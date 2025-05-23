package main

import (
	"log"
	"net/http"
	"os"
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"

	"health-monitor/internal/db"
	"health-monitor/internal/handlers"
	"health-monitor/internal/views/components"
	"health-monitor/internal/views/pages"
)

func main() {
	// Set up more verbose logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Use default port if not specified
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	database, err := sql.Open("sqlite", "health.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Set up more verbose logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	queries := db.New(database)

	r := chi.NewRouter()

	// Add detailed logging middleware
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: log.New(os.Stdout, "HTTP: ", log.LstdFlags), NoColor: false}))
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

	log.Printf("Server starting on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}
