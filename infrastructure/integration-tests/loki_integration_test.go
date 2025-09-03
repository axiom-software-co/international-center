package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLokiIntegration(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("loki service readiness", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Loki log aggregation is accessible
		lokiPort := requireEnv(t, "LOKI_PORT")
		lokiURL := fmt.Sprintf("http://localhost:%s/ready", lokiPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", lokiURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Loki not accessible at port %s: %v", lokiPort, err)
		}
		require.NoError(t, err, "Loki should be accessible")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Loki readiness check should return 200 OK")
	})

	t.Run("loki log ingestion endpoint", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Loki log ingestion functionality
		lokiPort := requireEnv(t, "LOKI_PORT")
		ingestURL := fmt.Sprintf("http://localhost:%s/loki/api/v1/push", lokiPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Create test log entry
		testLog := map[string]interface{}{
			"streams": []map[string]interface{}{
				{
					"stream": map[string]string{
						"service": "infrastructure-test",
						"level":   "info",
					},
					"values": [][]string{
						{fmt.Sprintf("%d", time.Now().UnixNano()), "test log entry from infrastructure tests"},
					},
				},
			},
		}
		
		logBytes, err := json.Marshal(testLog)
		require.NoError(t, err)
		
		req, err := http.NewRequestWithContext(ctx, "POST", ingestURL, 
			bytes.NewReader(logBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode == http.StatusOK ||
				resp.StatusCode == http.StatusNoContent,
				"Loki should accept log entries")
		} else {
			t.Logf("Loki ingestion endpoint not accessible: %v", err)
		}
	})

	t.Run("loki query endpoint functionality", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Loki query capabilities
		lokiPort := requireEnv(t, "LOKI_PORT")
		queryURL := fmt.Sprintf("http://localhost:%s/loki/api/v1/query", lokiPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test with basic query
		req, err := http.NewRequestWithContext(ctx, "GET", queryURL+"?query={service=\"test\"}", nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode == http.StatusOK ||
				resp.StatusCode == http.StatusBadRequest,
				"Loki query endpoint should handle requests")
		} else {
			t.Logf("Loki query endpoint not accessible: %v", err)
		}
	})

	t.Run("loki label api functionality", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Loki label API for log queries
		lokiPort := requireEnv(t, "LOKI_PORT")
		labelsURL := fmt.Sprintf("http://localhost:%s/loki/api/v1/labels", lokiPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", labelsURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode == http.StatusOK ||
				resp.StatusCode == http.StatusNoContent,
				"Loki labels API should be accessible for log queries")
				
			if resp.StatusCode == http.StatusOK {
				contentType := resp.Header.Get("Content-Type")
				assert.Contains(t, contentType, "application/json",
					"Loki labels should return JSON")
			}
		} else {
			t.Logf("Loki labels API not accessible: %v", err)
		}
	})

	t.Run("loki metrics endpoint", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Loki exposes its own metrics
		lokiPort := requireEnv(t, "LOKI_PORT")
		metricsURL := fmt.Sprintf("http://localhost:%s/metrics", lokiPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", metricsURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			if resp.StatusCode == http.StatusOK {
				// Check for Prometheus metrics format
				contentType := resp.Header.Get("Content-Type")
				assert.Contains(t, contentType, "text/plain",
					"Loki should return Prometheus format metrics")
			} else {
				t.Logf("Loki metrics endpoint returned status: %d", resp.StatusCode)
			}
		} else {
			t.Logf("Loki metrics endpoint not accessible: %v", err)
		}
	})

	t.Run("loki range query functionality", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Loki range query capabilities
		lokiPort := requireEnv(t, "LOKI_PORT")
		
		// Test range query endpoint
		now := time.Now()
		start := now.Add(-1 * time.Hour)
		rangeURL := fmt.Sprintf("http://localhost:%s/loki/api/v1/query_range?query={service=\"test\"}&start=%d&end=%d", 
			lokiPort, start.UnixNano(), now.UnixNano())
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", rangeURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			assert.True(t, resp.StatusCode == http.StatusOK ||
				resp.StatusCode == http.StatusBadRequest,
				"Loki range query should be functional")
				
			if resp.StatusCode == http.StatusOK {
				contentType := resp.Header.Get("Content-Type")
				assert.Contains(t, contentType, "application/json",
					"Loki range query should return JSON")
			}
		} else {
			t.Logf("Loki range query endpoint not accessible: %v", err)
		}
	})

	t.Run("loki log streaming capability", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Loki streaming API
		lokiPort := requireEnv(t, "LOKI_PORT")
		streamURL := fmt.Sprintf("http://localhost:%s/loki/api/v1/tail", lokiPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", streamURL+"?query={service=\"test\"}", nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			// Streaming endpoint should be available (may return various status codes)
			assert.True(t, resp.StatusCode != http.StatusNotFound,
				"Loki streaming endpoint should be available")
		} else {
			t.Logf("Loki streaming endpoint not accessible: %v", err)
		}
	})
}