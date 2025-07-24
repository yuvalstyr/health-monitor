package app

import (
	"context"
	"net/http"
	"sync"

	"health-monitor/internal/config"
	"health-monitor/internal/db"
	"health-monitor/internal/logger"
	"health-monitor/internal/scheduler"
	"health-monitor/internal/server"
	"health-monitor/internal/services"
	"health-monitor/internal/shutdown"
)

// App represents the main application
type App struct {
	config *config.Config
}

// New creates a new application instance
func New() *App {
	return &App{
		config: config.Load(),
	}
}

// Run starts the application and all its services
func (a *App) Run() error {
	// Set up structured logging
	logger.Setup()
	logger.Info().Msg("Starting health-monitor service")

	// Open database connection
	database, err := db.Open()
	if err != nil {
		logger.Fatal().Err(err).Msg("Error opening database")
		return err
	}
	defer database.Close()
	logger.Debug().Msg("Connected to database")

	queries := db.New(database)

	// Set up HTTP router
	router := server.SetupRouter(queries)
	httpServer := server.New(a.config.Port, router)

	// Create and start background scheduling service
	schedulingService := services.NewSchedulingService(queries)
	schedulerRunner := scheduler.NewDailyRunner(schedulingService)

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Start background scheduling service
	schedulerRunner.Start(ctx, &wg)

	// Start the HTTP server in a goroutine
	go func() {
		logger.Info().Str("port", a.config.Port).Msg("Server listening")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Set up shutdown manager and wait for shutdown
	shutdownManager := shutdown.New(httpServer, cancel, &wg)
	shutdownManager.WaitForShutdown()

	return nil
}