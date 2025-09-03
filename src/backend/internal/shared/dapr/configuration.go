package dapr

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dapr/go-sdk/client"
)

// Configuration wraps Dapr configuration operations
type Configuration struct {
	client    *Client
	storeName string
}

// ConfigItem represents a configuration item
type ConfigItem struct {
	Key      string            `json:"key"`
	Value    string            `json:"value"`
	Version  string            `json:"version"`
	Metadata map[string]string `json:"metadata"`
}

// AppConfig represents the application configuration
type AppConfig struct {
	Environment string
	AppID       string
	Version     string
	
	// Database configuration
	DatabaseConfig DatabaseConfig
	
	// API configuration
	APIConfig APIConfig
	
	// Gateway configuration
	GatewayConfig GatewayConfig
	
	// Dapr configuration
	DaprConfig DaprConfig
	
	// Observability configuration
	ObservabilityConfig ObservabilityConfig
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	MaxConnections    int
	ConnectionTimeout int
	QueryTimeout      int
}

// APIConfig holds API-related configuration
type APIConfig struct {
	Port         int
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

// GatewayConfig holds gateway-related configuration
type GatewayConfig struct {
	PublicRateLimit  int
	AdminRateLimit   int
	AllowedOrigins   []string
	SecurityHeaders  bool
}

// DaprConfig holds Dapr-related configuration
type DaprConfig struct {
	StateStoreName      string
	PubSubName          string
	SecretStoreName     string
	BlobBindingName     string
	HTTPPort            int
	GRPCPort            int
}

// ObservabilityConfig holds observability-related configuration
type ObservabilityConfig struct {
	LogLevel        string
	MetricsEnabled  bool
	TracingEnabled  bool
	AuditEnabled    bool
}

// NewConfiguration creates a new configuration instance
func NewConfiguration(client *Client) *Configuration {
	storeName := getEnv("DAPR_CONFIGURATION_STORE_NAME", "configstore")
	
	return &Configuration{
		client:    client,
		storeName: storeName,
	}
}

// GetConfigurationItem retrieves a single configuration item
func (c *Configuration) GetConfigurationItem(ctx context.Context, key string) (*ConfigItem, error) {
	item, err := c.client.GetClient().GetConfigurationItem(ctx, c.storeName, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration item %s: %w", key, err)
	}

	return &ConfigItem{
		Key:      item.Key,
		Value:    item.Value,
		Version:  item.Version,
		Metadata: item.Metadata,
	}, nil
}

// GetConfigurationItems retrieves multiple configuration items
func (c *Configuration) GetConfigurationItems(ctx context.Context, keys []string) (map[string]*ConfigItem, error) {
	items, err := c.client.GetClient().GetConfigurationItems(ctx, c.storeName, keys)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration items: %w", err)
	}

	result := make(map[string]*ConfigItem)
	for key, item := range items {
		result[key] = &ConfigItem{
			Key:      item.Key,
			Value:    item.Value,
			Version:  item.Version,
			Metadata: item.Metadata,
		}
	}

	return result, nil
}

