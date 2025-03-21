package components

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"health-monitor/internal/db"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
)

func TestGaugeCard(t *testing.T) {
	gauge := &db.Gauge{
		ID:          1,
		Name:        "Test Gauge",
		Description: sql.NullString{String: "Test Description", Valid: true},
		Target:      100,
		Unit:        "units",
		Icon:        "star",
	}

	t.Run("renders gauge card", func(t *testing.T) {
		component := GaugeCard(gauge, 75.0)
		html := renderComponent(t, component)

		// Check basic content
		assert.Contains(t, html, gauge.Name)
		assert.Contains(t, html, gauge.Description.String)
		assert.Contains(t, html, gauge.Unit)
		assert.Contains(t, html, "75.0")
		assert.Contains(t, html, "100.0")

		// Check HTMX attributes for dynamic updates
		assert.Contains(t, html, `hx-post="/gauges/1/increment"`)
		assert.Contains(t, html, `hx-post="/gauges/1/decrement"`)
		assert.Contains(t, html, `hx-target="#gauge-1"`)
		assert.Contains(t, html, `hx-swap="outerHTML"`)
		assert.Contains(t, html, `hx-push-url="false"`)
		assert.Contains(t, html, `hx-indicator="#loading-1"`)

		// Check responsive classes
		assert.Contains(t, html, `class="card w-full bg-base-100 shadow-xl"`)
		assert.Contains(t, html, `class="card-body"`)
		assert.Contains(t, html, `class="card-title"`)
		assert.Contains(t, html, `class="text-base-content/70"`)
		assert.Contains(t, html, `class="btn btn-circle btn-md sm:btn-lg"`)
		assert.Contains(t, html, `class="stats shadow"`)
		assert.Contains(t, html, `class="stat hidden sm:inline"`)

		// Check layout and structure
		assert.Contains(t, html, `class="flex items-center justify-between"`)
		assert.Contains(t, html, `class="flex items-center justify-center gap-4 mt-4"`)
	})

	t.Run("shows warning when over target", func(t *testing.T) {
		component := GaugeCard(gauge, 150.0)
		html := renderComponent(t, component)

		assert.Contains(t, html, `class="stat-value text-error animate-pulse"`)
	})
}

func TestGauge(t *testing.T) {
	gauge := &db.Gauge{
		ID:          1,
		Name:        "Test Gauge",
		Description: sql.NullString{String: "Test Description", Valid: true},
		Target:      100,
		Value:       75,
		Unit:        "units",
		Icon:        "star",
	}

	t.Run("renders gauge details", func(t *testing.T) {
		component := Gauge(gauge)
		html := renderComponent(t, component)

		// Check basic content
		assert.Contains(t, html, gauge.Name)
		assert.Contains(t, html, gauge.Description.String)
		assert.Contains(t, html, gauge.Unit)
		assert.Contains(t, html, "75.0")
		assert.Contains(t, html, "100.0")

		// Check HTMX form attributes
		assert.Contains(t, html, `hx-boost="true"`)
		assert.Contains(t, html, `hx-target="#gauge-1"`)
		assert.Contains(t, html, `hx-swap="outerHTML"`)

		// Check action buttons with proper paths
		assert.Contains(t, html, `action="/gauges/1/increment"`)
		assert.Contains(t, html, `action="/gauges/1/decrement"`)
		assert.Contains(t, html, `href="/admin/gauges/1"`)
		assert.Contains(t, html, `action="/admin/gauges/1"`)

		// Check stats component
		assert.Contains(t, html, `class="stats shadow mb-4"`)
		assert.Contains(t, html, `class="stat-title"`)
		assert.Contains(t, html, `class="stat-value"`)
		assert.Contains(t, html, `class="stat-desc"`)

		// Check responsive design classes
		assert.Contains(t, html, `class="text-lg sm:text-xl font-bold"`)
		assert.Contains(t, html, `class="text-sm sm:text-base mb-4"`)
		assert.Contains(t, html, `id="gauge-1" class="flex flex-col sm:flex-row gap-4"`)
	})

	t.Run("handles missing description", func(t *testing.T) {
		gauge := &db.Gauge{
			ID:          1,
			Name:        "Test Gauge",
			Description: sql.NullString{Valid: false},
			Target:      100,
			Value:       75,
			Unit:        "units",
			Icon:        "star",
		}
		component := Gauge(gauge)
		html := renderComponent(t, component)

		assert.NotContains(t, html, `text-base-content/70`)
	})
}

// Helper function to render a component to HTML string
func renderComponent(t *testing.T, component templ.Component) string {
	t.Helper()
	var sb strings.Builder
	err := component.Render(context.Background(), &sb)
	assert.NoError(t, err)
	return sb.String()
}
