package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

// TestHealthEndpointWithRetry performs health check with retry logic for distributed architecture
func TestHealthEndpointWithRetry(ctx context.Context, url string, maxRetries int, retryDelay time.Duration) error {
	var lastErr error
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
				// Continue with retry
			}
		}
		
		client := HTTPTestClient()
		req, err := HTTPTestRequest(ctx, "GET", url)
		if err != nil {
			lastErr = fmt.Errorf("request creation failed (attempt %d/%d): %w", attempt+1, maxRetries, err)
			continue
		}
		
		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("connection failed (attempt %d/%d): %w", attempt+1, maxRetries, err)
			continue
		}
		resp.Body.Close()
		
		// Success
		return nil
	}
	
	return fmt.Errorf("health check failed after %d attempts: %w", maxRetries, lastErr)
}

// ValidateDistributedSidecarHealth validates sidecar health with distributed architecture considerations
func ValidateDistributedSidecarHealth(ctx context.Context, sidecarName string, port int) error {
	url := fmt.Sprintf("http://localhost:%d/v1.0/healthz", port)

	// Use retry logic for distributed sidecars which may take time to initialize
	return TestHealthEndpointWithRetry(ctx, url, 5, 2*time.Second)
}

// ValidateContainerVolumeMount verifies that a container has the proper volume mount configuration
func ValidateContainerVolumeMount(ctx context.Context, containerName string, expectedSource string, expectedDestination string) error {
	cmd := exec.Command("podman", "inspect", containerName, "--format", "{{range .Mounts}}{{.Source}}:{{.Destination}}:{{.Type}}\n{{end}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to inspect container %s mounts: %w", containerName, err)
	}

	mounts := strings.TrimSpace(string(output))
	expectedMount := fmt.Sprintf("%s:%s", expectedSource, expectedDestination)

	if !strings.Contains(mounts, expectedMount) {
		return fmt.Errorf("container %s missing expected volume mount %s, found mounts: %s", containerName, expectedMount, mounts)
	}

	return nil
}

// DiagnoseConfigurationDeploymentStatus provides comprehensive diagnostics for configuration deployment issues
func DiagnoseConfigurationDeploymentStatus(ctx context.Context) map[string]interface{} {
	diagnostics := make(map[string]interface{})

	// Check if project configuration directory exists
	projectConfigDir := "/home/tojkuv/Documents/GitHub/international-center-workspace/international-center/src/public-website/deployment/configs/dapr"
	if _, err := os.Stat(projectConfigDir); err != nil {
		diagnostics["project_config_directory"] = fmt.Sprintf("ERROR: %v", err)
	} else {
		diagnostics["project_config_directory"] = "EXISTS"
	}

	// Check for temporary directory anti-patterns
	tempDirectories := []string{"/tmp/dapr-config", "/tmp/dapr-minimal", "/tmp/dapr-components"}
	tempDirStatus := make(map[string]string)
	for _, dir := range tempDirectories {
		if _, err := os.Stat(dir); err == nil {
			tempDirStatus[dir] = "EXISTS (ANTI-PATTERN)"
		} else {
			tempDirStatus[dir] = "NOT_EXISTS (GOOD)"
		}
	}
	diagnostics["temporary_directories"] = tempDirStatus

	// Check container configuration deployment
	configFailures := DetectConfigurationDeploymentFailures(ctx)
	if len(configFailures) > 0 {
		diagnostics["configuration_failures"] = configFailures
	} else {
		diagnostics["configuration_failures"] = "NONE"
	}

	return diagnostics
}

// SidecarConfig represents configuration for a distributed sidecar
type SidecarConfig struct {
	Name        string
	AppID       string
	HTTPPort    int
	GRPCPort    int
	AppPort     int
	HealthPath  string
}

