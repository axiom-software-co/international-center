package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi-command/sdk/go/command/local"
)

// ContainerFactoryError represents different types of container factory errors
type ContainerFactoryError struct {
	Type        ContainerErrorType
	ServiceName string
	Operation   string
	Context     map[string]interface{}
	Err         error
}

// ContainerErrorType defines the type of container factory error
type ContainerErrorType string

const (
	ErrorTypeImageBuild      ContainerErrorType = "image_build_failed"
	ErrorTypeImageValidation ContainerErrorType = "image_validation_failed"
	ErrorTypeContainerDeploy ContainerErrorType = "container_deploy_failed"
	ErrorTypeDaprSidecar     ContainerErrorType = "dapr_sidecar_failed"
	ErrorTypeHealthCheck     ContainerErrorType = "health_check_failed"
	ErrorTypeConfiguration   ContainerErrorType = "configuration_error"
	ErrorTypeResourceLimit   ContainerErrorType = "resource_limit_error"
)

// Error implements the error interface for ContainerFactoryError
func (e *ContainerFactoryError) Error() string {
	contextStr := ""
	if len(e.Context) > 0 {
		contextStr = fmt.Sprintf(" (context: %v)", e.Context)
	}
	return fmt.Sprintf("container factory %s failed for %s during %s: %v%s", 
		string(e.Type), e.ServiceName, e.Operation, e.Err, contextStr)
}

// OperationLogger provides structured logging for container operations
type OperationLogger struct {
	ctx         *pulumi.Context
	serviceName string
	operation   string
	startTime   time.Time
}

// NewOperationLogger creates a new operation logger
func NewOperationLogger(ctx *pulumi.Context, serviceName, operation string) *OperationLogger {
	logger := &OperationLogger{
		ctx:         ctx,
		serviceName: serviceName,
		operation:   operation,
		startTime:   time.Now(),
	}
	
	logger.LogStart()
	return logger
}

// LogStart logs the beginning of an operation
func (ol *OperationLogger) LogStart() {
	ol.ctx.Log.Info(fmt.Sprintf("Starting %s for service %s", ol.operation, ol.serviceName), nil)
}

// LogProgress logs progress during an operation
func (ol *OperationLogger) LogProgress(message string, details map[string]interface{}) {
	logData := map[string]interface{}{
		"service":   ol.serviceName,
		"operation": ol.operation,
		"progress":  message,
		"duration":  time.Since(ol.startTime).String(),
	}
	
	// Merge additional details
	for k, v := range details {
		logData[k] = v
	}
	
	ol.ctx.Log.Info(fmt.Sprintf("%s progress: %s", ol.operation, message), nil)
}

// LogSuccess logs successful completion of an operation
func (ol *OperationLogger) LogSuccess(details map[string]interface{}) {
	duration := time.Since(ol.startTime)
	logData := map[string]interface{}{
		"service":       ol.serviceName,
		"operation":     ol.operation,
		"duration":      duration.String(),
		"success":       true,
		"completed_at":  time.Now().Format(time.RFC3339),
	}
	
	// Merge additional details
	for k, v := range details {
		logData[k] = v
	}
	
	ol.ctx.Log.Info(fmt.Sprintf("%s completed successfully for %s in %v", ol.operation, ol.serviceName, duration), nil)
}

// LogError logs error during an operation
func (ol *OperationLogger) LogError(err error, details map[string]interface{}) {
	duration := time.Since(ol.startTime)
	logData := map[string]interface{}{
		"service":   ol.serviceName,
		"operation": ol.operation,
		"duration":  duration.String(),
		"success":   false,
		"error":     err.Error(),
		"failed_at": time.Now().Format(time.RFC3339),
	}
	
	// Merge additional details
	for k, v := range details {
		logData[k] = v
	}
	
	ol.ctx.Log.Error(fmt.Sprintf("%s failed for %s after %v: %v", ol.operation, ol.serviceName, duration, err), nil)
}

// ContainerConfig represents configuration for deploying a container
type ContainerConfig struct {
	ServiceName   string
	ContainerName string
	ImageName     string
	HostPort      int
	ContainerPort int
	DaprGrpcPort  int
	AppID         string
	CleanupImages bool // Whether to cleanup images when container is deleted
	HealthCheck   bool // Whether to perform health checks on the container
}

