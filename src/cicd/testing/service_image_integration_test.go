package testing

import (
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/cicd/components"
	"github.com/axiom-software-co/international-center/src/cicd/shared"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServiceImageIntegration validates consolidated service container deployments succeed with proper Dockerfiles
func TestServiceImageIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	
	t.Run("ValidateConsolidatedArchitectureDeployments", func(t *testing.T) {
		framework := shared.NewContractTestingFramework("international-center", "consolidated-architecture-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Validate that consolidated services can be deployed (Dockerfiles exist)
			// This validates our 5-container consolidated architecture:
			// inquiries, content, notifications, admin-gateway, public-gateway
			
			// Test inquiries service has proper Dockerfile
			_, err := components.DeployInquiriesServices(ctx)
			assert.NoError(t, err, "Inquiries consolidated service should deploy successfully")
			
			// Test content service has proper Dockerfile
			_, err = components.DeployContentServices(ctx)
			assert.NoError(t, err, "Content consolidated service should deploy successfully")
			
			// Test notifications service has proper Dockerfile
			_, err = components.DeployNotificationServices(ctx)
			assert.NoError(t, err, "Notifications consolidated service should deploy successfully")
			
			// Test gateway services have proper Dockerfiles
			_, err = components.DeployGatewayServices(ctx)
			assert.NoError(t, err, "Gateway services should deploy successfully")
			
			// Test website has proper Dockerfile
			_, err = components.DeployWebsiteContainer(ctx)
			assert.NoError(t, err, "Website should deploy successfully")
			
			return nil
		})
	})
}

// TestServiceImageBuildIntegration validates service containers deploy successfully with images
func TestServiceImageBuildIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	
	t.Run("ValidateInquiriesServicesSucceedWithImages", func(t *testing.T) {
		framework := shared.NewContractTestingFramework("international-center", "inquiries-build-integration-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Build required consolidated inquiries image
			builder := shared.NewImageBuilder(ctx, "development")
			
			// Build single consolidated inquiries service image (business, donations, media, volunteers)
			_, err := builder.BuildServiceImage("inquiries", "inquiries")
			require.NoError(t, err, "Should build consolidated inquiries service image")
			
			// Test that inquiries services deployment succeeds with consolidated image
			inquiriesMap, err := components.DeployInquiriesServices(ctx)
			require.NoError(t, err, "Inquiries services deployment should succeed with images")
			assert.NotNil(t, inquiriesMap, "Inquiries services should return deployment map")
			
			// Validate consolidated inquiries service is deployed
			inquiriesOutput := inquiriesMap["inquiries"]
			assert.NotNil(t, inquiriesOutput, "Consolidated inquiries service should be deployed")
			
			return nil
		})
	})
	
	t.Run("ValidateContentServicesSucceedWithImages", func(t *testing.T) {
		framework := shared.NewContractTestingFramework("international-center", "content-build-integration-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Build required consolidated content image
			builder := shared.NewImageBuilder(ctx, "development")
			
			// Build single consolidated content service image (events, news, research, services)
			_, err := builder.BuildServiceImage("content", "content")
			require.NoError(t, err, "Should build consolidated content service image")
			
			// Test that content services deployment succeeds with consolidated image
			contentMap, err := components.DeployContentServices(ctx)
			require.NoError(t, err, "Content services deployment should succeed with images")
			assert.NotNil(t, contentMap, "Content services should return deployment map")
			
			// Validate consolidated content service is deployed
			contentOutput := contentMap["content"]
			assert.NotNil(t, contentOutput, "Consolidated content service should be deployed")
			
			return nil
		})
	})
	
	t.Run("ValidateGatewayServicesSucceedWithImages", func(t *testing.T) {
		framework := shared.NewContractTestingFramework("international-center", "gateway-build-integration-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Build required images first
			builder := shared.NewImageBuilder(ctx, "development")
			
			gatewayServices := []string{"admin", "public"}
			for _, gateway := range gatewayServices {
				_, err := builder.BuildGatewayImage(gateway)
				require.NoError(t, err, "Should build %s gateway image", gateway)
			}
			
			// Test that gateway services deployment succeeds with images
			gatewayMap, err := components.DeployGatewayServices(ctx)
			require.NoError(t, err, "Gateway services deployment should succeed with images")
			assert.NotNil(t, gatewayMap, "Gateway services should return deployment map")
			
			// Validate all expected gateways are deployed
			for _, gateway := range gatewayServices {
				gatewayOutput := gatewayMap[gateway]
				assert.NotNil(t, gatewayOutput, "Gateway %s should be deployed", gateway)
			}
			
			return nil
		})
	})
	
	t.Run("ValidateNotificationServicesSucceedWithImages", func(t *testing.T) {
		framework := shared.NewContractTestingFramework("international-center", "notification-build-integration-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Build required consolidated notifications image
			builder := shared.NewImageBuilder(ctx, "development")
			
			// Build single consolidated notifications service image
			_, err := builder.BuildServiceImage("notifications", "notifications")
			require.NoError(t, err, "Should build consolidated notifications service image")
			
			// Test that notifications services deployment succeeds with consolidated image
			notificationsMap, err := components.DeployNotificationServices(ctx)
			require.NoError(t, err, "Notifications services deployment should succeed with images")
			assert.NotNil(t, notificationsMap, "Notifications services should return deployment map")
			
			// Validate consolidated notifications service is deployed
			notificationsOutput := notificationsMap["notifications"]
			assert.NotNil(t, notificationsOutput, "Consolidated notifications service should be deployed")
			
			return nil
		})
	})
	
	t.Run("ValidateWebsiteSucceedsWithImages", func(t *testing.T) {
		framework := shared.NewContractTestingFramework("international-center", "website-build-integration-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Build required image first
			builder := shared.NewImageBuilder(ctx, "development")
			_, err := builder.BuildWebsiteImage()
			require.NoError(t, err, "Should build website image")
			
			// Test that website deployment succeeds with image
			websiteCmd, err := components.DeployWebsiteContainer(ctx)
			require.NoError(t, err, "Website deployment should succeed with image")
			assert.NotNil(t, websiteCmd, "Website should return command resource")
			
			return nil
		})
	})
}

