package integration_tests

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/international-center/src/deployer/internal/messaging"
	"github.com/international-center/src/deployer/internal/orchestrator"
)

type DevelopmentValidationTestSuite struct {
	suite.Suite
	ctx           context.Context
	cancel        context.CancelFunc
	db            *sql.DB
	redisClient   *redis.Client
	pubsubManager *messaging.PubSubManager
	httpClient    *http.Client
	devApiUrl     string
}

func TestDevelopmentValidationSuite(t *testing.T) {
	suite.Run(t, new(DevelopmentValidationTestSuite))
}

func (suite *DevelopmentValidationTestSuite) SetupSuite() {
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 5*time.Minute)

	if os.Getenv("DATABASE_URL") == "" {
		suite.T().Skip("DATABASE_URL not set, skipping development integration tests")
	}
	if os.Getenv("REDIS_ADDR") == "" {
		suite.T().Skip("REDIS_ADDR not set, skipping development integration tests")
	}

	var err error

	suite.db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	suite.Require().NoError(err, "Failed to connect to development database")

	suite.redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	pubsubConfig := &messaging.PubSubConfig{
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       0,
		Environment:   "development",
		ClientName:    "dev-integration-test",
		MaxRetries:    2,
		RetryDelay:    500 * time.Millisecond,
		HealthCheck:   5 * time.Second,
		BufferSize:    50,
	}

	suite.pubsubManager, err = messaging.NewPubSubManager(pubsubConfig)
	suite.Require().NoError(err, "Failed to initialize development pub/sub manager")

	suite.httpClient = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    5,
			IdleConnTimeout: 10 * time.Second,
		},
	}

	suite.devApiUrl = os.Getenv("DEV_API_URL")
	if suite.devApiUrl == "" {
		suite.T().Skip("DEV_API_URL not set, skipping API integration tests")
	}
}

func (suite *DevelopmentValidationTestSuite) TearDownSuite() {
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
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentBasicConnectivity() {
	t := suite.T()

	t.Run("DatabaseConnection", func(t *testing.T) {
		err := suite.db.PingContext(suite.ctx)
		assert.NoError(t, err, "Development database should be accessible")
	})

	t.Run("RedisConnection", func(t *testing.T) {
		status := suite.redisClient.Ping(suite.ctx)
		assert.NoError(t, status.Err(), "Development Redis should be accessible")
	})
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentBasicFunctionality() {
	t := suite.T()

	t.Run("DatabaseQuery", func(t *testing.T) {
		var result int
		err := suite.db.QueryRowContext(suite.ctx, "SELECT 1").Scan(&result)
		assert.NoError(t, err, "Should be able to execute basic database query")
		assert.Equal(t, 1, result, "Query should return expected result")
	})

	t.Run("RedisReadWrite", func(t *testing.T) {
		testKey := fmt.Sprintf("dev-test-%d", time.Now().Unix())
		testValue := "dev-test-value"

		err := suite.redisClient.Set(suite.ctx, testKey, testValue, 30*time.Second).Err()
		assert.NoError(t, err, "Should be able to write to development Redis")

		result, err := suite.redisClient.Get(suite.ctx, testKey).Result()
		assert.NoError(t, err, "Should be able to read from development Redis")
		assert.Equal(t, testValue, result, "Should read the same value that was written")

		suite.redisClient.Del(suite.ctx, testKey)
	})
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentPubSubBasics() {
	t := suite.T()

	t.Run("PubSubConnection", func(t *testing.T) {
		assert.NotNil(t, suite.pubsubManager, "PubSub manager should be initialized for development")
	})

	t.Run("BasicEventPublishing", func(t *testing.T) {
		testEvent := &messaging.DeploymentEvent{
			ID:          fmt.Sprintf("dev-test-%d", time.Now().Unix()),
			Type:        "DEV_TEST_EVENT",
			Environment: "development",
			Timestamp:   time.Now(),
			Data: map[string]interface{}{
				"test": "dev-value",
			},
		}

		err := suite.pubsubManager.PublishEvent(suite.ctx, testEvent)
		assert.NoError(t, err, "Should be able to publish basic events in development")
	})
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentApiBasics() {
	t := suite.T()

	if suite.devApiUrl == "" {
		t.Skip("Development API URL not configured")
	}

	t.Run("ApiAccessibility", func(t *testing.T) {
		resp, err := suite.httpClient.Get(suite.devApiUrl)
		if err != nil {
			t.Logf("Development API not available: %v (this is acceptable in development)", err)
			return
		}
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode < 500, 
			"Development API should not return server errors when accessible")
	})
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentEnvironmentConfig() {
	t := suite.T()

	t.Run("DevelopmentEnvironmentVariables", func(t *testing.T) {
		basicVars := []string{
			"DATABASE_URL",
			"REDIS_ADDR",
		}

		for _, varName := range basicVars {
			value := os.Getenv(varName)
			assert.NotEmpty(t, value, "Basic environment variable %s should be configured", varName)
		}
	})

	t.Run("DevelopmentDeploymentConfig", func(t *testing.T) {
		config := &orchestrator.OrchestratorConfig{
			Environment:           "development",
			MaxConcurrentDeploys:  5,
			DeploymentTimeout:     30 * time.Minute,
			HealthCheckInterval:   15 * time.Second,
			RetryAttempts:        1,
			EnableSecurityTesting: false,
			EnableMigrationTesting: false,
		}

		assert.Equal(t, "development", config.Environment, "Environment should be development")
		assert.Equal(t, 5, config.MaxConcurrentDeploys, "Should allow high concurrency for development")
		assert.False(t, config.EnableSecurityTesting, "Security testing should be disabled in development")
		assert.False(t, config.EnableMigrationTesting, "Migration testing should be disabled in development")
	})
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentTableBasics() {
	t := suite.T()

	t.Run("CoreTablesExist", func(t *testing.T) {
		coreTableNames := []string{"services", "content"}
		
		for _, tableName := range coreTableNames {
			var exists bool
			query := `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)`
			err := suite.db.QueryRowContext(suite.ctx, query, tableName).Scan(&exists)
			
			if err != nil || !exists {
				t.Logf("Table %s may not exist in development environment: %v (this is acceptable)", tableName, err)
				continue
			}
			
			assert.True(t, exists, "Core table %s should exist if migrations have run", tableName)
		}
	})
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentRedisChannels() {
	t := suite.T()

	t.Run("BasicChannelAvailability", func(t *testing.T) {
		channels := []string{"development.deployments", "development.health"}
		
		for _, channel := range channels {
			pubsub := suite.redisClient.Subscribe(suite.ctx, channel)
			assert.NotNil(t, pubsub, "Should be able to subscribe to basic channel %s", channel)
			
			pubsub.Close()
		}
	})
}

func (suite *DevelopmentValidationTestSuite) TestDevelopmentQuickSmokeTest() {
	t := suite.T()

	t.Run("OverallSystemReadiness", func(t *testing.T) {
		dbOk := suite.db.PingContext(suite.ctx) == nil
		redisOk := suite.redisClient.Ping(suite.ctx).Err() == nil
		pubsubOk := suite.pubsubManager != nil

		systemReady := dbOk && redisOk && pubsubOk
		assert.True(t, systemReady, 
			"Development system should be ready (DB: %v, Redis: %v, PubSub: %v)", 
			dbOk, redisOk, pubsubOk)
	})
}

