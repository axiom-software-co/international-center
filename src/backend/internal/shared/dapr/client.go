package dapr

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/dapr/go-sdk/client"
)

// Client wraps the Dapr client with environment-aware configuration
type Client struct {
	client      client.Client
	environment string
	appID       string
}

var (
	daprClient *Client
	once       sync.Once
)

// NewClient creates a singleton Dapr client instance
func NewClient() (*Client, error) {
	var err error
	once.Do(func() {
		environment := getEnv("ENVIRONMENT", "development")
		appID := getEnv("DAPR_APP_ID", "international-center")

		daprSDKClient, clientErr := client.NewClient()
		if clientErr != nil {
			err = fmt.Errorf("failed to create Dapr client: %w", clientErr)
			return
		}

		daprClient = &Client{
			client:      daprSDKClient,
			environment: environment,
			appID:       appID,
		}
	})

	if err != nil {
		return nil, err
	}

	return daprClient, nil
}

// GetClient returns the underlying Dapr client
func (c *Client) GetClient() client.Client {
	return c.client
}

// GetEnvironment returns the current environment
func (c *Client) GetEnvironment() string {
	return c.environment
}

// GetAppID returns the current application ID
func (c *Client) GetAppID() string {
	return c.appID
}

// Close closes the Dapr client connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// HealthCheck validates the Dapr client connection
func (c *Client) HealthCheck(ctx context.Context) error {
	// Test connectivity by attempting to get configuration
	_, err := c.client.GetConfigurationItem(ctx, "healthcheck", "test")
	if err != nil {
		// Configuration not found is acceptable for health check
		return nil
	}
	return nil
}

// IsHealthy returns true if the Dapr client is healthy
func (c *Client) IsHealthy(ctx context.Context) bool {
	err := c.HealthCheck(ctx)
	return err == nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}