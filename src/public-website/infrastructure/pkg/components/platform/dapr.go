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
	
	err := ctx.RegisterComponentResource("international-center:platform:Dapr", name, component, opts...)
	if err != nil {
		return nil, err
	}

	var controlPlaneURL, placementService, healthEndpoint pulumi.StringOutput
	var sidecarEnabled pulumi.BoolOutput

	switch args.Environment {
	case "development":
		controlPlaneURL = pulumi.String("http://localhost:50001").ToStringOutput()
		placementService = pulumi.String("localhost:50005").ToStringOutput()
		sidecarEnabled = pulumi.Bool(true).ToBoolOutput()
		healthEndpoint = pulumi.String("http://localhost:50001/v1.0/healthz").ToStringOutput()
	case "staging":
		controlPlaneURL = pulumi.String("https://dapr-staging.azurecontainerapp.io").ToStringOutput()
		placementService = pulumi.String("dapr-staging.azurecontainerapp.io:50005").ToStringOutput()
		sidecarEnabled = pulumi.Bool(true).ToBoolOutput()
		healthEndpoint = pulumi.String("https://dapr-staging.azurecontainerapp.io/v1.0/healthz").ToStringOutput()
	case "production":
		controlPlaneURL = pulumi.String("https://dapr-production.azurecontainerapp.io").ToStringOutput()
		placementService = pulumi.String("dapr-production.azurecontainerapp.io:50005").ToStringOutput()
		sidecarEnabled = pulumi.Bool(true).ToBoolOutput()
		healthEndpoint = pulumi.String("https://dapr-production.azurecontainerapp.io/v1.0/healthz").ToStringOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.ControlPlaneURL = controlPlaneURL
	component.PlacementService = placementService
	component.SidecarEnabled = sidecarEnabled
	component.HealthEndpoint = healthEndpoint

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"controlPlaneURL":  component.ControlPlaneURL,
		"placementService": component.PlacementService,
		"sidecarEnabled":   component.SidecarEnabled,
		"healthEndpoint":   component.HealthEndpoint,
	}); err != nil {
		return nil, err
	}

	return component, nil
}