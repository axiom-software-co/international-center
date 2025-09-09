package validation

import (
	"testing"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/components/infrastructure"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/components/platform"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/components/services"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/components/websites/public-website"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

func TestComponentProperties_InfrastructureEndpoints(t *testing.T) {
	config := GetDefaultPropertyTestConfig()
	
	test := ComponentPropertyTest{
		Name: "infrastructure_endpoints_always_defined",
		Property: func(environment string, config map[string]interface{}) bool {
			var endpointsValid bool
			
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				component, err := infrastructure.NewInfrastructureComponent(ctx, "test-infra", &infrastructure.InfrastructureArgs{
					Environment: environment,
				})
				if err != nil {
					return err
				}
				
				// Property: All infrastructure endpoints must be defined for any environment
				component.DatabaseEndpoint.ApplyT(func(endpoint string) string {
					endpointsValid = endpoint != ""
					return endpoint
				})
				
				component.StorageEndpoint.ApplyT(func(endpoint string) string {
					endpointsValid = endpointsValid && endpoint != ""
					return endpoint
				})
				
				component.VaultEndpoint.ApplyT(func(endpoint string) string {
					endpointsValid = endpointsValid && endpoint != ""
					return endpoint
				})
				
				component.MessagingEndpoint.ApplyT(func(endpoint string) string {
					endpointsValid = endpointsValid && endpoint != ""
					return endpoint
				})
				
				component.ObservabilityEndpoint.ApplyT(func(endpoint string) string {
					endpointsValid = endpointsValid && endpoint != ""
					return endpoint
				})
				
				return nil
			}, pulumi.WithMocks("project", "stack", &SharedMockResourceMonitor{}))
			
			return err == nil && endpointsValid
		},
	}
	
	RunPropertyBasedTest(t, test, config)
}

func TestComponentProperties_SecurityConfiguration(t *testing.T) {
	config := GetDefaultPropertyTestConfig()
	
	test := ComponentPropertyTest{
		Name: "security_authentication_enabled_in_production",
		Property: func(environment string, config map[string]interface{}) bool {
			if environment != "production" {
				return true // Property only applies to production
			}
			
			var authenticationEnabled bool
			
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				component, err := platform.NewPlatformComponent(ctx, "test-platform", &platform.PlatformArgs{
					Environment: environment,
					InfrastructureOutputs: pulumi.Map{},
				})
				if err != nil {
					return err
				}
				
				// Property: Authentication must be enabled in production environments
				component.SecurityConfig.ApplyT(func(config interface{}) interface{} {
					configMap := config.(map[string]interface{})
					if authValue, exists := configMap["authentication_enabled"]; exists {
						if authBool, ok := authValue.(bool); ok {
							authenticationEnabled = authBool
						}
					}
					return config
				})
				
				return nil
			}, pulumi.WithMocks("project", "stack", &SharedMockResourceMonitor{}))
			
			return err == nil && authenticationEnabled
		},
	}
	
	RunPropertyBasedTest(t, test, config)
}

func TestComponentProperties_ServiceDeploymentConsistency(t *testing.T) {
	config := GetDefaultPropertyTestConfig()
	
	test := ComponentPropertyTest{
		Name: "service_deployment_type_consistency",
		Property: func(environment string, config map[string]interface{}) bool {
			var deploymentTypeValid bool
			
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				component, err := services.NewServicesComponent(ctx, "test-services", &services.ServicesArgs{
					Environment: environment,
					InfrastructureOutputs: pulumi.Map{},
					PlatformOutputs: pulumi.Map{},
				})
				if err != nil {
					return err
				}
				
				// Property: Development uses podman_containers, staging/production use container_apps
				component.DeploymentType.ApplyT(func(deploymentType string) string {
					switch environment {
					case "development":
						deploymentTypeValid = deploymentType == "podman_containers"
					case "staging", "production":
						deploymentTypeValid = deploymentType == "container_apps"
					default:
						deploymentTypeValid = false
					}
					return deploymentType
				})
				
				return nil
			}, pulumi.WithMocks("project", "stack", &SharedMockResourceMonitor{}))
			
			return err == nil && deploymentTypeValid
		},
	}
	
	RunPropertyBasedTest(t, test, config)
}

