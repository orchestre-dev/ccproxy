#!/usr/bin/env bash
# CCProxy release script

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="ccproxy"
BUILD_DIR="build"
DIST_DIR="dist"

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

# Print step message
step() {
    print_msg "$BLUE" "▶️  $*"
}

# Check prerequisites
check_prerequisites() {
    step "Checking prerequisites..."
    
    local missing=()
    
    # Check for required commands
    command -v git >/dev/null 2>&1 || missing+=("git")
    command -v go >/dev/null 2>&1 || missing+=("go")
    command -v zip >/dev/null 2>&1 || missing+=("zip")
    command -v tar >/dev/null 2>&1 || missing+=("tar")
    
    if [ ${#missing[@]} -ne 0 ]; then
        error "Missing required commands: ${missing[*]}"
        exit 1
    fi
    
    # Check for clean git state
    if [ -n "$(git status --porcelain)" ]; then
        warn "Git working directory is not clean"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    info "Prerequisites check passed"
}

# Get version
get_version() {
    local version="${1:-}"
    
    if [ -z "$version" ]; then
        # Try to get version from git tag
        version=$(git describe --tags --exact-match 2>/dev/null || echo "")
        
        if [ -z "$version" ]; then
            # Prompt for version
            read -p "Enter version (e.g., v1.0.0): " version
        fi
    fi
    
    # Validate version format
    if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-.*)?$ ]]; then
        error "Invalid version format: $version"
        error "Expected format: vX.Y.Z or vX.Y.Z-suffix"
        exit 1
    fi
    
    echo "$version"
}

# Create git tag
create_tag() {
    local version=$1
    
    step "Creating git tag $version..."
    
    if git rev-parse "$version" >/dev/null 2>&1; then
        warn "Tag $version already exists"
        read -p "Delete and recreate? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git tag -d "$version"
            git push origin --delete "$version" 2>/dev/null || true
        else
            return
        fi
    fi
    
    # Create annotated tag
    git tag -a "$version" -m "Release $version"
    info "Created tag $version"
    
    read -p "Push tag to origin? (Y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        git push origin "$version"
        info "Pushed tag to origin"
    fi
}

# Build release artifacts
build_release() {
    local version=$1
    
    step "Building release artifacts..."
    
    # Clean previous builds
    rm -rf "$BUILD_DIR" "$DIST_DIR"
    mkdir -p "$DIST_DIR"
    
    # Run tests
    info "Running tests..."
    go test -short ./... || {
        error "Tests failed"
        exit 1
    }
    
    # Build for all platforms
    VERSION="$version" make build-all
    
    # Create archives
    step "Creating release archives..."
    
    local platforms=(
        "darwin-amd64"
        "darwin-arm64"
        "linux-amd64"
        "linux-arm64"
        "windows-amd64"
    )
    
    for platform in "${platforms[@]}"; do
        local binary="$BUILD_DIR/$BINARY_NAME-$platform"
        local archive="$DIST_DIR/$BINARY_NAME-$version-$platform"
        
        if [[ "$platform" == "windows-amd64" ]]; then
            binary="$binary.exe"
            # Create zip for Windows
            info "Creating $archive.zip..."
            (cd "$BUILD_DIR" && zip -q "../$archive.zip" "$(basename "$binary")")
            # Add README and LICENSE if they exist
            [ -f "README.md" ] && zip -q "$archive.zip" README.md
            [ -f "LICENSE" ] && zip -q "$archive.zip" LICENSE
        else
            # Create tar.gz for Unix
            info "Creating $archive.tar.gz..."
            tar -czf "$archive.tar.gz" -C "$BUILD_DIR" "$(basename "$binary")"
            # Add README and LICENSE if they exist
            if [ -f "README.md" ] || [ -f "LICENSE" ]; then
                tar -rf "$archive.tar" -C . $(ls README.md LICENSE 2>/dev/null)
                gzip -f "$archive.tar"
            fi
        fi
    done
    
    info "Release artifacts created in $DIST_DIR/"
    ls -lh "$DIST_DIR"
}

# Create checksums
create_checksums() {
    local version=$1
    
    step "Creating checksums..."
    
    cd "$DIST_DIR"
    
    # Create SHA256 checksums
    shasum -a 256 *.{tar.gz,zip} > "$BINARY_NAME-$version-checksums.txt"
    
    info "Checksums created"
    cat "$BINARY_NAME-$version-checksums.txt"
    
    cd - >/dev/null
}

# Create release notes
create_release_notes() {
    local version=$1
    local notes_file="$DIST_DIR/RELEASE_NOTES.md"
    
    step "Creating release notes..."
    
    cat > "$notes_file" << EOF
# Release Notes for $version

## Changes in this Release

$(git log $(git describe --tags --abbrev=0 HEAD^)..HEAD --pretty=format:"- %s" 2>/dev/null || echo "- Initial release")

## Installation

### macOS/Linux

\`\`\`bash
# Download and extract
curl -sSL https://github.com/orchestre-dev/ccproxy/releases/download/$version/ccproxy-$version-\$(uname -s | tr '[:upper:]' '[:lower:]')-\$(uname -m | sed 's/x86_64/amd64/').tar.gz | tar xz

# Move to PATH
sudo mv ccproxy /usr/local/bin/

# Verify installation
ccproxy --version
\`\`\`

### Windows

1. Download the Windows executable from the releases page
2. Add to your PATH or move to a directory in your PATH
3. Run \`ccproxy --version\` to verify

### Docker

\`\`\`bash
docker pull ghcr.io/orchestre-dev/ccproxy:$version
\`\`\`

### Homebrew (macOS/Linux)

\`\`\`bash
brew tap orchestre-dev/tap
brew install ccproxy
\`\`\`

## Checksums

\`\`\`
$(cat "$DIST_DIR/$BINARY_NAME-$version-checksums.txt")
\`\`\`
EOF
    
    info "Release notes created: $notes_file"
}

# Main execution
main() {
    info "CCProxy Release Script"
    echo ""
    
    # Check prerequisites
    check_prerequisites
    
    # Get version
    VERSION=$(get_version "${1:-}")
    info "Releasing version: $VERSION"
    echo ""
    
    # Create git tag
    create_tag "$VERSION"
    echo ""
    
    # Build release artifacts
    build_release "$VERSION"
    echo ""
    
    # Create checksums
    create_checksums "$VERSION"
    echo ""
    
    # Create release notes
    create_release_notes "$VERSION"
    echo ""
    
    step "Release preparation complete! 🎉"
    echo ""
    info "Next steps:"
    echo "1. Review the release artifacts in $DIST_DIR/"
    echo "2. Create a GitHub release for tag $VERSION"
    echo "3. Upload the artifacts from $DIST_DIR/"
    echo "4. Paste the content from RELEASE_NOTES.md"
    echo ""
    echo "Or use GitHub CLI:"
    echo "  gh release create $VERSION $DIST_DIR/* --notes-file $DIST_DIR/RELEASE_NOTES.md"
}

# Run main function
main "$@"