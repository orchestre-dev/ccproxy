#!/bin/bash

# CCProxy Installation Script
# Downloads and installs the latest CCProxy release from GitHub

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# GitHub repository
REPO="orchestre-dev/ccproxy"
GITHUB_API="https://api.github.com/repos/${REPO}"
GITHUB_DOWNLOAD="https://github.com/${REPO}/releases/download"

# Installation directory
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="ccproxy"

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    # Map OS names
    case "$OS" in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            ;;
        *)
            echo -e "${RED}Unsupported operating system: $OS${NC}"
            exit 1
            ;;
    esac
    
    # Map architecture names
    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    # Set platform string
    PLATFORM="${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="${BINARY_NAME}.exe"
    fi
    
    echo -e "${BLUE}Detected platform: ${PLATFORM}${NC}"
}

# Get latest release version
get_latest_version() {
    echo -e "${BLUE}Fetching latest release information...${NC}"
    
    # Try to get latest release from GitHub API
    if command -v curl &> /dev/null; then
        VERSION=$(curl -s "${GITHUB_API}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v?([^"]+)".*/\1/')
    elif command -v wget &> /dev/null; then
        VERSION=$(wget -qO- "${GITHUB_API}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v?([^"]+)".*/\1/')
    else
        echo -e "${RED}Neither curl nor wget found. Please install one of them.${NC}"
        exit 1
    fi
    
    if [ -z "$VERSION" ]; then
        echo -e "${RED}Failed to fetch latest version. Please check your internet connection.${NC}"
        exit 1
    fi
    
    # Remove 'v' prefix if present
    VERSION="${VERSION#v}"
    echo -e "${GREEN}Latest version: v${VERSION}${NC}"
}

# Download binary
download_binary() {
    local url="${GITHUB_DOWNLOAD}/v${VERSION}/ccproxy-${PLATFORM}"
    if [ "$OS" = "windows" ]; then
        url="${url}.exe"
    fi
    
    local temp_file="/tmp/ccproxy-download"
    
    echo -e "${BLUE}Downloading CCProxy v${VERSION} for ${PLATFORM}...${NC}"
    echo -e "${BLUE}URL: ${url}${NC}"
    
    # Download using curl or wget
    if command -v curl &> /dev/null; then
        curl -L -f -o "$temp_file" "$url" || {
            echo -e "${RED}Download failed. The binary for ${PLATFORM} might not be available.${NC}"
            echo -e "${YELLOW}Available binaries: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64${NC}"
            exit 1
        }
    else
        wget -O "$temp_file" "$url" || {
            echo -e "${RED}Download failed. The binary for ${PLATFORM} might not be available.${NC}"
            echo -e "${YELLOW}Available binaries: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64${NC}"
            exit 1
        }
    fi
    
    echo -e "${GREEN}Download completed successfully${NC}"
    echo "$temp_file"
}

# Install binary
install_binary() {
    local temp_file="$1"
    
    # Check if we need sudo
    local sudo_cmd=""
    if [ "$OS" != "windows" ] && [ ! -w "$INSTALL_DIR" ] && [ "$EUID" -ne 0 ]; then
        if command -v sudo &> /dev/null; then
            echo -e "${YELLOW}Installation requires sudo privileges${NC}"
            sudo_cmd="sudo"
        else
            echo -e "${RED}Cannot write to $INSTALL_DIR. Please run as root or specify a different INSTALL_DIR${NC}"
            exit 1
        fi
    fi
    
    # Create install directory if needed
    if [ ! -d "$INSTALL_DIR" ]; then
        echo -e "${BLUE}Creating installation directory: $INSTALL_DIR${NC}"
        $sudo_cmd mkdir -p "$INSTALL_DIR"
    fi
    
    # Move binary to install location
    echo -e "${BLUE}Installing to ${INSTALL_DIR}/${BINARY_NAME}...${NC}"
    $sudo_cmd mv "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    $sudo_cmd chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    
    # Verify installation
    if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        echo -e "${GREEN}CCProxy installed successfully!${NC}"
    else
        echo -e "${RED}Installation failed${NC}"
        exit 1
    fi
}

# Verify installation
verify_installation() {
    # Check if binary is in PATH
    if command -v ccproxy &> /dev/null; then
        echo -e "${GREEN}CCProxy is available in your PATH${NC}"
        ccproxy version
    else
        echo -e "${YELLOW}CCProxy installed to ${INSTALL_DIR}/${BINARY_NAME}${NC}"
        if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
            echo -e "${YELLOW}Add ${INSTALL_DIR} to your PATH to use ccproxy from anywhere${NC}"
            echo -e "${YELLOW}You can run: export PATH=\"\$PATH:${INSTALL_DIR}\"${NC}"
        fi
        # Show version using full path
        "${INSTALL_DIR}/${BINARY_NAME}" version
    fi
}

# Main installation flow
main() {
    echo -e "${GREEN}=== CCProxy Installation ===${NC}"
    echo
    
    # Detect platform
    detect_platform
    
    # Get latest version (or use provided version)
    if [ -n "$1" ]; then
        VERSION="${1#v}"  # Remove 'v' prefix if present
        echo -e "${BLUE}Installing specific version: v${VERSION}${NC}"
    else
        get_latest_version
    fi
    
    # Download binary
    temp_file=$(download_binary)
    
    # Install binary
    install_binary "$temp_file"
    
    # Verify installation
    verify_installation
    
    echo
    echo -e "${GREEN}=== Installation Complete ===${NC}"
    echo
    echo -e "${BLUE}Quick start:${NC}"
    echo "  1. Create a config file with your provider API keys"
    echo "  2. Run: ccproxy start"
    echo "  3. Run: ccproxy code  # For Claude Code integration"
    echo
    echo -e "${BLUE}For more information, visit: https://github.com/${REPO}${NC}"
}

# Run main function
main "$@"