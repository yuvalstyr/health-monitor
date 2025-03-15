package models

import (
	"health-monitor/internal/db"
)

// GaugeStatus represents the status of a gauge
type GaugeStatus struct {
	Value   float64 `json:"value"`
	Target  float64 `json:"target"`
	Unit    string  `json:"unit"`
	Icon    string  `json:"icon"`
	Percent float64 `json:"percent"`
}

// MonthlyValue represents aggregated gauge values for a month
type MonthlyValue struct {
	Month        string  `json:"month"`
	AverageValue float64 `json:"average_value"`
}

// GaugeWithValue combines a gauge with its latest value
type GaugeWithValue struct {
	*db.Gauge
	Status *GaugeStatus `json:"status"`
}

// GaugeHistory represents historical data for a gauge
type GaugeHistory struct {
	*db.Gauge
	Month        string  `json:"month"`
	AverageValue float64 `json:"average_value"`
	Values       []MonthlyValue `json:"values"`
}

// NewGaugeWithValue creates a new GaugeWithValue instance
func NewGaugeWithValue(gauge *db.Gauge) *GaugeWithValue {
	percent := 0.0
	if gauge.Target > 0 {
		percent = (gauge.Value / gauge.Target) * 100
	}

	return &GaugeWithValue{
		Gauge: gauge,
		Status: &GaugeStatus{
			Value:   gauge.Value,
			Target:  gauge.Target,
			Unit:    gauge.Unit,
			Icon:    gauge.Icon,
			Percent: percent,
		},
	}
}

// NewGaugeHistory creates a new GaugeHistory instance
func NewGaugeHistory(gauge *db.Gauge, history []db.GetGaugeHistoryRow) *GaugeHistory {
	values := make([]MonthlyValue, len(history))
	for i, h := range history {
		values[i] = MonthlyValue{
			Month:        h.Month.(string),
			AverageValue: h.AverageValue,
		}
	}

	return &GaugeHistory{
		Gauge:  gauge,
		Month:  "",
		Values: values,
	}
}
