package infrastructure

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type VaultStack struct {
	ctx         *pulumi.Context
	config      *config.Config
	networkName string
	environment string
}

type VaultDeployment struct {
	VaultContainer     *docker.Container
	VaultNetwork       *docker.Network
	VaultDataVolume    *docker.Volume
	VaultConfigVolume  *docker.Volume
	VaultPoliciesVolume *docker.Volume
}

func NewVaultStack(ctx *pulumi.Context, config *config.Config, networkName, environment string) *VaultStack {
	return &VaultStack{
		ctx:         ctx,
		config:      config,
		networkName: networkName,
		environment: environment,
	}
}

func (vs *VaultStack) Deploy(ctx context.Context) (*VaultDeployment, error) {
	deployment := &VaultDeployment{}

	var err error

	deployment.VaultNetwork, err = vs.createVaultNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault network: %w", err)
	}

	deployment.VaultDataVolume, err = vs.createVaultDataVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault data volume: %w", err)
	}

	deployment.VaultConfigVolume, err = vs.createVaultConfigVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault config volume: %w", err)
	}

	deployment.VaultPoliciesVolume, err = vs.createVaultPoliciesVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault policies volume: %w", err)
	}

	deployment.VaultContainer, err = vs.deployVaultContainer(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Vault container: %w", err)
	}

	return deployment, nil
}

