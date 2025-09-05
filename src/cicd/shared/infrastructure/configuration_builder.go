package infrastructure

import (
	"fmt"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

// DaprConfig holds Dapr-specific configuration
type DaprConfig struct {
	HTTPPort        int
	GRPCPort        int
	PlacementPort   int
	ComponentsPath  string
	LogLevel        string
	Version         string
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	Name         string
	Port         int
	Host         string
	HealthPath   string
	Version      string
	Environment  string
}

// VaultConfig holds Vault connection configuration  
type VaultConfig struct {
	Address  string
	Token    string
	SkipTLS  bool
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	MetricsEnabled bool
	TracingEnabled bool
	LogLevel       string
	GrafanaURL     string
	PrometheusURL  string
	LokiURL        string
}

// ConfigurationBuilder provides centralized configuration management
type ConfigurationBuilder struct {
	config      *config.Config
	environment string
}

// NewConfigurationBuilder creates a new configuration builder
func NewConfigurationBuilder(config *config.Config, environment string) *ConfigurationBuilder {
	return &ConfigurationBuilder{
		config:      config,
		environment: environment,
	}
}

// BuildDatabaseConfig creates standardized database configuration
func (cb *ConfigurationBuilder) BuildDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     cb.config.Get("postgres_host", "postgres"),
		Port:     cb.config.RequireInt("postgres_port"),
		Database: cb.config.Get("postgres_db", fmt.Sprintf("%s_db", cb.environment)),
		User:     cb.config.Get("postgres_user", "postgres"),
		Password: cb.config.RequireSecret("postgres_password").AsStringOutput(),
		SSLMode:  cb.config.Get("postgres_sslmode", "disable"),
	}
}

// BuildRedisConfig creates standardized Redis configuration
func (cb *ConfigurationBuilder) BuildRedisConfig() RedisConfig {
	return RedisConfig{
		Host:     cb.config.Get("redis_host", "redis"),
		Port:     cb.config.RequireInt("redis_port"),
		Password: cb.config.Get("redis_password", ""),
	}
}

// BuildVaultConfig creates standardized Vault configuration
func (cb *ConfigurationBuilder) BuildVaultConfig() VaultConfig {
	return VaultConfig{
		Address: cb.config.Get("vault_addr", "http://vault:8200"),
		Token:   cb.config.Get("vault_token", "dev-root-token"),
		SkipTLS: cb.config.GetBool("vault_skip_tls", true),
	}
}

// BuildServiceEnvironment creates standardized environment variables for services
func (cb *ConfigurationBuilder) BuildServiceEnvironment(service ServiceConfig, db DatabaseConfig, redis RedisConfig, dapr DaprConfig) map[string]pulumi.StringInput {
	environment := make(map[string]pulumi.StringInput)
	
	// Database configuration
	environment["DATABASE_URL"] = pulumi.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.Database, db.SSLMode)
	environment["DATABASE_HOST"] = pulumi.String(db.Host)
	environment["DATABASE_PORT"] = pulumi.Sprintf("%d", db.Port)
	environment["DATABASE_NAME"] = pulumi.String(db.Database)
	environment["DATABASE_USER"] = pulumi.String(db.User)
	
	// Redis configuration
	environment["REDIS_URL"] = pulumi.Sprintf("redis://%s:%d", redis.Host, redis.Port)
	environment["REDIS_HOST"] = pulumi.String(redis.Host)
	environment["REDIS_PORT"] = pulumi.Sprintf("%d", redis.Port)
	if redis.Password != "" {
		environment["REDIS_PASSWORD"] = pulumi.String(redis.Password)
	}
	
	// Dapr configuration
	environment["DAPR_HTTP_PORT"] = pulumi.Sprintf("%d", dapr.HTTPPort)
	environment["DAPR_GRPC_PORT"] = pulumi.Sprintf("%d", dapr.GRPCPort)
	environment["DAPR_PLACEMENT_PORT"] = pulumi.Sprintf("%d", dapr.PlacementPort)
	
	// Service configuration
	environment["SERVICE_NAME"] = pulumi.String(service.Name)
	environment["SERVICE_PORT"] = pulumi.Sprintf("%d", service.Port)
	environment["SERVICE_HOST"] = pulumi.String(service.Host)
	environment["HEALTH_PATH"] = pulumi.String(service.HealthPath)
	
	// Application configuration
	environment["APP_VERSION"] = pulumi.String(service.Version)
	environment["ENVIRONMENT"] = pulumi.String(service.Environment)
	environment["LOG_LEVEL"] = pulumi.String(cb.config.Get("log_level", "info"))
	
	return environment
}

