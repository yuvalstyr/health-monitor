package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"health-monitor/internal/db"
	"health-monitor/internal/views/components"
	"health-monitor/internal/views/layouts"
	"health-monitor/internal/views/pages"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Querier interface {
	// Gauge Template methods
	ListGaugeTemplates(ctx context.Context) ([]db.GaugeTemplate, error)
	GetGaugeTemplate(ctx context.Context, id int64) (db.GaugeTemplate, error)
	CreateGaugeTemplate(ctx context.Context, params db.CreateGaugeTemplateParams) (db.GaugeTemplate, error)
	UpdateGaugeTemplate(ctx context.Context, params db.UpdateGaugeTemplateParams) error
	DeleteGaugeTemplate(ctx context.Context, id int64) error
	
	// Gauge Instance methods (for increment/decrement operations)
	GetGaugeInstance(ctx context.Context, id int64) (db.GaugeInstance, error)
	UpdateGaugeInstanceValue(ctx context.Context, params db.UpdateGaugeInstanceValueParams) error
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
			// Add POST handler to support form submissions with X-HTTP-Method-Override
			r.Post("/", h.handleUpdateGauge)
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
	gaugeTemplates, err := h.queries.ListGaugeTemplates(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get gauge templates: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = layouts.Base("Admin", pages.Admin(gaugeTemplates)).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleNewGaugeForm renders the form for creating a new gauge template
func (h *GaugeHandler) handleNewGaugeForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	err := layouts.Base("New Gauge", components.GaugeTemplateForm("POST", "/admin/gauges", nil, []components.FormError{})).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// validateGaugeTemplateForm parses and validates gauge template form input from an HTTP request.
// It returns the parsed name, description, icon, unit, target value, frequency, direction, active status, and a slice of form errors.
// Frequency defaults to "weekly" and direction defaults to "under" if missing or invalid. Target is set to 0 if not a valid number.
func validateGaugeTemplateForm(r *http.Request) (string, string, string, string, int64, string, string, bool, []components.FormError) {
	var errors []components.FormError

	// Validate name
	name := r.FormValue("name")
	if name == "" {
		errors = append(errors, components.FormError{Field: "name", Message: "Name is required"})
	}

	// Get description (optional)
	description := r.FormValue("description")

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
	target, err := strconv.ParseInt(targetStr, 10, 64)
	if err != nil || targetStr == "" {
		errors = append(errors, components.FormError{Field: "target", Message: "Target must be a valid integer"})
		target = 0
	}

	// Validate frequency
	frequency := r.FormValue("frequency")
	valid := false
	for _, validFreq := range []string{"weekly", "bi-weekly", "monthly"} {
		if frequency == validFreq {
			valid = true
			break
		}
	}
	if !valid {
		frequency = "weekly" // Default to weekly if invalid
	}

	// Validate direction
	direction := r.FormValue("direction")
	if direction != "under" && direction != "over" {
		direction = "under" // Default to under if invalid
	}

	// Parse active status (checkbox)
	active := r.FormValue("active") == "on" || r.FormValue("active") == "true"

	return name, description, icon, unit, target, frequency, direction, active, errors
}

// handleCreateGauge handles the creation of a new gauge template
func (h *GaugeHandler) handleCreateGauge(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	name, description, icon, unit, target, frequency, direction, active, errors := validateGaugeTemplateForm(r)

	// If there are validation errors, re-render the form
	if len(errors) > 0 {
		w.Header().Set("Content-Type", "text/html")
		// Create a dummy gauge template to maintain form values
		dummyTemplate := &db.GaugeTemplate{
			Name:        name,
			Description: sql.NullString{String: description, Valid: description != ""},
			Icon:        icon,
			Unit:        unit,
			Target:      target,
			Frequency:   frequency,
			Direction:   direction,
			Active:      active,
		}
		err := layouts.Base("New Gauge", components.GaugeTemplateForm("POST", "/admin/gauges", dummyTemplate, errors)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Create the gauge template
	_, err := h.queries.CreateGaugeTemplate(r.Context(), db.CreateGaugeTemplateParams{
		Name:        name,
		Description: sql.NullString{String: description, Valid: description != ""},
		Icon:        icon,
		Unit:        unit,
		Target:      target,
		Frequency:   frequency,
		Direction:   direction,
		Active:      active,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create gauge template: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to admin page after successful creation
	h.handleAdmin(w, r)
}

// handleEditGaugeForm renders the form for editing an existing gauge template
func (h *GaugeHandler) handleEditGaugeForm(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid gauge template ID: %v", err), http.StatusBadRequest)
		return
	}

	// Get the gauge template
	gaugeTemplate, err := h.queries.GetGaugeTemplate(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get gauge template: %v", err), http.StatusInternalServerError)
		return
	}

	// Render the edit form
	w.Header().Set("Content-Type", "text/html")
	err = layouts.Base("Edit Gauge", components.GaugeTemplateForm("PUT", fmt.Sprintf("/admin/gauges/%d", id), &gaugeTemplate, []components.FormError{})).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleUpdateGauge handles updating an existing gauge template
func (h *GaugeHandler) handleUpdateGauge(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid gauge template ID: %v", err), http.StatusBadRequest)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	// Validate form data
	name, description, icon, unit, target, frequency, direction, active, errors := validateGaugeTemplateForm(r)

	// If there are validation errors, re-render the form
	if len(errors) > 0 {
		w.Header().Set("Content-Type", "text/html")
		// Create a gauge template with the current values to maintain form state
		currentTemplate := db.GaugeTemplate{
			ID:          id,
			Name:        name,
			Description: sql.NullString{String: description, Valid: description != ""},
			Icon:        icon,
			Unit:        unit,
			Target:      target,
			Frequency:   frequency,
			Direction:   direction,
			Active:      active,
		}
		err := layouts.Base("Edit Gauge", components.GaugeTemplateForm("PUT", fmt.Sprintf("/admin/gauges/%d", id), &currentTemplate, errors)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Update the gauge template
	err = h.queries.UpdateGaugeTemplate(r.Context(), db.UpdateGaugeTemplateParams{
		ID:          id,
		Name:        name,
		Description: sql.NullString{String: description, Valid: description != ""},
		Icon:        icon,
		Unit:        unit,
		Target:      target,
		Frequency:   frequency,
		Direction:   direction,
		Active:      active,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update gauge template: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to admin page after successful update
	h.handleAdmin(w, r)
}

// handleDeleteGauge handles the deletion of a gauge template
func (h *GaugeHandler) handleDeleteGauge(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid gauge template ID: %v", err), http.StatusBadRequest)
		return
	}

	// Delete the gauge template
	err = h.queries.DeleteGaugeTemplate(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete gauge template: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to admin page after successful deletion
	h.handleAdmin(w, r)
}

// handleIncrementGauge handles incrementing a gauge instance's value
func (h *GaugeHandler) handleIncrementGauge(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid gauge instance ID: %v", err), http.StatusBadRequest)
		return
	}

	// Get the current gauge instance
	gaugeInstance, err := h.queries.GetGaugeInstance(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get gauge instance: %v", err), http.StatusInternalServerError)
		return
	}

	// Increment the value
	err = h.queries.UpdateGaugeInstanceValue(r.Context(), db.UpdateGaugeInstanceValueParams{
		ID:    id,
		Value: gaugeInstance.Value + 1,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to increment gauge instance: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the updated gauge instance
	updatedInstance, err := h.queries.GetGaugeInstance(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get updated gauge instance: %v", err), http.StatusInternalServerError)
		return
	}

	// Render just the updated gauge value component
	w.Header().Set("Content-Type", "text/html")
	err = components.GaugeInstanceValue(&updatedInstance).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleDecrementGauge handles decrementing a gauge instance's value
func (h *GaugeHandler) handleDecrementGauge(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid gauge instance ID: %v", err), http.StatusBadRequest)
		return
	}

	// Get the current gauge instance
	gaugeInstance, err := h.queries.GetGaugeInstance(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get gauge instance: %v", err), http.StatusInternalServerError)
		return
	}

	// Only decrement if value is greater than 0
	if gaugeInstance.Value > 0 {
		err = h.queries.UpdateGaugeInstanceValue(r.Context(), db.UpdateGaugeInstanceValueParams{
			ID:    id,
			Value: gaugeInstance.Value - 1,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to decrement gauge instance: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Get the updated gauge instance
	updatedInstance, err := h.queries.GetGaugeInstance(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get updated gauge instance: %v", err), http.StatusInternalServerError)
		return
	}

	// Render just the updated gauge value component
	w.Header().Set("Content-Type", "text/html")
	err = components.GaugeInstanceValue(&updatedInstance).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
