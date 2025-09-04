package infrastructure

import (
	"context"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type StorageStack interface {
	Deploy(ctx context.Context) (StorageDeployment, error)
	CreateStorageContainers(ctx context.Context, deployment StorageDeployment) error
	GetStorageConnectionInfo() map[string]interface{}
	GetBlobStorageEndpoint() string
	GetDaprBindingConfiguration(containerName string) map[string]interface{}
	ValidateDeployment(ctx context.Context, deployment StorageDeployment) error
}

type StorageDeployment interface {
	GetPrimaryStorageEndpoint() pulumi.StringOutput
	GetBackupStorageEndpoint() pulumi.StringOutput
	GetConnectionString() pulumi.StringOutput
	GetContainerEndpoint(name string) string
	GetQueueEndpoint(name string) string
}

type StorageConfiguration struct {
	Environment           string
	StorageAccountName    string
	BackupAccountName     string
	ContainerNames        []string
	QueueNames           []string
	RedundancyType       string
	AccessTier           string
	TierTransitionDays   int
	RetentionDays        int
	EnableGeoReplication bool
	EnableBackup         bool
	PrivateEndpoints     bool
	EncryptionEnabled    bool
	LifecycleManagement  bool
}

type StorageFactory interface {
	CreateStorageStack(ctx *pulumi.Context, config *config.Config, environment string) StorageStack
}

func GetStorageConfiguration(environment string, config *config.Config) *StorageConfiguration {
	switch environment {
	case "development":
		return &StorageConfiguration{
			Environment:           "development",
			StorageAccountName:    "devstoreaccount1", // Azurite default
			BackupAccountName:     "",
			ContainerNames:        []string{"content", "backups", "logs", "temp", "uploads"},
			QueueNames:           []string{"processing", "notifications"},
			RedundancyType:       "LRS", // Local for dev
			AccessTier:           "Hot",
			TierTransitionDays:   30,
			RetentionDays:        7,
			EnableGeoReplication: false,
			EnableBackup:         false,
			PrivateEndpoints:     false,
			EncryptionEnabled:    false,
			LifecycleManagement:  false,
		}
	case "staging":
		return &StorageConfiguration{
			Environment:           "staging",
			StorageAccountName:    "internationalcenterstaging",
			BackupAccountName:     "",
			ContainerNames:        []string{"content", "media", "documents", "backups", "temp"},
			QueueNames:           []string{"content-processing", "image-processing", "document-processing", "notification-queue", "audit-events"},
			RedundancyType:       "LRS",
			AccessTier:           "Hot",
			TierTransitionDays:   30,
			RetentionDays:        35,
			EnableGeoReplication: true,
			EnableBackup:         false,
			PrivateEndpoints:     true,
			EncryptionEnabled:    true,
			LifecycleManagement:  false,
		}
	case "production":
		return &StorageConfiguration{
			Environment:           "production",
			StorageAccountName:    "intcenterproduction",
			BackupAccountName:     "intcenterprodbackup",
			ContainerNames:        []string{"content", "media", "documents", "backups", "logs", "compliance", "disaster-recovery"},
			QueueNames:           []string{"content-processing", "image-processing", "document-processing", "notification-queue", "audit-events", "compliance-events", "backup-events", "virus-scan-events"},
			RedundancyType:       "GRS", // Geo-redundant for production
			AccessTier:           "Hot",
			TierTransitionDays:   30,
			RetentionDays:        2555, // 7 years
			EnableGeoReplication: true,
			EnableBackup:         true,
			PrivateEndpoints:     true,
			EncryptionEnabled:    true,
			LifecycleManagement:  true,
		}
	default:
		return &StorageConfiguration{
			Environment:           environment,
			StorageAccountName:    "internationalcenter" + environment,
			BackupAccountName:     "",
			ContainerNames:        []string{"content", "media", "documents", "backups"},
			QueueNames:           []string{"content-processing", "notification-queue"},
			RedundancyType:       "LRS",
			AccessTier:           "Hot",
			TierTransitionDays:   30,
			RetentionDays:        30,
			EnableGeoReplication: false,
			EnableBackup:         false,
			PrivateEndpoints:     false,
			EncryptionEnabled:    true,
			LifecycleManagement:  false,
		}
	}
}

// StorageMetrics defines storage performance and cost metrics for environment-specific policies
type StorageMetrics struct {
	MaxIOPS              int
	MaxThroughputMBps    int
	MaxStorageSizeGB     int
	MaxRetentionDays     int
	CostOptimized        bool
	PerformanceOptimized bool
	ComplianceRequired   bool
}

func GetStorageMetrics(environment string) StorageMetrics {
	switch environment {
	case "development":
		return StorageMetrics{
			MaxIOPS:              100,
			MaxThroughputMBps:    10,
			MaxStorageSizeGB:     100,
			MaxRetentionDays:     7,
			CostOptimized:        true,
			PerformanceOptimized: false,
			ComplianceRequired:   false,
		}
	case "staging":
		return StorageMetrics{
			MaxIOPS:              1000,
			MaxThroughputMBps:    100,
			MaxStorageSizeGB:     1000,
			MaxRetentionDays:     90,
			CostOptimized:        true,
			PerformanceOptimized: false,
			ComplianceRequired:   true,
		}
	case "production":
		return StorageMetrics{
			MaxIOPS:              10000,
			MaxThroughputMBps:    1000,
			MaxStorageSizeGB:     10000,
			MaxRetentionDays:     3650,
			CostOptimized:        false,
			PerformanceOptimized: true,
			ComplianceRequired:   true,
		}
	default:
		return StorageMetrics{
			MaxIOPS:              500,
			MaxThroughputMBps:    50,
			MaxStorageSizeGB:     500,
			MaxRetentionDays:     30,
			CostOptimized:        true,
			PerformanceOptimized: false,
			ComplianceRequired:   false,
		}
	}
}