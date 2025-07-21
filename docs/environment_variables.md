# CCProxy Environment Variables

This document describes all environment variables supported by CCProxy.

## Core Environment Variables

### CCPROXY_SPAWN_DEPTH
- **Description**: Tracks the depth of process spawning to prevent infinite loops
- **Required**: No
- **Default**: `0`
- **Valid Values**: Integer between 0 and 10
- **Usage**: Used internally to prevent infinite process spawning. Should not be set manually.

### CCPROXY_FOREGROUND
- **Description**: Set to '1' to indicate the process is running in foreground mode
- **Required**: No
- **Default**: `0`
- **Valid Values**: `0` or `1`
- **Usage**: Used internally to track foreground/background mode. Set automatically by the start command.

### CCPROXY_TEST_MODE
- **Description**: Set to '1' to enable test mode which disables background spawning
- **Required**: No
- **Default**: `0`
- **Valid Values**: `0` or `1`
- **Usage**: Should be set when running tests to prevent background process spawning.

### CCPROXY_VERSION
- **Description**: Override the version string reported by the server
- **Required**: No
- **Default**: Build-time version
- **Valid Values**: Any string
- **Usage**: Can be used to override the version reported in server responses.

## Server Configuration

### CCPROXY_HOST
- **Description**: Host address to bind the server to
- **Required**: No
- **Default**: `127.0.0.1`
- **Valid Values**: Any valid host address (e.g., `0.0.0.0`, `localhost`, `192.168.1.100`)
- **Usage**: Determines which network interface the server listens on.

### CCPROXY_PORT
- **Description**: Port number to bind the server to
- **Required**: No
- **Default**: `3456`
- **Valid Values**: Integer between 1 and 65535
- **Usage**: Determines which port the server listens on.

## Logging Configuration

### CCPROXY_LOG
- **Description**: Enable logging
- **Required**: No
- **Default**: `false`
- **Valid Values**: `true`, `false`, `1`, or `0`
- **Usage**: Enables or disables logging output.

### CCPROXY_LOG_FILE
- **Description**: Path to the log file
- **Required**: No
- **Default**: Empty (logs to stdout/stderr)
- **Valid Values**: Any valid file path
- **Usage**: When set, logs are written to the specified file instead of stdout/stderr.

## Authentication & Proxy

### CCPROXY_API_KEY
- **Description**: API key for authentication
- **Required**: No
- **Default**: Empty
- **Valid Values**: Any string
- **Usage**: When set, requires clients to provide this API key for authentication.

### CCPROXY_PROXY_URL
- **Description**: HTTP/HTTPS proxy URL for outbound connections
- **Required**: No
- **Default**: Empty
- **Valid Values**: URL starting with `http://` or `https://`
- **Usage**: Routes outbound connections through the specified proxy server.

## Provider-Specific Variables

### Provider Configuration (Docker/Kubernetes)
When running in containerized environments, additional provider configuration can be set:

- `CCPROXY_PROVIDERS_0_NAME`: Name of the first provider
- `CCPROXY_PROVIDERS_0_API_KEY`: API key for the first provider
- `CCPROXY_PROVIDERS_0_API_BASE_URL`: Base URL for the provider's API
- `CCPROXY_PROVIDERS_0_ENABLED`: Whether the provider is enabled

## Standard Proxy Variables

CCProxy also respects standard proxy environment variables:
- `HTTP_PROXY` / `http_proxy`: HTTP proxy for outbound connections
- `HTTPS_PROXY` / `https_proxy`: HTTPS proxy for outbound connections
- `NO_PROXY` / `no_proxy`: Comma-separated list of hosts to bypass proxy

## Example Usage

```bash
# Basic usage with custom port
export CCPROXY_PORT=8080
export CCPROXY_HOST=0.0.0.0
ccproxy start

# Enable logging to file
export CCPROXY_LOG=true
export CCPROXY_LOG_FILE=/var/log/ccproxy.log
ccproxy start

# Use with authentication
export CCPROXY_API_KEY=your-secret-key
ccproxy start

# Test mode (for running tests)
export CCPROXY_TEST_MODE=1
go test ./...

# Use with proxy
export CCPROXY_PROXY_URL=http://proxy.company.com:8080
ccproxy start
```

## Validation

All environment variables are validated when the service starts. Invalid values will cause the service to fail with a descriptive error message.

To see all environment variables and their current validation status, run:
```bash
ccproxy env
```