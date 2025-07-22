---
title: Groq with Claude Code - Ultra-Fast AI Inference via CCProxy
description: Experience lightning-fast AI development with Groq LPU infrastructure through CCProxy. Access high-performance models with Claude Code for sub-second inference speeds.
keywords: Groq, Claude Code, CCProxy, fast AI inference, LPU, ultra-fast AI, AI proxy, sub-second responses
---

# Groq Provider

**Groq revolutionizes AI development** with ultra-fast inference speeds through their groundbreaking LPU (Language Processing Unit) technology. When combined with **Claude Code and CCProxy**, Groq delivers an unmatched development experience with **sub-second response times**.

## 🚀 Why Choose Groq for Claude Code?

- ⚡ **Ultra-fast inference**: Sub-second response times with LPU technology
- 🚀 **High-performance models**: Access to optimized models for speed
- 💰 **Cost-effective**: Competitive pricing with generous free tier
- 🎯 **Simple API**: Easy integration with Claude Code via CCProxy
- 🔄 **High throughput**: Excellent for high-volume AI development workflows

## Setup

### 1. Get a Groq API Key

1. Visit [console.groq.com](https://console.groq.com)
2. Sign up for a free account
3. Navigate to the API Keys section
4. Generate a new API key

### 2. Configure CCProxy

Create or update your CCProxy configuration file:

```bash
mkdir -p ~/.ccproxy
cat > ~/.ccproxy/config.json << 'EOF'
{
  "providers": [
    {
      "name": "groq",
      "api_base_url": "https://api.groq.com/openai/v1",
      "api_key": "gsk_your_groq_api_key_here",
      "models": ["llama3-8b-8192", "mixtral-8x7b-32768"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "groq",
      "model": "llama3-8b-8192"
    }
  }
}
EOF
```

### 3. Start CCProxy and Claude Code

```bash
# Start CCProxy
ccproxy start

# Configure Claude Code to use CCProxy
ccproxy code
```

## Available Models

Groq provides several models optimized for speed:

| Model | Context Length | Best For | Function Calling |
|-------|----------------|----------|------------------|
| `llama3-70b-8192` | 8,192 tokens | High-quality responses | ✅ Yes |
| `llama3-8b-8192` | 8,192 tokens | Fast, cost-effective | ✅ Yes |
| `mixtral-8x7b-32768` | 32,768 tokens | Long context tasks | ✅ Yes |
| `gemma-7b-it` | 8,192 tokens | Efficient inference | ❌ No |

**⚠️ Important**: Claude Code requires models with function calling support. Use models marked with ✅ above.

## Configuration Options

### Complete Configuration Example

```json
{
  "providers": [
    {
      "name": "groq",
      "api_base_url": "https://api.groq.com/openai/v1",
      "api_key": "gsk_your_api_key",
      "models": ["llama3-70b-8192", "llama3-8b-8192", "mixtral-8x7b-32768"],
      "enabled": true,
      "timeout": "30s",
      "max_retries": 3,
      "headers": {
        "X-Custom-Header": "value"
      }
    }
  ],
  "routes": {
    "default": {
      "provider": "groq",
      "model": "llama3-8b-8192"
    }
  }
}
```

### Using Environment Variables

While CCProxy primarily uses config.json, you can use environment variables for API key security:

```bash
# Store API key in environment
export GROQ_API_KEY="gsk_your_api_key"

# Reference in config.json
{
  "providers": [{
    "name": "groq",
    "api_key": "${GROQ_API_KEY}",
    ...
  }]
}
```

## Troubleshooting

### Model Not Found
If you get a "model not found" error:
- Check available models at [console.groq.com](https://console.groq.com)
- Update your config.json with currently available models
- Ensure the model supports function calling for Claude Code

### Rate Limits
Groq has rate limits based on your plan:
- Free tier: Limited requests per minute
- Paid plans: Higher limits available
- Monitor usage at [console.groq.com](https://console.groq.com)

### Function Calling Errors
If Claude Code features aren't working:
- Verify you're using a model with function calling support
- Check the models table above for compatibility
- Consider switching to `llama3-70b-8192` or `mixtral-8x7b-32768`

## Best Practices

1. **Model Selection**: Start with `llama3-8b-8192` for best speed/quality balance
2. **Context Management**: Groq models have varying context limits - choose based on your needs
3. **Error Handling**: Implement retries for transient errors
4. **Monitoring**: Track your usage to stay within rate limits

## Next Steps

- [Quick Start Guide](/guide/quick-start) - Get running in 2 minutes
- [Configuration Guide](/guide/configuration) - Advanced configuration options
- [Provider Comparison](/providers/) - Compare with other providers