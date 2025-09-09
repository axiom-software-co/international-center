package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

type InfrastructureConfig struct {
	Environment       Environment
	ProjectName       string
	Stack             StackConfig
	Deployment        DeploymentConfig
	Resources         ResourceConfig
	Security          SecurityConfig
	Compliance        ComplianceConfig
	Backup            BackupConfig
	Monitoring        MonitoringConfig
}

type StackConfig struct {
	Name               string
	Region             string
	Organization       string
	BackendURL         string
	StateStorage       string
	EncryptionEnabled  bool
	AccessLogsEnabled  bool
	Tags               map[string]string
}

type DeploymentConfig struct {
	Strategy           string
	TimeoutMinutes     int
	ParallelDeployment bool
	MaxRetries         int
	RetryDelay         time.Duration
	RequireApproval    bool
	AutoRollback       bool
	ValidationEnabled  bool
	DriftDetection     bool
}

type ResourceConfig struct {
	Database    DatabaseResourceConfig
	Storage     StorageResourceConfig
	Messaging   MessagingResourceConfig
	Vault       VaultResourceConfig
	Networking  NetworkingResourceConfig
	Compute     ComputeResourceConfig
}

type DatabaseResourceConfig struct {
	Provider          string
	InstanceClass     string
	StorageSize       int
	BackupRetention   int
	MultiAZ           bool
	EncryptionEnabled bool
	MonitoringEnabled bool
	MaintenanceWindow string
}

type StorageResourceConfig struct {
	Provider          string
	StorageClass      string
	ReplicationEnabled bool
	VersioningEnabled bool
	LifecycleEnabled  bool
	EncryptionEnabled bool
	AccessLogging     bool
	TransferAcceleration bool
}

type MessagingResourceConfig struct {
	Provider          string
	InstanceType      string
	ClusterMode       bool
	ReplicationFactor int
	RetentionHours    int
	CompressionType   string
	EncryptionEnabled bool
	AuthenticationEnabled bool
}

type VaultResourceConfig struct {
	Provider          string
	HAEnabled         bool
	SealType          string
	StorageBackend    string
	AuditEnabled      bool
	PerformanceMode   string
	ThroughputMode    string
	UIEnabled         bool
}

type NetworkingResourceConfig struct {
	VPCEnabled        bool
	SubnetConfiguration string
	NATGatewayEnabled bool
	VPNEnabled        bool
	LoadBalancerType  string
	CDNEnabled        bool
	WAFEnabled        bool
	DDoSProtection    bool
}

type ComputeResourceConfig struct {
	Provider          string
	DefaultInstanceType string
	AutoScalingEnabled bool
	ContainerOrchestration string
	ServerlessEnabled bool
	GPUSupport        bool
	SpotInstancesEnabled bool
	ReservedInstancesEnabled bool
}

type SecurityConfig struct {
	EncryptionAtRest     bool
	EncryptionInTransit  bool
	KeyManagement        string
	CertificateManagement string
	AccessControlEnabled bool
	AuditLoggingEnabled  bool
	VulnerabilityScanning bool
	ComplianceFrameworks []string
	SecretRotationEnabled bool
	NetworkPolicyEnabled bool
}

type ComplianceConfig struct {
	DataRetentionPeriod   time.Duration
	AuditLogRetention     time.Duration
	BackupEncryption      bool
	DataClassification    string
	PrivacyControlsEnabled bool
	RegulatoryFrameworks  []string
	DataResidencyRegion   string
	PIIProtectionEnabled  bool
}

type BackupConfig struct {
	Enabled               bool
	Frequency             string
	RetentionPeriod       time.Duration
	CrossRegionReplication bool
	PointInTimeRecovery   bool
	BackupEncryption      bool
	BackupCompression     bool
	TestRestoreEnabled    bool
}

type MonitoringConfig struct {
	MetricsEnabled        bool
	LoggingEnabled        bool
	TracingEnabled        bool
	AlertingEnabled       bool
	DashboardsEnabled     bool
	SLIMonitoringEnabled  bool
	ErrorTrackingEnabled  bool
	PerformanceMonitoring bool
}

