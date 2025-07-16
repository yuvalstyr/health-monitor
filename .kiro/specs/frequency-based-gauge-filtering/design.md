# Design Document

## Overview

This design implements frequency-based gauge filtering for the health monitor dashboard. The system will enhance the existing gauge functionality to support time-based activation and automated instance creation based on frequency settings (weekly, bi-weekly, monthly). The dashboard will intelligently display only relevant gauges for the current time period while providing visual feedback during transition phases.

Since this is a work-in-progress application not yet deployed to production, we can implement aggressive changes including database schema modifications, tooling improvements (migrating from Atlas to Goose), and structural enhancements without backward compatibility concerns.

## Architecture

The solution extends the existing MVC architecture with enhanced tooling and database schema changes:

### Core Components
- **Migration Layer**: Goose-based database migrations replacing Atlas
- **Database Layer**: Extended schema with new fields for gauge scheduling
- **Service Layer**: New scheduling service for automated gauge management
- **Handler Layer**: Enhanced dashboard handler with filtering logic
- **View Layer**: Updated dashboard templates with visual status indicators

### Data Flow
1. User creates/configures gauges with frequency and active status
2. Scheduling service runs periodically to create new gauge instances
3. Dashboard handler filters gauges based on current time period
4. View layer renders gauges with appropriate visual styling

## Components and Interfaces

### 1. Migration Tooling (Atlas → Goose)

**Goose Setup:**
```bash
# Install Goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Initialize migrations directory
mkdir -p migrations
```

**Migration File Structure:**
```
migrations/
├── 00001_initial_schema.sql
├── 00002_add_gauge_scheduling_fields.sql
└── 00003_create_scheduling_indexes.sql
```

**Goose Configuration:**
```go
// internal/db/migrate.go
package db

import (
    "database/sql"
    "embed"
    
    "github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func init() {
    // Register modernc.org/sqlite driver as "sqlite3" for Goose compatibility
    sql.Register("sqlite3", &sqlite.Driver{})
}

func RunMigrations(db *sql.DB) error {
    goose.SetBaseFS(embedMigrations)
    
    if err := goose.SetDialect("sqlite3"); err != nil {
        return err
    }
    
    return goose.Up(db, "migrations")
}
```

**Benefits of Goose over Atlas:**
- **Native Go Integration**: Embed migrations in binary
- **Simpler Configuration**: No external config files needed
- **Up/Down Migrations**: Easy rollback during development
- **SQLC Compatibility**: Works seamlessly with existing setup
- **Lightweight**: Minimal dependencies

### 2. Database Schema Design

**Separate Tables Approach:**
```sql
-- Gauge templates (user-created configurations)
CREATE TABLE gauge_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    target REAL NOT NULL,
    unit TEXT NOT NULL,
    icon TEXT NOT NULL DEFAULT 'chart-bar',
    frequency TEXT NOT NULL, -- 'weekly', 'bi-weekly', 'monthly'
    direction TEXT NOT NULL DEFAULT 'under',
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
```

**Key Design Decisions:**
- **Separate Tables**: Clean separation between templates and instances
- **Computed period_end**: Calculate from `period_start` + frequency (no storage needed)
- **Simple Structure**: Only essential fields, no complex relationships

### 3. Enhanced Database Queries

**New Queries:**
```sql
-- Get active gauge templates
-- name: ListActiveGaugeTemplates :many
SELECT * FROM gauge_templates WHERE active = true;

-- Get current period gauge instances for dashboard
-- name: ListCurrentPeriodGaugeInstances :many
SELECT gi.*, gt.name, gt.description, gt.target, gt.unit, gt.icon, gt.frequency, gt.direction
FROM gauge_instances gi
JOIN gauge_templates gt ON gi.template_id = gt.id
WHERE (gt.frequency = 'weekly' AND gi.period_start = ?) 
   OR (gt.frequency = 'bi-weekly' AND gi.period_start = ?)
   OR (gt.frequency = 'monthly' AND gi.period_start = ?)
ORDER BY gt.name;

-- Create gauge instance from template
-- name: CreateGaugeInstance :one
INSERT INTO gauge_instances (template_id, period_start, value)
VALUES (?, ?, 0)
RETURNING *;

-- Check if instance exists for period
-- name: InstanceExistsForPeriod :one
SELECT COUNT(*) FROM gauge_instances 
WHERE template_id = ? AND period_start = ?;
```

