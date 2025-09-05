package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi-azure-native-sdk/resources/v3"
	"github.com/pulumi/pulumi-azure-native-sdk/storage/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AzureStorageStack struct {
	resourceGroup   *resources.ResourceGroup
	storageAccount  *storage.StorageAccount
	containers      map[string]*storage.BlobContainer
	queues          map[string]*storage.Queue
	accessKeys      storage.ListStorageAccountKeysResultOutput
}

func NewAzureStorageStack(resourceGroup *resources.ResourceGroup) *AzureStorageStack {
	return &AzureStorageStack{
		resourceGroup: resourceGroup,
		containers:    make(map[string]*storage.BlobContainer),
		queues:       make(map[string]*storage.Queue),
	}
}

func (stack *AzureStorageStack) Deploy(ctx *pulumi.Context) error {
	if err := stack.createStorageAccount(ctx); err != nil {
		return fmt.Errorf("failed to create storage account: %w", err)
	}

	if err := stack.createBlobContainers(ctx); err != nil {
		return fmt.Errorf("failed to create blob containers: %w", err)
	}

	if err := stack.createQueues(ctx); err != nil {
		return fmt.Errorf("failed to create queues: %w", err)
	}

	if err := stack.retrieveAccessKeys(ctx); err != nil {
		return fmt.Errorf("failed to retrieve access keys: %w", err)
	}

	return nil
}

func (stack *AzureStorageStack) createStorageAccount(ctx *pulumi.Context) error {
	storageAccount, err := storage.NewStorageAccount(ctx, "staging-storage", &storage.StorageAccountArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		AccountName:       pulumi.String("internationalcenterstaging"),
		Location:         stack.resourceGroup.Location,
		Kind:             pulumi.String("StorageV2"),
		Sku: &storage.SkuArgs{
			Name: pulumi.String("Standard_LRS"), // Locally redundant for staging
		},
		AccessTier:                   storage.AccessTierHot,
		AllowBlobPublicAccess:        pulumi.Bool(false), // Security best practice
		AllowSharedKeyAccess:         pulumi.Bool(true),  // Required for some integrations
		MinimumTlsVersion:           pulumi.String("TLS1_2"),
		NetworkRuleSet: &storage.NetworkRuleSetArgs{
			DefaultAction: storage.DefaultActionAllow, // More permissive for staging
			Bypass:        pulumi.String(string(storage.BypassAzureServices)),
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	stack.storageAccount = storageAccount
	return nil
}

func (stack *AzureStorageStack) createBlobContainers(ctx *pulumi.Context) error {
	containers := []string{"content", "media", "documents", "backups", "temp"}
	
	for _, containerName := range containers {
		if err := stack.createBlobContainer(ctx, containerName); err != nil {
			return fmt.Errorf("failed to create container %s: %w", containerName, err)
		}
	}

	return nil
}

func (stack *AzureStorageStack) createBlobContainer(ctx *pulumi.Context, containerName string) error {
	container, err := storage.NewBlobContainer(ctx, fmt.Sprintf("staging-%s-container", containerName), &storage.BlobContainerArgs{
		ResourceGroupName:  stack.resourceGroup.Name,
		AccountName:        stack.storageAccount.Name,
		ContainerName:      pulumi.String(containerName),
		PublicAccess:       storage.PublicAccessNone, // Private containers
		Metadata: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	stack.containers[containerName] = container
	return nil
}

func (stack *AzureStorageStack) createQueues(ctx *pulumi.Context) error {
	queueNames := []string{
		"content-processing",
		"image-processing", 
		"document-processing",
		"notification-queue",
		"audit-events",
	}
	
	for _, queueName := range queueNames {
		if err := stack.createQueue(ctx, queueName); err != nil {
			return fmt.Errorf("failed to create queue %s: %w", queueName, err)
		}
	}

	return nil
}

func (stack *AzureStorageStack) createQueue(ctx *pulumi.Context, queueName string) error {
	queue, err := storage.NewQueue(ctx, fmt.Sprintf("staging-%s-queue", queueName), &storage.QueueArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		AccountName:       stack.storageAccount.Name,
		QueueName:         pulumi.String(queueName),
		Metadata: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	stack.queues[queueName] = queue
	return nil
}

func (stack *AzureStorageStack) retrieveAccessKeys(ctx *pulumi.Context) error {
	stack.accessKeys = storage.ListStorageAccountKeysOutput(ctx, storage.ListStorageAccountKeysOutputArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		AccountName:       stack.storageAccount.Name,
	})

	return nil
}

func (stack *AzureStorageStack) GetStorageAccount() *storage.StorageAccount {
	return stack.storageAccount
}

func (stack *AzureStorageStack) GetContainer(name string) *storage.BlobContainer {
	return stack.containers[name]
}

func (stack *AzureStorageStack) GetQueue(name string) *storage.Queue {
	return stack.queues[name]
}

func (stack *AzureStorageStack) GetConnectionString() pulumi.StringOutput {
	return pulumi.Sprintf("DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;EndpointSuffix=core.windows.net",
		stack.storageAccount.Name,
		stack.accessKeys.Keys().Index(pulumi.Int(0)).Value(),
	)
}

func (stack *AzureStorageStack) GetBlobEndpoint() pulumi.StringOutput {
	return stack.storageAccount.PrimaryEndpoints.Blob()
}

func (stack *AzureStorageStack) GetQueueEndpoint() pulumi.StringOutput {
	return stack.storageAccount.PrimaryEndpoints.Queue()
}

func (stack *AzureStorageStack) GetDaprBindingConfiguration(containerName string) map[string]interface{} {
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
					"value": "internationalcenterstaging",
				},
				{
					"name":      "storageAccessKey",
					"secretRef": "storage-access-key",
				},
				{
					"name":  "container",
					"value": containerName,
				},
			},
		},
	}
}

func (stack *AzureStorageStack) GetDaprPubSubQueueConfiguration(queueName string) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "dapr.io/v1alpha1",
		"kind":       "Component", 
		"metadata": map[string]interface{}{
			"name": fmt.Sprintf("%s-queue-pubsub", queueName),
		},
		"spec": map[string]interface{}{
			"type":    "pubsub.azure.servicebus.queues", 
			"version": "v1",
			"metadata": []map[string]interface{}{
				{
					"name":      "connectionString",
					"secretRef": "servicebus-connection-string",
				},
			},
		},
	}
}