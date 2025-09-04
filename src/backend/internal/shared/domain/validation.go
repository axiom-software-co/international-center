package domain

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Common validation patterns
var (
	slugRegex     = regexp.MustCompile(`^[a-z0-9-]+$`)
	httpsURLRegex = regexp.MustCompile(`^https://[^\s]+$`)
)

// ValidateUUID checks if a string is a valid UUID
func ValidateUUID(id string) error {
	if id == "" {
		return NewValidationFieldError("id", "UUID cannot be empty")
	}
	
	if _, err := uuid.Parse(id); err != nil {
		return NewValidationFieldError("id", "invalid UUID format")
	}
	
	return nil
}

// ValidateSlug checks if a string is a valid slug
func ValidateSlug(slug string) error {
	if slug == "" {
		return NewValidationFieldError("slug", "slug cannot be empty")
	}
	
	if len(slug) > 255 {
		return NewValidationFieldError("slug", "slug cannot exceed 255 characters")
	}
	
	if !slugRegex.MatchString(slug) {
		return NewValidationFieldError("slug", "slug must contain only lowercase letters, numbers, and hyphens")
	}
	
	return nil
}

// ValidateTitle checks if a title is valid
func ValidateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return NewValidationFieldError("title", "title cannot be empty")
	}
	
	if len(title) > 255 {
		return NewValidationFieldError("title", "title cannot exceed 255 characters")
	}
	
	return nil
}

// ValidateRequiredString checks if a required string field is valid
func ValidateRequiredString(fieldName, value string) error {
	if strings.TrimSpace(value) == "" {
		return NewValidationFieldError(fieldName, fieldName+" cannot be empty")
	}
	
	return nil
}

// ValidateRequiredStringWithLength checks if a required string field is valid and within length limit
func ValidateRequiredStringWithLength(fieldName, value string, maxLength int) error {
	if err := ValidateRequiredString(fieldName, value); err != nil {
		return err
	}
	
	if len(value) > maxLength {
		return NewValidationFieldError(fieldName, fieldName+" cannot exceed "+strconv.Itoa(maxLength)+" characters")
	}
	
	return nil
}

// ValidateHTTPSURL checks if a URL is a valid HTTPS URL
func ValidateHTTPSURL(fieldName, url string) error {
	if url == "" {
		return nil // Optional field
	}
	
	if !httpsURLRegex.MatchString(url) {
		return NewValidationFieldError(fieldName, fieldName+" must be a valid HTTPS URL")
	}
	
	return nil
}

// GenerateSlug creates a URL-friendly slug from a title
func GenerateSlug(input string) string {
	if input == "" {
		return ""
	}
	
	// Convert to lowercase
	slug := strings.ToLower(input)
	
	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	
	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	
	// Limit length
	if len(slug) > 255 {
		slug = slug[:255]
		slug = strings.TrimRight(slug, "-")
	}
	
	return slug
}

// ValidateEnum validates that a value is one of the allowed enum values
func ValidateEnum(fieldName, value string, allowedValues []string) error {
	if value == "" {
		return NewValidationFieldError(fieldName, fieldName+" cannot be empty")
	}
	
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}
	
	return NewValidationFieldError(fieldName, "invalid "+fieldName+" value: "+value)
}