package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"health-monitor/internal/db"
	"health-monitor/internal/testutil"
	"health-monitor/internal/timeutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGaugeHistoryByTemplate_Weekly(t *testing.T) {
	queries := testutil.NewTestDB(t)
	handler := NewGaugeHandler(queries)

	// Create a test gauge template
	template, err := queries.CreateGaugeTemplate(context.Background(), db.CreateGaugeTemplateParams{
		Name:        "Test Weekly Gauge",
		Description: sql.NullString{String: "Test description", Valid: true},
		Target:      10,
		Unit:        "hours",
		Icon:        "clock",
		Frequency:   "weekly",
		Direction:   "under",
		Active:      true,
	})
	require.NoError(t, err)

	// Create gauge instances for different periods
	baseTime := time.Date(2024, 1, 7, 12, 0, 0, 0, time.UTC) // Sunday, Jan 7, 2024

	// Create instances for 4 consecutive weeks
	instances := make([]db.GaugeInstance, 4)
	expectedAverages := make([]float64, 4)
	
	for i := 0; i < 4; i++ {
		weekStart := timeutil.CalculateCurrentPeriodStart("weekly", baseTime.AddDate(0, 0, i*7))
		instance, err := queries.CreateGaugeInstance(context.Background(), db.CreateGaugeInstanceParams{
			TemplateID:  template.ID,
			PeriodStart: weekStart,
		})
		require.NoError(t, err)
		instances[i] = instance

		// Add some gauge values to each instance with varying patterns
		values := []int64{1, 3, 5, 7, 9}
		sum := int64(0)
		for j, val := range values {
			if j < (i+1)*2 { // Different number of values per week
				valueTime := weekStart.AddDate(0, 0, j+1)
				actualValue := val + int64(i*2) // Different base values per week
				err = queries.CreateGaugeValue(context.Background(), db.CreateGaugeValueParams{
					GaugeID: instance.ID,
					Value:   actualValue,
					Date:    valueTime,
				})
				require.NoError(t, err)
				sum += actualValue
			}
		}
		
		// Calculate expected average
		numValues := (i + 1) * 2
		if numValues > len(values) {
			numValues = len(values)
		}
		expectedAverages[i] = float64(sum) / float64(numValues)

		// Update instance value to latest
		finalValue := int64(5 + i*2 + (numValues-1)*2)
		err = queries.UpdateGaugeInstanceValue(context.Background(), db.UpdateGaugeInstanceValueParams{
			ID:    instance.ID,
			Value: finalValue,
		})
		require.NoError(t, err)
	}

	// Test getting history for the template
	history, err := handler.getGaugeHistoryByTemplate(context.Background(), template.ID)
	require.NoError(t, err)

	// Should have 4 periods of data
	assert.Len(t, history, 4)

	// Verify the data is grouped correctly by period (newest first)
	for i, period := range history {
		reverseIndex := 3 - i
		assert.Equal(t, instances[reverseIndex].PeriodStart.Format("2006-01-02"), period.PeriodStart)
		assert.Equal(t, "weekly", period.Frequency)
		assert.Equal(t, expectedAverages[reverseIndex], period.AverageValue)
		assert.Greater(t, period.ValueCount, int64(0)) // Should have some values
	}
}

