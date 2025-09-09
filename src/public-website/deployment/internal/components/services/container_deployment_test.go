package services

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatewayContainerDeployment_Development(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &GatewayArgs{
			Environment: "development",
			InfrastructureOutputs: pulumi.Map{
				"database_connection_string": pulumi.String("postgresql://postgres:5432/development"),
			},
			PlatformOutputs: pulumi.Map{
				"dapr_control_plane_url": pulumi.String("http://localhost:50001"),
			},
		}

		// Act
		component, err := NewGatewayComponent(ctx, "test-gateway", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate actual container deployment for public gateway
		component.PublicGatewayURL.ApplyT(func(url string) string {
			assert.Contains(t, url, "127.0.0.1:9001")
			return url
		})

		// Validate actual container deployment for admin gateway
		component.AdminGatewayURL.ApplyT(func(url string) string {
			assert.Contains(t, url, "127.0.0.1:9000")
			return url
		})

		// Validate Dapr sidecar deployment
		component.Services.ApplyT(func(services interface{}) interface{} {
			servicesMap := services.(map[string]interface{})
			
			// Validate public gateway has Dapr sidecar
			publicGateway := servicesMap["public"].(map[string]interface{})
			assert.Equal(t, "public-gateway", publicGateway["dapr_app_id"])
			
			// Validate admin gateway has Dapr sidecar
			adminGateway := servicesMap["admin"].(map[string]interface{})
			assert.Equal(t, "admin-gateway", adminGateway["dapr_app_id"])
			
			// TODO: Add validation for actual Dapr sidecar containers once implemented
			// This should validate:
			// - Dapr sidecar containers are running alongside app containers
			// - Dapr sidecar can communicate with control plane
			// - Service discovery through Dapr is working
			
			return services
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestContentServiceContainerDeployment_Development(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &ContentArgs{
			Environment: "development",
			InfrastructureOutputs: pulumi.Map{
				"database_connection_string": pulumi.String("postgresql://postgres:5432/development"),
			},
			PlatformOutputs: pulumi.Map{
				"dapr_control_plane_url": pulumi.String("http://localhost:50001"),
			},
		}

		// Act
		component, err := NewContentComponent(ctx, "test-content", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate actual container deployment for content services
		component.Services.ApplyT(func(services interface{}) interface{} {
			servicesMap := services.(map[string]interface{})
			
			// Test news service container deployment
			newsService := servicesMap["news"].(map[string]interface{})
			assert.Equal(t, "localhost/backend/content:latest", newsService["image"])
			assert.Equal(t, "content-news", newsService["container_id"])
			assert.Equal(t, int(3001), newsService["port"])
			assert.Equal(t, "content-news", newsService["dapr_app_id"])
			
			// TODO: Add actual container accessibility test
			// This should validate the container is running and accessible on port 3001
			
			// Test events service container deployment
			eventsService := servicesMap["events"].(map[string]interface{})
			assert.Equal(t, "localhost/backend/content:latest", eventsService["image"])
			assert.Equal(t, "content-events", eventsService["container_id"])
			assert.Equal(t, int(3002), eventsService["port"])
			assert.Equal(t, "content-events", eventsService["dapr_app_id"])
			
			// Test research service container deployment
			researchService := servicesMap["research"].(map[string]interface{})
			assert.Equal(t, "localhost/backend/content:latest", researchService["image"])
			assert.Equal(t, "content-research", researchService["container_id"])
			assert.Equal(t, int(3003), researchService["port"])
			assert.Equal(t, "content-research", researchService["dapr_app_id"])
			
			return services
		})

		// Validate health endpoints are configured correctly
		component.HealthEndpoints.ApplyT(func(endpoints interface{}) interface{} {
			endpointsMap := endpoints.(map[string]interface{})
			
			// Test health endpoint accessibility for each service
			newsHealthURL := endpointsMap["news"].(string)
			assert.Equal(t, "http://localhost:3001/health", newsHealthURL)
			
			eventsHealthURL := endpointsMap["events"].(string)
			assert.Equal(t, "http://localhost:3002/health", eventsHealthURL)
			
			researchHealthURL := endpointsMap["research"].(string)
			assert.Equal(t, "http://localhost:3003/health", researchHealthURL)
			
			// TODO: Add actual health endpoint accessibility tests once containers are deployed
			// This should validate:
			// - Each service container responds to health checks
			// - Health endpoints return proper status codes
			// - Health checks include dependency validation (database, Dapr)
			
			return endpoints
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestServiceContainerDeployment_Staging(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &GatewayArgs{
			Environment: "staging",
			InfrastructureOutputs: pulumi.Map{
				"database_connection_string": pulumi.String("postgresql://staging-db:5432/staging"),
			},
			PlatformOutputs: pulumi.Map{
				"dapr_control_plane_url": pulumi.String("https://dapr-staging.azurecontainerapp.io"),
			},
		}

		// Act
		component, err := NewGatewayComponent(ctx, "test-gateway", args)
		if err != nil {
			return err
		}

		// Assert - Validate Azure Container Apps deployment
		require.NotNil(t, component)

		// Validate Azure Container Apps configuration
		component.PublicGatewayURL.ApplyT(func(url string) string {
			assert.Contains(t, url, "public-gateway-staging.azurecontainerapp.io")
			assert.Contains(t, url, "https://")
			return url
		})

		component.AdminGatewayURL.ApplyT(func(url string) string {
			assert.Contains(t, url, "admin-gateway-staging.azurecontainerapp.io")
			assert.Contains(t, url, "https://")
			return url
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestServiceContainerDeployment_DaprSidecarIntegration(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &ServicesArgs{
			Environment: "development",
			InfrastructureOutputs: pulumi.Map{
				"database_connection_string": pulumi.String("postgresql://postgres:5432/development"),
			},
			PlatformOutputs: pulumi.Map{
				"dapr_control_plane_url": pulumi.String("http://localhost:50001"),
			},
		}

		// Act
		component, err := NewServicesComponent(ctx, "test-services", args)
		if err != nil {
			return err
		}

		// Assert - Validate Dapr sidecar integration
		require.NotNil(t, component)

		// This test validates that all services have proper Dapr sidecar integration
		// Currently this will fail because we don't have actual container deployment
		
		// Validate Dapr sidecar is enabled
		component.DaprSidecarEnabled.ApplyT(func(enabled bool) bool {
			assert.True(t, enabled, "Dapr sidecars should be enabled for all services")
			return enabled
		})

		// TODO: Add validation for actual Dapr sidecar deployment once implemented
		// This should validate:
		// - Each service container has a corresponding Dapr sidecar container
		// - Dapr sidecars can communicate with the control plane
		// - Service-to-service communication works through Dapr
		// - Dapr middleware components are properly configured
		// - Pub/sub, state management, and service invocation work

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestServiceContainerDeployment_ResourceLimits(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &ContentArgs{
			Environment: "development",
			InfrastructureOutputs: pulumi.Map{
				"database_connection_string": pulumi.String("postgresql://postgres:5432/development"),
			},
			PlatformOutputs: pulumi.Map{
				"dapr_control_plane_url": pulumi.String("http://localhost:50001"),
			},
		}

		// Act
		component, err := NewContentComponent(ctx, "test-content", args)
		if err != nil {
			return err
		}

		// Assert - Validate resource limits are applied
		require.NotNil(t, component)

		// Validate resource limits configuration
		component.Services.ApplyT(func(services interface{}) interface{} {
			servicesMap := services.(map[string]interface{})
			
			// Test resource limits for news service
			newsService := servicesMap["news"].(map[string]interface{})
			resourceLimits := newsService["resource_limits"].(map[string]interface{})
			assert.Equal(t, "500m", resourceLimits["cpu"])
			assert.Equal(t, "256Mi", resourceLimits["memory"])
			
			// TODO: Add validation for actual resource limits enforcement once containers are deployed
			// This should validate:
			// - Container resource limits are applied correctly
			// - Containers respect CPU and memory limits
			// - Resource usage monitoring is working
			// - Container scaling respects resource constraints
			
			return services
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

