package components

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
)

// TestBackendServicesContainerDeployment_Development validates that backend service containers are deployed and running
func TestBackendServicesContainerDeployment_Development(t *testing.T) {
	t.Run("BackendServiceContainersExist_Development", func(t *testing.T) {
		expectedBackendServices := []string{
			"inquiries-service", "content-service", "notifications-service", 
			"admin-gateway", "public-gateway",
		}
		for _, serviceName := range expectedBackendServices {
			validateBackendServiceContainerExists(t, serviceName)
		}
	})

	t.Run("BackendServiceContainersRunning_Development", func(t *testing.T) {
		// Test specific service containers with expected configurations
		validateBackendServiceContainerRunning(t, "inquiries-service", "localhost/inquiries:latest", []string{})
		validateBackendServiceContainerRunning(t, "content-service", "localhost/content:latest", []string{})
		validateBackendServiceContainerRunning(t, "notifications-service", "localhost/notifications:latest", []string{})
		validateBackendServiceContainerRunning(t, "admin-gateway", "localhost/admin-gateway:latest", []string{"9000"})
		validateBackendServiceContainerRunning(t, "public-gateway", "localhost/public-gateway:latest", []string{"9001"})
	})

	t.Run("BackendServiceContainersHealthy_Development", func(t *testing.T) {
		expectedBackendServices := []string{
			"inquiries-service", "content-service", "notifications-service", 
			"admin-gateway", "public-gateway",
		}
		for _, serviceName := range expectedBackendServices {
			validateBackendServiceContainerHealthy(t, serviceName)
		}
	})
}

// validateBackendServiceContainerExists checks if backend service container exists
func validateBackendServiceContainerExists(t *testing.T, name string) {
	cmd := exec.Command("podman", "ps", "-a", "--filter", "name="+name, "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check for backend service container %s: %v", name, err)
	}

	containerNames := strings.TrimSpace(string(output))
	assert.Contains(t, containerNames, name, "Backend service container %s should exist", name)
}

// validateBackendServiceContainerRunning checks if backend service container is running with correct image and ports
func validateBackendServiceContainerRunning(t *testing.T, name, expectedImage string, expectedPorts []string) {
	cmd := exec.Command("podman", "ps", "--filter", "name="+name, "--format", "{{.Names}}\t{{.Image}}\t{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check backend service container %s status: %v", name, err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("Backend service container %s is not running", name)
	}

	parts := strings.Split(lines[0], "\t")
	if len(parts) < 2 {
		t.Fatalf("Unexpected backend service container %s output format", name)
	}

	containerName := parts[0]
	containerImage := parts[1]
	var containerPorts string
	if len(parts) >= 3 {
		containerPorts = parts[2]
	}

	assert.Equal(t, name, containerName, "Container name should match")
	assert.Equal(t, expectedImage, containerImage, "Container should use correct backend service image")
	
	for _, port := range expectedPorts {
		assert.Contains(t, containerPorts, port, "Container should expose port %s", port)
	}
}

// validateBackendServiceContainerHealthy checks if backend service container is healthy
func validateBackendServiceContainerHealthy(t *testing.T, name string) {
	cmd := exec.Command("podman", "ps", "--filter", "name="+name, "--format", "{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check backend service container %s health: %v", name, err)
	}

	status := strings.TrimSpace(string(output))
	assert.NotEmpty(t, status, "Backend service container %s should have status", name)
	assert.Contains(t, strings.ToLower(status), "up", "Backend service container %s should be running (Up status)", name)
}

