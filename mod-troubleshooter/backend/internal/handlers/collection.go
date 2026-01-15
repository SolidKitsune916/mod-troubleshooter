package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/mod-troubleshooter/backend/internal/nexus"
)

// NexusClientGetter provides dynamic access to a Nexus client.
type NexusClientGetter interface {
	Get() *nexus.Client
}

// CollectionHandler handles collection-related HTTP requests.
type CollectionHandler struct {
	client *nexus.Client
}

// NewCollectionHandler creates a new collection handler with a static client.
func NewCollectionHandler(client *nexus.Client) *CollectionHandler {
	return &CollectionHandler{client: client}
}

// DynamicCollectionHandler handles collection-related HTTP requests with a dynamic client.
type DynamicCollectionHandler struct {
	clientGetter NexusClientGetter
}

// NewDynamicCollectionHandler creates a new collection handler with a dynamic client getter.
func NewDynamicCollectionHandler(getter NexusClientGetter) *DynamicCollectionHandler {
	return &DynamicCollectionHandler{clientGetter: getter}
}

// extractSlug extracts the collection slug from either a full Nexus URL or a slug string.
// It handles URL-encoded URLs and extracts just the slug part.
func extractSlug(input string) string {
	if input == "" {
		return ""
	}

	// Try to decode URL encoding first
	decoded, err := url.QueryUnescape(input)
	if err == nil {
		input = decoded
	}

	// Try to parse as URL
	parsedURL, err := url.Parse(input)
	if err == nil && parsedURL.Host != "" {
		// It's a full URL, extract slug from path
		// Pattern: /games/{game}/collections/{slug}
		path := parsedURL.Path
		re := regexp.MustCompile(`/collections/([^/?#]+)`)
		matches := re.FindStringSubmatch(path)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// If it contains nexusmods.com pattern, try regex extraction
	if strings.Contains(input, "nexusmods.com") {
		re := regexp.MustCompile(`nexusmods\.com/[^/]+/collections/([^/?#]+)`)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// Otherwise, assume it's already a slug
	return input
}

// GetCollection handles GET /api/collections/{slug}
func (h *DynamicCollectionHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
	client := h.clientGetter.Get()
	if client == nil {
		WriteError(w, http.StatusServiceUnavailable, "Nexus API key not configured. Please configure it in Settings.")
		return
	}

	ctx := r.Context()
	rawSlug := r.PathValue("slug")
	if rawSlug == "" {
		WriteError(w, http.StatusBadRequest, "Collection slug is required")
		return
	}

	slug := extractSlug(rawSlug)
	if slug == "" {
		WriteError(w, http.StatusBadRequest, "Invalid collection slug or URL")
		return
	}

	log.Printf("Fetching collection with slug: %q", slug)
	collection, err := client.GetCollection(ctx, slug)
	if err != nil {
		log.Printf("Error fetching collection %q: %v (error type: %T)", slug, err, err)
		handleNexusError(w, err, "fetch collection")
		return
	}
	log.Printf("Successfully fetched collection: %s", collection.Name)

	WriteJSON(w, http.StatusOK, collection)
}

// GetCollectionRevisions handles GET /api/collections/{slug}/revisions
func (h *DynamicCollectionHandler) GetCollectionRevisions(w http.ResponseWriter, r *http.Request) {
	client := h.clientGetter.Get()
	if client == nil {
		WriteError(w, http.StatusServiceUnavailable, "Nexus API key not configured. Please configure it in Settings.")
		return
	}

	ctx := r.Context()
	rawSlug := r.PathValue("slug")
	if rawSlug == "" {
		WriteError(w, http.StatusBadRequest, "Collection slug is required")
		return
	}

	slug := extractSlug(rawSlug)
	if slug == "" {
		WriteError(w, http.StatusBadRequest, "Invalid collection slug or URL")
		return
	}

	domainName := r.URL.Query().Get("domain")
	revisions, err := client.GetCollectionRevisions(ctx, domainName, slug)
	if err != nil {
		handleNexusError(w, err, "fetch revisions")
		return
	}

	WriteJSON(w, http.StatusOK, revisions)
}

// GetCollectionRevisionMods handles GET /api/collections/{slug}/revisions/{revision}
func (h *DynamicCollectionHandler) GetCollectionRevisionMods(w http.ResponseWriter, r *http.Request) {
	client := h.clientGetter.Get()
	if client == nil {
		WriteError(w, http.StatusServiceUnavailable, "Nexus API key not configured. Please configure it in Settings.")
		return
	}

	ctx := r.Context()
	rawSlug := r.PathValue("slug")
	if rawSlug == "" {
		WriteError(w, http.StatusBadRequest, "Collection slug is required")
		return
	}

	slug := extractSlug(rawSlug)
	if slug == "" {
		WriteError(w, http.StatusBadRequest, "Invalid collection slug or URL")
		return
	}

	revisionStr := r.PathValue("revision")
	if revisionStr == "" {
		WriteError(w, http.StatusBadRequest, "Revision number is required")
		return
	}

	revision, err := strconv.Atoi(revisionStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid revision number")
		return
	}

	revisionDetails, err := client.GetCollectionRevisionMods(ctx, slug, revision)
	if err != nil {
		handleNexusError(w, err, "fetch revision mods")
		return
	}

	WriteJSON(w, http.StatusOK, revisionDetails)
}

// GetCollection handles GET /api/collections/{slug}
// Returns collection metadata including the latest revision's mod list.
func (h *CollectionHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rawSlug := r.PathValue("slug")
	if rawSlug == "" {
		WriteError(w, http.StatusBadRequest, "Collection slug is required")
		return
	}

	slug := extractSlug(rawSlug)
	if slug == "" {
		WriteError(w, http.StatusBadRequest, "Invalid collection slug or URL")
		return
	}

	collection, err := h.client.GetCollection(ctx, slug)
	if err != nil {
		handleNexusError(w, err, "fetch collection")
		return
	}

	WriteJSON(w, http.StatusOK, collection)
}

// GetCollectionRevisions handles GET /api/collections/{slug}/revisions
// Returns revision history for a collection.
// Optional query param: domain (game domain name for filtering)
func (h *CollectionHandler) GetCollectionRevisions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rawSlug := r.PathValue("slug")
	if rawSlug == "" {
		WriteError(w, http.StatusBadRequest, "Collection slug is required")
		return
	}

	slug := extractSlug(rawSlug)
	if slug == "" {
		WriteError(w, http.StatusBadRequest, "Invalid collection slug or URL")
		return
	}

	domainName := r.URL.Query().Get("domain")

	revisions, err := h.client.GetCollectionRevisions(ctx, domainName, slug)
	if err != nil {
		handleNexusError(w, err, "fetch revisions")
		return
	}

	WriteJSON(w, http.StatusOK, revisions)
}

// GetCollectionRevisionMods handles GET /api/collections/{slug}/revisions/{revision}
// Returns mod files for a specific collection revision.
func (h *CollectionHandler) GetCollectionRevisionMods(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rawSlug := r.PathValue("slug")
	if rawSlug == "" {
		WriteError(w, http.StatusBadRequest, "Collection slug is required")
		return
	}

	slug := extractSlug(rawSlug)
	if slug == "" {
		WriteError(w, http.StatusBadRequest, "Invalid collection slug or URL")
		return
	}

	revisionStr := r.PathValue("revision")
	if revisionStr == "" {
		WriteError(w, http.StatusBadRequest, "Revision number is required")
		return
	}

	revision, err := strconv.Atoi(revisionStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid revision number")
		return
	}

	revisionDetails, err := h.client.GetCollectionRevisionMods(ctx, slug, revision)
	if err != nil {
		handleNexusError(w, err, "fetch revision mods")
		return
	}

	WriteJSON(w, http.StatusOK, revisionDetails)
}

// handleNexusError maps Nexus client errors to HTTP responses.
func handleNexusError(w http.ResponseWriter, err error, action string) {
	if err == nil {
		log.Printf("Warning: handleNexusError called with nil error for %s", action)
		WriteError(w, http.StatusInternalServerError, "Unknown error occurred")
		return
	}
	
	// Always log the full error details
	log.Printf("Nexus API error during %s: %+v (type: %T)", action, err, err)
	
	// Build error message with details
	errorDetail := err.Error()
	
	switch {
	case errors.Is(err, nexus.ErrNotFound):
		WriteError(w, http.StatusNotFound, "Resource not found: "+errorDetail)
		return
	case errors.Is(err, nexus.ErrUnauthorized):
		WriteError(w, http.StatusUnauthorized, "Invalid or missing Nexus API key: "+errorDetail)
		return
	case errors.Is(err, nexus.ErrPremiumOnly):
		WriteError(w, http.StatusForbidden, "This feature requires a Nexus Mods Premium account: "+errorDetail)
		return
	case errors.Is(err, nexus.ErrRateLimited):
		WriteError(w, http.StatusTooManyRequests, "Nexus API rate limit exceeded, please try again later: "+errorDetail)
		return
	case errors.Is(err, nexus.ErrNoAPIKey):
		WriteError(w, http.StatusServiceUnavailable, "Nexus API key not configured: "+errorDetail)
		return
	case errors.Is(err, nexus.ErrGraphQLErrors):
		WriteError(w, http.StatusInternalServerError, "GraphQL error: "+errorDetail)
		return
	case errors.Is(err, nexus.ErrServerError):
		WriteError(w, http.StatusBadGateway, "Nexus server error: "+errorDetail)
		return
	default:
		// Include full error details for debugging
		WriteError(w, http.StatusInternalServerError, "Failed to "+action+": "+errorDetail)
		return
	}
}
