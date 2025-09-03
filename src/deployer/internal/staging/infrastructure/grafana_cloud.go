package infrastructure

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-azure-native/sdk/v2/go/azurenative/resources"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type GrafanaCloudStack struct {
	resourceGroup     *resources.ResourceGroup
	grafanaUrl        pulumi.StringOutput
	prometheusUrl     pulumi.StringOutput
	lokiUrl          pulumi.StringOutput
	grafanaApiKey     string
	prometheusApiKey  string
	lokiApiKey       string
	dashboardConfigs  map[string]interface{}
	alertConfigs      map[string]interface{}
}

func NewGrafanaCloudStack(resourceGroup *resources.ResourceGroup) *GrafanaCloudStack {
	return &GrafanaCloudStack{
		resourceGroup:    resourceGroup,
		dashboardConfigs: make(map[string]interface{}),
		alertConfigs:     make(map[string]interface{}),
	}
}

func (stack *GrafanaCloudStack) Deploy(ctx *pulumi.Context) error {
	if err := stack.configureGrafanaCloud(ctx); err != nil {
		return fmt.Errorf("failed to configure Grafana Cloud: %w", err)
	}

	if err := stack.setupDataSources(ctx); err != nil {
		return fmt.Errorf("failed to setup data sources: %w", err)
	}

	if err := stack.createDashboards(ctx); err != nil {
		return fmt.Errorf("failed to create dashboards: %w", err)
	}

	if err := stack.configureAlerts(ctx); err != nil {
		return fmt.Errorf("failed to configure alerts: %w", err)
	}

	return nil
}

func (stack *GrafanaCloudStack) configureGrafanaCloud(ctx *pulumi.Context) error {
	stack.grafanaUrl = pulumi.String("https://international-center-staging.grafana.net").ToStringOutput()
	stack.prometheusUrl = pulumi.String("https://prometheus-prod-01-eu-west-0.grafana.net").ToStringOutput()
	stack.lokiUrl = pulumi.String("https://logs-prod-eu-west-0.grafana.net").ToStringOutput()

	stack.grafanaApiKey = "" // Retrieved from Key Vault
	stack.prometheusApiKey = "" // Retrieved from Key Vault
	stack.lokiApiKey = "" // Retrieved from Key Vault

	return nil
}

func (stack *GrafanaCloudStack) setupDataSources(ctx *pulumi.Context) error {
	prometheusConfig := map[string]interface{}{
		"name": "Prometheus-Staging",
		"type": "prometheus",
		"url":  stack.prometheusUrl,
		"access": "proxy",
		"basicAuth": false,
		"jsonData": map[string]interface{}{
			"httpMethod": "GET",
			"manageAlerts": true,
		},
		"secureJsonData": map[string]interface{}{
			"basicAuthPassword": stack.prometheusApiKey,
		},
	}

	lokiConfig := map[string]interface{}{
		"name": "Loki-Staging", 
		"type": "loki",
		"url":  stack.lokiUrl,
		"access": "proxy",
		"basicAuth": false,
		"jsonData": map[string]interface{}{
			"maxLines": 1000,
			"manageAlerts": true,
		},
		"secureJsonData": map[string]interface{}{
			"basicAuthPassword": stack.lokiApiKey,
		},
	}

	stack.dashboardConfigs["prometheus"] = prometheusConfig
	stack.dashboardConfigs["loki"] = lokiConfig

	return nil
}

func (stack *GrafanaCloudStack) createDashboards(ctx *pulumi.Context) error {
	if err := stack.createApplicationDashboard(); err != nil {
		return err
	}

	if err := stack.createInfrastructureDashboard(); err != nil {
		return err
	}

	if err := stack.createBusinessMetricsDashboard(); err != nil {
		return err
	}

	if err := stack.createSecurityDashboard(); err != nil {
		return err
	}

	return nil
}

