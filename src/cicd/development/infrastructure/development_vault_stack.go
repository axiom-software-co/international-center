package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
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

type VaultStack struct {
	ctx           *pulumi.Context
	config        *config.Config
	configManager *sharedconfig.ConfigManager
	networkName   string
	environment   string
	
	// Outputs
	VaultEndpoint   pulumi.StringOutput `pulumi:"vaultEndpoint"`
	VaultToken      pulumi.StringOutput `pulumi:"vaultToken"`
	VaultNetworkID  pulumi.StringOutput `pulumi:"vaultNetworkId"`
	VaultContainerID pulumi.StringOutput `pulumi:"vaultContainerId"`
}

type VaultDeployment struct {
	pulumi.ComponentResource
	VaultContainer     *docker.Container
	VaultNetwork       *docker.Network
	VaultDataVolume    *docker.Volume
	VaultConfigVolume  *docker.Volume
	VaultPoliciesVolume *docker.Volume
	
	// Outputs
	VaultEndpoint   pulumi.StringOutput `pulumi:"vaultEndpoint"`
	VaultToken      pulumi.StringOutput `pulumi:"vaultToken"`
	NetworkID       pulumi.StringOutput `pulumi:"networkId"`
}

// Implement the shared VaultDeployment interface
func (vd *VaultDeployment) GetVaultEndpoint() pulumi.StringOutput {
	return vd.VaultEndpoint
}

func (vd *VaultDeployment) GetUIEndpoint() pulumi.StringOutput {
	return vd.VaultEndpoint
}

func (vd *VaultDeployment) GetRootToken() pulumi.StringOutput {
	return vd.VaultToken
}

func (vd *VaultDeployment) GetUnsealKeys() []pulumi.StringOutput {
	// In development mode, Vault runs in dev mode with no unseal keys
	return []pulumi.StringOutput{}
}

func (vd *VaultDeployment) GetServiceAccount() string {
	return "development-vault"
}

func NewVaultStack(ctx *pulumi.Context, config *config.Config, networkName, environment string) *VaultStack {
	// Create ConfigManager for centralized configuration
	configManager, err := sharedconfig.NewConfigManager(ctx)
	if err != nil {
		ctx.Log.Warn(fmt.Sprintf("Failed to create ConfigManager, using legacy configuration: %v", err), nil)
		configManager = nil
	}
	
	component := &VaultStack{
		ctx:           ctx,
		config:        config,
		configManager: configManager,
		networkName:   networkName,
		environment:   environment,
	}
	
	return component
}

