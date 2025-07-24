package shutdown

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"health-monitor/internal/logger"
)

// Manager handles graceful shutdown of the application
type Manager struct {
	server *http.Server
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

// New creates a new shutdown manager
func New(server *http.Server, cancel context.CancelFunc, wg *sync.WaitGroup) *Manager {
	return &Manager{
		server: server,
		cancel: cancel,
		wg:     wg,
	}
}

// WaitForShutdown waits for shutdown signals and handles graceful shutdown
func (m *Manager) WaitForShutdown() {
	// Channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal
	<-sigChan
	logger.Info().Msg("Shutdown signal received, stopping services...")

	// Cancel context to stop background services
	m.cancel()

	// Shutdown HTTP server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := m.server.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("Server shutdown error")
	}

	// Wait for background services to finish
	m.wg.Wait()
	logger.Info().Msg("All services stopped gracefully")
}