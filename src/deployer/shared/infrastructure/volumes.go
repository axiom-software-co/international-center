package infrastructure

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/axiom-software-co/international-center/src/deployer/shared/config"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// VolumeType represents the type of volume
type VolumeType string

const (
	VolumeTypeLocal      VolumeType = "local"
	VolumeTypeNFS        VolumeType = "nfs"
	VolumeTypeAzureFiles VolumeType = "azure-files"
	VolumeTypeTmpfs      VolumeType = "tmpfs"
	VolumeTypeSecret     VolumeType = "secret"
	VolumeTypeConfigMap  VolumeType = "configmap"
)

// VolumeAccessMode defines how a volume can be accessed
type VolumeAccessMode string

const (
	AccessModeReadOnly       VolumeAccessMode = "ro"
	AccessModeReadWrite      VolumeAccessMode = "rw"
	AccessModeReadWriteOnce  VolumeAccessMode = "rwo"
	AccessModeReadWriteMany  VolumeAccessMode = "rwx"
)

// VolumeMount represents a volume mount point
type VolumeMount struct {
	Name        string            `json:"name"`
	Source      string            `json:"source"`
	Target      string            `json:"target"`
	Type        VolumeType        `json:"type"`
	AccessMode  VolumeAccessMode  `json:"access_mode"`
	Options     map[string]string `json:"options"`
	Readonly    bool              `json:"readonly"`
	Propagation string            `json:"propagation"`
}

// VolumeConfiguration defines volume configuration for a service
type VolumeConfiguration struct {
	ServiceName   string                    `json:"service_name"`
	Volumes       map[string]VolumeMount    `json:"volumes"`
	SharedVolumes map[string]VolumeMount    `json:"shared_volumes"`
	TempVolumes   map[string]VolumeMount    `json:"temp_volumes"`
	SecretMounts  map[string]VolumeMount    `json:"secret_mounts"`
	ConfigMounts  map[string]VolumeMount    `json:"config_mounts"`
}

// VolumeManager manages volume operations across environments
type VolumeManager struct {
	config *config.DeploymentConfig
	ctx    *pulumi.Context
}

// NewVolumeManager creates a new volume manager
func NewVolumeManager(ctx *pulumi.Context, cfg *config.DeploymentConfig) *VolumeManager {
	return &VolumeManager{
		config: cfg,
		ctx:    ctx,
	}
}

// CreateVolumeConfigurations creates volume configurations for all services
func (vm *VolumeManager) CreateVolumeConfigurations() (map[string]*VolumeConfiguration, error) {
	configurations := make(map[string]*VolumeConfiguration)
	
	// PostgreSQL volumes
	configurations["postgresql"] = vm.createPostgreSQLVolumes()
	
	// Redis volumes
	configurations["redis"] = vm.createRedisVolumes()
	
	// Storage service volumes (Azurite)
	configurations["azurite"] = vm.createAzuriteVolumes()
	
	// Vault volumes
	configurations["vault"] = vm.createVaultVolumes()
	
	// Grafana volumes
	configurations["grafana"] = vm.createGrafanaVolumes()
	
	// Loki volumes
	configurations["loki"] = vm.createLokiVolumes()
	
	// Application service volumes
	applications := map[string]config.ApplicationConfig{
		"content-api":    vm.config.Applications.ContentAPI,
		"services-api":   vm.config.Applications.ServicesAPI,
		"public-gateway": vm.config.Applications.PublicGateway,
		"admin-gateway":  vm.config.Applications.AdminGateway,
	}
	for appName, appConfig := range applications {
		configurations[appName] = vm.createApplicationVolumes(appName, appConfig)
	}
	
	return configurations, nil
}

// CreateVolume creates a single volume based on configuration
func (vm *VolumeManager) CreateVolume(volumeName string, volumeMount VolumeMount) (pulumi.Resource, error) {
	switch volumeMount.Type {
	case VolumeTypeLocal:
		return vm.createLocalVolume(volumeName, volumeMount)
	case VolumeTypeAzureFiles:
		return vm.createAzureFilesVolume(volumeName, volumeMount)
	case VolumeTypeTmpfs:
		return vm.createTmpfsVolume(volumeName, volumeMount)
	default:
		return nil, fmt.Errorf("unsupported volume type: %s", volumeMount.Type)
	}
}

