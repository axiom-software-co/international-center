package components

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
)

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

