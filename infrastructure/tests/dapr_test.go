package tests

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDaprComponentHealth(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("redis state store connectivity", func(t *testing.T) {
		// Test: Redis state store for Dapr components is accessible
		redisPort := requireEnv(t, "REDIS_PORT")
		
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

	t.Run("component configuration validation", func(t *testing.T) {
		// Test: Dapr component configuration exists and is valid
		// This tests that local development has proper component configuration
		
		// Check if Dapr configuration exists
		daprConfigPath := "infrastructure/tests/dapr/config.yaml"
		_, err := os.Stat(daprConfigPath)
		if err != nil {
			t.Logf("Dapr configuration file not found at %s: %v", daprConfigPath, err)
		}
		assert.NoError(t, err, "Dapr configuration should exist for component definition")
		
		// Verify configuration contains required settings for components
		if err == nil {
			content, err := os.ReadFile(daprConfigPath)
			require.NoError(t, err, "Should be able to read Dapr configuration")
			
			configContent := string(content)
			assert.Contains(t, configContent, "apiVersion", "Dapr config should have apiVersion")
			assert.Contains(t, configContent, "kind: Configuration", "Dapr config should be Configuration kind")
			assert.Contains(t, configContent, "state", "Dapr config should include state component access")
		}
	})
}

func TestDaprComponentUsage(t *testing.T) {
	// Integration test - requires full podman compose environment

	t.Run("state store component accessibility", func(t *testing.T) {
		// Test: Redis state store component is accessible for Dapr state operations
		// This validates the primary component usage pattern in local development
		
		redisPort := requireEnv(t, "REDIS_PORT")
		
		// Test Redis connectivity for state store operations
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%s", redisPort), 5*time.Second)
		if err != nil {
			t.Logf("Redis state store not accessible on port %s: %v", redisPort, err)
		}
		require.NoError(t, err, "Redis state store should be accessible for Dapr components")
		
		if conn != nil {
			conn.Close()
			t.Logf("Redis state store successfully validated on port %s", redisPort)
		}
	})

	t.Run("configuration validation for components", func(t *testing.T) {
		// Test: Dapr configuration supports required component features
		// This validates component configuration for actual usage patterns
		
		daprConfigPath := "infrastructure/tests/dapr/config.yaml"
		content, err := os.ReadFile(daprConfigPath)
		if err != nil {
			t.Logf("Cannot read Dapr configuration: %v", err)
			t.Skip("Skipping component configuration validation")
		}
		
		configContent := string(content)
		
		// Validate essential component access is enabled
		assert.Contains(t, configContent, "StateManagement", "State management should be enabled for Redis component")
		assert.Contains(t, configContent, "ServiceInvocation", "Service invocation should be enabled for API communication")
		assert.Contains(t, configContent, "enabled: true", "Components should be explicitly enabled")
	})
}

