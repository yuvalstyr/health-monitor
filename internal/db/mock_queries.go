package db

import "context"

// MockQueries is a mock implementation of the Querier interface for testing
type MockQueries struct {
	CreateGaugeFn       func(ctx context.Context, params CreateGaugeParams) (Gauge, error)
	UpdateGaugeFn       func(ctx context.Context, params UpdateGaugeParams) error
	DeleteGaugeFn       func(ctx context.Context, id int64) error
	GetGaugeFn         func(ctx context.Context, id int64) (Gauge, error)
	ListGaugesFn     func(ctx context.Context) ([]Gauge, error)
	UpdateGaugeValueFn func(ctx context.Context, params UpdateGaugeValueParams) error
}

func (m *MockQueries) CreateGauge(ctx context.Context, params CreateGaugeParams) (Gauge, error) {
	return m.CreateGaugeFn(ctx, params)
}

func (m *MockQueries) UpdateGauge(ctx context.Context, params UpdateGaugeParams) error {
	return m.UpdateGaugeFn(ctx, params)
}

func (m *MockQueries) DeleteGauge(ctx context.Context, id int64) error {
	return m.DeleteGaugeFn(ctx, id)
}

func (m *MockQueries) GetGauge(ctx context.Context, id int64) (Gauge, error) {
	return m.GetGaugeFn(ctx, id)
}

func (m *MockQueries) ListGauges(ctx context.Context) ([]Gauge, error) {
	return m.ListGaugesFn(ctx)
}

func (m *MockQueries) UpdateGaugeValue(ctx context.Context, params UpdateGaugeValueParams) error {
	return m.UpdateGaugeValueFn(ctx, params)
}
