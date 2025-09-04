package infrastructure

import (
	"context"
	"fmt"

	shared "github.com/axiom-software-co/international-center/src/deployer/shared/infrastructure"
	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	sharedconfig "github.com/axiom-software-co/international-center/src/deployer/shared/config"
)

type IntegrationStack struct {
	pulumi.ComponentResource
	ctx         *pulumi.Context
	config      *config.Config
	configManager *sharedconfig.ConfigManager
	environment string
	projectRoot string
	
	// Stack components
	daprStack      *DaprStack
	serviceStack   *ServiceStack
	databaseStack  shared.DatabaseStack
	storageStack   *StorageStack
	vaultStack     *VaultStack
	observabilityStack *ObservabilityStack
	
	// Outputs
	MainNetworkID         pulumi.StringOutput `pulumi:"mainNetworkId"`
	DatabaseEndpoint      pulumi.StringOutput `pulumi:"databaseEndpoint"`
	VaultEndpoint         pulumi.StringOutput `pulumi:"vaultEndpoint"`
	DaprHTTPEndpoint      pulumi.StringOutput `pulumi:"daprHTTPEndpoint"`
	PublicGatewayEndpoint pulumi.StringOutput `pulumi:"publicGatewayEndpoint"`
	AdminGatewayEndpoint  pulumi.StringOutput `pulumi:"adminGatewayEndpoint"`
}

type IntegratedDeployment struct {
	pulumi.ComponentResource
	// Infrastructure
	DatabaseDeployment     shared.DatabaseDeployment
	StorageDeployment      *StorageDeployment
	VaultDeployment        *VaultDeployment
	ObservabilityDeployment *ObservabilityDeployment
	
	// Dapr Control Plane
	DaprDeployment         *DaprDeployment
	
	// Containerized Services
	ServiceDeployment      *ServiceDeployment
	
	// Main Network
	MainNetwork            *docker.Network
	
	// Integration Test Dapr Sidecar
	IntegrationTestSidecar *docker.Container
	
	// Outputs
	MainNetworkID         pulumi.StringOutput `pulumi:"mainNetworkId"`
	DatabaseEndpoint      pulumi.StringOutput `pulumi:"databaseEndpoint"`
	VaultEndpoint         pulumi.StringOutput `pulumi:"vaultEndpoint"`
	DaprHTTPEndpoint      pulumi.StringOutput `pulumi:"daprHTTPEndpoint"`
	PublicGatewayEndpoint pulumi.StringOutput `pulumi:"publicGatewayEndpoint"`
	AdminGatewayEndpoint  pulumi.StringOutput `pulumi:"adminGatewayEndpoint"`
}

func NewIntegrationStack(ctx *pulumi.Context, config *config.Config, environment, projectRoot string) *IntegrationStack {
	// Create ConfigManager for centralized configuration
	configManager, err := sharedconfig.NewConfigManager(ctx)
	if err != nil {
		ctx.Log.Warn(fmt.Sprintf("Failed to create ConfigManager, using legacy configuration: %v", err), nil)
		configManager = nil
	}
	
	component := &IntegrationStack{
		ctx:           ctx,
		config:        config,
		configManager: configManager,
		environment:   environment,
		projectRoot:   projectRoot,
	}
	
	err = ctx.RegisterComponentResource("international-center:integration:DevelopmentStack",
		fmt.Sprintf("%s-integration-stack", environment), component)
	if err != nil {
		return nil
	}
	
	return component
}

