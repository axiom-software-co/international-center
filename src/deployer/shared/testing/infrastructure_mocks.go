package testing

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// InfrastructureMocks provides Pulumi mock implementations for infrastructure testing
type InfrastructureMocks struct {
	environment string
	resources   map[string]resource.PropertyMap
	calls       map[string]resource.PropertyMap
}

// NewInfrastructureMocks creates a new mock provider for infrastructure testing
func NewInfrastructureMocks(environment string) *InfrastructureMocks {
	return &InfrastructureMocks{
		environment: environment,
		resources:   make(map[string]resource.PropertyMap),
		calls:       make(map[string]resource.PropertyMap),
	}
}

// NewResource implements pulumi.MockResourceArgs for infrastructure resources
func (m *InfrastructureMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}
	
	switch args.TypeToken {
	case "azure-native:resources:ResourceGroup":
		outputs["name"] = resource.NewStringProperty(args.Name)
		outputs["location"] = resource.NewStringProperty("East US")
		outputs["id"] = resource.NewStringProperty("/subscriptions/test/resourceGroups/" + args.Name)
		
	case "azure-native:dbforpostgresql:Server":
		outputs["name"] = resource.NewStringProperty(args.Name)
		outputs["fullyQualifiedDomainName"] = resource.NewStringProperty(args.Name + ".postgres.database.azure.com")
		outputs["id"] = resource.NewStringProperty("/subscriptions/test/servers/" + args.Name)
		
	case "azure-native:storage:StorageAccount":
		outputs["name"] = resource.NewStringProperty(args.Name)
		outputs["primaryEndpoints"] = resource.NewObjectProperty(resource.PropertyMap{
			"blob":  resource.NewStringProperty("https://" + args.Name + ".blob.core.windows.net/"),
			"queue": resource.NewStringProperty("https://" + args.Name + ".queue.core.windows.net/"),
		})
		outputs["id"] = resource.NewStringProperty("/subscriptions/test/storageAccounts/" + args.Name)
		
	case "azure-native:keyvault:Vault":
		outputs["name"] = resource.NewStringProperty(args.Name)
		outputs["properties"] = resource.NewObjectProperty(resource.PropertyMap{
			"vaultUri": resource.NewStringProperty("https://" + args.Name + ".vault.azure.net/"),
		})
		outputs["id"] = resource.NewStringProperty("/subscriptions/test/vaults/" + args.Name)
		
	case "azure-native:app:ContainerApp":
		outputs["name"] = resource.NewStringProperty(args.Name)
		outputs["configuration"] = resource.NewObjectProperty(resource.PropertyMap{
			"ingress": resource.NewObjectProperty(resource.PropertyMap{
				"fqdn": resource.NewStringProperty(args.Name + "." + m.environment + ".azurecontainerapps.io"),
			}),
		})
		outputs["id"] = resource.NewStringProperty("/subscriptions/test/containerApps/" + args.Name)
		
	default:
		outputs["name"] = resource.NewStringProperty(args.Name)
		outputs["id"] = resource.NewStringProperty("/subscriptions/test/resources/" + args.Name)
	}
	
	// Store outputs for later retrieval
	m.resources[args.Name] = outputs
	
	return args.Name + "-id", outputs, nil
}

// Call implements pulumi.MockCallArgs for function calls
func (m *InfrastructureMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}
	
	switch args.Token {
	case "azure-native:storage:listStorageAccountKeys":
		outputs["keys"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{
				"value": resource.NewStringProperty("mock-storage-key"),
			}),
		})
		
	case "azure-native:dbforpostgresql:listServerConfigurations":
		outputs["value"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{
				"name":  resource.NewStringProperty("max_connections"),
				"value": resource.NewStringProperty("100"),
			}),
		})
		
	default:
		outputs["result"] = resource.NewStringProperty("mock-result")
	}
	
	// Store call outputs for later retrieval
	m.calls[args.Token] = outputs
	
	return outputs, nil
}

// GetProviderName returns the provider name for this mock
func (m *InfrastructureMocks) GetProviderName() string {
	return "azure-native"
}

// GetResourceTypes returns the resource types this mock supports
func (m *InfrastructureMocks) GetResourceTypes() []string {
	return []string{
		"azure-native:resources:ResourceGroup",
		"azure-native:dbforpostgresql:Server",
		"azure-native:storage:StorageAccount",
		"azure-native:keyvault:Vault",
		"azure-native:app:ContainerApp",
	}
}

// GetMockedResource retrieves mocked resource outputs by name
func (m *InfrastructureMocks) GetMockedResource(name string) (resource.PropertyMap, bool) {
	outputs, exists := m.resources[name]
	return outputs, exists
}

// GetMockedCall retrieves mocked call outputs by token
func (m *InfrastructureMocks) GetMockedCall(token string) (resource.PropertyMap, bool) {
	outputs, exists := m.calls[token]
	return outputs, exists
}