package testing

import (
	"fmt"
	"time"
	
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// MockProvider interface for infrastructure providers
type MockProvider interface {
	NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error)
	Call(args pulumi.MockCallArgs) (resource.PropertyMap, error)
	GetProviderName() string
	GetResourceTypes() []string
}

// InfrastructureMocks provides environment-specific infrastructure providers
type InfrastructureMocks struct {
	environment string
	providers   map[string]MockProvider
}

// DatabaseMockProvider provides database infrastructure behaviors
type DatabaseMockProvider struct {
	environment string
}

// StorageMockProvider provides storage infrastructure behaviors  
type StorageMockProvider struct {
	environment string
}

// VaultMockProvider provides vault infrastructure behaviors
type VaultMockProvider struct {
	environment string
}

// DaprMockProvider provides dapr infrastructure behaviors
type DaprMockProvider struct {
	environment string
}

// NewInfrastructureMocks creates environment-specific infrastructure providers
func NewInfrastructureMocks(environment string) *InfrastructureMocks {
	providers := make(map[string]MockProvider)
	providers["database"] = &DatabaseMockProvider{environment: environment}
	providers["storage"] = &StorageMockProvider{environment: environment}
	providers["vault"] = &VaultMockProvider{environment: environment}
	providers["dapr"] = &DaprMockProvider{environment: environment}
	
	return &InfrastructureMocks{
		environment: environment,
		providers:   providers,
	}
}

// NewResource implements pulumi.Mocks interface for infrastructure testing
func (m *InfrastructureMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	// Route to appropriate provider based on resource type
	switch {
	case contains(args.TypeToken, "postgresql") || contains(args.TypeToken, "database"):
		if provider, exists := m.providers["database"]; exists {
			return provider.NewResource(args)
		}
	case contains(args.TypeToken, "storage") || contains(args.TypeToken, "blob"):
		if provider, exists := m.providers["storage"]; exists {
			return provider.NewResource(args)
		}
	case contains(args.TypeToken, "keyvault") || contains(args.TypeToken, "vault"):
		if provider, exists := m.providers["vault"]; exists {
			return provider.NewResource(args)
		}
	case contains(args.TypeToken, "container") || contains(args.TypeToken, "podman") || contains(args.TypeToken, "file"):
		if provider, exists := m.providers["dapr"]; exists {
			return provider.NewResource(args)
		}
	}
	
	// Default fallback
	resourceID := fmt.Sprintf("%s-%s-%s", m.environment, args.Name, generateResourceSuffix())
	return resourceID, resource.PropertyMap{
		"name": resource.NewStringProperty(args.Name),
		"type": resource.NewStringProperty(args.TypeToken),
	}, nil
}

// Call implements pulumi.Mocks interface for infrastructure testing
func (m *InfrastructureMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	// Route to appropriate provider based on call token
	switch {
	case contains(args.Token, "postgresql") || contains(args.Token, "database"):
		if provider, exists := m.providers["database"]; exists {
			return provider.Call(args)
		}
	case contains(args.Token, "storage") || contains(args.Token, "blob"):
		if provider, exists := m.providers["storage"]; exists {
			return provider.Call(args)
		}
	case contains(args.Token, "keyvault") || contains(args.Token, "vault"):
		if provider, exists := m.providers["vault"]; exists {
			return provider.Call(args)
		}
	case contains(args.Token, "container") || contains(args.Token, "podman"):
		if provider, exists := m.providers["dapr"]; exists {
			return provider.Call(args)
		}
	}
	
	// Default fallback
	return resource.PropertyMap{}, nil
}

// Helper function for string containment check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		   len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		   (len(s) > len(substr) && findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// DatabaseMockProvider implementations
func (d *DatabaseMockProvider) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	resourceID := fmt.Sprintf("%s-%s-%s", d.environment, args.Name, generateResourceSuffix())
	
	switch args.TypeToken {
	case "azure:postgresql/server:Server":
		return resourceID, d.generatePostgreSQLServerProperties(args), nil
	case "azure:postgresql/database:Database":
		return resourceID, d.generatePostgreSQLDatabaseProperties(args), nil
	case "azure:postgresql/firewallRule:FirewallRule":
		return resourceID, d.generateFirewallRuleProperties(args), nil
	default:
		return resourceID, resource.PropertyMap{
			"name": resource.NewStringProperty(args.Name),
			"type": resource.NewStringProperty(args.TypeToken),
		}, nil
	}
}

func (d *DatabaseMockProvider) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	switch args.Token {
	case "azure:postgresql/getServer:getServer":
		return resource.PropertyMap{
			"name":                resource.NewStringProperty("test-server"),
			"fullyQualifiedDomainName": resource.NewStringProperty(fmt.Sprintf("%s-postgresql.postgres.database.azure.com", d.environment)),
			"version":             resource.NewStringProperty("13"),
		}, nil
	default:
		return resource.PropertyMap{}, nil
	}
}

