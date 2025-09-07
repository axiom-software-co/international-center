package observability

import (
	"context"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TDD Cycle 7 RED Phase: Service Metrics Collection Contract Tests
//
// WHY: Performance monitoring and SLA compliance require comprehensive metrics across all services
// SCOPE: All services (content-api, inquiries-api, notification-api) and gateways (admin/public)
// DEPENDENCIES: Prometheus endpoint integration, service health monitoring
// CONTEXT: Gateway architecture requiring performance visibility and alerting capabilities

func TestPrometheusMetricsEndpoint_Contract(t *testing.T) {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tests := []struct {
		name             string
		service          string
		expectedMetrics  []string
		endpoint         string
		expectedFormat   string
	}{
		{
			name:     "content-api prometheus metrics endpoint",
			service:  "content-api",
			endpoint: "/metrics",
			expectedFormat: "text/plain",
			expectedMetrics: []string{
				"http_requests_total",
				"http_request_duration_seconds",
				"service_uptime_seconds",
				"active_connections",
				"dapr_service_invocations_total",
				"database_queries_total",
				"database_query_duration_seconds",
			},
		},
		{
			name:     "inquiries-api prometheus metrics endpoint",
			service:  "inquiries-api",
			endpoint: "/metrics",
			expectedFormat: "text/plain",
			expectedMetrics: []string{
				"http_requests_total",
				"http_request_duration_seconds",
				"service_uptime_seconds",
				"inquiry_submissions_total",
				"inquiry_processing_duration_seconds",
				"dapr_service_invocations_total",
			},
		},
		{
			name:     "admin-gateway prometheus metrics endpoint",
			service:  "admin-gateway",
			endpoint: "/metrics",
			expectedFormat: "text/plain",
			expectedMetrics: []string{
				"gateway_requests_total",
				"gateway_request_duration_seconds",
				"gateway_upstream_duration_seconds",
				"authenticated_requests_total",
				"authorization_failures_total",
				"audit_events_total",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TDD RED Phase - This will fail until implementation exists
			metricsCollector, err := NewPrometheusMetricsCollector(tt.service)
			require.NoError(t, err, "metrics collector creation should succeed")
			require.NotNil(t, metricsCollector, "metrics collector should not be nil")

			// Contract: Service must expose Prometheus metrics endpoint
			metrics, contentType, err := metricsCollector.GetMetrics(ctx, tt.endpoint)
			assert.NoError(t, err, "metrics endpoint should be accessible")
			assert.Equal(t, tt.expectedFormat, contentType, "metrics should be in Prometheus text format")
			assert.NotEmpty(t, metrics, "metrics endpoint should return data")

			// Contract: Required metrics must be present
			for _, metric := range tt.expectedMetrics {
				assert.Contains(t, metrics, metric, "metric %s should be present", metric)
			}
		})
	}
}

func TestServicePerformanceMetrics_Contract(t *testing.T) {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tests := []struct {
		name            string
		operation       string
		service         string
		expectedLabels  []string
		expectedValues  []string
	}{
		{
			name:      "HTTP request performance metrics",
			operation: "http.request",
			service:   "content-api",
			expectedLabels: []string{
				"method",
				"route",
				"status_code",
				"service_name",
			},
			expectedValues: []string{
				"duration",
				"count",
			},
		},
		{
			name:      "database operation performance metrics",
			operation: "database.query",
			service:   "inquiries-api",
			expectedLabels: []string{
				"operation_type",
				"table_name",
				"service_name",
			},
			expectedValues: []string{
				"duration",
				"count",
				"error_count",
			},
		},
		{
			name:      "service invocation performance metrics",
			operation: "service.invocation",
			service:   "admin-gateway",
			expectedLabels: []string{
				"target_service",
				"method",
				"gateway_name",
			},
			expectedValues: []string{
				"duration",
				"count",
				"success_count",
				"error_count",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TDD RED Phase - This will fail until implementation exists
			performanceTracker, err := NewPerformanceMetricsTracker(tt.service)
			require.NoError(t, err, "performance tracker creation should succeed")
			require.NotNil(t, performanceTracker, "performance tracker should not be nil")

			// Contract: Performance tracking must capture operation metrics
			metric, err := performanceTracker.TrackOperation(ctx, tt.operation)
			assert.NoError(t, err, "operation tracking should succeed")
			assert.NotNil(t, metric, "metric should be created")

			// Contract: Metrics must contain required labels and values
			labels := metric.GetLabels()
			for _, label := range tt.expectedLabels {
				assert.Contains(t, labels, label, "metric should contain label %s", label)
			}

			values := metric.GetValues()
			for _, value := range tt.expectedValues {
				assert.Contains(t, values, value, "metric should contain value %s", value)
			}

			// Contract: Metrics must be exportable to Prometheus
			err = metric.Export()
			assert.NoError(t, err, "metric export should succeed")
		})
	}
}

func TestBusinessMetrics_Contract(t *testing.T) {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tests := []struct {
		name           string
		domain         string
		operation      string
		expectedMetrics []string
	}{
		{
			name:      "content publication business metrics",
			domain:    "content",
			operation: "content.publish",
			expectedMetrics: []string{
				"content_publications_total",
				"content_publication_duration_seconds",
				"content_categories_used_total",
			},
		},
		{
			name:      "inquiry submission business metrics",
			domain:    "inquiries",
			operation: "inquiry.submit",
			expectedMetrics: []string{
				"inquiry_submissions_total",
				"inquiry_types_total",
				"inquiry_processing_time_seconds",
				"inquiry_response_time_seconds",
			},
		},
		{
			name:      "admin operations business metrics",
			domain:    "admin",
			operation: "admin.audit",
			expectedMetrics: []string{
				"admin_operations_total",
				"admin_login_attempts_total",
				"admin_unauthorized_access_attempts_total",
				"content_modifications_total",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TDD RED Phase - This will fail until implementation exists
			businessMetrics, err := NewBusinessMetricsCollector(tt.domain)
			require.NoError(t, err, "business metrics collector creation should succeed")
			require.NotNil(t, businessMetrics, "business metrics collector should not be nil")

			// Contract: Business operations must generate domain-specific metrics
			err = businessMetrics.RecordBusinessOperation(ctx, tt.operation)
			assert.NoError(t, err, "business operation recording should succeed")

			// Contract: Business metrics must be available via Prometheus endpoint
			metrics, err := businessMetrics.GetBusinessMetrics(ctx)
			assert.NoError(t, err, "business metrics retrieval should succeed")
			assert.NotEmpty(t, metrics, "business metrics should be available")

			// Contract: Domain-specific metrics must be present
			for _, metric := range tt.expectedMetrics {
				assert.Contains(t, metrics, metric, "business metric %s should be present", metric)
			}
		})
	}
}

func TestGatewayMetrics_Contract(t *testing.T) {
	timeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tests := []struct {
		name               string
		gateway            string
		requestType        string
		expectedMetrics    []string
		expectAuthMetrics  bool
		expectAuditMetrics bool
	}{
		{
			name:        "admin gateway metrics with authentication and audit",
			gateway:     "admin-gateway",
			requestType: "admin.api.request",
			expectedMetrics: []string{
				"gateway_requests_total",
				"gateway_response_time_seconds",
				"upstream_service_calls_total",
				"gateway_errors_total",
			},
			expectAuthMetrics:  true,
			expectAuditMetrics: true,
		},
		{
			name:        "public gateway metrics without authentication",
			gateway:     "public-gateway",
			requestType: "public.api.request",
			expectedMetrics: []string{
				"gateway_requests_total",
				"gateway_response_time_seconds",
				"rate_limit_hits_total",
				"upstream_service_calls_total",
			},
			expectAuthMetrics:  false,
			expectAuditMetrics: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TDD RED Phase - This will fail until implementation exists
			gatewayMetrics, err := NewGatewayMetricsCollector(tt.gateway)
			require.NoError(t, err, "gateway metrics collector creation should succeed")
			require.NotNil(t, gatewayMetrics, "gateway metrics collector should not be nil")

			// Contract: Gateway must track all request metrics
			err = gatewayMetrics.RecordRequest(ctx, tt.requestType)
			assert.NoError(t, err, "gateway request recording should succeed")

			// Contract: Gateway metrics must be comprehensive
			metrics, err := gatewayMetrics.GetGatewayMetrics(ctx)
			assert.NoError(t, err, "gateway metrics retrieval should succeed")

			for _, metric := range tt.expectedMetrics {
				assert.Contains(t, metrics, metric, "gateway metric %s should be present", metric)
			}

			// Contract: Admin gateway must include authentication metrics
			if tt.expectAuthMetrics {
				authMetrics := []string{
					"authentication_attempts_total",
					"authentication_failures_total",
					"authorization_checks_total",
				}
				for _, metric := range authMetrics {
					assert.Contains(t, metrics, metric, "auth metric %s should be present for admin gateway", metric)
				}
			}

			// Contract: Admin gateway must include audit metrics
			if tt.expectAuditMetrics {
				auditMetrics := []string{
					"audit_events_total",
					"compliance_violations_total",
					"admin_actions_total",
				}
				for _, metric := range auditMetrics {
					assert.Contains(t, metrics, metric, "audit metric %s should be present for admin gateway", metric)
				}
			}
		})
	}
}

