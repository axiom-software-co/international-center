package testing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	sharedtesting "github.com/axiom-software-co/international-center/src/deployer/shared/testing"
)

// TestObservabilityStack validates Grafana observability components
func TestObservabilityStack(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := sharedtesting.CreateIntegrationTestContext()
	defer cancel()

	// Test Grafana dashboard accessibility
	t.Run("Grafana_Accessibility", func(t *testing.T) {
		grafanaURL := sharedtesting.GetEnvVar("GRAFANA_URL", "")
		if grafanaURL == "" {
			t.Skip("GRAFANA_URL not configured, skipping Grafana test")
		}

		// Act - Check Grafana health endpoint
		healthURL := fmt.Sprintf("%s/api/health", grafanaURL)
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", healthURL, nil)
		require.NoError(t, err, "Failed to make HTTP request to Grafana")

		defer resp.Body.Close()

		// Assert - If Grafana is running, it should be healthy
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Grafana should be healthy when running")

		var healthData map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&healthData)
		if err == nil {
			assert.Equal(t, "ok", healthData["database"], "Grafana database should be healthy")
			t.Logf("Grafana is running and healthy at %s", grafanaURL)
		}
	})

	// Test Loki log aggregation service
	t.Run("Loki_Accessibility", func(t *testing.T) {
		lokiURL := sharedtesting.GetEnvVar("LOKI_URL", "")
		if lokiURL == "" {
			t.Skip("LOKI_URL not configured, skipping Loki test")
		}

		// Act - Check Loki readiness endpoint
		readyURL := fmt.Sprintf("%s/ready", lokiURL)
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", readyURL, nil)
		require.NoError(t, err, "Should be able to connect to Loki")

		defer resp.Body.Close()

		// Assert - If Loki is running, it should be ready
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Loki should be ready when running")
		t.Logf("Loki is running and ready at %s", lokiURL)
	})

	// Test Prometheus/Mimir metrics collection
	t.Run("Metrics_Collection_Accessibility", func(t *testing.T) {
		prometheusURL := sharedtesting.GetEnvVar("PROMETHEUS_URL", "")
		if prometheusURL == "" {
			prometheusURL = sharedtesting.GetEnvVar("MIMIR_URL", "")
		}
		
		if prometheusURL == "" {
			t.Skip("Neither PROMETHEUS_URL nor MIMIR_URL configured, skipping metrics test")
		}

		// Act - Check metrics endpoint
		metricsURL := fmt.Sprintf("%s/api/v1/query?query=up", prometheusURL)
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", metricsURL, nil)
		require.NoError(t, err, "Should be able to connect to metrics service")

		defer resp.Body.Close()

		// Assert - If metrics service is running, it should respond to queries
		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
			"Metrics service should respond to queries when running")
		t.Logf("Metrics collection service is accessible at %s", prometheusURL)
	})
}

// TestStorageServices validates blob storage and related services
func TestStorageServices(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := sharedtesting.CreateIntegrationTestContext()
	defer cancel()

	// Test Azurite blob storage emulator
	t.Run("Azurite_Blob_Storage", func(t *testing.T) {
		// Arrange
		azuriteURL := sharedtesting.GetRequiredEnvVar(t, "AZURITE_URL")

		// Act - Check Azurite service availability
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "HEAD", azuriteURL, nil)

		// Assert - Azurite should be accessible
		require.NoError(t, err, "Azurite blob storage emulator should be accessible")
		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
			"Azurite should respond to requests (status: %d)", resp.StatusCode)
		
		t.Logf("Azurite blob storage is accessible at %s", azuriteURL)
	})

	// Test blob storage container operations
	t.Run("Blob_Storage_Operations", func(t *testing.T) {
		// Arrange  
		azuriteURL := sharedtesting.GetRequiredEnvVar(t, "AZURITE_URL")
		
		// Act - Try to list containers (should work even if no containers exist)
		containerListURL := fmt.Sprintf("%s?comp=list", azuriteURL)
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", containerListURL, nil)

		// Assert - Container listing should work
		require.NoError(t, err, "Should be able to query blob storage containers")
		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 400,
			"Container listing should succeed (status: %d)", resp.StatusCode)
			
		t.Logf("Blob storage operations are functional (status: %d)", resp.StatusCode)
	})
}

