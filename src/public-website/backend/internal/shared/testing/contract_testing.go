package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"io"
)

// ContractTestSuite provides contract testing capabilities for TDD workflows
type ContractTestSuite struct {
	adminSpec    *openapi3.T
	publicSpec   *openapi3.T
	adminRouter  routers.Router
	publicRouter routers.Router
}

// NewContractTestSuite creates a new contract test suite
func NewContractTestSuite(adminSpecPath, publicSpecPath string) (*ContractTestSuite, error) {
	ctx := context.Background()
	
	// Load specifications
	adminLoader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	adminSpec, err := adminLoader.LoadFromFile(adminSpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load admin spec: %w", err)
	}
	
	publicLoader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	publicSpec, err := publicLoader.LoadFromFile(publicSpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public spec: %w", err)
	}
	
	// Validate specifications
	if err := adminSpec.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invalid admin spec: %w", err)
	}
	
	if err := publicSpec.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invalid public spec: %w", err)
	}
	
	// Create routers
	adminRouter, err := gorillamux.NewRouter(adminSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin router: %w", err)
	}
	
	publicRouter, err := gorillamux.NewRouter(publicSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to create public router: %w", err)
	}
	
	return &ContractTestSuite{
		adminSpec:    adminSpec,
		publicSpec:   publicSpec,
		adminRouter:  adminRouter,
		publicRouter: publicRouter,
	}, nil
}

// ContractTestCase represents a test case for contract validation
type ContractTestCase struct {
	Name        string
	Method      string
	Path        string
	RequestBody interface{}
	Headers     map[string]string
	ExpectedStatus int
}

// ValidateHandlerCompliance validates that a handler complies with the OpenAPI contract
func (suite *ContractTestSuite) ValidateHandlerCompliance(handler http.Handler, testCase ContractTestCase) error {
	// Determine which spec to use based on path
	isAdminAPI := strings.HasPrefix(testCase.Path, "/admin/api/")
	
	var router routers.Router
	if isAdminAPI {
		router = suite.adminRouter
	} else {
		router = suite.publicRouter
	}
	
	// Create test request
	var requestBody []byte
	if testCase.RequestBody != nil {
		var err error
		requestBody, err = json.Marshal(testCase.RequestBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}
	
	req := httptest.NewRequest(testCase.Method, testCase.Path, bytes.NewReader(requestBody))
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range testCase.Headers {
		req.Header.Set(key, value)
	}
	
	// Find route in specification
	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		return fmt.Errorf("route not found in spec: %s %s: %w", testCase.Method, testCase.Path, err)
	}
	
	// Validate request against spec
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}
	
	ctx := context.Background()
	if err := openapi3filter.ValidateRequest(ctx, requestValidationInput); err != nil {
		return fmt.Errorf("request validation failed: %w", err)
	}
	
	// Execute request
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	// Validate response against spec
	if err := suite.validateResponse(rr, route, req); err != nil {
		return fmt.Errorf("response validation failed: %w", err)
	}
	
	// Validate expected status code
	if testCase.ExpectedStatus != 0 && rr.Code != testCase.ExpectedStatus {
		return fmt.Errorf("expected status %d, got %d", testCase.ExpectedStatus, rr.Code)
	}
	
	return nil
}

// validateResponse validates an HTTP response against the OpenAPI specification
func (suite *ContractTestSuite) validateResponse(rr *httptest.ResponseRecorder, route *routers.Route, originalRequest *http.Request) error {
	// Create response for validation
	response := &http.Response{
		StatusCode: rr.Code,
		Header:     rr.Header(),
		Body:       nil, // Will be set below if there's content
	}
	
	// Set response body if present
	responseBody := rr.Body.Bytes()
	if len(responseBody) > 0 {
		response.Body = io.NopCloser(strings.NewReader(string(responseBody)))
	}
	
	// Create validation input
	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: originalRequest,
			Route:   route,
		},
		Status: rr.Code,
		Header: response.Header,
	}
	
	// Set body for validation if present
	if len(responseBody) > 0 {
		responseValidationInput.SetBodyBytes(responseBody)
	}
	
	// Validate response
	ctx := context.Background()
	if err := openapi3filter.ValidateResponse(ctx, responseValidationInput); err != nil {
		return fmt.Errorf("response validation error: %w", err)
	}
	
	return nil
}

// ContractTestRunner provides utilities for running contract tests in TDD workflows
type ContractTestRunner struct {
	suite *ContractTestSuite
}

// NewContractTestRunner creates a new contract test runner
func NewContractTestRunner(adminSpecPath, publicSpecPath string) (*ContractTestRunner, error) {
	suite, err := NewContractTestSuite(adminSpecPath, publicSpecPath)
	if err != nil {
		return nil, err
	}
	
	return &ContractTestRunner{
		suite: suite,
	}, nil
}

