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

func TestGaugeTemplateHandler(t *testing.T) {
	queries := &db.MockQueries{}
	handler := NewGaugeHandler(queries)
	
	// Setup router for URL parameter extraction
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	t.Run("CreateGaugeTemplate", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// Mock database calls
			queries.CreateGaugeTemplateFn = func(ctx context.Context, params db.CreateGaugeTemplateParams) (db.GaugeTemplate, error) {
				return db.GaugeTemplate{
					ID:          1,
					Name:        params.Name,
					Description: params.Description,
					Icon:        params.Icon,
					Unit:        params.Unit,
					Target:      params.Target,
					Frequency:   params.Frequency,
					Direction:   params.Direction,
					Active:      params.Active,
				}, nil
			}
			queries.ListGaugeTemplatesFn = func(ctx context.Context) ([]db.GaugeTemplate, error) {
				return []db.GaugeTemplate{}, nil
			}

			// Create test request with all required fields including active
			r := createFormRequest("POST", "/admin/gauges", map[string]string{
				"name":        "Test Gauge",
				"description": "Test Description",
				"icon":        "test-icon",
				"unit":        "test-unit",
				"target":      "10",
				"frequency":   "weekly",
				"direction":   "under",
				"active":      "on",
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

		t.Run("frequency validation", func(t *testing.T) {
			// Mock database calls
			queries.CreateGaugeTemplateFn = func(ctx context.Context, params db.CreateGaugeTemplateParams) (db.GaugeTemplate, error) {
				// Verify that invalid frequency defaults to "weekly"
				assert.Equal(t, "weekly", params.Frequency)
				return db.GaugeTemplate{
					ID:        1,
					Frequency: params.Frequency,
				}, nil
			}
			queries.ListGaugeTemplatesFn = func(ctx context.Context) ([]db.GaugeTemplate, error) {
				return []db.GaugeTemplate{}, nil
			}

			// Create test request with invalid frequency
			r := createFormRequest("POST", "/admin/gauges", map[string]string{
				"name":      "Test Gauge",
				"icon":      "test-icon",
				"unit":      "test-unit",
				"target":    "10",
				"frequency": "invalid-frequency", // Invalid frequency should default to weekly
				"direction": "under",
			})

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleCreateGauge(w, r)

			// Check response
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("active status validation", func(t *testing.T) {
			// Mock database calls
			queries.CreateGaugeTemplateFn = func(ctx context.Context, params db.CreateGaugeTemplateParams) (db.GaugeTemplate, error) {
				// Verify that active status is correctly parsed
				assert.True(t, params.Active)
				return db.GaugeTemplate{
					ID:     1,
					Active: params.Active,
				}, nil
			}
			queries.ListGaugeTemplatesFn = func(ctx context.Context) ([]db.GaugeTemplate, error) {
				return []db.GaugeTemplate{}, nil
			}

			// Create test request with active checkbox checked
			r := createFormRequest("POST", "/admin/gauges", map[string]string{
				"name":      "Test Gauge",
				"icon":      "test-icon",
				"unit":      "test-unit",
				"target":    "10",
				"frequency": "weekly",
				"direction": "under",
				"active":    "on", // Checkbox checked
			})

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleCreateGauge(w, r)

			// Check response
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("inactive status validation", func(t *testing.T) {
			// Mock database calls
			queries.CreateGaugeTemplateFn = func(ctx context.Context, params db.CreateGaugeTemplateParams) (db.GaugeTemplate, error) {
				// Verify that active status is correctly parsed as false when not provided
				assert.False(t, params.Active)
				return db.GaugeTemplate{
					ID:     1,
					Active: params.Active,
				}, nil
			}
			queries.ListGaugeTemplatesFn = func(ctx context.Context) ([]db.GaugeTemplate, error) {
				return []db.GaugeTemplate{}, nil
			}

			// Create test request without active checkbox (should be false)
			r := createFormRequest("POST", "/admin/gauges", map[string]string{
				"name":      "Test Gauge",
				"icon":      "test-icon",
				"unit":      "test-unit",
				"target":    "10",
				"frequency": "weekly",
				"direction": "under",
				// No "active" field - should default to false
			})

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleCreateGauge(w, r)

			// Check response
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("database error", func(t *testing.T) {
			// Mock database calls with error
			queries.CreateGaugeTemplateFn = func(ctx context.Context, params db.CreateGaugeTemplateParams) (db.GaugeTemplate, error) {
				return db.GaugeTemplate{}, fmt.Errorf("database connection failed")
			}

			// Create test request with valid data
			r := createFormRequest("POST", "/admin/gauges", map[string]string{
				"name":      "Test Gauge",
				"icon":      "test-icon",
				"unit":      "test-unit",
				"target":    "10",
				"frequency": "weekly",
				"direction": "under",
			})

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleCreateGauge(w, r)

			// Check response contains error
			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Contains(t, w.Body.String(), "Failed to create gauge template")
		})
	})

	t.Run("UpdateGaugeTemplate", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// Mock database calls
			queries.GetGaugeTemplateFn = func(ctx context.Context, id int64) (db.GaugeTemplate, error) {
				return db.GaugeTemplate{
					ID:        1,
					Name:      "Original Gauge",
					Icon:      "original-icon",
					Unit:      "original-unit",
					Target:    5.0,
					Frequency: "weekly",
					Direction: "under",
					Active:    false,
				}, nil
			}
			queries.UpdateGaugeTemplateFn = func(ctx context.Context, params db.UpdateGaugeTemplateParams) error {
				return nil
			}
			queries.ListGaugeTemplatesFn = func(ctx context.Context) ([]db.GaugeTemplate, error) {
				return []db.GaugeTemplate{}, nil
			}

			// Create test request with all required fields including active
			r := createFormRequest("PUT", "/admin/gauges/1", map[string]string{
				"name":        "Updated Gauge",
				"description": "Updated Description",
				"icon":        "updated-icon",
				"unit":        "updated-unit",
				"target":      "20",
				"frequency":   "monthly",
				"direction":   "over",
				"active":      "on",
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

		t.Run("validation error", func(t *testing.T) {
			// Create test request with invalid data (missing required fields)
			r := createFormRequest("PUT", "/admin/gauges/1", map[string]string{
				"name":   "",
				"icon":   "",
				"unit":   "",
				"target": "invalid",
			})

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "1")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleUpdateGauge(w, r)

			// Check response contains validation errors
			assert.Equal(t, http.StatusOK, w.Code) // Form validation returns OK with errors in form
			assert.Contains(t, w.Body.String(), "errors")
			assert.Contains(t, w.Body.String(), "required")
		})
	})

	t.Run("DeleteGaugeTemplate", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// Mock database calls
			queries.DeleteGaugeTemplateFn = func(ctx context.Context, id int64) error {
				return nil
			}
			queries.ListGaugeTemplatesFn = func(ctx context.Context) ([]db.GaugeTemplate, error) {
				return []db.GaugeTemplate{}, nil
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
			queries.DeleteGaugeTemplateFn = func(ctx context.Context, id int64) error {
				return fmt.Errorf("failed to delete gauge template")
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
			assert.Contains(t, w.Body.String(), "failed to delete gauge template")
		})
	})

	t.Run("IncrementGaugeInstance", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// Mock database calls
			queries.GetGaugeInstanceFn = func(ctx context.Context, id int64) (db.GaugeInstance, error) {
				if id == 1 {
					return db.GaugeInstance{
						ID:    1,
						Value: 10,
					}, nil
				}
				// Return updated value for second call
				return db.GaugeInstance{
					ID:    1,
					Value: 11,
				}, nil
			}
			queries.UpdateGaugeInstanceValueFn = func(ctx context.Context, params db.UpdateGaugeInstanceValueParams) error {
				// Verify correct parameters
				assert.Equal(t, int64(1), params.ID)
				assert.Equal(t, int64(11), params.Value)
				return nil
			}
			queries.CreateGaugeValueFn = func(ctx context.Context, params db.CreateGaugeValueParams) error {
				// Verify gauge value is created for historical tracking
				assert.Equal(t, int64(1), params.GaugeID)
				assert.Equal(t, int64(11), params.Value)
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
			assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
		})

		t.Run("invalid gauge instance ID", func(t *testing.T) {
			// Create test request with invalid ID
			r := httptest.NewRequest("POST", "/gauges/invalid/increment", nil)

			// Setup chi router context with invalid ID
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "invalid")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleIncrementGauge(w, r)

			// Check response
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "Invalid gauge instance ID")
		})

		t.Run("gauge instance not found", func(t *testing.T) {
			// Mock database calls with error
			queries.GetGaugeInstanceFn = func(ctx context.Context, id int64) (db.GaugeInstance, error) {
				return db.GaugeInstance{}, fmt.Errorf("gauge instance not found")
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
			assert.Contains(t, w.Body.String(), "Failed to get gauge instance")
		})

		t.Run("update gauge instance value fails", func(t *testing.T) {
			// Mock database calls
			queries.GetGaugeInstanceFn = func(ctx context.Context, id int64) (db.GaugeInstance, error) {
				return db.GaugeInstance{
					ID:    1,
					Value: 10,
				}, nil
			}
			queries.UpdateGaugeInstanceValueFn = func(ctx context.Context, params db.UpdateGaugeInstanceValueParams) error {
				return fmt.Errorf("database update failed")
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
			assert.Contains(t, w.Body.String(), "Failed to increment gauge instance")
		})

		t.Run("create gauge value fails", func(t *testing.T) {
			// Mock database calls
			queries.GetGaugeInstanceFn = func(ctx context.Context, id int64) (db.GaugeInstance, error) {
				if id == 1 {
					return db.GaugeInstance{
						ID:    1,
						Value: 10,
					}, nil
				}
				// Return updated value for second call
				return db.GaugeInstance{
					ID:    1,
					Value: 11,
				}, nil
			}
			queries.UpdateGaugeInstanceValueFn = func(ctx context.Context, params db.UpdateGaugeInstanceValueParams) error {
				return nil
			}
			queries.CreateGaugeValueFn = func(ctx context.Context, params db.CreateGaugeValueParams) error {
				return fmt.Errorf("failed to create gauge value")
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
			assert.Contains(t, w.Body.String(), "Failed to create gauge value")
		})

		t.Run("get updated gauge instance fails", func(t *testing.T) {
			callCount := 0
			// Mock database calls
			queries.GetGaugeInstanceFn = func(ctx context.Context, id int64) (db.GaugeInstance, error) {
				callCount++
				if callCount == 1 {
					return db.GaugeInstance{
						ID:    1,
						Value: 10,
					}, nil
				}
				// Second call fails
				return db.GaugeInstance{}, fmt.Errorf("failed to get updated gauge instance")
			}
			queries.UpdateGaugeInstanceValueFn = func(ctx context.Context, params db.UpdateGaugeInstanceValueParams) error {
				return nil
			}
			queries.CreateGaugeValueFn = func(ctx context.Context, params db.CreateGaugeValueParams) error {
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
			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Contains(t, w.Body.String(), "Failed to get updated gauge instance")
		})
	})

	t.Run("DecrementGaugeInstance", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// Mock database calls
			queries.GetGaugeInstanceFn = func(ctx context.Context, id int64) (db.GaugeInstance, error) {
				if id == 1 {
					return db.GaugeInstance{
						ID:    1,
						Value: 10,
					}, nil
				}
				// Return updated value for second call
				return db.GaugeInstance{
					ID:    1,
					Value: 9,
				}, nil
			}
			queries.UpdateGaugeInstanceValueFn = func(ctx context.Context, params db.UpdateGaugeInstanceValueParams) error {
				// Verify correct parameters
				assert.Equal(t, int64(1), params.ID)
				assert.Equal(t, int64(9), params.Value)
				return nil
			}
			queries.CreateGaugeValueFn = func(ctx context.Context, params db.CreateGaugeValueParams) error {
				// Verify gauge value is created for historical tracking
				assert.Equal(t, int64(1), params.GaugeID)
				assert.Equal(t, int64(9), params.Value)
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
			assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
		})

		t.Run("invalid gauge instance ID", func(t *testing.T) {
			// Create test request with invalid ID
			r := httptest.NewRequest("POST", "/gauges/invalid/decrement", nil)

			// Setup chi router context with invalid ID
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "invalid")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler directly
			handler.handleDecrementGauge(w, r)

			// Check response
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "Invalid gauge instance ID")
		})

		t.Run("gauge instance not found", func(t *testing.T) {
			// Mock database calls with error
			queries.GetGaugeInstanceFn = func(ctx context.Context, id int64) (db.GaugeInstance, error) {
				return db.GaugeInstance{}, fmt.Errorf("gauge instance not found")
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
			assert.Contains(t, w.Body.String(), "Failed to get gauge instance")
		})

		t.Run("prevents negative values", func(t *testing.T) {
			// Mock database calls - gauge instance with value 0
			queries.GetGaugeInstanceFn = func(ctx context.Context, id int64) (db.GaugeInstance, error) {
				return db.GaugeInstance{
					ID:    1,
					Value: 0,
				}, nil
			}
			// Should not call UpdateGaugeInstanceValue when value is 0
			queries.UpdateGaugeInstanceValueFn = func(ctx context.Context, params db.UpdateGaugeInstanceValueParams) error {
				t.Error("UpdateGaugeInstanceValue should not be called when value is 0")
				return nil
			}
			// Should not call CreateGaugeValue when value is 0
			queries.CreateGaugeValueFn = func(ctx context.Context, params db.CreateGaugeValueParams) error {
				t.Error("CreateGaugeValue should not be called when value is 0")
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

		t.Run("update gauge instance value fails", func(t *testing.T) {
			// Mock database calls
			queries.GetGaugeInstanceFn = func(ctx context.Context, id int64) (db.GaugeInstance, error) {
				return db.GaugeInstance{
					ID:    1,
					Value: 10,
				}, nil
			}
			queries.UpdateGaugeInstanceValueFn = func(ctx context.Context, params db.UpdateGaugeInstanceValueParams) error {
				return fmt.Errorf("database update failed")
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
			assert.Contains(t, w.Body.String(), "Failed to decrement gauge instance")
		})

		t.Run("create gauge value fails", func(t *testing.T) {
			// Mock database calls
			queries.GetGaugeInstanceFn = func(ctx context.Context, id int64) (db.GaugeInstance, error) {
				if id == 1 {
					return db.GaugeInstance{
						ID:    1,
						Value: 10,
					}, nil
				}
				// Return updated value for second call
				return db.GaugeInstance{
					ID:    1,
					Value: 9,
				}, nil
			}
			queries.UpdateGaugeInstanceValueFn = func(ctx context.Context, params db.UpdateGaugeInstanceValueParams) error {
				return nil
			}
			queries.CreateGaugeValueFn = func(ctx context.Context, params db.CreateGaugeValueParams) error {
				return fmt.Errorf("failed to create gauge value")
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
			assert.Contains(t, w.Body.String(), "Failed to create gauge value")
		})

		t.Run("get updated gauge instance fails", func(t *testing.T) {
			callCount := 0
			// Mock database calls
			queries.GetGaugeInstanceFn = func(ctx context.Context, id int64) (db.GaugeInstance, error) {
				callCount++
				if callCount == 1 {
					return db.GaugeInstance{
						ID:    1,
						Value: 10,
					}, nil
				}
				// Second call fails
				return db.GaugeInstance{}, fmt.Errorf("failed to get updated gauge instance")
			}
			queries.UpdateGaugeInstanceValueFn = func(ctx context.Context, params db.UpdateGaugeInstanceValueParams) error {
				return nil
			}
			queries.CreateGaugeValueFn = func(ctx context.Context, params db.CreateGaugeValueParams) error {
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
			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Contains(t, w.Body.String(), "Failed to get updated gauge instance")
		})
	})
}
