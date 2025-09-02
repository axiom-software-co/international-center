package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/axiom-software-co/international-center/internal/gateway"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting public-gateway...")
	
	// Environment configuration
	port := getEnv("PUBLIC_GATEWAY_PORT", "8080")
	rateLimitStr := getEnv("RATE_LIMIT_PER_MINUTE", "1000")
	
	rateLimit, err := strconv.Atoi(rateLimitStr)
	if err != nil {
		log.Fatalf("Invalid RATE_LIMIT_PER_MINUTE value: %v", err)
	}
	
	// Initialize gateway
	config := &gateway.ProxyConfig{
		RateLimit: rateLimit,
	}
	
	publicGateway := gateway.NewPublicGateway(config)
	
	// Setup routes
	router := mux.NewRouter()
	publicGateway.SetupRoutes(router)
	
	// Start server
	address := fmt.Sprintf(":%s", port)
	log.Printf("Public Gateway listening on %s", address)
	log.Println("Using Dapr service invocation for backend communication")
	log.Printf("Rate limit: %d requests/minute per IP", rateLimit)
	
	if err := http.ListenAndServe(address, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}