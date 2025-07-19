package db

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*Queries, func()) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Load schema
	schema, err := os.ReadFile("schema.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(schema))
	require.NoError(t, err)

	q := New(db)
	cleanup := func() { db.Close() }
	return q, cleanup
}

func TestListGaugeTemplates_Empty(t *testing.T) {
	q, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	templates, err := q.ListGaugeTemplates(ctx)
	require.NoError(t, err)
	require.Len(t, templates, 0)
}

func TestCreateGaugeTemplate(t *testing.T) {
	q, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	
	params := CreateGaugeTemplateParams{
		Name:        "Test Gauge",
		Description: sql.NullString{String: "Test Description", Valid: true},
		Target:      10.0,
		Unit:        "hours",
		Icon:        "chart-bar",
		Frequency:   "weekly",
		Direction:   "under",
		Active:      true,
	}
	
	template, err := q.CreateGaugeTemplate(ctx, params)
	require.NoError(t, err)
	require.Equal(t, "Test Gauge", template.Name)
	require.Equal(t, "weekly", template.Frequency)
	require.True(t, template.Active)
}

func TestListActiveGaugeTemplates(t *testing.T) {
	q, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	
	// Create an active template
	activeParams := CreateGaugeTemplateParams{
		Name:        "Active Gauge",
		Description: sql.NullString{String: "Active Description", Valid: true},
		Target:      5.0,
		Unit:        "times",
		Icon:        "chart-bar",
		Frequency:   "monthly",
		Direction:   "under",
		Active:      true,
	}
	
	// Create an inactive template
	inactiveParams := CreateGaugeTemplateParams{
		Name:        "Inactive Gauge",
		Description: sql.NullString{String: "Inactive Description", Valid: true},
		Target:      3.0,
		Unit:        "days",
		Icon:        "chart-bar",
		Frequency:   "weekly",
		Direction:   "under",
		Active:      false,
	}
	
	_, err := q.CreateGaugeTemplate(ctx, activeParams)
	require.NoError(t, err)
	
	_, err = q.CreateGaugeTemplate(ctx, inactiveParams)
	require.NoError(t, err)
	
	// List only active templates
	activeTemplates, err := q.ListActiveGaugeTemplates(ctx)
	require.NoError(t, err)
	require.Len(t, activeTemplates, 1)
	require.Equal(t, "Active Gauge", activeTemplates[0].Name)
	require.True(t, activeTemplates[0].Active)
}

// Add more DB tests for edge cases and with inserted data
