package platform

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type PlatformArgs struct {
	Environment string
}

type PlatformComponent struct {
	pulumi.ResourceState

	DaprControlPlaneURL     pulumi.StringOutput `pulumi:"daprControlPlaneURL"`
	DaprPlacementService    pulumi.StringOutput `pulumi:"daprPlacementService"`
	ContainerOrchestrator   pulumi.StringOutput `pulumi:"containerOrchestrator"`
	ServiceMeshEnabled      pulumi.BoolOutput   `pulumi:"serviceMeshEnabled"`
	NetworkingConfiguration pulumi.MapOutput    `pulumi:"networkingConfiguration"`
	SecurityPolicies        pulumi.MapOutput    `pulumi:"securityPolicies"`
	HealthCheckEnabled      pulumi.BoolOutput   `pulumi:"healthCheckEnabled"`
}

func NewPlatformComponent(ctx *pulumi.Context, name string, args *PlatformArgs, opts ...pulumi.ResourceOption) (*PlatformComponent, error) {
	component := &PlatformComponent{}
	
	err := ctx.RegisterComponentResource("international-center:platform:Platform", name, component, opts...)
	if err != nil {
		return nil, err
	}

	// Deploy DAPR component
	dapr, err := NewDaprComponent(ctx, "dapr", &DaprArgs{
		Environment: args.Environment,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy container orchestration component
	orchestration, err := NewOrchestrationComponent(ctx, "orchestration", &OrchestrationArgs{
		Environment: args.Environment,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy networking component
	networking, err := NewNetworkingComponent(ctx, "networking", &NetworkingArgs{
		Environment: args.Environment,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy security component
	security, err := NewSecurityComponent(ctx, "security", &SecurityArgs{
		Environment: args.Environment,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Configure environment-specific settings
	var serviceMeshEnabled, healthCheckEnabled pulumi.BoolOutput
	
	switch args.Environment {
	case "development":
		serviceMeshEnabled = pulumi.Bool(true).ToBoolOutput()
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
	case "staging":
		serviceMeshEnabled = pulumi.Bool(true).ToBoolOutput()
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
	case "production":
		serviceMeshEnabled = pulumi.Bool(true).ToBoolOutput()
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
	default:
		serviceMeshEnabled = pulumi.Bool(true).ToBoolOutput()
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
	}

	// Set component outputs
	component.DaprControlPlaneURL = dapr.ControlPlaneURL
	component.DaprPlacementService = dapr.PlacementService
	component.ContainerOrchestrator = orchestration.OrchestratorType
	component.ServiceMeshEnabled = serviceMeshEnabled
	component.NetworkingConfiguration = networking.Configuration
	component.SecurityPolicies = security.Policies
	component.HealthCheckEnabled = healthCheckEnabled

	// Register outputs
	ctx.Export("platform:dapr_control_plane_url", component.DaprControlPlaneURL)
	ctx.Export("platform:dapr_placement_service", component.DaprPlacementService)
	ctx.Export("platform:container_orchestrator", component.ContainerOrchestrator)
	ctx.Export("platform:service_mesh_enabled", component.ServiceMeshEnabled)

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"daprControlPlaneURL":     component.DaprControlPlaneURL,
		"daprPlacementService":    component.DaprPlacementService,
		"containerOrchestrator":   component.ContainerOrchestrator,
		"serviceMeshEnabled":      component.ServiceMeshEnabled,
		"networkingConfiguration": component.NetworkingConfiguration,
		"securityPolicies":        component.SecurityPolicies,
		"healthCheckEnabled":      component.HealthCheckEnabled,
	}); err != nil {
		return nil, err
	}

	return component, nil
}