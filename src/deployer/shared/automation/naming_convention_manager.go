package automation

import (
	"fmt"
	"regexp"
	"strings"
)

// NamingConventionManager enforces consistent naming patterns across environments
type NamingConventionManager struct {
	environment string
	conventions map[ResourceType]*NamingConvention
	validators  map[ResourceType]*NamingValidator
}

// ResourceType defines types of resources for naming conventions
type ResourceType string

const (
	ResourceTypeResourceGroup    ResourceType = "resource_group"
	ResourceTypeSQLServer        ResourceType = "sql_server"
	ResourceTypeSQLDatabase      ResourceType = "sql_database"
	ResourceTypeStorageAccount   ResourceType = "storage_account"
	ResourceTypeKeyVault         ResourceType = "key_vault"
	ResourceTypeContainerApp     ResourceType = "container_app"
	ResourceTypeContainerEnvironment ResourceType = "container_environment"
	ResourceTypeLogAnalytics     ResourceType = "log_analytics"
	ResourceTypeSecurityGroup    ResourceType = "security_group"
	ResourceTypeVirtualNetwork   ResourceType = "virtual_network"
	ResourceTypeSubnet          ResourceType = "subnet"
)

// NamingConvention defines naming pattern for resource type
type NamingConvention struct {
	ResourceType ResourceType
	Pattern      string
	Description  string
	Examples     []string
	MinLength    int
	MaxLength    int
	Required     []string
	Optional     []string
	Forbidden    []string
}

// NamingValidator validates names against conventions
type NamingValidator struct {
	ResourceType ResourceType
	Pattern      *regexp.Regexp
	MinLength    int
	MaxLength    int
	Required     []string
	Forbidden    []string
}

// NamingValidationResult represents validation result
type NamingValidationResult struct {
	Valid       bool
	Name        string
	ResourceType ResourceType
	Environment string
	Errors      []string
	Warnings    []string
	Suggestions []string
}

// NewNamingConventionManager creates naming convention manager
func NewNamingConventionManager(environment string) *NamingConventionManager {
	ncm := &NamingConventionManager{
		environment: environment,
		conventions: make(map[ResourceType]*NamingConvention),
		validators:  make(map[ResourceType]*NamingValidator),
	}
	
	ncm.configureConventions()
	ncm.buildValidators()
	
	return ncm
}

