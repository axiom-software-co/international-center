package config

import (
	"fmt"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// DeploymentConfig holds deployment-wide configuration
type DeploymentConfig struct {
	Environment    Environment      `json:"environment"`
	ProjectName    string          `json:"project_name"`
	StackName      string          `json:"stack_name"`
	Version        string          `json:"version"`
	DeploymentID   string          `json:"deployment_id"`
	
	// Infrastructure configuration
	Infrastructure InfrastructureConfig `json:"infrastructure"`
	
	// Application configuration
	Applications ApplicationsConfig `json:"applications"`
	
	// Migration configuration
	Migration MigrationConfig `json:"migration"`
	
	// Monitoring configuration
	Monitoring MonitoringConfig `json:"monitoring"`
	
	// Security configuration
	Security SecurityConfig `json:"security"`
	
	// Resource management
	Resources ResourcesConfig `json:"resources"`
}

// InfrastructureConfig defines infrastructure deployment settings
type InfrastructureConfig struct {
	// Container runtime configuration
	ContainerRuntime ContainerRuntimeConfig `json:"container_runtime"`
	
	// Network configuration
	Network NetworkConfig `json:"network"`
	
	// Storage configuration
	Storage StorageDeploymentConfig `json:"storage"`
	
	// Database configuration
	Database DatabaseDeploymentConfig `json:"database"`
	
	// Redis configuration
	Redis RedisDeploymentConfig `json:"redis"`
	
	// Observability infrastructure
	Observability ObservabilityInfraConfig `json:"observability"`
	
	// Secrets management
	Secrets SecretsManagementConfig `json:"secrets"`
}

// ContainerRuntimeConfig defines container runtime settings
type ContainerRuntimeConfig struct {
	Runtime         string                `json:"runtime"`          // "podman" for development, "containerd" for production
	RegistryHost    string                `json:"registry_host"`
	RegistryPort    int                   `json:"registry_port"`
	ImagePullPolicy string                `json:"image_pull_policy"`
	NetworkMode     string                `json:"network_mode"`
	LogDriver       string                `json:"log_driver"`
	LogOptions      map[string]string     `json:"log_options"`
	ResourceLimits  ContainerResourceLimits `json:"resource_limits"`
}

// ContainerResourceLimits defines container resource constraints
type ContainerResourceLimits struct {
	DefaultCPULimit    string `json:"default_cpu_limit"`
	DefaultMemoryLimit string `json:"default_memory_limit"`
	DefaultDiskLimit   string `json:"default_disk_limit"`
	MaxCPULimit        string `json:"max_cpu_limit"`
	MaxMemoryLimit     string `json:"max_memory_limit"`
	MaxDiskLimit       string `json:"max_disk_limit"`
}

// NetworkConfig defines network deployment settings
type NetworkConfig struct {
	CIDR            string            `json:"cidr"`
	SubnetCIDRs     []string          `json:"subnet_cidrs"`
	DNSServers      []string          `json:"dns_servers"`
	EnableIPv6      bool              `json:"enable_ipv6"`
	FirewallRules   []FirewallRule    `json:"firewall_rules"`
	LoadBalancing   LoadBalancingConfig `json:"load_balancing"`
}

// FirewallRule defines network firewall rules
type FirewallRule struct {
	Name        string   `json:"name"`
	Protocol    string   `json:"protocol"`
	Ports       []string `json:"ports"`
	Sources     []string `json:"sources"`
	Destinations []string `json:"destinations"`
	Action      string   `json:"action"`
}

// LoadBalancingConfig defines load balancing settings
type LoadBalancingConfig struct {
	Enabled         bool              `json:"enabled"`
	Algorithm       string            `json:"algorithm"`
	HealthChecks    HealthCheckConfig `json:"health_checks"`
	SessionAffinity bool              `json:"session_affinity"`
}

// HealthCheckConfig defines health check settings
type HealthCheckConfig struct {
	Enabled         bool          `json:"enabled"`
	Path            string        `json:"path"`
	Port            int           `json:"port"`
	Protocol        string        `json:"protocol"`
	IntervalSeconds int           `json:"interval_seconds"`
	TimeoutSeconds  int           `json:"timeout_seconds"`
	HealthyThreshold int          `json:"healthy_threshold"`
	UnhealthyThreshold int        `json:"unhealthy_threshold"`
}

// StorageDeploymentConfig defines storage deployment settings
type StorageDeploymentConfig struct {
	Provider        string                `json:"provider"`     // "azurite" for dev, "azure-blob" for staging/prod
	Replication     string                `json:"replication"`
	AccessTier      string                `json:"access_tier"`
	BackupPolicy    BackupPolicyConfig    `json:"backup_policy"`
	LifecyclePolicy LifecyclePolicyConfig `json:"lifecycle_policy"`
}

// BackupPolicyConfig defines backup policy settings
type BackupPolicyConfig struct {
	Enabled             bool   `json:"enabled"`
	RetentionDays       int    `json:"retention_days"`
	BackupIntervalHours int    `json:"backup_interval_hours"`
	CrossRegionBackup   bool   `json:"cross_region_backup"`
}

// LifecyclePolicyConfig defines storage lifecycle policy
type LifecyclePolicyConfig struct {
	Enabled                bool `json:"enabled"`
	CoolTierAfterDays      int  `json:"cool_tier_after_days"`
	ArchiveTierAfterDays   int  `json:"archive_tier_after_days"`
	DeleteAfterDays        int  `json:"delete_after_days"`
}

// DatabaseDeploymentConfig defines database deployment settings
type DatabaseDeploymentConfig struct {
	Provider         string                `json:"provider"`    // "postgresql" for all environments
	Version          string                `json:"version"`
	SkuTier          string                `json:"sku_tier"`
	StorageGB        int                   `json:"storage_gb"`
	BackupPolicy     DatabaseBackupPolicy  `json:"backup_policy"`
	MaintenanceWindow MaintenanceWindow    `json:"maintenance_window"`
	HighAvailability  HighAvailabilityConfig `json:"high_availability"`
}

// DatabaseBackupPolicy defines database backup settings
type DatabaseBackupPolicy struct {
	Enabled                   bool `json:"enabled"`
	BackupRetentionDays       int  `json:"backup_retention_days"`
	PointInTimeRetentionDays  int  `json:"point_in_time_retention_days"`
	CrossRegionBackupEnabled  bool `json:"cross_region_backup_enabled"`
}

// MaintenanceWindow defines maintenance scheduling
type MaintenanceWindow struct {
	DayOfWeek   string `json:"day_of_week"`
	StartHour   int    `json:"start_hour"`
	StartMinute int    `json:"start_minute"`
	DurationMinutes int `json:"duration_minutes"`
}

// HighAvailabilityConfig defines HA settings
type HighAvailabilityConfig struct {
	Enabled           bool   `json:"enabled"`
	StandbyZone       string `json:"standby_zone"`
	AutoFailoverEnabled bool `json:"auto_failover_enabled"`
}

// RedisDeploymentConfig defines Redis deployment settings
type RedisDeploymentConfig struct {
	Provider         string               `json:"provider"`    // "redis" for dev, "upstash" for staging/prod
	Version          string               `json:"version"`
	SkuTier          string               `json:"sku_tier"`
	MemoryGB         int                  `json:"memory_gb"`
	MaxMemoryPolicy  string               `json:"max_memory_policy"`
	PersistencePolicy PersistencePolicyConfig `json:"persistence_policy"`
	ClusterConfig    ClusterConfig        `json:"cluster_config"`
}

// PersistencePolicyConfig defines Redis persistence settings
type PersistencePolicyConfig struct {
	Enabled       bool   `json:"enabled"`
	BackupPolicy  string `json:"backup_policy"`
	SnapshotPolicy string `json:"snapshot_policy"`
}

// ClusterConfig defines Redis cluster settings
type ClusterConfig struct {
	Enabled       bool `json:"enabled"`
	ShardCount    int  `json:"shard_count"`
	ReplicaCount  int  `json:"replica_count"`
}

// ObservabilityInfraConfig defines observability infrastructure settings
type ObservabilityInfraConfig struct {
	Provider            string                 `json:"provider"`    // "local" for dev, "grafana-cloud" for staging/prod
	MetricsStorage      MetricsStorageConfig   `json:"metrics_storage"`
	LogsStorage         LogsStorageConfig      `json:"logs_storage"`
	TracesStorage       TracesStorageConfig    `json:"traces_storage"`
	AlertingConfig      AlertingConfig         `json:"alerting"`
}

// MetricsStorageConfig defines metrics storage settings
type MetricsStorageConfig struct {
	RetentionDays       int    `json:"retention_days"`
	StorageClass        string `json:"storage_class"`
	CompressionEnabled  bool   `json:"compression_enabled"`
}

// LogsStorageConfig defines logs storage settings
type LogsStorageConfig struct {
	RetentionDays       int    `json:"retention_days"`
	StorageClass        string `json:"storage_class"`
	CompressionEnabled  bool   `json:"compression_enabled"`
	IndexingEnabled     bool   `json:"indexing_enabled"`
}

// TracesStorageConfig defines traces storage settings
type TracesStorageConfig struct {
	RetentionDays       int    `json:"retention_days"`
	SamplingRate        float64 `json:"sampling_rate"`
	StorageClass        string `json:"storage_class"`
}

// AlertingConfig defines alerting settings
type AlertingConfig struct {
	Enabled           bool              `json:"enabled"`
	NotificationChannels []string       `json:"notification_channels"`
	AlertRules        []AlertRule       `json:"alert_rules"`
}

// AlertRule defines alerting rules
type AlertRule struct {
	Name        string            `json:"name"`
	Query       string            `json:"query"`
	Threshold   float64           `json:"threshold"`
	Comparison  string            `json:"comparison"`
	Duration    string            `json:"duration"`
	Labels      map[string]string `json:"labels"`
}

// SecretsManagementConfig defines secrets management settings
type SecretsManagementConfig struct {
	Provider        string                 `json:"provider"`    // "vault" for all environments
	VaultAddress    string                 `json:"vault_address"`
	AuthMethod      string                 `json:"auth_method"`
	SecretPaths     map[string]string      `json:"secret_paths"`
	RotationPolicy  RotationPolicyConfig   `json:"rotation_policy"`
}

// RotationPolicyConfig defines secret rotation policy
type RotationPolicyConfig struct {
	Enabled         bool `json:"enabled"`
	IntervalDays    int  `json:"interval_days"`
	AutoRotate      bool `json:"auto_rotate"`
}

// ApplicationsConfig defines application deployment settings
type ApplicationsConfig struct {
	ContentAPI    ApplicationConfig `json:"content_api"`
	ServicesAPI   ApplicationConfig `json:"services_api"`
	PublicGateway ApplicationConfig `json:"public_gateway"`
	AdminGateway  ApplicationConfig `json:"admin_gateway"`
	
	// Common application settings
	CommonSettings CommonApplicationSettings `json:"common_settings"`
}

// ApplicationConfig defines individual application deployment settings
type ApplicationConfig struct {
	Enabled         bool                    `json:"enabled"`
	Replicas        int                     `json:"replicas"`
	ImageTag        string                  `json:"image_tag"`
	Resources       ApplicationResources    `json:"resources"`
	HealthCheck     ApplicationHealthCheck  `json:"health_check"`
	Environment     map[string]string       `json:"environment"`
	Ports           []ApplicationPort       `json:"ports"`
	Volumes         []ApplicationVolume     `json:"volumes"`
}

// ApplicationResources defines application resource requirements
type ApplicationResources struct {
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
	StorageRequest string `json:"storage_request"`
}

// ApplicationHealthCheck defines application health check settings
type ApplicationHealthCheck struct {
	Enabled             bool   `json:"enabled"`
	Path                string `json:"path"`
	Port                int    `json:"port"`
	InitialDelaySeconds int    `json:"initial_delay_seconds"`
	PeriodSeconds       int    `json:"period_seconds"`
	TimeoutSeconds      int    `json:"timeout_seconds"`
	FailureThreshold    int    `json:"failure_threshold"`
	SuccessThreshold    int    `json:"success_threshold"`
}

// ApplicationPort defines application port settings
type ApplicationPort struct {
	Name     string `json:"name"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Expose   bool   `json:"expose"`
}

// ApplicationVolume defines application volume settings
type ApplicationVolume struct {
	Name       string `json:"name"`
	MountPath  string `json:"mount_path"`
	VolumeType string `json:"volume_type"`
	Size       string `json:"size"`
	ReadOnly   bool   `json:"read_only"`
}

// CommonApplicationSettings defines settings shared across applications
type CommonApplicationSettings struct {
	ImagePullPolicy   string            `json:"image_pull_policy"`
	RestartPolicy     string            `json:"restart_policy"`
	DNSPolicy         string            `json:"dns_policy"`
	ServiceAccountName string           `json:"service_account_name"`
	SecurityContext   SecurityContext   `json:"security_context"`
	Tolerations       []Toleration      `json:"tolerations"`
	NodeSelector      map[string]string `json:"node_selector"`
}

// SecurityContext defines security context settings
type SecurityContext struct {
	RunAsNonRoot             bool   `json:"run_as_non_root"`
	RunAsUser                int64  `json:"run_as_user"`
	RunAsGroup               int64  `json:"run_as_group"`
	FSGroup                  int64  `json:"fs_group"`
	ReadOnlyRootFilesystem   bool   `json:"read_only_root_filesystem"`
	AllowPrivilegeEscalation bool   `json:"allow_privilege_escalation"`
}

// Toleration defines node toleration settings
type Toleration struct {
	Key      string `json:"key"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
	Effect   string `json:"effect"`
}

// MigrationConfig defines migration deployment settings
type MigrationConfig struct {
	Enabled         bool                  `json:"enabled"`
	Strategy        MigrationStrategy     `json:"strategy"`
	Timeout         time.Duration         `json:"timeout"`
	BackupPolicy    MigrationBackupPolicy `json:"backup_policy"`
	RollbackPolicy  RollbackPolicy        `json:"rollback_policy"`
	Validation      MigrationValidation   `json:"validation"`
}

// MigrationStrategy defines migration execution strategy
type MigrationStrategy struct {
	Type                string        `json:"type"`    // "aggressive", "careful", "conservative"
	ParallelExecution   bool          `json:"parallel_execution"`
	BatchSize           int           `json:"batch_size"`
	DelayBetweenBatches time.Duration `json:"delay_between_batches"`
	MaxRetries          int           `json:"max_retries"`
}

// MigrationBackupPolicy defines migration backup settings
type MigrationBackupPolicy struct {
	Enabled           bool   `json:"enabled"`
	BackupBeforeMigration bool `json:"backup_before_migration"`
	BackupLocation    string `json:"backup_location"`
	RetentionDays     int    `json:"retention_days"`
}

// RollbackPolicy defines rollback settings
type RollbackPolicy struct {
	Enabled              bool          `json:"enabled"`
	AutoRollbackEnabled  bool          `json:"auto_rollback_enabled"`
	RollbackTimeout      time.Duration `json:"rollback_timeout"`
	RequireApproval      bool          `json:"require_approval"`
}

// MigrationValidation defines migration validation settings
type MigrationValidation struct {
	Enabled                 bool          `json:"enabled"`
	ValidateBeforeMigration bool          `json:"validate_before_migration"`
	ValidateAfterMigration  bool          `json:"validate_after_migration"`
	ValidationTimeout       time.Duration `json:"validation_timeout"`
	FailOnValidationError   bool          `json:"fail_on_validation_error"`
}

// MonitoringConfig defines deployment monitoring settings
type MonitoringConfig struct {
	Enabled           bool                    `json:"enabled"`
	MetricsCollection MetricsCollectionConfig `json:"metrics_collection"`
	LogsCollection    LogsCollectionConfig    `json:"logs_collection"`
	TracesCollection  TracesCollectionConfig  `json:"traces_collection"`
	Dashboards        DashboardsConfig        `json:"dashboards"`
}

// MetricsCollectionConfig defines metrics collection settings
type MetricsCollectionConfig struct {
	Enabled         bool              `json:"enabled"`
	CollectionInterval time.Duration  `json:"collection_interval"`
	MetricSources   []string          `json:"metric_sources"`
	CustomMetrics   []CustomMetric    `json:"custom_metrics"`
}

// CustomMetric defines custom metric settings
type CustomMetric struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	Query       string            `json:"query"`
}

