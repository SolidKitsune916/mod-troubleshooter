package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mod-troubleshooter/backend/internal/archive"
	"github.com/mod-troubleshooter/backend/internal/cache"
	"github.com/mod-troubleshooter/backend/internal/conflict"
	"github.com/mod-troubleshooter/backend/internal/manifest"
	"github.com/mod-troubleshooter/backend/internal/nexus"
)

// ConflictAnalyzeRequest is the request body for conflict analysis.
type ConflictAnalyzeRequest struct {
	// Mods is a list of mods to analyze for conflicts in their intended load order.
	// Each mod should include game, modId, and fileId for downloading from Nexus.
	Mods []ModReference `json:"mods"`
	// IncludeContentHashes enables content-based duplicate detection (slower).
	IncludeContentHashes bool `json:"includeContentHashes,omitempty"`
}

// ModReference identifies a mod for conflict analysis.
type ModReference struct {
	// ModID is a unique identifier for this mod (used for display and tracking).
	ModID string `json:"modId"`
	// ModName is the display name of the mod.
	ModName string `json:"modName"`
	// Game is the game domain for downloading from Nexus.
	Game string `json:"game"`
	// NexusModID is the mod ID on Nexus.
	NexusModID int `json:"nexusModId"`
	// FileID is the file ID on Nexus.
	FileID int `json:"fileId"`
}

// ConflictAnalyzeResponse is the response from conflict analysis.
type ConflictAnalyzeResponse struct {
	*conflict.AnalysisResult
	Cached bool `json:"cached"`
}

// ConflictHandler handles conflict analysis HTTP requests.
type ConflictHandler struct {
	clientGetter      NexusClientGetter
	downloader        *archive.Downloader
	manifestExtractor *manifest.Extractor
	cache             *cache.Cache
	analyzer          *conflict.Analyzer
}

// ConflictHandlerConfig holds configuration for the ConflictHandler.
type ConflictHandlerConfig struct {
	ClientGetter NexusClientGetter
	Downloader   *archive.Downloader
	Cache        *cache.Cache
}

// NewConflictHandler creates a new conflict handler.
func NewConflictHandler(cfg ConflictHandlerConfig) *ConflictHandler {
	return &ConflictHandler{
		clientGetter:      cfg.ClientGetter,
		downloader:        cfg.Downloader,
		manifestExtractor: manifest.NewExtractor(),
		cache:             cfg.Cache,
		analyzer:          conflict.NewAnalyzer(),
	}
}

// AnalyzeConflicts handles POST /api/conflicts/analyze
// Analyzes a list of mods and returns file conflict information.
func (h *ConflictHandler) AnalyzeConflicts(w http.ResponseWriter, r *http.Request) {
	client := h.clientGetter.Get()
	if client == nil {
		WriteError(w, http.StatusServiceUnavailable, "Nexus API key not configured. Please configure it in Settings.")
		return
	}

	ctx := r.Context()

	// Parse request body
	var req ConflictAnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Mods) == 0 {
		WriteError(w, http.StatusBadRequest, "At least one mod is required")
		return
	}

	if len(req.Mods) < 2 {
		WriteError(w, http.StatusBadRequest, "At least two mods are required for conflict analysis")
		return
	}

	// Validate all mod references
	for i, mod := range req.Mods {
		if mod.ModID == "" {
			WriteError(w, http.StatusBadRequest, fmt.Sprintf("ModID is required for mod at index %d", i))
			return
		}
		if mod.Game == "" {
			WriteError(w, http.StatusBadRequest, fmt.Sprintf("Game domain is required for mod '%s'", mod.ModID))
			return
		}
		if mod.NexusModID <= 0 {
			WriteError(w, http.StatusBadRequest, fmt.Sprintf("Valid Nexus mod ID is required for mod '%s'", mod.ModID))
			return
		}
		if mod.FileID <= 0 {
			WriteError(w, http.StatusBadRequest, fmt.Sprintf("Valid file ID is required for mod '%s'", mod.ModID))
			return
		}
	}

	// Build list of mod manifests for analysis
	modManifests, err := h.fetchModManifests(ctx, client, req.Mods, req.IncludeContentHashes)
	if err != nil {
		if errors.Is(err, nexus.ErrPremiumOnly) {
			WriteError(w, http.StatusForbidden, "This feature requires a Nexus Mods Premium account")
			return
		}
		if errors.Is(err, context.Canceled) {
			WriteError(w, http.StatusRequestTimeout, "Request cancelled")
			return
		}
		log.Printf("Error fetching mod manifests: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch mod information")
		return
	}

	// Perform conflict analysis
	result, err := h.analyzer.Analyze(ctx, modManifests)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			WriteError(w, http.StatusRequestTimeout, "Request cancelled")
			return
		}
		log.Printf("Error analyzing conflicts: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to analyze conflicts")
		return
	}

	response := ConflictAnalyzeResponse{
		AnalysisResult: result,
		Cached:         false,
	}

	WriteJSON(w, http.StatusOK, response)
}

