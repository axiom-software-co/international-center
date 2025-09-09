package health

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/containers"
)

type HealthServer struct {
	port         int
	server       *http.Server
	orchestrator *containers.ContainerOrchestrator
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Component string            `json:"component"`
	Healthy   bool              `json:"healthy"`
	Details   map[string]string `json:"details"`
	Timestamp time.Time         `json:"timestamp"`
}

type OverallHealthResponse struct {
	Status         string                     `json:"status"`
	Environment    string                     `json:"environment"`
	TotalServices  int                        `json:"total_services"`
	HealthyCount   int                        `json:"healthy_count"`
	UnhealthyCount int                        `json:"unhealthy_count"`
	Services       map[string]HealthResponse `json:"services"`
	Timestamp      time.Time                  `json:"timestamp"`
}

func NewHealthServer(port int, environment string) *HealthServer {
	orchestrator := containers.NewContainerOrchestrator(environment)
	orchestrator.SetupDevelopmentInfrastructure()
	
	return &HealthServer{
		port:         port,
		orchestrator: orchestrator,
	}
}

func (hs *HealthServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	
	// Individual component health checks
	mux.HandleFunc("/health/", hs.handleComponentHealth)
	
	// Overall health check
	mux.HandleFunc("/health", hs.handleOverallHealth)
	
	// Status endpoint
	mux.HandleFunc("/status", hs.handleStatus)
	
	hs.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", hs.port),
		Handler: mux,
	}
	
	log.Printf("Starting health check server on port %d", hs.port)
	
	go func() {
		if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health server error: %v", err)
		}
	}()
	
	// Wait a moment for server to start
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

func (hs *HealthServer) Stop(ctx context.Context) error {
	if hs.server != nil {
		return hs.server.Shutdown(ctx)
	}
	return nil
}

func (hs *HealthServer) handleComponentHealth(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/health/")
	component := strings.TrimSuffix(path, "/")
	
	log.Printf("Health check request for component: %s", component)
	
	response := hs.checkComponentHealth(component)
	
	w.Header().Set("Content-Type", "application/json")
	if response.Healthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	
	json.NewEncoder(w).Encode(response)
}

func (hs *HealthServer) handleOverallHealth(w http.ResponseWriter, r *http.Request) {
	response := hs.checkOverallHealth()
	
	w.Header().Set("Content-Type", "application/json")
	if response.Status == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	
	json.NewEncoder(w).Encode(response)
}

func (hs *HealthServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := hs.orchestrator.GetContainerStatus()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"container_status": status,
		"timestamp":        time.Now(),
	})
}

func (hs *HealthServer) checkComponentHealth(component string) HealthResponse {
	response := HealthResponse{
		Component: component,
		Details:   make(map[string]string),
		Timestamp: time.Now(),
	}
	
	// Map component names to container names
	containerName := hs.mapComponentToContainer(component)
	
	if containerName == "" {
		response.Status = "unknown"
		response.Healthy = false
		response.Details["error"] = fmt.Sprintf("Unknown component: %s", component)
		return response
	}
	
	// Check if container is running
	isRunning := hs.isContainerHealthy(containerName)
	
	if isRunning {
		response.Status = "healthy"
		response.Healthy = true
		response.Details["container"] = containerName
		response.Details["container_status"] = "running"
	} else {
		response.Status = "unhealthy"
		response.Healthy = false
		response.Details["container"] = containerName
		response.Details["container_status"] = "not_running"
		response.Details["error"] = fmt.Sprintf("Container %s is not running", containerName)
	}
	
	return response
}

func (hs *HealthServer) checkOverallHealth() OverallHealthResponse {
	components := []string{
		"database_connection_string",
		"storage_connection_string", 
		"vault_address",
		"rabbitmq_endpoint",
		"grafana_url",
		"website_url",
	}
	
	response := OverallHealthResponse{
		Environment: "development",
		Services:    make(map[string]HealthResponse),
		Timestamp:   time.Now(),
	}
	
	healthyCount := 0
	for _, component := range components {
		health := hs.checkComponentHealth(component)
		response.Services[component] = health
		
		if health.Healthy {
			healthyCount++
		}
	}
	
	response.TotalServices = len(components)
	response.HealthyCount = healthyCount
	response.UnhealthyCount = response.TotalServices - healthyCount
	
	if healthyCount == response.TotalServices {
		response.Status = "healthy"
	} else if healthyCount > 0 {
		response.Status = "partially_healthy"
	} else {
		response.Status = "unhealthy"
	}
	
	return response
}

func (hs *HealthServer) mapComponentToContainer(component string) string {
	// Map Pulumi output names to actual container names
	componentMap := map[string]string{
		"database_connection_string": "postgres",
		"storage_connection_string":  "azurite",
		"vault_address":             "vault",
		"rabbitmq_endpoint":         "rabbitmq",
		"grafana_url":               "grafana",
		"website_url":               "", // Website container doesn't exist yet
	}
	
	if containerName, exists := componentMap[component]; exists {
		return containerName
	}
	
	return ""
}

func (hs *HealthServer) isContainerHealthy(containerName string) bool {
	if containerName == "" {
		return false
	}
	
	// Check if container is running using podman
	cmd := exec.Command("podman", "ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	return strings.TrimSpace(string(output)) == containerName
}