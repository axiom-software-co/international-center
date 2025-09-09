package refactored

import (
	"testing"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/pkg/components/infrastructure"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/test/shared"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

// TestInfrastructureComponent_RefactoredWithSharedUtilities demonstrates the improved
// maintainability and reduced code duplication using shared testing utilities
func TestInfrastructureComponent_RefactoredWithSharedUtilities(t *testing.T) {
	environments := shared.GetStandardEnvironmentValidations()
	
	for envName, envConfig := range environments {
		t.Run("Environment_"+envName, func(t *testing.T) {
			testCase := shared.ComponentTestCase{
				Name:        "Infrastructure_" + envName,
				Environment: envConfig.Environment,
				Validations: []shared.ValidationFunc{
					func(t *testing.T, component interface{}) {
						infraComponent := component.(*infrastructure.InfrastructureComponent)
						
						// Validate all required endpoints are present
						shared.ValidateStringOutput(t, infraComponent.DatabaseEndpoint, "", "Database endpoint should be defined")
						shared.ValidateStringOutput(t, infraComponent.StorageEndpoint, "", "Storage endpoint should be defined")
						shared.ValidateStringOutput(t, infraComponent.VaultEndpoint, "", "Vault endpoint should be defined")
						shared.ValidateStringOutput(t, infraComponent.MessagingEndpoint, "", "Messaging endpoint should be defined")
						shared.ValidateStringOutput(t, infraComponent.ObservabilityEndpoint, "", "Observability endpoint should be defined")
						
						// Validate environment-specific configurations
						validateEnvironmentSpecificInfrastructure(t, infraComponent, envConfig)
					},
				},
			}
			
			shared.RunPulumiComponentTest(t, testCase, func(ctx *pulumi.Context) (interface{}, error) {
				return infrastructure.NewInfrastructureComponent(ctx, "refactored-test-infrastructure", &infrastructure.InfrastructureArgs{
					Environment: envConfig.Environment,
				})
			})
		})
	}
}

func validateEnvironmentSpecificInfrastructure(t *testing.T, component *infrastructure.InfrastructureComponent, envConfig shared.EnvironmentValidation) {
	switch envConfig.Environment {
	case "development":
		// Development should use local/container-based endpoints
		shared.ValidateStringOutput(t, component.DatabaseEndpoint, "postgres:5432", "Development database should use postgres container")
		shared.ValidateStringOutput(t, component.StorageEndpoint, "azurite:10000", "Development storage should use azurite emulator")
		shared.ValidateStringOutput(t, component.VaultEndpoint, "vault:8200", "Development vault should use local vault container")
		shared.ValidateStringOutput(t, component.MessagingEndpoint, "rabbitmq:5672", "Development messaging should use rabbitmq container")
		shared.ValidateStringOutput(t, component.ObservabilityEndpoint, "grafana:3000", "Development observability should use local grafana")
		
	case "staging":
		// Staging should use Azure-managed services with staging naming
		shared.ValidateStringOutput(t, component.DatabaseEndpoint, "staging", "Staging database should contain staging identifier")
		shared.ValidateStringOutput(t, component.StorageEndpoint, "staging", "Staging storage should contain staging identifier")
		shared.ValidateStringOutput(t, component.VaultEndpoint, "staging", "Staging vault should contain staging identifier")
		shared.ValidateStringOutput(t, component.MessagingEndpoint, "staging", "Staging messaging should contain staging identifier")
		shared.ValidateStringOutput(t, component.ObservabilityEndpoint, "staging", "Staging observability should contain staging identifier")
		
	case "production":
		// Production should use Azure-managed services with production naming
		shared.ValidateStringOutput(t, component.DatabaseEndpoint, "production", "Production database should contain production identifier")
		shared.ValidateStringOutput(t, component.StorageEndpoint, "production", "Production storage should contain production identifier")
		shared.ValidateStringOutput(t, component.VaultEndpoint, "production", "Production vault should contain production identifier")
		shared.ValidateStringOutput(t, component.MessagingEndpoint, "production", "Production messaging should contain production identifier")
		shared.ValidateStringOutput(t, component.ObservabilityEndpoint, "production", "Production observability should contain production identifier")
	}
}

func TestInfrastructureComponent_BooleanOutputs(t *testing.T) {
	testCase := shared.ComponentTestCase{
		Name:        "Infrastructure_Boolean_Outputs_Validation",
		Environment: "production",
		Validations: []shared.ValidationFunc{
			func(t *testing.T, component interface{}) {
				infraComponent := component.(*infrastructure.InfrastructureComponent)
				
				// Validate boolean outputs for production environment
				shared.ValidateBoolOutput(t, infraComponent.HealthCheckEnabled, true, "Health checks should be enabled in production")
				shared.ValidateBoolOutput(t, infraComponent.SecurityPolicies, true, "Security policies should be enabled in production")
				shared.ValidateBoolOutput(t, infraComponent.AuditLogging, true, "Audit logging should be enabled in production")
			},
		},
	}
	
	shared.RunPulumiComponentTest(t, testCase, func(ctx *pulumi.Context) (interface{}, error) {
		return infrastructure.NewInfrastructureComponent(ctx, "config-test-infrastructure", &infrastructure.InfrastructureArgs{
			Environment: "production",
		})
	})
}

func TestInfrastructureComponent_UnsupportedEnvironment_RefactoredWithSharedUtilities(t *testing.T) {
	// Test that unsupported environments are handled gracefully
	// Expect this test to fail during component creation, not during validation
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		_, err := infrastructure.NewInfrastructureComponent(ctx, "unsupported-test-infrastructure", &infrastructure.InfrastructureArgs{
			Environment: "unsupported",
		})
		return err
	}, pulumi.WithMocks("project", "stack", &shared.SharedMockResourceMonitor{}))
	
	// Should get an error for unsupported environment
	assert.Error(t, err, "Should receive error for unsupported environment")
	assert.Contains(t, err.Error(), "unsupported environment", "Error message should indicate unsupported environment")
}