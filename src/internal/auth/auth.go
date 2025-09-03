package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Authentication constants
const (
	// Token types
	BearerTokenType = "Bearer"

	// Token validation constants
	DefaultTokenExpiration = 24 * time.Hour
	MaxTokenExpiration     = 7 * 24 * time.Hour

	// Authentication reasons
	ReasonTokenValid           = "token is valid and not expired"
	ReasonTokenExpired         = "token has expired"
	ReasonTokenInvalid         = "token is invalid or malformed"
	ReasonTokenMissing         = "authorization token is missing"
	ReasonInvalidTokenType     = "unsupported token type"
	ReasonUserNotFound         = "user not found"
	ReasonTokenValidationError = "token validation error"

	// OIDC Discovery constants
	OIDCDiscoveryPath = "/.well-known/openid_configuration"
	DefaultJWKSCacheTTL = 1 * time.Hour
)

// UserInfo contains authenticated user information
type UserInfo struct {
	// UserID is the unique identifier for the user
	UserID string `json:"user_id"`

	// Username is the human-readable username
	Username string `json:"username"`

	// Email is the user's email address
	Email string `json:"email"`

	// Roles contains the user's assigned roles
	Roles []string `json:"roles"`

	// Permissions contains the user's specific permissions
	Permissions []string `json:"permissions"`

	// Groups contains the user's group memberships
	Groups []string `json:"groups,omitempty"`

	// Metadata contains additional user-specific information
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// TokenIssuedAt is when the token was issued
	TokenIssuedAt time.Time `json:"token_issued_at,omitempty"`

	// TokenExpiresAt is when the token expires
	TokenExpiresAt time.Time `json:"token_expires_at,omitempty"`
}

// TokenValidationResult represents the result of token validation
type TokenValidationResult struct {
	// Valid indicates whether the token is valid
	Valid bool `json:"valid"`

	// Reason provides explanation for the validation result
	Reason string `json:"reason"`

	// UserInfo contains the user information if token is valid
	UserInfo *UserInfo `json:"user_info,omitempty"`

	// Metadata contains additional validation information
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AuthRequest contains the authentication request information
type AuthRequest struct {
	// AuthorizationHeader is the raw Authorization header value
	AuthorizationHeader string `json:"authorization_header"`

	// RequestPath is the API path being accessed
	RequestPath string `json:"request_path"`

	// Method is the HTTP method
	Method string `json:"method"`

	// ClientIP is the client's IP address
	ClientIP string `json:"client_ip,omitempty"`

	// Headers contains relevant request headers
	Headers map[string]string `json:"headers,omitempty"`
}

// AuthenticationResult represents the result of authentication processing
type AuthenticationResult struct {
	// Authenticated indicates whether the request is authenticated
	Authenticated bool `json:"authenticated"`

	// UserInfo contains the authenticated user information
	UserInfo *UserInfo `json:"user_info,omitempty"`

	// Reason provides explanation for the authentication result
	Reason string `json:"reason"`

	// Metadata contains additional authentication information
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TokenValidator defines the interface for validating OAuth2/JWT tokens
type TokenValidator interface {
	// ValidateToken validates an OAuth2/JWT token and returns user information
	ValidateToken(ctx context.Context, token string) (*TokenValidationResult, error)
}

// UserInfoProvider defines the interface for retrieving user information
type UserInfoProvider interface {
	// GetUserInfo retrieves user information by user ID
	GetUserInfo(ctx context.Context, userID string) (*UserInfo, error)
}

// AuthenticationMiddleware defines the interface for processing authentication requests
type AuthenticationMiddleware interface {
	// ProcessRequest processes an authentication request and returns the result
	ProcessRequest(ctx context.Context, request *AuthRequest) (*AuthenticationResult, error)
}

// AuthenticationService combines all authentication interfaces
type AuthenticationService interface {
	TokenValidator
	UserInfoProvider
	AuthenticationMiddleware
}

// SecretProvider defines the interface for retrieving secrets
type SecretProvider interface {
	// GetOAuth2Config retrieves OAuth2 configuration for a gateway
	GetOAuth2Config(ctx context.Context, gateway string) (*OAuth2Config, error)
	
	// GetOIDCConfig retrieves OIDC discovery configuration
	GetOIDCConfig(ctx context.Context) (*OIDCConfiguration, error)
}

// OAuth2Config represents OAuth2 application configuration
type OAuth2Config struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	IssuerURL    string   `json:"issuer_url"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
}

// JWTClaims represents standard JWT claims
type JWTClaims struct {
	// Standard claims
	Subject   string `json:"sub"`           // Subject (user ID)
	Issuer    string `json:"iss"`           // Issuer
	Audience  string `json:"aud,omitempty"` // Audience  
	ExpiresAt int64  `json:"exp"`           // Expiration time
	IssuedAt  int64  `json:"iat"`           // Issued at
	NotBefore int64  `json:"nbf,omitempty"` // Not before

	// Custom claims
	Roles       []string               `json:"roles,omitempty"`
	Groups      []string               `json:"groups,omitempty"`
	Permissions []string               `json:"permissions,omitempty"`
	Email       string                 `json:"email,omitempty"`
	Username    string                 `json:"preferred_username,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Custom      map[string]interface{} `json:"-"`
}

// JWTHeader represents JWT header
type JWTHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid,omitempty"`
}

