.PHONY: help build run test clean deps fmt lint

help:
	@echo "Release Control System - Backend"
	@echo ""
	@echo "Available targets:"
	@echo "  make build       - Build the application"
	@echo "  make run         - Run the application"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make deps        - Download dependencies"
	@echo "  make fmt         - Format code"
	@echo "  make lint        - Run linter (requires golangci-lint)"

build:
	@echo "Building release-control..."
	@go build -o release-control cmd/server/main.go
	@echo "Build complete: ./release-control"

run:
	@echo "Running release-control..."
	@go run cmd/server/main.go

dev:
	@echo "Running in development mode with auto-reload..."
	@command -v air >/dev/null 2>&1 || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	@air

test:
	@echo "Running tests..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean:
	@echo "Cleaning..."
	@rm -f release-control release-control* coverage.* *.out
	@rm -f release_control.db
	@echo "Clean complete"

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

lint:
	@echo "Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run ./...

.DEFAULT_GOAL := help
