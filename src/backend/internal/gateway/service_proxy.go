package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/internal/shared/domain"
)

// ServiceProxy handles proxying requests to backend services via Dapr service invocation
type ServiceProxy struct {
	serviceInvocation *dapr.ServiceInvocation
	configuration     *GatewayConfiguration
}

// NewServiceProxy creates a new service proxy
func NewServiceProxy(client *dapr.Client, config *GatewayConfiguration) *ServiceProxy {
	return &ServiceProxy{
		serviceInvocation: dapr.NewServiceInvocation(client),
		configuration:     config,
	}
}

// ProxyRequest proxies HTTP request to backend service via Dapr service invocation
func (p *ServiceProxy) ProxyRequest(ctx context.Context, w http.ResponseWriter, r *http.Request, targetService string) error {
	// Determine target service and method based on path
	serviceName, httpMethod, targetPath, err := p.parseTargetService(r.URL.Path, targetService)
	if err != nil {
		return domain.NewValidationError(fmt.Sprintf("failed to parse target service: %v", err))
	}

	// Add correlation context to request
	correlationCtx := domain.FromContext(ctx)
	if correlationCtx.CorrelationID == "" {
		correlationCtx.CorrelationID = domain.GenerateCorrelationID()
	}

	// Prepare request data
	var requestData interface{}
	if httpMethod != "GET" && r.ContentLength > 0 {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			return domain.NewValidationError("invalid request body")
		}
	}

	// Extract headers to forward
	headers := p.extractForwardableHeaders(r)
	
	// Add authentication context if present
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		headers["X-User-ID"] = userID
	}

	// Add correlation ID
	headers["X-Correlation-ID"] = correlationCtx.CorrelationID

	// Create request context with timeout
	requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Invoke service based on target
	var response interface{}
	switch serviceName {
	case "content-api":
		response, err = p.invokeContentAPI(requestCtx, httpMethod, targetPath, requestData, headers)
	case "services-api":
		response, err = p.invokeServicesAPI(requestCtx, httpMethod, targetPath, requestData, headers)
	default:
		return domain.NewValidationError(fmt.Sprintf("unknown target service: %s", serviceName))
	}

	if err != nil {
		return err
	}

	// Write response
	return p.writeProxyResponse(w, response, correlationCtx)
}

// parseTargetService parses the request path to determine target service
func (p *ServiceProxy) parseTargetService(path, targetService string) (string, string, string, error) {
	// Remove leading slash
	path = strings.TrimPrefix(path, "/")
	
	// Parse path components
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[0] != "api" {
		return "", "", "", fmt.Errorf("invalid API path format")
	}

	// Extract version and service from path
	version := parts[1] // e.g., "v1"
	service := parts[2] // e.g., "content" or "services"
	
	// Determine service name and remaining path
	var serviceName string
	switch service {
	case "content":
		serviceName = "content-api"
	case "services":
		serviceName = "services-api"
	default:
		return "", "", "", fmt.Errorf("unknown service: %s", service)
	}

	// Reconstruct target path
	targetPath := "/" + strings.Join(parts, "/")
	
	return serviceName, "GET", targetPath, nil // Currently only GET endpoints
}

// invokeContentAPI invokes content API service
func (p *ServiceProxy) invokeContentAPI(ctx context.Context, method, path string, data interface{}, headers map[string]string) (interface{}, error) {
	switch {
	case strings.HasPrefix(path, "/api/v1/content"):
		return p.serviceInvocation.InvokeContentAPI(ctx, method, path, data, headers)
	default:
		return nil, domain.NewNotFoundError("content API endpoint", path)
	}
}

// invokeServicesAPI invokes services API service
func (p *ServiceProxy) invokeServicesAPI(ctx context.Context, method, path string, data interface{}, headers map[string]string) (interface{}, error) {
	switch {
	case strings.HasPrefix(path, "/api/v1/services"):
		return p.serviceInvocation.InvokeServicesAPI(ctx, method, path, data, headers)
	default:
		return nil, domain.NewNotFoundError("services API endpoint", path)
	}
}

// extractForwardableHeaders extracts headers that should be forwarded to backend services
func (p *ServiceProxy) extractForwardableHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)
	
	// Forward specific headers
	forwardableHeaders := []string{
		"Authorization",
		"X-User-ID",
		"X-Request-ID",
		"X-Forwarded-For",
		"User-Agent",
		"Accept",
		"Accept-Language",
		"Content-Type",
	}
	
	for _, headerName := range forwardableHeaders {
		if value := r.Header.Get(headerName); value != "" {
			headers[headerName] = value
		}
	}
	
	return headers
}

// writeProxyResponse writes the proxied response back to the client
func (p *ServiceProxy) writeProxyResponse(w http.ResponseWriter, response interface{}, correlationCtx *domain.CorrelationContext) error {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationCtx.CorrelationID)
	
	// Add security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	
	// Set cache control based on gateway configuration
	if p.configuration.CacheControl.Enabled {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", p.configuration.CacheControl.MaxAge))
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	
	w.WriteHeader(http.StatusOK)
	
	// Encode response as JSON
	return json.NewEncoder(w).Encode(response)
}

// HealthCheck performs health check for the service proxy
func (p *ServiceProxy) HealthCheck(ctx context.Context) error {
	// Check content API health
	contentHealthy, err := p.serviceInvocation.CheckContentAPIHealth(ctx)
	if err != nil || !contentHealthy {
		return fmt.Errorf("content API health check failed: %v", err)
	}
	
	// Check services API health
	servicesHealthy, err := p.serviceInvocation.CheckServicesAPIHealth(ctx)
	if err != nil || !servicesHealthy {
		return fmt.Errorf("services API health check failed: %v", err)
	}
	
	return nil
}

// GetServiceMetrics returns metrics for proxied services
func (p *ServiceProxy) GetServiceMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
	
	// Get content API metrics
	contentMetrics, err := p.serviceInvocation.GetContentAPIMetrics(ctx)
	if err == nil {
		metrics["content_api"] = contentMetrics
	}
	
	// Get services API metrics
	servicesMetrics, err := p.serviceInvocation.GetServicesAPIMetrics(ctx)
	if err == nil {
		metrics["services_api"] = servicesMetrics
	}
	
	// Add gateway metrics
	metrics["gateway"] = map[string]interface{}{
		"uptime":        time.Now().UTC(),
		"version":       "1.0.0",
		"configuration": p.configuration.Name,
	}
	
	return metrics, nil
}