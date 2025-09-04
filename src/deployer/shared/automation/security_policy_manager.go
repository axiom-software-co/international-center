package automation

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// SecurityPolicyManager enforces security policies and compliance requirements
type SecurityPolicyManager struct {
	environment         string
	policies           map[string]*SecurityPolicy
	complianceRules    []ComplianceRule
	violationHandlers  map[SecurityViolationType]ViolationHandler
	auditLogger       *SecurityAuditLogger
}

// SecurityPolicy defines security policy with enforcement rules
type SecurityPolicy struct {
	Name            string
	Description     string
	Environment     string
	Scope          SecurityScope
	Rules          []SecurityRule
	Enforcement    EnforcementLevel
	Created        time.Time
	LastModified   time.Time
}

// SecurityRule defines specific security requirement
type SecurityRule struct {
	ID          string
	Name        string
	Description string
	RuleType    SecurityRuleType
	Parameters  map[string]interface{}
	Severity    SecuritySeverity
	Enabled     bool
}

// SecurityScope defines policy application scope
type SecurityScope string

const (
	SecurityScopeGlobal      SecurityScope = "global"
	SecurityScopeEnvironment SecurityScope = "environment"
	SecurityScopeResource    SecurityScope = "resource"
	SecurityScopeComponent   SecurityScope = "component"
)

// SecurityRuleType defines types of security rules
type SecurityRuleType string

const (
	SecurityRuleTypeNetworkAccess    SecurityRuleType = "network_access"
	SecurityRuleTypeDataEncryption   SecurityRuleType = "data_encryption"
	SecurityRuleTypeAccessControl    SecurityRuleType = "access_control"
	SecurityRuleTypeAuditLogging     SecurityRuleType = "audit_logging"
	SecurityRuleTypeSecretManagement SecurityRuleType = "secret_management"
	SecurityRuleTypeCompliance       SecurityRuleType = "compliance"
)

// SecuritySeverity defines severity levels
type SecuritySeverity string

const (
	SecuritySeverityLow      SecuritySeverity = "low"
	SecuritySeverityMedium   SecuritySeverity = "medium"
	SecuritySeverityHigh     SecuritySeverity = "high"
	SecuritySeverityCritical SecuritySeverity = "critical"
)

// EnforcementLevel defines policy enforcement levels
type EnforcementLevel string

const (
	EnforcementLevelWarn    EnforcementLevel = "warn"
	EnforcementLevelBlock   EnforcementLevel = "block"
	EnforcementLevelAudit   EnforcementLevel = "audit"
)

// SecurityViolationType defines types of security violations
type SecurityViolationType string

const (
	SecurityViolationUnauthorizedAccess SecurityViolationType = "unauthorized_access"
	SecurityViolationDataExposure       SecurityViolationType = "data_exposure"
	SecurityViolationPolicyViolation    SecurityViolationType = "policy_violation"
	SecurityViolationComplianceFailure  SecurityViolationType = "compliance_failure"
)

// ViolationHandler handles security violations
type ViolationHandler interface {
	HandleViolation(ctx context.Context, violation *SecurityViolation) error
}

// SecurityViolation represents security violation
type SecurityViolation struct {
	ID          string
	Type        SecurityViolationType
	Severity    SecuritySeverity
	Description string
	Environment string
	Resource    string
	Principal   string
	Timestamp   time.Time
	Details     map[string]interface{}
}

// SecurityAuditLogger logs security events for compliance
type SecurityAuditLogger struct {
	environment string
	logEntries  []SecurityLogEntry
}

// SecurityLogEntry represents security audit log entry
type SecurityLogEntry struct {
	Timestamp   time.Time
	Environment string
	EventType   string
	Principal   string
	Resource    string
	Action      string
	Result      string
	Details     map[string]interface{}
}

// NewSecurityPolicyManager creates security policy manager
func NewSecurityPolicyManager(environment string) *SecurityPolicyManager {
	spm := &SecurityPolicyManager{
		environment:       environment,
		policies:         make(map[string]*SecurityPolicy),
		complianceRules:  []ComplianceRule{},
		violationHandlers: make(map[SecurityViolationType]ViolationHandler),
		auditLogger:      NewSecurityAuditLogger(environment),
	}

	// Configure default policies
	spm.configureDefaultPolicies()
	spm.configureDefaultViolationHandlers()

	return spm
}

