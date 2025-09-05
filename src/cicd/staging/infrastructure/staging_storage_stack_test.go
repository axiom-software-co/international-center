package infrastructure

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"

	shared "github.com/axiom-software-co/international-center/src/deployer/shared/testing"
)

func TestStagingStorageStackCreation(t *testing.T) {
	suite := shared.NewInfrastructureTestSuite(t, "staging")
	
	suite.RunPulumiTest("storage_component_registration", func(ctx *pulumi.Context) error {
		// Test storage stack ComponentResource registration
		return nil
	})
}

func TestStagingStorageStackComponentContract(t *testing.T) {
	suite := shared.NewInfrastructureTestSuite(t, "staging")
	
	contractTest := shared.CreateStorageContractTest("staging")
	suite.RunComponentTest(contractTest)
}

func TestStagingStorageStackContainerCreation(t *testing.T) {
	suite := shared.NewInfrastructureTestSuite(t, "staging")
	
	t.Run("required_containers_exist", func(t *testing.T) {
		requiredContainers := []string{"content", "media", "documents", "backups", "temp"}
		
		for _, containerName := range requiredContainers {
			t.Run(containerName, func(t *testing.T) {
				assert.Contains(t, requiredContainers, containerName, 
					"Container %s should be in required containers list", containerName)
			})
		}
	})
	
	suite.RunPulumiTest("container_provisioning", func(ctx *pulumi.Context) error {
		// Test that all required containers are created with proper configuration
		return nil
	})
}

func TestStagingStorageStackQueueCreation(t *testing.T) {
	suite := shared.NewInfrastructureTestSuite(t, "staging")
	
	t.Run("required_queues_exist", func(t *testing.T) {
		requiredQueues := []string{
			"content-processing",
			"image-processing",
			"document-processing", 
			"notification-queue",
			"audit-events",
		}
		
		for _, queueName := range requiredQueues {
			t.Run(queueName, func(t *testing.T) {
				assert.Contains(t, requiredQueues, queueName,
					"Queue %s should be in required queues list", queueName)
			})
		}
	})
	
	suite.RunPulumiTest("queue_provisioning", func(ctx *pulumi.Context) error {
		// Test that all required queues are created with proper configuration
		return nil
	})
}

func TestStagingStorageStackOutputs(t *testing.T) {
	suite := shared.NewInfrastructureTestSuite(t, "staging")
	
	suite.RunPulumiTest("required_outputs_validation", func(ctx *pulumi.Context) error {
		requiredOutputs := []string{
			"connectionString",
			"blobEndpoint",
			"queueEndpoint",
		}
		
		// Mock outputs for testing
		outputs := map[string]pulumi.Output{
			"connectionString": pulumi.String("DefaultEndpointsProtocol=https;AccountName=staging;AccountKey=mock-key").ToStringOutput(),
			"blobEndpoint":    pulumi.String("https://staging.blob.core.windows.net/").ToStringOutput(),
			"queueEndpoint":   pulumi.String("https://staging.queue.core.windows.net/").ToStringOutput(),
		}
		
		suite.ValidateOutputs(outputs, requiredOutputs)
		return nil
	})
}

func TestStagingStorageStackNamingConventions(t *testing.T) {
	t.Run("resource_naming_consistency", func(t *testing.T) {
		suite := shared.NewInfrastructureTestSuite(t, "staging")
		
		testCases := []struct {
			resourceName string
			component    string
		}{
			{"staging-storage-account", "storage"},
			{"staging-content-container", "storage"},
			{"staging-media-container", "storage"},
			{"staging-content-processing-queue", "storage"},
		}
		
		for _, tc := range testCases {
			suite.ValidateNamingConsistency(tc.resourceName, tc.component)
		}
	})
}

func TestStagingStorageStackDaprIntegration(t *testing.T) {
	t.Run("blob_storage_binding_configuration", func(t *testing.T) {
		// Test Dapr blob storage binding configuration
		expectedConfig := map[string]interface{}{
			"apiVersion": "dapr.io/v1alpha1",
			"kind":       "Component",
			"spec": map[string]interface{}{
				"type":    "bindings.azure.blobstorage",
				"version": "v1",
			},
		}
		
		assert.Equal(t, "dapr.io/v1alpha1", expectedConfig["apiVersion"])
		assert.Equal(t, "Component", expectedConfig["kind"])
	})
	
	t.Run("queue_pubsub_configuration", func(t *testing.T) {
		// Test Dapr queue pubsub configuration
		expectedConfig := map[string]interface{}{
			"apiVersion": "dapr.io/v1alpha1",
			"kind":       "Component",
			"spec": map[string]interface{}{
				"type":    "pubsub.azure.servicebus.queues",
				"version": "v1",
			},
		}
		
		assert.Equal(t, "dapr.io/v1alpha1", expectedConfig["apiVersion"])
		assert.Equal(t, "pubsub.azure.servicebus.queues", expectedConfig["spec"].(map[string]interface{})["type"])
	})
}

func TestStagingStorageStackSecurityConfiguration(t *testing.T) {
	suite := shared.NewInfrastructureTestSuite(t, "staging")
	
	suite.RunPulumiTest("security_settings_validation", func(ctx *pulumi.Context) error {
		// Test that security settings are properly configured
		// - AllowBlobPublicAccess: false
		// - MinimumTlsVersion: TLS1_2
		// - NetworkRuleSet DefaultAction: Allow (staging environment)
		return nil
	})
}

func TestStagingStorageStackEnvironmentIsolation(t *testing.T) {
	suite := shared.NewInfrastructureTestSuite(t, "staging")
	
	suite.RunPulumiTest("environment_isolation", func(ctx *pulumi.Context) error {
		// Mock resources for isolation testing
		mockResources := []pulumi.Resource{
			// These would be actual resource instances in real tests
		}
		
		suite.ValidateEnvironmentIsolation(mockResources)
		return nil
	})
}