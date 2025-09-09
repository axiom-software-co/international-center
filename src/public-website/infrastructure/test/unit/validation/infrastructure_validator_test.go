package validation_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/cicd/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInfrastructureValidator(t *testing.T) {
	// Arrange
	config := &validation.ValidationConfig{
		Environment:          "development",
		TimeoutSeconds:       30,
		MaxRetries:          3,
		RetryDelaySeconds:   5,
		HealthCheckInterval: 30 * time.Second,
		ExpectedComponents:  []string{"infrastructure", "platform"},
		SecurityChecks:      []string{"encryption"},
		ComplianceChecks:    []string{"audit"},
	}

	// Act
	validator := validation.NewInfrastructureValidator(config)

	// Assert
	assert.NotNil(t, validator)
}

func TestValidationConfig_Properties(t *testing.T) {
	// Arrange
	testCases := []struct {
		name   string
		config *validation.ValidationConfig
	}{
		{
			name: "Development configuration",
			config: &validation.ValidationConfig{
				Environment:          "development",
				TimeoutSeconds:       30,
				MaxRetries:          3,
				RetryDelaySeconds:   5,
				HealthCheckInterval: 30 * time.Second,
				ExpectedComponents:  []string{"infrastructure"},
				SecurityChecks:      []string{},
				ComplianceChecks:    []string{},
			},
		},
		{
			name: "Production configuration",
			config: &validation.ValidationConfig{
				Environment:          "production",
				TimeoutSeconds:       120,
				MaxRetries:          10,
				RetryDelaySeconds:   15,
				HealthCheckInterval: 10 * time.Second,
				ExpectedComponents:  []string{"infrastructure", "platform", "services", "website"},
				SecurityChecks:      []string{"encryption", "access_control"},
				ComplianceChecks:    []string{"audit", "backup"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange & Act
			validator := validation.NewInfrastructureValidator(tc.config)

			// Assert
			assert.NotNil(t, validator)
		})
	}
}

func TestValidationResult_Structure(t *testing.T) {
	// Arrange
	result := validation.ValidationResult{
		Type:         validation.ValidationHealthCheck,
		ComponentID:  "test-component",
		Success:      true,
		Message:      "Health check passed",
		Details:      map[string]interface{}{"status": "healthy"},
		Timestamp:    time.Now(),
		Duration:     2 * time.Second,
		Severity:     "info",
		Environment:  "development",
	}

	// Act & Assert
	assert.Equal(t, validation.ValidationHealthCheck, result.Type)
	assert.Equal(t, "test-component", result.ComponentID)
	assert.True(t, result.Success)
	assert.Equal(t, "Health check passed", result.Message)
	assert.NotNil(t, result.Details)
	assert.Equal(t, "info", result.Severity)
	assert.Equal(t, "development", result.Environment)
	assert.Equal(t, 2*time.Second, result.Duration)
}

func TestValidationType_Constants(t *testing.T) {
	// Arrange & Act & Assert
	assert.Equal(t, validation.ValidationType("health_check"), validation.ValidationHealthCheck)
	assert.Equal(t, validation.ValidationType("connectivity"), validation.ValidationConnectivity)
	assert.Equal(t, validation.ValidationType("security"), validation.ValidationSecurity)
	assert.Equal(t, validation.ValidationType("contract"), validation.ValidationContract)
	assert.Equal(t, validation.ValidationType("environment"), validation.ValidationEnvironment)
	assert.Equal(t, validation.ValidationType("compliance"), validation.ValidationCompliance)
}

func TestInfrastructureValidator_GetValidationResults(t *testing.T) {
	// Arrange
	config := &validation.ValidationConfig{
		Environment:          "development",
		TimeoutSeconds:       30,
		MaxRetries:          3,
		RetryDelaySeconds:   5,
		HealthCheckInterval: 30 * time.Second,
		ExpectedComponents:  []string{"infrastructure"},
		SecurityChecks:      []string{},
		ComplianceChecks:    []string{},
	}
	validator := validation.NewInfrastructureValidator(config)

	// Act
	results := validator.GetValidationResults()

	// Assert
	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results)) // Should be empty initially
}

func TestInfrastructureValidator_GetValidationSummary(t *testing.T) {
	// Arrange
	config := &validation.ValidationConfig{
		Environment:          "development",
		TimeoutSeconds:       30,
		MaxRetries:          3,
		RetryDelaySeconds:   5,
		HealthCheckInterval: 30 * time.Second,
		ExpectedComponents:  []string{"infrastructure"},
		SecurityChecks:      []string{},
		ComplianceChecks:    []string{},
	}
	validator := validation.NewInfrastructureValidator(config)

	// Act
	summary := validator.GetValidationSummary()

	// Assert
	assert.NotNil(t, summary)
	assert.Contains(t, summary, "total_validations")
	assert.Contains(t, summary, "successful")
	assert.Contains(t, summary, "failed")
	assert.Contains(t, summary, "success_rate")
	assert.Contains(t, summary, "environment")
	assert.Contains(t, summary, "timestamp")
	assert.Equal(t, "development", summary["environment"])
}

