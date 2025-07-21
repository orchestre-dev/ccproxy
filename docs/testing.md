# CCProxy Testing Documentation

## Overview

CCProxy includes a comprehensive testing framework designed to ensure reliability while preventing system resource exhaustion during test execution. This guide covers test modes, environment variables, safe execution practices, and troubleshooting.

## Table of Contents

- [Test Environment Variables](#test-environment-variables)
- [Test Mode vs Production](#test-mode-vs-production)
- [Safe Test Execution](#safe-test-execution)
- [Test Organization](#test-organization)
- [Running Tests](#running-tests)
- [Troubleshooting Guide](#troubleshooting-guide)
- [Best Practices](#best-practices)

## Test Environment Variables

CCProxy uses several environment variables to control test behavior and prevent resource exhaustion:

### Core Test Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CCPROXY_TEST_MODE` | `0` | Set to `1` to enable test mode. Disables background process spawning and other production behaviors. |
| `CCPROXY_SPAWN_DEPTH` | `0` | Tracks process spawning depth to prevent infinite loops. Maximum value is 10. |
| `CCPROXY_FOREGROUND` | `0` | Set to `1` to force foreground execution. Used in tests to prevent daemon mode. |

### Additional Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CCPROXY_VERSION` | (empty) | Override the version string reported by the server |
| `CCPROXY_HOST` | `127.0.0.1` | Host address to bind the server to |
| `CCPROXY_PORT` | `3456` | Port number to bind the server to |
| `CCPROXY_LOG` | `false` | Enable logging (`true`/`false`) |
| `CCPROXY_LOG_FILE` | (empty) | Path to the log file |
| `CCPROXY_API_KEY` | (empty) | API key for authentication |
| `CCPROXY_PROXY_URL` | (empty) | HTTP/HTTPS proxy URL for outbound connections |
| `GOMAXPROCS` | (system) | Limit Go runtime CPU usage during tests (recommended: `4`) |

### Environment Variable Validation

All environment variables are validated at startup:
- `CCPROXY_SPAWN_DEPTH`: Must be 0-10
- `CCPROXY_TEST_MODE`, `CCPROXY_FOREGROUND`: Must be `0` or `1`
- `CCPROXY_PORT`: Must be 1-65535
- `CCPROXY_PROXY_URL`: Must start with `http://` or `https://`

## Test Mode vs Production

### Test Mode (`CCPROXY_TEST_MODE=1`)

When test mode is enabled:
- **Process Spawning Disabled**: No background processes are created
- **Foreground Execution**: Server runs in foreground only
- **Resource Limits**: Stricter resource consumption controls
- **Simplified Lifecycle**: No daemon mode or process management
- **Deterministic Behavior**: More predictable for testing

### Production Mode (Default)

When running in production:
- **Background Processes**: Can spawn background daemon processes
- **Process Management**: Full lifecycle management with PID files
- **Auto-restart**: Can automatically restart on failures
- **Full Features**: All production features enabled

### Key Differences

| Feature | Test Mode | Production |
|---------|-----------|------------|
| Background spawning | Disabled | Enabled |
| Process forking | Disabled | Enabled |
| Resource limits | Strict | Normal |
| Daemon mode | Disabled | Available |
| PID file management | Simplified | Full |
| Signal handling | Basic | Complete |

## Safe Test Execution

### Using the Safe Test Scripts

CCProxy provides several test scripts designed to prevent system crashes:

#### 1. Basic Safe Test Runner (`test_safe.sh`)

```bash
#!/bin/bash
# Runs tests with safety measures
./test_safe.sh

# What it does:
# - Sets CCPROXY_TEST_MODE=1
# - Limits parallelism with GOMAXPROCS=4
# - Kills existing ccproxy processes
# - Runs tests in smaller batches
```

#### 2. Advanced Safe Test Runner (`scripts/test_safe.sh`)

```bash
# Run all tests with resource monitoring
./scripts/test_safe.sh all

# Run specific test suites
./scripts/test_safe.sh unit        # Unit tests only
./scripts/test_safe.sh integration # Integration tests
./scripts/test_safe.sh load       # Load tests (use cautiously)
./scripts/test_safe.sh bench      # Benchmarks

# Features:
# - Memory usage monitoring (4GB limit)
# - CPU usage limiting (75%)
# - Automatic process cleanup
# - Resource violation detection
# - Detailed logging
```

#### 3. Batch Test Runner (`run_tests_safe.sh`)

```bash
#!/bin/bash
# Runs tests package by package
./run_tests_safe.sh

# Executes tests in this order:
# 1. Unit tests (package by package)
# 2. Integration tests
# 3. E2E tests
# 4. Skips benchmark/load tests by default
```

### Manual Safe Test Execution

For manual test execution with safety measures:

```bash
# Set environment variables
export CCPROXY_TEST_MODE=1
export CCPROXY_SPAWN_DEPTH=1
export CCPROXY_FOREGROUND=1
export GOMAXPROCS=4

# Kill any existing processes
pkill -f "ccproxy start" || true

# Run tests with limited parallelism
go test -p 2 -timeout 5m ./internal/...

# Run integration tests one at a time
go test -p 1 -timeout 10m ./tests/integration
```

## Test Organization

```
ccproxy/
├── tests/
│   ├── unit/           # Unit tests (in each package)
│   ├── integration/    # Integration test suites
│   ├── e2e/           # End-to-end tests
│   ├── benchmark/     # Performance benchmarks
│   └── load/          # Load and stress tests
│
├── testing/           # Test utilities
│   ├── framework.go   # Test framework core
│   ├── fixtures.go    # Test data fixtures
│   ├── assertions.go  # Enhanced assertions
│   ├── mock_server.go # Mock HTTP server
│   └── mock_provider.go # Mock provider implementation
│
└── scripts/
    ├── test.sh        # Main test runner
    ├── test_safe.sh   # Safe test runner with monitoring
    └── test_fixed.sh  # Fixed test runner for CI
```

## Running Tests

### Quick Commands

```bash
# Run all tests safely
make test

# Run specific test types
make test-unit         # Unit tests only
make test-integration  # Integration tests
make test-e2e         # End-to-end tests
make test-benchmark   # Benchmarks
make test-load        # Load tests

# Generate coverage report
make test-coverage
```

### Advanced Test Execution

```bash
# Run tests for specific package
go test -v ./internal/router

# Run with race detection
go test -race ./...

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./tests/benchmark/

# Run specific test by name
go test -v -run TestSelectProvider ./internal/router

# Run tests with custom timeout
go test -timeout 30s ./...

# Run tests in short mode (skip slow tests)
go test -short ./...
```

## Troubleshooting Guide

### Common Test Failures

#### 1. System Resource Exhaustion

**Symptoms:**
- System becomes unresponsive
- "Too many open files" errors
- "Cannot allocate memory" errors
- System crashes or freezes

**Solutions:**
```bash
# Always use test mode
export CCPROXY_TEST_MODE=1

# Limit parallelism
go test -p 1 ./...  # Run tests sequentially
go test -p 2 ./...  # Run with limited parallelism

# Use safe test scripts
./test_safe.sh
./scripts/test_safe.sh all

# Increase file descriptor limit
ulimit -n 2048
```

#### 2. Infinite Process Spawning

**Symptoms:**
- Hundreds of ccproxy processes
- System runs out of PIDs
- Fork bombs

**Solutions:**
```bash
# Set spawn depth limit
export CCPROXY_SPAWN_DEPTH=1

# Kill existing processes
pkill -f ccproxy

# Use test mode to disable spawning
export CCPROXY_TEST_MODE=1
```

#### 3. Port Already in Use

**Symptoms:**
- "bind: address already in use" errors
- Tests fail to start server

**Solutions:**
```bash
# Find and kill process using the port
lsof -i :3456
kill -9 <PID>

# Or use a different port
export CCPROXY_PORT=3457
```

#### 4. Test Timeouts

**Symptoms:**
- Tests exceed timeout limits
- "panic: test timed out" messages

**Solutions:**
```bash
# Increase test timeout
go test -timeout 30m ./tests/integration

# Run problematic tests individually
go test -v -run TestSpecificTest -timeout 10m ./package

# Skip slow tests in CI
go test -short ./...
```

#### 5. Race Conditions

**Symptoms:**
- Intermittent test failures
- "WARNING: DATA RACE" messages

**Solutions:**
```bash
# Run with race detector
go test -race ./...

# Fix data races in code
# Use proper synchronization (mutexes, channels)
```

### Debugging Test Failures

#### 1. Enable Verbose Output

```bash
# Verbose test output
go test -v ./...

# With detailed logging
export CCPROXY_LOG=true
export CCPROXY_LOG_FILE=test.log
go test -v ./...
```

#### 2. Run Specific Failing Test

```bash
# Isolate failing test
go test -v -run TestFailingFunction ./package

# With additional debugging
go test -v -run TestFailingFunction -count=1 -timeout=5m ./package
```

#### 3. Check Test Logs

```bash
# Safe test runner creates logs
ls -la test-logs/
tail -f test-logs/unit_router_*.log

# Check for panics
grep -r "panic:" test-logs/

# Check for resource issues
grep -r "too many open files\|cannot allocate memory" test-logs/
```

#### 4. Monitor System Resources

```bash
# Watch process count
watch -n 1 'ps aux | grep ccproxy | wc -l'

# Monitor memory usage
top -o MEM

# Check file descriptors
lsof | grep ccproxy | wc -l
```

## Best Practices

### 1. Always Use Test Mode

```bash
# In test scripts
export CCPROXY_TEST_MODE=1

# In Go test setup
func TestMain(m *testing.M) {
    os.Setenv("CCPROXY_TEST_MODE", "1")
    os.Setenv("CCPROXY_SPAWN_DEPTH", "1")
    code := m.Run()
    os.Exit(code)
}
```

### 2. Clean Up Resources

```go
func TestWithServer(t *testing.T) {
    // Always use cleanup
    server := startTestServer()
    defer server.Close()
    
    // Or use t.Cleanup
    t.Cleanup(func() {
        server.Close()
    })
}
```

### 3. Limit Parallelism

```bash
# In Makefile
test:
    go test -p 2 ./...

# In CI
go test -p 1 -timeout 30m ./...
```

### 4. Use Test Helpers

```go
// Use the testing framework
import "github.com/yourusername/ccproxy/testing"

func TestAPI(t *testing.T) {
    // Use test context
    ctx := testing.NewTestContext(t)
    
    // Use mock server
    mock := testing.NewMockServer()
    defer mock.Close()
    
    // Use fixtures
    req := testing.GetRequestFixture("anthropic_messages")
}
```

### 5. Handle Timeouts Gracefully

```go
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Use context in operations
    result, err := operationWithContext(ctx)
    if err == context.DeadlineExceeded {
        t.Skip("Test timed out - skipping")
    }
}
```

### 6. Document Test Requirements

```go
func TestRequiresDocker(t *testing.T) {
    if !dockerAvailable() {
        t.Skip("Test requires Docker")
    }
    // Test implementation
}

func TestSlowOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping slow test in short mode")
    }
    // Test implementation
}
```

## Continuous Integration

### GitHub Actions Configuration

```yaml
env:
  CCPROXY_TEST_MODE: "1"
  CCPROXY_SPAWN_DEPTH: "1"
  CCPROXY_FOREGROUND: "1"
  GOMAXPROCS: "2"

jobs:
  test:
    steps:
      - name: Run Safe Tests
        run: |
          ./scripts/test_safe.sh unit
          ./scripts/test_safe.sh integration
```

### Local Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Set test environment
export CCPROXY_TEST_MODE=1
export GOMAXPROCS=2

# Run fast tests only
go test -short -p 2 ./...
```

## Summary

The CCProxy testing framework provides comprehensive tools and practices for safe test execution:

1. **Use Test Mode**: Always set `CCPROXY_TEST_MODE=1`
2. **Limit Resources**: Use safe test scripts and limit parallelism
3. **Clean Up**: Kill processes and clean resources
4. **Monitor**: Watch for resource exhaustion
5. **Debug**: Use verbose output and logs for troubleshooting

Following these guidelines ensures reliable test execution without system crashes or resource exhaustion.