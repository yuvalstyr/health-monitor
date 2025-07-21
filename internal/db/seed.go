package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"health-monitor/internal/timeutil"
)

// SeedData contains sample data for seeding the database
type SeedData struct {
	queries *Queries
}

// NewSeedData creates a new SeedData instance
func NewSeedData(db *sql.DB) *SeedData {
	return &SeedData{
		queries: New(db),
	}
}

// SeedDatabase populates the database with sample gauge templates and instances
func (s *SeedData) SeedDatabase(ctx context.Context) error {
	log.Println("Starting database seeding...")

	// Clear existing data
	if err := s.clearExistingData(ctx); err != nil {
		return fmt.Errorf("failed to clear existing data: %w", err)
	}

	// Create sample gauge templates
	templates, err := s.createSampleTemplates(ctx)
	if err != nil {
		return fmt.Errorf("failed to create sample templates: %w", err)
	}

	// Create gauge instances for current periods
	if err := s.createCurrentPeriodInstances(ctx, templates); err != nil {
		return fmt.Errorf("failed to create current period instances: %w", err)
	}

	// Add some sample values to instances
	if err := s.addSampleValues(ctx); err != nil {
		return fmt.Errorf("failed to add sample values: %w", err)
	}

	log.Println("Database seeding completed successfully!")
	return nil
}

// clearExistingData removes all existing gauge data
func (s *SeedData) clearExistingData(ctx context.Context) error {
	log.Println("Clearing existing data...")
	
	// Delete in correct order due to foreign key constraints
	_, err := s.queries.db.ExecContext(ctx, "DELETE FROM gauge_values")
	if err != nil {
		return err
	}
	
	_, err = s.queries.db.ExecContext(ctx, "DELETE FROM gauge_instances")
	if err != nil {
		return err
	}
	
	_, err = s.queries.db.ExecContext(ctx, "DELETE FROM gauge_templates")
	if err != nil {
		return err
	}
	
	return nil
}

