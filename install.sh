#!/bin/bash

# CCProxy Installation Script
# Downloads and installs the latest CCProxy release from GitHub

set -euo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# GitHub repository
readonly REPO="orchestre-dev/ccproxy"
readonly GITHUB_API="https://api.github.com/repos/${REPO}"
readonly GITHUB_DOWNLOAD="https://github.com/${REPO}/releases/download"

# Installation directory
readonly DEFAULT_INSTALL_DIR="/usr/local/bin"
INSTALL_DIR="${INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"
BINARY_NAME="ccproxy"

# Security: Validate installation directory
validate_install_dir() {
    # Remove trailing slashes
    INSTALL_DIR="${INSTALL_DIR%/}"
    
    # Check for path traversal attempts
    if [[ "$INSTALL_DIR" =~ \.\. ]]; then
        echo -e "${RED}Error: Invalid installation directory (contains ..)${NC}"
        exit 1
    fi
    
    # Ensure absolute path
    if [[ ! "$INSTALL_DIR" = /* ]]; then
        echo -e "${RED}Error: Installation directory must be an absolute path${NC}"
        exit 1
    fi
    
    # Validate path characters
    if [[ ! "$INSTALL_DIR" =~ ^[a-zA-Z0-9/_-]+$ ]]; then
        echo -e "${RED}Error: Invalid characters in installation directory${NC}"
        exit 1
    fi
}

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

# Validate version format
validate_version() {
    local version="$1"
    # Version should be in format: digits.digits.digits (optionally with v prefix)
    # Only allow numbers, dots, and optional v prefix
    if [[ ! "$version" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?(\+[a-zA-Z0-9]+)?$ ]]; then
        echo -e "${RED}Invalid version format: $version${NC}"
        echo -e "${RED}Expected format: v1.2.3 or 1.2.3${NC}"
        exit 1
    fi
    
    # Additional check for reasonable version numbers
    local major minor patch
    IFS='.' read -r major minor patch <<< "${version#v}"
    if (( major > 999 )) || (( minor > 999 )) || (( ${patch%%[-+]*} > 999 )); then
        echo -e "${RED}Version numbers seem unreasonably high${NC}"
        exit 1
    fi
}

# Get latest release version with validation
get_latest_version() {
    echo -e "${BLUE}Fetching latest release information...${NC}"
    
    # Create temp file for API response with secure permissions
    local temp_response=$(mktemp -t ccproxy-api-XXXXXX)
    chmod 600 "$temp_response"
    trap "rm -f $temp_response" EXIT
    
    # Try to get latest release from GitHub API
    if command -v curl &> /dev/null; then
        if ! curl -sfL \
            -H "Accept: application/vnd.github.v3+json" \
            "${GITHUB_API}/releases/latest" \
            -o "$temp_response"; then
            echo -e "${RED}Failed to fetch latest version from GitHub${NC}"
            exit 1
        fi
    elif command -v wget &> /dev/null; then
        if ! wget -qO "$temp_response" \
            --header="Accept: application/vnd.github.v3+json" \
            "${GITHUB_API}/releases/latest"; then
            echo -e "${RED}Failed to fetch latest version from GitHub${NC}"
            exit 1
        fi
    else
        echo -e "${RED}Neither curl nor wget found. Please install one of them.${NC}"
        exit 1
    fi
    
    # Validate JSON response first
    if ! grep -q '"tag_name"' "$temp_response"; then
        echo -e "${RED}Invalid GitHub API response - no tag_name found${NC}"
        exit 1
    fi
    
    # Extract version safely using a more restrictive pattern
    VERSION=$(grep '"tag_name"' "$temp_response" | head -1 | sed -E 's/.*"tag_name"[[:space:]]*:[[:space:]]*"(v?[0-9]+\.[0-9]+\.[0-9]+[^"]*)".*/\1/')
    
    if [ -z "$VERSION" ]; then
        echo -e "${RED}Failed to parse version from GitHub response${NC}"
        exit 1
    fi
    
    # Validate version format
    validate_version "$VERSION"
    
    # Remove 'v' prefix if present
    VERSION="${VERSION#v}"
    echo -e "${GREEN}Latest version: v${VERSION}${NC}"
}

# Validate platform before using in URL
validate_platform() {
    # Ensure PLATFORM only contains expected values
    case "$PLATFORM" in
        linux-amd64|linux-arm64|darwin-amd64|darwin-arm64|windows-amd64)
            # Valid platform
            ;;
        *)
            echo -e "${RED}Invalid platform: $PLATFORM${NC}"
            exit 1
            ;;
    esac
}

