package services

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type NotificationArgs struct {
	Environment           string
	InfrastructureOutputs pulumi.Map
	PlatformOutputs      pulumi.Map
}

type NotificationComponent struct {
	pulumi.ResourceState

	Services         pulumi.MapOutput    `pulumi:"services"`
	HealthEndpoints  pulumi.MapOutput    `pulumi:"healthEndpoints"`
	DaprSidecars     pulumi.MapOutput    `pulumi:"daprSidecars"`
	ResourceLimits   pulumi.MapOutput    `pulumi:"resourceLimits"`
	ServiceEndpoints pulumi.MapOutput    `pulumi:"serviceEndpoints"`
}

func NewNotificationComponent(ctx *pulumi.Context, name string, args *NotificationArgs, opts ...pulumi.ResourceOption) (*NotificationComponent, error) {
	component := &NotificationComponent{}
	
	// Safe registration for mock contexts
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:services:Notification", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	// Deploy actual notification containers using the container deployment component
	notificationDeployment, err := NewContainerDeploymentComponent(ctx, "notification-deployment", &ContainerDeploymentArgs{
		Environment:           args.Environment,
		InfrastructureOutputs: args.InfrastructureOutputs,
		PlatformOutputs:      args.PlatformOutputs,
		ServiceType:          "notifications",
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("failed to deploy notification containers: %w", err)
	}

	component.Services = notificationDeployment.Containers
	component.HealthEndpoints = notificationDeployment.HealthEndpoints
	component.DaprSidecars = notificationDeployment.DaprSidecars
	component.ResourceLimits = notificationDeployment.ResourceLimits
	component.ServiceEndpoints = notificationDeployment.ServiceEndpoints

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