// LogsCollectionConfig defines logs collection settings
type LogsCollectionConfig struct {
	Enabled       bool              `json:"enabled"`
	LogLevel      string            `json:"log_level"`
	LogSources    []string          `json:"log_sources"`
	LogFilters    []LogFilter       `json:"log_filters"`
}

// LogFilter defines log filtering settings
type LogFilter struct {
	Field     string `json:"field"`
	Operator  string `json:"operator"`
	Value     string `json:"value"`
	Action    string `json:"action"`
}

// TracesCollectionConfig defines traces collection settings
type TracesCollectionConfig struct {
	Enabled      bool    `json:"enabled"`
	SamplingRate float64 `json:"sampling_rate"`
	TraceSources []string `json:"trace_sources"`
}

// DashboardsConfig defines dashboard settings
type DashboardsConfig struct {
	Enabled       bool     `json:"enabled"`
	DefaultDashboards []string `json:"default_dashboards"`
	CustomDashboards  []string `json:"custom_dashboards"`
}

// SecurityConfig defines deployment security settings
type SecurityConfig struct {
	Enabled            bool                      `json:"enabled"`
	NetworkPolicies    NetworkPoliciesConfig     `json:"network_policies"`
	SecurityScanning   SecurityScanningConfig    `json:"security_scanning"`
	AccessControl      AccessControlConfig       `json:"access_control"`
	Compliance         ComplianceConfig          `json:"compliance"`
}