# Download binary with checksum verification
download_binary() {
    # Validate platform before constructing URL
    validate_platform
    
    # URL encode the version (in case it contains special chars)
    local encoded_version=$(printf '%s' "v${VERSION}" | sed 's/[^a-zA-Z0-9._-]/_/g')
    
    local url="${GITHUB_DOWNLOAD}/${encoded_version}/ccproxy-${PLATFORM}"
    if [ "$OS" = "windows" ]; then
        url="${url}.exe"
    fi
    
    local checksum_url="${GITHUB_DOWNLOAD}/v${VERSION}/checksums.txt"
    # Create secure temp directory
    local temp_dir=$(mktemp -d -t ccproxy-install-XXXXXX)
    chmod 700 "$temp_dir"
    # Note: Cleanup is handled in main function after installation
    
    local temp_file="${temp_dir}/ccproxy-download"
    local checksum_file="${temp_dir}/checksums.txt"
    
    echo -e "${BLUE}Downloading CCProxy v${VERSION} for ${PLATFORM}...${NC}" >&2
    
    # Download binary
    if command -v curl &> /dev/null; then
        if ! curl -fL -o "$temp_file" "$url"; then
            echo -e "${RED}Download failed. The binary for ${PLATFORM} might not be available.${NC}" >&2
            echo -e "${YELLOW}Available binaries: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64${NC}" >&2
            exit 1
        fi
    else
        if ! wget -O "$temp_file" "$url"; then
            echo -e "${RED}Download failed. The binary for ${PLATFORM} might not be available.${NC}" >&2
            echo -e "${YELLOW}Available binaries: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64${NC}" >&2
            exit 1
        fi
    fi
    
    # Verify the downloaded file is actually a binary
    if [ -f "$temp_file" ]; then
        if command -v file &> /dev/null; then
            local file_type=$(file -b "$temp_file" 2>/dev/null || true)
            case "$file_type" in
                *executable*|*binary*|*Mach-O*|*PE32*|*ELF*)
                    # Valid binary file
                    ;;
                *HTML*|*text*|*ASCII*)
                    echo -e "${RED}Downloaded file appears to be HTML/text, not a binary${NC}" >&2
                    echo -e "${RED}This might indicate an error page was downloaded${NC}" >&2
                    exit 1
                    ;;
            esac
        fi
        
        # Check file size (binaries should be at least 1MB)
        local file_size=$(stat -f%z "$temp_file" 2>/dev/null || stat -c%s "$temp_file" 2>/dev/null || echo 0)
        if [ "$file_size" -lt 1048576 ]; then
            echo -e "${YELLOW}Warning: Downloaded file is unusually small (${file_size} bytes)${NC}" >&2
        fi
    else
        echo -e "${RED}Download failed - file not created${NC}" >&2
        exit 1
    fi
    
    echo -e "${BLUE}Downloading checksums...${NC}" >&2
    
    # Download checksums
    if command -v curl &> /dev/null; then
        if ! curl -fL -o "$checksum_file" "$checksum_url" 2>/dev/null; then
            echo -e "${YELLOW}Warning: Checksums not available for this release${NC}" >&2
            echo -e "${YELLOW}Proceeding without checksum verification${NC}" >&2
        else
            verify_checksum "$temp_file" "$checksum_file"
        fi
    else
        if ! wget -O "$checksum_file" "$checksum_url" 2>/dev/null; then
            echo -e "${YELLOW}Warning: Checksums not available for this release${NC}" >&2
            echo -e "${YELLOW}Proceeding without checksum verification${NC}" >&2
        else
            verify_checksum "$temp_file" "$checksum_file"
        fi
    fi
    
    echo -e "${GREEN}Download completed successfully${NC}" >&2
    echo "$temp_file:$temp_dir"
}

