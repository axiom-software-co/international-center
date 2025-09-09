package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// WorkerPool manages a pool of workers for concurrent processing
type WorkerPool struct {
	workerCount   int
	taskChannel   chan func()
	stopChannel   chan struct{}
	waitGroup     sync.WaitGroup
	logger        *slog.Logger
	activeWorkers int32
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workerCount int, logger *slog.Logger) *WorkerPool {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}
	
	return &WorkerPool{
		workerCount: workerCount,
		taskChannel: make(chan func(), workerCount*2), // Buffer to prevent blocking
		stopChannel: make(chan struct{}),
		logger:      logger,
	}
}

// Start initializes and starts all workers
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.logger.Info("Starting worker pool", "workers", wp.workerCount)
	
	for i := 0; i < wp.workerCount; i++ {
		wp.waitGroup.Add(1)
		go wp.worker(ctx, i)
	}
}

// worker is the main worker goroutine
func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.waitGroup.Done()
	atomic.AddInt32(&wp.activeWorkers, 1)
	
	wp.logger.Debug("Worker started", "worker_id", id)
	
	for {
		select {
		case <-ctx.Done():
			wp.logger.Debug("Worker stopped due to context cancellation", "worker_id", id)
			atomic.AddInt32(&wp.activeWorkers, -1)
			return
		case <-wp.stopChannel:
			wp.logger.Debug("Worker stopped", "worker_id", id)
			atomic.AddInt32(&wp.activeWorkers, -1)
			return
		case task := <-wp.taskChannel:
			if task != nil {
				wp.executeTask(task, id)
			}
		}
	}
}

// executeTask safely executes a task with panic recovery
func (wp *WorkerPool) executeTask(task func(), workerID int) {
	defer func() {
		if r := recover(); r != nil {
			wp.logger.Error("Worker panic recovered",
				"worker_id", workerID,
				"panic", r)
		}
	}()
	
	task()
}

// Submit submits a task to the worker pool
func (wp *WorkerPool) Submit(task func()) error {
	select {
	case wp.taskChannel <- task:
		return nil
	default:
		return fmt.Errorf("worker pool is full, cannot submit task")
	}
}

// SubmitWithTimeout submits a task with timeout
func (wp *WorkerPool) SubmitWithTimeout(task func(), timeout time.Duration) error {
	select {
	case wp.taskChannel <- task:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout submitting task to worker pool")
	}
}

// Stop gracefully shuts down the worker pool
func (wp *WorkerPool) Stop(timeout time.Duration) error {
	wp.logger.Info("Stopping worker pool")
	
	close(wp.stopChannel)
	
	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		wp.waitGroup.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		wp.logger.Info("Worker pool stopped successfully")
		return nil
	case <-time.After(timeout):
		wp.logger.Warn("Worker pool shutdown timeout")
		return fmt.Errorf("worker pool shutdown timeout after %v", timeout)
	}
}

// GetActiveWorkers returns the number of active workers
func (wp *WorkerPool) GetActiveWorkers() int32 {
	return atomic.LoadInt32(&wp.activeWorkers)
}

// GetMetrics returns worker pool metrics
func (wp *WorkerPool) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"total_workers":  wp.workerCount,
		"active_workers": wp.GetActiveWorkers(),
		"pending_tasks":  len(wp.taskChannel),
		"channel_cap":    cap(wp.taskChannel),
	}
}

// BatchProcessor handles batch processing of notifications
type BatchProcessor struct {
	batchSize     int
	flushInterval time.Duration
	processor     func([]interface{}) error
	items         []interface{}
	mutex         sync.Mutex
	ticker        *time.Ticker
	stopChannel   chan struct{}
	logger        *slog.Logger
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(
	batchSize int,
	flushInterval time.Duration,
	processor func([]interface{}) error,
	logger *slog.Logger,
) *BatchProcessor {
	return &BatchProcessor{
		batchSize:     batchSize,
		flushInterval: flushInterval,
		processor:     processor,
		items:         make([]interface{}, 0, batchSize),
		stopChannel:   make(chan struct{}),
		logger:        logger,
	}
}

// Start begins batch processing
func (bp *BatchProcessor) Start(ctx context.Context) {
	bp.ticker = time.NewTicker(bp.flushInterval)
	
	go func() {
		for {
			select {
			case <-ctx.Done():
				bp.flushBatch()
				return
			case <-bp.stopChannel:
				bp.flushBatch()
				return
			case <-bp.ticker.C:
				bp.flushBatch()
			}
		}
	}()
}

// Add adds an item to the batch
func (bp *BatchProcessor) Add(item interface{}) error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()
	
	bp.items = append(bp.items, item)
	
	if len(bp.items) >= bp.batchSize {
		return bp.flushBatchUnsafe()
	}
	
	return nil
}

// flushBatch processes the current batch
func (bp *BatchProcessor) flushBatch() error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()
	return bp.flushBatchUnsafe()
}

// flushBatchUnsafe processes the current batch without locking
func (bp *BatchProcessor) flushBatchUnsafe() error {
	if len(bp.items) == 0 {
		return nil
	}
	
	batch := make([]interface{}, len(bp.items))
	copy(batch, bp.items)
	bp.items = bp.items[:0] // Reset slice but keep capacity
	
	bp.logger.Debug("Processing batch", "size", len(batch))
	
	if err := bp.processor(batch); err != nil {
		bp.logger.Error("Batch processing failed", 
			"size", len(batch),
			"error", err)
		return err
	}
	
	return nil
}

