package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mod-troubleshooter/backend/internal/config"
	"github.com/mod-troubleshooter/backend/internal/handlers"
	"github.com/mod-troubleshooter/backend/internal/nexus"
	"github.com/rs/cors"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /api/health", healthHandler)

	// Initialize Nexus client if API key is configured
	if cfg.NexusAPIKey != "" {
		nexusClient, err := nexus.NewClient(nexus.ClientConfig{
			APIKey: cfg.NexusAPIKey,
		})
		if err != nil {
			log.Fatalf("Failed to create Nexus client: %v", err)
		}

		// Collection endpoints
		collectionHandler := handlers.NewCollectionHandler(nexusClient)
		mux.HandleFunc("GET /api/collections/{slug}", collectionHandler.GetCollection)
		mux.HandleFunc("GET /api/collections/{slug}/revisions", collectionHandler.GetCollectionRevisions)
		mux.HandleFunc("GET /api/collections/{slug}/revisions/{revision}", collectionHandler.GetCollectionRevisionMods)
	} else {
		log.Println("Warning: Nexus API key not configured, collection endpoints disabled")
	}

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
	log.Println("Server stopped")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
