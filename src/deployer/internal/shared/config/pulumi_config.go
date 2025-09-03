package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// PulumiConfig wraps Pulumi configuration with environment-specific handling
type PulumiConfig struct {
	config *config.Config
	ctx    *pulumi.Context
	env    Environment
}

// NewPulumiConfig creates a new Pulumi configuration wrapper
func NewPulumiConfig(ctx *pulumi.Context, env Environment) *PulumiConfig {
	cfg := config.New(ctx, "")
	
	return &PulumiConfig{
		config: cfg,
		ctx:    ctx,
		env:    env,
	}
}

// GetString returns a string configuration value with environment prefix
func (pc *PulumiConfig) GetString(key string) string {
	// Try environment-specific key first
	envKey := pc.getEnvironmentKey(key)
	if value := pc.config.Get(envKey); value != "" {
		return value
	}
	
	// Fallback to generic key
	return pc.config.Get(key)
}

// GetRequiredString returns a required string configuration value
func (pc *PulumiConfig) GetRequiredString(key string) string {
	value := pc.GetString(key)
	if value == "" {
		// This will cause Pulumi to fail with a descriptive error
		return pc.config.Require(key)
	}
	return value
}

// GetInt returns an integer configuration value
func (pc *PulumiConfig) GetInt(key string, defaultValue int) int {
	strValue := pc.GetString(key)
	if strValue == "" {
		return defaultValue
	}
	
	intValue, err := strconv.Atoi(strValue)
	if err != nil {
		pc.ctx.Log.Warn(fmt.Sprintf("Invalid integer value for %s: %s, using default %d", key, strValue, defaultValue), nil)
		return defaultValue
	}
	
	return intValue
}

// GetBool returns a boolean configuration value
func (pc *PulumiConfig) GetBool(key string, defaultValue bool) bool {
	strValue := pc.GetString(key)
	if strValue == "" {
		return defaultValue
	}
	
	boolValue, err := strconv.ParseBool(strValue)
	if err != nil {
		pc.ctx.Log.Warn(fmt.Sprintf("Invalid boolean value for %s: %s, using default %t", key, strValue, defaultValue), nil)
		return defaultValue
	}
	
	return boolValue
}

// GetStringSlice returns a slice of strings from comma-separated configuration
func (pc *PulumiConfig) GetStringSlice(key string, defaultValue []string) []string {
	strValue := pc.GetString(key)
	if strValue == "" {
		return defaultValue
	}
	
	// Split by comma and trim whitespace
	values := strings.Split(strValue, ",")
	result := make([]string, 0, len(values))
	
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	if len(result) == 0 {
		return defaultValue
	}
	
	return result
}

// GetSecret returns a secret configuration value
func (pc *PulumiConfig) GetSecret(key string) pulumi.StringOutput {
	envKey := pc.getEnvironmentKey(key)
	
	// Try environment-specific secret first
	if pc.config.Get(envKey) != "" {
		return pc.config.GetSecret(envKey)
	}
	
	// Fallback to generic secret key
	return pc.config.GetSecret(key)
}

// GetRequiredSecret returns a required secret configuration value
func (pc *PulumiConfig) GetRequiredSecret(key string) pulumi.StringOutput {
	envKey := pc.getEnvironmentKey(key)
	
	// Try environment-specific secret first
	if pc.config.Get(envKey) != "" {
		return pc.config.RequireSecret(envKey)
	}
	
	// Fallback to generic secret key
	return pc.config.RequireSecret(key)
}

// GetDatabaseConfig returns database configuration for the environment
func (pc *PulumiConfig) GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     pc.GetString("database.host"),
		Port:     pc.GetInt("database.port", pc.getDefaultDatabasePort()),
		Database: pc.GetString("database.name"),
		Username: pc.GetString("database.username"),
		Password: pc.GetSecret("database.password"),
		SSLMode:  pc.GetString("database.ssl_mode"),
		
		// Connection pool settings
		MaxConnections:     pc.GetInt("database.max_connections", pc.getDefaultMaxConnections()),
		MaxIdleConnections: pc.GetInt("database.max_idle_connections", pc.getDefaultMaxIdleConnections()),
		ConnMaxLifetime:    pc.GetString("database.conn_max_lifetime"),
	}
}

