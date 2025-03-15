package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"health-monitor/internal/db"
	"health-monitor/internal/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestGaugeHandler(t *testing.T) {
	q := testutil.NewTestDB(t)
	h := NewGaugeHandler(q)

	t.Run("GetGauge", func(t *testing.T) {
		gauge := testutil.CreateTestGauge(t, q)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", fmt.Sprintf("/gauges/%d", gauge.ID), nil)

		// Set up chi router context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", fmt.Sprint(gauge.ID))
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

		h.GetGauge(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		var response db.Gauge
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, gauge.ID, response.ID)
	})

	t.Run("CreateGauge", func(t *testing.T) {
		params := db.CreateGaugeParams{
			Name:   "Test Gauge",
			Target: 100,
			Unit:   "units",
			Icon:   "star",
		}

		body, err := json.Marshal(params)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/gauges", bytes.NewReader(body))

		h.CreateGauge(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		var response db.Gauge
		err = json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, params.Name, response.Name)
	})

	t.Run("UpdateGauge", func(t *testing.T) {
		gauge := testutil.CreateTestGauge(t, q)

		params := db.UpdateGaugeParams{
			ID:     gauge.ID,
			Name:   "Updated Gauge",
			Target: 200,
			Unit:   "updated_units",
			Icon:   "updated_star",
		}

		body, err := json.Marshal(params)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("PUT", fmt.Sprintf("/gauges/%d", gauge.ID), bytes.NewReader(body))

		// Set up chi router context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", fmt.Sprint(gauge.ID))
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

		h.UpdateGauge(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DeleteGauge", func(t *testing.T) {
		gauge := testutil.CreateTestGauge(t, q)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("DELETE", fmt.Sprintf("/gauges/%d", gauge.ID), nil)

		// Set up chi router context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", fmt.Sprint(gauge.ID))
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

		h.DeleteGauge(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetAllGauges", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/gauges", nil)

		h.GetAllGauges(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []db.Gauge
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
	})

	t.Run("CreateGaugeValue", func(t *testing.T) {
		gauge := testutil.CreateTestGauge(t, q)

		input := struct {
			Value float64 `json:"value"`
			Date  string  `json:"date"`
		}{
			Value: 85.5,
			Date:  "2025-03-15T14:25:27+11:00",
		}

		body, err := json.Marshal(input)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", fmt.Sprintf("/gauges/%d/values", gauge.ID), bytes.NewReader(body))

		// Set up chi router context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", fmt.Sprint(gauge.ID))
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

		h.CreateGaugeValue(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
