package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"health-monitor/internal/db"
	"health-monitor/internal/logger"
)

// MigrationHandler handles migration management requests
type MigrationHandler struct {
	db *sql.DB
}

// MigrationStatusResponse represents the migration status response
type MigrationStatusResponse struct {
	CurrentVersion        int64     `json:"current_version"`
	IsUpToDate           bool      `json:"is_up_to_date"`
	HasPendingMigrations bool      `json:"has_pending_migrations"`
	Timestamp            time.Time `json:"timestamp"`
	Environment          string    `json:"environment"`
}

// MigrationRollbackRequest represents a rollback request
type MigrationRollbackRequest struct {
	TargetVersion *int64 `json:"target_version,omitempty"` // If nil, rollback one version
	Confirm       bool   `json:"confirm"`                  // Safety confirmation
}

// MigrationRollbackResponse represents a rollback response
type MigrationRollbackResponse struct {
	Success         bool      `json:"success"`
	Message         string    `json:"message"`
	PreviousVersion int64     `json:"previous_version"`
	NewVersion      int64     `json:"new_version"`
	Timestamp       time.Time `json:"timestamp"`
	Error           string    `json:"error,omitempty"`
}

// NewMigrationHandler creates a new migration management handler
func NewMigrationHandler(database *sql.DB) *MigrationHandler {
	return &MigrationHandler{
		db: database,
	}
}

// GetMigrationStatus handles GET /admin/migrations/status requests
func (h *MigrationHandler) GetMigrationStatus(w http.ResponseWriter, r *http.Request) {
	logger.Info().Msg("Migration status requested")

	if h.db == nil {
		logger.Error().Msg("Database connection is nil")
		http.Error(w, "Database not available", http.StatusServiceUnavailable)
		return
	}

	status, err := db.GetMigrationStatus(h.db)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get migration status")
		http.Error(w, "Failed to get migration status", http.StatusInternalServerError)
		return
	}

	response := MigrationStatusResponse{
		CurrentVersion:        status.CurrentVersion,
		IsUpToDate:           status.IsUpToDate,
		HasPendingMigrations: status.HasPendingMigrations,
		Timestamp:            time.Now().UTC(),
		Environment:          getEnvironment(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().Err(err).Msg("Failed to encode migration status response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info().
		Int64("current_version", status.CurrentVersion).
		Bool("up_to_date", status.IsUpToDate).
		Bool("has_pending", status.HasPendingMigrations).
		Msg("Migration status retrieved successfully")
}

// PostMigrationRollback handles POST /admin/migrations/rollback requests
func (h *MigrationHandler) PostMigrationRollback(w http.ResponseWriter, r *http.Request) {
	logger.Info().Msg("Migration rollback requested")

	// Only allow rollback in development or with explicit production override
	if !isRollbackAllowed(r) {
		logger.Warn().Msg("Migration rollback not allowed in this environment")
		http.Error(w, "Migration rollback not allowed in production environment", http.StatusForbidden)
		return
	}

	if h.db == nil {
		logger.Error().Msg("Database connection is nil")
		http.Error(w, "Database not available", http.StatusServiceUnavailable)
		return
	}

	// Parse request body
	var req MigrationRollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Msg("Failed to parse rollback request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Safety check - require explicit confirmation
	if !req.Confirm {
		logger.Warn().Msg("Migration rollback attempted without confirmation")
		http.Error(w, "Rollback requires explicit confirmation", http.StatusBadRequest)
		return
	}

	// Get current version before rollback
	currentVersion, err := db.GetMigrationStatus(h.db)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get current migration version")
		response := MigrationRollbackResponse{
			Success:   false,
			Message:   "Failed to get current migration version",
			Timestamp: time.Now().UTC(),
			Error:     err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	previousVersion := currentVersion.CurrentVersion

	// Determine target version
	var targetVersion int64
	if req.TargetVersion != nil {
		targetVersion = *req.TargetVersion
		if targetVersion >= previousVersion {
			logger.Warn().
				Int64("target", targetVersion).
				Int64("current", previousVersion).
				Msg("Invalid rollback target version")
			response := MigrationRollbackResponse{
				Success:         false,
				Message:         "Target version must be lower than current version",
				PreviousVersion: previousVersion,
				NewVersion:      previousVersion,
				Timestamp:       time.Now().UTC(),
				Error:           "Invalid target version",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}
	} else {
		// Default: rollback one version
		targetVersion = previousVersion - 1
		if targetVersion < 0 {
			targetVersion = 0
		}
	}

	logger.Info().
		Int64("from_version", previousVersion).
		Int64("to_version", targetVersion).
		Msg("Starting migration rollback")

	// Perform rollback
	var rollbackErr error
	if targetVersion == previousVersion-1 {
		// Single version rollback
		rollbackErr = db.RollbackMigration(h.db)
	} else {
		// Multi-version rollback (not implemented in basic version)
		rollbackErr = db.RollbackMigration(h.db) // This will only rollback one version
		logger.Warn().
			Int64("requested", targetVersion).
			Int64("actual", previousVersion-1).
			Msg("Multi-version rollback not fully implemented, rolled back one version")
	}

	// Get final version after rollback
	finalStatus, statusErr := db.GetMigrationStatus(h.db)
	finalVersion := previousVersion // Default if we can't get status
	if statusErr == nil {
		finalVersion = finalStatus.CurrentVersion
	}

	// Prepare response
	response := MigrationRollbackResponse{
		Success:         rollbackErr == nil,
		PreviousVersion: previousVersion,
		NewVersion:      finalVersion,
		Timestamp:       time.Now().UTC(),
	}

	if rollbackErr != nil {
		response.Message = "Migration rollback failed"
		response.Error = rollbackErr.Error()
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error().
			Err(rollbackErr).
			Int64("from_version", previousVersion).
			Int64("to_version", targetVersion).
			Msg("Migration rollback failed")
	} else {
		response.Message = "Migration rollback completed successfully"
		w.WriteHeader(http.StatusOK)
		logger.Info().
			Int64("from_version", previousVersion).
			Int64("to_version", finalVersion).
			Msg("Migration rollback completed successfully")
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().Err(err).Msg("Failed to encode rollback response")
		return
	}
}

// isRollbackAllowed checks if migration rollback is allowed in the current environment
func isRollbackAllowed(r *http.Request) bool {
	// Allow in development
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		return true
	}

	// Allow in production only with explicit override header
	if r.Header.Get("X-Allow-Production-Rollback") == "true" {
		logger.Warn().Msg("Production migration rollback explicitly allowed via header")
		return true
	}

	// Allow if query parameter is set (for emergency use)
	if r.URL.Query().Get("allow_production") == "true" {
		logger.Warn().Msg("Production migration rollback explicitly allowed via query parameter")
		return true
	}

	return false
}

// getEnvironment returns the current environment string
func getEnvironment() string {
	if os.Getenv("RAILWAY_ENVIRONMENT") != "" {
		return "production"
	}
	return "development"
}

// RegisterRoutes registers migration management routes
func (h *MigrationHandler) RegisterRoutes(r chi.Router) {
	r.Route("/admin/migrations", func(r chi.Router) {
		r.Get("/status", h.GetMigrationStatus)
		r.Post("/rollback", h.PostMigrationRollback)
	})
}