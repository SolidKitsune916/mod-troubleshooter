package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

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

// GetCollection handles GET /api/collections/{slug}
func (h *DynamicCollectionHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
	client := h.clientGetter.Get()
	if client == nil {
		WriteError(w, http.StatusServiceUnavailable, "Nexus API key not configured. Please configure it in Settings.")
		return
	}

	ctx := r.Context()
	slug := r.PathValue("slug")
	if slug == "" {
		WriteError(w, http.StatusBadRequest, "Collection slug is required")
		return
	}

	collection, err := client.GetCollection(ctx, slug)
	if err != nil {
		handleNexusError(w, err, "fetch collection")
		return
	}

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
	slug := r.PathValue("slug")
	if slug == "" {
		WriteError(w, http.StatusBadRequest, "Collection slug is required")
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
	slug := r.PathValue("slug")
	if slug == "" {
		WriteError(w, http.StatusBadRequest, "Collection slug is required")
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

	slug := r.PathValue("slug")
	if slug == "" {
		WriteError(w, http.StatusBadRequest, "Collection slug is required")
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

	slug := r.PathValue("slug")
	if slug == "" {
		WriteError(w, http.StatusBadRequest, "Collection slug is required")
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

	slug := r.PathValue("slug")
	if slug == "" {
		WriteError(w, http.StatusBadRequest, "Collection slug is required")
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
	switch {
	case errors.Is(err, nexus.ErrNotFound):
		WriteError(w, http.StatusNotFound, "Resource not found")
	case errors.Is(err, nexus.ErrUnauthorized):
		WriteError(w, http.StatusUnauthorized, "Invalid or missing Nexus API key")
	case errors.Is(err, nexus.ErrPremiumOnly):
		WriteError(w, http.StatusForbidden, "This feature requires a Nexus Mods Premium account")
	case errors.Is(err, nexus.ErrRateLimited):
		WriteError(w, http.StatusTooManyRequests, "Nexus API rate limit exceeded, please try again later")
	case errors.Is(err, nexus.ErrNoAPIKey):
		WriteError(w, http.StatusServiceUnavailable, "Nexus API key not configured")
	default:
		log.Printf("Error: failed to %s: %v", action, err)
		WriteError(w, http.StatusInternalServerError, "Failed to "+action)
	}
}
