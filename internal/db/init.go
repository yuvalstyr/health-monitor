package db

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

// Open initializes and returns a new database connection
func Open() (*sql.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "health-monitor.db"
	}

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Printf("Creating new database at %s\n", dbPath)
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, err
		}
		file.Close()
	}

	// Open database connection
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Run Goose migrations
	if err := RunMigrations(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}