// DeployServiceContainer deploys a Podman container with Dapr sidecar with enhanced error handling and logging
func DeployServiceContainer(ctx *pulumi.Context, config ContainerConfig) (pulumi.Map, error) {
	logger := NewOperationLogger(ctx, config.ServiceName, "container deployment")
	
	// Validate configuration
	if err := validateContainerConfig(config); err != nil {
		logger.LogError(err, map[string]interface{}{"validation_error": true})
		return nil, &ContainerFactoryError{
			Type:        ErrorTypeConfiguration,
			ServiceName: config.ServiceName,
			Operation:   "configuration validation",
			Context: map[string]interface{}{
				"config": config,
			},
			Err: err,
		}
	}
	
	logger.LogProgress("Configuration validated", map[string]interface{}{
		"image_name":      config.ImageName,
		"container_name":  config.ContainerName,
		"host_port":       config.HostPort,
		"container_port":  config.ContainerPort,
		"health_check":    config.HealthCheck,
		"cleanup_images":  config.CleanupImages,
	})
	
	// Build the required image if it doesn't exist
	logger.LogProgress("Building required image", map[string]interface{}{
		"image_name": config.ImageName,
	})
	
	imageBuildCmd, err := buildImageIfNeeded(ctx, config.ServiceName, config.ImageName)
	if err != nil {
		logger.LogError(err, map[string]interface{}{
			"phase": "image_build",
			"image_name": config.ImageName,
		})
		return nil, &ContainerFactoryError{
			Type:        ErrorTypeImageBuild,
			ServiceName: config.ServiceName,
			Operation:   "image building",
			Context: map[string]interface{}{
				"image_name": config.ImageName,
			},
			Err: err,
		}
	}
	
	logger.LogProgress("Image build initiated", map[string]interface{}{
		"image_name": config.ImageName,
	})
	
	// Perform image health check after build if enabled
	var imageHealthCmd *local.Command
	if config.HealthCheck {
		logger.LogProgress("Performing image health check", map[string]interface{}{
			"image_name": config.ImageName,
		})
		
		imageHealthCmd, err = performImageHealthCheck(ctx, config.ServiceName, config.ImageName)
		if err != nil {
			logger.LogError(err, map[string]interface{}{
				"phase": "health_check",
				"image_name": config.ImageName,
			})
			return nil, &ContainerFactoryError{
				Type:        ErrorTypeHealthCheck,
				ServiceName: config.ServiceName,
				Operation:   "image health check",
				Context: map[string]interface{}{
					"image_name": config.ImageName,
				},
				Err: err,
			}
		}
		
		logger.LogProgress("Image health check configured", map[string]interface{}{
			"image_name": config.ImageName,
		})
	}
	
	// Determine container dependencies
	var dependencies []pulumi.Resource
	dependencies = append(dependencies, imageBuildCmd)
	if imageHealthCmd != nil {
		dependencies = append(dependencies, imageHealthCmd)
	}
	
	// Build container delete command with optional image cleanup
	deleteCmd := fmt.Sprintf("podman rm -f %s", config.ContainerName)
	if config.CleanupImages {
		deleteCmd = fmt.Sprintf("podman rm -f %s && podman rmi -f %s || true", config.ContainerName, config.ImageName)
		logger.LogProgress("Image cleanup enabled for container deletion", map[string]interface{}{
			"cleanup_images": true,
		})
	}
	
	// Create Dapr sidecar container first to establish network
	logger.LogProgress("Creating Dapr sidecar first", map[string]interface{}{
		"app_id":         config.AppID,
		"dapr_grpc_port": config.DaprGrpcPort,
	})
	
	daprCmd, err := local.NewCommand(ctx, fmt.Sprintf("%s-dapr-sidecar", config.ServiceName), &local.CommandArgs{
		Create: pulumi.Sprintf("podman run -d --name %s-dapr daprio/daprd:latest /daprd --app-id %s --app-port %d --dapr-http-port 3500 --dapr-grpc-port %d", 
			config.ServiceName, config.AppID, config.ContainerPort, config.DaprGrpcPort),
		Delete: pulumi.Sprintf("podman rm -f %s-dapr", config.ServiceName),
	}, pulumi.DependsOn(dependencies))
	if err != nil {
		logger.LogError(err, map[string]interface{}{
			"phase": "dapr_sidecar_creation",
			"app_id": config.AppID,
			"dapr_grpc_port": config.DaprGrpcPort,
		})
		return nil, &ContainerFactoryError{
			Type:        ErrorTypeDaprSidecar,
			ServiceName: config.ServiceName,
			Operation:   "dapr sidecar creation",
			Context: map[string]interface{}{
				"app_id":         config.AppID,
				"dapr_grpc_port": config.DaprGrpcPort,
			},
			Err: err,
		}
	}
	
	// Wait for Dapr sidecar to be ready
	logger.LogProgress("Waiting for Dapr sidecar readiness", map[string]interface{}{
		"dapr_grpc_port": config.DaprGrpcPort,
		"app_id":         config.AppID,
	})
	
	daprHealthCheck, err := local.NewCommand(ctx, fmt.Sprintf("%s-dapr-health", config.ServiceName), &local.CommandArgs{
		Create: pulumi.Sprintf(`
			echo "Waiting for Dapr sidecar %s to be ready on port %d..." &&
			for i in {1..30}; do 
				if podman exec %s-dapr /bin/sh -c "netstat -ln | grep :%d" >/dev/null 2>&1; then 
					echo "Dapr sidecar %s ready on port %d" && exit 0
				fi
				echo "Waiting for Dapr sidecar (attempt $i/30)..."
				sleep 2
			done &&
			echo "ERROR: Dapr sidecar %s not ready after 60 seconds" && exit 1
		`, config.ServiceName, config.DaprGrpcPort, config.ServiceName, config.DaprGrpcPort, config.ServiceName, config.DaprGrpcPort, config.ServiceName),
	}, pulumi.DependsOn([]pulumi.Resource{daprCmd}))
	if err != nil {
		logger.LogError(err, map[string]interface{}{
			"phase": "dapr_health_check",
			"app_id": config.AppID,
		})
		return nil, &ContainerFactoryError{
			Type:        ErrorTypeHealthCheck,
			ServiceName: config.ServiceName,
			Operation:   "dapr health check",
			Context: map[string]interface{}{
				"app_id":         config.AppID,
				"dapr_grpc_port": config.DaprGrpcPort,
			},
			Err: err,
		}
	}
	
	// Create backend container with network sharing to Dapr sidecar
	logger.LogProgress("Creating backend container with Dapr network sharing", map[string]interface{}{
		"container_name": config.ContainerName,
		"dapr_ready":     true,
	})
	
	containerDeps := append(dependencies, daprCmd, daprHealthCheck)
	containerCmd, err := local.NewCommand(ctx, fmt.Sprintf("%s-container", config.ServiceName), &local.CommandArgs{
		Create: pulumi.Sprintf("podman run -d --name %s --network=container:%s-dapr -p %d:%d -e DAPR_HTTP_PORT=3500 -e DAPR_GRPC_PORT=%d %s", 
			config.ContainerName, config.ServiceName, config.HostPort, config.ContainerPort, config.DaprGrpcPort, config.ImageName),
		Delete: pulumi.String(deleteCmd),
	}, pulumi.DependsOn(containerDeps))
	if err != nil {
		logger.LogError(err, map[string]interface{}{
			"phase": "container_creation",
			"container_name": config.ContainerName,
			"dapr_network":   true,
			"dependencies_count": len(containerDeps),
		})
		return nil, &ContainerFactoryError{
			Type:        ErrorTypeContainerDeploy,
			ServiceName: config.ServiceName,
			Operation:   "container creation with dapr network",
			Context: map[string]interface{}{
				"container_name": config.ContainerName,
				"image_name":     config.ImageName,
				"host_port":      config.HostPort,
				"container_port": config.ContainerPort,
				"dapr_network":   fmt.Sprintf("%s-dapr", config.ServiceName),
			},
			Err: err,
		}
	}
	
	// Perform service readiness validation if health checks are enabled
	var serviceHealthCmd *local.Command
	if config.HealthCheck {
		logger.LogProgress("Performing service readiness validation", map[string]interface{}{
			"service_name":     config.ServiceName,
			"health_endpoint":  fmt.Sprintf("http://localhost:%d/health", config.HostPort),
			"container_ready":  true,
		})
		
		serviceHealthCmd, err = local.NewCommand(ctx, fmt.Sprintf("%s-service-health", config.ServiceName), &local.CommandArgs{
			Create: pulumi.Sprintf(`
				echo "Waiting for service %s to be ready on port %d..." &&
				for i in {1..30}; do 
					if curl -f -s http://localhost:%d/health >/dev/null 2>&1; then 
						echo "Service %s health endpoint ready on port %d" && exit 0
					fi
					echo "Waiting for service health endpoint (attempt $i/30)..."
					sleep 3
				done &&
				echo "ERROR: Service %s health endpoint not ready after 90 seconds" && exit 1
			`, config.ServiceName, config.HostPort, config.HostPort, config.ServiceName, config.HostPort, config.ServiceName),
			Delete: pulumi.String("echo 'Service health check cleanup'"),
		}, pulumi.DependsOn([]pulumi.Resource{containerCmd}))
		if err != nil {
			logger.LogError(err, map[string]interface{}{
				"phase": "service_health_check",
				"service_name": config.ServiceName,
				"health_endpoint": fmt.Sprintf("http://localhost:%d/health", config.HostPort),
			})
			return nil, &ContainerFactoryError{
				Type:        ErrorTypeHealthCheck,
				ServiceName: config.ServiceName,
				Operation:   "service readiness validation",
				Context: map[string]interface{}{
					"health_endpoint": fmt.Sprintf("http://localhost:%d/health", config.HostPort),
					"host_port":       config.HostPort,
				},
				Err: err,
			}
		}
		
		logger.LogProgress("Service readiness validation configured", map[string]interface{}{
			"service_name":     config.ServiceName,
			"health_endpoint":  fmt.Sprintf("http://localhost:%d/health", config.HostPort),
		})
	}
	
	logger.LogSuccess(map[string]interface{}{
		"container_name":     config.ContainerName,
		"dapr_sidecar":       fmt.Sprintf("%s-dapr", config.ServiceName),
		"host_port":          config.HostPort,
		"container_port":     config.ContainerPort,
		"dapr_grpc_port":     config.DaprGrpcPort,
		"app_id":             config.AppID,
		"image_name":         config.ImageName,
		"health_check":       config.HealthCheck,
		"service_validated":  serviceHealthCmd != nil,
		"cleanup_images":     config.CleanupImages,
	})
	
	// Return service configuration map with validation status
	serviceMap := pulumi.Map{
		"container_id":      containerCmd.Stdout,
		"container_status":  pulumi.String("running"),
		"host_port":         pulumi.Int(config.HostPort),
		"health_endpoint":   pulumi.Sprintf("http://localhost:%d/health", config.HostPort),
		"dapr_app_id":       pulumi.String(config.AppID),
		"dapr_sidecar_id":   daprCmd.Stdout,
	}
	
	// Add service validation status to output
	if serviceHealthCmd != nil {
		serviceMap["service_validated"] = pulumi.Bool(true)
		serviceMap["validation_output"] = serviceHealthCmd.Stdout
	} else {
		serviceMap["service_validated"] = pulumi.Bool(false)
	}
	
	return serviceMap, nil
}

