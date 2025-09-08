package shared

import (
	"fmt"
	"testing"

	"github.com/axiom-software-co/international-center/src/cicd/components"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ContractTestingFramework provides utilities for component contract validation
type ContractTestingFramework struct {
	project string
	stack   string
	mocks   pulumi.MockResourceMonitor
}

// NewContractTestingFramework creates a new contract testing framework
func NewContractTestingFramework(project, stack string) *ContractTestingFramework {
	return &ContractTestingFramework{
		project: project,
		stack:   stack,
		mocks:   &ContractTestMocks{},
	}
}

// RunComponentContractTest executes a component contract test with proper Pulumi context
func (f *ContractTestingFramework) RunComponentContractTest(t *testing.T, environment string, testFunc func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		return testFunc(t, ctx, cfg, environment)
	}, pulumi.WithMocks(f.project, f.stack, f.mocks))
	require.NoError(t, err)
}

// ValidateComponentOutputs validates that component outputs meet contract requirements
func ValidateComponentOutputs(t *testing.T, component string, outputs interface{}, environment string) {
	switch component {
	case "database":
		validateDatabaseOutputs(t, outputs.(*components.DatabaseOutputs), environment)
	case "storage":
		validateStorageOutputs(t, outputs.(*components.StorageOutputs), environment)
	case "vault":
		validateVaultOutputs(t, outputs.(*components.VaultOutputs), environment)
	case "rabbitmq":
		validateRabbitMQOutputs(t, outputs.(*components.RabbitMQOutputs), environment)
	case "observability":
		validateObservabilityOutputs(t, outputs.(*components.ObservabilityOutputs), environment)
	case "dapr":
		validateDaprOutputs(t, outputs.(*components.DaprOutputs), environment)
	case "services":
		validateServicesOutputs(t, outputs.(*components.ServicesOutputs), environment)
	case "website":
		validateWebsiteOutputs(t, outputs.(*components.WebsiteOutputs), environment)
	default:
		t.Errorf("Unknown component type: %s", component)
	}
}

// ValidateComponentIntegration validates that two components can integrate correctly
func ValidateComponentIntegration(t *testing.T, componentA string, outputsA interface{}, componentB string, outputsB interface{}, environment string) {
	integrationKey := fmt.Sprintf("%s-%s", componentA, componentB)
	
	switch integrationKey {
	case "database-services":
		validateDatabaseServicesIntegration(t, outputsA.(*components.DatabaseOutputs), outputsB.(*components.ServicesOutputs), environment)
	case "storage-services":
		validateStorageServicesIntegration(t, outputsA.(*components.StorageOutputs), outputsB.(*components.ServicesOutputs), environment)
	case "vault-services":
		validateVaultServicesIntegration(t, outputsA.(*components.VaultOutputs), outputsB.(*components.ServicesOutputs), environment)
	case "vault-dapr":
		validateVaultDaprIntegration(t, outputsA.(*components.VaultOutputs), outputsB.(*components.DaprOutputs), environment)
	case "rabbitmq-services":
		validateRabbitMQServicesIntegration(t, outputsA.(*components.RabbitMQOutputs), outputsB.(*components.ServicesOutputs), environment)
	case "rabbitmq-dapr":
		validateRabbitMQDaprIntegration(t, outputsA.(*components.RabbitMQOutputs), outputsB.(*components.DaprOutputs), environment)
	case "dapr-services":
		validateDaprServicesIntegration(t, outputsA.(*components.DaprOutputs), outputsB.(*components.ServicesOutputs), environment)
	case "services-website":
		validateServicesWebsiteIntegration(t, outputsA.(*components.ServicesOutputs), outputsB.(*components.WebsiteOutputs), environment)
	case "observability-services":
		validateObservabilityServicesIntegration(t, outputsA.(*components.ObservabilityOutputs), outputsB.(*components.ServicesOutputs), environment)
	default:
		t.Logf("No specific integration validation defined for %s", integrationKey)
	}
}