// Private methods for service-specific volume configurations

func (vm *VolumeManager) createPostgreSQLVolumes() *VolumeConfiguration {
	volumes := make(map[string]VolumeMount)
	secretMounts := make(map[string]VolumeMount)
	
	// Data volume
	volumes["data"] = VolumeMount{
		Name:       "postgresql-data",
		Source:     vm.getVolumePath("postgresql", "data"),
		Target:     "/var/lib/postgresql/data",
		Type:       vm.getVolumeType("postgresql"),
		AccessMode: AccessModeReadWrite,
		Options:    vm.getVolumeOptions("postgresql"),
	}
	
	// Configuration volume
	volumes["config"] = VolumeMount{
		Name:       "postgresql-config",
		Source:     vm.getVolumePath("postgresql", "config"),
		Target:     "/etc/postgresql",
		Type:       VolumeTypeLocal,
		AccessMode: AccessModeReadOnly,
	}
	
	// Secrets mount
	secretMounts["password"] = VolumeMount{
		Name:       "postgresql-password",
		Source:     "postgresql-password-secret",
		Target:     "/var/secrets/postgresql",
		Type:       VolumeTypeSecret,
		AccessMode: AccessModeReadOnly,
		Readonly:   true,
	}
	
	return &VolumeConfiguration{
		ServiceName:  "postgresql",
		Volumes:      volumes,
		SecretMounts: secretMounts,
	}
}

func (vm *VolumeManager) createRedisVolumes() *VolumeConfiguration {
	volumes := make(map[string]VolumeMount)
	secretMounts := make(map[string]VolumeMount)
	
	// Data volume
	volumes["data"] = VolumeMount{
		Name:       "redis-data",
		Source:     vm.getVolumePath("redis", "data"),
		Target:     "/data",
		Type:       vm.getVolumeType("redis"),
		AccessMode: AccessModeReadWrite,
		Options:    vm.getVolumeOptions("redis"),
	}
	
	// Configuration volume
	volumes["config"] = VolumeMount{
		Name:       "redis-config",
		Source:     vm.getVolumePath("redis", "config"),
		Target:     "/usr/local/etc/redis",
		Type:       VolumeTypeLocal,
		AccessMode: AccessModeReadOnly,
	}
	
	// Secrets mount
	secretMounts["password"] = VolumeMount{
		Name:       "redis-password",
		Source:     "redis-password-secret",
		Target:     "/var/secrets/redis",
		Type:       VolumeTypeSecret,
		AccessMode: AccessModeReadOnly,
		Readonly:   true,
	}
	
	return &VolumeConfiguration{
		ServiceName:  "redis",
		Volumes:      volumes,
		SecretMounts: secretMounts,
	}
}

func (vm *VolumeManager) createAzuriteVolumes() *VolumeConfiguration {
	volumes := make(map[string]VolumeMount)
	
	// Only for development environment - staging/production use real Azure Storage
	if vm.config.Environment.IsDevelopment() {
		volumes["data"] = VolumeMount{
			Name:       "azurite-data",
			Source:     vm.getVolumePath("azurite", "data"),
			Target:     "/data",
			Type:       VolumeTypeLocal,
			AccessMode: AccessModeReadWrite,
		}
	}
	
	return &VolumeConfiguration{
		ServiceName: "azurite",
		Volumes:     volumes,
	}
}

