package shared

import (
	"fmt"
	"time"

	"github.com/axiom-software-co/international-center/src/cicd/components"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// HealthMonitor provides health monitoring and rollback capabilities for deployments
type HealthMonitor struct {
	ctx         *pulumi.Context
	environment string
}

// ComponentHealth represents the health status of a component
type ComponentHealth struct {
	Name         string
	Healthy      bool
	Status       string
	LastChecked  time.Time
	Dependencies []string
	Errors       []string
}

// DeploymentHealth aggregates health status for all components
type DeploymentHealth struct {
	OverallHealthy bool
	Components     map[string]*ComponentHealth
	Issues         []string
	LastUpdated    time.Time
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(ctx *pulumi.Context, environment string) *HealthMonitor {
	return &HealthMonitor{
		ctx:         ctx,
		environment: environment,
	}
}

// ValidateComponentHealth validates the health of a specific component
func (h *HealthMonitor) ValidateComponentHealth(componentName string, outputs interface{}) *ComponentHealth {
	health := &ComponentHealth{
		Name:        componentName,
		LastChecked: time.Now(),
		Healthy:     true,
		Status:      "healthy",
		Errors:      []string{},
	}

	switch componentName {
	case "database":
		health = h.validateDatabaseHealth(outputs.(*components.DatabaseOutputs))
	case "storage":
		health = h.validateStorageHealth(outputs.(*components.StorageOutputs))
	case "vault":
		health = h.validateVaultHealth(outputs.(*components.VaultOutputs))
	case "observability":
		health = h.validateObservabilityHealth(outputs.(*components.ObservabilityOutputs))
	case "dapr":
		health = h.validateDaprHealth(outputs.(*components.DaprOutputs))
	case "services":
		health = h.validateServicesHealth(outputs.(*components.ServicesOutputs))
	case "website":
		health = h.validateWebsiteHealth(outputs.(*components.WebsiteOutputs))
	default:
		health.Healthy = false
		health.Status = "unknown_component"
		health.Errors = append(health.Errors, fmt.Sprintf("Unknown component: %s", componentName))
	}

	h.ctx.Log.Info(fmt.Sprintf("%s health check: %s", componentName, health.Status), nil)
	
	return health
}

// validateDatabaseHealth validates database component health
func (h *HealthMonitor) validateDatabaseHealth(outputs *components.DatabaseOutputs) *ComponentHealth {
	health := &ComponentHealth{
		Name:         "database",
		LastChecked:  time.Now(),
		Healthy:      true,
		Status:       "healthy",
		Dependencies: []string{},
		Errors:       []string{},
	}

	// In a real deployment, this would:
	// - Test database connectivity
	// - Validate database schema
	// - Check database performance metrics
	// - Verify backup systems
	
	// For now, we validate that required outputs are present
	if outputs == nil {
		health.Healthy = false
		health.Status = "failed"
		health.Errors = append(health.Errors, "Database outputs are nil")
		return health
	}

	// Validate connection string is available
	outputs.ConnectionString.ApplyT(func(connStr string) error {
		if connStr == "" {
			health.Healthy = false
			health.Status = "misconfigured"
			health.Errors = append(health.Errors, "Database connection string is empty")
		}
		return nil
	})

	return health
}

// validateStorageHealth validates storage component health
func (h *HealthMonitor) validateStorageHealth(outputs *components.StorageOutputs) *ComponentHealth {
	health := &ComponentHealth{
		Name:         "storage",
		LastChecked:  time.Now(),
		Healthy:      true,
		Status:       "healthy",
		Dependencies: []string{},
		Errors:       []string{},
	}

	// In a real deployment, this would:
	// - Test storage connectivity
	// - Validate storage permissions
	// - Check storage capacity
	// - Verify backup and redundancy
	
	if outputs == nil {
		health.Healthy = false
		health.Status = "failed"
		health.Errors = append(health.Errors, "Storage outputs are nil")
		return health
	}

	outputs.ConnectionString.ApplyT(func(connStr string) error {
		if connStr == "" {
			health.Healthy = false
			health.Status = "misconfigured"
			health.Errors = append(health.Errors, "Storage connection string is empty")
		}
		return nil
	})

	return health
}

// validateVaultHealth validates vault component health
func (h *HealthMonitor) validateVaultHealth(outputs *components.VaultOutputs) *ComponentHealth {
	health := &ComponentHealth{
		Name:         "vault",
		LastChecked:  time.Now(),
		Healthy:      true,
		Status:       "healthy",
		Dependencies: []string{},
		Errors:       []string{},
	}

	// In a real deployment, this would:
	// - Test vault connectivity
	// - Validate vault unsealing
	// - Check secret access permissions
	// - Verify audit logging
	
	if outputs == nil {
		health.Healthy = false
		health.Status = "failed"
		health.Errors = append(health.Errors, "Vault outputs are nil")
		return health
	}

	outputs.VaultAddress.ApplyT(func(address string) error {
		if address == "" {
			health.Healthy = false
			health.Status = "misconfigured"
			health.Errors = append(health.Errors, "Vault address is empty")
		}
		return nil
	})

	return health
}

// validateObservabilityHealth validates observability component health
func (h *HealthMonitor) validateObservabilityHealth(outputs *components.ObservabilityOutputs) *ComponentHealth {
	health := &ComponentHealth{
		Name:         "observability",
		LastChecked:  time.Now(),
		Healthy:      true,
		Status:       "healthy",
		Dependencies: []string{},
		Errors:       []string{},
	}

	// In a real deployment, this would:
	// - Test Grafana connectivity
	// - Validate metrics collection
	// - Check alert manager configuration
	// - Verify log aggregation
	
	if outputs == nil {
		health.Healthy = false
		health.Status = "failed"
		health.Errors = append(health.Errors, "Observability outputs are nil")
		return health
	}

	outputs.GrafanaURL.ApplyT(func(url string) error {
		if url == "" {
			health.Healthy = false
			health.Status = "misconfigured"
			health.Errors = append(health.Errors, "Grafana URL is empty")
		}
		return nil
	})

	return health
}

// validateDaprHealth validates Dapr component health
func (h *HealthMonitor) validateDaprHealth(outputs *components.DaprOutputs) *ComponentHealth {
	health := &ComponentHealth{
		Name:         "dapr",
		LastChecked:  time.Now(),
		Healthy:      true,
		Status:       "healthy",
		Dependencies: []string{"database", "storage", "vault"},
		Errors:       []string{},
	}

	// In a real deployment, this would:
	// - Test Dapr control plane connectivity
	// - Validate service mesh configuration
	// - Check middleware and policy configuration
	// - Verify component connectivity
	
	if outputs == nil {
		health.Healthy = false
		health.Status = "failed"
		health.Errors = append(health.Errors, "Dapr outputs are nil")
		return health
	}

	outputs.ControlPlaneURL.ApplyT(func(url string) error {
		if url == "" {
			health.Healthy = false
			health.Status = "misconfigured"
			health.Errors = append(health.Errors, "Dapr control plane URL is empty")
		}
		return nil
	})

	return health
}

// validateServicesHealth validates services component health
func (h *HealthMonitor) validateServicesHealth(outputs *components.ServicesOutputs) *ComponentHealth {
	health := &ComponentHealth{
		Name:         "services",
		LastChecked:  time.Now(),
		Healthy:      true,
		Status:       "healthy",
		Dependencies: []string{"database", "storage", "vault", "dapr", "observability"},
		Errors:       []string{},
	}

	// In a real deployment, this would:
	// - Test service endpoints
	// - Validate service mesh communication
	// - Check health endpoints
	// - Verify load balancer configuration
	
	if outputs == nil {
		health.Healthy = false
		health.Status = "failed"
		health.Errors = append(health.Errors, "Services outputs are nil")
		return health
	}

	pulumi.All(outputs.PublicGatewayURL, outputs.AdminGatewayURL).ApplyT(func(args []interface{}) error {
		publicURL := args[0].(string)
		adminURL := args[1].(string)

		if publicURL == "" {
			health.Healthy = false
			health.Status = "misconfigured"
			health.Errors = append(health.Errors, "Public gateway URL is empty")
		}

		if adminURL == "" {
			health.Healthy = false
			health.Status = "misconfigured"
			health.Errors = append(health.Errors, "Admin gateway URL is empty")
		}

		return nil
	})

	return health
}

// validateWebsiteHealth validates website component health
func (h *HealthMonitor) validateWebsiteHealth(outputs *components.WebsiteOutputs) *ComponentHealth {
	health := &ComponentHealth{
		Name:         "website",
		LastChecked:  time.Now(),
		Healthy:      true,
		Status:       "healthy",
		Dependencies: []string{"services"},
		Errors:       []string{},
	}

	// In a real deployment, this would:
	// - Test website accessibility
	// - Validate CDN configuration
	// - Check API connectivity
	// - Verify SSL/TLS configuration
	
	if outputs == nil {
		health.Healthy = false
		health.Status = "failed"
		health.Errors = append(health.Errors, "Website outputs are nil")
		return health
	}

	outputs.ServerURL.ApplyT(func(url string) error {
		if url == "" {
			health.Healthy = false
			health.Status = "misconfigured"
			health.Errors = append(health.Errors, "Website server URL is empty")
		}
		return nil
	})

	return health
}

// GetOverallHealth aggregates health status for all components
func (h *HealthMonitor) GetOverallHealth(componentHealths map[string]*ComponentHealth) *DeploymentHealth {
	deploymentHealth := &DeploymentHealth{
		Components:     componentHealths,
		Issues:         []string{},
		LastUpdated:    time.Now(),
		OverallHealthy: true,
	}

	// Check overall health
	for _, health := range componentHealths {
		if !health.Healthy {
			deploymentHealth.OverallHealthy = false
			deploymentHealth.Issues = append(deploymentHealth.Issues, 
				fmt.Sprintf("%s: %s", health.Name, health.Status))
			
			// Add specific errors
			for _, err := range health.Errors {
				deploymentHealth.Issues = append(deploymentHealth.Issues, 
					fmt.Sprintf("%s error: %s", health.Name, err))
			}
		}
	}

	// Log overall health status
	if deploymentHealth.OverallHealthy {
		h.ctx.Log.Info("All components are healthy", nil)
	} else {
		h.ctx.Log.Warn(fmt.Sprintf("Deployment health issues detected: %v", deploymentHealth.Issues), nil)
	}

	return deploymentHealth
}

// CheckDependencyHealth validates that component dependencies are healthy
func (h *HealthMonitor) CheckDependencyHealth(componentName string, allHealth map[string]*ComponentHealth) []string {
	var issues []string

	componentHealth, exists := allHealth[componentName]
	if !exists {
		return []string{fmt.Sprintf("Component %s not found", componentName)}
	}

	// Check each dependency
	for _, dep := range componentHealth.Dependencies {
		depHealth, depExists := allHealth[dep]
		if !depExists {
			issues = append(issues, fmt.Sprintf("Dependency %s not found for %s", dep, componentName))
			continue
		}

		if !depHealth.Healthy {
			issues = append(issues, fmt.Sprintf("Dependency %s is unhealthy for %s: %s", dep, componentName, depHealth.Status))
		}
	}

	return issues
}

// ShouldRollback determines if a deployment should be rolled back based on health status
func (h *HealthMonitor) ShouldRollback(health *DeploymentHealth, environment string) bool {
	// For development, we're more lenient
	if environment == "development" {
		// Only rollback if more than 50% of components are failing
		totalComponents := len(health.Components)
		unhealthyComponents := 0
		
		for _, componentHealth := range health.Components {
			if !componentHealth.Healthy {
				unhealthyComponents++
			}
		}
		
		return float64(unhealthyComponents)/float64(totalComponents) > 0.5
	}

	// For staging and production, rollback if any critical component is unhealthy
	criticalComponents := []string{"database", "vault", "services"}
	for _, critical := range criticalComponents {
		if componentHealth, exists := health.Components[critical]; exists && !componentHealth.Healthy {
			h.ctx.Log.Error(fmt.Sprintf("Critical component %s is unhealthy, rollback recommended", critical), nil)
			return true
		}
	}

	return false
}