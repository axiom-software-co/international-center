package config

import (
	"fmt"
	"path/filepath"
)

// MigrationConfig represents domain-specific migration configuration
type MigrationConfig struct {
	Domain        string
	MigrationPath string
	DatabaseURL   string
}

// DomainMigrationConfigs holds all domain-specific migration configurations
type DomainMigrationConfigs struct {
	Content struct {
		Services  MigrationConfig
		News      MigrationConfig
		Research  MigrationConfig
		Events    MigrationConfig
	}
	Inquiries struct {
		Donations  MigrationConfig
		Business   MigrationConfig
		Media      MigrationConfig
		Volunteers MigrationConfig
	}
	Notifications MigrationConfig
	Gateway       MigrationConfig
	Shared        MigrationConfig
}

// NewDomainMigrationConfigs creates migration configurations for all domains
func NewDomainMigrationConfigs(basePath, databaseURL string) *DomainMigrationConfigs {
	config := &DomainMigrationConfigs{}
	
	// Content domain configurations
	config.Content.Services = MigrationConfig{
		Domain:        "content-services",
		MigrationPath: filepath.Join(basePath, "content", "services"),
		DatabaseURL:   databaseURL,
	}
	config.Content.News = MigrationConfig{
		Domain:        "content-news",
		MigrationPath: filepath.Join(basePath, "content", "news"),
		DatabaseURL:   databaseURL,
	}
	config.Content.Research = MigrationConfig{
		Domain:        "content-research",
		MigrationPath: filepath.Join(basePath, "content", "research"),
		DatabaseURL:   databaseURL,
	}
	config.Content.Events = MigrationConfig{
		Domain:        "content-events",
		MigrationPath: filepath.Join(basePath, "content", "events"),
		DatabaseURL:   databaseURL,
	}
	
	// Inquiries domain configurations
	config.Inquiries.Donations = MigrationConfig{
		Domain:        "inquiries-donations",
		MigrationPath: filepath.Join(basePath, "inquiries", "donations"),
		DatabaseURL:   databaseURL,
	}
	config.Inquiries.Business = MigrationConfig{
		Domain:        "inquiries-business",
		MigrationPath: filepath.Join(basePath, "inquiries", "business"),
		DatabaseURL:   databaseURL,
	}
	config.Inquiries.Media = MigrationConfig{
		Domain:        "inquiries-media",
		MigrationPath: filepath.Join(basePath, "inquiries", "media"),
		DatabaseURL:   databaseURL,
	}
	config.Inquiries.Volunteers = MigrationConfig{
		Domain:        "inquiries-volunteers",
		MigrationPath: filepath.Join(basePath, "inquiries", "volunteers"),
		DatabaseURL:   databaseURL,
	}
	
	// Supporting domain configurations
	config.Notifications = MigrationConfig{
		Domain:        "notifications",
		MigrationPath: filepath.Join(basePath, "notifications"),
		DatabaseURL:   databaseURL,
	}
	config.Gateway = MigrationConfig{
		Domain:        "gateway",
		MigrationPath: filepath.Join(basePath, "gateway"),
		DatabaseURL:   databaseURL,
	}
	config.Shared = MigrationConfig{
		Domain:        "shared",
		MigrationPath: filepath.Join(basePath, "shared"),
		DatabaseURL:   databaseURL,
	}
	
	return config
}

// GetMigrationURL returns the file:// URL for golang-migrate
func (c MigrationConfig) GetMigrationURL() string {
	return fmt.Sprintf("file://%s", c.MigrationPath)
}

// GetAllConfigs returns a slice of all migration configurations for iteration
func (d *DomainMigrationConfigs) GetAllConfigs() []MigrationConfig {
	return []MigrationConfig{
		d.Content.Services,
		d.Content.News,
		d.Content.Research,
		d.Content.Events,
		d.Inquiries.Donations,
		d.Inquiries.Business,
		d.Inquiries.Media,
		d.Inquiries.Volunteers,
		d.Notifications,
		d.Gateway,
		d.Shared,
	}
}