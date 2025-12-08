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

.PHONY: all build clean test coverage lint fmt vet install help run

# Default target
all: fmt vet test build

## help: Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^##' $(MAKEFILE_LIST) | sed -e 's/## //' | column -t -s ':'

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@chmod +x $(BIN_DIR)/$(BINARY_NAME)
	@echo "Binary built: $(BIN_DIR)/$(BINARY_NAME)"

## run: Run the application
run:
	$(GO) run $(MAIN_PATH)

## install: Install the binary to $GOPATH/bin
install:
	$(GO) install $(LDFLAGS) $(MAIN_PATH)

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@$(GO) clean
	@echo "Clean complete"

## test: Run tests
test:
	$(GO) test -v ./...

## test-short: Run short tests
test-short:
	$(GO) test -short -v ./...

## coverage: Run tests with coverage
coverage:
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

## fmt: Format code
fmt:
	$(GO) fmt ./...

## vet: Run go vet
vet:
	$(GO) vet ./...

## lint: Run linters (requires golangci-lint)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install: https://golangci-lint.run/usage/install/"; \
	fi

## mod-tidy: Tidy go.mod
mod-tidy:
	$(GO) mod tidy

## mod-download: Download dependencies
mod-download:
	$(GO) mod download

## mod-verify: Verify dependencies
mod-verify:
	$(GO) mod verify

## docker-build: Build Docker image
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .

## deps: Install development dependencies
deps:
	@echo "Installing development dependencies..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
