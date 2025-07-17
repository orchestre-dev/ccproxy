# Claude Code Integration

CCProxy is specifically designed to work seamlessly with [Claude Code](https://claude.ai/code), Anthropic's official CLI tool. This integration allows you to use any of the 7 supported providers while maintaining the familiar Claude Code interface.

## Quick Setup

### 1. Start CCProxy

```bash
# Configure your preferred provider
export PROVIDER=groq
export GROQ_API_KEY=your_api_key_here

# Start CCProxy
./ccproxy
```

### 2. Configure Claude Code

```bash
# Point Claude Code to CCProxy
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED

# Use Claude Code normally
claude "Help me optimize this Python function"
```

That's it! Claude Code will now use your configured provider through CCProxy.

## Environment Configuration

### Method 1: Environment Variables

```bash
# Set CCProxy endpoint
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED

# Provider configuration
export PROVIDER=groq
export GROQ_API_KEY=your_groq_key_here
```

### Method 2: Claude Code Configuration

```bash
# Configure Claude Code directly
claude config set anthropic.base_url http://localhost:7187
claude config set anthropic.api_key NOT_NEEDED
```

### Method 3: Shell Profile

Add to your `~/.bashrc`, `~/.zshrc`, or `~/.profile`:

```bash
# CCProxy + Claude Code Integration
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED

# Default provider (change as needed)
export PROVIDER=groq
export GROQ_API_KEY=your_groq_key_here
```

## Provider Switching

One of the key benefits of CCProxy is the ability to switch providers without changing Claude Code configuration:

### Switch to Different Providers

```bash
# Switch to OpenAI
export PROVIDER=openai
export OPENAI_API_KEY=your_openai_key
./ccproxy  # Restart CCProxy

# Switch to Ollama (local)
export PROVIDER=ollama
export OLLAMA_MODEL=llama3.2
./ccproxy  # Restart CCProxy

# Switch to XAI for real-time data
export PROVIDER=xai
export XAI_API_KEY=your_xai_key
./ccproxy  # Restart CCProxy
```

Claude Code commands remain the same - only the underlying provider changes!

## Common Usage Patterns

### Development Workflow

```bash
# Fast iteration with Groq
export PROVIDER=groq
claude "Explain this error message"

# Switch to OpenAI for complex reasoning
export PROVIDER=openai
claude "Design the architecture for this system"

# Use Ollama for private/sensitive code
export PROVIDER=ollama
claude "Review this proprietary algorithm"
```

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

| Feature | Groq | OpenRouter | OpenAI | XAI | Gemini | Mistral | Ollama |
|---------|------|------------|--------|-----|--------|---------|--------|
| **Code Analysis** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Code Generation** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Documentation** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Image Analysis** | ❌ | ✅* | ✅ | ✅ | ✅ | ❌ | ✅* |
| **Real-time Data** | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| **Function Calling** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

*Depends on specific model

## Configuration Examples

### Development Setup

```bash
# .env for development
PROVIDER=groq
GROQ_API_KEY=your_groq_key
GROQ_MODEL=llama-3.1-8b-instant  # Fast model for development

# Claude Code config
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED
```

### Production Setup

```bash
# .env for production
PROVIDER=openai
OPENAI_API_KEY=your_openai_key
OPENAI_MODEL=gpt-4o  # High-quality model for production

# Claude Code config
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED
```

### Privacy-Focused Setup

```bash
# .env for sensitive work
PROVIDER=ollama
OLLAMA_MODEL=codellama:13b  # Local model for privacy
OLLAMA_BASE_URL=http://localhost:11434

# Claude Code config
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED
```

### Multi-Provider Setup

Create different configurations for different use cases:

```bash
# groq-config.sh
export PROVIDER=groq
export GROQ_API_KEY=your_groq_key
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED

# openai-config.sh  
export PROVIDER=openai
export OPENAI_API_KEY=your_openai_key
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED

# Usage:
# source groq-config.sh && claude "fast iteration task"
# source openai-config.sh && claude "complex reasoning task"
```

## Automation Scripts

### Provider Switcher

```bash
#!/bin/bash
# switch-provider.sh

PROVIDER=$1

case $PROVIDER in
    groq)
        export PROVIDER=groq
        export GROQ_API_KEY=$GROQ_API_KEY
        ;;
    openai)
        export PROVIDER=openai
        export OPENAI_API_KEY=$OPENAI_API_KEY
        ;;
    ollama)
        export PROVIDER=ollama
        export OLLAMA_MODEL=llama3.2
        ;;
    *)
        echo "Usage: $0 {groq|openai|ollama}"
        exit 1
        ;;
esac

# Restart CCProxy with new provider
pkill ccproxy
./ccproxy &

echo "Switched to $PROVIDER provider"
```

### Claude Code Wrapper

```bash
#!/bin/bash
# claude-with-provider.sh

PROVIDER=$1
shift  # Remove provider from arguments

# Set provider configuration
case $PROVIDER in
    groq)
        export PROVIDER=groq
        ;;
    openai)
        export PROVIDER=openai
        ;;
    # ... other providers
esac

# Ensure CCProxy is running with correct provider
./switch-provider.sh $PROVIDER

# Run Claude Code with remaining arguments
claude "$@"
```

Usage:
```bash
./claude-with-provider.sh groq "Explain this code"
./claude-with-provider.sh openai "Design this architecture"
```

## Troubleshooting

### CCProxy Not Responding

```bash
# Check if CCProxy is running
curl http://localhost:7187/health

# If not running, start it
./ccproxy

# Check logs for errors
tail -f ccproxy.log
```

### Claude Code Connection Issues

```bash
# Verify environment variables
echo $ANTHROPIC_BASE_URL  # Should be http://localhost:7187
echo $ANTHROPIC_API_KEY   # Should be NOT_NEEDED

# Test direct API call
curl -X POST http://localhost:7187/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-3-sonnet","messages":[{"role":"user","content":"test"}],"max_tokens":10}'
```

### Provider Configuration Issues

```bash
# Check provider status
curl http://localhost:7187/status

# Verify API keys are set
env | grep API_KEY

# Check provider-specific configuration
env | grep GROQ     # For Groq
env | grep OPENAI   # For OpenAI
env | grep OLLAMA   # For Ollama
```

### Performance Issues

```bash
# Check CCProxy performance
curl http://localhost:7187/health

# Monitor response times
time claude "simple test"

# Switch to faster provider if needed
export PROVIDER=groq  # Groq is fastest
```

## Best Practices

### 1. Provider Selection by Use Case

```bash
# Development/iteration: Use Groq (fastest)
export PROVIDER=groq

# Production/quality: Use OpenAI or Claude via OpenRouter
export PROVIDER=openrouter

# Privacy/sensitive: Use Ollama (local)
export PROVIDER=ollama

# Real-time data: Use XAI
export PROVIDER=xai
```

### 2. Configuration Management

```bash
# Keep provider configs in separate files
# Use environment-specific configurations
# Version control your .env.example files
```

### 3. Error Handling

```bash
# Always check CCProxy health before long tasks
curl -s http://localhost:7187/health | jq '.status'

# Implement fallback providers for critical workflows
```

### 4. Cost Optimization

```bash
# Use free tiers for development (Groq, Gemini)
# Switch to paid providers only for production
# Monitor usage with provider dashboards
```

## Advanced Integration

### IDE Integration

Configure your IDE to use Claude Code with CCProxy:

#### VS Code Settings
```json
{
  "claude.anthropic.baseUrl": "http://localhost:7187",
  "claude.anthropic.apiKey": "NOT_NEEDED"
}
```

#### Vim Configuration
```vim
" In your .vimrc
let g:claude_base_url = 'http://localhost:7187'
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
          export ANTHROPIC_BASE_URL=http://localhost:7187
          export ANTHROPIC_API_KEY=NOT_NEEDED
          
      - name: AI Code Review
        run: |
          claude "Review the changes in this PR for bugs and improvements" < diff.txt
```

### Docker Integration

```dockerfile
# Dockerfile for development environment
FROM anthropic/claude-code:latest

# Install CCProxy
COPY ccproxy /usr/local/bin/
COPY .env /app/.env

# Configure Claude Code
ENV ANTHROPIC_BASE_URL=http://localhost:7187
ENV ANTHROPIC_API_KEY=NOT_NEEDED

# Start CCProxy and keep container running
CMD ["sh", "-c", "ccproxy & && tail -f /dev/null"]
```

## Monitoring and Logging

### Request Tracking

```bash
# Monitor Claude Code requests through CCProxy
tail -f ccproxy.log | grep claude-code

# Track usage patterns
curl http://localhost:7187/status | jq '.metrics'
```

### Performance Monitoring

```bash
# Monitor response times by provider
time claude "test query"

# Track token usage
curl http://localhost:7187/health | jq '.requests'
```

## Next Steps

- Learn about [provider-specific features](/providers/) to optimize for your use case
- Set up [monitoring and alerting](/guide/monitoring) for production use
- Explore [advanced configuration](/guide/configuration) options
- Check out [best practices](/guide/best-practices) for Claude Code + CCProxy workflows