package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"health-monitor/internal/db"
)

func TestSchedulingService_CreateInstancesForActiveTemplates(t *testing.T) {
	ctx := context.Background()

	t.Run("creates instances for active templates without existing instances", func(t *testing.T) {
		// Setup mock queries
		mockQueries := &db.MockQueries{}
		
		// Mock active templates
		activeTemplates := []db.GaugeTemplate{
			{
				ID:        1,
				Name:      "Weekly Exercise",
				Frequency: "weekly",
				Active:    true,
			},
			{
				ID:        2,
				Name:      "Monthly Reading",
				Frequency: "monthly",
				Active:    true,
			},
		}
		
		mockQueries.ListActiveGaugeTemplatesFn = func(ctx context.Context) ([]db.GaugeTemplate, error) {
			return activeTemplates, nil
		}
		
		// Mock that no instances exist for the next periods
		mockQueries.InstanceExistsForPeriodFn = func(ctx context.Context, params db.InstanceExistsForPeriodParams) (int64, error) {
			return 0, nil // No instances exist
		}
		
		// Track created instances
		var createdInstances []db.CreateGaugeInstanceParams
		mockQueries.CreateGaugeInstanceFn = func(ctx context.Context, params db.CreateGaugeInstanceParams) (db.GaugeInstance, error) {
			createdInstances = append(createdInstances, params)
			return db.GaugeInstance{
				ID:          int64(len(createdInstances)),
				TemplateID:  params.TemplateID,
				PeriodStart: params.PeriodStart,
				Value:       0,
			}, nil
		}
		
		// Create service and run
		service := NewSchedulingService(mockQueries)
		err := service.CreateInstancesForActiveTemplates(ctx)
		
		// Verify results
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if len(createdInstances) != 2 {
			t.Fatalf("Expected 2 instances to be created, got: %d", len(createdInstances))
		}
		
		// Verify template IDs
		if createdInstances[0].TemplateID != 1 || createdInstances[1].TemplateID != 2 {
			t.Errorf("Expected template IDs 1 and 2, got: %d and %d", 
				createdInstances[0].TemplateID, createdInstances[1].TemplateID)
		}
	})

	t.Run("skips templates that already have instances for next period", func(t *testing.T) {
		mockQueries := &db.MockQueries{}
		
		// Mock active template
		activeTemplates := []db.GaugeTemplate{
			{
				ID:        1,
				Name:      "Weekly Exercise",
				Frequency: "weekly",
				Active:    true,
			},
		}
		
		mockQueries.ListActiveGaugeTemplatesFn = func(ctx context.Context) ([]db.GaugeTemplate, error) {
			return activeTemplates, nil
		}
		
		// Mock that instance already exists
		mockQueries.InstanceExistsForPeriodFn = func(ctx context.Context, params db.InstanceExistsForPeriodParams) (int64, error) {
			return 1, nil // Instance exists
		}
		
		// Track created instances
		var createdInstances []db.CreateGaugeInstanceParams
		mockQueries.CreateGaugeInstanceFn = func(ctx context.Context, params db.CreateGaugeInstanceParams) (db.GaugeInstance, error) {
			createdInstances = append(createdInstances, params)
			return db.GaugeInstance{}, nil
		}
		
		// Create service and run
		service := NewSchedulingService(mockQueries)
		err := service.CreateInstancesForActiveTemplates(ctx)
		
		// Verify results
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if len(createdInstances) != 0 {
			t.Fatalf("Expected 0 instances to be created, got: %d", len(createdInstances))
		}
	})

	t.Run("handles different frequencies correctly", func(t *testing.T) {
		mockQueries := &db.MockQueries{}
		
		// Mock templates with different frequencies
		activeTemplates := []db.GaugeTemplate{
			{
				ID:        1,
				Name:      "Weekly Exercise",
				Frequency: "weekly",
				Active:    true,
			},
			{
				ID:        2,
				Name:      "Bi-weekly Review",
				Frequency: "bi-weekly",
				Active:    true,
			},
			{
				ID:        3,
				Name:      "Monthly Reading",
				Frequency: "monthly",
				Active:    true,
			},
		}
		
		mockQueries.ListActiveGaugeTemplatesFn = func(ctx context.Context) ([]db.GaugeTemplate, error) {
			return activeTemplates, nil
		}
		
		// Mock that no instances exist
		mockQueries.InstanceExistsForPeriodFn = func(ctx context.Context, params db.InstanceExistsForPeriodParams) (int64, error) {
			return 0, nil
		}
		
		// Track created instances with their period starts
		var createdInstances []db.CreateGaugeInstanceParams
		mockQueries.CreateGaugeInstanceFn = func(ctx context.Context, params db.CreateGaugeInstanceParams) (db.GaugeInstance, error) {
			createdInstances = append(createdInstances, params)
			return db.GaugeInstance{
				ID:          int64(len(createdInstances)),
				TemplateID:  params.TemplateID,
				PeriodStart: params.PeriodStart,
			}, nil
		}
		
		// Create service and run
		service := NewSchedulingService(mockQueries)
		err := service.CreateInstancesForActiveTemplates(ctx)
		
		// Verify results
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if len(createdInstances) != 3 {
			t.Fatalf("Expected 3 instances to be created, got: %d", len(createdInstances))
		}
		
		// Verify that different period starts are calculated for different frequencies
		// We can't test exact dates without mocking time, but we can verify they're different
		weeklyStart := createdInstances[0].PeriodStart
		biWeeklyStart := createdInstances[1].PeriodStart
		monthlyStart := createdInstances[2].PeriodStart
		
		if weeklyStart.Equal(biWeeklyStart) || weeklyStart.Equal(monthlyStart) || biWeeklyStart.Equal(monthlyStart) {
			t.Error("Expected different period start dates for different frequencies")
		}
	})

	t.Run("continues processing when one template fails", func(t *testing.T) {
		mockQueries := &db.MockQueries{}
		
		// Mock active templates
		activeTemplates := []db.GaugeTemplate{
			{
				ID:        1,
				Name:      "Template 1",
				Frequency: "weekly",
				Active:    true,
			},
			{
				ID:        2,
				Name:      "Template 2",
				Frequency: "weekly",
				Active:    true,
			},
		}
		
		mockQueries.ListActiveGaugeTemplatesFn = func(ctx context.Context) ([]db.GaugeTemplate, error) {
			return activeTemplates, nil
		}
		
		// Mock that no instances exist
		mockQueries.InstanceExistsForPeriodFn = func(ctx context.Context, params db.InstanceExistsForPeriodParams) (int64, error) {
			return 0, nil
		}
		
		// Mock creation failure for first template, success for second
		var createdInstances []db.CreateGaugeInstanceParams
		mockQueries.CreateGaugeInstanceFn = func(ctx context.Context, params db.CreateGaugeInstanceParams) (db.GaugeInstance, error) {
			if params.TemplateID == 1 {
				return db.GaugeInstance{}, errors.New("database error")
			}
			createdInstances = append(createdInstances, params)
			return db.GaugeInstance{
				ID:          params.TemplateID,
				TemplateID:  params.TemplateID,
				PeriodStart: params.PeriodStart,
			}, nil
		}
		
		// Create service and run
		service := NewSchedulingService(mockQueries)
		err := service.CreateInstancesForActiveTemplates(ctx)
		
		// Verify results - should not return error but should continue processing
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		// Should have created instance for template 2 only
		if len(createdInstances) != 1 {
			t.Fatalf("Expected 1 instance to be created, got: %d", len(createdInstances))
		}
		
		if createdInstances[0].TemplateID != 2 {
			t.Errorf("Expected template ID 2, got: %d", createdInstances[0].TemplateID)
		}
	})

	t.Run("handles error when checking if instance exists", func(t *testing.T) {
		mockQueries := &db.MockQueries{}
		
		// Mock active templates
		activeTemplates := []db.GaugeTemplate{
			{
				ID:        1,
				Name:      "Template 1",
				Frequency: "weekly",
				Active:    true,
			},
			{
				ID:        2,
				Name:      "Template 2",
				Frequency: "weekly",
				Active:    true,
			},
		}
		
		mockQueries.ListActiveGaugeTemplatesFn = func(ctx context.Context) ([]db.GaugeTemplate, error) {
			return activeTemplates, nil
		}
		
		// Mock error for first template, success for second
		mockQueries.InstanceExistsForPeriodFn = func(ctx context.Context, params db.InstanceExistsForPeriodParams) (int64, error) {
			if params.TemplateID == 1 {
				return 0, errors.New("database error")
			}
			return 0, nil // No instance exists for template 2
		}
		
		// Track created instances
		var createdInstances []db.CreateGaugeInstanceParams
		mockQueries.CreateGaugeInstanceFn = func(ctx context.Context, params db.CreateGaugeInstanceParams) (db.GaugeInstance, error) {
			createdInstances = append(createdInstances, params)
			return db.GaugeInstance{
				ID:          params.TemplateID,
				TemplateID:  params.TemplateID,
				PeriodStart: params.PeriodStart,
			}, nil
		}
		
		// Create service and run
		service := NewSchedulingService(mockQueries)
		err := service.CreateInstancesForActiveTemplates(ctx)
		
		// Verify results - should not return error but should continue processing
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		// Should have created instance for template 2 only
		if len(createdInstances) != 1 {
			t.Fatalf("Expected 1 instance to be created, got: %d", len(createdInstances))
		}
		
		if createdInstances[0].TemplateID != 2 {
			t.Errorf("Expected template ID 2, got: %d", createdInstances[0].TemplateID)
		}
	})

	t.Run("returns error when listing active templates fails", func(t *testing.T) {
		mockQueries := &db.MockQueries{}
		
		// Mock error when listing templates
		mockQueries.ListActiveGaugeTemplatesFn = func(ctx context.Context) ([]db.GaugeTemplate, error) {
			return nil, errors.New("database connection error")
		}
		
		// Create service and run
		service := NewSchedulingService(mockQueries)
		err := service.CreateInstancesForActiveTemplates(ctx)
		
		// Verify error is returned
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		
		if !errors.Is(err, errors.New("database connection error")) && 
		   err.Error() != "failed to list active gauge templates: database connection error" {
			t.Errorf("Expected wrapped database error, got: %v", err)
		}
	})

	t.Run("handles empty list of active templates", func(t *testing.T) {
		mockQueries := &db.MockQueries{}
		
		// Mock empty list of templates
		mockQueries.ListActiveGaugeTemplatesFn = func(ctx context.Context) ([]db.GaugeTemplate, error) {
			return []db.GaugeTemplate{}, nil
		}
		
		// Create service and run
		service := NewSchedulingService(mockQueries)
		err := service.CreateInstancesForActiveTemplates(ctx)
		
		// Verify no error and no instances created
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	t.Run("creates instances with correct default values", func(t *testing.T) {
		mockQueries := &db.MockQueries{}
		
		// Mock active template
		activeTemplates := []db.GaugeTemplate{
			{
				ID:        1,
				Name:      "Test Template",
				Frequency: "weekly",
				Active:    true,
			},
		}
		
		mockQueries.ListActiveGaugeTemplatesFn = func(ctx context.Context) ([]db.GaugeTemplate, error) {
			return activeTemplates, nil
		}
		
		mockQueries.InstanceExistsForPeriodFn = func(ctx context.Context, params db.InstanceExistsForPeriodParams) (int64, error) {
			return 0, nil
		}
		
		// Track created instances
		var createdInstances []db.CreateGaugeInstanceParams
		mockQueries.CreateGaugeInstanceFn = func(ctx context.Context, params db.CreateGaugeInstanceParams) (db.GaugeInstance, error) {
			createdInstances = append(createdInstances, params)
			return db.GaugeInstance{
				ID:          1,
				TemplateID:  params.TemplateID,
				PeriodStart: params.PeriodStart,
				Value:       0, // Should start with 0 value
				CreatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
				UpdatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
			}, nil
		}
		
		// Create service and run
		service := NewSchedulingService(mockQueries)
		err := service.CreateInstancesForActiveTemplates(ctx)
		
		// Verify results
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if len(createdInstances) != 1 {
			t.Fatalf("Expected 1 instance to be created, got: %d", len(createdInstances))
		}
		
		// Verify the instance was created with correct template ID
		if createdInstances[0].TemplateID != 1 {
			t.Errorf("Expected template ID 1, got: %d", createdInstances[0].TemplateID)
		}
		
		// Verify period start is in the future (next period)
		now := time.Now()
		if createdInstances[0].PeriodStart.Before(now) {
			t.Error("Expected period start to be in the future")
		}
	})
}