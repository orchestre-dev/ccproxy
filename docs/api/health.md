# Health Endpoints

CCProxy provides several endpoints for monitoring the health and status of the service and underlying providers.

## Endpoints Overview

| Endpoint | Method | Purpose | Authentication |
|----------|--------|---------|----------------|
| `/` | GET | Basic API info | None |
| `/health` | GET | Service health status | API key or localhost |
| `/status` | GET | Service information | API key or localhost |

## Basic API Info

### Endpoint
```
GET /
```

### Description
Returns basic API information including version.

### Response

```json
{
  "message": "LLMs API",
  "version": "1.0.0"
}
```

### Example

```bash
curl http://localhost:3456/
```

### Use Cases
- Quick service availability checks
- Version verification
- Basic connectivity test

## Health Check

### Endpoint
```
GET /health
```

### Description
Returns the current health status of the service. Requires authentication (API key or localhost access).

### Authentication
This endpoint requires one of the following:
- Request from localhost (127.0.0.1)
- Valid API key via `Authorization: Bearer <key>` or `x-api-key: <key>` header

### Response

#### Without Authentication (or basic status)
```json
{
  "status": "healthy"
}
```

#### With Authentication (detailed status)
```json
{
  "status": "healthy",
  "timestamp": "2025-01-17T10:30:00Z",
  "uptime": 7945000000000,
  "version": "1.0.0",
  "config": {
    "host": "127.0.0.1",
    "port": 3456,
    "providers": 2,
    "api_key_configured": true
  },
  "performance": {
    "memory_mb": 15.2,
    "goroutines": 23,
    "requests": {
      "total": 1250,
      "success": 1198,
      "failed": 52
    },
    "latency": {
      "avg_ms": 245,
      "p50_ms": 200,
      "p95_ms": 450,
      "p99_ms": 890
    }
  }
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | Always "healthy" if service is running |
| `timestamp` | string | Current time (ISO format) |
| `uptime` | number | Uptime in nanoseconds |
| `version` | string | CCProxy version |
| `config` | object | Configuration summary (authenticated only) |
| `performance` | object | Performance metrics (authenticated only) |

### Examples

#### Unauthenticated Request
```bash
curl http://your-server:3456/health
# Returns: {"status":"healthy"}
```

#### Authenticated Request from Localhost
```bash
curl http://localhost:3456/health
# Returns detailed health information
```

#### Authenticated Request with API Key
```bash
curl -H "x-api-key: your-api-key" http://your-server:3456/health
# Returns detailed health information
```

## Status Endpoint

### Endpoint
```
GET /status
```

### Description
Provides information about the service status and configuration. Requires authentication (API key or localhost access).

### Authentication
This endpoint requires one of the following:
- Request from localhost (127.0.0.1)
- Valid API key via `Authorization: Bearer <key>` or `x-api-key: <key>` header

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
| `service` | string | Service name |
| `version` | string | Service version |
| `status` | string | Service status (running/stopped) |
| `uptime` | string | Human-readable uptime |
| `pid` | number | Process ID |
| `host` | string | Bind host |
| `port` | number | Bind port |
| `api_key_configured` | boolean | Whether API key is set |
| `providers` | object | Provider information |
| `build` | object | Build information |

### Examples

#### Request from Localhost
```bash
curl http://localhost:3456/status
```

#### Request with API Key
```bash
curl -H "x-api-key: your-api-key" http://your-server:3456/status
```

### Provider-Specific Information

CCProxy supports the following providers with their specific capabilities:

#### Anthropic Provider
- Models: claude-3-opus, claude-3-sonnet, claude-3-haiku
- Full API compatibility with native format
- Supports all Claude features including vision and tools

#### OpenAI Provider  
- Models: gpt-4, gpt-4-turbo, gpt-3.5-turbo
- Automatic API translation from Anthropic format
- Full support for function calling and streaming

#### Google Gemini Provider
- Models: gemini-1.5-flash, gemini-1.5-pro
- Multimodal support with vision capabilities
- API translation for tool use

#### DeepSeek Provider
- Models: deepseek-coder, deepseek-chat
- Optimized for code generation tasks
- Full streaming support

#### OpenRouter Provider
- Access to 100+ models through unified API
- Model-specific routing and optimization
- Pay-per-use pricing model

## Monitoring Integration

### Health Check Integration

CCProxy provides health endpoints that can be integrated with various monitoring systems. The `/health` endpoint provides detailed metrics when authenticated, including request counts, latency percentiles, and resource usage.

### Health Check Scripts

#### Simple Health Check
```bash
#!/bin/bash
# health-check.sh

response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3456/)

if [ $response -eq 200 ]; then
    echo "CCProxy is healthy"
    exit 0
else
    echo "CCProxy is unhealthy (HTTP $response)"
    exit 1
