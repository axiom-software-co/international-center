package application

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-azure-native/sdk/v2/go/azurenative/app"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/axiom-software-co/international-center/src/deployer/internal/production/infrastructure"
)

type SecurityConfiguration struct {
	containerAppsStack *infrastructure.AzureProductionAppsStack
	vaultStack        *infrastructure.VaultProductionStack
	securityComponents map[string]*app.DaprComponent
	securityPolicies  map[string]*SecurityPolicy
	threatProtection  *AdvancedThreatProtection
	complianceRules   map[string]*ComplianceRule
}

type SecurityPolicy struct {
	Name                  string
	Type                  string
	EnforcementLevel      string
	Rules                []SecurityRule
	Exceptions           []SecurityException
	AuditLevel           string
	ComplianceFrameworks []string
}

type SecurityRule struct {
	ID          string
	Name        string
	Type        string
	Condition   string
	Action      string
	Severity    string
	Enabled     bool
	Description string
}

type SecurityException struct {
	ID            string
	ResourceType  string
	ResourceName  string
	Rule          string
	Reason        string
	ApprovedBy    string
	ExpiresAt     string
	AuditRequired bool
}

type AdvancedThreatProtection struct {
	Enabled                    bool
	RealTimeScanning          bool
	BehaviorAnalysis          bool
	AnomalyDetection          bool
	ThreatIntelligenceFeeds   []string
	AutomaticResponse         bool
	QuarantineThreshold       string
	NotificationEndpoints     []string
	IncidentEscalationRules   []EscalationRule
}

type EscalationRule struct {
	ThreatLevel     string
	ResponseTime    string
	NotificationList []string
	AutomaticActions []string
}

type ComplianceRule struct {
	Framework    string
	Requirement  string
	Control      string
	Implementation string
	ValidationMethod string
	Evidence     []string
	Status       string
	LastAudited  string
	NextAudit    string
}

func NewSecurityConfiguration(
	containerAppsStack *infrastructure.AzureProductionAppsStack,
	vaultStack *infrastructure.VaultProductionStack,
) *SecurityConfiguration {
	return &SecurityConfiguration{
		containerAppsStack: containerAppsStack,
		vaultStack:        vaultStack,
		securityComponents: make(map[string]*app.DaprComponent),
		securityPolicies:  make(map[string]*SecurityPolicy),
		complianceRules:   make(map[string]*ComplianceRule),
		threatProtection:  getAdvancedThreatProtectionConfig(),
	}
}

func (sc *SecurityConfiguration) Deploy(ctx *pulumi.Context) error {
	if err := sc.createSecurityMiddleware(ctx); err != nil {
		return fmt.Errorf("failed to create security middleware: %w", err)
	}

	if err := sc.configureSecurityPolicies(ctx); err != nil {
		return fmt.Errorf("failed to configure security policies: %w", err)
	}

	if err := sc.setupAdvancedThreatProtection(ctx); err != nil {
		return fmt.Errorf("failed to setup advanced threat protection: %w", err)
	}

	if err := sc.configureComplianceRules(ctx); err != nil {
		return fmt.Errorf("failed to configure compliance rules: %w", err)
	}

	return nil
}

func (sc *SecurityConfiguration) createSecurityMiddleware(ctx *pulumi.Context) error {
	if err := sc.createAuthenticationMiddleware(ctx); err != nil {
		return err
	}

	if err := sc.createAuthorizationMiddleware(ctx); err != nil {
		return err
	}

	if err := sc.createRateLimitingMiddleware(ctx); err != nil {
		return err
	}

	if err := sc.createSecurityHeadersMiddleware(ctx); err != nil {
		return err
	}

	if err := sc.createInputValidationMiddleware(ctx); err != nil {
		return err
	}

	if err := sc.createAuditLoggingMiddleware(ctx); err != nil {
		return err
	}

	return nil
}