func TestDaprServiceInvocationMetrics_Contract(t *testing.T) {
	timeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Integration with existing DAPR service invocation
	client := &dapr.Client{}
	serviceInvocation := dapr.NewServiceInvocation(client)

	tests := []struct {
		name            string
		targetService   string
		method          string
		expectedMetrics []string
	}{
		{
			name:          "content-api service invocation metrics",
			targetService: "content-api",
			method:        "api/v1/content",
			expectedMetrics: []string{
				"dapr_service_invocations_total",
				"dapr_service_invocation_duration_seconds",
				"dapr_service_invocation_errors_total",
			},
		},
		{
			name:          "inquiries-api service invocation metrics",
			targetService: "inquiries-api",
			method:        "api/v1/inquiries/business",
			expectedMetrics: []string{
				"dapr_service_invocations_total",
				"dapr_service_invocation_duration_seconds",
				"dapr_service_invocation_errors_total",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TDD RED Phase - This will fail until implementation exists
			metricsInvocation, err := NewMetricsServiceInvocation(serviceInvocation)
			require.NoError(t, err, "metrics service invocation creation should succeed")
			require.NotNil(t, metricsInvocation, "metrics invocation should not be nil")

			// Contract: Service invocation must be automatically instrumented with metrics
			req := &dapr.ServiceRequest{
				AppID:      tt.targetService,
				MethodName: tt.method,
				HTTPVerb:   "GET",
			}

			_, err = metricsInvocation.InvokeWithMetrics(ctx, req)
			// Note: This may fail in test mode but should record metrics attempt
			
			// Contract: Metrics must be recorded for service invocations
			metrics, err := metricsInvocation.GetServiceInvocationMetrics(ctx)
			assert.NoError(t, err, "service invocation metrics retrieval should succeed")

			for _, metric := range tt.expectedMetrics {
				assert.Contains(t, metrics, metric, "service invocation metric %s should be present", metric)
			}
		})
	}
}

