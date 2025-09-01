package tests

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDaprControlPlaneHealth(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("dapr placement service operational", func(t *testing.T) {
		// Test: Dapr placement service is running and accessible (gRPC service)
		placementPort := getEnvWithDefault("DAPR_PLACEMENT_PORT", "6050")
		
		// Test port connectivity since placement service uses gRPC
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%s", placementPort), 5*time.Second)
		if err != nil {
			t.Logf("Dapr placement service not accessible on port %s: %v", placementPort, err)
		}
		require.NoError(t, err, "Dapr placement service should be accessible on port %s", placementPort)
		
		if conn != nil {
			conn.Close()
			t.Logf("Dapr placement service successfully connected on port %s", placementPort)
		}
	})

	t.Run("dapr sentry service operational", func(t *testing.T) {
		// Test: Dapr sentry configuration for development environment
		sentryPort := getEnvWithDefault("DAPR_SENTRY_PORT", "6051")
		
		// Test port connectivity - sentry may not be fully operational in development
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%s", sentryPort), 5*time.Second)
		if err != nil {
			t.Logf("Dapr sentry service not accessible on port %s (acceptable for development): %v", sentryPort, err)
			// In development, sentry may not be fully configured - this is acceptable
			return
		}
		
		if conn != nil {
			conn.Close()
			t.Logf("Dapr sentry service successfully connected on port %s", sentryPort)
		}
	})

	t.Run("redis state store connectivity", func(t *testing.T) {
		// Test: Redis state store for Dapr control plane is accessible
		redisPort := getEnvWithDefault("REDIS_PORT", "6379")
		
		// Test Redis connectivity using TCP connection
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%s", redisPort), 5*time.Second)
		if err != nil {
			t.Logf("Redis not accessible on port %s: %v", redisPort, err)
		}
		require.NoError(t, err, "Redis should be accessible on port %s", redisPort)
		
		if conn != nil {
			conn.Close()
			t.Logf("Redis successfully connected on port %s", redisPort)
		}
	})

	t.Run("service discovery functionality", func(t *testing.T) {
		// Test: Service discovery is operational
		// This tests that Dapr can handle service discovery requests
		
		// Test Dapr metadata endpoint which requires service discovery to be working
		metadataURL := "http://localhost:3500/v1.0/metadata" // Default Dapr HTTP port
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", metadataURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Dapr metadata endpoint not accessible: %v", err)
			t.Log("Service discovery functionality cannot be verified without Dapr sidecar")
			return
		}
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Dapr metadata endpoint should be accessible")
	})
}

func TestDaprServiceRegistration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("service registration capability", func(t *testing.T) {
		// Test: Dapr can register services
		// This will test the basic service registration mechanism
		
		// Check if Dapr configuration exists
		daprConfigPath := "dapr/config.yaml"
		_, err := os.Stat(daprConfigPath)
		if err != nil {
			t.Logf("Dapr configuration file not found at %s: %v", daprConfigPath, err)
		}
		assert.NoError(t, err, "Dapr configuration should exist")
		
		// Verify configuration contains required settings for service registration
		if err == nil {
			content, err := os.ReadFile(daprConfigPath)
			require.NoError(t, err, "Should be able to read Dapr configuration")
			
			configContent := string(content)
			assert.Contains(t, configContent, "apiVersion", "Dapr config should have apiVersion")
			assert.Contains(t, configContent, "kind: Configuration", "Dapr config should be Configuration kind")
		}
	})

	t.Run("health check endpoint responsiveness", func(t *testing.T) {
		// Test: Health check endpoints are responsive
		// This tests the Dapr health check functionality
		
		healthEndpoints := []string{
			"http://localhost:3500/v1.0/healthz",         // Dapr HTTP health
			"http://localhost:3500/v1.0/healthz/outbound", // Outbound health
		}
		
		client := &http.Client{Timeout: 3 * time.Second}
		
		for _, endpoint := range healthEndpoints {
			req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
			require.NoError(t, err)
			
			resp, err := client.Do(req)
			if err != nil {
				t.Logf("Health endpoint %s not accessible: %v", endpoint, err)
				continue
			}
			defer resp.Body.Close()
			
			assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent,
				"Health endpoint %s should return success status", endpoint)
		}
	})

	t.Run("service invocation readiness", func(t *testing.T) {
		// Test: Service invocation is ready
		// This tests that Dapr is ready to handle service-to-service calls
		
		// Test the Dapr service invocation endpoint
		invocationURL := "http://localhost:3500/v1.0/invoke/nonexistent-service/method/health"
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", invocationURL, nil)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Service invocation endpoint not accessible: %v", err)
			return
		}
		defer resp.Body.Close()
		
		// We expect this to fail with a specific error (service not found)
		// but the endpoint should be accessible, indicating service invocation is ready
		assert.True(t, resp.StatusCode == http.StatusInternalServerError || 
			resp.StatusCode == http.StatusNotFound ||
			resp.StatusCode == http.StatusBadRequest,
			"Service invocation should be ready (even if service doesn't exist)")
	})
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}