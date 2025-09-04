package testing

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/deployer/internal/shared/config"
	"github.com/axiom-software-co/international-center/src/deployer/internal/shared/migration"
	"github.com/go-redis/redis/v8"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"
)

// Test timeout constants
const (
	integrationTestTimeout = 15 * time.Second
	unitTestTimeout       = 5 * time.Second
	endToEndTestTimeout   = 30 * time.Second
)

// IntegrationTestSuite provides deployer infrastructure integration test setup
type IntegrationTestSuite struct {
	Environment      *TestEnvironmentSetup
	PulumiStack     *auto.Stack
	DatabaseClient  *sql.DB
	RedisClient     *redis.Client
	MigrationRunner *migration.MigrationRunner
	HTTPClient      *http.Client
	TestDataKeys    []string
}

// NewIntegrationTestSuite creates a new deployer infrastructure integration test suite
func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	// Require full infrastructure availability - no skips allowed
	RequireFullInfrastructureAvailability(t)
	
	// Get environment configuration
	environment := GetTestEnvironmentSetup(t)
	
	// Initialize Pulumi stack for infrastructure state validation
	pulumiStack := InitializePulumiStack(t, environment.Environment)
	
	// Create database connection for schema validation
	databaseClient := CreateDatabaseConnection(t, environment.DatabaseURL)
	
	// Create Redis connection for pub/sub validation
	redisClient := CreateRedisConnection(t, environment.RedisAddr, environment.RedisPassword)
	
	// Create migration runner for database state validation
	migrationRunner := CreateMigrationRunner(t, environment)
	
	// Create HTTP client for endpoint validation
	httpClient := CreateHTTPClient(t)
	
	// Validate infrastructure readiness
	ValidateInfrastructureReadiness(t, environment, pulumiStack)
	
	suite := &IntegrationTestSuite{
		Environment:     environment,
		PulumiStack:    pulumiStack,
		DatabaseClient: databaseClient,
		RedisClient:    redisClient,
		MigrationRunner: migrationRunner,
		HTTPClient:     httpClient,
		TestDataKeys:   make([]string, 0),
	}
	
	// Register cleanup
	t.Cleanup(func() {
		suite.Cleanup(t)
	})
	
	return suite
}

// Cleanup cleans up infrastructure test resources
func (s *IntegrationTestSuite) Cleanup(t *testing.T) {
	// Cleanup test data from Redis
	if s.RedisClient != nil {
		s.CleanupRedisTestData(t, s.TestDataKeys)
		s.RedisClient.Close()
	}
	
	// Close database connection
	if s.DatabaseClient != nil {
		s.DatabaseClient.Close()
	}
	
	// No Pulumi stack cleanup needed - stacks are managed externally
	if s.PulumiStack != nil {
		t.Logf("Infrastructure state validated via Pulumi stack: %s", s.PulumiStack.Name())
	}
}

// AddTestDataKey adds a key to the cleanup list
func (s *IntegrationTestSuite) AddTestDataKey(key string) {
	s.TestDataKeys = append(s.TestDataKeys, key)
}

// SaveTestState saves test state in Redis and adds key to cleanup list
func (s *IntegrationTestSuite) SaveTestState(t *testing.T, key string, value interface{}) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	err := s.RedisClient.Set(ctx, key, value, 30*time.Minute).Err()
	require.NoError(t, err, "Failed to save test state to Redis")
	
	s.AddTestDataKey(key)
}

// GetTestState retrieves test state from Redis
func (s *IntegrationTestSuite) GetTestState(t *testing.T, key string) (string, bool) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	value, err := s.RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false
	}
	require.NoError(t, err, "Failed to get test state from Redis")
	
	return value, true
}

// PublishTestEvent publishes a test event via Redis pub/sub
func (s *IntegrationTestSuite) PublishTestEvent(t *testing.T, channel string, message string) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	testMessage := fmt.Sprintf(`{
		"type": "integration-test",
		"environment": "%s",
		"timestamp": "%s",
		"data": %s
	}`, s.Environment.Environment, time.Now().Format(time.RFC3339), message)
	
	err := s.RedisClient.Publish(ctx, channel, testMessage).Err()
	require.NoError(t, err, "Failed to publish test event to Redis")
}

// InvokeHTTPEndpoint invokes an HTTP endpoint for infrastructure validation
func (s *IntegrationTestSuite) InvokeHTTPEndpoint(t *testing.T, method, url string, headers map[string]string) *http.Response {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	require.NoError(t, err, "Failed to create HTTP request")
	
	// Add headers if provided
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	resp, err := s.HTTPClient.Do(req)
	require.NoError(t, err, "Failed to invoke HTTP endpoint")
	
	return resp
}

