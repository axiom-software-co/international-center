package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type MessagingArgs struct {
	Environment string
}

type MessagingComponent struct {
	pulumi.ResourceState

	Endpoint       pulumi.StringOutput `pulumi:"endpoint"`
	Username       pulumi.StringOutput `pulumi:"username"`
	Password       pulumi.StringOutput `pulumi:"password"`
	HealthEndpoint pulumi.StringOutput `pulumi:"healthEndpoint"`
}

func NewMessagingComponent(ctx *pulumi.Context, name string, args *MessagingArgs, opts ...pulumi.ResourceOption) (*MessagingComponent, error) {
	component := &MessagingComponent{}
	
	err := ctx.RegisterComponentResource("international-center:infrastructure:Messaging", name, component, opts...)
	if err != nil {
		return nil, err
	}

	var endpoint, username, password, healthEndpoint pulumi.StringOutput

	switch args.Environment {
	case "development":
		endpoint = pulumi.String("amqp://guest:guest@localhost:5672/").ToStringOutput()
		username = pulumi.String("guest").ToStringOutput()
		password = pulumi.String("guest").ToStringOutput()
		healthEndpoint = pulumi.String("http://localhost:15672/api/healthchecks/node").ToStringOutput()
	case "staging":
		endpoint = pulumi.String("amqps://rabbitmq-staging:staging_password@rabbitmq-staging.servicebus.windows.net:5671/").ToStringOutput()
		username = pulumi.String("rabbitmq-staging").ToStringOutput()
		password = pulumi.String("staging_password").ToStringOutput()
		healthEndpoint = pulumi.String("https://rabbitmq-staging.servicebus.windows.net/api/healthchecks/node").ToStringOutput()
	case "production":
		endpoint = pulumi.String("amqps://rabbitmq-production:production_password@rabbitmq-production.servicebus.windows.net:5671/").ToStringOutput()
		username = pulumi.String("rabbitmq-production").ToStringOutput()
		password = pulumi.String("production_password").ToStringOutput()
		healthEndpoint = pulumi.String("https://rabbitmq-production.servicebus.windows.net/api/healthchecks/node").ToStringOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.Endpoint = endpoint
	component.Username = username
	component.Password = password
	component.HealthEndpoint = healthEndpoint

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"endpoint":       component.Endpoint,
		"username":       component.Username,
		"password":       component.Password,
		"healthEndpoint": component.HealthEndpoint,
	}); err != nil {
		return nil, err
	}

	return component, nil
}