// TestBackendServiceHealthEndpoints_Development validates that backend services have proper health check endpoints
func TestBackendServiceHealthEndpoints_Development(t *testing.T) {
	serviceHealthConfigs := map[string]struct {
		port         string
		healthPath   string
		readyPath    string
		expectedKeys []string
	}{
		"inquiries-service": {
			port:         "8080",
			healthPath:   "/health",
			readyPath:    "/health/ready",
			expectedKeys: []string{"status", "timestamp", "service"},
		},
		"content-service": {
			port:         "8080",
			healthPath:   "/health",
			readyPath:    "/health/ready",
			expectedKeys: []string{"status", "timestamp", "service"},
		},
		"notifications-service": {
			port:         "8080",
			healthPath:   "/health",
			readyPath:    "/health/ready",
			expectedKeys: []string{"status", "timestamp", "service"},
		},
		"admin-gateway": {
			port:         "9000",
			healthPath:   "/health",
			readyPath:    "/health/ready",
			expectedKeys: []string{"status", "timestamp", "gateway", "services"},
		},
		"public-gateway": {
			port:         "9001",
			healthPath:   "/health",
			readyPath:    "/health/ready",
			expectedKeys: []string{"status", "timestamp", "gateway", "services"},
		},
	}

	t.Run("ServiceHealthEndpointsAccessible_Development", func(t *testing.T) {
		for serviceName, config := range serviceHealthConfigs {
			t.Run(fmt.Sprintf("HealthEndpoint_%s", serviceName), func(t *testing.T) {
				validateServiceHealthEndpoint(t, serviceName, "localhost:"+config.port, config.healthPath, config.expectedKeys)
			})
		}
	})

	t.Run("ServiceReadinessEndpointsAccessible_Development", func(t *testing.T) {
		for serviceName, config := range serviceHealthConfigs {
			t.Run(fmt.Sprintf("ReadinessEndpoint_%s", serviceName), func(t *testing.T) {
				validateServiceReadinessEndpoint(t, serviceName, "localhost:"+config.port, config.readyPath)
			})
		}
	})

	t.Run("ServiceHealthResponseValidation_Development", func(t *testing.T) {
		for serviceName, config := range serviceHealthConfigs {
			t.Run(fmt.Sprintf("HealthResponse_%s", serviceName), func(t *testing.T) {
				validateServiceHealthResponse(t, serviceName, "localhost:"+config.port, config.healthPath)
			})
		}
	})
}

// validateServiceHealthEndpoint validates service health endpoint accessibility and response structure
func validateServiceHealthEndpoint(t *testing.T, serviceName, baseURL, healthPath string, expectedKeys []string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("http://%s%s", baseURL, healthPath)
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("Failed to access health endpoint for service %s at %s: %v", serviceName, url, err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Service %s health endpoint should return 200 OK", serviceName)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Service %s health endpoint should return JSON", serviceName)

	var healthResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	assert.NoError(t, err, "Service %s health response should be valid JSON", serviceName)

	for _, key := range expectedKeys {
		assert.Contains(t, healthResponse, key, "Service %s health response should contain key '%s'", serviceName, key)
	}

	if status, ok := healthResponse["status"].(string); ok {
		assert.Equal(t, "healthy", strings.ToLower(status), "Service %s should report healthy status", serviceName)
	} else {
		t.Errorf("Service %s health response status should be a string", serviceName)
	}
}

// validateServiceReadinessEndpoint validates service readiness endpoint accessibility
func validateServiceReadinessEndpoint(t *testing.T, serviceName, baseURL, readyPath string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("http://%s%s", baseURL, readyPath)
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("Failed to access readiness endpoint for service %s at %s: %v", serviceName, url, err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Service %s readiness endpoint should return 200 OK", serviceName)
}

// validateServiceHealthResponse validates detailed health response structure and content
func validateServiceHealthResponse(t *testing.T, serviceName, baseURL, healthPath string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("http://%s%s", baseURL, healthPath)
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("Failed to get health response for service %s at %s: %v", serviceName, url, err)
	}
	defer resp.Body.Close()

	var healthResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	assert.NoError(t, err, "Service %s health response should decode successfully", serviceName)

	// Validate timestamp is recent (within last 10 seconds)
	if timestampStr, ok := healthResponse["timestamp"].(string); ok {
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err == nil {
			timeDiff := time.Since(timestamp)
			assert.True(t, timeDiff < 10*time.Second, "Service %s health timestamp should be recent", serviceName)
		}
	}

	// Validate service/gateway specific fields
	if strings.Contains(serviceName, "gateway") {
		if services, ok := healthResponse["services"].(map[string]interface{}); ok {
			assert.NotEmpty(t, services, "Gateway %s should report backend service states", serviceName)
		} else {
			t.Errorf("Gateway %s should include services health information", serviceName)
		}
	} else {
		if service, ok := healthResponse["service"].(string); ok {
			assert.Equal(t, serviceName, service, "Service %s should report correct service name", serviceName)
		} else {
			t.Errorf("Service %s should report its service name", serviceName)
		}
	}
}