func (is *IntegrationStack) Deploy(ctx context.Context) (*IntegratedDeployment, error) {
	deployment := &IntegratedDeployment{}
	
	var err error
	
	// 1. Create main container network
	deployment.MainNetwork, err = is.createMainNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to create main network: %w", err)
	}
	
	// 2. Deploy infrastructure services (database, storage, vault, observability)
	deployment.DatabaseDeployment, err = is.deployDatabase(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy database: %w", err)
	}
	
	deployment.StorageDeployment, err = is.deployStorage(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy storage: %w", err)
	}
	
	deployment.VaultDeployment, err = is.deployVault(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy vault: %w", err)
	}
	
	deployment.ObservabilityDeployment, err = is.deployObservability(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy observability: %w", err)
	}
	
	// 3. Deploy Dapr control plane
	is.daprStack = NewDaprStack(is.ctx, is.config, "dev-network", is.environment)
	daprDeployment, err := is.daprStack.Deploy(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Dapr stack: %w", err)
	}
	deployment.DaprDeployment = daprDeployment.(*DaprDeployment)
	
	// 4. Deploy containerized services with Dapr sidecars
	is.serviceStack = NewServiceStack(is.ctx, is.config, deployment.DaprDeployment, "dev-network", is.environment, is.projectRoot)
	serviceDeployment, err := is.serviceStack.Deploy(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy service stack: %w", err)
	}
	deployment.ServiceDeployment = serviceDeployment.(*ServiceDeployment)
	
	// 5. Deploy integration test Dapr sidecar
	deployment.IntegrationTestSidecar, err = is.deployIntegrationTestSidecar(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy integration test sidecar: %w", err)
	}
	
	return deployment, nil
}

func (is *IntegrationStack) createMainNetwork() (*docker.Network, error) {
	network, err := docker.NewNetwork(is.ctx, "main-network", &docker.NetworkArgs{
		Name:   pulumi.Sprintf("%s-main-network", is.environment),
		Driver: pulumi.String("bridge"),
		IpamConfigs: docker.NetworkIpamConfigArray{
			&docker.NetworkIpamConfigArgs{
				Subnet:  pulumi.String("172.18.0.0/16"),
				Gateway: pulumi.String("172.18.0.1"),
			},
		},
		Options: pulumi.StringMap{
			"com.docker.network.bridge.name": pulumi.String("br-main"),
			"com.docker.network.driver.mtu":  pulumi.String("1500"),
		},
		Labels: docker.NetworkLabelArray{
			&docker.NetworkLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(is.environment),
			},
			&docker.NetworkLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("main-infrastructure"),
			},
			&docker.NetworkLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	
	return network, nil
}

func (is *IntegrationStack) deployDatabase(deployment *IntegratedDeployment) (shared.DatabaseDeployment, error) {
	is.databaseStack = NewDatabaseStack(is.ctx, is.config, "dev-network", is.environment)
	return is.databaseStack.Deploy(context.Background())
}

func (is *IntegrationStack) deployStorage(deployment *IntegratedDeployment) (*StorageDeployment, error) {
	is.storageStack = NewStorageStack(is.ctx, is.config, "dev-network", is.environment)
	storageDeployment, err := is.storageStack.Deploy(context.Background())
	if err != nil {
		return nil, err
	}
	return storageDeployment.(*StorageDeployment), nil
}

func (is *IntegrationStack) deployVault(deployment *IntegratedDeployment) (*VaultDeployment, error) {
	is.vaultStack = NewVaultStack(is.ctx, is.config, "dev-network", is.environment)
	vaultDeployment, err := is.vaultStack.Deploy(context.Background())
	if err != nil {
		return nil, err
	}
	return vaultDeployment.(*VaultDeployment), nil
}

func (is *IntegrationStack) deployObservability(deployment *IntegratedDeployment) (*ObservabilityDeployment, error) {
	is.observabilityStack = NewObservabilityStack(is.ctx, is.config, "dev-network", is.environment)
	observabilityDeployment, err := is.observabilityStack.Deploy(context.Background())
	if err != nil {
		return nil, err
	}
	return observabilityDeployment.(*ObservabilityDeployment), nil
}

func (is *IntegrationStack) deployIntegrationTestSidecar(deployment *IntegratedDeployment) (*docker.Container, error) {
	daprVersion := is.config.Get("dapr_version")
	if daprVersion == "" {
		daprVersion = "1.12.0"
	}
	
	container, err := docker.NewContainer(is.ctx, "integration-test-dapr-sidecar", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-integration-test-dapr-sidecar", is.environment),
		Image:   pulumi.Sprintf("daprio/daprd:%s", daprVersion),
		Restart: pulumi.String("unless-stopped"),
		
		Command: pulumi.StringArray{
			pulumi.String("./daprd"),
			pulumi.String("--app-id=integration-test"),
			pulumi.String("--dapr-http-port=3500"),
			pulumi.String("--dapr-grpc-port=50001"),
			pulumi.String("--placement-host-address=dapr-placement:50005"),
			pulumi.String("--components-path=/components"),
			pulumi.String("--log-level=info"),
			pulumi.String("--app-ssl=false"),
		},
		
		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(3500),
				External: pulumi.Int(3500),
				Protocol: pulumi.String("tcp"),
			},
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(50001),
				External: pulumi.Int(50001),
				Protocol: pulumi.String("tcp"),
			},
		},
		
		Mounts: docker.ContainerMountArray{
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.DaprDeployment.DaprComponentsVolume.Name,
				Target: pulumi.String("/components"),
			},
		},
		
		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: deployment.ServiceDeployment.ServiceNetwork.Name,
			},
			&docker.ContainerNetworksAdvancedArgs{
				Name: deployment.DaprDeployment.RedisNetwork.Name,
			},
		},
		
		Labels: docker.ContainerLabelArray{
			&docker.ContainerLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(is.environment),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("integration-test"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{
		deployment.DaprDeployment.DaprPlacementContainer,
		deployment.DaprDeployment.RedisContainer,
	}))
	if err != nil {
		return nil, err
	}
	
	return container, nil
}

