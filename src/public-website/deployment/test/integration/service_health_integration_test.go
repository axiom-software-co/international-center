package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ServiceHealthResponse represents the expected health check response format
type ServiceHealthResponse struct {
	Status      string            `json:"status"`
	Timestamp   string            `json:"timestamp"`
	Service     string            `json:"service"`
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	Dependencies map[string]string `json:"dependencies"`
	Uptime      string            `json:"uptime"`
}

func TestServiceHealth_DaprControlPlane_Integration(t *testing.T) {
	// Skip integration tests unless explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests skipped - set RUN_INTEGRATION_TESTS=true to enable")
	}

	// Test Dapr control plane health endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 15 * time.Second}
	
	// Test Dapr control plane health
	healthURL := "http://localhost:50001/v1.0/healthz"
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err, "Dapr control plane health endpoint should be accessible")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Dapr control plane should return healthy status")

	// Test Dapr placement service
	placementURL := "http://localhost:50005/healthz"
	req, err = http.NewRequestWithContext(ctx, "GET", placementURL, nil)
	if err == nil { // Placement service might not have HTTP health endpoint
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound, 
				"Placement service should be accessible or not expose HTTP health endpoint")
		}
	}
}

func TestServiceHealth_Gateways_Integration(t *testing.T) {
	// Skip integration tests unless explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests skipped - set RUN_INTEGRATION_TESTS=true to enable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 15 * time.Second}

	// Test public gateway health
	publicHealthURL := "http://127.0.0.1:9001/health"
	req, err := http.NewRequestWithContext(ctx, "GET", publicHealthURL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err, "Public gateway health endpoint should be accessible")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Public gateway should return healthy status")

	// Validate public gateway health response format
	var publicHealth ServiceHealthResponse
	err = json.NewDecoder(resp.Body).Decode(&publicHealth)
	require.NoError(t, err, "Public gateway should return valid health JSON")

	assert.Equal(t, "healthy", publicHealth.Status)
	assert.Equal(t, "public-gateway", publicHealth.Service)
	assert.Equal(t, "development", publicHealth.Environment)
	assert.NotEmpty(t, publicHealth.Timestamp)
	assert.NotEmpty(t, publicHealth.Version)

	// Validate dependencies are healthy
	assert.Contains(t, publicHealth.Dependencies, "database")
	assert.Contains(t, publicHealth.Dependencies, "dapr")
	assert.Equal(t, "healthy", publicHealth.Dependencies["database"])
	assert.Equal(t, "healthy", publicHealth.Dependencies["dapr"])

	// Test admin gateway health  
	adminHealthURL := "http://127.0.0.1:9000/health"
	req, err = http.NewRequestWithContext(ctx, "GET", adminHealthURL, nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err, "Admin gateway health endpoint should be accessible")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Admin gateway should return healthy status")

	// Validate admin gateway health response format
	var adminHealth ServiceHealthResponse
	err = json.NewDecoder(resp.Body).Decode(&adminHealth)
	require.NoError(t, err, "Admin gateway should return valid health JSON")

	assert.Equal(t, "healthy", adminHealth.Status)
	assert.Equal(t, "admin-gateway", adminHealth.Service)
	assert.Equal(t, "development", adminHealth.Environment)
	
	// Admin gateway should have additional dependencies
	assert.Contains(t, adminHealth.Dependencies, "database")
	assert.Contains(t, adminHealth.Dependencies, "dapr")
	assert.Contains(t, adminHealth.Dependencies, "vault")
	assert.Equal(t, "healthy", adminHealth.Dependencies["vault"])
}

func TestServiceHealth_ContentServices_Integration(t *testing.T) {
	// Skip integration tests unless explicitly enabled  
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests skipped - set RUN_INTEGRATION_TESTS=true to enable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 15 * time.Second}

	// Test content services health endpoints
	contentServices := []struct {
		name string
		port int
	}{
		{"news", 3001},
		{"events", 3002},
		{"research", 3003},
	}

	for _, service := range contentServices {
		t.Run(fmt.Sprintf("content-%s", service.name), func(t *testing.T) {
			healthURL := fmt.Sprintf("http://localhost:%d/health", service.port)
			req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err, fmt.Sprintf("Content %s service health endpoint should be accessible", service.name))
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Content %s service should return healthy status", service.name))

			// Validate health response format
			var health ServiceHealthResponse
			err = json.NewDecoder(resp.Body).Decode(&health)
			require.NoError(t, err, fmt.Sprintf("Content %s service should return valid health JSON", service.name))

			assert.Equal(t, "healthy", health.Status)
			assert.Equal(t, fmt.Sprintf("content-%s", service.name), health.Service)
			assert.Equal(t, "development", health.Environment)

			// Validate content service dependencies
			assert.Contains(t, health.Dependencies, "database")
			assert.Contains(t, health.Dependencies, "dapr")
			assert.Equal(t, "healthy", health.Dependencies["database"])
			assert.Equal(t, "healthy", health.Dependencies["dapr"])
		})
	}
}

