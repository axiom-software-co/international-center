package infrastructure

import (
	"context"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type DaprStack interface {
	Deploy(ctx context.Context) (DaprDeployment, error)
	GetRedisConnectionInfo() (string, int, string)
	GetDaprEndpoints() map[string]string
	GenerateDaprComponents(ctx context.Context, deployment DaprDeployment) error
	ValidateDeployment(ctx context.Context, deployment DaprDeployment) error
}

type DaprDeployment interface {
	GetSidecarEndpoint() pulumi.StringOutput
	GetPlacementEndpoint() pulumi.StringOutput
	GetStateStoreEndpoint() pulumi.StringOutput
	GetPubSubEndpoint() pulumi.StringOutput
	GetConfigurationEndpoint() pulumi.StringOutput
}

type DaprConfiguration struct {
	Environment           string
	DeploymentMode        string // "sidecar", "kubernetes", "standalone"
	StateStoreProvider    string // "redis", "cosmosdb", "postgresql"
	PubSubProvider        string // "redis", "servicebus", "eventhubs"
	SecretsProvider       string // "kubernetes", "vault", "local-file"
	ConfigurationProvider string // "redis", "kubernetes"
	TracingEnabled        bool
	MetricsEnabled        bool
	LogLevel              string
	AppPort               int
	HTTPPort              int
	GRPCPort              int
	PlacementPort         int
	EnableMTLS            bool
	TrustDomain           string
	ComponentsPath        string
	ConfigPath            string
}

type StateStoreConfiguration struct {
	Name           string
	Type           string
	Version        string
	ConsistencyMode string // "strong", "eventual"
	ReplicationFactor int
	Metadata       map[string]string
	Scopes         []string
}

type PubSubConfiguration struct {
	Name             string
	Type             string
	Version          string
	DeliveryMode     string // "at-least-once", "exactly-once"
	MaxRedeliveries  int
	Metadata         map[string]string
	Scopes          []string
	BulkSubscribe    bool
}

type SecretsConfiguration struct {
	Name     string
	Type     string
	Version  string
	Metadata map[string]string
	Scopes   []string
}

type ConfigurationStoreConfiguration struct {
	Name       string
	Type       string
	Version    string
	Metadata   map[string]string
	Scopes     []string
	PollInterval string
}

type DaprComponent struct {
	ApiVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   DaprComponentMetadata   `yaml:"metadata"`
	Spec       DaprComponentSpec      `yaml:"spec"`
}

type DaprComponentMetadata struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace,omitempty"`
	Labels    map[string]string `yaml:"labels,omitempty"`
}

type DaprComponentSpec struct {
	Type     string                    `yaml:"type"`
	Version  string                    `yaml:"version"`
	Metadata []DaprComponentSpecMeta   `yaml:"metadata,omitempty"`
	Scopes   []string                  `yaml:"scopes,omitempty"`
}

type DaprComponentSpecMeta struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type DaprFactory interface {
	CreateDaprStack(ctx *pulumi.Context, config *config.Config, environment string) DaprStack
}

func GetDaprConfiguration(environment string, config *config.Config) *DaprConfiguration {
	switch environment {
	case "development":
		return &DaprConfiguration{
			Environment:           "development",
			DeploymentMode:        "standalone",
			StateStoreProvider:    "redis",
			PubSubProvider:        "redis",
			SecretsProvider:       "local-file",
			ConfigurationProvider: "redis",
			TracingEnabled:        true,
			MetricsEnabled:        true,
			LogLevel:              "debug",
			AppPort:               8080,
			HTTPPort:              3500,
			GRPCPort:              50001,
			PlacementPort:         50005,
			EnableMTLS:            false,
			TrustDomain:           "public",
			ComponentsPath:        "./dapr/components",
			ConfigPath:            "./dapr/config.yaml",
		}
	case "staging":
		return &DaprConfiguration{
			Environment:           "staging",
			DeploymentMode:        "kubernetes",
			StateStoreProvider:    "cosmosdb",
			PubSubProvider:        "servicebus",
			SecretsProvider:       "vault",
			ConfigurationProvider: "redis",
			TracingEnabled:        true,
			MetricsEnabled:        true,
			LogLevel:              "info",
			AppPort:               8080,
			HTTPPort:              3500,
			GRPCPort:              50001,
			PlacementPort:         50005,
			EnableMTLS:            true,
			TrustDomain:           "staging.international-center.com",
			ComponentsPath:        "/dapr/components",
			ConfigPath:            "/dapr/config.yaml",
		}
	case "production":
		return &DaprConfiguration{
			Environment:           "production",
			DeploymentMode:        "kubernetes",
			StateStoreProvider:    "cosmosdb",
			PubSubProvider:        "servicebus",
			SecretsProvider:       "vault",
			ConfigurationProvider: "redis",
			TracingEnabled:        true,
			MetricsEnabled:        true,
			LogLevel:              "warn",
			AppPort:               8080,
			HTTPPort:              3500,
			GRPCPort:              50001,
			PlacementPort:         50005,
			EnableMTLS:            true,
			TrustDomain:           "production.international-center.com",
			ComponentsPath:        "/dapr/components",
			ConfigPath:            "/dapr/config.yaml",
		}
	default:
		return &DaprConfiguration{
			Environment:           environment,
			DeploymentMode:        "standalone",
			StateStoreProvider:    "redis",
			PubSubProvider:        "redis",
			SecretsProvider:       "local-file",
			ConfigurationProvider: "redis",
			TracingEnabled:        true,
			MetricsEnabled:        true,
			LogLevel:              "info",
			AppPort:               8080,
			HTTPPort:              3500,
			GRPCPort:              50001,
			PlacementPort:         50005,
			EnableMTLS:            false,
			TrustDomain:           "public",
			ComponentsPath:        "./dapr/components",
			ConfigPath:            "./dapr/config.yaml",
		}
	}
}