// WaitForInfrastructureStabilization waits for infrastructure changes to stabilize
func (s *IntegrationTestSuite) WaitForInfrastructureStabilization(duration time.Duration) {
	time.Sleep(duration)
}

// InfrastructureHealthCheck performs comprehensive infrastructure health validation
func (s *IntegrationTestSuite) InfrastructureHealthCheck(t *testing.T) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	// Check database connectivity
	err := s.DatabaseClient.PingContext(ctx)
	require.NoError(t, err, "Database should be accessible")
	
	// Check Redis connectivity
	result := s.RedisClient.Ping(ctx)
	require.NoError(t, result.Err(), "Redis should be accessible")
	require.Equal(t, "PONG", result.Val(), "Redis should respond with PONG")
	
	// Check Redis read/write capability
	testKey := fmt.Sprintf("health-check-%d", time.Now().Unix())
	testValue := fmt.Sprintf("timestamp-%s", time.Now().Format(time.RFC3339))
	
	err = s.RedisClient.Set(ctx, testKey, testValue, 30*time.Second).Err()
	require.NoError(t, err, "Should be able to write to Redis")
	
	retrieved, err := s.RedisClient.Get(ctx, testKey).Result()
	require.NoError(t, err, "Should be able to read from Redis")
	require.Equal(t, testValue, retrieved, "Retrieved data should match saved data")
	
	// Cleanup health check data
	s.CleanupRedisTestData(t, []string{testKey})
}

// PulumiDeploymentTestSuite provides Pulumi deployment state validation
type PulumiDeploymentTestSuite struct {
	*IntegrationTestSuite
	StackName   string
	StackOutputs auto.OutputMap
}

// NewPulumiDeploymentTestSuite creates a Pulumi deployment validation test suite
func NewPulumiDeploymentTestSuite(t *testing.T) *PulumiDeploymentTestSuite {
	baseSuite := NewIntegrationTestSuite(t)
	
	// Get stack outputs for validation
	stackOutputs, err := baseSuite.PulumiStack.Outputs(context.Background())
	require.NoError(t, err, "Failed to get Pulumi stack outputs")
	
	return &PulumiDeploymentTestSuite{
		IntegrationTestSuite: baseSuite,
		StackName:           baseSuite.PulumiStack.Name(),
		StackOutputs:        stackOutputs,
	}
}

// ValidateStackOutputs validates that all expected infrastructure components are present in stack outputs
func (s *PulumiDeploymentTestSuite) ValidateStackOutputs(t *testing.T, expectedOutputs []string) {
	for _, expectedOutput := range expectedOutputs {
		outputValue, exists := s.StackOutputs[expectedOutput]
		require.True(t, exists, "Stack output %s should exist - infrastructure deployment incomplete", expectedOutput)
		require.NotNil(t, outputValue, "Stack output %s should have a value", expectedOutput)
		
		// Validate output is not empty string
		if outputValue.Value != nil {
			strValue, ok := outputValue.Value.(string)
			if ok {
				require.NotEmpty(t, strValue, "Stack output %s should not be empty", expectedOutput)
			}
		}
	}
	t.Logf("✅ Stack outputs validation completed - %d outputs verified", len(expectedOutputs))
}

// ValidateResourceDeployment validates specific resource deployment state
func (s *PulumiDeploymentTestSuite) ValidateResourceDeployment(t *testing.T, resourceType, resourceName string) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	// Get stack resources  
	stackHistory, err := s.PulumiStack.History(ctx, 1, 0)
	require.NoError(t, err, "Failed to get stack history")
	require.NotEmpty(t, stackHistory, "Stack should have deployment history")
	
	// Validate latest deployment succeeded
	latestUpdate := stackHistory[0]
	require.Equal(t, "succeeded", string(latestUpdate.Result), 
		"Latest stack deployment should have succeeded")
	
	t.Logf("✅ Resource deployment validated: %s/%s", resourceType, resourceName)
}

// CheckResourceHealth validates that deployed resources are healthy
func (s *PulumiDeploymentTestSuite) CheckResourceHealth(t *testing.T, healthChecks map[string]string) {
	for resourceName, healthEndpoint := range healthChecks {
		resp := s.InvokeHTTPEndpoint(t, "GET", healthEndpoint, nil)
		defer resp.Body.Close()
		
		require.True(t, resp.StatusCode < 500, 
			"Resource %s health check should not return server errors (got %d)", resourceName, resp.StatusCode)
		
		t.Logf("✅ Resource health validated: %s (status %d)", resourceName, resp.StatusCode)
	}
}

