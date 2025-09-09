package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type VaultProvider string

const (
	VaultProviderHashiCorp   VaultProvider = "hashicorp_vault"
	VaultProviderAzureKeyVault VaultProvider = "azure_key_vault"
	VaultProviderAWSSecretsManager VaultProvider = "aws_secrets_manager"
	VaultProviderGCPSecretManager VaultProvider = "gcp_secret_manager"
	VaultProviderKubernetes  VaultProvider = "kubernetes_secrets"
)

type VaultConfig struct {
	Provider          VaultProvider
	Address           string
	Port              int
	Token             string
	Username          string
	Password          string
	Namespace         string
	MountPath         string
	UseTLS            bool
	HealthCheckPath   string
	HealthCheckPort   int
	MaxRetries        int
	Timeout           int
	EngineType        string
	RotationEnabled   bool
	AuditEnabled      bool
	SealingEnabled    bool
	AdditionalParams  map[string]string
}

type VaultArgs struct {
	Config      *VaultConfig
	Environment string
	ProjectName string
}

type VaultComponent struct {
	pulumi.ResourceState

	VaultAddress    pulumi.StringOutput `pulumi:"vaultAddress"`
	Address         pulumi.StringOutput `pulumi:"address"`
	Port            pulumi.IntOutput    `pulumi:"port"`
	Namespace       pulumi.StringOutput `pulumi:"namespace"`
	MountPath       pulumi.StringOutput `pulumi:"mountPath"`
	HealthEndpoint  pulumi.StringOutput `pulumi:"healthEndpoint"`
	Provider        pulumi.StringOutput `pulumi:"provider"`
	UseTLS          pulumi.BoolOutput   `pulumi:"useTLS"`
	AuditEnabled    pulumi.BoolOutput   `pulumi:"auditEnabled"`
	SealingEnabled  pulumi.BoolOutput   `pulumi:"sealingEnabled"`
}

