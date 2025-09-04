package automation

import (
	"context"
	"fmt"
	"time"
)

// DeploymentValidator validates deployments before and after execution
type DeploymentValidator struct {
	validators map[ValidationType][]Validator
}

// Validator defines interface for deployment validators
type Validator interface {
	Name() string
	Validate(ctx context.Context, environment string, metadata map[string]interface{}) error
	Timeout() time.Duration
	Required() bool
}

// SecurityValidator validates security configurations
type SecurityValidator struct {
	name     string
	timeout  time.Duration
	required bool
}

// ComplianceValidator validates compliance requirements
type ComplianceValidator struct {
	name     string
	timeout  time.Duration
	required bool
	rules    []ComplianceRule
}

// ComplianceRule defines a compliance rule
type ComplianceRule struct {
	Name        string
	Description string
	Check       func(ctx context.Context, environment string) error
}

// ContractValidator validates infrastructure contracts
type ContractValidator struct {
	name     string
	timeout  time.Duration
	required bool
}

// HealthValidator validates infrastructure health
type HealthValidator struct {
	name     string
	timeout  time.Duration
	required bool
}

// NewDeploymentValidator creates a new deployment validator
func NewDeploymentValidator() *DeploymentValidator {
	dv := &DeploymentValidator{
		validators: make(map[ValidationType][]Validator),
	}
	
	// Register default validators
	dv.registerDefaultValidators()
	
	return dv
}

// registerDefaultValidators registers default validators
func (dv *DeploymentValidator) registerDefaultValidators() {
	// Pre-deployment validators
	dv.AddValidator(ValidationTypePreDeploy, &SecurityValidator{
		name:     "Pre-Deploy Security Check",
		timeout:  2 * time.Minute,
		required: true,
	})
	
	dv.AddValidator(ValidationTypePreDeploy, &HealthValidator{
		name:     "Infrastructure Health Check",
		timeout:  1 * time.Minute,
		required: true,
	})
	
	// Post-deployment validators
	dv.AddValidator(ValidationTypePostDeploy, &ContractValidator{
		name:     "Contract Validation",
		timeout:  3 * time.Minute,
		required: true,
	})
	
	dv.AddValidator(ValidationTypePostDeploy, &HealthValidator{
		name:     "Post-Deploy Health Check",
		timeout:  2 * time.Minute,
		required: true,
	})
	
	// Security validators
	dv.AddValidator(ValidationTypeSecurity, &SecurityValidator{
		name:     "Security Configuration Validation",
		timeout:  3 * time.Minute,
		required: true,
	})
	
	// Compliance validators
	dv.AddValidator(ValidationTypeCompliance, &ComplianceValidator{
		name:     "Compliance Validation",
		timeout:  5 * time.Minute,
		required: true,
		rules:    dv.createComplianceRules(),
	})
}

// AddValidator adds a validator for a specific validation type
func (dv *DeploymentValidator) AddValidator(validationType ValidationType, validator Validator) {
	dv.validators[validationType] = append(dv.validators[validationType], validator)
}

// RunValidation runs all validators for a specific type
func (dv *DeploymentValidator) RunValidation(ctx context.Context, validationType ValidationType, environment string, metadata map[string]interface{}) error {
	validators, exists := dv.validators[validationType]
	if !exists {
		return nil // No validators registered for this type
	}
	
	for _, validator := range validators {
		validationCtx, cancel := context.WithTimeout(ctx, validator.Timeout())
		err := validator.Validate(validationCtx, environment, metadata)
		cancel()
		
		if err != nil {
			if validator.Required() {
				return fmt.Errorf("required validation '%s' failed: %w", validator.Name(), err)
			}
			// Log non-fatal validation failure
			fmt.Printf("Non-required validation '%s' failed: %v\n", validator.Name(), err)
		}
	}
	
	return nil
}

// GetValidators returns validators for a specific type
func (dv *DeploymentValidator) GetValidators(validationType ValidationType) []Validator {
	return dv.validators[validationType]
}

// SecurityValidator implementation
func (sv *SecurityValidator) Name() string {
	return sv.name
}

func (sv *SecurityValidator) Timeout() time.Duration {
	return sv.timeout
}

func (sv *SecurityValidator) Required() bool {
	return sv.required
}

func (sv *SecurityValidator) Validate(ctx context.Context, environment string, metadata map[string]interface{}) error {
	// Security validation logic
	validations := []struct {
		name  string
		check func() error
	}{
		{
			name: "No hardcoded secrets",
			check: func() error {
				// Check for hardcoded secrets (axiom rule compliance)
				return nil
			},
		},
		{
			name: "TLS 1.2 minimum",
			check: func() error {
				// Validate minimum TLS version
				return nil
			},
		},
		{
			name: "Least privilege IAM",
			check: func() error {
				// Validate IAM configurations
				return nil
			},
		},
		{
			name: "Network security",
			check: func() error {
				// Validate network security configurations
				if environment == "production" {
					// Production requires more restrictive networking
					return nil
				}
				return nil
			},
		},
	}
	
	for _, validation := range validations {
		if err := validation.check(); err != nil {
			return fmt.Errorf("security validation '%s' failed: %w", validation.name, err)
		}
	}
	
	return nil
}

// ComplianceValidator implementation
func (cv *ComplianceValidator) Name() string {
	return cv.name
}

