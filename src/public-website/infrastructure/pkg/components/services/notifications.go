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

	Services        pulumi.MapOutput    `pulumi:"services"`
	HealthEndpoints pulumi.MapOutput    `pulumi:"healthEndpoints"`
	DaprAppId       pulumi.StringOutput `pulumi:"daprAppId"`
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

	var services, healthEndpoints pulumi.MapOutput
	var daprAppId pulumi.StringOutput

	switch args.Environment {
	case "development":
		services = pulumi.Map{
			"email": pulumi.Map{
				"image":         pulumi.String("localhost/backend/notifications:latest"),
				"container_id":  pulumi.String("notifications-email"),
				"port":          pulumi.Int(3007),
				"replicas":      pulumi.Int(1),
				"health_check":  pulumi.String("/health"),
				"dapr_app_id":   pulumi.String("notifications-email"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("300m"),
					"memory": pulumi.String("128Mi"),
				},
			},
			"newsletter": pulumi.Map{
				"image":         pulumi.String("localhost/backend/notifications:latest"),
				"container_id":  pulumi.String("notifications-newsletter"),
				"port":          pulumi.Int(3008),
				"replicas":      pulumi.Int(1),
				"health_check":  pulumi.String("/health"),
				"dapr_app_id":   pulumi.String("notifications-newsletter"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("300m"),
					"memory": pulumi.String("128Mi"),
				},
			},
		}.ToMapOutput()
		healthEndpoints = pulumi.Map{
			"email":      pulumi.String("http://localhost:3007/health"),
			"newsletter": pulumi.String("http://localhost:3008/health"),
		}.ToMapOutput()
		daprAppId = pulumi.String("notifications").ToStringOutput()
	case "staging":
		services = pulumi.Map{
			"notifications": pulumi.Map{
				"image":        pulumi.String("registry.azurecr.io/backend/notifications:staging"),
				"replicas":     pulumi.Int(2),
				"health_check": pulumi.String("/health"),
				"dapr_app_id":  pulumi.String("notifications"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
			},
		}.ToMapOutput()
		healthEndpoints = pulumi.Map{
			"notifications": pulumi.String("https://notifications-staging.azurecontainerapp.io/health"),
		}.ToMapOutput()
		daprAppId = pulumi.String("notifications").ToStringOutput()
	case "production":
		services = pulumi.Map{
			"notifications": pulumi.Map{
				"image":        pulumi.String("registry.azurecr.io/backend/notifications:production"),
				"replicas":     pulumi.Int(3),
				"health_check": pulumi.String("/health"),
				"dapr_app_id":  pulumi.String("notifications"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("1000m"),
					"memory": pulumi.String("512Mi"),
				},
			},
		}.ToMapOutput()
		healthEndpoints = pulumi.Map{
			"notifications": pulumi.String("https://notifications-production.azurecontainerapp.io/health"),
		}.ToMapOutput()
		daprAppId = pulumi.String("notifications").ToStringOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.Services = services
	component.HealthEndpoints = healthEndpoints
	component.DaprAppId = daprAppId

	if canRegister(ctx) {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"services":        component.Services,
			"healthEndpoints": component.HealthEndpoints,
			"daprAppId":       component.DaprAppId,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}