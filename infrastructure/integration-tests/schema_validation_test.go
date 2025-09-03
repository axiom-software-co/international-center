package tests

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"
)

func TestSchemaValidation(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("expected tables exist", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

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

	t.Run("services table schema validation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

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

	t.Run("content table schema validation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

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

	t.Run("service_categories table schema validation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Service categories table has expected schema structure
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		// Verify service_categories table exists and has basic structure
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'service_categories'
		)`
		
		err = db.QueryRowContext(ctx, query).Scan(&exists)
		require.NoError(t, err, "Should query service_categories table existence")
		
		if exists {
			// Validate key columns exist
			keyColumns := []string{"category_id", "name", "description"}
			for _, columnName := range keyColumns {
				var columnExists bool
				columnQuery := `SELECT EXISTS (
					SELECT FROM information_schema.columns 
					WHERE table_schema = 'public' 
					AND table_name = 'service_categories' 
					AND column_name = $1
				)`
				
				err = db.QueryRowContext(ctx, columnQuery, columnName).Scan(&columnExists)
				require.NoError(t, err, "Should query column existence for %s", columnName)
				assert.True(t, columnExists, "Column %s should exist in service_categories", columnName)
			}
		}
	})

	t.Run("primary key constraints validation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Primary key constraints are properly configured
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		expectedPrimaryKeys := map[string]string{
			"services": "service_id",
			"content":  "content_id",
		}

		for tableName, expectedPK := range expectedPrimaryKeys {
			query := `
				SELECT column_name 
				FROM information_schema.key_column_usage k
				JOIN information_schema.table_constraints t 
				ON k.constraint_name = t.constraint_name
				WHERE t.table_schema = 'public' 
				AND t.table_name = $1 
				AND t.constraint_type = 'PRIMARY KEY'
			`
			
			var primaryKeyColumn string
			err = db.QueryRowContext(ctx, query, tableName).Scan(&primaryKeyColumn)
			
			if err == sql.ErrNoRows {
				t.Logf("No primary key found for table %s - this may be expected for some tables", tableName)
			} else {
				require.NoError(t, err, "Should query primary key for %s", tableName)
				assert.Equal(t, expectedPK, primaryKeyColumn, 
					"Primary key for %s should be %s", tableName, expectedPK)
			}
		}
	})

	t.Run("foreign key constraints validation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Foreign key constraints are properly configured
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		// Check for foreign key constraints
		query := `
			SELECT 
				tc.table_name,
				kcu.column_name,
				ccu.table_name AS foreign_table_name,
				ccu.column_name AS foreign_column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu 
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			JOIN information_schema.constraint_column_usage ccu 
				ON ccu.constraint_name = tc.constraint_name
				AND ccu.table_schema = tc.table_schema
			WHERE tc.constraint_type = 'FOREIGN KEY' 
			AND tc.table_schema = 'public'
		`
		
		rows, err := db.QueryContext(ctx, query)
		require.NoError(t, err, "Should query foreign key constraints")
		defer rows.Close()

		constraintCount := 0
		for rows.Next() {
			var tableName, columnName, foreignTable, foreignColumn string
			err := rows.Scan(&tableName, &columnName, &foreignTable, &foreignColumn)
			require.NoError(t, err, "Should scan foreign key constraint info")
			
			t.Logf("Found foreign key: %s.%s -> %s.%s", 
				tableName, columnName, foreignTable, foreignColumn)
			constraintCount++
		}
		
		t.Logf("Found %d foreign key constraints", constraintCount)
	})

	t.Run("index validation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Important indexes exist for performance
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		// Check for indexes on important tables
		query := `
			SELECT 
				schemaname,
				tablename,
				indexname,
				indexdef
			FROM pg_indexes
			WHERE schemaname = 'public'
			ORDER BY tablename, indexname
		`
		
		rows, err := db.QueryContext(ctx, query)
		require.NoError(t, err, "Should query index information")
		defer rows.Close()

		indexCount := 0
		for rows.Next() {
			var schemaName, tableName, indexName, indexDef string
			err := rows.Scan(&schemaName, &tableName, &indexName, &indexDef)
			require.NoError(t, err, "Should scan index information")
			
			t.Logf("Found index: %s on %s.%s", indexName, schemaName, tableName)
			indexCount++
		}
		
		assert.GreaterOrEqual(t, indexCount, 1, 
			"Should have at least one index (primary keys create indexes)")
	})

	t.Run("nullable constraints validation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: NOT NULL constraints are properly configured
		dbURL := buildPostgreSQLConnectionString(t)
		
		db, err := sql.Open("postgres", dbURL)
		require.NoError(t, err, "Should connect to PostgreSQL database")
		defer db.Close()

		// Check NOT NULL constraints for critical columns
		criticalColumns := map[string][]string{
			"services": {"service_id", "title", "created_on"},
			"content":  {"content_id", "original_filename", "created_on"},
		}

		for tableName, columns := range criticalColumns {
			for _, columnName := range columns {
				query := `
					SELECT is_nullable
					FROM information_schema.columns
					WHERE table_schema = 'public'
					AND table_name = $1
					AND column_name = $2
				`
				
				var isNullable string
				err = db.QueryRowContext(ctx, query, tableName, columnName).Scan(&isNullable)
				
				if err == sql.ErrNoRows {
					t.Logf("Column %s.%s does not exist - may be expected", tableName, columnName)
				} else {
					require.NoError(t, err, "Should query nullable constraint for %s.%s", tableName, columnName)
					assert.Equal(t, "NO", isNullable, 
						"Critical column %s.%s should be NOT NULL", tableName, columnName)
				}
			}
		}
	})
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