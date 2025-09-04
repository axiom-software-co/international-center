package validation

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type EnvironmentValidator struct {
	environment string
	config      *ValidationConfig
}

type ValidationConfig struct {
	DatabaseURL    string
	RedisURL       string
	VaultURL       string
	AzuriteURL     string
	GrafanaURL     string
	LokiURL        string
	Timeouts       map[string]time.Duration
	RequiredEnvVars []string
}

type EnvironmentValidationResult struct {
	Environment         string
	IsValid             bool
	Errors             []ValidationError
	Warnings           []ValidationWarning
	DatabaseStatus      ServiceStatus
	RedisStatus        ServiceStatus
	VaultStatus        ServiceStatus
	AzuriteStatus      ServiceStatus
	GrafanaStatus      ServiceStatus
	LokiStatus         ServiceStatus
	NetworkConnectivity bool
	RequiredPorts       []PortCheck
	EnvironmentVars     map[string]EnvVarStatus
}

type ServiceStatus struct {
	Name        string
	URL         string
	Available   bool
	ResponseTime time.Duration
	Error       error
	Version     string
	Health      HealthStatus
}

type HealthStatus struct {
	Status    string
	Timestamp time.Time
	Details   map[string]interface{}
}

type PortCheck struct {
	Port      int
	Available bool
	Service   string
	Error     error
}

type EnvVarStatus struct {
	Name     string
	Present  bool
	Value    string
	Required bool
	Valid    bool
	Error    error
}

type ValidationError struct {
	Component string
	Message   string
	Critical  bool
	Details   map[string]interface{}
}

type ValidationWarning struct {
	Component string
	Message   string
	Severity  string
}

func NewEnvironmentValidator(environment string, config *ValidationConfig) *EnvironmentValidator {
	return &EnvironmentValidator{
		environment: environment,
		config:      config,
	}
}

func (ev *EnvironmentValidator) ValidateEnvironment(ctx context.Context) (*EnvironmentValidationResult, error) {
	result := &EnvironmentValidationResult{
		Environment:     ev.environment,
		EnvironmentVars: make(map[string]EnvVarStatus),
		IsValid:        true,
	}

	result.DatabaseStatus = ev.validateDatabaseConnection(ctx)
	if !result.DatabaseStatus.Available {
		result.Errors = append(result.Errors, ValidationError{
			Component: "database",
			Message:   fmt.Sprintf("Database connection failed: %v", result.DatabaseStatus.Error),
			Critical:  true,
		})
		result.IsValid = false
	}

	result.RedisStatus = ev.validateRedisConnection(ctx)
	if !result.RedisStatus.Available {
		result.Errors = append(result.Errors, ValidationError{
			Component: "redis",
			Message:   fmt.Sprintf("Redis connection failed: %v", result.RedisStatus.Error),
			Critical:  true,
		})
		result.IsValid = false
	}

	result.VaultStatus = ev.validateVaultConnection(ctx)
	if !result.VaultStatus.Available {
		result.Errors = append(result.Errors, ValidationError{
			Component: "vault",
			Message:   fmt.Sprintf("Vault connection failed: %v", result.VaultStatus.Error),
			Critical:  true,
		})
		result.IsValid = false
	}

	result.AzuriteStatus = ev.validateAzuriteConnection(ctx)
	if !result.AzuriteStatus.Available {
		result.Errors = append(result.Errors, ValidationError{
			Component: "azurite",
			Message:   fmt.Sprintf("Azurite connection failed: %v", result.AzuriteStatus.Error),
			Critical:  true,
		})
		result.IsValid = false
	}

	result.GrafanaStatus = ev.validateGrafanaConnection(ctx)
	if !result.GrafanaStatus.Available {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Component: "grafana",
			Message:   fmt.Sprintf("Grafana connection failed: %v", result.GrafanaStatus.Error),
			Severity:  "medium",
		})
	}

	result.LokiStatus = ev.validateLokiConnection(ctx)
	if !result.LokiStatus.Available {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Component: "loki",
			Message:   fmt.Sprintf("Loki connection failed: %v", result.LokiStatus.Error),
			Severity:  "medium",
		})
	}

	result.RequiredPorts = ev.validateRequiredPorts(ctx)
	for _, portCheck := range result.RequiredPorts {
		if !portCheck.Available {
			result.Errors = append(result.Errors, ValidationError{
				Component: "network",
				Message:   fmt.Sprintf("Required port %d for %s is not available: %v", portCheck.Port, portCheck.Service, portCheck.Error),
				Critical:  ev.isPortCritical(portCheck.Port),
			})
			result.IsValid = false
		}
	}

	ev.validateEnvironmentVariables(result)
	for _, envVar := range result.EnvironmentVars {
		if envVar.Required && (!envVar.Present || !envVar.Valid) {
			result.Errors = append(result.Errors, ValidationError{
				Component: "environment",
				Message:   fmt.Sprintf("Required environment variable %s is missing or invalid: %v", envVar.Name, envVar.Error),
				Critical:  true,
			})
			result.IsValid = false
		}
	}

	result.NetworkConnectivity = ev.validateNetworkConnectivity(ctx)
	if !result.NetworkConnectivity {
		result.Errors = append(result.Errors, ValidationError{
			Component: "network",
			Message:   "Network connectivity validation failed",
			Critical:  true,
		})
		result.IsValid = false
	}

	return result, nil
}