func GetStateStoreConfiguration(environment string) StateStoreConfiguration {
	switch environment {
	case "development":
		return StateStoreConfiguration{
			Name:            "statestore",
			Type:            "state.redis",
			Version:         "v1",
			ConsistencyMode: "eventual",
			ReplicationFactor: 1,
			Metadata: map[string]string{
				"redisHost":     "localhost:6379",
				"redisPassword": "",
			},
			Scopes: []string{"content-api", "services-api"},
		}
	case "staging":
		return StateStoreConfiguration{
			Name:            "statestore",
			Type:            "state.azure.cosmosdb",
			Version:         "v1",
			ConsistencyMode: "strong",
			ReplicationFactor: 2,
			Metadata: map[string]string{
				"url":          "https://int-staging-cosmos.documents.azure.com:443/",
				"database":     "StateDB",
				"collection":   "StateStore",
			},
			Scopes: []string{"content-api", "services-api"},
		}
	case "production":
		return StateStoreConfiguration{
			Name:            "statestore",
			Type:            "state.azure.cosmosdb",
			Version:         "v1",
			ConsistencyMode: "strong",
			ReplicationFactor: 3,
			Metadata: map[string]string{
				"url":          "https://int-prod-cosmos.documents.azure.com:443/",
				"database":     "StateDB",
				"collection":   "StateStore",
			},
			Scopes: []string{"content-api", "services-api"},
		}
	default:
		return StateStoreConfiguration{
			Name:            "statestore",
			Type:            "state.redis",
			Version:         "v1",
			ConsistencyMode: "eventual",
			ReplicationFactor: 1,
			Metadata: map[string]string{
				"redisHost":     "localhost:6379",
				"redisPassword": "",
			},
			Scopes: []string{"content-api", "services-api"},
		}
	}
}

func GetPubSubConfiguration(environment string) PubSubConfiguration {
	switch environment {
	case "development":
		return PubSubConfiguration{
			Name:            "pubsub",
			Type:            "pubsub.redis",
			Version:         "v1",
			DeliveryMode:    "at-least-once",
			MaxRedeliveries: 3,
			Metadata: map[string]string{
				"redisHost":     "localhost:6379",
				"redisPassword": "",
			},
			Scopes:        []string{"content-api", "services-api"},
			BulkSubscribe: false,
		}
	case "staging":
		return PubSubConfiguration{
			Name:            "pubsub",
			Type:            "pubsub.azure.servicebus",
			Version:         "v1",
			DeliveryMode:    "at-least-once",
			MaxRedeliveries: 5,
			Metadata: map[string]string{
				"connectionString": "",
				"timeoutInSec":     "60",
			},
			Scopes:        []string{"content-api", "services-api"},
			BulkSubscribe: true,
		}
	case "production":
		return PubSubConfiguration{
			Name:            "pubsub",
			Type:            "pubsub.azure.servicebus",
			Version:         "v1",
			DeliveryMode:    "exactly-once",
			MaxRedeliveries: 10,
			Metadata: map[string]string{
				"connectionString": "",
				"timeoutInSec":     "60",
			},
			Scopes:        []string{"content-api", "services-api"},
			BulkSubscribe: true,
		}
	default:
		return PubSubConfiguration{
			Name:            "pubsub",
			Type:            "pubsub.redis",
			Version:         "v1",
			DeliveryMode:    "at-least-once",
			MaxRedeliveries: 3,
			Metadata: map[string]string{
				"redisHost":     "localhost:6379",
				"redisPassword": "",
			},
			Scopes:        []string{"content-api", "services-api"},
			BulkSubscribe: false,
		}
	}
}

// DaprMetrics defines performance and reliability metrics for environment-specific policies
type DaprMetrics struct {
	MaxStateOperationsPerSec int
	MaxPubSubThroughput      int
	MaxSidecarMemoryMB       int
	EnableCircuitBreaker     bool
	RetryPolicy             string
	TimeoutSeconds          int
}

func GetDaprMetrics(environment string) DaprMetrics {
	switch environment {
	case "development":
		return DaprMetrics{
			MaxStateOperationsPerSec: 100,
			MaxPubSubThroughput:      50,
			MaxSidecarMemoryMB:       512,
			EnableCircuitBreaker:     false,
			RetryPolicy:              "linear",
			TimeoutSeconds:           30,
		}
	case "staging":
		return DaprMetrics{
			MaxStateOperationsPerSec: 1000,
			MaxPubSubThroughput:      500,
			MaxSidecarMemoryMB:       1024,
			EnableCircuitBreaker:     true,
			RetryPolicy:              "exponential",
			TimeoutSeconds:           15,
		}
	case "production":
		return DaprMetrics{
			MaxStateOperationsPerSec: 10000,
			MaxPubSubThroughput:      5000,
			MaxSidecarMemoryMB:       2048,
			EnableCircuitBreaker:     true,
			RetryPolicy:              "exponential",
			TimeoutSeconds:           10,
		}
	default:
		return DaprMetrics{
			MaxStateOperationsPerSec: 500,
			MaxPubSubThroughput:      100,
			MaxSidecarMemoryMB:       768,
			EnableCircuitBreaker:     false,
			RetryPolicy:              "linear",
			TimeoutSeconds:           20,
		}
	}
}