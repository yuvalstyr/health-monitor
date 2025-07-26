package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test development default port
	os.Unsetenv("PORT")
	os.Unsetenv("RAILWAY_ENVIRONMENT")
	config := Load()
	if config.Port != "3000" {
		t.Errorf("Expected development default port 3000, got %s", config.Port)
	}

	// Test production default port
	os.Unsetenv("PORT")
	os.Setenv("RAILWAY_ENVIRONMENT", "production")
	defer os.Unsetenv("RAILWAY_ENVIRONMENT")
	config = Load()
	if config.Port != "8080" {
		t.Errorf("Expected production default port 8080, got %s", config.Port)
	}

	// Test custom port override
	os.Setenv("PORT", "9000")
	defer os.Unsetenv("PORT")
	config = Load()
	if config.Port != "9000" {
		t.Errorf("Expected custom port 9000, got %s", config.Port)
	}

	// Test default version
	os.Unsetenv("APP_VERSION")
	config = Load()
	if config.Version != "1.0.0" {
		t.Errorf("Expected default version 1.0.0, got %s", config.Version)
	}

	// Test custom version
	os.Setenv("APP_VERSION", "2.1.0")
	defer os.Unsetenv("APP_VERSION")
	config = Load()
	if config.Version != "2.1.0" {
		t.Errorf("Expected custom version 2.1.0, got %s", config.Version)
	}
}