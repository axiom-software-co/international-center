package testing

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// HealthValidationMocks provides health-focused mocks for component validation
type HealthValidationMocks struct{}

func (mocks *HealthValidationMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}

	switch args.TypeToken {
	// Database component health mocks
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
		// Health-specific outputs
		outputs["status"] = resource.NewStringProperty("healthy")
		outputs["uptime"] = resource.NewStringProperty("99.9%")
		outputs["lastHealthCheck"] = resource.NewStringProperty("2024-01-01T12:00:00Z")

	// Storage component health mocks
	case "azure-native:storage:StorageAccount":
		outputs["name"] = resource.NewStringProperty("intlcenterdevstorage")
		outputs["connectionString"] = resource.NewStringProperty("DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;")
		outputs["deploymentType"] = resource.NewStringProperty("azurite")
		outputs["storageType"] = resource.NewStringProperty("blob_storage")
		outputs["accessTier"] = resource.NewStringProperty("hot")
		outputs["encryption"] = resource.NewBoolProperty(true)
		outputs["publicAccess"] = resource.NewBoolProperty(false)
		// Health-specific outputs
		outputs["healthStatus"] = resource.NewStringProperty("available")
		outputs["responseTime"] = resource.NewStringProperty("50ms")
		outputs["lastHealthCheck"] = resource.NewStringProperty("2024-01-01T12:00:00Z")

	// Vault component health mocks
	case "vault:index/mount:Mount":
		outputs["path"] = resource.NewStringProperty("secret")
		outputs["type"] = resource.NewStringProperty("kv-v2")
		outputs["vaultURL"] = resource.NewStringProperty("http://localhost:8200")
		outputs["deploymentType"] = resource.NewStringProperty("container")
		outputs["vaultType"] = resource.NewStringProperty("dev_server")
		outputs["sealStatus"] = resource.NewStringProperty("unsealed")
		outputs["authMethod"] = resource.NewStringProperty("token")
		outputs["highAvailability"] = resource.NewBoolProperty(false)
		// Health-specific outputs
		outputs["vaultHealth"] = resource.NewStringProperty("healthy")
		outputs["initialized"] = resource.NewBoolProperty(true)
		outputs["sealed"] = resource.NewBoolProperty(false)
		outputs["standby"] = resource.NewBoolProperty(false)
		outputs["lastHealthCheck"] = resource.NewStringProperty("2024-01-01T12:00:00Z")

	// Observability component health mocks
	case "grafana:index/dashboard:Dashboard":
		outputs["title"] = resource.NewStringProperty("International Center Development Dashboard")
		outputs["grafanaURL"] = resource.NewStringProperty("http://localhost:3000")
		outputs["deploymentType"] = resource.NewStringProperty("containers")
		outputs["observabilityStack"] = resource.NewStringProperty("grafana_stack")
		outputs["dataRetention"] = resource.NewStringProperty("7 days")
		outputs["alerting"] = resource.NewBoolProperty(true)
		outputs["auditLogging"] = resource.NewBoolProperty(true)
		// Health-specific outputs
		outputs["grafanaHealth"] = resource.NewStringProperty("ok")
		outputs["mimiHealth"] = resource.NewStringProperty("ready")
		outputs["lokiHealth"] = resource.NewStringProperty("ready")
		outputs["tempoHealth"] = resource.NewStringProperty("ready")
		outputs["pyroscopeHealth"] = resource.NewStringProperty("ready")
		outputs["lastHealthCheck"] = resource.NewStringProperty("2024-01-01T12:00:00Z")

	// Dapr component health mocks
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
		// Health-specific outputs
		outputs["daprHealth"] = resource.NewStringProperty("healthy")
		outputs["controlPlaneHealth"] = resource.NewStringProperty("running")
		outputs["sidecarHealth"] = resource.NewStringProperty("connected")
		outputs["componentsHealth"] = resource.NewStringProperty("all_healthy")
		outputs["lastHealthCheck"] = resource.NewStringProperty("2024-01-01T12:00:00Z")

	// Services component health mocks
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
		// Health-specific outputs
		outputs["serviceHealth"] = resource.NewStringProperty("healthy")
		outputs["gatewayHealth"] = resource.NewStringProperty("up")
		outputs["sidecarHealth"] = resource.NewStringProperty("connected")
		outputs["healthEndpoint"] = resource.NewStringProperty("/health")
		outputs["healthCheckInterval"] = resource.NewStringProperty("30s")
		outputs["lastHealthCheck"] = resource.NewStringProperty("2024-01-01T12:00:00Z")

	// Website component health mocks
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
		// Health-specific outputs
		outputs["deploymentHealth"] = resource.NewStringProperty("active")
		outputs["cdnHealth"] = resource.NewStringProperty("operational")
		outputs["buildStatus"] = resource.NewStringProperty("success")
		outputs["lastBuildTime"] = resource.NewStringProperty("2024-01-01T12:00:00Z")
		outputs["lastHealthCheck"] = resource.NewStringProperty("2024-01-01T12:00:00Z")

	default:
		// Default health mock outputs for unknown resource types
		outputs["name"] = resource.NewStringProperty(args.Name)
		outputs["id"] = resource.NewStringProperty(args.Name + "_health_test_id")
		outputs["status"] = resource.NewStringProperty("healthy")
		outputs["lastHealthCheck"] = resource.NewStringProperty("2024-01-01T12:00:00Z")
	}

	return args.Name + "_health_id", outputs, nil
}

