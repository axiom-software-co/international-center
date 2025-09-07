package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/dapr"
	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// DistributedTracingMiddleware provides DAPR middleware integration for distributed tracing
type DistributedTracingMiddleware struct {
	config *Configuration
}

// NewDistributedTracingMiddleware creates a new distributed tracing middleware
func NewDistributedTracingMiddleware() (*DistributedTracingMiddleware, error) {
	config, err := LoadTracingConfiguration(getEnv("ENVIRONMENT", "development"))
	if err != nil {
		return nil, fmt.Errorf("failed to load tracing configuration: %w", err)
	}

	return &DistributedTracingMiddleware{
		config: config,
	}, nil
}

// PropagateTraceHeaders creates and propagates trace headers for HTTP requests
func (dtm *DistributedTracingMiddleware) PropagateTraceHeaders(ctx context.Context, correlationCtx *domain.CorrelationContext) (map[string]string, error) {
	if !dtm.config.GetBool("enabled", true) {
		return nil, fmt.Errorf("distributed tracing is disabled")
	}

	headers := make(map[string]string)

	if correlationCtx != nil {
		headers["X-Correlation-ID"] = correlationCtx.CorrelationID
		headers["X-Trace-ID"] = correlationCtx.TraceID
		headers["X-Request-ID"] = correlationCtx.RequestID
		headers["X-User-ID"] = correlationCtx.UserID
		headers["X-App-Version"] = correlationCtx.AppVersion
	} else {
		// Generate new correlation context for requests without one
		newCtx := domain.NewCorrelationContext()
		headers["X-Correlation-ID"] = newCtx.CorrelationID
		headers["X-Trace-ID"] = newCtx.TraceID
		headers["X-Request-ID"] = newCtx.RequestID
	}

	return headers, nil
}

// TraceSpan represents a distributed trace span
type TraceSpan struct {
	spanID     string
	operation  string
	service    string
	startTime  time.Time
	endTime    *time.Time
	fields     map[string]interface{}
	finished   bool
}

// NewTraceSpan creates a new trace span
func NewTraceSpan(operation, service string) *TraceSpan {
	return &TraceSpan{
		spanID:    domain.GetTraceID(context.Background()), // Generate unique span ID
		operation: operation,
		service:   service,
		startTime: time.Now().UTC(),
		fields:    make(map[string]interface{}),
	}
}

// SetField sets a field on the trace span
func (ts *TraceSpan) SetField(key string, value interface{}) {
	ts.fields[key] = value
}

// GetFields returns all fields in the trace span
func (ts *TraceSpan) GetFields() map[string]interface{} {
	// Ensure required fields are present
	if ts.fields["service.name"] == "" {
		ts.fields["service.name"] = ts.service
	}
	if ts.fields["operation.name"] == "" {
		ts.fields["operation.name"] = ts.operation
	}
	
	// Always calculate and set span duration
	var durationMs int64
	if ts.endTime != nil {
		// Span is finished, use exact duration
		duration := ts.endTime.Sub(ts.startTime)
		durationMs = duration.Milliseconds()
	} else {
		// Span is still active, calculate current duration
		duration := time.Now().UTC().Sub(ts.startTime)
		durationMs = duration.Milliseconds()
	}
	
	// Ensure minimum duration of 1ms for observability
	if durationMs == 0 {
		durationMs = 1
	}
	ts.fields["span.duration"] = durationMs

	return ts.fields
}

// Finish marks the span as completed
func (ts *TraceSpan) Finish() error {
	if ts.finished {
		return fmt.Errorf("span already finished")
	}

	now := time.Now().UTC()
	ts.endTime = &now
	ts.finished = true

	// Set final span fields
	ts.fields["span.duration"] = ts.endTime.Sub(ts.startTime).Milliseconds()
	ts.fields["span.finish_time"] = ts.endTime.Format(time.RFC3339)

	return nil
}

// DistributedTracer provides distributed tracing capabilities
type DistributedTracer struct {
	config *Configuration
}

// NewDistributedTracer creates a new distributed tracer
func NewDistributedTracer() (*DistributedTracer, error) {
	config, err := LoadTracingConfiguration(getEnv("ENVIRONMENT", "development"))
	if err != nil {
		return nil, fmt.Errorf("failed to load tracing configuration: %w", err)
	}

	return &DistributedTracer{
		config: config,
	}, nil
}

