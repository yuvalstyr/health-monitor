package handlers

import (
	"context"
	"fmt"
	"health-monitor/internal/db"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a test request with form values
func createFormRequest(method, path string, formValues map[string]string) *http.Request {
	form := url.Values{}
	for key, value := range formValues {
		form.Add(key, value)
	}
	body := strings.NewReader(form.Encode())
	r := httptest.NewRequest(method, path, body)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}

func TestGaugeHandler(t *testing.T) {
	queries := &db.MockQueries{}
	handler := NewGaugeHandler(queries)
	
	// Setup router for URL parameter extraction
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	t.Run("Create", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// Mock database calls
			queries.CreateGaugeFn = func(ctx context.Context, params db.CreateGaugeParams) (db.Gauge, error) {
				return db.Gauge{
					ID:     1,
					Name:   params.Name,
					Icon:   params.Icon,
					Unit:   params.Unit,
					Target: params.Target,
					Value:  0,
				}, nil
			}
			queries.ListGaugesFn = func(ctx context.Context) ([]db.Gauge, error) {
				return []db.Gauge{}, nil
			}

			// Create test request
			r := createFormRequest("POST", "/admin/gauges", map[string]string{
				"name":   "Test Gauge",
				"icon":   "test-icon",
				"unit":   "test-unit",
				"target": "10",
			})

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleCreateGauge(w, r)

			// Check response
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "html")
		})

		t.Run("validation error", func(t *testing.T) {
			// Create test request with invalid data
			r := createFormRequest("POST", "/admin/gauges", map[string]string{
				"name":   "",
				"icon":   "",
				"unit":   "",
				"target": "abc",
			})

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleCreateGauge(w, r)

			// Check response contains validation errors
			assert.Equal(t, http.StatusOK, w.Code) // Form validation returns OK with errors in form
			assert.Contains(t, w.Body.String(), "errors")
			assert.Contains(t, w.Body.String(), "required")
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// Mock database calls
			queries.GetGaugeFn = func(ctx context.Context, id int64) (db.Gauge, error) {
				return db.Gauge{
					ID:     1,
					Name:   "Original Gauge",
					Icon:   "original-icon",
					Unit:   "original-unit",
					Target: 5.0,
				}, nil
			}
			queries.UpdateGaugeFn = func(ctx context.Context, params db.UpdateGaugeParams) error {
				return nil
			}
			queries.ListGaugesFn = func(ctx context.Context) ([]db.Gauge, error) {
				return []db.Gauge{}, nil
			}

			// Create test request
			r := createFormRequest("PUT", "/admin/gauges/1", map[string]string{
				"name":   "Updated Gauge",
				"icon":   "updated-icon",
				"unit":   "updated-unit",
				"target": "20",
			})

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "1")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleUpdateGauge(w, r)

			// Check response
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "html")
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// Mock database calls
			queries.DeleteGaugeFn = func(ctx context.Context, id int64) error {
				return nil
			}
			queries.ListGaugesFn = func(ctx context.Context) ([]db.Gauge, error) {
				return []db.Gauge{}, nil
			}

			// Create test request
			r := httptest.NewRequest("DELETE", "/admin/gauges/1", nil)

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "1")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleDeleteGauge(w, r)

			// Check response
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("error", func(t *testing.T) {
			// Mock database calls with error
			queries.DeleteGaugeFn = func(ctx context.Context, id int64) error {
				return fmt.Errorf("failed to delete gauge")
			}

			// Create test request
			r := httptest.NewRequest("DELETE", "/admin/gauges/1", nil)

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "1")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleDeleteGauge(w, r)

			// Check response
			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Contains(t, w.Body.String(), "failed to delete gauge")
		})
	})

	t.Run("Increment", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// Mock database calls
			queries.GetGaugeFn = func(ctx context.Context, id int64) (db.Gauge, error) {
				return db.Gauge{
					ID:    1,
					Value: 10,
				}, nil
			}
			queries.UpdateGaugeValueFn = func(ctx context.Context, params db.UpdateGaugeValueParams) error {
				return nil
			}

			// Create test request
			r := httptest.NewRequest("POST", "/gauges/1/increment", nil)

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "1")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleIncrementGauge(w, r)

			// Check response
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("error", func(t *testing.T) {
			// Mock database calls with error
			queries.GetGaugeFn = func(ctx context.Context, id int64) (db.Gauge, error) {
				return db.Gauge{}, fmt.Errorf("failed to get gauge")
			}

			// Create test request
			r := httptest.NewRequest("POST", "/gauges/1/increment", nil)

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "1")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleIncrementGauge(w, r)

			// Check response
			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Contains(t, w.Body.String(), "failed to get gauge")
		})
	})

	t.Run("Decrement", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// Mock database calls
			queries.GetGaugeFn = func(ctx context.Context, id int64) (db.Gauge, error) {
				return db.Gauge{
					ID:    1,
					Value: 10,
				}, nil
			}
			queries.UpdateGaugeValueFn = func(ctx context.Context, params db.UpdateGaugeValueParams) error {
				return nil
			}

			// Create test request
			r := httptest.NewRequest("POST", "/gauges/1/decrement", nil)

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "1")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleDecrementGauge(w, r)

			// Check response
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("error", func(t *testing.T) {
			// Mock database calls with error
			queries.GetGaugeFn = func(ctx context.Context, id int64) (db.Gauge, error) {
				return db.Gauge{}, fmt.Errorf("failed to get gauge")
			}

			// Create test request
			r := httptest.NewRequest("POST", "/gauges/1/decrement", nil)

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "1")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleDecrementGauge(w, r)

			// Check response
			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Contains(t, w.Body.String(), "failed to get gauge")
		})
	})
}
