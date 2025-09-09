package validation

import (
	"database/sql"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

type ModelValidator struct {
	db         *sql.DB
	schemaValidator *SchemaValidator
}

type ModelValidationResult struct {
	Domain           string
	ModelPath        string
	Valid            bool
	ModelCount       int
	Errors           []string
	Warnings         []string
	TypeMismatches   []TypeMismatch
	MissingFields    []MissingField
	ExtraFields      []ExtraField
}

type TypeMismatch struct {
	Model        string
	Field        string
	GoType       string
	DatabaseType string
	Table        string
	Column       string
}

type MissingField struct {
	Model  string
	Table  string
	Column string
	Issue  string
}

type ExtraField struct {
	Model string
	Field string
	Issue string
}

type GoModel struct {
	Name       string
	Package    string
	Fields     map[string]GoField
	TableName  string // Inferred from struct name or tags
	FilePath   string
}

type GoField struct {
	Name     string
	Type     string
	Tags     map[string]string
	Optional bool
	Column   string // Mapped column name
}

type DatabaseColumn struct {
	Name         string
	Type         string
	Nullable     bool
	Default      *string
	IsPrimaryKey bool
}

func NewModelValidator(db *sql.DB, schemaValidator *SchemaValidator) *ModelValidator {
	return &ModelValidator{
		db:              db,
		schemaValidator: schemaValidator,
	}
}

func (mv *ModelValidator) ValidateAllDomains(modelsBasePath string) ([]ModelValidationResult, error) {
	domains := map[string]string{
		"content_services": filepath.Join(modelsBasePath, "content", "services"),
		"content_news":     filepath.Join(modelsBasePath, "content", "news"),
		"content_research": filepath.Join(modelsBasePath, "content", "research"),
		"content_events":   filepath.Join(modelsBasePath, "content", "events"),
		"inquiries":        filepath.Join(modelsBasePath, "inquiries"),
		"notifications":    filepath.Join(modelsBasePath, "notifications"),
		"gateway":          filepath.Join(modelsBasePath, "gateway"),
		"shared":           filepath.Join(modelsBasePath, "shared"),
	}

	results := make([]ModelValidationResult, 0, len(domains))

	for domain, path := range domains {
		result, err := mv.ValidateDomain(domain, path)
		if err != nil {
			return nil, fmt.Errorf("failed to validate domain %s: %w", domain, err)
		}
		results = append(results, result)
	}

	return results, nil
}

func (mv *ModelValidator) ValidateDomain(domain, modelsPath string) (ModelValidationResult, error) {
	result := ModelValidationResult{
		Domain:           domain,
		ModelPath:        modelsPath,
		Valid:            true,
		Errors:           make([]string, 0),
		Warnings:         make([]string, 0),
		TypeMismatches:   make([]TypeMismatch, 0),
		MissingFields:    make([]MissingField, 0),
		ExtraFields:      make([]ExtraField, 0),
	}

	// Parse Go models from the domain package
	models, err := mv.parseGoModels(modelsPath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse Go models: %v", err))
		return result, nil
	}

	result.ModelCount = len(models)

	if len(models) == 0 {
		result.Warnings = append(result.Warnings, "No Go models found in domain package")
		return result, nil
	}

	// Get database schema information
	dbTables, err := mv.getDatabaseTables(domain)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get database tables: %v", err))
		return result, nil
	}

	// Validate each model against database schema
	for _, model := range models {
		mv.validateModel(model, dbTables, &result)
	}

	return result, nil
}

func (mv *ModelValidator) parseGoModels(modelsPath string) ([]GoModel, error) {
	models := make([]GoModel, 0)

	// Parse all Go files in the directory
	fset := token.NewFileSet()
	
	// Try to parse the directory - if it doesn't exist, return empty slice
	pkgs, err := parser.ParseDir(fset, modelsPath, nil, parser.ParseComments)
	if err != nil {
		// Directory might not exist or have no Go files
		return models, nil
	}

	for _, pkg := range pkgs {
		for filePath, file := range pkg.Files {
			fileModels := mv.parseFileModels(file, pkg.Name, filePath)
			models = append(models, fileModels...)
		}
	}

	return models, nil
}