func (vm *VolumeManager) createVaultVolumes() *VolumeConfiguration {
	volumes := make(map[string]VolumeMount)
	secretMounts := make(map[string]VolumeMount)
	
	// Data volume
	volumes["data"] = VolumeMount{
		Name:       "vault-data",
		Source:     vm.getVolumePath("vault", "data"),
		Target:     "/vault/data",
		Type:       vm.getVolumeType("vault"),
		AccessMode: AccessModeReadWrite,
		Options:    vm.getVolumeOptions("vault"),
	}
	
	// Configuration volume
	volumes["config"] = VolumeMount{
		Name:       "vault-config",
		Source:     vm.getVolumePath("vault", "config"),
		Target:     "/vault/config",
		Type:       VolumeTypeLocal,
		AccessMode: AccessModeReadOnly,
	}
	
	// TLS certificates
	secretMounts["tls"] = VolumeMount{
		Name:       "vault-tls",
		Source:     "vault-tls-secret",
		Target:     "/vault/tls",
		Type:       VolumeTypeSecret,
		AccessMode: AccessModeReadOnly,
		Readonly:   true,
	}
	
	return &VolumeConfiguration{
		ServiceName:  "vault",
		Volumes:      volumes,
		SecretMounts: secretMounts,
	}
}

func (vm *VolumeManager) createGrafanaVolumes() *VolumeConfiguration {
	volumes := make(map[string]VolumeMount)
	secretMounts := make(map[string]VolumeMount)
	configMounts := make(map[string]VolumeMount)
	
	// Data volume
	volumes["data"] = VolumeMount{
		Name:       "grafana-data",
		Source:     vm.getVolumePath("grafana", "data"),
		Target:     "/var/lib/grafana",
		Type:       vm.getVolumeType("grafana"),
		AccessMode: AccessModeReadWrite,
		Options:    vm.getVolumeOptions("grafana"),
	}
	
	// Configuration mounts
	configMounts["config"] = VolumeMount{
		Name:       "grafana-config",
		Source:     "grafana-config",
		Target:     "/etc/grafana",
		Type:       VolumeTypeConfigMap,
		AccessMode: AccessModeReadOnly,
	}
	
	// Dashboards
	configMounts["dashboards"] = VolumeMount{
		Name:       "grafana-dashboards",
		Source:     "grafana-dashboards",
		Target:     "/var/lib/grafana/dashboards",
		Type:       VolumeTypeConfigMap,
		AccessMode: AccessModeReadOnly,
	}
	
	// Secrets mount
	secretMounts["admin"] = VolumeMount{
		Name:       "grafana-admin",
		Source:     "grafana-admin-secret",
		Target:     "/var/secrets/grafana",
		Type:       VolumeTypeSecret,
		AccessMode: AccessModeReadOnly,
		Readonly:   true,
	}
	
	return &VolumeConfiguration{
		ServiceName:  "grafana",
		Volumes:      volumes,
		ConfigMounts: configMounts,
		SecretMounts: secretMounts,
	}
}

func (vm *VolumeManager) createLokiVolumes() *VolumeConfiguration {
	volumes := make(map[string]VolumeMount)
	configMounts := make(map[string]VolumeMount)
	
	// Data volume
	volumes["data"] = VolumeMount{
		Name:       "loki-data",
		Source:     vm.getVolumePath("loki", "data"),
		Target:     "/loki",
		Type:       vm.getVolumeType("loki"),
		AccessMode: AccessModeReadWrite,
		Options:    vm.getVolumeOptions("loki"),
	}
	
	// Configuration mount
	configMounts["config"] = VolumeMount{
		Name:       "loki-config",
		Source:     "loki-config",
		Target:     "/etc/loki",
		Type:       VolumeTypeConfigMap,
		AccessMode: AccessModeReadOnly,
	}
	
	return &VolumeConfiguration{
		ServiceName:  "loki",
		Volumes:      volumes,
		ConfigMounts: configMounts,
	}
}

