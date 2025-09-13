package dapr

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
)

// DataLayerMonitor provides comprehensive monitoring and observability for Dapr data operations
type DataLayerMonitor struct {
	stateStore      *StateStore
	pubsub          *PubSub
	secrets         *Secrets
	transactionMgr  *TransactionManager
	configManager   *ConfigManager
	aggregator      *MetricsAggregator
	healthChecker   *HealthChecker
	alertManager    *AlertManager
	config          *MonitoringConfig
	enabled         bool
}

// MetricsAggregator collects and aggregates metrics from all data layer components
type MetricsAggregator struct {
	stateStoreMetrics   *StateMetrics
	pubsubMetrics       *PubSubMetrics
	secretsMetrics      *SecretsMetrics
	transactionMetrics  *TransactionMetrics
	systemMetrics       *SystemMetrics
	customMetrics       map[string]interface{}
	lastCollection      time.Time
	collectionInterval  time.Duration
	mu                  sync.RWMutex
}

// HealthChecker monitors the health of all data layer components
type HealthChecker struct {
	checks          map[string]*HealthCheck
	checkInterval   time.Duration
	timeout         time.Duration
	lastCheck       time.Time
	overallStatus   HealthStatus
	mu              sync.RWMutex
}

// HealthCheck represents a single health check
type HealthCheck struct {
	Name            string                 `json:"name"`
	Status          HealthStatus           `json:"status"`
	LastCheck       time.Time              `json:"last_check"`
	Duration        time.Duration          `json:"duration"`
	Error           string                 `json:"error,omitempty"`
	Details         map[string]interface{} `json:"details,omitempty"`
	CheckFunc       func(ctx context.Context) error
}

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// AlertManager handles alerting based on metrics and health checks
type AlertManager struct {
	rules         []*AlertRule
	notifications []AlertNotification
	thresholds    map[string]*Threshold
	enabled       bool
	mu            sync.RWMutex
}

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	Name          string                 `json:"name"`
	Condition     string                 `json:"condition"`
	Threshold     float64                `json:"threshold"`
	Duration      time.Duration          `json:"duration"`
	Severity      AlertSeverity          `json:"severity"`
	Description   string                 `json:"description"`
	Tags          map[string]string      `json:"tags"`
	EvaluateFunc  func(metrics interface{}) bool
}

