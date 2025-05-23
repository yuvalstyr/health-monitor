package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// Dummy dependencies for now, replace with real ones as needed
func setupTestRouter() http.Handler {
	r := chi.NewRouter()
	// TODO: Register your real routes here, e.g. RegisterAdminRoutes(r, queries, templates)
	r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ADMIN ROUTE HANDLER TRIGGERED"))
	})
	return r
}

func TestAdminRoute_OK(t *testing.T) {
	r := setupTestRouter()
	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "ADMIN ROUTE HANDLER TRIGGERED")
}

// Add similar tests for other main routes (/, /gauges, etc.)
