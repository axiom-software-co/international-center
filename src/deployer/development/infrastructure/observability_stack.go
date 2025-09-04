package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	oslib "os"
	"strconv"
	"time"

	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	sharedconfig "github.com/axiom-software-co/international-center/src/deployer/shared/config"
	sharedinfra "github.com/axiom-software-co/international-center/src/deployer/shared/infrastructure"
)

type ObservabilityStack struct {
	pulumi.ComponentResource
	ctx           *pulumi.Context
	config        *config.Config
	configManager *sharedconfig.ConfigManager
	networkName   string
	environment   string
	
	// Outputs
	GrafanaEndpoint    pulumi.StringOutput `pulumi:"grafanaEndpoint"`
	LokiEndpoint       pulumi.StringOutput `pulumi:"lokiEndpoint"`
	PrometheusEndpoint pulumi.StringOutput `pulumi:"prometheusEndpoint"`
	ObservabilityNetworkID pulumi.StringOutput `pulumi:"observabilityNetworkId"`
}

type ObservabilityDeployment struct {
	pulumi.ComponentResource
	GrafanaContainer       *docker.Container
	LokiContainer          *docker.Container
	PrometheusContainer    *docker.Container
	ObservabilityNetwork   *docker.Network
	GrafanaDataVolume      *docker.Volume
	GrafanaConfigVolume    *docker.Volume
	LokiDataVolume         *docker.Volume
	LokiConfigVolume       *docker.Volume
	PrometheusDataVolume   *docker.Volume
	PrometheusConfigVolume *docker.Volume
	
	// Outputs
	GrafanaEndpoint    pulumi.StringOutput `pulumi:"grafanaEndpoint"`
	LokiEndpoint       pulumi.StringOutput `pulumi:"lokiEndpoint"`
	PrometheusEndpoint pulumi.StringOutput `pulumi:"prometheusEndpoint"`
	NetworkID          pulumi.StringOutput `pulumi:"networkId"`
}

// Implement the shared ObservabilityDeployment interface
func (od *ObservabilityDeployment) GetMetricsEndpoint() pulumi.StringOutput {
	return od.PrometheusEndpoint
}

func (od *ObservabilityDeployment) GetLogsEndpoint() pulumi.StringOutput {
	return od.LokiEndpoint
}

func (od *ObservabilityDeployment) GetTracingEndpoint() pulumi.StringOutput {
	// In development, we don't have separate tracing, return empty
	return pulumi.String("").ToStringOutput()
}

func (od *ObservabilityDeployment) GetDashboardEndpoint() pulumi.StringOutput {
	return od.GrafanaEndpoint
}

func (od *ObservabilityDeployment) GetAlertManagerEndpoint() pulumi.StringOutput {
	// In development, we don't have separate alertmanager, return Grafana endpoint
	return od.GrafanaEndpoint
}

func NewObservabilityStack(ctx *pulumi.Context, config *config.Config, networkName, environment string) *ObservabilityStack {
	// Create ConfigManager for centralized configuration
	configManager, err := sharedconfig.NewConfigManager(ctx)
	if err != nil {
		ctx.Log.Warn(fmt.Sprintf("Failed to create ConfigManager, using legacy configuration: %v", err), nil)
		configManager = nil
	}
	
	component := &ObservabilityStack{
		ctx:           ctx,
		config:        config,
		configManager: configManager,
		networkName:   networkName,
		environment:   environment,
	}
	
	err = ctx.RegisterComponentResource("international-center:observability:DevelopmentStack",
		fmt.Sprintf("%s-observability-stack", environment), component)
	if err != nil {
		return nil
	}
	
	return component
}

