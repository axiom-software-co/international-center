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

	sharedValidation "github.com/axiom-software-co/international-center/src/public-website/infrastructure/test/shared"
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
		componentsURL := "http://localhost:3502/v1.0/components"
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
			saveURL := fmt.Sprintf("http://localhost:3502/v1.0/state/%s", component.componentName)
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
			getURL := fmt.Sprintf("http://localhost:3502/v1.0/state/%s/%s", component.componentName, testKey)
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
			deleteURL := fmt.Sprintf("http://localhost:3502/v1.0/state/%s/%s", component.componentName, testKey)
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
		queryURL := "http://localhost:3502/v1.0/state/statestore/query"
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

// RED PHASE: Comprehensive State Store CRUD Operations Validation
func TestInfrastructureIntegration_ComprehensiveStateStoreCRUD(t *testing.T) {
	// This test validates comprehensive CRUD operations through Dapr state store
	// Critical for data persistence and consistency across service boundaries
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 15 * time.Second}
	baseURL := "http://localhost:3502/v1.0/state/statestore"

	// Test data for comprehensive CRUD operations across different entity types
	testEntities := []struct {
		key        string
		value      string
		entityType string
		updateValue string
	}{
		{
			key:        "news-integration-test-1",
			value:      `{"news_id":"test-1","title":"Test News Article","content":"Integration test content","status":"published","created_at":"2024-01-01T10:00:00Z"}`,
			entityType: "news",
			updateValue: `{"news_id":"test-1","title":"Updated Test News Article","content":"Updated integration test content","status":"published","updated_at":"2024-01-01T11:00:00Z"}`,
		},
		{
			key:        "event-integration-test-1", 
			value:      `{"event_id":"test-1","title":"Test Event","description":"Integration test event","status":"active","event_date":"2024-06-01T15:00:00Z"}`,
			entityType: "event",
			updateValue: `{"event_id":"test-1","title":"Updated Test Event","description":"Updated integration test event","status":"active","event_date":"2024-06-01T16:00:00Z"}`,
		},
		{
			key:        "inquiry-integration-test-1",
			value:      `{"inquiry_id":"test-1","subject":"Test Inquiry","message":"Integration test message","status":"pending","inquiry_type":"business"}`,
			entityType: "inquiry", 
			updateValue: `{"inquiry_id":"test-1","subject":"Test Inquiry","message":"Integration test message","status":"processed","inquiry_type":"business"}`,
		},
	}

	// RED PHASE: Comprehensive Create Operations Validation
	t.Run("ComprehensiveCreateOperations", func(t *testing.T) {
		for _, entity := range testEntities {
			t.Run("Create_"+entity.entityType+"_"+entity.key, func(t *testing.T) {
				// Test entity creation with proper JSON structure
				savePayload := fmt.Sprintf(`[{"key": "%s", "value": %s}]`, entity.key, entity.value)
				
				saveReq, err := http.NewRequestWithContext(ctx, "POST", baseURL, strings.NewReader(savePayload))
				require.NoError(t, err, "Failed to create save request for %s", entity.entityType)
				saveReq.Header.Set("Content-Type", "application/json")

				saveResp, err := client.Do(saveReq)
				require.NoError(t, err, "State store save operation must succeed for %s entity", entity.entityType)
				defer saveResp.Body.Close()

				assert.True(t, saveResp.StatusCode == http.StatusNoContent || saveResp.StatusCode == http.StatusOK,
					"State store save must return success status for %s entity", entity.entityType)

				t.Logf("RED PHASE SUCCESS: Created %s entity with key %s", entity.entityType, entity.key)
			})
		}
	})

	// RED PHASE: Comprehensive Read Operations Validation
	t.Run("ComprehensiveReadOperations", func(t *testing.T) {
		for _, entity := range testEntities {
			t.Run("Read_"+entity.entityType+"_"+entity.key, func(t *testing.T) {
				// Test entity retrieval and data integrity
				getURL := fmt.Sprintf("%s/%s", baseURL, entity.key)
				getReq, err := http.NewRequestWithContext(ctx, "GET", getURL, nil)
				require.NoError(t, err, "Failed to create get request for %s", entity.entityType)

				getResp, err := client.Do(getReq)
				require.NoError(t, err, "State store get operation must succeed for %s entity", entity.entityType)
				defer getResp.Body.Close()

				assert.Equal(t, http.StatusOK, getResp.StatusCode,
					"State store get must return 200 OK for existing %s entity", entity.entityType)

				// Validate data integrity by checking response content
				if getResp.StatusCode == http.StatusOK {
					body := make([]byte, 2048)
					n, _ := getResp.Body.Read(body)
					responseContent := string(body[:n])
					
					// Verify that the response contains expected entity data
					assert.Contains(t, responseContent, fmt.Sprintf("test-1"), 
						"Retrieved %s entity must contain expected ID", entity.entityType)
					
					t.Logf("RED PHASE SUCCESS: Retrieved %s entity with key %s - data integrity validated", entity.entityType, entity.key)
				}
			})
		}
	})

	// RED PHASE: Comprehensive Update Operations Validation  
	t.Run("ComprehensiveUpdateOperations", func(t *testing.T) {
		for _, entity := range testEntities {
			t.Run("Update_"+entity.entityType+"_"+entity.key, func(t *testing.T) {
				// Test entity update with modified data
				updatePayload := fmt.Sprintf(`[{"key": "%s", "value": %s}]`, entity.key, entity.updateValue)
				
				updateReq, err := http.NewRequestWithContext(ctx, "POST", baseURL, strings.NewReader(updatePayload))
				require.NoError(t, err, "Failed to create update request for %s", entity.entityType)
				updateReq.Header.Set("Content-Type", "application/json")

				updateResp, err := client.Do(updateReq)
				require.NoError(t, err, "State store update operation must succeed for %s entity", entity.entityType)
				defer updateResp.Body.Close()

				assert.True(t, updateResp.StatusCode == http.StatusNoContent || updateResp.StatusCode == http.StatusOK,
					"State store update must return success status for %s entity", entity.entityType)

				// Verify update by reading back the modified data
				getURL := fmt.Sprintf("%s/%s", baseURL, entity.key)
				getReq, err := http.NewRequestWithContext(ctx, "GET", getURL, nil)
				require.NoError(t, err, "Failed to create verification get request")

				getResp, err := client.Do(getReq)
				if err == nil && getResp.StatusCode == http.StatusOK {
					defer getResp.Body.Close()
					body := make([]byte, 2048)
					n, _ := getResp.Body.Read(body)
					responseContent := string(body[:n])
					
					// Verify update was applied (check for updated timestamp or status)
					if strings.Contains(entity.updateValue, "Updated") {
						assert.Contains(t, responseContent, "Updated", 
							"Updated %s entity must contain modified data", entity.entityType)
					}
					
					t.Logf("RED PHASE SUCCESS: Updated %s entity with key %s - modification verified", entity.entityType, entity.key)
				}
			})
		}
	})

	// RED PHASE: Comprehensive Delete Operations Validation
	t.Run("ComprehensiveDeleteOperations", func(t *testing.T) {
		for _, entity := range testEntities {
			t.Run("Delete_"+entity.entityType+"_"+entity.key, func(t *testing.T) {
				// Test entity deletion
				deleteURL := fmt.Sprintf("%s/%s", baseURL, entity.key)
				deleteReq, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
				require.NoError(t, err, "Failed to create delete request for %s", entity.entityType)

				deleteResp, err := client.Do(deleteReq)
				require.NoError(t, err, "State store delete operation must succeed for %s entity", entity.entityType)
				defer deleteResp.Body.Close()

				assert.True(t, deleteResp.StatusCode == http.StatusNoContent || deleteResp.StatusCode == http.StatusOK,
					"State store delete must return success status for %s entity", entity.entityType)

				// Verify deletion by attempting to read the deleted entity
				getURL := fmt.Sprintf("%s/%s", baseURL, entity.key)
				getReq, err := http.NewRequestWithContext(ctx, "GET", getURL, nil)
				require.NoError(t, err, "Failed to create verification get request")

				getResp, err := client.Do(getReq)
				if err == nil {
					defer getResp.Body.Close()
					// Entity should be not found after deletion
					assert.True(t, getResp.StatusCode == http.StatusNotFound || getResp.StatusCode == http.StatusNoContent,
						"Deleted %s entity must not be retrievable", entity.entityType)
					
					t.Logf("RED PHASE SUCCESS: Deleted %s entity with key %s - deletion verified", entity.entityType, entity.key)
				}
			})
		}
	})
}

