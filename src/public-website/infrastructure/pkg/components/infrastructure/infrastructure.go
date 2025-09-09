package infrastructure

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/containers"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/health"
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

	// For development environment, start actual containers (but not during tests)
	if args.Environment == "development" && canRegister(ctx) && !isTestEnvironment() {
		if err := startDevelopmentContainers(); err != nil {
			log.Printf("Warning: Failed to start development containers: %v", err)
			// Continue with deployment even if containers fail to start
		}
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
	component.DatabaseEndpoint = database.ConnectionString
	component.StorageEndpoint = storage.ConnectionString
	component.VaultEndpoint = vault.VaultAddress
	component.MessagingEndpoint = messaging.Endpoint
	component.ObservabilityEndpoint = observability.GrafanaURL
	component.HealthCheckEnabled = healthCheckEnabled
	component.SecurityPolicies = securityPolicies
	component.AuditLogging = auditLogging

	// Register outputs (only if context supports it)
	if canRegister(ctx) {
		ctx.Export("infrastructure:database_endpoint", component.DatabaseEndpoint)
		ctx.Export("infrastructure:storage_endpoint", component.StorageEndpoint)
		ctx.Export("infrastructure:vault_endpoint", component.VaultEndpoint)
		ctx.Export("infrastructure:messaging_endpoint", component.MessagingEndpoint)
		ctx.Export("infrastructure:observability_endpoint", component.ObservabilityEndpoint)
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
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}

func startDevelopmentContainers() error {
	log.Printf("Starting development infrastructure containers and health server")
	
	co := containers.NewContainerOrchestrator("development")
	co.SetupDevelopmentInfrastructure()
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	
	// Start containers
	if err := co.StartAllContainers(ctx); err != nil {
		return err
	}
	
	// Start health check server
	healthServer := health.NewHealthServer(8080, "development")
	if err := healthServer.Start(ctx); err != nil {
		log.Printf("Warning: Failed to start health server: %v", err)
		// Continue even if health server fails
	}
	
	log.Printf("Development environment started successfully")
	return nil
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

func isTestEnvironment() bool {
	// Check if we're running in a test environment
	if os.Getenv("GO_TESTING") == "true" {
		return true
	}
	
	// Check if any test-related environment variables are set
	if os.Getenv("SKIP_CONTAINERS") == "true" {
		return true
	}
	
	// Check for common test patterns in command line
	args := os.Args
	for _, arg := range args {
		if arg == "-test.v" || arg == "-test.run" || arg == "-test.timeout" {
			return true
		}
	}
	
	return false
}