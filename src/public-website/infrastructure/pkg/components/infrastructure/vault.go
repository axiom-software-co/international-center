package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type VaultArgs struct {
	Environment string
}

type VaultComponent struct {
	pulumi.ResourceState

	VaultAddress   pulumi.StringOutput `pulumi:"vaultAddress"`
	VaultToken     pulumi.StringOutput `pulumi:"vaultToken"`
	HealthEndpoint pulumi.StringOutput `pulumi:"healthEndpoint"`
}

func NewVaultComponent(ctx *pulumi.Context, name string, args *VaultArgs, opts ...pulumi.ResourceOption) (*VaultComponent, error) {
	component := &VaultComponent{}
	
	err := ctx.RegisterComponentResource("international-center:infrastructure:Vault", name, component, opts...)
	if err != nil {
		return nil, err
	}

	var vaultAddress, vaultToken, healthEndpoint pulumi.StringOutput

	switch args.Environment {
	case "development":
		vaultAddress = pulumi.String("http://127.0.0.1:8200").ToStringOutput()
		vaultToken = pulumi.String("dev-token").ToStringOutput()
		healthEndpoint = pulumi.String("http://127.0.0.1:8200/v1/sys/health").ToStringOutput()
	case "staging":
		vaultAddress = pulumi.String("https://vault-staging.azurewebsites.net").ToStringOutput()
		vaultToken = pulumi.String("staging-token").ToStringOutput()
		healthEndpoint = pulumi.String("https://vault-staging.azurewebsites.net/v1/sys/health").ToStringOutput()
	case "production":
		vaultAddress = pulumi.String("https://vault-production.azurewebsites.net").ToStringOutput()
		vaultToken = pulumi.String("production-token").ToStringOutput()
		healthEndpoint = pulumi.String("https://vault-production.azurewebsites.net/v1/sys/health").ToStringOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.VaultAddress = vaultAddress
	component.VaultToken = vaultToken
	component.HealthEndpoint = healthEndpoint

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"vaultAddress":   component.VaultAddress,
		"vaultToken":     component.VaultToken,
		"healthEndpoint": component.HealthEndpoint,
	}); err != nil {
		return nil, err
	}

	return component, nil
}