func TestMetricsConfiguration_Contract(t *testing.T) {
	tests := []struct {
		name           string
		environment    string
		expectedConfig map[string]interface{}
	}{
		{
			name:        "development environment metrics configuration",
			environment: "development",
			expectedConfig: map[string]interface{}{
				"enabled":           true,
				"collection_interval": "15s",
				"retention_period":    "7d",
				"export_endpoint":     "localhost:9090",
			},
		},
		{
			name:        "production environment metrics configuration",
			environment: "production",
			expectedConfig: map[string]interface{}{
				"enabled":           true,
				"collection_interval": "30s",
				"retention_period":    "90d",
				"export_endpoint":     "grafana-cloud:443",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TDD RED Phase - This will fail until implementation exists
			config, err := LoadMetricsConfiguration(tt.environment)
			assert.NoError(t, err, "metrics configuration loading should succeed")
			assert.NotNil(t, config, "configuration should not be nil")

			// Contract: Configuration must contain required metrics settings
			for key, expectedValue := range tt.expectedConfig {
				actualValue, exists := config.GetValue(key)
				assert.True(t, exists, "configuration should contain key %s", key)
				assert.Equal(t, expectedValue, actualValue, "configuration value for %s should match expected", key)
			}
		})
	}
}

func TestHealthCheckMetrics_Contract(t *testing.T) {
	timeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tests := []struct {
		name            string
		service         string
		healthEndpoint  string
		expectedMetrics []string
	}{
		{
			name:           "content-api health check metrics",
			service:        "content-api",
			healthEndpoint: "/health",
			expectedMetrics: []string{
				"health_check_total",
				"health_check_duration_seconds",
				"health_check_failures_total",
				"service_availability_ratio",
			},
		},
		{
			name:           "inquiries-api readiness check metrics", 
			service:        "inquiries-api",
			healthEndpoint: "/health/ready",
			expectedMetrics: []string{
				"readiness_check_total",
				"readiness_check_duration_seconds",
				"readiness_check_failures_total",
				"dependency_health_status",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TDD RED Phase - This will fail until implementation exists
			healthMetrics, err := NewServiceHealthMetricsCollector(tt.service)
			require.NoError(t, err, "service health metrics collector creation should succeed")
			require.NotNil(t, healthMetrics, "service health metrics collector should not be nil")

			// Contract: Health checks must be instrumented with metrics
			err = healthMetrics.RecordHealthCheck(ctx, tt.healthEndpoint)
			assert.NoError(t, err, "health check recording should succeed")

			// Contract: Health metrics must be available
			metrics, err := healthMetrics.GetHealthMetrics(ctx)
			assert.NoError(t, err, "health metrics retrieval should succeed")

			for _, metric := range tt.expectedMetrics {
				assert.Contains(t, metrics, metric, "health metric %s should be present", metric)
			}
		})
	}
}