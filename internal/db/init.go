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

	// Create tables if they don't exist
	if err := initSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// initSchema creates the database tables if they don't exist
func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS gauges (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		target INTEGER NOT NULL,
		unit TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS gauge_values (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		gauge_id INTEGER NOT NULL,
		value INTEGER NOT NULL,
		week INTEGER NOT NULL,
		year INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (gauge_id) REFERENCES gauges(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_gauge_values_week_year ON gauge_values(week, year);
	CREATE INDEX IF NOT EXISTS idx_gauge_values_gauge_id ON gauge_values(gauge_id);
	`

	_, err := db.Exec(schema)
	return err
}
