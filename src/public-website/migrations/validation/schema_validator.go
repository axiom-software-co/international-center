package validation

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

type SchemaValidator struct {
	db             *sql.DB
	validationPath string
}

type ValidationResult struct {
	Domain     string
	Valid      bool
	Errors     []string
	Warnings   []string
	TableCount int
	IndexCount int
}

type SchemaComparison struct {
	Domain             string
	MissingTables      []string
	ExtraTables        []string
	MissingIndexes     []string
	ExtraIndexes       []string
	MissingConstraints []string
	ExtraConstraints   []string
	ColumnMismatches   []ColumnMismatch
}

type ColumnMismatch struct {
	Table        string
	Column       string
	ExpectedType string
	ActualType   string
	Issue        string
}

func NewSchemaValidator(db *sql.DB, validationPath string) *SchemaValidator {
	return &SchemaValidator{
		db:             db,
		validationPath: validationPath,
	}
}

func (sv *SchemaValidator) ValidateAllDomains() ([]ValidationResult, error) {
	domains := []string{"content_services", "content_news", "content_research", "content_events", "inquiries", "notifications", "gateway", "shared"}
	results := make([]ValidationResult, 0, len(domains))

	for _, domain := range domains {
		result, err := sv.ValidateDomain(domain)
		if err != nil {
			return nil, fmt.Errorf("failed to validate domain %s: %w", domain, err)
		}
		results = append(results, result)
	}

	return results, nil
}

func (sv *SchemaValidator) ValidateDomain(domain string) (ValidationResult, error) {
	result := ValidationResult{
		Domain: domain,
		Valid:  true,
		Errors: make([]string, 0),
	}

	// Read validation schema file
	schemaPath := filepath.Join(sv.validationPath, domain+"_schema.sql")
	expectedSchema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to read validation schema: %v", err))
		return result, nil
	}

	// Parse expected schema
	expectedTables, expectedIndexes, expectedConstraints, err := sv.parseSchema(string(expectedSchema))
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse validation schema: %v", err))
		return result, nil
	}

	// Get actual schema from database
	actualTables, err := sv.getActualTables(domain)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to retrieve actual tables: %v", err))
		return result, nil
	}

	actualIndexes, err := sv.getActualIndexes(domain)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to retrieve actual indexes: %v", err))
		return result, nil
	}

	actualConstraints, err := sv.getActualConstraints(domain)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to retrieve actual constraints: %v", err))
		return result, nil
	}

	// Compare schemas
	comparison := sv.compareSchemas(domain, expectedTables, actualTables, expectedIndexes, actualIndexes, expectedConstraints, actualConstraints)

	// Populate result
	result.TableCount = len(actualTables)
	result.IndexCount = len(actualIndexes)

	if len(comparison.MissingTables) > 0 {
		result.Valid = false
		for _, table := range comparison.MissingTables {
			result.Errors = append(result.Errors, fmt.Sprintf("Missing table: %s", table))
		}
	}

	if len(comparison.ExtraTables) > 0 {
		for _, table := range comparison.ExtraTables {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Extra table found: %s", table))
		}
	}

	if len(comparison.MissingIndexes) > 0 {
		result.Valid = false
		for _, index := range comparison.MissingIndexes {
			result.Errors = append(result.Errors, fmt.Sprintf("Missing index: %s", index))
		}
	}

	if len(comparison.MissingConstraints) > 0 {
		result.Valid = false
		for _, constraint := range comparison.MissingConstraints {
			result.Errors = append(result.Errors, fmt.Sprintf("Missing constraint: %s", constraint))
		}
	}

	if len(comparison.ColumnMismatches) > 0 {
		result.Valid = false
		for _, mismatch := range comparison.ColumnMismatches {
			result.Errors = append(result.Errors, fmt.Sprintf("Column mismatch in %s.%s: %s", mismatch.Table, mismatch.Column, mismatch.Issue))
		}
	}

	return result, nil
}

func (sv *SchemaValidator) parseSchema(schema string) (map[string][]string, []string, []string, error) {
	tables := make(map[string][]string)
	indexes := make([]string, 0)
	constraints := make([]string, 0)

	lines := strings.Split(schema, "\n")
	var currentTable string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(strings.ToUpper(line), "CREATE TABLE") {
			// Extract table name
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				currentTable = strings.Trim(parts[2], "(")
				tables[currentTable] = make([]string, 0)
			}
		} else if strings.HasPrefix(strings.ToUpper(line), "CREATE INDEX") {
			// Extract index name
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				indexes = append(indexes, parts[2])
			}
		} else if strings.Contains(strings.ToUpper(line), "CONSTRAINT") && currentTable != "" {
			// Extract constraint
			parts := strings.Fields(line)
			for i, part := range parts {
				if strings.ToUpper(part) == "CONSTRAINT" && i+1 < len(parts) {
					constraints = append(constraints, parts[i+1])
					break
				}
			}
		} else if currentTable != "" && strings.Contains(line, " ") && !strings.HasPrefix(line, "--") && line != "" {
			// This might be a column definition
			parts := strings.Fields(line)
			if len(parts) >= 2 && !strings.HasPrefix(strings.ToUpper(parts[0]), "CREATE") {
				columnName := strings.Trim(parts[0], ",")
				if columnName != "" && !strings.HasPrefix(columnName, "(") && !strings.HasPrefix(columnName, ")") {
					tables[currentTable] = append(tables[currentTable], columnName)
				}
			}
		}
	}

	return tables, indexes, constraints, nil
}

