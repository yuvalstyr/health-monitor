package db

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupMigrationTestDB(t *testing.T) *sql.DB {
	// Create in-memory database for testing
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err, "Should be able to open in-memory database")
	
	t.Cleanup(func() {
		db.Close()
	})
	
	return db
}

func TestRunMigrations(t *testing.T) {
	// Create test database
	db := setupMigrationTestDB(t)

	// Test running migrations
	err := RunMigrations(db)
	require.NoError(t, err, "RunMigrations should succeed")

	// Verify migrations were applied by checking for expected tables
	tables := []string{"gauge_templates", "gauge_instances", "gauge_values"}
	for _, table := range tables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		require.NoError(t, err, "Should be able to query sqlite_master")
		assert.Equal(t, 1, count, "Table %s should exist", table)
	}
}

func TestGetMigrationStatus(t *testing.T) {
	// Create test database
	db := setupMigrationTestDB(t)

	// Run migrations first
	err := RunMigrations(db)
	require.NoError(t, err, "RunMigrations should succeed")

	// Test getting migration status
	status, err := GetMigrationStatus(db)
	require.NoError(t, err, "GetMigrationStatus should succeed")

	assert.True(t, status.CurrentVersion > 0, "Current version should be greater than 0")
	assert.True(t, status.IsUpToDate, "Database should be up to date after migrations")
}

func TestMigrationResult(t *testing.T) {
	// Create test database
	db := setupMigrationTestDB(t)

	// Test migration with result
	result := runMigrationsWithResult(db)
	
	assert.True(t, result.Success, "Migration should succeed")
	assert.True(t, result.AppliedCount >= 0, "Applied count should be non-negative")
	assert.True(t, result.CurrentVersion > 0, "Current version should be greater than 0")
	assert.True(t, result.Duration > 0, "Duration should be positive")
	assert.NoError(t, result.Error, "Error should be nil on success")
}

func TestMigrationWithProductionEnvironment(t *testing.T) {
	// Set production environment
	originalEnv := os.Getenv("RAILWAY_ENVIRONMENT")
	os.Setenv("RAILWAY_ENVIRONMENT", "production")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("RAILWAY_ENVIRONMENT")
		} else {
			os.Setenv("RAILWAY_ENVIRONMENT", originalEnv)
		}
	}()

	// Create test database
	db := setupMigrationTestDB(t)

	// Test migration in production environment
	result := runMigrationsWithResult(db)
	
	assert.True(t, result.Success, "Migration should succeed in production")
	assert.True(t, result.AppliedCount >= 0, "Applied count should be non-negative")
	assert.NoError(t, result.Error, "Error should be nil on success")
}

func TestRollbackMigration(t *testing.T) {
	// Create test database
	db := setupMigrationTestDB(t)

	// Run migrations first
	err := RunMigrations(db)
	require.NoError(t, err, "RunMigrations should succeed")

	// Get current version
	initialStatus, err := GetMigrationStatus(db)
	require.NoError(t, err, "GetMigrationStatus should succeed")

	// Only test rollback if we have migrations to rollback
	if initialStatus.CurrentVersion > 0 {
		// Test rollback
		err = RollbackMigration(db)
		require.NoError(t, err, "RollbackMigration should succeed")

		// Verify rollback worked
		finalStatus, err := GetMigrationStatus(db)
		require.NoError(t, err, "GetMigrationStatus should succeed after rollback")

		assert.Equal(t, initialStatus.CurrentVersion-1, finalStatus.CurrentVersion, 
			"Version should be decremented by 1 after rollback")
	} else {
		t.Skip("No migrations to rollback")
	}
}

func TestMigrationErrorHandling(t *testing.T) {
	// Test status with nil database
	_, err := GetMigrationStatus(nil)
	assert.Error(t, err, "GetMigrationStatus should fail with nil database")
	
	// Test rollback with nil database
	err = RollbackMigration(nil)
	assert.Error(t, err, "RollbackMigration should fail with nil database")
}

func TestCriticalMigrationErrorDetection(t *testing.T) {
	testCases := []struct {
		name        string
		errorMsg    string
		isCritical  bool
	}{
		{"syntax error", "syntax error in SQL", true},
		{"no such table", "no such table: test", true},
		{"duplicate column", "duplicate column name", true},
		{"constraint failed", "constraint failed", true},
		{"database locked", "database is locked", true},
		{"rollback failed", "rollback failed", true},
		{"connection error", "connection refused", false},
		{"timeout error", "timeout occurred", false},
		{"nil error", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			if tc.errorMsg != "" {
				// Create a real error with the expected message
				err = fmt.Errorf("migration failed: %s", tc.errorMsg)
			}
			
			// Test the actual error message matching logic
			result := isCriticalMigrationError(err)
			
			assert.Equal(t, tc.isCritical, result, 
				"Expected isCritical=%v for error message: %s", tc.isCritical, tc.errorMsg)
		})
	}
}

func TestMigrationBackup(t *testing.T) {
	// Create test database
	db := setupMigrationTestDB(t)

	// Run migrations to have some data
	err := RunMigrations(db)
	require.NoError(t, err, "RunMigrations should succeed")

	// Test backup creation
	err = createMigrationBackup(db)
	assert.NoError(t, err, "createMigrationBackup should succeed")
}

func TestRollbackToVersion(t *testing.T) {
	// Create test database
	db := setupMigrationTestDB(t)

	// Run migrations first
	err := RunMigrations(db)
	require.NoError(t, err, "RunMigrations should succeed")

	// Get current version
	currentStatus, err := GetMigrationStatus(db)
	require.NoError(t, err, "GetMigrationStatus should succeed")

	if currentStatus.CurrentVersion > 0 {
		// Test rollback to previous version
		targetVersion := currentStatus.CurrentVersion - 1
		err = rollbackToVersion(db, targetVersion)
		require.NoError(t, err, "rollbackToVersion should succeed")

		// Verify rollback
		finalStatus, err := GetMigrationStatus(db)
		require.NoError(t, err, "GetMigrationStatus should succeed after rollback")
		assert.Equal(t, targetVersion, finalStatus.CurrentVersion, 
			"Should rollback to target version")

		// Test rollback to same version (should be no-op)
		err = rollbackToVersion(db, targetVersion)
		assert.NoError(t, err, "rollbackToVersion to same version should succeed")
	} else {
		t.Skip("No migrations to rollback")
	}
}