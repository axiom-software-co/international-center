package infrastructure

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type StorageStack struct {
	ctx         *pulumi.Context
	config      *config.Config
	networkName string
	environment string
}

type StorageDeployment struct {
	AzuriteContainer    *docker.Container
	StorageNetwork      *docker.Network
	AzuriteDataVolume   *docker.Volume
	AzuriteConfigVolume *docker.Volume
}

func NewStorageStack(ctx *pulumi.Context, config *config.Config, networkName, environment string) *StorageStack {
	return &StorageStack{
		ctx:         ctx,
		config:      config,
		networkName: networkName,
		environment: environment,
	}
}

func (ss *StorageStack) Deploy(ctx context.Context) (*StorageDeployment, error) {
	deployment := &StorageDeployment{}

	var err error

	deployment.StorageNetwork, err = ss.createStorageNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to create storage network: %w", err)
	}

	deployment.AzuriteDataVolume, err = ss.createAzuriteDataVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create Azurite data volume: %w", err)
	}

	deployment.AzuriteConfigVolume, err = ss.createAzuriteConfigVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create Azurite config volume: %w", err)
	}

	deployment.AzuriteContainer, err = ss.deployAzuriteContainer(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Azurite container: %w", err)
	}

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
	azuriteBlobPort := ss.config.RequireInt("azurite_blob_port")
	azuriteQueuePort := ss.config.RequireInt("azurite_queue_port")
	azuriteTablePort := ss.config.RequireInt("azurite_table_port")

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

func (ss *StorageStack) CreateStorageContainers(ctx context.Context, deployment *StorageDeployment) error {
	containers := []string{
		"content",
		"backups",
		"logs",
		"temp",
		"uploads",
	}

	_ = containers

	return nil
}

func (ss *StorageStack) ValidateDeployment(ctx context.Context, deployment *StorageDeployment) error {
	if deployment.AzuriteContainer == nil {
		return fmt.Errorf("Azurite container is not deployed")
	}

	return nil
}

func (ss *StorageStack) GetStorageConnectionInfo() map[string]interface{} {
	azuriteBlobPort := ss.config.RequireInt("azurite_blob_port")
	azuriteQueuePort := ss.config.RequireInt("azurite_queue_port")
	azuriteTablePort := ss.config.RequireInt("azurite_table_port")

	return map[string]interface{}{
		"account_name":        "devstoreaccount1",
		"account_key":         "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==",
		"blob_endpoint":       fmt.Sprintf("http://localhost:%d/devstoreaccount1", azuriteBlobPort),
		"queue_endpoint":      fmt.Sprintf("http://localhost:%d/devstoreaccount1", azuriteQueuePort),
		"table_endpoint":      fmt.Sprintf("http://localhost:%d/devstoreaccount1", azuriteTablePort),
		"connection_string":   fmt.Sprintf("DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://localhost:%d/devstoreaccount1;QueueEndpoint=http://localhost:%d/devstoreaccount1;TableEndpoint=http://localhost:%d/devstoreaccount1;", azuriteBlobPort, azuriteQueuePort, azuriteTablePort),
		"blob_port":          azuriteBlobPort,
		"queue_port":         azuriteQueuePort,
		"table_port":         azuriteTablePort,
		"host":               "localhost",
	}
}

func (ss *StorageStack) GetBlobStorageEndpoint() string {
	azuriteBlobPort := ss.config.RequireInt("azurite_blob_port")
	return fmt.Sprintf("http://localhost:%d/devstoreaccount1", azuriteBlobPort)
}

func (ss *StorageStack) GetDaprBindingConfiguration() map[string]interface{} {
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