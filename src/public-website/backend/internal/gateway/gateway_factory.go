package gateway

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
)

// DAPRMiddleware represents a DAPR middleware component
type DAPRMiddleware struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Version string      `json:"version"`
	Config  interface{} `json:"config"`
}

// DAPRConfig represents DAPR configuration for gateway
type DAPRConfig struct {
	Environment     string                 `json:"environment"`
	AppID          string                 `json:"app_id"`
	DAPREnabled    bool                   `json:"dapr_enabled"`
	VaultEndpoint  string                 `json:"vault_endpoint"`
	RedisEndpoint  string                 `json:"redis_endpoint"`
	Middleware     []DAPRMiddleware       `json:"middleware"`
	Components     map[string]interface{} `json:"components"`
	Configuration  map[string]interface{} `json:"configuration"`
}

// GetMiddlewareChain returns the DAPR middleware chain
func (d *DAPRConfig) GetMiddlewareChain() []DAPRMiddleware {
	return d.Middleware
}

// NewPublicGateway creates a new public gateway with environment-specific configuration
func NewPublicGateway(ctx context.Context, environment string) (http.Handler, error) {
	// Set test mode for DAPR client during testing
	if environment == "development" || environment == "testing" {
		os.Setenv("DAPR_TEST_MODE", "true")
		// Set test-friendly environment variables if not already set
		if os.Getenv("PUBLIC_GATEWAY_PORT") == "" {
			os.Setenv("PUBLIC_GATEWAY_PORT", "8080")
		}
		if os.Getenv("PUBLIC_ALLOWED_ORIGINS") == "" {
			os.Setenv("PUBLIC_ALLOWED_ORIGINS", "http://localhost:3000,https://international-center.dev,http://localhost:3001")
		}
		if os.Getenv("ENVIRONMENT") == "" {
			os.Setenv("ENVIRONMENT", environment)
		}
	}
	
	// Create DAPR client
	daprClient, err := dapr.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create DAPR client: %w", err)
	}
	
	// Create public gateway configuration
	config := NewPublicGatewayConfiguration()
	config.Environment = environment
	
	// Apply environment-specific configuration
	if err := ApplyEnvironmentConfiguration(config, environment); err != nil {
		return nil, fmt.Errorf("failed to apply environment configuration: %w", err)
	}
	
	// Create gateway service with test service proxy for testing
	var gatewayService *GatewayService
	if environment == "development" || environment == "testing" {
		gatewayService = NewGatewayServiceWithTestProxy(config, daprClient)
	} else {
		gatewayService = NewGatewayService(config, daprClient)
	}
	
	// Create the base HTTP handler
	baseHandler := gatewayService.GetHandler().CreateRouter()
	
	// For public gateways in testing environments, wrap with DAPR middleware simulator for CORS headers
	if environment == "development" || environment == "testing" {
		return NewDAPRMiddlewareSimulator(baseHandler, false), nil
	}
	
	return baseHandler, nil
}

// NewAdminGateway creates a new admin gateway with environment-specific configuration
func NewAdminGateway(ctx context.Context, environment string) (http.Handler, error) {
	// Set test mode for DAPR client during testing
	if environment == "development" || environment == "testing" {
		os.Setenv("DAPR_TEST_MODE", "true")
		// Set test-friendly environment variables if not already set
		if os.Getenv("ADMIN_GATEWAY_PORT") == "" {
			os.Setenv("ADMIN_GATEWAY_PORT", "8081")
		}
		if os.Getenv("ADMIN_ALLOWED_ORIGINS") == "" {
			os.Setenv("ADMIN_ALLOWED_ORIGINS", "http://localhost:3001,https://admin.international-center.dev,http://localhost:3000")
		}
		if os.Getenv("ENVIRONMENT") == "" {
			os.Setenv("ENVIRONMENT", environment)
		}
	}
	
	// Create DAPR client
	daprClient, err := dapr.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create DAPR client: %w", err)
	}
	
	// Create admin gateway configuration
	config := NewAdminGatewayConfiguration()
	config.Environment = environment
	
	// Apply environment-specific configuration
	if err := ApplyEnvironmentConfiguration(config, environment); err != nil {
		return nil, fmt.Errorf("failed to apply environment configuration: %w", err)
	}
	
	// Create gateway service with test service proxy for testing
	var gatewayService *GatewayService
	if environment == "development" || environment == "testing" {
		gatewayService = NewGatewayServiceWithTestProxy(config, daprClient)
	} else {
		gatewayService = NewGatewayService(config, daprClient)
	}
	
	// Create the base HTTP handler
	baseHandler := gatewayService.GetHandler().CreateRouter()
	
	// For admin gateways in testing environments, wrap with DAPR middleware simulator
	if config.IsAdmin() && (environment == "development" || environment == "testing") {
		return NewDAPRMiddlewareSimulator(baseHandler, true), nil
	}
	
	return baseHandler, nil
}

