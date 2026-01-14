package nexus

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Common errors returned by the client.
var (
	ErrNoAPIKey      = errors.New("nexus API key is required")
	ErrUnauthorized  = errors.New("invalid or expired API key")
	ErrRateLimited   = errors.New("rate limit exceeded")
	ErrNotFound      = errors.New("resource not found")
	ErrServerError   = errors.New("nexus server error")
	ErrGraphQLErrors = errors.New("graphql query returned errors")
)

// ClientConfig holds configuration for the Nexus client.
type ClientConfig struct {
	APIKey         string
	HTTPClient     *http.Client
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

// Client handles communication with the Nexus Mods API.
type Client struct {
	apiKey         string
	httpClient     *http.Client
	maxRetries     int
	initialBackoff time.Duration
	maxBackoff     time.Duration

	// Rate limiting state
	mu              sync.RWMutex
	lastRequest     time.Time
	minRequestDelay time.Duration
	rateLimitInfo   *RateLimitInfo
}

// NewClient creates a new Nexus API client with the given configuration.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, ErrNoAPIKey
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	initialBackoff := cfg.InitialBackoff
	if initialBackoff <= 0 {
		initialBackoff = 1 * time.Second
	}

	maxBackoff := cfg.MaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = 30 * time.Second
	}

	return &Client{
		apiKey:          cfg.APIKey,
		httpClient:      httpClient,
		maxRetries:      maxRetries,
		initialBackoff:  initialBackoff,
		maxBackoff:      maxBackoff,
		minRequestDelay: 100 * time.Millisecond, // ~10 requests per second max
	}, nil
}