// GetRedisConfig returns Redis configuration for the environment
func (pc *PulumiConfig) GetRedisConfig() RedisConfig {
	return RedisConfig{
		Host:     pc.GetString("redis.host"),
		Port:     pc.GetInt("redis.port", pc.getDefaultRedisPort()),
		Password: pc.GetSecret("redis.password"),
		Database: pc.GetInt("redis.database", 0),
		
		// Connection pool settings
		PoolSize:        pc.GetInt("redis.pool_size", pc.getDefaultRedisPoolSize()),
		MinIdleConns:    pc.GetInt("redis.min_idle_conns", pc.getDefaultRedisMinIdleConns()),
		MaxRetries:      pc.GetInt("redis.max_retries", 3),
		DialTimeout:     pc.GetString("redis.dial_timeout"),
		ReadTimeout:     pc.GetString("redis.read_timeout"),
		WriteTimeout:    pc.GetString("redis.write_timeout"),
		PoolTimeout:     pc.GetString("redis.pool_timeout"),
		IdleTimeout:     pc.GetString("redis.idle_timeout"),
	}
}

// GetStorageConfig returns storage configuration for the environment
func (pc *PulumiConfig) GetStorageConfig() StorageConfig {
	return StorageConfig{
		AccountName:   pc.GetString("storage.account_name"),
		AccountKey:    pc.GetSecret("storage.account_key"),
		ContainerName: pc.GetString("storage.container_name"),
		Endpoint:      pc.GetString("storage.endpoint"),
		
		// Access settings
		AccessTier:     pc.GetString("storage.access_tier"),
		Redundancy:     pc.GetString("storage.redundancy"),
		EnableHTTPS:    pc.GetBool("storage.enable_https", true),
		AllowBlobPublicAccess: pc.GetBool("storage.allow_blob_public_access", false),
	}
}

// GetObservabilityConfig returns observability configuration
func (pc *PulumiConfig) GetObservabilityConfig() ObservabilityConfig {
	return ObservabilityConfig{
		GrafanaEndpoint:    pc.GetString("observability.grafana_endpoint"),
		GrafanaAPIKey:      pc.GetSecret("observability.grafana_api_key"),
		LokiEndpoint:       pc.GetString("observability.loki_endpoint"),
		PrometheusEndpoint: pc.GetString("observability.prometheus_endpoint"),
		
		// Feature toggles
		EnableMetrics: pc.GetBool("observability.enable_metrics", true),
		EnableTracing: pc.GetBool("observability.enable_tracing", true),
		EnableLogging: pc.GetBool("observability.enable_logging", true),
		
		// Retention settings
		MetricsRetentionDays: pc.GetInt("observability.metrics_retention_days", pc.getDefaultMetricsRetention()),
		LogsRetentionDays:    pc.GetInt("observability.logs_retention_days", pc.getDefaultLogsRetention()),
		TracesRetentionDays:  pc.GetInt("observability.traces_retention_days", pc.getDefaultTracesRetention()),
	}
}

// GetVaultConfig returns Vault configuration
func (pc *PulumiConfig) GetVaultConfig() VaultConfig {
	return VaultConfig{
		Address:   pc.GetString("vault.address"),
		Token:     pc.GetSecret("vault.token"),
		Namespace: pc.GetString("vault.namespace"),
		
		// Auth settings
		AuthMethod: pc.GetString("vault.auth_method"),
		RoleID:     pc.GetString("vault.role_id"),
		SecretID:   pc.GetSecret("vault.secret_id"),
		
		// Connection settings
		MaxRetries:    pc.GetInt("vault.max_retries", 3),
		Timeout:       pc.GetString("vault.timeout"),
		TLSSkipVerify: pc.GetBool("vault.tls_skip_verify", false),
	}
}

// Private helper methods

func (pc *PulumiConfig) getEnvironmentKey(key string) string {
	return fmt.Sprintf("%s.%s", pc.env, key)
}

func (pc *PulumiConfig) getDefaultDatabasePort() int {
	return 5432 // PostgreSQL default
}

func (pc *PulumiConfig) getDefaultRedisPort() int {
	return 6379 // Redis default
}

func (pc *PulumiConfig) getDefaultMaxConnections() int {
	switch pc.env {
	case EnvironmentDevelopment:
		return 10
	case EnvironmentStaging:
		return 25
	case EnvironmentProduction:
		return 50
	default:
		return 10
	}
}