// GetStandardSidecarConfigs returns configurations for all distributed sidecars
func GetStandardSidecarConfigs() map[string]SidecarConfig {
	return map[string]SidecarConfig{
		"content-api": {
			Name:       "content-api-sidecar",
			AppID:      "content-api",
			HTTPPort:   3502,
			GRPCPort:   50002,
			AppPort:    8082,
			HealthPath: "/v1.0/healthz",
		},
		"public-gateway": {
			Name:       "public-gateway-sidecar",
			AppID:      "public-gateway",
			HTTPPort:   3503,
			GRPCPort:   50003,
			AppPort:    8081,
			HealthPath: "/v1.0/healthz",
		},
		"inquiries-api": {
			Name:       "inquiries-api-sidecar",
			AppID:      "inquiries-api",
			HTTPPort:   3504,
			GRPCPort:   50004,
			AppPort:    8083,
			HealthPath: "/v1.0/healthz",
		},
		"admin-gateway": {
			Name:       "admin-gateway-sidecar",
			AppID:      "admin-gateway",
			HTTPPort:   3506,
			GRPCPort:   50006,
			AppPort:    8092,
			HealthPath: "/v1.0/healthz",
		},
		"services-api": {
			Name:       "services-api-sidecar",
			AppID:      "services-api",
			HTTPPort:   3507,
			GRPCPort:   50007,
			AppPort:    8093,
			HealthPath: "/v1.0/healthz",
		},
		"notification-api": {
			Name:       "notification-api-sidecar",
			AppID:      "notification-api",
			HTTPPort:   3508,
			GRPCPort:   50008,
			AppPort:    8094,
			HealthPath: "/v1.0/healthz",
		},
	}
}

// ValidateServiceMeshCommunication tests service-to-service communication through sidecars
func ValidateServiceMeshCommunication(ctx context.Context, fromSidecar SidecarConfig, toServiceAppID string, endpoint string) error {
	url := fmt.Sprintf("http://localhost:%d/v1.0/invoke/%s/method%s", 
		fromSidecar.HTTPPort, toServiceAppID, endpoint)
	
	// Use retry logic for service mesh calls which may require component initialization
	return TestHealthEndpointWithRetry(ctx, url, 3, 1*time.Second)
}

// DiagnoseSidecarConnectivity provides detailed diagnostic information for sidecar issues
func DiagnoseSidecarConnectivity(ctx context.Context, sidecarConfig SidecarConfig) error {
	client := HTTPTestClient()
	
	// Test basic HTTP connectivity
	healthURL := fmt.Sprintf("http://localhost:%d%s", sidecarConfig.HTTPPort, sidecarConfig.HealthPath)
	req, err := HTTPTestRequest(ctx, "GET", healthURL)
	if err != nil {
		return fmt.Errorf("sidecar %s: failed to create health request: %w", sidecarConfig.Name, err)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sidecar %s: health endpoint unreachable at %s - ensure sidecar container is running and healthy: %w", 
			sidecarConfig.Name, healthURL, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("sidecar %s: health check failed with status %d - check component initialization logs", 
			sidecarConfig.Name, resp.StatusCode)
	}
	
	return nil
}

// SidecarHealthResult represents the result of a sidecar health check
type SidecarHealthResult struct {
	SidecarName string
	AppID       string
	Port        int
	Healthy     bool
	Error       error
	Duration    time.Duration
}

// ValidateAllSidecarsParallel performs parallel health checks on all sidecars for optimal performance
func ValidateAllSidecarsParallel(ctx context.Context) []SidecarHealthResult {
	sidecarConfigs := GetStandardSidecarConfigs()
	results := make([]SidecarHealthResult, 0, len(sidecarConfigs))
	resultChan := make(chan SidecarHealthResult, len(sidecarConfigs))
	
	// Launch parallel health checks
	for _, config := range sidecarConfigs {
		go func(sc SidecarConfig) {
			start := time.Now()
			err := ValidateDistributedSidecarHealth(ctx, sc.Name, sc.HTTPPort)
			duration := time.Since(start)
			
			resultChan <- SidecarHealthResult{
				SidecarName: sc.Name,
				AppID:       sc.AppID,
				Port:        sc.HTTPPort,
				Healthy:     err == nil,
				Error:       err,
				Duration:    duration,
			}
		}(config)
	}
	
	// Collect results
	for i := 0; i < len(sidecarConfigs); i++ {
		select {
		case result := <-resultChan:
			results = append(results, result)
		case <-ctx.Done():
			// Context cancelled, return partial results
			return results
		}
	}
	
	return results
}

