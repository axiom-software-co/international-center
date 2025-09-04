package infrastructure

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-azure-native-sdk/keyvault"
	"github.com/pulumi/pulumi-azure-native-sdk/network"
	"github.com/pulumi/pulumi-azure-native-sdk/resources"
	"github.com/pulumi/pulumi-azure-native-sdk/security"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type VaultProductionStack struct {
	resourceGroup       *resources.ResourceGroup
	vnet               *network.VirtualNetwork
	privateSubnet      *network.Subnet
	keyVault           *keyvault.Vault
	secrets            map[string]*keyvault.Secret
	accessPolicies     []*keyvault.AccessPolicyEntry
	certificates       map[string]*keyvault.Certificate
	keys               map[string]*keyvault.Key
	privateEndpoint    *network.PrivateEndpoint
	privateDnsZone     *network.PrivateZone
	securityAssessment *security.Assessment
	backupVault        *keyvault.Vault
}

func NewVaultProductionStack(resourceGroup *resources.ResourceGroup, vnet *network.VirtualNetwork, privateSubnet *network.Subnet) *VaultProductionStack {
	return &VaultProductionStack{
		resourceGroup: resourceGroup,
		vnet:         vnet,
		privateSubnet: privateSubnet,
		secrets:      make(map[string]*keyvault.Secret),
		certificates: make(map[string]*keyvault.Certificate),
		keys:         make(map[string]*keyvault.Key),
	}
}

func (stack *VaultProductionStack) Deploy(ctx *pulumi.Context) error {
	if err := stack.createPrivateDnsZone(ctx); err != nil {
		return fmt.Errorf("failed to create private DNS zone: %w", err)
	}

	if err := stack.createProductionKeyVault(ctx); err != nil {
		return fmt.Errorf("failed to create production key vault: %w", err)
	}

	if err := stack.createBackupKeyVault(ctx); err != nil {
		return fmt.Errorf("failed to create backup key vault: %w", err)
	}

	if err := stack.createPrivateEndpoint(ctx); err != nil {
		return fmt.Errorf("failed to create private endpoint: %w", err)
	}

	if err := stack.configureProductionAccessPolicies(ctx); err != nil {
		return fmt.Errorf("failed to configure access policies: %w", err)
	}

	if err := stack.createProductionSecrets(ctx); err != nil {
		return fmt.Errorf("failed to create production secrets: %w", err)
	}

	if err := stack.createProductionCertificates(ctx); err != nil {
		return fmt.Errorf("failed to create production certificates: %w", err)
	}

	if err := stack.createProductionKeys(ctx); err != nil {
		return fmt.Errorf("failed to create production keys: %w", err)
	}

	if err := stack.enableSecurityAssessment(ctx); err != nil {
		return fmt.Errorf("failed to enable security assessment: %w", err)
	}

	return nil
}

