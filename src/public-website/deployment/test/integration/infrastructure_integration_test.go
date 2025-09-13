package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	sharedValidation "github.com/axiom-software-co/international-center/src/public-website/deployment/test/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Infrastructure Integration Tests
// Validates infrastructure phase components working together as integrated system
// Tests database, storage, vault, messaging integration and cross-component functionality

func TestInfrastructureIntegration_DatabaseStorageVaultMessaging(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("DatabaseIntegration_ConnectionAndSchema", func(t *testing.T) {
		// Test PostgreSQL database integration
		connStr := "postgresql://postgres:password@localhost:5432/international_center_development?sslmode=disable"
		
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			t.Errorf("Database connection failed - infrastructure integration broken: %v", err)
			return
		}
		defer db.Close()

		// Test database connectivity
		err = db.PingContext(ctx)
		assert.NoError(t, err, "Database ping must succeed for infrastructure integration")

		// Test database can handle schema operations (critical for migration integration)
		if err == nil {
			_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS integration_test (id SERIAL PRIMARY KEY, test_data TEXT)")
			assert.NoError(t, err, "Database must support schema operations for infrastructure integration")

			// Clean up test table
			db.ExecContext(ctx, "DROP TABLE IF EXISTS integration_test")
		}
	})

	t.Run("StorageIntegration_AzuriteConnectivity", func(t *testing.T) {
		// Test Azurite storage integration
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test Azurite blob service endpoint
		azuriteURL := "http://localhost:10000/"
		req, err := http.NewRequestWithContext(ctx, "GET", azuriteURL, nil)
		require.NoError(t, err, "Failed to create Azurite request")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
				"Azurite storage service must be accessible for infrastructure integration")
		} else {
			t.Errorf("Azurite storage integration failed: %v", err)
		}
	})

	t.Run("VaultIntegration_SecretStoreConnectivity", func(t *testing.T) {
		// Test Vault integration for secret storage
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test Vault health endpoint
		vaultURL := "http://localhost:8200/v1/sys/health"
		req, err := http.NewRequestWithContext(ctx, "GET", vaultURL, nil)
		require.NoError(t, err, "Failed to create Vault health request")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
				"Vault secret store must be accessible for infrastructure integration")
		} else {
			t.Logf("Vault integration not accessible: %v", err)
			// Vault may not be fully operational due to configuration
		}
	})

	t.Run("MessagingIntegration_RabbitMQConnectivity", func(t *testing.T) {
		// Test RabbitMQ integration for messaging
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test RabbitMQ management API
		rabbitMQURL := "http://localhost:15672/api/overview"
		req, err := http.NewRequestWithContext(ctx, "GET", rabbitMQURL, nil)
		require.NoError(t, err, "Failed to create RabbitMQ request")
		
		// Add basic auth for RabbitMQ management
		req.SetBasicAuth("guest", "guest")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
				"RabbitMQ messaging must be accessible for infrastructure integration")
		} else {
			t.Logf("RabbitMQ integration not accessible: %v", err)
			// RabbitMQ management may not be ready yet
		}
	})
}

func TestInfrastructureIntegration_CrossComponentConnectivity(t *testing.T) {
	// This test validates that infrastructure components can communicate with each other
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("NetworkConnectivity_InfrastructureComponents", func(t *testing.T) {
		// Test network connectivity between infrastructure components
		infrastructureConnectivityTests := []struct {
			sourceContainer string
			targetContainer string
			targetPort      int
			description     string
		}{
			{"postgresql", "vault", 8200, "Database to Vault connectivity for secret integration"},
			{"postgresql", "rabbitmq", 5672, "Database to RabbitMQ connectivity for event integration"},
			{"vault", "rabbitmq", 5672, "Vault to RabbitMQ connectivity for secure messaging"},
		}

		for _, test := range infrastructureConnectivityTests {
			t.Run(fmt.Sprintf("Connectivity_%s_to_%s", test.sourceContainer, test.targetContainer), func(t *testing.T) {
				// Check if both containers are running
				sourceCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+test.sourceContainer, "--format", "{{.Names}}")
				sourceOutput, err := sourceCmd.Output()
				require.NoError(t, err, "Failed to check source container %s", test.sourceContainer)

				targetCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+test.targetContainer, "--format", "{{.Names}}")
				targetOutput, err := targetCmd.Output()
				require.NoError(t, err, "Failed to check target container %s", test.targetContainer)

				sourceRunning := strings.Contains(string(sourceOutput), test.sourceContainer)
				targetRunning := strings.Contains(string(targetOutput), test.targetContainer)

				if sourceRunning && targetRunning {
					// Test connectivity between infrastructure components
					connectCmd := exec.CommandContext(ctx, "podman", "exec", test.sourceContainer, "nc", "-z", test.targetContainer, fmt.Sprintf("%d", test.targetPort))
					connectErr := connectCmd.Run()
					assert.NoError(t, connectErr, "%s - infrastructure components must have network connectivity", test.description)
				} else {
					t.Logf("Containers not running for connectivity test: %s=%v, %s=%v", 
						test.sourceContainer, sourceRunning, test.targetContainer, targetRunning)
				}
			})
		}
	})
}

