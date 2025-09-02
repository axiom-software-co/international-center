package gateway

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

type AuditEvent struct {
	Timestamp    time.Time     `json:"timestamp"`
	UserID       string        `json:"user_id"`
	UserEmail    string        `json:"user_email"`
	Action       string        `json:"action"`
	Resource     string        `json:"resource"`
	IPAddress    string        `json:"ip_address"`
	UserAgent    string        `json:"user_agent"`
	StatusCode   int           `json:"status_code"`
	ResponseTime time.Duration `json:"response_time_ms"`
	RequestSize  int64         `json:"request_size_bytes"`
	ResponseSize int64         `json:"response_size_bytes"`
	Success      bool          `json:"success"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

type AuditLogger struct {
	// In production, this would integrate with Grafana Loki or similar
	// For CICD testing, log to structured JSON
}

func NewAuditLogger() *AuditLogger {
	return &AuditLogger{}
}

func (al *AuditLogger) Log(event *AuditEvent) {
	// Determine if the request was successful based on status code
	event.Success = event.StatusCode >= 200 && event.StatusCode < 400
	
	// Convert response time to milliseconds
	event.ResponseTime = event.ResponseTime / time.Millisecond
	
	// Marshal to JSON for structured logging
	auditJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("AUDIT_ERROR - Failed to marshal audit event: %v", err)
		return
	}
	
	// Log the audit event
	// In production, this would be sent to Grafana Loki or audit storage
	log.Printf("AUDIT - %s", string(auditJSON))
}

// Context helpers for user claims
type contextKey string

const userClaimsKey contextKey = "user_claims"

func contextWithUserClaims(ctx context.Context, claims *UserClaims) context.Context {
	return context.WithValue(ctx, userClaimsKey, claims)
}