### 4. Scheduling Service

**Simplified Interface:**
```go
type SchedulingService interface {
    CreateInstancesForActiveTemplates(ctx context.Context) error
}
```

**Implementation Details:**
- Runs as a background goroutine with configurable interval (default: daily)
- For each active template, checks if **next period** instance exists
- If not, creates new instance with `period_start` = **next period start date**
- Creates instances ahead of time so they're ready when the period begins
- Simple logic: one function that handles everything

**Period Calculation Logic:**
```go
func CalculateCurrentPeriodStart(frequency string, currentTime time.Time) time.Time {
    switch frequency {
    case "weekly":
        // Monday of current week
        weekday := int(currentTime.Weekday())
        if weekday == 0 { weekday = 7 } // Sunday = 7
        return currentTime.AddDate(0, 0, -(weekday-1)).Truncate(24 * time.Hour)
    
    case "bi-weekly":
        // First or 15th of current month based on current date
        if currentTime.Day() <= 14 {
            return time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())
        } else {
            return time.Date(currentTime.Year(), currentTime.Month(), 15, 0, 0, 0, 0, currentTime.Location())
        }
    
    case "monthly":
        // First day of current month
        return time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())
    }
}

func CalculateNextPeriodStart(frequency string, currentTime time.Time) time.Time {
    switch frequency {
    case "weekly":
        // Monday of next week
        currentStart := CalculateCurrentPeriodStart(frequency, currentTime)
        return currentStart.AddDate(0, 0, 7)
    
    case "bi-weekly":
        // Next bi-weekly period
        if currentTime.Day() <= 14 {
            // Currently in first half, next is 15th
            return time.Date(currentTime.Year(), currentTime.Month(), 15, 0, 0, 0, 0, currentTime.Location())
        } else {
            // Currently in second half, next is 1st of next month
            return time.Date(currentTime.Year(), currentTime.Month()+1, 1, 0, 0, 0, 0, currentTime.Location())
        }
    
    case "monthly":
        // First day of next month
        return time.Date(currentTime.Year(), currentTime.Month()+1, 1, 0, 0, 0, 0, currentTime.Location())
    
    default:
        // Default to monthly for unknown frequencies
        return time.Date(currentTime.Year(), currentTime.Month()+1, 1, 0, 0, 0, 0, currentTime.Location())
    }
}
```

### 5. Enhanced Dashboard Handler

**Simplified Dashboard Logic:**
```go
func (h *GaugeHandler) handleDashboard(w http.ResponseWriter, r *http.Request) {
    currentTime := time.Now()
    
    // Get current period gauge instances only
    currentGauges := h.getCurrentPeriodGaugeInstances(r.Context(), currentTime)
    
    // Render simple dashboard
    h.renderDashboard(w, r, currentGauges)
}

func (h *GaugeHandler) getCurrentPeriodGaugeInstances(ctx context.Context, currentTime time.Time) []GaugeInstanceView {
    // Calculate current period start for each frequency
    weeklyStart := CalculateCurrentPeriodStart("weekly", currentTime)
    biWeeklyStart := CalculateCurrentPeriodStart("bi-weekly", currentTime)
    monthlyStart := CalculateCurrentPeriodStart("monthly", currentTime)
    
    // Query instances that match current periods
    return h.queries.ListCurrentPeriodGaugeInstances(ctx, weeklyStart, biWeeklyStart, monthlyStart)
}
```

**Simple Filtering Logic:**
- Only show gauge instances where `period_start` matches current period for that frequency
- No transition logic - just current active gauges
- Clean and simple dashboard display

### 6. Simplified View Components

**Dashboard Template Structure:**
```go
templ Dashboard(currentGauges []GaugeInstanceView) {
    <div class="dashboard-header">
        @CurrentPeriodIndicator()
    </div>
    
    <div class="current-gauges">
        if len(currentGauges) == 0 {
            <p>No active gauges for the current period.</p>
        } else {
            for _, gauge := range currentGauges {
                @components.GaugeCard(&gauge)
            }
        }
    </div>
}
```

**Simple Design:**
- Only show current period gauges
- Clean, minimal interface
- No complex transition states

