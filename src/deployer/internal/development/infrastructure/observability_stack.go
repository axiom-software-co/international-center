package infrastructure

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type ObservabilityStack struct {
	ctx         *pulumi.Context
	config      *config.Config
	networkName string
	environment string
}

type ObservabilityDeployment struct {
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
}

func NewObservabilityStack(ctx *pulumi.Context, config *config.Config, networkName, environment string) *ObservabilityStack {
	return &ObservabilityStack{
		ctx:         ctx,
		config:      config,
		networkName: networkName,
		environment: environment,
	}
}

func (os *ObservabilityStack) Deploy(ctx context.Context) (*ObservabilityDeployment, error) {
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
		Labels: pulumi.StringMap{
			"environment": pulumi.String(os.environment),
			"component":   pulumi.String("observability"),
			"managed-by":  pulumi.String("pulumi"),
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
		Labels: pulumi.StringMap{
			"environment": pulumi.String(os.environment),
			"component":   pulumi.String("grafana"),
			"data-type":   pulumi.String("persistent"),
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
		Labels: pulumi.StringMap{
			"environment": pulumi.String(os.environment),
			"component":   pulumi.String("grafana"),
			"data-type":   pulumi.String("configuration"),
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
		Labels: pulumi.StringMap{
			"environment": pulumi.String(os.environment),
			"component":   pulumi.String("loki"),
			"data-type":   pulumi.String("persistent"),
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
		Labels: pulumi.StringMap{
			"environment": pulumi.String(os.environment),
			"component":   pulumi.String("loki"),
			"data-type":   pulumi.String("configuration"),
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
		Labels: pulumi.StringMap{
			"environment": pulumi.String(os.environment),
			"component":   pulumi.String("prometheus"),
			"data-type":   pulumi.String("persistent"),
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
		Labels: pulumi.StringMap{
			"environment": pulumi.String(os.environment),
			"component":   pulumi.String("prometheus"),
			"data-type":   pulumi.String("configuration"),
		},
	})
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (os *ObservabilityStack) deployLokiContainer(deployment *ObservabilityDeployment) (*docker.Container, error) {
	lokiPort := os.config.RequireInt("loki_port")

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

		Labels: pulumi.StringMap{
			"environment": pulumi.String(os.environment),
			"component":   pulumi.String("loki"),
			"service":     pulumi.String("logging"),
			"managed-by":  pulumi.String("pulumi"),
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
	prometheusPort := os.config.RequireInt("prometheus_port")

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

		Labels: pulumi.StringMap{
			"environment": pulumi.String(os.environment),
			"component":   pulumi.String("prometheus"),
			"service":     pulumi.String("metrics"),
			"managed-by":  pulumi.String("pulumi"),
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
	grafanaPort := os.config.RequireInt("grafana_port")
	grafanaUser := os.config.Get("grafana_admin_user")
	if grafanaUser == "" {
		grafanaUser = "admin"
	}
	grafanaPassword := os.config.Get("grafana_admin_password")
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

		Labels: pulumi.StringMap{
			"environment": pulumi.String(os.environment),
			"component":   pulumi.String("grafana"),
			"service":     pulumi.String("dashboard"),
			"managed-by":  pulumi.String("pulumi"),
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

func (os *ObservabilityStack) ConfigureDataSources(ctx context.Context, deployment *ObservabilityDeployment) error {
	lokiPort := os.config.RequireInt("loki_port")
	prometheusPort := os.config.RequireInt("prometheus_port")

	lokiDatasourceConfig := fmt.Sprintf(`
apiVersion: 1

datasources:
  - name: Loki
    type: loki
    access: proxy
    url: http://loki:%d
    isDefault: false
    jsonData:
      maxLines: 1000
    editable: true
`, lokiPort)

	prometheusDatasourceConfig := fmt.Sprintf(`
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:%d
    isDefault: true
    jsonData:
      timeInterval: "5s"
      queryTimeout: "60s"
    editable: true
`, prometheusPort)

	_ = lokiDatasourceConfig
	_ = prometheusDatasourceConfig

	return nil
}

func (os *ObservabilityStack) ValidateDeployment(ctx context.Context, deployment *ObservabilityDeployment) error {
	if deployment.GrafanaContainer == nil {
		return fmt.Errorf("Grafana container is not deployed")
	}

	if deployment.LokiContainer == nil {
		return fmt.Errorf("Loki container is not deployed")
	}

	if deployment.PrometheusContainer == nil {
		return fmt.Errorf("Prometheus container is not deployed")
	}

	return nil
}

func (os *ObservabilityStack) GetObservabilityEndpoints() map[string]string {
	grafanaPort := os.config.RequireInt("grafana_port")
	lokiPort := os.config.RequireInt("loki_port")
	prometheusPort := os.config.RequireInt("prometheus_port")

	return map[string]string{
		"grafana":    fmt.Sprintf("http://localhost:%d", grafanaPort),
		"loki":       fmt.Sprintf("http://localhost:%d", lokiPort),
		"prometheus": fmt.Sprintf("http://localhost:%d", prometheusPort),
	}
}

func (os *ObservabilityStack) GetGrafanaCredentials() (string, string) {
	grafanaUser := os.config.Get("grafana_admin_user")
	if grafanaUser == "" {
		grafanaUser = "admin"
	}
	grafanaPassword := os.config.Get("grafana_admin_password")
	if grafanaPassword == "" {
		grafanaPassword = "admin"
	}

	return grafanaUser, grafanaPassword
}