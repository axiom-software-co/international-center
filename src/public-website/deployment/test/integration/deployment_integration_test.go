package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeploymentIntegration_Development(t *testing.T) {
	// Skip integration tests unless explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests skipped - set RUN_INTEGRATION_TESTS=true to enable")
	}

	// This test validates infrastructure deployment for development environment
	cwd, err := os.Getwd()
	require.NoError(t, err)

	// Get the infrastructure directory path
	infraDir := filepath.Join(cwd, "../..")

	// Integration test configuration
	opts := &integration.ProgramTestOptions{
		Dir:          infraDir,
		Dependencies: []string{"github.com/pulumi/pulumi/sdk/v3"},
		Config: map[string]string{
			"environment": "development",
		},
		Secrets: map[string]string{
			// These would be set from environment variables in actual deployment
		},
		Quick:                true,
		SkipRefresh:          false,
		ExpectRefreshChanges: false,
		SkipPreview:          false,
		SkipUpdate:           false,
		SkipExportImport:     true,
		SkipEmptyPreviewUpdate: false,
		DestroyOnCleanup:       true,
		DebugUpdates:          false,
		Verbose:               true,
		ExtraRuntimeValidation: func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
			// Validate infrastructure component outputs
			assert.NotEmpty(t, stack.Outputs["infrastructure:database_endpoint"])
			assert.NotEmpty(t, stack.Outputs["infrastructure:storage_endpoint"])
			assert.NotEmpty(t, stack.Outputs["infrastructure:vault_endpoint"])
			assert.NotEmpty(t, stack.Outputs["infrastructure:messaging_endpoint"])
			assert.NotEmpty(t, stack.Outputs["infrastructure:observability_endpoint"])

			// Validate platform component outputs
			assert.NotEmpty(t, stack.Outputs["platform:dapr_endpoint"])
			assert.NotEmpty(t, stack.Outputs["platform:orchestration_endpoint"])

			// Validate services component outputs
			assert.NotEmpty(t, stack.Outputs["services:public_gateway_url"])
			assert.NotEmpty(t, stack.Outputs["services:admin_gateway_url"])
			assert.NotEmpty(t, stack.Outputs["services:content_service_url"])
			assert.NotEmpty(t, stack.Outputs["services:inquiries_service_url"])
			assert.NotEmpty(t, stack.Outputs["services:notifications_service_url"])

			// Validate website component outputs
			assert.NotEmpty(t, stack.Outputs["website:url"])
			assert.Equal(t, "podman_container", stack.Outputs["website:deployment_type"])
			assert.Equal(t, false, stack.Outputs["website:cdn_enabled"])
			assert.Equal(t, false, stack.Outputs["website:ssl_enabled"])

			// Validate admin portal component outputs
			assert.NotEmpty(t, stack.Outputs["admin-portal:url"])
			assert.Equal(t, "podman_container", stack.Outputs["admin-portal:deployment_type"])
			assert.NotEmpty(t, stack.Outputs["admin-portal:api_endpoint"])

			// Validate development-specific configurations
			databaseEndpoint := stack.Outputs["infrastructure:database_endpoint"].(string)
			assert.Contains(t, databaseEndpoint, "postgres:5432")

			storageEndpoint := stack.Outputs["infrastructure:storage_endpoint"].(string)
			assert.Contains(t, storageEndpoint, "azurite:10000")

			websiteURL := stack.Outputs["website:url"].(string)
			assert.Contains(t, websiteURL, "localhost")

			// Validate development admin portal URL
			adminPortalURL := stack.Outputs["admin-portal:url"].(string)
			assert.Contains(t, adminPortalURL, "localhost")
			assert.Contains(t, adminPortalURL, "8055")

			adminAPIEndpoint := stack.Outputs["admin-portal:api_endpoint"].(string)
			assert.Contains(t, adminAPIEndpoint, "localhost")
			assert.Contains(t, adminAPIEndpoint, "9000")

			t.Log("Infrastructure deployment validation completed successfully")
		},
	}

	// Run the integration test
	integration.ProgramTest(t, opts)
}


