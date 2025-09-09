package website

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebsiteComponent_Development(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &WebsiteArgs{
			Environment:           "development",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs:       pulumi.Map{},
			ServicesOutputs:       pulumi.Map{},
		}

		// Act
		component, err := NewWebsiteComponent(ctx, "test-website", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate outputs are defined
		assert.NotNil(t, component.WebsiteURL)
		assert.NotNil(t, component.DeploymentType)
		assert.NotNil(t, component.CDNEnabled)
		assert.NotNil(t, component.SSLEnabled)
		assert.NotNil(t, component.HealthCheckEnabled)

		// Validate development-specific configurations
		component.WebsiteURL.ApplyT(func(url string) string {
			assert.Contains(t, url, "localhost")
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

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestWebsiteComponent_Staging(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &WebsiteArgs{
			Environment:           "staging",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs:       pulumi.Map{},
			ServicesOutputs:       pulumi.Map{},
		}

		// Act
		component, err := NewWebsiteComponent(ctx, "test-website", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate outputs are defined
		assert.NotNil(t, component.WebsiteURL)
		assert.NotNil(t, component.DeploymentType)
		assert.NotNil(t, component.CDNEnabled)
		assert.NotNil(t, component.SSLEnabled)

		// Validate staging-specific configurations
		component.WebsiteURL.ApplyT(func(url string) string {
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

func TestWebsiteComponent_Production(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &WebsiteArgs{
			Environment:           "production",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs:       pulumi.Map{},
			ServicesOutputs:       pulumi.Map{},
		}

		// Act
		component, err := NewWebsiteComponent(ctx, "test-website", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate outputs are defined
		assert.NotNil(t, component.WebsiteURL)
		assert.NotNil(t, component.DeploymentType)
		assert.NotNil(t, component.CDNEnabled)
		assert.NotNil(t, component.SSLEnabled)

		// Validate production-specific configurations
		component.WebsiteURL.ApplyT(func(url string) string {
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

func TestWebsiteComponent_ContainerConfiguration(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &WebsiteArgs{
			Environment:           "production",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs:       pulumi.Map{},
			ServicesOutputs:       pulumi.Map{},
		}

		// Act
		component, err := NewWebsiteComponent(ctx, "test-website", args)
		if err != nil {
			return err
		}

		// Assert - Validate container configurations
		require.NotNil(t, component.ContainerConfig)

		component.ContainerConfig.ApplyT(func(config interface{}) interface{} {
			configMap := config.(map[string]interface{})
			assert.NotNil(t, configMap["image"])
			assert.NotNil(t, configMap["replicas"])
			assert.NotNil(t, configMap["health_check"])
			assert.NotNil(t, configMap["resource_limits"])
			return config
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestWebsiteComponent_CacheConfiguration(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &WebsiteArgs{
			Environment:           "production",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs:       pulumi.Map{},
			ServicesOutputs:       pulumi.Map{},
		}

		// Act
		component, err := NewWebsiteComponent(ctx, "test-website", args)
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
			return config
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestWebsiteComponent_StaticAssets(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &WebsiteArgs{
			Environment:           "production",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs:       pulumi.Map{},
			ServicesOutputs:       pulumi.Map{},
		}

		// Act
		component, err := NewWebsiteComponent(ctx, "test-website", args)
		if err != nil {
			return err
		}

		// Assert - Validate static asset configurations
		require.NotNil(t, component.StaticAssets)

		component.StaticAssets.ApplyT(func(config interface{}) interface{} {
			configMap := config.(map[string]interface{})
			assert.NotNil(t, configMap["build_command"])
			assert.NotNil(t, configMap["dist_folder"])
			assert.NotNil(t, configMap["serve_static"])
			assert.NotNil(t, configMap["spa_mode"])
			return config
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