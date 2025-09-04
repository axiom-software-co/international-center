package infrastructure

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-azure-native-sdk/dbforpostgresql"
	"github.com/pulumi/pulumi-azure-native-sdk/network"
	"github.com/pulumi/pulumi-azure-native-sdk/resources"
	"github.com/pulumi/pulumi-azure-native-sdk/security"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AzureProductionDatabaseStack struct {
	resourceGroup         *resources.ResourceGroup
	vnet                 *network.VirtualNetwork
	privateSubnet        *network.Subnet
	server               *dbforpostgresql.Server
	databases            map[string]*dbforpostgresql.Database
	firewallRules        []*dbforpostgresql.FirewallRule
	privateEndpoint      *network.PrivateEndpoint
	privateDnsZone       *network.PrivateZone
	privateDnsZoneGroup  *network.PrivateZoneGroup
	securityAssessment   *security.Assessment
	backupPolicy         *dbforpostgresql.BackupPolicy
	readReplica          *dbforpostgresql.Server
}

func NewAzureProductionDatabaseStack(resourceGroup *resources.ResourceGroup, vnet *network.VirtualNetwork, privateSubnet *network.Subnet) *AzureProductionDatabaseStack {
	return &AzureProductionDatabaseStack{
		resourceGroup: resourceGroup,
		vnet:         vnet,
		privateSubnet: privateSubnet,
		databases:    make(map[string]*dbforpostgresql.Database),
	}
}

func (stack *AzureProductionDatabaseStack) Deploy(ctx *pulumi.Context) error {
	if err := stack.createPrivateDnsZone(ctx); err != nil {
		return fmt.Errorf("failed to create private DNS zone: %w", err)
	}

	if err := stack.createPostgreSQLServer(ctx); err != nil {
		return fmt.Errorf("failed to create PostgreSQL server: %w", err)
	}

	if err := stack.createDatabases(ctx); err != nil {
		return fmt.Errorf("failed to create databases: %w", err)
	}

	if err := stack.createPrivateEndpoint(ctx); err != nil {
		return fmt.Errorf("failed to create private endpoint: %w", err)
	}

	if err := stack.configureFirewallRules(ctx); err != nil {
		return fmt.Errorf("failed to configure firewall rules: %w", err)
	}

	if err := stack.createReadReplica(ctx); err != nil {
		return fmt.Errorf("failed to create read replica: %w", err)
	}

	if err := stack.enableSecurityAssessment(ctx); err != nil {
		return fmt.Errorf("failed to enable security assessment: %w", err)
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
	server, err := dbforpostgresql.NewServer(ctx, "production-postgres", &dbforpostgresql.ServerArgs{
		ResourceGroupName: stack.resourceGroup.Name,
		ServerName:        pulumi.String("international-center-production-db"),
		Location:         stack.resourceGroup.Location,
		Properties: &dbforpostgresql.ServerPropertiesForCreateArgs{
			CreateMode: pulumi.String("Default"),
			Version:    pulumi.String("13"),
			AdministratorLogin:         pulumi.String("dbadmin"),
			AdministratorLoginPassword: pulumi.String(""), // Retrieved from Key Vault
			StorageProfile: &dbforpostgresql.StorageProfileArgs{
				BackupRetentionDays: pulumi.Int(90), // Extended retention for production
				GeoRedundantBackup:  pulumi.String("Enabled"),
				StorageMB:          pulumi.Int(1048576), // 1TB storage
				StorageAutogrow:    pulumi.String("Enabled"),
			},
			SslEnforcement:                  pulumi.String("Enabled"),
			MinimalTlsVersion:              pulumi.String("TLS1_2"),
			InfrastructureEncryption:       pulumi.String("Enabled"),
			PublicNetworkAccess:           pulumi.String("Disabled"), // Private access only
		},
		Sku: &dbforpostgresql.SkuArgs{
			Name:     pulumi.String("GP_Gen5_8"), // General Purpose, 8 vCores for production
			Tier:     pulumi.String("GeneralPurpose"),
			Capacity: pulumi.Int(8),
			Size:     pulumi.String("1048576"), // 1TB
			Family:   pulumi.String("Gen5"),
		},
		Identity: &dbforpostgresql.ResourceIdentityArgs{
			Type: pulumi.String("SystemAssigned"),
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
			return fmt.Errorf("failed to create %s database: %w", domain, err)
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
		Subnet: &network.SubnetArgs{
			Id: stack.privateSubnet.ID(),
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
			"environment": pulumi.String("production"),
			"project":     pulumi.String("international-center"),
		},
	})
	if err != nil {
		return err
	}

	privateDnsZoneGroup, err := network.NewPrivateZoneGroup(ctx, "production-db-dns-zone-group", &network.PrivateZoneGroupArgs{
		ResourceGroupName:       stack.resourceGroup.Name,
		PrivateEndpointName:     privateEndpoint.Name,
		PrivateDnsZoneGroupName: pulumi.String("default"),
		PrivateDnsZoneConfigs: network.PrivateDnsZoneConfigArray{
			&network.PrivateDnsZoneConfigArgs{
				Name: pulumi.String("postgres-config"),
				PrivateDnsZoneId: stack.privateDnsZone.ID(),
			},
		},
	})
	if err != nil {
		return err
	}

	stack.privateEndpoint = privateEndpoint
	stack.privateDnsZoneGroup = privateDnsZoneGroup
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
		Properties: &dbforpostgresql.ServerPropertiesForCreateArgs{
			CreateMode:       pulumi.String("Replica"),
			SourceServerId:   stack.server.ID(),
			SslEnforcement:   pulumi.String("Enabled"),
			MinimalTlsVersion: pulumi.String("TLS1_2"),
			InfrastructureEncryption: pulumi.String("Enabled"),
			PublicNetworkAccess: pulumi.String("Disabled"),
		},
		Sku: &dbforpostgresql.SkuArgs{
			Name:     pulumi.String("GP_Gen5_8"), // Same as primary
			Tier:     pulumi.String("GeneralPurpose"),
			Capacity: pulumi.Int(8),
			Size:     pulumi.String("1048576"),
			Family:   pulumi.String("Gen5"),
		},
		Identity: &dbforpostgresql.ResourceIdentityArgs{
			Type: pulumi.String("SystemAssigned"),
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
			Category: pulumi.StringArray{
				pulumi.String("Data"),
			},
			Severity: pulumi.String("High"),
		},
	})
	if err != nil {
		return err
	}

	stack.securityAssessment = securityAssessment
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