func TestComponentProperties_WebsiteCDNandSSLAlignment(t *testing.T) {
	config := GetDefaultPropertyTestConfig()
	
	test := ComponentPropertyTest{
		Name: "website_cdn_ssl_alignment",
		Property: func(environment string, config map[string]interface{}) bool {
			var cdnSSLAligned bool
			
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				component, err := website.NewWebsiteComponent(ctx, "test-website", &website.WebsiteArgs{
					Environment: environment,
					InfrastructureOutputs: pulumi.Map{},
					PlatformOutputs: pulumi.Map{},
					ServicesOutputs: pulumi.Map{},
				})
				if err != nil {
					return err
				}
				
				// Property: CDN and SSL should have same enablement state
				var cdnEnabled, sslEnabled bool
				
				component.CDNEnabled.ApplyT(func(enabled bool) bool {
					cdnEnabled = enabled
					return enabled
				})
				
				component.SSLEnabled.ApplyT(func(enabled bool) bool {
					sslEnabled = enabled
					cdnSSLAligned = (cdnEnabled == sslEnabled)
					return enabled
				})
				
				return nil
			}, pulumi.WithMocks("project", "stack", &SharedMockResourceMonitor{}))
			
			return err == nil && cdnSSLAligned
		},
	}
	
	RunPropertyBasedTest(t, test, config)
}

func TestComponentProperties_HealthCheckEnabled(t *testing.T) {
	config := GetDefaultPropertyTestConfig()
	
	test := ComponentPropertyTest{
		Name: "health_checks_always_enabled",
		Property: func(environment string, config map[string]interface{}) bool {
			// Test all component types to ensure health checks are always enabled
			componentsHealthEnabled := true
			
			// Test Platform Component
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				platformComponent, err := platform.NewPlatformComponent(ctx, "test-platform", &platform.PlatformArgs{
					Environment: environment,
					InfrastructureOutputs: pulumi.Map{},
				})
				if err != nil {
					return err
				}
				
				platformComponent.HealthCheckEnabled.ApplyT(func(enabled bool) bool {
					componentsHealthEnabled = componentsHealthEnabled && enabled
					return enabled
				})
				
				return nil
			}, pulumi.WithMocks("project", "stack", &SharedMockResourceMonitor{}))
			
			if err != nil {
				return false
			}
			
			// Test Services Component
			err = pulumi.RunErr(func(ctx *pulumi.Context) error {
				servicesComponent, err := services.NewServicesComponent(ctx, "test-services", &services.ServicesArgs{
					Environment: environment,
					InfrastructureOutputs: pulumi.Map{},
					PlatformOutputs: pulumi.Map{},
				})
				if err != nil {
					return err
				}
				
				servicesComponent.HealthCheckEnabled.ApplyT(func(enabled bool) bool {
					componentsHealthEnabled = componentsHealthEnabled && enabled
					return enabled
				})
				
				return nil
			}, pulumi.WithMocks("project", "stack", &SharedMockResourceMonitor{}))
			
			if err != nil {
				return false
			}
			
			// Test Website Component
			err = pulumi.RunErr(func(ctx *pulumi.Context) error {
				websiteComponent, err := website.NewWebsiteComponent(ctx, "test-website", &website.WebsiteArgs{
					Environment: environment,
					InfrastructureOutputs: pulumi.Map{},
					PlatformOutputs: pulumi.Map{},
					ServicesOutputs: pulumi.Map{},
				})
				if err != nil {
					return err
				}
				
				websiteComponent.HealthCheckEnabled.ApplyT(func(enabled bool) bool {
					componentsHealthEnabled = componentsHealthEnabled && enabled
					return enabled
				})
				
				return nil
			}, pulumi.WithMocks("project", "stack", &SharedMockResourceMonitor{}))
			
			return err == nil && componentsHealthEnabled
		},
	}
	
	RunPropertyBasedTest(t, test, config)
}

func TestComponentProperties_EnvironmentSpecificURLPatterns(t *testing.T) {
	config := GetDefaultPropertyTestConfig()
	
	test := ComponentPropertyTest{
		Name: "environment_specific_url_patterns",
		Property: func(environment string, config map[string]interface{}) bool {
			var urlPatternValid bool
			
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				servicesComponent, err := services.NewServicesComponent(ctx, "test-services", &services.ServicesArgs{
					Environment: environment,
					InfrastructureOutputs: pulumi.Map{},
					PlatformOutputs: pulumi.Map{},
				})
				if err != nil {
					return err
				}
				
				// Property: URLs should contain environment-specific patterns
				servicesComponent.PublicGatewayURL.ApplyT(func(url string) string {
					switch environment {
					case "development":
						urlPatternValid = assert.Contains(t, url, "127.0.0.1")
					case "staging":
						urlPatternValid = assert.Contains(t, url, "staging") && assert.Contains(t, url, "azurecontainerapp.io")
					case "production":
						urlPatternValid = assert.Contains(t, url, "production") && assert.Contains(t, url, "azurecontainerapp.io")
					default:
						urlPatternValid = false
					}
					return url
				})
				
				return nil
			}, pulumi.WithMocks("project", "stack", &SharedMockResourceMonitor{}))
			
			return err == nil && urlPatternValid
		},
	}
	
	RunPropertyBasedTest(t, test, config)
}