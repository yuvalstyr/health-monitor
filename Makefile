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
	@echo "🚀 RAILWAY BUILD STARTING"
	@echo "==============================================="
	@echo "📋 GIT INFORMATION:"
	@echo "  Branch: $${RAILWAY_GIT_BRANCH:-unknown}"
	@echo "  Commit: $${RAILWAY_GIT_COMMIT_SHA:-unknown}"
	@echo "  Author: $${RAILWAY_GIT_AUTHOR:-unknown}"
	@echo "  Message: $${RAILWAY_GIT_COMMIT_MESSAGE:-unknown}"
	@echo "  Date: $${RAILWAY_GIT_COMMIT_DATE:-unknown}"
	@echo "  Railway Environment: $(RAILWAY_ENVIRONMENT)"
	@echo "  Railway Service: $(RAILWAY_SERVICE_NAME)"
	@echo "==============================================="
	@echo "📦 INSTALLING REQUIRED TOOLS..."
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo "🔧 GENERATING TEMPLATE AND DATABASE CODE..."
	@GOBIN=$$(go env GOBIN); if [ -z "$$GOBIN" ]; then GOBIN=$$(go env GOPATH)/bin; fi; $$GOBIN/templ generate
	@GOBIN=$$(go env GOBIN); if [ -z "$$GOBIN" ]; then GOBIN=$$(go env GOPATH)/bin; fi; $$GOBIN/sqlc generate
	@echo "🏗️ BUILDING GO APPLICATION..."
	@mkdir -p bin
	@echo "  Working directory: $$(pwd)"
	@echo "  Checking main.go: $$(ls -la cmd/server/main.go 2>/dev/null || echo 'not found')"
	@echo "🔍 DEBUGGING GO ENVIRONMENT..."
	@echo "  GOROOT: $$(go env GOROOT)"
	@echo "  GOPATH: $$(go env GOPATH)"
	@echo "  GOMOD: $$(go env GOMOD)"
	@echo "  PWD: $$(pwd)"
	@echo "  Go version: $$(go version)"
	@ls -la go.* || echo "No go files found"
	@echo "🏗️ BUILDING SERVER WITH MODULE MODE..."
	GOWORK=off go build -ldflags="-w -s" -o bin/server ./cmd/server || \
	(echo "⚠️ FALLBACK: Building without workspace..." && \
	 cd cmd/server && go build -ldflags="-w -s" -o ../../bin/server .)
	@echo "🌱 BUILDING SEED UTILITY..."
	GOWORK=off go build -ldflags="-w -s" -o bin/seed ./cmd/seed || \
	(echo "⚠️ FALLBACK: Building seed without workspace..." && \
	 cd cmd/seed && go build -ldflags="-w -s" -o ../../bin/seed .)
	@echo "==============================================="
	@echo "✅ RAILWAY BUILD COMPLETED SUCCESSFULLY"
	@echo "==============================================="

# Railway-specific start target
railway-start:
	@echo "🚀 Starting health-monitor server..."
	@echo "📋 Deployment Information:"
	@echo "  Branch: $${RAILWAY_GIT_BRANCH:-unknown}"
	@echo "  Commit: $${RAILWAY_GIT_COMMIT_SHA:-unknown}"
	@echo "  Author: $${RAILWAY_GIT_AUTHOR:-unknown}"
	@echo "  Message: $${RAILWAY_GIT_COMMIT_MESSAGE:-unknown}"
	@echo "  Build Time: $$(date -u '+%Y-%m-%d %H:%M:%S UTC')"
	@echo "🌍 Environment Variables:"
	@echo "  RAILWAY_ENVIRONMENT=$(RAILWAY_ENVIRONMENT)"
	@echo "  PORT=$(PORT)"
	@echo "  DB_PATH=$(DB_PATH)"
	@echo "  LOG_LEVEL=$(LOG_LEVEL)"
	@echo "📁 Data directory setup:"
	@mkdir -p /data || echo "  ⚠️ Warning: Could not create /data directory (this is normal in some environments)"
	@ls -la /data 2>/dev/null || echo "  📂 /data directory status: not accessible or doesn't exist yet"
	@echo "🌱 Database seeding for development branches..."
	@echo "🔍 Debug: Checking branch information..."
	@echo "  RAILWAY_GIT_BRANCH: $${RAILWAY_GIT_BRANCH:-not_set}"
	@echo "  Git command result: $$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'git_failed')"
	@CURRENT_BRANCH=$${RAILWAY_GIT_BRANCH:-$$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')}; \
	echo "  📊 Determined branch: $$CURRENT_BRANCH"; \
	if [ "$$CURRENT_BRANCH" != "main" ] && [ "$$CURRENT_BRANCH" != "master" ] && [ "$$CURRENT_BRANCH" != "unknown" ] && [ "$$CURRENT_BRANCH" != "" ]; then \
		echo "  🌱 Seeding database for branch: $$CURRENT_BRANCH"; \
		echo "  📍 Database path: $(DB_PATH)"; \
		go run cmd/seed/main.go -db "$(DB_PATH)" || echo "  ⚠️ Seeding failed - continuing without seed data"; \
	else \
		echo "  ⏭️ Skipping seed - branch: $$CURRENT_BRANCH (production or unknown)"; \
	fi
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