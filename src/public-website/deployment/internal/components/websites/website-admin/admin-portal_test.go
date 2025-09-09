package admin

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminPortalComponent_Development(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &AdminPortalArgs{
			Environment: "development",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs: pulumi.Map{},
			ServicesOutputs: pulumi.Map{
				"admin_gateway_url": pulumi.String("http://localhost:9002"),
			},
		}

		// Act
		component, err := NewAdminPortalComponent(ctx, "test-admin-portal", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate outputs are defined
		assert.NotNil(t, component.AdminPortalURL)
		assert.NotNil(t, component.DeploymentType)
		assert.NotNil(t, component.CDNEnabled)
		assert.NotNil(t, component.SSLEnabled)
		assert.NotNil(t, component.CacheConfiguration)
		assert.NotNil(t, component.HealthCheckEnabled)
		assert.NotNil(t, component.ContainerConfig)
		assert.NotNil(t, component.StaticAssets)

		// Validate development-specific configurations
		component.AdminPortalURL.ApplyT(func(url string) string {
			assert.Contains(t, url, "localhost")
			assert.Contains(t, url, "3001")
			return url
		})

		component.DeploymentType.ApplyT(func(deploymentType string) string {
			assert.Equal(t, "podman_container", deploymentType)
			return deploymentType
		})

		component.CDNEnabled.ApplyT(func(enabled bool) bool {
			assert.False(t, enabled)
			return enabled
		})

		component.SSLEnabled.ApplyT(func(enabled bool) bool {
			assert.False(t, enabled)
			return enabled
		})

		component.HealthCheckEnabled.ApplyT(func(enabled bool) bool {
			assert.True(t, enabled)
			return enabled
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestAdminPortalComponent_Staging(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &AdminPortalArgs{
			Environment: "staging",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs: pulumi.Map{},
			ServicesOutputs: pulumi.Map{
				"admin_gateway_url": pulumi.String("https://admin-gateway-staging.azurecontainerapp.io"),
			},
		}

		// Act
		component, err := NewAdminPortalComponent(ctx, "test-admin-portal", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate outputs are defined
		assert.NotNil(t, component.AdminPortalURL)
		assert.NotNil(t, component.DeploymentType)
		assert.NotNil(t, component.CDNEnabled)
		assert.NotNil(t, component.SSLEnabled)

		// Validate staging-specific configurations
		component.AdminPortalURL.ApplyT(func(url string) string {
			assert.Contains(t, url, "staging")
			assert.Contains(t, url, "azurecontainerapp.io")
			return url
		})

		component.DeploymentType.ApplyT(func(deploymentType string) string {
			assert.Equal(t, "container_app", deploymentType)
			return deploymentType
		})

		component.CDNEnabled.ApplyT(func(enabled bool) bool {
			assert.True(t, enabled)
			return enabled
		})

		component.SSLEnabled.ApplyT(func(enabled bool) bool {
			assert.True(t, enabled)
			return enabled
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestAdminPortalComponent_Production(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &AdminPortalArgs{
			Environment: "production",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs: pulumi.Map{},
			ServicesOutputs: pulumi.Map{
				"admin_gateway_url": pulumi.String("https://admin-gateway-production.azurecontainerapp.io"),
			},
		}

		// Act
		component, err := NewAdminPortalComponent(ctx, "test-admin-portal", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate outputs are defined
		assert.NotNil(t, component.AdminPortalURL)
		assert.NotNil(t, component.DeploymentType)
		assert.NotNil(t, component.CDNEnabled)
		assert.NotNil(t, component.SSLEnabled)

		// Validate production-specific configurations
		component.AdminPortalURL.ApplyT(func(url string) string {
			assert.Contains(t, url, "production")
			assert.Contains(t, url, "azurecontainerapp.io")
			return url
		})

		component.DeploymentType.ApplyT(func(deploymentType string) string {
			assert.Equal(t, "container_app", deploymentType)
			return deploymentType
		})

		component.CDNEnabled.ApplyT(func(enabled bool) bool {
			assert.True(t, enabled)
			return enabled
		})

		component.SSLEnabled.ApplyT(func(enabled bool) bool {
			assert.True(t, enabled)
			return enabled
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestAdminPortalComponent_ContainerConfiguration(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &AdminPortalArgs{
			Environment: "development",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs: pulumi.Map{},
			ServicesOutputs: pulumi.Map{
				"admin_gateway_url": pulumi.String("http://localhost:9002"),
			},
		}

		// Act
		component, err := NewAdminPortalComponent(ctx, "test-admin-portal", args)
		if err != nil {
			return err
		}

		// Assert - Validate container configurations
		require.NotNil(t, component.ContainerConfig)

		component.ContainerConfig.ApplyT(func(config interface{}) interface{} {
			configMap := config.(map[string]interface{})
			assert.NotNil(t, configMap["image"])
			assert.NotNil(t, configMap["container_id"])
			assert.NotNil(t, configMap["port"])
			assert.NotNil(t, configMap["health_check"])
			assert.NotNil(t, configMap["environment_variables"])
			assert.Contains(t, configMap["container_id"], "admin-portal")
			assert.Equal(t, configMap["health_check"], "/health")
			return config
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestAdminPortalComponent_CacheConfiguration(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &AdminPortalArgs{
			Environment: "production",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs: pulumi.Map{},
			ServicesOutputs: pulumi.Map{
				"admin_gateway_url": pulumi.String("https://admin-gateway-production.azurecontainerapp.io"),
			},
		}

		// Act
		component, err := NewAdminPortalComponent(ctx, "test-admin-portal", args)
		if err != nil {
			return err
		}

		// Assert - Validate cache configurations
		require.NotNil(t, component.CacheConfiguration)

		component.CacheConfiguration.ApplyT(func(config interface{}) interface{} {
			configMap := config.(map[string]interface{})
			assert.NotNil(t, configMap["enabled"])
			assert.NotNil(t, configMap["static_content_ttl"])
			assert.NotNil(t, configMap["compression"])
			
			// Validate cache is enabled for production
			assert.Equal(t, true, configMap["enabled"])
			return config
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestAdminPortalComponent_StaticAssets(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange - Test static assets configuration
		args := &AdminPortalArgs{
			Environment: "production",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs: pulumi.Map{},
			ServicesOutputs: pulumi.Map{
				"admin_gateway_url": pulumi.String("https://admin-gateway-production.azurecontainerapp.io"),
			},
		}

		// Act
		component, err := NewAdminPortalComponent(ctx, "test-admin-portal", args)
		if err != nil {
			return err
		}

		// Assert - Validate static assets configuration
		require.NotNil(t, component.StaticAssets)

		component.StaticAssets.ApplyT(func(assets interface{}) interface{} {
			assetsMap := assets.(map[string]interface{})
			assert.NotNil(t, assetsMap["build_command"])
			assert.NotNil(t, assetsMap["dist_folder"])
			assert.NotNil(t, assetsMap["serve_static"])
			assert.NotNil(t, assetsMap["spa_mode"])
			assert.Contains(t, assetsMap["build_command"], "pnpm")
			assert.Equal(t, "dist", assetsMap["dist_folder"])
			return assets
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

// mockResourceMonitor implements the Pulumi resource monitoring interface for testing
type mockResourceMonitor struct{}

func (m *mockResourceMonitor) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return resource.PropertyMap{}, nil
}

func (m *mockResourceMonitor) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	return args.Name + "_id", args.Inputs, nil
}