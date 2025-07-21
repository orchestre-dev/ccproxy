---
title: Development Setup - CCProxy Development Guide
description: Complete development setup guide for CCProxy. Learn how to set up your development environment, build from source, and contribute to the project.
keywords: CCProxy development, development setup, build from source, Go development, contribution guide
---

# Development Setup

<SocialShare />

This guide will help you set up a development environment for CCProxy and start contributing to the project.

## Prerequisites

### Required Software

- **Go 1.21+** - [Download Go](https://golang.org/dl/)
- **Git** - [Download Git](https://git-scm.com/downloads)
- **Make** - Usually pre-installed on Linux/macOS
- **Docker** (optional) - [Download Docker](https://www.docker.com/get-started)

### Recommended Tools

- **VS Code** with Go extension
- **GoLand** IDE
- **golangci-lint** - For code linting
- **air** - For hot reload during development

## Getting Started

### 1. Fork and Clone

```bash
# Fork the repository on GitHub first

# Clone your fork
git clone https://github.com/YOUR_USERNAME/ccproxy.git
cd ccproxy

# Add upstream remote
git remote add upstream https://github.com/orchestre-dev/ccproxy.git

# Verify remotes
git remote -v
```

### 2. Install Dependencies

```bash
# Download Go modules
go mod download

# Verify dependencies
go mod verify

# Install development tools
make deps
```

### 3. Set Up Environment

Create a `.env` file for development:

```bash
# .env.development
PROVIDER=anthropic
ANTHROPIC_API_KEY=your_dev_key_here
LOG_LEVEL=debug
LOG_FORMAT=text
PORT=8080
```

### 4. Build and Run

```bash
# Build the binary
make build

# Run in development mode
make dev

# Or run directly
go run ./cmd/ccproxy start
```

## Project Structure

```
ccproxy/
├── cmd/ccproxy/        # Main application entry point
├── internal/           # Internal packages (not importable)
│   ├── auth/          # Authentication middleware
│   ├── config/        # Configuration management
│   ├── errors/        # Error handling
│   ├── logger/        # Logging system
│   ├── provider/      # Provider implementations
│   ├── router/        # Request routing
│   ├── server/        # HTTP server
│   ├── token/         # Token counting
│   └── transformer/   # Message transformation
├── testing/           # Test utilities
├── tests/            # Test suites
├── docs/             # VitePress documentation
├── scripts/          # Build and utility scripts
└── deployments/      # Deployment configurations
```

## Development Workflow

### 1. Create a Feature Branch

```bash
# Update main branch
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name
```

### 2. Make Changes

Follow the coding standards:

```go
// Package comment should be present
package router

import (
    "context"
    "fmt"
    
    // Group standard library imports
    // Then third-party imports
    // Then local imports
    "github.com/orchestre-dev/ccproxy/internal/config"
)

// RouterService handles request routing
type RouterService struct {
    config *config.Config
    // fields should be unexported unless necessary
}

// NewRouterService creates a new router service
func NewRouterService(cfg *config.Config) *RouterService {
    return &RouterService{
        config: cfg,
    }
}
```

### 3. Write Tests

Always include tests for new features:

```go
func TestNewRouterService(t *testing.T) {
    cfg := &config.Config{}
    router := NewRouterService(cfg)
    
    assert.NotNil(t, router)
    assert.Equal(t, cfg, router.config)
}
```

### 4. Run Tests and Linting

```bash
# Run tests
make test

# Run linter
make lint

# Run everything
make check
```

### 5. Commit Changes

Follow conventional commits:

```bash
# Format: <type>(<scope>): <subject>
git commit -m "feat(router): add provider selection logic"
git commit -m "fix(auth): handle empty API keys"
git commit -m "docs(api): update endpoint documentation"
git commit -m "test(router): add failover test cases"
```

Commit types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Build/tooling changes

### 6. Push and Create PR

```bash
# Push to your fork
git push origin feature/your-feature-name

# Create pull request on GitHub
```

## Development Tools

### Hot Reload

Use `air` for automatic reloading:

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with hot reload
air

# Or use make
make dev
```

Air configuration (`.air.toml`):

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/ccproxy"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "docs"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_error = true

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false
```

### Debugging

#### VS Code Configuration

`.vscode/launch.json`:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug ccproxy",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/ccproxy",
            "args": ["start"],
            "env": {
                "ANTHROPIC_API_KEY": "your_key",
                "LOG_LEVEL": "debug"
            }
        },
        {
            "name": "Debug current test",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${file}"
        }
    ]
}
```

#### GoLand Configuration

1. Create a Go Build configuration
2. Set program arguments: `start`
3. Set environment variables in the configuration
4. Enable "Run with debugger"

### Profiling

Profile CPU and memory usage:

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./tests/benchmark/
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./tests/benchmark/
go tool pprof mem.prof

# Trace execution
go test -trace=trace.out ./internal/router
go tool trace trace.out
```

## Code Style Guide

### General Guidelines

1. **Format**: Use `gofmt` (enforced by CI)
2. **Imports**: Group and sort imports
3. **Comments**: Add comments for all exported types and functions
4. **Error Handling**: Always handle errors explicitly
5. **Testing**: Maintain >80% test coverage

### Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("failed to process request: %w", err)
}

// Bad
if err != nil {
    return err
}
```

### Context Usage

Always accept context as first parameter:

```go
func ProcessRequest(ctx context.Context, req *Request) (*Response, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // Process request
    }
}
```

### Logging

Use structured logging:

```go
logger.Info("Processing request",
    "provider", provider,
    "model", model,
    "request_id", requestID,
)
```

## Building for Production

### Local Build

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build with specific version
VERSION=v1.2.3 make build
```

### Docker Build

```bash
# Build Docker image
make docker

# Build and push
make docker-push

# Multi-platform build
docker buildx build --platform linux/amd64,linux/arm64 -t ccproxy:latest .
```

### Release Build

```bash
# Create release artifacts
make release

# Files will be in dist/
ls -la dist/
```

## Testing Guidelines

### Unit Tests

```go
// Place tests in same package
package router

import "testing"

func TestSelectProvider(t *testing.T) {
    // Arrange
    router := New(testConfig)
    
    // Act
    provider, model := router.SelectProvider(request)
    
    // Assert
    assert.Equal(t, "expected", provider)
}
```

### Integration Tests

```go
// Place in tests/integration/
package integration

func TestEndToEndFlow(t *testing.T) {
    // Start test server
    server := StartTestServer(t)
    defer server.Close()
    
    // Test complete flow
}
```

### Benchmarks

```go
func BenchmarkTokenCounting(b *testing.B) {
    counter := NewCounter()
    text := "sample text"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        counter.Count(text)
    }
}
```

## Troubleshooting Development

### Common Issues

**Module errors:**
```bash
# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download
```

**Build errors:**
```bash
# Clean build cache
go clean -cache

# Rebuild
make clean build
```

**Test failures:**
```bash
# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestName ./package
```

### Getting Help

- Check existing [GitHub Issues](https://github.com/orchestre-dev/ccproxy/issues)
- Join [Discussions](https://github.com/orchestre-dev/ccproxy/discussions)
- Read the [Contributing Guide](/guide/contributing)

## Next Steps

- [Testing Guide](/guide/testing) - Learn about the testing framework
- [API Documentation](/api/) - Understand the API structure
- [Architecture](/api/architecture) - Learn about the system design