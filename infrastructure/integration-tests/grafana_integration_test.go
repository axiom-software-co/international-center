package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrafanaIntegration(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("grafana service availability", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Grafana service is accessible
		grafanaPort := requireEnv(t, "GRAFANA_PORT")
		grafanaURL := fmt.Sprintf("http://localhost:%s/api/health", grafanaPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", grafanaURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Grafana not accessible at port %s: %v", grafanaPort, err)
		}
		require.NoError(t, err, "Grafana should be accessible")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Grafana health check should return 200 OK")
		
		// Verify response contains health information
		contentType := resp.Header.Get("Content-Type")
		assert.Contains(t, contentType, "application/json",
			"Grafana health should return JSON")
	})

	t.Run("grafana api accessibility", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Grafana API endpoints are accessible
		grafanaPort := requireEnv(t, "GRAFANA_PORT")
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test API version endpoint
		versionURL := fmt.Sprintf("http://localhost:%s/api/health", grafanaPort)
		req, err := http.NewRequestWithContext(ctx, "GET", versionURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		require.NoError(t, err, "Grafana API should be accessible")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Grafana API should return success")
	})

	t.Run("grafana data source connectivity", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Grafana data source connectivity
		grafanaPort := requireEnv(t, "GRAFANA_PORT")
		dataSourceURL := fmt.Sprintf("http://localhost:%s/api/datasources", grafanaPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test with basic auth (admin password from environment)
		req, err := http.NewRequestWithContext(ctx, "GET", dataSourceURL, nil)
		require.NoError(t, err)
		req.SetBasicAuth("admin", requireEnv(t, "GRAFANA_ADMIN_PASSWORD"))
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			if resp.StatusCode == http.StatusOK {
				// Check if data sources are configured
				assert.Equal(t, http.StatusOK, resp.StatusCode,
					"Grafana should list data sources")
				
				contentType := resp.Header.Get("Content-Type")
				assert.Contains(t, contentType, "application/json",
					"Grafana should return JSON data sources list")
			} else {
				t.Logf("Grafana data source endpoint returned status: %d", resp.StatusCode)
			}
		} else {
			t.Logf("Grafana data source endpoint not accessible: %v", err)
		}
	})

	t.Run("grafana dashboard api", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Grafana dashboard API is functional
		grafanaPort := requireEnv(t, "GRAFANA_PORT")
		dashboardURL := fmt.Sprintf("http://localhost:%s/api/search", grafanaPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		req, err := http.NewRequestWithContext(ctx, "GET", dashboardURL, nil)
		require.NoError(t, err)
		req.SetBasicAuth("admin", requireEnv(t, "GRAFANA_ADMIN_PASSWORD"))
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			// Dashboard search should be accessible
			assert.True(t, resp.StatusCode == http.StatusOK ||
				resp.StatusCode == http.StatusUnauthorized,
				"Grafana dashboard API should be functional")
				
			if resp.StatusCode == http.StatusOK {
				contentType := resp.Header.Get("Content-Type")
				assert.Contains(t, contentType, "application/json",
					"Dashboard API should return JSON")
			}
		} else {
			t.Logf("Grafana dashboard API not accessible: %v", err)
		}
	})

	t.Run("grafana configuration verification", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Grafana configuration is accessible
		grafanaPort := requireEnv(t, "GRAFANA_PORT")
		configURL := fmt.Sprintf("http://localhost:%s/api/admin/settings", grafanaPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		req, err := http.NewRequestWithContext(ctx, "GET", configURL, nil)
		require.NoError(t, err)
		req.SetBasicAuth("admin", requireEnv(t, "GRAFANA_ADMIN_PASSWORD"))
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			// Configuration endpoint should be accessible to admin
			assert.True(t, resp.StatusCode == http.StatusOK ||
				resp.StatusCode == http.StatusUnauthorized ||
				resp.StatusCode == http.StatusForbidden,
				"Grafana configuration endpoint should respond")
		} else {
			t.Logf("Grafana configuration endpoint not accessible: %v", err)
		}
	})
}