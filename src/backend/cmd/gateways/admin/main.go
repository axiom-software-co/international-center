package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/axiom-software-co/international-center/src/backend/internal/gateway"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
)

// AdminGatewayApplication represents the admin gateway application
type AdminGatewayApplication struct {
	daprClient     *dapr.Client
	gatewayService *gateway.GatewayService
}

func main() {
	// Create application
	app, err := NewAdminGatewayApplication()
	if err != nil {
		log.Fatalf("Failed to create admin gateway application: %v", err)
	}
	
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle shutdown signals
	go handleShutdownSignals(cancel)
	
	// Start the application
	if err := app.Start(ctx); err != nil {
		log.Fatalf("Admin gateway application failed: %v", err)
	}
	
	log.Println("Admin gateway application shutdown complete")
}

// NewAdminGatewayApplication creates a new admin gateway application
func NewAdminGatewayApplication() (*AdminGatewayApplication, error) {
	// Initialize Dapr client
	daprClient, err := dapr.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Dapr client: %w", err)
	}
	
	// Create admin gateway service with environment-specific configuration
	gatewayService := createAdminGatewayService(daprClient)
	
	return &AdminGatewayApplication{
		daprClient:     daprClient,
		gatewayService: gatewayService,
	}, nil
}

// Start starts the admin gateway application
func (app *AdminGatewayApplication) Start(ctx context.Context) error {
	config := app.gatewayService.GetConfiguration()
	
	log.Printf("Starting Admin Gateway application")
	log.Printf("Gateway Name: %s", config.Name)
	log.Printf("Gateway Type: %s", config.Type)
	log.Printf("Environment: %s", config.Environment)
	log.Printf("Version: %s", config.Version)
	log.Printf("Listen Address: %s", config.GetListenAddress())
	
	// Log configuration details
	logGatewayConfiguration(config)
	
	// Validate Dapr connectivity
	if err := app.validateDaprConnectivity(ctx); err != nil {
		return fmt.Errorf("Dapr connectivity validation failed: %w", err)
	}
	
	// Start gateway service
	if err := app.gatewayService.Start(ctx); err != nil {
		return fmt.Errorf("gateway service failed: %w", err)
	}
	
	return nil
}

// validateDaprConnectivity validates Dapr connectivity
func (app *AdminGatewayApplication) validateDaprConnectivity(ctx context.Context) error {
	// Check if Dapr client is healthy
	if !app.daprClient.IsHealthy(ctx) {
		return fmt.Errorf("Dapr client is not healthy")
	}
	
	// Perform gateway health check (which includes backend service checks)
	if err := app.gatewayService.HealthCheck(ctx); err != nil {
		log.Printf("Warning: Gateway health check failed: %v", err)
		// Don't fail startup - let the gateway start and report unhealthy
	}
	
	log.Println("Dapr connectivity validated successfully")
	return nil
}

// createAdminGatewayService creates an admin gateway service with environment-specific configuration
func createAdminGatewayService(daprClient *dapr.Client) *gateway.GatewayService {
	// Create base admin gateway configuration
	config := gateway.NewAdminGatewayConfiguration()
	
	// Override configuration from environment variables
	updateConfigurationFromEnvironment(config)
	
	// Create gateway service
	return gateway.NewGatewayService(config, daprClient)
}

