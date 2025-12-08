# =============================================================================
# KubeBridge-MCP Makefile
# =============================================================================
# This Makefile provides common development tasks for KubeBridge-MCP.
# Run 'make help' or just 'make' to see all available targets with descriptions.
#
# Quick Start:
#   make build          # Build the binary
#   make test           # Run tests
#   make check           # Run all checks (fmt, vet, lint, test)
#   make help            # Show this help message
# =============================================================================

# Project variables
BINARY_NAME=kubebridge-mcp
MAIN_PATH=./cmd/kubebridge-mcp
BIN_DIR=./bin
GO=go
GOFLAGS=-v

# Build information
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: all build clean test coverage lint fmt vet install help run mod-tidy mod-download mod-verify docker-build deps check test-short

# Default target - show help
.DEFAULT_GOAL := help

# =============================================================================
# Help Target
# =============================================================================

## help: Show this help message and exit
help:
	@echo ""
	@echo "╔══════════════════════════════════════════════════════════════════════════╗"
	@echo "║                    KubeBridge-MCP Makefile Help                          ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Available targets:"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed -e 's/## //' | awk 'BEGIN {FS = ": "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make build              # Build the binary to ./bin/kubebridge-mcp"
	@echo "  make test               # Run all tests"
	@echo "  make check              # Run all checks (fmt, vet, lint, test)"
	@echo "  make coverage           # Generate test coverage report"
	@echo "  make clean              # Remove build artifacts"
	@echo "  make docker-build       # Build Docker image"
	@echo ""
	@echo "For more information, see README.md or CONTRIBUTING.md"
	@echo ""

# =============================================================================
# Build Targets
# =============================================================================

## build: Build the binary to ./bin/kubebridge-mcp
##         Example: make build
##         Output: ./bin/kubebridge-mcp (executable)
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@chmod +x $(BIN_DIR)/$(BINARY_NAME)
	@echo "✅ Binary built: $(BIN_DIR)/$(BINARY_NAME)"
	@echo "   Run with: $(BIN_DIR)/$(BINARY_NAME) --help"

## run: Run the application directly (without building)
##      Example: make run
##      Note: This runs via 'go run', useful for quick testing
run: build
	@echo "Running $(BINARY_NAME)..."
	@echo "Note: MCP server expects JSON-RPC on stdin. Use 'make build' and configure MCP client instead."
	@$(BIN_DIR)/$(BINARY_NAME) --help || true

## install: Install the binary to $GOPATH/bin (makes it available system-wide)
##          Example: make install
##          Note: Requires $GOPATH/bin to be in your PATH
install:
	@echo "Installing $(BINARY_NAME) to $$GOPATH/bin..."
	$(GO) install $(LDFLAGS) $(MAIN_PATH)
	@echo "✅ Installed. Run with: $(BINARY_NAME) --help"

# =============================================================================
# Test Targets
# =============================================================================

## test: Run all tests with verbose output
##       Example: make test
##       Runs all tests in the project
test:
	@echo "Running tests..."
	$(GO) test -v ./...

## test-short: Run only short tests (skips long-running tests)
##             Example: make test-short
##             Useful for quick validation during development
test-short:
	@echo "Running short tests..."
	$(GO) test -short -v ./...

## coverage: Generate test coverage report
##           Example: make coverage
##           Output: coverage.out (text) and coverage.html (visual report)
coverage:
	@echo "Generating test coverage..."
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated:"
	@echo "   - Text: coverage.out"
	@echo "   - HTML: coverage.html (open in browser)"

# =============================================================================
# Code Quality Targets
# =============================================================================

## fmt: Format all Go code using gofmt
##      Example: make fmt
##      Automatically formats code to Go standards
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "✅ Code formatted"

## vet: Run go vet to check for common mistakes
##      Example: make vet
##      Checks for suspicious constructs, shadowed variables, etc.
vet:
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo "✅ go vet passed"

## lint: Run golangci-lint (comprehensive linter)
##       Example: make lint
##       Note: Requires golangci-lint (install with: make deps)
##       Checks code style, best practices, and potential bugs
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "✅ Linting passed"; \
	else \
		echo "❌ golangci-lint not installed."; \
		echo "   Install with: make deps"; \
		echo "   Or visit: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

# =============================================================================
# Dependency Management Targets
# =============================================================================

## mod-tidy: Clean up go.mod and go.sum files
##          Example: make mod-tidy
##          Removes unused dependencies and updates go.sum
mod-tidy:
	@echo "Tidying go.mod..."
	$(GO) mod tidy
	@echo "✅ Dependencies tidied"

## mod-download: Download all dependencies
##               Example: make mod-download
##               Downloads modules to local cache
mod-download:
	@echo "Downloading dependencies..."
	$(GO) mod download
	@echo "✅ Dependencies downloaded"

## mod-verify: Verify dependencies are valid and unchanged
##             Example: make mod-verify
##             Ensures module checksums match
mod-verify:
	@echo "Verifying dependencies..."
	$(GO) mod verify
	@echo "✅ Dependencies verified"

# =============================================================================
# Docker Targets
# =============================================================================

## docker-build: Build Docker image
##              Example: make docker-build
##              Example: make docker-build VERSION=v1.0.0
##              Builds Docker image with tag: kubebridge-mcp:$(VERSION)
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	@echo "✅ Docker image built: $(BINARY_NAME):$(VERSION)"
	@echo "   Run with: docker run $(BINARY_NAME):$(VERSION)"

# =============================================================================
# Development Setup Targets
# =============================================================================

## deps: Install development dependencies
##       Example: make deps
##       Installs golangci-lint for code linting
deps:
	@echo "Installing development dependencies..."
	@echo "Installing golangci-lint..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Development dependencies installed"
	@echo "   Run 'make lint' to use golangci-lint"

# =============================================================================
# Combined Targets
# =============================================================================

## check: Run all code quality checks (fmt, vet, lint, test)
##        Example: make check
##        Use this before committing code
##        Runs: formatting, vet, linting, and tests
check:
	@echo "╔══════════════════════════════════════════════════════════════════════════╗"
	@echo "║                    Running All Checks                                    ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@$(MAKE) fmt
	@echo ""
	@$(MAKE) vet
	@echo ""
	@$(MAKE) lint
	@echo ""
	@$(MAKE) test
	@echo ""
	@echo "╔══════════════════════════════════════════════════════════════════════════╗"
	@echo "║                    ✅ All Checks Passed                                 ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════╝"

## clean: Remove build artifacts and generated files
##        Example: make clean
##        Removes: ./bin directory, coverage files, and Go cache
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@$(GO) clean
	@echo "✅ Clean complete"

## all: Run fmt, vet, test, and build (full build pipeline)
##      Example: make all
##      Complete build and validation pipeline
all: fmt vet test build
	@echo "✅ Full build pipeline complete"
