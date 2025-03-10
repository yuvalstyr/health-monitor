package db

import (
	"database/sql"
	"log"
)

// Migrate creates the necessary database tables if they don't exist
func Migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS gauges (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		target REAL NOT NULL,
		unit TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS gauge_values (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		gauge_id INTEGER NOT NULL,
		value REAL NOT NULL,
		date DATETIME NOT NULL,
		FOREIGN KEY (gauge_id) REFERENCES gauges(id) ON DELETE CASCADE
	);
	`

	log.Printf("Running database migrations...")
	_, err := db.Exec(schema)
	if err != nil {
		log.Printf("Error running migrations: %v", err)
		return err
	}
	log.Printf("Database migrations completed successfully")
	return nil
}
