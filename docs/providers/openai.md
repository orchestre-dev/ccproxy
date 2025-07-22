---
title: OpenAI with Claude Code - Industry-Leading AI Models via CCProxy
description: Access GPT-4o, GPT-4.1, and o3/o4 reasoning models with Claude Code through CCProxy. Experience enterprise-grade AI with vision capabilities, function calling, and production reliability.
keywords: OpenAI, Claude Code, CCProxy, GPT-4o, GPT-4.1, o3, o4-mini, AI proxy, enterprise AI, vision AI, function calling
---

# OpenAI Provider

**OpenAI sets the industry standard** for AI models, offering the most mature ecosystem and reliable performance for production applications. Through **CCProxy integration with Claude Code**, you can access GPT-4o, GPT-4.1, and the latest o3/o4 reasoning models while maintaining your familiar development workflow.

## 🏭 Why Choose OpenAI for Claude Code?

- 🥇 **Industry standard**: Most mature and reliable AI models with proven enterprise adoption
- 🛠️ **Complete parameter support**: Temperature, top_p, presence_penalty, frequency_penalty all work perfectly
- 👁️ **Advanced vision**: Best-in-class image understanding with GPT-4o and GPT-4.1
- 🎯 **Function calling**: Full support for Claude Code's tool use capabilities
- 🧠 **Reasoning models**: Access to o3 and o4-mini for complex problem-solving (text-only)
- 🔧 **Zero-config integration**: Works immediately with CCProxy, no modifications needed

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
      "models": ["gpt-4o", "gpt-4o-mini", "gpt-4.1", "gpt-4.1-mini", "gpt-4.1-nano", "o3", "o3-pro", "o4-mini"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "gpt-4o"
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

OpenAI provides various model families optimized for different use cases:

### Latest Models (July 2025)

| Model | Context Length | Best For | Function Calling |
|-------|----------------|----------|------------------|
| `gpt-4o` | 128K tokens | Flagship model with vision, most versatile | ✅ Yes |
| `gpt-4o-mini` | 128K tokens | Smaller, cost-efficient version of GPT-4o | ✅ Yes |
| `gpt-4.1` | 128K tokens | Specialized for coding tasks, precise instruction following | ✅ Yes |
| `gpt-4.1-mini` | 128K tokens | Smaller variant of GPT-4.1 | ✅ Yes |
| `gpt-4.1-nano` | 128K tokens | Smallest, most efficient GPT-4.1 variant | ✅ Yes |
| `o3` | 128K tokens | Most powerful reasoning model, excels at complex problems | ❌ No* |
| `o3-pro` | 128K tokens | Extended thinking version of o3 | ❌ No* |
| `o4-mini` | 128K tokens | Fast, cost-efficient reasoning model | ❌ No* |

### Legacy Models

| Model | Status | Notes |
|-------|---------|-------|
| `gpt-4-turbo` | Available | Previous generation, consider GPT-4o |
| `gpt-4` | Available | Original GPT-4 |
| `gpt-3.5-turbo` | Available | Budget option, fastest response times |

**⚠️ Important**: Claude Code requires models with function calling support. Models marked with ✅ work perfectly with Claude Code.

*Note: o3/o4 reasoning models don't support function calling but can be used for pure text generation tasks when tool use is not required.

## Configuration Options

### Complete Configuration Example

```json
{
  "providers": [
    {
      "name": "openai",
      "api_base_url": "https://api.openai.com/v1",
      "api_key": "sk-your_api_key",
      "models": ["gpt-4o", "gpt-4o-mini", "gpt-4.1", "gpt-4.1-mini", "gpt-4.1-nano", "o3", "o3-pro", "o4-mini"],
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
      "model": "gpt-4o"
    },
    "longContext": {
      "provider": "openai",
      "model": "gpt-4o",
      "conditions": [{
        "type": "tokenCount",
        "operator": ">",
        "value": 60000
      }]
    },
    "background": {
      "provider": "openai",
      "model": "gpt-4.1-nano"
    },
    "think": {
      "provider": "openai",
      "model": "o3"
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

### Supported Parameters

OpenAI has the most complete parameter support in CCProxy:

```json
{
  "model": "gpt-4o",
  "messages": [...],
  "temperature": 0.7,          // 0.0 to 2.0
  "top_p": 0.9,               // 0.0 to 1.0
  "presence_penalty": 0.0,     // -2.0 to 2.0
  "frequency_penalty": 0.0,    // -2.0 to 2.0
  "max_tokens": 4096,          // Maximum tokens to generate
  "stream": true,              // Enable streaming responses
  "n": 1,                      // Number of completions
  "stop": ["\\n", "END"],      // Stop sequences
  "logprobs": true             // Include log probabilities
}
```

### Vision Capabilities

GPT-4o and GPT-4.1 series models support image inputs:

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

All GPT models support function calling, making them perfect for Claude Code:

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
   - Use `gpt-4o` for most tasks (flagship model with vision support)
   - Use `gpt-4.1` for coding-specific tasks (best instruction following)
   - Use `gpt-4o-mini` or `gpt-4.1-nano` for cost-efficient tasks
   - Use `o3` or `o3-pro` for complex reasoning (no function calling)
   - Use `o4-mini` for fast reasoning tasks
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

Check [openai.com/pricing](https://openai.com/pricing) for current pricing. As of July 2025:

- **GPT-4o**: Flagship model with best capabilities
- **GPT-4o-mini**: Cost-efficient version of GPT-4o
- **GPT-4.1 family**: Specialized coding models at various price points
- **o3/o3-pro**: Premium reasoning models (higher cost)
- **o4-mini**: Cost-efficient reasoning model
- **GPT-3.5-turbo**: Most affordable option for simple tasks

Pricing varies by model and usage tier. Enterprise customers should contact OpenAI for volume pricing.

## Next Steps

- [Quick Start Guide](/guide/quick-start) - Get running in 2 minutes
- [Configuration Guide](/guide/configuration) - Advanced configuration options
- [Provider Comparison](/providers/) - Compare with other providers