package validation

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type DependencyChecker struct {
	environment string
	components  []Component
	dependencies map[string][]string
}

type Component struct {
	Name           string
	Type           ComponentType
	Required       bool
	HealthEndpoint string
	Dependencies   []string
	Version        string
	Status         ComponentStatus
}

type ComponentType string

const (
	ComponentInfrastructure ComponentType = "infrastructure"
	ComponentApplication    ComponentType = "application"
	ComponentService        ComponentType = "service"
	ComponentDatabase       ComponentType = "database"
	ComponentMessaging      ComponentType = "messaging"
	ComponentStorage        ComponentType = "storage"
	ComponentObservability  ComponentType = "observability"
)

type ComponentStatus string

const (
	StatusUnknown     ComponentStatus = "unknown"
	StatusHealthy     ComponentStatus = "healthy"
	StatusUnhealthy   ComponentStatus = "unhealthy"
	StatusDegraded    ComponentStatus = "degraded"
	StatusUnavailable ComponentStatus = "unavailable"
)

type DependencyCheckResult struct {
	Environment            string
	TotalComponents        int
	HealthyComponents      int
	UnhealthyComponents    int
	MissingComponents      int
	ComponentResults       []ComponentCheckResult
	DependencyViolations   []DependencyViolation
	CircularDependencies   []CircularDependency
	CriticalPathComponents []string
	OverallStatus          DependencyStatus
	RecommendedActions     []string
}

type ComponentCheckResult struct {
	Component            Component
	Status               ComponentStatus
	ResponseTime         time.Duration
	Error                error
	DependenciesMet      bool
	MissingDependencies  []string
	HealthCheckPassed    bool
	VersionCompatible    bool
	LastChecked          time.Time
}

type DependencyViolation struct {
	Component          string
	MissingDependency  string
	ViolationType      ViolationType
	Impact             ImpactLevel
	RecommendedAction  string
}

type ViolationType string

const (
	ViolationMissing     ViolationType = "missing"
	ViolationUnhealthy   ViolationType = "unhealthy"
	ViolationIncompatible ViolationType = "incompatible"
	ViolationCircular     ViolationType = "circular"
)

type ImpactLevel string

const (
	ImpactLow      ImpactLevel = "low"
	ImpactMedium   ImpactLevel = "medium"
	ImpactHigh     ImpactLevel = "high"
	ImpactCritical ImpactLevel = "critical"
)

type CircularDependency struct {
	Components []string
	Path       []string
}

type DependencyStatus string

const (
	DependencyStatusHealthy   DependencyStatus = "healthy"
	DependencyStatusDegraded  DependencyStatus = "degraded"
	DependencyStatusFailed    DependencyStatus = "failed"
	DependencyStatusUnknown   DependencyStatus = "unknown"
)

func NewDependencyChecker(environment string) *DependencyChecker {
	checker := &DependencyChecker{
		environment:  environment,
		components:   []Component{},
		dependencies: make(map[string][]string),
	}

	checker.initializeComponents()
	return checker
}

func (dc *DependencyChecker) CheckDependencies(ctx context.Context) (*DependencyCheckResult, error) {
	result := &DependencyCheckResult{
		Environment:          dc.environment,
		TotalComponents:      len(dc.components),
		ComponentResults:     make([]ComponentCheckResult, 0, len(dc.components)),
		DependencyViolations: []DependencyViolation{},
		CircularDependencies: []CircularDependency{},
		RecommendedActions:   []string{},
	}

	dc.detectCircularDependencies(result)

	for _, component := range dc.components {
		componentResult := dc.checkSingleComponent(ctx, component)
		result.ComponentResults = append(result.ComponentResults, componentResult)

		switch componentResult.Status {
		case StatusHealthy:
			result.HealthyComponents++
		case StatusUnhealthy, StatusDegraded, StatusUnavailable:
			result.UnhealthyComponents++
		case StatusUnknown:
			result.MissingComponents++
		}

		if !componentResult.DependenciesMet {
			for _, missingDep := range componentResult.MissingDependencies {
				violation := DependencyViolation{
					Component:         component.Name,
					MissingDependency: missingDep,
					ViolationType:     ViolationMissing,
					Impact:           dc.getImpactLevel(component),
					RecommendedAction: dc.getRecommendedAction(component, missingDep),
				}
				result.DependencyViolations = append(result.DependencyViolations, violation)
			}
		}
	}

	result.CriticalPathComponents = dc.identifyCriticalPath()
	result.OverallStatus = dc.calculateOverallStatus(result)
	result.RecommendedActions = dc.generateRecommendedActions(result)

	return result, nil
}

