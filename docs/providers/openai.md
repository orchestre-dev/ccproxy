---
title: OpenAI with Claude Code - Industry-Leading AI Models via CCProxy
description: Access GPT-4, GPT-4o, and o1 reasoning models with Claude Code through CCProxy. Experience enterprise-grade AI with vision capabilities, function calling, and production reliability.
keywords: OpenAI, Claude Code, CCProxy, GPT-4, GPT-4o, o1 reasoning, AI proxy, enterprise AI, vision AI, function calling
---

# OpenAI Provider

**OpenAI sets the industry standard** for AI models, offering the most mature ecosystem and reliable performance for production applications. Through **CCProxy integration with Claude Code**, you can harness the full power of GPT-4, GPT-4o, and advanced reasoning models while maintaining your familiar development workflow.

## 🏭 Why Choose OpenAI for Claude Code?

- 🥇 **Industry standard**: Most mature and reliable AI models with proven enterprise adoption
- 🛠️ **Rich ecosystem**: Extensive tooling and integrations that work seamlessly with Claude Code
- 👁️ **Advanced vision**: Best-in-class image understanding and multimodal capabilities
- 🎯 **Proven reliability**: Battle-tested in production environments worldwide
- 🧠 **Advanced reasoning**: Access to o1 models for complex problem-solving
- 🔧 **Perfect Claude Code integration**: Zero configuration changes required with CCProxy

## Setup

### 1. Get an OpenAI API Key

1. Visit [platform.openai.com](https://platform.openai.com)
2. Sign up for an account
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
      "name": "openai",
      "api_base_url": "https://api.openai.com/v1",
      "api_key": "sk-your_openai_api_key_here",
      "models": ["gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-3.5-turbo"],
      "enabled": true
    }
  ]
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

OpenAI provides various model families optimized for different use cases:

| Model | Context Length | Best For | Function Calling |
|-------|----------------|----------|------------------|
| `gpt-4o` | 128,000 tokens | Most capable, multimodal | ✅ Yes |
| `gpt-4o-mini` | 128,000 tokens | Cost-effective, fast | ✅ Yes |
| `gpt-4-turbo` | 128,000 tokens | High quality, vision support | ✅ Yes |
| `gpt-4` | 8,192 tokens | Classic GPT-4 | ✅ Yes |
| `gpt-3.5-turbo` | 16,385 tokens | Fast, affordable | ✅ Yes |
| `o1-preview` | 128,000 tokens | Complex reasoning | ❌ No* |
| `o1-mini` | 128,000 tokens | Fast reasoning | ❌ No* |

**⚠️ Important**: Claude Code requires models with function calling support. Models marked with ✅ work perfectly with Claude Code.

*Note: o1 models don't support function calling but can be used for pure text generation tasks.

## Configuration Options

### Complete Configuration Example

```json
{
  "providers": [
    {
      "name": "openai",
      "api_base_url": "https://api.openai.com/v1",
      "api_key": "sk-your_api_key",
      "models": ["gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-3.5-turbo"],
      "enabled": true,
      "timeout": "60s",
      "max_retries": 3,
      "organization": "org-your_org_id",
      "headers": {
        "OpenAI-Beta": "assistants=v2"
      }
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "gpt-4o-mini"
    },
    "complex": {
      "provider": "openai",
      "model": "gpt-4o",
      "conditions": [{
        "type": "tokenCount",
        "operator": ">",
        "value": 8000
      }]
    }
  }
}
```

### Using Environment Variables

While CCProxy primarily uses config.json, you can use environment variables for API key security:

```bash
# Store API key in environment
export OPENAI_API_KEY="sk-your_api_key"

# Reference in config.json
{
  "providers": [{
    "name": "openai",
    "api_key": "${OPENAI_API_KEY}",
    ...
  }]
}
```

## Advanced Features

### Vision Capabilities

GPT-4o and GPT-4-turbo support image inputs:

```json
{
  "messages": [
    {
      "role": "user",
      "content": [
        {"type": "text", "text": "What's in this image?"},
        {"type": "image_url", "image_url": {"url": "data:image/jpeg;base64,..."}}
      ]
    }
  ]
}
```

### Function Calling

All OpenAI models (except o1 series) support function calling, making them perfect for Claude Code:

```json
{
  "model": "gpt-4o",
  "messages": [...],
  "tools": [{
    "type": "function",
    "function": {
      "name": "get_weather",
      "description": "Get the current weather",
      "parameters": {...}
    }
  }]
}
```

## Troubleshooting

### Rate Limits
OpenAI has different rate limits based on your usage tier:
- Monitor usage at [platform.openai.com/usage](https://platform.openai.com/usage)
- Consider upgrading your plan for higher limits
- Implement exponential backoff for rate limit errors

### API Errors
Common errors and solutions:
- **401 Unauthorized**: Check your API key
- **429 Too Many Requests**: You've hit rate limits
- **500 Server Error**: OpenAI service issue, retry with backoff

### Organization Access
If using an organization:
1. Ensure your API key has access to the organization
2. Add organization ID to your configuration
3. Check organization settings at platform.openai.com

## Best Practices

1. **Model Selection**: 
   - Use `gpt-4o-mini` for most tasks (best price/performance)
   - Use `gpt-4o` for complex or multimodal tasks
   - Use `gpt-3.5-turbo` for simple, high-volume tasks

2. **Cost Management**:
   - Set usage limits in OpenAI dashboard
   - Monitor token usage in responses
   - Use smaller models when possible

3. **Performance Optimization**:
   - Implement caching for repeated queries
   - Use streaming for better perceived performance
   - Batch requests when possible

## Pricing

Current pricing (subject to change):
- **GPT-4o**: $5.00 / 1M input tokens, $15.00 / 1M output tokens
- **GPT-4o-mini**: $0.15 / 1M input tokens, $0.60 / 1M output tokens
- **GPT-3.5-turbo**: $0.50 / 1M input tokens, $1.50 / 1M output tokens

Check [openai.com/pricing](https://openai.com/pricing) for current rates.

## Next Steps

- [Quick Start Guide](/guide/quick-start) - Get running in 2 minutes
- [Configuration Guide](/guide/configuration) - Advanced configuration options
- [Provider Comparison](/providers/) - Compare with other providers