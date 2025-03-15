package testutil

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"health-monitor/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

// NewTestDB creates a new test database and returns a Queries instance
func NewTestDB(t *testing.T) *db.Queries {
	// Create a temporary database file
	f, err := os.CreateTemp("", "test.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })

	// Open the database
	database, err := sql.Open("sqlite3", f.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	// Create the schema
	schema, err := os.ReadFile("../db/schema.sql")
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	if _, err := database.Exec(string(schema)); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db.New(database)
}

// CreateTestGauge creates a test gauge and returns it
func CreateTestGauge(t *testing.T, q *db.Queries) db.Gauge {
	params := db.CreateGaugeParams{
		Name:        "Test Gauge",
		Description: sql.NullString{String: "Test Description", Valid: true},
		Target:      100,
		Unit:        "units",
		Icon:        "star",
	}

	gauge, err := q.CreateGauge(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create test gauge: %v", err)
	}

	return gauge
}

// CreateTestGaugeValue creates a test gauge value
func CreateTestGaugeValue(t *testing.T, q *db.Queries, gaugeID int64, value float64) {
	params := db.CreateGaugeValueParams{
		GaugeID: gaugeID,
		Column2: value,
		Date:    time.Now(),
	}

	err := q.CreateGaugeValue(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create test gauge value: %v", err)
	}
}