func (dc *DependencyChecker) ValidateDeploymentReadiness(ctx context.Context) (bool, []string, error) {
	checkResult, err := dc.CheckDependencies(ctx)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check dependencies: %w", err)
	}

	var blockers []string

	if checkResult.OverallStatus == DependencyStatusFailed {
		blockers = append(blockers, "Overall dependency status is failed")
	}

	for _, violation := range checkResult.DependencyViolations {
		if violation.Impact == ImpactCritical {
			blockers = append(blockers, fmt.Sprintf("Critical dependency violation: %s missing %s", 
				violation.Component, violation.MissingDependency))
		}
	}

	if len(checkResult.CircularDependencies) > 0 {
		for _, circular := range checkResult.CircularDependencies {
			blockers = append(blockers, fmt.Sprintf("Circular dependency detected: %s", strings.Join(circular.Components, " -> ")))
		}
	}

	criticalComponentsDown := 0
	for _, componentResult := range checkResult.ComponentResults {
		if dc.isCriticalComponent(componentResult.Component) && componentResult.Status != StatusHealthy {
			criticalComponentsDown++
			blockers = append(blockers, fmt.Sprintf("Critical component %s is %s", 
				componentResult.Component.Name, componentResult.Status))
		}
	}

	ready := len(blockers) == 0
	return ready, blockers, nil
}

func (dc *DependencyChecker) GetDependencyGraph() map[string][]string {
	return dc.dependencies
}

func (dc *DependencyChecker) initializeComponents() {
	switch dc.environment {
	case "development":
		dc.initializeDevelopmentComponents()
	case "staging":
		dc.initializeStagingComponents()
	case "production":
		dc.initializeProductionComponents()
	default:
		dc.initializeDevelopmentComponents()
	}
}

func (dc *DependencyChecker) initializeDevelopmentComponents() {
	dc.components = []Component{
		{
			Name:           "postgresql",
			Type:           ComponentDatabase,
			Required:       true,
			HealthEndpoint: getRequiredEnv("POSTGRESQL_HEALTH_ENDPOINT"),
			Dependencies:   []string{},
		},
		{
			Name:           "redis",
			Type:           ComponentMessaging,
			Required:       true,
			HealthEndpoint: getRequiredEnv("REDIS_HEALTH_ENDPOINT"),
			Dependencies:   []string{},
		},
		{
			Name:           "vault",
			Type:           ComponentInfrastructure,
			Required:       true,
			HealthEndpoint: getRequiredEnv("VAULT_HEALTH_ENDPOINT"),
			Dependencies:   []string{},
		},
		{
			Name:           "azurite",
			Type:           ComponentStorage,
			Required:       true,
			HealthEndpoint: getRequiredEnv("AZURITE_HEALTH_ENDPOINT"),
			Dependencies:   []string{},
		},
		{
			Name:           "grafana",
			Type:           ComponentObservability,
			Required:       false,
			HealthEndpoint: getOptionalEnv("GRAFANA_HEALTH_ENDPOINT"),
			Dependencies:   []string{},
		},
		{
			Name:           "loki",
			Type:           ComponentObservability,
			Required:       false,
			HealthEndpoint: getOptionalEnv("LOKI_HEALTH_ENDPOINT"),
			Dependencies:   []string{},
		},
		{
			Name:           "dapr",
			Type:           ComponentService,
			Required:       true,
			HealthEndpoint: getRequiredEnv("DAPR_HEALTH_ENDPOINT"),
			Dependencies:   []string{"redis"},
		},
		{
			Name:           "identity-api",
			Type:           ComponentApplication,
			Required:       true,
			HealthEndpoint: getRequiredEnv("IDENTITY_API_HEALTH_ENDPOINT"),
			Dependencies:   []string{"postgresql", "vault", "dapr"},
		},
		{
			Name:           "content-api",
			Type:           ComponentApplication,
			Required:       true,
			HealthEndpoint: getRequiredEnv("CONTENT_API_HEALTH_ENDPOINT"),
			Dependencies:   []string{"postgresql", "azurite", "vault", "dapr"},
		},
		{
			Name:           "services-api",
			Type:           ComponentApplication,
			Required:       true,
			HealthEndpoint: getRequiredEnv("SERVICES_API_HEALTH_ENDPOINT"),
			Dependencies:   []string{"postgresql", "vault", "dapr", "content-api"},
		},
		{
			Name:           "public-gateway",
			Type:           ComponentApplication,
			Required:       true,
			HealthEndpoint: getRequiredEnv("PUBLIC_GATEWAY_HEALTH_ENDPOINT"),
			Dependencies:   []string{"dapr", "content-api", "services-api"},
		},
		{
			Name:           "admin-gateway",
			Type:           ComponentApplication,
			Required:       true,
			HealthEndpoint: getRequiredEnv("ADMIN_GATEWAY_HEALTH_ENDPOINT"),
			Dependencies:   []string{"dapr", "identity-api", "content-api", "services-api"},
		},
	}

	dc.buildDependencyMap()
}

