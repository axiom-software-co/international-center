package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi-azure-native-sdk/dbforpostgresql/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/network/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	shared "github.com/axiom-software-co/international-center/src/deployer/shared/infrastructure"
)

type AzureDatabaseStack struct {
	pulumi.ComponentResource
	resourceGroup    *resources.ResourceGroup
	server          *dbforpostgresql.Server
	databases       map[string]*dbforpostgresql.Database
	firewallRules   []*dbforpostgresql.FirewallRule
	privateEndpoint *network.PrivateEndpoint
	vnet            *network.VirtualNetwork
	subnet          *network.Subnet
	errorHandler    *shared.ErrorHandler
	
	// Outputs
	DatabaseEndpoint pulumi.StringOutput `pulumi:"databaseEndpoint"`
	ConnectionString pulumi.StringOutput `pulumi:"connectionString"`
	NetworkID       pulumi.StringOutput `pulumi:"networkId"`
}

func NewAzureDatabaseStack(ctx *pulumi.Context, resourceGroup *resources.ResourceGroup) *AzureDatabaseStack {
	errorHandler := shared.NewErrorHandler(ctx, "staging", "database")
	
	component := &AzureDatabaseStack{
		resourceGroup: resourceGroup,
		databases:     make(map[string]*dbforpostgresql.Database),
		errorHandler:  errorHandler,
	}
	
	err := ctx.RegisterComponentResource("custom:staging:AzureDatabaseStack", "staging-database-stack", component)
	if err != nil {
		resourceErr := shared.NewResourceError("register_component", "database", "staging", "AzureDatabaseStack", err)
		errorHandler.HandleError(resourceErr)
		panic(err) // Still panic for critical component registration failures
	}
	
	return component
}

func (stack *AzureDatabaseStack) Deploy(ctx *pulumi.Context) error {
	if err := stack.createVirtualNetwork(ctx); err != nil {
		return fmt.Errorf("failed to create virtual network: %w", err)
	}

	if err := stack.createPostgreSQLServer(ctx); err != nil {
		return fmt.Errorf("failed to create PostgreSQL server: %w", err)
	}

	if err := stack.createDatabases(ctx); err != nil {
		return fmt.Errorf("failed to create databases: %w", err)
	}

	if err := stack.configureFirewallRules(ctx); err != nil {
		return fmt.Errorf("failed to configure firewall rules: %w", err)
	}

	if err := stack.createPrivateEndpoint(ctx); err != nil {
		return fmt.Errorf("failed to create private endpoint: %w", err)
	}

	return nil
}