// updateConfigurationFromEnvironment updates configuration from environment variables
func updateConfigurationFromEnvironment(config *gateway.GatewayConfiguration) {
	// Update environment
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Environment = env
	}
	
	// Update port
	if port := getEnvInt("ADMIN_GATEWAY_PORT", config.Port); port > 0 {
		config.Port = port
	}
	
	// Update version
	if version := os.Getenv("APP_VERSION"); version != "" {
		config.Version = version
	}
	
	// Update CORS origins from environment (admin specific)
	if origins := os.Getenv("ADMIN_CORS_ALLOWED_ORIGINS"); origins != "" {
		// In production, this would parse the comma-separated origins
		// For now, keeping the default configuration
	}
	
	// Update rate limiting from environment (admin specific)
	if rpmStr := os.Getenv("ADMIN_RATE_LIMIT_REQUESTS_PER_MINUTE"); rpmStr != "" {
		if rpm := getEnvInt("ADMIN_RATE_LIMIT_REQUESTS_PER_MINUTE", config.RateLimit.RequestsPerMinute); rpm > 0 {
			config.RateLimit.RequestsPerMinute = rpm
		}
	}
	
	// Admin gateway always requires authentication (override from environment if needed)
	if requireAuth := os.Getenv("ADMIN_REQUIRE_AUTHENTICATION"); requireAuth == "false" {
		// Allow disabling auth for development/testing
		config.Security.RequireAuthentication = false
		log.Printf("WARNING: Admin gateway authentication disabled for development")
	}
	
	// Update authentication provider settings from environment
	if authProvider := os.Getenv("AUTH_PROVIDER"); authProvider != "" {
		// In production, this would configure the authentication provider
		log.Printf("Using authentication provider: %s", authProvider)
	}
}

// logGatewayConfiguration logs gateway configuration details
func logGatewayConfiguration(config *gateway.GatewayConfiguration) {
	log.Printf("Admin Gateway Configuration:")
	log.Printf("  - Port: %d", config.Port)
	log.Printf("  - Rate Limiting: %v", config.RateLimit.Enabled)
	if config.RateLimit.Enabled {
		log.Printf("    - Requests Per Minute: %d", config.RateLimit.RequestsPerMinute)
		log.Printf("    - Burst Size: %d", config.RateLimit.BurstSize)
		log.Printf("    - Key Extractor: %s (user-based for admin)", config.RateLimit.KeyExtractor)
	}
	log.Printf("  - CORS: %v", config.CORS.Enabled)
	if config.CORS.Enabled {
		log.Printf("    - Allowed Origins: %v", config.CORS.AllowedOrigins)
		log.Printf("    - Allowed Methods: %v", config.CORS.AllowedMethods)
		log.Printf("    - Allow Credentials: %v", config.CORS.AllowCredentials)
	}
	log.Printf("  - Authentication Required: %v", config.ShouldRequireAuth())
	if config.ShouldRequireAuth() {
		log.Printf("    - Admin authentication enforced")
		log.Printf("    - Role-based access control enabled")
	} else {
		log.Printf("    - WARNING: Admin authentication DISABLED for development")
	}
	log.Printf("  - Security Headers: %v", config.Security.SecurityHeaders.Enabled)
	if config.Security.SecurityHeaders.Enabled {
		log.Printf("    - Strict security headers for admin interface")
	}
	log.Printf("  - Cache Control: %v (disabled for admin)", config.CacheControl.Enabled)
	log.Printf("  - Service Routing:")
	log.Printf("    - Content API: %v", config.ServiceRouting.ContentAPIEnabled)
	log.Printf("    - Services API: %v", config.ServiceRouting.ServicesAPIEnabled)
	log.Printf("  - Observability:")
	log.Printf("    - Health Check: %s", config.Observability.HealthCheckPath)
	log.Printf("    - Readiness: %s", config.Observability.ReadinessPath)
	log.Printf("    - Metrics: %s", config.Observability.MetricsPath)
	log.Printf("  - Timeouts:")
	log.Printf("    - Read Timeout: %v (extended for admin operations)", config.Timeouts.ReadTimeout)
	log.Printf("    - Write Timeout: %v", config.Timeouts.WriteTimeout)
	log.Printf("    - Request Timeout: %v", config.Timeouts.RequestTimeout)
}

// handleShutdownSignals handles OS shutdown signals
func handleShutdownSignals(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	sig := <-sigChan
	log.Printf("Received shutdown signal: %v", sig)
	cancel()
}

// Helper functions

// getEnvInt returns an integer environment variable or default value
func getEnvInt(key string, defaultValue int) int {
	if str := os.Getenv(key); str != "" {
		// In production, this would use strconv.Atoi with error handling
		// For now, return default value
	}
	return defaultValue
}