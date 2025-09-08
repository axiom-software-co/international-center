package components

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
)

// TestRabbitMQContainerDeployment_Development validates that rabbitmq-dev container is deployed and running
func TestRabbitMQContainerDeployment_Development(t *testing.T) {
	t.Run("RabbitMQContainerExists_Development", func(t *testing.T) {
		validateRabbitMQContainerExists(t, "rabbitmq-dev")
	})

	t.Run("RabbitMQContainerRunning_Development", func(t *testing.T) {
		validateRabbitMQContainerRunning(t, "rabbitmq-dev", "rabbitmq:3-management-alpine", []string{"5672", "15672"})
	})

	t.Run("RabbitMQContainerHealthy_Development", func(t *testing.T) {
		validateRabbitMQContainerHealthy(t, "rabbitmq-dev")
	})
}

// validateRabbitMQContainerExists checks if rabbitmq container exists
func validateRabbitMQContainerExists(t *testing.T, name string) {
	cmd := exec.Command("podman", "ps", "-a", "--filter", "name="+name, "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check for rabbitmq container %s: %v", name, err)
	}

	containerNames := strings.TrimSpace(string(output))
	assert.Contains(t, containerNames, name, "RabbitMQ container %s should exist", name)
}

// validateRabbitMQContainerRunning checks if rabbitmq container is running with correct image and ports
func validateRabbitMQContainerRunning(t *testing.T, name, expectedImage string, expectedPorts []string) {
	cmd := exec.Command("podman", "ps", "--filter", "name="+name, "--format", "{{.Names}}\t{{.Image}}\t{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check rabbitmq container %s status: %v", name, err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("RabbitMQ container %s is not running", name)
	}

	parts := strings.Split(lines[0], "\t")
	if len(parts) < 3 {
		t.Fatalf("Unexpected rabbitmq container %s output format", name)
	}

	containerName := parts[0]
	containerImage := parts[1]
	containerPorts := parts[2]

	assert.Equal(t, name, containerName, "Container name should match")
	assert.Equal(t, expectedImage, containerImage, "Container should use correct rabbitmq image")
	
	for _, port := range expectedPorts {
		assert.Contains(t, containerPorts, port, "Container should expose port %s", port)
	}
}

// validateRabbitMQContainerHealthy checks if rabbitmq container is healthy
func validateRabbitMQContainerHealthy(t *testing.T, name string) {
	cmd := exec.Command("podman", "ps", "--filter", "name="+name, "--format", "{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check rabbitmq container %s health: %v", name, err)
	}

	status := strings.TrimSpace(string(output))
	assert.NotEmpty(t, status, "RabbitMQ container %s should have status", name)
	assert.Contains(t, strings.ToLower(status), "up", "RabbitMQ container %s should be running (Up status)", name)
}

// TestRabbitMQComponent_DevelopmentEnvironment tests rabbitmq component for development environment
func TestRabbitMQComponent_DevelopmentEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployRabbitMQ(ctx, cfg, "development")
		if err != nil {
			return err
		}

		// Verify development environment generates local RabbitMQ container configuration
		pulumi.All(outputs.DeploymentType, outputs.Endpoint, outputs.Port, outputs.ManagementPort).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			endpoint := args[1].(string)
			port := args[2].(int)
			managementPort := args[3].(int)

			assert.Equal(t, "podman_container", deploymentType, "Development should use local RabbitMQ container")
			assert.Contains(t, endpoint, "localhost:5672", "Should use local RabbitMQ endpoint")
			assert.Equal(t, 5672, port, "Should use RabbitMQ standard port")
			assert.Equal(t, 15672, managementPort, "Should use RabbitMQ management port")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &RabbitMQMocks{}))

	assert.NoError(t, err)
}

