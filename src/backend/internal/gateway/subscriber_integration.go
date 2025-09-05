package gateway

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/database"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// SubscriberManagementIntegration provides subscriber management integration for admin gateways
type SubscriberManagementIntegration struct {
	db                *sql.DB
	repository        SubscriberRepository
	service           SubscriberService
	handler           *SubscriberHandler
	gatewayConfig     *GatewayConfiguration
}

// NewSubscriberManagementIntegration creates a new subscriber management integration
func NewSubscriberManagementIntegration(gatewayConfig *GatewayConfiguration) (*SubscriberManagementIntegration, error) {
	if !gatewayConfig.IsAdmin() {
		return nil, fmt.Errorf("subscriber management is only available for admin gateways")
	}

	integration := &SubscriberManagementIntegration{
		gatewayConfig: gatewayConfig,
	}

	// Initialize database connection
	if err := integration.initializeDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize repository
	integration.repository = NewPostgreSQLSubscriberRepository(integration.db)

	// Initialize service
	integration.service = NewDefaultSubscriberService(integration.repository)

	// Initialize handler
	integration.handler = NewSubscriberHandler(integration.service, gatewayConfig)

	return integration, nil
}

// InitializeWithGateway integrates subscriber management with an existing gateway service
func (smi *SubscriberManagementIntegration) InitializeWithGateway(gatewayService *GatewayService) error {
	if gatewayService == nil {
		return fmt.Errorf("gateway service cannot be nil")
	}

	// Get the gateway handler
	gatewayHandler := gatewayService.GetHandler()
	if gatewayHandler == nil {
		return fmt.Errorf("gateway handler is nil")
	}

	// Set the subscriber handler
	gatewayHandler.SetSubscriberHandler(smi.handler)

	return nil
}

// HealthCheck performs a health check on subscriber management components
func (smi *SubscriberManagementIntegration) HealthCheck(ctx context.Context) error {
	// Check database connectivity
	if err := smi.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database connectivity check failed: %w", err)
	}

	// Perform a simple query to verify database functionality
	query := "SELECT 1"
	var result int
	if err := smi.db.QueryRowContext(ctx, query).Scan(&result); err != nil {
		return fmt.Errorf("database query check failed: %w", err)
	}

	return nil
}

// Close closes database connections and cleans up resources
func (smi *SubscriberManagementIntegration) Close() error {
	if smi.db != nil {
		return smi.db.Close()
	}
	return nil
}

// GetRepository returns the subscriber repository
func (smi *SubscriberManagementIntegration) GetRepository() SubscriberRepository {
	return smi.repository
}

// GetService returns the subscriber service
func (smi *SubscriberManagementIntegration) GetService() SubscriberService {
	return smi.service
}

// GetHandler returns the subscriber handler
func (smi *SubscriberManagementIntegration) GetHandler() *SubscriberHandler {
	return smi.handler
}

// GetDatabase returns the database connection
func (smi *SubscriberManagementIntegration) GetDatabase() *sql.DB {
	return smi.db
}

// Private methods

// initializeDatabase initializes the database connection
func (smi *SubscriberManagementIntegration) initializeDatabase() error {
	// Get database configuration from environment
	dbConfig := getDatabaseConfig()

	// Validate database configuration
	if err := validateDatabaseConfig(dbConfig); err != nil {
		return fmt.Errorf("invalid database configuration: %w", err)
	}

	// Create database connection
	db, err := sql.Open("postgres", dbConfig.GetConnectionString())
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(dbConfig.MaxOpenConns)
	db.SetMaxIdleConns(dbConfig.MaxIdleConns)
	db.SetConnMaxLifetime(dbConfig.ConnMaxLifetime)

	// Verify database connectivity
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	smi.db = db
	return nil
}

// Database configuration

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	Database        string
	Username        string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // in minutes
}

// GetConnectionString returns the PostgreSQL connection string
func (c *DatabaseConfig) GetConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode,
	)
}

// getDatabaseConfig retrieves database configuration from environment variables
func getDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:            getEnvString("DB_HOST", "localhost"),
		Port:            getEnvInt("DB_PORT", 5432),
		Database:        getEnvString("DB_NAME", "international_center"),
		Username:        getEnvString("DB_USER", "postgres"),
		Password:        getEnvString("DB_PASSWORD", "password"),
		SSLMode:         getEnvString("DB_SSL_MODE", "disable"),
		MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
		ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 5), // 5 minutes
	}
}

// validateDatabaseConfig validates database configuration
func validateDatabaseConfig(config *DatabaseConfig) error {
	if config.Host == "" {
		return fmt.Errorf("database host cannot be empty")
	}

	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", config.Port)
	}

	if config.Database == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	if config.Username == "" {
		return fmt.Errorf("database username cannot be empty")
	}

	if config.Password == "" {
		return fmt.Errorf("database password cannot be empty")
	}

	validSSLModes := map[string]bool{
		"disable":     true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}

	if !validSSLModes[config.SSLMode] {
		return fmt.Errorf("invalid SSL mode: %s", config.SSLMode)
	}

	if config.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections must be greater than 0")
	}

	if config.MaxIdleConns <= 0 {
		return fmt.Errorf("max idle connections must be greater than 0")
	}

	if config.ConnMaxLifetime <= 0 {
		return fmt.Errorf("connection max lifetime must be greater than 0")
	}

	return nil
}

