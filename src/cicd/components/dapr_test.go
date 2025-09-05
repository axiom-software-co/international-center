package components

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
)

// TestDaprComponent_DevelopmentEnvironment tests dapr component for development environment
func TestDaprComponent_DevelopmentEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployDapr(ctx, cfg, "development")
		if err != nil {
			return err
		}

		// Verify development environment generates self-hosted Dapr configuration
		pulumi.All(outputs.DeploymentType, outputs.RuntimePort, outputs.ControlPlaneURL, outputs.SidecarConfig).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			runtimePort := args[1].(int)
			controlPlaneURL := args[2].(string)
			sidecarConfig := args[3].(string)

			assert.Equal(t, "podman_dapr", deploymentType, "Development should use self-hosted Dapr")
			assert.Equal(t, 3500, runtimePort, "Should use standard Dapr runtime port")
			assert.Contains(t, controlPlaneURL, "http://127.0.0.1:50005", "Should use local control plane")
			assert.Equal(t, "development", sidecarConfig, "Should use development sidecar config")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &DaprMocks{}))

	assert.NoError(t, err)
}

// TestDaprComponent_StagingEnvironment tests dapr component for staging environment
func TestDaprComponent_StagingEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployDapr(ctx, cfg, "staging")
		if err != nil {
			return err
		}

		// Verify staging environment generates Container Apps managed Dapr
		pulumi.All(outputs.DeploymentType, outputs.ControlPlaneURL, outputs.SidecarConfig, outputs.MiddlewareEnabled).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			controlPlaneURL := args[1].(string)
			sidecarConfig := args[2].(string)
			middlewareEnabled := args[3].(bool)

			assert.Equal(t, "container_apps", deploymentType, "Staging should use Container Apps managed Dapr")
			assert.Contains(t, controlPlaneURL, "containerapp", "Should use Container Apps control plane")
			assert.Equal(t, "staging", sidecarConfig, "Should use staging sidecar config")
			assert.True(t, middlewareEnabled, "Should enable middleware for staging")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &DaprMocks{}))

	assert.NoError(t, err)
}

// TestDaprComponent_ProductionEnvironment tests dapr component for production environment
func TestDaprComponent_ProductionEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployDapr(ctx, cfg, "production")
		if err != nil {
			return err
		}

		// Verify production environment generates Container Apps with production features
		pulumi.All(outputs.DeploymentType, outputs.ControlPlaneURL, outputs.SidecarConfig, outputs.MiddlewareEnabled, outputs.PolicyEnabled).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			controlPlaneURL := args[1].(string)
			sidecarConfig := args[2].(string)
			middlewareEnabled := args[3].(bool)
			policyEnabled := args[4].(bool)

			assert.Equal(t, "container_apps", deploymentType, "Production should use Container Apps managed Dapr")
			assert.Contains(t, controlPlaneURL, "containerapp", "Should use Container Apps control plane")
			assert.Equal(t, "production", sidecarConfig, "Should use production sidecar config")
			assert.True(t, middlewareEnabled, "Should enable middleware for production")
			assert.True(t, policyEnabled, "Should enable OPA policies for production")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &DaprMocks{}))

	assert.NoError(t, err)
}

// TestDaprComponent_EnvironmentParity tests that all environments support required features
func TestDaprComponent_EnvironmentParity(t *testing.T) {
	environments := []string{"development", "staging", "production"}

	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				cfg := config.New(ctx, "")

				outputs, err := DeployDapr(ctx, cfg, env)
				if err != nil {
					return err
				}

				// Verify all environments provide required outputs
				pulumi.All(outputs.DeploymentType, outputs.ControlPlaneURL, outputs.SidecarConfig).ApplyT(func(args []interface{}) error {
					deploymentType := args[0].(string)
					controlPlaneURL := args[1].(string)
					sidecarConfig := args[2].(string)

					assert.NotEmpty(t, deploymentType, "All environments should provide deployment type")
					assert.NotEmpty(t, controlPlaneURL, "All environments should provide control plane URL")
					assert.NotEmpty(t, sidecarConfig, "All environments should provide sidecar config")
					return nil
				})

				return nil
			}, pulumi.WithMocks("test", "stack", &DaprMocks{}))

			assert.NoError(t, err)
		})
	}
}

// DaprMocks provides mocks for Pulumi testing
type DaprMocks struct{}

func (mocks *DaprMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}

	switch args.TypeToken {
	case "docker:index/container:Container":
		if args.Name == "dapr-sidecar" {
			outputs["name"] = resource.NewStringProperty("dapr-sidecar-dev")
			outputs["image"] = resource.NewStringProperty("daprio/daprd:latest")
			outputs["ports"] = resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewObjectProperty(resource.PropertyMap{
					"internal": resource.NewNumberProperty(3500),
					"external": resource.NewNumberProperty(3500),
				}),
			})
		} else if args.Name == "dapr-placement" {
			outputs["name"] = resource.NewStringProperty("dapr-placement-dev")
			outputs["image"] = resource.NewStringProperty("daprio/dapr:latest")
			outputs["ports"] = resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewObjectProperty(resource.PropertyMap{
					"internal": resource.NewNumberProperty(50005),
					"external": resource.NewNumberProperty(50005),
				}),
			})
		}

	case "azure-native:app:ContainerApp":
		outputs["configuration"] = resource.NewObjectProperty(resource.PropertyMap{
			"dapr": resource.NewObjectProperty(resource.PropertyMap{
				"enabled": resource.NewBoolProperty(true),
				"appId":   resource.NewStringProperty("international-center-app"),
			}),
		})
		outputs["name"] = resource.NewStringProperty("international-center-dapr")
	}

	return args.Name + "_id", outputs, nil
}

func (mocks *DaprMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}
	return outputs, nil
}