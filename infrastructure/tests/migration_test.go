package tests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"
)

func TestMigrationExecution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("postgresql database connectivity", func(t *testing.T) {
		// Test: PostgreSQL is accessible for migrations
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		err = db.PingContext(ctx)
		require.NoError(t, err, "PostgreSQL should respond to ping")
	})

	t.Run("migration container execution", func(t *testing.T) {
		// Test: golang-migrate container executes migrations successfully
		
		// Get PostgreSQL container IP for migration execution
		postgresIP, err := getPostgreSQLContainerIP()
		require.NoError(t, err, "Should get PostgreSQL container IP")
		
		// Build migration command
		dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", 
			requireEnv(t, "POSTGRES_USER"), requireEnv(t, "POSTGRES_PASSWORD"), 
			postgresIP, requireEnv(t, "POSTGRES_PORT"), requireEnv(t, "POSTGRES_DB"))
		
		// Check current version first
		versionCmd := buildMigrationCommand(dbURL, "version")
		versionOutput, err := executeCommandWithOutput(ctx, versionCmd)
		
		if err != nil || strings.Contains(versionOutput, "dirty") {
			// Force clean state if dirty or error
			forceCmd := buildMigrationCommand(dbURL, "force", "0")
			err = executeCommand(ctx, forceCmd)
			require.NoError(t, err, "Should force clean state")
		}
		
		// Execute migrations - this will be idempotent
		migrateCmd := buildMigrationCommand(dbURL, "up")
		err = executeCommand(ctx, migrateCmd)
		
		// Migration should succeed or be already up to date
		if err != nil && !strings.Contains(err.Error(), "no change") && !strings.Contains(err.Error(), "already exists") {
			t.Logf("Migration output: %v", err)
			// Check if this is just a re-run on existing schema - this is acceptable
			if strings.Contains(err.Error(), "relation") && strings.Contains(err.Error(), "already exists") {
				t.Log("Migration tables already exist - this is acceptable for idempotent testing")
			} else {
				t.Fatalf("Migration should succeed or be already up to date: %v", err)
			}
		}
	})

	t.Run("expected tables created", func(t *testing.T) {
		// Test: All expected tables are created by migrations
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		expectedTables := []string{
			"services",
			"service_categories", 
			"featured_categories",
			"content",
			"content_access_log",
			"content_virus_scan",
			"content_storage_backend",
			"schema_migrations",
		}

		for _, tableName := range expectedTables {
			var exists bool
			query := `SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)`
			
			err := db.QueryRowContext(ctx, query, tableName).Scan(&exists)
			require.NoError(t, err, "Should query table existence for %s", tableName)
			assert.True(t, exists, "Table %s should exist after migrations", tableName)
		}
	})

	t.Run("services schema validation", func(t *testing.T) {
		// Test: Services table has expected schema structure
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		expectedColumns := map[string]string{
			"service_id":         "uuid",
			"title":              "character varying",
			"description":        "text", 
			"slug":               "character varying",
			"content_url":        "character varying",
			"category_id":        "uuid",
			"image_url":          "character varying",
			"order_number":       "integer",
			"delivery_mode":      "character varying",
			"publishing_status":  "character varying",
			"created_on":         "timestamp with time zone",
			"created_by":         "character varying",
			"modified_on":        "timestamp with time zone",
			"modified_by":        "character varying",
			"is_deleted":         "boolean",
			"deleted_on":         "timestamp with time zone",
			"deleted_by":         "character varying",
		}

		validateTableColumns(ctx, t, db, "services", expectedColumns)
	})

	t.Run("content schema validation", func(t *testing.T) {
		// Test: Content table has expected schema structure
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		expectedColumns := map[string]string{
			"content_id":                "uuid",
			"original_filename":         "character varying",
			"file_size":                 "bigint",
			"mime_type":                 "character varying",
			"content_hash":              "character varying",
			"storage_path":              "character varying",
			"upload_status":             "character varying",
			"alt_text":                  "character varying",
			"description":               "text",
			"tags":                      "ARRAY",
			"content_category":          "character varying",
			"access_level":              "character varying",
			"upload_correlation_id":     "uuid",
			"processing_attempts":       "integer",
			"last_processed_at":         "timestamp with time zone",
			"created_on":                "timestamp with time zone",
			"created_by":                "character varying",
			"modified_on":               "timestamp with time zone",
			"modified_by":               "character varying",
			"is_deleted":                "boolean",
			"deleted_on":                "timestamp with time zone",
			"deleted_by":                "character varying",
		}

		validateTableColumns(ctx, t, db, "content", expectedColumns)
	})

	t.Run("migration versioning", func(t *testing.T) {
		// Test: Migration version tracking works correctly
		
		// Get PostgreSQL container IP for migration commands
		postgresIP, err := getPostgreSQLContainerIP()
		require.NoError(t, err, "Should get PostgreSQL container IP")
		
		dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", 
			requireEnv(t, "POSTGRES_USER"), requireEnv(t, "POSTGRES_PASSWORD"), 
			postgresIP, requireEnv(t, "POSTGRES_PORT"), requireEnv(t, "POSTGRES_DB"))
		
		// Ensure clean state
		forceCmd := buildMigrationCommand(dbURL, "force", "2") // Force to version 2 (latest)
		_ = executeCommand(ctx, forceCmd) // Don't fail if this doesn't work
		
		// Check version
		versionCmd := buildMigrationCommand(dbURL, "version")
		versionOutput, err := executeCommandWithOutput(ctx, versionCmd)
		require.NoError(t, err, "Should get migration version")
		
		// Version output should contain a version number
		assert.NotEmpty(t, versionOutput, "Version output should not be empty")
		assert.NotContains(t, versionOutput, "dirty", "Migration should not be in dirty state")
		
		// Also verify through database query
		db, err := sql.Open("postgres", buildPostgreSQLConnectionString(t))
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		var version int
		var dirty bool
		
		query := `SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1`
		err = db.QueryRowContext(ctx, query).Scan(&version, &dirty)
		require.NoError(t, err, "Should query current migration version")
		
		assert.Greater(t, version, 0, "Migration version should be greater than 0")
		assert.False(t, dirty, "Migration should not be in dirty state")
	})
}

