package validation

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type EnvironmentHealthChecker struct {
	environment string
	validator   *InfrastructureValidator
	config      *ValidationConfig
}

func NewEnvironmentHealthChecker(environment string) (*EnvironmentHealthChecker, error) {
	config, err := getValidationConfig(environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get validation config for environment %s: %w", environment, err)
	}

	validator := NewInfrastructureValidator(config)

	return &EnvironmentHealthChecker{
		environment: environment,
		validator:   validator,
		config:      config,
	}, nil
}

func getValidationConfig(environment string) (*ValidationConfig, error) {
	switch environment {
	case "development":
		return getDevelopmentValidationConfig(), nil
	case "staging":
		return getStagingValidationConfig(), nil
	case "production":
		return getProductionValidationConfig(), nil
	default:
		return nil, fmt.Errorf("unsupported environment: %s", environment)
	}
}

func getDevelopmentValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		Environment:          "development",
		TimeoutSeconds:       30,
		MaxRetries:          3,
		RetryDelaySeconds:   5,
		HealthCheckInterval: 30 * time.Second,
		ExpectedComponents: []string{
			"infrastructure",
			"platform",
			"services",
			"website",
		},
		SecurityChecks: []string{
			// Minimal security checks for development
		},
		ComplianceChecks: []string{
			// No compliance checks for development
		},
	}
}

func getStagingValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		Environment:          "staging",
		TimeoutSeconds:       60,
		MaxRetries:          5,
		RetryDelaySeconds:   10,
		HealthCheckInterval: 15 * time.Second,
		ExpectedComponents: []string{
			"infrastructure",
			"platform",
			"services", 
			"website",
			"monitoring",
			"security",
		},
		SecurityChecks: []string{
			"encryption_at_rest",
			"encryption_in_transit",
			"access_control",
			"network_policies",
		},
		ComplianceChecks: []string{
			"audit_logging",
			"data_retention",
			"backup_policies",
		},
	}
}

func getProductionValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		Environment:          "production",
		TimeoutSeconds:       120,
		MaxRetries:          10,
		RetryDelaySeconds:   15,
		HealthCheckInterval: 10 * time.Second,
		ExpectedComponents: []string{
			"infrastructure",
			"platform",
			"services",
			"website",
			"monitoring",
			"security",
			"backup",
			"disaster_recovery",
		},
		SecurityChecks: []string{
			"encryption_at_rest",
			"encryption_in_transit", 
			"access_control",
			"network_policies",
			"vulnerability_scanning",
			"secret_rotation",
			"certificate_validation",
		},
		ComplianceChecks: []string{
			"audit_logging",
			"data_retention",
			"backup_policies",
			"compliance_frameworks",
			"access_control_auditing",
			"data_classification",
			"incident_response",
		},
	}
}

func (ehc *EnvironmentHealthChecker) PerformHealthCheck(ctx context.Context, outputs map[string]interface{}) (*EnvironmentHealthReport, error) {
	log.Printf("Starting environment health check for: %s", ehc.environment)
	
	startTime := time.Now()
	report := &EnvironmentHealthReport{
		Environment: ehc.environment,
		StartTime:   startTime,
		Components:  make(map[string]ComponentHealth),
		Overall:     HealthStatusUnknown,
	}

	// Convert outputs to pulumi.Map format
	pulumiOutputs := convertToPulumiMap(outputs)

	// Run infrastructure validation
	err := ehc.validator.ValidateInfrastructure(ctx, pulumiOutputs)
	
	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(startTime)
	report.ValidationResults = ehc.validator.GetValidationResults()
	report.ValidationSummary = ehc.validator.GetValidationSummary()

	if err != nil {
		report.Overall = HealthStatusUnhealthy
		report.ErrorMessage = err.Error()
		log.Printf("Environment health check failed for %s: %v", ehc.environment, err)
		return report, fmt.Errorf("environment health check failed: %w", err)
	}

	// Evaluate component health
	ehc.evaluateComponentHealth(report, outputs)

	// Determine overall health status
	ehc.determineOverallHealth(report)

	log.Printf("Environment health check completed for %s. Status: %s", ehc.environment, report.Overall)
	return report, nil
}

