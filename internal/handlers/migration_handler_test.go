package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"health-monitor/internal/db"
	_ "modernc.org/sqlite"
)

func setupMigrationHandlerTestDB(t *testing.T) *sql.DB {
	// Create in-memory database for testing
	database, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err, "Should be able to open in-memory database")
	
	// Run migrations to have a proper database state
	err = db.RunMigrations(database)
	require.NoError(t, err, "Should be able to run migrations")
	
	t.Cleanup(func() {
		database.Close()
	})
	
	return database
}

func TestMigrationHandler_GetMigrationStatus(t *testing.T) {
	// Create test database with migrations
	database := setupMigrationHandlerTestDB(t)
	
	// Create migration handler
	handler := NewMigrationHandler(database)
	
	// Create test request
	req := httptest.NewRequest("GET", "/admin/migrations/status", nil)
	w := httptest.NewRecorder()
	
	// Call handler
	handler.GetMigrationStatus(w, req)
	
	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	
	// Parse response
	var response MigrationStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Should be able to parse response")
	
	// Verify response structure
	assert.True(t, response.CurrentVersion > 0, "Should have migrations applied")
	assert.True(t, response.IsUpToDate, "Should be up to date after migrations")
	assert.NotEmpty(t, response.Environment)
	assert.False(t, response.Timestamp.IsZero())
}

func TestMigrationHandler_GetMigrationStatusWithNilDB(t *testing.T) {
	// Create migration handler with nil database
	handler := NewMigrationHandler(nil)
	
	// Create test request
	req := httptest.NewRequest("GET", "/admin/migrations/status", nil)
	w := httptest.NewRecorder()
	
	// Call handler
	handler.GetMigrationStatus(w, req)
	
	// Check response
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestMigrationHandler_PostMigrationRollback_Development(t *testing.T) {
	// Ensure we're in development environment
	originalEnv := os.Getenv("RAILWAY_ENVIRONMENT")
	os.Unsetenv("RAILWAY_ENVIRONMENT")
	defer func() {
		if originalEnv != "" {
			os.Setenv("RAILWAY_ENVIRONMENT", originalEnv)
		}
	}()
	
	// Create test database with migrations
	database := setupMigrationHandlerTestDB(t)
	
	// Create migration handler
	handler := NewMigrationHandler(database)
	
	// Create rollback request
	rollbackReq := MigrationRollbackRequest{
		Confirm: true,
	}
	reqBody, err := json.Marshal(rollbackReq)
	require.NoError(t, err)
	
	// Create test request
	req := httptest.NewRequest("POST", "/admin/migrations/rollback", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Call handler
	handler.PostMigrationRollback(w, req)
	
	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	
	// Parse response
	var response MigrationRollbackResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Should be able to parse response")
	
	// Verify response structure
	assert.True(t, response.Success, "Rollback should succeed")
	assert.True(t, response.PreviousVersion > response.NewVersion, "Version should be decremented")
	assert.NotEmpty(t, response.Message)
	assert.False(t, response.Timestamp.IsZero())
}

func TestMigrationHandler_PostMigrationRollback_Production(t *testing.T) {
	// Set production environment
	originalEnv := os.Getenv("RAILWAY_ENVIRONMENT")
	os.Setenv("RAILWAY_ENVIRONMENT", "production")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("RAILWAY_ENVIRONMENT")
		} else {
			os.Setenv("RAILWAY_ENVIRONMENT", originalEnv)
		}
	}()
	
	// Create test database with migrations
	database := setupMigrationHandlerTestDB(t)
	
	// Create migration handler
	handler := NewMigrationHandler(database)
	
	// Create rollback request
	rollbackReq := MigrationRollbackRequest{
		Confirm: true,
	}
	reqBody, err := json.Marshal(rollbackReq)
	require.NoError(t, err)
	
	// Create test request without production override
	req := httptest.NewRequest("POST", "/admin/migrations/rollback", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Call handler
	handler.PostMigrationRollback(w, req)
	
	// Check response - should be forbidden in production
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMigrationHandler_PostMigrationRollback_ProductionWithOverride(t *testing.T) {
	// Set production environment
	originalEnv := os.Getenv("RAILWAY_ENVIRONMENT")
	os.Setenv("RAILWAY_ENVIRONMENT", "production")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("RAILWAY_ENVIRONMENT")
		} else {
			os.Setenv("RAILWAY_ENVIRONMENT", originalEnv)
		}
	}()
	
	// Create test database with migrations
	database := setupMigrationHandlerTestDB(t)
	
	// Create migration handler
	handler := NewMigrationHandler(database)
	
	// Create rollback request
	rollbackReq := MigrationRollbackRequest{
		Confirm: true,
	}
	reqBody, err := json.Marshal(rollbackReq)
	require.NoError(t, err)
	
	// Create test request with production override header
	req := httptest.NewRequest("POST", "/admin/migrations/rollback", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Allow-Production-Rollback", "true")
	w := httptest.NewRecorder()
	
	// Call handler
	handler.PostMigrationRollback(w, req)
	
	// Check response - should succeed with override
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMigrationHandler_PostMigrationRollback_WithoutConfirmation(t *testing.T) {
	// Ensure we're in development environment
	originalEnv := os.Getenv("RAILWAY_ENVIRONMENT")
	os.Unsetenv("RAILWAY_ENVIRONMENT")
	defer func() {
		if originalEnv != "" {
			os.Setenv("RAILWAY_ENVIRONMENT", originalEnv)
		}
	}()
	
	// Create test database with migrations
	database := setupMigrationHandlerTestDB(t)
	
	// Create migration handler
	handler := NewMigrationHandler(database)
	
	// Create rollback request without confirmation
	rollbackReq := MigrationRollbackRequest{
		Confirm: false,
	}
	reqBody, err := json.Marshal(rollbackReq)
	require.NoError(t, err)
	
	// Create test request
	req := httptest.NewRequest("POST", "/admin/migrations/rollback", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Call handler
	handler.PostMigrationRollback(w, req)
	
	// Check response - should be bad request without confirmation
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMigrationHandler_RegisterRoutes(t *testing.T) {
	// Create test database
	database := setupMigrationHandlerTestDB(t)
	
	// Create migration handler
	handler := NewMigrationHandler(database)
	
	// Create router and register routes
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	
	// Test that routes are registered by making requests
	
	// Test status endpoint
	req := httptest.NewRequest("GET", "/admin/migrations/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Test rollback endpoint (should fail without confirmation)
	rollbackReq := MigrationRollbackRequest{Confirm: false}
	reqBody, err := json.Marshal(rollbackReq)
	require.NoError(t, err)
	
	req = httptest.NewRequest("POST", "/admin/migrations/rollback", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code) // Should fail due to no confirmation
}