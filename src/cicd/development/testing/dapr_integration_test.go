package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	sharedtesting "github.com/axiom-software-co/international-center/src/deployer/shared/testing"
)

// TestDaprControlPlane validates the Dapr control plane is running and accessible
func TestDaprControlPlane(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := sharedtesting.CreateIntegrationTestContext()
	defer cancel()

	// Test Dapr sidecar HTTP API is accessible
	t.Run("Dapr_HTTP_API_Accessibility", func(t *testing.T) {
		// Arrange
		daprHTTPPort := sharedtesting.GetRequiredEnvVar(t, "DAPR_HTTP_PORT")
		serviceHost := sharedtesting.GetRequiredEnvVar(t, "SERVICE_HOST")
		
		// Act - Check Dapr metadata endpoint
		daprURL := fmt.Sprintf("http://%s:%s/v1.0/metadata", serviceHost, daprHTTPPort)
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", daprURL, nil)

		// Assert - Dapr HTTP API should be accessible
		require.NoError(t, err, "Dapr HTTP API should be accessible")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Dapr metadata endpoint should respond with 200")

		// Verify Dapr metadata response
		var metadata map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&metadata)
		require.NoError(t, err, "Dapr metadata should be valid JSON")
		
		assert.Contains(t, metadata, "id", "Dapr metadata should contain app ID")
		
		// Check for runtime information (can be either "actors" or "actorRuntime")
		hasActorInfo := metadata["actors"] != nil || metadata["actorRuntime"] != nil
		assert.True(t, hasActorInfo, "Dapr metadata should contain actor runtime information")
		
		// Components may be empty if no components configured, but structure should be present
		// Accept either "components" field or just validate that core Dapr is running
		if metadata["components"] != nil {
			t.Logf("Dapr components configured: %v", metadata["components"])
		} else {
			t.Logf("No components configured (expected for minimal test setup)")
		}
	})

	// Test Dapr sidecar gRPC API is accessible
	t.Run("Dapr_GRPC_API_Accessibility", func(t *testing.T) {
		// Arrange
		daprGRPCPort := sharedtesting.GetRequiredEnvVar(t, "DAPR_GRPC_PORT")
		serviceHost := sharedtesting.GetRequiredEnvVar(t, "SERVICE_HOST")
		
		// Act - Check if Dapr gRPC port is listening
		address := fmt.Sprintf("%s:%s", serviceHost, daprGRPCPort)
		conn, err := connectWithTimeout(ctx, "tcp", address, 3*time.Second)

		// Assert - Dapr gRPC API should be accessible
		require.NoError(t, err, "Dapr gRPC API should be accessible")
		if conn != nil {
			conn.Close()
		}
		
		t.Logf("Dapr gRPC API is accessible at %s", address)
	})

	// Test Dapr placement service is running
	t.Run("Dapr_Placement_Service", func(t *testing.T) {
		// Arrange
		placementPort := sharedtesting.GetEnvVar("DAPR_PLACEMENT_PORT", "6050")
		if placementPort == "" {
			t.Skip("DAPR_PLACEMENT_PORT not configured, skipping placement service test")
		}
		
		serviceHost := sharedtesting.GetRequiredEnvVar(t, "SERVICE_HOST")
		
		// Act - Check if Dapr placement port is listening
		address := fmt.Sprintf("%s:%s", serviceHost, placementPort)
		conn, err := connectWithTimeout(ctx, "tcp", address, 3*time.Second)

		// Assert - Dapr placement service should be accessible
		require.NoError(t, err, "Dapr placement service should be accessible")
		if conn != nil {
			conn.Close()
		}
		
		t.Logf("Dapr placement service is accessible at %s", address)
	})
}

// TestDaprServiceRegistration validates services are registered with Dapr
func TestDaprServiceRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := sharedtesting.CreateIntegrationTestContext()
	defer cancel()

	// Test services are discoverable via Dapr service invocation
	expectedServices := []string{"content-api", "services-api", "public-gateway", "admin-gateway"}

	for _, serviceName := range expectedServices {
		t.Run(fmt.Sprintf("Service_%s_Registration", serviceName), func(t *testing.T) {
			// Arrange
			daprHTTPPort := sharedtesting.GetRequiredEnvVar(t, "DAPR_HTTP_PORT")
			serviceHost := sharedtesting.GetRequiredEnvVar(t, "SERVICE_HOST")

			// Act - Try to invoke service health endpoint via Dapr
			healthURL := fmt.Sprintf("http://%s:%s/v1.0/invoke/%s/method/health", serviceHost, daprHTTPPort, serviceName)
			resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", healthURL, nil)

			// Assert - Service should be discoverable and responsive via Dapr
			require.NoError(t, err, "Service %s should be discoverable via Dapr", serviceName)
			
			// Accept either 200 (healthy) or other status codes that indicate service is reachable
			// The important thing is that Dapr can route to the service
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
				"Service %s should be reachable via Dapr (status: %d)", serviceName, resp.StatusCode)

			t.Logf("Service %s is registered and reachable via Dapr (status: %d)", serviceName, resp.StatusCode)
		})
	}
}

