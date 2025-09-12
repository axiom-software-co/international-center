package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DaprServiceTestClient provides utilities for testing services through Dapr service mesh
type DaprServiceTestClient struct {
	serviceAppID string
	daprPort     string
	httpClient   *http.Client
}

// NewDaprServiceTestClient creates a new Dapr service test client
func NewDaprServiceTestClient(serviceAppID, daprPort string) *DaprServiceTestClient {
	return &DaprServiceTestClient{
		serviceAppID: serviceAppID,
		daprPort:     daprPort,
		httpClient:   &http.Client{Timeout: 5 * time.Second},
	}
}

// InvokeService performs Dapr service invocation for testing
func (client *DaprServiceTestClient) InvokeService(ctx context.Context, targetAppID, method, endpoint string, data []byte) (*http.Response, error) {
	// Use Dapr service invocation instead of direct HTTP
	invokeURL := fmt.Sprintf("http://localhost:%s/v1.0/invoke/%s/method%s", client.daprPort, targetAppID, endpoint)
	
	req, err := http.NewRequestWithContext(ctx, method, invokeURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create Dapr service invocation request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("dapr-app-id", client.serviceAppID)
	
	return client.httpClient.Do(req)
}

// PublishEvent publishes an event through Dapr pub/sub for testing
func (client *DaprServiceTestClient) PublishEvent(ctx context.Context, topic string, data interface{}) (*http.Response, error) {
	publishURL := fmt.Sprintf("http://localhost:%s/v1.0/publish/pubsub/%s", client.daprPort, topic)
	
	eventData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", publishURL, bytes.NewReader(eventData))
	if err != nil {
		return nil, fmt.Errorf("failed to create Dapr pub/sub request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("dapr-app-id", client.serviceAppID)
	
	return client.httpClient.Do(req)
}

// GetMetadata retrieves Dapr metadata for the service
func (client *DaprServiceTestClient) GetMetadata(ctx context.Context) (map[string]interface{}, error) {
	metadataURL := fmt.Sprintf("http://localhost:%s/v1.0/metadata", client.daprPort)
	
	req, err := http.NewRequestWithContext(ctx, "GET", metadataURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata request: %w", err)
	}
	
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("metadata request failed: status %d", resp.StatusCode)
	}
	
	var metadata map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}
	
	return metadata, nil
}

// GetComponents retrieves Dapr components configuration for the service
func (client *DaprServiceTestClient) GetComponents(ctx context.Context) ([]map[string]interface{}, error) {
	componentsURL := fmt.Sprintf("http://localhost:%s/v1.0/components", client.daprPort)
	
	req, err := http.NewRequestWithContext(ctx, "GET", componentsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create components request: %w", err)
	}
	
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get components: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("components request failed: status %d", resp.StatusCode)
	}
	
	var components []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&components); err != nil {
		return nil, fmt.Errorf("failed to decode components: %w", err)
	}
	
	return components, nil
}

// ValidateStateStoreAccess tests if the service can access the state store component
func (client *DaprServiceTestClient) ValidateStateStoreAccess(ctx context.Context, stateStoreName string) error {
	stateURL := fmt.Sprintf("http://localhost:%s/v1.0/state/%s/test-validation-key", client.daprPort, stateStoreName)
	
	req, err := http.NewRequestWithContext(ctx, "GET", stateURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create state store request: %w", err)
	}
	
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("state store access failed: %w", err)
	}
	defer resp.Body.Close()
	
	// State store should be accessible (200, 204, or 404 acceptable - not 500)
	if resp.StatusCode >= 500 {
		return fmt.Errorf("state store error: status %d", resp.StatusCode)
	}
	
	return nil
}