func (is *IntegrationStack) ValidateDeployment(ctx context.Context, deployment *IntegratedDeployment) error {
	// Validate infrastructure
	if deployment.DatabaseDeployment == nil {
		return fmt.Errorf("database deployment is missing")
	}
	
	if deployment.StorageDeployment == nil {
		return fmt.Errorf("storage deployment is missing")
	}
	
	if deployment.VaultDeployment == nil {
		return fmt.Errorf("vault deployment is missing")
	}
	
	// Validate Dapr control plane
	if deployment.DaprDeployment == nil {
		return fmt.Errorf("Dapr deployment is missing")
	}
	
	if err := is.daprStack.ValidateDeployment(ctx, deployment.DaprDeployment); err != nil {
		return fmt.Errorf("Dapr deployment validation failed: %w", err)
	}
	
	// Validate containerized services
	if deployment.ServiceDeployment == nil {
		return fmt.Errorf("service deployment is missing")
	}
	
	if err := is.serviceStack.ValidateDeployment(ctx, deployment.ServiceDeployment); err != nil {
		return fmt.Errorf("service deployment validation failed: %w", err)
	}
	
	// Validate integration test setup
	if deployment.IntegrationTestSidecar == nil {
		return fmt.Errorf("integration test sidecar is missing")
	}
	
	return nil
}

