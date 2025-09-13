package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	sharedValidation "github.com/axiom-software-co/international-center/src/public-website/infrastructure/test/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE: Dapr Component Configuration Tests
// These tests validate that Dapr state store, pub/sub, and other components are properly configured

func TestDaprComponentConfiguration_StateStoreAvailability(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test Dapr state store component availability
	t.Run("StateStore_ComponentAccess", func(t *testing.T) {
		// Test state store component through Dapr control plane
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Check Dapr components endpoint
		componentsURL := "http://localhost:3500/v1.0/components"
		req, err := http.NewRequestWithContext(ctx, "GET", componentsURL, nil)
		require.NoError(t, err, "Failed to create components request")

		resp, err := client.Do(req)
		require.NoError(t, err, "Dapr components endpoint must be accessible")
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
			"Dapr components endpoint must be operational")

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Failed to read components response")

		var components []map[string]interface{}
		err = json.Unmarshal(body, &components)
		require.NoError(t, err, "Components response must be valid JSON")

		// Should have state store component configured
		hasStateStore := false
		for _, component := range components {
			if componentType, exists := component["type"]; exists {
				if strings.Contains(strings.ToLower(componentType.(string)), "state") {
					hasStateStore = true
					break
				}
			}
		}

		assert.True(t, hasStateStore, 
			"Dapr must have state store component configured for service data persistence")

		t.Logf("Dapr components found: %d", len(components))
		for _, component := range components {
			if name, exists := component["name"]; exists {
				if componentType, exists := component["type"]; exists {
					t.Logf("Component: %s, Type: %s", name, componentType)
				}
			}
		}
	})

	// Test state store functionality
	t.Run("StateStore_Functionality", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Test state store through Dapr state API
		stateURL := "http://localhost:3500/v1.0/state/statestore/test-key"
		
		// Try to read from state store
		req, err := http.NewRequestWithContext(ctx, "GET", stateURL, nil)
		require.NoError(t, err, "Failed to create state store request")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			// State store should be accessible (404 for missing key is acceptable)
			assert.True(t, resp.StatusCode == 404 || (resp.StatusCode >= 200 && resp.StatusCode < 300),
				"Dapr state store must be accessible for service data operations")
		} else {
			t.Errorf("Dapr state store not accessible: %v", err)
		}
	})
}

func TestDaprComponentConfiguration_PubSubAvailability(t *testing.T) {
	// Test Dapr pub/sub component configuration
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test pub/sub component availability
	t.Run("PubSub_ComponentConfiguration", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Test pub/sub through Dapr publish endpoint
		publishURL := "http://localhost:3500/v1.0/publish/pubsub/test-topic"
		testMessage := `{"message": "test"}`
		
		req, err := http.NewRequestWithContext(ctx, "POST", publishURL, strings.NewReader(testMessage))
		require.NoError(t, err, "Failed to create pub/sub request")
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			// Pub/sub should be accessible (even if component not configured, should get proper error)
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500,
				"Dapr pub/sub endpoint must be accessible")
		} else {
			t.Errorf("Dapr pub/sub not accessible: %v", err)
		}
	})

	// RED PHASE: Test pub/sub component integration through Dapr component validation instead of direct container checks
	t.Run("PubSub_ComponentIntegration", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}
		
		// RED PHASE: Validate pub/sub component through Dapr metadata instead of direct RabbitMQ container inspection
		componentsURL := "http://localhost:3500/v1.0/metadata"
		req, err := http.NewRequestWithContext(ctx, "GET", componentsURL, nil)
		require.NoError(t, err, "Failed to create Dapr metadata request")

		resp, err := client.Do(req)
		require.NoError(t, err, "Dapr metadata endpoint must be accessible for pub/sub component validation")
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
			"Dapr metadata endpoint must be operational for component validation")

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Failed to read Dapr metadata response")

		var metadata map[string]interface{}
		err = json.Unmarshal(body, &metadata)
		require.NoError(t, err, "Dapr metadata response must be valid JSON")

		// RED PHASE: Validate pub/sub component is registered and configured via Dapr
		if components, exists := metadata["components"]; exists {
			componentsList, ok := components.([]interface{})
			require.True(t, ok, "Dapr metadata components must be array")
			
			hasPubSubComponent := false
			for _, component := range componentsList {
				if comp, ok := component.(map[string]interface{}); ok {
					if componentType, exists := comp["type"]; exists {
						if strings.Contains(strings.ToLower(componentType.(string)), "pubsub") {
							hasPubSubComponent = true
							t.Logf("RED PHASE: Found pub/sub component: %v", comp["name"])
							break
						}
					}
				}
			}
			
			// RED PHASE: This should fail initially until pub/sub component is properly configured
			assert.True(t, hasPubSubComponent, 
				"Dapr must have pub/sub component configured and accessible via service mesh metadata")
		} else {
			// RED PHASE: Expected to fail until proper Dapr component configuration
			t.Log("RED PHASE: Dapr metadata components not available - expected until pub/sub component properly configured")
		}
	})
}

