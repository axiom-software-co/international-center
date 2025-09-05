package dapr

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/dapr/go-sdk/client"
)

// simpleMockDaprClient is a minimal mock implementation that only implements methods we actually use
type simpleMockDaprClient struct {
	environment string
	appID       string
}

// Close implements the Close method for mock client
func (m *simpleMockDaprClient) Close() {}

// GetSecret implements a mock GetSecret method
func (m *simpleMockDaprClient) GetSecret(ctx context.Context, storeName, key string, meta map[string]string) (map[string]string, error) {
	// Return mock secret data for testing
	mockSecrets := map[string]map[string]string{
		"database-connection-string": {"database-connection-string": "mock-database-connection"},
		"redis-connection-string":    {"redis-connection-string": "mock-redis-connection"},
		"oauth2-client-secret":       {"oauth2-client-secret": "mock-oauth2-secret-with-sufficient-length"},
		"external-api-key":           {"external-api-key": "mock-api-key"},
		"vault-token":                {"vault-token": "mock-vault-token"},
		"blob-storage-key":           {"blob-storage-key": "mock-blob-key"},
		"grafana-api-key":            {"grafana-api-key": "mock-grafana-key"},
	}
	
	if secret, exists := mockSecrets[key]; exists {
		return secret, nil
	}
	
	return nil, fmt.Errorf("secret %s not found", key)
}

// GetConfigurationItem implements a mock GetConfigurationItem method
func (m *simpleMockDaprClient) GetConfigurationItem(ctx context.Context, storeName, key string) (*client.ConfigurationItem, error) {
	// Return mock configuration for testing
	return &client.ConfigurationItem{
		Value: "mock-config-value",
	}, nil
}

// InvokeBinding implements a mock InvokeBinding method
func (m *simpleMockDaprClient) InvokeBinding(ctx context.Context, in *client.InvokeBindingRequest) (out *client.BindingEvent, err error) {
	// Return mock binding response for testing
	return &client.BindingEvent{
		Data: []byte("mock-binding-response"),
		Metadata: map[string]string{
			"mock": "true",
		},
	}, nil
}

// SaveState implements a mock SaveState method
func (m *simpleMockDaprClient) SaveState(ctx context.Context, storeName, key string, data []byte, meta map[string]string, so ...client.StateOption) error {
	// Mock successful state save
	return nil
}

// GetState implements a mock GetState method
func (m *simpleMockDaprClient) GetState(ctx context.Context, storeName, key string, meta map[string]string) (item *client.StateItem, err error) {
	// Return mock state data
	return &client.StateItem{
		Key:   key,
		Value: []byte("mock-state-data"),
	}, nil
}

// DeleteState implements a mock DeleteState method
func (m *simpleMockDaprClient) DeleteState(ctx context.Context, storeName, key string, meta map[string]string) error {
	// Mock successful state deletion
	return nil
}

// PublishEvent implements a mock PublishEvent method
func (m *simpleMockDaprClient) PublishEvent(ctx context.Context, pubsubName, topicName string, data []byte, meta map[string]string) error {
	// Mock successful event publishing
	return nil
}

// GetBulkState implements a mock GetBulkState method
func (m *simpleMockDaprClient) GetBulkState(ctx context.Context, storeName string, keys []string, meta map[string]string, parallelism int) ([]*client.BulkStateItem, error) {
	// Return mock bulk state data
	var items []*client.BulkStateItem
	for _, key := range keys {
		items = append(items, &client.BulkStateItem{
			Key:   key,
			Value: []byte("mock-bulk-state-data"),
		})
	}
	return items, nil
}

// All other methods return not implemented errors or default values
// This approach avoids having to implement the full Dapr client interface

func (m *simpleMockDaprClient) InvokeMethod(ctx context.Context, appID, methodName, verb string) (out []byte, err error) {
	return []byte("mock-service-response"), nil
}

func (m *simpleMockDaprClient) InvokeMethodWithContent(ctx context.Context, appID, methodName, verb string, content *client.DataContent) (out []byte, err error) {
	return []byte("mock-service-response-with-content"), nil
}

func (m *simpleMockDaprClient) SaveBulkState(ctx context.Context, storeName string, items ...*client.SetStateItem) error {
	return nil
}

func (m *simpleMockDaprClient) DeleteBulkState(ctx context.Context, storeName string, keys []string, meta map[string]string) error {
	return nil
}

func (m *simpleMockDaprClient) Decrypt(ctx context.Context, data io.Reader, options client.DecryptOptions) (io.Reader, error) {
	return strings.NewReader("mock-decrypted-data"), nil
}

func (m *simpleMockDaprClient) DeleteBulkStateItems(ctx context.Context, storeName string, items []*client.DeleteStateItem) error {
	return nil
}

func (m *simpleMockDaprClient) DeleteStateWithETag(ctx context.Context, storeName, key string, etag *client.ETag, meta map[string]string, so *client.StateOptions) error {
	return nil
}

func (m *simpleMockDaprClient) Encrypt(ctx context.Context, data io.Reader, options client.EncryptOptions) (io.Reader, error) {
	return strings.NewReader("mock-encrypted-data"), nil
}

// Add more methods as needed when they're discovered to be missing