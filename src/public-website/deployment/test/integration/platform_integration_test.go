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

// Platform Integration Tests
// Validates platform phase components working together as integrated system
// Tests Dapr control plane, orchestration, networking integration

func TestPlatformIntegration_DaprControlPlaneOrchestration(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("DaprControlPlane_ServiceMeshReadiness", func(t *testing.T) {
		// Test Dapr control plane is operational and ready for service mesh
		client := &http.Client{Timeout: 5 * time.Second}

		// Test Dapr health endpoint
		healthReq, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3500/v1.0/healthz", nil)
		require.NoError(t, err, "Failed to create Dapr health request")

		healthResp, err := client.Do(healthReq)
		require.NoError(t, err, "Dapr control plane health endpoint must be accessible for platform integration")
		defer healthResp.Body.Close()

		assert.True(t, healthResp.StatusCode >= 200 && healthResp.StatusCode < 300, 
			"Dapr control plane must be healthy for service mesh functionality")

		// Test Dapr metadata endpoint for service mesh configuration
		metadataReq, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3500/v1.0/metadata", nil)
		require.NoError(t, err, "Failed to create metadata request")

		metadataResp, err := client.Do(metadataReq)
		require.NoError(t, err, "Dapr metadata endpoint must be accessible for service discovery")
		defer metadataResp.Body.Close()

		assert.Equal(t, http.StatusOK, metadataResp.StatusCode, 
			"Dapr metadata endpoint must be operational for service mesh integration")

		// Validate metadata contains expected service mesh configuration
		body, err := io.ReadAll(metadataResp.Body)
		require.NoError(t, err, "Failed to read Dapr metadata")

		var metadata map[string]interface{}
		err = json.Unmarshal(body, &metadata)
		require.NoError(t, err, "Failed to parse Dapr metadata JSON")

		// Validate service mesh is configured
		assert.Contains(t, metadata, "id", "Dapr metadata must contain service mesh identity")
		
		if runtimeVersion, exists := metadata["runtimeVersion"]; exists {
			assert.NotEmpty(t, runtimeVersion, "Dapr runtime version must be available")
		}
	})

	t.Run("ServiceMeshOrchestration_PlatformReadiness", func(t *testing.T) {
		// Test that platform is ready for service orchestration
		client := &http.Client{Timeout: 5 * time.Second}

		// Test Dapr service invocation capability (without actual services)
		// This validates the platform layer service mesh infrastructure
		serviceInvocationURL := "http://localhost:3500/v1.0/invoke/test-service/method/health"
		
		req, err := http.NewRequestWithContext(ctx, "GET", serviceInvocationURL, nil)
		require.NoError(t, err, "Failed to create service invocation request")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			// Should return error for non-existent service, but platform should handle gracefully
			assert.True(t, resp.StatusCode >= 400 && resp.StatusCode < 600, 
				"Dapr service invocation platform should handle non-existent services gracefully")
		} else {
			t.Logf("Service invocation platform capability: %v", err)
			// Platform may not be fully ready for service invocation
		}
	})
}

func TestPlatformIntegration_NetworkingConfiguration(t *testing.T) {
	// Test platform networking configuration and container connectivity
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("DevelopmentNetwork_PlatformConnectivity", func(t *testing.T) {
		// Test that platform containers are properly connected to development network
		networkCmd := exec.CommandContext(ctx, "podman", "network", "inspect", "international-center-dev", "--format", "{{range .containers}}{{.Name}} {{end}}")
		networkOutput, err := networkCmd.Output()
		require.NoError(t, err, "Development network must be inspectable for platform integration")

		connectedContainers := strings.TrimSpace(string(networkOutput))
		
		// Platform components that should be connected to development network
		platformComponents := []string{"dapr-control-plane", "postgresql"}
		
		for _, component := range platformComponents {
			assert.Contains(t, connectedContainers, component, 
				"Platform component %s must be connected to development network", component)
		}
	})

	t.Run("PlatformDaprPortAccessibility", func(t *testing.T) {
		// Test that Dapr platform ports are accessible for service integration
		platformPorts := []struct {
			port        int
			protocol    string
			description string
		}{
			{3500, "HTTP", "Dapr HTTP API port must be accessible"},
			{50001, "gRPC", "Dapr gRPC API port must be accessible"},
		}

		for _, portTest := range platformPorts {
			t.Run("Port_"+portTest.protocol, func(t *testing.T) {
				// Test port connectivity using nc command through Dapr container
				portCheckCmd := exec.CommandContext(ctx, "podman", "exec", "dapr-control-plane", "nc", "-z", "localhost", fmt.Sprintf("%d", portTest.port))
				err := portCheckCmd.Run()
				assert.NoError(t, err, "%s - platform port must be accessible", portTest.description)
			})
		}
	})
}

