package platform

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlatformComponent_Development(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &PlatformArgs{
			Environment:           "development",
			InfrastructureOutputs: pulumi.Map{},
		}

		// Act
		component, err := NewPlatformComponent(ctx, "test-platform", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate outputs are defined
		assert.NotNil(t, component.DaprEndpoint)
		assert.NotNil(t, component.OrchestrationEndpoint)
		assert.NotNil(t, component.NetworkingConfig)
		assert.NotNil(t, component.SecurityConfig)

		// Validate development-specific configurations
		component.DaprEndpoint.ApplyT(func(endpoint string) string {
			assert.Contains(t, endpoint, "localhost")
			return endpoint
		})

		component.OrchestrationEndpoint.ApplyT(func(endpoint string) string {
			assert.Contains(t, endpoint, "podman")
			return endpoint
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestPlatformComponent_Staging(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &PlatformArgs{
			Environment:           "staging",
			InfrastructureOutputs: pulumi.Map{},
		}

		// Act
		component, err := NewPlatformComponent(ctx, "test-platform", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate outputs are defined
		assert.NotNil(t, component.DaprEndpoint)
		assert.NotNil(t, component.OrchestrationEndpoint)
		assert.NotNil(t, component.NetworkingConfig)
		assert.NotNil(t, component.SecurityConfig)

		// Validate staging-specific configurations
		component.DaprEndpoint.ApplyT(func(endpoint string) string {
			assert.Contains(t, endpoint, "staging")
			return endpoint
		})

		component.OrchestrationEndpoint.ApplyT(func(endpoint string) string {
			assert.Contains(t, endpoint, "container")
			return endpoint
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestPlatformComponent_Production(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &PlatformArgs{
			Environment:           "production",
			InfrastructureOutputs: pulumi.Map{},
		}

		// Act
		component, err := NewPlatformComponent(ctx, "test-platform", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate outputs are defined
		assert.NotNil(t, component.DaprEndpoint)
		assert.NotNil(t, component.OrchestrationEndpoint)
		assert.NotNil(t, component.NetworkingConfig)
		assert.NotNil(t, component.SecurityConfig)

		// Validate production-specific configurations
		component.DaprEndpoint.ApplyT(func(endpoint string) string {
			assert.Contains(t, endpoint, "production")
			return endpoint
		})

		component.OrchestrationEndpoint.ApplyT(func(endpoint string) string {
			assert.Contains(t, endpoint, "container")
			return endpoint
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestPlatformComponent_SecurityConfiguration(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &PlatformArgs{
			Environment:           "production",
			InfrastructureOutputs: pulumi.Map{},
		}

		// Act
		component, err := NewPlatformComponent(ctx, "test-platform", args)
		if err != nil {
			return err
		}

		// Assert - Validate security configurations
		require.NotNil(t, component.SecurityConfig)

		component.SecurityConfig.ApplyT(func(config interface{}) interface{} {
			configMap := config.(map[string]interface{})
			assert.NotNil(t, configMap["authentication_enabled"])
			assert.NotNil(t, configMap["authorization_enabled"])
			assert.NotNil(t, configMap["audit_logging_enabled"])
			return config
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestPlatformComponent_NetworkingConfiguration(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &PlatformArgs{
			Environment:           "development",
			InfrastructureOutputs: pulumi.Map{},
		}

		// Act
		component, err := NewPlatformComponent(ctx, "test-platform", args)
		if err != nil {
			return err
		}

		// Assert - Validate networking configurations
		require.NotNil(t, component.NetworkingConfig)

		component.NetworkingConfig.ApplyT(func(config interface{}) interface{} {
			configMap := config.(map[string]interface{})
			assert.NotNil(t, configMap["service_discovery_enabled"])
			assert.NotNil(t, configMap["load_balancing_enabled"])
			assert.NotNil(t, configMap["circuit_breaker_enabled"])
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