// NetworkPoliciesConfig defines network security policies
type NetworkPoliciesConfig struct {
	Enabled       bool                `json:"enabled"`
	DefaultPolicy string              `json:"default_policy"`
	Policies      []NetworkPolicy     `json:"policies"`
}

// NetworkPolicy defines individual network policy
type NetworkPolicy struct {
	Name      string              `json:"name"`
	Selectors map[string]string   `json:"selectors"`
	Ingress   []NetworkPolicyRule `json:"ingress"`
	Egress    []NetworkPolicyRule `json:"egress"`
}

// NetworkPolicyRule defines network policy rule
type NetworkPolicyRule struct {
	Ports  []NetworkPolicyPort `json:"ports"`
	From   []NetworkPolicyPeer `json:"from"`
	To     []NetworkPolicyPeer `json:"to"`
}

// NetworkPolicyPort defines network policy port
type NetworkPolicyPort struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
}

// NetworkPolicyPeer defines network policy peer
type NetworkPolicyPeer struct {
	PodSelector       map[string]string `json:"pod_selector"`
	NamespaceSelector map[string]string `json:"namespace_selector"`
	IPBlock           IPBlockConfig     `json:"ip_block"`
}

// IPBlockConfig defines IP block configuration
type IPBlockConfig struct {
	CIDR   string   `json:"cidr"`
	Except []string `json:"except"`
}

