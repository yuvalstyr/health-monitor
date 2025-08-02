.PHONY: all build build-prod run clean generate test test-coverage dev dev-restart migrate-status migrate-up migrate-down migrate-create migrate-reset seed railway-build railway-start

all: generate build

build:
	go build -o bin/server cmd/server/main.go

# Production build with optimizations (local build for testing)
build-prod: generate
	go build -ldflags="-w -s" -o bin/server cmd/server/main.go

# Railway-specific build target (called by Railway)
railway-build:
	@echo "==============================================="
	@echo "� GRAILWAY BUILD STARTING"
	@echo "==============================================="
	@echo "📋 GIT INFORMATION:"
	@echo "  Branch: $$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')"
	@echo "  Commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "  Author: $$(git log -1 --pretty=format:'%an' 2>/dev/null || echo 'unknown')"
	@echo "  Message: $$(git log -1 --pretty=format:'%s' 2>/dev/null || echo 'unknown')"
	@echo "  Date: $$(git log -1 --pretty=format:'%ci' 2>/dev/null || echo 'unknown')"
	@echo "==============================================="
	echo "📦 INSTALLING REQUIRED TOOLS..."
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	echo "� GENERATINnG TEMPLATE AND DATABASE CODE..."
	@GOBIN=$$(go env GOBIN); if [ -z "$$GOBIN" ]; then GOBIN=$$(go env GOPATH)/bin; fi; $$GOBIN/templ generate
	@GOBIN=$$(go env GOBIN); if [ -z "$$GOBIN" ]; then GOBIN=$$(go env GOPATH)/bin; fi; $$GOBIN/sqlc generate
	echo "🏗️ BUILDING GO APPLICATION..."
	mkdir -p bin
	echo "  Working directory: $$(pwd)"
	echo "  Checking main.go: $$(ls -la cmd/server/main.go 2>/dev/null || echo 'not found')"
	go build -ldflags="-w -s" -o bin/server ./cmd/server
	@echo "==============================================="
	@echo "✅ RAILWAY BUILD COMPLETED SUCCESSFULLY"
	@echo "==============================================="

# Railway-specific start target
railway-start:
	@echo "🚀 Starting health-monitor server..."
	@echo "📋 Deployment Information:"
	@echo "  Branch: $$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')"
	@echo "  Commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "  Author: $$(git log -1 --pretty=format:'%an' 2>/dev/null || echo 'unknown')"
	@echo "  Build Time: $$(date -u '+%Y-%m-%d %H:%M:%S UTC')"
	@echo "🌍 Environment Variables:"
	@echo "  RAILWAY_ENVIRONMENT=$(RAILWAY_ENVIRONMENT)"
	@echo "  PORT=$(PORT)"
	@echo "  DB_PATH=$(DB_PATH)"
	@echo "  LOG_LEVEL=$(LOG_LEVEL)"
	@echo "📁 Data directory setup:"
	@mkdir -p /data || echo "  ⚠️ Warning: Could not create /data directory (this is normal in some environments)"
	@ls -la /data 2>/dev/null || echo "  📂 /data directory status: not accessible or doesn't exist yet"
	@echo "🎯 Starting server binary..."
	@echo "  Binary: ./bin/server"
	@echo "  Ready to serve requests!"
	./bin/server

run: generate
	go run cmd/server/main.go

clean:
	rm -rf bin/
	rm -rf tmp/
	rm -f coverage.out

generate:
	@GOBIN=$$(go env GOBIN); if [ -z "$$GOBIN" ]; then GOBIN=$$(go env GOPATH)/bin; fi; $$GOBIN/templ generate
	@GOBIN=$$(go env GOBIN); if [ -z "$$GOBIN" ]; then GOBIN=$$(go env GOPATH)/bin; fi; $$GOBIN/sqlc generate

dev-restart: clean
	lsof -i :3000 | awk 'NR!=1 {print $2}' | xargs kill -9 2>/dev/null || true
	make dev

dev:
	$(HOME)/go/bin/air

test:
	go test -v -race -timeout 10m -parallel 4 -count=1 ./...

test-short:
	go test -v -short -timeout 2m -parallel 4 -count=1 ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

install-tools:
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/air-verse/air@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest

tidy:
	go mod tidy

migrate-status:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations sqlite3 health-monitor.db status

migrate-up:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations sqlite3 health-monitor.db up

migrate-down:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations sqlite3 health-monitor.db down

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-create NAME=your_migration_name"; \
		exit 1; \
	fi
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations create $(NAME) sql

migrate-reset:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations sqlite3 health-monitor.db reset

# Database seeding for development
seed:
	go run cmd/seed/main.go -db health-monitor.db

# Production deployment helpers
deploy-check:
	@echo "Checking production build requirements..."
	@command -v go >/dev/null 2>&1 || { echo "Go is required but not installed. Aborting." >&2; exit 1; }
	@echo "✓ Go is installed"
	@make generate >/dev/null 2>&1 && echo "✓ Code generation successful" || { echo "✗ Code generation failed" >&2; exit 1; }
	@make test-short >/dev/null 2>&1 && echo "✓ Tests passing" || { echo "✗ Tests failing" >&2; exit 1; }
	@echo "✓ Ready for production deployment"

# Clean production artifacts
clean-prod:
	rm -rf bin/
	rm -rf tmp/
	rm -f coverage.out
	rm -f *.log