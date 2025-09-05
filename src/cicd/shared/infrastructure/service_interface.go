package infrastructure

import (
	"context"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type ServiceStack interface {
	Deploy(ctx context.Context) (ServiceDeployment, error)
	ValidateDeployment(ctx context.Context, deployment ServiceDeployment) error
	GetServiceEndpoints() map[string]string
	GetDaprConfiguration() DaprConfiguration
}

type ServiceDeployment interface {
	GetServiceEndpoint(serviceName string) pulumi.StringOutput
	GetServiceHealthEndpoint(serviceName string) string
	GetServiceMetrics() ServiceMetrics
	GetScalingConfiguration() ScalingConfiguration
	GetNetworkConfiguration() ServiceNetworkConfiguration
}

type ServiceConfiguration struct {
	Environment       string
	Services          []ServiceConfig
	NetworkType       string // "bridge", "vnet", "private"
	DaprEnabled       bool
	LoadBalancing     bool
	AutoScaling       bool
	HealthChecks      bool
	Monitoring        bool
	SecurityEnabled   bool
	ComplianceMode    bool
}

type ServiceConfig struct {
	Name              string
	Image             string
	Port              int
	Protocol          string
	Environment       map[string]string
	Resources         ResourceConfig
	Replicas          ReplicaConfig
	HealthCheck       HealthCheckConfig
	SecurityContext   SecurityConfig
	DaprConfig        DaprServiceConfig
}

type ResourceConfig struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string
	Storage       string
}

type ReplicaConfig struct {
	Min     int
	Max     int
	Target  int
	Metrics []ScalingMetric
}

type HealthCheckConfig struct {
	Enabled             bool
	Path                string
	InitialDelaySeconds int
	PeriodSeconds       int
	TimeoutSeconds      int
	FailureThreshold    int
	SuccessThreshold    int
}

type SecurityConfig struct {
	RunAsNonRoot     bool
	ReadOnlyRootFS   bool
	AllowPrivileged  bool
	RequireAuth      bool
	EnableTLS        bool
	EnableCSRF       bool
	RateLimit        int
	CORSOrigins      []string
}

type DaprServiceConfig struct {
	Enabled     bool
	AppID       string
	AppPort     int
	HTTPPort    int
	GRPCPort    int
	LogLevel    string
	Components  []string
	EnableTLS   bool
	MTLSEnabled bool
}


type ScalingMetric struct {
	Type       string // "cpu", "memory", "http", "custom"
	Threshold  int
	Window     string
}

type ServiceMetrics struct {
	Availability        float64
	ResponseTime        float64
	ThroughputRPS      int
	ErrorRate          float64
	ResourceUtilization map[string]float64
}

type ScalingConfiguration struct {
	Strategy     string // "reactive", "predictive", "scheduled"
	MinInstances int
	MaxInstances int
	Metrics      []ScalingMetric
	Cooldown     int
}

type ServiceNetworkConfiguration struct {
	Type             string // "bridge", "overlay", "vnet"
	ExternalAccess   bool
	InternalDNS      bool
	LoadBalancer     ServiceLoadBalancerConfig
	ServiceMesh      ServiceMeshConfig
}

type ServiceLoadBalancerConfig struct {
	Enabled    bool
	Algorithm  string // "round_robin", "least_connections", "weighted"
	HealthPath string
	Timeout    int
}

type ServiceMeshConfig struct {
	Enabled     bool
	Provider    string // "dapr", "istio", "linkerd"
	Encryption  bool
	Observability bool
}

type ServiceFactory interface {
	CreateServiceStack(ctx *pulumi.Context, config *config.Config, environment string) ServiceStack
}