func (mv *ModelValidator) parseFileModels(file *ast.File, packageName, filePath string) []GoModel {
	models := make([]GoModel, 0)

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.TypeSpec:
			if structType, ok := node.Type.(*ast.StructType); ok {
				model := mv.parseStructType(node.Name.Name, structType, packageName, filePath)
				if model != nil {
					models = append(models, *model)
				}
			}
		}
		return true
	})

	return models
}

func (mv *ModelValidator) parseStructType(name string, structType *ast.StructType, packageName, filePath string) *GoModel {
	model := &GoModel{
		Name:      name,
		Package:   packageName,
		Fields:    make(map[string]GoField),
		TableName: mv.inferTableName(name),
		FilePath:  filePath,
	}

	for _, field := range structType.Fields.List {
		for _, fieldName := range field.Names {
			goField := GoField{
				Name:   fieldName.Name,
				Type:   mv.getTypeString(field.Type),
				Tags:   mv.parseFieldTags(field.Tag),
				Column: mv.inferColumnName(fieldName.Name, field.Tag),
			}

			// Check if field is optional (pointer type or has omitempty tag)
			goField.Optional = mv.isOptionalField(field.Type, goField.Tags)

			model.Fields[fieldName.Name] = goField
		}
	}

	// Only return models that look like database entities
	if mv.isDBModel(model) {
		return model
	}

	return nil
}

func (mv *ModelValidator) getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + mv.getTypeString(t.X)
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", mv.getTypeString(t.X), t.Sel.Name)
	case *ast.ArrayType:
		return "[]" + mv.getTypeString(t.Elt)
	default:
		return "unknown"
	}
}

func (mv *ModelValidator) parseFieldTags(tag *ast.BasicLit) map[string]string {
	tags := make(map[string]string)
	
	if tag == nil {
		return tags
	}

	// Remove backticks and parse tag string
	tagValue := strings.Trim(tag.Value, "`")
	
	// Simple tag parsing - in real implementation would use reflect.StructTag
	if strings.Contains(tagValue, "db:") {
		// Extract db tag value
		parts := strings.Split(tagValue, "db:")
		if len(parts) > 1 {
			dbTag := strings.Trim(strings.Split(parts[1], " ")[0], "\"")
			tags["db"] = dbTag
		}
	}

	if strings.Contains(tagValue, "json:") {
		// Extract json tag value
		parts := strings.Split(tagValue, "json:")
		if len(parts) > 1 {
			jsonTag := strings.Trim(strings.Split(parts[1], " ")[0], "\"")
			tags["json"] = jsonTag
		}
	}

	return tags
}

func (mv *ModelValidator) inferTableName(structName string) string {
	// Convert PascalCase to snake_case and pluralize
	// This is a simplified version - real implementation would be more sophisticated
	result := ""
	for i, r := range structName {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result += "_"
		}
		result += strings.ToLower(string(r))
	}
	
	// Simple pluralization
	if strings.HasSuffix(result, "y") {
		result = result[:len(result)-1] + "ies"
	} else if strings.HasSuffix(result, "s") {
		result += "es"
	} else {
		result += "s"
	}
	
	return result
}

func (mv *ModelValidator) inferColumnName(fieldName string, tag *ast.BasicLit) string {
	// Check for db tag first
	if tag != nil {
		tags := mv.parseFieldTags(tag)
		if dbTag, exists := tags["db"]; exists && dbTag != "-" {
			return dbTag
		}
	}

	// Convert PascalCase to snake_case
	result := ""
	for i, r := range fieldName {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result += "_"
		}
		result += strings.ToLower(string(r))
	}
	
	return result
}

func (mv *ModelValidator) isOptionalField(expr ast.Expr, tags map[string]string) bool {
	// Check if it's a pointer type
	if _, ok := expr.(*ast.StarExpr); ok {
		return true
	}

	// Check for omitempty in json tag
	if jsonTag, exists := tags["json"]; exists {
		return strings.Contains(jsonTag, "omitempty")
	}

	return false
}

func (mv *ModelValidator) isDBModel(model *GoModel) bool {
	// Check if model has typical database fields
	hasID := false
	hasTimestamp := false

	for fieldName, field := range model.Fields {
		lowerName := strings.ToLower(fieldName)
		lowerType := strings.ToLower(field.Type)

		if strings.Contains(lowerName, "id") {
			hasID = true
		}

		if strings.Contains(lowerType, "time") || strings.Contains(lowerName, "created") || strings.Contains(lowerName, "updated") {
			hasTimestamp = true
		}
	}

	// Consider it a DB model if it has ID-like field or timestamp fields
	return hasID || hasTimestamp
}

