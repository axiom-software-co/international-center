package main

import (
	"fmt"
	
	"github.com/axiom-software-co/international-center/src/public-website/migrations/config"
	"github.com/axiom-software-co/international-center/src/public-website/migrations/runner"
)

func main() {
	// Test that we can import and use the migrations package
	testConfig := config.MigrationConfig{
		Domain:        "test-domain",
		MigrationPath: "/tmp/test-migrations",
		DatabaseURL:   "postgres://localhost:5432/test",
	}
	
	migrationRunner := runner.NewMigrationRunner(testConfig)
	if migrationRunner != nil {
		fmt.Println("✓ Successfully imported and used migrations package")
	}
	
	configs := config.NewDomainMigrationConfigs("/tmp/test", "postgres://localhost:5432/test")
	if configs != nil {
		fmt.Println("✓ Successfully created domain migration configs")
	}
	
	orchestrator := runner.NewDomainMigrationOrchestrator(configs)
	if orchestrator != nil {
		fmt.Println("✓ Successfully created migration orchestrator")
	}
	
	fmt.Println("Migration package integration successful!")
}