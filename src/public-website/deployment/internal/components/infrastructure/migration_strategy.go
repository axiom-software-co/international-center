package infrastructure

import (
	"fmt"
	"log"
)

type MigrationEnvironmentStrategy struct {
	Environment     string
	Strategy        MigrationStrategy
	Approach        string
	RollbackPolicy  string
	SafetyChecks    string
	Automation      string
	RequireApproval bool
}

var EnvironmentStrategies = map[string]MigrationEnvironmentStrategy{
	"development": {
		Environment:     "development",
		Strategy:        MigrationStrategyAggressive,
		Approach:        "Aggressive - always migrate to latest",
		RollbackPolicy:  "Easy - can destroy and recreate",
		SafetyChecks:    "Minimal validation",
		Automation:      "Full automation",
		RequireApproval: false,
	},
	"staging": {
		Environment:     "staging",
		Strategy:        MigrationStrategyCareful,
		Approach:        "Careful - migrate with validation",
		RollbackPolicy:  "Supported with confirmation",
		SafetyChecks:    "Moderate validation",
		Automation:      "Pulumi orchestrated via GitHub Actions",
		RequireApproval: false,
	},
	"production": {
		Environment:     "production",
		Strategy:        MigrationStrategyConservativeWithExtensiveValidation,
		Approach:        "Conservative - extensive validation",
		RollbackPolicy:  "Manual approval required",
		SafetyChecks:    "Full validation and backup",
		Automation:      "Pulumi orchestrated with human approval",
		RequireApproval: true,
	},
}

// GetEnvironmentStrategy returns the migration strategy for the given environment
func GetEnvironmentStrategy(environment string) (MigrationEnvironmentStrategy, error) {
	if strategy, exists := EnvironmentStrategies[environment]; exists {
		return strategy, nil
	}
	
	// Default to staging strategy for unknown environments
	log.Printf("Unknown environment '%s', defaulting to staging strategy", environment)
	return EnvironmentStrategies["staging"], nil
}

// ValidateEnvironmentStrategy ensures the environment has a valid migration strategy
func ValidateEnvironmentStrategy(environment string) error {
	validEnvironments := []string{"development", "staging", "production"}
	for _, validEnv := range validEnvironments {
		if environment == validEnv {
			return nil
		}
	}
	return fmt.Errorf("invalid environment '%s'. Valid environments: %v", environment, validEnvironments)
}

// LogMigrationStrategyInfo logs the migration strategy information for the environment
func LogMigrationStrategyInfo(environment string) {
	strategy, err := GetEnvironmentStrategy(environment)
	if err != nil {
		log.Printf("Error getting migration strategy: %v", err)
		return
	}
	
	log.Printf("Migration Strategy for %s:", strategy.Environment)
	log.Printf("  Approach: %s", strategy.Approach)
	log.Printf("  Rollback Policy: %s", strategy.RollbackPolicy)
	log.Printf("  Safety Checks: %s", strategy.SafetyChecks)
	log.Printf("  Automation: %s", strategy.Automation)
	log.Printf("  Requires Approval: %t", strategy.RequireApproval)
}

// GetMigrationStrategyDescription returns a human-readable description of the strategy
func GetMigrationStrategyDescription(strategy MigrationStrategy) string {
	descriptions := map[MigrationStrategy]string{
		MigrationStrategyAggressive: "Aggressive migration approach with minimal validation for rapid development",
		MigrationStrategyCareful:    "Careful migration approach with moderate validation for staging environments",
		MigrationStrategyConservative: "Conservative migration approach with full validation for production",
		MigrationStrategyConservativeWithExtensiveValidation: "Conservative migration with extensive validation and manual approval gates for critical production systems",
	}
	
	if desc, exists := descriptions[strategy]; exists {
		return desc
	}
	return "Unknown migration strategy"
}