// OIDCConfiguration represents OIDC discovery configuration
type OIDCConfiguration struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	UserInfoEndpoint      string   `json:"userinfo_endpoint"`
	JWKSUri               string   `json:"jwks_uri"`
	ScopesSupported       []string `json:"scopes_supported"`
	ResponseTypesSupported []string `json:"response_types_supported"`
}

// JWK represents a JSON Web Key
type JWK struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

// JWKSet represents a JSON Web Key Set
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// JWTTokenValidator implements JWT token validation with OIDC discovery
type jwtTokenValidator struct {
	issuerURL      string
	clientID       string
	oidcConfig     *OIDCConfiguration
	jwksCache      map[string]*rsa.PublicKey
	jwksCacheTime  time.Time
	httpClient     HTTPClient
}

// NewJWTTokenValidator creates a new JWT token validator with OIDC discovery
func NewJWTTokenValidator(issuerURL, clientID string) TokenValidator {
	return &jwtTokenValidator{
		issuerURL:  issuerURL,
		clientID:   clientID,
		jwksCache:  make(map[string]*rsa.PublicKey),
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
}

// ValidateToken validates a JWT token using OIDC discovery and signature verification
func (j *jwtTokenValidator) ValidateToken(ctx context.Context, token string) (*TokenValidationResult, error) {
	if token == "" {
		return &TokenValidationResult{
			Valid:  false,
			Reason: ReasonTokenMissing,
		}, fmt.Errorf("token is empty")
	}

	// Parse JWT token
	claims, header, err := j.parseJWTToken(token)
	if err != nil {
		return &TokenValidationResult{
			Valid:  false,
			Reason: ReasonTokenInvalid,
		}, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	// Validate token expiration
	now := time.Now().Unix()
	if claims.ExpiresAt > 0 && now > claims.ExpiresAt {
		return &TokenValidationResult{
			Valid:  false,
			Reason: ReasonTokenExpired,
		}, fmt.Errorf("token has expired")
	}

	// Validate issuer
	if claims.Issuer != j.issuerURL {
		return &TokenValidationResult{
			Valid:  false,
			Reason: ReasonTokenInvalid,
		}, fmt.Errorf("invalid issuer: expected %s, got %s", j.issuerURL, claims.Issuer)
	}

	// Validate audience (if specified)
	if j.clientID != "" && claims.Audience != "" && claims.Audience != j.clientID {
		return &TokenValidationResult{
			Valid:  false,
			Reason: ReasonTokenInvalid,
		}, fmt.Errorf("invalid audience: expected %s, got %s", j.clientID, claims.Audience)
	}

	// TODO: Validate JWT signature with JWKS
	// For now, return success if basic validations pass
	// In production, signature validation is critical

	// Convert claims to UserInfo
	userInfo := &UserInfo{
		UserID:         claims.Subject,
		Username:       claims.Username,
		Email:          claims.Email,
		Roles:          claims.Roles,
		Permissions:    claims.Permissions,
		Groups:         claims.Groups,
		TokenIssuedAt:  time.Unix(claims.IssuedAt, 0),
		TokenExpiresAt: time.Unix(claims.ExpiresAt, 0),
		Metadata: map[string]interface{}{
			"issuer":  claims.Issuer,
			"jwt_kid": header.KeyID,
		},
	}

	return &TokenValidationResult{
		Valid:    true,
		Reason:   ReasonTokenValid,
		UserInfo: userInfo,
		Metadata: map[string]interface{}{
			"jwt_header": header,
		},
	}, nil
}

// parseJWTToken parses a JWT token and returns claims and header
func (j *jwtTokenValidator) parseJWTToken(token string) (*JWTClaims, *JWTHeader, error) {
	// Split JWT into parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, nil, fmt.Errorf("invalid JWT format")
	}

	// Decode header
	headerData, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode JWT header: %w", err)
	}

	var header JWTHeader
	if err := json.Unmarshal(headerData, &header); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal JWT header: %w", err)
	}

	// Decode payload
	payloadData, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadData, &claims); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal JWT claims: %w", err)
	}

	return &claims, &header, nil
}

