package platform

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HealthCheckResult represents the result of a health check operation
type HealthCheckResult struct {
	Healthy   bool
	Status    string
	Message   string
	Timestamp time.Time
	Endpoint  string
}

// ContainerHealthChecker defines the interface for container-specific health operations
type ContainerHealthChecker interface {
	// CheckContainerStatus checks if the container is in a running state
	CheckContainerStatus(ctx context.Context, containerName string) (string, error)
	
	// GetContainerEndpoint returns the health endpoint for a container
	GetContainerEndpoint(containerName string) string
	
	// GetContainerLogs retrieves logs from the container
	GetContainerLogs(ctx context.Context, containerName string, lines int) (string, error)
}

// UnifiedHealthChecker provides common health checking functionality for all container providers
type UnifiedHealthChecker struct {
	HTTPClient     *http.Client
	DefaultTimeout time.Duration
	RetryInterval  time.Duration
	MaxRetries     int
}

// NewUnifiedHealthChecker creates a new health checker with default configuration
func NewUnifiedHealthChecker() *UnifiedHealthChecker {
	return &UnifiedHealthChecker{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		DefaultTimeout: 120 * time.Second,
		RetryInterval:  5 * time.Second,
		MaxRetries:     3,
	}
}

// WaitForContainerHealth waits for a container to become healthy using the provided checker
func (h *UnifiedHealthChecker) WaitForContainerHealth(ctx context.Context, containerName string, checker ContainerHealthChecker, timeout time.Duration) error {
	if timeout == 0 {
		timeout = h.DefaultTimeout
	}
	
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(h.RetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for container %s to become healthy", containerName)
		case <-ticker.C:
			result, err := h.CheckContainerHealth(ctx, containerName, checker)
			if err != nil {
				continue // Continue retrying on error
			}
			
			if result.Healthy {
				return nil
			}
			
			if result.Status == "failed" || result.Status == "unhealthy" {
				return fmt.Errorf("container %s is in failed state: %s", containerName, result.Message)
			}
		}
	}
}

// CheckContainerHealth performs a comprehensive health check on a container
func (h *UnifiedHealthChecker) CheckContainerHealth(ctx context.Context, containerName string, checker ContainerHealthChecker) (*HealthCheckResult, error) {
	result := &HealthCheckResult{
		Timestamp: time.Now(),
		Healthy:   false,
	}

	// Check container status first
	status, err := checker.CheckContainerStatus(ctx, containerName)
	if err != nil {
		return result, fmt.Errorf("failed to check container status: %w", err)
	}
	
	result.Status = status
	
	// Basic status validation
	switch status {
	case "running", "Succeeded":
		// Container is running, proceed to health endpoint check
		break
	case "failed", "Failed", "unhealthy":
		result.Message = fmt.Sprintf("Container is in failed state: %s", status)
		return result, nil
	default:
		result.Message = fmt.Sprintf("Container is not ready: %s", status)
		return result, nil
	}

	// Check HTTP health endpoint if available
	endpoint := checker.GetContainerEndpoint(containerName)
	if endpoint != "" {
		result.Endpoint = endpoint
		healthy, message, err := h.CheckHTTPHealthEndpoint(ctx, endpoint)
		if err != nil {
			result.Message = fmt.Sprintf("Health endpoint check failed: %v", err)
			return result, nil
		}
		
		result.Healthy = healthy
		result.Message = message
		return result, nil
	}

	// If no health endpoint, consider running container as healthy
	result.Healthy = true
	result.Message = "Container is running (no health endpoint configured)"
	return result, nil
}

// CheckHTTPHealthEndpoint validates an HTTP health endpoint
func (h *UnifiedHealthChecker) CheckHTTPHealthEndpoint(ctx context.Context, endpoint string) (bool, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return false, "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return false, "Health endpoint not accessible", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, "Health endpoint returned success", nil
	}

	return false, fmt.Sprintf("Health endpoint returned status: %d", resp.StatusCode), nil
}

// ValidateDaprHealth performs Dapr-specific health validation
func (h *UnifiedHealthChecker) ValidateDaprHealth(ctx context.Context, appID string, daprEndpoint string) error {
	if appID == "" {
		return fmt.Errorf("Dapr app ID cannot be empty")
	}

	if daprEndpoint == "" {
		return fmt.Errorf("Dapr endpoint not configured for app %s", appID)
	}

	// Check Dapr health endpoint
	healthURL := fmt.Sprintf("%s/v1.0/healthz", daprEndpoint)
	healthy, message, err := h.CheckHTTPHealthEndpoint(ctx, healthURL)
	if err != nil {
		return fmt.Errorf("Dapr health check failed for %s: %w", appID, err)
	}

	if !healthy {
		return fmt.Errorf("Dapr sidecar for %s is unhealthy: %s", appID, message)
	}

	return nil
}

// CheckMultipleContainers checks health of multiple containers concurrently
func (h *UnifiedHealthChecker) CheckMultipleContainers(ctx context.Context, containers []string, checker ContainerHealthChecker) map[string]*HealthCheckResult {
	results := make(map[string]*HealthCheckResult)
	resultChan := make(chan struct {
		name   string
		result *HealthCheckResult
		err    error
	}, len(containers))

	// Start health checks concurrently
	for _, containerName := range containers {
		go func(name string) {
			result, err := h.CheckContainerHealth(ctx, name, checker)
			resultChan <- struct {
				name   string
				result *HealthCheckResult
				err    error
			}{name, result, err}
		}(containerName)
	}

	// Collect results
	for i := 0; i < len(containers); i++ {
		res := <-resultChan
		if res.err != nil {
			results[res.name] = &HealthCheckResult{
				Healthy:   false,
				Status:    "error",
				Message:   res.err.Error(),
				Timestamp: time.Now(),
			}
		} else {
			results[res.name] = res.result
		}
	}

	return results
}

// WaitForMultipleContainers waits for multiple containers to become healthy
func (h *UnifiedHealthChecker) WaitForMultipleContainers(ctx context.Context, containers []string, checker ContainerHealthChecker, timeout time.Duration) error {
	if timeout == 0 {
		timeout = h.DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(h.RetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for containers to become healthy")
		case <-ticker.C:
			results := h.CheckMultipleContainers(ctx, containers, checker)
			
			allHealthy := true
			var unhealthyContainers []string
			
			for containerName, result := range results {
				if !result.Healthy {
					allHealthy = false
					unhealthyContainers = append(unhealthyContainers, fmt.Sprintf("%s (%s)", containerName, result.Message))
				}
			}
			
			if allHealthy {
				return nil
			}
			
			// Continue waiting if containers are still starting up
		}
	}
}

// GetHealthSummary returns a summary of health check results
func (h *UnifiedHealthChecker) GetHealthSummary(results map[string]*HealthCheckResult) (int, int, []string) {
	healthy := 0
	unhealthy := 0
	var issues []string

	for containerName, result := range results {
		if result.Healthy {
			healthy++
		} else {
			unhealthy++
			issues = append(issues, fmt.Sprintf("%s: %s", containerName, result.Message))
		}
	}

	return healthy, unhealthy, issues
}