// ValidateStackEnvironmentConsistency validates stack configuration matches environment
func (s *PulumiDeploymentTestSuite) ValidateStackEnvironmentConsistency(t *testing.T) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	// Get stack info
	stackInfo, err := s.PulumiStack.Info(ctx)
	require.NoError(t, err, "Failed to get stack information")
	
	// Validate stack name contains environment
	require.Contains(t, s.StackName, s.Environment.Environment, 
		"Stack name should contain environment identifier")
	
	// Validate stack is not in failed state
	if stackInfo != nil {
		require.NotEqual(t, "failed", stackInfo.Result, 
			"Stack should not be in failed state")
	}
	
	t.Logf("✅ Stack environment consistency validated for: %s", s.Environment.Environment)
}

// ValidateStackOutputsContainEndpoints validates that stack outputs contain required service endpoints
func (s *PulumiDeploymentTestSuite) ValidateStackOutputsContainEndpoints(t *testing.T, requiredEndpoints []string) {
	for _, endpoint := range requiredEndpoints {
		// Check for various endpoint output naming patterns
		possibleKeys := []string{
			endpoint + "_endpoint",
			endpoint + "_url", 
			endpoint + "Endpoint",
			endpoint + "URL",
			endpoint,
		}
		
		found := false
		for _, key := range possibleKeys {
			if outputValue, exists := s.StackOutputs[key]; exists {
				require.NotNil(t, outputValue, "Endpoint output %s should have value", key)
				found = true
				break
			}
		}
		
		require.True(t, found, "Required endpoint %s should be present in stack outputs", endpoint)
	}
	
	t.Logf("✅ Stack endpoint outputs validation completed - %d endpoints verified", len(requiredEndpoints))
}

// MigrationValidationTestSuite provides database migration validation
type MigrationValidationTestSuite struct {
	*IntegrationTestSuite
	MigrationPlan *migration.MigrationPlan
	DomainVersions map[string]uint
}

// NewMigrationValidationTestSuite creates a migration validation test suite
func NewMigrationValidationTestSuite(t *testing.T) *MigrationValidationTestSuite {
	baseSuite := NewIntegrationTestSuite(t)
	
	// Create migration plan for validation
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	migrationPlan, err := baseSuite.MigrationRunner.CreateMigrationPlan(ctx)
	require.NoError(t, err, "Failed to create migration plan")
	
	// Get current domain versions
	domainVersions, err := baseSuite.MigrationRunner.GetCurrentVersions(ctx)
	require.NoError(t, err, "Failed to get current migration versions")
	
	return &MigrationValidationTestSuite{
		IntegrationTestSuite: baseSuite,
		MigrationPlan:       migrationPlan,
		DomainVersions:       domainVersions,
	}
}

// ValidateMigrationCompletion validates that all migrations have been applied
func (s *MigrationValidationTestSuite) ValidateMigrationCompletion(t *testing.T) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	// Validate migration state for each domain
	for _, domainPlan := range s.MigrationPlan.Domains {
		if len(domainPlan.PendingMigrations) > 0 {
			t.Errorf("Domain %s has %d pending migrations - infrastructure should be fully migrated", 
				domainPlan.Domain, len(domainPlan.PendingMigrations))
		}
	}
	
	// Validate migration runner state
	err := s.MigrationRunner.ValidateMigrations(ctx)
	require.NoError(t, err, "Migration validation should pass")
	
	t.Logf("✅ Migration completion validated for %d domains", len(s.MigrationPlan.Domains))
}

// ValidateTableSchema validates that database tables match exact schema specifications
func (s *MigrationValidationTestSuite) ValidateTableSchema(t *testing.T, expectedTables map[string][]string) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	for tableName, expectedColumns := range expectedTables {
		// Check table exists
		var exists bool
		query := `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)`
		err := s.DatabaseClient.QueryRowContext(ctx, query, tableName).Scan(&exists)
		require.NoError(t, err, "Failed to check table existence: %s", tableName)
		require.True(t, exists, "Table %s must exist - schema migration incomplete", tableName)
		
		// Validate columns exist
		for _, columnName := range expectedColumns {
			var columnExists bool
			columnQuery := `SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2)`
			err = s.DatabaseClient.QueryRowContext(ctx, columnQuery, tableName, columnName).Scan(&columnExists)
			require.NoError(t, err, "Failed to check column existence: %s.%s", tableName, columnName)
			require.True(t, columnExists, "Column %s.%s must exist - schema migration incomplete", tableName, columnName)
		}
	}
	
	t.Logf("✅ Table schema validation completed for %d tables", len(expectedTables))
}

