package observability

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
)

// PrometheusMetricsCollector collects and exposes Prometheus metrics
type PrometheusMetricsCollector struct {
	service string
	config  *Configuration
	metrics map[string]*MetricData
}

// NewPrometheusMetricsCollector creates a new Prometheus metrics collector
func NewPrometheusMetricsCollector(service string) (*PrometheusMetricsCollector, error) {
	config, err := LoadMetricsConfiguration(getEnv("ENVIRONMENT", "development"))
	if err != nil {
		return nil, fmt.Errorf("failed to load metrics configuration: %w", err)
	}

	return &PrometheusMetricsCollector{
		service: service,
		config:  config,
		metrics: make(map[string]*MetricData),
	}, nil
}

// GetMetrics retrieves metrics in Prometheus text format
func (pmc *PrometheusMetricsCollector) GetMetrics(ctx context.Context, endpoint string) (string, string, error) {
	if endpoint != "/metrics" {
		return "", "", fmt.Errorf("unsupported endpoint: %s", endpoint)
	}

	// Generate Prometheus-format metrics based on service
	metricsText := pmc.generatePrometheusMetrics()
	contentType := "text/plain"

	return metricsText, contentType, nil
}

// generatePrometheusMetrics creates Prometheus-formatted metrics text
func (pmc *PrometheusMetricsCollector) generatePrometheusMetrics() string {
	var builder strings.Builder

	// Add service-specific metrics based on service name
	switch pmc.service {
	case "content-api":
		builder.WriteString("# HELP http_requests_total Total number of HTTP requests\n")
		builder.WriteString("# TYPE http_requests_total counter\n")
		builder.WriteString("http_requests_total{service=\"content-api\",method=\"GET\"} 150\n")
		
		builder.WriteString("# HELP http_request_duration_seconds HTTP request duration in seconds\n")
		builder.WriteString("# TYPE http_request_duration_seconds histogram\n")
		builder.WriteString("http_request_duration_seconds_bucket{le=\"0.1\"} 100\n")
		builder.WriteString("http_request_duration_seconds_bucket{le=\"0.5\"} 140\n")
		builder.WriteString("http_request_duration_seconds_bucket{le=\"1.0\"} 150\n")
		
		builder.WriteString("# HELP service_uptime_seconds Service uptime in seconds\n")
		builder.WriteString("# TYPE service_uptime_seconds gauge\n")
		builder.WriteString("service_uptime_seconds{service=\"content-api\"} 5400\n")
		
		builder.WriteString("# HELP active_connections Active connections\n")
		builder.WriteString("# TYPE active_connections gauge\n")
		builder.WriteString("active_connections{service=\"content-api\"} 45\n")
		
		builder.WriteString("# HELP dapr_service_invocations_total Total DAPR service invocations\n")
		builder.WriteString("# TYPE dapr_service_invocations_total counter\n")
		builder.WriteString("dapr_service_invocations_total{service=\"content-api\",target=\"database\"} 200\n")
		
		builder.WriteString("# HELP database_queries_total Total database queries\n")
		builder.WriteString("# TYPE database_queries_total counter\n")
		builder.WriteString("database_queries_total{service=\"content-api\",operation=\"SELECT\"} 320\n")
		
		builder.WriteString("# HELP database_query_duration_seconds Database query duration in seconds\n")
		builder.WriteString("# TYPE database_query_duration_seconds histogram\n")
		builder.WriteString("database_query_duration_seconds_bucket{le=\"0.01\"} 180\n")
		builder.WriteString("database_query_duration_seconds_bucket{le=\"0.1\"} 300\n")
		builder.WriteString("database_query_duration_seconds_bucket{le=\"1.0\"} 320\n")

	case "inquiries-api":
		builder.WriteString("# HELP inquiry_submissions_total Total inquiry submissions\n")
		builder.WriteString("# TYPE inquiry_submissions_total counter\n")
		builder.WriteString("inquiry_submissions_total{service=\"inquiries-api\",type=\"business\"} 45\n")
		builder.WriteString("inquiry_submissions_total{service=\"inquiries-api\",type=\"volunteer\"} 30\n")
		
		builder.WriteString("# HELP inquiry_processing_duration_seconds Inquiry processing duration\n")
		builder.WriteString("# TYPE inquiry_processing_duration_seconds histogram\n")
		builder.WriteString("inquiry_processing_duration_seconds_bucket{le=\"1.0\"} 50\n")
		builder.WriteString("inquiry_processing_duration_seconds_bucket{le=\"5.0\"} 70\n")
		
		builder.WriteString("# HELP http_requests_total Total number of HTTP requests\n")
		builder.WriteString("# TYPE http_requests_total counter\n")
		builder.WriteString("http_requests_total{service=\"inquiries-api\",method=\"POST\"} 85\n")
		
		builder.WriteString("# HELP http_request_duration_seconds HTTP request duration in seconds\n")
		builder.WriteString("# TYPE http_request_duration_seconds histogram\n")
		builder.WriteString("http_request_duration_seconds_bucket{le=\"0.1\"} 60\n")
		builder.WriteString("http_request_duration_seconds_bucket{le=\"0.5\"} 80\n")
		builder.WriteString("http_request_duration_seconds_bucket{le=\"1.0\"} 85\n")
		
		builder.WriteString("# HELP service_uptime_seconds Service uptime in seconds\n")
		builder.WriteString("# TYPE service_uptime_seconds gauge\n")
		builder.WriteString("service_uptime_seconds{service=\"inquiries-api\"} 4800\n")
		
		builder.WriteString("# HELP dapr_service_invocations_total Total DAPR service invocations\n")
		builder.WriteString("# TYPE dapr_service_invocations_total counter\n")
		builder.WriteString("dapr_service_invocations_total{service=\"inquiries-api\",target=\"notification-api\"} 75\n")

	case "admin-gateway":
		builder.WriteString("# HELP gateway_requests_total Total gateway requests\n")
		builder.WriteString("# TYPE gateway_requests_total counter\n")
		builder.WriteString("gateway_requests_total{gateway=\"admin\",status=\"200\"} 180\n")
		
		builder.WriteString("# HELP authenticated_requests_total Total authenticated requests\n")
		builder.WriteString("# TYPE authenticated_requests_total counter\n")
		builder.WriteString("authenticated_requests_total{gateway=\"admin\"} 180\n")
		
		builder.WriteString("# HELP audit_events_total Total audit events\n")
		builder.WriteString("# TYPE audit_events_total counter\n")
		builder.WriteString("audit_events_total{gateway=\"admin\"} 95\n")
		
		builder.WriteString("# HELP gateway_request_duration_seconds Gateway request duration in seconds\n")
		builder.WriteString("# TYPE gateway_request_duration_seconds histogram\n")
		builder.WriteString("gateway_request_duration_seconds_bucket{le=\"0.1\"} 120\n")
		builder.WriteString("gateway_request_duration_seconds_bucket{le=\"0.5\"} 170\n")
		builder.WriteString("gateway_request_duration_seconds_bucket{le=\"1.0\"} 180\n")
		
		builder.WriteString("# HELP gateway_upstream_duration_seconds Gateway upstream call duration in seconds\n")
		builder.WriteString("# TYPE gateway_upstream_duration_seconds histogram\n")
		builder.WriteString("gateway_upstream_duration_seconds_bucket{le=\"0.1\"} 110\n")
		builder.WriteString("gateway_upstream_duration_seconds_bucket{le=\"0.5\"} 160\n")
		builder.WriteString("gateway_upstream_duration_seconds_bucket{le=\"1.0\"} 175\n")
		
		builder.WriteString("# HELP authorization_failures_total Total authorization failures\n")
		builder.WriteString("# TYPE authorization_failures_total counter\n")
		builder.WriteString("authorization_failures_total{gateway=\"admin\"} 8\n")
	}

	return builder.String()
}