// TestInquiryEndpointCompliance tests inquiry endpoint compliance with contracts
func (runner *ContractTestRunner) TestInquiryEndpointCompliance(handler http.Handler) []error {
	var errors []error
	
	// Test cases for inquiry endpoints
	testCases := []ContractTestCase{
		{
			Name:           "GetInquiries - valid request",
			Method:         "GET", 
			Path:           "/admin/api/v1/inquiries",
			Headers:        map[string]string{"X-User-ID": "admin-test-user"},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "GetInquiryById - valid UUID",
			Method:         "GET",
			Path:           "/admin/api/v1/inquiries/550e8400-e29b-41d4-a716-446655440000",
			Headers:        map[string]string{"X-User-ID": "admin-test-user"},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "UpdateInquiryStatus - valid status update",
			Method: "PUT",
			Path:   "/admin/api/v1/inquiries/550e8400-e29b-41d4-a716-446655440000",
			RequestBody: map[string]interface{}{
				"status": "in_progress",
				"notes":  "Processing inquiry",
			},
			Headers:        map[string]string{"X-User-ID": "admin-test-user"},
			ExpectedStatus: http.StatusOK,
		},
	}
	
	// Run each test case
	for _, testCase := range testCases {
		if err := runner.suite.ValidateHandlerCompliance(handler, testCase); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", testCase.Name, err))
		}
	}
	
	return errors
}

// ContractValidationResult holds the results of contract validation
type ContractValidationResult struct {
	TestName    string
	Endpoint    string
	Method      string
	Passed      bool
	Error       error
	Duration    time.Duration
	Timestamp   time.Time
}

// RunContractValidationSuite runs a full contract validation suite and returns results
func (runner *ContractTestRunner) RunContractValidationSuite(handler http.Handler) []ContractValidationResult {
	var results []ContractValidationResult
	
	// Get all inquiry contract validation errors
	start := time.Now()
	errors := runner.TestInquiryEndpointCompliance(handler)
	duration := time.Since(start)
	
	if len(errors) == 0 {
		results = append(results, ContractValidationResult{
			TestName:  "InquiryEndpointCompliance",
			Endpoint:  "/admin/api/v1/inquiries/*",
			Method:    "ALL",
			Passed:    true,
			Error:     nil,
			Duration:  duration,
			Timestamp: time.Now(),
		})
	} else {
		for _, err := range errors {
			results = append(results, ContractValidationResult{
				TestName:  "InquiryEndpointCompliance",
				Endpoint:  "/admin/api/v1/inquiries/*",
				Method:    "ALL",
				Passed:    false,
				Error:     err,
				Duration:  duration,
				Timestamp: time.Now(),
			})
		}
	}
	
	return results
}

// TDDContractPhase represents the TDD phases with contract validation
type TDDContractPhase string

const (
	TDDPhaseRed     TDDContractPhase = "red"     // Write failing tests including contract tests
	TDDPhaseGreen   TDDContractPhase = "green"   // Make tests pass including contract compliance
	TDDPhaseRefactor TDDContractPhase = "refactor" // Refactor while maintaining contract compliance
)

// TDDContractWorkflow provides contract-aware TDD workflow utilities
type TDDContractWorkflow struct {
	testRunner *ContractTestRunner
	phase      TDDContractPhase
}

// NewTDDContractWorkflow creates a new TDD workflow with contract validation
func NewTDDContractWorkflow(adminSpecPath, publicSpecPath string) (*TDDContractWorkflow, error) {
	runner, err := NewContractTestRunner(adminSpecPath, publicSpecPath)
	if err != nil {
		return nil, err
	}
	
	return &TDDContractWorkflow{
		testRunner: runner,
		phase:      TDDPhaseRed,
	}, nil
}

// SetPhase sets the current TDD phase
func (workflow *TDDContractWorkflow) SetPhase(phase TDDContractPhase) {
	workflow.phase = phase
}

// ValidatePhaseCompliance validates that the current implementation meets the phase requirements
func (workflow *TDDContractWorkflow) ValidatePhaseCompliance(handler http.Handler) error {
	results := workflow.testRunner.RunContractValidationSuite(handler)
	
	switch workflow.phase {
	case TDDPhaseRed:
		// In red phase, we expect some tests to fail (this is normal)
		// Just validate that the contract structure is correct
		return nil
		
	case TDDPhaseGreen:
		// In green phase, all contract tests must pass
		for _, result := range results {
			if !result.Passed {
				return fmt.Errorf("contract validation failed in green phase: %w", result.Error)
			}
		}
		return nil
		
	case TDDPhaseRefactor:
		// In refactor phase, contract compliance must be maintained
		for _, result := range results {
			if !result.Passed {
				return fmt.Errorf("contract compliance broken during refactor: %w", result.Error)
			}
		}
		return nil
		
	default:
		return fmt.Errorf("unknown TDD phase: %s", workflow.phase)
	}
}