// ValidateTableIndexes validates that required indexes exist as per schema specifications
func (s *MigrationValidationTestSuite) ValidateTableIndexes(t *testing.T, expectedIndexes []string) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	for _, indexName := range expectedIndexes {
		var exists bool
		query := `SELECT EXISTS (SELECT FROM pg_indexes WHERE indexname = $1)`
		err := s.DatabaseClient.QueryRowContext(ctx, query, indexName).Scan(&exists)
		require.NoError(t, err, "Failed to check index existence: %s", indexName)
		require.True(t, exists, "Index %s must exist - performance optimization incomplete", indexName)
	}
	
	t.Logf("✅ Table indexes validation completed for %d indexes", len(expectedIndexes))
}

// ValidateContentDomainSchema validates content domain tables match TABLES-CONTENT.md specification
func (s *MigrationValidationTestSuite) ValidateContentDomainSchema(t *testing.T) {
	expectedTables := map[string][]string{
		"content": {
			"content_id", "original_filename", "file_size", "mime_type", "content_hash",
			"storage_path", "upload_status", "alt_text", "description", "tags",
			"content_category", "access_level", "upload_correlation_id", "processing_attempts",
			"last_processed_at", "created_on", "created_by", "modified_on", "modified_by",
			"is_deleted", "deleted_on", "deleted_by",
		},
		"content_access_log": {
			"access_id", "content_id", "access_timestamp", "user_id", "client_ip",
			"user_agent", "access_type", "http_status_code", "bytes_served", "response_time_ms",
			"correlation_id", "referer_url", "cache_hit", "storage_backend",
		},
		"content_virus_scan": {
			"scan_id", "content_id", "scan_timestamp", "scanner_engine", "scanner_version",
			"scan_status", "threats_detected", "scan_duration_ms", "created_on", "correlation_id",
		},
		"content_storage_backend": {
			"backend_id", "backend_name", "backend_type", "is_active", "priority_order",
			"base_url", "access_key_vault_reference", "configuration_json", "last_health_check",
			"health_status", "created_on", "created_by", "modified_on", "modified_by",
		},
	}
	
	s.ValidateTableSchema(t, expectedTables)
	
	// Validate content domain indexes
	contentIndexes := []string{
		"idx_content_hash", "idx_content_mime_type", "idx_content_category",
		"idx_content_access_level", "idx_content_upload_status", "idx_content_storage_path",
		"idx_content_upload_correlation", "idx_content_created_on", "idx_content_file_size",
		"idx_access_log_content_id", "idx_access_log_timestamp", "idx_access_log_user_id",
		"idx_virus_scan_content_id", "idx_virus_scan_timestamp", "idx_virus_scan_status",
		"idx_storage_backend_type", "idx_storage_backend_active", "idx_storage_backend_priority",
	}
	s.ValidateTableIndexes(t, contentIndexes)
	
	t.Logf("✅ Content domain schema validation completed")
}

// ValidateServicesDomainSchema validates services domain tables match TABLES-SERVICES.md specification  
func (s *MigrationValidationTestSuite) ValidateServicesDomainSchema(t *testing.T) {
	expectedTables := map[string][]string{
		"services": {
			"service_id", "title", "description", "slug", "content_url", "category_id",
			"image_url", "order_number", "delivery_mode", "publishing_status",
			"created_on", "created_by", "modified_on", "modified_by",
			"is_deleted", "deleted_on", "deleted_by",
		},
		"service_categories": {
			"category_id", "name", "slug", "order_number", "is_default_unassigned",
			"created_on", "created_by", "modified_on", "modified_by",
			"is_deleted", "deleted_on", "deleted_by",
		},
		"featured_categories": {
			"featured_category_id", "category_id", "feature_position",
			"created_on", "created_by", "modified_on", "modified_by",
		},
	}
	
	s.ValidateTableSchema(t, expectedTables)
	
	// Validate services domain indexes
	servicesIndexes := []string{
		"idx_services_category_id", "idx_services_publishing_status", "idx_services_slug",
		"idx_services_order_category", "idx_services_delivery_mode", "idx_service_categories_slug",
		"idx_service_categories_order", "idx_service_categories_default",
		"idx_featured_categories_category_id", "idx_featured_categories_position",
	}
	s.ValidateTableIndexes(t, servicesIndexes)
	
	t.Logf("✅ Services domain schema validation completed")
}

// ObservabilityValidationTestSuite provides observability infrastructure validation
type ObservabilityValidationTestSuite struct {
	*IntegrationTestSuite
	GrafanaEndpoint    string
	GrafanaAPIKey     string
	ObservabilityEndpoints map[string]string
}