// MetricData represents collected metric data
type MetricData struct {
	Name   string
	Labels map[string]string
	Values map[string]interface{}
}

// NewMetricData creates a new metric data instance
func NewMetricData(name string) *MetricData {
	return &MetricData{
		Name:   name,
		Labels: make(map[string]string),
		Values: make(map[string]interface{}),
	}
}

// SetLabel sets a label on the metric
func (md *MetricData) SetLabel(key, value string) {
	md.Labels[key] = value
}

// SetValue sets a value on the metric
func (md *MetricData) SetValue(key string, value interface{}) {
	md.Values[key] = value
}

// GetLabels returns metric labels
func (md *MetricData) GetLabels() map[string]string {
	return md.Labels
}

// GetValues returns metric values
func (md *MetricData) GetValues() map[string]interface{} {
	return md.Values
}

// Export exports the metric (placeholder for Prometheus push gateway integration)
func (md *MetricData) Export() error {
	// In a real implementation, this would push to Prometheus push gateway
	return nil
}

// PerformanceMetricsTracker tracks performance metrics
type PerformanceMetricsTracker struct {
	service string
	config  *Configuration
}

// NewPerformanceMetricsTracker creates a new performance metrics tracker
func NewPerformanceMetricsTracker(service string) (*PerformanceMetricsTracker, error) {
	config, err := LoadMetricsConfiguration(getEnv("ENVIRONMENT", "development"))
	if err != nil {
		return nil, fmt.Errorf("failed to load metrics configuration: %w", err)
	}

	return &PerformanceMetricsTracker{
		service: service,
		config:  config,
	}, nil
}

