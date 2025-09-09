package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/containers"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/health"
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeploymentIntegration_Development(t *testing.T) {
	// Skip integration tests unless explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests skipped - set RUN_INTEGRATION_TESTS=true to enable")
	}

	// This test requires the entire development environment to be up
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

			// Validate development-specific configurations
			databaseEndpoint := stack.Outputs["infrastructure:database_endpoint"].(string)
			assert.Contains(t, databaseEndpoint, "postgres:5432")

			storageEndpoint := stack.Outputs["infrastructure:storage_endpoint"].(string)
			assert.Contains(t, storageEndpoint, "azurite:10000")

			websiteURL := stack.Outputs["website:url"].(string)
			assert.Contains(t, websiteURL, "localhost")
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

func TestDeploymentIntegration_HealthValidation(t *testing.T) {
	// Skip integration tests unless explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests skipped - set RUN_INTEGRATION_TESTS=true to enable")
	}

	// This test validates health endpoints after deployment
	cwd, err := os.Getwd()
	require.NoError(t, err)

	// Get the infrastructure directory path
	infraDir := filepath.Join(cwd, "../..")

	// Integration test configuration for health validation
	opts := &integration.ProgramTestOptions{
		Dir:          infraDir,
		Dependencies: []string{"github.com/pulumi/pulumi/sdk/v3"},
		Config: map[string]string{
			"environment": "development",
		},
		Quick:                true,
		SkipRefresh:          true,
		ExpectRefreshChanges: false,
		SkipPreview:          true,
		SkipUpdate:           true,
		SkipExportImport:     true,
		SkipEmptyPreviewUpdate: true,
		DestroyOnCleanup:       false,
		DebugUpdates:          false,
		Verbose:               false,
		ExtraRuntimeValidation: func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
			// Validate that health check endpoints are accessible
			// This would require the entire development environment to be running

			// Note: Actual HTTP health checks would be implemented here
			// For now, just validate that health check configurations are present
			websiteURL := stack.Outputs["website:url"].(string)
			assert.NotEmpty(t, websiteURL)

			publicGatewayURL := stack.Outputs["services:public_gateway_url"].(string)
			assert.NotEmpty(t, publicGatewayURL)

			adminGatewayURL := stack.Outputs["services:admin_gateway_url"].(string)
			assert.NotEmpty(t, adminGatewayURL)

			// Validate health check enabled flag
			assert.Equal(t, true, stack.Outputs["website:health_check_enabled"])
		},
	}

	// Run the integration test
	integration.ProgramTest(t, opts)
}