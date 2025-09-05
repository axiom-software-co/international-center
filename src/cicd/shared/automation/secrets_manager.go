package automation

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

// SecretsManager manages secrets for infrastructure deployment
type SecretsManager struct {
	providers map[string]SecretsProvider
	cache     map[string]string
	cacheMux  sync.RWMutex
}

// SecretsProvider defines interface for different secret providers
type SecretsProvider interface {
	GetSecret(ctx context.Context, key string) (string, error)
	SetSecret(ctx context.Context, key, value string) error
	DeleteSecret(ctx context.Context, key string) error
	ListSecrets(ctx context.Context) ([]string, error)
}

// EnvironmentSecretsProvider retrieves secrets from environment variables
type EnvironmentSecretsProvider struct {
	prefix string
}

// KeyVaultSecretsProvider retrieves secrets from Azure Key Vault
type KeyVaultSecretsProvider struct {
	vaultURL     string
	tenantID     string
	clientID     string
	clientSecret string
}

// NewSecretsManager creates a new secrets manager
func NewSecretsManager() *SecretsManager {
	sm := &SecretsManager{
		providers: make(map[string]SecretsProvider),
		cache:     make(map[string]string),
	}
	
	// Register default environment provider
	sm.RegisterProvider("env", &EnvironmentSecretsProvider{prefix: "PULUMI_"})
	
	return sm
}

// RegisterProvider registers a secrets provider
func (sm *SecretsManager) RegisterProvider(name string, provider SecretsProvider) {
	sm.providers[name] = provider
}

// GetSecret retrieves a secret from the appropriate provider
func (sm *SecretsManager) GetSecret(ctx context.Context, key string) (string, error) {
	// Check cache first
	sm.cacheMux.RLock()
	if value, exists := sm.cache[key]; exists {
		sm.cacheMux.RUnlock()
		return value, nil
	}
	sm.cacheMux.RUnlock()

	// Try providers in order
	for providerName, provider := range sm.providers {
		value, err := provider.GetSecret(ctx, key)
		if err == nil && value != "" {
			// Cache the value
			sm.cacheMux.Lock()
			sm.cache[key] = value
			sm.cacheMux.Unlock()
			
			return value, nil
		}
		
		// Log but continue to next provider
		fmt.Printf("Provider %s failed to retrieve secret %s: %v\n", providerName, key, err)
	}

	return "", fmt.Errorf("secret %s not found in any provider", key)
}

// SetSecret sets a secret using the appropriate provider
func (sm *SecretsManager) SetSecret(ctx context.Context, key, value string) error {
	// Clear cache entry
	sm.cacheMux.Lock()
	delete(sm.cache, key)
	sm.cacheMux.Unlock()

	// Try to set in first available provider
	for _, provider := range sm.providers {
		if err := provider.SetSecret(ctx, key, value); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to set secret %s in any provider", key)
}

// ClearCache clears the secrets cache
func (sm *SecretsManager) ClearCache() {
	sm.cacheMux.Lock()
	sm.cache = make(map[string]string)
	sm.cacheMux.Unlock()
}

// GetEnvironmentSecrets retrieves all secrets for a specific environment
func (sm *SecretsManager) GetEnvironmentSecrets(ctx context.Context, environment string) (map[string]string, error) {
	secrets := make(map[string]string)
	
	// Define required secrets per environment
	requiredSecrets := sm.getRequiredSecretsForEnvironment(environment)
	
	for _, secretKey := range requiredSecrets {
		value, err := sm.GetSecret(ctx, secretKey)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve required secret %s for environment %s: %w", secretKey, environment, err)
		}
		secrets[secretKey] = value
	}
	
	return secrets, nil
}

// getRequiredSecretsForEnvironment returns required secrets for each environment
func (sm *SecretsManager) getRequiredSecretsForEnvironment(environment string) []string {
	baseSecrets := []string{
		"azure:clientId",
		"azure:clientSecret",
		"azure:tenantId",
		"azure:subscriptionId",
		"database:adminPassword",
		"storage:accessKey",
	}
	
	switch environment {
	case "production":
		return append(baseSecrets, []string{
			"grafana:apiKey",
			"prometheus:apiKey",
			"loki:apiKey",
			"jwt:signingKey",
			"encryption:key",
			"backup:encryptionKey",
		}...)
	case "staging":
		return append(baseSecrets, []string{
			"grafana:apiKey",
			"prometheus:apiKey",
			"loki:apiKey",
			"jwt:signingKey",
		}...)
	default:
		return baseSecrets
	}
}

