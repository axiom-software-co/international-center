package migration

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
)

type RollbackManager struct {
	databaseURL string
	environment string
	basePath    string
}

type RollbackPlan struct {
	Environment     string
	RequestedBy     string
	RequestTime     time.Time
	TargetVersions  map[string]uint
	RollbackReason  string
	ApprovalStatus  ApprovalStatus
	EstimatedTime   time.Duration
	DataLossRisk    RiskLevel
	Dependencies    []RollbackDependency
}

type RollbackResult struct {
	Success           bool
	Error             error
	StartTime         time.Time
	EndTime           time.Time
	RolledBackDomains map[string]uint
	FailedDomains     []string
	DataLoss          bool
	RecoverySteps     []string
}

type ApprovalStatus int

const (
	ApprovalPending ApprovalStatus = iota
	ApprovalGranted
	ApprovalDenied
	ApprovalExpired
)

type RiskLevel int

const (
	RiskLow RiskLevel = iota
	RiskModerate
	RiskHigh
	RiskCritical
)

type RollbackDependency struct {
	Domain        string
	RequiredBy    string
	Reason        string
	CanProceed    bool
}

type RollbackSnapshot struct {
	Domain      string
	Version     uint
	Timestamp   time.Time
	DataHash    string
	BackupPath  string
	Size        int64
}

func NewRollbackManager(databaseURL, basePath, environment string) *RollbackManager {
	return &RollbackManager{
		databaseURL: databaseURL,
		environment: environment,
		basePath:    basePath,
	}
}

func (rm *RollbackManager) CreateRollbackPlan(ctx context.Context, targetVersions map[string]uint, reason string, requestedBy string) (*RollbackPlan, error) {
	plan := &RollbackPlan{
		Environment:    rm.environment,
		RequestedBy:    requestedBy,
		RequestTime:    time.Now(),
		TargetVersions: targetVersions,
		RollbackReason: reason,
		ApprovalStatus: rm.getInitialApprovalStatus(),
	}

	dependencies, err := rm.analyzeDependencies(ctx, targetVersions)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze dependencies: %w", err)
	}
	plan.Dependencies = dependencies

	plan.DataLossRisk = rm.assessDataLossRisk(ctx, targetVersions)
	plan.EstimatedTime = rm.calculateEstimatedTime(targetVersions)

	if !rm.canProceedWithRollback(plan) {
		return nil, fmt.Errorf("rollback cannot proceed due to dependencies or risk level")
	}

	return plan, nil
}

func (rm *RollbackManager) ExecuteRollback(ctx context.Context, plan *RollbackPlan) (*RollbackResult, error) {
	if plan.ApprovalStatus != ApprovalGranted && rm.requiresApproval() {
		return nil, fmt.Errorf("rollback requires approval but status is %v", plan.ApprovalStatus)
	}

	result := &RollbackResult{
		StartTime:         time.Now(),
		RolledBackDomains: make(map[string]uint),
		Success:           false,
	}

	snapshots, err := rm.createPreRollbackSnapshots(ctx, plan.TargetVersions)
	if err != nil {
		result.Error = fmt.Errorf("failed to create snapshots: %w", err)
		result.EndTime = time.Now()
		return result, result.Error
	}

	orderedDomains := rm.getRollbackOrder(plan.TargetVersions, plan.Dependencies)

	for _, domain := range orderedDomains {
		targetVersion, exists := plan.TargetVersions[domain]
		if !exists {
			continue
		}

		err := rm.rollbackDomain(ctx, domain, targetVersion)
		if err != nil {
			result.FailedDomains = append(result.FailedDomains, domain)
			result.Error = fmt.Errorf("failed to rollback domain %s: %w", domain, err)

			if rm.shouldStopOnError() {
				break
			}
		} else {
			result.RolledBackDomains[domain] = targetVersion
		}
	}

	result.Success = len(result.FailedDomains) == 0
	result.EndTime = time.Now()

	if result.Success {
		if err := rm.cleanupSnapshots(ctx, snapshots); err != nil {
			fmt.Printf("Warning: failed to cleanup snapshots: %v\n", err)
		}
	} else {
		result.RecoverySteps = rm.generateRecoverySteps(result.FailedDomains, snapshots)
	}

	return result, result.Error
}