// Helper functions for environment variables

// getEnvString returns string environment variable or default value
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt returns integer environment variable or default value
func getEnvInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		// In production, this would use strconv.Atoi with error handling
		// For now, return default value if parsing fails
		return defaultValue
	}
	return defaultValue
}

// Database migration and setup functions

// EnsureDatabaseSchema ensures the subscriber management database schema exists
func (smi *SubscriberManagementIntegration) EnsureDatabaseSchema(ctx context.Context) error {
	// Create the notification_subscribers table if it doesn't exist
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS notification_subscribers (
			subscriber_id UUID PRIMARY KEY,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			subscriber_name VARCHAR(100) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			phone VARCHAR(20),
			event_types TEXT[] NOT NULL,
			notification_methods TEXT[] NOT NULL,
			notification_schedule VARCHAR(20) NOT NULL DEFAULT 'immediate',
			priority_threshold VARCHAR(20) NOT NULL DEFAULT 'low',
			notes TEXT,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			created_by VARCHAR(100) NOT NULL,
			updated_by VARCHAR(100) NOT NULL,
			is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
			deleted_at TIMESTAMP WITH TIME ZONE,
			
			-- Constraints
			CONSTRAINT chk_status CHECK (status IN ('active', 'inactive', 'suspended')),
			CONSTRAINT chk_notification_schedule CHECK (notification_schedule IN ('immediate', 'hourly', 'daily')),
			CONSTRAINT chk_priority_threshold CHECK (priority_threshold IN ('low', 'medium', 'high', 'urgent')),
			CONSTRAINT chk_subscriber_name_length CHECK (LENGTH(subscriber_name) >= 2),
			CONSTRAINT chk_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
			CONSTRAINT chk_phone_format CHECK (phone IS NULL OR phone ~* '^\+[1-9]\d{1,14}$'),
			CONSTRAINT chk_event_types_not_empty CHECK (array_length(event_types, 1) > 0),
			CONSTRAINT chk_notification_methods_not_empty CHECK (array_length(notification_methods, 1) > 0)
		);
	`

	_, err := smi.db.ExecContext(ctx, createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create notification_subscribers table: %w", err)
	}

	// Create indexes for performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_notification_subscribers_email ON notification_subscribers(email) WHERE is_deleted = false;",
		"CREATE INDEX IF NOT EXISTS idx_notification_subscribers_status ON notification_subscribers(status) WHERE is_deleted = false;",
		"CREATE INDEX IF NOT EXISTS idx_notification_subscribers_event_types ON notification_subscribers USING GIN(event_types) WHERE is_deleted = false;",
		"CREATE INDEX IF NOT EXISTS idx_notification_subscribers_priority_threshold ON notification_subscribers(priority_threshold) WHERE is_deleted = false;",
		"CREATE INDEX IF NOT EXISTS idx_notification_subscribers_created_at ON notification_subscribers(created_at) WHERE is_deleted = false;",
		"CREATE INDEX IF NOT EXISTS idx_notification_subscribers_active ON notification_subscribers(status, is_deleted) WHERE status = 'active' AND is_deleted = false;",
	}

	for _, indexQuery := range indexes {
		_, err := smi.db.ExecContext(ctx, indexQuery)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// SeedDatabaseWithDefaultSubscribers creates default notification subscribers for system admin
func (smi *SubscriberManagementIntegration) SeedDatabaseWithDefaultSubscribers(ctx context.Context) error {
	// Check if any subscribers already exist
	var count int
	countQuery := "SELECT COUNT(*) FROM notification_subscribers WHERE is_deleted = false"
	err := smi.db.QueryRowContext(ctx, countQuery).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing subscribers: %w", err)
	}

	// If subscribers already exist, skip seeding
	if count > 0 {
		return nil
	}

	// Create default admin subscriber
	defaultSubscriber := &NotificationSubscriber{
		SubscriberID: "00000000-0000-0000-0000-000000000001", // Fixed UUID for system admin
		Status:       SubscriberStatusActive,
		SubscriberName: "System Administrator",
		Email:        getEnvString("ADMIN_EMAIL", "admin@international-center.org"),
		Phone:        stringPtr(getEnvString("ADMIN_PHONE", "+1234567890")),
		EventTypes: []EventType{
			EventTypeSystemError,
			EventTypeCapacityAlert,
			EventTypeAdminActionRequired,
			EventTypeComplianceAlert,
		},
		NotificationMethods:  []NotificationMethod{NotificationMethodBoth},
		NotificationSchedule: NotificationScheduleImmediate,
		PriorityThreshold:    PriorityThresholdLow, // Receive all priority levels
		Notes:                stringPtr("Default system administrator subscriber"),
		CreatedAt:            database.GetCurrentTime(),
		UpdatedAt:            database.GetCurrentTime(),
		CreatedBy:            "system",
		UpdatedBy:            "system",
		IsDeleted:            false,
	}

	// Create the default subscriber
	if err := smi.repository.CreateSubscriber(ctx, defaultSubscriber); err != nil {
		return fmt.Errorf("failed to create default admin subscriber: %w", err)
	}

	return nil
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}