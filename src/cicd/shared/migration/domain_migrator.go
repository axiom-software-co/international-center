package migration

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
)

type DomainMigrator struct {
	domain        string
	databaseURL   string
	migrationsPath string
	environment   string
	strategy      MigrationStrategy
}

type MigrationStrategy interface {
	ShouldExecute(migration *MigrationInfo) bool
	GetValidationLevel() ValidationLevel
	GetRollbackPolicy() RollbackPolicy
	GetConcurrency() int
}

type ValidationLevel int

const (
	ValidationMinimal ValidationLevel = iota
	ValidationModerate
	ValidationFull
)

type RollbackPolicy int

const (
	RollbackAutomatic RollbackPolicy = iota
	RollbackConfirmation
	RollbackManual
)

type MigrationInfo struct {
	Version     uint
	Filename    string
	UpSQL       string
	DownSQL     string
	Domain      string
	Environment string
	Size        int64
	Checksum    string
}

type DomainMigrationResult struct {
	Domain            string
	StartTime         time.Time
	EndTime           time.Time
	Success           bool
	Error             error
	MigrationsApplied []MigrationInfo
	RolledBackVersions []uint
	ValidationErrors  []ValidationError
}

type ValidationError struct {
	Version  uint
	Type     string
	Message  string
	Severity string
}

func NewDomainMigrator(domain, databaseURL, basePath, environment string) *DomainMigrator {
	migrationsPath := filepath.Join(basePath, "../backend/internal", domain, "migrations")
	
	return &DomainMigrator{
		domain:         domain,
		databaseURL:    databaseURL,
		migrationsPath: migrationsPath,
		environment:    environment,
		strategy:       createMigrationStrategy(environment),
	}
}

func (dm *DomainMigrator) ExecuteDomainMigrations(ctx context.Context) (*DomainMigrationResult, error) {
	result := &DomainMigrationResult{
		Domain:    dm.domain,
		StartTime: time.Now(),
		Success:   false,
	}

	migrations, err := dm.discoverMigrations(ctx)
	if err != nil {
		result.Error = fmt.Errorf("failed to discover migrations: %w", err)
		result.EndTime = time.Now()
		return result, err
	}

	validationErrors := dm.validateMigrations(migrations)
	result.ValidationErrors = validationErrors

	if dm.hasBlockingValidationErrors(validationErrors) {
		result.Error = fmt.Errorf("blocking validation errors found")
		result.EndTime = time.Now()
		return result, result.Error
	}

	for _, migration := range migrations {
		if !dm.strategy.ShouldExecute(&migration) {
			continue
		}

		err := dm.executeSingleMigration(ctx, &migration)
		if err != nil {
			result.Error = fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
			
			if dm.shouldRollback(err) {
				rollbackVersions, rollbackErr := dm.performRollback(ctx, migration.Version)
				result.RolledBackVersions = rollbackVersions
				if rollbackErr != nil {
					result.Error = fmt.Errorf("migration failed and rollback failed: %w", rollbackErr)
				}
			}
			
			result.EndTime = time.Now()
			return result, result.Error
		}

		result.MigrationsApplied = append(result.MigrationsApplied, migration)
	}

	result.Success = true
	result.EndTime = time.Now()
	return result, nil
}

func (dm *DomainMigrator) GetDomainStatus(ctx context.Context) (*DomainStatus, error) {
	db, err := sql.Open("postgres", dm.databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	var currentVersion sql.NullInt64
	var dirty sql.NullBool
	
	schemaName := fmt.Sprintf("%s_schema", dm.domain)
	query := fmt.Sprintf(`
		SELECT version, dirty 
		FROM %s.schema_migrations 
		ORDER BY version DESC 
		LIMIT 1`, schemaName)
	
	err = db.QueryRowContext(ctx, query).Scan(&currentVersion, &dirty)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get migration status: %w", err)
	}

	status := &DomainStatus{
		Domain:      dm.domain,
		Environment: dm.environment,
		Connected:   true,
	}

	if currentVersion.Valid {
		status.CurrentVersion = uint(currentVersion.Int64)
		status.HasMigrations = true
	}

	if dirty.Valid && dirty.Bool {
		status.IsDirty = true
		status.Status = "dirty"
	} else {
		status.Status = "clean"
	}

	pendingMigrations, err := dm.getPendingMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending migrations: %w", err)
	}

	status.PendingMigrations = len(pendingMigrations)
	status.Ready = !status.IsDirty && status.Connected

	return status, nil
}

type DomainStatus struct {
	Domain            string
	Environment       string  
	CurrentVersion    uint
	PendingMigrations int
	HasMigrations     bool
	IsDirty          bool
	Connected        bool
	Status           string
	Ready            bool
}

func (dm *DomainMigrator) discoverMigrations(ctx context.Context) ([]MigrationInfo, error) {
	fileSource, err := (&file.File{}).Open(fmt.Sprintf("file://%s", dm.migrationsPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open migrations directory: %w", err)
	}
	defer fileSource.Close()

	var migrations []MigrationInfo

	first, err := fileSource.First()
	if err != nil {
		return migrations, nil
	}

	version := first
	for {
		upSQL, downSQL, err := dm.readMigrationFiles(version)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration %d: %w", version, err)
		}

		migration := MigrationInfo{
			Version:     version,
			Filename:    fmt.Sprintf("%d", version),
			UpSQL:       upSQL,
			DownSQL:     downSQL,
			Domain:      dm.domain,
			Environment: dm.environment,
			Size:        int64(len(upSQL)),
		}

		migrations = append(migrations, migration)

		next, err := fileSource.Next(version)
		if err != nil {
			break
		}
		version = next
	}

	return migrations, nil
}

