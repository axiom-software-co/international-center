package infrastructure

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-azure-native/sdk/v2/go/azurenative/network"
	"github.com/pulumi/pulumi-azure-native/sdk/v2/go/azurenative/resources"
	"github.com/pulumi/pulumi-azure-native/sdk/v2/go/azurenative/security"
	"github.com/pulumi/pulumi-azure-native/sdk/v2/go/azurenative/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AzureProductionStorageStack struct {
	resourceGroup         *resources.ResourceGroup
	vnet                 *network.VirtualNetwork
	privateSubnet        *network.Subnet
	primaryStorageAccount *storage.StorageAccount
	backupStorageAccount  *storage.StorageAccount
	containers           map[string]*storage.BlobContainer
	queues               map[string]*storage.Queue
	privateEndpoint      *network.PrivateEndpoint
	privateDnsZone       *network.PrivateZone
	accessKeys           storage.ListStorageAccountKeysResultOutput
	backupAccessKeys     storage.ListStorageAccountKeysResultOutput
	securityAssessment   *security.Assessment
	lifecyclePolicy      *storage.ManagementPolicy
}

func NewAzureProductionStorageStack(resourceGroup *resources.ResourceGroup, vnet *network.VirtualNetwork, privateSubnet *network.Subnet) *AzureProductionStorageStack {
	return &AzureProductionStorageStack{
		resourceGroup: resourceGroup,
		vnet:         vnet,
		privateSubnet: privateSubnet,
		containers:   make(map[string]*storage.BlobContainer),
		queues:       make(map[string]*storage.Queue),
	}
}

func (stack *AzureProductionStorageStack) Deploy(ctx *pulumi.Context) error {
	if err := stack.createPrivateDnsZone(ctx); err != nil {
		return fmt.Errorf("failed to create private DNS zone: %w", err)
	}

	if err := stack.createPrimaryStorageAccount(ctx); err != nil {
		return fmt.Errorf("failed to create primary storage account: %w", err)
	}

	if err := stack.createBackupStorageAccount(ctx); err != nil {
		return fmt.Errorf("failed to create backup storage account: %w", err)
	}

	if err := stack.createBlobContainers(ctx); err != nil {
		return fmt.Errorf("failed to create blob containers: %w", err)
	}

	if err := stack.createQueues(ctx); err != nil {
		return fmt.Errorf("failed to create queues: %w", err)
	}

	if err := stack.createPrivateEndpoint(ctx); err != nil {
		return fmt.Errorf("failed to create private endpoint: %w", err)
	}

	if err := stack.createLifecycleManagementPolicy(ctx); err != nil {
		return fmt.Errorf("failed to create lifecycle policy: %w", err)
	}

	if err := stack.retrieveAccessKeys(ctx); err != nil {
		return fmt.Errorf("failed to retrieve access keys: %w", err)
	}

	if err := stack.enableSecurityAssessment(ctx); err != nil {
		return fmt.Errorf("failed to enable security assessment: %w", err)
	}

	return nil
}