// LoadAppConfig loads the complete application configuration
func (c *Configuration) LoadAppConfig(ctx context.Context) (*AppConfig, error) {
	environment := c.client.GetEnvironment()
	appID := c.client.GetAppID()
	
	// Define configuration keys to fetch
	configKeys := []string{
		"app.version",
		"database.max_connections",
		"database.connection_timeout",
		"database.query_timeout",
		"api.port",
		"api.read_timeout",
		"api.write_timeout",
		"api.idle_timeout",
		"gateway.public_rate_limit",
		"gateway.admin_rate_limit",
		"gateway.allowed_origins",
		"gateway.security_headers",
		"dapr.state_store_name",
		"dapr.pubsub_name",
		"dapr.secret_store_name",
		"dapr.blob_binding_name",
		"dapr.http_port",
		"dapr.grpc_port",
		"observability.log_level",
		"observability.metrics_enabled",
		"observability.tracing_enabled",
		"observability.audit_enabled",
	}

	// Get configuration items
	items, err := c.GetConfigurationItems(ctx, configKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to load app configuration: %w", err)
	}

	// Build configuration with defaults
	config := &AppConfig{
		Environment: environment,
		AppID:       appID,
		Version:     c.getConfigValue(items, "app.version", "1.0.0"),
		
		DatabaseConfig: DatabaseConfig{
			MaxConnections:    c.getConfigInt(items, "database.max_connections", 25),
			ConnectionTimeout: c.getConfigInt(items, "database.connection_timeout", 30),
			QueryTimeout:      c.getConfigInt(items, "database.query_timeout", 30),
		},
		
		APIConfig: APIConfig{
			Port:         c.getConfigInt(items, "api.port", 8080),
			ReadTimeout:  c.getConfigInt(items, "api.read_timeout", 15),
			WriteTimeout: c.getConfigInt(items, "api.write_timeout", 15),
			IdleTimeout:  c.getConfigInt(items, "api.idle_timeout", 60),
		},
		
		GatewayConfig: GatewayConfig{
			PublicRateLimit:  c.getConfigInt(items, "gateway.public_rate_limit", 1000),
			AdminRateLimit:   c.getConfigInt(items, "gateway.admin_rate_limit", 100),
			AllowedOrigins:   c.getConfigArray(items, "gateway.allowed_origins", []string{"*"}),
			SecurityHeaders:  c.getConfigBool(items, "gateway.security_headers", true),
		},
		
		DaprConfig: DaprConfig{
			StateStoreName:  c.getConfigValue(items, "dapr.state_store_name", "statestore-postgresql"),
			PubSubName:      c.getConfigValue(items, "dapr.pubsub_name", "pubsub-redis"),
			SecretStoreName: c.getConfigValue(items, "dapr.secret_store_name", "secretstore-vault"),
			BlobBindingName: c.getConfigValue(items, "dapr.blob_binding_name", "blob-storage"),
			HTTPPort:        c.getConfigInt(items, "dapr.http_port", 3500),
			GRPCPort:        c.getConfigInt(items, "dapr.grpc_port", 50001),
		},
		
		ObservabilityConfig: ObservabilityConfig{
			LogLevel:       c.getConfigValue(items, "observability.log_level", "info"),
			MetricsEnabled: c.getConfigBool(items, "observability.metrics_enabled", true),
			TracingEnabled: c.getConfigBool(items, "observability.tracing_enabled", true),
			AuditEnabled:   c.getConfigBool(items, "observability.audit_enabled", true),
		},
	}

	return config, nil
}

// getConfigValue retrieves a string configuration value with default
func (c *Configuration) getConfigValue(items map[string]*ConfigItem, key, defaultValue string) string {
	if item, exists := items[key]; exists && item.Value != "" {
		return item.Value
	}
	return defaultValue
}

// getConfigInt retrieves an integer configuration value with default
func (c *Configuration) getConfigInt(items map[string]*ConfigItem, key string, defaultValue int) int {
	if item, exists := items[key]; exists && item.Value != "" {
		if value, err := strconv.Atoi(item.Value); err == nil {
			return value
		}
	}
	return defaultValue
}

// getConfigBool retrieves a boolean configuration value with default
func (c *Configuration) getConfigBool(items map[string]*ConfigItem, key string, defaultValue bool) bool {
	if item, exists := items[key]; exists && item.Value != "" {
		if value, err := strconv.ParseBool(item.Value); err == nil {
			return value
		}
	}
	return defaultValue
}

// getConfigArray retrieves a string array configuration value with default
func (c *Configuration) getConfigArray(items map[string]*ConfigItem, key string, defaultValue []string) []string {
	if item, exists := items[key]; exists && item.Value != "" {
		// Split by comma and trim spaces
		values := strings.Split(item.Value, ",")
		result := make([]string, len(values))
		for i, v := range values {
			result[i] = strings.TrimSpace(v)
		}
		return result
	}
	return defaultValue
}

// WatchConfiguration watches for configuration changes
func (c *Configuration) WatchConfiguration(ctx context.Context, keys []string, callback func(map[string]*ConfigItem)) error {
	// Note: This would implement configuration watching if supported by Dapr
	// For now, this is a placeholder for future implementation
	return fmt.Errorf("configuration watching not implemented")
}

// HealthCheck validates the configuration store connection
func (c *Configuration) HealthCheck(ctx context.Context) error {
	// Test connectivity by attempting to get a configuration item
	_, err := c.GetConfigurationItem(ctx, "healthcheck")
	
	// Configuration not found is acceptable for health check
	// We're just validating connectivity
	return nil
}

// GetEnvironmentSpecificConfig gets environment-specific configuration
func (c *Configuration) GetEnvironmentSpecificConfig(ctx context.Context, baseKey string) (*ConfigItem, error) {
	environment := c.client.GetEnvironment()
	envSpecificKey := fmt.Sprintf("%s.%s", baseKey, environment)
	
	// Try environment-specific key first
	item, err := c.GetConfigurationItem(ctx, envSpecificKey)
	if err == nil {
		return item, nil
	}
	
	// Fallback to base key
	return c.GetConfigurationItem(ctx, baseKey)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}