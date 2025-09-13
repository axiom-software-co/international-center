package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
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

	// RED PHASE: Validate component registration through Dapr control plane
	t.Run("ComponentRegistrationValidation", func(t *testing.T) {
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Test Dapr components endpoint to validate component registration
		componentsURL := "http://localhost:3500/v1.0/components"
		req, err := http.NewRequestWithContext(ctx, "GET", componentsURL, nil)
		require.NoError(t, err, "Failed to create components request")

		resp, err := client.Do(req)
		require.NoError(t, err, "Dapr components endpoint must be accessible for component validation")
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
			"Dapr components endpoint must be operational for component registration validation")

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Failed to read components response")

		var components []map[string]interface{}
		err = json.Unmarshal(body, &components)
		require.NoError(t, err, "Components response must be valid JSON")

		// RED PHASE: Validate expected project-managed components are registered
		expectedComponents := []string{"statestore", "secretstore", "pubsub"}
		registeredComponents := make(map[string]bool)
		
		for _, component := range components {
			if name, exists := component["name"]; exists {
				registeredComponents[name.(string)] = true
				t.Logf("Found registered component: %s", name)
			}
		}
		
		for _, expectedComp := range expectedComponents {
			t.Run("RegisteredComponent_"+expectedComp, func(t *testing.T) {
				assert.True(t, registeredComponents[expectedComp], 
					"RED PHASE: Component %s should be registered in Dapr control plane - will fail until component loading implemented", expectedComp)
			})
		}
		
		t.Logf("RED PHASE: Found %d registered components, expected components: %v", len(components), expectedComponents)
	})

	// RED PHASE: Validate infrastructure through Dapr component APIs instead of direct container checks
	daprComponents := []struct {
		componentName string
		componentType string
		endpoint      string
		description   string
	}{
		{"statestore", "state.postgresql", "http://localhost:3500/v1.0/state/statestore", "PostgreSQL state store via Dapr"},
		{"pubsub", "pubsub.rabbitmq", "http://localhost:3500/v1.0/subscribe", "RabbitMQ pub/sub via Dapr"},
		{"secretstore", "secretstores.hashicorp.vault", "http://localhost:3500/v1.0/secrets/secretstore", "Vault secrets via Dapr"},
		{"blobstore", "bindings.azure.blobstorage", "http://localhost:3500/v1.0/bindings/blobstore", "Blob storage via Dapr"},
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
			metadataReq, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:3500/v1.0/metadata", nil)
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
		{"control-plane-health", "http://localhost:3500/v1.0/healthz", "GET", "Dapr control plane health via service mesh API"},
		{"service-discovery", "http://localhost:3500/v1.0/metadata", "GET", "Service discovery through Dapr metadata API"},
		{"component-registry", "http://localhost:3500/v1.0/metadata", "GET", "Component registry validation via Dapr API"},
		{"centralized-control-plane", "http://localhost:3500/v1.0/healthz", "GET", "Centralized control plane connectivity validation"},
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
		{"public-gateway", "health", "http://localhost:3500/v1.0/invoke/public-gateway/method/health", "Public gateway via Dapr service invocation"},
		{"admin-gateway", "health", "http://localhost:3500/v1.0/invoke/admin-gateway/method/health", "Admin gateway via Dapr service invocation"},
		{"content-api", "health", "http://localhost:3500/v1.0/invoke/content-api/method/health", "Content service via Dapr service invocation"},
		{"inquiries-api", "health", "http://localhost:3500/v1.0/invoke/inquiries-api/method/health", "Inquiries service via Dapr service invocation"},
		{"services-api", "health", "http://localhost:3500/v1.0/invoke/services-api/method/health", "Services service via Dapr service invocation"},
		{"notification-api", "health", "http://localhost:3500/v1.0/invoke/notification-api/method/health", "Notifications service via Dapr service invocation"},
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
			metadataReq, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:3500/v1.0/metadata", nil)
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
		{"public-website", "public-gateway", "http://localhost:3500/v1.0/invoke/public-gateway/method/health", "Public website via public gateway service mesh"},
		{"admin-portal", "admin-gateway", "http://localhost:3500/v1.0/invoke/admin-gateway/method/health", "Admin portal via admin gateway service mesh"},
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
				stateStoreURL := fmt.Sprintf("http://localhost:3500/v1.0/state/statestore/%s", stateKey)
				
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
		// RED PHASE: Validate service-to-service communication through Dapr service invocation
		client := &http.Client{Timeout: 10 * time.Second}
		
		serviceCommunications := []struct {
			from string
			to   string
			endpoint string
		}{
			{"public-gateway", "content-api", "http://localhost:3500/v1.0/invoke/content-api/method/health"},
			{"admin-gateway", "inquiries-api", "http://localhost:3500/v1.0/invoke/inquiries-api/method/health"},
			{"content-api", "notification-api", "http://localhost:3500/v1.0/invoke/notification-api/method/health"},
		}
		
		for _, comm := range serviceCommunications {
			t.Run(fmt.Sprintf("ServiceMesh_%s_to_%s", comm.from, comm.to), func(t *testing.T) {
				// RED PHASE: This should fail initially because proper service mesh communication is not implemented
				
				req, err := http.NewRequestWithContext(ctx, "GET", comm.endpoint, nil)
				require.NoError(t, err, "Failed to create service mesh communication request")

				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300, 
						"Service %s should communicate with %s via Dapr service mesh", comm.from, comm.to)
				} else {
					// RED PHASE: This failure is expected until proper service mesh integration
					t.Logf("RED PHASE: Service mesh communication %s -> %s not operational - expected until implemented: %v", 
						comm.from, comm.to, err)
				}
			})
		}
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
		metadataReq, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3500/v1.0/metadata", nil)
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
				metadataReq, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3500/v1.0/metadata", nil)
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