// Query executes a GraphQL query against the Nexus API.
func (c *Client) Query(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := c.calculateBackoff(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		// Enforce rate limiting
		if err := c.waitForRateLimit(ctx); err != nil {
			return err
		}

		resp, err := c.doRequest(ctx, bodyBytes)
		if err != nil {
			lastErr = err
			// Only retry on transient errors
			if isRetryable(err) {
				continue
			}
			return err
		}

		// Parse and decode response
		if err := c.decodeResponse(resp, result); err != nil {
			lastErr = err
			if isRetryable(err) {
				continue
			}
			return err
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// doRequest performs the HTTP request and handles response status codes.
func (c *Client) doRequest(ctx context.Context, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, GraphQLEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("User-Agent", "ModTroubleshooter/1.0")

	c.mu.Lock()
	c.lastRequest = time.Now()
	c.mu.Unlock()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}

	// Parse rate limit headers
	c.parseRateLimitHeaders(resp)

	// Handle error status codes
	switch resp.StatusCode {
	case http.StatusOK:
		return resp, nil
	case http.StatusUnauthorized:
		resp.Body.Close()
		return nil, ErrUnauthorized
	case http.StatusTooManyRequests:
		resp.Body.Close()
		return nil, ErrRateLimited
	case http.StatusNotFound:
		resp.Body.Close()
		return nil, ErrNotFound
	default:
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			return nil, fmt.Errorf("%w: status %d", ErrServerError, resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
}

// decodeResponse parses the GraphQL response and checks for errors.
func (c *Client) decodeResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	// Check for GraphQL errors
	if len(gqlResp.Errors) > 0 {
		return fmt.Errorf("%w: %s", ErrGraphQLErrors, gqlResp.Errors[0].Message)
	}

	// Decode data into result
	if result != nil && gqlResp.Data != nil {
		dataBytes, err := json.Marshal(gqlResp.Data)
		if err != nil {
			return fmt.Errorf("marshal data: %w", err)
		}
		if err := json.Unmarshal(dataBytes, result); err != nil {
			return fmt.Errorf("decode data: %w", err)
		}
	}

	return nil
}

// waitForRateLimit ensures we don't exceed rate limits.
func (c *Client) waitForRateLimit(ctx context.Context) error {
	c.mu.RLock()
	lastReq := c.lastRequest
	minDelay := c.minRequestDelay
	c.mu.RUnlock()

	elapsed := time.Since(lastReq)
	if elapsed < minDelay {
		waitTime := minDelay - elapsed
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}

	return nil
}

// calculateBackoff returns the backoff duration for a given retry attempt.
func (c *Client) calculateBackoff(attempt int) time.Duration {
	backoff := float64(c.initialBackoff) * math.Pow(2, float64(attempt-1))
	if backoff > float64(c.maxBackoff) {
		backoff = float64(c.maxBackoff)
	}
	return time.Duration(backoff)
}

// parseRateLimitHeaders extracts rate limiting info from response headers.
func (c *Client) parseRateLimitHeaders(resp *http.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	info := &RateLimitInfo{}

	if v := resp.Header.Get("X-RL-Hourly-Limit"); v != "" {
		info.HourlyLimit, _ = strconv.Atoi(v)
	}
	if v := resp.Header.Get("X-RL-Hourly-Remaining"); v != "" {
		info.HourlyRemaining, _ = strconv.Atoi(v)
	}
	if v := resp.Header.Get("X-RL-Daily-Limit"); v != "" {
		info.DailyLimit, _ = strconv.Atoi(v)
	}
	if v := resp.Header.Get("X-RL-Daily-Remaining"); v != "" {
		info.DailyRemaining, _ = strconv.Atoi(v)
	}

	c.rateLimitInfo = info

	// Adjust rate limiting if running low
	if info.HourlyRemaining > 0 && info.HourlyRemaining < 10 {
		c.minRequestDelay = 1 * time.Second // Slow down when running low
	} else if info.HourlyRemaining > 100 {
		c.minRequestDelay = 100 * time.Millisecond // Normal rate
	}
}

// GetRateLimitInfo returns the current rate limit information.
func (c *Client) GetRateLimitInfo() *RateLimitInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.rateLimitInfo == nil {
		return nil
	}
	info := *c.rateLimitInfo
	return &info
}

// isRetryable returns true if the error is transient and can be retried.
func isRetryable(err error) bool {
	return errors.Is(err, ErrRateLimited) || errors.Is(err, ErrServerError)
}

// GetCollection fetches a collection by slug.
func (c *Client) GetCollection(ctx context.Context, slug string) (*Collection, error) {
	variables := map[string]interface{}{
		"slug": slug,
	}

	var resp CollectionResponse
	if err := c.Query(ctx, CollectionQuery, variables, &resp); err != nil {
		return nil, err
	}

	if resp.Collection == nil {
		return nil, ErrNotFound
	}

	return resp.Collection, nil
}

// GetCollectionRevisions fetches revision history for a collection.
func (c *Client) GetCollectionRevisions(ctx context.Context, domainName, slug string) ([]Revision, error) {
	variables := map[string]interface{}{
		"slug": slug,
	}
	if domainName != "" {
		variables["domainName"] = domainName
	}

	var resp CollectionRevisionsResponse
	if err := c.Query(ctx, CollectionRevisionsQuery, variables, &resp); err != nil {
		return nil, err
	}

	if resp.Collection == nil {
		return nil, ErrNotFound
	}

	return resp.Collection.Revisions, nil
}

// GetCollectionRevisionMods fetches mod files for a specific collection revision.
func (c *Client) GetCollectionRevisionMods(ctx context.Context, slug string, revision int) (*RevisionDetails, error) {
	variables := map[string]interface{}{
		"slug":     slug,
		"revision": revision,
	}

	var resp CollectionRevisionModsResponse
	if err := c.Query(ctx, CollectionRevisionModsQuery, variables, &resp); err != nil {
		return nil, err
	}

	if resp.CollectionRevision == nil {
		return nil, ErrNotFound
	}

	return resp.CollectionRevision, nil
}

// ValidateAPIKey checks if the API key is valid by making a test query.
func (c *Client) ValidateAPIKey(ctx context.Context) (bool, error) {
	// Use a simple query to validate the API key
	var resp struct {
		CurrentUser *struct {
			MemberID int `json:"memberId"`
		} `json:"currentUser"`
	}

	query := `query { currentUser { memberId } }`

	if err := c.Query(ctx, query, nil, &resp); err != nil {
		if errors.Is(err, ErrUnauthorized) {
			return false, nil
		}
		return false, err
	}

	return resp.CurrentUser != nil, nil
}
