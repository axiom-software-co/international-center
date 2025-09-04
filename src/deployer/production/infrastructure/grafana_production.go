package infrastructure

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-azure-native-sdk/resources"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type GrafanaProductionStack struct {
	resourceGroup         *resources.ResourceGroup
	grafanaUrl           pulumi.StringOutput
	prometheusUrl        pulumi.StringOutput
	lokiUrl              pulumi.StringOutput
	tempoUrl             pulumi.StringOutput
	grafanaApiKey        string
	prometheusApiKey     string
	lokiApiKey           string
	tempoApiKey          string
	dashboardConfigs     map[string]interface{}
	alertConfigs         map[string]interface{}
	complianceConfigs    map[string]interface{}
}

func NewGrafanaProductionStack(resourceGroup *resources.ResourceGroup) *GrafanaProductionStack {
	return &GrafanaProductionStack{
		resourceGroup:     resourceGroup,
		dashboardConfigs:  make(map[string]interface{}),
		alertConfigs:      make(map[string]interface{}),
		complianceConfigs: make(map[string]interface{}),
	}
}

func (stack *GrafanaProductionStack) Deploy(ctx *pulumi.Context) error {
	if err := stack.configureGrafanaCloudProduction(ctx); err != nil {
		return fmt.Errorf("failed to configure Grafana Cloud production: %w", err)
	}

	if err := stack.setupProductionDataSources(ctx); err != nil {
		return fmt.Errorf("failed to setup production data sources: %w", err)
	}

	if err := stack.createProductionDashboards(ctx); err != nil {
		return fmt.Errorf("failed to create production dashboards: %w", err)
	}

	if err := stack.configureProductionAlerts(ctx); err != nil {
		return fmt.Errorf("failed to configure production alerts: %w", err)
	}

	if err := stack.setupComplianceMonitoring(ctx); err != nil {
		return fmt.Errorf("failed to setup compliance monitoring: %w", err)
	}

	return nil
}

func (stack *GrafanaProductionStack) configureGrafanaCloudProduction(ctx *pulumi.Context) error {
	stack.grafanaUrl = pulumi.String("https://international-center-production.grafana.net").ToStringOutput()
	stack.prometheusUrl = pulumi.String("https://prometheus-prod-13-prod-eu-west-0.grafana.net").ToStringOutput()
	stack.lokiUrl = pulumi.String("https://logs-prod-eu-west-0.grafana.net").ToStringOutput()
	stack.tempoUrl = pulumi.String("https://tempo-prod-04-prod-eu-west-0.grafana.net").ToStringOutput()

	stack.grafanaApiKey = "" // Retrieved from Key Vault
	stack.prometheusApiKey = "" // Retrieved from Key Vault
	stack.lokiApiKey = "" // Retrieved from Key Vault
	stack.tempoApiKey = "" // Retrieved from Key Vault

	return nil
}

func (stack *GrafanaProductionStack) setupProductionDataSources(ctx *pulumi.Context) error {
	prometheusConfig := map[string]interface{}{
		"name":      "Prometheus-Production",
		"type":      "prometheus",
		"url":       stack.prometheusUrl,
		"access":    "proxy",
		"basicAuth": false,
		"jsonData": map[string]interface{}{
			"httpMethod":     "GET",
			"manageAlerts":   true,
			"prometheusType": "Mimir",
			"prometheusVersion": "2.40.0",
		},
		"secureJsonData": map[string]interface{}{
			"basicAuthPassword": stack.prometheusApiKey,
		},
	}

	lokiConfig := map[string]interface{}{
		"name":      "Loki-Production",
		"type":      "loki",
		"url":       stack.lokiUrl,
		"access":    "proxy",
		"basicAuth": false,
		"jsonData": map[string]interface{}{
			"maxLines":         1000,
			"manageAlerts":     true,
			"derivedFields": []map[string]interface{}{
				{
					"datasourceUid": "tempo-production",
					"matcherRegex":  "trace_id=(\\w+)",
					"name":          "TraceID",
					"url":           "${__value.raw}",
				},
			},
		},
		"secureJsonData": map[string]interface{}{
			"basicAuthPassword": stack.lokiApiKey,
		},
	}

	tempoConfig := map[string]interface{}{
		"name":      "Tempo-Production",
		"type":      "tempo",
		"uid":       "tempo-production",
		"url":       stack.tempoUrl,
		"access":    "proxy",
		"basicAuth": false,
		"jsonData": map[string]interface{}{
			"tracesToLogsV2": map[string]interface{}{
				"datasourceUid": "loki-production",
				"tags": []map[string]interface{}{
					{"key": "service.name", "value": "service_name"},
					{"key": "service.namespace", "value": "namespace"},
				},
			},
			"tracesToMetrics": map[string]interface{}{
				"datasourceUid": "prometheus-production",
				"tags": []map[string]interface{}{
					{"key": "service.name", "value": "service_name"},
					{"key": "job", "value": "job"},
				},
			},
		},
		"secureJsonData": map[string]interface{}{
			"basicAuthPassword": stack.tempoApiKey,
		},
	}

	stack.dashboardConfigs["prometheus"] = prometheusConfig
	stack.dashboardConfigs["loki"] = lokiConfig
	stack.dashboardConfigs["tempo"] = tempoConfig

	return nil
}

