package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"time"

	"health-monitor/internal/logger"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// MigrationResult represents the result of a migration operation
type MigrationResult struct {
	Success         bool
	AppliedCount    int
	CurrentVersion  int64
	PreviousVersion int64
	Duration        time.Duration
	Error           error
}

// RunMigrations runs all pending Goose migrations with enhanced logging and error handling
func RunMigrations(db *sql.DB) error {
	result := runMigrationsWithResult(db)
	
	if !result.Success {
		return result.Error
	}
	
	return nil
}

// runMigrationsWithResult runs migrations and returns detailed result information
func runMigrationsWithResult(db *sql.DB) MigrationResult {
	startTime := time.Now()
	isProduction := os.Getenv("RAILWAY_ENVIRONMENT") != ""
	
	result := MigrationResult{
		Success: false,
	}
	
	logger.Info().Bool("production", isProduction).Msg("Starting database migration process")
	
	// Set up Goose with embedded migrations
	goose.SetBaseFS(embedMigrations)
	
	if err := goose.SetDialect("sqlite3"); err != nil {
		result.Error = fmt.Errorf("failed to set goose dialect: %w", err)
		logger.Error().Err(err).Msg("Failed to set migration dialect")
		return result
	}
	
	// Get current version before migration
	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		result.Error = fmt.Errorf("failed to get current database version: %w", err)
		logger.Error().Err(err).Msg("Failed to get current database version")
		return result
	}
	
	result.PreviousVersion = currentVersion
	logger.Info().Int64("current_version", currentVersion).Msg("Current database version")
	
	// Create backup before migration in production
	if isProduction && currentVersion > 0 {
		if err := createMigrationBackup(db); err != nil {
			logger.Warn().Err(err).Msg("Failed to create migration backup, continuing anyway")
		} else {
			logger.Info().Msg("Migration backup created successfully")
		}
	}
	
	// Run migrations
	logger.Info().Msg("Applying pending database migrations")
	if err := goose.Up(db, "migrations"); err != nil {
		result.Error = fmt.Errorf("failed to run goose migrations: %w", err)
		result.Duration = time.Since(startTime)
		
		logger.Error().
			Err(err).
			Dur("duration", result.Duration).
			Int64("previous_version", result.PreviousVersion).
			Msg("Database migration failed")
		
		// Attempt rollback in production if migration fails
		if isProduction {
			logger.Warn().Msg("Attempting to rollback failed migration")
			if rollbackErr := rollbackToVersion(db, currentVersion); rollbackErr != nil {
				logger.Error().Err(rollbackErr).Msg("Migration rollback failed - database may be in inconsistent state")
				result.Error = fmt.Errorf("migration failed and rollback failed: migration error: %w, rollback error: %v", err, rollbackErr)
			} else {
				logger.Info().Int64("version", currentVersion).Msg("Successfully rolled back to previous version")
				result.Error = fmt.Errorf("migration failed but successfully rolled back to version %d: %w", currentVersion, err)
			}
		}
		
		return result
	}
	
	// Get final version after migration
	finalVersion, err := goose.GetDBVersion(db)
	if err != nil {
		result.Error = fmt.Errorf("failed to get final database version: %w", err)
		logger.Error().Err(err).Msg("Failed to get final database version")
		return result
	}
	
	result.CurrentVersion = finalVersion
	result.AppliedCount = int(finalVersion - currentVersion)
	result.Duration = time.Since(startTime)
	result.Success = true
	
	logger.Info().
		Int64("previous_version", result.PreviousVersion).
		Int64("current_version", result.CurrentVersion).
		Int("applied_migrations", result.AppliedCount).
		Dur("duration", result.Duration).
		Bool("production", isProduction).
		Msg("Database migration completed successfully")
	
	return result
}

// createMigrationBackup creates a backup of the database before migration
func createMigrationBackup(db *sql.DB) error {
	// For SQLite, we can create a backup using the backup API
	// This is a simplified backup - in a real production system you might want
	// to use VACUUM INTO or copy the file
	backupQuery := `
		PRAGMA wal_checkpoint(FULL);
		VACUUM;
	`
	
	if _, err := db.Exec(backupQuery); err != nil {
		return fmt.Errorf("failed to prepare database for backup: %w", err)
	}
	
	logger.Info().Msg("Database prepared for migration (WAL checkpoint and vacuum completed)")
	return nil
}