// TestServiceCommunicationViaDapr_Development validates service-to-service communication through DAPR
func TestServiceCommunicationViaDapr_Development(t *testing.T) {
	daprPort := "3500"

	t.Run("DaprSidecarHealthValidation_Development", func(t *testing.T) {
		expectedDaprSidecars := []string{
			"inquiries-dapr", "content-dapr", "notifications-dapr", 
			"admin-dapr", "public-dapr",
		}
		for _, sidecarName := range expectedDaprSidecars {
			validateDaprSidecarHealth(t, sidecarName, daprPort)
		}
	})

	t.Run("ServiceDiscoveryViaDapr_Development", func(t *testing.T) {
		serviceAppIds := []string{"inquiries", "content", "notifications", "admin-gateway", "public-gateway"}
		for _, appId := range serviceAppIds {
			validateServiceDiscoveryViaDapr(t, appId, daprPort)
		}
	})

	t.Run("ServiceInvocationViaDapr_Development", func(t *testing.T) {
		// Test cross-service communication patterns
		validateServiceInvocation(t, "public-gateway", "content", "/api/v1/events", "GET", daprPort)
		validateServiceInvocation(t, "public-gateway", "inquiries", "/api/v1/business/inquiries", "POST", daprPort)
		validateServiceInvocation(t, "admin-gateway", "content", "/api/v1/admin/news", "GET", daprPort)
		validateServiceInvocation(t, "admin-gateway", "notifications", "/api/v1/admin/notifications", "GET", daprPort)
	})

	t.Run("ServiceMessagingViaDapr_Development", func(t *testing.T) {
		// Test pub/sub messaging through DAPR using RabbitMQ
		validatePubSubMessaging(t, "inquiries", "business.inquiry.created", daprPort)
		validatePubSubMessaging(t, "notifications", "notification.email.send", daprPort)
		validatePubSubMessaging(t, "content", "content.published", daprPort)
	})
}

// validateDaprSidecarHealth validates DAPR sidecar container health
func validateDaprSidecarHealth(t *testing.T, sidecarName, daprPort string) {
	cmd := exec.Command("podman", "ps", "--filter", "name="+sidecarName, "--format", "{{.Names}}\t{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check DAPR sidecar %s health: %v", sidecarName, err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("DAPR sidecar %s is not running", sidecarName)
	}

	parts := strings.Split(lines[0], "\t")
	if len(parts) < 2 {
		t.Fatalf("Unexpected DAPR sidecar %s output format", sidecarName)
	}

	containerName := parts[0]
	containerStatus := parts[1]

	assert.Equal(t, sidecarName, containerName, "DAPR sidecar name should match")
	assert.Contains(t, strings.ToLower(containerStatus), "up", "DAPR sidecar %s should be running", sidecarName)
}

// validateServiceDiscoveryViaDapr validates service discovery through DAPR service invocation
func validateServiceDiscoveryViaDapr(t *testing.T, appId, daprPort string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Use DAPR service invocation to check if service is discoverable
	url := fmt.Sprintf("http://localhost:%s/v1.0/invoke/%s/method/health", daprPort, appId)
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("Failed to discover service %s via DAPR at %s: %v", appId, url, err)
	}
	defer resp.Body.Close()

	// Service should be discoverable (even if it returns error, DAPR should route the call)
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode, "Service %s should be discoverable via DAPR", appId)
}

// validateServiceInvocation validates cross-service communication via DAPR service invocation
func validateServiceInvocation(t *testing.T, fromService, toService, endpoint, method, daprPort string) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Use DAPR service invocation for cross-service communication
	url := fmt.Sprintf("http://localhost:%s/v1.0/invoke/%s/method%s", daprPort, toService, endpoint)
	
	var req *http.Request
	var err error
	
	switch method {
	case "GET":
		req, err = http.NewRequest("GET", url, nil)
	case "POST":
		req, err = http.NewRequest("POST", url, strings.NewReader(`{"test": "data"}`))
		req.Header.Set("Content-Type", "application/json")
	default:
		t.Fatalf("Unsupported HTTP method: %s", method)
	}

	if err != nil {
		t.Fatalf("Failed to create request for %s -> %s via DAPR: %v", fromService, toService, err)
	}

	// Add DAPR headers to simulate call from source service
	req.Header.Set("dapr-app-id", fromService)
	
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to invoke %s from %s via DAPR at %s: %v", toService, fromService, url, err)
	}
	defer resp.Body.Close()

	// Verify DAPR routing works (service should be reachable even if endpoint doesn't exist)
	assert.NotEqual(t, http.StatusBadGateway, resp.StatusCode, "DAPR should successfully route from %s to %s", fromService, toService)
}