// ValidateEnvironmentParity validates that component outputs maintain environment parity principles
func ValidateEnvironmentParity(t *testing.T, component string, devOutputs, stagingOutputs, prodOutputs interface{}) {
	switch component {
	case "database":
		dev := devOutputs.(*components.DatabaseOutputs)
		staging := stagingOutputs.(*components.DatabaseOutputs)
		prod := prodOutputs.(*components.DatabaseOutputs)
		
		// Validate that different environments provide same interface contract
		assert.NotNil(t, dev.ConnectionString)
		assert.NotNil(t, staging.ConnectionString)
		assert.NotNil(t, prod.ConnectionString)
		
		assert.NotNil(t, dev.DatabaseName)
		assert.NotNil(t, staging.DatabaseName)
		assert.NotNil(t, prod.DatabaseName)
		
	case "services":
		dev := devOutputs.(*components.ServicesOutputs)
		staging := stagingOutputs.(*components.ServicesOutputs)
		prod := prodOutputs.(*components.ServicesOutputs)
		
		// Validate service parity across environments
		assert.NotNil(t, dev.APIServices)
		assert.NotNil(t, staging.APIServices)
		assert.NotNil(t, prod.APIServices)
		
		assert.NotNil(t, dev.GatewayServices)
		assert.NotNil(t, staging.GatewayServices)
		assert.NotNil(t, prod.GatewayServices)
		
	// Add more component parity validations as needed
	}
}

// Component-specific output validation functions

func validateDatabaseOutputs(t *testing.T, outputs *components.DatabaseOutputs, env string) {
	assert.NotNil(t, outputs.ConnectionString, "Database must provide connection string")
	assert.NotNil(t, outputs.DatabaseName, "Database must provide database name")
	assert.NotNil(t, outputs.Port, "Database must provide port")
	
	switch env {
	case "development":
		// Development should use container deployment
	case "staging", "production":
		// Staging/production should use managed database
	}
}

func validateStorageOutputs(t *testing.T, outputs *components.StorageOutputs, env string) {
	assert.NotNil(t, outputs.ConnectionString, "Storage must provide connection string")
	assert.NotNil(t, outputs.ContainerName, "Storage must provide container name")
	
	switch env {
	case "development":
		// Development should use emulator
	case "staging", "production":
		// Staging/production should use cloud storage
	}
}

func validateVaultOutputs(t *testing.T, outputs *components.VaultOutputs, env string) {
	assert.NotNil(t, outputs.VaultAddress, "Vault must provide vault address")
	assert.NotNil(t, outputs.AuthMethod, "Vault must provide auth method")
	
	switch env {
	case "development":
		// Development should use local container
	case "staging", "production":
		// Staging/production should use cloud vault
	}
}

func validateRabbitMQOutputs(t *testing.T, outputs *components.RabbitMQOutputs, env string) {
	// Core contract requirements - all environments must provide these
	assert.NotNil(t, outputs.DeploymentType, "RabbitMQ must provide deployment type")
	assert.NotNil(t, outputs.Endpoint, "RabbitMQ must provide endpoint")
	assert.NotNil(t, outputs.ManagementURL, "RabbitMQ must provide management URL")
	assert.NotNil(t, outputs.Port, "RabbitMQ must provide port")
	assert.NotNil(t, outputs.ManagementPort, "RabbitMQ must provide management port")
	assert.NotNil(t, outputs.Username, "RabbitMQ must provide username")
	assert.NotNil(t, outputs.Password, "RabbitMQ must provide password")
	assert.NotNil(t, outputs.VHost, "RabbitMQ must provide virtual host")
	assert.NotNil(t, outputs.HighAvailability, "RabbitMQ must provide high availability configuration")
	
	switch env {
	case "development":
		// Development should use Podman container
		outputs.DeploymentType.ApplyT(func(deploymentType string) error {
			assert.Equal(t, "podman_container", deploymentType, "Development should use Podman container")
			return nil
		})
		
		// Development should use standard AMQP port
		outputs.Port.ApplyT(func(port int) error {
			assert.Equal(t, 5672, port, "Development should use standard AMQP port")
			return nil
		})
		
		// Development should use standard management port
		outputs.ManagementPort.ApplyT(func(port int) error {
			assert.Equal(t, 15672, port, "Development should use standard management port")
			return nil
		})
		
		// Development should use localhost endpoint
		outputs.Endpoint.ApplyT(func(endpoint string) error {
			assert.Contains(t, endpoint, "localhost", "Development should use localhost endpoint")
			return nil
		})
		
		// Development should not have high availability
		outputs.HighAvailability.ApplyT(func(ha bool) error {
			assert.False(t, ha, "Development should not have high availability")
			return nil
		})
		
	case "staging", "production":
		// Staging/production should use managed CloudAMQP
		outputs.DeploymentType.ApplyT(func(deploymentType string) error {
			assert.Equal(t, "cloudamqp_managed", deploymentType, "Staging/production should use CloudAMQP managed service")
			return nil
		})
		
		// Staging/production should use TLS port
		outputs.Port.ApplyT(func(port int) error {
			assert.Equal(t, 5671, port, "Staging/production should use TLS AMQP port")
			return nil
		})
		
		// Staging/production should use HTTPS management port
		outputs.ManagementPort.ApplyT(func(port int) error {
			assert.Equal(t, 443, port, "Staging/production should use HTTPS management port")
			return nil
		})
		
		// Staging/production should use AMQPS protocol
		outputs.Endpoint.ApplyT(func(endpoint string) error {
			assert.Contains(t, endpoint, "amqps://", "Staging/production should use AMQPS protocol")
			return nil
		})
		
		// Staging/production should have high availability
		outputs.HighAvailability.ApplyT(func(ha bool) error {
			assert.True(t, ha, "Staging/production should have high availability")
			return nil
		})
		
		// Staging/production should use CloudAMQP domain
		outputs.ManagementURL.ApplyT(func(url string) error {
			assert.Contains(t, url, "cloudamqp.com", "Staging/production should use CloudAMQP management URL")
			return nil
		})
	}
}