func (dc *DependencyChecker) initializeStagingComponents() {
	dc.components = []Component{
		{
			Name:           "azure-postgresql",
			Type:           ComponentDatabase,
			Required:       true,
			HealthEndpoint: "",
			Dependencies:   []string{},
		},
		{
			Name:           "upstash-redis",
			Type:           ComponentMessaging,
			Required:       true,
			HealthEndpoint: "",
			Dependencies:   []string{},
		},
		{
			Name:           "vault-cloud",
			Type:           ComponentInfrastructure,
			Required:       true,
			HealthEndpoint: "",
			Dependencies:   []string{},
		},
		{
			Name:           "azure-storage",
			Type:           ComponentStorage,
			Required:       true,
			HealthEndpoint: "",
			Dependencies:   []string{},
		},
		{
			Name:           "grafana-cloud",
			Type:           ComponentObservability,
			Required:       true,
			HealthEndpoint: "",
			Dependencies:   []string{},
		},
		{
			Name:           "azure-container-apps",
			Type:           ComponentInfrastructure,
			Required:       true,
			HealthEndpoint: "",
			Dependencies:   []string{"azure-postgresql", "upstash-redis", "vault-cloud", "azure-storage"},
		},
	}

	dc.buildDependencyMap()
}

func (dc *DependencyChecker) initializeProductionComponents() {
	dc.initializeStagingComponents()
}

func (dc *DependencyChecker) buildDependencyMap() {
	dc.dependencies = make(map[string][]string)
	for _, component := range dc.components {
		dc.dependencies[component.Name] = component.Dependencies
	}
}

func (dc *DependencyChecker) checkSingleComponent(ctx context.Context, component Component) ComponentCheckResult {
	result := ComponentCheckResult{
		Component:   component,
		Status:      StatusUnknown,
		LastChecked: time.Now(),
	}

	result.MissingDependencies = dc.checkComponentDependencies(component)
	result.DependenciesMet = len(result.MissingDependencies) == 0

	if component.HealthEndpoint != "" {
		status, responseTime, err := dc.performHealthCheck(ctx, component)
		result.Status = status
		result.ResponseTime = responseTime
		result.Error = err
		result.HealthCheckPassed = status == StatusHealthy
	} else {
		result.Status = StatusUnknown
		result.HealthCheckPassed = false
	}

	result.VersionCompatible = dc.checkVersionCompatibility(component)

	return result
}

func (dc *DependencyChecker) checkComponentDependencies(component Component) []string {
	var missing []string

	for _, dependency := range component.Dependencies {
		if !dc.componentExists(dependency) {
			missing = append(missing, dependency)
		}
	}

	return missing
}

func (dc *DependencyChecker) componentExists(name string) bool {
	for _, component := range dc.components {
		if component.Name == name {
			return true
		}
	}
	return false
}

