package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
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

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.Info().Int("attempt", attempt).Bool("production", isProduction).Msg("Running database migrations")
		
		if err := RunMigrations(db); err != nil {
			logger.Error().Err(err).Int("attempt", attempt).Msg("Database migration failed")
			if attempt < maxRetries {
				logger.Info().Dur("delay", retryDelay).Int("attempt", attempt).Msg("Retrying database migrations")
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("database migration failed after %d attempts: %w", maxRetries, err)
		}

		logger.Info().Int("attempt", attempt).Msg("Database migrations completed successfully")
		return nil
	}

	return fmt.Errorf("database migration failed after %d attempts", maxRetries)
}


