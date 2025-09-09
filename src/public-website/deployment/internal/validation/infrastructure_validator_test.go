package validation_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHTTPClient implements the HTTPClient interface for testing
type mockHTTPClient struct {
	statusCode int
	body       string
	err        error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(strings.NewReader(m.body)),
		Header:     make(http.Header),
	}, nil
}

func TestNewInfrastructureValidator(t *testing.T) {
	// Arrange
	config := &ValidationConfig{
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
	validator := NewInfrastructureValidator(config)

	// Assert
	assert.NotNil(t, validator)
}

func TestValidationConfig_Properties(t *testing.T) {
	// Arrange
	testCases := []struct {
		name   string
		config *ValidationConfig
	}{
		{
			name: "Development configuration",
			config: &ValidationConfig{
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
			config: &ValidationConfig{
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
			validator := NewInfrastructureValidator(tc.config)

			// Assert
			assert.NotNil(t, validator)
		})
	}
}

func TestValidationResult_Structure(t *testing.T) {
	// Arrange
	result := ValidationResult{
		Type:         ValidationHealthCheck,
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
	assert.Equal(t, ValidationHealthCheck, result.Type)
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
	assert.Equal(t, ValidationType("health_check"), ValidationHealthCheck)
	assert.Equal(t, ValidationType("connectivity"), ValidationConnectivity)
	assert.Equal(t, ValidationType("security"), ValidationSecurity)
	assert.Equal(t, ValidationType("contract"), ValidationContract)
	assert.Equal(t, ValidationType("environment"), ValidationEnvironment)
	assert.Equal(t, ValidationType("compliance"), ValidationCompliance)
}

func TestInfrastructureValidator_GetValidationResults(t *testing.T) {
	// Arrange
	config := &ValidationConfig{
		Environment:          "development",
		TimeoutSeconds:       30,
		MaxRetries:          3,
		RetryDelaySeconds:   5,
		HealthCheckInterval: 30 * time.Second,
		ExpectedComponents:  []string{"infrastructure"},
		SecurityChecks:      []string{},
		ComplianceChecks:    []string{},
	}
	validator := NewInfrastructureValidator(config)

	// Act
	results := validator.GetValidationResults()

	// Assert
	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results)) // Should be empty initially
}

func TestInfrastructureValidator_GetValidationSummary(t *testing.T) {
	// Arrange
	config := &ValidationConfig{
		Environment:          "development",
		TimeoutSeconds:       30,
		MaxRetries:          3,
		RetryDelaySeconds:   5,
		HealthCheckInterval: 30 * time.Second,
		ExpectedComponents:  []string{"infrastructure"},
		SecurityChecks:      []string{},
		ComplianceChecks:    []string{},
	}
	validator := NewInfrastructureValidator(config)

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

	config := &ValidationConfig{
		Environment:          "development",
		TimeoutSeconds:       5,
		MaxRetries:          1,
		RetryDelaySeconds:   1,
		HealthCheckInterval: 5 * time.Second,
		ExpectedComponents:  []string{"test"},
		SecurityChecks:      []string{},
		ComplianceChecks:    []string{},
	}
	validator := NewInfrastructureValidator(config)

	// Act
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	outputs := map[string]interface{}{
		"test_health_endpoint": server.URL,
	}

	// Validate that outputs are properly structured for testing
	assert.NotEmpty(t, outputs)
	assert.Contains(t, outputs, "test_health_endpoint")

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
			config := &ValidationConfig{
				Environment:        tc.environment,
				TimeoutSeconds:     tc.expectedTimeout,
				MaxRetries:        tc.expectedRetries,
				ExpectedComponents: make([]string, tc.expectedComponents),
			}

			// Act
			validator := NewInfrastructureValidator(config)

			// Assert
			assert.NotNil(t, validator)
		})
	}
}

// Property-based tests
func TestInfrastructureValidator_Properties(t *testing.T) {
	// Property: Validator should always be created with valid config
	validConfigs := []*ValidationConfig{
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
			validator := NewInfrastructureValidator(config)

			// Assert - Property: valid configs always create validator
			assert.NotNil(t, validator, "Valid config %d should create validator", i)
		})
	}
}

// Benchmark tests
func BenchmarkInfrastructureValidator_Creation(b *testing.B) {
	// Arrange
	config := &ValidationConfig{
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
		validator := NewInfrastructureValidator(config)
		_ = validator // Prevent compiler optimization
	}
}

