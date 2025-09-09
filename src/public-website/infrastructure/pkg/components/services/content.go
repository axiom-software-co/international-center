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
	DaprAppId       pulumi.StringOutput `pulumi:"daprAppId"`
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

	var services, healthEndpoints pulumi.MapOutput
	var daprAppId pulumi.StringOutput

	switch args.Environment {
	case "development":
		services = pulumi.Map{
			"news": pulumi.Map{
				"image":         pulumi.String("localhost/backend/content:latest"),
				"container_id":  pulumi.String("content-news"),
				"port":          pulumi.Int(3001),
				"replicas":      pulumi.Int(1),
				"health_check":  pulumi.String("/health"),
				"dapr_app_id":   pulumi.String("content-news"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
			},
			"events": pulumi.Map{
				"image":         pulumi.String("localhost/backend/content:latest"),
				"container_id":  pulumi.String("content-events"),
				"port":          pulumi.Int(3002),
				"replicas":      pulumi.Int(1),
				"health_check":  pulumi.String("/health"),
				"dapr_app_id":   pulumi.String("content-events"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
			},
			"research": pulumi.Map{
				"image":         pulumi.String("localhost/backend/content:latest"),
				"container_id":  pulumi.String("content-research"),
				"port":          pulumi.Int(3003),
				"replicas":      pulumi.Int(1),
				"health_check":  pulumi.String("/health"),
				"dapr_app_id":   pulumi.String("content-research"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
			},
		}.ToMapOutput()
		healthEndpoints = pulumi.Map{
			"news":     pulumi.String("http://localhost:3001/health"),
			"events":   pulumi.String("http://localhost:3002/health"),
			"research": pulumi.String("http://localhost:3003/health"),
		}.ToMapOutput()
		daprAppId = pulumi.String("content").ToStringOutput()
	case "staging":
		services = pulumi.Map{
			"content": pulumi.Map{
				"image":        pulumi.String("registry.azurecr.io/backend/content:staging"),
				"replicas":     pulumi.Int(3),
				"health_check": pulumi.String("/health"),
				"dapr_app_id":  pulumi.String("content"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("1000m"),
					"memory": pulumi.String("512Mi"),
				},
			},
		}.ToMapOutput()
		healthEndpoints = pulumi.Map{
			"content": pulumi.String("https://content-staging.azurecontainerapp.io/health"),
		}.ToMapOutput()
		daprAppId = pulumi.String("content").ToStringOutput()
	case "production":
		services = pulumi.Map{
			"content": pulumi.Map{
				"image":        pulumi.String("registry.azurecr.io/backend/content:production"),
				"replicas":     pulumi.Int(5),
				"health_check": pulumi.String("/health"),
				"dapr_app_id":  pulumi.String("content"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("2000m"),
					"memory": pulumi.String("1Gi"),
				},
			},
		}.ToMapOutput()
		healthEndpoints = pulumi.Map{
			"content": pulumi.String("https://content-production.azurecontainerapp.io/health"),
		}.ToMapOutput()
		daprAppId = pulumi.String("content").ToStringOutput()
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