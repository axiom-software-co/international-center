package infrastructure

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/axiom-software-co/international-center/src/deployer/shared/infrastructure"
	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type ServiceStack struct {
	ctx             *pulumi.Context
	config          *config.Config
	containerRuntime *infrastructure.ContainerRuntime
	daprDeployment  *DaprDeployment
	networkName     string
	environment     string
	projectRoot     string
}

type ServiceDeployment struct {
	// Container Images
	ContentAPIImage     *docker.Image
	ServicesAPIImage    *docker.Image
	PublicGatewayImage  *docker.Image
	AdminGatewayImage   *docker.Image
	
	// Service Containers (with integrated Dapr sidecars)
	ContentAPIContainer     *docker.Container
	ServicesAPIContainer    *docker.Container
	PublicGatewayContainer  *docker.Container
	AdminGatewayContainer   *docker.Container
	
	// Service Networks
	ServiceNetwork *docker.Network
	
	// Dapr Sidecar Containers
	ContentAPIDaprSidecar   *docker.Container
	ServicesAPIDaprSidecar  *docker.Container
	PublicGatewayDaprSidecar *docker.Container
	AdminGatewayDaprSidecar *docker.Container
}

func NewServiceStack(ctx *pulumi.Context, config *config.Config, daprDeployment *DaprDeployment, networkName, environment, projectRoot string) *ServiceStack {
	containerRuntime := infrastructure.NewContainerRuntime(ctx, "podman", "localhost", 5000)
	
	return &ServiceStack{
		ctx:              ctx,
		config:           config,
		containerRuntime: containerRuntime,
		daprDeployment:   daprDeployment,
		networkName:      networkName,
		environment:      environment,
		projectRoot:      projectRoot,
	}
}

func (ss *ServiceStack) Deploy(ctx context.Context) (*ServiceDeployment, error) {
	deployment := &ServiceDeployment{}
	
	var err error
	
	// Create service network that connects to Dapr network
	deployment.ServiceNetwork, err = ss.createServiceNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to create service network: %w", err)
	}
	
	// Build container images
	deployment.ContentAPIImage, err = ss.buildContentAPIImage()
	if err != nil {
		return nil, fmt.Errorf("failed to build content-api image: %w", err)
	}
	
	deployment.ServicesAPIImage, err = ss.buildServicesAPIImage()
	if err != nil {
		return nil, fmt.Errorf("failed to build services-api image: %w", err)
	}
	
	deployment.PublicGatewayImage, err = ss.buildPublicGatewayImage()
	if err != nil {
		return nil, fmt.Errorf("failed to build public-gateway image: %w", err)
	}
	
	deployment.AdminGatewayImage, err = ss.buildAdminGatewayImage()
	if err != nil {
		return nil, fmt.Errorf("failed to build admin-gateway image: %w", err)
	}
	
	// Deploy Dapr sidecars
	deployment.ContentAPIDaprSidecar, err = ss.deployDaprSidecar("content-api", 3501, 50002, deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy content-api Dapr sidecar: %w", err)
	}
	
	deployment.ServicesAPIDaprSidecar, err = ss.deployDaprSidecar("services-api", 3502, 50003, deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy services-api Dapr sidecar: %w", err)
	}
	
	deployment.PublicGatewayDaprSidecar, err = ss.deployDaprSidecar("public-gateway", 3503, 50004, deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy public-gateway Dapr sidecar: %w", err)
	}
	
	deployment.AdminGatewayDaprSidecar, err = ss.deployDaprSidecar("admin-gateway", 3504, 50006, deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy admin-gateway Dapr sidecar: %w", err)
	}
	
	// Deploy service containers
	deployment.ContentAPIContainer, err = ss.deployContentAPIContainer(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy content-api container: %w", err)
	}
	
	deployment.ServicesAPIContainer, err = ss.deployServicesAPIContainer(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy services-api container: %w", err)
	}
	
	deployment.PublicGatewayContainer, err = ss.deployPublicGatewayContainer(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy public-gateway container: %w", err)
	}
	
	deployment.AdminGatewayContainer, err = ss.deployAdminGatewayContainer(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy admin-gateway container: %w", err)
	}
	
	return deployment, nil
}

