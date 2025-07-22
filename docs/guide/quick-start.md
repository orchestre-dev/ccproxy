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
    "api_base_url": "https://api.openai.com/v1",
    "api_key": "your-openai-api-key",
    "models": ["gpt-4", "gpt-3.5-turbo"],
    "enabled": true
  }]
}
EOF
```

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