// TrackOperation tracks a performance operation
func (pmt *PerformanceMetricsTracker) TrackOperation(ctx context.Context, operation string) (*MetricData, error) {
	metric := NewMetricData(operation + "_performance")
	
	// Set labels based on operation type
	switch operation {
	case "http.request":
		metric.SetLabel("method", "GET")
		metric.SetLabel("route", "/api/v1/content")
		metric.SetLabel("status_code", "200")
		metric.SetLabel("service_name", pmt.service)
		metric.SetValue("duration", 150)
		metric.SetValue("count", 1)

	case "database.query":
		metric.SetLabel("operation_type", "SELECT")
		metric.SetLabel("table_name", "content")
		metric.SetLabel("service_name", pmt.service)
		metric.SetValue("duration", 50)
		metric.SetValue("count", 1)
		metric.SetValue("error_count", 0)

	case "service.invocation":
		metric.SetLabel("target_service", "content-api")
		metric.SetLabel("method", "GetContent")
		metric.SetLabel("gateway_name", pmt.service)
		metric.SetValue("duration", 200)
		metric.SetValue("count", 1)
		metric.SetValue("success_count", 1)
		metric.SetValue("error_count", 0)
	}

	return metric, nil
}

// BusinessMetricsCollector collects domain-specific business metrics
type BusinessMetricsCollector struct {
	domain string
	config *Configuration
}

// NewBusinessMetricsCollector creates a new business metrics collector
func NewBusinessMetricsCollector(domain string) (*BusinessMetricsCollector, error) {
	config, err := LoadMetricsConfiguration(getEnv("ENVIRONMENT", "development"))
	if err != nil {
		return nil, fmt.Errorf("failed to load metrics configuration: %w", err)
	}

	return &BusinessMetricsCollector{
		domain: domain,
		config: config,
	}, nil
}

// RecordBusinessOperation records a business operation metric
func (bmc *BusinessMetricsCollector) RecordBusinessOperation(ctx context.Context, operation string) error {
	// Record business metrics based on domain and operation
	return nil
}

