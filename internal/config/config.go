package config

import (
	"os"

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

	// Note: Database directory initialization is handled by the db package during Open()

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