func NewVaultComponent(ctx *pulumi.Context, name string, args *VaultArgs, opts ...pulumi.ResourceOption) (*VaultComponent, error) {
	component := &VaultComponent{}
	
	if ctx != nil {
		err := ctx.RegisterComponentResource("framework:vault:Vault", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	config := args.Config
	if config == nil {
		return nil, fmt.Errorf("vault config is required")
	}

	vaultAddress := buildVaultAddress(config)
	healthEndpoint := buildVaultHealthEndpoint(config)

	component.VaultAddress = pulumi.String(vaultAddress).ToStringOutput()
	component.Address = pulumi.String(config.Address).ToStringOutput()
	component.Port = pulumi.Int(config.Port).ToIntOutput()
	component.Namespace = pulumi.String(config.Namespace).ToStringOutput()
	component.MountPath = pulumi.String(config.MountPath).ToStringOutput()
	component.HealthEndpoint = pulumi.String(healthEndpoint).ToStringOutput()
	component.Provider = pulumi.String(string(config.Provider)).ToStringOutput()
	component.UseTLS = pulumi.Bool(config.UseTLS).ToBoolOutput()
	component.AuditEnabled = pulumi.Bool(config.AuditEnabled).ToBoolOutput()
	component.SealingEnabled = pulumi.Bool(config.SealingEnabled).ToBoolOutput()

	if ctx != nil {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"vaultAddress":    component.VaultAddress,
			"address":         component.Address,
			"port":            component.Port,
			"namespace":       component.Namespace,
			"mountPath":       component.MountPath,
			"healthEndpoint":  component.HealthEndpoint,
			"provider":        component.Provider,
			"useTLS":          component.UseTLS,
			"auditEnabled":    component.AuditEnabled,
			"sealingEnabled":  component.SealingEnabled,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}

func buildVaultAddress(config *VaultConfig) string {
	protocol := "http"
	if config.UseTLS {
		protocol = "https"
	}
	
	switch config.Provider {
	case VaultProviderHashiCorp:
		return fmt.Sprintf("%s://%s:%d", protocol, config.Address, config.Port)
		
	case VaultProviderAzureKeyVault:
		return fmt.Sprintf("https://%s.vault.azure.net/", config.Address)
		
	case VaultProviderAWSSecretsManager:
		return fmt.Sprintf("https://secretsmanager.%s.amazonaws.com", config.Address)
		
	case VaultProviderGCPSecretManager:
		return "https://secretmanager.googleapis.com"
		
	case VaultProviderKubernetes:
		return "kubernetes://secrets"
		
	default:
		return fmt.Sprintf("%s://%s:%d", protocol, config.Address, config.Port)
	}
}

func buildVaultHealthEndpoint(config *VaultConfig) string {
	if config.HealthCheckPath == "" {
		return ""
	}
	
	port := config.HealthCheckPort
	if port == 0 {
		port = config.Port
	}
	
	protocol := "http"
	if config.UseTLS {
		protocol = "https"
	}
	
	switch config.Provider {
	case VaultProviderHashiCorp:
		return fmt.Sprintf("%s://%s:%d%s", protocol, config.Address, port, config.HealthCheckPath)
		
	case VaultProviderAzureKeyVault:
		return fmt.Sprintf("https://%s.vault.azure.net/health", config.Address)
		
	case VaultProviderAWSSecretsManager:
		return fmt.Sprintf("https://secretsmanager.%s.amazonaws.com/health", config.Address)
		
	case VaultProviderGCPSecretManager:
		return "https://secretmanager.googleapis.com/health"
		
	case VaultProviderKubernetes:
		return "kubernetes://health"
		
	default:
		return fmt.Sprintf("%s://%s:%d%s", protocol, config.Address, port, config.HealthCheckPath)
	}
}

func DefaultHashiCorpVaultConfig(address string) *VaultConfig {
	return &VaultConfig{
		Provider:         VaultProviderHashiCorp,
		Address:          address,
		Port:             8200,
		Token:            "",
		Username:         "",
		Password:         "",
		Namespace:        "",
		MountPath:        "secret",
		UseTLS:           true,
		HealthCheckPath:  "/v1/sys/health",
		HealthCheckPort:  8200,
		MaxRetries:       3,
		Timeout:          30,
		EngineType:       "kv-v2",
		RotationEnabled:  true,
		AuditEnabled:     true,
		SealingEnabled:   true,
		AdditionalParams: make(map[string]string),
	}
}

func DefaultAzureKeyVaultConfig(vaultName string) *VaultConfig {
	return &VaultConfig{
		Provider:         VaultProviderAzureKeyVault,
		Address:          vaultName,
		Port:             443,
		Token:            "",
		Username:         "",
		Password:         "",
		Namespace:        "",
		MountPath:        "",
		UseTLS:           true,
		HealthCheckPath:  "/health",
		HealthCheckPort:  443,
		MaxRetries:       3,
		Timeout:          30,
		EngineType:       "azure",
		RotationEnabled:  true,
		AuditEnabled:     true,
		SealingEnabled:   false,
		AdditionalParams: make(map[string]string),
	}
}

func DefaultAWSSecretsManagerConfig(region string) *VaultConfig {
	return &VaultConfig{
		Provider:         VaultProviderAWSSecretsManager,
		Address:          region,
		Port:             443,
		Token:            "",
		Username:         "",
		Password:         "",
		Namespace:        "",
		MountPath:        "",
		UseTLS:           true,
		HealthCheckPath:  "/health",
		HealthCheckPort:  443,
		MaxRetries:       3,
		Timeout:          30,
		EngineType:       "aws",
		RotationEnabled:  true,
		AuditEnabled:     true,
		SealingEnabled:   false,
		AdditionalParams: make(map[string]string),
	}
}

func DefaultGCPSecretManagerConfig() *VaultConfig {
	return &VaultConfig{
		Provider:         VaultProviderGCPSecretManager,
		Address:          "secretmanager.googleapis.com",
		Port:             443,
		Token:            "",
		Username:         "",
		Password:         "",
		Namespace:        "",
		MountPath:        "",
		UseTLS:           true,
		HealthCheckPath:  "/health",
		HealthCheckPort:  443,
		MaxRetries:       3,
		Timeout:          30,
		EngineType:       "gcp",
		RotationEnabled:  true,
		AuditEnabled:     true,
		SealingEnabled:   false,
		AdditionalParams: make(map[string]string),
	}
}

func DefaultKubernetesSecretsConfig() *VaultConfig {
	return &VaultConfig{
		Provider:         VaultProviderKubernetes,
		Address:          "kubernetes",
		Port:             0,
		Token:            "",
		Username:         "",
		Password:         "",
		Namespace:        "default",
		MountPath:        "",
		UseTLS:           false,
		HealthCheckPath:  "/health",
		HealthCheckPort:  0,
		MaxRetries:       3,
		Timeout:          30,
		EngineType:       "kubernetes",
		RotationEnabled:  false,
		AuditEnabled:     false,
		SealingEnabled:   false,
		AdditionalParams: make(map[string]string),
	}
}