func (stack *VaultProductionStack) createPrivateDnsZone(ctx *pulumi.Context) error {
	privateDnsZone, err := network.NewPrivateZone(ctx, "production-keyvault-dns-zone", &network.PrivateZoneArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		PrivateZoneName:   pulumi.String("privatelink.vaultcore.azure.net"),
		Location:         pulumi.String("Global"),
		Tags: pulumi.StringMap{
			"environment": pulumi.String("production"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	vnetLink, err := network.NewVirtualNetworkLink(ctx, "production-keyvault-vnet-link", &network.VirtualNetworkLinkArgs{
		ResourceGroupName:      stack.resourceGroup.Name,
		PrivateZoneName:        privateDnsZone.Name,
		VirtualNetworkLinkName: pulumi.String("production-keyvault-vnet-link"),
		Location:              pulumi.String("Global"),
		VirtualNetwork: &network.SubResourceArgs{
			Id: stack.vnet.ID(),
		},
		RegistrationEnabled: pulumi.Bool(false),
		Tags: pulumi.StringMap{
			"environment": pulumi.String("production"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}
	_ = vnetLink

	stack.privateDnsZone = privateDnsZone
	return nil
}

func (stack *VaultProductionStack) createProductionKeyVault(ctx *pulumi.Context) error {
	keyVault, err := keyvault.NewVault(ctx, "production-keyvault", &keyvault.VaultArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		VaultName:        pulumi.String("international-center-production-kv"),
		Location:         stack.resourceGroup.Location,
		Properties: &keyvault.VaultPropertiesArgs{
			TenantId: pulumi.String(""), // From environment
			Sku: &keyvault.SkuArgs{
				Name:   pulumi.String("Premium"), // Premium for HSM support
				Family: pulumi.String("A"),
			},
			EnabledForDeployment:         pulumi.Bool(true),
			EnabledForTemplateDeployment: pulumi.Bool(true),
			EnabledForDiskEncryption:     pulumi.Bool(true),
			EnableRbacAuthorization:      pulumi.Bool(false),
			SoftDeleteRetentionInDays:    pulumi.Int(90),
			EnableSoftDelete:            pulumi.Bool(true),
			EnablePurgeProtection:       pulumi.Bool(true), // Required for production
			PublicNetworkAccess:         pulumi.String("Disabled"), // Private access only
			NetworkAcls: &keyvault.NetworkRuleSetArgs{
				Bypass:                pulumi.String("AzureServices"),
				DefaultAction:         pulumi.String("Deny"),
				VirtualNetworkRules: keyvault.VirtualNetworkRuleArray{
					&keyvault.VirtualNetworkRuleArgs{
						Id: stack.vnet.ID(),
					},
				},
			},
		},
		Tags: pulumi.StringMap{
			"environment":      pulumi.String("production"),
			"project":         pulumi.String("international-center"),
			"tier":           pulumi.String("vault"),
			"compliance":     pulumi.String("required"),
			"backup-required": pulumi.String("true"),
			"hsm-enabled":    pulumi.String("true"),
		},
	})
	if err != nil {
		return err
	}

	stack.keyVault = keyVault
	return nil
}

func (stack *VaultProductionStack) createBackupKeyVault(ctx *pulumi.Context) error {
	backupVault, err := keyvault.NewVault(ctx, "production-backup-keyvault", &keyvault.VaultArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		VaultName:        pulumi.String("international-center-prod-backup-kv"),
		Location:         pulumi.String("West US 2"), // Different region for backup
		Properties: &keyvault.VaultPropertiesArgs{
			TenantId: pulumi.String(""), // From environment
			Sku: &keyvault.SkuArgs{
				Name:   pulumi.String("Premium"),
				Family: pulumi.String("A"),
			},
			EnabledForDeployment:         pulumi.Bool(false), // Backup vault - restricted
			EnabledForTemplateDeployment: pulumi.Bool(false),
			EnabledForDiskEncryption:     pulumi.Bool(false),
			EnableRbacAuthorization:      pulumi.Bool(false),
			SoftDeleteRetentionInDays:    pulumi.Int(90),
			EnableSoftDelete:            pulumi.Bool(true),
			EnablePurgeProtection:       pulumi.Bool(true),
			PublicNetworkAccess:         pulumi.String("Disabled"),
			NetworkAcls: &keyvault.NetworkRuleSetArgs{
				Bypass:        pulumi.String("None"), // No bypass for backup vault
				DefaultAction: pulumi.String("Deny"),
			},
		},
		Tags: pulumi.StringMap{
			"environment":      pulumi.String("production"),
			"project":         pulumi.String("international-center"),
			"tier":           pulumi.String("vault-backup"),
			"role":           pulumi.String("backup"),
			"compliance":     pulumi.String("required"),
		},
	})
	if err != nil {
		return err
	}

	stack.backupVault = backupVault
	return nil
}

func (stack *VaultProductionStack) createPrivateEndpoint(ctx *pulumi.Context) error {
	privateEndpoint, err := network.NewPrivateEndpoint(ctx, "production-keyvault-private-endpoint", &network.PrivateEndpointArgs{
		ResourceGroupName:   stack.resourceGroup.Name,
		PrivateEndpointName: pulumi.String("international-center-production-kv-pe"),
		Location:           stack.resourceGroup.Location,
		Subnet: &network.SubnetArgs{
			Id: stack.privateSubnet.ID(),
		},
		PrivateLinkServiceConnections: network.PrivateLinkServiceConnectionArray{
			&network.PrivateLinkServiceConnectionArgs{
				Name:                 pulumi.String("keyvault-connection"),
				PrivateLinkServiceId: stack.keyVault.ID(),
				GroupIds: pulumi.StringArray{
					pulumi.String("vault"),
				},
			},
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("production"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	privateDnsZoneGroup, err := network.NewPrivateZoneGroup(ctx, "production-keyvault-dns-zone-group", &network.PrivateZoneGroupArgs{
		ResourceGroupName:       stack.resourceGroup.Name,
		PrivateEndpointName:     privateEndpoint.Name,
		PrivateDnsZoneGroupName: pulumi.String("default"),
		PrivateDnsZoneConfigs: network.PrivateDnsZoneConfigArray{
			&network.PrivateDnsZoneConfigArgs{
				Name: pulumi.String("keyvault-config"),
				PrivateDnsZoneId: stack.privateDnsZone.ID(),
			},
		},
	})
	if err != nil {
		return err
	}
	_ = privateDnsZoneGroup

	stack.privateEndpoint = privateEndpoint
	return nil
}

func (stack *VaultProductionStack) configureProductionAccessPolicies(ctx *pulumi.Context) error {
	containerAppsPolicy := &keyvault.AccessPolicyEntry{
		TenantId: pulumi.String(""), // From environment
		ObjectId: pulumi.String(""), // Container Apps managed identity
		Permissions: &keyvault.PermissionsArgs{
			Secrets: pulumi.StringArray{
				pulumi.String("get"),
				pulumi.String("list"),
			},
			Certificates: pulumi.StringArray{
				pulumi.String("get"),
				pulumi.String("list"),
			},
			Keys: pulumi.StringArray{
				pulumi.String("get"),
				pulumi.String("decrypt"),
				pulumi.String("encrypt"),
				pulumi.String("sign"),
				pulumi.String("verify"),
				pulumi.String("wrapKey"),
				pulumi.String("unwrapKey"),
			},
		},
	}

	deployerServicePrincipalPolicy := &keyvault.AccessPolicyEntry{
		TenantId: pulumi.String(""), // From environment
		ObjectId: pulumi.String(""), // Deployer service principal
		Permissions: &keyvault.PermissionsArgs{
			Secrets: pulumi.StringArray{
				pulumi.String("get"),
				pulumi.String("list"),
				pulumi.String("set"),
				pulumi.String("delete"),
				pulumi.String("recover"),
				pulumi.String("backup"),
				pulumi.String("restore"),
			},
			Certificates: pulumi.StringArray{
				pulumi.String("get"),
				pulumi.String("list"),
				pulumi.String("create"),
				pulumi.String("update"),
				pulumi.String("delete"),
				pulumi.String("import"),
				pulumi.String("backup"),
				pulumi.String("restore"),
				pulumi.String("recover"),
			},
			Keys: pulumi.StringArray{
				pulumi.String("get"),
				pulumi.String("list"),
				pulumi.String("create"),
				pulumi.String("update"),
				pulumi.String("delete"),
				pulumi.String("decrypt"),
				pulumi.String("encrypt"),
				pulumi.String("sign"),
				pulumi.String("verify"),
				pulumi.String("wrapKey"),
				pulumi.String("unwrapKey"),
				pulumi.String("backup"),
				pulumi.String("restore"),
				pulumi.String("recover"),
			},
		},
	}

	backupServicePolicy := &keyvault.AccessPolicyEntry{
		TenantId: pulumi.String(""), // From environment
		ObjectId: pulumi.String(""), // Backup service principal
		Permissions: &keyvault.PermissionsArgs{
			Secrets: pulumi.StringArray{
				pulumi.String("backup"),
				pulumi.String("restore"),
			},
			Certificates: pulumi.StringArray{
				pulumi.String("backup"),
				pulumi.String("restore"),
			},
			Keys: pulumi.StringArray{
				pulumi.String("backup"),
				pulumi.String("restore"),
			},
		},
	}

	stack.accessPolicies = append(stack.accessPolicies, containerAppsPolicy, deployerServicePrincipalPolicy, backupServicePolicy)
	return nil
}

func (stack *VaultProductionStack) createProductionSecrets(ctx *pulumi.Context) error {
	secretConfigs := map[string]string{
		"database-admin-password":            "", // Generated high-entropy
		"database-read-replica-password":     "", // Generated high-entropy
		"redis-connection-string":           "", // From Azure Redis
		"storage-account-access-key":        "", // From Azure Storage
		"backup-storage-access-key":         "", // From backup Azure Storage
		"grafana-api-key":                   "", // From Grafana Cloud
		"prometheus-api-key":                "", // From Grafana Cloud  
		"loki-api-key":                      "", // From Grafana Cloud
		"tempo-api-key":                     "", // From Grafana Cloud
		"azure-client-secret":               "", // For service principal
		"jwt-signing-key":                   "", // Generated high-entropy
		"encryption-key":                    "", // Generated high-entropy
		"smtp-password":                     "", // For email service
		"webhook-secret":                    "", // For GitHub webhooks
		"api-keys-external-service-primary": "", // External integrations
		"api-keys-external-service-backup":  "", // External integrations backup
		"audit-signing-key":                 "", // For audit integrity
		"compliance-encryption-key":        "", // For compliance data
		"disaster-recovery-key":             "", // For DR procedures
	}

	for secretName, secretValue := range secretConfigs {
		if err := stack.createProductionSecret(ctx, secretName, secretValue); err != nil {
			return fmt.Errorf("failed to create secret %s: %w", secretName, err)
		}
	}

	return nil
}

func (stack *VaultProductionStack) createProductionSecret(ctx *pulumi.Context, secretName, secretValue string) error {
	secret, err := keyvault.NewSecret(ctx, fmt.Sprintf("production-%s-secret", secretName), &keyvault.SecretArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		VaultName:        stack.keyVault.Name,
		SecretName:       pulumi.String(secretName),
		Properties: &keyvault.SecretPropertiesArgs{
			Value:       pulumi.String(secretValue),
			ContentType: pulumi.String("text/plain"),
			Attributes: &keyvault.SecretAttributesArgs{
				Enabled:   pulumi.Bool(true),
				NotBefore: pulumi.Int(0),
				// Expires: set based on secret type and compliance requirements
			},
		},
		Tags: pulumi.StringMap{
			"environment":      pulumi.String("production"),
			"project":         pulumi.String("international-center"),
			"compliance":     pulumi.String("required"),
			"backup-required": pulumi.String("true"),
		},
	})
	if err != nil {
		return err
	}

	stack.secrets[secretName] = secret
	return nil
}

func (stack *VaultProductionStack) createProductionCertificates(ctx *pulumi.Context) error {
	certificateConfigs := []string{
		"api-international-center-com",
		"admin-api-international-center-com", 
		"app-international-center-com",
		"www-international-center-com",
		"wildcard-international-center-com",
	}

	for _, certName := range certificateConfigs {
		if err := stack.createProductionCertificate(ctx, certName); err != nil {
			return fmt.Errorf("failed to create certificate %s: %w", certName, err)
		}
	}

	return nil
}

func (stack *VaultProductionStack) createProductionCertificate(ctx *pulumi.Context, certName string) error {
	certificate, err := keyvault.NewCertificate(ctx, fmt.Sprintf("production-%s-cert", certName), &keyvault.CertificateArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		VaultName:        stack.keyVault.Name,
		CertificateName:  pulumi.String(certName),
		Properties: &keyvault.CertificatePropertiesArgs{
			IssuerParameters: &keyvault.IssuerParametersArgs{
				Name: pulumi.String("DigiCert"), // Production CA
			},
			KeyProperties: &keyvault.KeyPropertiesArgs{
				Exportable: pulumi.Bool(false), // Non-exportable for production security
				KeySize:    pulumi.Int(4096),   // Larger key size for production
				KeyType:    pulumi.String("RSA"),
				ReuseKey:   pulumi.Bool(false),
			},
			SecretProperties: &keyvault.SecretPropertiesArgs{
				ContentType: pulumi.String("application/x-pkcs12"),
			},
			X509CertificateProperties: &keyvault.X509CertificatePropertiesArgs{
				Subject: pulumi.String(fmt.Sprintf("CN=%s", certName)),
				ValidityInMonths: pulumi.Int(24), // 2 year validity for production
				KeyUsage: pulumi.StringArray{
					pulumi.String("digitalSignature"),
					pulumi.String("keyEncipherment"),
					pulumi.String("keyAgreement"),
				},
				Ekus: pulumi.StringArray{
					pulumi.String("1.3.6.1.5.5.7.3.1"), // Server Authentication
					pulumi.String("1.3.6.1.5.5.7.3.2"), // Client Authentication
				},
				SubjectAlternativeNames: &keyvault.SubjectAlternativeNamesArgs{
					DnsNames: pulumi.StringArray{
						pulumi.String(certName),
					},
				},
			},
		},
		Tags: pulumi.StringMap{
			"environment":      pulumi.String("production"),
			"project":         pulumi.String("international-center"),
			"compliance":     pulumi.String("required"),
			"backup-required": pulumi.String("true"),
		},
	})
	if err != nil {
		return err
	}

	stack.certificates[certName] = certificate
	return nil
}

func (stack *VaultProductionStack) createProductionKeys(ctx *pulumi.Context) error {
	keyConfigs := []string{
		"data-encryption-key",
		"jwt-signing-key",
		"api-signing-key",
		"audit-signing-key",
		"compliance-encryption-key",
		"disaster-recovery-key",
	}

	for _, keyName := range keyConfigs {
		if err := stack.createProductionKey(ctx, keyName); err != nil {
			return fmt.Errorf("failed to create key %s: %w", keyName, err)
		}
	}

	return nil
}

func (stack *VaultProductionStack) createProductionKey(ctx *pulumi.Context, keyName string) error {
	key, err := keyvault.NewKey(ctx, fmt.Sprintf("production-%s-key", keyName), &keyvault.KeyArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		VaultName:        stack.keyVault.Name,
		KeyName:          pulumi.String(keyName),
		Properties: &keyvault.KeyPropertiesArgs{
			Kty:     pulumi.String("RSA-HSM"), // HSM-backed keys for production
			KeySize: pulumi.Int(4096),         // Larger key size for production
			KeyOps: pulumi.StringArray{
				pulumi.String("encrypt"),
				pulumi.String("decrypt"),
				pulumi.String("sign"),
				pulumi.String("verify"),
				pulumi.String("wrapKey"),
				pulumi.String("unwrapKey"),
			},
			Attributes: &keyvault.KeyAttributesArgs{
				Enabled:    pulumi.Bool(true),
				Exportable: pulumi.Bool(false), // Non-exportable for production security
			},
		},
		Tags: pulumi.StringMap{
			"environment":      pulumi.String("production"),
			"project":         pulumi.String("international-center"),
			"compliance":     pulumi.String("required"),
			"backup-required": pulumi.String("true"),
			"hsm-backed":     pulumi.String("true"),
		},
	})
	if err != nil {
		return err
	}

	stack.keys[keyName] = key
	return nil
}

func (stack *VaultProductionStack) enableSecurityAssessment(ctx *pulumi.Context) error {
	securityAssessment, err := security.NewAssessment(ctx, "production-keyvault-security-assessment", &security.AssessmentArgs{
		ResourceId: stack.keyVault.ID(),
		AssessmentName: pulumi.String("production-keyvault-assessment"),
		Status: &security.AssessmentStatusArgs{
			Code: pulumi.String("Healthy"),
		},
		Metadata: &security.SecurityAssessmentMetadataPropertiesArgs{
			DisplayName: pulumi.String("Production Key Vault Security Assessment"),
			Description: pulumi.String("Security assessment for production Azure Key Vault"),
			AssessmentType: pulumi.String("BuiltIn"),
			Category: pulumi.StringArray{
				pulumi.String("IdentityAndAccess"),
			},
			Severity: pulumi.String("High"),
		},
	})
	if err != nil {
		return err
	}

	stack.securityAssessment = securityAssessment
	return nil
}

func (stack *VaultProductionStack) GetKeyVault() *keyvault.Vault {
	return stack.keyVault
}

func (stack *VaultProductionStack) GetBackupKeyVault() *keyvault.Vault {
	return stack.backupVault
}

func (stack *VaultProductionStack) GetSecret(name string) *keyvault.Secret {
	return stack.secrets[name]
}

func (stack *VaultProductionStack) GetCertificate(name string) *keyvault.Certificate {
	return stack.certificates[name]
}

func (stack *VaultProductionStack) GetKey(name string) *keyvault.Key {
	return stack.keys[name]
}

func (stack *VaultProductionStack) GetVaultUri() pulumi.StringOutput {
	return stack.keyVault.Properties.VaultUri()
}

func (stack *VaultProductionStack) GetBackupVaultUri() pulumi.StringOutput {
	return stack.backupVault.Properties.VaultUri()
}

func (stack *VaultProductionStack) GetDaprSecretStoreConfiguration() map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "dapr.io/v1alpha1",
		"kind":       "Component",
		"metadata": map[string]interface{}{
			"name": "secretstore",
		},
		"spec": map[string]interface{}{
			"type":    "secretstores.azure.keyvault",
			"version": "v1",
			"metadata": []map[string]interface{}{
				{
					"name":  "vaultName",
					"value": "international-center-production-kv",
				},
				{
					"name":  "azureTenantId",
					"value": "", // From environment
				},
				{
					"name":  "azureClientId", 
					"value": "", // From environment
				},
				{
					"name":      "azureClientSecret",
					"secretRef": "azure-client-secret",
				},
			},
		},
	}
}

func (stack *VaultProductionStack) GetPrivateEndpoint() *network.PrivateEndpoint {
	return stack.privateEndpoint
}