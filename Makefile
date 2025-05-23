.PHONY: all build run clean generate test test-coverage dev dev-restart

all: generate build

build:
	go build -o bin/server cmd/server/main.go

run: generate
	go run cmd/server/main.go

clean:
	rm -rf bin/
	rm -rf tmp/
	rm -f coverage.out

generate:
	templ generate
	sqlc generate

dev-restart: clean
	lsof -i :3000 | awk 'NR!=1 {print $$2}' | xargs kill -9 2>/dev/null || true
	make dev

dev:
	air

test:
	go test -v ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

install-tools:
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/air-verse/air@latest

tidy:
	go mod tidy

.PHONY: migrations
migrations:
	atlas migrate diff --env local

.PHONY: migrate
migrate:
	atlas migrate apply --env local

dev:
	air