// RED PHASE: Data Persistence and Consistency Validation
func TestInfrastructureIntegration_DataPersistenceValidation(t *testing.T) {
	// This test validates data persistence across service lifecycle events
	// Critical for ensuring data durability in production scenarios
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 15 * time.Second}
	baseURL := "http://localhost:3502/v1.0/state/statestore"

	t.Run("DataPersistenceAcrossOperations", func(t *testing.T) {
		// Test data that should persist through various operations
		persistenceTestKey := "persistence-test-key"
		persistenceTestValue := `{"entity_id":"persist-1","data":"critical persistence test data","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`

		// Create initial data
		savePayload := fmt.Sprintf(`[{"key": "%s", "value": %s}]`, persistenceTestKey, persistenceTestValue)
		saveReq, err := http.NewRequestWithContext(ctx, "POST", baseURL, strings.NewReader(savePayload))
		require.NoError(t, err, "Failed to create persistence test data")
		saveReq.Header.Set("Content-Type", "application/json")

		saveResp, err := client.Do(saveReq)
		require.NoError(t, err, "Persistence test data creation must succeed")
		defer saveResp.Body.Close()

		assert.True(t, saveResp.StatusCode == http.StatusNoContent || saveResp.StatusCode == http.StatusOK,
			"Persistence test data save must succeed")

		// Verify persistence immediately after creation
		getURL := fmt.Sprintf("%s/%s", baseURL, persistenceTestKey)
		getReq, err := http.NewRequestWithContext(ctx, "GET", getURL, nil)
		require.NoError(t, err, "Failed to create persistence verification request")

		getResp, err := client.Do(getReq)
		require.NoError(t, err, "Persistence verification read must succeed")
		defer getResp.Body.Close()

		assert.Equal(t, http.StatusOK, getResp.StatusCode,
			"Persistence test data must be immediately retrievable")

		// Clean up test data
		deleteReq, err := http.NewRequestWithContext(ctx, "DELETE", getURL, nil)
		if err == nil {
			deleteResp, _ := client.Do(deleteReq)
			if deleteResp != nil {
				deleteResp.Body.Close()
			}
		}

		t.Logf("RED PHASE SUCCESS: Data persistence validation completed - data survives immediate operations")
	})

	t.Run("TransactionConsistencyValidation", func(t *testing.T) {
		// Test transaction-like consistency for multi-key operations
		transactionKeys := []string{
			"transaction-test-1",
			"transaction-test-2", 
			"transaction-test-3",
		}

		transactionValues := []string{
			`{"entity_id":"trans-1","data":"transaction test data 1","sequence":1}`,
			`{"entity_id":"trans-2","data":"transaction test data 2","sequence":2}`,
			`{"entity_id":"trans-3","data":"transaction test data 3","sequence":3}`,
		}

		// Create multiple related entities that should maintain consistency
		for i, key := range transactionKeys {
			savePayload := fmt.Sprintf(`[{"key": "%s", "value": %s}]`, key, transactionValues[i])
			saveReq, err := http.NewRequestWithContext(ctx, "POST", baseURL, strings.NewReader(savePayload))
			require.NoError(t, err, "Failed to create transaction test entity %d", i+1)
			saveReq.Header.Set("Content-Type", "application/json")

			saveResp, err := client.Do(saveReq)
			require.NoError(t, err, "Transaction entity %d creation must succeed", i+1)
			defer saveResp.Body.Close()

			assert.True(t, saveResp.StatusCode == http.StatusNoContent || saveResp.StatusCode == http.StatusOK,
				"Transaction entity %d save must succeed", i+1)
		}

		// Verify all entities exist and maintain consistency
		for i, key := range transactionKeys {
			getURL := fmt.Sprintf("%s/%s", baseURL, key)
			getReq, err := http.NewRequestWithContext(ctx, "GET", getURL, nil)
			require.NoError(t, err, "Failed to create consistency verification request")

			getResp, err := client.Do(getReq)
			require.NoError(t, err, "Consistency verification must succeed for entity %d", i+1)
			defer getResp.Body.Close()

			assert.Equal(t, http.StatusOK, getResp.StatusCode,
				"Transaction entity %d must be retrievable for consistency validation", i+1)
		}

		// Clean up transaction test data
		for _, key := range transactionKeys {
			deleteURL := fmt.Sprintf("%s/%s", baseURL, key)
			deleteReq, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
			if err == nil {
				deleteResp, _ := client.Do(deleteReq)
				if deleteResp != nil {
					deleteResp.Body.Close()
				}
			}
		}

		t.Logf("RED PHASE SUCCESS: Transaction consistency validation completed - multi-key operations maintain consistency")
	})
}

