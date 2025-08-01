package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"health-monitor/internal/logger"
	_ "modernc.org/sqlite"
)

// Open initializes and returns a new database connection with retry logic for production
func Open(dbPath string) (*sql.DB, error) {
	// Determine if we're in production environment
	isProduction := os.Getenv("RAILWAY_ENVIRONMENT") != ""
	
	// Initialize database directory and file with retry logic
	if err := initializeDatabasePath(dbPath, isProduction); err != nil {
		return nil, fmt.Errorf("failed to initialize database path: %w", err)
	}

	// Open database connection with retry logic
	db, err := openDatabaseWithRetry(dbPath, isProduction)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Run migrations with retry logic
	if err := runMigrationsWithRetry(db, isProduction); err != nil {
		db.Close()
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	logger.Info().Str("path", dbPath).Bool("production", isProduction).Msg("Database initialization completed successfully")
	return db, nil
}

// initializeDatabasePath ensures the database directory and file exist
func initializeDatabasePath(dbPath string, isProduction bool) error {
	// Ensure the database directory exists with appropriate permissions
	dbDir := filepath.Dir(dbPath)
	dirMode := os.FileMode(0755)
	if isProduction {
		// More restrictive permissions in production
		dirMode = 0750
	}

	if err := os.MkdirAll(dbDir, dirMode); err != nil {
		logger.Error().Err(err).Str("dir", dbDir).Bool("production", isProduction).Msg("Failed to create database directory")
		return fmt.Errorf("failed to create database directory %s: %w", dbDir, err)
	}

	// Verify directory is writable
	testFile := filepath.Join(dbDir, ".write_test")
	if file, err := os.Create(testFile); err != nil {
		logger.Error().Err(err).Str("dir", dbDir).Msg("Database directory is not writable")
		return fmt.Errorf("database directory %s is not writable: %w", dbDir, err)
	} else {
		file.Close()
		os.Remove(testFile)
	}

	// Check if database file exists, create if needed
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		logger.Info().Str("path", dbPath).Bool("production", isProduction).Msg("Creating new database file")
		file, err := os.Create(dbPath)
		if err != nil {
			logger.Error().Err(err).Str("path", dbPath).Msg("Failed to create database file")
			return fmt.Errorf("failed to create database file %s: %w", dbPath, err)
		}
		file.Close()
		
		// Set appropriate file permissions
		fileMode := os.FileMode(0644)
		if isProduction {
			fileMode = 0640
		}
		if err := os.Chmod(dbPath, fileMode); err != nil {
			logger.Warn().Err(err).Str("path", dbPath).Msg("Failed to set database file permissions")
		}
	} else {
		logger.Info().Str("path", dbPath).Bool("production", isProduction).Msg("Using existing database file")
	}

	return nil
}

