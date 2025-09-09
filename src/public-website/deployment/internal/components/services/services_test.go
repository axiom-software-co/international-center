package services

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServicesComponent_Development(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &ServicesArgs{
			Environment:           "development",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs:       pulumi.Map{},
		}

		// Act
		component, err := NewServicesComponent(ctx, "test-services", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate outputs are defined
		assert.NotNil(t, component.PublicGatewayURL)
		assert.NotNil(t, component.AdminGatewayURL)
		assert.NotNil(t, component.ContentServiceURL)
		assert.NotNil(t, component.InquiriesServiceURL)
		assert.NotNil(t, component.NotificationsServiceURL)

		// Validate development-specific configurations
		component.PublicGatewayURL.ApplyT(func(url string) string {
			assert.Contains(t, url, "127.0.0.1")
			return url
		})

		component.AdminGatewayURL.ApplyT(func(url string) string {
			assert.Contains(t, url, "127.0.0.1")
			return url
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestServicesComponent_Staging(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &ServicesArgs{
			Environment:           "staging",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs:       pulumi.Map{},
		}

		// Act
		component, err := NewServicesComponent(ctx, "test-services", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate outputs are defined
		assert.NotNil(t, component.PublicGatewayURL)
		assert.NotNil(t, component.AdminGatewayURL)
		assert.NotNil(t, component.ContentServiceURL)
		assert.NotNil(t, component.InquiriesServiceURL)
		assert.NotNil(t, component.NotificationsServiceURL)

		// Validate staging-specific configurations
		component.PublicGatewayURL.ApplyT(func(url string) string {
			assert.Contains(t, url, "staging")
			assert.Contains(t, url, "azurecontainerapp.io")
			return url
		})

		component.AdminGatewayURL.ApplyT(func(url string) string {
			assert.Contains(t, url, "staging")
			assert.Contains(t, url, "azurecontainerapp.io")
			return url
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestServicesComponent_Production(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &ServicesArgs{
			Environment:           "production",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs:       pulumi.Map{},
		}

		// Act
		component, err := NewServicesComponent(ctx, "test-services", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate outputs are defined
		assert.NotNil(t, component.PublicGatewayURL)
		assert.NotNil(t, component.AdminGatewayURL)
		assert.NotNil(t, component.ContentServiceURL)
		assert.NotNil(t, component.InquiriesServiceURL)
		assert.NotNil(t, component.NotificationsServiceURL)

		// Validate production-specific configurations
		component.PublicGatewayURL.ApplyT(func(url string) string {
			assert.Contains(t, url, "production")
			assert.Contains(t, url, "azurecontainerapp.io")
			return url
		})

		component.AdminGatewayURL.ApplyT(func(url string) string {
			assert.Contains(t, url, "production")
			assert.Contains(t, url, "azurecontainerapp.io")
			return url
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestServicesComponent_GatewayConfiguration(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &ServicesArgs{
			Environment:           "production",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs:       pulumi.Map{},
		}

		// Act
		component, err := NewServicesComponent(ctx, "test-services", args)
		if err != nil {
			return err
		}

		// Assert - Validate gateway configurations
		require.NotNil(t, component.GatewayConfiguration)

		component.GatewayConfiguration.ApplyT(func(config interface{}) interface{} {
			configMap := config.(map[string]interface{})
			assert.NotNil(t, configMap["rate_limiting_enabled"])
			assert.NotNil(t, configMap["cors_enabled"])
			assert.NotNil(t, configMap["security_headers_enabled"])
			assert.NotNil(t, configMap["audit_logging_enabled"])
			return config
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestServicesComponent_ServiceConfiguration(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &ServicesArgs{
			Environment:           "development",
			InfrastructureOutputs: pulumi.Map{},
			PlatformOutputs:       pulumi.Map{},
		}

		// Act
		component, err := NewServicesComponent(ctx, "test-services", args)
		if err != nil {
			return err
		}

		// Assert - Validate service configurations
		require.NotNil(t, component.ServiceConfiguration)

		component.ServiceConfiguration.ApplyT(func(config interface{}) interface{} {
			configMap := config.(map[string]interface{})
			assert.NotNil(t, configMap["health_checks_enabled"])
			assert.NotNil(t, configMap["metrics_enabled"])
			assert.NotNil(t, configMap["distributed_tracing_enabled"])
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