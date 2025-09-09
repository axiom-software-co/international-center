package builders_test

import (
	"testing"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/builders"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewComponentBuilder(t *testing.T) {
	// Arrange
	testCases := []struct {
		name        string
		environment string
		expectNil   bool
	}{
		{
			name:        "Valid development environment",
			environment: "development",
			expectNil:   false,
		},
		{
			name:        "Valid staging environment",
			environment: "staging",
			expectNil:   false,
		},
		{
			name:        "Valid production environment",
			environment: "production",
			expectNil:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := createMockPulumiContext(t)

			// Act
			builder := builders.NewComponentBuilder(ctx, tc.environment)

			// Assert
			if tc.expectNil {
				assert.Nil(t, builder)
			} else {
				assert.NotNil(t, builder)
			}
		})
	}
}

func TestComponentBuilder_ValidateEnvironment(t *testing.T) {
	// Arrange
	testCases := []struct {
		name          string
		environment   string
		expectError   bool
		expectedError string
	}{
		{
			name:        "Development environment is valid",
			environment: "development",
			expectError: false,
		},
		{
			name:        "Staging environment is valid", 
			environment: "staging",
			expectError: false,
		},
		{
			name:        "Production environment is valid",
			environment: "production",
			expectError: false,
		},
		{
			name:          "Invalid environment",
			environment:   "invalid",
			expectError:   true,
			expectedError: "invalid environment: invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := createMockPulumiContext(t)
			builder := builders.NewComponentBuilder(ctx, tc.environment)
			require.NotNil(t, builder)

			// Act
			err := builder.ValidateEnvironment()

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != "" {
					assert.Contains(t, err.Error(), tc.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestComponentBuilder_BuildInfrastructure(t *testing.T) {
	// Arrange
	testCases := []struct {
		name        string
		environment string
	}{
		{
			name:        "Build infrastructure for development",
			environment: "development",
		},
		{
			name:        "Build infrastructure for staging",
			environment: "staging",
		},
		{
			name:        "Build infrastructure for production",
			environment: "production",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := createMockPulumiContext(t)
			builder := builders.NewComponentBuilder(ctx, tc.environment)
			require.NotNil(t, builder)

			// Act
			component, err := builder.BuildInfrastructure()

			// Assert
			// Component should be successfully created with mock context
			assert.NoError(t, err)
			assert.NotNil(t, component)
			// Verify component has expected structure
			assert.NotNil(t, component.DatabaseEndpoint)
			assert.NotNil(t, component.StorageEndpoint)
			assert.NotNil(t, component.VaultEndpoint)
			assert.NotNil(t, component.MessagingEndpoint)
		})
	}
}

func TestComponentBuilder_BuildPlatform(t *testing.T) {
	// Arrange
	testCases := []struct {
		name        string
		environment string
	}{
		{
			name:        "Build platform for development",
			environment: "development",
		},
		{
			name:        "Build platform for staging",
			environment: "staging",
		},
		{
			name:        "Build platform for production",
			environment: "production",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := createMockPulumiContext(t)
			builder := builders.NewComponentBuilder(ctx, tc.environment)
			require.NotNil(t, builder)

			// Mock infrastructure outputs
			infraOutputs := createMockOutputs("infrastructure")

			// Act
			component, err := builder.BuildPlatform(infraOutputs)

			// Assert
			// Component should be successfully created with mock context
			assert.NoError(t, err)
			assert.NotNil(t, component)
			// Verify component has expected structure
			assert.NotNil(t, component.DaprEndpoint)
			assert.NotNil(t, component.OrchestrationEndpoint)
			assert.NotNil(t, component.ServiceMeshEnabled)
		})
	}
}

func TestComponentBuilder_BuildServices(t *testing.T) {
	// Arrange
	testCases := []struct {
		name        string
		environment string
	}{
		{
			name:        "Build services for development",
			environment: "development",
		},
		{
			name:        "Build services for staging",
			environment: "staging",
		},
		{
			name:        "Build services for production",
			environment: "production",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := createMockPulumiContext(t)
			builder := builders.NewComponentBuilder(ctx, tc.environment)
			require.NotNil(t, builder)

			// Mock outputs
			infraOutputs := createMockOutputs("infrastructure")
			platformOutputs := createMockOutputs("platform")

			// Act
			component, err := builder.BuildServices(infraOutputs, platformOutputs)

			// Assert
			// Component should be successfully created with mock context
			assert.NoError(t, err)
			assert.NotNil(t, component)
			// Verify component has expected structure
			assert.NotNil(t, component.PublicGatewayURL)
			assert.NotNil(t, component.AdminGatewayURL)
			assert.NotNil(t, component.DeploymentType)
		})
	}
}

func TestComponentBuilder_BuildWebsite(t *testing.T) {
	// Arrange
	testCases := []struct {
		name        string
		environment string
	}{
		{
			name:        "Build website for development",
			environment: "development",
		},
		{
			name:        "Build website for staging",
			environment: "staging",
		},
		{
			name:        "Build website for production",
			environment: "production",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := createMockPulumiContext(t)
			builder := builders.NewComponentBuilder(ctx, tc.environment)
			require.NotNil(t, builder)

			// Mock outputs
			infraOutputs := createMockOutputs("infrastructure")
			platformOutputs := createMockOutputs("platform")
			servicesOutputs := createMockOutputs("services")

			// Act
			component, err := builder.BuildWebsite(infraOutputs, platformOutputs, servicesOutputs)

			// Assert
			// Component should be successfully created with mock context
			assert.NoError(t, err)
			assert.NotNil(t, component)
			// Verify component has expected structure
			assert.NotNil(t, component.WebsiteURL)
			assert.NotNil(t, component.DeploymentType)
			assert.NotNil(t, component.HealthCheckEnabled)
		})
	}
}

func TestComponentBuilder_EnvironmentValidation_Properties(t *testing.T) {
	// Property: Valid environments should never return validation errors
	validEnvironments := []string{"development", "staging", "production"}

	for _, env := range validEnvironments {
		t.Run("Property_ValidEnvironment_"+env, func(t *testing.T) {
			// Arrange
			ctx := createMockPulumiContext(t)
			builder := builders.NewComponentBuilder(ctx, env)
			require.NotNil(t, builder)

			// Act
			err := builder.ValidateEnvironment()

			// Assert - Property: valid environments never error
			assert.NoError(t, err, "Valid environment %s should not return validation error", env)
		})
	}
}

func TestComponentBuilder_InvalidEnvironment_Properties(t *testing.T) {
	// Property: Invalid environments should always return validation errors
	invalidEnvironments := []string{"invalid", "test", "", "dev", "prod"}

	for _, env := range invalidEnvironments {
		t.Run("Property_InvalidEnvironment_"+env, func(t *testing.T) {
			// Arrange
			ctx := createMockPulumiContext(t)
			builder := builders.NewComponentBuilder(ctx, env)
			require.NotNil(t, builder)

			// Act
			err := builder.ValidateEnvironment()

			// Assert - Property: invalid environments always error
			assert.Error(t, err, "Invalid environment %s should return validation error", env)
		})
	}
}

// Helper functions for testing
func createMockPulumiContext(t *testing.T) *pulumi.Context {
	// For unit tests, we create a minimal context that allows basic operations
	// This enables testing component structure without full Pulumi runtime
	return &pulumi.Context{}
}

func createMockOutputs(componentType string) pulumi.Map {
	// Create mock outputs using proper Pulumi types for testing
	return pulumi.Map{
		componentType + "_output": pulumi.String("mock_value"),
		"connection_string":       pulumi.String("mock_connection"),
		"endpoint":               pulumi.String("mock_endpoint"),
	}
}

// Benchmark tests
func BenchmarkComponentBuilder_Creation(b *testing.B) {
	// Arrange
	ctx := createMockPulumiContext(nil)

	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := builders.NewComponentBuilder(ctx, "development")
		_ = builder // Prevent compiler optimization
	}
}

