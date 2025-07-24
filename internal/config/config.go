package config

import (
	"os"

	"health-monitor/internal/logger"
)

// Config holds application configuration
type Config struct {
	Port string
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
		logger.Debug().Str("port", port).Msg("Using default port")
	}

	return &Config{
		Port: port,
	}
}