fi
```

#### Detailed Health Check
```bash
#!/bin/bash
# detailed-health-check.sh

health=$(curl -s http://localhost:3456/health | jq -r '.status')

case $health in
    "healthy")
        echo "✅ CCProxy is healthy"
        exit 0
        ;;
    "degraded") 
        echo "⚠️  CCProxy is degraded"
        exit 1
        ;;
    "unhealthy")
        echo "❌ CCProxy is unhealthy"
        exit 2
        ;;
    *)
        echo "❓ CCProxy status unknown"
        exit 3
        ;;
esac
```

### Docker Health Checks

#### Dockerfile
```dockerfile
FROM golang:1.23-alpine AS builder
# ... build steps ...

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/
COPY --from=builder /app/ccproxy .

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:3456/ || exit 1

EXPOSE 3456
CMD ["./ccproxy", "start", "--foreground"]
```

#### Docker Compose
```yaml
version: '3.8'
services:
  ccproxy:
    build: .
    ports:
      - "3456:3456"
    volumes:
      - ./config.json:/home/ccproxy/.ccproxy/config.json
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3456/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

### Kubernetes Probes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ccproxy
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ccproxy
  template:
    metadata:
      labels:
        app: ccproxy
    spec:
      containers:
      - name: ccproxy
        image: ccproxy:latest
        ports:
        - containerPort: 3456
        volumeMounts:
        - name: config
          mountPath: /home/ccproxy/.ccproxy/config.json
          subPath: config.json
        livenessProbe:
          httpGet:
            path: /
            port: 3456
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health
            port: 3456
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 10
          successThreshold: 1
          failureThreshold: 3
      volumes:
      - name: config
        configMap:
          name: ccproxy-config
```

## Alerting

### Health-Based Monitoring

Set up alerts based on the health endpoint responses:

#### Service Availability
Monitor the basic endpoint for service availability:
```bash
# Alert if service returns non-200 status
curl -f http://localhost:3456/ || alert "CCProxy is down"
```

#### Performance Monitoring
Use the authenticated health endpoint to monitor performance metrics:
```bash
# Check error rate from health endpoint
ERROR_RATE=$(curl -s -H "x-api-key: $API_KEY" http://localhost:3456/health | \
  jq -r '(.performance.requests.failed / .performance.requests.total) * 100')

if (( $(echo "$ERROR_RATE > 10" | bc -l) )); then
  alert "High error rate: ${ERROR_RATE}%"
fi
```

#### Latency Monitoring
```bash
# Check P95 latency
P95_LATENCY=$(curl -s -H "x-api-key: $API_KEY" http://localhost:3456/health | \
  jq -r '.performance.latency.p95_ms')

if [ "$P95_LATENCY" -gt 1000 ]; then
  alert "High P95 latency: ${P95_LATENCY}ms"
fi
```

### PagerDuty Integration

```bash
#!/bin/bash
# pagerduty-alert.sh

HEALTH_STATUS=$(curl -s http://localhost:3456/health | jq -r '.status')

if [ "$HEALTH_STATUS" != "healthy" ]; then
    curl -X POST https://events.pagerduty.com/v2/enqueue \
      -H 'Content-Type: application/json' \
      -d '{
        "routing_key": "'$PAGERDUTY_ROUTING_KEY'",
        "event_action": "trigger",
        "payload": {
          "summary": "CCProxy is '$HEALTH_STATUS'",
          "source": "ccproxy-health-check",
          "severity": "error"
        }
      }'
fi
```

## Best Practices

### 1. Health Check Frequency

- **Load balancers**: Check `/` every 5-10 seconds
- **Monitoring systems**: Check `/health` every 30-60 seconds  
- **Alerting**: Poll authenticated endpoints every 1-5 minutes

### 2. Timeout Configuration

- **Basic health check** (`/`): 1-2 second timeout
- **Detailed health check** (`/health`): 5-10 second timeout
- **Status check** (`/status`): 5-10 second timeout

### 3. Authentication for Monitoring

- Use API keys for production monitoring
- Restrict health endpoint access to monitoring systems
- Rotate API keys regularly

### 4. Endpoint Usage Guidelines

- Use `/` for simple uptime checks
- Use `/health` with auth for detailed metrics
- Use `/status` with auth for configuration info

## Troubleshooting

### Common Issues

#### Health Check Timeouts
- Increase timeout values in monitoring config
- Check network connectivity
- Verify CCProxy resource usage

#### Authentication Failures
- Verify API key is correctly configured
- Check request headers format
- Ensure monitoring from allowed IPs

#### Incomplete Metrics
- Confirm authentication is working
- Check if requesting from localhost
- Verify API key permissions

## Next Steps

- Set up [monitoring dashboard](/guide/monitoring) with Grafana
- Configure [alerting rules](/guide/monitoring) for your environment
- Learn about [performance optimization](/guide/monitoring)