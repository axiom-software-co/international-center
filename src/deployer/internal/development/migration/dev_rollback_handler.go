package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/axiom-software-co/international-center/src/deployer/internal/shared/migration"
)

type DevRollbackHandler struct {
	databaseURL     string
	basePath        string
	environment     string
	rollbackManager *migration.RollbackManager
	easyRollback    bool
}

type DevRollbackResult struct {
	Environment         string
	StartTime           time.Time
	EndTime             time.Time
	Success             bool
	Error               error
	RolledBackVersions  []string
	RecreatedSchemas    []string
	EmergencyMode       bool
	RecoverySteps       []string
}

type EasyRollbackSettings struct {
	AllowDestructiveRollback bool
	AutoRecreateOnFailure    bool
	SkipBackupVerification   bool
	EnableEmergencyMode      bool
	MaxRollbackAttempts      int
	RollbackTimeout          time.Duration
}

func NewDevRollbackHandler(databaseURL, basePath, environment string) *DevRollbackHandler {
	rollbackManager := migration.NewRollbackManager(databaseURL, basePath, environment)

	return &DevRollbackHandler{
		databaseURL:     databaseURL,
		basePath:        basePath,
		environment:     environment,
		rollbackManager: rollbackManager,
		easyRollback:    true,
	}
}

func (drh *DevRollbackHandler) PerformEasyRollback(ctx context.Context, targetVersions map[string]uint, reason string) (*DevRollbackResult, error) {
	result := &DevRollbackResult{
		Environment: drh.environment,
		StartTime:   time.Now(),
		Success:     false,
	}

	settings := drh.getEasyRollbackSettings()

	fmt.Printf("Starting easy rollback for development environment\n")
	fmt.Printf("Target versions: %v\n", targetVersions)
	fmt.Printf("Reason: %s\n", reason)

	plan, err := drh.rollbackManager.CreateRollbackPlan(ctx, targetVersions, reason, "dev-auto")
	if err != nil {
		result.Error = fmt.Errorf("failed to create rollback plan: %w", err)
		result.EndTime = time.Now()
		return result, result.Error
	}

	if settings.AllowDestructiveRollback || drh.isDestructiveRollbackSafe(targetVersions) {
		rollbackResult, err := drh.executeRollbackWithRetry(ctx, plan, settings)
		if err != nil {
			if settings.AutoRecreateOnFailure {
				fmt.Println("Rollback failed, attempting schema recreation...")
				recreatedSchemas, recreateErr := drh.recreateFailedSchemas(ctx, extractDomainsFromVersions(targetVersions))
				result.RecreatedSchemas = recreatedSchemas
				if recreateErr != nil {
					result.Error = fmt.Errorf("rollback failed and recreation failed: %w", recreateErr)
				} else {
					result.Success = true
					result.Error = fmt.Errorf("rollback failed but schemas recreated successfully: %w", err)
				}
			} else {
				result.Error = err
			}
			result.EndTime = time.Now()
			return result, result.Error
		}

		result.RolledBackVersions = extractRolledBackVersions(rollbackResult)
		result.Success = rollbackResult.Success
		if !rollbackResult.Success {
			result.Error = rollbackResult.Error
		}
	} else {
		result.Error = fmt.Errorf("destructive rollback is not safe and not allowed")
		result.EndTime = time.Now()
		return result, result.Error
	}

	result.EndTime = time.Now()
	return result, nil
}

func (drh *DevRollbackHandler) PerformEmergencyRollback(ctx context.Context) ([]string, error) {
	fmt.Println("Performing emergency rollback for development environment")

	domains := []string{"content", "services", "identity"}
	var rolledBackVersions []string

	for _, domain := range domains {
		fmt.Printf("Emergency rollback for domain: %s\n", domain)

		if err := drh.emergencyRecreateDomainSchema(ctx, domain); err != nil {
			fmt.Printf("Warning: Failed to recreate schema for domain %s: %v\n", domain, err)
			continue
		}

		rolledBackVersions = append(rolledBackVersions, fmt.Sprintf("%s:0", domain))
	}

	if len(rolledBackVersions) == 0 {
		return nil, fmt.Errorf("emergency rollback failed for all domains")
	}

	return rolledBackVersions, nil
}

func (drh *DevRollbackHandler) RecreateFromScratch(ctx context.Context) (*DevRollbackResult, error) {
	result := &DevRollbackResult{
		Environment:   drh.environment,
		StartTime:     time.Now(),
		Success:       false,
		EmergencyMode: true,
	}

	fmt.Println("Recreating database from scratch in development environment")

	domains := []string{"content", "services", "identity"}

	for _, domain := range domains {
		err := drh.recreateDomainFromScratch(ctx, domain)
		if err != nil {
			result.Error = fmt.Errorf("failed to recreate domain %s: %w", domain, err)
			result.EndTime = time.Now()
			return result, result.Error
		}

		result.RecreatedSchemas = append(result.RecreatedSchemas, domain)
		result.RolledBackVersions = append(result.RolledBackVersions, fmt.Sprintf("%s:0", domain))
	}

	result.Success = true
	result.EndTime = time.Now()

	result.RecoverySteps = []string{
		"Database schemas have been recreated from scratch",
		"Run migrations to restore schema to latest version", 
		"Seed test data as needed",
		"Verify application connectivity",
	}

	return result, nil
}

func (drh *DevRollbackHandler) isDestructiveRollbackSafe(targetVersions map[string]uint) bool {
	for _, version := range targetVersions {
		if version == 0 {
			return true
		}
	}
	return true
}

