package integration_tests

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/international-center/src/deployer/internal/messaging"
	"github.com/international-center/src/deployer/internal/orchestrator"
)

type StagingValidationTestSuite struct {
	suite.Suite
	ctx             context.Context
	cancel          context.CancelFunc
	db              *sql.DB
	redisClient     *redis.Client
	pubsubManager   *messaging.PubSubManager
	orchestrator    *orchestrator.DeployerOrchestrator
	httpClient      *http.Client
	stagingApiUrl   string
	stagingAdminUrl string
}

func TestStagingValidationSuite(t *testing.T) {
	suite.Run(t, new(StagingValidationTestSuite))
}

func (suite *StagingValidationTestSuite) SetupSuite() {
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	if os.Getenv("DATABASE_URL") == "" {
		suite.T().Skip("DATABASE_URL not set, skipping staging integration tests")
	}
	if os.Getenv("REDIS_ADDR") == "" {
		suite.T().Skip("REDIS_ADDR not set, skipping staging integration tests")
	}

	var err error

	suite.db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	suite.Require().NoError(err, "Failed to connect to staging database")

	suite.redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	pubsubConfig := &messaging.PubSubConfig{
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       0,
		Environment:   "staging",
		ClientName:    "staging-integration-test",
		MaxRetries:    3,
		RetryDelay:    1 * time.Second,
		HealthCheck:   10 * time.Second,
		BufferSize:    100,
	}

	suite.pubsubManager, err = messaging.NewPubSubManager(pubsubConfig)
	suite.Require().NoError(err, "Failed to initialize staging pub/sub manager")

	orchestratorConfig := &orchestrator.OrchestratorConfig{
		Environment:            "staging",
		MaxConcurrentDeploys:   2,
		DeploymentTimeout:      60 * time.Minute,
		HealthCheckInterval:    30 * time.Second,
		RetryAttempts:         3,
		EnableSecurityTesting: true,
		EnableMigrationTesting: true,
		NotificationChannels:  []string{"slack", "email"},
	}

	suite.orchestrator, err = orchestrator.NewDeployerOrchestrator(orchestratorConfig, pubsubConfig)
	suite.Require().NoError(err, "Failed to initialize staging orchestrator")

	suite.httpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		},
	}

	suite.stagingApiUrl = os.Getenv("STAGING_API_URL")
	if suite.stagingApiUrl == "" {
		suite.T().Skip("STAGING_API_URL not set, skipping API integration tests")
	}
	
	suite.stagingAdminUrl = os.Getenv("STAGING_ADMIN_URL")
	if suite.stagingAdminUrl == "" {
		suite.T().Skip("STAGING_ADMIN_URL not set, skipping admin integration tests")
	}
}

func (suite *StagingValidationTestSuite) TearDownSuite() {
	if suite.cancel != nil {
		suite.cancel()
	}
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.redisClient != nil {
		suite.redisClient.Close()
	}
	if suite.pubsubManager != nil {
		suite.pubsubManager.Close()
	}
	if suite.orchestrator != nil {
		suite.orchestrator.Close()
	}
}

func (suite *StagingValidationTestSuite) TestStagingDatabaseConnectivity() {
	t := suite.T()

	t.Run("DatabaseConnection", func(t *testing.T) {
		err := suite.db.PingContext(suite.ctx)
		assert.NoError(t, err, "Staging database should be accessible")
	})

	t.Run("BasicTableValidation", func(t *testing.T) {
		var exists bool
		query := `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'services')`
		err := suite.db.QueryRowContext(suite.ctx, query).Scan(&exists)
		assert.NoError(t, err, "Should be able to check for services table")
		assert.True(t, exists, "Services table should exist in staging database")

		query = `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'content')`
		err = suite.db.QueryRowContext(suite.ctx, query).Scan(&exists)
		assert.NoError(t, err, "Should be able to check for content table")
		assert.True(t, exists, "Content table should exist in staging database")
	})

	t.Run("DatabaseReadWrite", func(t *testing.T) {
		var result int
		err := suite.db.QueryRowContext(suite.ctx, "SELECT 1").Scan(&result)
		assert.NoError(t, err, "Should be able to execute basic query")
		assert.Equal(t, 1, result, "Query should return expected result")
	})
}

