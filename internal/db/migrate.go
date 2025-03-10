package db

import (
	"database/sql"
	"fmt"
	"log"
)

func Migrate(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS gauges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			target REAL NOT NULL,
			unit TEXT NOT NULL,
			icon TEXT DEFAULT 'chart-bar'
		)`,
		`CREATE TABLE IF NOT EXISTS measurements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			gauge_id INTEGER NOT NULL,
			value REAL NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (gauge_id) REFERENCES gauges(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_measurements_gauge_id ON measurements(gauge_id)`,
		`CREATE INDEX IF NOT EXISTS idx_measurements_timestamp ON measurements(timestamp)`,
	}

	for _, migration := range migrations {
		_, err := db.Exec(migration)
		if err != nil {
			return fmt.Errorf("%w\n%s", err, migration)
		}
	}

	// Check if icon column exists, if not add it
	var hasIconColumn bool
	err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('gauges') WHERE name='icon'").Scan(&hasIconColumn)
	if err != nil {
		return fmt.Errorf("error checking for icon column: %w", err)
	}

	if !hasIconColumn {
		log.Println("Adding icon column to gauges table...")
		_, err := db.Exec(`ALTER TABLE gauges ADD COLUMN icon TEXT DEFAULT 'chart-bar'`)
		if err != nil {
			return fmt.Errorf("error adding icon column: %w", err)
		}
	}

	return nil
}
