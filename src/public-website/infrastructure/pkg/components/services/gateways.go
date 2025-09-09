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
}

func NewGatewayComponent(ctx *pulumi.Context, name string, args *GatewayArgs, opts ...pulumi.ResourceOption) (*GatewayComponent, error) {
	component := &GatewayComponent{}
	
	err := ctx.RegisterComponentResource("international-center:services:Gateway", name, component, opts...)
	if err != nil {
		return nil, err
	}

	var services, healthEndpoints pulumi.MapOutput
	var publicGatewayURL, adminGatewayURL pulumi.StringOutput

	switch args.Environment {
	case "development":
		services = pulumi.Map{
			"public": pulumi.Map{
				"image":         pulumi.String("localhost/backend/public-gateway:latest"),
				"container_id":  pulumi.String("public-gateway"),
				"port":          pulumi.Int(9001),
				"replicas":      pulumi.Int(1),
				"health_check":  pulumi.String("/health"),
				"dapr_app_id":   pulumi.String("public-gateway"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
			},
			"admin": pulumi.Map{
				"image":         pulumi.String("localhost/backend/admin-gateway:latest"),
				"container_id":  pulumi.String("admin-gateway"),
				"port":          pulumi.Int(9000),
				"replicas":      pulumi.Int(1),
				"health_check":  pulumi.String("/health"),
				"dapr_app_id":   pulumi.String("admin-gateway"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
			},
		}.ToMapOutput()
		publicGatewayURL = pulumi.String("http://127.0.0.1:9001").ToStringOutput()
		adminGatewayURL = pulumi.String("http://127.0.0.1:9000").ToStringOutput()
		healthEndpoints = pulumi.Map{
			"public": pulumi.String("http://localhost:9001/health"),
			"admin":  pulumi.String("http://localhost:9000/health"),
		}.ToMapOutput()
	case "staging":
		services = pulumi.Map{
			"public": pulumi.Map{
				"image":        pulumi.String("registry.azurecr.io/backend/public-gateway:staging"),
				"replicas":     pulumi.Int(3),
				"health_check": pulumi.String("/health"),
				"dapr_app_id":  pulumi.String("public-gateway"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("1000m"),
					"memory": pulumi.String("512Mi"),
				},
			},
			"admin": pulumi.Map{
				"image":        pulumi.String("registry.azurecr.io/backend/admin-gateway:staging"),
				"replicas":     pulumi.Int(2),
				"health_check": pulumi.String("/health"),
				"dapr_app_id":  pulumi.String("admin-gateway"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("1000m"),
					"memory": pulumi.String("512Mi"),
				},
			},
		}.ToMapOutput()
		publicGatewayURL = pulumi.String("https://public-gateway-staging.azurecontainerapp.io").ToStringOutput()
		adminGatewayURL = pulumi.String("https://admin-gateway-staging.azurecontainerapp.io").ToStringOutput()
		healthEndpoints = pulumi.Map{
			"public": pulumi.String("https://public-gateway-staging.azurecontainerapp.io/health"),
			"admin":  pulumi.String("https://admin-gateway-staging.azurecontainerapp.io/health"),
		}.ToMapOutput()
	case "production":
		services = pulumi.Map{
			"public": pulumi.Map{
				"image":        pulumi.String("registry.azurecr.io/backend/public-gateway:production"),
				"replicas":     pulumi.Int(5),
				"health_check": pulumi.String("/health"),
				"dapr_app_id":  pulumi.String("public-gateway"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("2000m"),
					"memory": pulumi.String("1Gi"),
				},
			},
			"admin": pulumi.Map{
				"image":        pulumi.String("registry.azurecr.io/backend/admin-gateway:production"),
				"replicas":     pulumi.Int(3),
				"health_check": pulumi.String("/health"),
				"dapr_app_id":  pulumi.String("admin-gateway"),
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("2000m"),
					"memory": pulumi.String("1Gi"),
				},
			},
		}.ToMapOutput()
		publicGatewayURL = pulumi.String("https://public-gateway-production.azurecontainerapp.io").ToStringOutput()
		adminGatewayURL = pulumi.String("https://admin-gateway-production.azurecontainerapp.io").ToStringOutput()
		healthEndpoints = pulumi.Map{
			"public": pulumi.String("https://public-gateway-production.azurecontainerapp.io/health"),
			"admin":  pulumi.String("https://admin-gateway-production.azurecontainerapp.io/health"),
		}.ToMapOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.Services = services
	component.PublicGatewayURL = publicGatewayURL
	component.AdminGatewayURL = adminGatewayURL
	component.HealthEndpoints = healthEndpoints

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"services":         component.Services,
		"publicGatewayURL": component.PublicGatewayURL,
		"adminGatewayURL":  component.AdminGatewayURL,
		"healthEndpoints":  component.HealthEndpoints,
	}); err != nil {
		return nil, err
	}

	return component, nil
}