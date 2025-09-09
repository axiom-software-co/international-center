package website

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type CDNArgs struct {
	Environment string
}

type CDNComponent struct {
	pulumi.ResourceState

	CDNEndpoint        pulumi.StringOutput `pulumi:"cdnEndpoint"`
	CacheConfiguration pulumi.MapOutput    `pulumi:"cacheConfiguration"`
	CompressionEnabled pulumi.BoolOutput   `pulumi:"compressionEnabled"`
	EdgeLocations      pulumi.ArrayOutput  `pulumi:"edgeLocations"`
}

func NewCDNComponent(ctx *pulumi.Context, name string, args *CDNArgs, opts ...pulumi.ResourceOption) (*CDNComponent, error) {
	component := &CDNComponent{}
	
	// Safe registration for mock contexts
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:website:CDN", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	var cdnEndpoint pulumi.StringOutput
	var cacheConfiguration pulumi.MapOutput
	var compressionEnabled pulumi.BoolOutput
	var edgeLocations pulumi.ArrayOutput

	switch args.Environment {
	case "development":
		cdnEndpoint = pulumi.String("").ToStringOutput() // No CDN in development
		cacheConfiguration = pulumi.Map{
			"enabled":     pulumi.Bool(false),
			"ttl":         pulumi.Int(0),
			"compression": pulumi.Bool(false),
		}.ToMapOutput()
		compressionEnabled = pulumi.Bool(false).ToBoolOutput()
		edgeLocations = pulumi.Array{}.ToArrayOutput()
	case "staging":
		cdnEndpoint = pulumi.String("https://international-center-staging.azureedge.net").ToStringOutput()
		cacheConfiguration = pulumi.Map{
			"enabled":              pulumi.Bool(true),
			"static_content_ttl":   pulumi.Int(3600),    // 1 hour
			"dynamic_content_ttl":  pulumi.Int(300),     // 5 minutes
			"compression":          pulumi.Bool(true),
			"compression_types":    pulumi.StringArray{
				pulumi.String("text/html"),
				pulumi.String("text/css"),
				pulumi.String("application/javascript"),
				pulumi.String("application/json"),
			},
			"cache_control_headers": pulumi.Bool(true),
		}.ToMapOutput()
		compressionEnabled = pulumi.Bool(true).ToBoolOutput()
		edgeLocations = pulumi.Array{
			pulumi.String("North America"),
			pulumi.String("Europe"),
		}.ToArrayOutput()
	case "production":
		cdnEndpoint = pulumi.String("https://international-center-production.azureedge.net").ToStringOutput()
		cacheConfiguration = pulumi.Map{
			"enabled":              pulumi.Bool(true),
			"static_content_ttl":   pulumi.Int(86400),   // 24 hours
			"dynamic_content_ttl":  pulumi.Int(900),     // 15 minutes
			"compression":          pulumi.Bool(true),
			"compression_types":    pulumi.StringArray{
				pulumi.String("text/html"),
				pulumi.String("text/css"),
				pulumi.String("application/javascript"),
				pulumi.String("application/json"),
				pulumi.String("image/svg+xml"),
			},
			"cache_control_headers": pulumi.Bool(true),
			"origin_shield":        pulumi.Bool(true),
			"waf_enabled":          pulumi.Bool(true),
		}.ToMapOutput()
		compressionEnabled = pulumi.Bool(true).ToBoolOutput()
		edgeLocations = pulumi.Array{
			pulumi.String("Global"),
		}.ToArrayOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.CDNEndpoint = cdnEndpoint
	component.CacheConfiguration = cacheConfiguration
	component.CompressionEnabled = compressionEnabled
	component.EdgeLocations = edgeLocations

	if canRegister(ctx) {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"cdnEndpoint":        component.CDNEndpoint,
			"cacheConfiguration": component.CacheConfiguration,
			"compressionEnabled": component.CompressionEnabled,
			"edgeLocations":      component.EdgeLocations,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}