// validatePubSubMessaging validates pub/sub messaging through DAPR and RabbitMQ
func validatePubSubMessaging(t *testing.T, publisherService, topicName, daprPort string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Test message publishing via DAPR pub/sub using RabbitMQ component
	url := fmt.Sprintf("http://localhost:%s/v1.0/publish/rabbitmq-pubsub/%s", daprPort, topicName)
	
	messageData := map[string]interface{}{
		"eventType": topicName,
		"timestamp": time.Now().Format(time.RFC3339),
		"source":    publisherService,
		"data": map[string]string{
			"test": "message",
		},
	}

	messageJSON, err := json.Marshal(messageData)
	if err != nil {
		t.Fatalf("Failed to marshal test message: %v", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(messageJSON)))
	if err != nil {
		t.Fatalf("Failed to create pub/sub request for %s: %v", publisherService, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("dapr-app-id", publisherService)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to publish message from %s to topic %s via DAPR: %v", publisherService, topicName, err)
	}
	defer resp.Body.Close()

	// Message publishing should succeed (indicates DAPR and RabbitMQ integration works)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Message publishing from %s should succeed via DAPR pub/sub", publisherService)
}

// TestServicesComponent_DevelopmentEnvironment tests services component for development environment
func TestServicesComponent_DevelopmentEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployServices(ctx, cfg, "development")
		if err != nil {
			return err
		}

		// Verify development environment deploys Podman containers
		pulumi.All(outputs.DeploymentType, outputs.InquiriesServices, outputs.ContentServices, outputs.NotificationServices, outputs.GatewayServices).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			inquiriesServices := args[1].(map[string]interface{})
			contentServices := args[2].(map[string]interface{})
			notificationServices := args[3].(map[string]interface{})
			gatewayServices := args[4].(map[string]interface{})

			assert.Equal(t, "podman_containers", deploymentType, "Development should use Podman containers")
			
			// Verify consolidated inquiries service container is deployed
			assert.Contains(t, inquiriesServices, "inquiries", "Should deploy consolidated inquiries service container")
			inquiriesConfig := inquiriesServices["inquiries"].(map[string]interface{})
			assert.Contains(t, inquiriesConfig, "container_id", "Inquiries should have container_id")
			assert.Contains(t, inquiriesConfig, "container_status", "Inquiries should have container_status")
			
			// Verify consolidated content service container is deployed
			assert.Contains(t, contentServices, "content", "Should deploy consolidated content service container")
			contentConfig := contentServices["content"].(map[string]interface{})
			assert.Contains(t, contentConfig, "container_id", "Content should have container_id")
			assert.Contains(t, contentConfig, "container_status", "Content should have container_status")
			
			// Verify consolidated notifications service container is deployed
			assert.Contains(t, notificationServices, "notifications", "Should deploy consolidated notifications service container")
			notificationsConfig := notificationServices["notifications"].(map[string]interface{})
			assert.Contains(t, notificationsConfig, "container_id", "Notifications should have container_id")
			assert.Contains(t, notificationsConfig, "container_status", "Notifications should have container_status")
			
			// Verify both gateway services containers are deployed
			expectedGateways := []string{"admin", "public"}
			for _, gateway := range expectedGateways {
				assert.Contains(t, gatewayServices, gateway, "Should deploy %s gateway service container", gateway)
				gatewayConfig := gatewayServices[gateway].(map[string]interface{})
				assert.Contains(t, gatewayConfig, "container_id", "%s gateway should have container_id", gateway)
				assert.Contains(t, gatewayConfig, "container_status", "%s gateway should have container_status", gateway)
			}

			return nil
		})

		// Verify development health check configuration
		pulumi.All(outputs.HealthCheckEnabled, outputs.DaprSidecarEnabled).ApplyT(func(args []interface{}) error {
			healthCheckEnabled := args[0].(bool)
			daprSidecarEnabled := args[1].(bool)

			assert.True(t, healthCheckEnabled, "Should enable health checks for development")
			assert.True(t, daprSidecarEnabled, "Should enable Dapr sidecars for development")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &ServicesMocks{}))

	assert.NoError(t, err)
}