func (vm *VolumeManager) createApplicationVolumes(serviceName string, appConfig config.ApplicationConfig) *VolumeConfiguration {
	volumes := make(map[string]VolumeMount)
	secretMounts := make(map[string]VolumeMount)
	configMounts := make(map[string]VolumeMount)
	tempVolumes := make(map[string]VolumeMount)
	
	// TODO: Application data volume
	// PersistentStorage field not yet implemented in ApplicationConfig
	// if appConfig.PersistentStorage.Enabled {
	//     volumes["data"] = VolumeMount{
	//         Name:       fmt.Sprintf("%s-data", serviceName),
	//         Source:     vm.getVolumePath(serviceName, "data"),
	//         Target:     appConfig.PersistentStorage.MountPath,
	//         Type:       vm.getVolumeType(serviceName),
	//         AccessMode: AccessModeReadWrite,
	//         Options:    vm.getVolumeOptions(serviceName),
	//     }
	// }
	
	// Configuration mount
	configMounts["config"] = VolumeMount{
		Name:       fmt.Sprintf("%s-config", serviceName),
		Source:     fmt.Sprintf("%s-config", serviceName),
		Target:     "/app/config",
		Type:       VolumeTypeConfigMap,
		AccessMode: AccessModeReadOnly,
	}
	
	// Secrets mount
	secretMounts["secrets"] = VolumeMount{
		Name:       fmt.Sprintf("%s-secrets", serviceName),
		Source:     fmt.Sprintf("%s-secrets", serviceName),
		Target:     "/app/secrets",
		Type:       VolumeTypeSecret,
		AccessMode: AccessModeReadOnly,
		Readonly:   true,
	}
	
	// Temporary volumes
	tempVolumes["tmp"] = VolumeMount{
		Name:       fmt.Sprintf("%s-tmp", serviceName),
		Source:     "",
		Target:     "/tmp",
		Type:       VolumeTypeTmpfs,
		AccessMode: AccessModeReadWrite,
		Options: map[string]string{
			"size": "100m",
			"mode": "1777",
		},
	}
	
	// Logs volume (shared for log aggregation)
	volumes["logs"] = VolumeMount{
		Name:       fmt.Sprintf("%s-logs", serviceName),
		Source:     vm.getVolumePath("shared", "logs"),
		Target:     "/app/logs",
		Type:       vm.getVolumeType("shared"),
		AccessMode: AccessModeReadWrite,
		Options:    vm.getVolumeOptions("shared"),
	}
	
	return &VolumeConfiguration{
		ServiceName:  serviceName,
		Volumes:      volumes,
		TempVolumes:  tempVolumes,
		ConfigMounts: configMounts,
		SecretMounts: secretMounts,
	}
}

// Helper methods for volume configuration

func (vm *VolumeManager) getVolumeType(serviceName string) VolumeType {
	if vm.config.Environment.IsDevelopment() {
		return VolumeTypeLocal
	}
	
	// For staging/production, use Azure Files for persistent storage
	if vm.needsPersistentStorage(serviceName) {
		return VolumeTypeAzureFiles
	}
	
	return VolumeTypeLocal
}

func (vm *VolumeManager) getVolumePath(serviceName, volumeName string) string {
	if vm.config.Environment.IsDevelopment() {
		return filepath.Join("/opt/data", serviceName, volumeName)
	}
	
	// For Azure Files, return the share name
	// TODO: Fix when ResourcePrefix field is added to DeploymentConfig
	// return fmt.Sprintf("%s-%s-%s", vm.config.ResourcePrefix, serviceName, volumeName)
	return fmt.Sprintf("ic-%s-%s", serviceName, volumeName) // Temporary hardcoded prefix
}

func (vm *VolumeManager) getVolumeOptions(serviceName string) map[string]string {
	options := make(map[string]string)
	
	if vm.config.Environment.IsDevelopment() {
		// Local volume options
		options["bind-propagation"] = "rprivate"
		options["consistency"] = "cached"
	} else {
		// Azure Files options
		options["driver"] = "azure-files"
		options["share"] = vm.getVolumePath(serviceName, "data")
	}
	
	return options
}

func (vm *VolumeManager) needsPersistentStorage(serviceName string) bool {
	persistentServices := []string{"postgresql", "redis", "vault", "grafana", "loki"}
	
	for _, service := range persistentServices {
		if serviceName == service {
			return true
		}
	}
	
	// TODO: Check application configurations
	// PersistentStorage field not yet implemented in ApplicationConfig
	// applications := map[string]config.ApplicationConfig{
	//     "content-api":    vm.config.Applications.ContentAPI,
	//     "services-api":   vm.config.Applications.ServicesAPI,
	//     "public-gateway": vm.config.Applications.PublicGateway,
	//     "admin-gateway":  vm.config.Applications.AdminGateway,
	// }
	// for appName, appConfig := range applications {
	//     if appName == serviceName && appConfig.PersistentStorage.Enabled {
	//         return true
	//     }
	// }
	
	return false
}