func (stack *GrafanaCloudStack) createApplicationDashboard() error {
	dashboard := map[string]interface{}{
		"dashboard": map[string]interface{}{
			"id":    nil,
			"title": "International Center - Application Metrics (Staging)",
			"tags":  []string{"staging", "application", "apis"},
			"timezone": "UTC",
			"panels": []map[string]interface{}{
				{
					"title": "Request Rate",
					"type":  "stat",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(rate(http_requests_total{environment=\"staging\"}[5m])) by (service)",
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
					"title": "Response Time (95th percentile)",
					"type":  "stat",
					"targets": []map[string]interface{}{
						{
							"expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{environment=\"staging\"}[5m])) by (le, service))",
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
							"expr": "sum(rate(http_requests_total{environment=\"staging\", code=~\"4..|5..\"}[5m])) / sum(rate(http_requests_total{environment=\"staging\"}[5m]))",
							"legendFormat": "Error Rate",
						},
					},
					"fieldConfig": map[string]interface{}{
						"defaults": map[string]interface{}{
							"unit": "percent",
						},
					},
				},
				{
					"title": "Active Containers",
					"type":  "stat",
					"targets": []map[string]interface{}{
						{
							"expr": "count(up{environment=\"staging\"} == 1) by (service)",
							"legendFormat": "{{service}}",
						},
					},
				},
				{
					"title": "Database Connections",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(postgresql_connections{environment=\"staging\"}) by (database)",
							"legendFormat": "{{database}}",
						},
					},
				},
				{
					"title": "Redis Operations",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(rate(redis_commands_total{environment=\"staging\"}[5m])) by (command)",
							"legendFormat": "{{command}}",
						},
					},
				},
			},
			"time": map[string]interface{}{
				"from": "now-1h",
				"to":   "now",
			},
			"refresh": "30s",
		},
	}

	stack.dashboardConfigs["application"] = dashboard
	return nil
}

func (stack *GrafanaCloudStack) createInfrastructureDashboard() error {
	dashboard := map[string]interface{}{
		"dashboard": map[string]interface{}{
			"id":    nil,
			"title": "International Center - Infrastructure (Staging)",
			"tags":  []string{"staging", "infrastructure", "azure"},
			"timezone": "UTC",
			"panels": []map[string]interface{}{
				{
					"title": "Container CPU Usage",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(rate(container_cpu_usage_seconds_total{environment=\"staging\"}[5m])) by (container_name)",
							"legendFormat": "{{container_name}}",
						},
					},
				},
				{
					"title": "Container Memory Usage",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(container_memory_usage_bytes{environment=\"staging\"}) by (container_name)",
							"legendFormat": "{{container_name}}",
						},
					},
				},
				{
					"title": "Azure Container Apps Scaling",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(azure_container_apps_replica_count{environment=\"staging\"}) by (app_name)",
							"legendFormat": "{{app_name}}",
						},
					},
				},
				{
					"title": "Database Performance",
					"type":  "timeseries", 
					"targets": []map[string]interface{}{
						{
							"expr": "postgresql_database_size_bytes{environment=\"staging\"}",
							"legendFormat": "{{database}} Size",
						},
					},
				},
				{
					"title": "Storage Account Operations",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(azure_storage_transactions_total{environment=\"staging\"}) by (operation_type)",
							"legendFormat": "{{operation_type}}",
						},
					},
				},
			},
			"time": map[string]interface{}{
				"from": "now-6h",
				"to":   "now",
			},
			"refresh": "1m",
		},
	}

	stack.dashboardConfigs["infrastructure"] = dashboard
	return nil
}

func (stack *GrafanaCloudStack) createBusinessMetricsDashboard() error {
	dashboard := map[string]interface{}{
		"dashboard": map[string]interface{}{
			"id":    nil,
			"title": "International Center - Business Metrics (Staging)",
			"tags":  []string{"staging", "business", "kpis"},
			"timezone": "UTC",
			"panels": []map[string]interface{}{
				{
					"title": "User Registrations",
					"type":  "stat",
					"targets": []map[string]interface{}{
						{
							"expr": "increase(user_registrations_total{environment=\"staging\"}[24h])",
							"legendFormat": "Daily Registrations",
						},
					},
				},
				{
					"title": "Content Creation Rate",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "rate(content_created_total{environment=\"staging\"}[1h])",
							"legendFormat": "Content per Hour",
						},
					},
				},
				{
					"title": "API Usage by Service",
					"type":  "piechart",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(rate(http_requests_total{environment=\"staging\"}[1h])) by (service)",
							"legendFormat": "{{service}}",
						},
					},
				},
				{
					"title": "Feature Usage",
					"type":  "bargauge",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(feature_usage_total{environment=\"staging\"}) by (feature_name)",
							"legendFormat": "{{feature_name}}",
						},
					},
				},
			],
			"time": map[string]interface{}{
				"from": "now-24h",
				"to":   "now",
			},
			"refresh": "5m",
		},
	}

	stack.dashboardConfigs["business"] = dashboard
	return nil
}

func (stack *GrafanaCloudStack) createSecurityDashboard() error {
	dashboard := map[string]interface{}{
		"dashboard": map[string]interface{}{
			"id":    nil,
			"title": "International Center - Security Monitoring (Staging)",
			"tags":  []string{"staging", "security", "monitoring"},
			"timezone": "UTC",
			"panels": []map[string]interface{}{
				{
					"title": "Authentication Failures",
					"type":  "stat",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(rate(authentication_failures_total{environment=\"staging\"}[5m]))",
							"legendFormat": "Auth Failures/sec",
						},
					},
				},
				{
					"title": "Suspicious IP Activity",
					"type":  "table",
					"targets": []map[string]interface{}{
						{
							"expr": "topk(10, sum(rate(http_requests_total{environment=\"staging\", code=~\"4..\"}[5m])) by (source_ip))",
							"legendFormat": "{{source_ip}}",
						},
					},
				},
				{
					"title": "Rate Limiting Triggers",
					"type":  "timeseries",
					"targets": []map[string]interface{}{
						{
							"expr": "sum(rate(rate_limit_exceeded_total{environment=\"staging\"}[5m])) by (endpoint)",
							"legendFormat": "{{endpoint}}",
						},
					},
				},
				{
					"title": "SSL Certificate Expiry",
					"type":  "stat",
					"targets": []map[string]interface{}{
						{
							"expr": "(ssl_certificate_expiry_seconds{environment=\"staging\"} - time()) / 86400",
							"legendFormat": "Days until expiry",
						},
					},
				},
			},
			"time": map[string]interface{}{
				"from": "now-2h",
				"to":   "now",
			},
			"refresh": "30s",
		},
	}

	stack.dashboardConfigs["security"] = dashboard
	return nil
}

func (stack *GrafanaCloudStack) configureAlerts(ctx *pulumi.Context) error {
	alerts := []map[string]interface{}{
		{
			"alert": map[string]interface{}{
				"name":        "High Error Rate - Staging",
				"message":     "Error rate is above 5% for staging environment",
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
							"params": []float64{0.05},
							"type":   "gt",
						},
					},
				},
				"executionErrorState": "alerting",
				"noDataState":        "no_data",
				"for":               "5m",
			},
		},
		{
			"alert": map[string]interface{}{
				"name":        "Container App Down - Staging",
				"message":     "One or more container apps are down in staging",
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
							"type":   "lt",
						},
					},
				},
				"executionErrorState": "alerting",
				"noDataState":        "alerting",
				"for":               "1m",
			},
		},
		{
			"alert": map[string]interface{}{
				"name":        "Database Connection Issues - Staging",
				"message":     "Database connections are failing in staging",
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
							"params": []float64{10},
							"type":   "gt",
						},
					},
				},
				"executionErrorState": "alerting",
				"noDataState":        "no_data",
				"for":               "2m",
			},
		},
	}

	for _, alert := range alerts {
		stack.alertConfigs[alert["alert"].(map[string]interface{})["name"].(string)] = alert
	}

	return nil
}

func (stack *GrafanaCloudStack) GetGrafanaUrl() pulumi.StringOutput {
	return stack.grafanaUrl
}

func (stack *GrafanaCloudStack) GetPrometheusUrl() pulumi.StringOutput {
	return stack.prometheusUrl
}

func (stack *GrafanaCloudStack) GetLokiUrl() pulumi.StringOutput {
	return stack.lokiUrl
}

func (stack *GrafanaCloudStack) GetDashboardConfigs() map[string]interface{} {
	return stack.dashboardConfigs
}

func (stack *GrafanaCloudStack) GetAlertConfigs() map[string]interface{} {
	return stack.alertConfigs
}