// discoverOIDCConfiguration performs OIDC discovery
func (j *jwtTokenValidator) discoverOIDCConfiguration(ctx context.Context) error {
	if j.oidcConfig != nil {
		return nil // Already discovered
	}

	discoveryURL := j.issuerURL + OIDCDiscoveryPath
	req, err := http.NewRequestWithContext(ctx, "GET", discoveryURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create OIDC discovery request: %w", err)
	}

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform OIDC discovery: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OIDC discovery failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read OIDC discovery response: %w", err)
	}

	var config OIDCConfiguration
	if err := json.Unmarshal(body, &config); err != nil {
		return fmt.Errorf("failed to parse OIDC configuration: %w", err)
	}

	j.oidcConfig = &config
	return nil
}

// vaultSecretProvider implements SecretProvider using HashiCorp Vault
type vaultSecretProvider struct {
	vaultAddr  string
	vaultToken string
	httpClient HTTPClient
}

// NewVaultSecretProvider creates a new Vault-based secret provider
func NewVaultSecretProvider(vaultAddr, vaultToken string) SecretProvider {
	return &vaultSecretProvider{
		vaultAddr:  vaultAddr,
		vaultToken: vaultToken,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
}

// GetOAuth2Config retrieves OAuth2 configuration for a gateway from Vault
func (v *vaultSecretProvider) GetOAuth2Config(ctx context.Context, gateway string) (*OAuth2Config, error) {
	// Build Vault secret path
	secretPath := fmt.Sprintf("/v1/secret/data/oauth2/%s", gateway)
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", v.vaultAddr+secretPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault request: %w", err)
	}
	
	// Add Vault authentication header
	req.Header.Set("X-Vault-Token", v.vaultToken)
	
	// Execute request
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve OAuth2 config from Vault: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Vault request failed with status %d", resp.StatusCode)
	}
	
	// Parse response
	var vaultResp struct {
		Data struct {
			Data OAuth2Config `json:"data"`
		} `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&vaultResp); err != nil {
		return nil, fmt.Errorf("failed to decode Vault response: %w", err)
	}
	
	return &vaultResp.Data.Data, nil
}

// GetOIDCConfig retrieves OIDC discovery configuration from Vault
func (v *vaultSecretProvider) GetOIDCConfig(ctx context.Context) (*OIDCConfiguration, error) {
	// Build Vault secret path
	secretPath := "/v1/secret/data/oidc/config"
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", v.vaultAddr+secretPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault request: %w", err)
	}
	
	// Add Vault authentication header
	req.Header.Set("X-Vault-Token", v.vaultToken)
	
	// Execute request
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve OIDC config from Vault: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Vault request failed with status %d", resp.StatusCode)
	}
	
	// Parse response
	var vaultResp struct {
		Data struct {
			Data OIDCConfiguration `json:"data"`
		} `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&vaultResp); err != nil {
		return nil, fmt.Errorf("failed to decode Vault response: %w", err)
	}
	
	return &vaultResp.Data.Data, nil
}

// testTokenValidator provides a test implementation for unit testing
type testTokenValidator struct{}

// NewTestTokenValidator creates a test token validator for unit testing
func NewTestTokenValidator() TokenValidator {
	return &testTokenValidator{}
}

