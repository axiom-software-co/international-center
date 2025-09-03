package testing

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/internal/shared/dapr"
	"github.com/stretchr/testify/require"
)

// IntegrationTestSuite provides common integration test setup
type IntegrationTestSuite struct {
	DaprClient   *DaprTestClient
	Environment  *TestEnvironmentSetup
	TestDataKeys []string
}

// NewIntegrationTestSuite creates a new integration test suite
func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	// Skip if infrastructure tests are disabled
	SkipIfNoInfrastructure(t)
	
	// Get environment configuration
	environment := GetTestEnvironmentSetup(t)
	
	// Create Dapr test client
	daprClient := NewDaprTestClient(t)
	
	// Wait for Dapr to be ready
	daprClient.WaitForDapr(t, 30*time.Second)
	
	// Validate test environment
	ValidateTestEnvironment(t, daprClient)
	
	suite := &IntegrationTestSuite{
		DaprClient:   daprClient,
		Environment:  environment,
		TestDataKeys: make([]string, 0),
	}
	
	// Register cleanup
	t.Cleanup(func() {
		suite.Cleanup(t)
	})
	
	return suite
}

// Cleanup cleans up test resources
func (s *IntegrationTestSuite) Cleanup(t *testing.T) {
	if s.DaprClient != nil {
		// Cleanup test data
		s.DaprClient.CleanupStateStore(t, s.TestDataKeys)
		
		// Close Dapr client
		err := s.DaprClient.Close()
		if err != nil {
			t.Logf("Warning: failed to close Dapr client: %v", err)
		}
	}
}

// AddTestDataKey adds a key to the cleanup list
func (s *IntegrationTestSuite) AddTestDataKey(key string) {
	s.TestDataKeys = append(s.TestDataKeys, key)
}

// SaveTestState saves test state and adds key to cleanup list
func (s *IntegrationTestSuite) SaveTestState(t *testing.T, key string, value interface{}) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	err := s.DaprClient.StateStore.Save(ctx, key, value, nil)
	require.NoError(t, err, "Failed to save test state")
	
	s.AddTestDataKey(key)
}

// GetTestState retrieves test state
func (s *IntegrationTestSuite) GetTestState(t *testing.T, key string, target interface{}) bool {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	found, err := s.DaprClient.StateStore.Get(ctx, key, target)
	require.NoError(t, err, "Failed to get test state")
	
	return found
}

// PublishTestEvent publishes a test event
func (s *IntegrationTestSuite) PublishTestEvent(t *testing.T, topic string, data interface{}) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	event := &dapr.EventMessage{
		Topic: topic,
		Data:  map[string]interface{}{"test_data": data},
		Metadata: map[string]string{
			"test":        "true",
			"environment": s.Environment.Environment,
		},
		ContentType: "application/json",
		Source:      "integration-test",
		Type:        "test.event",
		Time:        time.Now(),
	}
	
	err := s.DaprClient.PubSub.PublishEvent(ctx, topic, event)
	require.NoError(t, err, "Failed to publish test event")
}

// InvokeTestService invokes a service for testing
func (s *IntegrationTestSuite) InvokeTestService(t *testing.T, appID, method string, data []byte) *dapr.ServiceResponse {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	req := &dapr.ServiceRequest{
		AppID:       appID,
		MethodName:  method,
		HTTPVerb:    "GET",
		Data:        data,
		ContentType: "application/json",
	}
	
	resp, err := s.DaprClient.ServiceInv.InvokeService(ctx, req)
	require.NoError(t, err, "Failed to invoke test service")
	
	return resp
}

// WaitForEventProcessing waits for async event processing to complete
func (s *IntegrationTestSuite) WaitForEventProcessing(duration time.Duration) {
	time.Sleep(duration)
}

// InfrastructureHealthCheck performs a comprehensive health check
func (s *IntegrationTestSuite) InfrastructureHealthCheck(t *testing.T) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	// Check Dapr health
	err := s.DaprClient.Client.HealthCheck(ctx)
	require.NoError(t, err, "Dapr should be healthy")
	
	// Check state store connectivity
	testKey := s.DaprClient.CreateTestStateKey("health", "check", "connectivity")
	testValue := map[string]string{"timestamp": time.Now().Format(time.RFC3339)}
	
	err = s.DaprClient.StateStore.Save(ctx, testKey, testValue, nil)
	require.NoError(t, err, "State store should be accessible")
	
	var retrieved map[string]string
	found, err := s.DaprClient.StateStore.Get(ctx, testKey, &retrieved)
	require.NoError(t, err, "Should be able to retrieve saved state")
	require.True(t, found, "Saved state should be found")
	require.Equal(t, testValue["timestamp"], retrieved["timestamp"], "Retrieved data should match saved data")
	
	// Cleanup health check data
	s.DaprClient.CleanupStateStore(t, []string{testKey})
}

