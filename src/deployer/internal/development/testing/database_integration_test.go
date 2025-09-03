package testing

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabaseConnectivity validates PostgreSQL database connectivity and basic operations
func TestDatabaseConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	// Arrange
	databaseURL := getRequiredEnvVar(t, "DATABASE_URL")

	// Act - Connect to database
	db, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err, "Should be able to open database connection")
	defer db.Close()

	// Assert - Database connection should work
	err = db.PingContext(ctx)
	require.NoError(t, err, "Database should be accessible and responsive")

	// Test basic query execution
	var result int
	err = db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err, "Should be able to execute basic queries")
	assert.Equal(t, 1, result, "Query result should be correct")
}

// TestDatabaseSchemaCompliance validates database schemas match the specification
func TestDatabaseSchemaCompliance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	// Arrange
	databaseURL := getRequiredEnvVar(t, "DATABASE_URL")
	db, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err)
	defer db.Close()

	// Test Services domain schema compliance
	t.Run("Services_Schema_Compliance", func(t *testing.T) {
		// Verify services table exists with correct structure
		t.Run("Services_Table_Structure", func(t *testing.T) {
			tableExists := checkTableExists(t, ctx, db, "services")
			assert.True(t, tableExists, "Services table must exist")

			if tableExists {
				// Verify required columns exist with correct types
				expectedColumns := map[string]string{
					"service_id":        "uuid",
					"title":            "character varying",
					"description":      "text", 
					"slug":             "character varying",
					"content_url":      "character varying",
					"category_id":      "uuid",
					"image_url":        "character varying",
					"order_number":     "integer",
					"delivery_mode":    "character varying",
					"publishing_status": "character varying",
					"created_on":       "timestamp with time zone",
					"created_by":       "character varying",
					"modified_on":      "timestamp with time zone",
					"modified_by":      "character varying",
					"is_deleted":       "boolean",
					"deleted_on":       "timestamp with time zone",
					"deleted_by":       "character varying",
				}

				validateTableColumns(t, ctx, db, "services", expectedColumns)
			}
		})

		// Verify service_categories table exists with correct structure
		t.Run("ServiceCategories_Table_Structure", func(t *testing.T) {
			tableExists := checkTableExists(t, ctx, db, "service_categories")
			assert.True(t, tableExists, "Service categories table must exist")

			if tableExists {
				expectedColumns := map[string]string{
					"category_id":              "uuid",
					"name":                    "character varying",
					"slug":                    "character varying", 
					"order_number":            "integer",
					"is_default_unassigned":   "boolean",
					"created_on":              "timestamp with time zone",
					"created_by":              "character varying",
					"modified_on":             "timestamp with time zone",
					"modified_by":             "character varying",
					"is_deleted":              "boolean",
					"deleted_on":              "timestamp with time zone",
					"deleted_by":              "character varying",
				}

				validateTableColumns(t, ctx, db, "service_categories", expectedColumns)
			}
		})

		// Verify featured_categories table exists with correct structure
		t.Run("FeaturedCategories_Table_Structure", func(t *testing.T) {
			tableExists := checkTableExists(t, ctx, db, "featured_categories")
			assert.True(t, tableExists, "Featured categories table must exist")

			if tableExists {
				expectedColumns := map[string]string{
					"featured_category_id": "uuid",
					"category_id":          "uuid",
					"feature_position":     "integer",
					"created_on":           "timestamp with time zone",
					"created_by":           "character varying",
					"modified_on":          "timestamp with time zone",
					"modified_by":          "character varying",
				}

				validateTableColumns(t, ctx, db, "featured_categories", expectedColumns)
			}
		})
	})

	// Test Content domain schema compliance
	t.Run("Content_Schema_Compliance", func(t *testing.T) {
		// Verify content table exists with correct structure
		t.Run("Content_Table_Structure", func(t *testing.T) {
			tableExists := checkTableExists(t, ctx, db, "content")
			assert.True(t, tableExists, "Content table must exist")

			if tableExists {
				expectedColumns := map[string]string{
					"content_id":              "uuid",
					"original_filename":       "character varying",
					"file_size":              "bigint",
					"mime_type":              "character varying",
					"content_hash":           "character varying",
					"storage_path":           "character varying",
					"upload_status":          "character varying",
					"alt_text":               "character varying",
					"description":            "text",
					"tags":                   "text[]",
					"content_category":       "character varying",
					"access_level":           "character varying",
					"upload_correlation_id":  "uuid",
					"processing_attempts":    "integer",
					"last_processed_at":     "timestamp with time zone",
					"created_on":            "timestamp with time zone",
					"created_by":            "character varying",
					"modified_on":           "timestamp with time zone",
					"modified_by":           "character varying",
					"is_deleted":            "boolean",
					"deleted_on":            "timestamp with time zone",
					"deleted_by":            "character varying",
				}

				validateTableColumns(t, ctx, db, "content", expectedColumns)
			}
		})

		// Verify content_access_log table exists with correct structure
		t.Run("ContentAccessLog_Table_Structure", func(t *testing.T) {
			tableExists := checkTableExists(t, ctx, db, "content_access_log")
			assert.True(t, tableExists, "Content access log table must exist")

			if tableExists {
				expectedColumns := map[string]string{
					"access_id":         "uuid",
					"content_id":        "uuid",
					"access_timestamp":  "timestamp with time zone",
					"user_id":          "character varying",
					"client_ip":        "inet",
					"user_agent":       "text",
					"access_type":      "character varying",
					"http_status_code": "integer",
					"bytes_served":     "bigint",
					"response_time_ms": "integer",
					"correlation_id":   "uuid",
					"referer_url":      "text",
					"cache_hit":        "boolean",
					"storage_backend":  "character varying",
				}

				validateTableColumns(t, ctx, db, "content_access_log", expectedColumns)
			}
		})

		// Verify content_virus_scan table exists with correct structure
		t.Run("ContentVirusScan_Table_Structure", func(t *testing.T) {
			tableExists := checkTableExists(t, ctx, db, "content_virus_scan")
			assert.True(t, tableExists, "Content virus scan table must exist")

			if tableExists {
				expectedColumns := map[string]string{
					"scan_id":          "uuid",
					"content_id":       "uuid",
					"scan_timestamp":   "timestamp with time zone",
					"scanner_engine":   "character varying",
					"scanner_version":  "character varying",
					"scan_status":      "character varying",
					"threats_detected": "text[]",
					"scan_duration_ms": "integer",
					"created_on":       "timestamp with time zone",
					"correlation_id":   "uuid",
				}

				validateTableColumns(t, ctx, db, "content_virus_scan", expectedColumns)
			}
		})

		// Verify content_storage_backend table exists with correct structure
		t.Run("ContentStorageBackend_Table_Structure", func(t *testing.T) {
			tableExists := checkTableExists(t, ctx, db, "content_storage_backend")
			assert.True(t, tableExists, "Content storage backend table must exist")

			if tableExists {
				expectedColumns := map[string]string{
					"backend_id":                 "uuid",
					"backend_name":               "character varying",
					"backend_type":               "character varying",
					"is_active":                  "boolean",
					"priority_order":             "integer",
					"base_url":                   "character varying",
					"access_key_vault_reference": "character varying",
					"configuration_json":         "jsonb",
					"last_health_check":         "timestamp with time zone",
					"health_status":             "character varying",
					"created_on":                "timestamp with time zone",
					"created_by":                "character varying",
					"modified_on":               "timestamp with time zone",
					"modified_by":               "character varying",
				}

				validateTableColumns(t, ctx, db, "content_storage_backend", expectedColumns)
			}
		})
	})
}

