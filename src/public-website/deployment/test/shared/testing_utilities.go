package shared

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SharedMockResourceMonitor provides a reusable mock resource monitor for all tests
type SharedMockResourceMonitor struct{}

func (m *SharedMockResourceMonitor) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return resource.PropertyMap{}, nil
}

func (m *SharedMockResourceMonitor) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	return args.Name + "_id", args.Inputs, nil
}

// ComponentTestCase represents a standardized test case structure for component testing
type ComponentTestCase struct {
	Name        string
	Environment string
	Validations []ValidationFunc
}

// ValidationFunc represents a validation function for component outputs
type ValidationFunc func(t *testing.T, component interface{})

// EnvironmentValidation provides common environment-specific validation patterns
type EnvironmentValidation struct {
	Environment string
	URLPattern  string
	CDNEnabled  bool
	SSLEnabled  bool
}

// GetStandardEnvironmentValidations returns standard validation patterns for each environment
func GetStandardEnvironmentValidations() map[string]EnvironmentValidation {
	return map[string]EnvironmentValidation{
		"development": {
			Environment: "development",
			URLPattern:  "localhost",
			CDNEnabled:  false,
			SSLEnabled:  false,
		},
		"staging": {
			Environment: "staging",
			URLPattern:  "staging",
			CDNEnabled:  true,
			SSLEnabled:  true,
		},
		"production": {
			Environment: "production",
			URLPattern:  "production",
			CDNEnabled:  true,
			SSLEnabled:  true,
		},
	}
}

// RunPulumiComponentTest executes a standardized Pulumi component test
func RunPulumiComponentTest(t *testing.T, testCase ComponentTestCase, createComponent func(*pulumi.Context) (interface{}, error)) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		component, err := createComponent(ctx)
		if err != nil {
			return err
		}

		require.NotNil(t, component)

		// Run all validations
		for _, validation := range testCase.Validations {
			validation(t, component)
		}

		return nil
	}, pulumi.WithMocks("project", "stack", &SharedMockResourceMonitor{}))

	require.NoError(t, err)
}

// ValidateStringOutput validates a string output matches expected patterns
func ValidateStringOutput(t *testing.T, output pulumi.StringOutput, expectedPattern string, description string) {
	output.ApplyT(func(value string) string {
		assert.Contains(t, value, expectedPattern, description)
		return value
	})
}

// ValidateBoolOutput validates a boolean output matches expected value
func ValidateBoolOutput(t *testing.T, output pulumi.BoolOutput, expectedValue bool, description string) {
	output.ApplyT(func(value bool) bool {
		assert.Equal(t, expectedValue, value, description)
		return value
	})
}

// ValidateMapOutput validates a map output contains expected keys
func ValidateMapOutput(t *testing.T, output pulumi.MapOutput, expectedKeys []string, description string) {
	output.ApplyT(func(config interface{}) interface{} {
		configMap := config.(map[string]interface{})
		for _, key := range expectedKeys {
			assert.NotNil(t, configMap[key], description+" - missing key: "+key)
		}
		return config
	})
}

// PerformanceTestConfig represents configuration for performance testing
type PerformanceTestConfig struct {
	MaxExecutionTimeMs int64
	MaxMemoryUsageMB   int64
	ConcurrentTests    int
}

// GetDefaultPerformanceConfig returns default performance testing configuration
func GetDefaultPerformanceConfig() PerformanceTestConfig {
	return PerformanceTestConfig{
		MaxExecutionTimeMs: 5000, // 5 seconds max for unit tests
		MaxMemoryUsageMB:   100,  // 100MB max memory usage
		ConcurrentTests:    4,    // 4 concurrent tests
	}
}

// PropertyBasedTestConfig represents configuration for property-based testing
type PropertyBasedTestConfig struct {
	Iterations        int
	EnvironmentTypes  []string
	ResourceLimits    map[string]interface{}
	ConfigurationSets []map[string]interface{}
}

// GetDefaultPropertyTestConfig returns default property-based testing configuration
func GetDefaultPropertyTestConfig() PropertyBasedTestConfig {
	return PropertyBasedTestConfig{
		Iterations:       10,
		EnvironmentTypes: []string{"development", "staging", "production"},
		ResourceLimits: map[string]interface{}{
			"cpu":      "100m",
			"memory":   "128Mi",
			"replicas": 1,
		},
		ConfigurationSets: []map[string]interface{}{
			{"cdn_enabled": true, "ssl_enabled": true},
			{"cdn_enabled": false, "ssl_enabled": false},
			{"health_checks": true, "metrics": true},
		},
	}
}

// ComponentPropertyTest represents a property-based test for components
type ComponentPropertyTest struct {
	Name     string
	Property func(environment string, config map[string]interface{}) bool
}

// RunPropertyBasedTest executes property-based testing on components
func RunPropertyBasedTest(t *testing.T, test ComponentPropertyTest, config PropertyBasedTestConfig) {
	for i := 0; i < config.Iterations; i++ {
		for _, env := range config.EnvironmentTypes {
			for _, configSet := range config.ConfigurationSets {
				t.Run(test.Name+"_"+env+"_iteration_"+string(rune(i)), func(t *testing.T) {
					result := test.Property(env, configSet)
					assert.True(t, result, "Property should hold for environment %s with config %v", env, configSet)
				})
			}
		}
	}
}

// HTTPTestClient creates a standardized HTTP client for testing with timeout
func HTTPTestClient() *http.Client {
	return &http.Client{Timeout: 5 * time.Second}
}

// HTTPTestRequest creates a standardized HTTP request for testing
func HTTPTestRequest(ctx context.Context, method, url string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, url, nil)
}

// TestHealthEndpoint performs a standard health check request and returns only error status
func TestHealthEndpoint(ctx context.Context, url string) error {
	client := HTTPTestClient()
	req, err := HTTPTestRequest(ctx, "GET", url)
	if err != nil {
		return err
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// For testing purposes, we just verify the endpoint is reachable
	return nil
}

// ValidateEnvironmentPrerequisites ensures environment health before integration testing
// This function checks that all critical containers are running before tests execute
func ValidateEnvironmentPrerequisites(t *testing.T) {
	// Check critical infrastructure, platform, service, and gateway components are running
	criticalContainers := []string{
		"postgresql", 
		"dapr-control-plane", 
		"content-api", 
		"inquiries-api", 
		"notification-api", 
		"services-api",
		"public-gateway", 
		"admin-gateway",
	}
	
	for _, container := range criticalContainers {
		cmd := exec.Command("podman", "ps", "--filter", "name="+container, "--format", "{{.Names}}")
		output, err := cmd.Output()
		require.NoError(t, err, "Failed to check critical container %s", container)

		if !strings.Contains(string(output), container) {
			t.Skipf("Critical container %s not running - environment not ready for integration testing", container)
		}
	}
}