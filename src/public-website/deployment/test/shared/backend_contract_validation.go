package shared

import (
	"context"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
)

// BackendContractValidator validates OpenAPI contract compliance for backend services
type BackendContractValidator struct {
	adminSpecPath  string
	publicSpecPath string
	environment    string
}

// NewBackendContractValidator creates a new backend contract compliance validator
func NewBackendContractValidator(environment string) *BackendContractValidator {
	return &BackendContractValidator{
		environment:    environment,
		adminSpecPath:  "/home/tojkuv/Documents/GitHub/international-center-workspace/international-center/src/public-website/contracts/openapi/admin-api.yaml",
		publicSpecPath: "/home/tojkuv/Documents/GitHub/international-center-workspace/international-center/src/public-website/contracts/openapi/public-api.yaml",
	}
}

// ValidateContractCompliance validates contract compliance for backend services
func (v *BackendContractValidator) ValidateContractCompliance(ctx context.Context) error {
	// Validate admin API specification
	if err := v.validateSpecification(v.adminSpecPath, "Admin API"); err != nil {
		return fmt.Errorf("admin API contract validation failed: %w", err)
	}
	
	// Validate public API specification
	if err := v.validateSpecification(v.publicSpecPath, "Public API"); err != nil {
		return fmt.Errorf("public API contract validation failed: %w", err)
	}
	
	return nil
}

// validateSpecification validates an individual OpenAPI specification
func (v *BackendContractValidator) validateSpecification(specPath, apiName string) error {
	// Check if specification file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		return fmt.Errorf("specification file not found: %s", specPath)
	}
	
	// Load and validate OpenAPI specification
	ctx := context.Background()
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	spec, err := loader.LoadFromFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to load %s specification: %w", apiName, err)
	}
	
	// Validate specification structure
	if err := spec.Validate(ctx); err != nil {
		return fmt.Errorf("%s specification validation failed: %w", apiName, err)
	}
	
	// Validate required specification properties
	if spec.Info == nil {
		return fmt.Errorf("%s specification missing info section", apiName)
	}
	
	if spec.Info.Title == "" {
		return fmt.Errorf("%s specification missing title", apiName)
	}
	
	if spec.Info.Version == "" {
		return fmt.Errorf("%s specification missing version", apiName)
	}
	
	return nil
}

// BackendContractValidation performs contract validation for backend services
type BackendContractValidation struct {
	Environment string
	Validator   *BackendContractValidator
}

// NewBackendContractValidation creates backend contract validation
func NewBackendContractValidation(environment string) *BackendContractValidation {
	validator := NewBackendContractValidator(environment)
	
	return &BackendContractValidation{
		Environment: environment,
		Validator:   validator,
	}
}

// RunPreDeploymentValidation runs contract validations for backend services
func (bv *BackendContractValidation) RunPreDeploymentValidation(ctx context.Context) error {
	// Step 1: Validate OpenAPI specifications
	if err := bv.Validator.ValidateContractCompliance(ctx); err != nil {
		return fmt.Errorf("contract compliance validation failed: %w", err)
	}
	
	// Step 2: Validate backend service implementations exist
	if err := bv.validateBackendImplementations(); err != nil {
		return fmt.Errorf("backend implementation validation failed: %w", err)
	}
	
	return nil
}

// validateBackendImplementations validates backend service implementations exist
func (bv *BackendContractValidation) validateBackendImplementations() error {
	// Check if backend service implementations exist
	expectedServices := []string{
		"/home/tojkuv/Documents/GitHub/international-center-workspace/international-center/src/public-website/backend/cmd/content",
		"/home/tojkuv/Documents/GitHub/international-center-workspace/international-center/src/public-website/backend/cmd/inquiries",
		"/home/tojkuv/Documents/GitHub/international-center-workspace/international-center/src/public-website/backend/cmd/notifications",
	}
	
	for _, servicePath := range expectedServices {
		if _, err := os.Stat(servicePath); os.IsNotExist(err) {
			return fmt.Errorf("backend service implementation missing: %s", servicePath)
		}
	}
	
	return nil
}