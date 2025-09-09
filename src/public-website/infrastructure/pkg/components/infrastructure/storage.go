package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type StorageArgs struct {
	Environment string
}

type StorageComponent struct {
	pulumi.ResourceState

	ConnectionString pulumi.StringOutput `pulumi:"connectionString"`
	ContainerName    pulumi.StringOutput `pulumi:"containerName"`
	HealthEndpoint   pulumi.StringOutput `pulumi:"healthEndpoint"`
}

func NewStorageComponent(ctx *pulumi.Context, name string, args *StorageArgs, opts ...pulumi.ResourceOption) (*StorageComponent, error) {
	component := &StorageComponent{}
	
	err := ctx.RegisterComponentResource("international-center:infrastructure:Storage", name, component, opts...)
	if err != nil {
		return nil, err
	}

	var connectionString, containerName, healthEndpoint pulumi.StringOutput

	switch args.Environment {
	case "development":
		connectionString = pulumi.String("DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1").ToStringOutput()
		containerName = pulumi.String("international-center-dev").ToStringOutput()
		healthEndpoint = pulumi.String("http://127.0.0.1:10000/health").ToStringOutput()
	case "staging":
		connectionString = pulumi.String("DefaultEndpointsProtocol=https;AccountName=internationalcenterstaging;AccountKey=staging_key;EndpointSuffix=core.windows.net").ToStringOutput()
		containerName = pulumi.String("international-center-staging").ToStringOutput()
		healthEndpoint = pulumi.String("https://internationalcenterstaging.blob.core.windows.net/health").ToStringOutput()
	case "production":
		connectionString = pulumi.String("DefaultEndpointsProtocol=https;AccountName=internationalcenterprod;AccountKey=production_key;EndpointSuffix=core.windows.net").ToStringOutput()
		containerName = pulumi.String("international-center-production").ToStringOutput()
		healthEndpoint = pulumi.String("https://internationalcenterprod.blob.core.windows.net/health").ToStringOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.ConnectionString = connectionString
	component.ContainerName = containerName
	component.HealthEndpoint = healthEndpoint

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"connectionString": component.ConnectionString,
		"containerName":    component.ContainerName,
		"healthEndpoint":   component.HealthEndpoint,
	}); err != nil {
		return nil, err
	}

	return component, nil
}