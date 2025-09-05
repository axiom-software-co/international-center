package notifications

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// CircuitBreakerState represents the current state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// CircuitBreakerConfig contains configuration for circuit breaker
type CircuitBreakerConfig struct {
	MaxFailures      int           `json:"max_failures"`
	ResetTimeout     time.Duration `json:"reset_timeout"`
	FailureThreshold float64       `json:"failure_threshold"`
	MinRequests      int           `json:"min_requests"`
}

// CircuitBreaker implements the circuit breaker pattern for fault tolerance
type CircuitBreaker struct {
	config       *CircuitBreakerConfig
	logger       *slog.Logger
	state        int32 // atomic access to CircuitBreakerState
	failures     int32 // atomic counter
	requests     int32 // atomic counter
	lastFailTime int64 // atomic timestamp
	mutex        sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig, logger *slog.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		config:   config,
		logger:   logger,
		state:    int32(CircuitBreakerClosed),
		failures: 0,
		requests: 0,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	if !cb.allowRequest() {
		return ErrCircuitBreakerOpen
	}

	atomic.AddInt32(&cb.requests, 1)

	err := fn(ctx)
	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

// allowRequest checks if the request should be allowed based on circuit breaker state
func (cb *CircuitBreaker) allowRequest() bool {
	state := CircuitBreakerState(atomic.LoadInt32(&cb.state))
	
	switch state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		return cb.shouldAttemptReset()
	case CircuitBreakerHalfOpen:
		return true
	default:
		return false
	}
}

// shouldAttemptReset checks if enough time has passed to attempt reset
func (cb *CircuitBreaker) shouldAttemptReset() bool {
	lastFailTime := atomic.LoadInt64(&cb.lastFailTime)
	return time.Since(time.Unix(0, lastFailTime)) >= cb.config.ResetTimeout
}

// onSuccess handles successful execution
func (cb *CircuitBreaker) onSuccess() {
	state := CircuitBreakerState(atomic.LoadInt32(&cb.state))
	
	if state == CircuitBreakerHalfOpen {
		cb.reset()
		cb.logger.Info("Circuit breaker reset to closed state")
	}
}

// onFailure handles failed execution
func (cb *CircuitBreaker) onFailure() {
	atomic.AddInt32(&cb.failures, 1)
	atomic.StoreInt64(&cb.lastFailTime, time.Now().UnixNano())
	
	failures := atomic.LoadInt32(&cb.failures)
	requests := atomic.LoadInt32(&cb.requests)
	
	if requests >= int32(cb.config.MinRequests) {
		failureRate := float64(failures) / float64(requests)
		if failureRate >= cb.config.FailureThreshold {
			cb.trip()
		}
	}
}

// trip opens the circuit breaker
func (cb *CircuitBreaker) trip() {
	if atomic.CompareAndSwapInt32(&cb.state, int32(CircuitBreakerClosed), int32(CircuitBreakerOpen)) ||
		atomic.CompareAndSwapInt32(&cb.state, int32(CircuitBreakerHalfOpen), int32(CircuitBreakerOpen)) {
		cb.logger.Warn("Circuit breaker tripped to open state",
			"failures", atomic.LoadInt32(&cb.failures),
			"requests", atomic.LoadInt32(&cb.requests))
	}
}

// reset closes the circuit breaker and resets counters
func (cb *CircuitBreaker) reset() {
	atomic.StoreInt32(&cb.state, int32(CircuitBreakerClosed))
	atomic.StoreInt32(&cb.failures, 0)
	atomic.StoreInt32(&cb.requests, 0)
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	return CircuitBreakerState(atomic.LoadInt32(&cb.state))
}

// GetMetrics returns current metrics
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"state":    cb.GetState(),
		"failures": atomic.LoadInt32(&cb.failures),
		"requests": atomic.LoadInt32(&cb.requests),
	}
}

