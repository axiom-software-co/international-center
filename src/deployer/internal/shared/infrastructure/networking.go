package infrastructure

import (
	"fmt"
	"net"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
)

// NetworkManager provides network configuration and management
type NetworkManager struct {
	ctx         *pulumi.Context
	environment string
	networkPrefix string
}

// NewNetworkManager creates a new network manager
func NewNetworkManager(ctx *pulumi.Context, environment, networkPrefix string) *NetworkManager {
	return &NetworkManager{
		ctx:           ctx,
		environment:   environment,
		networkPrefix: networkPrefix,
	}
}

// NetworkConfiguration defines complete network setup
type NetworkConfiguration struct {
	// Core networking
	MainNetwork    NetworkSpec            `json:"main_network"`
	ServiceNetworks []ServiceNetworkSpec  `json:"service_networks"`
	
	// Service discovery
	ServiceDiscovery ServiceDiscoveryConfig `json:"service_discovery"`
	
	// Load balancing
	LoadBalancer LoadBalancerConfig `json:"load_balancer"`
	
	// Security
	SecurityRules []NetworkSecurityRule `json:"security_rules"`
	
	// Monitoring
	NetworkMonitoring NetworkMonitoringConfig `json:"network_monitoring"`
}

// NetworkSpec defines network specification
type NetworkSpec struct {
	Name       string                 `json:"name"`
	CIDR       string                 `json:"cidr"`
	Driver     string                 `json:"driver"`
	EnableIPv6 bool                   `json:"enable_ipv6"`
	Internal   bool                   `json:"internal"`
	Options    map[string]string      `json:"options"`
	Labels     map[string]string      `json:"labels"`
	IPAM       IPAMConfiguration      `json:"ipam"`
}

// ServiceNetworkSpec defines service-specific network configuration
type ServiceNetworkSpec struct {
	ServiceName string                 `json:"service_name"`
	NetworkName string                 `json:"network_name"`
	IPAddress   string                 `json:"ip_address"`
	Aliases     []string               `json:"aliases"`
	Ports       []ServicePort          `json:"ports"`
	HealthCheck NetworkHealthCheck     `json:"health_check"`
}

// ServicePort defines service port configuration
type ServicePort struct {
	Name        string `json:"name"`
	Port        int    `json:"port"`
	TargetPort  int    `json:"target_port"`
	Protocol    string `json:"protocol"`
	ExternalIP  string `json:"external_ip"`
	LoadBalance bool   `json:"load_balance"`
}

// NetworkHealthCheck defines network-level health check
type NetworkHealthCheck struct {
	Enabled         bool   `json:"enabled"`
	Protocol        string `json:"protocol"`
	Path            string `json:"path"`
	Port            int    `json:"port"`
	IntervalSeconds int    `json:"interval_seconds"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
	HealthyThreshold int   `json:"healthy_threshold"`
	UnhealthyThreshold int `json:"unhealthy_threshold"`
}

// ServiceDiscoveryConfig defines service discovery configuration
type ServiceDiscoveryConfig struct {
	Enabled     bool                   `json:"enabled"`
	Provider    string                 `json:"provider"` // "dns", "consul", "etcd"
	Namespace   string                 `json:"namespace"`
	Domain      string                 `json:"domain"`
	Services    []ServiceRegistration  `json:"services"`
	HealthChecks []DiscoveryHealthCheck `json:"health_checks"`
}

// ServiceRegistration defines service registration for discovery
type ServiceRegistration struct {
	Name        string            `json:"name"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Tags        []string          `json:"tags"`
	Meta        map[string]string `json:"meta"`
	HealthCheck DiscoveryHealthCheck `json:"health_check"`
}

// DiscoveryHealthCheck defines health check for service discovery
type DiscoveryHealthCheck struct {
	Type                string `json:"type"` // "http", "tcp", "grpc"
	URL                 string `json:"url"`
	Method              string `json:"method"`
	Headers             map[string]string `json:"headers"`
	IntervalSeconds     int    `json:"interval_seconds"`
	TimeoutSeconds      int    `json:"timeout_seconds"`
	DeregisterAfterSeconds int `json:"deregister_after_seconds"`
}

