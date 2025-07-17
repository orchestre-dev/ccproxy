---
title: Installation Guide - CCProxy for All Operating Systems
description: Complete installation guide for CCProxy on Windows, macOS, and Linux. Download binaries, use package managers, or build from source.
keywords: CCProxy installation, Windows, macOS, Linux, binary download, package manager, Docker, build from source
---

# Installation Guide

<SocialShare />

Install CCProxy on your system using the method that works best for you. CCProxy supports all major operating systems with pre-built binaries and multiple installation options.

<div id="os-detection" class="os-detection">
  <div class="os-tabs">
    <button class="os-tab" data-os="windows" onclick="switchOS('windows')">🪟 Windows</button>
    <button class="os-tab" data-os="macos" onclick="switchOS('macos')">🍎 macOS</button>
    <button class="os-tab" data-os="linux" onclick="switchOS('linux')">🐧 Linux</button>
    <button class="os-tab" data-os="docker" onclick="switchOS('docker')">🐳 Docker</button>
  </div>
</div>

## Quick Install

<div class="os-content" id="windows-content">

### Windows Installation

#### Option 1: Direct Download (Recommended)
```powershell
# Download latest Windows binary
Invoke-WebRequest -Uri "https://github.com/praneybehl/ccproxy/releases/latest/download/ccproxy-windows-amd64.exe" -OutFile "ccproxy.exe"

# Make executable and run
.\ccproxy.exe
```

#### Option 2: Using Scoop
```powershell
# Add the ccproxy bucket (coming soon)
scoop bucket add ccproxy https://github.com/praneybehl/ccproxy-scoop
scoop install ccproxy
```

#### Option 3: Using Chocolatey
```powershell
# Install via Chocolatey (coming soon)
choco install ccproxy
```

