# End-to-End Testing

This directory contains end-to-end tests for ccproxy. These tests validate the complete flow from client request through the proxy to the provider and back.

## Test Categories

1. **Basic Proxy Tests** - Simple request/response validation
2. **Authentication Tests** - API key and security validation
3. **Routing Tests** - Model routing and provider selection
4. **Streaming Tests** - SSE streaming functionality
5. **Error Handling Tests** - Error propagation and handling
6. **Transform Tests** - Request/response transformation

## Running Tests

```bash
# Run all e2e tests
go test ./tests/e2e/...

# Run with verbose output
go test -v ./tests/e2e/...

# Run specific test
go test -v ./tests/e2e/... -run TestBasicProxy
```

## Requirements

- ccproxy must be built and available in PATH
- Port 8080 must be available for test server
- Mock provider server runs on port 9090