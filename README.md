# Personal Health Monitor

A personal health monitoring application built with Go, HTMX, and DaisyUI. Track weekly health metrics with visual gauges and historical trends.

## Features
- Weekly health metrics dashboard with target-based gauges
- Admin interface for managing metrics and targets
- Historical trends visualization (monthly and yearly)
- Visual indicators for above/below target metrics

## Tech Stack
- Backend: Go with Chi router
- Database: SQLite with SQLC
- Frontend: HTMX + DaisyUI + Templ
- Templates: Templ for type-safe HTML templates

## Project Structure
```
health-monitor/
├── cmd/
│   └── server/          # Main application entry point
│       └── main.go
├── data/               # Application data files
│   └── *.db           # SQLite database files
├── internal/
│   ├── db/            # Database layer (SQLC generated code)
│   │   ├── db.go      # Generated database interface
│   │   ├── models.go  # Generated database models
│   │   ├── queries.sql # SQL queries
│   │   └── schema.sql # Database schema
│   ├── handlers/      # HTTP request handlers
│   ├── models/        # Domain models and business logic
│   └── views/
│       └── components/ # Templ components
│           ├── gauge.templ
│           ├── gauge_form.templ
│           ├── gauge_list.templ
│           └── layout.templ
├── migrations/        # Database migration files
├── scripts/          # Development and maintenance scripts
├── .env              # Environment configuration
├── Makefile         # Build and development commands
└── sqlc.yaml        # SQLC configuration
```

## Development Guide

### Prerequisites

1. **Go**: Version 1.21+ required
   ```bash
   # Check your Go version
   go version
   ```

2. **Environment Setup**:
   - Add Go bin directory to your PATH for easy access to installed tools:
   ```bash
   # Add this to your ~/.zshrc or ~/.bash_profile
   echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.zshrc
   source ~/.zshrc
   ```

### Required Tools Installation

1. **Templ**: For type-safe HTML templates
   ```bash
   go install github.com/a-h/templ/cmd/templ@latest
   ```

2. **SQLC**: For generating type-safe database code
   ```bash
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   ```

3. **Air**: For live reloading during development
   ```bash
   go install github.com/cosmtrek/air@latest
   ```

4. **Atlas**: For database migrations
   ```bash
   # Using Homebrew
   brew install ariga/tap/atlas
   # Or direct download from https://atlasgo.io
   ```

5. **SQLite**: For database management
   ```bash
   # On macOS (usually pre-installed)
   brew install sqlite
   ```

Alternatively, you can run:
```bash
make install-tools
```

### Database Setup

There is currently a mismatch between the schema in migrations/schema.sql and what the application expects. The application requires additional columns in the `gauges` table:

1. **Using manual setup** (recommended for now):
   ```bash
   # Create the database with the correct schema
   # Apply the schema from the provided file
   sqlite3 health.db < updated_schema.sql
   cat updated_schema.sql | sqlite3 health.db
   ```

2. **Using Atlas** (currently needs schema update):
   ```bash
   # Note: This will use the current schema.sql which may be missing required columns
   # First, update the atlas.hcl file to point to health.db instead of health-monitor.db
   # Then run:
   make migrations
   ```

### Running the Application

1. **Code Generation**:
   ```bash
   make generate
   ```

2. **Development Mode with Live Reload**:
   ```bash
   make dev
   # Or directly:
   $HOME/go/bin/air
   ```

3. **Building for Production**:
   ```bash
   make build
   ```

4. **Running Tests**:
   ```bash
   make test
   ```

### Common Issues and Solutions

1. **PATH Issues**: Ensure $HOME/go/bin is in your PATH to access installed Go tools
   ```bash
   echo $PATH | grep "go/bin"
   ```

2. **Database Schema Mismatch**: If you encounter errors like "no such column: value":
   - Use the manual schema creation steps described above
   - Update the migrations/schema.sql file to match the application's expectations

3. **Air Not Found**: If air is not found, use the full path:
   ```bash
   $HOME/go/bin/air
   ```

4. **Port Conflicts**: If port 3000 is already in use:
   ```bash
   # Find and kill processes using port 3000
   lsof -i :3000 | awk 'NR!=1 {print $2}' | xargs kill -9 2>/dev/null || true
   ```

## Notes on Migration Management

Currently, there is a discrepancy between:
1. The schema defined in `migrations/schema.sql` (missing columns like `value`, `icon`, etc.)
2. The schema expected by the application (queries.sql references columns not in schema.sql)
3. The database file referenced in atlas.hcl (health-monitor.db) vs main.go (health.db)

To properly align these, consider:
1. Updating `migrations/schema.sql` with the complete schema
2. Updating `atlas.hcl` to reference the correct database file
3. Running a proper migration with Atlas once everything is aligned

### Installation
   ```bash
   # Clone the repository
   git clone https://github.com/yourusername/health-monitor.git
   cd health-monitor

   # Install dependencies
   go mod download

   # Initialize database
   make migrate
   ```

### Database Changes

1. **Modifying the Schema**:
   - Edit `internal/db/schema.sql` to add/modify tables
   - Run migrations:
   ```bash
   make migrate
   ```

