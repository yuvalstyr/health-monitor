package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"health-monitor/internal/db"
	"health-monitor/internal/handlers"
	"health-monitor/internal/logger"
)

// SetupRouter creates and configures the HTTP router with all routes and middleware
func SetupRouter(queries db.Querier) *chi.Mux {
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

	// Create gauge handler and register all gauge-related routes
	gaugeHandler := handlers.NewGaugeHandler(queries)
	gaugeHandler.RegisterRoutes(r)

	// Add static file server for assets
	fs := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	return r
}

// New creates a new HTTP server with the given configuration
func New(port string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}
}