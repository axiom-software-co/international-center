package automation

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// IAMManager manages identity and access management with least privilege principles
type IAMManager struct {
	environment     string
	policies        map[string]*IAMPolicy
	roleDefinitions map[string]*IAMRoleDefinition
	principalMappings map[string][]string
}

// IAMPolicy defines access policy with least privilege principles
type IAMPolicy struct {
	Name         string
	Description  string
	Environment  string
	Actions      []string
	Resources    []string
	Conditions   []IAMCondition
	MaxDuration  time.Duration
	RequiresMFA  bool
}

// IAMRoleDefinition defines role with specific permissions
type IAMRoleDefinition struct {
	Name               string
	Description        string
	Environment        string
	Policies           []string
	MaxSessionDuration time.Duration
	TrustedPrincipals  []string
	RequiresMFA        bool
	Tags               map[string]string
}

// IAMCondition defines conditional access requirements
type IAMCondition struct {
	Type      IAMConditionType
	Field     string
	Operator  string
	Values    []string
}

// IAMConditionType defines types of IAM conditions
type IAMConditionType string

const (
	IAMConditionTypeSourceIP       IAMConditionType = "source_ip"
	IAMConditionTypeTimeOfDay      IAMConditionType = "time_of_day"
	IAMConditionTypeMFAAge         IAMConditionType = "mfa_age"
	IAMConditionTypeEnvironment    IAMConditionType = "environment"
	IAMConditionTypeResourceAccess IAMConditionType = "resource_access"
)

// NewIAMManager creates new IAM manager with environment-specific policies
func NewIAMManager(environment string) *IAMManager {
	iam := &IAMManager{
		environment:       environment,
		policies:         make(map[string]*IAMPolicy),
		roleDefinitions:  make(map[string]*IAMRoleDefinition),
		principalMappings: make(map[string][]string),
	}
	
	// Configure environment-specific policies
	iam.configureDefaultPolicies()
	iam.configureDefaultRoles()
	
	return iam
}

