package website

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type WebsiteArgs struct {
	Environment           string
	InfrastructureOutputs pulumi.Map
	PlatformOutputs      pulumi.Map
	ServicesOutputs      pulumi.Map
}

type WebsiteComponent struct {
	pulumi.ResourceState

	WebsiteURL           pulumi.StringOutput `pulumi:"websiteURL"`
	DeploymentType       pulumi.StringOutput `pulumi:"deploymentType"`
	CDNEnabled           pulumi.BoolOutput   `pulumi:"cdnEnabled"`
	SSLEnabled           pulumi.BoolOutput   `pulumi:"sslEnabled"`
	CacheConfiguration   pulumi.MapOutput    `pulumi:"cacheConfiguration"`
	HealthCheckEnabled   pulumi.BoolOutput   `pulumi:"healthCheckEnabled"`
	ContainerConfig      pulumi.MapOutput    `pulumi:"containerConfig"`
	StaticAssets         pulumi.MapOutput    `pulumi:"staticAssets"`
}

func NewWebsiteComponent(ctx *pulumi.Context, name string, args *WebsiteArgs, opts ...pulumi.ResourceOption) (*WebsiteComponent, error) {
	component := &WebsiteComponent{}
	
	err := ctx.RegisterComponentResource("international-center:website:Website", name, component, opts...)
	if err != nil {
		return nil, err
	}

	// Deploy frontend configuration
	frontend, err := NewFrontendComponent(ctx, "frontend", &FrontendArgs{
		Environment:      args.Environment,
		ServicesOutputs: args.ServicesOutputs,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy CDN configuration
	cdn, err := NewCDNComponent(ctx, "cdn", &CDNArgs{
		Environment: args.Environment,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy SSL configuration
	_, err = NewSSLComponent(ctx, "ssl", &SSLArgs{
		Environment: args.Environment,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Configure environment-specific settings
	var deploymentType pulumi.StringOutput
	var cdnEnabled, sslEnabled, healthCheckEnabled pulumi.BoolOutput
	
	switch args.Environment {
	case "development":
		deploymentType = pulumi.String("podman_container").ToStringOutput()
		cdnEnabled = pulumi.Bool(false).ToBoolOutput()
		sslEnabled = pulumi.Bool(false).ToBoolOutput()
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
	case "staging":
		deploymentType = pulumi.String("container_app").ToStringOutput()
		cdnEnabled = pulumi.Bool(true).ToBoolOutput()
		sslEnabled = pulumi.Bool(true).ToBoolOutput()
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
	case "production":
		deploymentType = pulumi.String("container_app").ToStringOutput()
		cdnEnabled = pulumi.Bool(true).ToBoolOutput()
		sslEnabled = pulumi.Bool(true).ToBoolOutput()
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
	default:
		deploymentType = pulumi.String("podman_container").ToStringOutput()
		cdnEnabled = pulumi.Bool(false).ToBoolOutput()
		sslEnabled = pulumi.Bool(false).ToBoolOutput()
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
	}

	// Set component outputs
	component.WebsiteURL = frontend.WebsiteURL
	component.DeploymentType = deploymentType
	component.CDNEnabled = cdnEnabled
	component.SSLEnabled = sslEnabled
	component.CacheConfiguration = cdn.CacheConfiguration
	component.HealthCheckEnabled = healthCheckEnabled
	component.ContainerConfig = frontend.ContainerConfig
	component.StaticAssets = frontend.StaticAssets

	// Register outputs
	ctx.Export("website:url", component.WebsiteURL)
	ctx.Export("website:deployment_type", component.DeploymentType)
	ctx.Export("website:cdn_enabled", component.CDNEnabled)
	ctx.Export("website:ssl_enabled", component.SSLEnabled)

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"websiteURL":           component.WebsiteURL,
		"deploymentType":       component.DeploymentType,
		"cdnEnabled":           component.CDNEnabled,
		"sslEnabled":           component.SSLEnabled,
		"cacheConfiguration":   component.CacheConfiguration,
		"healthCheckEnabled":   component.HealthCheckEnabled,
		"containerConfig":      component.ContainerConfig,
		"staticAssets":         component.StaticAssets,
	}); err != nil {
		return nil, err
	}

	return component, nil
}