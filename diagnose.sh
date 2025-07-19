#!/bin/bash

echo "=== CCProxy Diagnostic Script ==="
echo "This script tests components in isolation to identify the freeze cause"
echo ""

# Clean up
pkill -f ccproxy 2>/dev/null
rm -f ~/.ccproxy/.ccproxy.pid
sleep 1

echo "1. Testing safe mode server (no providers)..."
echo "   Building ccproxy-safe..."
go build -o ccproxy-safe ./cmd/ccproxy-safe
if [ $? -ne 0 ]; then
    echo "   ❌ Build failed"
    exit 1
fi

echo "   Starting safe mode server for 10 seconds..."
timeout 10s ./ccproxy-safe &
SAFE_PID=$!
sleep 2

# Test if it's running
if kill -0 $SAFE_PID 2>/dev/null; then
    echo "   ✅ Safe mode server running successfully"
    curl -s http://localhost:3456/health | jq . || echo "   ❌ Health check failed"
    kill $SAFE_PID 2>/dev/null
    wait $SAFE_PID 2>/dev/null
else
    echo "   ❌ Safe mode server failed to start"
fi

echo ""
echo "2. Testing regular ccproxy with safety checks..."
echo "   Building ccproxy..."
go build -o ccproxy ./cmd/ccproxy
if [ $? -ne 0 ]; then
    echo "   ❌ Build failed"
    exit 1
fi

echo "   Running version command..."
./ccproxy version
echo ""

echo "   Starting with minimal config for 10 seconds..."
timeout 10s ./ccproxy start --config test-minimal.json --foreground 2>&1 | head -20

echo ""
echo "3. Checking system resources..."
echo "   Goroutines before: $(ps -M $(pgrep ccproxy) 2>/dev/null | wc -l || echo "0")"
echo "   Available memory: $(vm_stat | grep 'Pages free' | awk '{print $3 * 4096 / 1048576 " MB"}' 2>/dev/null || free -m | grep available | awk '{print $NF " MB"}')"
echo "   Load average: $(uptime | awk -F'load average:' '{print $2}')"

echo ""
echo "=== Diagnosis Complete ==="
echo ""
echo "If step 1 (safe mode) works but step 2 (regular) freezes,"
echo "the issue is in provider/middleware initialization."
echo ""
echo "If both freeze, the issue is in basic server setup."