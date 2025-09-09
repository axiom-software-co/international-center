package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/health"
)

func main() {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	port := 8080
	
	log.Printf("Starting standalone health server for environment: %s on port %d", environment, port)
	
	server := health.NewHealthServer(port, environment)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Start the health server
	if err := server.Start(ctx); err != nil {
		log.Fatalf("Failed to start health server: %v", err)
	}
	
	log.Printf("Health server started successfully on port %d", port)
	log.Printf("Available endpoints:")
	log.Printf("  - http://localhost:%d/health - Overall health check", port)
	log.Printf("  - http://localhost:%d/health/<component> - Component health check", port)
	log.Printf("  - http://localhost:%d/status - Container status", port)
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigChan
	log.Printf("Shutting down health server...")
	
	if err := server.Stop(ctx); err != nil {
		log.Printf("Error stopping health server: %v", err)
	} else {
		log.Printf("Health server stopped successfully")
	}
}