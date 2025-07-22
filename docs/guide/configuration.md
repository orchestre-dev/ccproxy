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
      "enabled": true
    }
  ]
}
```

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

CCProxy can route requests based on various criteria:

### Default Routing
```json
{
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
    }
  }
}
```

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

# Test message
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [{"role": "user", "content": "Hello!"}],
    "max_tokens": 100
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

**Rate Limiting:**
- Switch to a different provider
- Implement retry logic
- Consider upgrading your API plan

### Getting Help

- 📖 [Installation Guide](/guide/installation)
- 🚀 [Quick Start](/guide/quick-start)  
- 🔧 [Provider Guide](/providers/)
- 💬 [Ask Questions](https://github.com/orchestre-dev/ccproxy/discussions) - Community support
- 🐛 [Report Issues](https://github.com/orchestre-dev/ccproxy/issues) - Bug reports and feature requests