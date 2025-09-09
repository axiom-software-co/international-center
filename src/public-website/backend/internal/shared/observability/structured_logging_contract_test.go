package observability

import (
	"context"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TDD Cycle 7 RED Phase: Structured Logging Contract Tests
//
// WHY: Medical-grade compliance requires comprehensive audit trails and developer-focused structured logging
// SCOPE: All services and gateways with correlation ID propagation and contextual information
// DEPENDENCIES: slog integration, correlation context propagation, log level management
// CONTEXT: Gateway architecture requiring complete request traceability and debugging capabilities

func TestStructuredLogger_Contract(t *testing.T) {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tests := []struct {
		name               string
		service            string
		logLevel           string
		expectedFields     []string
		requiresCorrelation bool
	}{
		{
			name:     "content-api structured logger with full context",
			service:  "content-api",
			logLevel: "info",
			expectedFields: []string{
				"service_name",
				"correlation_id",
				"trace_id",
				"request_id",
				"user_id",
				"app_version",
				"timestamp",
				"level",
				"message",
			},
			requiresCorrelation: true,
		},
		{
			name:     "inquiries-api structured logger with business context",
			service:  "inquiries-api",
			logLevel: "warning",
			expectedFields: []string{
				"service_name",
				"correlation_id",
				"trace_id",
				"inquiry_type",
				"processing_stage",
				"timestamp",
				"level",
			},
			requiresCorrelation: true,
		},
		{
			name:     "admin-gateway structured logger with security context",
			service:  "admin-gateway",
			logLevel: "error",
			expectedFields: []string{
				"service_name",
				"correlation_id",
				"user_id",
				"admin_role",
				"security_event",
				"request_url",
				"timestamp",
				"level",
			},
			requiresCorrelation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TDD RED Phase - This will fail until implementation exists
			logger, err := NewStructuredLogger(tt.service)
			require.NoError(t, err, "structured logger creation should succeed")
			require.NotNil(t, logger, "logger instance should not be nil")

			// Set up correlation context for testing
			correlationCtx := domain.NewCorrelationContext()
			correlationCtx.SetUserContext("test-user-123", "v1.0.0")
			ctx = correlationCtx.ToContext(ctx)

			// Contract: Logger must accept and structure log entries
			logEntry, err := logger.CreateLogEntry(ctx, tt.logLevel, "Test log message")
			assert.NoError(t, err, "log entry creation should succeed")
			assert.NotNil(t, logEntry, "log entry should not be nil")

			// Contract: Log entry must contain all required structured fields
			fields := logEntry.GetFields()
			for _, field := range tt.expectedFields {
				assert.Contains(t, fields, field, "log entry should contain field %s", field)
				
				// Ensure critical fields are not empty
				if field == "correlation_id" || field == "trace_id" || field == "timestamp" {
					assert.NotEmpty(t, fields[field], "critical field %s should not be empty", field)
				}
			}

			// Contract: Log level must be properly set
			assert.Equal(t, tt.logLevel, fields["level"], "log level should match expected")

			// Contract: Logger must be able to output structured format
			output, err := logEntry.ToJSON()
			assert.NoError(t, err, "log entry should be serializable to JSON")
			assert.NotEmpty(t, output, "JSON output should not be empty")
		})
	}
}

func TestLogLevels_Contract(t *testing.T) {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Set up correlation context
	correlationCtx := domain.NewCorrelationContext()
	ctx = correlationCtx.ToContext(ctx)

	tests := []struct {
		name        string
		level       string
		message     string
		shouldLog   bool
		environment string
	}{
		{
			name:        "debug level in development environment",
			level:       "debug",
			message:     "Debug information for development",
			shouldLog:   true,
			environment: "development",
		},
		{
			name:        "debug level in production environment",
			level:       "debug", 
			message:     "Debug information for production",
			shouldLog:   false,
			environment: "production",
		},
		{
			name:        "info level in all environments",
			level:       "info",
			message:     "Informational message",
			shouldLog:   true,
			environment: "production",
		},
		{
			name:        "warning level in all environments",
			level:       "warning",
			message:     "Warning message",
			shouldLog:   true,
			environment: "production",
		},
		{
			name:        "error level in all environments",
			level:       "error",
			message:     "Error message",
			shouldLog:   true,
			environment: "production",
		},
		{
			name:        "critical level in all environments",
			level:       "critical",
			message:     "Critical system error",
			shouldLog:   true,
			environment: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TDD RED Phase - This will fail until implementation exists
			logger, err := NewEnvironmentLogger("test-service", tt.environment)
			require.NoError(t, err, "environment logger creation should succeed")
			require.NotNil(t, logger, "logger should not be nil")

			// Contract: Logger must respect log level filtering based on environment
			shouldLog := logger.ShouldLog(tt.level)
			assert.Equal(t, tt.shouldLog, shouldLog, "log level filtering should match environment expectations")

			if tt.shouldLog {
				// Contract: Logging should succeed when level is enabled
				err = logger.Log(ctx, tt.level, tt.message)
				assert.NoError(t, err, "logging should succeed for enabled level")
			}
		})
	}
}

