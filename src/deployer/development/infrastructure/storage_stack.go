package infrastructure

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	sharedconfig "github.com/axiom-software-co/international-center/src/deployer/shared/config"
	sharedinfra "github.com/axiom-software-co/international-center/src/deployer/shared/infrastructure"
)

type StorageStack struct {
	pulumi.ComponentResource
	ctx           *pulumi.Context
	config        *config.Config
	configManager *sharedconfig.ConfigManager
	networkName   string
	environment   string
	
	// Outputs
	BlobEndpoint        pulumi.StringOutput `pulumi:"blobEndpoint"`
	ConnectionString    pulumi.StringOutput `pulumi:"connectionString"`
	StorageNetworkID    pulumi.StringOutput `pulumi:"storageNetworkId"`
	AzuriteContainerID  pulumi.StringOutput `pulumi:"azuriteContainerId"`
}

type StorageDeployment struct {
	pulumi.ComponentResource
	AzuriteContainer    *docker.Container
	StorageNetwork      *docker.Network
	AzuriteDataVolume   *docker.Volume
	AzuriteConfigVolume *docker.Volume
	
	// Outputs
	BlobEndpoint      pulumi.StringOutput `pulumi:"blobEndpoint"`
	QueueEndpoint     pulumi.StringOutput `pulumi:"queueEndpoint"`
	TableEndpoint     pulumi.StringOutput `pulumi:"tableEndpoint"`
	ConnectionString  pulumi.StringOutput `pulumi:"connectionString"`
}

// Implement the shared StorageDeployment interface
func (sd *StorageDeployment) GetPrimaryStorageEndpoint() pulumi.StringOutput {
	return sd.BlobEndpoint
}

func (sd *StorageDeployment) GetBackupStorageEndpoint() pulumi.StringOutput {
	// In development, we don't have backup storage, return the primary endpoint
	return sd.BlobEndpoint
}

func (sd *StorageDeployment) GetConnectionString() pulumi.StringOutput {
	return sd.ConnectionString
}

func (sd *StorageDeployment) GetContainerEndpoint(name string) string {
	return "http://localhost:10000/devstoreaccount1/" + name
}

func (sd *StorageDeployment) GetQueueEndpoint(name string) string {
	return "http://localhost:10001/devstoreaccount1/" + name
}

func NewStorageStack(ctx *pulumi.Context, config *config.Config, networkName, environment string) *StorageStack {
	// Create ConfigManager for centralized configuration
	configManager, err := sharedconfig.NewConfigManager(ctx)
	if err != nil {
		ctx.Log.Warn(fmt.Sprintf("Failed to create ConfigManager, using legacy configuration: %v", err), nil)
		configManager = nil
	}
	
	component := &StorageStack{
		ctx:           ctx,
		config:        config,
		configManager: configManager,
		networkName:   networkName,
		environment:   environment,
	}
	
	err = ctx.RegisterComponentResource("international-center:storage:DevelopmentStack", 
		fmt.Sprintf("%s-storage-stack", environment), component)
	if err != nil {
		ctx.Log.Error(fmt.Sprintf("Failed to register StorageStack component: %v", err), nil)
		return nil
	}
	
	return component
}

