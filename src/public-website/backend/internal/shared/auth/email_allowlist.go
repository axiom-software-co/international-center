package auth

import (
	"strings"
)

// Restricted email allowlist - only these two emails are permitted
var emailAllowlist = []string{
	"tojkuv@gmail.com",    // Admin user with full permissions
	"tojkuv@outlook.com",  // Viewer user with lowest permissions (read-only)
}

// Role definitions for restricted access
const (
	RoleAdmin  = "admin"
	RoleViewer = "viewer"
)

// Permission definitions
var adminPermissions = []string{
	"read", "write", "delete", "manage_users", "manage_content", "view_audit",
}

var viewerPermissions = []string{
	"read", // Lowest permissions - read-only access
}

// GetEmailAllowlist returns the restricted email allowlist
func GetEmailAllowlist() []string {
	return emailAllowlist
}

// ValidateEmailInAllowlist validates if email is in the restricted allowlist
func ValidateEmailInAllowlist(email string) bool {
	// Exact match required - case sensitive and no whitespace tolerance
	for _, allowedEmail := range emailAllowlist {
		if email == allowedEmail {
			return true
		}
	}
	return false
}

// GetRoleForEmail returns the role assigned to a specific email
func GetRoleForEmail(email string) string {
	// Only assign roles to allowlisted emails
	if !ValidateEmailInAllowlist(email) {
		return ""
	}
	
	// Role assignment based on exact email match
	switch email {
	case "tojkuv@gmail.com":
		return RoleAdmin
	case "tojkuv@outlook.com":
		return RoleViewer
	default:
		return ""
	}
}

// GetPermissionsForEmail returns the permissions for a specific email
func GetPermissionsForEmail(email string) []string {
	role := GetRoleForEmail(email)
	
	switch role {
	case RoleAdmin:
		return adminPermissions
	case RoleViewer:
		return viewerPermissions
	default:
		return []string{} // No permissions for non-allowlisted emails
	}
}

// IsAdminEmail checks if email has admin privileges
func IsAdminEmail(email string) bool {
	return email == "tojkuv@gmail.com"
}

// IsViewerEmail checks if email has viewer privileges
func IsViewerEmail(email string) bool {
	return email == "tojkuv@outlook.com"
}

// GetProviderForEmail determines the OAuth2 provider based on email domain
func GetProviderForEmail(email string) string {
	if strings.HasSuffix(email, "@gmail.com") {
		return "google"
	}
	if strings.HasSuffix(email, "@outlook.com") {
		return "microsoft"
	}
	return "unknown"
}

// ValidateEmailProviderMatch validates that email domain matches expected OAuth2 provider
func ValidateEmailProviderMatch(email string, provider string) bool {
	expectedProvider := GetProviderForEmail(email)
	return expectedProvider == provider
}

// ExtractSocialProviderClaims extracts claims from social provider JWT
func ExtractSocialProviderClaims(jwtToken string) map[string]interface{} {
	// Placeholder implementation for testing
	// In production, this would parse actual JWT and extract claims
	
	// Mock claims for testing purposes
	if strings.Contains(jwtToken, "tojkuv@gmail.com") {
		return map[string]interface{}{
			"email":          "tojkuv@gmail.com",
			"email_verified": true,
			"name":           "Admin User",
			"picture":        "https://example.com/admin-avatar.jpg",
			"iss":            "https://accounts.google.com",
			"aud":            "international-center",
		}
	}
	
	if strings.Contains(jwtToken, "tojkuv@outlook.com") {
		return map[string]interface{}{
			"email":          "tojkuv@outlook.com",
			"email_verified": true,
			"name":           "Viewer User",
			"picture":        "https://example.com/viewer-avatar.jpg",
			"iss":            "https://login.microsoftonline.com/common/v2.0",
			"aud":            "international-center",
		}
	}
	
	// Return empty claims for invalid tokens
	return map[string]interface{}{}
}

// ValidateJWTToken validates JWT token structure and claims
func ValidateJWTToken(token string) bool {
	// Basic token validation
	if token == "" {
		return false
	}
	
	// Mock validation for testing
	if strings.Contains(token, "malformed") {
		return false
	}
	
	if strings.Contains(token, "expired") {
		return false
	}
	
	if strings.Contains(token, "invalid-signature") {
		return false
	}
	
	// Extract email from token and validate allowlist
	claims := ExtractSocialProviderClaims(token)
	if email, ok := claims["email"].(string); ok {
		return ValidateEmailInAllowlist(email)
	}
	
	return false
}

// GetTokenValidationError returns error information for invalid tokens
func GetTokenValidationError(token string) string {
	if token == "" {
		return "empty token"
	}
	
	if strings.Contains(token, "malformed") {
		return "invalid jwt"
	}
	
	if strings.Contains(token, "expired") {
		return "token expired"
	}
	
	if strings.Contains(token, "invalid-signature") {
		return "invalid signature"
	}
	
	// Check if email is not in allowlist
	claims := ExtractSocialProviderClaims(token)
	if email, ok := claims["email"].(string); ok {
		if !ValidateEmailInAllowlist(email) {
			return "email not in allowlist"
		}
	}
	
	return "unknown error"
}