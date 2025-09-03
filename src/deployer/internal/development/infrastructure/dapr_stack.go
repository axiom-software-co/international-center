package infrastructure

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type DaprStack struct {
	ctx           *pulumi.Context
	config        *config.Config
	networkName   string
	environment   string
}

type DaprDeployment struct {
	RedisContainer      *docker.Container
	DaprPlacementContainer *docker.Container
	DaprSentryContainer   *docker.Container
	DaprOperatorContainer *docker.Container
	RedisNetwork        *docker.Network
	DaprComponentsVolume *docker.Volume
	RedisDataVolume     *docker.Volume
}

func NewDaprStack(ctx *pulumi.Context, config *config.Config, networkName, environment string) *DaprStack {
	return &DaprStack{
		ctx:         ctx,
		config:      config,
		networkName: networkName,
		environment: environment,
	}
}

func (ds *DaprStack) Deploy(ctx context.Context) (*DaprDeployment, error) {
	deployment := &DaprDeployment{}

	var err error

	deployment.RedisNetwork, err = ds.createRedisNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis network: %w", err)
	}

	deployment.RedisDataVolume, err = ds.createRedisDataVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis data volume: %w", err)
	}

	deployment.DaprComponentsVolume, err = ds.createDaprComponentsVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create Dapr components volume: %w", err)
	}

	deployment.RedisContainer, err = ds.deployRedisContainer(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Redis container: %w", err)
	}

	deployment.DaprPlacementContainer, err = ds.deployDaprPlacementContainer(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Dapr placement container: %w", err)
	}

	deployment.DaprSentryContainer, err = ds.deployDaprSentryContainer(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Dapr sentry container: %w", err)
	}

	deployment.DaprOperatorContainer, err = ds.deployDaprOperatorContainer(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Dapr operator container: %w", err)
	}

	return deployment, nil
}

