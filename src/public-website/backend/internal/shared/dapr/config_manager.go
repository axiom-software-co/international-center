package dapr

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ConfigManager provides centralized configuration management through Dapr secret store
// with fallback to environment variables for migration
type ConfigManager struct {
	secrets     *Secrets
	serviceName string
	environment string
	fallbackToEnv bool
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(client *Client, serviceName string) *ConfigManager {
	return &ConfigManager{
		secrets:       NewSecrets(client),
		serviceName:   serviceName,
		environment:   client.GetEnvironment(),
		fallbackToEnv: getEnv("ALLOW_ENV_FALLBACK", "true") == "true",
	}
}

// GetString retrieves a string configuration value
func (c *ConfigManager) GetString(ctx context.Context, key, defaultValue string) string {
	// Try Dapr secret store first
	value, err := c.secrets.GetSecret(ctx, key)
	if err == nil && value != "" {
		return value
	}

	// Fallback to environment variables if allowed
	if c.fallbackToEnv {
		if envValue := os.Getenv(strings.ToUpper(key)); envValue != "" {
			return envValue
		}
		if envValue := os.Getenv(key); envValue != "" {
			return envValue
		}
	}

	return defaultValue
}

// GetInt retrieves an integer configuration value
func (c *ConfigManager) GetInt(ctx context.Context, key string, defaultValue int) int {
	stringValue := c.GetString(ctx, key, "")
	if stringValue == "" {
		return defaultValue
	}

	if intValue, err := strconv.Atoi(stringValue); err == nil {
		return intValue
	}

	return defaultValue
}

// GetBool retrieves a boolean configuration value
func (c *ConfigManager) GetBool(ctx context.Context, key string, defaultValue bool) bool {
	stringValue := c.GetString(ctx, key, "")
	if stringValue == "" {
		return defaultValue
	}

	stringValue = strings.ToLower(stringValue)
	switch stringValue {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

// GetDuration retrieves a duration configuration value
func (c *ConfigManager) GetDuration(ctx context.Context, key string, defaultValue time.Duration) time.Duration {
	stringValue := c.GetString(ctx, key, "")
	if stringValue == "" {
		return defaultValue
	}

	if duration, err := time.ParseDuration(stringValue); err == nil {
		return duration
	}

	return defaultValue
}

// GetDatabaseConfig retrieves database configuration for the service
func (c *ConfigManager) GetDatabaseConfig(ctx context.Context) (*DatabaseConfig, error) {
	connectionString, err := c.secrets.GetDatabaseConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection string: %w", err)
	}

	return &DatabaseConfig{
		ConnectionString: connectionString,
		MaxConnections:   c.GetInt(ctx, "database-max-connections", 25),
		MaxIdleTime:      c.GetDuration(ctx, "database-max-idle-time", 15*time.Minute),
		ConnTimeout:      c.GetDuration(ctx, "database-connection-timeout", 30*time.Second),
		ReadTimeout:      c.GetDuration(ctx, "database-read-timeout", 30*time.Second),
		WriteTimeout:     c.GetDuration(ctx, "database-write-timeout", 30*time.Second),
	}, nil
}

// GetRedisConfig retrieves Redis configuration for the service
func (c *ConfigManager) GetRedisConfig(ctx context.Context) (*RedisConfig, error) {
	connectionString, err := c.secrets.GetRedisConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis connection string: %w", err)
	}

	return &RedisConfig{
		ConnectionString: connectionString,
		MaxConnections:   c.GetInt(ctx, "redis-max-connections", 10),
		IdleTimeout:      c.GetDuration(ctx, "redis-idle-timeout", 5*time.Minute),
		ConnTimeout:      c.GetDuration(ctx, "redis-connection-timeout", 5*time.Second),
		ReadTimeout:      c.GetDuration(ctx, "redis-read-timeout", 3*time.Second),
		WriteTimeout:     c.GetDuration(ctx, "redis-write-timeout", 3*time.Second),
	}, nil
}

// GetServerConfig retrieves HTTP server configuration
func (c *ConfigManager) GetServerConfig(ctx context.Context) *ServerConfig {
	return &ServerConfig{
		Port:            c.GetString(ctx, "server-port", "8080"),
		Host:            c.GetString(ctx, "server-host", "0.0.0.0"),
		ReadTimeout:     c.GetDuration(ctx, "server-read-timeout", 30*time.Second),
		WriteTimeout:    c.GetDuration(ctx, "server-write-timeout", 30*time.Second),
		IdleTimeout:     c.GetDuration(ctx, "server-idle-timeout", 120*time.Second),
		ShutdownTimeout: c.GetDuration(ctx, "server-shutdown-timeout", 30*time.Second),
		EnableTLS:       c.GetBool(ctx, "server-enable-tls", false),
		TLSCertFile:     c.GetString(ctx, "server-tls-cert-file", ""),
		TLSKeyFile:      c.GetString(ctx, "server-tls-key-file", ""),
	}
}

// GetMonitoringConfig retrieves monitoring and observability configuration
func (c *ConfigManager) GetMonitoringConfig(ctx context.Context) (*MonitoringConfig, error) {
	grafanaKey, err := c.secrets.GetGrafanaAPIKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Grafana API key: %w", err)
	}

	return &MonitoringConfig{
		GrafanaAPIKey:      grafanaKey,
		MetricsEnabled:     c.GetBool(ctx, "monitoring-metrics-enabled", true),
		TracingEnabled:     c.GetBool(ctx, "monitoring-tracing-enabled", true),
		LoggingLevel:       c.GetString(ctx, "monitoring-logging-level", "info"),
		HealthCheckPath:    c.GetString(ctx, "monitoring-health-check-path", "/health"),
		MetricsPath:        c.GetString(ctx, "monitoring-metrics-path", "/metrics"),
		PrometheusEndpoint: c.GetString(ctx, "monitoring-prometheus-endpoint", "http://localhost:9090"),
		JaegerEndpoint:     c.GetString(ctx, "monitoring-jaeger-endpoint", "http://localhost:14268"),
	}, nil
}