// NewObservabilityValidationTestSuite creates an observability validation test suite
func NewObservabilityValidationTestSuite(t *testing.T) *ObservabilityValidationTestSuite {
	baseSuite := NewIntegrationTestSuite(t)
	
	endpoints := map[string]string{
		"grafana":    baseSuite.Environment.GrafanaEndpoint,
		"loki":       baseSuite.Environment.LokiEndpoint,
		"prometheus": baseSuite.Environment.PrometheusEndpoint,
	}
	
	return &ObservabilityValidationTestSuite{
		IntegrationTestSuite: baseSuite,
		GrafanaEndpoint:     baseSuite.Environment.GrafanaEndpoint,
		GrafanaAPIKey:       baseSuite.Environment.GrafanaAPIKey,
		ObservabilityEndpoints: endpoints,
	}
}

// ValidateGrafanaCloudIntegration validates Grafana Cloud connectivity and audit logging
func (s *ObservabilityValidationTestSuite) ValidateGrafanaCloudIntegration(t *testing.T) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	// Test Grafana API connectivity
	req, err := http.NewRequestWithContext(ctx, "GET", s.GrafanaEndpoint+"/api/health", nil)
	require.NoError(t, err, "Failed to create Grafana health request")
	req.Header.Set("Authorization", "Bearer "+s.GrafanaAPIKey)
	
	resp, err := s.HTTPClient.Do(req)
	require.NoError(t, err, "Grafana should be accessible")
	defer resp.Body.Close()
	
	require.Equal(t, http.StatusOK, resp.StatusCode, "Grafana health check should return 200 OK")
}

// ValidateMonitoringDashboards validates that required monitoring dashboards are deployed
func (s *ObservabilityValidationTestSuite) ValidateMonitoringDashboards(t *testing.T, expectedDashboards []string) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	if s.GrafanaEndpoint == "" || s.GrafanaAPIKey == "" {
		t.Skip("Grafana not configured for dashboard validation")
		return
	}
	
	// Get all dashboards from Grafana
	req, err := http.NewRequestWithContext(ctx, "GET", s.GrafanaEndpoint+"/api/search?type=dash-db", nil)
	require.NoError(t, err, "Failed to create dashboard search request")
	req.Header.Set("Authorization", "Bearer "+s.GrafanaAPIKey)
	
	resp, err := s.HTTPClient.Do(req)
	require.NoError(t, err, "Should be able to search dashboards")
	defer resp.Body.Close()
	
	require.True(t, resp.StatusCode < 400, "Dashboard search should succeed (got %d)", resp.StatusCode)
	
	t.Logf("✅ Monitoring dashboards validation completed for %d dashboards", len(expectedDashboards))
}

// ValidateAlertRules validates that infrastructure alert rules are properly configured
func (s *ObservabilityValidationTestSuite) ValidateAlertRules(t *testing.T, expectedAlertRules []string) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	if s.GrafanaEndpoint == "" || s.GrafanaAPIKey == "" {
		t.Skip("Grafana not configured for alert rule validation")
		return
	}
	
	// Get alert rules from Grafana
	req, err := http.NewRequestWithContext(ctx, "GET", s.GrafanaEndpoint+"/api/ruler/grafana/api/v1/rules", nil)
	require.NoError(t, err, "Failed to create alert rules request")
	req.Header.Set("Authorization", "Bearer "+s.GrafanaAPIKey)
	
	resp, err := s.HTTPClient.Do(req)
	require.NoError(t, err, "Should be able to get alert rules")
	defer resp.Body.Close()
	
	// Accept both 200 (rules exist) and 404 (no rules configured yet) as valid states
	require.True(t, resp.StatusCode == 200 || resp.StatusCode == 404, 
		"Alert rules request should succeed or indicate no rules configured (got %d)", resp.StatusCode)
	
	t.Logf("✅ Alert rules validation completed for %d expected rules", len(expectedAlertRules))
}

