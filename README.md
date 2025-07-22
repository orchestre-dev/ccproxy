# CCProxy

CCProxy is a high-performance Go proxy server that enables Claude Code to work with multiple AI providers through intelligent routing and API translation.

## 🌟 Features

- **Multi-Provider Support**: Anthropic, OpenAI, Google Gemini, Mistral, DeepSeek, Groq, OpenRouter, Ollama, and more
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

CCProxy uses a JSON configuration file. Create `config.json` based on `example.config.json`:

```json
{
  "host": "127.0.0.1",
  "port": 3456,
  "log": true,
  "apikey": "your-api-key-here",
  "providers": [
    {
      "name": "anthropic",
      "api_base_url": "https://api.anthropic.com",
      "api_key": "your-anthropic-api-key",
      "models": ["claude-3-opus-20240229", "claude-3-sonnet-20240229"],
      "enabled": true
    },
    {
      "name": "openai",
      "api_base_url": "https://api.openai.com/v1",
      "api_key": "your-openai-api-key",
      "models": ["gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"],
      "enabled": false
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

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CCPROXY_PORT` | `3456` | Override default port |
| `CCPROXY_HOST` | `127.0.0.1` | Override default host |
| `CCPROXY_API_KEY` | | Set API key for authentication |
| `CCPROXY_CONFIG` | | Path to configuration file |
| `LOG` | `false` | Enable file logging |

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
- **Mistral AI** - European privacy-focused models
- **DeepSeek** - Cost-effective coding models
- **Groq** - Ultra-fast inference with LPU technology
- **OpenRouter** - Access to 100+ models from various providers
- **XAI** - Grok models with real-time data
- **Ollama** - Local models for complete privacy

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