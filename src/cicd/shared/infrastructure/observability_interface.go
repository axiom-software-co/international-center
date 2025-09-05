package infrastructure

import (
	"context"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type ObservabilityStack interface {
	Deploy(ctx context.Context) (ObservabilityDeployment, error)
	ConfigureDataSources(ctx context.Context, deployment ObservabilityDeployment) error
	CreateDashboards(ctx context.Context) error
	ConfigureAlerts(ctx context.Context) error
	ValidateDeployment(ctx context.Context, deployment ObservabilityDeployment) error
	GetObservabilityEndpoints() map[string]string
}

type ObservabilityDeployment interface {
	GetMetricsEndpoint() pulumi.StringOutput
	GetLogsEndpoint() pulumi.StringOutput
	GetTracingEndpoint() pulumi.StringOutput
	GetDashboardEndpoint() pulumi.StringOutput
	GetAlertManagerEndpoint() pulumi.StringOutput
}

type ObservabilityConfiguration struct {
	Environment            string
	DeploymentType        string // "self-hosted", "cloud", "hybrid"
	MetricsProvider       string // "prometheus", "azure-monitor"
	LogsProvider          string // "loki", "azure-logs"
	TracingProvider       string // "jaeger", "tempo", "azure-insights"
	DashboardProvider     string // "grafana", "azure-dashboard"
	AlertingProvider      string // "alertmanager", "azure-alerts"
	RetentionDays         int
	SamplingRate          float64
	EnableCompliance      bool
	EnableAuditLogging    bool
	SecurityMonitoring    bool
	BusinessMetrics       bool
	InfrastructureMetrics bool
	ApplicationMetrics    bool
}

type MetricsConfiguration struct {
	Enabled           bool
	ScrapeInterval    string
	RetentionDays     int
	StorageSize       string
	AlertThresholds   MetricsThresholds
	CustomMetrics     []CustomMetric
}

type LoggingConfiguration struct {
	Enabled         bool
	LogLevel        string
	RetentionDays   int
	IndexPattern    string
	AlertRules      []LogAlertRule
	CompactionHours int
}

type TracingConfiguration struct {
	Enabled        bool
	SamplingRate   float64
	RetentionDays  int
	MaxTraceLength int
	TracingHeaders []string
}

type DashboardConfiguration struct {
	Enabled           bool
	AutoRefresh       string
	TimeRange         string
	Dashboards        []Dashboard
	AlertDashboards   []AlertDashboard
	CompliancePanels  bool
}

type AlertingConfiguration struct {
	Enabled              bool
	NotificationChannels []NotificationChannel
	AlertRules           []AlertRule
	EscalationPolicies   []EscalationPolicy
	SilenceRules         []SilenceRule
}

type MetricsThresholds struct {
	CPUUtilization     float64
	MemoryUtilization  float64
	DiskUtilization    float64
	ErrorRate          float64
	ResponseTime       float64
	ThroughputRPS      int
}

type CustomMetric struct {
	Name        string
	Type        string // "counter", "gauge", "histogram"
	Labels      []string
	Description string
	Query       string
}

type LogAlertRule struct {
	Name        string
	Query       string
	Threshold   int
	Window      string
	Severity    string
	Description string
}

type Dashboard struct {
	Name        string
	Category    string // "application", "infrastructure", "business", "security"
	Panels      []Panel
	TimeRange   string
	RefreshRate string
}

type Panel struct {
	Title       string
	Type        string // "graph", "stat", "table", "heatmap"
	Queries     []Query
	Thresholds  []Threshold
	Units       string
}

type Query struct {
	Expression string
	Legend     string
	Datasource string
}

type Threshold struct {
	Value float64
	Color string
	State string // "ok", "warning", "critical"
}

type AlertDashboard struct {
	Name       string
	AlertRules []AlertRule
	Severity   string
}

type AlertRule struct {
	Name        string
	Expression  string
	Duration    string
	Severity    string // "info", "warning", "critical"
	Description string
	Labels      map[string]string
	Annotations map[string]string
}

type NotificationChannel struct {
	Name     string
	Type     string // "email", "slack", "webhook", "pagerduty"
	Settings map[string]interface{}
}

type EscalationPolicy struct {
	Name    string
	Steps   []EscalationStep
	Timeout string
}

type EscalationStep struct {
	Wait     string
	Channels []string
}

type SilenceRule struct {
	Matchers []Matcher
	Duration string
	Creator  string
	Comment  string
}

type Matcher struct {
	Name    string
	Value   string
	IsRegex bool
}

type ObservabilityFactory interface {
	CreateObservabilityStack(ctx *pulumi.Context, config *config.Config, environment string) ObservabilityStack
}

func GetObservabilityConfiguration(environment string, config *config.Config) *ObservabilityConfiguration {
	switch environment {
	case "development":
		return &ObservabilityConfiguration{
			Environment:            "development",
			DeploymentType:        "self-hosted",
			MetricsProvider:       "prometheus",
			LogsProvider:          "loki",
			TracingProvider:       "jaeger",
			DashboardProvider:     "grafana",
			AlertingProvider:      "alertmanager",
			RetentionDays:         7,
			SamplingRate:          1.0, // 100% sampling in dev
			EnableCompliance:      false,
			EnableAuditLogging:    false,
			SecurityMonitoring:    false,
			BusinessMetrics:       false,
			InfrastructureMetrics: true,
			ApplicationMetrics:    true,
		}
	case "staging":
		return &ObservabilityConfiguration{
			Environment:            "staging",
			DeploymentType:        "cloud",
			MetricsProvider:       "prometheus",
			LogsProvider:          "loki",
			TracingProvider:       "tempo",
			DashboardProvider:     "grafana",
			AlertingProvider:      "grafana-alerting",
			RetentionDays:         30,
			SamplingRate:          0.1, // 10% sampling in staging
			EnableCompliance:      true,
			EnableAuditLogging:    true,
			SecurityMonitoring:    true,
			BusinessMetrics:       true,
			InfrastructureMetrics: true,
			ApplicationMetrics:    true,
		}
	case "production":
		return &ObservabilityConfiguration{
			Environment:            "production",
			DeploymentType:        "cloud",
			MetricsProvider:       "prometheus",
			LogsProvider:          "loki",
			TracingProvider:       "tempo",
			DashboardProvider:     "grafana",
			AlertingProvider:      "grafana-alerting",
			RetentionDays:         90,
			SamplingRate:          0.05, // 5% sampling in production
			EnableCompliance:      true,
			EnableAuditLogging:    true,
			SecurityMonitoring:    true,
			BusinessMetrics:       true,
			InfrastructureMetrics: true,
			ApplicationMetrics:    true,
		}
	default:
		return &ObservabilityConfiguration{
			Environment:            environment,
			DeploymentType:        "self-hosted",
			MetricsProvider:       "prometheus",
			LogsProvider:          "loki",
			TracingProvider:       "jaeger",
			DashboardProvider:     "grafana",
			AlertingProvider:      "alertmanager",
			RetentionDays:         14,
			SamplingRate:          0.5,
			EnableCompliance:      false,
			EnableAuditLogging:    false,
			SecurityMonitoring:    true,
			BusinessMetrics:       false,
			InfrastructureMetrics: true,
			ApplicationMetrics:    true,
		}
	}
}

func GetMetricsConfiguration(environment string) MetricsConfiguration {
	switch environment {
	case "development":
		return MetricsConfiguration{
			Enabled:        true,
			ScrapeInterval: "15s",
			RetentionDays:  7,
			StorageSize:    "10Gi",
			AlertThresholds: MetricsThresholds{
				CPUUtilization:    0.8,
				MemoryUtilization: 0.8,
				DiskUtilization:   0.9,
				ErrorRate:         0.1,
				ResponseTime:      5.0,
				ThroughputRPS:     100,
			},
		}
	case "staging":
		return MetricsConfiguration{
			Enabled:        true,
			ScrapeInterval: "15s",
			RetentionDays:  30,
			StorageSize:    "50Gi",
			AlertThresholds: MetricsThresholds{
				CPUUtilization:    0.75,
				MemoryUtilization: 0.75,
				DiskUtilization:   0.85,
				ErrorRate:         0.05,
				ResponseTime:      2.0,
				ThroughputRPS:     1000,
			},
		}
	case "production":
		return MetricsConfiguration{
			Enabled:        true,
			ScrapeInterval: "10s",
			RetentionDays:  90,
			StorageSize:    "200Gi",
			AlertThresholds: MetricsThresholds{
				CPUUtilization:    0.7,
				MemoryUtilization: 0.7,
				DiskUtilization:   0.8,
				ErrorRate:         0.01,
				ResponseTime:      1.0,
				ThroughputRPS:     10000,
			},
		}
	default:
		return MetricsConfiguration{
			Enabled:        true,
			ScrapeInterval: "30s",
			RetentionDays:  14,
			StorageSize:    "20Gi",
			AlertThresholds: MetricsThresholds{
				CPUUtilization:    0.8,
				MemoryUtilization: 0.8,
				DiskUtilization:   0.9,
				ErrorRate:         0.1,
				ResponseTime:      3.0,
				ThroughputRPS:     500,
			},
		}
	}
}

func GetLoggingConfiguration(environment string) LoggingConfiguration {
	switch environment {
	case "development":
		return LoggingConfiguration{
			Enabled:         true,
			LogLevel:        "debug",
			RetentionDays:   7,
			IndexPattern:    "dev-logs-*",
			CompactionHours: 24,
		}
	case "staging":
		return LoggingConfiguration{
			Enabled:         true,
			LogLevel:        "info",
			RetentionDays:   30,
			IndexPattern:    "staging-logs-*",
			CompactionHours: 12,
		}
	case "production":
		return LoggingConfiguration{
			Enabled:         true,
			LogLevel:        "warn",
			RetentionDays:   90,
			IndexPattern:    "prod-logs-*",
			CompactionHours: 6,
		}
	default:
		return LoggingConfiguration{
			Enabled:         true,
			LogLevel:        "info",
			RetentionDays:   14,
			IndexPattern:    "logs-*",
			CompactionHours: 24,
		}
	}
}

func GetTracingConfiguration(environment string) TracingConfiguration {
	switch environment {
	case "development":
		return TracingConfiguration{
			Enabled:        true,
			SamplingRate:   1.0,
			RetentionDays:  3,
			MaxTraceLength: 1000,
		}
	case "staging":
		return TracingConfiguration{
			Enabled:        true,
			SamplingRate:   0.1,
			RetentionDays:  7,
			MaxTraceLength: 2000,
		}
	case "production":
		return TracingConfiguration{
			Enabled:        true,
			SamplingRate:   0.05,
			RetentionDays:  30,
			MaxTraceLength: 5000,
		}
	default:
		return TracingConfiguration{
			Enabled:        true,
			SamplingRate:   0.5,
			RetentionDays:  7,
			MaxTraceLength: 1000,
		}
	}
}

func GetAlertingConfiguration(environment string) AlertingConfiguration {
	base := AlertingConfiguration{
		Enabled: true,
		NotificationChannels: []NotificationChannel{
			{
				Name: "email-alerts",
				Type: "email",
				Settings: map[string]interface{}{
					"addresses": []string{"alerts@international-center.com"},
				},
			},
		},
	}

	switch environment {
	case "development":
		base.AlertRules = []AlertRule{
			{
				Name:        "high-error-rate-dev",
				Expression:  "rate(http_requests_total{code=~\"5..\"}[5m]) > 0.1",
				Duration:    "5m",
				Severity:    "warning",
				Description: "High error rate in development",
			},
		}
	case "staging":
		base.AlertRules = []AlertRule{
			{
				Name:        "high-error-rate-staging",
				Expression:  "rate(http_requests_total{code=~\"5..\"}[5m]) > 0.05",
				Duration:    "3m",
				Severity:    "warning",
				Description: "High error rate in staging",
			},
			{
				Name:        "service-down-staging",
				Expression:  "up == 0",
				Duration:    "1m",
				Severity:    "critical",
				Description: "Service down in staging",
			},
		}
	case "production":
		base.AlertRules = []AlertRule{
			{
				Name:        "critical-error-rate-prod",
				Expression:  "rate(http_requests_total{code=~\"5..\"}[5m]) > 0.01",
				Duration:    "1m",
				Severity:    "critical",
				Description: "Critical error rate in production",
			},
			{
				Name:        "service-down-prod",
				Expression:  "up == 0",
				Duration:    "30s",
				Severity:    "critical",
				Description: "Service down in production",
			},
			{
				Name:        "high-memory-usage-prod",
				Expression:  "container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.85",
				Duration:    "5m",
				Severity:    "warning",
				Description: "High memory usage in production",
			},
		}
		// Add PagerDuty for production
		base.NotificationChannels = append(base.NotificationChannels, NotificationChannel{
			Name: "pagerduty-critical",
			Type: "pagerduty",
			Settings: map[string]interface{}{
				"integrationKey": "",
				"severity":       "critical",
			},
		})
	}

	return base
}

// ObservabilityMetrics defines key metrics for environment-specific monitoring policies
type ObservabilityMetrics struct {
	MaxDataRetentionDays  int
	MaxSamplingRate      float64
	AlertResponseTime    string
	MonitoringCoverage   float64
	ComplianceRequired   bool
}

func GetObservabilityMetrics(environment string) ObservabilityMetrics {
	switch environment {
	case "development":
		return ObservabilityMetrics{
			MaxDataRetentionDays: 7,
			MaxSamplingRate:      1.0,
			AlertResponseTime:    "5m",
			MonitoringCoverage:   0.7,
			ComplianceRequired:   false,
		}
	case "staging":
		return ObservabilityMetrics{
			MaxDataRetentionDays: 30,
			MaxSamplingRate:      0.5,
			AlertResponseTime:    "2m",
			MonitoringCoverage:   0.9,
			ComplianceRequired:   true,
		}
	case "production":
		return ObservabilityMetrics{
			MaxDataRetentionDays: 365,
			MaxSamplingRate:      0.1,
			AlertResponseTime:    "30s",
			MonitoringCoverage:   0.99,
			ComplianceRequired:   true,
		}
	default:
		return ObservabilityMetrics{
			MaxDataRetentionDays: 14,
			MaxSamplingRate:      0.5,
			AlertResponseTime:    "3m",
			MonitoringCoverage:   0.8,
			ComplianceRequired:   false,
		}
	}
}