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
			name: "calculate percentage - target > 0",
			gauge: &db.Gauge{
				ID:     1,
				Name:   "Test Gauge",
				Value:  80,
				Target: 100,
				Unit:   "units",
				Icon:   "star",
			},
			expected: &GaugeWithValue{
				Gauge: &db.Gauge{
					ID:     1,
					Name:   "Test Gauge",
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
			name: "zero target - no percentage",
			gauge: &db.Gauge{
				ID:     2,
				Name:   "Zero Target",
				Value:  80,
				Target: 0,
				Unit:   "units",
				Icon:   "warning",
			},
			expected: &GaugeWithValue{
				Gauge: &db.Gauge{
					ID:     2,
					Name:   "Zero Target",
					Value:  80,
					Target: 0,
					Unit:   "units",
					Icon:   "warning",
				},
				Status: &GaugeStatus{
					Value:   80,
					Target:  0,
					Unit:    "units",
					Icon:    "warning",
					Percent: 0,
				},
			},
		},
		{
			name: "zero value - zero percentage",
			gauge: &db.Gauge{
				ID:     3,
				Name:   "Zero Value",
				Value:  0,
				Target: 100,
				Unit:   "units",
				Icon:   "error",
			},
			expected: &GaugeWithValue{
				Gauge: &db.Gauge{
					ID:     3,
					Name:   "Zero Value",
					Value:  0,
					Target: 100,
					Unit:   "units",
					Icon:   "error",
				},
				Status: &GaugeStatus{
					Value:   0,
					Target:  100,
					Unit:    "units",
					Icon:    "error",
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
		ID:     1,
		Name:   "Test Gauge",
		Value:  80,
		Target: 100,
		Unit:   "units",
		Icon:   "star",
	}

	tests := []struct {
		name     string
		gauge    *db.Gauge
		history  []db.GetGaugeHistoryRow
		expected *GaugeHistory
	}{
		{
			name:  "multiple history entries",
			gauge: gauge,
			history: []db.GetGaugeHistoryRow{
				{
					Month:        "2025-03",
					AverageValue: 75.5,
				},
				{
					Month:        "2025-02",
					AverageValue: 82.3,
				},
			},
			expected: &GaugeHistory{
				Gauge: gauge,
				Month: "",
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
			},
		},
		{
			name:    "empty history",
			gauge:   gauge,
			history: []db.GetGaugeHistoryRow{},
			expected: &GaugeHistory{
				Gauge:  gauge,
				Month:  "",
				Values: []MonthlyValue{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGaugeHistory(tt.gauge, tt.history)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestMonthTypeAssertion(t *testing.T) {
	gauge := &db.Gauge{
		ID:     1,
		Name:   "Test Gauge",
		Value:  80,
		Target: 100,
		Unit:   "units",
		Icon:   "star",
	}

	// Test with string month
	history := []db.GetGaugeHistoryRow{
		{
			Month:        "2025-03",
			AverageValue: 75.5,
		},
	}

	result := NewGaugeHistory(gauge, history)
	assert.Equal(t, "2025-03", result.Values[0].Month)

	// Test with nil month
	historyNil := []db.GetGaugeHistoryRow{
		{
			Month:        nil,
			AverageValue: 82.3,
		},
	}

	resultNil := NewGaugeHistory(gauge, historyNil)
	assert.Equal(t, "", resultNil.Values[0].Month)
}