// TestDatabaseConstraints validates database constraints and business rules
func TestDatabaseConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()

	// Arrange
	databaseURL := getRequiredEnvVar(t, "DATABASE_URL")
	db, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err)
	defer db.Close()

	// Test primary key constraints exist
	t.Run("Primary_Key_Constraints", func(t *testing.T) {
		tables := []string{"services", "service_categories", "featured_categories", "content", "content_access_log", "content_virus_scan", "content_storage_backend"}
		
		for _, tableName := range tables {
			t.Run(fmt.Sprintf("%s_primary_key", tableName), func(t *testing.T) {
				if !checkTableExists(t, ctx, db, tableName) {
					t.Skipf("Table %s does not exist", tableName)
				}

				hasPrimaryKey := checkPrimaryKeyExists(t, ctx, db, tableName)
				assert.True(t, hasPrimaryKey, "Table %s must have a primary key constraint", tableName)
			})
		}
	})

	// Test foreign key constraints exist
	t.Run("Foreign_Key_Constraints", func(t *testing.T) {
		foreignKeys := map[string][]string{
			"services":                {"category_id"},
			"featured_categories":     {"category_id"},
			"content_access_log":      {"content_id"},
			"content_virus_scan":      {"content_id"},
		}

		for tableName, columns := range foreignKeys {
			t.Run(fmt.Sprintf("%s_foreign_keys", tableName), func(t *testing.T) {
				if !checkTableExists(t, ctx, db, tableName) {
					t.Skipf("Table %s does not exist", tableName)
				}

				for _, column := range columns {
					hasForeignKey := checkForeignKeyExists(t, ctx, db, tableName, column)
					assert.True(t, hasForeignKey, "Table %s column %s must have foreign key constraint", tableName, column)
				}
			})
		}
	})

	// Test unique constraints exist
	t.Run("Unique_Constraints", func(t *testing.T) {
		uniqueConstraints := map[string][]string{
			"services":                {"slug"},
			"service_categories":      {"slug"},
			"featured_categories":     {"feature_position"},
			"content_storage_backend": {"backend_name"},
		}

		for tableName, columns := range uniqueConstraints {
			t.Run(fmt.Sprintf("%s_unique_constraints", tableName), func(t *testing.T) {
				if !checkTableExists(t, ctx, db, tableName) {
					t.Skipf("Table %s does not exist", tableName)
				}

				for _, column := range columns {
					hasUniqueConstraint := checkUniqueConstraintExists(t, ctx, db, tableName, column)
					assert.True(t, hasUniqueConstraint, "Table %s column %s must have unique constraint", tableName, column)
				}
			})
		}
	})
}

