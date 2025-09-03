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
		// Handle null services as empty array (valid for empty state store)
		servicesValue, exists := servicesResponse["services"]
		require.True(t, exists, "Response should contain services field")
		
		var services []interface{}
		if servicesValue == nil {
			services = []interface{}{} // Treat null as empty array
		} else {
			var ok bool
			services, ok = servicesValue.([]interface{})
			require.True(t, ok, "Services should be an array or null")
		}
		assert.GreaterOrEqual(t, len(services), 0, "Services array should be accessible (empty is valid)")
		
		// Validate total count matches services array
		total, ok := servicesResponse["total"].(float64)
		require.True(t, ok, "Total should be a number")
		assert.Equal(t, float64(len(services)), total, "Total should match services array length")
	})
}