package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"health-monitor/internal/db"
	"health-monitor/internal/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestGaugeHandler(t *testing.T) {
	q := testutil.NewTestDB(t)
	h := NewGaugeHandler(q)

	t.Run("GetGauge", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			gauge := testutil.CreateTestGauge(t, q)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", fmt.Sprintf("/gauges/%d", gauge.ID), nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", fmt.Sprint(gauge.ID))
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.GetGauge(w, r)

			assert.Equal(t, http.StatusOK, w.Code)

			var response db.Gauge
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, gauge.ID, response.ID)
			assert.Equal(t, gauge.Name, response.Name)
			assert.Equal(t, gauge.Target, response.Target)
			assert.Equal(t, gauge.Unit, response.Unit)
		})

		t.Run("not found", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/gauges/999", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "999")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.GetGauge(w, r)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})

		t.Run("invalid id", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/gauges/invalid", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "invalid")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.GetGauge(w, r)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})

	t.Run("CreateGauge", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
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
			assert.Equal(t, params.Target, response.Target)
			assert.Equal(t, params.Unit, response.Unit)
			assert.Equal(t, params.Icon, response.Icon)
		})

		t.Run("invalid json", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/gauges", bytes.NewReader([]byte("invalid json")))

			h.CreateGauge(w, r)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("missing required fields", func(t *testing.T) {
			params := struct {
				Name string `json:"name"`
			}{
				Name: "Test Gauge",
			}

			body, err := json.Marshal(params)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/gauges", bytes.NewReader(body))

			h.CreateGauge(w, r)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})

	t.Run("UpdateGauge", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
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

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", fmt.Sprint(gauge.ID))
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.UpdateGauge(w, r)

			assert.Equal(t, http.StatusOK, w.Code)

			var response db.Gauge
			err = json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, params.Name, response.Name)
			assert.Equal(t, params.Target, response.Target)
			assert.Equal(t, params.Unit, response.Unit)
			assert.Equal(t, params.Icon, response.Icon)
		})

		t.Run("not found", func(t *testing.T) {
			params := db.UpdateGaugeParams{
				ID:     999,
				Name:   "Updated Gauge",
				Target: 200,
				Unit:   "updated_units",
				Icon:   "updated_star",
			}

			body, err := json.Marshal(params)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("PUT", "/gauges/999", bytes.NewReader(body))

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "999")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.UpdateGauge(w, r)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})

		t.Run("invalid json", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("PUT", "/gauges/1", bytes.NewReader([]byte("invalid json")))

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "1")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.UpdateGauge(w, r)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})

	t.Run("DeleteGauge", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			gauge := testutil.CreateTestGauge(t, q)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("DELETE", fmt.Sprintf("/gauges/%d", gauge.ID), nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", fmt.Sprint(gauge.ID))
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.DeleteGauge(w, r)

			assert.Equal(t, http.StatusOK, w.Code)

			// Verify gauge is deleted
			w = httptest.NewRecorder()
			r = httptest.NewRequest("GET", fmt.Sprintf("/gauges/%d", gauge.ID), nil)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.GetGauge(w, r)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})

		t.Run("not found", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("DELETE", "/gauges/999", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "999")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.DeleteGauge(w, r)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})

	t.Run("GetAllGauges", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// Create multiple gauges
			gauge1 := testutil.CreateTestGauge(t, q)
			gauge2 := testutil.CreateTestGauge(t, q)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/gauges", nil)

			h.GetAllGauges(w, r)

			assert.Equal(t, http.StatusOK, w.Code)

			var response []db.Gauge
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(response), 2)

			ids := make(map[int64]bool)
			for _, g := range response {
				ids[g.ID] = true
			}
			assert.True(t, ids[gauge1.ID])
			assert.True(t, ids[gauge2.ID])
		})
	})

	t.Run("CreateGaugeValue", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			gauge := testutil.CreateTestGauge(t, q)
			now := time.Now()

			input := struct {
				Value float64 `json:"value"`
				Date  string  `json:"date"`
			}{
				Value: 85.5,
				Date:  now.Format(time.RFC3339),
			}

			body, err := json.Marshal(input)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", fmt.Sprintf("/gauges/%d/values", gauge.ID), bytes.NewReader(body))

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", fmt.Sprint(gauge.ID))
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.CreateGaugeValue(w, r)

			assert.Equal(t, http.StatusOK, w.Code)

			// Verify value was created
			values, err := q.GetGaugeValues(context.Background(), gauge.ID)
			assert.NoError(t, err)
			assert.Len(t, values, 1)
			assert.Equal(t, input.Value, values[0].Value)
		})

		t.Run("gauge not found", func(t *testing.T) {
			input := struct {
				Value float64 `json:"value"`
				Date  string  `json:"date"`
			}{
				Value: 85.5,
				Date:  time.Now().Format(time.RFC3339),
			}

			body, err := json.Marshal(input)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/gauges/999/values", bytes.NewReader(body))

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "999")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.CreateGaugeValue(w, r)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})

		t.Run("invalid date format", func(t *testing.T) {
			gauge := testutil.CreateTestGauge(t, q)

			input := struct {
				Value float64 `json:"value"`
				Date  string  `json:"date"`
			}{
				Value: 85.5,
				Date:  "invalid-date",
			}

			body, err := json.Marshal(input)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", fmt.Sprintf("/gauges/%d/values", gauge.ID), bytes.NewReader(body))

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", fmt.Sprint(gauge.ID))
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.CreateGaugeValue(w, r)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})

	t.Run("GetGaugeHistory", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			gauge := testutil.CreateTestGauge(t, q)

			// Create some historical values
			dates := []time.Time{
				time.Now().AddDate(0, -2, 0), // 2 months ago
				time.Now().AddDate(0, -1, 0), // 1 month ago
				time.Now(),                   // current month
			}

			for _, date := range dates {
				err := testutil.CreateTestGaugeValue(t, q, gauge.ID, 85.5, date)
				assert.NoError(t, err)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", fmt.Sprintf("/gauges/%d/history", gauge.ID), nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", fmt.Sprint(gauge.ID))
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.GetGaugeHistory(w, r)

			assert.Equal(t, http.StatusOK, w.Code)

			var response []db.GetGaugeHistoryRow
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Len(t, response, 3)
		})

		t.Run("gauge not found", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/gauges/999/history", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "999")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.GetGaugeHistory(w, r)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})

		t.Run("no history", func(t *testing.T) {
			gauge := testutil.CreateTestGauge(t, q)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", fmt.Sprintf("/gauges/%d/history", gauge.ID), nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", fmt.Sprint(gauge.ID))
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h.GetGaugeHistory(w, r)

			assert.Equal(t, http.StatusOK, w.Code)

			var response []db.GetGaugeHistoryRow
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Empty(t, response)
		})
	})
}