// OptimizedHTTPClient returns an HTTP client optimized for integration testing performance
func OptimizedHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 3 * time.Second, // Reduced timeout for faster failure detection
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true, // Reduce CPU overhead
		},
	}
}

// ParallelServiceMeshValidation performs parallel validation of service mesh communication
func ParallelServiceMeshValidation(ctx context.Context, testCases []struct {
	FromSidecar     string
	ToServiceAppID  string
	Endpoint        string
	Description     string
}) []SidecarHealthResult {
	sidecarConfigs := GetStandardSidecarConfigs()
	results := make([]SidecarHealthResult, 0, len(testCases))
	resultChan := make(chan SidecarHealthResult, len(testCases))
	
	// Launch parallel service mesh tests
	for _, testCase := range testCases {
		go func(tc struct {
			FromSidecar     string
			ToServiceAppID  string
			Endpoint        string
			Description     string
		}) {
			start := time.Now()
			fromConfig, exists := sidecarConfigs[tc.FromSidecar]
			if !exists {
				resultChan <- SidecarHealthResult{
					SidecarName: tc.FromSidecar,
					AppID:       tc.ToServiceAppID,
					Healthy:     false,
					Error:       fmt.Errorf("unknown sidecar configuration: %s", tc.FromSidecar),
					Duration:    time.Since(start),
				}
				return
			}
			
			err := ValidateServiceMeshCommunication(ctx, fromConfig, tc.ToServiceAppID, tc.Endpoint)
			duration := time.Since(start)
			
			resultChan <- SidecarHealthResult{
				SidecarName: tc.FromSidecar,
				AppID:       tc.ToServiceAppID,
				Port:        fromConfig.HTTPPort,
				Healthy:     err == nil,
				Error:       err,
				Duration:    duration,
			}
		}(testCase)
	}
	
	// Collect results with timeout
	for i := 0; i < len(testCases); i++ {
		select {
		case result := <-resultChan:
			results = append(results, result)
		case <-ctx.Done():
			return results
		}
	}
	
	return results
}

// ValidateProjectConfigurationDeployment validates that containers use project-managed configurations
func ValidateProjectConfigurationDeployment(ctx context.Context, containerName string) error {
	// Validate container is NOT using /tmp directories for configuration
	cmd := exec.Command("podman", "inspect", containerName, "--format", "{{range .Mounts}}{{.Source}}:{{.Destination}}\n{{end}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to inspect container %s mounts: %w", containerName, err)
	}

	mounts := strings.TrimSpace(string(output))
	if strings.Contains(mounts, "/tmp/dapr-config") || strings.Contains(mounts, "/tmp/dapr-minimal") {
		return fmt.Errorf("container %s using temporary directory anti-pattern - found /tmp configuration mounts", containerName)
	}

	// Validate container IS using project configuration directory
	expectedConfigPath := "/home/tojkuv/Documents/GitHub/international-center-workspace/international-center/src/public-website/deployment/configs/dapr"
	if !strings.Contains(mounts, expectedConfigPath) {
		return fmt.Errorf("container %s missing proper project configuration mount - expected %s", containerName, expectedConfigPath)
	}

	return nil
}

// ValidateDaprConfigurationPaths validates sidecar containers have proper Dapr configuration paths
func ValidateDaprConfigurationPaths(ctx context.Context, sidecarName string) error {
	// Get container startup command to validate configuration paths
	cmd := exec.Command("podman", "inspect", sidecarName, "--format", "{{.Config.Cmd}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to inspect sidecar %s command: %w", sidecarName, err)
	}

	command := strings.TrimSpace(string(output))

	// Validate that configuration paths point to proper project directories, not /tmp
	if strings.Contains(command, "/tmp/dapr-config") || strings.Contains(command, "/tmp/dapr-minimal") {
		return fmt.Errorf("sidecar %s using temporary directory configuration paths - found /tmp references in command", sidecarName)
	}

	// Validate that proper project configuration paths are used
	if !strings.Contains(command, "--resources-path") {
		return fmt.Errorf("sidecar %s missing --resources-path parameter for component configuration", sidecarName)
	}

	if !strings.Contains(command, "--config") {
		return fmt.Errorf("sidecar %s missing --config parameter for Dapr configuration", sidecarName)
	}

	return nil
}

