package migration

import (
	"testing"

	"github.com/axiom-software-co/international-center/src/cicd/components"
	"github.com/axiom-software-co/international-center/src/cicd/shared"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSchemaValidator_BusinessDomain validates business domain schema matches markdown specification
func TestSchemaValidator_BusinessDomain(t *testing.T) {
	environments := []string{"development", "staging", "production"}
	
	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			framework := shared.NewContractTestingFramework("international-center", "schema-business-"+env+"-test")
			
			framework.RunComponentContractTest(t, env, func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, testEnv string) error {
				// Deploy database component
				databaseOutputs, err := components.DeployDatabase(ctx, cfg, testEnv)
				require.NoError(t, err)
				
				// Initialize schema validator
				validator, err := NewSchemaValidator(ctx, cfg, testEnv, databaseOutputs)
				require.NoError(t, err)
				
				// Validate business domain schema against markdown specification
				result, err := validator.ValidateDomainSchema(ctx, "business")
				require.NoError(t, err)
				
				// Verify schema validation results
				result.IsValid.ApplyT(func(isValid bool) error {
					assert.True(t, isValid, "Business domain schema should match markdown specification in %s environment", testEnv)
					return nil
				})
				
				result.RequiredTables.ApplyT(func(tables []interface{}) error {
					tableNames := make([]string, len(tables))
					for i, t := range tables {
						tableNames[i] = t.(string)
					}
					
					// Business domain should have these tables per business-TABLE.md
					expectedTables := []string{
						"business_profiles",
						"business_categories", 
						"business_locations",
						"business_hours",
						"business_contacts",
						"business_reviews",
					}
					
					for _, expectedTable := range expectedTables {
						assert.Contains(t, tableNames, expectedTable, "Business domain should have %s table", expectedTable)
					}
					return nil
				})
				
				return nil
			})
		})
	}
}

// TestSchemaValidator_DonationsDomain validates donations domain schema matches markdown specification  
func TestSchemaValidator_DonationsDomain(t *testing.T) {
	environments := []string{"development", "staging", "production"}
	
	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			framework := shared.NewContractTestingFramework("international-center", "schema-donations-"+env+"-test")
			
			framework.RunComponentContractTest(t, env, func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, testEnv string) error {
				// Deploy database component
				databaseOutputs, err := components.DeployDatabase(ctx, cfg, testEnv)
				require.NoError(t, err)
				
				// Initialize schema validator
				validator, err := NewSchemaValidator(ctx, cfg, testEnv, databaseOutputs)
				require.NoError(t, err)
				
				// Validate donations domain schema
				result, err := validator.ValidateDomainSchema(ctx, "donations")
				require.NoError(t, err)
				
				// Verify schema validation results
				result.RequiredTables.ApplyT(func(tables []interface{}) error {
					tableNames := make([]string, len(tables))
					for i, t := range tables {
						tableNames[i] = t.(string)
					}
					
					// Donations domain should have these tables per donations-TABLE.md
					expectedTables := []string{
						"donation_campaigns",
						"donation_transactions", 
						"donation_receipts",
						"donation_goals",
						"donation_categories",
					}
					
					for _, expectedTable := range expectedTables {
						assert.Contains(t, tableNames, expectedTable, "Donations domain should have %s table", expectedTable)
					}
					return nil
				})
				
				return nil
			})
		})
	}
}

