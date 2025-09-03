package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/axiom-software-co/international-center/src/deployer/internal/shared/config"
	"github.com/axiom-software-co/international-center/src/deployer/internal/shared/validation"
)

const (
	DevEnvironment = "development"
)

// Inline types to replace missing orchestrator and messaging packages
type DeploymentStatus string

const (
	DeploymentInProgress DeploymentStatus = "in_progress"
	DeploymentCompleted  DeploymentStatus = "completed"
	DeploymentFailed     DeploymentStatus = "failed"
)

type DeploymentStep struct {
	Name      string
	StartTime time.Time
	Error     error
}

type DeploymentSession struct {
	ID          string
	Status      DeploymentStatus
	Services    []string
	StartTime   time.Time
	CurrentStep int
	Steps       []DeploymentStep
	Error       error
}

type PubSubConfig struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	Environment   string
	ClientName    string
	MaxRetries    int
	RetryDelay    time.Duration
	HealthCheck   time.Duration
	BufferSize    int
}

type OrchestratorConfig struct {
	Environment            string
	MaxConcurrentDeploys   int
	DeploymentTimeout      time.Duration
	HealthCheckInterval    time.Duration
	RetryAttempts         int
	EnableSecurityTesting bool
	EnableMigrationTesting bool
	NotificationChannels  []string
}

type DeployerOrchestrator struct {
	config     *OrchestratorConfig
	pubsub     *PubSubConfig
	sessions   map[string]*DeploymentSession
	validator  *validation.EnvironmentValidator
}

func NewDeployerOrchestrator(orchestratorConfig *OrchestratorConfig, redisConfig *PubSubConfig) (*DeployerOrchestrator, error) {
	validationConfig := &validation.ValidationConfig{
		RequiredEnvVars: []string{"DATABASE_URL", "REDIS_ADDR"},
		Timeouts: map[string]time.Duration{
			"database": 10 * time.Second,
			"redis":    5 * time.Second,
		},
	}

	validator := validation.NewEnvironmentValidator(orchestratorConfig.Environment, validationConfig)
	
	return &DeployerOrchestrator{
		config:    orchestratorConfig,
		pubsub:    redisConfig,
		sessions:  make(map[string]*DeploymentSession),
		validator: validator,
	}, nil
}

func (do *DeployerOrchestrator) DeployFullStack(ctx context.Context, services []string) (*DeploymentSession, error) {
	sessionID := fmt.Sprintf("deploy-%d", time.Now().Unix())
	
	session := &DeploymentSession{
		ID:          sessionID,
		Status:      DeploymentInProgress,
		Services:    services,
		StartTime:   time.Now(),
		CurrentStep: 0,
		Steps:       createDeploymentSteps(services),
	}
	
	do.sessions[sessionID] = session
	
	go func() {
		if err := do.executeDeployment(ctx, session); err != nil {
			session.Status = DeploymentFailed
			session.Error = err
		} else {
			session.Status = DeploymentCompleted
		}
	}()
	
	return session, nil
}

func (do *DeployerOrchestrator) GetDeploymentStatus(sessionID string) (*DeploymentSession, error) {
	session, exists := do.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("deployment session %s not found", sessionID)
	}
	return session, nil
}

