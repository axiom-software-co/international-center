package infrastructure

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type InfrastructureArgs struct {
	Environment              string
	DatabaseConnectionString pulumi.StringInput
	StorageConnectionString  pulumi.StringInput
	VaultAddress            pulumi.StringInput
	RabbitMQEndpoint        pulumi.StringInput
	GrafanaURL              pulumi.StringInput
}

type InfrastructureComponent struct {
	pulumi.ResourceState

	DatabaseConnectionString pulumi.StringOutput `pulumi:"databaseConnectionString"`
	StorageConnectionString  pulumi.StringOutput `pulumi:"storageConnectionString"`
	VaultAddress            pulumi.StringOutput `pulumi:"vaultAddress"`
	RabbitMQEndpoint        pulumi.StringOutput `pulumi:"rabbitMQEndpoint"`
	GrafanaURL              pulumi.StringOutput `pulumi:"grafanaURL"`
	HealthCheckEnabled      pulumi.BoolOutput   `pulumi:"healthCheckEnabled"`
	SecurityPolicies        pulumi.BoolOutput   `pulumi:"securityPolicies"`
	AuditLogging           pulumi.BoolOutput   `pulumi:"auditLogging"`
}

func NewInfrastructureComponent(ctx *pulumi.Context, name string, args *InfrastructureArgs, opts ...pulumi.ResourceOption) (*InfrastructureComponent, error) {
	component := &InfrastructureComponent{}
	
	err := ctx.RegisterComponentResource("international-center:infrastructure:Infrastructure", name, component, opts...)
	if err != nil {
		return nil, err
	}

	// Deploy database component
	database, err := NewDatabaseComponent(ctx, "database", &DatabaseArgs{
		Environment: args.Environment,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy storage component
	storage, err := NewStorageComponent(ctx, "storage", &StorageArgs{
		Environment: args.Environment,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy vault component
	vault, err := NewVaultComponent(ctx, "vault", &VaultArgs{
		Environment: args.Environment,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy messaging (RabbitMQ) component
	messaging, err := NewMessagingComponent(ctx, "messaging", &MessagingArgs{
		Environment: args.Environment,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy observability component
	observability, err := NewObservabilityComponent(ctx, "observability", &ObservabilityArgs{
		Environment: args.Environment,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Configure environment-specific settings
	var healthCheckEnabled, securityPolicies, auditLogging pulumi.BoolOutput
	
	switch args.Environment {
	case "development":
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
		securityPolicies = pulumi.Bool(true).ToBoolOutput()
		auditLogging = pulumi.Bool(false).ToBoolOutput()
	case "staging":
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
		securityPolicies = pulumi.Bool(false).ToBoolOutput()
		auditLogging = pulumi.Bool(true).ToBoolOutput()
	case "production":
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
		securityPolicies = pulumi.Bool(true).ToBoolOutput()
		auditLogging = pulumi.Bool(true).ToBoolOutput()
	default:
		healthCheckEnabled = pulumi.Bool(true).ToBoolOutput()
		securityPolicies = pulumi.Bool(true).ToBoolOutput()
		auditLogging = pulumi.Bool(true).ToBoolOutput()
	}

	// Set component outputs
	component.DatabaseConnectionString = database.ConnectionString
	component.StorageConnectionString = storage.ConnectionString
	component.VaultAddress = vault.VaultAddress
	component.RabbitMQEndpoint = messaging.Endpoint
	component.GrafanaURL = observability.GrafanaURL
	component.HealthCheckEnabled = healthCheckEnabled
	component.SecurityPolicies = securityPolicies
	component.AuditLogging = auditLogging

	// Register outputs
	ctx.Export("infrastructure:database_connection_string", component.DatabaseConnectionString)
	ctx.Export("infrastructure:storage_connection_string", component.StorageConnectionString)
	ctx.Export("infrastructure:vault_address", component.VaultAddress)
	ctx.Export("infrastructure:rabbitmq_endpoint", component.RabbitMQEndpoint)
	ctx.Export("infrastructure:grafana_url", component.GrafanaURL)

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"databaseConnectionString": component.DatabaseConnectionString,
		"storageConnectionString":  component.StorageConnectionString,
		"vaultAddress":            component.VaultAddress,
		"rabbitMQEndpoint":        component.RabbitMQEndpoint,
		"grafanaURL":              component.GrafanaURL,
		"healthCheckEnabled":      component.HealthCheckEnabled,
		"securityPolicies":        component.SecurityPolicies,
		"auditLogging":           component.AuditLogging,
	}); err != nil {
		return nil, err
	}

	return component, nil
}