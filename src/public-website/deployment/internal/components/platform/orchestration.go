package platform

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type OrchestrationArgs struct {
	Environment string
}

type OrchestrationComponent struct {
	pulumi.ResourceState

	OrchestratorType     pulumi.StringOutput `pulumi:"orchestratorType"`
	DeploymentStrategy   pulumi.StringOutput `pulumi:"deploymentStrategy"`
	ScalingPolicy        pulumi.StringOutput `pulumi:"scalingPolicy"`
	ResourceLimits       pulumi.MapOutput    `pulumi:"resourceLimits"`
	HealthCheckConfig    pulumi.MapOutput    `pulumi:"healthCheckConfig"`
}

func NewOrchestrationComponent(ctx *pulumi.Context, name string, args *OrchestrationArgs, opts ...pulumi.ResourceOption) (*OrchestrationComponent, error) {
	component := &OrchestrationComponent{}
	
	// Safe registration for mock contexts
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:platform:Orchestration", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	var orchestratorType, deploymentStrategy, scalingPolicy pulumi.StringOutput
	var resourceLimits, healthCheckConfig pulumi.MapOutput

	switch args.Environment {
	case "development":
		orchestratorType = pulumi.String("podman").ToStringOutput()
		deploymentStrategy = pulumi.String("rolling_update").ToStringOutput()
		scalingPolicy = pulumi.String("none").ToStringOutput()
		resourceLimits = pulumi.Map{
			"cpu":    pulumi.String("1000m"),
			"memory": pulumi.String("512Mi"),
		}.ToMapOutput()
		healthCheckConfig = pulumi.Map{
			"enabled":         pulumi.Bool(true),
			"interval":        pulumi.String("30s"),
			"timeout":         pulumi.String("10s"),
			"retries":         pulumi.Int(3),
			"start_period":    pulumi.String("40s"),
		}.ToMapOutput()
	case "staging":
		orchestratorType = pulumi.String("container_apps").ToStringOutput()
		deploymentStrategy = pulumi.String("blue_green").ToStringOutput()
		scalingPolicy = pulumi.String("moderate").ToStringOutput()
		resourceLimits = pulumi.Map{
			"cpu":    pulumi.String("2000m"),
			"memory": pulumi.String("1Gi"),
		}.ToMapOutput()
		healthCheckConfig = pulumi.Map{
			"enabled":         pulumi.Bool(true),
			"interval":        pulumi.String("15s"),
			"timeout":         pulumi.String("5s"),
			"retries":         pulumi.Int(5),
			"start_period":    pulumi.String("60s"),
		}.ToMapOutput()
	case "production":
		orchestratorType = pulumi.String("container_apps").ToStringOutput()
		deploymentStrategy = pulumi.String("canary").ToStringOutput()
		scalingPolicy = pulumi.String("aggressive").ToStringOutput()
		resourceLimits = pulumi.Map{
			"cpu":    pulumi.String("4000m"),
			"memory": pulumi.String("2Gi"),
		}.ToMapOutput()
		healthCheckConfig = pulumi.Map{
			"enabled":         pulumi.Bool(true),
			"interval":        pulumi.String("10s"),
			"timeout":         pulumi.String("3s"),
			"retries":         pulumi.Int(10),
			"start_period":    pulumi.String("90s"),
		}.ToMapOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.OrchestratorType = orchestratorType
	component.DeploymentStrategy = deploymentStrategy
	component.ScalingPolicy = scalingPolicy
	component.ResourceLimits = resourceLimits
	component.HealthCheckConfig = healthCheckConfig

	if canRegister(ctx) {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"orchestratorType":   component.OrchestratorType,
			"deploymentStrategy": component.DeploymentStrategy,
			"scalingPolicy":      component.ScalingPolicy,
			"resourceLimits":     component.ResourceLimits,
			"healthCheckConfig":  component.HealthCheckConfig,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}