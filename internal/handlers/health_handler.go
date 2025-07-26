package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"health-monitor/internal/logger"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db      *sql.DB
	version string
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Database  bool      `json:"database"`
	Version   string    `json:"version"`
	Uptime    string    `json:"uptime,omitempty"`
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
	
	// Determine overall status
	status := "healthy"
	if !dbHealthy {
		status = "unhealthy"
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC(),
		Database:  dbHealthy,
		Version:   h.version,
	}

	// Set appropriate HTTP status code
	if !dbHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
		logger.Warn().Bool("database", dbHealthy).Msg("Health check failed")
	} else {
		w.WriteHeader(http.StatusOK)
		logger.Debug().Bool("database", dbHealthy).Msg("Health check passed")
	}

	// Set content type and encode response
	w.Header().Set("Content-Type", "application/json")
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

// RegisterRoutes registers health check routes
func (h *HealthHandler) RegisterRoutes(r http.Handler) {
	if mux, ok := r.(*http.ServeMux); ok {
		mux.HandleFunc("/health", h.HealthCheck)
	}
}