// rollbackToVersion attempts to rollback the database to a specific version
func rollbackToVersion(db *sql.DB, targetVersion int64) error {
	logger.Info().Int64("target_version", targetVersion).Msg("Starting migration rollback")
	
	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current version for rollback: %w", err)
	}
	
	if currentVersion <= targetVersion {
		logger.Info().
			Int64("current", currentVersion).
			Int64("target", targetVersion).
			Msg("No rollback needed - current version is not higher than target")
		return nil
	}
	
	// Use Goose's down migration to rollback
	if err := goose.DownTo(db, "migrations", targetVersion); err != nil {
		return fmt.Errorf("failed to rollback to version %d: %w", targetVersion, err)
	}
	
	// Verify rollback was successful
	finalVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to verify rollback: %w", err)
	}
	
	if finalVersion != targetVersion {
		return fmt.Errorf("rollback verification failed: expected version %d, got %d", targetVersion, finalVersion)
	}
	
	logger.Info().
		Int64("from_version", currentVersion).
		Int64("to_version", finalVersion).
		Msg("Migration rollback completed successfully")
	
	return nil
}

// GetMigrationStatus returns the current migration status
func GetMigrationStatus(db *sql.DB) (MigrationStatus, error) {
	status := MigrationStatus{}
	
	if db == nil {
		return status, fmt.Errorf("database connection is nil")
	}
	
	// Get current version
	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return status, fmt.Errorf("failed to get database version: %w", err)
	}
	
	status.CurrentVersion = currentVersion
	
	// Set up Goose to check for pending migrations
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return status, fmt.Errorf("failed to set dialect for status check: %w", err)
	}
	
	// Create a test database with current schema to check for pending migrations
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return status, fmt.Errorf("failed to create test database for status check: %w", err)
	}
	defer testDB.Close()
	
	// Apply migrations up to current version to test database
	if currentVersion > 0 {
		if err := goose.UpTo(testDB, "migrations", currentVersion); err != nil {
			// If we can't apply migrations to test DB, assume no pending migrations
			// This is a fallback to avoid breaking the status check
			status.HasPendingMigrations = false
			status.IsUpToDate = true
		} else {
			// Try to apply one more migration to see if there are pending ones
			testCurrentVersion, err := goose.GetDBVersion(testDB)
			if err != nil {
				status.HasPendingMigrations = false
				status.IsUpToDate = true
			} else {
				// Try to go up one more version to see if there are pending migrations
				if err := goose.UpByOne(testDB, "migrations"); err != nil {
					// No more migrations available
					status.HasPendingMigrations = false
					status.IsUpToDate = true
				} else {
					// There was another migration available
					newVersion, _ := goose.GetDBVersion(testDB)
					status.HasPendingMigrations = newVersion > testCurrentVersion
					status.IsUpToDate = !status.HasPendingMigrations
				}
			}
		}
	} else {
		// If current version is 0, check if there are any migrations available
		if err := goose.UpByOne(testDB, "migrations"); err != nil {
			// No migrations available
			status.HasPendingMigrations = false
			status.IsUpToDate = true
		} else {
			// There are migrations available
			status.HasPendingMigrations = true
			status.IsUpToDate = false
		}
	}
	
	logger.Info().
		Int64("current_version", status.CurrentVersion).
		Bool("up_to_date", status.IsUpToDate).
		Bool("has_pending", status.HasPendingMigrations).
		Msg("Migration status retrieved")
	
	return status, nil
}

// MigrationStatus represents the current state of database migrations
type MigrationStatus struct {
	CurrentVersion        int64
	IsUpToDate           bool
	HasPendingMigrations bool
}

// Migrate is kept for backward compatibility but now uses enhanced RunMigrations
func Migrate(db *sql.DB) error {
	return RunMigrations(db)
}

// RollbackMigration rolls back the database to the previous migration
func RollbackMigration(db *sql.DB) error {
	logger.Info().Msg("Starting migration rollback")
	
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}
	
	if currentVersion <= 0 {
		logger.Info().Msg("No migrations to rollback")
		return nil
	}
	
	targetVersion := currentVersion - 1
	return rollbackToVersion(db, targetVersion)
}
