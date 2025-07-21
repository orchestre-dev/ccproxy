# CCProxy - Claude Code Router (Go Implementation)

## Project Overview

CCProxy is a high-performance Go implementation of the Claude Code Router, providing intelligent model routing, multi-provider support, and comprehensive request/response transformation for AI services. This is a complete rewrite from TypeScript to Go with enhanced security, performance, and reliability features.

## Architecture

### Core Components

- **CLI Interface** (`cmd/ccproxy/`) - Cobra-based command-line interface
- **Server** (`internal/server/`) - Gin-based HTTP server with middleware stack
- **Pipeline** (`internal/pipeline/`) - Request processing pipeline with transformations
- **Providers** (`internal/providers/`) - Multi-provider AI service integration
- **Transformers** (`internal/transformer/`) - Request/response transformation engine
- **Router** (`internal/router/`) - Intelligent model routing based on token count and parameters
- **Performance** (`internal/performance/`) - Monitoring, rate limiting, and resource management
- **Security** (`internal/security/`) - Authentication, authorization, and audit logging

### Key Features

- **Intelligent Routing**: Automatic model selection based on token count (>60K → longContext), model type, and thinking parameters
- **Multi-Provider Support**: Anthropic, OpenAI, Gemini, DeepSeek, OpenRouter, Groq
- **Streaming Support**: Server-Sent Events (SSE) for real-time responses
- **Process Management**: Background service with PID file locking and graceful shutdown
- **Claude Code Integration**: Auto-start, environment variable management, reference counting

## Security Features

### Authentication & Authorization
- API key validation (Bearer token and x-api-key header)
- Localhost-only enforcement when no API key configured
- IP-based access controls with whitelist/blacklist
- Health endpoints with graduated access (basic status public, details require auth)

### Security Hardening
- Request size limits (10MB default) to prevent DoS attacks
- Configurable timeouts (30 seconds default) instead of 60-minute hardcoded values
- Provider error response sanitization (removes API keys, tokens, emails)
- CORS headers sanitized (x-api-key removed from public headers)
- Resource limit enforcement with circuit breaker patterns

### Process Security
- Exclusive PID file locking to prevent multiple instances
- No unsafe fallbacks - fails fast if locking unavailable
- Atomic operations for metrics tracking to prevent race conditions

## Configuration

### Default Configuration
```json
{
  "host": "127.0.0.1",
  "port": 3456,
  "performance": {
    "request_timeout": "30s",
    "max_request_body_size": 10485760,
    "metrics_enabled": true,
    "rate_limit_enabled": false,
    "circuit_breaker_enabled": true
  }
}
```

### Environment Variables
- `CCPROXY_PORT` - Override default port
- `CCPROXY_HOST` - Override default host  
- `CCPROXY_API_KEY` - Set API key for authentication
- `CCPROXY_CONFIG` - Path to configuration file
- `LOG` - Enable file logging to ~/.ccproxy/ccproxy.log

## Commands

### Primary Commands
- `ccproxy start` - Start the service in background
- `ccproxy stop` - Stop the running service
- `ccproxy status` - Show service status with emoji indicators
- `ccproxy code` - Auto-start service and configure Claude Code environment
- `ccproxy claude` - Manage ~/.claude.json configuration
- `ccproxy env` - Show environment variable documentation
- `ccproxy version` - Show version information

### Command Options
- `--config` - Specify custom configuration file
- `--foreground` - Run service in foreground (for start command)

## Build & Development

### Building
```bash
make build              # Build for current platform
make build-all          # Build for all platforms
make docker-build       # Build Docker image
make test               # Run tests
make test-race          # Run tests with race detection
```

### Testing
- **Unit Tests**: Comprehensive coverage for all components
- **Integration Tests**: End-to-end request processing
- **Race Detection**: All tests pass with `-race` flag
- **Benchmark Tests**: Performance validation
- **Security Tests**: Authentication and authorization flows

