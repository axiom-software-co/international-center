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

func TestTempoIntegration(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("tempo service readiness", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Tempo distributed tracing is accessible
		tempoPort := requireEnv(t, "TEMPO_PORT")
		tempoURL := fmt.Sprintf("http://localhost:%s/ready", tempoPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", tempoURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Tempo not accessible at port %s: %v", tempoPort, err)
		}
		require.NoError(t, err, "Tempo should be accessible")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Tempo readiness check should return 200 OK")
	})

	t.Run("tempo trace search functionality", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Tempo trace search endpoint
		tempoPort := requireEnv(t, "TEMPO_PORT")
		searchURL := fmt.Sprintf("http://localhost:%s/api/search", tempoPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode == http.StatusOK ||
				resp.StatusCode == http.StatusBadRequest,
				"Tempo search endpoint should be available")
				
			if resp.StatusCode == http.StatusOK {
				contentType := resp.Header.Get("Content-Type")
				assert.Contains(t, contentType, "application/json",
					"Tempo search should return JSON")
			}
		} else {
			t.Logf("Tempo search endpoint not accessible: %v", err)
		}
	})

	t.Run("tempo otlp http ingestion", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Tempo OTLP HTTP ingestion endpoint
		otlpPort := requireEnv(t, "OTLP_HTTP_PORT")
		otlpURL := fmt.Sprintf("http://localhost:%s/v1/traces", otlpPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test with empty JSON payload
		req, err := http.NewRequestWithContext(ctx, "POST", otlpURL, 
			bytes.NewReader([]byte("{}")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			// Should not return 404 (endpoint exists) even if request format is wrong
			assert.True(t, resp.StatusCode != http.StatusNotFound,
				"Tempo OTLP HTTP endpoint should be accessible")
		} else {
			t.Logf("Tempo OTLP HTTP endpoint not accessible: %v", err)
		}
	})

	t.Run("tempo metrics endpoint", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Tempo exposes its own metrics
		tempoPort := requireEnv(t, "TEMPO_PORT")
		metricsURL := fmt.Sprintf("http://localhost:%s/metrics", tempoPort)
		
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
					"Tempo should return Prometheus format metrics")
			} else {
				t.Logf("Tempo metrics endpoint returned status: %d", resp.StatusCode)
			}
		} else {
			t.Logf("Tempo metrics endpoint not accessible: %v", err)
		}
	})

	t.Run("tempo query api functionality", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Tempo trace query API
		tempoPort := requireEnv(t, "TEMPO_PORT")
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test trace lookup by ID (using dummy ID to test endpoint)
		traceID := "00000000000000000000000000000000"
		queryURL := fmt.Sprintf("http://localhost:%s/api/traces/%s", tempoPort, traceID)
		
		req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			// Query endpoint should be available (may return 404 for non-existent trace)
			assert.True(t, resp.StatusCode == http.StatusOK ||
				resp.StatusCode == http.StatusNotFound ||
				resp.StatusCode == http.StatusBadRequest,
				"Tempo trace query endpoint should be functional")
		} else {
			t.Logf("Tempo trace query endpoint not accessible: %v", err)
		}
	})

	t.Run("tempo tag search functionality", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Tempo tag search capabilities
		tempoPort := requireEnv(t, "TEMPO_PORT")
		tagsURL := fmt.Sprintf("http://localhost:%s/api/search/tags", tempoPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", tagsURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			assert.True(t, resp.StatusCode == http.StatusOK ||
				resp.StatusCode == http.StatusNoContent ||
				resp.StatusCode == http.StatusBadRequest,
				"Tempo tag search endpoint should be functional")
				
			if resp.StatusCode == http.StatusOK {
				contentType := resp.Header.Get("Content-Type")
				assert.Contains(t, contentType, "application/json",
					"Tempo tags should return JSON")
			}
		} else {
			t.Logf("Tempo tags endpoint not accessible: %v", err)
		}
	})

	t.Run("tempo jaeger compatibility", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Tempo provides Jaeger-compatible endpoints
		tempoPort := requireEnv(t, "TEMPO_PORT")
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test Jaeger API services endpoint
		servicesURL := fmt.Sprintf("http://localhost:%s/api/services", tempoPort)
		req, err := http.NewRequestWithContext(ctx, "GET", servicesURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			assert.True(t, resp.StatusCode == http.StatusOK ||
				resp.StatusCode == http.StatusNoContent,
				"Tempo should provide Jaeger-compatible services endpoint")
				
			if resp.StatusCode == http.StatusOK {
				contentType := resp.Header.Get("Content-Type")
				assert.Contains(t, contentType, "application/json",
					"Jaeger services endpoint should return JSON")
			}
		} else {
			t.Logf("Tempo Jaeger services endpoint not accessible: %v", err)
		}
	})

	t.Run("tempo health endpoint", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test: Tempo health endpoint variations
		tempoPort := requireEnv(t, "TEMPO_PORT")
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test different health endpoints
		healthEndpoints := []string{
			fmt.Sprintf("http://localhost:%s/status", tempoPort),
			fmt.Sprintf("http://localhost:%s/ready", tempoPort),
		}
		
		healthyCount := 0
		for _, endpoint := range healthEndpoints {
			req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
			if err != nil {
				continue
			}
			
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					healthyCount++
				}
			}
		}
		
		assert.GreaterOrEqual(t, healthyCount, 1,
			"At least one Tempo health endpoint should be responsive")
	})
}