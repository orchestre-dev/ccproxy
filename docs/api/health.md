# Health Endpoints

CCProxy provides several endpoints for monitoring the health and status of the service and underlying providers.

## Endpoints Overview

| Endpoint | Method | Purpose | Response Time |
|----------|--------|---------|---------------|
| `/` | GET | Basic health check | < 1ms |
| `/health` | GET | Detailed health status | < 100ms |
| `/status` | GET | Provider status & config | < 100ms |

## Basic Health Check

### Endpoint
```
GET /
```

### Description
Simple health check that returns immediately if the service is running.

### Response

```json
{
  "status": "ok",
  "message": "CCProxy is running"
}
```

### Example

```bash
curl http://localhost:7187/
```

### Use Cases
- Load balancer health checks
- Container orchestration probes
- Quick service availability checks

## Detailed Health Check

### Endpoint
```
GET /health
```

### Description
Comprehensive health check that validates the service configuration and provider connectivity.

### Response

```json
{
  "status": "healthy",
  "timestamp": "2025-01-17T10:30:00.000Z",
  "version": "1.0.0",
  "uptime": "2h 15m 30s",
  "provider": {
    "name": "groq",
    "status": "healthy",
    "model": "moonshotai/kimi-k2-instruct",
    "last_check": "2025-01-17T10:29:45.000Z",
    "response_time": "250ms"
  },
  "system": {
    "memory_usage": "45.2%",
    "cpu_usage": "12.8%",
    "goroutines": 23
  },
  "requests": {
    "total": 1250,
    "successful": 1198,
    "failed": 52,
    "success_rate": "95.8%"
  }
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | Overall health status |
| `timestamp` | string | Current timestamp (ISO 8601) |
| `version` | string | CCProxy version |
| `uptime` | string | Service uptime |
| `provider` | object | Provider-specific health info |
| `system` | object | System resource usage |
| `requests` | object | Request statistics |

### Health Status Values

| Status | Description |
|--------|-------------|
| `healthy` | All systems operational |
| `degraded` | Service running but with issues |
| `unhealthy` | Service not functioning properly |

### Example

```bash
curl http://localhost:7187/health
```

### Error Response

```json
{
  "status": "unhealthy",
  "timestamp": "2025-01-17T10:30:00.000Z",
  "version": "1.0.0",
  "uptime": "2h 15m 30s",
  "provider": {
    "name": "groq",
    "status": "unhealthy",
    "error": "API key invalid",
    "last_check": "2025-01-17T10:29:45.000Z"
  },
  "system": {
    "memory_usage": "45.2%",
    "cpu_usage": "12.8%",
    "goroutines": 23
  },
  "requests": {
    "total": 1250,
    "successful": 1100,
    "failed": 150,
    "success_rate": "88.0%"
  }
}
```

## Status Endpoint

### Endpoint
```
GET /status
```

### Description
Provides detailed information about the current provider configuration and operational status.

### Response

```json
{
  "service": {
    "name": "CCProxy",
    "version": "1.0.0",
    "build": "abc123",
    "start_time": "2025-01-17T08:15:00.000Z",
    "uptime": "2h 15m 30s"
  },
  "provider": {
    "name": "groq",
    "model": "moonshotai/kimi-k2-instruct",
    "base_url": "https://api.groq.com/openai/v1",
    "max_tokens": 16384,
    "status": "active",
    "capabilities": [
      "text_generation",
      "function_calling",
      "streaming"
    ]
  },
  "configuration": {
    "port": 7187,
    "log_level": "info",
    "timeout": "120s",
    "cors_enabled": true
  },
  "metrics": {
    "requests_total": 1250,
    "requests_successful": 1198,
    "requests_failed": 52,
    "avg_response_time": "850ms",
    "tokens_processed": 2485000,
    "uptime_percentage": "99.2%"
  }
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `service` | object | Service information |
| `provider` | object | Current provider configuration |
| `configuration` | object | Runtime configuration |
| `metrics` | object | Operational metrics |

### Example

```bash
curl http://localhost:7187/status
```

### Provider-Specific Status

Different providers return different capability information:

#### Groq Provider
```json
{
  "provider": {
    "name": "groq",
    "model": "moonshotai/kimi-k2-instruct",
    "capabilities": ["text_generation", "function_calling", "streaming"],
    "rate_limits": {
      "requests_per_minute": 30,
      "tokens_per_minute": 6000
    }
  }
}
```

#### OpenAI Provider
```json
{
  "provider": {
    "name": "openai",
    "model": "gpt-4o",
    "capabilities": ["text_generation", "function_calling", "vision", "streaming"],
    "rate_limits": {
      "requests_per_minute": 10000,
      "tokens_per_minute": 30000000
    }
  }
}
```

#### Ollama Provider
```json
{
  "provider": {
    "name": "ollama", 
    "model": "llama3.2",
    "base_url": "http://localhost:11434",
    "capabilities": ["text_generation", "function_calling", "local_processing"],
    "rate_limits": {
      "requests_per_minute": "unlimited",
      "tokens_per_minute": "unlimited"
    }
  }
}
```

## Monitoring Integration

### Prometheus Metrics

CCProxy exposes metrics in Prometheus format at `/metrics`:

```
# HELP ccproxy_requests_total Total number of requests
# TYPE ccproxy_requests_total counter
ccproxy_requests_total{provider="groq",status="success"} 1198
ccproxy_requests_total{provider="groq",status="error"} 52

# HELP ccproxy_request_duration_seconds Request duration
# TYPE ccproxy_request_duration_seconds histogram
ccproxy_request_duration_seconds_bucket{provider="groq",le="0.1"} 245
ccproxy_request_duration_seconds_bucket{provider="groq",le="0.5"} 856
ccproxy_request_duration_seconds_bucket{provider="groq",le="1.0"} 1150
ccproxy_request_duration_seconds_bucket{provider="groq",le="+Inf"} 1250

# HELP ccproxy_tokens_processed_total Total tokens processed
# TYPE ccproxy_tokens_processed_total counter
ccproxy_tokens_processed_total{provider="groq",type="input"} 1245000
ccproxy_tokens_processed_total{provider="groq",type="output"} 1240000
```

### Health Check Scripts

#### Simple Health Check
```bash
#!/bin/bash
# health-check.sh

response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:7187/)

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

health=$(curl -s http://localhost:7187/health | jq -r '.status')

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
FROM golang:1.21-alpine AS builder
# ... build steps ...

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/
COPY --from=builder /app/ccproxy .

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:7187/health || exit 1

EXPOSE 7187
CMD ["./ccproxy"]
```

#### Docker Compose
```yaml
version: '3.8'
services:
  ccproxy:
    build: .
    ports:
      - "7187:7187"
    environment:
      - PROVIDER=groq
      - GROQ_API_KEY=${GROQ_API_KEY}
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:7187/health"]
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
        - containerPort: 7187
        env:
        - name: PROVIDER
          value: "groq"
        - name: GROQ_API_KEY
          valueFrom:
            secretKeyRef:
              name: ccproxy-secrets
              key: groq-api-key
        livenessProbe:
          httpGet:
            path: /
            port: 7187
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health
            port: 7187
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 10
          successThreshold: 1
          failureThreshold: 3
```

## Alerting

### Grafana Alerts

Example Grafana alert rules:

```yaml
# High error rate
- alert: CCProxy High Error Rate
  expr: rate(ccproxy_requests_total{status="error"}[5m]) / rate(ccproxy_requests_total[5m]) > 0.1
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "CCProxy error rate is high"
    description: "Error rate is {{ $value | humanizePercentage }} for the last 5 minutes"

# Service down
- alert: CCProxy Down
  expr: up{job="ccproxy"} == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "CCProxy is down"
    description: "CCProxy has been down for more than 1 minute"

# High response time
- alert: CCProxy High Response Time
  expr: histogram_quantile(0.95, rate(ccproxy_request_duration_seconds_bucket[5m])) > 2
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "CCProxy response time is high"
    description: "95th percentile response time is {{ $value }}s"
```

### PagerDuty Integration

```bash
#!/bin/bash
# pagerduty-alert.sh

HEALTH_STATUS=$(curl -s http://localhost:7187/health | jq -r '.status')

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

```bash
# For load balancers: every 5-10 seconds
# For monitoring systems: every 30-60 seconds
# For alerting: every 1-5 minutes
```

### 2. Timeout Configuration

```bash
# Set appropriate timeouts for health checks
# Basic health check: 1-2 seconds
# Detailed health check: 5-10 seconds
```

### 3. Monitoring Stack

```bash
# Recommended monitoring stack:
# - Prometheus for metrics collection
# - Grafana for visualization
# - AlertManager for alerting
# - PagerDuty/Slack for notifications
```

### 4. Health Check Endpoints

```bash
# Use `/` for simple availability checks
# Use `/health` for detailed status monitoring
# Use `/status` for configuration validation
```

## Troubleshooting

### Common Issues

#### Health Check Timeouts
```bash
# Increase timeout values
# Check system resources
# Verify provider connectivity
```

#### False Positives
```bash
# Tune health check sensitivity
# Implement retry logic
# Use appropriate intervals
```

#### Missing Metrics
```bash
# Verify Prometheus endpoint
# Check metric labels
# Validate scrape configuration
```

## Next Steps

- Set up [monitoring dashboard](/guide/monitoring) with Grafana
- Configure [alerting rules](/guide/monitoring) for your environment
- Learn about [performance optimization](/guide/monitoring)