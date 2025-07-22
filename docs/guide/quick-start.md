---
title: Quick Start - Get CCProxy Running in 2 Minutes
description: Get CCProxy running with Claude Code in under 2 minutes. Fast setup guide for immediate AI development productivity.
keywords: CCProxy quick start, Claude Code integration, AI proxy setup
---

# Quick Start

<SocialShare />

Get CCProxy running with Claude Code in under 2 minutes.

## 1. Install CCProxy

Install with one command:

```bash
curl -sSL https://raw.githubusercontent.com/orchestre-dev/ccproxy/main/install.sh | bash
```

Or download manually from the [releases page](https://github.com/orchestre-dev/ccproxy/releases).

## 2. Configure and Start

Create a configuration file:

```bash
mkdir -p ~/.ccproxy
cat > ~/.ccproxy/config.json << 'EOF'
{
  "providers": [{
    "name": "openai",
    "api_key": "your-openai-api-key",
    "models": ["gpt-4o", "gpt-4o-mini"],
    "enabled": true
  }],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "gpt-4o"
    }
  }
}
EOF
```

**Important**: The `models` array lists available models for validation. The `routes` section defines which provider and model handle your requests.

Start CCProxy:
```bash
ccproxy start
```

## 3. Connect Claude Code

Use the auto-configuration:
```bash
./ccproxy code
```

Or configure manually:
```bash
export ANTHROPIC_BASE_URL=http://localhost:3456
export ANTHROPIC_AUTH_TOKEN=test
claude "Write a Python function to reverse a string"
```

## 🎉 Done!

Claude Code now uses your configured AI provider. Try:

```bash
claude "Explain this code and suggest improvements" < your-file.py
claude "Create a REST API for user management"
claude "Debug this error: TypeError: 'int' object is not subscriptable"
```

## Multiple Providers

Add more providers to your config.json:

```json
{
  "providers": [
    {
      "name": "anthropic",
      "api_key": "sk-ant-...",
      "models": ["claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022"],
      "enabled": true
    },
    {
      "name": "openai",
      "api_key": "sk-...",
      "models": ["gpt-4o", "gpt-4o-mini"],
      "enabled": true
    },
    {
      "name": "gemini",
      "api_key": "AI...",
      "models": ["gemini-2.0-flash-exp", "gemini-1.5-pro"],
      "enabled": true
    },
    {
      "name": "deepseek",
      "api_key": "sk-...",
      "models": ["deepseek-chat", "deepseek-coder"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "gpt-4o"
    },
    "longContext": {
      "provider": "anthropic",
      "model": "claude-3-5-sonnet-20241022"
    }
  }
}
```

The `routes` section controls which provider handles different types of requests. Requests with >60K tokens automatically use the `longContext` route.

**Note:** For Claude Code integration, ensure your selected models support function calling. Most modern models from major providers (Anthropic Claude, OpenAI, Google Gemini, DeepSeek) include this capability.

## Understanding Model Selection

CCProxy uses intelligent routing to select the appropriate model based on your request:

1. **Explicit model routes** - If you define a route with the exact model name, it uses that
2. **Long context routing** - Requests exceeding 60,000 tokens automatically use the `longContext` route
3. **Background routing** - Claude Haiku models (claude-3-5-haiku-*) use the `background` route if defined
4. **Thinking mode** - Requests with `thinking: true` parameter use the `think` route if defined
5. **Default routing** - All other requests use the `default` route

💡 **Tip:** For latest model information, check [models.dev](https://models.dev)

## Next Steps

- **[Full Installation Guide](/guide/installation)** - Multiple installation methods
- **[Configuration Guide](/guide/configuration)** - Advanced provider setup
- **[Provider Guide](/providers/)** - Supported AI providers

## Troubleshooting

**Connection refused?** Check if CCProxy is running:
```bash
./ccproxy status
```

**API key error?** Verify your provider API key in config.json.

**Need help?** [💬 GitHub Discussions](https://github.com/orchestre-dev/ccproxy/discussions) • [🐛 Report Issues](https://github.com/orchestre-dev/ccproxy/issues)