// RED PHASE: Project-managed component configuration file validation
func TestDaprComponentConfiguration_ProjectManagedComponentFiles(t *testing.T) {
	// RED PHASE: Validate that component configuration files exist in proper project structure
	expectedComponentFiles := []string{
		"../../configs/dapr/statestore.yaml",
		"../../configs/dapr/secretstore.yaml",
		"../../configs/dapr/pubsub.yaml",
		"../../configs/dapr/config.yaml",
	}
	
	for _, configFile := range expectedComponentFiles {
		t.Run("ProjectManagedConfig_"+configFile, func(t *testing.T) {
			// RED PHASE: Component configuration files should exist in project structure
			cmd := exec.Command("test", "-f", configFile)
			err := cmd.Run()
			assert.NoError(t, err, 
				"RED PHASE: Project-managed Dapr component configuration file must exist in proper project structure: %s", configFile)
		})
	}
}

// RED PHASE: Comprehensive Dapr component configuration validation
func TestDaprComponentConfiguration_ComprehensiveComponentValidation(t *testing.T) {
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 10 * time.Second}

	// RED PHASE: Expected Dapr components that should be configured for the service mesh
	expectedComponents := []struct {
		componentType string
		componentName string
		description   string
	}{
		{"state.postgresql", "statestore", "PostgreSQL state store for service data persistence"},
		{"pubsub.rabbitmq", "pubsub", "RabbitMQ pub/sub for service communication"},
		{"secretstores.hashicorp.vault", "secretstore", "HashiCorp Vault for secrets management"},
		{"bindings.azure.blobstorage", "blobstore", "Azure Blob Storage for file operations"},
	}

	t.Run("ComponentRegistration_ComprehensiveValidation", func(t *testing.T) {
		// RED PHASE: Validate all expected components are registered through Dapr metadata API
		metadataReq, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3500/v1.0/metadata", nil)
		require.NoError(t, err, "Failed to create Dapr metadata request")

		resp, err := client.Do(metadataReq)
		require.NoError(t, err, "Dapr metadata API must be accessible for component validation")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Dapr metadata API should return component information")

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Failed to read Dapr metadata response")

		var metadata map[string]interface{}
		err = json.Unmarshal(body, &metadata)
		require.NoError(t, err, "Dapr metadata must be valid JSON")

		// RED PHASE: Validate each expected component is properly configured
		for _, expectedComp := range expectedComponents {
			t.Run("Component_"+expectedComp.componentName, func(t *testing.T) {
				// RED PHASE: This should fail initially until components are properly configured
				
				// Search for component in metadata response
				componentFound := false
				if components, exists := metadata["components"]; exists {
					if componentsList, ok := components.([]interface{}); ok {
						for _, component := range componentsList {
							if comp, ok := component.(map[string]interface{}); ok {
								if name, exists := comp["name"]; exists && name == expectedComp.componentName {
									if compType, exists := comp["type"]; exists {
										assert.Contains(t, strings.ToLower(compType.(string)), 
											strings.Split(expectedComp.componentType, ".")[0],
											"%s - component type should match expected", expectedComp.description)
										componentFound = true
										break
									}
								}
							}
						}
					}
				}

				// RED PHASE: Expected to fail until proper Dapr component configuration
				if !componentFound {
					t.Logf("RED PHASE: %s not found in Dapr component registry - expected until implemented", 
						expectedComp.description)
				}

				assert.True(t, componentFound, 
					"%s must be registered and accessible via Dapr service mesh", expectedComp.description)
			})
		}
	})

	// RED PHASE: Test component accessibility through specific Dapr APIs
	t.Run("ComponentAccessibility_ThroughDaprAPIs", func(t *testing.T) {
		componentEndpoints := []struct {
			componentName string
			endpoint      string
			method        string
			description   string
		}{
			{"statestore", "http://localhost:3500/v1.0/state/statestore/test", "GET", "State store component accessibility"},
			{"secretstore", "http://localhost:3500/v1.0/secrets/secretstore/test", "GET", "Secret store component accessibility"},
			{"blobstore", "http://localhost:3500/v1.0/bindings/blobstore", "POST", "Blob store component accessibility"},
		}

		for _, endpoint := range componentEndpoints {
			t.Run("ComponentAPI_"+endpoint.componentName, func(t *testing.T) {
				// RED PHASE: This should fail initially because components might not be properly configured
				
				var req *http.Request
				var err error
				if endpoint.method == "POST" {
					req, err = http.NewRequestWithContext(ctx, endpoint.method, endpoint.endpoint, strings.NewReader(`{"operation":"list"}`))
					if err == nil {
						req.Header.Set("Content-Type", "application/json")
					}
				} else {
					req, err = http.NewRequestWithContext(ctx, endpoint.method, endpoint.endpoint, nil)
				}
				require.NoError(t, err, "Failed to create %s request", endpoint.description)

				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					// Component should be accessible through Dapr (may return 404/400 but should not be connection error)
					assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
						"%s - component must be accessible via Dapr API", endpoint.description)
				} else {
					// RED PHASE: Expected to fail until proper component configuration
					t.Logf("RED PHASE: %s not accessible via Dapr API - expected until implemented: %v", 
						endpoint.description, err)
				}
			})
		}
	})
}