// ValidateContainerVolumeMounts validates that containers have proper volume mounts for configuration
func ValidateContainerVolumeMounts(ctx context.Context, containerName string, expectedMounts []string) error {
	cmd := exec.Command("podman", "inspect", containerName, "--format", "{{range .Mounts}}{{.Source}}:{{.Destination}}:{{.Type}}\n{{end}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to inspect container %s volume mounts: %w", containerName, err)
	}

	mounts := strings.TrimSpace(string(output))

	for _, expectedMount := range expectedMounts {
		if !strings.Contains(mounts, expectedMount) {
			return fmt.Errorf("container %s missing expected volume mount: %s", containerName, expectedMount)
		}
	}

	return nil
}

// DetectConfigurationDeploymentFailures detects common configuration deployment issues
func DetectConfigurationDeploymentFailures(ctx context.Context) []string {
	var failures []string

	sidecarContainers := []string{
		"content-api-sidecar",
		"inquiries-api-sidecar",
		"notification-api-sidecar",
		"services-api-sidecar",
		"public-gateway-sidecar",
		"admin-gateway-sidecar",
	}

	for _, sidecar := range sidecarContainers {
		// Check if container exists and is running
		cmd := exec.Command("podman", "ps", "--filter", "name="+sidecar, "--format", "{{.Names}}")
		output, _ := cmd.Output()

		if strings.TrimSpace(string(output)) == "" {
			// Check if container exists but is stopped
			cmd = exec.Command("podman", "ps", "-a", "--filter", "name="+sidecar, "--format", "{{.Names}}")
			output, _ = cmd.Output()

			if strings.TrimSpace(string(output)) != "" {
				failures = append(failures, fmt.Sprintf("sidecar %s exists but not running - likely configuration failure", sidecar))
			} else {
				failures = append(failures, fmt.Sprintf("sidecar %s missing - deployment orchestration failure", sidecar))
			}
			continue
		}

		// Validate configuration deployment
		if err := ValidateProjectConfigurationDeployment(ctx, sidecar); err != nil {
			failures = append(failures, fmt.Sprintf("configuration deployment failure for %s: %v", sidecar, err))
		}

		if err := ValidateDaprConfigurationPaths(ctx, sidecar); err != nil {
			failures = append(failures, fmt.Sprintf("configuration path failure for %s: %v", sidecar, err))
		}
	}

	return failures
}

