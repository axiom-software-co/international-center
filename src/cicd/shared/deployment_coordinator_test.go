package shared

import (
	"sync"
	"testing"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeploymentCoordinator validates the deployment process coordination pattern
func TestDeploymentCoordinator(t *testing.T) {
	t.Run("EnforceSingletonDeploymentProcess", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-coordinator-singleton-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			coordinator := NewDeploymentCoordinator()
			
			// First deployment should succeed
			deployment1, err := coordinator.StartDeployment("development", ctx, cfg)
			require.NoError(t, err, "First deployment should start successfully")
			assert.NotNil(t, deployment1, "First deployment should return deployment instance")
			
			// Second concurrent deployment should be rejected
			deployment2, err := coordinator.StartDeployment("development", ctx, cfg)
			require.Error(t, err, "Second concurrent deployment should be rejected")
			assert.Nil(t, deployment2, "Second deployment should not return instance")
			assert.Contains(t, err.Error(), "deployment already in progress", "Error should indicate deployment in progress")
			
			// Complete first deployment
			err = coordinator.CompleteDeployment(deployment1)
			require.NoError(t, err, "First deployment should complete successfully")
			
			// Now second deployment should succeed
			deployment3, err := coordinator.StartDeployment("development", ctx, cfg)
			require.NoError(t, err, "Third deployment should start after first completes")
			assert.NotNil(t, deployment3, "Third deployment should return deployment instance")
			
			// Cleanup
			err = coordinator.CompleteDeployment(deployment3)
			require.NoError(t, err, "Third deployment should complete successfully")
			
			return nil
		})
	})
	
	t.Run("PreventProcessProliferation", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-coordinator-proliferation-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			coordinator := NewDeploymentCoordinator()
			
			// Simulate multiple concurrent requests (this causes current proliferation issue)
			var wg sync.WaitGroup
			successCount := 0
			rejectionCount := 0
			var mu sync.Mutex
			
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func(deploymentID int) {
					defer wg.Done()
					
					deployment, err := coordinator.StartDeployment("development", ctx, cfg)
					mu.Lock()
					defer mu.Unlock()
					
					if err != nil {
						rejectionCount++
						assert.Contains(t, err.Error(), "deployment already in progress", "Rejection should indicate deployment in progress")
					} else {
						successCount++
						assert.NotNil(t, deployment, "Successful deployment should return instance")
						
						// Simulate deployment work
						time.Sleep(50 * time.Millisecond)
						
						// Complete deployment
						coordinator.CompleteDeployment(deployment)
					}
				}(i)
			}
			
			wg.Wait()
			
			// Only one deployment should succeed, others should be rejected
			assert.Equal(t, 1, successCount, "Only one deployment should succeed")
			assert.Equal(t, 4, rejectionCount, "Four deployments should be rejected")
			
			return nil
		})
	})
	
	t.Run("HandleDeploymentTimeout", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-coordinator-timeout-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			coordinator := NewDeploymentCoordinator()
			
			// Start deployment with timeout
			deployment, err := coordinator.StartDeploymentWithTimeout("development", ctx, cfg, 100*time.Millisecond)
			require.NoError(t, err, "Deployment should start successfully")
			
			// Wait longer than timeout
			time.Sleep(150 * time.Millisecond)
			
			// Coordinator should automatically clean up timed out deployment
			assert.True(t, coordinator.IsDeploymentTimedOut(deployment), "Deployment should be marked as timed out")
			
			// New deployment should now be allowed
			deployment2, err := coordinator.StartDeployment("development", ctx, cfg)
			require.NoError(t, err, "New deployment should start after timeout cleanup")
			assert.NotNil(t, deployment2, "New deployment should return instance")
			
			// Cleanup
			err = coordinator.CompleteDeployment(deployment2)
			require.NoError(t, err, "Second deployment should complete successfully")
			
			return nil
		})
	})
}

// TestDeploymentProcessCleanup validates proper cleanup of deployment processes
func TestDeploymentProcessCleanup(t *testing.T) {
	t.Run("CleanupOrphanedProcesses", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-cleanup-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			coordinator := NewDeploymentCoordinator()
			
			// Start deployment
			deployment, err := coordinator.StartDeployment("development", ctx, cfg)
			require.NoError(t, err, "Deployment should start successfully")
			
			// Simulate process termination without proper cleanup (current issue)
			processID := coordinator.GetProcessID(deployment)
			assert.NotEmpty(t, processID, "Deployment should have associated process ID")
			
			// Force cleanup of orphaned process
			err = coordinator.CleanupOrphanedProcesses()
			require.NoError(t, err, "Cleanup should succeed")
			
			// Verify process is cleaned up
			assert.False(t, coordinator.IsProcessRunning(processID), "Process should no longer be running")
			
			// New deployment should be allowed after cleanup
			deployment2, err := coordinator.StartDeployment("development", ctx, cfg)
			require.NoError(t, err, "New deployment should start after cleanup")
			assert.NotNil(t, deployment2, "New deployment should return instance")
			
			// Cleanup properly
			err = coordinator.CompleteDeployment(deployment2)
			require.NoError(t, err, "Second deployment should complete successfully")
			
			return nil
		})
	})
	
	t.Run("PreventResourceLeaks", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-resource-leaks-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			coordinator := NewDeploymentCoordinator()
			
			// Get baseline resource count
			initialProcessCount := coordinator.GetActiveProcessCount()
			
			// Start and complete multiple deployments
			for i := 0; i < 3; i++ {
				deployment, err := coordinator.StartDeployment("development", ctx, cfg)
				require.NoError(t, err, "Deployment %d should start successfully", i)
				
				// Complete deployment
				err = coordinator.CompleteDeployment(deployment)
				require.NoError(t, err, "Deployment %d should complete successfully", i)
			}
			
			// Resource count should return to baseline
			finalProcessCount := coordinator.GetActiveProcessCount()
			assert.Equal(t, initialProcessCount, finalProcessCount, "Process count should return to baseline after deployments")
			
			return nil
		})
	})
}

