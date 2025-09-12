package platform

import (
	"context"
	"fmt"
	"time"
)

// DaprSidecarConfig defines the configuration for Dapr sidecar injection
type DaprSidecarConfig struct {
	AppID                string
	AppPort              int
	DaprHTTPPort         int
	DaprGRPCPort         int
	DaprMetricsPort      int
	DaprProfilePort      int
	PlacementHostAddress string
	LogLevel             string
	EnableProfiling      bool
	EnableMetrics        bool
	MaxConcurrency       int
	ResourceLimits       ResourceLimits
}

// DaprSidecarInjector defines the interface for Dapr sidecar injection
type DaprSidecarInjector interface {
	// InjectSidecar injects a Dapr sidecar for the given container specification
	InjectSidecar(ctx context.Context, spec *ContainerSpec, config *DaprSidecarConfig) error
	
	// ValidateSidecarConfig validates the Dapr sidecar configuration
	ValidateSidecarConfig(config *DaprSidecarConfig) error
	
	// GetSidecarName returns the name of the sidecar container
	GetSidecarName(appID string) string
}

// UnifiedDaprSidecarManager provides common Dapr sidecar functionality
type UnifiedDaprSidecarManager struct {
	Environment    string
	HealthChecker  *UnifiedHealthChecker
}

// NewUnifiedDaprSidecarManager creates a new Dapr sidecar manager
func NewUnifiedDaprSidecarManager(environment string) *UnifiedDaprSidecarManager {
	return &UnifiedDaprSidecarManager{
		Environment:   environment,
		HealthChecker: NewUnifiedHealthChecker(),
	}
}

// BuildDefaultDaprConfig builds default Dapr configuration for a container
func (d *UnifiedDaprSidecarManager) BuildDefaultDaprConfig(appID string, appPort int) *DaprSidecarConfig {
	return &DaprSidecarConfig{
		AppID:                appID,
		AppPort:              appPort,
		DaprHTTPPort:         d.calculateDaprHTTPPort(appPort),
		DaprGRPCPort:         d.calculateDaprGRPCPort(appPort),
		DaprMetricsPort:      9090,
		DaprProfilePort:      7777,
		PlacementHostAddress: d.getPlacementHostAddress(),
		LogLevel:             d.getLogLevel(),
		EnableProfiling:      d.shouldEnableProfiling(),
		EnableMetrics:        true,
		MaxConcurrency:       d.getMaxConcurrency(),
		ResourceLimits: ResourceLimits{
			CPU:    d.getSidecarCPULimit(),
			Memory: d.getSidecarMemoryLimit(),
		},
	}
}

// ValidateContainerForDapr validates that a container is ready for Dapr sidecar injection
func (d *UnifiedDaprSidecarManager) ValidateContainerForDapr(spec *ContainerSpec) error {
	if spec.Name == "" {
		return fmt.Errorf("container name is required for Dapr sidecar injection")
	}
	
	if spec.DaprAppID == "" {
		return fmt.Errorf("Dapr app ID is required for sidecar injection")
	}
	
	if spec.Port <= 0 {
		return fmt.Errorf("valid application port is required for Dapr sidecar injection")
	}
	
	if !spec.DaprEnabled {
		return fmt.Errorf("Dapr must be enabled for sidecar injection")
	}
	
	// Validate that the app ID follows Dapr naming conventions
	if err := d.validateAppID(spec.DaprAppID); err != nil {
		return fmt.Errorf("invalid Dapr app ID: %w", err)
	}
	
	return nil
}

// EnrichContainerSpecWithDapr adds Dapr-specific configuration to a container spec
func (d *UnifiedDaprSidecarManager) EnrichContainerSpecWithDapr(spec *ContainerSpec) error {
	if !spec.DaprEnabled {
		return nil // Skip if Dapr is not enabled
	}
	
	// Validate first
	if err := d.ValidateContainerForDapr(spec); err != nil {
		return err
	}
	
	// Build default Dapr configuration if not provided
	if len(spec.DaprConfig) == 0 {
		spec.DaprConfig = make(map[string]interface{})
	}
	
	// Set default Dapr configuration values
	daprConfig := d.BuildDefaultDaprConfig(spec.DaprAppID, spec.Port)
	
	spec.DaprConfig["app_port"] = daprConfig.AppPort
	spec.DaprConfig["placement_host_address"] = daprConfig.PlacementHostAddress
	spec.DaprConfig["log_level"] = daprConfig.LogLevel
	spec.DaprConfig["enable_profiling"] = daprConfig.EnableProfiling
	spec.DaprConfig["enable_metrics"] = daprConfig.EnableMetrics
	spec.DaprConfig["metrics_port"] = daprConfig.DaprMetricsPort
	spec.DaprConfig["max_concurrency"] = daprConfig.MaxConcurrency
	
	// Set Dapr ports if not already configured
	if spec.DaprPort == 0 {
		spec.DaprPort = daprConfig.DaprHTTPPort
	}
	
	// Add Dapr-specific environment variables to the main container
	if spec.Environment == nil {
		spec.Environment = make(map[string]string)
	}
	
	spec.Environment["DAPR_HTTP_PORT"] = fmt.Sprintf("%d", daprConfig.DaprHTTPPort)
	spec.Environment["DAPR_GRPC_PORT"] = fmt.Sprintf("%d", daprConfig.DaprGRPCPort)
	
	return nil
}

