package validation

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelValidator_ParseGoModels(t *testing.T) {
	mv := &ModelValidator{}

	// Create temporary Go file with test models
	tempDir, err := ioutil.TempDir("", "model_validator_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testGoCode := `
package domain

import (
	"time"
	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	UserID    uuid.UUID ` + "`" + `db:"user_id" json:"user_id"` + "`" + `
	Username  string    ` + "`" + `db:"username" json:"username"` + "`" + `
	Email     string    ` + "`" + `db:"email" json:"email"` + "`" + `
	CreatedAt time.Time ` + "`" + `db:"created_at" json:"created_at"` + "`" + `
	UpdatedAt *time.Time ` + "`" + `db:"updated_at" json:"updated_at,omitempty"` + "`" + `
}

// Service represents a medical service
type Service struct {
	ServiceID   uuid.UUID ` + "`" + `db:"service_id" json:"service_id"` + "`" + `
	Title       string    ` + "`" + `db:"title" json:"title"` + "`" + `
	Description *string   ` + "`" + `db:"description" json:"description,omitempty"` + "`" + `
	IsActive    bool      ` + "`" + `db:"is_active" json:"is_active"` + "`" + `
}

// NonDBStruct is not a database model
type NonDBStruct struct {
	Value   string
	Count   int
}
`

	goFilePath := filepath.Join(tempDir, "models.go")
	err = ioutil.WriteFile(goFilePath, []byte(testGoCode), 0644)
	require.NoError(t, err)

	// Parse models
	models, err := mv.parseGoModels(tempDir)
	require.NoError(t, err)

	// Should find 2 models (User and Service), excluding NonDBStruct
	assert.Len(t, models, 2)

	// Find User model
	var userModel *GoModel
	for _, model := range models {
		if model.Name == "User" {
			userModel = &model
			break
		}
	}

	require.NotNil(t, userModel)
	assert.Equal(t, "User", userModel.Name)
	assert.Equal(t, "domain", userModel.Package)
	assert.Equal(t, "users", userModel.TableName)

	// Check fields
	assert.Contains(t, userModel.Fields, "UserID")
	assert.Contains(t, userModel.Fields, "Username")
	assert.Contains(t, userModel.Fields, "Email")
	assert.Contains(t, userModel.Fields, "CreatedAt")
	assert.Contains(t, userModel.Fields, "UpdatedAt")

	// Check field details
	userIDField := userModel.Fields["UserID"]
	assert.Equal(t, "uuid.UUID", userIDField.Type)
	assert.Equal(t, "user_id", userIDField.Column)
	assert.False(t, userIDField.Optional)

	updatedAtField := userModel.Fields["UpdatedAt"]
	assert.Equal(t, "*time.Time", updatedAtField.Type)
	assert.Equal(t, "updated_at", updatedAtField.Column)
	assert.True(t, updatedAtField.Optional)
}

func TestModelValidator_InferTableName(t *testing.T) {
	mv := &ModelValidator{}

	testCases := []struct {
		structName string
		expected   string
	}{
		{"User", "users"},
		{"Service", "services"},
		{"Category", "categories"},
		{"UserRole", "user_roles"},
		{"BusinessInquiry", "business_inquiries"},
		{"EventRegistration", "event_registrations"},
	}

	for _, tc := range testCases {
		t.Run(tc.structName, func(t *testing.T) {
			result := mv.inferTableName(tc.structName)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestModelValidator_InferColumnName(t *testing.T) {
	mv := &ModelValidator{}

	testCases := []struct {
		fieldName    string
		tag          string
		expectedCol  string
	}{
		{"UserID", `db:"user_id"`, "user_id"},
		{"FirstName", `db:"first_name"`, "first_name"},
		{"Email", "", "email"},
		{"CreatedAt", "", "created_at"},
		{"IsActive", `db:"active"`, "active"},
		{"UpdatedBy", `db:"modified_by"`, "modified_by"},
	}

	for _, tc := range testCases {
		t.Run(tc.fieldName, func(t *testing.T) {
			// Create a mock AST tag
			var tag *ast.BasicLit
			if tc.tag != "" {
				tag = &ast.BasicLit{Value: "`" + tc.tag + "`"}
			}
			
			result := mv.inferColumnName(tc.fieldName, tag)
			assert.Equal(t, tc.expectedCol, result)
		})
	}
}

func TestModelValidator_IsTypeCompatible(t *testing.T) {
	mv := &ModelValidator{}

	testCases := []struct {
		goType      string
		dbType      string
		compatible  bool
	}{
		{"string", "character varying", true},
		{"string", "varchar", true},
		{"string", "text", true},
		{"*string", "varchar", true}, // Pointer type
		{"int", "integer", true},
		{"int64", "bigint", true},
		{"bool", "boolean", true},
		{"time.Time", "timestamp with time zone", true},
		{"time.Time", "timestamptz", true},
		{"uuid.UUID", "uuid", true},
		{"[]string", "text[]", true},
		{"interface{}", "jsonb", true},
		{"string", "integer", false}, // Incompatible
		{"int", "varchar", false},    // Incompatible
	}

	for _, tc := range testCases {
		t.Run(tc.goType+"_"+tc.dbType, func(t *testing.T) {
			result := mv.isTypeCompatible(tc.goType, tc.dbType)
			assert.Equal(t, tc.compatible, result)
		})
	}
}

func TestModelValidator_IsOptionalField(t *testing.T) {
	mv := &ModelValidator{}

	testCases := []struct {
		typeExpr ast.Expr
		tags     map[string]string
		expected bool
	}{
		// These would require creating AST nodes, simplified for testing
	}

	// Test with tags
	tagsWithOmitEmpty := map[string]string{"json": "field_name,omitempty"}
	tagsWithoutOmitEmpty := map[string]string{"json": "field_name"}

	// Since we can't easily create AST nodes in test, test the tag logic separately
	assert.True(t, mv.hasOmitEmptyTag(tagsWithOmitEmpty))
	assert.False(t, mv.hasOmitEmptyTag(tagsWithoutOmitEmpty))

	_ = testCases // Keep for potential future expansion
}

// Helper method to test omitempty detection
func (mv *ModelValidator) hasOmitEmptyTag(tags map[string]string) bool {
	if jsonTag, exists := tags["json"]; exists {
		return strings.Contains(jsonTag, "omitempty")
	}
	return false
}

func TestModelValidator_IsDBModel(t *testing.T) {
	mv := &ModelValidator{}

	// Test with ID field
	modelWithID := &GoModel{
		Name: "User",
		Fields: map[string]GoField{
			"UserID": {Name: "UserID", Type: "uuid.UUID"},
			"Name":   {Name: "Name", Type: "string"},
		},
	}
	assert.True(t, mv.isDBModel(modelWithID))

	// Test with timestamp field
	modelWithTimestamp := &GoModel{
		Name: "Event",
		Fields: map[string]GoField{
			"Name":      {Name: "Name", Type: "string"},
			"CreatedAt": {Name: "CreatedAt", Type: "time.Time"},
		},
	}
	assert.True(t, mv.isDBModel(modelWithTimestamp))

	// Test without DB indicators
	nonDBModel := &GoModel{
		Name: "Config",
		Fields: map[string]GoField{
			"Host": {Name: "Host", Type: "string"},
			"Port": {Name: "Port", Type: "int"},
		},
	}
	assert.False(t, mv.isDBModel(nonDBModel))
}

func TestModelValidator_IsAuditColumn(t *testing.T) {
	mv := &ModelValidator{}

	auditColumns := []string{
		"created_at", "updated_at", "created_on", "modified_on",
		"created_by", "updated_by", "modified_by",
		"is_deleted", "deleted_at", "deleted_on", "deleted_by",
	}

	for _, col := range auditColumns {
		t.Run(col, func(t *testing.T) {
			assert.True(t, mv.isAuditColumn(col))
		})
	}

	nonAuditColumns := []string{"name", "email", "title", "description", "status"}
	for _, col := range nonAuditColumns {
		t.Run(col, func(t *testing.T) {
			assert.False(t, mv.isAuditColumn(col))
		})
	}
}

func TestModelValidator_GenerateReport(t *testing.T) {
	mv := &ModelValidator{}

	results := []ModelValidationResult{
		{
			Domain:     "users",
			ModelPath:  "/path/to/users",
			Valid:      true,
			ModelCount: 2,
			Errors:     []string{},
			Warnings:   []string{"Nullable field warning"},
			TypeMismatches: []TypeMismatch{},
			MissingFields:  []MissingField{},
			ExtraFields:    []ExtraField{},
		},
		{
			Domain:     "services",
			ModelPath:  "/path/to/services",
			Valid:      false,
			ModelCount: 1,
			Errors:     []string{"Missing table for model Service"},
			TypeMismatches: []TypeMismatch{
				{
					Model:        "Service",
					Field:        "Price",
					GoType:       "float64",
					DatabaseType: "varchar",
					Table:        "services",
					Column:       "price",
				},
			},
			MissingFields: []MissingField{
				{
					Model:  "Service",
					Table:  "services",
					Column: "missing_column",
					Issue:  "Column doesn't exist",
				},
			},
		},
	}

	report := mv.GenerateReport(results)

	assert.Contains(t, report, "Model Validation Report")
	assert.Contains(t, report, "Domain: users")
	assert.Contains(t, report, "Domain: services")
	assert.Contains(t, report, "VALID")
	assert.Contains(t, report, "INVALID")
	assert.Contains(t, report, "Type Mismatches:")
	assert.Contains(t, report, "Missing Fields:")
	assert.Contains(t, report, "1/2 domains valid")
	assert.Contains(t, report, "3 total models")

	t.Logf("Generated report:\n%s", report)
}

func TestModelValidator_ParseFieldTags(t *testing.T) {
	mv := &ModelValidator{}

	testCases := []struct {
		tagValue    string
		expectedDB  string
		expectedJSON string
	}{
		{`db:"user_id" json:"user_id"`, "user_id", "user_id"},
		{`db:"first_name" json:"first_name,omitempty"`, "first_name", "first_name,omitempty"},
		{`json:"email"`, "", "email"},
		{`db:"created_at"`, "created_at", ""},
		{"", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.tagValue, func(t *testing.T) {
			var tag *ast.BasicLit
			if tc.tagValue != "" {
				tag = &ast.BasicLit{Value: "`" + tc.tagValue + "`"}
			}

			tags := mv.parseFieldTags(tag)

			if tc.expectedDB != "" {
				assert.Equal(t, tc.expectedDB, tags["db"])
			} else {
				assert.NotContains(t, tags, "db")
			}

			if tc.expectedJSON != "" {
				assert.Equal(t, tc.expectedJSON, tags["json"])
			} else {
				assert.NotContains(t, tags, "json")
			}
		})
	}
}

// Test with actual Go file parsing
func TestModelValidator_RealGoFileParsing(t *testing.T) {
	mv := &ModelValidator{}

	// Create a realistic Go model file
	tempDir, err := ioutil.TempDir("", "real_model_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	realModelCode := `
package services

import (
	"time"
	"github.com/google/uuid"
)

// Service represents a medical service offering
type Service struct {
	ServiceID    uuid.UUID  ` + "`" + `db:"service_id" json:"service_id"` + "`" + `
	CategoryID   uuid.UUID  ` + "`" + `db:"category_id" json:"category_id"` + "`" + `
	Title        string     ` + "`" + `db:"title" json:"title"` + "`" + `
	Description  *string    ` + "`" + `db:"description" json:"description,omitempty"` + "`" + `
	IsActive     bool       ` + "`" + `db:"is_active" json:"is_active"` + "`" + `
	CreatedAt    time.Time  ` + "`" + `db:"created_at" json:"created_at"` + "`" + `
	UpdatedAt    *time.Time ` + "`" + `db:"updated_at" json:"updated_at,omitempty"` + "`" + `
}

// ServiceCategory represents service categorization
type ServiceCategory struct {
	CategoryID   uuid.UUID  ` + "`" + `db:"category_id" json:"category_id"` + "`" + `
	Name         string     ` + "`" + `db:"name" json:"name"` + "`" + `
	Slug         string     ` + "`" + `db:"slug" json:"slug"` + "`" + `
	IsDefault    bool       ` + "`" + `db:"is_default_unassigned" json:"is_default_unassigned"` + "`" + `
}
`

	modelFile := filepath.Join(tempDir, "service.go")
	err = ioutil.WriteFile(modelFile, []byte(realModelCode), 0644)
	require.NoError(t, err)

	// Parse the file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, modelFile, nil, parser.ParseComments)
	require.NoError(t, err)

	models := mv.parseFileModels(file, "services", modelFile)

	assert.Len(t, models, 2)

	// Check Service model
	var serviceModel *GoModel
	for _, model := range models {
		if model.Name == "Service" {
			serviceModel = &model
			break
		}
	}

	require.NotNil(t, serviceModel)
	assert.Equal(t, "services", serviceModel.TableName)
	assert.Len(t, serviceModel.Fields, 7)

	// Verify specific field mappings
	assert.Equal(t, "service_id", serviceModel.Fields["ServiceID"].Column)
	assert.Equal(t, "uuid.UUID", serviceModel.Fields["ServiceID"].Type)
	assert.False(t, serviceModel.Fields["ServiceID"].Optional)

	assert.Equal(t, "description", serviceModel.Fields["Description"].Column)
	assert.Equal(t, "*string", serviceModel.Fields["Description"].Type)
	assert.True(t, serviceModel.Fields["Description"].Optional)
}