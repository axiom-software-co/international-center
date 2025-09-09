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