2. **Adding/Modifying Queries**:
   - Add or modify queries in `internal/db/queries.sql`
   - Generate SQLC code:
   ```bash
   make sqlc
   ```
   - This will update:
     - `internal/db/db.go`: Database interface
     - `internal/db/models.go`: Go structs for database models
     - `internal/db/querier.go`: Query interface

3. **Query Best Practices**:
   - Use named parameters (e.g., `@name`, `@id`)
   - Add clear comments for complex queries
   - Consider indexing for performance
   - Test queries with sample data

### Template Development

1. **Creating New Templates**:
   - Create a new `.templ` file in `internal/views/components/`
   - Use the Templ syntax for type-safe templates:
   ```go
   package components

   templ MyComponent(data string) {
       <div>{ data }</div>
   }
   ```

2. **Modifying Existing Templates**:
   - Edit the `.templ` files
   - Generate Templ code:
   ```bash
   make templ
   ```
   - Hot reload is active during development

3. **Template Best Practices**:
   - Use components for reusable UI elements
   - Leverage HTMX attributes for dynamic behavior
   - Follow DaisyUI classes for consistent styling
   - Keep components focused and maintainable
   - Use proper HTML semantics
   - Consider accessibility

### Development Workflow

1. **Starting Development Server**:
   ```bash
   make run
   ```
   This will:
   - Generate SQLC code
   - Generate Templ code
   - Start the server with hot reload

2. **Making Changes**:
   - Database changes:
     1. Modify schema.sql or queries.sql
     2. Run `make migrate` and `make sqlc`
   - Template changes:
     1. Edit .templ files
     2. Run `make templ`
   - Server changes:
     1. Edit Go files
     2. Server will auto-reload

3. **Common Commands**:
   ```bash
   make run          # Start development server
   make migrate      # Run database migrations
   make sqlc        # Generate SQLC code
   make templ       # Generate Templ code
   make clean       # Clean generated files
   ```

### HTMX Integration

1. **Form Submissions**:
   - Use `hx-post`, `hx-put`, `hx-delete` for form actions
   - Set `hx-target` and `hx-swap` for response handling
   - Add `hx-push-url` for proper URL history
   - Example:
   ```html
   <form 
     hx-put="/admin/gauges/1"
     hx-target="body"
     hx-swap="outerHTML"
     hx-push-url="/admin">
   ```

2. **Dynamic Updates**:
   - Use `hx-get` for polling or triggering updates
   - Set appropriate swap strategies
   - Consider using `hx-boost` for enhanced links
   - Example:
   ```html
   <div 
     hx-get="/gauges/1/value"
     hx-trigger="every 5s">
   ```

### Project Organization

1. **Code Structure**:
   - `cmd/server/`: Application entry point and routing
   - `data/`: Application data storage
     - SQLite database files
     - Backup and temporary files
   - `internal/`: Core application code
     - `db/`: Database layer (SQLC generated)
       - `schema.sql`: Database schema definition
       - `queries.sql`: SQL queries
       - `models.go`: Generated structs
       - `db.go`: Generated database interface
     - `handlers/`: HTTP request handlers
       - Route-specific request handling
       - Request validation
       - Response formatting
     - `models/`: Domain models and business logic
       - Core business entities
       - Business rules and validation
       - Service layer interfaces
     - `views/components/`: UI templates
       - Reusable Templ components
       - Page layouts
       - Form templates
   - `migrations/`: Database migration files
     - Version-controlled schema changes
     - Migration scripts
   - `scripts/`: Development and maintenance
     - Build scripts
     - Database maintenance
     - Development utilities
   
2. **Dependencies**:
   - Chi: HTTP routing
   - SQLC: Type-safe SQL
   - Templ: Type-safe templates
   - HTMX: Dynamic UI updates
   - DaisyUI: Tailwind-based UI components

3. **Best Practices**:
   - Keep SQL queries in queries.sql
   - Use components for UI reusability
   - Follow Go standard project layout
   - Leverage type safety with SQLC and Templ
   - Write clear commit messages
   - Document complex logic
   - Follow consistent code formatting

### Testing

The project includes comprehensive unit tests for handlers, models, and utilities. Tests are written using Go's standard testing package.

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# View coverage report in browser
make test-coverage-html
```

### Test Structure

- `internal/testutil/`: Test utilities and helpers
  * `db.go`: Database test utilities
- `internal/handlers/`: Handler tests
  * `gauge_handler_test.go`: Tests for gauge-related handlers
- `internal/models/`: Model tests
  * `gauge_test.go`: Tests for gauge-related models and utilities

### Test Coverage

We aim to maintain high test coverage for critical components:
- Database operations
- HTTP handlers
- Business logic
- Utility functions

### Troubleshooting

1. **Common Issues**:
   - Port already in use: Kill existing process
   - Database locked: Check connections
   - Template errors: Verify syntax

2. **Debugging Tools**:
   - SQLite CLI for database inspection
   - Browser dev tools for HTMX
   - Go debugger for server code

### CodeRabbit Reviews
![CodeRabbit Pull Request Reviews](https://img.shields.io/coderabbit/prs/github/yuvalstyr/health-monitor?utm_source=oss&utm_medium=github&utm_campaign=yuvalstyr%2Fhealth-monitor&labelColor=171717&color=FF570A&link=https%3A%2F%2Fcoderabbit.ai&label=CodeRabbit+Reviews)