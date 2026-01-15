package nexus

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ClientConfig
		wantErr error
	}{
		{
			name:    "missing API key",
			cfg:     ClientConfig{},
			wantErr: ErrNoAPIKey,
		},
		{
			name: "valid config",
			cfg: ClientConfig{
				APIKey: "test-api-key",
			},
			wantErr: nil,
		},
		{
			name: "custom settings",
			cfg: ClientConfig{
				APIKey:         "test-api-key",
				MaxRetries:     5,
				InitialBackoff: 2 * time.Second,
				MaxBackoff:     60 * time.Second,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.cfg)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("got error %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if client == nil {
				t.Error("expected client, got nil")
			}
		})
	}
}

func TestClient_Query(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantErr    bool
		errContain string
	}{
		{
			name: "successful query",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Verify headers
				if r.Header.Get("apikey") != "test-api-key" {
					t.Error("missing apikey header")
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Error("missing content-type header")
				}

				// Return valid response
				resp := GraphQLResponse{
					Data: map[string]interface{}{
						"collection": map[string]interface{}{
							"id":   "123",
							"name": "Test Collection",
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
			wantErr: false,
		},
		{
			name: "graphql errors",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := GraphQLResponse{
					Errors: []GraphQLError{
						{Message: "Collection not found"},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
			wantErr:    true,
			errContain: "Collection not found",
		},
		{
			name: "unauthorized",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			wantErr:    true,
			errContain: "invalid or expired API key",
		},
		{
			name: "rate limited",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
			},
			wantErr:    true,
			errContain: "max retries exceeded",
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr:    true,
			errContain: "max retries exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client, err := NewClient(ClientConfig{
				APIKey:         "test-api-key",
				MaxRetries:     1,
				InitialBackoff: 10 * time.Millisecond,
			})
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			// Create custom client that uses test server
			client.httpClient = &http.Client{
				Transport: &testTransport{
					server: server,
				},
			}

			var result CollectionResponse
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = client.Query(ctx, CollectionQuery, map[string]interface{}{"slug": "test"}, &result)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("error %q doesn't contain %q", err.Error(), tt.errContain)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_GetCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request body to verify variables
		var req GraphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		slug, ok := req.Variables["slug"].(string)
		if !ok || slug == "" {
			t.Error("missing slug variable")
		}

		resp := GraphQLResponse{
			Data: map[string]interface{}{
				"collection": map[string]interface{}{
					"id":             123,
					"slug":           slug,
					"name":           "Test Collection",
					"summary":        "A test collection",
					"endorsements":   100,
					"totalDownloads": 5000,
					"user": map[string]interface{}{
						"name":     "TestUser",
						"memberId": 12345,
					},
					"game": map[string]interface{}{
						"domainName": "skyrimspecialedition",
						"name":       "Skyrim Special Edition",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		APIKey: "test-api-key",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	client.httpClient = &http.Client{
		Transport: &testTransport{server: server},
	}

	ctx := context.Background()
	collection, err := client.GetCollection(ctx, "test-slug")
	if err != nil {
		t.Fatalf("GetCollection failed: %v", err)
	}

	if collection.ID != 123 {
		t.Errorf("got ID %d, want %d", collection.ID, 123)
	}
	if collection.Name != "Test Collection" {
		t.Errorf("got Name %q, want %q", collection.Name, "Test Collection")
	}
	if collection.User.Name != "TestUser" {
		t.Errorf("got User.Name %q, want %q", collection.User.Name, "TestUser")
	}
	if collection.Game.DomainName != "skyrimspecialedition" {
		t.Errorf("got Game.DomainName %q, want %q", collection.Game.DomainName, "skyrimspecialedition")
	}
}

func TestClient_GetCollection_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GraphQLResponse{
			Data: map[string]interface{}{
				"collection": nil,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		APIKey: "test-api-key",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	client.httpClient = &http.Client{
		Transport: &testTransport{server: server},
	}

	ctx := context.Background()
	_, err = client.GetCollection(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("got error %v, want %v", err, ErrNotFound)
	}
}

func TestClient_RateLimitHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RL-Hourly-Limit", "100")
		w.Header().Set("X-RL-Hourly-Remaining", "50")
		w.Header().Set("X-RL-Daily-Limit", "2500")
		w.Header().Set("X-RL-Daily-Remaining", "2000")

		resp := GraphQLResponse{
			Data: map[string]interface{}{
				"collection": map[string]interface{}{
					"id": "test",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		APIKey: "test-api-key",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	client.httpClient = &http.Client{
		Transport: &testTransport{server: server},
	}

	// Make a request to trigger header parsing
	ctx := context.Background()
	_, _ = client.GetCollection(ctx, "test")

	info := client.GetRateLimitInfo()
	if info == nil {
		t.Fatal("expected rate limit info, got nil")
	}
	if info.HourlyLimit != 100 {
		t.Errorf("got HourlyLimit %d, want %d", info.HourlyLimit, 100)
	}
	if info.HourlyRemaining != 50 {
		t.Errorf("got HourlyRemaining %d, want %d", info.HourlyRemaining, 50)
	}
	if info.DailyLimit != 2500 {
		t.Errorf("got DailyLimit %d, want %d", info.DailyLimit, 2500)
	}
	if info.DailyRemaining != 2000 {
		t.Errorf("got DailyRemaining %d, want %d", info.DailyRemaining, 2000)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		APIKey: "test-api-key",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	client.httpClient = &http.Client{
		Transport: &testTransport{server: server},
	}

	// Cancel context immediately
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var result CollectionResponse
	err = client.Query(ctx, CollectionQuery, map[string]interface{}{"slug": "test"}, &result)
	if err == nil {
		t.Error("expected context cancellation error")
	}
}

func TestCalculateBackoff(t *testing.T) {
	client, err := NewClient(ClientConfig{
		APIKey:         "test",
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     10 * time.Second,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{5, 10 * time.Second}, // Capped at max
		{6, 10 * time.Second}, // Still capped
	}

	for _, tt := range tests {
		got := client.calculateBackoff(tt.attempt)
		if got != tt.want {
			t.Errorf("attempt %d: got %v, want %v", tt.attempt, got, tt.want)
		}
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{ErrRateLimited, true},
		{ErrServerError, true},
		{ErrUnauthorized, false},
		{ErrNotFound, false},
		{ErrNoAPIKey, false},
	}

	for _, tt := range tests {
		got := isRetryable(tt.err)
		if got != tt.want {
			t.Errorf("isRetryable(%v) = %v, want %v", tt.err, got, tt.want)
		}
	}
}

// testTransport redirects requests to a test server
type testTransport struct {
	server *httptest.Server
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Redirect to test server
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(t.server.URL, "http://")
	return http.DefaultTransport.RoundTrip(req)
}
