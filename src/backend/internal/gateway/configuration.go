package gateway

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// GatewayType defines the type of gateway (public or admin)
type GatewayType string

const (
	GatewayTypePublic GatewayType = "public"
	GatewayTypeAdmin  GatewayType = "admin"
)

// GatewayConfiguration holds gateway-specific configuration
type GatewayConfiguration struct {
	Name         string      `json:"name"`
	Type         GatewayType `json:"type"`
	Port         int         `json:"port"`
	Environment  string      `json:"environment"`
	Version      string      `json:"version"`
	
	// Security configuration
	Security SecurityConfig `json:"security"`
	
	// Rate limiting configuration
	RateLimit RateLimitConfig `json:"rate_limit"`
	
	// CORS configuration
	CORS CORSConfig `json:"cors"`
	
	// Cache configuration
	CacheControl CacheControlConfig `json:"cache_control"`
	
	// Service routing configuration
	ServiceRouting ServiceRoutingConfig `json:"service_routing"`
	
	// Timeout configuration
	Timeouts TimeoutConfig `json:"timeouts"`
	
	// Observability configuration
	Observability ObservabilityConfig `json:"observability"`
}

// SecurityConfig defines security-related configuration
type SecurityConfig struct {
	RequireAuthentication bool     `json:"require_authentication"`
	AllowedOrigins       []string `json:"allowed_origins"`
	SecurityHeaders      SecurityHeadersConfig `json:"security_headers"`
}

// SecurityHeadersConfig defines security headers configuration
type SecurityHeadersConfig struct {
	Enabled                    bool   `json:"enabled"`
	ContentTypeOptions         string `json:"content_type_options"`
	FrameOptions              string `json:"frame_options"`
	XSSProtection             string `json:"xss_protection"`
	StrictTransportSecurity   string `json:"strict_transport_security"`
	ContentSecurityPolicy     string `json:"content_security_policy"`
	ReferrerPolicy            string `json:"referrer_policy"`
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool          `json:"enabled"`
	RequestsPerMinute int           `json:"requests_per_minute"`
	BurstSize         int           `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
	KeyExtractor      string        `json:"key_extractor"` // "ip" or "user"
	BackingStore      string        `json:"backing_store"` // "redis" or "memory"
}

// CORSConfig defines CORS configuration
type CORSConfig struct {
	Enabled          bool     `json:"enabled"`
	AllowedOrigins   []string `json:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

// CacheControlConfig defines cache control configuration
type CacheControlConfig struct {
	Enabled bool `json:"enabled"`
	MaxAge  int  `json:"max_age"`
}

// ServiceRoutingConfig defines service routing configuration
type ServiceRoutingConfig struct {
	ContentAPIEnabled  bool   `json:"content_api_enabled"`
	ServicesAPIEnabled bool   `json:"services_api_enabled"`
	NewsAPIEnabled     bool   `json:"news_api_enabled"`
	HealthCheckPath    string `json:"health_check_path"`
	MetricsPath        string `json:"metrics_path"`
}

// TimeoutConfig defines timeout configuration
type TimeoutConfig struct {
	ReadTimeout       time.Duration `json:"read_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout"`
	IdleTimeout       time.Duration `json:"idle_timeout"`
	RequestTimeout    time.Duration `json:"request_timeout"`
	ShutdownTimeout   time.Duration `json:"shutdown_timeout"`
}

// ObservabilityConfig defines observability configuration
type ObservabilityConfig struct {
	Enabled           bool   `json:"enabled"`
	MetricsEnabled    bool   `json:"metrics_enabled"`
	TracingEnabled    bool   `json:"tracing_enabled"`
	LoggingEnabled    bool   `json:"logging_enabled"`
	HealthCheckPath   string `json:"health_check_path"`
	ReadinessPath     string `json:"readiness_path"`
	MetricsPath       string `json:"metrics_path"`
}

// NewPublicGatewayConfiguration creates configuration for public gateway
func NewPublicGatewayConfiguration() *GatewayConfiguration {
	port := os.Getenv("PUBLIC_GATEWAY_PORT")
	if port == "" {
		log.Fatalf("PUBLIC_GATEWAY_PORT environment variable is required")
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("PUBLIC_GATEWAY_PORT must be a valid integer: %v", err)
	}

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		log.Fatalf("ENVIRONMENT environment variable is required")
	}

	allowedOrigins := os.Getenv("PUBLIC_ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		log.Fatalf("PUBLIC_ALLOWED_ORIGINS environment variable is required")
	}

	return &GatewayConfiguration{
		Name:        "public-gateway",
		Type:        GatewayTypePublic,
		Port:        portInt,
		Environment: environment,
		Version:     "1.0.0",
		
		Security: SecurityConfig{
			RequireAuthentication: false, // Public gateway allows anonymous access
			AllowedOrigins:        strings.Split(allowedOrigins, ","),
			SecurityHeaders: SecurityHeadersConfig{
				Enabled:                  true,
				ContentTypeOptions:       "nosniff",
				FrameOptions:            "DENY",
				XSSProtection:           "1; mode=block",
				StrictTransportSecurity: "max-age=31536000; includeSubDomains",
				ContentSecurityPolicy:   "default-src 'self'; object-src 'none'",
				ReferrerPolicy:          "strict-origin-when-cross-origin",
			},
		},
		
		RateLimit: RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 1000, // Higher limit for public access
			BurstSize:         100,
			WindowSize:        time.Minute,
			KeyExtractor:      "ip",
			BackingStore:      "redis",
		},
		
		CORS: CORSConfig{
			Enabled:          true,
			AllowedOrigins:   strings.Split(allowedOrigins, ","),
			AllowedMethods:   []string{"GET", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
			ExposedHeaders:   []string{"X-Correlation-ID"},
			AllowCredentials: false,
			MaxAge:           3600,
		},
		
		CacheControl: CacheControlConfig{
			Enabled: true,
			MaxAge:  300, // 5 minutes for public content
		},
		
		ServiceRouting: ServiceRoutingConfig{
			ContentAPIEnabled:  true,
			ServicesAPIEnabled: true,
			NewsAPIEnabled:     true,
			HealthCheckPath:    "/health",
			MetricsPath:        "/metrics",
		},
		
		Timeouts: TimeoutConfig{
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     60 * time.Second,
			RequestTimeout:  30 * time.Second,
			ShutdownTimeout: 15 * time.Second,
		},
		
		Observability: ObservabilityConfig{
			Enabled:         true,
			MetricsEnabled:  true,
			TracingEnabled:  true,
			LoggingEnabled:  true,
			HealthCheckPath: "/health",
			ReadinessPath:   "/ready",
			MetricsPath:     "/metrics",
		},
	}
}

