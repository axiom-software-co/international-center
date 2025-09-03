package integration_tests

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/international-center/src/deployer/internal/messaging"
	"github.com/international-center/src/deployer/internal/orchestrator"
	"github.com/international-center/src/deployer/internal/production/application"
	"github.com/international-center/src/deployer/internal/production/infrastructure"
	"github.com/international-center/src/deployer/internal/production/migration"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	_ "github.com/lib/pq"
)

type ProductionValidationTestSuite struct {
	suite.Suite
	ctx                   context.Context
	cancel                context.CancelFunc
	db                    *sql.DB
	redisClient           *redis.Client
	pubsubManager         *messaging.PubSubManager
	orchestrator          *orchestrator.DeployerOrchestrator
	infrastructureManager *infrastructure.AzureProductionAppsStack
	applicationManager    *application.ProductionApplicationDeployer
	migrationManager      *migration.ProductionMigrationOrchestrator
	httpClient            *http.Client
	testTimeout           time.Duration
}

func (suite *ProductionValidationTestSuite) SetupSuite() {
	suite.testTimeout = 15 * time.Second
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), suite.testTimeout)

	suite.requireEnvironmentVariables()
	suite.setupDatabaseConnection()
	suite.setupRedisConnection()
	suite.setupHTTPClient()
	suite.setupPubSubManager()
	suite.setupOrchestrator()
	suite.setupInfrastructureManager()
	suite.setupApplicationManager()
	suite.setupMigrationManager()
}

