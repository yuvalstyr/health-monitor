package main

import (
	"health-monitor/internal/app"
)

// main starts the health-monitor web service
func main() {
	application := app.New()
	application.Run()
}