// configureDefaultPolicies creates environment-specific security policies
func (spm *SecurityPolicyManager) configureDefaultPolicies() {
	// Network security policy
	networkPolicy := &SecurityPolicy{
		Name:         fmt.Sprintf("%s-network-security", spm.environment),
		Description:  "Network security policy with environment isolation",
		Environment:  spm.environment,
		Scope:        SecurityScopeEnvironment,
		Enforcement:  EnforcementLevelBlock,
		Created:      time.Now(),
		LastModified: time.Now(),
		Rules: []SecurityRule{
			{
				ID:          "network-001",
				Name:        "Environment Network Isolation",
				Description: "Network resources must be isolated by environment",
				RuleType:    SecurityRuleTypeNetworkAccess,
				Severity:    SecuritySeverityHigh,
				Enabled:     true,
				Parameters: map[string]interface{}{
					"allowed_environments": []string{spm.environment},
					"cross_environment_access": false,
				},
			},
			{
				ID:          "network-002", 
				Name:        "TLS Encryption Required",
				Description: "All network traffic must use TLS 1.2 or higher",
				RuleType:    SecurityRuleTypeDataEncryption,
				Severity:    SecuritySeverityHigh,
				Enabled:     true,
				Parameters: map[string]interface{}{
					"min_tls_version": "1.2",
					"require_tls": true,
				},
			},
		},
	}

	// Data encryption policy  
	encryptionPolicy := &SecurityPolicy{
		Name:         fmt.Sprintf("%s-data-encryption", spm.environment),
		Description:  "Data encryption requirements",
		Environment:  spm.environment,
		Scope:        SecurityScopeGlobal,
		Enforcement:  EnforcementLevelBlock,
		Created:      time.Now(),
		LastModified: time.Now(),
		Rules: []SecurityRule{
			{
				ID:          "encryption-001",
				Name:        "Data at Rest Encryption",
				Description: "All data must be encrypted at rest",
				RuleType:    SecurityRuleTypeDataEncryption,
				Severity:    SecuritySeverityCritical,
				Enabled:     true,
				Parameters: map[string]interface{}{
					"encryption_required": true,
					"min_key_length": 256,
				},
			},
			{
				ID:          "encryption-002",
				Name:        "Data in Transit Encryption",
				Description: "All data in transit must be encrypted",
				RuleType:    SecurityRuleTypeDataEncryption,
				Severity:    SecuritySeverityCritical,
				Enabled:     true,
				Parameters: map[string]interface{}{
					"transit_encryption_required": true,
				},
			},
		},
	}

	// Access control policy with environment-specific restrictions
	accessControlPolicy := &SecurityPolicy{
		Name:         fmt.Sprintf("%s-access-control", spm.environment),
		Description:  "Access control and least privilege enforcement",
		Environment:  spm.environment,
		Scope:        SecurityScopeEnvironment,
		Enforcement:  EnforcementLevelBlock,
		Created:      time.Now(),
		LastModified: time.Now(),
		Rules: []SecurityRule{
			{
				ID:          "access-001",
				Name:        "Least Privilege Access",
				Description: "Access must follow least privilege principles",
				RuleType:    SecurityRuleTypeAccessControl,
				Severity:    SecuritySeverityHigh,
				Enabled:     true,
				Parameters: map[string]interface{}{
					"least_privilege_required": true,
					"max_permissions": "read-write", // No admin by default
				},
			},
		},
	}

	// Production requires additional restrictions
	if spm.environment == "production" {
		accessControlPolicy.Rules = append(accessControlPolicy.Rules, SecurityRule{
			ID:          "access-002",
			Name:        "Production MFA Requirement",
			Description: "Multi-factor authentication required for production access",
			RuleType:    SecurityRuleTypeAccessControl,
			Severity:    SecuritySeverityCritical,
			Enabled:     true,
			Parameters: map[string]interface{}{
				"mfa_required": true,
				"mfa_max_age": 3600, // 1 hour
			},
		})
	}

	// Secret management policy
	secretPolicy := &SecurityPolicy{
		Name:         fmt.Sprintf("%s-secret-management", spm.environment),
		Description:  "Secret management and protection requirements",
		Environment:  spm.environment,
		Scope:        SecurityScopeGlobal,
		Enforcement:  EnforcementLevelBlock,
		Created:      time.Now(),
		LastModified: time.Now(),
		Rules: []SecurityRule{
			{
				ID:          "secret-001",
				Name:        "No Hardcoded Secrets",
				Description: "Secrets must not be hardcoded in configuration or code",
				RuleType:    SecurityRuleTypeSecretManagement,
				Severity:    SecuritySeverityCritical,
				Enabled:     true,
				Parameters: map[string]interface{}{
					"hardcoded_secrets_forbidden": true,
					"secret_patterns": []string{"password", "secret", "key", "token", "credential"},
				},
			},
			{
				ID:          "secret-002",
				Name:        "Secret Rotation Required",
				Description: "Secrets must be rotated regularly",
				RuleType:    SecurityRuleTypeSecretManagement,
				Severity:    SecuritySeverityMedium,
				Enabled:     true,
				Parameters: map[string]interface{}{
					"rotation_required": true,
					"max_age_days": 90,
				},
			},
		},
	}

	// Audit logging policy
	auditPolicy := &SecurityPolicy{
		Name:         fmt.Sprintf("%s-audit-logging", spm.environment),
		Description:  "Audit logging requirements for compliance",
		Environment:  spm.environment,
		Scope:        SecurityScopeGlobal,
		Enforcement:  EnforcementLevelAudit,
		Created:      time.Now(),
		LastModified: time.Now(),
		Rules: []SecurityRule{
			{
				ID:          "audit-001",
				Name:        "Security Event Logging",
				Description: "All security events must be logged",
				RuleType:    SecurityRuleTypeAuditLogging,
				Severity:    SecuritySeverityHigh,
				Enabled:     true,
				Parameters: map[string]interface{}{
					"log_all_access": true,
					"log_failures": true,
					"retention_days": 2555, // 7 years for compliance
				},
			},
		},
	}

	// Store policies
	spm.policies["network-security"] = networkPolicy
	spm.policies["data-encryption"] = encryptionPolicy
	spm.policies["access-control"] = accessControlPolicy
	spm.policies["secret-management"] = secretPolicy
	spm.policies["audit-logging"] = auditPolicy
}

