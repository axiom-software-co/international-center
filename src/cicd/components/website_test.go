package components

import (
	"sync"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
)

// TestWebsiteComponent_DevelopmentEnvironment tests website component for development environment
func TestWebsiteComponent_DevelopmentEnvironment(t *testing.T) {
	var wg sync.WaitGroup
	
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployWebsite(ctx, cfg, "development")
		if err != nil {
			return err
		}

		// Verify development environment deploys Podman container
		wg.Add(1)
		pulumi.All(outputs.DeploymentType, outputs.ContainerID, outputs.ContainerStatus, outputs.ServerURL).ApplyT(func(args []interface{}) error {
			defer wg.Done()
			deploymentType := args[0].(string)
			containerID := args[1].(string)
			containerStatus := args[2].(string)
			serverURL := args[3].(string)

			assert.Equal(t, "podman_container", deploymentType, "Development should use Podman container")
			assert.NotEmpty(t, containerID, "Should have container ID")
			assert.Equal(t, "running", containerStatus, "Website container should be running")
			assert.Contains(t, serverURL, "http://localhost:3001", "Should use local development server")
			return nil
		})

		// Verify development container configuration and API integration
		wg.Add(1)
		pulumi.All(outputs.NodeVersion, outputs.APIGatewayURL, outputs.CDNEnabled).ApplyT(func(args []interface{}) error {
			defer wg.Done()
			nodeVersion := args[0].(string)
			apiGatewayURL := args[1].(string)
			cdnEnabled := args[2].(bool)

			assert.Contains(t, nodeVersion, "20", "Should use Node.js 20+")
			assert.Contains(t, apiGatewayURL, "localhost:9001", "Should use local public gateway")
			assert.False(t, cdnEnabled, "Should not enable CDN for development")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &WebsiteMocks{}))

	wg.Wait()
	assert.NoError(t, err)
}

// TestWebsiteComponent_StagingEnvironment tests website component for staging environment
func TestWebsiteComponent_StagingEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployWebsite(ctx, cfg, "staging")
		if err != nil {
			return err
		}

		// Verify staging environment generates Cloudflare Pages configuration
		pulumi.All(outputs.DeploymentType, outputs.ServerURL, outputs.BuildCommand).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			serverURL := args[1].(string)
			buildCommand := args[2].(string)

			assert.Equal(t, "cloudflare_pages", deploymentType, "Staging should use Cloudflare Pages")
			assert.Contains(t, serverURL, "staging.international-center.org", "Should use staging domain")
			assert.Equal(t, "npm run build", buildCommand, "Should use production build command")
			return nil
		})

		// Verify staging CDN and caching configuration
		pulumi.All(outputs.CDNEnabled, outputs.CachePolicy).ApplyT(func(args []interface{}) error {
			cdnEnabled := args[0].(bool)
			cachePolicy := args[1].(string)

			assert.True(t, cdnEnabled, "Should enable CDN for staging")
			assert.Equal(t, "moderate", cachePolicy, "Should use moderate cache policy")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &WebsiteMocks{}))

	assert.NoError(t, err)
}

// TestWebsiteComponent_ProductionEnvironment tests website component for production environment
func TestWebsiteComponent_ProductionEnvironment(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployWebsite(ctx, cfg, "production")
		if err != nil {
			return err
		}

		// Verify production environment generates Cloudflare Pages with production features
		pulumi.All(outputs.DeploymentType, outputs.ServerURL, outputs.CachePolicy).ApplyT(func(args []interface{}) error {
			deploymentType := args[0].(string)
			serverURL := args[1].(string)
			cachePolicy := args[2].(string)

			assert.Equal(t, "cloudflare_pages", deploymentType, "Production should use Cloudflare Pages")
			assert.Contains(t, serverURL, "international-center.org", "Should use production domain")
			assert.Equal(t, "aggressive", cachePolicy, "Should use aggressive cache policy")
			return nil
		})

		// Verify production has full CDN optimization
		pulumi.All(outputs.CDNEnabled, outputs.CompressionEnabled, outputs.SecurityHeaders).ApplyT(func(args []interface{}) error {
			cdnEnabled := args[0].(bool)
			compressionEnabled := args[1].(bool)
			securityHeaders := args[2].(bool)

			assert.True(t, cdnEnabled, "Should enable CDN for production")
			assert.True(t, compressionEnabled, "Should enable compression for production")
			assert.True(t, securityHeaders, "Should enable security headers for production")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &WebsiteMocks{}))

	assert.NoError(t, err)
}

// TestWebsiteComponent_APIIntegration tests API gateway integration across environments
func TestWebsiteComponent_APIIntegration(t *testing.T) {
	environments := []string{"development", "staging", "production"}

	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				cfg := config.New(ctx, "")

				outputs, err := DeployWebsite(ctx, cfg, env)
				if err != nil {
					return err
				}

				// Verify all environments provide API gateway integration
				pulumi.All(outputs.APIGatewayURL, outputs.APIIntegrationEnabled).ApplyT(func(args []interface{}) error {
					apiGatewayURL := args[0].(string)
					apiIntegrationEnabled := args[1].(bool)

					assert.NotEmpty(t, apiGatewayURL, "All environments should provide API gateway URL")
					assert.True(t, apiIntegrationEnabled, "All environments should enable API integration")

					// Verify environment-specific API URLs
					switch env {
					case "development":
						assert.Contains(t, apiGatewayURL, "localhost", "Development should use local API gateway")
					case "staging":
						assert.Contains(t, apiGatewayURL, "staging", "Staging should use staging API gateway")
					case "production":
						assert.Contains(t, apiGatewayURL, "international-center", "Production should use production API gateway")
						assert.NotContains(t, apiGatewayURL, "staging", "Production should not contain staging URL")
					}

					return nil
				})

				return nil
			}, pulumi.WithMocks("test", "stack", &WebsiteMocks{}))

			assert.NoError(t, err)
		})
	}
}