// RetryConfig contains configuration for retry logic
type RetryConfig struct {
	MaxAttempts    int           `json:"max_attempts"`
	InitialDelay   time.Duration `json:"initial_delay"`
	MaxDelay       time.Duration `json:"max_delay"`
	BackoffFactor  float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// RetryExecutor implements exponential backoff retry logic
type RetryExecutor struct {
	config *RetryConfig
	logger *slog.Logger
}

// NewRetryExecutor creates a new retry executor
func NewRetryExecutor(config *RetryConfig, logger *slog.Logger) *RetryExecutor {
	return &RetryExecutor{
		config: config,
		logger: logger,
	}
}

// Execute executes a function with exponential backoff retry
func (r *RetryExecutor) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	var lastErr error
	
	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}
		
		err := fn(ctx)
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		if attempt == r.config.MaxAttempts {
			break
		}
		
		if !r.isRetryableError(err) {
			r.logger.Debug("Non-retryable error, not retrying", 
				"error", err,
				"attempt", attempt)
			return err
		}
		
		delay := r.calculateDelay(attempt)
		r.logger.Debug("Retrying operation after delay",
			"attempt", attempt,
			"delay", delay,
			"error", err)
			
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during retry delay: %w", ctx.Err())
		case <-time.After(delay):
		}
	}
	
	return fmt.Errorf("operation failed after %d attempts: %w", r.config.MaxAttempts, lastErr)
}

// calculateDelay calculates the delay for exponential backoff
func (r *RetryExecutor) calculateDelay(attempt int) time.Duration {
	delay := float64(r.config.InitialDelay) * math.Pow(r.config.BackoffFactor, float64(attempt-1))
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}
	return time.Duration(delay)
}

// isRetryableError checks if an error is retryable
func (r *RetryExecutor) isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for specific retryable errors
	for _, retryableError := range r.config.RetryableErrors {
		if errors.Is(err, errors.New(retryableError)) {
			return true
		}
	}
	
	// Check for timeout errors
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}
	
	// Default to retryable for network-related errors
	return true
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	tokens    int64
	maxTokens int64
	refillRate time.Duration
	lastRefill int64
	mutex     sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int64, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now().UnixNano(),
	}
}

// Allow checks if an operation is allowed based on rate limit
func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now().UnixNano()
	elapsed := time.Duration(now - rl.lastRefill)
	
	// Refill tokens based on elapsed time
	tokensToAdd := int64(elapsed / rl.refillRate)
	if tokensToAdd > 0 {
		rl.tokens = min(rl.maxTokens, rl.tokens+tokensToAdd)
		rl.lastRefill = now
	}
	
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	
	return false
}

// WaitFor waits for a token to become available
func (rl *RateLimiter) WaitFor(ctx context.Context) error {
	for {
		if rl.Allow() {
			return nil
		}
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			continue
		}
	}
}

// min returns the minimum of two int64 values
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// Common errors
var (
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
)

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Service   string                 `json:"service"`
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// HealthChecker performs health checks on services
type HealthChecker struct {
	logger *slog.Logger
	checks map[string]func(ctx context.Context) error
	mutex  sync.RWMutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(logger *slog.Logger) *HealthChecker {
	return &HealthChecker{
		logger: logger,
		checks: make(map[string]func(ctx context.Context) error),
	}
}

// RegisterCheck registers a health check for a service
func (hc *HealthChecker) RegisterCheck(name string, checkFn func(ctx context.Context) error) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	hc.checks[name] = checkFn
}

// CheckAll performs all registered health checks
func (hc *HealthChecker) CheckAll(ctx context.Context) map[string]*HealthCheckResult {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	
	results := make(map[string]*HealthCheckResult)
	
	for name, checkFn := range hc.checks {
		result := &HealthCheckResult{
			Service:   name,
			Timestamp: time.Now(),
		}
		
		err := checkFn(ctx)
		if err != nil {
			result.Status = "unhealthy"
			result.Error = err.Error()
		} else {
			result.Status = "healthy"
		}
		
		results[name] = result
	}
	
	return results
}