// StartSpan starts a new distributed trace span
func (dt *DistributedTracer) StartSpan(ctx context.Context, operation, service string) (*TraceSpan, error) {
	span := NewTraceSpan(operation, service)

	// Add correlation context to span
	correlationCtx := domain.FromContext(ctx)
	span.SetField("correlation.id", correlationCtx.CorrelationID)
	span.SetField("trace.id", correlationCtx.TraceID)
	span.SetField("service.name", service)
	span.SetField("operation.name", operation)

	// Set operation-specific fields
	switch operation {
	case "database.query":
		span.SetField("database.operation.type", "query")
	case "dapr.service.invoke":
		span.SetField("dapr.target.service", service)
	}

	return span, nil
}

// TracedServiceInvocation wraps DAPR service invocation with tracing
type TracedServiceInvocation struct {
	serviceInvocation *dapr.ServiceInvocation
	tracer           *DistributedTracer
}

// NewTracedServiceInvocation creates a new traced service invocation wrapper
func NewTracedServiceInvocation(serviceInvocation *dapr.ServiceInvocation) (*TracedServiceInvocation, error) {
	tracer, err := NewDistributedTracer()
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer: %w", err)
	}

	return &TracedServiceInvocation{
		serviceInvocation: serviceInvocation,
		tracer:           tracer,
	}, nil
}

// InvokeWithTracing performs a service invocation with automatic tracing
func (tsi *TracedServiceInvocation) InvokeWithTracing(ctx context.Context, req *dapr.ServiceRequest) (*dapr.ServiceResponse, *TraceSpan, error) {
	// Start trace span
	span, err := tsi.tracer.StartSpan(ctx, "dapr.service.invoke", req.AppID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start trace span: %w", err)
	}

	// Add service invocation metadata to span
	span.SetField("service.invocation.target", req.AppID)
	span.SetField("service.invocation.method", req.MethodName)
	span.SetField("service.invocation.verb", req.HTTPVerb)

	// Perform the actual service invocation
	startTime := time.Now()
	response, invokeErr := tsi.serviceInvocation.InvokeService(ctx, req)
	duration := time.Since(startTime)

	// Add timing and status to span
	span.SetField("service.invocation.duration", duration.Milliseconds())
	if invokeErr != nil {
		span.SetField("service.invocation.status", "error")
		span.SetField("service.invocation.error", invokeErr.Error())
	} else {
		span.SetField("service.invocation.status", "success")
	}

	// Finish the span
	span.Finish()

	return response, span, invokeErr
}

// GatewayTracer provides gateway-specific tracing
type GatewayTracer struct {
	gateway string
	tracer  *DistributedTracer
}

// NewGatewayTracer creates a new gateway tracer
func NewGatewayTracer(gateway string) (*GatewayTracer, error) {
	tracer, err := NewDistributedTracer()
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer: %w", err)
	}

	return &GatewayTracer{
		gateway: gateway,
		tracer:  tracer,
	}, nil
}

// TraceRequest traces a gateway request with optional audit trail
func (gt *GatewayTracer) TraceRequest(ctx context.Context, endpoint string) (*TraceSpan, *AuditTrail, error) {
	// Start trace span for gateway request
	span, err := gt.tracer.StartSpan(ctx, "gateway.request", gt.gateway)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start gateway trace span: %w", err)
	}

	// Add gateway-specific fields
	span.SetField("gateway.name", gt.gateway)
	span.SetField("gateway.endpoint", endpoint)
	span.SetField("request.timestamp", time.Now().UTC().Format(time.RFC3339))

	// Create audit trail for admin gateway
	var auditTrail *AuditTrail
	if gt.gateway == "admin-gateway" {
		auditTrail = &AuditTrail{
			Gateway:  gt.gateway,
			Endpoint: endpoint,
			fields:   make(map[string]interface{}),
		}
		auditTrail.SetComplianceField("audit.gateway", gt.gateway)
		auditTrail.SetComplianceField("audit.endpoint", endpoint)
		auditTrail.SetComplianceField("audit.timestamp", time.Now().UTC().Format(time.RFC3339))
	}

	return span, auditTrail, nil
}

// AuditTrail represents an audit trail for compliance
type AuditTrail struct {
	Gateway  string
	Endpoint string
	fields   map[string]interface{}
}

// SetComplianceField sets a compliance-related field
func (at *AuditTrail) SetComplianceField(key string, value interface{}) {
	at.fields[key] = value
}

// GetComplianceFields returns all compliance fields
func (at *AuditTrail) GetComplianceFields() map[string]interface{} {
	return at.fields
}