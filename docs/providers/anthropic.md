# Anthropic Provider

CCProxy provides native support for Anthropic's Claude models, including the latest Claude 4 family with enhanced reasoning and coding capabilities.

## Latest Models (July 2025)

### Claude 4 Family
- **claude-opus-4-20250720** - Most powerful model, best for complex reasoning and coding (SWE-bench 72.5%)
  - Exceptional performance on complex tasks
  - Superior reasoning and analysis capabilities
  - Best choice for long-context processing (>60K tokens)
  - Ideal for advanced coding tasks

- **claude-sonnet-4-20250720** - Balanced performance and speed
  - Excellent balance of capabilities and response time
  - Great for general-purpose tasks
  - Strong reasoning with faster responses than Opus
  - Recommended as default model

### Claude 3 Family
- **claude-3-7-sonnet-20250721** - Most intelligent Claude 3 model
  - Latest iteration of Claude 3 Sonnet
  - Enhanced capabilities over previous Claude 3 models
  - Good alternative when Claude 4 models are not required

- **claude-3-5-haiku-20241022** - Fast and efficient for simple tasks
  - Optimized for speed and cost-effectiveness
  - Ideal for background tasks and simple queries
  - Low latency responses

## Configuration

### Provider Setup

```json
{
  "providers": {
    "anthropic": {
      "enabled": true,
      "api_key": "${ANTHROPIC_API_KEY}",
      "base_url": "https://api.anthropic.com/v1",
      "models": {
        "claude-opus-4-20250720": {
          "enabled": true,
          "max_tokens": 200000,
          "context_window": 200000
        },
        "claude-sonnet-4-20250720": {
          "enabled": true,
          "max_tokens": 200000,
          "context_window": 200000
        },
        "claude-3-7-sonnet-20250721": {
          "enabled": true,
          "max_tokens": 200000,
          "context_window": 200000
        },
        "claude-3-5-haiku-20241022": {
          "enabled": true,
          "max_tokens": 200000,
          "context_window": 200000
        }
      }
    }
  }
}
```

### Routing Configuration

```json
{
  "router": {
    "routes": {
      "default": {
        "provider": "anthropic",
        "model": "claude-sonnet-4-20250720",
        "description": "Balanced performance for most tasks"
      },
      "longContext": {
        "provider": "anthropic",
        "model": "claude-opus-4-20250720",
        "description": "Best for >60K tokens or complex reasoning",
        "token_threshold": 60000
      },
      "background": {
        "provider": "anthropic",
        "model": "claude-3-5-haiku-20241022",
        "description": "Fast, cost-effective for simple tasks"
      }
    }
  }
}
```

## Environment Variables

```bash
# Required - CCProxy will auto-detect this
export ANTHROPIC_API_KEY="sk-ant-..."

# Optional - for CCProxy to intercept Claude Code
export ANTHROPIC_BASE_URL="http://127.0.0.1:3456"
```

## Request Format

CCProxy accepts standard Anthropic API format with the following supported parameters:

```json
{
  "model": "claude-sonnet-4-20250720",
  "messages": [
    {
      "role": "user",
      "content": "Hello, Claude!"
    }
  ],
  "max_tokens": 1024,
  "temperature": 0.7,
  "top_p": 0.9
}
```

### Supported Parameters

- **model** (required): The model identifier
- **messages** (required): Array of message objects with role and content
- **max_tokens** (optional): Maximum tokens to generate (default: model-specific)
- **temperature** (optional): Sampling temperature between 0 and 1 (default: 1.0)
- **top_p** (optional): Nucleus sampling between 0 and 1 (default: 1.0)
- **stream** (optional): Enable streaming responses (default: false)
- **system** (optional): System message (extracted from messages array)
- **tools** (optional): Array of tool definitions for function calling
- **tool_choice** (optional): Tool selection strategy

### Removed Parameters

The following OpenAI-style parameters are automatically removed when routing to Anthropic:
- **frequency_penalty**: Not supported by Anthropic API
- **presence_penalty**: Not supported by Anthropic API

These parameters will be silently dropped to ensure compatibility.

## Streaming Support

All Anthropic models support streaming responses:

```bash
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-sonnet-4-20250720",
    "messages": [{"role": "user", "content": "Stream a response"}],
    "stream": true,
    "max_tokens": 1000
  }'
```

## Model Selection Guidelines

### Use claude-opus-4-20250720 for:
- Complex reasoning tasks requiring deep analysis
- Long-context processing (>60K tokens)
- Advanced coding and software engineering tasks
- Tasks requiring maximum intelligence and capability

### Use claude-sonnet-4-20250720 for:
- General-purpose conversations
- Balanced performance requirements
- Standard coding tasks
- Most default routing scenarios

### Use claude-3-7-sonnet-20250721 for:
- When Claude 4 models are unavailable
- Tasks requiring Claude 3 specific behavior
- Compatibility with older integrations

### Use claude-3-5-haiku-20241022 for:
- High-volume, simple tasks
- Background processing
- Cost-sensitive applications
- Low-latency requirements

## Advanced Features

### Context Window Management

All models support 200K token context windows. CCProxy automatically handles:
- Token counting using tiktoken-compatible tokenizer
- Automatic routing to longContext models when needed
- Context window overflow prevention

### Tool Use

All Anthropic models support function calling:

```json
{
  "model": "claude-sonnet-4-20250720",
  "messages": [
    {
      "role": "user",
      "content": "What's the weather like?"
    }
  ],
  "tools": [
    {
      "name": "get_weather",
      "description": "Get current weather",
      "input_schema": {
        "type": "object",
        "properties": {
          "location": {"type": "string"}
        }
      }
    }
  ]
}
```

## Error Handling

CCProxy provides standardized error responses:

```json
{
  "error": {
    "type": "invalid_request_error",
    "message": "Model not found: claude-3-opus-20240229",
    "suggestion": "Use claude-opus-4-20250720 instead"
  }
}
```

## Rate Limiting

Anthropic enforces rate limits per model tier:
- Opus models: Lower rate limits, higher priority
- Sonnet models: Balanced rate limits
- Haiku models: Higher rate limits for volume usage

CCProxy respects these limits and provides appropriate retry mechanisms.

## Best Practices

1. **Model Selection**: Choose the appropriate model based on task complexity
2. **Token Management**: Monitor token usage to optimize routing
3. **Streaming**: Use streaming for better user experience on long responses
4. **Error Handling**: Implement retry logic for transient errors
5. **Cost Optimization**: Use Haiku for simple tasks to reduce costs
6. **Context Management**: Be mindful of context window limits

## Migration from Older Models

If you're using deprecated models, update your configuration:

| Old Model | Recommended Replacement |
|-----------|------------------------|
| claude-3-opus-20240229 | claude-opus-4-20250720 |
| claude-3-sonnet-20240229 | claude-sonnet-4-20250720 |
| claude-3-haiku-20240307 | claude-3-5-haiku-20241022 |
| claude-2.1 | claude-sonnet-4-20250720 |
| claude-2.0 | claude-sonnet-4-20250720 |
| claude-instant-1.2 | claude-3-5-haiku-20241022 |