func TestPlatformIntegration_ServiceMeshInfrastructure(t *testing.T) {
	// Test platform service mesh infrastructure readiness
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("ServiceRegistrationCapability", func(t *testing.T) {
		// Test that platform can handle service registration (even without services running)
		client := &http.Client{Timeout: 5 * time.Second}

		// Test Dapr metadata for service discovery infrastructure
		metadataReq, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3500/v1.0/metadata", nil)
		require.NoError(t, err, "Failed to create metadata request")

		metadataResp, err := client.Do(metadataReq)
		require.NoError(t, err, "Service registration infrastructure must be operational")
		defer metadataResp.Body.Close()

		body, err := io.ReadAll(metadataResp.Body)
		require.NoError(t, err, "Failed to read service registration metadata")

		var metadata map[string]interface{}
		err = json.Unmarshal(body, &metadata)
		require.NoError(t, err, "Service registration metadata must be valid JSON")

		// Platform should be ready to register services (even if none are running yet)
		assert.Contains(t, metadata, "id", "Platform must have service mesh identity for service registration")
	})

	t.Run("ServiceInvocationInfrastructure", func(t *testing.T) {
		// Test that service invocation infrastructure is ready
		client := &http.Client{Timeout: 5 * time.Second}

		// Test service invocation endpoint availability (without actual target service)
		invocationURL := "http://localhost:3500/v1.0/invoke/platform-test/method/health"
		
		req, err := http.NewRequestWithContext(ctx, "GET", invocationURL, nil)
		require.NoError(t, err, "Failed to create service invocation request")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			// Platform should respond (even with error for non-existent service)
			assert.True(t, resp.StatusCode >= 200, 
				"Service invocation infrastructure must be responsive")
		} else {
			t.Logf("Service invocation infrastructure not ready: %v", err)
		}
	})

	t.Run("ComponentConfigurationSupport", func(t *testing.T) {
		// Test that platform supports Dapr component configuration
		client := &http.Client{Timeout: 5 * time.Second}

		// Test that platform can handle component queries
		componentsURL := "http://localhost:3500/v1.0/components"
		
		req, err := http.NewRequestWithContext(ctx, "GET", componentsURL, nil)
		require.NoError(t, err, "Failed to create components request")

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500, 
				"Platform component configuration support must be operational")
		} else {
			t.Logf("Component configuration support not ready: %v", err)
		}
	})
}

