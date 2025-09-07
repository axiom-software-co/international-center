package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// AuditEventType represents the type of operation being audited
type AuditEventType string

const (
	AuditEventCreate     AuditEventType = "CREATE"
	AuditEventUpdate     AuditEventType = "UPDATE"
	AuditEventDelete     AuditEventType = "DELETE"
	AuditEventPublish    AuditEventType = "PUBLISH"
	AuditEventArchive    AuditEventType = "ARCHIVE"
	AuditEventView       AuditEventType = "VIEW"
	AuditEventUpload     AuditEventType = "UPLOAD"
	AuditEventDownload   AuditEventType = "DOWNLOAD"
	AuditEventAccess     AuditEventType = "ACCESS"
)

// AuditResourceType represents the type of resource being operated on
type AuditResourceType string

const (
	AuditResourceServices    AuditResourceType = "SERVICES"
	AuditResourceNews        AuditResourceType = "NEWS"
	AuditResourceResearch    AuditResourceType = "RESEARCH"
	AuditResourceEvents      AuditResourceType = "EVENTS"
	AuditResourceInquiries   AuditResourceType = "INQUIRIES"
	AuditResourceCategories  AuditResourceType = "CATEGORIES"
	AuditResourceUsers       AuditResourceType = "USERS"
	AuditResourceReports     AuditResourceType = "REPORTS"
	AuditResourceAuditLogs   AuditResourceType = "AUDIT_LOGS"
)