func GetServiceConfiguration(environment string, config *config.Config) *ServiceConfiguration {
	switch environment {
	case "development":
		return &ServiceConfiguration{
			Environment:     "development",
			Services:        getDevelopmentServices(),
			NetworkType:     "bridge",
			DaprEnabled:     true,
			LoadBalancing:   false,
			AutoScaling:     false,
			HealthChecks:    true,
			Monitoring:      false,
			SecurityEnabled: false,
			ComplianceMode:  false,
		}
	case "staging":
		return &ServiceConfiguration{
			Environment:     "staging",
			Services:        getStagingServices(),
			NetworkType:     "vnet",
			DaprEnabled:     true,
			LoadBalancing:   true,
			AutoScaling:     true,
			HealthChecks:    true,
			Monitoring:      true,
			SecurityEnabled: true,
			ComplianceMode:  true,
		}
	case "production":
		return &ServiceConfiguration{
			Environment:     "production",
			Services:        getProductionServices(),
			NetworkType:     "private",
			DaprEnabled:     true,
			LoadBalancing:   true,
			AutoScaling:     true,
			HealthChecks:    true,
			Monitoring:      true,
			SecurityEnabled: true,
			ComplianceMode:  true,
		}
	default:
		return &ServiceConfiguration{
			Environment:     environment,
			Services:        getDefaultServices(),
			NetworkType:     "bridge",
			DaprEnabled:     true,
			LoadBalancing:   false,
			AutoScaling:     false,
			HealthChecks:    true,
			Monitoring:      false,
			SecurityEnabled: true,
			ComplianceMode:  false,
		}
	}
}

func getDevelopmentServices() []ServiceConfig {
	return []ServiceConfig{
		{
			Name:     "content-api",
			Image:    "content-api:latest",
			Port:     8080,
			Protocol: "http",
			Resources: ResourceConfig{
				CPURequest:    "100m",
				CPULimit:      "500m",
				MemoryRequest: "256Mi",
				MemoryLimit:   "512Mi",
			},
			Replicas: ReplicaConfig{Min: 1, Max: 1, Target: 1},
			HealthCheck: HealthCheckConfig{
				Enabled:             true,
				Path:                "/health",
				InitialDelaySeconds: 10,
				PeriodSeconds:       10,
				TimeoutSeconds:      5,
			},
			DaprConfig: DaprServiceConfig{
				Enabled:  true,
				AppID:    "content-api",
				AppPort:  8080,
				HTTPPort: 3501,
				GRPCPort: 50002,
			},
		},
		{
			Name:     "services-api",
			Image:    "services-api:latest",
			Port:     8081,
			Protocol: "http",
			Resources: ResourceConfig{
				CPURequest:    "100m",
				CPULimit:      "500m",
				MemoryRequest: "256Mi",
				MemoryLimit:   "512Mi",
			},
			Replicas: ReplicaConfig{Min: 1, Max: 1, Target: 1},
			HealthCheck: HealthCheckConfig{
				Enabled:             true,
				Path:                "/health",
				InitialDelaySeconds: 10,
				PeriodSeconds:       10,
				TimeoutSeconds:      5,
			},
			DaprConfig: DaprServiceConfig{
				Enabled:  true,
				AppID:    "services-api",
				AppPort:  8081,
				HTTPPort: 3502,
				GRPCPort: 50003,
			},
		},
		{
			Name:     "public-gateway",
			Image:    "public-gateway:latest",
			Port:     8082,
			Protocol: "http",
			Resources: ResourceConfig{
				CPURequest:    "100m",
				CPULimit:      "500m",
				MemoryRequest: "256Mi",
				MemoryLimit:   "512Mi",
			},
			Replicas: ReplicaConfig{Min: 1, Max: 1, Target: 1},
			HealthCheck: HealthCheckConfig{
				Enabled:             true,
				Path:                "/health",
				InitialDelaySeconds: 15,
				PeriodSeconds:       10,
				TimeoutSeconds:      5,
			},
			DaprConfig: DaprServiceConfig{
				Enabled:  true,
				AppID:    "public-gateway",
				AppPort:  8082,
				HTTPPort: 3503,
				GRPCPort: 50004,
			},
		},
		{
			Name:     "admin-gateway",
			Image:    "admin-gateway:latest",
			Port:     8083,
			Protocol: "http",
			Resources: ResourceConfig{
				CPURequest:    "100m",
				CPULimit:      "500m",
				MemoryRequest: "256Mi",
				MemoryLimit:   "512Mi",
			},
			Replicas: ReplicaConfig{Min: 1, Max: 1, Target: 1},
			HealthCheck: HealthCheckConfig{
				Enabled:             true,
				Path:                "/health",
				InitialDelaySeconds: 15,
				PeriodSeconds:       10,
				TimeoutSeconds:      5,
			},
			DaprConfig: DaprServiceConfig{
				Enabled:  true,
				AppID:    "admin-gateway",
				AppPort:  8083,
				HTTPPort: 3504,
				GRPCPort: 50006,
			},
		},
	}
}

