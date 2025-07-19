#!/bin/bash

echo "=== CCProxy Safe Start Script ==="
echo "This script runs ccproxy with resource limits to prevent system freeze"
echo ""

# Kill any existing ccproxy processes
echo "Cleaning up existing processes..."
pkill -f ccproxy 2>/dev/null
rm -f ~/.ccproxy/.ccproxy.pid
sleep 1

# Build the binary
echo "Building ccproxy..."
go build -o ccproxy ./cmd/ccproxy
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo ""
echo "Starting ccproxy with resource limits:"
echo "- CPU limit: 50%"
echo "- Memory limit: 512MB"
echo "- Timeout: 30 seconds"
echo ""

# On macOS, we can use different tools
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS doesn't have cgroups, but we can use nice and ulimit
    echo "Running on macOS with nice priority and ulimits..."
    
    # Set memory limit (512MB in KB)
    ulimit -v 524288
    
    # Run with lower priority and timeout
    nice -n 10 timeout 30s ./ccproxy start --config test-minimal.json --foreground
else
    # On Linux, we can use cgroups or systemd-run
    if command -v systemd-run &> /dev/null; then
        echo "Running with systemd-run resource limits..."
        systemd-run --scope -p CPUQuota=50% -p MemoryLimit=512M timeout 30s ./ccproxy start --config test-minimal.json --foreground
    else
        echo "Running with basic resource limits..."
        # Set memory limit
        ulimit -v 524288
        # Run with timeout
        timeout 30s ./ccproxy start --config test-minimal.json --foreground
    fi
fi

exit_code=$?
echo ""
echo "Process exited with code: $exit_code"

if [ $exit_code -eq 124 ]; then
    echo "WARNING: Process was killed by timeout (possible freeze)"
fi