# Claude Code Integration

CCProxy is specifically designed to work seamlessly with [Claude Code](https://claude.ai/code), Anthropic's official CLI tool. This integration allows you to use multiple AI providers while maintaining the familiar Claude Code interface.

<SocialShare />

## Quick Setup with Auto-Configuration

### One-Command Setup

```bash
# CCProxy will auto-start and configure Claude Code
./ccproxy code
```

This command:
- Starts CCProxy if not already running
- Sets environment variables for Claude Code
- Manages reference counting for auto-shutdown
- Returns Claude Code configuration

## Manual Setup

### 1. Configure Providers

Create or edit `config.json`:

```json
{
  "host": "127.0.0.1",
  "port": 3456,
  "providers": [
    {
      "name": "anthropic",
      "api_key": "your-anthropic-key",
      "models": ["claude-3-sonnet-20240229", "claude-3-opus-20240229"],
      "enabled": true
    },
    {
      "name": "openai",
      "api_key": "your-openai-key",
      "models": ["gpt-4", "gpt-3.5-turbo"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "anthropic",
      "model": "claude-3-sonnet-20240229"
    }
  }
}
```

### 2. Start CCProxy

```bash
./ccproxy start
```

### 3. Configure Claude Code

```bash
# Point Claude Code to CCProxy
export ANTHROPIC_BASE_URL=http://localhost:3456
export ANTHROPIC_AUTH_TOKEN=test

# Use Claude Code normally
claude "Help me optimize this Python function"
```

## Environment Configuration

### Method 1: Auto-Configuration (Recommended)

```bash
# Let CCProxy configure everything
./ccproxy code
```

### Method 2: Manual Environment Variables

```bash
# Set CCProxy endpoint
export ANTHROPIC_BASE_URL=http://localhost:3456
export ANTHROPIC_AUTH_TOKEN=test
export API_TIMEOUT_MS=600000
```

### Method 3: Shell Profile

Add to your `~/.bashrc`, `~/.zshrc`, or `~/.profile`:

```bash
# CCProxy + Claude Code Integration
export ANTHROPIC_BASE_URL=http://localhost:3456
export ANTHROPIC_AUTH_TOKEN=test
export API_TIMEOUT_MS=600000

# Optional: Auto-start CCProxy
alias claude-start="ccproxy code"
```

### Method 4: Claude Configuration File

Edit `~/.claude.json`:

```json
{
  "providers": {
    "anthropic": {
      "baseUrl": "http://localhost:3456",
      "authToken": "test"
    }
  },
  "timeout": 600000
}
```

## Provider Management

CCProxy uses a configuration-based approach for provider management:

### Configuring Multiple Providers

```json
{
  "providers": [
    {
      "name": "anthropic",
      "api_key": "sk-ant-...",
      "models": ["claude-3-opus-20240229", "claude-3-sonnet-20240229"],
      "enabled": true
    },
    {
      "name": "openai",
      "api_key": "sk-...",
      "models": ["gpt-4", "gpt-3.5-turbo"],
      "enabled": true
    },
    {
      "name": "gemini",
      "api_key": "AI...",
      "models": ["gemini-1.5-flash", "gemini-1.5-pro"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "anthropic",
      "model": "claude-3-sonnet-20240229"
    },
    "longContext": {
      "provider": "anthropic",
      "model": "claude-3-opus-20240229"
    }
  }
}
```

### Dynamic Provider Management

```bash
# Add a new provider
curl -X POST http://localhost:3456/providers \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-ccproxy-api-key" \
  -d '{
    "name": "openai",
    "api_key": "sk-...",
    "enabled": true
  }'

# Update provider configuration
curl -X PUT http://localhost:3456/providers/openai \
  -H "x-api-key: your-ccproxy-api-key" \
  -d '{"enabled": false}'
```

## Common Usage Patterns

### Development Workflow

```bash
# Use default provider (configured in config.json)
claude "Explain this error message"

# For long contexts (>60K tokens), CCProxy automatically routes to longContext model
claude "Analyze this large codebase" < large_file.js

# Use with different models via config.json routes
claude "Design the architecture for this system"

### Code Analysis

```bash
# Analyze code files
claude "Review this JavaScript function for performance issues" < app.js

# Get explanations
claude "Explain what this Python script does" < script.py

# Generate tests
claude "Create unit tests for this Go function" < handler.go
```

### Interactive Development

```bash
# Enter interactive mode
claude

# Use different providers for different tasks
# Type your questions and get responses from configured provider
```

### Batch Processing

```bash
# Process multiple files
for file in *.py; do
    claude "Document this Python file" < "$file" > "${file%.py}_docs.md"
done
```

## Claude Code Features with CCProxy

### All Standard Features Work

- **Interactive mode**: Multi-turn conversations
- **File operations**: Reading and writing files
- **Code generation**: Creating new files and functions
- **Code analysis**: Understanding existing code
- **Documentation**: Generating docs and comments
- **Debugging**: Error analysis and fixes

### Feature Availability by Provider

| Feature | Anthropic | OpenAI | Gemini | DeepSeek | OpenRouter |
|---------|-----------|--------|--------|----------|------------|
| **Code Analysis** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Code Generation** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Documentation** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Image Analysis** | ✅ | ✅ | ✅ | ❌ | ✅* |
| **Function Calling** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Streaming** | ✅ | ✅ | ✅ | ✅ | ✅ |

*Depends on specific model

## Configuration Examples

### Development Setup

```json
{
  "host": "127.0.0.1",
  "port": 3456,
  "providers": [
    {
      "name": "anthropic",
      "api_key": "${ANTHROPIC_API_KEY}",
      "enabled": true
    },
    {
      "name": "deepseek",
      "api_key": "${DEEPSEEK_API_KEY}",
      "models": ["deepseek-coder"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "deepseek",
      "model": "deepseek-coder"
    }
  }
}
```

### Production Setup

```json
{
  "host": "0.0.0.0",
  "port": 3456,
  "apikey": "${CCPROXY_APIKEY}",
  "performance": {
    "rate_limit_enabled": true,
    "circuit_breaker_enabled": true
  },
  "providers": [
    {
      "name": "openai",
      "api_key": "${OPENAI_API_KEY}",
      "models": ["gpt-4", "gpt-3.5-turbo"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "gpt-4"
    }
  }
}
```

### Multi-Model Setup

```json
{
  "routes": {
    "default": {
      "provider": "anthropic",
      "model": "claude-3-sonnet-20240229"
    },
    "longContext": {
      "provider": "anthropic",
      "model": "claude-3-opus-20240229"
    },
    "vision": {
      "provider": "gemini",
      "model": "gemini-1.5-pro"
    }
  }
}
```

### Configuration Management

Use different configuration files for different use cases:

```bash
# Development configuration
cp config.dev.json config.json
./ccproxy start

# Production configuration  
cp config.prod.json config.json
./ccproxy start

# Test configuration
cp config.test.json config.json
./ccproxy start
```

#### Example config.dev.json:
```json
{
  "log": true,
  "providers": [
    {
      "name": "anthropic",
      "api_key": "${ANTHROPIC_API_KEY}",
      "models": ["claude-3-sonnet-20240229"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "anthropic",
      "model": "claude-3-sonnet-20240229"
    }
  }
}
```

#### Example config.prod.json:
```json
{
  "host": "0.0.0.0",
  "port": 443,
  "apikey": "${CCPROXY_APIKEY}",
  "log": false,
  "performance": {
    "rate_limit_enabled": true,
    "metrics_enabled": true
  },
  "providers": [
    {
      "name": "openai",
      "api_key": "${OPENAI_API_KEY}",
      "models": ["gpt-4", "gpt-3.5-turbo"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "gpt-4"
    }
  }
}
```

## Automation Scripts

### CCProxy Service Manager

```bash
#!/bin/bash
# ccproxy-manager.sh

case $1 in
    start)
        ./ccproxy start
        echo "✅ CCProxy started"
        ;;
    stop)
        ./ccproxy stop
        echo "❌ CCProxy stopped"
        ;;
    restart)
        ./ccproxy stop
        sleep 1
        ./ccproxy start
        echo "🔄 CCProxy restarted"
        ;;
    status)
        ./ccproxy status
        ;;
    claude)
        ./ccproxy code
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|claude}"
        exit 1
        ;;