func (sc *SecurityConfiguration) createAuthenticationMiddleware(ctx *pulumi.Context) error {
	component, err := app.NewDaprComponent(ctx, "production-auth-middleware", &app.DaprComponentArgs{
		ResourceGroupName:    sc.containerAppsStack.GetResourceGroup().Name,
		EnvironmentName:     sc.containerAppsStack.GetEnvironment().Name,
		ComponentName:       pulumi.String("authentication"),
		ComponentType:       pulumi.String("middleware.http.oauth2"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("clientId"),
				Value: pulumi.String(""), // From Azure AD App Registration
			},
			&app.DaprMetadataArgs{
				Name:      pulumi.String("clientSecret"),
				SecretRef: pulumi.String("oauth-client-secret"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("scopes"),
				Value: pulumi.String("openid profile email api.read api.write"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("authURL"),
				Value: pulumi.String("https://login.microsoftonline.com/common/oauth2/v2.0/authorize"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("tokenURL"),
				Value: pulumi.String("https://login.microsoftonline.com/common/oauth2/v2.0/token"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("redirectURL"),
				Value: pulumi.String("https://api.international-center.com/auth/callback"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("sessionEncryptionKey"),
				SecretRef: pulumi.String("jwt-signing-key"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("forceHTTPS"),
				Value: pulumi.String("true"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("cookieSecure"),
				Value: pulumi.String("true"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("cookieHttpOnly"),
				Value: pulumi.String("true"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("cookieSameSite"),
				Value: pulumi.String("Strict"),
			},
		},
		Secrets: app.SecretArray{
			&app.SecretArgs{
				Name:        pulumi.String("oauth-client-secret"),
				KeyVaultUrl: sc.vaultStack.GetVaultUri(),
				Identity:    pulumi.String("system"),
			},
			&app.SecretArgs{
				Name:        pulumi.String("jwt-signing-key"),
				KeyVaultUrl: sc.vaultStack.GetVaultUri(),
				Identity:    pulumi.String("system"),
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("public-gateway"),
			pulumi.String("admin-gateway"),
		},
	})
	if err != nil {
		return err
	}

	sc.securityComponents["authentication"] = component
	return nil
}

func (sc *SecurityConfiguration) createAuthorizationMiddleware(ctx *pulumi.Context) error {
	component, err := app.NewDaprComponent(ctx, "production-authz-middleware", &app.DaprComponentArgs{
		ResourceGroupName:    sc.containerAppsStack.GetResourceGroup().Name,
		EnvironmentName:     sc.containerAppsStack.GetEnvironment().Name,
		ComponentName:       pulumi.String("authorization"),
		ComponentType:       pulumi.String("middleware.http.opa"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("defaultStatus"),
				Value: pulumi.String("403"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("defaultMessage"),
				Value: pulumi.String("Forbidden: Access denied by authorization policy"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("includedHeaders"),
				Value: pulumi.String("Authorization,X-User-ID,X-User-Role,X-Correlation-ID"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("rego"),
				Value: pulumi.String(`
					package authz
					
					default allow = false
					
					# Allow public read operations
					allow {
						input.request.method == "GET"
						startswith(input.request.path, "/api/v1/services")
						not contains(input.request.path, "/admin")
					}
					
					# Allow authenticated users for standard operations
					allow {
						input.request.headers.Authorization
						input.request.headers["X-User-ID"]
						input.request.method in ["GET", "POST", "PUT", "PATCH"]
					}
					
					# Allow admin users for all operations
					allow {
						input.request.headers.Authorization
						input.request.headers["X-User-Role"] == "admin"
					}
					
					# Deny by default for security
				`),
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("public-gateway"),
			pulumi.String("admin-gateway"),
			pulumi.String("identity-api"),
			pulumi.String("content-api"),
			pulumi.String("services-api"),
		},
	})
	if err != nil {
		return err
	}

	sc.securityComponents["authorization"] = component
	return nil
}

func (sc *SecurityConfiguration) createRateLimitingMiddleware(ctx *pulumi.Context) error {
	component, err := app.NewDaprComponent(ctx, "production-ratelimit-middleware", &app.DaprComponentArgs{
		ResourceGroupName:    sc.containerAppsStack.GetResourceGroup().Name,
		EnvironmentName:     sc.containerAppsStack.GetEnvironment().Name,
		ComponentName:       pulumi.String("ratelimit"),
		ComponentType:       pulumi.String("middleware.http.ratelimit"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("maxRequestsPerSecond"),
				Value: pulumi.String("100"), // Production rate limit
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("statusCode"),
				Value: pulumi.String("429"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("statusMessage"),
				Value: pulumi.String("Too Many Requests - Rate limit exceeded"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("headerRateLimit"),
				Value: pulumi.String("X-RateLimit-Limit"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("headerRateRemaining"),
				Value: pulumi.String("X-RateLimit-Remaining"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("headerRetryAfter"),
				Value: pulumi.String("Retry-After"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("strategy"),
				Value: pulumi.String("sliding-window"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("windowSize"),
				Value: pulumi.String("60s"),
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("public-gateway"),
			pulumi.String("admin-gateway"),
		},
	})
	if err != nil {
		return err
	}

	sc.securityComponents["ratelimit"] = component
	return nil
}

func (sc *SecurityConfiguration) createSecurityHeadersMiddleware(ctx *pulumi.Context) error {
	component, err := app.NewDaprComponent(ctx, "production-security-headers-middleware", &app.DaprComponentArgs{
		ResourceGroupName:    sc.containerAppsStack.GetResourceGroup().Name,
		EnvironmentName:     sc.containerAppsStack.GetEnvironment().Name,
		ComponentName:       pulumi.String("security-headers"),
		ComponentType:       pulumi.String("middleware.http.headers"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("X-Frame-Options"),
				Value: pulumi.String("DENY"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("X-Content-Type-Options"),
				Value: pulumi.String("nosniff"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("X-XSS-Protection"),
				Value: pulumi.String("1; mode=block"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("Strict-Transport-Security"),
				Value: pulumi.String("max-age=31536000; includeSubDomains; preload"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("Content-Security-Policy"),
				Value: pulumi.String("default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self' https://api.international-center.com"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("Referrer-Policy"),
				Value: pulumi.String("strict-origin-when-cross-origin"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("Permissions-Policy"),
				Value: pulumi.String("geolocation=(), microphone=(), camera=()"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("X-Permitted-Cross-Domain-Policies"),
				Value: pulumi.String("none"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("Cross-Origin-Embedder-Policy"),
				Value: pulumi.String("require-corp"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("Cross-Origin-Opener-Policy"),
				Value: pulumi.String("same-origin"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("Cross-Origin-Resource-Policy"),
				Value: pulumi.String("same-origin"),
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("public-gateway"),
			pulumi.String("admin-gateway"),
		},
	})
	if err != nil {
		return err
	}

	sc.securityComponents["security-headers"] = component
	return nil
}

func (sc *SecurityConfiguration) createInputValidationMiddleware(ctx *pulumi.Context) error {
	component, err := app.NewDaprComponent(ctx, "production-input-validation-middleware", &app.DaprComponentArgs{
		ResourceGroupName:    sc.containerAppsStack.GetResourceGroup().Name,
		EnvironmentName:     sc.containerAppsStack.GetEnvironment().Name,
		ComponentName:       pulumi.String("input-validation"),
		ComponentType:       pulumi.String("middleware.http.validator"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("maxRequestSize"),
				Value: pulumi.String("10485760"), // 10MB max request size
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("validateJSON"),
				Value: pulumi.String("true"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("validateXML"),
				Value: pulumi.String("false"), // Disable XML for security
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("sanitizeHTML"),
				Value: pulumi.String("true"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("blockSQLInjection"),
				Value: pulumi.String("true"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("blockXSS"),
				Value: pulumi.String("true"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("blockCommandInjection"),
				Value: pulumi.String("true"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("statusCode"),
				Value: pulumi.String("400"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("statusMessage"),
				Value: pulumi.String("Bad Request - Input validation failed"),
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("public-gateway"),
			pulumi.String("admin-gateway"),
			pulumi.String("identity-api"),
			pulumi.String("content-api"),
			pulumi.String("services-api"),
		},
	})
	if err != nil {
		return err
	}

	sc.securityComponents["input-validation"] = component
	return nil
}

func (sc *SecurityConfiguration) createAuditLoggingMiddleware(ctx *pulumi.Context) error {
	component, err := app.NewDaprComponent(ctx, "production-audit-logging-middleware", &app.DaprComponentArgs{
		ResourceGroupName:    sc.containerAppsStack.GetResourceGroup().Name,
		EnvironmentName:     sc.containerAppsStack.GetEnvironment().Name,
		ComponentName:       pulumi.String("audit-logging"),
		ComponentType:       pulumi.String("middleware.http.audit"),
		Version:            pulumi.String("v1"),
		Metadata: app.DaprMetadataArray{
			&app.DaprMetadataArgs{
				Name:  pulumi.String("logLevel"),
				Value: pulumi.String("INFO"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("logFormat"),
				Value: pulumi.String("json"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("includeRequestHeaders"),
				Value: pulumi.String("Authorization,X-User-ID,X-Correlation-ID,User-Agent,X-Forwarded-For"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("includeResponseHeaders"),
				Value: pulumi.String("Content-Type,X-Request-ID"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("includeRequestBody"),
				Value: pulumi.String("false"), // For performance and privacy
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("includeResponseBody"),
				Value: pulumi.String("false"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("auditFailedRequests"),
				Value: pulumi.String("true"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("auditSuccessRequests"),
				Value: pulumi.String("true"),
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("destination"),
				Value: pulumi.String("grafana-loki"), // Send to Grafana Cloud Loki
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("retentionDays"),
				Value: pulumi.String("2555"), // 7 years for compliance
			},
			&app.DaprMetadataArgs{
				Name:  pulumi.String("encryptLogs"),
				Value: pulumi.String("true"),
			},
			&app.DaprMetadataArgs{
				Name:      pulumi.String("encryptionKey"),
				SecretRef: pulumi.String("audit-signing-key"),
			},
		},
		Secrets: app.SecretArray{
			&app.SecretArgs{
				Name:        pulumi.String("audit-signing-key"),
				KeyVaultUrl: sc.vaultStack.GetVaultUri(),
				Identity:    pulumi.String("system"),
			},
		},
		Scopes: pulumi.StringArray{
			pulumi.String("public-gateway"),
			pulumi.String("admin-gateway"),
			pulumi.String("identity-api"),
			pulumi.String("content-api"),
			pulumi.String("services-api"),
		},
	})
	if err != nil {
		return err
	}

	sc.securityComponents["audit-logging"] = component
	return nil
}

func (sc *SecurityConfiguration) configureSecurityPolicies(ctx *pulumi.Context) error {
	// Configure security policies for production environment
	sc.securityPolicies["authentication-policy"] = &SecurityPolicy{
		Name:             "Production Authentication Policy",
		Type:             "authentication",
		EnforcementLevel: "strict",
		Rules: []SecurityRule{
			{
				ID:          "auth-001",
				Name:        "Require Authentication",
				Type:        "authentication",
				Condition:   "request.path != '/health' AND request.path != '/ready'",
				Action:      "require_auth",
				Severity:    "HIGH",
				Enabled:     true,
				Description: "All requests except health checks must be authenticated",
			},
			{
				ID:          "auth-002",
				Name:        "Multi-Factor Authentication",
				Type:        "authentication",
				Condition:   "user.role == 'admin' OR request.path starts_with '/admin'",
				Action:      "require_mfa",
				Severity:    "CRITICAL",
				Enabled:     true,
				Description: "Admin users and admin endpoints require MFA",
			},
		},
		AuditLevel:           "FULL",
		ComplianceFrameworks: []string{"SOC2-Type2", "ISO27001"},
	}

	sc.securityPolicies["authorization-policy"] = &SecurityPolicy{
		Name:             "Production Authorization Policy",
		Type:             "authorization",
		EnforcementLevel: "strict",
		Rules: []SecurityRule{
			{
				ID:          "authz-001",
				Name:        "Role-Based Access Control",
				Type:        "authorization",
				Condition:   "request.path starts_with '/admin'",
				Action:      "check_admin_role",
				Severity:    "HIGH",
				Enabled:     true,
				Description: "Admin endpoints require admin role",
			},
			{
				ID:          "authz-002",
				Name:        "Resource-Based Access Control",
				Type:        "authorization",
				Condition:   "request.method in ['PUT', 'DELETE', 'PATCH']",
				Action:      "check_resource_owner",
				Severity:    "HIGH",
				Enabled:     true,
				Description: "Modification operations require resource ownership",
			},
		},
		AuditLevel:           "FULL",
		ComplianceFrameworks: []string{"SOC2-Type2", "GDPR"},
	}

	return nil
}

func (sc *SecurityConfiguration) setupAdvancedThreatProtection(ctx *pulumi.Context) error {
	return nil
}

func (sc *SecurityConfiguration) configureComplianceRules(ctx *pulumi.Context) error {
	sc.complianceRules["soc2-access-control"] = &ComplianceRule{
		Framework:        "SOC2-Type2",
		Requirement:     "CC6.1",
		Control:         "Logical and Physical Access Controls",
		Implementation:  "Role-based access control with MFA for privileged access",
		ValidationMethod: "Automated policy enforcement + quarterly access reviews",
		Evidence:        []string{"access-logs", "mfa-logs", "policy-configs"},
		Status:          "COMPLIANT",
		LastAudited:     "2024-01-01",
		NextAudit:       "2024-04-01",
	}

	sc.complianceRules["gdpr-data-protection"] = &ComplianceRule{
		Framework:        "GDPR",
		Requirement:     "Article 32",
		Control:         "Security of Processing",
		Implementation:  "Encryption at rest and in transit + audit logging",
		ValidationMethod: "Automated encryption validation + audit log analysis",
		Evidence:        []string{"encryption-configs", "audit-logs", "security-scans"},
		Status:          "COMPLIANT",
		LastAudited:     "2024-01-01",
		NextAudit:       "2024-04-01",
	}

	return nil
}

func getAdvancedThreatProtectionConfig() *AdvancedThreatProtection {
	return &AdvancedThreatProtection{
		Enabled:                 true,
		RealTimeScanning:        true,
		BehaviorAnalysis:        true,
		AnomalyDetection:        true,
		ThreatIntelligenceFeeds: []string{"microsoft-defender", "azure-security-center"},
		AutomaticResponse:       true,
		QuarantineThreshold:     "MEDIUM",
		NotificationEndpoints:   []string{"security-team@international-center.com"},
		IncidentEscalationRules: []EscalationRule{
			{
				ThreatLevel:      "CRITICAL",
				ResponseTime:     "5m",
				NotificationList: []string{"security-lead", "ciso", "incident-commander"},
				AutomaticActions: []string{"isolate-service", "notify-stakeholders"},
			},
			{
				ThreatLevel:      "HIGH",
				ResponseTime:     "15m",
				NotificationList: []string{"security-team", "ops-team"},
				AutomaticActions: []string{"increase-monitoring", "notify-team"},
			},
		},
	}
}

func (sc *SecurityConfiguration) GetSecurityComponent(name string) *app.DaprComponent {
	return sc.securityComponents[name]
}

func (sc *SecurityConfiguration) GetSecurityPolicy(name string) *SecurityPolicy {
	return sc.securityPolicies[name]
}

func (sc *SecurityConfiguration) GetComplianceRule(name string) *ComplianceRule {
	return sc.complianceRules[name]
}

func (sc *SecurityConfiguration) GetThreatProtectionConfig() *AdvancedThreatProtection {
	return sc.threatProtection
}