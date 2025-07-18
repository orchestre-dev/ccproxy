---
title: Status Endpoint - CCProxy API Reference
description: Monitor CCProxy health and provider status with the /status endpoint. Real-time information about proxy health and provider connectivity.
keywords: CCProxy status, API health check, proxy monitoring, provider status
---

# Status Endpoint

<SocialShare />

Monitor CCProxy and provider health with the status endpoint.

## GET /status

Returns detailed status information about CCProxy and the configured provider.

### Request

```bash
curl http://localhost:7187/status
```

### Response

```json
{
  "status": "healthy",
  "timestamp": "2025-01-17T10:30:45Z",
  "proxy": {
    "version": "1.0.0",
    "uptime": "2h 15m 30s",
    "requests_served": 1247
  },
  "provider": {
    "name": "groq",
    "status": "connected",
    "model": "moonshotai/kimi-k2-instruct",
    "last_check": "2025-01-17T10:30:44Z",
    "response_time_ms": 185
  }
}
```

### Status Values

| Status | Description |
|--------|-------------|
| `healthy` | All systems operational |
| `degraded` | Provider issues, but proxy operational |
| `unhealthy` | Proxy or provider connection failed |

## Provider-Specific Status

### Groq Status
```json
{
  "provider": {
    "name": "groq",
    "status": "connected",
    "model": "moonshotai/kimi-k2-instruct",
    "tokens_per_second": 185,
    "rate_limit_remaining": 5000
  }
}
```

### OpenAI Status
```json
{
  "provider": {
    "name": "openai",
    "status": "connected", 
    "model": "gpt-4o",
    "rate_limit_remaining": 150,
    "organization": "org-..."
  }
}
```

## Monitoring Integration

### Prometheus Metrics
```bash
# Enable metrics endpoint
export ENABLE_METRICS=true

# Scrape metrics
curl http://localhost:7187/metrics
```

### Health Check Script
```bash
#!/bin/bash
STATUS=$(curl -s http://localhost:7187/status | jq -r '.status')
if [ "$STATUS" != "healthy" ]; then
  echo "CCProxy unhealthy: $STATUS"
  exit 1
fi
echo "CCProxy healthy"
```

## Error Responses

### Provider Connection Failed
```json
{
  "status": "unhealthy",
  "error": "provider_connection_failed",
  "message": "Failed to connect to Groq API",
  "details": {
    "provider": "groq",
    "error_code": "authentication_failed"
  }
}
```

### Rate Limited
```json
{
  "status": "degraded", 
  "warning": "rate_limited",
  "message": "Provider rate limit exceeded",
  "retry_after": 60
}
```

## See Also

- [Health Endpoints](/api/health) - Simple health checks
- [Messages Endpoint](/api/messages) - Main API endpoint
- [Error Handling](/api/errors) - Error response format