esac
```

### Health Check Script

```bash
#!/bin/bash
# health-check.sh

# Check CCProxy health
HEALTH=$(curl -s http://localhost:3456/health | jq -r '.status')

if [ "$HEALTH" = "healthy" ]; then
    echo "✅ CCProxy is healthy"
    
    # Get detailed status if authenticated
    if [ -n "$CCPROXY_API_KEY" ]; then
        curl -s -H "x-api-key: $CCPROXY_API_KEY" http://localhost:3456/status | jq
    fi
else
    echo "❌ CCProxy is unhealthy"
    exit 1
fi
```

### Claude Code Session Manager

```bash
#!/bin/bash
# claude-session.sh

# Start CCProxy and configure Claude Code
eval $(./ccproxy code)

# Verify configuration
if [ "$ANTHROPIC_BASE_URL" = "http://127.0.0.1:3456" ]; then
    echo "✅ Claude Code configured for CCProxy"
    
    # Run Claude Code with auto-cleanup
    trap "./ccproxy stop" EXIT
    claude "$@"
else
    echo "❌ Configuration failed"
    exit 1
fi
```

## Troubleshooting

### CCProxy Not Responding

```bash
# Check if CCProxy is running
./ccproxy status

# If not running, start it
./ccproxy start

# Check logs for errors (if logging enabled)
tail -f ~/.ccproxy/ccproxy.log

# Verify process
ps aux | grep ccproxy
```

### Claude Code Connection Issues

```bash
# Verify environment variables
echo $ANTHROPIC_BASE_URL    # Should be http://127.0.0.1:3456
echo $ANTHROPIC_AUTH_TOKEN  # Should be test
echo $API_TIMEOUT_MS        # Should be 600000

# Use ccproxy code to auto-configure
./ccproxy code

# Test direct API call
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-3-sonnet-20240229","messages":[{"role":"user","content":"test"}],"max_tokens":10}'
```

### Provider Configuration Issues

```bash
# Check provider status (requires auth or localhost)
curl http://localhost:3456/status

# Verify configuration file
cat config.json | jq '.providers'

# Check enabled providers
cat config.json | jq '.providers[] | select(.enabled==true) | .name'

# List available providers (requires auth)
curl -H "x-api-key: $CCPROXY_API_KEY" http://localhost:3456/providers
```

### Performance Issues

```bash
# Check CCProxy performance (authenticated)
curl -H "x-api-key: $CCPROXY_API_KEY" http://localhost:3456/health | jq '.performance'

# Monitor response times
time claude "simple test"

# Check latency metrics
curl -H "x-api-key: $CCPROXY_API_KEY" http://localhost:3456/health | \
  jq '.performance.latency'

# Review configuration for performance settings
cat config.json | jq '.performance'
```

## Best Practices

### 1. Provider Selection by Use Case

```json
// config.json - Route by use case
{
  "routes": {
    "default": {
      "provider": "deepseek",      // Fast, cost-effective for development
      "model": "deepseek-coder"
    },
    "production": {
      "provider": "openai",         // High quality for production
      "model": "gpt-4"
    },
    "longContext": {
      "provider": "anthropic",      // Best for large context windows
      "model": "claude-3-opus-20240229"
    }
  }
}
```

### 2. Configuration Management

```bash
# Use separate config files for different environments
config.json          # Active configuration
config.dev.json      # Development settings
config.prod.json     # Production settings
config.test.json     # Test settings

# Use environment variables for sensitive data
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENAI_API_KEY="sk-..."

# Reference in config.json
"api_key": "${ANTHROPIC_API_KEY}"
```

### 3. Error Handling

```bash
# Check CCProxy health before long tasks
HEALTH=$(curl -s http://localhost:3456/health | jq -r '.status')
if [ "$HEALTH" != "healthy" ]; then
    echo "CCProxy not healthy, restarting..."
    ./ccproxy stop && ./ccproxy start
fi

# Configure multiple providers for fallback
# CCProxy will handle provider failures gracefully
```

### 4. Security Best Practices

```bash
# Use API key for production
cat config.json | jq '.apikey = "${CCPROXY_API_KEY}"' > config.prod.json

# Restrict to localhost for development
cat config.json | jq '.host = "127.0.0.1"' > config.dev.json

# Enable security features
cat config.json | jq '.security.audit_enabled = true' > config.secure.json
```

## Advanced Integration

### IDE Integration

Configure your IDE to use Claude Code with CCProxy:

#### VS Code Settings
```json
{
  "claude.anthropic.baseUrl": "http://localhost:3456",
  "claude.anthropic.apiKey": "NOT_NEEDED"
}
```

#### Vim Configuration
```vim
" In your .vimrc
let g:claude_base_url = 'http://localhost:3456'
let g:claude_api_key = 'NOT_NEEDED'
```

### CI/CD Integration

```yaml
# .github/workflows/ai-review.yml
name: AI Code Review
on: [pull_request]

jobs:
  ai-review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup CCProxy
        run: |
          export PROVIDER=groq
          export GROQ_API_KEY=${{ secrets.GROQ_API_KEY }}
          ./ccproxy &
          
      - name: Configure Claude Code
        run: |
          export ANTHROPIC_BASE_URL=http://localhost:3456
          export ANTHROPIC_API_KEY=NOT_NEEDED
          
      - name: AI Code Review
        run: |
          claude "Review the changes in this PR for bugs and improvements" < diff.txt
```

### Docker Integration

```dockerfile
# Dockerfile for development environment
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o ccproxy ./cmd/ccproxy

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/ccproxy /usr/local/bin/
COPY config.json /home/ccproxy/.ccproxy/config.json

# Configure Claude Code environment
ENV ANTHROPIC_BASE_URL=http://localhost:3456
ENV ANTHROPIC_AUTH_TOKEN=test
ENV API_TIMEOUT_MS=600000

EXPOSE 3456
CMD ["ccproxy", "start", "--foreground"]
```

## Monitoring and Logging

### Request Tracking

```bash
# Monitor Claude Code requests through CCProxy
tail -f ccproxy.log | grep claude

# Track usage patterns
curl http://localhost:3456/status | jq '.metrics'
```

### Performance Monitoring

```bash
# Monitor response times by provider
time claude "test query"

# Track token usage
curl http://localhost:3456/health | jq '.requests'
```

## Next Steps

- Learn about [provider-specific features](/providers/) to optimize for your use case
- Set up [monitoring and alerting](/guide/monitoring) for production use
- Explore [advanced configuration](/guide/configuration) options
- Check out [advanced workflows](/guide/advanced-workflows) for Claude Code + CCProxy integration