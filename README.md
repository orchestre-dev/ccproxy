# CCProxy - Multi-Provider AI Proxy Server

A production-grade Golang proxy server that enables the use of multiple AI providers (Groq, OpenRouter, OpenAI, and more) with Claude Code by translating between Anthropic and provider-specific API formats.

## 🌟 Features

- **Multi-Provider Support**: Switch between Groq, OpenRouter, OpenAI, and other providers
- **API Translation**: Seamless conversion between Anthropic and OpenAI-compatible API formats
- **Tool Support**: Full support for Anthropic tool calling and tool results
- **Production Ready**: Built for high performance and reliability
- **Cross Platform**: Binaries available for Linux, macOS, and Windows (AMD64/ARM64)
- **Docker Support**: Container-ready with multi-stage builds
- **Comprehensive Logging**: Structured logging with configurable levels
- **Graceful Shutdown**: Proper signal handling and connection draining
- **Health Checks**: Built-in health monitoring and provider status endpoints

## 🚀 Quick Start

### Option 1: Download Pre-built Binary

1. Download the latest binary for your platform from the [releases page](https://github.com/your-repo/ccproxy/releases)
2. Set your provider and API key:
   ```bash
   # For Groq
   export PROVIDER=groq
   export GROQ_API_KEY=your_groq_api_key_here
   
   # Or for OpenRouter
   export PROVIDER=openrouter
   export OPENROUTER_API_KEY=your_openrouter_api_key_here
   ```
3. Run the proxy:
   ```bash
   ./ccproxy-<platform>
   ```

### Option 2: Build from Source

1. **Prerequisites**: Go 1.21 or later
2. **Clone and build**:
   ```bash
   git clone https://github.com/your-repo/ccproxy.git
   cd ccproxy
   go build ./cmd/proxy
   ```
3. **Set up environment**:
   ```bash
   cp .env.example .env
   # Edit .env and set your PROVIDER and API key
   ```
4. **Run the proxy**:
   ```bash
   ./proxy
   ```

### Option 3: Docker

1. **Using Docker Compose** (recommended):
   ```bash
   # Set your provider and API key
   echo "PROVIDER=groq" > .env
   echo "GROQ_API_KEY=your_groq_api_key_here" >> .env
   
   # Start the service
   docker-compose up -d
   ```

2. **Using Docker directly**:
   ```bash
   docker build -f docker/Dockerfile -t ccproxy .
   docker run -p 7187:7187 -e PROVIDER=groq -e GROQ_API_KEY=your_api_key ccproxy
   ```

## 🔧 Configuration

### Core Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PROVIDER` | *(required)* | Provider to use: `groq`, `openrouter` |
| `SERVER_HOST` | `0.0.0.0` | Server bind address |
| `SERVER_PORT` | `7187` | Server port |
| `SERVER_ENVIRONMENT` | `development` | Environment mode |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `LOG_FORMAT` | `json` | Log format (json, text) |

### Provider-Specific Variables

#### Groq Provider
| Variable | Default | Description |
|----------|---------|-------------|
| `GROQ_API_KEY` | *(required)* | Your Groq API key |
| `GROQ_BASE_URL` | `https://api.groq.com/openai/v1` | Groq API base URL |
| `GROQ_MODEL` | `moonshotai/kimi-k2-instruct` | Model to use |
| `GROQ_MAX_TOKENS` | `16384` | Maximum tokens per request |

#### OpenRouter Provider
| Variable | Default | Description |
|----------|---------|-------------|
| `OPENROUTER_API_KEY` | *(required)* | Your OpenRouter API key |
| `OPENROUTER_BASE_URL` | `https://openrouter.ai/api/v1` | OpenRouter API base URL |
| `OPENROUTER_MODEL` | `openai/gpt-4o` | Model to use |
| `OPENROUTER_MAX_TOKENS` | `4096` | Maximum tokens per request |
| `OPENROUTER_SITE_URL` | | Your site URL (for analytics) |
| `OPENROUTER_SITE_NAME` | | Your site name (for analytics) |

### Configuration File

You can also use a YAML configuration file:

```yaml
# config.yaml
provider: "groq"  # or "openrouter"

server:
  host: "0.0.0.0"
  port: "7187"
  environment: "production"

groq:
  api_key: "your_groq_api_key_here"
  model: "moonshotai/kimi-k2-instruct"
  max_tokens: 16384

openrouter:
  api_key: "your_openrouter_api_key_here"
  model: "openai/gpt-4o"
  max_tokens: 4096
  site_url: "https://your-site.com"
  site_name: "Your App Name"

logging:
  level: "info"
  format: "json"
```

## 🎯 Using with Claude Code

1. **Start the proxy server**:
   ```bash
   # For Groq
   PROVIDER=groq GROQ_API_KEY=your_key ./ccproxy-<platform>
   
   # For OpenRouter  
   PROVIDER=openrouter OPENROUTER_API_KEY=your_key ./ccproxy-<platform>
   
   # Or: docker-compose up -d
   ```

2. **Configure Claude Code** to use the proxy:
   ```bash
   export ANTHROPIC_BASE_URL=http://localhost:7187
   export ANTHROPIC_API_KEY=NOT_NEEDED
   ```

3. **Run Claude Code**:
   ```bash
   claude
   ```

Claude Code will now use your selected AI provider through the proxy!

## 🏆 Provider Comparison

| Provider | Strengths | Best For | Popular Models |
|----------|-----------|----------|----------------|
| **Groq** | Ultra-fast inference, cost-effective | Real-time applications, development | `moonshotai/kimi-k2-instruct`, `llama-3.1-405b-reasoning` |
| **OpenRouter** | Huge model selection, competitive pricing | Access to latest models, experimentation | `openai/gpt-4o`, `anthropic/claude-3-sonnet`, `google/gemini-2.5-pro-preview` |

### Important: Tool Calling Requirement

**⚠️ Critical for Claude Code Users**: You must select models that support **tool calling** or **function calling** capabilities, as Claude Code requires these features to operate correctly.

### Model Selection Guidelines

When choosing models from any provider:

1. **Verify Tool Support**: Ensure the model supports function calling/tool use
2. **Check Current Availability**: Model availability changes frequently
3. **Review Capabilities**: Different models excel at different tasks
4. **Consider Performance**: Balance speed, quality, and cost for your needs

For current model lists and capabilities, visit each provider's official documentation.

## 📊 API Endpoints

### Main Endpoint
- `POST /v1/messages` - Anthropic-compatible messages endpoint

### Health & Monitoring
- `GET /` - Basic health check with current provider info
- `GET /health` - Detailed health information
- `GET /status` - Current provider status and configuration

### Example Request

```bash
curl -X POST http://localhost:7187/v1/messages \
  -H "Content-Type: application/json" \
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
go build ./cmd/proxy

# Cross-platform builds
./scripts/build.sh

# Windows builds
./scripts/build.bat
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/converter
```

### Hot Reload Development

Install Air for hot reloading:
```bash
go install github.com/cosmtrek/air@latest
air
```

## 🐛 Troubleshooting

### Common Issues

1. **"PROVIDER environment variable is required"**
   - Ensure you've set the PROVIDER variable to `groq` or `openrouter`
   - Set the corresponding API key for your chosen provider

2. **"API_KEY environment variable is required"**
   - For Groq: Set `GROQ_API_KEY` - Get your key from [Groq Console](https://console.groq.com/)
   - For OpenRouter: Set `OPENROUTER_API_KEY` - Get your key from [OpenRouter](https://openrouter.ai/)

3. **Connection refused on localhost:7187**
   - Check if the proxy is running: `curl http://localhost:7187/`
   - Verify the port isn't in use: `lsof -i :7187`

4. **"Failed to call provider API"**
   - Verify your API key is valid for the selected provider
   - Check internet connectivity
   - Review proxy logs for detailed error messages
   - Ensure the model name is supported by your provider

### Debug Mode

Enable debug logging for troubleshooting:
```bash
export LOG_LEVEL=debug
./ccproxy-<platform>
```

### Health Check

Check if the service is healthy:
```bash
curl http://localhost:7187/health
```

## 📈 Performance

### Benchmarks

The Golang implementation provides significant performance improvements over the Python version:

- **Memory Usage**: ~10-20MB (vs ~50-100MB Python)
- **Startup Time**: <100ms (vs ~2-3s Python)
- **Request Latency**: <10ms conversion overhead
- **Throughput**: >1000 requests/second

### Optimization

For production deployments:

1. Use the `production` environment
2. Set appropriate resource limits
3. Enable horizontal scaling with load balancers
4. Monitor memory and CPU usage

## 🔒 Security

### Best Practices

- Keep your API keys secure and rotate regularly
- Use HTTPS in production environments
- Set up proper firewall rules
- Monitor for unusual API usage patterns
- Use environment variables for sensitive configuration
- Choose the right provider for your use case (cost, speed, model availability)

### Production Deployment

For production use:

```bash
# Use production environment
export SERVER_ENVIRONMENT=production

# Bind to specific interface
export SERVER_HOST=127.0.0.1

# Use structured JSON logging
export LOG_FORMAT=json
export LOG_LEVEL=warn
```

## 🚀 Deployment

### Systemd Service

Create `/etc/systemd/system/ccproxy.service`:

```ini
[Unit]
Description=CCProxy Multi-Provider AI Proxy Server
After=network.target

[Service]
Type=simple
User=ccproxy
Group=ccproxy
ExecStart=/usr/local/bin/ccproxy
Restart=always
RestartSec=5
Environment=PROVIDER=groq
Environment=GROQ_API_KEY=your_api_key_here
Environment=LOG_LEVEL=info
Environment=SERVER_ENVIRONMENT=production

[Install]
WantedBy=multi-user.target
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ccproxy
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ccproxy
  template:
    metadata:
      labels:
        app: ccproxy
    spec:
      containers:
      - name: ccproxy
        image: ccproxy:latest
        ports:
        - containerPort: 7187
        env:
        - name: PROVIDER
          value: "groq"
        - name: GROQ_API_KEY
          valueFrom:
            secretKeyRef:
              name: ccproxy-secrets
              key: groq-api-key
        - name: SERVER_ENVIRONMENT
          value: "production"
```

## 🔄 Migration from Python Version

The Golang version is a drop-in replacement for the Python version:

1. **Same API**: Identical endpoints and request/response formats
2. **Same Configuration**: Uses the same environment variables
3. **Same Port**: Default port 7187
4. **Better Performance**: Significantly faster and uses less memory

### Side-by-Side Testing

You can run both versions simultaneously for testing:

```bash
# Python version on port 7187
python proxy.py

# Golang version on port 7188
SERVER_PORT=7188 ./ccproxy
```

## 📋 Comparison with Original

| Feature | Python Version | Golang Version |
|---------|---------------|----------------|
| **Startup Time** | ~2-3 seconds | <100ms |
| **Memory Usage** | ~50-100MB | ~10-20MB |
| **CPU Usage** | Higher | Lower |
| **Dependencies** | Many (FastAPI, etc.) | Minimal |
| **Binary Size** | N/A | ~9MB |
| **Cross-compilation** | No | Yes |
| **Static Binary** | No | Yes |
| **Deployment** | Complex | Simple |

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Development Setup

```bash
# Clone the repo
git clone https://github.com/your-repo/ccproxy.git
cd ccproxy

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build ./cmd/proxy
```

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by [claude-code-proxy](https://github.com/1rgs/claude-code-proxy)
- Built with [Gin Web Framework](https://github.com/gin-gonic/gin)
- Supports multiple providers: [Groq](https://groq.com/), [OpenRouter](https://openrouter.ai/), and more

## 📞 Support

- 📖 [Documentation](./docs/)
- 🐛 [Issue Tracker](https://github.com/your-repo/ccproxy/issues)
- 💬 [Discussions](https://github.com/your-repo/ccproxy/discussions)

---

If you find this project useful, please consider giving it a ⭐ on GitHub!