func (pc *PulumiConfig) getDefaultMaxIdleConnections() int {
	return pc.getDefaultMaxConnections() / 2
}

func (pc *PulumiConfig) getDefaultRedisPoolSize() int {
	switch pc.env {
	case EnvironmentDevelopment:
		return 10
	case EnvironmentStaging:
		return 20
	case EnvironmentProduction:
		return 50
	default:
		return 10
	}
}

func (pc *PulumiConfig) getDefaultRedisMinIdleConns() int {
	return pc.getDefaultRedisPoolSize() / 4
}

func (pc *PulumiConfig) getDefaultMetricsRetention() int {
	switch pc.env {
	case EnvironmentDevelopment:
		return 7   // 1 week
	case EnvironmentStaging:
		return 30  // 1 month
	case EnvironmentProduction:
		return 365 // 1 year
	default:
		return 7
	}
}

func (pc *PulumiConfig) getDefaultLogsRetention() int {
	switch pc.env {
	case EnvironmentDevelopment:
		return 7   // 1 week
	case EnvironmentStaging:
		return 90  // 3 months
	case EnvironmentProduction:
		return 2555 // 7 years for compliance
	default:
		return 7
	}
}

func (pc *PulumiConfig) getDefaultTracesRetention() int {
	switch pc.env {
	case EnvironmentDevelopment:
		return 3   // 3 days
	case EnvironmentStaging:
		return 14  // 2 weeks
	case EnvironmentProduction:
		return 30  // 1 month
	default:
		return 3
	}
}

// Configuration structures

type DatabaseConfig struct {
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Database string            `json:"database"`
	Username string            `json:"username"`
	Password pulumi.StringOutput `json:"password"`
	SSLMode  string            `json:"ssl_mode"`
	
	// Connection pool settings
	MaxConnections     int    `json:"max_connections"`
	MaxIdleConnections int    `json:"max_idle_connections"`
	ConnMaxLifetime    string `json:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Password pulumi.StringOutput `json:"password"`
	Database int               `json:"database"`
	
	// Connection pool settings
	PoolSize        int    `json:"pool_size"`
	MinIdleConns    int    `json:"min_idle_conns"`
	MaxRetries      int    `json:"max_retries"`
	DialTimeout     string `json:"dial_timeout"`
	ReadTimeout     string `json:"read_timeout"`
	WriteTimeout    string `json:"write_timeout"`
	PoolTimeout     string `json:"pool_timeout"`
	IdleTimeout     string `json:"idle_timeout"`
}

type StorageConfig struct {
	AccountName   string            `json:"account_name"`
	AccountKey    pulumi.StringOutput `json:"account_key"`
	ContainerName string            `json:"container_name"`
	Endpoint      string            `json:"endpoint"`
	
	// Access settings
	AccessTier            string `json:"access_tier"`
	Redundancy            string `json:"redundancy"`
	EnableHTTPS           bool   `json:"enable_https"`
	AllowBlobPublicAccess bool   `json:"allow_blob_public_access"`
}

type ObservabilityConfig struct {
	GrafanaEndpoint    string            `json:"grafana_endpoint"`
	GrafanaAPIKey      pulumi.StringOutput `json:"grafana_api_key"`
	LokiEndpoint       string            `json:"loki_endpoint"`
	PrometheusEndpoint string            `json:"prometheus_endpoint"`
	
	// Feature toggles
	EnableMetrics bool `json:"enable_metrics"`
	EnableTracing bool `json:"enable_tracing"`
	EnableLogging bool `json:"enable_logging"`
	
	// Retention settings
	MetricsRetentionDays int `json:"metrics_retention_days"`
	LogsRetentionDays    int `json:"logs_retention_days"`
	TracesRetentionDays  int `json:"traces_retention_days"`
}

type VaultConfig struct {
	Address   string            `json:"address"`
	Token     pulumi.StringOutput `json:"token"`
	Namespace string            `json:"namespace"`
	
	// Auth settings
	AuthMethod string            `json:"auth_method"`
	RoleID     string            `json:"role_id"`
	SecretID   pulumi.StringOutput `json:"secret_id"`
	
	// Connection settings
	MaxRetries    int    `json:"max_retries"`
	Timeout       string `json:"timeout"`
	TLSSkipVerify bool   `json:"tls_skip_verify"`
}