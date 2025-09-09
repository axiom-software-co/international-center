package services

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type InquiriesArgs struct {
	Environment           string
	InfrastructureOutputs pulumi.Map
	PlatformOutputs      pulumi.Map
}

type InquiriesComponent struct {
	pulumi.ResourceState

	Services         pulumi.MapOutput    `pulumi:"services"`
	HealthEndpoints  pulumi.MapOutput    `pulumi:"healthEndpoints"`
	DaprSidecars     pulumi.MapOutput    `pulumi:"daprSidecars"`
	ResourceLimits   pulumi.MapOutput    `pulumi:"resourceLimits"`
	ServiceEndpoints pulumi.MapOutput    `pulumi:"serviceEndpoints"`
}

func NewInquiriesComponent(ctx *pulumi.Context, name string, args *InquiriesArgs, opts ...pulumi.ResourceOption) (*InquiriesComponent, error) {
	component := &InquiriesComponent{}
	
	// Safe registration for mock contexts
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:services:Inquiries", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	// Deploy actual inquiries containers using the container deployment component
	inquiriesDeployment, err := NewContainerDeploymentComponent(ctx, "inquiries-deployment", &ContainerDeploymentArgs{
		Environment:           args.Environment,
		InfrastructureOutputs: args.InfrastructureOutputs,
		PlatformOutputs:      args.PlatformOutputs,
		ServiceType:          "inquiries",
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("failed to deploy inquiries containers: %w", err)
	}

	component.Services = inquiriesDeployment.Containers
	component.HealthEndpoints = inquiriesDeployment.HealthEndpoints
	component.DaprSidecars = inquiriesDeployment.DaprSidecars
	component.ResourceLimits = inquiriesDeployment.ResourceLimits
	component.ServiceEndpoints = inquiriesDeployment.ServiceEndpoints

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