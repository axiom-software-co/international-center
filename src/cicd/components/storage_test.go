package components

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
)

// TestAzuriteContainerDeployment_Development validates that azurite-dev container is deployed and running
func TestAzuriteContainerDeployment_Development(t *testing.T) {
	t.Run("AzuriteContainerExists_Development", func(t *testing.T) {
		validateAzuriteContainerExists(t, "azurite-dev")
	})

	t.Run("AzuriteContainerRunning_Development", func(t *testing.T) {
		validateAzuriteContainerRunning(t, "azurite-dev", "mcr.microsoft.com/azure-storage/azurite", []string{"10000", "10001", "10002"})
	})

	t.Run("AzuriteContainerHealthy_Development", func(t *testing.T) {
		validateAzuriteContainerHealthy(t, "azurite-dev")
	})
}

// validateAzuriteContainerExists checks if azurite container exists
func validateAzuriteContainerExists(t *testing.T, name string) {
	cmd := exec.Command("podman", "ps", "-a", "--filter", "name="+name, "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check for azurite container %s: %v", name, err)
	}

	containerNames := strings.TrimSpace(string(output))
	assert.Contains(t, containerNames, name, "Azurite container %s should exist", name)
}

// validateAzuriteContainerRunning checks if azurite container is running with correct image and ports
func validateAzuriteContainerRunning(t *testing.T, name, expectedImage string, expectedPorts []string) {
	cmd := exec.Command("podman", "ps", "--filter", "name="+name, "--format", "{{.Names}}\t{{.Image}}\t{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check azurite container %s status: %v", name, err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("Azurite container %s is not running", name)
	}

	parts := strings.Split(lines[0], "\t")
	if len(parts) < 3 {
		t.Fatalf("Unexpected azurite container %s output format", name)
	}

	containerName := parts[0]
	containerImage := parts[1]
	containerPorts := parts[2]

	assert.Equal(t, name, containerName, "Container name should match")
	assert.Equal(t, expectedImage, containerImage, "Container should use correct azurite image")
	
	for _, port := range expectedPorts {
		assert.Contains(t, containerPorts, port, "Container should expose port %s", port)
	}
}

// validateAzuriteContainerHealthy checks if azurite container is healthy
func validateAzuriteContainerHealthy(t *testing.T, name string) {
	cmd := exec.Command("podman", "ps", "--filter", "name="+name, "--format", "{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check azurite container %s health: %v", name, err)
	}

	status := strings.TrimSpace(string(output))
	assert.NotEmpty(t, status, "Azurite container %s should have status", name)
	assert.Contains(t, strings.ToLower(status), "up", "Azurite container %s should be running (Up status)", name)
}

// TestStorageComponent_DevelopmentEnvironment tests storage component for development environment
func TestStorageComponent_DevelopmentEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployStorage(ctx, cfg, "development")
		if err != nil {
			return err
		}

		// Verify development environment generates Azurite emulator configuration
		pulumi.All(outputs.StorageType, outputs.ConnectionString, outputs.AccountName, outputs.ContainerName).ApplyT(func(args []interface{}) error {
			storageType := args[0].(string)
			connectionString := args[1].(string)
			accountName := args[2].(string)
			containerName := args[3].(string)

			assert.Equal(t, "azurite_podman", storageType, "Development should use Azurite emulator")
			assert.Contains(t, connectionString, "AccountName=devstoreaccount1", "Should use Azurite default connection string")
			assert.Equal(t, "devstoreaccount1", accountName, "Should use Azurite default account name")
			assert.Equal(t, "international-center-dev", containerName, "Should use development container name")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &StorageMocks{}))

	assert.NoError(t, err)
}

