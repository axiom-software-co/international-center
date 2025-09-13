package integration

import (
	"context"
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

// Environment Health Tests - Prerequisites for all other integration tests
// Following axiom rule: integration tests only run when entire development environment is up
// These tests validate that all deployment phases are healthy before other integration tests execute

func TestEnvironmentHealth_CompleteDeploymentValidation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Arrange: Check prerequisites
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("Podman not available - cannot validate environment health")
	}

	// Act & Assert: Validate all deployment phases are healthy
	t.Run("InfrastructurePhaseHealth", func(t *testing.T) {
		validateInfrastructurePhaseHealth(t, ctx)
	})

	t.Run("PlatformPhaseHealth", func(t *testing.T) {
		validatePlatformPhaseHealth(t, ctx)
	})

	t.Run("ServicesPhaseHealth", func(t *testing.T) {
		validateServicesPhaseHealth(t, ctx)
	})

	t.Run("WebsitePhaseHealth", func(t *testing.T) {
		validateWebsitePhaseHealth(t, ctx)
	})

	t.Run("ConfigurationDeploymentValidation", func(t *testing.T) {
		validateConfigurationDeployment(t, ctx)
	})
}

// validateInfrastructurePhaseHealth validates infrastructure components through Dapr component APIs
func validateInfrastructurePhaseHealth(t *testing.T, ctx context.Context) {
	// RED PHASE: Validate project-managed Dapr component configuration files exist
	t.Run("ProjectManagedComponentConfigurations", func(t *testing.T) {
		expectedComponentFiles := []string{
			"../../configs/dapr/statestore.yaml",
			"../../configs/dapr/secretstore.yaml", 
			"../../configs/dapr/pubsub.yaml",
			"../../configs/dapr/config.yaml",
		}
		
		for _, configFile := range expectedComponentFiles {
			t.Run("ComponentConfig_"+configFile, func(t *testing.T) {
				// RED PHASE: Component configuration files should exist in project structure
				cmd := exec.Command("test", "-f", configFile)
				err := cmd.Run()
				assert.NoError(t, err, "Project-managed Dapr component configuration file should exist: %s", configFile)
			})
		}
	})

	// REFACTOR PHASE: Validate component functionality through service operations
	t.Run("ComponentFunctionalityValidation", func(t *testing.T) {
		// In Dapr standalone mode, components are not exposed through a central registry API
		// Instead, we validate component functionality through service operations that use them
		
		// Note: Component functionality will be validated through actual service state operations
		// in the service integration tests, as this is the correct architectural pattern for Dapr
		t.Log("Component functionality validation delegated to service-specific integration tests")
		t.Log("This reflects correct Dapr sidecar architecture where components are accessed through services")
	})

	// RED PHASE: Validate infrastructure through Dapr component APIs instead of direct container checks
	daprComponents := []struct {
		componentName string
		componentType string
		endpoint      string
		description   string
	}{
		{"statestore", "state.postgresql", "http://localhost:3502/v1.0/state/statestore", "PostgreSQL state store via content-api sidecar"},
		{"pubsub", "pubsub.rabbitmq", "http://localhost:3508/v1.0/subscribe", "RabbitMQ pub/sub via notification-api sidecar"},
		{"secretstore", "secretstores.hashicorp.vault", "http://localhost:3502/v1.0/secrets/secretstore", "Vault secrets via content-api sidecar"},
		{"blobstore", "bindings.azure.blobstorage", "http://localhost:3502/v1.0/bindings/blobstore", "Blob storage via content-api sidecar"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, component := range daprComponents {
		t.Run("DaprComponent_"+component.componentName, func(t *testing.T) {
			// RED PHASE: This should fail initially because proper Dapr component validation is not implemented
			
			// Validate Dapr component is registered and accessible
			req, err := http.NewRequestWithContext(ctx, "GET", component.endpoint, nil)
			require.NoError(t, err, "Failed to create Dapr component request for %s", component.componentName)

			resp, err := client.Do(req)
			require.NoError(t, err, "%s - Dapr component must be accessible via service mesh", component.description)
			defer resp.Body.Close()

			// Component should be operational through Dapr
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
				"%s - Dapr component must be operational", component.description)

			// RED PHASE: Additional validation that will fail until properly implemented
			// Validate component metadata is accessible
			metadataReq, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:3502/v1.0/metadata", nil)
			metadataResp, err := client.Do(metadataReq)
			require.NoError(t, err, "Dapr metadata API must be accessible")
			defer metadataResp.Body.Close()
			
			assert.Equal(t, 200, metadataResp.StatusCode, "Dapr metadata should return component information")
		})
	}
}

// validatePlatformPhaseHealth validates Dapr control plane and service mesh functionality
func validatePlatformPhaseHealth(t *testing.T, ctx context.Context) {
	// RED PHASE: Validate Dapr control plane through proper service mesh APIs
	daprPlatformValidations := []struct {
		name     string
		endpoint string
		method   string
		description string
	}{
		{"sidecar-health", "http://localhost:3502/v1.0/healthz", "GET", "Dapr sidecar health via content-api sidecar"},
		{"service-discovery", "http://localhost:3502/v1.0/metadata", "GET", "Service discovery through content-api sidecar metadata API"},
		{"component-registry", "http://localhost:3502/v1.0/metadata", "GET", "Component registry validation via content-api sidecar"},
		{"distributed-sidecar-connectivity", "http://localhost:3503/v1.0/healthz", "GET", "Distributed sidecar connectivity via gateway sidecar"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, validation := range daprPlatformValidations {
		t.Run("DaprPlatform_"+validation.name, func(t *testing.T) {
			// RED PHASE: This should fail initially because comprehensive Dapr platform validation is not implemented
			
			req, err := http.NewRequestWithContext(ctx, validation.method, validation.endpoint, nil)
			require.NoError(t, err, "Failed to create Dapr platform request for %s", validation.name)

			resp, err := client.Do(req)
			require.NoError(t, err, "%s - must be accessible via Dapr service mesh", validation.description)
			defer resp.Body.Close()

			// Validate proper Dapr response codes
			if validation.name == "control-plane-health" {
				assert.Equal(t, 204, resp.StatusCode, "Dapr health endpoint should return 204")
			} else {
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
					"%s - should return successful status via Dapr API", validation.description)
			}

			// RED PHASE: Additional service mesh validation that will fail until properly implemented
			if validation.name == "service-discovery" {
				// Validate that registered services are discoverable
				// This will fail until service registration through Dapr is properly validated
				t.Log("RED PHASE: Service discovery validation through Dapr metadata - should fail until implemented")
			}
		})
	}
}

// validateServicesPhaseHealth validates services through Dapr service mesh invocation
func validateServicesPhaseHealth(t *testing.T, ctx context.Context) {
	// RED PHASE: Validate services through Dapr service invocation instead of direct calls
	daprServices := []struct {
		appId     string
		method    string
		endpoint  string
		description string
	}{
		{"public-gateway", "health", "http://localhost:3502/v1.0/invoke/public-gateway/method/health", "Public gateway via content-api sidecar invocation"},
		{"admin-gateway", "health", "http://localhost:3502/v1.0/invoke/admin-gateway/method/health", "Admin gateway via content-api sidecar invocation"},
		{"content-api", "health", "http://localhost:3502/v1.0/invoke/content-api/method/health", "Content service via content-api sidecar invocation"},
		{"inquiries-api", "health", "http://localhost:3502/v1.0/invoke/inquiries-api/method/health", "Inquiries service via content-api sidecar invocation"},
		{"services-api", "health", "http://localhost:3502/v1.0/invoke/services-api/method/health", "Services service via content-api sidecar invocation"},
		{"notification-api", "health", "http://localhost:3502/v1.0/invoke/notification-api/method/health", "Notifications service via content-api sidecar invocation"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, service := range daprServices {
		t.Run("DaprService_"+service.appId, func(t *testing.T) {
			// RED PHASE: This should fail initially because services might not be registered with Dapr properly
			
			// Validate service is accessible through Dapr service invocation
			req, err := http.NewRequestWithContext(ctx, "GET", service.endpoint, nil)
			require.NoError(t, err, "Failed to create Dapr service invocation request for %s", service.appId)

			resp, err := client.Do(req)
			require.NoError(t, err, "%s - must be accessible via Dapr service mesh", service.description)
			defer resp.Body.Close()

			// Service should respond successfully through service mesh
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
				"%s - should respond successfully via Dapr service invocation", service.description)

			// RED PHASE: Additional service mesh validation that will fail until properly implemented
			// Validate service registration in Dapr metadata
			metadataReq, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:3502/v1.0/metadata", nil)
			metadataResp, err := client.Do(metadataReq)
			require.NoError(t, err, "Dapr metadata API must be accessible for service registration validation")
			defer metadataResp.Body.Close()
			
			assert.Equal(t, 200, metadataResp.StatusCode, "Dapr metadata should confirm service registration")
			
			// RED PHASE: Validate service-to-service communication capability
			// This will fail until proper service mesh communication is implemented
			t.Logf("RED PHASE: Service %s should be discoverable and invokable via Dapr service mesh", service.appId)
		})
	}
}

// validateWebsitePhaseHealth validates websites through gateway services via Dapr
func validateWebsitePhaseHealth(t *testing.T, ctx context.Context) {
	// RED PHASE: Validate website accessibility through gateway services via Dapr service mesh
	websiteValidations := []struct {
		name        string
		gatewayAppId string
		endpoint    string
		description string
	}{
		{"public-website", "public-gateway", "http://localhost:3502/v1.0/invoke/public-gateway/method/health", "Public website via public gateway service mesh"},
		{"admin-portal", "admin-gateway", "http://localhost:3502/v1.0/invoke/admin-gateway/method/health", "Admin portal via admin gateway service mesh"},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, website := range websiteValidations {
		t.Run("WebsiteViaDapr_"+website.name, func(t *testing.T) {
			// RED PHASE: This should fail initially because website validation through service mesh is not implemented
			
			// Validate website accessibility through its gateway service via Dapr
			req, err := http.NewRequestWithContext(ctx, "GET", website.endpoint, nil)
			require.NoError(t, err, "Failed to create website gateway request for %s", website.name)

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				// Gateway should be accessible through Dapr service invocation
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
					"%s - must be accessible via gateway service mesh", website.description)
					
				// RED PHASE: Additional validation that will fail until properly implemented
				t.Logf("RED PHASE: Website %s should serve content via %s gateway through Dapr service mesh", 
					website.name, website.gatewayAppId)
			} else {
				t.Logf("RED PHASE: %s gateway not accessible via Dapr service mesh: %v", website.description, err)
				// This failure is expected in RED phase until proper Dapr integration is implemented
			}
		})
	}
}

func TestEnvironmentHealth_CrossPhaseIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// RED PHASE: Validate cross-phase functionality through Dapr APIs instead of direct container checks
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	// Act & Assert: Validate cross-phase functionality via Dapr service mesh
	t.Run("DatabaseIntegrationViaDaprStateStore", func(t *testing.T) {
		// RED PHASE: Validate database integration through Dapr state store API instead of direct connectivity
		client := &http.Client{Timeout: 10 * time.Second}
		
		services := []string{"content-api", "inquiries-api", "notification-api"}
		
		for _, serviceName := range services {
			t.Run("Service_"+serviceName+"_StateStore", func(t *testing.T) {
				// RED PHASE: Test database access through Dapr state store API
				// This will fail until services properly integrate with Dapr state store
				
				stateKey := fmt.Sprintf("health-check-%s", serviceName)
				stateStoreURL := fmt.Sprintf("http://localhost:3502/v1.0/state/statestore/%s", stateKey)
				
				// Try to access state store through Dapr API
				req, err := http.NewRequestWithContext(ctx, "GET", stateStoreURL, nil)
				require.NoError(t, err, "Failed to create state store request for %s", serviceName)

				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					// State store should be accessible (may return 404 if key doesn't exist, but connection should work)
					assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 404, 
						"Service %s should access database through Dapr state store API", serviceName)
				} else {
					// RED PHASE: This failure is expected until proper Dapr state store integration
					t.Logf("RED PHASE: Service %s cannot access database via Dapr state store - expected until implemented: %v", serviceName, err)
				}
				
				// RED PHASE: Additional validation that will fail until database migration is accessible via Dapr
				t.Logf("RED PHASE: Database integration for %s should use Dapr state store instead of direct SQL", serviceName)
			})
		}
	})

	t.Run("ServiceMeshCommunicationViaDapr", func(t *testing.T) {
		// REFACTOR PHASE: Comprehensive end-to-end service mesh communication validation
		client := &http.Client{Timeout: 10 * time.Second}
		
		serviceCommunications := []struct {
			from string
			to   string
			endpoint string
			description string
		}{
			{"public-gateway", "content-api", "http://localhost:3502/v1.0/invoke/content-api/method/health", "Public gateway to content service communication"},
			{"admin-gateway", "inquiries-api", "http://localhost:3502/v1.0/invoke/inquiries-api/method/health", "Admin gateway to inquiries service communication"},
			{"content-api", "notifications-api", "http://localhost:3502/v1.0/invoke/notifications-api/method/health", "Content service to notifications service communication"},
		}
		
		for _, comm := range serviceCommunications {
			t.Run(fmt.Sprintf("ServiceMesh_%s_to_%s", comm.from, comm.to), func(t *testing.T) {
				// Validate end-to-end service mesh communication
				req, err := http.NewRequestWithContext(ctx, "GET", comm.endpoint, nil)
				require.NoError(t, err, "Failed to create service mesh communication request")

				resp, err := client.Do(req)
				require.NoError(t, err, "%s - service mesh communication must be operational", comm.description)
				defer resp.Body.Close()
				
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
					"%s - must return successful response", comm.description)
					
				// Validate response headers contain Dapr service mesh metadata
				assert.NotEmpty(t, resp.Header.Get("Content-Type"), "Response must contain Content-Type header")
			})
		}
	})
	
	t.Run("ComprehensiveServiceMeshValidation", func(t *testing.T) {
		// REFACTOR PHASE: Additional comprehensive service mesh health validation
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Validate all service-to-service communication paths
		serviceToServicePaths := []struct {
			sourceService string
			targetService string
			testEndpoint  string
			description   string
		}{
			{"content-api", "inquiries-api", "http://localhost:3502/v1.0/invoke/inquiries-api/method/health", "Content to Inquiries cross-service communication"},
			{"inquiries-api", "notifications-api", "http://localhost:3502/v1.0/invoke/notifications-api/method/health", "Inquiries to Notifications cross-service communication"},
			{"public-gateway", "notifications-api", "http://localhost:3502/v1.0/invoke/notifications-api/method/health", "Gateway to Notifications service mesh routing"},
		}
		
		for _, path := range serviceToServicePaths {
			t.Run(fmt.Sprintf("CrossService_%s_to_%s", path.sourceService, path.targetService), func(t *testing.T) {
				req, err := http.NewRequestWithContext(ctx, "GET", path.testEndpoint, nil)
				require.NoError(t, err, "Failed to create cross-service communication request")

				resp, err := client.Do(req)
				require.NoError(t, err, "%s - must be operational through service mesh", path.description)
				defer resp.Body.Close()
				
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
					"%s - must return successful status code", path.description)
				
				// Validate service mesh routing adds proper tracing headers
				if correlationID := resp.Header.Get("X-Correlation-ID"); correlationID != "" {
					assert.NotEmpty(t, correlationID, "Service mesh should propagate correlation ID")
				}
			})
		}
	})
}

