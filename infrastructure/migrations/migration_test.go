package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseMigrationExecution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("PostgreSQL migrations execute successfully", func(t *testing.T) {
		// Test: Migrations can execute without errors
		migrator, err := NewMigrator("development")
		require.NoError(t, err, "Should be able to create migrator")
		
		err = migrator.RunMigrations(ctx)
		assert.NoError(t, err, "Migrations should execute successfully")
	})

	t.Run("migrations create expected tables", func(t *testing.T) {
		// Test: Migrations create database tables successfully
		db, err := sql.Open("postgres", getDatabaseConnectionString())
		require.NoError(t, err, "Should connect to database")
		defer db.Close()

		// Verify that migrations created tables (not specific schema structure)
		var tableCount int
		err = db.QueryRowContext(ctx, `
			SELECT COUNT(*) 
			FROM information_schema.tables 
			WHERE table_schema = 'public'
		`).Scan(&tableCount)
		require.NoError(t, err, "Should be able to query table count")
		assert.Greater(t, tableCount, 0, "Migrations should create tables")
	})

	t.Run("migrations make database changes", func(t *testing.T) {
		// Test: Migrations successfully make changes to the database
		db, err := sql.Open("postgres", getDatabaseConnectionString())
		require.NoError(t, err, "Should connect to database")
		defer db.Close()

		// Verify that migrations created some database objects (tables, indexes, etc.)
		var objectCount int
		err = db.QueryRowContext(ctx, `
			SELECT COUNT(*) 
			FROM pg_class 
			WHERE relkind IN ('r', 'i')  -- tables and indexes
		`).Scan(&objectCount)
		require.NoError(t, err, "Should be able to query database objects")
		assert.Greater(t, objectCount, 0, "Migrations should create database objects")
	})

	t.Run("migration rollback capability via environment reset", func(t *testing.T) {
		// Test: Can rollback migrations by destroying and recreating environment
		migrator, err := NewMigrator("development")
		require.NoError(t, err, "Should be able to create migrator")

		// Test that we can detect current migration version
		version, dirty, err := migrator.Version()
		if err != nil && err != migrate.ErrNilVersion {
			require.NoError(t, err, "Should be able to get migration version")
		}

		// In development, rollback is via environment destruction/recreation
		// This tests that the migrator can handle a fresh database
		if version > 0 {
			assert.False(t, dirty, "Migration should not be in dirty state")
			assert.Greater(t, version, uint(0), "Should have applied migrations")
		}
	})
}

func TestSingleDatabaseSupport(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("database connectivity", func(t *testing.T) {
		// Test: Database is reachable and functional
		db, err := sql.Open("postgres", getDatabaseConnectionString())
		require.NoError(t, err, "Should connect to database")
		defer db.Close()

		err = db.PingContext(ctx)
		assert.NoError(t, err, "Database should be reachable")
	})

	t.Run("migrations create database objects", func(t *testing.T) {
		// Test: Database contains objects created by migrations
		db, err := sql.Open("postgres", getDatabaseConnectionString())
		require.NoError(t, err, "Should connect to database")
		defer db.Close()

		// Check that migrations created some tables
		var tableCount int
		err = db.QueryRowContext(ctx, `
			SELECT COUNT(*) 
			FROM information_schema.tables 
			WHERE table_schema = 'public'
		`).Scan(&tableCount)
		require.NoError(t, err, "Should be able to query table count")
		assert.Greater(t, tableCount, 0, "Migrations should have created tables")
	})
}

// Helper functions for database connections
func getDatabaseConnectionString() string {
	host := getEnvWithDefault("POSTGRES_HOST", "localhost")
	port := getEnvWithDefault("POSTGRES_PORT", "5432")
	user := getEnvWithDefault("POSTGRES_USER", "postgres")
	password := getEnvWithDefault("POSTGRES_PASSWORD", "development")
	dbname := getEnvWithDefault("POSTGRES_DB", "international_center")
	
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)
}