// SecurityScanningConfig defines security scanning settings
type SecurityScanningConfig struct {
	Enabled             bool     `json:"enabled"`
	VulnerabilityScanning bool   `json:"vulnerability_scanning"`
	ComplianceScanning  bool     `json:"compliance_scanning"`
	ScanningSchedule    string   `json:"scanning_schedule"`
	ScanningTools       []string `json:"scanning_tools"`
}

// AccessControlConfig defines access control settings
type AccessControlConfig struct {
	Enabled           bool                        `json:"enabled"`
	RoleBasedAccess   RoleBasedAccessConfig       `json:"role_based_access"`
	ServiceAccounts   []ServiceAccountConfig      `json:"service_accounts"`
	APIKeyManagement  APIKeyManagementConfig      `json:"api_key_management"`
}

// RoleBasedAccessConfig defines RBAC settings
type RoleBasedAccessConfig struct {
	Enabled bool                `json:"enabled"`
	Roles   []RoleConfiguration `json:"roles"`
}

// RoleConfiguration defines role configuration
type RoleConfiguration struct {
	Name        string              `json:"name"`
	Permissions []PermissionConfig  `json:"permissions"`
	Subjects    []SubjectConfig     `json:"subjects"`
}

// PermissionConfig defines permission configuration
type PermissionConfig struct {
	APIGroups []string `json:"api_groups"`
	Resources []string `json:"resources"`
	Verbs     []string `json:"verbs"`
}

// SubjectConfig defines subject configuration
type SubjectConfig struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// ServiceAccountConfig defines service account configuration
type ServiceAccountConfig struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
}

// APIKeyManagementConfig defines API key management settings
type APIKeyManagementConfig struct {
	Enabled         bool          `json:"enabled"`
	KeyRotationDays int           `json:"key_rotation_days"`
	KeyLength       int           `json:"key_length"`
	EncryptionAlgorithm string    `json:"encryption_algorithm"`
}

// ComplianceConfig defines compliance settings
type ComplianceConfig struct {
	Enabled       bool              `json:"enabled"`
	Standards     []string          `json:"standards"`
	AuditLogging  AuditLoggingConfig `json:"audit_logging"`
	DataRetention DataRetentionConfig `json:"data_retention"`
}

// AuditLoggingConfig defines audit logging settings
type AuditLoggingConfig struct {
	Enabled       bool     `json:"enabled"`
	LogLevel      string   `json:"log_level"`
	LogDestination string  `json:"log_destination"`
	RetentionDays int      `json:"retention_days"`
	EncryptLogs   bool     `json:"encrypt_logs"`
}