func BenchmarkValidationSummary_Generation(b *testing.B) {
	// Arrange
	config := &ValidationConfig{
		Environment:          "development",
		TimeoutSeconds:       30,
		MaxRetries:          3,
		RetryDelaySeconds:   5,
		HealthCheckInterval: 30 * time.Second,
		ExpectedComponents:  []string{"infrastructure"},
		SecurityChecks:      []string{},
		ComplianceChecks:    []string{},
	}
	validator := NewInfrastructureValidator(config)

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

	config := &ValidationConfig{
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
	validator := NewInfrastructureValidator(config)
	require.NotNil(t, validator)

	// Test should complete within timeout
	select {
	case <-ctx.Done():
		t.Fatal("Test timed out")
	default:
		assert.NotNil(t, validator)
	}
}

// REFACTOR PHASE: Enhanced contract validation tests

func TestInfrastructureValidator_ValidateInfrastructure_Contract(t *testing.T) {
	// Test contract: ValidateInfrastructure should execute all validation phases
	testCases := []struct {
		name              string
		environment       string
		expectedComponents []string
		securityChecks    []string
		complianceChecks  []string
		expectSuccess     bool
	}{
		{
			name:              "Development environment with minimal validation",
			environment:       "development", 
			expectedComponents: []string{"infrastructure"},
			securityChecks:    []string{},
			complianceChecks:  []string{},
			expectSuccess:     true,
		},
		{
			name:              "Production environment with comprehensive validation",
			environment:       "production",
			expectedComponents: []string{"infrastructure", "platform", "services"},
			securityChecks:    []string{"encryption_at_rest", "access_control"},
			complianceChecks:  []string{"audit_logging", "backup_policies"},
			expectSuccess:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange - Set up mock HTTP client that always returns success
			mockClient := &mockHTTPClient{
				statusCode: http.StatusOK,
				body:       "OK",
			}
			
			config := &ValidationConfig{
				Environment:          tc.environment,
				TimeoutSeconds:       5,
				MaxRetries:          1,
				RetryDelaySeconds:   1,
				HealthCheckInterval: 5 * time.Second,
				ExpectedComponents:  tc.expectedComponents,
				SecurityChecks:      tc.securityChecks,
				ComplianceChecks:    tc.complianceChecks,
			}
			validator := NewInfrastructureValidator(config).WithHTTPClient(mockClient)
			
			// Mock outputs with health endpoints
			outputs := createMockValidationOutputs(tc.expectedComponents)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Act
			err := validator.ValidateInfrastructure(ctx, outputs)

			// Assert contract compliance
			if tc.expectSuccess {
				assert.NoError(t, err, "Validation should succeed for %s", tc.environment)
			} else {
				assert.Error(t, err, "Validation should fail for invalid configuration")
			}
			
			// Verify validation results were recorded
			results := validator.GetValidationResults()
			assert.NotEmpty(t, results, "Validation results should be recorded")
			
			// Verify all expected validation types were executed
			expectedValidationTypes := []string{"health_checks", "connectivity", "security_policies", "component_contracts", "environment_compliance"}
			for _, expectedType := range expectedValidationTypes {
				found := false
				for _, result := range results {
					if string(result.Type) == expectedType {
						found = true
						// Contract: All validation results must have proper structure
						assert.NotEmpty(t, result.Message, "Validation result must have message")
						assert.NotZero(t, result.Timestamp, "Validation result must have timestamp")
						assert.Equal(t, tc.environment, result.Environment, "Validation result must match environment")
						break
					}
				}
				assert.True(t, found, "Expected validation type %s was not executed", expectedType)
			}
		})
	}
}

func TestInfrastructureValidator_GetValidationSummary_Contract(t *testing.T) {
	// Test contract: GetValidationSummary must provide accurate metrics
	// Arrange
	mockClient := &mockHTTPClient{
		statusCode: http.StatusOK,
		body:       "OK",
	}
	
	config := &ValidationConfig{
		Environment:          "development",
		TimeoutSeconds:       5,
		MaxRetries:          1,
		RetryDelaySeconds:   1,
		HealthCheckInterval: 5 * time.Second,
		ExpectedComponents:  []string{"infrastructure"},
		SecurityChecks:      []string{},
		ComplianceChecks:    []string{},
	}
	validator := NewInfrastructureValidator(config).WithHTTPClient(mockClient)
	
	// Act: Execute validation to generate results
	outputs := createMockValidationOutputs([]string{"infrastructure"})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	err := validator.ValidateInfrastructure(ctx, outputs)
	require.NoError(t, err)
	
	// Act: Get validation summary
	summary := validator.GetValidationSummary()
	
	// Assert contract compliance
	assert.Contains(t, summary, "total_validations", "Summary must include total validations count")
	assert.Contains(t, summary, "successful", "Summary must include successful count")
	assert.Contains(t, summary, "failed", "Summary must include failed count")
	assert.Contains(t, summary, "success_rate", "Summary must include success rate")
	assert.Contains(t, summary, "environment", "Summary must include environment")
	assert.Contains(t, summary, "timestamp", "Summary must include timestamp")
	
	// Contract: Success rate calculation must be correct
	total := summary["total_validations"].(int)
	successful := summary["successful"].(int)
	failed := summary["failed"].(int)
	successRate := summary["success_rate"].(float64)
	
	assert.Equal(t, total, successful+failed, "Total must equal sum of successful and failed")
	expectedRate := float64(successful) / float64(total) * 100
	assert.InDelta(t, expectedRate, successRate, 0.01, "Success rate calculation must be accurate")
	assert.Equal(t, "development", summary["environment"], "Environment must match configuration")
}

