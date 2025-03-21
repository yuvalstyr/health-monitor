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

func TestQueries_CreateAndGetGauge(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	t.Run("create and get gauge", func(t *testing.T) {
		params := db.CreateGaugeParams{
			Name:        "Test Gauge",
			Description: sql.NullString{String: "Test Description", Valid: true},
			Target:      100,
			Unit:        "units",
			Icon:        "star",
		}

		// Create gauge
		gauge, err := q.CreateGauge(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, params.Name, gauge.Name)
		assert.Equal(t, params.Description, gauge.Description)
		assert.Equal(t, params.Target, gauge.Target)
		assert.Equal(t, params.Unit, gauge.Unit)
		assert.Equal(t, params.Icon, gauge.Icon)

		// Get gauge
		retrieved, err := q.GetGauge(ctx, gauge.ID)
		assert.NoError(t, err)
		assert.Equal(t, gauge, retrieved)
	})
}

func TestQueries_GaugeValues(t *testing.T) {
	q := testutil.NewTestDB(t)
	ctx := context.Background()

	t.Run("create and get gauge values", func(t *testing.T) {
		// Create gauge
		gauge := testutil.CreateTestGauge(t, q)

		// Create values
		now := time.Now().UTC()
		values := []struct {
			value float64
			date  time.Time
		}{
			{50, now.AddDate(0, -2, 0)},
			{75, now.AddDate(0, -1, 0)},
			{100, now},
		}

		for _, v := range values {
			err := testutil.CreateTestGaugeValue(t, q, gauge.ID, v.value, v.date)
			assert.NoError(t, err)
		}

		// Get history
		history, err := q.GetGaugeHistory(ctx, gauge.ID)
		assert.NoError(t, err)
		assert.Len(t, history, 3)

		// Values should be ordered by date DESC (newest first)
		assert.Equal(t, 100.0, history[0].AverageValue)
		assert.Equal(t, 75.0, history[1].AverageValue)
		assert.Equal(t, 50.0, history[2].AverageValue)
	})
}
