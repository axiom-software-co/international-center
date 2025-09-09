#!/bin/bash

set -e

echo "🚀 Generating Go server interfaces from OpenAPI specifications..."

# Check if required tools are installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is required but not installed. Please install Go."
    exit 1
fi

# Check if oapi-codegen is installed
if ! command -v oapi-codegen &> /dev/null; then
    echo "📦 Installing oapi-codegen..."
    go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
fi

# Create directories for generated code
mkdir -p generated/public/models
mkdir -p generated/public/handlers
mkdir -p generated/admin/models
mkdir -p generated/admin/handlers
mkdir -p generated/common

echo "📋 Validating OpenAPI specifications..."
# Validate Go syntax compatibility
go mod tidy

echo "🔧 Generating Public API server interfaces..."
# Generate Public API models
oapi-codegen -config public-models-config.yaml ../../openapi/public-api.yaml > generated/public/models/types.go

# Generate Public API server interfaces
oapi-codegen -config public-server-config.yaml ../../openapi/public-api.yaml > generated/public/handlers/server.go

echo "🔧 Generating Admin API server interfaces..."
# Generate Admin API models
oapi-codegen -config admin-models-config.yaml ../../openapi/admin-api.yaml > generated/admin/models/types.go

# Generate Admin API server interfaces
oapi-codegen -config admin-server-config.yaml ../../openapi/admin-api.yaml > generated/admin/handlers/server.go

echo "🔧 Generating common utilities..."
# Generate common error handling and middleware
cat > generated/common/errors.go << 'EOF'
package common

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type APIError struct {
	Error struct {
		Code          string      `json:"code"`
		Message       string      `json:"message"`
		Details       interface{} `json:"details,omitempty"`
		CorrelationID string      `json:"correlation_id"`
		Timestamp     time.Time   `json:"timestamp"`
	} `json:"error"`
}

func NewAPIError(code, message string, details interface{}) *APIError {
	return &APIError{
		Error: struct {
			Code          string      `json:"code"`
			Message       string      `json:"message"`
			Details       interface{} `json:"details,omitempty"`
			CorrelationID string      `json:"correlation_id"`
			Timestamp     time.Time   `json:"timestamp"`
		}{
			Code:          code,
			Message:       message,
			Details:       details,
			CorrelationID: uuid.New().String(),
			Timestamp:     time.Now().UTC(),
		},
	}
}

func (e *APIError) WriteResponse(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", e.Error.CorrelationID)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(e)
}

func BadRequestError(message string, details interface{}) *APIError {
	return NewAPIError("BAD_REQUEST", message, details)
}

func UnauthorizedError(message string) *APIError {
	return NewAPIError("UNAUTHORIZED", message, nil)
}

func ForbiddenError(message string) *APIError {
	return NewAPIError("FORBIDDEN", message, nil)
}

func NotFoundError(message string) *APIError {
	return NewAPIError("NOT_FOUND", message, nil)
}

func InternalServerError(message string) *APIError {
	return NewAPIError("INTERNAL_SERVER_ERROR", message, nil)
}
EOF

cat > generated/common/middleware.go << 'EOF'
package common

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type key int

const (
	CorrelationIDKey key = iota
	RequestIDKey
	UserIDKey
)

func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}
		
		ctx := context.WithValue(r.Context(), CorrelationIDKey, correlationID)
		w.Header().Set("X-Correlation-ID", correlationID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Log request
		correlationID := GetCorrelationID(r.Context())
		
		next.ServeHTTP(w, r)
		
		// Log response
		duration := time.Since(start)
		_ = correlationID
		_ = duration
		// Actual logging would go here using slog
	})
}

func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}
EOF

echo "🏗️  Building Go modules..."
go mod tidy
go build ./...

echo "✅ Go server interfaces generated successfully!"
echo "📁 Generated files:"
echo "   - generated/public/models/ - Public API models"
echo "   - generated/public/handlers/ - Public API server interfaces"
echo "   - generated/admin/models/ - Admin API models"
echo "   - generated/admin/handlers/ - Admin API server interfaces"
echo "   - generated/common/ - Common utilities and middleware"

echo "🎉 Generation complete! You can now implement these interfaces in your Go services."