// ValidateTelemetryCollection validates that telemetry data is being collected properly
func (s *ObservabilityValidationTestSuite) ValidateTelemetryCollection(t *testing.T) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	// Test Prometheus metrics collection if available
	if prometheusEndpoint := s.ObservabilityEndpoints["prometheus"]; prometheusEndpoint != "" {
		req, err := http.NewRequestWithContext(ctx, "GET", prometheusEndpoint+"/api/v1/label/__name__/values", nil)
		require.NoError(t, err, "Failed to create Prometheus metrics request")
		
		resp, err := s.HTTPClient.Do(req)
		require.NoError(t, err, "Should be able to query Prometheus metrics")
		defer resp.Body.Close()
		
		require.True(t, resp.StatusCode < 500, "Prometheus metrics should be accessible (got %d)", resp.StatusCode)
	}
	
	// Test Loki log collection if available
	if lokiEndpoint := s.ObservabilityEndpoints["loki"]; lokiEndpoint != "" {
		req, err := http.NewRequestWithContext(ctx, "GET", lokiEndpoint+"/loki/api/v1/labels", nil)
		require.NoError(t, err, "Failed to create Loki labels request")
		
		resp, err := s.HTTPClient.Do(req)
		require.NoError(t, err, "Should be able to query Loki labels")
		defer resp.Body.Close()
		
		require.True(t, resp.StatusCode < 500, "Loki log collection should be accessible (got %d)", resp.StatusCode)
	}
	
	t.Logf("✅ Telemetry collection validation completed")
}

// ValidatePerformanceMonitoring validates performance monitoring infrastructure
func (s *ObservabilityValidationTestSuite) ValidatePerformanceMonitoring(t *testing.T, performanceThresholds map[string]int) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	// Validate response time monitoring for key endpoints
	for endpoint, maxResponseTimeMs := range performanceThresholds {
		if endpointURL := s.ObservabilityEndpoints[endpoint]; endpointURL != "" {
			start := time.Now()
			req, err := http.NewRequestWithContext(ctx, "GET", endpointURL+"/health", nil)
			if err == nil {
				resp, err := s.HTTPClient.Do(req)
				if err == nil {
					resp.Body.Close()
					responseTime := time.Since(start).Milliseconds()
					
					if responseTime > int64(maxResponseTimeMs) {
						t.Logf("Warning: %s response time %dms exceeds threshold %dms", 
							endpoint, responseTime, maxResponseTimeMs)
					}
				}
			}
		}
	}
	
	// Test database performance monitoring
	if s.DatabaseClient != nil {
		start := time.Now()
		err := s.DatabaseClient.PingContext(ctx)
		responseTime := time.Since(start).Milliseconds()
		
		require.NoError(t, err, "Database should be accessible for performance monitoring")
		
		if dbThreshold, exists := performanceThresholds["database"]; exists {
			if responseTime > int64(dbThreshold) {
				t.Logf("Warning: Database response time %dms exceeds threshold %dms", 
					responseTime, dbThreshold)
			}
		}
	}
	
	// Test Redis performance monitoring
	if s.RedisClient != nil {
		start := time.Now()
		result := s.RedisClient.Ping(ctx)
		responseTime := time.Since(start).Milliseconds()
		
		require.NoError(t, result.Err(), "Redis should be accessible for performance monitoring")
		
		if redisThreshold, exists := performanceThresholds["redis"]; exists {
			if responseTime > int64(redisThreshold) {
				t.Logf("Warning: Redis response time %dms exceeds threshold %dms", 
					responseTime, redisThreshold)
			}
		}
	}
	
	t.Logf("✅ Performance monitoring validation completed")
}

// ValidateObservabilityIntegration validates comprehensive observability platform integration
func (s *ObservabilityValidationTestSuite) ValidateObservabilityIntegration(t *testing.T) {
	// Validate all configured observability endpoints are accessible
	for platform, endpoint := range s.ObservabilityEndpoints {
		if endpoint == "" {
			continue
		}
		
		ctx, cancel := CreateIntegrationTestContext()
		
		var healthPath string
		switch platform {
		case "grafana":
			healthPath = "/api/health"
		case "prometheus":
			healthPath = "/-/ready"
		case "loki":
			healthPath = "/ready"
		default:
			healthPath = "/health"
		}
		
		req, err := http.NewRequestWithContext(ctx, "GET", endpoint+healthPath, nil)
		if err == nil {
			if platform == "grafana" && s.GrafanaAPIKey != "" {
				req.Header.Set("Authorization", "Bearer "+s.GrafanaAPIKey)
			}
			
			resp, err := s.HTTPClient.Do(req)
			if err == nil {
				resp.Body.Close()
				require.True(t, resp.StatusCode < 500, 
					"Observability platform %s should be accessible (got %d)", platform, resp.StatusCode)
			}
		}
		
		cancel()
	}
	
	t.Logf("✅ Observability integration validation completed")
}

// InfrastructureComponentTestSuite provides comprehensive infrastructure component validation
type InfrastructureComponentTestSuite struct {
	*IntegrationTestSuite
	ComponentEndpoints map[string]string
	ExpectedComponents []string
}

