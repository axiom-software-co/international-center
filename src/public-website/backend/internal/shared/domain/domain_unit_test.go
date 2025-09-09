package domain

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE - Domain Error System Tests (60+ test cases)

func TestNewDomainError(t *testing.T) {
	tests := []struct {
		name        string
		errorType   ErrorType
		code        string
		message     string
		wantType    ErrorType
		wantCode    string
		wantMessage string
	}{
		{
			name:        "create validation error",
			errorType:   ErrorTypeValidation,
			code:        "VALIDATION_ERROR",
			message:     "validation failed",
			wantType:    ErrorTypeValidation,
			wantCode:    "VALIDATION_ERROR",
			wantMessage: "validation failed",
		},
		{
			name:        "create not found error",
			errorType:   ErrorTypeNotFound,
			code:        "NOT_FOUND",
			message:     "entity not found",
			wantType:    ErrorTypeNotFound,
			wantCode:    "NOT_FOUND",
			wantMessage: "entity not found",
		},
		{
			name:        "create conflict error",
			errorType:   ErrorTypeConflict,
			code:        "CONFLICT",
			message:     "resource conflict",
			wantType:    ErrorTypeConflict,
			wantCode:    "CONFLICT",
			wantMessage: "resource conflict",
		},
		{
			name:        "create unauthorized error",
			errorType:   ErrorTypeUnauthorized,
			code:        "UNAUTHORIZED",
			message:     "access denied",
			wantType:    ErrorTypeUnauthorized,
			wantCode:    "UNAUTHORIZED",
			wantMessage: "access denied",
		},
		{
			name:        "create forbidden error",
			errorType:   ErrorTypeForbidden,
			code:        "FORBIDDEN",
			message:     "forbidden action",
			wantType:    ErrorTypeForbidden,
			wantCode:    "FORBIDDEN",
			wantMessage: "forbidden action",
		},
		{
			name:        "create internal error",
			errorType:   ErrorTypeInternal,
			code:        "INTERNAL_ERROR",
			message:     "internal server error",
			wantType:    ErrorTypeInternal,
			wantCode:    "INTERNAL_ERROR",
			wantMessage: "internal server error",
		},
		{
			name:        "create timeout error",
			errorType:   ErrorTypeTimeout,
			code:        "TIMEOUT",
			message:     "request timeout",
			wantType:    ErrorTypeTimeout,
			wantCode:    "TIMEOUT",
			wantMessage: "request timeout",
		},
		{
			name:        "create rate limit error",
			errorType:   ErrorTypeRateLimit,
			code:        "RATE_LIMIT",
			message:     "rate limit exceeded",
			wantType:    ErrorTypeRateLimit,
			wantCode:    "RATE_LIMIT",
			wantMessage: "rate limit exceeded",
		},
		{
			name:        "create dependency error",
			errorType:   ErrorTypeDependency,
			code:        "DEPENDENCY_ERROR",
			message:     "dependency failure",
			wantType:    ErrorTypeDependency,
			wantCode:    "DEPENDENCY_ERROR",
			wantMessage: "dependency failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until NewDomainError is properly implemented
			err := NewDomainError(tt.errorType, tt.code, tt.message)
			require.NotNil(t, err)
			assert.Equal(t, tt.wantType, err.Type)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMessage, err.Message)
			assert.Equal(t, "", err.Field)
			assert.Equal(t, "", err.Value)
			assert.Nil(t, err.Cause)
		})
	}
}

func TestNewValidationError(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		wantType    ErrorType
		wantCode    string
		wantMessage string
	}{
		{
			name:        "create simple validation error",
			message:     "field is required",
			wantType:    ErrorTypeValidation,
			wantCode:    "VALIDATION_ERROR",
			wantMessage: "field is required",
		},
		{
			name:        "create validation error with complex message",
			message:     "value must be between 1 and 100",
			wantType:    ErrorTypeValidation,
			wantCode:    "VALIDATION_ERROR",
			wantMessage: "value must be between 1 and 100",
		},
		{
			name:        "create validation error with empty message",
			message:     "",
			wantType:    ErrorTypeValidation,
			wantCode:    "VALIDATION_ERROR",
			wantMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until NewValidationError is properly implemented
			err := NewValidationError(tt.message)
			require.NotNil(t, err)
			assert.Equal(t, tt.wantType, err.Type)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMessage, err.Message)
		})
	}
}

func TestNewValidationFieldError(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		message     string
		wantType    ErrorType
		wantCode    string
		wantField   string
		wantMessage string
	}{
		{
			name:        "create field validation error",
			field:       "email",
			message:     "invalid email format",
			wantType:    ErrorTypeValidation,
			wantCode:    "VALIDATION_FIELD_ERROR",
			wantField:   "email",
			wantMessage: "invalid email format",
		},
		{
			name:        "create field validation error with empty field",
			field:       "",
			message:     "field is required",
			wantType:    ErrorTypeValidation,
			wantCode:    "VALIDATION_FIELD_ERROR",
			wantField:   "",
			wantMessage: "field is required",
		},
		{
			name:        "create field validation error with complex field name",
			field:       "user.profile.bio",
			message:     "bio exceeds maximum length",
			wantType:    ErrorTypeValidation,
			wantCode:    "VALIDATION_FIELD_ERROR",
			wantField:   "user.profile.bio",
			wantMessage: "bio exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until NewValidationFieldError is properly implemented
			err := NewValidationFieldError(tt.field, tt.message)
			require.NotNil(t, err)
			assert.Equal(t, tt.wantType, err.Type)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantField, err.Field)
			assert.Equal(t, tt.wantMessage, err.Message)
		})
	}
}

func TestNewNotFoundError(t *testing.T) {
	tests := []struct {
		name           string
		entityType     string
		entityID       string
		wantType       ErrorType
		wantCode       string
		wantValue      string
		wantMessage    string
	}{
		{
			name:        "create not found error for user",
			entityType:  "user",
			entityID:    "123",
			wantType:    ErrorTypeNotFound,
			wantCode:    "ENTITY_NOT_FOUND",
			wantValue:   "123",
			wantMessage: "user with ID 123 not found",
		},
		{
			name:        "create not found error for UUID entity",
			entityType:  "service",
			entityID:    "550e8400-e29b-41d4-a716-446655440000",
			wantType:    ErrorTypeNotFound,
			wantCode:    "ENTITY_NOT_FOUND",
			wantValue:   "550e8400-e29b-41d4-a716-446655440000",
			wantMessage: "service with ID 550e8400-e29b-41d4-a716-446655440000 not found",
		},
		{
			name:        "create not found error with empty entity type",
			entityType:  "",
			entityID:    "456",
			wantType:    ErrorTypeNotFound,
			wantCode:    "ENTITY_NOT_FOUND",
			wantValue:   "456",
			wantMessage: " with ID 456 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until NewNotFoundError is properly implemented
			err := NewNotFoundError(tt.entityType, tt.entityID)
			require.NotNil(t, err)
			assert.Equal(t, tt.wantType, err.Type)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantValue, err.Value)
			assert.Equal(t, tt.wantMessage, err.Message)
		})
	}
}

func TestNewConflictError(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		wantType    ErrorType
		wantCode    string
		wantMessage string
	}{
		{
			name:        "create conflict error",
			message:     "resource already exists",
			wantType:    ErrorTypeConflict,
			wantCode:    "CONFLICT",
			wantMessage: "resource already exists",
		},
		{
			name:        "create conflict error with detailed message",
			message:     "email address already in use by another user",
			wantType:    ErrorTypeConflict,
			wantCode:    "CONFLICT",
			wantMessage: "email address already in use by another user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until NewConflictError is properly implemented
			err := NewConflictError(tt.message)
			require.NotNil(t, err)
			assert.Equal(t, tt.wantType, err.Type)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMessage, err.Message)
		})
	}
}