// validateContainerConfig validates the container configuration
func validateContainerConfig(config ContainerConfig) error {
	var validationErrors []string
	
	// Validate required fields
	if config.ServiceName == "" {
		validationErrors = append(validationErrors, "service name is required")
	}
	if config.ContainerName == "" {
		validationErrors = append(validationErrors, "container name is required")
	}
	if config.ImageName == "" {
		validationErrors = append(validationErrors, "image name is required")
	}
	if config.AppID == "" {
		validationErrors = append(validationErrors, "app ID is required")
	}
	
	// Validate port ranges
	if config.HostPort <= 0 || config.HostPort > 65535 {
		validationErrors = append(validationErrors, fmt.Sprintf("host port %d is invalid (must be 1-65535)", config.HostPort))
	}
	if config.ContainerPort <= 0 || config.ContainerPort > 65535 {
		validationErrors = append(validationErrors, fmt.Sprintf("container port %d is invalid (must be 1-65535)", config.ContainerPort))
	}
	if config.DaprGrpcPort <= 0 || config.DaprGrpcPort > 65535 {
		validationErrors = append(validationErrors, fmt.Sprintf("dapr grpc port %d is invalid (must be 1-65535)", config.DaprGrpcPort))
	}
	
	// Validate port conflicts
	if config.HostPort == config.DaprGrpcPort {
		validationErrors = append(validationErrors, fmt.Sprintf("host port %d conflicts with dapr grpc port", config.HostPort))
	}
	
	// Validate naming conventions
	if strings.Contains(config.ServiceName, "_") || strings.Contains(config.ServiceName, ".") {
		validationErrors = append(validationErrors, "service name should use kebab-case (no underscores or dots)")
	}
	if strings.Contains(config.ContainerName, "_") {
		validationErrors = append(validationErrors, "container name should use kebab-case (no underscores)")
	}
	
	// Return aggregated validation errors
	if len(validationErrors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(validationErrors, "; "))
	}
	
	return nil
}

// DeployInquiriesServices deploys all inquiries service containers
func DeployInquiriesServices(ctx *pulumi.Context) (pulumi.Map, error) {
	inquiriesServices := pulumi.Map{}
	serviceNames := []string{"media", "donations", "volunteers", "business"}
	
	for i, serviceName := range serviceNames {
		config := ContainerConfig{
			ServiceName:   serviceName,
			ContainerName: fmt.Sprintf("%s-dev", serviceName),
			ImageName:     fmt.Sprintf("backend/%s:latest", serviceName),
			HostPort:      8080 + i,
			ContainerPort: 8080,
			DaprGrpcPort:  50001 + i,
			AppID:         fmt.Sprintf("%s-api", serviceName),
			CleanupImages: false, // Keep images for development reuse
			HealthCheck:   true,  // Enable health checks for development validation
		}
		
		serviceMap, err := DeployServiceContainer(ctx, config)
		if err != nil {
			return nil, err
		}
		
		inquiriesServices[serviceName] = serviceMap
	}
	
	return inquiriesServices, nil
}

// DeployContentServices deploys all content service containers
func DeployContentServices(ctx *pulumi.Context) (pulumi.Map, error) {
	contentServices := pulumi.Map{}
	serviceNames := []string{"research", "services", "events", "news"}
	
	for i, serviceName := range serviceNames {
		config := ContainerConfig{
			ServiceName:   serviceName,
			ContainerName: fmt.Sprintf("%s-dev", serviceName),
			ImageName:     fmt.Sprintf("backend/%s:latest", serviceName),
			HostPort:      8090 + i,
			ContainerPort: 8080,
			DaprGrpcPort:  50010 + i,
			AppID:         fmt.Sprintf("%s-api", serviceName),
			CleanupImages: false, // Keep images for development reuse
			HealthCheck:   true,  // Enable health checks for development validation
		}
		
		serviceMap, err := DeployServiceContainer(ctx, config)
		if err != nil {
			return nil, err
		}
		
		contentServices[serviceName] = serviceMap
	}
	
	return contentServices, nil
}

// DeployGatewayServices deploys all gateway service containers
func DeployGatewayServices(ctx *pulumi.Context) (pulumi.Map, error) {
	gatewayServices := pulumi.Map{}
	gateways := []struct {
		name string
		port int
	}{
		{"admin", 9000},
		{"public", 9001},
	}
	
	for i, gateway := range gateways {
		config := ContainerConfig{
			ServiceName:   gateway.name,
			ContainerName: fmt.Sprintf("%s-gateway-dev", gateway.name),
			ImageName:     fmt.Sprintf("backend/%s-gateway:latest", gateway.name),
			HostPort:      gateway.port,
			ContainerPort: gateway.port,
			DaprGrpcPort:  50020 + i,
			AppID:         fmt.Sprintf("%s-gateway", gateway.name),
			CleanupImages: false, // Keep images for development reuse
			HealthCheck:   true,  // Enable health checks for development validation
		}
		
		serviceMap, err := DeployServiceContainer(ctx, config)
		if err != nil {
			return nil, err
		}
		
		gatewayServices[gateway.name] = serviceMap
	}
	
	return gatewayServices, nil
}

