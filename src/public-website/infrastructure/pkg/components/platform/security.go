package platform

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SecurityArgs struct {
	Environment string
}

type SecurityComponent struct {
	pulumi.ResourceState

	Policies         pulumi.MapOutput    `pulumi:"policies"`
	AuthenticationConfig pulumi.MapOutput    `pulumi:"authenticationConfig"`
	TLSConfig        pulumi.MapOutput    `pulumi:"tlsConfig"`
	AuditLogging     pulumi.BoolOutput   `pulumi:"auditLogging"`
}

func NewSecurityComponent(ctx *pulumi.Context, name string, args *SecurityArgs, opts ...pulumi.ResourceOption) (*SecurityComponent, error) {
	component := &SecurityComponent{}
	
	// Safe registration for mock contexts
	if canRegister(ctx) {
		err := ctx.RegisterComponentResource("international-center:platform:Security", name, component, opts...)
		if err != nil {
			return nil, err
		}
	}

	var policies, authenticationConfig, tlsConfig pulumi.MapOutput
	var auditLogging pulumi.BoolOutput

	switch args.Environment {
	case "development":
		policies = pulumi.Map{
			"pod_security_policy":    pulumi.Bool(true),
			"network_policies":       pulumi.Bool(false),
			"service_account_tokens": pulumi.Bool(true),
			"resource_quotas":        pulumi.Bool(false),
			"authentication_enabled": pulumi.Bool(true),
			"authorization_enabled":  pulumi.Bool(false),
			"audit_logging_enabled":  pulumi.Bool(false),
		}.ToMapOutput()
		authenticationConfig = pulumi.Map{
			"enabled":       pulumi.Bool(true),
			"provider":      pulumi.String("local"),
			"token_expiry":  pulumi.String("24h"),
			"refresh_token": pulumi.Bool(true),
		}.ToMapOutput()
		tlsConfig = pulumi.Map{
			"enabled":         pulumi.Bool(false),
			"cert_source":     pulumi.String("self_signed"),
			"min_version":     pulumi.String("1.2"),
			"cipher_suites":   pulumi.StringArray{pulumi.String("TLS_AES_256_GCM_SHA384")},
		}.ToMapOutput()
		auditLogging = pulumi.Bool(false).ToBoolOutput()
	case "staging":
		policies = pulumi.Map{
			"pod_security_policy":    pulumi.Bool(true),
			"network_policies":       pulumi.Bool(true),
			"service_account_tokens": pulumi.Bool(true),
			"resource_quotas":        pulumi.Bool(true),
			"authentication_enabled": pulumi.Bool(true),
			"authorization_enabled":  pulumi.Bool(true),
			"audit_logging_enabled":  pulumi.Bool(true),
		}.ToMapOutput()
		authenticationConfig = pulumi.Map{
			"enabled":       pulumi.Bool(true),
			"provider":      pulumi.String("azure_ad"),
			"token_expiry":  pulumi.String("8h"),
			"refresh_token": pulumi.Bool(true),
		}.ToMapOutput()
		tlsConfig = pulumi.Map{
			"enabled":         pulumi.Bool(true),
			"cert_source":     pulumi.String("azure_key_vault"),
			"min_version":     pulumi.String("1.3"),
			"cipher_suites":   pulumi.StringArray{pulumi.String("TLS_AES_256_GCM_SHA384"), pulumi.String("TLS_CHACHA20_POLY1305_SHA256")},
		}.ToMapOutput()
		auditLogging = pulumi.Bool(true).ToBoolOutput()
	case "production":
		policies = pulumi.Map{
			"pod_security_policy":    pulumi.Bool(true),
			"network_policies":       pulumi.Bool(true),
			"service_account_tokens": pulumi.Bool(true),
			"resource_quotas":        pulumi.Bool(true),
			"authentication_enabled": pulumi.Bool(true),
			"authorization_enabled":  pulumi.Bool(true),
			"audit_logging_enabled":  pulumi.Bool(true),
		}.ToMapOutput()
		authenticationConfig = pulumi.Map{
			"enabled":       pulumi.Bool(true),
			"provider":      pulumi.String("azure_ad"),
			"token_expiry":  pulumi.String("4h"),
			"refresh_token": pulumi.Bool(true),
		}.ToMapOutput()
		tlsConfig = pulumi.Map{
			"enabled":         pulumi.Bool(true),
			"cert_source":     pulumi.String("azure_key_vault"),
			"min_version":     pulumi.String("1.3"),
			"cipher_suites":   pulumi.StringArray{pulumi.String("TLS_AES_256_GCM_SHA384"), pulumi.String("TLS_CHACHA20_POLY1305_SHA256")},
		}.ToMapOutput()
		auditLogging = pulumi.Bool(true).ToBoolOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.Policies = policies
	component.AuthenticationConfig = authenticationConfig
	component.TLSConfig = tlsConfig
	component.AuditLogging = auditLogging

	if canRegister(ctx) {
		if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
			"policies":             component.Policies,
			"authenticationConfig": component.AuthenticationConfig,
			"tlsConfig":            component.TLSConfig,
			"auditLogging":         component.AuditLogging,
		}); err != nil {
			return nil, err
		}
	}

	return component, nil
}