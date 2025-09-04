package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
)

// GatewayService provides the main gateway service implementation
type GatewayService struct {
	config        *GatewayConfiguration
	daprClient    *dapr.Client
	serviceProxy  *ServiceProxy
	middleware    *Middleware
	handler       *GatewayHandler
	server        *http.Server
}

// NewGatewayService creates a new gateway service
func NewGatewayService(config *GatewayConfiguration, daprClient *dapr.Client) *GatewayService {
	// Initialize service proxy
	serviceProxy := NewServiceProxy(daprClient, config)
	
	// Initialize middleware
	middleware := NewMiddleware(config)
	
	// Initialize handler
	handler := NewGatewayHandler(config, serviceProxy, middleware)
	
	// Create HTTP server
	server := &http.Server{
		Addr:         config.GetListenAddress(),
		Handler:      handler.CreateRouter(),
		ReadTimeout:  config.Timeouts.ReadTimeout,
		WriteTimeout: config.Timeouts.WriteTimeout,
		IdleTimeout:  config.Timeouts.IdleTimeout,
	}
	
	return &GatewayService{
		config:       config,
		daprClient:   daprClient,
		serviceProxy: serviceProxy,
		middleware:   middleware,
		handler:      handler,
		server:       server,
	}
}

// NewPublicGatewayService creates a new public gateway service
func NewPublicGatewayService(daprClient *dapr.Client) *GatewayService {
	config := NewPublicGatewayConfiguration()
	return NewGatewayService(config, daprClient)
}

// NewAdminGatewayService creates a new admin gateway service
func NewAdminGatewayService(daprClient *dapr.Client) *GatewayService {
	config := NewAdminGatewayConfiguration()
	return NewGatewayService(config, daprClient)
}

// Start starts the gateway service
func (g *GatewayService) Start(ctx context.Context) error {
	// Log startup information
	fmt.Printf("Starting %s gateway service on %s\n", g.config.Name, g.config.GetListenAddress())
	fmt.Printf("Gateway Type: %s\n", g.config.Type)
	fmt.Printf("Environment: %s\n", g.config.Environment)
	fmt.Printf("Version: %s\n", g.config.Version)
	
	// Print configuration details
	g.logConfiguration()
	
	// Validate configuration
	if err := g.validateConfiguration(); err != nil {
		return fmt.Errorf("invalid gateway configuration: %w", err)
	}
	
	// Check Dapr connectivity
	if err := g.checkDaprConnectivity(ctx); err != nil {
		return fmt.Errorf("Dapr connectivity check failed: %w", err)
	}
	
	// Start HTTP server in goroutine
	go func() {
		fmt.Printf("Gateway %s listening on %s\n", g.config.Name, g.config.GetListenAddress())
		if err := g.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Gateway server error: %v\n", err)
		}
	}()
	
	// Wait for context cancellation
	<-ctx.Done()
	
	// Shutdown gracefully
	return g.Shutdown()
}

// Shutdown gracefully shuts down the gateway service
func (g *GatewayService) Shutdown() error {
	fmt.Printf("Shutting down %s gateway service...\n", g.config.Name)
	
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), g.config.Timeouts.ShutdownTimeout)
	defer cancel()
	
	// Shutdown HTTP server
	if err := g.server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Gateway server shutdown error: %v\n", err)
		return err
	}
	
	fmt.Printf("Gateway %s shut down successfully\n", g.config.Name)
	return nil
}

// HealthCheck performs a health check of the gateway service
func (g *GatewayService) HealthCheck(ctx context.Context) error {
	// Check Dapr connectivity
	if err := g.checkDaprConnectivity(ctx); err != nil {
		return fmt.Errorf("Dapr connectivity failed: %w", err)
	}
	
	// Check backend service health via service proxy
	if err := g.serviceProxy.HealthCheck(ctx); err != nil {
		return fmt.Errorf("backend services health check failed: %w", err)
	}
	
	return nil
}

// GetMetrics returns gateway service metrics
func (g *GatewayService) GetMetrics(ctx context.Context) (map[string]interface{}, error) {
	// Get service metrics from proxy
	serviceMetrics, err := g.serviceProxy.GetServiceMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get service metrics: %w", err)
	}
	
	// Add gateway-specific metrics
	metrics := map[string]interface{}{
		"gateway": map[string]interface{}{
			"name":        g.config.Name,
			"type":        g.config.Type,
			"version":     g.config.Version,
			"environment": g.config.Environment,
			"uptime":      time.Now().UTC(),
			"server": map[string]interface{}{
				"addr":          g.server.Addr,
				"read_timeout":  g.config.Timeouts.ReadTimeout,
				"write_timeout": g.config.Timeouts.WriteTimeout,
				"idle_timeout":  g.config.Timeouts.IdleTimeout,
			},
			"configuration": map[string]interface{}{
				"rate_limit_enabled":   g.config.RateLimit.Enabled,
				"cors_enabled":         g.config.CORS.Enabled,
				"auth_required":        g.config.ShouldRequireAuth(),
				"content_api_enabled":  g.config.ServiceRouting.ContentAPIEnabled,
				"services_api_enabled": g.config.ServiceRouting.ServicesAPIEnabled,
				"news_api_enabled":     g.config.ServiceRouting.NewsAPIEnabled,
			},
		},
		"services": serviceMetrics,
	}
	
	return metrics, nil
}

