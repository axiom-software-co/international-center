package dapr

import "os"

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as integer with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		// Simple conversion, in production would handle errors
		if parsed := parseInt(value); parsed > 0 {
			return parsed
		}
	}
	return defaultValue
}

// parseInt simple integer parsing helper
func parseInt(s string) int {
	// Simple integer parsing - in production would use strconv.Atoi
	result := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		} else {
			return 0
		}
	}
	return result
}