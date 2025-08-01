package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"health-monitor/internal/db"
	"health-monitor/internal/logger"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db      *sql.DB
	version string
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status     string           `json:"status"`
	Timestamp  time.Time        `json:"timestamp"`
	Database   bool             `json:"database"`
	Version    string           `json:"version"`
	Uptime     string           `json:"uptime,omitempty"`
	Migrations *MigrationHealth `json:"migrations,omitempty"`
}

// MigrationHealth represents migration status in health checks
type MigrationHealth struct {
	CurrentVersion int64 `json:"current_version"`
	IsUpToDate     bool  `json:"is_up_to_date"`
	Status         string `json:"status"`
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(db *sql.DB, version string) *HealthHandler {
	return &HealthHandler{
		db:      db,
		version: version,
	}
}

// HealthCheck handles GET /health requests
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	logger.Debug().Msg("Health check requested")

	// Check database connectivity
	dbHealthy := h.checkDatabase()
	
	// Check migration status
	migrationHealth := h.checkMigrationStatus()
	
	// Determine overall status
	status := "healthy"
	if !dbHealthy {
		status = "unhealthy"
	} else if migrationHealth != nil && !migrationHealth.IsUpToDate {
		status = "degraded" // Database works but migrations are pending
	}

	response := HealthResponse{
		Status:     status,
		Timestamp:  time.Now().UTC(),
		Database:   dbHealthy,
		Version:    h.version,
		Migrations: migrationHealth,
	}

	// Set content type first
	w.Header().Set("Content-Type", "application/json")

	// Set appropriate HTTP status code and log
	migrationStatus := "unknown"
	if migrationHealth != nil {
		migrationStatus = migrationHealth.Status
	}

	switch {
	case !dbHealthy:
		w.WriteHeader(http.StatusServiceUnavailable)
		logger.Warn().
			Bool("database", dbHealthy).
			Str("migration_status", migrationStatus).
			Msg("Health check failed")
	
	case migrationHealth != nil && !migrationHealth.IsUpToDate:
		w.WriteHeader(http.StatusOK) // Still OK but with degraded status
		logger.Warn().
			Bool("database", dbHealthy).
			Int64("migration_version", migrationHealth.CurrentVersion).
			Msg("Health check passed but migrations are pending")
	
	default:
		w.WriteHeader(http.StatusOK)
		logger.Debug().
			Bool("database", dbHealthy).
			Str("migration_status", migrationStatus).
			Msg("Health check passed")
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().Err(err).Msg("Failed to encode health check response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// checkDatabase verifies database connectivity
func (h *HealthHandler) checkDatabase() bool {
	if h.db == nil {
		return false
	}

	// Simple ping to check if database is accessible
	if err := h.db.Ping(); err != nil {
		logger.Error().Err(err).Msg("Database health check failed")
		return false
	}

	// Try a simple query to ensure database is functional
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count)
	if err != nil {
		logger.Error().Err(err).Msg("Database query health check failed")
		return false
	}

	return true
}

// checkMigrationStatus verifies migration status for health checks
func (h *HealthHandler) checkMigrationStatus() *MigrationHealth {
	if h.db == nil {
		return &MigrationHealth{
			CurrentVersion: -1,
			IsUpToDate:     false,
			Status:         "no_database",
		}
	}

	status, err := db.GetMigrationStatus(h.db)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get migration status for health check")
		return &MigrationHealth{
			CurrentVersion: -1,
			IsUpToDate:     false,
			Status:         "error",
		}
	}

	migrationStatus := "up_to_date"
	if status.HasPendingMigrations {
		migrationStatus = "pending"
	} else if !status.IsUpToDate {
		migrationStatus = "outdated"
	}

	return &MigrationHealth{
		CurrentVersion: status.CurrentVersion,
		IsUpToDate:     status.IsUpToDate && !status.HasPendingMigrations,
		Status:         migrationStatus,
	}
}

// RegisterRoutes registers health check routes
func (h *HealthHandler) RegisterRoutes(r http.Handler) {
	if mux, ok := r.(*http.ServeMux); ok {
		mux.HandleFunc("/health", h.HealthCheck)
	}
}