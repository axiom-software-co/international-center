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

	"github.com/axiom-software-co/international-center/src/internal/services"
	"github.com/axiom-software-co/international-center/src/internal/shared/dapr"
	"github.com/gorilla/mux"
)

// ServicesAPIApplication represents the services API application
type ServicesAPIApplication struct {
	daprClient         *dapr.Client
	servicesRepository *services.ServicesRepository
	servicesService    *services.ServicesService
	servicesHandler    *services.ServicesHandler
	server             *http.Server
}

func main() {
	// Create application
	app, err := NewServicesAPIApplication()
	if err != nil {
		log.Fatalf("Failed to create services API application: %v", err)
	}
	
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle shutdown signals
	go handleShutdownSignals(cancel)
	
	// Start the application
	if err := app.Start(ctx); err != nil {
		log.Fatalf("Services API application failed: %v", err)
	}
	
	log.Println("Services API application shutdown complete")
}

// NewServicesAPIApplication creates a new services API application
func NewServicesAPIApplication() (*ServicesAPIApplication, error) {
	// Initialize Dapr client
	daprClient, err := dapr.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Dapr client: %w", err)
	}
	
	// Initialize services repository
	servicesRepository := services.NewServicesRepository(daprClient)
	
	// Initialize services service
	servicesService := services.NewServicesService(servicesRepository)
	
	// Initialize services handler
	servicesHandler := services.NewServicesHandler(servicesService)
	
	// Create HTTP server
	server := &http.Server{
		Addr:         getServerAddress(),
		Handler:      createRouter(servicesHandler),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	return &ServicesAPIApplication{
		daprClient:         daprClient,
		servicesRepository: servicesRepository,
		servicesService:    servicesService,
		servicesHandler:    servicesHandler,
		server:             server,
	}, nil
}

// Start starts the services API application
func (app *ServicesAPIApplication) Start(ctx context.Context) error {
	log.Printf("Starting Services API application on %s", app.server.Addr)
	log.Printf("Environment: %s", getEnvironment())
	log.Printf("Version: %s", getVersion())
	
	// Validate Dapr connectivity
	if err := app.validateDaprConnectivity(ctx); err != nil {
		return fmt.Errorf("Dapr connectivity validation failed: %w", err)
	}
	
	// Start HTTP server in goroutine
	go func() {
		log.Printf("Services API server listening on %s", app.server.Addr)
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Services API server error: %v", err)
		}
	}()
	
	// Wait for context cancellation
	<-ctx.Done()
	
	// Shutdown gracefully
	return app.Shutdown()
}

// Shutdown gracefully shuts down the application
func (app *ServicesAPIApplication) Shutdown() error {
	log.Println("Shutting down Services API application...")
	
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	// Shutdown HTTP server
	if err := app.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Services API server shutdown error: %v", err)
		return err
	}
	
	// Close Dapr client
	if err := app.daprClient.Close(); err != nil {
		log.Printf("Dapr client close error: %v", err)
	}
	
	log.Println("Services API application shut down successfully")
	return nil
}

// validateDaprConnectivity validates Dapr connectivity
func (app *ServicesAPIApplication) validateDaprConnectivity(ctx context.Context) error {
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

// createRouter creates the HTTP router with all services routes
func createRouter(handler *services.ServicesHandler) http.Handler {
	router := mux.NewRouter()
	
	// Register services routes
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
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

// getServerAddress returns the server address from environment or default
func getServerAddress() string {
	if addr := os.Getenv("SERVICES_API_ADDR"); addr != "" {
		return addr
	}
	return ":8081"
}

// getEnvironment returns the environment from environment variable or default
func getEnvironment() string {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return env
	}
	return "development"
}

// getVersion returns the application version from environment variable or default
func getVersion() string {
	if version := os.Getenv("APP_VERSION"); version != "" {
		return version
	}
	return "1.0.0"
}