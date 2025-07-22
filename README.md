# CCProxy

[![CI](https://github.com/orchestre-dev/ccproxy/actions/workflows/ci.yml/badge.svg)](https://github.com/orchestre-dev/ccproxy/actions/workflows/ci.yml)
[![Pre-Release](https://github.com/orchestre-dev/ccproxy/actions/workflows/pre-release.yml/badge.svg)](https://github.com/orchestre-dev/ccproxy/actions/workflows/pre-release.yml)
[![Release](https://github.com/orchestre-dev/ccproxy/actions/workflows/release.yml/badge.svg)](https://github.com/orchestre-dev/ccproxy/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/orchestre-dev/ccproxy)](https://goreportcard.com/report/github.com/orchestre-dev/ccproxy)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Documentation](https://img.shields.io/badge/docs-ccproxy.orchestre.dev-blue)](https://ccproxy.orchestre.dev)

A high-performance proxy that enables Claude Code to work with multiple AI providers. Route requests to OpenAI, Google Gemini, DeepSeek, and more through a single interface.

📚 **[Documentation](https://ccproxy.orchestre.dev)** | 🐛 **[Issues](https://github.com/orchestre-dev/ccproxy/issues)** | 💬 **[Discussions](https://github.com/orchestre-dev/ccproxy/discussions)**

## 🎯 Why CCProxy?

CCProxy enables Claude Code to work with multiple AI providers:

- **Top-Ranked Model Access**: Use Qwen3 235B (#1 on AIME with 70.3 score) via OpenRouter
- **100+ Models Through OpenRouter**: Access Qwen3, Kimi K2, Grok, and many more
- **Cost Optimization**: Route requests to the most cost-effective provider
- **Failover Protection**: Automatic fallback when providers are unavailable
- **Token-Based Routing**: Intelligent routing based on context length
- **Drop-in Replacement**: Works seamlessly without code changes

## 🌟 Features

- **Multi-Provider Support**: Anthropic, OpenAI, Google Gemini, DeepSeek, OpenRouter (100+ models)
- **Intelligent Routing**: Automatic model selection based on context
- **API Translation**: Seamless format conversion between providers
- **Tool Support**: Function calling required for Claude Code compatibility
- **Streaming Support**: Real-time responses via SSE
- **Cross Platform**: Linux, macOS, and Windows (AMD64/ARM64)
- **Process Management**: Background service with auto-startup
- **Health Monitoring**: Built-in status tracking
- **Security**: API validation, access control, rate limiting

## 🎯 Model Selection Strategy

CCProxy intelligently routes requests based on:
- Token count (>60K → longContext route)
- Model type (haiku models → background route)
- Thinking parameter (thinking: true → think route if configured)
- Explicit selection (provider,model format)
- Access to top models like Qwen3 235B via OpenRouter

**Note**: For Claude Code compatibility, models must support function calling (tool use).

## 🚀 Quick Start

### Automated Installation (Recommended)

**macOS/Linux:**
```bash
curl -sSL https://raw.githubusercontent.com/orchestre-dev/ccproxy/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/orchestre-dev/ccproxy/main/install.ps1 | iex
```

Both installers will:
- Download and install CCProxy
- Create configuration directory with starter config
- Update your PATH
- Show you exactly what to do next

### Manual Installation

Download the latest binary for your platform from the [releases page](https://github.com/orchestre-dev/ccproxy/releases):
- **macOS**: `ccproxy-darwin-amd64` (Intel) or `ccproxy-darwin-arm64` (Apple Silicon)
- **Linux**: `ccproxy-linux-amd64` or `ccproxy-linux-arm64`
- **Windows**: `ccproxy-windows-amd64.exe`

### Option 2: Build from Source

1. **Prerequisites**: Go 1.23 or later
2. **Clone and build**:
   ```bash
   git clone https://github.com/orchestre-dev/ccproxy.git
   cd ccproxy
   go build ./cmd/ccproxy
   ```
3. **Configure**:
   ```bash
   cp example.config.json config.json
   # Edit config.json to add your provider API keys
   ```
4. **Start the service**:
   ```bash
   ./ccproxy start
   ```

### Option 3: Docker

```bash
docker build -t ccproxy .
docker run -d -p 3456:3456 -v $(pwd)/config.json:/home/ccproxy/.ccproxy/config.json ccproxy
```

### Configure and Start

After installation, CCProxy needs your API keys:

**Config file location:**
- macOS/Linux: `~/.ccproxy/config.json`
- Windows: `%USERPROFILE%\.ccproxy\config.json`

Edit the config to add your API key(s), then:

```bash
# Start the proxy server
ccproxy start

# Configure Claude Code (sets environment variables)
ccproxy code
```

## 🔧 Configuration Guide

### Quick Start Configuration

Create a configuration file with your provider API keys:

1. **Create configuration directory** (if it doesn't exist):
   ```bash
   mkdir -p ~/.ccproxy
   ```

2. **Create config file** with your provider API keys:
   ```bash
   cat > ~/.ccproxy/config.json << 'EOF'
   {
     "providers": [
       {
         "name": "openai",
         "api_base_url": "https://api.openai.com/v1",
         "api_key": "your-openai-api-key",
         "models": ["gpt-4.1", "gpt-4.1-mini"],  // Available models for validation
         "enabled": true
       }
     ],
     "routes": {
       "default": {
         "provider": "openai",
         "model": "gpt-4.1"  // This is the actual model that will be used
       }
     }
   }
   EOF
   ```

3. **Start CCProxy and configure Claude Code**:
   ```bash
   ./ccproxy code
   ```

This single command will:
- Start the CCProxy service
- Set up environment variables for Claude Code
- Enable Claude Code to use your configured AI providers

### Configuration Priority System

CCProxy uses a layered configuration system. Settings are applied in this order (highest priority first):

1. **Command-line flags** - Override any other setting
   ```bash
   ccproxy start --port 8080 --config /custom/config.json
   ```

2. **Environment variables** - Override config file values
   ```bash
   export CCPROXY_PORT=8080
   export CCPROXY_API_KEY=my-secret-key
   ```

3. **Configuration file** - Your main configuration
   ```json
   { "port": 3456, "apikey": "configured-key" }
   ```

4. **Built-in defaults** - Used when nothing else is specified

### Configuration File Locations

CCProxy searches for `config.json` in these locations (in order):
1. Current directory: `./config.json`
2. User home directory: `~/.ccproxy/config.json` 
3. System directory: `/etc/ccproxy/config.json`

The first found file is used. You can also specify a custom path:
```bash
ccproxy start --config /path/to/my/config.json
```

### 🔑 Understanding the Two-Level API Key System

CCProxy uses two types of API keys:

#### 1. **CCProxy Authentication Key** (Optional)

<table>
<tr><td><b>What it does</b></td><td>Secures access to your CCProxy server</td></tr>
<tr><td><b>Who uses it</b></td><td>Claude Code when connecting to CCProxy</td></tr>
<tr><td><b>Configuration</b></td><td><code>"apikey": "your-secret-key"</code></td></tr>
<tr><td><b>When to use</b></td><td>When CCProxy is accessible from non-localhost addresses</td></tr>
<tr><td><b>Default behavior</b></td><td>If not set, only localhost connections are allowed</td></tr>
</table>

**Example scenario**: You're running CCProxy on a server and Claude Code connects from your laptop:
```json
{
  "host": "0.0.0.0",  // Accessible from network
  "apikey": "my-secret-ccproxy-key"  // Required for security
}
```

#### 2. **AI Provider API Keys** (Required)

<table>
<tr><td><b>What they do</b></td><td>Authenticate CCProxy to AI services (OpenAI, Anthropic, etc.)</td></tr>
<tr><td><b>Who uses them</b></td><td>CCProxy when forwarding your requests to AI providers</td></tr>
<tr><td><b>Configuration</b></td><td>Each provider has its own <code>"api_key"</code></td></tr>
<tr><td><b>When to use</b></td><td>Always - you need valid keys from each AI provider you want to use</td></tr>
<tr><td><b>How to get them</b></td><td>Sign up at each provider's website (OpenAI, Anthropic, Google AI, etc.)</td></tr>
</table>

**Visual Flow**:
```
Claude Code → (uses CCProxy API key) → CCProxy → (uses Provider API key) → OpenAI/Anthropic/etc
```

### 📋 Complete Configuration Example

```json
{
  "host": "127.0.0.1",
  "port": 3456,
  "log": true,
  "log_file": "~/.ccproxy/ccproxy.log",
  "apikey": "my-secret-ccproxy-key",  // Optional: for CCProxy authentication
  "providers": [
    {
      "name": "anthropic",
      "api_base_url": "https://api.anthropic.com",
      "api_key": "sk-ant-...",  // Required: your Anthropic API key
      "models": ["claude-opus-4-20250720", "claude-sonnet-4-20250720"],
      "enabled": true
    },
    {
      "name": "openai",
      "api_base_url": "https://api.openai.com/v1",
      "api_key": "sk-...",  // Required: your OpenAI API key
      "models": ["gpt-4.1", "gpt-4.1-turbo", "gpt-4.1-mini"],
      "enabled": true
    },
    {
      "name": "gemini",
      "api_base_url": "https://generativelanguage.googleapis.com/v1",
      "api_key": "AIza...",  // Required: your Google AI API key
      "models": ["gemini-2.5-flash", "gemini-2.5-pro"],
      "enabled": false
    },
    {
      "name": "deepseek",
      "api_base_url": "https://api.deepseek.com",
      "api_key": "sk-...",  // Required: your DeepSeek API key
      "models": ["deepseek-chat", "deepseek-coder"],
      "enabled": true
    },
    {
      "name": "openrouter",
      "api_base_url": "https://openrouter.ai/api/v1",
      "api_key": "sk-or-...",  // Required: your OpenRouter API key
      "models": ["qwen/qwen-3-235b", "moonshotai/kimi-k2-instruct", "xai/grok-beta"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "anthropic",
      "model": "claude-sonnet-4-20250720"
    },
    "longContext": {
      "provider": "anthropic",
      "model": "claude-opus-4-20250720"
    },
    "background": {
      "provider": "openai",
      "model": "gpt-4.1-mini"
    },
    "think": {  // Optional: route for thinking parameter
      "provider": "anthropic",
      "model": "claude-opus-4-20250720"
    },
    "gpt-4.1": {
      "provider": "openai",
      "model": "gpt-4.1-turbo"
    },
    "qwen3-235b": {  // Top-ranked model via OpenRouter
      "provider": "openrouter",
      "model": "qwen/qwen-3-235b"
    }
  },
  "performance": {
    "request_timeout": "30s",
    "max_request_body_size": 10485760,
    "metrics_enabled": true,
    "rate_limit_enabled": false,
    "circuit_breaker_enabled": true
  }
}
```

### 🎯 How Model Selection Works

CCProxy uses intelligent routing to select the appropriate model and provider:

**1. Explicit Provider Selection** (Highest Priority)
```json
// Force a specific provider/model combination
{"model": "openai,gpt-4.1-turbo"}
```

**2. Direct Model Routes**
```json
// If "gpt-4.1" is defined in routes, use that configuration
{"model": "gpt-4.1"}
```

**3. Automatic Token-Based Routing**
```json
// Requests with >60K tokens automatically use longContext route
{"model": "claude-sonnet-4", "messages": [/* very long context */]}
```

**4. Background Task Routing**
```json
// Models starting with "claude-3-5-haiku" use background route
{"model": "claude-3-5-haiku-20241022"}
```

**5. Thinking Mode Routing**
```json
// Requests with thinking parameter use think route (if configured)
{"model": "claude-sonnet-4", "thinking": true}
```

**6. Default Route**
```json
// Any unmatched model uses the default route
{"model": "some-model"}
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CCPROXY_PORT` | `3456` | Port for CCProxy to listen on |
| `CCPROXY_HOST` | `127.0.0.1` | Host/IP for CCProxy to bind to |
| `CCPROXY_API_KEY` | | API key for CCProxy authentication |
| `CCPROXY_CONFIG` | | Path to configuration file |
| `CCPROXY_LOG` | `false` | Enable file logging |
| `CCPROXY_LOG_FILE` | `~/.ccproxy/ccproxy.log` | Log file path |
| `CCPROXY_PROVIDERS_0_API_KEY` | | Override first provider's API key |
| `CCPROXY_PROVIDERS_1_API_KEY` | | Override second provider's API key |
| `LOG` | `false` | Alternative way to enable logging |

### Real-World Configuration Examples

#### Example 1: Simple Setup (Local Development)
```json
{
  "providers": [
    {
      "name": "openai",
      "api_key": "sk-proj-...",
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "gpt-4.1-turbo"
    }
  }
}
```

#### Example 2: Multi-Provider Setup with Smart Routing
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
      "api_key": "sk-proj-...",
      "enabled": true
    },
    {
      "name": "deepseek",
      "api_key": "sk-...",
      "enabled": true
    },
    {
      "name": "openrouter",
      "api_base_url": "https://openrouter.ai/api/v1",
      "api_key": "sk-or-...",  // Required: your OpenRouter API key
      "models": ["qwen/qwen-3-235b", "moonshotai/kimi-k2-instruct", "xai/grok-beta"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "deepseek",
      "model": "deepseek-chat"
    },
    "longContext": {  // Auto-selected for >60K tokens
      "provider": "anthropic",
      "model": "claude-opus-4-20250720"
    },
    "gpt-4.1": {  // Direct model mapping
      "provider": "openai",
      "model": "gpt-4.1-turbo"
    },
    "claude-opus-4": {  // Another direct mapping
      "provider": "anthropic",
      "model": "claude-opus-4-20250720"
    },
    "qwen3-235b": {  // Access top-ranked Qwen3 235B
      "provider": "openrouter",
      "model": "qwen/qwen-3-235b"
    }
  }
}
```

#### Example 3: Priority Override Demonstration
```bash
# Config file sets: port 8080
# Environment variable: CCPROXY_PORT=9090
# Command flag: ccproxy start --port 7070
# Result: CCProxy uses port 7070 (command flag wins)
```

## 🎯 Using CCProxy

### Automatic Setup (Recommended)

```bash
# One command setup
./ccproxy code
```

✨ **What happens**: Automatic configuration enabling access to all configured providers.

### Manual Setup

1. **Start CCProxy**:
   ```bash
   ./ccproxy start
   ```

2. **Configure Claude Code**:
   ```bash
   export ANTHROPIC_BASE_URL=http://localhost:3456
   export ANTHROPIC_AUTH_TOKEN=test
   ```

3. **Use Claude Code**:
   ```bash
   claude "Help me with my code"
   ```

## 🏆 Supported AI Providers

CCProxy provides seamless integration with 5 major providers:

- **Anthropic** - Claude models with native support
- **OpenAI** - GPT-4.1, GPT-4.1-mini models
- **Google Gemini** - Advanced multimodal models
- **DeepSeek** - Cost-effective coding models
- **OpenRouter** - Gateway to 100+ models including:
  - 🏆 **Qwen3 235B** - Top-ranked model with 70.3 AIME score
  - ⚡ **Kimi K2** - Ultra-fast inference from Moonshot AI
  - 🌐 **Grok** - Real-time data access from xAI
  - And 100+ more models from various providers

**Note**: Additional providers like Groq, Mistral, XAI, and Ollama are accessible through OpenRouter, giving you the flexibility to use virtually any AI model through a single interface.

### Configuration Example

Add providers to your `config.json`:

```json
{
  "providers": [
    {
      "name": "openai",
      "api_base_url": "https://api.openai.com/v1",
      "api_key": "your-api-key",
      "models": ["gpt-4.1", "gpt-4.1-mini"],
      "enabled": true
    }
  ]
}
```

## 📊 API Endpoints

- `POST /v1/messages` - Anthropic-compatible messages endpoint
- `GET /health` - Health check endpoint
- `GET /status` - Service status information
- `GET /providers` - List configured providers
- `POST /providers` - Add new provider
- `PUT /providers/:name` - Update provider
- `DELETE /providers/:name` - Remove provider

### Example Requests

#### Basic Request (uses default route)
```bash
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -d '{
    "model": "claude-sonnet-4-20250720",
    "max_tokens": 1000,
    "messages": [
      {
        "role": "user",
        "content": "Hello, how are you?"
      }
    ]
  }'
```

#### Force Specific Provider
```bash
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -d '{
    "model": "openai,gpt-4.1-turbo",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

#### Long Context (auto-routes to longContext route)
```bash
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -d '{
    "model": "any-model",
    "messages": [{"role": "user", "content": "/* 100K+ token content */"}]
  }'
```

#### Thinking Mode (routes to think route if configured)
```bash
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -d '{
    "model": "claude-sonnet-4",
    "thinking": true,
    "messages": [{"role": "user", "content": "Solve this complex problem"}]
  }'
```

## 🔨 Development Guide

### Building

```bash
# Build for current platform
make build

# Cross-platform builds
make build-all

# Run tests
make test

# Run tests with race detection
make test-race
```

### Commands

- `ccproxy start` - Start the service
- `ccproxy stop` - Stop the service
- `ccproxy status` - Check service status
- `ccproxy code` - Configure Claude Code
- `ccproxy version` - Show version
- `ccproxy env` - Show environment variables

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guidelines](docs/guide/contributing.md) first.

## 🙏 Acknowledgments

Built with ❤️ for the Claude Code community.

Inspired by the original [Claude Code Router](https://github.com/musistudio/claude-router) project.

## 🐛 Troubleshooting

### Common Issues

1. **Service won't start**
   - Check if port 3456 is available: `lsof -i :3456`
   - Verify configuration file syntax
   - Check logs: `tail -f ~/.ccproxy/ccproxy.log`

2. **Authentication errors**
   - Verify API keys in config.json
   - Check provider is enabled in configuration
   - Ensure API key has correct permissions

3. **Connection refused**
   - Check service status: `./ccproxy status`
   - Verify firewall settings
   - Ensure service is bound to correct interface

### Debug Mode

Enable debug logging:
```bash
LOG=true ./ccproxy start
# Check logs at ~/.ccproxy/ccproxy.log
```

## 📞 Support

- 📖 [Documentation](https://ccproxy.orchestre.dev)
- 🐛 [Issue Tracker](https://github.com/orchestre-dev/ccproxy/issues)
- 💬 [Discussions](https://github.com/orchestre-dev/ccproxy/discussions)

## 🌟 Summary

CCProxy provides seamless multi-provider integration for AI-powered development. Access top-ranked models like Qwen3 235B, switch easily between OpenAI, Google Gemini, Anthropic Claude, DeepSeek, and leverage 100+ models through OpenRouter.

### Key Benefits:
- ✅ **Top-Ranked Models**: Access Qwen3 235B (70.3 AIME score) and 100+ models via OpenRouter
- ✅ **Multi-Provider Support**: 5 major AI providers with full implementation
- ✅ **Cost Optimization**: Route to cost-effective providers
- ✅ **High Performance**: Minimal latency with Go
- ✅ **Easy Setup**: One command configuration
- ✅ **Enterprise Ready**: Built-in security and monitoring

Start using CCProxy today to unlock multi-provider AI development!

---

If you find this project useful, please consider giving it a ⭐ on GitHub!