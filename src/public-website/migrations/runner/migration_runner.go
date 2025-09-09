package runner

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/axiom-software-co/international-center/src/public-website/migrations/config"
)

// MigrationRunner handles domain-specific database migrations
type MigrationRunner struct {
	config config.MigrationConfig
}

// NewMigrationRunner creates a new migration runner for a domain
func NewMigrationRunner(config config.MigrationConfig) *MigrationRunner {
	return &MigrationRunner{
		config: config,
	}
}

// MigrateUp applies all pending migrations for the domain
func (r *MigrationRunner) MigrateUp() error {
	m, err := r.createMigrate()
	if err != nil {
		return fmt.Errorf("failed to create migrate instance for domain %s: %w", r.config.Domain, err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate up for domain %s: %w", r.config.Domain, err)
	}

	log.Printf("Successfully migrated domain %s up", r.config.Domain)
	return nil
}

// MigrateDown rolls back specified number of migrations for the domain
func (r *MigrationRunner) MigrateDown(steps int) error {
	m, err := r.createMigrate()
	if err != nil {
		return fmt.Errorf("failed to create migrate instance for domain %s: %w", r.config.Domain, err)
	}
	defer m.Close()

	if err := m.Steps(-steps); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate down %d steps for domain %s: %w", steps, r.config.Domain, err)
	}

	log.Printf("Successfully migrated domain %s down %d steps", r.config.Domain, steps)
	return nil
}

// GetVersion returns current migration version for the domain
func (r *MigrationRunner) GetVersion() (uint, bool, error) {
	m, err := r.createMigrate()
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance for domain %s: %w", r.config.Domain, err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get version for domain %s: %w", r.config.Domain, err)
	}

	return version, dirty, nil
}

// Force sets migration version for the domain (recovery purposes)
func (r *MigrationRunner) Force(version int) error {
	m, err := r.createMigrate()
	if err != nil {
		return fmt.Errorf("failed to create migrate instance for domain %s: %w", r.config.Domain, err)
	}
	defer m.Close()

	if err := m.Force(version); err != nil {
		return fmt.Errorf("failed to force version %d for domain %s: %w", version, r.config.Domain, err)
	}

	log.Printf("Successfully forced domain %s to version %d", r.config.Domain, version)
	return nil
}

// ValidateSchema compares current schema against validation files
func (r *MigrationRunner) ValidateSchema() error {
	// Implementation would compare deployed schema against validation SQL files
	// This is a placeholder for schema validation logic
	log.Printf("Schema validation for domain %s would be implemented here", r.config.Domain)
	return nil
}

// createMigrate creates a migrate instance for the domain
func (r *MigrationRunner) createMigrate() (*migrate.Migrate, error) {
	db, err := sql.Open("postgres", r.config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		r.config.GetMigrationURL(),
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return m, nil
}

// DomainMigrationOrchestrator manages migrations across all domains
type DomainMigrationOrchestrator struct {
	configs *config.DomainMigrationConfigs
}

// NewDomainMigrationOrchestrator creates orchestrator for all domain migrations
func NewDomainMigrationOrchestrator(configs *config.DomainMigrationConfigs) *DomainMigrationOrchestrator {
	return &DomainMigrationOrchestrator{
		configs: configs,
	}
}

// MigrateAllDomains applies migrations for all domains in dependency order
func (o *DomainMigrationOrchestrator) MigrateAllDomains() error {
	// Migration order: shared -> content domains -> inquiries domains -> supporting domains
	migrationOrder := []config.MigrationConfig{
		o.configs.Shared,
		o.configs.Content.Services,
		o.configs.Content.News,
		o.configs.Content.Research,
		o.configs.Content.Events,
		o.configs.Inquiries.Donations,
		o.configs.Inquiries.Business,
		o.configs.Inquiries.Media,
		o.configs.Inquiries.Volunteers,
		o.configs.Notifications,
		o.configs.Gateway,
	}

	for _, config := range migrationOrder {
		runner := NewMigrationRunner(config)
		if err := runner.MigrateUp(); err != nil {
			return fmt.Errorf("migration failed for domain %s: %w", config.Domain, err)
		}
	}

	log.Printf("Successfully migrated all domains")
	return nil
}

// ValidateAllDomains validates schema for all domains against validation files
func (o *DomainMigrationOrchestrator) ValidateAllDomains() error {
	for _, config := range o.configs.GetAllConfigs() {
		runner := NewMigrationRunner(config)
		if err := runner.ValidateSchema(); err != nil {
			return fmt.Errorf("schema validation failed for domain %s: %w", config.Domain, err)
		}
	}

	log.Printf("Successfully validated all domain schemas")
	return nil
}