#!/bin/bash

# CCProxy Installation Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture names
case "$ARCH" in
    x86_64)
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

# Set binary name based on OS
BINARY_NAME="ccproxy-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
fi

# Installation directory
INSTALL_DIR="/usr/local/bin"
if [ "$OS" = "windows" ]; then
    INSTALL_DIR="$HOME/bin"
fi

echo -e "${GREEN}Installing CCProxy for ${OS}/${ARCH}...${NC}"

# Check if binary exists
if [ ! -f "dist/$BINARY_NAME" ]; then
    echo -e "${RED}Binary not found: dist/$BINARY_NAME${NC}"
    echo "Please run 'make build-all' first"
    exit 1
fi

# Create install directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Copy binary
echo "Installing to $INSTALL_DIR/ccproxy..."
cp "dist/$BINARY_NAME" "$INSTALL_DIR/ccproxy"
chmod +x "$INSTALL_DIR/ccproxy"

# Verify installation
if command -v ccproxy &> /dev/null; then
    echo -e "${GREEN}CCProxy installed successfully!${NC}"
    ccproxy version
else
    echo -e "${YELLOW}CCProxy installed to $INSTALL_DIR/ccproxy${NC}"
    echo -e "${YELLOW}Please add $INSTALL_DIR to your PATH${NC}"
fi