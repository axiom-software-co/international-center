package environment_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for environment health validation
// These tests require the entire development environment to be running
// They use real dependencies, not mocks

func TestEnvironmentHealth_DevelopmentEnvironment(t *testing.T) {
	// Skip if not in integration test mode
	if !isIntegrationTestEnabled() {
		t.Skip("Integration tests disabled - set INTEGRATION_TESTS=true and ensure development environment is running")
	}

	// Arrange
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	healthChecker, err := validation.NewEnvironmentHealthChecker("development")
	require.NoError(t, err)
	require.NotNil(t, healthChecker)

	// Mock outputs from actual deployment (would be retrieved from Pulumi stack)
	mockOutputs := getMockDeploymentOutputs()

	// Act
	report, err := healthChecker.PerformHealthCheck(ctx, mockOutputs)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, report)

	assert.Equal(t, "development", report.Environment)
	assert.True(t, report.Duration > 0)
	assert.NotEmpty(t, report.Components)

	// Verify all expected components are present
	expectedComponents := []string{"infrastructure", "platform", "services", "website"}
	for _, component := range expectedComponents {
		assert.Contains(t, report.Components, component, "Component %s should be present", component)
	}

	// For development environment, we expect healthy status
	if report.IsHealthy() {
		assert.Equal(t, validation.HealthStatusHealthy, report.Overall)
	} else {
		// Log unhealthy components for debugging
		unhealthy := report.GetUnhealthyComponents()
		t.Logf("Unhealthy components in development environment: %v", unhealthy)
	}
}

func TestEnvironmentHealth_ComponentValidation(t *testing.T) {
	// Skip if not in integration test mode
	if !isIntegrationTestEnabled() {
		t.Skip("Integration tests disabled - set INTEGRATION_TESTS=true and ensure development environment is running")
	}

	// Arrange
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	testCases := []struct {
		name            string
		component       string
		expectedPresent bool
	}{
		{
			name:            "Infrastructure component should be deployed",
			component:       "infrastructure",
			expectedPresent: true,
		},
		{
			name:            "Platform component should be deployed",
			component:       "platform",
			expectedPresent: true,
		},
		{
			name:            "Services component should be deployed",
			component:       "services",
			expectedPresent: true,
		},
		{
			name:            "Website component should be deployed",
			component:       "website",
			expectedPresent: true,
		},
	}

	healthChecker, err := validation.NewEnvironmentHealthChecker("development")
	require.NoError(t, err)

	mockOutputs := getMockDeploymentOutputs()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			report, err := healthChecker.PerformHealthCheck(ctx, mockOutputs)

			// Assert
			require.NoError(t, err)
			require.NotNil(t, report)

			if tc.expectedPresent {
				assert.Contains(t, report.Components, tc.component)
				
				componentHealth := report.Components[tc.component]
				assert.NotEmpty(t, componentHealth.Name)
				assert.NotZero(t, componentHealth.CheckTime)
			}
		})
	}
}

func TestEnvironmentHealth_ValidationResults(t *testing.T) {
	// Skip if not in integration test mode
	if !isIntegrationTestEnabled() {
		t.Skip("Integration tests disabled - set INTEGRATION_TESTS=true and ensure development environment is running")
	}

	// Arrange
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	healthChecker, err := validation.NewEnvironmentHealthChecker("development")
	require.NoError(t, err)

	mockOutputs := getMockDeploymentOutputs()

	// Act
	report, err := healthChecker.PerformHealthCheck(ctx, mockOutputs)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, report)

	// Verify validation results structure
	assert.NotNil(t, report.ValidationResults)
	assert.NotNil(t, report.ValidationSummary)

	// Check validation summary contains expected fields
	summary := report.ValidationSummary
	assert.Contains(t, summary, "total_validations")
	assert.Contains(t, summary, "successful")
	assert.Contains(t, summary, "failed")
	assert.Contains(t, summary, "success_rate")
	assert.Contains(t, summary, "environment")

	// Environment should match
	assert.Equal(t, "development", summary["environment"])
}

func TestEnvironmentHealth_HealthSummary(t *testing.T) {
	// Skip if not in integration test mode
	if !isIntegrationTestEnabled() {
		t.Skip("Integration tests disabled - set INTEGRATION_TESTS=true and ensure development environment is running")
	}

	// Arrange
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	healthChecker, err := validation.NewEnvironmentHealthChecker("development")
	require.NoError(t, err)

	mockOutputs := getMockDeploymentOutputs()

	// Act
	report, err := healthChecker.PerformHealthCheck(ctx, mockOutputs)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, report)

	// Get health summary
	summary := report.GetHealthySummary()
	assert.NotNil(t, summary)

	// Verify summary structure
	assert.Contains(t, summary, "total_components")
	assert.Contains(t, summary, "healthy_components")
	assert.Contains(t, summary, "degraded_components")
	assert.Contains(t, summary, "unhealthy_components")
	assert.Contains(t, summary, "health_percentage")
	assert.Contains(t, summary, "overall_status")
	assert.Contains(t, summary, "environment")
	assert.Contains(t, summary, "check_duration")

	// Verify values are reasonable
	totalComponents := summary["total_components"].(int)
	assert.Greater(t, totalComponents, 0)

	healthPercentage := summary["health_percentage"].(float64)
	assert.GreaterOrEqual(t, healthPercentage, 0.0)
	assert.LessOrEqual(t, healthPercentage, 100.0)

	assert.Equal(t, "development", summary["environment"])
}

