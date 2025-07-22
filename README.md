# CCProxy - AI Request Proxy for Claude Code | Multi-Provider LLM Gateway

[![CI](https://github.com/orchestre-dev/ccproxy/actions/workflows/ci.yml/badge.svg)](https://github.com/orchestre-dev/ccproxy/actions/workflows/ci.yml)
[![Pre-Release](https://github.com/orchestre-dev/ccproxy/actions/workflows/pre-release.yml/badge.svg)](https://github.com/orchestre-dev/ccproxy/actions/workflows/pre-release.yml)
[![Release](https://github.com/orchestre-dev/ccproxy/actions/workflows/release.yml/badge.svg)](https://github.com/orchestre-dev/ccproxy/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/orchestre-dev/ccproxy)](https://goreportcard.com/report/github.com/orchestre-dev/ccproxy)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Documentation](https://img.shields.io/badge/docs-ccproxy.dev-blue)](https://ccproxy.pages.dev)

**CCProxy is the premier AI request proxy for Claude Code**, providing seamless integration with multiple LLM providers including OpenAI, Google Gemini, DeepSeek, and more. This high-performance proxy server enables Claude Code to access various AI models through intelligent routing and automatic API translation, making it the essential tool for developers using Claude Code who need multi-provider flexibility.

📚 **[Full Documentation](https://ccproxy.pages.dev)** | 🐛 **[Report Issues](https://github.com/orchestre-dev/ccproxy/issues)** | 💬 **[Discussions](https://github.com/orchestre-dev/ccproxy/discussions)**

> **Keywords**: AI proxy for Claude Code, Claude Code proxy server, LLM gateway, AI model router, Anthropic Claude proxy, OpenAI GPT proxy, multi-provider AI gateway, Claude Code integration

## 🎯 Why CCProxy for Claude Code?

CCProxy is specifically designed as an **AI request proxy for Claude Code**, solving key challenges:

- **Multi-Provider Access**: Use Claude Code with OpenAI GPT-4, Google Gemini, DeepSeek, and more
- **Cost Optimization**: Route requests to the most cost-effective provider based on your needs
- **Failover Protection**: Automatic fallback to alternative providers if one is unavailable
- **Token-Based Routing**: Intelligently route long-context requests to appropriate models
- **Drop-in Replacement**: Works seamlessly with Claude Code without any code changes

## 🌟 Features for Claude Code Users

- **Multi-Provider Support**: Anthropic, OpenAI, Google Gemini, DeepSeek, OpenRouter
- **Intelligent Routing**: Automatic model selection based on token count and parameters
- **API Translation**: Seamless conversion between Anthropic and provider-specific formats
- **Tool Support**: Full support for function calling across all compatible providers
- **Streaming Support**: Server-Sent Events (SSE) for real-time responses
- **Cross Platform**: Binaries available for Linux, macOS, and Windows (AMD64/ARM64)
- **Process Management**: Background service with automatic startup and graceful shutdown
- **Health Monitoring**: Built-in health checks and provider status tracking
- **Security**: API key validation, IP-based access control, rate limiting

## 🚀 Quick Start

### Option 1: Download Pre-built Binary

1. Download the latest binary for your platform from the [releases page](https://github.com/orchestre-dev/ccproxy/releases)
2. Create a configuration file:
   ```bash
   cp example.config.json config.json
   # Edit config.json to add your provider API keys
   ```
3. Start CCProxy:
   ```bash
   ./ccproxy start
   ```

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

### Option 4: One-Command Setup with Claude Code (Recommended)

```bash
# The fastest way to start using CCProxy with Claude Code
# This command auto-configures everything for Claude Code integration
./ccproxy code
```

This is the **recommended method for Claude Code users** as it automatically:
- Starts the CCProxy AI proxy server
- Configures Claude Code environment variables
- Sets up the proxy connection for Claude Code
- Enables access to all configured AI providers

## 🔧 Configuration Guide for Claude Code Integration

### Quick Start Configuration for Claude Code

The easiest way to configure CCProxy as an AI proxy for Claude Code is to create a configuration file with your AI provider API keys:

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
         "models": ["gpt-4", "gpt-3.5-turbo"],
         "enabled": true
       }
     ]
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

CCProxy, as an AI proxy for Claude Code, uses two types of API keys for different purposes:

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

### 📋 Complete Configuration Example for Claude Code

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
      "models": ["claude-3-opus-20240229", "claude-3-sonnet-20240229"],
      "enabled": true
    },
    {
      "name": "openai",
      "api_base_url": "https://api.openai.com/v1",
      "api_key": "sk-...",  // Required: your OpenAI API key
      "models": ["gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"],
      "enabled": true
    },
    {
      "name": "gemini",
      "api_base_url": "https://generativelanguage.googleapis.com/v1",
      "api_key": "AIza...",  // Required: your Google AI API key
      "models": ["gemini-pro", "gemini-pro-vision"],
      "enabled": false
    }
  ],
  "routes": {
    "default": {
      "provider": "anthropic",
      "model": "claude-3-sonnet-20240229"
    },
    "longContext": {
      "provider": "anthropic",
      "model": "claude-3-opus-20240229",
      "conditions": [{
        "type": "tokenCount",
        "operator": ">",
        "value": 60000
      }]
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

#### Example 1: Simple Claude Code Setup (Local Development)
```json
{
  "providers": [
    {
      "name": "openai",
      "api_key": "sk-proj-...",
      "enabled": true
    }
  ]
}
```

#### Example 2: Multi-Provider Setup for Claude Code
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
    }
  ],
  "routes": {
    "default": {
      "provider": "deepseek",
      "model": "deepseek-coder"
    },
    "longContext": {
      "provider": "anthropic",
      "model": "claude-3-opus-20240229"
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

## 🎯 Using CCProxy as an AI Proxy for Claude Code

### Automatic Setup for Claude Code (Recommended)

```bash
# One command to set up CCProxy as your AI proxy for Claude Code
./ccproxy code
```

✨ **What happens**: CCProxy automatically configures itself as the AI proxy for Claude Code, enabling access to all your configured providers.

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

## 🏆 Supported AI Providers for Claude Code Integration

**CCProxy acts as an AI request proxy for Claude Code**, enabling seamless integration with multiple language model providers:

- **Anthropic** - Claude models with native support
- **OpenAI** - GPT-4, GPT-3.5 models
- **Google Gemini** - Advanced multimodal models
- **DeepSeek** - Cost-effective coding models
- **OpenRouter** - Access to 100+ models from various providers

### Provider Support Levels

CCProxy offers different levels of support for various providers:

#### Full Support (with dedicated transformers)
These providers have complete API translation and all features work seamlessly:
- **Anthropic** - Native Claude API support
- **OpenAI** - Complete GPT model compatibility
- **Google Gemini** - Full multimodal support
- **DeepSeek** - Optimized for coding tasks
- **OpenRouter** - Unified access to multiple providers

#### Basic Routing Support
These providers have basic routing capabilities but may have limited functionality:
- Groq, Mistral, XAI, Ollama - Basic message routing only

For production use, we recommend using providers with full support.

### Configuration Example

Add providers to your `config.json`:

```json
{
  "providers": [
    {
      "name": "openai",
      "api_base_url": "https://api.openai.com/v1",
      "api_key": "your-api-key",
      "models": ["gpt-4", "gpt-3.5-turbo"],
      "enabled": true
    }
  ]
}
```

## 📊 API Endpoints for Claude Code Proxy

- `POST /v1/messages` - Anthropic-compatible messages endpoint
- `GET /health` - Health check endpoint
- `GET /status` - Service status information
- `GET /providers` - List configured providers
- `POST /providers` - Add new provider
- `PUT /providers/:name` - Update provider
- `DELETE /providers/:name` - Remove provider

### Example Request

```bash
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -d '{
    "model": "claude-3-sonnet-20240229",
    "max_tokens": 1000,
    "messages": [
      {
        "role": "user",
        "content": "Hello, how are you?"
      }
    ]
  }'
```

## 🔨 Development Guide for CCProxy AI Proxy

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

Inspired by the original [Claude Code Router](https://github.com/musistudio/claude-code-router) project.

## 🐛 Troubleshooting Claude Code AI Proxy Issues

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

**CCProxy is the essential AI request proxy for Claude Code**, providing seamless multi-provider integration for AI-powered development. Whether you're using OpenAI's GPT-4, Google's Gemini, Anthropic's Claude, or other LLM providers, CCProxy makes it easy to switch between them while using Claude Code.

### Key Benefits of Using CCProxy with Claude Code:
- ✅ **Multi-Provider Support**: Access OpenAI, Google Gemini, DeepSeek, and more through Claude Code
- ✅ **Cost Optimization**: Route requests to the most cost-effective AI provider
- ✅ **High Performance**: Written in Go for minimal latency and resource usage
- ✅ **Easy Setup**: One command (`ccproxy code`) configures everything
- ✅ **Enterprise Ready**: Built-in security, monitoring, and rate limiting

Start using CCProxy as your AI proxy for Claude Code today and unlock the full potential of multi-provider AI development!

---

**Keywords**: CCProxy, AI proxy for Claude Code, Claude Code proxy server, LLM gateway, multi-provider AI proxy, Anthropic Claude proxy, OpenAI GPT proxy, Google Gemini proxy, AI model router, Claude Code integration, AI request routing, LLM proxy server

If you find this project useful, please consider giving it a ⭐ on GitHub!