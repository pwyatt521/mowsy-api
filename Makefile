# Mowsy API Makefile

.PHONY: help build test test-verbose test-coverage clean run-local run-lambda-local deps fmt lint swagger-init swagger-gen swagger-fmt swagger-serve swagger-docs

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the Lambda binary"
	@echo "  test          - Run all tests"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  run-local     - Run the API locally"
	@echo "  deps          - Download dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  swagger-init  - Install swagger CLI tool"
	@echo "  swagger-gen   - Generate swagger documentation"
	@echo "  swagger-fmt   - Format swagger comments"
	@echo "  swagger-serve - Generate docs and show integration instructions"
	@echo "  swagger-docs  - Generate docs and show file locations"

# Build the Lambda binary
build:
	GOOS=linux GOARCH=amd64 go build -o bootstrap cmd/lambda/main.go
	zip lambda-deployment.zip bootstrap

# Run all tests
test:
	JWT_SECRET=test-secret-key go test ./...

# Run tests with verbose output
test-verbose:
	JWT_SECRET=test-secret-key go test -v ./...

# Run tests with coverage
test-coverage:
	JWT_SECRET=test-secret-key go test -cover ./...
	JWT_SECRET=test-secret-key go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run specific test packages
test-auth:
	JWT_SECRET=test-secret-key go test -v ./pkg/auth

test-services:
	JWT_SECRET=test-secret-key go test -v ./internal/services

test-handlers:
	JWT_SECRET=test-secret-key go test -v ./internal/handlers

test-middleware:
	JWT_SECRET=test-secret-key go test -v ./internal/middleware

test-utils:
	JWT_SECRET=test-secret-key go test -v ./internal/utils

# Clean build artifacts
clean:
	rm -f bootstrap
	rm -f lambda-deployment.zip
	rm -f coverage.out
	rm -f coverage.html

# Run the API locally
run-local:
	go run cmd/local/main.go

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint to be installed)
lint:
	golangci-lint run

# Development helpers
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Docker commands (if you want to containerize)
docker-build:
	docker build -t mowsy-api .

docker-run:
	docker run -p 8080:8080 mowsy-api

# Database commands (for local development)
db-create:
	createdb mowsy_db

db-drop:
	dropdb mowsy_db

# Test database commands
test-db-setup:
	@echo "Test database uses in-memory SQLite, no setup required"

# Environment setup
env-example:
	cp .env.example .env
	@echo "Please edit .env with your actual configuration values"

# Build for different platforms
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/mowsy-api-linux cmd/local/main.go

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/mowsy-api-mac cmd/local/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/mowsy-api-windows.exe cmd/local/main.go

build-all: build-linux build-mac build-windows

# Performance testing
benchmark:
	go test -bench=. ./...

# Security scanning (requires gosec)
security:
	gosec ./...

# Install security scanner
install-security:
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Swagger documentation
swagger-init:
	go install github.com/swaggo/swag/cmd/swag@v1.8.12

swagger-gen:
	~/go/bin/swag init -g cmd/local/main.go -o docs

swagger-fmt:
	~/go/bin/swag fmt

swagger-serve:
	@./scripts/serve-swagger.sh

swagger-docs: swagger-gen
	@echo "Swagger documentation generated successfully!"
	@echo "View the generated files:"
	@echo "  - JSON: docs/swagger.json"
	@echo "  - YAML: docs/swagger.yaml"
	@echo "  - Go docs: docs/docs.go"
	@echo ""
	@echo "To integrate with your Go 1.22+ application:"
	@echo "  1. Uncomment swagger imports in internal/routes/routes.go"
	@echo "  2. Uncomment docs import in cmd/local/main.go"
	@echo "  3. Start server and visit http://localhost:8080/swagger/index.html"