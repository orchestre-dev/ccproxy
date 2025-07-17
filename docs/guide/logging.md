# Logging

CCProxy provides comprehensive logging for monitoring, debugging, and auditing purposes.

## Log Configuration

### Environment Variables

```bash
# Log level configuration
LOG_LEVEL=info    # Options: debug, info, warn, error

# Log format configuration
LOG_FORMAT=json   # Options: json, text
```

### Log Levels

- **debug**: Detailed debugging information
- **info**: General information about application flow
- **warn**: Warning messages for potential issues
- **error**: Error messages for failures

## Log Formats

### JSON Format (Recommended for Production)

```json
{
  "level": "info",
  "msg": "Successfully initialized groq provider with model moonshotai/kimi-k2-instruct",
  "time": "2025-01-17T10:30:00Z"
}
```

### Text Format (Human-Readable)

```
INFO[2025-01-17T10:30:00Z] Successfully initialized groq provider with model moonshotai/kimi-k2-instruct
```

## Log Types

### Application Logs

Standard application lifecycle logs:

```json
{
  "level": "info",
  "msg": "Starting server on 0.0.0.0:7187",
  "time": "2025-01-17T10:30:00Z"
}
```

### API Action Logs

Structured logs for API requests and responses:

```json
{
  "action": "anthropic_request",
  "level": "info",
  "max_tokens": 100,
  "messages": 1,
  "model": "claude-3-sonnet",
  "msg": "API action",
  "provider": "groq",
  "request_id": "req_123456789",
  "time": "2025-01-17T10:30:00Z",
  "tools": 0,
  "type": "api_action"
}
```

### Error Logs

Detailed error information:

```json
{
  "level": "error",
  "msg": "Failed to call groq API",
  "error": "connection timeout",
  "request_id": "req_123456789",
  "time": "2025-01-17T10:30:00Z"
}
```

## Request Tracking

### Request ID

Each request is assigned a unique ID for tracking:

```json
{
  "request_id": "f658e30a-cdfa-40c6-9d82-d8036a6157ec",
  "msg": "Processing request",
  "level": "info"
}
```

### Request Flow

Complete request lifecycle logging:

1. **Request received**: Log incoming request details
2. **Provider call**: Log outgoing API call
3. **Provider response**: Log response from provider
4. **Response sent**: Log final response to client

## Log Aggregation

### ELK Stack

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
    date {
      match => [ "time", "ISO8601" ]
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

### Fluentd

```xml
<source>
  @type tail
  path /var/log/ccproxy/app.log
  pos_file /var/log/fluentd/ccproxy.log.pos
  tag ccproxy
  format json
  time_format %Y-%m-%dT%H:%M:%S.%L%z
</source>

<match ccproxy>
  @type elasticsearch
  host elasticsearch
  port 9200
  index_name ccproxy
  type_name logs
</match>
```

## Monitoring and Alerting

### Log-Based Metrics

Extract metrics from logs for monitoring:

```bash
# Count errors per minute
grep '"level":"error"' /var/log/ccproxy/app.log | wc -l

# Track request latency
grep '"action":".*_response"' /var/log/ccproxy/app.log | jq '.duration_ms'

# Monitor provider errors
grep '"provider":"groq"' /var/log/ccproxy/app.log | grep '"level":"error"'
```

### Alerting Rules

Set up alerts based on log patterns:

```yaml
# Alert on high error rate
- alert: HighErrorRate
  expr: increase(log_errors_total[5m]) > 10
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "High error rate detected in CCProxy"

# Alert on provider failures
- alert: ProviderDown
  expr: increase(log_provider_errors_total[5m]) > 5
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Provider {{ $labels.provider }} is failing"
```

## Log Rotation

### Logrotate Configuration

```bash
# /etc/logrotate.d/ccproxy
/var/log/ccproxy/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0644 ccproxy ccproxy
    postrotate
        /bin/kill -USR1 `cat /var/run/ccproxy.pid 2> /dev/null` 2> /dev/null || true
    endscript
}
```

## Best Practices

1. **Use JSON format** in production for easier parsing
2. **Include request IDs** in all logs for request tracing
3. **Log at appropriate levels** - avoid debug logs in production
4. **Rotate logs regularly** to prevent disk space issues
5. **Monitor log volume** to detect issues early
6. **Sanitize sensitive data** - never log API keys or personal information

## Troubleshooting

### Common Issues

1. **No logs appearing**: Check log level and format configuration
2. **Logs filling disk**: Implement log rotation
3. **Missing request IDs**: Ensure middleware is properly configured
4. **Performance impact**: Use appropriate log levels for production

### Debug Logging

Enable debug logging for troubleshooting:

```bash
# Temporary debug mode
LOG_LEVEL=debug ./ccproxy

# Or set environment variable
export LOG_LEVEL=debug
```

### Log Analysis

```bash
# Search for specific errors
grep '"level":"error"' /var/log/ccproxy/app.log | jq '.msg'

# Filter by request ID
grep '"request_id":"req_123456789"' /var/log/ccproxy/app.log

# Monitor real-time logs
tail -f /var/log/ccproxy/app.log | jq .
```