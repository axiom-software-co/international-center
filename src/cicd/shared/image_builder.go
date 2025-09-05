package shared

import (
	"github.com/axiom-software-co/international-center/src/cicd/components"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// NewImageBuilder creates a new image builder for the specified environment
// This is a wrapper around the components.NewImageBuilder to maintain the shared package interface
func NewImageBuilder(ctx *pulumi.Context, environment string) *components.ImageBuilder {
	return components.NewImageBuilder(ctx, environment)
}