func (stack *AzureDatabaseStack) createVirtualNetwork(ctx *pulumi.Context) error {
	vnet, err := network.NewVirtualNetwork(ctx, "staging-vnet", &network.VirtualNetworkArgs{
		ResourceGroupName:    stack.resourceGroup.Name,
		VirtualNetworkName:   pulumi.String("international-center-staging-vnet"),
		Location:            stack.resourceGroup.Location,
		AddressSpace: &network.AddressSpaceArgs{
			AddressPrefixes: pulumi.StringArray{
				pulumi.String("10.1.0.0/16"),
			},
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return shared.NewNetworkError("create_virtual_network", "database", "staging", "international-center-staging-vnet", err)
	}

	subnet, err := network.NewSubnet(ctx, "staging-db-subnet", &network.SubnetArgs{
		ResourceGroupName:  stack.resourceGroup.Name,
		VirtualNetworkName: vnet.Name,
		SubnetName:         pulumi.String("database-subnet"),
		AddressPrefix:      pulumi.String("10.1.1.0/24"),
		ServiceEndpoints: network.ServiceEndpointPropertiesFormatArray{
			&network.ServiceEndpointPropertiesFormatArgs{
				Service: pulumi.String("Microsoft.Sql"),
			},
		},
		PrivateEndpointNetworkPolicies:    pulumi.String("Disabled"),
		PrivateLinkServiceNetworkPolicies: pulumi.String("Enabled"),
	})
	if err != nil {
		return shared.NewNetworkError("create_virtual_network", "database", "staging", "international-center-staging-vnet", err)
	}

	stack.vnet = vnet
	stack.subnet = subnet
	return nil
}

func (stack *AzureDatabaseStack) createPostgreSQLServer(ctx *pulumi.Context) error {
	// TODO: Fix PostgreSQL server configuration for Azure Native SDK v1.104.0 
	// The API has changed significantly and requires different configuration pattern
	server, err := dbforpostgresql.NewServer(ctx, "staging-postgres", &dbforpostgresql.ServerArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		ServerName:        pulumi.String("international-center-staging-db"),
		Location:         stack.resourceGroup.Location,
		Tags: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return shared.NewNetworkError("create_virtual_network", "database", "staging", "international-center-staging-vnet", err)
	}

	stack.server = server
	return nil
}

func (stack *AzureDatabaseStack) createDatabases(ctx *pulumi.Context) error {
	domainDatabases := []string{"identity", "content", "services"}
	
	for _, domain := range domainDatabases {
		if err := stack.createDomainDatabase(ctx, domain); err != nil {
			return fmt.Errorf("failed to create %s database: %w", domain, err)
		}
	}

	return nil
}

func (stack *AzureDatabaseStack) createDomainDatabase(ctx *pulumi.Context, domainName string) error {
	database, err := dbforpostgresql.NewDatabase(ctx, fmt.Sprintf("staging-%s-db", domainName), &dbforpostgresql.DatabaseArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		ServerName:       stack.server.Name,
		DatabaseName:     pulumi.String(fmt.Sprintf("%s_staging", domainName)),
		Charset:          pulumi.String("UTF8"),
		Collation:       pulumi.String("en_US.utf8"),
	})
	if err != nil {
		return shared.NewNetworkError("create_virtual_network", "database", "staging", "international-center-staging-vnet", err)
	}

	stack.databases[domainName] = database
	return nil
}

func (stack *AzureDatabaseStack) configureFirewallRules(ctx *pulumi.Context) error {
	containerAppsRule, err := dbforpostgresql.NewFirewallRule(ctx, "staging-container-apps-rule", &dbforpostgresql.FirewallRuleArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		ServerName:       stack.server.Name,
		FirewallRuleName: pulumi.String("AllowContainerApps"),
		StartIpAddress:   pulumi.String("0.0.0.0"), // Container Apps dynamic IPs
		EndIpAddress:     pulumi.String("255.255.255.255"),
	})
	if err != nil {
		return shared.NewNetworkError("create_virtual_network", "database", "staging", "international-center-staging-vnet", err)
	}

	azureServicesRule, err := dbforpostgresql.NewFirewallRule(ctx, "staging-azure-services-rule", &dbforpostgresql.FirewallRuleArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		ServerName:       stack.server.Name,
		FirewallRuleName: pulumi.String("AllowAzureServices"),
		StartIpAddress:   pulumi.String("0.0.0.0"),
		EndIpAddress:     pulumi.String("0.0.0.0"), // Special case for Azure services
	})
	if err != nil {
		return shared.NewNetworkError("create_virtual_network", "database", "staging", "international-center-staging-vnet", err)
	}

	stack.firewallRules = append(stack.firewallRules, containerAppsRule, azureServicesRule)
	return nil
}

func (stack *AzureDatabaseStack) createPrivateEndpoint(ctx *pulumi.Context) error {
	privateEndpoint, err := network.NewPrivateEndpoint(ctx, "staging-db-private-endpoint", &network.PrivateEndpointArgs{
		ResourceGroupName:   stack.resourceGroup.Name,
		PrivateEndpointName: pulumi.String("international-center-staging-db-pe"),
		Location:           stack.resourceGroup.Location,
		Subnet: &network.SubnetTypeArgs{
			Id: stack.subnet.ID(),
		},
		PrivateLinkServiceConnections: network.PrivateLinkServiceConnectionArray{
			&network.PrivateLinkServiceConnectionArgs{
				Name:                 pulumi.String("database-connection"),
				PrivateLinkServiceId: stack.server.ID(),
				GroupIds: pulumi.StringArray{
					pulumi.String("postgresqlServer"),
				},
			},
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return shared.NewNetworkError("create_virtual_network", "database", "staging", "international-center-staging-vnet", err)
	}

	stack.privateEndpoint = privateEndpoint
	return nil
}

func (stack *AzureDatabaseStack) GetConnectionString() pulumi.StringOutput {
	return pulumi.Sprintf("Server=%s.postgres.database.azure.com;Database=postgres;Port=5432;User Id=dbadmin@%s;Password=%s;Ssl Mode=Require;",
		stack.server.Name,
		stack.server.Name,
		"", // Password retrieved from Key Vault
	)
}

func (stack *AzureDatabaseStack) GetDomainConnectionString(domain string) pulumi.StringOutput {
	return pulumi.Sprintf("Server=%s.postgres.database.azure.com;Database=%s_staging;Port=5432;User Id=dbadmin@%s;Password=%s;Ssl Mode=Require;",
		stack.server.Name,
		domain,
		stack.server.Name,
		"", // Password retrieved from Key Vault
	)
}

func (stack *AzureDatabaseStack) GetServer() *dbforpostgresql.Server {
	return stack.server
}

func (stack *AzureDatabaseStack) GetDatabase(domain string) *dbforpostgresql.Database {
	return stack.databases[domain]
}

func (stack *AzureDatabaseStack) GetVirtualNetwork() *network.VirtualNetwork {
	return stack.vnet
}

func (stack *AzureDatabaseStack) GetSubnet() *network.Subnet {
	return stack.subnet
}