// NewGatewayDAPRConfiguration creates DAPR-specific configuration for the gateway
func NewGatewayDAPRConfiguration(environment string) (*DAPRConfig, error) {
	if environment == "" {
		return nil, fmt.Errorf("environment cannot be empty")
	}
	
	config := &DAPRConfig{
		Environment:   environment,
		AppID:        "international-center-gateway",
		DAPREnabled:  true,
		Middleware:   []DAPRMiddleware{},
		Components:   make(map[string]interface{}),
		Configuration: make(map[string]interface{}),
	}
	
	// Environment-specific DAPR configuration
	switch environment {
	case "development", "testing":
		config.VaultEndpoint = "http://vault-dev:8200"
		config.RedisEndpoint = "redis-dev:6379"
		
		config.Middleware = []DAPRMiddleware{
			{Name: "routeChecker", Type: "middleware.http.routeguard", Version: "v1", Config: map[string]interface{}{
				"allowedRoutes": []string{
					"/api/v1/services", "/api/v1/services/{id}", "/api/v1/services/slug/{slug}",
					"/api/v1/services/featured", "/api/v1/services/categories", "/api/v1/services/categories/{id}/services",
					"/api/v1/services/search", "/api/v1/news", "/api/v1/news/{id}", "/api/v1/news/slug/{slug}",
					"/api/v1/news/featured", "/api/v1/news/categories", "/api/v1/news/categories/{id}/news",
					"/api/v1/news/search", "/api/v1/research", "/api/v1/research/{id}", "/api/v1/research/slug/{slug}",
					"/api/v1/research/featured", "/api/v1/research/categories", "/api/v1/research/categories/{id}/research",
					"/api/v1/research/search", "/api/v1/research/{id}/report", "/api/v1/events", "/api/v1/events/{id}",
					"/api/v1/events/slug/{slug}", "/api/v1/events/featured", "/api/v1/events/categories",
					"/api/v1/events/categories/{id}/events", "/api/v1/events/search", "/api/v1/events/{id}/register",
					"/api/v1/events/{id}/registrations", "/api/v1/inquiries/media", "/api/v1/inquiries/business",
					"/api/v1/inquiries/donations", "/api/v1/inquiries/volunteers", "/health", "/health/ready",
				},
			}},
			{Name: "cors", Type: "middleware.http.cors", Version: "v1", Config: map[string]interface{}{
				"allowOrigins": "http://localhost:3000,https://international-center.dev,http://localhost:3001",
				"allowMethods": "GET,POST,OPTIONS",
				"allowHeaders": "Content-Type,Authorization,X-Requested-With",
				"exposeHeaders": "X-Correlation-ID",
				"maxAge": "3600",
				"allowCredentials": "false",
			}},
			{Name: "ratelimit", Type: "middleware.http.ratelimit", Version: "v1", Config: map[string]interface{}{
				"maxRequestsPerMin": 2000,
				"burstSize": 200,
				"keyExtractor": "ip",
				"store": "redis",
			}},
		}
		
		config.Components["vault"] = map[string]interface{}{
			"type":     "secretstores.local.env",
			"version":  "v1",
			"metadata": map[string]interface{}{},
		}
		config.Components["redis"] = map[string]interface{}{
			"type":    "state.redis",
			"version": "v1",
			"metadata": map[string]interface{}{
				"redisHost":     "localhost:6379",
				"redisPassword": "",
			},
		}
		
	case "staging":
		config.VaultEndpoint = "https://vault-staging.axiomcloud.dev"
		config.RedisEndpoint = "redis-staging.upstash.io:6379"
		
		config.Middleware = []DAPRMiddleware{
			{Name: "routeChecker", Type: "middleware.http.routeguard", Version: "v1", Config: map[string]interface{}{
				"allowedRoutes": []string{
					"/api/v1/services", "/api/v1/services/{id}", "/api/v1/services/slug/{slug}",
					"/api/v1/services/featured", "/api/v1/services/categories", "/api/v1/services/categories/{id}/services",
					"/api/v1/services/search", "/api/v1/news", "/api/v1/news/{id}", "/api/v1/news/slug/{slug}",
					"/api/v1/news/featured", "/api/v1/news/categories", "/api/v1/news/categories/{id}/news",
					"/api/v1/news/search", "/api/v1/research", "/api/v1/research/{id}", "/api/v1/research/slug/{slug}",
					"/api/v1/research/featured", "/api/v1/research/categories", "/api/v1/research/categories/{id}/research",
					"/api/v1/research/search", "/api/v1/research/{id}/report", "/api/v1/events", "/api/v1/events/{id}",
					"/api/v1/events/slug/{slug}", "/api/v1/events/featured", "/api/v1/events/categories",
					"/api/v1/events/categories/{id}/events", "/api/v1/events/search", "/api/v1/events/{id}/register",
					"/api/v1/events/{id}/registrations", "/api/v1/inquiries/media", "/api/v1/inquiries/business",
					"/api/v1/inquiries/donations", "/api/v1/inquiries/volunteers", "/health", "/health/ready",
				},
			}},
			{Name: "cors", Type: "middleware.http.cors", Version: "v1", Config: map[string]interface{}{
				"allowOrigins": "https://staging.international-center.dev",
				"allowMethods": "GET,POST,OPTIONS",
				"allowHeaders": "Content-Type,Authorization,X-Requested-With",
				"exposeHeaders": "X-Correlation-ID",
				"maxAge": "3600",
				"allowCredentials": "false",
			}},
			{Name: "ratelimit", Type: "middleware.http.ratelimit", Version: "v1", Config: map[string]interface{}{
				"maxRequestsPerMin": 1500,
				"burstSize": 150,
				"keyExtractor": "ip",
				"store": "redis",
			}},
		}
		
		config.Components["vault"] = map[string]interface{}{
			"type":     "secretstores.azure.keyvault",
			"version":  "v1",
			"metadata": map[string]interface{}{
				"vaultName": "staging-vault",
			},
		}
		config.Components["redis"] = map[string]interface{}{
			"type":    "state.redis",
			"version": "v1",
			"metadata": map[string]interface{}{
				"redisHost":     "staging-redis:6379",
				"redisPassword": "{vault-redis-password}",
			},
		}
		
	case "production":
		config.VaultEndpoint = "https://vault.axiomcloud.dev"
		config.RedisEndpoint = "redis-prod.upstash.io:6379"
		
		config.Middleware = []DAPRMiddleware{
			{Name: "routeChecker", Type: "middleware.http.routeguard", Version: "v1", Config: map[string]interface{}{
				"allowedRoutes": []string{
					"/api/v1/services", "/api/v1/services/{id}", "/api/v1/services/slug/{slug}",
					"/api/v1/services/featured", "/api/v1/services/categories", "/api/v1/services/categories/{id}/services",
					"/api/v1/services/search", "/api/v1/news", "/api/v1/news/{id}", "/api/v1/news/slug/{slug}",
					"/api/v1/news/featured", "/api/v1/news/categories", "/api/v1/news/categories/{id}/news",
					"/api/v1/news/search", "/api/v1/research", "/api/v1/research/{id}", "/api/v1/research/slug/{slug}",
					"/api/v1/research/featured", "/api/v1/research/categories", "/api/v1/research/categories/{id}/research",
					"/api/v1/research/search", "/api/v1/research/{id}/report", "/api/v1/events", "/api/v1/events/{id}",
					"/api/v1/events/slug/{slug}", "/api/v1/events/featured", "/api/v1/events/categories",
					"/api/v1/events/categories/{id}/events", "/api/v1/events/search", "/api/v1/events/{id}/register",
					"/api/v1/events/{id}/registrations", "/api/v1/inquiries/media", "/api/v1/inquiries/business",
					"/api/v1/inquiries/donations", "/api/v1/inquiries/volunteers", "/health", "/health/ready",
				},
			}},
			{Name: "bearer", Type: "middleware.http.bearer", Version: "v1", Config: map[string]interface{}{
				"authHeader": "Authorization",
				"token": "{vault-jwt-secret}",
				"issuer": "https://auth.production.international-center.dev/application/o/gateway/",
				"audience": "international-center",
				"clockSkew": "5m",
			}},
			{Name: "oauth2", Type: "middleware.http.oauth2", Version: "v1", Config: map[string]interface{}{
				"clientId": "{vault-oauth2-client-id}",
				"clientSecret": "{vault-oauth2-client-secret}",
				"scopes": []string{"openid", "profile", "email"},
				"authURL": "https://auth.production.international-center.dev/application/o/authorize/",
				"tokenURL": "https://auth.production.international-center.dev/application/o/token/",
				"redirectURL": "https://international-center.dev/auth/callback",
				"authStyle": "2",
			}},
			{Name: "opa", Type: "middleware.http.opa", Version: "v1", Config: map[string]interface{}{
				"defaultStatus": "403",
				"includedHeaders": []string{"authorization", "content-type"},
				"rego": "package http_authz\n\ndefault allow = false\n\nallow {\n  input.method == \"GET\"\n  startswith(input.path, \"/api/v1/\")\n  token_valid\n}\n\nallow {\n  input.method == \"POST\"\n  startswith(input.path, \"/admin/api/v1/\")\n  token_valid\n  user_has_role(\"admin\")\n}\n\ntoken_valid {\n  jwt.valid_es256(input.headers.authorization, \"{vault-jwt-secret}\")\n}\n\nuser_has_role(role) {\n  [_, payload, _] := io.jwt.decode(input.headers.authorization)\n  payload.roles[_] == role\n}",
			}},
			{Name: "cors", Type: "middleware.http.cors", Version: "v1", Config: map[string]interface{}{
				"allowOrigins": "https://international-center.dev",
				"allowMethods": "GET,POST,PUT,DELETE,OPTIONS",
				"allowHeaders": "Content-Type,Authorization,X-Requested-With",
				"exposeHeaders": "X-Correlation-ID",
				"maxAge": "86400",
				"allowCredentials": "true",
			}},
			{Name: "ratelimit", Type: "middleware.http.ratelimit", Version: "v1", Config: map[string]interface{}{
				"maxRequestsPerMin": 1000,
				"burstSize": 100,
				"keyExtractor": "ip",
				"store": "redis",
				"headers": map[string]interface{}{
					"X-RateLimit-Limit": "1000",
					"X-RateLimit-Remaining": "{remaining}",
					"X-RateLimit-Reset": "{reset}",
				},
			}},
		}
		
		config.Components["vault"] = map[string]interface{}{
			"type":     "secretstores.azure.keyvault",
			"version":  "v1",
			"metadata": map[string]interface{}{
				"vaultName": "production-vault",
			},
		}
		config.Components["redis"] = map[string]interface{}{
			"type":    "state.redis",
			"version": "v1",
			"metadata": map[string]interface{}{
				"redisHost":     "production-redis:6379",
				"redisPassword": "{vault-redis-password}",
				"enableTLS":     true,
			},
		}
		
	default:
		return nil, fmt.Errorf("unsupported environment: %s", environment)
	}
	
	// Add authentication configuration
	config.Configuration["authentication"] = map[string]interface{}{
		"enabled":    true,
		"provider":   "authentik",
		"jwksURL":    fmt.Sprintf("https://auth.%s.international-center.dev/application/o/gateway/.well-known/jwks.json", environment),
		"audience":   "international-center",
		"issuer":     fmt.Sprintf("https://auth.%s.international-center.dev/application/o/gateway/", environment),
	}
	
	// Add observability configuration
	config.Configuration["observability"] = map[string]interface{}{
		"tracing": map[string]interface{}{
			"enabled": true,
			"sampler": map[string]interface{}{
				"type":  "probability",
				"value": getTracingSampleRate(environment),
			},
		},
		"metrics": map[string]interface{}{
			"enabled": true,
			"path":    "/metrics",
		},
	}
	
	return config, nil
}