// TestDeploymentConcurrency validates concurrent access patterns
func TestDeploymentConcurrency(t *testing.T) {
	t.Run("DetectCleanupDeadlockCondition", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-concurrency-deadlock-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			coordinator := NewDeploymentCoordinator()
			
			// Create an active deployment to trigger cleanup condition
			deployment, err := coordinator.StartDeployment("development", ctx, cfg)
			require.NoError(t, err, "Deployment should start successfully")
			
			// Test that cleanup completes without deadlock - the key is it doesn't hang
			done := make(chan error, 1)
			go func() {
				done <- coordinator.CleanupOrphanedProcesses()
			}()
			
			select {
			case err := <-done:
				// The critical test - cleanup completed without deadlock
				assert.NoError(t, err, "Cleanup should complete without deadlock")
				
				// If cleanup cleared the deployment, that's acceptable behavior
				// The key test was that it didn't deadlock
				if coordinator.activeDeployment != nil {
					// Only try to complete if deployment still active
					err = coordinator.CompleteDeployment(deployment)
					require.NoError(t, err, "Deployment should complete successfully")
				}
			case <-time.After(2 * time.Second):
				t.Fatal("CleanupOrphanedProcesses deadlocked - mutex ordering issue detected")
			}
			
			return nil
		})
	})
}

// TestDeploymentStateLock validates Pulumi state lock handling
func TestDeploymentStateLock(t *testing.T) {
	t.Run("HandleStateLockContention", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-state-lock-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			coordinator := NewDeploymentCoordinator()
			
			// Start deployment that holds state lock
			deployment1, err := coordinator.StartDeployment("development", ctx, cfg)
			require.NoError(t, err, "First deployment should start successfully")
			
			// Verify state lock is held
			assert.True(t, coordinator.IsStateLocked("development"), "State should be locked during deployment")
			
			// Attempt concurrent deployment should detect active deployment
			_, err = coordinator.StartDeployment("development", ctx, cfg)
			require.Error(t, err, "Second deployment should be rejected due to active deployment")
			assert.Contains(t, err.Error(), "deployment already in progress", "Error should indicate deployment in progress")
			
			// Complete first deployment
			err = coordinator.CompleteDeployment(deployment1)
			require.NoError(t, err, "First deployment should complete successfully")
			
			// State lock should be released
			assert.False(t, coordinator.IsStateLocked("development"), "State should be unlocked after deployment completion")
			
			// New deployment should succeed
			deployment3, err := coordinator.StartDeployment("development", ctx, cfg)
			require.NoError(t, err, "Third deployment should start after lock release")
			
			// Cleanup
			err = coordinator.CompleteDeployment(deployment3)
			require.NoError(t, err, "Third deployment should complete successfully")
			
			return nil
		})
	})
	
	t.Run("RecoverFromStaleLocks", func(t *testing.T) {
		framework := NewContractTestingFramework("international-center", "deployment-stale-lock-test")
		
		framework.RunComponentContractTest(t, "development", func(t *testing.T, ctx *pulumi.Context, cfg *config.Config, env string) error {
			coordinator := NewDeploymentCoordinator()
			
			// Simulate stale lock (process died without cleanup)
			err := coordinator.CreateStaleLock("development")
			require.NoError(t, err, "Should be able to create stale lock for testing")
			
			// Verify lock exists
			assert.True(t, coordinator.IsStateLocked("development"), "Stale lock should be detected")
			
			// Attempt deployment should detect and recover from stale lock
			deployment, err := coordinator.StartDeploymentWithRecovery("development", ctx, cfg)
			require.NoError(t, err, "Deployment should succeed after stale lock recovery")
			assert.NotNil(t, deployment, "Deployment should return instance after recovery")
			
			// State should be locked by new deployment
			assert.True(t, coordinator.IsStateLocked("development"), "State should be locked by new deployment")
			
			// Complete deployment
			err = coordinator.CompleteDeployment(deployment)
			require.NoError(t, err, "Deployment should complete successfully")
			
			// State should be unlocked
			assert.False(t, coordinator.IsStateLocked("development"), "State should be unlocked after completion")
			
			return nil
		})
	})
}