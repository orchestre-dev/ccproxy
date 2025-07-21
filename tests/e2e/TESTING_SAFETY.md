# End-to-End Test Safety Guide

This document describes the safety measures implemented in the CCProxy end-to-end tests to prevent process leaks and ensure clean test execution.

## Safety Features

### 1. Test Mode Environment Variables

The tests set several environment variables to prevent issues:

- `CCPROXY_TEST_MODE=1` - Disables background process spawning
- `CCPROXY_SPAWN_DEPTH=1` - Prevents infinite process spawning
- `CCPROXY_FOREGROUND=1` - Forces foreground execution

### 2. Process Isolation

Each test runs with:
- Isolated HOME directory (using `t.TempDir()`)
- Isolated PID files in temporary directories
- Separate configuration files per test

### 3. Automatic Cleanup

Multiple layers of cleanup ensure no processes are left running:

1. **Test-level cleanup** - Each test registers cleanup with `t.Cleanup()`
2. **TestMain cleanup** - Runs before and after all tests
3. **Signal handling** - Catches interrupts for graceful shutdown
4. **PID file cleanup** - Removes PID and lock files

### 4. Process Management

- Graceful shutdown with SIGTERM before SIGKILL
- Timeout-based process termination
- PID file tracking for precise process management
- Cleanup of lock files to prevent deadlocks

## Running Tests Safely

### Option 1: Use the Safety Script (Recommended)

```bash
./scripts/test_e2e_safe.sh
```

This script:
- Cleans up any existing processes before tests
- Sets proper environment variables
- Runs tests with timeout protection
- Performs final cleanup

### Option 2: Run Tests Directly

```bash
# Set test mode
export CCPROXY_TEST_MODE=1

# Run tests with limited parallelism
go test -v ./tests/e2e/... -count=1 -parallel=1
```

### Option 3: Run Individual Tests

```bash
# Run a specific test
go test -v ./tests/e2e -run TestBasicProxy
```

## Troubleshooting

### Check for Running Processes

```bash
# Check if any ccproxy processes are running
ps aux | grep ccproxy | grep -v grep

# Check specific ports
lsof -i :3456  # Default ccproxy port
```

### Manual Cleanup

If tests fail to clean up properly:

```bash
# Kill test processes
pkill -f "ccproxy.*--foreground"

# Remove PID files
rm -f ~/.ccproxy/.ccproxy.pid*
```

### Debug Failed Tests

Tests log output when they fail:
- Process stdout/stderr is captured
- Log files are created in temp directories
- Cleanup operations are logged

## Test Architecture

### Test Environment (`testEnv`)

Each test creates an isolated environment with:
- Temporary directories
- Isolated HOME directory
- Mock provider instance
- CCProxy instance

### Mock Provider

- Graceful shutdown with context timeout
- Proper cleanup in `t.Cleanup()`
- Non-blocking error reporting

### Port Management

- Dynamic port allocation using `GetFreePorts()`
- Prevents port conflicts between parallel tests
- Waits for ports to be released after cleanup

## Best Practices

1. **Always use `newTestEnv()`** for test isolation
2. **Don't share state** between tests
3. **Use sub-tests** with `t.Run()` for organization
4. **Check test logs** if cleanup issues occur
5. **Run with `-parallel=1`** if experiencing issues

## Environment Variables Reference

| Variable | Purpose | Default |
|----------|---------|---------|
| `CCPROXY_TEST_MODE` | Enables test mode | `0` |
| `CCPROXY_SPAWN_DEPTH` | Tracks spawn depth | `0` |
| `CCPROXY_FOREGROUND` | Forces foreground mode | `0` |
| `HOME` | Isolated home directory | System default |

## Known Issues and Solutions

1. **Port already in use**
   - Solution: Wait longer between tests or use dynamic ports
   
2. **PID file locked**
   - Solution: Remove `.lock` files manually
   
3. **Process won't die**
   - Solution: Use `kill -9` as last resort

## Contributing

When adding new e2e tests:
1. Use the `testEnv` helper for isolation
2. Register cleanup with `t.Cleanup()`
3. Handle errors gracefully
4. Log important operations for debugging