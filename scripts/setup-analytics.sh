#!/bin/bash

# CCProxy Analytics Setup Script
# Quick setup for analytics providers

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}📊 CCProxy Analytics Setup${NC}"
echo "============================="
echo ""

# Check if we're in the right directory
if [ ! -f "docs/package.json" ]; then
    echo -e "${RED}❌ Please run this script from the project root directory${NC}"
    exit 1
fi

ENV_FILE="docs/.env"

echo -e "${BLUE}🔧 Setting up analytics configuration...${NC}"
echo ""

# Check if .env already exists
if [ -f "$ENV_FILE" ]; then
    echo -e "${YELLOW}⚠️  .env file already exists. Creating backup...${NC}"
    cp "$ENV_FILE" "$ENV_FILE.backup"
    echo -e "${GREEN}✅ Backup created: $ENV_FILE.backup${NC}"
    echo ""
fi

# Create or update .env file
echo "# Analytics Configuration for CCProxy Documentation" > "$ENV_FILE"
echo "# Generated on $(date)" >> "$ENV_FILE"
echo "" >> "$ENV_FILE"

# Plausible setup
echo -e "${BLUE}📈 Setting up Plausible Analytics (Recommended)${NC}"
echo ""
echo "Plausible is privacy-focused, GDPR-compliant, and perfect for documentation sites."
echo "Benefits: No cookies, lightweight (<1KB), open source"
echo ""
read -p "Do you want to set up Plausible Analytics? (y/n): " setup_plausible

if [[ $setup_plausible == "y" || $setup_plausible == "Y" ]]; then
    echo ""
    echo "Choose Plausible setup type:"
    echo "1) Plausible Cloud (plausible.io) - Easiest setup"
    echo "2) Self-hosted Plausible - Free, complete control"
    echo ""
    read -p "Enter choice (1 or 2): " plausible_choice
    
    echo "" >> "$ENV_FILE"
    echo "# Plausible Analytics (Privacy-focused, GDPR compliant)" >> "$ENV_FILE"
    echo "PLAUSIBLE_DOMAIN=ccproxy.orchestre.dev" >> "$ENV_FILE"
    
    if [[ $plausible_choice == "2" ]]; then
        echo ""
        read -p "Enter your self-hosted Plausible domain (e.g., analytics.yoursite.com): " plausible_host
        echo "PLAUSIBLE_API_HOST=https://$plausible_host/api/event" >> "$ENV_FILE"
        echo "PLAUSIBLE_SRC=https://$plausible_host/js/script.js" >> "$ENV_FILE"
        echo ""
        echo -e "${GREEN}✅ Self-hosted Plausible configured${NC}"
        echo -e "${BLUE}📝 Next steps:${NC}"
        echo "   1. Deploy Plausible to $plausible_host"
        echo "   2. Add ccproxy.orchestre.dev as a site"
        echo "   3. Deploy your documentation"
    else
        echo ""
        echo -e "${GREEN}✅ Plausible Cloud configured${NC}"
        echo -e "${BLUE}📝 Next steps:${NC}"
        echo "   1. Create account at https://plausible.io"
        echo "   2. Add site: ccproxy.orchestre.dev"
        echo "   3. Deploy your documentation"
    fi
fi

echo "" >> "$ENV_FILE"

# Google Analytics setup
echo ""
echo -e "${BLUE}📊 Setting up Google Analytics (Optional)${NC}"
echo ""
echo "Google Analytics provides detailed insights but uses cookies and requires GDPR compliance."
echo ""
read -p "Do you want to set up Google Analytics? (y/n): " setup_ga

if [[ $setup_ga == "y" || $setup_ga == "Y" ]]; then
    echo ""
    read -p "Enter your Google Analytics Measurement ID (format G-XXXXXXXXXX): " ga_id
    
    # Validate GA ID format
    if [[ $ga_id =~ ^G-[A-Z0-9]{8,}$ ]]; then
        echo "" >> "$ENV_FILE"
        echo "# Google Analytics (Optional - requires cookie consent for GDPR)" >> "$ENV_FILE"
        echo "GA_MEASUREMENT_ID=$ga_id" >> "$ENV_FILE"
        echo ""
        echo -e "${GREEN}✅ Google Analytics configured${NC}"
        echo -e "${BLUE}📝 Next steps:${NC}"
        echo "   1. Verify GA4 property is set up for ccproxy.orchestre.dev"
        echo "   2. Consider cookie consent for GDPR compliance"
        echo "   3. Deploy your documentation"
    else
        echo -e "${RED}❌ Invalid Google Analytics ID format. Expected: G-XXXXXXXXXX${NC}"
        echo "   Skipping Google Analytics setup."
    fi
fi

echo ""
echo -e "${GREEN}✅ Analytics configuration complete!${NC}"
echo ""
echo -e "${BLUE}📁 Configuration saved to: $ENV_FILE${NC}"
echo ""

# Show current configuration
echo -e "${BLUE}📋 Current configuration:${NC}"
echo "---"
cat "$ENV_FILE" | grep -v '^#' | grep -v '^$'
echo "---"
echo ""

# Test build
echo -e "${BLUE}🔨 Testing build with analytics...${NC}"
cd docs

if npm run build > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Build successful with analytics configured!${NC}"
else
    echo -e "${RED}❌ Build failed. Please check your configuration.${NC}"
    exit 1
fi

cd ..

echo ""
echo -e "${BLUE}🚀 Ready to deploy!${NC}"
echo ""
echo "Deployment options:"
echo "  Local test:    ./scripts/deploy-local.sh"
echo "  Preview:       ./scripts/deploy-cloudflare.sh --preview"
echo "  Production:    ./scripts/deploy-cloudflare.sh --production"
echo ""
echo "Analytics dashboards:"
if [[ $setup_plausible == "y" || $setup_plausible == "Y" ]]; then
    echo "  Plausible:     https://plausible.io/ccproxy.orchestre.dev"
fi
if [[ $setup_ga == "y" || $setup_ga == "Y" ]]; then
    echo "  Google Analytics: https://analytics.google.com"
fi
echo ""
echo -e "${GREEN}🎉 Setup complete!${NC}"