// DatabaseIntegrationTestSuite provides database-specific integration testing
type DatabaseIntegrationTestSuite struct {
	*IntegrationTestSuite
	DatabaseURL string
}

// NewDatabaseIntegrationTestSuite creates a database integration test suite
func NewDatabaseIntegrationTestSuite(t *testing.T) *DatabaseIntegrationTestSuite {
	baseSuite := NewIntegrationTestSuite(t)
	
	return &DatabaseIntegrationTestSuite{
		IntegrationTestSuite: baseSuite,
		DatabaseURL:          baseSuite.Environment.DatabaseURL,
	}
}

// ContentIntegrationTestSuite provides content-specific integration testing
type ContentIntegrationTestSuite struct {
	*IntegrationTestSuite
	TestContentDir string
}

// NewContentIntegrationTestSuite creates a content integration test suite
func NewContentIntegrationTestSuite(t *testing.T) *ContentIntegrationTestSuite {
	baseSuite := NewIntegrationTestSuite(t)
	
	// Create temporary directory for test content
	testDir := fmt.Sprintf("/tmp/integration-test-content-%d", time.Now().Unix())
	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err, "Failed to create test content directory")
	
	// Register cleanup for test directory
	t.Cleanup(func() {
		os.RemoveAll(testDir)
	})
	
	return &ContentIntegrationTestSuite{
		IntegrationTestSuite: baseSuite,
		TestContentDir:       testDir,
	}
}

// CreateTestContent creates test content in blob storage
func (s *ContentIntegrationTestSuite) CreateTestContent(t *testing.T, contentID, content string) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	blobName := fmt.Sprintf("test-content-%s.txt", contentID)
	err := s.DaprClient.Bindings.UploadBlob(ctx, blobName, []byte(content), "text/plain")
	require.NoError(t, err, "Failed to upload test content")
}

// GatewayIntegrationTestSuite provides gateway-specific integration testing
type GatewayIntegrationTestSuite struct {
	*IntegrationTestSuite
	ContentAPIEndpoint  string
	ServicesAPIEndpoint string
}

// NewGatewayIntegrationTestSuite creates a gateway integration test suite
func NewGatewayIntegrationTestSuite(t *testing.T) *GatewayIntegrationTestSuite {
	baseSuite := NewIntegrationTestSuite(t)
	
	endpoints := baseSuite.DaprClient.ServiceInv.GetServiceEndpoints()
	
	return &GatewayIntegrationTestSuite{
		IntegrationTestSuite: baseSuite,
		ContentAPIEndpoint:   endpoints.ContentAPI,
		ServicesAPIEndpoint:  endpoints.ServicesAPI,
	}
}

// TestGatewayServiceInvocation tests service invocation through gateways
func (s *GatewayIntegrationTestSuite) TestGatewayServiceInvocation(t *testing.T, endpoint, path string) {
	resp := s.InvokeTestService(t, endpoint, path, nil)
	require.NotNil(t, resp, "Gateway should return response")
	require.NotNil(t, resp.Data, "Gateway response should contain data")
}

// MigrationIntegrationTestSuite provides migration-specific integration testing
type MigrationIntegrationTestSuite struct {
	*DatabaseIntegrationTestSuite
	MigrationEndpoint string
}

// NewMigrationIntegrationTestSuite creates a migration integration test suite
func NewMigrationIntegrationTestSuite(t *testing.T) *MigrationIntegrationTestSuite {
	baseSuite := NewDatabaseIntegrationTestSuite(t)
	
	migrationEndpoint := RequireEnvWithDefault(t, "MIGRATION_SERVICE_ENDPOINT", "migration-service")
	
	return &MigrationIntegrationTestSuite{
		DatabaseIntegrationTestSuite: baseSuite,
		MigrationEndpoint:            migrationEndpoint,
	}
}

// TestMigrationServiceHealth tests migration service health
func (s *MigrationIntegrationTestSuite) TestMigrationServiceHealth(t *testing.T) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	err := s.DaprClient.ServiceInv.CheckServiceHealth(ctx, s.MigrationEndpoint)
	require.NoError(t, err, "Migration service should be healthy")
}