func (stack *AzureProductionStorageStack) createPrivateDnsZone(ctx *pulumi.Context) error {
	privateDnsZone, err := network.NewPrivateZone(ctx, "production-storage-dns-zone", &network.PrivateZoneArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		PrivateZoneName:   pulumi.String("privatelink.blob.core.windows.net"),
		Location:         pulumi.String("Global"),
		Tags: pulumi.StringMap{
			"environment": pulumi.String("production"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	vnetLink, err := network.NewVirtualNetworkLink(ctx, "production-storage-vnet-link", &network.VirtualNetworkLinkArgs{
		ResourceGroupName:      stack.resourceGroup.Name,
		PrivateZoneName:        privateDnsZone.Name,
		VirtualNetworkLinkName: pulumi.String("production-storage-vnet-link"),
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

func (stack *AzureProductionStorageStack) createPrimaryStorageAccount(ctx *pulumi.Context) error {
	storageAccount, err := storage.NewStorageAccount(ctx, "production-storage-primary", &storage.StorageAccountArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		AccountName:       pulumi.String("intcenterproduction"),
		Location:         stack.resourceGroup.Location,
		Kind:             pulumi.String("StorageV2"),
		Sku: &storage.SkuArgs{
			Name: pulumi.String("Standard_GRS"), // Geo-redundant storage for production
		},
		AccessTier:                          pulumi.String("Hot"),
		AllowBlobPublicAccess:               pulumi.Bool(false),
		AllowSharedKeyAccess:                pulumi.Bool(true), // Required for some integrations
		MinimumTlsVersion:                  pulumi.String("TLS1_2"),
		SupportsHttpsTrafficOnly:           pulumi.Bool(true),
		AllowCrossTenantReplication:        pulumi.Bool(false),
		DefaultToOAuthAuthentication:       pulumi.Bool(true),
		PublicNetworkAccess:                pulumi.String("Disabled"), // Private access only
		NetworkRuleSet: &storage.NetworkRuleSetArgs{
			DefaultAction: pulumi.String("Deny"),
			Bypass:        pulumi.String("AzureServices"),
			VirtualNetworkRules: storage.VirtualNetworkRuleArray{
				&storage.VirtualNetworkRuleArgs{
					VirtualNetworkResourceId: stack.vnet.ID(),
					Action:                  pulumi.String("Allow"),
				},
			},
		},
		Encryption: &storage.EncryptionArgs{
			Services: &storage.EncryptionServicesArgs{
				Blob: &storage.EncryptionServiceArgs{
					Enabled: pulumi.Bool(true),
					KeyType: pulumi.String("Account"),
				},
				File: &storage.EncryptionServiceArgs{
					Enabled: pulumi.Bool(true),
					KeyType: pulumi.String("Account"),
				},
				Queue: &storage.EncryptionServiceArgs{
					Enabled: pulumi.Bool(true),
					KeyType: pulumi.String("Account"),
				},
				Table: &storage.EncryptionServiceArgs{
					Enabled: pulumi.Bool(true),
					KeyType: pulumi.String("Account"),
				},
			},
			KeySource:                       pulumi.String("Microsoft.Storage"),
			RequireInfrastructureEncryption: pulumi.Bool(true),
		},
		Identity: &storage.IdentityArgs{
			Type: pulumi.String("SystemAssigned"),
		},
		Tags: pulumi.StringMap{
			"environment":      pulumi.String("production"),
			"project":         pulumi.String("international-center"),
			"tier":           pulumi.String("storage"),
			"role":           pulumi.String("primary"),
			"compliance":     pulumi.String("required"),
			"backup-required": pulumi.String("true"),
			"encryption":     pulumi.String("enabled"),
		},
	})
	if err != nil {
		return err
	}

	stack.primaryStorageAccount = storageAccount
	return nil
}

func (stack *AzureProductionStorageStack) createBackupStorageAccount(ctx *pulumi.Context) error {
	backupStorageAccount, err := storage.NewStorageAccount(ctx, "production-storage-backup", &storage.StorageAccountArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		AccountName:       pulumi.String("intcenterprodbackup"),
		Location:         pulumi.String("West US 2"), // Different region for backup
		Kind:             pulumi.String("StorageV2"),
		Sku: &storage.SkuArgs{
			Name: pulumi.String("Standard_GRS"), // Geo-redundant storage for backup
		},
		AccessTier:                          pulumi.String("Cool"), // Cool tier for backup storage
		AllowBlobPublicAccess:               pulumi.Bool(false),
		AllowSharedKeyAccess:                pulumi.Bool(true),
		MinimumTlsVersion:                  pulumi.String("TLS1_2"),
		SupportsHttpsTrafficOnly:           pulumi.Bool(true),
		AllowCrossTenantReplication:        pulumi.Bool(false),
		DefaultToOAuthAuthentication:       pulumi.Bool(true),
		PublicNetworkAccess:                pulumi.String("Disabled"),
		NetworkRuleSet: &storage.NetworkRuleSetArgs{
			DefaultAction: pulumi.String("Deny"),
			Bypass:        pulumi.String("AzureServices"),
		},
		Encryption: &storage.EncryptionArgs{
			Services: &storage.EncryptionServicesArgs{
				Blob: &storage.EncryptionServiceArgs{
					Enabled: pulumi.Bool(true),
					KeyType: pulumi.String("Account"),
				},
				File: &storage.EncryptionServiceArgs{
					Enabled: pulumi.Bool(true),
					KeyType: pulumi.String("Account"),
				},
			},
			KeySource:                       pulumi.String("Microsoft.Storage"),
			RequireInfrastructureEncryption: pulumi.Bool(true),
		},
		Identity: &storage.IdentityArgs{
			Type: pulumi.String("SystemAssigned"),
		},
		Tags: pulumi.StringMap{
			"environment":      pulumi.String("production"),
			"project":         pulumi.String("international-center"),
			"tier":           pulumi.String("storage"),
			"role":           pulumi.String("backup"),
			"compliance":     pulumi.String("required"),
			"backup-required": pulumi.String("false"), // This is the backup
		},
	})
	if err != nil {
		return err
	}

	stack.backupStorageAccount = backupStorageAccount
	return nil
}

func (stack *AzureProductionStorageStack) createBlobContainers(ctx *pulumi.Context) error {
	containers := []string{"content", "media", "documents", "backups", "logs", "compliance", "disaster-recovery"}
	
	for _, containerName := range containers {
		if err := stack.createBlobContainer(ctx, containerName); err != nil {
			return fmt.Errorf("failed to create container %s: %w", containerName, err)
		}

		// Create backup container in backup storage account
		if err := stack.createBackupBlobContainer(ctx, containerName); err != nil {
			return fmt.Errorf("failed to create backup container %s: %w", containerName, err)
		}
	}

	return nil
}

func (stack *AzureProductionStorageStack) createBlobContainer(ctx *pulumi.Context, containerName string) error {
	container, err := storage.NewBlobContainer(ctx, fmt.Sprintf("production-%s-container", containerName), &storage.BlobContainerArgs{
		ResourceGroupName:  stack.resourceGroup.Name,
		AccountName:        stack.primaryStorageAccount.Name,
		ContainerName:      pulumi.String(containerName),
		PublicAccess:       pulumi.String("None"),
		DefaultEncryptionScope: pulumi.String("$account-encryption-key"),
		DenyEncryptionScopeOverride: pulumi.Bool(true),
		Metadata: pulumi.StringMap{
			"environment":      pulumi.String("production"),
			"project":         pulumi.String("international-center"),
			"compliance":      pulumi.String("required"),
			"backup-required": pulumi.String("true"),
		},
	})
	if err != nil {
		return err
	}

	stack.containers[containerName] = container
	return nil
}

func (stack *AzureProductionStorageStack) createBackupBlobContainer(ctx *pulumi.Context, containerName string) error {
	backupContainer, err := storage.NewBlobContainer(ctx, fmt.Sprintf("production-backup-%s-container", containerName), &storage.BlobContainerArgs{
		ResourceGroupName:  stack.resourceGroup.Name,
		AccountName:        stack.backupStorageAccount.Name,
		ContainerName:      pulumi.String(fmt.Sprintf("%s-backup", containerName)),
		PublicAccess:       pulumi.String("None"),
		DefaultEncryptionScope: pulumi.String("$account-encryption-key"),
		DenyEncryptionScopeOverride: pulumi.Bool(true),
		Metadata: pulumi.StringMap{
			"environment": pulumi.String("production"),
			"project":     pulumi.String("international-center"),
			"role":        pulumi.String("backup"),
		},
	})
	if err != nil {
		return err
	}

	stack.containers[fmt.Sprintf("%s-backup", containerName)] = backupContainer
	return nil
}

func (stack *AzureProductionStorageStack) createQueues(ctx *pulumi.Context) error {
	queueNames := []string{
		"content-processing",
		"image-processing", 
		"document-processing",
		"notification-queue",
		"audit-events",
		"compliance-events",
		"backup-events",
		"virus-scan-events",
	}
	
	for _, queueName := range queueNames {
		if err := stack.createQueue(ctx, queueName); err != nil {
			return fmt.Errorf("failed to create queue %s: %w", queueName, err)
		}
	}

	return nil
}

func (stack *AzureProductionStorageStack) createQueue(ctx *pulumi.Context, queueName string) error {
	queue, err := storage.NewQueue(ctx, fmt.Sprintf("production-%s-queue", queueName), &storage.QueueArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		AccountName:       stack.primaryStorageAccount.Name,
		QueueName:         pulumi.String(queueName),
		Metadata: pulumi.StringMap{
			"environment":      pulumi.String("production"),
			"project":         pulumi.String("international-center"),
			"compliance":      pulumi.String("required"),
			"backup-required": pulumi.String("true"),
		},
	})
	if err != nil {
		return err
	}

	stack.queues[queueName] = queue
	return nil
}

func (stack *AzureProductionStorageStack) createPrivateEndpoint(ctx *pulumi.Context) error {
	privateEndpoint, err := network.NewPrivateEndpoint(ctx, "production-storage-private-endpoint", &network.PrivateEndpointArgs{
		ResourceGroupName:   stack.resourceGroup.Name,
		PrivateEndpointName: pulumi.String("international-center-production-storage-pe"),
		Location:           stack.resourceGroup.Location,
		Subnet: &network.SubnetArgs{
			Id: stack.privateSubnet.ID(),
		},
		PrivateLinkServiceConnections: network.PrivateLinkServiceConnectionArray{
			&network.PrivateLinkServiceConnectionArgs{
				Name:                 pulumi.String("storage-blob-connection"),
				PrivateLinkServiceId: stack.primaryStorageAccount.ID(),
				GroupIds: pulumi.StringArray{
					pulumi.String("blob"),
				},
			},
			&network.PrivateLinkServiceConnectionArgs{
				Name:                 pulumi.String("storage-queue-connection"),
				PrivateLinkServiceId: stack.primaryStorageAccount.ID(),
				GroupIds: pulumi.StringArray{
					pulumi.String("queue"),
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

	privateDnsZoneGroup, err := network.NewPrivateZoneGroup(ctx, "production-storage-dns-zone-group", &network.PrivateZoneGroupArgs{
		ResourceGroupName:       stack.resourceGroup.Name,
		PrivateEndpointName:     privateEndpoint.Name,
		PrivateDnsZoneGroupName: pulumi.String("default"),
		PrivateDnsZoneConfigs: network.PrivateDnsZoneConfigArray{
			&network.PrivateDnsZoneConfigArgs{
				Name: pulumi.String("storage-blob-config"),
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

func (stack *AzureProductionStorageStack) createLifecycleManagementPolicy(ctx *pulumi.Context) error {
	lifecyclePolicy, err := storage.NewManagementPolicy(ctx, "production-storage-lifecycle-policy", &storage.ManagementPolicyArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		AccountName:          stack.primaryStorageAccount.Name,
		ManagementPolicyName: pulumi.String("default"),
		Policy: &storage.ManagementPolicySchemaArgs{
			Rules: storage.ManagementPolicyRuleArray{
				&storage.ManagementPolicyRuleArgs{
					Name:    pulumi.String("moveToArchiveAfter90Days"),
					Enabled: pulumi.Bool(true),
					Type:    pulumi.String("Lifecycle"),
					Definition: &storage.ManagementPolicyDefinitionArgs{
						Actions: &storage.ManagementPolicyActionArgs{
							BaseBlob: &storage.ManagementPolicyBaseBlobArgs{
								TierToArchive: &storage.DateAfterModificationArgs{
									DaysAfterModificationGreaterThan: pulumi.Float64(90),
								},
								TierToCool: &storage.DateAfterModificationArgs{
									DaysAfterModificationGreaterThan: pulumi.Float64(30),
								},
								Delete: &storage.DateAfterModificationArgs{
									DaysAfterModificationGreaterThan: pulumi.Float64(2555), // 7 years retention
								},
							},
							Snapshot: &storage.ManagementPolicySnapShotArgs{
								TierToArchive: &storage.DateAfterCreationArgs{
									DaysAfterCreationGreaterThan: pulumi.Float64(90),
								},
								Delete: &storage.DateAfterCreationArgs{
									DaysAfterCreationGreaterThan: pulumi.Float64(2555), // 7 years retention
								},
							},
						},
						Filters: &storage.ManagementPolicyFilterArgs{
							BlobTypes: pulumi.StringArray{
								pulumi.String("blockBlob"),
							},
							PrefixMatch: pulumi.StringArray{
								pulumi.String("content/"),
								pulumi.String("media/"),
								pulumi.String("documents/"),
							},
						},
					},
				},
				&storage.ManagementPolicyRuleArgs{
					Name:    pulumi.String("complianceDataRetention"),
					Enabled: pulumi.Bool(true),
					Type:    pulumi.String("Lifecycle"),
					Definition: &storage.ManagementPolicyDefinitionArgs{
						Actions: &storage.ManagementPolicyActionArgs{
							BaseBlob: &storage.ManagementPolicyBaseBlobArgs{
								TierToArchive: &storage.DateAfterModificationArgs{
									DaysAfterModificationGreaterThan: pulumi.Float64(30),
								},
								Delete: &storage.DateAfterModificationArgs{
									DaysAfterModificationGreaterThan: pulumi.Float64(3650), // 10 years for compliance
								},
							},
						},
						Filters: &storage.ManagementPolicyFilterArgs{
							BlobTypes: pulumi.StringArray{
								pulumi.String("blockBlob"),
							},
							PrefixMatch: pulumi.StringArray{
								pulumi.String("compliance/"),
								pulumi.String("audit/"),
								pulumi.String("logs/"),
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	stack.lifecyclePolicy = lifecyclePolicy
	return nil
}

func (stack *AzureProductionStorageStack) retrieveAccessKeys(ctx *pulumi.Context) error {
	stack.accessKeys = storage.ListStorageAccountKeysOutput(ctx, &storage.ListStorageAccountKeysOutputArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		AccountName:       stack.primaryStorageAccount.Name,
	})

	stack.backupAccessKeys = storage.ListStorageAccountKeysOutput(ctx, &storage.ListStorageAccountKeysOutputArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		AccountName:       stack.backupStorageAccount.Name,
	})

	return nil
}

func (stack *AzureProductionStorageStack) enableSecurityAssessment(ctx *pulumi.Context) error {
	securityAssessment, err := security.NewAssessment(ctx, "production-storage-security-assessment", &security.AssessmentArgs{
		ResourceId: stack.primaryStorageAccount.ID(),
		AssessmentName: pulumi.String("production-storage-assessment"),
		Status: &security.AssessmentStatusArgs{
			Code: pulumi.String("Healthy"),
		},
		Metadata: &security.SecurityAssessmentMetadataPropertiesArgs{
			DisplayName: pulumi.String("Production Storage Security Assessment"),
			Description: pulumi.String("Security assessment for production Azure storage account"),
			AssessmentType: pulumi.String("BuiltIn"),
			Category: pulumi.StringArray{
				pulumi.String("Data"),
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

func (stack *AzureProductionStorageStack) GetPrimaryStorageAccount() *storage.StorageAccount {
	return stack.primaryStorageAccount
}

func (stack *AzureProductionStorageStack) GetBackupStorageAccount() *storage.StorageAccount {
	return stack.backupStorageAccount
}

func (stack *AzureProductionStorageStack) GetContainer(name string) *storage.BlobContainer {
	return stack.containers[name]
}

func (stack *AzureProductionStorageStack) GetQueue(name string) *storage.Queue {
	return stack.queues[name]
}

func (stack *AzureProductionStorageStack) GetConnectionString() pulumi.StringOutput {
	return pulumi.Sprintf("DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;EndpointSuffix=core.windows.net",
		stack.primaryStorageAccount.Name,
		stack.accessKeys.Keys().Index(pulumi.Int(0)).Value(),
	)
}

func (stack *AzureProductionStorageStack) GetBackupConnectionString() pulumi.StringOutput {
	return pulumi.Sprintf("DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;EndpointSuffix=core.windows.net",
		stack.backupStorageAccount.Name,
		stack.backupAccessKeys.Keys().Index(pulumi.Int(0)).Value(),
	)
}

func (stack *AzureProductionStorageStack) GetBlobEndpoint() pulumi.StringOutput {
	return stack.primaryStorageAccount.PrimaryEndpoints.Blob()
}

func (stack *AzureProductionStorageStack) GetQueueEndpoint() pulumi.StringOutput {
	return stack.primaryStorageAccount.PrimaryEndpoints.Queue()
}

func (stack *AzureProductionStorageStack) GetDaprBindingConfiguration(containerName string) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "dapr.io/v1alpha1",
		"kind":       "Component",
		"metadata": map[string]interface{}{
			"name": fmt.Sprintf("%s-storage-binding", containerName),
		},
		"spec": map[string]interface{}{
			"type":    "bindings.azure.blobstorage",
			"version": "v1",
			"metadata": []map[string]interface{}{
				{
					"name":  "storageAccount",
					"value": "intcenterproduction",
				},
				{
					"name":      "storageAccessKey",
					"secretRef": "storage-access-key",
				},
				{
					"name":  "container",
					"value": containerName,
				},
				{
					"name":  "decodeBase64",
					"value": "false",
				},
				{
					"name":  "getBlobRetryCount",
					"value": "5",
				},
			},
		},
	}
}

func (stack *AzureProductionStorageStack) GetPrivateEndpoint() *network.PrivateEndpoint {
	return stack.privateEndpoint
}

func (stack *AzureProductionStorageStack) GetLifecyclePolicy() *storage.ManagementPolicy {
	return stack.lifecyclePolicy
}