# CCProxy Makefile
# Automated build and distribution system

# Variables
BINARY_NAME := ccproxy
PACKAGE := github.com/orchestre-dev/ccproxy
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go build flags
LDFLAGS := -ldflags "\
	-X main.Version=${VERSION} \
	-X main.BuildTime=${BUILD_TIME} \
	-X main.Commit=${COMMIT} \
	-s -w"

# Directories
BUILD_DIR := build
DIST_DIR := dist
COVERAGE_DIR := coverage

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOLINT := golangci-lint

# Platforms
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

# Default target
.DEFAULT_GOAL := help

# Phony targets
.PHONY: all build test clean help install lint fmt coverage release docker

## help: Show this help message
help:
	@echo "CCProxy Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## all: Build for all platforms
all: clean test build-all

## build: Build binary for current platform
build:
	@echo "Building $(BINARY_NAME) v$(VERSION) for current platform..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/ccproxy

## build-all: Build binaries for all platforms
build-all:
	@echo "Building $(BINARY_NAME) v$(VERSION) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d/ -f1) \
		GOARCH=$$(echo $$platform | cut -d/ -f2) \
		output=$(BUILD_DIR)/$(BINARY_NAME)-$$(echo $$platform | tr / -); \
		if [ "$$(echo $$platform | cut -d/ -f1)" = "windows" ]; then \
			output="$$output.exe"; \
		fi; \
		echo "Building for $$platform..."; \
		GOOS=$$(echo $$platform | cut -d/ -f1) \
		GOARCH=$$(echo $$platform | cut -d/ -f2) \
		CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $$output ./cmd/ccproxy || exit 1; \
	done
	@echo "Build complete!"

## install: Install binary to system
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) installed to /usr/local/bin/"

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -timeout 5m ./...

## test-short: Run short tests
test-short:
	@echo "Running short tests..."
	$(GOTEST) -v -short -race ./...

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -race -tags=integration ./tests/integration/...

## coverage: Generate test coverage report
coverage:
	@echo "Generating coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated at $(COVERAGE_DIR)/coverage.html"

## lint: Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		$(GOLINT) run ./...; \
	else \
		echo "golangci-lint not installed. Install with: brew install golangci-lint"; \
		exit 1; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	$(GOMOD) tidy

## clean: Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR) $(COVERAGE_DIR)
	@rm -f *.out *.test *.coverage
	@echo "Clean complete!"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

## update-deps: Update dependencies
update-deps:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

## release: Create release artifacts
release: clean test build-all
	@echo "Creating release artifacts for v$(VERSION)..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		binary=$(BUILD_DIR)/$(BINARY_NAME)-$$(echo $$platform | tr / -); \
		if [ "$$(echo $$platform | cut -d/ -f1)" = "windows" ]; then \
			binary="$$binary.exe"; \
		fi; \
		if [ -f "$$binary" ]; then \
			archive=$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-$$(echo $$platform | tr / -); \
			if [ "$$(echo $$platform | cut -d/ -f1)" = "windows" ]; then \
				zip -j $$archive.zip $$binary README.md LICENSE 2>/dev/null || true; \
				echo "Created $$archive.zip"; \
			else \
				tar -czf $$archive.tar.gz -C $(BUILD_DIR) $$(basename $$binary) -C .. README.md LICENSE 2>/dev/null || true; \
				echo "Created $$archive.tar.gz"; \
			fi; \
		fi; \
	done
	@echo "Release artifacts created in $(DIST_DIR)/"

## docker: Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t ccproxy:$(VERSION) -t ccproxy:latest .

## docker-push: Push Docker image to registry
docker-push: docker
	@echo "Pushing Docker image..."
	docker tag ccproxy:$(VERSION) ghcr.io/orchestre-dev/ccproxy:$(VERSION)
	docker tag ccproxy:latest ghcr.io/orchestre-dev/ccproxy:latest
	docker push ghcr.io/orchestre-dev/ccproxy:$(VERSION)
	docker push ghcr.io/orchestre-dev/ccproxy:latest

## run: Run ccproxy locally
run: build
	@echo "Running $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME) start

## dev: Run in development mode with hot reload
dev:
	@echo "Running in development mode..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		exit 1; \
	fi

## check: Run all checks (test, lint, fmt)
check: fmt lint test

## info: Show build information
info:
	@echo "Build Information:"
	@echo "  Binary Name:  $(BINARY_NAME)"
	@echo "  Package:      $(PACKAGE)"
	@echo "  Version:      $(VERSION)"
	@echo "  Commit:       $(COMMIT)"
	@echo "  Build Time:   $(BUILD_TIME)"
	@echo "  Go Version:   $(shell go version)"
	@echo "  Platform:     $(shell go env GOOS)/$(shell go env GOARCH)"

## version: Show version
version:
	@echo "$(VERSION)"

# Docker targets (if Makefile.docker exists)
-include Makefile.docker