func (os *ObservabilityStack) Deploy(ctx context.Context) (sharedinfra.ObservabilityDeployment, error) {
	deployment := &ObservabilityDeployment{}

	var err error

	deployment.ObservabilityNetwork, err = os.createObservabilityNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to create observability network: %w", err)
	}

	// Create volumes
	deployment.GrafanaDataVolume, err = os.createGrafanaDataVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create Grafana data volume: %w", err)
	}

	deployment.GrafanaConfigVolume, err = os.createGrafanaConfigVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create Grafana config volume: %w", err)
	}

	deployment.LokiDataVolume, err = os.createLokiDataVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create Loki data volume: %w", err)
	}

	deployment.LokiConfigVolume, err = os.createLokiConfigVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create Loki config volume: %w", err)
	}

	deployment.PrometheusDataVolume, err = os.createPrometheusDataVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus data volume: %w", err)
	}

	deployment.PrometheusConfigVolume, err = os.createPrometheusConfigVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus config volume: %w", err)
	}

	// Deploy containers
	deployment.LokiContainer, err = os.deployLokiContainer(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Loki container: %w", err)
	}

	deployment.PrometheusContainer, err = os.deployPrometheusContainer(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Prometheus container: %w", err)
	}

	deployment.GrafanaContainer, err = os.deployGrafanaContainer(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy Grafana container: %w", err)
	}

	return deployment, nil
}

