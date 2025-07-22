# Monitoring

Monitor CCProxy performance, health, and usage with comprehensive monitoring solutions.

<SocialShare />

## Key Metrics to Monitor

### Service Health
- **Service uptime**: Monitor `/health` endpoint
- **Response time**: Track API response latency
- **Error rate**: Monitor failed requests
- **Request volume**: Track requests per second

### Provider Performance
- **Provider latency**: Time taken for provider API calls
- **Provider errors**: Failed provider API calls
- **Token usage**: Monitor token consumption per provider
- **Rate limiting**: Track rate limit hits

### System Resources
- **Memory usage**: Monitor container/process memory
- **CPU usage**: Track CPU utilization
- **Network I/O**: Monitor network bandwidth
- **Disk usage**: Track log file sizes

## Monitoring Stack

### Prometheus + Grafana

#### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'ccproxy'
    static_configs:
      - targets: ['localhost:3456']
    metrics_path: /health
    scrape_interval: 30s
```

#### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "CCProxy Dashboard",
    "panels": [
      {
        "title": "Service Health",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=\"ccproxy\"}",
            "legendFormat": "Service Status"
          }
        ]
      },
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "Requests/sec"
          }
        ]
      }
    ]
  }
}
```

### Log-Based Monitoring

#### ELK Stack Integration

```yaml
# logstash.conf
input {
  file {
    path => "/var/log/ccproxy/app.log"
    codec => "json"
    type => "ccproxy"
  }
}

filter {
  if [type] == "ccproxy" {
    if [action] == "anthropic_request" {
      mutate {
        add_field => { "metric_type" => "request" }
      }
    }
    if [action] == "anthropic_response" {
      mutate {
        add_field => { "metric_type" => "response" }
      }
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "ccproxy-%{+YYYY.MM.dd}"
  }
}
```

#### Kibana Visualizations

1. **Request Volume**: Line chart showing requests over time
2. **Error Rate**: Pie chart showing error distribution
3. **Response Time**: Histogram of API response times
4. **Provider Usage**: Bar chart showing provider utilization

## Health Check Monitoring

### Uptime Monitoring

```bash
#!/bin/bash
# health-check.sh
ENDPOINT="http://localhost:3456/health"
TIMEOUT=10

while true; do
  if curl -f -s --max-time $TIMEOUT $ENDPOINT > /dev/null; then
    echo "$(date): Service is healthy"
  else
    echo "$(date): Service is unhealthy!"
    # Send alert
  fi
  sleep 30
done
```

### Kubernetes Monitoring

```yaml
apiVersion: v1
kind: Service
metadata:
  name: ccproxy-monitoring
  labels:
    app: ccproxy
spec:
  selector:
    app: ccproxy
  ports:
    - name: http
      port: 3456
      targetPort: 3456

---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: ccproxy
spec:
  selector:
    matchLabels:
      app: ccproxy
  endpoints:
    - port: http
      path: /health
      interval: 30s
```

## Alerting

### Prometheus Alerts

```yaml
# alerts.yml
groups:
  - name: ccproxy
    rules:
      - alert: CCProxyDown
        expr: up{job="ccproxy"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "CCProxy service is down"
          description: "CCProxy has been down for more than 1 minute"

      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors per second"

      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 2
        for: 3m
        labels:
          severity: warning
        annotations:
          summary: "High latency detected"
          description: "95th percentile latency is {{ $value }} seconds"
```

### Notification Channels

```yaml
# alertmanager.yml
route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
  - name: 'web.hook'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'
        channel: '#alerts'
        title: 'CCProxy Alert'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
```

## Custom Monitoring Scripts

### Provider Health Check

```bash
#!/bin/bash
# provider-health.sh
PROVIDERS=("groq" "openai" "gemini" "mistral" "xai" "ollama")

for provider in "${PROVIDERS[@]}"; do
  echo "Checking $provider provider..."
  
  # Test provider endpoint
  response=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Content-Type: application/json" \
    -X POST http://localhost:3456/v1/messages \
    -d '{
      "model": "test",
      "messages": [{"role": "user", "content": "test"}],
      "max_tokens": 10
    }')
  
  if [ "$response" -eq 200 ]; then
    echo "$provider: OK"
  else
    echo "$provider: ERROR (HTTP $response)"
  fi
done
```

### Performance Metrics

```bash
#!/bin/bash
# performance-metrics.sh
LOG_FILE="/var/log/ccproxy/app.log"

# Request count in last hour
echo "Requests in last hour:"
grep "$(date -d '1 hour ago' '+%Y-%m-%dT%H')" $LOG_FILE | \
  grep '"action":"anthropic_request"' | wc -l

# Average response time
echo "Average response time:"
grep '"duration_ms"' $LOG_FILE | \
  jq '.duration_ms' | \
  awk '{sum+=$1; count++} END {print sum/count "ms"}'

# Error rate
echo "Error rate:"
total=$(grep '"level":"' $LOG_FILE | wc -l)
errors=$(grep '"level":"error"' $LOG_FILE | wc -l)
echo "scale=2; $errors * 100 / $total" | bc -l
```

## Integration with External Services

### Datadog

```yaml
# datadog.yaml
logs:
  - type: file
    path: /var/log/ccproxy/app.log
    service: ccproxy
    source: go
    sourcecategory: sourcecode
```

### New Relic

```bash
# Install New Relic Go agent
go get github.com/newrelic/go-agent/v3/newrelic

# Configuration
export NEW_RELIC_LICENSE_KEY=your_license_key
export NEW_RELIC_APP_NAME=CCProxy
```

## Best Practices

1. **Monitor key metrics**: Focus on SLIs (Service Level Indicators)
2. **Set up proactive alerting**: Don't wait for users to report issues
3. **Use dashboards**: Visualize metrics for quick understanding
4. **Regular review**: Analyze trends and adjust thresholds
5. **Test alerts**: Ensure notifications work correctly
6. **Document runbooks**: Create incident response procedures

## Troubleshooting

### Common Monitoring Issues

1. **No data in dashboards**: Check Prometheus scraping configuration
2. **False alerts**: Adjust alert thresholds and timing
3. **Missing metrics**: Verify log format and parsing
4. **High monitoring overhead**: Optimize scraping intervals

### Performance Impact

Monitor the monitoring system itself:

```bash
# Check Prometheus memory usage
ps aux | grep prometheus

# Monitor log file sizes
du -h /var/log/ccproxy/

# Check disk space
df -h
```