// DataRetentionConfig defines data retention settings
type DataRetentionConfig struct {
	Enabled               bool `json:"enabled"`
	PersonalDataRetentionDays int `json:"personal_data_retention_days"`
	AuditDataRetentionDays    int `json:"audit_data_retention_days"`
	BackupRetentionDays       int `json:"backup_retention_days"`
}

// ResourcesConfig defines resource management settings
type ResourcesConfig struct {
	Quotas          ResourceQuotasConfig          `json:"quotas"`
	LimitRanges     []LimitRangeConfig           `json:"limit_ranges"`
	AutoScaling     AutoScalingConfig            `json:"auto_scaling"`
	CostManagement  CostManagementConfig         `json:"cost_management"`
}

// ResourceQuotasConfig defines resource quotas
type ResourceQuotasConfig struct {
	Enabled       bool              `json:"enabled"`
	CPULimit      string            `json:"cpu_limit"`
	MemoryLimit   string            `json:"memory_limit"`
	StorageLimit  string            `json:"storage_limit"`
	PodLimit      int               `json:"pod_limit"`
	ServiceLimit  int               `json:"service_limit"`
}

// LimitRangeConfig defines limit ranges
type LimitRangeConfig struct {
	Name          string                       `json:"name"`
	Type          string                       `json:"type"`
	Limits        []LimitRangeItem             `json:"limits"`
}

// LimitRangeItem defines limit range item
type LimitRangeItem struct {
	Type           string            `json:"type"`
	Default        map[string]string `json:"default"`
	DefaultRequest map[string]string `json:"default_request"`
	Max            map[string]string `json:"max"`
	Min            map[string]string `json:"min"`
}

// AutoScalingConfig defines auto-scaling settings
type AutoScalingConfig struct {
	Enabled             bool    `json:"enabled"`
	MinReplicas         int     `json:"min_replicas"`
	MaxReplicas         int     `json:"max_replicas"`
	TargetCPUUtilization int    `json:"target_cpu_utilization"`
	TargetMemoryUtilization int `json:"target_memory_utilization"`
	ScaleUpCooldown     time.Duration `json:"scale_up_cooldown"`
	ScaleDownCooldown   time.Duration `json:"scale_down_cooldown"`
}

// CostManagementConfig defines cost management settings
type CostManagementConfig struct {
	Enabled         bool              `json:"enabled"`
	BudgetLimit     float64           `json:"budget_limit"`
	BudgetAlerts    []BudgetAlert     `json:"budget_alerts"`
	CostAllocation  CostAllocationConfig `json:"cost_allocation"`
}

// BudgetAlert defines budget alert settings
type BudgetAlert struct {
	Name        string  `json:"name"`
	Threshold   float64 `json:"threshold"`
	Type        string  `json:"type"`
	Recipients  []string `json:"recipients"`
}

// CostAllocationConfig defines cost allocation settings
type CostAllocationConfig struct {
	Enabled       bool              `json:"enabled"`
	TagKeys       []string          `json:"tag_keys"`
	Departments   []string          `json:"departments"`
	CostCenters   []string          `json:"cost_centers"`
}

// NewDeploymentConfig creates a new deployment configuration
func NewDeploymentConfig(ctx *pulumi.Context, env Environment) (*DeploymentConfig, error) {
	pulumiConfig := NewPulumiConfig(ctx, env)
	envConfig, err := GetEnvironmentConfig(env)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment config: %w", err)
	}
	
	deploymentID := fmt.Sprintf("%s-%d", env, time.Now().Unix())
	
	config := &DeploymentConfig{
		Environment:  env,
		ProjectName:  envConfig.ProjectName,
		StackName:    envConfig.StackName,
		Version:      pulumiConfig.GetString("version"),
		DeploymentID: deploymentID,
		
		Infrastructure: createInfrastructureConfig(env, pulumiConfig),
		Applications:   createApplicationsConfig(env, pulumiConfig),
		Migration:      createMigrationConfig(env, pulumiConfig),
		Monitoring:     createMonitoringConfig(env, pulumiConfig),
		Security:       createSecurityConfig(env, pulumiConfig),
		Resources:      createResourcesConfig(env, pulumiConfig),
	}
	
	return config, nil
}

// Helper functions for creating configuration sections

func createInfrastructureConfig(env Environment, pc *PulumiConfig) InfrastructureConfig {
	return InfrastructureConfig{
		ContainerRuntime: createContainerRuntimeConfig(env, pc),
		Network:         createNetworkConfig(env, pc),
		Storage:         createStorageDeploymentConfig(env, pc),
		Database:        createDatabaseDeploymentConfig(env, pc),
		Redis:           createRedisDeploymentConfig(env, pc),
		Observability:   createObservabilityInfraConfig(env, pc),
		Secrets:         createSecretsManagementConfig(env, pc),
	}
}

func createContainerRuntimeConfig(env Environment, pc *PulumiConfig) ContainerRuntimeConfig {
	runtime := "podman"
	if env.IsProduction() || env.IsStaging() {
		runtime = "containerd"
	}
	
	return ContainerRuntimeConfig{
		Runtime:         runtime,
		RegistryHost:    pc.GetString("container_registry.host"),
		RegistryPort:    pc.GetInt("container_registry.port", 5000),
		ImagePullPolicy: pc.GetString("container_runtime.image_pull_policy"),
		NetworkMode:     pc.GetString("container_runtime.network_mode"),
		LogDriver:       pc.GetString("container_runtime.log_driver"),
		LogOptions:      map[string]string{
			"max-size": "10m",
			"max-file": "3",
		},
		ResourceLimits: ContainerResourceLimits{
			DefaultCPULimit:    pc.GetString("container_runtime.default_cpu_limit"),
			DefaultMemoryLimit: pc.GetString("container_runtime.default_memory_limit"),
			DefaultDiskLimit:   pc.GetString("container_runtime.default_disk_limit"),
			MaxCPULimit:        pc.GetString("container_runtime.max_cpu_limit"),
			MaxMemoryLimit:     pc.GetString("container_runtime.max_memory_limit"),
			MaxDiskLimit:       pc.GetString("container_runtime.max_disk_limit"),
		},
	}
}