func (ds *DaprStack) createRedisNetwork() (*docker.Network, error) {
	network, err := docker.NewNetwork(ds.ctx, "redis-network", &docker.NetworkArgs{
		Name: pulumi.Sprintf("%s-redis-network", ds.environment),
		Driver: pulumi.String("bridge"),
		Options: pulumi.StringMap{
			"com.docker.network.driver.mtu": pulumi.String("1500"),
		},
		Labels: docker.NetworkLabelArray{
			&docker.NetworkLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ds.environment),
			},
			&docker.NetworkLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("redis"),
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

func (ds *DaprStack) createRedisDataVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(ds.ctx, "redis-data", &docker.VolumeArgs{
		Name: pulumi.Sprintf("%s-redis-data", ds.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ds.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("redis"),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("data-type"),
				Value: pulumi.String("persistent"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (ds *DaprStack) createDaprComponentsVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(ds.ctx, "dapr-components", &docker.VolumeArgs{
		Name: pulumi.Sprintf("%s-dapr-components", ds.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ds.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("dapr"),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("data-type"),
				Value: pulumi.String("configuration"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (ds *DaprStack) deployRedisContainer(deployment *DaprDeployment) (*docker.Container, error) {
	redisPort := ds.config.RequireInt("redis_port")
	redisPassword := ds.config.Get("redis_password")
	
	envVars := pulumi.StringArray{
		pulumi.String("REDIS_PASSWORD=" + redisPassword),
		pulumi.String("REDIS_PORT=" + fmt.Sprintf("%d", redisPort)),
		pulumi.String("REDIS_DATABASES=16"),
	}

	container, err := docker.NewContainer(ds.ctx, "redis-pubsub", &docker.ContainerArgs{
		Name:  pulumi.Sprintf("%s-redis-pubsub", ds.environment),
		Image: pulumi.String("redis:7-alpine"),
		Restart: pulumi.String("unless-stopped"),
		
		Command: pulumi.StringArray{
			pulumi.String("redis-server"),
			pulumi.String("--save"), pulumi.String("60"), pulumi.String("1"),
			pulumi.String("--loglevel"), pulumi.String("notice"),
			pulumi.String("--maxmemory"), pulumi.String("256mb"),
			pulumi.String("--maxmemory-policy"), pulumi.String("allkeys-lru"),
			pulumi.String("--notify-keyspace-events"), pulumi.String("Ex"),
		},
		
		Envs: envVars,
		
		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(6379),
				External: pulumi.Int(redisPort),
				Protocol: pulumi.String("tcp"),
			},
		},
		
		Mounts: docker.ContainerMountArray{
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.RedisDataVolume.Name,
				Target: pulumi.String("/data"),
			},
		},
		
		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: deployment.RedisNetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("redis"),
					pulumi.String("redis-pubsub"),
				},
			},
		},
		
		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests: pulumi.StringArray{
				pulumi.String("CMD-SHELL"),
				pulumi.String("redis-cli ping"),
			},
			Interval: pulumi.String("30s"),
			Timeout:  pulumi.String("10s"),
			Retries:  pulumi.Int(3),
		},
		
		Labels: docker.ContainerLabelArray{
			&docker.ContainerLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ds.environment),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("redis"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("service"),
				Value: pulumi.String("pubsub"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("dapr-component"),
				Value: pulumi.String("pubsub"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},
		
		LogDriver: pulumi.String("json-file"),
		LogOpts: pulumi.StringMap{
			"max-size": pulumi.String("10m"),
			"max-file": pulumi.String("3"),
		},
	})
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (ds *DaprStack) deployDaprPlacementContainer(deployment *DaprDeployment) (*docker.Container, error) {
	daprVersion := ds.config.Get("dapr_version")
	if daprVersion == "" {
		daprVersion = "1.12.0"
	}

	container, err := docker.NewContainer(ds.ctx, "dapr-placement", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-dapr-placement", ds.environment),
		Image:   pulumi.Sprintf("daprio/dapr:%s", daprVersion),
		Restart: pulumi.String("unless-stopped"),
		
		Command: pulumi.StringArray{
			pulumi.String("./placement"),
			pulumi.String("--port"), pulumi.String("50005"),
			pulumi.String("--log-level"), pulumi.String("info"),
			pulumi.String("--tls-enabled=false"),
		},
		
		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(50005),
				External: pulumi.Int(50005),
				Protocol: pulumi.String("tcp"),
			},
		},
		
		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: deployment.RedisNetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("dapr-placement"),
					pulumi.String("placement"),
				},
			},
		},
		
		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests: pulumi.StringArray{
				pulumi.String("CMD-SHELL"),
				pulumi.String("nc -z localhost 50005"),
			},
			Interval: pulumi.String("30s"),
			Timeout:  pulumi.String("10s"),
			Retries:  pulumi.Int(3),
		},
		
		Labels: docker.ContainerLabelArray{
			&docker.ContainerLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ds.environment),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("dapr"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("service"),
				Value: pulumi.String("placement"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (ds *DaprStack) deployDaprSentryContainer(deployment *DaprDeployment) (*docker.Container, error) {
	daprVersion := ds.config.Get("dapr_version")
	if daprVersion == "" {
		daprVersion = "1.12.0"
	}

	container, err := docker.NewContainer(ds.ctx, "dapr-sentry", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-dapr-sentry", ds.environment),
		Image:   pulumi.Sprintf("daprio/dapr:%s", daprVersion),
		Restart: pulumi.String("unless-stopped"),
		
		Command: pulumi.StringArray{
			pulumi.String("./sentry"),
			pulumi.String("--port"), pulumi.String("50001"),
			pulumi.String("--log-level"), pulumi.String("info"),
			pulumi.String("--trust-domain"), pulumi.String("localhost"),
		},
		
		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(50001),
				External: pulumi.Int(50001),
				Protocol: pulumi.String("tcp"),
			},
		},
		
		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: deployment.RedisNetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("dapr-sentry"),
					pulumi.String("sentry"),
				},
			},
		},
		
		Labels: docker.ContainerLabelArray{
			&docker.ContainerLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ds.environment),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("dapr"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("service"),
				Value: pulumi.String("sentry"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (ds *DaprStack) deployDaprOperatorContainer(deployment *DaprDeployment) (*docker.Container, error) {
	daprVersion := ds.config.Get("dapr_version")
	if daprVersion == "" {
		daprVersion = "1.12.0"
	}

	container, err := docker.NewContainer(ds.ctx, "dapr-operator", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-dapr-operator", ds.environment),
		Image:   pulumi.Sprintf("daprio/dapr:%s", daprVersion),
		Restart: pulumi.String("unless-stopped"),
		
		Command: pulumi.StringArray{
			pulumi.String("./operator"),
			pulumi.String("--log-level"), pulumi.String("info"),
		},
		
		Mounts: docker.ContainerMountArray{
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.DaprComponentsVolume.Name,
				Target: pulumi.String("/components"),
			},
			&docker.ContainerMountArgs{
				Type:   pulumi.String("bind"),
				Source: pulumi.String("/var/run/docker.sock"),
				Target: pulumi.String("/var/run/docker.sock"),
			},
		},
		
		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: deployment.RedisNetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("dapr-operator"),
					pulumi.String("operator"),
				},
			},
		},
		
		Labels: docker.ContainerLabelArray{
			&docker.ContainerLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ds.environment),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("dapr"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("service"),
				Value: pulumi.String("operator"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (ds *DaprStack) ValidateDeployment(ctx context.Context, deployment *DaprDeployment) error {
	if deployment.RedisContainer == nil {
		return fmt.Errorf("Redis container is not deployed")
	}

	if deployment.DaprPlacementContainer == nil {
		return fmt.Errorf("Dapr placement container is not deployed")
	}

	if deployment.DaprSentryContainer == nil {
		return fmt.Errorf("Dapr sentry container is not deployed")
	}

	if deployment.DaprOperatorContainer == nil {
		return fmt.Errorf("Dapr operator container is not deployed")
	}

	return nil
}

func (ds *DaprStack) GenerateDaprComponents(ctx context.Context, deployment *DaprDeployment) error {
	redisPort := ds.config.RequireInt("redis_port")
	redisPassword := ds.config.Get("redis_password")
	
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
    value: redis:%d
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
`, redisPort, redisPassword, ds.environment)

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
- identity-api
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

func (ds *DaprStack) GetRedisConnectionInfo() (string, int, string) {
	redisHost := "localhost"
	redisPort := ds.config.RequireInt("redis_port")
	redisPassword := ds.config.Get("redis_password")
	
	return redisHost, redisPort, redisPassword
}

func (ds *DaprStack) GetDaprEndpoints() map[string]string {
	return map[string]string{
		"http":      "http://localhost:3500",
		"grpc":      "localhost:50001",
		"placement": "localhost:50005",
		"sentry":    "localhost:50001",
	}
}