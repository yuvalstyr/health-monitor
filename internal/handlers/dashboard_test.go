package handlers

import (
	"context"
	"database/sql"
	"health-monitor/internal/db"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockQuerier implements the Querier interface for testing
type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) ListGaugeTemplates(ctx context.Context) ([]db.GaugeTemplate, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.GaugeTemplate), args.Error(1)
}

func (m *MockQuerier) GetGaugeTemplate(ctx context.Context, id int64) (db.GaugeTemplate, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.GaugeTemplate), args.Error(1)
}

func (m *MockQuerier) CreateGaugeTemplate(ctx context.Context, params db.CreateGaugeTemplateParams) (db.GaugeTemplate, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.GaugeTemplate), args.Error(1)
}

func (m *MockQuerier) UpdateGaugeTemplate(ctx context.Context, params db.UpdateGaugeTemplateParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockQuerier) DeleteGaugeTemplate(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQuerier) GetGaugeInstance(ctx context.Context, id int64) (db.GaugeInstance, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.GaugeInstance), args.Error(1)
}

func (m *MockQuerier) UpdateGaugeInstanceValue(ctx context.Context, params db.UpdateGaugeInstanceValueParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockQuerier) CreateGaugeValue(ctx context.Context, params db.CreateGaugeValueParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockQuerier) ListCurrentPeriodGaugeInstances(ctx context.Context, params db.ListCurrentPeriodGaugeInstancesParams) ([]db.ListCurrentPeriodGaugeInstancesRow, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]db.ListCurrentPeriodGaugeInstancesRow), args.Error(1)
}

func TestHandleDashboard_WithActiveGauges(t *testing.T) {
	// Create mock querier
	mockQuerier := new(MockQuerier)
	handler := NewGaugeHandler(mockQuerier)

	// Create test data - current period gauge instances
	testInstances := []db.ListCurrentPeriodGaugeInstancesRow{
		{
			ID:          1,
			TemplateID:  1,
			PeriodStart: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), // Sunday
			Value:       5,
			CreatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
			UpdatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
			Name:        "Weekly Exercise",
			Description: sql.NullString{String: "Track weekly exercise", Valid: true},
			Target:      10,
			Unit:        "hours",
			Icon:        "chart-bar",
			Frequency:   "weekly",
			Direction:   "under",
		},
		{
			ID:          2,
			TemplateID:  2,
			PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // First of month
			Value:       3,
			CreatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
			UpdatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
			Name:        "Monthly Reading",
			Description: sql.NullString{String: "Track monthly reading", Valid: true},
			Target:      5,
			Unit:        "books",
			Icon:        "book",
			Frequency:   "monthly",
			Direction:   "under",
		},
	}

	// Set up mock expectations
	mockQuerier.On("ListCurrentPeriodGaugeInstances", mock.Anything, mock.AnythingOfType("db.ListCurrentPeriodGaugeInstancesParams")).Return(testInstances, nil)

	// Create test request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.handleDashboard(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	
	// Verify the response contains gauge data
	body := w.Body.String()
	assert.Contains(t, body, "Weekly Exercise")
	assert.Contains(t, body, "Monthly Reading")
	assert.NotContains(t, body, "No Active Gauges")

	// Verify mock was called
	mockQuerier.AssertExpectations(t)
}

func TestHandleDashboard_NoActiveGauges(t *testing.T) {
	// Create mock querier
	mockQuerier := new(MockQuerier)
	handler := NewGaugeHandler(mockQuerier)

	// Set up mock to return empty slice
	emptyInstances := []db.ListCurrentPeriodGaugeInstancesRow{}
	mockQuerier.On("ListCurrentPeriodGaugeInstances", mock.Anything, mock.AnythingOfType("db.ListCurrentPeriodGaugeInstancesParams")).Return(emptyInstances, nil)

	// Create test request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.handleDashboard(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	
	// Verify the response shows "no active gauges" message
	body := w.Body.String()
	assert.Contains(t, body, "No Active Gauges")
	assert.Contains(t, body, "No gauge instances are active for the current time period")
	assert.Contains(t, body, "Create Gauge Templates")

	// Verify mock was called
	mockQuerier.AssertExpectations(t)
}

func TestHandleDashboard_DatabaseError(t *testing.T) {
	// Create mock querier
	mockQuerier := new(MockQuerier)
	handler := NewGaugeHandler(mockQuerier)

	// Set up mock to return error
	mockQuerier.On("ListCurrentPeriodGaugeInstances", mock.Anything, mock.AnythingOfType("db.ListCurrentPeriodGaugeInstancesParams")).Return([]db.ListCurrentPeriodGaugeInstancesRow{}, assert.AnError)

	// Create test request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.handleDashboard(w, req)

	// Verify response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to fetch current period gauge instances")

	// Verify mock was called
	mockQuerier.AssertExpectations(t)
}

func TestHandleDashboard_FilteringLogic(t *testing.T) {
	// Create mock querier
	mockQuerier := new(MockQuerier)
	handler := NewGaugeHandler(mockQuerier)

	// Create test data with different frequencies
	testInstances := []db.ListCurrentPeriodGaugeInstancesRow{
		{
			ID:          1,
			TemplateID:  1,
			PeriodStart: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), // Sunday - weekly
			Value:       2,
			Name:        "Weekly Task",
			Frequency:   "weekly",
			Target:      5,
			Unit:        "tasks",
			Icon:        "check",
			Direction:   "under",
		},
		{
			ID:          2,
			TemplateID:  2,
			PeriodStart: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), // Sunday - bi-weekly
			Value:       4,
			Name:        "Bi-weekly Goal",
			Frequency:   "bi-weekly",
			Target:      8,
			Unit:        "goals",
			Icon:        "target",
			Direction:   "under",
		},
		{
			ID:          3,
			TemplateID:  3,
			PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // First of month - monthly
			Value:       1,
			Name:        "Monthly Habit",
			Frequency:   "monthly",
			Target:      3,
			Unit:        "habits",
			Icon:        "calendar",
			Direction:   "under",
		},
	}

	// Set up mock expectations
	mockQuerier.On("ListCurrentPeriodGaugeInstances", mock.Anything, mock.MatchedBy(func(params db.ListCurrentPeriodGaugeInstancesParams) bool {
		// Verify that the parameters contain proper period start dates
		// The exact dates will depend on when the test runs, but we can verify the structure
		return !params.PeriodStart.IsZero() && !params.PeriodStart_2.IsZero() && !params.PeriodStart_3.IsZero()
	})).Return(testInstances, nil)

	// Create test request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.handleDashboard(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify all frequency types are represented
	body := w.Body.String()
	assert.Contains(t, body, "Weekly Task")
	assert.Contains(t, body, "Bi-weekly Goal")
	assert.Contains(t, body, "Monthly Habit")

	// Verify mock was called
	mockQuerier.AssertExpectations(t)
}