// configureConventions sets up naming conventions for each resource type
func (ncm *NamingConventionManager) configureConventions() {
	// Resource Group convention
	ncm.conventions[ResourceTypeResourceGroup] = &NamingConvention{
		ResourceType: ResourceTypeResourceGroup,
		Pattern:      fmt.Sprintf("rg-%s-[component]-[region]", ncm.environment),
		Description:  "Resource groups must follow environment-component-region pattern",
		Examples:     []string{fmt.Sprintf("rg-%s-database-eastus2", ncm.environment), fmt.Sprintf("rg-%s-storage-eastus2", ncm.environment)},
		MinLength:    10,
		MaxLength:    80,
		Required:     []string{"rg", ncm.environment},
		Forbidden:    []string{"prod", "production", "dev", "development", "test", "testing"},
	}

	// SQL Server convention
	ncm.conventions[ResourceTypeSQLServer] = &NamingConvention{
		ResourceType: ResourceTypeSQLServer,
		Pattern:      fmt.Sprintf("sql-%s-[component]-[region]-[suffix]", ncm.environment),
		Description:  "SQL servers must include environment, component, region and unique suffix",
		Examples:     []string{fmt.Sprintf("sql-%s-app-eastus2-001", ncm.environment)},
		MinLength:    15,
		MaxLength:    60,
		Required:     []string{"sql", ncm.environment},
		Forbidden:    []string{"server", "database", "prod", "dev"},
	}

	// SQL Database convention
	ncm.conventions[ResourceTypeSQLDatabase] = &NamingConvention{
		ResourceType: ResourceTypeSQLDatabase,
		Pattern:      fmt.Sprintf("db-%s-[component]", ncm.environment),
		Description:  "SQL databases must include environment and component",
		Examples:     []string{fmt.Sprintf("db-%s-app", ncm.environment), fmt.Sprintf("db-%s-identity", ncm.environment)},
		MinLength:    8,
		MaxLength:    120,
		Required:     []string{"db", ncm.environment},
		Forbidden:    []string{"database", "prod", "dev"},
	}

	// Storage Account convention (must be globally unique and lowercase)
	ncm.conventions[ResourceTypeStorageAccount] = &NamingConvention{
		ResourceType: ResourceTypeStorageAccount,
		Pattern:      fmt.Sprintf("st%s[component][suffix]", ncm.environment),
		Description:  "Storage accounts must be lowercase, globally unique with environment and component",
		Examples:     []string{fmt.Sprintf("st%sapp001", ncm.environment), fmt.Sprintf("st%sdata002", ncm.environment)},
		MinLength:    8,
		MaxLength:    24,
		Required:     []string{"st", ncm.environment},
		Forbidden:    []string{"storage", "account", "prod", "dev", "test"},
	}

	// Key Vault convention (globally unique)
	ncm.conventions[ResourceTypeKeyVault] = &NamingConvention{
		ResourceType: ResourceTypeKeyVault,
		Pattern:      fmt.Sprintf("kv-%s-[component]-[suffix]", ncm.environment),
		Description:  "Key vaults must be globally unique with environment and component",
		Examples:     []string{fmt.Sprintf("kv-%s-app-001", ncm.environment), fmt.Sprintf("kv-%s-shared-002", ncm.environment)},
		MinLength:    10,
		MaxLength:    24,
		Required:     []string{"kv", ncm.environment},
		Forbidden:    []string{"vault", "key", "secret", "prod", "dev"},
	}

	// Container App convention
	ncm.conventions[ResourceTypeContainerApp] = &NamingConvention{
		ResourceType: ResourceTypeContainerApp,
		Pattern:      fmt.Sprintf("ca-%s-[component]", ncm.environment),
		Description:  "Container apps must include environment and component",
		Examples:     []string{fmt.Sprintf("ca-%s-api", ncm.environment), fmt.Sprintf("ca-%s-gateway", ncm.environment)},
		MinLength:    8,
		MaxLength:    32,
		Required:     []string{"ca", ncm.environment},
		Forbidden:    []string{"container", "app", "prod", "dev"},
	}

	// Container Environment convention
	ncm.conventions[ResourceTypeContainerEnvironment] = &NamingConvention{
		ResourceType: ResourceTypeContainerEnvironment,
		Pattern:      fmt.Sprintf("cae-%s-[region]", ncm.environment),
		Description:  "Container environments must include environment and region",
		Examples:     []string{fmt.Sprintf("cae-%s-eastus2", ncm.environment)},
		MinLength:    12,
		MaxLength:    32,
		Required:     []string{"cae", ncm.environment},
		Forbidden:    []string{"environment", "prod", "dev"},
	}

	// Log Analytics Workspace convention
	ncm.conventions[ResourceTypeLogAnalytics] = &NamingConvention{
		ResourceType: ResourceTypeLogAnalytics,
		Pattern:      fmt.Sprintf("log-%s-[region]", ncm.environment),
		Description:  "Log Analytics workspaces must include environment and region",
		Examples:     []string{fmt.Sprintf("log-%s-eastus2", ncm.environment)},
		MinLength:    10,
		MaxLength:    60,
		Required:     []string{"log", ncm.environment},
		Forbidden:    []string{"analytics", "workspace", "prod", "dev"},
	}

	// Virtual Network convention
	ncm.conventions[ResourceTypeVirtualNetwork] = &NamingConvention{
		ResourceType: ResourceTypeVirtualNetwork,
		Pattern:      fmt.Sprintf("vnet-%s-[region]", ncm.environment),
		Description:  "Virtual networks must include environment and region",
		Examples:     []string{fmt.Sprintf("vnet-%s-eastus2", ncm.environment)},
		MinLength:    12,
		MaxLength:    64,
		Required:     []string{"vnet", ncm.environment},
		Forbidden:    []string{"network", "virtual", "prod", "dev"},
	}

	// Subnet convention
	ncm.conventions[ResourceTypeSubnet] = &NamingConvention{
		ResourceType: ResourceTypeSubnet,
		Pattern:      fmt.Sprintf("snet-%s-[component]", ncm.environment),
		Description:  "Subnets must include environment and component",
		Examples:     []string{fmt.Sprintf("snet-%s-database", ncm.environment), fmt.Sprintf("snet-%s-app", ncm.environment)},
		MinLength:    10,
		MaxLength:    64,
		Required:     []string{"snet", ncm.environment},
		Forbidden:    []string{"subnet", "network", "prod", "dev"},
	}

	// Network Security Group convention
	ncm.conventions[ResourceTypeSecurityGroup] = &NamingConvention{
		ResourceType: ResourceTypeSecurityGroup,
		Pattern:      fmt.Sprintf("nsg-%s-[component]", ncm.environment),
		Description:  "Network security groups must include environment and component",
		Examples:     []string{fmt.Sprintf("nsg-%s-database", ncm.environment), fmt.Sprintf("nsg-%s-app", ncm.environment)},
		MinLength:    10,
		MaxLength:    64,
		Required:     []string{"nsg", ncm.environment},
		Forbidden:    []string{"security", "group", "prod", "dev"},
	}
}

