# Health Checks

CCProxy provides comprehensive health check endpoints for monitoring and deployment orchestration.

## Health Check Endpoints

### Basic Health Check

```bash
GET /
```

Returns basic service status:

```json
{
  "message": "CCProxy Multi-Provider Anthropic API is alive 💡",
  "provider": "groq"
}
```

### Detailed Health Check

```bash
GET /health
```

Returns detailed service information:

```json
{
  "status": "healthy",
  "service": "ccproxy",
  "version": "1.0.0"
}
```

### Provider Status

```bash
GET /status
```

Returns provider-specific status:

```json
{
  "provider": "groq",
  "model": "moonshotai/kimi-k2-instruct",
  "base_url": "https://api.groq.com/openai/v1",
  "max_tokens": 16384,
  "status": "active",
  "service": "ccproxy",
  "version": "1.0.0"
}
```

## Docker Health Checks

CCProxy includes built-in Docker health checks:

```dockerfile
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3456/health || exit 1
```

## Kubernetes Health Checks

### Liveness Probe

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 3456
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

### Readiness Probe

```yaml
readinessProbe:
  httpGet:
    path: /health
    port: 3456
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

## Monitoring Integration

### Prometheus Metrics

While CCProxy doesn't expose Prometheus metrics directly, you can monitor health check endpoints:

```yaml
# prometheus.yml
- job_name: 'ccproxy'
  static_configs:
    - targets: ['localhost:3456']
  metrics_path: /health
  scrape_interval: 30s
```

### Alerting

Set up alerts based on health check failures:

```yaml
# alertmanager rules
- alert: CCProxyDown
  expr: up{job="ccproxy"} == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "CCProxy service is down"
    description: "CCProxy has been down for more than 1 minute"
```

## Load Balancer Configuration

### Nginx

```nginx
upstream ccproxy {
    server localhost:3456;
}

server {
    location / {
        proxy_pass http://ccproxy;
        
        # Health check
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    location /health {
        proxy_pass http://ccproxy/health;
        access_log off;
    }
}
```

### Traefik

```yaml
http:
  services:
    ccproxy:
      loadBalancer:
        servers:
          - url: "http://localhost:3456"
        healthCheck:
          path: /health
          interval: 30s
          timeout: 10s
```

## Best Practices

1. **Regular Health Checks**: Monitor health endpoints every 30 seconds
2. **Graceful Degradation**: Use health checks to route traffic away from unhealthy instances
3. **Startup Time**: Allow sufficient time for the service to initialize
4. **Timeout Configuration**: Set appropriate timeouts based on your network conditions
5. **Failure Thresholds**: Configure reasonable failure thresholds to avoid false positives

## Troubleshooting

### Common Issues

1. **Health check timeouts**: Check network connectivity and server load
2. **Provider errors**: Verify API keys and provider configuration
3. **Service startup failures**: Check logs for configuration errors

### Debugging Health Checks

```bash
# Test health check manually
curl -f http://localhost:3456/health

# Check service logs
docker logs ccproxy

# Monitor health check responses
watch -n 5 'curl -s http://localhost:3456/health | jq'
```