func TestCorrelationContextLogging_Contract(t *testing.T) {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tests := []struct {
		name           string
		correlationID  string
		userID         string
		appVersion     string
		expectPropagation bool
	}{
		{
			name:              "full correlation context propagation",
			correlationID:     "corr-12345",
			userID:            "user-67890",
			appVersion:        "v1.2.3",
			expectPropagation: true,
		},
		{
			name:              "partial correlation context propagation",
			correlationID:     "corr-54321",
			userID:            "",
			appVersion:        "v1.2.3",
			expectPropagation: true,
		},
		{
			name:              "missing correlation context handling",
			correlationID:     "",
			userID:            "",
			appVersion:        "",
			expectPropagation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TDD RED Phase - This will fail until implementation exists
			logger, err := NewCorrelationLogger("test-service")
			require.NoError(t, err, "correlation logger creation should succeed")
			require.NotNil(t, logger, "correlation logger should not be nil")

			// Set up correlation context based on test case
			var testCtx context.Context
			if tt.expectPropagation {
				correlationCtx := domain.NewCorrelationContextWithID(tt.correlationID)
				correlationCtx.SetUserContext(tt.userID, tt.appVersion)
				testCtx = correlationCtx.ToContext(ctx)
			} else {
				testCtx = ctx
			}

			// Contract: Logger must extract and include correlation information
			logEntry, err := logger.LogWithCorrelation(testCtx, "info", "Test correlation message")
			assert.NoError(t, err, "correlation logging should succeed")
			assert.NotNil(t, logEntry, "log entry should be created")

			fields := logEntry.GetFields()

			if tt.expectPropagation {
				// Contract: Correlation fields must be present when context is available
				if tt.correlationID != "" {
					assert.Equal(t, tt.correlationID, fields["correlation_id"], "correlation ID should match")
				} else {
					assert.NotEmpty(t, fields["correlation_id"], "correlation ID should be generated when missing")
				}

				if tt.userID != "" {
					assert.Equal(t, tt.userID, fields["user_id"], "user ID should match")
				}

				if tt.appVersion != "" {
					assert.Equal(t, tt.appVersion, fields["app_version"], "app version should match")
				}

				assert.NotEmpty(t, fields["trace_id"], "trace ID should be present")
			} else {
				// Contract: Logger must generate correlation information when missing
				assert.NotEmpty(t, fields["correlation_id"], "correlation ID should be generated")
				assert.NotEmpty(t, fields["trace_id"], "trace ID should be generated")
			}
		})
	}
}

func TestServiceInvocationLogging_Contract(t *testing.T) {
	timeout := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Set up correlation context
	correlationCtx := domain.NewCorrelationContext()
	correlationCtx.SetUserContext("test-user", "v1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	tests := []struct {
		name            string
		targetService   string
		method          string
		expectedFields  []string
	}{
		{
			name:          "content-api service invocation logging",
			targetService: "content-api",
			method:        "api/v1/content",
			expectedFields: []string{
				"service_invocation.target",
				"service_invocation.method",
				"service_invocation.start_time",
				"correlation_id",
				"trace_id",
			},
		},
		{
			name:          "inquiries-api service invocation logging",
			targetService: "inquiries-api",
			method:        "api/v1/inquiries/business",
			expectedFields: []string{
				"service_invocation.target",
				"service_invocation.method",
				"service_invocation.duration",
				"correlation_id",
				"trace_id",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TDD RED Phase - This will fail until implementation exists
			invocationLogger, err := NewServiceInvocationLogger()
			require.NoError(t, err, "service invocation logger creation should succeed")
			require.NotNil(t, invocationLogger, "invocation logger should not be nil")

			// Contract: Service invocations must be automatically logged with context
			logEntry, err := invocationLogger.LogServiceInvocation(ctx, tt.targetService, tt.method)
			assert.NoError(t, err, "service invocation logging should succeed")
			assert.NotNil(t, logEntry, "log entry should be created")

			// Contract: Service invocation logs must contain required fields
			fields := logEntry.GetFields()
			for _, field := range tt.expectedFields {
				assert.Contains(t, fields, field, "service invocation log should contain field %s", field)
			}

			// Contract: Service invocation completion must be logged
			err = invocationLogger.LogServiceInvocationComplete(ctx, logEntry, 200, 150*time.Millisecond)
			assert.NoError(t, err, "service invocation completion logging should succeed")
		})
	}
}