func LoadConfig() (*InfrastructureConfig, error) {
	environment := getEnvironment()
	
	config := &InfrastructureConfig{
		Environment: environment,
		ProjectName: getEnvString("PROJECT_NAME", "international-center"),
	}
	
	var err error
	config.Stack, err = loadStackConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load stack config: %w", err)
	}
	
	config.Deployment, err = loadDeploymentConfig(environment)
	if err != nil {
		return nil, fmt.Errorf("failed to load deployment config: %w", err)
	}
	
	config.Resources, err = loadResourceConfig(environment)
	if err != nil {
		return nil, fmt.Errorf("failed to load resource config: %w", err)
	}
	
	config.Security, err = loadSecurityConfig(environment)
	if err != nil {
		return nil, fmt.Errorf("failed to load security config: %w", err)
	}
	
	config.Compliance, err = loadComplianceConfig(environment)
	if err != nil {
		return nil, fmt.Errorf("failed to load compliance config: %w", err)
	}
	
	config.Backup, err = loadBackupConfig(environment)
	if err != nil {
		return nil, fmt.Errorf("failed to load backup config: %w", err)
	}
	
	config.Monitoring, err = loadMonitoringConfig(environment)
	if err != nil {
		return nil, fmt.Errorf("failed to load monitoring config: %w", err)
	}
	
	return config, nil
}

func getEnvironment() Environment {
	env := os.Getenv("ENVIRONMENT")
	switch env {
	case "development":
		return Development
	case "staging":
		return Staging
	case "production":
		return Production
	default:
		return Development
	}
}

func loadStackConfig() (StackConfig, error) {
	return StackConfig{
		Name:              getEnvString("PULUMI_STACK_NAME", "dev"),
		Region:            getEnvString("PULUMI_REGION", "us-central1"),
		Organization:      getEnvString("PULUMI_ORGANIZATION", "axiom-software"),
		BackendURL:        getEnvString("PULUMI_BACKEND_URL", ""),
		StateStorage:      getEnvString("PULUMI_STATE_STORAGE", "azure-blob"),
		EncryptionEnabled: getEnvBool("PULUMI_ENCRYPTION_ENABLED", true),
		AccessLogsEnabled: getEnvBool("PULUMI_ACCESS_LOGS_ENABLED", true),
		Tags: map[string]string{
			"project":     getEnvString("PROJECT_NAME", "international-center"),
			"environment": getEnvString("ENVIRONMENT", "development"),
			"managed-by":  "pulumi",
		},
	}, nil
}

func loadDeploymentConfig(environment Environment) (DeploymentConfig, error) {
	config := DeploymentConfig{
		MaxRetries:         getEnvInt("DEPLOYMENT_MAX_RETRIES", 3),
		RetryDelay:         getEnvDuration("DEPLOYMENT_RETRY_DELAY", 30*time.Second),
		ParallelDeployment: getEnvBool("DEPLOYMENT_PARALLEL", true),
		ValidationEnabled:  getEnvBool("DEPLOYMENT_VALIDATION_ENABLED", true),
		DriftDetection:     getEnvBool("DEPLOYMENT_DRIFT_DETECTION", true),
	}
	
	switch environment {
	case Development:
		config.Strategy = "aggressive"
		config.TimeoutMinutes = 15
		config.RequireApproval = false
		config.AutoRollback = true
	case Staging:
		config.Strategy = "careful_validation"
		config.TimeoutMinutes = 30
		config.RequireApproval = false
		config.AutoRollback = true
	case Production:
		config.Strategy = "conservative_extensive_validation"
		config.TimeoutMinutes = 60
		config.RequireApproval = true
		config.AutoRollback = false
	}
	
	return config, nil
}

