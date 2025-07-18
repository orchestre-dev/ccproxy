#!/bin/bash

# Local Development Deployment Script for CCProxy Documentation
# Quick script for testing deployments locally

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}🚀 CCProxy Local Deployment${NC}"
echo "=================================="

# Check if we're in the right directory
if [ ! -f "docs/package.json" ]; then
    echo -e "${YELLOW}⚠️  Please run this script from the project root directory${NC}"
    exit 1
fi

# Install dependencies
echo -e "${BLUE}📦 Installing dependencies...${NC}"
cd docs
npm install

# Build the site
echo -e "${BLUE}🔨 Building VitePress site...${NC}"
npm run build

# Check if build was successful
if [ -d ".vitepress/dist" ]; then
    echo -e "${GREEN}✅ Build successful!${NC}"
    echo -e "${BLUE}📁 Build output location: docs/.vitepress/dist/${NC}"
    echo ""
    echo "🌐 To serve locally:"
    echo "   cd docs && npm run preview"
    echo ""
    echo "☁️  To deploy to Cloudflare Pages:"
    echo "   ./scripts/deploy-cloudflare.sh --production"
else
    echo "❌ Build failed!"
    exit 1
fi

cd ..