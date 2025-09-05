package testing

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	
	sharedtesting "github.com/axiom-software-co/international-center/src/deployer/shared/testing"
)

// TestDevelopmentInfrastructureDeployment validates complete infrastructure deployment
func TestDevelopmentInfrastructureDeployment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite := sharedtesting.NewIntegrationTestSuite(t)
	defer suite.Cleanup(t)

	t.Run("Infrastructure_Deployment_Readiness", func(t *testing.T) {
		// Arrange - Validate infrastructure is expected to be deployed
		require.NotNil(t, suite.Environment, "Environment configuration must be available")
		assert.Equal(t, "development", suite.Environment.Environment, "Must be running in development environment")
		
		// Act - Perform infrastructure health check
		suite.InfrastructureHealthCheck(t)
		
		// Assert - All critical infrastructure components should be healthy
		// This test will fail until infrastructure is properly deployed
		t.Log("Infrastructure deployment validation completed")
	})
}

// TestPostgreSQLContainerConnectivity validates database container readiness
func TestPostgreSQLContainerConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("PostgreSQL_Container_Health", func(t *testing.T) {
		// Arrange - Get expected database connection details
		connectionString := sharedtesting.GetEnvVar("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/development_test?sslmode=disable")
		
		// Act - Attempt database connection
		db, err := sql.Open("postgres", connectionString)
		require.NoError(t, err, "Database connection should be created successfully")
		defer db.Close()
		
		// Assert - Database should accept connections and be migration-ready
		err = db.PingContext(ctx)
		require.NoError(t, err, "PostgreSQL container should accept connections")
		
		// Validate database is ready for migrations
		var version string
		err = db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
		require.NoError(t, err, "Should be able to query database version")
		assert.Contains(t, version, "PostgreSQL", "Should be running PostgreSQL")
		
		t.Logf("PostgreSQL container is ready: %s", version)
	})

	t.Run("PostgreSQL_Migration_Readiness", func(t *testing.T) {
		// Arrange - Connect to database
		connectionString := sharedtesting.GetEnvVar("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/development_test?sslmode=disable")
		db, err := sql.Open("postgres", connectionString)
		require.NoError(t, err, "Database connection should be established")
		defer db.Close()
		
		// Act - Validate schema migration capabilities
		var canCreateSchema bool
		err = db.QueryRowContext(ctx, "SELECT has_database_privilege('international_center_dev', 'CREATE')").Scan(&canCreateSchema)
		require.NoError(t, err, "Should be able to check database privileges")
		
		// Assert - Database should support schema creation
		assert.True(t, canCreateSchema, "Database should allow schema creation for migrations")
	})
}

// TestRedisContainerConnectivity validates Redis container for state store and pub/sub
func TestRedisContainerConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("Redis_Container_Health", func(t *testing.T) {
		// Arrange - Get Redis connection details
		redisAddr := sharedtesting.GetEnvVar("REDIS_ADDR", "localhost:6379")
		redisPassword := sharedtesting.GetEnvVar("REDIS_PASSWORD", "")
		
		client := redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: redisPassword,
			DB:       0,
		})
		defer client.Close()
		
		// Act - Test Redis connectivity
		err := client.Ping(ctx).Err()
		require.NoError(t, err, "Redis container should accept connections")
		
		// Assert - Redis should support state operations
		err = client.Set(ctx, "test_key", "test_value", time.Minute).Err()
		require.NoError(t, err, "Redis should support SET operations")
		
		value, err := client.Get(ctx, "test_key").Result()
		require.NoError(t, err, "Redis should support GET operations")
		assert.Equal(t, "test_value", value, "Redis should return correct values")
		
		// Cleanup test key
		client.Del(ctx, "test_key")
		
		t.Log("Redis container is ready for state store and pub/sub operations")
	})

	t.Run("Redis_PubSub_Readiness", func(t *testing.T) {
		// Arrange - Create Redis client
		redisAddr := sharedtesting.GetEnvVar("REDIS_ADDR", "localhost:6379")
		client := redis.NewClient(&redis.Options{
			Addr: redisAddr,
		})
		defer client.Close()
		
		// Act - Test pub/sub capabilities
		pubsub := client.Subscribe(ctx, "test_channel")
		defer pubsub.Close()
		
		// Assert - Should be able to subscribe to channels
		err := client.Publish(ctx, "test_channel", "test_message").Err()
		require.NoError(t, err, "Redis should support pub/sub operations")
	})
}