// openDatabaseWithRetry opens database connection with retry logic for production reliability
func openDatabaseWithRetry(dbPath string, isProduction bool) (*sql.DB, error) {
	maxRetries := 1
	retryDelay := time.Second

	if isProduction {
		maxRetries = 5
		retryDelay = 2 * time.Second
	}

	var db *sql.DB
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Configure connection string with production-optimized settings
		connectionString := buildConnectionString(dbPath, isProduction)
		
		db, err = sql.Open("sqlite", connectionString)
		if err != nil {
			logger.Error().Err(err).Str("path", dbPath).Int("attempt", attempt).Msg("Failed to open database connection")
			if attempt < maxRetries {
				logger.Info().Dur("delay", retryDelay).Int("attempt", attempt).Msg("Retrying database connection")
				time.Sleep(retryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to open database connection after %d attempts: %w", maxRetries, err)
		}

		// Configure connection pool settings
		configureDatabasePool(db, isProduction)

		// Test the connection with retry
		if err = testDatabaseConnection(db, attempt, maxRetries); err != nil {
			db.Close()
			if attempt < maxRetries {
				logger.Info().Dur("delay", retryDelay).Int("attempt", attempt).Msg("Retrying database connection test")
				time.Sleep(retryDelay)
				continue
			}
			return nil, fmt.Errorf("database connection test failed after %d attempts: %w", maxRetries, err)
		}

		logger.Info().Str("path", dbPath).Int("attempt", attempt).Bool("production", isProduction).Msg("Database connection established")
		return db, nil
	}

	return nil, fmt.Errorf("failed to establish database connection after %d attempts", maxRetries)
}

// buildConnectionString creates optimized connection string based on environment
func buildConnectionString(dbPath string, isProduction bool) string {
	baseString := dbPath + "?_foreign_keys=on"
	
	if isProduction {
		// Production-optimized settings for reliability and performance
		return baseString + "&_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=memory"
	} else {
		// Development settings for faster iteration
		return baseString + "&_busy_timeout=5000&_journal_mode=DELETE&_synchronous=NORMAL"
	}
}

// configureDatabasePool sets connection pool settings based on environment
func configureDatabasePool(db *sql.DB, isProduction bool) {
	if isProduction {
		// Production settings for better reliability
		db.SetMaxOpenConns(1)    // SQLite works best with single writer
		db.SetMaxIdleConns(1)    // Keep connection alive
		db.SetConnMaxLifetime(0) // Connections never expire
		db.SetConnMaxIdleTime(5 * time.Minute) // Close idle connections after 5 minutes
	} else {
		// Development settings
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
		db.SetConnMaxLifetime(0)
	}
}

// testDatabaseConnection tests database connectivity with proper error handling
func testDatabaseConnection(db *sql.DB, attempt, maxRetries int) error {
	if err := db.Ping(); err != nil {
		logger.Error().Err(err).Int("attempt", attempt).Msg("Database connection test failed")
		return err
	}

	// Additional connectivity test - try a simple query
	var result int
	if err := db.QueryRow("SELECT 1").Scan(&result); err != nil {
		logger.Error().Err(err).Int("attempt", attempt).Msg("Database query test failed")
		return err
	}

	return nil
}

// runMigrationsWithRetry runs database migrations with retry logic for production
func runMigrationsWithRetry(db *sql.DB, isProduction bool) error {
	maxRetries := 1
	retryDelay := time.Second

	if isProduction {
		maxRetries = 3
		retryDelay = 5 * time.Second
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.Info().Int("attempt", attempt).Bool("production", isProduction).Msg("Running database migrations")
		
		// Get migration status before attempting migration
		status, statusErr := GetMigrationStatus(db)
		if statusErr != nil {
			logger.Warn().Err(statusErr).Msg("Failed to get migration status, proceeding with migration")
		} else {
			logger.Info().
				Int64("current_version", status.CurrentVersion).
				Bool("up_to_date", status.IsUpToDate).
				Msg("Pre-migration database status")
		}
		
		if err := RunMigrations(db); err != nil {
			lastErr = err
			logger.Error().
				Err(err).
				Int("attempt", attempt).
				Int("max_retries", maxRetries).
				Msg("Database migration failed")
			
			// Don't retry if this is a critical migration error that won't be fixed by retrying
			if isCriticalMigrationError(err) {
				logger.Error().Err(err).Msg("Critical migration error detected, not retrying")
				return fmt.Errorf("critical database migration error on attempt %d: %w", attempt, err)
			}
			
			if attempt < maxRetries {
				logger.Info().
					Dur("delay", retryDelay).
					Int("attempt", attempt).
					Int("remaining", maxRetries-attempt).
					Msg("Retrying database migrations")
				time.Sleep(retryDelay)
				continue
			}
			
			// Final attempt failed
			logger.Error().
				Err(err).
				Int("total_attempts", maxRetries).
				Msg("All migration attempts failed")
			return fmt.Errorf("database migration failed after %d attempts, last error: %w", maxRetries, err)
		}

		// Migration succeeded, log final status
		finalStatus, statusErr := GetMigrationStatus(db)
		if statusErr != nil {
			logger.Warn().Err(statusErr).Msg("Failed to get final migration status")
		} else {
			logger.Info().
				Int64("final_version", finalStatus.CurrentVersion).
				Bool("up_to_date", finalStatus.IsUpToDate).
				Int("attempt", attempt).
				Msg("Database migrations completed successfully")
		}
		
		return nil
	}

	return fmt.Errorf("database migration failed after %d attempts, last error: %w", maxRetries, lastErr)
}

// isCriticalMigrationError determines if a migration error is critical and shouldn't be retried
func isCriticalMigrationError(err error) bool {
	if err == nil {
		return false
	}
	
	errorStr := err.Error()
	
	// Critical errors that indicate structural problems that won't be fixed by retrying
	criticalPatterns := []string{
		"syntax error",
		"no such table",
		"duplicate column name",
		"constraint failed",
		"database is locked", // This might be retryable in some cases, but often indicates a deeper issue
		"rollback failed",
		"backup failed",
	}
	
	for _, pattern := range criticalPatterns {
		if strings.Contains(errorStr, pattern) {
			return true
		}
	}
	
	return false
}




