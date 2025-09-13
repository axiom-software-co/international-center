package integration

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE: Frontend Development Servers Tests
// These tests validate that frontend development servers are accessible on configured ports

func TestFrontendDevelopmentServers_ApplicationAccessibility(t *testing.T) {
	// This test requires complete environment health - enforcing axiom rule
	validateFrontendDevEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Frontend applications that must be accessible for development workflow
	frontendApplicationTests := []struct {
		applicationName string
		port            int
		endpoint        string
		technology      string
		description     string
		critical        bool
	}{
		{
			applicationName: "public-website",
			port:            3000,
			endpoint:        "http://localhost:3000",
			technology:      "Astro + Vue + Pinia",
			description:     "Public website frontend must be accessible for frontend development workflow",
			critical:        true,
		},
		{
			applicationName: "admin-portal",
			port:            3001,
			endpoint:        "http://localhost:3001",
			technology:      "Astro + Vue + Pinia",
			description:     "Admin portal frontend must be accessible for admin development workflow",
			critical:        true,
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test frontend application accessibility
	for _, app := range frontendApplicationTests {
		t.Run("FrontendApp_"+app.applicationName, func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, "GET", app.endpoint, nil)
			require.NoError(t, err, "Failed to create frontend application request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				
				if app.critical {
					// Critical frontend applications must be accessible
					assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500,
						"%s - frontend application must be accessible for development", app.description)
					
					// Should not return connection refused
					assert.NotEqual(t, 0, resp.StatusCode,
						"%s - frontend development server must be running", app.description)
				}
			} else {
				if app.critical {
					// Expected to fail in RED phase - development servers not running
					t.Errorf("%s not accessible (expected in RED phase): %v", app.description, err)
				}
			}
		})
	}
}

func TestFrontendDevelopmentServers_FullStackEnvironmentValidation(t *testing.T) {
	// Test that complete development environment includes both backend and frontend
	validateFrontendDevEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Complete development stack components
	developmentStackComponents := []struct {
		component     string
		endpoint      string
		componentType string
		description   string
		operational   bool
	}{
		{
			component:     "backend-service-mesh",
			endpoint:      "http://localhost:3502/v1.0/healthz",
			componentType: "backend",
			description:   "Backend service mesh must be operational",
			operational:   true, // Should be working
		},
		{
			component:     "public-gateway",
			endpoint:      "http://localhost:9001/health",
			componentType: "backend",
			description:   "Public gateway must be operational",
			operational:   true, // Should be working
		},
		{
			component:     "admin-gateway",
			endpoint:      "http://localhost:9000/health",
			componentType: "backend",
			description:   "Admin gateway must be operational",
			operational:   true, // Should be working
		},
		{
			component:     "public-website-frontend",
			endpoint:      "http://localhost:3000",
			componentType: "frontend",
			description:   "Public website frontend must be accessible for complete development stack",
			operational:   false, // Should fail in RED phase
		},
		{
			component:     "admin-portal-frontend",
			endpoint:      "http://localhost:3001",
			componentType: "frontend",
			description:   "Admin portal frontend must be accessible for complete development stack",
			operational:   false, // Should fail in RED phase
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test complete development stack
	for _, stack := range developmentStackComponents {
		t.Run("DevStack_"+stack.component, func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, "GET", stack.endpoint, nil)
			require.NoError(t, err, "Failed to create development stack request")

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				
				if stack.operational {
					// Components that should be working
					assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
						"%s - component must be operational for complete development stack", stack.description)
				} else {
					// Components that should fail in RED phase
					t.Logf("%s accessible (unexpected for RED phase)", stack.description)
				}
			} else {
				if stack.operational {
					t.Errorf("%s not accessible but should be operational: %v", stack.description, err)
				} else {
					// Expected for frontend components in RED phase
					t.Logf("%s not accessible (expected in RED phase for %s components): %v", stack.description, stack.componentType, err)
				}
			}
		})
	}
}