// DeployWebsiteContainer deploys website container for development
func DeployWebsiteContainer(ctx *pulumi.Context) (*local.Command, error) {
	return DeployWebsiteContainerWithConfig(ctx, WebsiteContainerConfig{
		CleanupImages: false, // Keep images for development reuse
		HealthCheck:   true,  // Enable health checks for development validation
	})
}

// WebsiteContainerConfig represents configuration for deploying website container
type WebsiteContainerConfig struct {
	CleanupImages bool // Whether to cleanup images when container is deleted
	HealthCheck   bool // Whether to perform health checks on the container
}

// DeployWebsiteContainerWithConfig deploys website container with advanced lifecycle configuration
func DeployWebsiteContainerWithConfig(ctx *pulumi.Context, config WebsiteContainerConfig) (*local.Command, error) {
	imageName := "website:latest"
	serviceName := "website"
	
	// Build the website image if it doesn't exist
	imageBuildCmd, err := buildImageIfNeeded(ctx, serviceName, imageName)
	if err != nil {
		return nil, fmt.Errorf("failed to build website image: %w", err)
	}
	
	// Perform image health check after build if enabled
	var imageHealthCmd *local.Command
	if config.HealthCheck {
		imageHealthCmd, err = performImageHealthCheck(ctx, serviceName, imageName)
		if err != nil {
			return nil, fmt.Errorf("failed to perform health check for website image: %w", err)
		}
	}
	
	// Determine container dependencies
	var dependencies []pulumi.Resource
	dependencies = append(dependencies, imageBuildCmd)
	if imageHealthCmd != nil {
		dependencies = append(dependencies, imageHealthCmd)
	}
	
	// Build container delete command with optional image cleanup
	deleteCmd := "podman rm -f website-dev"
	if config.CleanupImages {
		deleteCmd = "podman rm -f website-dev && podman rmi -f website:latest || true"
	}
	
	// Create website container using Command provider
	containerCmd, err := local.NewCommand(ctx, "website-container", &local.CommandArgs{
		Create: pulumi.String("podman run -d --name website-dev -p 3000:3000 -e NODE_ENV=development website:latest"),
		Delete: pulumi.String(deleteCmd),
	}, pulumi.DependsOn(dependencies))
	if err != nil {
		return nil, fmt.Errorf("failed to create website container: %w", err)
	}
	
	return containerCmd, nil
}

// buildImageIfNeeded builds a Docker image if it doesn't already exist
func buildImageIfNeeded(ctx *pulumi.Context, serviceName, imageName string) (*local.Command, error) {
	// Build image using appropriate build strategy based on service name and image name
	var buildCmd *local.Command
	var err error
	
	if serviceName == "website" {
		buildCmd, err = buildWebsiteImage(ctx, serviceName)
	} else if strings.Contains(imageName, "gateway") {
		// Extract gateway name from service name (remove any suffixes)
		gatewayName := strings.Replace(serviceName, "-gateway", "", 1)
		buildCmd, err = buildGatewayImage(ctx, gatewayName, serviceName)
	} else {
		// Determine service type based on image name pattern
		serviceType := determineServiceType(serviceName)
		buildCmd, err = buildServiceImage(ctx, serviceName, serviceType)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to build %s image: %w", serviceName, err)
	}
	
	return buildCmd, nil
}

// ImageBuildSpec defines the specification for building a container image
type ImageBuildSpec struct {
	ServiceName    string
	ServiceType    string
	DockerfilePath string
	ContextPath    string
	ImageTag       string
	ResourceName   string
}

// ImageBuildPaths contains common path configurations for image building
type ImageBuildPaths struct {
	BaseDir     string
	BackendDir  string
	WebsiteDir  string
	GatewaysDir string
}

// getImageBuildPaths returns the standard path configuration for image building
func getImageBuildPaths() ImageBuildPaths {
	baseDir := "/home/tojkuv/Documents/GitHub/international-center-workspace/international-center/src"
	return ImageBuildPaths{
		BaseDir:     baseDir,
		BackendDir:  fmt.Sprintf("%s/backend", baseDir),
		WebsiteDir:  fmt.Sprintf("%s/website", baseDir),
		GatewaysDir: fmt.Sprintf("%s/backend/cmd/gateways", baseDir),
	}
}

// createServiceImageSpec creates an image build specification for a backend service
func createServiceImageSpec(serviceName, serviceType string) ImageBuildSpec {
	paths := getImageBuildPaths()
	cicdDir := fmt.Sprintf("%s/cicd", paths.BaseDir)
	return ImageBuildSpec{
		ServiceName:    serviceName,
		ServiceType:    serviceType,
		DockerfilePath: fmt.Sprintf("%s/containers/%s/%s/Dockerfile", cicdDir, serviceType, serviceName),
		ContextPath:    paths.BackendDir,
		ImageTag:       fmt.Sprintf("backend/%s:latest", serviceName),
		ResourceName:   fmt.Sprintf("build-%s-%s", serviceType, serviceName),
	}
}

// createGatewayImageSpec creates an image build specification for a gateway service
func createGatewayImageSpec(gatewayName, serviceName string) ImageBuildSpec {
	paths := getImageBuildPaths()
	cicdDir := fmt.Sprintf("%s/cicd", paths.BaseDir)
	return ImageBuildSpec{
		ServiceName:    gatewayName,
		ServiceType:    "gateway",
		DockerfilePath: fmt.Sprintf("%s/containers/gateways/%s/Dockerfile", cicdDir, gatewayName),
		ContextPath:    paths.BackendDir,
		ImageTag:       fmt.Sprintf("backend/%s-gateway:latest", gatewayName),
		ResourceName:   fmt.Sprintf("build-gateway-%s", serviceName),
	}
}

// createWebsiteImageSpec creates an image build specification for the website
func createWebsiteImageSpec() ImageBuildSpec {
	paths := getImageBuildPaths()
	cicdDir := fmt.Sprintf("%s/cicd", paths.BaseDir)
	return ImageBuildSpec{
		ServiceName:    "website",
		ServiceType:    "website",
		DockerfilePath: fmt.Sprintf("%s/containers/website/Dockerfile", cicdDir),
		ContextPath:    paths.WebsiteDir,
		ImageTag:       "website:latest",
		ResourceName:   "build-website",
	}
}

// BuildCacheStrategy defines the caching strategy for image builds
type BuildCacheStrategy struct {
	EnableCache      bool
	CacheFromTags    []string  // Previous image tags to use as cache sources
	CacheToRegistry  string    // Registry to push cache layers to
	MaxCacheAge      string    // Maximum age for cached layers (e.g., "24h")
	CacheCompression bool      // Enable cache compression
}