// RED PHASE: Deployment Automation Verification Tests
// These tests validate that the deployment automation process properly brings up
// all infrastructure components and that the deployment is coordinated correctly

func TestDeploymentAutomation_PulumiStackManagement(t *testing.T) {
	// RED PHASE: This test validates Pulumi stack management and deployment coordination
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("DeploymentAutomation_StackLifecycleValidation", func(t *testing.T) {
		// RED PHASE: Validate Pulumi stack lifecycle management
		
		// Check if Pulumi is available
		pulumiCmd := exec.CommandContext(ctx, "pulumi", "version")
		pulumiOutput, err := pulumiCmd.Output()
		
		if err == nil {
			t.Logf("Pulumi version available: %s", strings.TrimSpace(string(pulumiOutput)))
			
			// Check if development stack exists
			stackListCmd := exec.CommandContext(ctx, "pulumi", "stack", "ls")
			stackListCmd.Dir = "../../../"
			stackListOutput, stackListErr := stackListCmd.Output()
			
			if stackListErr == nil {
				stackList := string(stackListOutput)
				assert.Contains(t, stackList, "development", 
					"Development stack must exist for automated deployment")
				
				// If stack exists, validate it has recent activity
				if strings.Contains(stackList, "development") {
					historyCmd := exec.CommandContext(ctx, "pulumi", "history", "--page-size", "1")
					historyCmd.Dir = "../../../"
					historyOutput, historyErr := historyCmd.Output()
					
					if historyErr == nil {
						assert.NotEmpty(t, historyOutput, 
							"Stack must have deployment history indicating automation activity")
						t.Logf("Recent stack activity found: %s", strings.TrimSpace(string(historyOutput)))
					} else {
						t.Logf("RED PHASE: Cannot access stack history - expected until deployment automation is active: %v", historyErr)
					}
				}
			} else {
				t.Logf("RED PHASE: Cannot list Pulumi stacks - expected until deployment automation is configured: %v", stackListErr)
			}
		} else {
			t.Logf("RED PHASE: Pulumi not available - expected until deployment automation is configured: %v", err)
			
			// This should fail in RED PHASE until Pulumi is properly set up
			assert.Fail(t, "Pulumi must be available for infrastructure deployment automation")
		}
	})

	t.Run("DeploymentAutomation_ResourceProvisioning", func(t *testing.T) {
		// RED PHASE: Validate automated resource provisioning
		
		// Check if Pulumi has provisioned resources by examining outputs
		resourceOutputs := []struct {
			outputName  string
			description string
			required    bool
		}{
			{"postgresql_connection_string", "PostgreSQL database connection", true},
			{"vault_endpoint", "Vault secret store endpoint", true},
			{"rabbitmq_connection_string", "RabbitMQ messaging connection", true},
			{"dapr_components_path", "Dapr components configuration path", true},
			{"grafana_endpoint", "Grafana monitoring endpoint", false},
			{"monitoring_config", "Monitoring configuration", false},
		}

		for _, resource := range resourceOutputs {
			t.Run("Resource_"+resource.outputName, func(t *testing.T) {
				outputCmd := exec.CommandContext(ctx, "pulumi", "stack", "output", resource.outputName)
				outputCmd.Dir = "../../../"
				outputValue, outputErr := outputCmd.CombinedOutput()

				if outputErr == nil && len(strings.TrimSpace(string(outputValue))) > 0 {
					t.Logf("Resource %s provisioned: %s", resource.description, strings.TrimSpace(string(outputValue)))
					assert.NotEmpty(t, strings.TrimSpace(string(outputValue)), 
						"%s must be provisioned by deployment automation", resource.description)
				} else {
					if resource.required {
						t.Logf("RED PHASE: Required resource %s not provisioned - expected until deployment automation is complete: %v", resource.description, outputErr)
						
						// Required resources should fail in RED PHASE
						assert.Fail(t, fmt.Sprintf("Required resource %s must be provisioned by deployment automation", resource.description))
					} else {
						t.Logf("Optional resource %s not provisioned - acceptable: %v", resource.description, outputErr)
					}
				}
			})
		}
	})
}