func loadResourceConfig(environment Environment) (ResourceConfig, error) {
	// Base configuration
	config := ResourceConfig{
		Database: DatabaseResourceConfig{
			Provider:          getEnvString("DATABASE_PROVIDER", "postgresql"),
			BackupRetention:   getEnvInt("DATABASE_BACKUP_RETENTION", 7),
			EncryptionEnabled: getEnvBool("DATABASE_ENCRYPTION_ENABLED", true),
			MonitoringEnabled: getEnvBool("DATABASE_MONITORING_ENABLED", true),
		},
		Storage: StorageResourceConfig{
			Provider:          getEnvString("STORAGE_PROVIDER", "azure_blob"),
			VersioningEnabled: getEnvBool("STORAGE_VERSIONING_ENABLED", true),
			EncryptionEnabled: getEnvBool("STORAGE_ENCRYPTION_ENABLED", true),
			AccessLogging:     getEnvBool("STORAGE_ACCESS_LOGGING", true),
		},
		Messaging: MessagingResourceConfig{
			Provider:          getEnvString("MESSAGING_PROVIDER", "rabbitmq"),
			ReplicationFactor: getEnvInt("MESSAGING_REPLICATION_FACTOR", 3),
			RetentionHours:    getEnvInt("MESSAGING_RETENTION_HOURS", 24),
			EncryptionEnabled: getEnvBool("MESSAGING_ENCRYPTION_ENABLED", true),
		},
		Vault: VaultResourceConfig{
			Provider:     getEnvString("VAULT_PROVIDER", "hashicorp_vault"),
			AuditEnabled: getEnvBool("VAULT_AUDIT_ENABLED", true),
			UIEnabled:    getEnvBool("VAULT_UI_ENABLED", true),
		},
	}
	
	// Environment-specific adjustments
	switch environment {
	case Development:
		config.Database.InstanceClass = "db.t3.micro"
		config.Database.StorageSize = 20
		config.Database.MultiAZ = false
		config.Storage.StorageClass = "standard"
		config.Storage.ReplicationEnabled = false
		config.Messaging.InstanceType = "t3.micro"
		config.Messaging.ClusterMode = false
		config.Vault.HAEnabled = false
	case Staging:
		config.Database.InstanceClass = "db.t3.small"
		config.Database.StorageSize = 100
		config.Database.MultiAZ = true
		config.Storage.StorageClass = "standard"
		config.Storage.ReplicationEnabled = true
		config.Messaging.InstanceType = "t3.small"
		config.Messaging.ClusterMode = true
		config.Vault.HAEnabled = true
	case Production:
		config.Database.InstanceClass = "db.r5.large"
		config.Database.StorageSize = 500
		config.Database.MultiAZ = true
		config.Storage.StorageClass = "premium"
		config.Storage.ReplicationEnabled = true
		config.Messaging.InstanceType = "m5.large"
		config.Messaging.ClusterMode = true
		config.Vault.HAEnabled = true
	}
	
	return config, nil
}

func loadSecurityConfig(environment Environment) (SecurityConfig, error) {
	config := SecurityConfig{
		KeyManagement:         getEnvString("SECURITY_KEY_MANAGEMENT", "azure_key_vault"),
		CertificateManagement: getEnvString("SECURITY_CERTIFICATE_MANAGEMENT", "lets_encrypt"),
		ComplianceFrameworks:  getEnvStringSlice("SECURITY_COMPLIANCE_FRAMEWORKS", []string{"SOC2"}),
	}
	
	switch environment {
	case Development:
		config.EncryptionAtRest = false
		config.EncryptionInTransit = false
		config.AccessControlEnabled = false
		config.AuditLoggingEnabled = false
		config.VulnerabilityScanning = false
		config.SecretRotationEnabled = false
		config.NetworkPolicyEnabled = false
	case Staging:
		config.EncryptionAtRest = true
		config.EncryptionInTransit = true
		config.AccessControlEnabled = true
		config.AuditLoggingEnabled = true
		config.VulnerabilityScanning = true
		config.SecretRotationEnabled = true
		config.NetworkPolicyEnabled = true
	case Production:
		config.EncryptionAtRest = true
		config.EncryptionInTransit = true
		config.AccessControlEnabled = true
		config.AuditLoggingEnabled = true
		config.VulnerabilityScanning = true
		config.SecretRotationEnabled = true
		config.NetworkPolicyEnabled = true
		config.ComplianceFrameworks = []string{"SOC2", "HIPAA", "PCI-DSS"}
	}
	
	return config, nil
}

