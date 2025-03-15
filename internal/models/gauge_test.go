package models

import (
	"testing"

	"health-monitor/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestNewGaugeWithValue(t *testing.T) {
	tests := []struct {
		name     string
		gauge    *db.Gauge
		expected *GaugeWithValue
	}{
		{
			name: "gauge with target",
			gauge: &db.Gauge{
				Value:  80,
				Target: 100,
				Unit:   "units",
				Icon:   "star",
			},
			expected: &GaugeWithValue{
				Gauge: &db.Gauge{
					Value:  80,
					Target: 100,
					Unit:   "units",
					Icon:   "star",
				},
				Status: &GaugeStatus{
					Value:   80,
					Target:  100,
					Unit:    "units",
					Icon:    "star",
					Percent: 80,
				},
			},
		},
		{
			name: "gauge with zero target",
			gauge: &db.Gauge{
				Value:  80,
				Target: 0,
				Unit:   "units",
				Icon:   "star",
			},
			expected: &GaugeWithValue{
				Gauge: &db.Gauge{
					Value:  80,
					Target: 0,
					Unit:   "units",
					Icon:   "star",
				},
				Status: &GaugeStatus{
					Value:   80,
					Target:  0,
					Unit:    "units",
					Icon:    "star",
					Percent: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGaugeWithValue(tt.gauge)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestNewGaugeHistory(t *testing.T) {
	gauge := &db.Gauge{
		Value:  80,
		Target: 100,
		Unit:   "units",
		Icon:   "star",
	}

	history := []db.GetGaugeHistoryRow{
		{
			Month:        "2025-03",
			AverageValue: 75.5,
		},
		{
			Month:        "2025-02",
			AverageValue: 82.3,
		},
	}

	expected := &GaugeHistory{
		Gauge: gauge,
		Values: []MonthlyValue{
			{
				Month:        "2025-03",
				AverageValue: 75.5,
			},
			{
				Month:        "2025-02",
				AverageValue: 82.3,
			},
		},
	}

	got := NewGaugeHistory(gauge, history)
	assert.Equal(t, expected, got)
}
