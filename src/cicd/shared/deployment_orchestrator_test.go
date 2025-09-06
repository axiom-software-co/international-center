package shared

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiom-software-co/international-center/src/cicd/components"
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
				Database:      &components.DatabaseOutputs{
					DeploymentType:   pulumi.String("container").ToStringOutput(),
					InstanceType:     pulumi.String("test").ToStringOutput(),
					ConnectionString: pulumi.String("postgresql://localhost:5432/test").ToStringOutput(),
					Port:            pulumi.Int(5432).ToIntOutput(),
					DatabaseName:    pulumi.String("test").ToStringOutput(),
					StorageSize:     pulumi.String("1GB").ToStringOutput(),
					BackupRetention: pulumi.String("7days").ToStringOutput(),
					HighAvailability: pulumi.Bool(false).ToBoolOutput(),
				},
				Storage:       &components.StorageOutputs{
					StorageType:      pulumi.String("container").ToStringOutput(),
					ConnectionString: pulumi.String("http://localhost:10000").ToStringOutput(),
					AccountName:      pulumi.String("test").ToStringOutput(),
					ContainerName:    pulumi.String("test-container").ToStringOutput(),
					ReplicationType:  pulumi.String("LRS").ToStringOutput(),
					AccessTier:       pulumi.String("Hot").ToStringOutput(),
					BackupEnabled:    pulumi.Bool(false).ToBoolOutput(),
				},
				Vault:         &components.VaultOutputs{
					VaultType:     pulumi.String("container").ToStringOutput(),
					VaultAddress:  pulumi.String("http://localhost:8200").ToStringOutput(),
					AuthMethod:    pulumi.String("token").ToStringOutput(),
					SecretEngine:  pulumi.String("kv").ToStringOutput(),
					ClusterTier:   pulumi.String("dev").ToStringOutput(),
					AuditEnabled:  pulumi.Bool(true).ToBoolOutput(),
				},
				Observability: &components.ObservabilityOutputs{
					StackType:       pulumi.String("container").ToStringOutput(),
					GrafanaURL:      pulumi.String("http://localhost:3000").ToStringOutput(),
					PrometheusURL:   pulumi.String("http://localhost:9090").ToStringOutput(),
					LokiURL:         pulumi.String("http://localhost:3100").ToStringOutput(),
					RetentionDays:   pulumi.Int(30).ToIntOutput(),
					AuditLogging:    pulumi.Bool(true).ToBoolOutput(),
					AlertingEnabled: pulumi.Bool(false).ToBoolOutput(),
				},
				Dapr:          &components.DaprOutputs{
					DeploymentType:      pulumi.String("container").ToStringOutput(),
					RuntimePort:         pulumi.Int(3500).ToIntOutput(),
					ControlPlaneURL:     pulumi.String("http://localhost:50005").ToStringOutput(),
					SidecarConfig:       pulumi.String("development").ToStringOutput(),
					MiddlewareEnabled:   pulumi.Bool(true).ToBoolOutput(),
					PolicyEnabled:       pulumi.Bool(true).ToBoolOutput(),
				},
				Services:      &components.ServicesOutputs{
					DeploymentType:        pulumi.String("container").ToStringOutput(),
					InquiriesServices:     pulumi.Map{"inquiries": pulumi.String("http://localhost:8001")}.ToMapOutput(),
					ContentServices:       pulumi.Map{"content": pulumi.String("http://localhost:8002")}.ToMapOutput(),
					GatewayServices:       pulumi.Map{"public": pulumi.String("http://localhost:8080")}.ToMapOutput(),
					TestServices:          pulumi.Map{"test": pulumi.String("http://localhost:8003")}.ToMapOutput(),
					APIServices:           pulumi.Map{"api": pulumi.String("http://localhost:8004")}.ToMapOutput(),
					PublicGatewayURL:      pulumi.String("http://localhost:8080").ToStringOutput(),
					AdminGatewayURL:       pulumi.String("http://localhost:8081").ToStringOutput(),
					HealthCheckEnabled:    pulumi.Bool(true).ToBoolOutput(),
					DaprSidecarEnabled:    pulumi.Bool(true).ToBoolOutput(),
					ObservabilityEnabled:  pulumi.Bool(true).ToBoolOutput(),
					TestingEnabled:        pulumi.Bool(true).ToBoolOutput(),
					ScalingPolicy:         pulumi.String("manual").ToStringOutput(),
					SecurityPolicies:      pulumi.Bool(true).ToBoolOutput(),
					AuditLogging:          pulumi.Bool(true).ToBoolOutput(),
				},
				Website:       &components.WebsiteOutputs{
					DeploymentType:        pulumi.String("container").ToStringOutput(),
					ContainerID:           pulumi.String("website-dev").ToStringOutput(),
					ContainerStatus:       pulumi.String("running").ToStringOutput(),
					ServerURL:             pulumi.String("http://localhost:3000").ToStringOutput(),
					BuildCommand:          pulumi.String("pnpm run build").ToStringOutput(),
					BuildDirectory:        pulumi.String("dist").ToStringOutput(),
					NodeVersion:           pulumi.String("20").ToStringOutput(),
					CDNEnabled:            pulumi.Bool(false).ToBoolOutput(),
					CachePolicy:           pulumi.String("no-cache").ToStringOutput(),
					CompressionEnabled:    pulumi.Bool(true).ToBoolOutput(),
					SecurityHeaders:       pulumi.Bool(true).ToBoolOutput(),
					APIGatewayURL:         pulumi.String("http://localhost:8080").ToStringOutput(),
					APIIntegrationEnabled: pulumi.Bool(true).ToBoolOutput(),
				},
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

