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

The installer will:
- ✅ Install ccproxy to `/usr/local/bin`
- ✅ Create `~/.ccproxy` directory
- ✅ Generate a starter configuration file
- ✅ Update your PATH if needed
- ✅ Show clear next steps

Or download manually from the [releases page](https://github.com/orchestre-dev/ccproxy/releases).

## 2. Configure Your API Key

The installer creates a configuration file at `~/.ccproxy/config.json`. Edit it to add your API key:

```bash
# Open in your preferred editor
nano ~/.ccproxy/config.json    # or vim, code, etc.
```

Replace `your-openai-api-key-here` with your actual API key:

```json
{
  "providers": [{
    "name": "openai",
    "api_key": "sk-proj-...",  // <- Your actual API key here
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
```

## 3. Start CCProxy

```bash
ccproxy start
```

You'll see:
```
Starting CCProxy on port 3456...
✅ Server started successfully
```

## 4. Connect Claude Code

The easiest way:
```bash
ccproxy code
```

This command:
- Sets environment variables for Claude Code
- Verifies the connection
- Shows you're ready to go

Or configure manually:
```bash
export ANTHROPIC_BASE_URL=http://localhost:3456
export ANTHROPIC_AUTH_TOKEN=test
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

### "ccproxy: command not found"
The installer adds `/usr/local/bin` to your PATH. Try:
```bash
# Reload your shell configuration
source ~/.bashrc    # or ~/.zshrc for zsh

# Or use the full path
/usr/local/bin/ccproxy start
```

### "Connection refused"
Check if CCProxy is running:
```bash
ccproxy status
```

If not running, start it:
```bash
ccproxy start
```

### "API key error"
1. Check your configuration file:
   ```bash
   cat ~/.ccproxy/config.json
   ```

2. Ensure you replaced `your-openai-api-key-here` with your actual API key

3. Verify the API key format:
   - OpenAI: Starts with `sk-`
   - Anthropic: Starts with `sk-ant-`
   - Google: Starts with `AI`

### "Config file not found"
The config should be at `~/.ccproxy/config.json`. If missing, create it:
```bash
mkdir -p ~/.ccproxy
ccproxy init    # Coming soon - for now, copy from examples above
```

**Need help?** [💬 GitHub Discussions](https://github.com/orchestre-dev/ccproxy/discussions) • [🐛 Report Issues](https://github.com/orchestre-dev/ccproxy/issues)