#!/bin/bash

# Generate checksums for release binaries

set -euo pipefail

# Colors
readonly GREEN='\033[0;32m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# Check if dist directory exists
if [ ! -d "dist" ]; then
    echo "Error: dist directory not found. Run 'make build-all' first."
    exit 1
fi

echo -e "${BLUE}Generating checksums for release binaries...${NC}"

# Change to dist directory
cd dist

# Generate SHA256 checksums
if command -v sha256sum &> /dev/null; then
    sha256sum ccproxy-* > checksums.txt
elif command -v shasum &> /dev/null; then
    shasum -a 256 ccproxy-* > checksums.txt
else
    echo "Error: No SHA256 tool found (sha256sum or shasum required)"
    exit 1
fi

echo -e "${GREEN}Checksums generated in dist/checksums.txt${NC}"
echo
echo "Contents:"
cat checksums.txt