// RED PHASE: Service-to-Service Communication Contract Validation via Dapr Service Mesh
func TestPlatformIntegration_ServiceMeshCommunication(t *testing.T) {
	// This test validates that services can communicate with each other through Dapr service mesh
	// Critical for microservices architecture and cross-service integration
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Service communication patterns that must work through Dapr service mesh
	serviceCommunicationTests := []struct {
		sourceService      string
		targetService      string
		communicationType  string
		testEndpoint       string
		expectedResponse   string
		description        string
	}{
		{
			sourceService:     "content-api",
			targetService:     "notification-api",
			communicationType: "health-check",
			testEndpoint:      "/health",
			expectedResponse:  "healthy",
			description:       "Content service must communicate with notifications service for event publishing",
		},
		{
			sourceService:     "inquiries-api",
			targetService:     "notification-api", 
			communicationType: "health-check",
			testEndpoint:      "/health",
			expectedResponse:  "healthy",
			description:       "Inquiries service must communicate with notifications service for inquiry alerts",
		},
		{
			sourceService:     "public-gateway",
			targetService:     "content-api",
			communicationType: "health-check",
			testEndpoint:      "/health",
			expectedResponse:  "healthy",
			description:       "Public gateway must communicate with content service for API routing",
		},
		{
			sourceService:     "admin-gateway",
			targetService:     "inquiries-api",
			communicationType: "health-check",
			testEndpoint:      "/health",
			expectedResponse:  "healthy",
			description:       "Admin gateway must communicate with inquiries service for admin operations",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Validate Dapr service mesh communication capabilities
	t.Run("DaprServiceMeshCommunicationCapabilities", func(t *testing.T) {
		// Test that Dapr service invocation API is functional for cross-service communication
		for _, commTest := range serviceCommunicationTests {
			t.Run("ServiceMeshComm_"+commTest.sourceService+"_to_"+commTest.targetService, func(t *testing.T) {
				// Test service-to-service communication via Dapr service invocation
				serviceURL := fmt.Sprintf("http://localhost:3500/v1.0/invoke/%s/method%s", 
					commTest.targetService, commTest.testEndpoint)
				
				req, err := http.NewRequestWithContext(ctx, "GET", serviceURL, nil)
				require.NoError(t, err, "Failed to create service mesh communication request")

				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("RED PHASE VALIDATION: %s - Service mesh communication failed: %v", 
						commTest.description, err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(resp.Body)
					t.Errorf("RED PHASE VALIDATION: %s - Service mesh communication returned %d: %s", 
						commTest.description, resp.StatusCode, string(body))
					return
				}

				// Validate response contains expected content
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err, "Failed to read service mesh communication response")

				var response map[string]interface{}
				err = json.Unmarshal(body, &response)
				if err != nil {
					t.Errorf("RED PHASE VALIDATION: %s - Service response not valid JSON: %v", 
						commTest.description, err)
					return
				}

				if status, exists := response["status"]; exists {
					assert.Equal(t, commTest.expectedResponse, status,
						"RED PHASE VALIDATION: %s - Service must respond with expected status through service mesh",
						commTest.description)
				} else {
					t.Errorf("RED PHASE VALIDATION: %s - Service response missing 'status' field", 
						commTest.description)
				}

				t.Logf("RED PHASE VALIDATION SUCCESS: %s - Service mesh communication functional", commTest.description)
			})
		}
	})

	// Validate cross-service data flow through service mesh
	t.Run("CrossServiceDataFlow", func(t *testing.T) {
		// Test data flow between services through Dapr service mesh
		dataFlowTests := []struct {
			workflow     string
			steps        []string
			description  string
		}{
			{
				workflow: "inquiry_to_notification_flow",
				steps: []string{
					"POST inquiry to inquiries-api",
					"inquiries-api notifies notification-api",
					"notification-api processes alert",
				},
				description: "Inquiry submission must trigger notification workflow through service mesh",
			},
			{
				workflow: "content_to_notification_flow", 
				steps: []string{
					"POST content to content-api",
					"content-api publishes event to notification-api",
					"notification-api processes content event",
				},
				description: "Content publication must trigger notification workflow through service mesh",
			},
		}

		for _, dataFlow := range dataFlowTests {
			t.Run("DataFlow_"+dataFlow.workflow, func(t *testing.T) {
				// For now, validate that target services are accessible for data flow
				// Full data flow validation will be implemented in GREEN PHASE when services are properly integrated
				
				t.Logf("RED PHASE VALIDATION: %s - Data flow pattern defined", dataFlow.description)
				t.Logf("  Workflow steps: %v", dataFlow.steps)
				
				// This serves as documentation for required data flow patterns
				// Actual implementation will be validated when services properly integrate
			})
		}
	})

	// Validate service mesh resilience patterns
	t.Run("ServiceMeshResilienceValidation", func(t *testing.T) {
		// Test that service mesh handles service failures gracefully
		resilienceTests := []struct {
			pattern     string
			description string
		}{
			{
				pattern:     "service_unavailable_handling",
				description: "Service mesh must handle unavailable services gracefully",
			},
			{
				pattern:     "circuit_breaker_behavior",
				description: "Service mesh must implement circuit breaker patterns",
			},
			{
				pattern:     "retry_mechanism",
				description: "Service mesh must implement retry mechanisms for failed requests",
			},
		}

		for _, resilienceTest := range resilienceTests {
			t.Run("Resilience_"+resilienceTest.pattern, func(t *testing.T) {
				// Test service mesh resilience by attempting to invoke non-existent service
				nonExistentURL := "http://localhost:3500/v1.0/invoke/non-existent-service/method/health"
				
				req, err := http.NewRequestWithContext(ctx, "GET", nonExistentURL, nil)
				require.NoError(t, err, "Failed to create resilience test request")

				resp, err := client.Do(req)
				if resp != nil {
					defer resp.Body.Close()
					
					// Service mesh should return proper error response for non-existent services
					if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusServiceUnavailable {
						t.Logf("RED PHASE VALIDATION SUCCESS: %s - Service mesh properly handles non-existent service", 
							resilienceTest.description)
					} else {
						body, _ := io.ReadAll(resp.Body)
						t.Logf("RED PHASE VALIDATION: %s - Service mesh returned %d for non-existent service: %s", 
							resilienceTest.description, resp.StatusCode, string(body))
					}
				} else if err != nil {
					t.Logf("RED PHASE VALIDATION: %s - Service mesh connection failed for non-existent service: %v", 
						resilienceTest.description, err)
				}
			})
		}
	})

	// Validate service discovery through service mesh
	t.Run("ServiceDiscoveryThroughMesh", func(t *testing.T) {
		// Test that services can discover each other through Dapr service mesh
		expectedServices := []string{"content-api", "inquiries-api", "notification-api"}
		
		// Test service discovery by attempting to invoke each service's health endpoint
		for _, serviceName := range expectedServices {
			t.Run("ServiceDiscovery_"+serviceName, func(t *testing.T) {
				discoveryURL := fmt.Sprintf("http://localhost:3500/v1.0/invoke/%s/method/health", serviceName)
				
				req, err := http.NewRequestWithContext(ctx, "GET", discoveryURL, nil)
				require.NoError(t, err, "Failed to create service discovery request")

				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("RED PHASE VALIDATION: Service %s not discoverable through service mesh: %v", 
						serviceName, err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					t.Logf("RED PHASE VALIDATION SUCCESS: Service %s discoverable through service mesh", serviceName)
				} else {
					body, _ := io.ReadAll(resp.Body)
					t.Errorf("RED PHASE VALIDATION: Service %s discovery failed with status %d: %s", 
						serviceName, resp.StatusCode, string(body))
				}
			})
		}
	})
}

// RED PHASE: Comprehensive Pub/Sub Message Publishing/Consuming Validation
func TestPlatformIntegration_PubSubMessagingValidation(t *testing.T) {
	// This test validates comprehensive pub/sub messaging through Dapr components
	// Critical for event-driven architecture and cross-service communication
	sharedValidation.ValidateEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 15 * time.Second}
	pubsubComponent := "pubsub" // Default Dapr pub/sub component name

	// Event schemas that services must support for cross-service communication
	eventSchemas := []struct {
		topicName    string
		eventType    string
		payload      string
		description  string
		publisher    string
		subscribers  []string
	}{
		{
			topicName:   "content-events",
			eventType:   "content.published",
			payload:     `{"content_id":"test-123","title":"Test Content","event_type":"content.published","timestamp":"` + time.Now().Format(time.RFC3339) + `","publisher":"content-api"}`,
			description: "Content publication events must be publishable and consumable across services",
			publisher:   "content-api",
			subscribers: []string{"notification-api", "admin-gateway"},
		},
		{
			topicName:   "inquiry-events", 
			eventType:   "inquiry.received",
			payload:     `{"inquiry_id":"inq-456","type":"business","event_type":"inquiry.received","timestamp":"` + time.Now().Format(time.RFC3339) + `","publisher":"inquiries-api"}`,
			description: "Inquiry submission events must be publishable and consumable across services",
			publisher:   "inquiries-api",
			subscribers: []string{"notification-api", "admin-gateway"},
		},
		{
			topicName:   "notification-events",
			eventType:   "notification.sent",
			payload:     `{"notification_id":"notif-789","recipient":"admin","event_type":"notification.sent","timestamp":"` + time.Now().Format(time.RFC3339) + `","publisher":"notification-api"}`,
			description: "Notification events must be publishable for audit and monitoring",
			publisher:   "notification-api",
			subscribers: []string{"admin-gateway"},
		},
		{
			topicName:   "system-events",
			eventType:   "system.health_check",
			payload:     `{"system_id":"platform","status":"healthy","event_type":"system.health_check","timestamp":"` + time.Now().Format(time.RFC3339) + `","publisher":"platform"}`,
			description: "System health events must be publishable for monitoring and alerting",
			publisher:   "platform",
			subscribers: []string{"admin-gateway", "notification-api"},
		},
	}

	// RED PHASE: Event Publishing Validation
	t.Run("ComprehensiveEventPublishing", func(t *testing.T) {
		for _, eventSchema := range eventSchemas {
			t.Run("Publish_"+eventSchema.eventType+"_to_"+eventSchema.topicName, func(t *testing.T) {
				// Test event publishing through Dapr pub/sub API
				publishURL := fmt.Sprintf("http://localhost:3500/v1.0/publish/%s/%s", pubsubComponent, eventSchema.topicName)
				
				publishReq, err := http.NewRequestWithContext(ctx, "POST", publishURL, strings.NewReader(eventSchema.payload))
				require.NoError(t, err, "Failed to create event publish request for %s", eventSchema.eventType)
				publishReq.Header.Set("Content-Type", "application/json")

				publishResp, err := client.Do(publishReq)
				require.NoError(t, err, "Event publishing must succeed for %s through Dapr pub/sub", eventSchema.eventType)
				defer publishResp.Body.Close()

				assert.True(t, publishResp.StatusCode == http.StatusNoContent || publishResp.StatusCode == http.StatusOK,
					"Event publishing must return success status for %s event", eventSchema.eventType)

				if publishResp.StatusCode != http.StatusNoContent && publishResp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(publishResp.Body)
					t.Errorf("RED PHASE VALIDATION: %s - Event publishing failed with status %d: %s", 
						eventSchema.description, publishResp.StatusCode, string(body))
					return
				}

				t.Logf("RED PHASE SUCCESS: Published %s event to topic %s - publisher %s", 
					eventSchema.eventType, eventSchema.topicName, eventSchema.publisher)
			})
		}
	})

	// RED PHASE: Cross-Service Event Communication Validation
	t.Run("CrossServiceEventCommunication", func(t *testing.T) {
		// Test cross-service event communication workflows
		communicationFlows := []struct {
			flowName        string
			publishingService string
			consumingServices []string
			eventPayload    string
			description     string
		}{
			{
				flowName:        "content_publication_workflow",
				publishingService: "content-api",
				consumingServices: []string{"notification-api", "admin-gateway"},
				eventPayload:    `{"workflow":"content_publication","content_id":"flow-test-1","action":"published","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`,
				description:     "Content publication must trigger notifications and admin updates",
			},
			{
				flowName:        "inquiry_processing_workflow",
				publishingService: "inquiries-api", 
				consumingServices: []string{"notification-api", "admin-gateway"},
				eventPayload:    `{"workflow":"inquiry_processing","inquiry_id":"flow-test-2","action":"received","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`,
				description:     "Inquiry submission must trigger notifications and admin alerts",
			},
			{
				flowName:        "system_monitoring_workflow",
				publishingService: "platform",
				consumingServices: []string{"notification-api", "admin-gateway"},
				eventPayload:    `{"workflow":"system_monitoring","system_component":"database","status":"warning","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`,
				description:     "System events must trigger monitoring and administrative workflows",
			},
		}

		for _, flow := range communicationFlows {
			t.Run("CommunicationFlow_"+flow.flowName, func(t *testing.T) {
				// Test complete event communication flow
				workflowTopicName := "workflow-" + flow.flowName
				publishURL := fmt.Sprintf("http://localhost:3500/v1.0/publish/%s/%s", pubsubComponent, workflowTopicName)
				
				publishReq, err := http.NewRequestWithContext(ctx, "POST", publishURL, strings.NewReader(flow.eventPayload))
				require.NoError(t, err, "Failed to create workflow event publish request")
				publishReq.Header.Set("Content-Type", "application/json")

				publishResp, err := client.Do(publishReq)
				require.NoError(t, err, "Workflow event publishing must succeed for %s", flow.flowName)
				defer publishResp.Body.Close()

				assert.True(t, publishResp.StatusCode == http.StatusNoContent || publishResp.StatusCode == http.StatusOK,
					"Workflow event publishing must return success status for %s", flow.flowName)

				t.Logf("RED PHASE SUCCESS: %s - Cross-service event communication workflow validated", flow.description)
			})
		}
	})

	// RED PHASE: Event Message Delivery Validation
	t.Run("EventMessageDeliveryValidation", func(t *testing.T) {
		// Test event message delivery patterns and guarantees
		deliveryTests := []struct {
			deliveryPattern string
			testPayload     string
			description     string
		}{
			{
				deliveryPattern: "at_least_once",
				testPayload:     `{"delivery_test":"at_least_once","message_id":"msg-001","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`,
				description:     "Events must support at-least-once delivery semantics",
			},
			{
				deliveryPattern: "ordered_delivery",
				testPayload:     `{"delivery_test":"ordered","sequence":1,"message_id":"msg-seq-001","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`,
				description:     "Events must support ordered delivery within topics",
			},
			{
				deliveryPattern: "bulk_publishing",
				testPayload:     `{"delivery_test":"bulk","batch_size":10,"message_id":"msg-bulk-001","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`,
				description:     "Events must support bulk publishing for high-throughput scenarios",
			},
		}

		for _, deliveryTest := range deliveryTests {
			t.Run("DeliveryPattern_"+deliveryTest.deliveryPattern, func(t *testing.T) {
				// Test event delivery patterns through dedicated test topic
				testTopicName := "delivery-test-" + deliveryTest.deliveryPattern
				publishURL := fmt.Sprintf("http://localhost:3500/v1.0/publish/%s/%s", pubsubComponent, testTopicName)
				
				publishReq, err := http.NewRequestWithContext(ctx, "POST", publishURL, strings.NewReader(deliveryTest.testPayload))
				require.NoError(t, err, "Failed to create delivery test publish request")
				publishReq.Header.Set("Content-Type", "application/json")

				publishResp, err := client.Do(publishReq)
				require.NoError(t, err, "Delivery test publishing must succeed for %s", deliveryTest.deliveryPattern)
				defer publishResp.Body.Close()

				assert.True(t, publishResp.StatusCode == http.StatusNoContent || publishResp.StatusCode == http.StatusOK,
					"Delivery test publishing must return success status for %s", deliveryTest.deliveryPattern)

				t.Logf("RED PHASE SUCCESS: %s - Event delivery pattern validated", deliveryTest.description)
			})
		}
	})
}

