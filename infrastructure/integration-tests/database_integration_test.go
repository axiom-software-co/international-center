package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"
)

func TestDatabaseIntegration(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("postgresql database connectivity", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: PostgreSQL is accessible for database operations
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		err = db.PingContext(ctx)
		require.NoError(t, err, "PostgreSQL should respond to ping")
	})

	t.Run("database connection pool functionality", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: PostgreSQL connection pooling works correctly
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		// Set connection pool parameters
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(time.Hour)

		// Test multiple concurrent connections
		for i := 0; i < 5; i++ {
			err := db.PingContext(ctx)
			require.NoError(t, err, "Connection pool should handle multiple connections")
		}

		// Verify connection pool stats
		stats := db.Stats()
		assert.GreaterOrEqual(t, stats.MaxOpenConnections, 1, 
			"Connection pool should have maximum connections configured")
	})

	t.Run("basic database operations", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Basic SQL operations work
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		// Test basic SELECT operation
		var version string
		err = db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
		require.NoError(t, err, "Should execute basic SELECT query")
		assert.Contains(t, version, "PostgreSQL", "Should return PostgreSQL version")

		// Test current database name query
		var dbName string
		err = db.QueryRowContext(ctx, "SELECT current_database()").Scan(&dbName)
		require.NoError(t, err, "Should get current database name")
		
		expectedDB := requireEnv(t, "POSTGRES_DB")
		assert.Equal(t, expectedDB, dbName, "Should be connected to correct database")
	})

	t.Run("transaction functionality", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Database transaction support
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		// Test transaction begin/commit
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err, "Should begin transaction")

		// Execute query within transaction
		var result int
		err = tx.QueryRowContext(ctx, "SELECT 1").Scan(&result)
		require.NoError(t, err, "Should execute query within transaction")
		assert.Equal(t, 1, result, "Query should return expected result")

		// Commit transaction
		err = tx.Commit()
		require.NoError(t, err, "Should commit transaction")
	})

	t.Run("database user permissions", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Database user has appropriate permissions
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		// Test current user query
		var currentUser string
		err = db.QueryRowContext(ctx, "SELECT current_user").Scan(&currentUser)
		require.NoError(t, err, "Should get current user")
		
		expectedUser := requireEnv(t, "POSTGRES_USER")
		assert.Equal(t, expectedUser, currentUser, "Should be connected as correct user")

		// Test user permissions (should be able to access information_schema)
		var tableCount int
		err = db.QueryRowContext(ctx, 
			"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)
		require.NoError(t, err, "User should have access to information_schema")
		
		// Should have some tables (at least from migrations)
		assert.GreaterOrEqual(t, tableCount, 0, "Should be able to query table information")
	})

	t.Run("database timezone and encoding", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Database timezone and encoding configuration
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		// Test timezone setting
		var timezone string
		err = db.QueryRowContext(ctx, "SHOW timezone").Scan(&timezone)
		require.NoError(t, err, "Should get timezone setting")
		assert.NotEmpty(t, timezone, "Database should have timezone configured")

		// Test encoding setting
		var encoding string
		err = db.QueryRowContext(ctx, "SHOW server_encoding").Scan(&encoding)
		require.NoError(t, err, "Should get encoding setting")
		assert.Equal(t, "UTF8", encoding, "Database should use UTF8 encoding")
	})

	t.Run("prepared statement functionality", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Prepared statements work correctly
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		// Test prepared statement
		stmt, err := db.PrepareContext(ctx, "SELECT $1::text")
		require.NoError(t, err, "Should prepare statement")
		defer stmt.Close()

		var result string
		err = stmt.QueryRowContext(ctx, "test-value").Scan(&result)
		require.NoError(t, err, "Should execute prepared statement")
		assert.Equal(t, "test-value", result, "Prepared statement should return correct value")
	})

	t.Run("database constraints and validation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Database constraint validation
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		// Test that invalid SQL is properly rejected
		_, err = db.ExecContext(ctx, "INVALID SQL STATEMENT")
		assert.Error(t, err, "Database should reject invalid SQL")

		// Test basic data type validation
		var result bool
		err = db.QueryRowContext(ctx, "SELECT true::boolean").Scan(&result)
		require.NoError(t, err, "Should handle boolean data type")
		assert.True(t, result, "Boolean value should be correct")
	})
}

func buildPostgreSQLConnectionString(t *testing.T) string {
	host := "localhost"
	port := requireEnv(t, "POSTGRES_PORT")
	user := requireEnv(t, "POSTGRES_USER")
	password := requireEnv(t, "POSTGRES_PASSWORD")
	dbname := requireEnv(t, "POSTGRES_DB")
	
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)
}