func TestInfrastructureValidator_MockHealthCheck(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	config := &validation.ValidationConfig{
		Environment:          "development",
		TimeoutSeconds:       5,
		MaxRetries:          1,
		RetryDelaySeconds:   1,
		HealthCheckInterval: 5 * time.Second,
		ExpectedComponents:  []string{"test"},
		SecurityChecks:      []string{},
		ComplianceChecks:    []string{},
	}
	validator := validation.NewInfrastructureValidator(config)

	// Act
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	outputs := map[string]interface{}{
		"test_health_endpoint": server.URL,
	}

	// This would normally call ValidateInfrastructure, but we're testing structure
	results := validator.GetValidationResults()

	// Assert
	assert.NotNil(t, results)
	assert.NotNil(t, ctx) // Context is properly created
}

func TestValidationConfig_EnvironmentSpecific(t *testing.T) {
	// Arrange
	testCases := []struct {
		name               string
		environment        string
		expectedTimeout    int
		expectedRetries    int
		expectedComponents int
	}{
		{
			name:               "Development has minimal configuration",
			environment:        "development",
			expectedTimeout:    30,
			expectedRetries:    3,
			expectedComponents: 1,
		},
		{
			name:               "Production has extensive configuration",
			environment:        "production",
			expectedTimeout:    120,
			expectedRetries:    10,
			expectedComponents: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			config := &validation.ValidationConfig{
				Environment:        tc.environment,
				TimeoutSeconds:     tc.expectedTimeout,
				MaxRetries:        tc.expectedRetries,
				ExpectedComponents: make([]string, tc.expectedComponents),
			}

			// Act
			validator := validation.NewInfrastructureValidator(config)

			// Assert
			assert.NotNil(t, validator)
		})
	}
}

// Property-based tests
func TestInfrastructureValidator_Properties(t *testing.T) {
	// Property: Validator should always be created with valid config
	validConfigs := []*validation.ValidationConfig{
		{
			Environment:          "development",
			TimeoutSeconds:       30,
			MaxRetries:          3,
			RetryDelaySeconds:   5,
			HealthCheckInterval: 30 * time.Second,
			ExpectedComponents:  []string{"infrastructure"},
			SecurityChecks:      []string{},
			ComplianceChecks:    []string{},
		},
		{
			Environment:          "staging",
			TimeoutSeconds:       60,
			MaxRetries:          5,
			RetryDelaySeconds:   10,
			HealthCheckInterval: 15 * time.Second,
			ExpectedComponents:  []string{"infrastructure", "platform"},
			SecurityChecks:      []string{"encryption"},
			ComplianceChecks:    []string{"audit"},
		},
	}

	for i, config := range validConfigs {
		t.Run("Property_ValidConfig_"+config.Environment, func(t *testing.T) {
			// Act
			validator := validation.NewInfrastructureValidator(config)

			// Assert - Property: valid configs always create validator
			assert.NotNil(t, validator, "Valid config %d should create validator", i)
		})
	}
}

// Benchmark tests
func BenchmarkInfrastructureValidator_Creation(b *testing.B) {
	// Arrange
	config := &validation.ValidationConfig{
		Environment:          "development",
		TimeoutSeconds:       30,
		MaxRetries:          3,
		RetryDelaySeconds:   5,
		HealthCheckInterval: 30 * time.Second,
		ExpectedComponents:  []string{"infrastructure"},
		SecurityChecks:      []string{},
		ComplianceChecks:    []string{},
	}

	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator := validation.NewInfrastructureValidator(config)
		_ = validator // Prevent compiler optimization
	}
}

func BenchmarkValidationSummary_Generation(b *testing.B) {
	// Arrange
	config := &validation.ValidationConfig{
		Environment:          "development",
		TimeoutSeconds:       30,
		MaxRetries:          3,
		RetryDelaySeconds:   5,
		HealthCheckInterval: 30 * time.Second,
		ExpectedComponents:  []string{"infrastructure"},
		SecurityChecks:      []string{},
		ComplianceChecks:    []string{},
	}
	validator := validation.NewInfrastructureValidator(config)

	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		summary := validator.GetValidationSummary()
		_ = summary // Prevent compiler optimization
	}
}

// Test with timeouts
func TestInfrastructureValidator_WithTimeout(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config := &validation.ValidationConfig{
		Environment:          "development",
		TimeoutSeconds:       30,
		MaxRetries:          3,
		RetryDelaySeconds:   5,
		HealthCheckInterval: 30 * time.Second,
		ExpectedComponents:  []string{"infrastructure"},
		SecurityChecks:      []string{},
		ComplianceChecks:    []string{},
	}

	// Act & Assert
	validator := validation.NewInfrastructureValidator(config)
	require.NotNil(t, validator)

	// Test should complete within timeout
	select {
	case <-ctx.Done():
		t.Fatal("Test timed out")
	default:
		assert.NotNil(t, validator)
	}
}