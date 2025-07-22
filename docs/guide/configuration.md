---
title: Configuration Guide - CCProxy Provider Setup
description: Complete configuration guide for CCProxy. Set up Anthropic, OpenAI, Gemini, DeepSeek, and OpenRouter providers with Claude Code.
keywords: CCProxy configuration, Anthropic setup, OpenAI config, Gemini setup, DeepSeek config, OpenRouter, Claude Code integration
---

# Configuration Guide

<SocialShare />

Configure CCProxy to work with your preferred AI providers. CCProxy uses a JSON configuration file to manage providers, routing, and server settings.

## Configuration File

CCProxy uses a `config.json` file for all configuration. The file is typically located at:
- Linux/macOS: `~/.ccproxy/config.json`
- Windows: `%USERPROFILE%\.ccproxy\config.json`
- Or specify a custom location with `--config` flag

## Basic Configuration

### Minimal Configuration

```json
{
  "host": "127.0.0.1",
  "port": 3456,
  "providers": [
    {
      "name": "anthropic",
      "api_key": "sk-ant-...",
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

**Note**: The `routes` section is required. Without it, CCProxy cannot determine which provider and model to use for requests.

### Full Configuration Example

```json
{
  "host": "127.0.0.1",
  "port": 3456,
  "log": true,
  "apikey": "your-ccproxy-api-key",
  "performance": {
    "request_timeout": "30s",
    "max_request_body_size": 10485760,
    "metrics_enabled": true,
    "rate_limit_enabled": false,
    "circuit_breaker_enabled": true
  },
  "security": {
    "audit_enabled": false,
    "ip_whitelist": ["127.0.0.1"],
    "allowed_headers": ["Content-Type", "Accept", "Authorization"]
  },
  "providers": [
    {
      "name": "anthropic",
      "api_key": "${ANTHROPIC_API_KEY}",
      "api_base_url": "https://api.anthropic.com",
      "models": ["claude-3-opus-20240229", "claude-3-sonnet-20240229"],
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

## Provider Configuration

### 🎯 Anthropic (Claude)

**Native support for all Claude models**

```json
{
  "providers": [
    {
      "name": "anthropic",
      "api_key": "sk-ant-...",
      "api_base_url": "https://api.anthropic.com",
      "models": [
        "claude-3-opus-20240229",
        "claude-3-sonnet-20240229",
        "claude-3-haiku-20240307"
      ],
      "enabled": true
    }
  ]
}
```

**Get your API key:** [console.anthropic.com](https://console.anthropic.com/)

### 🤖 OpenAI

**GPT-4 and GPT-3.5 models with full compatibility**

```json
{
  "providers": [
    {
      "name": "openai",
      "api_key": "sk-...",
      "api_base_url": "https://api.openai.com/v1",
      "models": [
        "gpt-4",
        "gpt-4-turbo-preview",
        "gpt-3.5-turbo"
      ],
      "enabled": true
    }
  ]
}
```

**Get your API key:** [platform.openai.com](https://platform.openai.com/)

### 🌟 Google Gemini

**Advanced multimodal models**

```json
{
  "providers": [
    {
      "name": "gemini",
      "api_key": "AI...",
      "api_base_url": "https://generativelanguage.googleapis.com",
      "models": [
        "gemini-1.5-flash",
        "gemini-1.5-pro"
      ],
      "enabled": true
    }
  ]
}
```

**Get your API key:** [makersuite.google.com](https://makersuite.google.com/app/apikey)

### 💻 DeepSeek

**Specialized coding models**

```json
{
  "providers": [
    {
      "name": "deepseek",
      "api_key": "sk-...",
      "api_base_url": "https://api.deepseek.com",
      "models": [
        "deepseek-coder",
        "deepseek-chat"
      ],
      "enabled": true
    }
  ]
}
```

**Get your API key:** [platform.deepseek.com](https://platform.deepseek.com/)

### 🌐 OpenRouter

**Access to 100+ models through unified API**

```json
{
  "providers": [
    {
      "name": "openrouter",
      "api_key": "sk-or-v1-...",
      "api_base_url": "https://openrouter.ai/api/v1",
      "models": [
        "anthropic/claude-3.5-sonnet",
        "google/gemini-pro-1.5",
        "meta-llama/llama-3.1-405b-instruct"
      ],
      "enabled": true
    }
  ]
}
```

**Get your API key:** [openrouter.ai](https://openrouter.ai/)

## Environment Variables

CCProxy supports environment variable substitution in configuration files:

```json
{
  "providers": [
    {
      "name": "anthropic",
      "api_key": "${ANTHROPIC_API_KEY}",
      "enabled": true
    }
  ]
}
```

Then set the environment variable:
```bash
export ANTHROPIC_API_KEY="sk-ant-..."
./ccproxy start
```

## Routing Configuration

CCProxy uses routes to determine which provider and model handle your requests. This is separate from the `models` array in provider configuration, which is used for validation.

### Understanding Routes vs Models

1. **`models` array** (in providers): Lists available models for validation
2. **`routes` section**: Defines which provider/model actually handles requests

## Model Selection Explained

The `models` array in provider config lists available models for validation.
The `routes` section defines which provider/model actually handles requests.

### Example: Multi-Provider Setup
```json
{
  "providers": [
    {
      "name": "anthropic",
      "models": ["claude-opus-4-20250720", "claude-sonnet-4-20250720"],
      "api_key": "..."
    },
    {
      "name": "openai", 
      "models": ["gpt-4.1", "o3", "o4-mini"],
      "api_key": "..."
    }
  ],
  "routes": {
    "default": {
      "provider": "anthropic",
      "model": "claude-sonnet-4-20250720"
    },
    "gpt-4": {  // Direct route for compatibility
      "provider": "openai",
      "model": "gpt-4.1"
    }
  }
}
```

## Model Currency

It's important to keep your model configurations up-to-date with the latest available models. AI providers frequently release new models with improved capabilities, better performance, and lower costs.

### Why Model Currency Matters
- **Better Performance**: Newer models often provide faster responses and better quality outputs
- **Cost Efficiency**: Latest models may offer better pricing tiers
- **Feature Support**: New models include the latest features like improved context windows, better tool use, and enhanced reasoning capabilities
- **Deprecation**: Older models are eventually deprecated and removed from service

### Staying Current
Check [models.dev](https://models.dev) for the latest model information across all providers. This resource provides:
- Current model names and versions
- Pricing information
- Context window sizes
- Feature support comparison
- Deprecation notices

Update your CCProxy configuration regularly to ensure you're using the best available models for your use case.

### Default Routing
```json
{
  "providers": [{
    "name": "anthropic",
    "api_key": "sk-ant-...",
    "models": ["claude-3-sonnet-20240229", "claude-3-opus-20240229"],
    "enabled": true
  }],
  "routes": {
    "default": {
      "provider": "anthropic",
      "model": "claude-3-sonnet-20240229"
    }
  }
}
```

### Context-Based Routing
Requests with >60K tokens automatically route to long-context models:
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
    }
  }
}
```

### Model-Specific Routing
When a request specifies a model name, CCProxy looks for a matching route:
```json
{
  "routes": {
    "claude-3-opus": {
      "provider": "anthropic",
      "model": "claude-3-opus-20240229"
    },
    "gpt-4": {
      "provider": "openai",
      "model": "gpt-4"
    },
    "default": {
      "provider": "openai",
      "model": "gpt-3.5-turbo"
    }
  }
}
```

**Important**: Without routes configuration, CCProxy cannot determine which provider to use. Always define at least a `default` route.

## Performance Configuration

Optimize CCProxy performance:

```json
{
  "performance": {
    "request_timeout": "30s",
    "max_request_body_size": 10485760,
    "metrics_enabled": true,
    "rate_limit_enabled": false,
    "circuit_breaker_enabled": true,
    "circuit_breaker_threshold": 5,
    "circuit_breaker_timeout": "60s"
  }
}
```

## Security Configuration

Configure security settings:

```json
{
  "apikey": "your-secure-api-key",
  "security": {
    "audit_enabled": true,
    "ip_whitelist": ["127.0.0.1", "192.168.1.0/24"],
    "ip_blacklist": ["10.0.0.0/8"],
    "allowed_headers": ["Content-Type", "Accept", "Authorization"],
    "cors_enabled": true,
    "cors_origins": ["http://localhost:3000"]
  }
}
```

## Multiple Configurations

Manage different environments with separate configuration files:

### Development Configuration
```json
{
  "host": "127.0.0.1",
  "port": 3456,
  "log": true,
  "providers": [{
    "name": "anthropic",
    "api_key": "${ANTHROPIC_API_KEY}",
    "enabled": true
  }]
}
```

### Production Configuration
```json
{
  "host": "0.0.0.0",
  "port": 443,
  "apikey": "${CCPROXY_API_KEY}",
  "log": false,
  "performance": {
    "rate_limit_enabled": true,
    "circuit_breaker_enabled": true
  },
  "security": {
    "audit_enabled": true,
    "ip_whitelist": ["10.0.0.0/8"]
  }
}
```

Use different configs:
```bash
./ccproxy start --config config.dev.json
./ccproxy start --config config.prod.json
```

## Claude Code Integration

Once CCProxy is configured and running, integrate with Claude Code:

### Automatic Setup
```bash
./ccproxy code
```

### Manual Setup
```bash
# Set Claude Code to use CCProxy
export ANTHROPIC_BASE_URL=http://localhost:3456
export ANTHROPIC_AUTH_TOKEN=test
export API_TIMEOUT_MS=600000

# Verify it's working
claude "Hello, can you help me with coding?"
```

## Request Parameters

### Standard Parameters

CCProxy supports standard request parameters that work across all providers:

```json
{
  "model": "claude-3-sonnet-20240229",
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "max_tokens": 1024,
  "temperature": 0.7,
  "top_p": 0.9,
  "top_k": 40,
  "stream": false,
  "stop_sequences": ["\n\n"],
  "tools": [],
  "tool_choice": "auto"
}
```

**Supported Parameters:**
- `model` - Model identifier (required)
- `messages` - Array of message objects (required)
- `max_tokens` - Maximum tokens to generate
- `temperature` - Randomness (0-2)
- `top_p` - Nucleus sampling threshold
- `top_k` - Top-k sampling
- `stream` - Enable streaming responses
- `stop_sequences` - Stop generation at these sequences
- `tools` - Function calling definitions
- `tool_choice` - How to use tools ("auto", "none", or specific tool)

**Note:** Provider-specific parameters like `thinkingBudget`, `frequency_penalty`, and `presence_penalty` are not supported and will be ignored or cause errors.

### Parameter Mapping

CCProxy automatically maps parameters for different providers:

#### Gemini Parameter Mapping
- `max_tokens` → `maxOutputTokens` (wrapped in `generationConfig`)
- All generation parameters wrapped in `generationConfig` object
- Example transformation:
  ```json
  // Input
  {
    "max_tokens": 1024,
    "temperature": 0.7
  }
  
  // Transformed for Gemini
  {
    "generationConfig": {
      "maxOutputTokens": 1024,
      "temperature": 0.7
    }
  }
  ```

#### Anthropic Parameter Handling
- Removes unsupported parameters: `frequency_penalty`, `presence_penalty`
- Native support for all other standard parameters
- Supports Anthropic-specific features like system messages

#### DeepSeek Constraints
- Enforces maximum `max_tokens` limit of 8192
- If you request more than 8192 tokens, it will be capped
- Supports all standard OpenAI-compatible parameters

#### OpenAI Compatibility
- Full support for all standard parameters
- Native `frequency_penalty` and `presence_penalty` support
- Compatible with GPT-3.5 and GPT-4 models

### Function Calling

Function calling (tools) requires specific formatting:

```json
{
  "messages": [
    {"role": "user", "content": "What's the weather in Paris?"}
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get current weather",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {
              "type": "string",
              "description": "City name"
            }
          },
          "required": ["location"]
        }
      }
    }
  ],
  "tool_choice": "auto"
}
```

**Provider Support:**
- ✅ Anthropic - Full support
- ✅ OpenAI - Full support
- ✅ Gemini - Full support (transformed to Gemini format)
- ❌ DeepSeek - Limited support
- ✅ OpenRouter - Depends on underlying model

## Advanced Configuration

### Complete Configuration Reference

```json
{
  "host": "127.0.0.1",
  "port": 3456,
  "log": true,
  "apikey": "your-ccproxy-api-key",
  
  "performance": {
    "request_timeout": "30s",
    "max_request_body_size": 10485760,
    "metrics_enabled": true,
    "rate_limit_enabled": false,
    "rate_limit_requests_per_minute": 60,
    "circuit_breaker_enabled": true,
    "circuit_breaker_threshold": 5,
    "circuit_breaker_timeout": "60s",
    "cache_enabled": true,
    "cache_ttl": "5m"
  },
  
  "security": {
    "audit_enabled": false,
    "audit_log_path": "~/.ccproxy/audit.log",
    "ip_whitelist": [],
    "ip_blacklist": [],
    "allowed_headers": ["Content-Type", "Accept", "Authorization"],
    "cors_enabled": true,
    "cors_origins": ["*"],
    "cors_credentials": false
  },
  
  "providers": [
    {
      "name": "anthropic",
      "api_key": "${ANTHROPIC_API_KEY}",
      "api_base_url": "https://api.anthropic.com",
      "models": ["claude-3-opus-20240229", "claude-3-sonnet-20240229"],
      "enabled": true,
      "priority": 1,
      "timeout": "30s",
      "max_retries": 3
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
  },
  
  "middleware": {
    "request_id": true,
    "logging": true,
    "recovery": true,
    "timeout": true,
    "cors": true,
    "security": true,
    "performance": true
  }
}
```

### Configuration Directory Structure

```
~/.ccproxy/
├── config.json          # Main configuration
├── config.dev.json      # Development config
├── config.prod.json     # Production config
├── ccproxy.log         # Log file (when enabled)
├── audit.log           # Audit log (when enabled)
└── .ccproxy.pid        # Process ID file
```

### Dynamic Configuration Updates

Update provider configuration without restarting:

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

# Disable a provider
curl -X PUT http://localhost:3456/providers/openai \
  -H "x-api-key: your-ccproxy-api-key" \
  -d '{"enabled": false}'

# Remove a provider
curl -X DELETE http://localhost:3456/providers/openai \
  -H "x-api-key: your-ccproxy-api-key"
```

## Validation

Test your configuration:

```bash
# Health check
curl http://localhost:3456/health

# Provider status
curl http://localhost:3456/status

# Test message with standard parameters only
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet-20240229",
    "messages": [{"role": "user", "content": "Hello!"}],
    "max_tokens": 100,
    "temperature": 0.7
  }'

# Test streaming
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet-20240229",
    "messages": [{"role": "user", "content": "Tell me a short story"}],
    "max_tokens": 200,
    "stream": true
  }'

# Test function calling
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet-20240229",
    "messages": [{"role": "user", "content": "What is 2+2?"}],
    "tools": [{
      "type": "function",
      "function": {
        "name": "calculate",
        "description": "Perform calculations",
        "parameters": {
          "type": "object",
          "properties": {
            "expression": {"type": "string"}
          }
        }
      }
    }]
  }'
```

## Troubleshooting

### Common Issues

**API Key Invalid:**
- Verify your API key is correct
- Check for extra spaces or newlines
- Ensure the key has proper permissions

**Connection Refused:**
- Check if the provider's API is accessible
- Verify your network connection
- Try a different provider

**Model Not Found:**
- Verify the model name is correct
- Check if the model is available for your API key
- Try the default model for your provider

**Parameter Errors:**
- Use only standard parameters (avoid provider-specific ones like `thinkingBudget`)
- Check parameter names match exactly (case-sensitive)
- Verify parameter values are within acceptable ranges
- For Gemini, parameters are automatically wrapped in `generationConfig`
- For DeepSeek, `max_tokens` is capped at 8192

**Rate Limiting:**
- Switch to a different provider
- Implement retry logic
- Consider upgrading your API plan

**Function Calling Errors:**
- Ensure tools array is properly formatted
- Check that the provider supports function calling
- Verify tool_choice is set to "auto", "none", or a specific tool name

### Getting Help

- 📖 [Installation Guide](/guide/installation)
- 🚀 [Quick Start](/guide/quick-start)  
- 🔧 [Provider Guide](/providers/)
- 💬 [Ask Questions](https://github.com/orchestre-dev/ccproxy/discussions) - Community support
- 🐛 [Report Issues](https://github.com/orchestre-dev/ccproxy/issues) - Bug reports and feature requests