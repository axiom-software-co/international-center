package migration

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// MigrationMocks provides mocks for Pulumi testing of migration runner
type MigrationMocks struct{}

func (mocks *MigrationMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}

	switch args.TypeToken {
	case "postgresql:index/database:Database":
		outputs["name"] = resource.NewStringProperty("international_center")
		outputs["owner"] = resource.NewStringProperty("postgres")
		outputs["connectionLimit"] = resource.NewNumberProperty(-1)

	case "migration:index/runner:Runner":
		outputs["strategy"] = resource.NewStringProperty("aggressive")
		outputs["connectionString"] = resource.NewStringProperty("postgresql://user:password@localhost:5432/international_center")
		outputs["migrationFiles"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty("/migrations/business.sql"),
			resource.NewStringProperty("/migrations/donations.sql"),
			resource.NewStringProperty("/migrations/events.sql"),
			resource.NewStringProperty("/migrations/media.sql"),
			resource.NewStringProperty("/migrations/news.sql"),
			resource.NewStringProperty("/migrations/research.sql"),
			resource.NewStringProperty("/migrations/services.sql"),
			resource.NewStringProperty("/migrations/volunteers.sql"),
		})
		outputs["safetyChecks"] = resource.NewBoolProperty(false)
		outputs["rollbackEnabled"] = resource.NewBoolProperty(false)
		outputs["timeoutMinutes"] = resource.NewNumberProperty(5)
		outputs["backupRequired"] = resource.NewBoolProperty(false)
		outputs["approvalRequired"] = resource.NewBoolProperty(false)
		outputs["executionStatus"] = resource.NewStringProperty("completed")

	case "golang-migrate:index/migration:Migration":
		outputs["version"] = resource.NewStringProperty("20240101000000")
		outputs["name"] = resource.NewStringProperty("initial_schema")
		outputs["up"] = resource.NewStringProperty("CREATE TABLE IF NOT EXISTS schema_migrations...")
		outputs["down"] = resource.NewStringProperty("DROP TABLE IF EXISTS schema_migrations...")

	default:
		// Default mock outputs for unknown resource types
		outputs["name"] = resource.NewStringProperty(args.Name)
		outputs["id"] = resource.NewStringProperty(args.Name + "_mock_id")
	}

	return args.Name + "_id", outputs, nil
}

func (mocks *MigrationMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}

	switch args.Token {
	case "golang-migrate:index/getVersion:getVersion":
		outputs["version"] = resource.NewStringProperty("20240101000000")
		outputs["dirty"] = resource.NewBoolProperty(false)

	case "postgresql:index/getDatabase:getDatabase":
		outputs["name"] = resource.NewStringProperty("international_center")
		outputs["encoding"] = resource.NewStringProperty("UTF8")
		outputs["owner"] = resource.NewStringProperty("postgres")

	default:
		// Default mock outputs for unknown function calls
		outputs["result"] = resource.NewStringProperty("mock-result")
	}

	return outputs, nil
}