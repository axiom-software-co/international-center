package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/inquiries/volunteers"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/gorilla/mux"
)

func main() {
	// Get environment variables
	port := getEnv("PORT", "8087")
	environment := getEnv("ENVIRONMENT", "development")
	
	log.Printf("Starting Volunteers API on port %s in %s environment", port, environment)

	// Initialize Dapr client
	daprClient, err := dapr.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Dapr client: %v", err)
	}
	defer daprClient.Close()

	// Initialize Dapr components
	stateStore := dapr.NewStateStore(daprClient)
	bindings := dapr.NewBindings(daprClient)
	pubsub := dapr.NewPubSub(daprClient)

	// Create repository
	volunteerRepository := volunteers.NewVolunteerRepository(stateStore, bindings, pubsub)

	// Create service
	volunteerService := volunteers.NewVolunteerService(volunteerRepository)

	// Create handler
	volunteerHandler := volunteers.NewVolunteerHandler(volunteerService)

	// Setup router
	router := mux.NewRouter()
	
	// Register routes
	volunteerHandler.RegisterRoutes(router)
	
	// Health endpoints
	router.HandleFunc("/health", volunteerHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/health/ready", volunteerHandler.ReadinessCheck).Methods("GET")

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
		log.Printf("Volunteers API server listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down Volunteers API server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Volunteers API server shutdown complete")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}