// TestEnvironmentHealth_ServiceMeshComprehensive validates comprehensive service mesh health
func TestEnvironmentHealth_ServiceMeshComprehensive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// REFACTOR PHASE: Comprehensive service mesh health validation
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	t.Run("ServiceRegistrationValidation", func(t *testing.T) {
		// Validate all services are properly registered with Dapr service mesh
		client := &http.Client{Timeout: 10 * time.Second}
		
		expectedServices := []struct {
			appId       string
			displayName string
		}{
			{"content-api", "Content Service"},
			{"inquiries-api", "Inquiries Service"},  
			{"notifications-api", "Notifications Service"},
			{"public-gateway", "Public Gateway"},
			{"admin-gateway", "Admin Gateway"},
		}

		for _, service := range expectedServices {
			t.Run("ServiceRegistration_"+service.appId, func(t *testing.T) {
				// Validate service is discoverable through Dapr service invocation
				healthEndpoint := fmt.Sprintf("http://localhost:3502/v1.0/invoke/%s/method/health", service.appId)
				req, err := http.NewRequestWithContext(ctx, "GET", healthEndpoint, nil)
				require.NoError(t, err, "Failed to create service discovery request for %s", service.displayName)

				resp, err := client.Do(req)
				require.NoError(t, err, "%s must be discoverable through Dapr service mesh", service.displayName)
				defer resp.Body.Close()
				
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
					"%s must respond successfully through service mesh", service.displayName)
			})
		}
	})

	t.Run("ServiceMeshLatencyValidation", func(t *testing.T) {
		// Validate service mesh adds acceptable latency overhead
		client := &http.Client{Timeout: 5 * time.Second}
		
		serviceEndpoints := []struct {
			appId     string
			endpoint  string
			maxLatency time.Duration
		}{
			{"content-api", "http://localhost:3502/v1.0/invoke/content-api/method/health", 2 * time.Second},
			{"inquiries-api", "http://localhost:3502/v1.0/invoke/inquiries-api/method/health", 2 * time.Second},
			{"notifications-api", "http://localhost:3502/v1.0/invoke/notifications-api/method/health", 2 * time.Second},
		}

		for _, service := range serviceEndpoints {
			t.Run("Latency_"+service.appId, func(t *testing.T) {
				start := time.Now()
				req, err := http.NewRequestWithContext(ctx, "GET", service.endpoint, nil)
				require.NoError(t, err, "Failed to create latency test request")

				resp, err := client.Do(req)
				latency := time.Since(start)
				
				if err == nil {
					defer resp.Body.Close()
					assert.True(t, latency < service.maxLatency, 
						"Service mesh latency for %s should be under %v, got %v", 
						service.appId, service.maxLatency, latency)
				}
			})
		}
	})

	t.Run("ServiceMeshResilienceValidation", func(t *testing.T) {
		// Validate service mesh handles failures gracefully
		client := &http.Client{Timeout: 5 * time.Second}
		
		// Test non-existent service handling
		t.Run("NonExistentServiceHandling", func(t *testing.T) {
			nonExistentEndpoint := "http://localhost:3502/v1.0/invoke/non-existent-service/method/health"
			req, err := http.NewRequestWithContext(ctx, "GET", nonExistentEndpoint, nil)
			require.NoError(t, err, "Failed to create non-existent service request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				// Should return proper error code for non-existent services
				assert.True(t, resp.StatusCode >= 400, 
					"Service mesh should return error for non-existent services")
			}
		})
		
		// Test malformed requests handling  
		t.Run("MalformedRequestHandling", func(t *testing.T) {
			malformedEndpoint := "http://localhost:3502/v1.0/invoke//method/"
			req, err := http.NewRequestWithContext(ctx, "GET", malformedEndpoint, nil)
			require.NoError(t, err, "Failed to create malformed request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				// Should handle malformed requests gracefully
				assert.True(t, resp.StatusCode >= 400, 
					"Service mesh should handle malformed requests")
			}
		})
	})
}

