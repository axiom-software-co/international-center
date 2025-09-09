package services

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ContainerDeploymentArgs struct {
	Environment           string
	InfrastructureOutputs pulumi.Map
	PlatformOutputs      pulumi.Map
	ServiceType          string // "gateway", "content", "inquiries", "notifications"
}

type ContainerDeploymentComponent struct {
	pulumi.ResourceState

	Containers           pulumi.MapOutput    `pulumi:"containers"`
	DaprSidecars         pulumi.MapOutput    `pulumi:"daprSidecars"`
	NetworkConfiguration pulumi.MapOutput    `pulumi:"networkConfiguration"`
	HealthEndpoints      pulumi.MapOutput    `pulumi:"healthEndpoints"`
	ServiceEndpoints     pulumi.MapOutput    `pulumi:"serviceEndpoints"`
	ResourceLimits       pulumi.MapOutput    `pulumi:"resourceLimits"`
}

type ContainerConfig struct {
	Image           string                 `json:"image"`
	ContainerID     string                 `json:"container_id"`
	Port            int                    `json:"port"`
	DaprAppID       string                 `json:"dapr_app_id"`
	HealthEndpoint  string                 `json:"health_endpoint"`
	Environment     map[string]interface{} `json:"environment"`
	ResourceLimits  map[string]interface{} `json:"resource_limits"`
	DaprConfig      map[string]interface{} `json:"dapr_config"`
}

