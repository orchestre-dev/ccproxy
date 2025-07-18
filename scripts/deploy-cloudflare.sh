#!/bin/bash

# CCProxy Cloudflare Pages Deployment Script
# This script automates deployment to Cloudflare Pages using Wrangler CLI

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="ccproxy"
BUILD_DIR="docs/.vitepress/dist"
BUILD_COMMAND="npm run build"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_requirements() {
    log_info "Checking requirements..."
    
    # Check if Node.js is available
    if ! command -v node &> /dev/null; then
        log_error "Node.js is not installed. Please install Node.js first:"
        echo "https://nodejs.org/"
        exit 1
    fi
    
    # Check if npm is available
    if ! command -v npm &> /dev/null; then
        log_error "npm is not available. Please install npm first."
        exit 1
    fi
    
    # Check if we're in the correct directory
    if [ ! -f "docs/package.json" ]; then
        log_error "This script must be run from the project root directory"
        exit 1
    fi
    
    # Check if logged in to Cloudflare using npx
    log_info "Checking Cloudflare authentication..."
    if ! npx wrangler whoami &> /dev/null; then
        log_error "Not logged in to Cloudflare. Please login first:"
        echo "npx wrangler login"
        exit 1
    fi
    
    log_success "All requirements met"
}

install_dependencies() {
    log_info "Installing dependencies..."
    cd docs
    npm ci
    cd ..
    log_success "Dependencies installed"
}

build_site() {
    log_info "Building VitePress site..."
    cd docs
    
    # Load environment variables if .env exists
    if [ -f ".env" ]; then
        log_info "Loading environment variables from .env"
        export $(cat .env | grep -v '^#' | xargs)
    fi
    
    $BUILD_COMMAND
    cd ..
    
    if [ ! -d "$BUILD_DIR" ]; then
        log_error "Build directory not found: $BUILD_DIR"
        exit 1
    fi
    
    log_success "Site built successfully"
}

deploy_to_pages() {
    log_info "Deploying to Cloudflare Pages..."
    
    # Deploy using npx wrangler
    npx wrangler pages deploy "$BUILD_DIR" --project-name="$PROJECT_NAME"
    
    log_success "Deployment completed!"
    log_info "Your site should be available at: https://ccproxy.orchestre.dev"
}

preview_deployment() {
    log_info "Creating preview deployment..."
    
    # Create a preview deployment
    npx wrangler pages deploy "$BUILD_DIR" --project-name="$PROJECT_NAME" --no-bundle
    
    log_success "Preview deployment completed!"
}

show_usage() {
    echo "Usage: $0 [OPTION]"
    echo ""
    echo "Options:"
    echo "  --production, -p    Deploy to production"
    echo "  --preview, -pr      Create preview deployment"
    echo "  --build-only, -b    Only build the site (no deployment)"
    echo "  --help, -h          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --production     # Deploy to production"
    echo "  $0 --preview        # Create preview deployment"
    echo "  $0 --build-only     # Just build the site"
}

main() {
    case "${1:-}" in
        --production|-p)
            log_info "Starting production deployment..."
            check_requirements
            install_dependencies
            build_site
            deploy_to_pages
            ;;
        --preview|-pr)
            log_info "Starting preview deployment..."
            check_requirements
            install_dependencies
            build_site
            preview_deployment
            ;;
        --build-only|-b)
            log_info "Building site only..."
            install_dependencies
            build_site
            log_success "Build completed. Files are in: $BUILD_DIR"
            ;;
        --help|-h)
            show_usage
            ;;
        "")
            log_info "No option specified. Defaulting to production deployment."
            check_requirements
            install_dependencies
            build_site
            deploy_to_pages
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
}

# Cleanup function
cleanup() {
    log_info "Cleaning up..."
    # Add any cleanup tasks here if needed
}

# Set up cleanup trap
trap cleanup EXIT

# Run main function with all arguments
main "$@"