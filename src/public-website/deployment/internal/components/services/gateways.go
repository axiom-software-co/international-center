package services

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type GatewayArgs struct {
	Environment           string
	InfrastructureOutputs pulumi.Map
	PlatformOutputs      pulumi.Map
}

type GatewayComponent struct {
	pulumi.ResourceState

	Services         pulumi.MapOutput    `pulumi:"services"`
	PublicGatewayURL  pulumi.StringOutput `pulumi:"publicGatewayURL"`
	AdminGatewayURL   pulumi.StringOutput `pulumi:"adminGatewayURL"`
	HealthEndpoints   pulumi.MapOutput    `pulumi:"healthEndpoints"`
	DaprSidecars     pulumi.MapOutput    `pulumi:"daprSidecars"`
	ResourceLimits   pulumi.MapOutput    `pulumi:"resourceLimits"`
}

func NewGatewayComponent(ctx *pulumi.Context, name string, args *GatewayArgs, opts ...pulumi.ResourceOption) (*GatewayComponent, error) {
	component := &GatewayComponent{}
	
	// Safe registration for mock contexts
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:services:Gateway", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	// Deploy actual gateway containers using the container deployment component
	gatewayDeployment, err := NewContainerDeploymentComponent(ctx, "gateway-deployment", &ContainerDeploymentArgs{
		Environment:           args.Environment,
		InfrastructureOutputs: args.InfrastructureOutputs,
		PlatformOutputs:      args.PlatformOutputs,
		ServiceType:          "gateway",
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("failed to deploy gateway containers: %w", err)
	}

	// Extract service endpoints from deployment
	var publicGatewayURL, adminGatewayURL pulumi.StringOutput

	switch args.Environment {
	case "development":
		publicGatewayURL = pulumi.String("http://127.0.0.1:9001").ToStringOutput()
		adminGatewayURL = pulumi.String("http://127.0.0.1:9000").ToStringOutput()
	case "staging":
		publicGatewayURL = pulumi.String("https://public-gateway-staging.azurecontainerapp.io").ToStringOutput()
		adminGatewayURL = pulumi.String("https://admin-gateway-staging.azurecontainerapp.io").ToStringOutput()
	case "production":
		publicGatewayURL = pulumi.String("https://public-gateway-production.azurecontainerapp.io").ToStringOutput()
		adminGatewayURL = pulumi.String("https://admin-gateway-production.azurecontainerapp.io").ToStringOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.Services = gatewayDeployment.Containers
	component.PublicGatewayURL = publicGatewayURL
	component.AdminGatewayURL = adminGatewayURL
	component.HealthEndpoints = gatewayDeployment.HealthEndpoints
	component.DaprSidecars = gatewayDeployment.DaprSidecars
	component.ResourceLimits = gatewayDeployment.ResourceLimits

	if canRegister(ctx) {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"services":         component.Services,
			"publicGatewayURL": component.PublicGatewayURL,
			"adminGatewayURL":  component.AdminGatewayURL,
			"healthEndpoints":  component.HealthEndpoints,
			"daprSidecars":     component.DaprSidecars,
			"resourceLimits":   component.ResourceLimits,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}