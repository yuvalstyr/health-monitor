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
	gaugeInstance := &db.GaugeInstance{
		ID:    1,
		Value: 75.0,
	}
	
	template := &db.GaugeTemplate{
		ID:          1,
		Name:        "Test Gauge",
		Description: sql.NullString{String: "Test Description", Valid: true},
		Target:      100,
		Unit:        "units",
		Icon:        "star",
		Frequency:   "weekly",
		Direction:   "under",
		Active:      true,
	}

	t.Run("renders gauge card with basic content", func(t *testing.T) {
		component := GaugeCard(gaugeInstance, template)
		html := renderComponent(t, component)

		// Check essential content is present
		assert.Contains(t, html, template.Name)
		assert.Contains(t, html, "75")
		assert.Contains(t, html, "weekly")
		
		// Check it has increment/decrement buttons
		assert.Contains(t, html, "/gauges/1/increment")
		assert.Contains(t, html, "/gauges/1/decrement")
		
		// Check it has edit link
		assert.Contains(t, html, "/admin/gauges/1")
	})

	t.Run("renders different value when changed", func(t *testing.T) {
		differentInstance := &db.GaugeInstance{
			ID:    1,
			Value: 200,
		}
		component := GaugeCard(differentInstance, template)
		html := renderComponent(t, component)

		assert.Contains(t, html, "200")
		// Check that the value 150 is not present as a gauge value (more specific check)
		assert.NotContains(t, html, ">150<")
		assert.NotContains(t, html, "font-bold text-primary\">150<")
	})
}

func TestGauge(t *testing.T) {
	gaugeInstance := &db.GaugeInstance{
		ID:    1,
		Value: 75,
	}
	
	template := &db.GaugeTemplate{
		ID:          1,
		Name:        "Test Gauge",
		Description: sql.NullString{String: "Test Description", Valid: true},
		Target:      100,
		Unit:        "units",
		Icon:        "star",
		Frequency:   "weekly",
		Direction:   "under",
		Active:      true,
	}

	t.Run("renders gauge details", func(t *testing.T) {
		component := Gauge(gaugeInstance, template)
		html := renderComponent(t, component)

		// Check basic content
		assert.Contains(t, html, template.Name)
		assert.Contains(t, html, "75")
		assert.Contains(t, html, "weekly")

		// Check it has increment/decrement functionality
		assert.Contains(t, html, "/gauges/1/increment")
		assert.Contains(t, html, "/gauges/1/decrement")
	})

	t.Run("handles missing description", func(t *testing.T) {
		templateWithoutDesc := &db.GaugeTemplate{
			ID:          1,
			Name:        "Test Gauge",
			Description: sql.NullString{Valid: false},
			Target:      100,
			Unit:        "units",
			Icon:        "star",
			Frequency:   "weekly",
			Direction:   "under",
			Active:      true,
		}
		component := Gauge(gaugeInstance, templateWithoutDesc)
		html := renderComponent(t, component)

		// Should still render the gauge name and value
		assert.Contains(t, html, "Test Gauge")
		assert.Contains(t, html, "75")
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
