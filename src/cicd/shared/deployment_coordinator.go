package shared

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// DeploymentCoordinator manages singleton deployment processes
type DeploymentCoordinator struct {
	mutex             sync.RWMutex
	activeDeployment  *Deployment
	stateLocks        map[string]time.Time
	processRegistry   map[string]*os.Process
}

// Deployment represents an active deployment process
type Deployment struct {
	ID          string
	Environment string
	StartTime   time.Time
	ProcessID   string
	Context     *pulumi.Context
	Config      *config.Config
	Timeout     time.Duration
}

// NewDeploymentCoordinator creates a new deployment coordinator
func NewDeploymentCoordinator() *DeploymentCoordinator {
	return &DeploymentCoordinator{
		stateLocks:      make(map[string]time.Time),
		processRegistry: make(map[string]*os.Process),
	}
}

// StartDeployment starts a new deployment if none is active
func (dc *DeploymentCoordinator) StartDeployment(environment string, ctx *pulumi.Context, cfg *config.Config) (*Deployment, error) {
	return dc.StartDeploymentWithTimeout(environment, ctx, cfg, 30*time.Minute)
}

// StartDeploymentWithTimeout starts a new deployment with specified timeout
// Mutex usage: Acquires write lock for the entire operation to ensure atomic deployment creation
func (dc *DeploymentCoordinator) StartDeploymentWithTimeout(environment string, ctx *pulumi.Context, cfg *config.Config, timeout time.Duration) (*Deployment, error) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	
	// Check if deployment already in progress - use internal timeout check to avoid nested locking
	if dc.activeDeployment != nil {
		if !dc.isDeploymentTimedOutLocked(dc.activeDeployment) {
			return nil, fmt.Errorf("deployment already in progress for environment: %s", dc.activeDeployment.Environment)
		} else {
			// Clear timed-out deployment atomically
			timedOutEnv := dc.activeDeployment.Environment
			dc.activeDeployment = nil
			delete(dc.stateLocks, timedOutEnv)
		}
	}
	
	// Check for state lock (already under write lock)
	if dc.isStateLocked(environment) {
		return nil, fmt.Errorf("state locked for environment: %s", environment)
	}
	
	// Create new deployment
	deployment := &Deployment{
		ID:          fmt.Sprintf("deploy-%d", time.Now().UnixNano()),
		Environment: environment,
		StartTime:   time.Now(),
		ProcessID:   fmt.Sprintf("proc-%d", os.Getpid()),
		Context:     ctx,
		Config:      cfg,
		Timeout:     timeout,
	}
	
	// Set state lock
	dc.stateLocks[environment] = time.Now()
	
	// Register process
	if process, err := os.FindProcess(os.Getpid()); err == nil {
		dc.processRegistry[deployment.ProcessID] = process
	}
	
	dc.activeDeployment = deployment
	
	return deployment, nil
}

// StartDeploymentWithRecovery starts deployment with stale lock recovery
// Mutex usage: Acquires write lock briefly for stale lock cleanup, then uses normal deployment flow
func (dc *DeploymentCoordinator) StartDeploymentWithRecovery(environment string, ctx *pulumi.Context, cfg *config.Config) (*Deployment, error) {
	// Minimize lock duration - only hold for stale lock check and cleanup
	dc.mutex.Lock()
	if dc.isStateLocked(environment) {
		lockTime := dc.stateLocks[environment]
		if time.Since(lockTime) > 5*time.Minute {
			// Consider lock stale and remove it atomically
			delete(dc.stateLocks, environment)
		}
	}
	dc.mutex.Unlock()
	
	// Now attempt normal deployment (which will acquire its own lock)
	return dc.StartDeployment(environment, ctx, cfg)
}

// CompleteDeployment marks a deployment as completed and releases locks
func (dc *DeploymentCoordinator) CompleteDeployment(deployment *Deployment) error {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	
	if dc.activeDeployment != deployment {
		return fmt.Errorf("deployment not active")
	}
	
	// Release state lock
	delete(dc.stateLocks, deployment.Environment)
	
	// Remove process from registry
	delete(dc.processRegistry, deployment.ProcessID)
	
	// Clear active deployment
	dc.activeDeployment = nil
	
	return nil
}

// IsDeploymentTimedOut checks if deployment has timed out
// Mutex usage: Thread-safe read-only operation, no locks needed for time comparison
func (dc *DeploymentCoordinator) IsDeploymentTimedOut(deployment *Deployment) bool {
	return dc.isDeploymentTimedOutLocked(deployment)
}

// isDeploymentTimedOutLocked internal timeout check without lock acquisition
// Must be called with mutex already held or for lock-free operations
func (dc *DeploymentCoordinator) isDeploymentTimedOutLocked(deployment *Deployment) bool {
	return time.Since(deployment.StartTime) > deployment.Timeout
}