## Data Models

### 1. Simplified Gauge Models

```go
type GaugeTemplate struct {
    ID          int64  `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Target      float64 `json:"target"`
    Unit        string `json:"unit"`
    Icon        string `json:"icon"`
    Frequency   string `json:"frequency"`
    Direction   string `json:"direction"`
    Active      bool   `json:"active"`
}

type GaugeInstanceView struct {
    ID          int64     `json:"id"`
    TemplateID  int64     `json:"template_id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Target      float64   `json:"target"`
    Value       float64   `json:"value"`
    Unit        string    `json:"unit"`
    Icon        string    `json:"icon"`
    Frequency   string    `json:"frequency"`
    Direction   string    `json:"direction"`
    PeriodStart time.Time `json:"period_start"`
}
```

### 2. Time Period Calculation

```go
type PeriodCalculator struct{}

func (pc *PeriodCalculator) CalculateWeeklyPeriod(t time.Time) (start, end time.Time) {
    // Monday to Sunday of current week
    weekday := int(t.Weekday())
    if weekday == 0 { weekday = 7 } // Sunday = 7
    start = t.AddDate(0, 0, -(weekday-1)).Truncate(24 * time.Hour)
    end = start.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
    return
}

func (pc *PeriodCalculator) CalculateBiWeeklyPeriod(t time.Time) (start, end time.Time) {
    // Determine which bi-weekly period of the month
    monthStart := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
    weekOfMonth := (t.Day()-1)/7 + 1
    
    if weekOfMonth <= 2 {
        // First bi-weekly period (weeks 1-2)
        start = monthStart
        end = monthStart.AddDate(0, 0, 13).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
    } else {
        // Second bi-weekly period (weeks 3-4+)
        start = monthStart.AddDate(0, 0, 14)
        end = monthStart.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
    }
    return
}

func (pc *PeriodCalculator) CalculateMonthlyPeriod(t time.Time) (start, end time.Time) {
    // First day to last day of current month
    start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
    end = start.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
    return
}
```

## Error Handling

### 1. Database Errors
- **Connection failures**: Graceful degradation, show cached data if available
- **Query failures**: Log error, return empty gauge list with user notification
- **Transaction failures**: Rollback and retry with exponential backoff

### 2. Scheduling Service Errors
- **Time calculation errors**: Log error, skip problematic gauge template
- **Instance creation failures**: Log error, continue with other templates
- **Duplicate instance detection**: Skip creation, log warning

### 3. Dashboard Rendering Errors
- **Template rendering failures**: Show basic HTML fallback
- **Data formatting errors**: Show gauge with default values
- **Missing gauge data**: Hide gauge, log warning

## Testing Strategy

### 1. Unit Tests

**Database Layer:**
- Test new queries with various time ranges
- Test gauge instance creation and retrieval
- Test period calculation functions

**Scheduling Service:**
- Test period calculations for all frequencies
- Test instance creation logic
- Test duplicate prevention
- Mock time for consistent testing

**Handler Layer:**
- Test dashboard filtering logic
- Test transition gauge detection
- Test error handling scenarios

### 2. Integration Tests

**End-to-End Scenarios:**
- Create gauge template → Activate → Verify instance creation
- Dashboard displays correct gauges for different time periods
- Transition periods show appropriate past/future gauges
- Automated scheduling creates instances at correct times

**Time-Based Testing:**
- Mock system time to test different periods
- Test month boundaries and edge cases
- Test leap year handling for monthly periods

### 3. Performance Tests

**Database Performance:**
- Query performance with large numbers of gauge instances
- Index optimization for time-based queries
- Cleanup of old gauge instances

**Scheduling Performance:**
- Background service resource usage
- Processing time for large numbers of templates
- Memory usage during bulk instance creation

## Migration Strategy

### 1. Goose Migration Setup

**Replace Atlas Configuration:**
```bash
# Remove Atlas files
rm atlas.hcl

# Install Goose
go install github.com/pressly/goose/v3/cmd/goose@latest
```

**Migration Files:**
```sql
-- migrations/00001_initial_schema.sql
-- +goose Up
CREATE TABLE gauge_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    target REAL NOT NULL,
    unit TEXT NOT NULL,
    icon TEXT NOT NULL DEFAULT 'chart-bar',
    frequency TEXT NOT NULL, -- 'weekly', 'bi-weekly', 'monthly'
    direction TEXT NOT NULL DEFAULT 'under',
    active BOOLEAN DEFAULT false,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE gauge_instances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INTEGER NOT NULL REFERENCES gauge_templates(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    value REAL NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE gauge_values (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gauge_id INTEGER NOT NULL,
    value REAL NOT NULL,
    date DATETIME NOT NULL,
    FOREIGN KEY (gauge_id) REFERENCES gauge_instances(id) ON DELETE CASCADE
);
```

**Data Model Relationships and Value Management:**

The gauge data model uses a three-tier structure to support frequency-based filtering:

1. **gauge_templates**: User-created configurations (the "blueprint")
2. **gauge_instances**: Auto-generated instances for specific time periods
3. **gauge_values**: Individual value entries over time within each instance

**Value Management Strategy:**
- **gauge_instances.value**: Current/latest value for the period (denormalized for performance)
- **gauge_values**: Historical log of all value changes within the period
- When a user updates a gauge, both tables are updated:
  - New entry added to `gauge_values` with timestamp
  - `gauge_instances.value` updated to reflect the latest value
  - `gauge_instances.updated_at` timestamp refreshed

**Relationship Flow:**
```
gauge_templates (1) → (many) gauge_instances (1) → (many) gauge_values
     ↓                        ↓                         ↓
  User config          Period-specific           Individual updates
                       gauge instance            within the period
```

**Why Both value Fields Exist:**
- **Performance**: `gauge_instances.value` allows fast dashboard queries without JOINs
- **History**: `gauge_values` preserves complete audit trail of changes
- **Flexibility**: Supports future features like value trends, undo operations, or detailed analytics

**Example Data Flow:**
1. User creates template: "Weekly Exercise" (target: 5 hours)
2. System creates instance: period_start="2024-01-01", value=0
3. User logs 2 hours: gauge_values entry created, instance.value=2
4. User logs 1.5 more hours: new gauge_values entry, instance.value=3.5
5. Dashboard shows current value (3.5) with history available if needed

```sql
-- Create indexes for performance
CREATE INDEX idx_gauge_templates_active ON gauge_templates(active);
CREATE INDEX idx_gauge_templates_frequency ON gauge_templates(frequency);
CREATE INDEX idx_gauge_instances_template ON gauge_instances(template_id);
CREATE INDEX idx_gauge_instances_period ON gauge_instances(period_start);

-- +goose Down
DROP INDEX IF EXISTS idx_gauge_instances_period;
DROP INDEX IF EXISTS idx_gauge_instances_template;
DROP INDEX IF EXISTS idx_gauge_templates_frequency;
DROP INDEX IF EXISTS idx_gauge_templates_active;
DROP TABLE IF EXISTS gauge_values;
DROP TABLE IF EXISTS gauge_instances;
DROP TABLE IF EXISTS gauge_templates;
```

### 2. Aggressive Schema Changes

Since this is a WIP application, we can:
- **Recreate Database**: Drop and recreate with new schema
- **No Backward Compatibility**: Focus on optimal design
- **Schema Optimization**: Add all necessary indexes from the start
- **Data Reset**: Start fresh with new gauge structure

**Database Initialization:**
```go
// Update internal/db/init.go
func InitializeDatabase(db *sql.DB) error {
    // Run Goose migrations
    if err := RunMigrations(db); err != nil {
        return fmt.Errorf("failed to run migrations: %w", err)
    }
    
    // Seed with sample data if needed
    return seedSampleData(db)
}
```

### 3. Development Workflow

**Migration Commands:**
```bash
# Apply migrations
goose -dir migrations sqlite3 health.db up

# Rollback migrations
goose -dir migrations sqlite3 health.db down

# Check migration status
goose -dir migrations sqlite3 health.db status

# Create new migration
goose -dir migrations create add_new_feature sql
```

**Integration with Build Process:**
```go
// Embed migrations in binary
//go:embed migrations/*.sql
var embedMigrations embed.FS

// Auto-migrate on startup
func main() {
    db := setupDatabase()
    if err := RunMigrations(db); err != nil {
        log.Fatal("Migration failed:", err)
    }
    // ... rest of application
}
```