func (ss *ServiceStack) createServiceNetwork() (*docker.Network, error) {
	network, err := docker.NewNetwork(ss.ctx, "service-network", &docker.NetworkArgs{
		Name:   pulumi.Sprintf("%s-service-network", ss.environment),
		Driver: pulumi.String("bridge"),
		Ipam: &docker.NetworkIpamArgs{
			Driver: pulumi.String("default"),
			Configs: docker.NetworkIpamConfigArray{
				&docker.NetworkIpamConfigArgs{
					Subnet:  pulumi.String("172.20.0.0/16"),
					Gateway: pulumi.String("172.20.0.1"),
				},
			},
		},
		Options: pulumi.StringMap{
			"com.docker.network.bridge.name": pulumi.String("br-service"),
			"com.docker.network.driver.mtu":  pulumi.String("1500"),
		},
		Labels: docker.NetworkLabelArray{
			&docker.NetworkLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ss.environment),
			},
			&docker.NetworkLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("services"),
			},
			&docker.NetworkLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	
	return network, nil
}

func (ss *ServiceStack) buildContentAPIImage() (*docker.Image, error) {
	return ss.containerRuntime.BuildImage(
		"content-api", 
		ss.projectRoot,
		filepath.Join(ss.projectRoot, "src/deployer/shared/infrastructure/containers/Dockerfile.content-api"),
		map[string]string{
			"APP_VERSION": ss.config.Get("app_version"),
			"BUILD_TIME":  fmt.Sprintf("%d", time.Now().Unix()),
			"GIT_COMMIT":  "development",
		},
	)
}

func (ss *ServiceStack) buildServicesAPIImage() (*docker.Image, error) {
	return ss.containerRuntime.BuildImage(
		"services-api", 
		ss.projectRoot,
		filepath.Join(ss.projectRoot, "src/deployer/shared/infrastructure/containers/Dockerfile.services-api"),
		map[string]string{
			"APP_VERSION": ss.config.Get("app_version"),
			"BUILD_TIME":  fmt.Sprintf("%d", time.Now().Unix()),
			"GIT_COMMIT":  "development",
		},
	)
}

func (ss *ServiceStack) buildPublicGatewayImage() (*docker.Image, error) {
	return ss.containerRuntime.BuildImage(
		"public-gateway", 
		ss.projectRoot,
		filepath.Join(ss.projectRoot, "src/deployer/shared/infrastructure/containers/Dockerfile.public-gateway"),
		map[string]string{
			"APP_VERSION": ss.config.Get("app_version"),
			"BUILD_TIME":  fmt.Sprintf("%d", time.Now().Unix()),
			"GIT_COMMIT":  "development",
		},
	)
}

func (ss *ServiceStack) buildAdminGatewayImage() (*docker.Image, error) {
	return ss.containerRuntime.BuildImage(
		"admin-gateway", 
		ss.projectRoot,
		filepath.Join(ss.projectRoot, "src/deployer/shared/infrastructure/containers/Dockerfile.admin-gateway"),
		map[string]string{
			"APP_VERSION": ss.config.Get("app_version"),
			"BUILD_TIME":  fmt.Sprintf("%d", time.Now().Unix()),
			"GIT_COMMIT":  "development",
		},
	)
}