func (stack *GrafanaProductionStack) createProductionDashboards(ctx *pulumi.Context) error {
	if err := stack.createApplicationProductionDashboard(); err != nil {
		return err
	}

	if err := stack.createInfrastructureProductionDashboard(); err != nil {
		return err
	}

	if err := stack.createBusinessMetricsProductionDashboard(); err != nil {
		return err
	}

	if err := stack.createSecurityProductionDashboard(); err != nil {
		return err
	}

	if err := stack.createComplianceDashboard(); err != nil {
		return err
	}

	return nil
}

func (stack *GrafanaProductionStack) createApplicationProductionDashboard() error {
	dashboard := map[string]interface{}{
		"dashboard": map[string]interface{}{
			"id":    nil,
			"title": "International Center - Application Metrics (Production)",
			"tags":  []string{"production", "application", "apis"},
			"timezone": "UTC",
			"panels": []map[string]interface{}{
				{
					"title": "Request Rate by Service",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(rate(http_requests_total{environment=\"production\"}[5m])) by (service)",
							"legendFormat": "{{service}}",
						},
					},
					"fieldConfig": map[string]interface{}{
						"defaults": map[string]interface{}{
							"unit": "reqps",
						},
					},
				},
				{
					"title": "Response Time P99",
					"type":  "stat",
					"targets": []map[string]interface{}{
						{
							"expr": "histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{environment=\"production\"}[5m])) by (le, service))",
							"legendFormat": "{{service}}",
						},
					},
					"fieldConfig": map[string]interface{}{
						"defaults": map[string]interface{}{
							"unit": "s",
						},
					},
				},
				{
					"title": "Error Rate",
					"type":  "stat",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(rate(http_requests_total{environment=\"production\", code=~\"4..|5..\"}[5m])) / sum(rate(http_requests_total{environment=\"production\"}[5m]))",
							"legendFormat": "Error Rate",
						},
					},
					"fieldConfig": map[string]interface{}{
						"defaults": map[string]interface{}{
							"unit": "percentunit",
							"thresholds": map[string]interface{}{
								"steps": []map[string]interface{}{
									{"color": "green", "value": 0},
									{"color": "yellow", "value": 0.01},
									{"color": "red", "value": 0.05},
								},
							},
						},
					},
				},
				{
					"title": "Active Container Replicas",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(kube_deployment_status_replicas_available{environment=\"production\"}) by (deployment)",
							"legendFormat": "{{deployment}}",
						},
					},
				},
				{
					"title": "Database Connection Pool",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(postgresql_connections{environment=\"production\", state=\"active\"}) by (database)",
							"legendFormat": "Active - {{database}}",
						},
						{
							"expr": "sum(postgresql_connections{environment=\"production\", state=\"idle\"}) by (database)",
							"legendFormat": "Idle - {{database}}",
						},
					},
				},
				{
					"title": "Redis Operations",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(rate(redis_commands_processed_total{environment=\"production\"}[5m])) by (cmd)",
							"legendFormat": "{{cmd}}",
						},
					},
				},
			},
			"time": map[string]interface{}{
				"from": "now-1h",
				"to":   "now",
			},
			"refresh": "15s",
		},
	}

	stack.dashboardConfigs["application-production"] = dashboard
	return nil
}

func (stack *GrafanaProductionStack) createInfrastructureProductionDashboard() error {
	dashboard := map[string]interface{}{
		"dashboard": map[string]interface{}{
			"id":    nil,
			"title": "International Center - Infrastructure (Production)",
			"tags":  []string{"production", "infrastructure", "azure"},
			"timezone": "UTC",
			"panels": []map[string]interface{}{
				{
					"title": "Container CPU Usage",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(rate(container_cpu_usage_seconds_total{environment=\"production\"}[5m])) by (container_name)",
							"legendFormat": "{{container_name}}",
						},
					},
				},
				{
					"title": "Container Memory Usage",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(container_memory_working_set_bytes{environment=\"production\"}) by (container_name)",
							"legendFormat": "{{container_name}}",
						},
					},
				},
				{
					"title": "Azure Container Apps Scaling",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(azure_container_apps_replica_count{environment=\"production\"}) by (app_name)",
							"legendFormat": "{{app_name}}",
						},
					},
				},
				{
					"title": "Database Performance",
					"type":  "timeseries", 
					"targets": []map[string]interface{}{
						{
							"expr": "postgresql_database_size_bytes{environment=\"production\"}",
							"legendFormat": "{{database}} Size",
						},
						{
							"expr": "rate(postgresql_stat_database_tup_inserted{environment=\"production\"}[5m])",
							"legendFormat": "{{database}} Inserts/sec",
						},
					},
				},
				{
					"title": "Storage Account Performance",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(azure_storage_transactions_total{environment=\"production\"}) by (operation_type)",
							"legendFormat": "{{operation_type}}",
						},
						{
							"expr": "sum(azure_storage_used_capacity_bytes{environment=\"production\"}) by (account_name)",
							"legendFormat": "{{account_name}}",
						},
					},
				},
			},
			"time": map[string]interface{}{
				"from": "now-6h",
				"to":   "now",
			},
			"refresh": "30s",
		},
	}

	stack.dashboardConfigs["infrastructure-production"] = dashboard
	return nil
}

func (stack *GrafanaProductionStack) createBusinessMetricsProductionDashboard() error {
	dashboard := map[string]interface{}{
		"dashboard": map[string]interface{}{
			"id":    nil,
			"title": "International Center - Business Metrics (Production)",
			"tags":  []string{"production", "business", "kpis"},
			"timezone": "UTC",
			"panels": []map[string]interface{}{
				{
					"title": "User Registrations (24h)",
					"type":  "stat",
					"targets": []map[string]interface{}{
						{
							"expr": "increase(user_registrations_total{environment=\"production\"}[24h])",
							"legendFormat": "Registrations",
						},
					},
					"fieldConfig": map[string]interface{}{
						"defaults": map[string]interface{}{
							"unit": "short",
						},
					},
				},
				{
					"title": "Content Creation Rate",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "rate(content_created_total{environment=\"production\"}[1h])",
							"legendFormat": "Content/hour",
						},
					},
				},
				{
					"title": "API Usage Distribution",
					"type":  "piechart",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(rate(http_requests_total{environment=\"production\"}[1h])) by (service)",
							"legendFormat": "{{service}}",
						},
					},
				},
				{
					"title": "Service Response Times",
					"type":  "bargauge",
					"targets": []map[string]interface{}{
						{
							"expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{environment=\"production\"}[5m])) by (le, service))",
							"legendFormat": "{{service}} P95",
						},
					},
				},
			},
			"time": map[string]interface{}{
				"from": "now-24h",
				"to":   "now",
			},
			"refresh": "2m",
		},
	}

	stack.dashboardConfigs["business-production"] = dashboard
	return nil
}

func (stack *GrafanaProductionStack) createSecurityProductionDashboard() error {
	dashboard := map[string]interface{}{
		"dashboard": map[string]interface{}{
			"id":    nil,
			"title": "International Center - Security Monitoring (Production)",
			"tags":  []string{"production", "security", "monitoring"},
			"timezone": "UTC",
			"panels": []map[string]interface{}{
				{
					"title": "Authentication Failures",
					"type":  "stat",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(rate(authentication_failures_total{environment=\"production\"}[5m]))",
							"legendFormat": "Auth Failures/sec",
						},
					},
					"fieldConfig": map[string]interface{}{
						"defaults": map[string]interface{}{
							"thresholds": map[string]interface{}{
								"steps": []map[string]interface{}{
									{"color": "green", "value": 0},
									{"color": "yellow", "value": 0.1},
									{"color": "red", "value": 1.0},
								},
							},
						},
					},
				},
				{
					"title": "Suspicious IP Activity",
					"type":  "table",
					"targets": []map[string]interface{}{
						{
							"expr": "topk(20, sum(rate(http_requests_total{environment=\"production\", code=~\"4..\"}[5m])) by (source_ip))",
							"legendFormat": "{{source_ip}}",
						},
					},
				},
				{
					"title": "Rate Limiting Events",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(rate(rate_limit_exceeded_total{environment=\"production\"}[5m])) by (endpoint)",
							"legendFormat": "{{endpoint}}",
						},
					},
				},
				{
					"title": "SSL Certificate Expiry",
					"type":  "stat",
					"targets": []map[string]interface{}{
						{
							"expr": "(ssl_certificate_expiry_seconds{environment=\"production\"} - time()) / 86400",
							"legendFormat": "Days until expiry",
						},
					},
					"fieldConfig": map[string]interface{}{
						"defaults": map[string]interface{}{
							"thresholds": map[string]interface{}{
								"steps": []map[string]interface{}{
									{"color": "red", "value": 0},
									{"color": "yellow", "value": 30},
									{"color": "green", "value": 90},
								},
							},
						},
					},
				},
			},
			"time": map[string]interface{}{
				"from": "now-2h",
				"to":   "now",
			},
			"refresh": "15s",
		},
	}

	stack.dashboardConfigs["security-production"] = dashboard
	return nil
}

func (stack *GrafanaProductionStack) createComplianceDashboard() error {
	dashboard := map[string]interface{}{
		"dashboard": map[string]interface{}{
			"id":    nil,
			"title": "International Center - Compliance Monitoring (Production)",
			"tags":  []string{"production", "compliance", "audit"},
			"timezone": "UTC",
			"panels": []map[string]interface{}{
				{
					"title": "Audit Event Volume",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(rate(audit_events_total{environment=\"production\"}[5m])) by (entity_type)",
							"legendFormat": "{{entity_type}}",
						},
					},
				},
				{
					"title": "Data Access Events",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(rate(data_access_events_total{environment=\"production\"}[5m])) by (access_level)",
							"legendFormat": "{{access_level}}",
						},
					},
				},
				{
					"title": "Backup Success Rate",
					"type":  "stat",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(backup_success_total{environment=\"production\"}) / sum(backup_attempts_total{environment=\"production\"})",
							"legendFormat": "Success Rate",
						},
					},
					"fieldConfig": map[string]interface{}{
						"defaults": map[string]interface{}{
							"unit": "percentunit",
							"thresholds": map[string]interface{}{
								"steps": []map[string]interface{}{
									{"color": "red", "value": 0.95},
									{"color": "yellow", "value": 0.98},
									{"color": "green", "value": 0.99},
								},
							},
						},
					},
				},
				{
					"title": "Compliance Violations",
					"type":  "stat",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(compliance_violations_total{environment=\"production\"})",
							"legendFormat": "Violations",
						},
					},
					"fieldConfig": map[string]interface{}{
						"defaults": map[string]interface{}{
							"thresholds": map[string]interface{}{
								"steps": []map[string]interface{}{
									{"color": "green", "value": 0},
									{"color": "yellow", "value": 1},
									{"color": "red", "value": 5},
								},
							},
						},
					},
				},
			},
		},
		"time": map[string]interface{}{
			"from": "now-24h",
			"to":   "now",
		},
		"refresh": "1m",
	}

	stack.complianceConfigs["compliance-production"] = dashboard
	return nil
}

func (stack *GrafanaProductionStack) configureProductionAlerts(ctx *pulumi.Context) error {
	alerts := []map[string]interface{}{
		{
			"alert": map[string]interface{}{
				"name":        "Critical Error Rate - Production",
				"message":     "Error rate exceeds 1% in production environment - immediate attention required",
				"frequency":   "10s",
				"conditions": []map[string]interface{}{
					{
						"query": map[string]interface{}{
							"queryType": "A",
							"refId":     "A",
						},
						"reducer": map[string]interface{}{
							"type": "last",
						},
						"evaluator": map[string]interface{}{
							"params": []float64{0.01},
							"type":   "gt",
						},
					},
				},
				"executionErrorState": "alerting",
				"noDataState":        "no_data",
				"for":               "2m",
			},
		},
		{
			"alert": map[string]interface{}{
				"name":        "Production Service Down",
				"message":     "One or more production services are down - critical incident",
				"frequency":   "5s",
				"conditions": []map[string]interface{}{
					{
						"query": map[string]interface{}{
							"queryType": "A",
							"refId":     "A",
						},
						"reducer": map[string]interface{}{
							"type": "last",
						},
						"evaluator": map[string]interface{}{
							"params": []float64{1},
							"type":   "lt",
						},
					},
				},
				"executionErrorState": "alerting",
				"noDataState":        "alerting",
				"for":               "30s",
			},
		},
		{
			"alert": map[string]interface{}{
				"name":        "Database Connection Issues - Production",
				"message":     "Database connections are failing in production - immediate investigation required",
				"frequency":   "15s",
				"conditions": []map[string]interface{}{
					{
						"query": map[string]interface{}{
							"queryType": "A",
							"refId":     "A",
						},
						"reducer": map[string]interface{}{
							"type": "last",
						},
						"evaluator": map[string]interface{}{
							"params": []float64{5},
							"type":   "gt",
						},
					},
				},
				"executionErrorState": "alerting",
				"noDataState":        "no_data",
				"for":               "1m",
			},
		},
		{
			"alert": map[string]interface{}{
				"name":        "High Memory Usage - Production",
				"message":     "Memory usage exceeds 85% in production containers",
				"frequency":   "30s",
				"conditions": []map[string]interface{}{
					{
						"query": map[string]interface{}{
							"queryType": "A",
							"refId":     "A",
						},
						"reducer": map[string]interface{}{
							"type": "last",
						},
						"evaluator": map[string]interface{}{
							"params": []float64{0.85},
							"type":   "gt",
						},
					},
				},
				"executionErrorState": "alerting",
				"noDataState":        "no_data",
				"for":               "5m",
			},
		},
	}

	for _, alert := range alerts {
		stack.alertConfigs[alert["alert"].(map[string]interface{})["name"].(string)] = alert
	}

	return nil
}

func (stack *GrafanaProductionStack) setupComplianceMonitoring(ctx *pulumi.Context) error {
	complianceAlerts := []map[string]interface{}{
		{
			"alert": map[string]interface{}{
				"name":        "Audit Event Loss - Production",
				"message":     "Audit events are being lost - compliance violation detected",
				"frequency":   "10s",
				"conditions": []map[string]interface{}{
					{
						"query": map[string]interface{}{
							"queryType": "A",
							"refId":     "A",
						},
						"reducer": map[string]interface{}{
							"type": "last",
						},
						"evaluator": map[string]interface{}{
							"params": []float64{1},
							"type":   "gt",
						},
					},
				},
				"executionErrorState": "alerting",
				"noDataState":        "alerting",
				"for":               "30s",
			},
		},
		{
			"alert": map[string]interface{}{
				"name":        "Backup Failure - Production",
				"message":     "Production backup has failed - compliance requirement at risk",
				"frequency":   "60s",
				"conditions": []map[string]interface{}{
					{
						"query": map[string]interface{}{
							"queryType": "A",
							"refId":     "A",
						},
						"reducer": map[string]interface{}{
							"type": "last",
						},
						"evaluator": map[string]interface{}{
							"params": []float64{1},
							"type":   "gt",
						},
					},
				},
				"executionErrorState": "alerting",
				"noDataState":        "no_data",
				"for":               "0s",
			},
		},
	}

	for _, alert := range complianceAlerts {
		stack.complianceConfigs[alert["alert"].(map[string]interface{})["name"].(string)] = alert
	}

	return nil
}

func (stack *GrafanaProductionStack) GetGrafanaUrl() pulumi.StringOutput {
	return stack.grafanaUrl
}

func (stack *GrafanaProductionStack) GetPrometheusUrl() pulumi.StringOutput {
	return stack.prometheusUrl
}

func (stack *GrafanaProductionStack) GetLokiUrl() pulumi.StringOutput {
	return stack.lokiUrl
}

func (stack *GrafanaProductionStack) GetTempoUrl() pulumi.StringOutput {
	return stack.tempoUrl
}

func (stack *GrafanaProductionStack) GetDashboardConfigs() map[string]interface{} {
	return stack.dashboardConfigs
}

func (stack *GrafanaProductionStack) GetAlertConfigs() map[string]interface{} {
	return stack.alertConfigs
}

func (stack *GrafanaProductionStack) GetComplianceConfigs() map[string]interface{} {
	return stack.complianceConfigs
}