package handlers

import (
	"context"
	"fmt"
	"health-monitor/internal/db"
	"testing"

	"github.com/stretchr/testify/assert"
	"net/http/httptest"
)

func TestGaugeHandler(t *testing.T) {
	queries := &db.MockQueries{}
	handler := NewGaugeHandler(queries)

	t.Run("Create", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			queries.CreateGaugeFn = func(ctx context.Context, params db.CreateGaugeParams) (db.Gauge, error) {
				return db.Gauge{
					ID:     1,
					Name:   params.Name,
					Icon:   params.Icon,
					Unit:   params.Unit,
					Target: params.Target,
					Value:  0,
				}, nil
			}
			queries.GetAllGaugesFn = func(ctx context.Context) ([]db.Gauge, error) {
				return []db.Gauge{}, nil
			}

			r := httptest.NewRequest("POST", "/gauges", nil)
			r.Form = map[string][]string{
				"name":   {"Test Gauge"},
				"icon":   {"test-icon"},
				"unit":   {"test-unit"},
				"target": {"10"},
			}

			component, err := handler.Create(r)
			assert.NoError(t, err)
			assert.NotNil(t, component)
		})

		t.Run("validation error", func(t *testing.T) {
			r := httptest.NewRequest("POST", "/gauges", nil)
			r.Form = map[string][]string{
				"name":   {""},
				"icon":   {""},
				"unit":   {""},
				"target": {"abc"},
			}

			_, err := handler.Create(r)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid target value")
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			queries.UpdateGaugeFn = func(ctx context.Context, params db.UpdateGaugeParams) error {
				return nil
			}
			queries.GetAllGaugesFn = func(ctx context.Context) ([]db.Gauge, error) {
				return []db.Gauge{}, nil
			}

			r := httptest.NewRequest("PUT", "/gauges/1", nil)
			r.Form = map[string][]string{
				"name":   {"Updated Gauge"},
				"icon":   {"updated-icon"},
				"unit":   {"updated-unit"},
				"target": {"20"},
			}

			component, err := handler.Update(1, r)
			assert.NoError(t, err)
			assert.NotNil(t, component)
		})

		t.Run("validation error", func(t *testing.T) {
			r := httptest.NewRequest("PUT", "/gauges/1", nil)
			r.Form = map[string][]string{
				"name":   {""},
				"icon":   {""},
				"unit":   {""},
				"target": {"abc"},
			}

			_, err := handler.Update(1, r)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid target value")
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			queries.DeleteGaugeFn = func(ctx context.Context, id int64) error {
				return nil
			}
			queries.GetAllGaugesFn = func(ctx context.Context) ([]db.Gauge, error) {
				return []db.Gauge{}, nil
			}

			r := httptest.NewRequest("DELETE", "/gauges/1", nil)
			component, err := handler.Delete(r, 1)
			assert.NoError(t, err)
			assert.NotNil(t, component)
		})

		t.Run("error", func(t *testing.T) {
			queries.DeleteGaugeFn = func(ctx context.Context, id int64) error {
				return fmt.Errorf("failed to delete gauge")
			}

			r := httptest.NewRequest("DELETE", "/gauges/1", nil)
			component, err := handler.Delete(r, 1)
			assert.Error(t, err)
			assert.Nil(t, component)
		})
	})

	t.Run("Increment", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			queries.GetGaugeFn = func(ctx context.Context, id int64) (db.Gauge, error) {
				return db.Gauge{
					ID:     id,
					Name:   "Test Gauge",
					Icon:   "test-icon",
					Unit:   "test-unit",
					Target: 10,
					Value:  5,
				}, nil
			}
			queries.UpdateGaugeValueFn = func(ctx context.Context, params db.UpdateGaugeValueParams) error {
				return nil
			}

			r := httptest.NewRequest("POST", "/gauges/1/increment", nil)
			component, err := handler.Increment(r, 1)
			assert.NoError(t, err)
			assert.NotNil(t, component)
		})

		t.Run("error", func(t *testing.T) {
			queries.GetGaugeFn = func(ctx context.Context, id int64) (db.Gauge, error) {
				return db.Gauge{}, fmt.Errorf("failed to get gauge")
			}

			r := httptest.NewRequest("POST", "/gauges/1/increment", nil)
			component, err := handler.Increment(r, 1)
			assert.Error(t, err)
			assert.Nil(t, component)
		})
	})

	t.Run("Decrement", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			queries.GetGaugeFn = func(ctx context.Context, id int64) (db.Gauge, error) {
				return db.Gauge{
					ID:     id,
					Name:   "Test Gauge",
					Icon:   "test-icon",
					Unit:   "test-unit",
					Target: 10,
					Value:  5,
				}, nil
			}
			queries.UpdateGaugeValueFn = func(ctx context.Context, params db.UpdateGaugeValueParams) error {
				return nil
			}

			r := httptest.NewRequest("POST", "/gauges/1/decrement", nil)
			component, err := handler.Decrement(r, 1)
			assert.NoError(t, err)
			assert.NotNil(t, component)
		})

		t.Run("error", func(t *testing.T) {
			queries.GetGaugeFn = func(ctx context.Context, id int64) (db.Gauge, error) {
				return db.Gauge{}, fmt.Errorf("failed to get gauge")
			}

			r := httptest.NewRequest("POST", "/gauges/1/decrement", nil)
			component, err := handler.Decrement(r, 1)
			assert.Error(t, err)
			assert.Nil(t, component)
		})
	})
}
