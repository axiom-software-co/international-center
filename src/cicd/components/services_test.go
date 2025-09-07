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

		// Verify development environment deploys Podman containers
		pulumi.All(outputs.DeploymentType, outputs.InquiriesServices, outputs.ContentServices, outputs.NotificationServices, outputs.GatewayServices).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			inquiriesServices := args[1].(map[string]interface{})
			contentServices := args[2].(map[string]interface{})
			notificationServices := args[3].(map[string]interface{})
			gatewayServices := args[4].(map[string]interface{})

			assert.Equal(t, "podman_containers", deploymentType, "Development should use Podman containers")
			
			// Verify consolidated inquiries service container is deployed
			assert.Contains(t, inquiriesServices, "inquiries", "Should deploy consolidated inquiries service container")
			inquiriesConfig := inquiriesServices["inquiries"].(map[string]interface{})
			assert.Contains(t, inquiriesConfig, "container_id", "Inquiries should have container_id")
			assert.Contains(t, inquiriesConfig, "container_status", "Inquiries should have container_status")
			
			// Verify consolidated content service container is deployed
			assert.Contains(t, contentServices, "content", "Should deploy consolidated content service container")
			contentConfig := contentServices["content"].(map[string]interface{})
			assert.Contains(t, contentConfig, "container_id", "Content should have container_id")
			assert.Contains(t, contentConfig, "container_status", "Content should have container_status")
			
			// Verify consolidated notifications service container is deployed
			assert.Contains(t, notificationServices, "notifications", "Should deploy consolidated notifications service container")
			notificationsConfig := notificationServices["notifications"].(map[string]interface{})
			assert.Contains(t, notificationsConfig, "container_id", "Notifications should have container_id")
			assert.Contains(t, notificationsConfig, "container_status", "Notifications should have container_status")
			
			// Verify both gateway services containers are deployed
			expectedGateways := []string{"admin", "public"}
			for _, gateway := range expectedGateways {
				assert.Contains(t, gatewayServices, gateway, "Should deploy %s gateway service container", gateway)
				gatewayConfig := gatewayServices[gateway].(map[string]interface{})
				assert.Contains(t, gatewayConfig, "container_id", "%s gateway should have container_id", gateway)
				assert.Contains(t, gatewayConfig, "container_status", "%s gateway should have container_status", gateway)
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

// TestServicesComponent_ContainerConfiguration tests deployed container configurations
func TestServicesComponent_ContainerConfiguration(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployServices(ctx, cfg, "development")
		if err != nil {
			return err
		}

		// Verify consolidated services containers have required deployment attributes
		pulumi.All(outputs.InquiriesServices, outputs.ContentServices, outputs.NotificationServices).ApplyT(func(args []interface{}) error {
			inquiriesServices := args[0].(map[string]interface{})
			contentServices := args[1].(map[string]interface{})
			notificationServices := args[2].(map[string]interface{})

			// Verify consolidated inquiries service has required attributes
			inquiriesConfig, ok := inquiriesServices["inquiries"].(map[string]interface{})
			assert.True(t, ok, "Consolidated inquiries service should have container configuration")
			assert.Contains(t, inquiriesConfig, "container_id", "Inquiries service should have container_id")
			assert.Contains(t, inquiriesConfig, "container_status", "Inquiries service should have container_status")
			assert.Contains(t, inquiriesConfig, "host_port", "Inquiries service should have host_port")
			assert.Contains(t, inquiriesConfig, "health_endpoint", "Inquiries service should have health_endpoint")
			assert.Contains(t, inquiriesConfig, "dapr_app_id", "Inquiries service should have Dapr app ID")
			assert.Contains(t, inquiriesConfig, "dapr_sidecar_id", "Inquiries service should have Dapr sidecar container")
			
			// Verify consolidated content service has required attributes
			contentConfig, ok := contentServices["content"].(map[string]interface{})
			assert.True(t, ok, "Consolidated content service should have container configuration")
			assert.Contains(t, contentConfig, "container_id", "Content service should have container_id")
			assert.Contains(t, contentConfig, "container_status", "Content service should have container_status")
			assert.Contains(t, contentConfig, "host_port", "Content service should have host_port")
			assert.Contains(t, contentConfig, "health_endpoint", "Content service should have health_endpoint")
			assert.Contains(t, contentConfig, "dapr_app_id", "Content service should have Dapr app ID")
			assert.Contains(t, contentConfig, "dapr_sidecar_id", "Content service should have Dapr sidecar container")
			
			// Verify consolidated notifications service has required attributes
			notificationsConfig, ok := notificationServices["notifications"].(map[string]interface{})
			assert.True(t, ok, "Consolidated notifications service should have container configuration")
			assert.Contains(t, notificationsConfig, "container_id", "Notifications service should have container_id")
			assert.Contains(t, notificationsConfig, "container_status", "Notifications service should have container_status")
			assert.Contains(t, notificationsConfig, "host_port", "Notifications service should have host_port")
			assert.Contains(t, notificationsConfig, "health_endpoint", "Notifications service should have health_endpoint")
			assert.Contains(t, notificationsConfig, "dapr_app_id", "Notifications service should have Dapr app ID")
			assert.Contains(t, notificationsConfig, "dapr_sidecar_id", "Notifications service should have Dapr sidecar container")
			
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
				pulumi.All(outputs.DeploymentType, outputs.InquiriesServices, outputs.ContentServices, outputs.NotificationServices, outputs.GatewayServices, outputs.APIServices).ApplyT(func(args []interface{}) error {
					deploymentType := args[0].(string)
					inquiriesServices := args[1].(map[string]interface{})
					contentServices := args[2].(map[string]interface{})
					notificationServices := args[3].(map[string]interface{})
					gatewayServices := args[4].(map[string]interface{})
					apiServices := args[5].(map[string]interface{})

					assert.NotEmpty(t, deploymentType, "All environments should provide deployment type")
					assert.NotEmpty(t, gatewayServices, "All environments should provide gateway services")
					
					// Different architectures for different environments
					if env == "development" {
						assert.NotEmpty(t, inquiriesServices, "Development should provide inquiries services")
						assert.NotEmpty(t, contentServices, "Development should provide content services")
						assert.NotEmpty(t, notificationServices, "Development should provide notification services")
					} else {
						// Staging and production use APIServices instead of InquiriesServices/ContentServices/NotificationServices
						assert.NotEmpty(t, apiServices, "Staging and production should provide API services")
					}
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