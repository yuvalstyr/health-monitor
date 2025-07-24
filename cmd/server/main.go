package main

import (
	"log"
	"os"

	"health-monitor/internal/app"
)

// main starts the health-monitor web service
func main() {
	application := app.New()
	if err := application.Run(); err != nil {
		log.Printf("Application failed to start: %v", err)
		os.Exit(1)
	}
}
