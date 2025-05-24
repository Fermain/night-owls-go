# Night Owls Go - Development Makefile

.PHONY: help build test seed seed-dev seed-reset seed-preview clean

# Default target
help:
	@echo "Night Owls Go - Available commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  build         Build the server binary"
	@echo "  run           Run the server in development mode"
	@echo ""
	@echo "Database & Seeding:"
	@echo "  seed          Seed the database with development data"
	@echo "  seed-reset    Reset and seed the database from scratch"
	@echo "  seed-preview  Preview what would be seeded (dry run)"
	@echo "  seed-test     Seed a test database (test-seed.db)"
	@echo ""
	@echo "Enhanced Seeding:"
	@echo "  seed-minimal  Create minimal seed (3 users only)"
	@echo "  seed-large    Seed database with 50 users"
	@echo "  seed-future   Seed with future bookings (next 30 days)"
	@echo "  seed-export   Seed and export data to JSON file"
	@echo "  seed-demo     Full demo (100 users, future bookings, export)"
	@echo ""
	@echo "Testing:"
	@echo "  test          Run all tests"
	@echo "  test-api      Run API integration tests"
	@echo "  test-service  Run service layer tests"
	@echo ""
	@echo "Development:"
	@echo "  clean         Clean build artifacts and test databases"
	@echo "  format        Format Go code"
	@echo "  lint          Run linter checks"

# Build targets
build:
	@echo "Building server..."
	go build -o server cmd/server/main.go

build-seed:
	@echo "Building seed command..."
	go build -o cmd/seed/seed cmd/seed/main.go

# Run targets
run: build
	@echo "Starting development server..."
	./server

# Database seeding targets
seed: build-seed
	@echo "Seeding database with development data..."
	./cmd/seed/seed

seed-reset: build-seed
	@echo "Resetting and seeding database..."
	./cmd/seed/seed --reset

seed-preview: build-seed
	@echo "Previewing seed data..."
	./cmd/seed/seed --dry-run

seed-test: build-seed
	@echo "Seeding test database..."
	./cmd/seed/seed --db "./test-seed.db" --reset

# Enhanced seeding targets
seed-large: build-seed
	@echo "Seeding database with 50 users..."
	./cmd/seed/seed --reset --users 50

seed-future: build-seed
	@echo "Seeding database with future bookings..."
	./cmd/seed/seed --reset --future-bookings

seed-export: build-seed
	@echo "Seeding and exporting data..."
	./cmd/seed/seed --reset --export "./seed-export.json"

seed-demo: build-seed
	@echo "Creating demo environment (100 users, future bookings, export)..."
	./cmd/seed/seed --reset --users 100 --future-bookings --export "./demo-data.json" --verbose

seed-minimal: build-seed
	@echo "Creating minimal seed (3 users only)..."
	./cmd/seed/seed --reset --users 3

# Testing targets
test:
	@echo "Running all tests..."
	go test ./... -v

test-api:
	@echo "Running API integration tests..."
	go test ./internal/api/... -v

test-service:
	@echo "Running service layer tests..."
	go test ./internal/service/... -v

test-with-coverage:
	@echo "Running tests with coverage..."
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Development tools
format:
	@echo "Formatting Go code..."
	go fmt ./...

lint:
	@echo "Running linter..."
	golangci-lint run --timeout=5m

vet:
	@echo "Running go vet..."
	go vet ./...

# Clean up
clean:
	@echo "Cleaning build artifacts..."
	rm -f server cmd/seed/seed
	rm -f test-seed.db
	rm -f coverage.out coverage.html
	rm -f seed-export.json demo-data.json
	@echo "Clean complete"

# Development workflow
dev-setup: build-seed seed-reset
	@echo "Development environment ready!"
	@echo ""
	@echo "Seeded users:"
	@echo "  Admin: Alice Admin (+27821234567)"
	@echo "  Admin: Bob Manager (+27821234568)"
	@echo "  Owls: Charlie, Diana, Eve, Frank, Grace, Henry"
	@echo "  Guests: Iris, Jack"
	@echo ""
	@echo "Run 'make run' to start the server"

# Advanced seeding
seed-custom:
	@echo "Usage: make seed-custom DB=path/to/db.db"
	@if [ -z "$(DB)" ]; then echo "Error: DB parameter required"; exit 1; fi
	./cmd/seed/seed --db "$(DB)" --reset 