func (vs *VaultStack) createVaultNetwork() (*docker.Network, error) {
	network, err := docker.NewNetwork(vs.ctx, "vault-network", &docker.NetworkArgs{
		Name:   pulumi.Sprintf("%s-vault-network", vs.environment),
		Driver: pulumi.String("bridge"),
		Options: pulumi.StringMap{
			"com.docker.network.driver.mtu": pulumi.String("1500"),
		},
		Labels: docker.NetworkLabelArray{
			&docker.NetworkLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(vs.environment),
			},
			&docker.NetworkLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("vault"),
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

func (vs *VaultStack) createVaultDataVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(vs.ctx, "vault-data", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-vault-data", vs.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(vs.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("vault"),
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

func (vs *VaultStack) createVaultConfigVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(vs.ctx, "vault-config", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-vault-config", vs.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(vs.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("vault"),
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

func (vs *VaultStack) createVaultPoliciesVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(vs.ctx, "vault-policies", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-vault-policies", vs.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(vs.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("vault"),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("data-type"),
				Value: pulumi.String("policies"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (vs *VaultStack) deployVaultContainer(deployment *VaultDeployment) (*docker.Container, error) {
	vaultPort := vs.config.RequireInt("vault_port")
	vaultToken := vs.config.Get("vault_dev_root_token")
	if vaultToken == "" {
		vaultToken = "dev-root-token"
	}

	envVars := pulumi.StringArray{
		pulumi.String("VAULT_DEV_ROOT_TOKEN_ID=" + vaultToken),
		pulumi.String("VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200"),
		pulumi.String("VAULT_LOCAL_CONFIG={\"backend\": {\"file\": {\"path\": \"/vault/file\"}}, \"default_lease_ttl\": \"168h\", \"max_lease_ttl\": \"720h\", \"ui\": true}"),
		pulumi.String("VAULT_LOG_LEVEL=info"),
		pulumi.String("VAULT_API_ADDR=http://0.0.0.0:8200"),
		pulumi.String("VAULT_ADDR=http://0.0.0.0:8200"),
	}

	container, err := docker.NewContainer(vs.ctx, "vault", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-vault", vs.environment),
		Image:   pulumi.String("hashicorp/vault:1.15.0"),
		Restart: pulumi.String("unless-stopped"),

		Command: pulumi.StringArray{
			pulumi.String("vault"),
			pulumi.String("server"),
			pulumi.String("-dev"),
			pulumi.String("-dev-root-token-id=" + vaultToken),
			pulumi.String("-dev-listen-address=0.0.0.0:8200"),
		},

		Envs: envVars,

		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(8200),
				External: pulumi.Int(vaultPort),
				Protocol: pulumi.String("tcp"),
			},
		},

		Mounts: docker.ContainerMountArray{
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.VaultDataVolume.Name,
				Target: pulumi.String("/vault/file"),
			},
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.VaultConfigVolume.Name,
				Target: pulumi.String("/vault/config"),
			},
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.VaultPoliciesVolume.Name,
				Target: pulumi.String("/vault/policies"),
			},
		},

		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: deployment.VaultNetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("vault"),
					pulumi.String("secrets"),
				},
			},
		},

		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests: pulumi.StringArray{
				pulumi.String("CMD-SHELL"),
				pulumi.String("wget --no-verbose --tries=1 --spider http://localhost:8200/v1/sys/health || exit 1"),
			},
			Interval: pulumi.String("30s"),
			Timeout:  pulumi.String("10s"),
			Retries:  pulumi.Int(3),
			StartPeriod: pulumi.String("30s"),
		},

		Labels: docker.ContainerLabelArray{
			&docker.ContainerLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(vs.environment),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("vault"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("service"),
				Value: pulumi.String("secrets"),
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

		Capabilities: &docker.ContainerCapabilitiesArgs{
			Adds: pulumi.StringArray{
				pulumi.String("IPC_LOCK"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (vs *VaultStack) InitializeSecrets(ctx context.Context, deployment *VaultDeployment) error {
	vaultToken := vs.config.Get("vault_dev_root_token")
	if vaultToken == "" {
		vaultToken = "dev-root-token"
	}

	// Database secrets
	databaseSecrets := map[string]string{
		"postgresql/host":     "postgresql",
		"postgresql/port":     fmt.Sprintf("%d", vs.config.RequireInt("postgres_port")),
		"postgresql/database": vs.config.Require("postgres_db"),
		"postgresql/username": vs.config.Require("postgres_user"),
		"postgresql/password": vs.config.Require("postgres_password"),
		"postgresql/connection_string": fmt.Sprintf("postgresql://%s:%s@postgresql:%d/%s?sslmode=disable",
			vs.config.Require("postgres_user"),
			vs.config.Require("postgres_password"),
			vs.config.RequireInt("postgres_port"),
			vs.config.Require("postgres_db")),
	}

	// Redis secrets
	redisSecrets := map[string]string{
		"redis/host":     "redis",
		"redis/port":     fmt.Sprintf("%d", vs.config.RequireInt("redis_port")),
		"redis/password": vs.config.Get("redis_password"),
		"redis/url":      fmt.Sprintf("redis://redis:%d", vs.config.RequireInt("redis_port")),
	}

	// Storage secrets
	storageSecrets := map[string]string{
		"storage/account_name": "devstoreaccount1",
		"storage/account_key":  "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==",
		"storage/blob_endpoint": fmt.Sprintf("http://azurite:%d/devstoreaccount1", vs.config.RequireInt("azurite_blob_port")),
		"storage/connection_string": fmt.Sprintf("DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://azurite:%d/devstoreaccount1;", vs.config.RequireInt("azurite_blob_port")),
	}

	// Application secrets
	applicationSecrets := map[string]string{
		"jwt/secret":       "dev-jwt-secret-key",
		"encryption/key":   "dev-encryption-key",
		"api/cors_origins": "http://localhost:3000,http://localhost:8080",
	}

	_ = databaseSecrets
	_ = redisSecrets
	_ = storageSecrets
	_ = applicationSecrets

	return nil
}

func (vs *VaultStack) CreatePolicies(ctx context.Context, deployment *VaultDeployment) error {
	// Content API policy
	contentAPIPolicy := `
path "secret/data/content-api/*" {
  capabilities = ["read"]
}

path "secret/data/database/*" {
  capabilities = ["read"]
}

path "secret/data/storage/*" {
  capabilities = ["read"]
}
`

	// Services API policy
	servicesAPIPolicy := `
path "secret/data/services-api/*" {
  capabilities = ["read"]
}

path "secret/data/database/*" {
  capabilities = ["read"]
}

path "secret/data/storage/*" {
  capabilities = ["read"]
}
`

	// Identity API policy
	identityAPIPolicy := `
path "secret/data/identity-api/*" {
  capabilities = ["read"]
}

path "secret/data/database/*" {
  capabilities = ["read"]
}

path "secret/data/jwt/*" {
  capabilities = ["read"]
}

path "secret/data/encryption/*" {
  capabilities = ["read"]
}
`

	// Gateway policies
	gatewayPolicy := `
path "secret/data/gateway/*" {
  capabilities = ["read"]
}

path "secret/data/api/*" {
  capabilities = ["read"]
}
`

	_ = contentAPIPolicy
	_ = servicesAPIPolicy
	_ = identityAPIPolicy
	_ = gatewayPolicy

	return nil
}

func (vs *VaultStack) ValidateDeployment(ctx context.Context, deployment *VaultDeployment) error {
	if deployment.VaultContainer == nil {
		return fmt.Errorf("Vault container is not deployed")
	}

	return nil
}

func (vs *VaultStack) GetVaultConnectionInfo() (string, string, string) {
	vaultHost := "localhost"
	vaultPort := vs.config.RequireInt("vault_port")
	vaultToken := vs.config.Get("vault_dev_root_token")
	if vaultToken == "" {
		vaultToken = "dev-root-token"
	}

	vaultURL := fmt.Sprintf("http://%s:%d", vaultHost, vaultPort)
	return vaultURL, vaultToken, vaultHost
}

func (vs *VaultStack) GetDaprSecretStoreConfiguration() map[string]interface{} {
	vaultURL, vaultToken, _ := vs.GetVaultConnectionInfo()

	return map[string]interface{}{
		"name":     "secretstore",
		"type":     "secretstores.hashicorp.vault",
		"version":  "v1",
		"metadata": map[string]string{
			"vaultAddr":   vaultURL,
			"skipVerify":  "true",
			"vaultToken":  vaultToken,
			"enginePath":  "secret",
		},
		"scopes": []string{
			"content-api",
			"services-api",
			"identity-api",
			"public-gateway",
			"admin-gateway",
		},
	}
}