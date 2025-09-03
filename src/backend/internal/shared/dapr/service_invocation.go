package dapr

import (
	"context"
	"fmt"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/dapr/go-sdk/client"
)

// ServiceInvocation wraps Dapr service invocation operations
type ServiceInvocation struct {
	client *Client
}

// ServiceRequest represents a request to another service
type ServiceRequest struct {
	AppID       string
	MethodName  string
	HTTPVerb    string
	Data        []byte
	ContentType string
	Metadata    map[string]string
}

// ServiceResponse represents a response from another service
type ServiceResponse struct {
	Data        []byte
	ContentType string
	StatusCode  int
	Headers     map[string]string
}

// ServiceEndpoints contains the configured service endpoints
type ServiceEndpoints struct {
	ContentAPI  string
	ServicesAPI string
	AdminGW     string
	PublicGW    string
}

// NewServiceInvocation creates a new service invocation instance
func NewServiceInvocation(client *Client) *ServiceInvocation {
	return &ServiceInvocation{
		client: client,
	}
}

// GetServiceEndpoints returns the configured service endpoints
func (s *ServiceInvocation) GetServiceEndpoints() *ServiceEndpoints {
	return &ServiceEndpoints{
		ContentAPI:  getEnv("CONTENT_API_APP_ID", "content-api"),
		ServicesAPI: getEnv("SERVICES_API_APP_ID", "services-api"),
		AdminGW:     getEnv("ADMIN_GATEWAY_APP_ID", "admin-gateway"),
		PublicGW:    getEnv("PUBLIC_GATEWAY_APP_ID", "public-gateway"),
	}
}

// InvokeService invokes a method on another service
func (s *ServiceInvocation) InvokeService(ctx context.Context, req *ServiceRequest) (*ServiceResponse, error) {
	// Create timeout context if not already set
	if deadline, hasDeadline := ctx.Deadline(); !hasDeadline || time.Until(deadline) > 30*time.Second {
		timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		ctx = timeoutCtx
	}

	invokeReq := &client.InvokeMethodRequest{
		AppID:       req.AppID,
		MethodName:  req.MethodName,
		HTTPVerb:    req.HTTPVerb,
		Data:        req.Data,
		ContentType: req.ContentType,
		Metadata:    req.Metadata,
	}

	resp, err := s.client.GetClient().InvokeMethodWithContent(ctx, invokeReq)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, domain.NewTimeoutError(fmt.Sprintf("service invocation %s/%s", req.AppID, req.MethodName))
		}
		return nil, domain.NewDependencyError("service invocation", domain.WrapError(err, fmt.Sprintf("failed to invoke %s/%s", req.AppID, req.MethodName)))
	}

	return &ServiceResponse{
		Data:        resp,
		ContentType: req.ContentType,
		StatusCode:  200, // Dapr abstracts status codes
	}, nil
}

// InvokeContentAPI invokes a method on the content API service
func (s *ServiceInvocation) InvokeContentAPI(ctx context.Context, method, httpVerb string, data []byte) (*ServiceResponse, error) {
	endpoints := s.GetServiceEndpoints()
	
	req := &ServiceRequest{
		AppID:       endpoints.ContentAPI,
		MethodName:  method,
		HTTPVerb:    httpVerb,
		Data:        data,
		ContentType: "application/json",
	}

	return s.InvokeService(ctx, req)
}

// InvokeServicesAPI invokes a method on the services API service
func (s *ServiceInvocation) InvokeServicesAPI(ctx context.Context, method, httpVerb string, data []byte) (*ServiceResponse, error) {
	endpoints := s.GetServiceEndpoints()
	
	req := &ServiceRequest{
		AppID:       endpoints.ServicesAPI,
		MethodName:  method,
		HTTPVerb:    httpVerb,
		Data:        data,
		ContentType: "application/json",
	}

	return s.InvokeService(ctx, req)
}

// GetContent retrieves content by ID from content API
func (s *ServiceInvocation) GetContent(ctx context.Context, contentID string) (*ServiceResponse, error) {
	method := fmt.Sprintf("api/v1/content/%s", contentID)
	return s.InvokeContentAPI(ctx, method, "GET", nil)
}

// GetAllContent retrieves all content from content API
func (s *ServiceInvocation) GetAllContent(ctx context.Context) (*ServiceResponse, error) {
	return s.InvokeContentAPI(ctx, "api/v1/content", "GET", nil)
}

// GetContentDownload retrieves content download URL from content API
func (s *ServiceInvocation) GetContentDownload(ctx context.Context, contentID string) (*ServiceResponse, error) {
	method := fmt.Sprintf("api/v1/content/%s/download", contentID)
	return s.InvokeContentAPI(ctx, method, "GET", nil)
}