// AlertNotification represents an active alert
type AlertNotification struct {
	ID          string                 `json:"id"`
	Rule        string                 `json:"rule"`
	Severity    AlertSeverity          `json:"severity"`
	Message     string                 `json:"message"`
	Timestamp   time.Time              `json:"timestamp"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Context     map[string]interface{} `json:"context"`
}

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityInfo     AlertSeverity = "info"
)

// Threshold defines alert thresholds for various metrics
type Threshold struct {
	Warning  float64 `json:"warning"`
	Critical float64 `json:"critical"`
	Unit     string  `json:"unit"`
}

// SecretsMetrics tracks metrics for secret store operations
type SecretsMetrics struct {
	SecretsRetrieved    int64            `json:"secrets_retrieved"`
	SecretsRotated      int64            `json:"secrets_rotated"`
	ValidationFailures  int64            `json:"validation_failures"`
	CacheHits          int64            `json:"cache_hits"`
	CacheMisses        int64            `json:"cache_misses"`
	OperationLatencies map[string]int64 `json:"operation_latencies_ms"`
	ErrorCounts        map[string]int64 `json:"error_counts"`
	mu                 sync.RWMutex
}

// TransactionMetrics tracks metrics for transaction manager operations
type TransactionMetrics struct {
	SagasStarted        int64            `json:"sagas_started"`
	SagasCompleted      int64            `json:"sagas_completed"`
	SagasFailed         int64            `json:"sagas_failed"`
	SagasCompensated    int64            `json:"sagas_compensated"`
	AvgSagaDuration     time.Duration    `json:"avg_saga_duration"`
	TransactionConflicts int64           `json:"transaction_conflicts"`
	CompensationEvents  int64            `json:"compensation_events"`
	OperationLatencies  map[string]int64 `json:"operation_latencies_ms"`
	mu                  sync.RWMutex
}

// SystemMetrics tracks overall system performance metrics
type SystemMetrics struct {
	CPUUsage           float64   `json:"cpu_usage_percent"`
	MemoryUsage        float64   `json:"memory_usage_percent"`
	DiskUsage          float64   `json:"disk_usage_percent"`
	NetworkIOBytes     int64     `json:"network_io_bytes"`
	ConnectionPoolSize int       `json:"connection_pool_size"`
	Uptime             time.Duration `json:"uptime"`
	LastUpdated        time.Time `json:"last_updated"`
}

// MonitoringConfig configures the monitoring system
type MonitoringConfig struct {
	Enabled                bool          `json:"enabled"`
	MetricsInterval        time.Duration `json:"metrics_interval"`
	HealthCheckInterval    time.Duration `json:"health_check_interval"`
	HealthCheckTimeout     time.Duration `json:"health_check_timeout"`
	AlertsEnabled          bool          `json:"alerts_enabled"`
	PrometheusEnabled      bool          `json:"prometheus_enabled"`
	PrometheusPort         int           `json:"prometheus_port"`
	LogMetrics             bool          `json:"log_metrics"`
	RetentionDays          int           `json:"retention_days"`
}

// NewDataLayerMonitor creates a new data layer monitor
func NewDataLayerMonitor(stateStore *StateStore, pubsub *PubSub, secrets *Secrets, 
	transactionMgr *TransactionManager, configManager *ConfigManager) *DataLayerMonitor {
	
	config := &MonitoringConfig{
		Enabled:             getEnv("MONITORING_ENABLED", "true") == "true",
		MetricsInterval:     parseDurationEnv("MONITORING_METRICS_INTERVAL", 30*time.Second),
		HealthCheckInterval: parseDurationEnv("MONITORING_HEALTH_INTERVAL", 60*time.Second),
		HealthCheckTimeout:  parseDurationEnv("MONITORING_HEALTH_TIMEOUT", 10*time.Second),
		AlertsEnabled:       getEnv("MONITORING_ALERTS_ENABLED", "true") == "true",
		PrometheusEnabled:   getEnv("MONITORING_PROMETHEUS_ENABLED", "true") == "true",
		PrometheusPort:      parseIntEnv("MONITORING_PROMETHEUS_PORT", 2112),
		LogMetrics:         getEnv("MONITORING_LOG_METRICS", "true") == "true",
		RetentionDays:      parseIntEnv("MONITORING_RETENTION_DAYS", 7),
	}

	monitor := &DataLayerMonitor{
		stateStore:     stateStore,
		pubsub:         pubsub,
		secrets:        secrets,
		transactionMgr: transactionMgr,
		configManager:  configManager,
		config:         config,
		enabled:        config.Enabled,
	}

	if monitor.enabled {
		monitor.initializeComponents()
		monitor.startMonitoring()
	}

	return monitor
}

// initializeComponents initializes all monitoring components
func (m *DataLayerMonitor) initializeComponents() {
	// Initialize metrics aggregator
	m.aggregator = &MetricsAggregator{
		customMetrics:      make(map[string]interface{}),
		collectionInterval: m.config.MetricsInterval,
	}

	// Initialize health checker
	m.healthChecker = &HealthChecker{
		checks:        make(map[string]*HealthCheck),
		checkInterval: m.config.HealthCheckInterval,
		timeout:       m.config.HealthCheckTimeout,
		overallStatus: HealthStatusUnknown,
	}

	// Initialize alert manager
	if m.config.AlertsEnabled {
		m.alertManager = &AlertManager{
			rules:         make([]*AlertRule, 0),
			notifications: make([]AlertNotification, 0),
			thresholds:    make(map[string]*Threshold),
			enabled:       true,
		}
		m.setupDefaultAlertRules()
	}

	// Setup health checks
	m.setupHealthChecks()
}

// startMonitoring starts the monitoring goroutines
func (m *DataLayerMonitor) startMonitoring() {
	// Start metrics collection
	go m.metricsCollectionLoop()
	
	// Start health checking
	go m.healthCheckLoop()
	
	// Start alert evaluation
	if m.alertManager != nil {
		go m.alertEvaluationLoop()
	}
}

// setupHealthChecks configures health checks for all components
func (m *DataLayerMonitor) setupHealthChecks() {
	// State store health check
	m.healthChecker.checks["state_store"] = &HealthCheck{
		Name: "State Store",
		CheckFunc: func(ctx context.Context) error {
			// Test basic state store connectivity
			testKey := "health:check:state"
			testValue := map[string]interface{}{"timestamp": time.Now()}
			
			err := m.stateStore.Save(ctx, testKey, testValue, nil)
			if err != nil {
				return fmt.Errorf("state store save failed: %w", err)
			}
			
			var retrieved map[string]interface{}
			found, err := m.stateStore.Get(ctx, testKey, &retrieved)
			if err != nil {
				return fmt.Errorf("state store get failed: %w", err)
			}
			
			if !found {
				return fmt.Errorf("state store data not found")
			}
			
			// Clean up test data
			m.stateStore.Delete(ctx, testKey, nil)
			return nil
		},
	}

	// PubSub health check
	m.healthChecker.checks["pubsub"] = &HealthCheck{
		Name: "Pub/Sub",
		CheckFunc: func(ctx context.Context) error {
			// Test basic pub/sub connectivity
			testTopic := "health-check"
			testData := map[string]interface{}{
				"type":      "health_check",
				"timestamp": time.Now(),
			}
			
			return m.pubsub.PublishEvent(ctx, testTopic, testData, map[string]string{
				"health_check": "true",
			})
		},
	}

	// Secrets health check
	m.healthChecker.checks["secrets"] = &HealthCheck{
		Name: "Secrets Store",
		CheckFunc: func(ctx context.Context) error {
			// Test secrets store connectivity
			_, err := m.secrets.ValidateServiceSecrets(ctx, "health-check")
			// It's okay if the service doesn't exist, we just want to test connectivity
			if err != nil && !strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("secrets store validation failed: %w", err)
			}
			return nil
		},
	}

	// Transaction manager health check
	if m.transactionMgr != nil {
		m.healthChecker.checks["transaction_manager"] = &HealthCheck{
			Name: "Transaction Manager",
			CheckFunc: func(ctx context.Context) error {
				// Test transaction manager basic functionality
				metrics := m.transactionMgr.GetMetrics()
				if metrics == nil {
					return fmt.Errorf("transaction manager metrics not available")
				}
				return nil
			},
		}
	}
}

// setupDefaultAlertRules configures default alert rules
func (m *DataLayerMonitor) setupDefaultAlertRules() {
	// High error rate alert
	m.alertManager.rules = append(m.alertManager.rules, &AlertRule{
		Name:        "HighErrorRate",
		Condition:   "error_rate > threshold",
		Threshold:   5.0, // 5% error rate
		Duration:    5 * time.Minute,
		Severity:    AlertSeverityWarning,
		Description: "High error rate detected in data layer operations",
		Tags:        map[string]string{"component": "data_layer"},
		EvaluateFunc: func(metrics interface{}) bool {
			// Implementation would check actual error rates
			return false // Placeholder
		},
	})

	// High latency alert
	m.alertManager.rules = append(m.alertManager.rules, &AlertRule{
		Name:        "HighLatency",
		Condition:   "avg_latency > threshold",
		Threshold:   1000.0, // 1 second
		Duration:    2 * time.Minute,
		Severity:    AlertSeverityWarning,
		Description: "High latency detected in data layer operations",
		Tags:        map[string]string{"component": "data_layer"},
		EvaluateFunc: func(metrics interface{}) bool {
			// Implementation would check actual latencies
			return false // Placeholder
		},
	})

	// Component unhealthy alert
	m.alertManager.rules = append(m.alertManager.rules, &AlertRule{
		Name:        "ComponentUnhealthy",
		Condition:   "health_status != healthy",
		Threshold:   0.0,
		Duration:    1 * time.Minute,
		Severity:    AlertSeverityCritical,
		Description: "Data layer component is unhealthy",
		Tags:        map[string]string{"component": "data_layer"},
		EvaluateFunc: func(metrics interface{}) bool {
			// Implementation would check actual health status
			return false // Placeholder
		},
	})

	// Set default thresholds
	m.alertManager.thresholds["error_rate"] = &Threshold{
		Warning:  2.0,
		Critical: 10.0,
		Unit:     "percent",
	}
	
	m.alertManager.thresholds["latency"] = &Threshold{
		Warning:  500.0,
		Critical: 2000.0,
		Unit:     "milliseconds",
	}
	
	m.alertManager.thresholds["cache_hit_rate"] = &Threshold{
		Warning:  70.0,
		Critical: 50.0,
		Unit:     "percent",
	}
}

// metricsCollectionLoop periodically collects metrics from all components
func (m *DataLayerMonitor) metricsCollectionLoop() {
	ticker := time.NewTicker(m.config.MetricsInterval)
	defer ticker.Stop()

	for range ticker.C {
		if !m.enabled {
			return
		}
		
		m.collectAllMetrics()
	}
}

// healthCheckLoop periodically runs health checks on all components
func (m *DataLayerMonitor) healthCheckLoop() {
	ticker := time.NewTicker(m.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		if !m.enabled {
			return
		}
		
		m.runAllHealthChecks()
	}
}

// alertEvaluationLoop periodically evaluates alert rules
func (m *DataLayerMonitor) alertEvaluationLoop() {
	ticker := time.NewTicker(30 * time.Second) // Check alerts every 30 seconds
	defer ticker.Stop()

	for range ticker.C {
		if !m.enabled || m.alertManager == nil {
			return
		}
		
		m.evaluateAlertRules()
	}
}

// collectAllMetrics collects metrics from all data layer components
func (m *DataLayerMonitor) collectAllMetrics() {
	m.aggregator.mu.Lock()
	defer m.aggregator.mu.Unlock()

	// Collect state store metrics
	if m.stateStore != nil {
		m.aggregator.stateStoreMetrics = m.stateStore.GetMetrics()
	}

	// Collect pub/sub metrics
	if m.pubsub != nil {
		m.aggregator.pubsubMetrics = m.pubsub.GetMetrics()
	}

	// Collect secrets metrics
	if m.secrets != nil {
		m.aggregator.secretsMetrics = m.secrets.GetMetrics()
	}

	// Collect transaction metrics
	if m.transactionMgr != nil {
		m.aggregator.transactionMetrics = m.transactionMgr.GetMetrics()
	}

	// Collect system metrics
	m.aggregator.systemMetrics = m.collectSystemMetrics()

	m.aggregator.lastCollection = time.Now()

	// Log metrics if enabled
	if m.config.LogMetrics {
		m.logMetrics()
	}
}

// runAllHealthChecks executes all configured health checks
func (m *DataLayerMonitor) runAllHealthChecks() {
	m.healthChecker.mu.Lock()
	defer m.healthChecker.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), m.config.HealthCheckTimeout)
	defer cancel()

	overallHealthy := true

	for name, check := range m.healthChecker.checks {
		startTime := time.Now()
		
		err := check.CheckFunc(ctx)
		duration := time.Since(startTime)
		
		check.LastCheck = time.Now()
		check.Duration = duration
		
		if err != nil {
			check.Status = HealthStatusUnhealthy
			check.Error = err.Error()
			overallHealthy = false
		} else {
			check.Status = HealthStatusHealthy
			check.Error = ""
		}
		
		m.healthChecker.checks[name] = check
	}

	if overallHealthy {
		m.healthChecker.overallStatus = HealthStatusHealthy
	} else {
		m.healthChecker.overallStatus = HealthStatusUnhealthy
	}

	m.healthChecker.lastCheck = time.Now()
}

// evaluateAlertRules evaluates all configured alert rules
func (m *DataLayerMonitor) evaluateAlertRules() {
	m.alertManager.mu.Lock()
	defer m.alertManager.mu.Unlock()

	for _, rule := range m.alertManager.rules {
		if rule.EvaluateFunc != nil {
			shouldAlert := rule.EvaluateFunc(m.aggregator)
			
			if shouldAlert {
				// Create alert notification
				notification := AlertNotification{
					ID:        fmt.Sprintf("%s-%d", rule.Name, time.Now().UnixNano()),
					Rule:      rule.Name,
					Severity:  rule.Severity,
					Message:   rule.Description,
					Timestamp: time.Now(),
					Resolved:  false,
					Context:   make(map[string]interface{}),
				}
				
				m.alertManager.notifications = append(m.alertManager.notifications, notification)
			}
		}
	}
}

// collectSystemMetrics gathers system-level performance metrics
func (m *DataLayerMonitor) collectSystemMetrics() *SystemMetrics {
	// In a real implementation, this would collect actual system metrics
	// For now, return mock data
	return &SystemMetrics{
		CPUUsage:           15.5,
		MemoryUsage:        67.8,
		DiskUsage:          23.4,
		NetworkIOBytes:     1024000,
		ConnectionPoolSize: 25,
		Uptime:             24 * time.Hour,
		LastUpdated:        time.Now(),
	}
}

// logMetrics logs current metrics to the application log
func (m *DataLayerMonitor) logMetrics() {
	summary := m.GetMetricsSummary()
	summaryJSON, _ := json.Marshal(summary)
	
	// This would typically use a structured logger like slog
	fmt.Printf("[METRICS] %s\n", string(summaryJSON))
}

// Public API methods

// GetMetricsSummary returns a comprehensive summary of all metrics
func (m *DataLayerMonitor) GetMetricsSummary() map[string]interface{} {
	m.aggregator.mu.RLock()
	defer m.aggregator.mu.RUnlock()

	return map[string]interface{}{
		"state_store":         m.aggregator.stateStoreMetrics,
		"pubsub":             m.aggregator.pubsubMetrics,
		"secrets":            m.aggregator.secretsMetrics,
		"transactions":       m.aggregator.transactionMetrics,
		"system":             m.aggregator.systemMetrics,
		"custom":             m.aggregator.customMetrics,
		"last_collection":    m.aggregator.lastCollection,
		"collection_interval": m.aggregator.collectionInterval,
	}
}

// GetHealthStatus returns the current health status of all components
func (m *DataLayerMonitor) GetHealthStatus() map[string]interface{} {
	m.healthChecker.mu.RLock()
	defer m.healthChecker.mu.RUnlock()

	checks := make(map[string]*HealthCheck)
	for name, check := range m.healthChecker.checks {
		checks[name] = check
	}

	return map[string]interface{}{
		"overall_status": m.healthChecker.overallStatus,
		"last_check":     m.healthChecker.lastCheck,
		"checks":         checks,
	}
}

// GetActiveAlerts returns all currently active alerts
func (m *DataLayerMonitor) GetActiveAlerts() []AlertNotification {
	if m.alertManager == nil {
		return []AlertNotification{}
	}

	m.alertManager.mu.RLock()
	defer m.alertManager.mu.RUnlock()

	var activeAlerts []AlertNotification
	for _, notification := range m.alertManager.notifications {
		if !notification.Resolved {
			activeAlerts = append(activeAlerts, notification)
		}
	}

	return activeAlerts
}

// AddCustomMetric adds a custom metric to be tracked
func (m *DataLayerMonitor) AddCustomMetric(name string, value interface{}) {
	if !m.enabled {
		return
	}

	m.aggregator.mu.Lock()
	defer m.aggregator.mu.Unlock()

	m.aggregator.customMetrics[name] = value
}

// SetEnabled enables or disables monitoring
func (m *DataLayerMonitor) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// Shutdown gracefully shuts down the monitoring system
func (m *DataLayerMonitor) Shutdown(ctx context.Context) error {
	m.enabled = false
	// Any cleanup logic would go here
	return nil
}