// ApplyEnvironmentConfiguration applies environment-specific overrides to gateway configuration
func ApplyEnvironmentConfiguration(config *GatewayConfiguration, environment string) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	if environment == "" {
		return fmt.Errorf("environment cannot be empty")
	}
	
	// Apply environment-specific timeout adjustments
	switch environment {
	case "development", "testing":
		// Shorter timeouts for development/testing
		config.Timeouts.ReadTimeout = 10 * time.Second
		config.Timeouts.WriteTimeout = 10 * time.Second
		config.Timeouts.IdleTimeout = 30 * time.Second
		config.Timeouts.RequestTimeout = 10 * time.Second
		config.Timeouts.ShutdownTimeout = 5 * time.Second
		
		// More permissive rate limiting for testing
		if config.RateLimit.Enabled {
			if config.IsPublic() {
				config.RateLimit.RequestsPerMinute = 2000 // Higher for public testing
				config.RateLimit.BurstSize = 200
			} else {
				config.RateLimit.RequestsPerMinute = 200 // Higher for admin testing
				config.RateLimit.BurstSize = 40
			}
		}
		
	case "staging":
		// Moderate timeouts for staging
		config.Timeouts.ReadTimeout = 30 * time.Second
		config.Timeouts.WriteTimeout = 30 * time.Second
		config.Timeouts.IdleTimeout = 60 * time.Second
		config.Timeouts.RequestTimeout = 30 * time.Second
		config.Timeouts.ShutdownTimeout = 15 * time.Second
		
		// Production-like rate limiting for staging
		if config.RateLimit.Enabled {
			if config.IsPublic() {
				config.RateLimit.RequestsPerMinute = 1500
				config.RateLimit.BurstSize = 150
			} else {
				config.RateLimit.RequestsPerMinute = 150
				config.RateLimit.BurstSize = 30
			}
		}
		
	case "production":
		// Production timeouts (more conservative)
		config.Timeouts.ReadTimeout = 60 * time.Second
		config.Timeouts.WriteTimeout = 60 * time.Second
		config.Timeouts.IdleTimeout = 120 * time.Second
		config.Timeouts.RequestTimeout = 60 * time.Second
		config.Timeouts.ShutdownTimeout = 30 * time.Second
		
		// Strict rate limiting for production
		if config.RateLimit.Enabled {
			if config.IsPublic() {
				config.RateLimit.RequestsPerMinute = 1000
				config.RateLimit.BurstSize = 100
			} else {
				config.RateLimit.RequestsPerMinute = 100
				config.RateLimit.BurstSize = 20
			}
		}
		
		// Enhanced security for production
		if config.Security.SecurityHeaders.Enabled {
			config.Security.SecurityHeaders.StrictTransportSecurity = "max-age=31536000; includeSubDomains; preload"
			config.Security.SecurityHeaders.ContentSecurityPolicy = "default-src 'self'; object-src 'none'; base-uri 'self'"
		}
		
	default:
		return fmt.Errorf("unsupported environment: %s", environment)
	}
	
	// Apply environment-specific observability settings
	if config.Observability.Enabled {
		switch environment {
		case "development", "testing":
			config.Observability.TracingEnabled = true
			config.Observability.MetricsEnabled = true
			config.Observability.LoggingEnabled = true
			
		case "staging":
			config.Observability.TracingEnabled = true
			config.Observability.MetricsEnabled = true
			config.Observability.LoggingEnabled = true
			
		case "production":
			config.Observability.TracingEnabled = true
			config.Observability.MetricsEnabled = true
			config.Observability.LoggingEnabled = true
		}
	}
	
	return nil
}

// Helper function to get tracing sample rate based on environment
func getTracingSampleRate(environment string) float64 {
	switch environment {
	case "development", "testing":
		return 1.0 // 100% sampling for development
	case "staging":
		return 0.1 // 10% sampling for staging
	case "production":
		return 0.01 // 1% sampling for production
	default:
		return 0.1
	}
}