func TestErrorLogging_Contract(t *testing.T) {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Set up correlation context
	correlationCtx := domain.NewCorrelationContext()
	correlationCtx.SetUserContext("test-user", "v1.0.0")
	ctx = correlationCtx.ToContext(ctx)

	tests := []struct {
		name           string
		errorType      string
		severity       string
		expectedFields []string
		expectAlert    bool
	}{
		{
			name:      "database connection error logging",
			errorType: "database.connection.failed",
			severity:  "critical",
			expectedFields: []string{
				"error.type",
				"error.severity",
				"error.stack_trace",
				"service_name",
				"correlation_id",
				"timestamp",
			},
			expectAlert: true,
		},
		{
			name:      "service invocation error logging",
			errorType: "service.invocation.timeout",
			severity:  "error",
			expectedFields: []string{
				"error.type",
				"error.target_service",
				"error.timeout_duration",
				"correlation_id",
				"trace_id",
			},
			expectAlert: false,
		},
		{
			name:      "validation error logging",
			errorType: "validation.failed",
			severity:  "warning",
			expectedFields: []string{
				"error.type",
				"validation.field",
				"validation.constraint",
				"user_id",
				"correlation_id",
			},
			expectAlert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TDD RED Phase - This will fail until implementation exists
			errorLogger, err := NewErrorLogger("test-service")
			require.NoError(t, err, "error logger creation should succeed")
			require.NotNil(t, errorLogger, "error logger should not be nil")

			// Contract: Error logging must capture comprehensive error context
			testError := &TestError{Type: tt.errorType, Message: "Test error message"}
			logEntry, alert, err := errorLogger.LogError(ctx, testError, tt.severity)
			
			assert.NoError(t, err, "error logging should succeed")
			assert.NotNil(t, logEntry, "error log entry should be created")

			// Contract: Error logs must contain required fields
			fields := logEntry.GetFields()
			for _, field := range tt.expectedFields {
				assert.Contains(t, fields, field, "error log should contain field %s", field)
			}

			// Contract: Critical errors must generate alerts
			if tt.expectAlert {
				assert.NotNil(t, alert, "critical error should generate alert")
				assert.Equal(t, tt.severity, alert.Severity, "alert severity should match error severity")
			} else {
				assert.Nil(t, alert, "non-critical error should not generate alert")
			}
		})
	}
}

func TestPerformanceLogging_Contract(t *testing.T) {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Set up correlation context
	correlationCtx := domain.NewCorrelationContext()
	ctx = correlationCtx.ToContext(ctx)

	tests := []struct {
		name           string
		operation      string
		duration       time.Duration
		expectedFields []string
		expectSLOAlert bool
	}{
		{
			name:      "HTTP request performance logging",
			operation: "http.request.process",
			duration:  50 * time.Millisecond,
			expectedFields: []string{
				"performance.operation",
				"performance.duration_ms",
				"performance.slo_compliance",
				"correlation_id",
			},
			expectSLOAlert: false,
		},
		{
			name:      "slow database query performance logging",
			operation: "database.query.execute",
			duration:  2500 * time.Millisecond,
			expectedFields: []string{
				"performance.operation",
				"performance.duration_ms",
				"performance.slo_violation",
				"database.query_type",
			},
			expectSLOAlert: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TDD RED Phase - This will fail until implementation exists
			perfLogger, err := NewPerformanceLogger("test-service")
			require.NoError(t, err, "performance logger creation should succeed")
			require.NotNil(t, perfLogger, "performance logger should not be nil")

			// Contract: Performance logging must track operation timing
			logEntry, sloAlert, err := perfLogger.LogPerformance(ctx, tt.operation, tt.duration)
			assert.NoError(t, err, "performance logging should succeed")
			assert.NotNil(t, logEntry, "performance log entry should be created")

			// Contract: Performance logs must contain timing and SLO information
			fields := logEntry.GetFields()
			for _, field := range tt.expectedFields {
				assert.Contains(t, fields, field, "performance log should contain field %s", field)
			}

			// Contract: SLO violations must generate alerts
			if tt.expectSLOAlert {
				assert.NotNil(t, sloAlert, "SLO violation should generate alert")
				assert.Contains(t, sloAlert.Message, "SLO violation", "alert should mention SLO violation")
			} else {
				assert.Nil(t, sloAlert, "compliant performance should not generate SLO alert")
			}
		})
	}
}

func TestLogConfiguration_Contract(t *testing.T) {
	tests := []struct {
		name           string
		environment    string
		expectedConfig map[string]interface{}
	}{
		{
			name:        "development environment logging configuration",
			environment: "development",
			expectedConfig: map[string]interface{}{
				"level":           "debug",
				"format":          "json",
				"output":          "stdout",
				"correlation":     true,
				"performance":     true,
			},
		},
		{
			name:        "production environment logging configuration",
			environment: "production",
			expectedConfig: map[string]interface{}{
				"level":           "info",
				"format":          "json",
				"output":          "grafana-loki",
				"correlation":     true,
				"sampling_rate":   0.1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TDD RED Phase - This will fail until implementation exists
			config, err := LoadLoggingConfiguration(tt.environment)
			assert.NoError(t, err, "logging configuration loading should succeed")
			assert.NotNil(t, config, "configuration should not be nil")

			// Contract: Configuration must contain required logging settings
			for key, expectedValue := range tt.expectedConfig {
				actualValue, exists := config.GetValue(key)
				assert.True(t, exists, "configuration should contain key %s", key)
				assert.Equal(t, expectedValue, actualValue, "configuration value for %s should match expected", key)
			}
		})
	}
}

// TestError is a helper type for testing error logging
type TestError struct {
	Type    string
	Message string
}

func (e *TestError) Error() string {
	return e.Message
}

func (e *TestError) GetType() string {
	return e.Type
}