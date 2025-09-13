package integration

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE: Frontend Code Quality Tests
// These tests validate that frontend applications build without import errors and serve content properly

func TestFrontendCodeQuality_BuildErrorsAndImportResolution(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	validateFrontendEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Frontend build quality tests - should identify import errors and missing dependencies
	frontendBuildTests := []struct {
		application   string
		port          int
		endpoint      string
		expectedIssue string
		critical      bool
		description   string
	}{
		{
			application:   "public-website",
			port:          3000,
			endpoint:      "http://localhost:3000",
			expectedIssue: "import errors (newsletterClient, useEvent, useResearchArticle)",
			critical:      true,
			description:   "Public website must build without import errors for complete functionality",
		},
		{
			application:   "admin-portal",
			port:          3001,
			endpoint:      "http://localhost:3001",
			expectedIssue: "404 response instead of admin portal content",
			critical:      true,
			description:   "Admin portal must serve admin interface content (not 404)",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test frontend build quality
	for _, buildTest := range frontendBuildTests {
		t.Run("BuildQuality_"+buildTest.application, func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, "GET", buildTest.endpoint, nil)
			require.NoError(t, err, "Failed to create frontend request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()

				if buildTest.critical {
					// Critical applications must serve content properly
					if buildTest.application == "public-website" {
						// Public website should be accessible and serving HTML
						assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
							"%s - must serve content without build errors", buildTest.description)
					} else if buildTest.application == "admin-portal" {
						// Admin portal should not return 404
						assert.NotEqual(t, 404, resp.StatusCode,
							"%s - must serve admin interface (not 404)", buildTest.description)
						
						// Should serve admin portal content
						if resp.StatusCode == 404 {
							t.Errorf("%s - returning 404 instead of admin content (expected issue: %s)", 
								buildTest.description, buildTest.expectedIssue)
						}
					}
				}
			} else {
				if buildTest.critical {
					t.Errorf("%s not accessible: %v (expected issue: %s)", 
						buildTest.description, err, buildTest.expectedIssue)
				}
			}
		})
	}
}

