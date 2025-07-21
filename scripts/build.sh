#!/usr/bin/env bash
# CCProxy build script

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="ccproxy"
BUILD_DIR="build"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}"

# Parse command line arguments
PLATFORM="${1:-current}"
VERBOSE="${VERBOSE:-false}"

# Print colored message
print_msg() {
    local color=$1
    shift
    echo -e "${color}$*${NC}"
}

# Print info message
info() {
    print_msg "$GREEN" "ℹ️  $*"
}

# Print error message
error() {
    print_msg "$RED" "❌ $*"
}

# Print warning message
warn() {
    print_msg "$YELLOW" "⚠️  $*"
}

# Build for current platform
build_current() {
    local output="$BUILD_DIR/$BINARY_NAME"
    
    info "Building $BINARY_NAME v$VERSION for current platform..."
    
    CGO_ENABLED=0 go build \
        -ldflags "-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.Commit=$COMMIT -s -w" \
        -o "$output" \
        ./cmd/ccproxy
    
    info "Build complete: $output"
    ls -lh "$output"
}

# Build for specific platform
build_platform() {
    local goos=$1
    local goarch=$2
    local output="$BUILD_DIR/$BINARY_NAME-$goos-$goarch"
    
    if [ "$goos" = "windows" ]; then
        output="$output.exe"
    fi
    
    info "Building for $goos/$goarch..."
    
    GOOS=$goos GOARCH=$goarch CGO_ENABLED=0 go build \
        -ldflags "-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.Commit=$COMMIT -s -w" \
        -o "$output" \
        ./cmd/ccproxy
    
    info "Built: $output"
}

# Build for all platforms
build_all() {
    local platforms=(
        "darwin/amd64"
        "darwin/arm64"
        "linux/amd64"
        "linux/arm64"
        "windows/amd64"
    )
    
    info "Building $BINARY_NAME v$VERSION for all platforms..."
    
    for platform in "${platforms[@]}"; do
        IFS='/' read -r goos goarch <<< "$platform"
        build_platform "$goos" "$goarch"
    done
    
    info "All builds complete!"
    ls -lh "$BUILD_DIR"
}

# Main execution
main() {
    # Create build directory
    mkdir -p "$BUILD_DIR"
    
    # Show build info
    info "Build Information:"
    echo "  Binary:    $BINARY_NAME"
    echo "  Version:   $VERSION"
    echo "  Commit:    $COMMIT"
    echo "  Time:      $BUILD_TIME"
    echo "  Platform:  $PLATFORM"
    echo ""
    
    # Ensure dependencies are up to date
    info "Checking dependencies..."
    go mod download
    
    # Run build based on platform
    case "$PLATFORM" in
        current)
            build_current
            ;;
        all)
            build_all
            ;;
        */*)
            IFS='/' read -r goos goarch <<< "$PLATFORM"
            build_platform "$goos" "$goarch"
            ;;
        *)
            error "Invalid platform: $PLATFORM"
            echo "Usage: $0 [current|all|os/arch]"
            echo "Examples:"
            echo "  $0             # Build for current platform"
            echo "  $0 all         # Build for all platforms"
            echo "  $0 linux/amd64 # Build for specific platform"
            exit 1
            ;;
    esac
    
    info "Build script completed successfully! ✅"
}

# Run main function
main