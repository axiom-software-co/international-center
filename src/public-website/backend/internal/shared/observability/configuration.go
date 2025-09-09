package observability

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Configuration provides environment-specific observability settings
type Configuration struct {
	environment string
	values      map[string]interface{}
}

// NewConfiguration creates a new observability configuration for the given environment
func NewConfiguration(environment string) *Configuration {
	config := &Configuration{
		environment: environment,
		values:      make(map[string]interface{}),
	}
	config.loadDefaults()
	return config
}

// GetValue retrieves a configuration value
func (c *Configuration) GetValue(key string) (interface{}, bool) {
	value, exists := c.values[key]
	return value, exists
}

// SetValue sets a configuration value
func (c *Configuration) SetValue(key string, value interface{}) {
	c.values[key] = value
}

// GetString retrieves a string configuration value
func (c *Configuration) GetString(key string, defaultValue string) string {
	if value, exists := c.values[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetBool retrieves a boolean configuration value
func (c *Configuration) GetBool(key string, defaultValue bool) bool {
	if value, exists := c.values[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// GetInt retrieves an integer configuration value
func (c *Configuration) GetInt(key string, defaultValue int) int {
	if value, exists := c.values[key]; exists {
		if i, ok := value.(int); ok {
			return i
		}
	}
	return defaultValue
}

// GetDuration retrieves a duration configuration value
func (c *Configuration) GetDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := c.values[key]; exists {
		if d, ok := value.(time.Duration); ok {
			return d
		}
		if str, ok := value.(string); ok {
			if duration, err := time.ParseDuration(str); err == nil {
				return duration
			}
		}
	}
	return defaultValue
}

// GetFloat64 retrieves a float64 configuration value
func (c *Configuration) GetFloat64(key string, defaultValue float64) float64 {
	if value, exists := c.values[key]; exists {
		if f, ok := value.(float64); ok {
			return f
		}
	}
	return defaultValue
}

// LoadTracingConfiguration loads distributed tracing configuration
func LoadTracingConfiguration(environment string) (*Configuration, error) {
	config := NewConfiguration(environment)
	
	switch environment {
	case "development":
		config.values = map[string]interface{}{
			"enabled":        true,
			"sampling_rate":  1.0,
			"export_timeout": "30s",
			"batch_size":     100,
			"endpoint":       "http://localhost:4317",
		}
	case "production":
		config.values = map[string]interface{}{
			"enabled":        true,
			"sampling_rate":  0.1,
			"export_timeout": "10s", 
			"batch_size":     1000,
			"endpoint":       getEnv("TRACING_ENDPOINT", "https://tempo.grafana.cloud"),
		}
	default:
		return nil, fmt.Errorf("unsupported environment: %s", environment)
	}
	
	return config, nil
}

// LoadMetricsConfiguration loads service metrics configuration
func LoadMetricsConfiguration(environment string) (*Configuration, error) {
	config := NewConfiguration(environment)
	
	switch environment {
	case "development":
		config.values = map[string]interface{}{
			"enabled":           true,
			"collection_interval": "15s",
			"retention_period":    "7d",
			"export_endpoint":     "localhost:9090",
			"push_gateway":        "localhost:9091",
		}
	case "production":
		config.values = map[string]interface{}{
			"enabled":           true,
			"collection_interval": "30s",
			"retention_period":    "90d",
			"export_endpoint":     "grafana-cloud:443",
			"push_gateway":        getEnv("PROMETHEUS_PUSHGATEWAY", ""),
		}
	default:
		return nil, fmt.Errorf("unsupported environment: %s", environment)
	}
	
	return config, nil
}

// LoadLoggingConfiguration loads structured logging configuration
func LoadLoggingConfiguration(environment string) (*Configuration, error) {
	config := NewConfiguration(environment)
	
	switch environment {
	case "development":
		config.values = map[string]interface{}{
			"level":           "debug",
			"format":          "json",
			"output":          "stdout",
			"correlation":     true,
			"performance":     true,
		}
	case "production":
		config.values = map[string]interface{}{
			"level":           "info",
			"format":          "json",
			"output":          "grafana-loki",
			"correlation":     true,
			"sampling_rate":   0.1,
			"endpoint":        getEnv("LOKI_ENDPOINT", "https://logs.grafana.cloud"),
		}
	default:
		return nil, fmt.Errorf("unsupported environment: %s", environment)
	}
	
	return config, nil
}

// LoadAuditConfiguration loads audit logging configuration
func LoadAuditConfiguration(environment string) (*Configuration, error) {
	config := NewConfiguration(environment)
	
	switch environment {
	case "development":
		config.values = map[string]interface{}{
			"enabled":              true,
			"persistence_endpoint": "local-loki",
			"encryption_enabled":   true,
			"retention_days":       30,
			"backup_enabled":       true,
			"alert_on_failure":     true,
		}
	case "production":
		config.values = map[string]interface{}{
			"enabled":              true,
			"persistence_endpoint": "grafana-cloud-loki",
			"encryption_enabled":   true,
			"retention_days":       2555, // 7 years for compliance
			"backup_enabled":       true,
			"alert_on_failure":     true,
			"compliance_frameworks": []string{"HIPAA", "SOC2", "GDPR"},
		}
	default:
		return nil, fmt.Errorf("unsupported environment: %s", environment)
	}
	
	return config, nil
}

// LoadHealthMonitoringConfiguration loads health monitoring configuration
func LoadHealthMonitoringConfiguration(environment string) (*Configuration, error) {
	config := NewConfiguration(environment)
	
	switch environment {
	case "development":
		config.values = map[string]interface{}{
			"enabled":                    true,
			"check_interval":             "60s",
			"dependency_timeout":         "30s",
			"circuit_breaker_enabled":    false,
			"alert_on_degradation":       false,
			"metrics_retention_days":     7,
		}
	case "production":
		config.values = map[string]interface{}{
			"enabled":                    true,
			"check_interval":             "30s",
			"dependency_timeout":         "10s",
			"circuit_breaker_enabled":    true,
			"alert_on_degradation":       true,
			"metrics_retention_days":     30,
			"comprehensive_checks":       true,
		}
	default:
		return nil, fmt.Errorf("unsupported environment: %s", environment)
	}
	
	return config, nil
}

// loadDefaults loads default configuration values
func (c *Configuration) loadDefaults() {
	// Default values that apply to all environments
	c.values["service_name"] = getEnv("SERVICE_NAME", "unknown-service")
	c.values["service_version"] = getEnv("SERVICE_VERSION", "unknown-version")
	c.values["environment"] = c.environment
}

// getEnv retrieves environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvBool retrieves boolean environment variable with fallback
func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return fallback
}

// getEnvInt retrieves integer environment variable with fallback
func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}