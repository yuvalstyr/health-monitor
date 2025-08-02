package services

import (
	"context"
	"fmt"
	"time"

	"health-monitor/internal/db"
	"health-monitor/internal/logger"
	"health-monitor/internal/timeutil"
)

// SchedulingService handles automated gauge instance creation
type SchedulingService interface {
	CreateInstancesForActiveTemplates(ctx context.Context) error
}

// schedulingService implements the SchedulingService interface
type schedulingService struct {
	queries db.Querier
}

// NewSchedulingService creates a new scheduling service
func NewSchedulingService(queries db.Querier) SchedulingService {
	return &schedulingService{
		queries: queries,
	}
}

// CreateInstancesForActiveTemplates creates gauge instances for the next period
// for all active gauge templates that don't already have instances for that period
func (s *schedulingService) CreateInstancesForActiveTemplates(ctx context.Context) error {
	// Get all active gauge templates
	templates, err := s.queries.ListActiveGaugeTemplates(ctx)
	if err != nil {
		return fmt.Errorf("failed to list active gauge templates: %w", err)
	}

	currentTime := time.Now()
	var createdCount int

	// Process each active template
	for _, template := range templates {
		// Calculate the next period start date for this template
		nextPeriodStart := timeutil.CalculateNextPeriodStart(template.Frequency, currentTime)

		// Check if an instance already exists for this period
		exists, err := s.queries.InstanceExistsForPeriod(ctx, db.InstanceExistsForPeriodParams{
			TemplateID:  template.ID,
			PeriodStart: nextPeriodStart,
		})
		if err != nil {
			logger.Error().Err(err).Int64("template_id", template.ID).Msg("Error checking if instance exists for template")
			continue // Skip this template and continue with others
		}

		// If instance doesn't exist, create it
		if exists == 0 {
			_, err := s.queries.CreateGaugeInstance(ctx, db.CreateGaugeInstanceParams{
				TemplateID:  template.ID,
				PeriodStart: nextPeriodStart,
			})
			if err != nil {
				logger.Error().Err(err).Int64("template_id", template.ID).Msg("Error creating gauge instance for template")
				continue // Skip this template and continue with others
			}

			createdCount++
			logger.Info().
				Str("template_name", template.Name).
				Int64("template_id", template.ID).
				Str("period_start", nextPeriodStart.Format("2006-01-02")).
				Msg("Created gauge instance for template")
		}
	}

	if createdCount > 0 {
		logger.Info().Int("created_count", createdCount).Msg("Scheduling service created new gauge instances")
	}

	return nil
}