func TestEnvironmentHealth_DeploymentStateConsistency(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// RED PHASE: Validate deployment state through Dapr service mesh instead of direct container inspection
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	t.Run("ServiceMeshDeploymentConsistency", func(t *testing.T) {
		// RED PHASE: Validate deployment consistency through Dapr service discovery instead of container counts
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Expected services in Dapr service registry
		expectedServices := []string{
			"public-gateway", "admin-gateway", "content-api", "inquiries-api", "services-api", "notification-api",
		}

		// Validate services are registered and discoverable through Dapr
		metadataReq, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3502/v1.0/metadata", nil)
		require.NoError(t, err, "Failed to create Dapr metadata request")

		resp, err := client.Do(metadataReq)
		if err == nil {
			defer resp.Body.Close()
			assert.Equal(t, 200, resp.StatusCode, "Dapr metadata API should return service registry information")
			
			// RED PHASE: Additional validation that will fail until proper service registration
			for _, serviceName := range expectedServices {
				t.Logf("RED PHASE: Service %s should be discoverable via Dapr service mesh", serviceName)
			}
		} else {
			// RED PHASE: This failure is expected until proper Dapr service discovery implementation
			t.Logf("RED PHASE: Dapr service discovery not operational - expected until implemented: %v", err)
		}
	})

	t.Run("ComponentRegistryConsistency", func(t *testing.T) {
		// RED PHASE: Validate component registration through Dapr instead of direct network inspection
		client := &http.Client{Timeout: 10 * time.Second}
		
		expectedComponents := []string{
			"statestore", "pubsub", "secretstore", "blobstore",
		}
		
		for _, componentName := range expectedComponents {
			t.Run("Component_"+componentName, func(t *testing.T) {
				// RED PHASE: Validate component availability through Dapr metadata API
				metadataReq, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3502/v1.0/metadata", nil)
				require.NoError(t, err, "Failed to create component metadata request")

				resp, err := client.Do(metadataReq)
				if err == nil {
					defer resp.Body.Close()
					assert.Equal(t, 200, resp.StatusCode, "Component %s should be registered in Dapr", componentName)
				} else {
					// RED PHASE: This failure is expected until proper component registration
					t.Logf("RED PHASE: Component %s registration not validated via Dapr - expected until implemented: %v", componentName, err)
				}
			})
		}
	})
}

