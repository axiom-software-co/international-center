package shared

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// EnvironmentConfiguration holds environment-specific deployment configuration
type EnvironmentConfiguration struct {
	Environment       string
	DeploymentOrder   []string
	OutputMappings    map[string]string
	ComponentSettings map[string]interface{}
}

// LoadEnvironmentConfiguration loads configuration for a specific environment
func LoadEnvironmentConfiguration(environment string, cfg *config.Config) (*EnvironmentConfiguration, error) {
	if environment == "" {
		return nil, fmt.Errorf("environment cannot be empty")
	}

	// Validate supported environments
	validEnvironments := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}

	if !validEnvironments[environment] {
		return nil, fmt.Errorf("unsupported environment: %s", environment)
	}

	// Define standard deployment order for RabbitMQ-only architecture
	deploymentOrder := []string{
		"database", "storage", "vault", "rabbitmq",
		"observability", "dapr", "services", "website",
	}

	// Define standard output mappings for RabbitMQ-only architecture
	outputMappings := map[string]string{
		"environment":                environment,
		"database_connection_string": "database.ConnectionString",
		"storage_connection_string":  "storage.ConnectionString",
		"vault_address":              "vault.VaultAddress",
		"rabbitmq_endpoint":          "rabbitmq.Endpoint",
		"grafana_url":                "observability.GrafanaURL",
		"dapr_control_plane_url":     "dapr.ControlPlaneURL",
		"public_gateway_url":         "services.PublicGatewayURL",
		"admin_gateway_url":          "services.AdminGatewayURL",
		"website_url":                "website.ServerURL",
	}

	// Environment-specific component settings (can be customized per environment)
	componentSettings := make(map[string]interface{})
	
	return &EnvironmentConfiguration{
		Environment:       environment,
		DeploymentOrder:   deploymentOrder,
		OutputMappings:    outputMappings,
		ComponentSettings: componentSettings,
	}, nil
}

// Validate checks if the configuration is valid
func (ec *EnvironmentConfiguration) Validate() error {
	if len(ec.DeploymentOrder) == 0 {
		return fmt.Errorf("deployment order cannot be empty")
	}

	if len(ec.OutputMappings) == 0 {
		return fmt.Errorf("output mappings cannot be empty")
	}

	return nil
}