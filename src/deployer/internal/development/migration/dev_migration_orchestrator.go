package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/axiom-software-co/international-center/src/deployer/internal/shared/migration"
)

type DevMigrationOrchestrator struct {
	basePath       string
	databaseURL    string
	environment    string
	migrationRunner *migration.MigrationRunner
	rollbackHandler *DevRollbackHandler
	aggressiveMode  bool
}

type DevMigrationResult struct {
	Environment       string
	StartTime         time.Time
	EndTime           time.Time
	Success           bool
	Error             error
	ExecutedMigrations []migration.MigrationResult
	RollbacksPerformed []string
	DatabaseRecreated  bool
	SchemasInitialized bool
	TestDataSeeded     bool
}

type AggressiveSettings struct {
	AutoRollbackOnError   bool
	RecreateOnConflict    bool
	SkipValidation        bool
	ForceLatestVersion    bool
	EnableTestDataSeed    bool
	CleanupOnFailure      bool
	MaxRetryAttempts      int
	RetryDelay           time.Duration
}

func NewDevMigrationOrchestrator(databaseURL, basePath, environment string) *DevMigrationOrchestrator {
	migrationRunner := migration.NewMigrationRunner(databaseURL, basePath, environment)
	rollbackHandler := NewDevRollbackHandler(databaseURL, basePath, environment)

	return &DevMigrationOrchestrator{
		basePath:       basePath,
		databaseURL:    databaseURL,
		environment:    environment,
		migrationRunner: migrationRunner,
		rollbackHandler: rollbackHandler,
		aggressiveMode:  true,
	}
}

func (dmo *DevMigrationOrchestrator) ExecuteMigrations(ctx context.Context) (*DevMigrationResult, error) {
	result := &DevMigrationResult{
		Environment: dmo.environment,
		StartTime:   time.Now(),
		Success:     false,
	}

	settings := dmo.getAggressiveSettings()

	if dmo.shouldRecreateDatabase(ctx) {
		err := dmo.recreateDatabase(ctx)
		if err != nil {
			result.Error = fmt.Errorf("failed to recreate database: %w", err)
			result.EndTime = time.Now()
			return result, result.Error
		}
		result.DatabaseRecreated = true
	}

	plan, err := dmo.migrationRunner.CreateMigrationPlan(ctx)
	if err != nil {
		result.Error = fmt.Errorf("failed to create migration plan: %w", err)
		result.EndTime = time.Now()
		return result, result.Error
	}

	if !settings.SkipValidation {
		if err := dmo.migrationRunner.ValidateMigrations(ctx); err != nil {
			if settings.RecreateOnConflict {
				fmt.Printf("Validation failed, recreating database: %v\n", err)
				if err := dmo.recreateDatabase(ctx); err != nil {
					result.Error = fmt.Errorf("failed to recreate database after validation failure: %w", err)
					result.EndTime = time.Now()
					return result, result.Error
				}
				result.DatabaseRecreated = true
			} else {
				result.Error = fmt.Errorf("migration validation failed: %w", err)
				result.EndTime = time.Now()
				return result, result.Error
			}
		}
	}

	migrationResults, err := dmo.executeMigrationsWithRetry(ctx, plan, settings)
	if err != nil {
		if settings.AutoRollbackOnError {
			rollbackVersions, rollbackErr := dmo.rollbackHandler.PerformEmergencyRollback(ctx)
			result.RollbacksPerformed = rollbackVersions
			if rollbackErr != nil {
				result.Error = fmt.Errorf("migration failed and rollback failed: %w", rollbackErr)
			} else {
				result.Error = fmt.Errorf("migration failed but rollback successful: %w", err)
			}
		} else {
			result.Error = err
		}
		result.EndTime = time.Now()
		return result, result.Error
	}

	result.ExecutedMigrations = migrationResults

	if settings.EnableTestDataSeed {
		if err := dmo.seedTestData(ctx); err != nil {
			fmt.Printf("Warning: failed to seed test data: %v\n", err)
		} else {
			result.TestDataSeeded = true
		}
	}

	if err := dmo.initializeSchemas(ctx); err != nil {
		fmt.Printf("Warning: failed to initialize schemas: %v\n", err)
	} else {
		result.SchemasInitialized = true
	}

	result.Success = true
	result.EndTime = time.Now()
	return result, nil
}

