-- +goose Up
-- Create gauges table
CREATE TABLE gauges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    target REAL NOT NULL,
    value REAL DEFAULT 0,
    unit TEXT NOT NULL,
    icon TEXT DEFAULT 'chart-bar',
    frequency TEXT,
    direction TEXT DEFAULT 'under',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create gauge_values table
CREATE TABLE gauge_values (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gauge_id INTEGER NOT NULL,
    value REAL NOT NULL,
    week INTEGER NOT NULL,
    year INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (gauge_id) REFERENCES gauges(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX idx_gauge_values_gauge_id ON gauge_values(gauge_id);
CREATE INDEX idx_gauge_values_week_year ON gauge_values(week, year);

-- +goose Down
DROP INDEX IF EXISTS idx_gauge_values_week_year;
DROP INDEX IF EXISTS idx_gauge_values_gauge_id;
DROP TABLE IF EXISTS gauge_values;
DROP TABLE IF EXISTS gauges;