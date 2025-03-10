package models

import (
	"database/sql"
	"time"
)

// Gauge represents a health metric that we want to track
type Gauge struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Target      float64   `json:"target"`
	Unit        string    `json:"unit"`
	Icon        string    `json:"icon"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GaugeValue represents a single value reading for a gauge
type GaugeValue struct {
	ID      int64     `json:"id"`
	GaugeID int64     `json:"gauge_id"`
	Value   float64   `json:"value"`
	Date    time.Time `json:"date"`
}

// GaugeWithValue represents a gauge with its current value
type GaugeWithValue struct {
	*Gauge
	Value float64 `json:"value"`
}

// GaugeHistory represents a gauge's historical data
type GaugeHistory struct {
	Month        string  `json:"month"`
	AverageValue float64 `json:"average_value"`
}

// MonthlyValue represents a gauge's average value for a specific month
type MonthlyValue struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Value  int64  `json:"value"`
	Month  int64  `json:"month"`
	Year   int64  `json:"year"`
	Target int64  `json:"target"`
}

// WeeklyValue represents a gauge's average value for a specific week
type WeeklyValue struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Value  int64  `json:"value"`
	Week   int64  `json:"week"`
	Year   int64  `json:"year"`
}

func ToGauge(g *sql.NullString, target float64, createdAt, updatedAt sql.NullTime) *Gauge {
	return &Gauge{
		Description: g.String,
		Target:      target,
		CreatedAt:   createdAt.Time,
		UpdatedAt:   updatedAt.Time,
	}
}
