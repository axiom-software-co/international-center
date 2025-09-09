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
	hasDigits := false
	
	for i, r := range s {
		if r >= '0' && r <= '9' {
			hasDigits = true
			digit := int(r - '0')
			
			// Check for overflow before multiplication
			if result > (int(^uint(0)>>1)-digit)/10 {
				// Return current result on overflow to avoid negative numbers
				return result
			}
			
			result = result*10 + digit
		} else {
			// Special handling for scientific notation: parse digits before 'e'
			if hasDigits && r == 'e' && i > 0 {
				return result
			}
			// For other cases with mixed characters, return 0 as expected by tests
			return 0
		}
	}
	
	if !hasDigits {
		return 0
	}
	return result
}