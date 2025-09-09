package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DatabaseProvider string

const (
	DatabaseProviderPostgreSQL DatabaseProvider = "postgresql"
	DatabaseProviderMySQL      DatabaseProvider = "mysql"
	DatabaseProviderSQLite     DatabaseProvider = "sqlite"
)

type DatabaseConfig struct {
	Provider           DatabaseProvider
	DatabaseName       string
	Host               string
	Port               int
	Username           string
	Password           string
	SSLMode            string
	MaxConnections     int
	ConnectionTimeout  int
	HealthCheckPath    string
	HealthCheckPort    int
	BackupEnabled      bool
	BackupRetention    int
	EncryptionEnabled  bool
	AdditionalParams   map[string]string
}

type DatabaseArgs struct {
	Config      *DatabaseConfig
	Environment string
	ProjectName string
}

type DatabaseComponent struct {
	pulumi.ResourceState

	ConnectionString  pulumi.StringOutput `pulumi:"connectionString"`
	DatabaseName      pulumi.StringOutput `pulumi:"databaseName"`
	Host              pulumi.StringOutput `pulumi:"host"`
	Port              pulumi.IntOutput    `pulumi:"port"`
	HealthEndpoint    pulumi.StringOutput `pulumi:"healthEndpoint"`
	Provider          pulumi.StringOutput `pulumi:"provider"`
	MigrationStatus   pulumi.StringOutput `pulumi:"migrationStatus"`
	SchemaVersion     pulumi.StringOutput `pulumi:"schemaVersion"`
	MigrationsApplied pulumi.IntOutput    `pulumi:"migrationsApplied"`
	ValidationStatus  pulumi.StringOutput `pulumi:"validationStatus"`
}

func NewDatabaseComponent(ctx *pulumi.Context, name string, args *DatabaseArgs, opts ...pulumi.ResourceOption) (*DatabaseComponent, error) {
	component := &DatabaseComponent{}
	
	if ctx != nil {
		err := ctx.RegisterComponentResource("framework:database:Database", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	config := args.Config
	if config == nil {
		return nil, fmt.Errorf("database config is required")
	}

	connectionString := buildConnectionString(config)
	healthEndpoint := buildHealthEndpoint(config)

	component.ConnectionString = pulumi.String(connectionString).ToStringOutput()
	component.DatabaseName = pulumi.String(config.DatabaseName).ToStringOutput()
	component.Host = pulumi.String(config.Host).ToStringOutput()
	component.Port = pulumi.Int(config.Port).ToIntOutput()
	component.HealthEndpoint = pulumi.String(healthEndpoint).ToStringOutput()
	component.Provider = pulumi.String(string(config.Provider)).ToStringOutput()

	// Execute migrations after database configuration
	migrationArgs := DefaultMigrationArgs(args.Environment, component.ConnectionString)
	migrationComponent, err := NewMigrationComponent(ctx, name+"-migrations", migrationArgs, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("failed to create migration component: %w", err)
	}

	// Set migration outputs
	component.MigrationStatus = migrationComponent.MigrationStatus
	component.SchemaVersion = migrationComponent.SchemaVersion
	component.MigrationsApplied = migrationComponent.MigrationsApplied
	component.ValidationStatus = migrationComponent.ValidationStatus

	if ctx != nil {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"connectionString":  component.ConnectionString,
			"databaseName":      component.DatabaseName,
			"host":              component.Host,
			"port":              component.Port,
			"healthEndpoint":    component.HealthEndpoint,
			"provider":          component.Provider,
			"migrationStatus":   component.MigrationStatus,
			"schemaVersion":     component.SchemaVersion,
			"migrationsApplied": component.MigrationsApplied,
			"validationStatus":  component.ValidationStatus,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}

func buildConnectionString(config *DatabaseConfig) string {
	var connectionString string
	
	switch config.Provider {
	case DatabaseProviderPostgreSQL:
		connectionString = fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
			config.Username, config.Password, config.Host, config.Port, config.DatabaseName)
		
		if config.SSLMode != "" {
			connectionString += fmt.Sprintf("?sslmode=%s", config.SSLMode)
		}
		
	case DatabaseProviderMySQL:
		connectionString = fmt.Sprintf("mysql://%s:%s@%s:%d/%s",
			config.Username, config.Password, config.Host, config.Port, config.DatabaseName)
			
	case DatabaseProviderSQLite:
		connectionString = fmt.Sprintf("sqlite://%s", config.DatabaseName)
	}
	
	return connectionString
}

func buildHealthEndpoint(config *DatabaseConfig) string {
	if config.HealthCheckPath == "" {
		return ""
	}
	
	port := config.HealthCheckPort
	if port == 0 {
		port = config.Port
	}
	
	return fmt.Sprintf("http://%s:%d%s", config.Host, port, config.HealthCheckPath)
}

func DefaultPostgreSQLConfig(databaseName, host string) *DatabaseConfig {
	return &DatabaseConfig{
		Provider:          DatabaseProviderPostgreSQL,
		DatabaseName:      databaseName,
		Host:              host,
		Port:              5432,
		Username:          "postgres",
		Password:          "password",
		SSLMode:           "prefer",
		MaxConnections:    100,
		ConnectionTimeout: 30,
		HealthCheckPath:   "/health",
		HealthCheckPort:   0,
		BackupEnabled:     true,
		BackupRetention:   7,
		EncryptionEnabled: true,
		AdditionalParams:  make(map[string]string),
	}
}

func DefaultMySQLConfig(databaseName, host string) *DatabaseConfig {
	return &DatabaseConfig{
		Provider:          DatabaseProviderMySQL,
		DatabaseName:      databaseName,
		Host:              host,
		Port:              3306,
		Username:          "root",
		Password:          "password",
		MaxConnections:    100,
		ConnectionTimeout: 30,
		HealthCheckPath:   "/health",
		HealthCheckPort:   0,
		BackupEnabled:     true,
		BackupRetention:   7,
		EncryptionEnabled: true,
		AdditionalParams:  make(map[string]string),
	}
}

func DefaultSQLiteConfig(databaseName string) *DatabaseConfig {
	return &DatabaseConfig{
		Provider:          DatabaseProviderSQLite,
		DatabaseName:      databaseName,
		Host:              "localhost",
		Port:              0,
		Username:          "",
		Password:          "",
		MaxConnections:    1,
		ConnectionTimeout: 5,
		BackupEnabled:     false,
		EncryptionEnabled: false,
		AdditionalParams:  make(map[string]string),
	}
}