package testing

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// IntegrationTestMocks provides comprehensive mocks for integration testing
type IntegrationTestMocks struct{}

func (mocks *IntegrationTestMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}

	switch args.TypeToken {
	// Database component mocks
	case "postgresql:index/database:Database":
		outputs["name"] = resource.NewStringProperty("international_center_dev")
		outputs["connectionString"] = resource.NewStringProperty("postgresql://postgres:password@localhost:5432/international_center_dev")
		outputs["port"] = resource.NewNumberProperty(5432)
		outputs["deploymentType"] = resource.NewStringProperty("container")
		outputs["instanceType"] = resource.NewStringProperty("postgresql")
		outputs["databaseName"] = resource.NewStringProperty("international_center_dev")
		outputs["storageSize"] = resource.NewStringProperty("10GB")
		outputs["backupRetention"] = resource.NewStringProperty("7 days")
		outputs["highAvailability"] = resource.NewBoolProperty(false)

	// Storage component mocks
	case "azure-native:storage:StorageAccount":
		outputs["name"] = resource.NewStringProperty("intlcenterdevstorage")
		outputs["connectionString"] = resource.NewStringProperty("DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;")
		outputs["deploymentType"] = resource.NewStringProperty("azurite")
		outputs["storageType"] = resource.NewStringProperty("blob_storage")
		outputs["accessTier"] = resource.NewStringProperty("hot")
		outputs["encryption"] = resource.NewBoolProperty(true)
		outputs["publicAccess"] = resource.NewBoolProperty(false)

	// Vault component mocks
	case "vault:index/mount:Mount":
		outputs["path"] = resource.NewStringProperty("secret")
		outputs["type"] = resource.NewStringProperty("kv-v2")
		outputs["vaultURL"] = resource.NewStringProperty("http://localhost:8200")
		outputs["deploymentType"] = resource.NewStringProperty("container")
		outputs["vaultType"] = resource.NewStringProperty("dev_server")
		outputs["sealStatus"] = resource.NewStringProperty("unsealed")
		outputs["authMethod"] = resource.NewStringProperty("token")
		outputs["highAvailability"] = resource.NewBoolProperty(false)

	// Observability component mocks
	case "grafana:index/dashboard:Dashboard":
		outputs["title"] = resource.NewStringProperty("International Center Development Dashboard")
		outputs["grafanaURL"] = resource.NewStringProperty("http://localhost:3000")
		outputs["deploymentType"] = resource.NewStringProperty("containers")
		outputs["observabilityStack"] = resource.NewStringProperty("grafana_stack")
		outputs["dataRetention"] = resource.NewStringProperty("7 days")
		outputs["alerting"] = resource.NewBoolProperty(true)
		outputs["auditLogging"] = resource.NewBoolProperty(true)

	// Dapr component mocks
	case "dapr:index/component:Component":
		outputs["name"] = resource.NewStringProperty("international-center-dapr")
		outputs["version"] = resource.NewStringProperty("v1.0")
		outputs["deploymentType"] = resource.NewStringProperty("self_hosted")
		outputs["runtimeVersion"] = resource.NewStringProperty("1.12")
		outputs["controlPlaneURL"] = resource.NewStringProperty("http://localhost:3500")
		outputs["sidecarEnabled"] = resource.NewBoolProperty(true)
		outputs["serviceInvocation"] = resource.NewBoolProperty(true)
		outputs["stateStore"] = resource.NewBoolProperty(true)
		outputs["pubSub"] = resource.NewBoolProperty(true)
		outputs["secretStore"] = resource.NewBoolProperty(true)
		outputs["observability"] = resource.NewBoolProperty(true)

	// Services component mocks
	case "docker:index/container:Container":
		outputs["name"] = resource.NewStringProperty(args.Name)
		outputs["image"] = resource.NewStringProperty("nginx:alpine")
		outputs["restart"] = resource.NewStringProperty("unless-stopped")
		outputs["deploymentType"] = resource.NewStringProperty("containers")
		outputs["publicGatewayURL"] = resource.NewStringProperty("http://localhost:8080")
		outputs["adminGatewayURL"] = resource.NewStringProperty("http://localhost:8081")
		outputs["healthCheckEnabled"] = resource.NewBoolProperty(true)
		outputs["daprSidecarEnabled"] = resource.NewBoolProperty(true)
		outputs["observabilityEnabled"] = resource.NewBoolProperty(true)
		outputs["scalingPolicy"] = resource.NewStringProperty("manual")
		outputs["securityPolicies"] = resource.NewBoolProperty(true)
		outputs["auditLogging"] = resource.NewBoolProperty(true)

	// Website component mocks
	case "cloudflare:index/pagesProject:PagesProject":
		outputs["name"] = resource.NewStringProperty("international-center-website")
		outputs["subdomain"] = resource.NewStringProperty("international-center")
		outputs["websiteURL"] = resource.NewStringProperty("https://international-center.axiomcloud.dev")
		outputs["deploymentType"] = resource.NewStringProperty("cloudflare_pages")
		outputs["buildCommand"] = resource.NewStringProperty("pnpm build")
		outputs["outputDirectory"] = resource.NewStringProperty("dist")
		outputs["cdnEnabled"] = resource.NewBoolProperty(true)
		outputs["compressionEnabled"] = resource.NewBoolProperty(true)
		outputs["cachePolicy"] = resource.NewStringProperty("aggressive")

	// Migration component mocks
	case "golang-migrate:index/migration:Migration":
		outputs["version"] = resource.NewStringProperty("20240101000000")
		outputs["name"] = resource.NewStringProperty("initial_schema")
		outputs["executionStatus"] = resource.NewStringProperty("completed")
		outputs["migrationsRun"] = resource.NewNumberProperty(8)
		outputs["errorMessage"] = resource.NewStringProperty("")
		outputs["duration"] = resource.NewStringProperty("2m30s")

	default:
		// Default mock outputs for unknown resource types
		outputs["name"] = resource.NewStringProperty(args.Name)
		outputs["id"] = resource.NewStringProperty(args.Name + "_integration_test_id")
		outputs["status"] = resource.NewStringProperty("healthy")
	}

	return args.Name + "_integration_id", outputs, nil
}

