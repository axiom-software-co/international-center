package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type InfrastructureFramework struct {
	Database  *DatabaseComponent
	Messaging *MessagingComponent
	Storage   *StorageComponent
	Vault     *VaultComponent
}

type FrameworkArgs struct {
	ProjectName      string
	Environment      string
	DatabaseConfig   *DatabaseConfig
	MessagingConfig  *MessagingConfig
	StorageConfig    *StorageConfig
	VaultConfig      *VaultConfig
}

type FrameworkComponent struct {
	pulumi.ResourceState

	Framework *InfrastructureFramework
	Outputs   pulumi.Map
}

func NewFrameworkComponent(ctx *pulumi.Context, name string, args *FrameworkArgs, opts ...pulumi.ResourceOption) (*FrameworkComponent, error) {
	component := &FrameworkComponent{}
	
	if ctx != nil {
		err := ctx.RegisterComponentResource("framework:infrastructure:Framework", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	framework := &InfrastructureFramework{}
	outputs := make(pulumi.Map)

	if args.DatabaseConfig != nil {
		database, err := NewDatabaseComponent(ctx, fmt.Sprintf("%s-database", name), &DatabaseArgs{
			Config:      args.DatabaseConfig,
			Environment: args.Environment,
			ProjectName: args.ProjectName,
		}, pulumi.Parent(component))
		if err != nil {
			return nil, fmt.Errorf("failed to create database component: %w", err)
		}
		framework.Database = database
		outputs["database_connection_string"] = database.ConnectionString
		outputs["database_host"] = database.Host
		outputs["database_port"] = database.Port
		outputs["database_health_endpoint"] = database.HealthEndpoint
	}

	if args.MessagingConfig != nil {
		messaging, err := NewMessagingComponent(ctx, fmt.Sprintf("%s-messaging", name), &MessagingArgs{
			Config:      args.MessagingConfig,
			Environment: args.Environment,
			ProjectName: args.ProjectName,
		}, pulumi.Parent(component))
		if err != nil {
			return nil, fmt.Errorf("failed to create messaging component: %w", err)
		}
		framework.Messaging = messaging
		outputs["messaging_endpoint"] = messaging.Endpoint
		outputs["messaging_host"] = messaging.Host
		outputs["messaging_port"] = messaging.Port
		outputs["messaging_health_endpoint"] = messaging.HealthEndpoint
	}

	if args.StorageConfig != nil {
		storage, err := NewStorageComponent(ctx, fmt.Sprintf("%s-storage", name), &StorageArgs{
			Config:      args.StorageConfig,
			Environment: args.Environment,
			ProjectName: args.ProjectName,
		}, pulumi.Parent(component))
		if err != nil {
			return nil, fmt.Errorf("failed to create storage component: %w", err)
		}
		framework.Storage = storage
		outputs["storage_connection_string"] = storage.ConnectionString
		outputs["storage_endpoint"] = storage.Endpoint
		outputs["storage_bucket_name"] = storage.BucketName
		outputs["storage_health_endpoint"] = storage.HealthEndpoint
	}

	if args.VaultConfig != nil {
		vault, err := NewVaultComponent(ctx, fmt.Sprintf("%s-vault", name), &VaultArgs{
			Config:      args.VaultConfig,
			Environment: args.Environment,
			ProjectName: args.ProjectName,
		}, pulumi.Parent(component))
		if err != nil {
			return nil, fmt.Errorf("failed to create vault component: %w", err)
		}
		framework.Vault = vault
		outputs["vault_address"] = vault.VaultAddress
		outputs["vault_health_endpoint"] = vault.HealthEndpoint
	}

	component.Framework = framework
	component.Outputs = outputs

	if ctx != nil {
		if err := ctx.RegisterResourceOutputs(component, outputs); err != nil {
			return nil, err
		}
	}

	return component, nil
}

func (f *FrameworkComponent) GetDatabaseEndpoint() pulumi.StringOutput {
	if f.Framework.Database != nil {
		return f.Framework.Database.ConnectionString
	}
	return pulumi.String("").ToStringOutput()
}

func (f *FrameworkComponent) GetMessagingEndpoint() pulumi.StringOutput {
	if f.Framework.Messaging != nil {
		return f.Framework.Messaging.Endpoint
	}
	return pulumi.String("").ToStringOutput()
}

func (f *FrameworkComponent) GetStorageEndpoint() pulumi.StringOutput {
	if f.Framework.Storage != nil {
		return f.Framework.Storage.ConnectionString
	}
	return pulumi.String("").ToStringOutput()
}

func (f *FrameworkComponent) GetVaultAddress() pulumi.StringOutput {
	if f.Framework.Vault != nil {
		return f.Framework.Vault.VaultAddress
	}
	return pulumi.String("").ToStringOutput()
}

func DefaultDevelopmentFramework(projectName string) *FrameworkArgs {
	return &FrameworkArgs{
		ProjectName:     projectName,
		Environment:     "development",
		DatabaseConfig:  DefaultPostgreSQLConfig(fmt.Sprintf("%s_dev", projectName), "localhost"),
		MessagingConfig: DefaultRabbitMQConfig("localhost"),
		StorageConfig:   DefaultLocalStorageConfig(fmt.Sprintf("./storage/%s", projectName)),
		VaultConfig:     DefaultHashiCorpVaultConfig("localhost"),
	}
}

func DefaultStagingFramework(projectName string) *FrameworkArgs {
	return &FrameworkArgs{
		ProjectName:     projectName,
		Environment:     "staging",
		DatabaseConfig:  DefaultPostgreSQLConfig(fmt.Sprintf("%s_staging", projectName), "staging-db.example.com"),
		MessagingConfig: DefaultRabbitMQConfig("staging-mq.example.com"),
		StorageConfig:   DefaultAzureBlobConfig("stagingstorage", "staging_key", fmt.Sprintf("%s-staging", projectName)),
		VaultConfig:     DefaultAzureKeyVaultConfig(fmt.Sprintf("%s-staging-vault", projectName)),
	}
}

func DefaultProductionFramework(projectName string) *FrameworkArgs {
	return &FrameworkArgs{
		ProjectName:     projectName,
		Environment:     "production",
		DatabaseConfig:  DefaultPostgreSQLConfig(fmt.Sprintf("%s_production", projectName), "production-db.example.com"),
		MessagingConfig: DefaultRabbitMQConfig("production-mq.example.com"),
		StorageConfig:   DefaultAzureBlobConfig("productionstorage", "production_key", fmt.Sprintf("%s-production", projectName)),
		VaultConfig:     DefaultAzureKeyVaultConfig(fmt.Sprintf("%s-production-vault", projectName)),
	}
}

type ComponentRegistry struct {
	components map[string]interface{}
}

func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		components: make(map[string]interface{}),
	}
}

func (r *ComponentRegistry) Register(name string, component interface{}) {
	r.components[name] = component
}

func (r *ComponentRegistry) Get(name string) (interface{}, bool) {
	component, exists := r.components[name]
	return component, exists
}

func (r *ComponentRegistry) List() []string {
	names := make([]string, 0, len(r.components))
	for name := range r.components {
		names = append(names, name)
	}
	return names
}

var DefaultRegistry = NewComponentRegistry()

func RegisterComponent(name string, component interface{}) {
	DefaultRegistry.Register(name, component)
}

func GetComponent(name string) (interface{}, bool) {
	return DefaultRegistry.Get(name)
}

func ListComponents() []string {
	return DefaultRegistry.List()
}