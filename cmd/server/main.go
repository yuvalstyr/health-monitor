package main

import (
	"database/sql"
	stdlogger "log" // Standard log package for middleware compatibility
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	_ "github.com/mattn/go-sqlite3"

	"health-monitor/internal/db"
	"health-monitor/internal/handlers"
	"health-monitor/internal/views/components"
	"health-monitor/internal/views/pages"
)

// setupLogger configures zerolog with appropriate settings
func setupLogger() {
	// Set up pretty console logging for development
	if os.Getenv("ENV") != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		// In production, use JSON format
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Override global level if DEBUG env var is set
	if os.Getenv("DEBUG") == "true" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

// Create a bridge from zerolog to standard logger for middleware
type zerologBridge struct {}

func (z zerologBridge) Write(p []byte) (n int, err error) {
	log.Info().Msg(string(p))
	return len(p), nil
}

// Create a standard logger that uses zerolog as backend
var stdLog = stdlogger.New(zerologBridge{}, "", 0)

func main() {
	// Set up structured logging
	setupLogger()
	log.Info().Msg("Starting health-monitor service")

	// Use default port if not specified
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
		log.Debug().Str("port", port).Msg("Using default port")
	}

	database, err := sql.Open("sqlite3", "data/health-monitor.db")
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening database")
	}
	defer database.Close()
	log.Debug().Msg("Connected to database")

	queries := db.New(database)

	r := chi.NewRouter()

	// Custom zerolog middleware for request logging
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: stdLog, NoColor: false}))
	r.Use(middleware.Recoverer)

	// Add custom debug middleware to trace route execution
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("DEBUG: Request received: %s %s\n", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
			log.Printf("DEBUG: Request completed: %s %s\n", r.Method, r.URL.Path)
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		gauges, err := queries.ListGauges(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		err = components.Layout(pages.Dashboard(gauges)).Render(r.Context(), w)
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
	log.Info().Str("port", port).Msg("Server listening")
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}
