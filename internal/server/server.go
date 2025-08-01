package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"health-monitor/internal/db"
	"health-monitor/internal/handlers"
	"health-monitor/internal/logger"
)

// SetupRouter creates and configures the HTTP router with all routes and middleware
func SetupRouter(queries db.Querier, database *sql.DB, version string) *chi.Mux {
	r := chi.NewRouter()

	// Custom zerolog middleware for request logging
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: logger.StdLogger(), NoColor: false}))
	r.Use(middleware.Recoverer)

	// Add custom debug middleware to trace route execution
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Debug().Str("method", r.Method).Str("path", r.URL.Path).Msg("Request received")
			next.ServeHTTP(w, r)
			logger.Debug().Str("method", r.Method).Str("path", r.URL.Path).Msg("Request completed")
		})
	})

	// Create and register health check handler
	healthHandler := handlers.NewHealthHandler(database, version)
	r.Get("/health", healthHandler.HealthCheck)

	// Create migration management handler and register routes
	migrationHandler := handlers.NewMigrationHandler(database)
	migrationHandler.RegisterRoutes(r)

	// Create gauge handler and register all gauge-related routes
	gaugeHandler := handlers.NewGaugeHandler(queries)
	gaugeHandler.RegisterRoutes(r)
	gaugeHandler.RegisterTrendsRoutes(r)

	// Add static file server for assets with directory listing protection
	fs := http.FileServer(neuteredFileSystem{http.Dir("./static")})
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	return r
}

// neuteredFileSystem wraps http.FileSystem to prevent directory listings
type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if s.IsDir() {
		return nil, os.ErrNotExist
	}

	return f, nil
}

// New creates a new HTTP server with the given configuration
func New(port string, handler http.Handler) *http.Server {
	// Validate port
	if port == "" {
		panic("port cannot be empty")
	}

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 1 || portNum > 65535 {
		panic(fmt.Sprintf("invalid port number: %s (must be between 1 and 65535)", port))
	}

	return &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: handler,
	}
}