// configureDefaultViolationHandlers sets up violation handling
func (spm *SecurityPolicyManager) configureDefaultViolationHandlers() {
	spm.violationHandlers[SecurityViolationUnauthorizedAccess] = &DefaultViolationHandler{}
	spm.violationHandlers[SecurityViolationDataExposure] = &DefaultViolationHandler{}
	spm.violationHandlers[SecurityViolationPolicyViolation] = &DefaultViolationHandler{}
	spm.violationHandlers[SecurityViolationComplianceFailure] = &DefaultViolationHandler{}
}

// ValidateSecurity validates security compliance for resource
func (spm *SecurityPolicyManager) ValidateSecurity(ctx context.Context, request *SecurityValidationRequest) (*SecurityValidationResult, error) {
	result := &SecurityValidationResult{
		RequestID:    request.ID,
		Valid:        true,
		Violations:   []SecurityViolation{},
		Warnings:     []string{},
		Environment:  spm.environment,
		Timestamp:    time.Now(),
	}

	// Audit the validation request
	spm.auditLogger.LogSecurityEvent("security_validation", request.Principal, request.Resource, "validate", map[string]interface{}{
		"request_id": request.ID,
		"resource_type": request.ResourceType,
	})

	// Apply each policy
	for _, policy := range spm.policies {
		policyResult := spm.validatePolicy(ctx, policy, request)
		
		result.Violations = append(result.Violations, policyResult.Violations...)
		result.Warnings = append(result.Warnings, policyResult.Warnings...)

		// Check if any critical violations occurred
		for _, violation := range policyResult.Violations {
			if violation.Severity == SecuritySeverityCritical {
				result.Valid = false
			}
		}
	}

	return result, nil
}

