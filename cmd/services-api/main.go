package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/axiom-software-co/international-center/internal/services"
	"github.com/dapr/go-sdk/client"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting services-api...")
	
	// Environment configuration
	port := getEnv("SERVICES_API_PORT", "8080")
	stateStoreName := getEnv("DAPR_STATE_STORE_NAME", "services-store")
	
	// Dapr client connection
	daprClient, err := client.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Dapr client: %v", err)
	}
	defer daprClient.Close()
	
	log.Println("Dapr client connection established")
	
	// Initialize layers
	repository := services.NewDaprStateStoreRepository(daprClient, stateStoreName)
	service := services.NewServicesService(repository)
	handler := services.NewServicesHandler(service)
	
	// Setup routes
	router := mux.NewRouter()
	
	// Health check endpoints
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")
	router.HandleFunc("/health/ready", readinessCheckHandler(daprClient, stateStoreName)).Methods("GET")
	
	// Register service routes
	handler.RegisterRoutes(router)
	
	// Start server
	address := fmt.Sprintf(":%s", port)
	log.Printf("Services API listening on %s", address)
	
	if err := http.ListenAndServe(address, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok","service":"services-api"}`)
}

func readinessCheckHandler(daprClient client.Client, stateStoreName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Test Dapr connection by attempting to get a state that doesn't exist
		_, err := daprClient.GetState(r.Context(), stateStoreName, "health-check", nil)
		if err != nil {
			// This is expected for a non-existent key, so we check if it's a connection error
			if err.Error() != "state not found" && err.Error() != "error getting state: state not found" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprintf(w, `{"status":"not_ready","error":"%s"}`, err.Error())
				return
			}
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ready","service":"services-api"}`)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}