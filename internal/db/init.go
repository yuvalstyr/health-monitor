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

	// Open database connection with proper SQLite settings for concurrency
	db, err := sql.Open("sqlite", dbPath+"?_busy_timeout=30000&_journal_mode=DELETE&_foreign_keys=on&_synchronous=NORMAL")
	if err != nil {
		return nil, err
	}
	
	// Set reasonable connection pool settings for SQLite
	db.SetMaxOpenConns(1)  // SQLite works best with single writer
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // Connections never expire

	// Run Goose migrations
	if err := RunMigrations(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}


