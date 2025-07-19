#!/bin/bash

echo "=== CCProxy Debug Start Script ==="
echo "Time: $(date)"
echo "Working Directory: $(pwd)"
echo ""

# Build the binary
echo "Building ccproxy..."
go build -o ccproxy-debug ./cmd/ccproxy

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Build successful"
echo ""

# Kill any existing ccproxy processes
echo "Checking for existing ccproxy processes..."
pkill -f ccproxy 2>/dev/null
sleep 1

# Clean up PID file
echo "Cleaning up PID file..."
rm -f ~/.ccproxy/.ccproxy.pid

# Set debug environment variables
export GIN_MODE=debug
export CCPROXY_DEBUG=1

echo ""
echo "Starting ccproxy with minimal config..."
echo "Command: ./ccproxy-debug start --config test-minimal.json --foreground"
echo ""
echo "=== Output ==="

# Run with timeout to prevent system freeze
timeout 30s ./ccproxy-debug start --config test-minimal.json --foreground 2>&1 | tee ccproxy-debug.log

exit_code=$?
echo ""
echo "=== Exit code: $exit_code ==="

if [ $exit_code -eq 124 ]; then
    echo "Process timed out after 30 seconds (possible freeze detected)"
fi

echo ""
echo "Debug log saved to: ccproxy-debug.log"