// TestServicesComponent_StagingEnvironment tests services component for staging environment
func TestServicesComponent_StagingEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployServices(ctx, cfg, "staging")
		if err != nil {
			return err
		}

		// Verify staging environment generates Container Apps configuration
		pulumi.All(outputs.DeploymentType, outputs.APIServices, outputs.ScalingPolicy).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			apiServices := args[1].(map[string]interface{})
			scalingPolicy := args[2].(string)

			assert.Equal(t, "container_apps", deploymentType, "Staging should use Container Apps")
			assert.NotEmpty(t, apiServices, "Should configure API services for staging")
			assert.Equal(t, "moderate", scalingPolicy, "Should use moderate scaling policy")
			return nil
		})

		// Verify staging Dapr and observability integration
		pulumi.All(outputs.DaprSidecarEnabled, outputs.ObservabilityEnabled).ApplyT(func(args []interface{}) error {
			daprSidecarEnabled := args[0].(bool)
			observabilityEnabled := args[1].(bool)

			assert.True(t, daprSidecarEnabled, "Should enable Dapr sidecars for staging")
			assert.True(t, observabilityEnabled, "Should enable observability for staging")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &ServicesMocks{}))

	assert.NoError(t, err)
}

// TestServicesComponent_ProductionEnvironment tests services component for production environment
func TestServicesComponent_ProductionEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployServices(ctx, cfg, "production")
		if err != nil {
			return err
		}

		// Verify production environment generates Container Apps with production features
		pulumi.All(outputs.DeploymentType, outputs.ScalingPolicy, outputs.SecurityPolicies).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			scalingPolicy := args[1].(string)
			securityPolicies := args[2].(bool)

			assert.Equal(t, "container_apps", deploymentType, "Production should use Container Apps")
			assert.Equal(t, "aggressive", scalingPolicy, "Should use aggressive scaling policy")
			assert.True(t, securityPolicies, "Should enable security policies for production")
			return nil
		})

		// Verify production has full observability and audit logging
		pulumi.All(outputs.ObservabilityEnabled, outputs.AuditLogging).ApplyT(func(args []interface{}) error {
			observabilityEnabled := args[0].(bool)
			auditLogging := args[1].(bool)

			assert.True(t, observabilityEnabled, "Should enable observability for production")
			assert.True(t, auditLogging, "Should enable audit logging for production")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &ServicesMocks{}))

	assert.NoError(t, err)
}

// TestServicesComponent_ContainerConfiguration tests deployed container configurations
func TestServicesComponent_ContainerConfiguration(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployServices(ctx, cfg, "development")
		if err != nil {
			return err
		}

		// Verify consolidated services containers have required deployment attributes
		pulumi.All(outputs.InquiriesServices, outputs.ContentServices, outputs.NotificationServices).ApplyT(func(args []interface{}) error {
			inquiriesServices := args[0].(map[string]interface{})
			contentServices := args[1].(map[string]interface{})
			notificationServices := args[2].(map[string]interface{})

			// Verify consolidated inquiries service has required attributes
			inquiriesConfig, ok := inquiriesServices["inquiries"].(map[string]interface{})
			assert.True(t, ok, "Consolidated inquiries service should have container configuration")
			assert.Contains(t, inquiriesConfig, "container_id", "Inquiries service should have container_id")
			assert.Contains(t, inquiriesConfig, "container_status", "Inquiries service should have container_status")
			assert.Contains(t, inquiriesConfig, "host_port", "Inquiries service should have host_port")
			assert.Contains(t, inquiriesConfig, "health_endpoint", "Inquiries service should have health_endpoint")
			assert.Contains(t, inquiriesConfig, "dapr_app_id", "Inquiries service should have Dapr app ID")
			assert.Contains(t, inquiriesConfig, "dapr_sidecar_id", "Inquiries service should have Dapr sidecar container")
			
			// Verify consolidated content service has required attributes
			contentConfig, ok := contentServices["content"].(map[string]interface{})
			assert.True(t, ok, "Consolidated content service should have container configuration")
			assert.Contains(t, contentConfig, "container_id", "Content service should have container_id")
			assert.Contains(t, contentConfig, "container_status", "Content service should have container_status")
			assert.Contains(t, contentConfig, "host_port", "Content service should have host_port")
			assert.Contains(t, contentConfig, "health_endpoint", "Content service should have health_endpoint")
			assert.Contains(t, contentConfig, "dapr_app_id", "Content service should have Dapr app ID")
			assert.Contains(t, contentConfig, "dapr_sidecar_id", "Content service should have Dapr sidecar container")
			
			// Verify consolidated notifications service has required attributes
			notificationsConfig, ok := notificationServices["notifications"].(map[string]interface{})
			assert.True(t, ok, "Consolidated notifications service should have container configuration")
			assert.Contains(t, notificationsConfig, "container_id", "Notifications service should have container_id")
			assert.Contains(t, notificationsConfig, "container_status", "Notifications service should have container_status")
			assert.Contains(t, notificationsConfig, "host_port", "Notifications service should have host_port")
			assert.Contains(t, notificationsConfig, "health_endpoint", "Notifications service should have health_endpoint")
			assert.Contains(t, notificationsConfig, "dapr_app_id", "Notifications service should have Dapr app ID")
			assert.Contains(t, notificationsConfig, "dapr_sidecar_id", "Notifications service should have Dapr sidecar container")
			
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &ServicesMocks{}))

	assert.NoError(t, err)
}