# Verify checksum
verify_checksum() {
    local file="$1"
    local checksum_file="$2"
    local filename="ccproxy-${PLATFORM}"
    
    if [ "$OS" = "windows" ]; then
        filename="${filename}.exe"
    fi
    
    echo -e "${BLUE}Verifying checksum...${NC}"
    
    # Check if we have sha256sum or shasum
    local sha_cmd=""
    if command -v sha256sum &> /dev/null; then
        sha_cmd="sha256sum"
    elif command -v shasum &> /dev/null; then
        sha_cmd="shasum -a 256"
    else
        echo -e "${YELLOW}Warning: No SHA256 tool found, skipping checksum verification${NC}"
        return
    fi
    
    # Calculate checksum of downloaded file
    local actual_checksum=$($sha_cmd "$file" | awk '{print $1}')
    
    # Extract expected checksum from file
    local expected_checksum=$(grep "$filename" "$checksum_file" 2>/dev/null | awk '{print $1}')
    
    if [ -z "$expected_checksum" ]; then
        echo -e "${YELLOW}Warning: Checksum not found for $filename${NC}" >&2
        return
    fi
    
    if [ "$actual_checksum" != "$expected_checksum" ]; then
        echo -e "${RED}Checksum verification failed!${NC}" >&2
        echo -e "${RED}Expected: $expected_checksum${NC}" >&2
        echo -e "${RED}Actual:   $actual_checksum${NC}" >&2
        exit 1
    fi
    
    echo -e "${GREEN}Checksum verification passed${NC}" >&2
}

# Install binary with proper permission checks
install_binary() {
    local temp_file="$1"
    
    # Validate install directory
    validate_install_dir
    
    # Check if we need sudo
    local use_sudo=false
    local sudo_cmd=""
    
    if [ "$OS" != "windows" ]; then
        if [ ! -w "$INSTALL_DIR" ] && [ "$EUID" -ne 0 ]; then
            if command -v sudo &> /dev/null; then
                echo -e "${YELLOW}Installation to $INSTALL_DIR requires administrator privileges${NC}"
                echo -e "${BLUE}The following operations will be performed:${NC}"
                echo -e "  1. Create directory (if needed): $INSTALL_DIR"
                echo -e "  2. Copy binary to: ${INSTALL_DIR}/${BINARY_NAME}"
                echo -e "  3. Set executable permissions"
                echo
                read -p "Do you want to proceed with sudo? [y/N] " -n 1 -r
                echo
                if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                    echo -e "${RED}Installation cancelled${NC}"
                    exit 1
                fi
                use_sudo=true
                sudo_cmd="sudo"
            else
                echo -e "${RED}Cannot write to $INSTALL_DIR. Please run as root or specify a different INSTALL_DIR${NC}"
                echo -e "${YELLOW}Example: INSTALL_DIR=~/bin $0${NC}"
                exit 1
            fi
        fi
    fi
    
    # Create install directory if needed
    if [ ! -d "$INSTALL_DIR" ]; then
        echo -e "${BLUE}Creating installation directory: $INSTALL_DIR${NC}"
        if [ "$use_sudo" = true ]; then
            $sudo_cmd mkdir -p "$INSTALL_DIR"
        else
            mkdir -p "$INSTALL_DIR"
        fi
    fi
    
    # Install binary
    echo -e "${BLUE}Installing to ${INSTALL_DIR}/${BINARY_NAME}...${NC}"
    
    # Debug: Check if temp file exists
    if [ ! -f "$temp_file" ]; then
        echo -e "${RED}Error: Downloaded file not found at: $temp_file${NC}"
        exit 1
    fi
    
    if [ "$use_sudo" = true ]; then
        $sudo_cmd cp "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
        $sudo_cmd chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        cp "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    # Verify installation
    if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        echo -e "${GREEN}CCProxy installed successfully!${NC}"
    else
        echo -e "${RED}Installation failed${NC}"
        exit 1
    fi
}

# Validate JSON configuration
validate_json_config() {
    local file="$1"
    
    if [ ! -f "$file" ]; then
        return 1
    fi
    
    # Try different JSON validators in order of preference
    if command -v python3 &> /dev/null; then
        if python3 -c "import json; json.load(open('$file'))" 2>/dev/null; then
            echo -e "${GREEN}Configuration validated successfully${NC}"
            return 0
        else
            echo -e "${YELLOW}Warning: Configuration has JSON syntax issues${NC}"
            return 1
        fi
    elif command -v python &> /dev/null; then
        if python -c "import json; json.load(open('$file'))" 2>/dev/null; then
            echo -e "${GREEN}Configuration validated successfully${NC}"
            return 0
        else
            echo -e "${YELLOW}Warning: Configuration has JSON syntax issues${NC}"
            return 1
        fi
    elif command -v jq &> /dev/null; then
        if jq . "$file" > /dev/null 2>&1; then
            echo -e "${GREEN}Configuration validated successfully${NC}"
            return 0
        else
            echo -e "${YELLOW}Warning: Configuration has JSON syntax issues${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}Note: Could not validate JSON (no validator found)${NC}"
        return 0  # Don't fail if no validator available
    fi
}

