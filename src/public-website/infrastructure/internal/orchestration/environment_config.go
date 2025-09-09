package orchestration

import (
	"fmt"
	"time"
)

type EnvironmentConfig struct {
	Name                    string
	DeploymentStrategy      string
	MigrationStrategy       string
	RollbackStrategy        string
	SafetyChecks           SafetyLevel
	AutomationLevel        AutomationLevel
	ResourceLimits         ResourceLimits
	SecurityPolicy         SecurityPolicy
	MonitoringConfig       MonitoringConfig
	ScalingConfiguration   ScalingConfig
	BackupConfiguration    BackupConfig
	ComplianceRequirements ComplianceRequirements
}

type SafetyLevel string

const (
	SafetyMinimal   SafetyLevel = "minimal"
	SafetyModerate  SafetyLevel = "moderate"
	SafetyExtensive SafetyLevel = "extensive"
)

type AutomationLevel string

const (
	AutomationFull      AutomationLevel = "full"
	AutomationPartial   AutomationLevel = "partial"
	AutomationManual    AutomationLevel = "manual"
)

type ResourceLimits struct {
	CPULimit       string
	MemoryLimit    string
	StorageLimit   string
	NetworkLimit   string
	ConcurrentOps  int
	TimeoutMinutes int
}

type SecurityPolicy struct {
	EncryptionRequired     bool
	AccessControlEnabled   bool
	AuditLoggingRequired   bool
	VulnerabilityScanning  bool
	SecretRotationEnabled  bool
	NetworkPolicyEnforced  bool
}

type MonitoringConfig struct {
	MetricsEnabled        bool
	LoggingLevel         string
	AlertingEnabled      bool
	HealthCheckInterval  time.Duration
	PerformanceTracking  bool
	ErrorTrackingEnabled bool
}

type ScalingConfig struct {
	AutoScalingEnabled bool
	MinReplicas        int
	MaxReplicas        int
	CPUThreshold       int
	MemoryThreshold    int
	ScaleUpCooldown    time.Duration
	ScaleDownCooldown  time.Duration
}

type BackupConfig struct {
	Enabled             bool
	Frequency           time.Duration
	RetentionPeriod     time.Duration
	CrossRegionBackup   bool
	PointInTimeRecovery bool
}

type ComplianceRequirements struct {
	AuditingRequired        bool
	DataRetentionPeriod     time.Duration
	EncryptionAtRest        bool
	EncryptionInTransit     bool
	AccessControlAuditing   bool
	ComplianceFrameworks    []string
}

func GetEnvironmentConfig(environment string) (*EnvironmentConfig, error) {
	switch environment {
	case "development":
		return getDevelopmentConfig(), nil
	case "staging":
		return getStagingConfig(), nil
	case "production":
		return getProductionConfig(), nil
	default:
		return nil, fmt.Errorf("unsupported environment: %s", environment)
	}
}

func getDevelopmentConfig() *EnvironmentConfig {
	return &EnvironmentConfig{
		Name:               "development",
		DeploymentStrategy: "aggressive",
		MigrationStrategy:  "always_migrate_latest",
		RollbackStrategy:   "easy_destroy_recreate",
		SafetyChecks:       SafetyMinimal,
		AutomationLevel:    AutomationFull,
		ResourceLimits: ResourceLimits{
			CPULimit:       "2",
			MemoryLimit:    "4Gi",
			StorageLimit:   "20Gi",
			NetworkLimit:   "1Gbps",
			ConcurrentOps:  10,
			TimeoutMinutes: 15,
		},
		SecurityPolicy: SecurityPolicy{
			EncryptionRequired:     false,
			AccessControlEnabled:   false,
			AuditLoggingRequired:   false,
			VulnerabilityScanning:  false,
			SecretRotationEnabled:  false,
			NetworkPolicyEnforced:  false,
		},
		MonitoringConfig: MonitoringConfig{
			MetricsEnabled:        true,
			LoggingLevel:         "debug",
			AlertingEnabled:      false,
			HealthCheckInterval:  30 * time.Second,
			PerformanceTracking:  true,
			ErrorTrackingEnabled: true,
		},
		ScalingConfiguration: ScalingConfig{
			AutoScalingEnabled: false,
			MinReplicas:        1,
			MaxReplicas:        1,
			CPUThreshold:       80,
			MemoryThreshold:    80,
			ScaleUpCooldown:    5 * time.Minute,
			ScaleDownCooldown:  3 * time.Minute,
		},
		BackupConfiguration: BackupConfig{
			Enabled:             false,
			Frequency:           24 * time.Hour,
			RetentionPeriod:     7 * 24 * time.Hour,
			CrossRegionBackup:   false,
			PointInTimeRecovery: false,
		},
		ComplianceRequirements: ComplianceRequirements{
			AuditingRequired:        false,
			DataRetentionPeriod:     30 * 24 * time.Hour,
			EncryptionAtRest:        false,
			EncryptionInTransit:     false,
			AccessControlAuditing:   false,
			ComplianceFrameworks:    []string{},
		},
	}
}

