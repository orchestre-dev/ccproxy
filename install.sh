#!/bin/bash

# CCProxy Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/praneybehl/ccproxy/main/install.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="praneybehl/ccproxy"
BINARY_NAME="ccproxy"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

detect_os() {
    local os
    case "$(uname -s)" in
        Darwin*)
            os="darwin"
            ;;
        Linux*)
            os="linux"
            ;;
        CYGWIN*|MINGW*|MSYS*)
            os="windows"
            ;;
        *)
            log_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    echo "$os"
}

detect_arch() {
    local arch
    case "$(uname -m)" in
        x86_64|amd64)
            arch="amd64"
            ;;
        arm64|aarch64)
            arch="arm64"
            ;;
        *)
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    echo "$arch"
}

get_latest_release() {
    local release_url="https://api.github.com/repos/${REPO}/releases/latest"
    
    if command -v curl >/dev/null 2>&1; then
        curl -s "$release_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$release_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        log_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
}

download_binary() {
    local os="$1"
    local arch="$2"
    local version="$3"
    local filename="${BINARY_NAME}-${os}-${arch}"
    
    if [ "$os" = "windows" ]; then
        filename="${filename}.exe"
    fi
    
    local download_url="https://github.com/${REPO}/releases/download/${version}/${filename}"
    local temp_file="/tmp/${filename}"
    
    log_info "Downloading ${filename} from ${download_url}"
    
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$temp_file" "$download_url"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$temp_file" "$download_url"
    else
        log_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
    
    if [ ! -f "$temp_file" ]; then
        log_error "Failed to download ${filename}"
        exit 1
    fi
    
    echo "$temp_file"
}

install_binary() {
    local temp_file="$1"
    local install_path="${INSTALL_DIR}/${BINARY_NAME}"
    
    log_info "Installing CCProxy to ${install_path}"
    
    # Check if we need sudo
    if [ ! -w "$INSTALL_DIR" ]; then
        if command -v sudo >/dev/null 2>&1; then
            sudo mv "$temp_file" "$install_path"
            sudo chmod +x "$install_path"
        else
            log_error "No write permission to ${INSTALL_DIR} and sudo not available"
            log_info "Please run: mv $temp_file $install_path && chmod +x $install_path"
            exit 1
        fi
    else
        mv "$temp_file" "$install_path"
        chmod +x "$install_path"
    fi
    
    log_success "CCProxy installed successfully to ${install_path}"
}

verify_installation() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local version
        version=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown")
        log_success "CCProxy is ready! Version: ${version}"
        log_info "Run 'ccproxy --help' to get started"
    else
        log_warning "CCProxy installed but not found in PATH"
        log_info "Add ${INSTALL_DIR} to your PATH or run: export PATH=\"${INSTALL_DIR}:\$PATH\""
    fi
}

print_next_steps() {
    echo
    log_info "Next steps:"
    echo "1. Configure your AI provider (e.g., export PROVIDER=groq)"
    echo "2. Set your API key (e.g., export GROQ_API_KEY=your_key)"
    echo "3. Start CCProxy: ccproxy"
    echo "4. Configure Claude Code: export ANTHROPIC_BASE_URL=http://localhost:7187"
    echo
    log_info "For detailed setup instructions, visit:"
    echo "   https://ccproxy.dev/guide/configuration"
}

# Main installation process
main() {
    echo
    log_info "🚀 CCProxy Installation Script"
    echo
    
    # Detect system
    local os arch version temp_file
    os=$(detect_os)
    arch=$(detect_arch)
    
    log_info "Detected system: ${os}/${arch}"
    
    # Get latest release
    log_info "Fetching latest release information..."
    version=$(get_latest_release)
    
    if [ -z "$version" ]; then
        log_error "Failed to get latest release information"
        exit 1
    fi
    
    log_info "Latest version: ${version}"
    
    # Download binary
    temp_file=$(download_binary "$os" "$arch" "$version")
    
    # Install binary
    install_binary "$temp_file"
    
    # Verify installation
    verify_installation
    
    # Print next steps
    print_next_steps
    
    log_success "Installation complete! 🎉"
}

# Run main function
main "$@"