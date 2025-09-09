package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/content"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/gorilla/mux"
)

func main() {
	// Get environment variables
	port := getEnv("PORT", "8080")
	environment := getEnv("ENVIRONMENT", "development")
	
	log.Printf("Starting Content Service (events, news, research, services) on port %s in %s environment", port, environment)

	// Initialize Dapr client
	daprClient, err := dapr.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Dapr client: %v", err)
	}
	defer daprClient.Close()

	// Create consolidated content handler
	contentHandler, err := content.NewContentHandler(daprClient)
	if err != nil {
		log.Fatalf("Failed to create content handler: %v", err)
	}

	// Setup router
	router := mux.NewRouter()
	
	// Register all content domain routes
	contentHandler.RegisterRoutes(router)
	
	// Health endpoints
	router.HandleFunc("/health", contentHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/health/ready", contentHandler.ReadinessCheck).Methods("GET")

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
		log.Printf("Content Service listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down Content Service...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Content Service shutdown complete")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}