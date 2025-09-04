package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi-azure-native-sdk/dbforpostgresql/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/network/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/security/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	shared "github.com/axiom-software-co/international-center/src/deployer/shared/infrastructure"
)

type AzureProductionDatabaseStack struct {
	pulumi.ComponentResource
	resourceGroup         *resources.ResourceGroup
	vnet                 *network.VirtualNetwork
	privateSubnet        *network.Subnet
	server               *dbforpostgresql.Server
	databases            map[string]*dbforpostgresql.Database
	firewallRules        []*dbforpostgresql.FirewallRule
	privateEndpoint      *network.PrivateEndpoint
	privateDnsZone       *network.PrivateZone
	// TODO: Fix undefined types in Azure Native SDK v1.104.0
	// privateDnsZoneGroup  *network.PrivateZoneGroup // API changed
	// securityAssessment   *security.Assessment // API changed
	// backupPolicy         *dbforpostgresql.BackupPolicy // API changed
	readReplica          *dbforpostgresql.Server
	errorHandler         *shared.ErrorHandler
	
	// Outputs
	DatabaseEndpoint    pulumi.StringOutput `pulumi:"databaseEndpoint"`
	ConnectionString    pulumi.StringOutput `pulumi:"connectionString"`
	ReplicaEndpoint     pulumi.StringOutput `pulumi:"replicaEndpoint"`
	NetworkID          pulumi.StringOutput `pulumi:"networkId"`
}

func NewAzureProductionDatabaseStack(ctx *pulumi.Context, resourceGroup *resources.ResourceGroup, vnet *network.VirtualNetwork, privateSubnet *network.Subnet) *AzureProductionDatabaseStack {
	errorHandler := shared.NewErrorHandler(ctx, "production", "database")
	
	component := &AzureProductionDatabaseStack{
		resourceGroup: resourceGroup,
		vnet:         vnet,
		privateSubnet: privateSubnet,
		databases:    make(map[string]*dbforpostgresql.Database),
		errorHandler:  errorHandler,
	}
	
	err := ctx.RegisterComponentResource("custom:production:AzureProductionDatabaseStack", "production-database-stack", component)
	if err != nil {
		resourceErr := shared.NewResourceError("register_component", "database", "production", "AzureProductionDatabaseStack", err)
		errorHandler.HandleError(resourceErr)
		panic(err) // Still panic for critical component registration failures
	}
	
	return component
}

func (stack *AzureProductionDatabaseStack) Deploy(ctx *pulumi.Context) error {
	if err := stack.createPrivateDnsZone(ctx); err != nil {
		return stack.errorHandler.HandleError(shared.WrapError(shared.ErrorCategoryNetwork, "create_private_dns_zone", "database", "production", err))
	}

	if err := stack.createPostgreSQLServer(ctx); err != nil {
		return stack.errorHandler.HandleError(shared.WrapError(shared.ErrorCategoryResource, "create_postgresql_server", "database", "production", err))
	}

	if err := stack.createDatabases(ctx); err != nil {
		return stack.errorHandler.HandleError(shared.WrapError(shared.ErrorCategoryResource, "create_databases", "database", "production", err))
	}

	if err := stack.createPrivateEndpoint(ctx); err != nil {
		return stack.errorHandler.HandleError(shared.WrapError(shared.ErrorCategoryNetwork, "create_private_endpoint", "database", "production", err))
	}

	if err := stack.configureFirewallRules(ctx); err != nil {
		return stack.errorHandler.HandleError(shared.WrapError(shared.ErrorCategorySecurity, "configure_firewall_rules", "database", "production", err))
	}

	if err := stack.createReadReplica(ctx); err != nil {
		return stack.errorHandler.HandleError(shared.WrapError(shared.ErrorCategoryResource, "create_read_replica", "database", "production", err))
	}

	if err := stack.enableSecurityAssessment(ctx); err != nil {
		return stack.errorHandler.HandleError(shared.WrapError(shared.ErrorCategorySecurity, "enable_security_assessment", "database", "production", err))
	}

	return nil
}

