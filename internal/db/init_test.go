package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	
	_ "modernc.org/sqlite"
)

func TestInitializeDatabasePath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	tests := []struct {
		name         string
		isProduction bool
		wantError    bool
	}{
		{
			name:         "development environment",
			isProduction: false,
			wantError:    false,
		},
		{
			name:         "production environment",
			isProduction: true,
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing database file
			os.Remove(dbPath)
			
			err := initializeDatabasePath(dbPath, tt.isProduction)
			
			if (err != nil) != tt.wantError {
				t.Errorf("initializeDatabasePath() error = %v, wantError %v", err, tt.wantError)
				return
			}

			// Verify database file was created
			if _, err := os.Stat(dbPath); os.IsNotExist(err) {
				t.Errorf("Database file was not created at %s", dbPath)
			}

			// Verify directory exists
			dbDir := filepath.Dir(dbPath)
			if _, err := os.Stat(dbDir); os.IsNotExist(err) {
				t.Errorf("Database directory was not created at %s", dbDir)
			}
		})
	}
}

func TestBuildConnectionString(t *testing.T) {
	dbPath := "/test/path/db.sqlite"

	tests := []struct {
		name         string
		isProduction bool
		wantContains []string
	}{
		{
			name:         "development connection string",
			isProduction: false,
			wantContains: []string{"_foreign_keys=on", "_busy_timeout=5000", "_journal_mode=DELETE"},
		},
		{
			name:         "production connection string",
			isProduction: true,
			wantContains: []string{"_foreign_keys=on", "_busy_timeout=30000", "_journal_mode=WAL", "_synchronous=NORMAL"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildConnectionString(dbPath, tt.isProduction)
			
			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("buildConnectionString() = %v, should contain %v", result, want)
				}
			}
		})
	}
}

func TestOpenDatabaseWithRetry_InvalidPath(t *testing.T) {
	// Test with an invalid path that should fail
	invalidPath := "/invalid/path/that/does/not/exist/db.sqlite"
	
	_, err := openDatabaseWithRetry(invalidPath, false)
	if err == nil {
		t.Error("Expected error for invalid database path, got nil")
	}
}

func TestConfigureDatabasePool(t *testing.T) {
	// This test verifies that the function doesn't panic
	// We can't easily test the actual pool settings without a real database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	
	// Initialize the database path first
	if err := initializeDatabasePath(dbPath, false); err != nil {
		t.Fatalf("Failed to initialize database path: %v", err)
	}
	
	// Create a simple connection string for testing
	connectionString := buildConnectionString(dbPath, false)
	db, err := sql.Open("sqlite", connectionString)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test both production and development configurations
	configureDatabasePool(db, false)
	configureDatabasePool(db, true)
	
	// If we get here without panicking, the test passes
}

