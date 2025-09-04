package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi-azure-native-sdk/keyvault/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type VaultCloudStack struct {
	resourceGroup    *resources.ResourceGroup
	keyVault        *keyvault.Vault
	secrets         map[string]*keyvault.Secret
	accessPolicies  []*keyvault.AccessPolicyEntry
	// certificates    map[string]*keyvault.Certificate // TODO: Fix certificate API in Azure Native SDK v3
	keys           map[string]*keyvault.Key
}

func NewVaultCloudStack(resourceGroup *resources.ResourceGroup) *VaultCloudStack {
	return &VaultCloudStack{
		resourceGroup:  resourceGroup,
		secrets:       make(map[string]*keyvault.Secret),
		// certificates:  make(map[string]*keyvault.Certificate), // TODO: Fix certificate API
		keys:         make(map[string]*keyvault.Key),
	}
}

func (stack *VaultCloudStack) Deploy(ctx *pulumi.Context) error {
	if err := stack.createKeyVault(ctx); err != nil {
		return fmt.Errorf("failed to create key vault: %w", err)
	}

	if err := stack.configureAccessPolicies(ctx); err != nil {
		return fmt.Errorf("failed to configure access policies: %w", err)
	}

	if err := stack.createSecrets(ctx); err != nil {
		return fmt.Errorf("failed to create secrets: %w", err)
	}

	if err := stack.createCertificates(ctx); err != nil {
		return fmt.Errorf("failed to create certificates: %w", err)
	}

	if err := stack.createKeys(ctx); err != nil {
		return fmt.Errorf("failed to create keys: %w", err)
	}

	return nil
}

