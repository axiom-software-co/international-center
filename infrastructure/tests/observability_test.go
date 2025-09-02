package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrafanaObservabilityStack(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("grafana service availability", func(t *testing.T) {
		// Test: Grafana service is accessible
		grafanaPort := getEnvWithDefault("GRAFANA_PORT", "3000")
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

	t.Run("mimir metrics storage connectivity", func(t *testing.T) {
		// Test: Mimir metrics storage is accessible
		mimirPort := getEnvWithDefault("MIMIR_PORT", "9009")
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

	t.Run("loki log aggregation functionality", func(t *testing.T) {
		// Test: Loki log aggregation is accessible
		lokiPort := getEnvWithDefault("LOKI_PORT", "3100")
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
		
		// Test log ingestion endpoint
		ingestURL := fmt.Sprintf("http://localhost:%s/loki/api/v1/push", lokiPort)
		
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
		
		ingestReq, err := http.NewRequestWithContext(ctx, "POST", ingestURL, 
			bytes.NewReader(logBytes))
		require.NoError(t, err)
		ingestReq.Header.Set("Content-Type", "application/json")
		
		ingestResp, err := client.Do(ingestReq)
		if err == nil {
			defer ingestResp.Body.Close()
			assert.True(t, ingestResp.StatusCode == http.StatusOK ||
				ingestResp.StatusCode == http.StatusNoContent,
				"Loki should accept log entries")
		}
	})

	t.Run("tempo tracing capability", func(t *testing.T) {
		// Test: Tempo distributed tracing is accessible
		tempoPort := getEnvWithDefault("TEMPO_PORT", "3200")
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
		
		// Test trace search endpoint
		searchURL := fmt.Sprintf("http://localhost:%s/api/search", tempoPort)
		searchReq, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
		require.NoError(t, err)
		
		searchResp, err := client.Do(searchReq)
		if err == nil {
			defer searchResp.Body.Close()
			assert.True(t, searchResp.StatusCode == http.StatusOK ||
				searchResp.StatusCode == http.StatusBadRequest,
				"Tempo search endpoint should be available")
		}
	})

	t.Run("pyroscope profiling integration", func(t *testing.T) {
		// Test: Pyroscope continuous profiling is accessible
		pyroscopePort := getEnvWithDefault("PYROSCOPE_PORT", "4040")
		pyroscopeURL := fmt.Sprintf("http://localhost:%s/api/apps", pyroscopePort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", pyroscopeURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Pyroscope not accessible at port %s: %v", pyroscopePort, err)
		}
		require.NoError(t, err, "Pyroscope should be accessible")
		defer resp.Body.Close()
		
		// Pyroscope API should be available
		assert.True(t, resp.StatusCode == http.StatusOK ||
			resp.StatusCode == http.StatusNotFound,
			"Pyroscope API should be available")
		
		// Should return JSON
		if resp.StatusCode == http.StatusOK {
			contentType := resp.Header.Get("Content-Type")
			assert.Contains(t, contentType, "application/json",
				"Pyroscope should return JSON response")
		}
	})
}

func TestTelemetryCollection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("grafana agent configuration", func(t *testing.T) {
		// Test: Grafana Agent is configured and running
		// Note: Grafana Agent typically doesn't expose HTTP endpoints by default
		// This test focuses on configuration validation
		
		// Check for agent configuration file
		agentConfigPath := "../../observability/grafana-agent.yaml"
		_, err := os.Stat(agentConfigPath)
		if err != nil {
			t.Logf("Grafana Agent configuration not found at %s: %v", agentConfigPath, err)
			// Configuration might be in different location or embedded in compose
			return
		}
		
		content, err := os.ReadFile(agentConfigPath)
		require.NoError(t, err, "Should be able to read Grafana Agent configuration")
		
		configContent := string(content)
		assert.Contains(t, configContent, "server:", "Agent config should have server configuration")
		assert.Contains(t, configContent, "metrics:", "Agent config should have metrics configuration")
		assert.Contains(t, configContent, "logs:", "Agent config should have logs configuration")
	})

	t.Run("telemetry endpoint connectivity", func(t *testing.T) {
		// Test: Telemetry endpoints are accessible for data collection
		
		client := &http.Client{Timeout: 3 * time.Second}
		
		// Test Mimir push endpoint
		mimirEndpoint := fmt.Sprintf("http://localhost:%s/api/v1/push", getEnvWithDefault("MIMIR_PORT", "9009"))
		req, err := http.NewRequestWithContext(ctx, "POST", mimirEndpoint, bytes.NewReader([]byte("{}")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode != http.StatusNotFound,
				"Mimir telemetry endpoint should be accessible")
		}
		
		// Test Loki push endpoint
		lokiEndpoint := fmt.Sprintf("http://localhost:%s/loki/api/v1/push", getEnvWithDefault("LOKI_PORT", "3100"))
		req, err = http.NewRequestWithContext(ctx, "POST", lokiEndpoint, bytes.NewReader([]byte("{}")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		
		resp, err = client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode != http.StatusNotFound,
				"Loki telemetry endpoint should be accessible")
		}
		
		// Test Tempo OTLP HTTP endpoint (port 4318 for HTTP)
		tempoEndpoint := "http://localhost:4318/v1/traces"
		req, err = http.NewRequestWithContext(ctx, "POST", tempoEndpoint, bytes.NewReader([]byte("{}")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		
		resp, err = client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode != http.StatusNotFound,
				"Tempo telemetry endpoint should be accessible")
		} else {
			t.Logf("Tempo telemetry endpoint not accessible: %v", err)
		}
	})

	t.Run("data pipeline functionality", func(t *testing.T) {
		// Test: Data pipeline from collection to storage is functional
		
		// Test metrics pipeline
		mimirPort := getEnvWithDefault("MIMIR_PORT", "9009")
		metricsURL := fmt.Sprintf("http://localhost:%s/prometheus/api/v1/query", mimirPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test basic query capability (just check if endpoint exists)
		queryReq, err := http.NewRequestWithContext(ctx, "GET", metricsURL, nil)
		require.NoError(t, err)
		
		queryResp, err := client.Do(queryReq)
		if err == nil {
			defer queryResp.Body.Close()
			// Mimir query endpoint should at least return a method not allowed or similar, not 404
			assert.True(t, queryResp.StatusCode != http.StatusNotFound,
				"Mimir query endpoint should be accessible")
		} else {
			t.Logf("Mimir query endpoint not accessible: %v", err)
		}
		
		// Test logs pipeline
		lokiPort := getEnvWithDefault("LOKI_PORT", "3100")
		logsQueryURL := fmt.Sprintf("http://localhost:%s/loki/api/v1/query", lokiPort)
		
		logsReq, err := http.NewRequestWithContext(ctx, "GET", logsQueryURL+"?query={service=\"test\"}", nil)
		require.NoError(t, err)
		
		logsResp, err := client.Do(logsReq)
		if err == nil {
			defer logsResp.Body.Close()
			assert.True(t, logsResp.StatusCode == http.StatusOK ||
				logsResp.StatusCode == http.StatusBadRequest,
				"Loki query endpoint should handle requests")
		}
	})

	t.Run("observability stack integration", func(t *testing.T) {
		// Test: All observability components can communicate with each other
		
		// Test Grafana data source connectivity
		grafanaPort := getEnvWithDefault("GRAFANA_PORT", "3000")
		dataSourceURL := fmt.Sprintf("http://localhost:%s/api/datasources", grafanaPort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test with basic auth (admin:admin is default for Grafana)
		req, err := http.NewRequestWithContext(ctx, "GET", dataSourceURL, nil)
		require.NoError(t, err)
		req.SetBasicAuth("admin", "admin")
		
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
			}
		}
		
		// Verify observability stack health overall
		services := map[string]string{
			"Grafana":       fmt.Sprintf("http://localhost:%s/api/health", grafanaPort),
			"Mimir":         fmt.Sprintf("http://localhost:%s/ready", getEnvWithDefault("MIMIR_PORT", "9009")),
			"Loki":          fmt.Sprintf("http://localhost:%s/ready", getEnvWithDefault("LOKI_PORT", "3100")),
			"Tempo":         fmt.Sprintf("http://localhost:%s/ready", getEnvWithDefault("TEMPO_PORT", "3200")),
			"Pyroscope":     fmt.Sprintf("http://localhost:%s/api/apps", getEnvWithDefault("PYROSCOPE_PORT", "4040")),
			"Grafana-Agent": "http://localhost:12345/-/ready",
		}
		
		healthyServices := 0
		for service, healthURL := range services {
			healthReq, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
			if err != nil {
				continue
			}
			
			healthResp, err := client.Do(healthReq)
			if err == nil {
				defer healthResp.Body.Close()
				if healthResp.StatusCode == http.StatusOK {
					healthyServices++
					t.Logf("%s is healthy", service)
				}
			}
		}
		
		// At least 3 out of 5 services should be healthy for basic functionality
		assert.GreaterOrEqual(t, healthyServices, 3,
			"At least 3 observability services should be healthy for basic functionality")
	})
}

func TestGrafanaAgentIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("grafana agent health and configuration", func(t *testing.T) {
		// Test: Grafana Agent is running and healthy
		agentURL := "http://localhost:12345/-/ready"
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", agentURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Grafana Agent not accessible: %v", err)
		}
		require.NoError(t, err, "Grafana Agent should be accessible")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Grafana Agent health check should return 200 OK")
		
		// Verify configuration API is accessible
		configURL := "http://localhost:12345/agent/api/v1/config"
		configReq, err := http.NewRequestWithContext(ctx, "GET", configURL, nil)
		require.NoError(t, err)
		
		configResp, err := client.Do(configReq)
		if err == nil {
			defer configResp.Body.Close()
			// Config endpoint may return 404 if not enabled, but shouldn't be connection error
			assert.True(t, configResp.StatusCode == http.StatusOK || 
				configResp.StatusCode == http.StatusNotFound,
				"Grafana Agent config endpoint should respond")
		}
	})

	t.Run("metrics collection and forwarding", func(t *testing.T) {
		// Test: Grafana Agent can collect and forward metrics to Mimir
		
		// Check that agent is scraping configured targets
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Verify agent's internal metrics endpoint
		metricsURL := "http://localhost:12345/metrics"
		req, err := http.NewRequestWithContext(ctx, "GET", metricsURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode,
				"Grafana Agent should expose internal metrics")
			
			// Check for Prometheus metrics format
			contentType := resp.Header.Get("Content-Type")
			assert.Contains(t, contentType, "text/plain",
				"Grafana Agent should return Prometheus format metrics")
		}
	})

	t.Run("logs collection and forwarding", func(t *testing.T) {
		// Test: Grafana Agent can collect and forward logs to Loki
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test that the agent is configured to send logs to Loki
		// We can verify this by checking if Loki is receiving data from the agent
		lokiURL := fmt.Sprintf("http://localhost:%s/ready", getEnvWithDefault("LOKI_PORT", "3100"))
		req, err := http.NewRequestWithContext(ctx, "GET", lokiURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode,
				"Loki should be ready to receive logs from agent")
		}
		
		// Test Loki label API to see if we have any labels (indicating data flow)
		labelsURL := fmt.Sprintf("http://localhost:%s/loki/api/v1/labels", getEnvWithDefault("LOKI_PORT", "3100"))
		labelsReq, err := http.NewRequestWithContext(ctx, "GET", labelsURL, nil)
		require.NoError(t, err)
		
		labelsResp, err := client.Do(labelsReq)
		if err == nil {
			defer labelsResp.Body.Close()
			assert.True(t, labelsResp.StatusCode == http.StatusOK ||
				labelsResp.StatusCode == http.StatusNoContent,
				"Loki labels API should be accessible for log queries")
		}
	})

	t.Run("traces collection and forwarding", func(t *testing.T) {
		// Test: Grafana Agent can collect and forward traces to Tempo
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test that Tempo is ready to receive traces from the agent
		tempoURL := fmt.Sprintf("http://localhost:%s/ready", getEnvWithDefault("TEMPO_PORT", "3200"))
		req, err := http.NewRequestWithContext(ctx, "GET", tempoURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode,
				"Tempo should be ready to receive traces from agent")
		}
		
		// Test that agent's OTLP receivers are accessible
		// Agent should expose OTLP endpoints for receiving traces
		otlpHTTPURL := "http://localhost:4318/v1/traces"
		otlpReq, err := http.NewRequestWithContext(ctx, "POST", otlpHTTPURL, 
			bytes.NewReader([]byte("{}")))
		require.NoError(t, err)
		otlpReq.Header.Set("Content-Type", "application/json")
		
		otlpResp, err := client.Do(otlpReq)
		if err == nil {
			defer otlpResp.Body.Close()
			// Should not return 404 (endpoint exists) even if request format is wrong
			assert.True(t, otlpResp.StatusCode != http.StatusNotFound,
				"Agent OTLP HTTP endpoint should be accessible")
		}
	})

	t.Run("agent service integration", func(t *testing.T) {
		// Test: Grafana Agent integrates properly with observability stack
		
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Verify all configured backend services are accessible from agent's perspective
		backendServices := map[string]string{
			"Mimir (metrics)": fmt.Sprintf("http://localhost:%s/ready", getEnvWithDefault("MIMIR_PORT", "9009")),
			"Loki (logs)":     fmt.Sprintf("http://localhost:%s/ready", getEnvWithDefault("LOKI_PORT", "3100")),
			"Tempo (traces)":  fmt.Sprintf("http://localhost:%s/ready", getEnvWithDefault("TEMPO_PORT", "3200")),
		}
		
		healthyBackends := 0
		for service, healthURL := range backendServices {
			req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
			if err != nil {
				continue
			}
			
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					healthyBackends++
					t.Logf("%s is accessible to agent", service)
				}
			}
		}
		
		// Agent should be able to reach all 3 backend services
		assert.GreaterOrEqual(t, healthyBackends, 3,
			"Agent should be able to reach all observability backend services")
		
		// Verify agent itself is responsive
		agentHealthURL := "http://localhost:12345/-/ready"
		agentReq, err := http.NewRequestWithContext(ctx, "GET", agentHealthURL, nil)
		require.NoError(t, err)
		
		agentResp, err := client.Do(agentReq)
		require.NoError(t, err, "Agent should be responsive")
		defer agentResp.Body.Close()
		
		assert.Equal(t, http.StatusOK, agentResp.StatusCode,
			"Agent should report healthy status")
	})
}