// BuildGatewayEnvironment creates environment variables specific to gateway services
func (cb *ConfigurationBuilder) BuildGatewayEnvironment(gatewayType string, service ServiceConfig, dapr DaprConfig, upstreamServices map[string]string) map[string]pulumi.StringInput {
	environment := make(map[string]pulumi.StringInput)
	
	// Gateway configuration
	environment["GATEWAY_TYPE"] = pulumi.String(gatewayType)
	environment["GATEWAY_PORT"] = pulumi.Sprintf("%d", service.Port)
	environment[fmt.Sprintf("%s_PORT", gatewayType)] = pulumi.Sprintf("%d", service.Port)
	
	// CORS configuration based on gateway type
	if gatewayType == "PUBLIC_GATEWAY" {
		environment["PUBLIC_ALLOWED_ORIGINS"] = pulumi.String("*")
		environment["CORS_ENABLED"] = pulumi.String("true")
	} else if gatewayType == "ADMIN_GATEWAY" {
		environment["ADMIN_ALLOWED_ORIGINS"] = pulumi.String("http://localhost:3000")
		environment["CORS_ENABLED"] = pulumi.String("true")
	}
	
	// Dapr configuration
	environment["DAPR_HTTP_PORT"] = pulumi.Sprintf("%d", dapr.HTTPPort)
	environment["DAPR_GRPC_PORT"] = pulumi.Sprintf("%d", dapr.GRPCPort)
	
	// Upstream service URLs (using Dapr service invocation)
	for serviceName, endpoint := range upstreamServices {
		envKey := fmt.Sprintf("%s_URL", serviceName)
		environment[envKey] = pulumi.String(endpoint)
	}
	
	// Standard configuration
	environment["SERVICE_NAME"] = pulumi.String(service.Name)
	environment["APP_VERSION"] = pulumi.String(service.Version)
	environment["ENVIRONMENT"] = pulumi.String(service.Environment)
	environment["LOG_LEVEL"] = pulumi.String(cb.config.Get("log_level", "info"))
	
	return environment
}

// BuildWebsiteEnvironment creates environment variables for website containers
func (cb *ConfigurationBuilder) BuildWebsiteEnvironment(service ServiceConfig, apiEndpoints map[string]string) map[string]pulumi.StringInput {
	environment := make(map[string]pulumi.StringInput)
	
	// Website configuration
	environment["NODE_ENV"] = pulumi.String(cb.environment)
	environment["WEBSITE_PORT"] = pulumi.Sprintf("%d", service.Port)
	environment["WEBSITE_HOST"] = pulumi.String(service.Host)
	
	// API endpoint configuration
	for serviceName, endpoint := range apiEndpoints {
		envKey := fmt.Sprintf("API_%s_URL", serviceName)
		environment[envKey] = pulumi.String(endpoint)
	}
	
	// Development features
	if cb.environment == "development" {
		environment["HOT_RELOAD_ENABLED"] = pulumi.String("true")
		environment["DEBUG_MODE"] = pulumi.String("true")
		environment["DEV_TOOLS_ENABLED"] = pulumi.String("true")
	}
	
	// Standard configuration
	environment["SERVICE_NAME"] = pulumi.String(service.Name)
	environment["APP_VERSION"] = pulumi.String(service.Version)
	environment["ENVIRONMENT"] = pulumi.String(service.Environment)
	
	return environment
}