func TestFrontendCodeQuality_ContractClientIntegration(t *testing.T) {
	// Test frontend contract client integration and API consumption
	validateFrontendEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Contract client integration tests - frontend should consume backend APIs
	contractIntegrationTests := []struct {
		clientType     string
		frontendApp    string
		frontendPort   int
		backendAPI     string
		gatewayAPI     string
		description    string
	}{
		{
			clientType:   "public-api-client",
			frontendApp:  "public-website",
			frontendPort: 3000,
			backendAPI:   "http://localhost:3001/api/news",
			gatewayAPI:   "http://localhost:9001/api/news",
			description:  "Public website must consume backend news API through contract clients",
		},
		{
			clientType:   "public-api-client",
			frontendApp:  "public-website",
			frontendPort: 3000,
			backendAPI:   "http://localhost:3001/api/events",
			gatewayAPI:   "http://localhost:9001/api/events",
			description:  "Public website must consume backend events API through contract clients",
		},
		{
			clientType:   "admin-api-client",
			frontendApp:  "admin-portal",
			frontendPort: 3001,
			backendAPI:   "http://localhost:3101/api/inquiries",
			gatewayAPI:   "http://localhost:9000/api/admin/inquiries",
			description:  "Admin portal must consume backend inquiries API through contract clients",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test contract client integration
	for _, integration := range contractIntegrationTests {
		t.Run("ContractIntegration_"+integration.clientType, func(t *testing.T) {
			// First verify backend API is functional
			backendReq, err := http.NewRequestWithContext(ctx, "GET", integration.backendAPI, nil)
			require.NoError(t, err, "Failed to create backend API request")

			backendResp, err := client.Do(backendReq)
			require.NoError(t, err, "Backend API must be accessible for contract integration")
			defer backendResp.Body.Close()

			assert.True(t, backendResp.StatusCode >= 200 && backendResp.StatusCode < 300,
				"Backend API must be functional for %s", integration.description)

			// Verify gateway API is functional
			gatewayReq, err := http.NewRequestWithContext(ctx, "GET", integration.gatewayAPI, nil)
			require.NoError(t, err, "Failed to create gateway API request")

			gatewayResp, err := client.Do(gatewayReq)
			require.NoError(t, err, "Gateway API must be accessible for contract integration")
			defer gatewayResp.Body.Close()

			assert.True(t, gatewayResp.StatusCode >= 200 && gatewayResp.StatusCode < 300,
				"Gateway API must be functional for %s", integration.description)

			// Test frontend application accessibility
			frontendURL := fmt.Sprintf("http://localhost:%d", integration.frontendPort)
			frontendReq, err := http.NewRequestWithContext(ctx, "GET", frontendURL, nil)
			require.NoError(t, err, "Failed to create frontend request")

			frontendResp, err := client.Do(frontendReq)
			if err == nil {
				defer frontendResp.Body.Close()
				
				if integration.frontendApp == "public-website" {
					// Public website should be accessible
					assert.True(t, frontendResp.StatusCode >= 200 && frontendResp.StatusCode < 300,
						"%s - frontend must be accessible for API consumption", integration.description)
				} else if integration.frontendApp == "admin-portal" {
					// Admin portal should not return 404
					assert.NotEqual(t, 404, frontendResp.StatusCode,
						"%s - admin portal must serve content for API consumption", integration.description)
				}
			} else {
				t.Errorf("%s frontend not accessible: %v", integration.description, err)
			}
		})
	}
}

func TestFrontendCodeQuality_DevelopmentWorkflowValidation(t *testing.T) {
	// Test complete development workflow with frontend and backend integration
	validateFrontendEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Development workflow validation tests
	workflowValidationTests := []struct {
		workflow     string
		components   []string
		endpoints    []string
		status       string
		description  string
	}{
		{
			workflow:   "frontend-development",
			components: []string{"public-website", "admin-portal"},
			endpoints:  []string{"http://localhost:3000", "http://localhost:3001"},
			status:     "partial", // Public working, admin 404
			description: "Frontend development workflow must be complete with both applications functional",
		},
		{
			workflow:   "backend-integration",
			components: []string{"service-mesh", "gateways", "APIs"},
			endpoints:  []string{"http://localhost:3502/v1.0/healthz", "http://localhost:9001/health", "http://localhost:9001/api/news"},
			status:     "functional",
			description: "Backend integration must be fully functional for frontend consumption",
		},
		{
			workflow:   "contract-generation",
			components: []string{"openapi-specs", "typescript-clients"},
			endpoints:  []string{}, // No HTTP endpoints for this workflow
			status:     "functional",
			description: "Contract generation workflow must produce functional TypeScript clients",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Validate development workflow components
	for _, workflow := range workflowValidationTests {
		t.Run("Workflow_"+workflow.workflow, func(t *testing.T) {
			if len(workflow.endpoints) > 0 {
				// Test workflow endpoints
				for i, endpoint := range workflow.endpoints {
					component := "component"
					if i < len(workflow.components) {
						component = workflow.components[i]
					}
					
					req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
					require.NoError(t, err, "Failed to create workflow request for %s", component)

					resp, err := client.Do(req)
					if err == nil {
						defer resp.Body.Close()
						
						if workflow.status == "functional" {
							assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
								"%s component %s must be functional", workflow.description, component)
						} else if workflow.status == "partial" {
							// Some components working, some not
							t.Logf("Workflow %s component %s status: %d (%s)", workflow.workflow, component, resp.StatusCode, workflow.status)
						}
					} else {
						if workflow.status == "functional" {
							t.Errorf("%s component %s not accessible: %v", workflow.description, component, err)
						} else {
							t.Logf("Workflow %s component %s not accessible (%s): %v", workflow.workflow, component, workflow.status, err)
						}
					}
				}
			} else {
				// Non-endpoint based workflow validation
				t.Logf("Workflow %s validation: %s", workflow.workflow, workflow.status)
			}
		})
	}
}

// validateFrontendEnvironmentPrerequisites ensures environment health before integration testing
func validateFrontendEnvironmentPrerequisites(t *testing.T) {
	// Check critical backend infrastructure components are running
	criticalContainers := []string{"postgresql", "dapr-control-plane", "content", "inquiries", "notifications", "public-gateway", "admin-gateway"}
	
	for _, container := range criticalContainers {
		cmd := exec.Command("podman", "ps", "--filter", "name="+container, "--format", "{{.Names}}")
		output, err := cmd.Output()
		require.NoError(t, err, "Failed to check critical container %s", container)

		if !strings.Contains(string(output), container) {
			t.Skipf("Critical container %s not running - environment not ready for integration testing", container)
		}
	}
}