func (suite *ProductionValidationTestSuite) TearDownSuite() {
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

func (suite *ProductionValidationTestSuite) requireEnvironmentVariables() {
	requiredVars := []string{
		"DATABASE_URL",
		"REDIS_ADDR",
		"REDIS_PASSWORD",
		"AZURE_SUBSCRIPTION_ID",
		"AZURE_CLIENT_ID",
		"AZURE_CLIENT_SECRET",
		"AZURE_TENANT_ID",
		"PRODUCTION_API_URL",
		"PRODUCTION_ADMIN_URL",
		"GRAFANA_ENDPOINT",
		"GRAFANA_API_KEY",
	}

	for _, envVar := range requiredVars {
		value := os.Getenv(envVar)
		require.NotEmpty(suite.T(), value, fmt.Sprintf("Environment variable %s must be set for production integration tests", envVar))
	}
}

func (suite *ProductionValidationTestSuite) setupDatabaseConnection() {
	var err error
	suite.db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	require.NoError(suite.T(), err, "Failed to connect to production database")

	ctx, cancel := context.WithTimeout(suite.ctx, 5*time.Second)
	defer cancel()

	err = suite.db.PingContext(ctx)
	require.NoError(suite.T(), err, "Failed to ping production database")
}

func (suite *ProductionValidationTestSuite) setupRedisConnection() {
	suite.redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(suite.ctx, 5*time.Second)
	defer cancel()

	err := suite.redisClient.Ping(ctx).Err()
	require.NoError(suite.T(), err, "Failed to connect to production Redis")
}

func (suite *ProductionValidationTestSuite) setupHTTPClient() {
	suite.httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
}

func (suite *ProductionValidationTestSuite) setupPubSubManager() {
	config := &messaging.PubSubConfig{
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       0,
		Environment:   "production",
		ClientName:    "production-integration-test",
		MaxRetries:    3,
		RetryDelay:    1 * time.Second,
		HealthCheck:   10 * time.Second,
		BufferSize:    100,
	}

	var err error
	suite.pubsubManager, err = messaging.NewPubSubManager(config)
	require.NoError(suite.T(), err, "Failed to initialize production pub/sub manager")
}

func (suite *ProductionValidationTestSuite) setupOrchestrator() {
	config := &orchestrator.OrchestratorConfig{
		Environment:            "production",
		MaxConcurrentDeploys:   1,
		DeploymentTimeout:      30 * time.Minute,
		HealthCheckInterval:    15 * time.Second,
		RetryAttempts:         3,
		EnableSecurityTesting: true,
		EnableMigrationTesting: true,
		NotificationChannels:  []string{"test"},
	}

	pubsubConfig := &messaging.PubSubConfig{
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       0,
		Environment:   "production",
		ClientName:    "production-orchestrator-test",
		MaxRetries:    3,
		RetryDelay:    1 * time.Second,
		HealthCheck:   10 * time.Second,
		BufferSize:    100,
	}

	var err error
	suite.orchestrator, err = orchestrator.NewDeployerOrchestrator(config, pubsubConfig)
	require.NoError(suite.T(), err, "Failed to initialize production orchestrator")
}

func (suite *ProductionValidationTestSuite) setupInfrastructureManager() {
	var err error
	suite.infrastructureManager, err = infrastructure.NewAzureProductionAppsStack()
	require.NoError(suite.T(), err, "Failed to initialize production infrastructure manager")
}

func (suite *ProductionValidationTestSuite) setupApplicationManager() {
	var err error
	suite.applicationManager, err = application.NewProductionApplicationDeployer()
	require.NoError(suite.T(), err, "Failed to initialize production application manager")
}

func (suite *ProductionValidationTestSuite) setupMigrationManager() {
	var err error
	suite.migrationManager, err = migration.NewProductionMigrationOrchestrator()
	require.NoError(suite.T(), err, "Failed to initialize production migration manager")
}

func (suite *ProductionValidationTestSuite) TestProductionDatabaseSchemaValidation() {
	t := suite.T()

	t.Run("ServicesTableExists", func(t *testing.T) {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = 'services'
			)
		`
		err := suite.db.QueryRowContext(suite.ctx, query).Scan(&exists)
		require.NoError(t, err, "Failed to check services table existence")
		assert.True(t, exists, "Services table must exist in production database")
	})

	t.Run("ServiceCategoriesTableExists", func(t *testing.T) {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = 'service_categories'
			)
		`
		err := suite.db.QueryRowContext(suite.ctx, query).Scan(&exists)
		require.NoError(t, err, "Failed to check service_categories table existence")
		assert.True(t, exists, "Service categories table must exist in production database")
	})

	t.Run("ContentTableExists", func(t *testing.T) {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = 'content'
			)
		`
		err := suite.db.QueryRowContext(suite.ctx, query).Scan(&exists)
		require.NoError(t, err, "Failed to check content table existence")
		assert.True(t, exists, "Content table must exist in production database")
	})

	t.Run("RequiredIndexesExist", func(t *testing.T) {
		requiredIndexes := []string{
			"idx_services_category_id",
			"idx_services_publishing_status",
			"idx_services_slug",
			"idx_content_hash",
			"idx_content_upload_status",
		}

		for _, indexName := range requiredIndexes {
			var exists bool
			query := `
				SELECT EXISTS (
					SELECT FROM pg_indexes 
					WHERE indexname = $1
				)
			`
			err := suite.db.QueryRowContext(suite.ctx, query, indexName).Scan(&exists)
			require.NoError(t, err, fmt.Sprintf("Failed to check index %s existence", indexName))
			assert.True(t, exists, fmt.Sprintf("Index %s must exist in production database", indexName))
		}
	})
}

func (suite *ProductionValidationTestSuite) TestProductionRedisConnectivity() {
	t := suite.T()

	t.Run("RedisHealthCheck", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(suite.ctx, 5*time.Second)
		defer cancel()

		result := suite.redisClient.Ping(ctx)
		assert.NoError(t, result.Err(), "Production Redis must be healthy")
		assert.Equal(t, "PONG", result.Val(), "Redis ping must return PONG")
	})

	t.Run("RedisPubSubCapability", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(suite.ctx, 10*time.Second)
		defer cancel()

		testChannel := "production-integration-test"
		testMessage := "test-message-" + fmt.Sprintf("%d", time.Now().UnixNano())

		pubsub := suite.redisClient.Subscribe(ctx, testChannel)
		defer pubsub.Close()

		_, err := pubsub.Receive(ctx)
		require.NoError(t, err, "Failed to subscribe to test channel")

		err = suite.redisClient.Publish(ctx, testChannel, testMessage).Err()
		require.NoError(t, err, "Failed to publish test message")

		msg, err := pubsub.ReceiveMessage(ctx)
		require.NoError(t, err, "Failed to receive test message")
		assert.Equal(t, testMessage, msg.Payload, "Published and received messages must match")
	})

	t.Run("RedisPerformance", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(suite.ctx, 10*time.Second)
		defer cancel()

		start := time.Now()
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("perf-test-%d", i)
			err := suite.redisClient.Set(ctx, key, fmt.Sprintf("value-%d", i), time.Minute).Err()
			require.NoError(t, err, "Redis SET operation failed")

			val, err := suite.redisClient.Get(ctx, key).Result()
			require.NoError(t, err, "Redis GET operation failed")
			assert.Equal(t, fmt.Sprintf("value-%d", i), val, "Retrieved value must match set value")
		}
		duration := time.Since(start)

		assert.Less(t, duration, 5*time.Second, "Redis performance must be acceptable for production")

		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("perf-test-%d", i)
			suite.redisClient.Del(ctx, key)
		}
	})
}

func (suite *ProductionValidationTestSuite) TestProductionAPIEndpointsAvailability() {
	t := suite.T()

	apiURL := os.Getenv("PRODUCTION_API_URL")
	adminURL := os.Getenv("PRODUCTION_ADMIN_URL")

	t.Run("APIHealthEndpoint", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(suite.ctx, 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", apiURL+"/health", nil)
		require.NoError(t, err, "Failed to create health check request")

		resp, err := suite.httpClient.Do(req)
		require.NoError(t, err, "Health check request failed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Health endpoint must return 200 OK")
	})

	t.Run("APIServicesEndpoint", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(suite.ctx, 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", apiURL+"/api/v1/services", nil)
		require.NoError(t, err, "Failed to create services request")

		resp, err := suite.httpClient.Do(req)
		require.NoError(t, err, "Services endpoint request failed")
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized,
			"Services endpoint must return 200 OK or 401 Unauthorized (if auth required)")
	})

	t.Run("AdminHealthEndpoint", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(suite.ctx, 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", adminURL+"/health", nil)
		require.NoError(t, err, "Failed to create admin health check request")

		resp, err := suite.httpClient.Do(req)
		require.NoError(t, err, "Admin health check request failed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Admin health endpoint must return 200 OK")
	})

	t.Run("AdminServicesEndpoint", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(suite.ctx, 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", adminURL+"/admin/api/v1/services", nil)
		require.NoError(t, err, "Failed to create admin services request")

		resp, err := suite.httpClient.Do(req)
		require.NoError(t, err, "Admin services endpoint request failed")
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized,
			"Admin services endpoint must return 200 OK or 401 Unauthorized")
	})
}

func (suite *ProductionValidationTestSuite) TestProductionSecurityHeaders() {
	t := suite.T()

	apiURL := os.Getenv("PRODUCTION_API_URL")

	t.Run("SecurityHeadersPresent", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(suite.ctx, 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", apiURL+"/health", nil)
		require.NoError(t, err, "Failed to create request")

		resp, err := suite.httpClient.Do(req)
		require.NoError(t, err, "Request failed")
		defer resp.Body.Close()

		requiredHeaders := map[string]string{
			"X-Content-Type-Options":   "nosniff",
			"X-Frame-Options":          "DENY",
			"X-XSS-Protection":         "1; mode=block",
			"Strict-Transport-Security": "",
			"Content-Security-Policy":   "",
			"Referrer-Policy":          "",
		}

		for headerName, expectedValue := range requiredHeaders {
			actualValue := resp.Header.Get(headerName)
			assert.NotEmpty(t, actualValue, fmt.Sprintf("Security header %s must be present", headerName))
			
			if expectedValue != "" {
				assert.Equal(t, expectedValue, actualValue, 
					fmt.Sprintf("Security header %s must have correct value", headerName))
			}
		}
	})

	t.Run("HTTPSEnforcement", func(t *testing.T) {
		assert.Contains(t, apiURL, "https://", "Production API must use HTTPS")

		adminURL := os.Getenv("PRODUCTION_ADMIN_URL")
		assert.Contains(t, adminURL, "https://", "Production Admin must use HTTPS")
	})
}

func (suite *ProductionValidationTestSuite) TestProductionMonitoringIntegration() {
	t := suite.T()

	grafanaEndpoint := os.Getenv("GRAFANA_ENDPOINT")
	grafanaAPIKey := os.Getenv("GRAFANA_API_KEY")

	t.Run("GrafanaConnectivity", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(suite.ctx, 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", grafanaEndpoint+"/api/health", nil)
		require.NoError(t, err, "Failed to create Grafana request")
		req.Header.Set("Authorization", "Bearer "+grafanaAPIKey)

		resp, err := suite.httpClient.Do(req)
		require.NoError(t, err, "Grafana health check failed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Grafana must be accessible")
	})

	t.Run("GrafanaDashboardsExist", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(suite.ctx, 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", grafanaEndpoint+"/api/search?type=dash-db", nil)
		require.NoError(t, err, "Failed to create Grafana dashboards request")
		req.Header.Set("Authorization", "Bearer "+grafanaAPIKey)

		resp, err := suite.httpClient.Do(req)
		require.NoError(t, err, "Grafana dashboards request failed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Must be able to access Grafana dashboards")
	})
}

func (suite *ProductionValidationTestSuite) TestProductionPubSubIntegration() {
	t := suite.T()

	t.Run("PubSubManagerHealthy", func(t *testing.T) {
		metrics := suite.pubsubManager.GetMetrics()
		assert.NotNil(t, metrics, "PubSub metrics must be available")
	})

	t.Run("PublishAndReceiveMessage", func(t *testing.T) {
		testChannel := messaging.InfrastructureChanged
		received := make(chan bool, 1)

		suite.pubsubManager.Subscribe(testChannel, func(ctx context.Context, msg *messaging.Message) error {
			assert.Equal(t, "production", msg.Environment)
			assert.Equal(t, "integration-test", msg.Source)
			received <- true
			return nil
		})

		go func() {
			ctx, cancel := context.WithTimeout(suite.ctx, 5*time.Second)
			defer cancel()
			suite.pubsubManager.StartListening(ctx)
		}()

		time.Sleep(1 * time.Second)

		ctx, cancel := context.WithTimeout(suite.ctx, 5*time.Second)
		defer cancel()

		err := suite.pubsubManager.PublishInfrastructureEvent(ctx, testChannel, "production", "test-component", map[string]interface{}{
			"test": true,
		})
		require.NoError(t, err, "Failed to publish test message")

		select {
		case <-received:
			assert.True(t, true, "Message received successfully")
		case <-time.After(10 * time.Second):
			assert.Fail(t, "Test message was not received within timeout")
		}
	})
}

func (suite *ProductionValidationTestSuite) TestProductionOrchestratorReadiness() {
	t := suite.T()

	t.Run("OrchestratorInitialized", func(t *testing.T) {
		assert.NotNil(t, suite.orchestrator, "Production orchestrator must be initialized")
	})

	t.Run("ActiveDeploymentsTracking", func(t *testing.T) {
		activeDeployments := suite.orchestrator.ListActiveDeployments()
		assert.NotNil(t, activeDeployments, "Active deployments list must be available")
	})
}

func (suite *ProductionValidationTestSuite) TestProductionComponentsReadiness() {
	t := suite.T()

	t.Run("InfrastructureManagerReady", func(t *testing.T) {
		assert.NotNil(t, suite.infrastructureManager, "Production infrastructure manager must be ready")
	})

	t.Run("ApplicationManagerReady", func(t *testing.T) {
		assert.NotNil(t, suite.applicationManager, "Production application manager must be ready")
	})

	t.Run("MigrationManagerReady", func(t *testing.T) {
		assert.NotNil(t, suite.migrationManager, "Production migration manager must be ready")
	})
}

func TestProductionValidationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping production integration tests in short mode")
	}

	if os.Getenv("INTEGRATION_TESTS") != "production" {
		t.Skip("Production integration tests require INTEGRATION_TESTS=production")
	}

	suite.Run(t, new(ProductionValidationTestSuite))
}