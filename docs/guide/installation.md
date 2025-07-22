---
title: Installation Guide - CCProxy for Claude Code
description: Install CCProxy on Windows, macOS, and Linux. Download pre-built binaries or build from source.
keywords: CCProxy installation, AI proxy for Claude Code, Windows, macOS, Linux, binary download, Docker
---

# Installation Guide

<SocialShare />

Install CCProxy on your system using the method that works best for you. CCProxy supports all major operating systems with pre-built binaries.

## Quick Install (All Platforms)

The fastest way to install CCProxy:

```bash
curl -sSL https://raw.githubusercontent.com/orchestre-dev/ccproxy/main/install.sh | bash
```

This script will:
- Detect your operating system and architecture
- Download the appropriate binary
- Install it to `/usr/local/bin` (or `~/bin` on Windows)

## Platform-Specific Installation

### Windows

#### Download Binary
1. Visit [CCProxy Releases](https://github.com/orchestre-dev/ccproxy/releases/latest)
2. Download `ccproxy-windows-amd64.exe`
3. Place in a directory in your PATH
4. Run from Command Prompt or PowerShell:
   ```powershell
   ccproxy.exe version
   ```

### macOS

#### For Apple Silicon (M1/M2/M3)
```bash
curl -L "https://github.com/orchestre-dev/ccproxy/releases/latest/download/ccproxy-darwin-arm64" -o ccproxy
chmod +x ccproxy
sudo mv ccproxy /usr/local/bin/
```

#### For Intel Macs
```bash
curl -L "https://github.com/orchestre-dev/ccproxy/releases/latest/download/ccproxy-darwin-amd64" -o ccproxy
chmod +x ccproxy
sudo mv ccproxy /usr/local/bin/
```

### Linux

#### For x86_64 systems
```bash
curl -L "https://github.com/orchestre-dev/ccproxy/releases/latest/download/ccproxy-linux-amd64" -o ccproxy
chmod +x ccproxy
sudo mv ccproxy /usr/local/bin/
```

#### For ARM64 systems
```bash
curl -L "https://github.com/orchestre-dev/ccproxy/releases/latest/download/ccproxy-linux-arm64" -o ccproxy
chmod +x ccproxy
sudo mv ccproxy /usr/local/bin/
```

## Docker Installation

Run CCProxy in a container:

```bash
# Quick start with Docker
docker run -d -p 3456:3456 \
  -v ~/.ccproxy:/home/ccproxy/.ccproxy \
  ghcr.io/orchestre-dev/ccproxy:latest

# Or use Docker Compose
curl -O https://raw.githubusercontent.com/orchestre-dev/ccproxy/main/docker-compose.yml
docker-compose up -d
```

## Building from Source

If you want to build CCProxy yourself:

### Prerequisites
- Go 1.21 or later
- Git

### Build Steps
```bash
# Clone the repository
git clone https://github.com/orchestre-dev/ccproxy.git
cd ccproxy

# Build for your current platform
make build

# Or build for all platforms
make build-all

# Install locally
sudo make install
```

## Verify Installation

After installation, verify CCProxy is working:

```bash
# Check version
ccproxy version

# Check status
ccproxy status

# Run health check
ccproxy start
curl http://localhost:3456/health
```

## Configuration

After installation, create a configuration file:

```bash
mkdir -p ~/.ccproxy
cat > ~/.ccproxy/config.json << 'EOF'
{
  "providers": [{
    "name": "openai",
    "api_base_url": "https://api.openai.com/v1",
    "api_key": "your-api-key",
    "models": ["gpt-4", "gpt-3.5-turbo"],
    "enabled": true
  }]
}
EOF
```

## Next Steps

- [Quick Start Guide](/guide/quick-start) - Get running in 2 minutes
- [Configuration Guide](/guide/configuration) - Detailed configuration options
- [Provider Setup](/providers/) - Configure AI providers

## Troubleshooting

### Permission Denied
If you get "permission denied" when running the install script:
```bash
chmod +x install.sh
./install.sh
```

### Command Not Found
If `ccproxy` is not found after installation, add the install directory to your PATH:
```bash
export PATH="$PATH:/usr/local/bin"
```

### Binary Not Found for Platform
Available binaries:
- `ccproxy-linux-amd64` - Linux x86_64
- `ccproxy-linux-arm64` - Linux ARM64
- `ccproxy-darwin-amd64` - macOS Intel
- `ccproxy-darwin-arm64` - macOS Apple Silicon
- `ccproxy-windows-amd64.exe` - Windows x86_64

If your platform isn't supported, please [build from source](#building-from-source).