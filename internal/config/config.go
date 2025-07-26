package config

import (
	"os"
	"path/filepath"

	"health-monitor/internal/logger"
)

// Config holds application configuration
type Config struct {
	Port     string
	DBPath   string
	LogLevel string
	IsProduction bool
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	// Detect if running in production (Railway sets RAILWAY_ENVIRONMENT)
	isProduction := os.Getenv("RAILWAY_ENVIRONMENT") != ""
	
	// Port configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
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

	// Ensure database directory exists in production
	if isProduction {
		dbDir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			logger.Error().Err(err).Str("dir", dbDir).Msg("Failed to create database directory")
		} else {
			logger.Info().Str("dir", dbDir).Msg("Database directory ensured")
		}
	}

	config := &Config{
		Port:         port,
		DBPath:       dbPath,
		LogLevel:     logLevel,
		IsProduction: isProduction,
	}

	logger.Info().
		Str("port", config.Port).
		Str("db_path", config.DBPath).
		Str("log_level", config.LogLevel).
		Bool("is_production", config.IsProduction).
		Msg("Configuration loaded")

	return config
}