// TestIntegrationTimeouts validates all integration tests respect timeout constraints
func TestIntegrationTimeouts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout tests in short mode")
	}
	
	t.Run("ValidateImageBuildTimeouts", func(t *testing.T) {
		framework := shared.NewContractTestingFramework("international-center", "timeout-integration-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			builder := shared.NewImageBuilder(ctx, "development")
			
			// Test that image building completes within integration test timeout
			start := time.Now()
			_, err := builder.BuildServiceImage("inquiries", "inquiries")
			duration := time.Since(start)
			
			assert.NoError(t, err, "Image building should complete without timeout")
			assert.Less(t, duration, 15*time.Second, "Image build should complete within 15 seconds")
			
			return nil
		})
	})
	
	t.Run("ValidateContainerDeploymentTimeouts", func(t *testing.T) {
		framework := shared.NewContractTestingFramework("international-center", "deployment-timeout-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Build required image first
			builder := shared.NewImageBuilder(ctx, "development")
			_, err := builder.BuildServiceImage("inquiries", "inquiries")
			require.NoError(t, err, "Should build inquiries service image")
			
			// Test that container deployment completes within timeout
			start := time.Now()
			config := components.ContainerConfig{
				ServiceName:   "inquiries",
				ContainerName: "inquiries-timeout-test",
				ImageName:     "backend/inquiries:latest",
				HostPort:      8080,
				ContainerPort: 8080,
				DaprGrpcPort:  50001,
				AppID:         "inquiries",
			}
			_, err = components.DeployServiceContainer(ctx, config)
			duration := time.Since(start)
			
			assert.NoError(t, err, "Container deployment should complete without timeout")
			assert.Less(t, duration, 15*time.Second, "Container deployment should complete within 15 seconds")
			
			return nil
		})
	})
}