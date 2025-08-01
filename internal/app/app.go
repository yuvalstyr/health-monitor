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

	// Open database connection with configured path
	database, err := db.Open(a.config.DBPath)
	if err != nil {
		logger.Error().Err(err).Msg("Error opening database")
		return err
	}
	defer database.Close()
	logger.Debug().Str("db_path", a.config.DBPath).Msg("Connected to database")

	queries := db.New(database)

	// Set up HTTP router with health check
	router := server.SetupRouter(queries, database, a.config.Version)
	httpServer := server.New(a.config.Port, router)

	// Create and start background scheduling service
	schedulingService := services.NewSchedulingService(queries)
	schedulerRunner := scheduler.NewDailyRunner(schedulingService)

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Create error channel for server startup errors
	serverErrChan := make(chan error, 1)

	// Start background scheduling service
	schedulerRunner.Start(ctx, &wg)

	// Start the HTTP server in a goroutine
	go func() {
		logger.Info().Str("port", a.config.Port).Msg("Server listening")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("Server failed to start")
			serverErrChan <- err
		}
	}()

	// Set up shutdown manager
	shutdownManager := shutdown.New(httpServer, cancel, &wg)

	// Wait for either shutdown signal or server error
	select {
	case err := <-serverErrChan:
		logger.Error().Err(err).Msg("Server startup failed, initiating shutdown")
		cancel() // Cancel context to stop other services
		shutdownManager.WaitForShutdown()
		return err
	case <-ctx.Done():
		// Normal shutdown initiated
		shutdownManager.WaitForShutdown()
		return nil
	}
}