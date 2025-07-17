#!/bin/bash

# Google Analytics Setup Script for CCProxy Documentation
# Quick setup for Google Analytics 4

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}📊 CCProxy Google Analytics Setup${NC}"
echo "=================================="
echo ""

# Check if we're in the right directory
if [ ! -f "docs/package.json" ]; then
    echo -e "${RED}❌ Please run this script from the project root directory${NC}"
    exit 1
fi

ENV_FILE="docs/.env"

echo -e "${BLUE}🔧 Setting up Google Analytics for CCProxy Documentation...${NC}"
echo ""

# Check if .env already exists
if [ -f "$ENV_FILE" ]; then
    echo -e "${YELLOW}⚠️  .env file already exists. Creating backup...${NC}"
    cp "$ENV_FILE" "$ENV_FILE.backup"
    echo -e "${GREEN}✅ Backup created: $ENV_FILE.backup${NC}"
    echo ""
fi

# Google Analytics setup
echo -e "${BLUE}📈 Google Analytics 4 Configuration${NC}"
echo ""
echo "To get your Google Analytics Measurement ID:"
echo "1. Go to https://analytics.google.com/"
echo "2. Create a new property for 'ccproxy.orchestre.dev'"
echo "3. Copy your Measurement ID (format: G-XXXXXXXXXX)"
echo ""

read -p "Enter your Google Analytics Measurement ID: " ga_id

# Validate GA ID format
if [[ ! $ga_id =~ ^G-[A-Z0-9]{8,}$ ]]; then
    echo -e "${RED}❌ Invalid Google Analytics ID format.${NC}"
    echo "   Expected format: G-XXXXXXXXXX"
    echo "   Example: G-ABC123DEF4"
    exit 1
fi

# Create .env file
echo "# Google Analytics Configuration for CCProxy Documentation" > "$ENV_FILE"
echo "# Generated on $(date)" >> "$ENV_FILE"
echo "" >> "$ENV_FILE"
echo "# Google Analytics 4 Measurement ID" >> "$ENV_FILE"
echo "GA_MEASUREMENT_ID=$ga_id" >> "$ENV_FILE"

echo ""
echo -e "${GREEN}✅ Google Analytics configured successfully!${NC}"
echo ""
echo -e "${BLUE}📁 Configuration saved to: $ENV_FILE${NC}"
echo ""

# Show current configuration
echo -e "${BLUE}📋 Current configuration:${NC}"
echo "---"
echo "GA_MEASUREMENT_ID=$ga_id"
echo "---"
echo ""

# Test build
echo -e "${BLUE}🔨 Testing build with Google Analytics...${NC}"
cd docs

if npm run build > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Build successful with Google Analytics configured!${NC}"
else
    echo -e "${RED}❌ Build failed. Please check your configuration.${NC}"
    exit 1
fi

cd ..

echo ""
echo -e "${BLUE}🚀 Deployment Instructions:${NC}"
echo ""
echo "For Cloudflare Pages:"
echo "  1. Go to Cloudflare Pages → Your project → Settings → Environment variables"
echo "  2. Add: GA_MEASUREMENT_ID = $ga_id"
echo "  3. Deploy your site"
echo ""
echo "For GitHub Actions:"
echo "  1. Go to GitHub → Settings → Secrets → Actions"
echo "  2. Add secret: GA_MEASUREMENT_ID = $ga_id"
echo "  3. Push to main branch"
echo ""
echo "Local deployment:"
echo "  ./scripts/deploy-local.sh"
echo "  ./scripts/deploy-cloudflare.sh --production"
echo ""

echo -e "${BLUE}📊 Verification:${NC}"
echo ""
echo "After deployment:"
echo "  1. Visit: https://ccproxy.orchestre.dev"
echo "  2. Check Google Analytics Realtime reports"
echo "  3. Verify tracking in browser dev tools"
echo ""

echo -e "${BLUE}📈 Google Analytics Dashboard:${NC}"
echo "  https://analytics.google.com/analytics/web/"
echo ""

echo -e "${GREEN}🎉 Google Analytics setup complete!${NC}"
echo ""
echo "Your CCProxy documentation will now track:"
echo "  • Page views and user engagement"
echo "  • Traffic sources and referrals"
echo "  • Geographic and device data"
echo "  • Documentation performance metrics"
echo ""
echo "Note: Analytics data may take 24-48 hours to fully populate."