// AuditEvent represents a single audit event
type AuditEvent struct {
	Timestamp      time.Time         `json:"timestamp"`
	CorrelationID  string            `json:"correlation_id"`
	UserID         string            `json:"user_id"`
	EventType      AuditEventType    `json:"event_type"`
	ResourceType   AuditResourceType `json:"resource_type"`
	ResourceID     string            `json:"resource_id,omitempty"`
	Path           string            `json:"path"`
	Method         string            `json:"method"`
	StatusCode     int               `json:"status_code"`
	Duration       time.Duration     `json:"duration"`
	RemoteAddr     string            `json:"remote_addr"`
	UserAgent      string            `json:"user_agent"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Success        bool              `json:"success"`
	ErrorMessage   string            `json:"error_message,omitempty"`
	Environment    string            `json:"environment"`
	GatewayVersion string            `json:"gateway_version"`
}

// AuditService handles audit logging for admin operations
type AuditService struct {
	logger      *slog.Logger
	environment string
	version     string
}

// NewAuditService creates a new audit service
func NewAuditService(environment, version string) *AuditService {
	// Create structured logger for audit events
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	
	return &AuditService{
		logger:      logger,
		environment: environment,
		version:     version,
	}
}

// LogAdminOperation logs an admin operation for audit purposes
func (s *AuditService) LogAdminOperation(ctx context.Context, r *http.Request, statusCode int, duration time.Duration, err error) {
	correlationID := domain.GetCorrelationID(ctx)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "unknown"
	}

	event := AuditEvent{
		Timestamp:      time.Now().UTC(),
		CorrelationID:  correlationID,
		UserID:         userID,
		Path:           r.URL.Path,
		Method:         r.Method,
		StatusCode:     statusCode,
		Duration:       duration,
		RemoteAddr:     s.extractClientIP(r),
		UserAgent:      r.Header.Get("User-Agent"),
		Success:        statusCode < 400 && err == nil,
		Environment:    s.environment,
		GatewayVersion: s.version,
	}

	// Set error message if present
	if err != nil {
		event.ErrorMessage = err.Error()
	}

	// Parse operation type and resource from path
	event.EventType, event.ResourceType, event.ResourceID = s.parseAdminOperation(r)

	// Add metadata based on operation type
	event.Metadata = s.buildMetadata(r, event.EventType)

	// Log the audit event
	s.logAuditEvent(event)
}

// parseAdminOperation parses the HTTP request to determine operation type and resource
func (s *AuditService) parseAdminOperation(r *http.Request) (AuditEventType, AuditResourceType, string) {
	path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/")
	parts := strings.Split(path, "/")
	
	if len(parts) == 0 {
		return AuditEventAccess, AuditResourceUsers, ""
	}

	// Determine resource type
	var resourceType AuditResourceType
	switch parts[0] {
	case "services":
		resourceType = AuditResourceServices
	case "news":
		resourceType = AuditResourceNews
	case "research":
		resourceType = AuditResourceResearch
	case "events":
		resourceType = AuditResourceEvents
	case "inquiries":
		resourceType = AuditResourceInquiries
	default:
		resourceType = AuditResourceUsers
	}

	// Extract resource ID if present
	var resourceID string
	if len(parts) > 1 && !strings.Contains(parts[1], "categories") && !strings.Contains(parts[1], "featured") {
		resourceID = parts[1]
	}

	// Determine operation type based on method and path
	var eventType AuditEventType
	switch r.Method {
	case "GET":
		if strings.Contains(path, "/audit") {
			eventType = AuditEventView
			resourceType = AuditResourceAuditLogs
		} else {
			eventType = AuditEventView
		}
	case "POST":
		if strings.Contains(path, "/upload") {
			eventType = AuditEventUpload
			resourceType = AuditResourceReports
		} else {
			eventType = AuditEventCreate
		}
	case "PUT", "PATCH":
		if strings.Contains(path, "/publish") {
			eventType = AuditEventPublish
		} else if strings.Contains(path, "/archive") {
			eventType = AuditEventArchive
		} else {
			eventType = AuditEventUpdate
		}
	case "DELETE":
		eventType = AuditEventDelete
	default:
		eventType = AuditEventAccess
	}

	// Handle categories as separate resource type
	if strings.Contains(path, "/categories") {
		resourceType = AuditResourceCategories
	}

	return eventType, resourceType, resourceID
}

// buildMetadata creates metadata based on the operation type
func (s *AuditService) buildMetadata(r *http.Request, eventType AuditEventType) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Add query parameters for GET requests
	if r.Method == "GET" && len(r.URL.RawQuery) > 0 {
		metadata["query_params"] = r.URL.RawQuery
	}

	// Add content length for data operations
	if r.ContentLength > 0 {
		metadata["content_length"] = r.ContentLength
	}

	// Add content type for upload operations
	if eventType == AuditEventUpload {
		metadata["content_type"] = r.Header.Get("Content-Type")
	}

	// Add referer for tracking navigation patterns
	if referer := r.Header.Get("Referer"); referer != "" {
		metadata["referer"] = referer
	}

	return metadata
}

// extractClientIP extracts the client IP address from request
func (s *AuditService) extractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}

	return ip
}

// logAuditEvent logs the audit event using structured logging
func (s *AuditService) logAuditEvent(event AuditEvent) {
	// Convert to JSON for structured logging
	eventJSON, err := json.Marshal(event)
	if err != nil {
		s.logger.Error("Failed to marshal audit event", 
			"error", err,
			"correlation_id", event.CorrelationID,
		)
		return
	}

	// Log with appropriate level based on operation criticality
	logLevel := slog.LevelInfo
	if !event.Success {
		logLevel = slog.LevelError
	} else if event.EventType == AuditEventDelete || 
			  event.EventType == AuditEventPublish || 
			  event.EventType == AuditEventArchive {
		logLevel = slog.LevelWarn // Higher visibility for critical operations
	}

	s.logger.Log(context.Background(), logLevel, "Admin operation audit",
		"audit_event", string(eventJSON),
		"correlation_id", event.CorrelationID,
		"user_id", event.UserID,
		"operation", fmt.Sprintf("%s_%s", event.EventType, event.ResourceType),
		"resource_id", event.ResourceID,
		"success", event.Success,
		"duration_ms", event.Duration.Milliseconds(),
	)
}

// AuditMiddleware returns middleware that logs admin operations
func (s *AuditService) AuditMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			ww := &auditResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(ww, r)

			// Log the audit event for admin operations only
			if s.isAdminOperation(r.URL.Path) {
				duration := time.Since(start)
				s.LogAdminOperation(r.Context(), r, ww.statusCode, duration, nil)
			}
		})
	}
}

// isAdminOperation checks if the path is an admin operation that should be audited
func (s *AuditService) isAdminOperation(path string) bool {
	return strings.HasPrefix(path, "/admin/api/v1/")
}

// auditResponseWriter wraps http.ResponseWriter to capture status code
type auditResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *auditResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}