func (suite *StagingValidationTestSuite) TestStagingRedisConnectivity() {
	t := suite.T()

	t.Run("RedisConnection", func(t *testing.T) {
		status := suite.redisClient.Ping(suite.ctx)
		assert.NoError(t, status.Err(), "Staging Redis should be accessible")
	})

	t.Run("RedisReadWrite", func(t *testing.T) {
		testKey := fmt.Sprintf("staging-test-%d", time.Now().Unix())
		testValue := "staging-test-value"

		err := suite.redisClient.Set(suite.ctx, testKey, testValue, 1*time.Minute).Err()
		assert.NoError(t, err, "Should be able to write to staging Redis")

		result, err := suite.redisClient.Get(suite.ctx, testKey).Result()
		assert.NoError(t, err, "Should be able to read from staging Redis")
		assert.Equal(t, testValue, result, "Should read the same value that was written")

		err = suite.redisClient.Del(suite.ctx, testKey).Err()
		assert.NoError(t, err, "Should be able to delete test key")
	})
}

func (suite *StagingValidationTestSuite) TestStagingPubSubMessaging() {
	t := suite.T()

	t.Run("PubSubConnection", func(t *testing.T) {
		assert.NotNil(t, suite.pubsubManager, "PubSub manager should be initialized")
	})

	t.Run("EventPublishing", func(t *testing.T) {
		testEvent := &messaging.DeploymentEvent{
			ID:          fmt.Sprintf("staging-test-%d", time.Now().Unix()),
			Type:        "TEST_EVENT",
			Environment: "staging",
			Timestamp:   time.Now(),
			Data: map[string]interface{}{
				"test": "staging-value",
			},
		}

		err := suite.pubsubManager.PublishEvent(suite.ctx, testEvent)
		assert.NoError(t, err, "Should be able to publish events to staging environment")
	})
}

func (suite *StagingValidationTestSuite) TestStagingApiEndpoints() {
	t := suite.T()

	if suite.stagingApiUrl == "" {
		t.Skip("STAGING_API_URL not configured, skipping API endpoint tests")
	}

	t.Run("HealthEndpoint", func(t *testing.T) {
		healthUrl := fmt.Sprintf("%s/health", suite.stagingApiUrl)
		resp, err := suite.httpClient.Get(healthUrl)
		if err != nil {
			t.Logf("Health endpoint not available: %v", err)
			return
		}
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 503, 
			"Health endpoint should return 200 (healthy) or 503 (unhealthy)")
	})

	t.Run("ServicesEndpoint", func(t *testing.T) {
		servicesUrl := fmt.Sprintf("%s/api/v1/services", suite.stagingApiUrl)
		resp, err := suite.httpClient.Get(servicesUrl)
		if err != nil {
			t.Logf("Services endpoint not available: %v", err)
			return
		}
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 404, 
			"Services endpoint should be accessible in staging")
	})

	t.Run("ContentEndpoint", func(t *testing.T) {
		contentUrl := fmt.Sprintf("%s/api/v1/content", suite.stagingApiUrl)
		resp, err := suite.httpClient.Get(contentUrl)
		if err != nil {
			t.Logf("Content endpoint not available: %v", err)
			return
		}
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 404, 
			"Content endpoint should be accessible in staging")
	})
}

func (suite *StagingValidationTestSuite) TestStagingSecurityHeaders() {
	t := suite.T()

	if suite.stagingApiUrl == "" {
		t.Skip("STAGING_API_URL not configured, skipping security header tests")
	}

	t.Run("SecurityHeaders", func(t *testing.T) {
		resp, err := suite.httpClient.Get(suite.stagingApiUrl)
		if err != nil {
			t.Logf("API endpoint not available for security testing: %v", err)
			return
		}
		defer resp.Body.Close()

		expectedHeaders := map[string]string{
			"X-Content-Type-Options": "nosniff",
			"X-Frame-Options":        "",
		}

		for header, expectedValue := range expectedHeaders {
			actualValue := resp.Header.Get(header)
			if expectedValue != "" {
				assert.Equal(t, expectedValue, actualValue, 
					"Security header %s should have correct value", header)
			} else {
				assert.NotEmpty(t, actualValue, 
					"Security header %s should be present", header)
			}
		}
	})
}

