---
title: Testing Guide - CCProxy Test Framework
description: Comprehensive testing guide for CCProxy. Learn how to write, run, and debug tests for the ccproxy codebase.
keywords: CCProxy testing, test framework, unit tests, integration tests, benchmark tests, load testing, test coverage
---

# Testing Guide

<SocialShare />

CCProxy includes a comprehensive testing framework to ensure reliability and performance. This guide covers all aspects of testing in the project.

## Test Structure

CCProxy organizes tests into several categories:

```
tests/
├── unit/           # Unit tests (in each package)
├── integration/    # Integration test suites
├── benchmark/      # Performance benchmarks
└── load/          # Load and stress tests

testing/           # Testing utilities and helpers
├── framework.go   # Test framework core
├── fixtures.go    # Test data fixtures
├── assertions.go  # Enhanced assertions
└── mock_server.go # Mock HTTP server
```

## Running Tests

### Quick Start

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run benchmarks
make test-benchmark

# Run load tests
make test-load

# Run with coverage
make test-coverage
```

### Using the Test Script

The `scripts/test.sh` script provides advanced testing capabilities:

```bash
# Run all tests with coverage
./scripts/test.sh all

# Run specific test suite
./scripts/test.sh unit
./scripts/test.sh integration
./scripts/test.sh benchmark
./scripts/test.sh load

# Run tests for specific package
./scripts/test.sh package internal/router

# Run race detection tests
./scripts/test.sh race

# Run short tests (for CI)
./scripts/test.sh short

# Enable verbose output
VERBOSE=true ./scripts/test.sh unit
```

## Writing Tests

### Unit Tests

Unit tests should be placed alongside the code they test:

```go
// internal/router/router_test.go
package router

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestSelectProvider(t *testing.T) {
    t.Run("explicit model selection", func(t *testing.T) {
        router := New(config)
        provider, model := router.SelectProvider(&Request{
            Model: "openai:gpt-4",
        })
        
        assert.Equal(t, "openai", provider)
        assert.Equal(t, "gpt-4", model)
    })
}
```

### Integration Tests

Integration tests use the test framework for complex scenarios:

```go
// tests/integration/provider_test.go
package integration

import (
    "testing"
    "github.com/orchestre-dev/ccproxy/testing"
)

