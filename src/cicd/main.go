package main

import (
	"github.com/axiom-software-co/international-center/src/cicd/shared"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "international-center-cicd")
		environment := cfg.Require("environment")
		
		ctx.Log.Info("Starting International Center infrastructure deployment", nil)
		ctx.Log.Info("Environment: "+environment, nil)
		
		// Initialize deployment coordinator to prevent process proliferation
		coordinator := shared.NewDeploymentCoordinator()
		
		// Cleanup any orphaned processes from previous runs
		if err := coordinator.CleanupOrphanedProcesses(); err != nil {
			ctx.Log.Warn("Warning during process cleanup: "+err.Error(), nil)
		}
		
		// Start coordinated deployment
		deployment, err := coordinator.StartDeployment(environment, ctx, cfg)
		if err != nil {
			ctx.Log.Error("Failed to start coordinated deployment: "+err.Error(), nil)
			return err
		}
		
		// Ensure deployment is completed even if process fails
		defer func() {
			if completeErr := coordinator.CompleteDeployment(deployment); completeErr != nil {
				ctx.Log.Error("Failed to complete deployment coordination: "+completeErr.Error(), nil)
			}
		}()
		
		// Create factory and deployment strategy
		factory := shared.NewEnvironmentFactory()
		strategy, err := factory.CreateDeploymentStrategy(environment, ctx, cfg)
		if err != nil {
			ctx.Log.Error("Failed to create deployment strategy: "+err.Error(), nil)
			return err
		}
		
		// Execute deployment using strategy pattern under coordination
		outputs, err := strategy.Deploy(ctx, cfg)
		if err != nil {
			ctx.Log.Error("Deployment failed: "+err.Error(), nil)
			return err
		}
		
		// Export outputs
		for key, value := range outputs {
			ctx.Export(key, pulumi.ToOutput(value))
		}
		
		ctx.Log.Info(environment+" infrastructure deployment completed successfully", nil)
		return nil
	})
}

