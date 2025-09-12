package testing

import (
	"context"
	"fmt"
	"time"
)

// ContractFirstTestingPatterns provides standardized contract-first testing patterns for all modules
type ContractFirstTestingPatterns struct {
	environmentValidator *SharedEnvironmentValidator
	contractValidator    *SimpleContractValidator
	daprTestRunner       *DaprServiceMeshTestRunner
}

// NewContractFirstTestingPatterns creates standardized contract-first testing patterns
func NewContractFirstTestingPatterns() *ContractFirstTestingPatterns {
	return &ContractFirstTestingPatterns{
		environmentValidator: NewSharedEnvironmentValidator(),
		contractValidator:    NewSimpleContractValidator(),
		daprTestRunner:       NewDaprServiceMeshTestRunner(),
	}
}

// StandardizedTDDPhase represents standardized TDD phases across all modules
type StandardizedTDDPhase string

const (
	StandardizedPhaseRed     StandardizedTDDPhase = "RED"     // Failing tests that define requirements
	StandardizedPhaseGreen   StandardizedTDDPhase = "GREEN"   // Implementation to make tests pass
	StandardizedPhaseRefactor StandardizedTDDPhase = "REFACTOR" // Optimization while maintaining tests
)

// ContractFirstTestSuite provides standardized test suite for all modules
type ContractFirstTestSuite struct {
	ModuleName    string
	Phase         StandardizedTDDPhase
	TestType      string // "unit", "integration", "contract", "e2e"
	RequiredServices []string
	patterns      *ContractFirstTestingPatterns
}

// NewContractFirstTestSuite creates a standardized contract-first test suite
func NewContractFirstTestSuite(moduleName string, phase StandardizedTDDPhase, testType string) *ContractFirstTestSuite {
	return &ContractFirstTestSuite{
		ModuleName: moduleName,
		Phase:      phase,
		TestType:   testType,
		patterns:   NewContractFirstTestingPatterns(),
	}
}

// SetRequiredServices sets the services required for this test suite
func (suite *ContractFirstTestSuite) SetRequiredServices(services []string) {
	suite.RequiredServices = services
}

// ValidateTestEnvironment validates the test environment using standardized patterns
func (suite *ContractFirstTestSuite) ValidateTestEnvironment(ctx context.Context) error {
	// Only validate environment for integration tests
	if suite.TestType == "integration" || suite.TestType == "contract" || suite.TestType == "e2e" {
		return suite.patterns.environmentValidator.ValidateEnvironmentPrerequisites(ctx, suite.RequiredServices)
	}
	
	// Unit tests don't need environment validation
	return nil
}

// RunContractValidation runs standardized contract validation
func (suite *ContractFirstTestSuite) RunContractValidation(ctx context.Context) (ContractValidationSummary, error) {
	// Contract validation is relevant for all test types
	summary := suite.patterns.contractValidator.GenerateValidationSummary(ctx)
	return summary, nil
}

// RunDaprServiceMeshValidation runs standardized Dapr service mesh validation
func (suite *ContractFirstTestSuite) RunDaprServiceMeshValidation(ctx context.Context) ([]DaprTestingResult, error) {
	// Dapr validation only for backend and integration tests
	if suite.TestType == "integration" && (suite.ModuleName == "backend" || suite.ModuleName == "deployment") {
		return suite.patterns.daprTestRunner.RunComprehensiveDaprTesting(ctx), nil
	}
	
	return nil, nil
}

// StandardizedTestResult holds standardized test results across all modules
type StandardizedTestResult struct {
	ModuleName           string
	Phase                StandardizedTDDPhase
	TestType             string
	EnvironmentValid     bool
	ContractCompliance   float64
	DaprServicesHealthy  int
	DaprServicesFailing  int
	TotalDuration        time.Duration
	CriticalIssues       []string
	MinorIssues          []string
	Success              bool
}

// RunStandardizedTesting runs comprehensive standardized testing
func (suite *ContractFirstTestSuite) RunStandardizedTesting(ctx context.Context) StandardizedTestResult {
	start := time.Now()
	
	result := StandardizedTestResult{
		ModuleName: suite.ModuleName,
		Phase:      suite.Phase,
		TestType:   suite.TestType,
	}
	
	// Step 1: Environment validation
	if err := suite.ValidateTestEnvironment(ctx); err != nil {
		result.EnvironmentValid = false
		result.CriticalIssues = append(result.CriticalIssues, fmt.Sprintf("Environment validation failed: %v", err))
	} else {
		result.EnvironmentValid = true
	}
	
	// Step 2: Contract validation
	contractSummary, err := suite.RunContractValidation(ctx)
	if err != nil {
		result.CriticalIssues = append(result.CriticalIssues, fmt.Sprintf("Contract validation failed: %v", err))
	} else {
		result.ContractCompliance = contractSummary.CompliancePercent
		
		for _, issue := range contractSummary.CriticalIssues {
			result.CriticalIssues = append(result.CriticalIssues, issue)
		}
		for _, issue := range contractSummary.MinorIssues {
			result.MinorIssues = append(result.MinorIssues, issue)
		}
	}
	
	// Step 3: Dapr validation (if applicable)
	daprResults, err := suite.RunDaprServiceMeshValidation(ctx)
	if err != nil {
		result.CriticalIssues = append(result.CriticalIssues, fmt.Sprintf("Dapr validation failed: %v", err))
	} else if daprResults != nil {
		for _, daprResult := range daprResults {
			if daprResult.Success {
				result.DaprServicesHealthy++
			} else {
				result.DaprServicesFailing++
				result.CriticalIssues = append(result.CriticalIssues, fmt.Sprintf("Dapr service %s: %v", daprResult.ServiceName, daprResult.Error))
			}
		}
	}
	
	// Calculate overall success
	result.Success = result.EnvironmentValid && len(result.CriticalIssues) == 0
	result.TotalDuration = time.Since(start)
	
	return result
}