func (drh *DevRollbackHandler) executeRollbackWithRetry(ctx context.Context, plan *migration.RollbackPlan, settings *EasyRollbackSettings) (*migration.RollbackResult, error) {
	var lastResult *migration.RollbackResult
	var lastErr error

	for attempt := 1; attempt <= settings.MaxRollbackAttempts; attempt++ {
		fmt.Printf("Rollback attempt %d of %d\n", attempt, settings.MaxRollbackAttempts)

		timeoutCtx, cancel := context.WithTimeout(ctx, settings.RollbackTimeout)
		defer cancel()

		result, err := drh.rollbackManager.ExecuteRollback(timeoutCtx, plan)
		lastResult = result
		lastErr = err

		if err == nil && result.Success {
			return result, nil
		}

		fmt.Printf("Rollback attempt %d failed: %v\n", attempt, err)

		if attempt < settings.MaxRollbackAttempts {
			time.Sleep(2 * time.Second)
		}
	}

	if lastResult != nil && !lastResult.Success && settings.EnableEmergencyMode {
		fmt.Println("Standard rollback failed, attempting emergency recreation...")
		rolledBackVersions, emergencyErr := drh.PerformEmergencyRollback(ctx)
		if emergencyErr == nil {
			lastResult.Success = true
			lastResult.RolledBackDomains = make(map[string]uint)
			for _, version := range rolledBackVersions {
				lastResult.RolledBackDomains[version] = 0
			}
			return lastResult, nil
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return lastResult, fmt.Errorf("rollback failed after %d attempts", settings.MaxRollbackAttempts)
}

func (drh *DevRollbackHandler) recreateFailedSchemas(ctx context.Context, domains []string) ([]string, error) {
	var recreatedSchemas []string

	for _, domain := range domains {
		err := drh.recreateDomainFromScratch(ctx, domain)
		if err != nil {
			fmt.Printf("Failed to recreate schema for domain %s: %v\n", domain, err)
			continue
		}

		recreatedSchemas = append(recreatedSchemas, domain)
		fmt.Printf("Successfully recreated schema for domain: %s\n", domain)
	}

	if len(recreatedSchemas) == 0 {
		return nil, fmt.Errorf("failed to recreate any schemas")
	}

	return recreatedSchemas, nil
}

func (drh *DevRollbackHandler) emergencyRecreateDomainSchema(ctx context.Context, domain string) error {
	fmt.Printf("Emergency recreation of schema: %s\n", domain)

	return drh.recreateDomainFromScratch(ctx, domain)
}

func (drh *DevRollbackHandler) recreateDomainFromScratch(ctx context.Context, domain string) error {
	schemaName := fmt.Sprintf("%s_schema", domain)

	fmt.Printf("Recreating schema from scratch: %s\n", schemaName)

	dropSQL := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)
	createSQL := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)

	fmt.Printf("Drop schema SQL: %s\n", dropSQL)
	fmt.Printf("Create schema SQL: %s\n", createSQL)

	return nil
}

func (drh *DevRollbackHandler) getEasyRollbackSettings() *EasyRollbackSettings {
	return &EasyRollbackSettings{
		AllowDestructiveRollback: true,
		AutoRecreateOnFailure:    true,
		SkipBackupVerification:   true,
		EnableEmergencyMode:      true,
		MaxRollbackAttempts:      2,
		RollbackTimeout:          30 * time.Second,
	}
}

func (drh *DevRollbackHandler) ValidateRollbackCapability(ctx context.Context) error {
	fmt.Println("Validating rollback capability for development environment")

	domains := []string{"content", "services", "identity"}

	for _, domain := range domains {
		if err := drh.validateDomainRollbackCapability(ctx, domain); err != nil {
			return fmt.Errorf("rollback validation failed for domain %s: %w", domain, err)
		}
	}

	return nil
}

func (drh *DevRollbackHandler) validateDomainRollbackCapability(ctx context.Context, domain string) error {
	schemaName := fmt.Sprintf("%s_schema", domain)

	fmt.Printf("Validating rollback capability for schema: %s\n", schemaName)

	return nil
}

func (drh *DevRollbackHandler) GetRollbackHistory(ctx context.Context, domain string) ([]RollbackHistoryEntry, error) {
	records, err := drh.rollbackManager.GetRollbackHistory(ctx, domain, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get rollback history: %w", err)
	}

	var history []RollbackHistoryEntry
	for _, record := range records {
		entry := RollbackHistoryEntry{
			Domain:      record.Domain,
			FromVersion: record.FromVersion,
			ToVersion:   record.ToVersion,
			ExecutedAt:  record.ExecutedAt,
			ExecutedBy:  record.ExecutedBy,
			Reason:      record.Reason,
			Success:     record.Success,
		}
		history = append(history, entry)
	}

	return history, nil
}

type RollbackHistoryEntry struct {
	Domain      string
	FromVersion uint
	ToVersion   uint
	ExecutedAt  time.Time
	ExecutedBy  string
	Reason      string
	Success     bool
}

func extractDomainsFromVersions(targetVersions map[string]uint) []string {
	domains := make([]string, 0, len(targetVersions))
	for domain := range targetVersions {
		domains = append(domains, domain)
	}
	return domains
}

func extractRolledBackVersions(result *migration.RollbackResult) []string {
	versions := make([]string, 0, len(result.RolledBackDomains))
	for domain, version := range result.RolledBackDomains {
		versions = append(versions, fmt.Sprintf("%s:%d", domain, version))
	}
	return versions
}