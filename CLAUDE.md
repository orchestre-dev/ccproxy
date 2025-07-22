# CCProxy - AI Request Proxy for Claude Code (Go Implementation)

## Project Overview

CCProxy is a high-performance Go-based API translation proxy that enables Claude Code to work with multiple AI providers. It acts as a middleware layer that translates Anthropic's API format to various provider-specific formats, enabling seamless integration without modifying Claude Code. This is a complete rewrite from TypeScript to Go with enhanced security, performance, and reliability features.

**Important**: CCProxy is a translation proxy - it does not add new capabilities beyond API format translation. It requires models with function calling support for Claude Code compatibility.

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

- **API Translation**: Converts Anthropic API format to provider-specific formats
- **Intelligent Routing**: Automatic model selection based on:
  - Token count (>60K → longContext route)
  - Model type (haiku models → background route)
  - Boolean thinking parameter (thinking: true → think route)
  - Explicit provider selection (format: "provider,model")
- **Provider Support**: Full transformers for Anthropic, OpenAI, Gemini, DeepSeek, OpenRouter
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
- `CCPROXY_PROVIDERS_0_API_KEY` - Override first provider's API key
- `CCPROXY_PROVIDERS_1_API_KEY` - Override second provider's API key
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

## Provider Support & Limitations

### Implemented Providers (with Transformers)
These providers have dedicated transformer implementations in `internal/transformer/`:

- **Anthropic** (`anthropic.go`) - Full Claude API support with function calling
  - Removes unsupported OpenAI parameters: `frequency_penalty`, `presence_penalty`
  - Complete tool/function calling support
  
- **OpenAI** (`openai.go`) - Most complete parameter support
  - Supports all standard parameters including `frequency_penalty`, `presence_penalty`
  - Full function calling support
  
- **Google Gemini** (`gemini.go`) - Limited tool support
  - Maps parameters to Gemini format: `max_tokens` → `maxOutputTokens`
  - Wraps parameters in `generationConfig` object
  - Function calling may have compatibility issues
  - No support for provider-specific features like `thinkingBudget`
  
- **DeepSeek** (`deepseek.go`) - NO function calling support
  - Hard limit of 8192 max_tokens
  - Special handling for `reasoning_content` in streaming
  - NOT recommended for Claude Code due to lack of tool support
  
- **OpenRouter** (`openrouter.go`) - Multi-provider gateway
  - Passes through to underlying providers
  - Function calling depends on selected model

### Not Implemented
These providers appear in the codebase but have NO transformer implementations:
- Groq, Mistral, XAI/Grok, Ollama - Referenced but not functional

### Claude Code Compatibility
For Claude Code usage, only providers with function calling support are recommended:
- ✅ **Best**: Anthropic, OpenAI (full tool support)
- ⚠️ **Limited**: Gemini (may have issues)
- ❌ **Not Compatible**: DeepSeek (no tool support)

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

## Model Selection & Routing

### Routing Priority (Highest to Lowest)
1. **Explicit Provider Selection**: "provider,model" format (e.g., "anthropic,claude-3-opus")
2. **Direct Model Routes**: Exact model name matches in routes config
3. **Long Context Routing**: Token count > 60,000 triggers longContext route
4. **Background Routing**: Models starting with "claude-3-5-haiku" use background route
5. **Thinking Routing**: Boolean `thinking: true` parameter triggers think route
6. **Default Route**: Fallback for all unmatched requests

### Route Configuration Structure

Routes are defined as a map where keys can be:
- **Special route names**: `default`, `longContext`, `background`, `think`
- **Anthropic model names**: Direct model-to-provider mappings (e.g., `"claude-opus-4"`, `"claude-3-5-sonnet-20241022"`)

Each route contains:
```json
{
  "provider": "provider_name",    // Must match a configured provider name
  "model": "actual_model_name",   // The actual model to use at the target provider
  "conditions": []                // Optional conditions (defined but not currently implemented)
}
```

**Important**: Direct model routes use Anthropic model names as keys because CCProxy acts as a proxy for Claude Code, which only sends Anthropic model names. These routes allow you to redirect specific Anthropic models to different providers or models.

**Note**: While the `conditions` field exists in the Route struct for future extensibility, it is not currently used by the router implementation. All routing decisions are based on the hardcoded logic described in the routing priority section above.

### Configuration Example
```json
{
  "providers": [
    {
      "name": "anthropic",
      "api_key": "sk-ant-...",
      "models": ["claude-opus-4", "claude-sonnet-4"],  // For validation only
      "enabled": true
    },
    {
      "name": "openai",
      "api_key": "sk-...",
      "models": ["gpt-4.1", "o3"],
      "enabled": true
    },
    {
      "name": "deepseek",
      "api_key": "sk-...",
      "models": ["deepseek-chat", "deepseek-reasoner"],
      "enabled": true
    }
  ],
  "routes": {
    // Special routes - triggered by conditions
    "default": {
      "provider": "anthropic",
      "model": "claude-sonnet-4-20250720"
    },
    "longContext": {
      "provider": "anthropic",
      "model": "claude-opus-4-20250720"
    },
    "background": {
      "provider": "openai",
      "model": "gpt-4.1-mini"
    },
    "think": {
      "provider": "deepseek",
      "model": "deepseek-reasoner"
    },
    // Direct model routes - map Anthropic model names to any provider
    "claude-opus-4": {
      "provider": "openai",
      "model": "gpt-4.1-turbo"  // Route claude-opus-4 requests to GPT-4.1 Turbo
    },
    "claude-3-5-sonnet-20241022": {
      "provider": "deepseek",
      "model": "deepseek-chat"  // Route specific Sonnet model to DeepSeek
    }
  }
}
```

### Parameter Support
**Standard Parameters** (all providers):
- `model`, `messages`, `max_tokens`, `temperature`, `top_p`, `stream`

**Provider-Specific Handling**:
- **OpenAI**: Also supports `top_k`, `presence_penalty`, `frequency_penalty`
- **Anthropic**: Strips `presence_penalty`, `frequency_penalty` 
- **Gemini**: Transforms to Gemini format, wraps in `generationConfig`
- **DeepSeek**: Enforces 8192 token limit

## Development Guidelines

- CCProxy is a translation proxy - do not add features beyond API translation
- Always verify function calling support for Claude Code compatibility
- Run tests with race detection: `go test -race ./...`
- Validate security implications of all changes
- Document only implemented features in provider docs
- Update this CLAUDE.md file when making significant changes