// AnalyzeCollectionConflicts handles GET /api/collections/{slug}/revisions/{revision}/conflicts
// Analyzes file conflicts for all mods in a collection revision.
func (h *ConflictHandler) AnalyzeCollectionConflicts(w http.ResponseWriter, r *http.Request) {
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

	// Check for optional query params
	includeHashes := r.URL.Query().Get("includeHashes") == "true"

	// Check cache
	cacheKey := fmt.Sprintf("conflicts:%s:%d:%t", slug, revision, includeHashes)
	var cachedResult ConflictAnalyzeResponse
	if h.cache != nil {
		if err := h.cache.Get(ctx, cacheKey, &cachedResult); err == nil {
			cachedResult.Cached = true
			WriteJSON(w, http.StatusOK, cachedResult)
			return
		}
	}

	// Get collection revision mods
	revisionDetails, err := client.GetCollectionRevisionMods(ctx, slug, revision)
	if err != nil {
		handleNexusError(w, err, "fetch collection revision")
		return
	}

	// Get the collection to determine the game
	collection, err := client.GetCollection(ctx, slug)
	if err != nil {
		handleNexusError(w, err, "fetch collection")
		return
	}

	gameDomain := collection.Game.DomainName

	// Extract mod manifests from the collection
	modManifests, err := h.extractManifestsFromCollection(ctx, client, gameDomain, revisionDetails, includeHashes)
	if err != nil {
		if errors.Is(err, nexus.ErrPremiumOnly) {
			WriteError(w, http.StatusForbidden, "This feature requires a Nexus Mods Premium account")
			return
		}
		log.Printf("Error extracting manifests: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to extract mod information")
		return
	}

	if len(modManifests) < 2 {
		// Not enough mods for conflict analysis, return empty result
		response := ConflictAnalyzeResponse{
			AnalysisResult: &conflict.AnalysisResult{
				Conflicts:    []conflict.Conflict{},
				ModSummaries: []conflict.ModConflictSummary{},
				FileToMods:   make(map[string][]string),
				Stats:        conflict.Stats{ByFileType: make(map[manifest.FileType]int)},
			},
			Cached: false,
		}
		WriteJSON(w, http.StatusOK, response)
		return
	}

	// Perform conflict analysis
	result, err := h.analyzer.Analyze(ctx, modManifests)
	if err != nil {
		log.Printf("Error analyzing conflicts: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to analyze conflicts")
		return
	}

	response := ConflictAnalyzeResponse{
		AnalysisResult: result,
		Cached:         false,
	}

	// Cache the result
	if h.cache != nil {
		if err := h.cache.Set(ctx, cacheKey, response); err != nil {
			log.Printf("Error caching result: %v", err)
		}
	}

	WriteJSON(w, http.StatusOK, response)
}

