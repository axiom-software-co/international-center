package platform

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DaprDeploymentArgs struct {
	Environment           string
	InfrastructureOutputs pulumi.Map
}

type DaprDeploymentComponent struct {
	pulumi.ResourceState

	ControlPlaneContainer pulumi.MapOutput    `pulumi:"controlPlaneContainer"`
	PlacementContainer   pulumi.MapOutput    `pulumi:"placementContainer"`
	SentryContainer      pulumi.MapOutput    `pulumi:"sentryContainer"`
	ContainerNetwork     pulumi.StringOutput `pulumi:"containerNetwork"`
	HealthEndpoints      pulumi.MapOutput    `pulumi:"healthEndpoints"`
}

func NewDaprDeploymentComponent(ctx *pulumi.Context, name string, args *DaprDeploymentArgs, opts ...pulumi.ResourceOption) (*DaprDeploymentComponent, error) {
	component := &DaprDeploymentComponent{}
	
	// Safe registration for mock contexts
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:platform:DaprDeployment", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	var controlPlaneContainer, placementContainer, sentryContainer pulumi.MapOutput
	var containerNetwork pulumi.StringOutput
	var healthEndpoints pulumi.MapOutput

	switch args.Environment {
	case "development":
		// Development: Deploy actual Podman containers for Dapr control plane
		containerNetwork = pulumi.String("podman").ToStringOutput()
		
		controlPlaneContainer = pulumi.Map{
			"image":         pulumi.String("daprio/dapr:latest"),
			"container_id":  pulumi.String("dapr-control-plane"),
			"command":       pulumi.Array{
				pulumi.String("./daprd"),
				pulumi.String("--mode"), pulumi.String("kubernetes"),
				pulumi.String("--dapr-http-port"), pulumi.String("3500"),
				pulumi.String("--dapr-grpc-port"), pulumi.String("50001"),
				pulumi.String("--dapr-internal-grpc-port"), pulumi.String("50002"),
				pulumi.String("--dapr-listen-addresses"), pulumi.String("0.0.0.0"),
				pulumi.String("--dapr-public-port"), pulumi.String("3501"),
				pulumi.String("--app-port"), pulumi.String("80"),
				pulumi.String("--app-id"), pulumi.String("dapr-control-plane"),
				pulumi.String("--control-plane-address"), pulumi.String("localhost:50001"),
				pulumi.String("--config"), pulumi.String("/dapr/config/config.yaml"),
				pulumi.String("--log-level"), pulumi.String("info"),
			},
			"ports": pulumi.Array{
				pulumi.Map{"container_port": pulumi.Int(3500), "host_port": pulumi.Int(3500)},
				pulumi.Map{"container_port": pulumi.Int(50001), "host_port": pulumi.Int(50001)},
				pulumi.Map{"container_port": pulumi.Int(50002), "host_port": pulumi.Int(50002)},
				pulumi.Map{"container_port": pulumi.Int(3501), "host_port": pulumi.Int(3501)},
			},
			"volumes": pulumi.Array{
				pulumi.Map{
					"host_path":      pulumi.String("src/public-website/deployment/configs/dapr"),
					"container_path": pulumi.String("/dapr/config"),
				},
			},
			"environment_variables": pulumi.Map{
				"DAPR_TRUST_ANCHORS":     pulumi.String("/dapr/certs/ca.crt"),
				"DAPR_CERT_CHAIN":        pulumi.String("/dapr/certs/issuer.crt"),
				"DAPR_CERT_KEY":          pulumi.String("/dapr/certs/issuer.key"),
				"NAMESPACE":              pulumi.String("default"),
			},
			"resource_limits": pulumi.Map{
				"cpu":    pulumi.String("500m"),
				"memory": pulumi.String("256Mi"),
			},
			"health_check": pulumi.Map{
				"test":     pulumi.Array{
					pulumi.String("CMD-SHELL"),
					pulumi.String("wget --no-verbose --tries=1 --spider http://localhost:3502/v1.0/healthz || exit 1"),
				},
				"interval": pulumi.String("30s"),
				"timeout":  pulumi.String("10s"),
				"retries":  pulumi.Int(3),
			},
			"restart_policy": pulumi.String("unless-stopped"),
			"network_mode":   pulumi.String("bridge"),
		}.ToMapOutput()

		placementContainer = pulumi.Map{
			"image":         pulumi.String("daprio/dapr:latest"),
			"container_id":  pulumi.String("dapr-placement"),
			"command":       pulumi.Array{
				pulumi.String("./placement"),
				pulumi.String("--port"), pulumi.String("50005"),
				pulumi.String("--log-level"), pulumi.String("info"),
				pulumi.String("--tls-enabled"), pulumi.String("false"),
			},
			"ports": pulumi.Array{
				pulumi.Map{"container_port": pulumi.Int(50005), "host_port": pulumi.Int(50005)},
			},
			"environment_variables": pulumi.Map{
				"NAMESPACE": pulumi.String("default"),
			},
			"resource_limits": pulumi.Map{
				"cpu":    pulumi.String("200m"),
				"memory": pulumi.String("128Mi"),
			},
			"health_check": pulumi.Map{
				"test":     pulumi.Array{
					pulumi.String("CMD-SHELL"),
					pulumi.String("nc -z localhost 50005 || exit 1"),
				},
				"interval": pulumi.String("30s"),
				"timeout":  pulumi.String("10s"),
				"retries":  pulumi.Int(3),
			},
			"restart_policy": pulumi.String("unless-stopped"),
			"network_mode":   pulumi.String("bridge"),
		}.ToMapOutput()

		sentryContainer = pulumi.Map{
			"image":         pulumi.String("daprio/dapr:latest"),
			"container_id":  pulumi.String("dapr-sentry"),
			"command":       pulumi.Array{
				pulumi.String("./sentry"),
				pulumi.String("--port"), pulumi.String("50003"),
				pulumi.String("--log-level"), pulumi.String("info"),
				pulumi.String("--trust-domain"), pulumi.String("public"),
			},
			"ports": pulumi.Array{
				pulumi.Map{"container_port": pulumi.Int(50003), "host_port": pulumi.Int(50003)},
			},
			"volumes": pulumi.Array{
				pulumi.Map{
					"host_path":      pulumi.String("src/public-website/deployment/configs/dapr/certs"),
					"container_path": pulumi.String("/dapr/certs"),
				},
			},
			"environment_variables": pulumi.Map{
				"NAMESPACE": pulumi.String("default"),
			},
			"resource_limits": pulumi.Map{
				"cpu":    pulumi.String("200m"),
				"memory": pulumi.String("128Mi"),
			},
			"health_check": pulumi.Map{
				"test":     pulumi.Array{
					pulumi.String("CMD-SHELL"),
					pulumi.String("nc -z localhost 50003 || exit 1"),
				},
				"interval": pulumi.String("30s"),
				"timeout":  pulumi.String("10s"),
				"retries":  pulumi.Int(3),
			},
			"restart_policy": pulumi.String("unless-stopped"),
			"network_mode":   pulumi.String("bridge"),
		}.ToMapOutput()

		healthEndpoints = pulumi.Map{
			"control_plane": pulumi.String("http://localhost:3502/v1.0/healthz"),
			"placement":     pulumi.String("tcp://localhost:50005"),
			"sentry":        pulumi.String("tcp://localhost:50003"),
		}.ToMapOutput()

	case "staging":
		// Staging: Deploy using Azure Container Apps with Dapr extension
		containerNetwork = pulumi.String("azure_container_apps").ToStringOutput()
		
		controlPlaneContainer = pulumi.Map{
			"container_app_name": pulumi.String("dapr-control-plane-staging"),
			"image":              pulumi.String("mcr.microsoft.com/dapr/dapr:latest"),
			"cpu":                pulumi.Float64(0.5),
			"memory":             pulumi.String("1Gi"),
			"replicas": pulumi.Map{
				"min_replicas": pulumi.Int(1),
				"max_replicas": pulumi.Int(3),
			},
			"dapr": pulumi.Map{
				"enabled":        pulumi.Bool(true),
				"app_id":         pulumi.String("dapr-control-plane"),
				"app_port":       pulumi.Int(3500),
				"app_protocol":   pulumi.String("http"),
			},
			"ingress": pulumi.Map{
				"external": pulumi.Bool(false),
				"target_port": pulumi.Int(3500),
			},
			"environment_variables": pulumi.Array{
				pulumi.Map{"name": pulumi.String("DAPR_HTTP_PORT"), "value": pulumi.String("3500")},
				pulumi.Map{"name": pulumi.String("DAPR_GRPC_PORT"), "value": pulumi.String("50001")},
				pulumi.Map{"name": pulumi.String("NAMESPACE"), "value": pulumi.String("default")},
			},
		}.ToMapOutput()

		// Azure Container Apps manage Dapr control plane automatically
		placementContainer = pulumi.Map{
			"managed_by_platform": pulumi.Bool(true),
			"service_type":        pulumi.String("dapr_placement"),
		}.ToMapOutput()

		sentryContainer = pulumi.Map{
			"managed_by_platform": pulumi.Bool(true),
			"service_type":        pulumi.String("dapr_sentry"),
		}.ToMapOutput()

		healthEndpoints = pulumi.Map{
			"control_plane": pulumi.String("https://dapr-control-plane-staging.azurecontainerapp.io/v1.0/healthz"),
			"placement":     pulumi.String("managed_by_azure"),
			"sentry":        pulumi.String("managed_by_azure"),
		}.ToMapOutput()

	case "production":
		// Production: Deploy using Azure Container Apps with enhanced security
		containerNetwork = pulumi.String("azure_container_apps").ToStringOutput()
		
		controlPlaneContainer = pulumi.Map{
			"container_app_name": pulumi.String("dapr-control-plane-production"),
			"image":              pulumi.String("mcr.microsoft.com/dapr/dapr:1.12.0"), // Fixed version for production
			"cpu":                pulumi.Float64(1.0),
			"memory":             pulumi.String("2Gi"),
			"replicas": pulumi.Map{
				"min_replicas": pulumi.Int(2),
				"max_replicas": pulumi.Int(5),
			},
			"dapr": pulumi.Map{
				"enabled":        pulumi.Bool(true),
				"app_id":         pulumi.String("dapr-control-plane"),
				"app_port":       pulumi.Int(3500),
				"app_protocol":   pulumi.String("https"),
			},
			"ingress": pulumi.Map{
				"external": pulumi.Bool(false),
				"target_port": pulumi.Int(3500),
				"traffic": pulumi.Array{
					pulumi.Map{"weight": pulumi.Int(100), "latest_revision": pulumi.Bool(true)},
				},
			},
			"environment_variables": pulumi.Array{
				pulumi.Map{"name": pulumi.String("DAPR_HTTP_PORT"), "value": pulumi.String("3500")},
				pulumi.Map{"name": pulumi.String("DAPR_GRPC_PORT"), "value": pulumi.String("50001")},
				pulumi.Map{"name": pulumi.String("NAMESPACE"), "value": pulumi.String("default")},
				pulumi.Map{"name": pulumi.String("DAPR_LOG_LEVEL"), "value": pulumi.String("warn")},
			},
			"security": pulumi.Map{
				"allow_insecure_connections": pulumi.Bool(false),
				"disable_anonymous_access":   pulumi.Bool(true),
			},
		}.ToMapOutput()

		placementContainer = pulumi.Map{
			"managed_by_platform": pulumi.Bool(true),
			"service_type":        pulumi.String("dapr_placement"),
			"high_availability":   pulumi.Bool(true),
		}.ToMapOutput()

		sentryContainer = pulumi.Map{
			"managed_by_platform": pulumi.Bool(true),
			"service_type":        pulumi.String("dapr_sentry"),
			"certificate_management": pulumi.Bool(true),
		}.ToMapOutput()

		healthEndpoints = pulumi.Map{
			"control_plane": pulumi.String("https://dapr-control-plane-production.azurecontainerapp.io/v1.0/healthz"),
			"placement":     pulumi.String("managed_by_azure"),
			"sentry":        pulumi.String("managed_by_azure"),
		}.ToMapOutput()

	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.ControlPlaneContainer = controlPlaneContainer
	component.PlacementContainer = placementContainer
	component.SentryContainer = sentryContainer
	component.ContainerNetwork = containerNetwork
	component.HealthEndpoints = healthEndpoints

	if canRegister(ctx) {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"controlPlaneContainer": component.ControlPlaneContainer,
			"placementContainer":    component.PlacementContainer,
			"sentryContainer":       component.SentryContainer,
			"containerNetwork":      component.ContainerNetwork,
			"healthEndpoints":       component.HealthEndpoints,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}

// deployDaprContainerWithPodman deploys Dapr container using Podman for development
func (component *DaprDeploymentComponent) deployDaprContainerWithPodman(ctx *pulumi.Context, containerConfig pulumi.Map) error {
	// This function would integrate with Podman to actually deploy containers
	// For now, this is a placeholder that would be called from the container orchestrator
	
	// TODO: Implement actual Podman container deployment
	// This should:
	// 1. Pull the Dapr container image
	// 2. Create and configure container networks
	// 3. Start containers with proper port mappings and volumes
	// 4. Set up health checks
	// 5. Configure container restart policies
	// 6. Validate containers are running and healthy
	
	return nil
}

// deployDaprContainerWithAzure deploys Dapr container using Azure Container Apps for staging/production
func (component *DaprDeploymentComponent) deployDaprContainerWithAzure(ctx *pulumi.Context, containerConfig pulumi.Map) error {
	// This function would integrate with Azure Container Apps to deploy containers
	// For now, this is a placeholder that would be called from the container orchestrator
	
	// TODO: Implement actual Azure Container Apps deployment
	// This should:
	// 1. Create Azure Container App resources
	// 2. Configure Dapr extension
	// 3. Set up ingress and networking
	// 4. Configure environment variables and secrets
	// 5. Set up health checks and monitoring
	// 6. Configure auto-scaling policies
	// 7. Validate deployment is successful
	
	return nil
}

// generateDaprConfiguration creates Dapr configuration files for the deployment
func generateDaprConfiguration(environment string) string {
	baseConfig := `
apiVersion: dapr.io/v1alpha1
kind: Configuration
metadata:
  name: daprConfig
  namespace: default
spec:
  tracing:
    samplingRate: "1"
  metric:
    enabled: true
`

	switch environment {
	case "development":
		return baseConfig + `
  middleware:
    http:
      - name: cors
        spec:
          allowedOrigins: ["*"]
      - name: rateLimiter
        spec:
          maxRequestsPerSecond: 1000
`
	case "staging":
		return baseConfig + `
  middleware:
    http:
      - name: cors
        spec:
          allowedOrigins: ["https://*.azurecontainerapp.io"]
      - name: rateLimiter
        spec:
          maxRequestsPerSecond: 500
  mtls:
    enabled: true
    workloadCertTTL: "24h"
    allowedClockSkew: "15m"
`
	case "production":
		return baseConfig + `
  middleware:
    http:
      - name: cors
        spec:
          allowedOrigins: ["https://production.domain.com"]
      - name: rateLimiter
        spec:
          maxRequestsPerSecond: 200
      - name: oauth2
        spec:
          clientId: "${OAUTH_CLIENT_ID}"
          clientSecret: "${OAUTH_CLIENT_SECRET}"
  mtls:
    enabled: true
    workloadCertTTL: "12h"
    allowedClockSkew: "5m"
  accessControl:
    defaultAction: "deny"
    trustDomain: "production"
`
	}

	return baseConfig
}

// Helper function to extract container ID from configuration
func extractContainerID(config pulumi.Map, key string) string {
	// In real deployment, this would properly resolve Pulumi inputs
	// For now, return a default based on the key
	switch key {
	case "control_plane_container":
		return "dapr-control-plane"
	case "placement_container":
		return "dapr-placement"
	case "sentry_container":
		return "dapr-sentry"
	default:
		return ""
	}
}

// Helper function to validate container deployment health
func validateContainerHealth(containerID string, healthEndpoint string) error {
	// This function would implement actual health validation
	// TODO: Implement health check validation for deployed containers
	return fmt.Errorf("container health validation not implemented for %s", containerID)
}

// Helper function to generate Podman run command for development
func generatePodmanRunCommand(containerConfig pulumi.Map) []string {
	var cmd []string
	cmd = append(cmd, "podman", "run", "-d")
	
	// Add basic configuration
	if name, ok := containerConfig["container_id"]; ok {
		cmd = append(cmd, "--name", fmt.Sprintf("%v", name))
	}
	
	// Add port mappings
	if ports, ok := containerConfig["ports"]; ok {
		if portsArray, ok := ports.(pulumi.Array); ok {
			for _, port := range portsArray {
				if portMap, ok := port.(pulumi.Map); ok {
					hostPort := portMap["host_port"]
					containerPort := portMap["container_port"]
					cmd = append(cmd, "-p", fmt.Sprintf("%v:%v", hostPort, containerPort))
				}
			}
		}
	}
	
	// Add volume mounts
	if volumes, ok := containerConfig["volumes"]; ok {
		if volumesArray, ok := volumes.(pulumi.Array); ok {
			for _, volume := range volumesArray {
				if volumeMap, ok := volume.(pulumi.Map); ok {
					hostPath := volumeMap["host_path"]
					containerPath := volumeMap["container_path"]
					cmd = append(cmd, "-v", fmt.Sprintf("%v:%v", hostPath, containerPath))
				}
			}
		}
	}
	
	// Add environment variables
	if envVars, ok := containerConfig["environment_variables"]; ok {
		if envMap, ok := envVars.(pulumi.Map); ok {
			for key, value := range envMap {
				cmd = append(cmd, "-e", fmt.Sprintf("%s=%v", key, value))
			}
		}
	}
	
	// Add image
	if image, ok := containerConfig["image"]; ok {
		cmd = append(cmd, fmt.Sprintf("%v", image))
	}
	
	// Add command arguments
	if command, ok := containerConfig["command"]; ok {
		if cmdArray, ok := command.(pulumi.Array); ok {
			for _, arg := range cmdArray {
				cmd = append(cmd, fmt.Sprintf("%v", arg))
			}
		}
	}
	
	return cmd
}

// Helper function to format Podman command as string for logging
func formatPodmanCommand(cmd []string) string {
	return strings.Join(cmd, " ")
}