// TestRabbitMQComponent_StagingEnvironment tests rabbitmq component for staging environment
func TestRabbitMQComponent_StagingEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployRabbitMQ(ctx, cfg, "staging")
		if err != nil {
			return err
		}

		// Verify staging environment generates CloudAMQP configuration
		pulumi.All(outputs.DeploymentType, outputs.Endpoint, outputs.Port, outputs.HighAvailability).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			endpoint := args[1].(string)
			port := args[2].(int)
			highAvailability := args[3].(bool)

			assert.Equal(t, "cloudamqp_managed", deploymentType, "Staging should use CloudAMQP managed service")
			assert.Contains(t, endpoint, "cloudamqp.com", "Should use CloudAMQP endpoint")
			assert.Equal(t, 5671, port, "Should use TLS port for staging")
			assert.True(t, highAvailability, "Should enable high availability for staging")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &RabbitMQMocks{}))

	assert.NoError(t, err)
}

// TestRabbitMQComponent_ProductionEnvironment tests rabbitmq component for production environment
func TestRabbitMQComponent_ProductionEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployRabbitMQ(ctx, cfg, "production")
		if err != nil {
			return err
		}

		// Verify production environment generates CloudAMQP with production features
		pulumi.All(outputs.DeploymentType, outputs.Endpoint, outputs.Port, outputs.HighAvailability).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			endpoint := args[1].(string)
			port := args[2].(int)
			highAvailability := args[3].(bool)

			assert.Equal(t, "cloudamqp_managed", deploymentType, "Production should use CloudAMQP managed service")
			assert.Contains(t, endpoint, "cloudamqp.com", "Should use CloudAMQP endpoint")
			assert.Equal(t, 5671, port, "Should use TLS port for production")
			assert.True(t, highAvailability, "Should enable high availability for production")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &RabbitMQMocks{}))

	assert.NoError(t, err)
}

// TestRabbitMQComponent_EnvironmentParity tests that all environments support required features
func TestRabbitMQComponent_EnvironmentParity(t *testing.T) {
	environments := []string{"development", "staging", "production"}

	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				cfg := config.New(ctx, "")

				outputs, err := DeployRabbitMQ(ctx, cfg, env)
				if err != nil {
					return err
				}

				// Verify all environments provide required outputs
				pulumi.All(outputs.Endpoint, outputs.Username, outputs.Password, outputs.VHost).ApplyT(func(args []interface{}) error {
					endpoint := args[0].(string)
					username := args[1].(string)
					password := args[2].(string)
					vhost := args[3].(string)

					assert.NotEmpty(t, endpoint, "All environments should provide endpoint")
					assert.NotEmpty(t, username, "All environments should provide username")
					assert.NotEmpty(t, password, "All environments should provide password")
					assert.NotEmpty(t, vhost, "All environments should provide vhost")
					return nil
				})

				return nil
			}, pulumi.WithMocks("test", "stack", &RabbitMQMocks{}))

			assert.NoError(t, err)
		})
	}
}

// RabbitMQMocks provides mocks for Pulumi testing
type RabbitMQMocks struct{}

func (mocks *RabbitMQMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}

	switch args.TypeToken {
	case "docker:index/container:Container":
		outputs["name"] = resource.NewStringProperty("rabbitmq-dev")
		outputs["image"] = resource.NewStringProperty("rabbitmq:3-management-alpine")
		outputs["ports"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{
				"internal": resource.NewNumberProperty(5672),
				"external": resource.NewNumberProperty(5672),
			}),
			resource.NewObjectProperty(resource.PropertyMap{
				"internal": resource.NewNumberProperty(15672),
				"external": resource.NewNumberProperty(15672),
			}),
		})

	case "cloudamqp:index/instance:Instance":
		outputs["name"] = resource.NewStringProperty("international-center-rabbitmq")
		outputs["url"] = resource.NewStringProperty("amqps://international-center.cloudamqp.com")
		outputs["apikey"] = resource.NewStringProperty("mock-api-key")
	}

	return args.Name + "_id", outputs, nil
}

func (mocks *RabbitMQMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}
	return outputs, nil
}