func (mocks *IntegrationTestMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}

	switch args.Token {
	// Database function calls
	case "postgresql:index/getDatabase:getDatabase":
		outputs["name"] = resource.NewStringProperty("international_center_dev")
		outputs["encoding"] = resource.NewStringProperty("UTF8")
		outputs["owner"] = resource.NewStringProperty("postgres")
		outputs["connectionString"] = resource.NewStringProperty("postgresql://postgres:password@localhost:5432/international_center_dev")

	// Storage function calls
	case "azure-native:storage/getBlobContainer:getBlobContainer":
		outputs["name"] = resource.NewStringProperty("app-data")
		outputs["publicAccess"] = resource.NewStringProperty("None")
		outputs["metadata"] = resource.NewObjectProperty(resource.PropertyMap{
			"environment": resource.NewStringProperty("development"),
			"purpose":     resource.NewStringProperty("application_data"),
		})

	// Vault function calls
	case "vault:index/getMount:getMount":
		outputs["path"] = resource.NewStringProperty("secret")
		outputs["type"] = resource.NewStringProperty("kv-v2")
		outputs["description"] = resource.NewStringProperty("Development secrets store")

	// Observability function calls
	case "grafana:index/getDashboard:getDashboard":
		outputs["title"] = resource.NewStringProperty("International Center Development")
		outputs["tags"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty("development"),
			resource.NewStringProperty("monitoring"),
		})

	// Dapr function calls
	case "dapr:index/getComponent:getComponent":
		outputs["name"] = resource.NewStringProperty("international-center-state-store")
		outputs["type"] = resource.NewStringProperty("state.postgresql")
		outputs["version"] = resource.NewStringProperty("v1")

	// Services function calls
	case "docker:index/getContainer:getContainer":
		outputs["name"] = resource.NewStringProperty("public-gateway")
		outputs["status"] = resource.NewStringProperty("running")
		outputs["ports"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{
				"internal": resource.NewNumberProperty(80),
				"external": resource.NewNumberProperty(8080),
			}),
		})

	// Website function calls
	case "cloudflare:index/getPagesProject:getPagesProject":
		outputs["name"] = resource.NewStringProperty("international-center-website")
		outputs["domains"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty("international-center.axiomcloud.dev"),
		})

	// Migration function calls
	case "golang-migrate:index/getVersion:getVersion":
		outputs["version"] = resource.NewStringProperty("20240101000000")
		outputs["dirty"] = resource.NewBoolProperty(false)
		outputs["migrationsApplied"] = resource.NewNumberProperty(8)

	default:
		// Default mock outputs for unknown function calls
		outputs["result"] = resource.NewStringProperty("integration-test-result")
		outputs["status"] = resource.NewStringProperty("success")
	}

	return outputs, nil
}