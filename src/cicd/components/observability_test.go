package components

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
)

// TestObservabilityComponent_DevelopmentEnvironment tests observability component for development environment
func TestObservabilityComponent_DevelopmentEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployObservability(ctx, cfg, "development")
		if err != nil {
			return err
		}

		// Verify development environment generates consolidated otel-lgtm stack configuration
		pulumi.All(outputs.StackType, outputs.GrafanaURL, outputs.PrometheusURL, outputs.LokiURL).ApplyT(func(args []interface{}) error {
			stackType := args[0].(string)
			grafanaURL := args[1].(string)
			prometheusURL := args[2].(string)
			lokiURL := args[3].(string)

			assert.Equal(t, "otel_lgtm_container", stackType, "Development should use consolidated otel-lgtm container")
			assert.Contains(t, grafanaURL, "http://127.0.0.1:3000", "Should use local Grafana URL")
			assert.Contains(t, prometheusURL, "http://127.0.0.1:9090", "Should use standard Prometheus URL")
			assert.Contains(t, lokiURL, "http://127.0.0.1:3100", "Should use local Loki URL")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &ObservabilityMocks{}))

	assert.NoError(t, err)
}

// TestObservabilityComponent_StagingEnvironment tests observability component for staging environment
func TestObservabilityComponent_StagingEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployObservability(ctx, cfg, "staging")
		if err != nil {
			return err
		}

		// Verify staging environment generates Grafana Cloud configuration
		pulumi.All(outputs.StackType, outputs.GrafanaURL, outputs.RetentionDays, outputs.AuditLogging).ApplyT(func(args []interface{}) error {
			stackType := args[0].(string)
			grafanaURL := args[1].(string)
			retentionDays := args[2].(int)
			auditLogging := args[3].(bool)

			assert.Equal(t, "grafana_cloud", stackType, "Staging should use Grafana Cloud")
			assert.Contains(t, grafanaURL, "grafana.net", "Should use Grafana Cloud URL")
			assert.Equal(t, 30, retentionDays, "Should configure staging retention days")
			assert.True(t, auditLogging, "Should enable audit logging")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &ObservabilityMocks{}))

	assert.NoError(t, err)
}

// TestObservabilityComponent_ProductionEnvironment tests observability component for production environment
func TestObservabilityComponent_ProductionEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployObservability(ctx, cfg, "production")
		if err != nil {
			return err
		}

		// Verify production environment generates Grafana Cloud with production features
		pulumi.All(outputs.StackType, outputs.GrafanaURL, outputs.RetentionDays, outputs.AuditLogging, outputs.AlertingEnabled).ApplyT(func(args []interface{}) error {
			stackType := args[0].(string)
			grafanaURL := args[1].(string)
			retentionDays := args[2].(int)
			auditLogging := args[3].(bool)
			alertingEnabled := args[4].(bool)

			assert.Equal(t, "grafana_cloud", stackType, "Production should use Grafana Cloud")
			assert.Contains(t, grafanaURL, "grafana.net", "Should use Grafana Cloud URL")
			assert.Equal(t, 90, retentionDays, "Should configure production retention days")
			assert.True(t, auditLogging, "Should enable audit logging for production")
			assert.True(t, alertingEnabled, "Should enable alerting for production")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &ObservabilityMocks{}))

	assert.NoError(t, err)
}

// TestObservabilityComponent_EnvironmentParity tests that all environments support required features
func TestObservabilityComponent_EnvironmentParity(t *testing.T) {
	environments := []string{"development", "staging", "production"}

	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				cfg := config.New(ctx, "")

				outputs, err := DeployObservability(ctx, cfg, env)
				if err != nil {
					return err
				}

				// Verify all environments provide required outputs
				pulumi.All(outputs.GrafanaURL, outputs.StackType, outputs.AuditLogging).ApplyT(func(args []interface{}) error {
					grafanaURL := args[0].(string)
					stackType := args[1].(string)
					auditLogging := args[2].(bool)

					assert.NotEmpty(t, grafanaURL, "All environments should provide Grafana URL")
					assert.NotEmpty(t, stackType, "All environments should provide stack type")
					assert.NotNil(t, auditLogging, "All environments should specify audit logging")
					return nil
				})

				return nil
			}, pulumi.WithMocks("test", "stack", &ObservabilityMocks{}))

			assert.NoError(t, err)
		})
	}
}

// ObservabilityMocks provides mocks for Pulumi testing
type ObservabilityMocks struct{}

func (mocks *ObservabilityMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}

	switch args.TypeToken {
	case "docker:index/container:Container":
		if args.Name == "grafana" {
			outputs["name"] = resource.NewStringProperty("grafana-dev")
			outputs["image"] = resource.NewStringProperty("grafana/grafana:latest")
			outputs["ports"] = resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewObjectProperty(resource.PropertyMap{
					"internal": resource.NewNumberProperty(3000),
					"external": resource.NewNumberProperty(3000),
				}),
			})
		} else if args.Name == "prometheus" {
			outputs["name"] = resource.NewStringProperty("prometheus-dev")
			outputs["image"] = resource.NewStringProperty("prom/prometheus:latest")
			outputs["ports"] = resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewObjectProperty(resource.PropertyMap{
					"internal": resource.NewNumberProperty(9090),
					"external": resource.NewNumberProperty(9090),
				}),
			})
		} else if args.Name == "loki" {
			outputs["name"] = resource.NewStringProperty("loki-dev")
			outputs["image"] = resource.NewStringProperty("grafana/loki:latest")
			outputs["ports"] = resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewObjectProperty(resource.PropertyMap{
					"internal": resource.NewNumberProperty(3100),
					"external": resource.NewNumberProperty(3100),
				}),
			})
		}

	case "grafana:index/cloudStack:CloudStack":
		outputs["url"] = resource.NewStringProperty("https://international-center.grafana.net")
		outputs["status"] = resource.NewStringProperty("active")
		outputs["orgId"] = resource.NewNumberProperty(1077975)
	}

	return args.Name + "_id", outputs, nil
}

func (mocks *ObservabilityMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}
	return outputs, nil
}