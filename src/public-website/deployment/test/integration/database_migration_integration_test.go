package integration

import (
	"context"
	"database/sql"
	"os/exec"
	"strings"
	"testing"
	"time"

	sharedValidation "github.com/axiom-software-co/international-center/src/public-website/deployment/test/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// RED PHASE: Database Migration Integration Tests
// These tests validate that database migrations execute during deployment process
// and that all services have required database schemas available

func TestDatabaseMigrationIntegration_MigrationExecution(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("MigrationExecution_AutomatedDuringDeployment", func(t *testing.T) {
		// Validate that migrations are executed (not deferred) during deployment
		
		// Check database connectivity first
		connStr := "postgresql://postgres:password@localhost:5432/international_center_development?sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		require.NoError(t, err, "Database must be accessible for migration validation")
		defer db.Close()

		err = db.PingContext(ctx)
		require.NoError(t, err, "Database must be ready for migration validation")

		// Check if schema_migrations table exists (indicates migrations have run)
		var migrationTableExists bool
		err = db.QueryRowContext(ctx, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations')").Scan(&migrationTableExists)
		require.NoError(t, err, "Failed to check for schema_migrations table")

		assert.True(t, migrationTableExists, "Schema migrations table must exist - migrations should be executed during deployment")

		// If migration table exists, check for applied migrations
		if migrationTableExists {
			var migrationCount int
			err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount)
			require.NoError(t, err, "Failed to count applied migrations")

			assert.Greater(t, migrationCount, 0, "Migrations must be applied during deployment (not deferred)")
		}
	})

	t.Run("MigrationExecution_ServiceSchemaRequirements", func(t *testing.T) {
		// Validate that all services have required database schemas
		
		connStr := "postgresql://postgres:password@localhost:5432/international_center_development?sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		require.NoError(t, err, "Database must be accessible for schema validation")
		defer db.Close()

		// Required tables for each service domain
		requiredTables := map[string][]string{
			"content": {
				"content_news",
				"content_events", 
				"content_research",
				"content_metadata",
			},
			"inquiries": {
				"inquiries_business",
				"inquiries_donations",
				"inquiries_media",
				"inquiries_volunteers",
				"inquiry_metadata",
			},
			"notifications": {
				"notification_subscribers",
				"notification_templates",
				"notification_events",
				"notification_delivery_log",
			},
			"gateways": {
				"gateway_rate_limits",
				"gateway_access_log",
				"gateway_configuration",
			},
		}

		// Validate each service has required tables
		for serviceDomain, tables := range requiredTables {
			t.Run("ServiceSchema_"+serviceDomain, func(t *testing.T) {
				for _, tableName := range tables {
					var tableExists bool
					err := db.QueryRowContext(ctx, 
						"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)", 
						tableName).Scan(&tableExists)
					require.NoError(t, err, "Failed to check for table %s", tableName)

					assert.True(t, tableExists, 
						"Service domain %s requires table %s - migrations must create all service schemas", 
						serviceDomain, tableName)
				}
			})
		}
	})
}

func TestDatabaseMigrationIntegration_ServiceDatabaseConnectivity(t *testing.T) {
	// Test that all services can connect to database with proper schema
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Services that require database connectivity
	databaseDependentServices := []struct {
		serviceName     string
		containerName   string
		requiredTables  []string
		description     string
	}{
		{
			serviceName:    "content",
			containerName:  "content",
			requiredTables: []string{"content_news", "content_events", "content_research"},
			description:    "Content service requires content management tables",
		},
		{
			serviceName:    "inquiries", 
			containerName:  "inquiries",
			requiredTables: []string{"inquiries_business", "inquiries_donations"},
			description:    "Inquiries service requires inquiry management tables",
		},
		{
			serviceName:    "notifications",
			containerName:  "notifications", 
			requiredTables: []string{"notification_subscribers", "notification_templates"},
			description:    "Notifications service requires notification management tables",
		},
		{
			serviceName:    "public-gateway",
			containerName:  "public-gateway",
			requiredTables: []string{"gateway_rate_limits", "gateway_access_log"},
			description:    "Public gateway requires gateway management tables",
		},
		{
			serviceName:    "admin-gateway",
			containerName:  "admin-gateway", 
			requiredTables: []string{"gateway_rate_limits", "gateway_configuration"},
			description:    "Admin gateway requires admin gateway tables",
		},
	}

	// Act & Assert: Validate service database connectivity and schema
	for _, service := range databaseDependentServices {
		t.Run("ServiceDatabaseConnectivity_"+service.serviceName, func(t *testing.T) {
			// Check if service is running
			serviceCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+service.containerName, "--format", "{{.Names}}")
			serviceOutput, err := serviceCmd.Output()
			require.NoError(t, err, "Failed to check service %s status", service.containerName)

			runningServices := strings.TrimSpace(string(serviceOutput))

			if strings.Contains(runningServices, service.containerName) {
				// Service is running - check if it can access required database tables
				
				// Test database connectivity from service perspective
				connStr := "postgresql://postgres:password@postgresql:5432/international_center_development?sslmode=disable"
				db, err := sql.Open("postgres", connStr)
				require.NoError(t, err, "Database must be accessible from %s", service.description)
				defer db.Close()

				// Validate service can access its required tables
				for _, tableName := range service.requiredTables {
					var tableExists bool
					err := db.QueryRowContext(ctx, 
						"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)", 
						tableName).Scan(&tableExists)
					require.NoError(t, err, "Failed to check table %s for %s", tableName, service.serviceName)

					assert.True(t, tableExists, 
						"%s requires table %s - migrations must create service-specific schemas", 
						service.description, tableName)
				}

				// Test that service can perform basic database operations
				for _, tableName := range service.requiredTables {
					_, err := db.ExecContext(ctx, "SELECT COUNT(*) FROM "+tableName)
					assert.NoError(t, err, 
						"%s must be able to query table %s", service.description, tableName)
				}
			} else {
				t.Logf("Service %s not running - skipping database connectivity test", service.serviceName)
			}
		})
	}
}

