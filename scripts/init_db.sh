#!/bin/bash
set -e

# Create the database directory if it doesn't exist
mkdir -p data

# Create the database file
cat > data/health-monitor.db.sql << 'EOF'
CREATE TABLE gauges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    target INTEGER NOT NULL,
    unit TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE gauge_values (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gauge_id INTEGER NOT NULL,
    value INTEGER NOT NULL,
    week INTEGER NOT NULL,
    year INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (gauge_id) REFERENCES gauges (id) ON DELETE CASCADE
);
EOF

# Initialize the database
sqlite3 data/health-monitor.db ".read data/health-monitor.db.sql"

echo "Database initialized successfully!"
