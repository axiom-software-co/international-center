package website

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

	WebsiteURL      pulumi.StringOutput `pulumi:"websiteURL"`
	ContainerConfig pulumi.MapOutput    `pulumi:"containerConfig"`
	StaticAssets    pulumi.MapOutput    `pulumi:"staticAssets"`
	HealthEndpoint  pulumi.StringOutput `pulumi:"healthEndpoint"`
}

func NewFrontendComponent(ctx *pulumi.Context, name string, args *FrontendArgs, opts ...pulumi.ResourceOption) (*FrontendComponent, error) {
	component := &FrontendComponent{}
	
	// Safe registration for mock contexts
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:website:Frontend", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	var websiteURL, healthEndpoint pulumi.StringOutput
	var containerConfig, staticAssets pulumi.MapOutput

	switch args.Environment {
	case "development":
		websiteURL = pulumi.String("http://localhost:5173").ToStringOutput()
		healthEndpoint = pulumi.String("http://localhost:5173/health").ToStringOutput()
		containerConfig = pulumi.Map{
			"image":         pulumi.String("localhost/frontend/website:latest"),
			"container_id":  pulumi.String("website"),
			"port":          pulumi.Int(5173),
			"replicas":      pulumi.Int(1),
			"health_check":  pulumi.String("/health"),
			"resource_limits": pulumi.Map{
				"cpu":    pulumi.String("500m"),
				"memory": pulumi.String("256Mi"),
			},
			"environment_variables": pulumi.Map{
				"NODE_ENV":          pulumi.String("development"),
				"VITE_API_BASE_URL": pulumi.String("http://127.0.0.1:9001"),
				"VITE_APP_ENV":      pulumi.String("development"),
			},
		}.ToMapOutput()
		staticAssets = pulumi.Map{
			"build_command": pulumi.String("npm run build"),
			"dist_folder":   pulumi.String("dist"),
			"serve_static":  pulumi.Bool(true),
			"spa_mode":      pulumi.Bool(true),
		}.ToMapOutput()
	case "staging":
		websiteURL = pulumi.String("https://website-staging.azurecontainerapp.io").ToStringOutput()
		healthEndpoint = pulumi.String("https://website-staging.azurecontainerapp.io/health").ToStringOutput()
		containerConfig = pulumi.Map{
			"image":        pulumi.String("registry.azurecr.io/frontend/website:staging"),
			"replicas":     pulumi.Int(2),
			"health_check": pulumi.String("/health"),
			"resource_limits": pulumi.Map{
				"cpu":    pulumi.String("1000m"),
				"memory": pulumi.String("512Mi"),
			},
			"environment_variables": pulumi.Map{
				"NODE_ENV":          pulumi.String("staging"),
				"VITE_API_BASE_URL": pulumi.String("https://public-gateway-staging.azurecontainerapp.io"),
				"VITE_APP_ENV":      pulumi.String("staging"),
			},
		}.ToMapOutput()
		staticAssets = pulumi.Map{
			"build_command": pulumi.String("npm run build:staging"),
			"dist_folder":   pulumi.String("dist"),
			"serve_static":  pulumi.Bool(true),
			"spa_mode":      pulumi.Bool(true),
			"cdn_enabled":   pulumi.Bool(true),
		}.ToMapOutput()
	case "production":
		websiteURL = pulumi.String("https://website-production.azurecontainerapp.io").ToStringOutput()
		healthEndpoint = pulumi.String("https://website-production.azurecontainerapp.io/health").ToStringOutput()
		containerConfig = pulumi.Map{
			"image":        pulumi.String("registry.azurecr.io/frontend/website:production"),
			"replicas":     pulumi.Int(3),
			"health_check": pulumi.String("/health"),
			"resource_limits": pulumi.Map{
				"cpu":    pulumi.String("2000m"),
				"memory": pulumi.String("1Gi"),
			},
			"environment_variables": pulumi.Map{
				"NODE_ENV":          pulumi.String("production"),
				"VITE_API_BASE_URL": pulumi.String("https://public-gateway-production.azurecontainerapp.io"),
				"VITE_APP_ENV":      pulumi.String("production"),
			},
		}.ToMapOutput()
		staticAssets = pulumi.Map{
			"build_command": pulumi.String("npm run build:production"),
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

	component.WebsiteURL = websiteURL
	component.ContainerConfig = containerConfig
	component.StaticAssets = staticAssets
	component.HealthEndpoint = healthEndpoint

	if canRegister(ctx) {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"websiteURL":      component.WebsiteURL,
			"containerConfig": component.ContainerConfig,
			"staticAssets":    component.StaticAssets,
			"healthEndpoint":  component.HealthEndpoint,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}