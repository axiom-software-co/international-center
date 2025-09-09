package notifications

import (
	"fmt"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// NotificationConfig contains configuration for the notification system
type NotificationConfig struct {
	Version           string                `json:"version"`
	MaxRetries        int                   `json:"max_retries"`
	RetryDelay        time.Duration         `json:"retry_delay"`
	BatchSize         int                   `json:"batch_size"`
	ProcessingTimeout time.Duration         `json:"processing_timeout"`
	Database          *DatabaseConfig       `json:"database"`
	MessageQueue      *MessageQueueConfig   `json:"message_queue"`
	Email             *EmailHandlerConfig   `json:"email"`
	SMS               *SMSHandlerConfig     `json:"sms"`
	Slack             *SlackHandlerConfig   `json:"slack"`
	Observability     *ObservabilityConfig  `json:"observability"`
	Reliability       *ReliabilityConfig    `json:"reliability"`
	Performance       *PerformanceConfig    `json:"performance"`
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	ConnectionString    string        `json:"connection_string"`
	MaxOpenConnections  int           `json:"max_open_connections"`
	MaxIdleConnections  int           `json:"max_idle_connections"`
	ConnectionTimeout   time.Duration `json:"connection_timeout"`
	QueryTimeout        time.Duration `json:"query_timeout"`
	MigrationsPath      string        `json:"migrations_path"`
}

// MessageQueueConfig contains message queue configuration
type MessageQueueConfig struct {
	ConnectionString   string        `json:"connection_string"`
	MaxRetries         int           `json:"max_retries"`
	RetryDelay         time.Duration `json:"retry_delay"`
	MessageTimeout     time.Duration `json:"message_timeout"`
	PrefetchCount      int           `json:"prefetch_count"`
	EnableDeadLetter   bool          `json:"enable_dead_letter"`
	DeadLetterExchange string        `json:"dead_letter_exchange"`
}

// AzureEmailConfig contains Azure Communication Services email configuration
type AzureEmailConfig struct {
	ConnectionString string `json:"connection_string"`
	SenderAddress    string `json:"sender_address"`
	SenderName       string `json:"sender_name"`
	MaxRetries       int    `json:"max_retries"`
	RetryDelay       int    `json:"retry_delay"` // seconds
	RequestTimeout   int    `json:"request_timeout"` // seconds
}

// AzureSMSConfig contains Azure Communication Services SMS configuration
type AzureSMSConfig struct {
	ConnectionString string `json:"connection_string"`
	FromNumber       string `json:"from_number"`
	MaxRetries       int    `json:"max_retries"`
	RetryDelay       int    `json:"retry_delay"` // seconds
	RequestTimeout   int    `json:"request_timeout"` // seconds
}

// SlackConfig contains Slack API configuration
type SlackConfig struct {
	BotToken         string            `json:"bot_token"`
	DefaultChannel   string            `json:"default_channel"`
	ChannelMapping   map[string]string `json:"channel_mapping"` // event_type -> channel
	MaxRetries       int               `json:"max_retries"`
	RetryDelay       int               `json:"retry_delay"` // seconds
	RequestTimeout   int               `json:"request_timeout"` // seconds
	RateLimit        int               `json:"rate_limit"` // requests per minute
}

// EmailHandlerConfig contains email handler configuration
type EmailHandlerConfig struct {
	Enabled          bool          `json:"enabled"`
	QueueName        string        `json:"queue_name"`
	Workers          int           `json:"workers"`
	ProcessingDelay  time.Duration `json:"processing_delay"`
	Azure            *AzureEmailConfig `json:"azure"`
}

// SMSHandlerConfig contains SMS handler configuration
type SMSHandlerConfig struct {
	Enabled         bool          `json:"enabled"`
	QueueName       string        `json:"queue_name"`
	Workers         int           `json:"workers"`
	ProcessingDelay time.Duration `json:"processing_delay"`
	Azure           *AzureSMSConfig `json:"azure"`
}

// SlackHandlerConfig contains Slack handler configuration
type SlackHandlerConfig struct {
	Enabled         bool          `json:"enabled"`
	QueueName       string        `json:"queue_name"`
	Workers         int           `json:"workers"`
	ProcessingDelay time.Duration `json:"processing_delay"`
	Slack           *SlackConfig  `json:"slack"`
}

// ObservabilityConfig contains observability configuration
type ObservabilityConfig struct {
	LogLevel          string `json:"log_level"`
	MetricsEnabled    bool   `json:"metrics_enabled"`
	TracingEnabled    bool   `json:"tracing_enabled"`
	HealthCheckPort   int    `json:"health_check_port"`
	MetricsPort       int    `json:"metrics_port"`
}

// ReliabilityConfig contains reliability and fault tolerance configuration  
type ReliabilityConfig struct {
	CircuitBreaker *CircuitBreakerConfig `json:"circuit_breaker"`
	Retry          *RetryConfig          `json:"retry"`
	RateLimit      *RateLimitConfig      `json:"rate_limit"`
	HealthCheck    *HealthCheckConfig    `json:"health_check"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	Enabled        bool          `json:"enabled"`
	MaxRequests    int64         `json:"max_requests"`
	WindowDuration time.Duration `json:"window_duration"`
	BurstSize      int64         `json:"burst_size"`
}

// HealthCheckConfig contains health check configuration
type HealthCheckConfig struct {
	Enabled         bool          `json:"enabled"`
	CheckInterval   time.Duration `json:"check_interval"`
	TimeoutDuration time.Duration `json:"timeout_duration"`
	MaxFailures     int           `json:"max_failures"`
}

// PerformanceConfig contains performance optimization configuration
type PerformanceConfig struct {
	WorkerPool     *WorkerPoolConfig     `json:"worker_pool"`
	BatchProcess   *BatchProcessConfig   `json:"batch_process"`
	Cache          *CacheConfig          `json:"cache"`
	ConnectionPool *ConnectionPoolConfig `json:"connection_pool"`
}

// WorkerPoolConfig contains worker pool configuration
type WorkerPoolConfig struct {
	Enabled     bool          `json:"enabled"`
	WorkerCount int           `json:"worker_count"`
	QueueSize   int           `json:"queue_size"`
	Timeout     time.Duration `json:"timeout"`
}

// BatchProcessConfig contains batch processing configuration
type BatchProcessConfig struct {
	Enabled       bool          `json:"enabled"`
	BatchSize     int           `json:"batch_size"`
	FlushInterval time.Duration `json:"flush_interval"`
	MaxBatches    int           `json:"max_batches"`
}

// CacheConfig contains caching configuration
type CacheConfig struct {
	Enabled         bool          `json:"enabled"`
	DefaultTTL      time.Duration `json:"default_ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	MaxSize         int           `json:"max_size"`
}

// ConnectionPoolConfig contains connection pool configuration
type ConnectionPoolConfig struct {
	MaxConnections    int           `json:"max_connections"`
	MinConnections    int           `json:"min_connections"`
	IdleTimeout       time.Duration `json:"idle_timeout"`
	ConnectTimeout    time.Duration `json:"connect_timeout"`
	MaxLifetime       time.Duration `json:"max_lifetime"`
}

// DefaultNotificationConfig returns a default configuration
func DefaultNotificationConfig() *NotificationConfig {
	return &NotificationConfig{
		Version:           "1.0.0",
		MaxRetries:        3,
		RetryDelay:        5 * time.Second,
		BatchSize:         100,
		ProcessingTimeout: 30 * time.Second,
		Database: &DatabaseConfig{
			MaxOpenConnections: 25,
			MaxIdleConnections: 5,
			ConnectionTimeout:  10 * time.Second,
			QueryTimeout:       30 * time.Second,
			MigrationsPath:     "./migrations",
		},
		MessageQueue: &MessageQueueConfig{
			ConnectionString:   "amqp://guest:guest@localhost:5672/",
			MaxRetries:         3,
			RetryDelay:         5 * time.Second,
			MessageTimeout:     30 * time.Second,
			PrefetchCount:      10,
			EnableDeadLetter:   true,
			DeadLetterExchange: "notifications.dead-letter",
		},
		Email: &EmailHandlerConfig{
			Enabled:         true,
			QueueName:       "email-notifications",
			Workers:         5,
			ProcessingDelay: 1 * time.Second,
			Azure: &AzureEmailConfig{
				ConnectionString: "endpoint=https://your-acs-resource.communication.azure.com/;accesskey=your-access-key",
				SenderAddress:    "noreply@your-domain.com",
				SenderName:       "International Center",
				MaxRetries:       3,
				RetryDelay:       5,
				RequestTimeout:   30,
			},
		},
		SMS: &SMSHandlerConfig{
			Enabled:         true,
			QueueName:       "sms-notifications", 
			Workers:         3,
			ProcessingDelay: 2 * time.Second,
			Azure: &AzureSMSConfig{
				ConnectionString: "endpoint=https://your-acs-resource.communication.azure.com/;accesskey=your-access-key",
				FromNumber:       "+1234567890",
				MaxRetries:       3,
				RetryDelay:       5,
				RequestTimeout:   30,
			},
		},
		Slack: &SlackHandlerConfig{
			Enabled:         true,
			QueueName:       "slack-notifications",
			Workers:         3,
			ProcessingDelay: 1 * time.Second,
			Slack: &SlackConfig{
				BotToken:       "xoxb-your-bot-token",
				DefaultChannel: "#general",
				ChannelMapping: map[string]string{
					"system-error":     "#alerts",
					"capacity-alert":   "#alerts",
					"compliance-alert": "#compliance",
				},
				MaxRetries:     3,
				RetryDelay:     5,
				RequestTimeout: 30,
				RateLimit:      60, // 60 requests per minute
			},
		},
		Observability: &ObservabilityConfig{
			LogLevel:        "info",
			MetricsEnabled:  true,
			TracingEnabled:  true,
			HealthCheckPort: 8080,
			MetricsPort:     9090,
		},
		Reliability: &ReliabilityConfig{
			CircuitBreaker: &CircuitBreakerConfig{
				MaxFailures:      5,
				ResetTimeout:     30 * time.Second,
				FailureThreshold: 0.6,
				MinRequests:      10,
			},
			Retry: &RetryConfig{
				MaxAttempts:     3,
				InitialDelay:    100 * time.Millisecond,
				MaxDelay:        10 * time.Second,
				BackoffFactor:   2.0,
				RetryableErrors: []string{"timeout", "connection_refused", "service_unavailable"},
			},
			RateLimit: &RateLimitConfig{
				Enabled:        true,
				MaxRequests:    100,
				WindowDuration: time.Minute,
				BurstSize:      20,
			},
			HealthCheck: &HealthCheckConfig{
				Enabled:         true,
				CheckInterval:   30 * time.Second,
				TimeoutDuration: 5 * time.Second,
				MaxFailures:     3,
			},
		},
		Performance: &PerformanceConfig{
			WorkerPool: &WorkerPoolConfig{
				Enabled:     true,
				WorkerCount: 10,
				QueueSize:   100,
				Timeout:     30 * time.Second,
			},
			BatchProcess: &BatchProcessConfig{
				Enabled:       true,
				BatchSize:     50,
				FlushInterval: 5 * time.Second,
				MaxBatches:    10,
			},
			Cache: &CacheConfig{
				Enabled:         true,
				DefaultTTL:      15 * time.Minute,
				CleanupInterval: 5 * time.Minute,
				MaxSize:         1000,
			},
			ConnectionPool: &ConnectionPoolConfig{
				MaxConnections: 25,
				MinConnections: 5,
				IdleTimeout:    5 * time.Minute,
				ConnectTimeout: 10 * time.Second,
				MaxLifetime:    30 * time.Minute,
			},
		},
	}
}

// ValidateConfig validates the notification configuration
func (c *NotificationConfig) Validate() error {
	if c == nil {
		return domain.NewValidationError("configuration cannot be nil")
	}

	if c.Version == "" {
		return domain.NewValidationError("version is required")
	}

	if c.MaxRetries < 0 {
		return domain.NewValidationError("max retries cannot be negative")
	}

	if c.RetryDelay < time.Second {
		return domain.NewValidationError("retry delay must be at least 1 second")
	}

	if c.BatchSize <= 0 {
		return domain.NewValidationError("batch size must be positive")
	}

	if c.ProcessingTimeout < 5*time.Second {
		return domain.NewValidationError("processing timeout must be at least 5 seconds")
	}

	// Validate database configuration
	if err := c.validateDatabaseConfig(); err != nil {
		return fmt.Errorf("database config validation failed: %w", err)
	}

	// Validate message queue configuration
	if err := c.validateMessageQueueConfig(); err != nil {
		return fmt.Errorf("message queue config validation failed: %w", err)
	}

	// Validate handler configurations
	if err := c.validateHandlerConfigs(); err != nil {
		return fmt.Errorf("handler config validation failed: %w", err)
	}

	// Validate observability configuration
	if err := c.validateObservabilityConfig(); err != nil {
		return fmt.Errorf("observability config validation failed: %w", err)
	}

	return nil
}

// validateDatabaseConfig validates database configuration
func (c *NotificationConfig) validateDatabaseConfig() error {
	if c.Database == nil {
		return domain.NewValidationError("database configuration is required")
	}

	if c.Database.ConnectionString == "" {
		return domain.NewValidationError("database connection string is required")
	}

	if c.Database.MaxOpenConnections <= 0 {
		return domain.NewValidationError("max open connections must be positive")
	}

	if c.Database.MaxIdleConnections < 0 {
		return domain.NewValidationError("max idle connections cannot be negative")
	}

	if c.Database.MaxIdleConnections > c.Database.MaxOpenConnections {
		return domain.NewValidationError("max idle connections cannot exceed max open connections")
	}

	if c.Database.ConnectionTimeout < time.Second {
		return domain.NewValidationError("connection timeout must be at least 1 second")
	}

	if c.Database.QueryTimeout < time.Second {
		return domain.NewValidationError("query timeout must be at least 1 second")
	}

	return nil
}

// validateMessageQueueConfig validates message queue configuration
func (c *NotificationConfig) validateMessageQueueConfig() error {
	if c.MessageQueue == nil {
		return domain.NewValidationError("message queue configuration is required")
	}

	if c.MessageQueue.ConnectionString == "" {
		return domain.NewValidationError("message queue connection string is required")
	}

	if c.MessageQueue.MaxRetries < 0 {
		return domain.NewValidationError("message queue max retries cannot be negative")
	}

	if c.MessageQueue.RetryDelay < time.Second {
		return domain.NewValidationError("message queue retry delay must be at least 1 second")
	}

	if c.MessageQueue.MessageTimeout < 5*time.Second {
		return domain.NewValidationError("message timeout must be at least 5 seconds")
	}

	if c.MessageQueue.PrefetchCount <= 0 {
		return domain.NewValidationError("prefetch count must be positive")
	}

	return nil
}

// validateHandlerConfigs validates notification handler configurations
func (c *NotificationConfig) validateHandlerConfigs() error {
	// At least one handler must be enabled
	if !c.Email.Enabled && !c.SMS.Enabled && !c.Slack.Enabled {
		return domain.NewValidationError("at least one notification handler must be enabled")
	}

	// Validate email handler config
	if c.Email.Enabled {
		if err := c.validateEmailHandlerConfig(); err != nil {
			return fmt.Errorf("email handler validation failed: %w", err)
		}
	}

	// Validate SMS handler config
	if c.SMS.Enabled {
		if err := c.validateSMSHandlerConfig(); err != nil {
			return fmt.Errorf("SMS handler validation failed: %w", err)
		}
	}

	// Validate Slack handler config
	if c.Slack.Enabled {
		if err := c.validateSlackHandlerConfig(); err != nil {
			return fmt.Errorf("Slack handler validation failed: %w", err)
		}
	}

	return nil
}

// validateEmailHandlerConfig validates email handler configuration
func (c *NotificationConfig) validateEmailHandlerConfig() error {
	if c.Email.QueueName == "" {
		return domain.NewValidationError("email queue name is required")
	}

	if c.Email.Workers <= 0 {
		return domain.NewValidationError("email workers must be positive")
	}

	if c.Email.ProcessingDelay < 0 {
		return domain.NewValidationError("email processing delay cannot be negative")
	}

	if c.Email.Azure == nil {
		return domain.NewValidationError("Azure email configuration is required")
	}

	if c.Email.Azure.ConnectionString == "" {
		return domain.NewValidationError("Azure email connection string is required")
	}

	if c.Email.Azure.SenderAddress == "" {
		return domain.NewValidationError("Azure email sender address is required")
	}

	return nil
}

// validateSMSHandlerConfig validates SMS handler configuration
func (c *NotificationConfig) validateSMSHandlerConfig() error {
	if c.SMS.QueueName == "" {
		return domain.NewValidationError("SMS queue name is required")
	}

	if c.SMS.Workers <= 0 {
		return domain.NewValidationError("SMS workers must be positive")
	}

	if c.SMS.ProcessingDelay < 0 {
		return domain.NewValidationError("SMS processing delay cannot be negative")
	}

	if c.SMS.Azure == nil {
		return domain.NewValidationError("Azure SMS configuration is required")
	}

	if c.SMS.Azure.ConnectionString == "" {
		return domain.NewValidationError("Azure SMS connection string is required")
	}

	if c.SMS.Azure.FromNumber == "" {
		return domain.NewValidationError("Azure SMS from number is required")
	}

	return nil
}

// validateSlackHandlerConfig validates Slack handler configuration
func (c *NotificationConfig) validateSlackHandlerConfig() error {
	if c.Slack.QueueName == "" {
		return domain.NewValidationError("Slack queue name is required")
	}

	if c.Slack.Workers <= 0 {
		return domain.NewValidationError("Slack workers must be positive")
	}

	if c.Slack.ProcessingDelay < 0 {
		return domain.NewValidationError("Slack processing delay cannot be negative")
	}

	if c.Slack.Slack == nil {
		return domain.NewValidationError("Slack configuration is required")
	}

	if c.Slack.Slack.BotToken == "" {
		return domain.NewValidationError("Slack bot token is required")
	}

	if c.Slack.Slack.DefaultChannel == "" {
		return domain.NewValidationError("Slack default channel is required")
	}

	return nil
}

// validateObservabilityConfig validates observability configuration
func (c *NotificationConfig) validateObservabilityConfig() error {
	if c.Observability == nil {
		return domain.NewValidationError("observability configuration is required")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[c.Observability.LogLevel] {
		return domain.NewValidationError("invalid log level")
	}

	if c.Observability.HealthCheckPort <= 0 || c.Observability.HealthCheckPort > 65535 {
		return domain.NewValidationError("invalid health check port")
	}

	if c.Observability.MetricsEnabled {
		if c.Observability.MetricsPort <= 0 || c.Observability.MetricsPort > 65535 {
			return domain.NewValidationError("invalid metrics port")
		}

		if c.Observability.MetricsPort == c.Observability.HealthCheckPort {
			return domain.NewValidationError("metrics port cannot be the same as health check port")
		}
	}

	return nil
}

// GetEnabledHandlers returns a list of enabled notification handlers
func (c *NotificationConfig) GetEnabledHandlers() []string {
	var handlers []string

	if c.Email.Enabled {
		handlers = append(handlers, "email")
	}

	if c.SMS.Enabled {
		handlers = append(handlers, "sms")
	}

	if c.Slack.Enabled {
		handlers = append(handlers, "slack")
	}

	return handlers
}

// GetTotalWorkers returns the total number of workers across all handlers
func (c *NotificationConfig) GetTotalWorkers() int {
	total := 0

	if c.Email.Enabled {
		total += c.Email.Workers
	}

	if c.SMS.Enabled {
		total += c.SMS.Workers
	}

	if c.Slack.Enabled {
		total += c.Slack.Workers
	}

	return total
}

// IsProduction checks if the configuration is for production environment
func (c *NotificationConfig) IsProduction() bool {
	// Production characteristics:
	// - Higher worker counts
	// - Longer timeouts
	// - Observability enabled
	// - Multiple handlers enabled

	return c.GetTotalWorkers() >= 10 &&
		   c.ProcessingTimeout >= 30*time.Second &&
		   c.Observability.MetricsEnabled &&
		   c.Observability.TracingEnabled &&
		   len(c.GetEnabledHandlers()) >= 2
}

// GetConfigSummary returns a summary of the configuration for logging
func (c *NotificationConfig) GetConfigSummary() map[string]interface{} {
	return map[string]interface{}{
		"version":            c.Version,
		"max_retries":        c.MaxRetries,
		"batch_size":         c.BatchSize,
		"processing_timeout": c.ProcessingTimeout.String(),
		"enabled_handlers":   c.GetEnabledHandlers(),
		"total_workers":      c.GetTotalWorkers(),
		"database": map[string]interface{}{
			"max_open_connections": c.Database.MaxOpenConnections,
			"max_idle_connections": c.Database.MaxIdleConnections,
			"connection_timeout":   c.Database.ConnectionTimeout.String(),
		},
		"observability": map[string]interface{}{
			"log_level":         c.Observability.LogLevel,
			"metrics_enabled":   c.Observability.MetricsEnabled,
			"tracing_enabled":   c.Observability.TracingEnabled,
			"health_check_port": c.Observability.HealthCheckPort,
		},
		"is_production": c.IsProduction(),
	}
}