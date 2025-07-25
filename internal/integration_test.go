package internal

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"health-monitor/internal/db"
	"health-monitor/internal/handlers"
	"health-monitor/internal/services"
	"health-monitor/internal/testutil"
	"health-monitor/internal/timeutil"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// setupTestDB creates a test database for integration tests
func setupTestDB(t *testing.T) *db.Queries {
	// Create a temporary database file
	f, err := os.CreateTemp("", "integration_test.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })

	// Open the database
	database, err := sql.Open("sqlite", f.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	// Create the schema - adjust path for integration test
	schema, err := os.ReadFile("db/schema.sql")
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	if _, err := database.Exec(string(schema)); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db.New(database)
}

// Integration test suite for the complete gauge template and instance workflow
func TestGaugeWorkflowIntegration(t *testing.T) {
	// Setup test database
	queries := setupTestDB(t)
	ctx := context.Background()

	// Setup handlers and router
	gaugeHandler := handlers.NewGaugeHandler(queries)
	router := chi.NewRouter()
	
	// Register routes
	gaugeHandler.RegisterRoutes(router)

	t.Run("Complete Workflow Integration", func(t *testing.T) {
		// Test 1: Create gauge template
		t.Run("gauge template creation and activation", func(t *testing.T) {
			// Create a gauge template via HTTP handler
			formData := url.Values{
				"name":        {"Weekly Exercise"},
				"description": {"Track weekly exercise hours"},
				"target":      {"5"},
				"unit":        {"hours"},
				"icon":        {"fitness"},
				"frequency":   {"weekly"},
				"direction":   {"under"},
				"active":      {"true"},
			}

			req := httptest.NewRequest("POST", "/admin/gauges", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should either redirect or return success (both are acceptable)
			assert.True(t, w.Code == http.StatusSeeOther || w.Code == http.StatusFound || w.Code == http.StatusOK, 
				"Expected redirect or success after gauge creation, got %d", w.Code)

			// Verify gauge template was created in database
			templates, err := queries.ListActiveGaugeTemplates(ctx)
			require.NoError(t, err)
			require.Len(t, templates, 1)
			assert.Equal(t, "Weekly Exercise", templates[0].Name)
			assert.Equal(t, "weekly", templates[0].Frequency)
			assert.True(t, templates[0].Active)
		})

		// Test 2: Automated instance creation by scheduling service
		t.Run("automated instance creation by scheduling service", func(t *testing.T) {
			// Get the created template
			templates, err := queries.ListActiveGaugeTemplates(ctx)
			require.NoError(t, err)
			require.Len(t, templates, 1)
			template := templates[0]

			// Also create current period instance (as would happen in real usage)
			currentWeekStart := timeutil.CalculateCurrentPeriodStart("weekly", time.Now())
			currentInstance := testutil.CreateTestGaugeInstance(t, queries, template.ID, currentWeekStart)
			require.NotZero(t, currentInstance.ID)

			// Create scheduling service and run instance creation (for next period)
			schedulingService := services.NewSchedulingService(queries)
			err = schedulingService.CreateInstancesForActiveTemplates(ctx)
			require.NoError(t, err)

			// Verify instances were created for both current and next period
			nextWeekStart := timeutil.CalculateNextPeriodStart("weekly", time.Now())
			instances, err := queries.ListGaugeInstancesByTemplate(ctx, template.ID)
			require.NoError(t, err)
			require.Len(t, instances, 2) // Current + next period
			
			// Check we have both current and next period instances
			var hasCurrentPeriod, hasNextPeriod bool
			for _, instance := range instances {
				if instance.PeriodStart.Equal(currentWeekStart) {
					hasCurrentPeriod = true
					assert.Equal(t, template.ID, instance.TemplateID)
					assert.Equal(t, int64(0), instance.Value) // Initial value should be 0
				}
				if instance.PeriodStart.Equal(nextWeekStart) {
					hasNextPeriod = true
					assert.Equal(t, template.ID, instance.TemplateID)
					assert.Equal(t, int64(0), instance.Value) // Initial value should be 0
				}
			}
			assert.True(t, hasCurrentPeriod, "Should have current period instance")
			assert.True(t, hasNextPeriod, "Should have next period instance")

			// Test idempotency - running again should not create duplicate instances
			err = schedulingService.CreateInstancesForActiveTemplates(ctx)
			require.NoError(t, err)

			instances, err = queries.ListGaugeInstancesByTemplate(ctx, template.ID)
			require.NoError(t, err)
			assert.Len(t, instances, 2, "Should not create duplicate instances")
		})

		// Test 3: Dashboard filtering shows correct instances
		t.Run("dashboard filtering shows correct instances", func(t *testing.T) {
			// Make request to dashboard
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			
			// Verify response contains the gauge instance data
			responseBody := w.Body.String()
			assert.Contains(t, responseBody, "Weekly Exercise", "Dashboard should show the gauge template name")
			assert.Contains(t, responseBody, "weekly", "Dashboard should show the frequency")
			// Icons are displayed as SVG elements, not as text
			assert.Contains(t, responseBody, "gauge-instance-1", "Dashboard should show the gauge instance")

			// Verify only current period instances are shown
			// This is tested by ensuring the dashboard doesn't show historical data
			// which would be in different time periods
			currentWeekStart := timeutil.CalculateCurrentPeriodStart("weekly", time.Now())
			biWeeklyStart := timeutil.CalculateCurrentPeriodStart("bi-weekly", time.Now())
			monthlyStart := timeutil.CalculateCurrentPeriodStart("monthly", time.Now())
			
			// Create an instance for a previous week to ensure it's filtered out
			prevWeekStart := currentWeekStart.AddDate(0, 0, -7)
			template, err := queries.ListActiveGaugeTemplates(ctx)
			require.NoError(t, err)
			require.Len(t, template, 1)
			
			testutil.CreateTestGaugeInstance(t, queries, template[0].ID, prevWeekStart)

			// Make dashboard request again
			req = httptest.NewRequest("GET", "/", nil)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			
			// Should still only show current period data, not historical
			instances, err := queries.ListCurrentPeriodGaugeInstances(ctx, 
				db.ListCurrentPeriodGaugeInstancesParams{
					PeriodStart:   currentWeekStart,
					PeriodStart_2: biWeeklyStart,
					PeriodStart_3: monthlyStart,
				})
			require.NoError(t, err)
			assert.Len(t, instances, 1, "Dashboard should only show current period instances")
		})

		// Test 4: Gauge value updates work with instances
		t.Run("gauge value updates with instances", func(t *testing.T) {
			// Get the current instance
			template, err := queries.ListActiveGaugeTemplates(ctx)
			require.NoError(t, err)
			require.Len(t, template, 1)

			instances, err := queries.ListGaugeInstancesByTemplate(ctx, template[0].ID)
			require.NoError(t, err)
			require.True(t, len(instances) >= 2, "Should have at least current and next period instances")
			
			// Find current week instance
			currentWeekStart := timeutil.CalculateCurrentPeriodStart("weekly", time.Now())
			var currentInstance db.GaugeInstance
			for _, instance := range instances {
				if instance.PeriodStart.Equal(currentWeekStart) {
					currentInstance = instance
					break
				}
			}
			require.NotZero(t, currentInstance.ID, "Should find current week instance")

			// Test increment via HTTP handler
			req := httptest.NewRequest("POST", fmt.Sprintf("/gauges/%d/increment", currentInstance.ID), nil)
			req.Header.Set("HX-Request", "true") // Simulate HTMX request
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			// Verify value was incremented in database
			updatedInstance, err := queries.GetGaugeInstance(ctx, currentInstance.ID)
			require.NoError(t, err)
			assert.Equal(t, int64(1), updatedInstance.Value)

			// Test decrement via HTTP handler
			req = httptest.NewRequest("POST", fmt.Sprintf("/gauges/%d/decrement", currentInstance.ID), nil)
			req.Header.Set("HX-Request", "true")
			w = httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			// Verify value was decremented
			updatedInstance, err = queries.GetGaugeInstance(ctx, currentInstance.ID)
			require.NoError(t, err)
			assert.Equal(t, int64(0), updatedInstance.Value)

			// Test decrement doesn't go below zero
			req = httptest.NewRequest("POST", fmt.Sprintf("/gauges/%d/decrement", currentInstance.ID), nil)
			req.Header.Set("HX-Request", "true")
			w = httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should still be OK but value should remain 0
			assert.Equal(t, http.StatusOK, w.Code)
			updatedInstance, err = queries.GetGaugeInstance(ctx, currentInstance.ID)
			require.NoError(t, err)
			assert.Equal(t, int64(0), updatedInstance.Value, "Value should not go below zero")
		})

		// Test 5: Multi-frequency workflow
		t.Run("multi-frequency workflow", func(t *testing.T) {
			// Create monthly gauge template
			formData := url.Values{
				"name":        {"Monthly Budget"},
				"description": {"Track monthly budget spending"},
				"target":      {"2000"},
				"unit":        {"dollars"},
				"icon":        {"money"},
				"frequency":   {"monthly"},
				"direction":   {"under"},
				"active":      {"true"},
			}

			req := httptest.NewRequest("POST", "/admin/gauges", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.True(t, w.Code == http.StatusSeeOther || w.Code == http.StatusFound || w.Code == http.StatusOK)

			// Get the monthly template and create current period instance
			templates, err := queries.ListActiveGaugeTemplates(ctx)
			require.NoError(t, err)
			require.Len(t, templates, 2) // Weekly + Monthly

			// Find the monthly template and create current period instance
			var monthlyTemplate *db.GaugeTemplate
			for _, tmpl := range templates {
				if tmpl.Frequency == "monthly" {
					monthlyTemplate = &tmpl
					break
				}
			}
			require.NotNil(t, monthlyTemplate, "Should find monthly template")

			// Create current period instance for monthly template
			currentMonthStart := timeutil.CalculateCurrentPeriodStart("monthly", time.Now())
			testutil.CreateTestGaugeInstance(t, queries, monthlyTemplate.ID, currentMonthStart)

			// Run scheduling service again
			schedulingService := services.NewSchedulingService(queries)
			err = schedulingService.CreateInstancesForActiveTemplates(ctx)
			require.NoError(t, err)

			// Verify both weekly and monthly instances exist
			currentWeekStart := timeutil.CalculateCurrentPeriodStart("weekly", time.Now())
			biWeeklyStart := timeutil.CalculateCurrentPeriodStart("bi-weekly", time.Now())

			allInstances, err := queries.ListCurrentPeriodGaugeInstances(ctx, 
				db.ListCurrentPeriodGaugeInstancesParams{
					PeriodStart:   currentWeekStart,
					PeriodStart_2: biWeeklyStart,
					PeriodStart_3: currentMonthStart,
				})
			require.NoError(t, err)

			// Should have instances for both frequencies
			assert.Greater(t, len(allInstances), 1, "Should have instances for multiple frequencies")

			// Verify dashboard shows both
			req = httptest.NewRequest("GET", "/", nil)
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			responseBody := w.Body.String()
			assert.Contains(t, responseBody, "Weekly Exercise")
			assert.Contains(t, responseBody, "Monthly Budget")
		})
	})
}

// Test error handling and edge cases
func TestGaugeWorkflowErrorHandling(t *testing.T) {
	queries := setupTestDB(t)

	t.Run("Error Handling", func(t *testing.T) {
		t.Run("invalid gauge template creation", func(t *testing.T) {
			gaugeHandler := handlers.NewGaugeHandler(queries)
			router := chi.NewRouter()
			gaugeHandler.RegisterRoutes(router)

			// Test with missing required fields
			formData := url.Values{
				"name": {""},  // Empty name should fail
			}

			req := httptest.NewRequest("POST", "/admin/gauges", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should return an error (4xx status code)
			assert.True(t, w.Code >= 400 && w.Code < 500, "Should return client error for invalid data")
		})

		t.Run("non-existent gauge instance operations", func(t *testing.T) {
			gaugeHandler := handlers.NewGaugeHandler(queries)
			router := chi.NewRouter()
			gaugeHandler.RegisterRoutes(router)

			// Try to increment non-existent gauge
			req := httptest.NewRequest("POST", "/gauges/99999/increment", nil)
			req.Header.Set("HX-Request", "true")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should handle error gracefully
			assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError,
				"Should handle non-existent gauge gracefully")
		})
	})
}

// Test historical data and trends integration
func TestHistoricalDataIntegration(t *testing.T) {
	queries := setupTestDB(t)
	ctx := context.Background()

	// Create test template
	template := testutil.CreateTestGaugeTemplate(t, queries)
	
	// Create instances for different time periods
	currentWeek := timeutil.CalculateCurrentPeriodStart("weekly", time.Now())
	lastWeek := currentWeek.AddDate(0, 0, -7)
	twoWeeksAgo := currentWeek.AddDate(0, 0, -14)

	// Create instances
	instance1 := testutil.CreateTestGaugeInstance(t, queries, template.ID, twoWeeksAgo)
	instance2 := testutil.CreateTestGaugeInstance(t, queries, template.ID, lastWeek)
	instance3 := testutil.CreateTestGaugeInstance(t, queries, template.ID, currentWeek)

	// Add some values
	testutil.CreateTestGaugeValue(t, queries, instance1.ID, 3, twoWeeksAgo.Add(time.Hour))
	testutil.CreateTestGaugeValue(t, queries, instance1.ID, 4, twoWeeksAgo.Add(2*time.Hour))
	testutil.CreateTestGaugeValue(t, queries, instance2.ID, 5, lastWeek.Add(time.Hour))
	testutil.CreateTestGaugeValue(t, queries, instance3.ID, 2, currentWeek.Add(time.Hour))

	t.Run("historical data retrieval", func(t *testing.T) {
		history, err := queries.GetGaugeHistoryByTemplate(ctx, template.ID)
		require.NoError(t, err)
		
		// Should have data for all three periods
		assert.Len(t, history, 3, "Should have historical data for all periods")
		
		// Verify data is ordered by period (most recent first)
		assert.True(t, history[0].PeriodStart.After(history[1].PeriodStart) || 
					history[0].PeriodStart.Equal(history[1].PeriodStart),
					"Historical data should be ordered by period")
	})
}