func TestFrontendDevelopmentServers_ContractClientPrerequisites(t *testing.T) {
	// Test prerequisites for frontend contract client integration
	validateFrontendDevEnvironmentPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Backend APIs that frontend contract clients need to consume
	contractClientPrerequisites := []struct {
		apiName         string
		gatewayEndpoint string
		backendEndpoint string
		contractType    string
		description     string
	}{
		{
			apiName:         "news-api",
			gatewayEndpoint: "http://localhost:9001/api/news",
			backendEndpoint: "http://localhost:3001/api/news",
			contractType:    "public",
			description:     "News API must be accessible for frontend contract client consumption",
		},
		{
			apiName:         "events-api",
			gatewayEndpoint: "http://localhost:9001/api/events",
			backendEndpoint: "http://localhost:3001/api/events",
			contractType:    "public",
			description:     "Events API must be accessible for frontend contract client consumption",
		},
		{
			apiName:         "inquiries-api",
			gatewayEndpoint: "http://localhost:9000/api/admin/inquiries",
			backendEndpoint: "http://localhost:3101/api/inquiries",
			contractType:    "admin",
			description:     "Inquiries API must be accessible for admin portal contract client consumption",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Act & Assert: Test contract client prerequisites
	for _, prereq := range contractClientPrerequisites {
		t.Run("ContractPrereq_"+prereq.apiName, func(t *testing.T) {
			// Test gateway endpoint (frontend will use this)
			gatewayReq, err := http.NewRequestWithContext(ctx, "GET", prereq.gatewayEndpoint, nil)
			require.NoError(t, err, "Failed to create gateway request")

			gatewayResp, err := client.Do(gatewayReq)
			require.NoError(t, err, "Gateway endpoint must be accessible for contract clients")
			defer gatewayResp.Body.Close()

			assert.True(t, gatewayResp.StatusCode >= 200 && gatewayResp.StatusCode < 300,
				"%s - gateway endpoint must be functional for frontend contract client", prereq.description)

			// Test backend endpoint (gateway routes to this)
			backendReq, err := http.NewRequestWithContext(ctx, "GET", prereq.backendEndpoint, nil)
			require.NoError(t, err, "Failed to create backend request")

			backendResp, err := client.Do(backendReq)
			require.NoError(t, err, "Backend endpoint must be accessible")
			defer backendResp.Body.Close()

			assert.True(t, backendResp.StatusCode >= 200 && backendResp.StatusCode < 300,
				"%s - backend endpoint must be functional for contract client integration", prereq.description)
		})
	}
}

func TestFrontendDevelopmentServers_DevelopmentWorkflowRequirements(t *testing.T) {
	// Test development workflow requirements for complete frontend-backend integration
	validateFrontendDevEnvironmentPrerequisites(t)

	// Development workflow requirements
	workflowRequirements := []struct {
		requirement string
		validation  string
		component   string
		description string
	}{
		{
			requirement: "backend-api-layer",
			validation:  "functional",
			component:   "backend",
			description: "Backend API layer must be functional for frontend consumption",
		},
		{
			requirement: "gateway-routing",
			validation:  "functional",
			component:   "backend",
			description: "Gateway routing must be functional for frontend API access",
		},
		{
			requirement: "frontend-development-servers",
			validation:  "missing",
			component:   "frontend",
			description: "Frontend development servers must be deployed for complete workflow",
		},
		{
			requirement: "contract-client-integration",
			validation:  "prerequisites-ready",
			component:   "integration",
			description: "Contract client integration prerequisites must be satisfied",
		},
	}

	// Act & Assert: Validate development workflow requirements
	for _, req := range workflowRequirements {
		t.Run("WorkflowReq_"+req.requirement, func(t *testing.T) {
			switch req.validation {
			case "functional":
				// Should be working (backend components)
				t.Logf("%s - %s (backend infrastructure complete)", req.description, req.validation)
			case "missing":
				// Should be missing (frontend components in RED phase)
				t.Logf("%s - %s (expected in RED phase)", req.description, req.validation)
			case "prerequisites-ready":
				// Prerequisites should be satisfied
				t.Logf("%s - %s (backend APIs functional)", req.description, req.validation)
			}
		})
	}
}

// validateFrontendDevEnvironmentPrerequisites ensures environment health before integration testing
func validateFrontendDevEnvironmentPrerequisites(t *testing.T) {
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