func (ss *ServiceStack) deployDaprSidecar(appID string, daprHTTPPort, daprGRPCPort int, deployment *ServiceDeployment) (*docker.Container, error) {
	daprVersion := ss.config.Get("dapr_version")
	if daprVersion == "" {
		daprVersion = "1.12.0"
	}
	
	container, err := docker.NewContainer(ss.ctx, fmt.Sprintf("%s-dapr-sidecar", appID), &docker.ContainerArgs{
		Name:    pulumi.Sprintf("%s-%s-dapr-sidecar", ss.environment, appID),
		Image:   pulumi.Sprintf("daprio/daprd:%s", daprVersion),
		Restart: pulumi.String("unless-stopped"),
		
		Command: pulumi.StringArray{
			pulumi.String("./daprd"),
			pulumi.Sprintf("--app-id=%s", appID),
			pulumi.Sprintf("--app-port=%d", ss.getAppPort(appID)),
			pulumi.Sprintf("--dapr-http-port=%d", daprHTTPPort),
			pulumi.Sprintf("--dapr-grpc-port=%d", daprGRPCPort),
			pulumi.String("--placement-host-address=dapr-placement:50005"),
			pulumi.String("--components-path=/components"),
			pulumi.String("--log-level=info"),
			pulumi.String("--app-ssl=false"),
		},
		
		Ports: docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(daprHTTPPort),
				External: pulumi.Int(daprHTTPPort),
				Protocol: pulumi.String("tcp"),
			},
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(daprGRPCPort),
				External: pulumi.Int(daprGRPCPort),
				Protocol: pulumi.String("tcp"),
			},
		},
		
		Mounts: docker.ContainerMountArray{
			&docker.ContainerMountArgs{
				Type:   pulumi.String("volume"),
				Source: ss.daprDeployment.DaprComponentsVolume.Name,
				Target: pulumi.String("/components"),
			},
		},
		
		NetworksAdvanced: docker.ContainerNetworksAdvancedArray{
			&docker.ContainerNetworksAdvancedArgs{
				Name: deployment.ServiceNetwork.Name,
			},
			&docker.ContainerNetworksAdvancedArgs{
				Name: ss.daprDeployment.RedisNetwork.Name,
			},
		},
		
		DependsOn: pulumi.Array{
			ss.daprDeployment.DaprPlacementContainer,
			ss.daprDeployment.RedisContainer,
		},
		
		Labels: docker.ContainerLabelArray{
			&docker.ContainerLabelArgs{
				Label: pulumi.String("environment"),
				Value: pulumi.String(ss.environment),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("component"),
				Value: pulumi.String("dapr-sidecar"),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("app-id"),
				Value: pulumi.String(appID),
			},
			&docker.ContainerLabelArgs{
				Label: pulumi.String("managed-by"),
				Value: pulumi.String("pulumi"),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	
	return container, nil
}

func (ss *ServiceStack) deployContentAPIContainer(deployment *ServiceDeployment) (*docker.Container, error) {
	return ss.deployServiceContainer(infrastructure.ContainerConfig{
		Name:  fmt.Sprintf("%s-content-api", ss.environment),
		Image: "content-api",
		Tag:   "latest",
		Environment: map[string]pulumi.StringInput{
			"DATABASE_URL":       pulumi.String(ss.getDatabaseURL()),
			"REDIS_ADDR":         pulumi.String("redis:6379"),
			"REDIS_PASSWORD":     pulumi.String(ss.config.Get("redis_password")),
			"CONTENT_API_ADDR":   pulumi.String(":8080"),
			"ENVIRONMENT":        pulumi.String(ss.environment),
			"APP_VERSION":        pulumi.String(ss.config.Get("app_version")),
			"DAPR_HTTP_PORT":     pulumi.String("3501"),
			"DAPR_GRPC_PORT":     pulumi.String("50002"),
		},
		Ports: []infrastructure.ContainerPort{
			{Internal: 8080, External: 8080, Protocol: "tcp"},
		},
		Networks: []string{deployment.ServiceNetwork.Name.ToStringOutput().AsStringOutput().ToStringOutput().ApplyT(func(name interface{}) string { return name.(string) }).(string)},
		RestartPolicy: "unless-stopped",
		HealthCheck: infrastructure.ContainerHealthCheck{
			Test:     []string{"CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1"},
			Interval: 30 * time.Second,
			Timeout:  10 * time.Second,
			Retries:  3,
		},
		Labels: map[string]string{
			"environment": ss.environment,
			"service":     "content-api",
			"component":   "backend-api",
			"managed-by":  "pulumi",
		},
		DependsOn: []pulumi.Resource{
			deployment.ContentAPIDaprSidecar,
			ss.daprDeployment.RedisContainer,
		},
	})
}

func (ss *ServiceStack) deployServicesAPIContainer(deployment *ServiceDeployment) (*docker.Container, error) {
	return ss.deployServiceContainer(infrastructure.ContainerConfig{
		Name:  fmt.Sprintf("%s-services-api", ss.environment),
		Image: "services-api",
		Tag:   "latest",
		Environment: map[string]pulumi.StringInput{
			"DATABASE_URL":       pulumi.String(ss.getDatabaseURL()),
			"REDIS_ADDR":         pulumi.String("redis:6379"),
			"REDIS_PASSWORD":     pulumi.String(ss.config.Get("redis_password")),
			"SERVICES_API_ADDR":  pulumi.String(":8081"),
			"ENVIRONMENT":        pulumi.String(ss.environment),
			"APP_VERSION":        pulumi.String(ss.config.Get("app_version")),
			"DAPR_HTTP_PORT":     pulumi.String("3502"),
			"DAPR_GRPC_PORT":     pulumi.String("50003"),
		},
		Ports: []infrastructure.ContainerPort{
			{Internal: 8081, External: 8081, Protocol: "tcp"},
		},
		Networks: []string{deployment.ServiceNetwork.Name.ToStringOutput().AsStringOutput().ToStringOutput().ApplyT(func(name interface{}) string { return name.(string) }).(string)},
		RestartPolicy: "unless-stopped",
		HealthCheck: infrastructure.ContainerHealthCheck{
			Test:     []string{"CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1"},
			Interval: 30 * time.Second,
			Timeout:  10 * time.Second,
			Retries:  3,
		},
		Labels: map[string]string{
			"environment": ss.environment,
			"service":     "services-api",
			"component":   "backend-api",
			"managed-by":  "pulumi",
		},
		DependsOn: []pulumi.Resource{
			deployment.ServicesAPIDaprSidecar,
			ss.daprDeployment.RedisContainer,
		},
	})
}

func (ss *ServiceStack) deployPublicGatewayContainer(deployment *ServiceDeployment) (*docker.Container, error) {
	return ss.deployServiceContainer(infrastructure.ContainerConfig{
		Name:  fmt.Sprintf("%s-public-gateway", ss.environment),
		Image: "public-gateway",
		Tag:   "latest",
		Environment: map[string]pulumi.StringInput{
			"PUBLIC_GATEWAY_PORT":     pulumi.String("8082"),
			"PUBLIC_ALLOWED_ORIGINS":  pulumi.String("*"),
			"ENVIRONMENT":             pulumi.String(ss.environment),
			"APP_VERSION":             pulumi.String(ss.config.Get("app_version")),
			"DAPR_HTTP_PORT":          pulumi.String("3503"),
			"DAPR_GRPC_PORT":          pulumi.String("50004"),
			"CONTENT_API_URL":         pulumi.String("http://localhost:3501"),
			"SERVICES_API_URL":        pulumi.String("http://localhost:3502"),
		},
		Ports: []infrastructure.ContainerPort{
			{Internal: 8082, External: 8082, Protocol: "tcp"},
		},
		Networks: []string{deployment.ServiceNetwork.Name.ToStringOutput().AsStringOutput().ToStringOutput().ApplyT(func(name interface{}) string { return name.(string) }).(string)},
		RestartPolicy: "unless-stopped",
		HealthCheck: infrastructure.ContainerHealthCheck{
			Test:     []string{"CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1"},
			Interval: 30 * time.Second,
			Timeout:  10 * time.Second,
			Retries:  3,
		},
		Labels: map[string]string{
			"environment": ss.environment,
			"service":     "public-gateway",
			"component":   "gateway",
			"managed-by":  "pulumi",
		},
		DependsOn: []pulumi.Resource{
			deployment.PublicGatewayDaprSidecar,
			deployment.ContentAPIContainer,
			deployment.ServicesAPIContainer,
		},
	})
}

func (ss *ServiceStack) deployAdminGatewayContainer(deployment *ServiceDeployment) (*docker.Container, error) {
	return ss.deployServiceContainer(infrastructure.ContainerConfig{
		Name:  fmt.Sprintf("%s-admin-gateway", ss.environment),
		Image: "admin-gateway",
		Tag:   "latest",
		Environment: map[string]pulumi.StringInput{
			"ADMIN_GATEWAY_PORT":      pulumi.String("8083"),
			"ADMIN_ALLOWED_ORIGINS":   pulumi.String("*"),
			"ENVIRONMENT":             pulumi.String(ss.environment),
			"APP_VERSION":             pulumi.String(ss.config.Get("app_version")),
			"DAPR_HTTP_PORT":          pulumi.String("3504"),
			"DAPR_GRPC_PORT":          pulumi.String("50006"),
			"CONTENT_API_URL":         pulumi.String("http://localhost:3501"),
			"SERVICES_API_URL":        pulumi.String("http://localhost:3502"),
		},
		Ports: []infrastructure.ContainerPort{
			{Internal: 8083, External: 8083, Protocol: "tcp"},
		},
		Networks: []string{deployment.ServiceNetwork.Name.ToStringOutput().AsStringOutput().ToStringOutput().ApplyT(func(name interface{}) string { return name.(string) }).(string)},
		RestartPolicy: "unless-stopped",
		HealthCheck: infrastructure.ContainerHealthCheck{
			Test:     []string{"CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8083/health || exit 1"},
			Interval: 30 * time.Second,
			Timeout:  10 * time.Second,
			Retries:  3,
		},
		Labels: map[string]string{
			"environment": ss.environment,
			"service":     "admin-gateway",
			"component":   "gateway",
			"managed-by":  "pulumi",
		},
		DependsOn: []pulumi.Resource{
			deployment.AdminGatewayDaprSidecar,
			deployment.ContentAPIContainer,
			deployment.ServicesAPIContainer,
		},
	})
}

func (ss *ServiceStack) deployServiceContainer(config infrastructure.ContainerConfig) (*docker.Container, error) {
	return ss.containerRuntime.CreateContainer(config)
}

func (ss *ServiceStack) getAppPort(appID string) int {
	switch appID {
	case "content-api":
		return 8080
	case "services-api":
		return 8081
	case "public-gateway":
		return 8082
	case "admin-gateway":
		return 8083
	default:
		return 8080
	}
}

func (ss *ServiceStack) getDatabaseURL() string {
	user := ss.config.Get("postgres_user")
	password := ss.config.Get("postgres_password")
	host := "postgres"  // Container hostname
	port := ss.config.RequireInt("postgres_port")
	dbname := ss.config.Get("postgres_db")
	
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, dbname)
}

func (ss *ServiceStack) ValidateDeployment(ctx context.Context, deployment *ServiceDeployment) error {
	if deployment.ContentAPIContainer == nil {
		return fmt.Errorf("content-api container is not deployed")
	}
	
	if deployment.ServicesAPIContainer == nil {
		return fmt.Errorf("services-api container is not deployed")
	}
	
	if deployment.PublicGatewayContainer == nil {
		return fmt.Errorf("public-gateway container is not deployed")
	}
	
	if deployment.AdminGatewayContainer == nil {
		return fmt.Errorf("admin-gateway container is not deployed")
	}
	
	if deployment.ContentAPIDaprSidecar == nil {
		return fmt.Errorf("content-api Dapr sidecar is not deployed")
	}
	
	if deployment.ServicesAPIDaprSidecar == nil {
		return fmt.Errorf("services-api Dapr sidecar is not deployed")
	}
	
	if deployment.PublicGatewayDaprSidecar == nil {
		return fmt.Errorf("public-gateway Dapr sidecar is not deployed")
	}
	
	if deployment.AdminGatewayDaprSidecar == nil {
		return fmt.Errorf("admin-gateway Dapr sidecar is not deployed")
	}
	
	return nil
}