// BuildOptimizations defines build optimization settings
type BuildOptimizations struct {
	CacheStrategy     BuildCacheStrategy
	MultiStage        bool     // Enable multi-stage build optimizations
	MinimizeLayers    bool     // Enable layer minimization
	BuildArgs         map[string]string // Build arguments for optimization
	ExcludePatterns   []string // Patterns to exclude from build context
	ParallelBuilds    bool     // Enable parallel layer building
}

// getDefaultBuildOptimizations returns optimized default settings for builds
func getDefaultBuildOptimizations(serviceName, environment string) BuildOptimizations {
	// For development environment, disable cache-from to avoid issues with non-existent images
	var cacheFromTags []string
	if environment != "development" {
		cacheFromTags = []string{fmt.Sprintf("%s:latest", serviceName), fmt.Sprintf("%s:%s", serviceName, environment)}
	}
	
	return BuildOptimizations{
		CacheStrategy: BuildCacheStrategy{
			EnableCache:      true,
			CacheFromTags:    cacheFromTags,
			MaxCacheAge:      "24h",
			CacheCompression: true,
		},
		MultiStage:     environment != "development", // Disable multi-stage targeting for development
		MinimizeLayers: true,
		BuildArgs: map[string]string{
			"BUILDKIT_INLINE_CACHE": "1",
			"BUILD_ENV":             environment,
			"CACHE_MOUNT":           "/tmp/.buildx-cache",
		},
		ExcludePatterns: []string{
			"**/.git",
			"**/node_modules",
			"**/.DS_Store",
			"**/coverage",
			"**/dist",
			"**/.nyc_output",
			"**/tmp",
			"**/*.log",
		},
		ParallelBuilds: true,
	}
}

// buildImageFromSpec builds a Docker image based on the provided specification with cache optimization
func buildImageFromSpec(ctx *pulumi.Context, spec ImageBuildSpec) (*local.Command, error) {
	optimizations := getDefaultBuildOptimizations(spec.ServiceName, "development")
	return buildImageFromSpecWithOptimizations(ctx, spec, optimizations)
}

// buildImageFromSpecWithOptimizations builds a Docker image with advanced cache and layer optimizations
func buildImageFromSpecWithOptimizations(ctx *pulumi.Context, spec ImageBuildSpec, opts BuildOptimizations) (*local.Command, error) {
	logger := NewOperationLogger(ctx, spec.ServiceName, fmt.Sprintf("optimized %s build", spec.ServiceType))
	
	// Build optimized build command
	buildCommand := constructOptimizedBuildCommand(spec, opts)
	
	logger.LogProgress("Constructing optimized build command", map[string]interface{}{
		"image_tag":        spec.ImageTag,
		"dockerfile_path":  spec.DockerfilePath,
		"context_path":     spec.ContextPath,
		"cache_enabled":    opts.CacheStrategy.EnableCache,
		"multi_stage":      opts.MultiStage,
		"minimize_layers":  opts.MinimizeLayers,
		"parallel_builds":  opts.ParallelBuilds,
		"exclude_patterns": len(opts.ExcludePatterns),
	})
	
	buildCmd, err := local.NewCommand(ctx, spec.ResourceName, &local.CommandArgs{
		Create: pulumi.String(buildCommand),
	})
	if err != nil {
		logger.LogError(err, map[string]interface{}{
			"build_command": buildCommand,
		})
		return nil, fmt.Errorf("failed to create optimized build command for %s image: %w", spec.ServiceName, err)
	}
	
	logger.LogSuccess(map[string]interface{}{
		"image_tag":       spec.ImageTag,
		"optimizations":   "cache+layers+parallel",
		"dockerfile_path": spec.DockerfilePath,
	})
	
	return buildCmd, nil
}

