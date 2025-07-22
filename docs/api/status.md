---
title: Status Endpoint - CCProxy API Reference
description: Monitor CCProxy health and provider status with the /status endpoint. Real-time information about proxy health and provider connectivity.
keywords: CCProxy status, API health check, proxy monitoring, provider status
---

# Status Endpoint

<SocialShare />

Get detailed information about CCProxy service status and configuration.

## GET /status

Returns service status, version information, and configuration details.

### Authentication

This endpoint requires authentication:
- Request from localhost (127.0.0.1)
- Valid API key via `Authorization: Bearer <key>` or `x-api-key: <key>` header

### Request

```bash
# From localhost
curl http://localhost:3456/status

# With API key
curl -H "x-api-key: your-api-key" http://your-server:3456/status
```

### Response

```json
{
  "service": "CCProxy",
  "version": "1.0.0",
  "status": "running",
  "uptime": "2h 15m 30s",
  "pid": 12345,
  "host": "127.0.0.1",
  "port": 3456,
  "api_key_configured": true,
  "providers": {
    "count": 2,
    "active": ["anthropic", "openai"]
  },
  "build": {
    "version": "1.0.0",
    "commit": "abc123def",
    "date": "2025-01-17T08:00:00Z"
  }
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `service` | string | Service name (always "CCProxy") |
| `version` | string | Service version |
| `status` | string | Service status (running/stopped) |
| `uptime` | string | Human-readable uptime |
| `pid` | number | Process ID |
| `host` | string | Bind host address |
| `port` | number | Bind port number |
| `api_key_configured` | boolean | Whether API key is configured |
| `providers` | object | Provider configuration info |
| `build` | object | Build information |

## Usage Examples

### Basic Status Check
```bash
#!/bin/bash
# Check if CCProxy is running
STATUS=$(curl -s http://localhost:3456/status | jq -r '.status')
if [ "$STATUS" = "running" ]; then
  echo "✅ CCProxy is running"
else
  echo "❌ CCProxy is not running"
  exit 1
fi
```

### Provider Information
```bash
# Get active providers
PROVIDERS=$(curl -s -H "x-api-key: $API_KEY" http://localhost:3456/status | 
  jq -r '.providers.active[]')
echo "Active providers: $PROVIDERS"
```

### Version Check
```bash
# Check CCProxy version
VERSION=$(curl -s http://localhost:3456/status | jq -r '.version')
echo "CCProxy version: $VERSION"
```

## Integration with Monitoring Systems

### Uptime Monitoring
```bash
#!/bin/bash
# Monitor service uptime
RESPONSE=$(curl -s -H "x-api-key: $API_KEY" http://localhost:3456/status)
if [ $? -ne 0 ]; then
  alert "CCProxy is down"
  exit 1
fi

UPTIME=$(echo "$RESPONSE" | jq -r '.uptime')
echo "CCProxy uptime: $UPTIME"
```

### Configuration Validation
```bash
#!/bin/bash
# Verify expected configuration
CONFIG=$(curl -s -H "x-api-key: $API_KEY" http://localhost:3456/status)

# Check if API key is configured
API_KEY_SET=$(echo "$CONFIG" | jq -r '.api_key_configured')
if [ "$API_KEY_SET" != "true" ]; then
  echo "WARNING: API key not configured"
fi

# Check provider count
PROVIDER_COUNT=$(echo "$CONFIG" | jq -r '.providers.count')
if [ "$PROVIDER_COUNT" -eq 0 ]; then
  echo "ERROR: No providers configured"
  exit 1
fi
```

## Error Responses

### Unauthorized Access
```json
{
  "error": {
    "type": "authentication_error",
    "message": "Invalid or missing API key"
  }
}
```
HTTP Status: 401

### Service Not Available
```json
{
  "error": {
    "type": "service_error",
    "message": "Service temporarily unavailable"
  }
}
```
HTTP Status: 503

## Docker Integration

```yaml
version: '3.8'
services:
  ccproxy:
    image: ccproxy:latest
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3456/status"]
      interval: 30s
      timeout: 10s
      retries: 3
```

## See Also

- [Health Endpoints](/api/health) - Simple health checks
- [Messages Endpoint](/api/messages) - Main API endpoint
- [Error Handling](/api/errors) - Error response format