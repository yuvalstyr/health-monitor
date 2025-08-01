package config

import (
	"fmt"
	"os"
	"path/filepath"

	"health-monitor/internal/logger"
)

// Config holds application configuration
type Config struct {
	Port     string
	DBPath   string
	LogLevel string
	Version  string
	IsProduction bool
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	// Detect if running in production (Railway sets RAILWAY_ENVIRONMENT)
	isProduction := os.Getenv("RAILWAY_ENVIRONMENT") != ""
	
	// Port configuration
	port := os.Getenv("PORT")
	if port == "" {
		if isProduction {
			port = "8080" // Production default
		} else {
			port = "3000" // Development default
		}
		logger.Debug().Str("port", port).Msg("Using default port")
	}

	// Database path configuration
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		if isProduction {
			// Production: use Railway volume path
			dbPath = "/data/health-monitor.db"
		} else {
			// Development: use local path
			dbPath = "health-monitor.db"
		}
	}

	// Log level configuration
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		if isProduction {
			logLevel = "info"
		} else {
			logLevel = "debug"
		}
	}

	// Version configuration
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "1.0.0" // Default version
	}

	// Ensure database directory exists and is properly configured for production
	if isProduction {
		if err := ensureProductionDatabaseDirectory(dbPath); err != nil {
			logger.Error().Err(err).Str("db_path", dbPath).Msg("Failed to ensure production database directory")
		}
	}

	config := &Config{
		Port:         port,
		DBPath:       dbPath,
		LogLevel:     logLevel,
		Version:      version,
		IsProduction: isProduction,
	}

	logger.Info().
		Str("port", config.Port).
		Str("db_path", config.DBPath).
		Str("log_level", config.LogLevel).
		Str("version", config.Version).
		Bool("is_production", config.IsProduction).
		Msg("Configuration loaded")

	return config
}

// ensureProductionDatabaseDirectory ensures the database directory exists and is properly configured for production
func ensureProductionDatabaseDirectory(dbPath string) error {
	dbDir := filepath.Dir(dbPath)
	
	// Create directory with restrictive permissions for production
	if err := os.MkdirAll(dbDir, 0750); err != nil {
		return fmt.Errorf("failed to create database directory %s: %w", dbDir, err)
	}

	// Verify directory exists and is accessible
	if stat, err := os.Stat(dbDir); err != nil {
		return fmt.Errorf("failed to verify database directory %s: %w", dbDir, err)
	} else if !stat.IsDir() {
		return fmt.Errorf("database path %s is not a directory", dbDir)
	}

	// Test write permissions by creating a temporary file
	testFile := filepath.Join(dbDir, ".write_test_config")
	if file, err := os.Create(testFile); err != nil {
		return fmt.Errorf("database directory %s is not writable: %w", dbDir, err)
	} else {
		file.Close()
		os.Remove(testFile)
	}

	logger.Info().Str("dir", dbDir).Msg("Production database directory verified and ready")
	return nil
}