// GetContentPreview retrieves content preview from content API
func (s *ServiceInvocation) GetContentPreview(ctx context.Context, contentID string) (*ServiceResponse, error) {
	method := fmt.Sprintf("api/v1/content/%s/preview", contentID)
	return s.InvokeContentAPI(ctx, method, "GET", nil)
}

// GetService retrieves service by ID from services API
func (s *ServiceInvocation) GetService(ctx context.Context, serviceID string) (*ServiceResponse, error) {
	method := fmt.Sprintf("api/v1/services/%s", serviceID)
	return s.InvokeServicesAPI(ctx, method, "GET", nil)
}

// GetAllServices retrieves all services from services API
func (s *ServiceInvocation) GetAllServices(ctx context.Context) (*ServiceResponse, error) {
	return s.InvokeServicesAPI(ctx, "api/v1/services", "GET", nil)
}

// GetServiceBySlug retrieves service by slug from services API
func (s *ServiceInvocation) GetServiceBySlug(ctx context.Context, slug string) (*ServiceResponse, error) {
	method := fmt.Sprintf("api/v1/services/slug/%s", slug)
	return s.InvokeServicesAPI(ctx, method, "GET", nil)
}

// GetFeaturedServices retrieves featured services from services API
func (s *ServiceInvocation) GetFeaturedServices(ctx context.Context) (*ServiceResponse, error) {
	return s.InvokeServicesAPI(ctx, "api/v1/services/featured", "GET", nil)
}

// GetServiceCategories retrieves service categories from services API
func (s *ServiceInvocation) GetServiceCategories(ctx context.Context) (*ServiceResponse, error) {
	return s.InvokeServicesAPI(ctx, "api/v1/services/categories", "GET", nil)
}

// GetServicesByCategory retrieves services by category from services API
func (s *ServiceInvocation) GetServicesByCategory(ctx context.Context, categoryID string) (*ServiceResponse, error) {
	method := fmt.Sprintf("api/v1/services/categories/%s/services", categoryID)
	return s.InvokeServicesAPI(ctx, method, "GET", nil)
}

// CheckServiceHealth checks if a service is healthy
func (s *ServiceInvocation) CheckServiceHealth(ctx context.Context, appID string) error {
	req := &ServiceRequest{
		AppID:      appID,
		MethodName: "health",
		HTTPVerb:   "GET",
	}

	_, err := s.InvokeService(ctx, req)
	if err != nil {
		return fmt.Errorf("service %s is not healthy: %w", appID, err)
	}

	return nil
}

// CheckServiceReadiness checks if a service is ready
func (s *ServiceInvocation) CheckServiceReadiness(ctx context.Context, appID string) error {
	req := &ServiceRequest{
		AppID:      appID,
		MethodName: "health/ready",
		HTTPVerb:   "GET",
	}

	_, err := s.InvokeService(ctx, req)
	if err != nil {
		return fmt.Errorf("service %s is not ready: %w", appID, err)
	}

	return nil
}

// CheckAllServicesHealth checks the health of all core services
func (s *ServiceInvocation) CheckAllServicesHealth(ctx context.Context) error {
	endpoints := s.GetServiceEndpoints()
	services := []string{endpoints.ContentAPI, endpoints.ServicesAPI}

	for _, service := range services {
		if err := s.CheckServiceHealth(ctx, service); err != nil {
			return fmt.Errorf("health check failed for service %s: %w", service, err)
		}
	}

	return nil
}

// CheckAllServicesReadiness checks the readiness of all core services
func (s *ServiceInvocation) CheckAllServicesReadiness(ctx context.Context) error {
	endpoints := s.GetServiceEndpoints()
	services := []string{endpoints.ContentAPI, endpoints.ServicesAPI}

	for _, service := range services {
		if err := s.CheckServiceReadiness(ctx, service); err != nil {
			return fmt.Errorf("readiness check failed for service %s: %w", service, err)
		}
	}

	return nil
}

// InvokeWithRetry invokes a service with retry logic
func (s *ServiceInvocation) InvokeWithRetry(ctx context.Context, req *ServiceRequest, maxRetries int) (*ServiceResponse, error) {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := s.InvokeService(ctx, req)
		if err == nil {
			return resp, nil
		}
		
		lastErr = err
		
		if attempt < maxRetries {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt+1) * time.Second):
				// Continue with next attempt
			}
		}
	}

	return nil, fmt.Errorf("service invocation failed after %d attempts: %w", maxRetries+1, lastErr)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}