func createNetworkConfig(env Environment, pc *PulumiConfig) NetworkConfig {
	return NetworkConfig{
		CIDR:        pc.GetString("network.cidr"),
		SubnetCIDRs: pc.GetStringSlice("network.subnet_cidrs", []string{}),
		DNSServers:  pc.GetStringSlice("network.dns_servers", []string{}),
		EnableIPv6:  pc.GetBool("network.enable_ipv6", false),
		FirewallRules: []FirewallRule{}, // Would be populated from configuration
		LoadBalancing: LoadBalancingConfig{
			Enabled:   pc.GetBool("load_balancing.enabled", true),
			Algorithm: pc.GetString("load_balancing.algorithm"),
			HealthChecks: HealthCheckConfig{
				Enabled:            pc.GetBool("health_checks.enabled", true),
				Path:               pc.GetString("health_checks.path"),
				Port:               pc.GetInt("health_checks.port", 8080),
				Protocol:           pc.GetString("health_checks.protocol"),
				IntervalSeconds:    pc.GetInt("health_checks.interval_seconds", 30),
				TimeoutSeconds:     pc.GetInt("health_checks.timeout_seconds", 5),
				HealthyThreshold:   pc.GetInt("health_checks.healthy_threshold", 2),
				UnhealthyThreshold: pc.GetInt("health_checks.unhealthy_threshold", 3),
			},
		},
	}
}

func createStorageDeploymentConfig(env Environment, pc *PulumiConfig) StorageDeploymentConfig {
	provider := "azurite"
	if env.IsStaging() || env.IsProduction() {
		provider = "azure-blob"
	}
	
	return StorageDeploymentConfig{
		Provider:    provider,
		Replication: pc.GetString("storage.replication"),
		AccessTier:  pc.GetString("storage.access_tier"),
		BackupPolicy: BackupPolicyConfig{
			Enabled:             pc.GetBool("storage.backup.enabled", env.IsProduction()),
			RetentionDays:       pc.GetInt("storage.backup.retention_days", 30),
			BackupIntervalHours: pc.GetInt("storage.backup.interval_hours", 24),
			CrossRegionBackup:   pc.GetBool("storage.backup.cross_region", env.IsProduction()),
		},
	}
}

func createDatabaseDeploymentConfig(env Environment, pc *PulumiConfig) DatabaseDeploymentConfig {
	return DatabaseDeploymentConfig{
		Provider:  "postgresql",
		Version:   pc.GetString("database.version"),
		SkuTier:   pc.GetString("database.sku_tier"),
		StorageGB: pc.GetInt("database.storage_gb", 100),
		BackupPolicy: DatabaseBackupPolicy{
			Enabled:                  pc.GetBool("database.backup.enabled", true),
			BackupRetentionDays:      pc.GetInt("database.backup.retention_days", 7),
			PointInTimeRetentionDays: pc.GetInt("database.backup.point_in_time_retention_days", 7),
			CrossRegionBackupEnabled: pc.GetBool("database.backup.cross_region_enabled", env.IsProduction()),
		},
		HighAvailability: HighAvailabilityConfig{
			Enabled:             pc.GetBool("database.high_availability.enabled", env.IsProduction()),
			AutoFailoverEnabled: pc.GetBool("database.high_availability.auto_failover", env.IsProduction()),
		},
	}
}

func createRedisDeploymentConfig(env Environment, pc *PulumiConfig) RedisDeploymentConfig {
	provider := "redis"
	if env.IsStaging() || env.IsProduction() {
		provider = "upstash"
	}
	
	return RedisDeploymentConfig{
		Provider:        provider,
		Version:         pc.GetString("redis.version"),
		SkuTier:         pc.GetString("redis.sku_tier"),
		MemoryGB:        pc.GetInt("redis.memory_gb", 1),
		MaxMemoryPolicy: pc.GetString("redis.max_memory_policy"),
		PersistencePolicy: PersistencePolicyConfig{
			Enabled:       pc.GetBool("redis.persistence.enabled", !env.IsDevelopment()),
			BackupPolicy:  pc.GetString("redis.persistence.backup_policy"),
			SnapshotPolicy: pc.GetString("redis.persistence.snapshot_policy"),
		},
		ClusterConfig: ClusterConfig{
			Enabled:      pc.GetBool("redis.cluster.enabled", env.IsProduction()),
			ShardCount:   pc.GetInt("redis.cluster.shard_count", 3),
			ReplicaCount: pc.GetInt("redis.cluster.replica_count", 1),
		},
	}
}

