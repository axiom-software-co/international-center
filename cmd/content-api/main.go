package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/axiom-software-co/international-center/internal/content"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("Starting content-api...")
	
	// Environment configuration
	port := getEnv("CONTENT_API_PORT", "8081")
	databaseConnectionString := getEnv("DATABASE_CONNECTION_STRING", "")
	
	if databaseConnectionString == "" {
		log.Fatal("DATABASE_CONNECTION_STRING environment variable is required")
	}
	
	// PostgreSQL connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	db, err := sql.Open("postgres", databaseConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close PostgreSQL connection: %v", err)
		}
	}()
	
	// Test PostgreSQL connection
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("PostgreSQL connection test failed: %v", err)
	}
	log.Println("PostgreSQL connection established")
	
	// Initialize layers
	repository := content.NewPostgreSQLContentRepository(db)
	service := content.NewContentService(repository)
	handler := content.NewContentHandler(service)
	
	// Setup routes
	router := mux.NewRouter()
	
	// Health check endpoints
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")
	router.HandleFunc("/health/ready", readinessCheckHandler(db)).Methods("GET")
	
	// Register content routes
	handler.RegisterRoutes(router)
	
	// Start server
	address := fmt.Sprintf(":%s", port)
	log.Printf("Content API listening on %s", address)
	
	if err := http.ListenAndServe(address, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok","service":"content-api"}`)
}

func readinessCheckHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := db.PingContext(ctx); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"not_ready","error":"%s"}`, err.Error())
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ready","service":"content-api"}`)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}