// GetConfiguration returns the gateway configuration
func (g *GatewayService) GetConfiguration() *GatewayConfiguration {
	return g.config
}

// GetServiceProxy returns the service proxy
func (g *GatewayService) GetServiceProxy() *ServiceProxy {
	return g.serviceProxy
}

// GetHandler returns the gateway handler
func (g *GatewayService) GetHandler() *GatewayHandler {
	return g.handler
}

// Private helper methods

// validateConfiguration validates the gateway configuration
func (g *GatewayService) validateConfiguration() error {
	if g.config.Name == "" {
		return fmt.Errorf("gateway name cannot be empty")
	}
	
	if g.config.Port <= 0 || g.config.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", g.config.Port)
	}
	
	if g.config.Type != GatewayTypePublic && g.config.Type != GatewayTypeAdmin {
		return fmt.Errorf("invalid gateway type: %s", g.config.Type)
	}
	
	if g.config.Version == "" {
		return fmt.Errorf("gateway version cannot be empty")
	}
	
	// Validate rate limiting configuration
	if g.config.RateLimit.Enabled {
		if g.config.RateLimit.RequestsPerMinute <= 0 {
			return fmt.Errorf("invalid requests per minute: %d", g.config.RateLimit.RequestsPerMinute)
		}
		
		if g.config.RateLimit.BurstSize <= 0 {
			return fmt.Errorf("invalid burst size: %d", g.config.RateLimit.BurstSize)
		}
		
		if g.config.RateLimit.KeyExtractor != "ip" && g.config.RateLimit.KeyExtractor != "user" {
			return fmt.Errorf("invalid key extractor: %s", g.config.RateLimit.KeyExtractor)
		}
	}
	
	// Validate CORS configuration
	if g.config.CORS.Enabled {
		if len(g.config.CORS.AllowedOrigins) == 0 {
			return fmt.Errorf("CORS enabled but no allowed origins specified")
		}
		
		if len(g.config.CORS.AllowedMethods) == 0 {
			return fmt.Errorf("CORS enabled but no allowed methods specified")
		}
	}
	
	// Validate timeout configuration
	if g.config.Timeouts.ReadTimeout <= 0 {
		return fmt.Errorf("invalid read timeout: %v", g.config.Timeouts.ReadTimeout)
	}
	
	if g.config.Timeouts.WriteTimeout <= 0 {
		return fmt.Errorf("invalid write timeout: %v", g.config.Timeouts.WriteTimeout)
	}
	
	return nil
}

// checkDaprConnectivity checks Dapr connectivity
func (g *GatewayService) checkDaprConnectivity(ctx context.Context) error {
	// Create a context with timeout for connectivity check
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	// Check if Dapr client is healthy
	if !g.daprClient.IsHealthy(checkCtx) {
		return fmt.Errorf("Dapr client is not healthy")
	}
	
	return nil
}

// logConfiguration logs gateway configuration details
func (g *GatewayService) logConfiguration() {
	fmt.Printf("Gateway Configuration:\n")
	fmt.Printf("  - Rate Limiting: %v", g.config.RateLimit.Enabled)
	if g.config.RateLimit.Enabled {
		fmt.Printf(" (%d req/min, burst: %d)", g.config.RateLimit.RequestsPerMinute, g.config.RateLimit.BurstSize)
	}
	fmt.Printf("\n")
	
	fmt.Printf("  - CORS: %v", g.config.CORS.Enabled)
	if g.config.CORS.Enabled {
		fmt.Printf(" (origins: %v)", g.config.CORS.AllowedOrigins)
	}
	fmt.Printf("\n")
	
	fmt.Printf("  - Authentication Required: %v\n", g.config.ShouldRequireAuth())
	fmt.Printf("  - Content API: %v\n", g.config.ServiceRouting.ContentAPIEnabled)
	fmt.Printf("  - Services API: %v\n", g.config.ServiceRouting.ServicesAPIEnabled)
	fmt.Printf("  - News API: %v\n", g.config.ServiceRouting.NewsAPIEnabled)
	fmt.Printf("  - Security Headers: %v\n", g.config.Security.SecurityHeaders.Enabled)
	fmt.Printf("  - Observability: %v\n", g.config.Observability.Enabled)
	
	if g.config.Observability.Enabled {
		fmt.Printf("    - Health Check: %s\n", g.config.Observability.HealthCheckPath)
		fmt.Printf("    - Readiness: %s\n", g.config.Observability.ReadinessPath)
		fmt.Printf("    - Metrics: %s\n", g.config.Observability.MetricsPath)
	}
	
	fmt.Printf("  - Timeouts:\n")
	fmt.Printf("    - Read: %v\n", g.config.Timeouts.ReadTimeout)
	fmt.Printf("    - Write: %v\n", g.config.Timeouts.WriteTimeout)
	fmt.Printf("    - Request: %v\n", g.config.Timeouts.RequestTimeout)
	fmt.Printf("    - Shutdown: %v\n", g.config.Timeouts.ShutdownTimeout)
}