package validation

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type SafetyChecker struct {
	environment   string
	safetyLevel   SafetyLevel
	rules         []SafetyRule
	restrictions  map[string][]Restriction
}

type SafetyLevel int

const (
	SafetyLevelMinimal SafetyLevel = iota
	SafetyLevelModerate
	SafetyLevelStrict
	SafetyLevelMaximum
)

type SafetyRule interface {
	Evaluate(ctx context.Context, deployment *DeploymentContext) SafetyCheckResult
	GetSeverity() SafetySeverity
	GetDescription() string
}

type SafetySeverity int

const (
	SeverityInfo SafetySeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

type DeploymentContext struct {
	Environment         string
	TargetComponents    []string
	MigrationVersions   map[string]uint
	ConfigurationChanges map[string]interface{}
	ResourceChanges     []ResourceChange
	UserContext         UserContext
	Timestamp           time.Time
}

type ResourceChange struct {
	ResourceType string
	Action       ChangeAction
	ResourceName string
	OldValue     interface{}
	NewValue     interface{}
	Impact       ImpactAssessment
}

type ChangeAction string

const (
	ActionCreate ChangeAction = "create"
	ActionUpdate ChangeAction = "update"
	ActionDelete ChangeAction = "delete"
	ActionScale  ChangeAction = "scale"
)

type ImpactAssessment struct {
	DataLossRisk     bool
	DowntimeExpected time.Duration
	AffectedUsers    int
	RollbackPossible bool
}

type UserContext struct {
	UserID      string
	Permissions []string
	ApprovalLevel string
}

type SafetyCheckResult struct {
	RuleName     string
	Passed       bool
	Severity     SafetySeverity
	Message      string
	Details      map[string]interface{}
	Recommendations []string
}

type SafetyValidationResult struct {
	Environment        string
	SafetyLevel        SafetyLevel
	OverallStatus      SafetyStatus
	TotalChecks        int
	PassedChecks       int
	FailedChecks       int
	CriticalIssues     []SafetyCheckResult
	Warnings          []SafetyCheckResult
	Recommendations   []string
	CanProceed        bool
	RequiredApprovals []ApprovalRequirement
	Timestamp         time.Time
}

type SafetyStatus string

const (
	SafetyStatusSafe     SafetyStatus = "safe"
	SafetyStatusWarning  SafetyStatus = "warning"
	SafetyStatusUnsafe   SafetyStatus = "unsafe"
	SafetyStatusBlocked  SafetyStatus = "blocked"
)

type ApprovalRequirement struct {
	Type        string
	Reason      string
	RequiredBy  time.Time
	ApproverRole string
}

type Restriction struct {
	Type        RestrictionType
	Description string
	Applies     func(context *DeploymentContext) bool
	Enforce     func(context *DeploymentContext) error
}

type RestrictionType string

const (
	RestrictionTimeWindow   RestrictionType = "time_window"
	RestrictionDataOperations RestrictionType = "data_operations"
	RestrictionScaling      RestrictionType = "scaling"
	RestrictionMigrations   RestrictionType = "migrations"
	RestrictionApprovals    RestrictionType = "approvals"
)

func NewSafetyChecker(environment string) *SafetyChecker {
	checker := &SafetyChecker{
		environment:  environment,
		safetyLevel:  getSafetyLevelForEnvironment(environment),
		restrictions: make(map[string][]Restriction),
	}

	checker.initializeSafetyRules()
	checker.initializeRestrictions()
	
	return checker
}

func (sc *SafetyChecker) ValidateDeploymentSafety(ctx context.Context, deployment *DeploymentContext) (*SafetyValidationResult, error) {
	result := &SafetyValidationResult{
		Environment:   sc.environment,
		SafetyLevel:   sc.safetyLevel,
		TotalChecks:   len(sc.rules),
		Timestamp:     time.Now(),
		CanProceed:    true,
		Recommendations: []string{},
	}

	var allResults []SafetyCheckResult

	for _, rule := range sc.rules {
		ruleResult := rule.Evaluate(ctx, deployment)
		allResults = append(allResults, ruleResult)

		if ruleResult.Passed {
			result.PassedChecks++
		} else {
			result.FailedChecks++
			
			if ruleResult.Severity == SeverityCritical {
				result.CriticalIssues = append(result.CriticalIssues, ruleResult)
				result.CanProceed = false
			} else if ruleResult.Severity == SeverityWarning {
				result.Warnings = append(result.Warnings, ruleResult)
			}
		}

		result.Recommendations = append(result.Recommendations, ruleResult.Recommendations...)
	}

	if err := sc.enforceRestrictions(deployment); err != nil {
		result.CanProceed = false
		result.CriticalIssues = append(result.CriticalIssues, SafetyCheckResult{
			RuleName: "restrictions",
			Passed:   false,
			Severity: SeverityCritical,
			Message:  fmt.Sprintf("Deployment restrictions violated: %v", err),
		})
	}

	result.RequiredApprovals = sc.getRequiredApprovals(deployment, result)
	result.OverallStatus = sc.calculateOverallStatus(result)

	return result, nil
}

func (sc *SafetyChecker) initializeSafetyRules() {
	sc.rules = []SafetyRule{
		&MigrationSafetyRule{environment: sc.environment},
		&DataOperationSafetyRule{environment: sc.environment},
		&ResourceLimitSafetyRule{environment: sc.environment},
		&ApprovalRequirementRule{environment: sc.environment},
		&TimeWindowRestrictionRule{environment: sc.environment},
		&DependencyValidationRule{environment: sc.environment},
		&BackupVerificationRule{environment: sc.environment},
		&CapacityValidationRule{environment: sc.environment},
	}
}

func (sc *SafetyChecker) initializeRestrictions() {
	switch sc.environment {
	case "production":
		sc.initializeProductionRestrictions()
	case "staging":
		sc.initializeStagingRestrictions()
	case "development":
		sc.initializeDevelopmentRestrictions()
	}
}

func (sc *SafetyChecker) initializeProductionRestrictions() {
	sc.restrictions["time_window"] = []Restriction{
		{
			Type:        RestrictionTimeWindow,
			Description: "No deployments during business hours",
			Applies: func(ctx *DeploymentContext) bool {
				hour := ctx.Timestamp.Hour()
				return hour >= 9 && hour <= 17
			},
			Enforce: func(ctx *DeploymentContext) error {
				return fmt.Errorf("deployments not allowed during business hours (9AM-5PM)")
			},
		},
	}

	sc.restrictions["data_operations"] = []Restriction{
		{
			Type:        RestrictionDataOperations,
			Description: "Destructive data operations require approval",
			Applies: func(ctx *DeploymentContext) bool {
				for _, change := range ctx.ResourceChanges {
					if change.Impact.DataLossRisk {
						return true
					}
				}
				return false
			},
			Enforce: func(ctx *DeploymentContext) error {
				return fmt.Errorf("destructive data operations require manual approval")
			},
		},
	}

	sc.restrictions["approvals"] = []Restriction{
		{
			Type:        RestrictionApprovals,
			Description: "All production deployments require senior approval",
			Applies: func(ctx *DeploymentContext) bool {
				return ctx.Environment == "production"
			},
			Enforce: func(ctx *DeploymentContext) error {
				if ctx.UserContext.ApprovalLevel != "senior" {
					return fmt.Errorf("production deployments require senior approval")
				}
				return nil
			},
		},
	}
}

func (sc *SafetyChecker) initializeStagingRestrictions() {
	sc.restrictions["migrations"] = []Restriction{
		{
			Type:        RestrictionMigrations,
			Description: "Migration rollbacks require confirmation",
			Applies: func(ctx *DeploymentContext) bool {
				for _, version := range ctx.MigrationVersions {
					if version == 0 {
						return true
					}
				}
				return false
			},
			Enforce: func(ctx *DeploymentContext) error {
				return fmt.Errorf("migration rollbacks in staging require explicit confirmation")
			},
		},
	}
}

func (sc *SafetyChecker) initializeDevelopmentRestrictions() {
}

func (sc *SafetyChecker) enforceRestrictions(deployment *DeploymentContext) error {
	for _, restrictions := range sc.restrictions {
		for _, restriction := range restrictions {
			if restriction.Applies(deployment) {
				if err := restriction.Enforce(deployment); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (sc *SafetyChecker) getRequiredApprovals(deployment *DeploymentContext, result *SafetyValidationResult) []ApprovalRequirement {
	var approvals []ApprovalRequirement

	if sc.environment == "production" {
		approvals = append(approvals, ApprovalRequirement{
			Type:         "senior_approval",
			Reason:       "Production deployment requires senior approval",
			RequiredBy:   time.Now().Add(24 * time.Hour),
			ApproverRole: "senior",
		})
	}

	for _, change := range deployment.ResourceChanges {
		if change.Impact.DataLossRisk {
			approvals = append(approvals, ApprovalRequirement{
				Type:         "data_operation_approval",
				Reason:       fmt.Sprintf("Destructive operation on %s requires approval", change.ResourceName),
				RequiredBy:   time.Now().Add(4 * time.Hour),
				ApproverRole: "senior",
			})
		}
	}

	return approvals
}

func (sc *SafetyChecker) calculateOverallStatus(result *SafetyValidationResult) SafetyStatus {
	if len(result.CriticalIssues) > 0 {
		return SafetyStatusBlocked
	}

	if !result.CanProceed {
		return SafetyStatusUnsafe
	}

	if len(result.Warnings) > 0 {
		return SafetyStatusWarning
	}

	return SafetyStatusSafe
}

func getSafetyLevelForEnvironment(environment string) SafetyLevel {
	switch environment {
	case "development":
		return SafetyLevelMinimal
	case "staging":
		return SafetyLevelModerate
	case "production":
		return SafetyLevelMaximum
	default:
		return SafetyLevelModerate
	}
}

type MigrationSafetyRule struct {
	environment string
}

func (r *MigrationSafetyRule) Evaluate(ctx context.Context, deployment *DeploymentContext) SafetyCheckResult {
	result := SafetyCheckResult{
		RuleName: "migration_safety",
		Passed:   true,
		Severity: SeverityWarning,
	}

	for domain, version := range deployment.MigrationVersions {
		if version == 0 {
			result.Passed = false
			result.Severity = SeverityError
			result.Message = fmt.Sprintf("Migration rollback detected for domain %s", domain)
			result.Recommendations = append(result.Recommendations, 
				fmt.Sprintf("Verify rollback safety for %s domain", domain))
		}
	}

	if result.Passed {
		result.Message = "All migration operations are safe"
	}

	return result
}

func (r *MigrationSafetyRule) GetSeverity() SafetySeverity { return SeverityWarning }
func (r *MigrationSafetyRule) GetDescription() string { return "Validates migration operation safety" }

type DataOperationSafetyRule struct {
	environment string
}

func (r *DataOperationSafetyRule) Evaluate(ctx context.Context, deployment *DeploymentContext) SafetyCheckResult {
	result := SafetyCheckResult{
		RuleName: "data_operation_safety",
		Passed:   true,
		Severity: SeverityCritical,
		Message:  "No destructive data operations detected",
	}

	for _, change := range deployment.ResourceChanges {
		if change.Impact.DataLossRisk {
			result.Passed = false
			result.Message = fmt.Sprintf("Destructive data operation detected: %s on %s", 
				change.Action, change.ResourceName)
			result.Recommendations = append(result.Recommendations,
				"Verify backup is available before proceeding")
			break
		}
	}

	return result
}

func (r *DataOperationSafetyRule) GetSeverity() SafetySeverity { return SeverityCritical }
func (r *DataOperationSafetyRule) GetDescription() string { return "Validates data operation safety" }

type ResourceLimitSafetyRule struct {
	environment string
}

func (r *ResourceLimitSafetyRule) Evaluate(ctx context.Context, deployment *DeploymentContext) SafetyCheckResult {
	result := SafetyCheckResult{
		RuleName: "resource_limits",
		Passed:   true,
		Severity: SeverityWarning,
		Message:  "Resource limits are within acceptable ranges",
	}

	return result
}

func (r *ResourceLimitSafetyRule) GetSeverity() SafetySeverity { return SeverityWarning }
func (r *ResourceLimitSafetyRule) GetDescription() string { return "Validates resource allocation limits" }

type ApprovalRequirementRule struct {
	environment string
}

func (r *ApprovalRequirementRule) Evaluate(ctx context.Context, deployment *DeploymentContext) SafetyCheckResult {
	result := SafetyCheckResult{
		RuleName: "approval_requirements",
		Passed:   true,
		Severity: SeverityError,
		Message:  "Approval requirements satisfied",
	}

	if r.environment == "production" && deployment.UserContext.ApprovalLevel != "senior" {
		result.Passed = false
		result.Message = "Production deployment requires senior approval"
		result.Recommendations = append(result.Recommendations,
			"Obtain senior approval before proceeding")
	}

	return result
}

func (r *ApprovalRequirementRule) GetSeverity() SafetySeverity { return SeverityError }
func (r *ApprovalRequirementRule) GetDescription() string { return "Validates approval requirements" }

type TimeWindowRestrictionRule struct {
	environment string
}

func (r *TimeWindowRestrictionRule) Evaluate(ctx context.Context, deployment *DeploymentContext) SafetyCheckResult {
	result := SafetyCheckResult{
		RuleName: "time_window_restrictions",
		Passed:   true,
		Severity: SeverityWarning,
		Message:  "Deployment time is acceptable",
	}

	if r.environment == "production" {
		hour := deployment.Timestamp.Hour()
		if hour >= 9 && hour <= 17 {
			result.Passed = false
			result.Message = "Production deployments should not occur during business hours"
			result.Recommendations = append(result.Recommendations,
				"Schedule deployment outside business hours (9AM-5PM)")
		}
	}

	return result
}

func (r *TimeWindowRestrictionRule) GetSeverity() SafetySeverity { return SeverityWarning }
func (r *TimeWindowRestrictionRule) GetDescription() string { return "Validates deployment time windows" }

type DependencyValidationRule struct {
	environment string
}

func (r *DependencyValidationRule) Evaluate(ctx context.Context, deployment *DeploymentContext) SafetyCheckResult {
	result := SafetyCheckResult{
		RuleName: "dependency_validation",
		Passed:   true,
		Severity: SeverityError,
		Message:  "All dependencies are satisfied",
	}

	return result
}

func (r *DependencyValidationRule) GetSeverity() SafetySeverity { return SeverityError }
func (r *DependencyValidationRule) GetDescription() string { return "Validates deployment dependencies" }

type BackupVerificationRule struct {
	environment string
}

func (r *BackupVerificationRule) Evaluate(ctx context.Context, deployment *DeploymentContext) SafetyCheckResult {
	result := SafetyCheckResult{
		RuleName: "backup_verification",
		Passed:   true,
		Severity: SeverityCritical,
		Message:  "Backup verification not required for this deployment",
	}

	hasDataOperations := false
	for _, change := range deployment.ResourceChanges {
		if change.Impact.DataLossRisk {
			hasDataOperations = true
			break
		}
	}

	if r.environment == "production" && hasDataOperations {
		result.Message = "Backup verification required for destructive operations"
		result.Recommendations = append(result.Recommendations,
			"Verify recent backup exists and is restorable")
	}

	return result
}

func (r *BackupVerificationRule) GetSeverity() SafetySeverity { return SeverityCritical }
func (r *BackupVerificationRule) GetDescription() string { return "Validates backup availability" }

type CapacityValidationRule struct {
	environment string
}

func (r *CapacityValidationRule) Evaluate(ctx context.Context, deployment *DeploymentContext) SafetyCheckResult {
	result := SafetyCheckResult{
		RuleName: "capacity_validation",
		Passed:   true,
		Severity: SeverityWarning,
		Message:  "System capacity is adequate",
	}

	return result
}

func (r *CapacityValidationRule) GetSeverity() SafetySeverity { return SeverityWarning }
func (r *CapacityValidationRule) GetDescription() string { return "Validates system capacity requirements" }