### Dependencies
- **Go 1.21+** required
- **Gin** - HTTP web framework
- **Cobra** - CLI framework
- **Viper** - Configuration management
- **gofrs/flock** - File locking for PID management
- **tiktoken-go** - Token counting compatible with OpenAI

## Deployment

### Production Requirements
- Memory: <20MB baseline usage
- Startup: <100ms cold start
- Port: 3456 (configurable)
- PID File: ~/.ccproxy/.ccproxy.pid
- Logs: ~/.ccproxy/ccproxy.log (when enabled)

### Docker Deployment
```bash
docker build -t ccproxy .
docker run -p 3456:3456 ccproxy
```

### Binary Deployment
- Single static binary with no external dependencies
- Cross-platform: Linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64)
- Self-contained with embedded version and build information

## Claude Code Integration

### Auto-Start Behavior
When `ccproxy code` is executed:
1. Checks if service is already running
2. Auto-starts service if not running (10-second timeout)
3. Sets environment variables for Claude Code:
   - `ANTHROPIC_AUTH_TOKEN=test`
   - `ANTHROPIC_BASE_URL=http://127.0.0.1:3456`
   - `API_TIMEOUT_MS=600000`
4. Manages reference counting for auto-shutdown

### Status Indicators
Service status is displayed with emoji indicators:
- ✅ Running
- ❌ Not Running  
- 🆔 Process ID
- 🌐 Port
- 📡 API Endpoint

## Performance Monitoring

### Metrics Collection
- Request counts (total, success, failed)
- Latency tracking (average, P50, P95, P99)
- Provider-specific metrics
- Resource usage (memory, CPU, goroutines)
- Rate limiting statistics

### Resource Limits
- Memory limits with circuit breaker
- Goroutine count monitoring
- CPU usage tracking
- Request body size enforcement
- Timeout management

## Troubleshooting

### Common Issues

**Service Won't Start**
- Check if port 3456 is available
- Verify PID file permissions in ~/.ccproxy/
- Check configuration file syntax

**Authentication Errors**
- Verify API key configuration
- Check IP whitelist/blacklist settings
- Ensure request format (Bearer token or x-api-key header)

**Performance Issues**
- Monitor resource usage with `/health` endpoint (authenticated)
- Check rate limiting configuration
- Verify timeout settings

### Debug Mode
Enable detailed logging by setting `LOG=true` environment variable. Logs are written to `~/.ccproxy/ccproxy.log`.

### Health Checks
- `GET /health` - Basic status (public)
- `GET /health` with auth - Detailed diagnostics
- `GET /status` - Full service information

## Code Quality Standards

### Security
- All input validation and sanitization
- No hardcoded secrets or credentials  
- Proper error handling without information leakage
- Race condition prevention with atomic operations
- Resource exhaustion protection

### Performance
- Bounded caches with LRU eviction
- Connection pooling and reuse
- Efficient memory management
- Minimal allocation patterns
- Streaming support for large responses

### Reliability  
- Graceful shutdown handling
- Circuit breaker patterns
- Comprehensive error recovery
- Process management with proper cleanup
- Atomic state transitions

## Recent Security Fixes (2025-07-21)

1. **API Key Exposure**: Removed x-api-key from CORS headers
2. **Memory Leaks**: Implemented bounded transformer chain cache (100 entries)
3. **Race Conditions**: Added atomic operations for provider metrics
4. **Health Endpoint**: Added authentication requirement for detailed info
5. **PID Security**: Eliminated unsafe fallbacks, enforced exclusive locking
6. **Request Limits**: Added 10MB body size limit and 30s timeout defaults
7. **Error Sanitization**: Provider responses stripped of sensitive data
8. **Resource Enforcement**: Added middleware to reject requests when limits exceeded

All fixes validated with comprehensive test suite and race detection. Build confirmed working on production binary v53140a8-dirty.

## Development Guidelines

- Use cognitive triangulation approach for complex changes
- Run tests with race detection: `go test -race ./...`
- Validate security implications of all changes
- Maintain 100% compatibility with Claude Code integration
- Follow Go best practices and project conventions
- Update this CLAUDE.md file when making significant changes