package dapr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/axiom-software-co/international-center/src/backend/internal/shared/domain"
	"github.com/dapr/go-sdk/client"
)

// StateStore wraps Dapr state store operations
type StateStore struct {
	client    *Client
	storeName string
}

// StateOptions contains options for state operations
type StateOptions struct {
	Consistency client.StateConsistency
	Concurrency client.StateConcurrency
}

// NewStateStore creates a new state store instance
func NewStateStore(client *Client) *StateStore {
	storeName := getEnv("DAPR_STATE_STORE_NAME", "statestore-postgresql")
	
	return &StateStore{
		client:    client,
		storeName: storeName,
	}
}

// Save saves an entity to the state store
func (s *StateStore) Save(ctx context.Context, key string, value interface{}, options *StateOptions) error {
	// Add operation-specific timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	data, err := json.Marshal(value)
	if err != nil {
		return domain.WrapError(err, fmt.Sprintf("failed to marshal value for state store key %s", key))
	}

	item := &client.SetStateItem{
		Key:   key,
		Value: data,
	}

	if options != nil {
		item.Options = &client.StateOptions{
			Consistency: options.Consistency,
			Concurrency: options.Concurrency,
		}
	}

	err = s.client.GetClient().SaveState(timeoutCtx, s.storeName, item)
	if err != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return domain.NewTimeoutError(fmt.Sprintf("state store save operation for key %s", key))
		}
		return domain.NewDependencyError("state store", domain.WrapError(err, fmt.Sprintf("failed to save state for key %s", key)))
	}

	return nil
}

// Get retrieves an entity from the state store
func (s *StateStore) Get(ctx context.Context, key string, target interface{}) (bool, error) {
	// Add operation-specific timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := s.client.GetClient().GetState(timeoutCtx, s.storeName, key, nil)
	if err != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return false, domain.NewTimeoutError(fmt.Sprintf("state store get operation for key %s", key))
		}
		return false, domain.NewDependencyError("state store", domain.WrapError(err, fmt.Sprintf("failed to get state for key %s", key)))
	}

	if result.Value == nil || len(result.Value) == 0 {
		return false, nil
	}

	err = json.Unmarshal(result.Value, target)
	if err != nil {
		return false, domain.WrapError(err, fmt.Sprintf("failed to unmarshal state for key %s", key))
	}

	return true, nil
}

// Delete removes an entity from the state store
func (s *StateStore) Delete(ctx context.Context, key string, options *StateOptions) error {
	deleteOptions := &client.DeleteStateOptions{}
	
	if options != nil {
		deleteOptions.Concurrency = options.Concurrency
	}

	err := s.client.GetClient().DeleteState(ctx, s.storeName, key, deleteOptions)
	if err != nil {
		return fmt.Errorf("failed to delete state for key %s: %w", key, err)
	}

	return nil
}

// GetBulk retrieves multiple entities from the state store
func (s *StateStore) GetBulk(ctx context.Context, keys []string, targets map[string]interface{}) error {
	results, err := s.client.GetClient().GetBulkState(ctx, s.storeName, keys, nil, 100)
	if err != nil {
		return fmt.Errorf("failed to get bulk state: %w", err)
	}

	for _, result := range results {
		if result.Value != nil && len(result.Value) > 0 {
			if target, exists := targets[result.Key]; exists {
				err = json.Unmarshal(result.Value, target)
				if err != nil {
					return fmt.Errorf("failed to unmarshal state for key %s: %w", result.Key, err)
				}
			}
		}
	}

	return nil
}

// SaveBulk saves multiple entities to the state store
func (s *StateStore) SaveBulk(ctx context.Context, items map[string]interface{}, options *StateOptions) error {
	var stateItems []*client.SetStateItem

	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}

		item := &client.SetStateItem{
			Key:   key,
			Value: data,
		}

		if options != nil {
			item.Options = &client.StateOptions{
				Consistency: options.Consistency,
				Concurrency: options.Concurrency,
			}
		}

		stateItems = append(stateItems, item)
	}

	err := s.client.GetClient().SaveBulkState(ctx, s.storeName, stateItems)
	if err != nil {
		return fmt.Errorf("failed to save bulk state: %w", err)
	}

	return nil
}

// Query executes a query against the state store
func (s *StateStore) Query(ctx context.Context, query string) ([]client.BulkStateItem, error) {
	results, err := s.client.GetClient().QueryStateAlpha1(ctx, s.storeName, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query state: %w", err)
	}

	return results.Results, nil
}

// Transaction executes multiple state operations as a transaction
func (s *StateStore) Transaction(ctx context.Context, operations []client.TransactionalStateOperation) error {
	err := s.client.GetClient().ExecuteStateTransaction(ctx, s.storeName, nil, operations)
	if err != nil {
		return fmt.Errorf("failed to execute state transaction: %w", err)
	}

	return nil
}

// CreateKey creates a standardized key for the given domain and entity
func (s *StateStore) CreateKey(domain, entityType, id string) string {
	return fmt.Sprintf("%s:%s:%s", domain, entityType, id)
}

// CreateIndexKey creates a standardized index key for efficient lookups
func (s *StateStore) CreateIndexKey(domain, entityType, indexName, indexValue string) string {
	return fmt.Sprintf("idx:%s:%s:%s:%s", domain, entityType, indexName, indexValue)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}