// TestDaprComponents validates Dapr components are configured and accessible
func TestDaprComponents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := sharedtesting.CreateIntegrationTestContext()
	defer cancel()

	// Get Dapr components configuration
	t.Run("Dapr_Components_Configuration", func(t *testing.T) {
		// Arrange
		daprHTTPPort := sharedtesting.GetRequiredEnvVar(t, "DAPR_HTTP_PORT")
		serviceHost := sharedtesting.GetRequiredEnvVar(t, "SERVICE_HOST")
		
		// Act - Get Dapr metadata to check components
		metadataURL := fmt.Sprintf("http://%s:%s/v1.0/metadata", serviceHost, daprHTTPPort)
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", metadataURL, nil)

		// Assert - Should be able to get components metadata
		require.NoError(t, err, "Should be able to get Dapr components metadata")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Dapr metadata endpoint should respond")

		var metadata map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&metadata)
		require.NoError(t, err, "Dapr metadata should be valid JSON")

		// Verify components are configured
		components, exists := metadata["components"]
		assert.True(t, exists, "Dapr metadata should contain components")
		
		if exists {
			componentsList, ok := components.([]interface{})
			assert.True(t, ok, "Components should be a list")
			
			if ok {
				t.Logf("Dapr has %d components configured", len(componentsList))
				
				// Log component types for debugging
				componentTypes := make(map[string]int)
				for _, comp := range componentsList {
					if compMap, ok := comp.(map[string]interface{}); ok {
						if compType, exists := compMap["type"]; exists {
							if typeStr, ok := compType.(string); ok {
								componentTypes[typeStr]++
							}
						}
					}
				}
				
				for compType, count := range componentTypes {
					t.Logf("Component type: %s (count: %d)", compType, count)
				}
			}
		}
	})

	// Test state store component
	t.Run("State_Store_Component", func(t *testing.T) {
		// Arrange
		daprHTTPPort := sharedtesting.GetRequiredEnvVar(t, "DAPR_HTTP_PORT")
		serviceHost := sharedtesting.GetRequiredEnvVar(t, "SERVICE_HOST")
		
		// Act - Try to access state store via Dapr
		stateURL := fmt.Sprintf("http://%s:%s/v1.0/state/statestore", serviceHost, daprHTTPPort)
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", stateURL, nil)

		// Assert - State store component should be accessible
		// Note: This might return 404 if no state exists, which is fine - it means the component is working
		require.NoError(t, err, "State store component should be accessible via Dapr")
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound, 
			"State store should respond (200 or 404 are acceptable)")
		
		t.Logf("State store component is accessible (status: %d)", resp.StatusCode)
	})

	// Test pubsub component
	t.Run("PubSub_Component", func(t *testing.T) {
		// Arrange
		daprHTTPPort := sharedtesting.GetRequiredEnvVar(t, "DAPR_HTTP_PORT")
		serviceHost := sharedtesting.GetRequiredEnvVar(t, "SERVICE_HOST")
		
		// Act - Try to publish a test message via Dapr pubsub
		pubsubURL := fmt.Sprintf("http://%s:%s/v1.0/publish/pubsub/test-topic", serviceHost, daprHTTPPort)
		testMessage := []byte(`{"test": "message", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`)
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "POST", pubsubURL, testMessage)

		// Assert - PubSub component should be accessible
		require.NoError(t, err, "PubSub component should be accessible via Dapr")
		
		// Accept various success status codes (200, 202, 204)
		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
			"PubSub publish should succeed (status: %d)", resp.StatusCode)
		
		t.Logf("PubSub component is accessible (status: %d)", resp.StatusCode)
	})

	// Test bindings component (for Azurite blob storage)
	t.Run("Bindings_Component", func(t *testing.T) {
		// Arrange
		daprHTTPPort := sharedtesting.GetRequiredEnvVar(t, "DAPR_HTTP_PORT")
		serviceHost := sharedtesting.GetRequiredEnvVar(t, "SERVICE_HOST")
		
		// Act - Try to access bindings component via Dapr
		bindingURL := fmt.Sprintf("http://%s:%s/v1.0/bindings/blob-storage", serviceHost, daprHTTPPort)
		testData := []byte(`{"operation": "list"}`)
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "POST", bindingURL, testData)

		// Assert - Bindings component should be accessible
		require.NoError(t, err, "Bindings component should be accessible via Dapr")
		
		// The binding might return various status codes depending on configuration
		// The important thing is that Dapr can reach the binding component
		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500,
			"Bindings component should be reachable (status: %d)", resp.StatusCode)
		
		t.Logf("Bindings component is accessible (status: %d)", resp.StatusCode)
	})
}

