package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"health-monitor/internal/logger"
	_ "modernc.org/sqlite"
)

// Open initializes and returns a new database connection
func Open(dbPath string) (*sql.DB, error) {
	// Ensure the database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		logger.Error().Err(err).Str("dir", dbDir).Msg("Failed to create database directory")
		return nil, fmt.Errorf("failed to create database directory %s: %w", dbDir, err)
	}

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		logger.Info().Str("path", dbPath).Msg("Creating new database")
		file, err := os.Create(dbPath)
		if err != nil {
			logger.Error().Err(err).Str("path", dbPath).Msg("Failed to create database file")
			return nil, fmt.Errorf("failed to create database file %s: %w", dbPath, err)
		}
		file.Close()
	} else {
		logger.Info().Str("path", dbPath).Msg("Using existing database")
	}

	// Open database connection with proper SQLite settings for concurrency
	connectionString := dbPath + "?_busy_timeout=30000&_journal_mode=DELETE&_foreign_keys=on&_synchronous=NORMAL"
	db, err := sql.Open("sqlite", connectionString)
	if err != nil {
		logger.Error().Err(err).Str("path", dbPath).Msg("Failed to open database connection")
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	
	// Set reasonable connection pool settings for SQLite
	db.SetMaxOpenConns(1)  // SQLite works best with single writer
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // Connections never expire

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		logger.Error().Err(err).Str("path", dbPath).Msg("Database connection test failed")
		return nil, fmt.Errorf("database connection test failed: %w", err)
	}

	logger.Info().Str("path", dbPath).Msg("Database connection established")

	// Run Goose migrations
	if err := RunMigrations(db); err != nil {
		db.Close()
		logger.Error().Err(err).Msg("Database migration failed")
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	logger.Info().Msg("Database migrations completed successfully")
	return db, nil
}