// NewAdminGatewayConfiguration creates configuration for admin gateway
func NewAdminGatewayConfiguration() *GatewayConfiguration {
	port := os.Getenv("ADMIN_GATEWAY_PORT")
	if port == "" {
		log.Fatalf("ADMIN_GATEWAY_PORT environment variable is required")
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("ADMIN_GATEWAY_PORT must be a valid integer: %v", err)
	}

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		log.Fatalf("ENVIRONMENT environment variable is required")
	}

	allowedOrigins := os.Getenv("ADMIN_ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		log.Fatalf("ADMIN_ALLOWED_ORIGINS environment variable is required")
	}

	return &GatewayConfiguration{
		Name:        "admin-gateway",
		Type:        GatewayTypeAdmin,
		Port:        portInt,
		Environment: environment,
		Version:     "1.0.0",
		
		Security: SecurityConfig{
			RequireAuthentication: true, // Admin gateway requires authentication
			AllowedOrigins:        strings.Split(allowedOrigins, ","),
			SecurityHeaders: SecurityHeadersConfig{
				Enabled:                  true,
				ContentTypeOptions:       "nosniff",
				FrameOptions:            "DENY",
				XSSProtection:           "1; mode=block",
				StrictTransportSecurity: "max-age=31536000; includeSubDomains",
				ContentSecurityPolicy:   "default-src 'self'; object-src 'none'",
				ReferrerPolicy:          "strict-origin-when-cross-origin",
			},
		},
		
		RateLimit: RateLimitConfig{
			Enabled:           true,
			RequestsPerMinute: 100, // Lower limit for admin access
			BurstSize:         20,
			WindowSize:        time.Minute,
			KeyExtractor:      "user", // Rate limit by user for admin
			BackingStore:      "redis",
		},
		
		CORS: CORSConfig{
			Enabled:          true,
			AllowedOrigins:   strings.Split(allowedOrigins, ","),
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With", "X-User-ID"},
			ExposedHeaders:   []string{"X-Correlation-ID"},
			AllowCredentials: true,
			MaxAge:           3600,
		},
		
		CacheControl: CacheControlConfig{
			Enabled: false, // No caching for admin interface
			MaxAge:  0,
		},
		
		ServiceRouting: ServiceRoutingConfig{
			ContentAPIEnabled:  true,
			ServicesAPIEnabled: true,
			NewsAPIEnabled:     true,
			HealthCheckPath:    "/health",
			MetricsPath:        "/metrics",
		},
		
		Timeouts: TimeoutConfig{
			ReadTimeout:     60 * time.Second, // Longer timeouts for admin operations
			WriteTimeout:    60 * time.Second,
			IdleTimeout:     120 * time.Second,
			RequestTimeout:  60 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
		
		Observability: ObservabilityConfig{
			Enabled:         true,
			MetricsEnabled:  true,
			TracingEnabled:  true,
			LoggingEnabled:  true,
			HealthCheckPath: "/health",
			ReadinessPath:   "/ready",
			MetricsPath:     "/metrics",
		},
	}
}

// IsPublic returns true if this is a public gateway
func (c *GatewayConfiguration) IsPublic() bool {
	return c.Type == GatewayTypePublic
}

// IsAdmin returns true if this is an admin gateway
func (c *GatewayConfiguration) IsAdmin() bool {
	return c.Type == GatewayTypeAdmin
}

// GetListenAddress returns the listen address for the gateway
func (c *GatewayConfiguration) GetListenAddress() string {
	return fmt.Sprintf(":%d", c.Port)
}

// ShouldRequireAuth returns true if authentication is required
func (c *GatewayConfiguration) ShouldRequireAuth() bool {
	return c.Security.RequireAuthentication
}