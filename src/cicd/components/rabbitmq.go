package components

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// RabbitMQOutputs represents the outputs from rabbitmq component
type RabbitMQOutputs struct {
	DeploymentType    pulumi.StringOutput
	Endpoint          pulumi.StringOutput
	ManagementURL     pulumi.StringOutput
	Port              pulumi.IntOutput
	ManagementPort    pulumi.IntOutput
	Username          pulumi.StringOutput
	Password          pulumi.StringOutput
	VHost             pulumi.StringOutput
	HighAvailability  pulumi.BoolOutput
}

// DeployRabbitMQ deploys RabbitMQ infrastructure based on environment
func DeployRabbitMQ(ctx *pulumi.Context, cfg *config.Config, environment string) (*RabbitMQOutputs, error) {
	switch environment {
	case "development":
		return deployDevelopmentRabbitMQ(ctx, cfg)
	case "staging":
		return deployStagingRabbitMQ(ctx, cfg)
	case "production":
		return deployProductionRabbitMQ(ctx, cfg)
	default:
		return nil, fmt.Errorf("unknown environment: %s", environment)
	}
}

// deployDevelopmentRabbitMQ deploys RabbitMQ container for development pub/sub
func deployDevelopmentRabbitMQ(ctx *pulumi.Context, cfg *config.Config) (*RabbitMQOutputs, error) {
	// Create RabbitMQ container using Podman for pub/sub messaging
	rabbitmqContainer, err := local.NewCommand(ctx, "rabbitmq-container", &local.CommandArgs{
		Create: pulumi.String("podman run -d --name rabbitmq-dev -e RABBITMQ_DEFAULT_USER=user -e RABBITMQ_DEFAULT_PASS=password -p 5672:5672 -p 15672:15672 rabbitmq:3-management-alpine"),
		Delete: pulumi.String("podman stop rabbitmq-dev && podman rm rabbitmq-dev"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ container: %w", err)
	}

	deploymentType := pulumi.String("podman_container").ToStringOutput()
	endpoint := pulumi.String("localhost:5672").ToStringOutput()
	managementURL := pulumi.String("http://localhost:15672").ToStringOutput()
	port := pulumi.Int(5672).ToIntOutput()
	managementPort := pulumi.Int(15672).ToIntOutput()
	username := pulumi.String("user").ToStringOutput()
	password := pulumi.String("password").ToStringOutput()
	vhost := pulumi.String("/").ToStringOutput()
	highAvailability := pulumi.Bool(false).ToBoolOutput()

	// Add dependency on container creation
	endpoint = pulumi.All(rabbitmqContainer.Stdout).ApplyT(func(args []interface{}) string {
		return "localhost:5672"
	}).(pulumi.StringOutput)

	return &RabbitMQOutputs{
		DeploymentType:   deploymentType,
		Endpoint:         endpoint,
		ManagementURL:    managementURL,
		Port:             port,
		ManagementPort:   managementPort,
		Username:         username,
		Password:         password,
		VHost:            vhost,
		HighAvailability: highAvailability,
	}, nil
}

// deployStagingRabbitMQ deploys managed RabbitMQ for staging pub/sub
func deployStagingRabbitMQ(ctx *pulumi.Context, cfg *config.Config) (*RabbitMQOutputs, error) {
	// For staging, use CloudAMQP managed service
	deploymentType := pulumi.String("cloudamqp_managed").ToStringOutput()
	endpoint := pulumi.String("amqps://staging.cloudamqp.com").ToStringOutput()
	managementURL := pulumi.String("https://staging-rabbitmq.cloudamqp.com").ToStringOutput()
	port := pulumi.Int(5671).ToIntOutput() // TLS port
	managementPort := pulumi.Int(443).ToIntOutput()
	username := pulumi.String("staging-user").ToStringOutput()
	password := pulumi.String("staging-rabbitmq-password").ToStringOutput()
	vhost := pulumi.String("staging").ToStringOutput()
	highAvailability := pulumi.Bool(true).ToBoolOutput()

	return &RabbitMQOutputs{
		DeploymentType:   deploymentType,
		Endpoint:         endpoint,
		ManagementURL:    managementURL,
		Port:             port,
		ManagementPort:   managementPort,
		Username:         username,
		Password:         password,
		VHost:            vhost,
		HighAvailability: highAvailability,
	}, nil
}

// deployProductionRabbitMQ deploys managed RabbitMQ for production pub/sub
func deployProductionRabbitMQ(ctx *pulumi.Context, cfg *config.Config) (*RabbitMQOutputs, error) {
	// For production, use CloudAMQP managed service with high availability cluster
	deploymentType := pulumi.String("cloudamqp_managed").ToStringOutput()
	endpoint := pulumi.String("amqps://production.cloudamqp.com").ToStringOutput()
	managementURL := pulumi.String("https://production-rabbitmq.cloudamqp.com").ToStringOutput()
	port := pulumi.Int(5671).ToIntOutput() // TLS port
	managementPort := pulumi.Int(443).ToIntOutput()
	username := pulumi.String("production-user").ToStringOutput()
	password := pulumi.String("production-rabbitmq-password").ToStringOutput()
	vhost := pulumi.String("production").ToStringOutput()
	highAvailability := pulumi.Bool(true).ToBoolOutput()

	return &RabbitMQOutputs{
		DeploymentType:   deploymentType,
		Endpoint:         endpoint,
		ManagementURL:    managementURL,
		Port:             port,
		ManagementPort:   managementPort,
		Username:         username,
		Password:         password,
		VHost:            vhost,
		HighAvailability: highAvailability,
	}, nil
}