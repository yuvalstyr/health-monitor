package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"health-monitor/internal/db"
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
			log.Printf("Error checking if instance exists for template %d: %v", template.ID, err)
			continue // Skip this template and continue with others
		}

		// If instance doesn't exist, create it
		if exists == 0 {
			_, err := s.queries.CreateGaugeInstance(ctx, db.CreateGaugeInstanceParams{
				TemplateID:  template.ID,
				PeriodStart: nextPeriodStart,
			})
			if err != nil {
				log.Printf("Error creating gauge instance for template %d: %v", template.ID, err)
				continue // Skip this template and continue with others
			}

			createdCount++
			log.Printf("Created gauge instance for template '%s' (ID: %d) for period starting %s",
				template.Name, template.ID, nextPeriodStart.Format("2006-01-02"))
		}
	}

	if createdCount > 0 {
		log.Printf("Scheduling service created %d new gauge instances", createdCount)
	}

	return nil
}