func (mocks *HealthValidationMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}

	switch args.Token {
	// Database health function calls
	case "postgresql:index/getDatabase:getDatabase":
		outputs["name"] = resource.NewStringProperty("international_center_dev")
		outputs["encoding"] = resource.NewStringProperty("UTF8")
		outputs["owner"] = resource.NewStringProperty("postgres")
		outputs["connectionString"] = resource.NewStringProperty("postgresql://postgres:password@localhost:5432/international_center_dev")
		outputs["health"] = resource.NewObjectProperty(resource.PropertyMap{
			"status":          resource.NewStringProperty("healthy"),
			"uptime":          resource.NewStringProperty("99.9%"),
			"connections":     resource.NewNumberProperty(10),
			"maxConnections": resource.NewNumberProperty(100),
			"responseTime":    resource.NewStringProperty("5ms"),
		})

	// Storage health function calls
	case "azure-native:storage/getBlobContainer:getBlobContainer":
		outputs["name"] = resource.NewStringProperty("app-data")
		outputs["publicAccess"] = resource.NewStringProperty("None")
		outputs["metadata"] = resource.NewObjectProperty(resource.PropertyMap{
			"environment": resource.NewStringProperty("development"),
			"purpose":     resource.NewStringProperty("application_data"),
		})
		outputs["health"] = resource.NewObjectProperty(resource.PropertyMap{
			"status":       resource.NewStringProperty("available"),
			"responseTime": resource.NewStringProperty("50ms"),
			"accessibility": resource.NewStringProperty("accessible"),
		})

	// Vault health function calls
	case "vault:index/getMount:getMount":
		outputs["path"] = resource.NewStringProperty("secret")
		outputs["type"] = resource.NewStringProperty("kv-v2")
		outputs["description"] = resource.NewStringProperty("Development secrets store")
		outputs["health"] = resource.NewObjectProperty(resource.PropertyMap{
			"initialized": resource.NewBoolProperty(true),
			"sealed":      resource.NewBoolProperty(false),
			"standby":     resource.NewBoolProperty(false),
			"version":     resource.NewStringProperty("1.15.0"),
		})

	// Observability health function calls
	case "grafana:index/getDashboard:getDashboard":
		outputs["title"] = resource.NewStringProperty("International Center Development")
		outputs["tags"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty("development"),
			resource.NewStringProperty("monitoring"),
		})
		outputs["health"] = resource.NewObjectProperty(resource.PropertyMap{
			"grafana":   resource.NewStringProperty("ok"),
			"mimir":     resource.NewStringProperty("ready"),
			"loki":      resource.NewStringProperty("ready"),
			"tempo":     resource.NewStringProperty("ready"),
			"pyroscope": resource.NewStringProperty("ready"),
		})

	// Dapr health function calls
	case "dapr:index/getComponent:getComponent":
		outputs["name"] = resource.NewStringProperty("international-center-state-store")
		outputs["type"] = resource.NewStringProperty("state.postgresql")
		outputs["version"] = resource.NewStringProperty("v1")
		outputs["health"] = resource.NewObjectProperty(resource.PropertyMap{
			"status":          resource.NewStringProperty("healthy"),
			"controlPlane":    resource.NewStringProperty("running"),
			"sidecar":         resource.NewStringProperty("connected"),
			"components":      resource.NewStringProperty("all_healthy"),
			"runtimeVersion":  resource.NewStringProperty("1.12"),
		})

	// Services health function calls
	case "docker:index/getContainer:getContainer":
		outputs["name"] = resource.NewStringProperty("public-gateway")
		outputs["status"] = resource.NewStringProperty("running")
		outputs["ports"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{
				"internal": resource.NewNumberProperty(80),
				"external": resource.NewNumberProperty(8080),
			}),
		})
		outputs["health"] = resource.NewObjectProperty(resource.PropertyMap{
			"status":             resource.NewStringProperty("healthy"),
			"gateway":            resource.NewStringProperty("up"),
			"sidecar":            resource.NewStringProperty("connected"),
			"healthEndpoint":     resource.NewStringProperty("/health"),
			"lastHealthCheck":    resource.NewStringProperty("2024-01-01T12:00:00Z"),
			"healthCheckResult":  resource.NewStringProperty("pass"),
		})

	// Website health function calls
	case "cloudflare:index/getPagesProject:getPagesProject":
		outputs["name"] = resource.NewStringProperty("international-center-website")
		outputs["domains"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty("international-center.axiomcloud.dev"),
		})
		outputs["health"] = resource.NewObjectProperty(resource.PropertyMap{
			"deployment": resource.NewStringProperty("active"),
			"cdn":        resource.NewStringProperty("operational"),
			"build":      resource.NewStringProperty("success"),
			"ssl":        resource.NewStringProperty("active"),
			"performance": resource.NewObjectProperty(resource.PropertyMap{
				"loadTime":     resource.NewStringProperty("200ms"),
				"availability": resource.NewStringProperty("99.9%"),
			}),
		})

	default:
		// Default health mock outputs for unknown function calls
		outputs["result"] = resource.NewStringProperty("health-check-success")
		outputs["status"] = resource.NewStringProperty("healthy")
		outputs["lastCheck"] = resource.NewStringProperty("2024-01-01T12:00:00Z")
	}

	return outputs, nil
}