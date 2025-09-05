package components

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// RedisOutputs represents the outputs from redis component
type RedisOutputs struct {
	DeploymentType    pulumi.StringOutput
	Endpoint          pulumi.StringOutput
	Port              pulumi.IntOutput
	Password          pulumi.StringOutput
	MaxMemory         pulumi.StringOutput
	Persistence       pulumi.BoolOutput
}

// DeployRedis deploys Redis infrastructure based on environment
func DeployRedis(ctx *pulumi.Context, cfg *config.Config, environment string) (*RedisOutputs, error) {
	switch environment {
	case "development":
		return deployDevelopmentRedis(ctx, cfg)
	case "staging":
		return deployStagingRedis(ctx, cfg)
	case "production":
		return deployProductionRedis(ctx, cfg)
	default:
		return nil, fmt.Errorf("unknown environment: %s", environment)
	}
}

// deployDevelopmentRedis deploys Redis container for development caching
func deployDevelopmentRedis(ctx *pulumi.Context, cfg *config.Config) (*RedisOutputs, error) {
	// Create Redis container using Podman for caching
	redisContainer, err := local.NewCommand(ctx, "redis-container", &local.CommandArgs{
		Create: pulumi.String("podman run -d --name redis-dev -p 6379:6379 redis:7-alpine"),
		Delete: pulumi.String("podman stop redis-dev && podman rm redis-dev"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis container: %w", err)
	}

	deploymentType := pulumi.String("podman_container").ToStringOutput()
	endpoint := pulumi.String("localhost:6379").ToStringOutput()
	port := pulumi.Int(6379).ToIntOutput()
	password := pulumi.String("").ToStringOutput() // No password for development
	maxMemory := pulumi.String("256mb").ToStringOutput()
	persistence := pulumi.Bool(false).ToBoolOutput() // No persistence for development

	// Add dependency on container creation
	endpoint = pulumi.All(redisContainer.Stdout).ApplyT(func(args []interface{}) string {
		return "localhost:6379"
	}).(pulumi.StringOutput)

	return &RedisOutputs{
		DeploymentType: deploymentType,
		Endpoint:       endpoint,
		Port:           port,
		Password:       password,
		MaxMemory:      maxMemory,
		Persistence:    persistence,
	}, nil
}

// deployStagingRedis deploys managed Redis for staging caching
func deployStagingRedis(ctx *pulumi.Context, cfg *config.Config) (*RedisOutputs, error) {
	// For staging, use Upstash Redis managed service
	deploymentType := pulumi.String("upstash_managed").ToStringOutput()
	endpoint := pulumi.String("redis-staging.upstash.io").ToStringOutput()
	port := pulumi.Int(6379).ToIntOutput()
	password := pulumi.String("staging-redis-password").ToStringOutput()
	maxMemory := pulumi.String("1gb").ToStringOutput()
	persistence := pulumi.Bool(true).ToBoolOutput()

	return &RedisOutputs{
		DeploymentType: deploymentType,
		Endpoint:       endpoint,
		Port:           port,
		Password:       password,
		MaxMemory:      maxMemory,
		Persistence:    persistence,
	}, nil
}

// deployProductionRedis deploys managed Redis for production caching
func deployProductionRedis(ctx *pulumi.Context, cfg *config.Config) (*RedisOutputs, error) {
	// For production, use Upstash Redis managed service with high availability
	deploymentType := pulumi.String("upstash_managed").ToStringOutput()
	endpoint := pulumi.String("redis-production.upstash.io").ToStringOutput()
	port := pulumi.Int(6379).ToIntOutput()
	password := pulumi.String("production-redis-password").ToStringOutput()
	maxMemory := pulumi.String("4gb").ToStringOutput()
	persistence := pulumi.Bool(true).ToBoolOutput()

	return &RedisOutputs{
		DeploymentType: deploymentType,
		Endpoint:       endpoint,
		Port:           port,
		Password:       password,
		MaxMemory:      maxMemory,
		Persistence:    persistence,
	}, nil
}