// WaitForDaprSidecarReady waits for the Dapr sidecar to become ready
func (d *UnifiedDaprSidecarManager) WaitForDaprSidecarReady(ctx context.Context, appID string, injector DaprSidecarInjector, timeout time.Duration) error {
	sidecarName := injector.GetSidecarName(appID)
	
	// Create a health checker that can check the sidecar container
	checker, ok := injector.(ContainerHealthChecker)
	if !ok {
		return fmt.Errorf("injector does not implement ContainerHealthChecker")
	}
	
	// Wait for the sidecar container to be healthy
	return d.HealthChecker.WaitForContainerHealth(ctx, sidecarName, checker, timeout)
}

// GetDaprEndpoint returns the Dapr HTTP endpoint for an application
func (d *UnifiedDaprSidecarManager) GetDaprEndpoint(appID string, daprPort int) string {
	switch d.Environment {
	case "development":
		return fmt.Sprintf("http://localhost:%d", daprPort)
	case "staging":
		return fmt.Sprintf("https://%s-staging.azurecontainerapp.io", appID)
	case "production":
		return fmt.Sprintf("https://%s-production.azurecontainerapp.io", appID)
	default:
		return fmt.Sprintf("http://localhost:%d", daprPort)
	}
}

// BuildDaprCommand builds the Dapr daprd command arguments
func (d *UnifiedDaprSidecarManager) BuildDaprCommand(config *DaprSidecarConfig) []string {
	args := []string{
		"./daprd",
		fmt.Sprintf("--app-id=%s", config.AppID),
		fmt.Sprintf("--app-port=%d", config.AppPort),
		fmt.Sprintf("--dapr-http-port=%d", config.DaprHTTPPort),
		fmt.Sprintf("--dapr-grpc-port=%d", config.DaprGRPCPort),
		fmt.Sprintf("--log-level=%s", config.LogLevel),
		fmt.Sprintf("--app-max-concurrency=%d", config.MaxConcurrency),
		fmt.Sprintf("--placement-host-address=%s", config.PlacementHostAddress),
		"--dapr-listen-addresses=0.0.0.0",
		"--components-path=/tmp/dapr-components", // CRITICAL: Load state store and pub/sub components
	}
	
	if config.EnableProfiling {
		args = append(args, "--enable-profiling")
		args = append(args, fmt.Sprintf("--profile-port=%d", config.DaprProfilePort))
	}
	
	if config.EnableMetrics {
		args = append(args, "--enable-metrics")
		args = append(args, fmt.Sprintf("--metrics-port=%d", config.DaprMetricsPort))
	}
	
	return args
}

// Private helper methods

func (d *UnifiedDaprSidecarManager) calculateDaprHTTPPort(appPort int) int {
	// Use a base port range for Dapr HTTP ports based on application port
	switch {
	case appPort >= 9000 && appPort < 10000: // Gateways
		return 50000 + (appPort - 9000)
	case appPort >= 3000 && appPort < 4000: // Content services
		return 50010 + (appPort - 3000)
	case appPort >= 3100 && appPort < 3200: // Inquiry services
		return 50020 + (appPort - 3100)
	case appPort >= 3200 && appPort < 3300: // Notification services
		return 50030 + (appPort - 3200)
	default:
		return 50100 + (appPort % 100) // Fallback range
	}
}

func (d *UnifiedDaprSidecarManager) calculateDaprGRPCPort(appPort int) int {
	// GRPC port is typically HTTP port + 10000
	return d.calculateDaprHTTPPort(appPort) + 10000
}

func (d *UnifiedDaprSidecarManager) getPlacementHostAddress() string {
	switch d.Environment {
	case "development":
		return "localhost:50005"
	case "staging":
		return "dapr-control-plane-staging.azurecontainerapp.io:50005"
	case "production":
		return "dapr-control-plane-production.azurecontainerapp.io:50005"
	default:
		return "localhost:50005"
	}
}

func (d *UnifiedDaprSidecarManager) getLogLevel() string {
	switch d.Environment {
	case "development":
		return "debug"
	case "staging":
		return "info"
	case "production":
		return "warn"
	default:
		return "info"
	}
}

func (d *UnifiedDaprSidecarManager) shouldEnableProfiling() bool {
	return d.Environment == "development"
}

func (d *UnifiedDaprSidecarManager) getMaxConcurrency() int {
	switch d.Environment {
	case "development":
		return -1 // Unlimited
	case "staging":
		return 100
	case "production":
		return 1000
	default:
		return 100
	}
}

func (d *UnifiedDaprSidecarManager) getSidecarCPULimit() string {
	switch d.Environment {
	case "development":
		return "200m"
	case "staging":
		return "500m"
	case "production":
		return "1000m"
	default:
		return "500m"
	}
}

func (d *UnifiedDaprSidecarManager) getSidecarMemoryLimit() string {
	switch d.Environment {
	case "development":
		return "128Mi"
	case "staging":
		return "256Mi"
	case "production":
		return "512Mi"
	default:
		return "256Mi"
	}
}

func (d *UnifiedDaprSidecarManager) validateAppID(appID string) error {
	if len(appID) == 0 {
		return fmt.Errorf("app ID cannot be empty")
	}
	
	if len(appID) > 60 {
		return fmt.Errorf("app ID cannot be longer than 60 characters")
	}
	
	// Validate that app ID contains only valid characters for Dapr
	// Dapr app IDs should contain only alphanumeric characters and hyphens
	for i, char := range appID {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '-') {
			return fmt.Errorf("app ID contains invalid character '%c' at position %d", char, i)
		}
	}
	
	// App ID cannot start or end with a hyphen
	if appID[0] == '-' || appID[len(appID)-1] == '-' {
		return fmt.Errorf("app ID cannot start or end with a hyphen")
	}
	
	return nil
}