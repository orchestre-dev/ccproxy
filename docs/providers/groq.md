---
title: Groq - Ultra-Fast AI Inference with Claude Code
description: Use Groq's lightning-fast AI models with Claude Code through CCProxy. Experience the fastest inference speeds available.
keywords: Groq, fast AI, Claude Code, CCProxy, Llama, Mixtral, Gemma, ultra-fast inference
---

# Groq Provider

Experience the fastest AI inference available with Groq's LPU (Language Processing Unit) technology. Groq provides an OpenAI-compatible API that works seamlessly with CCProxy, delivering responses at unprecedented speeds.

## Why Choose Groq?

- ⚡ **Ultra-Fast Inference**: 10-100x faster than traditional GPU inference
- 🚀 **Low Latency**: Near-instant responses for better user experience
- 💰 **Competitive Pricing**: Cost-effective for high-volume applications
- 🛠️ **OpenAI Compatible**: Drop-in replacement for OpenAI API
- 🔧 **Function Calling**: Full support for Claude Code's tool use

## Setup

### 1. Get a Groq API Key

1. Visit [console.groq.com](https://console.groq.com)
2. Sign up for a free account
3. Generate an API key

### 2. Configure CCProxy

Create or update your CCProxy configuration:

```json
{
  "providers": [
    {
      "name": "openai",
      "api_base_url": "https://api.groq.com/openai/v1",
      "api_key": "gsk_your_groq_api_key",
      "models": ["llama-3.1-70b-versatile", "llama-3.1-8b-instant", "mixtral-8x7b-32768"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "llama-3.1-70b-versatile"
    }
  }
}
```

**Important**: Use `"openai"` as the provider name since Groq provides an OpenAI-compatible API.

### 3. Using Environment Variables

You can also configure via environment variables:

```bash
export GROQ_API_KEY="gsk_your_groq_api_key"
```

Then in your config:
```json
{
  "providers": [
    {
      "name": "openai",
      "api_base_url": "https://api.groq.com/openai/v1",
      "api_key": "${GROQ_API_KEY}",
      "enabled": true
    }
  ]
}
```

### 4. Start CCProxy

```bash
ccproxy start
ccproxy code
```

## Available Models

### Llama 3.1 Series
- **llama-3.1-70b-versatile** - Most capable model with tool support
- **llama-3.1-8b-instant** - Fastest responses, good for simple tasks
- **llama3-70b-8192** - Previous generation, still excellent
- **llama3-8b-8192** - Smaller, faster variant

### Mixtral Series
- **mixtral-8x7b-32768** - 32K context window, excellent for code
- **gemma-7b-it** - Google's efficient model
- **gemma2-9b-it** - Latest Gemma model

### Function Calling Support

All Groq models support function calling, making them perfect for Claude Code:

✅ **Full Claude Code Compatibility** - All models work with Claude Code's tool use

## Configuration Examples

### Basic Setup

Simple configuration with Llama 3.1:

```json
{
  "providers": [
    {
      "name": "openai",
      "api_base_url": "https://api.groq.com/openai/v1",
      "api_key": "gsk_your_api_key",
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "llama-3.1-70b-versatile"
    }
  }
}
```

### Speed-Optimized Setup

Different models for different speed requirements:

```json
{
  "providers": [
    {
      "name": "openai",
      "api_base_url": "https://api.groq.com/openai/v1",
      "api_key": "gsk_your_api_key",
      "models": ["llama-3.1-70b-versatile", "llama-3.1-8b-instant", "mixtral-8x7b-32768"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "llama-3.1-70b-versatile"
    },
    "background": {
      "provider": "openai",
      "model": "llama-3.1-8b-instant"
    },
    "longContext": {
      "provider": "openai",
      "model": "mixtral-8x7b-32768"
    }
  }
}
```

### Multi-Provider with Groq for Speed

Use Groq for fast responses, other providers for specialized tasks:

```json
{
  "providers": [
    {
      "name": "openai",
      "api_base_url": "https://api.groq.com/openai/v1",
      "api_key": "${GROQ_API_KEY}",
      "models": ["llama-3.1-8b-instant"],
      "enabled": true
    },
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
    },
    "background": {
      "provider": "openai",
      "model": "llama-3.1-8b-instant"
    }
  }
}
```

## Performance Benchmarks

Groq's LPU technology delivers exceptional performance:

| Model | Tokens/Second | First Token Latency |
|-------|---------------|---------------------|
| llama-3.1-8b-instant | ~1000 | <100ms |
| llama-3.1-70b-versatile | ~500 | <200ms |
| mixtral-8x7b-32768 | ~700 | <150ms |

*Actual speeds may vary based on load and request complexity*

## Best Practices

1. **Model Selection**:
   - Use `llama-3.1-8b-instant` for maximum speed
   - Use `llama-3.1-70b-versatile` for complex reasoning
   - Use `mixtral-8x7b-32768` for code and long contexts

2. **Rate Limits**:
   - Free tier: 30 requests/minute
   - Paid tiers: Higher limits available
   - Implement retry logic for rate limit errors

3. **Cost Optimization**:
   - Groq charges per token like other providers
   - Faster inference can reduce overall costs
   - Monitor usage in Groq console

## Troubleshooting

### API Key Issues

Ensure your API key:
- Starts with `gsk_`
- Has no extra spaces or newlines
- Is properly quoted in JSON

### Rate Limiting

If you hit rate limits:
```json
{
  "error": {
    "message": "Rate limit exceeded",
    "type": "rate_limit_exceeded"
  }
}
```

Solutions:
1. Implement exponential backoff
2. Upgrade to a paid plan
3. Distribute requests over time

### Connection Issues

Test the connection:
```bash
curl https://api.groq.com/openai/v1/models \
  -H "Authorization: Bearer gsk_your_api_key" \
  -H "Content-Type: application/json"
```

## Limitations

- **Context Windows**: Vary by model (8K-32K tokens)
- **Rate Limits**: Based on your plan
- **Model Availability**: Some models may have limited availability during peak times

## Next Steps

- Experiment with different models for your use case
- Monitor response times and optimize routing
- Consider Groq for latency-critical applications
- Combine with other providers for best results

For more information, visit [groq.com](https://groq.com).