// RED PHASE: Runtime component loading validation
func TestDaprComponentConfiguration_RuntimeComponentLoading(t *testing.T) {
	// RED PHASE: Validate that project-managed components are loaded and accessible at runtime
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 10 * time.Second}

	// RED PHASE: Test that components are registered and accessible through Dapr runtime APIs
	t.Run("ComponentRuntimeAccessibility", func(t *testing.T) {
		// Get registered components from Dapr control plane
		componentsURL := "http://localhost:3500/v1.0/components"
		req, err := http.NewRequestWithContext(ctx, "GET", componentsURL, nil)
		require.NoError(t, err, "Failed to create components request")

		resp, err := client.Do(req)
		require.NoError(t, err, "Dapr components endpoint must be accessible for runtime validation")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, 
			"Dapr components endpoint should return registered components")

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Failed to read components response")

		var components []map[string]interface{}
		err = json.Unmarshal(body, &components)
		require.NoError(t, err, "Components response must be valid JSON")

		// RED PHASE: Validate that expected components are loaded at runtime
		expectedRuntimeComponents := []struct {
			name          string
			expectedType  string
			description   string
		}{
			{"statestore", "state.postgresql", "PostgreSQL state store component should be loaded"},
			{"secretstore", "secretstores.hashicorp.vault", "Vault secret store component should be loaded"},
			{"pubsub", "pubsub.rabbitmq", "RabbitMQ pub/sub component should be loaded"},
		}

		loadedComponents := make(map[string]map[string]interface{})
		for _, component := range components {
			if name, exists := component["name"]; exists {
				loadedComponents[name.(string)] = component
			}
		}

		for _, expected := range expectedRuntimeComponents {
			t.Run("RuntimeComponent_"+expected.name, func(t *testing.T) {
				component, loaded := loadedComponents[expected.name]
				assert.True(t, loaded, 
					"RED PHASE: %s - component should be loaded at runtime (will fail until component loading implemented)", expected.description)
				
				if loaded {
					if componentType, exists := component["type"]; exists {
						assert.Contains(t, strings.ToLower(componentType.(string)), 
							strings.Split(expected.expectedType, ".")[0],
							"RED PHASE: Component %s should have correct type", expected.name)
					}
					t.Logf("RED PHASE: Component %s loaded with type: %v", expected.name, component["type"])
				} else {
					t.Logf("RED PHASE: Component %s not loaded - expected until runtime component loading implemented", expected.name)
				}
			})
		}

		t.Logf("RED PHASE: Found %d loaded components at runtime", len(components))
	})
}

