package migration

import (
	"testing"

	"github.com/axiom-software-co/international-center/src/cicd/components"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigrationRunner_DevelopmentEnvironment validates aggressive migration strategy
func TestMigrationRunner_DevelopmentEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		
		// Deploy database component first to get connection details
		databaseOutputs, err := components.DeployDatabase(ctx, cfg, "development")
		require.NoError(t, err)
		
		// Initialize migration runner with development strategy
		runner, err := NewMigrationRunner(ctx, cfg, "development", databaseOutputs)
		require.NoError(t, err)
		
		// Verify development strategy configuration
		pulumi.All(runner.Strategy, runner.SafetyChecks, runner.RollbackEnabled, runner.TimeoutMinutes).ApplyT(func(args []interface{}) error {
			strategy := args[0].(string)
			safetyChecks := args[1].(bool)
			rollbackEnabled := args[2].(bool)
			timeoutMinutes := args[3].(int)
			
			assert.Equal(t, "aggressive", strategy, "Development should use aggressive migration strategy")
			assert.False(t, safetyChecks, "Development should disable safety checks for speed")
			assert.False(t, rollbackEnabled, "Development should disable rollback for simplicity")
			assert.Equal(t, 5, timeoutMinutes, "Development should use short timeout")
			return nil
		})
		
		return nil
	}, pulumi.WithMocks("test", "stack", &MigrationMocks{}))
	
	assert.NoError(t, err)
}

// TestMigrationRunner_StagingEnvironment validates careful migration strategy
func TestMigrationRunner_StagingEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		
		// Deploy database component first to get connection details
		databaseOutputs, err := components.DeployDatabase(ctx, cfg, "staging")
		require.NoError(t, err)
		
		// Initialize migration runner with staging strategy
		runner, err := NewMigrationRunner(ctx, cfg, "staging", databaseOutputs)
		require.NoError(t, err)
		
		// Verify staging strategy configuration
		pulumi.All(runner.Strategy, runner.SafetyChecks, runner.RollbackEnabled, runner.TimeoutMinutes, runner.BackupRequired).ApplyT(func(args []interface{}) error {
			strategy := args[0].(string)
			safetyChecks := args[1].(bool)
			rollbackEnabled := args[2].(bool)
			timeoutMinutes := args[3].(int)
			backupRequired := args[4].(bool)
			
			assert.Equal(t, "careful", strategy, "Staging should use careful migration strategy")
			assert.True(t, safetyChecks, "Staging should enable safety checks")
			assert.True(t, rollbackEnabled, "Staging should enable rollback capability")
			assert.Equal(t, 15, timeoutMinutes, "Staging should use moderate timeout")
			assert.True(t, backupRequired, "Staging should require backup before migration")
			return nil
		})
		
		return nil
	}, pulumi.WithMocks("test", "stack", &MigrationMocks{}))
	
	assert.NoError(t, err)
}

// TestMigrationRunner_ProductionEnvironment validates conservative migration strategy
func TestMigrationRunner_ProductionEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		
		// Deploy database component first to get connection details
		databaseOutputs, err := components.DeployDatabase(ctx, cfg, "production")
		require.NoError(t, err)
		
		// Initialize migration runner with production strategy
		runner, err := NewMigrationRunner(ctx, cfg, "production", databaseOutputs)
		require.NoError(t, err)
		
		// Verify production strategy configuration
		pulumi.All(runner.Strategy, runner.SafetyChecks, runner.RollbackEnabled, runner.TimeoutMinutes, runner.BackupRequired, runner.ApprovalRequired).ApplyT(func(args []interface{}) error {
			strategy := args[0].(string)
			safetyChecks := args[1].(bool)
			rollbackEnabled := args[2].(bool)
			timeoutMinutes := args[3].(int)
			backupRequired := args[4].(bool)
			approvalRequired := args[5].(bool)
			
			assert.Equal(t, "conservative", strategy, "Production should use conservative migration strategy")
			assert.True(t, safetyChecks, "Production should enable safety checks")
			assert.True(t, rollbackEnabled, "Production should enable rollback capability")
			assert.Equal(t, 30, timeoutMinutes, "Production should use extended timeout")
			assert.True(t, backupRequired, "Production should require backup before migration")
			assert.True(t, approvalRequired, "Production should require manual approval")
			return nil
		})
		
		return nil
	}, pulumi.WithMocks("test", "stack", &MigrationMocks{}))
	
	assert.NoError(t, err)
}

