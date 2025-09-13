package dapr

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/google/uuid"
)

// TransactionManager handles distributed transactions and data consistency across services
type TransactionManager struct {
	stateStore      *StateStore
	pubsub          *PubSub
	client          *Client
	sagaStore       map[string]*SagaExecution
	sagaMutex       sync.RWMutex
	consistencyLevel ConsistencyLevel
}

// ConsistencyLevel defines the level of consistency required
type ConsistencyLevel string

const (
	ConsistencyEventual ConsistencyLevel = "eventual"
	ConsistencyStrong   ConsistencyLevel = "strong"
	ConsistencySession  ConsistencyLevel = "session"
)

// TransactionOptions configures transaction behavior
type TransactionOptions struct {
	ConsistencyLevel ConsistencyLevel      `json:"consistency_level"`
	Timeout         time.Duration         `json:"timeout"`
	RetryPolicy     *RetryPolicy          `json:"retry_policy,omitempty"`
	CompensationTimeout time.Duration     `json:"compensation_timeout"`
}

// RetryPolicy defines retry behavior for failed operations
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// SagaDefinition defines a saga transaction pattern
type SagaDefinition struct {
	SagaID      string                `json:"saga_id"`
	Name        string                `json:"name"`
	Steps       []*SagaStep          `json:"steps"`
	Timeout     time.Duration        `json:"timeout"`
	CreatedAt   time.Time            `json:"created_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SagaStep represents a step in a saga transaction
type SagaStep struct {
	StepID          string                 `json:"step_id"`
	Name            string                 `json:"name"`
	ServiceName     string                 `json:"service_name"`
	Operation       string                 `json:"operation"`
	Data            map[string]interface{} `json:"data"`
	CompensationOp  string                 `json:"compensation_operation,omitempty"`
	CompensationData map[string]interface{} `json:"compensation_data,omitempty"`
	Timeout         time.Duration          `json:"timeout"`
	RetryPolicy     *RetryPolicy           `json:"retry_policy,omitempty"`
}

// SagaExecution tracks the execution state of a saga
type SagaExecution struct {
	SagaID          string                 `json:"saga_id"`
	Definition      *SagaDefinition        `json:"definition"`
	Status          SagaStatus             `json:"status"`
	CurrentStep     int                    `json:"current_step"`
	CompletedSteps  []string               `json:"completed_steps"`
	FailedStep      *SagaStepResult        `json:"failed_step,omitempty"`
	StartedAt       time.Time              `json:"started_at"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	Compensation    *CompensationExecution `json:"compensation,omitempty"`
}

// SagaStatus represents the status of a saga execution
type SagaStatus string

const (
	SagaStatusPending      SagaStatus = "pending"
	SagaStatusExecuting    SagaStatus = "executing"
	SagaStatusCompleted    SagaStatus = "completed"
	SagaStatusFailed       SagaStatus = "failed"
	SagaStatusCompensating SagaStatus = "compensating"
	SagaStatusCompensated  SagaStatus = "compensated"
)

