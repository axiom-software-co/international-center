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

// TestServiceImageIntegration validates service container deployments require valid images
func TestServiceImageIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	
	t.Run("ValidateInquiriesServicesFailWithoutImages", func(t *testing.T) {
		framework := shared.NewContractTestingFramework("international-center", "inquiries-integration-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Test that inquiries services deployment fails when images don't exist
			_, err := components.DeployInquiriesServices(ctx)
			require.Error(t, err, "Inquiries services deployment should fail without images")
			assert.Contains(t, err.Error(), "image not found", "Error should indicate missing image")
			
			return nil
		})
	})
	
	t.Run("ValidateContentServicesFailWithoutImages", func(t *testing.T) {
		framework := shared.NewContractTestingFramework("international-center", "content-integration-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Test that content services deployment fails when images don't exist
			_, err := components.DeployContentServices(ctx)
			require.Error(t, err, "Content services deployment should fail without images")
			assert.Contains(t, err.Error(), "image not found", "Error should indicate missing image")
			
			return nil
		})
	})
	
	t.Run("ValidateGatewayServicesFailWithoutImages", func(t *testing.T) {
		framework := shared.NewContractTestingFramework("international-center", "gateway-integration-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Test that gateway services deployment fails when images don't exist
			_, err := components.DeployGatewayServices(ctx)
			require.Error(t, err, "Gateway services deployment should fail without images")
			assert.Contains(t, err.Error(), "image not found", "Error should indicate missing image")
			
			return nil
		})
	})
	
	t.Run("ValidateWebsiteFailsWithoutImages", func(t *testing.T) {
		framework := shared.NewContractTestingFramework("international-center", "website-integration-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Test that website deployment fails when image doesn't exist
			_, err := components.DeployWebsiteContainer(ctx)
			require.Error(t, err, "Website deployment should fail without image")
			assert.Contains(t, err.Error(), "image not found", "Error should indicate missing image")
			
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
			// Build required images first
			builder := shared.NewImageBuilder(ctx, "development")
			
			inquiriesServices := []string{"media", "donations", "volunteers", "business"}
			for _, service := range inquiriesServices {
				_, err := builder.BuildServiceImage(service, "inquiries")
				require.NoError(t, err, "Should build %s service image", service)
			}
			
			// Test that inquiries services deployment succeeds with images
			inquiriesMap, err := components.DeployInquiriesServices(ctx)
			require.NoError(t, err, "Inquiries services deployment should succeed with images")
			assert.NotNil(t, inquiriesMap, "Inquiries services should return deployment map")
			
			// Validate all expected services are deployed
			for _, service := range inquiriesServices {
				serviceOutput := inquiriesMap[service]
				assert.NotNil(t, serviceOutput, "Service %s should be deployed", service)
			}
			
			return nil
		})
	})
	
	t.Run("ValidateContentServicesSucceedWithImages", func(t *testing.T) {
		framework := shared.NewContractTestingFramework("international-center", "content-build-integration-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Build required images first
			builder := shared.NewImageBuilder(ctx, "development")
			
			contentServices := []string{"research", "services", "events", "news"}
			for _, service := range contentServices {
				_, err := builder.BuildServiceImage(service, "content")
				require.NoError(t, err, "Should build %s content service image", service)
			}
			
			// Test that content services deployment succeeds with images
			contentMap, err := components.DeployContentServices(ctx)
			require.NoError(t, err, "Content services deployment should succeed with images")
			assert.NotNil(t, contentMap, "Content services should return deployment map")
			
			// Validate all expected services are deployed
			for _, service := range contentServices {
				serviceOutput := contentMap[service]
				assert.NotNil(t, serviceOutput, "Service %s should be deployed", service)
			}
			
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
			_, err := builder.BuildServiceImage("media", "inquiries")
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
			_, err := builder.BuildServiceImage("media", "inquiries")
			require.NoError(t, err, "Should build media service image")
			
			// Test that container deployment completes within timeout
			start := time.Now()
			config := components.ContainerConfig{
				ServiceName:   "media",
				ContainerName: "media-timeout-test",
				ImageName:     "backend/media:latest",
				HostPort:      8080,
				ContainerPort: 8080,
				DaprGrpcPort:  50001,
				AppID:         "media-api",
			}
			_, err = components.DeployServiceContainer(ctx, config)
			duration := time.Since(start)
			
			assert.NoError(t, err, "Container deployment should complete without timeout")
			assert.Less(t, duration, 15*time.Second, "Container deployment should complete within 15 seconds")
			
			return nil
		})
	})
}