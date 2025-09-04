package infrastructure

import (
	"fmt"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/deployer/shared/config"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// SecretProvider represents the type of secret provider
type SecretProvider string

const (
	SecretProviderLocal     SecretProvider = "local"
	SecretProviderVault     SecretProvider = "vault"
	SecretProviderAzureKV   SecretProvider = "azure-keyvault"
	SecretProviderK8sSecret SecretProvider = "kubernetes-secret"
)

// SecretType represents the type of secret
type SecretType string

const (
	SecretTypeGeneric     SecretType = "generic"
	SecretTypeDatabase    SecretType = "database"
	SecretTypeAPI         SecretType = "api"
	SecretTypeTLS         SecretType = "tls"
	SecretTypeOAuth       SecretType = "oauth"
	SecretTypeJWT         SecretType = "jwt"
	SecretTypeStorage     SecretType = "storage"
)

// SecretRotationPolicy defines how secrets should be rotated
type SecretRotationPolicy struct {
	Enabled       bool          `json:"enabled"`
	RotationDays  int           `json:"rotation_days"`
	RetainOldDays int           `json:"retain_old_days"`
	AutoRotate    bool          `json:"auto_rotate"`
	NotifyDays    int           `json:"notify_days"`
}

// SecretMetadata contains metadata for a secret
type SecretMetadata struct {
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	SecretType   SecretType            `json:"secret_type"`
	Provider     SecretProvider        `json:"provider"`
	Tags         map[string]string     `json:"tags"`
	CreatedAt    time.Time             `json:"created_at"`
	ExpiresAt    *time.Time            `json:"expires_at,omitempty"`
	RotationPolicy *SecretRotationPolicy `json:"rotation_policy,omitempty"`
}

// SecretValue represents a secret value with metadata
type SecretValue struct {
	Metadata SecretMetadata    `json:"metadata"`
	Data     map[string]string `json:"data"`
	Version  string            `json:"version"`
}

// SecretConfiguration defines secret configuration for services
type SecretConfiguration struct {
	ServiceName   string                    `json:"service_name"`
	Secrets       map[string]SecretMetadata `json:"secrets"`
	SecretMounts  map[string]string         `json:"secret_mounts"`
	EnvSecrets    map[string]string         `json:"env_secrets"`
}

// SecretsManager manages secrets across different providers and environments
type SecretsManager struct {
	configManager *config.ConfigManager
	ctx           *pulumi.Context
	provider      SecretProvider
}

// NewSecretsManager creates a new secrets manager
func NewSecretsManager(ctx *pulumi.Context, configManager *config.ConfigManager) *SecretsManager {
	sm := &SecretsManager{
		configManager: configManager,
		ctx:           ctx,
	}
	sm.provider = sm.determineSecretProvider(configManager)
	
	return sm
}

// CreateSecretConfigurations creates secret configurations for all services
func (sm *SecretsManager) CreateSecretConfigurations() (map[string]*SecretConfiguration, error) {
	configurations := make(map[string]*SecretConfiguration)
	
	// Database secrets
	configurations["postgresql"] = sm.createPostgreSQLSecrets()
	configurations["redis"] = sm.createRedisSecrets()
	
	// Infrastructure secrets
	configurations["vault"] = sm.createVaultSecrets()
	configurations["grafana"] = sm.createGrafanaSecrets()
	configurations["azurite"] = sm.createAzuriteSecrets()
	
	// Application secrets
	applications := []string{
		"content-api",
		"services-api", 
		"public-gateway",
		"admin-gateway",
	}
	for _, appName := range applications {
		configurations[appName] = sm.createApplicationSecrets(appName)
	}
	
	// System secrets
	configurations["system"] = sm.createSystemSecrets()
	
	return configurations, nil
}

// CreateSecret creates a secret in the configured provider
func (sm *SecretsManager) CreateSecret(secretName string, secretData map[string]string, metadata SecretMetadata) (pulumi.Resource, error) {
	switch sm.provider {
	case SecretProviderVault:
		return sm.createVaultSecret(secretName, secretData, metadata)
	case SecretProviderAzureKV:
		return sm.createAzureKeyVaultSecret(secretName, secretData, metadata)
	case SecretProviderLocal:
		return sm.createLocalSecret(secretName, secretData, metadata)
	case SecretProviderK8sSecret:
		return sm.createKubernetesSecret(secretName, secretData, metadata)
	default:
		return nil, fmt.Errorf("unsupported secret provider: %s", sm.provider)
	}
}

// Service-specific secret configurations

func (sm *SecretsManager) createPostgreSQLSecrets() *SecretConfiguration {
	secrets := make(map[string]SecretMetadata)
	secretMounts := make(map[string]string)
	envSecrets := make(map[string]string)
	
	// Database password
	secrets["password"] = SecretMetadata{
		Name:        "postgresql-password",
		Description: "PostgreSQL database password",
		SecretType:  SecretTypeDatabase,
		Provider:    sm.provider,
		Tags: map[string]string{
			"service":     "postgresql",
			"type":        "password",
			"environment": string(sm.configManager.GetEnvironment()),
		},
		RotationPolicy: &SecretRotationPolicy{
			Enabled:       true,
			RotationDays:  90,
			RetainOldDays: 7,
			AutoRotate:    !sm.configManager.IsProduction(),
			NotifyDays:    14,
		},
	}
	
	// Replication password
	secrets["replication-password"] = SecretMetadata{
		Name:        "postgresql-replication-password",
		Description: "PostgreSQL replication password",
		SecretType:  SecretTypeDatabase,
		Provider:    sm.provider,
		Tags: map[string]string{
			"service":     "postgresql",
			"type":        "replication",
			"environment": string(sm.configManager.GetEnvironment()),
		},
		RotationPolicy: &SecretRotationPolicy{
			Enabled:       true,
			RotationDays:  90,
			RetainOldDays: 7,
			AutoRotate:    false,
			NotifyDays:    14,
		},
	}
	
	// TLS certificates for SSL connections
	secrets["tls-cert"] = SecretMetadata{
		Name:        "postgresql-tls-cert",
		Description: "PostgreSQL TLS certificate",
		SecretType:  SecretTypeTLS,
		Provider:    sm.provider,
		Tags: map[string]string{
			"service":     "postgresql",
			"type":        "tls",
			"environment": string(sm.configManager.GetEnvironment()),
		},
		RotationPolicy: &SecretRotationPolicy{
			Enabled:       true,
			RotationDays:  365,
			RetainOldDays: 30,
			AutoRotate:    false,
			NotifyDays:    30,
		},
	}
	
	// Environment variables mapping
	envSecrets["POSTGRES_PASSWORD"] = "postgresql-password"
	envSecrets["POSTGRES_REPLICATION_PASSWORD"] = "postgresql-replication-password"
	
	// Secret mounts
	secretMounts["password"] = "/var/secrets/postgresql/password"
	secretMounts["tls-cert"] = "/var/secrets/postgresql/tls"
	
	return &SecretConfiguration{
		ServiceName:  "postgresql",
		Secrets:      secrets,
		SecretMounts: secretMounts,
		EnvSecrets:   envSecrets,
	}
}

func (sm *SecretsManager) createRedisSecrets() *SecretConfiguration {
	secrets := make(map[string]SecretMetadata)
	secretMounts := make(map[string]string)
	envSecrets := make(map[string]string)
	
	// Redis password
	secrets["password"] = SecretMetadata{
		Name:        "redis-password",
		Description: "Redis authentication password",
		SecretType:  SecretTypeDatabase,
		Provider:    sm.provider,
		Tags: map[string]string{
			"service":     "redis",
			"type":        "password",
			"environment": string(sm.configManager.GetEnvironment()),
		},
		RotationPolicy: &SecretRotationPolicy{
			Enabled:       true,
			RotationDays:  60,
			RetainOldDays: 7,
			AutoRotate:    !sm.configManager.IsProduction(),
			NotifyDays:    14,
		},
	}
	
	// TLS certificate for SSL connections
	secrets["tls-cert"] = SecretMetadata{
		Name:        "redis-tls-cert",
		Description: "Redis TLS certificate",
		SecretType:  SecretTypeTLS,
		Provider:    sm.provider,
		Tags: map[string]string{
			"service":     "redis",
			"type":        "tls",
			"environment": string(sm.configManager.GetEnvironment()),
		},
		RotationPolicy: &SecretRotationPolicy{
			Enabled:       true,
			RotationDays:  365,
			RetainOldDays: 30,
			AutoRotate:    false,
			NotifyDays:    30,
		},
	}
	
	envSecrets["REDIS_PASSWORD"] = "redis-password"
	secretMounts["password"] = "/var/secrets/redis/password"
	secretMounts["tls-cert"] = "/var/secrets/redis/tls"
	
	return &SecretConfiguration{
		ServiceName:  "redis",
		Secrets:      secrets,
		SecretMounts: secretMounts,
		EnvSecrets:   envSecrets,
	}
}

func (sm *SecretsManager) createVaultSecrets() *SecretConfiguration {
	secrets := make(map[string]SecretMetadata)
	secretMounts := make(map[string]string)
	envSecrets := make(map[string]string)
	
	// Vault root token (only for initialization)
	if sm.configManager.IsDevelopment() {
		secrets["root-token"] = SecretMetadata{
			Name:        "vault-root-token",
			Description: "Vault root token (development only)",
			SecretType:  SecretTypeAPI,
			Provider:    sm.provider,
			Tags: map[string]string{
				"service":     "vault",
				"type":        "root-token",
				"environment": "development",
			},
			RotationPolicy: &SecretRotationPolicy{
				Enabled:    false,
				AutoRotate: false,
			},
		}
		envSecrets["VAULT_ROOT_TOKEN"] = "vault-root-token"
	}
	
	// Vault unseal keys
	for i := 1; i <= 5; i++ {
		secrets[fmt.Sprintf("unseal-key-%d", i)] = SecretMetadata{
			Name:        fmt.Sprintf("vault-unseal-key-%d", i),
			Description: fmt.Sprintf("Vault unseal key %d", i),
			SecretType:  SecretTypeAPI,
			Provider:    sm.provider,
			Tags: map[string]string{
				"service":     "vault",
				"type":        "unseal-key",
				"key-number":  fmt.Sprintf("%d", i),
				"environment": string(sm.configManager.GetEnvironment()),
			},
			RotationPolicy: &SecretRotationPolicy{
				Enabled:    false,
				AutoRotate: false,
			},
		}
	}
	
	// TLS certificates
	secrets["tls-cert"] = SecretMetadata{
		Name:        "vault-tls-cert",
		Description: "Vault TLS certificate",
		SecretType:  SecretTypeTLS,
		Provider:    sm.provider,
		Tags: map[string]string{
			"service":     "vault",
			"type":        "tls",
			"environment": string(sm.configManager.GetEnvironment()),
		},
		RotationPolicy: &SecretRotationPolicy{
			Enabled:       true,
			RotationDays:  365,
			RetainOldDays: 30,
			AutoRotate:    false,
			NotifyDays:    30,
		},
	}
	
	secretMounts["tls-cert"] = "/vault/tls"
	secretMounts["unseal-keys"] = "/vault/keys"
	
	return &SecretConfiguration{
		ServiceName:  "vault",
		Secrets:      secrets,
		SecretMounts: secretMounts,
		EnvSecrets:   envSecrets,
	}
}

func (sm *SecretsManager) createGrafanaSecrets() *SecretConfiguration {
	secrets := make(map[string]SecretMetadata)
	secretMounts := make(map[string]string)
	envSecrets := make(map[string]string)
	
	// Admin password
	secrets["admin-password"] = SecretMetadata{
		Name:        "grafana-admin-password",
		Description: "Grafana admin password",
		SecretType:  SecretTypeAPI,
		Provider:    sm.provider,
		Tags: map[string]string{
			"service":     "grafana",
			"type":        "password",
			"environment": string(sm.configManager.GetEnvironment()),
		},
		RotationPolicy: &SecretRotationPolicy{
			Enabled:       true,
			RotationDays:  90,
			RetainOldDays: 7,
			AutoRotate:    !sm.configManager.IsProduction(),
			NotifyDays:    14,
		},
	}
	
	// Secret key for session encryption
	secrets["secret-key"] = SecretMetadata{
		Name:        "grafana-secret-key",
		Description: "Grafana secret key for session encryption",
		SecretType:  SecretTypeAPI,
		Provider:    sm.provider,
		Tags: map[string]string{
			"service":     "grafana",
			"type":        "encryption-key",
			"environment": string(sm.configManager.GetEnvironment()),
		},
		RotationPolicy: &SecretRotationPolicy{
			Enabled:       true,
			RotationDays:  180,
			RetainOldDays: 14,
			AutoRotate:    false,
			NotifyDays:    30,
		},
	}
	
	// Database connection
	secrets["database-password"] = SecretMetadata{
		Name:        "grafana-database-password",
		Description: "Grafana database password",
		SecretType:  SecretTypeDatabase,
		Provider:    sm.provider,
		Tags: map[string]string{
			"service":     "grafana",
			"type":        "database-password",
			"environment": string(sm.configManager.GetEnvironment()),
		},
		RotationPolicy: &SecretRotationPolicy{
			Enabled:       true,
			RotationDays:  90,
			RetainOldDays: 7,
			AutoRotate:    false,
			NotifyDays:    14,
		},
	}
	
	envSecrets["GF_SECURITY_ADMIN_PASSWORD"] = "grafana-admin-password"
	envSecrets["GF_SECURITY_SECRET_KEY"] = "grafana-secret-key"
	envSecrets["GF_DATABASE_PASSWORD"] = "grafana-database-password"
	
	secretMounts["admin"] = "/var/secrets/grafana/admin"
	
	return &SecretConfiguration{
		ServiceName:  "grafana",
		Secrets:      secrets,
		SecretMounts: secretMounts,
		EnvSecrets:   envSecrets,
	}
}

func (sm *SecretsManager) createAzuriteSecrets() *SecretConfiguration {
	secrets := make(map[string]SecretMetadata)
	envSecrets := make(map[string]string)
	
	// Only for development - staging/production use real Azure Storage with managed identity
	if sm.configManager.IsDevelopment() {
		secrets["account-key"] = SecretMetadata{
			Name:        "azurite-account-key",
			Description: "Azurite storage account key",
			SecretType:  SecretTypeStorage,
			Provider:    sm.provider,
			Tags: map[string]string{
				"service":     "azurite",
				"type":        "storage-key",
				"environment": "development",
			},
			RotationPolicy: &SecretRotationPolicy{
				Enabled:    false,
				AutoRotate: false,
			},
		}
		
		envSecrets["AZURITE_ACCOUNT_KEY"] = "azurite-account-key"
	}
	
	return &SecretConfiguration{
		ServiceName: "azurite",
		Secrets:     secrets,
		EnvSecrets:  envSecrets,
	}
}

func (sm *SecretsManager) createApplicationSecrets(serviceName string) *SecretConfiguration {
	secrets := make(map[string]SecretMetadata)
	secretMounts := make(map[string]string)
	envSecrets := make(map[string]string)
	
	// JWT signing key
	secrets["jwt-key"] = SecretMetadata{
		Name:        fmt.Sprintf("%s-jwt-key", serviceName),
		Description: fmt.Sprintf("%s JWT signing key", serviceName),
		SecretType:  SecretTypeJWT,
		Provider:    sm.provider,
		Tags: map[string]string{
			"service":     serviceName,
			"type":        "jwt-key",
			"environment": string(sm.configManager.GetEnvironment()),
		},
		RotationPolicy: &SecretRotationPolicy{
			Enabled:       true,
			RotationDays:  180,
			RetainOldDays: 30,
			AutoRotate:    false,
			NotifyDays:    30,
		},
	}
	
	// API keys for external services
	secrets["external-api-key"] = SecretMetadata{
		Name:        fmt.Sprintf("%s-external-api-key", serviceName),
		Description: fmt.Sprintf("%s external API key", serviceName),
		SecretType:  SecretTypeAPI,
		Provider:    sm.provider,
		Tags: map[string]string{
			"service":     serviceName,
			"type":        "external-api-key",
			"environment": string(sm.configManager.GetEnvironment()),
		},
		RotationPolicy: &SecretRotationPolicy{
			Enabled:       true,
			RotationDays:  90,
			RetainOldDays: 7,
			AutoRotate:    false,
			NotifyDays:    14,
		},
	}
	
	// TODO: Database connection password (if service has its own DB)
	// Database field not yet implemented in ApplicationConfig
	// if appConfig.Database.Enabled {
	//     secrets["db-password"] = SecretMetadata{...}
	//     envSecrets["DATABASE_PASSWORD"] = fmt.Sprintf("%s-db-password", serviceName)
	// }
	
	// TODO: TLS certificates
	// TLS field not yet implemented in ApplicationConfig
	// if appConfig.TLS.Enabled {
	//     secrets["tls-cert"] = SecretMetadata{...}
	//     secretMounts["tls-cert"] = "/app/tls"
	// }
	
	envSecrets["JWT_SIGNING_KEY"] = fmt.Sprintf("%s-jwt-key", serviceName)
	envSecrets["EXTERNAL_API_KEY"] = fmt.Sprintf("%s-external-api-key", serviceName)
	
	secretMounts["secrets"] = "/app/secrets"
	
	return &SecretConfiguration{
		ServiceName:  serviceName,
		Secrets:      secrets,
		SecretMounts: secretMounts,
		EnvSecrets:   envSecrets,
	}
}

func (sm *SecretsManager) createSystemSecrets() *SecretConfiguration {
	secrets := make(map[string]SecretMetadata)
	envSecrets := make(map[string]string)
	
	// Dapr API token
	secrets["dapr-api-token"] = SecretMetadata{
		Name:        "dapr-api-token",
		Description: "Dapr API authentication token",
		SecretType:  SecretTypeAPI,
		Provider:    sm.provider,
		Tags: map[string]string{
			"service":     "dapr",
			"type":        "api-token",
			"environment": string(sm.configManager.GetEnvironment()),
		},
		RotationPolicy: &SecretRotationPolicy{
			Enabled:       true,
			RotationDays:  90,
			RetainOldDays: 7,
			AutoRotate:    false,
			NotifyDays:    14,
		},
	}
	
	// Container registry credentials
	secrets["registry-credentials"] = SecretMetadata{
		Name:        "container-registry-credentials",
		Description: "Container registry authentication credentials",
		SecretType:  SecretTypeAPI,
		Provider:    sm.provider,
		Tags: map[string]string{
			"service":     "registry",
			"type":        "credentials",
			"environment": string(sm.configManager.GetEnvironment()),
		},
		RotationPolicy: &SecretRotationPolicy{
			Enabled:       true,
			RotationDays:  180,
			RetainOldDays: 14,
			AutoRotate:    false,
			NotifyDays:    30,
		},
	}
	
	// Backup encryption key
	secrets["backup-encryption-key"] = SecretMetadata{
		Name:        "backup-encryption-key",
		Description: "Backup encryption key",
		SecretType:  SecretTypeGeneric,
		Provider:    sm.provider,
		Tags: map[string]string{
			"service":     "backup",
			"type":        "encryption-key",
			"environment": string(sm.configManager.GetEnvironment()),
		},
		RotationPolicy: &SecretRotationPolicy{
			Enabled:       true,
			RotationDays:  365,
			RetainOldDays: 90,
			AutoRotate:    false,
			NotifyDays:    60,
		},
	}
	
	envSecrets["DAPR_API_TOKEN"] = "dapr-api-token"
	envSecrets["REGISTRY_PASSWORD"] = "registry-credentials"
	envSecrets["BACKUP_ENCRYPTION_KEY"] = "backup-encryption-key"
	
	return &SecretConfiguration{
		ServiceName: "system",
		Secrets:     secrets,
		EnvSecrets:  envSecrets,
	}
}

// Secret creation methods for different providers

func (sm *SecretsManager) createVaultSecret(secretName string, secretData map[string]string, metadata SecretMetadata) (pulumi.Resource, error) {
	// For Vault, we would typically use the Vault API or Vault provider
	// This is a placeholder for the actual Vault secret creation
	sm.ctx.Log.Info(fmt.Sprintf("Creating Vault secret: %s", secretName), nil)
	return nil, nil
}

func (sm *SecretsManager) createAzureKeyVaultSecret(secretName string, secretData map[string]string, metadata SecretMetadata) (pulumi.Resource, error) {
	// TODO: Fix when Vault and ResourceGroupName fields are added to DeploymentConfig
	// Convert secret data to JSON for Azure Key Vault
	// dataJSON, err := json.Marshal(secretData)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to marshal secret data: %w", err)
	// }
	// secret, err := keyvault.NewSecret(sm.ctx, secretName, &keyvault.SecretArgs{
	//     SecretName:        pulumi.String(metadata.Name),
	//     VaultName:         pulumi.String(sm.config.Vault.VaultName),
	//     ResourceGroupName: pulumi.String(sm.config.ResourceGroupName),
	//     Properties: keyvault.SecretPropertiesArgs{
	//         Value: pulumi.String(string(dataJSON)),
	//         Attributes: keyvault.SecretAttributesArgs{
	//             Enabled: pulumi.Bool(true),
	//         },
	//     },
	// })
	
	// Temporary placeholder - return nil until fields are implemented
	sm.ctx.Log.Info(fmt.Sprintf("Azure Key Vault secret creation disabled - missing config fields: %s", secretName), nil)
	return nil, nil
}

func (sm *SecretsManager) createLocalSecret(secretName string, secretData map[string]string, metadata SecretMetadata) (pulumi.Resource, error) {
	// For local development, we log the secret creation
	// In practice, this might write to a local file or environment variable
	sm.ctx.Log.Info(fmt.Sprintf("Creating local secret: %s", secretName), nil)
	return nil, nil
}

func (sm *SecretsManager) createKubernetesSecret(secretName string, secretData map[string]string, metadata SecretMetadata) (pulumi.Resource, error) {
	// For Kubernetes, we would create a Secret resource
	// This is a placeholder for actual Kubernetes secret creation
	sm.ctx.Log.Info(fmt.Sprintf("Creating Kubernetes secret: %s", secretName), nil)
	return nil, nil
}

// Helper methods

func (sm *SecretsManager) determineSecretProvider(configManager *config.ConfigManager) SecretProvider {
	if configManager.GetEnvironment().IsDevelopment() {
		return SecretProviderVault // Use Vault even in development
	}
	
	// Use Vault for all environments through ConfigManager
	return SecretProviderVault
}

// GetSecretReference returns a reference to a secret for environment variables
func (sm *SecretsManager) GetSecretReference(secretName string) string {
	switch sm.provider {
	case SecretProviderVault:
		return fmt.Sprintf("vault:%s", secretName)
	case SecretProviderAzureKV:
		return fmt.Sprintf("azurekv:%s", secretName)
	case SecretProviderK8sSecret:
		return fmt.Sprintf("k8s-secret:%s", secretName)
	default:
		return fmt.Sprintf("local:%s", secretName)
	}
}

// RotateSecret initiates secret rotation
func (sm *SecretsManager) RotateSecret(secretName string) error {
	sm.ctx.Log.Info(fmt.Sprintf("Initiating rotation for secret: %s", secretName), nil)
	
	// Implementation would depend on the secret provider and rotation strategy
	switch sm.provider {
	case SecretProviderVault:
		return sm.rotateVaultSecret(secretName)
	case SecretProviderAzureKV:
		return sm.rotateAzureKeyVaultSecret(secretName)
	default:
		return fmt.Errorf("secret rotation not supported for provider: %s", sm.provider)
	}
}

func (sm *SecretsManager) rotateVaultSecret(secretName string) error {
	// Vault secret rotation implementation
	return fmt.Errorf("vault secret rotation not yet implemented")
}

func (sm *SecretsManager) rotateAzureKeyVaultSecret(secretName string) error {
	// Azure Key Vault secret rotation implementation
	return fmt.Errorf("azure key vault secret rotation not yet implemented")
}

// ValidateSecretConfiguration validates secret configuration
func (sm *SecretsManager) ValidateSecretConfiguration(config *SecretConfiguration) error {
	if config.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}
	
	// Validate secret names
	for _, secret := range config.Secrets {
		if secret.Name == "" {
			return fmt.Errorf("secret name is required")
		}
		
		if !sm.isValidSecretName(secret.Name) {
			return fmt.Errorf("invalid secret name: %s", secret.Name)
		}
	}
	
	return nil
}

func (sm *SecretsManager) isValidSecretName(name string) bool {
	// Basic validation - no spaces, special characters
	return !strings.ContainsAny(name, " \t\n\r") && name != ""
}

// GetEnvironmentSecrets returns all secrets for environment variables
func (sc *SecretConfiguration) GetEnvironmentSecrets() map[string]string {
	return sc.EnvSecrets
}

// GetMountedSecrets returns all secrets that should be mounted as files
func (sc *SecretConfiguration) GetMountedSecrets() map[string]string {
	return sc.SecretMounts
}