func (ev *EnvironmentValidator) validateDatabaseConnection(ctx context.Context) ServiceStatus {
	status := ServiceStatus{
		Name: "PostgreSQL",
		URL:  ev.config.DatabaseURL,
	}

	timeout := ev.getTimeout("database")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()

	conn, err := net.DialTimeout("tcp", ev.extractHostPort(ev.config.DatabaseURL), timeout)
	if err != nil {
		status.Error = err
		return status
	}
	defer conn.Close()

	status.Available = true
	status.ResponseTime = time.Since(start)

	healthEndpoint := ev.buildHealthEndpoint(ev.config.DatabaseURL, "/health")
	health, err := ev.checkHTTPHealth(ctx, healthEndpoint)
	if err == nil {
		status.Health = health
	}

	return status
}

func (ev *EnvironmentValidator) validateRedisConnection(ctx context.Context) ServiceStatus {
	status := ServiceStatus{
		Name: "Redis",
		URL:  ev.config.RedisURL,
	}

	timeout := ev.getTimeout("redis")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()

	conn, err := net.DialTimeout("tcp", ev.extractHostPort(ev.config.RedisURL), timeout)
	if err != nil {
		status.Error = err
		return status
	}
	defer conn.Close()

	status.Available = true
	status.ResponseTime = time.Since(start)

	return status
}

func (ev *EnvironmentValidator) validateVaultConnection(ctx context.Context) ServiceStatus {
	status := ServiceStatus{
		Name: "Vault",
		URL:  ev.config.VaultURL,
	}

	timeout := ev.getTimeout("vault")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()

	healthEndpoint := fmt.Sprintf("%s/v1/sys/health", ev.config.VaultURL)
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(healthEndpoint)
	if err != nil {
		status.Error = err
		return status
	}
	defer resp.Body.Close()

	status.Available = resp.StatusCode == http.StatusOK
	status.ResponseTime = time.Since(start)

	if !status.Available {
		status.Error = fmt.Errorf("vault health check returned status %d", resp.StatusCode)
	}

	return status
}

func (ev *EnvironmentValidator) validateAzuriteConnection(ctx context.Context) ServiceStatus {
	status := ServiceStatus{
		Name: "Azurite",
		URL:  ev.config.AzuriteURL,
	}

	timeout := ev.getTimeout("azurite")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()

	conn, err := net.DialTimeout("tcp", ev.extractHostPort(ev.config.AzuriteURL), timeout)
	if err != nil {
		status.Error = err
		return status
	}
	defer conn.Close()

	status.Available = true
	status.ResponseTime = time.Since(start)

	return status
}