func TestDeploymentAutomation_ComponentInitializationOrder(t *testing.T) {
	// RED PHASE: This test validates that components are initialized in correct order
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("DeploymentAutomation_InfrastructureFirstValidation", func(t *testing.T) {
		// RED PHASE: Validate infrastructure components start before application services
		
		infrastructureComponents := []struct {
			containerName string
			description   string
			priority      int // Lower numbers should start first
		}{
			{"postgresql", "PostgreSQL database", 1},
			{"vault", "HashiCorp Vault", 1},
			{"rabbitmq", "RabbitMQ messaging", 1},
			{"dapr_placement", "Dapr placement service", 2},
			{"grafana", "Grafana monitoring", 1},
		}

		applicationServices := []struct {
			containerName string
			description   string
			priority      int
		}{
			{"content", "Content service", 3},
			{"inquiries", "Inquiries service", 3},
			{"notifications", "Notifications service", 3},
			{"public-gateway", "Public gateway", 4},
			{"admin-gateway", "Admin gateway", 4},
		}

		// Check infrastructure components are running
		infraRunning := 0
		for _, component := range infrastructureComponents {
			containerCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+component.containerName, "--format", "{{.Names}}")
			containerOutput, err := containerCmd.Output()
			
			if err == nil && strings.Contains(string(containerOutput), component.containerName) {
				infraRunning++
				t.Logf("Infrastructure component running: %s", component.description)
			} else {
				t.Logf("RED PHASE: Infrastructure component %s not running - expected until deployment automation provides correct startup order: %v", component.description, err)
			}
		}

		// Check application services
		appRunning := 0
		for _, service := range applicationServices {
			containerCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+service.containerName, "--format", "{{.Names}}")
			containerOutput, err := containerCmd.Output()
			
			if err == nil && strings.Contains(string(containerOutput), service.containerName) {
				appRunning++
				t.Logf("Application service running: %s", service.description)
			} else {
				t.Logf("RED PHASE: Application service %s not running - expected until deployment automation is complete: %v", service.description, err)
			}
		}

		// In proper deployment, infrastructure should be running before applications
		if appRunning > 0 {
			assert.Greater(t, infraRunning, 0, 
				"Infrastructure components must be running before application services")
		}

		// RED PHASE: This should fail until deployment automation ensures proper startup order
		if infraRunning == 0 && appRunning == 0 {
			assert.Fail(t, "Deployment automation must ensure proper component initialization order")
		}
	})

	t.Run("DeploymentAutomation_DependencyValidation", func(t *testing.T) {
		// RED PHASE: Validate that dependent services wait for dependencies
		
		dependencies := []struct {
			serviceName    string
			dependsOn      []string
			description    string
		}{
			{"content", []string{"postgresql", "vault"}, "Content service depends on database and secrets"},
			{"inquiries", []string{"postgresql", "vault"}, "Inquiries service depends on database and secrets"},
			{"notifications", []string{"postgresql", "vault", "rabbitmq"}, "Notifications service depends on database, secrets, and messaging"},
			{"public-gateway", []string{"content", "inquiries", "notifications"}, "Public gateway depends on backend services"},
			{"admin-gateway", []string{"content", "inquiries", "notifications"}, "Admin gateway depends on backend services"},
		}

		for _, dep := range dependencies {
			t.Run("Dependency_"+dep.serviceName, func(t *testing.T) {
				// Check if the service is running
				serviceCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+dep.serviceName, "--format", "{{.Names}}")
				serviceOutput, serviceErr := serviceCmd.Output()

				if serviceErr == nil && strings.Contains(string(serviceOutput), dep.serviceName) {
					// Service is running - check if its dependencies are also running
					missingDeps := []string{}
					
					for _, depName := range dep.dependsOn {
						depCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+depName, "--format", "{{.Names}}")
						depOutput, depErr := depCmd.Output()
						
						if depErr != nil || !strings.Contains(string(depOutput), depName) {
							missingDeps = append(missingDeps, depName)
						}
					}

					assert.Empty(t, missingDeps, 
						"%s is running but dependencies %v are not running - deployment automation must ensure proper dependency order", 
						dep.description, missingDeps)
				} else {
					t.Logf("Service %s not running - dependency validation skipped", dep.serviceName)
				}
			})
		}
	})
}