func TestDeploymentIntegration_Staging(t *testing.T) {
	// Skip integration tests unless explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests skipped - set RUN_INTEGRATION_TESTS=true to enable")
	}

	// This test requires proper Azure credentials and staging environment setup
	cwd, err := os.Getwd()
	require.NoError(t, err)

	// Get the infrastructure directory path
	infraDir := filepath.Join(cwd, "../..")

	// Integration test configuration for staging
	opts := &integration.ProgramTestOptions{
		Dir:          infraDir,
		Dependencies: []string{"github.com/pulumi/pulumi/sdk/v3"},
		Config: map[string]string{
			"environment": "staging",
		},
		Secrets: map[string]string{
			// These would be set from environment variables in actual deployment
		},
		Quick:                true,
		SkipRefresh:          false,
		ExpectRefreshChanges: false,
		SkipPreview:          false,
		SkipUpdate:           false,
		SkipExportImport:     true,
		SkipEmptyPreviewUpdate: false,
		DestroyOnCleanup:       true,
		DebugUpdates:          false,
		Verbose:               true,
		ExtraRuntimeValidation: func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
			// Validate staging-specific configurations
			websiteURL := stack.Outputs["website:url"].(string)
			assert.Contains(t, websiteURL, "staging")
			assert.Contains(t, websiteURL, "azurecontainerapp.io")

			// Validate CDN and SSL are enabled in staging
			assert.Equal(t, true, stack.Outputs["website:cdn_enabled"])
			assert.Equal(t, true, stack.Outputs["website:ssl_enabled"])
			assert.Equal(t, "container_app", stack.Outputs["website:deployment_type"])

			// Validate staging gateway URLs
			publicGatewayURL := stack.Outputs["services:public_gateway_url"].(string)
			assert.Contains(t, publicGatewayURL, "staging")
			assert.Contains(t, publicGatewayURL, "azurecontainerapp.io")

			adminGatewayURL := stack.Outputs["services:admin_gateway_url"].(string)
			assert.Contains(t, adminGatewayURL, "staging")
			assert.Contains(t, adminGatewayURL, "azurecontainerapp.io")

			// Validate staging admin portal
			adminPortalURL := stack.Outputs["admin-portal:url"].(string)
			assert.Contains(t, adminPortalURL, "staging")
			assert.Contains(t, adminPortalURL, "azurecontainerapp.io")
			assert.Equal(t, "container_app", stack.Outputs["admin-portal:deployment_type"])
		},
	}

	// Run the integration test
	integration.ProgramTest(t, opts)
}

func TestDeploymentIntegration_Production(t *testing.T) {
	// Skip integration tests unless explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests skipped - set RUN_INTEGRATION_TESTS=true to enable")
	}

	// This test requires proper Azure credentials and production environment setup
	cwd, err := os.Getwd()
	require.NoError(t, err)

	// Get the infrastructure directory path
	infraDir := filepath.Join(cwd, "../..")

	// Integration test configuration for production
	opts := &integration.ProgramTestOptions{
		Dir:          infraDir,
		Dependencies: []string{"github.com/pulumi/pulumi/sdk/v3"},
		Config: map[string]string{
			"environment": "production",
		},
		Secrets: map[string]string{
			// These would be set from environment variables in actual deployment
		},
		Quick:                true,
		SkipRefresh:          false,
		ExpectRefreshChanges: false,
		SkipPreview:          false,
		SkipUpdate:           false,
		SkipExportImport:     true,
		SkipEmptyPreviewUpdate: false,
		DestroyOnCleanup:       false, // Don't auto-destroy production
		DebugUpdates:          false,
		Verbose:               true,
		ExtraRuntimeValidation: func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
			// Validate production-specific configurations
			websiteURL := stack.Outputs["website:url"].(string)
			assert.Contains(t, websiteURL, "production")
			assert.Contains(t, websiteURL, "azurecontainerapp.io")

			// Validate enhanced production settings
			assert.Equal(t, true, stack.Outputs["website:cdn_enabled"])
			assert.Equal(t, true, stack.Outputs["website:ssl_enabled"])
			assert.Equal(t, "container_app", stack.Outputs["website:deployment_type"])

			// Validate production gateway URLs
			publicGatewayURL := stack.Outputs["services:public_gateway_url"].(string)
			assert.Contains(t, publicGatewayURL, "production")
			assert.Contains(t, publicGatewayURL, "azurecontainerapp.io")

			adminGatewayURL := stack.Outputs["services:admin_gateway_url"].(string)
			assert.Contains(t, adminGatewayURL, "production")
			assert.Contains(t, adminGatewayURL, "azurecontainerapp.io")

			// Validate production admin portal
			adminPortalURL := stack.Outputs["admin-portal:url"].(string)
			assert.Contains(t, adminPortalURL, "production")
			assert.Contains(t, adminPortalURL, "azurecontainerapp.io")
			assert.Equal(t, "container_app", stack.Outputs["admin-portal:deployment_type"])

			// Validate all required production services are deployed
			assert.NotEmpty(t, stack.Outputs["infrastructure:database_endpoint"])
			assert.NotEmpty(t, stack.Outputs["infrastructure:storage_endpoint"])
			assert.NotEmpty(t, stack.Outputs["infrastructure:vault_endpoint"])
			assert.NotEmpty(t, stack.Outputs["infrastructure:messaging_endpoint"])
			assert.NotEmpty(t, stack.Outputs["infrastructure:observability_endpoint"])
		},
	}

	// Run the integration test
	integration.ProgramTest(t, opts)
}