func (ev *EnvironmentValidator) validateGrafanaConnection(ctx context.Context) ServiceStatus {
	status := ServiceStatus{
		Name: "Grafana",
		URL:  ev.config.GrafanaURL,
	}

	timeout := ev.getTimeout("grafana")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()

	healthEndpoint := fmt.Sprintf("%s/api/health", ev.config.GrafanaURL)
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(healthEndpoint)
	if err != nil {
		status.Error = err
		return status
	}
	defer resp.Body.Close()

	status.Available = resp.StatusCode == http.StatusOK
	status.ResponseTime = time.Since(start)

	if !status.Available {
		status.Error = fmt.Errorf("grafana health check returned status %d", resp.StatusCode)
	}

	return status
}

func (ev *EnvironmentValidator) validateLokiConnection(ctx context.Context) ServiceStatus {
	status := ServiceStatus{
		Name: "Loki",
		URL:  ev.config.LokiURL,
	}

	timeout := ev.getTimeout("loki")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()

	healthEndpoint := fmt.Sprintf("%s/ready", ev.config.LokiURL)
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(healthEndpoint)
	if err != nil {
		status.Error = err
		return status
	}
	defer resp.Body.Close()

	status.Available = resp.StatusCode == http.StatusOK
	status.ResponseTime = time.Since(start)

	if !status.Available {
		status.Error = fmt.Errorf("loki health check returned status %d", resp.StatusCode)
	}

	return status
}

func (ev *EnvironmentValidator) validateRequiredPorts(ctx context.Context) []PortCheck {
	requiredPorts := ev.getRequiredPorts()
	checks := make([]PortCheck, 0, len(requiredPorts))

	for port, service := range requiredPorts {
		check := PortCheck{
			Port:    port,
			Service: service,
		}

		timeout := ev.getTimeout("port")
		host := os.Getenv("SERVICE_HOST")
		if host == "" {
			check.Error = fmt.Errorf("SERVICE_HOST environment variable not set")
			check.Available = false
		} else {
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
			if err != nil {
				check.Error = err
				check.Available = false
			} else {
				check.Available = true
				conn.Close()
			}
		}

		checks = append(checks, check)
	}

	return checks
}

func (ev *EnvironmentValidator) validateEnvironmentVariables(result *EnvironmentValidationResult) {
	for _, envVar := range ev.config.RequiredEnvVars {
		status := EnvVarStatus{
			Name:     envVar,
			Required: true,
		}

		value, present := os.LookupEnv(envVar)
		status.Present = present
		status.Value = value

		if present {
			status.Valid = ev.validateEnvVarValue(envVar, value)
			if !status.Valid {
				status.Error = fmt.Errorf("environment variable %s has invalid value", envVar)
			}
		} else {
			status.Error = fmt.Errorf("required environment variable %s is not set", envVar)
		}

		result.EnvironmentVars[envVar] = status
	}
}

func (ev *EnvironmentValidator) validateNetworkConnectivity(ctx context.Context) bool {
	timeout := ev.getTimeout("network")
	
	// Get test endpoints from environment variables
	testEndpoints := []string{}
	
	if endpoint1 := os.Getenv("NETWORK_TEST_ENDPOINT_1"); endpoint1 != "" {
		testEndpoints = append(testEndpoints, endpoint1)
	}
	if endpoint2 := os.Getenv("NETWORK_TEST_ENDPOINT_2"); endpoint2 != "" {
		testEndpoints = append(testEndpoints, endpoint2)
	}
	
	// If no test endpoints configured, skip network connectivity test
	if len(testEndpoints) == 0 {
		return true
	}

	for _, endpoint := range testEndpoints {
		conn, err := net.DialTimeout("tcp", endpoint, timeout)
		if err == nil {
			conn.Close()
			return true
		}
	}

	return false
}

