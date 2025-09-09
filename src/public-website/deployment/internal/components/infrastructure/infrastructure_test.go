package infrastructure

import (
	"testing"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/validation"
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

// Refactored tests using shared validation utilities

func TestInfrastructureComponent_RefactoredWithSharedUtilities(t *testing.T) {
	environments := validation.GetStandardEnvironmentValidations()
	
	for envName, envConfig := range environments {
		t.Run("Environment_"+envName, func(t *testing.T) {
			testCase := validation.ComponentTestCase{
				Name:        "Infrastructure_" + envName,
				Environment: envConfig.Environment,
				Validations: []validation.ValidationFunc{
					func(t *testing.T, component interface{}) {
						infraComponent := component.(*InfrastructureComponent)
						
						// Validate all required endpoints are present
						validation.ValidateStringOutput(t, infraComponent.DatabaseEndpoint, "", "Database endpoint should be defined")
						validation.ValidateStringOutput(t, infraComponent.StorageEndpoint, "", "Storage endpoint should be defined")
						validation.ValidateStringOutput(t, infraComponent.VaultEndpoint, "", "Vault endpoint should be defined")
						validation.ValidateStringOutput(t, infraComponent.MessagingEndpoint, "", "Messaging endpoint should be defined")
						validation.ValidateStringOutput(t, infraComponent.ObservabilityEndpoint, "", "Observability endpoint should be defined")
						
						// Validate environment-specific configurations
						validateEnvironmentSpecificInfrastructure(t, infraComponent, envConfig)
					},
				},
			}
			
			validation.RunPulumiComponentTest(t, testCase, func(ctx *pulumi.Context) (interface{}, error) {
				return NewInfrastructureComponent(ctx, "refactored-test-infrastructure", &InfrastructureArgs{
					Environment: envConfig.Environment,
				})
			})
		})
	}
}

func validateEnvironmentSpecificInfrastructure(t *testing.T, component *InfrastructureComponent, envConfig validation.EnvironmentValidation) {
	switch envConfig.Environment {
	case "development":
		// Development should use local/container-based endpoints
		validation.ValidateStringOutput(t, component.DatabaseEndpoint, "postgres:5432", "Development database should use postgres container")
		validation.ValidateStringOutput(t, component.StorageEndpoint, "azurite:10000", "Development storage should use azurite emulator")
		validation.ValidateStringOutput(t, component.VaultEndpoint, "vault:8200", "Development vault should use local vault container")
		validation.ValidateStringOutput(t, component.MessagingEndpoint, "rabbitmq:5672", "Development messaging should use rabbitmq container")
		validation.ValidateStringOutput(t, component.ObservabilityEndpoint, "grafana:3000", "Development observability should use local grafana")
		
	case "staging":
		// Staging should use Azure-managed services with staging naming
		validation.ValidateStringOutput(t, component.DatabaseEndpoint, "staging", "Staging database should contain staging identifier")
		validation.ValidateStringOutput(t, component.StorageEndpoint, "staging", "Staging storage should contain staging identifier")
		validation.ValidateStringOutput(t, component.VaultEndpoint, "staging", "Staging vault should contain staging identifier")
		validation.ValidateStringOutput(t, component.MessagingEndpoint, "staging", "Staging messaging should contain staging identifier")
		validation.ValidateStringOutput(t, component.ObservabilityEndpoint, "staging", "Staging observability should contain staging identifier")
		
	case "production":
		// Production should use Azure-managed services with production naming
		validation.ValidateStringOutput(t, component.DatabaseEndpoint, "production", "Production database should contain production identifier")
		validation.ValidateStringOutput(t, component.StorageEndpoint, "production", "Production storage should contain production identifier")
		validation.ValidateStringOutput(t, component.VaultEndpoint, "production", "Production vault should contain production identifier")
		validation.ValidateStringOutput(t, component.MessagingEndpoint, "production", "Production messaging should contain production identifier")
		validation.ValidateStringOutput(t, component.ObservabilityEndpoint, "production", "Production observability should contain production identifier")
	}
}

func TestInfrastructureComponent_BooleanOutputs(t *testing.T) {
	testCase := validation.ComponentTestCase{
		Name:        "Infrastructure_Boolean_Outputs_Validation",
		Environment: "production",
		Validations: []validation.ValidationFunc{
			func(t *testing.T, component interface{}) {
				infraComponent := component.(*InfrastructureComponent)
				
				// Validate boolean outputs for production environment
				validation.ValidateBoolOutput(t, infraComponent.HealthCheckEnabled, true, "Health checks should be enabled in production")
				validation.ValidateBoolOutput(t, infraComponent.SecurityPolicies, true, "Security policies should be enabled in production")
				validation.ValidateBoolOutput(t, infraComponent.AuditLogging, true, "Audit logging should be enabled in production")
			},
		},
	}
	
	validation.RunPulumiComponentTest(t, testCase, func(ctx *pulumi.Context) (interface{}, error) {
		return NewInfrastructureComponent(ctx, "config-test-infrastructure", &InfrastructureArgs{
			Environment: "production",
		})
	})
}