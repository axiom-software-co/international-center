package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Migrator handles database migrations with environment-specific strategies
type Migrator struct {
	Environment   string
	DatabaseURL   string
	MigrationsDir string
	migrate       *migrate.Migrate
}

// NewMigrator creates a new migrator instance for the specified environment
func NewMigrator(environment string) (*Migrator, error) {
	migrator := &Migrator{
		Environment:   environment,
		MigrationsDir: "infrastructure/migrations/migrations",
	}

	// For development environment, use PostgreSQL connection
	if environment == "development" {
		host := getEnvWithDefault("POSTGRES_HOST", "localhost")
		port := getEnvWithDefault("POSTGRES_PORT", "5432")
		user := getEnvWithDefault("POSTGRES_USER", "postgres")
		password := getEnvWithDefault("POSTGRES_PASSWORD", "development")
		dbname := getEnvWithDefault("POSTGRES_DB", "international_center")
		
		migrator.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			user, password, host, port, dbname)
	}

	return migrator, nil
}

// RunMigrations executes database migrations based on environment strategy
func (m *Migrator) RunMigrations(ctx context.Context) error {
	// Initialize migration instance if not already done
	if m.migrate == nil {
		if err := m.initMigrate(); err != nil {
			return fmt.Errorf("failed to initialize migrator: %w", err)
		}
	}

	switch m.Environment {
	case "development":
		return m.aggressiveMigration(ctx)
	case "staging":
		return m.carefulMigration(ctx)
	case "production":
		return m.conservativeMigration(ctx)
	default:
		return fmt.Errorf("unknown environment: %s", m.Environment)
	}
}

// Version returns the current migration version and dirty state
func (m *Migrator) Version() (version uint, dirty bool, err error) {
	if m.migrate == nil {
		if err := m.initMigrate(); err != nil {
			return 0, false, fmt.Errorf("failed to initialize migrator: %w", err)
		}
	}

	return m.migrate.Version()
}

// initMigrate initializes the migrate instance
func (m *Migrator) initMigrate() error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	
	// Build absolute path to migrations directory
	var migrationsDir string
	if filepath.IsAbs(m.MigrationsDir) {
		migrationsDir = m.MigrationsDir
	} else {
		migrationsDir = filepath.Join(cwd, m.MigrationsDir)
	}
	
	// Check if migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", migrationsDir)
	}

	// Create file source with absolute path
	migrationsPath := fmt.Sprintf("file://%s", migrationsDir)
	source, err := (&file.File{}).Open(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to open migration source: %w", err)
	}

	// Create database connection
	db, err := sql.Open("postgres", m.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m.migrate, err = migrate.NewWithInstance("file", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return nil
}

// aggressiveMigration implements the development migration strategy
func (m *Migrator) aggressiveMigration(ctx context.Context) error {
	// Development: Always migrate to latest version
	// Easy rollback via environment destruction/recreation
	
	// Set timeout for aggressive approach
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Always migrate up to latest
	if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("aggressive migration failed: %w", err)
	}

	return nil
}

// carefulMigration implements the staging migration strategy
func (m *Migrator) carefulMigration(ctx context.Context) error {
	// Staging: Careful migration with validation
	// Supported rollback with confirmation
	
	// For now, implement basic migration
	if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("careful migration failed: %w", err)
	}

	return nil
}

// conservativeMigration implements the production migration strategy
func (m *Migrator) conservativeMigration(ctx context.Context) error {
	// Production: Conservative with extensive validation
	// Manual approval required for rollback
	
	// For now, implement basic migration
	if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("conservative migration failed: %w", err)
	}

	return nil
}

// getEnvWithDefault returns environment variable value or default
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}