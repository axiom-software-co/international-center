package components

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
)

// TestServicesComponent_DevelopmentEnvironment tests services component for development environment
func TestServicesComponent_DevelopmentEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployServices(ctx, cfg, "development")
		if err != nil {
			return err
		}

		// Verify development environment generates local container configurations
		pulumi.All(outputs.DeploymentType, outputs.APIServices, outputs.GatewayServices).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			apiServices := args[1].(map[string]interface{})
			gatewayServices := args[2].(map[string]interface{})

			assert.Equal(t, "containers", deploymentType, "Development should use local containers")
			
			// Verify all 8 API services are configured
			expectedAPIs := []string{"business", "donations", "events", "media", "news", "research", "services", "volunteers"}
			for _, api := range expectedAPIs {
				assert.Contains(t, apiServices, api, "Should configure %s API service", api)
			}
			
			// Verify both gateway services are configured
			expectedGateways := []string{"admin", "public"}
			for _, gateway := range expectedGateways {
				assert.Contains(t, gatewayServices, gateway, "Should configure %s gateway service", gateway)
			}

			return nil
		})

		// Verify development health check configuration
		pulumi.All(outputs.HealthCheckEnabled, outputs.DaprSidecarEnabled).ApplyT(func(args []interface{}) error {
			healthCheckEnabled := args[0].(bool)
			daprSidecarEnabled := args[1].(bool)

			assert.True(t, healthCheckEnabled, "Should enable health checks for development")
			assert.True(t, daprSidecarEnabled, "Should enable Dapr sidecars for development")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &ServicesMocks{}))

	assert.NoError(t, err)
}

// TestServicesComponent_StagingEnvironment tests services component for staging environment
func TestServicesComponent_StagingEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployServices(ctx, cfg, "staging")
		if err != nil {
			return err
		}

		// Verify staging environment generates Container Apps configuration
		pulumi.All(outputs.DeploymentType, outputs.APIServices, outputs.ScalingPolicy).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			apiServices := args[1].(map[string]interface{})
			scalingPolicy := args[2].(string)

			assert.Equal(t, "container_apps", deploymentType, "Staging should use Container Apps")
			assert.NotEmpty(t, apiServices, "Should configure API services for staging")
			assert.Equal(t, "moderate", scalingPolicy, "Should use moderate scaling policy")
			return nil
		})

		// Verify staging Dapr and observability integration
		pulumi.All(outputs.DaprSidecarEnabled, outputs.ObservabilityEnabled).ApplyT(func(args []interface{}) error {
			daprSidecarEnabled := args[0].(bool)
			observabilityEnabled := args[1].(bool)

			assert.True(t, daprSidecarEnabled, "Should enable Dapr sidecars for staging")
			assert.True(t, observabilityEnabled, "Should enable observability for staging")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &ServicesMocks{}))

	assert.NoError(t, err)
}

// TestServicesComponent_ProductionEnvironment tests services component for production environment
func TestServicesComponent_ProductionEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployServices(ctx, cfg, "production")
		if err != nil {
			return err
		}

		// Verify production environment generates Container Apps with production features
		pulumi.All(outputs.DeploymentType, outputs.ScalingPolicy, outputs.SecurityPolicies).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			scalingPolicy := args[1].(string)
			securityPolicies := args[2].(bool)

			assert.Equal(t, "container_apps", deploymentType, "Production should use Container Apps")
			assert.Equal(t, "aggressive", scalingPolicy, "Should use aggressive scaling policy")
			assert.True(t, securityPolicies, "Should enable security policies for production")
			return nil
		})

		// Verify production has full observability and audit logging
		pulumi.All(outputs.ObservabilityEnabled, outputs.AuditLogging).ApplyT(func(args []interface{}) error {
			observabilityEnabled := args[0].(bool)
			auditLogging := args[1].(bool)

			assert.True(t, observabilityEnabled, "Should enable observability for production")
			assert.True(t, auditLogging, "Should enable audit logging for production")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &ServicesMocks{}))

	assert.NoError(t, err)
}