// configureDefaultPolicies creates environment-specific IAM policies
func (iam *IAMManager) configureDefaultPolicies() {
	// Database access policies
	iam.policies["database-read"] = &IAMPolicy{
		Name:         fmt.Sprintf("%s-database-read", iam.environment),
		Description:  "Read-only access to database resources",
		Environment:  iam.environment,
		Actions:      []string{"Microsoft.Sql/servers/databases/read"},
		Resources:    []string{fmt.Sprintf("/subscriptions/*/resourceGroups/%s-*/providers/Microsoft.Sql/servers/*", iam.environment)},
		MaxDuration:  8 * time.Hour,
		RequiresMFA:  iam.environment == "production",
		Conditions: []IAMCondition{
			{
				Type:     IAMConditionTypeEnvironment,
				Field:    "Environment",
				Operator: "StringEquals",
				Values:   []string{iam.environment},
			},
		},
	}

	iam.policies["database-write"] = &IAMPolicy{
		Name:         fmt.Sprintf("%s-database-write", iam.environment),
		Description:  "Write access to database resources",
		Environment:  iam.environment,
		Actions: []string{
			"Microsoft.Sql/servers/databases/read",
			"Microsoft.Sql/servers/databases/write",
		},
		Resources:   []string{fmt.Sprintf("/subscriptions/*/resourceGroups/%s-*/providers/Microsoft.Sql/servers/*", iam.environment)},
		MaxDuration: 4 * time.Hour,
		RequiresMFA: true,
		Conditions: []IAMCondition{
			{
				Type:     IAMConditionTypeEnvironment,
				Field:    "Environment",
				Operator: "StringEquals",
				Values:   []string{iam.environment},
			},
			{
				Type:     IAMConditionTypeMFAAge,
				Field:    "MultiFactorAuthAge",
				Operator: "NumericLessThan",
				Values:   []string{"3600"}, // 1 hour
			},
		},
	}

	// Storage access policies
	iam.policies["storage-read"] = &IAMPolicy{
		Name:         fmt.Sprintf("%s-storage-read", iam.environment),
		Description:  "Read-only access to storage resources",
		Environment:  iam.environment,
		Actions:      []string{"Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read"},
		Resources:    []string{fmt.Sprintf("/subscriptions/*/resourceGroups/%s-*/providers/Microsoft.Storage/storageAccounts/*", iam.environment)},
		MaxDuration:  8 * time.Hour,
		RequiresMFA:  iam.environment == "production",
	}

	iam.policies["storage-write"] = &IAMPolicy{
		Name:         fmt.Sprintf("%s-storage-write", iam.environment),
		Description:  "Write access to storage resources", 
		Environment:  iam.environment,
		Actions: []string{
			"Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read",
			"Microsoft.Storage/storageAccounts/blobServices/containers/blobs/write",
		},
		Resources:   []string{fmt.Sprintf("/subscriptions/*/resourceGroups/%s-*/providers/Microsoft.Storage/storageAccounts/*", iam.environment)},
		MaxDuration: 2 * time.Hour,
		RequiresMFA: true,
	}

	// Vault access policies with enhanced security for production
	vaultPolicy := &IAMPolicy{
		Name:         fmt.Sprintf("%s-vault-access", iam.environment),
		Description:  "Access to vault resources",
		Environment:  iam.environment,
		Actions:      []string{"Microsoft.KeyVault/vaults/secrets/read"},
		Resources:    []string{fmt.Sprintf("/subscriptions/*/resourceGroups/%s-*/providers/Microsoft.KeyVault/vaults/*", iam.environment)},
		MaxDuration:  4 * time.Hour,
		RequiresMFA:  true,
		Conditions: []IAMCondition{
			{
				Type:     IAMConditionTypeEnvironment,
				Field:    "Environment", 
				Operator: "StringEquals",
				Values:   []string{iam.environment},
			},
		},
	}

	// Production vault requires additional restrictions
	if iam.environment == "production" {
		vaultPolicy.MaxDuration = 1 * time.Hour
		vaultPolicy.Conditions = append(vaultPolicy.Conditions, IAMCondition{
			Type:     IAMConditionTypeTimeOfDay,
			Field:    "DateTimeUtc",
			Operator: "DateTimeGreaterThan",
			Values:   []string{"08:00:00Z"},
		}, IAMCondition{
			Type:     IAMConditionTypeTimeOfDay,
			Field:    "DateTimeUtc", 
			Operator: "DateTimeLessThan",
			Values:   []string{"18:00:00Z"},
		})
	}

	iam.policies["vault-access"] = vaultPolicy
}