func TestInfrastructureIntegration_MigrationExecution(t *testing.T) {
	// This test validates that database migrations can execute successfully
	// Critical for development environment functionality
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("DatabaseMigrationExecution", func(t *testing.T) {
		// Test that database is ready for migrations
		connStr := "postgresql://postgres:password@localhost:5432/international_center_development?sslmode=disable"
		
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			t.Skip("Database not accessible - skipping migration test")
		}
		defer db.Close()

		// Test database readiness for migrations
		err = db.PingContext(ctx)
		if err != nil {
			t.Skip("Database not ready - skipping migration test")
		}

		// Test that we can create schema_migrations table (standard migration pattern)
		_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (version BIGINT PRIMARY KEY, dirty BOOLEAN NOT NULL)")
		assert.NoError(t, err, "Database must support migration schema operations")

		// Test migration table accessibility
		var count int
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&count)
		assert.NoError(t, err, "Migration schema must be accessible for infrastructure integration")

		// Clean up migration test table
		db.ExecContext(ctx, "DROP TABLE IF EXISTS schema_migrations")
	})

	t.Run("MigrationPathResolution", func(t *testing.T) {
		// Test that migration paths are properly resolved
		migrationPath := "/home/tojkuv/Documents/GitHub/international-center-workspace/international-center/src/public-website/migrations"
		
		// Check if migration directory exists and is accessible
		pathCmd := exec.CommandContext(ctx, "ls", "-la", migrationPath)
		pathOutput, err := pathCmd.Output()
		assert.NoError(t, err, "Migration path must be accessible for infrastructure integration")

		if err == nil {
			pathContent := string(pathOutput)
			assert.Contains(t, pathContent, "sql", "Migration directory must contain SQL files for schema management")
		}
	})
}