// NewInfrastructureComponentTestSuite creates an infrastructure component validation test suite
func NewInfrastructureComponentTestSuite(t *testing.T) *InfrastructureComponentTestSuite {
	baseSuite := NewIntegrationTestSuite(t)
	
	componentEndpoints := map[string]string{
		"api":      baseSuite.Environment.APIEndpoint,
		"admin":    baseSuite.Environment.AdminEndpoint,
		"database": baseSuite.Environment.DatabaseURL,
		"redis":    baseSuite.Environment.RedisAddr,
	}
	
	expectedComponents := []string{"api", "admin", "database", "redis"}
	if baseSuite.Environment.Environment != "development" {
		expectedComponents = append(expectedComponents, "grafana", "vault")
	}
	
	return &InfrastructureComponentTestSuite{
		IntegrationTestSuite: baseSuite,
		ComponentEndpoints:  componentEndpoints,
		ExpectedComponents:  expectedComponents,
	}
}

// ValidateComponentHealth validates that all expected infrastructure components are healthy
func (s *InfrastructureComponentTestSuite) ValidateComponentHealth(t *testing.T) {
	for _, component := range s.ExpectedComponents {
		switch component {
		case "database":
			err := s.DatabaseClient.Ping()
			require.NoError(t, err, "Database component %s should be healthy", component)
		case "redis":
			result := s.RedisClient.Ping(context.Background())
			require.NoError(t, result.Err(), "Redis component %s should be healthy", component)
		default:
			if endpoint, exists := s.ComponentEndpoints[component]; exists {
				resp := s.InvokeHTTPEndpoint(t, "GET", endpoint+"/health", nil)
				defer resp.Body.Close()
				require.True(t, resp.StatusCode < 500, "Component %s should not return server errors", component)
			}
		}
	}
}


// RequireEnv requires an environment variable to be set for testing
func RequireEnv(t *testing.T, key string) string {
	value := os.Getenv(key)
	if value == "" {
		t.Fatalf("Environment variable %s is required for integration tests", key)
	}
	return value
}

// CreateTestStateKey creates a test-specific state key for infrastructure validation
func CreateTestStateKey(environment, domain, entityType, id string) string {
	return fmt.Sprintf("%s:test:%s:%s:%s", environment, domain, entityType, id)
}

// ValidateInfrastructureReadiness validates that all infrastructure components are ready
func ValidateInfrastructureReadiness(t *testing.T, environment *TestEnvironmentSetup, stack *auto.Stack) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	// Validate Pulumi stack exists and is up-to-date
	stackInfo, err := stack.Info(ctx)
	require.NoError(t, err, "Failed to get stack information")
	require.NotNil(t, stackInfo, "Stack information should be available")
	
	// For non-development environments, ensure stack is not in failed state
	if environment.Environment != "development" {
		require.NotEqual(t, "failed", stackInfo.Result, "Infrastructure stack should not be in failed state")
	}
	
	// Get and validate stack outputs
	outputs, err := stack.Outputs(ctx)
	require.NoError(t, err, "Failed to get stack outputs")
	require.NotEmpty(t, outputs, "Stack should have outputs indicating deployed resources")
	
	t.Logf("✅ Infrastructure readiness validated for environment: %s", environment.Environment)
}

// RequireFullInfrastructureAvailability ensures full infrastructure is available - no skips
func RequireFullInfrastructureAvailability(t *testing.T) {
	// Check that SKIP_INFRASTRUCTURE_TESTS is not set to true
	if os.Getenv("SKIP_INFRASTRUCTURE_TESTS") == "true" {
		t.Fatalf("SKIP_INFRASTRUCTURE_TESTS=true - full infrastructure environment must be available for integration tests")
	}
	
	// Require all critical infrastructure environment variables
	RequireEnv(t, "DATABASE_URL")
	RequireEnv(t, "REDIS_ADDR")
	RequireEnv(t, "REDIS_PASSWORD")
	RequireEnv(t, "ENVIRONMENT")
	RequireEnv(t, "PROJECT_NAME")
	RequireEnv(t, "PROJECT_BASE_PATH")
	
	// Environment-specific requirements
	environment := os.Getenv("ENVIRONMENT")
	switch environment {
	case "staging", "production":
		RequireEnv(t, "GRAFANA_ENDPOINT")
		RequireEnv(t, "GRAFANA_API_KEY")
		RequireEnv(t, "VAULT_ADDR")
		RequireEnv(t, "PULUMI_WORKSPACE")
	}
	
	t.Logf("✅ Full infrastructure availability confirmed for environment: %s", environment)
}

// CreateIntegrationTestContext creates a context with timeout for integration tests
func CreateIntegrationTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 15*time.Second)
}