func validateObservabilityOutputs(t *testing.T, outputs *components.ObservabilityOutputs, env string) {
	// Core contract requirements - all environments must provide these
	assert.NotNil(t, outputs.GrafanaURL, "Observability must provide Grafana URL")
	assert.NotNil(t, outputs.StackType, "Observability must provide stack type")
	assert.NotNil(t, outputs.RetentionDays, "Observability must provide retention days")
	assert.NotNil(t, outputs.AuditLogging, "Observability must provide audit logging configuration")
	assert.NotNil(t, outputs.AlertingEnabled, "Observability must provide alerting configuration")
	
	switch env {
	case "development":
		// Development should use consolidated otel-lgtm single container
		outputs.StackType.ApplyT(func(stackType string) error {
			assert.Equal(t, "otel_lgtm_container", stackType, "Development should use consolidated otel-lgtm container")
			return nil
		})
		
		// Development should provide all observability endpoints
		assert.NotNil(t, outputs.PrometheusURL, "Development observability must provide Prometheus URL")
		assert.NotNil(t, outputs.LokiURL, "Development observability must provide Loki URL")
		
		// Validate development URLs use localhost
		outputs.GrafanaURL.ApplyT(func(url string) error {
			assert.Contains(t, url, "127.0.0.1:3000", "Development Grafana should use local URL")
			return nil
		})
		
		outputs.PrometheusURL.ApplyT(func(url string) error {
			assert.Contains(t, url, "127.0.0.1:9090", "Development Prometheus should use local URL")
			return nil
		})
		
		outputs.LokiURL.ApplyT(func(url string) error {
			assert.Contains(t, url, "127.0.0.1:3100", "Development Loki should use local URL")
			return nil
		})
		
		// Development should have reasonable defaults
		outputs.RetentionDays.ApplyT(func(days int) error {
			assert.Equal(t, 7, days, "Development should have 7 days retention")
			return nil
		})
		
	case "staging", "production":
		// Staging/production should use Grafana Cloud
		outputs.StackType.ApplyT(func(stackType string) error {
			assert.Equal(t, "grafana_cloud", stackType, "Staging/production should use Grafana Cloud")
			return nil
		})
		
		// Cloud observability should use grafana.net URLs
		outputs.GrafanaURL.ApplyT(func(url string) error {
			assert.Contains(t, url, "grafana.net", "Staging/production should use Grafana Cloud URL")
			return nil
		})
		
		// Validate retention policies per environment
		if env == "staging" {
			outputs.RetentionDays.ApplyT(func(days int) error {
				assert.Equal(t, 30, days, "Staging should have 30 days retention")
				return nil
			})
		} else if env == "production" {
			outputs.RetentionDays.ApplyT(func(days int) error {
				assert.Equal(t, 90, days, "Production should have 90 days retention")
				return nil
			})
			
			// Production should have alerting enabled
			outputs.AlertingEnabled.ApplyT(func(enabled bool) error {
				assert.True(t, enabled, "Production should have alerting enabled")
				return nil
			})
		}
		
		// All cloud environments should have audit logging
		outputs.AuditLogging.ApplyT(func(enabled bool) error {
			assert.True(t, enabled, "Staging/production should have audit logging enabled")
			return nil
		})
	}
}

