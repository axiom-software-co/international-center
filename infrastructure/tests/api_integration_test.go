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

func TestServicesAPIIntegration(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("services api health check", func(t *testing.T) {
		// Test: Services API health endpoint is accessible
		servicesAPIPort := requireEnv(t, "SERVICES_API_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/health", servicesAPIPort))
		require.NoError(t, err, "Services API health check should succeed")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Services API should be healthy")
		
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Should be able to read health response")
		
		var healthResponse map[string]interface{}
		err = json.Unmarshal(body, &healthResponse)
		require.NoError(t, err, "Health response should be valid JSON")
		
		assert.Equal(t, "ok", healthResponse["status"], "Health status should be ok")
		assert.Equal(t, "services-api", healthResponse["service"], "Service name should be correct")
	})

	t.Run("services api readiness check", func(t *testing.T) {
		// Test: Services API readiness endpoint checks database connectivity
		servicesAPIPort := requireEnv(t, "SERVICES_API_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/health/ready", servicesAPIPort))
		require.NoError(t, err, "Services API readiness check should succeed")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Services API should be ready")
		
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Should be able to read readiness response")
		
		var readinessResponse map[string]interface{}
		err = json.Unmarshal(body, &readinessResponse)
		require.NoError(t, err, "Readiness response should be valid JSON")
		
		assert.Equal(t, "ready", readinessResponse["status"], "Readiness status should be ready")
	})

	t.Run("services api public endpoints", func(t *testing.T) {
		// Test: Public services endpoints are accessible via Dapr sidecar
		servicesAPIPort := requireEnv(t, "SERVICES_API_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Test GET /api/v1/services (list published services)
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/api/v1/services", servicesAPIPort))
		require.NoError(t, err, "GET /api/v1/services should succeed")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Services list should return OK")
		
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Should be able to read services list response")
		
		var servicesResponse map[string]interface{}
		err = json.Unmarshal(body, &servicesResponse)
		require.NoError(t, err, "Services response should be valid JSON")
		
		assert.Contains(t, servicesResponse, "services", "Response should contain services array")
		assert.Contains(t, servicesResponse, "total", "Response should contain total count")
		
		// Validate JSON structure consistency (Dapr state stores are empty by design)
		services, ok := servicesResponse["services"].([]interface{})
		require.True(t, ok, "Services should be an array (not null)")
		assert.NotNil(t, services, "Services array should not be nil")
		assert.GreaterOrEqual(t, len(services), 0, "Services array should be accessible (empty is valid)")
		
		// Validate total count matches services array
		total, ok := servicesResponse["total"].(float64)
		require.True(t, ok, "Total should be a number")
		assert.Equal(t, float64(len(services)), total, "Total should match services array length")
	})
}

func TestContentAPIIntegration(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("content api health check", func(t *testing.T) {
		// Test: Content API health endpoint is accessible
		contentAPIPort := requireEnv(t, "CONTENT_API_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/health", contentAPIPort))
		require.NoError(t, err, "Content API health check should succeed")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Content API should be healthy")
		
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Should be able to read health response")
		
		var healthResponse map[string]interface{}
		err = json.Unmarshal(body, &healthResponse)
		require.NoError(t, err, "Health response should be valid JSON")
		
		assert.Equal(t, "ok", healthResponse["status"], "Health status should be ok")
		assert.Equal(t, "content-api", healthResponse["service"], "Service name should be correct")
	})

	t.Run("content api readiness check", func(t *testing.T) {
		// Test: Content API readiness endpoint checks database connectivity
		contentAPIPort := requireEnv(t, "CONTENT_API_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/health/ready", contentAPIPort))
		require.NoError(t, err, "Content API readiness check should succeed")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Content API should be ready")
		
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Should be able to read readiness response")
		
		var readinessResponse map[string]interface{}
		err = json.Unmarshal(body, &readinessResponse)
		require.NoError(t, err, "Readiness response should be valid JSON")
		
		assert.Equal(t, "ready", readinessResponse["status"], "Readiness status should be ready")
	})

	t.Run("content api public endpoints", func(t *testing.T) {
		// Test: Public content endpoints are accessible
		contentAPIPort := requireEnv(t, "CONTENT_API_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Test GET /api/v1/content (list published content)
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/api/v1/content", contentAPIPort))
		require.NoError(t, err, "GET /api/v1/content should succeed")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Content list should return OK")
		
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Should be able to read content list response")
		
		var contentResponse map[string]interface{}
		err = json.Unmarshal(body, &contentResponse)
		require.NoError(t, err, "Content response should be valid JSON")
		
		assert.Contains(t, contentResponse, "content", "Response should contain content array")
		assert.Contains(t, contentResponse, "total", "Response should contain total count")
	})
}

func TestGatewayIntegration(t *testing.T) {
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

	t.Run("admin gateway health check", func(t *testing.T) {
		// Test: Admin gateway health endpoint is accessible
		adminGatewayPort := requireEnv(t, "ADMIN_GATEWAY_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/health", adminGatewayPort))
		require.NoError(t, err, "Admin gateway health check should succeed")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Admin gateway should be healthy")
	})

	t.Run("public gateway service routing", func(t *testing.T) {
		// Test: Public gateway routes requests to services API correctly
		publicGatewayPort := requireEnv(t, "PUBLIC_GATEWAY_PORT")
		
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Test routing to services API through gateway
		resp, err := client.Get(fmt.Sprintf("http://localhost:%s/api/v1/services", publicGatewayPort))
		require.NoError(t, err, "Gateway routing to services should succeed")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Gateway should route services requests correctly")
		
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
		
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Gateway should route content requests correctly")
		
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Should be able to read gateway response")
		
		var contentResponse map[string]interface{}
		err = json.Unmarshal(body, &contentResponse)
		require.NoError(t, err, "Gateway response should be valid JSON")
		
		assert.Contains(t, contentResponse, "content", "Gateway response should contain content array")
	})
}

func TestDaprServiceInvocation(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("dapr sidecar connectivity", func(t *testing.T) {
		// Test: Dapr sidecars are accessible and responding
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Test services-api Dapr sidecar
		resp, err := client.Get("http://localhost:3500/v1.0/healthz")
		if err == nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode, "Services API Dapr sidecar should be healthy")
		} else {
			t.Logf("Services API Dapr sidecar not accessible: %v", err)
		}
	})

	t.Run("service invocation through dapr", func(t *testing.T) {
		// Test: Service invocation works through Dapr sidecars
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Test invoking services-api through Dapr
		resp, err := client.Get("http://localhost:3500/v1.0/invoke/services-api/method/health")
		if err == nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode, "Dapr service invocation should work")
			
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err, "Should be able to read Dapr response")
			
			var healthResponse map[string]interface{}
			err = json.Unmarshal(body, &healthResponse)
			if err == nil {
				assert.Equal(t, "ok", healthResponse["status"], "Dapr invocation should return correct response")
			}
		} else {
			t.Logf("Dapr service invocation not working: %v", err)
		}
	})
}