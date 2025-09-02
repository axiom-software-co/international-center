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
	log.Println("Starting admin-gateway...")
	
	// Environment configuration
	port := getEnv("ADMIN_GATEWAY_PORT", "8082")
	servicesAPIURL := getEnv("SERVICES_API_URL", "")
	contentAPIURL := getEnv("CONTENT_API_URL", "")
	rateLimitStr := getEnv("RATE_LIMIT_PER_MINUTE", "100")
	
	if servicesAPIURL == "" {
		log.Fatal("SERVICES_API_URL environment variable is required")
	}
	
	if contentAPIURL == "" {
		log.Fatal("CONTENT_API_URL environment variable is required")
	}
	
	rateLimit, err := strconv.Atoi(rateLimitStr)
	if err != nil {
		log.Fatalf("Invalid RATE_LIMIT_PER_MINUTE value: %v", err)
	}
	
	// Initialize gateway
	config := &gateway.AdminProxyConfig{
		ServicesAPIURL: servicesAPIURL,
		ContentAPIURL:  contentAPIURL,
		RateLimit:      rateLimit,
	}
	
	adminGateway := gateway.NewAdminGateway(config)
	
	// Setup routes
	router := mux.NewRouter()
	adminGateway.SetupRoutes(router)
	
	// Start server
	address := fmt.Sprintf(":%s", port)
	log.Printf("Admin Gateway listening on %s", address)
	log.Printf("Proxying services API: %s", servicesAPIURL)
	log.Printf("Proxying content API: %s", contentAPIURL)
	log.Printf("Rate limit: %d requests/minute per user", rateLimit)
	log.Println("Authentication required for all admin routes")
	log.Println("Audit logging enabled for compliance")
	
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