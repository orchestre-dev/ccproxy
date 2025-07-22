---
title: Quick Start - Get CCProxy Running in 2 Minutes
description: Get CCProxy running with Claude Code in under 2 minutes. Fast setup guide for immediate AI development productivity.
keywords: CCProxy quick start, Claude Code integration, AI proxy setup
---

# Quick Start

<SocialShare />

Get CCProxy running with Claude Code in under 2 minutes.

## 1. Download CCProxy

Download the latest binary for your platform from the [releases page](https://github.com/orchestre-dev/ccproxy/releases).

```bash
# Example for macOS/Linux
chmod +x ccproxy
```

## 2. Configure and Start

Create a `config.json` file:

```json
{
  "host": "127.0.0.1",
  "port": 3456,
  "providers": [{
    "name": "anthropic",
    "api_key": "your-anthropic-api-key",
    "enabled": true
  }]
}
```

Start CCProxy:
```bash
./ccproxy start
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
      "enabled": true
    },
    {
      "name": "openai",
      "api_key": "sk-...",
      "enabled": true
    },
    {
      "name": "gemini",
      "api_key": "AI...",
      "enabled": true
    }
  ]
}
```

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