// validatePolicy validates specific security policy
func (spm *SecurityPolicyManager) validatePolicy(ctx context.Context, policy *SecurityPolicy, request *SecurityValidationRequest) *SecurityValidationResult {
	result := &SecurityValidationResult{
		RequestID:   request.ID,
		Valid:       true,
		Violations:  []SecurityViolation{},
		Warnings:    []string{},
		Environment: spm.environment,
		Timestamp:   time.Now(),
	}

	// Apply policy rules
	for _, rule := range policy.Rules {
		if !rule.Enabled {
			continue
		}

		violation := spm.validateRule(ctx, rule, request)
		if violation != nil {
			if policy.Enforcement == EnforcementLevelBlock {
				result.Violations = append(result.Violations, *violation)
				if rule.Severity == SecuritySeverityCritical || rule.Severity == SecuritySeverityHigh {
					result.Valid = false
				}
			} else if policy.Enforcement == EnforcementLevelWarn {
				result.Warnings = append(result.Warnings, violation.Description)
			}
		}
	}

	return result
}

// validateRule validates specific security rule
func (spm *SecurityPolicyManager) validateRule(ctx context.Context, rule SecurityRule, request *SecurityValidationRequest) *SecurityViolation {
	switch rule.RuleType {
	case SecurityRuleTypeNetworkAccess:
		return spm.validateNetworkAccess(rule, request)
	case SecurityRuleTypeDataEncryption:
		return spm.validateDataEncryption(rule, request)
	case SecurityRuleTypeAccessControl:
		return spm.validateAccessControl(rule, request)
	case SecurityRuleTypeSecretManagement:
		return spm.validateSecretManagement(rule, request)
	case SecurityRuleTypeAuditLogging:
		return spm.validateAuditLogging(rule, request)
	default:
		return nil
	}
}

// validateNetworkAccess validates network access rules
func (spm *SecurityPolicyManager) validateNetworkAccess(rule SecurityRule, request *SecurityValidationRequest) *SecurityViolation {
	// Check environment isolation
	if allowedEnvs, exists := rule.Parameters["allowed_environments"].([]string); exists {
		environmentAllowed := false
		for _, env := range allowedEnvs {
			if env == spm.environment {
				environmentAllowed = true
				break
			}
		}
		
		if !environmentAllowed {
			return &SecurityViolation{
				ID:          fmt.Sprintf("network-violation-%d", time.Now().Unix()),
				Type:        SecurityViolationPolicyViolation,
				Severity:    rule.Severity,
				Description: fmt.Sprintf("Network access not allowed for environment %s", spm.environment),
				Environment: spm.environment,
				Resource:    request.Resource,
				Principal:   request.Principal,
				Timestamp:   time.Now(),
			}
		}
	}

	return nil
}

// validateDataEncryption validates encryption requirements
func (spm *SecurityPolicyManager) validateDataEncryption(rule SecurityRule, request *SecurityValidationRequest) *SecurityViolation {
	// Check if encryption is required and enabled
	if encryptionRequired, exists := rule.Parameters["encryption_required"].(bool); exists && encryptionRequired {
		// In real implementation, this would check actual resource encryption settings
		// For now, validate that encryption parameters are present
		if request.Configuration == nil {
			return &SecurityViolation{
				ID:          fmt.Sprintf("encryption-violation-%d", time.Now().Unix()),
				Type:        SecurityViolationPolicyViolation,
				Severity:    rule.Severity,
				Description: "Encryption configuration missing",
				Environment: spm.environment,
				Resource:    request.Resource,
				Principal:   request.Principal,
				Timestamp:   time.Now(),
			}
		}
	}

	return nil
}