// Test helper functions for database validation

// checkTableExists verifies that a table exists in the database
func checkTableExists(t *testing.T, ctx context.Context, db *sql.DB, tableName string) bool {
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = $1
	)`
	
	var exists bool
	err := db.QueryRowContext(ctx, query, tableName).Scan(&exists)
	require.NoError(t, err, "Should be able to check if table exists")
	
	return exists
}

// validateTableColumns verifies table has expected columns with correct data types
func validateTableColumns(t *testing.T, ctx context.Context, db *sql.DB, tableName string, expectedColumns map[string]string) {
	query := `SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_schema = 'public' AND table_name = $1`
	
	rows, err := db.QueryContext(ctx, query, tableName)
	require.NoError(t, err, "Should be able to query table columns")
	defer rows.Close()

	actualColumns := make(map[string]string)
	for rows.Next() {
		var columnName, dataType string
		err := rows.Scan(&columnName, &dataType)
		require.NoError(t, err, "Should be able to scan column information")
		actualColumns[columnName] = dataType
	}

	// Verify all expected columns exist with correct types
	for expectedColumn, expectedType := range expectedColumns {
		actualType, exists := actualColumns[expectedColumn]
		assert.True(t, exists, "Table %s should have column %s", tableName, expectedColumn)
		if exists {
			// Handle type variations (e.g., varchar vs character varying)
			assert.True(t, 
				strings.Contains(actualType, expectedType) || strings.Contains(expectedType, actualType),
				"Table %s column %s should have type %s, but has type %s", 
				tableName, expectedColumn, expectedType, actualType)
		}
	}
}

// checkPrimaryKeyExists verifies that a table has a primary key constraint
func checkPrimaryKeyExists(t *testing.T, ctx context.Context, db *sql.DB, tableName string) bool {
	query := `SELECT EXISTS (
		SELECT FROM information_schema.table_constraints 
		WHERE table_schema = 'public' AND table_name = $1 AND constraint_type = 'PRIMARY KEY'
	)`
	
	var exists bool
	err := db.QueryRowContext(ctx, query, tableName).Scan(&exists)
	require.NoError(t, err, "Should be able to check for primary key constraint")
	
	return exists
}

// checkForeignKeyExists verifies that a foreign key constraint exists for a column
func checkForeignKeyExists(t *testing.T, ctx context.Context, db *sql.DB, tableName, columnName string) bool {
	query := `SELECT EXISTS (
		SELECT FROM information_schema.key_column_usage kcu
		JOIN information_schema.table_constraints tc ON kcu.constraint_name = tc.constraint_name
		WHERE tc.table_schema = 'public' AND tc.table_name = $1 
		AND kcu.column_name = $2 AND tc.constraint_type = 'FOREIGN KEY'
	)`
	
	var exists bool
	err := db.QueryRowContext(ctx, query, tableName, columnName).Scan(&exists)
	require.NoError(t, err, "Should be able to check for foreign key constraint")
	
	return exists
}

// checkUniqueConstraintExists verifies that a unique constraint exists for a column
func checkUniqueConstraintExists(t *testing.T, ctx context.Context, db *sql.DB, tableName, columnName string) bool {
	query := `SELECT EXISTS (
		SELECT FROM information_schema.key_column_usage kcu
		JOIN information_schema.table_constraints tc ON kcu.constraint_name = tc.constraint_name
		WHERE tc.table_schema = 'public' AND tc.table_name = $1 
		AND kcu.column_name = $2 AND tc.constraint_type = 'UNIQUE'
	)`
	
	var exists bool
	err := db.QueryRowContext(ctx, query, tableName, columnName).Scan(&exists)
	require.NoError(t, err, "Should be able to check for unique constraint")
	
	return exists
}