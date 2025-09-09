package infrastructure

import (
	"fmt"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/integrations"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type InfrastructureArgs struct {
	Environment string
}

type InfrastructureComponent struct {
	pulumi.ResourceState

	DatabaseEndpoint       pulumi.StringOutput `pulumi:"databaseEndpoint"`
	StorageEndpoint        pulumi.StringOutput `pulumi:"storageEndpoint"`
	VaultEndpoint          pulumi.StringOutput `pulumi:"vaultEndpoint"`
	MessagingEndpoint      pulumi.StringOutput `pulumi:"messagingEndpoint"`
	ObservabilityEndpoint  pulumi.StringOutput `pulumi:"observabilityEndpoint"`
	HealthCheckEnabled     pulumi.BoolOutput   `pulumi:"healthCheckEnabled"`
	SecurityPolicies       pulumi.BoolOutput   `pulumi:"securityPolicies"`
	AuditLogging          pulumi.BoolOutput   `pulumi:"auditLogging"`
	MigrationStatus       pulumi.StringOutput `pulumi:"migrationStatus"`
	SchemaVersion         pulumi.StringOutput `pulumi:"schemaVersion"`
	MigrationsApplied     pulumi.IntOutput    `pulumi:"migrationsApplied"`
	ValidationStatus      pulumi.StringOutput `pulumi:"validationStatus"`
}

func NewInfrastructureComponent(ctx *pulumi.Context, name string, args *InfrastructureArgs, opts ...pulumi.ResourceOption) (*InfrastructureComponent, error) {
	component := &InfrastructureComponent{}
	
	// Safe registration for mock contexts
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:infrastructure:Infrastructure", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	// Infrastructure components only define cloud resources
	// Runtime operations are handled by the operations project

	// Deploy database component
	database, err := NewDatabaseComponent(ctx, "database", &DatabaseArgs{
		Config:      DefaultPostgreSQLConfig("international_center_"+args.Environment, "localhost"),
		Environment: args.Environment,
		ProjectName: "international-center",
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy storage component
	storage, err := NewStorageComponent(ctx, "storage", &StorageArgs{
		Config:      DefaultLocalStorageConfig(fmt.Sprintf("./storage/international-center-%s", args.Environment)),
		Environment: args.Environment,
		ProjectName: "international-center",
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy vault component
	vault, err := NewVaultComponent(ctx, "vault", &VaultArgs{
		Config:      DefaultHashiCorpVaultConfig("localhost"),
		Environment: args.Environment,
		ProjectName: "international-center",
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy messaging (RabbitMQ) component
	messaging, err := NewMessagingComponent(ctx, "messaging", &MessagingArgs{
		Config:      DefaultRabbitMQConfig("localhost"),
		Environment: args.Environment,
		ProjectName: "international-center",
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Deploy Grafana Cloud observability integration
	var grafanaConfig *integrations.GrafanaCloudConfig
	switch args.Environment {
	case "development":
		grafanaConfig = integrations.DevelopmentGrafanaCloudConfig()
	case "staging":
		grafanaConfig = integrations.DefaultGrafanaCloudConfig("staging-stack", "1", "staging-token")
	case "production":
		grafanaConfig = integrations.ProductionGrafanaCloudConfig("prod-stack", "1", "prod-token")
	default:
		grafanaConfig = integrations.DevelopmentGrafanaCloudConfig()
	}
	
	observability, err := integrations.NewGrafanaCloudComponent(ctx, "observability", &integrations.GrafanaCloudArgs{
		Config:      grafanaConfig,
		Environment: args.Environment,
		ProjectName: "international-center",
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
	component.DatabaseEndpoint = database.ConnectionString
	component.StorageEndpoint = storage.ConnectionString
	component.VaultEndpoint = vault.VaultAddress
	component.MessagingEndpoint = messaging.Endpoint
	component.ObservabilityEndpoint = observability.GrafanaURL
	component.HealthCheckEnabled = healthCheckEnabled
	component.SecurityPolicies = securityPolicies
	component.AuditLogging = auditLogging
	component.MigrationStatus = database.MigrationStatus
	component.SchemaVersion = database.SchemaVersion
	component.MigrationsApplied = database.MigrationsApplied
	component.ValidationStatus = database.ValidationStatus

	// Register outputs (only if context supports it)
	if canRegister(ctx) {
		ctx.Export("infrastructure:database_endpoint", component.DatabaseEndpoint)
		ctx.Export("infrastructure:storage_endpoint", component.StorageEndpoint)
		ctx.Export("infrastructure:vault_endpoint", component.VaultEndpoint)
		ctx.Export("infrastructure:messaging_endpoint", component.MessagingEndpoint)
		ctx.Export("infrastructure:observability_endpoint", component.ObservabilityEndpoint)
		ctx.Export("infrastructure:migration_status", component.MigrationStatus)
		ctx.Export("infrastructure:schema_version", component.SchemaVersion)
		ctx.Export("infrastructure:migrations_applied", component.MigrationsApplied)
		ctx.Export("infrastructure:validation_status", component.ValidationStatus)
	}

	if canRegister(ctx) {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"databaseEndpoint":       component.DatabaseEndpoint,
			"storageEndpoint":        component.StorageEndpoint,
			"vaultEndpoint":          component.VaultEndpoint,
			"messagingEndpoint":      component.MessagingEndpoint,
			"observabilityEndpoint":  component.ObservabilityEndpoint,
			"healthCheckEnabled":     component.HealthCheckEnabled,
			"securityPolicies":       component.SecurityPolicies,
			"auditLogging":          component.AuditLogging,
			"migrationStatus":       component.MigrationStatus,
			"schemaVersion":         component.SchemaVersion,
			"migrationsApplied":     component.MigrationsApplied,
			"validationStatus":      component.ValidationStatus,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}


func canRegister(ctx *pulumi.Context) bool {
	if ctx == nil {
		return false
	}
	
	// Use a defer/recover pattern to safely test if registration works
	canRegisterSafely := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				// If panic occurred, registration is not safe
				canRegisterSafely = false
			}
		}()
		
		// Try to detect if this is a real Pulumi context vs a mock
		// Mock contexts created with &pulumi.Context{} will panic on export
		// Real contexts will have internal state initialized
		// We use a simple test - try to export a dummy value like canExport does
		testOutput := pulumi.String("test").ToStringOutput()
		ctx.Export("__test_register_capability", testOutput)
		canRegisterSafely = true
	}()
	
	return canRegisterSafely
}