// BuildHealthCheckConfig creates standardized health check configuration
func (cb *ConfigurationBuilder) BuildHealthCheckConfig(service ServiceConfig) HealthCheckConfig {
	// Default health check settings based on environment
	interval := 30 * time.Second
	timeout := 10 * time.Second
	retries := 3
	
	// Adjust for environment
	if cb.environment == "development" {
		interval = 15 * time.Second
		timeout = 5 * time.Second
	} else if cb.environment == "production" {
		interval = 60 * time.Second
		timeout = 15 * time.Second
		retries = 5
	}
	
	return HealthCheckConfig{
		Path:     service.HealthPath,
		Port:     service.Port,
		Interval: interval,
		Timeout:  timeout,
		Retries:  retries,
	}
}

// BuildDaprSidecarConfig creates standardized Dapr sidecar configuration
func (cb *ConfigurationBuilder) BuildDaprSidecarConfig(serviceName string, appPort int, httpPort int, grpcPort int) DaprConfig {
	return DaprConfig{
		HTTPPort:       httpPort,
		GRPCPort:       grpcPort,
		PlacementPort:  50005,
		ComponentsPath: "/components",
		LogLevel:       cb.config.Get("dapr_log_level", "info"),
		Version:        cb.config.Get("dapr_version", "1.12.0"),
	}
}

// BuildNetworkConfig creates standardized network configuration
func (cb *ConfigurationBuilder) BuildNetworkConfig(component string) NetworkConfig {
	// Define subnet ranges for different components
	subnetRanges := map[string]struct{ subnet, gateway string }{
		"service":       {"172.20.0.0/16", "172.20.0.1"},
		"database":      {"172.21.0.0/16", "172.21.0.1"},
		"dapr":          {"172.22.0.0/16", "172.22.0.1"},
		"observability": {"172.23.0.0/16", "172.23.0.1"},
		"website":       {"172.24.0.0/16", "172.24.0.1"},
	}
	
	networkRange, exists := subnetRanges[component]
	if !exists {
		networkRange = subnetRanges["service"] // Default fallback
	}
	
	return NetworkConfig{
		Name:        fmt.Sprintf("%s-%s", cb.environment, component),
		Environment: cb.environment,
		Component:   component,
		Driver:      "bridge",
		Subnet:      networkRange.subnet,
		Gateway:     networkRange.gateway,
		Labels: map[string]string{
			"project": "international-center",
			"layer":   "infrastructure",
		},
	}
}

// BuildObservabilityConfig creates observability configuration
func (cb *ConfigurationBuilder) BuildObservabilityConfig() ObservabilityConfig {
	return ObservabilityConfig{
		MetricsEnabled: cb.config.GetBool("metrics_enabled", true),
		TracingEnabled: cb.config.GetBool("tracing_enabled", cb.environment != "production"),
		LogLevel:       cb.config.Get("log_level", "info"),
		GrafanaURL:     cb.config.Get("grafana_url", "http://grafana:3000"),
		PrometheusURL:  cb.config.Get("prometheus_url", "http://prometheus:9090"),
		LokiURL:        cb.config.Get("loki_url", "http://loki:3100"),
	}
}

// Helper method to get standard service port mapping
func (cb *ConfigurationBuilder) GetStandardServicePorts() map[string]int {
	return map[string]int{
		"content-api":     8080,
		"services-api":    8081,
		"public-gateway":  8082,
		"admin-gateway":   8083,
		"website":         3000,
	}
}

// Helper method to get standard Dapr port mapping
func (cb *ConfigurationBuilder) GetStandardDaprPorts() map[string]struct{ http, grpc int } {
	return map[string]struct{ http, grpc int }{
		"content-api":     {3501, 50002},
		"services-api":    {3502, 50003},
		"public-gateway":  {3503, 50004},
		"admin-gateway":   {3504, 50006},
		"website":         {3505, 50007},
	}
}

// HealthCheckConfig defines health check parameters
type HealthCheckConfig struct {
	Path     string
	Port     int
	Interval time.Duration
	Timeout  time.Duration
	Retries  int
}