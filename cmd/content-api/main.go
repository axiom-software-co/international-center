package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/axiom-software-co/international-center/internal/content"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	log.Println("Starting content-api...")
	
	// Environment configuration
	port := getEnv("CONTENT_API_PORT", "8081")
	mongoConnectionString := getEnv("MONGO_CONNECTION_STRING", "")
	mongoDatabaseName := getEnv("MONGO_DATABASE_NAME", "international_center")
	
	if mongoConnectionString == "" {
		log.Fatal("MONGO_CONNECTION_STRING environment variable is required")
	}
	
	// MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoConnectionString))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()
	
	// Test MongoDB connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB connection test failed: %v", err)
	}
	log.Println("MongoDB connection established")
	
	db := client.Database(mongoDatabaseName)
	
	// Initialize layers
	repository := content.NewMongoContentRepository(db)
	service := content.NewContentService(repository)
	handler := content.NewContentHandler(service)
	
	// Setup routes
	router := mux.NewRouter()
	
	// Health check endpoints
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")
	router.HandleFunc("/health/ready", readinessCheckHandler(client)).Methods("GET")
	
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

func readinessCheckHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := client.Ping(ctx, nil); err != nil {
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