// buildValidators creates regex validators from conventions
func (ncm *NamingConventionManager) buildValidators() {
	for resourceType, convention := range ncm.conventions {
		// Convert naming pattern to regex
		pattern := ncm.conventionToRegex(convention)
		regex, err := regexp.Compile(pattern)
		if err != nil {
			fmt.Printf("Warning: Failed to compile regex for %s: %v\n", resourceType, err)
			continue
		}

		ncm.validators[resourceType] = &NamingValidator{
			ResourceType: resourceType,
			Pattern:      regex,
			MinLength:    convention.MinLength,
			MaxLength:    convention.MaxLength,
			Required:     convention.Required,
			Forbidden:    convention.Forbidden,
		}
	}
}

// conventionToRegex converts naming convention pattern to regex
func (ncm *NamingConventionManager) conventionToRegex(convention *NamingConvention) string {
	pattern := convention.Pattern

	// Replace placeholders with regex patterns
	pattern = strings.ReplaceAll(pattern, "[component]", "[a-z0-9]+")
	pattern = strings.ReplaceAll(pattern, "[region]", "[a-z0-9]+")
	pattern = strings.ReplaceAll(pattern, "[suffix]", "[a-z0-9]+")

	// Ensure start and end anchors
	if !strings.HasPrefix(pattern, "^") {
		pattern = "^" + pattern
	}
	if !strings.HasSuffix(pattern, "$") {
		pattern = pattern + "$"
	}

	return pattern
}

