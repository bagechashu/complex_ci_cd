#!/bin/bash

# Run all tests with coverage report
set -e

echo "Running unit tests..."
go test -v -race -tags=unit -cover -coverprofile=coverage-unit.out ./...

echo ""
echo "Running integration tests..."
go test -v -race -tags=integration -cover -coverprofile=coverage-integration.out ./...

echo ""
echo "Generating coverage report..."
go tool cover -html=coverage-unit.out -o coverage-unit.html
go tool cover -html=coverage-integration.out -o coverage-integration.html

echo ""
echo "Coverage reports generated:"
echo "  - coverage-unit.html"
echo "  - coverage-integration.html"

# Clean up
rm -f coverage-unit.out coverage-integration.out