func (ss *StorageStack) Deploy(ctx context.Context) (sharedinfra.StorageDeployment, error) {
	deployment := &StorageDeployment{}
	
	// Register the deployment as a child ComponentResource
	err := ss.ctx.RegisterComponentResource("international-center:storage:DevelopmentDeployment",
		fmt.Sprintf("%s-storage-deployment", ss.environment), deployment, pulumi.Parent(ss))
	if err != nil {
		return nil, fmt.Errorf("failed to register StorageDeployment component: %w", err)
	}

	deployment.StorageNetwork, err = ss.createStorageNetworkWithParent(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage network: %w", err)
	}

	deployment.AzuriteDataVolume, err = ss.createAzuriteDataVolumeWithParent(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azurite data volume: %w", err)
	}

	deployment.AzuriteConfigVolume, err = ss.createAzuriteConfigVolumeWithParent(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azurite config volume: %w", err)
	}

	deployment.AzuriteContainer, err = ss.deployAzuriteContainerWithParent(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Azurite container: %w", err)
	}
	
	// Set outputs
	deployment.BlobEndpoint = deployment.AzuriteContainer.Ports.Index(pulumi.Int(0)).External().ApplyT(func(port *int) string {
		if port != nil {
			return fmt.Sprintf("http://localhost:%d/devstoreaccount1", *port)
		}
		return "http://localhost:10000/devstoreaccount1"
	}).(pulumi.StringOutput)
		
	deployment.QueueEndpoint = deployment.AzuriteContainer.Ports.Index(pulumi.Int(1)).External().ApplyT(func(port *int) string {
		if port != nil {
			return fmt.Sprintf("http://localhost:%d/devstoreaccount1", *port)
		}
		return "http://localhost:10001/devstoreaccount1"
	}).(pulumi.StringOutput)
		
	deployment.TableEndpoint = deployment.AzuriteContainer.Ports.Index(pulumi.Int(2)).External().ApplyT(func(port *int) string {
		if port != nil {
			return fmt.Sprintf("http://localhost:%d/devstoreaccount1", *port)
		}
		return "http://localhost:10002/devstoreaccount1"
	}).(pulumi.StringOutput)
		
	deployment.ConnectionString = pulumi.All(
		deployment.AzuriteContainer.Ports.Index(pulumi.Int(0)).External(),
		deployment.AzuriteContainer.Ports.Index(pulumi.Int(1)).External(),
		deployment.AzuriteContainer.Ports.Index(pulumi.Int(2)).External(),
	).ApplyT(func(args []interface{}) string {
		var blobPort, queuePort, tablePort int
		if args[0] != nil && args[0].(*int) != nil {
			blobPort = *args[0].(*int)
		} else {
			blobPort = 10000
		}
		if args[1] != nil && args[1].(*int) != nil {
			queuePort = *args[1].(*int)
		} else {
			queuePort = 10001
		}
		if args[2] != nil && args[2].(*int) != nil {
			tablePort = *args[2].(*int)
		} else {
			tablePort = 10002
		}
		return fmt.Sprintf("DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://localhost:%d/devstoreaccount1;QueueEndpoint=http://localhost:%d/devstoreaccount1;TableEndpoint=http://localhost:%d/devstoreaccount1;",
			blobPort, queuePort, tablePort)
	}).(pulumi.StringOutput)

	return deployment, nil
}