// ValidateName validates resource name against conventions
func (ncm *NamingConventionManager) ValidateName(name string, resourceType ResourceType) *NamingValidationResult {
	result := &NamingValidationResult{
		Name:         name,
		ResourceType: resourceType,
		Environment:  ncm.environment,
		Valid:        true,
		Errors:       []string{},
		Warnings:     []string{},
		Suggestions:  []string{},
	}

	validator, exists := ncm.validators[resourceType]
	if !exists {
		result.Errors = append(result.Errors, fmt.Sprintf("No naming convention defined for resource type %s", resourceType))
		result.Valid = false
		return result
	}

	convention := ncm.conventions[resourceType]

	// Check length constraints
	if len(name) < validator.MinLength {
		result.Errors = append(result.Errors, fmt.Sprintf("Name too short, minimum length is %d characters", validator.MinLength))
		result.Valid = false
	}

	if len(name) > validator.MaxLength {
		result.Errors = append(result.Errors, fmt.Sprintf("Name too long, maximum length is %d characters", validator.MaxLength))
		result.Valid = false
	}

	// Check pattern match
	if !validator.Pattern.MatchString(name) {
		result.Errors = append(result.Errors, fmt.Sprintf("Name does not match required pattern: %s", convention.Pattern))
		result.Valid = false
		result.Suggestions = append(result.Suggestions, fmt.Sprintf("Example valid names: %v", convention.Examples))
	}

	// Check required elements
	namelower := strings.ToLower(name)
	for _, required := range validator.Required {
		if !strings.Contains(namelower, strings.ToLower(required)) {
			result.Errors = append(result.Errors, fmt.Sprintf("Name must contain required element: %s", required))
			result.Valid = false
		}
	}

	// Check forbidden elements
	for _, forbidden := range validator.Forbidden {
		if strings.Contains(namelower, strings.ToLower(forbidden)) {
			result.Errors = append(result.Errors, fmt.Sprintf("Name must not contain forbidden element: %s", forbidden))
			result.Valid = false
		}
	}

	// Environment-specific validations
	ncm.validateEnvironmentSpecific(name, resourceType, result)

	return result
}

// validateEnvironmentSpecific applies environment-specific validation rules
func (ncm *NamingConventionManager) validateEnvironmentSpecific(name string, resourceType ResourceType, result *NamingValidationResult) {
	namelower := strings.ToLower(name)

	// Production-specific rules
	if ncm.environment == "production" {
		// Production resources should not contain "test", "dev", "temp", "demo"
		prodForbidden := []string{"test", "dev", "temp", "demo", "example", "sample"}
		for _, forbidden := range prodForbidden {
			if strings.Contains(namelower, forbidden) {
				result.Errors = append(result.Errors, fmt.Sprintf("Production resources must not contain: %s", forbidden))
				result.Valid = false
			}
		}

		// Storage accounts in production must have sufficient entropy
		if resourceType == ResourceTypeStorageAccount && len(name) < 12 {
			result.Warnings = append(result.Warnings, "Production storage accounts should be at least 12 characters for better uniqueness")
		}
	}

	// Development-specific rules
	if ncm.environment == "development" {
		if strings.Contains(namelower, "prod") || strings.Contains(namelower, "production") {
			result.Errors = append(result.Errors, "Development resources must not contain production indicators")
			result.Valid = false
		}
	}

	// Staging-specific rules
	if ncm.environment == "staging" {
		if strings.Contains(namelower, "prod") || strings.Contains(namelower, "production") {
			result.Errors = append(result.Errors, "Staging resources must not contain production indicators")
			result.Valid = false
		}
		if strings.Contains(namelower, "dev") || strings.Contains(namelower, "development") {
			result.Warnings = append(result.Warnings, "Consider avoiding development indicators in staging resources")
		}
	}

	// Global uniqueness requirements
	globallyUnique := []ResourceType{ResourceTypeStorageAccount, ResourceTypeKeyVault}
	for _, uniqueType := range globallyUnique {
		if resourceType == uniqueType {
			// Check for sufficient uniqueness (simplified check)
			if !ncm.hasUniqueSuffix(name) {
				result.Warnings = append(result.Warnings, fmt.Sprintf("%s requires global uniqueness, consider adding unique suffix", resourceType))
			}
		}
	}
}

// hasUniqueSuffix checks if name has unique suffix for global uniqueness
func (ncm *NamingConventionManager) hasUniqueSuffix(name string) bool {
	// Simple check for numeric suffix or longer names that likely have uniqueness
	if len(name) > 15 {
		return true
	}
	
	// Check for numeric suffix
	suffixRegex := regexp.MustCompile(`\d{3,}$`)
	return suffixRegex.MatchString(name)
}