# Backup existing configuration
backup_config() {
    local config_file="$1"
    
    if [ -f "$config_file" ]; then
        local backup_file="${config_file}.backup.$(date +%Y%m%d_%H%M%S)"
        cp "$config_file" "$backup_file"
        echo -e "${BLUE}Backed up existing config to: $backup_file${NC}"
    fi
}

# Setup CCProxy configuration
setup_config() {
    local config_dir="$HOME/.ccproxy"
    local config_file="$config_dir/config.json"
    
    # Create config directory
    if [ ! -d "$config_dir" ]; then
        echo -e "${BLUE}Creating configuration directory: $config_dir${NC}"
        mkdir -p "$config_dir"
    fi
    
    # Create default config if it doesn't exist
    if [ ! -f "$config_file" ]; then
        echo -e "${BLUE}Creating default configuration file...${NC}"
        cat > "$config_file" << 'EOF'
{
  "providers": [
    {
      "name": "openai",
      "api_key": "your-openai-api-key-here",
      "api_base_url": "https://api.openai.com/v1",
      "models": ["gpt-4o", "gpt-4o-mini"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "gpt-4o"
    }
  }
}
EOF
        echo -e "${GREEN}Created default configuration at: $config_file${NC}"
        
        # Create example config with additional providers
        local example_file="$config_dir/config.example.json"
        cat > "$example_file" << 'EOF'
{
  "_comment": "Example configuration with multiple providers",
  "providers": [
    {
      "name": "openai",
      "api_key": "your-openai-api-key-here",
      "api_base_url": "https://api.openai.com/v1",
      "models": ["gpt-4o", "gpt-4o-mini"],
      "enabled": true
    },
    {
      "_comment": "Anthropic Claude models",
      "name": "anthropic",
      "api_key": "sk-ant-...",
      "api_base_url": "https://api.anthropic.com",
      "models": ["claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022"],
      "enabled": false
    },
    {
      "_comment": "Google Gemini models",
      "name": "gemini",
      "api_key": "AI...",
      "api_base_url": "https://generativelanguage.googleapis.com/v1",
      "models": ["gemini-2.0-flash-exp", "gemini-1.5-pro"],
      "enabled": false
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "gpt-4o"
    },
    "_comment_routes": "Special routes can be added here",
    "longContext": {
      "provider": "anthropic",
      "model": "claude-3-5-sonnet-20241022"
    }
  }
}
EOF
        echo -e "${BLUE}Example configuration saved at: $example_file${NC}"
        
        # Validate the generated config
        validate_json_config "$config_file"
    else
        echo -e "${YELLOW}Configuration already exists at: $config_file${NC}"
        # Optionally backup existing config
        backup_config "$config_file"
    fi
}

# Update PATH in shell configuration
update_shell_path() {
    local shell_rc=""
    local shell_name=""
    
    # Determine shell configuration file
    if [ -n "$ZSH_VERSION" ]; then
        shell_rc="$HOME/.zshrc"
        shell_name="zsh"
    elif [ -n "$BASH_VERSION" ]; then
        shell_rc="$HOME/.bashrc"
        shell_name="bash"
    else
        # Try to detect from SHELL variable
        case "$SHELL" in
            */zsh)
                shell_rc="$HOME/.zshrc"
                shell_name="zsh"
                ;;
            */bash)
                shell_rc="$HOME/.bashrc"
                shell_name="bash"
                ;;
            *)
                echo -e "${YELLOW}Warning: Could not detect shell type${NC}"
                return
                ;;
        esac
    fi
    
    # Enhanced PATH detection - check more thoroughly
    local path_needs_update=false
    
    # Check if directory is in PATH using multiple methods
    if ! command -v "$BINARY_NAME" &> /dev/null; then
        # Binary not found, check if install dir is in PATH
        if ! echo "$PATH" | tr ':' '\n' | grep -Fx "$INSTALL_DIR" > /dev/null 2>&1; then
            path_needs_update=true
        fi
    fi
    
    if [ "$path_needs_update" = true ]; then
        echo -e "${BLUE}Adding $INSTALL_DIR to PATH in $shell_rc${NC}"
        
        # Add PATH update to shell rc file
        if [ -f "$shell_rc" ]; then
            # Check if PATH export already exists for this directory
            if ! grep -q "export PATH=.*$INSTALL_DIR" "$shell_rc"; then
                echo "" >> "$shell_rc"
                echo "# Added by CCProxy installer" >> "$shell_rc"
                echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$shell_rc"
                echo -e "${GREEN}Updated PATH in $shell_rc${NC}"
                echo -e "${YELLOW}Note: Run 'source $shell_rc' or start a new terminal for PATH changes to take effect${NC}"
            fi
        else
            echo -e "${YELLOW}Warning: Shell configuration file $shell_rc not found${NC}"
            echo -e "${YELLOW}You may need to manually add $INSTALL_DIR to your PATH${NC}"
        fi
    else
        echo -e "${GREEN}$INSTALL_DIR is already in your PATH${NC}"
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

