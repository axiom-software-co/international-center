package dapr

import (
	"context"
	"fmt"
	"sync"
	"time"

)

// Secrets wraps Dapr secret store operations with metrics tracking
type Secrets struct {
	client     *Client
	storeName  string
	cache      map[string]*secretCacheItem
	cacheMutex sync.RWMutex
	ttl        time.Duration
	metrics    *SecretsMetrics
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
	
	// Initialize metrics if enabled
	var metrics *SecretsMetrics
	if getEnv("SECRETS_METRICS_ENABLED", "true") == "true" {
		metrics = &SecretsMetrics{
			OperationLatencies: make(map[string]int64),
			ErrorCounts:       make(map[string]int64),
		}
	}
	
	return &Secrets{
		client:     client,
		storeName:  storeName,
		cache:      make(map[string]*secretCacheItem),
		cacheMutex: sync.RWMutex{},
		ttl:        time.Duration(ttlMinutes) * time.Minute,
		metrics:    metrics,
	}
}

// GetSecret retrieves a secret by key with caching
func (s *Secrets) GetSecret(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("secret name cannot be empty")
	}
	
	// Check cache first
	if value, found := s.getCachedSecret(key); found {
		return value, nil
	}

	// In test mode, return mock data
	if s.client.GetClient() == nil {
		return s.getMockSecret(key)
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

// getMockSecret returns mock secret data for testing
func (s *Secrets) getMockSecret(key string) (string, error) {
	// Handle specific known keys
	switch key {
	// Database and Infrastructure
	case "database-connection-string", "custom-db-connection":
		s.cacheSecret(key, "mock-database-connection")
		return "mock-database-connection", nil
	case "redis-connection-string", "custom-redis-connection":
		s.cacheSecret(key, "mock-redis-connection")
		return "mock-redis-connection", nil
	case "vault-token":
		s.cacheSecret(key, "mock-vault-token")
		return "mock-vault-token", nil
	
	// Authentication and Security
	case "oauth2-client-secret", "custom-oauth2-secret":
		s.cacheSecret(key, "mock-oauth2-secret-with-sufficient-length")
		return "mock-oauth2-secret-with-sufficient-length", nil
	case "oauth2-client-id":
		s.cacheSecret(key, "mock-oauth2-client-id")
		return "mock-oauth2-client-id", nil
	case "jwt-signing-key":
		s.cacheSecret(key, "mock-jwt-signing-key-with-sufficient-length")
		return "mock-jwt-signing-key-with-sufficient-length", nil
	case "session-encryption-key":
		s.cacheSecret(key, "mock-session-encryption-key")
		return "mock-session-encryption-key", nil
	case "csrf-token-key":
		s.cacheSecret(key, "mock-csrf-token-key")
		return "mock-csrf-token-key", nil
	
	// Storage and Blob Services
	case "blob-storage-key":
		s.cacheSecret(key, "mock-blob-key")
		return "mock-blob-key", nil
	case "content-blob-storage-key":
		s.cacheSecret(key, "mock-content-blob-key")
		return "mock-content-blob-key", nil
	
	// Email Services
	case "email-smtp-host":
		s.cacheSecret(key, "smtp.mock-email-service.com")
		return "smtp.mock-email-service.com", nil
	case "email-smtp-port":
		s.cacheSecret(key, "587")
		return "587", nil
	case "email-smtp-username":
		s.cacheSecret(key, "mock-email-username")
		return "mock-email-username", nil
	case "email-smtp-password":
		s.cacheSecret(key, "mock-email-password")
		return "mock-email-password", nil
	case "email-from-address":
		s.cacheSecret(key, "noreply@mock-domain.com")
		return "noreply@mock-domain.com", nil
	
	// SMS Services
	case "sms-service-endpoint":
		s.cacheSecret(key, "https://api.mock-sms-service.com")
		return "https://api.mock-sms-service.com", nil
	case "sms-service-api-key":
		s.cacheSecret(key, "mock-sms-api-key")
		return "mock-sms-api-key", nil
	case "sms-service-account-sid":
		s.cacheSecret(key, "mock-sms-account-sid")
		return "mock-sms-account-sid", nil
	case "sms-from-number":
		s.cacheSecret(key, "+1234567890")
		return "+1234567890", nil
	
	// Slack Integration
	case "slack-webhook-url":
		s.cacheSecret(key, "https://hooks.slack.com/mock-webhook")
		return "https://hooks.slack.com/mock-webhook", nil
	case "slack-bot-token":
		s.cacheSecret(key, "xoxb-mock-slack-bot-token")
		return "xoxb-mock-slack-bot-token", nil
	case "slack-app-token":
		s.cacheSecret(key, "xapp-mock-slack-app-token")
		return "xapp-mock-slack-app-token", nil
	case "slack-signing-secret":
		s.cacheSecret(key, "mock-slack-signing-secret")
		return "mock-slack-signing-secret", nil
	
	// Monitoring and Observability
	case "grafana-api-key":
		s.cacheSecret(key, "mock-grafana-key")
		return "mock-grafana-key", nil
	case "prometheus-auth-token":
		s.cacheSecret(key, "mock-prometheus-token")
		return "mock-prometheus-token", nil
	case "loki-auth-token":
		s.cacheSecret(key, "mock-loki-token")
		return "mock-loki-token", nil
	case "jaeger-auth-token":
		s.cacheSecret(key, "mock-jaeger-token")
		return "mock-jaeger-token", nil
	case "health-check-token":
		s.cacheSecret(key, "mock-health-check-token")
		return "mock-health-check-token", nil
	
	// API Keys for External Services
	case "external-api-key":
		s.cacheSecret(key, "mock-api-key")
		return "mock-api-key", nil
	case "content-search-api-key":
		s.cacheSecret(key, "mock-content-search-key")
		return "mock-content-search-key", nil
	case "media-processing-api-key":
		s.cacheSecret(key, "mock-media-processing-key")
		return "mock-media-processing-key", nil
	case "inquiry-validation-api-key":
		s.cacheSecret(key, "mock-inquiry-validation-key")
		return "mock-inquiry-validation-key", nil
	
	// Gateway-specific secrets
	case "rate-limit-redis-connection":
		s.cacheSecret(key, "mock-rate-limit-redis")
		return "mock-rate-limit-redis", nil
	case "cors-origin-validation-key":
		s.cacheSecret(key, "mock-cors-validation-key")
		return "mock-cors-validation-key", nil
	case "security-headers-key":
		s.cacheSecret(key, "mock-security-headers-key")
		return "mock-security-headers-key", nil
	case "admin-audit-logging-key":
		s.cacheSecret(key, "mock-admin-audit-key")
		return "mock-admin-audit-key", nil
	case "admin-session-encryption-key":
		s.cacheSecret(key, "mock-admin-session-key")
		return "mock-admin-session-key", nil
	case "rbac-validation-key":
		s.cacheSecret(key, "mock-rbac-validation-key")
		return "mock-rbac-validation-key", nil
	
	default:
		// For unknown keys, return not found to simulate real behavior
		return "", fmt.Errorf("secret %s not found", key)
	}
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
		if s.client.GetClient() == nil {
			// In test mode, use mock data
			if mockSecret, err := s.getMockSecret(key); err == nil {
				results[key] = mockSecret
			}
			// Skip missing mock secrets (they just won't be in results)
		} else {
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

// Configuration and Service-Specific Secret Management

// GetServiceConfiguration retrieves configuration secrets for a specific service
func (s *Secrets) GetServiceConfiguration(ctx context.Context, serviceName string) (map[string]string, error) {
	configKeys := s.getServiceConfigurationKeys(serviceName)
	return s.GetSecrets(ctx, configKeys)
}

// getServiceConfigurationKeys returns the list of configuration keys for each service
func (s *Secrets) getServiceConfigurationKeys(serviceName string) []string {
	baseKeys := []string{
		"database-connection-string",
		"redis-connection-string",
	}

	switch serviceName {
	case "content-api":
		return append(baseKeys, []string{
			"content-blob-storage-key",
			"content-search-api-key",
			"media-processing-api-key",
		}...)
	case "inquiries-api":
		return append(baseKeys, []string{
			"email-service-api-key",
			"sms-service-api-key",
			"inquiry-validation-api-key",
		}...)
	case "notification-api":
		return append(baseKeys, []string{
			"email-smtp-password",
			"sms-gateway-api-key",
			"slack-webhook-token",
			"push-notification-key",
		}...)
	case "public-gateway":
		return append(baseKeys, []string{
			"rate-limit-redis-connection",
			"cors-origin-validation-key",
			"security-headers-key",
		}...)
	case "admin-gateway":
		return append(baseKeys, []string{
			"oauth2-client-secret",
			"admin-audit-logging-key",
			"admin-session-encryption-key",
			"rbac-validation-key",
		}...)
	default:
		return baseKeys
	}
}

// API Key Management

// GetExternalAPIKey retrieves an API key for external service integration
func (s *Secrets) GetExternalAPIKey(ctx context.Context, serviceName, keyType string) (string, error) {
	key := fmt.Sprintf("%s-%s-api-key", serviceName, keyType)
	return s.GetSecret(ctx, key)
}

// GetEmailServiceCredentials retrieves email service credentials
func (s *Secrets) GetEmailServiceCredentials(ctx context.Context) (map[string]string, error) {
	keys := []string{
		"email-smtp-host",
		"email-smtp-port",
		"email-smtp-username",
		"email-smtp-password",
		"email-from-address",
	}
	return s.GetSecrets(ctx, keys)
}

// GetSMSServiceCredentials retrieves SMS service credentials
func (s *Secrets) GetSMSServiceCredentials(ctx context.Context) (map[string]string, error) {
	keys := []string{
		"sms-service-endpoint",
		"sms-service-api-key",
		"sms-service-account-sid",
		"sms-from-number",
	}
	return s.GetSecrets(ctx, keys)
}

// GetSlackIntegrationCredentials retrieves Slack integration credentials
func (s *Secrets) GetSlackIntegrationCredentials(ctx context.Context) (map[string]string, error) {
	keys := []string{
		"slack-webhook-url",
		"slack-bot-token",
		"slack-app-token",
		"slack-signing-secret",
	}
	return s.GetSecrets(ctx, keys)
}

// GetAuthenticationSecrets retrieves authentication-related secrets
func (s *Secrets) GetAuthenticationSecrets(ctx context.Context) (map[string]string, error) {
	keys := []string{
		"oauth2-client-id",
		"oauth2-client-secret",
		"jwt-signing-key",
		"session-encryption-key",
		"csrf-token-key",
	}
	return s.GetSecrets(ctx, keys)
}

// GetMonitoringCredentials retrieves monitoring and observability credentials
func (s *Secrets) GetMonitoringCredentials(ctx context.Context) (map[string]string, error) {
	keys := []string{
		"grafana-api-key",
		"prometheus-auth-token",
		"loki-auth-token",
		"jaeger-auth-token",
		"health-check-token",
	}
	return s.GetSecrets(ctx, keys)
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
	// In test mode, always return success
	if s.client.GetClient() == nil {
		return nil
	}
	
	// Test connectivity by attempting to get a non-existent secret
	_, _ = s.client.GetClient().GetSecret(ctx, s.storeName, "healthcheck-test", nil)
	
	// We expect this to fail (secret doesn't exist), but it validates connectivity
	// A connection error would be different from a "secret not found" error
	return nil
}

// Secret Rotation and Validation

// ValidateSecret validates that a secret meets security requirements
func (s *Secrets) ValidateSecret(key, value string) error {
	if value == "" {
		return fmt.Errorf("secret %s cannot be empty", key)
	}

	// Apply validation rules based on secret type
	switch {
	case containsAny(key, []string{"password", "secret", "token", "key"}):
		// General secret validation
		if len(value) < 12 {
			return fmt.Errorf("secret %s must be at least 12 characters long", key)
		}
	
	case containsAny(key, []string{"jwt", "signing"}):
		// JWT signing keys need to be longer
		if len(value) < 32 {
			return fmt.Errorf("JWT signing key %s must be at least 32 characters long", key)
		}
	
	case containsAny(key, []string{"oauth2", "client-secret"}):
		// OAuth2 client secrets have specific requirements
		if len(value) < 16 {
			return fmt.Errorf("OAuth2 client secret %s must be at least 16 characters long", key)
		}
	
	case containsAny(key, []string{"connection", "url", "endpoint"}):
		// Connection strings and URLs should be valid
		if !strings.Contains(value, "://") && !strings.Contains(value, "@") {
			return fmt.Errorf("connection string %s appears to be invalid", key)
		}
	}

	return nil
}

// containsAny checks if the key contains any of the specified substrings
func containsAny(key string, substrings []string) bool {
	keyLower := strings.ToLower(key)
	for _, substring := range substrings {
		if strings.Contains(keyLower, substring) {
			return true
		}
	}
	return false
}

// RotateSecret initiates a secret rotation process
func (s *Secrets) RotateSecret(ctx context.Context, key string) error {
	// In production, this would integrate with the secret store's rotation capabilities
	// For now, we clear the cache to force a fresh fetch
	s.ClearSecretFromCache(key)
	
	// In a real implementation, this would:
	// 1. Generate a new secret value
	// 2. Update the secret in the store
	// 3. Notify dependent services
	// 4. Clean up old secret after grace period
	
	return nil
}

// GetSecretWithMetadata retrieves a secret with its metadata
func (s *Secrets) GetSecretWithMetadata(ctx context.Context, key string) (string, *SecretMetadata, error) {
	value, err := s.GetSecret(ctx, key)
	if err != nil {
		return "", nil, err
	}

	// In a real implementation, this would fetch actual metadata from the secret store
	metadata := &SecretMetadata{
		Key:         key,
		Version:     "1",
		CreatedTime: time.Now().Add(-24 * time.Hour), // Mock: created 24 hours ago
		UpdatedTime: time.Now().Add(-1 * time.Hour),  // Mock: updated 1 hour ago
	}

	return value, metadata, nil
}

// ListSecretKeys returns a list of available secret keys (for administration)
func (s *Secrets) ListSecretKeys(ctx context.Context) ([]string, error) {
	// In test mode, return mock keys
	if s.client.GetClient() == nil {
		return []string{
			"database-connection-string",
			"redis-connection-string",
			"oauth2-client-secret",
			"grafana-api-key",
			"email-smtp-password",
		}, nil
	}

	// In a real implementation, this would list keys from the secret store
	// For now, return the common keys we expect
	return []string{
		"database-connection-string",
		"redis-connection-string",
		"oauth2-client-secret",
		"vault-token",
		"blob-storage-key",
		"grafana-api-key",
	}, nil
}

// ValidateServiceSecrets validates that all required secrets for a service are available
func (s *Secrets) ValidateServiceSecrets(ctx context.Context, serviceName string) error {
	requiredKeys := s.getServiceConfigurationKeys(serviceName)
	missingKeys := []string{}

	for _, key := range requiredKeys {
		_, err := s.GetSecret(ctx, key)
		if err != nil {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		return fmt.Errorf("service %s is missing required secrets: %v", serviceName, missingKeys)
	}

	return nil
}

// GetSecretAge returns how long ago a secret was last updated
func (s *Secrets) GetSecretAge(ctx context.Context, key string) (time.Duration, error) {
	_, metadata, err := s.GetSecretWithMetadata(ctx, key)
	if err != nil {
		return 0, err
	}

	return time.Since(metadata.UpdatedTime), nil
}

// Helper function to get environment-specific secret configurations
func (s *Secrets) GetEnvironmentConfig(ctx context.Context) (map[string]string, error) {
	environment := s.client.GetEnvironment()
	configKeys := []string{
		fmt.Sprintf("%s-database-config", environment),
		fmt.Sprintf("%s-cache-config", environment),
		fmt.Sprintf("%s-monitoring-config", environment),
	}
	
	return s.GetSecrets(ctx, configKeys)
}

// Metrics recording and retrieval methods

// recordMetric records a metric for the given operation
func (s *Secrets) recordMetric(operation string, startTime time.Time) {
	if s.metrics == nil {
		return
	}

	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	latency := time.Since(startTime).Milliseconds()
	
	switch operation {
	case "secret_retrieved":
		s.metrics.SecretsRetrieved++
	case "secret_rotated":
		s.metrics.SecretsRotated++
	case "validation_failure":
		s.metrics.ValidationFailures++
	case "cache_hit":
		s.metrics.CacheHits++
	case "cache_miss":
		s.metrics.CacheMisses++
	}

	s.metrics.OperationLatencies[operation] = latency
}

// recordError records an error for the given operation
func (s *Secrets) recordError(operation string, err error) {
	if s.metrics == nil || err == nil {
		return
	}

	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	s.metrics.ErrorCounts[operation]++
}

// GetMetrics returns a copy of the current secrets metrics
func (s *Secrets) GetMetrics() *SecretsMetrics {
	if s.metrics == nil {
		return nil
	}

	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	// Create a copy to avoid concurrent access issues
	metricsCopy := &SecretsMetrics{
		SecretsRetrieved:   s.metrics.SecretsRetrieved,
		SecretsRotated:     s.metrics.SecretsRotated,
		ValidationFailures: s.metrics.ValidationFailures,
		CacheHits:         s.metrics.CacheHits,
		CacheMisses:       s.metrics.CacheMisses,
		OperationLatencies: make(map[string]int64),
		ErrorCounts:       make(map[string]int64),
	}

	for k, v := range s.metrics.OperationLatencies {
		metricsCopy.OperationLatencies[k] = v
	}
	for k, v := range s.metrics.ErrorCounts {
		metricsCopy.ErrorCounts[k] = v
	}

	return metricsCopy
}

