# Integration Tests for CCProxy

This directory contains comprehensive integration tests for the CCProxy system.

## Test Categories

### 1. Server Integration Tests (`server_test.go`)
- Server lifecycle (start, run, shutdown)
- Health and status endpoints
- Provider configuration
- Authentication middleware
- CORS handling
- Graceful shutdown

### 2. CLI Integration Tests (`cli_test.go`)
- Command execution (start, stop, status, version)
- Claude configuration commands (init, show, set, reset)
- Service lifecycle management
- Configuration file loading
- Error handling for invalid inputs
- Help documentation

### 3. Pipeline Integration Tests (`pipeline_test.go`)
- Complete request pipeline processing
- Streaming response handling
- Routing logic with multiple providers
- Transformer chain execution
- Request/response transformation

### 4. Error Handling Tests (`error_handling_test.go`)
- Provider error responses (401, 429, 500, etc.)
- Request validation errors
- Provider unavailability
- Timeout handling
- Concurrent request processing

## Running the Tests

### Run all integration tests:
```bash
go test ./tests/integration/... -v
```

### Run specific test file:
```bash
go test ./tests/integration/server_test.go -v
```

### Run with coverage:
```bash
go test ./tests/integration/... -coverprofile=integration_coverage.out -v
go tool cover -html=integration_coverage.out
```

### Run specific test:
```bash
go test ./tests/integration/... -run TestServerLifecycle -v
```

## Environment Variables

Some tests require specific environment variables:

- `TEST_ANTHROPIC_API_KEY`: API key for testing with real Anthropic provider
- `CI`: Set to "true" to skip tests that require special permissions

## Test Requirements

1. **Network Access**: Tests create local HTTP servers on random ports
2. **File System**: Tests create temporary files and directories
3. **Process Management**: Some tests spawn child processes
4. **Build Tools**: Tests build the ccproxy binary

## Adding New Tests

When adding new integration tests:

1. Place them in the appropriate test file based on the component being tested
2. Use meaningful test names that describe what is being tested
3. Clean up any resources (files, processes, servers) created during tests
4. Use `t.Skip()` for tests that require special setup or environment
5. Document any special requirements or environment variables

## Test Isolation

Each test should:
- Use unique ports (port 0 for random assignment)
- Create temporary directories for file operations
- Clean up all resources in defer statements
- Not depend on the state from other tests

## Debugging Failed Tests

1. Run with `-v` flag for verbose output
2. Check for leftover processes: `ps aux | grep ccproxy`
3. Check for leftover files in temp directories
4. Look for port conflicts if tests fail intermittently
5. Use `t.Logf()` to add debug output

## Performance Considerations

- Integration tests are slower than unit tests
- Use `t.Parallel()` where possible for independent tests
- Mock external services instead of calling real APIs
- Set reasonable timeouts to prevent hanging tests