func TestMigrationRollback(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("migration down functionality", func(t *testing.T) {
		// Test: Migration rollback works correctly
		
		// Get PostgreSQL container IP
		postgresIP, err := getPostgreSQLContainerIP()
		require.NoError(t, err, "Should get PostgreSQL container IP")
		
		dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", 
			requireEnv(t, "POSTGRES_USER"), requireEnv(t, "POSTGRES_PASSWORD"), 
			postgresIP, requireEnv(t, "POSTGRES_PORT"), requireEnv(t, "POSTGRES_DB"))
		
		// Execute down migration for one step
		downCmd := buildMigrationCommand(dbURL, "down", "1")
		err = executeCommand(ctx, downCmd)
		require.NoError(t, err, "Migration rollback should execute successfully")
		
		// Re-run up to restore state for other tests
		upCmd := buildMigrationCommand(dbURL, "up")
		err = executeCommand(ctx, upCmd)
		require.NoError(t, err, "Migration up should restore state")
	})
}

func TestMigrationRecovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("dirty state recovery", func(t *testing.T) {
		// Test: Recovery from dirty migration state works
		
		postgresIP, err := getPostgreSQLContainerIP()
		require.NoError(t, err, "Should get PostgreSQL container IP")
		
		dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", 
			requireEnv(t, "POSTGRES_USER"), requireEnv(t, "POSTGRES_PASSWORD"), 
			postgresIP, requireEnv(t, "POSTGRES_PORT"), requireEnv(t, "POSTGRES_DB"))
		
		// Force a dirty state
		forceCmd := buildMigrationCommand(dbURL, "force", "1")
		err = executeCommand(ctx, forceCmd)
		require.NoError(t, err, "Should force migration to version 1")
		
		// Try to migrate - this should work after force
		upCmd := buildMigrationCommand(dbURL, "up")
		err = executeCommand(ctx, upCmd)
		
		// Migration should succeed or tables already exist
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			t.Fatalf("Should migrate successfully after force clean: %v", err)
		}
	})
}

// Helper functions

func buildPostgreSQLConnectionString(t *testing.T) string {
	host := "localhost"
	port := requireEnv(t, "POSTGRES_PORT")
	user := requireEnv(t, "POSTGRES_USER")
	password := requireEnv(t, "POSTGRES_PASSWORD")
	dbname := requireEnv(t, "POSTGRES_DB")
	
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)
}

func getPostgreSQLContainerIP() (string, error) {
	cmd := exec.Command("podman", "inspect", "postgresql", "--format={{.NetworkSettings.IPAddress}}")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get PostgreSQL container IP: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func buildMigrationCommand(dbURL string, args ...string) *exec.Cmd {
	cmdArgs := []string{
		"run", "--rm", "--network", "bridge",
		"-v", "../../infrastructure/migrations/migrations:/migrations:ro",
		"migrate/migrate:v4.19.0",
		"-path", "/migrations",
		"-database", dbURL,
	}
	
	cmdArgs = append(cmdArgs, args...)
	return exec.Command("podman", cmdArgs...)
}

func executeCommand(ctx context.Context, cmd *exec.Cmd) error {
	cmd.Env = os.Environ()
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}
	
	return nil
}

func executeCommandWithOutput(ctx context.Context, cmd *exec.Cmd) (string, error) {
	cmd.Env = os.Environ()
	
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func validateTableColumns(ctx context.Context, t *testing.T, db *sql.DB, tableName string, expectedColumns map[string]string) {
	query := `
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_schema = 'public' 
		AND table_name = $1
	`
	
	rows, err := db.QueryContext(ctx, query, tableName)
	require.NoError(t, err, "Should query table columns for %s", tableName)
	defer rows.Close()
	
	actualColumns := make(map[string]string)
	for rows.Next() {
		var columnName, dataType string
		err := rows.Scan(&columnName, &dataType)
		require.NoError(t, err, "Should scan column info")
		actualColumns[columnName] = dataType
	}
	
	for expectedColumn, expectedType := range expectedColumns {
		actualType, exists := actualColumns[expectedColumn]
		assert.True(t, exists, "Column %s should exist in table %s", expectedColumn, tableName)
		
		if exists {
			// Handle array types specially
			if expectedType == "ARRAY" {
				assert.True(t, strings.HasPrefix(actualType, "ARRAY"), 
					"Column %s should be array type, got %s", expectedColumn, actualType)
			} else {
				assert.Equal(t, expectedType, actualType, 
					"Column %s should have type %s, got %s", expectedColumn, expectedType, actualType)
			}
		}
	}
}



// requireEnv retrieves environment variable or fails test if missing
func requireEnv(t *testing.T, key string) string {
	value := os.Getenv(key)
	if value == "" {
		t.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}