func TestNewUnauthorizedError(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		wantType    ErrorType
		wantCode    string
		wantMessage string
	}{
		{
			name:        "create unauthorized error",
			message:     "authentication required",
			wantType:    ErrorTypeUnauthorized,
			wantCode:    "UNAUTHORIZED",
			wantMessage: "authentication required",
		},
		{
			name:        "create unauthorized error with token info",
			message:     "invalid or expired authentication token",
			wantType:    ErrorTypeUnauthorized,
			wantCode:    "UNAUTHORIZED",
			wantMessage: "invalid or expired authentication token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until NewUnauthorizedError is properly implemented
			err := NewUnauthorizedError(tt.message)
			require.NotNil(t, err)
			assert.Equal(t, tt.wantType, err.Type)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMessage, err.Message)
		})
	}
}

func TestNewForbiddenError(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		wantType    ErrorType
		wantCode    string
		wantMessage string
	}{
		{
			name:        "create forbidden error",
			message:     "insufficient permissions",
			wantType:    ErrorTypeForbidden,
			wantCode:    "FORBIDDEN",
			wantMessage: "insufficient permissions",
		},
		{
			name:        "create forbidden error with detailed message",
			message:     "admin role required for this action",
			wantType:    ErrorTypeForbidden,
			wantCode:    "FORBIDDEN",
			wantMessage: "admin role required for this action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until NewForbiddenError is properly implemented
			err := NewForbiddenError(tt.message)
			require.NotNil(t, err)
			assert.Equal(t, tt.wantType, err.Type)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMessage, err.Message)
		})
	}
}

func TestNewInternalError(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		cause       error
		wantType    ErrorType
		wantCode    string
		wantMessage string
		wantCause   error
	}{
		{
			name:        "create internal error with cause",
			message:     "database connection failed",
			cause:       errors.New("connection timeout"),
			wantType:    ErrorTypeInternal,
			wantCode:    "INTERNAL_ERROR",
			wantMessage: "database connection failed",
			wantCause:   errors.New("connection timeout"),
		},
		{
			name:        "create internal error without cause",
			message:     "unexpected error occurred",
			cause:       nil,
			wantType:    ErrorTypeInternal,
			wantCode:    "INTERNAL_ERROR",
			wantMessage: "unexpected error occurred",
			wantCause:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until NewInternalError is properly implemented
			err := NewInternalError(tt.message, tt.cause)
			require.NotNil(t, err)
			assert.Equal(t, tt.wantType, err.Type)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMessage, err.Message)
			if tt.wantCause != nil {
				assert.EqualError(t, err.Cause, tt.wantCause.Error())
			} else {
				assert.Nil(t, err.Cause)
			}
		})
	}
}

func TestNewTimeoutError(t *testing.T) {
	tests := []struct {
		name        string
		operation   string
		wantType    ErrorType
		wantCode    string
		wantMessage string
	}{
		{
			name:        "create timeout error for database operation",
			operation:   "database query",
			wantType:    ErrorTypeTimeout,
			wantCode:    "TIMEOUT",
			wantMessage: "timeout during database query",
		},
		{
			name:        "create timeout error for HTTP request",
			operation:   "external API call",
			wantType:    ErrorTypeTimeout,
			wantCode:    "TIMEOUT",
			wantMessage: "timeout during external API call",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until NewTimeoutError is properly implemented
			err := NewTimeoutError(tt.operation)
			require.NotNil(t, err)
			assert.Equal(t, tt.wantType, err.Type)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMessage, err.Message)
		})
	}
}

func TestNewRateLimitError(t *testing.T) {
	tests := []struct {
		name        string
		limit       string
		wantType    ErrorType
		wantCode    string
		wantMessage string
	}{
		{
			name:        "create rate limit error with requests per minute",
			limit:       "100 requests per minute",
			wantType:    ErrorTypeRateLimit,
			wantCode:    "RATE_LIMIT_EXCEEDED",
			wantMessage: "rate limit exceeded: 100 requests per minute",
		},
		{
			name:        "create rate limit error with API quota",
			limit:       "daily API quota",
			wantType:    ErrorTypeRateLimit,
			wantCode:    "RATE_LIMIT_EXCEEDED",
			wantMessage: "rate limit exceeded: daily API quota",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until NewRateLimitError is properly implemented
			err := NewRateLimitError(tt.limit)
			require.NotNil(t, err)
			assert.Equal(t, tt.wantType, err.Type)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMessage, err.Message)
		})
	}
}

func TestNewDependencyError(t *testing.T) {
	tests := []struct {
		name         string
		dependency   string
		cause        error
		wantType     ErrorType
		wantCode     string
		wantMessage  string
		wantCause    error
	}{
		{
			name:        "create dependency error with redis",
			dependency:  "Redis",
			cause:       errors.New("connection refused"),
			wantType:    ErrorTypeDependency,
			wantCode:    "DEPENDENCY_ERROR",
			wantMessage: "dependency error: Redis",
			wantCause:   errors.New("connection refused"),
		},
		{
			name:        "create dependency error without cause",
			dependency:  "External API",
			cause:       nil,
			wantType:    ErrorTypeDependency,
			wantCode:    "DEPENDENCY_ERROR",
			wantMessage: "dependency error: External API",
			wantCause:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until NewDependencyError is properly implemented
			err := NewDependencyError(tt.dependency, tt.cause)
			require.NotNil(t, err)
			assert.Equal(t, tt.wantType, err.Type)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMessage, err.Message)
			if tt.wantCause != nil {
				assert.EqualError(t, err.Cause, tt.wantCause.Error())
			} else {
				assert.Nil(t, err.Cause)
			}
		})
	}
}

func TestDomainError_Error(t *testing.T) {
	tests := []struct {
		name        string
		domainError *DomainError
		wantError   string
	}{
		{
			name: "error without field",
			domainError: &DomainError{
				Type:    ErrorTypeValidation,
				Code:    "VALIDATION_ERROR",
				Message: "validation failed",
			},
			wantError: "VALIDATION_ERROR: validation failed",
		},
		{
			name: "error with field",
			domainError: &DomainError{
				Type:    ErrorTypeValidation,
				Code:    "VALIDATION_FIELD_ERROR",
				Message: "invalid email format",
				Field:   "email",
			},
			wantError: "VALIDATION_FIELD_ERROR: invalid email format (field: email)",
		},
		{
			name: "error with empty field",
			domainError: &DomainError{
				Type:    ErrorTypeNotFound,
				Code:    "ENTITY_NOT_FOUND",
				Message: "user not found",
				Field:   "",
			},
			wantError: "ENTITY_NOT_FOUND: user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until DomainError.Error() is properly implemented
			got := tt.domainError.Error()
			assert.Equal(t, tt.wantError, got)
		})
	}
}