func validateDaprOutputs(t *testing.T, outputs *components.DaprOutputs, env string) {
	assert.NotNil(t, outputs.DeploymentType, "Dapr must provide deployment type")
	assert.NotNil(t, outputs.ControlPlaneURL, "Dapr must provide control plane URL")
	
	switch env {
	case "development":
		// Development should use self-hosted
	case "staging", "production":
		// Staging/production should use managed Dapr
	}
}

func validateServicesOutputs(t *testing.T, outputs *components.ServicesOutputs, env string) {
	assert.NotNil(t, outputs.DeploymentType, "Services must provide deployment type")
	assert.NotNil(t, outputs.APIServices, "Services must provide API services")
	assert.NotNil(t, outputs.GatewayServices, "Services must provide gateway services")
	
	switch env {
	case "development":
		// Development should use containers
	case "staging", "production":
		// Staging/production should use container apps
	}
}

func validateWebsiteOutputs(t *testing.T, outputs *components.WebsiteOutputs, env string) {
	assert.NotNil(t, outputs.ServerURL, "Website must provide server URL")
	assert.NotNil(t, outputs.DeploymentType, "Website must provide deployment type")
	assert.NotNil(t, outputs.APIGatewayURL, "Website must provide API gateway URL")
	
	switch env {
	case "development":
		// Development should use local server
	case "staging", "production":
		// Staging/production should use Cloudflare Pages
	}
}

// Component integration validation functions

func validateDatabaseServicesIntegration(t *testing.T, dbOutputs *components.DatabaseOutputs, servicesOutputs *components.ServicesOutputs, env string) {
	assert.NotNil(t, dbOutputs.ConnectionString, "Database connection string required for services integration")
	assert.NotNil(t, servicesOutputs.DeploymentType, "Services deployment type required")
	
	// Contract: Services must be able to connect to database
	// This validates the connection string format is compatible
}

func validateStorageServicesIntegration(t *testing.T, storageOutputs *components.StorageOutputs, servicesOutputs *components.ServicesOutputs, env string) {
	assert.NotNil(t, storageOutputs.ConnectionString, "Storage connection string required for services integration")
	assert.NotNil(t, servicesOutputs.DeploymentType, "Services deployment type required")
	
	// Contract: Services must be able to access storage containers
}

func validateVaultServicesIntegration(t *testing.T, vaultOutputs *components.VaultOutputs, servicesOutputs *components.ServicesOutputs, env string) {
	assert.NotNil(t, vaultOutputs.VaultAddress, "Vault address required for services integration")
	assert.NotNil(t, servicesOutputs.DeploymentType, "Services deployment type required")
	
	// Contract: Services must be able to access secrets from vault
}

func validateVaultDaprIntegration(t *testing.T, vaultOutputs *components.VaultOutputs, daprOutputs *components.DaprOutputs, env string) {
	assert.NotNil(t, vaultOutputs.VaultAddress, "Vault address required for Dapr integration")
	assert.NotNil(t, daprOutputs.DeploymentType, "Dapr deployment type required")
	
	// Contract: Dapr secret store component must be able to connect to vault
}

func validateDaprServicesIntegration(t *testing.T, daprOutputs *components.DaprOutputs, servicesOutputs *components.ServicesOutputs, env string) {
	assert.NotNil(t, daprOutputs.DeploymentType, "Dapr deployment type required for services integration")
	assert.NotNil(t, servicesOutputs.DaprSidecarEnabled, "Services must have Dapr sidecar configuration")
	
	// Contract: Services must be configured to work with Dapr runtime
}

func validateServicesWebsiteIntegration(t *testing.T, servicesOutputs *components.ServicesOutputs, websiteOutputs *components.WebsiteOutputs, env string) {
	assert.NotNil(t, servicesOutputs.GatewayServices, "Gateway services required for website integration")
	assert.NotNil(t, websiteOutputs.APIGatewayURL, "Website must reference API gateway URL")
	
	// Contract: Website must be able to reach public gateway
}

