package shared

import (
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/cicd/components"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestImageBuildingValidation validates Docker image building contracts
func TestImageBuildingValidation(t *testing.T) {
	t.Run("ValidateImageBuildingContract", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "image-build-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Test that image builder interface exists and validates correctly
			builder := NewImageBuilder(ctx, env)
			require.NotNil(t, builder, "ImageBuilder should be instantiated")
			
			return nil
		})
	})
	
	t.Run("ValidateInquiriesServiceImageBuilding", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "inquiries-image-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			builder := NewImageBuilder(ctx, env)
			
			// Test building inquiries service images
			inquiriesServices := []string{"media", "donations", "volunteers", "business"}
			for _, service := range inquiriesServices {
				imageRef, err := builder.BuildServiceImage(service, "inquiries")
				require.NoError(t, err, "Should build %s service image successfully", service)
				assert.NotEmpty(t, imageRef, "Image reference should not be empty for %s", service)
				
				// Validate image exists locally
				exists, err := builder.ImageExists(imageRef)
				require.NoError(t, err, "Should check image existence without error")
				assert.True(t, exists, "Image %s should exist after building", imageRef)
			}
			
			return nil
		})
	})
	
	t.Run("ValidateContentServiceImageBuilding", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "content-image-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			builder := NewImageBuilder(ctx, env)
			
			// Test building content service images  
			contentServices := []string{"research", "services", "events", "news"}
			for _, service := range contentServices {
				imageRef, err := builder.BuildServiceImage(service, "content")
				require.NoError(t, err, "Should build %s content service image successfully", service)
				assert.NotEmpty(t, imageRef, "Image reference should not be empty for %s", service)
				
				// Validate image exists locally
				exists, err := builder.ImageExists(imageRef)
				require.NoError(t, err, "Should check image existence without error")
				assert.True(t, exists, "Image %s should exist after building", imageRef)
			}
			
			return nil
		})
	})
	
	t.Run("ValidateGatewayServiceImageBuilding", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "gateway-image-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			builder := NewImageBuilder(ctx, env)
			
			// Test building gateway service images
			gatewayServices := []string{"admin", "public"}
			for _, service := range gatewayServices {
				imageRef, err := builder.BuildGatewayImage(service)
				require.NoError(t, err, "Should build %s gateway image successfully", service)
				assert.NotEmpty(t, imageRef, "Image reference should not be empty for %s", service)
				
				// Validate image exists locally
				exists, err := builder.ImageExists(imageRef)
				require.NoError(t, err, "Should check image existence without error")
				assert.True(t, exists, "Image %s should exist after building", imageRef)
			}
			
			return nil
		})
	})
	
	t.Run("ValidateWebsiteImageBuilding", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "website-image-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			builder := NewImageBuilder(ctx, env)
			
			// Test building website image
			imageRef, err := builder.BuildWebsiteImage()
			require.NoError(t, err, "Should build website image successfully")
			assert.NotEmpty(t, imageRef, "Website image reference should not be empty")
			
			// Validate image exists locally
			exists, err := builder.ImageExists(imageRef)
			require.NoError(t, err, "Should check website image existence without error")
			assert.True(t, exists, "Website image %s should exist after building", imageRef)
			
			return nil
		})
	})
}

// TestContainerFactoryImageValidation validates container factory validates images before deployment
func TestContainerFactoryImageValidation(t *testing.T) {
	t.Run("ValidateImageExistenceBeforeContainerDeployment", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "factory-image-validation")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			// Test that container factory validates image existence before deployment
			config := components.ContainerConfig{
				ServiceName:   "media", 
				ContainerName: "media-test",
				ImageName:     "backend/media:latest",
				HostPort:      8080,
				ContainerPort: 8080,
				DaprGrpcPort:  50001,
				AppID:         "media-api",
			}
			
			// This should fail because image doesn't exist yet
			_, err := components.DeployServiceContainer(ctx, config)
			require.Error(t, err, "Should fail when image doesn't exist")
			assert.Contains(t, err.Error(), "image not found", "Error should indicate missing image")
			
			return nil
		})
	})
}

// TestImageBuildIntegration validates image building integrates with deployment orchestrator
func TestImageBuildIntegration(t *testing.T) {
	t.Run("ValidateOrchestratorImageBuildingPhase", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "orchestrator-image-build")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			orchestrator := NewDeploymentOrchestrator(ctx, cfg, "development")
			
			// Test that orchestrator has image building phase
			err := orchestrator.BuildRequiredImages()
			require.NoError(t, err, "Orchestrator should build all required images successfully")
			
			// Validate that all expected images exist after building
			builder := NewImageBuilder(ctx, "development")
			
			expectedImages := []struct {
				name string
				ref  string
			}{
				{"media", "backend/media:latest"},
				{"donations", "backend/donations:latest"},
				{"volunteers", "backend/volunteers:latest"},
				{"business", "backend/business:latest"},
				{"research", "backend/research:latest"},
				{"services", "backend/services:latest"},
				{"events", "backend/events:latest"},
				{"news", "backend/news:latest"},
				{"admin-gateway", "backend/admin-gateway:latest"},
				{"public-gateway", "backend/public-gateway:latest"},
				{"website", "website:latest"},
			}
			
			for _, img := range expectedImages {
				exists, err := builder.ImageExists(img.ref)
				require.NoError(t, err, "Should check %s image existence without error", img.name)
				assert.True(t, exists, "Image %s should exist after orchestrator build", img.ref)
			}
			
			return nil
		})
	})
}

// TestImageBuildingTimeouts validates all image operations respect timeout constraints  
func TestImageBuildingTimeouts(t *testing.T) {
	t.Run("ValidateImageBuildingTimeout", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping timeout test in short mode")
		}
		
		framework := NewContractTestingFramework("international-center", "image-build-timeout")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			builder := NewImageBuilder(ctx, "development")
			
			// Test that image building respects reasonable timeouts
			start := time.Now()
			_, err := builder.BuildServiceImage("media", "inquiries")
			duration := time.Since(start)
			
			// Image builds should complete within 15 seconds for integration tests
			assert.NoError(t, err, "Image building should not timeout")
			assert.Less(t, duration, 15*time.Second, "Image build should complete within 15 seconds")
			
			return nil
		})
	})
}