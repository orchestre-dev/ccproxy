.PHONY: build clean test install build-all

# Binary name
BINARY_NAME=ccproxy

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT) -s -w"

# Default build
build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/ccproxy

# Build for all platforms
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/ccproxy
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/ccproxy

build-darwin:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/ccproxy
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/ccproxy

build-windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/ccproxy

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Install locally
install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

# Run the application
run: build
	./$(BINARY_NAME)

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# Lint code
lint:
	golangci-lint run

# Tidy dependencies
tidy:
	go mod tidy

# Generate test coverage report
coverage: test
	go tool cover -html=coverage.out -o coverage.html