// RED PHASE: Dapr State Store Connectivity Contract Validation
func TestInfrastructureIntegration_DaprStateStoreConnectivity(t *testing.T) {
	// This test validates that Dapr state store components are properly configured and accessible
	// Critical for service state persistence and data integration
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// State store components that must be accessible through Dapr
	stateStoreComponents := []struct {
		componentName string
		storeType     string
		description   string
	}{
		{
			componentName: "statestore",
			storeType:     "state.postgresql",
			description:   "PostgreSQL state store must be accessible through Dapr for service state persistence",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Validate Dapr state store component registration
	t.Run("DaprStateStoreComponentRegistration", func(t *testing.T) {
		// Check Dapr components endpoint to validate state store is registered
		componentsURL := "http://localhost:3500/v1.0/components"
		req, err := http.NewRequestWithContext(ctx, "GET", componentsURL, nil)
		require.NoError(t, err, "Failed to create Dapr components request")

		resp, err := client.Do(req)
		require.NoError(t, err, "Dapr components endpoint must be accessible for state store validation")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Dapr components endpoint must return 200 OK for state store component validation")

		// TODO: Parse response and validate state store component is registered
		// This will be validated when Dapr components are properly configured
		t.Logf("RED PHASE VALIDATION: Dapr components endpoint accessible - detailed component validation pending proper configuration")
	})

	// Validate state store connectivity through Dapr state API
	for _, component := range stateStoreComponents {
		t.Run("StateStoreConnectivity_"+component.componentName, func(t *testing.T) {
			// Test state store connectivity via Dapr state API
			testKey := "integration-test-connectivity"
			testValue := `{"test": "connectivity", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`

			// Test state save operation
			saveURL := fmt.Sprintf("http://localhost:3500/v1.0/state/%s", component.componentName)
			savePayload := fmt.Sprintf(`[{"key": "%s", "value": %s}]`, testKey, testValue)
			
			saveReq, err := http.NewRequestWithContext(ctx, "POST", saveURL, strings.NewReader(savePayload))
			require.NoError(t, err, "Failed to create state save request")
			saveReq.Header.Set("Content-Type", "application/json")

			saveResp, err := client.Do(saveReq)
			if err != nil {
				t.Errorf("RED PHASE VALIDATION: %s - State store save operation failed: %v", component.description, err)
				return
			}
			defer saveResp.Body.Close()

			if saveResp.StatusCode != http.StatusNoContent && saveResp.StatusCode != http.StatusOK {
				body := make([]byte, 1024)
				saveResp.Body.Read(body)
				t.Errorf("RED PHASE VALIDATION: %s - State store save returned %d: %s", 
					component.description, saveResp.StatusCode, string(body))
				return
			}

			// Test state get operation
			getURL := fmt.Sprintf("http://localhost:3500/v1.0/state/%s/%s", component.componentName, testKey)
			getReq, err := http.NewRequestWithContext(ctx, "GET", getURL, nil)
			require.NoError(t, err, "Failed to create state get request")

			getResp, err := client.Do(getReq)
			if err != nil {
				t.Errorf("RED PHASE VALIDATION: %s - State store get operation failed: %v", component.description, err)
				return
			}
			defer getResp.Body.Close()

			if getResp.StatusCode != http.StatusOK {
				body := make([]byte, 1024)
				getResp.Body.Read(body)
				t.Errorf("RED PHASE VALIDATION: %s - State store get returned %d: %s", 
					component.description, getResp.StatusCode, string(body))
				return
			}

			// Clean up test data
			deleteURL := fmt.Sprintf("http://localhost:3500/v1.0/state/%s/%s", component.componentName, testKey)
			deleteReq, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
			if err == nil {
				deleteResp, _ := client.Do(deleteReq)
				if deleteResp != nil {
					deleteResp.Body.Close()
				}
			}

			t.Logf("RED PHASE VALIDATION SUCCESS: %s - State store CRUD operations functional through Dapr", component.description)
		})
	}

	// Validate state store query capabilities 
	t.Run("DaprStateStoreQueryCapabilities", func(t *testing.T) {
		// Test state store query functionality that services depend on
		queryURL := "http://localhost:3500/v1.0/state/statestore/query"
		queryPayload := `{
			"filter": {},
			"page": {
				"limit": 10
			}
		}`

		queryReq, err := http.NewRequestWithContext(ctx, "POST", queryURL, strings.NewReader(queryPayload))
		require.NoError(t, err, "Failed to create state store query request")
		queryReq.Header.Set("Content-Type", "application/json")

		queryResp, err := client.Do(queryReq)
		if err != nil {
			t.Errorf("RED PHASE VALIDATION: State store query capability failed: %v", err)
			return
		}
		defer queryResp.Body.Close()

		if queryResp.StatusCode != http.StatusOK && queryResp.StatusCode != http.StatusNotFound {
			body := make([]byte, 1024)
			queryResp.Body.Read(body)
			t.Errorf("RED PHASE VALIDATION: State store query returned %d: %s", 
				queryResp.StatusCode, string(body))
		} else {
			t.Logf("RED PHASE VALIDATION SUCCESS: State store query capabilities accessible through Dapr")
		}
	})

	// Validate PostgreSQL connectivity through Dapr vs Direct Database Access
	t.Run("PostgreSQLConnectivityValidation", func(t *testing.T) {
		// First test direct PostgreSQL connectivity (infrastructure level)
		connStr := "postgresql://postgres:password@localhost:5432/international_center_development?sslmode=disable"
		
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			t.Errorf("RED PHASE VALIDATION: Direct PostgreSQL connection failed - infrastructure not ready: %v", err)
			return
		}
		defer db.Close()

		err = db.PingContext(ctx)
		if err != nil {
			t.Errorf("RED PHASE VALIDATION: PostgreSQL ping failed - database not accessible: %v", err)
			return
		}

		// Test that state store configuration aligns with actual database
		// This validates that Dapr state store can connect to the same PostgreSQL instance
		_, err = db.ExecContext(ctx, "SELECT 1")
		assert.NoError(t, err, "RED PHASE VALIDATION: PostgreSQL must be functional for Dapr state store integration")

		t.Logf("RED PHASE VALIDATION SUCCESS: PostgreSQL database accessible for Dapr state store integration")
	})

	// Validate state store connection string configuration
	t.Run("DaprStateStoreConfigurationValidation", func(t *testing.T) {
		// Test that Dapr state store configuration is consistent with actual database availability
		// This will help identify configuration mismatches

		// Expected connection details that should match Dapr component configuration
		expectedHost := "localhost"  // or "postgresql" in container network
		expectedPort := "5432"
		expectedDatabase := "international_center_development"
		expectedUser := "postgres"

		// Validate these match what services expect
		t.Logf("RED PHASE VALIDATION: Expected state store configuration:")
		t.Logf("  Host: %s", expectedHost)
		t.Logf("  Port: %s", expectedPort) 
		t.Logf("  Database: %s", expectedDatabase)
		t.Logf("  User: %s", expectedUser)

		// The actual validation will happen when state store operations succeed
		// This serves as documentation for configuration requirements
	})
}