func TestGetGaugeHistoryByTemplate_BiWeekly(t *testing.T) {
	queries := testutil.NewTestDB(t)
	handler := NewGaugeHandler(queries)

	// Create a bi-weekly gauge template
	template, err := queries.CreateGaugeTemplate(context.Background(), db.CreateGaugeTemplateParams{
		Name:        "Test Bi-Weekly Gauge",
		Description: sql.NullString{String: "Test description", Valid: true},
		Target:      20,
		Unit:        "tasks",
		Icon:        "check",
		Frequency:   "bi-weekly",
		Direction:   "under",
		Active:      true,
	})
	require.NoError(t, err)

	// Create instances for different bi-weekly periods
	baseTime := time.Date(2024, 1, 7, 12, 0, 0, 0, time.UTC) // Sunday, Jan 7, 2024

	// Create instances for 3 bi-weekly periods
	instances := make([]db.GaugeInstance, 3)
	expectedAverages := make([]float64, 3)
	
	for i := 0; i < 3; i++ {
		periodStart := timeutil.CalculateCurrentPeriodStart("bi-weekly", baseTime.AddDate(0, 0, i*14))
		instance, err := queries.CreateGaugeInstance(context.Background(), db.CreateGaugeInstanceParams{
			TemplateID:  template.ID,
			PeriodStart: periodStart,
		})
		require.NoError(t, err)
		instances[i] = instance

		// Add gauge values with realistic bi-weekly patterns
		values := []int64{2, 5, 8, 12, 15, 18, 20}
		sum := int64(0)
		numValues := 5 + i // Different number of values per period
		if numValues > len(values) {
			numValues = len(values)
		}
		
		for j := 0; j < numValues; j++ {
			valueTime := periodStart.AddDate(0, 0, j*2) // Every 2 days
			actualValue := values[j] + int64(i*3) // Different base values per period
			err = queries.CreateGaugeValue(context.Background(), db.CreateGaugeValueParams{
				GaugeID: instance.ID,
				Value:   actualValue,
				Date:    valueTime,
			})
			require.NoError(t, err)
			sum += actualValue
		}
		
		expectedAverages[i] = float64(sum) / float64(numValues)

		// Update instance value to latest
		finalValue := values[numValues-1] + int64(i*3)
		err = queries.UpdateGaugeInstanceValue(context.Background(), db.UpdateGaugeInstanceValueParams{
			ID:    instance.ID,
			Value: finalValue,
		})
		require.NoError(t, err)
	}

	// Test getting history
	history, err := handler.getGaugeHistoryByTemplate(context.Background(), template.ID)
	require.NoError(t, err)

	// Should have 3 periods
	assert.Len(t, history, 3)

	// Verify bi-weekly grouping (newest first)
	for i, period := range history {
		reverseIndex := 2 - i
		assert.Equal(t, instances[reverseIndex].PeriodStart.Format("2006-01-02"), period.PeriodStart)
		assert.Equal(t, "bi-weekly", period.Frequency)
		assert.Equal(t, expectedAverages[reverseIndex], period.AverageValue)
		assert.Greater(t, period.ValueCount, int64(0))
	}
}

func TestGetGaugeHistoryByTemplate_Monthly(t *testing.T) {
	queries := testutil.NewTestDB(t)
	handler := NewGaugeHandler(queries)

	// Create a monthly gauge template
	template, err := queries.CreateGaugeTemplate(context.Background(), db.CreateGaugeTemplateParams{
		Name:        "Test Monthly Gauge",
		Description: sql.NullString{String: "Test description", Valid: true},
		Target:      100,
		Unit:        "points",
		Icon:        "star",
		Frequency:   "monthly",
		Direction:   "over",
		Active:      true,
	})
	require.NoError(t, err)

	// Create instances for different months
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC) // Mid January

	instances := make([]db.GaugeInstance, 3)
	expectedAverages := make([]float64, 3)
	
	for i := 0; i < 3; i++ {
		monthTime := baseTime.AddDate(0, i, 0)
		periodStart := timeutil.CalculateCurrentPeriodStart("monthly", monthTime)
		instance, err := queries.CreateGaugeInstance(context.Background(), db.CreateGaugeInstanceParams{
			TemplateID:  template.ID,
			PeriodStart: periodStart,
		})
		require.NoError(t, err)
		instances[i] = instance

		// Add gauge values throughout the month with realistic patterns
		numValues := 8 + i*2 // Different number of values per month
		sum := int64(0)
		baseValue := int64((i + 1) * 10)
		
		for j := 0; j < numValues; j++ {
			valueTime := periodStart.AddDate(0, 0, j*3) // Every 3 days
			actualValue := baseValue + int64(j*2) // Progressive increase
			err = queries.CreateGaugeValue(context.Background(), db.CreateGaugeValueParams{
				GaugeID: instance.ID,
				Value:   actualValue,
				Date:    valueTime,
			})
			require.NoError(t, err)
			sum += actualValue
		}
		
		expectedAverages[i] = float64(sum) / float64(numValues)

		// Update instance value to latest
		finalValue := baseValue + int64((numValues-1)*2)
		err = queries.UpdateGaugeInstanceValue(context.Background(), db.UpdateGaugeInstanceValueParams{
			ID:    instance.ID,
			Value: finalValue,
		})
		require.NoError(t, err)
	}

	// Test getting history
	history, err := handler.getGaugeHistoryByTemplate(context.Background(), template.ID)
	require.NoError(t, err)

	// Should have 3 periods
	assert.Len(t, history, 3)

	// Verify monthly grouping (newest first)
	for i, period := range history {
		reverseIndex := 2 - i
		assert.Equal(t, instances[reverseIndex].PeriodStart.Format("2006-01-02"), period.PeriodStart)
		assert.Equal(t, "monthly", period.Frequency)
		assert.Equal(t, expectedAverages[reverseIndex], period.AverageValue)
		assert.Greater(t, period.ValueCount, int64(0))
	}
}

