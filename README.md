# CCProxy - AI Request Proxy for Claude Code

[![CI](https://github.com/orchestre-dev/ccproxy/actions/workflows/ci.yml/badge.svg)](https://github.com/orchestre-dev/ccproxy/actions/workflows/ci.yml)
[![Pre-Release](https://github.com/orchestre-dev/ccproxy/actions/workflows/pre-release.yml/badge.svg)](https://github.com/orchestre-dev/ccproxy/actions/workflows/pre-release.yml)
[![Release](https://github.com/orchestre-dev/ccproxy/actions/workflows/release.yml/badge.svg)](https://github.com/orchestre-dev/ccproxy/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/orchestre-dev/ccproxy)](https://goreportcard.com/report/github.com/orchestre-dev/ccproxy)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Documentation](https://img.shields.io/badge/docs-ccproxy.dev-blue)](https://ccproxy.pages.dev)

CCProxy is a high-performance AI request proxy for Claude Code, enabling it to work with multiple AI providers through intelligent routing and API translation.

📚 **[Full Documentation](https://ccproxy.pages.dev)** | 🐛 **[Report Issues](https://github.com/orchestre-dev/ccproxy/issues)** | 💬 **[Discussions](https://github.com/orchestre-dev/ccproxy/discussions)**

## 🌟 Features

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

### Option 4: One-Command Setup with Claude Code

```bash
# CCProxy will auto-start and configure Claude Code
./ccproxy code
```

## 🔧 Configuration

CCProxy uses a layered configuration system with the following priority (highest to lowest):
1. **Command-line flags** (e.g., `--config /path/to/config.json`)
2. **Environment variables** (e.g., `CCPROXY_PORT=3456`)
3. **Configuration file** (`config.json`)
4. **Default values**

### Configuration File Locations

CCProxy searches for `config.json` in these locations (in order):
1. Current directory: `./config.json`
2. User home directory: `~/.ccproxy/config.json` 
3. System directory: `/etc/ccproxy/config.json`

The first found file is used. You can also specify a custom path:
```bash
ccproxy start --config /path/to/my/config.json
```

### Understanding API Keys

CCProxy uses API keys at two different levels, which can be confusing at first:

#### 1. **CCProxy API Key** (Root Level)
- **Purpose**: Authenticates requests TO CCProxy itself
- **Used by**: Claude Code or other clients connecting to CCProxy
- **Configuration**: `"apikey": "your-ccproxy-api-key"`
- **Security**: 
  - If not set, CCProxy only accepts requests from localhost (127.0.0.1)
  - If set, required for all requests (via `Authorization: Bearer` or `x-api-key` header)
- **Example**: When Claude Code connects to CCProxy, it uses this key

#### 2. **Provider API Keys** (Per Provider)
- **Purpose**: Authenticates CCProxy's requests TO each AI provider (Anthropic, OpenAI, etc.)
- **Used by**: CCProxy when forwarding requests to providers
- **Configuration**: Each provider has its own `"api_key"` field
- **Required**: Each enabled provider must have a valid API key from that provider
- **Example**: When CCProxy forwards a request to OpenAI, it uses the OpenAI API key

### Complete Configuration Example

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

### Configuration Priority Example

```bash
# 1. Default: port 3456
# 2. Config file sets: port 8080
# 3. Environment variable: CCPROXY_PORT=9090
# 4. Command flag: ccproxy start --port 7070
# Result: CCProxy uses port 7070 (command flag wins)
```

## 🎯 Using with Claude Code

### Automatic Setup (Recommended)

```bash
# CCProxy will auto-start and configure Claude Code environment
./ccproxy code
```

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

## 🏆 Supported Providers

CCProxy supports multiple AI providers with full API translation:

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

## 📊 API Endpoints

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

## 🔨 Development

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

---

If you find this project useful, please consider giving it a ⭐ on GitHub!