// TestVaultContainerInitialization validates Vault container setup
func TestVaultContainerInitialization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("Vault_Container_Health", func(t *testing.T) {
		// Arrange - Get Vault connection details
		vaultAddr := sharedtesting.GetEnvVar("VAULT_ADDR", "http://localhost:8200")
		
		// Act - Check Vault health endpoint
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", vaultAddr+"/v1/sys/health", nil)
		require.NoError(t, err, "Vault container should be accessible")
		defer resp.Body.Close()
		
		// Assert - Vault should be initialized and unsealed
		assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 429, 
			"Vault should return healthy status (200 or 429)")
		
		t.Log("Vault container is accessible and initialized")
	})

	t.Run("Vault_Secret_Store_Readiness", func(t *testing.T) {
		// Arrange - This test validates Vault is ready for secret operations
		vaultAddr := sharedtesting.GetEnvVar("VAULT_ADDR", "http://localhost:8200")
		
		// Act - Test Vault secret engine availability
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", vaultAddr+"/v1/sys/mounts", nil)
		require.NoError(t, err, "Should be able to query Vault mounts")
		defer resp.Body.Close()
		
		// Assert - Vault should have secret engines available
		assert.True(t, resp.StatusCode < 500, "Vault secret engines should be accessible")
	})
}

// TestAzuriteStorageEmulatorConnectivity validates blob storage emulator
func TestAzuriteStorageEmulatorConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("Azurite_Blob_Service_Health", func(t *testing.T) {
		// Arrange - Get Azurite connection details
		azuriteEndpoint := sharedtesting.GetEnvVar("AZURITE_BLOB_ENDPOINT", "http://localhost:10000")
		
		// Act - Check Azurite blob service health
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", azuriteEndpoint+"/devstoreaccount1?comp=list", nil)
		require.NoError(t, err, "Azurite blob service should be accessible")
		defer resp.Body.Close()
		
		// Assert - Azurite should provide blob storage API
		assert.True(t, resp.StatusCode < 500, "Azurite should provide valid blob storage API")
		
		t.Log("Azurite blob storage emulator is ready")
	})

	t.Run("Azurite_Storage_Operations", func(t *testing.T) {
		// Arrange - This test validates basic storage operations are available
		azuriteEndpoint := sharedtesting.GetEnvVar("AZURITE_BLOB_ENDPOINT", "http://localhost:10000")
		
		// Act - Test container operations endpoint
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", azuriteEndpoint+"/devstoreaccount1/test-container?restype=container", nil)
		
		// Assert - Should be able to interact with storage API (404 is acceptable for non-existent container)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode == 404 || resp.StatusCode < 500, 
				"Azurite should provide container operations")
		}
	})
}