func (rm *RollbackManager) ValidateRollbackSafety(ctx context.Context, targetVersions map[string]uint) error {
	for domain, targetVersion := range targetVersions {
		if err := rm.validateDomainRollback(ctx, domain, targetVersion); err != nil {
			return fmt.Errorf("rollback validation failed for domain %s: %w", domain, err)
		}
	}
	return nil
}

func (rm *RollbackManager) GetRollbackHistory(ctx context.Context, domain string, limit int) ([]RollbackRecord, error) {
	db, err := sql.Open("postgres", rm.databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	query := `
		SELECT rollback_id, domain, from_version, to_version, executed_at, executed_by, reason, success
		FROM rollback_history 
		WHERE domain = $1 
		ORDER BY executed_at DESC 
		LIMIT $2`

	rows, err := db.QueryContext(ctx, query, domain, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query rollback history: %w", err)
	}
	defer rows.Close()

	var records []RollbackRecord
	for rows.Next() {
		var record RollbackRecord
		err := rows.Scan(
			&record.RollbackID,
			&record.Domain,
			&record.FromVersion,
			&record.ToVersion,
			&record.ExecutedAt,
			&record.ExecutedBy,
			&record.Reason,
			&record.Success,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rollback record: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}

type RollbackRecord struct {
	RollbackID  string
	Domain      string
	FromVersion uint
	ToVersion   uint
	ExecutedAt  time.Time
	ExecutedBy  string
	Reason      string
	Success     bool
}

func (rm *RollbackManager) rollbackDomain(ctx context.Context, domain string, targetVersion uint) error {
	db, err := sql.Open("postgres", rm.databaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{
		DatabaseName: fmt.Sprintf("%s_schema", domain),
		SchemaName:   fmt.Sprintf("%s_schema", domain),
	})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrationsPath := fmt.Sprintf("%s/../backend/internal/%s/migrations", rm.basePath, domain)
	fileSource, err := (&file.File{}).Open(fmt.Sprintf("file://%s", migrationsPath))
	if err != nil {
		return fmt.Errorf("failed to open migration files: %w", err)
	}
	defer fileSource.Close()

	m, err := migrate.NewWithInstance("file", fileSource, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Migrate(targetVersion); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback to version %d: %w", targetVersion, err)
	}

	if err := rm.recordRollback(ctx, domain, targetVersion); err != nil {
		fmt.Printf("Warning: failed to record rollback: %v\n", err)
	}

	return nil
}

func (rm *RollbackManager) recordRollback(ctx context.Context, domain string, targetVersion uint) error {
	return nil
}

func (rm *RollbackManager) analyzeDependencies(ctx context.Context, targetVersions map[string]uint) ([]RollbackDependency, error) {
	var dependencies []RollbackDependency

	domainDeps := map[string][]string{
		"content":  {},
		"services": {"content"},
	}

	for domain, requiredDomains := range domainDeps {
		targetVersion, rollbackRequested := targetVersions[domain]
		if !rollbackRequested {
			continue
		}

		for _, requiredDomain := range requiredDomains {
			dependency := RollbackDependency{
				Domain:     requiredDomain,
				RequiredBy: domain,
				Reason:     fmt.Sprintf("Domain %s depends on %s", domain, requiredDomain),
				CanProceed: true,
			}

			if requiredTargetVersion, requiredRollback := targetVersions[requiredDomain]; requiredRollback {
				if requiredTargetVersion > targetVersion {
					dependency.CanProceed = false
					dependency.Reason = fmt.Sprintf("Required domain %s target version (%d) is higher than dependent domain %s target version (%d)", 
						requiredDomain, requiredTargetVersion, domain, targetVersion)
				}
			}

			dependencies = append(dependencies, dependency)
		}
	}

	return dependencies, nil
}

func (rm *RollbackManager) assessDataLossRisk(ctx context.Context, targetVersions map[string]uint) RiskLevel {
	highestRisk := RiskLow

	for domain, targetVersion := range targetVersions {
		currentVersion, err := rm.getCurrentVersion(ctx, domain)
		if err != nil {
			return RiskCritical
		}

		versionDiff := currentVersion - targetVersion
		
		risk := RiskLow
		if versionDiff > 10 {
			risk = RiskCritical
		} else if versionDiff > 5 {
			risk = RiskHigh
		} else if versionDiff > 2 {
			risk = RiskModerate
		}

		if risk > highestRisk {
			highestRisk = risk
		}
	}

	return highestRisk
}

func (rm *RollbackManager) getCurrentVersion(ctx context.Context, domain string) (uint, error) {
	db, err := sql.Open("postgres", rm.databaseURL)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var version sql.NullInt64
	schemaName := fmt.Sprintf("%s_schema", domain)
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

func (rm *RollbackManager) calculateEstimatedTime(targetVersions map[string]uint) time.Duration {
	baseDuration := 30 * time.Second
	return time.Duration(len(targetVersions)) * baseDuration
}

func (rm *RollbackManager) getInitialApprovalStatus() ApprovalStatus {
	switch rm.environment {
	case "development":
		return ApprovalGranted
	case "staging":
		return ApprovalPending
	case "production":
		return ApprovalPending
	default:
		return ApprovalPending
	}
}

func (rm *RollbackManager) requiresApproval() bool {
	return rm.environment != "development"
}

func (rm *RollbackManager) canProceedWithRollback(plan *RollbackPlan) bool {
	for _, dep := range plan.Dependencies {
		if !dep.CanProceed {
			return false
		}
	}

	if plan.DataLossRisk == RiskCritical && rm.environment == "production" {
		return false
	}

	return true
}

func (rm *RollbackManager) shouldStopOnError() bool {
	return rm.environment != "development"
}

func (rm *RollbackManager) getRollbackOrder(targetVersions map[string]uint, dependencies []RollbackDependency) []string {
	var ordered []string

	for domain := range targetVersions {
		if domain == "services" {
			ordered = append(ordered, domain)
		}
	}
	
	for domain := range targetVersions {
		if domain == "content" {
			ordered = append(ordered, domain)
		}
	}

	return ordered
}

func (rm *RollbackManager) createPreRollbackSnapshots(ctx context.Context, targetVersions map[string]uint) ([]RollbackSnapshot, error) {
	var snapshots []RollbackSnapshot

	for domain := range targetVersions {
		currentVersion, err := rm.getCurrentVersion(ctx, domain)
		if err != nil {
			return nil, fmt.Errorf("failed to get current version for domain %s: %w", domain, err)
		}

		snapshot := RollbackSnapshot{
			Domain:    domain,
			Version:   currentVersion,
			Timestamp: time.Now(),
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

func (rm *RollbackManager) cleanupSnapshots(ctx context.Context, snapshots []RollbackSnapshot) error {
	return nil
}

func (rm *RollbackManager) generateRecoverySteps(failedDomains []string, snapshots []RollbackSnapshot) []string {
	steps := []string{
		"Review rollback error logs",
		"Verify database connectivity",
		"Check migration file integrity",
	}

	for _, domain := range failedDomains {
		steps = append(steps, fmt.Sprintf("Manually review %s domain migration state", domain))
	}

	steps = append(steps, "Consider restoring from backup if necessary")
	return steps
}

func (rm *RollbackManager) validateDomainRollback(ctx context.Context, domain string, targetVersion uint) error {
	currentVersion, err := rm.getCurrentVersion(ctx, domain)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if targetVersion >= currentVersion {
		return fmt.Errorf("target version %d is not less than current version %d", targetVersion, currentVersion)
	}

	return nil
}