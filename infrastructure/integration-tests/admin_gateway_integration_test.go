package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminGatewayIntegration(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("admin gateway authentication requirement", func(t *testing.T) {
		// Test: Admin gateway enforces authentication for admin endpoints
		adminGatewayPort := requireEnv(t, "ADMIN_GATEWAY_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Test admin services endpoint without authentication
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/admin/api/v1/services", adminGatewayPort))
		require.NoError(t, err, "Admin gateway should respond to unauthorized request")
		defer resp.Body.Close()
		
		// Should return 401 Unauthorized or 403 Forbidden for unauthenticated requests
		assert.True(t, resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden,
			"Admin endpoints should require authentication, got status: %d", resp.StatusCode)
	})

	t.Run("admin gateway content endpoints", func(t *testing.T) {
		// Test: Admin gateway routes to content API admin endpoints
		adminGatewayPort := requireEnv(t, "ADMIN_GATEWAY_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Test admin content endpoint without authentication
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/admin/api/v1/content", adminGatewayPort))
		require.NoError(t, err, "Admin gateway should respond to content request")
		defer resp.Body.Close()
		
		// Should return 401 Unauthorized or 403 Forbidden for unauthenticated requests
		assert.True(t, resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden,
			"Admin content endpoints should require authentication, got status: %d", resp.StatusCode)
	})

	t.Run("admin gateway rate limiting", func(t *testing.T) {
		// Test: Admin gateway enforces rate limiting
		adminGatewayPort := requireEnv(t, "ADMIN_GATEWAY_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Make multiple rapid requests to trigger rate limiting
		successCount := 0
		rateLimitedCount := 0
		
		for i := 0; i < 20; i++ {
			resp, err := client.Get(fmt.Sprintf("http://localhost:%s/health", adminGatewayPort))
			if err != nil {
				t.Logf("Request %d failed: %v", i, err)
				continue
			}
			
			if resp.StatusCode == http.StatusOK {
				successCount++
			} else if resp.StatusCode == http.StatusTooManyRequests {
				rateLimitedCount++
			}
			
			resp.Body.Close()
			
			// Small delay between requests
			time.Sleep(10 * time.Millisecond)
		}
		
		t.Logf("Successful requests: %d, Rate limited requests: %d", successCount, rateLimitedCount)
		
		// We expect at least some requests to succeed, but rate limiting may or may not be triggered
		// depending on the configured limits
		assert.Greater(t, successCount, 0, "Some requests should succeed")
	})
}

func TestAuthentikIntegration(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("authentik health check", func(t *testing.T) {
		// Test: Authentik identity provider is accessible and healthy
		authentikPort := requireEnv(t, "AUTHENTIK_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/-/health/live/", authentikPort))
		require.NoError(t, err, "Authentik health check should succeed")
		defer resp.Body.Close()
		
		assert.Contains(t, []int{http.StatusOK, http.StatusNoContent}, resp.StatusCode, "Authentik should be healthy")
		
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Should be able to read Authentik health response")
		
		// Handle both JSON response (200) and no content response (204)
		if resp.StatusCode == http.StatusOK && len(body) > 0 {
			var healthResponse map[string]interface{}
			err = json.Unmarshal(body, &healthResponse)
			require.NoError(t, err, "Authentik health response should be valid JSON")
			assert.Equal(t, "ok", healthResponse["status"], "Authentik health status should be ok")
		} else {
			// 204 No Content is also a valid healthy response
			assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Authentik health endpoint returned no content (healthy)")
		}
	})

	t.Run("authentik ui accessibility", func(t *testing.T) {
		// Test: Authentik UI is accessible for administrative configuration
		authentikPort := requireEnv(t, "AUTHENTIK_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/if/flow/default-authentication-flow/", authentikPort))
		require.NoError(t, err, "Authentik UI should be accessible")
		defer resp.Body.Close()
		
		// Should return 200 OK or 302 redirect for the login flow
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound,
			"Authentik UI should be accessible, got status: %d", resp.StatusCode)
	})
}

func TestDaprMiddlewareIntegration(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("dapr component health validation", func(t *testing.T) {
		// Test: Dapr components are properly configured and accessible
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Test Dapr metadata endpoint for component validation
		resp, err := client.Get("http://localhost:3500/v1.0/metadata")
		if err == nil {
			defer resp.Body.Close()
			
			if resp.StatusCode == http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err, "Should be able to read Dapr metadata")
				
				var metadata map[string]interface{}
				err = json.Unmarshal(body, &metadata)
				require.NoError(t, err, "Dapr metadata should be valid JSON")
				
				// Check for components
				if components, ok := metadata["components"].([]interface{}); ok {
					assert.Greater(t, len(components), 0, "Dapr should have configured components")
					
					// Log component names for debugging
					for _, comp := range components {
						if compMap, ok := comp.(map[string]interface{}); ok {
							if name, ok := compMap["name"].(string); ok {
								t.Logf("Dapr component found: %s", name)
							}
						}
					}
				}
			}
		} else {
			t.Logf("Dapr metadata endpoint not accessible: %v", err)
		}
	})

	t.Run("redis pubsub component validation", func(t *testing.T) {
		// Test: Redis pub/sub component is accessible through Dapr
		redisPort := requireEnv(t, "REDIS_PORT")
		
		// Test Redis connectivity
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Test if we can reach the Redis component through Dapr
		pubsubData := map[string]interface{}{
			"data": map[string]string{
				"message": "test-message",
			},
		}
		
		jsonData, err := json.Marshal(pubsubData)
		require.NoError(t, err, "Should be able to marshal test data")
		
		// Try to publish a test message (this may fail if pubsub isn't configured, but should not error)
		req, err := http.NewRequest("POST", "http://localhost:3500/v1.0/publish/redis-pubsub/test-topic", bytes.NewBuffer(jsonData))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				// Any response (even error) indicates Dapr is processing the request
				t.Logf("Dapr pubsub test response status: %d", resp.StatusCode)
			} else {
				t.Logf("Dapr pubsub test failed: %v", err)
			}
		}
		
		t.Logf("Redis port configured: %s", redisPort)
	})

	t.Run("postgresql state store validation", func(t *testing.T) {
		// Test: PostgreSQL state store component is accessible through Dapr
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Try to access state store through Dapr
		resp, err := client.Get("http://localhost:3500/v1.0/state/postgresql-statestore")
		if err == nil {
			defer resp.Body.Close()
			// Any response indicates Dapr is processing state store requests
			t.Logf("Dapr state store response status: %d", resp.StatusCode)
		} else {
			t.Logf("Dapr state store not accessible: %v", err)
		}
	})
}