func (d *DatabaseMockProvider) GetProviderName() string {
	return "database-provider"
}

func (d *DatabaseMockProvider) GetResourceTypes() []string {
	return []string{"azure:postgresql/server:Server", "azure:postgresql/database:Database", "azure:postgresql/firewallRule:FirewallRule"}
}

func (d *DatabaseMockProvider) generatePostgreSQLServerProperties(args pulumi.MockResourceArgs) resource.PropertyMap {
	props := resource.PropertyMap{
		"name":                     resource.NewStringProperty(args.Name),
		"administratorLogin":       resource.NewStringProperty("internationalcenteradmin"),
		"version":                 resource.NewStringProperty("13"),
		"sslEnforcement":          resource.NewStringProperty("Enabled"),
		"minimumTlsVersion":       resource.NewStringProperty("TLS1_2"),
		"fullyQualifiedDomainName": resource.NewStringProperty(fmt.Sprintf("%s-postgresql.postgres.database.azure.com", d.environment)),
	}
	
	// Environment-specific properties
	switch d.environment {
	case "production":
		props["skuName"] = resource.NewStringProperty("GP_Gen5_4")
		props["storageMb"] = resource.NewNumberProperty(512000)
		props["backupRetentionDays"] = resource.NewNumberProperty(35)
		props["geoRedundantBackupEnabled"] = resource.NewBoolProperty(true)
	case "staging":
		props["skuName"] = resource.NewStringProperty("GP_Gen5_2")
		props["storageMb"] = resource.NewNumberProperty(102400)
		props["backupRetentionDays"] = resource.NewNumberProperty(7)
		props["geoRedundantBackupEnabled"] = resource.NewBoolProperty(false)
	case "development":
		props["skuName"] = resource.NewStringProperty("B_Gen5_1")
		props["storageMb"] = resource.NewNumberProperty(51200)
		props["backupRetentionDays"] = resource.NewNumberProperty(7)
		props["geoRedundantBackupEnabled"] = resource.NewBoolProperty(false)
	}
	
	return props
}

func (d *DatabaseMockProvider) generatePostgreSQLDatabaseProperties(args pulumi.MockResourceArgs) resource.PropertyMap {
	return resource.PropertyMap{
		"name":     resource.NewStringProperty(args.Name),
		"charset":  resource.NewStringProperty("UTF8"),
		"collation": resource.NewStringProperty("English_United States.1252"),
	}
}

func (d *DatabaseMockProvider) generateFirewallRuleProperties(args pulumi.MockResourceArgs) resource.PropertyMap {
	return resource.PropertyMap{
		"name":           resource.NewStringProperty(args.Name),
		"startIpAddress": resource.NewStringProperty("10.0.0.0"),
		"endIpAddress":   resource.NewStringProperty("10.0.255.255"),
	}
}

// StorageMockProvider implementations
func (s *StorageMockProvider) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	resourceID := fmt.Sprintf("%s-%s-%s", s.environment, args.Name, generateResourceSuffix())
	
	switch args.TypeToken {
	case "azure:storage/account:Account":
		return resourceID, s.generateStorageAccountProperties(args), nil
	case "azure:storage/container:Container":
		return resourceID, s.generateStorageContainerProperties(args), nil
	case "azure:storage/queue:Queue":
		return resourceID, s.generateStorageQueueProperties(args), nil
	default:
		return resourceID, resource.PropertyMap{
			"name": resource.NewStringProperty(args.Name),
			"type": resource.NewStringProperty(args.TypeToken),
		}, nil
	}
}

func (s *StorageMockProvider) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	switch args.Token {
	case "azure:storage/getAccount:getAccount":
		return resource.PropertyMap{
			"name": resource.NewStringProperty("test-storage"),
			"primaryBlobEndpoint": resource.NewStringProperty(fmt.Sprintf("https://%sstorage.blob.core.windows.net/", s.environment)),
			"primaryQueueEndpoint": resource.NewStringProperty(fmt.Sprintf("https://%sstorage.queue.core.windows.net/", s.environment)),
		}, nil
	default:
		return resource.PropertyMap{}, nil
	}
}

func (s *StorageMockProvider) GetProviderName() string {
	return "storage-provider"
}

func (s *StorageMockProvider) GetResourceTypes() []string {
	return []string{"azure:storage/account:Account", "azure:storage/container:Container", "azure:storage/queue:Queue"}
}

