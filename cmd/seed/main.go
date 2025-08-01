package main

import (
	"context"
	"flag"
	"log"
	"os"

	"health-monitor/internal/db"
)

func main() {
	var (
		dbPath = flag.String("db", "health-monitor.db", "Path to the SQLite database file")
		help   = flag.Bool("help", false, "Show help message")
	)
	
	flag.Usage = func() {
		log.Printf("Database Seeding Tool for Health Monitor\n")
		log.Printf("This tool populates the database with sample gauge templates and instances for development and testing.\n\n")
		log.Printf("Usage: %s [options]\n\n", os.Args[0])
		log.Printf("Options:\n")
		flag.PrintDefaults()
		log.Printf("\nExamples:\n")
		log.Printf("  %s                    # Seed default database (health-monitor.db)\n", os.Args[0])
		log.Printf("  %s -db test.db        # Seed custom database file\n", os.Args[0])
		log.Printf("  make seed             # Use Makefile target\n")
	}
	
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	log.Printf("Starting database seeding with database: %s", *dbPath)

	// Open database connection
	database, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Create seeder
	seeder := db.NewSeedData(database)

	// Run seeding
	ctx := context.Background()
	if err := seeder.SeedDatabase(ctx); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	log.Println("Database seeding completed successfully!")
}