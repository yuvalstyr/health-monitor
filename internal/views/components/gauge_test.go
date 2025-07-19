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
		gauge.Value = 75.0
		component := GaugeCard(gauge)
		html := renderComponent(t, component)

		// Check basic content
		assert.Contains(t, html, gauge.Name)
		assert.Contains(t, html, gauge.Description.String)
		assert.Contains(t, html, gauge.Unit)
		assert.Contains(t, html, "75.0")
		assert.Contains(t, html, "Target: 100.0")

		// Check card structure
		assert.Contains(t, html, `class="card bg-base-100 shadow-xl hover:shadow-2xl transition-all group"`)
		assert.Contains(t, html, `class="card-body p-3 sm:p-6"`)
		assert.Contains(t, html, `class="card-title text-base sm:text-lg mb-0 sm:mb-1"`)

		// Check value display
		assert.Contains(t, html, `text-4xl sm:text-5xl font-bold transition-all`)
		assert.Contains(t, html, `badge badge-lg badge-outline`)

		// Check progress bar
		assert.Contains(t, html, `w-full h-2.5 sm:h-3 bg-base-200/50 rounded-lg overflow-hidden shadow-inner`)

		// Check controls
		assert.Contains(t, html, `btn btn-error btn-sm w-full font-bold`)
		assert.Contains(t, html, `btn btn-success btn-sm w-full font-bold`)
	})

	t.Run("shows warning when over target", func(t *testing.T) {
		gauge.Value = 150.0
		component := GaugeCard(gauge)
		html := renderComponent(t, component)

		assert.Contains(t, html, `text-error animate-pulse`)
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

		// Check HTMX attributes and paths
		assert.Contains(t, html, `hx-post="/gauges/1/increment"`)
		assert.Contains(t, html, `hx-post="/gauges/1/decrement"`)
		assert.Contains(t, html, `hx-target="#gauge-1"`)
		assert.Contains(t, html, `hx-swap="outerHTML"`)
		assert.Contains(t, html, `hx-push-url="false"`)
		assert.Contains(t, html, `href="/admin/gauges/1"`)
		assert.Contains(t, html, `hx-delete="/admin/gauges/1"`)
		assert.Contains(t, html, `hx-confirm="Are you sure you want to delete this gauge?"`)


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