func (os *ObservabilityStack) createObservabilityNetwork() (*docker.Network, error) {
	network, err := docker.NewNetwork(os.ctx, "observability-network", &docker.NetworkArgs{
		Name:   pulumi.Sprintf("%s-observability-network", os.environment),
		Driver: pulumi.String("bridge"),
		Options: pulumi.StringMap{
			"com.docker.network.driver.mtu": pulumi.String("1500"),
		},
		Labels: docker.NetworkLabelArray{
			&docker.NetworkLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(os.environment),
			},
			&docker.NetworkLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("observability"),
			},
			&docker.NetworkLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return network, nil
}

func (os *ObservabilityStack) createGrafanaDataVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(os.ctx, "grafana-data", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-grafana-data", os.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(os.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("grafana"),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("data-type"),
				Value: pulumi.String("persistent"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (os *ObservabilityStack) createGrafanaConfigVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(os.ctx, "grafana-config", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-grafana-config", os.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(os.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("grafana"),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("data-type"),
				Value: pulumi.String("configuration"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (os *ObservabilityStack) createLokiDataVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(os.ctx, "loki-data", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-loki-data", os.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(os.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("loki"),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("data-type"),
				Value: pulumi.String("persistent"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (os *ObservabilityStack) createLokiConfigVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(os.ctx, "loki-config", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-loki-config", os.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(os.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("loki"),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("data-type"),
				Value: pulumi.String("configuration"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (os *ObservabilityStack) createPrometheusDataVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(os.ctx, "prometheus-data", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-prometheus-data", os.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(os.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("prometheus"),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("data-type"),
				Value: pulumi.String("persistent"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (os *ObservabilityStack) createPrometheusConfigVolume() (*docker.Volume, error) {
	volume, err := docker.NewVolume(os.ctx, "prometheus-config", &docker.VolumeArgs{
		Name:   pulumi.Sprintf("%s-prometheus-config", os.environment),
		Driver: pulumi.String("local"),
		Labels: docker.VolumeLabelArray{
			&docker.VolumeLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(os.environment),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("prometheus"),
			},
			&docker.VolumeLabelArgs{
				Label: pulumi.String("data-type"),
				Value: pulumi.String("configuration"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (os *ObservabilityStack) deployLokiContainer(deployment *ObservabilityDeployment) (*docker.Container, error) {
	lokiPort, err := strconv.Atoi(oslib.Getenv("LOKI_PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid LOKI_PORT: %w", err)
	}

	container, err := docker.NewContainer(os.ctx, "loki", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-loki", os.environment),
		Image:   pulumi.String("grafana/loki:2.9.0"),
		Restart: pulumi.String("unless-stopped"),

		Command: pulumi.StringArray{
			pulumi.String("-config.file=/etc/loki/local-config.yaml"),
			pulumi.String("-target=all"),
			pulumi.String("-server.http-listen-port=3100"),
			pulumi.String("-log.level=info"),
		},

		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(3100),
				External: pulumi.Int(lokiPort),
				Protocol: pulumi.String("tcp"),
			},
		},

		Mounts: docker.ContainerMountArray{
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.LokiDataVolume.Name,
				Target: pulumi.String("/loki"),
			},
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.LokiConfigVolume.Name,
				Target: pulumi.String("/etc/loki/config"),
			},
		},

		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: deployment.ObservabilityNetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("loki"),
					pulumi.String("logs"),
				},
			},
		},

		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests: pulumi.StringArray{
				pulumi.String("CMD-SHELL"),
				pulumi.String("wget --no-verbose --tries=1 --spider http://localhost:3100/ready || exit 1"),
			},
			Interval: pulumi.String("30s"),
			Timeout:  pulumi.String("10s"),
			Retries:  pulumi.Int(3),
			StartPeriod: pulumi.String("60s"),
		},

		Labels: docker.ContainerLabelArray{
			&docker.ContainerLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(os.environment),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("loki"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("service"),
				Value: pulumi.String("logging"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},

		LogDriver: pulumi.String("json-file"),
		LogOpts: pulumi.StringMap{
			"max-size": pulumi.String("10m"),
			"max-file": pulumi.String("3"),
		},
	})
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (os *ObservabilityStack) deployPrometheusContainer(deployment *ObservabilityDeployment) (*docker.Container, error) {
	prometheusPort, err := strconv.Atoi(oslib.Getenv("PROMETHEUS_PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid PROMETHEUS_PORT: %w", err)
	}

	container, err := docker.NewContainer(os.ctx, "prometheus", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-prometheus", os.environment),
		Image:   pulumi.String("prom/prometheus:v2.47.0"),
		Restart: pulumi.String("unless-stopped"),

		Command: pulumi.StringArray{
			pulumi.String("--config.file=/etc/prometheus/prometheus.yml"),
			pulumi.String("--storage.tsdb.path=/prometheus"),
			pulumi.String("--web.console.libraries=/etc/prometheus/console_libraries"),
			pulumi.String("--web.console.templates=/etc/prometheus/consoles"),
			pulumi.String("--storage.tsdb.retention.time=15d"),
			pulumi.String("--web.enable-lifecycle"),
			pulumi.String("--log.level=info"),
		},

		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(9090),
				External: pulumi.Int(prometheusPort),
				Protocol: pulumi.String("tcp"),
			},
		},

		Mounts: docker.ContainerMountArray{
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.PrometheusDataVolume.Name,
				Target: pulumi.String("/prometheus"),
			},
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.PrometheusConfigVolume.Name,
				Target: pulumi.String("/etc/prometheus/config"),
			},
		},

		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: deployment.ObservabilityNetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("prometheus"),
					pulumi.String("metrics"),
				},
			},
		},

		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests: pulumi.StringArray{
				pulumi.String("CMD-SHELL"),
				pulumi.String("wget --no-verbose --tries=1 --spider http://localhost:9090/-/healthy || exit 1"),
			},
			Interval: pulumi.String("30s"),
			Timeout:  pulumi.String("10s"),
			Retries:  pulumi.Int(3),
		},

		Labels: docker.ContainerLabelArray{
			&docker.ContainerLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(os.environment),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("prometheus"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("service"),
				Value: pulumi.String("metrics"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},

		LogDriver: pulumi.String("json-file"),
		LogOpts: pulumi.StringMap{
			"max-size": pulumi.String("10m"),
			"max-file": pulumi.String("3"),
		},
	})
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (os *ObservabilityStack) deployGrafanaContainer(deployment *ObservabilityDeployment) (*docker.Container, error) {
	grafanaPort, err := strconv.Atoi(oslib.Getenv("GRAFANA_PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid GRAFANA_PORT: %w", err)
	}
	grafanaUser := oslib.Getenv("GRAFANA_ADMIN_USER")
	if grafanaUser == "" {
		grafanaUser = "admin"
	}
	grafanaPassword := oslib.Getenv("GRAFANA_ADMIN_PASSWORD")
	if grafanaPassword == "" {
		grafanaPassword = "admin"
	}

	envVars := pulumi.StringArray{
		pulumi.String("GF_SECURITY_ADMIN_USER=" + grafanaUser),
		pulumi.String("GF_SECURITY_ADMIN_PASSWORD=" + grafanaPassword),
		pulumi.String("GF_USERS_ALLOW_SIGN_UP=false"),
		pulumi.String("GF_LOG_LEVEL=info"),
		pulumi.String("GF_INSTALL_PLUGINS=grafana-clock-panel,grafana-simple-json-datasource"),
		pulumi.String("GF_FEATURE_TOGGLES_ENABLE=publicDashboards"),
		pulumi.String("GF_SERVER_ROOT_URL=http://localhost:" + fmt.Sprintf("%d", grafanaPort)),
		pulumi.String("GF_ANALYTICS_REPORTING_ENABLED=false"),
		pulumi.String("GF_ANALYTICS_CHECK_FOR_UPDATES=false"),
	}

	container, err := docker.NewContainer(os.ctx, "grafana", &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-grafana", os.environment),
		Image:   pulumi.String("grafana/grafana:10.1.0"),
		Restart: pulumi.String("unless-stopped"),

		Envs: envVars,

		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(3000),
				External: pulumi.Int(grafanaPort),
				Protocol: pulumi.String("tcp"),
			},
		},

		Mounts: docker.ContainerMountArray{
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.GrafanaDataVolume.Name,
				Target: pulumi.String("/var/lib/grafana"),
			},
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: deployment.GrafanaConfigVolume.Name,
				Target: pulumi.String("/etc/grafana/provisioning"),
			},
		},

		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: deployment.ObservabilityNetwork.Name,
				Aliases: pulumi.StringArray{
					pulumi.String("grafana"),
					pulumi.String("dashboard"),
				},
			},
		},

		Healthcheck: &docker.ContainerHealthcheckArgs{
			Tests: pulumi.StringArray{
				pulumi.String("CMD-SHELL"),
				pulumi.String("wget --no-verbose --tries=1 --spider http://localhost:3000/api/health || exit 1"),
			},
			Interval: pulumi.String("30s"),
			Timeout:  pulumi.String("10s"),
			Retries:  pulumi.Int(3),
			StartPeriod: pulumi.String("60s"),
		},

		Labels: docker.ContainerLabelArray{
			&docker.ContainerLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(os.environment),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("grafana"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("service"),
				Value: pulumi.String("dashboard"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},

		LogDriver: pulumi.String("json-file"),
		LogOpts: pulumi.StringMap{
			"max-size": pulumi.String("10m"),
			"max-file": pulumi.String("3"),
		},

		User: pulumi.String("472:472"),
	})
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (os *ObservabilityStack) ConfigureDataSources(ctx context.Context, deployment sharedinfra.ObservabilityDeployment) error {
	lokiPort, err := strconv.Atoi(oslib.Getenv("LOKI_PORT"))
	if err != nil {
		return fmt.Errorf("invalid LOKI_PORT: %w", err)
	}
	prometheusPort, err := strconv.Atoi(oslib.Getenv("PROMETHEUS_PORT"))
	if err != nil {
		return fmt.Errorf("invalid PROMETHEUS_PORT: %w", err)
	}

	grafanaPort, err := strconv.Atoi(oslib.Getenv("GRAFANA_PORT"))
	if err != nil {
		return fmt.Errorf("invalid GRAFANA_PORT: %w", err)
	}
	grafanaUser := oslib.Getenv("GRAFANA_ADMIN_USER")
	if grafanaUser == "" {
		grafanaUser = "admin"
	}
	grafanaPassword := oslib.Getenv("GRAFANA_ADMIN_PASSWORD")
	if grafanaPassword == "" {
		grafanaPassword = "admin"
	}

	grafanaURL := fmt.Sprintf("http://localhost:%d", grafanaPort)

	if err := os.configureGrafanaDatasource(ctx, grafanaURL, grafanaUser, grafanaPassword, "Prometheus", "prometheus", fmt.Sprintf("http://prometheus:%d", prometheusPort), true); err != nil {
		return fmt.Errorf("failed to configure Prometheus datasource: %w", err)
	}

	if err := os.configureGrafanaDatasource(ctx, grafanaURL, grafanaUser, grafanaPassword, "Loki", "loki", fmt.Sprintf("http://loki:%d", lokiPort), false); err != nil {
		return fmt.Errorf("failed to configure Loki datasource: %w", err)
	}

	return nil
}