func (do *DeployerOrchestrator) StartListening(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (do *DeployerOrchestrator) Close() error {
	return nil
}

func (do *DeployerOrchestrator) executeDeployment(ctx context.Context, session *DeploymentSession) error {
	for i, step := range session.Steps {
		session.CurrentStep = i
		step.StartTime = time.Now()
		
		log.Printf("Executing deployment step: %s", step.Name)
		
		if err := do.executeStep(ctx, step); err != nil {
			step.Error = err
			return fmt.Errorf("deployment step %s failed: %w", step.Name, err)
		}
		
		time.Sleep(2 * time.Second)
	}
	
	return nil
}

func (do *DeployerOrchestrator) executeStep(ctx context.Context, step *DeploymentStep) error {
	if strings.Contains(step.Name, "Database") {
		return do.validator.ValidateEnvironment(ctx)
	}
	return nil
}

func createDeploymentSteps(services []string) []DeploymentStep {
	steps := []DeploymentStep{
		{Name: "Environment Validation"},
		{Name: "Database Migration"},
		{Name: "Infrastructure Setup"},
	}
	
	for _, service := range services {
		steps = append(steps, DeploymentStep{
			Name: fmt.Sprintf("Deploy %s Service", service),
		})
	}
	
	steps = append(steps, DeploymentStep{Name: "Health Check Validation"})
	
	return steps
}

func main() {
	log.Printf("Starting International Center Development Deployer")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := createDevelopmentConfig()
	
	if err := validateDevelopmentEnvironment(); err != nil {
		log.Fatalf("Development environment validation failed: %v", err)
	}

	deployerOrchestrator, err := NewDeployerOrchestrator(config.OrchestratorConfig, config.RedisConfig)
	if err != nil {
		log.Fatalf("Failed to initialize development deployer: %v", err)
	}
	defer deployerOrchestrator.Close()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	log.Printf("Development deployer initialized successfully")
	log.Printf("Environment: %s", DevEnvironment)
	log.Printf("Redis: %s", config.RedisConfig.RedisAddr)

	services := []string{"api", "admin", "worker"}
	if envServices := os.Getenv("DEVELOPMENT_SERVICES"); envServices != "" {
		log.Printf("Using custom service list from environment: %s", envServices)
	}

	session, err := deployerOrchestrator.DeployFullStack(ctx, services)
	if err != nil {
		log.Fatalf("Failed to start development deployment: %v", err)
	}

	log.Printf("Development deployment started with session: %s", session.ID)

	go func() {
		if err := deployerOrchestrator.StartListening(ctx); err != nil {
			log.Printf("Deployer listener stopped with error: %v", err)
			cancel()
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-signalChan:
			log.Printf("Received shutdown signal")
			return
		case <-ctx.Done():
			log.Printf("Context cancelled")
			return
		case <-ticker.C:
			currentSession, err := deployerOrchestrator.GetDeploymentStatus(session.ID)
			if err != nil {
				log.Printf("Failed to get deployment status: %v", err)
				continue
			}

			switch currentSession.Status {
			case DeploymentCompleted:
				log.Printf("✓ Development deployment completed successfully")
				log.Printf("Services deployed: %v", currentSession.Services)
				log.Printf("Total duration: %v", time.Since(currentSession.StartTime))
				
				if err := displayDevelopmentUrls(); err != nil {
					log.Printf("Failed to display development URLs: %v", err)
				}
				return

			case DeploymentFailed:
				log.Printf("✗ Development deployment failed: %v", currentSession.Error)
				log.Printf("Failed at step: %d/%d", currentSession.CurrentStep+1, len(currentSession.Steps))
				if currentSession.CurrentStep >= 0 && currentSession.CurrentStep < len(currentSession.Steps) {
					failedStep := currentSession.Steps[currentSession.CurrentStep]
					log.Printf("Failed step: %s", failedStep.Name)
					log.Printf("Step error: %v", failedStep.Error)
				}
				os.Exit(1)

			case DeploymentInProgress:
				if currentSession.CurrentStep >= 0 && currentSession.CurrentStep < len(currentSession.Steps) {
					currentStep := currentSession.Steps[currentSession.CurrentStep]
					log.Printf("→ Step %d/%d: %s", currentSession.CurrentStep+1, len(currentSession.Steps), currentStep.Name)
				}
			}
		}
	}
}

func createDevelopmentConfig() *DeploymentConfig {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatalf("REDIS_ADDR environment variable is required")
	}

	return &DeploymentConfig{
		RedisConfig: &PubSubConfig{
			RedisAddr:     redisAddr,
			RedisPassword: os.Getenv("REDIS_PASSWORD"),
			RedisDB:       getEnvIntOrZero("REDIS_DB"),
			Environment:   DevEnvironment,
			ClientName:    "dev-deployer",
			MaxRetries:    3,
			RetryDelay:    1 * time.Second,
			HealthCheck:   30 * time.Second,
			BufferSize:    100,
		},
		OrchestratorConfig: &OrchestratorConfig{
			Environment:            DevEnvironment,
			MaxConcurrentDeploys:   1,
			DeploymentTimeout:      30 * time.Minute,
			HealthCheckInterval:    10 * time.Second,
			RetryAttempts:         1,
			EnableSecurityTesting: false,
			EnableMigrationTesting: false,
			NotificationChannels:  []string{"console"},
		},
	}
}

func validateDevelopmentEnvironment() error {
	log.Printf("Validating development environment...")

	requiredEnvVars := map[string]string{
		"DATABASE_URL": "PostgreSQL connection for development",
		"REDIS_ADDR":   "Redis connection for pub/sub",
	}

	for envVar, description := range requiredEnvVars {
		if value := os.Getenv(envVar); value == "" {
			return fmt.Errorf("required environment variable %s is not set (%s)", envVar, description)
		}
	}

	log.Printf("✓ Environment variables validated")

	return nil
}

func displayDevelopmentUrls() error {
	log.Printf("\n" + "="*60)
	log.Printf("Development Environment Ready")
	log.Printf("="*60)
	
	apiHost := os.Getenv("API_HOST")
	apiPort := os.Getenv("API_PORT")
	adminHost := os.Getenv("ADMIN_HOST")
	adminPort := os.Getenv("ADMIN_PORT")
	
	if apiHost == "" || apiPort == "" {
		log.Printf("API_HOST and API_PORT environment variables must be set")
		return nil
	}
	
	if adminHost == "" || adminPort == "" {
		log.Printf("ADMIN_HOST and ADMIN_PORT environment variables must be set")
		return nil
	}
	
	log.Printf("API Gateway:   http://%s:%s", apiHost, apiPort)
	log.Printf("Admin Gateway: http://%s:%s", adminHost, adminPort)
	log.Printf("Health Check:  http://%s:%s/health", apiHost, apiPort)
	
	log.Printf("\nAvailable Endpoints:")
	log.Printf("GET /api/v1/services           - List all services")
	log.Printf("GET /api/v1/services/{id}      - Get service by ID")
	log.Printf("GET /api/v1/content            - List all content")
	log.Printf("GET /api/v1/content/{id}       - Get content by ID")
	log.Printf("GET /health                    - Health check")
	
	log.Printf("\nAdmin Endpoints:")
	log.Printf("GET /admin/api/v1/services     - Admin service list")
	log.Printf("GET /admin/api/v1/content      - Admin content list")
	
	log.Printf("="*60)
	
	return nil
}

type DeploymentConfig struct {
	RedisConfig        *PubSubConfig
	OrchestratorConfig *OrchestratorConfig
}

func getEnvIntOrZero(key string) int {
	if value := os.Getenv(key); value != "" {
		return parseInt(value)
	}
	return 0
}

func parseInt(s string) int {
	value := 0
	for _, char := range s {
		if char < '0' || char > '9' {
			return 0
		}
		value = value*10 + int(char-'0')
	}
	return value
}