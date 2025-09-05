package shared

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeploymentOrchestratorImageBuilding validates orchestrator handles image building
func TestDeploymentOrchestratorImageBuilding(t *testing.T) {
	t.Run("ValidateOrchestratorHasImageBuildingPhase", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "orchestrator-image-build-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			orchestrator := NewDeploymentOrchestrator(ctx, cfg, "development")
			require.NotNil(t, orchestrator, "Orchestrator should be instantiated")
			
			// Test that orchestrator has BuildRequiredImages method
			err := orchestrator.BuildRequiredImages()
			require.NoError(t, err, "Orchestrator should have BuildRequiredImages method")
			
			return nil
		})
	})
	
	t.Run("ValidateOrchestratorImageBuildingContract", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "orchestrator-build-contract-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			orchestrator := NewDeploymentOrchestrator(ctx, cfg, "development")
			
			// Test that image building precedes infrastructure deployment
			err := orchestrator.BuildRequiredImages()
			require.NoError(t, err, "Image building should complete successfully")
			
			// After building images, validate they exist
			builder := NewImageBuilder(ctx, "development")
			
			// Test critical service images exist after orchestrator build
			expectedImages := []struct {
				name string
				ref  string
			}{
				{"media", "backend/media:latest"},
				{"donations", "backend/donations:latest"},
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
	
	t.Run("ValidateOrchestratorFailsWithoutImages", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "orchestrator-no-images-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			orchestrator := NewDeploymentOrchestrator(ctx, cfg, "development")
			
			// Test that infrastructure deployment fails without images
			_, err := orchestrator.DeployInfrastructure()
			require.Error(t, err, "Infrastructure deployment should fail without images")
			assert.Contains(t, err.Error(), "image not found", "Error should indicate missing images")
			
			return nil
		})
	})
	
	t.Run("ValidateOrchestratorSucceedsWithImages", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "orchestrator-with-images-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			orchestrator := NewDeploymentOrchestrator(ctx, cfg, "development")
			
			// Build required images first
			err := orchestrator.BuildRequiredImages()
			require.NoError(t, err, "Image building should succeed")
			
			// Test that infrastructure deployment succeeds with images
			outputs, err := orchestrator.DeployInfrastructure()
			require.NoError(t, err, "Infrastructure deployment should succeed with images")
			require.NotNil(t, outputs, "Infrastructure outputs should not be nil")
			
			// Validate all components are deployed successfully
			assert.NotNil(t, outputs.Database, "Database should be deployed")
			assert.NotNil(t, outputs.Storage, "Storage should be deployed")
			assert.NotNil(t, outputs.Vault, "Vault should be deployed")
			assert.NotNil(t, outputs.Observability, "Observability should be deployed")
			assert.NotNil(t, outputs.Dapr, "Dapr should be deployed")
			assert.NotNil(t, outputs.Services, "Services should be deployed")
			assert.NotNil(t, outputs.Website, "Website should be deployed")
			
			return nil
		})
	})
}

// TestDeploymentOrchestratorImageLifecycle validates orchestrator manages image lifecycle
func TestDeploymentOrchestratorImageLifecycle(t *testing.T) {
	t.Run("ValidateOrchestratorImageRollback", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "orchestrator-image-rollback-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			orchestrator := NewDeploymentOrchestrator(ctx, cfg, "development")
			
			// Build images for rollback scenario
			err := orchestrator.BuildRequiredImages()
			require.NoError(t, err, "Image building should succeed")
			
			// Test orchestrator can rollback with image cleanup
			outputs := &ComponentOutputs{
				Database:      &DummyDatabaseOutputs{},
				Storage:       &DummyStorageOutputs{},
				Vault:         &DummyVaultOutputs{},
				Observability: &DummyObservabilityOutputs{},
				Dapr:          &DummyDaprOutputs{},
				Services:      &DummyServicesOutputs{},
				Website:       &DummyWebsiteOutputs{},
			}
			
			err = orchestrator.performRollback(outputs)
			assert.NoError(t, err, "Orchestrator should perform rollback successfully")
			
			return nil
		})
	})
	
	t.Run("ValidateOrchestratorImageHealthCheck", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "orchestrator-image-health-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			orchestrator := NewDeploymentOrchestrator(ctx, cfg, "development")
			
			// Build images for health check scenario
			err := orchestrator.BuildRequiredImages()
			require.NoError(t, err, "Image building should succeed")
			
			// Test that orchestrator can validate image health
			builder := NewImageBuilder(ctx, "development")
			
			// Validate images are healthy after build
			serviceImages := []string{"backend/media:latest", "backend/donations:latest", "website:latest"}
			for _, image := range serviceImages {
				exists, err := builder.ImageExists(image)
				require.NoError(t, err, "Should check image %s existence", image)
				assert.True(t, exists, "Image %s should exist and be healthy", image)
			}
			
			return nil
		})
	})
}

// TestDeploymentOrchestratorImageIntegration validates orchestrator integrates with image operations
func TestDeploymentOrchestratorImageIntegration(t *testing.T) {
	t.Run("ValidateOrchestratorImageBuildingIntegration", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "orchestrator-integration-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			orchestrator := NewDeploymentOrchestrator(ctx, cfg, "development")
			
			// Test integrated workflow: build images then deploy infrastructure
			err := orchestrator.BuildRequiredImages()
			require.NoError(t, err, "Image building phase should succeed")
			
			// Deploy infrastructure with built images
			outputs, err := orchestrator.DeployInfrastructure()
			require.NoError(t, err, "Infrastructure deployment should succeed after image build")
			
			// Validate deployment health includes image validation
			health := orchestrator.GetDeploymentHealth(outputs)
			assert.True(t, health["database"], "Database should be healthy")
			assert.True(t, health["services"], "Services should be healthy")
			assert.True(t, health["website"], "Website should be healthy")
			
			return nil
		})
	})
	
	t.Run("ValidateOrchestratorImageErrorHandling", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "orchestrator-error-handling-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			orchestrator := NewDeploymentOrchestrator(ctx, cfg, "development")
			
			// Test orchestrator handles image build failures gracefully
			// This test validates error propagation without building actual images
			_, err := orchestrator.DeployInfrastructure()
			require.Error(t, err, "Should fail when images don't exist")
			
			// Validate error contains image-related information
			assert.Contains(t, err.Error(), "image", "Error should reference image problems")
			
			return nil
		})
	})
}

// Dummy output types for testing rollback scenarios
type DummyDatabaseOutputs struct{}
type DummyStorageOutputs struct{}
type DummyVaultOutputs struct{}
type DummyObservabilityOutputs struct{}
type DummyDaprOutputs struct{}
type DummyServicesOutputs struct{}
type DummyWebsiteOutputs struct{}