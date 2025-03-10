.PHONY: all build run clean generate

all: generate build

build:
	go build -o bin/server cmd/server/main.go

run: generate
	go run cmd/server/main.go

clean:
	rm -rf bin/

generate:
	templ generate
	sqlc generate

install-tools:
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

tidy:
	go mod tidy

.PHONY: migrations
migrations:
	atlas migrate diff --env local

.PHONY: migrate
migrate:
	atlas migrate apply --env local