func (vs *VaultStack) Deploy(ctx context.Context) (sharedinfra.VaultDeployment, error) {
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

	// Set deployment outputs
	vaultURL, vaultToken, _ := vs.GetVaultConnectionInfo()
	deployment.VaultEndpoint = pulumi.String(vaultURL).ToStringOutput()
	deployment.VaultToken = pulumi.String(vaultToken).ToStringOutput()
	deployment.NetworkID = deployment.VaultNetwork.ID().ToStringOutput()

	// Outputs are handled through stack outputs

	// Set stack outputs
	vs.VaultEndpoint = deployment.VaultEndpoint
	vs.VaultToken = deployment.VaultToken
	vs.VaultNetworkID = deployment.NetworkID
	vs.VaultContainerID = deployment.VaultContainer.ID().ToStringOutput()

	// Stack outputs are set in the struct fields and can be accessed directly

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

func (vs *VaultStack) InitializeSecrets(ctx context.Context, deployment sharedinfra.VaultDeployment) error {
	// Cast to concrete type to access implementation details
	_, ok := deployment.(*VaultDeployment)
	if !ok {
		return fmt.Errorf("deployment is not a valid VaultDeployment implementation")
	}
	vaultURL, vaultToken, _ := vs.GetVaultConnectionInfo()
	
	var databaseSecrets, redisSecrets, storageSecrets, applicationSecrets map[string]interface{}
	
	if vs.configManager == nil {
		return fmt.Errorf("configManager is required for vault secret initialization")
	}
	
	// Use ConfigManager for centralized configuration
	dbConfig := vs.configManager.GetDatabaseConfig()
	redisConfig := vs.configManager.GetRedisConfig()
	storageConfig := vs.configManager.GetStorageConfig()
	
	// Database secrets
	databaseSecrets = map[string]interface{}{
		"host":              dbConfig.ContainerHost,
		"port":              strconv.Itoa(dbConfig.Port),
		"database":          dbConfig.Database,
		"username":          dbConfig.User,
		"password":          dbConfig.Password,
		"connection_string": dbConfig.ContainerURL,
	}

	// Redis secrets
	redisSecrets = map[string]interface{}{
		"host":     redisConfig.ContainerHost,
		"port":     strconv.Itoa(redisConfig.Port),
		"password": redisConfig.Password,
		"addr":     redisConfig.ContainerAddr,
	}

	// Storage secrets
	storageSecrets = map[string]interface{}{
		"account_name":      "devstoreaccount1",
		"account_key":       "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==",
		"blob_endpoint":     storageConfig.BlobEndpoint,
		"connection_string": storageConfig.ConnectionString,
	}
	
	// Application secrets
	applicationSecrets = map[string]interface{}{
		"jwt_secret":     "dev-jwt-secret-key",
		"encryption_key": "dev-encryption-key",
		"cors_origins":   "", // TODO: Add to ConfigManager if needed
	}

	secretPaths := []struct {
		path string
		data map[string]interface{}
	}{
		{"secret/data/database", map[string]interface{}{"data": databaseSecrets}},
		{"secret/data/redis", map[string]interface{}{"data": redisSecrets}},
		{"secret/data/storage", map[string]interface{}{"data": storageSecrets}},
		{"secret/data/application", map[string]interface{}{"data": applicationSecrets}},
	}

	for _, secret := range secretPaths {
		if err := vs.storeSecret(ctx, vaultURL, vaultToken, secret.path, secret.data); err != nil {
			return fmt.Errorf("failed to store secret at %s: %w", secret.path, err)
		}
	}

	return nil
}

func (vs *VaultStack) storeSecret(ctx context.Context, vaultURL, vaultToken, path string, data map[string]interface{}) error {
	secretURL := fmt.Sprintf("%s/v1/%s", vaultURL, path)
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal secret data: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", secretURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("X-Vault-Token", vaultToken)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to store secret: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	return nil
}

func (vs *VaultStack) CreatePolicies(ctx context.Context, deployment sharedinfra.VaultDeployment) error {
	vaultURL, vaultToken, _ := vs.GetVaultConnectionInfo()

	policies := []struct {
		name   string
		policy string
	}{
		{
			name: "content-api-policy",
			policy: `
path "secret/data/database/*" {
  capabilities = ["read"]
}

path "secret/data/storage/*" {
  capabilities = ["read"]
}

path "secret/data/application/*" {
  capabilities = ["read"]
}
`,
		},
		{
			name: "services-api-policy",
			policy: `
path "secret/data/database/*" {
  capabilities = ["read"]
}

path "secret/data/storage/*" {
  capabilities = ["read"]
}

path "secret/data/application/*" {
  capabilities = ["read"]
}
`,
		},
		{
			name: "gateway-policy",
			policy: `
path "secret/data/database/*" {
  capabilities = ["read"]
}

path "secret/data/storage/*" {
  capabilities = ["read"]
}

path "secret/data/redis/*" {
  capabilities = ["read"]
}

path "secret/data/application/*" {
  capabilities = ["read"]
}
`,
		},
	}

	for _, policy := range policies {
		if err := vs.createPolicy(ctx, vaultURL, vaultToken, policy.name, policy.policy); err != nil {
			return fmt.Errorf("failed to create policy %s: %w", policy.name, err)
		}
	}

	return nil
}

func (vs *VaultStack) createPolicy(ctx context.Context, vaultURL, vaultToken, policyName, policyContent string) error {
	policyURL := fmt.Sprintf("%s/v1/sys/policy/%s", vaultURL, policyName)
	
	policyData := map[string]interface{}{
		"policy": policyContent,
	}
	
	jsonData, err := json.Marshal(policyData)
	if err != nil {
		return fmt.Errorf("failed to marshal policy data: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "PUT", policyURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("X-Vault-Token", vaultToken)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	return nil
}

func (vs *VaultStack) ValidateDeployment(ctx context.Context, deployment sharedinfra.VaultDeployment) error {
	// Cast to concrete type to access implementation details
	concreteDeployment, ok := deployment.(*VaultDeployment)
	if !ok {
		return fmt.Errorf("deployment is not a valid VaultDeployment implementation")
	}

	if concreteDeployment.VaultContainer == nil {
		return fmt.Errorf("Vault container is not deployed")
	}

	return nil
}

func (vs *VaultStack) GetVaultConnectionInfo() (string, string, error) {
	vaultHost := "localhost"
	vaultPort := vs.config.RequireInt("vault_port")
	vaultToken := vs.config.Get("vault_dev_root_token")
	if vaultToken == "" {
		vaultToken = "dev-root-token"
	}

	vaultURL := fmt.Sprintf("http://%s:%d", vaultHost, vaultPort)
	return vaultURL, vaultToken, nil
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