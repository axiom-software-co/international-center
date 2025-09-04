package infrastructure

import (
	"context"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type VaultStack interface {
	Deploy(ctx context.Context) (VaultDeployment, error)
	GetVaultConnectionInfo() (string, string, error)
	InitializeSecrets(ctx context.Context, deployment VaultDeployment) error
	CreatePolicies(ctx context.Context, deployment VaultDeployment) error
	ValidateDeployment(ctx context.Context, deployment VaultDeployment) error
}

type VaultDeployment interface {
	GetVaultEndpoint() pulumi.StringOutput
	GetUIEndpoint() pulumi.StringOutput
	GetRootToken() pulumi.StringOutput
	GetUnsealKeys() []pulumi.StringOutput
	GetServiceAccount() string
}

type VaultConfiguration struct {
	Environment        string
	DeploymentMode     string // "dev", "standalone", "ha", "cloud"
	StorageBackend     string // "file", "consul", "azure", "raft"
	UIEnabled          bool
	APIAddr            string
	ClusterAddr        string
	DisableMlock       bool
	DefaultLeaseTTL    string
	MaxLeaseTTL        string
	EnableAuditLog     bool
	AuditLogPath       string
	SecretsEngines     []SecretsEngine
	AuthMethods        []AuthMethod
	Policies           []Policy
	TLSConfig          TLSConfiguration
	HAConfig           HAConfiguration
	SealConfig         SealConfiguration
}

type SecretsEngine struct {
	Path        string
	Type        string // "kv", "database", "azure", "kubernetes"
	Version     string
	Description string
	Config      map[string]interface{}
	Options     map[string]interface{}
}

type AuthMethod struct {
	Path        string
	Type        string // "userpass", "jwt", "azure", "kubernetes"
	Description string
	Config      map[string]interface{}
}

type Policy struct {
	Name        string
	Description string
	Rules       string
}

type TLSConfiguration struct {
	Enabled          bool
	CertFile         string
	KeyFile          string
	CAFile           string
	MinVersion       string
	CipherSuites     []string
	PreferServerCiphers bool
}

type HAConfiguration struct {
	Enabled           bool
	DisableClustering bool
	ClusterName       string
	APIAddr           string
	ClusterAddr       string
	RedirectAddr      string
}

type SealConfiguration struct {
	Type           string // "shamir", "azure-keyvault", "aws-kms", "gcpckms"
	KeyName        string
	VaultName      string
	ClientID       string
	ClientSecret   string
	TenantID       string
	ResourceGroup  string
	SubscriptionID string
}

type VaultSecrets struct {
	DatabaseCredentials map[string]DatabaseCredential
	APIKeys            map[string]string
	Certificates       map[string]Certificate
	ServiceAccounts    map[string]ServiceAccount
}

type DatabaseCredential struct {
	Username string
	Password string
	Host     string
	Port     int
	Database string
	TTL      string
}

type Certificate struct {
	Certificate  string
	PrivateKey   string
	CAChain      []string
	SerialNumber string
	ExpiryDate   string
}

type ServiceAccount struct {
	ClientID     string
	ClientSecret string
	TenantID     string
	Scope        []string
}

type VaultFactory interface {
	CreateVaultStack(ctx *pulumi.Context, config *config.Config, environment string) VaultStack
}

func GetVaultConfiguration(environment string, config *config.Config) *VaultConfiguration {
	switch environment {
	case "development":
		return &VaultConfiguration{
			Environment:     "development",
			DeploymentMode:  "dev",
			StorageBackend:  "file",
			UIEnabled:       true,
			APIAddr:         "http://127.0.0.1:8200",
			ClusterAddr:     "https://127.0.0.1:8201",
			DisableMlock:    true,
			DefaultLeaseTTL: "768h",
			MaxLeaseTTL:     "8760h",
			EnableAuditLog:  false,
			AuditLogPath:    "/vault/logs/audit.log",
			SecretsEngines: []SecretsEngine{
				{
					Path:        "secret/",
					Type:        "kv",
					Version:     "v2",
					Description: "Key-Value secrets engine",
				},
				{
					Path:        "database/",
					Type:        "database",
					Version:     "v1",
					Description: "Database credentials",
				},
			},
			AuthMethods: []AuthMethod{
				{
					Path:        "userpass/",
					Type:        "userpass",
					Description: "Username & Password auth",
				},
			},
			TLSConfig: TLSConfiguration{
				Enabled: false,
			},
			HAConfig: HAConfiguration{
				Enabled: false,
			},
			SealConfig: SealConfiguration{
				Type: "shamir",
			},
		}
	case "staging":
		return &VaultConfiguration{
			Environment:     "staging",
			DeploymentMode:  "standalone",
			StorageBackend:  "azure",
			UIEnabled:       true,
			APIAddr:         "https://vault-staging.international-center.com",
			ClusterAddr:     "https://vault-staging.international-center.com:8201",
			DisableMlock:    false,
			DefaultLeaseTTL: "168h", // 7 days
			MaxLeaseTTL:     "720h", // 30 days
			EnableAuditLog:  true,
			AuditLogPath:    "/vault/logs/audit.log",
			SecretsEngines: []SecretsEngine{
				{
					Path:        "secret/",
					Type:        "kv",
					Version:     "v2",
					Description: "Application secrets",
				},
				{
					Path:        "database/",
					Type:        "database",
					Version:     "v1",
					Description: "Database dynamic credentials",
				},
				{
					Path:        "azure/",
					Type:        "azure",
					Version:     "v1",
					Description: "Azure service principal credentials",
				},
			},
			AuthMethods: []AuthMethod{
				{
					Path:        "jwt/",
					Type:        "jwt",
					Description: "JWT/OIDC authentication",
				},
				{
					Path:        "azure/",
					Type:        "azure",
					Description: "Azure Active Directory authentication",
				},
			},
			TLSConfig: TLSConfiguration{
				Enabled:    true,
				MinVersion: "tls12",
				PreferServerCiphers: true,
			},
			HAConfig: HAConfiguration{
				Enabled:     false,
				ClusterName: "staging-cluster",
			},
			SealConfig: SealConfiguration{
				Type:      "azure-keyvault",
				KeyName:   "vault-unseal-key",
				VaultName: "int-staging-keyvault",
			},
		}
	case "production":
		return &VaultConfiguration{
			Environment:     "production",
			DeploymentMode:  "ha",
			StorageBackend:  "raft",
			UIEnabled:       true,
			APIAddr:         "https://vault.international-center.com",
			ClusterAddr:     "https://vault.international-center.com:8201",
			DisableMlock:    false,
			DefaultLeaseTTL: "24h",
			MaxLeaseTTL:     "168h", // 7 days
			EnableAuditLog:  true,
			AuditLogPath:    "/vault/logs/audit.log",
			SecretsEngines: []SecretsEngine{
				{
					Path:        "secret/",
					Type:        "kv",
					Version:     "v2",
					Description: "Application secrets",
				},
				{
					Path:        "database/",
					Type:        "database",
					Version:     "v1",
					Description: "Database dynamic credentials",
				},
				{
					Path:        "azure/",
					Type:        "azure",
					Version:     "v1",
					Description: "Azure service principal credentials",
				},
				{
					Path:        "pki/",
					Type:        "pki",
					Version:     "v1",
					Description: "Public Key Infrastructure",
				},
			},
			AuthMethods: []AuthMethod{
				{
					Path:        "jwt/",
					Type:        "jwt",
					Description: "JWT/OIDC authentication",
				},
				{
					Path:        "azure/",
					Type:        "azure",
					Description: "Azure Active Directory authentication",
				},
				{
					Path:        "kubernetes/",
					Type:        "kubernetes",
					Description: "Kubernetes service account authentication",
				},
			},
			TLSConfig: TLSConfiguration{
				Enabled:    true,
				MinVersion: "tls12",
				PreferServerCiphers: true,
			},
			HAConfig: HAConfiguration{
				Enabled:           true,
				DisableClustering: false,
				ClusterName:       "production-cluster",
			},
			SealConfig: SealConfiguration{
				Type:      "azure-keyvault",
				KeyName:   "vault-unseal-key",
				VaultName: "int-prod-keyvault",
			},
		}
	default:
		return &VaultConfiguration{
			Environment:     environment,
			DeploymentMode:  "standalone",
			StorageBackend:  "file",
			UIEnabled:       true,
			APIAddr:         "http://127.0.0.1:8200",
			ClusterAddr:     "https://127.0.0.1:8201",
			DisableMlock:    true,
			DefaultLeaseTTL: "768h",
			MaxLeaseTTL:     "8760h",
			EnableAuditLog:  false,
			SecretsEngines: []SecretsEngine{
				{
					Path:        "secret/",
					Type:        "kv",
					Version:     "v2",
					Description: "Key-Value secrets engine",
				},
			},
			AuthMethods: []AuthMethod{
				{
					Path:        "userpass/",
					Type:        "userpass",
					Description: "Username & Password auth",
				},
			},
			TLSConfig: TLSConfiguration{
				Enabled: false,
			},
			HAConfig: HAConfiguration{
				Enabled: false,
			},
			SealConfig: SealConfiguration{
				Type: "shamir",
			},
		}
	}
}

func GetVaultPolicies(environment string) []Policy {
	basePolicies := []Policy{
		{
			Name:        "content-api-policy",
			Description: "Policy for content API service",
			Rules: `
path "secret/data/content-api/*" {
  capabilities = ["read", "list"]
}
path "database/creds/content-readonly" {
  capabilities = ["read"]
}`,
		},
		{
			Name:        "services-api-policy",
			Description: "Policy for services API service",
			Rules: `
path "secret/data/services-api/*" {
  capabilities = ["read", "list"]
}
path "database/creds/services-readwrite" {
  capabilities = ["read"]
}`,
		},
	}

	switch environment {
	case "development":
		basePolicies = append(basePolicies, Policy{
			Name:        "dev-admin-policy",
			Description: "Admin policy for development environment",
			Rules: `
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "database/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}`,
		})
	case "staging":
		basePolicies = append(basePolicies, Policy{
			Name:        "staging-deploy-policy",
			Description: "Deployment policy for staging environment",
			Rules: `
path "secret/data/staging/*" {
  capabilities = ["read", "list"]
}
path "azure/creds/staging-sp" {
  capabilities = ["read"]
}`,
		})
	case "production":
		basePolicies = append(basePolicies, Policy{
			Name:        "prod-deploy-policy",
			Description: "Deployment policy for production environment",
			Rules: `
path "secret/data/production/*" {
  capabilities = ["read", "list"]
}
path "azure/creds/production-sp" {
  capabilities = ["read"]
}
path "pki/issue/production" {
  capabilities = ["create", "update"]
}`,
		})
	}

	return basePolicies
}

// VaultMetrics defines security and performance metrics for environment-specific policies
type VaultMetrics struct {
	MaxSecretsCount        int
	MaxTokenTTL           string
	EnableAuditLog        bool
	RequireApprovalFlow   bool
	RotationIntervalDays  int
	BackupIntervalHours   int
	ComplianceRequired    bool
}

func GetVaultMetrics(environment string) VaultMetrics {
	switch environment {
	case "development":
		return VaultMetrics{
			MaxSecretsCount:       1000,
			MaxTokenTTL:          "8760h", // 1 year
			EnableAuditLog:       false,
			RequireApprovalFlow:  false,
			RotationIntervalDays: 90,
			BackupIntervalHours:  24,
			ComplianceRequired:   false,
		}
	case "staging":
		return VaultMetrics{
			MaxSecretsCount:       5000,
			MaxTokenTTL:          "720h", // 30 days
			EnableAuditLog:       true,
			RequireApprovalFlow:  true,
			RotationIntervalDays: 30,
			BackupIntervalHours:  12,
			ComplianceRequired:   true,
		}
	case "production":
		return VaultMetrics{
			MaxSecretsCount:       10000,
			MaxTokenTTL:          "168h", // 7 days
			EnableAuditLog:       true,
			RequireApprovalFlow:  true,
			RotationIntervalDays: 7,
			BackupIntervalHours:  4,
			ComplianceRequired:   true,
		}
	default:
		return VaultMetrics{
			MaxSecretsCount:       2000,
			MaxTokenTTL:          "720h",
			EnableAuditLog:       false,
			RequireApprovalFlow:  false,
			RotationIntervalDays: 60,
			BackupIntervalHours:  24,
			ComplianceRequired:   false,
		}
	}
}