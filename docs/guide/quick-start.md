---
title: Quick Start - Get CCProxy Running in 2 Minutes
description: Get CCProxy running with Claude Code in under 2 minutes. Fast setup guide for immediate AI development productivity.
keywords: CCProxy quick start, Claude Code integration, AI proxy setup
---

# Quick Start

<SocialShare />

Get CCProxy running with Claude Code in under 2 minutes.

## 1. Install CCProxy

### macOS/Linux

Install with one command:

```bash
curl -sSL https://raw.githubusercontent.com/orchestre-dev/ccproxy/main/install.sh | bash
```

The installer will:
- ✅ Install ccproxy binary to `/usr/local/bin/ccproxy`
- ✅ Create configuration directory at `/Users/yourname/.ccproxy` (macOS) or `/home/yourname/.ccproxy` (Linux)
- ✅ Generate starter config at `~/.ccproxy/config.json` with example API key placeholder
- ✅ Add `/usr/local/bin` to PATH in `.bashrc` or `.zshrc` if not already present
- ✅ Display exact next steps with your specific file paths

### Windows

**Option 1: Automated Installation (Recommended)**

In PowerShell:
```powershell
# Download and run the installer
irm https://raw.githubusercontent.com/orchestre-dev/ccproxy/main/install.ps1 | iex

# Or download first, then run
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/orchestre-dev/ccproxy/main/install.ps1" -OutFile install.ps1
.\install.ps1
```

The installer will:
- ✅ Download latest ccproxy.exe to `C:\Program Files\CCProxy\`
- ✅ Create configuration directory at `C:\Users\YourName\.ccproxy\`
- ✅ Generate starter config with example API key placeholder
- ✅ Add CCProxy to your PATH automatically
- ✅ Display exact next steps with your specific file paths

**Option 2: Manual Installation**

1. Go to [releases page](https://github.com/orchestre-dev/ccproxy/releases/latest)
2. Download `ccproxy-windows-amd64.exe`
3. Create folder: `C:\Program Files\CCProxy`
4. Move downloaded file there and rename to `ccproxy.exe`
5. Add `C:\Program Files\CCProxy` to your PATH
6. Create the configuration directory and file:
```powershell
# In PowerShell
# Create directory
mkdir $env:USERPROFILE\.ccproxy

# Create config file with starter template
@'
{
  "providers": [{
    "name": "openai",
    "api_key": "your-openai-api-key-here",
    "api_base_url": "https://api.openai.com/v1",
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
'@ | Out-File -FilePath "$env:USERPROFILE\.ccproxy\config.json" -Encoding UTF8
```

This creates:
- Config directory at `C:\Users\YourName\.ccproxy`
- Config file at `C:\Users\YourName\.ccproxy\config.json`

## 2. Configure Your API Key

### Option A: Use Example Configuration (Recommended)

We provide ready-to-use configurations:

```bash
# Copy an example configuration
cp examples/configs/openai-gpt4.json ~/.ccproxy/config.json

# Set your API key via environment variable
export OPENAI_API_KEY="sk-your-openai-key"
```

Available examples:
- **`openai-gpt4.json`** - Latest GPT-4.1 models
- **`openai-o-series.json`** - O-series reasoning models (O3, O1)
- **`openai-mixed.json`** - Multi-provider setup
- **`openai-budget.json`** - Cost-optimized configuration

### Option B: Edit Config Manually

Find your config file:
- **macOS**: `/Users/YourName/.ccproxy/config.json`
- **Linux**: `/home/YourName/.ccproxy/config.json`
- **Windows**: `C:\Users\YourName\.ccproxy\config.json`

**macOS/Linux:**
```bash
# Using nano (simple editor)
nano ~/.ccproxy/config.json

# Or using VS Code
code ~/.ccproxy/config.json

# Or using vim
vim ~/.ccproxy/config.json
```

**Windows (PowerShell):**
```powershell
# Using Notepad
notepad $env:USERPROFILE\.ccproxy\config.json

# Or using VS Code
code $env:USERPROFILE\.ccproxy\config.json
```

Replace `your-openai-api-key-here` with your actual API key:

```json
{
  "providers": [{
    "name": "openai",
    "api_key": "sk-proj-...",  // <- Your actual API key here
    "models": ["gpt-4.1", "gpt-4.1-mini", "gpt-4o"],
    "enabled": true
  }],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "gpt-4.1"
    }
  }
}
```

## 3. Start CCProxy

**All platforms:**
```bash
ccproxy start
```

You'll see:
```
Starting CCProxy on port 3456...
✅ Server started successfully
```

## 4. Connect Claude Code

**All platforms:**
```bash
ccproxy code
```

This command:
- Sets environment variables for Claude Code
- Verifies the connection
- Shows you're ready to go

**Manual configuration:**

macOS/Linux:
```bash
export ANTHROPIC_BASE_URL=http://localhost:3456
export ANTHROPIC_AUTH_TOKEN=test
```

Windows (PowerShell):
```powershell
$env:ANTHROPIC_BASE_URL = "http://localhost:3456"
$env:ANTHROPIC_AUTH_TOKEN = "test"
```

Windows (Command Prompt):
```cmd
set ANTHROPIC_BASE_URL=http://localhost:3456
set ANTHROPIC_AUTH_TOKEN=test
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
      "models": ["claude-opus-4-20250720", "claude-sonnet-4-20250720", "claude-3-5-haiku-20241022"],
      "enabled": true
    },
    {
      "name": "openai",
      "api_key": "sk-...",
      "models": ["gpt-4.1", "gpt-4.1-mini", "gpt-4o", "o3", "o4-mini"],
      "enabled": true
    },
    {
      "name": "gemini",
      "api_key": "AI...",
      "models": ["gemini-2.5-pro", "gemini-2.5-flash"],
      "enabled": true
    },
    {
      "name": "deepseek",
      "api_key": "sk-...",
      "models": ["deepseek-chat", "deepseek-reasoner"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "gpt-4.1"
    },
    "longContext": {
      "provider": "anthropic",
      "model": "claude-opus-4-20250720"
    },
    "think": {
      "provider": "openai",
      "model": "o3"
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

**macOS/Linux:**
```bash
# Option 1: Reload your shell configuration
source ~/.bashrc    # For bash
source ~/.zshrc     # For zsh

# Option 2: Use the full path
/usr/local/bin/ccproxy start

# Option 3: Check if it's installed
ls -la /usr/local/bin/ccproxy
```

**Windows:**
```powershell
# Option 1: Restart your terminal (PATH changes need this)

# Option 2: Use the full path
"C:\Program Files\CCProxy\ccproxy.exe" start

# Option 3: Check your PATH
echo $env:PATH

# Option 4: Verify it's installed
Test-Path "C:\Program Files\CCProxy\ccproxy.exe"
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

**macOS/Linux:**
```bash
# Check if config exists
ls -la ~/.ccproxy/config.json

# If missing, create it:
mkdir -p ~/.ccproxy
# Then copy the example config from Step 2 above
```

**Windows:**
```powershell
# Check if config exists
Test-Path "$env:USERPROFILE\.ccproxy\config.json"

# If missing, run the PowerShell commands from Step 1 to create it
```

**Need help?** [💬 GitHub Discussions](https://github.com/orchestre-dev/ccproxy/discussions) • [🐛 Report Issues](https://github.com/orchestre-dev/ccproxy/issues)