// LoadBalancerConfig defines load balancer configuration
type LoadBalancerConfig struct {
	Enabled         bool                    `json:"enabled"`
	Type            string                  `json:"type"` // "round_robin", "least_connections", "ip_hash"
	Algorithm       string                  `json:"algorithm"`
	StickySession   bool                    `json:"sticky_session"`
	HealthCheck     LoadBalancerHealthCheck `json:"health_check"`
	Backends        []LoadBalancerBackend   `json:"backends"`
	SSL             SSLConfiguration        `json:"ssl"`
}

// LoadBalancerBackend defines backend server configuration
type LoadBalancerBackend struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	Weight      int    `json:"weight"`
	MaxConns    int    `json:"max_conns"`
	FailTimeout int    `json:"fail_timeout"`
}

// LoadBalancerHealthCheck defines load balancer health check
type LoadBalancerHealthCheck struct {
	Enabled         bool   `json:"enabled"`
	Path            string `json:"path"`
	Method          string `json:"method"`
	ExpectedStatus  int    `json:"expected_status"`
	IntervalSeconds int    `json:"interval_seconds"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
	HealthyThreshold int   `json:"healthy_threshold"`
	UnhealthyThreshold int `json:"unhealthy_threshold"`
}

// SSLConfiguration defines SSL/TLS configuration
type SSLConfiguration struct {
	Enabled            bool     `json:"enabled"`
	CertificateFile    string   `json:"certificate_file"`
	PrivateKeyFile     string   `json:"private_key_file"`
	CACertificateFile  string   `json:"ca_certificate_file"`
	Protocols          []string `json:"protocols"`
	Ciphers            []string `json:"ciphers"`
	VerifyClient       bool     `json:"verify_client"`
}

// NetworkSecurityRule defines network security rules
type NetworkSecurityRule struct {
	Name         string              `json:"name"`
	Priority     int                 `json:"priority"`
	Direction    string              `json:"direction"` // "inbound", "outbound"
	Action       string              `json:"action"`    // "allow", "deny"
	Protocol     string              `json:"protocol"`  // "tcp", "udp", "icmp", "*"
	SourceRanges []string            `json:"source_ranges"`
	TargetRanges []string            `json:"target_ranges"`
	Ports        []PortRange         `json:"ports"`
	Tags         map[string]string   `json:"tags"`
}

// PortRange defines port range for security rules
type PortRange struct {
	From int `json:"from"`
	To   int `json:"to"`
}

// NetworkMonitoringConfig defines network monitoring configuration
type NetworkMonitoringConfig struct {
	Enabled           bool                      `json:"enabled"`
	CollectMetrics    bool                      `json:"collect_metrics"`
	CollectFlowLogs   bool                      `json:"collect_flow_logs"`
	MetricsInterval   int                       `json:"metrics_interval_seconds"`
	FlowLogFormat     string                    `json:"flow_log_format"`
	AlertRules        []NetworkAlertRule        `json:"alert_rules"`
	TrafficAnalysis   TrafficAnalysisConfig     `json:"traffic_analysis"`
}

// NetworkAlertRule defines network monitoring alert rules
type NetworkAlertRule struct {
	Name        string            `json:"name"`
	Metric      string            `json:"metric"`
	Threshold   float64           `json:"threshold"`
	Comparison  string            `json:"comparison"`
	Duration    string            `json:"duration"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// TrafficAnalysisConfig defines traffic analysis configuration
type TrafficAnalysisConfig struct {
	Enabled             bool     `json:"enabled"`
	SamplingRate        float64  `json:"sampling_rate"`
	RetentionDays       int      `json:"retention_days"`
	AnalyzedProtocols   []string `json:"analyzed_protocols"`
	DetectAnomalies     bool     `json:"detect_anomalies"`
	GeoLocationTracking bool     `json:"geo_location_tracking"`
}

// IPAMConfiguration defines IPAM configuration
type IPAMConfiguration struct {
	Driver  string              `json:"driver"`
	Options map[string]string   `json:"options"`
	Configs []IPAMSubnet        `json:"configs"`
}

// IPAMSubnet defines IPAM subnet configuration
type IPAMSubnet struct {
	Subnet     string            `json:"subnet"`
	IPRange    string            `json:"ip_range"`
	Gateway    string            `json:"gateway"`
	AuxAddress map[string]string `json:"aux_addresses"`
}

// CreateNetworkConfiguration creates the complete network setup
func (nm *NetworkManager) CreateNetworkConfiguration(config NetworkConfiguration) (*NetworkResources, error) {
	resources := &NetworkResources{
		Networks: make(map[string]*docker.Network),
		Services: make(map[string]*ServiceNetworkConfig),
	}
	
	// Create main network
	mainNetwork, err := nm.createNetwork(config.MainNetwork)
	if err != nil {
		return nil, fmt.Errorf("failed to create main network: %w", err)
	}
	resources.Networks[config.MainNetwork.Name] = mainNetwork
	
	// Create service-specific networks
	for _, serviceNet := range config.ServiceNetworks {
		serviceConfig := &ServiceNetworkConfig{
			NetworkName: serviceNet.NetworkName,
			IPAddress:   serviceNet.IPAddress,
			Aliases:     serviceNet.Aliases,
			Ports:       serviceNet.Ports,
		}
		resources.Services[serviceNet.ServiceName] = serviceConfig
	}
	
	return resources, nil
}

// createNetwork creates a Docker network
func (nm *NetworkManager) createNetwork(spec NetworkSpec) (*docker.Network, error) {
	// Build network name with environment prefix
	networkName := nm.buildNetworkName(spec.Name)
	
	// Configure IPAM
	var ipamConfig *docker.NetworkIpamArgs
	if spec.IPAM.Driver != "" {
		ipamConfig = &docker.NetworkIpamArgs{
			Driver:  pulumi.String(spec.IPAM.Driver),
			Options: nm.convertStringMapToPulumiMap(spec.IPAM.Options),
		}
		
		if len(spec.IPAM.Configs) > 0 {
			configs := make(docker.NetworkIpamConfigArray, len(spec.IPAM.Configs))
			for i, cfg := range spec.IPAM.Configs {
				config := &docker.NetworkIpamConfigArgs{
					Subnet:  pulumi.String(cfg.Subnet),
					Gateway: pulumi.String(cfg.Gateway),
				}
				
				if cfg.IPRange != "" {
					config.IpRange = pulumi.String(cfg.IPRange)
				}
				
				if len(cfg.AuxAddress) > 0 {
					config.AuxAddress = nm.convertStringMapToPulumiMap(cfg.AuxAddress)
				}
				
				configs[i] = config
			}
			ipamConfig.Configs = configs
		}
	}
	
	// Create network arguments
	networkArgs := &docker.NetworkArgs{
		Name:     pulumi.String(networkName),
		Driver:   pulumi.String(spec.Driver),
		Internal: pulumi.Bool(spec.Internal),
		Ipam:     ipamConfig,
	}
	
	// Add options if provided
	if len(spec.Options) > 0 {
		networkArgs.Options = nm.convertStringMapToPulumiMap(spec.Options)
	}
	
	// Add labels
	labels := nm.buildNetworkLabels(spec.Labels)
	if len(labels) > 0 {
		networkArgs.Labels = labels
	}
	
	// Create the network
	network, err := docker.NewNetwork(nm.ctx, networkName, networkArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to create network %s: %w", networkName, err)
	}
	
	return network, nil
}

// CreateServiceNetworkConfiguration creates network configuration for services
func (nm *NetworkManager) CreateServiceNetworkConfiguration() *ServiceNetworkConfiguration {
	return &ServiceNetworkConfiguration{
		// Content API network configuration
		ContentAPI: ServiceNetworkSpec{
			ServiceName: "content-api",
			NetworkName: nm.buildNetworkName("main"),
			Aliases:     []string{"content-api", "content"},
			Ports: []ServicePort{
				{Name: "http", Port: 8080, TargetPort: 8080, Protocol: "tcp"},
				{Name: "metrics", Port: 9090, TargetPort: 9090, Protocol: "tcp"},
			},
			HealthCheck: NetworkHealthCheck{
				Enabled:            true,
				Protocol:           "http",
				Path:               "/health",
				Port:               8080,
				IntervalSeconds:    30,
				TimeoutSeconds:     5,
				HealthyThreshold:   2,
				UnhealthyThreshold: 3,
			},
		},
		
		// Services API network configuration
		ServicesAPI: ServiceNetworkSpec{
			ServiceName: "services-api",
			NetworkName: nm.buildNetworkName("main"),
			Aliases:     []string{"services-api", "services"},
			Ports: []ServicePort{
				{Name: "http", Port: 8081, TargetPort: 8081, Protocol: "tcp"},
				{Name: "metrics", Port: 9091, TargetPort: 9091, Protocol: "tcp"},
			},
			HealthCheck: NetworkHealthCheck{
				Enabled:            true,
				Protocol:           "http",
				Path:               "/health",
				Port:               8081,
				IntervalSeconds:    30,
				TimeoutSeconds:     5,
				HealthyThreshold:   2,
				UnhealthyThreshold: 3,
			},
		},
		
		// Public Gateway network configuration
		PublicGateway: ServiceNetworkSpec{
			ServiceName: "public-gateway",
			NetworkName: nm.buildNetworkName("main"),
			Aliases:     []string{"public-gateway", "gateway"},
			Ports: []ServicePort{
				{Name: "http", Port: 80, TargetPort: 8080, Protocol: "tcp", ExternalIP: "0.0.0.0", LoadBalance: true},
				{Name: "https", Port: 443, TargetPort: 8443, Protocol: "tcp", ExternalIP: "0.0.0.0", LoadBalance: true},
			},
			HealthCheck: NetworkHealthCheck{
				Enabled:            true,
				Protocol:           "http",
				Path:               "/health",
				Port:               8080,
				IntervalSeconds:    15,
				TimeoutSeconds:     5,
				HealthyThreshold:   2,
				UnhealthyThreshold: 5,
			},
		},
		
		// Admin Gateway network configuration
		AdminGateway: ServiceNetworkSpec{
			ServiceName: "admin-gateway",
			NetworkName: nm.buildNetworkName("main"),
			Aliases:     []string{"admin-gateway", "admin"},
			Ports: []ServicePort{
				{Name: "https", Port: 8443, TargetPort: 8443, Protocol: "tcp", LoadBalance: true},
			},
			HealthCheck: NetworkHealthCheck{
				Enabled:            true,
				Protocol:           "https",
				Path:               "/health",
				Port:               8443,
				IntervalSeconds:    30,
				TimeoutSeconds:     10,
				HealthyThreshold:   2,
				UnhealthyThreshold: 3,
			},
		},
		
		// PostgreSQL network configuration
		PostgreSQL: ServiceNetworkSpec{
			ServiceName: "postgresql",
			NetworkName: nm.buildNetworkName("data"),
			Aliases:     []string{"postgresql", "postgres", "db"},
			Ports: []ServicePort{
				{Name: "postgresql", Port: 5432, TargetPort: 5432, Protocol: "tcp"},
			},
			HealthCheck: NetworkHealthCheck{
				Enabled:            true,
				Protocol:           "tcp",
				Port:               5432,
				IntervalSeconds:    30,
				TimeoutSeconds:     10,
				HealthyThreshold:   1,
				UnhealthyThreshold: 3,
			},
		},
		
		// Redis network configuration (for pub/sub and sessions)
		Redis: ServiceNetworkSpec{
			ServiceName: "redis",
			NetworkName: nm.buildNetworkName("data"),
			Aliases:     []string{"redis", "redis-pubsub", "cache"},
			Ports: []ServicePort{
				{Name: "redis", Port: 6379, TargetPort: 6379, Protocol: "tcp"},
			},
			HealthCheck: NetworkHealthCheck{
				Enabled:            true,
				Protocol:           "tcp",
				Port:               6379,
				IntervalSeconds:    30,
				TimeoutSeconds:     5,
				HealthyThreshold:   1,
				UnhealthyThreshold: 3,
			},
		},
		
		// Azurite (blob storage emulator) network configuration
		Azurite: ServiceNetworkSpec{
			ServiceName: "azurite",
			NetworkName: nm.buildNetworkName("data"),
			Aliases:     []string{"azurite", "blob-storage", "storage"},
			Ports: []ServicePort{
				{Name: "blob", Port: 10000, TargetPort: 10000, Protocol: "tcp"},
				{Name: "queue", Port: 10001, TargetPort: 10001, Protocol: "tcp"},
				{Name: "table", Port: 10002, TargetPort: 10002, Protocol: "tcp"},
			},
			HealthCheck: NetworkHealthCheck{
				Enabled:            true,
				Protocol:           "http",
				Path:               "/",
				Port:               10000,
				IntervalSeconds:    30,
				TimeoutSeconds:     5,
				HealthyThreshold:   1,
				UnhealthyThreshold: 3,
			},
		},
		
		// Vault network configuration
		Vault: ServiceNetworkSpec{
			ServiceName: "vault",
			NetworkName: nm.buildNetworkName("data"),
			Aliases:     []string{"vault", "secrets"},
			Ports: []ServicePort{
				{Name: "http", Port: 8200, TargetPort: 8200, Protocol: "tcp"},
			},
			HealthCheck: NetworkHealthCheck{
				Enabled:            true,
				Protocol:           "http",
				Path:               "/v1/sys/health",
				Port:               8200,
				IntervalSeconds:    30,
				TimeoutSeconds:     10,
				HealthyThreshold:   1,
				UnhealthyThreshold: 3,
			},
		},
		
		// Grafana network configuration
		Grafana: ServiceNetworkSpec{
			ServiceName: "grafana",
			NetworkName: nm.buildNetworkName("observability"),
			Aliases:     []string{"grafana", "dashboard"},
			Ports: []ServicePort{
				{Name: "http", Port: 3000, TargetPort: 3000, Protocol: "tcp"},
			},
			HealthCheck: NetworkHealthCheck{
				Enabled:            true,
				Protocol:           "http",
				Path:               "/api/health",
				Port:               3000,
				IntervalSeconds:    30,
				TimeoutSeconds:     5,
				HealthyThreshold:   1,
				UnhealthyThreshold: 3,
			},
		},
		
		// Loki network configuration
		Loki: ServiceNetworkSpec{
			ServiceName: "loki",
			NetworkName: nm.buildNetworkName("observability"),
			Aliases:     []string{"loki", "logs"},
			Ports: []ServicePort{
				{Name: "http", Port: 3100, TargetPort: 3100, Protocol: "tcp"},
				{Name: "grpc", Port: 9095, TargetPort: 9095, Protocol: "tcp"},
			},
			HealthCheck: NetworkHealthCheck{
				Enabled:            true,
				Protocol:           "http",
				Path:               "/ready",
				Port:               3100,
				IntervalSeconds:    30,
				TimeoutSeconds:     5,
				HealthyThreshold:   1,
				UnhealthyThreshold: 3,
			},
		},
		
		// Dapr placement service network configuration
		DaprPlacement: ServiceNetworkSpec{
			ServiceName: "dapr-placement",
			NetworkName: nm.buildNetworkName("main"),
			Aliases:     []string{"dapr-placement", "placement"},
			Ports: []ServicePort{
				{Name: "http", Port: 50005, TargetPort: 50005, Protocol: "tcp"},
				{Name: "grpc", Port: 50006, TargetPort: 50006, Protocol: "tcp"},
			},
			HealthCheck: NetworkHealthCheck{
				Enabled:            true,
				Protocol:           "http",
				Path:               "/v1.0/healthz",
				Port:               50005,
				IntervalSeconds:    30,
				TimeoutSeconds:     5,
				HealthyThreshold:   1,
				UnhealthyThreshold: 3,
			},
		},
	}
}

// GetNetworkSecurityRules returns environment-appropriate security rules
func (nm *NetworkManager) GetNetworkSecurityRules() []NetworkSecurityRule {
	baseRules := []NetworkSecurityRule{
		// Allow internal communication
		{
			Name:      "allow-internal",
			Priority:  100,
			Direction: "inbound",
			Action:    "allow",
			Protocol:  "*",
			SourceRanges: []string{
				nm.getNetworkCIDR(),
			},
			TargetRanges: []string{
				nm.getNetworkCIDR(),
			},
		},
		
		// Allow public gateway external access
		{
			Name:      "allow-public-gateway",
			Priority:  200,
			Direction: "inbound",
			Action:    "allow",
			Protocol:  "tcp",
			SourceRanges: []string{"0.0.0.0/0"}, // Allow from anywhere for public gateway
			TargetRanges: []string{nm.getNetworkCIDR()},
			Ports: []PortRange{
				{From: 80, To: 80},   // HTTP
				{From: 443, To: 443}, // HTTPS
			},
		},
		
		// Deny all other external access by default
		{
			Name:      "deny-external",
			Priority:  9999,
			Direction: "inbound",
			Action:    "deny",
			Protocol:  "*",
			SourceRanges: []string{"0.0.0.0/0"},
			TargetRanges: []string{nm.getNetworkCIDR()},
		},
	}
	
	// Add environment-specific rules
	if nm.environment == "development" {
		// More permissive rules for development
		baseRules = append(baseRules, NetworkSecurityRule{
			Name:      "allow-dev-access",
			Priority:  150,
			Direction: "inbound",
			Action:    "allow",
			Protocol:  "tcp",
			SourceRanges: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
			TargetRanges: []string{nm.getNetworkCIDR()},
			Ports: []PortRange{
				{From: 8000, To: 9000}, // Development ports
			},
		})
	}
	
	return baseRules
}

// Helper methods

func (nm *NetworkManager) buildNetworkName(baseName string) string {
	return fmt.Sprintf("%s-%s-%s", nm.networkPrefix, nm.environment, baseName)
}

func (nm *NetworkManager) buildNetworkLabels(customLabels map[string]string) pulumi.StringMap {
	labels := pulumi.StringMap{
		"environment":  pulumi.String(nm.environment),
		"managed-by":   pulumi.String("pulumi"),
		"project":      pulumi.String("international-center"),
	}
	
	// Add custom labels
	for key, value := range customLabels {
		labels[key] = pulumi.String(value)
	}
	
	return labels
}

func (nm *NetworkManager) convertStringMapToPulumiMap(input map[string]string) pulumi.StringMap {
	if len(input) == 0 {
		return nil
	}
	
	result := make(pulumi.StringMap)
	for key, value := range input {
		result[key] = pulumi.String(value)
	}
	return result
}

func (nm *NetworkManager) getNetworkCIDR() string {
	// Return environment-specific CIDR
	switch nm.environment {
	case "development":
		return "10.0.0.0/16"
	case "staging":
		return "10.1.0.0/16"
	case "production":
		return "10.2.0.0/16"
	default:
		return "10.0.0.0/16"
	}
}

// ValidateNetworkConfiguration validates network configuration
func (nm *NetworkManager) ValidateNetworkConfiguration(config NetworkConfiguration) error {
	// Validate CIDR ranges
	if config.MainNetwork.CIDR != "" {
		_, _, err := net.ParseCIDR(config.MainNetwork.CIDR)
		if err != nil {
			return fmt.Errorf("invalid CIDR range %s: %w", config.MainNetwork.CIDR, err)
		}
	}
	
	// Validate service ports
	for _, service := range config.ServiceNetworks {
		for _, port := range service.Ports {
			if port.Port <= 0 || port.Port > 65535 {
				return fmt.Errorf("invalid port %d for service %s", port.Port, service.ServiceName)
			}
			if port.TargetPort <= 0 || port.TargetPort > 65535 {
				return fmt.Errorf("invalid target port %d for service %s", port.TargetPort, service.ServiceName)
			}
		}
	}
	
	// Validate security rules
	for _, rule := range config.SecurityRules {
		if rule.Direction != "inbound" && rule.Direction != "outbound" {
			return fmt.Errorf("invalid direction %s for security rule %s", rule.Direction, rule.Name)
		}
		if rule.Action != "allow" && rule.Action != "deny" {
			return fmt.Errorf("invalid action %s for security rule %s", rule.Action, rule.Name)
		}
	}
	
	return nil
}

// NetworkResources represents created network resources
type NetworkResources struct {
	Networks map[string]*docker.Network    `json:"networks"`
	Services map[string]*ServiceNetworkConfig `json:"services"`
}

// ServiceNetworkConfig represents service network configuration
type ServiceNetworkConfig struct {
	NetworkName string        `json:"network_name"`
	IPAddress   string        `json:"ip_address"`
	Aliases     []string      `json:"aliases"`
	Ports       []ServicePort `json:"ports"`
}

// ServiceNetworkConfiguration represents the complete service network setup
type ServiceNetworkConfiguration struct {
	ContentAPI    ServiceNetworkSpec `json:"content_api"`
	ServicesAPI   ServiceNetworkSpec `json:"services_api"`
	PublicGateway ServiceNetworkSpec `json:"public_gateway"`
	AdminGateway  ServiceNetworkSpec `json:"admin_gateway"`
	PostgreSQL    ServiceNetworkSpec `json:"postgresql"`
	Redis         ServiceNetworkSpec `json:"redis"`
	Azurite       ServiceNetworkSpec `json:"azurite"`
	Vault         ServiceNetworkSpec `json:"vault"`
	Grafana       ServiceNetworkSpec `json:"grafana"`
	Loki          ServiceNetworkSpec `json:"loki"`
	DaprPlacement ServiceNetworkSpec `json:"dapr_placement"`
}