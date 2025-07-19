DROP TABLE IF EXISTS gauge_values;
DROP TABLE IF EXISTS gauge_instances;
DROP TABLE IF EXISTS gauge_templates;

-- Gauge templates (user-created configurations)
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

-- Gauge instances (auto-generated for specific time periods)
CREATE TABLE gauge_instances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INTEGER NOT NULL REFERENCES gauge_templates(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    value REAL NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Gauge values (historical log of value changes within each instance)
CREATE TABLE gauge_values (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gauge_id INTEGER NOT NULL,
    value REAL NOT NULL,
    date DATETIME NOT NULL,
    FOREIGN KEY (gauge_id) REFERENCES gauge_instances(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_gauge_templates_active ON gauge_templates(active);
CREATE INDEX idx_gauge_templates_frequency ON gauge_templates(frequency);
CREATE INDEX idx_gauge_instances_template ON gauge_instances(template_id);
CREATE INDEX idx_gauge_instances_period ON gauge_instances(period_start);
CREATE UNIQUE INDEX idx_gauge_instances_template_period ON gauge_instances(template_id, period_start);
