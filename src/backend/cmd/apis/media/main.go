package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/media"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/gorilla/mux"
)

// MediaAPIApplication represents the media API application
type MediaAPIApplication struct {
	daprClient      *dapr.Client
	mediaRepository *media.MediaInquiryRepository
	mediaService    *media.MediaService
	mediaHandler    *media.MediaHandler
	server          *http.Server
}

func main() {
	// Create application
	app, err := NewMediaAPIApplication()
	if err != nil {
		log.Fatalf("Failed to create media API application: %v", err)
	}
	
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle shutdown signals
	go handleShutdownSignals(cancel)
	
	// Start the application
	if err := app.Start(ctx); err != nil {
		log.Fatalf("Media API application failed: %v", err)
	}
	
	log.Println("Media API application shutdown complete")
}

// NewMediaAPIApplication creates a new media API application
func NewMediaAPIApplication() (*MediaAPIApplication, error) {
	// Initialize Dapr client
	daprClient, err := dapr.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Dapr client: %w", err)
	}
	
	// Initialize media repository
	mediaRepository := media.NewMediaInquiryRepository(daprClient)
	
	// Initialize media service
	mediaService := media.NewMediaService(mediaRepository)
	
	// Initialize media handler
	mediaHandler := media.NewMediaHandler(mediaService)
	
	// Create HTTP server
	server := &http.Server{
		Addr:         getServerAddress(),
		Handler:      createRouter(mediaHandler),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	return &MediaAPIApplication{
		daprClient:      daprClient,
		mediaRepository: mediaRepository,
		mediaService:    mediaService,
		mediaHandler:    mediaHandler,
		server:          server,
	}, nil
}

// Start starts the media API application
func (app *MediaAPIApplication) Start(ctx context.Context) error {
	log.Printf("Starting Media API application on %s", app.server.Addr)
	log.Printf("Environment: %s", getEnvironment())
	log.Printf("Version: %s", getVersion())
	
	// Validate Dapr connectivity
	if err := app.validateDaprConnectivity(ctx); err != nil {
		return fmt.Errorf("Dapr connectivity validation failed: %w", err)
	}
	
	// Start HTTP server in goroutine
	go func() {
		log.Printf("Media API server listening on %s", app.server.Addr)
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Media API server error: %v", err)
		}
	}()
	
	// Wait for context cancellation
	<-ctx.Done()
	
	// Shutdown gracefully
	return app.Shutdown()
}

// Shutdown gracefully shuts down the application
func (app *MediaAPIApplication) Shutdown() error {
	log.Println("Shutting down Media API application...")
	
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	// Shutdown HTTP server
	if err := app.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Media API server shutdown error: %v", err)
		return err
	}
	
	// Close Dapr client
	if err := app.daprClient.Close(); err != nil {
		log.Printf("Dapr client close error: %v", err)
	}
	
	log.Println("Media API application shut down successfully")
	return nil
}

// validateDaprConnectivity validates Dapr connectivity
func (app *MediaAPIApplication) validateDaprConnectivity(ctx context.Context) error {
	// Create a context with timeout for connectivity check
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	// Check if Dapr client is healthy
	if !app.daprClient.IsHealthy(checkCtx) {
		return fmt.Errorf("Dapr client is not healthy")
	}
	
	log.Println("Dapr connectivity validated successfully")
	return nil
}

// createRouter creates the HTTP router with all media routes
func createRouter(handler *media.MediaHandler) http.Handler {
	router := mux.NewRouter()
	
	// Register media routes
	handler.RegisterRoutes(router)
	
	// Add health check routes
	router.HandleFunc("/health", handler.HealthCheck).Methods("GET")
	router.HandleFunc("/ready", handler.ReadinessCheck).Methods("GET")
	
	// Add CORS middleware for development
	return addCORSMiddleware(router)
}

// addCORSMiddleware adds CORS middleware for development
func addCORSMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for development
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Correlation-ID")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		handler.ServeHTTP(w, r)
	})
}

// handleShutdownSignals handles OS shutdown signals
func handleShutdownSignals(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	sig := <-sigChan
	log.Printf("Received shutdown signal: %v", sig)
	cancel()
}

// Configuration helpers

// getServerAddress returns the server address from environment
func getServerAddress() string {
	addr := os.Getenv("MEDIA_API_ADDR")
	if addr == "" {
		log.Fatalf("MEDIA_API_ADDR environment variable is required")
	}
	return addr
}

// getEnvironment returns the environment from environment variable
func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		log.Fatalf("ENVIRONMENT environment variable is required")
	}
	return env
}

// getVersion returns the application version from environment variable
func getVersion() string {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		log.Fatalf("APP_VERSION environment variable is required")
	}
	return version
}