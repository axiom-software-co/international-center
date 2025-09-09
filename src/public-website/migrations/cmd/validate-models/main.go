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
		// Try to find backend models directory
		backendPath := filepath.Join(wd, "..", "..", "backend", "internal")
		if _, err := os.Stat(backendPath); os.IsNotExist(err) {
			if *verbose {
				fmt.Printf("Backend models directory not found at %s, will check for empty results\n", backendPath)
			}
		}
		*modelsBasePath = backendPath
	}

	if *verbose {
		fmt.Printf("Database URL: %s\n", *dbURL)
		fmt.Printf("Validation Path: %s\n", *validationPath)
		fmt.Printf("Models Base Path: %s\n", *modelsBasePath)
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

	// Create schema validator first
	schemaValidator := validation.NewSchemaValidator(db, *validationPath)
	
	// Create model validator
	modelValidator := validation.NewModelValidator(db, schemaValidator)

	// Run model validation
	var results []validation.ModelValidationResult
	if *domain != "" {
		domainPath := filepath.Join(*modelsBasePath, *domain)
		result, err := modelValidator.ValidateDomain(*domain, domainPath)
		if err != nil {
			log.Fatalf("Failed to validate domain %s: %v", *domain, err)
		}
		results = []validation.ModelValidationResult{result}
	} else {
		results, err = modelValidator.ValidateAllDomains(*modelsBasePath)
		if err != nil {
			log.Fatalf("Failed to validate models: %v", err)
		}
	}

	// Generate and print report
	report := modelValidator.GenerateReport(results)
	fmt.Print(report)

	// Set exit code based on validation results
	hasErrors := false
	for _, result := range results {
		if !result.Valid {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		if *verbose {
			fmt.Fprintf(os.Stderr, "\nModel validation failed - see errors above\n")
		}
		os.Exit(1)
	} else {
		if *verbose {
			fmt.Fprintf(os.Stderr, "\nAll model validations passed successfully\n")
		}
		os.Exit(0)
	}
}