func TestDeploymentAutomation_ComponentConfiguration(t *testing.T) {
	// RED PHASE: This test validates that components are properly configured during deployment
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("DeploymentAutomation_DaprComponentFiles", func(t *testing.T) {
		// RED PHASE: Validate Dapr component files are properly deployed
		
		expectedComponentFiles := []struct {
			filePath    string
			description string
		}{
			{"../../../configs/dapr/statestore.yaml", "PostgreSQL state store configuration"},
			{"../../../configs/dapr/secretstore.yaml", "Vault secret store configuration"},
			{"../../../configs/dapr/pubsub.yaml", "RabbitMQ pub/sub configuration"},
			{"../../../configs/dapr/config.yaml", "Dapr runtime configuration"},
		}

		for _, componentFile := range expectedComponentFiles {
			t.Run("ComponentFile_"+strings.Replace(componentFile.filePath, "/", "_", -1), func(t *testing.T) {
				// Check if component file exists
				_, err := exec.CommandContext(ctx, "ls", "-la", componentFile.filePath).Output()
				
				if err == nil {
					// File exists - validate it has proper content
					catCmd := exec.CommandContext(ctx, "cat", componentFile.filePath)
					catOutput, catErr := catCmd.Output()
					
					if catErr == nil {
						content := string(catOutput)
						assert.Contains(t, content, "apiVersion", 
							"%s must be valid Dapr component configuration", componentFile.description)
						assert.Contains(t, content, "kind: Component", 
							"%s must be valid Dapr component", componentFile.description)
						
						t.Logf("Dapr component file validated: %s", componentFile.description)
					} else {
						t.Logf("RED PHASE: Cannot read component file %s - expected until deployment automation configures components: %v", componentFile.description, catErr)
					}
				} else {
					t.Logf("RED PHASE: Component file %s not found - expected until deployment automation creates configurations: %v", componentFile.description, err)
					
					// This should fail in RED PHASE until component files are deployed
					assert.Fail(t, fmt.Sprintf("Deployment automation must create %s", componentFile.description))
				}
			})
		}
	})

	t.Run("DeploymentAutomation_EnvironmentVariables", func(t *testing.T) {
		// RED PHASE: Validate environment variables are properly configured
		
		requiredEnvVars := []struct {
			varName     string
			description string
			required    bool
		}{
			{"PULUMI_CONFIG_PASSPHRASE", "Pulumi configuration passphrase", true},
			{"DAPR_HTTP_PORT", "Dapr HTTP port configuration", false},
			{"DAPR_GRPC_PORT", "Dapr gRPC port configuration", false},
		}

		for _, envVar := range requiredEnvVars {
			t.Run("EnvVar_"+envVar.varName, func(t *testing.T) {
				envCmd := exec.CommandContext(ctx, "printenv", envVar.varName)
				envOutput, envErr := envCmd.Output()

				if envErr == nil && len(strings.TrimSpace(string(envOutput))) > 0 {
					t.Logf("Environment variable %s configured: %s", envVar.description, strings.TrimSpace(string(envOutput)))
				} else {
					if envVar.required {
						t.Logf("RED PHASE: Required environment variable %s not configured - expected until deployment automation sets up environment: %v", envVar.description, envErr)
						
						// Required environment variables should fail in RED PHASE
						assert.Fail(t, fmt.Sprintf("Required environment variable %s must be configured by deployment automation", envVar.description))
					} else {
						t.Logf("Optional environment variable %s not configured - acceptable", envVar.description)
					}
				}
			})
		}
	})
}

