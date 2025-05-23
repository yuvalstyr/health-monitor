package handlers

import (
	"context"
	"fmt"
	"health-monitor/internal/db"
	"health-monitor/internal/views/components"
	"health-monitor/internal/views/pages"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Querier interface {
	ListGauges(ctx context.Context) ([]db.Gauge, error)
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

// RegisterRoutes registers all gauge-related routes on the provided router
func (h *GaugeHandler) RegisterRoutes(r chi.Router) {
	// Admin dashboard
	r.Get("/admin", h.handleAdmin)

	// Gauge routes
	r.Route("/admin/gauges", func(r chi.Router) {
		// New gauge form
		r.Get("/new", h.handleNewGaugeForm)

		// Create gauge
		r.Post("/", h.handleCreateGauge)

		// Edit gauge routes
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.handleEditGaugeForm)
			r.Put("/", h.handleUpdateGauge)
			r.Delete("/", h.handleDeleteGauge)
		})
	})

	// Gauge HTMX actions
	r.Route("/gauges/{id}", func(r chi.Router) {
		r.Post("/increment", h.handleIncrementGauge)
		r.Post("/decrement", h.handleDecrementGauge)
	})
}

// handleAdmin renders the admin dashboard page
func (h *GaugeHandler) handleAdmin(w http.ResponseWriter, r *http.Request) {
	gauges, err := h.queries.ListGauges(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get gauges: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = components.Layout(pages.Admin(gauges)).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleNewGaugeForm renders the form for creating a new gauge
func (h *GaugeHandler) handleNewGaugeForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	err := components.Layout(components.GaugeForm("POST", "/admin/gauges", nil, []components.FormError{})).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// validateGaugeForm validates the form data for gauge creation/updates
func validateGaugeForm(r *http.Request) (string, string, string, float64, []components.FormError) {
	var errors []components.FormError
	
	// Validate name
	name := r.FormValue("name")
	if name == "" {
		errors = append(errors, components.FormError{Field: "name", Message: "Name is required"})
	}

	// Validate icon
	icon := r.FormValue("icon")
	if icon == "" {
		errors = append(errors, components.FormError{Field: "icon", Message: "Icon is required"})
	}

	// Validate unit
	unit := r.FormValue("unit")
	if unit == "" {
		errors = append(errors, components.FormError{Field: "unit", Message: "Unit is required"})
	}

	// Validate target
	targetStr := r.FormValue("target")
	target, err := strconv.ParseFloat(targetStr, 64)
	if err != nil || targetStr == "" {
		errors = append(errors, components.FormError{Field: "target", Message: "Target must be a valid number"})
		target = 0
	}

	return name, icon, unit, target, errors
}

// handleCreateGauge handles the creation of a new gauge
func (h *GaugeHandler) handleCreateGauge(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	name, icon, unit, target, errors := validateGaugeForm(r)

	// If there are validation errors, re-render the form
	if len(errors) > 0 {
		w.Header().Set("Content-Type", "text/html")
		// Create a dummy gauge to maintain form values
		dummyGauge := &db.Gauge{
			Name: name,
			Icon: icon,
			Unit: unit,
			Target: target,
		}
		err := components.Layout(components.GaugeForm("POST", "/admin/gauges", dummyGauge, errors)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Create the gauge
	_, err := h.queries.CreateGauge(r.Context(), db.CreateGaugeParams{
		Name:   name,
		Icon:   icon,
		Unit:   unit,
		Target: target,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create gauge: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to admin page after successful creation
	h.handleAdmin(w, r)
}

// handleEditGaugeForm renders the form for editing an existing gauge
func (h *GaugeHandler) handleEditGaugeForm(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid gauge ID: %v", err), http.StatusBadRequest)
		return
	}

	// Get the gauge
	gauge, err := h.queries.GetGauge(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get gauge: %v", err), http.StatusInternalServerError)
		return
	}

	// Render the edit form
	w.Header().Set("Content-Type", "text/html")
	err = components.Layout(components.GaugeForm("PUT", fmt.Sprintf("/admin/gauges/%d", id), &gauge, []components.FormError{})).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleUpdateGauge handles updating an existing gauge
func (h *GaugeHandler) handleUpdateGauge(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid gauge ID: %v", err), http.StatusBadRequest)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	// Validate form data
	name, icon, unit, target, errors := validateGaugeForm(r)

	// If there are validation errors, re-render the form
	if len(errors) > 0 {
		w.Header().Set("Content-Type", "text/html")
		// Create a gauge with the current values to maintain form state
		currentGauge := db.Gauge{
			ID:     id,
			Name:   name,
			Icon:   icon,
			Unit:   unit,
			Target: target,
		}
		err := components.Layout(components.GaugeForm("PUT", fmt.Sprintf("/admin/gauges/%d", id), &currentGauge, errors)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Update the gauge
	err = h.queries.UpdateGauge(r.Context(), db.UpdateGaugeParams{
		ID:     id,
		Name:   name,
		Icon:   icon,
		Unit:   unit,
		Target: target,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update gauge: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to admin page after successful update
	h.handleAdmin(w, r)
}

// handleDeleteGauge handles the deletion of a gauge
func (h *GaugeHandler) handleDeleteGauge(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid gauge ID: %v", err), http.StatusBadRequest)
		return
	}

	// Delete the gauge
	err = h.queries.DeleteGauge(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete gauge: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to admin page after successful deletion
	h.handleAdmin(w, r)
}

// handleIncrementGauge handles incrementing a gauge's value
func (h *GaugeHandler) handleIncrementGauge(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid gauge ID: %v", err), http.StatusBadRequest)
		return
	}

	// Get the current gauge
	gauge, err := h.queries.GetGauge(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get gauge: %v", err), http.StatusInternalServerError)
		return
	}

	// Increment the value
	err = h.queries.UpdateGaugeValue(r.Context(), db.UpdateGaugeValueParams{
		ID:    id,
		Value: gauge.Value + 1,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to increment gauge: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the updated gauge
	updatedGauge, err := h.queries.GetGauge(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get updated gauge: %v", err), http.StatusInternalServerError)
		return
	}

	// Render just the updated gauge component
	w.Header().Set("Content-Type", "text/html")
	err = components.Gauge(&updatedGauge).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleDecrementGauge handles decrementing a gauge's value
func (h *GaugeHandler) handleDecrementGauge(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid gauge ID: %v", err), http.StatusBadRequest)
		return
	}

	// Get the current gauge
	gauge, err := h.queries.GetGauge(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get gauge: %v", err), http.StatusInternalServerError)
		return
	}

	// Only decrement if value is greater than 0
	if gauge.Value > 0 {
		err = h.queries.UpdateGaugeValue(r.Context(), db.UpdateGaugeValueParams{
			ID:    id,
			Value: gauge.Value - 1,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to decrement gauge: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Get the updated gauge
	updatedGauge, err := h.queries.GetGauge(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get updated gauge: %v", err), http.StatusInternalServerError)
		return
	}

	// Render just the updated gauge component
	w.Header().Set("Content-Type", "text/html")
	err = components.Gauge(&updatedGauge).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
