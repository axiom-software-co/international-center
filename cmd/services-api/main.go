package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/axiom-software-co/international-center/internal/services"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("Starting services-api...")
	
	// Environment configuration
	port := getEnv("SERVICES_API_PORT", "8080")
	dbConnectionString := getEnv("DATABASE_CONNECTION_STRING", "")
	
	if dbConnectionString == "" {
		log.Fatal("DATABASE_CONNECTION_STRING environment variable is required")
	}
	
	// Database connection
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Database connection test failed: %v", err)
	}
	log.Println("Database connection established")
	
	// Initialize layers
	repository := services.NewPostgreSQLServicesRepository(db)
	service := services.NewServicesService(repository)
	handler := services.NewServicesHandler(service)
	
	// Setup routes
	router := mux.NewRouter()
	
	// Health check endpoints
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")
	router.HandleFunc("/health/ready", readinessCheckHandler(db)).Methods("GET")
	
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

func readinessCheckHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"not_ready","error":"%s"}`, err.Error())
			return
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