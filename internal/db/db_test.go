package db_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"health-monitor/internal/db"
	"health-monitor/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func TestQueries_CreateAndGetGaugeTemplate(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	t.Run("create and get gauge template", func(t *testing.T) {
		params := db.CreateGaugeTemplateParams{
			Name:        "Test Gauge",
			Description: sql.NullString{String: "Test Description", Valid: true},
			Target:      100,
			Unit:        "units",
			Icon:        "star",
			Frequency:   "weekly",
			Direction:   "under",
			Active:      true,
		}

		// Create gauge template
		template, err := q.CreateGaugeTemplate(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, params.Name, template.Name)
		assert.Equal(t, params.Description, template.Description)
		assert.Equal(t, params.Target, template.Target)
		assert.Equal(t, params.Unit, template.Unit)
		assert.Equal(t, params.Icon, template.Icon)
		assert.Equal(t, params.Frequency, template.Frequency)
		assert.Equal(t, params.Active, template.Active)

		// Get gauge template
		retrieved, err := q.GetGaugeTemplate(ctx, template.ID)
		assert.NoError(t, err)
		assert.Equal(t, template, retrieved)
	})
}

func TestQueries_GaugeValues(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	t.Run("create and get gauge values", func(t *testing.T) {
		// Create gauge template and instance
		template := testutil.CreateTestGaugeTemplate(t, q)
		instance := testutil.CreateTestGaugeInstance(t, q, template.ID, time.Now().UTC())

		// Create values (all in the same month to test averaging)
		now := time.Now().UTC()
		values := []struct {
			value float64
			date  time.Time
		}{
			{50, now.AddDate(0, 0, -2)},  // 2 days ago
			{75, now.AddDate(0, 0, -1)},  // 1 day ago
			{100, now},                   // today
		}

		for _, v := range values {
			err := testutil.CreateTestGaugeValue(t, q, instance.ID, v.value, v.date)
			assert.NoError(t, err)
		}

		// Get history (should be grouped by month and averaged)
		history, err := q.GetGaugeHistory(ctx, instance.ID)
		assert.NoError(t, err)
		assert.Len(t, history, 1) // All values are in the same month

		// Average of 50, 75, 100 should be 75
		expectedAverage := (50.0 + 75.0 + 100.0) / 3.0
		assert.Equal(t, expectedAverage, history[0].AverageValue)
	})
}
