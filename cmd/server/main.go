package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"health-monitor/internal/db"
	"health-monitor/internal/handlers"
	"health-monitor/internal/logger"
	"health-monitor/internal/views/layouts"
	"health-monitor/internal/views/pages"
)

// main starts the health-monitor web service, initializing logging, database connections, HTTP routing, middleware, and serving HTTP requests.
func main() {
	// Set up structured logging
	logger.Setup()
	logger.Info().Msg("Starting health-monitor service")

	// Use default port if not specified
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
		logger.Debug().Str("port", port).Msg("Using default port")
	}

	database, err := db.Open()
	if err != nil {
		logger.Fatal().Err(err).Msg("Error opening database")
	}
	defer database.Close()
	logger.Debug().Msg("Connected to database")

	queries := db.New(database)

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

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// For now, show empty dashboard - this will be updated in task 7 to show current period instances
		gaugeInstances := []db.ListCurrentPeriodGaugeInstancesRow{}

		w.Header().Set("Content-Type", "text/html")
		err = layouts.Base("Dashboard", pages.Dashboard(gaugeInstances)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Create gauge handler and register all gauge-related routes
	gaugeHandler := handlers.NewGaugeHandler(queries)
	gaugeHandler.RegisterRoutes(r)

	// Add static file server for assets
	fs := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	// Start the HTTP server
	logger.Info().Str("port", port).Msg("Server listening")
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		logger.Fatal().Err(err).Msg("Server failed to start")
	}
}
