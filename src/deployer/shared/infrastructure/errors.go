package infrastructure

import (
	"fmt"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ErrorCategory defines the type of infrastructure error
type ErrorCategory string

const (
	ErrorCategoryValidation     ErrorCategory = "validation"
	ErrorCategoryConfiguration  ErrorCategory = "configuration"
	ErrorCategoryResource       ErrorCategory = "resource"
	ErrorCategoryNetwork        ErrorCategory = "network"
	ErrorCategoryDeployment     ErrorCategory = "deployment"
	ErrorCategorySecurity       ErrorCategory = "security"
	ErrorCategoryObservability  ErrorCategory = "observability"
)

// ErrorSeverity defines the severity level of the error
type ErrorSeverity string

const (
	ErrorSeverityCritical ErrorSeverity = "critical"
	ErrorSeverityHigh     ErrorSeverity = "high"
	ErrorSeverityMedium   ErrorSeverity = "medium"
	ErrorSeverityLow      ErrorSeverity = "low"
)

// InfrastructureError provides structured error information
type InfrastructureError struct {
	Category      ErrorCategory          `json:"category"`
	Severity      ErrorSeverity          `json:"severity"`
	Operation     string                 `json:"operation"`
	Component     string                 `json:"component"`
	Environment   string                 `json:"environment"`
	ResourceName  string                 `json:"resource_name,omitempty"`
	Message       string                 `json:"message"`
	OriginalError error                  `json:"-"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Retryable     bool                   `json:"retryable"`
	RecoveryHint  string                 `json:"recovery_hint,omitempty"`
}

// Error implements the error interface
func (e *InfrastructureError) Error() string {
	if e.OriginalError != nil {
		return fmt.Sprintf("[%s:%s] %s failed in %s/%s: %s (original: %v)", 
			e.Category, e.Severity, e.Operation, e.Environment, e.Component, e.Message, e.OriginalError)
	}
	return fmt.Sprintf("[%s:%s] %s failed in %s/%s: %s", 
		e.Category, e.Severity, e.Operation, e.Environment, e.Component, e.Message)
}

// Unwrap returns the original error for error chain compatibility
func (e *InfrastructureError) Unwrap() error {
	return e.OriginalError
}

// ErrorBuilder provides fluent interface for building infrastructure errors
type ErrorBuilder struct {
	err *InfrastructureError
}

// NewError creates a new error builder
func NewError(category ErrorCategory, operation, component, environment string) *ErrorBuilder {
	return &ErrorBuilder{
		err: &InfrastructureError{
			Category:    category,
			Severity:    ErrorSeverityMedium, // Default severity
			Operation:   operation,
			Component:   component,
			Environment: environment,
			Timestamp:   time.Now(),
			Context:     make(map[string]interface{}),
		},
	}
}

// WithSeverity sets the error severity
func (b *ErrorBuilder) WithSeverity(severity ErrorSeverity) *ErrorBuilder {
	b.err.Severity = severity
	return b
}

// WithResourceName sets the resource name
func (b *ErrorBuilder) WithResourceName(name string) *ErrorBuilder {
	b.err.ResourceName = name
	return b
}

// WithMessage sets the error message
func (b *ErrorBuilder) WithMessage(message string) *ErrorBuilder {
	b.err.Message = message
	return b
}

// WithOriginalError sets the underlying error
func (b *ErrorBuilder) WithOriginalError(err error) *ErrorBuilder {
	b.err.OriginalError = err
	return b
}

// WithContext adds context information
func (b *ErrorBuilder) WithContext(key string, value interface{}) *ErrorBuilder {
	b.err.Context[key] = value
	return b
}

// WithCorrelationID sets the correlation ID
func (b *ErrorBuilder) WithCorrelationID(id string) *ErrorBuilder {
	b.err.CorrelationID = id
	return b
}

// WithRetryable marks the error as retryable
func (b *ErrorBuilder) WithRetryable(retryable bool) *ErrorBuilder {
	b.err.Retryable = retryable
	return b
}

// WithRecoveryHint provides recovery guidance
func (b *ErrorBuilder) WithRecoveryHint(hint string) *ErrorBuilder {
	b.err.RecoveryHint = hint
	return b
}

// Build returns the constructed error
func (b *ErrorBuilder) Build() *InfrastructureError {
	return b.err
}

// ErrorLogger provides structured logging for infrastructure errors
type ErrorLogger struct {
	ctx *pulumi.Context
	environment string
	component   string
}

// NewErrorLogger creates a new error logger
func NewErrorLogger(ctx *pulumi.Context, environment, component string) *ErrorLogger {
	return &ErrorLogger{
		ctx:         ctx,
		environment: environment,
		component:   component,
	}
}

// LogError logs an infrastructure error with structured information
func (l *ErrorLogger) LogError(err *InfrastructureError) {
	logLevel := l.getSeverityLogLevel(err.Severity)
	
	// Create structured log message
	structuredMessage := fmt.Sprintf("Infrastructure error: %s [category=%s severity=%s operation=%s component=%s environment=%s timestamp=%s retryable=%t",
		err.Error(),
		err.Category,
		err.Severity,
		err.Operation,
		err.Component,
		err.Environment,
		err.Timestamp.Format(time.RFC3339),
		err.Retryable)
	
	// Add optional fields
	if err.ResourceName != "" {
		structuredMessage += fmt.Sprintf(" resource=%s", err.ResourceName)
	}
	if err.CorrelationID != "" {
		structuredMessage += fmt.Sprintf(" correlation_id=%s", err.CorrelationID)
	}
	if err.RecoveryHint != "" {
		structuredMessage += fmt.Sprintf(" recovery_hint=%s", err.RecoveryHint)
	}
	
	// Add context information
	for key, value := range err.Context {
		structuredMessage += fmt.Sprintf(" %s=%v", key, value)
	}
	
	structuredMessage += "]"
	
	// Log based on severity using simple string logging
	switch logLevel {
	case "error":
		l.ctx.Log.Error(structuredMessage, nil)
	case "warn":
		l.ctx.Log.Warn(structuredMessage, nil)
	case "info":
		l.ctx.Log.Info(structuredMessage, nil)
	default:
		l.ctx.Log.Debug(structuredMessage, nil)
	}
}

// getSeverityLogLevel maps error severity to log level
func (l *ErrorLogger) getSeverityLogLevel(severity ErrorSeverity) string {
	switch severity {
	case ErrorSeverityCritical, ErrorSeverityHigh:
		return "error"
	case ErrorSeverityMedium:
		return "warn"
	case ErrorSeverityLow:
		return "info"
	default:
		return "debug"
	}
}

// ErrorHandler provides centralized error handling for infrastructure operations
type ErrorHandler struct {
	logger      *ErrorLogger
	environment string
	component   string
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(ctx *pulumi.Context, environment, component string) *ErrorHandler {
	return &ErrorHandler{
		logger:      NewErrorLogger(ctx, environment, component),
		environment: environment,
		component:   component,
	}
}

// HandleError processes and logs infrastructure errors
func (h *ErrorHandler) HandleError(err *InfrastructureError) error {
	// Log the error
	h.logger.LogError(err)
	
	// Apply environment-specific handling
	switch h.environment {
	case "development":
		// In development, provide more detailed error information
		return fmt.Errorf("%s - %s", err.Error(), h.getDetailedErrorInfo(err))
	case "staging":
		// In staging, provide moderate detail for debugging
		return fmt.Errorf("%s", err.Error())
	case "production":
		// In production, provide minimal detail but ensure correlation ID is included
		if err.CorrelationID != "" {
			return fmt.Errorf("infrastructure operation failed (correlation_id: %s)", err.CorrelationID)
		}
		return fmt.Errorf("infrastructure operation failed in %s", err.Component)
	default:
		return err
	}
}

// getDetailedErrorInfo provides detailed error information for development
func (h *ErrorHandler) getDetailedErrorInfo(err *InfrastructureError) string {
	details := fmt.Sprintf("Component: %s, Resource: %s, Retryable: %t", 
		err.Component, err.ResourceName, err.Retryable)
	
	if err.RecoveryHint != "" {
		details += fmt.Sprintf(", Recovery: %s", err.RecoveryHint)
	}
	
	return details
}

// Common error creation helpers

// NewResourceError creates a resource-related error
func NewResourceError(operation, component, environment, resourceName string, originalErr error) *InfrastructureError {
	return NewError(ErrorCategoryResource, operation, component, environment).
		WithResourceName(resourceName).
		WithOriginalError(originalErr).
		WithSeverity(ErrorSeverityHigh).
		WithRetryable(true).
		WithRecoveryHint("Check resource configuration and dependencies").
		Build()
}

// NewValidationError creates a validation error
func NewValidationError(operation, component, environment string, message string) *InfrastructureError {
	return NewError(ErrorCategoryValidation, operation, component, environment).
		WithMessage(message).
		WithSeverity(ErrorSeverityMedium).
		WithRetryable(false).
		WithRecoveryHint("Review and correct configuration parameters").
		Build()
}

// NewConfigurationError creates a configuration error
func NewConfigurationError(operation, component, environment, configKey string, originalErr error) *InfrastructureError {
	return NewError(ErrorCategoryConfiguration, operation, component, environment).
		WithMessage(fmt.Sprintf("Configuration error for key: %s", configKey)).
		WithOriginalError(originalErr).
		WithContext("config_key", configKey).
		WithSeverity(ErrorSeverityHigh).
		WithRetryable(false).
		WithRecoveryHint("Verify environment configuration and required variables").
		Build()
}

// NewNetworkError creates a network-related error
func NewNetworkError(operation, component, environment, networkResource string, originalErr error) *InfrastructureError {
	return NewError(ErrorCategoryNetwork, operation, component, environment).
		WithResourceName(networkResource).
		WithOriginalError(originalErr).
		WithSeverity(ErrorSeverityHigh).
		WithRetryable(true).
		WithRecoveryHint("Check network connectivity and firewall rules").
		Build()
}

// NewDeploymentError creates a deployment error
func NewDeploymentError(operation, component, environment string, originalErr error) *InfrastructureError {
	return NewError(ErrorCategoryDeployment, operation, component, environment).
		WithOriginalError(originalErr).
		WithSeverity(ErrorSeverityCritical).
		WithRetryable(true).
		WithRecoveryHint("Review deployment configuration and resource availability").
		Build()
}

// WrapError wraps a standard error into an InfrastructureError
func WrapError(category ErrorCategory, operation, component, environment string, originalErr error) *InfrastructureError {
	if originalErr == nil {
		return nil
	}
	
	// If it's already an InfrastructureError, return it
	if infraErr, ok := originalErr.(*InfrastructureError); ok {
		return infraErr
	}
	
	return NewError(category, operation, component, environment).
		WithOriginalError(originalErr).
		WithMessage(originalErr.Error()).
		Build()
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if infraErr, ok := err.(*InfrastructureError); ok {
		return infraErr.Retryable
	}
	return false
}

// GetErrorCategory extracts the error category
func GetErrorCategory(err error) ErrorCategory {
	if infraErr, ok := err.(*InfrastructureError); ok {
		return infraErr.Category
	}
	return ErrorCategoryResource // Default category
}

// GetCorrelationID extracts the correlation ID from an error
func GetCorrelationID(err error) string {
	if infraErr, ok := err.(*InfrastructureError); ok {
		return infraErr.CorrelationID
	}
	return ""
}