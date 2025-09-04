package infrastructure

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-azure-native-sdk/dbforpostgresql"
	"github.com/pulumi/pulumi-azure-native-sdk/network"
	"github.com/pulumi/pulumi-azure-native-sdk/resources"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AzureDatabaseStack struct {
	resourceGroup    *resources.ResourceGroup
	server          *dbforpostgresql.Server
	databases       map[string]*dbforpostgresql.Database
	firewallRules   []*dbforpostgresql.FirewallRule
	privateEndpoint *network.PrivateEndpoint
	vnet            *network.VirtualNetwork
	subnet          *network.Subnet
}

func NewAzureDatabaseStack(resourceGroup *resources.ResourceGroup) *AzureDatabaseStack {
	return &AzureDatabaseStack{
		resourceGroup: resourceGroup,
		databases:     make(map[string]*dbforpostgresql.Database),
	}
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
		return err
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
		return err
	}

	stack.vnet = vnet
	stack.subnet = subnet
	return nil
}

func (stack *AzureDatabaseStack) createPostgreSQLServer(ctx *pulumi.Context) error {
	server, err := dbforpostgresql.NewServer(ctx, "staging-postgres", &dbforpostgresql.ServerArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		ServerName:        pulumi.String("international-center-staging-db"),
		Location:         stack.resourceGroup.Location,
		Properties: &dbforpostgresql.ServerPropertiesForCreateArgs{
			CreateMode: pulumi.String("Default"),
			Version:    pulumi.String("13"),
			AdministratorLogin:         pulumi.String("dbadmin"),
			AdministratorLoginPassword: pulumi.String(""), // Retrieved from Key Vault
			StorageProfile: &dbforpostgresql.StorageProfileArgs{
				BackupRetentionDays: pulumi.Int(35), // Extended retention for staging
				GeoRedundantBackup:  pulumi.String("Enabled"),
				StorageMB:          pulumi.Int(102400), // 100GB
				StorageAutogrow:    pulumi.String("Enabled"),
			},
		},
		Sku: &dbforpostgresql.SkuArgs{
			Name:     pulumi.String("GP_Gen5_4"), // General Purpose, 4 vCores
			Tier:     pulumi.String("GeneralPurpose"),
			Capacity: pulumi.Int(4),
			Size:     pulumi.String("102400"),
			Family:   pulumi.String("Gen5"),
		},
		Tags: pulumi.StringMap{
			"environment": pulumi.String("staging"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
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
		return err
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
		return err
	}

	azureServicesRule, err := dbforpostgresql.NewFirewallRule(ctx, "staging-azure-services-rule", &dbforpostgresql.FirewallRuleArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		ServerName:       stack.server.Name,
		FirewallRuleName: pulumi.String("AllowAzureServices"),
		StartIpAddress:   pulumi.String("0.0.0.0"),
		EndIpAddress:     pulumi.String("0.0.0.0"), // Special case for Azure services
	})
	if err != nil {
		return err
	}

	stack.firewallRules = append(stack.firewallRules, containerAppsRule, azureServicesRule)
	return nil
}

func (stack *AzureDatabaseStack) createPrivateEndpoint(ctx *pulumi.Context) error {
	privateEndpoint, err := network.NewPrivateEndpoint(ctx, "staging-db-private-endpoint", &network.PrivateEndpointArgs{
		ResourceGroupName:   stack.resourceGroup.Name,
		PrivateEndpointName: pulumi.String("international-center-staging-db-pe"),
		Location:           stack.resourceGroup.Location,
		Subnet: &network.SubnetArgs{
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
		return err
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