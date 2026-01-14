package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/mod-troubleshooter/backend/internal/archive"
	"github.com/mod-troubleshooter/backend/internal/cache"
	"github.com/mod-troubleshooter/backend/internal/config"
	"github.com/mod-troubleshooter/backend/internal/handlers"
	"github.com/mod-troubleshooter/backend/internal/nexus"
	"github.com/rs/cors"
)

// clientManager manages the Nexus client lifecycle with thread-safe updates.
type clientManager struct {
	mu     sync.RWMutex
	client *nexus.Client
}

func (m *clientManager) Get() *nexus.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.client
}

func (m *clientManager) Set(client *nexus.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.client = client
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /api/health", healthHandler)

	// Initialize settings store with initial API key
	settingsStore := handlers.NewSettingsStore(cfg.NexusAPIKey)

	// Client manager for dynamic client updates
	clientMgr := &clientManager{}

	// Initialize Nexus client if API key is configured
	if cfg.NexusAPIKey != "" {
		nexusClient, err := nexus.NewClient(nexus.ClientConfig{
			APIKey: cfg.NexusAPIKey,
		})
		if err != nil {
			log.Fatalf("Failed to create Nexus client: %v", err)
		}
		clientMgr.Set(nexusClient)
	} else {
		log.Println("Warning: Nexus API key not configured, collection endpoints will return errors until configured")
	}

	// Set up callback to update client when API key changes
	settingsStore.SetOnKeyChange(func(newKey string) {
		if newKey == "" {
			clientMgr.Set(nil)
			log.Println("Nexus API key cleared")
			return
		}

		newClient, err := nexus.NewClient(nexus.ClientConfig{
			APIKey: newKey,
		})
		if err != nil {
			log.Printf("Failed to create new Nexus client: %v", err)
			return
		}
		clientMgr.Set(newClient)
		log.Println("Nexus API key updated")
	})

	// Settings endpoints (always available)
	settingsHandler := handlers.NewSettingsHandler(settingsStore)
	mux.HandleFunc("GET /api/settings", settingsHandler.GetSettings)
	mux.HandleFunc("POST /api/settings", settingsHandler.UpdateSettings)
	mux.HandleFunc("POST /api/settings/validate", settingsHandler.ValidateAPIKey)

	// Collection endpoints with dynamic client lookup
	collectionHandler := handlers.NewDynamicCollectionHandler(clientMgr)
	mux.HandleFunc("GET /api/collections/{slug}", collectionHandler.GetCollection)
	mux.HandleFunc("GET /api/collections/{slug}/revisions", collectionHandler.GetCollectionRevisions)
	mux.HandleFunc("GET /api/collections/{slug}/revisions/{revision}", collectionHandler.GetCollectionRevisionMods)

	// Download endpoints (requires Premium)
	downloadHandler := handlers.NewDownloadHandler(clientMgr)
	mux.HandleFunc("GET /api/games/{game}/mods/{modId}/files/{fileId}/download", downloadHandler.GetModFileDownloadLinks)

	// Initialize archive downloader and extractor
	downloader, err := archive.NewDownloader(archive.DownloaderConfig{
		TempDir:     filepath.Join(cfg.DataDir, "downloads"),
		MaxFileSize: 5 * 1024 * 1024 * 1024, // 5GB max
	})
	if err != nil {
		log.Fatalf("Failed to create downloader: %v", err)
	}

	extractor, err := archive.NewExtractor(archive.ExtractorConfig{
		TempDir:      filepath.Join(cfg.DataDir, "extracted"),
		MaxFileSize:  100 * 1024 * 1024,        // 100MB per file
		MaxTotalSize: 1024 * 1024 * 1024,       // 1GB total
	})
	if err != nil {
		log.Fatalf("Failed to create extractor: %v", err)
	}

	// Initialize cache for FOMOD analysis results
	fomodCache, err := cache.New(cache.Config{
		DBPath: filepath.Join(cfg.DataDir, "cache.db"),
		TTL:    time.Duration(cfg.CacheTTLHours) * time.Hour,
	})
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}

	// FOMOD analysis endpoints (requires Premium)
	fomodHandler := handlers.NewFomodHandler(handlers.FomodHandlerConfig{
		ClientGetter: clientMgr,
		Downloader:   downloader,
		Extractor:    extractor,
		Cache:        fomodCache,
	})
	mux.HandleFunc("POST /api/fomod/analyze", fomodHandler.AnalyzeFomod)

	// Load order analysis endpoints (requires Premium for collection analysis)
	loadOrderHandler := handlers.NewLoadOrderHandler(handlers.LoadOrderHandlerConfig{
		ClientGetter: clientMgr,
		Downloader:   downloader,
		Extractor:    extractor,
		Cache:        fomodCache,
	})
	mux.HandleFunc("POST /api/loadorder/analyze", loadOrderHandler.AnalyzeLoadOrder)
	mux.HandleFunc("GET /api/collections/{slug}/revisions/{revision}/loadorder", loadOrderHandler.AnalyzeCollectionLoadOrder)

	// Conflict analysis endpoints (requires Premium for downloading mod archives)
	conflictHandler := handlers.NewConflictHandler(handlers.ConflictHandlerConfig{
		ClientGetter: clientMgr,
		Downloader:   downloader,
		Cache:        fomodCache,
	})
	mux.HandleFunc("POST /api/conflicts/analyze", conflictHandler.AnalyzeConflicts)
	mux.HandleFunc("GET /api/collections/{slug}/revisions/{revision}/conflicts", conflictHandler.AnalyzeCollectionConflicts)

	// Configure CORS for React frontend
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	handler := c.Handler(mux)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on http://localhost:%s", cfg.Port)
		log.Printf("Environment: %s", cfg.Environment)
		log.Printf("Data directory: %s", cfg.DataDir)
		if cfg.NexusAPIKey != "" {
			log.Printf("Nexus API key: configured")
		} else {
			log.Printf("Nexus API key: not configured")
		}
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	// Cleanup resources
	if err := fomodCache.Close(); err != nil {
		log.Printf("Error closing cache: %v", err)
	}
	if err := downloader.Cleanup(); err != nil {
		log.Printf("Error cleaning up downloads: %v", err)
	}

	log.Println("Server stopped")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
