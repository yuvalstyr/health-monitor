package main

// Force Railway deployment - updated Makefile with seeding fixes
import (
	"log"
	"os"

	"health-monitor/internal/app"
)

// main starts the health-monitor web service
func main() {
	log.Println("🚀 Starting health-monitor application...")
	
	application := app.New()
	if err := application.Run(); err != nil {
		log.Printf("❌ CRITICAL ERROR: Application failed to start: %v", err)
		log.Printf("❌ Error details: %+v", err)
		os.Exit(1)
	}
	
	log.Println("✅ Application shut down gracefully")
}