// createSampleTemplates creates a variety of gauge templates with different frequencies
func (s *SeedData) createSampleTemplates(ctx context.Context) ([]GaugeTemplate, error) {
	log.Println("Creating sample gauge templates...")

	sampleTemplates := []struct {
		name        string
		description string
		target      int64
		unit        string
		icon        string
		frequency   string
		direction   string
		active      bool
	}{
		// Weekly gauges
		{
			name:        "Weekly Exercise",
			description: "Track weekly workout sessions",
			target:      5,
			unit:        "sessions",
			icon:        "dumbbell",
			frequency:   "weekly",
			direction:   "over",
			active:      true,
		},
		{
			name:        "Screen Time",
			description: "Limit daily screen time hours",
			target:      35,
			unit:        "hours",
			icon:        "phone",
			frequency:   "weekly",
			direction:   "under",
			active:      true,
		},
		{
			name:        "Books Read",
			description: "Weekly reading goal",
			target:      2,
			unit:        "books",
			icon:        "book",
			frequency:   "weekly",
			direction:   "over",
			active:      true,
		},
		
		// Bi-weekly gauges
		{
			name:        "Social Activities",
			description: "Bi-weekly social engagement tracking",
			target:      4,
			unit:        "activities",
			icon:        "users",
			frequency:   "bi-weekly",
			direction:   "over",
			active:      true,
		},
		{
			name:        "Grocery Budget",
			description: "Bi-weekly grocery spending limit",
			target:      200,
			unit:        "dollars",
			icon:        "shopping-cart",
			frequency:   "bi-weekly",
			direction:   "under",
			active:      true,
		},
		{
			name:        "Deep Work Sessions",
			description: "Focused work periods bi-weekly",
			target:      10,
			unit:        "sessions",
			icon:        "brain",
			frequency:   "bi-weekly",
			direction:   "over",
			active:      false, // Inactive for testing
		},
		
		// Monthly gauges
		{
			name:        "Monthly Savings",
			description: "Monthly savings goal",
			target:      1000,
			unit:        "dollars",
			icon:        "piggy-bank",
			frequency:   "monthly",
			direction:   "over",
			active:      true,
		},
		{
			name:        "Coffee Shop Visits",
			description: "Limit monthly coffee shop expenses",
			target:      15,
			unit:        "visits",
			icon:        "coffee",
			frequency:   "monthly",
			direction:   "under",
			active:      true,
		},
		{
			name:        "Learning Hours",
			description: "Monthly skill development time",
			target:      40,
			unit:        "hours",
			icon:        "graduation-cap",
			frequency:   "monthly",
			direction:   "over",
			active:      true,
		},
		{
			name:        "Meditation Sessions",
			description: "Monthly mindfulness practice",
			target:      20,
			unit:        "sessions",
			icon:        "heart",
			frequency:   "monthly",
			direction:   "over",
			active:      false, // Inactive for testing
		},
	}

	var templates []GaugeTemplate
	for _, template := range sampleTemplates {
		created, err := s.queries.CreateGaugeTemplate(ctx, CreateGaugeTemplateParams{
			Name:        template.name,
			Description: sql.NullString{String: template.description, Valid: true},
			Target:      template.target,
			Unit:        template.unit,
			Icon:        template.icon,
			Frequency:   template.frequency,
			Direction:   template.direction,
			Active:      template.active,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create template %s: %w", template.name, err)
		}
		templates = append(templates, created)
		log.Printf("Created template: %s (%s, %s)", created.Name, created.Frequency, 
			map[bool]string{true: "active", false: "inactive"}[created.Active])
	}

	return templates, nil
}

// createCurrentPeriodInstances creates gauge instances for the current time period
func (s *SeedData) createCurrentPeriodInstances(ctx context.Context, templates []GaugeTemplate) error {
	log.Println("Creating current period gauge instances...")

	currentTime := time.Now()
	
	for _, template := range templates {
		if !template.Active {
			continue // Skip inactive templates
		}

		periodStart := timeutil.CalculateCurrentPeriodStart(template.Frequency, currentTime)
		
		// Check if instance already exists
		existsResult, err := s.queries.InstanceExistsForPeriod(ctx, InstanceExistsForPeriodParams{
			TemplateID:  template.ID,
			PeriodStart: periodStart,
		})
		if err != nil {
			return fmt.Errorf("failed to check if instance exists for template %s: %w", template.Name, err)
		}
		
		if existsResult > 0 {
			log.Printf("Instance already exists for template %s, period %s", template.Name, periodStart.Format("2006-01-02"))
			continue
		}

		// Create new instance
		instance, err := s.queries.CreateGaugeInstance(ctx, CreateGaugeInstanceParams{
			TemplateID:  template.ID,
			PeriodStart: periodStart,
		})
		if err != nil {
			return fmt.Errorf("failed to create instance for template %s: %w", template.Name, err)
		}

		log.Printf("Created instance for %s: period %s", template.Name, periodStart.Format("2006-01-02"))
		
		// Add some realistic sample progress
		if err := s.addSampleProgressToInstance(ctx, instance, template); err != nil {
			log.Printf("Warning: failed to add sample progress to instance %d: %v", instance.ID, err)
		}
	}

	return nil
}

// addSampleProgressToInstance adds realistic sample progress to a gauge instance
func (s *SeedData) addSampleProgressToInstance(ctx context.Context, instance GaugeInstance, template GaugeTemplate) error {
	// Generate realistic progress based on the gauge type and target
	var currentValue int64
	
	switch template.Direction {
	case "over":
		// For "over" targets, generate 30-80% progress
		progress := 0.3 + rand.Float64()*0.5 // 30-80%
		currentValue = int64(float64(template.Target) * progress)
	case "under":
		// For "under" targets, generate 20-70% of the limit
		progress := 0.2 + rand.Float64()*0.5 // 20-70%
		currentValue = int64(float64(template.Target) * progress)
	}

	// Update the instance value
	err := s.queries.UpdateGaugeInstanceValue(ctx, UpdateGaugeInstanceValueParams{
		Value: currentValue,
		ID:    instance.ID,
	})
	if err != nil {
		return err
	}

	// Add some historical values to show progression
	now := time.Now()
	periodStart := instance.PeriodStart
	
	// Calculate how many days into the period we are
	daysSincePeriodStart := int(now.Sub(periodStart).Hours() / 24)
	if daysSincePeriodStart < 1 {
		daysSincePeriodStart = 1
	}
	
	// Add 2-5 historical entries showing progression
	numEntries := 2 + rand.Intn(4) // 2-5 entries
	if numEntries > daysSincePeriodStart {
		numEntries = daysSincePeriodStart
	}
	
	for i := 0; i < numEntries; i++ {
		// Distribute entries across the period
		daysAgo := daysSincePeriodStart - (i * daysSincePeriodStart / numEntries)
		entryDate := now.AddDate(0, 0, -daysAgo)
		
		// Calculate progressive value (building up to current value)
		progressRatio := float64(numEntries-i) / float64(numEntries)
		entryValue := int64(float64(currentValue) * progressRatio)
		
		err := s.queries.CreateGaugeValue(ctx, CreateGaugeValueParams{
			GaugeID: instance.ID,
			Value:   entryValue,
			Date:    entryDate,
		})
		if err != nil {
			return fmt.Errorf("failed to create gauge value: %w", err)
		}
	}

	return nil
}

// addSampleValues adds some additional sample values for demonstration
func (s *SeedData) addSampleValues(ctx context.Context) error {
	log.Println("Adding sample gauge values...")
	
	// Get all current instances
	instances, err := s.queries.ListGaugeInstances(ctx)
	if err != nil {
		return err
	}

	// Add a few more recent values to some instances to show activity
	for i, instance := range instances {
		if i%3 == 0 { // Add to every third instance
			now := time.Now()
			
			// Add a value from yesterday
			yesterday := now.AddDate(0, 0, -1)
			err := s.queries.CreateGaugeValue(ctx, CreateGaugeValueParams{
				GaugeID: instance.ID,
				Value:   instance.Value + int64(rand.Intn(3)-1), // Slight variation
				Date:    yesterday,
			})
			if err != nil {
				log.Printf("Warning: failed to add yesterday value for instance %d: %v", instance.ID, err)
			}
		}
	}

	return nil
}