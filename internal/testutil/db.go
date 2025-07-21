package testutil

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"health-monitor/internal/db"

	_ "modernc.org/sqlite"
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
	database, err := sql.Open("sqlite", f.Name())
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

// CreateTestGaugeTemplate creates a test gauge template and returns it
func CreateTestGaugeTemplate(t *testing.T, q *db.Queries) db.GaugeTemplate {
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

	template, err := q.CreateGaugeTemplate(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create test gauge template: %v", err)
	}

	return template
}

// CreateTestGaugeInstance creates a test gauge instance and returns it
func CreateTestGaugeInstance(t *testing.T, q *db.Queries, templateID int64, periodStart time.Time) db.GaugeInstance {
	params := db.CreateGaugeInstanceParams{
		TemplateID:  templateID,
		PeriodStart: periodStart,
	}

	instance, err := q.CreateGaugeInstance(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create test gauge instance: %v", err)
	}

	return instance
}

// CreateTestGaugeValue creates a test gauge value
func CreateTestGaugeValue(t *testing.T, q *db.Queries, gaugeID int64, value int64, date time.Time) error {
	params := db.CreateGaugeValueParams{
		GaugeID: gaugeID,
		Value:   value,
		Date:    date,
	}

	return q.CreateGaugeValue(context.Background(), params)
}
