package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/notifications"
	_ "github.com/lib/pq"
)

func main() {
	// Get environment variables
	port := getEnv("PORT", "8095")
	environment := getEnv("ENVIRONMENT", "development")
	
	log.Printf("Starting Notifications Service (email, sms, slack routing) on port %s in %s environment", port, environment)

	// Initialize configuration
	config := notifications.DefaultNotificationConfig()
	
	// Override config from environment
	if dbConnStr := os.Getenv("DATABASE_CONNECTION_STRING"); dbConnStr != "" {
		config.Database.ConnectionString = dbConnStr
	}
	
	if mqConnStr := os.Getenv("MESSAGE_QUEUE_CONNECTION_STRING"); mqConnStr != "" {
		config.MessageQueue.ConnectionString = mqConnStr
	}

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

	// Initialize database connection
	db, err := sql.Open("postgres", config.Database.ConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Configure database connection pool
	db.SetMaxOpenConns(config.Database.MaxOpenConnections)
	db.SetMaxIdleConns(config.Database.MaxIdleConnections)
	db.SetConnMaxLifetime(30 * time.Minute)

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize subscriber repository
	subscriberRepo, err := notifications.NewPostgreSQLSubscriberRepository(db, logger)
	if err != nil {
		log.Fatalf("Failed to create subscriber repository: %v", err)
	}

	// Initialize message queue client
	messageQueue := notifications.NewRabbitMQClient(
		config.MessageQueue.ConnectionString,
		logger,
		config.MessageQueue,
	)

	// Initialize notification publishers
	emailPublisher := notifications.NewRabbitMQEmailPublisher(messageQueue, logger)
	smsPublisher := notifications.NewRabbitMQSMSPublisher(messageQueue, logger)
	slackPublisher := notifications.NewRabbitMQSlackPublisher(messageQueue, logger)

	// Create notification router service
	notificationService := notifications.NewNotificationRouterService(
		subscriberRepo,
		messageQueue,
		emailPublisher,
		smsPublisher,
		slackPublisher,
		logger,
		config,
	)

	// Start health check server if enabled
	var healthServer *http.Server
	if config.Observability.HealthCheckPort > 0 {
		healthServer = startHealthCheckServer(config.Observability.HealthCheckPort, notificationService)
	}

	// Start notification service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := notificationService.Start(ctx); err != nil {
			log.Printf("Notification service error: %v", err)
			cancel()
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	select {
	case <-c:
		log.Println("Shutting down Notifications Service...")
	case <-ctx.Done():
		log.Println("Notifications Service context cancelled...")
	}

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop notification service
	if err := notificationService.Stop(shutdownCtx); err != nil {
		log.Printf("Notification service shutdown error: %v", err)
	}

	// Stop health check server if running
	if healthServer != nil {
		if err := healthServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("Health check server shutdown error: %v", err)
		}
	}

	log.Println("Notifications Service shutdown complete")
}

// startHealthCheckServer starts a simple health check HTTP server
func startHealthCheckServer(port int, service *notifications.NotificationRouterService) *http.Server {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		
		status, err := service.GetHealthStatus(ctx)
		if err != nil {
			http.Error(w, "Health check failed", http.StatusInternalServerError)
			return
		}
		
		if status.Status == "unhealthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"` + status.Status + `","service":"notifications-service"}`))
	})
	
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready","service":"notifications-service"}`))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Health check server listening on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health check server error: %v", err)
		}
	}()

	return server
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}