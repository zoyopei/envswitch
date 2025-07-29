# EnvSwitch Makefile

.PHONY: all build test clean install help

# Build variables
BINARY_NAME=envswitch
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Directories
BUILD_DIR=dist
COVERAGE_DIR=coverage

all: build

## Build: Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) .

## Test: Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

## Test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	@go test -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

## Lint: Run linting
lint:
	@echo "Running linter..."
	@golangci-lint run

## Format: Format code
format:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

## Clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@go clean

## Install: Install the application
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BINARY_NAME) $(GOPATH)/bin/

## Cross-compile: Build for multiple platforms
cross-compile: clean
	@echo "Cross-compiling..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	# Windows
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	
	# macOS
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	
	@echo "Cross-compilation completed. Binaries are in $(BUILD_DIR)/"

## Release: Create release archives
release: cross-compile
	@echo "Creating release archives..."
	@cd $(BUILD_DIR) && \
	tar -czf $(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
	tar -czf $(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
	tar -czf $(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
	tar -czf $(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
	zip $(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	
	@cd $(BUILD_DIR) && sha256sum *.tar.gz *.zip > checksums.txt
	@echo "Release archives created in $(BUILD_DIR)/"

## Deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify

## Tidy: Clean up dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

## Dev: Start development server
dev:
	@echo "Starting development server..."
	@go run . server --port 8080

## Version: Show version
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"

## Help: Show this help message
help:
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST) 