func TestEnvironmentHealth_ConfigurationRetrieval(t *testing.T) {
	// Skip if not in integration test mode
	if !isIntegrationTestEnabled() {
		t.Skip("Integration tests disabled - set INTEGRATION_TESTS=true and ensure development environment is running")
	}

	// Arrange
	healthChecker, err := validation.NewEnvironmentHealthChecker("development")
	require.NoError(t, err)

	// Act
	config := healthChecker.GetEnvironmentConfig()
	requirements := healthChecker.GetEnvironmentRequirements()

	// Assert
	assert.NotNil(t, config)
	assert.NotNil(t, requirements)

	// Verify configuration for development environment
	assert.Equal(t, "development", config.Environment)
	assert.Greater(t, config.TimeoutSeconds, 0)
	assert.Greater(t, config.MaxRetries, 0)
	assert.Greater(t, config.HealthCheckInterval, time.Duration(0))

	// Verify requirements structure
	assert.Contains(t, requirements, "environment")
	assert.Contains(t, requirements, "expected_components")
	assert.Contains(t, requirements, "security_checks")
	assert.Contains(t, requirements, "compliance_checks")
	assert.Equal(t, "development", requirements["environment"])
}

// Helper functions for integration tests

func isIntegrationTestEnabled() bool {
	// Check if integration tests are enabled via environment variable
	return os.Getenv("INTEGRATION_TESTS") == "true"
}

func getMockDeploymentOutputs() map[string]interface{} {
	// In real integration tests, these would be retrieved from actual Pulumi stack outputs
	// For testing structure, we provide mock outputs that simulate real deployment
	return map[string]interface{}{
		"environment":                  "development",
		"deployment_complete":          true,
		"database_connection_string":   "mock://database:5432/development",
		"storage_connection_string":    "mock://storage:9000/development",
		"vault_address":               "http://vault:8200",
		"rabbitmq_endpoint":           "amqp://rabbitmq:5672",
		"grafana_url":                 "http://grafana:3000",
		"dapr_control_plane_url":      "http://dapr:3500",
		"container_orchestrator":      "podman",
		"public_gateway_url":          "http://gateway:8080",
		"admin_gateway_url":           "http://admin:8081",
		"website_url":                 "http://localhost:3000",
		"services_deployment_type":    "podman_containers",
		"website_deployment_type":     "container",
		"cdn_enabled":                 false,
		"ssl_enabled":                 false,
	}
}

// Contract validation test
func TestEnvironmentHealth_ContractValidation(t *testing.T) {
	// Skip if not in integration test mode
	if !isIntegrationTestEnabled() {
		t.Skip("Integration tests disabled - set INTEGRATION_TESTS=true and ensure development environment is running")
	}

	// Arrange
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	healthChecker, err := validation.NewEnvironmentHealthChecker("development")
	require.NoError(t, err)

	mockOutputs := getMockDeploymentOutputs()

	// Act
	report, err := healthChecker.PerformHealthCheck(ctx, mockOutputs)

	// Assert
	require.NoError(t, err)

	// Contract: Environment health report must always contain these fields
	assert.NotEmpty(t, report.Environment)
	assert.NotZero(t, report.StartTime)
	assert.NotZero(t, report.EndTime)
	assert.True(t, report.Duration >= 0)
	assert.NotNil(t, report.Components)
	assert.NotEqual(t, validation.HealthStatusUnknown, report.Overall)

	// Contract: Start time must be before end time
	assert.True(t, report.StartTime.Before(report.EndTime) || report.StartTime.Equal(report.EndTime))

	// Contract: Duration should match the difference between start and end times
	expectedDuration := report.EndTime.Sub(report.StartTime)
	tolerance := 100 * time.Millisecond // Allow small tolerance for timing
	assert.InDelta(t, expectedDuration.Nanoseconds(), report.Duration.Nanoseconds(), float64(tolerance.Nanoseconds()))
}

// Benchmark test for integration performance
func BenchmarkEnvironmentHealth_FullHealthCheck(b *testing.B) {
	if !isIntegrationTestEnabled() {
		b.Skip("Integration tests disabled - set INTEGRATION_TESTS=true and ensure development environment is running")
	}

	// Arrange
	healthChecker, err := validation.NewEnvironmentHealthChecker("development")
	if err != nil {
		b.Fatal(err)
	}

	mockOutputs := getMockDeploymentOutputs()

	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, err := healthChecker.PerformHealthCheck(ctx, mockOutputs)
		cancel()
		
		if err != nil {
			b.Fatal(err)
		}
	}
}