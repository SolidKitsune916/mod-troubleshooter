package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/mod-troubleshooter/backend/internal/nexus"
)

// DownloadHandler handles download-related HTTP requests.
type DownloadHandler struct {
	clientGetter NexusClientGetter
}

// NewDownloadHandler creates a new download handler with a dynamic client getter.
func NewDownloadHandler(getter NexusClientGetter) *DownloadHandler {
	return &DownloadHandler{clientGetter: getter}
}

// GetModFileDownloadLinks handles GET /api/games/{game}/mods/{modId}/files/{fileId}/download
// Returns download URLs for the specified mod file.
// This endpoint requires a Nexus Mods Premium account.
func (h *DownloadHandler) GetModFileDownloadLinks(w http.ResponseWriter, r *http.Request) {
	client := h.clientGetter.Get()
	if client == nil {
		WriteError(w, http.StatusServiceUnavailable, "Nexus API key not configured. Please configure it in Settings.")
		return
	}

	ctx := r.Context()

	// Extract path parameters
	game := r.PathValue("game")
	if game == "" {
		WriteError(w, http.StatusBadRequest, "Game domain is required")
		return
	}

	modIDStr := r.PathValue("modId")
	if modIDStr == "" {
		WriteError(w, http.StatusBadRequest, "Mod ID is required")
		return
	}

	modID, err := strconv.Atoi(modIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid mod ID")
		return
	}

	fileIDStr := r.PathValue("fileId")
	if fileIDStr == "" {
		WriteError(w, http.StatusBadRequest, "File ID is required")
		return
	}

	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid file ID")
		return
	}

	// Map game ID to Nexus domain name
	gameDomain := GetNexusDomain(game)

	// Fetch download links from Nexus API
	links, err := client.GetModFileDownloadLinks(ctx, gameDomain, modID, fileID)
	if err != nil {
		handleDownloadError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, links)
}

// handleDownloadError maps Nexus client errors to HTTP responses for download endpoints.
func handleDownloadError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, nexus.ErrNotFound):
		WriteError(w, http.StatusNotFound, "Mod file not found")
	case errors.Is(err, nexus.ErrUnauthorized):
		WriteError(w, http.StatusUnauthorized, "Invalid or missing Nexus API key")
	case errors.Is(err, nexus.ErrPremiumOnly):
		WriteError(w, http.StatusForbidden, "This feature requires a Nexus Mods Premium account")
	case errors.Is(err, nexus.ErrRateLimited):
		WriteError(w, http.StatusTooManyRequests, "Nexus API rate limit exceeded, please try again later")
	case errors.Is(err, nexus.ErrNoAPIKey):
		WriteError(w, http.StatusServiceUnavailable, "Nexus API key not configured")
	default:
		log.Printf("Error: failed to fetch download links: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch download links")
	}
}