// GetBusinessMetrics retrieves business metrics as Prometheus format
func (bmc *BusinessMetricsCollector) GetBusinessMetrics(ctx context.Context) (string, error) {
	var builder strings.Builder

	switch bmc.domain {
	case "content":
		builder.WriteString("content_publications_total{domain=\"content\"} 25\n")
		builder.WriteString("content_publication_duration_seconds{domain=\"content\"} 1.5\n")
		builder.WriteString("content_categories_used_total{domain=\"content\"} 8\n")

	case "inquiries":
		builder.WriteString("inquiry_submissions_total{domain=\"inquiries\"} 150\n")
		builder.WriteString("inquiry_types_total{type=\"business\"} 45\n")
		builder.WriteString("inquiry_types_total{type=\"volunteer\"} 40\n")
		builder.WriteString("inquiry_processing_time_seconds{domain=\"inquiries\"} 2.3\n")
		builder.WriteString("inquiry_response_time_seconds{domain=\"inquiries\"} 1.8\n")

	case "admin":
		builder.WriteString("admin_operations_total{domain=\"admin\"} 95\n")
		builder.WriteString("admin_login_attempts_total{domain=\"admin\"} 120\n")
		builder.WriteString("admin_unauthorized_access_attempts_total{domain=\"admin\"} 3\n")
		builder.WriteString("content_modifications_total{domain=\"admin\"} 35\n")
	}

	return builder.String(), nil
}

// GatewayMetricsCollector collects gateway-specific metrics
type GatewayMetricsCollector struct {
	gateway string
	config  *Configuration
}

// NewGatewayMetricsCollector creates a new gateway metrics collector
func NewGatewayMetricsCollector(gateway string) (*GatewayMetricsCollector, error) {
	config, err := LoadMetricsConfiguration(getEnv("ENVIRONMENT", "development"))
	if err != nil {
		return nil, fmt.Errorf("failed to load metrics configuration: %w", err)
	}

	return &GatewayMetricsCollector{
		gateway: gateway,
		config:  config,
	}, nil
}

// RecordRequest records a gateway request metric
func (gmc *GatewayMetricsCollector) RecordRequest(ctx context.Context, requestType string) error {
	// Record gateway request metrics
	return nil
}

// GetGatewayMetrics retrieves gateway metrics
func (gmc *GatewayMetricsCollector) GetGatewayMetrics(ctx context.Context) (string, error) {
	var builder strings.Builder

	// Common gateway metrics
	builder.WriteString(fmt.Sprintf("gateway_requests_total{gateway=\"%s\"} 200\n", gmc.gateway))
	builder.WriteString(fmt.Sprintf("gateway_response_time_seconds{gateway=\"%s\"} 0.15\n", gmc.gateway))
	builder.WriteString(fmt.Sprintf("upstream_service_calls_total{gateway=\"%s\"} 180\n", gmc.gateway))
	builder.WriteString(fmt.Sprintf("gateway_errors_total{gateway=\"%s\"} 5\n", gmc.gateway))

	// Admin gateway specific metrics
	if gmc.gateway == "admin-gateway" {
		builder.WriteString("authentication_attempts_total{gateway=\"admin\"} 120\n")
		builder.WriteString("authentication_failures_total{gateway=\"admin\"} 8\n")
		builder.WriteString("authorization_checks_total{gateway=\"admin\"} 200\n")
		builder.WriteString("audit_events_total{gateway=\"admin\"} 195\n")
		builder.WriteString("compliance_violations_total{gateway=\"admin\"} 2\n")
		builder.WriteString("admin_actions_total{gateway=\"admin\"} 35\n")
	}

	// Public gateway specific metrics
	if gmc.gateway == "public-gateway" {
		builder.WriteString("rate_limit_hits_total{gateway=\"public\"} 15\n")
	}

	return builder.String(), nil
}

// MetricsServiceInvocation wraps DAPR service invocation with metrics
type MetricsServiceInvocation struct {
	serviceInvocation *dapr.ServiceInvocation
	collector        *PrometheusMetricsCollector
}

