package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"

	"health-monitor/internal/db"
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
		log.Println("ROOT ROUTE HANDLER TRIGGERED")
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

	r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		log.Println("ADMIN ROUTE HANDLER TRIGGERED")
		gauges, err := queries.ListGauges(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		err = components.Layout(pages.Admin(gauges)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Special debug handler for admin/gauges/{id}
	log.Println("*** REGISTERING /admin/gauges/{id} ROUTE ***")

	// Define a special debug subrouter to handle the admin/gauges paths
	r.Route("/admin/gauges", func(r chi.Router) {
		// GET handler for displaying the edit form
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			log.Printf("*** ROUTE TRIGGERED: /admin/gauges/{id} with id=%s, full URL=%s ***\n", id, r.URL.String())

			idNum, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			gauge, err := queries.GetGauge(r.Context(), idNum)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("*** GAUGE RETRIEVED (ORIGINAL): %+v ***\n", gauge)
			
			// Fix any missing data to ensure form fields are populated
			if gauge.Name == "" {
				gauge.Name = fmt.Sprintf("Gauge %d", idNum)
				log.Printf("*** Fixed missing name: %s ***\n", gauge.Name)
			}
			if gauge.Unit == "" {
				gauge.Unit = "units"
				log.Printf("*** Fixed missing unit: %s ***\n", gauge.Unit)
			}
			if gauge.Icon == "" {
				gauge.Icon = "chart-bar"
				log.Printf("*** Fixed missing icon: %s ***\n", gauge.Icon)
			}
			
			log.Printf("*** GAUGE AFTER FIXES: %+v ***\n", gauge)
			w.Header().Set("Content-Type", "text/html")
			// Render the edit form with the gauge data
			err = components.Layout(components.GaugeForm("PUT", fmt.Sprintf("/admin/gauges/%d", idNum), &gauge, []components.FormError{})).Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
		
		// PUT handler for processing form submissions
		r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
			log.Println("*** EDIT FORM SUBMISSION HANDLER TRIGGERED ***")
			id := chi.URLParam(r, "id")
			log.Printf("Processing update for gauge ID: %s\n", id)
			
			// Parse form data
			err := r.ParseForm()
			if err != nil {
				log.Printf("Error parsing form: %v\n", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			
			// Log all form values for debugging
			log.Println("Form values:")
			for key, values := range r.Form {
				log.Printf("  %s: %v\n", key, values)
			}
			
			// Convert ID to int64
			idNum, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				log.Printf("Error parsing ID: %v\n", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			
			// Parse and validate form fields
			name := r.FormValue("name")
			if name == "" {
				log.Println("Name is required")
				http.Error(w, "Name is required", http.StatusBadRequest)
				return
			}
			
			targetStr := r.FormValue("target")
			target, err := strconv.ParseFloat(targetStr, 64)
			if err != nil {
				log.Printf("Error parsing target: %v\n", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			
			unit := r.FormValue("unit")
			icon := r.FormValue("icon")
			
			// Update the gauge in the database
			log.Printf("Updating gauge with ID: %d, Name: %s, Target: %f, Unit: %s, Icon: %s\n", 
				idNum, name, target, unit, icon)
			
			err = queries.UpdateGauge(r.Context(), db.UpdateGaugeParams{
				ID:          idNum,
				Name:        name,
				Description: sql.NullString{String: r.FormValue("description"), Valid: true},
				Target:      target,
				Unit:        unit,
				Icon:        icon,
			})
			
			if err != nil {
				log.Printf("Error updating gauge: %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			log.Println("Gauge updated successfully")
			
			// Redirect to admin page
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
		})
	})

	r.Get("/admin/gauges/new", func(w http.ResponseWriter, r *http.Request) {
		log.Println("NEW GAUGE ROUTE HANDLER TRIGGERED")
		w.Header().Set("Content-Type", "text/html")
		err := components.Layout(components.GaugeForm("POST", "/admin/gauges", nil, []components.FormError{})).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Post("/admin/gauges", func(w http.ResponseWriter, r *http.Request) {
		log.Println("CREATE GAUGE ROUTE HANDLER TRIGGERED")
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

		w.Header().Set("Content-Type", "text/html")
		err = components.Layout(pages.Admin(gauges)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Define a shared handler for both PUT and POST (with method override) requests
	gaugesEditHandler := func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Read the raw request body for debugging
		bodyBytes, _ := io.ReadAll(r.Body)
		bodyString := string(bodyBytes)
		fmt.Printf("Raw request body: %s\n", bodyString)

		// Important: Restore the body for parsing
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse the form
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf("Method: %s\n", r.Method)
		fmt.Printf("Content-Type: %s\n", r.Header.Get("Content-Type"))
		fmt.Printf("Form data: %+v\n", r.Form)

		// Get form values
		name := r.FormValue("name")
		targetStr := r.FormValue("target")
		unit := r.FormValue("unit")
		icon := r.FormValue("icon")

		// Debug log
		fmt.Printf("Form values - name: '%s', target: '%s', unit: '%s', icon: '%s'\n", name, targetStr, unit, icon)

		// Validate required fields
		var errors []components.FormError
		var target float64

		// Print each form field for debugging
		for key, values := range r.Form {
			fmt.Printf("Form field %s: %v\n", key, values)
		}

		fmt.Printf("\nLooking at specific fields:\n")
		fmt.Printf("name (direct): '%s'\n", r.Form.Get("name"))
		fmt.Printf("target (direct): '%s'\n", r.Form.Get("target"))
		fmt.Printf("unit (direct): '%s'\n", r.Form.Get("unit"))
		fmt.Printf("icon (direct): '%s'\n", r.Form.Get("icon"))

		// Only validate fields if they're truly empty, not just missing from the form
		var validationErrors []components.FormError

		// Parse target value if present
		if targetStr != "" {
			t, err := strconv.ParseFloat(targetStr, 64)
			if err != nil {
				validationErrors = append(validationErrors, components.FormError{Field: "target", Message: "Target must be a valid number"})
			} else {
				target = t
			}
		}

		// Only use the validation errors if we actually have any
		if len(validationErrors) > 0 {
			errors = validationErrors
		}

		if len(errors) > 0 {
			// Get the current gauge for re-rendering the form
			gauge, err := queries.GetGauge(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			err = components.Layout(components.GaugeForm("PUT", fmt.Sprintf("/admin/gauges/%d", id), &gauge, errors)).Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Log the values we're about to save
		fmt.Printf("\nAttempting to update gauge with ID %d with values:\n", id)
		fmt.Printf("Name: '%s'\n", name)
		fmt.Printf("Description: '%s'\n", r.FormValue("description"))
		fmt.Printf("Target: %.2f\n", target)
		fmt.Printf("Unit: '%s'\n", unit)
		fmt.Printf("Icon: '%s'\n", icon)

		// Create a gauge for logging
		params := db.UpdateGaugeParams{
			Name:        name,
			Description: sql.NullString{String: r.FormValue("description"), Valid: true},
			Target:      target,
			Unit:        unit,
			Icon:        icon,
			ID:          id,
		}

		// Execute the update
		err = queries.UpdateGauge(r.Context(), params)
		if err != nil {
			fmt.Printf("ERROR updating gauge: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("Database update successful!\n")

		// Get the updated gauge to display the edit form again
		updatedGauge, err := queries.GetGauge(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Debug the gauge data from the database
		fmt.Printf("\nUpdated gauge from database:\n")
		fmt.Printf("ID: %d\n", updatedGauge.ID)
		fmt.Printf("Name: '%s'\n", updatedGauge.Name)
		fmt.Printf("Description: '%v'\n", updatedGauge.Description)
		fmt.Printf("Target: %.2f\n", updatedGauge.Target)
		fmt.Printf("Unit: '%s'\n", updatedGauge.Unit)
		fmt.Printf("Icon: '%s'\n", updatedGauge.Icon)

		// Instead of using HTMX redirect, use a standard HTTP redirect
		// This ensures the browser completely reloads the page with fresh data
		http.Redirect(w, r, fmt.Sprintf("/admin/gauges/%d", id), http.StatusSeeOther)
	}

	// Register the handler for both PUT and POST with method override
	// Commented out to avoid conflict with subrouter handler
	// r.Put("/admin/gauges/{id}", gaugesEditHandler)

	// Handle POST requests with X-HTTP-Method-Override: PUT
	r.Post("/admin/gauges/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-HTTP-Method-Override") == "PUT" {
			gaugesEditHandler(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	r.Delete("/admin/gauges/{id}", func(w http.ResponseWriter, r *http.Request) {
		log.Println("DELETE GAUGE ROUTE HANDLER TRIGGERED")
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

		w.Header().Set("Content-Type", "text/html")
		err = components.Layout(pages.Admin(gauges)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

		w.Header().Set("Content-Type", "text/html")
		err = components.GaugeValue(&gauge, gauge.Value).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

		w.Header().Set("Content-Type", "text/html")
		err = components.GaugeValue(&gauge, gauge.Value).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Printf("Server starting on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}
