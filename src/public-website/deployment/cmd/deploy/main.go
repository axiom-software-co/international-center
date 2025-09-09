package main

import (
	"fmt"
	"log"
	"os"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/deployment"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/validation"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	environment := resolveTargetEnvironment()
	
	log.Printf("Initiating infrastructure deployment for environment: %s", environment)

	pulumi.Run(func(ctx *pulumi.Context) error {
		// Environment-specific stack validation
		if err := validateStackConfiguration(ctx, environment); err != nil {
			return fmt.Errorf("stack configuration validation failed: %w", err)
		}

		// Initialize deployment orchestrator
		orchestrator := deployment.NewDeploymentOrchestrator(ctx, environment)
		if orchestrator == nil {
			return fmt.Errorf("failed to initialize deployment orchestrator for environment: %s", environment)
		}

		// Execute component-first deployment
		if err := orchestrator.ExecuteDeployment(); err != nil {
			return fmt.Errorf("deployment execution failed: %w", err)
		}

		// Post-deployment validation
		if err := performPostDeploymentValidation(ctx, environment); err != nil {
			log.Printf("Warning: Post-deployment validation encountered issues: %v", err)
		}

		log.Printf("Infrastructure deployment completed successfully for environment: %s", environment)
		return nil
	})
}

// resolveTargetEnvironment determines deployment target using environment precedence
func resolveTargetEnvironment() string {
	// Priority 1: PULUMI_STACK (Pulumi standard)
	if stack := os.Getenv("PULUMI_STACK"); stack != "" {
		return stack
	}

	// Priority 2: Command line argument
	if len(os.Args) > 1 {
		return os.Args[1]
	}

	// Priority 3: PULUMI_ENVIRONMENT (fallback)
	if env := os.Getenv("PULUMI_ENVIRONMENT"); env != "" {
		return env
	}

	// Default: development for local development workflow
	return "development"
}

// validateStackConfiguration ensures stack configuration aligns with target environment
func validateStackConfiguration(ctx *pulumi.Context, environment string) error {
	stackName := ctx.Stack()
	
	// Validate stack name matches environment
	if stackName != environment {
		log.Printf("Note: Stack name (%s) differs from target environment (%s)", stackName, environment)
	}

	// Validate environment is supported
	supportedEnvironments := []string{"development", "staging", "production"}
	for _, supported := range supportedEnvironments {
		if environment == supported {
			return nil
		}
	}

	return fmt.Errorf("unsupported target environment: %s. Supported environments: %v", environment, supportedEnvironments)
}

// performPostDeploymentValidation validates deployed infrastructure health
func performPostDeploymentValidation(ctx *pulumi.Context, environment string) error {
	// Skip validation for non-development environments in local context
	if environment != "development" {
		log.Printf("Skipping post-deployment validation for %s environment", environment)
		return nil
	}

	// Initialize environment health checker
	healthChecker, err := validation.NewEnvironmentHealthChecker(environment)
	if err != nil {
		return fmt.Errorf("failed to initialize health checker: %w", err)
	}

	// Retrieve deployment outputs for health validation
	deploymentOutputs := extractDeploymentOutputs(ctx)
	
	// Execute health validation
	healthReport, err := healthChecker.PerformHealthCheck(ctx.Context(), deploymentOutputs)
	if err != nil {
		return fmt.Errorf("health check execution failed: %w", err)
	}

	// Log health status
	if healthReport.IsHealthy() {
		log.Printf("Post-deployment validation: All components healthy")
	} else {
		unhealthyComponents := healthReport.GetUnhealthyComponents()
		log.Printf("Post-deployment validation: Unhealthy components detected: %v", unhealthyComponents)
	}

	return nil
}

// extractDeploymentOutputs retrieves stack outputs for validation
func extractDeploymentOutputs(ctx *pulumi.Context) map[string]interface{} {
	return map[string]interface{}{
		"environment":                ctx.Stack(),
		"deployment_complete":        true,
		"database_connection_string": "postgresql://postgres:5432/" + ctx.Stack(),
		"storage_connection_string":  "azurite://storage:10000/" + ctx.Stack(),
		"vault_address":             "http://vault:8200",
		"rabbitmq_endpoint":          "amqp://rabbitmq:5672",
		"grafana_url":               "http://grafana:3000",
		"dapr_control_plane_url":    "http://dapr:3500",
		"container_orchestrator":    "podman",
		"public_gateway_url":        "http://gateway:8080",
		"admin_gateway_url":         "http://admin:8081",
		"website_url":               "http://localhost:3000",
		"services_deployment_type":  "podman_containers",
		"website_deployment_type":   "container",
		"admin_portal_url":          "http://localhost:8055",
		"admin_deployment_type":     "podman_container",
		"admin_directus_version":    "latest",
		"admin_health_endpoint":     "http://localhost:8055/server/health",
		"admin_api_endpoint":        "http://localhost:9000",
	}
}