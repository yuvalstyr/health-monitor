package scheduler

import (
	"context"
	"sync"
	"time"

	"health-monitor/internal/logger"
	"health-monitor/internal/services"
)

// Runner manages the background execution of scheduling services
type Runner struct {
	schedulingService services.SchedulingService
	interval          time.Duration
}

// NewRunner creates a new scheduler runner with the specified service and interval
func NewRunner(schedulingService services.SchedulingService, interval time.Duration) *Runner {
	return &Runner{
		schedulingService: schedulingService,
		interval:          interval,
	}
}

// NewDailyRunner creates a new scheduler runner with daily (24-hour) intervals
func NewDailyRunner(schedulingService services.SchedulingService) *Runner {
	return NewRunner(schedulingService, 24*time.Hour)
}

// Start runs the scheduling service in the background with the configured interval
func (r *Runner) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		r.run(ctx)
	}()
}

// run is the internal method that handles the scheduling loop
func (r *Runner) run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	logger.Info().Msg("Background scheduling service started")

	// Run once immediately on startup
	if err := r.schedulingService.CreateInstancesForActiveTemplates(ctx); err != nil {
		logger.Error().Err(err).Msg("Error running initial scheduling service")
	}

	// Then run on schedule
	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Background scheduling service stopped")
			return
		case <-ticker.C:
			logger.Debug().Msg("Running scheduled gauge instance creation")
			if err := r.schedulingService.CreateInstancesForActiveTemplates(ctx); err != nil {
				logger.Error().Err(err).Msg("Error running scheduled gauge instance creation")
			}
		}
	}
}