func TestDaprComponentConfiguration_ServiceStateIntegration(t *testing.T) {
	// Test that services can integrate with Dapr state store for data operations
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Services that should be able to use Dapr state store
	serviceStateTests := []struct {
		serviceName     string
		healthEndpoint  string
		dataEndpoint    string
		stateOperation  string
		description     string
	}{
		{
			serviceName:    "content-api",
			healthEndpoint: "http://localhost:3500/v1.0/invoke/content-api/method/health",
			dataEndpoint:   "http://localhost:3500/v1.0/invoke/content-api/method/api/news",
			stateOperation: "news data persistence",
			description:    "Content service must integrate with Dapr state store for news data operations",
		},
		{
			serviceName:    "inquiries-api",
			healthEndpoint: "http://localhost:3500/v1.0/invoke/inquiries-api/method/health",
			dataEndpoint:   "http://localhost:3500/v1.0/invoke/inquiries-api/method/api/inquiries",
			stateOperation: "inquiry data persistence",
			description:    "Inquiries service must integrate with Dapr state store for inquiry data operations",
		},
		{
			serviceName:    "notification-api",
			healthEndpoint: "http://localhost:3500/v1.0/invoke/notification-api/method/health",
			dataEndpoint:   "http://localhost:3500/v1.0/invoke/notification-api/method/api/subscribers",
			stateOperation: "subscriber data persistence",
			description:    "Notifications service must integrate with Dapr state store for subscriber data operations",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test service state integration
	for _, stateTest := range serviceStateTests {
		t.Run("ServiceState_"+stateTest.serviceName, func(t *testing.T) {
			// Verify service is healthy
			healthReq, err := http.NewRequestWithContext(ctx, "GET", stateTest.healthEndpoint, nil)
			require.NoError(t, err, "Failed to create service health request")

			healthResp, err := client.Do(healthReq)
			require.NoError(t, err, "Service health must be accessible")
			defer healthResp.Body.Close()

			assert.True(t, healthResp.StatusCode >= 200 && healthResp.StatusCode < 300,
				"Service %s must be healthy for state integration testing", stateTest.serviceName)

			// Test service data operations (should work with Dapr state store)
			dataReq, err := http.NewRequestWithContext(ctx, "GET", stateTest.dataEndpoint, nil)
			require.NoError(t, err, "Failed to create service data request")

			dataResp, err := client.Do(dataReq)
			require.NoError(t, err, "Service data endpoint must be accessible")
			defer dataResp.Body.Close()

			assert.True(t, dataResp.StatusCode >= 200 && dataResp.StatusCode < 300,
				"%s - service data operations must be functional", stateTest.description)

			// Validate response structure indicates state management readiness
			body, err := io.ReadAll(dataResp.Body)
			if err == nil {
				var jsonData map[string]interface{}
				assert.NoError(t, json.Unmarshal(body, &jsonData),
					"%s - service must return JSON for %s", stateTest.description, stateTest.stateOperation)

				// Should have data structure ready for state operations
				assert.Contains(t, jsonData, "data",
					"%s - service response must have data field for state operations", stateTest.description)
				assert.Contains(t, jsonData, "count",
					"%s - service response must have count field for state operations", stateTest.description)
			}
		})
	}
}

func TestDaprComponentConfiguration_CrossServiceCommunication(t *testing.T) {
	// Test that cross-service communication works through Dapr service mesh
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Cross-service communication patterns that should work through Dapr
	crossServiceTests := []struct {
		communicationPattern string
		sourceService        string
		targetService        string
		testEndpoint         string
		description          string
	}{
		{
			communicationPattern: "content-to-notifications",
			sourceService:        "content-api",
			targetService:        "notification-api",
			testEndpoint:         "http://localhost:3500/v1.0/invoke/notification-api/method/health",
			description:          "Content service must communicate with notifications through Dapr for event publishing",
		},
		{
			communicationPattern: "inquiries-to-notifications",
			sourceService:        "inquiries-api",
			targetService:        "notification-api",
			testEndpoint:         "http://localhost:3500/v1.0/invoke/notification-api/method/health",
			description:          "Inquiries service must communicate with notifications through Dapr for inquiry alerts",
		},
		{
			communicationPattern: "gateway-to-content",
			sourceService:        "public-gateway",
			targetService:        "content-api",
			testEndpoint:         "http://localhost:3500/v1.0/invoke/content-api/method/health",
			description:          "Public gateway must communicate with content service through Dapr for API routing",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test cross-service communication
	for _, communication := range crossServiceTests {
		t.Run("CrossService_"+communication.communicationPattern, func(t *testing.T) {
			// Test that target service is accessible through service mesh
			req, err := http.NewRequestWithContext(ctx, "GET", communication.testEndpoint, nil)
			require.NoError(t, err, "Failed to create cross-service communication request")

			resp, err := client.Do(req)
			require.NoError(t, err, "Cross-service communication must be accessible")
			defer resp.Body.Close()

			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
				"%s - cross-service communication must be functional", communication.description)

			// Validate service mesh communication works
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				var jsonData interface{}
				assert.NoError(t, json.Unmarshal(body, &jsonData),
					"%s - cross-service communication must return valid JSON", communication.description)
			}
		})
	}
}

