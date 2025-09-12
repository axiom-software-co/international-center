package testing

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SharedEnvironmentValidator provides centralized environment validation for all modules
type SharedEnvironmentValidator struct {
	httpClient *http.Client
	timeout    time.Duration
}

// NewSharedEnvironmentValidator creates a new shared environment validator
func NewSharedEnvironmentValidator() *SharedEnvironmentValidator {
	return &SharedEnvironmentValidator{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		timeout:    15 * time.Second,
	}
}

// ValidateEnvironmentPrerequisites validates environment health for integration testing
// This replaces the 16+ duplicated validateEnvironmentPrerequisites() functions
func (validator *SharedEnvironmentValidator) ValidateEnvironmentPrerequisites(ctx context.Context, requiredServices []string) error {
	// Define all available services and their health endpoints
	serviceEndpoints := map[string]string{
		"postgresql":        "http://localhost:5432", // Note: PostgreSQL doesn't have HTTP endpoint, this is for reference
		"dapr-control-plane": "http://localhost:3500/v1.0/healthz",
		"content":           "http://localhost:3001/health",
		"inquiries":         "http://localhost:3101/health", 
		"notifications":     "http://localhost:3201/health",
		"public-gateway":    "http://localhost:9001/health",
		"admin-gateway":     "http://localhost:9000/health",
		"rabbitmq":          "http://localhost:15672", // RabbitMQ management interface
		"vault":             "http://localhost:8200/v1/sys/health",
		"azurite":           "http://localhost:10000", // Azurite blob service
	}
	
	// Validate each required service
	for _, serviceName := range requiredServices {
		endpoint, exists := serviceEndpoints[serviceName]
		if !exists {
			return fmt.Errorf("unknown service: %s", serviceName)
		}
		
		// Skip services that don't have HTTP health endpoints
		if serviceName == "postgresql" {
			// PostgreSQL validation would be done differently
			continue
		}
		
		if err := validator.validateServiceHealth(ctx, serviceName, endpoint); err != nil {
			return fmt.Errorf("service %s not healthy: %w", serviceName, err)
		}
	}
	
	return nil
}

// validateServiceHealth validates a single service health
func (validator *SharedEnvironmentValidator) validateServiceHealth(ctx context.Context, serviceName, endpoint string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}
	
	resp, err := validator.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Accept various health check response codes
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return nil // Healthy
	}
	
	return fmt.Errorf("unhealthy status: %d", resp.StatusCode)
}

// ValidateBasicEnvironment validates the basic development environment
func (validator *SharedEnvironmentValidator) ValidateBasicEnvironment(ctx context.Context) error {
	basicServices := []string{
		"dapr-control-plane",
		"content", 
		"inquiries",
		"notifications",
		"public-gateway",
		"admin-gateway",
	}
	
	return validator.ValidateEnvironmentPrerequisites(ctx, basicServices)
}

// ValidateFullEnvironment validates the complete development environment
func (validator *SharedEnvironmentValidator) ValidateFullEnvironment(ctx context.Context) error {
	fullServices := []string{
		"dapr-control-plane",
		"content",
		"inquiries", 
		"notifications",
		"public-gateway",
		"admin-gateway",
		"vault",
		"azurite",
	}
	
	return validator.ValidateEnvironmentPrerequisites(ctx, fullServices)
}

// SharedHTTPTestClient provides centralized HTTP testing utilities
type SharedHTTPTestClient struct {
	client  *http.Client
	baseURL string
	headers map[string]string
}

// NewSharedHTTPTestClient creates a new shared HTTP test client
func NewSharedHTTPTestClient(baseURL string) *SharedHTTPTestClient {
	return &SharedHTTPTestClient{
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: baseURL,
		headers: make(map[string]string),
	}
}

// SetHeader sets a default header for all requests
func (client *SharedHTTPTestClient) SetHeader(key, value string) {
	client.headers[key] = value
}

// Get performs a GET request with shared client configuration
func (client *SharedHTTPTestClient) Get(ctx context.Context, path string) (*http.Response, error) {
	return client.request(ctx, "GET", path, nil)
}

// Post performs a POST request with shared client configuration  
func (client *SharedHTTPTestClient) Post(ctx context.Context, path string, body []byte) (*http.Response, error) {
	return client.request(ctx, "POST", path, body)
}

// request performs an HTTP request with shared configuration
func (client *SharedHTTPTestClient) request(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	url := client.baseURL + path
	
	var bodyReader io.Reader
	if body != nil {
		bodyReader = strings.NewReader(string(body))
	}
	
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Apply default headers
	for key, value := range client.headers {
		req.Header.Set(key, value)
	}
	
	// Set content type for POST requests
	if method == "POST" && body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	return client.client.Do(req)
}

// SharedTestingUtilities provides common utilities for all integration tests
type SharedTestingUtilities struct {
	EnvironmentValidator *SharedEnvironmentValidator
	HTTPClient           *SharedHTTPTestClient
}

// NewSharedTestingUtilities creates comprehensive shared testing utilities
func NewSharedTestingUtilities(baseURL string) *SharedTestingUtilities {
	return &SharedTestingUtilities{
		EnvironmentValidator: NewSharedEnvironmentValidator(),
		HTTPClient:           NewSharedHTTPTestClient(baseURL),
	}
}

// SetupIntegrationTestEnvironment sets up the environment for integration testing
func (utils *SharedTestingUtilities) SetupIntegrationTestEnvironment(ctx context.Context, requiredServices []string) error {
	// Validate environment prerequisites using shared validator
	if err := utils.EnvironmentValidator.ValidateEnvironmentPrerequisites(ctx, requiredServices); err != nil {
		return fmt.Errorf("integration test environment not ready: %w", err)
	}
	
	// Set up common HTTP client configuration
	utils.HTTPClient.SetHeader("User-Agent", "Integration-Test-Client/1.0")
	utils.HTTPClient.SetHeader("Accept", "application/json")
	
	return nil
}