func (stack *VaultCloudStack) createKeyVault(ctx *pulumi.Context) error {
	keyVault, err := keyvault.NewVault(ctx, "staging-keyvault", &keyvault.VaultArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		VaultName:        pulumi.String("international-center-staging-kv"),
		Location:         stack.resourceGroup.Location,
		Properties: &keyvault.VaultPropertiesArgs{
			TenantId: pulumi.String(""), // From environment
			Sku: &keyvault.SkuArgs{
				Name:   keyvault.SkuNameStandard,
				Family: pulumi.String(string(keyvault.SkuFamilyA)),
			},
			EnabledForDeployment:         pulumi.Bool(true),
			EnabledForTemplateDeployment: pulumi.Bool(true),
			EnabledForDiskEncryption:     pulumi.Bool(true),
			EnableRbacAuthorization:      pulumi.Bool(false), // Using access policies
			SoftDeleteRetentionInDays:    pulumi.Int(90),
			EnableSoftDelete:            pulumi.Bool(true),
			EnablePurgeProtection:       pulumi.Bool(false), // More flexible for staging
			NetworkAcls: &keyvault.NetworkRuleSetArgs{
				Bypass:        pulumi.String("AzureServices"),
				DefaultAction: pulumi.String("Allow"),
			},
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	stack.keyVault = keyVault
	return nil
}

func (stack *VaultCloudStack) configureAccessPolicies(ctx *pulumi.Context) error {
	containerAppsPolicy := &keyvault.AccessPolicyEntry{
		TenantId: "", // From environment
		ObjectId: "", // Container Apps managed identity
		Permissions: keyvault.Permissions{
			Secrets: []string{
				"get",
				"list",
			},
			Certificates: []string{
				"get",
				"list",
			},
			Keys: []string{
				"get",
				"decrypt",
				"encrypt",
				"sign",
				"verify",
			},
		},
	}

	deployerServicePrincipalPolicy := &keyvault.AccessPolicyEntry{
		TenantId: "", // From environment
		ObjectId: "", // Deployer service principal
		Permissions: keyvault.Permissions{
			Secrets: []string{
				"get",
				"list",
				"set",
				"delete",
				"recover",
				"backup",
				"restore",
			},
			Certificates: []string{
				"get",
				"list",
				"create",
				"update",
				"delete",
				"import",
			},
			Keys: []string{
				"get",
				"list",
				"create",
				"update",
				"delete",
				"decrypt",
				"encrypt",
				"sign",
				"verify",
			},
		},
	}

	stack.accessPolicies = append(stack.accessPolicies, containerAppsPolicy, deployerServicePrincipalPolicy)
	return nil
}

func (stack *VaultCloudStack) createSecrets(ctx *pulumi.Context) error {
	secretConfigs := map[string]string{
		"database-admin-password":       "", // Generated
		"redis-connection-string":       "", // From Azure Redis
		"storage-account-access-key":    "", // From Azure Storage
		"grafana-api-key":              "", // From Grafana Cloud
		"prometheus-api-key":           "", // From Grafana Cloud  
		"loki-api-key":                "", // From Grafana Cloud
		"azure-client-secret":          "", // For service principal
		"jwt-signing-key":              "", // Generated
		"encryption-key":               "", // Generated
		"smtp-password":                "", // For email service
		"webhook-secret":               "", // For GitHub webhooks
		"api-keys-external-service-a":  "", // External integrations
		"api-keys-external-service-b":  "", // External integrations
	}

	for secretName, secretValue := range secretConfigs {
		if err := stack.createSecret(ctx, secretName, secretValue); err != nil {
			return fmt.Errorf("failed to create secret %s: %w", secretName, err)
		}
	}

	return nil
}

func (stack *VaultCloudStack) createSecret(ctx *pulumi.Context, secretName, secretValue string) error {
	secret, err := keyvault.NewSecret(ctx, fmt.Sprintf("staging-%s-secret", secretName), &keyvault.SecretArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		VaultName:        stack.keyVault.Name,
		SecretName:       pulumi.String(secretName),
		Properties: &keyvault.SecretPropertiesArgs{
			Value: pulumi.String(secretValue),
			ContentType: pulumi.String("text/plain"),
			Attributes: &keyvault.SecretAttributesArgs{
				Enabled:   pulumi.Bool(true),
				NotBefore: pulumi.Int(0), // Available immediately
				// Expires: pulumi.Int(), // Set based on secret type
			},
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	stack.secrets[secretName] = secret
	return nil
}

func (stack *VaultCloudStack) createCertificates(ctx *pulumi.Context) error {
	certificateConfigs := []string{
		"api-staging-international-center-com",
		"admin-staging-international-center-com", 
		"app-staging-international-center-com",
	}

	for _, certName := range certificateConfigs {
		if err := stack.createCertificate(ctx, certName); err != nil {
			return fmt.Errorf("failed to create certificate %s: %w", certName, err)
		}
	}

	return nil
}

func (stack *VaultCloudStack) createCertificate(ctx *pulumi.Context, certName string) error {
	// TODO: Fix certificate API in Azure Native SDK v3.7.1 - API removed/changed
	return nil
}

// TODO: Fix certificate API in Azure Native SDK v3
// func (stack *VaultCloudStack) createCertificate(ctx *pulumi.Context, certName string) error {
//     certificate, err := keyvault.NewCertificate(ctx, fmt.Sprintf("staging-%s-cert", certName), &keyvault.CertificateArgs{
//         ResourceGroupName: stack.resourceGroup.Name,
//         VaultName:        stack.keyVault.Name,
//         CertificateName:  pulumi.String(certName),
//         Properties: &keyvault.CertificatePropertiesArgs{
//             IssuerParameters: &keyvault.IssuerParametersArgs{
//                 Name: pulumi.String("Self"), // Self-signed for staging
//             },
//             KeyProperties: &keyvault.KeyPropertiesArgs{
//                 Exportable: pulumi.Bool(true),
//                 KeySize:    pulumi.Int(2048),
//                 KeyType:    pulumi.String("RSA"),
//                 ReuseKey:   pulumi.Bool(false),
//             },
//             SecretProperties: &keyvault.SecretPropertiesArgs{
//                 ContentType: pulumi.String("application/x-pkcs12"),
//             },
//             X509CertificateProperties: &keyvault.X509CertificatePropertiesArgs{
//                 Subject: pulumi.String(fmt.Sprintf("CN=%s", certName)),
//                 ValidityInMonths: pulumi.Int(12),
//                 KeyUsage: pulumi.StringArray{
//                     pulumi.String("digitalSignature"),
//                     pulumi.String("keyEncipherment"),
//                 },
//                 Ekus: pulumi.StringArray{
//                     pulumi.String("1.3.6.1.5.5.7.3.1"), // Server Authentication
//                     pulumi.String("1.3.6.1.5.5.7.3.2"), // Client Authentication
//                 },
//                 SubjectAlternativeNames: &keyvault.SubjectAlternativeNamesArgs{
//                     DnsNames: pulumi.StringArray{
//                         pulumi.String(certName),
//                     },
//                 },
//             },
//         },
//         Tags: pulumi.StringMap{
//             "environment": pulumi.String("staging"),
//             "project":     pulumi.String("international-center"),
//         },
//     })
//     if err != nil {
//         return err
//     }
//
//     stack.certificates[certName] = certificate
//     return nil
// }

func (stack *VaultCloudStack) createKeys(ctx *pulumi.Context) error {
	keyConfigs := []string{
		"data-encryption-key",
		"jwt-signing-key",
		"api-signing-key",
	}

	for _, keyName := range keyConfigs {
		if err := stack.createKey(ctx, keyName); err != nil {
			return fmt.Errorf("failed to create key %s: %w", keyName, err)
		}
	}

	return nil
}

func (stack *VaultCloudStack) createKey(ctx *pulumi.Context, keyName string) error {
	key, err := keyvault.NewKey(ctx, fmt.Sprintf("staging-%s-key", keyName), &keyvault.KeyArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		VaultName:        stack.keyVault.Name,
		KeyName:          pulumi.String(keyName),
		Properties: &keyvault.KeyPropertiesArgs{
			Kty: pulumi.String("RSA"),
			KeySize: pulumi.Int(2048),
			KeyOps: pulumi.StringArray{
				pulumi.String("encrypt"),
				pulumi.String("decrypt"),
				pulumi.String("sign"),
				pulumi.String("verify"),
				pulumi.String("wrapKey"),
				pulumi.String("unwrapKey"),
			},
			Attributes: &keyvault.KeyAttributesArgs{
				Enabled: pulumi.Bool(true),
			},
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	stack.keys[keyName] = key
	return nil
}

func (stack *VaultCloudStack) GetKeyVault() *keyvault.Vault {
	return stack.keyVault
}

func (stack *VaultCloudStack) GetSecret(name string) *keyvault.Secret {
	return stack.secrets[name]
}

// func (stack *VaultCloudStack) GetCertificate(name string) *keyvault.Certificate {
//     return stack.certificates[name]
// } // TODO: Fix certificate API in Azure Native SDK v3

func (stack *VaultCloudStack) GetKey(name string) *keyvault.Key {
	return stack.keys[name]
}

func (stack *VaultCloudStack) GetVaultUri() pulumi.StringOutput {
	return stack.keyVault.Properties.VaultUri()
}

func (stack *VaultCloudStack) GetDaprSecretStoreConfiguration() map[string]interface{} {
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
					"value": "international-center-staging-kv",
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