// TestServicesComponent_ServiceConfiguration tests service-specific configurations
func TestServicesComponent_ServiceConfiguration(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployServices(ctx, cfg, "development")
		if err != nil {
			return err
		}

		// Verify service configuration includes required parameters
		pulumi.All(outputs.APIServices).ApplyT(func(args []interface{}) error {
			apiServices := args[0].(map[string]interface{})

			// Each API service should have required configuration
			for serviceName, serviceConfig := range apiServices {
				config, ok := serviceConfig.(map[string]interface{})
				assert.True(t, ok, "Service %s should have configuration map", serviceName)
				
				// Verify each service has required configuration
				assert.Contains(t, config, "image", "Service %s should have container image", serviceName)
				assert.Contains(t, config, "port", "Service %s should have port configuration", serviceName)
				assert.Contains(t, config, "health_check", "Service %s should have health check", serviceName)
				assert.Contains(t, config, "dapr_app_id", "Service %s should have Dapr app ID", serviceName)
			}
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &ServicesMocks{}))

	assert.NoError(t, err)
}

// TestServicesComponent_EnvironmentParity tests that all environments support required features
func TestServicesComponent_EnvironmentParity(t *testing.T) {
	environments := []string{"development", "staging", "production"}

	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				cfg := config.New(ctx, "")

				outputs, err := DeployServices(ctx, cfg, env)
				if err != nil {
					return err
				}

				// Verify all environments provide required outputs
				pulumi.All(outputs.DeploymentType, outputs.APIServices, outputs.GatewayServices).ApplyT(func(args []interface{}) error {
					deploymentType := args[0].(string)
					apiServices := args[1].(map[string]interface{})
					gatewayServices := args[2].(map[string]interface{})

					assert.NotEmpty(t, deploymentType, "All environments should provide deployment type")
					assert.NotEmpty(t, apiServices, "All environments should provide API services")
					assert.NotEmpty(t, gatewayServices, "All environments should provide gateway services")
					return nil
				})

				return nil
			}, pulumi.WithMocks("test", "stack", &ServicesMocks{}))

			assert.NoError(t, err)
		})
	}
}

// ServicesMocks provides mocks for Pulumi testing
type ServicesMocks struct{}

func (mocks *ServicesMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}

	switch args.TypeToken {
	case "docker:index/container:Container":
		// Mock docker container for development
		outputs["name"] = resource.NewStringProperty(args.Name + "-dev")
		outputs["image"] = resource.NewStringProperty("backend/" + args.Name + ":latest")
		outputs["ports"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{
				"internal": resource.NewNumberProperty(8080),
				"external": resource.NewNumberProperty(8080),
			}),
		})
		outputs["env"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty("DAPR_HTTP_PORT=3500"),
			resource.NewStringProperty("DAPR_GRPC_PORT=50001"),
		})

	case "azure-native:app:ContainerApp":
		// Mock Azure Container App for staging/production
		outputs["name"] = resource.NewStringProperty("international-center-" + args.Name)
		outputs["configuration"] = resource.NewObjectProperty(resource.PropertyMap{
			"dapr": resource.NewObjectProperty(resource.PropertyMap{
				"enabled": resource.NewBoolProperty(true),
				"appId":   resource.NewStringProperty(args.Name),
			}),
		})
		outputs["template"] = resource.NewObjectProperty(resource.PropertyMap{
			"containers": resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewObjectProperty(resource.PropertyMap{
					"image": resource.NewStringProperty("backend/" + args.Name + ":latest"),
					"name":  resource.NewStringProperty(args.Name),
					"resources": resource.NewObjectProperty(resource.PropertyMap{
						"cpu":    resource.NewNumberProperty(0.25),
						"memory": resource.NewStringProperty("0.5Gi"),
					}),
				}),
			}),
		})

	case "azure-native:operationalinsights:Workspace":
		// Mock Log Analytics Workspace for observability
		outputs["customerId"] = resource.NewStringProperty("workspace-customer-id")
		outputs["name"] = resource.NewStringProperty("international-center-logs")
	}

	return args.Name + "_id", outputs, nil
}

func (mocks *ServicesMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}
	return outputs, nil
}