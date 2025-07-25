package models

import (
	"testing"

	"health-monitor/internal/db"

	"github.com/stretchr/testify/assert"
)

func TestNewGaugeTemplateWithStatus(t *testing.T) {
	tests := []struct {
		name         string
		template     *db.GaugeTemplate
		currentValue int64
		expected     *GaugeTemplateWithStatus
	}{
		{
			name: "calculate percentage - target > 0",
			template: &db.GaugeTemplate{
				ID:     1,
				Name:   "Test Template",
				Target: 100,
				Unit:   "units",
				Icon:   "star",
			},
			currentValue: 80,
			expected: &GaugeTemplateWithStatus{
				GaugeTemplate: &db.GaugeTemplate{
					ID:     1,
					Name:   "Test Template",
					Target: 100,
					Unit:   "units",
					Icon:   "star",
				},
				Status: &GaugeStatus{
					Value:   80,
					Target:  100,
					Unit:    "units",
					Icon:    "star",
					Percent: 80,
				},
			},
		},
		{
			name: "zero target - no percentage",
			template: &db.GaugeTemplate{
				ID:     2,
				Name:   "Zero Target",
				Target: 0,
				Unit:   "units",
				Icon:   "warning",
			},
			currentValue: 80,
			expected: &GaugeTemplateWithStatus{
				GaugeTemplate: &db.GaugeTemplate{
					ID:     2,
					Name:   "Zero Target",
					Target: 0,
					Unit:   "units",
					Icon:   "warning",
				},
				Status: &GaugeStatus{
					Value:   80,
					Target:  0,
					Unit:    "units",
					Icon:    "warning",
					Percent: 0,
				},
			},
		},
		{
			name:         "nil template",
			template:     nil,
			currentValue: 80,
			expected:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGaugeTemplateWithStatus(tt.template, tt.currentValue)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestNewGaugeHistory(t *testing.T) {
	tests := []struct {
		name         string
		instanceID   int64
		templateName string
		history      []db.GetGaugeHistoryRow
		expected     *GaugeHistory
	}{
		{
			name:         "multiple history entries",
			instanceID:   1,
			templateName: "Test Template",
			history: []db.GetGaugeHistoryRow{
				{
					Month:        "2025-03",
					AverageValue: 75,
				},
				{
					Month:        "2025-02",
					AverageValue: 82,
				},
			},
			expected: &GaugeHistory{
				InstanceID:   1,
				TemplateName: "Test Template",
				Month:        "",
				Values: []MonthlyValue{
					{
						Month:        "2025-03",
						AverageValue: 75,
					},
					{
						Month:        "2025-02",
						AverageValue: 82,
					},
				},
			},
		},
		{
			name:         "empty history",
			instanceID:   2,
			templateName: "Empty Template",
			history:      []db.GetGaugeHistoryRow{},
			expected: &GaugeHistory{
				InstanceID:   2,
				TemplateName: "Empty Template",
				Month:        "",
				Values:       []MonthlyValue{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGaugeHistory(tt.instanceID, tt.templateName, tt.history)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestMonthTypeAssertion(t *testing.T) {
	// Test with string month
	history := []db.GetGaugeHistoryRow{
		{
			Month:        "2025-03",
			AverageValue: 75,
		},
	}

	result := NewGaugeHistory(1, "Test Template", history)
	assert.Equal(t, "2025-03", result.Values[0].Month)

	// Test with nil month
	historyNil := []db.GetGaugeHistoryRow{
		{
			Month:        nil,
			AverageValue: 82,
		},
	}

	resultNil := NewGaugeHistory(2, "Test Template", historyNil)
	assert.Equal(t, "", resultNil.Values[0].Month)
}

func TestNewGaugeInstanceWithStatus(t *testing.T) {
	tests := []struct {
		name     string
		instance *db.ListCurrentPeriodGaugeInstancesRow
		expected *GaugeInstanceWithStatus
	}{
		{
			name: "calculate percentage - target > 0",
			instance: &db.ListCurrentPeriodGaugeInstancesRow{
				ID:     1,
				Name:   "Test Instance",
				Value:  80,
				Target: 100,
				Unit:   "units",
				Icon:   "star",
			},
			expected: &GaugeInstanceWithStatus{
				ListCurrentPeriodGaugeInstancesRow: &db.ListCurrentPeriodGaugeInstancesRow{
					ID:     1,
					Name:   "Test Instance",
					Value:  80,
					Target: 100,
					Unit:   "units",
					Icon:   "star",
				},
				Status: &GaugeStatus{
					Value:   80,
					Target:  100,
					Unit:    "units",
					Icon:    "star",
					Percent: 80,
				},
			},
		},
		{
			name: "zero target - no percentage",
			instance: &db.ListCurrentPeriodGaugeInstancesRow{
				ID:     2,
				Name:   "Zero Target",
				Value:  80,
				Target: 0,
				Unit:   "units",
				Icon:   "warning",
			},
			expected: &GaugeInstanceWithStatus{
				ListCurrentPeriodGaugeInstancesRow: &db.ListCurrentPeriodGaugeInstancesRow{
					ID:     2,
					Name:   "Zero Target",
					Value:  80,
					Target: 0,
					Unit:   "units",
					Icon:   "warning",
				},
				Status: &GaugeStatus{
					Value:   80,
					Target:  0,
					Unit:    "units",
					Icon:    "warning",
					Percent: 0,
				},
			},
		},
		{
			name:     "nil instance",
			instance: nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGaugeInstanceWithStatus(tt.instance)
			assert.Equal(t, tt.expected, got)
		})
	}
}
