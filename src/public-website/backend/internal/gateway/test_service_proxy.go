package gateway

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
)

// TestServiceInvocation implements ServiceInvocationInterface for testing
type TestServiceInvocation struct{}

// InvokeContentAPI returns test response for content API
func (t *TestServiceInvocation) InvokeContentAPI(ctx context.Context, method, httpVerb string, data []byte) (*dapr.ServiceResponse, error) {
	response := map[string]interface{}{
		"data":    []interface{}{},
		"status":  "success",
		"message": "Test content API response",
	}
	
	// Determine status code based on HTTP method and route
	var statusCode int
	switch httpVerb {
	case "POST":
		statusCode = http.StatusCreated // 201 for creating resources (news, services, etc.)
	case "PUT", "DELETE":
		statusCode = http.StatusOK // 200 for updates and deletes
	default: // GET and others
		statusCode = http.StatusOK // 200 for GET requests
	}
	
	responseData, _ := json.Marshal(response)
	return &dapr.ServiceResponse{
		Data:       responseData,
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

// InvokeInquiriesAPI returns test response for inquiries API
func (t *TestServiceInvocation) InvokeInquiriesAPI(ctx context.Context, method, httpVerb string, data []byte) (*dapr.ServiceResponse, error) {
	response := map[string]interface{}{
		"data":    []interface{}{},
		"status":  "success", 
		"message": "Test inquiries API response",
	}
	
	responseData, _ := json.Marshal(response)
	return &dapr.ServiceResponse{
		Data:       responseData,
		StatusCode: http.StatusAccepted, // Changed to 202 as expected by tests
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

// InvokeServicesAPI returns test response for services API
func (t *TestServiceInvocation) InvokeServicesAPI(ctx context.Context, method, httpVerb string, data []byte) (*dapr.ServiceResponse, error) {
	response := map[string]interface{}{
		"data":    []interface{}{},
		"status":  "success",
		"message": "Test services API response", 
	}
	
	responseData, _ := json.Marshal(response)
	return &dapr.ServiceResponse{
		Data:       responseData,
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

// InvokeNewsAPI returns test response for news API
func (t *TestServiceInvocation) InvokeNewsAPI(ctx context.Context, method, httpVerb string, data []byte) (*dapr.ServiceResponse, error) {
	var statusCode int
	switch httpVerb {
	case "POST":
		statusCode = http.StatusCreated // 201 for creating news
	case "PUT", "DELETE":
		statusCode = http.StatusOK
	default:
		statusCode = http.StatusOK
	}

	response := map[string]interface{}{
		"data":    []interface{}{},
		"status":  "success",
		"message": "Test news API response", 
	}
	
	responseData, _ := json.Marshal(response)
	return &dapr.ServiceResponse{
		Data:       responseData,
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

// InvokeNotificationAPI returns test response for notification API
func (t *TestServiceInvocation) InvokeNotificationAPI(ctx context.Context, method, httpVerb string, data []byte) (*dapr.ServiceResponse, error) {
	response := map[string]interface{}{
		"data":    []interface{}{},
		"status":  "success",
		"message": "Test notification API response", 
	}
	
	responseData, _ := json.Marshal(response)
	return &dapr.ServiceResponse{
		Data:       responseData,
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

// CheckContentAPIHealth returns healthy status for content API
func (t *TestServiceInvocation) CheckContentAPIHealth(ctx context.Context) (bool, error) {
	return true, nil
}

// CheckInquiriesAPIHealth returns healthy status for inquiries API  
func (t *TestServiceInvocation) CheckInquiriesAPIHealth(ctx context.Context) (bool, error) {
	return true, nil
}

// CheckServicesAPIHealth returns healthy status for services API  
func (t *TestServiceInvocation) CheckServicesAPIHealth(ctx context.Context) (bool, error) {
	return true, nil
}

// CheckNewsAPIHealth returns healthy status for news API  
func (t *TestServiceInvocation) CheckNewsAPIHealth(ctx context.Context) (bool, error) {
	return true, nil
}

// CheckNotificationAPIHealth returns healthy status for notification API
func (t *TestServiceInvocation) CheckNotificationAPIHealth(ctx context.Context) (bool, error) {
	return true, nil
}

// GetContentAPIMetrics returns test metrics for content API
func (t *TestServiceInvocation) GetContentAPIMetrics(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"requests_total": 100,
		"errors_total":   0,
		"latency_ms":     50,
	}, nil
}

// GetInquiriesAPIMetrics returns test metrics for inquiries API
func (t *TestServiceInvocation) GetInquiriesAPIMetrics(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"requests_total": 75,
		"errors_total":   0, 
		"latency_ms":     30,
	}, nil
}

// GetServicesAPIMetrics returns test metrics for services API
func (t *TestServiceInvocation) GetServicesAPIMetrics(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"requests_total": 150,
		"errors_total":   0, 
		"latency_ms":     40,
	}, nil
}

// GetNewsAPIMetrics returns test metrics for news API
func (t *TestServiceInvocation) GetNewsAPIMetrics(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"requests_total": 200,
		"errors_total":   0, 
		"latency_ms":     35,
	}, nil
}

// GetNotificationAPIMetrics returns test metrics for notification API
func (t *TestServiceInvocation) GetNotificationAPIMetrics(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"requests_total": 25,
		"errors_total":   0,
		"latency_ms":     20,
	}, nil
}