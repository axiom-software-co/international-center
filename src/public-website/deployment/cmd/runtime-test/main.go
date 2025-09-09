package main

import (
	"context"
	"log"
	"time"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/internal/components/platform"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Simple runtime orchestrator test to debug container execution
func main() {
	log.Printf("Testing runtime orchestrator outside of Pulumi context")

	// Create runtime orchestrator
	orchestrator := platform.NewRuntimeOrchestrator("development")
	
	// Create test runtime execution args
	args := &platform.RuntimeExecutionArgs{
		Environment:           "development",
		InfrastructureOutputs: pulumi.Map{},
		PlatformOutputs:      pulumi.Map{},
		ServicesOutputs:      pulumi.Map{},
		ExecutionContext:     context.Background(),
		ExecutionTimeout:     5 * time.Minute,
	}

	// Execute runtime deployment
	ctx := context.Background()
	if err := orchestrator.ExecuteRuntimeDeployment(ctx, args); err != nil {
		log.Fatalf("Runtime orchestration failed: %v", err)
	}

	log.Printf("Runtime orchestration test completed successfully")
}