func (os *ObservabilityStack) ValidateDeployment(ctx context.Context, deployment sharedinfra.ObservabilityDeployment) error {
	// Cast to concrete type to access implementation details
	concreteDeployment, ok := deployment.(*ObservabilityDeployment)
	if !ok {
		return fmt.Errorf("deployment is not a valid ObservabilityDeployment implementation")
	}

	if concreteDeployment.GrafanaContainer == nil {
		return fmt.Errorf("Grafana container is not deployed")
	}

	if concreteDeployment.LokiContainer == nil {
		return fmt.Errorf("Loki container is not deployed")
	}

	if concreteDeployment.PrometheusContainer == nil {
		return fmt.Errorf("Prometheus container is not deployed")
	}

	return nil
}

func (os *ObservabilityStack) GetObservabilityEndpoints() map[string]string {
	grafanaPort, err := strconv.Atoi(oslib.Getenv("GRAFANA_PORT"))
	if err != nil {
		grafanaPort = 3000
	}
	lokiPort, err := strconv.Atoi(oslib.Getenv("LOKI_PORT"))
	if err != nil {
		lokiPort = 3100
	}
	prometheusPort, err := strconv.Atoi(oslib.Getenv("PROMETHEUS_PORT"))
	if err != nil {
		prometheusPort = 9090
	}

	return map[string]string{
		"grafana":    fmt.Sprintf("http://localhost:%d", grafanaPort),
		"loki":       fmt.Sprintf("http://localhost:%d", lokiPort),
		"prometheus": fmt.Sprintf("http://localhost:%d", prometheusPort),
	}
}