func TestProviderFailover(t *testing.T) {
    // Create test context
    ctx := testing.NewTestContext(t)
    
    // Create mock providers
    primary := testing.NewMockServer()
    backup := testing.NewMockServer()
    
    // Configure primary to fail
    primary.SetError("POST", "/v1/messages", 
        errors.New("provider unavailable"))
    
    // Test failover behavior
    client := testing.NewHTTPClient(serverURL)
    resp, err := client.Request("POST", "/v1/messages", request)
    
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

### Benchmark Tests

Benchmark tests measure performance:

```go
// tests/benchmark/token_benchmark_test.go
func BenchmarkTokenCounting(b *testing.B) {
    counter := token.NewCounter()
    message := fixtures.GenerateLargeMessage(10000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = counter.CountTokens(message, "claude-3-sonnet")
    }
}
```

### Load Tests

Load tests simulate real-world traffic:

```go
// tests/load/endpoint_load_test.go
func TestLoadBasicEndpoint(t *testing.T) {
    config := testing.LoadTestConfig{
        Duration:        30 * time.Second,
        ConcurrentUsers: 50,
        RampUpTime:      5 * time.Second,
        RequestsPerUser: 100,
        ThinkTime:       100 * time.Millisecond,
    }
    
    loadTester := testing.NewLoadTester(framework, config)
    results := loadTester.Run(testFunc)
    
    // Assert performance metrics
    assert.Less(t, results.ErrorRate, 0.05) // < 5% errors
    assert.Greater(t, results.RequestsPerSec, 10.0)
}
```

## Test Utilities

### Fixtures

Pre-defined test data:

```go
fixtures := testing.NewFixtures()

// Get request fixtures
req, _ := fixtures.GetRequest("anthropic_messages")
req, _ := fixtures.GetRequest("openai_chat")

// Get response fixtures
resp, _ := fixtures.GetResponse("anthropic_messages")
resp, _ := fixtures.GetResponse("streaming_chunks")

// Generate test data
largeMsg := fixtures.GenerateLargeMessage(10000) // ~10k tokens
messages := fixtures.GenerateMessages(20)         // 20-turn conversation
```

### Mock Server

Create mock HTTP servers for testing:

```go
// Create mock server
mock := testing.NewMockServer()
defer mock.Close()

// Add routes
mock.AddRoute("POST", "/v1/messages", response, 200)

// Add streaming route
mock.AddStreamingRoute("POST", "/stream", chunks)

// Set delays
mock.SetDelay("POST", "/v1/messages", 100*time.Millisecond)

// Record requests
requests := mock.GetRequests()
```

### Assertions

Enhanced assertion helpers:

```go
assertions := testing.NewAssertions(t)

// JSON assertions
assertions.AssertJSONEqual(expected, actual)
assertions.AssertJSONPath(jsonStr, "result.id", "123")

// Eventually assertions
assertions.AssertEventually(func() bool {
    return server.IsReady()
}, 5*time.Second, 100*time.Millisecond)

// Duration assertions  
assertions.AssertDuration(actual, expected, 10*time.Millisecond)
```

## Test Coverage

### Generating Coverage Reports

```bash
# Generate coverage report
make coverage

# Or using the test script
./scripts/test.sh coverage

# View HTML report
open coverage/coverage.html
```

### Coverage Requirements

- Unit tests: Minimum 80% coverage
- Integration tests: Cover all major workflows
- Critical packages: 90%+ coverage required

### Checking Coverage by Package

```bash
# Show coverage by package
go test -cover ./...

# Detailed coverage for specific package
go test -coverprofile=coverage.out ./internal/router
go tool cover -func=coverage.out
```

## Performance Testing

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./tests/benchmark/...

# Run specific benchmark
go test -bench=BenchmarkTokenCounting ./tests/benchmark/

# Run with memory profiling
go test -bench=. -benchmem ./tests/benchmark/

# Run for specific duration
go test -bench=. -benchtime=10s ./tests/benchmark/
```

### Interpreting Results

```
BenchmarkTokenCounting/Small-100-8   	   50000	     23456 ns/op	    2048 B/op	      10 allocs/op
BenchmarkTokenCounting/Large-10000-8 	     500	   2345678 ns/op	  204800 B/op	     100 allocs/op

Legend:
- Small-100-8: Test name with 8 CPU cores
- 50000: Number of iterations
- 23456 ns/op: Nanoseconds per operation
- 2048 B/op: Bytes allocated per operation
- 10 allocs/op: Number of allocations per operation
```

### Load Testing

```bash
# Run basic load test
./scripts/test.sh load

# Run sustained load test
go test -v -timeout=30m ./tests/load -run TestLoadSustained
```

## Test Best Practices

### 1. Test Organization

- Keep unit tests next to the code
- Use table-driven tests for multiple scenarios
- Group related tests with subtests

```go
func TestRouter(t *testing.T) {
    tests := []struct {
        name     string
        input    Request
        expected string
    }{
        {"explicit model", Request{Model: "gpt-4"}, "openai"},
        {"default model", Request{}, "anthropic"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 2. Test Isolation

- Each test should be independent
- Use setup and teardown properly
- Clean up resources

```go
func TestWithServer(t *testing.T) {
    // Setup
    server := StartTestServer()
    defer server.Close() // Always cleanup
    
    // Test implementation
}
```

### 3. Mock External Dependencies

```go
// Use interfaces for mockability
type Provider interface {
    SendRequest(req Request) (Response, error)
}

// Create mock implementation
type MockProvider struct {
    mock.Mock
}

func (m *MockProvider) SendRequest(req Request) (Response, error) {
    args := m.Called(req)
    return args.Get(0).(Response), args.Error(1)
}
```

### 4. Test Error Cases

Always test error conditions:

```go
t.Run("handles provider error", func(t *testing.T) {
    mock.SetError("POST", "/api", errors.New("provider error"))
    
    resp, err := client.SendRequest(request)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "provider error")
})
```

## Debugging Tests

### Verbose Output

```bash
# Run with verbose output
go test -v ./internal/router

# Run specific test with verbose
go test -v -run TestSelectProvider ./internal/router
```

### Debug Logging

```go
// Add debug logging in tests
t.Logf("Router decision: provider=%s, model=%s", provider, model)

// Enable debug logging
os.Setenv("LOG_LEVEL", "debug")
```

### Test Timeouts

```go
// Set custom timeout for slow tests
func TestSlowOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping slow test in short mode")
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Test implementation
}
```

## Continuous Integration

### GitHub Actions

Tests run automatically on:
- Pull requests
- Pushes to main branch
- Release tags

### Pre-commit Hooks

Run tests before committing:

```bash
# Install pre-commit hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
make test-short
EOF
chmod +x .git/hooks/pre-commit
```

## Troubleshooting Tests

### Common Issues

**Tests failing locally but passing in CI:**
- Check environment variables
- Verify timezone settings
- Check file permissions

**Flaky tests:**
- Add retry logic for network operations
- Use eventually assertions for async operations
- Increase timeouts for slow operations

**Race conditions:**
```bash
# Run with race detector
go test -race ./...
```

**Memory leaks:**
```bash
# Check for goroutine leaks
go test -run TestName -trace trace.out
go tool trace trace.out
```

## Next Steps

- [Development Setup](/guide/development) - Set up your development environment
- [Performance Tuning](/guide/performance) - Optimize ccproxy performance
- [Contributing](/guide/contributing) - Contribute to the project