func getStagingConfig() *EnvironmentConfig {
	return &EnvironmentConfig{
		Name:               "staging",
		DeploymentStrategy: "careful_validation",
		MigrationStrategy:  "migrate_with_validation",
		RollbackStrategy:   "supported_with_confirmation",
		SafetyChecks:       SafetyModerate,
		AutomationLevel:    AutomationPartial,
		ResourceLimits: ResourceLimits{
			CPULimit:       "4",
			MemoryLimit:    "8Gi",
			StorageLimit:   "100Gi",
			NetworkLimit:   "10Gbps",
			ConcurrentOps:  25,
			TimeoutMinutes: 30,
		},
		SecurityPolicy: SecurityPolicy{
			EncryptionRequired:     true,
			AccessControlEnabled:   true,
			AuditLoggingRequired:   true,
			VulnerabilityScanning:  true,
			SecretRotationEnabled:  true,
			NetworkPolicyEnforced:  true,
		},
		MonitoringConfig: MonitoringConfig{
			MetricsEnabled:        true,
			LoggingLevel:         "info",
			AlertingEnabled:      true,
			HealthCheckInterval:  15 * time.Second,
			PerformanceTracking:  true,
			ErrorTrackingEnabled: true,
		},
		ScalingConfiguration: ScalingConfig{
			AutoScalingEnabled: true,
			MinReplicas:        2,
			MaxReplicas:        10,
			CPUThreshold:       70,
			MemoryThreshold:    70,
			ScaleUpCooldown:    3 * time.Minute,
			ScaleDownCooldown:  5 * time.Minute,
		},
		BackupConfiguration: BackupConfig{
			Enabled:             true,
			Frequency:           12 * time.Hour,
			RetentionPeriod:     30 * 24 * time.Hour,
			CrossRegionBackup:   false,
			PointInTimeRecovery: true,
		},
		ComplianceRequirements: ComplianceRequirements{
			AuditingRequired:        true,
			DataRetentionPeriod:     90 * 24 * time.Hour,
			EncryptionAtRest:        true,
			EncryptionInTransit:     true,
			AccessControlAuditing:   true,
			ComplianceFrameworks:    []string{"SOC2"},
		},
	}
}

func getProductionConfig() *EnvironmentConfig {
	return &EnvironmentConfig{
		Name:               "production",
		DeploymentStrategy: "conservative_extensive_validation",
		MigrationStrategy:  "manual_approval_required",
		RollbackStrategy:   "manual_approval_required",
		SafetyChecks:       SafetyExtensive,
		AutomationLevel:    AutomationManual,
		ResourceLimits: ResourceLimits{
			CPULimit:       "8",
			MemoryLimit:    "16Gi",
			StorageLimit:   "500Gi",
			NetworkLimit:   "100Gbps",
			ConcurrentOps:  100,
			TimeoutMinutes: 60,
		},
		SecurityPolicy: SecurityPolicy{
			EncryptionRequired:     true,
			AccessControlEnabled:   true,
			AuditLoggingRequired:   true,
			VulnerabilityScanning:  true,
			SecretRotationEnabled:  true,
			NetworkPolicyEnforced:  true,
		},
		MonitoringConfig: MonitoringConfig{
			MetricsEnabled:        true,
			LoggingLevel:         "warn",
			AlertingEnabled:      true,
			HealthCheckInterval:  10 * time.Second,
			PerformanceTracking:  true,
			ErrorTrackingEnabled: true,
		},
		ScalingConfiguration: ScalingConfig{
			AutoScalingEnabled: true,
			MinReplicas:        3,
			MaxReplicas:        50,
			CPUThreshold:       60,
			MemoryThreshold:    60,
			ScaleUpCooldown:    2 * time.Minute,
			ScaleDownCooldown:  10 * time.Minute,
		},
		BackupConfiguration: BackupConfig{
			Enabled:             true,
			Frequency:           6 * time.Hour,
			RetentionPeriod:     365 * 24 * time.Hour,
			CrossRegionBackup:   true,
			PointInTimeRecovery: true,
		},
		ComplianceRequirements: ComplianceRequirements{
			AuditingRequired:        true,
			DataRetentionPeriod:     7 * 365 * 24 * time.Hour,
			EncryptionAtRest:        true,
			EncryptionInTransit:     true,
			AccessControlAuditing:   true,
			ComplianceFrameworks:    []string{"SOC2", "HIPAA", "PCI-DSS"},
		},
	}
}

func (ec *EnvironmentConfig) IsProductionEnvironment() bool {
	return ec.Name == "production"
}

func (ec *EnvironmentConfig) RequiresManualApproval() bool {
	return ec.AutomationLevel == AutomationManual
}

func (ec *EnvironmentConfig) HasExtensiveSafetyChecks() bool {
	return ec.SafetyChecks == SafetyExtensive
}

func (ec *EnvironmentConfig) GetDeploymentTimeout() time.Duration {
	return time.Duration(ec.ResourceLimits.TimeoutMinutes) * time.Minute
}