// EnvironmentSecretsProvider implementation
func NewEnvironmentSecretsProvider(prefix string) *EnvironmentSecretsProvider {
	return &EnvironmentSecretsProvider{prefix: prefix}
}

func (esp *EnvironmentSecretsProvider) GetSecret(ctx context.Context, key string) (string, error) {
	// Convert key format: "azure:clientId" -> "PULUMI_AZURE_CLIENT_ID"
	envKey := esp.keyToEnvVar(key)
	value := os.Getenv(envKey)
	if value == "" {
		return "", fmt.Errorf("environment variable %s not found", envKey)
	}
	return value, nil
}

func (esp *EnvironmentSecretsProvider) SetSecret(ctx context.Context, key, value string) error {
	envKey := esp.keyToEnvVar(key)
	return os.Setenv(envKey, value)
}

func (esp *EnvironmentSecretsProvider) DeleteSecret(ctx context.Context, key string) error {
	envKey := esp.keyToEnvVar(key)
	return os.Unsetenv(envKey)
}

func (esp *EnvironmentSecretsProvider) ListSecrets(ctx context.Context) ([]string, error) {
	var secrets []string
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, esp.prefix) {
			key := strings.SplitN(env, "=", 2)[0]
			secrets = append(secrets, esp.envVarToKey(key))
		}
	}
	return secrets, nil
}

func (esp *EnvironmentSecretsProvider) keyToEnvVar(key string) string {
	// Convert "azure:clientId" to "PULUMI_AZURE_CLIENT_ID"
	parts := strings.Split(key, ":")
	envKey := esp.prefix + strings.ToUpper(strings.Join(parts, "_"))
	return envKey
}

func (esp *EnvironmentSecretsProvider) envVarToKey(envVar string) string {
	// Convert "PULUMI_AZURE_CLIENT_ID" to "azure:clientId"
	key := strings.TrimPrefix(envVar, esp.prefix)
	parts := strings.Split(strings.ToLower(key), "_")
	if len(parts) >= 2 {
		return parts[0] + ":" + strings.Join(parts[1:], "")
	}
	return strings.ToLower(key)
}

// KeyVaultSecretsProvider implementation
func NewKeyVaultSecretsProvider(vaultURL, tenantID, clientID, clientSecret string) *KeyVaultSecretsProvider {
	return &KeyVaultSecretsProvider{
		vaultURL:     vaultURL,
		tenantID:     tenantID,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (kvsp *KeyVaultSecretsProvider) GetSecret(ctx context.Context, key string) (string, error) {
	// TODO: Implement Azure Key Vault integration
	// This would use the Azure SDK to retrieve secrets from Key Vault
	return "", fmt.Errorf("key vault integration not yet implemented")
}

func (kvsp *KeyVaultSecretsProvider) SetSecret(ctx context.Context, key, value string) error {
	// TODO: Implement Azure Key Vault integration
	return fmt.Errorf("key vault integration not yet implemented")
}

func (kvsp *KeyVaultSecretsProvider) DeleteSecret(ctx context.Context, key string) error {
	// TODO: Implement Azure Key Vault integration  
	return fmt.Errorf("key vault integration not yet implemented")
}

func (kvsp *KeyVaultSecretsProvider) ListSecrets(ctx context.Context) ([]string, error) {
	// TODO: Implement Azure Key Vault integration
	return nil, fmt.Errorf("key vault integration not yet implemented")
}

// ValidateSecretsAvailability validates that all required secrets are available
func (sm *SecretsManager) ValidateSecretsAvailability(ctx context.Context, environment string) error {
	requiredSecrets := sm.getRequiredSecretsForEnvironment(environment)
	var missingSecrets []string
	
	for _, secretKey := range requiredSecrets {
		_, err := sm.GetSecret(ctx, secretKey)
		if err != nil {
			missingSecrets = append(missingSecrets, secretKey)
		}
	}
	
	if len(missingSecrets) > 0 {
		return fmt.Errorf("missing required secrets for environment %s: %v", environment, missingSecrets)
	}
	
	return nil
}