package handlers

import (
	"context"
	"io"
	"net/http"
	"health-monitor/internal/charts"
	"health-monitor/internal/views/components"
	"health-monitor/internal/views/layouts"
	"github.com/a-h/templ"
)

// ChartTestHandler provides endpoints for testing chart rendering with various data sets
type ChartTestHandler struct{}

func NewChartTestHandler() *ChartTestHandler {
	return &ChartTestHandler{}
}

// HandleChartTest renders a test page with various chart configurations
func (h *ChartTestHandler) HandleChartTest(w http.ResponseWriter, r *http.Request) {
	// Test data set 1: Weekly data
	weeklyData := charts.ChartData{
		Labels: []string{"Week 1", "Week 2", "Week 3", "Week 4", "Week 5", "Week 6"},
		Values: []float64{2.5, 3.2, 1.8, 4.1, 3.7, 2.9},
		Target: 3.0,
		Unit:   "hours",
		Title:  "Weekly Exercise",
	}

	// Test data set 2: Monthly data
	monthlyData := charts.ChartData{
		Labels: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"},
		Values: []float64{85, 92, 78, 88, 95, 82},
		Target: 90,
		Unit:   "points",
		Title:  "Monthly Health Score",
	}

	// Test data set 3: Bi-weekly data with higher frequency
	biWeeklyData := charts.ChartData{
		Labels: []string{"Period 1", "Period 2", "Period 3", "Period 4", "Period 5", "Period 6", "Period 7", "Period 8"},
		Values: []float64{12, 15, 8, 18, 22, 16, 14, 20},
		Target: 15,
		Unit:   "sessions",
		Title:  "Bi-weekly Training Sessions",
	}

	// Test data set 4: Empty data
	emptyData := charts.ChartData{
		Labels: []string{},
		Values: []float64{},
		Target: 10,
		Unit:   "units",
		Title:  "Empty Dataset",
	}

	// Test data set 5: Single data point
	singleData := charts.ChartData{
		Labels: []string{"Current"},
		Values: []float64{7.5},
		Target: 8.0,
		Unit:   "kg",
		Title:  "Single Data Point",
	}

	content := chartTestContent(weeklyData, monthlyData, biWeeklyData, emptyData, singleData)
	layouts.Base("Chart Test", content).Render(r.Context(), w)
}

// chartTestContent renders the test page content with multiple charts
func chartTestContent(weekly, monthly, biWeekly, empty, single charts.ChartData) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(`
			<div class="container mx-auto px-4 py-8">
				<h1 class="text-3xl font-bold mb-8">ApexCharts Integration Test</h1>
				<p class="text-base-content/70 mb-8">Testing chart rendering with various data sets and frequencies</p>
				
				<div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
					<!-- Weekly Data Test -->
					<div class="card bg-base-100 shadow-xl">
						<div class="card-body">
							<h2 class="card-title">Weekly Data Test</h2>
							<p class="text-sm text-base-content/70 mb-4">6 data points, target line at 3.0</p>
		`))
		if err != nil {
			return err
		}

		// Render weekly chart
		err = components.LineChart(weekly, "weeklyChart").Render(ctx, w)
		if err != nil {
			return err
		}

		_, err = w.Write([]byte(`
						</div>
					</div>

					<!-- Monthly Data Test -->
					<div class="card bg-base-100 shadow-xl">
						<div class="card-body">
							<h2 class="card-title">Monthly Data Test</h2>
							<p class="text-sm text-base-content/70 mb-4">6 data points, target line at 90</p>
		`))
		if err != nil {
			return err
		}

		// Render monthly chart
		err = components.LineChart(monthly, "monthlyChart").Render(ctx, w)
		if err != nil {
			return err
		}

		_, err = w.Write([]byte(`
						</div>
					</div>

					<!-- Bi-weekly Data Test -->
					<div class="card bg-base-100 shadow-xl">
						<div class="card-body">
							<h2 class="card-title">Bi-weekly Data Test</h2>
							<p class="text-sm text-base-content/70 mb-4">8 data points, target line at 15</p>
		`))
		if err != nil {
			return err
		}

		// Render bi-weekly chart
		err = components.LineChart(biWeekly, "biWeeklyChart").Render(ctx, w)
		if err != nil {
			return err
		}

		_, err = w.Write([]byte(`
						</div>
					</div>

					<!-- Empty Data Test -->
					<div class="card bg-base-100 shadow-xl">
						<div class="card-body">
							<h2 class="card-title">Empty Data Test</h2>
							<p class="text-sm text-base-content/70 mb-4">No data points - should show empty state</p>
		`))
		if err != nil {
			return err
		}

		// Render empty chart
		if len(empty.Values) == 0 {
			err = components.EmptyChart("No data available for this test").Render(ctx, w)
		} else {
			err = components.LineChart(empty, "emptyChart").Render(ctx, w)
		}
		if err != nil {
			return err
		}

		_, err = w.Write([]byte(`
						</div>
					</div>

					<!-- Single Data Point Test -->
					<div class="card bg-base-100 shadow-xl">
						<div class="card-body">
							<h2 class="card-title">Single Data Point Test</h2>
							<p class="text-sm text-base-content/70 mb-4">1 data point, target line at 8.0</p>
		`))
		if err != nil {
			return err
		}

		// Render single data chart
		err = components.LineChart(single, "singleChart").Render(ctx, w)
		if err != nil {
			return err
		}

		_, err = w.Write([]byte(`
						</div>
					</div>

					<!-- HTMX Update Test -->
					<div class="card bg-base-100 shadow-xl">
						<div class="card-body">
							<h2 class="card-title">HTMX Update Test</h2>
							<p class="text-sm text-base-content/70 mb-4">Chart updates every 10 seconds</p>
		`))
		if err != nil {
			return err
		}

		// Render HTMX chart
		err = components.ChartWithHTMX(weekly, "htmxChart", "/test/chart-data", "10s").Render(ctx, w)
		if err != nil {
			return err
		}

		_, err = w.Write([]byte(`
						</div>
					</div>
				</div>

				<div class="mt-8">
					<div class="alert alert-info">
						<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
						<div>
							<h3 class="font-bold">Test Results</h3>
							<div class="text-xs">
								<p>✓ Charts should render with ApexCharts library</p>
								<p>✓ Target lines should be visible as dashed red lines</p>
								<p>✓ Empty data should show placeholder message</p>
								<p>✓ Single data points should render correctly</p>
								<p>✓ HTMX chart should update automatically</p>
							</div>
						</div>
					</div>
				</div>
			</div>
		`))
		return err
	})
}

// HandleChartData provides dynamic data for HTMX chart updates
func (h *ChartTestHandler) HandleChartData(w http.ResponseWriter, r *http.Request) {
	// Generate random-ish data for testing HTMX updates
	dynamicData := charts.ChartData{
		Labels: []string{"Week 1", "Week 2", "Week 3", "Week 4", "Week 5", "Week 6"},
		Values: []float64{2.1, 3.8, 2.2, 4.5, 3.1, 3.6}, // Different values each time
		Target: 3.0,
		Unit:   "hours",
		Title:  "Dynamic Weekly Exercise",
	}

	// Return just the chart component for HTMX to swap
	components.LineChart(dynamicData, "htmxChart").Render(r.Context(), w)
}

// HandleChartUpdate provides dynamic chart updates for HTMX testing (alias for HandleChartData)
func (h *ChartTestHandler) HandleChartUpdate(w http.ResponseWriter, r *http.Request) {
	h.HandleChartData(w, r)
}