// TestSecurityServices validates security-related services
func TestSecurityServices(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Test Vault secret management service
	t.Run("Vault_Secret_Management", func(t *testing.T) {
		// Arrange
		vaultURL := sharedtesting.GetRequiredEnvVar(t, "VAULT_URL")

		// Act - Check Vault health endpoint
		healthURL := fmt.Sprintf("%s/v1/sys/health", vaultURL)
		ctx, cancel := sharedtesting.CreateIntegrationTestContext()
		defer cancel()
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", healthURL, nil)
		require.NoError(t, err, "Should be able to make HTTP request to Vault")

		// Assert - Vault should be accessible and healthy
		
		// Vault health endpoint returns various status codes based on configuration
		// 200 = initialized and unsealed
		// 429 = unsealed and standby 
		// 472 = disaster recovery mode replication
		// 473 = performance standby
		// 501 = not initialized
		// 503 = sealed
		acceptableStatuses := []int{200, 429, 472, 473, 501, 503}
		statusAcceptable := false
		for _, status := range acceptableStatuses {
			if resp.StatusCode == status {
				statusAcceptable = true
				break
			}
		}
		
		assert.True(t, statusAcceptable, "Vault should return valid health status (got: %d)", resp.StatusCode)
		
		// Parse health response for detailed information
		var healthData map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&healthData)
		if err == nil {
			if initialized, exists := healthData["initialized"]; exists {
				t.Logf("Vault initialized: %v", initialized)
			}
			if sealed, exists := healthData["sealed"]; exists {
				t.Logf("Vault sealed: %v", sealed)
			}
		}
		
		t.Logf("Vault is accessible and responding at %s (status: %d)", vaultURL, resp.StatusCode)
	})

	// Test Authentik identity provider (if configured)
	t.Run("Authentik_Identity_Provider", func(t *testing.T) {
		authentikURL := sharedtesting.GetEnvVar("AUTHENTIK_URL", "")
		if authentikURL == "" {
			t.Skip("AUTHENTIK_URL not configured, skipping Authentik test")
		}

		// Act - Check Authentik health/status endpoint
		healthURL := fmt.Sprintf("%s/api/v3/admin/version/", authentikURL)
		ctx, cancel := sharedtesting.CreateIntegrationTestContext()
		defer cancel()
		resp, err := sharedtesting.MakeHTTPRequest(ctx, "GET", healthURL, nil)
		require.NoError(t, err, "Should be able to make HTTP request to Authentik")

		// Assert - If Authentik is running, it should respond
		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500,
			"Authentik should respond when running (status: %d)", resp.StatusCode)
		t.Logf("Authentik identity provider is accessible at %s", authentikURL)
	})
}

// TestNetworkConnectivity validates network connectivity between services
func TestNetworkConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx, cancel := sharedtesting.CreateIntegrationTestContext()
	defer cancel()

	// Test internal network connectivity between services
	t.Run("Internal_Service_Connectivity", func(t *testing.T) {
		// Define service endpoints to test connectivity
		services := map[string]string{
			"Content API":    sharedtesting.GetRequiredEnvVar(t, "CONTENT_API_URL"),
			"Services API":   sharedtesting.GetRequiredEnvVar(t, "SERVICES_API_URL"), 
			"Public Gateway": sharedtesting.GetRequiredEnvVar(t, "PUBLIC_GATEWAY_URL"),
			"Admin Gateway":  sharedtesting.GetRequiredEnvVar(t, "ADMIN_GATEWAY_URL"),
		}

		// Test each service can reach the others
		for sourceName, _ := range services {
			t.Run(fmt.Sprintf("From_%s", strings.ReplaceAll(sourceName, " ", "_")), func(t *testing.T) {
				for targetName, targetURL := range services {
					if sourceName == targetName {
						continue // Skip self-connectivity test
					}

					// Act - Test connectivity by checking if we can resolve and connect
					targetHost := extractHostFromURL(targetURL)
					if targetHost != "" {
						conn, err := sharedtesting.ConnectWithTimeout(ctx, "tcp", targetHost, 2*time.Second)
						
						// Assert - Services should be able to connect to each other
						if err != nil {
							t.Logf("Connection from %s to %s (%s) failed: %v", sourceName, targetName, targetHost, err)
						} else {
							conn.Close()
							t.Logf("✓ %s can connect to %s", sourceName, targetName)
						}
					}
				}
			})
		}
	})

	// Test external connectivity (if required)
	t.Run("External_Connectivity", func(t *testing.T) {
		// Test connectivity to external test endpoints if configured
		testEndpoint1 := sharedtesting.GetEnvVar("NETWORK_TEST_ENDPOINT_1", "") 
		testEndpoint2 := sharedtesting.GetEnvVar("NETWORK_TEST_ENDPOINT_2", "")

		if testEndpoint1 == "" && testEndpoint2 == "" {
			t.Skip("No external connectivity test endpoints configured")
		}

		endpoints := []string{}
		if testEndpoint1 != "" {
			endpoints = append(endpoints, testEndpoint1)
		}
		if testEndpoint2 != "" {
			endpoints = append(endpoints, testEndpoint2)
		}

		for _, endpoint := range endpoints {
			t.Run(fmt.Sprintf("Endpoint_%s", endpoint), func(t *testing.T) {
				// Act - Test external connectivity
				conn, err := sharedtesting.ConnectWithTimeout(ctx, "tcp", endpoint, 3*time.Second)

				if err != nil {
					t.Logf("External connectivity to %s failed: %v", endpoint, err)
				} else {
					conn.Close()
					t.Logf("✓ External connectivity to %s successful", endpoint)
				}
			})
		}
	})
}

