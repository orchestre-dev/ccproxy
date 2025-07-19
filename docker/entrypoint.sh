#!/bin/sh
# Docker entrypoint script for ccproxy

set -e

# Function to generate config from environment variables
generate_config() {
    echo "Generating configuration from environment variables..."
    
    # Start with base config
    cat > /home/ccproxy/.ccproxy/config.json <<EOF
{
    "host": "${CCPROXY_HOST:-0.0.0.0}",
    "port": ${CCPROXY_PORT:-3456},
    "log": ${CCPROXY_LOG:-true},
    "log_file": "${CCPROXY_LOG_FILE:-}",
    "apikey": "${CCPROXY_API_KEY:-}",
    "proxy_url": "${CCPROXY_PROXY_URL:-}",
    "providers": [],
    "routes": {}
}
EOF

    # Add providers from environment
    if [ -n "$CCPROXY_PROVIDERS_JSON" ]; then
        # If full providers JSON is provided
        echo "Using provided providers configuration..."
        jq ".providers = $CCPROXY_PROVIDERS_JSON" /home/ccproxy/.ccproxy/config.json > /tmp/config.json && \
        mv /tmp/config.json /home/ccproxy/.ccproxy/config.json
    fi
    
    # Add routes from environment
    if [ -n "$CCPROXY_ROUTES_JSON" ]; then
        # If full routes JSON is provided
        echo "Using provided routes configuration..."
        jq ".routes = $CCPROXY_ROUTES_JSON" /home/ccproxy/.ccproxy/config.json > /tmp/config.json && \
        mv /tmp/config.json /home/ccproxy/.ccproxy/config.json
    fi
}

# Check if config file exists
if [ ! -f "/home/ccproxy/.ccproxy/config.json" ]; then
    # No config file, check if we should generate from env
    if [ "$CCPROXY_GENERATE_CONFIG" = "true" ]; then
        generate_config
    elif [ -f "/home/ccproxy/.ccproxy/config.example.json" ]; then
        # Use example config as default
        cp /home/ccproxy/.ccproxy/config.example.json /home/ccproxy/.ccproxy/config.json
    fi
fi

# Handle special commands
case "$1" in
    "generate-config")
        generate_config
        exit 0
        ;;
    "validate-config")
        ccproxy validate --config /home/ccproxy/.ccproxy/config.json
        exit $?
        ;;
esac

# Execute ccproxy with arguments
exec ccproxy "$@"