// ValidatePubSubAccess tests if the service can access the pub/sub component
func (client *DaprServiceTestClient) ValidatePubSubAccess(ctx context.Context, pubsubName, topic string) error {
	publishURL := fmt.Sprintf("http://localhost:%s/v1.0/publish/%s/%s", client.daprPort, pubsubName, topic)
	
	testEvent := map[string]interface{}{
		"test":      "validation",
		"timestamp": time.Now().Unix(),
	}
	
	eventData, err := json.Marshal(testEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal test event: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", publishURL, bytes.NewReader(eventData))
	if err != nil {
		return fmt.Errorf("failed to create pub/sub request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("pub/sub access failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Pub/sub should accept events (200, 204 acceptable)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("pub/sub error: status %d", resp.StatusCode)
	}
	
	return nil
}

// DaprServiceMeshTestRunner provides comprehensive service mesh testing
type DaprServiceMeshTestRunner struct {
	services map[string]*DaprServiceTestClient
}

// NewDaprServiceMeshTestRunner creates a new service mesh test runner
func NewDaprServiceMeshTestRunner() *DaprServiceMeshTestRunner {
	services := map[string]*DaprServiceTestClient{
		"content":       NewDaprServiceTestClient("content", "50030"),
		"inquiries":     NewDaprServiceTestClient("inquiries", "50040"),
		"notifications": NewDaprServiceTestClient("notifications", "50050"),
	}
	
	return &DaprServiceMeshTestRunner{
		services: services,
	}
}

// ValidateServiceMeshCommunication validates communication between all services
func (runner *DaprServiceMeshTestRunner) ValidateServiceMeshCommunication(ctx context.Context) []error {
	var errors []error
	
	// Test communication patterns between services
	communicationTests := []struct {
		fromService string
		toService   string
		endpoint    string
		method      string
	}{
		{"content", "inquiries", "/api/inquiries/content-related", "POST"},
		{"content", "notifications", "/api/notifications/send", "POST"},
		{"inquiries", "content", "/api/content/inquiry-context", "GET"},
		{"inquiries", "notifications", "/api/notifications/inquiry-submitted", "POST"},
		{"notifications", "content", "/api/content/notification-context", "GET"},
	}
	
	for _, test := range communicationTests {
		fromClient := runner.services[test.fromService]
		if fromClient == nil {
			errors = append(errors, fmt.Errorf("service %s not configured for testing", test.fromService))
			continue
		}
		
		// Test service-to-service communication
		resp, err := fromClient.InvokeService(ctx, test.toService, test.method, test.endpoint, nil)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s → %s communication failed: %w", test.fromService, test.toService, err))
			continue
		}
		defer resp.Body.Close()
		
		// Communication should work (any status < 500 acceptable for testing)
		if resp.StatusCode >= 500 {
			errors = append(errors, fmt.Errorf("%s → %s communication error: status %d", test.fromService, test.toService, resp.StatusCode))
		}
	}
	
	return errors
}

// ValidateComponentConfiguration validates all Dapr components are properly configured
func (runner *DaprServiceMeshTestRunner) ValidateComponentConfiguration(ctx context.Context) map[string][]error {
	results := make(map[string][]error)
	
	for serviceName, client := range runner.services {
		var serviceErrors []error
		
		// Validate state store access
		if err := client.ValidateStateStoreAccess(ctx, "statestore"); err != nil {
			serviceErrors = append(serviceErrors, fmt.Errorf("state store validation failed: %w", err))
		}
		
		// Validate pub/sub access
		if err := client.ValidatePubSubAccess(ctx, "pubsub", "test-events"); err != nil {
			serviceErrors = append(serviceErrors, fmt.Errorf("pub/sub validation failed: %w", err))
		}
		
		// Get service metadata for diagnostics
		_, err := client.GetMetadata(ctx)
		if err != nil {
			serviceErrors = append(serviceErrors, fmt.Errorf("metadata validation failed: %w", err))
		} else {
			// Log component configuration status
			if components, err := client.GetComponents(ctx); err == nil {
				if len(components) == 0 {
					serviceErrors = append(serviceErrors, fmt.Errorf("no components configured for service"))
				}
			}
		}
		
		if len(serviceErrors) > 0 {
			results[serviceName] = serviceErrors
		}
	}
	
	return results
}

// DaprTestingResult holds the results of Dapr testing operations
type DaprTestingResult struct {
	ServiceName string
	TestType    string
	Success     bool
	Error       error
	Duration    time.Duration
}

// RunComprehensiveDaprTesting runs comprehensive Dapr service mesh testing
func (runner *DaprServiceMeshTestRunner) RunComprehensiveDaprTesting(ctx context.Context) []DaprTestingResult {
	var results []DaprTestingResult
	
	// Test service mesh communication
	start := time.Now()
	communicationErrors := runner.ValidateServiceMeshCommunication(ctx)
	communicationDuration := time.Since(start)
	
	if len(communicationErrors) == 0 {
		results = append(results, DaprTestingResult{
			ServiceName: "all-services",
			TestType:    "service-mesh-communication",
			Success:     true,
			Duration:    communicationDuration,
		})
	} else {
		for _, err := range communicationErrors {
			results = append(results, DaprTestingResult{
				ServiceName: "service-mesh",
				TestType:    "communication",
				Success:     false,
				Error:       err,
				Duration:    communicationDuration,
			})
		}
	}
	
	// Test component configuration
	start = time.Now()
	componentErrors := runner.ValidateComponentConfiguration(ctx)
	componentDuration := time.Since(start)
	
	for serviceName, errors := range componentErrors {
		for _, err := range errors {
			results = append(results, DaprTestingResult{
				ServiceName: serviceName,
				TestType:    "component-configuration",
				Success:     false,
				Error:       err,
				Duration:    componentDuration,
			})
		}
	}
	
	// If no component errors, mark as success
	if len(componentErrors) == 0 {
		results = append(results, DaprTestingResult{
			ServiceName: "all-services",
			TestType:    "component-configuration",
			Success:     true,
			Duration:    componentDuration,
		})
	}
	
	return results
}