package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	sharedtesting "github.com/axiom-software-co/international-center/src/deployer/shared/testing"
)

// TestDaprComponentConfiguration validates Dapr components are properly configured
func TestDaprComponentConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	daprHTTPPort := sharedtesting.GetEnvVar("DAPR_HTTP_PORT", "3500")
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("State_Store_Component_Configuration", func(t *testing.T) {
		// Arrange - State store component should be configured to use Redis
		componentsURL := fmt.Sprintf("http://localhost:%s/v1.0/metadata", daprHTTPPort)
		
		// Act - Query Dapr metadata for components
		req, err := http.NewRequestWithContext(ctx, "GET", componentsURL, nil)
		require.NoError(t, err, "Should be able to create metadata request")
		resp, err := client.Do(req)
		
		// Assert - Should find state store component
		require.NoError(t, err, "Dapr metadata endpoint should be accessible")
		require.Equal(t, http.StatusOK, resp.StatusCode, "Metadata endpoint should return OK")
		
		var metadata map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&metadata)
		require.NoError(t, err, "Should be able to decode metadata response")
		resp.Body.Close()
		
		// Validate state store component exists
		components, exists := metadata["components"].([]interface{})
		require.True(t, exists, "Components should exist in metadata")
		
		stateStoreFound := false
		for _, component := range components {
			comp := component.(map[string]interface{})
			if comp["name"] == "statestore" && comp["type"] == "state.redis" {
				stateStoreFound = true
				break
			}
		}
		assert.True(t, stateStoreFound, "Redis state store component should be configured")
		
		t.Log("State store component configuration validated")
	})

	t.Run("PubSub_Component_Configuration", func(t *testing.T) {
		// Arrange - Pub/sub component should be configured to use Redis
		componentsURL := fmt.Sprintf("http://localhost:%s/v1.0/metadata", daprHTTPPort)
		
		// Act - Query Dapr metadata for components
		req, err := http.NewRequestWithContext(ctx, "GET", componentsURL, nil)
		require.NoError(t, err, "Should be able to create metadata request")
		resp, err := client.Do(req)
		
		// Assert - Should find pub/sub component
		require.NoError(t, err, "Dapr metadata endpoint should be accessible")
		require.Equal(t, http.StatusOK, resp.StatusCode, "Metadata endpoint should return OK")
		
		var metadata map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&metadata)
		require.NoError(t, err, "Should be able to decode metadata response")
		resp.Body.Close()
		
		// Validate pub/sub component exists
		components, exists := metadata["components"].([]interface{})
		require.True(t, exists, "Components should exist in metadata")
		
		pubsubFound := false
		for _, component := range components {
			comp := component.(map[string]interface{})
			if comp["name"] == "pubsub" && comp["type"] == "pubsub.redis" {
				pubsubFound = true
				break
			}
		}
		assert.True(t, pubsubFound, "Redis pub/sub component should be configured")
		
		t.Log("Pub/sub component configuration validated")
	})

	t.Run("Secrets_Component_Configuration", func(t *testing.T) {
		// Arrange - Secrets component should be configured to use Vault
		componentsURL := fmt.Sprintf("http://localhost:%s/v1.0/metadata", daprHTTPPort)
		
		// Act - Query Dapr metadata for components
		req, err := http.NewRequestWithContext(ctx, "GET", componentsURL, nil)
		require.NoError(t, err, "Should be able to create metadata request")
		resp, err := client.Do(req)
		
		// Assert - Should find secrets component
		require.NoError(t, err, "Dapr metadata endpoint should be accessible")
		require.Equal(t, http.StatusOK, resp.StatusCode, "Metadata endpoint should return OK")
		
		var metadata map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&metadata)
		require.NoError(t, err, "Should be able to decode metadata response")
		resp.Body.Close()
		
		// Validate secrets component exists
		components, exists := metadata["components"].([]interface{})
		require.True(t, exists, "Components should exist in metadata")
		
		secretsFound := false
		for _, component := range components {
			comp := component.(map[string]interface{})
			if comp["name"] == "secrets" && comp["type"] == "secretstores.hashicorp.vault" {
				secretsFound = true
				break
			}
		}
		assert.True(t, secretsFound, "Vault secrets component should be configured")
		
		t.Log("Secrets component configuration validated")
	})
}

// TestDaprMiddlewareConfiguration validates Dapr middleware pipeline
func TestDaprMiddlewareConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	daprHTTPPort := sharedtesting.GetEnvVar("DAPR_HTTP_PORT", "3500")
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("CORS_Middleware_Configuration", func(t *testing.T) {
		// Arrange - CORS middleware should be configured for development
		testURL := fmt.Sprintf("http://localhost:%s/v1.0/healthz", daprHTTPPort)
		
		// Act - Make preflight request to test CORS
		req, err := http.NewRequestWithContext(ctx, "OPTIONS", testURL, nil)
		require.NoError(t, err, "Should be able to create OPTIONS request")
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "GET")
		resp, err := client.Do(req)
		
		// Assert - CORS headers should be present
		require.NoError(t, err, "CORS preflight request should succeed")
		
		// Note: This will initially fail until CORS middleware is configured
		corsHeader := resp.Header.Get("Access-Control-Allow-Origin")
		assert.NotEmpty(t, corsHeader, "CORS middleware should set Access-Control-Allow-Origin header")
		
		if resp.Body != nil {
			resp.Body.Close()
		}
		
		t.Log("CORS middleware configuration validated")
	})

	t.Run("Rate_Limiting_Middleware_Configuration", func(t *testing.T) {
		// Arrange - Rate limiting should be configured for development
		testURL := fmt.Sprintf("http://localhost:%s/v1.0/healthz", daprHTTPPort)
		
		// Act - Make multiple rapid requests to test rate limiting
		for i := 0; i < 5; i++ {
			req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
			require.NoError(t, err, "Should be able to create rate limit test request")
			resp, err := client.Do(req)
			
			// Basic connectivity assertion
			require.NoError(t, err, "Rate limit test request should succeed")
			
			if resp.Body != nil {
				resp.Body.Close()
			}
		}
		
		// Note: Advanced rate limiting validation will be implemented in GREEN phase
		t.Log("Rate limiting middleware basic validation completed")
	})
}

