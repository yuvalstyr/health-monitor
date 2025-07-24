package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test default port
	os.Unsetenv("PORT")
	config := Load()
	if config.Port != "3000" {
		t.Errorf("Expected default port 3000, got %s", config.Port)
	}

	// Test custom port
	os.Setenv("PORT", "8080")
	defer os.Unsetenv("PORT")
	config = Load()
	if config.Port != "8080" {
		t.Errorf("Expected port 8080, got %s", config.Port)
	}
}