func getStagingServices() []ServiceConfig {
	services := getDevelopmentServices()
	for i := range services {
		// Staging enhancements
		services[i].Resources = ResourceConfig{
			CPURequest:    "250m",
			CPULimit:      "750m",
			MemoryRequest: "512Mi",
			MemoryLimit:   "1Gi",
		}
		services[i].Replicas = ReplicaConfig{
			Min:    2,
			Max:    10,
			Target: 2,
			Metrics: []ScalingMetric{
				{Type: "http", Threshold: 30, Window: "1m"},
			},
		}
		services[i].SecurityContext = SecurityConfig{
			RunAsNonRoot:    true,
			ReadOnlyRootFS:  false,
			AllowPrivileged: false,
			RequireAuth:     true,
			EnableTLS:       true,
		}
	}
	return services
}

func getProductionServices() []ServiceConfig {
	services := getStagingServices()
	for i := range services {
		// Production enhancements
		services[i].Resources = ResourceConfig{
			CPURequest:    "500m",
			CPULimit:      "1500m",
			MemoryRequest: "1Gi",
			MemoryLimit:   "3Gi",
		}
		
		if services[i].Name == "content-api" || services[i].Name == "services-api" {
			services[i].Replicas = ReplicaConfig{
				Min:    3,
				Max:    50,
				Target: 3,
				Metrics: []ScalingMetric{
					{Type: "http", Threshold: 50, Window: "1m"},
					{Type: "cpu", Threshold: 60, Window: "2m"},
					{Type: "memory", Threshold: 70, Window: "2m"},
				},
			}
		} else {
			// Gateways need higher capacity
			services[i].Replicas = ReplicaConfig{
				Min:    5,
				Max:    100,
				Target: 5,
				Metrics: []ScalingMetric{
					{Type: "http", Threshold: 200, Window: "1m"},
					{Type: "cpu", Threshold: 50, Window: "2m"},
				},
			}
		}
		
		services[i].SecurityContext = SecurityConfig{
			RunAsNonRoot:     true,
			ReadOnlyRootFS:   false,
			AllowPrivileged:  false,
			RequireAuth:      true,
			EnableTLS:        true,
			EnableCSRF:       true,
		}
		
		services[i].HealthCheck = HealthCheckConfig{
			Enabled:             true,
			Path:                "/health",
			InitialDelaySeconds: 60,
			PeriodSeconds:       30,
			TimeoutSeconds:      15,
			FailureThreshold:    3,
			SuccessThreshold:    1,
		}
		
		if services[i].Name == "public-gateway" {
			services[i].SecurityContext.RateLimit = 1000
			services[i].SecurityContext.CORSOrigins = []string{
				"https://www.international-center.com",
				"https://app.international-center.com",
			}
		} else if services[i].Name == "admin-gateway" {
			services[i].SecurityContext.RateLimit = 100
			services[i].SecurityContext.CORSOrigins = []string{
				"https://admin.international-center.com",
			}
		}
	}
	return services
}

func getDefaultServices() []ServiceConfig {
	return getDevelopmentServices()
}

