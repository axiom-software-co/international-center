package components

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

// TestDatabaseComponent_DevelopmentEnvironment tests database component for development environment
func TestDatabaseComponent_DevelopmentEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployDatabase(ctx, cfg, "development")
		if err != nil {
			return err
		}

		// Verify development environment generates container-based PostgreSQL
		pulumi.All(outputs.DeploymentType, outputs.InstanceType, outputs.ConnectionString, outputs.Port).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			instanceType := args[1].(string)
			connectionString := args[2].(string)
			port := args[3].(int)

			assert.Equal(t, "podman_container", deploymentType, "Development should use container deployment type")
			assert.Equal(t, "postgresql", instanceType, "Development should use PostgreSQL instance type")
			assert.Contains(t, connectionString, "postgresql://", "Should generate PostgreSQL connection string")
			assert.Equal(t, 5432, port, "Should use standard PostgreSQL port")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &DatabaseMocks{}))

	assert.NoError(t, err)
}

// TestDatabaseComponent_StagingEnvironment tests database component for staging environment
func TestDatabaseComponent_StagingEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployDatabase(ctx, cfg, "staging")
		if err != nil {
			return err
		}

		// Verify staging environment generates Azure PostgreSQL with appropriate scaling
		pulumi.All(outputs.DeploymentType, outputs.InstanceType, outputs.ConnectionString, outputs.StorageSize, outputs.BackupRetention).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			instanceType := args[1].(string)
			connectionString := args[2].(string)
			storageSize := args[3].(string)
			backupRetention := args[4].(string)

			assert.Equal(t, "azure_postgresql", deploymentType, "Staging should use Azure PostgreSQL deployment type")
			assert.Equal(t, "flexible_server", instanceType, "Staging should use flexible server instance type")
			assert.Contains(t, connectionString, ".postgres.database.azure.com", "Should generate Azure PostgreSQL connection string")
			assert.Equal(t, "50GB", storageSize, "Should configure staging storage size")
			assert.Equal(t, "14d", backupRetention, "Should configure staging backup retention")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &DatabaseMocks{}))

	assert.NoError(t, err)
}

// TestDatabaseComponent_ProductionEnvironment tests database component for production environment
func TestDatabaseComponent_ProductionEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployDatabase(ctx, cfg, "production")
		if err != nil {
			return err
		}

		// Verify production environment generates Azure PostgreSQL with production features
		pulumi.All(outputs.DeploymentType, outputs.InstanceType, outputs.ConnectionString, outputs.StorageSize, outputs.BackupRetention, outputs.HighAvailability).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			instanceType := args[1].(string)
			connectionString := args[2].(string)
			storageSize := args[3].(string)
			backupRetention := args[4].(string)
			highAvailability := args[5].(bool)

			assert.Equal(t, "azure_postgresql", deploymentType, "Production should use Azure PostgreSQL deployment type")
			assert.Equal(t, "flexible_server", instanceType, "Production should use flexible server instance type")
			assert.Contains(t, connectionString, ".postgres.database.azure.com", "Should generate Azure PostgreSQL connection string")
			assert.Equal(t, "100GB", storageSize, "Should configure production storage size")
			assert.Equal(t, "30d", backupRetention, "Should configure production backup retention")
			assert.True(t, highAvailability, "Should enable high availability for production")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &DatabaseMocks{}))

	assert.NoError(t, err)
}

// TestDatabaseComponent_EnvironmentParity tests that all environments support required features
func TestDatabaseComponent_EnvironmentParity(t *testing.T) {
	environments := []string{"development", "staging", "production"}

	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				cfg := config.New(ctx, "")

				outputs, err := DeployDatabase(ctx, cfg, env)
				if err != nil {
					return err
				}

				// Verify all environments provide required outputs
				pulumi.All(outputs.ConnectionString, outputs.Port, outputs.DatabaseName).ApplyT(func(args []interface{}) error {
					connectionString := args[0].(string)
					port := args[1].(int)
					databaseName := args[2].(string)

					assert.NotEmpty(t, connectionString, "All environments should provide connection string")
					assert.Greater(t, port, 0, "All environments should provide valid port")
					assert.NotEmpty(t, databaseName, "All environments should provide database name")
					return nil
				})

				return nil
			}, pulumi.WithMocks("test", "stack", &DatabaseMocks{}))

			assert.NoError(t, err)
		})
	}
}

// DatabaseMocks provides mocks for Pulumi testing
type DatabaseMocks struct{}

func (mocks *DatabaseMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}

	switch args.TypeToken {
	case "docker:index/container:Container":
		outputs["name"] = resource.NewStringProperty("postgresql-dev")
		outputs["image"] = resource.NewStringProperty("postgres:15")
		outputs["ports"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{
				"internal": resource.NewNumberProperty(5432),
				"external": resource.NewNumberProperty(5432),
			}),
		})

	case "azure-native:dbforpostgresql:Server":
		outputs["fullyQualifiedDomainName"] = resource.NewStringProperty("test-postgresql.postgres.database.azure.com")
		outputs["administratorLogin"] = resource.NewStringProperty("admin")
		outputs["version"] = resource.NewStringProperty("15")

	case "azure-native:dbforpostgresql:Configuration":
		outputs["name"] = resource.NewStringProperty("shared_preload_libraries")
		outputs["value"] = resource.NewStringProperty("pg_stat_statements")
	}

	return args.Name + "_id", outputs, nil
}

func (mocks *DatabaseMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}
	return outputs, nil
}

