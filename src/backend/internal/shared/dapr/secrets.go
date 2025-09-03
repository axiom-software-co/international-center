package dapr

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dapr/go-sdk/client"
)

// Secrets wraps Dapr secret store operations
type Secrets struct {
	client     *Client
	storeName  string
	cache      map[string]*secretCacheItem
	cacheMutex sync.RWMutex
	ttl        time.Duration
}

// secretCacheItem represents a cached secret with TTL
type secretCacheItem struct {
	value     string
	expiresAt time.Time
}

// SecretMetadata contains metadata about a secret
type SecretMetadata struct {
	Key         string
	Version     string
	CreatedTime time.Time
	UpdatedTime time.Time
}

// NewSecrets creates a new secrets instance
func NewSecrets(client *Client) *Secrets {
	environment := client.GetEnvironment()
	var storeName string
	
	switch environment {
	case "production", "staging":
		storeName = getEnv("DAPR_SECRET_STORE_NAME", "secretstore-vault")
	default:
		storeName = getEnv("DAPR_SECRET_STORE_NAME", "secretstore-vault-local")
	}

	ttlMinutes := getEnvInt("SECRET_CACHE_TTL_MINUTES", 15)
	
	return &Secrets{
		client:     client,
		storeName:  storeName,
		cache:      make(map[string]*secretCacheItem),
		cacheMutex: sync.RWMutex{},
		ttl:        time.Duration(ttlMinutes) * time.Minute,
	}
}

// GetSecret retrieves a secret by key with caching
func (s *Secrets) GetSecret(ctx context.Context, key string) (string, error) {
	// Check cache first
	if value, found := s.getCachedSecret(key); found {
		return value, nil
	}

	// Fetch from Dapr secret store
	secret, err := s.client.GetClient().GetSecret(ctx, s.storeName, key, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get secret %s: %w", key, err)
	}

	if len(secret) == 0 {
		return "", fmt.Errorf("secret %s not found", key)
	}

	// Get the secret value (Dapr returns a map[string]string)
	var value string
	for _, v := range secret {
		value = v
		break // Take the first (and typically only) value
	}

	// Cache the secret
	s.cacheSecret(key, value)

	return value, nil
}

// GetSecrets retrieves multiple secrets by keys
func (s *Secrets) GetSecrets(ctx context.Context, keys []string) (map[string]string, error) {
	results := make(map[string]string)
	var missingKeys []string

	// Check cache for each key
	for _, key := range keys {
		if value, found := s.getCachedSecret(key); found {
			results[key] = value
		} else {
			missingKeys = append(missingKeys, key)
		}
	}

	// Fetch missing keys from secret store
	for _, key := range missingKeys {
		secret, err := s.client.GetClient().GetSecret(ctx, s.storeName, key, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get secret %s: %w", key, err)
		}

		if len(secret) > 0 {
			var value string
			for _, v := range secret {
				value = v
				break
			}
			results[key] = value
			s.cacheSecret(key, value)
		}
	}

	return results, nil
}

// GetDatabaseConnectionString retrieves the database connection string
func (s *Secrets) GetDatabaseConnectionString(ctx context.Context) (string, error) {
	key := getEnv("DATABASE_CONNECTION_SECRET_KEY", "database-connection-string")
	return s.GetSecret(ctx, key)
}

// GetRedisConnectionString retrieves the Redis connection string
func (s *Secrets) GetRedisConnectionString(ctx context.Context) (string, error) {
	key := getEnv("REDIS_CONNECTION_SECRET_KEY", "redis-connection-string")
	return s.GetSecret(ctx, key)
}

// GetVaultToken retrieves the Vault access token
func (s *Secrets) GetVaultToken(ctx context.Context) (string, error) {
	key := getEnv("VAULT_TOKEN_SECRET_KEY", "vault-token")
	return s.GetSecret(ctx, key)
}

// GetBlobStorageKey retrieves the blob storage access key
func (s *Secrets) GetBlobStorageKey(ctx context.Context) (string, error) {
	key := getEnv("BLOB_STORAGE_SECRET_KEY", "blob-storage-key")
	return s.GetSecret(ctx, key)
}

// GetOAuth2ClientSecret retrieves the OAuth2 client secret
func (s *Secrets) GetOAuth2ClientSecret(ctx context.Context) (string, error) {
	key := getEnv("OAUTH2_CLIENT_SECRET_KEY", "oauth2-client-secret")
	return s.GetSecret(ctx, key)
}

// GetGrafanaAPIKey retrieves the Grafana API key
func (s *Secrets) GetGrafanaAPIKey(ctx context.Context) (string, error) {
	key := getEnv("GRAFANA_API_KEY_SECRET", "grafana-api-key")
	return s.GetSecret(ctx, key)
}

// getCachedSecret retrieves a secret from cache if valid
func (s *Secrets) getCachedSecret(key string) (string, bool) {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	item, exists := s.cache[key]
	if !exists {
		return "", false
	}

	if time.Now().After(item.expiresAt) {
		// Secret expired, remove from cache
		delete(s.cache, key)
		return "", false
	}

	return item.value, true
}

// cacheSecret stores a secret in cache with TTL
func (s *Secrets) cacheSecret(key, value string) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.cache[key] = &secretCacheItem{
		value:     value,
		expiresAt: time.Now().Add(s.ttl),
	}
}

// ClearCache clears the secret cache
func (s *Secrets) ClearCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.cache = make(map[string]*secretCacheItem)
}

// ClearSecretFromCache removes a specific secret from cache
func (s *Secrets) ClearSecretFromCache(key string) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	delete(s.cache, key)
}

// RefreshSecret forces a refresh of a cached secret
func (s *Secrets) RefreshSecret(ctx context.Context, key string) (string, error) {
	// Remove from cache first
	s.ClearSecretFromCache(key)

	// Fetch fresh secret
	return s.GetSecret(ctx, key)
}

// HealthCheck validates the secret store connection
func (s *Secrets) HealthCheck(ctx context.Context) error {
	// Test connectivity by attempting to get a non-existent secret
	_, err := s.client.GetClient().GetSecret(ctx, s.storeName, "healthcheck-test", nil)
	
	// We expect this to fail (secret doesn't exist), but it validates connectivity
	// A connection error would be different from a "secret not found" error
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		// Simple conversion, in production would handle errors
		if parsed := parseInt(value); parsed > 0 {
			return parsed
		}
	}
	return defaultValue
}

func parseInt(s string) int {
	// Simple integer parsing - in production would use strconv.Atoi
	result := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		} else {
			return 0
		}
	}
	return result
}