func TestGetGaugeHistoryByTemplate_EmptyHistory(t *testing.T) {
	queries := testutil.NewTestDB(t)
	handler := NewGaugeHandler(queries)

	// Create a gauge template with no instances
	template, err := queries.CreateGaugeTemplate(context.Background(), db.CreateGaugeTemplateParams{
		Name:        "Empty Gauge",
		Description: sql.NullString{String: "No data", Valid: true},
		Target:      5,
		Unit:        "items",
		Icon:        "box",
		Frequency:   "weekly",
		Direction:   "under",
		Active:      false,
	})
	require.NoError(t, err)

	// Test getting history for template with no instances
	history, err := handler.getGaugeHistoryByTemplate(context.Background(), template.ID)
	require.NoError(t, err)

	// Should return empty slice
	assert.Len(t, history, 0)
}

func TestGetGaugeHistoryByTemplate_IncompleteData(t *testing.T) {
	queries := testutil.NewTestDB(t)
	handler := NewGaugeHandler(queries)

	// Create a gauge template
	template, err := queries.CreateGaugeTemplate(context.Background(), db.CreateGaugeTemplateParams{
		Name:        "Incomplete Data Gauge",
		Description: sql.NullString{String: "Some missing data", Valid: true},
		Target:      15,
		Unit:        "hours",
		Icon:        "clock",
		Frequency:   "weekly",
		Direction:   "under",
		Active:      true,
	})
	require.NoError(t, err)

	baseTime := time.Date(2024, 1, 7, 12, 0, 0, 0, time.UTC)

	// Create instances with varying amounts of data
	instances := make([]db.GaugeInstance, 3)
	for i := 0; i < 3; i++ {
		weekStart := timeutil.CalculateCurrentPeriodStart("weekly", baseTime.AddDate(0, 0, i*7))
		instance, err := queries.CreateGaugeInstance(context.Background(), db.CreateGaugeInstanceParams{
			TemplateID:  template.ID,
			PeriodStart: weekStart,
		})
		require.NoError(t, err)
		instances[i] = instance

		// First instance: no gauge_values (only instance.value = 0)
		// Second instance: one gauge_value
		// Third instance: multiple gauge_values
		if i == 1 {
			err = queries.CreateGaugeValue(context.Background(), db.CreateGaugeValueParams{
				GaugeID: instance.ID,
				Value:   5,
				Date:    weekStart.AddDate(0, 0, 1),
			})
			require.NoError(t, err)
			err = queries.UpdateGaugeInstanceValue(context.Background(), db.UpdateGaugeInstanceValueParams{
				ID:    instance.ID,
				Value: 5,
			})
			require.NoError(t, err)
		} else if i == 2 {
			values := []int64{0, 2, 4, 6}
			for j, val := range values {
				err = queries.CreateGaugeValue(context.Background(), db.CreateGaugeValueParams{
					GaugeID: instance.ID,
					Value:   val,
					Date:    weekStart.AddDate(0, 0, j+1),
				})
				require.NoError(t, err)
			}
			err = queries.UpdateGaugeInstanceValue(context.Background(), db.UpdateGaugeInstanceValueParams{
				ID:    instance.ID,
				Value: 6,
			})
			require.NoError(t, err)
		}
	}

	// Test getting history
	history, err := handler.getGaugeHistoryByTemplate(context.Background(), template.ID)
	require.NoError(t, err)

	// Should have 3 periods
	assert.Len(t, history, 3)

	// Verify handling of incomplete data (newest first)
	// Third instance (index 2, but first in results due to reverse order)
	assert.Equal(t, 3.0, history[0].AverageValue) // Average of [0, 2, 4, 6]
	assert.Equal(t, int64(4), history[0].ValueCount) // Should have 4 values

	// Second instance
	assert.Equal(t, 5.0, history[1].AverageValue) // Single value
	assert.Equal(t, int64(1), history[1].ValueCount) // Should have 1 value

	// First instance (no gauge_values, should use instance.value)
	assert.Equal(t, 0.0, history[2].AverageValue) // Default value
	assert.Equal(t, int64(0), history[2].ValueCount) // Should have 0 values
}

