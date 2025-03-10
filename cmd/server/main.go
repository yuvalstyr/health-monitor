package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"health-monitor/internal/db"
	"health-monitor/internal/models"
	"health-monitor/internal/views/components"
	"health-monitor/internal/views/pages"
)

//go:embed web/static
var staticFiles embed.FS

func toGauge(g db.Gauge) *models.Gauge {
	return &models.Gauge{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description.String,
		Target:      g.Target,
		Unit:        g.Unit,
		CreatedAt:   g.CreatedAt.Time,
		UpdatedAt:   g.UpdatedAt.Time,
	}
}

func getIncrementValue(unit string) float64 {
	unit = strings.ToLower(unit)
	switch {
	case strings.Contains(unit, "step"):
		return 1000 // Increment steps by 1000
	case strings.Contains(unit, "liter"):
		return 0.1 // Increment water by 0.1 liters
	case strings.Contains(unit, "hour"):
		return 0.5 // Increment sleep by 0.5 hours
	default:
		return 1.0
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "health.db"
	}

	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Run database migrations
	if err := db.Migrate(database); err != nil {
		log.Fatal(err)
	}

	queries := db.New(database)

	r := chi.NewRouter()

	// Static file server
	r.Handle("/static/*", http.FileServer(http.FS(staticFiles)))

	// Dashboard routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		gauges, err := queries.ListGauges(r.Context())
		if err != nil {
			log.Printf("Error listing gauges: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		gaugesWithValues := make([]models.GaugeWithValue, len(gauges))
		for i, g := range gauges {
			value, err := queries.GetCurrentValue(r.Context(), g.ID)
			if err != nil {
				log.Printf("Error getting current value for gauge %d: %v", g.ID, err)
				value = 0
			}
			gaugesWithValues[i] = models.GaugeWithValue{
				Gauge: toGauge(g),
				Value: value,
			}
		}

		err = pages.DashboardPage(gaugesWithValues).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering dashboard page: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Admin routes
	r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		gauges, err := queries.ListGauges(r.Context())
		if err != nil {
			log.Printf("Error listing gauges: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		gaugesWithValues := make([]models.GaugeWithValue, len(gauges))
		for i, g := range gauges {
			value, err := queries.GetCurrentValue(r.Context(), g.ID)
			if err != nil {
				log.Printf("Error getting current value for gauge %d: %v", g.ID, err)
				value = 0
			}
			gaugesWithValues[i] = models.GaugeWithValue{
				Gauge: toGauge(g),
				Value: value,
			}
		}

		err = pages.AdminPage(gaugesWithValues).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering admin page: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/admin/gauges/new", func(w http.ResponseWriter, r *http.Request) {
		err := pages.NewGaugePage().Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering new gauge page: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/admin/gauges/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			log.Printf("Error parsing gauge ID: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		gauge, err := queries.GetGauge(r.Context(), id)
		if err != nil {
			log.Printf("Error getting gauge %d: %v", id, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		modelGauge := toGauge(gauge)
		err = components.GaugeForm(fmt.Sprintf("/admin/gauges/%d", id), "PUT", modelGauge).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering gauge form: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/admin/gauges/{id}/history", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			log.Printf("Error parsing gauge ID: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		history, err := queries.GetGaugeHistory(r.Context(), id)
		if err != nil {
			log.Printf("Error getting gauge history for %d: %v", id, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var historyData []models.GaugeHistory
		for _, h := range history {
			month, ok := h.Month.(string)
			if !ok {
				log.Printf("Error converting month to string for gauge %d", id)
				continue
			}
			historyData = append(historyData, models.GaugeHistory{
				Month:        month,
				AverageValue: h.AverageValue,
			})
		}

		err = components.GaugeHistory(historyData).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering gauge history: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// API routes
	r.Post("/gauges/{id}/increment", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			log.Printf("Error parsing gauge ID: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		gauge, err := queries.GetGauge(r.Context(), id)
		if err != nil {
			log.Printf("Error getting gauge %d: %v", id, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		currentValue, err := queries.GetCurrentValue(r.Context(), id)
		if err != nil {
			log.Printf("Error getting current value for gauge %d: %v", id, err)
			currentValue = 0
		}

		incrementValue := getIncrementValue(gauge.Unit)
		newValue := currentValue + incrementValue

		err = queries.CreateGaugeValue(r.Context(), db.CreateGaugeValueParams{
			GaugeID: id,
			Column2: newValue,
			Date:    time.Now(),
		})
		if err != nil {
			log.Printf("Error creating gauge value for %d: %v", id, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		modelGauge := toGauge(gauge)
		err = components.GaugeCard(modelGauge, newValue).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering gauge card: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Post("/gauges/{id}/decrement", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			log.Printf("Error parsing gauge ID: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		gauge, err := queries.GetGauge(r.Context(), id)
		if err != nil {
			log.Printf("Error getting gauge %d: %v", id, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		currentValue, err := queries.GetCurrentValue(r.Context(), id)
		if err != nil {
			log.Printf("Error getting current value for gauge %d: %v", id, err)
			currentValue = 0
		}

		decrementValue := getIncrementValue(gauge.Unit)
		newValue := currentValue - decrementValue
		if newValue < 0 {
			newValue = 0
		}

		err = queries.CreateGaugeValue(r.Context(), db.CreateGaugeValueParams{
			GaugeID: id,
			Column2: newValue,
			Date:    time.Now(),
		})
		if err != nil {
			log.Printf("Error creating gauge value for %d: %v", id, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		modelGauge := toGauge(gauge)
		err = components.GaugeCard(modelGauge, newValue).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering gauge card: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Post("/admin/gauges", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		target, err := strconv.ParseFloat(r.FormValue("target"), 64)
		if err != nil {
			log.Printf("Error parsing target value: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = queries.CreateGauge(r.Context(), db.CreateGaugeParams{
			Name:        r.FormValue("name"),
			Description: sql.NullString{String: r.FormValue("description"), Valid: true},
			Column3:     target,
			Unit:        r.FormValue("unit"),
		})
		if err != nil {
			log.Printf("Error creating gauge: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	})

	r.Put("/admin/gauges/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			log.Printf("Error parsing gauge ID: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		target, err := strconv.ParseFloat(r.FormValue("target"), 64)
		if err != nil {
			log.Printf("Error parsing target value: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = queries.UpdateGauge(r.Context(), db.UpdateGaugeParams{
			ID:          id,
			Name:        r.FormValue("name"),
			Description: sql.NullString{String: r.FormValue("description"), Valid: true},
			Column3:     target,
			Unit:        r.FormValue("unit"),
		})
		if err != nil {
			log.Printf("Error updating gauge %d: %v", id, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	})

	r.Delete("/admin/gauges/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			log.Printf("Error parsing gauge ID: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = queries.DeleteGauge(r.Context(), id)
		if err != nil {
			log.Printf("Error deleting gauge %d: %v", id, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