func NewContainerDeploymentComponent(ctx *pulumi.Context, name string, args *ContainerDeploymentArgs, opts ...pulumi.ResourceOption) (*ContainerDeploymentComponent, error) {
	component := &ContainerDeploymentComponent{}
	
	// Safe registration for mock contexts
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:services:ContainerDeployment", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	// Get infrastructure and platform dependencies
	var databaseConnectionString, daprControlPlaneURL string
	
	// Extract database connection from infrastructure outputs
	if dbConnOutput, exists := args.InfrastructureOutputs["database_connection_string"]; exists {
		databaseConnectionString = extractStringFromPulumiInput(dbConnOutput)
	}
	
	// Extract Dapr control plane URL from platform outputs  
	if daprOutput, exists := args.PlatformOutputs["dapr_control_plane_url"]; exists {
		daprControlPlaneURL = extractStringFromPulumiInput(daprOutput)
	}

	var containers, daprSidecars, networkConfiguration, healthEndpoints, serviceEndpoints, resourceLimits pulumi.MapOutput

	switch args.Environment {
	case "development":
		containers, daprSidecars, networkConfiguration, healthEndpoints, serviceEndpoints, resourceLimits = 
			component.createDevelopmentDeployment(args.ServiceType, databaseConnectionString, daprControlPlaneURL)
			
	case "staging":
		containers, daprSidecars, networkConfiguration, healthEndpoints, serviceEndpoints, resourceLimits = 
			component.createStagingDeployment(args.ServiceType, databaseConnectionString, daprControlPlaneURL)
			
	case "production":
		containers, daprSidecars, networkConfiguration, healthEndpoints, serviceEndpoints, resourceLimits = 
			component.createProductionDeployment(args.ServiceType, databaseConnectionString, daprControlPlaneURL)
			
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.Containers = containers
	component.DaprSidecars = daprSidecars
	component.NetworkConfiguration = networkConfiguration
	component.HealthEndpoints = healthEndpoints
	component.ServiceEndpoints = serviceEndpoints
	component.ResourceLimits = resourceLimits

	if canRegister(ctx) {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"containers":           component.Containers,
			"daprSidecars":         component.DaprSidecars,
			"networkConfiguration": component.NetworkConfiguration,
			"healthEndpoints":      component.HealthEndpoints,
			"serviceEndpoints":     component.ServiceEndpoints,
			"resourceLimits":       component.ResourceLimits,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}

func (component *ContainerDeploymentComponent) createDevelopmentDeployment(serviceType, databaseConnectionString, daprControlPlaneURL string) (pulumi.MapOutput, pulumi.MapOutput, pulumi.MapOutput, pulumi.MapOutput, pulumi.MapOutput, pulumi.MapOutput) {
	var containers, daprSidecars, healthEndpoints, serviceEndpoints pulumi.Map
	
	// Resource limits for development
	resourceLimits := pulumi.Map{
		"cpu_limit":    pulumi.String("500m"),
		"memory_limit": pulumi.String("256Mi"),
		"cpu_request":  pulumi.String("100m"),
		"memory_request": pulumi.String("128Mi"),
	}.ToMapOutput()

	// Network configuration for development (Podman)
	networkConfiguration := pulumi.Map{
		"network_name": pulumi.String("international-center-dev"),
		"driver":       pulumi.String("bridge"),
		"subnet":       pulumi.String("172.20.0.0/16"),
		"gateway":      pulumi.String("172.20.0.1"),
	}.ToMapOutput()

	switch serviceType {
	case "gateway":
		containers = pulumi.Map{
			"public": pulumi.Map{
				"image":         pulumi.String("localhost/backend/public-gateway:latest"),
				"container_id":  pulumi.String("public-gateway"),
				"port":          pulumi.Int(9001),
				"dapr_app_id":   pulumi.String("public-gateway"),
				"dapr_port":     pulumi.Int(50001),
				"environment": pulumi.Map{
					"DATABASE_CONNECTION_STRING": pulumi.String(databaseConnectionString),
					"DAPR_HTTP_ENDPOINT":        pulumi.String(daprControlPlaneURL),
					"ENVIRONMENT":               pulumi.String("development"),
					"LOG_LEVEL":                 pulumi.String("debug"),
				},
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
				"dapr_config": pulumi.Map{
					"enabled":           pulumi.Bool(true),
					"app_port":          pulumi.Int(9001),
					"dapr_http_port":    pulumi.Int(50001),
					"dapr_grpc_port":    pulumi.Int(60001),
					"control_plane_address": pulumi.String(daprControlPlaneURL),
				},
			},
			"admin": pulumi.Map{
				"image":         pulumi.String("localhost/backend/admin-gateway:latest"),
				"container_id":  pulumi.String("admin-gateway"),
				"port":          pulumi.Int(9000),
				"dapr_app_id":   pulumi.String("admin-gateway"),
				"dapr_port":     pulumi.Int(50000),
				"environment": pulumi.Map{
					"DATABASE_CONNECTION_STRING": pulumi.String(databaseConnectionString),
					"DAPR_HTTP_ENDPOINT":        pulumi.String(daprControlPlaneURL),
					"ENVIRONMENT":               pulumi.String("development"),
					"LOG_LEVEL":                 pulumi.String("debug"),
				},
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
				"dapr_config": pulumi.Map{
					"enabled":           pulumi.Bool(true),
					"app_port":          pulumi.Int(9000),
					"dapr_http_port":    pulumi.Int(50000),
					"dapr_grpc_port":    pulumi.Int(60000),
					"control_plane_address": pulumi.String(daprControlPlaneURL),
				},
			},
		}

		healthEndpoints = pulumi.Map{
			"public_gateway": pulumi.String("http://127.0.0.1:9001/health"),
			"admin_gateway":  pulumi.String("http://127.0.0.1:9000/health"),
		}

		serviceEndpoints = pulumi.Map{
			"public_gateway": pulumi.String("http://127.0.0.1:9001"),
			"admin_gateway":  pulumi.String("http://127.0.0.1:9000"),
		}

	case "content":
		containers = pulumi.Map{
			"news": pulumi.Map{
				"image":         pulumi.String("localhost/backend/content:latest"),
				"container_id":  pulumi.String("content-news"),
				"port":          pulumi.Int(3001),
				"dapr_app_id":   pulumi.String("content-news"),
				"dapr_port":     pulumi.Int(50011),
				"environment": pulumi.Map{
					"DATABASE_CONNECTION_STRING": pulumi.String(databaseConnectionString),
					"DAPR_HTTP_ENDPOINT":        pulumi.String(daprControlPlaneURL),
					"ENVIRONMENT":               pulumi.String("development"),
					"SERVICE_TYPE":              pulumi.String("news"),
					"LOG_LEVEL":                 pulumi.String("debug"),
				},
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
				"dapr_config": pulumi.Map{
					"enabled":           pulumi.Bool(true),
					"app_port":          pulumi.Int(3001),
					"dapr_http_port":    pulumi.Int(50011),
					"dapr_grpc_port":    pulumi.Int(60011),
					"control_plane_address": pulumi.String(daprControlPlaneURL),
				},
			},
			"events": pulumi.Map{
				"image":         pulumi.String("localhost/backend/content:latest"),
				"container_id":  pulumi.String("content-events"),
				"port":          pulumi.Int(3002),
				"dapr_app_id":   pulumi.String("content-events"),
				"dapr_port":     pulumi.Int(50012),
				"environment": pulumi.Map{
					"DATABASE_CONNECTION_STRING": pulumi.String(databaseConnectionString),
					"DAPR_HTTP_ENDPOINT":        pulumi.String(daprControlPlaneURL),
					"ENVIRONMENT":               pulumi.String("development"),
					"SERVICE_TYPE":              pulumi.String("events"),
					"LOG_LEVEL":                 pulumi.String("debug"),
				},
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
				"dapr_config": pulumi.Map{
					"enabled":           pulumi.Bool(true),
					"app_port":          pulumi.Int(3002),
					"dapr_http_port":    pulumi.Int(50012),
					"dapr_grpc_port":    pulumi.Int(60012),
					"control_plane_address": pulumi.String(daprControlPlaneURL),
				},
			},
			"research": pulumi.Map{
				"image":         pulumi.String("localhost/backend/content:latest"),
				"container_id":  pulumi.String("content-research"),
				"port":          pulumi.Int(3003),
				"dapr_app_id":   pulumi.String("content-research"),
				"dapr_port":     pulumi.Int(50013),
				"environment": pulumi.Map{
					"DATABASE_CONNECTION_STRING": pulumi.String(databaseConnectionString),
					"DAPR_HTTP_ENDPOINT":        pulumi.String(daprControlPlaneURL),
					"ENVIRONMENT":               pulumi.String("development"),
					"SERVICE_TYPE":              pulumi.String("research"),
					"LOG_LEVEL":                 pulumi.String("debug"),
				},
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
				"dapr_config": pulumi.Map{
					"enabled":           pulumi.Bool(true),
					"app_port":          pulumi.Int(3003),
					"dapr_http_port":    pulumi.Int(50013),
					"dapr_grpc_port":    pulumi.Int(60013),
					"control_plane_address": pulumi.String(daprControlPlaneURL),
				},
			},
		}

		healthEndpoints = pulumi.Map{
			"news":     pulumi.String("http://localhost:3001/health"),
			"events":   pulumi.String("http://localhost:3002/health"),
			"research": pulumi.String("http://localhost:3003/health"),
		}

		serviceEndpoints = pulumi.Map{
			"news":     pulumi.String("http://localhost:3001"),
			"events":   pulumi.String("http://localhost:3002"),
			"research": pulumi.String("http://localhost:3003"),
		}

	case "inquiries":
		containers = pulumi.Map{
			"business": pulumi.Map{
				"image":         pulumi.String("localhost/backend/inquiries:latest"),
				"container_id":  pulumi.String("inquiries-business"),
				"port":          pulumi.Int(3101),
				"dapr_app_id":   pulumi.String("inquiries-business"),
				"dapr_port":     pulumi.Int(50021),
				"environment": pulumi.Map{
					"DATABASE_CONNECTION_STRING": pulumi.String(databaseConnectionString),
					"DAPR_HTTP_ENDPOINT":        pulumi.String(daprControlPlaneURL),
					"ENVIRONMENT":               pulumi.String("development"),
					"SERVICE_TYPE":              pulumi.String("business"),
					"LOG_LEVEL":                 pulumi.String("debug"),
				},
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
				"dapr_config": pulumi.Map{
					"enabled":           pulumi.Bool(true),
					"app_port":          pulumi.Int(3101),
					"dapr_http_port":    pulumi.Int(50021),
					"dapr_grpc_port":    pulumi.Int(60021),
					"control_plane_address": pulumi.String(daprControlPlaneURL),
				},
			},
			"donations": pulumi.Map{
				"image":         pulumi.String("localhost/backend/inquiries:latest"),
				"container_id":  pulumi.String("inquiries-donations"),
				"port":          pulumi.Int(3102),
				"dapr_app_id":   pulumi.String("inquiries-donations"),
				"dapr_port":     pulumi.Int(50022),
				"environment": pulumi.Map{
					"DATABASE_CONNECTION_STRING": pulumi.String(databaseConnectionString),
					"DAPR_HTTP_ENDPOINT":        pulumi.String(daprControlPlaneURL),
					"ENVIRONMENT":               pulumi.String("development"),
					"SERVICE_TYPE":              pulumi.String("donations"),
					"LOG_LEVEL":                 pulumi.String("debug"),
				},
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
				"dapr_config": pulumi.Map{
					"enabled":           pulumi.Bool(true),
					"app_port":          pulumi.Int(3102),
					"dapr_http_port":    pulumi.Int(50022),
					"dapr_grpc_port":    pulumi.Int(60022),
					"control_plane_address": pulumi.String(daprControlPlaneURL),
				},
			},
			"media": pulumi.Map{
				"image":         pulumi.String("localhost/backend/inquiries:latest"),
				"container_id":  pulumi.String("inquiries-media"),
				"port":          pulumi.Int(3103),
				"dapr_app_id":   pulumi.String("inquiries-media"),
				"dapr_port":     pulumi.Int(50023),
				"environment": pulumi.Map{
					"DATABASE_CONNECTION_STRING": pulumi.String(databaseConnectionString),
					"DAPR_HTTP_ENDPOINT":        pulumi.String(daprControlPlaneURL),
					"ENVIRONMENT":               pulumi.String("development"),
					"SERVICE_TYPE":              pulumi.String("media"),
					"LOG_LEVEL":                 pulumi.String("debug"),
				},
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
				"dapr_config": pulumi.Map{
					"enabled":           pulumi.Bool(true),
					"app_port":          pulumi.Int(3103),
					"dapr_http_port":    pulumi.Int(50023),
					"dapr_grpc_port":    pulumi.Int(60023),
					"control_plane_address": pulumi.String(daprControlPlaneURL),
				},
			},
			"volunteers": pulumi.Map{
				"image":         pulumi.String("localhost/backend/inquiries:latest"),
				"container_id":  pulumi.String("inquiries-volunteers"),
				"port":          pulumi.Int(3104),
				"dapr_app_id":   pulumi.String("inquiries-volunteers"),
				"dapr_port":     pulumi.Int(50024),
				"environment": pulumi.Map{
					"DATABASE_CONNECTION_STRING": pulumi.String(databaseConnectionString),
					"DAPR_HTTP_ENDPOINT":        pulumi.String(daprControlPlaneURL),
					"ENVIRONMENT":               pulumi.String("development"),
					"SERVICE_TYPE":              pulumi.String("volunteers"),
					"LOG_LEVEL":                 pulumi.String("debug"),
				},
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
				"dapr_config": pulumi.Map{
					"enabled":           pulumi.Bool(true),
					"app_port":          pulumi.Int(3104),
					"dapr_http_port":    pulumi.Int(50024),
					"dapr_grpc_port":    pulumi.Int(60024),
					"control_plane_address": pulumi.String(daprControlPlaneURL),
				},
			},
		}

		healthEndpoints = pulumi.Map{
			"business":   pulumi.String("http://localhost:3101/health"),
			"donations":  pulumi.String("http://localhost:3102/health"),
			"media":      pulumi.String("http://localhost:3103/health"),
			"volunteers": pulumi.String("http://localhost:3104/health"),
		}

		serviceEndpoints = pulumi.Map{
			"business":   pulumi.String("http://localhost:3101"),
			"donations":  pulumi.String("http://localhost:3102"),
			"media":      pulumi.String("http://localhost:3103"),
			"volunteers": pulumi.String("http://localhost:3104"),
		}

	case "notifications":
		containers = pulumi.Map{
			"notification_service": pulumi.Map{
				"image":         pulumi.String("localhost/backend/notifications:latest"),
				"container_id":  pulumi.String("notification-service"),
				"port":          pulumi.Int(3201),
				"dapr_app_id":   pulumi.String("notification-service"),
				"dapr_port":     pulumi.Int(50031),
				"environment": pulumi.Map{
					"DATABASE_CONNECTION_STRING": pulumi.String(databaseConnectionString),
					"DAPR_HTTP_ENDPOINT":        pulumi.String(daprControlPlaneURL),
					"ENVIRONMENT":               pulumi.String("development"),
					"LOG_LEVEL":                 pulumi.String("debug"),
				},
				"resource_limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("256Mi"),
				},
				"dapr_config": pulumi.Map{
					"enabled":           pulumi.Bool(true),
					"app_port":          pulumi.Int(3201),
					"dapr_http_port":    pulumi.Int(50031),
					"dapr_grpc_port":    pulumi.Int(60031),
					"control_plane_address": pulumi.String(daprControlPlaneURL),
				},
			},
		}

		healthEndpoints = pulumi.Map{
			"notification_service": pulumi.String("http://localhost:3201/health"),
		}

		serviceEndpoints = pulumi.Map{
			"notification_service": pulumi.String("http://localhost:3201"),
		}

	default:
		containers = pulumi.Map{}
		healthEndpoints = pulumi.Map{}
		serviceEndpoints = pulumi.Map{}
	}

	// Create Dapr sidecar configurations
	daprSidecars = pulumi.Map{
		"enabled": pulumi.Bool(true),
		"config": pulumi.Map{
			"control_plane_address": pulumi.String(daprControlPlaneURL),
			"placement_host_address": pulumi.String("localhost:50005"),
			"log_level":             pulumi.String("debug"),
			"metrics_port":          pulumi.String("9090"),
			"profile_port":          pulumi.String("7777"),
		},
	}

	return containers.ToMapOutput(), daprSidecars.ToMapOutput(), networkConfiguration, healthEndpoints.ToMapOutput(), serviceEndpoints.ToMapOutput(), resourceLimits
}

func (component *ContainerDeploymentComponent) createStagingDeployment(serviceType, databaseConnectionString, daprControlPlaneURL string) (pulumi.MapOutput, pulumi.MapOutput, pulumi.MapOutput, pulumi.MapOutput, pulumi.MapOutput, pulumi.MapOutput) {
	// Resource limits for staging
	resourceLimits := pulumi.Map{
		"cpu_limit":    pulumi.String("1000m"),
		"memory_limit": pulumi.String("512Mi"),
		"cpu_request":  pulumi.String("250m"),
		"memory_request": pulumi.String("256Mi"),
	}.ToMapOutput()

	// Network configuration for staging (Azure Container Apps)
	networkConfiguration := pulumi.Map{
		"vnet_integration": pulumi.Bool(true),
		"subnet_id":       pulumi.String("/subscriptions/{subscription}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{vnet}/subnets/container-apps-staging"),
		"dns_suffix":      pulumi.String("staging.azurecontainerapp.io"),
		"ingress_enabled": pulumi.Bool(true),
		"traffic_splitting": pulumi.Map{
			"enabled":       pulumi.Bool(true),
			"blue_weight":   pulumi.Int(100),
			"green_weight":  pulumi.Int(0),
		},
	}.ToMapOutput()

	var containers, healthEndpoints, serviceEndpoints pulumi.Map

	switch serviceType {
	case "gateway":
		containers = pulumi.Map{
			"public": pulumi.Map{
				"image":        pulumi.String("registry.azurecr.io/backend/public-gateway:staging"),
				"container_app_name": pulumi.String("public-gateway-staging"),
				"dapr_app_id":  pulumi.String("public-gateway"),
				"environment": pulumi.Map{
					"DATABASE_CONNECTION_STRING": pulumi.String(databaseConnectionString),
					"DAPR_HTTP_ENDPOINT":        pulumi.String(daprControlPlaneURL),
					"ENVIRONMENT":               pulumi.String("staging"),
					"LOG_LEVEL":                 pulumi.String("info"),
				},
				"scaling": pulumi.Map{
					"min_replicas": pulumi.Int(1),
					"max_replicas": pulumi.Int(5),
				},
			},
			"admin": pulumi.Map{
				"image":        pulumi.String("registry.azurecr.io/backend/admin-gateway:staging"),
				"container_app_name": pulumi.String("admin-gateway-staging"),
				"dapr_app_id":  pulumi.String("admin-gateway"),
				"environment": pulumi.Map{
					"DATABASE_CONNECTION_STRING": pulumi.String(databaseConnectionString),
					"DAPR_HTTP_ENDPOINT":        pulumi.String(daprControlPlaneURL),
					"ENVIRONMENT":               pulumi.String("staging"),
					"LOG_LEVEL":                 pulumi.String("info"),
				},
				"scaling": pulumi.Map{
					"min_replicas": pulumi.Int(1),
					"max_replicas": pulumi.Int(3),
				},
			},
		}

		healthEndpoints = pulumi.Map{
			"public_gateway": pulumi.String("https://public-gateway-staging.azurecontainerapp.io/health"),
			"admin_gateway":  pulumi.String("https://admin-gateway-staging.azurecontainerapp.io/health"),
		}

		serviceEndpoints = pulumi.Map{
			"public_gateway": pulumi.String("https://public-gateway-staging.azurecontainerapp.io"),
			"admin_gateway":  pulumi.String("https://admin-gateway-staging.azurecontainerapp.io"),
		}

	default:
		containers = pulumi.Map{}
		healthEndpoints = pulumi.Map{}
		serviceEndpoints = pulumi.Map{}
	}

	// Dapr configuration for Azure Container Apps
	daprSidecars := pulumi.Map{
		"enabled": pulumi.Bool(true),
		"config": pulumi.Map{
			"dapr_managed": pulumi.Bool(true), // Azure Container Apps manages Dapr
			"log_level":    pulumi.String("info"),
		},
	}

	return containers.ToMapOutput(), daprSidecars.ToMapOutput(), networkConfiguration, healthEndpoints.ToMapOutput(), serviceEndpoints.ToMapOutput(), resourceLimits
}

func (component *ContainerDeploymentComponent) createProductionDeployment(serviceType, databaseConnectionString, daprControlPlaneURL string) (pulumi.MapOutput, pulumi.MapOutput, pulumi.MapOutput, pulumi.MapOutput, pulumi.MapOutput, pulumi.MapOutput) {
	// Resource limits for production
	resourceLimits := pulumi.Map{
		"cpu_limit":    pulumi.String("2000m"),
		"memory_limit": pulumi.String("1Gi"),
		"cpu_request":  pulumi.String("500m"),
		"memory_request": pulumi.String("512Mi"),
	}.ToMapOutput()

	// Network configuration for production (Azure Container Apps)
	networkConfiguration := pulumi.Map{
		"vnet_integration": pulumi.Bool(true),
		"subnet_id":       pulumi.String("/subscriptions/{subscription}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{vnet}/subnets/container-apps-production"),
		"dns_suffix":      pulumi.String("production.azurecontainerapp.io"),
		"ingress_enabled": pulumi.Bool(true),
		"traffic_splitting": pulumi.Map{
			"enabled":        pulumi.Bool(true),
			"canary_weight":  pulumi.Int(10),
			"stable_weight":  pulumi.Int(90),
		},
	}.ToMapOutput()

	var containers, healthEndpoints, serviceEndpoints pulumi.Map

	switch serviceType {
	case "gateway":
		containers = pulumi.Map{
			"public": pulumi.Map{
				"image":        pulumi.String("registry.azurecr.io/backend/public-gateway:production"),
				"container_app_name": pulumi.String("public-gateway-production"),
				"dapr_app_id":  pulumi.String("public-gateway"),
				"environment": pulumi.Map{
					"DATABASE_CONNECTION_STRING": pulumi.String(databaseConnectionString),
					"DAPR_HTTP_ENDPOINT":        pulumi.String(daprControlPlaneURL),
					"ENVIRONMENT":               pulumi.String("production"),
					"LOG_LEVEL":                 pulumi.String("warn"),
				},
				"scaling": pulumi.Map{
					"min_replicas": pulumi.Int(3),
					"max_replicas": pulumi.Int(20),
				},
			},
			"admin": pulumi.Map{
				"image":        pulumi.String("registry.azurecr.io/backend/admin-gateway:production"),
				"container_app_name": pulumi.String("admin-gateway-production"),
				"dapr_app_id":  pulumi.String("admin-gateway"),
				"environment": pulumi.Map{
					"DATABASE_CONNECTION_STRING": pulumi.String(databaseConnectionString),
					"DAPR_HTTP_ENDPOINT":        pulumi.String(daprControlPlaneURL),
					"ENVIRONMENT":               pulumi.String("production"),
					"LOG_LEVEL":                 pulumi.String("warn"),
				},
				"scaling": pulumi.Map{
					"min_replicas": pulumi.Int(2),
					"max_replicas": pulumi.Int(10),
				},
			},
		}

		healthEndpoints = pulumi.Map{
			"public_gateway": pulumi.String("https://public-gateway-production.azurecontainerapp.io/health"),
			"admin_gateway":  pulumi.String("https://admin-gateway-production.azurecontainerapp.io/health"),
		}

		serviceEndpoints = pulumi.Map{
			"public_gateway": pulumi.String("https://public-gateway-production.azurecontainerapp.io"),
			"admin_gateway":  pulumi.String("https://admin-gateway-production.azurecontainerapp.io"),
		}

	default:
		containers = pulumi.Map{}
		healthEndpoints = pulumi.Map{}
		serviceEndpoints = pulumi.Map{}
	}

	// Dapr configuration for Azure Container Apps
	daprSidecars := pulumi.Map{
		"enabled": pulumi.Bool(true),
		"config": pulumi.Map{
			"dapr_managed": pulumi.Bool(true), // Azure Container Apps manages Dapr
			"log_level":    pulumi.String("warn"),
		},
	}

	return containers.ToMapOutput(), daprSidecars.ToMapOutput(), networkConfiguration, healthEndpoints.ToMapOutput(), serviceEndpoints.ToMapOutput(), resourceLimits
}

// Helper function to extract string from Pulumi input (simplified for this context)
func extractStringFromPulumiInput(input interface{}) string {
	// In real deployment, this would properly resolve the Pulumi input
	// For now, return a simplified extraction
	switch v := input.(type) {
	case string:
		return v
	case pulumi.StringInput:
		// This would be resolved properly in actual deployment context
		return "resolved-connection-string"
	default:
		return "default-connection-string"
	}
}