func (mv *ModelValidator) getDatabaseTables(domain string) (map[string]map[string]DatabaseColumn, error) {
	tables := make(map[string]map[string]DatabaseColumn)

	// Get table names for domain (reuse logic from schema validator)
	var pattern string
	switch domain {
	case "content_services":
		pattern = "(service_categories|services|featured_services)"
	case "content_news":
		pattern = "(news_categories|news|featured_news|news_external_sources)"
	case "content_research":
		pattern = "(research_categories|research|featured_research)"
	case "content_events":
		pattern = "(event_categories|events|featured_events|event_registrations)"
	case "inquiries":
		pattern = "(donations_inquiries|business_inquiries|media_inquiries|volunteer_applications)"
	case "notifications":
		pattern = "(notification_subscribers)"
	case "gateway":
		pattern = "(users|roles|user_roles)"
	case "shared":
		pattern = "(audit_events|correlation_tracking)"
	default:
		return nil, fmt.Errorf("unknown domain: %s", domain)
	}

	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		AND table_name SIMILAR TO $1
	`

	rows, err := mv.db.Query(query, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}

		columns, err := mv.getTableColumns(tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
		}

		tables[tableName] = columns
	}

	return tables, nil
}

func (mv *ModelValidator) getTableColumns(tableName string) (map[string]DatabaseColumn, error) {
	query := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default,
			CASE WHEN tc.constraint_type = 'PRIMARY KEY' THEN true ELSE false END as is_primary_key
		FROM information_schema.columns c
		LEFT JOIN information_schema.key_column_usage kcu ON 
			c.table_name = kcu.table_name AND c.column_name = kcu.column_name
		LEFT JOIN information_schema.table_constraints tc ON 
			kcu.constraint_name = tc.constraint_name AND tc.constraint_type = 'PRIMARY KEY'
		WHERE c.table_schema = 'public'
		AND c.table_name = $1
		ORDER BY c.ordinal_position
	`

	rows, err := mv.db.Query(query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	columns := make(map[string]DatabaseColumn)
	for rows.Next() {
		var col DatabaseColumn
		var nullable string
		var isPK *bool

		err := rows.Scan(&col.Name, &col.Type, &nullable, &col.Default, &isPK)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		col.Nullable = nullable == "YES"
		if isPK != nil {
			col.IsPrimaryKey = *isPK
		}

		columns[col.Name] = col
	}

	return columns, nil
}

func (mv *ModelValidator) validateModel(model GoModel, dbTables map[string]map[string]DatabaseColumn, result *ModelValidationResult) {
	// Find matching database table
	dbTable, exists := dbTables[model.TableName]
	if !exists {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("No database table found for model %s (expected table: %s)", model.Name, model.TableName))
		return
	}

	// Check each model field against database columns
	for fieldName, field := range model.Fields {
		dbColumn, exists := dbTable[field.Column]
		if !exists {
			result.MissingFields = append(result.MissingFields, MissingField{
				Model:  model.Name,
				Table:  model.TableName,
				Column: field.Column,
				Issue:  fmt.Sprintf("Model field %s maps to column %s which doesn't exist", fieldName, field.Column),
			})
			result.Valid = false
			continue
		}

		// Validate type compatibility
		if !mv.isTypeCompatible(field.Type, dbColumn.Type) {
			result.TypeMismatches = append(result.TypeMismatches, TypeMismatch{
				Model:        model.Name,
				Field:        fieldName,
				GoType:       field.Type,
				DatabaseType: dbColumn.Type,
				Table:        model.TableName,
				Column:       field.Column,
			})
			result.Valid = false
		}

		// Validate nullability
		if !field.Optional && dbColumn.Nullable {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Field %s.%s is required in Go but nullable in database", model.Name, fieldName))
		}
	}

	// Check for database columns not represented in model
	for columnName, dbColumn := range dbTable {
		found := false
		for _, field := range model.Fields {
			if field.Column == columnName {
				found = true
				break
			}
		}

		if !found && !mv.isAuditColumn(columnName) {
			result.ExtraFields = append(result.ExtraFields, ExtraField{
				Model: model.Name,
				Field: columnName,
				Issue: fmt.Sprintf("Database column %s not represented in model %s", columnName, model.Name),
			})
			
			if !dbColumn.Nullable {
				result.Valid = false
			}
		}
	}
}

