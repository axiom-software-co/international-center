package tests

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	envLoadOnce sync.Once
	envLoadErr  error
)

// loadDevelopmentEnv loads environment variables from .env.development file
func loadDevelopmentEnv() error {
	// Get the project root directory (two levels up from infrastructure/tests)
	projectRoot := filepath.Join("..", "..")
	envFile := filepath.Join(projectRoot, ".env.development")

	file, err := os.Open(envFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			
			// Only set if not already set (allows override from actual environment)
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
	
	return scanner.Err()
}

// initTestEnvironment loads development environment variables once per test run
func initTestEnvironment(t *testing.T) {
	envLoadOnce.Do(func() {
		envLoadErr = loadDevelopmentEnv()
	})
	
	if envLoadErr != nil {
		t.Fatalf("Failed to load development environment: %v", envLoadErr)
	}
}

// requireEnv retrieves environment variable or fails test if missing
// Automatically loads .env.development on first call
func requireEnv(t *testing.T, key string) string {
	// Initialize environment if not already done
	initTestEnvironment(t)
	
	value := os.Getenv(key)
	if value == "" {
		t.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}