-- +goose Up
-- Change gauge values from REAL to INTEGER

-- Update gauge_templates table to use INTEGER for target
ALTER TABLE gauge_templates RENAME TO gauge_templates_old;

CREATE TABLE gauge_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    target INTEGER NOT NULL,
    unit TEXT NOT NULL,
    icon TEXT NOT NULL DEFAULT 'chart-bar',
    frequency TEXT NOT NULL CHECK (frequency IN ('weekly', 'bi-weekly', 'monthly')),
    direction TEXT NOT NULL DEFAULT 'under' CHECK (direction IN ('under', 'over')),
    active BOOLEAN DEFAULT false,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Copy data from old table, converting REAL to INTEGER
INSERT INTO gauge_templates (id, name, description, target, unit, icon, frequency, direction, active, created_at, updated_at)
SELECT id, name, description, CAST(target AS INTEGER), unit, icon, frequency, direction, active, created_at, updated_at
FROM gauge_templates_old;

-- Drop old table
DROP TABLE gauge_templates_old;

-- Update gauge_instances table to use INTEGER for value
ALTER TABLE gauge_instances RENAME TO gauge_instances_old;

CREATE TABLE gauge_instances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INTEGER NOT NULL,
    period_start DATE NOT NULL,
    value INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES gauge_templates(id) ON DELETE CASCADE
);

-- Copy data from old table, converting REAL to INTEGER
INSERT INTO gauge_instances (id, template_id, period_start, value, created_at, updated_at)
SELECT id, template_id, period_start, CAST(value AS INTEGER), created_at, updated_at
FROM gauge_instances_old;

-- Drop old table
DROP TABLE gauge_instances_old;

-- Update gauge_values table to use INTEGER for value
ALTER TABLE gauge_values RENAME TO gauge_values_old;

CREATE TABLE gauge_values (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gauge_id INTEGER NOT NULL,
    value INTEGER NOT NULL,
    date DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (gauge_id) REFERENCES gauge_instances(id) ON DELETE CASCADE
);

-- Copy data from old table, converting REAL to INTEGER
INSERT INTO gauge_values (id, gauge_id, value, date, created_at)
SELECT id, gauge_id, CAST(value AS INTEGER), date, created_at
FROM gauge_values_old;

-- Drop old table
DROP TABLE gauge_values_old;

-- Recreate indexes for performance
CREATE INDEX idx_gauge_templates_active ON gauge_templates(active);
CREATE INDEX idx_gauge_templates_frequency ON gauge_templates(frequency);
CREATE INDEX idx_gauge_instances_template ON gauge_instances(template_id);
CREATE INDEX idx_gauge_instances_period ON gauge_instances(period_start);
CREATE INDEX idx_gauge_values_gauge_id ON gauge_values(gauge_id);
CREATE INDEX idx_gauge_values_date ON gauge_values(date);

-- +goose Down
-- Revert gauge values back to REAL

-- Revert gauge_templates table to use REAL for target
ALTER TABLE gauge_templates RENAME TO gauge_templates_old;

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

-- Copy data back, converting INTEGER to REAL
INSERT INTO gauge_templates (id, name, description, target, unit, icon, frequency, direction, active, created_at, updated_at)
SELECT id, name, description, CAST(target AS REAL), unit, icon, frequency, direction, active, created_at, updated_at
FROM gauge_templates_old;

DROP TABLE gauge_templates_old;

-- Revert gauge_instances table to use REAL for value
ALTER TABLE gauge_instances RENAME TO gauge_instances_old;

CREATE TABLE gauge_instances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INTEGER NOT NULL,
    period_start DATE NOT NULL,
    value REAL NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES gauge_templates(id) ON DELETE CASCADE
);

-- Copy data back, converting INTEGER to REAL
INSERT INTO gauge_instances (id, template_id, period_start, value, created_at, updated_at)
SELECT id, template_id, period_start, CAST(value AS REAL), created_at, updated_at
FROM gauge_instances_old;

DROP TABLE gauge_instances_old;

-- Revert gauge_values table to use REAL for value
ALTER TABLE gauge_values RENAME TO gauge_values_old;

CREATE TABLE gauge_values (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gauge_id INTEGER NOT NULL,
    value REAL NOT NULL,
    date DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (gauge_id) REFERENCES gauge_instances(id) ON DELETE CASCADE
);

-- Copy data back, converting INTEGER to REAL
INSERT INTO gauge_values (id, gauge_id, value, date, created_at)
SELECT id, gauge_id, CAST(value AS REAL), date, created_at
FROM gauge_values_old;

DROP TABLE gauge_values_old;

-- Recreate indexes
CREATE INDEX idx_gauge_templates_active ON gauge_templates(active);
CREATE INDEX idx_gauge_templates_frequency ON gauge_templates(frequency);
CREATE INDEX idx_gauge_instances_template ON gauge_instances(template_id);
CREATE INDEX idx_gauge_instances_period ON gauge_instances(period_start);
CREATE INDEX idx_gauge_values_gauge_id ON gauge_values(gauge_id);
CREATE INDEX idx_gauge_values_date ON gauge_values(date);