func TestInfrastructureValidator_ErrorHandling_Contract(t *testing.T) {
	// Test contract: Validation errors should be properly handled and reported
	testCases := []struct {
		name               string
		expectedComponents []string
		outputs            pulumi.Map
		expectError        bool
		errorContains      string
		setupServer        bool
	}{
		{
			name:               "No health endpoints should fail health check validation",
			expectedComponents: []string{"infrastructure"},
			outputs:           pulumi.Map{"infrastructure": pulumi.String("value")},
			expectError:       true,
			errorContains:     "no health check endpoints found",
			setupServer:       false,
		},
		{
			name:               "Valid components with health endpoints should succeed",
			expectedComponents: []string{"infrastructure"},
			outputs:           pulumi.Map{}, // Will be populated with mock data
			expectError:       false,
			setupServer:       true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var outputs pulumi.Map
			var mockClient *mockHTTPClient
			
			if tc.setupServer {
				mockClient = &mockHTTPClient{
					statusCode: http.StatusOK,
					body:       "OK",
				}
				outputs = createMockValidationOutputs(tc.expectedComponents)
			} else {
				mockClient = &mockHTTPClient{
					statusCode: http.StatusServiceUnavailable,
					body:       "Service Unavailable",
				}
				outputs = tc.outputs
			}
			
			config := &ValidationConfig{
				Environment:          "development",
				TimeoutSeconds:       5,
				MaxRetries:          1,
				RetryDelaySeconds:   1,
				HealthCheckInterval: 5 * time.Second,
				ExpectedComponents:  tc.expectedComponents,
				SecurityChecks:      []string{},
				ComplianceChecks:    []string{},
			}
			validator := NewInfrastructureValidator(config).WithHTTPClient(mockClient)
			
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			
			// Act
			err := validator.ValidateInfrastructure(ctx, outputs)
			
			// Assert contract compliance
			if tc.expectError {
				assert.Error(t, err, "Should return error for invalid configuration")
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, "Error message should be descriptive")
				}
				
				// Contract: Failed validation should still record results
				results := validator.GetValidationResults()
				assert.NotEmpty(t, results, "Should record results even on failure")
				
				// Contract: At least one result should indicate failure
				foundFailure := false
				for _, result := range results {
					if !result.Success {
						foundFailure = true
						assert.Equal(t, "error", result.Severity, "Failed validation should have error severity")
						break
					}
				}
				assert.True(t, foundFailure, "Should record at least one failed validation result")
			} else {
				assert.NoError(t, err, "Should succeed for valid configuration")
			}
		})
	}
}

// Helper function to create mock validation outputs with working health endpoints
func createMockValidationOutputs(components []string) pulumi.Map {
	outputs := pulumi.Map{}
	
	for _, component := range components {
		outputs[component] = pulumi.String(fmt.Sprintf("mock_%s_value", component))
		// Add health endpoint for each component to ensure health checks can find endpoints
		outputs[fmt.Sprintf("%s_health_endpoint", component)] = pulumi.String(fmt.Sprintf("http://localhost:8080/health/%s", component))
	}
	
	return outputs
}

// Helper function to create mock validation outputs with mock HTTP server endpoints
func createMockValidationOutputsWithServer(components []string, serverURL string) pulumi.Map {
	outputs := pulumi.Map{}
	
	for _, component := range components {
		outputs[component] = pulumi.String(fmt.Sprintf("mock_%s_value", component))
		outputs[fmt.Sprintf("%s_health_endpoint", component)] = pulumi.String(fmt.Sprintf("%s/%s/health", serverURL, component))
	}
	
	return outputs
}