func (ev *EnvironmentValidator) getRequiredPorts() map[int]string {
	ports := make(map[int]string)
	
	// Get port configurations from environment variables
	envPorts := map[string]string{
		"DATABASE_PORT":   "postgresql",
		"REDIS_PORT":     "redis", 
		"VAULT_PORT":     "vault",
		"AZURITE_PORT":   "azurite",
		"GRAFANA_PORT":   "grafana",
		"LOKI_PORT":      "loki",
		"DAPR_HTTP_PORT": "dapr-http",
		"DAPR_GRPC_PORT": "dapr-grpc",
		"HTTP_PORT":      "http",
		"HTTPS_PORT":     "https",
	}
	
	for envVar, serviceName := range envPorts {
		if portStr := os.Getenv(envVar); portStr != "" {
			if port := parseInt(portStr); port > 0 {
				ports[port] = serviceName
			}
		}
	}
	
	return ports
}

func (ev *EnvironmentValidator) isPortCritical(port int) bool {
	// Check if port is defined in environment configuration
	requiredPorts := ev.getRequiredPorts()
	_, exists := requiredPorts[port]
	return exists
}

func (ev *EnvironmentValidator) validateEnvVarValue(name, value string) bool {
	if value == "" {
		return false
	}

	switch name {
	case "DATABASE_URL":
		return ev.isValidDatabaseURL(value)
	case "REDIS_URL":
		return ev.isValidRedisURL(value)
	case "VAULT_URL":
		return ev.isValidVaultURL(value)
	default:
		return true
	}
}

func (ev *EnvironmentValidator) isValidDatabaseURL(url string) bool {
	return len(url) > 10 && (url[:10] == "postgresql" || url[:8] == "postgres")
}

func (ev *EnvironmentValidator) isValidRedisURL(url string) bool {
	return len(url) > 6 && url[:6] == "redis:"
}

func (ev *EnvironmentValidator) isValidVaultURL(url string) bool {
	return len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://")
}

func (ev *EnvironmentValidator) extractHostPort(url string) string {
	// Extract actual host:port from URL instead of hardcoding
	if url[:10] == "postgresql" || url[:8] == "postgres" {
		// Extract from postgresql://user:pass@host:port/db
		start := strings.Index(url, "@")
		if start != -1 {
			end := strings.Index(url[start+1:], "/")
			if end != -1 {
				return url[start+1 : start+1+end]
			}
		}
		return os.Getenv("DATABASE_HOST") + ":" + os.Getenv("DATABASE_PORT")
	}
	
	if len(url) > 6 && url[:6] == "redis:" {
		return os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
	}
	
	if len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://") {
		// Extract from http://host:port or https://host:port
		start := strings.Index(url, "://")
		if start != -1 {
			hostPort := url[start+3:]
			end := strings.Index(hostPort, "/")
			if end != -1 {
				return hostPort[:end]
			}
			return hostPort
		}
	}
	
	return os.Getenv("SERVICE_HOST") + ":" + os.Getenv("SERVICE_PORT")
}

func (ev *EnvironmentValidator) buildHealthEndpoint(baseURL, path string) string {
	return fmt.Sprintf("%s%s", baseURL, path)
}

func (ev *EnvironmentValidator) checkHTTPHealth(ctx context.Context, endpoint string) (HealthStatus, error) {
	client := &http.Client{
		Timeout: ev.getTimeout("health"),
	}

	resp, err := client.Get(endpoint)
	if err != nil {
		return HealthStatus{}, err
	}
	defer resp.Body.Close()

	return HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"status_code": resp.StatusCode,
		},
	}, nil
}

func (ev *EnvironmentValidator) getTimeout(component string) time.Duration {
	// Only use timeout from configuration, no fallbacks
	if timeout, exists := ev.config.Timeouts[component]; exists {
		return timeout
	}

	// Require timeout to be explicitly configured
	panic(fmt.Sprintf("timeout for component %s must be explicitly configured in ValidationConfig", component))
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