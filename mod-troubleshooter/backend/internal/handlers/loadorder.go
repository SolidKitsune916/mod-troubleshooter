package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mod-troubleshooter/backend/internal/archive"
	"github.com/mod-troubleshooter/backend/internal/cache"
	"github.com/mod-troubleshooter/backend/internal/loadorder"
	"github.com/mod-troubleshooter/backend/internal/nexus"
	"github.com/mod-troubleshooter/backend/internal/plugin"
)

// LoadOrderAnalyzeRequest is the request body for load order analysis.
type LoadOrderAnalyzeRequest struct {
	// Plugins is a list of plugins to analyze in their intended load order.
	// Each plugin should include game, modId, and fileId for downloading,
	// or just filename for manual analysis.
	Plugins []PluginReference `json:"plugins"`
}

// PluginReference identifies a plugin for analysis.
type PluginReference struct {
	// Filename is the plugin filename (required).
	Filename string `json:"filename"`
	// Game is the game domain for downloading from Nexus (optional).
	Game string `json:"game,omitempty"`
	// ModID is the mod ID on Nexus (optional).
	ModID int `json:"modId,omitempty"`
	// FileID is the file ID on Nexus (optional).
	FileID int `json:"fileId,omitempty"`
}

// LoadOrderAnalyzeResponse is the response from load order analysis.
type LoadOrderAnalyzeResponse struct {
	*loadorder.AnalysisResult
	Cached bool `json:"cached"`
}

// LoadOrderHandler handles load order analysis HTTP requests.
type LoadOrderHandler struct {
	clientGetter NexusClientGetter
	downloader   *archive.Downloader
	extractor    *archive.Extractor
	cache        *cache.Cache
	analyzer     *loadorder.Analyzer
	parser       *plugin.Parser
}

// LoadOrderHandlerConfig holds configuration for the LoadOrderHandler.
type LoadOrderHandlerConfig struct {
	ClientGetter NexusClientGetter
	Downloader   *archive.Downloader
	Extractor    *archive.Extractor
	Cache        *cache.Cache
}

// NewLoadOrderHandler creates a new load order handler.
func NewLoadOrderHandler(cfg LoadOrderHandlerConfig) *LoadOrderHandler {
	return &LoadOrderHandler{
		clientGetter: cfg.ClientGetter,
		downloader:   cfg.Downloader,
		extractor:    cfg.Extractor,
		cache:        cfg.Cache,
		analyzer:     loadorder.NewAnalyzer(),
		parser:       plugin.NewParser(),
	}
}

// AnalyzeLoadOrder handles POST /api/loadorder/analyze
// Analyzes a list of plugins and returns dependency issues and stats.
func (h *LoadOrderHandler) AnalyzeLoadOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req LoadOrderAnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Plugins) == 0 {
		WriteError(w, http.StatusBadRequest, "At least one plugin is required")
		return
	}

	// Build list of plugin files for analysis
	pluginFiles := make([]loadorder.PluginFile, 0, len(req.Plugins))

	for _, ref := range req.Plugins {
		if ref.Filename == "" {
			WriteError(w, http.StatusBadRequest, "Plugin filename is required")
			return
		}

		pf := loadorder.PluginFile{
			Filename: ref.Filename,
		}

		// If Nexus info is provided, try to fetch and parse the plugin
		if ref.Game != "" && ref.ModID > 0 && ref.FileID > 0 {
			header, err := h.fetchAndParsePlugin(ctx, ref)
			if err != nil {
				// Log the error but continue with just the filename
				log.Printf("Warning: could not fetch plugin %s: %v", ref.Filename, err)
			} else {
				pf.Header = header
			}
		}

		pluginFiles = append(pluginFiles, pf)
	}

	// Perform analysis
	result, err := h.analyzer.Analyze(ctx, pluginFiles)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			WriteError(w, http.StatusRequestTimeout, "Request cancelled")
			return
		}
		log.Printf("Error analyzing load order: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to analyze load order")
		return
	}

	response := LoadOrderAnalyzeResponse{
		AnalysisResult: result,
		Cached:         false,
	}

	WriteJSON(w, http.StatusOK, response)
}

