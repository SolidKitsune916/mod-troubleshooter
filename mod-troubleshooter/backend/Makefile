.PHONY: build run dev test lint fmt clean help

# Variables
BINARY_NAME=server
BUILD_DIR=bin
MAIN_PATH=./cmd/server

# Build the server binary
build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Run the server
run:
	go run $(MAIN_PATH)

# Run with hot reload (requires air)
dev:
	@command -v air >/dev/null 2>&1 || { echo "Installing air..."; go install github.com/air-verse/air@latest; }
	air

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Lint code (requires golangci-lint)
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found. Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	golangci-lint run

# Vet code
vet:
	go vet ./...

# Format code
fmt:
	go fmt ./...
	@command -v goimports >/dev/null 2>&1 && goimports -w . || echo "goimports not installed, skipping"

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	go mod download
	go mod tidy

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the server binary"
	@echo "  run           - Run the server"
	@echo "  dev           - Run with hot reload (air)"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run golangci-lint"
	@echo "  vet           - Run go vet"
	@echo "  fmt           - Format code"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  help          - Show this help"