// TestStorageComponent_StagingEnvironment tests storage component for staging environment
func TestStorageComponent_StagingEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployStorage(ctx, cfg, "staging")
		if err != nil {
			return err
		}

		// Verify staging environment generates Azure Blob Storage with appropriate configuration
		pulumi.All(outputs.StorageType, outputs.ConnectionString, outputs.ReplicationType, outputs.AccessTier).ApplyT(func(args []interface{}) error {
			storageType := args[0].(string)
			connectionString := args[1].(string)
			replicationType := args[2].(string)
			accessTier := args[3].(string)

			assert.Equal(t, "azure_blob_storage", storageType, "Staging should use Azure Blob Storage")
			assert.Contains(t, connectionString, ".blob.core.windows.net", "Should generate Azure Blob Storage connection string")
			assert.Equal(t, "LRS", replicationType, "Should configure staging replication type")
			assert.Equal(t, "hot", accessTier, "Should configure staging access tier")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &StorageMocks{}))

	assert.NoError(t, err)
}

// TestStorageComponent_ProductionEnvironment tests storage component for production environment
func TestStorageComponent_ProductionEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployStorage(ctx, cfg, "production")
		if err != nil {
			return err
		}

		// Verify production environment generates Azure Blob Storage with production features
		pulumi.All(outputs.StorageType, outputs.ConnectionString, outputs.ReplicationType, outputs.BackupEnabled, outputs.AccessTier).ApplyT(func(args []interface{}) error {
			storageType := args[0].(string)
			connectionString := args[1].(string)
			replicationType := args[2].(string)
			backupEnabled := args[3].(bool)
			accessTier := args[4].(string)

			assert.Equal(t, "azure_blob_storage", storageType, "Production should use Azure Blob Storage")
			assert.Contains(t, connectionString, ".blob.core.windows.net", "Should generate Azure Blob Storage connection string")
			assert.Equal(t, "ZRS", replicationType, "Should configure production replication type")
			assert.True(t, backupEnabled, "Should enable backup for production")
			assert.Equal(t, "hot", accessTier, "Should configure production access tier")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &StorageMocks{}))

	assert.NoError(t, err)
}

// TestStorageComponent_EnvironmentParity tests that all environments support required features
func TestStorageComponent_EnvironmentParity(t *testing.T) {
	environments := []string{"development", "staging", "production"}

	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				cfg := config.New(ctx, "")

				outputs, err := DeployStorage(ctx, cfg, env)
				if err != nil {
					return err
				}

				// Verify all environments provide required outputs
				pulumi.All(outputs.ConnectionString, outputs.AccountName, outputs.ContainerName).ApplyT(func(args []interface{}) error {
					connectionString := args[0].(string)
					accountName := args[1].(string)
					containerName := args[2].(string)

					assert.NotEmpty(t, connectionString, "All environments should provide connection string")
					assert.NotEmpty(t, accountName, "All environments should provide account name")
					assert.NotEmpty(t, containerName, "All environments should provide container name")
					return nil
				})

				return nil
			}, pulumi.WithMocks("test", "stack", &StorageMocks{}))

			assert.NoError(t, err)
		})
	}
}

// StorageMocks provides mocks for Pulumi testing
type StorageMocks struct{}

func (mocks *StorageMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}

	switch args.TypeToken {
	case "docker:index/container:Container":
		outputs["name"] = resource.NewStringProperty("azurite-dev")
		outputs["image"] = resource.NewStringProperty("mcr.microsoft.com/azure-storage/azurite")
		outputs["ports"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{
				"internal": resource.NewNumberProperty(10000),
				"external": resource.NewNumberProperty(10000),
			}),
		})

	case "azure-native:storage:StorageAccount":
		outputs["name"] = resource.NewStringProperty("internationalcenterstore")
		outputs["primaryEndpoints"] = resource.NewObjectProperty(resource.PropertyMap{
			"blob": resource.NewStringProperty("https://internationalcenterstore.blob.core.windows.net/"),
		})
		outputs["primaryAccessKey"] = resource.NewStringProperty("mock-access-key")

	case "azure-native:storage:BlobContainer":
		outputs["name"] = resource.NewStringProperty("international-center")
		outputs["publicAccess"] = resource.NewStringProperty("None")
	}

	return args.Name + "_id", outputs, nil
}

func (mocks *StorageMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}
	return outputs, nil
}

