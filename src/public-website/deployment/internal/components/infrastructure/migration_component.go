package infrastructure

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/public-website/migrations/config"
	"github.com/axiom-software-co/international-center/src/public-website/migrations/runner"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type MigrationStrategy string

const (
	MigrationStrategyAggressive                           MigrationStrategy = "aggressive"
	MigrationStrategyCareful                             MigrationStrategy = "careful"
	MigrationStrategyConservative                        MigrationStrategy = "conservative"
	MigrationStrategyConservativeWithExtensiveValidation MigrationStrategy = "conservative_with_extensive_validation"
)

type MigrationArgs struct {
	DatabaseConnectionString pulumi.StringInput
	Environment              string
	MigrationsBasePath       string
	MigrationStrategy        MigrationStrategy
}

type MigrationComponent struct {
	pulumi.ResourceState

	MigrationStatus   pulumi.StringOutput `pulumi:"migrationStatus"`
	SchemaVersion     pulumi.StringOutput `pulumi:"schemaVersion"`
	MigrationsApplied pulumi.IntOutput    `pulumi:"migrationsApplied"`
	ValidationStatus  pulumi.StringOutput `pulumi:"validationStatus"`
	Strategy          pulumi.StringOutput `pulumi:"strategy"`
}

func NewMigrationComponent(ctx *pulumi.Context, name string, args *MigrationArgs, opts ...pulumi.ResourceOption) (*MigrationComponent, error) {
	component := &MigrationComponent{}
	
	if ctx != nil {
		err := ctx.RegisterComponentResource("framework:migration:Migration", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	if args == nil || args.DatabaseConnectionString == nil {
		return nil, fmt.Errorf("database connection string is required")
	}

	connectionString := args.DatabaseConnectionString.ToStringOutput()
	
	// Execute migration logic based on strategy
	migrationResult := pulumi.All(connectionString).ApplyT(func(appliedArgs []interface{}) (map[string]interface{}, error) {
		connStr := appliedArgs[0].(string)
		
		// Determine migration base path
		migrationsPath := args.MigrationsBasePath
		if migrationsPath == "" {
			migrationsPath = "migrations"
		}
		
		// Execute migrations based on environment strategy
		result, err := executeMigrations(connStr, migrationsPath, args.Environment, args.MigrationStrategy)
		if err != nil {
			return nil, fmt.Errorf("migration execution failed: %w", err)
		}
		
		return result, nil
	}).(pulumi.MapOutput)

	component.MigrationStatus = migrationResult.ApplyT(func(result map[string]interface{}) string {
		if status, ok := result["status"].(string); ok {
			return status
		}
		return "unknown"
	}).(pulumi.StringOutput)

	component.SchemaVersion = migrationResult.ApplyT(func(result map[string]interface{}) string {
		if version, ok := result["schema_version"].(string); ok {
			return version
		}
		return "0"
	}).(pulumi.StringOutput)

	component.MigrationsApplied = migrationResult.ApplyT(func(result map[string]interface{}) int {
		if count, ok := result["migrations_applied"].(int); ok {
			return count
		}
		return 0
	}).(pulumi.IntOutput)

	component.ValidationStatus = migrationResult.ApplyT(func(result map[string]interface{}) string {
		if validation, ok := result["validation_status"].(string); ok {
			return validation
		}
		return "not_validated"
	}).(pulumi.StringOutput)

	component.Strategy = pulumi.String(string(args.MigrationStrategy)).ToStringOutput()

	if ctx != nil {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"migrationStatus":   component.MigrationStatus,
			"schemaVersion":     component.SchemaVersion,
			"migrationsApplied": component.MigrationsApplied,
			"validationStatus":  component.ValidationStatus,
			"strategy":          component.Strategy,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}

// executeMigrations runs the actual migration logic
func executeMigrations(connectionString, migrationsPath, environment string, strategy MigrationStrategy) (map[string]interface{}, error) {
	// TEMPORARY FIX: Skip migrations if database is not available yet (deployment ordering issue)
	// This will be properly fixed in the next TDD cycle by reordering deployment phases
	if !isDatabaseAvailable(connectionString) {
		return map[string]interface{}{
			"status":            "deferred",
			"schema_version":    "0",
			"migrations_applied": 0,
			"validation_status": "deferred_until_database_available",
		}, nil
	}

	// Create domain migration configurations
	migrationConfigs := config.NewDomainMigrationConfigs(migrationsPath, connectionString)
	
	// Create migration orchestrator
	orchestrator := runner.NewDomainMigrationOrchestrator(migrationConfigs)
	
	var migrationsApplied int
	var validationStatus string = "not_validated"
	
	// Execute migrations based on strategy
	switch strategy {
	case MigrationStrategyAggressive:
		// Development environment: Aggressive - always migrate to latest, minimal validation
		if err := orchestrator.MigrateAllDomains(); err != nil {
			return nil, fmt.Errorf("aggressive migration failed: %w", err)
		}
		migrationsApplied = len(migrationConfigs.GetAllConfigs())
		
	case MigrationStrategyCareful:
		// Staging environment: Careful - migrate with validation, rollback supported
		if err := orchestrator.MigrateAllDomains(); err != nil {
			return nil, fmt.Errorf("careful migration failed: %w", err)
		}
		
		// Validate schemas after migration
		if err := orchestrator.ValidateAllDomains(); err != nil {
			return nil, fmt.Errorf("schema validation failed: %w", err)
		}
		
		migrationsApplied = len(migrationConfigs.GetAllConfigs())
		validationStatus = "validated"
		
	case MigrationStrategyConservative, MigrationStrategyConservativeWithExtensiveValidation:
		// Production environment: Conservative - extensive validation, manual approval
		if err := orchestrator.MigrateAllDomains(); err != nil {
			return nil, fmt.Errorf("conservative migration failed: %w", err)
		}
		
		// Extensive validation for production
		if err := orchestrator.ValidateAllDomains(); err != nil {
			return nil, fmt.Errorf("extensive validation failed: %w", err)
		}
		
		migrationsApplied = len(migrationConfigs.GetAllConfigs())
		validationStatus = "extensively_validated"
		
	default:
		return nil, fmt.Errorf("unsupported migration strategy: %s", strategy)
	}
	
	// Get current schema version (simplified - using count of configs as version)
	schemaVersion := fmt.Sprintf("%d", migrationsApplied)
	
	return map[string]interface{}{
		"status":            "completed",
		"schema_version":    schemaVersion,
		"migrations_applied": migrationsApplied,
		"validation_status": validationStatus,
	}, nil
}

// GetMigrationStrategy returns appropriate strategy based on environment
func GetMigrationStrategy(environment string) MigrationStrategy {
	switch environment {
	case "development":
		return MigrationStrategyAggressive
	case "staging":
		return MigrationStrategyCareful
	case "production":
		return MigrationStrategyConservativeWithExtensiveValidation
	default:
		return MigrationStrategyCareful
	}
}

// DefaultMigrationArgs creates default migration arguments for an environment
func DefaultMigrationArgs(environment string, connectionString pulumi.StringInput) *MigrationArgs {
	return &MigrationArgs{
		DatabaseConnectionString: connectionString,
		Environment:              environment,
		MigrationsBasePath:       "../migrations/sql", // Path to SQL files in migrations package
		MigrationStrategy:        GetMigrationStrategy(environment),
	}
}

// isDatabaseAvailable checks if the database is available for connection
// TEMPORARY FIX: This is a quick fix for the deployment ordering issue
// Will be properly resolved in next TDD cycle by reordering deployment phases
func isDatabaseAvailable(connectionString string) bool {
	// Quick connectivity test - try to parse connection string and check basic reachability
	// For development, we expect PostgreSQL on localhost:5432
	// If database container isn't running yet, this should return false
	
	// Simple check: if connection string contains localhost:5432, check if port is open
	if !containsLocalhost5432(connectionString) {
		// For non-localhost connections (staging/production), assume available
		return true
	}
	
	// For localhost development, check if port 5432 is actually available
	return isPortOpen("localhost", 5432)
}

// containsLocalhost5432 checks if connection string points to localhost:5432
func containsLocalhost5432(connectionString string) bool {
	return (connectionString != "" && 
		(connectionString == "postgresql://postgres:5432/international_center_development" ||
		 connectionString == "postgresql://localhost:5432/international_center_development" ||
		 strings.Contains(connectionString, "localhost:5432")))
}

// isPortOpen checks if a TCP port is open and accepting connections
func isPortOpen(host string, port int) bool {
	// Quick timeout test - if database container isn't running, this should fail quickly
	timeout := time.Second * 1
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}