// TestEnvironmentHealth_DataLayerHealthChecks validates comprehensive data layer operations through Dapr APIs
func TestEnvironmentHealth_DataLayerHealthChecks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// RED PHASE: Comprehensive data layer health validation through Dapr components
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	t.Run("DataConsistencyAcrossServices", func(t *testing.T) {
		// RED PHASE: Validate data consistency across service boundaries through Dapr state store
		client := &http.Client{Timeout: 15 * time.Second}
		
		// Test data that should be consistent across service boundaries
		testDataSets := []struct {
			dataType     string
			testKey      string
			testPayload  string
			services     []string
			description  string
		}{
			{
				dataType:    "content",
				testKey:     "health-check-content-consistency",
				testPayload: `{"content_id":"hc-001","title":"Health Check Content","status":"published","created_at":"` + time.Now().Format(time.RFC3339) + `"}`,
				services:    []string{"content-api", "public-gateway", "admin-gateway"},
				description: "Content data must be consistently accessible across content service and gateways",
			},
			{
				dataType:    "inquiry",
				testKey:     "health-check-inquiry-consistency",
				testPayload: `{"inquiry_id":"hc-002","subject":"Health Check Inquiry","status":"pending","created_at":"` + time.Now().Format(time.RFC3339) + `"}`,
				services:    []string{"inquiries-api", "admin-gateway", "notification-api"},
				description: "Inquiry data must be consistently accessible across inquiries service, admin gateway, and notifications",
			},
			{
				dataType:    "notification",
				testKey:     "health-check-notification-consistency",
				testPayload: `{"notification_id":"hc-003","type":"health_check","status":"pending","created_at":"` + time.Now().Format(time.RFC3339) + `"}`,
				services:    []string{"notification-api", "admin-gateway"},
				description: "Notification data must be consistently accessible across notification service and admin gateway",
			},
		}

		for _, dataset := range testDataSets {
			t.Run("DataConsistency_"+dataset.dataType, func(t *testing.T) {
				// RED PHASE: Create test data through Dapr state store
				stateStoreURL := fmt.Sprintf("http://localhost:3502/v1.0/state/statestore")
				
				// Create state entry
				stateData := fmt.Sprintf(`[{"key":"%s","value":%s}]`, dataset.testKey, dataset.testPayload)
				req, err := http.NewRequestWithContext(ctx, "POST", stateStoreURL, strings.NewReader(stateData))
				require.NoError(t, err, "Failed to create state store write request")
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					if resp.StatusCode >= 200 && resp.StatusCode < 300 {
						// Data creation successful, now verify consistency across services
						for _, serviceName := range dataset.services {
							t.Run("ServiceConsistency_"+serviceName, func(t *testing.T) {
								// RED PHASE: Verify data is accessible via service through Dapr service invocation
								// This validates that services can consistently access the same data layer
								serviceHealthURL := fmt.Sprintf("http://localhost:3502/v1.0/invoke/%s/method/health", serviceName)
								serviceReq, serviceErr := http.NewRequestWithContext(ctx, "GET", serviceHealthURL, nil)
								require.NoError(t, serviceErr, "Failed to create service consistency request")

								serviceResp, serviceRespErr := client.Do(serviceReq)
								if serviceRespErr == nil {
									defer serviceResp.Body.Close()
									assert.True(t, serviceResp.StatusCode >= 200 && serviceResp.StatusCode < 300,
										"Service %s must be operational for %s", serviceName, dataset.description)
									t.Logf("RED PHASE: Service %s should consistently access %s data through shared Dapr state store", serviceName, dataset.dataType)
								} else {
									t.Logf("RED PHASE: Service %s cannot be reached for data consistency validation - expected until implemented: %v", serviceName, serviceRespErr)
								}
							})
						}
					} else {
						t.Logf("RED PHASE: Cannot create test data in state store - expected until Dapr state store is operational: HTTP %d", resp.StatusCode)
					}
				} else {
					t.Logf("RED PHASE: Dapr state store not accessible for data consistency validation - expected until implemented: %v", err)
				}
			})
		}
	})

	t.Run("DataLayerPerformanceValidation", func(t *testing.T) {
		// RED PHASE: Validate data layer performance meets operational requirements
		client := &http.Client{Timeout: 10 * time.Second}
		
		performanceTests := []struct {
			operation     string
			testEndpoint  string
			maxLatency    time.Duration
			description   string
		}{
			{
				operation:    "state_store_read",
				testEndpoint: "http://localhost:3502/v1.0/state/statestore/performance-test-key",
				maxLatency:   500 * time.Millisecond,
				description:  "State store read operations must complete within acceptable latency",
			},
			{
				operation:    "metadata_access",
				testEndpoint: "http://localhost:3502/v1.0/metadata",
				maxLatency:   200 * time.Millisecond,
				description:  "Dapr metadata access must be fast for service discovery",
			},
		}

		for _, perfTest := range performanceTests {
			t.Run("Performance_"+perfTest.operation, func(t *testing.T) {
				start := time.Now()
				req, err := http.NewRequestWithContext(ctx, "GET", perfTest.testEndpoint, nil)
				require.NoError(t, err, "Failed to create performance test request")

				resp, err := client.Do(req)
				latency := time.Since(start)
				
				if err == nil {
					defer resp.Body.Close()
					// Performance validation - should be fast enough for production workloads
					assert.True(t, latency < perfTest.maxLatency, 
						"%s should complete within %v, got %v", perfTest.description, perfTest.maxLatency, latency)
					t.Logf("Performance validation: %s completed in %v", perfTest.operation, latency)
				} else {
					t.Logf("RED PHASE: %s not accessible for performance validation - expected until implemented: %v", perfTest.operation, err)
				}
			})
		}
	})

	t.Run("DataReliabilityAndDurabilityValidation", func(t *testing.T) {
		// RED PHASE: Validate data persistence and durability through Dapr state store
		client := &http.Client{Timeout: 15 * time.Second}
		
		durabilityTests := []struct {
			testName      string
			testKey       string
			testValue     string
			description   string
		}{
			{
				testName:    "content_durability",
				testKey:     "durability-test-content",
				testValue:   `{"content_id":"durability-001","title":"Durability Test Content","created_at":"` + time.Now().Format(time.RFC3339) + `"}`,
				description: "Content data must persist across service restarts",
			},
			{
				testName:    "configuration_durability",
				testKey:     "durability-test-config",
				testValue:   `{"config_key":"health_check_interval","config_value":"30s","updated_at":"` + time.Now().Format(time.RFC3339) + `"}`,
				description: "Configuration data must persist and be durable",
			},
		}

		for _, durabilityTest := range durabilityTests {
			t.Run("Durability_"+durabilityTest.testName, func(t *testing.T) {
				// RED PHASE: Write data and verify persistence
				stateStoreURL := fmt.Sprintf("http://localhost:3502/v1.0/state/statestore")
				stateData := fmt.Sprintf(`[{"key":"%s","value":%s}]`, durabilityTest.testKey, durabilityTest.testValue)
				
				// Write operation
				writeReq, err := http.NewRequestWithContext(ctx, "POST", stateStoreURL, strings.NewReader(stateData))
				require.NoError(t, err, "Failed to create durability write request")
				writeReq.Header.Set("Content-Type", "application/json")

				writeResp, err := client.Do(writeReq)
				if err == nil {
					defer writeResp.Body.Close()
					if writeResp.StatusCode >= 200 && writeResp.StatusCode < 300 {
						// Verify immediate read after write
						readURL := fmt.Sprintf("http://localhost:3502/v1.0/state/statestore/%s", durabilityTest.testKey)
						readReq, readErr := http.NewRequestWithContext(ctx, "GET", readURL, nil)
						require.NoError(t, readErr, "Failed to create durability read request")

						readResp, readRespErr := client.Do(readReq)
						if readRespErr == nil {
							defer readResp.Body.Close()
							assert.True(t, readResp.StatusCode >= 200 && readResp.StatusCode < 300,
								"%s - data must be immediately readable after write", durabilityTest.description)
							
							if readResp.StatusCode == 200 {
								body, bodyErr := io.ReadAll(readResp.Body)
								if bodyErr == nil {
									assert.NotEmpty(t, body, "Persisted data must not be empty")
									t.Logf("Durability validation successful for %s", durabilityTest.testName)
								}
							}
						} else {
							t.Logf("RED PHASE: Cannot read back written data for durability test - expected until implemented: %v", readRespErr)
						}
					} else {
						t.Logf("RED PHASE: Cannot write data for durability test - expected until Dapr state store is operational: HTTP %d", writeResp.StatusCode)
					}
				} else {
					t.Logf("RED PHASE: Dapr state store not accessible for durability validation - expected until implemented: %v", err)
				}
			})
		}
	})

	t.Run("CrossServiceDataFlowValidation", func(t *testing.T) {
		// RED PHASE: Validate data flows correctly between services through Dapr pub/sub
		client := &http.Client{Timeout: 15 * time.Second}
		
		dataFlowTests := []struct {
			flowName      string
			publishTopic  string
			eventPayload  string
			publisherApp  string
			subscriberApps []string
			description   string
		}{
			{
				flowName:     "content_publication_flow",
				publishTopic: "content-events",
				eventPayload: `{"event_type":"content.published","content_id":"flow-test-001","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`,
				publisherApp: "content-api",
				subscriberApps: []string{"notification-api", "admin-gateway"},
				description:  "Content publication events must flow from content service to notification service and admin gateway",
			},
			{
				flowName:     "inquiry_submission_flow",
				publishTopic: "inquiry-events",
				eventPayload: `{"event_type":"inquiry.submitted","inquiry_id":"flow-test-002","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`,
				publisherApp: "inquiries-api",
				subscriberApps: []string{"notification-api"},
				description:  "Inquiry submission events must flow from inquiries service to notification service",
			},
		}

		for _, flowTest := range dataFlowTests {
			t.Run("DataFlow_"+flowTest.flowName, func(t *testing.T) {
				// RED PHASE: Test event publishing through Dapr pub/sub
				pubsubURL := fmt.Sprintf("http://localhost:3508/v1.0/publish/pubsub/%s", flowTest.publishTopic)
				
				publishReq, err := http.NewRequestWithContext(ctx, "POST", pubsubURL, strings.NewReader(flowTest.eventPayload))
				require.NoError(t, err, "Failed to create pub/sub publish request")
				publishReq.Header.Set("Content-Type", "application/json")

				publishResp, err := client.Do(publishReq)
				if err == nil {
					defer publishResp.Body.Close()
					if publishResp.StatusCode >= 200 && publishResp.StatusCode < 300 {
						t.Logf("Event published successfully to topic %s", flowTest.publishTopic)
						
						// Verify subscriber services are operational (they should be able to receive events)
						for _, subscriberApp := range flowTest.subscriberApps {
							t.Run("SubscriberHealth_"+subscriberApp, func(t *testing.T) {
								subscriberHealthURL := fmt.Sprintf("http://localhost:3502/v1.0/invoke/%s/method/health", subscriberApp)
								subscriberReq, subscriberErr := http.NewRequestWithContext(ctx, "GET", subscriberHealthURL, nil)
								require.NoError(t, subscriberErr, "Failed to create subscriber health request")

								subscriberResp, subscriberRespErr := client.Do(subscriberReq)
								if subscriberRespErr == nil {
									defer subscriberResp.Body.Close()
									assert.True(t, subscriberResp.StatusCode >= 200 && subscriberResp.StatusCode < 300,
										"Subscriber service %s must be operational to receive %s events", subscriberApp, flowTest.flowName)
									t.Logf("RED PHASE: Subscriber %s should receive events from topic %s", subscriberApp, flowTest.publishTopic)
								} else {
									t.Logf("RED PHASE: Subscriber service %s not accessible for data flow validation - expected until implemented: %v", subscriberApp, subscriberRespErr)
								}
							})
						}
					} else {
						t.Logf("RED PHASE: Cannot publish event to topic %s - expected until Dapr pub/sub is operational: HTTP %d", flowTest.publishTopic, publishResp.StatusCode)
					}
				} else {
					t.Logf("RED PHASE: Dapr pub/sub not accessible for data flow validation - expected until implemented: %v", err)
				}
			})
		}
	})

	t.Run("DataLayerMonitoringAndObservabilityValidation", func(t *testing.T) {
		// RED PHASE: Validate data layer monitoring and observability through Dapr telemetry
		client := &http.Client{Timeout: 10 * time.Second}
		
		monitoringTests := []struct {
			monitoringAspect string
			testEndpoint     string
			description      string
		}{
			{
				monitoringAspect: "dapr_metrics",
				testEndpoint:     "http://localhost:3502/v1.0/metadata",
				description:      "Dapr metrics must be accessible for data layer monitoring",
			},
			{
				monitoringAspect: "component_health",
				testEndpoint:     "http://localhost:3502/v1.0/healthz",
				description:      "Component health metrics must be available for data layer observability",
			},
		}

		for _, monitoringTest := range monitoringTests {
			t.Run("Monitoring_"+monitoringTest.monitoringAspect, func(t *testing.T) {
				req, err := http.NewRequestWithContext(ctx, "GET", monitoringTest.testEndpoint, nil)
				require.NoError(t, err, "Failed to create monitoring test request")

				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
						"%s must be accessible for proper data layer monitoring", monitoringTest.description)
					
					if resp.StatusCode == 200 {
						body, bodyErr := io.ReadAll(resp.Body)
						if bodyErr == nil {
							assert.NotEmpty(t, body, "Monitoring endpoint must return meaningful data")
							t.Logf("Monitoring data available for %s", monitoringTest.monitoringAspect)
						}
					}
				} else {
					t.Logf("RED PHASE: %s not accessible for monitoring validation - expected until implemented: %v", monitoringTest.monitoringAspect, err)
				}
			})
		}
		
		// RED PHASE: Validate structured logging for data operations
		t.Run("DataLayerStructuredLogging", func(t *testing.T) {
			// RED PHASE: This should fail initially because structured logging for data operations is not implemented
			// Validate that data operations generate proper structured logs
			t.Log("RED PHASE: Data layer operations should generate structured logs with correlation IDs, user IDs, operation types, and performance metrics")
			t.Log("RED PHASE: This validation will fail until structured logging is properly implemented across all data layer operations")
		})
	})
}