// AnalyzeCollectionLoadOrder handles GET /api/collections/{slug}/revisions/{revision}/loadorder
// Analyzes the load order of all plugins in a collection revision.
func (h *LoadOrderHandler) AnalyzeCollectionLoadOrder(w http.ResponseWriter, r *http.Request) {
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

	// Check cache
	cacheKey := fmt.Sprintf("loadorder:%s:%d", slug, revision)
	var cachedResult LoadOrderAnalyzeResponse
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

	// Extract plugin files from the collection mods
	pluginFiles, err := h.extractPluginsFromCollection(ctx, client, gameDomain, revisionDetails)
	if err != nil {
		if errors.Is(err, nexus.ErrPremiumOnly) {
			WriteError(w, http.StatusForbidden, "This feature requires a Nexus Mods Premium account")
			return
		}
		log.Printf("Error extracting plugins: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to extract plugin information")
		return
	}

	// Perform analysis
	result, err := h.analyzer.Analyze(ctx, pluginFiles)
	if err != nil {
		log.Printf("Error analyzing load order: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to analyze load order")
		return
	}

	response := LoadOrderAnalyzeResponse{
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

// fetchAndParsePlugin downloads a plugin and parses its header.
func (h *LoadOrderHandler) fetchAndParsePlugin(ctx context.Context, ref PluginReference) (*plugin.PluginHeader, error) {
	client := h.clientGetter.Get()
	if client == nil {
		return nil, errors.New("nexus client not available")
	}

	// Get download links
	links, err := client.GetModFileDownloadLinks(ctx, ref.Game, ref.ModID, ref.FileID)
	if err != nil {
		return nil, fmt.Errorf("get download links: %w", err)
	}

	if len(links) == 0 {
		return nil, errors.New("no download links available")
	}

	// Download the file
	downloadResult, err := h.downloader.Download(ctx, links[0].URI, nil)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer h.downloader.CleanupPath(downloadResult.FilePath)

	// If it's an archive, try to extract the plugin
	if isArchive(downloadResult.FilePath) {
		return h.extractAndParsePluginFromArchive(ctx, downloadResult.FilePath, ref.Filename)
	}

	// If it's a direct plugin file, parse it
	if plugin.IsPluginFile(downloadResult.FilePath) {
		return h.parser.ParseFile(ctx, downloadResult.FilePath)
	}

	return nil, fmt.Errorf("unknown file type: %s", downloadResult.FilePath)
}

// extractAndParsePluginFromArchive extracts a specific plugin from an archive and parses it.
func (h *LoadOrderHandler) extractAndParsePluginFromArchive(ctx context.Context, archivePath, pluginFilename string) (*plugin.PluginHeader, error) {
	// List files to find the plugin
	files, err := h.extractor.ListFiles(ctx, archivePath)
	if err != nil {
		return nil, fmt.Errorf("list archive: %w", err)
	}

	// Find the plugin file in the archive
	var pluginPath string
	pluginLower := strings.ToLower(pluginFilename)
	for _, f := range files {
		if strings.ToLower(filepath.Base(f)) == pluginLower {
			pluginPath = f
			break
		}
	}

	if pluginPath == "" {
		return nil, fmt.Errorf("plugin %s not found in archive", pluginFilename)
	}

	// Extract just this plugin
	result, err := h.extractor.ExtractPaths(ctx, archivePath, []string{pluginPath})
	if err != nil {
		return nil, fmt.Errorf("extract plugin: %w", err)
	}
	defer h.extractor.Cleanup(result.OutputDir)

	if len(result.Files) == 0 {
		return nil, fmt.Errorf("plugin %s not extracted", pluginFilename)
	}

	// Parse the extracted plugin
	extractedPath := filepath.Join(result.OutputDir, result.Files[0])
	return h.parser.ParseFile(ctx, extractedPath)
}

// extractPluginsFromCollection extracts plugin information from collection mods.
func (h *LoadOrderHandler) extractPluginsFromCollection(ctx context.Context, client *nexus.Client, gameDomain string, revision *nexus.RevisionDetails) ([]loadorder.PluginFile, error) {
	var pluginFiles []loadorder.PluginFile

	for _, modFile := range revision.ModFiles {
		if modFile.File == nil || modFile.File.Mod == nil {
			continue
		}

		// Check if this mod file might contain plugins
		filename := modFile.File.Name
		lowerName := strings.ToLower(filename)

		// If the file itself is a plugin
		if plugin.IsPluginFile(filename) {
			pf := loadorder.PluginFile{
				Filename: filename,
			}

			// Try to get actual plugin header
			header, err := h.fetchModFilePlugin(ctx, client, gameDomain, modFile)
			if err != nil {
				log.Printf("Warning: could not fetch plugin %s: %v", filename, err)
			} else if header != nil {
				pf.Header = header
			}

			pluginFiles = append(pluginFiles, pf)
			continue
		}

		// If it's an archive, try to find plugins inside
		if isArchiveFilename(lowerName) {
			plugins, err := h.extractPluginsFromModFile(ctx, client, gameDomain, modFile)
			if err != nil {
				log.Printf("Warning: could not extract plugins from %s: %v", filename, err)
				continue
			}
			pluginFiles = append(pluginFiles, plugins...)
		}
	}

	return pluginFiles, nil
}

// fetchModFilePlugin downloads a mod file and parses its plugin header.
func (h *LoadOrderHandler) fetchModFilePlugin(ctx context.Context, client *nexus.Client, gameDomain string, modFile nexus.ModFileReference) (*plugin.PluginHeader, error) {
	if modFile.File == nil || modFile.File.Mod == nil {
		return nil, errors.New("incomplete mod file reference")
	}

	// Get download links
	links, err := client.GetModFileDownloadLinks(ctx, gameDomain, modFile.File.Mod.ModID, modFile.File.FileID)
	if err != nil {
		return nil, err
	}

	if len(links) == 0 {
		return nil, errors.New("no download links")
	}

	// Download
	downloadResult, err := h.downloader.Download(ctx, links[0].URI, nil)
	if err != nil {
		return nil, err
	}
	defer h.downloader.CleanupPath(downloadResult.FilePath)

	// Parse plugin
	return h.parser.ParseFile(ctx, downloadResult.FilePath)
}

// extractPluginsFromModFile extracts plugin files from an archive mod file.
func (h *LoadOrderHandler) extractPluginsFromModFile(ctx context.Context, client *nexus.Client, gameDomain string, modFile nexus.ModFileReference) ([]loadorder.PluginFile, error) {
	if modFile.File == nil || modFile.File.Mod == nil {
		return nil, errors.New("incomplete mod file reference")
	}

	// Get download links
	links, err := client.GetModFileDownloadLinks(ctx, gameDomain, modFile.File.Mod.ModID, modFile.File.FileID)
	if err != nil {
		return nil, err
	}

	if len(links) == 0 {
		return nil, errors.New("no download links")
	}

	// Download
	downloadResult, err := h.downloader.Download(ctx, links[0].URI, nil)
	if err != nil {
		return nil, err
	}
	defer h.downloader.CleanupPath(downloadResult.FilePath)

	// List files in archive
	files, err := h.extractor.ListFiles(ctx, downloadResult.FilePath)
	if err != nil {
		return nil, err
	}

	// Find all plugin files
	var pluginPaths []string
	for _, f := range files {
		if plugin.IsPluginFile(f) {
			pluginPaths = append(pluginPaths, f)
		}
	}

	if len(pluginPaths) == 0 {
		return nil, nil
	}

	// Extract plugin files
	extractResult, err := h.extractor.ExtractPaths(ctx, downloadResult.FilePath, pluginPaths)
	if err != nil {
		return nil, err
	}
	defer h.extractor.Cleanup(extractResult.OutputDir)

	// Parse each plugin
	var pluginFiles []loadorder.PluginFile
	for _, extractedFile := range extractResult.Files {
		extractedPath := filepath.Join(extractResult.OutputDir, extractedFile)
		filename := filepath.Base(extractedFile)

		pf := loadorder.PluginFile{
			Filename: filename,
		}

		header, err := h.parser.ParseFile(ctx, extractedPath)
		if err != nil {
			log.Printf("Warning: could not parse plugin %s: %v", filename, err)
		} else {
			pf.Header = header
		}

		pluginFiles = append(pluginFiles, pf)
	}

	return pluginFiles, nil
}

// isArchive checks if a file is an archive based on content type or extension.
func isArchive(filePath string) bool {
	// Try to identify by reading file header
	f, err := os.Open(filePath)
	if err != nil {
		return isArchiveFilename(strings.ToLower(filePath))
	}
	defer f.Close()

	// Read first few bytes
	header := make([]byte, 10)
	n, err := io.ReadFull(f, header)
	if err != nil || n < 4 {
		return isArchiveFilename(strings.ToLower(filePath))
	}

	// Check magic bytes
	// ZIP: PK\x03\x04
	if header[0] == 'P' && header[1] == 'K' && header[2] == 0x03 && header[3] == 0x04 {
		return true
	}
	// 7z: 7z\xBC\xAF\x27\x1C
	if header[0] == '7' && header[1] == 'z' && header[2] == 0xBC && header[3] == 0xAF {
		return true
	}
	// RAR: Rar!\x1A\x07
	if header[0] == 'R' && header[1] == 'a' && header[2] == 'r' && header[3] == '!' {
		return true
	}

	return isArchiveFilename(strings.ToLower(filePath))
}

// isArchiveFilename checks if a filename has an archive extension.
func isArchiveFilename(filename string) bool {
	switch {
	case strings.HasSuffix(filename, ".zip"):
		return true
	case strings.HasSuffix(filename, ".7z"):
		return true
	case strings.HasSuffix(filename, ".rar"):
		return true
	default:
		return false
	}
}
