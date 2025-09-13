package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/notifications"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/gorilla/mux"
)

func main() {
	// Get environment variables
	port := getEnv("PORT", "8095")
	environment := getEnv("ENVIRONMENT", "development")
	
	log.Printf("Starting Notifications Service (email, sms, slack routing) on port %s in %s environment", port, environment)

	// Initialize Dapr client
	daprClient, err := dapr.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Dapr client: %v", err)
	}
	defer daprClient.Close()

	// Validate Dapr service registration for notifications-api
	ctx := context.Background()
	if err := daprClient.ValidateServiceRegistration(ctx); err != nil {
		log.Fatalf("Notifications service registration validation failed: %v", err)
	}
	log.Printf("Notifications service (notifications-api) successfully registered with Dapr runtime")

	// Initialize configuration
	config := notifications.DefaultNotificationConfig()
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize logger
	logLevel := slog.LevelInfo
	if config.Observability.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}
	
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Create notification service using Dapr client
	notificationService, err := notifications.NewNotificationHandler(daprClient)
	if err != nil {
		log.Fatalf("Failed to create notification handler: %v", err)
	}

	// Setup router
	router := mux.NewRouter()
	
	// Register all notification domain routes
	notificationService.RegisterRoutes(router)
	
	// Health endpoints
	router.HandleFunc("/health", notificationService.HealthCheck).Methods("GET")
	router.HandleFunc("/health/ready", notificationService.ReadinessCheck).Methods("GET")

	// Create server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Notifications Service listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down Notifications Service...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Notifications Service shutdown complete")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}