package platform

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DaprArgs struct {
	Environment string
}

type DaprComponent struct {
	pulumi.ResourceState

	ControlPlaneURL  pulumi.StringOutput `pulumi:"controlPlaneURL"`
	PlacementService pulumi.StringOutput `pulumi:"placementService"`
	SidecarEnabled   pulumi.BoolOutput   `pulumi:"sidecarEnabled"`
	HealthEndpoint   pulumi.StringOutput `pulumi:"healthEndpoint"`
}

func NewDaprComponent(ctx *pulumi.Context, name string, args *DaprArgs, opts ...pulumi.ResourceOption) (*DaprComponent, error) {
	component := &DaprComponent{}
	
	// Safe registration for mock contexts
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:platform:Dapr", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	// Deploy actual Dapr containers using the deployment component
	daprDeployment, err := NewDaprDeploymentComponent(ctx, "dapr-deployment", &DaprDeploymentArgs{
		Environment: args.Environment,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Dapr containers: %w", err)
	}

	var controlPlaneURL, placementService, healthEndpoint pulumi.StringOutput
	var sidecarEnabled pulumi.BoolOutput

	switch args.Environment {
	case "development":
		controlPlaneURL = pulumi.String("http://localhost:3502").ToStringOutput()
		placementService = pulumi.String("localhost:50005").ToStringOutput()
		sidecarEnabled = pulumi.Bool(true).ToBoolOutput()
		healthEndpoint = pulumi.String("http://localhost:3502/v1.0/healthz").ToStringOutput()
	case "staging":
		controlPlaneURL = pulumi.String("https://dapr-control-plane-staging.azurecontainerapp.io").ToStringOutput()
		placementService = pulumi.String("dapr-control-plane-staging.azurecontainerapp.io:50005").ToStringOutput()
		sidecarEnabled = pulumi.Bool(true).ToBoolOutput()
		healthEndpoint = pulumi.String("https://dapr-control-plane-staging.azurecontainerapp.io/v1.0/healthz").ToStringOutput()
	case "production":
		controlPlaneURL = pulumi.String("https://dapr-control-plane-production.azurecontainerapp.io").ToStringOutput()
		placementService = pulumi.String("dapr-control-plane-production.azurecontainerapp.io:50005").ToStringOutput()
		sidecarEnabled = pulumi.Bool(true).ToBoolOutput()
		healthEndpoint = pulumi.String("https://dapr-control-plane-production.azurecontainerapp.io/v1.0/healthz").ToStringOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.ControlPlaneURL = controlPlaneURL
	component.PlacementService = placementService
	component.SidecarEnabled = sidecarEnabled
	component.HealthEndpoint = healthEndpoint

	// Export additional deployment information
	if canRegister(ctx) {
		ctx.Export("dapr:control_plane_container", daprDeployment.ControlPlaneContainer)
		ctx.Export("dapr:placement_container", daprDeployment.PlacementContainer)
		ctx.Export("dapr:sentry_container", daprDeployment.SentryContainer)
		ctx.Export("dapr:container_network", daprDeployment.ContainerNetwork)
		ctx.Export("dapr:health_endpoints", daprDeployment.HealthEndpoints)
	}

	if canRegister(ctx) {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"controlPlaneURL":        component.ControlPlaneURL,
			"placementService":       component.PlacementService,
			"sidecarEnabled":         component.SidecarEnabled,
			"healthEndpoint":         component.HealthEndpoint,
			"controlPlaneContainer":  daprDeployment.ControlPlaneContainer,
			"placementContainer":     daprDeployment.PlacementContainer,
			"sentryContainer":        daprDeployment.SentryContainer,
			"containerNetwork":       daprDeployment.ContainerNetwork,
			"healthEndpoints":        daprDeployment.HealthEndpoints,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}

