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

### Initial Setup

1. **Prerequisites**:
   - Go 1.21 or later
   - SQLite
   - Make

2. **Installation**:
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
