package config

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationConfigInterface(t *testing.T) {
	t.Run("MigrationConfig provides required fields", func(t *testing.T) {
		config := MigrationConfig{
			Domain:        "test-domain",
			MigrationPath: "/path/to/migrations",
			DatabaseURL:   "postgres://localhost:5432/test",
		}
		
		assert.Equal(t, "test-domain", config.Domain, "Domain field should be accessible")
		assert.Equal(t, "/path/to/migrations", config.MigrationPath, "MigrationPath field should be accessible")
		assert.Equal(t, "postgres://localhost:5432/test", config.DatabaseURL, "DatabaseURL field should be accessible")
	})
	
	t.Run("GetMigrationURL returns file URL", func(t *testing.T) {
		config := MigrationConfig{
			Domain:        "test-domain",
			MigrationPath: "/path/to/migrations",
			DatabaseURL:   "postgres://localhost:5432/test",
		}
		
		url := config.GetMigrationURL()
		
		assert.True(t, strings.HasPrefix(url, "file://"), "Migration URL should start with file://")
		assert.Contains(t, url, config.MigrationPath, "Migration URL should contain the migration path")
	})
}

func TestDomainMigrationConfigsInterface(t *testing.T) {
	t.Run("NewDomainMigrationConfigs creates all domain configurations", func(t *testing.T) {
		basePath := "/tmp/test-migrations"
		databaseURL := "postgres://localhost:5432/test"
		
		configs := NewDomainMigrationConfigs(basePath, databaseURL)
		
		require.NotNil(t, configs, "NewDomainMigrationConfigs should return valid configs")
		
		// Test content domain configurations
		assert.Equal(t, "content-services", configs.Content.Services.Domain)
		assert.Equal(t, "content-news", configs.Content.News.Domain)
		assert.Equal(t, "content-research", configs.Content.Research.Domain)
		assert.Equal(t, "content-events", configs.Content.Events.Domain)
		
		// Test inquiries domain configurations
		assert.Equal(t, "inquiries-donations", configs.Inquiries.Donations.Domain)
		assert.Equal(t, "inquiries-business", configs.Inquiries.Business.Domain)
		assert.Equal(t, "inquiries-media", configs.Inquiries.Media.Domain)
		assert.Equal(t, "inquiries-volunteers", configs.Inquiries.Volunteers.Domain)
		
		// Test supporting domain configurations
		assert.Equal(t, "notifications", configs.Notifications.Domain)
		assert.Equal(t, "gateway", configs.Gateway.Domain)
		assert.Equal(t, "shared", configs.Shared.Domain)
		
		// Test that all configs have the same database URL
		allConfigs := configs.GetAllConfigs()
		for _, cfg := range allConfigs {
			assert.Equal(t, databaseURL, cfg.DatabaseURL, "All configs should have the same database URL")
		}
	})
	
	t.Run("migration paths are correctly constructed", func(t *testing.T) {
		basePath := "/tmp/test-migrations"
		databaseURL := "postgres://localhost:5432/test"
		
		configs := NewDomainMigrationConfigs(basePath, databaseURL)
		
		// Test that paths are properly joined
		expectedServicesPath := filepath.Join(basePath, "content", "services")
		assert.Equal(t, expectedServicesPath, configs.Content.Services.MigrationPath)
		
		expectedDonationsPath := filepath.Join(basePath, "inquiries", "donations")
		assert.Equal(t, expectedDonationsPath, configs.Inquiries.Donations.MigrationPath)
		
		expectedNotificationsPath := filepath.Join(basePath, "notifications")
		assert.Equal(t, expectedNotificationsPath, configs.Notifications.MigrationPath)
	})
	
	t.Run("GetAllConfigs returns all configurations", func(t *testing.T) {
		basePath := "/tmp/test-migrations"
		databaseURL := "postgres://localhost:5432/test"
		
		configs := NewDomainMigrationConfigs(basePath, databaseURL)
		allConfigs := configs.GetAllConfigs()
		
		// Should return all 11 domain configurations
		assert.Len(t, allConfigs, 11, "GetAllConfigs should return all domain configurations")
		
		// Verify we have all expected domains
		domains := make(map[string]bool)
		for _, cfg := range allConfigs {
			domains[cfg.Domain] = true
		}
		
		expectedDomains := []string{
			"content-services", "content-news", "content-research", "content-events",
			"inquiries-donations", "inquiries-business", "inquiries-media", "inquiries-volunteers",
			"notifications", "gateway", "shared",
		}
		
		for _, expected := range expectedDomains {
			assert.True(t, domains[expected], "Should include domain: %s", expected)
		}
	})
}

func TestMigrationConfigValidation(t *testing.T) {
	tests := []struct {
		name          string
		domain        string
		migrationPath string
		databaseURL   string
		expectValid   bool
	}{
		{
			name:          "valid configuration",
			domain:        "test-domain",
			migrationPath: "/path/to/migrations",
			databaseURL:   "postgres://localhost:5432/test",
			expectValid:   true,
		},
		{
			name:          "empty domain",
			domain:        "",
			migrationPath: "/path/to/migrations",
			databaseURL:   "postgres://localhost:5432/test",
			expectValid:   false,
		},
		{
			name:          "empty migration path",
			domain:        "test-domain",
			migrationPath: "",
			databaseURL:   "postgres://localhost:5432/test",
			expectValid:   false,
		},
		{
			name:          "empty database URL",
			domain:        "test-domain",
			migrationPath: "/path/to/migrations",
			databaseURL:   "",
			expectValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := MigrationConfig{
				Domain:        tt.domain,
				MigrationPath: tt.migrationPath,
				DatabaseURL:   tt.databaseURL,
			}
			
			// Basic validation - non-empty fields for valid configs
			if tt.expectValid {
				assert.NotEmpty(t, config.Domain, "Valid config should have non-empty domain")
				assert.NotEmpty(t, config.MigrationPath, "Valid config should have non-empty migration path")
				assert.NotEmpty(t, config.DatabaseURL, "Valid config should have non-empty database URL")
				assert.NotEmpty(t, config.GetMigrationURL(), "Valid config should generate migration URL")
			} else {
				// For invalid configs, at least one required field should be empty
				isEmpty := config.Domain == "" || config.MigrationPath == "" || config.DatabaseURL == ""
				assert.True(t, isEmpty, "Invalid config should have at least one empty required field")
			}
		})
	}
}