package components

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
)

// TestVaultComponent_DevelopmentEnvironment tests vault component for development environment
func TestVaultComponent_DevelopmentEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployVault(ctx, cfg, "development")
		if err != nil {
			return err
		}

		// Verify development environment generates local Vault container configuration
		pulumi.All(outputs.VaultType, outputs.VaultAddress, outputs.AuthMethod, outputs.SecretEngine).ApplyT(func(args []interface{}) error {
			vaultType := args[0].(string)
			vaultAddress := args[1].(string)
			authMethod := args[2].(string)
			secretEngine := args[3].(string)

			assert.Equal(t, "podman_vault", vaultType, "Development should use local Vault container")
			assert.Contains(t, vaultAddress, "http://127.0.0.1:8200", "Should use local Vault address")
			assert.Equal(t, "dev_token", authMethod, "Should use development token auth")
			assert.Equal(t, "secret", secretEngine, "Should use KV secret engine")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &VaultMocks{}))

	assert.NoError(t, err)
}

// TestVaultComponent_StagingEnvironment tests vault component for staging environment
func TestVaultComponent_StagingEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployVault(ctx, cfg, "staging")
		if err != nil {
			return err
		}

		// Verify staging environment generates HashiCorp Vault Cloud configuration
		pulumi.All(outputs.VaultType, outputs.VaultAddress, outputs.AuthMethod, outputs.ClusterTier).ApplyT(func(args []interface{}) error {
			vaultType := args[0].(string)
			vaultAddress := args[1].(string)
			authMethod := args[2].(string)
			clusterTier := args[3].(string)

			assert.Equal(t, "hashicorp_cloud", vaultType, "Staging should use HashiCorp Vault Cloud")
			assert.Contains(t, vaultAddress, "vault.cloud.hashicorp.com", "Should use HashiCorp Cloud address")
			assert.Equal(t, "service_principal", authMethod, "Should use service principal auth")
			assert.Equal(t, "development", clusterTier, "Should configure staging cluster tier")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &VaultMocks{}))

	assert.NoError(t, err)
}

// TestVaultComponent_ProductionEnvironment tests vault component for production environment
func TestVaultComponent_ProductionEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployVault(ctx, cfg, "production")
		if err != nil {
			return err
		}

		// Verify production environment generates HashiCorp Vault Cloud with production features
		pulumi.All(outputs.VaultType, outputs.VaultAddress, outputs.AuthMethod, outputs.ClusterTier, outputs.AuditEnabled).ApplyT(func(args []interface{}) error {
			vaultType := args[0].(string)
			vaultAddress := args[1].(string)
			authMethod := args[2].(string)
			clusterTier := args[3].(string)
			auditEnabled := args[4].(bool)

			assert.Equal(t, "hashicorp_cloud", vaultType, "Production should use HashiCorp Vault Cloud")
			assert.Contains(t, vaultAddress, "vault.cloud.hashicorp.com", "Should use HashiCorp Cloud address")
			assert.Equal(t, "service_principal", authMethod, "Should use service principal auth")
			assert.Equal(t, "standard", clusterTier, "Should configure production cluster tier")
			assert.True(t, auditEnabled, "Should enable audit logging for production")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &VaultMocks{}))

	assert.NoError(t, err)
}

// TestVaultComponent_EnvironmentParity tests that all environments support required features
func TestVaultComponent_EnvironmentParity(t *testing.T) {
	environments := []string{"development", "staging", "production"}

	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				cfg := config.New(ctx, "")

				outputs, err := DeployVault(ctx, cfg, env)
				if err != nil {
					return err
				}

				// Verify all environments provide required outputs
				pulumi.All(outputs.VaultAddress, outputs.AuthMethod, outputs.SecretEngine).ApplyT(func(args []interface{}) error {
					vaultAddress := args[0].(string)
					authMethod := args[1].(string)
					secretEngine := args[2].(string)

					assert.NotEmpty(t, vaultAddress, "All environments should provide vault address")
					assert.NotEmpty(t, authMethod, "All environments should provide auth method")
					assert.NotEmpty(t, secretEngine, "All environments should provide secret engine")
					return nil
				})

				return nil
			}, pulumi.WithMocks("test", "stack", &VaultMocks{}))

			assert.NoError(t, err)
		})
	}
}

// VaultMocks provides mocks for Pulumi testing
type VaultMocks struct{}

func (mocks *VaultMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}

	switch args.TypeToken {
	case "docker:index/container:Container":
		outputs["name"] = resource.NewStringProperty("vault-dev")
		outputs["image"] = resource.NewStringProperty("hashicorp/vault:latest")
		outputs["ports"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{
				"internal": resource.NewNumberProperty(8200),
				"external": resource.NewNumberProperty(8200),
			}),
		})

	case "hcp:index/vaultCluster:VaultCluster":
		outputs["vaultPublicEndpointUrl"] = resource.NewStringProperty("https://international-center-vault.vault.cloud.hashicorp.com:8200")
		outputs["clusterId"] = resource.NewStringProperty("international-center-vault")
		outputs["tier"] = resource.NewStringProperty("standard")
	}

	return args.Name + "_id", outputs, nil
}

func (mocks *VaultMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}
	return outputs, nil
}