// TestSchemaValidator_EventsDomain validates events domain schema matches markdown specification
func TestSchemaValidator_EventsDomain(t *testing.T) {
	environments := []string{"development", "staging", "production"}
	
	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			framework := shared.NewContractTestingFramework("international-center", "schema-events-"+env+"-test")
			
			framework.RunComponentContractTest(t, env, func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, testEnv string) error {
				// Deploy database component
				databaseOutputs, err := components.DeployDatabase(ctx, cfg, testEnv)
				require.NoError(t, err)
				
				// Initialize schema validator
				validator, err := NewSchemaValidator(ctx, cfg, testEnv, databaseOutputs)
				require.NoError(t, err)
				
				// Validate events domain schema
				result, err := validator.ValidateDomainSchema(ctx, "events")
				require.NoError(t, err)
				
				// Verify schema validation results
				result.RequiredTables.ApplyT(func(tables []interface{}) error {
					tableNames := make([]string, len(tables))
					for i, t := range tables {
						tableNames[i] = t.(string)
					}
					
					// Events domain should have these tables per events-TABLE.md
					expectedTables := []string{
						"events",
						"event_categories",
						"event_registrations", 
						"event_speakers",
						"event_venues",
						"event_schedules",
					}
					
					for _, expectedTable := range expectedTables {
						assert.Contains(t, tableNames, expectedTable, "Events domain should have %s table", expectedTable)
					}
					return nil
				})
				
				return nil
			})
		})
	}
}

// TestSchemaValidator_AllDomains validates all domain schemas match markdown specifications
func TestSchemaValidator_AllDomains(t *testing.T) {
	domains := []string{"business", "donations", "events", "media", "news", "research", "services", "volunteers"}
	
	framework := shared.NewContractTestingFramework("international-center", "schema-all-domains-test")
	
	framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, testEnv string) error {
		// Deploy database component
		databaseOutputs, err := components.DeployDatabase(ctx, cfg, testEnv)
		require.NoError(t, err)
		
		// Initialize schema validator
		validator, err := NewSchemaValidator(ctx, cfg, testEnv, databaseOutputs)
		require.NoError(t, err)
		
		for _, domain := range domains {
			// Validate each domain schema
			result, err := validator.ValidateDomainSchema(ctx, domain)
			require.NoError(t, err, "Should validate %s domain schema", domain)
			
			// Verify each domain has valid schema
			result.IsValid.ApplyT(func(isValid bool) error {
				assert.True(t, isValid, "Domain %s should have valid schema matching markdown specification", domain)
				return nil
			})
		}
		
		return nil
	})
}

// TestSchemaValidator_EnvironmentConsistency validates schema consistency across environments
func TestSchemaValidator_EnvironmentConsistency(t *testing.T) {
	domains := []string{"business", "donations", "events", "media", "news", "research", "services", "volunteers"}
	environments := []string{"development", "staging", "production"}
	
	for _, domain := range domains {
		t.Run("Domain_"+domain, func(t *testing.T) {
			results := make(map[string]*SchemaValidationResult)
			
			for _, env := range environments {
				framework := shared.NewContractTestingFramework("international-center", "schema-consistency-"+domain+"-"+env+"-test")
				
				framework.RunComponentContractTest(t, env, func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, testEnv string) error {
					// Deploy database component
					databaseOutputs, err := components.DeployDatabase(ctx, cfg, testEnv)
					require.NoError(t, err)
					
					// Initialize schema validator
					validator, err := NewSchemaValidator(ctx, cfg, testEnv, databaseOutputs)
					require.NoError(t, err)
					
					// Validate domain schema
					result, err := validator.ValidateDomainSchema(ctx, domain)
					require.NoError(t, err)
					
					results[testEnv] = result
					return nil
				})
			}
			
			// Verify environment consistency after all environments are validated
			if devResult, devOk := results["development"]; devOk {
				if stagingResult, stagingOk := results["staging"]; stagingOk {
					if prodResult, prodOk := results["production"]; prodOk {
						// Verify environment consistency
						pulumi.All(devResult.RequiredTables, stagingResult.RequiredTables, prodResult.RequiredTables).ApplyT(func(args []interface{}) error {
							devTables := args[0].([]interface{})
							stagingTables := args[1].([]interface{})
							prodTables := args[2].([]interface{})
							
							assert.Equal(t, len(devTables), len(stagingTables), "Development and staging should have same number of tables for domain %s", domain)
							assert.Equal(t, len(stagingTables), len(prodTables), "Staging and production should have same number of tables for domain %s", domain)
							
							return nil
						})
					}
				}
			}
		})
	}
}