func TestDatabaseMigrationIntegration_MigrationDeploymentCoordination(t *testing.T) {
	// Test that migrations are properly coordinated with deployment phases
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("MigrationDeploymentCoordination_InfrastructurePhase", func(t *testing.T) {
		// Validate that migrations execute during infrastructure phase (not deferred)
		
		connStr := "postgresql://postgres:password@localhost:5432/international_center_development?sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		require.NoError(t, err, "Database must be accessible for migration coordination validation")
		defer db.Close()

		// Check current migration status in database
		var migrationCount int
		err = db.QueryRowContext(ctx, 
			"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE '%_'").Scan(&migrationCount)
		require.NoError(t, err, "Failed to count migrated tables")

		assert.Greater(t, migrationCount, 5, 
			"Migrations must execute during infrastructure phase creating service tables")

		// Validate specific infrastructure tables exist
		infrastructureTables := []string{
			"schema_migrations",
			// Core domain tables that should exist from migrations
		}

		for _, tableName := range infrastructureTables {
			var tableExists bool
			err := db.QueryRowContext(ctx, 
				"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)", 
				tableName).Scan(&tableExists)
			require.NoError(t, err, "Failed to check infrastructure table %s", tableName)

			assert.True(t, tableExists, 
				"Infrastructure table %s must be created during migration execution", tableName)
		}
	})

	t.Run("MigrationDeploymentCoordination_ServicePhaseReadiness", func(t *testing.T) {
		// Validate that database is ready for services when service phase starts
		
		connStr := "postgresql://postgres:password@localhost:5432/international_center_development?sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		require.NoError(t, err, "Database must be ready for services phase")
		defer db.Close()

		// Test that database can handle concurrent service connections
		for i := 0; i < 3; i++ {
			_, err := db.ExecContext(ctx, "SELECT 1")
			assert.NoError(t, err, "Database must handle concurrent service connections")
		}

		// Validate database is ready for service operations
		err = db.PingContext(ctx)
		assert.NoError(t, err, "Database must be ready when services phase starts")
	})
}