// Volume creation methods for different types

func (vm *VolumeManager) createLocalVolume(volumeName string, volumeMount VolumeMount) (pulumi.Resource, error) {
	// For local volumes, we typically don't create Pulumi resources
	// They are managed by the container runtime
	vm.ctx.Log.Info(fmt.Sprintf("Local volume configured: %s -> %s", volumeMount.Source, volumeMount.Target), nil)
	return nil, nil
}

func (vm *VolumeManager) createAzureFilesVolume(volumeName string, volumeMount VolumeMount) (pulumi.Resource, error) {
	// Create Azure File Share for persistent storage
	// TODO: Fix when Storage and ResourceGroupName fields are added to DeploymentConfig
	// fileShare, err := storage.NewFileShare(vm.ctx, volumeName, &storage.FileShareArgs{
	//     ShareName:          pulumi.String(volumeMount.Source),
	//     AccountName:        pulumi.String(vm.config.Storage.AccountName),
	//     ResourceGroupName:  pulumi.String(vm.config.ResourceGroupName),
	//     ShareQuota:         pulumi.Int(vm.getShareQuota(volumeName)),
	//     AccessTier:         pulumi.StringPtr("Hot"),
	//     EnabledProtocols:   pulumi.StringPtr("SMB"),
	// })
	
	// Temporary placeholder - return nil until fields are implemented
	vm.ctx.Log.Info(fmt.Sprintf("Azure File Share creation disabled - missing config fields: %s", volumeName), nil)
	return nil, nil
}

func (vm *VolumeManager) createTmpfsVolume(volumeName string, volumeMount VolumeMount) (pulumi.Resource, error) {
	// tmpfs volumes are created by the container runtime
	vm.ctx.Log.Info(fmt.Sprintf("Tmpfs volume configured: %s", volumeMount.Target), nil)
	return nil, nil
}

func (vm *VolumeManager) getShareQuota(volumeName string) int {
	// Default quotas in GB based on service type
	quotas := map[string]int{
		"postgresql": 50,
		"redis":      10,
		"vault":      5,
		"grafana":    10,
		"loki":       100,
		"shared":     20,
	}
	
	for service, quota := range quotas {
		if strings.Contains(volumeName, service) {
			return quota
		}
	}
	
	return 5 // Default 5GB
}

// GetVolumeMount formats volume mount for container runtime
func (vm *VolumeConfiguration) GetVolumeMount(volumeName string) (string, error) {
	if mount, exists := vm.Volumes[volumeName]; exists {
		return vm.formatVolumeMount(mount), nil
	}
	
	if mount, exists := vm.SharedVolumes[volumeName]; exists {
		return vm.formatVolumeMount(mount), nil
	}
	
	return "", fmt.Errorf("volume %s not found in configuration", volumeName)
}

func (vm *VolumeConfiguration) formatVolumeMount(mount VolumeMount) string {
	mountStr := fmt.Sprintf("%s:%s", mount.Source, mount.Target)
	
	if mount.Readonly {
		mountStr += ":ro"
	} else {
		mountStr += ":" + string(mount.AccessMode)
	}
	
	return mountStr
}

// GetAllVolumeMounts returns all volume mounts formatted for container runtime
func (vm *VolumeConfiguration) GetAllVolumeMounts() []string {
	var mounts []string
	
	// Regular volumes
	for _, mount := range vm.Volumes {
		mounts = append(mounts, vm.formatVolumeMount(mount))
	}
	
	// Shared volumes
	for _, mount := range vm.SharedVolumes {
		mounts = append(mounts, vm.formatVolumeMount(mount))
	}
	
	// Temp volumes
	for _, mount := range vm.TempVolumes {
		mounts = append(mounts, vm.formatVolumeMount(mount))
	}
	
	return mounts
}