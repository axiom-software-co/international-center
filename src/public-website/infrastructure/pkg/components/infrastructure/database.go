package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DatabaseArgs struct {
	Environment string
}

type DatabaseComponent struct {
	pulumi.ResourceState

	ConnectionString pulumi.StringOutput `pulumi:"connectionString"`
	DatabaseName     pulumi.StringOutput `pulumi:"databaseName"`
	HealthEndpoint   pulumi.StringOutput `pulumi:"healthEndpoint"`
}

func NewDatabaseComponent(ctx *pulumi.Context, name string, args *DatabaseArgs, opts ...pulumi.ResourceOption) (*DatabaseComponent, error) {
	component := &DatabaseComponent{}
	
	err := ctx.RegisterComponentResource("international-center:infrastructure:Database", name, component, opts...)
	if err != nil {
		return nil, err
	}

	var connectionString, databaseName, healthEndpoint pulumi.StringOutput

	switch args.Environment {
	case "development":
		connectionString = pulumi.String("postgresql://postgres:password@localhost:5432/international_center_dev").ToStringOutput()
		databaseName = pulumi.String("international_center_dev").ToStringOutput()
		healthEndpoint = pulumi.String("http://localhost:5432/health").ToStringOutput()
	case "staging":
		connectionString = pulumi.String("postgresql://postgres:staging_password@staging-db.azurewebsites.net:5432/international_center_staging").ToStringOutput()
		databaseName = pulumi.String("international_center_staging").ToStringOutput()
		healthEndpoint = pulumi.String("https://staging-db.azurewebsites.net/health").ToStringOutput()
	case "production":
		connectionString = pulumi.String("postgresql://postgres:production_password@production-db.azurewebsites.net:5432/international_center_production").ToStringOutput()
		databaseName = pulumi.String("international_center_production").ToStringOutput()
		healthEndpoint = pulumi.String("https://production-db.azurewebsites.net/health").ToStringOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.ConnectionString = connectionString
	component.DatabaseName = databaseName
	component.HealthEndpoint = healthEndpoint

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"connectionString": component.ConnectionString,
		"databaseName":     component.DatabaseName,
		"healthEndpoint":   component.HealthEndpoint,
	}); err != nil {
		return nil, err
	}

	return component, nil
}