func loadComplianceConfig(environment Environment) (ComplianceConfig, error) {
	config := ComplianceConfig{
		DataClassification:    getEnvString("COMPLIANCE_DATA_CLASSIFICATION", "internal"),
		DataResidencyRegion:   getEnvString("COMPLIANCE_DATA_RESIDENCY_REGION", "us-central1"),
		PIIProtectionEnabled:  getEnvBool("COMPLIANCE_PII_PROTECTION_ENABLED", true),
		PrivacyControlsEnabled: getEnvBool("COMPLIANCE_PRIVACY_CONTROLS_ENABLED", true),
	}
	
	switch environment {
	case Development:
		config.DataRetentionPeriod = 30 * 24 * time.Hour
		config.AuditLogRetention = 7 * 24 * time.Hour
		config.BackupEncryption = false
		config.RegulatoryFrameworks = []string{}
	case Staging:
		config.DataRetentionPeriod = 90 * 24 * time.Hour
		config.AuditLogRetention = 30 * 24 * time.Hour
		config.BackupEncryption = true
		config.RegulatoryFrameworks = []string{"SOC2"}
	case Production:
		config.DataRetentionPeriod = 7 * 365 * 24 * time.Hour
		config.AuditLogRetention = 365 * 24 * time.Hour
		config.BackupEncryption = true
		config.RegulatoryFrameworks = []string{"SOC2", "HIPAA", "GDPR"}
	}
	
	return config, nil
}

func loadBackupConfig(environment Environment) (BackupConfig, error) {
	config := BackupConfig{
		BackupCompression:  getEnvBool("BACKUP_COMPRESSION", true),
		TestRestoreEnabled: getEnvBool("BACKUP_TEST_RESTORE_ENABLED", true),
	}
	
	switch environment {
	case Development:
		config.Enabled = false
		config.Frequency = "daily"
		config.RetentionPeriod = 7 * 24 * time.Hour
		config.CrossRegionReplication = false
		config.PointInTimeRecovery = false
		config.BackupEncryption = false
	case Staging:
		config.Enabled = true
		config.Frequency = "every-12h"
		config.RetentionPeriod = 30 * 24 * time.Hour
		config.CrossRegionReplication = false
		config.PointInTimeRecovery = true
		config.BackupEncryption = true
	case Production:
		config.Enabled = true
		config.Frequency = "every-6h"
		config.RetentionPeriod = 365 * 24 * time.Hour
		config.CrossRegionReplication = true
		config.PointInTimeRecovery = true
		config.BackupEncryption = true
	}
	
	return config, nil
}

func loadMonitoringConfig(environment Environment) (MonitoringConfig, error) {
	return MonitoringConfig{
		MetricsEnabled:        getEnvBool("MONITORING_METRICS_ENABLED", true),
		LoggingEnabled:        getEnvBool("MONITORING_LOGGING_ENABLED", true),
		TracingEnabled:        getEnvBool("MONITORING_TRACING_ENABLED", true),
		AlertingEnabled:       getEnvBool("MONITORING_ALERTING_ENABLED", true),
		DashboardsEnabled:     getEnvBool("MONITORING_DASHBOARDS_ENABLED", true),
		SLIMonitoringEnabled:  getEnvBool("MONITORING_SLI_ENABLED", true),
		ErrorTrackingEnabled:  getEnvBool("MONITORING_ERROR_TRACKING_ENABLED", true),
		PerformanceMonitoring: getEnvBool("MONITORING_PERFORMANCE_ENABLED", true),
	}, nil
}

// Utility functions (same as operations config)
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return []string{value}
	}
	return defaultValue
}

func (c *InfrastructureConfig) IsProduction() bool {
	return c.Environment == Production
}

func (c *InfrastructureConfig) IsDevelopment() bool {
	return c.Environment == Development
}

func (c *InfrastructureConfig) IsStaging() bool {
	return c.Environment == Staging
}

func (c *InfrastructureConfig) GetEnvironmentString() string {
	return string(c.Environment)
}