func (sv *SchemaValidator) getActualTables(domain string) (map[string][]string, error) {
	tables := make(map[string][]string)

	// Get table names for domain
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		AND table_name SIMILAR TO $1
	`
	
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

	rows, err := sv.db.Query(query, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}

		// Get columns for this table
		columns, err := sv.getTableColumns(tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
		}
		
		tables[tableName] = columns
	}

	return tables, nil
}

func (sv *SchemaValidator) getTableColumns(tableName string) ([]string, error) {
	query := `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = 'public'
		AND table_name = $1
		ORDER BY ordinal_position
	`

	rows, err := sv.db.Query(query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	columns := make([]string, 0)
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, fmt.Errorf("failed to scan column name: %w", err)
		}
		columns = append(columns, columnName)
	}

	return columns, nil
}

func (sv *SchemaValidator) getActualIndexes(domain string) ([]string, error) {
	query := `
		SELECT indexname
		FROM pg_indexes
		WHERE schemaname = 'public'
		AND indexname SIMILAR TO $1
	`

	var pattern string
	switch domain {
	case "content_services":
		pattern = "idx_(service|featured_service)%"
	case "content_news":
		pattern = "idx_(news|featured_news)%"
	case "content_research":
		pattern = "idx_(research|featured_research)%"
	case "content_events":
		pattern = "idx_(event|featured_event)%"
	case "inquiries":
		pattern = "idx_(donations|business|media|volunteer)%"
	case "notifications":
		pattern = "idx_notification%"
	case "gateway":
		pattern = "idx_(users|roles|user_roles)%"
	case "shared":
		pattern = "idx_(audit|correlation)%"
	default:
		return nil, fmt.Errorf("unknown domain: %s", domain)
	}

	rows, err := sv.db.Query(query, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	indexes := make([]string, 0)
	for rows.Next() {
		var indexName string
		if err := rows.Scan(&indexName); err != nil {
			return nil, fmt.Errorf("failed to scan index name: %w", err)
		}
		indexes = append(indexes, indexName)
	}

	return indexes, nil
}

func (sv *SchemaValidator) getActualConstraints(domain string) ([]string, error) {
	query := `
		SELECT constraint_name
		FROM information_schema.table_constraints
		WHERE table_schema = 'public'
		AND constraint_type IN ('CHECK', 'UNIQUE', 'FOREIGN KEY')
		AND table_name SIMILAR TO $1
	`

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

	rows, err := sv.db.Query(query, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query constraints: %w", err)
	}
	defer rows.Close()

	constraints := make([]string, 0)
	for rows.Next() {
		var constraintName string
		if err := rows.Scan(&constraintName); err != nil {
			return nil, fmt.Errorf("failed to scan constraint name: %w", err)
		}
		constraints = append(constraints, constraintName)
	}

	return constraints, nil
}

func (sv *SchemaValidator) compareSchemas(domain string, expectedTables map[string][]string, actualTables map[string][]string, expectedIndexes []string, actualIndexes []string, expectedConstraints []string, actualConstraints []string) SchemaComparison {
	comparison := SchemaComparison{
		Domain:             domain,
		MissingTables:      make([]string, 0),
		ExtraTables:        make([]string, 0),
		MissingIndexes:     make([]string, 0),
		ExtraIndexes:       make([]string, 0),
		MissingConstraints: make([]string, 0),
		ExtraConstraints:   make([]string, 0),
		ColumnMismatches:   make([]ColumnMismatch, 0),
	}

	// Compare tables
	for expectedTable := range expectedTables {
		if _, exists := actualTables[expectedTable]; !exists {
			comparison.MissingTables = append(comparison.MissingTables, expectedTable)
		}
	}

	for actualTable := range actualTables {
		if _, exists := expectedTables[actualTable]; !exists {
			comparison.ExtraTables = append(comparison.ExtraTables, actualTable)
		}
	}

	// Compare indexes (simplified)
	expectedIndexMap := make(map[string]bool)
	for _, index := range expectedIndexes {
		expectedIndexMap[index] = true
	}

	actualIndexMap := make(map[string]bool)
	for _, index := range actualIndexes {
		actualIndexMap[index] = true
	}

	for expectedIndex := range expectedIndexMap {
		if !actualIndexMap[expectedIndex] {
			comparison.MissingIndexes = append(comparison.MissingIndexes, expectedIndex)
		}
	}

	for actualIndex := range actualIndexMap {
		if !expectedIndexMap[actualIndex] {
			comparison.ExtraIndexes = append(comparison.ExtraIndexes, actualIndex)
		}
	}

	return comparison
}

func (sv *SchemaValidator) GenerateReport(results []ValidationResult) string {
	var report strings.Builder
	
	report.WriteString("Schema Validation Report\n")
	report.WriteString("========================\n\n")

	totalDomains := len(results)
	validDomains := 0
	
	for _, result := range results {
		if result.Valid {
			validDomains++
		}
		
		report.WriteString(fmt.Sprintf("Domain: %s\n", result.Domain))
		report.WriteString(fmt.Sprintf("Status: %s\n", func() string {
			if result.Valid {
				return "VALID"
			}
			return "INVALID"
		}()))
		report.WriteString(fmt.Sprintf("Tables: %d\n", result.TableCount))
		report.WriteString(fmt.Sprintf("Indexes: %d\n", result.IndexCount))
		
		if len(result.Errors) > 0 {
			report.WriteString("Errors:\n")
			for _, err := range result.Errors {
				report.WriteString(fmt.Sprintf("  - %s\n", err))
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
	
	report.WriteString(fmt.Sprintf("Summary: %d/%d domains valid\n", validDomains, totalDomains))
	
	return report.String()
}