func (s *StorageMockProvider) generateStorageAccountProperties(args pulumi.MockResourceArgs) resource.PropertyMap {
	props := resource.PropertyMap{
		"name":                    resource.NewStringProperty(args.Name),
		"kind":                   resource.NewStringProperty("StorageV2"),
		"accessTier":             resource.NewStringProperty("Hot"),
		"allowBlobPublicAccess":  resource.NewBoolProperty(false),
		"minimumTlsVersion":      resource.NewStringProperty("TLS1_2"),
		"primaryBlobEndpoint":    resource.NewStringProperty(fmt.Sprintf("https://%s.blob.core.windows.net/", args.Name)),
		"primaryQueueEndpoint":   resource.NewStringProperty(fmt.Sprintf("https://%s.queue.core.windows.net/", args.Name)),
		"primaryTableEndpoint":   resource.NewStringProperty(fmt.Sprintf("https://%s.table.core.windows.net/", args.Name)),
	}
	
	// Environment-specific properties
	switch s.environment {
	case "production":
		props["accountTier"] = resource.NewStringProperty("Standard")
		props["accountReplicationType"] = resource.NewStringProperty("GRS")
		props["allowSharedKeyAccess"] = resource.NewBoolProperty(false)
	case "staging":
		props["accountTier"] = resource.NewStringProperty("Standard")
		props["accountReplicationType"] = resource.NewStringProperty("LRS")
		props["allowSharedKeyAccess"] = resource.NewBoolProperty(true)
	case "development":
		props["accountTier"] = resource.NewStringProperty("Standard")
		props["accountReplicationType"] = resource.NewStringProperty("LRS")
		props["allowSharedKeyAccess"] = resource.NewBoolProperty(true)
	}
	
	return props
}

func (s *StorageMockProvider) generateStorageContainerProperties(args pulumi.MockResourceArgs) resource.PropertyMap {
	return resource.PropertyMap{
		"name":                 resource.NewStringProperty(args.Name),
		"containerAccessType":  resource.NewStringProperty("private"),
		"metadata":            resource.NewObjectProperty(resource.PropertyMap{
			"environment": resource.NewStringProperty(s.environment),
		}),
	}
}

func (s *StorageMockProvider) generateStorageQueueProperties(args pulumi.MockResourceArgs) resource.PropertyMap {
	return resource.PropertyMap{
		"name": resource.NewStringProperty(args.Name),
		"metadata": resource.NewObjectProperty(resource.PropertyMap{
			"environment": resource.NewStringProperty(s.environment),
			"purpose":     resource.NewStringProperty("async-processing"),
		}),
	}
}

// VaultMockProvider implementations
func (v *VaultMockProvider) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	resourceID := fmt.Sprintf("%s-%s-%s", v.environment, args.Name, generateResourceSuffix())
	
	switch args.TypeToken {
	case "azure:keyvault/vault:Vault":
		return resourceID, v.generateKeyVaultProperties(args), nil
	case "azure:keyvault/secret:Secret":
		return resourceID, v.generateKeyVaultSecretProperties(args), nil
	case "azure:keyvault/key:Key":
		return resourceID, v.generateKeyVaultKeyProperties(args), nil
	default:
		return resourceID, resource.PropertyMap{
			"name": resource.NewStringProperty(args.Name),
			"type": resource.NewStringProperty(args.TypeToken),
		}, nil
	}
}

func (v *VaultMockProvider) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	switch args.Token {
	case "azure:keyvault/getVault:getVault":
		return resource.PropertyMap{
			"name":    resource.NewStringProperty("test-vault"),
			"vaultUri": resource.NewStringProperty(fmt.Sprintf("https://%s-vault.vault.azure.net/", v.environment)),
		}, nil
	case "azure:keyvault/getSecret:getSecret":
		return resource.PropertyMap{
			"value": resource.NewSecretProperty(&resource.Secret{Element: resource.NewStringProperty("mock-secret-value")}),
		}, nil
	default:
		return resource.PropertyMap{}, nil
	}
}

func (v *VaultMockProvider) GetProviderName() string {
	return "vault-provider"
}

func (v *VaultMockProvider) GetResourceTypes() []string {
	return []string{"azure:keyvault/vault:Vault", "azure:keyvault/secret:Secret", "azure:keyvault/key:Key"}
}

func (v *VaultMockProvider) generateKeyVaultProperties(args pulumi.MockResourceArgs) resource.PropertyMap {
	props := resource.PropertyMap{
		"name":                          resource.NewStringProperty(args.Name),
		"vaultUri":                     resource.NewStringProperty(fmt.Sprintf("https://%s.vault.azure.net/", args.Name)),
		"enabledForDeployment":         resource.NewBoolProperty(true),
		"enabledForTemplateDeployment": resource.NewBoolProperty(true),
		"enabledForDiskEncryption":     resource.NewBoolProperty(true),
		"enableSoftDelete":             resource.NewBoolProperty(true),
		"softDeleteRetentionInDays":    resource.NewNumberProperty(90),
	}
	
	// Environment-specific properties
	switch v.environment {
	case "production":
		props["enablePurgeProtection"] = resource.NewBoolProperty(true)
		props["networkAcls"] = resource.NewObjectProperty(resource.PropertyMap{
			"defaultAction": resource.NewStringProperty("Deny"),
			"bypass":        resource.NewStringProperty("AzureServices"),
		})
	case "staging":
		props["enablePurgeProtection"] = resource.NewBoolProperty(false)
		props["networkAcls"] = resource.NewObjectProperty(resource.PropertyMap{
			"defaultAction": resource.NewStringProperty("Allow"),
		})
	case "development":
		props["enablePurgeProtection"] = resource.NewBoolProperty(false)
		props["networkAcls"] = resource.NewObjectProperty(resource.PropertyMap{
			"defaultAction": resource.NewStringProperty("Allow"),
		})
	}
	
	return props
}