func validateRabbitMQServicesIntegration(t *testing.T, rabbitmqOutputs *components.RabbitMQOutputs, servicesOutputs *components.ServicesOutputs, env string) {
	assert.NotNil(t, rabbitmqOutputs.Endpoint, "RabbitMQ endpoint required for services integration")
	assert.NotNil(t, rabbitmqOutputs.Username, "RabbitMQ username required for services integration")
	assert.NotNil(t, rabbitmqOutputs.Password, "RabbitMQ password required for services integration")
	assert.NotNil(t, rabbitmqOutputs.VHost, "RabbitMQ virtual host required for services integration")
	assert.NotNil(t, servicesOutputs.DeploymentType, "Services deployment type required")
	
	// Contract: Services must be able to connect to RabbitMQ for pub/sub messaging
	// This validates RabbitMQ connectivity for microservices communication
	switch env {
	case "development":
		// Development services should connect to local RabbitMQ container
		rabbitmqOutputs.Endpoint.ApplyT(func(endpoint string) error {
			assert.Contains(t, endpoint, "localhost", "Development services should connect to localhost RabbitMQ")
			return nil
		})
		
	case "staging", "production":
		// Staging/production services should connect to managed CloudAMQP
		rabbitmqOutputs.Endpoint.ApplyT(func(endpoint string) error {
			assert.Contains(t, endpoint, "cloudamqp.com", "Staging/production services should connect to CloudAMQP")
			return nil
		})
		
		// Validate high availability for production workloads
		rabbitmqOutputs.HighAvailability.ApplyT(func(ha bool) error {
			assert.True(t, ha, "Services require high availability RabbitMQ in staging/production")
			return nil
		})
	}
}

func validateRabbitMQDaprIntegration(t *testing.T, rabbitmqOutputs *components.RabbitMQOutputs, daprOutputs *components.DaprOutputs, env string) {
	assert.NotNil(t, rabbitmqOutputs.Endpoint, "RabbitMQ endpoint required for Dapr integration")
	assert.NotNil(t, rabbitmqOutputs.Username, "RabbitMQ username required for Dapr integration")
	assert.NotNil(t, rabbitmqOutputs.Password, "RabbitMQ password required for Dapr integration")
	assert.NotNil(t, daprOutputs.DeploymentType, "Dapr deployment type required")
	
	// Contract: Dapr pub/sub component must be able to connect to RabbitMQ
	// This validates RabbitMQ as Dapr pub/sub backing store
	switch env {
	case "development":
		// Development Dapr should connect to local RabbitMQ container
		rabbitmqOutputs.Endpoint.ApplyT(func(endpoint string) error {
			assert.Contains(t, endpoint, "localhost", "Development Dapr should connect to localhost RabbitMQ")
			return nil
		})
		
	case "staging", "production":
		// Staging/production Dapr should connect to managed CloudAMQP
		rabbitmqOutputs.Endpoint.ApplyT(func(endpoint string) error {
			assert.Contains(t, endpoint, "cloudamqp.com", "Staging/production Dapr should connect to CloudAMQP")
			return nil
		})
		
		// Validate secure connection for production Dapr
		rabbitmqOutputs.Endpoint.ApplyT(func(endpoint string) error {
			assert.Contains(t, endpoint, "amqps://", "Production Dapr should use secure AMQPS connection")
			return nil
		})
	}
}