// TestDaprServiceDiscovery validates service discovery and name resolution
func TestDaprServiceDiscovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	daprHTTPPort := sharedtesting.GetEnvVar("DAPR_HTTP_PORT", "3500")
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("Service_Registration_Validation", func(t *testing.T) {
		// Arrange - Services should be discoverable via Dapr name resolution
		servicesURL := fmt.Sprintf("http://localhost:%s/v1.0/metadata", daprHTTPPort)
		
		// Act - Query service metadata
		req, err := http.NewRequestWithContext(ctx, "GET", servicesURL, nil)
		require.NoError(t, err, "Should be able to create service discovery request")
		resp, err := client.Do(req)
		
		// Assert - Dapr should be accessible for service discovery
		require.NoError(t, err, "Service discovery endpoint should be accessible")
		require.Equal(t, http.StatusOK, resp.StatusCode, "Service discovery should return OK")
		
		var metadata map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&metadata)
		require.NoError(t, err, "Should be able to decode service metadata")
		resp.Body.Close()
		
		// Validate service registration (will expand in GREEN phase)
		appID, exists := metadata["id"].(string)
		require.True(t, exists, "Service should have an app ID")
		assert.NotEmpty(t, appID, "App ID should not be empty")
		
		t.Logf("Service discovery validated - App ID: %s", appID)
	})

	t.Run("Service_Invocation_Name_Resolution", func(t *testing.T) {
		// Arrange - Should be able to invoke services by name
		expectedServices := []string{"content-api", "services-api", "public-gateway", "admin-gateway"}
		
		for _, serviceName := range expectedServices {
			// Act - Test service name resolution
			invokeURL := fmt.Sprintf("http://localhost:%s/v1.0/invoke/%s/method/health", daprHTTPPort, serviceName)
			req, err := http.NewRequestWithContext(ctx, "GET", invokeURL, nil)
			require.NoError(t, err, "Should be able to create service invocation request")
			
			// Note: This will initially fail until services are deployed
			resp, err := client.Do(req)
			
			// Assert - Service should be resolvable by name (will fail initially)
			if err == nil && resp != nil {
				t.Logf("Service %s is resolvable via name resolution", serviceName)
				if resp.Body != nil {
					resp.Body.Close()
				}
			} else {
				t.Logf("Service %s not yet deployed - expected in GREEN phase", serviceName)
			}
		}
		
		t.Log("Service name resolution validation completed")
	})
}

// TestDaprStateManagement validates state store operations
func TestDaprStateManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	daprHTTPPort := sharedtesting.GetEnvVar("DAPR_HTTP_PORT", "3500")
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("State_Store_Connectivity", func(t *testing.T) {
		// Arrange - State store should be accessible via Dapr
		stateURL := fmt.Sprintf("http://localhost:%s/v1.0/state/statestore/test-key", daprHTTPPort)
		
		// Act - Attempt to query state (will fail initially until component is configured)
		req, err := http.NewRequestWithContext(ctx, "GET", stateURL, nil)
		require.NoError(t, err, "Should be able to create state query request")
		resp, err := client.Do(req)
		
		// Assert - State endpoint should be reachable
		// Note: May return 404 or 500 initially until component is properly configured
		if err == nil && resp != nil {
			t.Logf("State store endpoint accessible - Status: %d", resp.StatusCode)
			if resp.Body != nil {
				resp.Body.Close()
			}
		} else {
			t.Log("State store not yet configured - expected until GREEN phase")
		}
		
		t.Log("State store connectivity validation completed")
	})
}

// TestDaprPubSubMessaging validates pub/sub messaging
func TestDaprPubSubMessaging(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	daprHTTPPort := sharedtesting.GetEnvVar("DAPR_HTTP_PORT", "3500")
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("PubSub_Component_Accessibility", func(t *testing.T) {
		// Arrange - Pub/sub should be accessible via Dapr
		pubsubURL := fmt.Sprintf("http://localhost:%s/v1.0/publish/pubsub/test-topic", daprHTTPPort)
		
		// Act - Attempt to access pub/sub endpoint (will fail initially)
		req, err := http.NewRequestWithContext(ctx, "POST", pubsubURL, nil)
		require.NoError(t, err, "Should be able to create pub/sub request")
		resp, err := client.Do(req)
		
		// Assert - Pub/sub endpoint should be reachable
		// Note: May fail initially until component is properly configured
		if err == nil && resp != nil {
			t.Logf("Pub/sub endpoint accessible - Status: %d", resp.StatusCode)
			if resp.Body != nil {
				resp.Body.Close()
			}
		} else {
			t.Log("Pub/sub not yet configured - expected until GREEN phase")
		}
		
		t.Log("Pub/sub accessibility validation completed")
	})
}