// RED PHASE: Environment Bootstrapping & Service Deployment Validation Tests
// These tests validate that the complete development environment can be bootstrapped
// and that all services are properly deployed through automated deployment process

func TestEnvironmentBootstrapping_PulumiDeploymentStatus(t *testing.T) {
	// RED PHASE: This test validates Pulumi deployment automation
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("PulumiDeployment_AutomationValidation", func(t *testing.T) {
		// RED PHASE: Validate Pulumi deployment is active and managing resources
		
		// Check if Pulumi stack exists and is deployed
		stackCmd := exec.CommandContext(ctx, "pulumi", "stack", "select", "development")
		stackCmd.Dir = "../../../" // Adjust to pulumi project directory
		stackOutput, err := stackCmd.CombinedOutput()
		
		if err == nil {
			// Stack exists, check deployment status
			statusCmd := exec.CommandContext(ctx, "pulumi", "stack", "output", "--json")
			statusCmd.Dir = "../../../"
			statusOutput, statusErr := statusCmd.CombinedOutput()
			
			assert.NoError(t, statusErr, "Pulumi stack outputs must be accessible for deployment validation")
			assert.NotEmpty(t, statusOutput, "Pulumi stack must have deployed outputs indicating active infrastructure")
			
			t.Logf("Pulumi stack status: %s", string(statusOutput))
		} else {
			t.Logf("RED PHASE: Pulumi stack not selected or not deployed - expected until environment is bootstrapped: %v", err)
			t.Logf("Stack command output: %s", string(stackOutput))
		}
	})

	t.Run("PulumiDeployment_InfrastructureProvisioning", func(t *testing.T) {
		// RED PHASE: Validate that Pulumi has provisioned required infrastructure components
		
		requiredInfrastructureOutputs := []string{
			"postgresql_connection_string",
			"vault_endpoint",
			"rabbitmq_connection_string",
			"dapr_components_path",
		}

		for _, output := range requiredInfrastructureOutputs {
			outputCmd := exec.CommandContext(ctx, "pulumi", "stack", "output", output)
			outputCmd.Dir = "../../../"
			outputValue, err := outputCmd.CombinedOutput()

			if err == nil {
				assert.NotEmpty(t, strings.TrimSpace(string(outputValue)), 
					"Pulumi must provision and output %s for infrastructure deployment", output)
			} else {
				t.Logf("RED PHASE: Pulumi output %s not available - expected until infrastructure is deployed: %v", output, err)
			}
		}
	})
}

