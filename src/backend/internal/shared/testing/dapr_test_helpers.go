package testing

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/internal/shared/domain"
	"github.com/stretchr/testify/require"
)

// DaprTestClient provides a test-friendly Dapr client wrapper
type DaprTestClient struct {
	Client      *dapr.Client
	StateStore  *dapr.StateStore
	Bindings    *dapr.Bindings
	PubSub      *dapr.PubSub
	Secrets     *dapr.Secrets
	ServiceInv  *dapr.ServiceInvocation
	Config      *dapr.Configuration
}

// NewDaprTestClient creates a new Dapr test client
func NewDaprTestClient(t *testing.T) *DaprTestClient {
	// Ensure we have required environment variables for testing
	RequireEnv(t, "DAPR_HTTP_PORT")
	RequireEnv(t, "DAPR_GRPC_PORT")
	
	client, err := dapr.NewClient()
	require.NoError(t, err, "Failed to create Dapr client")
	
	return &DaprTestClient{
		Client:     client,
		StateStore: dapr.NewStateStore(client),
		Bindings:   dapr.NewBindings(client),
		PubSub:     dapr.NewPubSub(client),
		Secrets:    dapr.NewSecrets(client),
		ServiceInv: dapr.NewServiceInvocation(client),
		Config:     dapr.NewConfiguration(client),
	}
}

// Close closes the Dapr test client
func (d *DaprTestClient) Close() error {
	return d.Client.Close()
}

// WaitForDapr waits for Dapr to be ready
func (d *DaprTestClient) WaitForDapr(t *testing.T, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Dapr not ready after %v timeout", timeout)
		case <-ticker.C:
			if err := d.Client.HealthCheck(ctx); err == nil {
				return
			}
		}
	}
}

// CleanupStateStore removes all test data from state store
func (d *DaprTestClient) CleanupStateStore(t *testing.T, keys []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	for _, key := range keys {
		err := d.StateStore.Delete(ctx, key, nil)
		if err != nil {
			t.Logf("Warning: failed to cleanup state key %s: %v", key, err)
		}
	}
}

// CreateTestStateKey creates a test-specific state key
func (d *DaprTestClient) CreateTestStateKey(domain, entityType, id string) string {
	return fmt.Sprintf("test:%s:%s:%s", domain, entityType, id)
}

// CreateTestCorrelationContext creates a test correlation context
func CreateTestCorrelationContext(userID string) *domain.CorrelationContext {
	ctx := domain.NewCorrelationContext()
	ctx.SetUserContext(userID, "test-1.0.0")
	return ctx
}

// CreateTestAuditEvent creates a test audit event
func CreateTestAuditEvent(entityType domain.EntityType, entityID, userID string, operationType domain.AuditEventType) *domain.AuditEvent {
	event := domain.NewAuditEvent(entityType, entityID, operationType, userID)
	event.SetEnvironmentContext("test", "test-1.0.0")
	return event
}

// RequireEnv requires an environment variable to be set for testing
func RequireEnv(t *testing.T, key string) string {
	value := os.Getenv(key)
	if value == "" {
		t.Fatalf("Environment variable %s is required for integration tests", key)
	}
	return value
}

// RequireEnvWithDefault gets environment variable with default for testing
func RequireEnvWithDefault(t *testing.T, key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		t.Logf("Using default value for %s: %s", key, defaultValue)
		return defaultValue
	}
	return value
}

// SkipIfNoInfrastructure skips the test if infrastructure is not available
func SkipIfNoInfrastructure(t *testing.T) {
	if os.Getenv("SKIP_INFRASTRUCTURE_TESTS") == "true" {
		t.Skip("Skipping test because SKIP_INFRASTRUCTURE_TESTS=true")
	}
}

// CreateIntegrationTestContext creates a context with timeout for integration tests
func CreateIntegrationTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 15*time.Second)
}

// CreateUnitTestContext creates a context with timeout for unit tests
func CreateUnitTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// TestEnvironmentSetup represents test environment configuration
type TestEnvironmentSetup struct {
	Environment     string
	DatabaseURL     string
	DaprHTTPPort    string
	DaprGRPCPort    string
	StateStoreName  string
	PubSubName      string
	SecretStoreName string
	BlobBindingName string
}

// GetTestEnvironmentSetup gets test environment configuration
func GetTestEnvironmentSetup(t *testing.T) *TestEnvironmentSetup {
	return &TestEnvironmentSetup{
		Environment:     RequireEnvWithDefault(t, "ENVIRONMENT", "test"),
		DatabaseURL:     RequireEnv(t, "DATABASE_URL"),
		DaprHTTPPort:    RequireEnv(t, "DAPR_HTTP_PORT"),
		DaprGRPCPort:    RequireEnv(t, "DAPR_GRPC_PORT"),
		StateStoreName:  RequireEnvWithDefault(t, "DAPR_STATE_STORE_NAME", "statestore-postgresql-test"),
		PubSubName:      RequireEnvWithDefault(t, "DAPR_PUBSUB_NAME", "pubsub-redis-test"),
		SecretStoreName: RequireEnvWithDefault(t, "DAPR_SECRET_STORE_NAME", "secretstore-vault-test"),
		BlobBindingName: RequireEnvWithDefault(t, "DAPR_BLOB_BINDING_NAME", "blob-storage-test"),
	}
}

// ValidateTestEnvironment validates that the test environment is properly configured
func ValidateTestEnvironment(t *testing.T, client *DaprTestClient) {
	ctx := CreateIntegrationTestContext()
	defer ctx()
	
	// Test state store connectivity
	testKey := client.CreateTestStateKey("test", "validation", "connectivity")
	testValue := map[string]string{"test": "data"}
	
	err := client.StateStore.Save(ctx, testKey, testValue, nil)
	require.NoError(t, err, "State store should be accessible")
	
	// Cleanup test data
	client.CleanupStateStore(t, []string{testKey})
}

// MockDaprComponents provides mock implementations for unit testing
type MockDaprComponents struct {
	StateStore  map[string]interface{}
	PubSubEvents []MockPubSubEvent
	Secrets     map[string]string
	Bindings    map[string][]byte
}

// MockPubSubEvent represents a mock pub/sub event
type MockPubSubEvent struct {
	Topic string
	Data  interface{}
}

// NewMockDaprComponents creates new mock Dapr components
func NewMockDaprComponents() *MockDaprComponents {
	return &MockDaprComponents{
		StateStore: make(map[string]interface{}),
		Secrets:    make(map[string]string),
		Bindings:   make(map[string][]byte),
	}
}

// AddMockSecret adds a mock secret
func (m *MockDaprComponents) AddMockSecret(key, value string) {
	m.Secrets[key] = value
}

// AddMockBinding adds a mock binding response
func (m *MockDaprComponents) AddMockBinding(operation string, data []byte) {
	m.Bindings[operation] = data
}

// GetMockPubSubEvents returns all published events
func (m *MockDaprComponents) GetMockPubSubEvents(topic string) []MockPubSubEvent {
	var events []MockPubSubEvent
	for _, event := range m.PubSubEvents {
		if event.Topic == topic {
			events = append(events, event)
		}
	}
	return events
}