// TestDaprControlPlaneDeployment validates Dapr orchestrator container
func TestDaprControlPlaneDeployment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("Dapr_Control_Plane_Health", func(t *testing.T) {
		// Arrange - Get Dapr control plane connection details
		daprHTTPPort := sharedtesting.GetEnvVar("DAPR_HTTP_PORT", "3500")
		daprEndpoint := fmt.Sprintf("http://localhost:%s", daprHTTPPort)
		
		// Act - Check Dapr control plane health
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", daprEndpoint+"/v1.0/healthz", nil)
		require.NoError(t, err, "Dapr control plane should be accessible")
		defer resp.Body.Close()
		
		// Assert - Dapr should be running and healthy
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Dapr control plane should be healthy")
		
		t.Log("Dapr control plane is running and healthy")
	})

	t.Run("Dapr_Service_Discovery_Readiness", func(t *testing.T) {
		// Arrange - Get Dapr connection details
		daprHTTPPort := sharedtesting.GetEnvVar("DAPR_HTTP_PORT", "3500")
		daprEndpoint := fmt.Sprintf("http://localhost:%s", daprHTTPPort)
		
		// Act - Test service discovery capabilities
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", daprEndpoint+"/v1.0/metadata", nil)
		require.NoError(t, err, "Should be able to query Dapr metadata")
		defer resp.Body.Close()
		
		// Assert - Dapr should provide service discovery
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Dapr should provide metadata endpoint")
	})
}

// TestContainerNetworkConnectivity validates inter-container networking
func TestContainerNetworkConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("Container_Network_Isolation", func(t *testing.T) {
		// Arrange - All containers should be on same network for service discovery
		suite := sharedtesting.NewIntegrationTestSuite(t)
		defer suite.Cleanup(t)
		
		// Act - Perform comprehensive network connectivity check
		suite.InfrastructureHealthCheck(t)
		
		// Assert - Network connectivity should be established
		assert.NotNil(t, suite.Environment, "Environment should be configured")
		
		t.Log("Container network connectivity validation completed")
	})

	t.Run("Service_Discovery_Network", func(t *testing.T) {
		// Arrange - Services should discover each other via container names/network aliases
		daprPort := sharedtesting.GetEnvVar("DAPR_HTTP_PORT", "3500")
		
		// Act - Verify Dapr can access other services on the network
		daprEndpoint := fmt.Sprintf("http://localhost:%s/v1.0/metadata", daprPort)
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", daprEndpoint, nil)
		
		// Assert - Network should support service discovery
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode < 500, "Network should support service discovery")
		}
	})
}

// TestInfrastructureReadinessForApplications validates infrastructure is ready for backend services
func TestInfrastructureReadinessForApplications(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	suite := sharedtesting.NewIntegrationTestSuite(t)
	defer suite.Cleanup(t)

	t.Run("Complete_Infrastructure_Readiness", func(t *testing.T) {
		// Arrange - All infrastructure components must be ready
		require.NotNil(t, suite.Environment, "Environment must be configured")
		
		// Act - Perform comprehensive infrastructure readiness check
		suite.InfrastructureHealthCheck(t)
		
		// Assert - Infrastructure should be ready for application deployment
		t.Run("Database_Ready", func(t *testing.T) {
			connectionString := suite.Environment.DatabaseURL
			if connectionString == "" {
				connectionString = "postgresql://postgres:postgres@localhost:5432/development_test?sslmode=disable"
			}
			
			db, err := sql.Open("postgres", connectionString)
			require.NoError(t, err, "Database should be accessible")
			defer db.Close()
			
			err = db.PingContext(ctx)
			require.NoError(t, err, "Database should accept connections")
		})
		
		t.Run("State_Store_Ready", func(t *testing.T) {
			client := redis.NewClient(&redis.Options{
				Addr: suite.Environment.RedisAddr,
				Password: suite.Environment.RedisPassword,
			})
			defer client.Close()
			
			err := client.Ping(ctx).Err()
			require.NoError(t, err, "Redis state store should be ready")
		})
		
		t.Run("Secret_Store_Ready", func(t *testing.T) {
			vaultAddr := suite.Environment.VaultAddr
			resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", vaultAddr+"/v1/sys/health", nil)
			require.NoError(t, err, "Vault should be accessible")
			defer resp.Body.Close()
			
			assert.True(t, resp.StatusCode < 500, "Vault should be healthy")
		})
		
		t.Log("Infrastructure is ready for application services deployment")
	})
}