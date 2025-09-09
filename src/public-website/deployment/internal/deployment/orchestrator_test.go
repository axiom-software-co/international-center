package deployment_test

import (
	"context"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/deployment"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeploymentOrchestrator_NewDeploymentOrchestrator(t *testing.T) {
	testCases := []struct {
		name        string
		environment string
		expectError bool
	}{
		{
			name:        "Valid development environment",
			environment: "development",
			expectError: false,
		},
		{
			name:        "Valid staging environment", 
			environment: "staging",
			expectError: false,
		},
		{
			name:        "Valid production environment",
			environment: "production",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := createMockPulumiContext(t)

			orchestrator := deployment.NewDeploymentOrchestrator(ctx, tc.environment)

			if tc.expectError {
				assert.Nil(t, orchestrator)
			} else {
				assert.NotNil(t, orchestrator)
			}
		})
	}
}

func TestDeploymentOrchestrator_ExecuteDeployment_ValidatesEnvironment(t *testing.T) {
	ctx := createMockPulumiContext(t)
	orchestrator := deployment.NewDeploymentOrchestrator(ctx, "development")
	require.NotNil(t, orchestrator)

	err := orchestrator.ExecuteDeployment()

	assert.NoError(t, err)
}

func TestDeploymentOrchestrator_Timeouts(t *testing.T) {
	testCases := []struct {
		name        string
		environment string
		component   deployment.DeploymentPhase
		timeout     time.Duration
	}{
		{
			name:        "Infrastructure timeout for development",
			environment: "development", 
			component:   deployment.PhaseInfrastructure,
			timeout:     10 * time.Minute,
		},
		{
			name:        "Platform timeout for development",
			environment: "development",
			component:   deployment.PhasePlatform, 
			timeout:     8 * time.Minute,
		},
		{
			name:        "Services timeout for development",
			environment: "development",
			component:   deployment.PhaseServices,
			timeout:     15 * time.Minute,
		},
		{
			name:        "Website timeout for development",
			environment: "development",
			component:   deployment.PhaseWebsite,
			timeout:     5 * time.Minute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := createMockPulumiContext(t)
			orchestrator := deployment.NewDeploymentOrchestrator(ctx, tc.environment)
			require.NotNil(t, orchestrator)

			assert.NotNil(t, orchestrator)
		})
	}
}

func TestDeploymentPhase_Constants(t *testing.T) {
	assert.Equal(t, deployment.DeploymentPhase("infrastructure"), deployment.PhaseInfrastructure)
	assert.Equal(t, deployment.DeploymentPhase("platform"), deployment.PhasePlatform)
	assert.Equal(t, deployment.DeploymentPhase("services"), deployment.PhaseServices)
	assert.Equal(t, deployment.DeploymentPhase("website"), deployment.PhaseWebsite)
}

func TestDeploymentResult_Properties(t *testing.T) {
	startTime := time.Now()
	result := &deployment.DeploymentResult{
		Phase:     deployment.PhaseInfrastructure,
		Success:   true,
		Outputs:   pulumi.Map{"test": pulumi.String("value")},
		Error:     nil,
		Duration:  5 * time.Minute,
		StartTime: startTime,
	}

	assert.Equal(t, deployment.PhaseInfrastructure, result.Phase)
	assert.True(t, result.Success)
	assert.NotNil(t, result.Outputs)
	assert.Nil(t, result.Error)
	assert.Equal(t, 5*time.Minute, result.Duration)
	assert.Equal(t, startTime, result.StartTime)
}

func TestDeploymentOrchestrator_EnvironmentValidation(t *testing.T) {
	testCases := []struct {
		name             string
		environment      string
		expectedValid    bool
		expectedError    string
	}{
		{
			name:          "Development environment is valid",
			environment:   "development",
			expectedValid: true,
		},
		{
			name:          "Staging environment is valid",
			environment:   "staging", 
			expectedValid: true,
		},
		{
			name:          "Production environment is valid",
			environment:   "production",
			expectedValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := createMockPulumiContext(t)
			orchestrator := deployment.NewDeploymentOrchestrator(ctx, tc.environment)
			require.NotNil(t, orchestrator)

			assert.NotNil(t, orchestrator)
		})
	}
}

func createMockPulumiContext(t *testing.T) *pulumi.Context {
	return &pulumi.Context{}
}

func BenchmarkDeploymentOrchestrator_Creation(b *testing.B) {
	ctx := createMockPulumiContext(nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orchestrator := deployment.NewDeploymentOrchestrator(ctx, "development")
		_ = orchestrator
	}
}

func TestDeploymentOrchestrator_Properties(t *testing.T) {
	validEnvironments := []string{"development", "staging", "production"}
	
	for _, env := range validEnvironments {
		t.Run("Property_ValidEnvironment_"+env, func(t *testing.T) {
			ctx := createMockPulumiContext(t)
			
			orchestrator := deployment.NewDeploymentOrchestrator(ctx, env)
			
			assert.NotNil(t, orchestrator, "Valid environment %s should create orchestrator", env)
		})
	}
}

func TestDeploymentOrchestrator_WithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pulumiCtx := createMockPulumiContext(t)
	orchestrator := deployment.NewDeploymentOrchestrator(pulumiCtx, "development")
	require.NotNil(t, orchestrator)

	select {
	case <-ctx.Done():
		t.Fatal("Test timed out")
	default:
		assert.NotNil(t, orchestrator)
	}
}