func (v *VaultMockProvider) generateKeyVaultSecretProperties(args pulumi.MockResourceArgs) resource.PropertyMap {
	return resource.PropertyMap{
		"name":  resource.NewStringProperty(args.Name),
		"value": resource.NewSecretProperty(&resource.Secret{Element: resource.NewStringProperty("mock-secret-value")}),
		"contentType": resource.NewStringProperty("text/plain"),
	}
}

func (v *VaultMockProvider) generateKeyVaultKeyProperties(args pulumi.MockResourceArgs) resource.PropertyMap {
	return resource.PropertyMap{
		"name":     resource.NewStringProperty(args.Name),
		"keyType":  resource.NewStringProperty("RSA"),
		"keySize":  resource.NewNumberProperty(2048),
		"keyOpts": resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty("encrypt"),
			resource.NewStringProperty("decrypt"),
			resource.NewStringProperty("sign"),
			resource.NewStringProperty("verify"),
			resource.NewStringProperty("wrapKey"),
			resource.NewStringProperty("unwrapKey"),
		}),
	}
}

// DaprMockProvider implementations
func (d *DaprMockProvider) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	resourceID := fmt.Sprintf("%s-%s-%s", d.environment, args.Name, generateResourceSuffix())
	
	switch args.TypeToken {
	case "podman:container/container:Container":
		return resourceID, d.generateContainerProperties(args), nil
	case "local:file/file:File":
		return resourceID, d.generateFileProperties(args), nil
	default:
		return resourceID, resource.PropertyMap{
			"name": resource.NewStringProperty(args.Name),
			"type": resource.NewStringProperty(args.TypeToken),
		}, nil
	}
}

func (d *DaprMockProvider) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	switch args.Token {
	case "podman:container/getContainer:getContainer":
		return resource.PropertyMap{
			"name": resource.NewStringProperty("test-container"),
			"state": resource.NewStringProperty("running"),
		}, nil
	default:
		return resource.PropertyMap{}, nil
	}
}

func (d *DaprMockProvider) GetProviderName() string {
	return "dapr-provider"
}

func (d *DaprMockProvider) GetResourceTypes() []string {
	return []string{"podman:container/container:Container", "local:file/file:File"}
}

func (d *DaprMockProvider) generateContainerProperties(args pulumi.MockResourceArgs) resource.PropertyMap {
	props := resource.PropertyMap{
		"name":  resource.NewStringProperty(args.Name),
		"state": resource.NewStringProperty("running"),
		"ports": resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty("3500:3500"), // Dapr HTTP port
			resource.NewStringProperty("50001:50001"), // Dapr GRPC port
		}),
		"environment": resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty(fmt.Sprintf("DAPR_ENVIRONMENT=%s", d.environment)),
		}),
	}
	
	// Environment-specific container properties
	switch d.environment {
	case "development":
		props["image"] = resource.NewStringProperty("daprio/daprd:latest")
		props["restart"] = resource.NewStringProperty("unless-stopped")
	case "staging":
		props["image"] = resource.NewStringProperty("daprio/daprd:1.12.0")
		props["restart"] = resource.NewStringProperty("always")
	case "production":
		props["image"] = resource.NewStringProperty("daprio/daprd:1.12.0")
		props["restart"] = resource.NewStringProperty("always")
		props["memoryLimit"] = resource.NewStringProperty("512m")
		props["cpuLimit"] = resource.NewStringProperty("0.5")
	}
	
	return props
}

func (d *DaprMockProvider) generateFileProperties(args pulumi.MockResourceArgs) resource.PropertyMap {
	return resource.PropertyMap{
		"filename": resource.NewStringProperty(args.Name),
		"content":  resource.NewStringProperty("# Dapr component configuration\napiVersion: dapr.io/v1alpha1\nkind: Component"),
		"filePermissions": resource.NewStringProperty("0644"),
	}
}

// Utility function to generate unique resource suffixes
func generateResourceSuffix() string {
	return fmt.Sprintf("%d", time.Now().Unix()%10000)
}