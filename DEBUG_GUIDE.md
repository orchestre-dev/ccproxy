# CCProxy System Freeze Debugging Guide

## Issue Description
CCProxy is causing system freezes when starting up. This guide provides tools and steps to diagnose and fix the issue.

## Debugging Tools Created

### 1. Minimal Test Configuration
**File:** `test-minimal.json`
- Disables all providers
- Enables logging to `/tmp/ccproxy-debug.log`
- Uses default port 3456

### 2. Debug Start Script
**File:** `debug-start.sh`
- Adds debug output at each initialization step
- Uses timeout to prevent indefinite freeze
- Captures output to `ccproxy-debug.log`

### 3. Safe Start Script
**File:** `safe-start.sh`
- Runs ccproxy with resource limits
- CPU limit: 50%
- Memory limit: 512MB
- Auto-terminates after 30 seconds

### 4. Minimal Test Program
**File:** `cmd/test-minimal/main.go`
- Tests basic HTTP server binding
- Helps isolate if the issue is with port binding

## Changes Made to Debug

1. **Disabled Health Checks**
   - File: `internal/server/server.go`
   - Health checks temporarily disabled to isolate the issue

2. **Added Debug Logging**
   - Files: `cmd/ccproxy/commands/start.go`, `internal/server/server.go`
   - Added [DEBUG] output at each initialization step

3. **Added Resource Monitoring**
   - File: `internal/server/server_safe.go`
   - Monitors goroutine count and memory usage

## Troubleshooting Steps

1. **Test Minimal HTTP Server**
   ```bash
   go run cmd/test-minimal/main.go
   ```
   If this freezes, the issue is with basic port binding.

2. **Run with Debug Script**
   ```bash
   ./debug-start.sh
   ```
   Check where the output stops to identify the freeze point.

3. **Run with Safe Limits**
   ```bash
   ./safe-start.sh
   ```
   This prevents system-wide freeze by limiting resources.

4. **Check System Resources**
   ```bash
   # Check if port is already in use
   lsof -i :3456
   
   # Check system load
   top
   
   # Check disk space
   df -h
   ```

5. **Check Logs**
   - Application log: `/tmp/ccproxy-debug.log`
   - System logs: `Console.app` on macOS

## Potential Causes

1. **Port Binding Issues**
   - Port 3456 might be in use
   - Firewall blocking the port

2. **Resource Exhaustion**
   - Too many goroutines spawned
   - Memory leak during initialization

3. **Deadlock**
   - Mutex deadlock in provider initialization
   - Channel deadlock in health checks

4. **External Dependencies**
   - DNS resolution hanging
   - Proxy configuration issues
   - Network timeouts

## Recommended Next Steps

1. Run `./safe-start.sh` first to safely test without freezing your system
2. Check the debug output to see where it stops
3. Run the minimal test program to verify basic functionality
4. Report back with:
   - The last [DEBUG] message before freeze
   - Any error messages
   - System specifications (OS version, available memory, etc.)

## Reverting Changes

To revert debugging changes and re-enable health checks:
```bash
git checkout internal/server/server.go
git checkout cmd/ccproxy/commands/start.go
```