package infrastructure

import (
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/public-website/migrations/config"
	"github.com/axiom-software-co/international-center/src/public-website/migrations/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationPackageIntegrationContract(t *testing.T) {
	t.Run("deployment can import migrations package", func(t *testing.T) {
		// This test validates that the deployment package can successfully import
		// and use the migrations package as an external dependency
		
		// Test that we can create migration configurations
		configs := config.NewDomainMigrationConfigs("/tmp/test", "postgres://localhost:5432/test")
		require.NotNil(t, configs, "Should be able to create migration configs from external package")
		
		// Test that we can create migration runners
		testConfig := config.MigrationConfig{
			Domain:        "test-domain",
			MigrationPath: "/tmp/test-migrations",
			DatabaseURL:   "postgres://localhost:5432/test",
		}
		
		migrationRunner := runner.NewMigrationRunner(testConfig)
		require.NotNil(t, migrationRunner, "Should be able to create migration runner from external package")
		
		// Test that we can create orchestrator
		orchestrator := runner.NewDomainMigrationOrchestrator(configs)
		require.NotNil(t, orchestrator, "Should be able to create migration orchestrator from external package")
	})
}

func TestMigrationComponentContractIntegration(t *testing.T) {
	t.Run("migration component can use external migration package", func(t *testing.T) {
		// Test the contract that migration component depends on from the migrations package
		
		// Test migration strategy mapping
		testCases := []struct {
			environment string
			expected    MigrationStrategy
		}{
			{"development", MigrationStrategyAggressive},
			{"staging", MigrationStrategyCareful},
			{"production", MigrationStrategyConservativeWithExtensiveValidation},
			{"unknown", MigrationStrategyCareful}, // Default fallback
		}
		
		for _, tc := range testCases {
			strategy := GetMigrationStrategy(tc.environment)
			assert.Equal(t, tc.expected, strategy, "Should get correct strategy for environment: %s", tc.environment)
		}
		
		// Test default migration args creation
		args := DefaultMigrationArgs("development", nil)
		require.NotNil(t, args, "Should be able to create default migration args")
		assert.Equal(t, "development", args.Environment)
		assert.Equal(t, MigrationStrategyAggressive, args.MigrationStrategy)
	})
}

func TestMigrationExecutionContract(t *testing.T) {
	t.Run("migration execution follows expected contract", func(t *testing.T) {
		// Test the contract for migration execution that deployment relies on
		
		testConnectionString := "postgres://localhost:5432/test"
		testMigrationsPath := "/tmp/test-migrations"
		testEnvironment := "development"
		testStrategy := MigrationStrategyAggressive
		
		// This should not panic and should return a result map
		// (it will likely fail with connection error, but the interface should work)
		result, err := executeMigrations(testConnectionString, testMigrationsPath, testEnvironment, testStrategy)
		
		// We expect this to fail with connection issues, but not panic
		if err != nil {
			// Error is expected due to no actual database connection in tests
			assert.Contains(t, err.Error(), "migration", "Error should be migration-related")
		} else {
			// If somehow it doesn't error, result should have expected structure
			require.NotNil(t, result, "Result should not be nil")
			
			// Result should have expected keys
			_, hasStatus := result["status"]
			_, hasVersion := result["schema_version"]
			_, hasApplied := result["migrations_applied"]
			_, hasValidation := result["validation_status"]
			
			assert.True(t, hasStatus, "Result should have status")
			assert.True(t, hasVersion, "Result should have schema_version")
			assert.True(t, hasApplied, "Result should have migrations_applied")
			assert.True(t, hasValidation, "Result should have validation_status")
		}
	})
}

func TestMigrationComponentTimeout(t *testing.T) {
	t.Run("migration operations complete within timeout", func(t *testing.T) {
		// Test that migration operations complete within the unit test timeout requirement
		
		done := make(chan bool, 1)
		timeout := time.After(5 * time.Second)
		
		go func() {
			// Test that all migration-related operations complete quickly
			
			// Test config creation
			configs := config.NewDomainMigrationConfigs("/tmp", "postgres://localhost:5432/test")
			_ = configs
			
			// Test runner creation
			testConfig := config.MigrationConfig{
				Domain:        "test-domain",
				MigrationPath: "/tmp/test-migrations",
				DatabaseURL:   "postgres://localhost:5432/test",
			}
			migrationRunner := runner.NewMigrationRunner(testConfig)
			_ = migrationRunner
			
			// Test orchestrator creation  
			orchestrator := runner.NewDomainMigrationOrchestrator(configs)
			_ = orchestrator
			
			// Test migration strategy functions
			_ = GetMigrationStrategy("development")
			_ = DefaultMigrationArgs("development", nil)
			
			done <- true
		}()
		
		select {
		case <-done:
			// Test passed - operations completed within timeout
		case <-timeout:
			t.Fatal("Migration package operations took longer than 5 seconds")
		}
	})
}

func TestMigrationEnvironmentStrategies(t *testing.T) {
	t.Run("environment strategies maintain expected contracts", func(t *testing.T) {
		// Test that environment-specific strategies are available and consistent
		
		// Test strategy mapping consistency
		envStrategies := map[string]MigrationStrategy{
			"development": MigrationStrategyAggressive,
			"staging":     MigrationStrategyCareful,
			"production":  MigrationStrategyConservativeWithExtensiveValidation,
		}
		
		for env, expectedStrategy := range envStrategies {
			strategy := GetMigrationStrategy(env)
			assert.Equal(t, expectedStrategy, strategy, "Strategy should be consistent for environment: %s", env)
			
			// Test that default args use the same strategy
			args := DefaultMigrationArgs(env, nil)
			assert.Equal(t, expectedStrategy, args.MigrationStrategy, "Default args should use same strategy for environment: %s", env)
		}
		
		// Test strategy descriptions exist
		descriptions := []MigrationStrategy{
			MigrationStrategyAggressive,
			MigrationStrategyCareful,
			MigrationStrategyConservative,
			MigrationStrategyConservativeWithExtensiveValidation,
		}
		
		for _, strategy := range descriptions {
			desc := GetMigrationStrategyDescription(strategy)
			assert.NotEmpty(t, desc, "Strategy should have non-empty description: %s", strategy)
			assert.NotEqual(t, "Unknown migration strategy", desc, "Strategy should have valid description: %s", strategy)
		}
	})
}