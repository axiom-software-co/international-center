package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// ProxyResponse wraps response data with status code information
type ProxyResponse struct {
	Data       interface{}
	StatusCode int
	Headers    map[string]string
}

// ServiceInvocationInterface defines the contract for service invocation operations
type ServiceInvocationInterface interface {
	InvokeContentAPI(ctx context.Context, method, httpVerb string, data []byte) (*dapr.ServiceResponse, error)
	InvokeInquiriesAPI(ctx context.Context, method, httpVerb string, data []byte) (*dapr.ServiceResponse, error)
	InvokeNotificationAPI(ctx context.Context, method, httpVerb string, data []byte) (*dapr.ServiceResponse, error)
	CheckContentAPIHealth(ctx context.Context) (bool, error)
	CheckInquiriesAPIHealth(ctx context.Context) (bool, error)
	CheckNotificationAPIHealth(ctx context.Context) (bool, error)
	GetContentAPIMetrics(ctx context.Context) (map[string]interface{}, error)
	GetInquiriesAPIMetrics(ctx context.Context) (map[string]interface{}, error)
	GetNotificationAPIMetrics(ctx context.Context) (map[string]interface{}, error)
}

// ServiceProxy handles proxying requests to backend services via Dapr service invocation
type ServiceProxy struct {
	serviceInvocation ServiceInvocationInterface
	configuration     *GatewayConfiguration
}

// NewServiceProxy creates a new service proxy
func NewServiceProxy(client *dapr.Client, config *GatewayConfiguration) *ServiceProxy {
	return &ServiceProxy{
		serviceInvocation: dapr.NewServiceInvocation(client),
		configuration:     config,
	}
}

// NewServiceProxyWithInvocation creates a service proxy with a specific service invocation implementation (for testing)
func NewServiceProxyWithInvocation(serviceInvocation ServiceInvocationInterface, config *GatewayConfiguration) *ServiceProxy {
	return &ServiceProxy{
		serviceInvocation: serviceInvocation,
		configuration:     config,
	}
}

