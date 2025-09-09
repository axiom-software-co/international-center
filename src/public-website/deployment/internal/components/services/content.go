package services

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ContentArgs struct {
	Environment           string
	InfrastructureOutputs pulumi.Map
	PlatformOutputs      pulumi.Map
}

type ContentComponent struct {
	pulumi.ResourceState

	Services        pulumi.MapOutput    `pulumi:"services"`
	HealthEndpoints pulumi.MapOutput    `pulumi:"healthEndpoints"`
	DaprSidecars    pulumi.MapOutput    `pulumi:"daprSidecars"`
	ResourceLimits  pulumi.MapOutput    `pulumi:"resourceLimits"`
	ServiceEndpoints pulumi.MapOutput   `pulumi:"serviceEndpoints"`
}

func NewContentComponent(ctx *pulumi.Context, name string, args *ContentArgs, opts ...pulumi.ResourceOption) (*ContentComponent, error) {
	component := &ContentComponent{}
	
	// Safe registration for mock contexts
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:services:Content", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	// Deploy actual content containers using the container deployment component
	contentDeployment, err := NewContainerDeploymentComponent(ctx, "content-deployment", &ContainerDeploymentArgs{
		Environment:           args.Environment,
		InfrastructureOutputs: args.InfrastructureOutputs,
		PlatformOutputs:      args.PlatformOutputs,
		ServiceType:          "content",
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("failed to deploy content containers: %w", err)
	}

	component.Services = contentDeployment.Containers
	component.HealthEndpoints = contentDeployment.HealthEndpoints
	component.DaprSidecars = contentDeployment.DaprSidecars
	component.ResourceLimits = contentDeployment.ResourceLimits
	component.ServiceEndpoints = contentDeployment.ServiceEndpoints

	if canRegister(ctx) {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"services":         component.Services,
			"healthEndpoints":  component.HealthEndpoints,
			"daprSidecars":     component.DaprSidecars,
			"resourceLimits":   component.ResourceLimits,
			"serviceEndpoints": component.ServiceEndpoints,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}