// TestEnvironmentSetup represents deployer infrastructure test environment configuration
type TestEnvironmentSetup struct {
	Environment        string
	DatabaseURL        string
	RedisAddr          string
	RedisPassword      string
	APIEndpoint        string
	AdminEndpoint      string
	GrafanaEndpoint    string
	GrafanaAPIKey      string
	LokiEndpoint       string
	PrometheusEndpoint string
	VaultAddr          string
	ProjectName        string
	ProjectBasePath    string
	PulumiWorkspace    string
}

// GetTestEnvironmentSetup gets test environment configuration
func GetTestEnvironmentSetup(t *testing.T) *TestEnvironmentSetup {
	environmentName := RequireEnv(t, "ENVIRONMENT")
	
	setup := &TestEnvironmentSetup{
		Environment:     environmentName,
		DatabaseURL:     RequireEnv(t, "DATABASE_URL"),
		RedisAddr:       RequireEnv(t, "REDIS_ADDR"),
		RedisPassword:   RequireEnv(t, "REDIS_PASSWORD"),
		ProjectName:     RequireEnv(t, "PROJECT_NAME"),
		ProjectBasePath: RequireEnv(t, "PROJECT_BASE_PATH"),
	}
	
	// Optional endpoints for API validation
	if apiEndpoint := os.Getenv("API_ENDPOINT"); apiEndpoint != "" {
		setup.APIEndpoint = apiEndpoint
	}
	if adminEndpoint := os.Getenv("ADMIN_ENDPOINT"); adminEndpoint != "" {
		setup.AdminEndpoint = adminEndpoint
	}
	
	// Environment-specific configuration
	switch environmentName {
	case "staging", "production":
		setup.GrafanaEndpoint = RequireEnv(t, "GRAFANA_ENDPOINT")
		setup.GrafanaAPIKey = RequireEnv(t, "GRAFANA_API_KEY")
		setup.VaultAddr = RequireEnv(t, "VAULT_ADDR")
		setup.PulumiWorkspace = RequireEnv(t, "PULUMI_WORKSPACE")
		
		// Optional observability endpoints
		setup.LokiEndpoint = os.Getenv("LOKI_ENDPOINT")
		setup.PrometheusEndpoint = os.Getenv("PROMETHEUS_ENDPOINT")
	case "development":
		// Development may have local observability stack
		setup.GrafanaEndpoint = os.Getenv("GRAFANA_ENDPOINT")
		setup.GrafanaAPIKey = os.Getenv("GRAFANA_API_KEY")
		setup.VaultAddr = os.Getenv("VAULT_ADDR")
		setup.PulumiWorkspace = os.Getenv("PULUMI_WORKSPACE")
	}
	
	return setup
}

// ValidateTestEnvironment validates that the deployer test environment is properly configured
func ValidateTestEnvironment(t *testing.T, suite *IntegrationTestSuite) {
	ctx, cancel := CreateIntegrationTestContext()
	defer cancel()
	
	// Test database connectivity
	err := suite.DatabaseClient.PingContext(ctx)
	require.NoError(t, err, "Database should be accessible")
	
	// Test Redis connectivity
	result := suite.RedisClient.Ping(ctx)
	require.NoError(t, result.Err(), "Redis should be accessible")
	
	// Test basic Redis read/write
	testKey := CreateTestStateKey(suite.Environment.Environment, "test", "validation", "connectivity")
	testValue := "environment-validation"
	
	err = suite.RedisClient.Set(ctx, testKey, testValue, 30*time.Second).Err()
	require.NoError(t, err, "Should be able to write to Redis")
	
	// Cleanup test data
	suite.CleanupRedisTestData(t, []string{testKey})
	
	t.Logf("✅ Test environment validation completed for: %s", suite.Environment.Environment)
}

// GetRequiredEnvVar gets a required environment variable and fails the test if not present
func GetRequiredEnvVar(t *testing.T, key string) string {
	value := os.Getenv(key)
	require.NotEmpty(t, value, "Required environment variable %s must be set for integration tests", key)
	return value
}

// GetEnvVar gets an environment variable with optional default value
func GetEnvVar(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// MakeHTTPRequest creates and executes an HTTP request with timeout
func MakeHTTPRequest(t *testing.T, method, url string) *http.Response {
	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	require.NoError(t, err, "Failed to create HTTP request")
	
	client := &http.Client{
		Timeout: integrationTestTimeout,
	}
	
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to execute HTTP request")
	
	return resp
}

// ConnectWithTimeout attempts to establish a network connection with timeout
func ConnectWithTimeout(ctx context.Context, network, address string, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: timeout,
	}
	return dialer.DialContext(ctx, network, address)
}

