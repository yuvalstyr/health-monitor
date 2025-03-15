package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"health-monitor/internal/db"
)

type GaugeHandler struct {
	queries *db.Queries
}

func NewGaugeHandler(queries *db.Queries) *GaugeHandler {
	return &GaugeHandler{
		queries: queries,
	}
}

// GetGauge handles retrieving a single gauge by ID
func (h *GaugeHandler) GetGauge(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid gauge id", http.StatusBadRequest)
		return
	}

	gauge, err := h.queries.GetGauge(r.Context(), id)
	if err != nil {
		http.Error(w, "gauge not found", http.StatusNotFound)
		return
	}

	WriteJSON(w, gauge)
}

// CreateGauge handles creating a new gauge
func (h *GaugeHandler) CreateGauge(w http.ResponseWriter, r *http.Request) {
	var params db.CreateGaugeParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	gauge, err := h.queries.CreateGauge(r.Context(), params)
	if err != nil {
		http.Error(w, "failed to create gauge", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, gauge)
}

// UpdateGauge handles updating an existing gauge
func (h *GaugeHandler) UpdateGauge(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid gauge id", http.StatusBadRequest)
		return
	}

	var params db.UpdateGaugeParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	params.ID = id

	if err := h.queries.UpdateGauge(r.Context(), params); err != nil {
		http.Error(w, "failed to update gauge", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"status": "ok"})
}

// DeleteGauge handles deleting a gauge
func (h *GaugeHandler) DeleteGauge(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid gauge id", http.StatusBadRequest)
		return
	}

	if err := h.queries.DeleteGauge(r.Context(), id); err != nil {
		http.Error(w, "failed to delete gauge", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"status": "ok"})
}

// GetAllGauges handles retrieving all gauges
func (h *GaugeHandler) GetAllGauges(w http.ResponseWriter, r *http.Request) {
	gauges, err := h.queries.ListGauges(r.Context())
	if err != nil {
		http.Error(w, "failed to list gauges", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, gauges)
}

// CreateGaugeValue handles creating a new gauge value
func (h *GaugeHandler) CreateGaugeValue(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid gauge id", http.StatusBadRequest)
		return
	}

	var input struct {
		Value float64   `json:"value"`
		Date  time.Time `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	params := db.CreateGaugeValueParams{
		GaugeID: id,
		Column2: input.Value,
		Date:    input.Date,
	}

	if err := h.queries.CreateGaugeValue(r.Context(), params); err != nil {
		http.Error(w, "failed to create gauge value", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"status": "ok"})
}

// WriteJSON writes JSON to response writer
func WriteJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