func BenchmarkComponentBuilder_EnvironmentValidation(b *testing.B) {
	// Arrange
	ctx := createMockPulumiContext(nil)
	builder := builders.NewComponentBuilder(ctx, "development")

	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := builder.ValidateEnvironment()
		_ = err // Prevent compiler optimization
	}
}

// Table-driven test for all valid environments
func TestComponentBuilder_AllValidEnvironments(t *testing.T) {
	// Arrange
	environments := []struct {
		name string
		env  string
	}{
		{"Development", "development"},
		{"Staging", "staging"},
		{"Production", "production"},
	}

	ctx := createMockPulumiContext(t)

	for _, environment := range environments {
		t.Run(environment.name, func(t *testing.T) {
			// Act
			builder := builders.NewComponentBuilder(ctx, environment.env)

			// Assert
			assert.NotNil(t, builder)

			// Validate environment
			err := builder.ValidateEnvironment()
			assert.NoError(t, err)
		})
	}
}

// Error handling tests
func TestComponentBuilder_ErrorHandling(t *testing.T) {
	// Arrange
	ctx := createMockPulumiContext(t)
	builder := builders.NewComponentBuilder(ctx, "invalid_environment")
	require.NotNil(t, builder)

	// Act
	err := builder.ValidateEnvironment()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid environment")
	assert.Contains(t, err.Error(), "Valid environments:")
}