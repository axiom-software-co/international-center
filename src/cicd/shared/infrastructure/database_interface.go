package infrastructure

import (
	"context"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type DatabaseStack interface {
	Deploy(ctx context.Context) (DatabaseDeployment, error)
	GetConnectionString() pulumi.StringOutput
	GetConnectionInfo() (string, int, string, string)
	ValidateDeployment(ctx context.Context, deployment DatabaseDeployment) error
}

type DatabaseDeployment interface {
	GetPrimaryEndpoint() pulumi.StringOutput
	GetReplicationEndpoints() []pulumi.StringOutput
	GetNetworkResources() DatabaseNetworkResources
	GetBackupConfiguration() BackupConfig
}

type DatabaseNetworkResources interface {
	GetNetworkID() pulumi.StringOutput
	GetSubnetIDs() []pulumi.StringOutput
	GetSecurityGroupIDs() []pulumi.StringOutput
}

type DatabaseFactory interface {
	CreateDatabaseStack(ctx *pulumi.Context, config *config.Config, environment string) DatabaseStack
}

type DatabaseSchemaManager interface {
	CreateDatabaseSchemas(ctx context.Context, deployment DatabaseDeployment) error
	CreateIndexes(ctx context.Context, deployment DatabaseDeployment) error
	ValidateSchemas(ctx context.Context, deployment DatabaseDeployment) error
}

type DatabaseConfiguration struct {
	Environment        string
	DatabaseName       string
	Username          string
	Host              string
	Port              int
	SSLMode           string
	ConnectionTimeout int
	MaxConnections    int
	BackupRetention   int
	GeoReplication    bool
	HighAvailability  bool
	ReadReplicas      []string
}

// BackupConfig represents backup configuration for database deployment
type BackupConfig struct {
	Enabled           bool
	RetentionDays     int
	BackupInterval    string
	StorageLocation   string
	EncryptionEnabled bool
}

func GetDatabaseConfiguration(environment string, config *config.Config) *DatabaseConfiguration {
	switch environment {
	case "development":
		return &DatabaseConfiguration{
			Environment:        "development",
			DatabaseName:       config.Require("postgres_db"),
			Username:          config.Require("postgres_user"),
			Host:              config.Get("postgres_host"),
			Port:              config.RequireInt("postgres_port"),
			SSLMode:           "disable",
			ConnectionTimeout: 30,
			MaxConnections:    100,
			BackupRetention:   7,
			GeoReplication:    false,
			HighAvailability:  false,
			ReadReplicas:      []string{},
		}
	case "staging":
		return &DatabaseConfiguration{
			Environment:        "staging",
			DatabaseName:       "international_center_staging",
			Username:          "dbadmin",
			Host:              "",
			Port:              5432,
			SSLMode:           "require",
			ConnectionTimeout: 45,
			MaxConnections:    200,
			BackupRetention:   35,
			GeoReplication:    true,
			HighAvailability:  false,
			ReadReplicas:      []string{},
		}
	case "production":
		return &DatabaseConfiguration{
			Environment:        "production",
			DatabaseName:       "international_center_production",
			Username:          "dbadmin",
			Host:              "",
			Port:              5432,
			SSLMode:           "require",
			ConnectionTimeout: 60,
			MaxConnections:    500,
			BackupRetention:   90,
			GeoReplication:    true,
			HighAvailability:  true,
			ReadReplicas:      []string{"West US 2"},
		}
	default:
		return &DatabaseConfiguration{
			Environment:        environment,
			DatabaseName:       "international_center_" + environment,
			Username:          "dbadmin",
			Host:              "",
			Port:              5432,
			SSLMode:           "require",
			ConnectionTimeout: 30,
			MaxConnections:    100,
			BackupRetention:   30,
			GeoReplication:    false,
			HighAvailability:  false,
			ReadReplicas:      []string{},
		}
	}
}