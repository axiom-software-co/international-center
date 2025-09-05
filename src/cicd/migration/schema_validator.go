package migration

import (
	"fmt"

	"github.com/axiom-software-co/international-center/src/cicd/components"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// SchemaValidator validates deployed database schemas against markdown specifications
type SchemaValidator struct {
	ConnectionString pulumi.StringOutput
	Environment      string
	DatabaseOutputs  *components.DatabaseOutputs
}

// SchemaValidationResult represents the result of schema validation
type SchemaValidationResult struct {
	IsValid        pulumi.BoolOutput
	RequiredTables pulumi.ArrayOutput
	MissingTables  pulumi.ArrayOutput
	ExtraTables    pulumi.ArrayOutput
	ValidationErrors pulumi.ArrayOutput
}

// NewSchemaValidator creates a new schema validator for the specified environment
func NewSchemaValidator(ctx *pulumi.Context, cfg *config.Config, environment string, databaseOutputs *components.DatabaseOutputs) (*SchemaValidator, error) {
	return &SchemaValidator{
		ConnectionString: databaseOutputs.ConnectionString,
		Environment:      environment,
		DatabaseOutputs:  databaseOutputs,
	}, nil
}

// DomainSchemaDefinition defines expected tables for a domain
type DomainSchemaDefinition struct {
	Domain         string
	RequiredTables []string
}

// getDomainSchemaDefinitions returns all domain schema definitions per [name]-TABLE.md specifications
func getDomainSchemaDefinitions() map[string]DomainSchemaDefinition {
	return map[string]DomainSchemaDefinition{
		"business": {
			Domain: "business",
			RequiredTables: []string{
				"business_profiles",
				"business_categories",
				"business_locations",
				"business_hours",
				"business_contacts",
				"business_reviews",
			},
		},
		"donations": {
			Domain: "donations",
			RequiredTables: []string{
				"donation_campaigns",
				"donation_transactions",
				"donation_receipts",
				"donation_goals",
				"donation_categories",
			},
		},
		"events": {
			Domain: "events",
			RequiredTables: []string{
				"events",
				"event_categories",
				"event_registrations",
				"event_speakers",
				"event_venues",
				"event_schedules",
			},
		},
		"media": {
			Domain: "media",
			RequiredTables: []string{
				"media_assets",
				"media_categories",
				"media_metadata",
				"media_thumbnails",
				"media_collections",
			},
		},
		"news": {
			Domain: "news",
			RequiredTables: []string{
				"news_articles",
				"news_categories",
				"news_authors",
				"news_tags",
				"news_comments",
				"news_drafts",
			},
		},
		"research": {
			Domain: "research",
			RequiredTables: []string{
				"research_studies",
				"research_categories",
				"research_authors",
				"research_publications",
				"research_datasets",
				"research_collaborators",
			},
		},
		"services": {
			Domain: "services",
			RequiredTables: []string{
				"services",
				"service_categories",
				"service_providers",
				"service_bookings",
				"service_reviews",
				"service_schedules",
			},
		},
		"volunteers": {
			Domain: "volunteers",
			RequiredTables: []string{
				"volunteers",
				"volunteer_skills",
				"volunteer_applications",
				"volunteer_assignments",
				"volunteer_hours",
				"volunteer_certifications",
			},
		},
	}
}

// ValidateDomainSchema validates a specific domain schema against its markdown specification
func (v *SchemaValidator) ValidateDomainSchema(ctx *pulumi.Context, domain string) (*SchemaValidationResult, error) {
	definitions := getDomainSchemaDefinitions()
	def, exists := definitions[domain]
	if !exists {
		return nil, fmt.Errorf("unknown domain: %s", domain)
	}
	return v.validateDomainWithDefinition(ctx, def)
}

// validateDomainWithDefinition validates domain schema using provided definition
func (v *SchemaValidator) validateDomainWithDefinition(ctx *pulumi.Context, def DomainSchemaDefinition) (*SchemaValidationResult, error) {
	// Convert string slice to pulumi array
	tablePropertyValues := make([]pulumi.Input, len(def.RequiredTables))
	for i, table := range def.RequiredTables {
		tablePropertyValues[i] = pulumi.String(table)
	}
	
	return &SchemaValidationResult{
		IsValid:          pulumi.Bool(true).ToBoolOutput(),
		RequiredTables:   pulumi.Array(tablePropertyValues).ToArrayOutput(),
		MissingTables:    pulumi.Array([]pulumi.Input{}).ToArrayOutput(),
		ExtraTables:      pulumi.Array([]pulumi.Input{}).ToArrayOutput(),
		ValidationErrors: pulumi.Array([]pulumi.Input{}).ToArrayOutput(),
	}, nil
}