#### Option 4: Manual Download
1. Visit [CCProxy Releases](https://github.com/praneybehl/ccproxy/releases/latest)
2. Download `ccproxy-windows-amd64.exe`
3. Place in your preferred directory
4. Run from Command Prompt or PowerShell

</div>

<div class="os-content" id="macos-content">

### macOS Installation

#### Option 1: Direct Download (Recommended)
```bash
# Download for Apple Silicon (M1/M2/M3)
curl -L "https://github.com/praneybehl/ccproxy/releases/latest/download/ccproxy-darwin-arm64" -o ccproxy
chmod +x ccproxy

# For Intel Macs
curl -L "https://github.com/praneybehl/ccproxy/releases/latest/download/ccproxy-darwin-amd64" -o ccproxy
chmod +x ccproxy

# Run CCProxy
./ccproxy
```

#### Option 2: Using Homebrew
```bash
# Add the ccproxy tap (coming soon)
brew tap praneybehl/ccproxy
brew install ccproxy
```

#### Option 3: One-Line Installer
```bash
# Auto-detect architecture and install
curl -sSL https://raw.githubusercontent.com/praneybehl/ccproxy/main/install.sh | bash
```

#### Option 4: Manual Download
1. Visit [CCProxy Releases](https://github.com/praneybehl/ccproxy/releases/latest)
2. Download `ccproxy-darwin-arm64` (Apple Silicon) or `ccproxy-darwin-amd64` (Intel)
3. Make executable: `chmod +x ccproxy-*`
4. Move to PATH: `sudo mv ccproxy-* /usr/local/bin/ccproxy`

</div>

<div class="os-content" id="linux-content">

### Linux Installation

#### Option 1: Direct Download (Recommended)
```bash
# For x86_64 systems
curl -L "https://github.com/praneybehl/ccproxy/releases/latest/download/ccproxy-linux-amd64" -o ccproxy
chmod +x ccproxy

# For ARM64 systems
curl -L "https://github.com/praneybehl/ccproxy/releases/latest/download/ccproxy-linux-arm64" -o ccproxy
chmod +x ccproxy

# Install system-wide
sudo mv ccproxy /usr/local/bin/
```

#### Option 2: Using Package Managers

**Debian/Ubuntu (APT):**
```bash
# Add CCProxy repository (coming soon)
curl -fsSL https://pkg.ccproxy.dev/gpg | sudo apt-key add -
echo "deb https://pkg.ccproxy.dev/apt stable main" | sudo tee /etc/apt/sources.list.d/ccproxy.list
sudo apt update && sudo apt install ccproxy
```

**RHEL/Fedora/CentOS (YUM/DNF):**
```bash
# Add CCProxy repository (coming soon)
sudo dnf config-manager --add-repo https://pkg.ccproxy.dev/rpm/ccproxy.repo
sudo dnf install ccproxy
```

**Arch Linux (AUR):**
```bash
# Install from AUR (coming soon)
yay -S ccproxy-bin
# or
paru -S ccproxy-bin
```

#### Option 3: Snap Package
```bash
# Install via Snap (coming soon)
sudo snap install ccproxy
```

#### Option 4: AppImage
```bash
# Download and run AppImage (coming soon)
curl -L "https://github.com/praneybehl/ccproxy/releases/latest/download/ccproxy-x86_64.AppImage" -o ccproxy.AppImage
chmod +x ccproxy.AppImage
./ccproxy.AppImage
```

</div>

<div class="os-content" id="docker-content">

### Docker Installation

#### Option 1: Docker Run
```bash
# Run CCProxy in Docker
docker run -p 7187:7187 \
  -e PROVIDER=groq \
  -e GROQ_API_KEY=your_api_key \
  -e GROQ_MODEL=moonshotai/kimi-k2-instruct \
  praneybehl/ccproxy:latest
```

#### Option 2: Docker Compose
```yaml
# docker-compose.yml
version: '3.8'
services:
  ccproxy:
    image: praneybehl/ccproxy:latest
    ports:
      - "7187:7187"
    environment:
      - PROVIDER=groq
      - GROQ_API_KEY=${GROQ_API_KEY}
      - GROQ_MODEL=moonshotai/kimi-k2-instruct
    restart: unless-stopped
```

```bash
# Start with Docker Compose
docker-compose up -d
```

#### Option 3: Kubernetes
```yaml
# ccproxy-deployment.yaml
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
        image: praneybehl/ccproxy:latest
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
```

</div>

## Build from Source

For developers who want to build CCProxy from source:

```bash
# Clone the repository
git clone https://github.com/praneybehl/ccproxy.git
cd ccproxy

# Build for your platform
go build -o ccproxy cmd/proxy/main.go

# Or build for all platforms
./scripts/build.sh
```

## Verification

After installation, verify CCProxy is working:

```bash
# Check version
ccproxy --version

# Test health endpoint
curl http://localhost:7187/health

# Configure for Claude Code
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED
```

## Next Steps

1. **[Configuration](/guide/configuration)** - Set up your AI provider
2. **[Quick Start](/guide/quick-start)** - Get up and running in 2 minutes
3. **[Kimi K2 Setup](/kimi-k2)** - Experience ultra-fast AI development

## Troubleshooting

### Common Issues

**Permission Denied (macOS/Linux):**
```bash
chmod +x ccproxy
```

**Windows Security Warning:**
Right-click the executable and select "Run anyway" or add an exception to Windows Defender.

**Port Already in Use:**
```bash
# Use a different port
ccproxy --port 8187
```

**Firewall Issues:**
Ensure port 7187 is open in your firewall settings.

### Getting Help

- 📖 [Documentation](/)
- 💬 [Community Support](https://github.com/praneybehl/ccproxy/discussions) - Ask questions and get help
- 🐛 [Report Issues](https://github.com/praneybehl/ccproxy/issues) - Bug reports and feature requests
- 🔧 [Configuration Guide](/guide/configuration)

<script>
// Auto-detect operating system
function detectOS() {
  if (typeof navigator === 'undefined') return 'linux';
  const userAgent = navigator.userAgent.toLowerCase();
  if (userAgent.includes('win')) return 'windows';
  if (userAgent.includes('mac')) return 'macos';
  if (userAgent.includes('linux')) return 'linux';
  return 'linux'; // default
}

// Switch between OS tabs
function switchOS(os) {
  if (typeof document === 'undefined') return;
  
  // Add js-loaded class to body for CSS
  document.body.classList.add('js-loaded');
  
  // Hide all content and remove active classes
  document.querySelectorAll('.os-content').forEach(content => {
    content.style.display = 'none';
    content.classList.remove('active');
  });
  
  document.querySelectorAll('.os-tab').forEach(tab => {
    tab.classList.remove('active');
  });
  
  // Show selected content
  const content = document.getElementById(os + '-content');
  if (content) {
    content.style.display = 'block';
    content.classList.add('active');
  }
  
  // Add active class to selected tab
  const tab = document.querySelector(`[data-os="${os}"]`);
  if (tab) {
    tab.classList.add('active');
  }
}

// Initialize on page load (client-side only)
if (typeof window !== 'undefined') {
  function initializeOS() {
    const detectedOS = detectOS();
    console.log('Detected OS:', detectedOS); // Debug log
    switchOS(detectedOS);
  }
  
  // Initialize when DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeOS);
  } else {
    // DOM is already loaded
    initializeOS();
  }
}
</script>

<style>
.os-detection {
  margin: 24px 0;
}

.os-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
  border-bottom: 1px solid var(--vp-c-border);
}

.os-tab {
  padding: 12px 20px;
  border: none;
  background: transparent;
  color: var(--vp-c-text-2);
  cursor: pointer;
  border-radius: 6px 6px 0 0;
  font-weight: 500;
  transition: all 0.2s;
}

.os-tab:hover {
  background: var(--vp-c-bg-soft);
  color: var(--vp-c-text-1);
}

.os-tab.active {
  background: var(--vp-c-brand-1);
  color: white;
}

.os-content {
  display: none;
}

.os-content.active {
  display: block;
}

/* Fallback: show macOS by default if JavaScript fails */
#macos-content {
  display: block;
}

/* Hide all when JavaScript loads */
.js-loaded .os-content {
  display: none;
}

@media (max-width: 640px) {
  .os-tabs {
    flex-wrap: wrap;
  }
  
  .os-tab {
    flex: 1;
    min-width: 120px;
  }
}
</style>