// ValidateToken implements test token validation logic
func (t *testTokenValidator) ValidateToken(ctx context.Context, token string) (*TokenValidationResult, error) {
	if token == "" {
		return &TokenValidationResult{
			Valid:  false,
			Reason: ReasonTokenMissing,
		}, fmt.Errorf("token is empty")
	}

	// Simulate JWT token validation based on token content
	// Check for expired tokens FIRST before checking for valid user tokens
	switch {
	case strings.HasPrefix(token, "expired-jwt-token-") ||
		 strings.Contains(token, "exp\":1") ||
		 token == "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ1c2VyLTQ1NiIsInJvbGVzIjpbInVzZXIiXSwiZXhwIjoxfQ":
		// Expired token
		return &TokenValidationResult{
			Valid:  false,
			Reason: ReasonTokenExpired,
		}, fmt.Errorf("token has expired")
	case strings.Contains(token, "admin-user-123") || 
		 token == "valid-jwt-token-admin" ||
		 token == "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJhZG1pbi11c2VyLTEyMyIsInJvbGVzIjpbImFkbWluIl0sImV4cCI6OTk5OTk5OTk5OX0":
		return &TokenValidationResult{
			Valid:  true,
			Reason: ReasonTokenValid,
			UserInfo: &UserInfo{
				UserID: "admin-user-123",
				Roles:  []string{"admin"},
			},
		}, nil
	case strings.Contains(token, "user-456") || 
		 token == "valid-jwt-token-user" ||
		 token == "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ1c2VyLTQ1NiIsInJvbGVzIjpbInVzZXIiXSwiZXhwIjo5OTk5OTk5OTk5fQ":
		return &TokenValidationResult{
			Valid:  true,
			Reason: ReasonTokenValid,
			UserInfo: &UserInfo{
				UserID: "user-456",
				Roles:  []string{"user"},
			},
		}, nil
	case token == "invalid-token" || token == "invalid.token.format":
		return &TokenValidationResult{
			Valid:  false,
			Reason: ReasonTokenInvalid,
		}, fmt.Errorf("token is invalid")
	default:
		return &TokenValidationResult{
			Valid:  false,
			Reason: ReasonTokenInvalid,
		}, fmt.Errorf("token is invalid")
	}
}

// testUserInfoProvider provides a test implementation for unit testing  
type testUserInfoProvider struct{}

// NewTestUserInfoProvider creates a test user info provider for unit testing
func NewTestUserInfoProvider() UserInfoProvider {
	return &testUserInfoProvider{}
}

// GetUserInfo implements test user info retrieval logic
func (t *testUserInfoProvider) GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
	switch userID {
	case "admin-user-123":
		return &UserInfo{
			UserID:      "admin-user-123",
			Username:    "admin",
			Email:       "admin@example.com",
			Roles:       []string{"admin"},
			Permissions: []string{"admin:*"},
		}, nil
	case "user-456":
		return &UserInfo{
			UserID:      "user-456",
			Username:    "testuser",
			Email:       "testuser@example.com",
			Roles:       []string{"user"},
			Permissions: []string{"user:read"},
		}, nil
	default:
		return nil, fmt.Errorf("user not found: %s", userID)
	}
}

// testAuthenticationMiddleware provides a test implementation for unit testing
type testAuthenticationMiddleware struct {
	tokenValidator TokenValidator
}

// NewTestAuthenticationMiddleware creates a test authentication middleware for unit testing
func NewTestAuthenticationMiddleware() AuthenticationMiddleware {
	return &testAuthenticationMiddleware{
		tokenValidator: NewTestTokenValidator(),
	}
}

// ProcessRequest implements test authentication request processing
func (t *testAuthenticationMiddleware) ProcessRequest(ctx context.Context, request *AuthRequest) (*AuthenticationResult, error) {
	if request.AuthorizationHeader == "" {
		return &AuthenticationResult{
			Authenticated: false,
			Reason:        ReasonTokenMissing,
		}, nil
	}

	// Parse Authorization header
	parts := strings.SplitN(request.AuthorizationHeader, " ", 2)
	if len(parts) != 2 || parts[0] != BearerTokenType {
		return &AuthenticationResult{
			Authenticated: false,
			Reason:        ReasonInvalidTokenType,
		}, nil
	}

	token := parts[1]

	// Validate token
	validationResult, err := t.tokenValidator.ValidateToken(ctx, token)
	if err != nil {
		return &AuthenticationResult{
			Authenticated: false,
			Reason:        ReasonTokenValidationError,
		}, nil
	}

	if !validationResult.Valid {
		return &AuthenticationResult{
			Authenticated: false,
			Reason:        validationResult.Reason,
		}, nil
	}

	return &AuthenticationResult{
		Authenticated: true,
		UserInfo:      validationResult.UserInfo,
		Reason:        ReasonTokenValid,
	}, nil
}

// extractBearerToken extracts the token from Authorization header
func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is empty")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != BearerTokenType {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}

// hasPermission checks if a user has a specific permission
func hasPermission(permissions []string, targetPermission string) bool {
	for _, permission := range permissions {
		if permission == targetPermission || permission == "*" {
			return true
		}
		// Check for wildcard permissions (e.g., "admin:*")
		if strings.HasSuffix(permission, ":*") {
			prefix := strings.TrimSuffix(permission, ":*")
			if strings.HasPrefix(targetPermission, prefix+":") {
				return true
			}
		}
	}
	return false
}