// GetProcessID returns the process ID for a deployment
// Mutex usage: No locks needed - read-only access to deployment field
func (dc *DeploymentCoordinator) GetProcessID(deployment *Deployment) string {
	return deployment.ProcessID
}

// IsProcessRunning checks if a process is still running
// Mutex usage: Acquires read lock to safely access process registry
func (dc *DeploymentCoordinator) IsProcessRunning(processID string) bool {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()
	
	return dc.isProcessRunningLocked(processID)
}

// isProcessRunningLocked checks if a process is running without acquiring locks
// Must be called with mutex already held
func (dc *DeploymentCoordinator) isProcessRunningLocked(processID string) bool {
	process, exists := dc.processRegistry[processID]
	if !exists {
		return false
	}
	
	// Try to send signal 0 to check if process exists
	err := process.Signal(os.Signal(nil))
	return err == nil
}

// CleanupOrphanedProcesses removes processes that are no longer running
// Mutex usage: Acquires write lock for the entire cleanup operation to ensure consistency
// This method resolves the previous deadlock by using internal lock-free methods
func (dc *DeploymentCoordinator) CleanupOrphanedProcesses() error {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	
	// Cleanup orphaned processes from registry
	for processID, process := range dc.processRegistry {
		err := process.Signal(os.Signal(nil))
		if err != nil {
			// Process no longer exists, remove from registry
			delete(dc.processRegistry, processID)
		}
	}
	
	// Check active deployment status using lock-free internal methods
	if dc.activeDeployment != nil {
		if dc.isDeploymentTimedOutLocked(dc.activeDeployment) {
			// Clear timed-out deployment atomically
			delete(dc.stateLocks, dc.activeDeployment.Environment)
			dc.activeDeployment = nil
		} else if !dc.isProcessRunningLocked(dc.activeDeployment.ProcessID) {
			// Process terminated without proper cleanup, clear deployment atomically
			delete(dc.stateLocks, dc.activeDeployment.Environment)
			dc.activeDeployment = nil
		}
	}
	
	return nil
}

// GetActiveProcessCount returns number of active processes
// Mutex usage: Acquires read lock for thread-safe registry access
func (dc *DeploymentCoordinator) GetActiveProcessCount() int {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()
	
	return len(dc.processRegistry)
}

// IsStateLocked checks if environment state is locked
// Mutex usage: Acquires read lock for thread-safe state lock access
func (dc *DeploymentCoordinator) IsStateLocked(environment string) bool {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()
	
	return dc.isStateLocked(environment)
}

// isStateLocked internal method to check state lock (assumes lock held)
func (dc *DeploymentCoordinator) isStateLocked(environment string) bool {
	_, exists := dc.stateLocks[environment]
	return exists
}

// CreateStaleLock creates a stale lock for testing purposes
func (dc *DeploymentCoordinator) CreateStaleLock(environment string) error {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	
	dc.stateLocks[environment] = time.Now().Add(-10 * time.Minute) // 10 minutes ago
	return nil
}

// KillBackgroundProcesses attempts to cleanup any background pulumi processes
// Mutex usage: Acquires write lock only for registry cleanup, minimizing lock duration
func (dc *DeploymentCoordinator) KillBackgroundProcesses() error {
	// Execute external process cleanup without holding locks (no shared state access)
	cmd := exec.Command("pkill", "-f", "pulumi")
	cmd.Run() // Ignore errors as processes may not exist
	
	cmd = exec.Command("pkill", "-f", "pulumi-program")
	cmd.Run() // Ignore errors as processes may not exist
	
	// Only acquire lock for internal state cleanup
	dc.mutex.Lock()
	dc.processRegistry = make(map[string]*os.Process)
	dc.activeDeployment = nil
	// Clear all state locks as we're doing a full reset
	dc.stateLocks = make(map[string]time.Time)
	dc.mutex.Unlock()
	
	return nil
}

// GetPulumiLockStatus checks if pulumi state is locked
func (dc *DeploymentCoordinator) GetPulumiLockStatus() (bool, error) {
	// Check if we can acquire pulumi lock by running a quick command
	cmd := exec.Command("pulumi", "stack", "ls", "--non-interactive")
	cmd.Env = append(os.Environ(), "PULUMI_CONFIG_PASSPHRASE=development")
	
	err := cmd.Run()
	return err != nil, err
}

// WaitForPulumiLock waits for pulumi state lock to be released
func (dc *DeploymentCoordinator) WaitForPulumiLock(timeout time.Duration) error {
	start := time.Now()
	
	for time.Since(start) < timeout {
		locked, err := dc.GetPulumiLockStatus()
		if err == nil && !locked {
			return nil
		}
		
		time.Sleep(5 * time.Second)
	}
	
	return fmt.Errorf("timeout waiting for pulumi lock")
}