func createObservabilityInfraConfig(env Environment, pc *PulumiConfig) ObservabilityInfraConfig {
	provider := "local"
	if env.IsStaging() || env.IsProduction() {
		provider = "grafana-cloud"
	}
	
	return ObservabilityInfraConfig{
		Provider: provider,
		MetricsStorage: MetricsStorageConfig{
			RetentionDays:      pc.GetInt("observability.metrics.retention_days", 30),
			StorageClass:       pc.GetString("observability.metrics.storage_class"),
			CompressionEnabled: pc.GetBool("observability.metrics.compression_enabled", true),
		},
		LogsStorage: LogsStorageConfig{
			RetentionDays:      pc.GetInt("observability.logs.retention_days", 90),
			StorageClass:       pc.GetString("observability.logs.storage_class"),
			CompressionEnabled: pc.GetBool("observability.logs.compression_enabled", true),
			IndexingEnabled:    pc.GetBool("observability.logs.indexing_enabled", true),
		},
		TracesStorage: TracesStorageConfig{
			RetentionDays: pc.GetInt("observability.traces.retention_days", 7),
			SamplingRate:  0.1, // 10% sampling
			StorageClass:  pc.GetString("observability.traces.storage_class"),
		},
	}
}

func createSecretsManagementConfig(env Environment, pc *PulumiConfig) SecretsManagementConfig {
	return SecretsManagementConfig{
		Provider:     "vault",
		VaultAddress: pc.GetString("vault.address"),
		AuthMethod:   pc.GetString("vault.auth_method"),
		SecretPaths: map[string]string{
			"database": "secret/database",
			"redis":    "secret/redis",
			"storage":  "secret/storage",
		},
		RotationPolicy: RotationPolicyConfig{
			Enabled:      pc.GetBool("vault.rotation.enabled", env.IsProduction()),
			IntervalDays: pc.GetInt("vault.rotation.interval_days", 90),
			AutoRotate:   pc.GetBool("vault.rotation.auto_rotate", !env.IsProduction()),
		},
	}
}

func createApplicationsConfig(env Environment, pc *PulumiConfig) ApplicationsConfig {
	return ApplicationsConfig{
		ContentAPI:    createApplicationConfig("content_api", env, pc),
		ServicesAPI:   createApplicationConfig("services_api", env, pc),
		PublicGateway: createApplicationConfig("public_gateway", env, pc),
		AdminGateway:  createApplicationConfig("admin_gateway", env, pc),
		CommonSettings: CommonApplicationSettings{
			ImagePullPolicy:   pc.GetString("applications.image_pull_policy"),
			RestartPolicy:     pc.GetString("applications.restart_policy"),
			DNSPolicy:         pc.GetString("applications.dns_policy"),
			ServiceAccountName: pc.GetString("applications.service_account_name"),
			SecurityContext: SecurityContext{
				RunAsNonRoot:             pc.GetBool("applications.security_context.run_as_non_root", true),
				ReadOnlyRootFilesystem:   pc.GetBool("applications.security_context.read_only_root_filesystem", true),
				AllowPrivilegeEscalation: pc.GetBool("applications.security_context.allow_privilege_escalation", false),
			},
		},
	}
}

func createApplicationConfig(appName string, env Environment, pc *PulumiConfig) ApplicationConfig {
	replicas := 1
	if env.IsProduction() {
		replicas = 3
	} else if env.IsStaging() {
		replicas = 2
	}
	
	return ApplicationConfig{
		Enabled:  pc.GetBool(fmt.Sprintf("applications.%s.enabled", appName), true),
		Replicas: pc.GetInt(fmt.Sprintf("applications.%s.replicas", appName), replicas),
		ImageTag: pc.GetString(fmt.Sprintf("applications.%s.image_tag", appName)),
		Resources: ApplicationResources{
			CPURequest:    pc.GetString(fmt.Sprintf("applications.%s.resources.cpu_request", appName)),
			CPULimit:      pc.GetString(fmt.Sprintf("applications.%s.resources.cpu_limit", appName)),
			MemoryRequest: pc.GetString(fmt.Sprintf("applications.%s.resources.memory_request", appName)),
			MemoryLimit:   pc.GetString(fmt.Sprintf("applications.%s.resources.memory_limit", appName)),
		},
		HealthCheck: ApplicationHealthCheck{
			Enabled:             pc.GetBool(fmt.Sprintf("applications.%s.health_check.enabled", appName), true),
			Path:                pc.GetString(fmt.Sprintf("applications.%s.health_check.path", appName)),
			Port:                pc.GetInt(fmt.Sprintf("applications.%s.health_check.port", appName), 8080),
			InitialDelaySeconds: pc.GetInt(fmt.Sprintf("applications.%s.health_check.initial_delay_seconds", appName), 30),
			PeriodSeconds:       pc.GetInt(fmt.Sprintf("applications.%s.health_check.period_seconds", appName), 10),
			TimeoutSeconds:      pc.GetInt(fmt.Sprintf("applications.%s.health_check.timeout_seconds", appName), 5),
			FailureThreshold:    pc.GetInt(fmt.Sprintf("applications.%s.health_check.failure_threshold", appName), 3),
			SuccessThreshold:    pc.GetInt(fmt.Sprintf("applications.%s.health_check.success_threshold", appName), 1),
		},
	}
}

