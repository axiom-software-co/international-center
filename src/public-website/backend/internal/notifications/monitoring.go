package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// MetricType represents different types of metrics
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeTimer     MetricType = "timer"
)

// Metric represents a single metric
type Metric struct {
	Name      string                 `json:"name"`
	Type      MetricType             `json:"type"`
	Value     float64                `json:"value"`
	Labels    map[string]string      `json:"labels,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Counter represents a monotonically increasing counter
type Counter struct {
	value int64
	name  string
	labels map[string]string
}

// NewCounter creates a new counter
func NewCounter(name string, labels map[string]string) *Counter {
	return &Counter{
		name:   name,
		labels: labels,
	}
}

// Inc increments the counter by 1
func (c *Counter) Inc() {
	atomic.AddInt64(&c.value, 1)
}

// Add adds a value to the counter
func (c *Counter) Add(value int64) {
	atomic.AddInt64(&c.value, value)
}

// Get returns the current value
func (c *Counter) Get() int64 {
	return atomic.LoadInt64(&c.value)
}

// Gauge represents a value that can go up and down
type Gauge struct {
	value int64
	name  string
	labels map[string]string
}

// NewGauge creates a new gauge
func NewGauge(name string, labels map[string]string) *Gauge {
	return &Gauge{
		name:   name,
		labels: labels,
	}
}

// Set sets the gauge value
func (g *Gauge) Set(value int64) {
	atomic.StoreInt64(&g.value, value)
}

// Inc increments the gauge by 1
func (g *Gauge) Inc() {
	atomic.AddInt64(&g.value, 1)
}

// Dec decrements the gauge by 1
func (g *Gauge) Dec() {
	atomic.AddInt64(&g.value, -1)
}

// Add adds a value to the gauge
func (g *Gauge) Add(value int64) {
	atomic.AddInt64(&g.value, value)
}

// Get returns the current value
func (g *Gauge) Get() int64 {
	return atomic.LoadInt64(&g.value)
}

// Timer measures durations and provides statistics
type Timer struct {
	durations []time.Duration
	mutex     sync.RWMutex
	name      string
	labels    map[string]string
}

// NewTimer creates a new timer
func NewTimer(name string, labels map[string]string) *Timer {
	return &Timer{
		durations: make([]time.Duration, 0),
		name:      name,
		labels:    labels,
	}
}

// Record records a duration
func (t *Timer) Record(duration time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	t.durations = append(t.durations, duration)
	
	// Keep only recent measurements to prevent unbounded growth
	if len(t.durations) > 1000 {
		t.durations = t.durations[len(t.durations)-500:]
	}
}

// Start returns a function that records the duration when called
func (t *Timer) Start() func() {
	start := time.Now()
	return func() {
		t.Record(time.Since(start))
	}
}

// GetStats returns timer statistics
func (t *Timer) GetStats() map[string]interface{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	
	if len(t.durations) == 0 {
		return map[string]interface{}{
			"count": 0,
			"min":   0,
			"max":   0,
			"avg":   0,
		}
	}
	
	var total, min, max time.Duration
	min = t.durations[0]
	
	for _, d := range t.durations {
		total += d
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}
	
	avg := total / time.Duration(len(t.durations))
	
	return map[string]interface{}{
		"count": len(t.durations),
		"min":   min.Milliseconds(),
		"max":   max.Milliseconds(),
		"avg":   avg.Milliseconds(),
		"total": total.Milliseconds(),
	}
}

// MetricsCollector collects and manages metrics
type MetricsCollector struct {
	counters map[string]*Counter
	gauges   map[string]*Gauge
	timers   map[string]*Timer
	mutex    sync.RWMutex
	logger   *slog.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *slog.Logger) *MetricsCollector {
	return &MetricsCollector{
		counters: make(map[string]*Counter),
		gauges:   make(map[string]*Gauge),
		timers:   make(map[string]*Timer),
		logger:   logger,
	}
}

// GetOrCreateCounter gets or creates a counter
func (mc *MetricsCollector) GetOrCreateCounter(name string, labels map[string]string) *Counter {
	key := mc.buildKey(name, labels)
	
	mc.mutex.RLock()
	counter, exists := mc.counters[key]
	mc.mutex.RUnlock()
	
	if exists {
		return counter
	}
	
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	// Double-check after acquiring write lock
	if counter, exists := mc.counters[key]; exists {
		return counter
	}
	
	counter = NewCounter(name, labels)
	mc.counters[key] = counter
	return counter
}

// GetOrCreateGauge gets or creates a gauge
func (mc *MetricsCollector) GetOrCreateGauge(name string, labels map[string]string) *Gauge {
	key := mc.buildKey(name, labels)
	
	mc.mutex.RLock()
	gauge, exists := mc.gauges[key]
	mc.mutex.RUnlock()
	
	if exists {
		return gauge
	}
	
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	// Double-check after acquiring write lock
	if gauge, exists := mc.gauges[key]; exists {
		return gauge
	}
	
	gauge = NewGauge(name, labels)
	mc.gauges[key] = gauge
	return gauge
}

// GetOrCreateTimer gets or creates a timer
func (mc *MetricsCollector) GetOrCreateTimer(name string, labels map[string]string) *Timer {
	key := mc.buildKey(name, labels)
	
	mc.mutex.RLock()
	timer, exists := mc.timers[key]
	mc.mutex.RUnlock()
	
	if exists {
		return timer
	}
	
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	// Double-check after acquiring write lock
	if timer, exists := mc.timers[key]; exists {
		return timer
	}
	
	timer = NewTimer(name, labels)
	mc.timers[key] = timer
	return timer
}

// buildKey builds a unique key for metrics
func (mc *MetricsCollector) buildKey(name string, labels map[string]string) string {
	key := name
	for k, v := range labels {
		key += fmt.Sprintf("_%s:%s", k, v)
	}
	return key
}

// CollectMetrics returns all current metrics
func (mc *MetricsCollector) CollectMetrics() []*Metric {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	var metrics []*Metric
	now := time.Now()
	
	// Collect counters
	for _, counter := range mc.counters {
		metrics = append(metrics, &Metric{
			Name:      counter.name,
			Type:      MetricTypeCounter,
			Value:     float64(counter.Get()),
			Labels:    counter.labels,
			Timestamp: now,
		})
	}
	
	// Collect gauges
	for _, gauge := range mc.gauges {
		metrics = append(metrics, &Metric{
			Name:      gauge.name,
			Type:      MetricTypeGauge,
			Value:     float64(gauge.Get()),
			Labels:    gauge.labels,
			Timestamp: now,
		})
	}
	
	// Collect timers
	for _, timer := range mc.timers {
		stats := timer.GetStats()
		metrics = append(metrics, &Metric{
			Name:      timer.name,
			Type:      MetricTypeTimer,
			Value:     0, // Timers don't have a single value
			Labels:    timer.labels,
			Timestamp: now,
			Metadata:  stats,
		})
	}
	
	return metrics
}

// NotificationMetrics tracks notification system metrics
type NotificationMetrics struct {
	collector *MetricsCollector
	
	// Message metrics
	messagesPublished *Counter
	messagesConsumed  *Counter
	messagesFailed    *Counter
	messagesRetried   *Counter
	
	// Processing metrics
	processingTime    *Timer
	queueDepth        *Gauge
	activeWorkers     *Gauge
	
	// Service metrics
	emailsSent        *Counter
	smssSent          *Counter
	slackMessagesSent *Counter
	
	// Error metrics
	databaseErrors    *Counter
	messageQueueErrors *Counter
	externalAPIErrors  *Counter
	
	// Circuit breaker metrics
	circuitBreakerTrips *Counter
	circuitBreakerState *Gauge
	
	logger *slog.Logger
}

// NewNotificationMetrics creates notification system metrics
func NewNotificationMetrics(logger *slog.Logger) *NotificationMetrics {
	collector := NewMetricsCollector(logger)
	
	return &NotificationMetrics{
		collector: collector,
		
		messagesPublished: collector.GetOrCreateCounter("messages_published_total", nil),
		messagesConsumed:  collector.GetOrCreateCounter("messages_consumed_total", nil),
		messagesFailed:    collector.GetOrCreateCounter("messages_failed_total", nil),
		messagesRetried:   collector.GetOrCreateCounter("messages_retried_total", nil),
		
		processingTime: collector.GetOrCreateTimer("message_processing_duration", nil),
		queueDepth:     collector.GetOrCreateGauge("queue_depth", nil),
		activeWorkers:  collector.GetOrCreateGauge("active_workers", nil),
		
		emailsSent:        collector.GetOrCreateCounter("emails_sent_total", nil),
		smssSent:          collector.GetOrCreateCounter("sms_sent_total", nil),
		slackMessagesSent: collector.GetOrCreateCounter("slack_messages_sent_total", nil),
		
		databaseErrors:     collector.GetOrCreateCounter("database_errors_total", nil),
		messageQueueErrors: collector.GetOrCreateCounter("message_queue_errors_total", nil),
		externalAPIErrors:  collector.GetOrCreateCounter("external_api_errors_total", nil),
		
		circuitBreakerTrips: collector.GetOrCreateCounter("circuit_breaker_trips_total", nil),
		circuitBreakerState: collector.GetOrCreateGauge("circuit_breaker_state", nil),
		
		logger: logger,
	}
}

// RecordMessagePublished records a published message
func (nm *NotificationMetrics) RecordMessagePublished() {
	nm.messagesPublished.Inc()
}

// RecordMessageConsumed records a consumed message
func (nm *NotificationMetrics) RecordMessageConsumed() {
	nm.messagesConsumed.Inc()
}

// RecordMessageFailed records a failed message
func (nm *NotificationMetrics) RecordMessageFailed() {
	nm.messagesFailed.Inc()
}

// RecordMessageRetried records a retried message
func (nm *NotificationMetrics) RecordMessageRetried() {
	nm.messagesRetried.Inc()
}

// RecordProcessingTime records message processing time
func (nm *NotificationMetrics) RecordProcessingTime(duration time.Duration) {
	nm.processingTime.Record(duration)
}

// StartProcessingTimer starts a processing timer
func (nm *NotificationMetrics) StartProcessingTimer() func() {
	return nm.processingTime.Start()
}

// SetQueueDepth sets the current queue depth
func (nm *NotificationMetrics) SetQueueDepth(depth int64) {
	nm.queueDepth.Set(depth)
}

// SetActiveWorkers sets the number of active workers
func (nm *NotificationMetrics) SetActiveWorkers(workers int64) {
	nm.activeWorkers.Set(workers)
}

// RecordEmailSent records a sent email
func (nm *NotificationMetrics) RecordEmailSent() {
	nm.emailsSent.Inc()
}

// RecordSMSSent records a sent SMS
func (nm *NotificationMetrics) RecordSMSSent() {
	nm.smssSent.Inc()
}

// RecordSlackMessageSent records a sent Slack message
func (nm *NotificationMetrics) RecordSlackMessageSent() {
	nm.slackMessagesSent.Inc()
}

// RecordDatabaseError records a database error
func (nm *NotificationMetrics) RecordDatabaseError() {
	nm.databaseErrors.Inc()
}

// RecordMessageQueueError records a message queue error
func (nm *NotificationMetrics) RecordMessageQueueError() {
	nm.messageQueueErrors.Inc()
}

// RecordExternalAPIError records an external API error
func (nm *NotificationMetrics) RecordExternalAPIError() {
	nm.externalAPIErrors.Inc()
}

// RecordCircuitBreakerTrip records a circuit breaker trip
func (nm *NotificationMetrics) RecordCircuitBreakerTrip() {
	nm.circuitBreakerTrips.Inc()
}

// SetCircuitBreakerState sets the circuit breaker state
func (nm *NotificationMetrics) SetCircuitBreakerState(state CircuitBreakerState) {
	nm.circuitBreakerState.Set(int64(state))
}

// GetAllMetrics returns all collected metrics
func (nm *NotificationMetrics) GetAllMetrics() []*Metric {
	return nm.collector.CollectMetrics()
}

// LogMetrics logs current metrics to structured logger
func (nm *NotificationMetrics) LogMetrics(ctx context.Context) {
	metrics := nm.GetAllMetrics()
	
	for _, metric := range metrics {
		nm.logger.InfoContext(ctx, "Metric",
			"name", metric.Name,
			"type", string(metric.Type),
			"value", metric.Value,
			"labels", metric.Labels,
			"metadata", metric.Metadata)
	}
}

// MetricsExporter exports metrics in various formats
type MetricsExporter struct {
	collector *MetricsCollector
	logger    *slog.Logger
}

// NewMetricsExporter creates a new metrics exporter
func NewMetricsExporter(collector *MetricsCollector, logger *slog.Logger) *MetricsExporter {
	return &MetricsExporter{
		collector: collector,
		logger:    logger,
	}
}

// ExportJSON exports metrics as JSON
func (me *MetricsExporter) ExportJSON() ([]byte, error) {
	metrics := me.collector.CollectMetrics()
	return json.MarshalIndent(metrics, "", "  ")
}

// ExportPrometheusFormat exports metrics in Prometheus format
func (me *MetricsExporter) ExportPrometheusFormat() string {
	metrics := me.collector.CollectMetrics()
	var output string
	
	for _, metric := range metrics {
		// Build labels string
		labelString := ""
		if len(metric.Labels) > 0 {
			var labels []string
			for k, v := range metric.Labels {
				labels = append(labels, fmt.Sprintf(`%s="%s"`, k, v))
			}
			labelString = fmt.Sprintf("{%s}", fmt.Sprintf("%s", labels))
		}
		
		metricName := fmt.Sprintf("notification_%s", metric.Name)
		
		switch metric.Type {
		case MetricTypeCounter:
			output += fmt.Sprintf("# TYPE %s counter\n", metricName)
			output += fmt.Sprintf("%s%s %.0f\n", metricName, labelString, metric.Value)
		case MetricTypeGauge:
			output += fmt.Sprintf("# TYPE %s gauge\n", metricName)
			output += fmt.Sprintf("%s%s %.0f\n", metricName, labelString, metric.Value)
		case MetricTypeTimer:
			if metric.Metadata != nil {
				output += fmt.Sprintf("# TYPE %s_duration_milliseconds summary\n", metricName)
				if count, ok := metric.Metadata["count"]; ok {
					output += fmt.Sprintf("%s_duration_milliseconds_count%s %.0f\n", 
						metricName, labelString, count)
				}
				if total, ok := metric.Metadata["total"]; ok {
					output += fmt.Sprintf("%s_duration_milliseconds_sum%s %.0f\n", 
						metricName, labelString, total)
				}
			}
		}
	}
	
	return output
}

// StartPeriodicExport starts periodic metric export
func (me *MetricsExporter) StartPeriodicExport(ctx context.Context, interval time.Duration, exportFunc func([]byte)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			data, err := me.ExportJSON()
			if err != nil {
				me.logger.Error("Failed to export metrics", "error", err)
				continue
			}
			
			exportFunc(data)
		}
	}
}