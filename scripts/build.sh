#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="ccproxy"
BUILD_DIR="bin"
VERSION=${VERSION:-"1.0.0"}
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS="-X main.version=${VERSION} -X main.commit=${COMMIT_HASH} -X main.buildTime=${BUILD_TIME} -s -w"

echo -e "${GREEN}Building ${APP_NAME} v${VERSION}${NC}"
echo -e "${YELLOW}Commit: ${COMMIT_HASH}${NC}"
echo -e "${YELLOW}Build Time: ${BUILD_TIME}${NC}"
echo ""

# Create build directory
mkdir -p ${BUILD_DIR}

# Build targets
declare -a targets=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for target in "${targets[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$target"
    
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${APP_NAME}-${GOOS}-${GOARCH}.exe"
    else
        OUTPUT_NAME="${APP_NAME}-${GOOS}-${GOARCH}"
    fi
    
    echo -e "${YELLOW}Building for ${GOOS}/${GOARCH}...${NC}"
    
    CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "$LDFLAGS" \
        -o ${BUILD_DIR}/${OUTPUT_NAME} \
        ./cmd/proxy
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Successfully built ${OUTPUT_NAME}${NC}"
        
        # Show file size
        if command -v ls >/dev/null 2>&1; then
            SIZE=$(ls -lh ${BUILD_DIR}/${OUTPUT_NAME} | awk '{print $5}')
            echo -e "  Size: ${SIZE}"
        fi
    else
        echo -e "${RED}✗ Failed to build ${OUTPUT_NAME}${NC}"
        exit 1
    fi
    echo ""
done

echo -e "${GREEN}All builds completed successfully!${NC}"
echo -e "${YELLOW}Binaries available in ${BUILD_DIR}/ directory${NC}"

# List all built binaries
echo ""
echo "Built binaries:"
ls -la ${BUILD_DIR}/

# Create checksums
echo ""
echo -e "${YELLOW}Generating checksums...${NC}"
cd ${BUILD_DIR}
if command -v sha256sum >/dev/null 2>&1; then
    sha256sum ${APP_NAME}-* > checksums.txt
    echo -e "${GREEN}✓ Checksums saved to ${BUILD_DIR}/checksums.txt${NC}"
elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 ${APP_NAME}-* > checksums.txt
    echo -e "${GREEN}✓ Checksums saved to ${BUILD_DIR}/checksums.txt${NC}"
else
    echo -e "${YELLOW}Warning: No checksum utility found${NC}"
fi

cd ..

echo ""
echo -e "${GREEN}Build process completed!${NC}"