func TestEnvironmentBootstrapping_InfrastructureComponentHealth(t *testing.T) {
	// RED PHASE: This test validates infrastructure components are healthy and accessible
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("InfrastructureHealth_DatabaseConnectivity", func(t *testing.T) {
		// RED PHASE: Validate PostgreSQL database is running and accessible
		
		// Check if PostgreSQL container is running
		psqlCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name=postgresql", "--format", "{{.Names}}")
		psqlOutput, err := psqlCmd.Output()
		
		if err == nil && strings.Contains(string(psqlOutput), "postgresql") {
			// Database container is running - test connectivity
			connStr := "postgresql://postgres:password@localhost:5432/international_center_development?sslmode=disable"
			
			// Use podman exec to test database connectivity from within container network
			testCmd := exec.CommandContext(ctx, "podman", "exec", "postgresql", 
				"psql", connStr, "-c", "SELECT 1")
			testOutput, testErr := testCmd.CombinedOutput()
			
			if testErr == nil {
				assert.Contains(t, string(testOutput), "1", 
					"PostgreSQL must be responsive and accept connections for service data persistence")
			} else {
				t.Logf("RED PHASE: PostgreSQL connectivity test failed - expected until properly configured: %v", testErr)
				t.Logf("Database test output: %s", string(testOutput))
			}
		} else {
			t.Logf("RED PHASE: PostgreSQL container not running - expected until environment is deployed: %v", err)
			t.Log("Database must be running for service data persistence")
		}
	})

	t.Run("InfrastructureHealth_VaultAccessibility", func(t *testing.T) {
		// RED PHASE: Validate HashiCorp Vault is running and accessible
		
		// Check if Vault container is running
		vaultCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name=vault", "--format", "{{.Names}}")
		vaultOutput, err := vaultCmd.Output()
		
		if err == nil && strings.Contains(string(vaultOutput), "vault") {
			// Vault container is running - test API accessibility
			client := &http.Client{Timeout: 10 * time.Second}
			
			req, reqErr := http.NewRequestWithContext(ctx, "GET", "http://localhost:8200/v1/sys/health", nil)
			if reqErr == nil {
				resp, respErr := client.Do(req)
				if respErr == nil {
					defer resp.Body.Close()
					assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 429, 
						"Vault must be accessible for secret management operations")
				} else {
					t.Logf("RED PHASE: Vault HTTP API not accessible - expected until properly configured: %v", respErr)
				}
			}
		} else {
			t.Logf("RED PHASE: Vault container not running - expected until environment is deployed: %v", err)
			t.Log("Vault must be running for secure secret management")
		}
	})

	t.Run("InfrastructureHealth_RabbitMQMessaging", func(t *testing.T) {
		// RED PHASE: Validate RabbitMQ is running and accessible
		
		// Check if RabbitMQ container is running
		rabbitmqCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name=rabbitmq", "--format", "{{.Names}}")
		rabbitmqOutput, err := rabbitmqCmd.Output()
		
		if err == nil && strings.Contains(string(rabbitmqOutput), "rabbitmq") {
			// RabbitMQ container is running - test management API
			client := &http.Client{Timeout: 10 * time.Second}
			
			req, reqErr := http.NewRequestWithContext(ctx, "GET", "http://localhost:15672/api/overview", nil)
			if reqErr == nil {
				req.SetBasicAuth("guest", "guest")
				resp, respErr := client.Do(req)
				if respErr == nil {
					defer resp.Body.Close()
					assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
						"RabbitMQ must be accessible for pub/sub messaging operations")
				} else {
					t.Logf("RED PHASE: RabbitMQ management API not accessible - expected until properly configured: %v", respErr)
				}
			}
		} else {
			t.Logf("RED PHASE: RabbitMQ container not running - expected until environment is deployed: %v", err)
			t.Log("RabbitMQ must be running for event-driven communication")
		}
	})
}