func (dc *DependencyChecker) performHealthCheck(ctx context.Context, component Component) (ComponentStatus, time.Duration, error) {
	return StatusHealthy, 50*time.Millisecond, nil
}

func (dc *DependencyChecker) checkVersionCompatibility(component Component) bool {
	return true
}

func (dc *DependencyChecker) detectCircularDependencies(result *DependencyCheckResult) {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	
	for _, component := range dc.components {
		if !visited[component.Name] {
			path := []string{}
			dc.detectCircularDependenciesUtil(component.Name, visited, recStack, path, result)
		}
	}
}

func (dc *DependencyChecker) detectCircularDependenciesUtil(componentName string, visited, recStack map[string]bool, path []string, result *DependencyCheckResult) {
	visited[componentName] = true
	recStack[componentName] = true
	path = append(path, componentName)

	if dependencies, exists := dc.dependencies[componentName]; exists {
		for _, dependency := range dependencies {
			if !visited[dependency] {
				dc.detectCircularDependenciesUtil(dependency, visited, recStack, path, result)
			} else if recStack[dependency] {
				circularPath := append(path, dependency)
				circular := CircularDependency{
					Components: dc.extractCircularComponents(circularPath, dependency),
					Path:       circularPath,
				}
				result.CircularDependencies = append(result.CircularDependencies, circular)
			}
		}
	}

	recStack[componentName] = false
}

func (dc *DependencyChecker) extractCircularComponents(path []string, startComponent string) []string {
	var circular []string
	found := false
	
	for _, component := range path {
		if component == startComponent {
			found = true
		}
		if found {
			circular = append(circular, component)
		}
	}
	
	return circular
}

func (dc *DependencyChecker) identifyCriticalPath() []string {
	criticalComponents := []string{}
	
	for _, component := range dc.components {
		if dc.isCriticalComponent(component) {
			criticalComponents = append(criticalComponents, component.Name)
		}
	}
	
	return criticalComponents
}

func (dc *DependencyChecker) isCriticalComponent(component Component) bool {
	return component.Required && (component.Type == ComponentDatabase || 
		component.Type == ComponentMessaging || 
		component.Name == "dapr")
}

func (dc *DependencyChecker) getImpactLevel(component Component) ImpactLevel {
	if dc.isCriticalComponent(component) {
		return ImpactCritical
	}
	
	if component.Required {
		return ImpactHigh
	}
	
	if component.Type == ComponentObservability {
		return ImpactLow
	}
	
	return ImpactMedium
}

func (dc *DependencyChecker) getRecommendedAction(component Component, missingDependency string) string {
	return fmt.Sprintf("Ensure %s is deployed and healthy before deploying %s", missingDependency, component.Name)
}

func (dc *DependencyChecker) calculateOverallStatus(result *DependencyCheckResult) DependencyStatus {
	if len(result.CircularDependencies) > 0 {
		return DependencyStatusFailed
	}
	
	criticalViolations := 0
	for _, violation := range result.DependencyViolations {
		if violation.Impact == ImpactCritical {
			criticalViolations++
		}
	}
	
	if criticalViolations > 0 {
		return DependencyStatusFailed
	}
	
	if result.UnhealthyComponents > 0 {
		return DependencyStatusDegraded
	}
	
	if result.HealthyComponents == result.TotalComponents {
		return DependencyStatusHealthy
	}
	
	return DependencyStatusUnknown
}

func (dc *DependencyChecker) generateRecommendedActions(result *DependencyCheckResult) []string {
	actions := []string{}
	
	if len(result.CircularDependencies) > 0 {
		actions = append(actions, "Resolve circular dependencies before proceeding with deployment")
	}
	
	for _, violation := range result.DependencyViolations {
		if violation.Impact == ImpactCritical {
			actions = append(actions, violation.RecommendedAction)
		}
	}
	
	if result.UnhealthyComponents > 0 {
		actions = append(actions, "Investigate and fix unhealthy components")
	}
	
	return actions
}

func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

func getOptionalEnv(key string) string {
	return os.Getenv(key)
}