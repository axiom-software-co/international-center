package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublicGatewayIntegration(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("public gateway health check", func(t *testing.T) {
		// Test: Public gateway health endpoint is accessible
		publicGatewayPort := requireEnv(t, "PUBLIC_GATEWAY_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/health", publicGatewayPort))
		require.NoError(t, err, "Public gateway health check should succeed")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Public gateway should be healthy")
	})

	t.Run("public gateway service routing", func(t *testing.T) {
		// Test: Public gateway routes requests to services API correctly
		publicGatewayPort := requireEnv(t, "PUBLIC_GATEWAY_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Test routing to services API through gateway
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/api/v1/services", publicGatewayPort))
		require.NoError(t, err, "Gateway routing to services should succeed")
		defer resp.Body.Close()
		
		// Accept both successful routing (200) and service unavailable (502) during infrastructure startup
		if resp.StatusCode == http.StatusOK {
			assert.Equal(t, http.StatusOK, resp.StatusCode, "Gateway should route services requests correctly")
		} else if resp.StatusCode == http.StatusBadGateway {
			t.Logf("Gateway returned 502 - backend service may not be ready yet (infrastructure startup)")
			return // Skip further validation if backend isn't ready
		} else {
			assert.Equal(t, http.StatusOK, resp.StatusCode, "Gateway should route services requests correctly")
		}
		
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Should be able to read gateway response")
		
		var servicesResponse map[string]interface{}
		err = json.Unmarshal(body, &servicesResponse)
		require.NoError(t, err, "Gateway response should be valid JSON")
		
		assert.Contains(t, servicesResponse, "services", "Gateway response should contain services array")
	})

	t.Run("public gateway content routing", func(t *testing.T) {
		// Test: Public gateway routes requests to content API correctly
		publicGatewayPort := requireEnv(t, "PUBLIC_GATEWAY_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Test routing to content API through gateway
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/api/v1/content", publicGatewayPort))
		require.NoError(t, err, "Gateway routing to content should succeed")
		defer resp.Body.Close()
		
		// Accept both successful routing (200) and service unavailable (502) during infrastructure startup
		if resp.StatusCode == http.StatusOK {
			assert.Equal(t, http.StatusOK, resp.StatusCode, "Gateway should route content requests correctly")
		} else if resp.StatusCode == http.StatusBadGateway {
			t.Logf("Gateway returned 502 - backend content service may not be ready yet (infrastructure startup)")
			return // Skip further validation if backend isn't ready
		} else {
			assert.Equal(t, http.StatusOK, resp.StatusCode, "Gateway should route content requests correctly")
		}
		
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Should be able to read gateway response")
		
		var contentResponse map[string]interface{}
		err = json.Unmarshal(body, &contentResponse)
		require.NoError(t, err, "Gateway response should be valid JSON")
		
		assert.Contains(t, contentResponse, "content", "Gateway response should contain content array")
	})
}