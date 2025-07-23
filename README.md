# CCProxy

[![CI](https://github.com/orchestre-dev/ccproxy/actions/workflows/ci.yml/badge.svg)](https://github.com/orchestre-dev/ccproxy/actions/workflows/ci.yml)
[![Release](https://github.com/orchestre-dev/ccproxy/actions/workflows/release.yml/badge.svg)](https://github.com/orchestre-dev/ccproxy/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/orchestre-dev/ccproxy)](https://goreportcard.com/report/github.com/orchestre-dev/ccproxy)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Documentation](https://img.shields.io/badge/docs-ccproxy.orchestre.dev-blue)](https://ccproxy.orchestre.dev)

A proxy that enables Claude Code to work with multiple AI providers (OpenAI, Google Gemini, DeepSeek, OpenRouter).

📚 **[Documentation](https://ccproxy.orchestre.dev)** | 🐛 **[Issues](https://github.com/orchestre-dev/ccproxy/issues)** | 💬 **[Discussions](https://github.com/orchestre-dev/ccproxy/discussions)**

## Installation

### Quick Install

**macOS/Linux:**
```bash
curl -sSL https://raw.githubusercontent.com/orchestre-dev/ccproxy/main/install.sh | bash
```

**Windows:**
```powershell
irm https://raw.githubusercontent.com/orchestre-dev/ccproxy/main/install.ps1 | iex
```

### Manual Install

Download from [releases](https://github.com/orchestre-dev/ccproxy/releases) or build from source:

```bash
git clone https://github.com/orchestre-dev/ccproxy.git
cd ccproxy
go build ./cmd/ccproxy
```

## Quick Start

1. **Configure** your API keys in `~/.ccproxy/config.json`:
```json
{
  "providers": [{
    "name": "openai",
    "api_key": "sk-...",
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

Or use one of our [example configurations](examples/configs/):
```bash
# Copy a ready-to-use configuration
cp examples/configs/openai-gpt4.json ~/.ccproxy/config.json

# Add your API key
export CCPROXY_PROVIDERS_0_API_KEY="sk-your-openai-key"
```

2. **Start** CCProxy and configure Claude Code:
```bash
ccproxy code
```

That's it. Claude Code now works with your configured providers.

## Configuration

See the [configuration guide](https://ccproxy.orchestre.dev/guide/configuration) for:
- Multi-provider setup
- Routing configuration
- Environment variables
- Advanced options

## Supported Providers

- **Anthropic** - Native Claude support
- **OpenAI** - GPT-4 models
- **Google Gemini** - Multimodal models
- **DeepSeek** - Cost-effective models (no tool support)
- **OpenRouter** - 100+ models gateway

Full details in the [provider documentation](https://ccproxy.orchestre.dev/providers/).

## Development

```bash
make build      # Build for current platform
make test       # Run tests
make build-all  # Build for all platforms
```

## Contributing

See [contributing guidelines](docs/guide/contributing.md).

## License

MIT License - see [LICENSE](LICENSE) file.

## Acknowledgments

Inspired by [Claude Code Router](https://github.com/musistudio/claude-router).