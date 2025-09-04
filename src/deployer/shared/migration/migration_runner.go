package migration

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type MigrationRunner struct {
	databaseURL string
	domains     []string
	basePath    string
	environment string
}

type MigrationResult struct {
	Domain         string
	Version        uint
	Success        bool
	Error          error
	ExecutionTime  int64
	AppliedMigrations []string
}

type MigrationPlan struct {
	Environment       string
	ExecutionStrategy string
	Domains          []DomainMigrationPlan
	TotalMigrations  int
	EstimatedTime    int64
}

type DomainMigrationPlan struct {
	Domain            string
	MigrationsPath    string
	PendingMigrations []string
	CurrentVersion    uint
	TargetVersion     uint
	Dependencies      []string
}

func NewMigrationRunner(databaseURL, basePath, environment string) *MigrationRunner {
	return &MigrationRunner{
		databaseURL: databaseURL,
		basePath:    basePath,
		environment: environment,
		domains:     []string{"content", "services"},
	}
}

func (mr *MigrationRunner) CreateMigrationPlan(ctx context.Context) (*MigrationPlan, error) {
	plan := &MigrationPlan{
		Environment:       mr.environment,
		ExecutionStrategy: mr.getExecutionStrategy(),
		Domains:          make([]DomainMigrationPlan, 0, len(mr.domains)),
	}

	for _, domain := range mr.domains {
		domainPlan, err := mr.createDomainMigrationPlan(ctx, domain)
		if err != nil {
			return nil, fmt.Errorf("failed to create migration plan for domain %s: %w", domain, err)
		}
		plan.Domains = append(plan.Domains, *domainPlan)
		plan.TotalMigrations += len(domainPlan.PendingMigrations)
	}

	plan.EstimatedTime = mr.calculateEstimatedTime(plan.TotalMigrations)
	return plan, nil
}

func (mr *MigrationRunner) ExecuteMigrationPlan(ctx context.Context, plan *MigrationPlan) ([]MigrationResult, error) {
	results := make([]MigrationResult, 0, len(plan.Domains))

	orderedDomains := mr.getExecutionOrder(plan.Domains)

	for _, domainPlan := range orderedDomains {
		result := mr.executeDomainMigrations(ctx, domainPlan)
		results = append(results, result)

		if !result.Success && mr.shouldStopOnError() {
			return results, fmt.Errorf("migration failed for domain %s: %w", result.Domain, result.Error)
		}
	}

	return results, nil
}

func (mr *MigrationRunner) GetCurrentVersions(ctx context.Context) (map[string]uint, error) {
	versions := make(map[string]uint)

	for _, domain := range mr.domains {
		version, err := mr.getCurrentVersion(ctx, domain)
		if err != nil {
			return nil, fmt.Errorf("failed to get current version for domain %s: %w", domain, err)
		}
		versions[domain] = version
	}

	return versions, nil
}

func (mr *MigrationRunner) ValidateMigrations(ctx context.Context) error {
	for _, domain := range mr.domains {
		if err := mr.validateDomainMigrations(ctx, domain); err != nil {
			return fmt.Errorf("validation failed for domain %s: %w", domain, err)
		}
	}
	return nil
}

func (mr *MigrationRunner) createDomainMigrationPlan(ctx context.Context, domain string) (*DomainMigrationPlan, error) {
	migrationsPath := filepath.Join(mr.basePath, "../backend/internal", domain, "migrations")
	
	currentVersion, err := mr.getCurrentVersion(ctx, domain)
	if err != nil {
		return nil, err
	}

	pendingMigrations, err := mr.getPendingMigrations(migrationsPath, currentVersion)
	if err != nil {
		return nil, err
	}

	targetVersion := mr.calculateTargetVersion(pendingMigrations, currentVersion)

	return &DomainMigrationPlan{
		Domain:            domain,
		MigrationsPath:    migrationsPath,
		PendingMigrations: pendingMigrations,
		CurrentVersion:    currentVersion,
		TargetVersion:     targetVersion,
		Dependencies:      mr.getDomainDependencies(domain),
	}, nil
}

