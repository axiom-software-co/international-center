package production

import (
	"log"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/pkg/components/infrastructure"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/pkg/components/platform"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/pkg/components/services"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/pkg/components/public-website"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployStack(ctx *pulumi.Context) error {
	log.Printf("Deploying Production Stack for International Center")

	// Step 1: Deploy Infrastructure Components
	log.Printf("Step 1: Deploying Infrastructure Components")
	infrastructureComponent, err := infrastructure.NewInfrastructureComponent(ctx, "infrastructure", &infrastructure.InfrastructureArgs{
		Environment: "production",
	})
	if err != nil {
		return err
	}

	// Step 2: Deploy Platform Components  
	log.Printf("Step 2: Deploying Platform Components")
	platformComponent, err := platform.NewPlatformComponent(ctx, "platform", &platform.PlatformArgs{
		Environment: "production",
		InfrastructureOutputs: pulumi.Map{
			"database_endpoint": infrastructureComponent.DatabaseEndpoint,
			"storage_endpoint":  infrastructureComponent.StorageEndpoint,
			"vault_endpoint":    infrastructureComponent.VaultEndpoint,
			"messaging_endpoint": infrastructureComponent.MessagingEndpoint,
			"observability_endpoint": infrastructureComponent.ObservabilityEndpoint,
		},
	})
	if err != nil {
		return err
	}

	// Step 3: Deploy Services Components
	log.Printf("Step 3: Deploying Services Components")
	servicesComponent, err := services.NewServicesComponent(ctx, "services", &services.ServicesArgs{
		Environment: "production",
		InfrastructureOutputs: pulumi.Map{
			"database_connection_string": infrastructureComponent.DatabaseEndpoint,
			"storage_connection_string":  infrastructureComponent.StorageEndpoint,
			"vault_address":             infrastructureComponent.VaultEndpoint,
			"rabbitmq_endpoint":         infrastructureComponent.MessagingEndpoint,
			"grafana_url":               infrastructureComponent.ObservabilityEndpoint,
		},
		PlatformOutputs: pulumi.Map{
			"dapr_control_plane_url":   platformComponent.DaprEndpoint,
			"container_orchestrator":   platformComponent.OrchestrationEndpoint,
			"service_mesh_enabled":     platformComponent.ServiceMeshEnabled,
			"networking_configuration": platformComponent.NetworkingConfig,
		},
	})
	if err != nil {
		return err
	}

	// Step 4: Deploy Website Components
	log.Printf("Step 4: Deploying Website Components")
	websiteComponent, err := website.NewWebsiteComponent(ctx, "website", &website.WebsiteArgs{
		Environment: "production",
		InfrastructureOutputs: pulumi.Map{
			"database_connection_string": infrastructureComponent.DatabaseEndpoint,
			"storage_connection_string":  infrastructureComponent.StorageEndpoint,
		},
		PlatformOutputs: pulumi.Map{
			"dapr_control_plane_url": platformComponent.DaprEndpoint,
			"container_orchestrator": platformComponent.OrchestrationEndpoint,
		},
		ServicesOutputs: pulumi.Map{
			"public_gateway_url": servicesComponent.PublicGatewayURL,
			"admin_gateway_url":  servicesComponent.AdminGatewayURL,
			"content_services":   servicesComponent.ContentServices,
		},
	})
	if err != nil {
		return err
	}

	// Export primary outputs for Production environment
	ctx.Export("environment", pulumi.String("production"))
	ctx.Export("deployment_complete", pulumi.Bool(true))
	
	// Infrastructure Outputs
	ctx.Export("database_connection_string", infrastructureComponent.DatabaseEndpoint)
	ctx.Export("storage_connection_string", infrastructureComponent.StorageEndpoint)
	ctx.Export("vault_address", infrastructureComponent.VaultEndpoint)
	ctx.Export("rabbitmq_endpoint", infrastructureComponent.MessagingEndpoint)
	ctx.Export("grafana_url", infrastructureComponent.ObservabilityEndpoint)
	
	// Platform Outputs
	ctx.Export("dapr_control_plane_url", platformComponent.DaprEndpoint)
	ctx.Export("container_orchestrator", platformComponent.OrchestrationEndpoint)
	
	// Services Outputs
	ctx.Export("public_gateway_url", servicesComponent.PublicGatewayURL)
	ctx.Export("admin_gateway_url", servicesComponent.AdminGatewayURL)
	ctx.Export("services_deployment_type", servicesComponent.DeploymentType)
	ctx.Export("services_scaling_policy", servicesComponent.ScalingPolicy)
	
	// Website Outputs
	ctx.Export("website_url", websiteComponent.WebsiteURL)
	ctx.Export("website_deployment_type", websiteComponent.DeploymentType)
	ctx.Export("cdn_enabled", websiteComponent.CDNEnabled)
	ctx.Export("ssl_enabled", websiteComponent.SSLEnabled)

	log.Printf("Production Stack deployment completed successfully")
	return nil
}