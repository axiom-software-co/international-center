package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ObservabilityArgs struct {
	Environment string
}

type ObservabilityComponent struct {
	pulumi.ResourceState

	GrafanaURL       pulumi.StringOutput `pulumi:"grafanaURL"`
	PrometheusURL    pulumi.StringOutput `pulumi:"prometheusURL"`
	JaegerURL        pulumi.StringOutput `pulumi:"jaegerURL"`
	HealthEndpoint   pulumi.StringOutput `pulumi:"healthEndpoint"`
}

func NewObservabilityComponent(ctx *pulumi.Context, name string, args *ObservabilityArgs, opts ...pulumi.ResourceOption) (*ObservabilityComponent, error) {
	component := &ObservabilityComponent{}
	
	// Safe registration for mock contexts
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:infrastructure:Observability", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	var grafanaURL, prometheusURL, jaegerURL, healthEndpoint pulumi.StringOutput

	switch args.Environment {
	case "development":
		grafanaURL = pulumi.String("http://localhost:3000").ToStringOutput()
		prometheusURL = pulumi.String("http://localhost:9090").ToStringOutput()
		jaegerURL = pulumi.String("http://localhost:16686").ToStringOutput()
		healthEndpoint = pulumi.String("http://localhost:3000/api/health").ToStringOutput()
	case "staging":
		grafanaURL = pulumi.String("https://grafana-staging.azurewebsites.net").ToStringOutput()
		prometheusURL = pulumi.String("https://prometheus-staging.azurewebsites.net").ToStringOutput()
		jaegerURL = pulumi.String("https://jaeger-staging.azurewebsites.net").ToStringOutput()
		healthEndpoint = pulumi.String("https://grafana-staging.azurewebsites.net/api/health").ToStringOutput()
	case "production":
		grafanaURL = pulumi.String("https://grafana-production.azurewebsites.net").ToStringOutput()
		prometheusURL = pulumi.String("https://prometheus-production.azurewebsites.net").ToStringOutput()
		jaegerURL = pulumi.String("https://jaeger-production.azurewebsites.net").ToStringOutput()
		healthEndpoint = pulumi.String("https://grafana-production.azurewebsites.net/api/health").ToStringOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.GrafanaURL = grafanaURL
	component.PrometheusURL = prometheusURL
	component.JaegerURL = jaegerURL
	component.HealthEndpoint = healthEndpoint

	if canRegister(ctx) {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"grafanaURL":       component.GrafanaURL,
			"prometheusURL":    component.PrometheusURL,
			"jaegerURL":        component.JaegerURL,
			"healthEndpoint":   component.HealthEndpoint,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}

