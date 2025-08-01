package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupHealthTestDB(t *testing.T) *sql.DB {
	// Create in-memory database for testing
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err, "Should be able to open in-memory database")
	
	t.Cleanup(func() {
		db.Close()
	})
	
	return db
}

func TestHealthHandler_HealthCheck(t *testing.T) {
	// Create test database
	db := setupHealthTestDB(t)
	
	// Create health handler
	handler := NewHealthHandler(db, "test-version")
	
	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	// Call handler
	handler.HealthCheck(w, req)
	
	// Check response - should be OK even without migrations (database is accessible)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	
	// Parse response
	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Should be able to parse response")
	
	// Verify response structure - status might be degraded due to migration issues
	assert.Contains(t, []string{"healthy", "degraded"}, response.Status)
	assert.Equal(t, "test-version", response.Version)
	assert.True(t, response.Database)
	assert.NotNil(t, response.Migrations)
	
	// Verify migration health is included
	assert.NotNil(t, response.Migrations)
	assert.NotEmpty(t, response.Migrations.Status)
}

func TestHealthHandler_HealthCheckWithNilDB(t *testing.T) {
	// Create health handler with nil database
	handler := NewHealthHandler(nil, "test-version")
	
	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	// Call handler
	handler.HealthCheck(w, req)
	
	// Check response
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	
	// Parse response
	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Should be able to parse response")
	
	// Verify response structure
	assert.Equal(t, "unhealthy", response.Status)
	assert.Equal(t, "test-version", response.Version)
	assert.False(t, response.Database)
	assert.NotNil(t, response.Migrations)
	assert.Equal(t, "no_database", response.Migrations.Status)
}

func TestHealthHandler_checkMigrationStatus(t *testing.T) {
	// Test with nil database
	handler := NewHealthHandler(nil, "test-version")
	migrationHealth := handler.checkMigrationStatus()
	
	assert.NotNil(t, migrationHealth)
	assert.Equal(t, int64(-1), migrationHealth.CurrentVersion)
	assert.False(t, migrationHealth.IsUpToDate)
	assert.Equal(t, "no_database", migrationHealth.Status)
	
	// Test with valid database (but no migrations applied)
	db := setupHealthTestDB(t)
	handler = NewHealthHandler(db, "test-version")
	migrationHealth = handler.checkMigrationStatus()
	
	assert.NotNil(t, migrationHealth)
	// Migration status might be error due to no goose table
	assert.NotEmpty(t, migrationHealth.Status)
}