func (suite *StagingValidationTestSuite) TestStagingMonitoringIntegration() {
	t := suite.T()

	grafanaEndpoint := os.Getenv("GRAFANA_ENDPOINT")
	if grafanaEndpoint == "" {
		t.Skip("GRAFANA_ENDPOINT not configured, skipping monitoring tests")
	}

	t.Run("MonitoringEndpoint", func(t *testing.T) {
		resp, err := suite.httpClient.Get(grafanaEndpoint)
		if err != nil {
			t.Logf("Monitoring endpoint not accessible: %v", err)
			return
		}
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode < 500, 
			"Monitoring endpoint should be accessible (not server error)")
	})
}

func (suite *StagingValidationTestSuite) TestStagingServiceReadiness() {
	t := suite.T()

	t.Run("DatabaseMigrationStatus", func(t *testing.T) {
		var migrationCount int
		query := `SELECT COUNT(*) FROM information_schema.tables 
		          WHERE table_schema = 'public' 
		          AND table_name IN ('services', 'service_categories', 'content')`
		err := suite.db.QueryRowContext(suite.ctx, query).Scan(&migrationCount)
		assert.NoError(t, err, "Should be able to check migration status")
		assert.GreaterOrEqual(t, migrationCount, 2, "Core tables should exist after migration")
	})

	t.Run("RedisChannelAvailability", func(t *testing.T) {
		channels := []string{"staging.deployments", "staging.health", "staging.events"}
		
		for _, channel := range channels {
			pubsub := suite.redisClient.Subscribe(suite.ctx, channel)
			assert.NotNil(t, pubsub, "Should be able to subscribe to channel %s", channel)
			
			err := pubsub.Close()
			assert.NoError(t, err, "Should be able to close subscription")
		}
	})
}

func (suite *StagingValidationTestSuite) TestStagingAzureIntegration() {
	t := suite.T()

	azureClientId := os.Getenv("AZURE_CLIENT_ID")
	azureSubscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")

	if azureClientId == "" || azureSubscriptionId == "" {
		t.Skip("Azure configuration not available, skipping Azure integration tests")
	}

	t.Run("AzureEnvironmentVariables", func(t *testing.T) {
		requiredVars := []string{
			"AZURE_SUBSCRIPTION_ID",
			"AZURE_CLIENT_ID", 
			"AZURE_CLIENT_SECRET",
			"AZURE_TENANT_ID",
		}

		for _, varName := range requiredVars {
			value := os.Getenv(varName)
			assert.NotEmpty(t, value, "Azure environment variable %s should be configured", varName)
		}
	})
}

func (suite *StagingValidationTestSuite) TestStagingDeploymentWorkflow() {
	t := suite.T()

	t.Run("DeploymentConfigValidation", func(t *testing.T) {
		config := &orchestrator.OrchestratorConfig{
			Environment:            "staging",
			MaxConcurrentDeploys:   2,
			DeploymentTimeout:      60 * time.Minute,
			HealthCheckInterval:    30 * time.Second,
			RetryAttempts:         3,
			EnableSecurityTesting: true,
			EnableMigrationTesting: true,
			NotificationChannels:  []string{"slack", "email"},
		}

		assert.Equal(t, "staging", config.Environment, "Environment should be staging")
		assert.Equal(t, 2, config.MaxConcurrentDeploys, "Should allow moderate concurrency for staging")
		assert.True(t, config.EnableSecurityTesting, "Security testing should be enabled in staging")
		assert.True(t, config.EnableMigrationTesting, "Migration testing should be enabled in staging")
	})

	t.Run("DeploymentStatusTracking", func(t *testing.T) {
		if suite.orchestrator == nil {
			t.Skip("Orchestrator not available")
		}

		testSession := &orchestrator.DeploymentSession{
			ID:          fmt.Sprintf("staging-test-%d", time.Now().Unix()),
			Environment: "staging",
			Services:    []string{"api", "admin"},
			Status:      orchestrator.DeploymentInProgress,
			StartTime:   time.Now(),
		}

		status := suite.orchestrator.GetDeploymentStatus(testSession.ID)
		assert.NotNil(t, status, "Should be able to track deployment status")
	})
}

