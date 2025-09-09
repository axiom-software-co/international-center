package admin

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type FrontendArgs struct {
	Environment     string
	ServicesOutputs pulumi.Map
}

type FrontendComponent struct {
	pulumi.ResourceState

	AdminPortalURL  pulumi.StringOutput `pulumi:"adminPortalURL"`
	ContainerConfig pulumi.MapOutput    `pulumi:"containerConfig"`
	StaticAssets    pulumi.MapOutput    `pulumi:"staticAssets"`
	HealthEndpoint  pulumi.StringOutput `pulumi:"healthEndpoint"`
}

func NewFrontendComponent(ctx *pulumi.Context, name string, args *FrontendArgs, opts ...pulumi.ResourceOption) (*FrontendComponent, error) {
	component := &FrontendComponent{}
	
	// Safe registration for mock contexts
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:admin:Frontend", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	var adminPortalURL, healthEndpoint pulumi.StringOutput
	var containerConfig, staticAssets pulumi.MapOutput

	switch args.Environment {
	case "development":
		adminPortalURL = pulumi.String("http://localhost:3001").ToStringOutput()
		healthEndpoint = pulumi.String("http://localhost:3001/health").ToStringOutput()
		containerConfig = pulumi.Map{
			"image":         pulumi.String("localhost/frontend/admin-portal:latest"),
			"container_id":  pulumi.String("admin-portal"),
			"port":          pulumi.Int(3001),
			"replicas":      pulumi.Int(1),
			"health_check":  pulumi.String("/health"),
			"resource_limits": pulumi.Map{
				"cpu":    pulumi.String("500m"),
				"memory": pulumi.String("256Mi"),
			},
			"environment_variables": pulumi.Map{
				"NODE_ENV":          pulumi.String("development"),
				"VITE_API_BASE_URL": pulumi.String("http://127.0.0.1:9002"),
				"VITE_APP_ENV":      pulumi.String("development"),
			},
		}.ToMapOutput()
		staticAssets = pulumi.Map{
			"build_command": pulumi.String("pnpm run build"),
			"dist_folder":   pulumi.String("dist"),
			"serve_static":  pulumi.Bool(true),
			"spa_mode":      pulumi.Bool(true),
		}.ToMapOutput()
	case "staging":
		adminPortalURL = pulumi.String("https://admin-portal-staging.azurecontainerapp.io").ToStringOutput()
		healthEndpoint = pulumi.String("https://admin-portal-staging.azurecontainerapp.io/health").ToStringOutput()
		containerConfig = pulumi.Map{
			"image":        pulumi.String("registry.azurecr.io/frontend/admin-portal:staging"),
			"replicas":     pulumi.Int(2),
			"health_check": pulumi.String("/health"),
			"resource_limits": pulumi.Map{
				"cpu":    pulumi.String("1000m"),
				"memory": pulumi.String("512Mi"),
			},
			"environment_variables": pulumi.Map{
				"NODE_ENV":          pulumi.String("staging"),
				"VITE_API_BASE_URL": pulumi.String("https://admin-gateway-staging.azurecontainerapp.io"),
				"VITE_APP_ENV":      pulumi.String("staging"),
			},
		}.ToMapOutput()
		staticAssets = pulumi.Map{
			"build_command": pulumi.String("pnpm run build:staging"),
			"dist_folder":   pulumi.String("dist"),
			"serve_static":  pulumi.Bool(true),
			"spa_mode":      pulumi.Bool(true),
			"cdn_enabled":   pulumi.Bool(true),
		}.ToMapOutput()
	case "production":
		adminPortalURL = pulumi.String("https://admin-portal-production.azurecontainerapp.io").ToStringOutput()
		healthEndpoint = pulumi.String("https://admin-portal-production.azurecontainerapp.io/health").ToStringOutput()
		containerConfig = pulumi.Map{
			"image":        pulumi.String("registry.azurecr.io/frontend/admin-portal:production"),
			"replicas":     pulumi.Int(3),
			"health_check": pulumi.String("/health"),
			"resource_limits": pulumi.Map{
				"cpu":    pulumi.String("2000m"),
				"memory": pulumi.String("1Gi"),
			},
			"environment_variables": pulumi.Map{
				"NODE_ENV":          pulumi.String("production"),
				"VITE_API_BASE_URL": pulumi.String("https://admin-gateway-production.azurecontainerapp.io"),
				"VITE_APP_ENV":      pulumi.String("production"),
			},
		}.ToMapOutput()
		staticAssets = pulumi.Map{
			"build_command": pulumi.String("pnpm run build:production"),
			"dist_folder":   pulumi.String("dist"),
			"serve_static":  pulumi.Bool(true),
			"spa_mode":      pulumi.Bool(true),
			"cdn_enabled":   pulumi.Bool(true),
			"minify":        pulumi.Bool(true),
			"tree_shaking":  pulumi.Bool(true),
		}.ToMapOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.AdminPortalURL = adminPortalURL
	component.ContainerConfig = containerConfig
	component.StaticAssets = staticAssets
	component.HealthEndpoint = healthEndpoint

	if canRegister(ctx) {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"adminPortalURL":  component.AdminPortalURL,
			"containerConfig": component.ContainerConfig,
			"staticAssets":    component.StaticAssets,
			"healthEndpoint":  component.HealthEndpoint,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}