func createMigrationConfig(env Environment, pc *PulumiConfig) MigrationConfig {
	strategy := "careful"
	if env.IsDevelopment() {
		strategy = "aggressive"
	} else if env.IsProduction() {
		strategy = "conservative"
	}
	
	return MigrationConfig{
		Enabled: pc.GetBool("migration.enabled", true),
		Strategy: MigrationStrategy{
			Type:                strategy,
			ParallelExecution:   pc.GetBool("migration.parallel_execution", env.IsDevelopment()),
			BatchSize:           pc.GetInt("migration.batch_size", 10),
			DelayBetweenBatches: time.Duration(pc.GetInt("migration.delay_between_batches_seconds", 5)) * time.Second,
			MaxRetries:          pc.GetInt("migration.max_retries", 3),
		},
		Timeout: time.Duration(pc.GetInt("migration.timeout_minutes", 30)) * time.Minute,
		BackupPolicy: MigrationBackupPolicy{
			Enabled:               pc.GetBool("migration.backup.enabled", !env.IsDevelopment()),
			BackupBeforeMigration: pc.GetBool("migration.backup.before_migration", env.IsProduction()),
			RetentionDays:         pc.GetInt("migration.backup.retention_days", 30),
		},
		RollbackPolicy: RollbackPolicy{
			Enabled:             pc.GetBool("migration.rollback.enabled", true),
			AutoRollbackEnabled: pc.GetBool("migration.rollback.auto_enabled", env.IsDevelopment()),
			RollbackTimeout:     time.Duration(pc.GetInt("migration.rollback.timeout_minutes", 15)) * time.Minute,
			RequireApproval:     pc.GetBool("migration.rollback.require_approval", env.IsProduction()),
		},
	}
}

func createMonitoringConfig(env Environment, pc *PulumiConfig) MonitoringConfig {
	return MonitoringConfig{
		Enabled: pc.GetBool("monitoring.enabled", true),
		MetricsCollection: MetricsCollectionConfig{
			Enabled:            pc.GetBool("monitoring.metrics.enabled", true),
			CollectionInterval: time.Duration(pc.GetInt("monitoring.metrics.collection_interval_seconds", 30)) * time.Second,
			MetricSources:      pc.GetStringSlice("monitoring.metrics.sources", []string{"applications", "infrastructure"}),
		},
		LogsCollection: LogsCollectionConfig{
			Enabled:    pc.GetBool("monitoring.logs.enabled", true),
			LogLevel:   pc.GetString("monitoring.logs.level"),
			LogSources: pc.GetStringSlice("monitoring.logs.sources", []string{"applications", "infrastructure"}),
		},
		TracesCollection: TracesCollectionConfig{
			Enabled:      pc.GetBool("monitoring.traces.enabled", true),
			SamplingRate: 0.1, // 10% sampling
			TraceSources: pc.GetStringSlice("monitoring.traces.sources", []string{"applications"}),
		},
	}
}

func createSecurityConfig(env Environment, pc *PulumiConfig) SecurityConfig {
	return SecurityConfig{
		Enabled: pc.GetBool("security.enabled", true),
		NetworkPolicies: NetworkPoliciesConfig{
			Enabled:       pc.GetBool("security.network_policies.enabled", !env.IsDevelopment()),
			DefaultPolicy: pc.GetString("security.network_policies.default_policy"),
		},
		SecurityScanning: SecurityScanningConfig{
			Enabled:               pc.GetBool("security.scanning.enabled", env.IsProduction()),
			VulnerabilityScanning: pc.GetBool("security.scanning.vulnerability_scanning", true),
			ComplianceScanning:    pc.GetBool("security.scanning.compliance_scanning", env.IsProduction()),
			ScanningSchedule:      pc.GetString("security.scanning.schedule"),
		},
		AccessControl: AccessControlConfig{
			Enabled: pc.GetBool("security.access_control.enabled", true),
			RoleBasedAccess: RoleBasedAccessConfig{
				Enabled: pc.GetBool("security.rbac.enabled", true),
			},
		},
		Compliance: ComplianceConfig{
			Enabled: pc.GetBool("security.compliance.enabled", env.IsProduction()),
			AuditLogging: AuditLoggingConfig{
				Enabled:       pc.GetBool("security.audit_logging.enabled", true),
				LogLevel:      pc.GetString("security.audit_logging.level"),
				RetentionDays: pc.GetInt("security.audit_logging.retention_days", 2555), // 7 years for compliance
				EncryptLogs:   pc.GetBool("security.audit_logging.encrypt_logs", env.IsProduction()),
			},
		},
	}
}

func createResourcesConfig(env Environment, pc *PulumiConfig) ResourcesConfig {
	return ResourcesConfig{
		Quotas: ResourceQuotasConfig{
			Enabled:      pc.GetBool("resources.quotas.enabled", !env.IsDevelopment()),
			CPULimit:     pc.GetString("resources.quotas.cpu_limit"),
			MemoryLimit:  pc.GetString("resources.quotas.memory_limit"),
			StorageLimit: pc.GetString("resources.quotas.storage_limit"),
		},
		AutoScaling: AutoScalingConfig{
			Enabled:                 pc.GetBool("resources.auto_scaling.enabled", env.IsProduction()),
			MinReplicas:             pc.GetInt("resources.auto_scaling.min_replicas", 1),
			MaxReplicas:             pc.GetInt("resources.auto_scaling.max_replicas", 10),
			TargetCPUUtilization:    pc.GetInt("resources.auto_scaling.target_cpu_utilization", 70),
			TargetMemoryUtilization: pc.GetInt("resources.auto_scaling.target_memory_utilization", 80),
			ScaleUpCooldown:         time.Duration(pc.GetInt("resources.auto_scaling.scale_up_cooldown_seconds", 300)) * time.Second,
			ScaleDownCooldown:       time.Duration(pc.GetInt("resources.auto_scaling.scale_down_cooldown_seconds", 600)) * time.Second,
		},
		CostManagement: CostManagementConfig{
			Enabled:     pc.GetBool("resources.cost_management.enabled", env.IsProduction()),
			BudgetLimit: 1000.0, // Default budget limit
		},
	}
}