func (is *IntegrationStack) GenerateComponentsConfiguration(ctx context.Context, deployment *IntegratedDeployment) error {
	// Generate comprehensive Dapr components configuration for containerized environment
	redisPassword := is.config.Get("redis_password")
	
	pubsubComponentYAML := fmt.Sprintf(`
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: pubsub
  namespace: default
spec:
  type: pubsub.redis
  version: v1
  metadata:
  - name: redisHost
    value: redis:6379
  - name: redisPassword
    value: "%s"
  - name: enableTLS
    value: false
  - name: maxRetries
    value: "3"
  - name: maxRetryBackoff
    value: "2s"
scopes:
- content-api
- services-api
- public-gateway
- admin-gateway
- integration-test
`, redisPassword)

	stateStoreComponentYAML := fmt.Sprintf(`
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: statestore
  namespace: default
spec:
  type: state.redis
  version: v1
  metadata:
  - name: redisHost
    value: redis:6379
  - name: redisPassword
    value: "%s"
  - name: enableTLS
    value: false
  - name: keyPrefix
    value: "%s"
scopes:
- content-api
- services-api
- public-gateway
- admin-gateway
- integration-test
`, redisPassword, is.environment)

	secretStoreComponentYAML := `
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: secretstore
  namespace: default
spec:
  type: secretstores.hashicorp.vault
  version: v1
  metadata:
  - name: vaultAddr
    value: "http://vault:8200"
  - name: skipVerify
    value: true
  - name: vaultToken
    value: "dev-root-token"
scopes:
- content-api
- services-api
`

	blobStorageComponentYAML := `
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: blob-storage
  namespace: default
spec:
  type: bindings.azure.blobstorage
  version: v1
  metadata:
  - name: accountName
    value: "devstoreaccount1"
  - name: accountKey
    value: "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="
  - name: containerName
    value: "content"
  - name: endpoint
    value: "http://azurite:10000/devstoreaccount1"
scopes:
- content-api
`

	_ = pubsubComponentYAML
	_ = stateStoreComponentYAML
	_ = secretStoreComponentYAML
	_ = blobStorageComponentYAML

	return nil
}

func (is *IntegrationStack) GetServiceEndpoints() map[string]string {
	return map[string]string{
		"content-api":    "http://localhost:8080",
		"services-api":   "http://localhost:8081",
		"public-gateway": "http://localhost:8082",
		"admin-gateway":  "http://localhost:8083",
		"dapr-http":      "http://localhost:3500",
		"dapr-grpc":      "localhost:50001",
		"dapr-placement": "localhost:50005",
		"database":       "postgres://dev_user:dev_password@localhost:5432/international_center_dev?sslmode=disable",
		"redis":          "localhost:6379",
		"vault":          "http://localhost:8200",
		"azurite":        "http://localhost:10000",
		"grafana":        "http://localhost:3000",
		"loki":           "http://localhost:3100",
		"prometheus":     "http://localhost:9090",
	}
}

func (is *IntegrationStack) GetContainerizedServiceInfo() map[string]ContainerizedServiceInfo {
	return map[string]ContainerizedServiceInfo{
		"content-api": {
			ContainerName: fmt.Sprintf("%s-content-api", is.environment),
			DaprSidecarName: fmt.Sprintf("%s-content-api-dapr-sidecar", is.environment),
			AppPort: 8080,
			DaprHTTPPort: 3501,
			DaprGRPCPort: 50002,
			HealthEndpoint: "/health",
		},
		"services-api": {
			ContainerName: fmt.Sprintf("%s-services-api", is.environment),
			DaprSidecarName: fmt.Sprintf("%s-services-api-dapr-sidecar", is.environment),
			AppPort: 8081,
			DaprHTTPPort: 3502,
			DaprGRPCPort: 50003,
			HealthEndpoint: "/health",
		},
		"public-gateway": {
			ContainerName: fmt.Sprintf("%s-public-gateway", is.environment),
			DaprSidecarName: fmt.Sprintf("%s-public-gateway-dapr-sidecar", is.environment),
			AppPort: 8082,
			DaprHTTPPort: 3503,
			DaprGRPCPort: 50004,
			HealthEndpoint: "/health",
		},
		"admin-gateway": {
			ContainerName: fmt.Sprintf("%s-admin-gateway", is.environment),
			DaprSidecarName: fmt.Sprintf("%s-admin-gateway-dapr-sidecar", is.environment),
			AppPort: 8083,
			DaprHTTPPort: 3504,
			DaprGRPCPort: 50006,
			HealthEndpoint: "/health",
		},
	}
}

type ContainerizedServiceInfo struct {
	ContainerName    string
	DaprSidecarName  string
	AppPort          int
	DaprHTTPPort     int
	DaprGRPCPort     int
	HealthEndpoint   string
}