// SagaStepResult represents the result of executing a saga step
type SagaStepResult struct {
	StepID      string                 `json:"step_id"`
	Status      string                 `json:"status"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	ExecutedAt  time.Time              `json:"executed_at"`
	Duration    time.Duration          `json:"duration"`
}

// CompensationExecution tracks compensation action execution
type CompensationExecution struct {
	StartedAt       time.Time            `json:"started_at"`
	CompletedAt     *time.Time           `json:"completed_at,omitempty"`
	CompensatedSteps []string            `json:"compensated_steps"`
	Status          string               `json:"status"`
	ErrorMessage    string               `json:"error_message,omitempty"`
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(client *Client) *TransactionManager {
	return &TransactionManager{
		stateStore:       NewStateStore(client),
		pubsub:           NewPubSub(client),
		client:           client,
		sagaStore:        make(map[string]*SagaExecution),
		sagaMutex:        sync.RWMutex{},
		consistencyLevel: ConsistencyEventual,
	}
}

// ExecuteDistributedTransaction executes a distributed transaction using the saga pattern
func (tm *TransactionManager) ExecuteDistributedTransaction(ctx context.Context, saga *SagaDefinition, options *TransactionOptions) (*SagaExecution, error) {
	if saga == nil {
		return nil, fmt.Errorf("saga definition cannot be nil")
	}

	if options == nil {
		options = &TransactionOptions{
			ConsistencyLevel:    ConsistencyEventual,
			Timeout:            30 * time.Second,
			CompensationTimeout: 60 * time.Second,
		}
	}

	// Create saga execution
	execution := &SagaExecution{
		SagaID:         saga.SagaID,
		Definition:     saga,
		Status:         SagaStatusPending,
		CurrentStep:    0,
		CompletedSteps: []string{},
		StartedAt:      time.Now(),
	}

	// Store saga execution state
	if err := tm.storeSagaExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to store saga execution: %w", err)
	}

	// Execute saga asynchronously
	go tm.executeSagaAsync(context.Background(), execution, options)

	return execution, nil
}

// executeSagaAsync executes the saga steps asynchronously
func (tm *TransactionManager) executeSagaAsync(ctx context.Context, execution *SagaExecution, options *TransactionOptions) {
	execution.Status = SagaStatusExecuting
	tm.updateSagaExecution(ctx, execution)

	// Execute each step
	for i, step := range execution.Definition.Steps {
		execution.CurrentStep = i

		stepResult, err := tm.executeStep(ctx, step, options)
		if err != nil {
			// Step failed, initiate compensation
			execution.Status = SagaStatusFailed
			execution.FailedStep = stepResult
			execution.ErrorMessage = err.Error()
			tm.updateSagaExecution(ctx, execution)

			// Start compensation
			tm.compensateSaga(ctx, execution, options)
			return
		}

		// Step succeeded
		execution.CompletedSteps = append(execution.CompletedSteps, step.StepID)
		tm.updateSagaExecution(ctx, execution)
	}

	// All steps completed successfully
	execution.Status = SagaStatusCompleted
	now := time.Now()
	execution.CompletedAt = &now
	tm.updateSagaExecution(ctx, execution)

	// Publish saga completion event
	tm.publishSagaEvent(ctx, execution, "saga.completed")
}

// executeStep executes a single saga step
func (tm *TransactionManager) executeStep(ctx context.Context, step *SagaStep, options *TransactionOptions) (*SagaStepResult, error) {
	startTime := time.Now()
	
	result := &SagaStepResult{
		StepID:     step.StepID,
		ExecutedAt: startTime,
	}

	// Create step timeout context
	stepCtx := ctx
	if step.Timeout > 0 {
		var cancel context.CancelFunc
		stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
		defer cancel()
	}

	// Execute step with retry logic
	var err error
	retryPolicy := step.RetryPolicy
	if retryPolicy == nil {
		retryPolicy = &RetryPolicy{
			MaxRetries:    3,
			InitialDelay:  100 * time.Millisecond,
			MaxDelay:      5 * time.Second,
			BackoffFactor: 2.0,
		}
	}

	for attempt := 0; attempt <= retryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(float64(retryPolicy.InitialDelay) * 
				float64(attempt) * retryPolicy.BackoffFactor)
			if delay > retryPolicy.MaxDelay {
				delay = retryPolicy.MaxDelay
			}
			time.Sleep(delay)
		}

		err = tm.executeStepOperation(stepCtx, step)
		if err == nil {
			result.Status = "completed"
			result.Duration = time.Since(startTime)
			return result, nil
		}

		// Don't retry on context cancellation
		if stepCtx.Err() != nil {
			break
		}
	}

	result.Status = "failed"
	result.Error = err.Error()
	result.Duration = time.Since(startTime)
	return result, err
}

// executeStepOperation executes the actual step operation
func (tm *TransactionManager) executeStepOperation(ctx context.Context, step *SagaStep) error {
	// Publish step execution event to the target service
	event := &CrossServiceEvent{
		EventID:       uuid.New().String(),
		EventType:     "saga.step.execute",
		SourceService: tm.client.GetAppID(),
		TargetService: step.ServiceName,
		EntityType:    "saga_step",
		EntityID:      step.StepID,
		OperationType: step.Operation,
		Payload: map[string]interface{}{
			"step_id":   step.StepID,
			"operation": step.Operation,
			"data":      step.Data,
			"timeout":   step.Timeout.String(),
		},
		Timestamp: time.Now(),
	}

	return tm.pubsub.PublishCrossServiceEvent(ctx, event)
}

// compensateSaga executes compensation actions for failed saga
func (tm *TransactionManager) compensateSaga(ctx context.Context, execution *SagaExecution, options *TransactionOptions) {
	execution.Status = SagaStatusCompensating
	execution.Compensation = &CompensationExecution{
		StartedAt:        time.Now(),
		CompensatedSteps: []string{},
		Status:          "running",
	}

	// Execute compensation in reverse order
	for i := len(execution.CompletedSteps) - 1; i >= 0; i-- {
		stepID := execution.CompletedSteps[i]
		step := tm.findStepByID(execution.Definition, stepID)
		if step == nil || step.CompensationOp == "" {
			continue
		}

		if err := tm.executeCompensationStep(ctx, step, options); err != nil {
			execution.Compensation.Status = "failed"
			execution.Compensation.ErrorMessage = err.Error()
			tm.updateSagaExecution(ctx, execution)
			return
		}

		execution.Compensation.CompensatedSteps = append(execution.Compensation.CompensatedSteps, stepID)
	}

	// Compensation completed
	execution.Status = SagaStatusCompensated
	execution.Compensation.Status = "completed"
	now := time.Now()
	execution.Compensation.CompletedAt = &now
	tm.updateSagaExecution(ctx, execution)

	// Publish saga compensation event
	tm.publishSagaEvent(ctx, execution, "saga.compensated")
}

// executeCompensationStep executes a compensation step
func (tm *TransactionManager) executeCompensationStep(ctx context.Context, step *SagaStep, options *TransactionOptions) error {
	// Create compensation timeout context
	compensationCtx := ctx
	if options.CompensationTimeout > 0 {
		var cancel context.CancelFunc
		compensationCtx, cancel = context.WithTimeout(ctx, options.CompensationTimeout)
		defer cancel()
	}

	// Publish compensation event to the target service
	event := &CrossServiceEvent{
		EventID:       uuid.New().String(),
		EventType:     "saga.step.compensate",
		SourceService: tm.client.GetAppID(),
		TargetService: step.ServiceName,
		EntityType:    "saga_step",
		EntityID:      step.StepID,
		OperationType: step.CompensationOp,
		Payload: map[string]interface{}{
			"step_id":           step.StepID,
			"operation":         step.CompensationOp,
			"compensation_data": step.CompensationData,
			"original_data":     step.Data,
		},
		Timestamp: time.Now(),
	}

	return tm.pubsub.PublishCrossServiceEvent(compensationCtx, event)
}

// HandleEventualConsistency manages eventual consistency through event sourcing
func (tm *TransactionManager) HandleEventualConsistency(ctx context.Context, entityType, entityID string, version int, updateFunc func() error) error {
	// Implement optimistic locking with version checking
	lockKey := fmt.Sprintf("lock:%s:%s", entityType, entityID)
	versionKey := fmt.Sprintf("version:%s:%s", entityType, entityID)

	// Check current version
	currentVersionStr, err := tm.stateStore.Get(ctx, versionKey, nil)
	if err != nil && !domain.IsNotFoundError(err) {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	var currentVersion int
	if currentVersionStr != nil {
		if v, ok := currentVersionStr.(float64); ok {
			currentVersion = int(v)
		}
	}

	// Version conflict check
	if currentVersion != version {
		return domain.NewConflictError(fmt.Sprintf("version conflict for %s %s: expected %d, got %d", 
			entityType, entityID, version, currentVersion))
	}

	// Acquire lock
	lockData := map[string]interface{}{
		"locked_at": time.Now(),
		"locked_by": tm.client.GetAppID(),
		"ttl":       30, // 30 seconds TTL
	}

	err = tm.stateStore.Save(ctx, lockKey, lockData, &StateOptions{
		Concurrency: 1, // Optimistic concurrency
	})
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	// Ensure lock is released
	defer func() {
		tm.stateStore.Delete(ctx, lockKey, nil)
	}()

	// Execute update
	if err := updateFunc(); err != nil {
		return err
	}

	// Update version
	newVersion := currentVersion + 1
	err = tm.stateStore.Save(ctx, versionKey, newVersion, nil)
	if err != nil {
		return fmt.Errorf("failed to update version: %w", err)
	}

	return nil
}

// ValidateDataConsistency validates data consistency across services
func (tm *TransactionManager) ValidateDataConsistency(ctx context.Context, validations []ConsistencyValidation) error {
	var errors []error

	for _, validation := range validations {
		if err := tm.validateSingleConsistency(ctx, validation); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("data consistency validation failed: %v", errors)
	}

	return nil
}

// ConsistencyValidation defines a consistency check
type ConsistencyValidation struct {
	EntityType   string                 `json:"entity_type"`
	EntityID     string                 `json:"entity_id"`
	ServiceName  string                 `json:"service_name"`
	ExpectedData map[string]interface{} `json:"expected_data"`
	Tolerance    time.Duration          `json:"tolerance"`
}

// validateSingleConsistency validates a single consistency check
func (tm *TransactionManager) validateSingleConsistency(ctx context.Context, validation ConsistencyValidation) error {
	// Get current data from the service
	event := &CrossServiceEvent{
		EventID:       uuid.New().String(),
		EventType:     "consistency.validate",
		SourceService: tm.client.GetAppID(),
		TargetService: validation.ServiceName,
		EntityType:    validation.EntityType,
		EntityID:      validation.EntityID,
		OperationType: "VALIDATE",
		Payload: map[string]interface{}{
			"expected_data": validation.ExpectedData,
			"tolerance":     validation.Tolerance.String(),
		},
		Timestamp: time.Now(),
	}

	return tm.pubsub.PublishCrossServiceEvent(ctx, event)
}

// Utility functions

func (tm *TransactionManager) storeSagaExecution(ctx context.Context, execution *SagaExecution) error {
	key := fmt.Sprintf("saga:execution:%s", execution.SagaID)
	return tm.stateStore.Save(ctx, key, execution, nil)
}

func (tm *TransactionManager) updateSagaExecution(ctx context.Context, execution *SagaExecution) {
	tm.sagaMutex.Lock()
	tm.sagaStore[execution.SagaID] = execution
	tm.sagaMutex.Unlock()

	tm.storeSagaExecution(ctx, execution)
}

func (tm *TransactionManager) findStepByID(saga *SagaDefinition, stepID string) *SagaStep {
	for _, step := range saga.Steps {
		if step.StepID == stepID {
			return step
		}
	}
	return nil
}

func (tm *TransactionManager) publishSagaEvent(ctx context.Context, execution *SagaExecution, eventType string) {
	event := &CrossServiceEvent{
		EventID:       uuid.New().String(),
		EventType:     eventType,
		SourceService: tm.client.GetAppID(),
		EntityType:    "saga",
		EntityID:      execution.SagaID,
		OperationType: "NOTIFY",
		Payload: map[string]interface{}{
			"saga_id": execution.SagaID,
			"status":  execution.Status,
		},
		Timestamp: time.Now(),
	}

	tm.pubsub.PublishCrossServiceEvent(ctx, event)
}

// GetSagaExecution retrieves a saga execution by ID
func (tm *TransactionManager) GetSagaExecution(ctx context.Context, sagaID string) (*SagaExecution, error) {
	tm.sagaMutex.RLock()
	execution, exists := tm.sagaStore[sagaID]
	tm.sagaMutex.RUnlock()

	if exists {
		return execution, nil
	}

	// Try to load from state store
	key := fmt.Sprintf("saga:execution:%s", sagaID)
	var storedExecution SagaExecution
	found, err := tm.stateStore.Get(ctx, key, &storedExecution)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, domain.NewNotFoundError("saga execution", sagaID)
	}

	return &storedExecution, nil
}

// GetMetrics returns current transaction manager metrics
func (tm *TransactionManager) GetMetrics() *TransactionMetrics {
	if tm.metrics == nil {
		return nil
	}

	tm.metricsMutex.RLock()
	defer tm.metricsMutex.RUnlock()

	// Create a copy to avoid concurrent access issues
	metricsCopy := &TransactionMetrics{
		SagasStarted:         tm.metrics.SagasStarted,
		SagasCompleted:       tm.metrics.SagasCompleted,
		SagasFailed:         tm.metrics.SagasFailed,
		SagasCompensated:    tm.metrics.SagasCompensated,
		AvgSagaDuration:     tm.metrics.AvgSagaDuration,
		TransactionConflicts: tm.metrics.TransactionConflicts,
		CompensationEvents:  tm.metrics.CompensationEvents,
		OperationLatencies:  make(map[string]int64),
	}

	for k, v := range tm.metrics.OperationLatencies {
		metricsCopy.OperationLatencies[k] = v
	}

	return metricsCopy
}