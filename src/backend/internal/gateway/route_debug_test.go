package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

// TestRouteRegistrationDebug helps debug what routes are actually being registered
func TestRouteRegistrationDebug(t *testing.T) {
	ctx := context.Background()
	
	t.Run("debug_public_gateway_routes", func(t *testing.T) {
		publicHandler, err := NewPublicGateway(ctx, "development")
		if err != nil {
			t.Fatalf("Failed to create public gateway: %v", err)
		}
		
		// Extract the underlying router from the DAPR middleware wrapper
		daprWrapper := publicHandler.(*DAPRMiddlewareSimulator)
		router := daprWrapper.GetHandler().(*mux.Router)
		routes := extractRoutes(router)
		
		t.Logf("Public Gateway Routes (%d total):", len(routes))
		for _, route := range routes {
			t.Logf("  - %s", route)
		}
		
		// Verify key public routes exist
		expectedPublicRoutes := []string{
			"/api/v1/services",
			"/api/v1/news",
			"/health",
		}
		
		for _, expected := range expectedPublicRoutes {
			found := false
			for _, route := range routes {
				if strings.Contains(route, expected) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected public route %s not found in registered routes", expected)
			}
		}
	})
	
	t.Run("debug_admin_gateway_routes", func(t *testing.T) {
		adminHandler, err := NewAdminGateway(ctx, "development")
		if err != nil {
			t.Fatalf("Failed to create admin gateway: %v", err)
		}
		
		// Extract the underlying router from the DAPR middleware wrapper
		daprWrapper := adminHandler.(*DAPRMiddlewareSimulator)
		router := daprWrapper.handler.(*mux.Router)
		routes := extractRoutes(router)
		
		t.Logf("Admin Gateway Routes (%d total):", len(routes))
		for _, route := range routes {
			t.Logf("  - %s", route)
		}
		
		// Verify key admin routes exist
		expectedAdminRoutes := []string{
			"/admin/api/v1/services",
			"/admin/api/v1/news",
			"/health",
		}
		
		for _, expected := range expectedAdminRoutes {
			found := false
			for _, route := range routes {
				if strings.Contains(route, expected) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected admin route %s not found in registered routes", expected)
			}
		}
	})
	
	t.Run("test_admin_route_matching", func(t *testing.T) {
		adminHandler, err := NewAdminGateway(ctx, "development")
		if err != nil {
			t.Fatalf("Failed to create admin gateway: %v", err)
		}
		
		// Test unauthenticated request (should get 401)
		req := httptest.NewRequest("GET", "/admin/api/v1/services", nil)
		recorder := httptest.NewRecorder()
		adminHandler.ServeHTTP(recorder, req)
		
		t.Logf("Unauthenticated request result: %d %s", recorder.Code, http.StatusText(recorder.Code))
		t.Logf("Response body: %s", recorder.Body.String())
		
		// Test authenticated request with valid admin token
		authReq := httptest.NewRequest("GET", "/admin/api/v1/services", nil)
		authReq.Header.Set("Authorization", "Bearer valid_admin_token")
		authRecorder := httptest.NewRecorder()
		adminHandler.ServeHTTP(authRecorder, authReq)
		
		t.Logf("Authenticated admin request result: %d %s", authRecorder.Code, http.StatusText(authRecorder.Code))
		t.Logf("Response headers: %v", authRecorder.Header())
		t.Logf("Response body: %s", authRecorder.Body.String())
		
		// Test authenticated POST request (should get 201)
		postReq := httptest.NewRequest("POST", "/admin/api/v1/services", strings.NewReader(`{"name":"test"}`))
		postReq.Header.Set("Authorization", "Bearer valid_admin_token")
		postReq.Header.Set("Content-Type", "application/json")
		postRecorder := httptest.NewRecorder()
		adminHandler.ServeHTTP(postRecorder, postReq)
		
		t.Logf("Authenticated POST request result: %d %s", postRecorder.Code, http.StatusText(postRecorder.Code))
		t.Logf("Response body: %s", postRecorder.Body.String())
	})
}

// extractRoutes extracts route patterns from a mux router for debugging
func extractRoutes(router *mux.Router) []string {
	var routes []string
	
	err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err != nil {
			return nil // Skip routes without path templates
		}
		
		methods, err := route.GetMethods()
		if err != nil {
			methods = []string{"*"} // Default if no specific methods
		}
		
		for _, method := range methods {
			routes = append(routes, method+" "+pathTemplate)
		}
		
		return nil
	})
	
	if err != nil {
		routes = append(routes, "Error walking routes: "+err.Error())
	}
	
	return routes
}