// TestWebsiteComponent_BuildConfiguration tests build pipeline configuration
func TestWebsiteComponent_BuildConfiguration(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		outputs, err := DeployWebsite(ctx, cfg, "production")
		if err != nil {
			return err
		}

		// Verify build configuration includes required parameters
		pulumi.All(outputs.BuildCommand, outputs.BuildDirectory, outputs.NodeVersion).ApplyT(func(args []interface{}) error {
			buildCommand := args[0].(string)
			buildDirectory := args[1].(string)
			nodeVersion := args[2].(string)

			assert.NotEmpty(t, buildCommand, "Should provide build command")
			assert.NotEmpty(t, buildDirectory, "Should provide build directory")
			assert.NotEmpty(t, nodeVersion, "Should provide Node.js version")

			// Verify Astro-specific build configuration
			assert.Equal(t, "npm run build", buildCommand, "Should use npm build command")
			assert.Equal(t, "dist", buildDirectory, "Should use Astro output directory")
			assert.Contains(t, nodeVersion, "20", "Should use Node.js 20+")
			return nil
		})

		return nil
	}, pulumi.WithMocks("test", "stack", &WebsiteMocks{}))

	assert.NoError(t, err)
}

// TestWebsiteComponent_EnvironmentParity tests that all environments support required features
func TestWebsiteComponent_EnvironmentParity(t *testing.T) {
	environments := []string{"development", "staging", "production"}

	for _, env := range environments {
		t.Run("Environment_"+env, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				cfg := config.New(ctx, "")

				outputs, err := DeployWebsite(ctx, cfg, env)
				if err != nil {
					return err
				}

				// Verify all environments provide required outputs
				pulumi.All(outputs.DeploymentType, outputs.ServerURL, outputs.BuildCommand).ApplyT(func(args []interface{}) error {
					deploymentType := args[0].(string)
					serverURL := args[1].(string)
					buildCommand := args[2].(string)

					assert.NotEmpty(t, deploymentType, "All environments should provide deployment type")
					assert.NotEmpty(t, serverURL, "All environments should provide server URL")
					assert.NotEmpty(t, buildCommand, "All environments should provide build command")
					return nil
				})

				return nil
			}, pulumi.WithMocks("test", "stack", &WebsiteMocks{}))

			assert.NoError(t, err)
		})
	}
}

// WebsiteMocks provides mocks for Pulumi testing
type WebsiteMocks struct{}

func (mocks *WebsiteMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}

	switch args.TypeToken {
	case "cloudflare:index/pagesProject:PagesProject":
		// Mock Cloudflare Pages project
		outputs["name"] = resource.NewStringProperty("international-center-website")
		outputs["productionBranch"] = resource.NewStringProperty("main")
		outputs["deploymentConfigs"] = resource.NewObjectProperty(resource.PropertyMap{
			"production": resource.NewObjectProperty(resource.PropertyMap{
				"buildCommand":        resource.NewStringProperty("npm run build"),
				"destinationDir":      resource.NewStringProperty("dist"),
				"rootDir":             resource.NewStringProperty("."),
				"webAnalyticsTag":     resource.NewStringProperty("website-analytics"),
				"webAnalyticsEnabled": resource.NewBoolProperty(true),
			}),
			"preview": resource.NewObjectProperty(resource.PropertyMap{
				"buildCommand":   resource.NewStringProperty("npm run build"),
				"destinationDir": resource.NewStringProperty("dist"),
				"rootDir":        resource.NewStringProperty("."),
			}),
		})
		outputs["subdomain"] = resource.NewStringProperty("international-center")

	case "cloudflare:index/pagesDomain:PagesDomain":
		// Mock Cloudflare Pages custom domain
		outputs["projectName"] = resource.NewStringProperty("international-center-website")
		outputs["domain"] = resource.NewStringProperty("international-center.org")

	case "cloudflare:index/record:Record":
		// Mock DNS records
		if args.Name == "website-cname" {
			outputs["name"] = resource.NewStringProperty("www")
			outputs["type"] = resource.NewStringProperty("CNAME")
			outputs["value"] = resource.NewStringProperty("international-center.pages.dev")
		}

	case "docker:index/container:Container":
		// Mock docker container for development
		if args.Name == "website-dev" {
			outputs["name"] = resource.NewStringProperty("website-dev-container")
			outputs["image"] = resource.NewStringProperty("node:20-alpine")
			outputs["id"] = resource.NewStringProperty("website-dev-container-id")
			outputs["ports"] = resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewObjectProperty(resource.PropertyMap{
					"internal": resource.NewNumberProperty(3001),
					"external": resource.NewNumberProperty(3001),
				}),
			})
			outputs["env"] = resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewStringProperty("NODE_ENV=development"),
			})
		}
	}

	return args.Name + "_id", outputs, nil
}

func (mocks *WebsiteMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := resource.PropertyMap{}
	return outputs, nil
}