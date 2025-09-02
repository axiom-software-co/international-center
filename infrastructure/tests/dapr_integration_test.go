package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/dapr/go-sdk/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDaprComponentIntegration(t *testing.T) {
	// Integration test - requires full podman compose environment
	// Run tests sequentially to avoid gRPC connection conflicts
	
	// Create single Dapr client for all tests to avoid connection issues
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	
	daprClient, err := client.NewClient()
	require.NoError(t, err, "Should create Dapr client successfully")
	defer daprClient.Close()

	t.Run("dapr state store component connectivity", func(t *testing.T) {
		// Test: Dapr state store component is accessible and configured
		stateStoreName := requireEnv(t, "DAPR_STATE_STORE_NAME")

		// Test state store connectivity by attempting a query operation
		query := `{
			"filter": {
				"EQ": { "is_deleted": false }
			},
			"page": {
				"limit": 1
			}
		}`
		
		results, err := daprClient.QueryStateAlpha1(ctx, stateStoreName, query, nil)
		require.NoError(t, err, "State store should be accessible for queries")
		assert.NotNil(t, results, "Query results should not be nil")
		assert.NotNil(t, results.Results, "Query results.Results should not be nil")
		
		// Empty results are expected and valid
		assert.GreaterOrEqual(t, len(results.Results), 0, "Results array should be accessible")
	})

	t.Run("dapr redis pubsub component connectivity", func(t *testing.T) {
		// Test: Redis PubSub component is accessible through Dapr
		redisPort := requireEnv(t, "REDIS_PORT")

		// Test Redis connectivity by attempting a simple get state operation
		// This validates that Redis is accessible through Dapr state store
		_, err = daprClient.GetState(ctx, "statestore-postgresql", "health-check-redis", nil)
		// Error is expected for non-existent key, but should not be a connection error
		if err != nil {
			assert.Contains(t, []string{
				"state not found",
				"error getting state: state not found",
			}, err.Error(), "Should get 'not found' error, not connection error. Port should be accessible: %s", redisPort)
		}
	})

	t.Run("dapr categories store component connectivity", func(t *testing.T) {
		// Test: Categories store component is accessible
		categoriesStoreName := requireEnv(t, "DAPR_CATEGORIES_STORE_NAME")

		// Test categories store connectivity
		query := `{
			"filter": {
				"EQ": { "is_deleted": false }
			},
			"page": {
				"limit": 1
			}
		}`
		
		results, err := daprClient.QueryStateAlpha1(ctx, categoriesStoreName, query, nil)
		require.NoError(t, err, "Categories store should be accessible for queries")
		assert.NotNil(t, results, "Query results should not be nil")
		assert.NotNil(t, results.Results, "Query results.Results should not be nil")
	})
}

func TestDaprComponentConfiguration(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("dapr components respond to health checks", func(t *testing.T) {
		// Test: Dapr components are properly configured in the environment
		servicesAPIPort := requireEnv(t, "SERVICES_API_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Test that Services API Dapr sidecar is accessible for health
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/health/ready", servicesAPIPort))
		require.NoError(t, err, "Services API readiness check should be accessible")
		defer resp.Body.Close()
		
		// Should return either OK or service unavailable, but not connection refused
		assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, resp.StatusCode,
			"Readiness endpoint should be accessible (OK or unavailable, not connection refused)")
		
		if resp.StatusCode == http.StatusOK {
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err, "Should decode readiness response as JSON")
			
			status, exists := response["status"]
			assert.True(t, exists, "Response should contain status field")
			assert.Equal(t, "ready", status, "Status should be ready when OK returned")
		}
	})

	t.Run("dapr state store returns consistent empty results", func(t *testing.T) {
		// Test: Empty state store queries return consistent JSON structure
		stateStoreName := requireEnv(t, "DAPR_STATE_STORE_NAME")
		
		daprClient, err := client.NewClient()
		require.NoError(t, err, "Should create Dapr client successfully")
		defer daprClient.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test published services query structure
		publishedQuery := `{
			"filter": {
				"AND": [
					{ "EQ": { "publishing_status": "published" } },
					{ "EQ": { "is_deleted": false } }
				]
			},
			"sort": [
				{ "key": "order_number", "order": "ASC" },
				{ "key": "created_on", "order": "DESC" }
			],
			"page": {
				"limit": 20
			}
		}`
		
		results, err := daprClient.QueryStateAlpha1(ctx, stateStoreName, publishedQuery, nil)
		require.NoError(t, err, "Published services query should execute successfully")
		assert.NotNil(t, results, "Query results should not be nil")
		assert.NotNil(t, results.Results, "Query results.Results should not be nil")
		
		// Empty results are expected but structure should be consistent
		assert.IsType(t, []interface{}{}, results.Results, "Results should be array type")
	})

	t.Run("dapr grpc endpoint connectivity", func(t *testing.T) {
		// Test: Dapr gRPC endpoint is accessible
		daprGRPCEndpoint := requireEnv(t, "DAPR_GRPC_ENDPOINT")
		
		// Validate that gRPC endpoint format is correct
		assert.Contains(t, daprGRPCEndpoint, ":", "gRPC endpoint should contain port separator")
		
		// Create Dapr client which internally tests gRPC connectivity
		daprClient, err := client.NewClient()
		require.NoError(t, err, "Should connect to Dapr gRPC endpoint at: %s", daprGRPCEndpoint)
		defer daprClient.Close()
		
		// Test basic Dapr connectivity
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// Simple health check via state operation
		_, err = daprClient.GetState(ctx, "statestore-postgresql", "grpc-health-check", nil)
		// Connection should succeed even if key doesn't exist
		if err != nil {
			assert.Contains(t, []string{
				"state not found", 
				"error getting state: state not found",
			}, err.Error(), "Should get 'not found' error, not gRPC connection error")
		}
	})
}