// fetchModManifests downloads mod archives and extracts their file manifests.
func (h *ConflictHandler) fetchModManifests(ctx context.Context, client *nexus.Client, mods []ModReference, includeHashes bool) ([]conflict.ModManifest, error) {
	modManifests := make([]conflict.ModManifest, 0, len(mods))

	for i, mod := range mods {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		modManifest := conflict.ModManifest{
			ModID:     mod.ModID,
			ModName:   mod.ModName,
			LoadOrder: i,
		}

		// Get download links (map game ID to Nexus domain)
		modGameDomain := GetNexusDomain(mod.Game)
		links, err := client.GetModFileDownloadLinks(ctx, modGameDomain, mod.NexusModID, mod.FileID)
		if err != nil {
			// Log and continue with empty manifest
			log.Printf("Warning: could not get download links for mod %s: %v", mod.ModID, err)
			modManifests = append(modManifests, modManifest)
			continue
		}

		if len(links) == 0 {
			log.Printf("Warning: no download links available for mod %s", mod.ModID)
			modManifests = append(modManifests, modManifest)
			continue
		}

		// Download the archive
		downloadResult, err := h.downloader.Download(ctx, links[0].URI, nil)
		if err != nil {
			log.Printf("Warning: could not download mod %s: %v", mod.ModID, err)
			modManifests = append(modManifests, modManifest)
			continue
		}

		// Extract manifest
		var manifestData *manifest.Manifest
		if includeHashes {
			manifestData, err = h.manifestExtractor.ExtractManifestWithHashes(ctx, downloadResult.FilePath)
		} else {
			manifestData, err = h.manifestExtractor.ExtractManifest(ctx, downloadResult.FilePath)
		}

		// Clean up downloaded file
		h.downloader.CleanupPath(downloadResult.FilePath)

		if err != nil {
			log.Printf("Warning: could not extract manifest for mod %s: %v", mod.ModID, err)
			modManifests = append(modManifests, modManifest)
			continue
		}

		modManifest.Manifest = manifestData
		modManifests = append(modManifests, modManifest)
	}

	return modManifests, nil
}

// extractManifestsFromCollection extracts file manifests from all mods in a collection.
func (h *ConflictHandler) extractManifestsFromCollection(ctx context.Context, client *nexus.Client, gameDomain string, revision *nexus.RevisionDetails, includeHashes bool) ([]conflict.ModManifest, error) {
	var modManifests []conflict.ModManifest

	for i, modFile := range revision.ModFiles {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if modFile.File == nil || modFile.File.Mod == nil {
			continue
		}

		modName := modFile.File.Mod.Name
		if modName == "" {
			modName = modFile.File.Name
		}

		modManifest := conflict.ModManifest{
			ModID:     fmt.Sprintf("%d-%d", modFile.File.Mod.ModID, modFile.File.FileID),
			ModName:   modName,
			LoadOrder: i,
		}

		// Only process archive files for conflict detection
		filename := modFile.File.Name
		lowerName := strings.ToLower(filename)
		if !isArchiveFilename(lowerName) {
			// Skip non-archive files (individual plugins, etc.)
			continue
		}

		// Get download links
		links, err := client.GetModFileDownloadLinks(ctx, gameDomain, modFile.File.Mod.ModID, modFile.File.FileID)
		if err != nil {
			log.Printf("Warning: could not get download links for %s: %v", filename, err)
			continue
		}

		if len(links) == 0 {
			log.Printf("Warning: no download links for %s", filename)
			continue
		}

		// Download the archive
		downloadResult, err := h.downloader.Download(ctx, links[0].URI, nil)
		if err != nil {
			log.Printf("Warning: could not download %s: %v", filename, err)
			continue
		}

		// Extract manifest
		var manifestData *manifest.Manifest
		if includeHashes {
			manifestData, err = h.manifestExtractor.ExtractManifestWithHashes(ctx, downloadResult.FilePath)
		} else {
			manifestData, err = h.manifestExtractor.ExtractManifest(ctx, downloadResult.FilePath)
		}

		// Clean up downloaded file
		h.downloader.CleanupPath(downloadResult.FilePath)

		if err != nil {
			log.Printf("Warning: could not extract manifest from %s: %v", filename, err)
			continue
		}

		modManifest.Manifest = manifestData
		modManifests = append(modManifests, modManifest)
	}

	return modManifests, nil
}