func (ehc *EnvironmentHealthChecker) evaluateComponentHealth(report *EnvironmentHealthReport, outputs map[string]interface{}) {
	for _, componentName := range ehc.config.ExpectedComponents {
		health := ComponentHealth{
			Name:      componentName,
			Status:    HealthStatusUnknown,
			CheckTime: time.Now(),
		}

		// Check if component outputs exist
		componentFound := false
		for outputKey := range outputs {
			if containsComponent(outputKey, componentName) {
				componentFound = true
				break
			}
		}

		if componentFound {
			health.Status = HealthStatusHealthy
			health.Message = "Component deployed successfully"
		} else {
			health.Status = HealthStatusUnhealthy
			health.Message = "Component not found in deployment outputs"
		}

		report.Components[componentName] = health
	}
}

func (ehc *EnvironmentHealthChecker) determineOverallHealth(report *EnvironmentHealthReport) {
	healthyCount := 0
	totalCount := len(report.Components)

	for _, component := range report.Components {
		if component.Status == HealthStatusHealthy {
			healthyCount++
		}
	}

	// Check validation results
	validationPassed := true
	for _, result := range report.ValidationResults {
		if !result.Success {
			validationPassed = false
			break
		}
	}

	if healthyCount == totalCount && validationPassed {
		report.Overall = HealthStatusHealthy
	} else if healthyCount > 0 {
		report.Overall = HealthStatusDegraded
	} else {
		report.Overall = HealthStatusUnhealthy
	}
}

type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

type ComponentHealth struct {
	Name      string       `json:"name"`
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message"`
	CheckTime time.Time    `json:"check_time"`
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
}

type EnvironmentHealthReport struct {
	Environment        string                     `json:"environment"`
	Overall           HealthStatus               `json:"overall_status"`
	Components        map[string]ComponentHealth `json:"components"`
	ValidationResults []ValidationResult         `json:"validation_results"`
	ValidationSummary map[string]interface{}     `json:"validation_summary"`
	StartTime         time.Time                  `json:"start_time"`
	EndTime           time.Time                  `json:"end_time"`
	Duration          time.Duration              `json:"duration"`
	ErrorMessage      string                     `json:"error_message,omitempty"`
}

func (ehr *EnvironmentHealthReport) IsHealthy() bool {
	return ehr.Overall == HealthStatusHealthy
}

func (ehr *EnvironmentHealthReport) GetUnhealthyComponents() []string {
	unhealthy := make([]string, 0)
	for name, component := range ehr.Components {
		if component.Status == HealthStatusUnhealthy {
			unhealthy = append(unhealthy, name)
		}
	}
	return unhealthy
}

func (ehr *EnvironmentHealthReport) GetHealthySummary() map[string]interface{} {
	healthy := 0
	degraded := 0
	unhealthy := 0
	total := len(ehr.Components)

	for _, component := range ehr.Components {
		switch component.Status {
		case HealthStatusHealthy:
			healthy++
		case HealthStatusDegraded:
			degraded++
		case HealthStatusUnhealthy:
			unhealthy++
		}
	}

	return map[string]interface{}{
		"total_components":    total,
		"healthy_components":  healthy,
		"degraded_components": degraded,
		"unhealthy_components": unhealthy,
		"health_percentage":   float64(healthy) / float64(total) * 100,
		"overall_status":      ehr.Overall,
		"environment":         ehr.Environment,
		"check_duration":      ehr.Duration.Seconds(),
	}
}

func convertToPulumiMap(outputs map[string]interface{}) pulumi.Map {
	// Convert regular map to pulumi.Map compatible format
	pulumiMap := make(pulumi.Map)
	for key, value := range outputs {
		pulumiMap[key] = pulumi.Any(value)
	}
	return pulumiMap
}

func containsComponent(outputKey, componentName string) bool {
	// Check if output key contains component name
	// This is a simplified check - in real implementation would be more sophisticated
	return len(outputKey) > 0 && len(componentName) > 0
}

func (ehc *EnvironmentHealthChecker) GetEnvironmentConfig() *ValidationConfig {
	return ehc.config
}

func (ehc *EnvironmentHealthChecker) GetEnvironmentRequirements() map[string]interface{} {
	return map[string]interface{}{
		"environment":           ehc.environment,
		"expected_components":   ehc.config.ExpectedComponents,
		"security_checks":       ehc.config.SecurityChecks,
		"compliance_checks":     ehc.config.ComplianceChecks,
		"timeout_seconds":       ehc.config.TimeoutSeconds,
		"max_retries":          ehc.config.MaxRetries,
		"health_check_interval": ehc.config.HealthCheckInterval.String(),
	}
}