func (mv *ModelValidator) isTypeCompatible(goType, dbType string) bool {
	// Normalize types for comparison
	goType = strings.TrimPrefix(goType, "*") // Remove pointer indicator
	dbType = strings.ToLower(dbType)

	typeMap := map[string][]string{
		"string":        {"character varying", "varchar", "text", "character", "char"},
		"int":           {"integer", "int", "int4"},
		"int64":         {"bigint", "int8"},
		"int32":         {"integer", "int", "int4"},
		"bool":          {"boolean", "bool"},
		"time.Time":     {"timestamp with time zone", "timestamptz", "timestamp", "date"},
		"uuid.UUID":     {"uuid"},
		"float64":       {"double precision", "float8", "numeric", "decimal"},
		"float32":       {"real", "float4"},
		"[]string":      {"text[]", "character varying[]", "varchar[]"},
		"[]byte":        {"bytea"},
		"interface{}":   {"jsonb", "json"},
	}

	if compatibleTypes, exists := typeMap[goType]; exists {
		for _, compatibleType := range compatibleTypes {
			if strings.Contains(dbType, compatibleType) {
				return true
			}
		}
	}

	return false
}

func (mv *ModelValidator) isAuditColumn(columnName string) bool {
	auditColumns := []string{
		"created_at", "updated_at", "created_on", "modified_on",
		"created_by", "updated_by", "modified_by",
		"is_deleted", "deleted_at", "deleted_on", "deleted_by",
	}

	lowerName := strings.ToLower(columnName)
	for _, auditCol := range auditColumns {
		if lowerName == auditCol {
			return true
		}
	}

	return false
}

func (mv *ModelValidator) GenerateReport(results []ModelValidationResult) string {
	var report strings.Builder
	
	report.WriteString("Model Validation Report\n")
	report.WriteString("=======================\n\n")

	totalDomains := len(results)
	validDomains := 0
	totalModels := 0
	
	for _, result := range results {
		if result.Valid {
			validDomains++
		}
		totalModels += result.ModelCount
		
		report.WriteString(fmt.Sprintf("Domain: %s\n", result.Domain))
		report.WriteString(fmt.Sprintf("Model Path: %s\n", result.ModelPath))
		report.WriteString(fmt.Sprintf("Status: %s\n", func() string {
			if result.Valid {
				return "VALID"
			}
			return "INVALID"
		}()))
		report.WriteString(fmt.Sprintf("Models Found: %d\n", result.ModelCount))
		
		if len(result.Errors) > 0 {
			report.WriteString("Errors:\n")
			for _, err := range result.Errors {
				report.WriteString(fmt.Sprintf("  - %s\n", err))
			}
		}

		if len(result.TypeMismatches) > 0 {
			report.WriteString("Type Mismatches:\n")
			for _, mismatch := range result.TypeMismatches {
				report.WriteString(fmt.Sprintf("  - %s.%s: Go type '%s' vs DB type '%s'\n", 
					mismatch.Model, mismatch.Field, mismatch.GoType, mismatch.DatabaseType))
			}
		}

		if len(result.MissingFields) > 0 {
			report.WriteString("Missing Fields:\n")
			for _, missing := range result.MissingFields {
				report.WriteString(fmt.Sprintf("  - %s\n", missing.Issue))
			}
		}

		if len(result.ExtraFields) > 0 {
			report.WriteString("Extra Database Columns:\n")
			for _, extra := range result.ExtraFields {
				report.WriteString(fmt.Sprintf("  - %s\n", extra.Issue))
			}
		}
		
		if len(result.Warnings) > 0 {
			report.WriteString("Warnings:\n")
			for _, warning := range result.Warnings {
				report.WriteString(fmt.Sprintf("  - %s\n", warning))
			}
		}
		
		report.WriteString("\n")
	}
	
	report.WriteString(fmt.Sprintf("Summary: %d/%d domains valid, %d total models\n", validDomains, totalDomains, totalModels))
	
	return report.String()
}