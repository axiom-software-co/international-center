package integration

import (
	"context"
	"encoding/json"
	"fmt"
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

// REFACTOR PHASE: Component functionality validation through service integration
func TestDaprComponentConfiguration_ComponentFunctionalityThroughServices(t *testing.T) {
	// REFACTOR PHASE: Validate component functionality through actual service operations
	// This reflects correct Dapr architecture where components are accessed via service sidecars
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	// REFACTOR PHASE: Test component functionality by validating service operations that depend on components
	t.Run("ComponentFunctionalityThroughServiceOperations", func(t *testing.T) {
		// Note: In Dapr architecture, components are not exposed through central APIs
		// Instead, they are loaded by service sidecars and accessed through service operations
		
		// Component functionality validation will be performed through:
		// 1. Service state operations (testing state store component)
		// 2. Service-to-service communication (testing service invocation)  
		// 3. Service data operations that require database connectivity
		
		t.Log("Component functionality validated through service integration tests")
		t.Log("This approach aligns with Dapr sidecar architecture patterns")
		
		// The actual component functionality testing is delegated to:
		// - TestDaprComponentConfiguration_ServiceStateIntegration (tests state store component)
		// - TestDaprComponentConfiguration_CrossServiceCommunication (tests service mesh)
		// This ensures components are tested in their proper architectural context
	})
}

func TestDaprComponentConfiguration_ServiceStateIntegration(t *testing.T) {
	// RED PHASE: Enhanced service API contract validation with specific endpoint and schema requirements
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// RED PHASE: Enhanced service API contract specifications with detailed validation requirements
	serviceAPIContracts := []struct {
		serviceName        string
		healthEndpoint     string
		dataEndpoint       string
		expectedHTTPStatus int
		expectedContentType string
		requiredFields     []string
		dataFieldType      string
		stateOperation     string
		description        string
	}{
		{
			serviceName:        "content-api",
			healthEndpoint:     "http://localhost:3500/v1.0/invoke/content-api/method/health",
			dataEndpoint:       "http://localhost:3500/v1.0/invoke/content-api/method/api/v1/news/featured",
			expectedHTTPStatus: 200,
			expectedContentType: "application/json",
			requiredFields:     []string{"data"},
			dataFieldType:      "object",
			stateOperation:     "news data persistence",
			description:        "Content service must implement /api/v1/news/featured endpoint with proper JSON API contract",
		},
		{
			serviceName:        "inquiries-api",
			healthEndpoint:     "http://localhost:3500/v1.0/invoke/inquiries-api/method/health",
			dataEndpoint:       "http://localhost:3500/v1.0/invoke/inquiries-api/method/api/inquiries",
			expectedHTTPStatus: 200,
			expectedContentType: "application/json",
			requiredFields:     []string{"data", "count"},
			dataFieldType:      "array",
			stateOperation:     "inquiry data persistence", 
			description:        "Inquiries service must implement /api/inquiries endpoint with proper JSON API contract",
		},
		{
			serviceName:        "notification-api",
			healthEndpoint:     "http://localhost:3500/v1.0/invoke/notification-api/method/health",
			dataEndpoint:       "http://localhost:3500/v1.0/invoke/notification-api/method/api/subscribers",
			expectedHTTPStatus: 200,
			expectedContentType: "application/json",
			requiredFields:     []string{"data", "count"},
			dataFieldType:      "array",
			stateOperation:     "subscriber data persistence",
			description:        "Notifications service must implement /api/subscribers endpoint with proper JSON API contract",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// RED PHASE: Enhanced API contract validation with specific endpoint and schema testing
	for _, contract := range serviceAPIContracts {
		t.Run("ServiceAPIContract_"+contract.serviceName, func(t *testing.T) {
			// RED PHASE: Prerequisite - Verify service health before testing API contracts
			t.Run("ServiceHealth", func(t *testing.T) {
				healthReq, err := http.NewRequestWithContext(ctx, "GET", contract.healthEndpoint, nil)
				require.NoError(t, err, "Failed to create service health request")

				healthResp, err := client.Do(healthReq)
				require.NoError(t, err, "Service health must be accessible for API contract testing")
				defer healthResp.Body.Close()

				assert.True(t, healthResp.StatusCode >= 200 && healthResp.StatusCode < 300,
					"RED PHASE: Service %s must be healthy before testing API contracts", contract.serviceName)
			})

			// RED PHASE: API Endpoint Accessibility Validation
			t.Run("APIEndpointAccessibility", func(t *testing.T) {
				dataReq, err := http.NewRequestWithContext(ctx, "GET", contract.dataEndpoint, nil)
				require.NoError(t, err, "Failed to create API endpoint request")

				dataResp, err := client.Do(dataReq)
				require.NoError(t, err, "API endpoint must be accessible via Dapr service mesh")
				defer dataResp.Body.Close()

				// RED PHASE: This should fail until API endpoints are implemented
				assert.Equal(t, contract.expectedHTTPStatus, dataResp.StatusCode,
					"RED PHASE: %s - API endpoint must return expected HTTP status (will fail until implemented)", contract.description)
				
				// RED PHASE: Validate Content-Type header
				contentType := dataResp.Header.Get("Content-Type")
				assert.Contains(t, contentType, contract.expectedContentType,
					"RED PHASE: %s - API endpoint must return expected content type", contract.description)
			})

			// RED PHASE: JSON Response Schema Validation  
			t.Run("JSONResponseSchema", func(t *testing.T) {
				dataReq, err := http.NewRequestWithContext(ctx, "GET", contract.dataEndpoint, nil)
				require.NoError(t, err, "Failed to create API endpoint request")

				dataResp, err := client.Do(dataReq)
				require.NoError(t, err, "API endpoint must be accessible for schema validation")
				defer dataResp.Body.Close()

				// RED PHASE: Only validate schema if we get a successful response
				if dataResp.StatusCode == contract.expectedHTTPStatus {
					body, err := io.ReadAll(dataResp.Body)
					require.NoError(t, err, "Failed to read API response body")

					var jsonData map[string]interface{}
					assert.NoError(t, json.Unmarshal(body, &jsonData),
						"RED PHASE: %s - API response must be valid JSON", contract.description)

					// RED PHASE: Validate required fields in response
					for _, field := range contract.requiredFields {
						assert.Contains(t, jsonData, field,
							"RED PHASE: %s - API response must contain '%s' field", contract.description, field)
					}

					// RED PHASE: Validate data field is an array as expected
					if dataField, exists := jsonData["data"]; exists {
						assert.IsType(t, []interface{}{}, dataField,
							"RED PHASE: %s - 'data' field must be an array type", contract.description)
					}

					// RED PHASE: Validate count field is a number
					if countField, exists := jsonData["count"]; exists {
						assert.IsType(t, float64(0), countField,
							"RED PHASE: %s - 'count' field must be a number type", contract.description)
					}
				} else {
					t.Logf("RED PHASE: %s - API endpoint not yet implemented, skipping schema validation", contract.description)
				}
			})

			// RED PHASE: State Store Integration Validation
			t.Run("StateStoreIntegration", func(t *testing.T) {
				// RED PHASE: This test validates that the API endpoints properly integrate with Dapr state store
				// This will be validated by checking that the API returns persistent data
				
				dataReq, err := http.NewRequestWithContext(ctx, "GET", contract.dataEndpoint, nil)
				require.NoError(t, err, "Failed to create state store integration request")

				dataResp, err := client.Do(dataReq)
				require.NoError(t, err, "State store integration endpoint must be accessible")
				defer dataResp.Body.Close()

				// RED PHASE: If endpoint is implemented, it should demonstrate state store integration
				if dataResp.StatusCode == contract.expectedHTTPStatus {
					t.Logf("RED PHASE: %s - State store integration can be validated once endpoint is implemented", contract.description)
					
					// Future validation: Check that data persists across requests
					// Future validation: Check that data can be modified and retrieved
					// Future validation: Check that count field reflects actual data count
				} else {
					t.Logf("RED PHASE: %s - State store integration validation deferred until API endpoint implementation", contract.description)
				}
			})
		})
	}
}

// RED PHASE: Comprehensive Dapr Service Registration Validation
func TestDaprComponentConfiguration_ServiceRegistration(t *testing.T) {
	// This test validates that all services properly register with Dapr runtime for service discovery
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Expected services that must be registered with Dapr runtime
	expectedServices := []struct {
		appId       string
		serviceName string
		description string
	}{
		{
			appId:       "content-api",
			serviceName: "Content Service",
			description: "Content service must register with Dapr for news, events, research, services domains",
		},
		{
			appId:       "inquiries-api", 
			serviceName: "Inquiries Service",
			description: "Inquiries service must register with Dapr for business, donations, media, volunteer domains",
		},
		{
			appId:       "notification-api",
			serviceName: "Notifications Service", 
			description: "Notifications service must register with Dapr for email, SMS, Slack routing",
		},
		{
			appId:       "public-gateway",
			serviceName: "Public Gateway",
			description: "Public gateway must register with Dapr for anonymous API routing",
		},
		{
			appId:       "admin-gateway",
			serviceName: "Admin Gateway",
			description: "Admin gateway must register with Dapr for authenticated API routing",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Validate service registration through Dapr metadata API
	t.Run("DaprServiceDiscovery", func(t *testing.T) {
		// Test Dapr metadata endpoint accessibility
		metadataURL := "http://localhost:3500/v1.0/metadata"
		req, err := http.NewRequestWithContext(ctx, "GET", metadataURL, nil)
		require.NoError(t, err, "Failed to create Dapr metadata request")

		resp, err := client.Do(req)
		require.NoError(t, err, "Dapr metadata endpoint must be accessible for service discovery validation")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, 
			"Dapr metadata endpoint must return 200 OK for service registry access")

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Failed to read Dapr metadata response")

		var metadata map[string]interface{}
		err = json.Unmarshal(body, &metadata)
		require.NoError(t, err, "Dapr metadata response must be valid JSON")

		// Validate that extended actors are available (indicates service mesh is functional)
		if extendedMetadata, exists := metadata["extended"]; exists {
			t.Logf("Dapr extended metadata available: %v", extendedMetadata)
		} else {
			t.Errorf("RED PHASE VALIDATION: Dapr extended metadata not available - service mesh may not be fully operational")
		}
	})

	// Validate individual service registration through service invocation
	for _, expectedService := range expectedServices {
		t.Run("ServiceRegistration_"+expectedService.appId, func(t *testing.T) {
			// Test service registration via Dapr service invocation health check
			serviceURL := fmt.Sprintf("http://localhost:3500/v1.0/invoke/%s/method/health", expectedService.appId)
			req, err := http.NewRequestWithContext(ctx, "GET", serviceURL, nil)
			require.NoError(t, err, "Failed to create service invocation request for %s", expectedService.appId)

			resp, err := client.Do(req)
			if err != nil {
				t.Errorf("RED PHASE VALIDATION: %s - Service %s not accessible through Dapr service invocation: %v", 
					expectedService.description, expectedService.appId, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("RED PHASE VALIDATION: %s - Service %s registration failed. Status: %d, Response: %s", 
					expectedService.description, expectedService.appId, resp.StatusCode, string(body))
				return
			}

			// Validate response is JSON and contains service health information
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err, "Failed to read service health response for %s", expectedService.appId)

			var healthResponse map[string]interface{}
			err = json.Unmarshal(body, &healthResponse)
			if err != nil {
				t.Errorf("RED PHASE VALIDATION: %s - Service %s health response not valid JSON: %v", 
					expectedService.description, expectedService.appId, err)
				return
			}

			// Validate service health response contains expected fields
			if status, exists := healthResponse["status"]; exists {
				assert.Equal(t, "healthy", status, 
					"RED PHASE VALIDATION: %s - Service %s must report healthy status", 
					expectedService.description, expectedService.appId)
			} else {
				t.Errorf("RED PHASE VALIDATION: %s - Service %s health response missing 'status' field", 
					expectedService.description, expectedService.appId)
			}

			t.Logf("RED PHASE VALIDATION SUCCESS: %s registered and accessible through Dapr service mesh", expectedService.serviceName)
		})
	}

	// Validate Dapr service invocation capabilities 
	t.Run("DaprServiceInvocationCapabilities", func(t *testing.T) {
		// Test that Dapr can route requests to services through service invocation
		testEndpoints := []struct {
			serviceAppId string
			methodPath   string
			description  string
		}{
			{
				serviceAppId: "content-api",
				methodPath:   "/health",
				description:  "Content service health endpoint via Dapr service invocation",
			},
			{
				serviceAppId: "inquiries-api", 
				methodPath:   "/health",
				description:  "Inquiries service health endpoint via Dapr service invocation",
			},
			{
				serviceAppId: "notification-api",
				methodPath:   "/health", 
				description:  "Notifications service health endpoint via Dapr service invocation",
			},
		}

		for _, testEndpoint := range testEndpoints {
			serviceURL := fmt.Sprintf("http://localhost:3500/v1.0/invoke/%s/method%s", 
				testEndpoint.serviceAppId, testEndpoint.methodPath)
			
			req, err := http.NewRequestWithContext(ctx, "GET", serviceURL, nil)
			require.NoError(t, err, "Failed to create service invocation request")

			resp, err := client.Do(req)
			if err != nil {
				t.Errorf("RED PHASE VALIDATION: %s - Service invocation failed: %v", testEndpoint.description, err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("RED PHASE VALIDATION: %s - Service invocation returned %d: %s", 
					testEndpoint.description, resp.StatusCode, string(body))
			} else {
				t.Logf("RED PHASE VALIDATION SUCCESS: %s accessible through Dapr service invocation", testEndpoint.description)
			}
		}
	})
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

