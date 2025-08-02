# Health Monitor Project

## Configuration Files
- `.github/workflows/ci-cd.yml` - CI/CD pipeline configuration
- `.gitignore` - Git ignore rules
- `.air.toml` - Air live reload configuration

## Project Specs (.kiro)
- `.kiro/specs/deployment-infrastructure/design.md` - Deployment infrastructure design
- `.kiro/specs/deployment-infrastructure/requirements.md` - Deployment infrastructure requirements
- `.kiro/specs/deployment-infrastructure/tasks.md` - Deployment infrastructure tasks
- `.kiro/specs/frequency-based-gauge-filtering/design.md` - Gauge filtering design
- `.kiro/specs/frequency-based-gauge-filtering/requirements.md` - Gauge filtering requirements
- `.kiro/specs/frequency-based-gauge-filtering/tasks.md` - Gauge filtering tasks
- `.kiro/specs/mobile-ux-improvements/design-guidelines.md` - Mobile UX design guidelines
- `.kiro/specs/mobile-ux-improvements/requirements.md` - Mobile UX requirements
- `.kiro/specs/mobile-ux-improvements/tasks.md` - Mobile UX tasks

## Key Application Files
- `cmd/server/main.go` - Main server entry point
- `cmd/seed/main.go` - Database seeding utility
- `internal/app/app.go` - Application setup and configuration
- `internal/server/server.go` - HTTP server implementation
- `internal/handlers/` - HTTP request handlers
- `internal/db/` - Database layer and migrations
- `internal/models/` - Domain models
- `internal/views/` - Template views (Templ)

## Configuration & Build
- `Makefile` - Build and development commands
- `go.mod` - Go module dependencies
- `sqlc.yaml` - SQLC configuration for SQL code generation
- `railway.toml` - Railway deployment configuration
- `package.json` - Node.js dependencies (for frontend tooling)

## Documentation
- `README.md` - Project overview and setup instructions
- `docs/railway-deployment-setup.md` - Railway deployment guide

## Scripts
- `scripts/hard-restart.sh` - Hard restart script for development

## Testing
Use `make test` to run all tests. Key test files are located alongside their implementation files with `_test.go` suffix.

## Development
- Run `make dev` to start development server with Air hot reloading
- Run `make dev-restart` to kill existing server and restart with hot reload
- Run `make generate` to generate code from SQL queries (Templ + SQLC)
- Run `make build-prod` to build production binary

## Architecture Overview
This is a Go web application using:
- **Framework**: Standard library HTTP server with Chi router
- **Database**: SQLite with SQLC for type-safe queries
- **Templates**: Templ for HTML templating
- **Frontend**: HTMX + Alpine.js for interactivity
- **Deployment**: Railway platform
- **CI/CD**: GitHub Actions

## Database
- SQLite database (`health-monitor.db`)
- Migrations in `internal/db/migrations/`
- Generated queries in `internal/db/queries.sql.go`
- Models in `internal/db/models.go`

## Common Commands
- `make test` - Run all tests
- `make test-coverage` - Generate test coverage report
- `make lint` - Run linter
- `make install-tools` - Install development tools
- `make migrate-up` - Run database migrations
- `make seed` - Seed database with test data

## Environment Variables
Check `internal/config/config.go` for configuration options. Set environment variables or use command-line flags.

### Development Environment Variables
- `PORT` - Server port (default: 3000)
- `DB_PATH` - Database file path (default: "health-monitor.db")
- `LOG_LEVEL` - Logging level (default: "debug")

### Production Environment Variables (Railway)
- `PORT` - Server port (automatically provided by Railway)
- `DB_PATH` - Database file path (set to "/data/health-monitor.db" for volume persistence)
- `LOG_LEVEL` - Logging level (set to "info" for production)
- `RAILWAY_ENVIRONMENT` - Railway environment indicator (automatically set by Railway)

## End of Task Workflow
When completing a task, follow these steps:
1. **Test**: Run `make test` to ensure all tests pass
2. **Lint**: Run `make lint` if available, or check code quality
3. **Build**: Run `make build-prod` to verify production build works
4. **Commit**: Create descriptive commit message following project conventions
5. **Push**: Push changes to feature branch
6. **PR**: Open pull request against main branch for review
7. **Approval**: Wait for approval before merging

Note: Always seek approval before creating commits and pull requests.