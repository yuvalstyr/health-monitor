-- +goose Up
-- Create gauge_templates table
CREATE TABLE gauge_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    target REAL NOT NULL,
    unit TEXT NOT NULL,
    icon TEXT NOT NULL DEFAULT 'chart-bar',
    frequency TEXT NOT NULL CHECK (frequency IN ('weekly', 'bi-weekly', 'monthly')),
    direction TEXT NOT NULL DEFAULT 'under' CHECK (direction IN ('under', 'over')),
    active BOOLEAN DEFAULT false,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create gauge_instances table
CREATE TABLE gauge_instances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INTEGER NOT NULL,
    period_start DATE NOT NULL,
    value REAL NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES gauge_templates(id) ON DELETE CASCADE
);

-- Drop the old gauge_values table
DROP TABLE IF EXISTS gauge_values;

-- Create new gauge_values table that references gauge_instances
CREATE TABLE gauge_values (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gauge_id INTEGER NOT NULL,
    value REAL NOT NULL,
    date DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (gauge_id) REFERENCES gauge_instances(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_gauge_templates_active ON gauge_templates(active);
CREATE INDEX idx_gauge_templates_frequency ON gauge_templates(frequency);
CREATE INDEX idx_gauge_instances_template ON gauge_instances(template_id);
CREATE INDEX idx_gauge_instances_period ON gauge_instances(period_start);
CREATE INDEX idx_gauge_values_gauge_id ON gauge_values(gauge_id);
CREATE INDEX idx_gauge_values_date ON gauge_values(date);

-- +goose Down
-- Drop indexes
DROP INDEX IF EXISTS idx_gauge_values_date;
DROP INDEX IF EXISTS idx_gauge_values_gauge_id;
DROP INDEX IF EXISTS idx_gauge_instances_period;
DROP INDEX IF EXISTS idx_gauge_instances_template;
DROP INDEX IF EXISTS idx_gauge_templates_frequency;
DROP INDEX IF EXISTS idx_gauge_templates_active;

-- Drop tables
DROP TABLE IF EXISTS gauge_values;
DROP TABLE IF EXISTS gauge_instances;
DROP TABLE IF EXISTS gauge_templates;

-- Recreate original gauge_values table
CREATE TABLE gauge_values (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gauge_id INTEGER NOT NULL,
    value REAL NOT NULL,
    week INTEGER NOT NULL,
    year INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (gauge_id) REFERENCES gauges(id) ON DELETE CASCADE
);

-- Recreate original indexes
CREATE INDEX idx_gauge_values_gauge_id ON gauge_values(gauge_id);
CREATE INDEX idx_gauge_values_week_year ON gauge_values(week, year);