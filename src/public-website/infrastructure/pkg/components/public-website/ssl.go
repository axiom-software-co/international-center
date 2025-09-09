package website

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SSLArgs struct {
	Environment string
}

type SSLComponent struct {
	pulumi.ResourceState

	CertificateSource   pulumi.StringOutput `pulumi:"certificateSource"`
	SSLEnabled          pulumi.BoolOutput   `pulumi:"sslEnabled"`
	CertificateSettings pulumi.MapOutput    `pulumi:"certificateSettings"`
	SecurityHeaders     pulumi.MapOutput    `pulumi:"securityHeaders"`
}

func NewSSLComponent(ctx *pulumi.Context, name string, args *SSLArgs, opts ...pulumi.ResourceOption) (*SSLComponent, error) {
	component := &SSLComponent{}
	
	err := ctx.RegisterComponentResource("international-center:website:SSL", name, component, opts...)
	if err != nil {
		return nil, err
	}

	var certificateSource pulumi.StringOutput
	var sslEnabled pulumi.BoolOutput
	var certificateSettings, securityHeaders pulumi.MapOutput

	switch args.Environment {
	case "development":
		certificateSource = pulumi.String("self_signed").ToStringOutput()
		sslEnabled = pulumi.Bool(false).ToBoolOutput()
		certificateSettings = pulumi.Map{
			"enabled":           pulumi.Bool(false),
			"auto_renewal":      pulumi.Bool(false),
			"minimum_tls_version": pulumi.String("1.2"),
		}.ToMapOutput()
		securityHeaders = pulumi.Map{
			"strict_transport_security": pulumi.Bool(false),
			"content_security_policy":   pulumi.Bool(false),
			"x_frame_options":           pulumi.Bool(false),
		}.ToMapOutput()
	case "staging":
		certificateSource = pulumi.String("letsencrypt").ToStringOutput()
		sslEnabled = pulumi.Bool(true).ToBoolOutput()
		certificateSettings = pulumi.Map{
			"enabled":           pulumi.Bool(true),
			"auto_renewal":      pulumi.Bool(true),
			"minimum_tls_version": pulumi.String("1.2"),
			"cipher_suites":     pulumi.StringArray{
				pulumi.String("TLS_AES_256_GCM_SHA384"),
				pulumi.String("TLS_CHACHA20_POLY1305_SHA256"),
			},
		}.ToMapOutput()
		securityHeaders = pulumi.Map{
			"strict_transport_security": pulumi.Bool(true),
			"content_security_policy":   pulumi.Bool(true),
			"x_frame_options":           pulumi.Bool(true),
			"x_content_type_options":    pulumi.Bool(true),
		}.ToMapOutput()
	case "production":
		certificateSource = pulumi.String("azure_key_vault").ToStringOutput()
		sslEnabled = pulumi.Bool(true).ToBoolOutput()
		certificateSettings = pulumi.Map{
			"enabled":           pulumi.Bool(true),
			"auto_renewal":      pulumi.Bool(true),
			"minimum_tls_version": pulumi.String("1.3"),
			"cipher_suites":     pulumi.StringArray{
				pulumi.String("TLS_AES_256_GCM_SHA384"),
				pulumi.String("TLS_CHACHA20_POLY1305_SHA256"),
				pulumi.String("TLS_AES_128_GCM_SHA256"),
			},
			"perfect_forward_secrecy": pulumi.Bool(true),
			"certificate_transparency": pulumi.Bool(true),
		}.ToMapOutput()
		securityHeaders = pulumi.Map{
			"strict_transport_security": pulumi.Bool(true),
			"content_security_policy":   pulumi.Bool(true),
			"x_frame_options":           pulumi.Bool(true),
			"x_content_type_options":    pulumi.Bool(true),
			"referrer_policy":           pulumi.Bool(true),
			"permissions_policy":        pulumi.Bool(true),
		}.ToMapOutput()
	default:
		return nil, fmt.Errorf("unsupported environment: %s", args.Environment)
	}

	component.CertificateSource = certificateSource
	component.SSLEnabled = sslEnabled
	component.CertificateSettings = certificateSettings
	component.SecurityHeaders = securityHeaders

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"certificateSource":   component.CertificateSource,
		"sslEnabled":          component.SSLEnabled,
		"certificateSettings": component.CertificateSettings,
		"securityHeaders":     component.SecurityHeaders,
	}); err != nil {
		return nil, err
	}

	return component, nil
}