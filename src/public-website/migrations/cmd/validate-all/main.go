package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/axiom-software-co/international-center/src/public-website/migrations/validation"
	_ "github.com/lib/pq"
)

func main() {
	var (
		dbURL          = flag.String("db-url", "", "Database connection URL")
		validationPath = flag.String("validation-path", "", "Path to validation schema files")
		modelsBasePath = flag.String("models-path", "", "Base path to Go domain models")
		domain         = flag.String("domain", "", "Specific domain to validate (optional)")
		verbose        = flag.Bool("verbose", false, "Enable verbose output")
		schemaOnly     = flag.Bool("schema-only", false, "Only validate database schemas")
		modelsOnly     = flag.Bool("models-only", false, "Only validate Go models")
	)
	flag.Parse()

	if *dbURL == "" {
		*dbURL = os.Getenv("DATABASE_URL")
		if *dbURL == "" {
			log.Fatal("Database URL must be provided via --db-url flag or DATABASE_URL environment variable")
		}
	}

	if *validationPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get working directory: %v", err)
		}
		*validationPath = filepath.Join(wd, "migrations", "validation")
	}

	if *modelsBasePath == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get working directory: %v", err)
		}
		backendPath := filepath.Join(wd, "..", "..", "backend", "internal")
		*modelsBasePath = backendPath
	}

	if *verbose {
		fmt.Printf("Database URL: %s\n", *dbURL)
		fmt.Printf("Validation Path: %s\n", *validationPath)
		fmt.Printf("Models Base Path: %s\n", *modelsBasePath)
		if *domain != "" {
			fmt.Printf("Target Domain: %s\n", *domain)
		}
	}

	// Connect to database
	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	var allValid = true

	// Schema Validation
	if !*modelsOnly {
		if *verbose {
			fmt.Println("\n=== SCHEMA VALIDATION ===")
		}

		schemaValidator := validation.NewSchemaValidator(db, *validationPath)

		var schemaResults []validation.ValidationResult
		if *domain != "" {
			result, err := schemaValidator.ValidateDomain(*domain)
			if err != nil {
				log.Fatalf("Failed to validate schema for domain %s: %v", *domain, err)
			}
			schemaResults = []validation.ValidationResult{result}
		} else {
			schemaResults, err = schemaValidator.ValidateAllDomains()
			if err != nil {
				log.Fatalf("Failed to validate schemas: %v", err)
			}
		}

		schemaReport := schemaValidator.GenerateReport(schemaResults)
		fmt.Print(schemaReport)

		// Check schema validation results
		for _, result := range schemaResults {
			if !result.Valid {
				allValid = false
			}
		}
	}

	// Model Validation
	if !*schemaOnly {
		if *verbose {
			fmt.Println("\n=== MODEL VALIDATION ===")
		}

		schemaValidator := validation.NewSchemaValidator(db, *validationPath)
		modelValidator := validation.NewModelValidator(db, schemaValidator)

		var modelResults []validation.ModelValidationResult
		if *domain != "" {
			domainPath := filepath.Join(*modelsBasePath, *domain)
			result, err := modelValidator.ValidateDomain(*domain, domainPath)
			if err != nil {
				log.Fatalf("Failed to validate models for domain %s: %v", *domain, err)
			}
			modelResults = []validation.ModelValidationResult{result}
		} else {
			modelResults, err = modelValidator.ValidateAllDomains(*modelsBasePath)
			if err != nil {
				log.Fatalf("Failed to validate models: %v", err)
			}
		}

		modelReport := modelValidator.GenerateReport(modelResults)
		fmt.Print(modelReport)

		// Check model validation results
		for _, result := range modelResults {
			if !result.Valid {
				allValid = false
			}
		}
	}

	// Overall summary
	if *verbose {
		fmt.Println("\n=== OVERALL VALIDATION SUMMARY ===")
		if allValid {
			fmt.Println("✓ All validations passed successfully")
		} else {
			fmt.Println("✗ Some validations failed - see details above")
		}
	}

	// Set exit code
	if !allValid {
		if *verbose {
			fmt.Fprintf(os.Stderr, "\nValidation failed - see errors above\n")
		}
		os.Exit(1)
	} else {
		if *verbose {
			fmt.Fprintf(os.Stderr, "\nAll validations passed successfully\n")
		}
		os.Exit(0)
	}
}