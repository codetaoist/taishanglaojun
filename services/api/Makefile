.PHONY: help build run test migrate-up migrate-down migrate-create clean

# Default target
help:
	@echo "Available commands:"
	@echo "  build         - Build the API service"
	@echo "  run           - Run the API service"
	@echo "  test          - Run tests"
	@echo "  migrate-up    - Run database migrations"
	@echo "  migrate-down  - Rollback the last migration"
	@echo "  migrate-create - Create a new migration (usage: make migrate-create NAME=migration_name)"
	@echo "  clean         - Clean build artifacts"

# Build the API service
build:
	go build -o bin/api ./cmd/api

# Run the API service
run:
	go run ./cmd/api

# Run tests
test:
	go test -v ./...

# Run database migrations
migrate-up:
	go run ./cmd/migrate -action=up

# Rollback the last migration
migrate-down:
	go run ./cmd/migrate -action=down

# Create a new migration
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: Migration name is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	go run ./cmd/migrate -action=create -name=$(NAME)

# Clean build artifacts
clean:
	rm -rf bin/