// PrintStandardizedTestResult prints a standardized test result
func (result *StandardizedTestResult) PrintStandardizedTestResult() string {
	output := fmt.Sprintf("\n=== STANDARDIZED TEST RESULT: %s ===\n", result.ModuleName)
	output += fmt.Sprintf("Phase: %s | Type: %s | Duration: %v\n", result.Phase, result.TestType, result.TotalDuration)
	output += fmt.Sprintf("Environment: %t | Contract Compliance: %.1f%%\n", result.EnvironmentValid, result.ContractCompliance)
	
	if result.DaprServicesHealthy > 0 || result.DaprServicesFailing > 0 {
		output += fmt.Sprintf("Dapr Services: %d healthy, %d failing\n", result.DaprServicesHealthy, result.DaprServicesFailing)
	}
	
	if result.Success {
		output += "✅ OVERALL: SUCCESS\n"
	} else {
		output += "❌ OVERALL: ISSUES IDENTIFIED\n"
	}
	
	if len(result.CriticalIssues) > 0 {
		output += "\n❌ CRITICAL ISSUES:\n"
		for _, issue := range result.CriticalIssues {
			output += fmt.Sprintf("  %s\n", issue)
		}
	}
	
	if len(result.MinorIssues) > 0 {
		output += "\n⚠️  MINOR ISSUES:\n"
		for _, issue := range result.MinorIssues {
			output += fmt.Sprintf("  %s\n", issue)
		}
	}
	
	return output
}

// Module-specific standardized test suite factories

// NewBackendContractFirstTestSuite creates standardized test suite for backend module
func NewBackendContractFirstTestSuite(phase StandardizedTDDPhase, testType string) *ContractFirstTestSuite {
	suite := NewContractFirstTestSuite("backend", phase, testType)
	
	// Backend module standard required services
	backendServices := []string{"dapr-control-plane", "content", "inquiries", "notifications", "public-gateway", "admin-gateway"}
	suite.SetRequiredServices(backendServices)
	
	return suite
}

// NewDeploymentContractFirstTestSuite creates standardized test suite for deployment module
func NewDeploymentContractFirstTestSuite(phase StandardizedTDDPhase, testType string) *ContractFirstTestSuite {
	suite := NewContractFirstTestSuite("deployment", phase, testType)
	
	// Deployment module standard required services (infrastructure focus)
	deploymentServices := []string{"dapr-control-plane", "vault", "azurite"}
	suite.SetRequiredServices(deploymentServices)
	
	return suite
}

// NewFrontendContractFirstTestSuite creates standardized test suite for frontend modules
func NewFrontendContractFirstTestSuite(phase StandardizedTDDPhase, testType string, frontendApp string) *ContractFirstTestSuite {
	suite := NewContractFirstTestSuite(fmt.Sprintf("frontend-%s", frontendApp), phase, testType)
	
	// Frontend integration tests require backend services
	if testType == "integration" {
		frontendServices := []string{"public-gateway", "admin-gateway"}
		suite.SetRequiredServices(frontendServices)
	}
	// Unit tests don't require services (proper isolation)
	
	return suite
}

// NewContractsModuleTestSuite creates standardized test suite for contracts module
func NewContractsModuleTestSuite(phase StandardizedTDDPhase, testType string) *ContractFirstTestSuite {
	suite := NewContractFirstTestSuite("contracts", phase, testType)
	
	// Contracts module tests require backend services for contract validation
	contractServices := []string{"content", "inquiries", "notifications", "public-gateway", "admin-gateway"}
	suite.SetRequiredServices(contractServices)
	
	return suite
}

// NewMigrationsModuleTestSuite creates standardized test suite for migrations module
func NewMigrationsModuleTestSuite(phase StandardizedTDDPhase, testType string) *ContractFirstTestSuite {
	suite := NewContractFirstTestSuite("migrations", phase, testType)
	
	// Migrations module tests require database and related services
	migrationServices := []string{"postgresql", "dapr-control-plane"}
	suite.SetRequiredServices(migrationServices)
	
	return suite
}