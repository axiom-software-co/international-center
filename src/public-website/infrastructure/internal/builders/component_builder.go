package builders

import (
	"fmt"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/pkg/components/infrastructure"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/pkg/components/platform"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/pkg/components/services"
	website "github.com/axiom-software-co/international-center/src/public-website/infrastructure/pkg/components/public-website"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ComponentBuilder struct {
	environment string
	ctx         *pulumi.Context
}

type InfrastructureComponent struct {
	DatabaseEndpoint        pulumi.StringOutput
	StorageEndpoint         pulumi.StringOutput
	VaultEndpoint           pulumi.StringOutput
	MessagingEndpoint       pulumi.StringOutput
	ObservabilityEndpoint   pulumi.StringOutput
	HealthCheckEnabled      pulumi.BoolOutput
	SecurityPolicies        pulumi.BoolOutput
	AuditLogging           pulumi.BoolOutput
}

type PlatformComponent struct {
	DaprEndpoint            pulumi.StringOutput
	OrchestrationEndpoint   pulumi.StringOutput
	ServiceMeshEnabled      pulumi.BoolOutput
	NetworkingConfig        pulumi.MapOutput
	SecurityConfig          pulumi.MapOutput
	HealthCheckEnabled      pulumi.BoolOutput
}

type ServicesComponent struct {
	ContentServices       pulumi.MapOutput
	InquiriesServices     pulumi.MapOutput
	NotificationServices  pulumi.MapOutput
	GatewayServices       pulumi.MapOutput
	PublicGatewayURL      pulumi.StringOutput
	AdminGatewayURL       pulumi.StringOutput
	DeploymentType        pulumi.StringOutput
	HealthCheckEnabled    pulumi.BoolOutput
	DaprSidecarEnabled    pulumi.BoolOutput
	ScalingPolicy         pulumi.StringOutput
}

type WebsiteComponent struct {
	WebsiteURL            pulumi.StringOutput
	DeploymentType        pulumi.StringOutput
	CDNEnabled           pulumi.BoolOutput
	SSLEnabled           pulumi.BoolOutput
	CacheConfiguration   pulumi.MapOutput
	HealthCheckEnabled   pulumi.BoolOutput
	ContainerConfig      pulumi.MapOutput
	StaticAssets         pulumi.MapOutput
}

func NewComponentBuilder(ctx *pulumi.Context, environment string) *ComponentBuilder {
	return &ComponentBuilder{
		environment: environment,
		ctx:         ctx,
	}
}

func (cb *ComponentBuilder) BuildInfrastructure() (*InfrastructureComponent, error) {
	// Deploy actual infrastructure using the infrastructure package
	infraComponent, err := infrastructure.NewInfrastructureComponent(cb.ctx, "infrastructure", &infrastructure.InfrastructureArgs{
		Environment: cb.environment,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create infrastructure component: %w", err)
	}

	// Export the outputs to the stack (only if context supports it)
	if cb.canExport() {
		cb.ctx.Export("database_connection_string", infraComponent.DatabaseEndpoint)
		cb.ctx.Export("storage_connection_string", infraComponent.StorageEndpoint)
		cb.ctx.Export("vault_address", infraComponent.VaultEndpoint)
		cb.ctx.Export("rabbitmq_endpoint", infraComponent.MessagingEndpoint)
		cb.ctx.Export("grafana_url", infraComponent.ObservabilityEndpoint)
		cb.ctx.Export("health_check_enabled", infraComponent.HealthCheckEnabled)
		cb.ctx.Export("security_policies", infraComponent.SecurityPolicies)
		cb.ctx.Export("audit_logging", infraComponent.AuditLogging)
	}

	return &InfrastructureComponent{
		DatabaseEndpoint:        infraComponent.DatabaseEndpoint,
		StorageEndpoint:         infraComponent.StorageEndpoint,
		VaultEndpoint:           infraComponent.VaultEndpoint,
		MessagingEndpoint:       infraComponent.MessagingEndpoint,
		ObservabilityEndpoint:   infraComponent.ObservabilityEndpoint,
		HealthCheckEnabled:      infraComponent.HealthCheckEnabled,
		SecurityPolicies:        infraComponent.SecurityPolicies,
		AuditLogging:           infraComponent.AuditLogging,
	}, nil
}

func (cb *ComponentBuilder) BuildPlatform(infraOutputs pulumi.Map) (*PlatformComponent, error) {
	// Deploy actual platform using the platform package
	platformComponent, err := platform.NewPlatformComponent(cb.ctx, "platform", &platform.PlatformArgs{
		Environment: cb.environment,
		InfrastructureOutputs: infraOutputs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create platform component: %w", err)
	}

	// Export the outputs to the stack (only if context supports it)
	if cb.canExport() {
		cb.ctx.Export("dapr_control_plane_url", platformComponent.DaprEndpoint)
		cb.ctx.Export("container_orchestrator", platformComponent.OrchestrationEndpoint)
		cb.ctx.Export("service_mesh_enabled", platformComponent.ServiceMeshEnabled)
		cb.ctx.Export("networking_configuration", platformComponent.NetworkingConfig)
		cb.ctx.Export("security_policies", platformComponent.SecurityConfig)
		cb.ctx.Export("health_check_enabled", platformComponent.HealthCheckEnabled)
	}

	return &PlatformComponent{
		DaprEndpoint:            platformComponent.DaprEndpoint,
		OrchestrationEndpoint:   platformComponent.OrchestrationEndpoint,
		ServiceMeshEnabled:      platformComponent.ServiceMeshEnabled,
		NetworkingConfig:        platformComponent.NetworkingConfig,
		SecurityConfig:          platformComponent.SecurityConfig,
		HealthCheckEnabled:      platformComponent.HealthCheckEnabled,
	}, nil
}

func (cb *ComponentBuilder) BuildServices(infraOutputs, platformOutputs pulumi.Map) (*ServicesComponent, error) {
	// Deploy actual services using the services package
	servicesComponent, err := services.NewServicesComponent(cb.ctx, "services", &services.ServicesArgs{
		Environment:           cb.environment,
		InfrastructureOutputs: infraOutputs,
		PlatformOutputs:      platformOutputs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create services component: %w", err)
	}

	// Export the outputs to the stack (only if context supports it)
	if cb.canExport() {
		cb.ctx.Export("content_services", servicesComponent.ContentServices)
		cb.ctx.Export("inquiries_services", servicesComponent.InquiriesServices)
		cb.ctx.Export("notification_services", servicesComponent.NotificationServices)
		cb.ctx.Export("gateway_services", servicesComponent.GatewayServices)
		cb.ctx.Export("public_gateway_url", servicesComponent.PublicGatewayURL)
		cb.ctx.Export("admin_gateway_url", servicesComponent.AdminGatewayURL)
		cb.ctx.Export("deployment_type", servicesComponent.DeploymentType)
		cb.ctx.Export("health_check_enabled", servicesComponent.HealthCheckEnabled)
		cb.ctx.Export("dapr_sidecar_enabled", servicesComponent.DaprSidecarEnabled)
		cb.ctx.Export("scaling_policy", servicesComponent.ScalingPolicy)
	}

	return &ServicesComponent{
		ContentServices:       servicesComponent.ContentServices,
		InquiriesServices:     servicesComponent.InquiriesServices,
		NotificationServices:  servicesComponent.NotificationServices,
		GatewayServices:       servicesComponent.GatewayServices,
		PublicGatewayURL:      servicesComponent.PublicGatewayURL,
		AdminGatewayURL:       servicesComponent.AdminGatewayURL,
		DeploymentType:        servicesComponent.DeploymentType,
		HealthCheckEnabled:    servicesComponent.HealthCheckEnabled,
		DaprSidecarEnabled:    servicesComponent.DaprSidecarEnabled,
		ScalingPolicy:         servicesComponent.ScalingPolicy,
	}, nil
}

func (cb *ComponentBuilder) BuildWebsite(infraOutputs, platformOutputs, servicesOutputs pulumi.Map) (*WebsiteComponent, error) {
	// Deploy actual website using the website package
	websiteComponent, err := website.NewWebsiteComponent(cb.ctx, "website", &website.WebsiteArgs{
		Environment:           cb.environment,
		InfrastructureOutputs: infraOutputs,
		PlatformOutputs:      platformOutputs,
		ServicesOutputs:      servicesOutputs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create website component: %w", err)
	}

	// Export the outputs to the stack (only if context supports it)
	if cb.canExport() {
		cb.ctx.Export("website_url", websiteComponent.WebsiteURL)
		cb.ctx.Export("deployment_type", websiteComponent.DeploymentType)
		cb.ctx.Export("cdn_enabled", websiteComponent.CDNEnabled)
		cb.ctx.Export("ssl_enabled", websiteComponent.SSLEnabled)
		cb.ctx.Export("cache_configuration", websiteComponent.CacheConfiguration)
		cb.ctx.Export("health_check_enabled", websiteComponent.HealthCheckEnabled)
		cb.ctx.Export("container_config", websiteComponent.ContainerConfig)
		cb.ctx.Export("static_assets", websiteComponent.StaticAssets)
	}

	return &WebsiteComponent{
		WebsiteURL:            websiteComponent.WebsiteURL,
		DeploymentType:        websiteComponent.DeploymentType,
		CDNEnabled:           websiteComponent.CDNEnabled,
		SSLEnabled:           websiteComponent.SSLEnabled,
		CacheConfiguration:   websiteComponent.CacheConfiguration,
		HealthCheckEnabled:   websiteComponent.HealthCheckEnabled,
		ContainerConfig:      websiteComponent.ContainerConfig,
		StaticAssets:         websiteComponent.StaticAssets,
	}, nil
}

func (cb *ComponentBuilder) ValidateEnvironment() error {
	validEnvironments := []string{"development", "staging", "production"}
	for _, env := range validEnvironments {
		if cb.environment == env {
			return nil
		}
	}
	return fmt.Errorf("invalid environment: %s. Valid environments: %v", cb.environment, validEnvironments)
}

func (cb *ComponentBuilder) canExport() bool {
	// Check if context is properly initialized and can handle export operations
	// For unit tests, we may have a mock context that doesn't support exports
	if cb.ctx == nil {
		return false
	}
	
	// Use a defer/recover pattern to safely test if exports work
	canExportSafely := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				// If panic occurred, exports are not safe
				canExportSafely = false
			}
		}()
		
		// Try to detect if this is a real Pulumi context vs a mock
		// Mock contexts created with &pulumi.Context{} will panic on export
		// Real contexts will have internal state initialized
		// We use a simple test - try to export a dummy value
		testOutput := pulumi.String("test").ToStringOutput()
		cb.ctx.Export("__test_export_capability", testOutput)
		canExportSafely = true
	}()
	
	return canExportSafely
}