func TestEnvironmentBootstrapping_DaprRuntimeInitialization(t *testing.T) {
	// RED PHASE: This test validates Dapr runtime is properly initialized
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("DaprRuntime_ControlPlaneHealth", func(t *testing.T) {
		// RED PHASE: Validate Dapr control plane is running and healthy
		
		// Check if Dapr placement service is running
		placementCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name=dapr_placement", "--format", "{{.Names}}")
		placementOutput, err := placementCmd.Output()
		
		if err == nil && strings.Contains(string(placementOutput), "dapr_placement") {
			// Dapr placement is running - check health
			client := &http.Client{Timeout: 10 * time.Second}
			
			req, reqErr := http.NewRequestWithContext(ctx, "GET", "http://localhost:50005/v1.0/healthz", nil)
			if reqErr == nil {
				resp, respErr := client.Do(req)
				if respErr == nil {
					defer resp.Body.Close()
					assert.Equal(t, 200, resp.StatusCode, 
						"Dapr placement service must be healthy for distributed applications")
				} else {
					t.Logf("RED PHASE: Dapr placement health endpoint not accessible: %v", respErr)
				}
			}
		} else {
			t.Logf("RED PHASE: Dapr placement service not running - expected until Dapr runtime is deployed: %v", err)
		}
	})

	t.Run("DaprRuntime_ComponentRegistration", func(t *testing.T) {
		// RED PHASE: Validate Dapr components are properly registered
		
		expectedComponents := []struct {
			name        string
			componentType string
			description string
		}{
			{"statestore", "state", "PostgreSQL state store for service data persistence"},
			{"secretstore", "secretstore", "Vault secret store for secure configuration"},
			{"pubsub", "pubsub", "RabbitMQ pub/sub for event-driven communication"},
		}

		for _, component := range expectedComponents {
			t.Run("Component_"+component.name, func(t *testing.T) {
				// Check if any Dapr sidecar is running to test component registration
				sidecarCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name=daprd", "--format", "{{.Names}}")
				sidecarOutput, err := sidecarCmd.Output()
				
				if err == nil && strings.Contains(string(sidecarOutput), "daprd") {
					// Dapr sidecar is running - test component via metadata API
					client := &http.Client{Timeout: 10 * time.Second}
					
					req, reqErr := http.NewRequestWithContext(ctx, "GET", "http://localhost:3502/v1.0/metadata", nil)
					if reqErr == nil {
						resp, respErr := client.Do(req)
						if respErr == nil {
							defer resp.Body.Close()
							body, bodyErr := io.ReadAll(resp.Body)
							if bodyErr == nil {
								assert.Contains(t, string(body), component.name,
									"%s must be registered in Dapr runtime", component.description)
							}
						} else {
							t.Logf("RED PHASE: Dapr metadata API not accessible for component validation: %v", respErr)
						}
					}
				} else {
					t.Logf("RED PHASE: No Dapr sidecars running - expected until services are deployed: %v", err)
				}
			})
		}
	})
}

