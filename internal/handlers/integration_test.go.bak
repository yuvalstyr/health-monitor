package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"health-monitor/internal/db"
	"health-monitor/internal/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func setupTestServer(t *testing.T) (*chi.Mux, *db.Queries) {
	queries := testutil.NewTestDB(t)
	handler := NewGaugeHandler(queries)

	r := chi.NewRouter()
	r.Get("/gauges", handler.GetAllGauges)
	r.Post("/gauges", handler.CreateGauge)
	r.Get("/gauges/{id}", handler.GetGauge)
	r.Put("/gauges/{id}", handler.UpdateGauge)
	r.Delete("/gauges/{id}", handler.DeleteGauge)
	r.Post("/gauges/{id}/values", handler.CreateGaugeValue)
	r.Get("/gauges/{id}/history", handler.GetGaugeHistory)

	return r, queries
}

func TestFullGaugeLifecycle(t *testing.T) {
	router, queries := setupTestServer(t)

	// Test creating a gauge
	createParams := db.CreateGaugeParams{
		Name:        "Steps",
		Description: sql.NullString{String: "Daily step count", Valid: true},
		Target:      10000,
		Unit:        "steps",
		Icon:        "footsteps",
	}
	createBody, _ := json.Marshal(createParams)

	req := httptest.NewRequest("POST", "/gauges", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var createdGauge db.Gauge
	err := json.NewDecoder(w.Body).Decode(&createdGauge)
	assert.NoError(t, err)
	assert.Equal(t, createParams.Name, createdGauge.Name)

	// Test getting the gauge
	req = httptest.NewRequest("GET", "/gauges/"+strconv.FormatInt(createdGauge.ID, 10), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var retrievedGauge db.Gauge
	err = json.NewDecoder(w.Body).Decode(&retrievedGauge)
	assert.NoError(t, err)
	assert.Equal(t, createdGauge, retrievedGauge)

	// Test creating gauge values
	now := time.Now().UTC()
	valueInput := struct {
		Value float64   `json:"value"`
		Date  time.Time `json:"date"`
	}{
		Value: 8000,
		Date:  now,
	}
	valueBody, _ := json.Marshal(valueInput)

	req = httptest.NewRequest("POST", "/gauges/"+strconv.FormatInt(createdGauge.ID, 10)+"/values", bytes.NewBuffer(valueBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify value was created
	values, err := queries.GetGaugeValues(req.Context(), createdGauge.ID)
	assert.NoError(t, err)
	assert.Len(t, values, 1)
	assert.Equal(t, valueInput.Value, values[0].Value)

	// Test getting gauge history
	req = httptest.NewRequest("GET", "/gauges/"+strconv.FormatInt(createdGauge.ID, 10)+"/history", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var history []db.GetGaugeHistoryRow
	err = json.NewDecoder(w.Body).Decode(&history)
	assert.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, 8000.0, history[0].AverageValue)

	// Test HTMX response
	t.Run("HTMX response", func(t *testing.T) {
		req = httptest.NewRequest("GET", "/gauges/"+strconv.FormatInt(createdGauge.ID, 10), nil)
		req.Header.Set("HX-Request", "true")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
		assert.Equal(t, "refresh", w.Header().Get("HX-Trigger"))
	})

	// Test form submission with HTMX
	t.Run("HTMX form submission", func(t *testing.T) {
		formData := url.Values{
			"name":        []string{"Weight"},
			"description": []string{"Daily weight measurement"},
			"target":      []string{"70"},
			"unit":        []string{"kg"},
			"icon":        []string{"scale"},
		}

		req = httptest.NewRequest("POST", "/gauges", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
		assert.Equal(t, "refresh", w.Header().Get("HX-Trigger"))
	})
}

func TestErrorHandling(t *testing.T) {
	router, _ := setupTestServer(t)

	t.Run("invalid gauge ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/gauges/invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("gauge not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/gauges/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/gauges", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGaugeHandlerIntegration(t *testing.T) {
	queries := &db.MockQueries{
		GetAllGaugesFn: func(ctx context.Context) ([]db.Gauge, error) {
			return []db.Gauge{}, nil
		},
		GetGaugeFn: func(ctx context.Context, id int64) (db.Gauge, error) {
			return db.Gauge{
				ID:     id,
				Name:   "Test Gauge",
				Icon:   "test-icon",
				Unit:   "test-unit",
				Target: 10.0,
				Value:  5.0,
			}, nil
		},
		CreateGaugeFn: func(ctx context.Context, params db.CreateGaugeParams) (db.Gauge, error) {
			return db.Gauge{
				ID:     1,
				Name:   params.Name,
				Icon:   params.Icon,
				Unit:   params.Unit,
				Target: params.Target,
				Value:  0,
			}, nil
		},
		UpdateGaugeFn: func(ctx context.Context, params db.UpdateGaugeParams) error {
			return nil
		},
		DeleteGaugeFn: func(ctx context.Context, id int64) error {
			return nil
		},
		UpdateGaugeValueFn: func(ctx context.Context, params db.UpdateGaugeValueParams) error {
			return nil
		},
	}

	handler := NewGaugeHandler(queries)

	t.Run("Admin", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/admin", nil)
		component, err := handler.Admin(r)
		assert.NoError(t, err)
		assert.NotNil(t, component)
	})

	t.Run("Create", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/gauges", nil)
		r.Form = map[string][]string{
			"name":   {"Test Gauge"},
			"icon":   {"test-icon"},
			"unit":   {"test-unit"},
			"target": {"10.0"},
		}

		component, err := handler.Create(r)
		assert.NoError(t, err)
		assert.NotNil(t, component)
	})

	t.Run("Update", func(t *testing.T) {
		r := httptest.NewRequest("PUT", "/gauges/1", nil)
		r.Form = map[string][]string{
			"name":   {"Updated Gauge"},
			"icon":   {"updated-icon"},
			"unit":   {"updated-unit"},
			"target": {"20.0"},
		}

		component, err := handler.Update(1, r)
		assert.NoError(t, err)
		assert.NotNil(t, component)
	})

	t.Run("Delete", func(t *testing.T) {
		r := httptest.NewRequest("DELETE", "/gauges/1", nil)
		component, err := handler.Delete(r, 1)
		assert.NoError(t, err)
		assert.NotNil(t, component)
	})

	t.Run("Increment", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/gauges/1/increment", nil)
		component, err := handler.Increment(r, 1)
		assert.NoError(t, err)
		assert.NotNil(t, component)
	})

	t.Run("Decrement", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/gauges/1/decrement", nil)
		component, err := handler.Decrement(r, 1)
		assert.NoError(t, err)
		assert.NotNil(t, component)
	})
}
