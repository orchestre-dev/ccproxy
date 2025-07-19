# CCProxy

CCProxy is a Go-based proxy server that acts as an intelligent intermediary between Claude Code and various Large Language Model (LLM) providers.

## Installation

```bash
# Build from source
make build

# Install to system
make install
```

## Usage

```bash
# Start the proxy server
ccproxy start

# Run Claude Code through CCProxy
ccproxy code

# Check server status
ccproxy status

# Stop the server
ccproxy stop
```

## Development

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Clean build artifacts
make clean
```

## License

MIT License - see LICENSE file for details