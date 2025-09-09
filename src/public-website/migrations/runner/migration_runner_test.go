package runner

import (
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/public-website/migrations/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationRunnerInterface(t *testing.T) {
	t.Run("NewMigrationRunner creates runner with valid config", func(t *testing.T) {
		cfg := config.MigrationConfig{
			Domain:        "test-domain",
			MigrationPath: "/tmp/test-migrations",
			DatabaseURL:   "postgres://localhost:5432/test",
		}
		
		runner := NewMigrationRunner(cfg)
		
		require.NotNil(t, runner, "NewMigrationRunner should return a valid runner")
		assert.Equal(t, cfg.Domain, runner.config.Domain, "Runner should store the provided config")
	})
}

func TestMigrationRunnerContracts(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		expectError bool
	}{
		{
			name:        "valid domain configuration",
			domain:      "content-services",
			expectError: false,
		},
		{
			name:        "empty domain should fail",
			domain:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.MigrationConfig{
				Domain:        tt.domain,
				MigrationPath: "/tmp/test-migrations",
				DatabaseURL:   "postgres://localhost:5432/test",
			}
			
			runner := NewMigrationRunner(cfg)
			
			if tt.expectError {
				// For invalid configurations, we expect methods to return errors
				err := runner.MigrateUp()
				assert.Error(t, err, "MigrateUp should return error for invalid config")
				
				err = runner.MigrateDown(1)
				assert.Error(t, err, "MigrateDown should return error for invalid config")
				
				_, _, err = runner.GetVersion()
				assert.Error(t, err, "GetVersion should return error for invalid config")
			} else {
				// For valid configurations, methods should exist (will fail with connection errors in tests, but methods should exist)
				assert.NotPanics(t, func() {
					runner.MigrateUp()
				}, "MigrateUp method should exist and not panic")
				
				assert.NotPanics(t, func() {
					runner.MigrateDown(1)
				}, "MigrateDown method should exist and not panic")
				
				assert.NotPanics(t, func() {
					runner.GetVersion()
				}, "GetVersion method should exist and not panic")
				
				assert.NotPanics(t, func() {
					runner.Force(1)
				}, "Force method should exist and not panic")
				
				assert.NotPanics(t, func() {
					runner.ValidateSchema()
				}, "ValidateSchema method should exist and not panic")
			}
		})
	}
}

func TestDomainMigrationOrchestratorInterface(t *testing.T) {
	t.Run("NewDomainMigrationOrchestrator creates orchestrator with configs", func(t *testing.T) {
		configs := &config.DomainMigrationConfigs{}
		
		orchestrator := NewDomainMigrationOrchestrator(configs)
		
		require.NotNil(t, orchestrator, "NewDomainMigrationOrchestrator should return a valid orchestrator")
	})
	
	t.Run("orchestrator provides expected methods", func(t *testing.T) {
		configs := &config.DomainMigrationConfigs{}
		orchestrator := NewDomainMigrationOrchestrator(configs)
		
		// Test that methods exist and don't panic (they may return errors due to missing DB)
		assert.NotPanics(t, func() {
			orchestrator.MigrateAllDomains()
		}, "MigrateAllDomains method should exist and not panic")
		
		assert.NotPanics(t, func() {
			orchestrator.ValidateAllDomains()
		}, "ValidateAllDomains method should exist and not panic")
	})
}

func TestMigrationRunnerTimeouts(t *testing.T) {
	t.Run("migration operations complete within timeout", func(t *testing.T) {
		cfg := config.MigrationConfig{
			Domain:        "test-domain",
			MigrationPath: "/tmp/test-migrations",
			DatabaseURL:   "postgres://localhost:5432/test",
		}
		
		runner := NewMigrationRunner(cfg)
		
		// Test that operations complete within 5 second timeout (unit test requirement)
		done := make(chan bool, 1)
		timeout := time.After(5 * time.Second)
		
		go func() {
			// These will likely fail with connection errors, but should complete quickly
			runner.MigrateUp()
			runner.MigrateDown(1)
			runner.GetVersion()
			runner.ValidateSchema()
			done <- true
		}()
		
		select {
		case <-done:
			// Test passed - operations completed within timeout
		case <-timeout:
			t.Fatal("Migration runner operations took longer than 5 seconds")
		}
	})
}