// TestMigrationRunner_DatabaseSchemaValidation validates migration runner loads correct database schemas
func TestMigrationRunner_DatabaseSchemaValidation(t *testing.T) {
	domains := []string{"business", "donations", "events", "media", "news", "research", "services", "volunteers"}
	
	for _, domain := range domains {
		t.Run("Domain_"+domain, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				cfg := config.New(ctx, "")
				
				// Deploy database component
				databaseOutputs, err := components.DeployDatabase(ctx, cfg, "development")
				require.NoError(t, err)
				
				// Initialize migration runner
				runner, err := NewMigrationRunner(ctx, cfg, "development", databaseOutputs)
				require.NoError(t, err)
				
				// Verify domain migration files are loaded
				runner.MigrationFiles.ApplyT(func(files []interface{}) error {
					fileList := make([]string, len(files))
					for i, f := range files {
						fileList[i] = f.(string)
					}
					
					// Check that domain-specific migration files are present
					found := false
					for _, file := range fileList {
						if file == "/migrations/"+domain+".sql" {
							found = true
							break
						}
					}
					assert.True(t, found, "Migration runner should load %s domain migration file", domain)
					return nil
				})
				
				return nil
			}, pulumi.WithMocks("test", "stack", &MigrationMocks{}))
			
			assert.NoError(t, err)
		})
	}
}

// TestMigrationRunner_EnvironmentParity validates migration runner provides consistent interface across environments
func TestMigrationRunner_EnvironmentParity(t *testing.T) {
	environments := []string{"development", "staging", "production"}
	
	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				cfg := config.New(ctx, "")
				
				// Deploy database component
				databaseOutputs, err := components.DeployDatabase(ctx, cfg, env)
				require.NoError(t, err)
				
				// Initialize migration runner
				runner, err := NewMigrationRunner(ctx, cfg, env, databaseOutputs)
				require.NoError(t, err)
				
				// Verify all environments provide required outputs
				pulumi.All(runner.ConnectionString, runner.MigrationFiles, runner.Strategy).ApplyT(func(args []interface{}) error {
					connectionString := args[0].(string)
					migrationFiles := args[1].([]interface{})
					strategy := args[2].(string)
					
					assert.NotEmpty(t, connectionString, "All environments should provide connection string")
					assert.NotEmpty(t, migrationFiles, "All environments should provide migration files")
					assert.NotEmpty(t, strategy, "All environments should provide migration strategy")
					return nil
				})
				
				return nil
			}, pulumi.WithMocks("test", "stack", &MigrationMocks{}))
			
			assert.NoError(t, err)
		})
	}
}

// TestMigrationRunner_MigrationExecution validates migration execution with environment-specific policies
func TestMigrationRunner_MigrationExecution(t *testing.T) {
	environments := []string{"development", "staging", "production"}
	
	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				cfg := config.New(ctx, "")
				
				// Deploy database component
				databaseOutputs, err := components.DeployDatabase(ctx, cfg, env)
				require.NoError(t, err)
				
				// Initialize migration runner
				runner, err := NewMigrationRunner(ctx, cfg, env, databaseOutputs)
				require.NoError(t, err)
				
				// Execute migration with environment-specific validation
				result, err := runner.ExecuteMigration(ctx)
				require.NoError(t, err)
				
				// Verify migration execution results
				result.ExecutionStatus.ApplyT(func(status string) error {
					assert.Equal(t, "completed", status, "Migration should complete successfully in %s environment", env)
					return nil
				})
				
				return nil
			}, pulumi.WithMocks("test", "stack", &MigrationMocks{}))
			
			assert.NoError(t, err)
		})
	}
}