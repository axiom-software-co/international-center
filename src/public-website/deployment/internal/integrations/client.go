package integrations

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type GrafanaCloudConfig struct {
	StackName   string
	OrgID       string
	AccessToken string
}

type GrafanaCloudArgs struct {
	Config      *GrafanaCloudConfig
	Environment string
	ProjectName string
}

type GrafanaCloudComponent struct {
	pulumi.ResourceState
	GrafanaURL pulumi.StringOutput
}

func DevelopmentGrafanaCloudConfig() *GrafanaCloudConfig {
	return &GrafanaCloudConfig{
		StackName:   "development-stack",
		OrgID:       "dev-org",
		AccessToken: "dev-token",
	}
}

func DefaultGrafanaCloudConfig(stackName, orgID, token string) *GrafanaCloudConfig {
	return &GrafanaCloudConfig{
		StackName:   stackName,
		OrgID:       orgID,
		AccessToken: token,
	}
}

func ProductionGrafanaCloudConfig(stackName, orgID, token string) *GrafanaCloudConfig {
	return &GrafanaCloudConfig{
		StackName:   stackName,
		OrgID:       orgID,
		AccessToken: token,
	}
}

func NewGrafanaCloudComponent(ctx *pulumi.Context, name string, args *GrafanaCloudArgs, opts ...pulumi.ResourceOption) (*GrafanaCloudComponent, error) {
	component := &GrafanaCloudComponent{}
	err := ctx.RegisterComponentResource("axiom:integrations:GrafanaCloud", name, component, opts...)
	if err != nil {
		return nil, err
	}

	// Initialize GrafanaURL based on config
	grafanaURL := pulumi.Sprintf("https://%s.grafana.net", args.Config.StackName)
	component.GrafanaURL = grafanaURL

	return component, nil
}