package domain

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// ContextKey represents a context key for correlation data
type ContextKey string

const (
	ContextKeyCorrelationID ContextKey = "correlation_id"
	ContextKeyTraceID       ContextKey = "trace_id"
	ContextKeyUserID        ContextKey = "user_id"
	ContextKeyRequestID     ContextKey = "request_id"
	ContextKeyAppVersion    ContextKey = "app_version"
)

// CorrelationContext contains correlation information for distributed tracing and request tracking.
// This context flows through all backend services to provide comprehensive request correlation,
// performance monitoring, and debugging capabilities across service boundaries.
//
// The context includes unique identifiers for correlation, tracing, and request tracking,
// along with user context and timing information for complete request lifecycle tracking.
type CorrelationContext struct {
	CorrelationID string    // Unique identifier that spans the entire user request across all services
	TraceID       string    // Distributed tracing identifier for monitoring and debugging
	UserID        string    // User identifier for security and audit context
	RequestID     string    // Unique identifier for this specific service request
	AppVersion    string    // Application version for compatibility tracking
	StartTime     time.Time // Request start time for performance monitoring
}

// NewCorrelationContext creates a new correlation context with generated IDs
func NewCorrelationContext() *CorrelationContext {
	return &CorrelationContext{
		CorrelationID: uuid.New().String(),
		TraceID:       generateTraceID(),
		RequestID:     uuid.New().String(),
		StartTime:     time.Now().UTC(),
	}
}

// NewCorrelationContextWithID creates a new correlation context with provided correlation ID
func NewCorrelationContextWithID(correlationID string) *CorrelationContext {
	return &CorrelationContext{
		CorrelationID: correlationID,
		TraceID:       generateTraceID(),
		RequestID:     uuid.New().String(),
		StartTime:     time.Now().UTC(),
	}
}

// SetUserContext sets user information in the correlation context
func (c *CorrelationContext) SetUserContext(userID, appVersion string) {
	c.UserID = userID
	c.AppVersion = appVersion
}

// ToContext adds correlation information to a context
func (c *CorrelationContext) ToContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, ContextKeyCorrelationID, c.CorrelationID)
	ctx = context.WithValue(ctx, ContextKeyTraceID, c.TraceID)
	ctx = context.WithValue(ctx, ContextKeyUserID, c.UserID)
	ctx = context.WithValue(ctx, ContextKeyRequestID, c.RequestID)
	ctx = context.WithValue(ctx, ContextKeyAppVersion, c.AppVersion)
	return ctx
}

// FromContext extracts correlation information from a context
func FromContext(ctx context.Context) *CorrelationContext {
	correlation := &CorrelationContext{
		StartTime: time.Now().UTC(),
	}
	
	if correlationID, ok := ctx.Value(ContextKeyCorrelationID).(string); ok {
		correlation.CorrelationID = correlationID
	}
	
	if traceID, ok := ctx.Value(ContextKeyTraceID).(string); ok {
		correlation.TraceID = traceID
	}
	
	if userID, ok := ctx.Value(ContextKeyUserID).(string); ok {
		correlation.UserID = userID
	}
	
	if requestID, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		correlation.RequestID = requestID
	}
	
	if appVersion, ok := ctx.Value(ContextKeyAppVersion).(string); ok {
		correlation.AppVersion = appVersion
	}
	
	// Generate missing IDs
	if correlation.CorrelationID == "" {
		correlation.CorrelationID = uuid.New().String()
	}
	
	if correlation.TraceID == "" {
		correlation.TraceID = generateTraceID()
	}
	
	if correlation.RequestID == "" {
		correlation.RequestID = uuid.New().String()
	}
	
	return correlation
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(ContextKeyCorrelationID).(string); ok {
		return correlationID
	}
	return uuid.New().String()
}

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(ContextKeyTraceID).(string); ok {
		return traceID
	}
	return generateTraceID()
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return userID
	}
	return ""
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		return requestID
	}
	return uuid.New().String()
}

// GetAppVersion extracts app version from context
func GetAppVersion(ctx context.Context) string {
	if appVersion, ok := ctx.Value(ContextKeyAppVersion).(string); ok {
		return appVersion
	}
	return "unknown"
}

// WithCorrelationID adds correlation ID to context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, ContextKeyCorrelationID, correlationID)
}

// WithTraceID adds trace ID to context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ContextKeyTraceID, traceID)
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextKeyRequestID, requestID)
}

// WithAppVersion adds app version to context
func WithAppVersion(ctx context.Context, appVersion string) context.Context {
	return context.WithValue(ctx, ContextKeyAppVersion, appVersion)
}

// CreateChildContext creates a child context with new trace but same correlation
func CreateChildContext(parentCtx context.Context) context.Context {
	correlationID := GetCorrelationID(parentCtx)
	userID := GetUserID(parentCtx)
	appVersion := GetAppVersion(parentCtx)
	
	childCorrelation := &CorrelationContext{
		CorrelationID: correlationID,
		TraceID:       generateTraceID(),
		UserID:        userID,
		RequestID:     uuid.New().String(),
		AppVersion:    appVersion,
		StartTime:     time.Now().UTC(),
	}
	
	return childCorrelation.ToContext(context.Background())
}

// GetElapsedTime calculates elapsed time from correlation context
func (c *CorrelationContext) GetElapsedTime() time.Duration {
	return time.Since(c.StartTime)
}

// ToLogFields returns correlation fields for structured logging
func (c *CorrelationContext) ToLogFields() map[string]interface{} {
	return map[string]interface{}{
		"correlation_id": c.CorrelationID,
		"trace_id":       c.TraceID,
		"user_id":        c.UserID,
		"request_id":     c.RequestID,
		"app_version":    c.AppVersion,
		"elapsed_ms":     c.GetElapsedTime().Milliseconds(),
	}
}

// generateTraceID generates a trace ID compatible with distributed tracing
// Returns a 32-character hex string optimized for performance
func generateTraceID() string {
	// Generate a 32-character hex string for trace ID
	// Optimized to avoid string formatting overhead
	id := uuid.New()
	var buf [32]byte
	hex.Encode(buf[:], id[:])
	return string(buf[:])
}