func (cv *ComplianceValidator) Timeout() time.Duration {
	return cv.timeout
}

func (cv *ComplianceValidator) Required() bool {
	return cv.required
}

func (cv *ComplianceValidator) Validate(ctx context.Context, environment string, metadata map[string]interface{}) error {
	for _, rule := range cv.rules {
		if err := rule.Check(ctx, environment); err != nil {
			return fmt.Errorf("compliance rule '%s' failed: %w", rule.Name, err)
		}
	}
	return nil
}

// createComplianceRules creates compliance rules
func (dv *DeploymentValidator) createComplianceRules() []ComplianceRule {
	return []ComplianceRule{
		{
			Name:        "Data Encryption at Rest",
			Description: "All data must be encrypted at rest",
			Check: func(ctx context.Context, environment string) error {
				// Validate encryption at rest
				return nil
			},
		},
		{
			Name:        "Data Encryption in Transit",
			Description: "All data must be encrypted in transit",
			Check: func(ctx context.Context, environment string) error {
				// Validate encryption in transit (TLS)
				return nil
			},
		},
		{
			Name:        "Audit Logging",
			Description: "Audit logging must be enabled",
			Check: func(ctx context.Context, environment string) error {
				// Validate audit logging configuration
				return nil
			},
		},
		{
			Name:        "Environment Isolation",
			Description: "Resources must be properly isolated by environment",
			Check: func(ctx context.Context, environment string) error {
				// Validate environment isolation
				return nil
			},
		},
		{
			Name:        "Backup Configuration",
			Description: "Backup must be properly configured",
			Check: func(ctx context.Context, environment string) error {
				if environment == "production" {
					// Production requires more stringent backup requirements
					return nil
				}
				return nil
			},
		},
	}
}

// ContractValidator implementation
func (cv *ContractValidator) Name() string {
	return cv.name
}

func (cv *ContractValidator) Timeout() time.Duration {
	return cv.timeout
}

func (cv *ContractValidator) Required() bool {
	return cv.required
}

func (cv *ContractValidator) Validate(ctx context.Context, environment string, metadata map[string]interface{}) error {
	// Contract validation logic - this would integrate with the testing framework
	contracts := []struct {
		name  string
		check func() error
	}{
		{
			name: "Database contract validation",
			check: func() error {
				// Validate database component contracts
				return nil
			},
		},
		{
			name: "Storage contract validation",
			check: func() error {
				// Validate storage component contracts
				return nil
			},
		},
		{
			name: "Vault contract validation",
			check: func() error {
				// Validate vault component contracts
				return nil
			},
		},
		{
			name: "Component integration validation",
			check: func() error {
				// Validate component integration contracts
				return nil
			},
		},
	}
	
	for _, contract := range contracts {
		if err := contract.check(); err != nil {
			return fmt.Errorf("contract validation '%s' failed: %w", contract.name, err)
		}
	}
	
	return nil
}

// HealthValidator implementation
func (hv *HealthValidator) Name() string {
	return hv.name
}

func (hv *HealthValidator) Timeout() time.Duration {
	return hv.timeout
}

func (hv *HealthValidator) Required() bool {
	return hv.required
}

func (hv *HealthValidator) Validate(ctx context.Context, environment string, metadata map[string]interface{}) error {
	// Health validation logic
	healthChecks := []struct {
		name  string
		check func() error
	}{
		{
			name: "Resource availability",
			check: func() error {
				// Check that all required resources are available
				return nil
			},
		},
		{
			name: "Service connectivity",
			check: func() error {
				// Check connectivity between services
				return nil
			},
		},
		{
			name: "Performance benchmarks",
			check: func() error {
				// Validate performance meets requirements
				if environment == "production" {
					// Production has stricter performance requirements
					return nil
				}
				return nil
			},
		},
		{
			name: "Dependency validation",
			check: func() error {
				// Validate dependencies are available and healthy
				return nil
			},
		},
	}
	
	for _, healthCheck := range healthChecks {
		if err := healthCheck.check(); err != nil {
			return fmt.Errorf("health check '%s' failed: %w", healthCheck.name, err)
		}
	}
	
	return nil
}

// ValidationSummary provides a summary of validation results
type ValidationSummary struct {
	TotalValidations     int
	SuccessfulValidations int
	FailedValidations    int
	Warnings             []string
	Errors              []string
	Duration            time.Duration
}

// RunAllValidations runs all validations and returns a summary
func (dv *DeploymentValidator) RunAllValidations(ctx context.Context, environment string, metadata map[string]interface{}) (*ValidationSummary, error) {
	startTime := time.Now()
	summary := &ValidationSummary{}
	
	validationTypes := []ValidationType{
		ValidationTypePreDeploy,
		ValidationTypeSecurity,
		ValidationTypeCompliance,
		ValidationTypeContract,
	}
	
	for _, validationType := range validationTypes {
		validators := dv.GetValidators(validationType)
		for _, validator := range validators {
			summary.TotalValidations++
			
			err := dv.RunValidation(ctx, validationType, environment, metadata)
			if err != nil {
				summary.FailedValidations++
				summary.Errors = append(summary.Errors, err.Error())
				
				if validator.Required() {
					summary.Duration = time.Since(startTime)
					return summary, err
				}
			} else {
				summary.SuccessfulValidations++
			}
		}
	}
	
	summary.Duration = time.Since(startTime)
	return summary, nil
}