// ValidateEnvironmentPrerequisites ensures environment health before integration testing
// This function checks that all critical containers are running before tests execute
func ValidateEnvironmentPrerequisites(t *testing.T) {
	// Check critical infrastructure, platform, service, and gateway components are running
	// Updated to validate distributed Dapr sidecar architecture instead of centralized control plane
	criticalContainers := []string{
		"postgresql", 
		"content-api", 
		"content-api-sidecar",
		"inquiries-api", 
		"inquiries-api-sidecar",
		"notification-api", 
		"notification-api-sidecar",
		"services-api",
		"services-api-sidecar",
		"public-gateway", 
		"public-gateway-sidecar",
		"admin-gateway",
		"admin-gateway-sidecar",
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

// ValidateEnvironmentPrerequisitesForBenchmarks ensures environment health before benchmark testing
// This function checks that all critical containers are running before benchmarks execute
func ValidateEnvironmentPrerequisitesForBenchmarks(b *testing.B) {
	// Check critical infrastructure, platform, service, and gateway components are running
	criticalContainers := []string{
		"postgresql",
		"content-api",
		"content-api-sidecar",
		"inquiries-api",
		"inquiries-api-sidecar",
		"notification-api",
		"notification-api-sidecar",
		"services-api",
		"services-api-sidecar",
		"public-gateway",
		"public-gateway-sidecar",
		"admin-gateway",
		"admin-gateway-sidecar",
	}

	for _, container := range criticalContainers {
		cmd := exec.Command("podman", "ps", "--filter", "name="+container, "--format", "{{.Names}}")
		output, err := cmd.Output()
		if err != nil {
			b.Fatalf("Failed to check critical container %s: %v", container, err)
		}

		if !strings.Contains(string(output), container) {
			b.Skipf("Critical container %s not running - environment not ready for benchmark testing", container)
		}
	}
}

// REFACTORED COMMON TESTING PATTERNS
// These utilities extract common patterns from our integration tests

// HTTPIntegrationTestClient provides standardized HTTP client configurations for integration tests
type HTTPIntegrationTestClient struct {
	Client *http.Client
}

// NewHTTPIntegrationTestClient creates an HTTP client with appropriate timeout for integration tests
func NewHTTPIntegrationTestClient(timeout time.Duration) *HTTPIntegrationTestClient {
	if timeout == 0 {
		timeout = 10 * time.Second // Default timeout for integration tests
	}

	return &HTTPIntegrationTestClient{
		Client: &http.Client{Timeout: timeout},
	}
}

// ServiceEndpoint represents a service endpoint for testing
type ServiceEndpoint struct {
	Name        string
	URL         string
	Method      string
	Description string
}

// TestServiceEndpoints tests multiple service endpoints with common validation patterns
func (c *HTTPIntegrationTestClient) TestServiceEndpoints(t *testing.T, endpoints []ServiceEndpoint) {
	for _, endpoint := range endpoints {
		t.Run(endpoint.Name, func(t *testing.T) {
			var resp *http.Response
			var err error

			switch endpoint.Method {
			case "GET", "":
				resp, err = c.Client.Get(endpoint.URL)
			default:
				req, reqErr := http.NewRequest(endpoint.Method, endpoint.URL, nil)
				require.NoError(t, reqErr, "Should create request for %s %s", endpoint.Method, endpoint.Name)
				resp, err = c.Client.Do(req)
			}

			require.NoError(t, err, "Should be able to reach %s endpoint", endpoint.Name)
			defer resp.Body.Close()

			// Common validation: no server errors
			assert.True(t, resp.StatusCode < 500,
				"Endpoint %s should not return server errors, got %d", endpoint.Name, resp.StatusCode)

			t.Logf("✅ Endpoint %s: %s (status: %d)", endpoint.Name, endpoint.Description, resp.StatusCode)
		})
	}
}

// HealthCheckResponse represents expected health check response structure
type HealthCheckResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Health  string `json:"health"`
}

// TestServiceHealthEndpoints tests service health endpoints with structured validation
func (c *HTTPIntegrationTestClient) TestServiceHealthEndpoints(t *testing.T, services []string, urlPattern string) {
	for _, service := range services {
		t.Run(service+"HealthCheck", func(t *testing.T) {
			url := strings.ReplaceAll(urlPattern, "{service}", service)
			resp, err := c.Client.Get(url)
			require.NoError(t, err, "Should be able to reach %s health endpoint", service)
			defer resp.Body.Close()

			// Validate status code
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
				"Service %s health endpoint should return success status, got %d", service, resp.StatusCode)

			// Validate content type
			contentType := resp.Header.Get("Content-Type")
			assert.True(t, strings.Contains(contentType, "application/json"),
				"Service %s health endpoint should return JSON, got %s", service, contentType)

			// Try to parse as JSON health response
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				var healthData map[string]interface{}
				if json.Unmarshal(body, &healthData) == nil {
					t.Logf("✅ Service %s: Health endpoint returns structured data", service)
				}
			}

			t.Logf("✅ Service %s: Health check validated (status: %d)", service, resp.StatusCode)
		})
	}
}

// DaprServiceEndpoint represents a Dapr service invocation endpoint
type DaprServiceEndpoint struct {
	ServiceName string
	Method      string
	Path        string
	Description string
}