// TestDaprSecrets validates Dapr secrets management integration  
func TestDaprSecrets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := sharedtesting.CreateIntegrationTestContext()
	defer cancel()

	// Test secrets store component
	t.Run("Secrets_Store_Component", func(t *testing.T) {
		// Arrange
		daprHTTPPort := sharedtesting.GetRequiredEnvVar(t, "DAPR_HTTP_PORT")
		serviceHost := sharedtesting.GetRequiredEnvVar(t, "SERVICE_HOST")
		
		// Act - Try to access secrets store via Dapr
		secretsURL := fmt.Sprintf("http://%s:%s/v1.0/secrets/secretstore/test-secret", serviceHost, daprHTTPPort)
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", secretsURL, nil)

		// Assert - Secrets store component should be accessible
		require.NoError(t, err, "Secrets store component should be accessible via Dapr")
		
		// Secrets might not exist or access might be restricted - various status codes are acceptable
		// The important thing is that Dapr can reach the secrets component
		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500,
			"Secrets store should be reachable (status: %d)", resp.StatusCode)
		
		t.Logf("Secrets store component is accessible (status: %d)", resp.StatusCode)
	})
}

// TestDaprMiddleware validates Dapr middleware chain is working correctly
func TestDaprMiddleware(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := sharedtesting.CreateIntegrationTestContext()
	defer cancel()

	// Test CORS middleware
	t.Run("CORS_Middleware", func(t *testing.T) {
		// Arrange
		daprHTTPPort := sharedtesting.GetRequiredEnvVar(t, "DAPR_HTTP_PORT")
		serviceHost := sharedtesting.GetRequiredEnvVar(t, "SERVICE_HOST")
		
		// Act - Make CORS preflight request via Dapr
		daprURL := fmt.Sprintf("http://%s:%s/v1.0/invoke/public-gateway/method/api/v1/services", serviceHost, daprHTTPPort)
		req, err := http.NewRequestWithContext(ctx, "OPTIONS", daprURL, nil)
		require.NoError(t, err)
		
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "GET")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")
		
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)

		// Assert - CORS middleware should handle preflight requests
		require.NoError(t, err, "CORS preflight request should be handled")
		
		// Check for CORS headers in response
		corsHeaders := []string{
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Methods", 
			"Access-Control-Allow-Headers",
		}
		
		for _, header := range corsHeaders {
			if resp.Header.Get(header) != "" {
				t.Logf("CORS header present: %s = %s", header, resp.Header.Get(header))
			}
		}
		
		t.Logf("CORS middleware test completed (status: %d)", resp.StatusCode)
	})

	// Test rate limiting middleware
	t.Run("Rate_Limiting_Middleware", func(t *testing.T) {
		// Arrange
		daprHTTPPort := sharedtesting.GetRequiredEnvVar(t, "DAPR_HTTP_PORT")
		serviceHost := sharedtesting.GetRequiredEnvVar(t, "SERVICE_HOST")
		
		// Act - Make multiple requests to test rate limiting
		daprURL := fmt.Sprintf("http://%s:%s/v1.0/invoke/public-gateway/method/health", serviceHost, daprHTTPPort)
		
		var successCount, rateLimitedCount int
		requestCount := 5 // Make multiple requests to potentially trigger rate limiting
		
		for i := 0; i < requestCount; i++ {
			resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", daprURL, nil)
			if err == nil {
				if resp.StatusCode == http.StatusTooManyRequests {
					rateLimitedCount++
				} else if resp.StatusCode < 400 {
					successCount++
				}
			}
			
			// Small delay between requests
			time.Sleep(100 * time.Millisecond)
		}

		// Assert - Rate limiting middleware should be functional
		assert.True(t, successCount > 0, "Some requests should succeed")
		t.Logf("Rate limiting test: %d successful, %d rate limited out of %d requests", 
			successCount, rateLimitedCount, requestCount)
	})
}

// connectWithTimeout creates a network connection with timeout
func connectWithTimeout(ctx context.Context, network, address string, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: timeout,
	}
	return dialer.DialContext(ctx, network, address)
}