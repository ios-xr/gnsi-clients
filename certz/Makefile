# Makefile for gNSI Certz Client

.PHONY: all build clean deps lint fmt install help

# Default target
all: build

# Build the client
build:
	@echo "Building certz_client..."
	@cd cmd/certz_client && go build -o certz_client

# Build with version information
build-release:
	@echo "Building certz_client with version info..."
	@cd cmd/certz_client && go build -ldflags="-s -w" -o certz_client

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f cmd/certz_client/certz_client
	@go clean

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -s -w cmd/ pkg/
	@go vet ./...

# Install the client to GOPATH/bin
install:
	@echo "Installing certz_client..."
	@cd cmd/certz_client && go install

# Run a quick connection test (requires environment variables)
test-connection:
	@if [ -z "$$TARGET_ADDR" ]; then \
		echo "Error: TARGET_ADDR environment variable not set"; \
		echo "Example: make test-connection TARGET_ADDR=192.168.1.1:57400"; \
		exit 1; \
	fi
	@./cmd/certz_client/certz_client \
		-target_addr $$TARGET_ADDR \
		-username $$USERNAME \
		-password $$PASSWORD \
		-insecure_skip_verify \
		-op get-profile-list \
		-v

# Display help
help:
	@echo "Available targets:"
	@echo "  build           - Build the certz_client binary"
	@echo "  build-release   - Build optimized binary for release"
	@echo "  clean           - Remove build artifacts"
	@echo "  deps            - Download and tidy dependencies"
	@echo "  lint            - Run golangci-lint (requires installation)"
	@echo "  fmt             - Format code with gofmt"
	@echo "  install         - Install binary to GOPATH/bin"
	@echo "  test-connection - Test connection to device (requires env vars)"
	@echo "  help            - Display this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make fmt"
	@echo "  TARGET_ADDR=192.168.1.1:57400 USERNAME=admin PASSWORD=pass make test-connection"
	@echo ""
	@echo "Note: Tests not yet implemented"