func TestServiceHealth_InquiriesServices_Integration(t *testing.T) {
	// Skip integration tests unless explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests skipped - set RUN_INTEGRATION_TESTS=true to enable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 15 * time.Second}

	// Test inquiries services health endpoints  
	inquiriesServices := []struct {
		name string
		port int
	}{
		{"business", 3011},
		{"donations", 3012},
		{"media", 3013},
		{"volunteers", 3014},
	}

	for _, service := range inquiriesServices {
		t.Run(fmt.Sprintf("inquiries-%s", service.name), func(t *testing.T) {
			healthURL := fmt.Sprintf("http://localhost:%d/health", service.port)
			req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err, fmt.Sprintf("Inquiries %s service health endpoint should be accessible", service.name))
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Inquiries %s service should return healthy status", service.name))

			// Validate health response format
			var health ServiceHealthResponse
			err = json.NewDecoder(resp.Body).Decode(&health)
			require.NoError(t, err, fmt.Sprintf("Inquiries %s service should return valid health JSON", service.name))

			assert.Equal(t, "healthy", health.Status)
			assert.Equal(t, fmt.Sprintf("inquiries-%s", service.name), health.Service)

			// Validate inquiries service dependencies
			assert.Contains(t, health.Dependencies, "database")
			assert.Contains(t, health.Dependencies, "dapr")
			assert.Contains(t, health.Dependencies, "messaging")
			assert.Equal(t, "healthy", health.Dependencies["messaging"])
		})
	}
}

func TestServiceHealth_NotificationService_Integration(t *testing.T) {
	// Skip integration tests unless explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests skipped - set RUN_INTEGRATION_TESTS=true to enable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 15 * time.Second}

	// Test notification service health
	healthURL := "http://localhost:3020/health"
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err, "Notification service health endpoint should be accessible")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Notification service should return healthy status")

	// Validate notification service health response
	var health ServiceHealthResponse
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err, "Notification service should return valid health JSON")

	assert.Equal(t, "healthy", health.Status)
	assert.Equal(t, "notification-service", health.Service)

	// Validate notification service dependencies
	assert.Contains(t, health.Dependencies, "database")
	assert.Contains(t, health.Dependencies, "dapr")
	assert.Contains(t, health.Dependencies, "messaging")
	assert.Contains(t, health.Dependencies, "smtp") // Email service dependency
}

func TestServiceHealth_AdminPortal_Integration(t *testing.T) {
	// Skip integration tests unless explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests skipped - set RUN_INTEGRATION_TESTS=true to enable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 15 * time.Second}

	// Test admin portal (Directus) health
	healthURL := "http://localhost:8055/server/health"
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err, "Admin portal health endpoint should be accessible")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Admin portal should return healthy status")

	// Test admin portal main interface accessibility
	portalURL := "http://localhost:8055"
	req, err = http.NewRequestWithContext(ctx, "GET", portalURL, nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err, "Admin portal main interface should be accessible")
	defer resp.Body.Close()

	// Admin portal should redirect to login or return main interface
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound,
		"Admin portal should be accessible")
}

func TestServiceHealth_DependencyValidation_Integration(t *testing.T) {
	// Skip integration tests unless explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests skipped - set RUN_INTEGRATION_TESTS=true to enable")
	}

	// This test validates that all services properly check their dependencies
	// and report dependency health status correctly

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 15 * time.Second}

	// Test that services report unhealthy when dependencies are down
	// This is a comprehensive dependency validation test

	t.Run("database_dependency_validation", func(t *testing.T) {
		// TODO: This test would temporarily stop database container
		// and verify services report unhealthy status
		t.Skip("Database dependency validation test requires container manipulation")
	})

	t.Run("dapr_dependency_validation", func(t *testing.T) {
		// TODO: This test would temporarily stop Dapr control plane
		// and verify services report unhealthy status
		t.Skip("Dapr dependency validation test requires container manipulation")
	})

	t.Run("messaging_dependency_validation", func(t *testing.T) {
		// TODO: This test would temporarily stop message queue
		// and verify services report unhealthy status
		t.Skip("Messaging dependency validation test requires container manipulation")
	})
}

func TestServiceHealth_LoadBalancing_Integration(t *testing.T) {
	// Skip integration tests unless explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests skipped - set RUN_INTEGRATION_TESTS=true to enable")
	}

	// This test validates that load balancing works correctly for services
	// with multiple replicas (mainly relevant for staging/production)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 5 * time.Second}

	// In development, we typically have single replicas, but we can still
	// test that the service is accessible consistently

	for i := 0; i < 10; i++ {
		// Test public gateway accessibility
		req, err := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:9001/health", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err, fmt.Sprintf("Request %d: Public gateway should be accessible", i+1))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Request %d: Public gateway should be healthy", i+1))
	}
}

func TestServiceHealth_ResponseTimeValidation_Integration(t *testing.T) {
	// Skip integration tests unless explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests skipped - set RUN_INTEGRATION_TESTS=true to enable")
	}

	// This test validates that health endpoints respond within acceptable time limits

	services := []struct {
		name string
		url  string
		maxResponseTime time.Duration
	}{
		{"dapr-control-plane", "http://localhost:50001/v1.0/healthz", 2 * time.Second},
		{"public-gateway", "http://127.0.0.1:9001/health", 3 * time.Second},
		{"admin-gateway", "http://127.0.0.1:9000/health", 3 * time.Second},
		{"content-news", "http://localhost:3001/health", 5 * time.Second},
		{"content-events", "http://localhost:3002/health", 5 * time.Second},
		{"content-research", "http://localhost:3003/health", 5 * time.Second},
		{"admin-portal", "http://localhost:8055/server/health", 5 * time.Second},
	}

	for _, service := range services {
		t.Run(fmt.Sprintf("%s_response_time", service.name), func(t *testing.T) {
			client := &http.Client{Timeout: service.maxResponseTime}
			
			start := time.Now()
			resp, err := client.Get(service.url)
			duration := time.Since(start)
			
			require.NoError(t, err, fmt.Sprintf("%s should respond within timeout", service.name))
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("%s should return healthy status", service.name))
			assert.Less(t, duration, service.maxResponseTime, 
				fmt.Sprintf("%s should respond within %v, took %v", service.name, service.maxResponseTime, duration))
		})
	}
}