# Check for concurrent installation
check_concurrent_install() {
    local lock_file="/tmp/ccproxy_install.lock"
    local pid_file="/tmp/ccproxy_install.pid"
    
    # Check if lock file exists
    if [ -f "$lock_file" ]; then
        # Check if the PID is still running
        if [ -f "$pid_file" ]; then
            local old_pid=$(cat "$pid_file" 2>/dev/null)
            if [ -n "$old_pid" ] && kill -0 "$old_pid" 2>/dev/null; then
                echo -e "${RED}Another installation is already in progress (PID: $old_pid)${NC}"
                echo -e "${YELLOW}If this is incorrect, remove $lock_file and try again${NC}"
                exit 1
            fi
        fi
        # Stale lock file, remove it
        rm -f "$lock_file" "$pid_file"
    fi
    
    # Create lock file
    touch "$lock_file"
    echo $$ > "$pid_file"
    
    # Ensure cleanup on exit
    trap "rm -f '$lock_file' '$pid_file' 2>/dev/null || true" EXIT INT TERM
}

# Main installation flow
main() {
    echo -e "${GREEN}=== CCProxy Installation ===${NC}"
    echo
    
    # Check for concurrent installation
    check_concurrent_install
    
    # Detect platform
    detect_platform
    
    # Get latest version (or use provided version)
    if [ -n "${1:-}" ]; then
        # Validate user input more strictly
        local user_version="$1"
        
        # Remove any potentially dangerous characters
        user_version=$(printf '%s' "$user_version" | tr -cd 'a-zA-Z0-9._+-')
        
        # Validate format
        validate_version "$user_version"
        
        VERSION="${user_version#v}"  # Remove 'v' prefix if present
        echo -e "${BLUE}Installing specific version: v${VERSION}${NC}"
    else
        get_latest_version
    fi
    
    # Download binary with checksum verification
    download_result=$(download_binary)
    temp_file="${download_result%%:*}"
    temp_dir="${download_result##*:}"
    
    # Update trap to include temp directory cleanup
    trap "rm -rf '$temp_dir' 2>/dev/null || true; rm -f '/tmp/ccproxy_install.lock' '/tmp/ccproxy_install.pid' 2>/dev/null || true" EXIT INT TERM
    
    # Install binary
    install_binary "$temp_file"
    
    # Setup configuration
    setup_config
    
    # Update PATH if needed
    update_shell_path
    
    # Verify installation
    verify_installation
    
    echo
    echo -e "${GREEN}=== Installation Complete ===${NC}"
    echo
    echo -e "${BLUE}Next Steps:${NC}"
    echo
    echo "1. ${YELLOW}Edit your configuration file:${NC}"
    echo "   Location: $HOME/.ccproxy/config.json"
    echo "   "
    if [ -n "$EDITOR" ]; then
        echo "   Run: $EDITOR $HOME/.ccproxy/config.json"
    elif command -v code &> /dev/null; then
        echo "   Run: code $HOME/.ccproxy/config.json"
    elif command -v vim &> /dev/null; then
        echo "   Run: vim $HOME/.ccproxy/config.json"
    elif command -v nano &> /dev/null; then
        echo "   Run: nano $HOME/.ccproxy/config.json"
    else
        echo "   Open in your text editor"
    fi
    echo "   "
    echo "   Replace 'your-openai-api-key-here' with your actual API key"
    echo
    echo "2. ${YELLOW}Start CCProxy:${NC}"
    echo "   ccproxy start"
    echo
    echo "3. ${YELLOW}Configure Claude Code:${NC}"
    echo "   ccproxy code"
    echo
    echo -e "${BLUE}Documentation:${NC}"
    echo "  Configuration Guide: https://ccproxy.orchestre.dev/guide/configuration"
    echo "  Provider Setup: https://ccproxy.orchestre.dev/providers/"
    echo "  Quick Start: https://ccproxy.orchestre.dev/guide/quick-start"
    echo
    echo -e "${GREEN}Tip:${NC} If 'ccproxy' command is not found, run:"
    echo "  source ~/.bashrc  # or ~/.zshrc for zsh users"
}

# Run main function with all arguments
main "$@"