// configureDefaultRoles creates environment-specific IAM roles
func (iam *IAMManager) configureDefaultRoles() {
	// Developer role - limited access based on environment
	devPolicies := []string{"database-read", "storage-read"}
	if iam.environment != "production" {
		devPolicies = append(devPolicies, "database-write", "storage-write")
	}

	iam.roleDefinitions["developer"] = &IAMRoleDefinition{
		Name:               fmt.Sprintf("%s-developer", iam.environment),
		Description:        fmt.Sprintf("Developer role for %s environment", iam.environment),
		Environment:        iam.environment,
		Policies:           devPolicies,
		MaxSessionDuration: 8 * time.Hour,
		TrustedPrincipals:  []string{"developers-group"},
		RequiresMFA:        iam.environment == "production",
		Tags: map[string]string{
			"Environment": iam.environment,
			"Role":        "Developer",
			"AccessLevel": "Limited",
		},
	}

	// Operations role - elevated access with strict controls
	opsPolicies := []string{"database-read", "database-write", "storage-read", "storage-write"}
	if iam.environment == "production" {
		opsPolicies = append(opsPolicies, "vault-access")
	}

	iam.roleDefinitions["operations"] = &IAMRoleDefinition{
		Name:               fmt.Sprintf("%s-operations", iam.environment),
		Description:        fmt.Sprintf("Operations role for %s environment", iam.environment),
		Environment:        iam.environment,
		Policies:           opsPolicies,
		MaxSessionDuration: 4 * time.Hour,
		TrustedPrincipals:  []string{"operations-group"},
		RequiresMFA:        true,
		Tags: map[string]string{
			"Environment": iam.environment,
			"Role":        "Operations",
			"AccessLevel": "Elevated",
		},
	}

	// Admin role - full access with maximum restrictions for production
	adminSessionDuration := 8 * time.Hour
	if iam.environment == "production" {
		adminSessionDuration = 2 * time.Hour
	}

	iam.roleDefinitions["admin"] = &IAMRoleDefinition{
		Name:               fmt.Sprintf("%s-admin", iam.environment),
		Description:        fmt.Sprintf("Administrator role for %s environment", iam.environment),
		Environment:        iam.environment,
		Policies:           []string{"database-read", "database-write", "storage-read", "storage-write", "vault-access"},
		MaxSessionDuration: adminSessionDuration,
		TrustedPrincipals:  []string{"admins-group"},
		RequiresMFA:        true,
		Tags: map[string]string{
			"Environment": iam.environment,
			"Role":        "Administrator",
			"AccessLevel": "Full",
		},
	}

	// Service role for automated systems
	iam.roleDefinitions["service"] = &IAMRoleDefinition{
		Name:               fmt.Sprintf("%s-service", iam.environment),
		Description:        fmt.Sprintf("Service role for automated systems in %s", iam.environment),
		Environment:        iam.environment,
		Policies:           []string{"database-read", "database-write", "storage-read", "storage-write"},
		MaxSessionDuration: 24 * time.Hour,
		TrustedPrincipals:  []string{fmt.Sprintf("%s-service-principal", iam.environment)},
		RequiresMFA:        false, // Services use certificates instead
		Tags: map[string]string{
			"Environment": iam.environment,
			"Role":        "Service",
			"AccessLevel": "Automated",
		},
	}
}

// ValidateAccess validates if access request meets policy requirements
func (iam *IAMManager) ValidateAccess(ctx context.Context, request *AccessRequest) (*AccessValidationResult, error) {
	result := &AccessValidationResult{
		RequestID:   request.ID,
		Granted:     false,
		Reason:      "",
		Conditions:  []string{},
		MaxDuration: 0,
	}

	// Check if role exists
	role, exists := iam.roleDefinitions[request.Role]
	if !exists {
		result.Reason = fmt.Sprintf("Role %s not found in environment %s", request.Role, iam.environment)
		return result, nil
	}

	// Validate environment isolation
	if role.Environment != iam.environment {
		result.Reason = fmt.Sprintf("Role %s not valid for environment %s", request.Role, iam.environment)
		return result, nil
	}

	// Check principal authorization
	if !iam.isPrincipalAuthorized(request.Principal, role.TrustedPrincipals) {
		result.Reason = fmt.Sprintf("Principal %s not authorized for role %s", request.Principal, request.Role)
		return result, nil
	}

	// Validate policy conditions
	for _, policyName := range role.Policies {
		policy, policyExists := iam.policies[policyName]
		if !policyExists {
			result.Reason = fmt.Sprintf("Policy %s not found", policyName)
			return result, nil
		}

		// Check policy conditions
		for _, condition := range policy.Conditions {
			if !iam.evaluateCondition(ctx, condition, request) {
				result.Reason = fmt.Sprintf("Policy condition failed: %s %s %s", condition.Field, condition.Operator, strings.Join(condition.Values, ","))
				return result, nil
			}
		}

		// Track maximum duration
		if result.MaxDuration == 0 || policy.MaxDuration < result.MaxDuration {
			result.MaxDuration = policy.MaxDuration
		}
	}

	// MFA validation
	if role.RequiresMFA && !request.MFAVerified {
		result.Reason = "Multi-factor authentication required"
		return result, nil
	}

	// Access granted
	result.Granted = true
	result.Reason = "Access granted based on policy validation"
	result.Conditions = iam.generateAccessConditions(role)

	return result, nil
}