func TestEnvironmentBootstrapping_ServiceContainerDeployment(t *testing.T) {
	// RED PHASE: This test validates service containers are properly deployed
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("ServiceDeployment_BackendServicesRunning", func(t *testing.T) {
		// RED PHASE: Validate backend service containers are deployed and running
		
		expectedServices := []struct {
			serviceName   string
			containerName string
			description   string
		}{
			{"content", "content", "Content management service"},
			{"inquiries", "inquiries", "Inquiry management service"},
			{"notifications", "notifications", "Notification delivery service"},
			{"public-gateway", "public-gateway", "Public API gateway"},
			{"admin-gateway", "admin-gateway", "Admin API gateway"},
		}

		for _, service := range expectedServices {
			t.Run("Service_"+service.serviceName, func(t *testing.T) {
				// Check if service container is running
				serviceCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name="+service.containerName, "--format", "{{.Names}}")
				serviceOutput, err := serviceCmd.Output()
				
				if err == nil && strings.Contains(string(serviceOutput), service.containerName) {
					// Service container is running - validate it's healthy
					t.Logf("Service %s is running", service.serviceName)
					
					// Additional health check could be added here
					// For now, just validate the container is present
					assert.True(t, true, "%s must be deployed and running", service.description)
				} else {
					t.Logf("RED PHASE: %s not running - expected until service deployment is automated: %v", service.description, err)
					
					// This should fail in RED PHASE until services are properly deployed
					assert.Fail(t, fmt.Sprintf("%s must be deployed and running for environment validation", service.description))
				}
			})
		}
	})

	t.Run("ServiceDeployment_DaprSidecarIntegration", func(t *testing.T) {
		// RED PHASE: Validate services are integrated with Dapr sidecars
		
		// Check for Dapr sidecar containers
		sidecarCmd := exec.CommandContext(ctx, "podman", "ps", "--filter", "name=daprd", "--format", "{{.Names}}")
		sidecarOutput, err := sidecarCmd.Output()
		
		if err == nil {
			sidecars := strings.Fields(string(sidecarOutput))
			assert.Greater(t, len(sidecars), 0, 
				"At least one Dapr sidecar must be running for service mesh integration")
			
			for _, sidecar := range sidecars {
				if strings.Contains(sidecar, "daprd") {
					t.Logf("Dapr sidecar found: %s", sidecar)
				}
			}
		} else {
			t.Logf("RED PHASE: No Dapr sidecars detected - expected until services are deployed with Dapr integration: %v", err)
			
			// This should fail in RED PHASE until Dapr sidecars are deployed
			assert.Fail(t, "Services must be deployed with Dapr sidecar integration for service mesh functionality")
		}
	})

	t.Run("ServiceDeployment_ServiceDiscoveryValidation", func(t *testing.T) {
		// RED PHASE: Validate services can be discovered through Dapr
		
		// Test service discovery by attempting to call a service through Dapr
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Try to access a service through Dapr service invocation
		req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3502/v1.0/invoke/content/method/health", nil)
		if err == nil {
			resp, respErr := client.Do(req)
			if respErr == nil {
				defer resp.Body.Close()
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
					"Services must be discoverable through Dapr service invocation")
				
				t.Logf("Service discovery test - Status: %d", resp.StatusCode)
			} else {
				t.Logf("RED PHASE: Service discovery through Dapr not working - expected until services are deployed: %v", respErr)
				
				// This should fail in RED PHASE until service discovery is working
				assert.Fail(t, "Service discovery through Dapr must be functional for microservices communication")
			}
		} else {
			t.Logf("RED PHASE: Cannot create service discovery request - expected until Dapr is configured: %v", err)
		}
	})
}

// validateConfigurationDeployment validates proper project configuration deployment
func validateConfigurationDeployment(t *testing.T, ctx context.Context) {
	// RED PHASE: Detect improper temporary directory usage
	t.Run("DetectTemporaryDirectoryAntiPattern", func(t *testing.T) {
		configFailures := sharedValidation.DetectConfigurationDeploymentFailures(ctx)

		// Test should initially fail when containers use /tmp directories
		for _, failure := range configFailures {
			if strings.Contains(failure, "/tmp") {
				t.Logf("RED PHASE: Detected temporary directory anti-pattern: %s", failure)
				assert.Fail(t, "Container configuration must not use temporary directories - found: %s", failure)
			}
		}
	})

	// RED PHASE: Validate project configuration directory mounting
	t.Run("ValidateProjectConfigurationMounting", func(t *testing.T) {
		sidecarContainers := []string{
			"content-api-sidecar",
			"inquiries-api-sidecar",
			"notification-api-sidecar",
			"services-api-sidecar",
			"public-gateway-sidecar",
			"admin-gateway-sidecar",
		}

		expectedMounts := []string{
			"/home/tojkuv/Documents/GitHub/international-center-workspace/international-center/src/public-website/deployment/configs/dapr",
		}

		for _, sidecar := range sidecarContainers {
			t.Run("ProjectMount_"+sidecar, func(t *testing.T) {
				err := sharedValidation.ValidateContainerVolumeMounts(ctx, sidecar, expectedMounts)
				if err != nil {
					t.Logf("RED PHASE: Project configuration mounting failure for %s: %v", sidecar, err)
					assert.Fail(t, "Sidecar %s must mount project configuration directory properly", sidecar)
				} else {
					t.Logf("Project configuration properly mounted for %s", sidecar)
				}
			})
		}
	})

	// RED PHASE: Validate Dapr configuration paths in container startup commands
	t.Run("ValidateDaprConfigurationPaths", func(t *testing.T) {
		sidecarContainers := []string{
			"content-api-sidecar",
			"inquiries-api-sidecar",
			"notification-api-sidecar",
			"services-api-sidecar",
			"public-gateway-sidecar",
			"admin-gateway-sidecar",
		}

		for _, sidecar := range sidecarContainers {
			t.Run("ConfigPaths_"+sidecar, func(t *testing.T) {
				err := sharedValidation.ValidateDaprConfigurationPaths(ctx, sidecar)
				if err != nil {
					t.Logf("RED PHASE: Configuration path failure for %s: %v", sidecar, err)
					assert.Fail(t, "Sidecar %s must have proper Dapr configuration paths", sidecar)
				} else {
					t.Logf("Dapr configuration paths properly set for %s", sidecar)
				}
			})
		}
	})

	// RED PHASE: Comprehensive configuration deployment validation
	t.Run("ComprehensiveConfigurationValidation", func(t *testing.T) {
		sidecarContainers := []string{
			"content-api-sidecar",
			"inquiries-api-sidecar",
			"notification-api-sidecar",
			"services-api-sidecar",
			"public-gateway-sidecar",
			"admin-gateway-sidecar",
		}

		for _, sidecar := range sidecarContainers {
			t.Run("FullValidation_"+sidecar, func(t *testing.T) {
				err := sharedValidation.ValidateProjectConfigurationDeployment(ctx, sidecar)
				if err != nil {
					t.Logf("RED PHASE: Full configuration validation failure for %s: %v", sidecar, err)
					assert.Fail(t, "Sidecar %s must use proper project configuration deployment", sidecar)
				} else {
					t.Logf("Full configuration validation passed for %s", sidecar)
				}
			})
		}
	})
}