// TestServicesComponent_EnvironmentParity tests that all environments support required features
func TestServicesComponent_EnvironmentParity(t *testing.T) {
	environments := []string{"development", "staging", "production"}

	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				cfg := config.New(ctx, "")

				outputs, err := DeployServices(ctx, cfg, env)
				if err != nil {
					return err
				}

				// Verify all environments provide required outputs
				pulumi.All(outputs.DeploymentType, outputs.InquiriesServices, outputs.ContentServices, outputs.NotificationServices, outputs.GatewayServices, outputs.APIServices).ApplyT(func(args []interface{}) error {
					deploymentType := args[0].(string)
					inquiriesServices := args[1].(map[string]interface{})
					contentServices := args[2].(map[string]interface{})
					notificationServices := args[3].(map[string]interface{})
					gatewayServices := args[4].(map[string]interface{})
					apiServices := args[5].(map[string]interface{})

					assert.NotEmpty(t, deploymentType, "All environments should provide deployment type")
					assert.NotEmpty(t, gatewayServices, "All environments should provide gateway services")
					
					// Different architectures for different environments
					if env == "development" {
						assert.NotEmpty(t, inquiriesServices, "Development should provide inquiries services")
						assert.NotEmpty(t, contentServices, "Development should provide content services")
						assert.NotEmpty(t, notificationServices, "Development should provide notification services")
					} else {
						// Staging and production use APIServices instead of InquiriesServices/ContentServices/NotificationServices
						assert.NotEmpty(t, apiServices, "Staging and production should provide API services")
					}
					return nil
				})

				return nil
			}, pulumi.WithMocks("test", "stack", &ServicesMocks{}))

			assert.NoError(t, err)
		})
	}
}

// ServicesMocks provides mocks for Pulumi testing
type ServicesMocks struct{}

func (mocks *ServicesMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}

	switch args.TypeToken {
	case "docker:index/container:Container":
		// Mock docker container for development
		outputs["name"] = resource.NewStringProperty(args.Name + "-dev")
		outputs["image"] = resource.NewStringProperty("backend/" + args.Name + ":latest")
		outputs["ports"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{
				"internal": resource.NewNumberProperty(8080),
				"external": resource.NewNumberProperty(8080),
			}),
		})
		outputs["env"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty("DAPR_HTTP_PORT=3500"),
			resource.NewStringProperty("DAPR_GRPC_PORT=50001"),
		})

	case "azure-native:app:ContainerApp":
		// Mock Azure Container App for staging/production
		outputs["name"] = resource.NewStringProperty("international-center-" + args.Name)
		outputs["configuration"] = resource.NewObjectProperty(resource.PropertyMap{
			"dapr": resource.NewObjectProperty(resource.PropertyMap{
				"enabled": resource.NewBoolProperty(true),
				"appId":   resource.NewStringProperty(args.Name),
			}),
		})
		outputs["template"] = resource.NewObjectProperty(resource.PropertyMap{
			"containers": resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewObjectProperty(resource.PropertyMap{
					"image": resource.NewStringProperty("backend/" + args.Name + ":latest"),
					"name":  resource.NewStringProperty(args.Name),
					"resources": resource.NewObjectProperty(resource.PropertyMap{
						"cpu":    resource.NewNumberProperty(0.25),
						"memory": resource.NewStringProperty("0.5Gi"),
					}),
				}),
			}),
		})

	case "azure-native:operationalinsights:Workspace":
		// Mock Log Analytics Workspace for observability
		outputs["customerId"] = resource.NewStringProperty("workspace-customer-id")
		outputs["name"] = resource.NewStringProperty("international-center-logs")
	}

	return args.Name + "_id", outputs, nil
}

func (mocks *ServicesMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}
	return outputs, nil
}