func (mr *MigrationRunner) executeDomainMigrations(ctx context.Context, domainPlan DomainMigrationPlan) MigrationResult {
	result := MigrationResult{
		Domain:  domainPlan.Domain,
		Success: false,
	}

	db, err := sql.Open("postgres", mr.databaseURL)
	if err != nil {
		result.Error = fmt.Errorf("failed to connect to database: %w", err)
		return result
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{
		DatabaseName: fmt.Sprintf("%s_schema", domainPlan.Domain),
		SchemaName:   fmt.Sprintf("%s_schema", domainPlan.Domain),
	})
	if err != nil {
		result.Error = fmt.Errorf("failed to create postgres driver: %w", err)
		return result
	}

	fileSource, err := (&file.File{}).Open(fmt.Sprintf("file://%s", domainPlan.MigrationsPath))
	if err != nil {
		result.Error = fmt.Errorf("failed to open migration files: %w", err)
		return result
	}

	m, err := migrate.NewWithInstance("file", fileSource, "postgres", driver)
	if err != nil {
		result.Error = fmt.Errorf("failed to create migrate instance: %w", err)
		return result
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		result.Error = fmt.Errorf("failed to run migrations: %w", err)
		return result
	}

	currentVersion, _, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		result.Error = fmt.Errorf("failed to get migration version: %w", err)
		return result
	}

	result.Version = currentVersion
	result.Success = true
	result.AppliedMigrations = domainPlan.PendingMigrations
	return result
}

func (mr *MigrationRunner) getCurrentVersion(ctx context.Context, domain string) (uint, error) {
	db, err := sql.Open("postgres", mr.databaseURL)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var version sql.NullInt64
	query := fmt.Sprintf("SELECT version FROM %s_schema.schema_migrations ORDER BY version DESC LIMIT 1", domain)
	err = db.QueryRowContext(ctx, query).Scan(&version)
	
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	if version.Valid {
		return uint(version.Int64), nil
	}
	return 0, nil
}

func (mr *MigrationRunner) getPendingMigrations(migrationsPath string, currentVersion uint) ([]string, error) {
	fileSource, err := (&file.File{}).Open(fmt.Sprintf("file://%s", migrationsPath))
	if err != nil {
		return nil, err
	}
	defer fileSource.Close()

	first, err := fileSource.First()
	if err != nil {
		return []string{}, nil
	}

	var migrations []string
	version := first
	for {
		if version > currentVersion {
			migrations = append(migrations, fmt.Sprintf("%d", version))
		}
		
		next, err := fileSource.Next(version)
		if err != nil {
			break
		}
		version = next
	}

	return migrations, nil
}

func (mr *MigrationRunner) validateDomainMigrations(ctx context.Context, domain string) error {
	migrationsPath := filepath.Join(mr.basePath, "../backend/internal", domain, "migrations")
	
	fileSource, err := (&file.File{}).Open(fmt.Sprintf("file://%s", migrationsPath))
	if err != nil {
		return fmt.Errorf("failed to open migration files for domain %s: %w", domain, err)
	}
	defer fileSource.Close()

	return nil
}

func (mr *MigrationRunner) calculateTargetVersion(pendingMigrations []string, currentVersion uint) uint {
	if len(pendingMigrations) == 0 {
		return currentVersion
	}

	versions := make([]uint, 0, len(pendingMigrations))
	for _, migration := range pendingMigrations {
		var version uint
		fmt.Sscanf(migration, "%d", &version)
		versions = append(versions, version)
	}

	if len(versions) > 0 {
		sort.Slice(versions, func(i, j int) bool { return versions[i] > versions[j] })
		return versions[0]
	}

	return currentVersion
}

func (mr *MigrationRunner) getExecutionStrategy() string {
	switch mr.environment {
	case "development":
		return "aggressive"
	case "staging":
		return "careful"
	case "production":
		return "conservative"
	default:
		return "careful"
	}
}

func (mr *MigrationRunner) getExecutionOrder(domainPlans []DomainMigrationPlan) []DomainMigrationPlan {
	ordered := make([]DomainMigrationPlan, 0, len(domainPlans))
	
	for _, plan := range domainPlans {
		if plan.Domain == "content" {
			ordered = append(ordered, plan)
		}
	}
	
	for _, plan := range domainPlans {
		if plan.Domain == "services" {
			ordered = append(ordered, plan)
		}
	}
	
	return ordered
}

func (mr *MigrationRunner) getDomainDependencies(domain string) []string {
	dependencies := map[string][]string{
		"content":  {},
		"services": {"content"},
	}
	
	if deps, exists := dependencies[domain]; exists {
		return deps
	}
	return []string{}
}

func (mr *MigrationRunner) calculateEstimatedTime(totalMigrations int) int64 {
	baseTime := int64(5000) 
	return baseTime * int64(totalMigrations)
}

func (mr *MigrationRunner) shouldStopOnError() bool {
	return mr.environment != "development"
}