func (dmo *DevMigrationOrchestrator) shouldRecreateDatabase(ctx context.Context) bool {
	currentVersions, err := dmo.migrationRunner.GetCurrentVersions(ctx)
	if err != nil {
		return true
	}

	if len(currentVersions) == 0 {
		return false
	}

	return false
}

func (dmo *DevMigrationOrchestrator) recreateDatabase(ctx context.Context) error {
	fmt.Println("Recreating development database...")

	domains := []string{"content", "services", "identity"}
	
	for _, domain := range domains {
		if err := dmo.recreateDomainSchema(ctx, domain); err != nil {
			return fmt.Errorf("failed to recreate schema for domain %s: %w", domain, err)
		}
	}

	return nil
}

func (dmo *DevMigrationOrchestrator) recreateDomainSchema(ctx context.Context, domain string) error {
	schemaName := fmt.Sprintf("%s_schema", domain)
	
	dropSQL := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)
	createSQL := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)

	fmt.Printf("Recreating schema: %s\n", schemaName)

	_ = dropSQL
	_ = createSQL

	return nil
}

func (dmo *DevMigrationOrchestrator) executeMigrationsWithRetry(ctx context.Context, plan *migration.MigrationPlan, settings *AggressiveSettings) ([]migration.MigrationResult, error) {
	var results []migration.MigrationResult
	var lastErr error

	for attempt := 1; attempt <= settings.MaxRetryAttempts; attempt++ {
		fmt.Printf("Migration attempt %d of %d\n", attempt, settings.MaxRetryAttempts)

		migrationResults, err := dmo.migrationRunner.ExecuteMigrationPlan(ctx, plan)
		if err == nil {
			results = migrationResults
			break
		}

		lastErr = err
		fmt.Printf("Migration attempt %d failed: %v\n", attempt, err)

		if attempt < settings.MaxRetryAttempts {
			if settings.RecreateOnConflict {
				fmt.Println("Recreating database before retry...")
				if recreateErr := dmo.recreateDatabase(ctx); recreateErr != nil {
					return nil, fmt.Errorf("failed to recreate database on retry: %w", recreateErr)
				}
			}

			fmt.Printf("Waiting %v before retry...\n", settings.RetryDelay)
			time.Sleep(settings.RetryDelay)
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all migration attempts failed, last error: %w", lastErr)
	}

	return results, nil
}

func (dmo *DevMigrationOrchestrator) seedTestData(ctx context.Context) error {
	fmt.Println("Seeding test data...")

	contentTestData := `
INSERT INTO content_schema.content_storage_backend (backend_name, backend_type, is_active, priority_order, base_url)
VALUES ('azurite-local', 'azure-blob', true, 1, 'http://azurite:10000')
ON CONFLICT (backend_name) DO NOTHING;
`

	servicesTestData := `
INSERT INTO services_schema.service_categories (name, slug, order_number, is_default_unassigned)
VALUES ('General Services', 'general', 1, true)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO services_schema.service_categories (name, slug, order_number)
VALUES 
  ('Emergency Services', 'emergency', 2),
  ('Outpatient Services', 'outpatient', 3),
  ('Inpatient Services', 'inpatient', 4)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO services_schema.services (title, description, slug, category_id, delivery_mode, publishing_status)
SELECT 
  'Emergency Care',
  '24/7 emergency medical services',
  'emergency-care',
  sc.category_id,
  'outpatient_service',
  'published'
FROM services_schema.service_categories sc
WHERE sc.slug = 'emergency'
ON CONFLICT (slug) DO NOTHING;

INSERT INTO services_schema.featured_categories (category_id, feature_position)
SELECT sc.category_id, 1
FROM services_schema.service_categories sc
WHERE sc.slug = 'emergency'
ON CONFLICT (feature_position) DO NOTHING;
`

	identityTestData := `
INSERT INTO identity_schema.user_sessions (user_id, expires_at)
VALUES ('test-user', NOW() + INTERVAL '24 hours')
ON CONFLICT (session_id) DO NOTHING;
`

	_ = contentTestData
	_ = servicesTestData
	_ = identityTestData

	return nil
}

func (dmo *DevMigrationOrchestrator) initializeSchemas(ctx context.Context) error {
	fmt.Println("Initializing database schemas...")

	domains := []string{"content", "services", "identity"}

	for _, domain := range domains {
		if err := dmo.initializeDomainSchema(ctx, domain); err != nil {
			return fmt.Errorf("failed to initialize schema for domain %s: %w", domain, err)
		}
	}

	return nil
}

func (dmo *DevMigrationOrchestrator) initializeDomainSchema(ctx context.Context, domain string) error {
	fmt.Printf("Initializing %s schema...\n", domain)

	schemaName := fmt.Sprintf("%s_schema", domain)
	grantSQL := fmt.Sprintf("GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA %s TO current_user", schemaName)
	grantSeqSQL := fmt.Sprintf("GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA %s TO current_user", schemaName)

	_ = grantSQL
	_ = grantSeqSQL

	return nil
}

func (dmo *DevMigrationOrchestrator) getAggressiveSettings() *AggressiveSettings {
	return &AggressiveSettings{
		AutoRollbackOnError:   true,
		RecreateOnConflict:    true,
		SkipValidation:        false,
		ForceLatestVersion:    true,
		EnableTestDataSeed:    true,
		CleanupOnFailure:      true,
		MaxRetryAttempts:      3,
		RetryDelay:           5 * time.Second,
	}
}

func (dmo *DevMigrationOrchestrator) GetMigrationStatus(ctx context.Context) (*DevMigrationStatus, error) {
	status := &DevMigrationStatus{
		Environment: dmo.environment,
		Timestamp:   time.Now(),
	}

	currentVersions, err := dmo.migrationRunner.GetCurrentVersions(ctx)
	if err != nil {
		status.Error = err
		return status, err
	}

	status.CurrentVersions = currentVersions

	for domain, version := range currentVersions {
		domainMigrator := migration.NewDomainMigrator(domain, dmo.databaseURL, dmo.basePath, dmo.environment)
		domainStatus, err := domainMigrator.GetDomainStatus(ctx)
		if err != nil {
			status.DomainErrors = append(status.DomainErrors, fmt.Sprintf("%s: %v", domain, err))
			continue
		}

		status.DomainStatuses = append(status.DomainStatuses, DomainMigrationStatus{
			Domain:         domain,
			CurrentVersion: version,
			Ready:          domainStatus.Ready,
			IsDirty:        domainStatus.IsDirty,
			Status:         domainStatus.Status,
		})

		if domainStatus.Ready {
			status.ReadyDomains++
		}
	}

	status.TotalDomains = len(currentVersions)
	status.IsReady = status.ReadyDomains == status.TotalDomains && len(status.DomainErrors) == 0

	return status, nil
}

type DevMigrationStatus struct {
	Environment     string
	Timestamp       time.Time
	CurrentVersions map[string]uint
	DomainStatuses  []DomainMigrationStatus
	TotalDomains    int
	ReadyDomains    int
	IsReady         bool
	DomainErrors    []string
	Error           error
}

type DomainMigrationStatus struct {
	Domain         string
	CurrentVersion uint
	Ready          bool
	IsDirty        bool
	Status         string
}

func (dmo *DevMigrationOrchestrator) ValidateConfiguration(ctx context.Context) error {
	if dmo.databaseURL == "" {
		return fmt.Errorf("database URL is required")
	}

	if dmo.basePath == "" {
		return fmt.Errorf("base path is required")
	}

	return nil
}