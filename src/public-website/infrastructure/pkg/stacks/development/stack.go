package development

import (
	"log"

	"github.com/axiom-software-co/international-center/src/cicd/pkg/components/infrastructure"
	"github.com/axiom-software-co/international-center/src/cicd/pkg/components/platform"
	"github.com/axiom-software-co/international-center/src/cicd/pkg/components/services"
	"github.com/axiom-software-co/international-center/src/cicd/pkg/components/website"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployStack(ctx *pulumi.Context) error {
	log.Printf("Deploying Development Stack for International Center")

	// Step 1: Deploy Infrastructure Components
	log.Printf("Step 1: Deploying Infrastructure Components")
	infrastructureComponent, err := infrastructure.NewInfrastructureComponent(ctx, "infrastructure", &infrastructure.InfrastructureArgs{
		Environment: "development",
	})
	if err != nil {
		return err
	}

	// Step 2: Deploy Platform Components  
	log.Printf("Step 2: Deploying Platform Components")
	platformComponent, err := platform.NewPlatformComponent(ctx, "platform", &platform.PlatformArgs{
		Environment: "development",
	})
	if err != nil {
		return err
	}

	// Step 3: Deploy Services Components
	log.Printf("Step 3: Deploying Services Components")
	servicesComponent, err := services.NewServicesComponent(ctx, "services", &services.ServicesArgs{
		Environment: "development",
		InfrastructureOutputs: pulumi.Map{
			"database_connection_string": infrastructureComponent.DatabaseConnectionString,
			"storage_connection_string":  infrastructureComponent.StorageConnectionString,
			"vault_address":             infrastructureComponent.VaultAddress,
			"rabbitmq_endpoint":         infrastructureComponent.RabbitMQEndpoint,
			"grafana_url":               infrastructureComponent.GrafanaURL,
		},
		PlatformOutputs: pulumi.Map{
			"dapr_control_plane_url":   platformComponent.DaprControlPlaneURL,
			"dapr_placement_service":   platformComponent.DaprPlacementService,
			"container_orchestrator":   platformComponent.ContainerOrchestrator,
			"service_mesh_enabled":     platformComponent.ServiceMeshEnabled,
			"networking_configuration": platformComponent.NetworkingConfiguration,
		},
	})
	if err != nil {
		return err
	}

	// Step 4: Deploy Website Components
	log.Printf("Step 4: Deploying Website Components")
	websiteComponent, err := website.NewWebsiteComponent(ctx, "website", &website.WebsiteArgs{
		Environment: "development",
		InfrastructureOutputs: pulumi.Map{
			"database_connection_string": infrastructureComponent.DatabaseConnectionString,
			"storage_connection_string":  infrastructureComponent.StorageConnectionString,
		},
		PlatformOutputs: pulumi.Map{
			"dapr_control_plane_url": platformComponent.DaprControlPlaneURL,
			"container_orchestrator": platformComponent.ContainerOrchestrator,
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

	// Export primary outputs for Development environment
	ctx.Export("environment", pulumi.String("development"))
	ctx.Export("deployment_complete", pulumi.Bool(true))
	
	// Infrastructure Outputs
	ctx.Export("database_connection_string", infrastructureComponent.DatabaseConnectionString)
	ctx.Export("storage_connection_string", infrastructureComponent.StorageConnectionString)
	ctx.Export("vault_address", infrastructureComponent.VaultAddress)
	ctx.Export("rabbitmq_endpoint", infrastructureComponent.RabbitMQEndpoint)
	ctx.Export("grafana_url", infrastructureComponent.GrafanaURL)
	
	// Platform Outputs
	ctx.Export("dapr_control_plane_url", platformComponent.DaprControlPlaneURL)
	ctx.Export("dapr_placement_service", platformComponent.DaprPlacementService)
	ctx.Export("container_orchestrator", platformComponent.ContainerOrchestrator)
	
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

	log.Printf("Development Stack deployment completed successfully")
	return nil
}