// GetEmailConfig retrieves email service configuration
func (c *ConfigManager) GetEmailConfig(ctx context.Context) (*EmailConfig, error) {
	credentials, err := c.secrets.GetEmailServiceCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get email service credentials: %w", err)
	}

	port, _ := strconv.Atoi(credentials["email-smtp-port"])
	if port == 0 {
		port = 587 // Default SMTP port
	}

	return &EmailConfig{
		SMTPHost:     credentials["email-smtp-host"],
		SMTPPort:     port,
		Username:     credentials["email-smtp-username"],
		Password:     credentials["email-smtp-password"],
		FromAddress:  credentials["email-from-address"],
		EnableTLS:    c.GetBool(ctx, "email-enable-tls", true),
		Timeout:      c.GetDuration(ctx, "email-timeout", 30*time.Second),
		MaxRetries:   c.GetInt(ctx, "email-max-retries", 3),
	}, nil
}

// GetSMSConfig retrieves SMS service configuration
func (c *ConfigManager) GetSMSConfig(ctx context.Context) (*SMSConfig, error) {
	credentials, err := c.secrets.GetSMSServiceCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SMS service credentials: %w", err)
	}

	return &SMSConfig{
		ServiceEndpoint: credentials["sms-service-endpoint"],
		APIKey:          credentials["sms-service-api-key"],
		AccountSID:      credentials["sms-service-account-sid"],
		FromNumber:      credentials["sms-from-number"],
		Timeout:         c.GetDuration(ctx, "sms-timeout", 30*time.Second),
		MaxRetries:      c.GetInt(ctx, "sms-max-retries", 3),
	}, nil
}

// GetAuthConfig retrieves authentication configuration
func (c *ConfigManager) GetAuthConfig(ctx context.Context) (*AuthConfig, error) {
	authSecrets, err := c.secrets.GetAuthenticationSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get authentication secrets: %w", err)
	}

	return &AuthConfig{
		OAuth2ClientID:       authSecrets["oauth2-client-id"],
		OAuth2ClientSecret:   authSecrets["oauth2-client-secret"],
		JWTSigningKey:        authSecrets["jwt-signing-key"],
		SessionEncryptionKey: authSecrets["session-encryption-key"],
		CSRFTokenKey:         authSecrets["csrf-token-key"],
		TokenExpiration:      c.GetDuration(ctx, "auth-token-expiration", 24*time.Hour),
		SessionTimeout:       c.GetDuration(ctx, "auth-session-timeout", 8*time.Hour),
		EnableMFA:            c.GetBool(ctx, "auth-enable-mfa", false),
	}, nil
}

// ValidateConfiguration validates that all required configuration is available
func (c *ConfigManager) ValidateConfiguration(ctx context.Context) error {
	return c.secrets.ValidateServiceSecrets(ctx, c.serviceName)
}

// Configuration structs
type DatabaseConfig struct {
	ConnectionString string
	MaxConnections   int
	MaxIdleTime      time.Duration
	ConnTimeout      time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
}

type RedisConfig struct {
	ConnectionString string
	MaxConnections   int
	IdleTimeout      time.Duration
	ConnTimeout      time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
}

type ServerConfig struct {
	Port            string
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	EnableTLS       bool
	TLSCertFile     string
	TLSKeyFile      string
}

type MonitoringConfig struct {
	GrafanaAPIKey      string
	MetricsEnabled     bool
	TracingEnabled     bool
	LoggingLevel       string
	HealthCheckPath    string
	MetricsPath        string
	PrometheusEndpoint string
	JaegerEndpoint     string
}

type EmailConfig struct {
	SMTPHost    string
	SMTPPort    int
	Username    string
	Password    string
	FromAddress string
	EnableTLS   bool
	Timeout     time.Duration
	MaxRetries  int
}

type SMSConfig struct {
	ServiceEndpoint string
	APIKey          string
	AccountSID      string
	FromNumber      string
	Timeout         time.Duration
	MaxRetries      int
}

type AuthConfig struct {
	OAuth2ClientID       string
	OAuth2ClientSecret   string
	JWTSigningKey        string
	SessionEncryptionKey string
	CSRFTokenKey         string
	TokenExpiration      time.Duration
	SessionTimeout       time.Duration
	EnableMFA            bool
}