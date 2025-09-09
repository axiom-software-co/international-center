package services

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type ServicesArgs struct {
	Environment                string
	InfrastructureOutputs      pulumi.Map
	PlatformOutputs           pulumi.Map
}

type ServicesComponent struct {
	pulumi.ResourceState

	ContentServices      pulumi.MapOutput    `pulumi:"contentServices"`
	InquiriesServices    pulumi.MapOutput    `pulumi:"inquiriesServices"`
	NotificationServices pulumi.MapOutput    `pulumi:"notificationServices"`
	GatewayServices      pulumi.MapOutput    `pulumi:"gatewayServices"`
	PublicGatewayURL     pulumi.StringOutput `pulumi:"publicGatewayURL"`
	AdminGatewayURL      pulumi.StringOutput `pulumi:"adminGatewayURL"`
	DeploymentType       pulumi.StringOutput `pulumi:"deploymentType"`
	HealthCheckEnabled   pulumi.BoolOutput   `pulumi:"healthCheckEnabled"`
	DaprSidecarEnabled   pulumi.BoolOutput   `pulumi:"daprSidecarEnabled"`
	ScalingPolicy        pulumi.StringOutput `pulumi:"scalingPolicy"`
}

func NewServicesComponent(ctx *pulumi.Context, name string, args *ServicesArgs, opts ...pulumi.ResourceOption) (*ServicesComponent, error) {
	component := &ServicesComponent{}
	
	err := ctx.RegisterComponentResource("international-center:services:Services", name, component, opts...)
	if err != nil {
		return nil, err
	}

	// Deploy content services
	content, err := NewContentComponent(ctx, "content", &ContentArgs{
		Environment:           args.Environment,
		InfrastructureOutputs: args.InfrastructureOutputs,
		PlatformOutputs:      args.PlatformOutputs,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy inquiries services
	inquiries, err := NewInquiriesComponent(ctx, "inquiries", &InquiriesArgs{
		Environment:           args.Environment,
		InfrastructureOutputs: args.InfrastructureOutputs,
		PlatformOutputs:      args.PlatformOutputs,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy notification services
	notifications, err := NewNotificationComponent(ctx, "notifications", &NotificationArgs{
		Environment:           args.Environment,
		InfrastructureOutputs: args.InfrastructureOutputs,
		PlatformOutputs:      args.PlatformOutputs,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy gateway services
	gateways, err := NewGatewayComponent(ctx, "gateways", &GatewayArgs{
		Environment:           args.Environment,
		InfrastructureOutputs: args.InfrastructureOutputs,
		PlatformOutputs:      args.PlatformOutputs,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Configure environment-specific settings
	var deploymentType, scalingPolicy pulumi.StringOutput
	var healthCheckEnabled, daprSidecarEnabled pulumi.BoolOutput
	
	switch args.Environment {
	case "development":
		deploymentType = pulumi.String("podman_containers").ToStringOutput()
		scalingPolicy = pulumi.String("none").ToStringOutput()
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
		daprSidecarEnabled = pulumi.Bool(true).ToBoolOutput()
	case "staging":
		deploymentType = pulumi.String("container_apps").ToStringOutput()
		scalingPolicy = pulumi.String("moderate").ToStringOutput()
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
		daprSidecarEnabled = pulumi.Bool(true).ToBoolOutput()
	case "production":
		deploymentType = pulumi.String("container_apps").ToStringOutput()
		scalingPolicy = pulumi.String("aggressive").ToStringOutput()
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
		daprSidecarEnabled = pulumi.Bool(true).ToBoolOutput()
	default:
		deploymentType = pulumi.String("podman_containers").ToStringOutput()
		scalingPolicy = pulumi.String("none").ToStringOutput()
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
		daprSidecarEnabled = pulumi.Bool(true).ToBoolOutput()
	}

	// Set component outputs
	component.ContentServices = content.Services
	component.InquiriesServices = inquiries.Services
	component.NotificationServices = notifications.Services
	component.GatewayServices = gateways.Services
	component.PublicGatewayURL = gateways.PublicGatewayURL
	component.AdminGatewayURL = gateways.AdminGatewayURL
	component.DeploymentType = deploymentType
	component.HealthCheckEnabled = healthCheckEnabled
	component.DaprSidecarEnabled = daprSidecarEnabled
	component.ScalingPolicy = scalingPolicy

	// Register outputs
	ctx.Export("services:public_gateway_url", component.PublicGatewayURL)
	ctx.Export("services:admin_gateway_url", component.AdminGatewayURL)
	ctx.Export("services:deployment_type", component.DeploymentType)
	ctx.Export("services:scaling_policy", component.ScalingPolicy)

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"contentServices":      component.ContentServices,
		"inquiriesServices":    component.InquiriesServices,
		"notificationServices": component.NotificationServices,
		"gatewayServices":      component.GatewayServices,
		"publicGatewayURL":     component.PublicGatewayURL,
		"adminGatewayURL":      component.AdminGatewayURL,
		"deploymentType":       component.DeploymentType,
		"healthCheckEnabled":   component.HealthCheckEnabled,
		"daprSidecarEnabled":   component.DaprSidecarEnabled,
		"scalingPolicy":        component.ScalingPolicy,
	}); err != nil {
		return nil, err
	}

	return component, nil
}