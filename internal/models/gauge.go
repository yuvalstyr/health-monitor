package models

import (
	"health-monitor/internal/db"
)

// GaugeStatus represents the status of a gauge
type GaugeStatus struct {
	Value   float64 `json:"value"`
	Target  float64 `json:"target"`
	Unit    string  `json:"unit"`
	Icon    string  `json:"icon"`
	Percent float64 `json:"percent"`
}

// MonthlyValue represents aggregated gauge values for a month
type MonthlyValue struct {
	Month        string  `json:"month"`
	AverageValue float64 `json:"average_value"`
}

// GaugeInstanceWithStatus combines a gauge instance with its status
type GaugeInstanceWithStatus struct {
	*db.ListCurrentPeriodGaugeInstancesRow
	Status *GaugeStatus `json:"status"`
}

// GaugeTemplateWithStatus combines a gauge template with its status
type GaugeTemplateWithStatus struct {
	*db.GaugeTemplate
	Status *GaugeStatus `json:"status"`
}

// GaugeHistory represents historical data for a gauge instance
type GaugeHistory struct {
	InstanceID   int64          `json:"instance_id"`
	TemplateName string         `json:"template_name"`
	Month        string         `json:"month"`
	AverageValue float64        `json:"average_value"`
	Values       []MonthlyValue `json:"values"`
}

// NewGaugeInstanceWithStatus creates a new GaugeInstanceWithStatus instance
func NewGaugeInstanceWithStatus(instance *db.ListCurrentPeriodGaugeInstancesRow) *GaugeInstanceWithStatus {
	if instance == nil {
		return nil
	}
	
	percent := 0.0
	if instance.Target > 0 {
		percent = (instance.Value / instance.Target) * 100
	}

	return &GaugeInstanceWithStatus{
		ListCurrentPeriodGaugeInstancesRow: instance,
		Status: &GaugeStatus{
			Value:   instance.Value,
			Target:  instance.Target,
			Unit:    instance.Unit,
			Icon:    instance.Icon,
			Percent: percent,
		},
	}
}

// NewGaugeTemplateWithStatus creates a new GaugeTemplateWithStatus instance
func NewGaugeTemplateWithStatus(template *db.GaugeTemplate, currentValue float64) *GaugeTemplateWithStatus {
	if template == nil {
		return nil
	}
	
	percent := 0.0
	if template.Target > 0 {
		percent = (currentValue / template.Target) * 100
	}

	return &GaugeTemplateWithStatus{
		GaugeTemplate: template,
		Status: &GaugeStatus{
			Value:   currentValue,
			Target:  template.Target,
			Unit:    template.Unit,
			Icon:    template.Icon,
			Percent: percent,
		},
	}
}

// NewGaugeHistory creates a new GaugeHistory instance
func NewGaugeHistory(instanceID int64, templateName string, history []db.GetGaugeHistoryRow) *GaugeHistory {
	values := make([]MonthlyValue, len(history))
	for i, h := range history {
		month := ""
		if h.Month != nil {
			if str, ok := h.Month.(string); ok {
				month = str
			}
		}
		values[i] = MonthlyValue{
			Month:        month,
			AverageValue: h.AverageValue,
		}
	}

	return &GaugeHistory{
		InstanceID:   instanceID,
		TemplateName: templateName,
		Month:        "",
		Values:       values,
	}
}