func TestHandleTrends(t *testing.T) {
	queries := testutil.NewTestDB(t)
	handler := NewGaugeHandler(queries)

	// Create a gauge template
	template, err := queries.CreateGaugeTemplate(context.Background(), db.CreateGaugeTemplateParams{
		Name:        "Test Trends Gauge",
		Description: sql.NullString{String: "For trends testing", Valid: true},
		Target:      10,
		Unit:        "hours",
		Icon:        "clock",
		Frequency:   "weekly",
		Direction:   "under",
		Active:      true,
	})
	require.NoError(t, err)

	// Create some test data
	baseTime := time.Date(2024, 1, 7, 12, 0, 0, 0, time.UTC)
	weekStart := timeutil.CalculateCurrentPeriodStart("weekly", baseTime)
	instance, err := queries.CreateGaugeInstance(context.Background(), db.CreateGaugeInstanceParams{
		TemplateID:  template.ID,
		PeriodStart: weekStart,
	})
	require.NoError(t, err)

	// Add some gauge values
	err = queries.CreateGaugeValue(context.Background(), db.CreateGaugeValueParams{
		GaugeID: instance.ID,
		Value:   5,
		Date:    weekStart.AddDate(0, 0, 1),
	})
	require.NoError(t, err)

	// Set up router and test request
	r := chi.NewRouter()
	handler.RegisterTrendsRoutes(r)

	// Test trends page request
	req := httptest.NewRequest("GET", "/trends/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.handleTrends(w, req)

	// Should return 200 OK
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")

	// Should contain the gauge name in the response
	assert.Contains(t, w.Body.String(), "Test Trends Gauge")
}

func TestHandleTrends_InvalidID(t *testing.T) {
	queries := testutil.NewTestDB(t)
	handler := NewGaugeHandler(queries)

	r := chi.NewRouter()
	handler.RegisterTrendsRoutes(r)

	// Test with invalid ID
	req := httptest.NewRequest("GET", "/trends/invalid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.handleTrends(w, req)

	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleTrends_NonexistentTemplate(t *testing.T) {
	queries := testutil.NewTestDB(t)
	handler := NewGaugeHandler(queries)

	r := chi.NewRouter()
	handler.RegisterTrendsRoutes(r)

	// Test with non-existent template ID
	req := httptest.NewRequest("GET", "/trends/999", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.handleTrends(w, req)

	// Should return 500 Internal Server Error (database error)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetGaugeHistoryByTemplate_MixedFrequencies(t *testing.T) {
	queries := testutil.NewTestDB(t)
	handler := NewGaugeHandler(queries)

	// Test that each frequency type groups data correctly
	frequencies := []string{"weekly", "bi-weekly", "monthly"}
	
	for _, frequency := range frequencies {
		t.Run(fmt.Sprintf("frequency_%s", frequency), func(t *testing.T) {
			// Create a gauge template for this frequency
			template, err := queries.CreateGaugeTemplate(context.Background(), db.CreateGaugeTemplateParams{
				Name:        fmt.Sprintf("Test %s Gauge", frequency),
				Description: sql.NullString{String: "Test description", Valid: true},
				Target:      10,
				Unit:        "units",
				Icon:        "chart",
				Frequency:   frequency,
				Direction:   "under",
				Active:      true,
			})
			require.NoError(t, err)

			// Create instances based on frequency
			baseTime := time.Date(2024, 1, 7, 12, 0, 0, 0, time.UTC)
			numPeriods := 3
			var periodIncrement int
			
			switch frequency {
			case "weekly":
				periodIncrement = 7
			case "bi-weekly":
				periodIncrement = 14
			case "monthly":
				periodIncrement = 30 // Approximate for testing
			}

			instances := make([]db.GaugeInstance, numPeriods)
			for i := 0; i < numPeriods; i++ {
				var periodStart time.Time
				if frequency == "monthly" {
					// For monthly, use proper month calculation
					monthTime := baseTime.AddDate(0, i, 0)
					periodStart = timeutil.CalculateCurrentPeriodStart("monthly", monthTime)
				} else {
					periodStart = timeutil.CalculateCurrentPeriodStart(frequency, baseTime.AddDate(0, 0, i*periodIncrement))
				}
				
				instance, err := queries.CreateGaugeInstance(context.Background(), db.CreateGaugeInstanceParams{
					TemplateID:  template.ID,
					PeriodStart: periodStart,
				})
				require.NoError(t, err)
				instances[i] = instance

				// Add some test values
				for j := 0; j < 3; j++ {
					err = queries.CreateGaugeValue(context.Background(), db.CreateGaugeValueParams{
						GaugeID: instance.ID,
						Value:   int64((i+1)*2 + j),
						Date:    periodStart.AddDate(0, 0, j+1),
					})
					require.NoError(t, err)
				}

				// Update instance value
				err = queries.UpdateGaugeInstanceValue(context.Background(), db.UpdateGaugeInstanceValueParams{
					ID:    instance.ID,
					Value: int64((i+1)*2 + 2),
				})
				require.NoError(t, err)
			}

			// Test getting history
			history, err := handler.getGaugeHistoryByTemplate(context.Background(), template.ID)
			require.NoError(t, err)

			// Should have the expected number of periods
			assert.Len(t, history, numPeriods)

			// Verify all periods have the correct frequency
			for _, period := range history {
				assert.Equal(t, frequency, period.Frequency)
				assert.Greater(t, period.ValueCount, int64(0))
				assert.Greater(t, period.AverageValue, 0.0)
			}
		})
	}
}

func TestGetGaugeHistoryByTemplate_LargeDataset(t *testing.T) {
	queries := testutil.NewTestDB(t)
	handler := NewGaugeHandler(queries)

	// Create a gauge template
	template, err := queries.CreateGaugeTemplate(context.Background(), db.CreateGaugeTemplateParams{
		Name:        "Large Dataset Gauge",
		Description: sql.NullString{String: "Testing with lots of data", Valid: true},
		Target:      50,
		Unit:        "items",
		Icon:        "database",
		Frequency:   "weekly",
		Direction:   "under",
		Active:      true,
	})
	require.NoError(t, err)

	// Create many instances (simulate 6 months of weekly data)
	baseTime := time.Date(2024, 1, 7, 12, 0, 0, 0, time.UTC)
	numWeeks := 26 // 6 months
	
	for i := 0; i < numWeeks; i++ {
		weekStart := timeutil.CalculateCurrentPeriodStart("weekly", baseTime.AddDate(0, 0, i*7))
		instance, err := queries.CreateGaugeInstance(context.Background(), db.CreateGaugeInstanceParams{
			TemplateID:  template.ID,
			PeriodStart: weekStart,
		})
		require.NoError(t, err)

		// Add multiple values per week (simulate daily updates)
		for j := 0; j < 7; j++ {
			err = queries.CreateGaugeValue(context.Background(), db.CreateGaugeValueParams{
				GaugeID: instance.ID,
				Value:   int64(j + (i % 10)), // Varying values
				Date:    weekStart.AddDate(0, 0, j),
			})
			require.NoError(t, err)
		}

		// Update instance value
		err = queries.UpdateGaugeInstanceValue(context.Background(), db.UpdateGaugeInstanceValueParams{
			ID:    instance.ID,
			Value: int64(6 + (i % 10)),
		})
		require.NoError(t, err)
	}

	// Test getting history
	history, err := handler.getGaugeHistoryByTemplate(context.Background(), template.ID)
	require.NoError(t, err)

	// Should have all weeks
	assert.Len(t, history, numWeeks)

	// Verify data is ordered correctly (newest first)
	for i := 1; i < len(history); i++ {
		prevDate, err := time.Parse("2006-01-02", history[i-1].PeriodStart)
		require.NoError(t, err)
		currDate, err := time.Parse("2006-01-02", history[i].PeriodStart)
		require.NoError(t, err)
		
		assert.True(t, prevDate.After(currDate), "History should be ordered newest first")
	}

	// Verify all periods have data
	for _, period := range history {
		assert.Equal(t, "weekly", period.Frequency)
		assert.Equal(t, int64(7), period.ValueCount) // Should have 7 values per week
		assert.Greater(t, period.AverageValue, 0.0)
	}
}