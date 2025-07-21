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
	GetCurrentValueFn  func(ctx context.Context, gaugeID int64) (int64, error)
	GetGaugeValuesFn   func(ctx context.Context, gaugeID int64) ([]GaugeValue, error)
	GetGaugeHistoryFn  func(ctx context.Context, gaugeID int64) ([]GetGaugeHistoryRow, error)
}

// Gauge Template methods
func (m *MockQueries) CreateGaugeTemplate(ctx context.Context, params CreateGaugeTemplateParams) (GaugeTemplate, error) {
	if m.CreateGaugeTemplateFn == nil {
		return GaugeTemplate{}, nil
	}
	return m.CreateGaugeTemplateFn(ctx, params)
}

func (m *MockQueries) UpdateGaugeTemplate(ctx context.Context, params UpdateGaugeTemplateParams) error {
	if m.UpdateGaugeTemplateFn == nil {
		return nil
	}
	return m.UpdateGaugeTemplateFn(ctx, params)
}

func (m *MockQueries) DeleteGaugeTemplate(ctx context.Context, id int64) error {
	if m.DeleteGaugeTemplateFn == nil {
		return nil
	}
	return m.DeleteGaugeTemplateFn(ctx, id)
}

func (m *MockQueries) GetGaugeTemplate(ctx context.Context, id int64) (GaugeTemplate, error) {
	if m.GetGaugeTemplateFn == nil {
		return GaugeTemplate{}, nil
	}
	return m.GetGaugeTemplateFn(ctx, id)
}

func (m *MockQueries) ListGaugeTemplates(ctx context.Context) ([]GaugeTemplate, error) {
	if m.ListGaugeTemplatesFn == nil {
		return []GaugeTemplate{}, nil
	}
	return m.ListGaugeTemplatesFn(ctx)
}

func (m *MockQueries) ListActiveGaugeTemplates(ctx context.Context) ([]GaugeTemplate, error) {
	if m.ListActiveGaugeTemplatesFn == nil {
		return []GaugeTemplate{}, nil
	}
	return m.ListActiveGaugeTemplatesFn(ctx)
}

// Gauge Instance methods
func (m *MockQueries) CreateGaugeInstance(ctx context.Context, params CreateGaugeInstanceParams) (GaugeInstance, error) {
	if m.CreateGaugeInstanceFn == nil {
		return GaugeInstance{}, nil
	}
	return m.CreateGaugeInstanceFn(ctx, params)
}

func (m *MockQueries) UpdateGaugeInstanceValue(ctx context.Context, params UpdateGaugeInstanceValueParams) error {
	if m.UpdateGaugeInstanceValueFn == nil {
		return nil
	}
	return m.UpdateGaugeInstanceValueFn(ctx, params)
}

func (m *MockQueries) DeleteGaugeInstance(ctx context.Context, id int64) error {
	if m.DeleteGaugeInstanceFn == nil {
		return nil
	}
	return m.DeleteGaugeInstanceFn(ctx, id)
}

func (m *MockQueries) GetGaugeInstance(ctx context.Context, id int64) (GaugeInstance, error) {
	if m.GetGaugeInstanceFn == nil {
		return GaugeInstance{}, nil
	}
	return m.GetGaugeInstanceFn(ctx, id)
}

func (m *MockQueries) ListGaugeInstances(ctx context.Context) ([]GaugeInstance, error) {
	if m.ListGaugeInstancesFn == nil {
		return []GaugeInstance{}, nil
	}
	return m.ListGaugeInstancesFn(ctx)
}

func (m *MockQueries) ListGaugeInstancesByTemplate(ctx context.Context, templateID int64) ([]GaugeInstance, error) {
	if m.ListGaugeInstancesByTemplateFn == nil {
		return []GaugeInstance{}, nil
	}
	return m.ListGaugeInstancesByTemplateFn(ctx, templateID)
}

func (m *MockQueries) InstanceExistsForPeriod(ctx context.Context, params InstanceExistsForPeriodParams) (int64, error) {
	if m.InstanceExistsForPeriodFn == nil {
		return 0, nil
	}
	return m.InstanceExistsForPeriodFn(ctx, params)
}

// Dashboard methods
func (m *MockQueries) ListCurrentPeriodGaugeInstances(ctx context.Context, params ListCurrentPeriodGaugeInstancesParams) ([]ListCurrentPeriodGaugeInstancesRow, error) {
	if m.ListCurrentPeriodGaugeInstancesFn == nil {
		return []ListCurrentPeriodGaugeInstancesRow{}, nil
	}
	return m.ListCurrentPeriodGaugeInstancesFn(ctx, params)
}

// Gauge Value methods
func (m *MockQueries) CreateGaugeValue(ctx context.Context, params CreateGaugeValueParams) error {
	if m.CreateGaugeValueFn == nil {
		return nil
	}
	return m.CreateGaugeValueFn(ctx, params)
}

func (m *MockQueries) GetCurrentValue(ctx context.Context, gaugeID int64) (int64, error) {
	if m.GetCurrentValueFn == nil {
		return 0, nil
	}
	return m.GetCurrentValueFn(ctx, gaugeID)
}

func (m *MockQueries) GetGaugeValues(ctx context.Context, gaugeID int64) ([]GaugeValue, error) {
	if m.GetGaugeValuesFn == nil {
		return []GaugeValue{}, nil
	}
	return m.GetGaugeValuesFn(ctx, gaugeID)
}

func (m *MockQueries) GetGaugeHistory(ctx context.Context, gaugeID int64) ([]GetGaugeHistoryRow, error) {
	if m.GetGaugeHistoryFn == nil {
		return []GetGaugeHistoryRow{}, nil
	}
	return m.GetGaugeHistoryFn(ctx, gaugeID)
}
