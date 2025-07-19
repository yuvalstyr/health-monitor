package db

import "context"

// MockQueries is a mock implementation of the Querier interface for testing
type MockQueries struct {
	// Gauge Template methods
	CreateGaugeTemplateFn       func(ctx context.Context, params CreateGaugeTemplateParams) (GaugeTemplate, error)
	UpdateGaugeTemplateFn       func(ctx context.Context, params UpdateGaugeTemplateParams) error
	DeleteGaugeTemplateFn       func(ctx context.Context, id int64) error
	GetGaugeTemplateFn         func(ctx context.Context, id int64) (GaugeTemplate, error)
	ListGaugeTemplatesFn       func(ctx context.Context) ([]GaugeTemplate, error)
	ListActiveGaugeTemplatesFn func(ctx context.Context) ([]GaugeTemplate, error)
	
	// Gauge Instance methods
	CreateGaugeInstanceFn       func(ctx context.Context, params CreateGaugeInstanceParams) (GaugeInstance, error)
	UpdateGaugeInstanceValueFn  func(ctx context.Context, params UpdateGaugeInstanceValueParams) error
	DeleteGaugeInstanceFn       func(ctx context.Context, id int64) error
	GetGaugeInstanceFn         func(ctx context.Context, id int64) (GaugeInstance, error)
	ListGaugeInstancesFn       func(ctx context.Context) ([]GaugeInstance, error)
	ListGaugeInstancesByTemplateFn func(ctx context.Context, templateID int64) ([]GaugeInstance, error)
	InstanceExistsForPeriodFn  func(ctx context.Context, params InstanceExistsForPeriodParams) (int64, error)
	
	// Dashboard methods
	ListCurrentPeriodGaugeInstancesFn func(ctx context.Context, params ListCurrentPeriodGaugeInstancesParams) ([]ListCurrentPeriodGaugeInstancesRow, error)
	
	// Gauge Value methods
	CreateGaugeValueFn func(ctx context.Context, params CreateGaugeValueParams) error
	GetCurrentValueFn  func(ctx context.Context, gaugeID int64) (float64, error)
	GetGaugeValuesFn   func(ctx context.Context, gaugeID int64) ([]GaugeValue, error)
	GetGaugeHistoryFn  func(ctx context.Context, gaugeID int64) ([]GetGaugeHistoryRow, error)
}

// Gauge Template methods
func (m *MockQueries) CreateGaugeTemplate(ctx context.Context, params CreateGaugeTemplateParams) (GaugeTemplate, error) {
	return m.CreateGaugeTemplateFn(ctx, params)
}

func (m *MockQueries) UpdateGaugeTemplate(ctx context.Context, params UpdateGaugeTemplateParams) error {
	return m.UpdateGaugeTemplateFn(ctx, params)
}

func (m *MockQueries) DeleteGaugeTemplate(ctx context.Context, id int64) error {
	return m.DeleteGaugeTemplateFn(ctx, id)
}

func (m *MockQueries) GetGaugeTemplate(ctx context.Context, id int64) (GaugeTemplate, error) {
	return m.GetGaugeTemplateFn(ctx, id)
}

func (m *MockQueries) ListGaugeTemplates(ctx context.Context) ([]GaugeTemplate, error) {
	return m.ListGaugeTemplatesFn(ctx)
}

func (m *MockQueries) ListActiveGaugeTemplates(ctx context.Context) ([]GaugeTemplate, error) {
	return m.ListActiveGaugeTemplatesFn(ctx)
}

// Gauge Instance methods
func (m *MockQueries) CreateGaugeInstance(ctx context.Context, params CreateGaugeInstanceParams) (GaugeInstance, error) {
	return m.CreateGaugeInstanceFn(ctx, params)
}

func (m *MockQueries) UpdateGaugeInstanceValue(ctx context.Context, params UpdateGaugeInstanceValueParams) error {
	return m.UpdateGaugeInstanceValueFn(ctx, params)
}

func (m *MockQueries) DeleteGaugeInstance(ctx context.Context, id int64) error {
	return m.DeleteGaugeInstanceFn(ctx, id)
}

func (m *MockQueries) GetGaugeInstance(ctx context.Context, id int64) (GaugeInstance, error) {
	return m.GetGaugeInstanceFn(ctx, id)
}

func (m *MockQueries) ListGaugeInstances(ctx context.Context) ([]GaugeInstance, error) {
	return m.ListGaugeInstancesFn(ctx)
}

func (m *MockQueries) ListGaugeInstancesByTemplate(ctx context.Context, templateID int64) ([]GaugeInstance, error) {
	return m.ListGaugeInstancesByTemplateFn(ctx, templateID)
}

func (m *MockQueries) InstanceExistsForPeriod(ctx context.Context, params InstanceExistsForPeriodParams) (int64, error) {
	return m.InstanceExistsForPeriodFn(ctx, params)
}

// Dashboard methods
func (m *MockQueries) ListCurrentPeriodGaugeInstances(ctx context.Context, params ListCurrentPeriodGaugeInstancesParams) ([]ListCurrentPeriodGaugeInstancesRow, error) {
	return m.ListCurrentPeriodGaugeInstancesFn(ctx, params)
}

// Gauge Value methods
func (m *MockQueries) CreateGaugeValue(ctx context.Context, params CreateGaugeValueParams) error {
	return m.CreateGaugeValueFn(ctx, params)
}

func (m *MockQueries) GetCurrentValue(ctx context.Context, gaugeID int64) (float64, error) {
	return m.GetCurrentValueFn(ctx, gaugeID)
}

func (m *MockQueries) GetGaugeValues(ctx context.Context, gaugeID int64) ([]GaugeValue, error) {
	return m.GetGaugeValuesFn(ctx, gaugeID)
}

func (m *MockQueries) GetGaugeHistory(ctx context.Context, gaugeID int64) ([]GetGaugeHistoryRow, error) {
	return m.GetGaugeHistoryFn(ctx, gaugeID)
}