// isPrincipalAuthorized checks if principal is authorized for role
func (iam *IAMManager) isPrincipalAuthorized(principal string, trustedPrincipals []string) bool {
	for _, trusted := range trustedPrincipals {
		if principal == trusted {
			return true
		}
		// Check group membership if applicable
		if groups, exists := iam.principalMappings[principal]; exists {
			for _, group := range groups {
				if group == trusted {
					return true
				}
			}
		}
	}
	return false
}

// evaluateCondition evaluates IAM condition
func (iam *IAMManager) evaluateCondition(ctx context.Context, condition IAMCondition, request *AccessRequest) bool {
	switch condition.Type {
	case IAMConditionTypeEnvironment:
		return condition.Values[0] == iam.environment
	case IAMConditionTypeMFAAge:
		if !request.MFAVerified {
			return false
		}
		// Check MFA age (simplified implementation)
		return true
	case IAMConditionTypeSourceIP:
		// Validate source IP (would integrate with actual IP validation)
		return true
	case IAMConditionTypeTimeOfDay:
		// Validate time restrictions (simplified implementation)
		return true
	default:
		return false
	}
}

// generateAccessConditions generates conditions for granted access
func (iam *IAMManager) generateAccessConditions(role *IAMRoleDefinition) []string {
	conditions := []string{
		fmt.Sprintf("Environment: %s", role.Environment),
		fmt.Sprintf("MaxSessionDuration: %v", role.MaxSessionDuration),
	}

	if role.RequiresMFA {
		conditions = append(conditions, "MFA: Required")
	}

	return conditions
}

// AccessRequest represents access request for validation
type AccessRequest struct {
	ID          string
	Principal   string
	Role        string
	Resource    string
	Action      string
	Environment string
	SourceIP    string
	MFAVerified bool
	Timestamp   time.Time
}

// AccessValidationResult represents validation result
type AccessValidationResult struct {
	RequestID   string
	Granted     bool
	Reason      string
	Conditions  []string
	MaxDuration time.Duration
}

// GetRoleDefinitions returns role definitions for environment
func (iam *IAMManager) GetRoleDefinitions() map[string]*IAMRoleDefinition {
	return iam.roleDefinitions
}

// GetPolicies returns policies for environment
func (iam *IAMManager) GetPolicies() map[string]*IAMPolicy {
	return iam.policies
}

// AddPolicy adds custom policy with validation
func (iam *IAMManager) AddPolicy(policy *IAMPolicy) error {
	if policy.Environment != iam.environment {
		return fmt.Errorf("policy environment %s does not match manager environment %s", policy.Environment, iam.environment)
	}

	if policy.MaxDuration > 24*time.Hour {
		return fmt.Errorf("policy max duration cannot exceed 24 hours")
	}

	if iam.environment == "production" && policy.MaxDuration > 4*time.Hour {
		return fmt.Errorf("production policies cannot exceed 4 hours duration")
	}

	iam.policies[policy.Name] = policy
	return nil
}

// AddRoleDefinition adds custom role definition with validation
func (iam *IAMManager) AddRoleDefinition(role *IAMRoleDefinition) error {
	if role.Environment != iam.environment {
		return fmt.Errorf("role environment %s does not match manager environment %s", role.Environment, iam.environment)
	}

	// Validate all policies exist
	for _, policyName := range role.Policies {
		if _, exists := iam.policies[policyName]; !exists {
			return fmt.Errorf("policy %s referenced by role does not exist", policyName)
		}
	}

	if iam.environment == "production" && role.MaxSessionDuration > 8*time.Hour {
		return fmt.Errorf("production roles cannot exceed 8 hours session duration")
	}

	iam.roleDefinitions[role.Name] = role
	return nil
}

// UpdatePrincipalMappings updates group memberships for principals
func (iam *IAMManager) UpdatePrincipalMappings(principal string, groups []string) {
	iam.principalMappings[principal] = groups
}