// validateAccessControl validates access control rules
func (spm *SecurityPolicyManager) validateAccessControl(rule SecurityRule, request *SecurityValidationRequest) *SecurityViolation {
	// Check MFA requirements for production
	if mfaRequired, exists := rule.Parameters["mfa_required"].(bool); exists && mfaRequired {
		if !request.MFAVerified {
			return &SecurityViolation{
				ID:          fmt.Sprintf("access-violation-%d", time.Now().Unix()),
				Type:        SecurityViolationUnauthorizedAccess,
				Severity:    rule.Severity,
				Description: "Multi-factor authentication required",
				Environment: spm.environment,
				Resource:    request.Resource,
				Principal:   request.Principal,
				Timestamp:   time.Now(),
			}
		}
	}

	return nil
}

// validateSecretManagement validates secret management rules
func (spm *SecurityPolicyManager) validateSecretManagement(rule SecurityRule, request *SecurityValidationRequest) *SecurityViolation {
	// Check for hardcoded secrets
	if forbidden, exists := rule.Parameters["hardcoded_secrets_forbidden"].(bool); exists && forbidden {
		if patterns, patternsExist := rule.Parameters["secret_patterns"].([]string); patternsExist {
			configStr := fmt.Sprintf("%v", request.Configuration)
			for _, pattern := range patterns {
				matched, _ := regexp.MatchString(pattern, strings.ToLower(configStr))
				if matched {
					return &SecurityViolation{
						ID:          fmt.Sprintf("secret-violation-%d", time.Now().Unix()),
						Type:        SecurityViolationDataExposure,
						Severity:    rule.Severity,
						Description: fmt.Sprintf("Potential hardcoded secret detected: %s", pattern),
						Environment: spm.environment,
						Resource:    request.Resource,
						Principal:   request.Principal,
						Timestamp:   time.Now(),
					}
				}
			}
		}
	}

	return nil
}

// validateAuditLogging validates audit logging requirements
func (spm *SecurityPolicyManager) validateAuditLogging(rule SecurityRule, request *SecurityValidationRequest) *SecurityViolation {
	// Audit logging validation would check that proper logging is configured
	// This is more of an infrastructure configuration check
	return nil
}

// SecurityValidationRequest represents security validation request
type SecurityValidationRequest struct {
	ID            string
	Principal     string
	Resource      string
	ResourceType  string
	Action        string
	Environment   string
	Configuration map[string]interface{}
	MFAVerified   bool
	Timestamp     time.Time
}

// SecurityValidationResult represents validation result
type SecurityValidationResult struct {
	RequestID   string
	Valid       bool
	Violations  []SecurityViolation
	Warnings    []string
	Environment string
	Timestamp   time.Time
}

// DefaultViolationHandler provides default violation handling
type DefaultViolationHandler struct{}

func (dvh *DefaultViolationHandler) HandleViolation(ctx context.Context, violation *SecurityViolation) error {
	// Log violation
	fmt.Printf("Security violation detected: %s - %s\n", violation.Type, violation.Description)
	
	// In production, this would integrate with alerting systems
	return nil
}

// NewSecurityAuditLogger creates security audit logger
func NewSecurityAuditLogger(environment string) *SecurityAuditLogger {
	return &SecurityAuditLogger{
		environment: environment,
		logEntries:  []SecurityLogEntry{},
	}
}

// LogSecurityEvent logs security event for audit trail
func (sal *SecurityAuditLogger) LogSecurityEvent(eventType, principal, resource, action string, details map[string]interface{}) {
	entry := SecurityLogEntry{
		Timestamp:   time.Now(),
		Environment: sal.environment,
		EventType:   eventType,
		Principal:   principal,
		Resource:    resource,
		Action:      action,
		Result:      "success", // Would be determined by actual result
		Details:     details,
	}

	sal.logEntries = append(sal.logEntries, entry)
	
	// In production, this would send to external audit system
	fmt.Printf("SECURITY AUDIT [%s]: %s by %s on %s - %s\n", 
		sal.environment, eventType, principal, resource, action)
}

// GetAuditLog returns audit log entries
func (sal *SecurityAuditLogger) GetAuditLog() []SecurityLogEntry {
	return sal.logEntries
}