package infrastructure

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfrastructureComponent_Development(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &InfrastructureArgs{
			Environment: "development",
		}

		// Act
		component, err := NewInfrastructureComponent(ctx, "test-infrastructure", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate outputs are defined
		assert.NotNil(t, component.DatabaseEndpoint)
		assert.NotNil(t, component.StorageEndpoint)
		assert.NotNil(t, component.VaultEndpoint)
		assert.NotNil(t, component.MessagingEndpoint)
		assert.NotNil(t, component.ObservabilityEndpoint)

		// Validate environment-specific configurations
		component.DatabaseEndpoint.ApplyT(func(endpoint string) string {
			assert.Contains(t, endpoint, "localhost:5432")
			assert.Contains(t, endpoint, "international_center_dev")
			return endpoint
		})

		component.StorageEndpoint.ApplyT(func(endpoint string) string {
			assert.Contains(t, endpoint, "127.0.0.1:10000")
			assert.Contains(t, endpoint, "devstoreaccount1")
			return endpoint
		})

		component.VaultEndpoint.ApplyT(func(endpoint string) string {
			assert.Contains(t, endpoint, "127.0.0.1:8200")
			return endpoint
		})

		component.MessagingEndpoint.ApplyT(func(endpoint string) string {
			assert.Contains(t, endpoint, "localhost:5672")
			return endpoint
		})

		component.ObservabilityEndpoint.ApplyT(func(endpoint string) string {
			assert.Contains(t, endpoint, "localhost:3000")
			return endpoint
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestInfrastructureComponent_Staging(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &InfrastructureArgs{
			Environment: "staging",
		}

		// Act
		component, err := NewInfrastructureComponent(ctx, "test-infrastructure", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate outputs are defined
		assert.NotNil(t, component.DatabaseEndpoint)
		assert.NotNil(t, component.StorageEndpoint)
		assert.NotNil(t, component.VaultEndpoint)
		assert.NotNil(t, component.MessagingEndpoint)
		assert.NotNil(t, component.ObservabilityEndpoint)

		// Validate staging-specific configurations
		component.DatabaseEndpoint.ApplyT(func(endpoint string) string {
			assert.Contains(t, endpoint, "staging")
			return endpoint
		})

		component.StorageEndpoint.ApplyT(func(endpoint string) string {
			assert.Contains(t, endpoint, "staging")
			return endpoint
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestInfrastructureComponent_Production(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &InfrastructureArgs{
			Environment: "production",
		}

		// Act
		component, err := NewInfrastructureComponent(ctx, "test-infrastructure", args)
		if err != nil {
			return err
		}

		// Assert - Validate component was created
		require.NotNil(t, component)

		// Validate outputs are defined
		assert.NotNil(t, component.DatabaseEndpoint)
		assert.NotNil(t, component.StorageEndpoint)
		assert.NotNil(t, component.VaultEndpoint)
		assert.NotNil(t, component.MessagingEndpoint)
		assert.NotNil(t, component.ObservabilityEndpoint)

		// Validate production-specific configurations
		component.DatabaseEndpoint.ApplyT(func(endpoint string) string {
			assert.Contains(t, endpoint, "production")
			return endpoint
		})

		component.StorageEndpoint.ApplyT(func(endpoint string) string {
			assert.Contains(t, endpoint, "production")
			return endpoint
		})

		return nil
	}, pulumi.WithMocks("project", "stack", &mockResourceMonitor{}))

	require.NoError(t, err)
}

func TestInfrastructureComponent_UnsupportedEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// Arrange
		args := &InfrastructureArgs{
			Environment: "unsupported",
		}

		// Act
		_, err := NewInfrastructureComponent(ctx, "test-infrastructure", args)

		// Assert - Should return error for unsupported environment
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported environment")

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