func (ss *StorageStack) createStorageNetwork() (*docker.Network, error) {
	network, err := docker.NewNetwork(ss.ctx, "storage-network", &docker.NetworkArgs{
		Name:   pulumi.Sprintf("%s-storage-network", ss.environment),
		Driver: pulumi.String("bridge"),
		Options: pulumi.StringMap{
			"com.docker.network.driver.mtu": pulumi.String("1500"),
		},
		Labels: docker.NetworkLabelArray{
			&docker.NetworkLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ss.environment),
			},
			&docker.NetworkLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("storage"),
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

func (ss *StorageStack) createAzuriteDataVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(ss.ctx, "azurite-data", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-azurite-data", ss.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ss.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("azurite"),
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

func (ss *StorageStack) createAzuriteConfigVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(ss.ctx, "azurite-config", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-azurite-config", ss.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ss.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("azurite"),
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

func (ss *StorageStack) deployAzuriteContainer(deployment *StorageDeployment) (*docker.Container, error) {
	var azuritePort int
	
	if ss.configManager == nil {
		return nil, fmt.Errorf("configManager is required for storage deployment")
	}
	
	storageConfig := ss.configManager.GetStorageConfig()
	var err error
	azuritePort, err = strconv.Atoi(storageConfig.AzuritePort)
	if err != nil {
		return nil, fmt.Errorf("invalid AZURITE_PORT from config: %w", err)
	}
	
	azuriteBlobPort := azuritePort     // 10000
	azuriteQueuePort := azuritePort + 1 // 10001
	azuriteTablePort := azuritePort + 2 // 10002

	container, err := docker.NewContainer(ss.ctx, "azurite", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-azurite", ss.environment),
		Image:   pulumi.String("mcr.microsoft.com/azure-storage/azurite:latest"),
		Restart: pulumi.String("unless-stopped"),

		Command: pulumi.StringArray{
			pulumi.String("azurite"),
			pulumi.String("--blobHost"), pulumi.String("0.0.0.0"),
			pulumi.String("--blobPort"), pulumi.String("10000"),
			pulumi.String("--queueHost"), pulumi.String("0.0.0.0"),
			pulumi.String("--queuePort"), pulumi.String("10001"),
			pulumi.String("--tableHost"), pulumi.String("0.0.0.0"),
			pulumi.String("--tablePort"), pulumi.String("10002"),
			pulumi.String("--location"), pulumi.String("/workspace"),
			pulumi.String("--debug"), pulumi.String("/workspace/debug.log"),
			pulumi.String("--loose"),
			pulumi.String("--skipApiVersionCheck"),
		},

		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(10000),
				External: pulumi.Int(azuriteBlobPort),
				Protocol: pulumi.String("tcp"),
			},
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(10001),
				External: pulumi.Int(azuriteQueuePort),
				Protocol: pulumi.String("tcp"),
			},
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(10002),
				External: pulumi.Int(azuriteTablePort),
				Protocol: pulumi.String("tcp"),
			},
		},

		Mounts: docker.ContainerMountArray{
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.AzuriteDataVolume.Name,
				Target: pulumi.String("/workspace"),
			},
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.AzuriteConfigVolume.Name,
				Target: pulumi.String("/config"),
			},
		},

		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: deployment.StorageNetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("azurite"),
					pulumi.String("storage"),
					pulumi.String("blob-storage"),
				},
			},
		},

		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests: pulumi.StringArray{
				pulumi.String("CMD-SHELL"),
				pulumi.String("curl -f http://localhost:10000/devstoreaccount1 || exit 1"),
			},
			Interval: pulumi.String("30s"),
			Timeout:  pulumi.String("10s"),
			Retries:  pulumi.Int(3),
			StartPeriod: pulumi.String("30s"),
		},

		Labels: docker.ContainerLabelArray{
			&docker.ContainerLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ss.environment),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("azurite"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("service"),
				Value: pulumi.String("storage"),
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

func (ss *StorageStack) CreateStorageContainers(ctx context.Context, deployment sharedinfra.StorageDeployment) error {
	var azuriteHost string
	var azuritePort int
	
	if ss.configManager == nil {
		return fmt.Errorf("configManager is required for storage container creation")
	}
	
	storageConfig := ss.configManager.GetStorageConfig()
	azuriteHost = storageConfig.AzuriteHost
	var err error
	azuritePort, err = strconv.Atoi(storageConfig.AzuritePort)
	if err != nil {
		return fmt.Errorf("invalid AZURITE_PORT from config: %w", err)
	}

	containers := []string{
		"content",
		"backups",
		"logs",
		"temp",
		"uploads",
	}

	azuriteEndpoint := fmt.Sprintf("http://%s:%d/devstoreaccount1", azuriteHost, azuritePort)
	
	for _, containerName := range containers {
		if err := ss.createBlobContainer(ctx, azuriteEndpoint, containerName); err != nil {
			return fmt.Errorf("failed to create container %s: %w", containerName, err)
		}
	}

	return nil
}

func (ss *StorageStack) createBlobContainer(ctx context.Context, endpoint, containerName string) error {
	containerURL := fmt.Sprintf("%s/%s?restype=container", endpoint, containerName)
	
	req, err := http.NewRequestWithContext(ctx, "PUT", containerURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("x-ms-version", "2020-08-04")
	req.Header.Set("x-ms-date", time.Now().UTC().Format(time.RFC1123))
	req.Header.Set("Content-Length", "0")
	
	client := &http.Client{Timeout: 30 * time.Second}
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	return nil
}

func (ss *StorageStack) ValidateDeployment(ctx context.Context, deployment sharedinfra.StorageDeployment) error {
	// Cast to concrete type to access implementation details
	concreteDeployment, ok := deployment.(*StorageDeployment)
	if !ok {
		return fmt.Errorf("deployment is not a valid StorageDeployment implementation")
	}
	if concreteDeployment.AzuriteContainer == nil {
		return fmt.Errorf("Azurite container is not deployed")
	}

	return nil
}

func (ss *StorageStack) GetStorageConnectionInfo() map[string]interface{} {
	var azuriteHost string
	var azuritePort int
	
	if ss.configManager == nil {
		return nil
	}
	
	storageConfig := ss.configManager.GetStorageConfig()
	azuriteHost = storageConfig.AzuriteHost
	var err error
	azuritePort, err = strconv.Atoi(storageConfig.AzuritePort)
	if err != nil {
		azuritePort = 10000
	}
	
	azuriteBlobPort := azuritePort     // 10000
	azuriteQueuePort := azuritePort + 1 // 10001
	azuriteTablePort := azuritePort + 2 // 10002

	return map[string]interface{}{
		"account_name":        "devstoreaccount1",
		"account_key":         "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==",
		"blob_endpoint":       fmt.Sprintf("http://%s:%d/devstoreaccount1", azuriteHost, azuriteBlobPort),
		"queue_endpoint":      fmt.Sprintf("http://%s:%d/devstoreaccount1", azuriteHost, azuriteQueuePort),
		"table_endpoint":      fmt.Sprintf("http://%s:%d/devstoreaccount1", azuriteHost, azuriteTablePort),
		"connection_string":   fmt.Sprintf("DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://%s:%d/devstoreaccount1;QueueEndpoint=http://%s:%d/devstoreaccount1;TableEndpoint=http://%s:%d/devstoreaccount1;", azuriteHost, azuriteBlobPort, azuriteHost, azuriteQueuePort, azuriteHost, azuriteTablePort),
		"blob_port":          azuriteBlobPort,
		"queue_port":         azuriteQueuePort,
		"table_port":         azuriteTablePort,
		"host":               azuriteHost,
	}
}

func (ss *StorageStack) GetBlobStorageEndpoint() string {
	var azuriteHost string
	var azuritePort int
	
	if ss.configManager == nil {
		return ""
	}
	
	storageConfig := ss.configManager.GetStorageConfig()
	azuriteHost = storageConfig.AzuriteHost
	var err error
	azuritePort, err = strconv.Atoi(storageConfig.AzuritePort)
	if err != nil {
		azuritePort = 10000
	}
	
	return fmt.Sprintf("http://%s:%d/devstoreaccount1", azuriteHost, azuritePort)
}

func (ss *StorageStack) GetDaprBindingConfiguration(serviceName string) map[string]interface{} {
	connectionInfo := ss.GetStorageConnectionInfo()
	
	return map[string]interface{}{
		"name":     "blob-storage-local",
		"type":     "bindings.azure.blobstorage",
		"version":  "v1",
		"metadata": map[string]string{
			"accountName":     connectionInfo["account_name"].(string),
			"accountKey":      connectionInfo["account_key"].(string),
			"containerName":   "content",
			"endpoint":        connectionInfo["blob_endpoint"].(string),
		},
		"scopes": []string{
			"content-api",
			"services-api",
		},
	}
}

// Helper methods that create resources with proper parent relationships for ComponentResource architecture

func (ss *StorageStack) createStorageNetworkWithParent(parent pulumi.ComponentResource) (*docker.Network, error) {
	network, err := docker.NewNetwork(ss.ctx, "storage-network", &docker.NetworkArgs{
		Name:   pulumi.Sprintf("%s-storage-network", ss.environment),
		Driver: pulumi.String("bridge"),
		Options: pulumi.StringMap{
			"com.docker.network.driver.mtu": pulumi.String("1500"),
		},
		Labels: docker.NetworkLabelArray{
			&docker.NetworkLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ss.environment),
			},
			&docker.NetworkLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("storage"),
			},
			&docker.NetworkLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},
	}, pulumi.Parent(parent))
	if err != nil {
		return nil, err
	}

	return network, nil
}

