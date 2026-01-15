package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/mod-troubleshooter/backend/internal/archive"
	"github.com/mod-troubleshooter/backend/internal/cache"
	"github.com/mod-troubleshooter/backend/internal/fomod"
	"github.com/mod-troubleshooter/backend/internal/nexus"
)

// FomodAnalyzeRequest is the request body for FOMOD analysis.
type FomodAnalyzeRequest struct {
	Game   string `json:"game"`
	ModID  int    `json:"modId"`
	FileID int    `json:"fileId"`
}

// FomodAnalyzeResponse is the response from FOMOD analysis.
type FomodAnalyzeResponse struct {
	Game     string          `json:"game"`
	ModID    int             `json:"modId"`
	FileID   int             `json:"fileId"`
	HasFomod bool            `json:"hasFomod"`
	Data     *fomod.FomodData `json:"data,omitempty"`
	Cached   bool            `json:"cached"`
}

// FomodHandler handles FOMOD analysis HTTP requests.
type FomodHandler struct {
	clientGetter NexusClientGetter
	downloader   *archive.Downloader
	extractor    *archive.Extractor
	cache        *cache.Cache
}

// FomodHandlerConfig holds configuration for the FomodHandler.
type FomodHandlerConfig struct {
	ClientGetter NexusClientGetter
	Downloader   *archive.Downloader
	Extractor    *archive.Extractor
	Cache        *cache.Cache
}

// NewFomodHandler creates a new FOMOD handler.
func NewFomodHandler(cfg FomodHandlerConfig) *FomodHandler {
	return &FomodHandler{
		clientGetter: cfg.ClientGetter,
		downloader:   cfg.Downloader,
		extractor:    cfg.Extractor,
		cache:        cfg.Cache,
	}
}

// AnalyzeFomod handles POST /api/fomod/analyze
// Downloads a mod archive, extracts the FOMOD data, and returns the parsed configuration.
func (h *FomodHandler) AnalyzeFomod(w http.ResponseWriter, r *http.Request) {
	client := h.clientGetter.Get()
	if client == nil {
		WriteError(w, http.StatusServiceUnavailable, "Nexus API key not configured. Please configure it in Settings.")
		return
	}

	ctx := r.Context()

	// Parse request body
	var req FomodAnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Game == "" {
		WriteError(w, http.StatusBadRequest, "Game domain is required")
		return
	}
	if req.ModID <= 0 {
		WriteError(w, http.StatusBadRequest, "Valid mod ID is required")
		return
	}
	if req.FileID <= 0 {
		WriteError(w, http.StatusBadRequest, "Valid file ID is required")
		return
	}

	// Check cache first
	cacheKey := cache.CacheKey(req.Game, req.ModID, req.FileID)
	var cachedResult FomodAnalyzeResponse
	if h.cache != nil {
		if err := h.cache.Get(ctx, cacheKey, &cachedResult); err == nil {
			cachedResult.Cached = true
			WriteJSON(w, http.StatusOK, cachedResult)
			return
		}
	}

	// Map game ID to Nexus domain name
	gameDomain := GetNexusDomain(req.Game)

	// Get download links from Nexus
	links, err := client.GetModFileDownloadLinks(ctx, gameDomain, req.ModID, req.FileID)
	if err != nil {
		handleFomodError(w, err)
		return
	}

	if len(links) == 0 {
		WriteError(w, http.StatusNotFound, "No download links available")
		return
	}

	// Use the first available download link
	downloadURL := links[0].URI

	// Download the archive
	log.Printf("Downloading mod archive from: %s", downloadURL)
	downloadResult, err := h.downloader.Download(ctx, downloadURL, nil)
	if err != nil {
		log.Printf("Error downloading archive: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to download mod archive")
		return
	}
	defer h.downloader.CleanupPath(downloadResult.FilePath)

	// Check if archive has FOMOD directory
	hasFomod, err := h.extractor.HasFomod(ctx, downloadResult.FilePath)
	if err != nil {
		log.Printf("Error checking for FOMOD: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to inspect archive")
		return
	}

	response := FomodAnalyzeResponse{
		Game:     req.Game,
		ModID:    req.ModID,
		FileID:   req.FileID,
		HasFomod: hasFomod,
		Cached:   false,
	}

	if !hasFomod {
		// Cache the negative result
		if h.cache != nil {
			if err := h.cache.Set(ctx, cacheKey, response); err != nil {
				log.Printf("Error caching result: %v", err)
			}
		}
		WriteJSON(w, http.StatusOK, response)
		return
	}

	// Extract FOMOD directory
	extractResult, err := h.extractor.ExtractFomod(ctx, downloadResult.FilePath)
	if err != nil {
		log.Printf("Error extracting FOMOD: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to extract FOMOD data")
		return
	}
	defer h.extractor.Cleanup(extractResult.OutputDir)

	// Parse FOMOD XML
	parser, err := fomod.NewParser(extractResult.OutputDir)
	if err != nil {
		if errors.Is(err, fomod.ErrNoFomodDir) {
			// This shouldn't happen since we checked HasFomod, but handle gracefully
			response.HasFomod = false
			if h.cache != nil {
				if err := h.cache.Set(ctx, cacheKey, response); err != nil {
					log.Printf("Error caching result: %v", err)
				}
			}
			WriteJSON(w, http.StatusOK, response)
			return
		}
		log.Printf("Error creating FOMOD parser: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to parse FOMOD data")
		return
	}

	fomodData, err := parser.Parse()
	if err != nil {
		if errors.Is(err, fomod.ErrNoModuleConfig) {
			// Has fomod directory but no ModuleConfig.xml
			response.HasFomod = false
			if h.cache != nil {
				if err := h.cache.Set(ctx, cacheKey, response); err != nil {
					log.Printf("Error caching result: %v", err)
				}
			}
			WriteJSON(w, http.StatusOK, response)
			return
		}
		if errors.Is(err, os.ErrNotExist) {
			// info.xml doesn't exist but ModuleConfig.xml does - this is okay
			// The parse should have continued, so this is an unexpected error
			log.Printf("Error parsing FOMOD: %v", err)
			WriteError(w, http.StatusInternalServerError, "Failed to parse FOMOD data")
			return
		}
		log.Printf("Error parsing FOMOD: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to parse FOMOD data")
		return
	}

	response.Data = fomodData

	// Cache the result
	if h.cache != nil {
		if err := h.cache.Set(ctx, cacheKey, response); err != nil {
			log.Printf("Error caching result: %v", err)
		}
	}

	WriteJSON(w, http.StatusOK, response)
}

// handleFomodError maps errors to HTTP responses for FOMOD analysis.
func handleFomodError(w http.ResponseWriter, err error) {
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
	case errors.Is(err, archive.ErrNoURL):
		WriteError(w, http.StatusBadRequest, "Download URL is required")
	case errors.Is(err, archive.ErrDownloadFailed):
		WriteError(w, http.StatusBadGateway, "Failed to download mod archive")
	case errors.Is(err, archive.ErrFileTooLarge):
		WriteError(w, http.StatusRequestEntityTooLarge, "Mod archive is too large")
	default:
		log.Printf("Error: FOMOD analysis failed: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to analyze FOMOD")
	}
}
