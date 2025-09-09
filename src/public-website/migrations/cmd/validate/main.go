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

	// Create validator
	validator := validation.NewSchemaValidator(db, *validationPath)

	// Run validation
	var results []validation.ValidationResult
	if *domain != "" {
		result, err := validator.ValidateDomain(*domain)
		if err != nil {
			log.Fatalf("Failed to validate domain %s: %v", *domain, err)
		}
		results = []validation.ValidationResult{result}
	} else {
		results, err = validator.ValidateAllDomains()
		if err != nil {
			log.Fatalf("Failed to validate domains: %v", err)
		}
	}

	// Generate and print report
	report := validator.GenerateReport(results)
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