func (ss *StorageStack) createAzuriteDataVolumeWithParent(parent pulumi.ComponentResource) (*docker.Volume, error) {
	volume, err := docker.NewVolume(ss.ctx, "azurite-data", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-azurite-data", ss.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ss.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("azurite"),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("data-type"),
				Value: pulumi.String("persistent"),
			},
		},
	}, pulumi.Parent(parent))
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (ss *StorageStack) createAzuriteConfigVolumeWithParent(parent pulumi.ComponentResource) (*docker.Volume, error) {
	volume, err := docker.NewVolume(ss.ctx, "azurite-config", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-azurite-config", ss.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ss.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("azurite"),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("data-type"),
				Value: pulumi.String("configuration"),
			},
		},
	}, pulumi.Parent(parent))
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (ss *StorageStack) deployAzuriteContainerWithParent(deployment *StorageDeployment) (*docker.Container, error) {
	var azuritePort int
	
	if ss.configManager == nil {
		return nil, fmt.Errorf("configManager is required for storage deployment")
	}
	
	storageConfig := ss.configManager.GetStorageConfig()
	var err error
	azuritePort, err = strconv.Atoi(storageConfig.AzuritePort)
	if err != nil {
		return nil, fmt.Errorf("invalid AZURITE_PORT from config: %w", err)
	}
	
	azuriteBlobPort := azuritePort     // 10000
	azuriteQueuePort := azuritePort + 1 // 10001
	azuriteTablePort := azuritePort + 2 // 10002

	container, err := docker.NewContainer(ss.ctx, "azurite", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-azurite", ss.environment),
		Image:   pulumi.String("mcr.microsoft.com/azure-storage/azurite:latest"),
		Restart: pulumi.String("unless-stopped"),

		Command: pulumi.StringArray{
			pulumi.String("azurite"),
			pulumi.String("--blobHost"), pulumi.String("0.0.0.0"),
			pulumi.String("--blobPort"), pulumi.String("10000"),
			pulumi.String("--queueHost"), pulumi.String("0.0.0.0"),
			pulumi.String("--queuePort"), pulumi.String("10001"),
			pulumi.String("--tableHost"), pulumi.String("0.0.0.0"),
			pulumi.String("--tablePort"), pulumi.String("10002"),
			pulumi.String("--location"), pulumi.String("/workspace"),
			pulumi.String("--debug"), pulumi.String("/workspace/debug.log"),
			pulumi.String("--loose"),
			pulumi.String("--skipApiVersionCheck"),
		},

		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(10000),
				External: pulumi.Int(azuriteBlobPort),
				Protocol: pulumi.String("tcp"),
			},
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(10001),
				External: pulumi.Int(azuriteQueuePort),
				Protocol: pulumi.String("tcp"),
			},
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(10002),
				External: pulumi.Int(azuriteTablePort),
				Protocol: pulumi.String("tcp"),
			},
		},

		Mounts: docker.ContainerMountArray{
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.AzuriteDataVolume.Name,
				Target: pulumi.String("/workspace"),
			},
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.AzuriteConfigVolume.Name,
				Target: pulumi.String("/config"),
			},
		},

		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: deployment.StorageNetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("azurite"),
					pulumi.String("storage"),
					pulumi.String("blob-storage"),
				},
			},
		},

		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests: pulumi.StringArray{
				pulumi.String("CMD-SHELL"),
				pulumi.String("curl -f http://localhost:10000/devstoreaccount1 || exit 1"),
			},
			Interval: pulumi.String("30s"),
			Timeout:  pulumi.String("10s"),
			Retries:  pulumi.Int(3),
			StartPeriod: pulumi.String("30s"),
		},

		Labels: docker.ContainerLabelArray{
			&docker.ContainerLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ss.environment),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("azurite"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("service"),
				Value: pulumi.String("storage"),
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
	}, pulumi.Parent(deployment))
	if err != nil {
		return nil, err
	}

	return container, nil
}