func (dm *DomainMigrator) readMigrationFiles(version uint) (upSQL, downSQL string, err error) {
	return "", "", nil
}

func (dm *DomainMigrator) validateMigrations(migrations []MigrationInfo) []ValidationError {
	var errors []ValidationError

	for _, migration := range migrations {
		migrationErrors := dm.validateSingleMigration(migration)
		errors = append(errors, migrationErrors...)
	}

	return errors
}

func (dm *DomainMigrator) validateSingleMigration(migration MigrationInfo) []ValidationError {
	var errors []ValidationError

	if migration.Size > 1024*1024 {
		errors = append(errors, ValidationError{
			Version:  migration.Version,
			Type:     "size",
			Message:  "migration file is larger than 1MB",
			Severity: "warning",
		})
	}

	if dm.containsDropTable(migration.UpSQL) {
		errors = append(errors, ValidationError{
			Version:  migration.Version,
			Type:     "destructive",
			Message:  "migration contains DROP TABLE statement",
			Severity: "error",
		})
	}

	return errors
}

func (dm *DomainMigrator) containsDropTable(sql string) bool {
	return false
}

func (dm *DomainMigrator) hasBlockingValidationErrors(errors []ValidationError) bool {
	for _, err := range errors {
		if err.Severity == "error" && dm.strategy.GetValidationLevel() == ValidationFull {
			return true
		}
	}
	return false
}

func (dm *DomainMigrator) executeSingleMigration(ctx context.Context, migration *MigrationInfo) error {
	db, err := sql.Open("postgres", dm.databaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{
		DatabaseName: fmt.Sprintf("%s_schema", dm.domain),
		SchemaName:   fmt.Sprintf("%s_schema", dm.domain),
	})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	fileSource, err := (&file.File{}).Open(fmt.Sprintf("file://%s", dm.migrationsPath))
	if err != nil {
		return fmt.Errorf("failed to open migration files: %w", err)
	}
	defer fileSource.Close()

	m, err := migrate.NewWithInstance("file", fileSource, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Migrate(migration.Version); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migration to version %d: %w", migration.Version, err)
	}

	return nil
}

func (dm *DomainMigrator) shouldRollback(err error) bool {
	switch dm.strategy.GetRollbackPolicy() {
	case RollbackAutomatic:
		return true
	case RollbackConfirmation:
		return dm.environment != "production"
	case RollbackManual:
		return false
	default:
		return false
	}
}

func (dm *DomainMigrator) performRollback(ctx context.Context, failedVersion uint) ([]uint, error) {
	db, err := sql.Open("postgres", dm.databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{
		DatabaseName: fmt.Sprintf("%s_schema", dm.domain),
		SchemaName:   fmt.Sprintf("%s_schema", dm.domain),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	fileSource, err := (&file.File{}).Open(fmt.Sprintf("file://%s", dm.migrationsPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open migration files: %w", err)
	}
	defer fileSource.Close()

	m, err := migrate.NewWithInstance("file", fileSource, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	targetVersion := failedVersion - 1
	if err := m.Migrate(targetVersion); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("failed to rollback to version %d: %w", targetVersion, err)
	}

	return []uint{failedVersion}, nil
}

func (dm *DomainMigrator) getPendingMigrations(ctx context.Context) ([]MigrationInfo, error) {
	currentVersion, err := dm.getCurrentVersion(ctx)
	if err != nil {
		return nil, err
	}

	allMigrations, err := dm.discoverMigrations(ctx)
	if err != nil {
		return nil, err
	}

	var pending []MigrationInfo
	for _, migration := range allMigrations {
		if migration.Version > currentVersion {
			pending = append(pending, migration)
		}
	}

	return pending, nil
}

func (dm *DomainMigrator) getCurrentVersion(ctx context.Context) (uint, error) {
	db, err := sql.Open("postgres", dm.databaseURL)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var version sql.NullInt64
	schemaName := fmt.Sprintf("%s_schema", dm.domain)
	query := fmt.Sprintf("SELECT version FROM %s.schema_migrations ORDER BY version DESC LIMIT 1", schemaName)
	
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

func createMigrationStrategy(environment string) MigrationStrategy {
	switch environment {
	case "development":
		return &AggressiveStrategy{}
	case "staging":
		return &CarefulStrategy{}
	case "production":
		return &ConservativeStrategy{}
	default:
		return &CarefulStrategy{}
	}
}

type AggressiveStrategy struct{}

func (s *AggressiveStrategy) ShouldExecute(migration *MigrationInfo) bool { return true }
func (s *AggressiveStrategy) GetValidationLevel() ValidationLevel { return ValidationMinimal }
func (s *AggressiveStrategy) GetRollbackPolicy() RollbackPolicy { return RollbackAutomatic }
func (s *AggressiveStrategy) GetConcurrency() int { return 1 }

type CarefulStrategy struct{}

func (s *CarefulStrategy) ShouldExecute(migration *MigrationInfo) bool { return true }
func (s *CarefulStrategy) GetValidationLevel() ValidationLevel { return ValidationModerate }
func (s *CarefulStrategy) GetRollbackPolicy() RollbackPolicy { return RollbackConfirmation }
func (s *CarefulStrategy) GetConcurrency() int { return 1 }

type ConservativeStrategy struct{}

func (s *ConservativeStrategy) ShouldExecute(migration *MigrationInfo) bool { return true }
func (s *ConservativeStrategy) GetValidationLevel() ValidationLevel { return ValidationFull }
func (s *ConservativeStrategy) GetRollbackPolicy() RollbackPolicy { return RollbackManual }
func (s *ConservativeStrategy) GetConcurrency() int { return 1 }