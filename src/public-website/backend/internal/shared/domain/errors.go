package domain

import (
	"errors"
	"fmt"
)

// ErrorType represents the classification of domain errors for consistent error handling
// across all backend services. Each type corresponds to specific HTTP status codes
// and error handling strategies.
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeNotFound      ErrorType = "not_found" 
	ErrorTypeConflict      ErrorType = "conflict"
	ErrorTypeUnauthorized  ErrorType = "unauthorized"
	ErrorTypeForbidden     ErrorType = "forbidden"
	ErrorTypeInternal      ErrorType = "internal"
	ErrorTypeTimeout       ErrorType = "timeout"
	ErrorTypeRateLimit     ErrorType = "rate_limit"
	ErrorTypeDependency    ErrorType = "dependency"
)

// DomainError represents a structured domain error with comprehensive context
// for debugging, logging, and API responses. It implements the error interface
// and provides structured error information for consistent error handling.
//
// The error includes type classification, structured codes, human-readable messages,
// field-specific validation context, and error chaining support.
type DomainError struct {
	Type    ErrorType `json:"type"`              // Error classification for handling strategy
	Code    string    `json:"code"`              // Structured error code for programmatic handling
	Message string    `json:"message"`           // Human-readable error message
	Field   string    `json:"field,omitempty"`   // Field name for validation errors
	Value   string    `json:"value,omitempty"`   // Field value that caused the error
	Cause   error     `json:"-"`                 // Underlying cause for error chaining
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s)", e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// Is implements error matching for errors.Is
func (e *DomainError) Is(target error) bool {
	if other, ok := target.(*DomainError); ok {
		return e.Type == other.Type && e.Code == other.Code
	}
	return false
}

// NewDomainError creates a new domain error
func NewDomainError(errorType ErrorType, code, message string) *DomainError {
	return &DomainError{
		Type:    errorType,
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeValidation,
		Code:    "VALIDATION_ERROR",
		Message: message,
	}
}

// NewValidationFieldError creates a validation error for a specific field
func NewValidationFieldError(field, message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeValidation,
		Code:    "VALIDATION_FIELD_ERROR", 
		Message: message,
		Field:   field,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(entityType, entityID string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeNotFound,
		Code:    "ENTITY_NOT_FOUND",
		Message: fmt.Sprintf("%s with ID %s not found", entityType, entityID),
		Value:   entityID,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeConflict,
		Code:    "CONFLICT",
		Message: message,
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeUnauthorized,
		Code:    "UNAUTHORIZED",
		Message: message,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeForbidden,
		Code:    "FORBIDDEN",
		Message: message,
	}
}

// NewInternalError creates an internal error with a cause
func NewInternalError(message string, cause error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeInternal,
		Code:    "INTERNAL_ERROR",
		Message: message,
		Cause:   cause,
	}
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeTimeout,
		Code:    "TIMEOUT",
		Message: fmt.Sprintf("timeout during %s", operation),
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(limit string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeRateLimit,
		Code:    "RATE_LIMIT_EXCEEDED",
		Message: fmt.Sprintf("rate limit exceeded: %s", limit),
	}
}

// NewDependencyError creates a dependency error
func NewDependencyError(dependency string, cause error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeDependency,
		Code:    "DEPENDENCY_ERROR",
		Message: fmt.Sprintf("dependency error: %s", dependency),
		Cause:   cause,
	}
}

// Predefined common validation errors for consistency
var (
	// Format validation errors
	ErrInvalidUUID  = NewValidationError("invalid UUID format")
	ErrInvalidSlug  = NewValidationError("invalid slug format")
	ErrInvalidEmail = NewValidationError("invalid email format")
	ErrInvalidURL   = NewValidationError("invalid URL format")
	ErrInvalidDate  = NewValidationError("invalid date format")
	ErrInvalidEnum  = NewValidationError("invalid enum value")
	
	// Required field validation errors
	ErrEmptyTitle           = NewValidationError("title is required")
	ErrEmptyDescription     = NewValidationError("description is required")
	ErrMissingCorrelationID = NewValidationError("correlation ID is required")
	ErrMissingUserID        = NewValidationError("user ID is required")
)

// IsValidationError checks if error is a validation error
func IsValidationError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr) && domainErr.Type == ErrorTypeValidation
}

// IsNotFoundError checks if error is a not found error
func IsNotFoundError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr) && domainErr.Type == ErrorTypeNotFound
}

// IsConflictError checks if error is a conflict error
func IsConflictError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr) && domainErr.Type == ErrorTypeConflict
}

// IsUnauthorizedError checks if error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr) && domainErr.Type == ErrorTypeUnauthorized
}

// IsForbiddenError checks if error is a forbidden error
func IsForbiddenError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr) && domainErr.Type == ErrorTypeForbidden
}

// IsInternalError checks if error is an internal error
func IsInternalError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr) && domainErr.Type == ErrorTypeInternal
}

// IsTimeoutError checks if error is a timeout error
func IsTimeoutError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr) && domainErr.Type == ErrorTypeTimeout
}

// IsRateLimitError checks if error is a rate limit error
func IsRateLimitError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr) && domainErr.Type == ErrorTypeRateLimit
}

// IsDependencyError checks if error is a dependency error
func IsDependencyError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr) && domainErr.Type == ErrorTypeDependency
}

// GetErrorType extracts the error type from a domain error
func GetErrorType(err error) ErrorType {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Type
	}
	return ErrorTypeInternal
}

// GetErrorCode extracts the error code from a domain error
func GetErrorCode(err error) string {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code
	}
	return "UNKNOWN_ERROR"
}

// WrapError wraps an error as an internal domain error
func WrapError(err error, message string) *DomainError {
	if err == nil {
		return nil
	}
	
	// If it's already a domain error, preserve the type
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return &DomainError{
			Type:    domainErr.Type,
			Code:    domainErr.Code,
			Message: message,
			Cause:   err,
		}
	}
	
	return NewInternalError(message, err)
}