package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"health-monitor/internal/db"
	"health-monitor/internal/views/components"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"
)

func main() {
	database, err := sql.Open("sqlite", "health.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	queries := db.New(database)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		gauges, err := queries.ListGauges(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		components.Layout(components.GaugeList(gauges)).Render(r.Context(), w)
	})

	r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		gauges, err := queries.ListGauges(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		components.Layout(components.GaugeList(gauges)).Render(r.Context(), w)
	})

	r.Get("/gauges/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		gauge, err := queries.GetGauge(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		components.Layout(components.GaugeView(&gauge)).Render(r.Context(), w)
	})

	r.Get("/admin/gauges/new", func(w http.ResponseWriter, r *http.Request) {
		components.Layout(components.GaugeForm("POST", "/admin/gauges", nil)).Render(r.Context(), w)
	})

	r.Get("/admin/gauges/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		gauge, err := queries.GetGauge(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		components.Layout(components.GaugeForm("PUT", fmt.Sprintf("/admin/gauges/%d", id), &gauge)).Render(r.Context(), w)
	})

	r.Post("/admin/gauges", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		target, err := strconv.ParseFloat(r.FormValue("target"), 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = queries.CreateGauge(r.Context(), db.CreateGaugeParams{
			Name:        r.FormValue("name"),
			Description: sql.NullString{String: r.FormValue("description"), Valid: true},
			Target:      target,
			Unit:        r.FormValue("unit"),
			Icon:        r.FormValue("icon"),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		gauges, err := queries.ListGauges(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		components.Layout(components.GaugeList(gauges)).Render(r.Context(), w)
	})

	r.Put("/admin/gauges/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		target, err := strconv.ParseFloat(r.FormValue("target"), 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = queries.UpdateGauge(r.Context(), db.UpdateGaugeParams{
			ID:          id,
			Name:        r.FormValue("name"),
			Description: sql.NullString{String: r.FormValue("description"), Valid: true},
			Target:      target,
			Unit:        r.FormValue("unit"),
			Icon:        r.FormValue("icon"),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		gauges, err := queries.ListGauges(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		components.Layout(components.GaugeList(gauges)).Render(r.Context(), w)
	})

	r.Delete("/admin/gauges/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = queries.DeleteGauge(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		gauges, err := queries.ListGauges(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		components.Layout(components.GaugeList(gauges)).Render(r.Context(), w)
	})

	r.Post("/gauges/{id}/increment", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		gauge, err := queries.GetGauge(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = queries.UpdateGaugeValue(r.Context(), db.UpdateGaugeValueParams{
			ID:    id,
			Value: gauge.Value + 1,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		gauge, err = queries.GetGauge(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%.1f %s", gauge.Value, gauge.Unit)
	})

	r.Post("/gauges/{id}/decrement", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		gauge, err := queries.GetGauge(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = queries.UpdateGaugeValue(r.Context(), db.UpdateGaugeValueParams{
			ID:    id,
			Value: gauge.Value - 1,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		gauge, err = queries.GetGauge(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%.1f %s", gauge.Value, gauge.Unit)
	})

	log.Printf("Server starting on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}
