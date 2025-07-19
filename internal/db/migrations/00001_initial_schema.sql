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
    week INTEGER NOT NULL CHECK (week >= 1 AND week <= 53),
    year INTEGER NOT NULL CHECK (year >= 1900 AND year <= 2100),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (gauge_id) REFERENCES gauges(id) ON DELETE CASCADE,
    -- Ensure week 53 only exists in years that actually have 53 weeks
    CHECK (week <= 52 OR (week = 53 AND (
        -- Years with 53 weeks: years starting on Thursday or leap years starting on Wednesday
        (year % 4 = 0 AND year % 100 != 0) OR (year % 400 = 0) OR
        (strftime('%w', year || '-01-01') = '4') OR
        ((year % 4 = 0 AND year % 100 != 0 OR year % 400 = 0) AND strftime('%w', year || '-01-01') = '3')
    )))
);

-- Create indexes
CREATE INDEX idx_gauge_values_gauge_id ON gauge_values(gauge_id);
CREATE INDEX idx_gauge_values_week_year ON gauge_values(week, year);

-- +goose Down
DROP INDEX IF EXISTS idx_gauge_values_week_year;
DROP INDEX IF EXISTS idx_gauge_values_gauge_id;
DROP TABLE IF EXISTS gauge_values;
DROP TABLE IF EXISTS gauges;