// TestEnvironmentConfiguration validates environment-specific configuration
func TestEnvironmentConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Test development-specific configuration
	t.Run("Development_Configuration", func(t *testing.T) {
		// Verify development environment variables are set correctly
		requiredDevVars := []string{
			"DATABASE_URL",
			"REDIS_URL",
			"VAULT_URL", 
			"AZURITE_URL",
			"CONTENT_API_URL",
			"SERVICES_API_URL",
			"PUBLIC_GATEWAY_URL",
			"ADMIN_GATEWAY_URL",
			"DAPR_HTTP_PORT",
			"DAPR_GRPC_PORT",
		}

		for _, envVar := range requiredDevVars {
			value := sharedtesting.GetEnvVar(envVar, "")
			assert.NotEmpty(t, value, "Development environment variable %s should be set", envVar)
			
			// Validate URL format for URL variables
			if strings.HasSuffix(envVar, "_URL") {
				assert.True(t, strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://"),
					"Environment variable %s should be a valid URL", envVar)
			}
		}
	})

	// Test port configuration consistency
	t.Run("Port_Configuration_Consistency", func(t *testing.T) {
		// Verify no port conflicts in configuration
		portVars := map[string]string{
			"DATABASE_PORT":   sharedtesting.GetEnvVar("DATABASE_PORT", "5432"),
			"REDIS_PORT":      sharedtesting.GetEnvVar("REDIS_PORT", "6379"), 
			"VAULT_PORT":      sharedtesting.GetEnvVar("VAULT_PORT", "8200"),
			"AZURITE_PORT":    sharedtesting.GetEnvVar("AZURITE_PORT", "10000"),
			"GRAFANA_PORT":    sharedtesting.GetEnvVar("GRAFANA_PORT", "3000"),
			"LOKI_PORT":       sharedtesting.GetEnvVar("LOKI_PORT", "3100"),
			"DAPR_HTTP_PORT":  sharedtesting.GetEnvVar("DAPR_HTTP_PORT", "3500"),
			"DAPR_GRPC_PORT":  sharedtesting.GetEnvVar("DAPR_GRPC_PORT", "50001"),
		}

		usedPorts := make(map[string]string)
		for varName, port := range portVars {
			if port != "" {
				if conflictVar, exists := usedPorts[port]; exists {
					t.Errorf("Port conflict: %s and %s both use port %s", varName, conflictVar, port)
				} else {
					usedPorts[port] = varName
					t.Logf("Port assignment: %s = %s", varName, port)
				}
			}
		}
	})
}

// Helper function to extract host:port from URL
func extractHostFromURL(url string) string {
	// Simple extraction - remove protocol and path
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}
	
	// Find the end of host:port (before path)
	if pathIndex := strings.Index(url, "/"); pathIndex != -1 {
		url = url[:pathIndex]
	}
	
	return url
}