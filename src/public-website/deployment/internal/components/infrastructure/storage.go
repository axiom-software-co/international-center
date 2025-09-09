package infrastructure

import (
	"fmt"
	"net/url"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type StorageProvider string

const (
	StorageProviderAzureBlob StorageProvider = "azure_blob"
	StorageProviderAWSS3     StorageProvider = "aws_s3"
	StorageProviderGCPCloud  StorageProvider = "gcp_cloud"
	StorageProviderMinIO     StorageProvider = "minio"
	StorageProviderLocal     StorageProvider = "local"
)

type StorageConfig struct {
	Provider          StorageProvider
	Host              string
	Port              int
	BucketName        string
	Region            string
	AccessKeyID       string
	SecretAccessKey   string
	UseTLS            bool
	HealthCheckPath   string
	HealthCheckPort   int
	MaxConnections    int
	ConnectionTimeout int
	RetryAttempts     int
	RetryDelay        int
	PublicRead        bool
	Versioning        bool
	EncryptionEnabled bool
	BackupEnabled     bool
	BackupRetention   int
	AdditionalParams  map[string]string
}

type StorageArgs struct {
	Config      *StorageConfig
	Environment string
	ProjectName string
}

type StorageComponent struct {
	pulumi.ResourceState

	ConnectionString  pulumi.StringOutput `pulumi:"connectionString"`
	Endpoint          pulumi.StringOutput `pulumi:"endpoint"`
	BucketName        pulumi.StringOutput `pulumi:"bucketName"`
	Region            pulumi.StringOutput `pulumi:"region"`
	HealthEndpoint    pulumi.StringOutput `pulumi:"healthEndpoint"`
	Provider          pulumi.StringOutput `pulumi:"provider"`
	PublicRead        pulumi.BoolOutput   `pulumi:"publicRead"`
	Versioning        pulumi.BoolOutput   `pulumi:"versioning"`
	EncryptionEnabled pulumi.BoolOutput   `pulumi:"encryptionEnabled"`
}

func NewStorageComponent(ctx *pulumi.Context, name string, args *StorageArgs, opts ...pulumi.ResourceOption) (*StorageComponent, error) {
	component := &StorageComponent{}
	
	if ctx != nil {
		err := ctx.RegisterComponentResource("framework:storage:Storage", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	config := args.Config
	if config == nil {
		return nil, fmt.Errorf("storage config is required")
	}

	connectionString := buildStorageConnectionString(config)
	endpoint := buildStorageEndpoint(config)
	healthEndpoint := buildStorageHealthEndpoint(config)

	component.ConnectionString = pulumi.String(connectionString).ToStringOutput()
	component.Endpoint = pulumi.String(endpoint).ToStringOutput()
	component.BucketName = pulumi.String(config.BucketName).ToStringOutput()
	component.Region = pulumi.String(config.Region).ToStringOutput()
	component.HealthEndpoint = pulumi.String(healthEndpoint).ToStringOutput()
	component.Provider = pulumi.String(string(config.Provider)).ToStringOutput()
	component.PublicRead = pulumi.Bool(config.PublicRead).ToBoolOutput()
	component.Versioning = pulumi.Bool(config.Versioning).ToBoolOutput()
	component.EncryptionEnabled = pulumi.Bool(config.EncryptionEnabled).ToBoolOutput()

	if ctx != nil {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"connectionString":  component.ConnectionString,
			"endpoint":          component.Endpoint,
			"bucketName":        component.BucketName,
			"region":            component.Region,
			"healthEndpoint":    component.HealthEndpoint,
			"provider":          component.Provider,
			"publicRead":        component.PublicRead,
			"versioning":        component.Versioning,
			"encryptionEnabled": component.EncryptionEnabled,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}

func buildStorageConnectionString(config *StorageConfig) string {
	switch config.Provider {
	case StorageProviderAzureBlob:
		return fmt.Sprintf("DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;EndpointSuffix=core.windows.net",
			config.AccessKeyID, config.SecretAccessKey)
			
	case StorageProviderAWSS3:
		return fmt.Sprintf("s3://%s:%s@%s/%s",
			config.AccessKeyID, config.SecretAccessKey, config.Region, config.BucketName)
			
	case StorageProviderGCPCloud:
		return fmt.Sprintf("gs://%s", config.BucketName)
		
	case StorageProviderMinIO:
		protocol := "http"
		if config.UseTLS {
			protocol = "https"
		}
		return fmt.Sprintf("%s://%s:%s@%s:%d/%s",
			protocol, config.AccessKeyID, config.SecretAccessKey, config.Host, config.Port, config.BucketName)
			
	case StorageProviderLocal:
		return fmt.Sprintf("file://%s", config.BucketName)
		
	default:
		return ""
	}
}

func buildStorageEndpoint(config *StorageConfig) string {
	protocol := "http"
	if config.UseTLS {
		protocol = "https"
	}
	
	switch config.Provider {
	case StorageProviderAzureBlob:
		return fmt.Sprintf("https://%s.blob.core.windows.net", config.AccessKeyID)
		
	case StorageProviderAWSS3:
		if config.Region != "" {
			return fmt.Sprintf("https://s3.%s.amazonaws.com", config.Region)
		}
		return "https://s3.amazonaws.com"
		
	case StorageProviderGCPCloud:
		return "https://storage.googleapis.com"
		
	case StorageProviderMinIO:
		return fmt.Sprintf("%s://%s:%d", protocol, config.Host, config.Port)
		
	case StorageProviderLocal:
		return fmt.Sprintf("file://%s", config.BucketName)
		
	default:
		return fmt.Sprintf("%s://%s:%d", protocol, config.Host, config.Port)
	}
}

func buildStorageHealthEndpoint(config *StorageConfig) string {
	if config.HealthCheckPath == "" {
		return ""
	}
	
	port := config.HealthCheckPort
	if port == 0 {
		port = config.Port
	}
	
	protocol := "http"
	if config.UseTLS {
		protocol = "https"
	}
	
	return fmt.Sprintf("%s://%s:%d%s", protocol, config.Host, port, config.HealthCheckPath)
}

func DefaultAzureBlobConfig(accountName, accountKey, containerName string) *StorageConfig {
	return &StorageConfig{
		Provider:          StorageProviderAzureBlob,
		Host:              fmt.Sprintf("%s.blob.core.windows.net", accountName),
		Port:              443,
		BucketName:        containerName,
		Region:            "",
		AccessKeyID:       accountName,
		SecretAccessKey:   accountKey,
		UseTLS:            true,
		HealthCheckPath:   "",
		HealthCheckPort:   0,
		MaxConnections:    100,
		ConnectionTimeout: 30,
		RetryAttempts:     3,
		RetryDelay:        1000,
		PublicRead:        false,
		Versioning:        true,
		EncryptionEnabled: true,
		BackupEnabled:     true,
		BackupRetention:   30,
		AdditionalParams:  make(map[string]string),
	}
}

func DefaultAWSS3Config(accessKeyID, secretAccessKey, bucketName, region string) *StorageConfig {
	return &StorageConfig{
		Provider:          StorageProviderAWSS3,
		Host:              fmt.Sprintf("s3.%s.amazonaws.com", region),
		Port:              443,
		BucketName:        bucketName,
		Region:            region,
		AccessKeyID:       accessKeyID,
		SecretAccessKey:   secretAccessKey,
		UseTLS:            true,
		HealthCheckPath:   "",
		HealthCheckPort:   0,
		MaxConnections:    100,
		ConnectionTimeout: 30,
		RetryAttempts:     3,
		RetryDelay:        1000,
		PublicRead:        false,
		Versioning:        true,
		EncryptionEnabled: true,
		BackupEnabled:     true,
		BackupRetention:   30,
		AdditionalParams:  make(map[string]string),
	}
}

func DefaultMinIOConfig(host string, accessKeyID, secretAccessKey, bucketName string) *StorageConfig {
	return &StorageConfig{
		Provider:          StorageProviderMinIO,
		Host:              host,
		Port:              9000,
		BucketName:        bucketName,
		Region:            "",
		AccessKeyID:       accessKeyID,
		SecretAccessKey:   secretAccessKey,
		UseTLS:            false,
		HealthCheckPath:   "/minio/health/live",
		HealthCheckPort:   9000,
		MaxConnections:    100,
		ConnectionTimeout: 30,
		RetryAttempts:     3,
		RetryDelay:        1000,
		PublicRead:        false,
		Versioning:        true,
		EncryptionEnabled: false,
		BackupEnabled:     true,
		BackupRetention:   30,
		AdditionalParams:  make(map[string]string),
	}
}

func DefaultLocalStorageConfig(basePath string) *StorageConfig {
	return &StorageConfig{
		Provider:          StorageProviderLocal,
		Host:              "localhost",
		Port:              0,
		BucketName:        basePath,
		Region:            "",
		AccessKeyID:       "",
		SecretAccessKey:   "",
		UseTLS:            false,
		HealthCheckPath:   "",
		HealthCheckPort:   0,
		MaxConnections:    1,
		ConnectionTimeout: 5,
		RetryAttempts:     1,
		RetryDelay:        0,
		PublicRead:        false,
		Versioning:        false,
		EncryptionEnabled: false,
		BackupEnabled:     false,
		BackupRetention:   0,
		AdditionalParams:  make(map[string]string),
	}
}

func ParseStorageURL(storageURL string) (*StorageConfig, error) {
	u, err := url.Parse(storageURL)
	if err != nil {
		return nil, err
	}
	
	config := &StorageConfig{
		Host:             u.Hostname(),
		AdditionalParams: make(map[string]string),
	}
	
	if u.Port() != "" {
		port := 80
		if u.Port() != "" {
			fmt.Sscanf(u.Port(), "%d", &port)
		}
		config.Port = port
	}
	
	config.UseTLS = u.Scheme == "https" || u.Scheme == "s3" || u.Scheme == "gs"
	
	switch u.Scheme {
	case "s3":
		config.Provider = StorageProviderAWSS3
	case "gs":
		config.Provider = StorageProviderGCPCloud
	case "file":
		config.Provider = StorageProviderLocal
	case "http", "https":
		if u.Hostname() == "s3.amazonaws.com" || u.Host == "amazonaws.com" {
			config.Provider = StorageProviderAWSS3
		} else {
			config.Provider = StorageProviderMinIO
		}
	}
	
	if u.User != nil {
		config.AccessKeyID = u.User.Username()
		if password, ok := u.User.Password(); ok {
			config.SecretAccessKey = password
		}
	}
	
	config.BucketName = u.Path
	if config.BucketName != "" && config.BucketName[0] == '/' {
		config.BucketName = config.BucketName[1:]
	}
	
	return config, nil
}