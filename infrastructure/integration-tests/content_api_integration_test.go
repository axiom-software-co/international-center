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