// Stop stops the batch processor
func (bp *BatchProcessor) Stop() {
	if bp.ticker != nil {
		bp.ticker.Stop()
	}
	close(bp.stopChannel)
}

// ConnectionPool manages a pool of database connections
type ConnectionPool struct {
	connections chan interface{}
	factory     func() (interface{}, error)
	cleanup     func(interface{}) error
	maxConns    int
	activeConns int32
	logger      *slog.Logger
	mutex       sync.Mutex
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(
	maxConns int,
	factory func() (interface{}, error),
	cleanup func(interface{}) error,
	logger *slog.Logger,
) *ConnectionPool {
	return &ConnectionPool{
		connections: make(chan interface{}, maxConns),
		factory:     factory,
		cleanup:     cleanup,
		maxConns:    maxConns,
		logger:      logger,
	}
}

// Get acquires a connection from the pool
func (cp *ConnectionPool) Get(ctx context.Context) (interface{}, error) {
	select {
	case conn := <-cp.connections:
		return conn, nil
	default:
		// Try to create new connection if under limit
		if atomic.LoadInt32(&cp.activeConns) < int32(cp.maxConns) {
			conn, err := cp.factory()
			if err != nil {
				return nil, fmt.Errorf("failed to create connection: %w", err)
			}
			atomic.AddInt32(&cp.activeConns, 1)
			return conn, nil
		}
		
		// Wait for available connection
		select {
		case conn := <-cp.connections:
			return conn, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// Put returns a connection to the pool
func (cp *ConnectionPool) Put(conn interface{}) {
	select {
	case cp.connections <- conn:
		// Connection returned to pool
	default:
		// Pool is full, cleanup connection
		if err := cp.cleanup(conn); err != nil {
			cp.logger.Error("Failed to cleanup connection", "error", err)
		}
		atomic.AddInt32(&cp.activeConns, -1)
	}
}

// Close closes all connections in the pool
func (cp *ConnectionPool) Close() error {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	
	close(cp.connections)
	
	var lastErr error
	for conn := range cp.connections {
		if err := cp.cleanup(conn); err != nil {
			lastErr = err
			cp.logger.Error("Failed to cleanup connection during close", "error", err)
		}
		atomic.AddInt32(&cp.activeConns, -1)
	}
	
	return lastErr
}

// GetMetrics returns connection pool metrics
func (cp *ConnectionPool) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"max_connections":       cp.maxConns,
		"active_connections":    atomic.LoadInt32(&cp.activeConns),
		"available_connections": len(cp.connections),
	}
}

// CacheEntry represents a cached item
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired
func (ce *CacheEntry) IsExpired() bool {
	return time.Now().After(ce.ExpiresAt)
}

// MemoryCache implements an in-memory cache with TTL
type MemoryCache struct {
	items       map[string]*CacheEntry
	mutex       sync.RWMutex
	defaultTTL  time.Duration
	cleanupTick *time.Ticker
	stopCleanup chan struct{}
	logger      *slog.Logger
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(defaultTTL time.Duration, cleanupInterval time.Duration, logger *slog.Logger) *MemoryCache {
	cache := &MemoryCache{
		items:       make(map[string]*CacheEntry),
		defaultTTL:  defaultTTL,
		cleanupTick: time.NewTicker(cleanupInterval),
		stopCleanup: make(chan struct{}),
		logger:      logger,
	}
	
	// Start cleanup goroutine
	go cache.cleanupExpired()
	
	return cache
}

// Get retrieves a value from the cache
func (mc *MemoryCache) Get(key string) (interface{}, bool) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	entry, exists := mc.items[key]
	if !exists {
		return nil, false
	}
	
	if entry.IsExpired() {
		// Remove expired entry
		delete(mc.items, key)
		return nil, false
	}
	
	return entry.Value, true
}

// Set stores a value in the cache with default TTL
func (mc *MemoryCache) Set(key string, value interface{}) {
	mc.SetWithTTL(key, value, mc.defaultTTL)
}

// SetWithTTL stores a value in the cache with custom TTL
func (mc *MemoryCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	mc.items[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Delete removes a value from the cache
func (mc *MemoryCache) Delete(key string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	delete(mc.items, key)
}

// Clear removes all items from the cache
func (mc *MemoryCache) Clear() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	mc.items = make(map[string]*CacheEntry)
}

// cleanupExpired removes expired entries
func (mc *MemoryCache) cleanupExpired() {
	for {
		select {
		case <-mc.stopCleanup:
			return
		case <-mc.cleanupTick.C:
			mc.mutex.Lock()
			now := time.Now()
			expired := 0
			
			for key, entry := range mc.items {
				if now.After(entry.ExpiresAt) {
					delete(mc.items, key)
					expired++
				}
			}
			
			if expired > 0 {
				mc.logger.Debug("Cleaned up expired cache entries", "count", expired)
			}
			mc.mutex.Unlock()
		}
	}
}

// Stop stops the cache cleanup goroutine
func (mc *MemoryCache) Stop() {
	mc.cleanupTick.Stop()
	close(mc.stopCleanup)
}

// GetMetrics returns cache metrics
func (mc *MemoryCache) GetMetrics() map[string]interface{} {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	return map[string]interface{}{
		"total_items": len(mc.items),
		"default_ttl": mc.defaultTTL.String(),
	}
}