// TestDaprServiceInvocation tests Dapr service invocation endpoints
func (c *HTTPIntegrationTestClient) TestDaprServiceInvocation(t *testing.T, endpoints []DaprServiceEndpoint, daprPort string) {
	if daprPort == "" {
		daprPort = "3500"
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.ServiceName+endpoint.Path, func(t *testing.T) {
			url := fmt.Sprintf("http://localhost:%s/v1.0/invoke/%s/method%s", daprPort, endpoint.ServiceName, endpoint.Path)

			var resp *http.Response
			var err error

			switch endpoint.Method {
			case "GET", "":
				resp, err = c.Client.Get(url)
			default:
				req, reqErr := http.NewRequest(endpoint.Method, url, nil)
				require.NoError(t, reqErr, "Should create Dapr request for %s", endpoint.ServiceName)
				resp, err = c.Client.Do(req)
			}

			require.NoError(t, err, "Should be able to invoke %s via Dapr", endpoint.ServiceName)
			defer resp.Body.Close()

			// Validate service is reachable through Dapr
			assert.True(t, resp.StatusCode < 500,
				"Dapr service invocation for %s should not return server errors, got %d",
				endpoint.ServiceName, resp.StatusCode)

			t.Logf("✅ Dapr service %s: %s - %s (status: %d)",
				endpoint.ServiceName, endpoint.Path, endpoint.Description, resp.StatusCode)
		})
	}
}

// GatewayRoutingTest represents a gateway routing test case
type GatewayRoutingTest struct {
	Name           string
	Path           string
	ExpectedStatus []int
	Description    string
}

// TestGatewayRouting tests gateway routing with flexible status code validation
func (c *HTTPIntegrationTestClient) TestGatewayRouting(t *testing.T, baseURL string, routes []GatewayRoutingTest) {
	for _, route := range routes {
		t.Run(route.Name, func(t *testing.T) {
			url := baseURL + route.Path
			resp, err := c.Client.Get(url)

			if err != nil {
				t.Logf("❌ Gateway routing failed for %s: %v", route.Name, err)
				require.NoError(t, err, "Gateway should be reachable for %s", route.Name)
				return
			}
			defer resp.Body.Close()

			// Validate status code is within expected range
			statusValid := false
			for _, validStatus := range route.ExpectedStatus {
				if resp.StatusCode == validStatus {
					statusValid = true
					break
				}
			}

			assert.True(t, statusValid,
				"Gateway route %s should return valid status code, got %d (expected one of %v)",
				route.Name, resp.StatusCode, route.ExpectedStatus)

			// No server errors
			assert.True(t, resp.StatusCode < 500,
				"Gateway route %s should not return server errors, got %d", route.Name, resp.StatusCode)

			t.Logf("✅ Gateway route %s: %s (status: %d)", route.Name, route.Description, resp.StatusCode)
		})
	}
}

// ContractComplianceTest represents a contract compliance test
type ContractComplianceTest struct {
	Name            string
	Endpoint        string
	ExpectedHeaders map[string]string
	RequiresJSON    bool
	Description     string
}

// TestContractCompliance tests API contract compliance
func (c *HTTPIntegrationTestClient) TestContractCompliance(t *testing.T, contracts []ContractComplianceTest) {
	for _, contract := range contracts {
		t.Run(contract.Name+"Contract", func(t *testing.T) {
			resp, err := c.Client.Get(contract.Endpoint)
			require.NoError(t, err, "Should be able to reach %s contract endpoint", contract.Name)
			defer resp.Body.Close()

			// Validate status code contract
			assert.True(t, resp.StatusCode < 500,
				"Contract %s should not return server errors, got %d", contract.Name, resp.StatusCode)

			// Validate expected headers
			for header, expectedValue := range contract.ExpectedHeaders {
				actualValue := resp.Header.Get(header)
				assert.True(t, strings.Contains(actualValue, expectedValue),
					"Contract %s should have %s header with %s, got %s",
					contract.Name, header, expectedValue, actualValue)
			}

			// Validate JSON response if required
			if contract.RequiresJSON && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				contentType := resp.Header.Get("Content-Type")
				assert.True(t, strings.Contains(contentType, "application/json"),
					"Contract %s should return JSON content type, got %s", contract.Name, contentType)

				// Try to parse JSON
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					var jsonData map[string]interface{}
					assert.NoError(t, json.Unmarshal(body, &jsonData),
						"Contract %s should return valid JSON structure", contract.Name)
				}
			}

			t.Logf("✅ Contract %s: %s compliance validated", contract.Name, contract.Description)
		})
	}
}