func (os *ObservabilityStack) ConfigureAlerts(ctx context.Context) error {
	// For development environment, we don't configure complex alerts
	// This method is required by the shared interface but can be no-op in dev
	os.ctx.Log.Info("ConfigureAlerts: Skipping alert configuration for development environment", nil)
	return nil
}

func (os *ObservabilityStack) CreateDashboards(ctx context.Context) error {
	// For development environment, we use default dashboards
	// This method provides basic dashboard setup
	os.ctx.Log.Info("CreateDashboards: Using default dashboards for development environment", nil)
	return nil
}

func (os *ObservabilityStack) GetGrafanaCredentials() (string, string) {
	grafanaUser := oslib.Getenv("GRAFANA_ADMIN_USER")
	if grafanaUser == "" {
		grafanaUser = "admin"
	}
	grafanaPassword := oslib.Getenv("GRAFANA_ADMIN_PASSWORD")
	if grafanaPassword == "" {
		grafanaPassword = "admin"
	}

	return grafanaUser, grafanaPassword
}

func (os *ObservabilityStack) configureGrafanaDatasource(ctx context.Context, grafanaURL, username, password, name, dsType, url string, isDefault bool) error {
	datasourceConfig := map[string]interface{}{
		"name":      name,
		"type":      dsType,
		"url":       url,
		"access":    "proxy",
		"isDefault": isDefault,
		"basicAuth": false,
		"editable":  true,
	}

	if dsType == "prometheus" {
		datasourceConfig["jsonData"] = map[string]interface{}{
			"timeInterval": "5s",
			"queryTimeout": "60s",
		}
	} else if dsType == "loki" {
		datasourceConfig["jsonData"] = map[string]interface{}{
			"maxLines": 1000,
		}
	}

	jsonData, err := json.Marshal(datasourceConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal datasource config: %w", err)
	}

	datasourceURL := fmt.Sprintf("%s/api/datasources", grafanaURL)
	req, err := http.NewRequestWithContext(ctx, "POST", datasourceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to configure datasource: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}