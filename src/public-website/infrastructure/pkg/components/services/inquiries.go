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

	Services        pulumi.MapOutput    `pulumi:"services"`
	HealthEndpoints pulumi.MapOutput    `pulumi:"healthEndpoints"`
	DaprAppId       pulumi.StringOutput `pulumi:"daprAppId"`
}

func NewInquiriesComponent(ctx *pulumi.Context, name string, args *InquiriesArgs, opts ...pulumi.ResourceOption) (*InquiriesComponent, error) {
	component := &InquiriesComponent{}
	
	err := ctx.RegisterComponentResource("international-center:services:Inquiries", name, component, opts...)
	if err != nil {
		return nil, err
	}

	var services, healthEndpoints pulumi.MapOutput
	var daprAppId pulumi.StringOutput

	switch args.Environment {
	case "development":
		services = pulumi.Map{
			"contact": pulumi.Map{
				"image":         pulumi.String("localhost/backend/inquiries:latest"),
				"container_id":  pulumi.String("inquiries-contact"),
				"port":          pulumi.Int(3004),
				"replicas":      pulumi.Int(1),
				"health_check":  pulumi.String("/health"),
				"dapr_app_id":   pulumi.String("inquiries-contact"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
			},
			"volunteer": pulumi.Map{
				"image":         pulumi.String("localhost/backend/inquiries:latest"),
				"container_id":  pulumi.String("inquiries-volunteer"),
				"port":          pulumi.Int(3005),
				"replicas":      pulumi.Int(1),
				"health_check":  pulumi.String("/health"),
				"dapr_app_id":   pulumi.String("inquiries-volunteer"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
			},
			"services": pulumi.Map{
				"image":         pulumi.String("localhost/backend/inquiries:latest"),
				"container_id":  pulumi.String("inquiries-services"),
				"port":          pulumi.Int(3006),
				"replicas":      pulumi.Int(1),
				"health_check":  pulumi.String("/health"),
				"dapr_app_id":   pulumi.String("inquiries-services"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
			},
		}.ToMapOutput()
		healthEndpoints = pulumi.Map{
			"contact":   pulumi.String("http://localhost:3004/health"),
			"volunteer": pulumi.String("http://localhost:3005/health"),
			"services":  pulumi.String("http://localhost:3006/health"),
		}.ToMapOutput()
		daprAppId = pulumi.String("inquiries").ToStringOutput()
	case "staging":
		services = pulumi.Map{
			"inquiries": pulumi.Map{
				"image":        pulumi.String("registry.azurecr.io/backend/inquiries:staging"),
				"replicas":     pulumi.Int(3),
				"health_check": pulumi.String("/health"),
				"dapr_app_id":  pulumi.String("inquiries"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("1000m"),
					"memory": pulumi.String("512Mi"),
				},
			},
		}.ToMapOutput()
		healthEndpoints = pulumi.Map{
			"inquiries": pulumi.String("https://inquiries-staging.azurecontainerapp.io/health"),
		}.ToMapOutput()
		daprAppId = pulumi.String("inquiries").ToStringOutput()
	case "production":
		services = pulumi.Map{
			"inquiries": pulumi.Map{
				"image":        pulumi.String("registry.azurecr.io/backend/inquiries:production"),
				"replicas":     pulumi.Int(5),
				"health_check": pulumi.String("/health"),
				"dapr_app_id":  pulumi.String("inquiries"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("2000m"),
					"memory": pulumi.String("1Gi"),
				},
			},
		}.ToMapOutput()
		healthEndpoints = pulumi.Map{
			"inquiries": pulumi.String("https://inquiries-production.azurecontainerapp.io/health"),
		}.ToMapOutput()
		daprAppId = pulumi.String("inquiries").ToStringOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.Services = services
	component.HealthEndpoints = healthEndpoints
	component.DaprAppId = daprAppId

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"services":        component.Services,
		"healthEndpoints": component.HealthEndpoints,
		"daprAppId":       component.DaprAppId,
	}); err != nil {
		return nil, err
	}

	return component, nil
}