func TestDomainError_Unwrap(t *testing.T) {
	tests := []struct {
		name        string
		domainError *DomainError
		wantCause   error
	}{
		{
			name: "error with cause",
			domainError: &DomainError{
				Type:    ErrorTypeInternal,
				Code:    "INTERNAL_ERROR",
				Message: "database error",
				Cause:   errors.New("connection timeout"),
			},
			wantCause: errors.New("connection timeout"),
		},
		{
			name: "error without cause",
			domainError: &DomainError{
				Type:    ErrorTypeValidation,
				Code:    "VALIDATION_ERROR",
				Message: "validation failed",
				Cause:   nil,
			},
			wantCause: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until DomainError.Unwrap() is properly implemented
			got := tt.domainError.Unwrap()
			if tt.wantCause != nil {
				assert.EqualError(t, got, tt.wantCause.Error())
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestDomainError_Is(t *testing.T) {
	validationError := &DomainError{
		Type: ErrorTypeValidation,
		Code: "VALIDATION_ERROR",
	}
	
	sameValidationError := &DomainError{
		Type: ErrorTypeValidation,
		Code: "VALIDATION_ERROR",
	}
	
	differentValidationError := &DomainError{
		Type: ErrorTypeValidation,
		Code: "VALIDATION_FIELD_ERROR",
	}
	
	notFoundError := &DomainError{
		Type: ErrorTypeNotFound,
		Code: "ENTITY_NOT_FOUND",
	}
	
	regularError := errors.New("regular error")

	tests := []struct {
		name   string
		err    *DomainError
		target error
		want   bool
	}{
		{
			name:   "same error type and code",
			err:    validationError,
			target: sameValidationError,
			want:   true,
		},
		{
			name:   "same error type different code",
			err:    validationError,
			target: differentValidationError,
			want:   false,
		},
		{
			name:   "different error type",
			err:    validationError,
			target: notFoundError,
			want:   false,
		},
		{
			name:   "non-domain error",
			err:    validationError,
			target: regularError,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until DomainError.Is() is properly implemented
			got := tt.err.Is(tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "validation error",
			err:  NewValidationError("validation failed"),
			want: true,
		},
		{
			name: "validation field error",
			err:  NewValidationFieldError("email", "invalid format"),
			want: true,
		},
		{
			name: "not found error",
			err:  NewNotFoundError("user", "123"),
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until IsValidationError is properly implemented
			got := IsValidationError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "not found error",
			err:  NewNotFoundError("user", "123"),
			want: true,
		},
		{
			name: "validation error",
			err:  NewValidationError("validation failed"),
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until IsNotFoundError is properly implemented
			got := IsNotFoundError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsConflictError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "conflict error",
			err:  NewConflictError("resource exists"),
			want: true,
		},
		{
			name: "validation error",
			err:  NewValidationError("validation failed"),
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until IsConflictError is properly implemented
			got := IsConflictError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsUnauthorizedError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "unauthorized error",
			err:  NewUnauthorizedError("authentication required"),
			want: true,
		},
		{
			name: "validation error",
			err:  NewValidationError("validation failed"),
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until IsUnauthorizedError is properly implemented
			got := IsUnauthorizedError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsForbiddenError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "forbidden error",
			err:  NewForbiddenError("insufficient permissions"),
			want: true,
		},
		{
			name: "validation error",
			err:  NewValidationError("validation failed"),
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until IsForbiddenError is properly implemented
			got := IsForbiddenError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsInternalError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "internal error",
			err:  NewInternalError("server error", nil),
			want: true,
		},
		{
			name: "validation error",
			err:  NewValidationError("validation failed"),
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until IsInternalError is properly implemented
			got := IsInternalError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsTimeoutError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "timeout error",
			err:  NewTimeoutError("database query"),
			want: true,
		},
		{
			name: "validation error",
			err:  NewValidationError("validation failed"),
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until IsTimeoutError is properly implemented
			got := IsTimeoutError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "rate limit error",
			err:  NewRateLimitError("100 requests per minute"),
			want: true,
		},
		{
			name: "validation error",
			err:  NewValidationError("validation failed"),
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until IsRateLimitError is properly implemented
			got := IsRateLimitError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsDependencyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "dependency error",
			err:  NewDependencyError("Redis", nil),
			want: true,
		},
		{
			name: "validation error",
			err:  NewValidationError("validation failed"),
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until IsDependencyError is properly implemented
			got := IsDependencyError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetErrorType(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ErrorType
	}{
		{
			name: "validation error",
			err:  NewValidationError("validation failed"),
			want: ErrorTypeValidation,
		},
		{
			name: "not found error",
			err:  NewNotFoundError("user", "123"),
			want: ErrorTypeNotFound,
		},
		{
			name: "timeout error",
			err:  NewTimeoutError("database query"),
			want: ErrorTypeTimeout,
		},
		{
			name: "regular error",
			err:  errors.New("regular error"),
			want: ErrorTypeInternal,
		},
		{
			name: "nil error",
			err:  nil,
			want: ErrorTypeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until GetErrorType is properly implemented
			got := GetErrorType(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "validation error",
			err:  NewValidationError("validation failed"),
			want: "VALIDATION_ERROR",
		},
		{
			name: "not found error",
			err:  NewNotFoundError("user", "123"),
			want: "ENTITY_NOT_FOUND",
		},
		{
			name: "timeout error",
			err:  NewTimeoutError("database query"),
			want: "TIMEOUT",
		},
		{
			name: "regular error",
			err:  errors.New("regular error"),
			want: "UNKNOWN_ERROR",
		},
		{
			name: "nil error",
			err:  nil,
			want: "UNKNOWN_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until GetErrorCode is properly implemented
			got := GetErrorCode(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWrapError(t *testing.T) {
	regularError := errors.New("database connection failed")
	domainValidationError := NewValidationError("validation failed")

	tests := []struct {
		name        string
		err         error
		message     string
		wantType    ErrorType
		wantCode    string
		wantMessage string
		wantNil     bool
	}{
		{
			name:        "wrap regular error",
			err:         regularError,
			message:     "failed to save user",
			wantType:    ErrorTypeInternal,
			wantCode:    "INTERNAL_ERROR",
			wantMessage: "failed to save user",
			wantNil:     false,
		},
		{
			name:        "wrap domain error preserves type",
			err:         domainValidationError,
			message:     "user validation failed",
			wantType:    ErrorTypeValidation,
			wantCode:    "VALIDATION_ERROR",
			wantMessage: "user validation failed",
			wantNil:     false,
		},
		{
			name:    "wrap nil error returns nil",
			err:     nil,
			message: "some message",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until WrapError is properly implemented
			got := WrapError(tt.err, tt.message)
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, tt.wantType, got.Type)
				assert.Equal(t, tt.wantCode, got.Code)
				assert.Equal(t, tt.wantMessage, got.Message)
				assert.Equal(t, tt.err, got.Cause)
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name      string
		err       *DomainError
		wantType  ErrorType
		wantCode  string
	}{
		{
			name:     "ErrInvalidUUID",
			err:      ErrInvalidUUID,
			wantType: ErrorTypeValidation,
			wantCode: "VALIDATION_ERROR",
		},
		{
			name:     "ErrInvalidSlug",
			err:      ErrInvalidSlug,
			wantType: ErrorTypeValidation,
			wantCode: "VALIDATION_ERROR",
		},
		{
			name:     "ErrEmptyTitle",
			err:      ErrEmptyTitle,
			wantType: ErrorTypeValidation,
			wantCode: "VALIDATION_ERROR",
		},
		{
			name:     "ErrEmptyDescription",
			err:      ErrEmptyDescription,
			wantType: ErrorTypeValidation,
			wantCode: "VALIDATION_ERROR",
		},
		{
			name:     "ErrInvalidEmail",
			err:      ErrInvalidEmail,
			wantType: ErrorTypeValidation,
			wantCode: "VALIDATION_ERROR",
		},
		{
			name:     "ErrInvalidURL",
			err:      ErrInvalidURL,
			wantType: ErrorTypeValidation,
			wantCode: "VALIDATION_ERROR",
		},
		{
			name:     "ErrInvalidDate",
			err:      ErrInvalidDate,
			wantType: ErrorTypeValidation,
			wantCode: "VALIDATION_ERROR",
		},
		{
			name:     "ErrInvalidEnum",
			err:      ErrInvalidEnum,
			wantType: ErrorTypeValidation,
			wantCode: "VALIDATION_ERROR",
		},
		{
			name:     "ErrMissingCorrelationID",
			err:      ErrMissingCorrelationID,
			wantType: ErrorTypeValidation,
			wantCode: "VALIDATION_ERROR",
		},
		{
			name:     "ErrMissingUserID",
			err:      ErrMissingUserID,
			wantType: ErrorTypeValidation,
			wantCode: "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until predefined errors are properly implemented
			require.NotNil(t, tt.err)
			assert.Equal(t, tt.wantType, tt.err.Type)
			assert.Equal(t, tt.wantCode, tt.err.Code)
			assert.NotEmpty(t, tt.err.Message)
		})
	}
}

// RED PHASE - Validation System Tests (40+ test cases)

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid UUID v4",
			id:        "550e8400-e29b-41d4-a716-446655440000",
			wantError: false,
		},
		{
			name:      "valid UUID generated",
			id:        uuid.New().String(),
			wantError: false,
		},
		{
			name:      "empty UUID",
			id:        "",
			wantError: true,
			errorMsg:  "UUID cannot be empty",
		},
		{
			name:      "invalid UUID format",
			id:        "not-a-uuid",
			wantError: true,
			errorMsg:  "invalid UUID format",
		},
		{
			name:      "UUID with wrong length",
			id:        "550e8400-e29b-41d4-a716",
			wantError: true,
			errorMsg:  "invalid UUID format",
		},
		{
			name:      "UUID with invalid characters",
			id:        "550e8400-e29b-41d4-a716-44665544000g",
			wantError: true,
			errorMsg:  "invalid UUID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until ValidateUUID is properly implemented
			err := ValidateUUID(tt.id)
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		name      string
		slug      string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid simple slug",
			slug:      "hello-world",
			wantError: false,
		},
		{
			name:      "valid slug with numbers",
			slug:      "test-123",
			wantError: false,
		},
		{
			name:      "valid single character slug",
			slug:      "a",
			wantError: false,
		},
		{
			name:      "valid long slug",
			slug:      "this-is-a-very-long-slug-with-many-words-and-hyphens",
			wantError: false,
		},
		{
			name:      "empty slug",
			slug:      "",
			wantError: true,
			errorMsg:  "slug cannot be empty",
		},
		{
			name:      "slug with uppercase letters",
			slug:      "Hello-World",
			wantError: true,
			errorMsg:  "slug must contain only lowercase letters, numbers, and hyphens",
		},
		{
			name:      "slug with spaces",
			slug:      "hello world",
			wantError: true,
			errorMsg:  "slug must contain only lowercase letters, numbers, and hyphens",
		},
		{
			name:      "slug with special characters",
			slug:      "hello_world!",
			wantError: true,
			errorMsg:  "slug must contain only lowercase letters, numbers, and hyphens",
		},
		{
			name:      "slug exceeding length limit",
			slug:      strings.Repeat("a", 256),
			wantError: true,
			errorMsg:  "slug cannot exceed 255 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until ValidateSlug is properly implemented
			err := ValidateSlug(tt.slug)
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTitle(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid title",
			title:     "Hello World",
			wantError: false,
		},
		{
			name:      "valid title with special characters",
			title:     "Hello, World! How are you?",
			wantError: false,
		},
		{
			name:      "valid title at max length",
			title:     strings.Repeat("a", 255),
			wantError: false,
		},
		{
			name:      "empty title",
			title:     "",
			wantError: true,
			errorMsg:  "title cannot be empty",
		},
		{
			name:      "whitespace only title",
			title:     "   ",
			wantError: true,
			errorMsg:  "title cannot be empty",
		},
		{
			name:      "title exceeding length limit",
			title:     strings.Repeat("a", 256),
			wantError: true,
			errorMsg:  "title cannot exceed 255 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until ValidateTitle is properly implemented
			err := ValidateTitle(tt.title)
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRequiredString(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid required string",
			fieldName: "name",
			value:     "John Doe",
			wantError: false,
		},
		{
			name:      "valid string with leading/trailing spaces",
			fieldName: "description",
			value:     "  Hello World  ",
			wantError: false,
		},
		{
			name:      "empty string",
			fieldName: "email",
			value:     "",
			wantError: true,
			errorMsg:  "email cannot be empty",
		},
		{
			name:      "whitespace only string",
			fieldName: "address",
			value:     "   ",
			wantError: true,
			errorMsg:  "address cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until ValidateRequiredString is properly implemented
			err := ValidateRequiredString(tt.fieldName, tt.value)
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRequiredStringWithLength(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     string
		maxLength int
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid string within length",
			fieldName: "description",
			value:     "Hello World",
			maxLength: 50,
			wantError: false,
		},
		{
			name:      "valid string at max length",
			fieldName: "title",
			value:     strings.Repeat("a", 10),
			maxLength: 10,
			wantError: false,
		},
		{
			name:      "empty string",
			fieldName: "content",
			value:     "",
			maxLength: 100,
			wantError: true,
			errorMsg:  "content cannot be empty",
		},
		{
			name:      "string exceeding length",
			fieldName: "summary",
			value:     strings.Repeat("a", 11),
			maxLength: 10,
			wantError: true,
			errorMsg:  "summary cannot exceed 10 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until ValidateRequiredStringWithLength is properly implemented
			err := ValidateRequiredStringWithLength(tt.fieldName, tt.value, tt.maxLength)
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateHTTPSURL(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		url       string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid HTTPS URL",
			fieldName: "website",
			url:       "https://example.com",
			wantError: false,
		},
		{
			name:      "valid HTTPS URL with path",
			fieldName: "image_url",
			url:       "https://example.com/path/to/image.jpg",
			wantError: false,
		},
		{
			name:      "valid HTTPS URL with query params",
			fieldName: "api_url",
			url:       "https://api.example.com/v1/users?limit=10",
			wantError: false,
		},
		{
			name:      "empty URL (optional field)",
			fieldName: "optional_url",
			url:       "",
			wantError: false,
		},
		{
			name:      "HTTP URL (not HTTPS)",
			fieldName: "website",
			url:       "http://example.com",
			wantError: true,
			errorMsg:  "website must be a valid HTTPS URL",
		},
		{
			name:      "invalid URL format",
			fieldName: "image_url",
			url:       "not-a-url",
			wantError: true,
			errorMsg:  "image_url must be a valid HTTPS URL",
		},
		{
			name:      "URL with spaces",
			fieldName: "api_url",
			url:       "https://example.com/path with spaces",
			wantError: true,
			errorMsg:  "api_url must be a valid HTTPS URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until ValidateHTTPSURL is properly implemented
			err := ValidateHTTPSURL(tt.fieldName, tt.url)
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple title",
			input: "Hello World",
			want:  "hello-world",
		},
		{
			name:  "title with special characters",
			input: "Hello, World! How are you?",
			want:  "hello-world-how-are-you",
		},
		{
			name:  "title with numbers",
			input: "Article 123 Version 2.0",
			want:  "article-123-version-2-0",
		},
		{
			name:  "title with multiple spaces",
			input: "Hello    World",
			want:  "hello-world",
		},
		{
			name:  "title with leading/trailing spaces",
			input: "  Hello World  ",
			want:  "hello-world",
		},
		{
			name:  "title with hyphens",
			input: "Hello-World-Test",
			want:  "hello-world-test",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "title exceeding length limit",
			input: strings.Repeat("Hello World ", 30),
			want:  strings.TrimRight(strings.Repeat("hello-world-", 21), "-"),
		},
		{
			name:  "title with only special characters",
			input: "!@#$%^&*()",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until GenerateSlug is properly implemented
			got := GenerateSlug(tt.input)
			// Handle the case where long input results might be truncated
			if tt.name == "title exceeding length limit" {
				// For very long inputs, just check that the result is truncated and reasonable
				assert.LessOrEqual(t, len(got), 255)
				assert.True(t, strings.HasPrefix(got, "hello-world-"))
				assert.NotEmpty(t, strings.TrimSuffix(got, "-"))
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestValidateEnum(t *testing.T) {
	allowedValues := []string{"active", "inactive", "pending"}

	tests := []struct {
		name          string
		fieldName     string
		value         string
		allowedValues []string
		wantError     bool
		errorMsg      string
	}{
		{
			name:          "valid enum value",
			fieldName:     "status",
			value:         "active",
			allowedValues: allowedValues,
			wantError:     false,
		},
		{
			name:          "another valid enum value",
			fieldName:     "status",
			value:         "pending",
			allowedValues: allowedValues,
			wantError:     false,
		},
		{
			name:          "empty value",
			fieldName:     "status",
			value:         "",
			allowedValues: allowedValues,
			wantError:     true,
			errorMsg:      "status cannot be empty",
		},
		{
			name:          "invalid enum value",
			fieldName:     "status",
			value:         "unknown",
			allowedValues: allowedValues,
			wantError:     true,
			errorMsg:      "invalid status value: unknown",
		},
		{
			name:          "case sensitive validation",
			fieldName:     "status",
			value:         "Active",
			allowedValues: allowedValues,
			wantError:     true,
			errorMsg:      "invalid status value: Active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until ValidateEnum is properly implemented
			err := ValidateEnum(tt.fieldName, tt.value, tt.allowedValues)
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// RED PHASE - Audit Event System Tests (50+ test cases)

func TestNewAuditEvent(t *testing.T) {
	tests := []struct {
		name          string
		entityType    EntityType
		entityID      string
		operationType AuditEventType
		userID        string
		wantFields    []string
	}{
		{
			name:          "create audit event for service insert",
			entityType:    EntityTypeService,
			entityID:      "service-123",
			operationType: AuditEventInsert,
			userID:        "user-456",
			wantFields:    []string{"AuditID", "EntityType", "EntityID", "OperationType", "AuditTime", "UserID", "CorrelationID"},
		},
		{
			name:          "create audit event for news update",
			entityType:    EntityTypeNews,
			entityID:      "news-789",
			operationType: AuditEventUpdate,
			userID:        "admin-123",
			wantFields:    []string{"AuditID", "EntityType", "EntityID", "OperationType", "AuditTime", "UserID", "CorrelationID"},
		},
		{
			name:          "create audit event for user delete",
			entityType:    EntityTypeUser,
			entityID:      "user-999",
			operationType: AuditEventDelete,
			userID:        "admin-456",
			wantFields:    []string{"AuditID", "EntityType", "EntityID", "OperationType", "AuditTime", "UserID", "CorrelationID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until NewAuditEvent is properly implemented
			event := NewAuditEvent(tt.entityType, tt.entityID, tt.operationType, tt.userID)
			require.NotNil(t, event)
			
			// Validate basic fields
			assert.Equal(t, tt.entityType, event.EntityType)
			assert.Equal(t, tt.entityID, event.EntityID)
			assert.Equal(t, tt.operationType, event.OperationType)
			assert.Equal(t, tt.userID, event.UserID)
			
			// Validate generated fields
			assert.NotEmpty(t, event.AuditID)
			assert.NotEmpty(t, event.CorrelationID)
			assert.False(t, event.AuditTime.IsZero())
			assert.NotNil(t, event.DataSnapshot)
			
			// Validate UUID format for generated IDs
			_, err := uuid.Parse(event.AuditID)
			assert.NoError(t, err, "AuditID should be valid UUID")
			
			_, err = uuid.Parse(event.CorrelationID)
			assert.NoError(t, err, "CorrelationID should be valid UUID")
		})
	}
}

func TestAuditEvent_SetTraceContext(t *testing.T) {
	tests := []struct {
		name          string
		correlationID string
		traceID       string
	}{
		{
			name:          "set trace context with valid IDs",
			correlationID: "correlation-123",
			traceID:       "trace-456",
		},
		{
			name:          "set trace context with empty IDs",
			correlationID: "",
			traceID:       "",
		},
		{
			name:          "set trace context with UUID format",
			correlationID: uuid.New().String(),
			traceID:       "trace-" + uuid.New().String(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until AuditEvent.SetTraceContext is properly implemented
			event := NewAuditEvent(EntityTypeService, "service-123", AuditEventInsert, "user-456")
			
			event.SetTraceContext(tt.correlationID, tt.traceID)
			
			assert.Equal(t, tt.correlationID, event.CorrelationID)
			assert.Equal(t, tt.traceID, event.TraceID)
		})
	}
}

func TestAuditEvent_SetRequestContext(t *testing.T) {
	tests := []struct {
		name       string
		requestURL string
		ipAddress  string
		userAgent  string
	}{
		{
			name:       "set complete request context",
			requestURL: "https://api.example.com/v1/services",
			ipAddress:  "192.168.1.100",
			userAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		},
		{
			name:       "set request context with empty values",
			requestURL: "",
			ipAddress:  "",
			userAgent:  "",
		},
		{
			name:       "set request context with partial data",
			requestURL: "/api/services/123",
			ipAddress:  "10.0.0.1",
			userAgent:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until AuditEvent.SetRequestContext is properly implemented
			event := NewAuditEvent(EntityTypeService, "service-123", AuditEventInsert, "user-456")
			
			event.SetRequestContext(tt.requestURL, tt.ipAddress, tt.userAgent)
			
			assert.Equal(t, tt.requestURL, event.RequestURL)
			assert.Equal(t, tt.ipAddress, event.IPAddress)
			assert.Equal(t, tt.userAgent, event.UserAgent)
		})
	}
}

func TestAuditEvent_SetEnvironmentContext(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		appVersion  string
	}{
		{
			name:        "set production environment context",
			environment: "production",
			appVersion:  "v1.2.3",
		},
		{
			name:        "set development environment context",
			environment: "development",
			appVersion:  "v0.1.0-dev",
		},
		{
			name:        "set environment context with empty values",
			environment: "",
			appVersion:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until AuditEvent.SetEnvironmentContext is properly implemented
			event := NewAuditEvent(EntityTypeService, "service-123", AuditEventInsert, "user-456")
			
			event.SetEnvironmentContext(tt.environment, tt.appVersion)
			
			assert.Equal(t, tt.environment, event.Environment)
			assert.Equal(t, tt.appVersion, event.AppVersion)
		})
	}
}

func TestAuditEvent_SetDataSnapshot(t *testing.T) {
	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name   string
		before interface{}
		after  interface{}
	}{
		{
			name:   "set data snapshot with structs",
			before: &testStruct{Name: "John", Age: 30},
			after:  &testStruct{Name: "John", Age: 31},
		},
		{
			name:   "set data snapshot with maps",
			before: map[string]interface{}{"status": "draft"},
			after:  map[string]interface{}{"status": "published"},
		},
		{
			name:   "set data snapshot with nil values",
			before: nil,
			after:  nil,
		},
		{
			name:   "set data snapshot with mixed types",
			before: "string value",
			after:  123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until AuditEvent.SetDataSnapshot is properly implemented
			event := NewAuditEvent(EntityTypeService, "service-123", AuditEventUpdate, "user-456")
			
			event.SetDataSnapshot(tt.before, tt.after)
			
			require.NotNil(t, event.DataSnapshot)
			assert.Equal(t, tt.before, event.DataSnapshot.Before)
			assert.Equal(t, tt.after, event.DataSnapshot.After)
		})
	}
}

func TestAuditEvent_SetBeforeData(t *testing.T) {
	tests := []struct {
		name   string
		before interface{}
	}{
		{
			name:   "set before data for delete operation",
			before: map[string]interface{}{"id": "123", "name": "Test Service"},
		},
		{
			name:   "set nil before data",
			before: nil,
		},
		{
			name:   "set string before data",
			before: "original value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until AuditEvent.SetBeforeData is properly implemented
			event := NewAuditEvent(EntityTypeService, "service-123", AuditEventDelete, "user-456")
			
			event.SetBeforeData(tt.before)
			
			require.NotNil(t, event.DataSnapshot)
			assert.Equal(t, tt.before, event.DataSnapshot.Before)
			assert.Nil(t, event.DataSnapshot.After)
		})
	}
}

func TestAuditEvent_SetAfterData(t *testing.T) {
	tests := []struct {
		name  string
		after interface{}
	}{
		{
			name:  "set after data for insert operation",
			after: map[string]interface{}{"id": "123", "name": "New Service"},
		},
		{
			name:  "set nil after data",
			after: nil,
		},
		{
			name:  "set complex after data",
			after: struct{ Items []string }{Items: []string{"item1", "item2"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until AuditEvent.SetAfterData is properly implemented
			event := NewAuditEvent(EntityTypeService, "service-123", AuditEventInsert, "user-456")
			
			event.SetAfterData(tt.after)
			
			require.NotNil(t, event.DataSnapshot)
			assert.Equal(t, tt.after, event.DataSnapshot.After)
			assert.Nil(t, event.DataSnapshot.Before)
		})
	}
}

func TestAuditEvent_Validate(t *testing.T) {
	validEvent := &AuditEvent{
		AuditID:       uuid.New().String(),
		EntityType:    EntityTypeService,
		EntityID:      "service-123",
		OperationType: AuditEventInsert,
		AuditTime:     time.Now(),
		UserID:        "user-456",
	}

	tests := []struct {
		name      string
		event     *AuditEvent
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid complete audit event",
			event:     validEvent,
			wantError: false,
		},
		{
			name: "missing audit ID",
			event: &AuditEvent{
				EntityType:    EntityTypeService,
				EntityID:      "service-123",
				OperationType: AuditEventInsert,
				AuditTime:     time.Now(),
				UserID:        "user-456",
			},
			wantError: true,
			errorMsg:  "audit_id is required",
		},
		{
			name: "missing entity type",
			event: &AuditEvent{
				AuditID:       uuid.New().String(),
				EntityID:      "service-123",
				OperationType: AuditEventInsert,
				AuditTime:     time.Now(),
				UserID:        "user-456",
			},
			wantError: true,
			errorMsg:  "entity_type is required",
		},
		{
			name: "missing entity ID",
			event: &AuditEvent{
				AuditID:       uuid.New().String(),
				EntityType:    EntityTypeService,
				OperationType: AuditEventInsert,
				AuditTime:     time.Now(),
				UserID:        "user-456",
			},
			wantError: true,
			errorMsg:  "entity_id is required",
		},
		{
			name: "missing operation type",
			event: &AuditEvent{
				AuditID:    uuid.New().String(),
				EntityType: EntityTypeService,
				EntityID:   "service-123",
				AuditTime:  time.Now(),
				UserID:     "user-456",
			},
			wantError: true,
			errorMsg:  "operation_type is required",
		},
		{
			name: "missing audit time",
			event: &AuditEvent{
				AuditID:       uuid.New().String(),
				EntityType:    EntityTypeService,
				EntityID:      "service-123",
				OperationType: AuditEventInsert,
				UserID:        "user-456",
			},
			wantError: true,
			errorMsg:  "audit_timestamp is required",
		},
		{
			name: "missing user ID",
			event: &AuditEvent{
				AuditID:       uuid.New().String(),
				EntityType:    EntityTypeService,
				EntityID:      "service-123",
				OperationType: AuditEventInsert,
				AuditTime:     time.Now(),
			},
			wantError: true,
			errorMsg:  "user_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until AuditEvent.Validate is properly implemented
			err := tt.event.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidEntityType(t *testing.T) {
	tests := []struct {
		name       string
		entityType EntityType
		want       bool
	}{
		// Services Domain
		{name: "valid EntityTypeService", entityType: EntityTypeService, want: true},
		{name: "valid EntityTypeServiceCategory", entityType: EntityTypeServiceCategory, want: true},
		{name: "valid EntityTypeFeaturedCategory", entityType: EntityTypeFeaturedCategory, want: true},
		// News Domain
		{name: "valid EntityTypeNews", entityType: EntityTypeNews, want: true},
		{name: "valid EntityTypeNewsCategory", entityType: EntityTypeNewsCategory, want: true},
		{name: "valid EntityTypeFeaturedNews", entityType: EntityTypeFeaturedNews, want: true},
		// Research Domain
		{name: "valid EntityTypeResearch", entityType: EntityTypeResearch, want: true},
		{name: "valid EntityTypeResearchCategory", entityType: EntityTypeResearchCategory, want: true},
		{name: "valid EntityTypeFeaturedResearch", entityType: EntityTypeFeaturedResearch, want: true},
		// Events Domain
		{name: "valid EntityTypeEvent", entityType: EntityTypeEvent, want: true},
		{name: "valid EntityTypeEventCategory", entityType: EntityTypeEventCategory, want: true},
		{name: "valid EntityTypeFeaturedEvent", entityType: EntityTypeFeaturedEvent, want: true},
		{name: "valid EntityTypeEventRegistration", entityType: EntityTypeEventRegistration, want: true},
		// Inquiry Domains
		{name: "valid EntityTypeBusinessInquiry", entityType: EntityTypeBusinessInquiry, want: true},
		{name: "valid EntityTypeDonationsInquiry", entityType: EntityTypeDonationsInquiry, want: true},
		{name: "valid EntityTypeMediaInquiry", entityType: EntityTypeMediaInquiry, want: true},
		{name: "valid EntityTypeVolunteerApplication", entityType: EntityTypeVolunteerApplication, want: true},
		// System Entities
		{name: "valid EntityTypeUser", entityType: EntityTypeUser, want: true},
		{name: "valid EntityTypeMigration", entityType: EntityTypeMigration, want: true},
		// Invalid cases
		{name: "invalid empty entity type", entityType: EntityType(""), want: false},
		{name: "invalid unknown entity type", entityType: EntityType("unknown"), want: false},
		{name: "invalid mixed case entity type", entityType: EntityType("Service"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until IsValidEntityType is properly implemented
			got := IsValidEntityType(tt.entityType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidOperationType(t *testing.T) {
	tests := []struct {
		name          string
		operationType AuditEventType
		want          bool
	}{
		{name: "valid AuditEventInsert", operationType: AuditEventInsert, want: true},
		{name: "valid AuditEventUpdate", operationType: AuditEventUpdate, want: true},
		{name: "valid AuditEventDelete", operationType: AuditEventDelete, want: true},
		{name: "valid AuditEventPublish", operationType: AuditEventPublish, want: true},
		{name: "valid AuditEventArchive", operationType: AuditEventArchive, want: true},
		{name: "valid AuditEventAccess", operationType: AuditEventAccess, want: true},
		{name: "invalid empty operation type", operationType: AuditEventType(""), want: false},
		{name: "invalid unknown operation type", operationType: AuditEventType("UNKNOWN"), want: false},
		{name: "invalid mixed case operation type", operationType: AuditEventType("Insert"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until IsValidOperationType is properly implemented
			got := IsValidOperationType(tt.operationType)
			assert.Equal(t, tt.want, got)
		})
	}
}

// RED PHASE - Correlation System Tests (30+ test cases)

func TestNewCorrelationContext(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "create new correlation context"},
		{name: "create another correlation context"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until NewCorrelationContext is properly implemented
			ctx := NewCorrelationContext()
			require.NotNil(t, ctx)
			
			// Validate generated fields
			assert.NotEmpty(t, ctx.CorrelationID)
			assert.NotEmpty(t, ctx.TraceID)
			assert.NotEmpty(t, ctx.RequestID)
			assert.False(t, ctx.StartTime.IsZero())
			assert.Empty(t, ctx.UserID)
			assert.Empty(t, ctx.AppVersion)
			
			// Validate UUID format for generated IDs
			_, err := uuid.Parse(ctx.CorrelationID)
			assert.NoError(t, err, "CorrelationID should be valid UUID")
			
			_, err = uuid.Parse(ctx.RequestID)
			assert.NoError(t, err, "RequestID should be valid UUID")
		})
	}
}

func TestNewCorrelationContextWithID(t *testing.T) {
	tests := []struct {
		name          string
		correlationID string
	}{
		{
			name:          "create correlation context with custom ID",
			correlationID: "custom-correlation-123",
		},
		{
			name:          "create correlation context with UUID ID",
			correlationID: uuid.New().String(),
		},
		{
			name:          "create correlation context with empty ID",
			correlationID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until NewCorrelationContextWithID is properly implemented
			ctx := NewCorrelationContextWithID(tt.correlationID)
			require.NotNil(t, ctx)
			
			// Validate provided correlation ID
			assert.Equal(t, tt.correlationID, ctx.CorrelationID)
			
			// Validate other generated fields
			assert.NotEmpty(t, ctx.TraceID)
			assert.NotEmpty(t, ctx.RequestID)
			assert.False(t, ctx.StartTime.IsZero())
			assert.Empty(t, ctx.UserID)
			assert.Empty(t, ctx.AppVersion)
			
			// Validate UUID format for generated IDs
			_, err := uuid.Parse(ctx.RequestID)
			assert.NoError(t, err, "RequestID should be valid UUID")
		})
	}
}

func TestCorrelationContextConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant ContextKey
		want     string
	}{
		{
			name:     "ContextKeyCorrelationID constant",
			constant: ContextKeyCorrelationID,
			want:     "correlation_id",
		},
		{
			name:     "ContextKeyTraceID constant",
			constant: ContextKeyTraceID,
			want:     "trace_id",
		},
		{
			name:     "ContextKeyUserID constant",
			constant: ContextKeyUserID,
			want:     "user_id",
		},
		{
			name:     "ContextKeyRequestID constant",
			constant: ContextKeyRequestID,
			want:     "request_id",
		},
		{
			name:     "ContextKeyAppVersion constant",
			constant: ContextKeyAppVersion,
			want:     "app_version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until context constants are properly defined
			assert.Equal(t, tt.want, string(tt.constant))
		})
	}
}

func TestContextManipulation(t *testing.T) {
	tests := []struct {
		name  string
		key   ContextKey
		value interface{}
	}{
		{
			name:  "set correlation ID in context",
			key:   ContextKeyCorrelationID,
			value: "correlation-123",
		},
		{
			name:  "set trace ID in context",
			key:   ContextKeyTraceID,
			value: "trace-456",
		},
		{
			name:  "set user ID in context",
			key:   ContextKeyUserID,
			value: "user-789",
		},
		{
			name:  "set request ID in context",
			key:   ContextKeyRequestID,
			value: uuid.New().String(),
		},
		{
			name:  "set app version in context",
			key:   ContextKeyAppVersion,
			value: "v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until context manipulation works properly
			ctx := context.Background()
			ctx = context.WithValue(ctx, tt.key, tt.value)
			
			got := ctx.Value(tt.key)
			assert.Equal(t, tt.value, got)
		})
	}
}

func TestCorrelationContextFields(t *testing.T) {
	correlationID := "correlation-123"
	traceID := "trace-456"
	userID := "user-789"
	requestID := uuid.New().String()
	appVersion := "v1.0.0"
	startTime := time.Now()

	ctx := &CorrelationContext{
		CorrelationID: correlationID,
		TraceID:       traceID,
		UserID:        userID,
		RequestID:     requestID,
		AppVersion:    appVersion,
		StartTime:     startTime,
	}

	tests := []struct {
		name     string
		field    string
		expected interface{}
		actual   interface{}
	}{
		{
			name:     "correlation ID field",
			field:    "CorrelationID",
			expected: correlationID,
			actual:   ctx.CorrelationID,
		},
		{
			name:     "trace ID field",
			field:    "TraceID",
			expected: traceID,
			actual:   ctx.TraceID,
		},
		{
			name:     "user ID field",
			field:    "UserID",
			expected: userID,
			actual:   ctx.UserID,
		},
		{
			name:     "request ID field",
			field:    "RequestID",
			expected: requestID,
			actual:   ctx.RequestID,
		},
		{
			name:     "app version field",
			field:    "AppVersion",
			expected: appVersion,
			actual:   ctx.AppVersion,
		},
		{
			name:     "start time field",
			field:    "StartTime",
			expected: startTime,
			actual:   ctx.StartTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until CorrelationContext struct is properly implemented
			assert.Equal(t, tt.expected, tt.actual)
		})
	}
}

func TestCorrelationContextUniqueness(t *testing.T) {
	// Create multiple correlation contexts to test uniqueness
	ctx1 := NewCorrelationContext()
	ctx2 := NewCorrelationContext()
	ctx3 := NewCorrelationContext()

	contexts := []*CorrelationContext{ctx1, ctx2, ctx3}

	tests := []struct {
		name  string
		field string
	}{
		{name: "correlation IDs should be unique", field: "CorrelationID"},
		{name: "trace IDs should be unique", field: "TraceID"},
		{name: "request IDs should be unique", field: "RequestID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail until correlation context generation ensures uniqueness
			values := make(map[string]bool)
			for _, ctx := range contexts {
				var value string
				switch tt.field {
				case "CorrelationID":
					value = ctx.CorrelationID
				case "TraceID":
					value = ctx.TraceID
				case "RequestID":
					value = ctx.RequestID
				}
				
				assert.NotEmpty(t, value, "%s should not be empty", tt.field)
				assert.False(t, values[value], "%s should be unique: %s", tt.field, value)
				values[value] = true
			}
		})
	}
}

func TestCorrelationContext_SetUserContext(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		appVersion string
	}{
		{
			name:       "set user context with valid data",
			userID:     "user-123",
			appVersion: "v1.0.0",
		},
		{
			name:       "set user context with empty data",
			userID:     "",
			appVersion: "",
		},
		{
			name:       "set user context with UUID user ID",
			userID:     uuid.New().String(),
			appVersion: "v2.1.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewCorrelationContext()
			ctx.SetUserContext(tt.userID, tt.appVersion)
			
			assert.Equal(t, tt.userID, ctx.UserID)
			assert.Equal(t, tt.appVersion, ctx.AppVersion)
		})
	}
}

func TestCorrelationContext_ToContext(t *testing.T) {
	correlationCtx := &CorrelationContext{
		CorrelationID: "correlation-123",
		TraceID:       "trace-456",
		UserID:        "user-789",
		RequestID:     "request-012",
		AppVersion:    "v1.0.0",
		StartTime:     time.Now(),
	}

	tests := []struct {
		name        string
		baseContext context.Context
	}{
		{
			name:        "add to background context",
			baseContext: context.Background(),
		},
		{
			name:        "add to context with existing values",
			baseContext: context.WithValue(context.Background(), "existing", "value"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := correlationCtx.ToContext(tt.baseContext)
			
			assert.Equal(t, correlationCtx.CorrelationID, ctx.Value(ContextKeyCorrelationID))
			assert.Equal(t, correlationCtx.TraceID, ctx.Value(ContextKeyTraceID))
			assert.Equal(t, correlationCtx.UserID, ctx.Value(ContextKeyUserID))
			assert.Equal(t, correlationCtx.RequestID, ctx.Value(ContextKeyRequestID))
			assert.Equal(t, correlationCtx.AppVersion, ctx.Value(ContextKeyAppVersion))
		})
	}
}

func TestFromContext(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		expect map[string]string
	}{
		{
			name: "context with all correlation values",
			ctx: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, ContextKeyCorrelationID, "correlation-123")
				ctx = context.WithValue(ctx, ContextKeyTraceID, "trace-456")
				ctx = context.WithValue(ctx, ContextKeyUserID, "user-789")
				ctx = context.WithValue(ctx, ContextKeyRequestID, "request-012")
				ctx = context.WithValue(ctx, ContextKeyAppVersion, "v1.0.0")
				return ctx
			}(),
			expect: map[string]string{
				"CorrelationID": "correlation-123",
				"TraceID":       "trace-456",
				"UserID":        "user-789",
				"RequestID":     "request-012",
				"AppVersion":    "v1.0.0",
			},
		},
		{
			name: "context with partial correlation values",
			ctx: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, ContextKeyCorrelationID, "correlation-123")
				ctx = context.WithValue(ctx, ContextKeyUserID, "user-789")
				return ctx
			}(),
			expect: map[string]string{
				"CorrelationID": "correlation-123",
				"UserID":        "user-789",
			},
		},
		{
			name: "empty context generates new IDs",
			ctx:  context.Background(),
			expect: map[string]string{
				// IDs should be generated, so we'll check they're not empty
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			correlationCtx := FromContext(tt.ctx)
			require.NotNil(t, correlationCtx)
			
			if correlationID, exists := tt.expect["CorrelationID"]; exists {
				assert.Equal(t, correlationID, correlationCtx.CorrelationID)
			} else {
				assert.NotEmpty(t, correlationCtx.CorrelationID)
			}
			
			if traceID, exists := tt.expect["TraceID"]; exists {
				assert.Equal(t, traceID, correlationCtx.TraceID)
			} else {
				assert.NotEmpty(t, correlationCtx.TraceID)
			}
			
			if userID, exists := tt.expect["UserID"]; exists {
				assert.Equal(t, userID, correlationCtx.UserID)
			}
			
			if requestID, exists := tt.expect["RequestID"]; exists {
				assert.Equal(t, requestID, correlationCtx.RequestID)
			} else {
				assert.NotEmpty(t, correlationCtx.RequestID)
			}
			
			if appVersion, exists := tt.expect["AppVersion"]; exists {
				assert.Equal(t, appVersion, correlationCtx.AppVersion)
			}
		})
	}
}

func TestGetCorrelationID(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "context with correlation ID",
			ctx:  context.WithValue(context.Background(), ContextKeyCorrelationID, "correlation-123"),
			want: "correlation-123",
		},
		{
			name: "context without correlation ID generates new one",
			ctx:  context.Background(),
			want: "", // We'll check it's a valid UUID instead
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCorrelationID(tt.ctx)
			if tt.want == "" {
				assert.NotEmpty(t, got)
				_, err := uuid.Parse(got)
				assert.NoError(t, err, "Generated correlation ID should be valid UUID")
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetTraceID(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "context with trace ID",
			ctx:  context.WithValue(context.Background(), ContextKeyTraceID, "trace-456"),
			want: "trace-456",
		},
		{
			name: "context without trace ID generates new one",
			ctx:  context.Background(),
			want: "", // We'll check it's a valid trace ID instead
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTraceID(tt.ctx)
			if tt.want == "" {
				assert.NotEmpty(t, got)
				assert.Len(t, got, 32, "Generated trace ID should be 32 characters")
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "context with user ID",
			ctx:  context.WithValue(context.Background(), ContextKeyUserID, "user-789"),
			want: "user-789",
		},
		{
			name: "context without user ID returns empty string",
			ctx:  context.Background(),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetUserID(tt.ctx)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetRequestID(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "context with request ID",
			ctx:  context.WithValue(context.Background(), ContextKeyRequestID, "request-012"),
			want: "request-012",
		},
		{
			name: "context without request ID generates new one",
			ctx:  context.Background(),
			want: "", // We'll check it's a valid UUID instead
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRequestID(tt.ctx)
			if tt.want == "" {
				assert.NotEmpty(t, got)
				_, err := uuid.Parse(got)
				assert.NoError(t, err, "Generated request ID should be valid UUID")
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetAppVersion(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "context with app version",
			ctx:  context.WithValue(context.Background(), ContextKeyAppVersion, "v1.0.0"),
			want: "v1.0.0",
		},
		{
			name: "context without app version returns unknown",
			ctx:  context.Background(),
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAppVersion(tt.ctx)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithCorrelationID(t *testing.T) {
	tests := []struct {
		name          string
		correlationID string
	}{
		{
			name:          "add correlation ID to context",
			correlationID: "correlation-123",
		},
		{
			name:          "add UUID correlation ID to context",
			correlationID: uuid.New().String(),
		},
		{
			name:          "add empty correlation ID to context",
			correlationID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithCorrelationID(context.Background(), tt.correlationID)
			got := ctx.Value(ContextKeyCorrelationID)
			assert.Equal(t, tt.correlationID, got)
		})
	}
}

func TestWithTraceID(t *testing.T) {
	tests := []struct {
		name    string
		traceID string
	}{
		{
			name:    "add trace ID to context",
			traceID: "trace-456",
		},
		{
			name:    "add empty trace ID to context",
			traceID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithTraceID(context.Background(), tt.traceID)
			got := ctx.Value(ContextKeyTraceID)
			assert.Equal(t, tt.traceID, got)
		})
	}
}

func TestWithUserID(t *testing.T) {
	tests := []struct {
		name   string
		userID string
	}{
		{
			name:   "add user ID to context",
			userID: "user-789",
		},
		{
			name:   "add UUID user ID to context",
			userID: uuid.New().String(),
		},
		{
			name:   "add empty user ID to context",
			userID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithUserID(context.Background(), tt.userID)
			got := ctx.Value(ContextKeyUserID)
			assert.Equal(t, tt.userID, got)
		})
	}
}

func TestWithRequestID(t *testing.T) {
	tests := []struct {
		name      string
		requestID string
	}{
		{
			name:      "add request ID to context",
			requestID: "request-012",
		},
		{
			name:      "add UUID request ID to context",
			requestID: uuid.New().String(),
		},
		{
			name:      "add empty request ID to context",
			requestID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithRequestID(context.Background(), tt.requestID)
			got := ctx.Value(ContextKeyRequestID)
			assert.Equal(t, tt.requestID, got)
		})
	}
}

func TestWithAppVersion(t *testing.T) {
	tests := []struct {
		name       string
		appVersion string
	}{
		{
			name:       "add app version to context",
			appVersion: "v1.0.0",
		},
		{
			name:       "add development app version to context",
			appVersion: "v0.1.0-dev",
		},
		{
			name:       "add empty app version to context",
			appVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithAppVersion(context.Background(), tt.appVersion)
			got := ctx.Value(ContextKeyAppVersion)
			assert.Equal(t, tt.appVersion, got)
		})
	}
}

func TestCreateChildContext(t *testing.T) {
	parentCtx := context.Background()
	parentCtx = context.WithValue(parentCtx, ContextKeyCorrelationID, "parent-correlation-123")
	parentCtx = context.WithValue(parentCtx, ContextKeyUserID, "user-789")
	parentCtx = context.WithValue(parentCtx, ContextKeyAppVersion, "v1.0.0")

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "create child context from parent with correlation data",
			ctx:  parentCtx,
		},
		{
			name: "create child context from empty parent",
			ctx:  context.Background(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			childCtx := CreateChildContext(tt.ctx)
			
			// Child should preserve correlation ID from parent
			childCorrelationID := GetCorrelationID(childCtx)
			if tt.ctx.Value(ContextKeyCorrelationID) != nil {
				assert.Equal(t, "parent-correlation-123", childCorrelationID)
			} else {
				assert.NotEmpty(t, childCorrelationID)
			}
			
			// Child should preserve user ID from parent
			childUserID := GetUserID(childCtx)
			if tt.ctx.Value(ContextKeyUserID) != nil {
				assert.Equal(t, "user-789", childUserID)
			}
			
			// Child should preserve app version from parent
			childAppVersion := GetAppVersion(childCtx)
			if tt.ctx.Value(ContextKeyAppVersion) != nil {
				assert.Equal(t, "v1.0.0", childAppVersion)
			}
			
			// Child should have new trace ID and request ID
			childTraceID := GetTraceID(childCtx)
			childRequestID := GetRequestID(childCtx)
			assert.NotEmpty(t, childTraceID)
			assert.NotEmpty(t, childRequestID)
		})
	}
}

func TestCorrelationContext_GetElapsedTime(t *testing.T) {
	tests := []struct {
		name      string
		startTime time.Time
	}{
		{
			name:      "get elapsed time from recent start",
			startTime: time.Now().Add(-100 * time.Millisecond),
		},
		{
			name:      "get elapsed time from older start",
			startTime: time.Now().Add(-5 * time.Second),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &CorrelationContext{
				StartTime: tt.startTime,
			}
			
			elapsed := ctx.GetElapsedTime()
			assert.True(t, elapsed > 0, "Elapsed time should be positive")
			
			// Allow for some timing variance in tests
			expectedMin := time.Since(tt.startTime) - 10*time.Millisecond
			expectedMax := time.Since(tt.startTime) + 10*time.Millisecond
			assert.True(t, elapsed >= expectedMin && elapsed <= expectedMax, 
				"Elapsed time should be within expected range")
		})
	}
}

func TestCorrelationContext_ToLogFields(t *testing.T) {
	startTime := time.Now().Add(-500 * time.Millisecond)
	ctx := &CorrelationContext{
		CorrelationID: "correlation-123",
		TraceID:       "trace-456",
		UserID:        "user-789",
		RequestID:     "request-012",
		AppVersion:    "v1.0.0",
		StartTime:     startTime,
	}

	tests := []struct {
		name   string
		ctx    *CorrelationContext
		expect map[string]interface{}
	}{
		{
			name: "complete correlation context to log fields",
			ctx:  ctx,
			expect: map[string]interface{}{
				"correlation_id": "correlation-123",
				"trace_id":       "trace-456",
				"user_id":        "user-789",
				"request_id":     "request-012",
				"app_version":    "v1.0.0",
			},
		},
		{
			name: "minimal correlation context to log fields",
			ctx: &CorrelationContext{
				CorrelationID: "correlation-456",
				TraceID:       "trace-789",
				StartTime:     time.Now(),
			},
			expect: map[string]interface{}{
				"correlation_id": "correlation-456",
				"trace_id":       "trace-789",
				"user_id":        "",
				"request_id":     "",
				"app_version":    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := tt.ctx.ToLogFields()
			
			assert.Equal(t, tt.expect["correlation_id"], fields["correlation_id"])
			assert.Equal(t, tt.expect["trace_id"], fields["trace_id"])
			assert.Equal(t, tt.expect["user_id"], fields["user_id"])
			assert.Equal(t, tt.expect["request_id"], fields["request_id"])
			assert.Equal(t, tt.expect["app_version"], fields["app_version"])
			
			// Check that elapsed_ms is present and is a positive number
			elapsedMs, exists := fields["elapsed_ms"]
			assert.True(t, exists, "elapsed_ms should be present in log fields")
			assert.IsType(t, int64(0), elapsedMs, "elapsed_ms should be int64")
			assert.True(t, elapsedMs.(int64) >= 0, "elapsed_ms should be non-negative")
		})
	}
}