// ProxyRequest proxies HTTP request to backend service via Dapr service invocation
func (p *ServiceProxy) ProxyRequest(ctx context.Context, w http.ResponseWriter, r *http.Request, targetService string) error {
	// Determine target service and method based on path
	serviceName, httpMethod, targetPath, err := p.parseTargetService(r.URL.Path, targetService, r.Method)
	if err != nil {
		return domain.NewValidationError(fmt.Sprintf("failed to parse target service: %v", err))
	}

	// Add correlation context to request
	correlationCtx := domain.FromContext(ctx)
	if correlationCtx.CorrelationID == "" {
		correlationCtx.CorrelationID = uuid.New().String()
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
	
	// Forward authentication headers for admin operations
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		headers["Authorization"] = authHeader
	}
	
	// Forward JWT-related headers
	if jwtHeader := r.Header.Get("X-JWT-Token"); jwtHeader != "" {
		headers["X-JWT-Token"] = jwtHeader
	}
	
	// Forward role information if present
	if rolesHeader := r.Header.Get("X-User-Roles"); rolesHeader != "" {
		headers["X-User-Roles"] = rolesHeader
	}

	// Add correlation ID
	headers["X-Correlation-ID"] = correlationCtx.CorrelationID
	
	// Add gateway context for service-to-service audit trail
	headers["X-Gateway-Type"] = string(p.configuration.Type)
	headers["X-Gateway-Version"] = p.configuration.Version

	// Create request context with timeout
	requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Invoke service based on target
	var response *ProxyResponse
	switch serviceName {
	case "content-api", "content":
		response, err = p.invokeContentAPI(requestCtx, httpMethod, targetPath, requestData, headers)
	case "inquiries-api", "inquiries":
		response, err = p.invokeInquiriesAPI(requestCtx, httpMethod, targetPath, requestData, headers)
	case "notification-api", "notifications":
		response, err = p.invokeNotificationAPI(requestCtx, httpMethod, targetPath, requestData, headers)
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
func (p *ServiceProxy) parseTargetService(path, targetService, httpMethod string) (string, string, string, error) {
	// Remove leading slash
	path = strings.TrimPrefix(path, "/")
	
	// Parse path components
	parts := strings.Split(path, "/")
	
	// Standardized routing: only support /api/v1/* pattern
	if len(parts) < 3 || parts[0] != "api" || parts[1] != "v1" {
		return "", "", "", fmt.Errorf("invalid API path format - only /api/v1/* paths supported")
	}

	// Extract service from standardized path: /api/v1/service
	service := parts[2] // e.g., "content", "services", "news", etc.
	
	// Determine service name based on domain consolidation
	var serviceName string
	switch service {
	case "content", "services", "news", "research", "events":
		// All content domains consolidated into content service
		serviceName = "content"
	case "inquiries":
		serviceName = "inquiries"
	case "notifications", "subscribers":
		// Notifications and subscribers handled by notifications service
		serviceName = "notifications"
	default:
		return "", "", "", fmt.Errorf("unknown service: %s", service)
	}

	// Use original path for backend service invocation
	targetPath := "/" + strings.Join(parts, "/")
	
	return serviceName, httpMethod, targetPath, nil
}

// invokeContentAPI invokes content API service (handles content, services, research, events, and news domains)
func (p *ServiceProxy) invokeContentAPI(ctx context.Context, method, path string, data interface{}, headers map[string]string) (*ProxyResponse, error) {
	// Standardized routing: only support /api/v1/* paths for content domains
	if !strings.HasPrefix(path, "/api/v1/") {
		return nil, domain.NewNotFoundError("content API endpoint", path)
	}

	// Convert data to []byte if needed
	var requestData []byte
	if data != nil {
		var err error
		requestData, err = json.Marshal(data)
		if err != nil {
			return nil, domain.NewValidationError("failed to marshal request data")
		}
	}
	
	response, err := p.serviceInvocation.InvokeContentAPI(ctx, path, method, requestData)
	if err != nil {
		return nil, err
	}
	
	// Parse response data back to interface{}
	var result interface{}
	if len(response.Data) > 0 {
		if err := json.Unmarshal(response.Data, &result); err != nil {
			return nil, domain.NewInternalError("failed to unmarshal response", err)
		}
	}
	
	return &ProxyResponse{
		Data:       result,
		StatusCode: response.StatusCode,
		Headers:    response.Headers,
	}, nil
}

// invokeInquiriesAPI invokes inquiries API service (handles business, donations, media, volunteers)
func (p *ServiceProxy) invokeInquiriesAPI(ctx context.Context, method, path string, data interface{}, headers map[string]string) (*ProxyResponse, error) {
	// Standardized routing: only support /api/v1/* paths for inquiries
	if !strings.HasPrefix(path, "/api/v1/inquiries") {
		return nil, domain.NewNotFoundError("inquiries API endpoint", path)
	}

	// Convert data to []byte if needed
	var requestData []byte
	if data != nil {
		var err error
		requestData, err = json.Marshal(data)
		if err != nil {
			return nil, domain.NewValidationError("failed to marshal request data")
		}
	}
	
	response, err := p.serviceInvocation.InvokeInquiriesAPI(ctx, path, method, requestData)
	if err != nil {
		return nil, err
	}
	
	// Parse response data back to interface{}
	var result interface{}
	if len(response.Data) > 0 {
		if err := json.Unmarshal(response.Data, &result); err != nil {
			return nil, domain.NewInternalError("failed to unmarshal response", err)
		}
	}
	
	return &ProxyResponse{
		Data:       result,
		StatusCode: response.StatusCode,
		Headers:    response.Headers,
	}, nil
}


// invokeNotificationAPI invokes notification API service
func (p *ServiceProxy) invokeNotificationAPI(ctx context.Context, method, path string, data interface{}, headers map[string]string) (*ProxyResponse, error) {
	// Standardized routing: only support /api/v1/* paths for notifications and subscribers
	if !strings.HasPrefix(path, "/api/v1/notifications") && !strings.HasPrefix(path, "/api/v1/subscribers") {
		return nil, domain.NewNotFoundError("notification API endpoint", path)
	}

	// Convert data to []byte if needed
	var requestData []byte
	if data != nil {
		// Check if data is already []byte (avoid double marshaling)
		if bytes, ok := data.([]byte); ok {
			requestData = bytes
		} else {
			var err error
			requestData, err = json.Marshal(data)
			if err != nil {
				return nil, domain.NewValidationError("failed to marshal request data")
			}
		}
	}
	
	response, err := p.serviceInvocation.InvokeNotificationAPI(ctx, path, method, requestData)
	if err != nil {
		return nil, err
	}
	
	// Parse response data back to interface{}
	var result interface{}
	if len(response.Data) > 0 {
		if err := json.Unmarshal(response.Data, &result); err != nil {
			return nil, domain.NewInternalError("failed to unmarshal response", err)
		}
	}
	
	return &ProxyResponse{
		Data:       result,
		StatusCode: response.StatusCode,
		Headers:    response.Headers,
	}, nil
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
func (p *ServiceProxy) writeProxyResponse(w http.ResponseWriter, response *ProxyResponse, correlationCtx *domain.CorrelationContext) error {
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
	
	w.WriteHeader(response.StatusCode)
	
	// Encode response data as JSON
	return json.NewEncoder(w).Encode(response.Data)
}

// HealthCheck performs health check for the service proxy
func (p *ServiceProxy) HealthCheck(ctx context.Context) error {
	// Check content API health
	contentHealthy, err := p.serviceInvocation.CheckContentAPIHealth(ctx)
	if err != nil || !contentHealthy {
		return fmt.Errorf("content API health check failed: %v", err)
	}
	
	// Check inquiries API health
	inquiriesHealthy, err := p.serviceInvocation.CheckInquiriesAPIHealth(ctx)
	if err != nil || !inquiriesHealthy {
		return fmt.Errorf("inquiries API health check failed: %v", err)
	}
	
	// Check notification API health
	notificationHealthy, err := p.serviceInvocation.CheckNotificationAPIHealth(ctx)
	if err != nil || !notificationHealthy {
		return fmt.Errorf("notification API health check failed: %v", err)
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
	
	// Get inquiries API metrics
	inquiriesMetrics, err := p.serviceInvocation.GetInquiriesAPIMetrics(ctx)
	if err == nil {
		metrics["inquiries_api"] = inquiriesMetrics
	}
	
	// Get notification API metrics
	notificationMetrics, err := p.serviceInvocation.GetNotificationAPIMetrics(ctx)
	if err == nil {
		metrics["notification_api"] = notificationMetrics
	}
	
	// Add gateway metrics
	metrics["gateway"] = map[string]interface{}{
		"uptime":        time.Now().UTC(),
		"version":       "1.0.0",
		"configuration": p.configuration.Name,
	}
	
	return metrics, nil
}