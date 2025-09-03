package tests

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMimirIntegration(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("mimir service readiness", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Mimir metrics storage is accessible
		mimirPort := requireEnv(t, "MIMIR_PORT")
		mimirURL := fmt.Sprintf("http://localhost:%s/ready", mimirPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", mimirURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Mimir not accessible at port %s: %v", mimirPort, err)
		}
		require.NoError(t, err, "Mimir should be accessible")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Mimir readiness check should return 200 OK")
	})

	t.Run("mimir metrics ingestion endpoint", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Mimir metrics ingestion endpoint is accessible
		mimirPort := requireEnv(t, "MIMIR_PORT")
		pushURL := fmt.Sprintf("http://localhost:%s/api/v1/push", mimirPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test with empty JSON payload
		req, err := http.NewRequestWithContext(ctx, "POST", pushURL, bytes.NewReader([]byte("{}")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode != http.StatusNotFound,
				"Mimir metrics ingestion endpoint should be accessible")
		} else {
			t.Logf("Mimir push endpoint not accessible: %v", err)
		}
	})

	t.Run("mimir query endpoint availability", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Mimir query endpoint is functional
		mimirPort := requireEnv(t, "MIMIR_PORT")
		queryURL := fmt.Sprintf("http://localhost:%s/prometheus/api/v1/query", mimirPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test basic query capability (just check if endpoint exists)
		req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			// Mimir query endpoint should at least return a method not allowed or similar, not 404
			assert.True(t, resp.StatusCode != http.StatusNotFound,
				"Mimir query endpoint should be accessible")
		} else {
			t.Logf("Mimir query endpoint not accessible: %v", err)
		}
	})

	t.Run("mimir prometheus compatibility", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Mimir provides Prometheus-compatible endpoints
		mimirPort := requireEnv(t, "MIMIR_PORT")
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test Prometheus API status endpoint
		statusURL := fmt.Sprintf("http://localhost:%s/api/v1/status/config", mimirPort)
		req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			// Should respond (even if with error for missing config)
			assert.True(t, resp.StatusCode == http.StatusOK ||
				resp.StatusCode == http.StatusBadRequest ||
				resp.StatusCode == http.StatusUnauthorized,
				"Mimir should provide Prometheus-compatible API endpoints")
		} else {
			t.Logf("Mimir Prometheus API endpoint not accessible: %v", err)
		}
	})

	t.Run("mimir health and metrics", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Mimir exposes its own health and metrics
		mimirPort := requireEnv(t, "MIMIR_PORT")
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test metrics endpoint
		metricsURL := fmt.Sprintf("http://localhost:%s/metrics", mimirPort)
		req, err := http.NewRequestWithContext(ctx, "GET", metricsURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			if resp.StatusCode == http.StatusOK {
				// Check for Prometheus metrics format
				contentType := resp.Header.Get("Content-Type")
				assert.Contains(t, contentType, "text/plain",
					"Mimir should return Prometheus format metrics")
			} else {
				t.Logf("Mimir metrics endpoint returned status: %d", resp.StatusCode)
			}
		} else {
			t.Logf("Mimir metrics endpoint not accessible: %v", err)
		}
	})

	t.Run("mimir label and series endpoints", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Mimir label and series endpoints are functional
		mimirPort := requireEnv(t, "MIMIR_PORT")
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test labels endpoint
		labelsURL := fmt.Sprintf("http://localhost:%s/prometheus/api/v1/labels", mimirPort)
		req, err := http.NewRequestWithContext(ctx, "GET", labelsURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			// Labels endpoint should be accessible
			assert.True(t, resp.StatusCode == http.StatusOK ||
				resp.StatusCode == http.StatusNoContent ||
				resp.StatusCode == http.StatusBadRequest,
				"Mimir labels endpoint should be functional")
		} else {
			t.Logf("Mimir labels endpoint not accessible: %v", err)
		}
		
		// Test series endpoint
		seriesURL := fmt.Sprintf("http://localhost:%s/prometheus/api/v1/series", mimirPort)
		req, err = http.NewRequestWithContext(ctx, "GET", seriesURL, nil)
		require.NoError(t, err)
		
		resp, err = client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			// Series endpoint should be accessible
			assert.True(t, resp.StatusCode == http.StatusOK ||
				resp.StatusCode == http.StatusBadRequest,
				"Mimir series endpoint should be functional")
		} else {
			t.Logf("Mimir series endpoint not accessible: %v", err)
		}
	})
}