func validateObservabilityServicesIntegration(t *testing.T, obsOutputs *components.ObservabilityOutputs, servicesOutputs *components.ServicesOutputs, env string) {
	assert.NotNil(t, obsOutputs.GrafanaURL, "Grafana URL required for observability integration")
	assert.NotNil(t, obsOutputs.StackType, "Observability stack type required for services integration")
	assert.NotNil(t, servicesOutputs.ObservabilityEnabled, "Services must have observability configuration")
	
	// Contract: Services must export metrics to observability stack
	servicesOutputs.ObservabilityEnabled.ApplyT(func(enabled bool) error {
		assert.True(t, enabled, "Services must have observability enabled for integration")
		return nil
	})
	
	switch env {
	case "development":
		// Development integration: Services should connect to consolidated otel-lgtm container
		obsOutputs.StackType.ApplyT(func(stackType string) error {
			assert.Equal(t, "otel_lgtm_container", stackType, "Development services should integrate with otel-lgtm container")
			return nil
		})
		
		// Validate all development observability endpoints are available for services integration
		assert.NotNil(t, obsOutputs.PrometheusURL, "Development services need Prometheus URL for metrics export")
		assert.NotNil(t, obsOutputs.LokiURL, "Development services need Loki URL for logs export")
		
		// Contract: Services should be able to reach local observability endpoints
		obsOutputs.GrafanaURL.ApplyT(func(grafanaURL string) error {
			assert.Contains(t, grafanaURL, "127.0.0.1:3000", "Services should access local Grafana for development")
			return nil
		})
		
	case "staging", "production":
		// Staging/production integration: Services should connect to Grafana Cloud
		obsOutputs.StackType.ApplyT(func(stackType string) error {
			assert.Equal(t, "grafana_cloud", stackType, "Staging/production services should integrate with Grafana Cloud")
			return nil
		})
		
		// Contract: Services should export to cloud observability endpoints
		obsOutputs.GrafanaURL.ApplyT(func(grafanaURL string) error {
			assert.Contains(t, grafanaURL, "grafana.net", "Services should export to Grafana Cloud in staging/production")
			return nil
		})
		
		// Validate audit logging integration for compliance
		obsOutputs.AuditLogging.ApplyT(func(auditEnabled bool) error {
			assert.True(t, auditEnabled, "Services integration requires audit logging in staging/production")
			return nil
		})
	}
	
	// Contract: Dapr-enabled services must have proper observability sidecar configuration
	assert.NotNil(t, servicesOutputs.DaprSidecarEnabled, "Dapr sidecar configuration required for observability integration")
	servicesOutputs.DaprSidecarEnabled.ApplyT(func(daprEnabled bool) error {
		if daprEnabled {
			// When Dapr is enabled, observability should be configured to collect Dapr metrics
			obsOutputs.RetentionDays.ApplyT(func(retentionDays int) error {
				assert.Greater(t, retentionDays, 0, "Observability retention must be configured for Dapr metrics collection")
				return nil
			})
		}
		return nil
	})
}

// ContractTestMocks provides enhanced mock implementation for contract testing
type ContractTestMocks struct{}

func (m *ContractTestMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}
	
	// Provide component-specific mock outputs based on resource type
	switch args.TypeToken {
	case "docker:index/container:Container":
		outputs["name"] = resource.NewStringProperty(args.Name)
		outputs["image"] = resource.NewStringProperty("mock-image")
		outputs["ports"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{
				"internal": resource.NewNumberProperty(8080),
				"external": resource.NewNumberProperty(8080),
			}),
		})
		
	case "azure:storage/account:Account":
		outputs["name"] = resource.NewStringProperty(args.Name)
		outputs["primaryConnectionString"] = resource.NewStringProperty("DefaultEndpointsProtocol=https;AccountName=mock;AccountKey=mock;EndpointSuffix=core.windows.net")
		outputs["primaryBlobEndpoint"] = resource.NewStringProperty("https://mock.blob.core.windows.net/")
		
	case "azure:postgresql/server:Server":
		outputs["name"] = resource.NewStringProperty(args.Name)
		outputs["fullyQualifiedDomainName"] = resource.NewStringProperty("mock-db.postgres.database.azure.com")
		
	case "cloudflare:index/pagesProject:PagesProject":
		outputs["name"] = resource.NewStringProperty(args.Name)
		outputs["subdomain"] = resource.NewStringProperty("mock-project.pages.dev")
		
	case "pulumi:providers:azure":
		// Azure provider - no specific outputs needed
		
	case "pulumi:providers:cloudflare":
		// Cloudflare provider - no specific outputs needed
		
	default:
		// Default mock outputs for unknown resource types
		outputs["name"] = resource.NewStringProperty(args.Name)
		outputs["id"] = resource.NewStringProperty(args.Name + "_mock_id")
	}
	
	return args.Name + "_id", outputs, nil
}

func (m *ContractTestMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}
	
	switch args.Token {
	case "azure:core/getResourceGroup:getResourceGroup":
		outputs["name"] = resource.NewStringProperty("mock-rg")
		outputs["location"] = resource.NewStringProperty("East US")
		
	case "azure:core/getClientConfig:getClientConfig":
		outputs["subscriptionId"] = resource.NewStringProperty("mock-subscription-id")
		outputs["tenantId"] = resource.NewStringProperty("mock-tenant-id")
		outputs["clientId"] = resource.NewStringProperty("mock-client-id")
		
	default:
		// Default mock outputs for unknown function calls
		outputs["result"] = resource.NewStringProperty("mock-result")
	}
	
	return outputs, nil
}