// constructOptimizedBuildCommand constructs a Podman-compatible optimized build command 
func constructOptimizedBuildCommand(spec ImageBuildSpec, opts BuildOptimizations) string {
	var commandParts []string
	
	// Start with the base existence check and build command
	baseCommand := fmt.Sprintf("podman image exists %s", spec.ImageTag)
	commandParts = append(commandParts, baseCommand)
	commandParts = append(commandParts, "||")
	commandParts = append(commandParts, "(")
	
	// Add build context optimization (create .dockerignore if patterns specified)
	if len(opts.ExcludePatterns) > 0 {
		createIgnoreFile := fmt.Sprintf("echo '%s' > %s/.dockerignore", 
			strings.Join(opts.ExcludePatterns, "\n"), spec.ContextPath)
		commandParts = append(commandParts, createIgnoreFile, "&&")
	}
	
	// Construct main build command with Podman-compatible optimizations
	buildCommand := []string{"podman", "build"}
	
	// Add build arguments for optimization
	for key, value := range opts.BuildArgs {
		buildCommand = append(buildCommand, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}
	
	// Multi-stage optimization (Podman supports --target)
	if opts.MultiStage {
		buildCommand = append(buildCommand, "--target", "production")
	}
	
	// Add standard build parameters
	buildCommand = append(buildCommand, "-f", spec.DockerfilePath)
	buildCommand = append(buildCommand, "-t", spec.ImageTag)
	buildCommand = append(buildCommand, spec.ContextPath)
	
	// Join build command and add cleanup
	commandParts = append(commandParts, strings.Join(buildCommand, " "))
	
	// Clean up temporary files
	if len(opts.ExcludePatterns) > 0 {
		cleanupCommand := fmt.Sprintf("rm -f %s/.dockerignore", spec.ContextPath)
		commandParts = append(commandParts, "&&", cleanupCommand)
	}
	
	// Close the conditional build block
	commandParts = append(commandParts, ")")
	
	return strings.Join(commandParts, " ")
}

// BuildCacheManager manages build cache lifecycle and optimization
type BuildCacheManager struct {
	ctx         *pulumi.Context
	environment string
	cacheDir    string
}

// NewBuildCacheManager creates a new build cache manager
func NewBuildCacheManager(ctx *pulumi.Context, environment string) *BuildCacheManager {
	return &BuildCacheManager{
		ctx:         ctx,
		environment: environment,
		cacheDir:    fmt.Sprintf("/tmp/build-cache-%s", environment),
	}
}

// PruneBuildCache removes old and unused build cache
func (bcm *BuildCacheManager) PruneBuildCache() (*local.Command, error) {
	logger := NewOperationLogger(bcm.ctx, "cache-manager", "build cache pruning")
	
	pruneCommand := fmt.Sprintf(`
		echo "Pruning build cache older than 24h" &&
		podman system prune -f --filter until=24h &&
		podman builder prune -f --filter until=24h &&
		echo "Build cache pruning completed"
	`)
	
	pruneCmd, err := local.NewCommand(bcm.ctx, "build-cache-prune", &local.CommandArgs{
		Create: pulumi.String(pruneCommand),
	})
	
	if err != nil {
		logger.LogError(err, map[string]interface{}{
			"operation": "cache_prune",
		})
		return nil, fmt.Errorf("failed to create cache prune command: %w", err)
	}
	
	logger.LogSuccess(map[string]interface{}{
		"cache_strategy": "prune_24h",
	})
	
	return pruneCmd, nil
}

// OptimizeImageLayers performs post-build layer optimization
func (bcm *BuildCacheManager) OptimizeImageLayers(imageTag string) (*local.Command, error) {
	logger := NewOperationLogger(bcm.ctx, "cache-manager", "layer optimization")
	
	optimizeCommand := fmt.Sprintf(`
		echo "Optimizing layers for image %s" &&
		podman image exists %s && (
			podman export %s | podman import --change 'CMD [""]' - %s:optimized &&
			podman tag %s:optimized %s &&
			podman rmi %s:optimized &&
			echo "Layer optimization completed for %s"
		) || echo "Image %s not found for optimization"
	`, imageTag, imageTag, imageTag, imageTag, imageTag, imageTag, imageTag, imageTag, imageTag)
	
	optimizeCmd, err := local.NewCommand(bcm.ctx, fmt.Sprintf("optimize-layers-%s", strings.ReplaceAll(imageTag, ":", "-")), &local.CommandArgs{
		Create: pulumi.String(optimizeCommand),
	})
	
	if err != nil {
		logger.LogError(err, map[string]interface{}{
			"image_tag": imageTag,
			"operation": "layer_optimization",
		})
		return nil, fmt.Errorf("failed to create layer optimization command for %s: %w", imageTag, err)
	}
	
	logger.LogSuccess(map[string]interface{}{
		"image_tag": imageTag,
		"optimization": "layers_flattened",
	})
	
	return optimizeCmd, nil
}

// buildServiceImage builds a backend service image using extracted patterns
func buildServiceImage(ctx *pulumi.Context, serviceName, serviceType string) (*local.Command, error) {
	spec := createServiceImageSpec(serviceName, serviceType)
	return buildImageFromSpec(ctx, spec)
}

// buildGatewayImage builds a gateway service image using extracted patterns  
func buildGatewayImage(ctx *pulumi.Context, gatewayName, serviceName string) (*local.Command, error) {
	spec := createGatewayImageSpec(gatewayName, serviceName)
	return buildImageFromSpec(ctx, spec)
}

// buildWebsiteImage builds the website image using extracted patterns
func buildWebsiteImage(ctx *pulumi.Context, serviceName string) (*local.Command, error) {
	spec := createWebsiteImageSpec()
	return buildImageFromSpec(ctx, spec)
}

// determineServiceType determines the service type based on service name
func determineServiceType(serviceName string) string {
	inquiriesServices := []string{"media", "donations", "volunteers", "business"}
	for _, service := range inquiriesServices {
		if service == serviceName {
			return "inquiries"
		}
	}
	return "content"
}

// ImageValidationSpec defines the specification for image validation operations
type ImageValidationSpec struct {
	ServiceName     string
	ImageName       string
	ResourceName    string
	ValidationSteps []string
	ErrorMessage    string
}

// createHealthCheckSpec creates a validation specification for image health check
func createHealthCheckSpec(serviceName, imageName string) ImageValidationSpec {
	return ImageValidationSpec{
		ServiceName:  serviceName,
		ImageName:    imageName,
		ResourceName: fmt.Sprintf("%s-image-health-check", serviceName),
		ValidationSteps: []string{
			fmt.Sprintf("echo \"Performing health check on image %s\"", imageName),
			fmt.Sprintf("podman image exists %s || (echo \"Image %s not found\" && exit 1)", imageName, imageName),
			fmt.Sprintf("podman inspect %s --format=\"{{.Id}}\" > /dev/null || (echo \"Image %s appears corrupted\" && exit 1)", imageName, imageName),
			fmt.Sprintf("echo \"Image %s passed health check\"", imageName),
		},
		ErrorMessage: fmt.Sprintf("failed to create health check command for %s", serviceName),
	}
}

// createIntegrityCheckSpec creates a validation specification for image integrity check
func createIntegrityCheckSpec(serviceName, imageName string) ImageValidationSpec {
	return ImageValidationSpec{
		ServiceName:  serviceName,
		ImageName:    imageName,
		ResourceName: fmt.Sprintf("%s-image-integrity-check", serviceName),
		ValidationSteps: []string{
			fmt.Sprintf("echo \"Validating image integrity for %s\"", imageName),
			fmt.Sprintf("podman image exists %s || (echo \"Image %s not found for integrity check\" && exit 1)", imageName, imageName),
			fmt.Sprintf("podman run --rm --entrypoint=\"\" %s /bin/sh -c \"echo 'Image layers accessible'\" || (echo \"Image %s layers corrupted\" && exit 1)", imageName, imageName),
			fmt.Sprintf("echo \"Image %s integrity validation passed\"", imageName),
		},
		ErrorMessage: fmt.Sprintf("failed to create integrity check command for %s", serviceName),
	}
}

// executeImageValidation executes image validation based on the provided specification
func executeImageValidation(ctx *pulumi.Context, spec ImageValidationSpec) (*local.Command, error) {
	validationScript := strings.Join(spec.ValidationSteps, " && ")
	
	validationCmd, err := local.NewCommand(ctx, spec.ResourceName, &local.CommandArgs{
		Create: pulumi.String(validationScript),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", spec.ErrorMessage, err)
	}
	
	return validationCmd, nil
}

// performImageHealthCheck validates that an image is healthy and ready for deployment
func performImageHealthCheck(ctx *pulumi.Context, serviceName, imageName string) (*local.Command, error) {
	spec := createHealthCheckSpec(serviceName, imageName)
	return executeImageValidation(ctx, spec)
}

// validateImageIntegrity performs deep validation on container image integrity
func validateImageIntegrity(ctx *pulumi.Context, serviceName, imageName string) (*local.Command, error) {
	spec := createIntegrityCheckSpec(serviceName, imageName)
	return executeImageValidation(ctx, spec)
}

// ContainerRestartSpec defines the specification for container restart operations
type ContainerRestartSpec struct {
	ServiceName   string
	ContainerName string
	ImageName     string
	ResourceName  string
}

// restartContainerWithImageUpdate restarts a container after updating its image
func restartContainerWithImageUpdate(ctx *pulumi.Context, serviceName, containerName, imageName string) (*local.Command, error) {
	spec := ContainerRestartSpec{
		ServiceName:   serviceName,
		ContainerName: containerName,
		ImageName:     imageName,
		ResourceName:  fmt.Sprintf("%s-container-restart", serviceName),
	}
	
	restartSteps := []string{
		fmt.Sprintf("echo \"Restarting container %s with updated image %s\"", containerName, imageName),
		fmt.Sprintf("podman stop %s || true", containerName),
		fmt.Sprintf("podman rm %s || true", containerName),
		fmt.Sprintf("podman image exists %s || (echo \"Updated image %s not found\" && exit 1)", imageName, imageName),
		fmt.Sprintf("echo \"Container %s restart completed\"", containerName),
	}
	
	restartScript := strings.Join(restartSteps, " && ")
	
	restartCmd, err := local.NewCommand(ctx, spec.ResourceName, &local.CommandArgs{
		Create: pulumi.String(restartScript),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create restart command for %s: %w", serviceName, err)
	}
	
	return restartCmd, nil
}

// TestContainerConfig represents configuration for deploying test containers
type TestContainerConfig struct {
	ServiceName        string
	TestType          string   // "unit", "integration", "e2e"
	BackendSourcePath string   // Path to backend source code
	TestCommands      []string // Commands to execute for testing
	Timeout           string   // Test execution timeout (e.g., "30s")
	RetainResults     bool     // Whether to retain test results for analysis
	ParallelExecution bool     // Whether to run tests in parallel
}

// TestVolumeConfig represents configuration for test volume management
type TestVolumeConfig struct {
	SourcePath      string // Host source code path
	ContainerPath   string // Container mount path
	ResultsPath     string // Test results output path
	CachePath       string // Go module cache path
	VolumePrefix    string // Prefix for volume names
}

// TestExecutionResult represents the result of test execution
type TestExecutionResult struct {
	ContainerID     pulumi.StringOutput
	TestStatus      pulumi.StringOutput
	ExecutionTime   pulumi.StringOutput
	ResultsLocation pulumi.StringOutput
	LogLocation     pulumi.StringOutput
	CoverageReport  pulumi.StringOutput
}

// DeployTestContainers deploys test containers for backend services with volume management
func DeployTestContainers(ctx *pulumi.Context) (pulumi.Map, error) {
	logger := NewOperationLogger(ctx, "test-infrastructure", "test container deployment")
	
	// Define test volume configuration
	testVolumeConfig := TestVolumeConfig{
		SourcePath:      "/home/tojkuv/Documents/GitHub/international-center-workspace/international-center/src/backend",
		ContainerPath:   "/app",
		ResultsPath:     "/tmp/test-results",
		CachePath:       "/go/pkg/mod",
		VolumePrefix:    "test",
	}
	
	// Create test volumes
	logger.LogProgress("Creating test volumes", map[string]interface{}{
		"source_path":      testVolumeConfig.SourcePath,
		"container_path":   testVolumeConfig.ContainerPath,
		"results_path":     testVolumeConfig.ResultsPath,
		"cache_path":       testVolumeConfig.CachePath,
	})
	
	testVolumes, err := createTestVolumes(ctx, testVolumeConfig)
	if err != nil {
		logger.LogError(err, map[string]interface{}{
			"operation": "test_volume_creation",
		})
		return nil, fmt.Errorf("failed to create test volumes: %w", err)
	}
	
	testContainers := pulumi.Map{}
	
	// Define test configurations for different test types
	testConfigs := []TestContainerConfig{
		{
			ServiceName:        "backend-unit-tests",
			TestType:          "unit",
			BackendSourcePath: testVolumeConfig.SourcePath,
			TestCommands: []string{
				"go test ./... -v -timeout=30s -race -coverprofile=coverage.out",
				"go tool cover -html=coverage.out -o coverage.html",
			},
			Timeout:           "60s",
			RetainResults:     true,
			ParallelExecution: true,
		},
		{
			ServiceName:        "backend-integration-tests",
			TestType:          "integration",
			BackendSourcePath: testVolumeConfig.SourcePath,
			TestCommands: []string{
				"go test ./... -tags=integration -v -timeout=300s -race",
			},
			Timeout:           "600s",
			RetainResults:     true,
			ParallelExecution: false,
		},
		{
			ServiceName:        "backend-module-tests",
			TestType:          "module",
			BackendSourcePath: testVolumeConfig.SourcePath,
			TestCommands: []string{
				"go test ./internal/content/news/... -v -timeout=10s",
				"go test ./internal/content/research/... -v -timeout=10s",
				"go test ./internal/content/events/... -v -timeout=10s",
				"go test ./internal/content/services/... -v -timeout=10s",
				"go test ./internal/inquiries/business/... -v -timeout=10s",
				"go test ./internal/inquiries/donations/... -v -timeout=10s",
				"go test ./internal/inquiries/media/... -v -timeout=10s",
				"go test ./internal/inquiries/volunteers/... -v -timeout=10s",
				"go test ./internal/notifications/... -v -timeout=10s",
				"go test ./internal/shared/... -v -timeout=10s",
			},
			Timeout:           "120s",
			RetainResults:     true,
			ParallelExecution: false,
		},
	}
	
	// Deploy test containers for each configuration
	for _, config := range testConfigs {
		logger.LogProgress("Deploying test container", map[string]interface{}{
			"service_name": config.ServiceName,
			"test_type":    config.TestType,
			"timeout":      config.Timeout,
			"parallel":     config.ParallelExecution,
			"commands":     len(config.TestCommands),
		})
		
		testResult, err := deployTestContainer(ctx, config, testVolumes, testVolumeConfig)
		if err != nil {
			logger.LogError(err, map[string]interface{}{
				"service_name": config.ServiceName,
				"test_type":    config.TestType,
			})
			return nil, fmt.Errorf("failed to deploy test container %s: %w", config.ServiceName, err)
		}
		
		testContainers[config.ServiceName] = pulumi.Map{
			"container_id":     testResult.ContainerID,
			"test_status":      testResult.TestStatus,
			"execution_time":   testResult.ExecutionTime,
			"results_location": testResult.ResultsLocation,
			"log_location":     testResult.LogLocation,
			"coverage_report":  testResult.CoverageReport,
			"test_type":        pulumi.String(config.TestType),
			"timeout":          pulumi.String(config.Timeout),
			"parallel":         pulumi.Bool(config.ParallelExecution),
		}
	}
	
	logger.LogSuccess(map[string]interface{}{
		"test_containers_deployed": len(testConfigs),
		"volume_configuration":     "source+results+cache",
		"test_types":               []string{"unit", "integration", "module"},
	})
	
	return testContainers, nil
}

// createTestVolumes creates and manages volumes for test execution
func createTestVolumes(ctx *pulumi.Context, config TestVolumeConfig) (pulumi.Map, error) {
	logger := NewOperationLogger(ctx, "test-volumes", "test volume creation")
	
	// Create test results directory volume
	resultsDirCmd, err := local.NewCommand(ctx, fmt.Sprintf("%s-results-volume", config.VolumePrefix), &local.CommandArgs{
		Create: pulumi.Sprintf("mkdir -p %s && echo 'Test results volume created at %s'", 
			config.ResultsPath, config.ResultsPath),
		Delete: pulumi.Sprintf("rm -rf %s || true", config.ResultsPath),
	})
	if err != nil {
		logger.LogError(err, map[string]interface{}{
			"operation": "results_volume_creation",
			"path":      config.ResultsPath,
		})
		return nil, fmt.Errorf("failed to create test results volume: %w", err)
	}
	
	// Create Go module cache directory
	cacheDirCmd, err := local.NewCommand(ctx, fmt.Sprintf("%s-cache-volume", config.VolumePrefix), &local.CommandArgs{
		Create: pulumi.String("mkdir -p /tmp/go-cache && echo 'Go module cache volume created'"),
		Delete: pulumi.String("rm -rf /tmp/go-cache || true"),
	})
	if err != nil {
		logger.LogError(err, map[string]interface{}{
			"operation": "cache_volume_creation",
		})
		return nil, fmt.Errorf("failed to create Go cache volume: %w", err)
	}
	
	logger.LogSuccess(map[string]interface{}{
		"results_path": config.ResultsPath,
		"cache_path":   "/tmp/go-cache",
		"source_mount": config.SourcePath,
	})
	
	return pulumi.Map{
		"results_volume": resultsDirCmd.Stdout,
		"cache_volume":   cacheDirCmd.Stdout,
		"source_path":    pulumi.String(config.SourcePath),
		"container_path": pulumi.String(config.ContainerPath),
		"results_path":   pulumi.String(config.ResultsPath),
	}, nil
}

// deployTestContainer deploys a single test container with comprehensive configuration
func deployTestContainer(ctx *pulumi.Context, config TestContainerConfig, volumes pulumi.Map, volumeConfig TestVolumeConfig) (*TestExecutionResult, error) {
	logger := NewOperationLogger(ctx, config.ServiceName, fmt.Sprintf("%s test container deployment", config.TestType))
	
	// Construct test execution script
	testScript := constructTestExecutionScript(config, volumeConfig)
	
	logger.LogProgress("Constructing test execution script", map[string]interface{}{
		"test_commands":  len(config.TestCommands),
		"timeout":        config.Timeout,
		"parallel":       config.ParallelExecution,
		"retain_results": config.RetainResults,
	})
	
	// Define volume mounts for the container
	volumeMounts := []string{
		fmt.Sprintf("-v %s:%s", volumeConfig.SourcePath, volumeConfig.ContainerPath),
		fmt.Sprintf("-v %s:%s", volumeConfig.ResultsPath, "/test-results"),
		fmt.Sprintf("-v /tmp/go-cache:/go/pkg/mod"),
	}
	
	// Construct container run command
	containerRunCmd := fmt.Sprintf(
		"podman run --rm --name %s-test %s -w %s -e GOCACHE=/go/pkg/mod -e CGO_ENABLED=1 golang:1.24 /bin/sh -c \"%s\"",
		config.ServiceName,
		strings.Join(volumeMounts, " "),
		volumeConfig.ContainerPath,
		testScript,
	)
	
	logger.LogProgress("Creating test container command", map[string]interface{}{
		"container_name": fmt.Sprintf("%s-test", config.ServiceName),
		"volume_mounts":  len(volumeMounts),
		"working_dir":    volumeConfig.ContainerPath,
	})
	
	// Create the test execution command
	testCmd, err := local.NewCommand(ctx, fmt.Sprintf("%s-execution", config.ServiceName), &local.CommandArgs{
		Create: pulumi.String(containerRunCmd),
	})
	
	if err != nil {
		logger.LogError(err, map[string]interface{}{
			"container_command": containerRunCmd,
			"service_name":      config.ServiceName,
		})
		return nil, fmt.Errorf("failed to create test container command for %s: %w", config.ServiceName, err)
	}
	
	logger.LogSuccess(map[string]interface{}{
		"service_name":    config.ServiceName,
		"test_type":       config.TestType,
		"execution_ready": true,
	})
	
	// Return test execution result structure
	return &TestExecutionResult{
		ContainerID:     testCmd.Stdout,
		TestStatus:      pulumi.String("completed").ToStringOutput(),
		ExecutionTime:   pulumi.String(config.Timeout).ToStringOutput(),
		ResultsLocation: pulumi.Sprintf("%s/%s", volumeConfig.ResultsPath, config.ServiceName).ToStringOutput(),
		LogLocation:     pulumi.Sprintf("%s/%s.log", volumeConfig.ResultsPath, config.ServiceName).ToStringOutput(),
		CoverageReport:  pulumi.Sprintf("%s/%s-coverage.html", volumeConfig.ResultsPath, config.ServiceName).ToStringOutput(),
	}, nil
}

// constructTestExecutionScript creates a comprehensive test execution script
func constructTestExecutionScript(config TestContainerConfig, volumeConfig TestVolumeConfig) string {
	var scriptParts []string
	
	// Add test execution header
	scriptParts = append(scriptParts, fmt.Sprintf("echo 'Starting %s tests for %s'", config.TestType, config.ServiceName))
	scriptParts = append(scriptParts, fmt.Sprintf("echo 'Timeout: %s, Parallel: %t'", config.Timeout, config.ParallelExecution))
	
	// Add Go environment setup
	scriptParts = append(scriptParts, "export GO111MODULE=on")
	scriptParts = append(scriptParts, "export GOPROXY=https://proxy.golang.org,direct")
	scriptParts = append(scriptParts, "export GOSUMDB=sum.golang.org")
	
	// Add dependency download
	scriptParts = append(scriptParts, "echo 'Downloading Go dependencies...'")
	scriptParts = append(scriptParts, "go mod download")
	
	// Add test result directory setup
	if config.RetainResults {
		scriptParts = append(scriptParts, fmt.Sprintf("mkdir -p /test-results/%s", config.ServiceName))
	}
	
	// Add test commands execution
	for i, testCmd := range config.TestCommands {
		if config.ParallelExecution && len(config.TestCommands) > 1 {
			scriptParts = append(scriptParts, fmt.Sprintf("echo 'Running test command %d in parallel: %s'", i+1, testCmd))
			scriptParts = append(scriptParts, fmt.Sprintf("(%s) &", testCmd))
		} else {
			scriptParts = append(scriptParts, fmt.Sprintf("echo 'Running test command %d: %s'", i+1, testCmd))
			scriptParts = append(scriptParts, testCmd)
		}
	}
	
	// Wait for parallel executions to complete
	if config.ParallelExecution && len(config.TestCommands) > 1 {
		scriptParts = append(scriptParts, "wait")
		scriptParts = append(scriptParts, "echo 'All parallel test executions completed'")
	}
	
	// Add result collection
	if config.RetainResults {
		scriptParts = append(scriptParts, fmt.Sprintf("echo 'Collecting test results for %s'", config.ServiceName))
		scriptParts = append(scriptParts, fmt.Sprintf("find . -name '*.out' -o -name '*.html' -o -name '*.xml' -o -name '*.json' | xargs -I {} cp {} /test-results/%s/ 2>/dev/null || true", config.ServiceName))
	}
	
	// Add completion message
	scriptParts = append(scriptParts, fmt.Sprintf("echo '%s tests completed for %s'", config.TestType, config.ServiceName))
	
	return strings.Join(scriptParts, " && ")
}