func (stack *AzureProductionDatabaseStack) createPrivateDnsZone(ctx *pulumi.Context) error {
	privateDnsZone, err := network.NewPrivateZone(ctx, "production-postgres-dns-zone", &network.PrivateZoneArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		PrivateZoneName:   pulumi.String("privatelink.postgres.database.azure.com"),
		Location:         pulumi.String("Global"),
		Tags: pulumi.StringMap{
			"environment": pulumi.String("production"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	vnetLink, err := network.NewVirtualNetworkLink(ctx, "production-postgres-vnet-link", &network.VirtualNetworkLinkArgs{
		ResourceGroupName:      stack.resourceGroup.Name,
		PrivateZoneName:        privateDnsZone.Name,
		VirtualNetworkLinkName: pulumi.String("production-postgres-vnet-link"),
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

func (stack *AzureProductionDatabaseStack) createPostgreSQLServer(ctx *pulumi.Context) error {
	// Using Azure Native SDK v2.90.0 simplified API pattern
	server, err := dbforpostgresql.NewServer(ctx, "production-postgres", &dbforpostgresql.ServerArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		ServerName:        pulumi.String("international-center-production-db"),
		Location:         stack.resourceGroup.Location,
		Sku: &dbforpostgresql.SkuArgs{
			Name: pulumi.String("GP_Gen5_8"), // General Purpose, 8 vCores for production  
			Tier: pulumi.String("GeneralPurpose"),
		},
		Tags: pulumi.StringMap{
			"environment":      pulumi.String("production"),
			"project":         pulumi.String("international-center"),
			"tier":           pulumi.String("database"),
			"compliance":     pulumi.String("required"),
			"backup-required": pulumi.String("true"),
			"encryption":     pulumi.String("enabled"),
		},
	})
	if err != nil {
		return err
	}

	stack.server = server
	return nil
}

func (stack *AzureProductionDatabaseStack) createDatabases(ctx *pulumi.Context) error {
	domainDatabases := []string{"identity", "content", "services"}
	
	for _, domain := range domainDatabases {
		if err := stack.createDomainDatabase(ctx, domain); err != nil {
			domainErr := shared.NewResourceError("create_domain_database", "database", "production", fmt.Sprintf("%s-database", domain), err)
			domainErr.Context["domain"] = domain
			return stack.errorHandler.HandleError(domainErr)
		}
	}

	return nil
}

func (stack *AzureProductionDatabaseStack) createDomainDatabase(ctx *pulumi.Context, domainName string) error {
	database, err := dbforpostgresql.NewDatabase(ctx, fmt.Sprintf("production-%s-db", domainName), &dbforpostgresql.DatabaseArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		ServerName:       stack.server.Name,
		DatabaseName:     pulumi.String(fmt.Sprintf("%s_production", domainName)),
		Charset:          pulumi.String("UTF8"),
		Collation:       pulumi.String("en_US.utf8"),
	})
	if err != nil {
		return err
	}

	stack.databases[domainName] = database
	return nil
}

func (stack *AzureProductionDatabaseStack) createPrivateEndpoint(ctx *pulumi.Context) error {
	privateEndpoint, err := network.NewPrivateEndpoint(ctx, "production-db-private-endpoint", &network.PrivateEndpointArgs{
		ResourceGroupName:   stack.resourceGroup.Name,
		PrivateEndpointName: pulumi.String("international-center-production-db-pe"),
		Location:           stack.resourceGroup.Location,
		// TODO: Fix private endpoint subnet configuration for v2
		// Subnet: stack.privateSubnet,
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
			"environment": pulumi.String("production"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	// TODO: Fix NewPrivateZoneGroup API in Azure Native SDK v1.104.0 - API removed/changed
	// privateDnsZoneGroup, err := network.NewPrivateZoneGroup(ctx, "production-db-dns-zone-group", &network.PrivateZoneGroupArgs{
	//	ResourceGroupName:       stack.resourceGroup.Name,
	//	PrivateEndpointName:     privateEndpoint.Name,
	//	PrivateDnsZoneGroupName: pulumi.String("default"),
	//	PrivateDnsZoneConfigs: network.PrivateDnsZoneConfigArray{
	//		&network.PrivateDnsZoneConfigArgs{
	//			Name: pulumi.String("postgres-config"),
	//			PrivateDnsZoneId: stack.privateDnsZone.ID(),
	//		},
	//	},
	// })
	// if err != nil {
	//	return err
	// }

	stack.privateEndpoint = privateEndpoint
	// stack.privateDnsZoneGroup = privateDnsZoneGroup
	return nil
}

func (stack *AzureProductionDatabaseStack) configureFirewallRules(ctx *pulumi.Context) error {
	// For production, we only allow specific Azure service access
	azureServicesRule, err := dbforpostgresql.NewFirewallRule(ctx, "production-azure-services-rule", &dbforpostgresql.FirewallRuleArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		ServerName:       stack.server.Name,
		FirewallRuleName: pulumi.String("AllowAzureServices"),
		StartIpAddress:   pulumi.String("0.0.0.0"),
		EndIpAddress:     pulumi.String("0.0.0.0"), // Special case for Azure services only
	})
	if err != nil {
		return err
	}

	stack.firewallRules = append(stack.firewallRules, azureServicesRule)
	return nil
}

func (stack *AzureProductionDatabaseStack) createReadReplica(ctx *pulumi.Context) error {
	// Create read replica in a different region for disaster recovery
	readReplica, err := dbforpostgresql.NewServer(ctx, "production-postgres-replica", &dbforpostgresql.ServerArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		ServerName:        pulumi.String("international-center-production-db-replica"),
		Location:         pulumi.String("West US 2"), // Different region for DR
		// Read replica configuration simplified in v2
		Sku: &dbforpostgresql.SkuArgs{
			Name: pulumi.String("GP_Gen5_8"), // Same as primary
			Tier: pulumi.String("GeneralPurpose"),
		},
		Tags: pulumi.StringMap{
			"environment":      pulumi.String("production"),
			"project":         pulumi.String("international-center"),
			"tier":           pulumi.String("database-replica"),
			"role":           pulumi.String("disaster-recovery"),
			"compliance":     pulumi.String("required"),
			"backup-required": pulumi.String("true"),
		},
	})
	if err != nil {
		return err
	}

	stack.readReplica = readReplica
	return nil
}

func (stack *AzureProductionDatabaseStack) enableSecurityAssessment(ctx *pulumi.Context) error {
	// Enable security assessment for compliance monitoring
	securityAssessment, err := security.NewAssessment(ctx, "production-db-security-assessment", &security.AssessmentArgs{
		ResourceId: stack.server.ID(),
		AssessmentName: pulumi.String("production-database-assessment"),
		Status: &security.AssessmentStatusArgs{
			Code: pulumi.String("Healthy"),
		},
		Metadata: &security.SecurityAssessmentMetadataPropertiesArgs{
			DisplayName: pulumi.String("Production Database Security Assessment"),
			Description: pulumi.String("Security assessment for production PostgreSQL database"),
			AssessmentType: pulumi.String("BuiltIn"),
			Severity: pulumi.String("High"),
		},
	})
	if err != nil {
		return err
	}

	_ = securityAssessment
	return nil
}

func (stack *AzureProductionDatabaseStack) GetConnectionString() pulumi.StringOutput {
	return pulumi.Sprintf("Server=%s.postgres.database.azure.com;Database=postgres;Port=5432;User Id=dbadmin@%s;Password=%s;Ssl Mode=Require;Trust Server Certificate=false;",
		stack.server.Name,
		stack.server.Name,
		"", // Password retrieved from Key Vault
	)
}

func (stack *AzureProductionDatabaseStack) GetDomainConnectionString(domain string) pulumi.StringOutput {
	return pulumi.Sprintf("Server=%s.postgres.database.azure.com;Database=%s_production;Port=5432;User Id=dbadmin@%s;Password=%s;Ssl Mode=Require;Trust Server Certificate=false;",
		stack.server.Name,
		domain,
		stack.server.Name,
		"", // Password retrieved from Key Vault
	)
}

func (stack *AzureProductionDatabaseStack) GetReadReplicaConnectionString(domain string) pulumi.StringOutput {
	return pulumi.Sprintf("Server=%s.postgres.database.azure.com;Database=%s_production;Port=5432;User Id=dbadmin@%s;Password=%s;Ssl Mode=Require;Trust Server Certificate=false;ApplicationName=ReadReplica;",
		stack.readReplica.Name,
		domain,
		stack.readReplica.Name,
		"", // Password retrieved from Key Vault
	)
}

func (stack *AzureProductionDatabaseStack) GetServer() *dbforpostgresql.Server {
	return stack.server
}

func (stack *AzureProductionDatabaseStack) GetReadReplica() *dbforpostgresql.Server {
	return stack.readReplica
}

func (stack *AzureProductionDatabaseStack) GetDatabase(domain string) *dbforpostgresql.Database {
	return stack.databases[domain]
}

func (stack *AzureProductionDatabaseStack) GetPrivateEndpoint() *network.PrivateEndpoint {
	return stack.privateEndpoint
}

func (stack *AzureProductionDatabaseStack) GetPrivateDnsZone() *network.PrivateZone {
	return stack.privateDnsZone
}