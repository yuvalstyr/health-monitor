package db

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*Queries, func()) {
	db, err := sql.Open("sqlite3", ":memory:")
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

func TestListGauges_Empty(t *testing.T) {
	q, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	gauges, err := q.ListGauges(ctx)
	require.NoError(t, err)
	require.Len(t, gauges, 0)
}

// Add more DB tests for edge cases and with inserted data
