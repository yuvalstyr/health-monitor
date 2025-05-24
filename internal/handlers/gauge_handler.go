package handlers

import (
	"context"
	"fmt"
	"health-monitor/internal/db"
	"health-monitor/internal/views/pages"
	"net/http"
	"strconv"
	"github.com/a-h/templ"
)

type Querier interface {
	GetAllGauges(ctx context.Context) ([]db.Gauge, error)
	GetGauge(ctx context.Context, id int64) (db.Gauge, error)
	CreateGauge(ctx context.Context, params db.CreateGaugeParams) (db.Gauge, error)
	UpdateGauge(ctx context.Context, params db.UpdateGaugeParams) error
	DeleteGauge(ctx context.Context, id int64) error
	UpdateGaugeValue(ctx context.Context, params db.UpdateGaugeValueParams) error
}

type GaugeHandler struct {
	queries Querier
}

func NewGaugeHandler(queries Querier) *GaugeHandler {
	return &GaugeHandler{
		queries: queries,
	}
}

func (h *GaugeHandler) Admin(r *http.Request) (templ.Component, error) {
	gauges, err := h.queries.GetAllGauges(r.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get gauges: %w", err)
	}
	return pages.Admin(gauges), nil
}

func (h *GaugeHandler) Create(r *http.Request) (templ.Component, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}

	name := r.FormValue("name")
	icon := r.FormValue("icon")
	unit := r.FormValue("unit")
	frequency := r.FormValue("frequency")
	direction := r.FormValue("direction")
	target, err := strconv.ParseFloat(r.FormValue("target"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid target value: %w", err)
	}

	// Validate frequency and direction
	if frequency != "weekly" && frequency != "biweekly" && frequency != "monthly" {
		frequency = "monthly" // Default to monthly if invalid
	}

	if direction != "under" && direction != "over" {
		direction = "under" // Default to under if invalid
	}

	_, err = h.queries.CreateGauge(r.Context(), db.CreateGaugeParams{
		Name:      name,
		Icon:      icon,
		Unit:      unit,
		Target:    target,
		Frequency: frequency,
		Direction: direction,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gauge: %w", err)
	}

	return h.Admin(r)
}

func (h *GaugeHandler) Update(id int64, r *http.Request) (templ.Component, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}

	name := r.FormValue("name")
	icon := r.FormValue("icon")
	unit := r.FormValue("unit")
	frequency := r.FormValue("frequency")
	direction := r.FormValue("direction")
	target, err := strconv.ParseFloat(r.FormValue("target"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid target value: %w", err)
	}

	// Validate frequency and direction
	if frequency != "weekly" && frequency != "biweekly" && frequency != "monthly" {
		frequency = "monthly" // Default to monthly if invalid
	}

	if direction != "under" && direction != "over" {
		direction = "under" // Default to under if invalid
	}

	err = h.queries.UpdateGauge(r.Context(), db.UpdateGaugeParams{
		ID:        id,
		Name:      name,
		Icon:      icon,
		Unit:      unit,
		Target:    target,
		Frequency: frequency,
		Direction: direction,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update gauge: %w", err)
	}

	return h.Admin(r)
}

func (h *GaugeHandler) Delete(r *http.Request, id int64) (templ.Component, error) {
	err := h.queries.DeleteGauge(r.Context(), id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete gauge: %w", err)
	}

	return h.Admin(r)
}

func (h *GaugeHandler) Increment(r *http.Request, id int64) (templ.Component, error) {
	gauge, err := h.queries.GetGauge(r.Context(), id)
	if err != nil {
		return nil, fmt.Errorf("failed to get gauge: %w", err)
	}

	err = h.queries.UpdateGaugeValue(r.Context(), db.UpdateGaugeValueParams{
		ID:    id,
		Value: gauge.Value + 1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to increment gauge: %w", err)
	}

	return h.Admin(r)
}

func (h *GaugeHandler) Decrement(r *http.Request, id int64) (templ.Component, error) {
	gauge, err := h.queries.GetGauge(r.Context(), id)
	if err != nil {
		return nil, fmt.Errorf("failed to get gauge: %w", err)
	}

	if gauge.Value > 0 {
		err = h.queries.UpdateGaugeValue(r.Context(), db.UpdateGaugeValueParams{
			ID:    id,
			Value: gauge.Value - 1,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to decrement gauge: %w", err)
		}
	}

	return h.Admin(r)
}
