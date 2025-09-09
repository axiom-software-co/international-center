package builders

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ComponentBuilder struct {
	environment string
	ctx         *pulumi.Context
}

type InfrastructureComponent struct {
	DatabaseConnectionString pulumi.StringOutput
	StorageConnectionString  pulumi.StringOutput
	VaultAddress            pulumi.StringOutput
	RabbitMQEndpoint        pulumi.StringOutput
	GrafanaURL              pulumi.StringOutput
	HealthCheckEnabled      pulumi.BoolOutput
	SecurityPolicies        pulumi.MapOutput
	AuditLogging           pulumi.BoolOutput
}

type PlatformComponent struct {
	DaprControlPlaneURL      pulumi.StringOutput
	DaprPlacementService     pulumi.StringOutput
	ContainerOrchestrator    pulumi.StringOutput
	ServiceMeshEnabled       pulumi.BoolOutput
	NetworkingConfiguration  pulumi.MapOutput
	SecurityPolicies         pulumi.MapOutput
	HealthCheckEnabled       pulumi.BoolOutput
}

type ServicesComponent struct {
	ContentServices       pulumi.ArrayOutput
	InquiriesServices     pulumi.ArrayOutput
	NotificationServices  pulumi.ArrayOutput
	GatewayServices       pulumi.ArrayOutput
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
	StaticAssets         pulumi.ArrayOutput
}

func NewComponentBuilder(ctx *pulumi.Context, environment string) *ComponentBuilder {
	return &ComponentBuilder{
		environment: environment,
		ctx:         ctx,
	}
}

func (cb *ComponentBuilder) BuildInfrastructure() (*InfrastructureComponent, error) {
	// For now, return mock component outputs that match the expected structure
	// This would be implemented with actual Pulumi infrastructure resources
	return &InfrastructureComponent{
		DatabaseConnectionString: cb.ctx.Export("database_connection_string", pulumi.String("postgresql://postgres:5432/development")).(pulumi.StringOutput),
		StorageConnectionString:  cb.ctx.Export("storage_connection_string", pulumi.String("s3://development-storage")).(pulumi.StringOutput),
		VaultAddress:            cb.ctx.Export("vault_address", pulumi.String("http://vault:8200")).(pulumi.StringOutput),
		RabbitMQEndpoint:        cb.ctx.Export("rabbitmq_endpoint", pulumi.String("amqp://rabbitmq:5672")).(pulumi.StringOutput),
		GrafanaURL:              cb.ctx.Export("grafana_url", pulumi.String("http://grafana:3000")).(pulumi.StringOutput),
		HealthCheckEnabled:      cb.ctx.Export("health_check_enabled", pulumi.Bool(true)).(pulumi.BoolOutput),
		SecurityPolicies:        cb.ctx.Export("security_policies", pulumi.Map{}).(pulumi.MapOutput),
		AuditLogging:           cb.ctx.Export("audit_logging", pulumi.Bool(false)).(pulumi.BoolOutput),
	}, nil
}

func (cb *ComponentBuilder) BuildPlatform() (*PlatformComponent, error) {
	// For now, return mock component outputs that match the expected structure
	return &PlatformComponent{
		DaprControlPlaneURL:      cb.ctx.Export("dapr_control_plane_url", pulumi.String("http://dapr:3500")).(pulumi.StringOutput),
		DaprPlacementService:     cb.ctx.Export("dapr_placement_service", pulumi.String("dapr-placement:50006")).(pulumi.StringOutput),
		ContainerOrchestrator:    cb.ctx.Export("container_orchestrator", pulumi.String("podman")).(pulumi.StringOutput),
		ServiceMeshEnabled:       cb.ctx.Export("service_mesh_enabled", pulumi.Bool(true)).(pulumi.BoolOutput),
		NetworkingConfiguration:  cb.ctx.Export("networking_configuration", pulumi.Map{"mode": pulumi.String("bridge")}).(pulumi.MapOutput),
		SecurityPolicies:         cb.ctx.Export("security_policies", pulumi.Map{}).(pulumi.MapOutput),
		HealthCheckEnabled:       cb.ctx.Export("health_check_enabled", pulumi.Bool(true)).(pulumi.BoolOutput),
	}, nil
}

func (cb *ComponentBuilder) BuildServices(infraOutputs, platformOutputs pulumi.Map) (*ServicesComponent, error) {
	// For now, return mock component outputs that match the expected structure
	return &ServicesComponent{
		ContentServices:       cb.ctx.Export("content_services", pulumi.Array{}).(pulumi.ArrayOutput),
		InquiriesServices:     cb.ctx.Export("inquiries_services", pulumi.Array{}).(pulumi.ArrayOutput),
		NotificationServices:  cb.ctx.Export("notification_services", pulumi.Array{}).(pulumi.ArrayOutput),
		GatewayServices:       cb.ctx.Export("gateway_services", pulumi.Array{}).(pulumi.ArrayOutput),
		PublicGatewayURL:      cb.ctx.Export("public_gateway_url", pulumi.String("http://gateway:8080")).(pulumi.StringOutput),
		AdminGatewayURL:       cb.ctx.Export("admin_gateway_url", pulumi.String("http://admin:8081")).(pulumi.StringOutput),
		DeploymentType:        cb.ctx.Export("deployment_type", pulumi.String("podman_containers")).(pulumi.StringOutput),
		HealthCheckEnabled:    cb.ctx.Export("health_check_enabled", pulumi.Bool(true)).(pulumi.BoolOutput),
		DaprSidecarEnabled:    cb.ctx.Export("dapr_sidecar_enabled", pulumi.Bool(true)).(pulumi.BoolOutput),
		ScalingPolicy:         cb.ctx.Export("scaling_policy", pulumi.String("manual")).(pulumi.StringOutput),
	}, nil
}

func (cb *ComponentBuilder) BuildWebsite(infraOutputs, platformOutputs, servicesOutputs pulumi.Map) (*WebsiteComponent, error) {
	// For now, return mock component outputs that match the expected structure
	return &WebsiteComponent{
		WebsiteURL:            cb.ctx.Export("website_url", pulumi.String("http://localhost:3000")).(pulumi.StringOutput),
		DeploymentType:        cb.ctx.Export("deployment_type", pulumi.String("container")).(pulumi.StringOutput),
		CDNEnabled:           cb.ctx.Export("cdn_enabled", pulumi.Bool(false)).(pulumi.BoolOutput),
		SSLEnabled:           cb.ctx.Export("ssl_enabled", pulumi.Bool(false)).(pulumi.BoolOutput),
		CacheConfiguration:   cb.ctx.Export("cache_configuration", pulumi.Map{"enabled": pulumi.Bool(false)}).(pulumi.MapOutput),
		HealthCheckEnabled:   cb.ctx.Export("health_check_enabled", pulumi.Bool(true)).(pulumi.BoolOutput),
		ContainerConfig:      cb.ctx.Export("container_config", pulumi.Map{"replicas": pulumi.Int(1)}).(pulumi.MapOutput),
		StaticAssets:         cb.ctx.Export("static_assets", pulumi.Array{}).(pulumi.ArrayOutput),
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