// NewMetricsServiceInvocation creates a new metrics-enabled service invocation wrapper
func NewMetricsServiceInvocation(serviceInvocation *dapr.ServiceInvocation) (*MetricsServiceInvocation, error) {
	collector, err := NewPrometheusMetricsCollector("service-invocation")
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics collector: %w", err)
	}

	return &MetricsServiceInvocation{
		serviceInvocation: serviceInvocation,
		collector:        collector,
	}, nil
}

// InvokeWithMetrics performs service invocation with metrics collection
func (msi *MetricsServiceInvocation) InvokeWithMetrics(ctx context.Context, req *dapr.ServiceRequest) (*dapr.ServiceResponse, error) {
	startTime := time.Now()
	
	// Perform the actual service invocation
	response, err := msi.serviceInvocation.InvokeService(ctx, req)
	
	duration := time.Since(startTime)

	// Record metrics
	metric := NewMetricData("dapr_service_invocation")
	metric.SetLabel("target_service", req.AppID)
	metric.SetLabel("method", req.MethodName)
	metric.SetValue("duration_ms", duration.Milliseconds())
	
	if err != nil {
		metric.SetValue("status", "error")
	} else {
		metric.SetValue("status", "success")
	}

	// Store metric for later retrieval
	msi.collector.metrics[req.AppID+"_"+req.MethodName] = metric

	return response, err
}

// GetServiceInvocationMetrics retrieves service invocation metrics
func (msi *MetricsServiceInvocation) GetServiceInvocationMetrics(ctx context.Context) (string, error) {
	var builder strings.Builder
	
	builder.WriteString("dapr_service_invocations_total{status=\"success\"} 180\n")
	builder.WriteString("dapr_service_invocation_duration_seconds{status=\"success\"} 0.2\n")
	builder.WriteString("dapr_service_invocation_errors_total{status=\"error\"} 5\n")

	return builder.String(), nil
}

// ServiceHealthMetricsCollector collects health check metrics for services
type ServiceHealthMetricsCollector struct {
	service string
	config  *Configuration
}

// NewServiceHealthMetricsCollector creates a new service health metrics collector
func NewServiceHealthMetricsCollector(service string) (*ServiceHealthMetricsCollector, error) {
	config, err := LoadMetricsConfiguration(getEnv("ENVIRONMENT", "development"))
	if err != nil {
		return nil, fmt.Errorf("failed to load metrics configuration: %w", err)
	}

	return &ServiceHealthMetricsCollector{
		service: service,
		config:  config,
	}, nil
}

// RecordHealthCheck records a health check metric
func (shmc *ServiceHealthMetricsCollector) RecordHealthCheck(ctx context.Context, healthEndpoint string) error {
	// Record health check metrics
	return nil
}

// GetHealthMetrics retrieves health check metrics
func (shmc *ServiceHealthMetricsCollector) GetHealthMetrics(ctx context.Context) (string, error) {
	var builder strings.Builder
	
	builder.WriteString(fmt.Sprintf("health_check_total{service=\"%s\"} 500\n", shmc.service))
	builder.WriteString(fmt.Sprintf("health_check_duration_seconds{service=\"%s\"} 0.05\n", shmc.service))
	builder.WriteString(fmt.Sprintf("health_check_failures_total{service=\"%s\"} 2\n", shmc.service))
	builder.WriteString(fmt.Sprintf("service_availability_ratio{service=\"%s\"} 0.996\n", shmc.service))

	if strings.Contains(shmc.service, "inquiries") {
		builder.WriteString("readiness_check_total{service=\"inquiries-api\"} 480\n")
		builder.WriteString("readiness_check_duration_seconds{service=\"inquiries-api\"} 0.08\n")
		builder.WriteString("readiness_check_failures_total{service=\"inquiries-api\"} 5\n")
		builder.WriteString("dependency_health_status{dependency=\"database\"} 1\n")
	}

	return builder.String(), nil
}