// GenerateName generates compliant name for resource type
func (ncm *NamingConventionManager) GenerateName(resourceType ResourceType, component string, options *NamingOptions) (string, error) {
	_, exists := ncm.conventions[resourceType]
	if !exists {
		return "", fmt.Errorf("no naming convention defined for resource type %s", resourceType)
	}

	var name string
	
	switch resourceType {
	case ResourceTypeResourceGroup:
		region := "eastus2"
		if options != nil && options.Region != "" {
			region = options.Region
		}
		name = fmt.Sprintf("rg-%s-%s-%s", ncm.environment, component, region)
		
	case ResourceTypeSQLServer:
		region := "eastus2"
		suffix := "001"
		if options != nil {
			if options.Region != "" {
				region = options.Region
			}
			if options.Suffix != "" {
				suffix = options.Suffix
			}
		}
		name = fmt.Sprintf("sql-%s-%s-%s-%s", ncm.environment, component, region, suffix)
		
	case ResourceTypeSQLDatabase:
		name = fmt.Sprintf("db-%s-%s", ncm.environment, component)
		
	case ResourceTypeStorageAccount:
		suffix := "001"
		if options != nil && options.Suffix != "" {
			suffix = options.Suffix
		}
		name = fmt.Sprintf("st%s%s%s", ncm.environment, component, suffix)
		
	case ResourceTypeKeyVault:
		suffix := "001"
		if options != nil && options.Suffix != "" {
			suffix = options.Suffix
		}
		name = fmt.Sprintf("kv-%s-%s-%s", ncm.environment, component, suffix)
		
	case ResourceTypeContainerApp:
		name = fmt.Sprintf("ca-%s-%s", ncm.environment, component)
		
	case ResourceTypeContainerEnvironment:
		region := "eastus2"
		if options != nil && options.Region != "" {
			region = options.Region
		}
		name = fmt.Sprintf("cae-%s-%s", ncm.environment, region)
		
	case ResourceTypeLogAnalytics:
		region := "eastus2"
		if options != nil && options.Region != "" {
			region = options.Region
		}
		name = fmt.Sprintf("log-%s-%s", ncm.environment, region)
		
	case ResourceTypeVirtualNetwork:
		region := "eastus2"
		if options != nil && options.Region != "" {
			region = options.Region
		}
		name = fmt.Sprintf("vnet-%s-%s", ncm.environment, region)
		
	case ResourceTypeSubnet:
		name = fmt.Sprintf("snet-%s-%s", ncm.environment, component)
		
	case ResourceTypeSecurityGroup:
		name = fmt.Sprintf("nsg-%s-%s", ncm.environment, component)
		
	default:
		return "", fmt.Errorf("name generation not implemented for resource type %s", resourceType)
	}

	// Validate generated name
	result := ncm.ValidateName(name, resourceType)
	if !result.Valid {
		return "", fmt.Errorf("generated name failed validation: %v", result.Errors)
	}

	return name, nil
}

// NamingOptions provides options for name generation
type NamingOptions struct {
	Region string
	Suffix string
	Tags   map[string]string
}

// GetConvention returns naming convention for resource type
func (ncm *NamingConventionManager) GetConvention(resourceType ResourceType) (*NamingConvention, error) {
	convention, exists := ncm.conventions[resourceType]
	if !exists {
		return nil, fmt.Errorf("no naming convention defined for resource type %s", resourceType)
	}
	return convention, nil
}

// ListConventions returns all naming conventions
func (ncm *NamingConventionManager) ListConventions() map[ResourceType]*NamingConvention {
	return ncm.conventions
}

// ValidateResourceNames validates multiple resource names
func (ncm *NamingConventionManager) ValidateResourceNames(names map[ResourceType]string) map[ResourceType]*NamingValidationResult {
	results := make(map[ResourceType]*NamingValidationResult)
	
	for resourceType, name := range names {
		results[resourceType] = ncm.ValidateName(name, resourceType)
	}
	
	return results
}