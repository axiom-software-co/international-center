package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type MessagingProvider string

const (
	MessagingProviderRabbitMQ MessagingProvider = "rabbitmq"
	MessagingProviderKafka    MessagingProvider = "kafka"
	MessagingProviderRedis    MessagingProvider = "redis"
	MessagingProviderNATS     MessagingProvider = "nats"
)

type MessagingConfig struct {
	Provider           MessagingProvider
	Host               string
	Port               int
	Username           string
	Password           string
	VirtualHost        string
	UseTLS             bool
	HealthCheckPath    string
	HealthCheckPort    int
	MaxConnections     int
	ConnectionTimeout  int
	RetryAttempts      int
	RetryDelay         int
	AdditionalParams   map[string]string
}

type MessagingArgs struct {
	Config      *MessagingConfig
	Environment string
	ProjectName string
}

type MessagingComponent struct {
	pulumi.ResourceState

	Endpoint       pulumi.StringOutput `pulumi:"endpoint"`
	Host           pulumi.StringOutput `pulumi:"host"`
	Port           pulumi.IntOutput    `pulumi:"port"`
	Username       pulumi.StringOutput `pulumi:"username"`
	Password       pulumi.StringOutput `pulumi:"password"`
	VirtualHost    pulumi.StringOutput `pulumi:"virtualHost"`
	HealthEndpoint pulumi.StringOutput `pulumi:"healthEndpoint"`
	Provider       pulumi.StringOutput `pulumi:"provider"`
	UseTLS         pulumi.BoolOutput   `pulumi:"useTLS"`
}

func NewMessagingComponent(ctx *pulumi.Context, name string, args *MessagingArgs, opts ...pulumi.ResourceOption) (*MessagingComponent, error) {
	component := &MessagingComponent{}
	
	if ctx != nil {
		err := ctx.RegisterComponentResource("framework:messaging:Messaging", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	config := args.Config
	if config == nil {
		return nil, fmt.Errorf("messaging config is required")
	}

	endpoint := buildMessagingEndpoint(config)
	healthEndpoint := buildMessagingHealthEndpoint(config)

	component.Endpoint = pulumi.String(endpoint).ToStringOutput()
	component.Host = pulumi.String(config.Host).ToStringOutput()
	component.Port = pulumi.Int(config.Port).ToIntOutput()
	component.Username = pulumi.String(config.Username).ToStringOutput()
	component.Password = pulumi.String(config.Password).ToStringOutput()
	component.VirtualHost = pulumi.String(config.VirtualHost).ToStringOutput()
	component.HealthEndpoint = pulumi.String(healthEndpoint).ToStringOutput()
	component.Provider = pulumi.String(string(config.Provider)).ToStringOutput()
	component.UseTLS = pulumi.Bool(config.UseTLS).ToBoolOutput()

	if ctx != nil {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"endpoint":       component.Endpoint,
			"host":           component.Host,
			"port":           component.Port,
			"username":       component.Username,
			"password":       component.Password,
			"virtualHost":    component.VirtualHost,
			"healthEndpoint": component.HealthEndpoint,
			"provider":       component.Provider,
			"useTLS":         component.UseTLS,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}

func buildMessagingEndpoint(config *MessagingConfig) string {
	protocol := "amqp"
	if config.UseTLS {
		protocol = "amqps"
	}
	
	switch config.Provider {
	case MessagingProviderRabbitMQ:
		if config.VirtualHost != "" && config.VirtualHost != "/" {
			return fmt.Sprintf("%s://%s:%s@%s:%d/%s",
				protocol, config.Username, config.Password, config.Host, config.Port, config.VirtualHost)
		}
		return fmt.Sprintf("%s://%s:%s@%s:%d/",
			protocol, config.Username, config.Password, config.Host, config.Port)
			
	case MessagingProviderKafka:
		return fmt.Sprintf("kafka://%s:%d", config.Host, config.Port)
		
	case MessagingProviderRedis:
		redisProtocol := "redis"
		if config.UseTLS {
			redisProtocol = "rediss"
		}
		if config.Password != "" {
			return fmt.Sprintf("%s://:%s@%s:%d", redisProtocol, config.Password, config.Host, config.Port)
		}
		return fmt.Sprintf("%s://%s:%d", redisProtocol, config.Host, config.Port)
		
	case MessagingProviderNATS:
		natsProtocol := "nats"
		if config.UseTLS {
			natsProtocol = "tls"
		}
		return fmt.Sprintf("%s://%s:%s@%s:%d", natsProtocol, config.Username, config.Password, config.Host, config.Port)
		
	default:
		return fmt.Sprintf("%s://%s:%s@%s:%d/", protocol, config.Username, config.Password, config.Host, config.Port)
	}
}

func buildMessagingHealthEndpoint(config *MessagingConfig) string {
	if config.HealthCheckPath == "" {
		return ""
	}
	
	port := config.HealthCheckPort
	if port == 0 {
		switch config.Provider {
		case MessagingProviderRabbitMQ:
			port = 15672
		case MessagingProviderKafka:
			port = 9092
		case MessagingProviderRedis:
			port = config.Port
		case MessagingProviderNATS:
			port = 8222
		default:
			port = config.Port
		}
	}
	
	protocol := "http"
	if config.UseTLS {
		protocol = "https"
	}
	
	return fmt.Sprintf("%s://%s:%d%s", protocol, config.Host, port, config.HealthCheckPath)
}

func DefaultRabbitMQConfig(host string) *MessagingConfig {
	return &MessagingConfig{
		Provider:          MessagingProviderRabbitMQ,
		Host:              host,
		Port:              5672,
		Username:          "guest",
		Password:          "guest",
		VirtualHost:       "/",
		UseTLS:            false,
		HealthCheckPath:   "/api/healthchecks/node",
		HealthCheckPort:   15672,
		MaxConnections:    100,
		ConnectionTimeout: 30,
		RetryAttempts:     3,
		RetryDelay:        1000,
		AdditionalParams:  make(map[string]string),
	}
}

func DefaultKafkaConfig(host string) *MessagingConfig {
	return &MessagingConfig{
		Provider:          MessagingProviderKafka,
		Host:              host,
		Port:              9092,
		Username:          "",
		Password:          "",
		VirtualHost:       "",
		UseTLS:            false,
		HealthCheckPath:   "/health",
		HealthCheckPort:   0,
		MaxConnections:    100,
		ConnectionTimeout: 30,
		RetryAttempts:     3,
		RetryDelay:        1000,
		AdditionalParams:  make(map[string]string),
	}
}

func DefaultRedisConfig(host string) *MessagingConfig {
	return &MessagingConfig{
		Provider:          MessagingProviderRedis,
		Host:              host,
		Port:              6379,
		Username:          "",
		Password:          "",
		VirtualHost:       "",
		UseTLS:            false,
		HealthCheckPath:   "/health",
		HealthCheckPort:   0,
		MaxConnections:    100,
		ConnectionTimeout: 30,
		RetryAttempts:     3,
		RetryDelay:        1000,
		AdditionalParams:  make(map[string]string),
	}
}

func DefaultNATSConfig(host string) *MessagingConfig {
	return &MessagingConfig{
		Provider:          MessagingProviderNATS,
		Host:              host,
		Port:              4222,
		Username:          "",
		Password:          "",
		VirtualHost:       "",
		UseTLS:            false,
		HealthCheckPath:   "/healthz",
		HealthCheckPort:   8222,
		MaxConnections:    100,
		ConnectionTimeout: 30,
		RetryAttempts:     3,
		RetryDelay:        1000,
		AdditionalParams:  make(map[string]string),
	}
}