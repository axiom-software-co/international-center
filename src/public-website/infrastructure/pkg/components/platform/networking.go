package platform

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type NetworkingArgs struct {
	Environment string
}

type NetworkingComponent struct {
	pulumi.ResourceState

	Configuration   pulumi.MapOutput    `pulumi:"configuration"`
	LoadBalancer    pulumi.StringOutput `pulumi:"loadBalancer"`
	ServiceMesh     pulumi.BoolOutput   `pulumi:"serviceMesh"`
	NetworkPolicies pulumi.MapOutput    `pulumi:"networkPolicies"`
}

func NewNetworkingComponent(ctx *pulumi.Context, name string, args *NetworkingArgs, opts ...pulumi.ResourceOption) (*NetworkingComponent, error) {
	component := &NetworkingComponent{}
	
	err := ctx.RegisterComponentResource("international-center:platform:Networking", name, component, opts...)
	if err != nil {
		return nil, err
	}

	var configuration, networkPolicies pulumi.MapOutput
	var loadBalancer pulumi.StringOutput
	var serviceMesh pulumi.BoolOutput

	switch args.Environment {
	case "development":
		configuration = pulumi.Map{
			"network_mode":    pulumi.String("bridge"),
			"dns_config":      pulumi.String("localhost"),
			"port_mapping":    pulumi.Bool(true),
			"host_networking": pulumi.Bool(false),
		}.ToMapOutput()
		loadBalancer = pulumi.String("nginx").ToStringOutput()
		serviceMesh = pulumi.Bool(true).ToBoolOutput()
		networkPolicies = pulumi.Map{
			"isolation_enabled": pulumi.Bool(false),
			"ingress_rules":     pulumi.StringArray{pulumi.String("allow_all")},
			"egress_rules":      pulumi.StringArray{pulumi.String("allow_all")},
		}.ToMapOutput()
	case "staging":
		configuration = pulumi.Map{
			"network_mode":    pulumi.String("container_app"),
			"dns_config":      pulumi.String("azure_dns"),
			"port_mapping":    pulumi.Bool(true),
			"host_networking": pulumi.Bool(false),
		}.ToMapOutput()
		loadBalancer = pulumi.String("azure_application_gateway").ToStringOutput()
		serviceMesh = pulumi.Bool(true).ToBoolOutput()
		networkPolicies = pulumi.Map{
			"isolation_enabled": pulumi.Bool(true),
			"ingress_rules":     pulumi.StringArray{pulumi.String("staging_only"), pulumi.String("internal_services")},
			"egress_rules":      pulumi.StringArray{pulumi.String("external_apis"), pulumi.String("azure_services")},
		}.ToMapOutput()
	case "production":
		configuration = pulumi.Map{
			"network_mode":    pulumi.String("container_app"),
			"dns_config":      pulumi.String("azure_dns"),
			"port_mapping":    pulumi.Bool(true),
			"host_networking": pulumi.Bool(false),
		}.ToMapOutput()
		loadBalancer = pulumi.String("azure_application_gateway").ToStringOutput()
		serviceMesh = pulumi.Bool(true).ToBoolOutput()
		networkPolicies = pulumi.Map{
			"isolation_enabled": pulumi.Bool(true),
			"ingress_rules":     pulumi.StringArray{pulumi.String("production_only"), pulumi.String("verified_sources")},
			"egress_rules":      pulumi.StringArray{pulumi.String("verified_apis"), pulumi.String("azure_services")},
		}.ToMapOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.Configuration = configuration
	component.LoadBalancer = loadBalancer
	component.ServiceMesh = serviceMesh
	component.NetworkPolicies = networkPolicies

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"configuration":   component.Configuration,
		"loadBalancer":    component.LoadBalancer,
		"serviceMesh":     component.ServiceMesh,
		"networkPolicies": component.NetworkPolicies,
	}); err != nil {
		return nil, err
	}

	return component, nil
}