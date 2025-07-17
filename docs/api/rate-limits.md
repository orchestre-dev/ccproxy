# Rate Limiting

CCProxy implements rate limiting to protect against abuse and ensure fair usage across all clients.

## Rate Limiting Overview

Rate limiting is applied per client based on:
- **Client IP address** (fallback identification)
- **API key** (when provided in Authorization header)

## Default Limits

### Per-Client Limits

- **Request Rate**: 100 requests per minute
- **Burst Capacity**: 10 requests (allows temporary bursts)
- **Window**: 60 seconds (sliding window)

### Global Limits

- **Total Throughput**: No global limits (depends on provider limits)
- **Concurrent Connections**: No artificial limits

## Rate Limit Headers

CCProxy includes standard rate limit headers in all responses:

```http
X-RateLimit-Limit: 100
X-RateLimit-Window: 60s
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642684800
```

## Rate Limit Exceeded Response

When rate limits are exceeded, CCProxy returns:

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
X-RateLimit-Limit: 100
X-RateLimit-Window: 60s
Retry-After: 60

{
  "error": "Rate limit exceeded",
  "message": "Too many requests. Limit: 100 per 60s",
  "retry_after": "60s"
}
```

## Client Identification

### API Key-Based (Recommended)

```http
POST /v1/messages
Authorization: Bearer your_api_key_here
Content-Type: application/json

{
  "model": "claude-3-sonnet",
  "messages": [{"role": "user", "content": "Hello"}]
}
```

### IP-Based (Fallback)

When no API key is provided, rate limiting falls back to IP address:

```http
POST /v1/messages
Content-Type: application/json

{
  "model": "claude-3-sonnet",
  "messages": [{"role": "user", "content": "Hello"}]
}
```

## Configuration

Rate limiting is currently not configurable via environment variables but uses sensible defaults.

### Future Configuration Options

```bash
# Planned environment variables
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60s
RATE_LIMIT_BURST=10
```

## Best Practices

### For API Clients

1. **Include API Key**: Always include your API key for consistent rate limiting
2. **Respect Headers**: Check rate limit headers and implement backoff
3. **Handle 429 Responses**: Implement retry logic with exponential backoff
4. **Distribute Requests**: Avoid sending all requests in bursts

### Example Client Implementation

```javascript
async function makeRequest(data, retries = 3) {
  try {
    const response = await fetch('/v1/messages', {
      method: 'POST',
      headers: {
        'Authorization': 'Bearer your_api_key',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    });
    
    if (response.status === 429) {
      const retryAfter = response.headers.get('Retry-After');
      if (retries > 0) {
        await new Promise(resolve => setTimeout(resolve, retryAfter * 1000));
        return makeRequest(data, retries - 1);
      }
      throw new Error('Rate limit exceeded');
    }
    
    return response.json();
  } catch (error) {
    console.error('Request failed:', error);
    throw error;
  }
}
```

### For Server Operators

1. **Monitor Rate Limiting**: Track rate limit hits and adjust if needed
2. **Whitelist Trusted Clients**: Consider IP whitelisting for trusted sources
3. **Scale Horizontally**: Use load balancers to distribute load
4. **Provider Limits**: Respect upstream provider rate limits

## Monitoring Rate Limits

### Log Analysis

Rate limit events are logged for monitoring:

```json
{
  "level": "warn",
  "msg": "Rate limit exceeded",
  "client_id": "api_abc123",
  "requests": 101,
  "limit": 100,
  "window": "60s",
  "time": "2025-01-17T10:30:00Z"
}
```

### Metrics Collection

Track rate limiting metrics:

```bash
# Count rate limit hits
grep "Rate limit exceeded" /var/log/ccproxy/app.log | wc -l

# Monitor by client
grep "Rate limit exceeded" /var/log/ccproxy/app.log | jq '.client_id' | sort | uniq -c
```

## Rate Limit Bypass

### Trusted Networks

For internal services, you might want to bypass rate limiting:

```bash
# Future configuration option
RATE_LIMIT_BYPASS_IPS=10.0.0.0/8,172.16.0.0/12,192.168.0.0/16
```

### Service Accounts

Dedicated service accounts with higher limits:

```bash
# Future configuration option
RATE_LIMIT_SERVICE_ACCOUNTS=service1:1000,service2:500
```

## Troubleshooting

### Common Issues

1. **Unexpected Rate Limits**: Check if multiple clients share the same IP
2. **Rate Limit Not Working**: Verify client identification (IP vs API key)
3. **False Positives**: Consider network setup (proxies, load balancers)

### Debugging Rate Limits

```bash
# Check current rate limit status
curl -I http://localhost:7187/health

# Monitor rate limit headers
curl -H "Authorization: Bearer your_key" -I http://localhost:7187/v1/messages

# Test rate limiting
for i in {1..105}; do
  curl -s -o /dev/null -w "%{http_code}\n" http://localhost:7187/health
done
```

## Provider-Specific Considerations

### Groq Rate Limits

- **Requests per minute**: 30 (varies by model)
- **Tokens per minute**: 7,000 (varies by model)
- **Burst capacity**: Higher for paid plans

### OpenAI Rate Limits

- **Requests per minute**: 3,500 (varies by tier)
- **Tokens per minute**: 90,000 (varies by tier)
- **Concurrent requests**: 100

### Gemini Rate Limits

- **Requests per minute**: 1,500 (free tier)
- **Tokens per minute**: 32,000 (free tier)
- **Daily quota**: 1,500 requests

CCProxy's rate limiting works in conjunction with provider limits to ensure smooth operation without hitting upstream limits.

## Future Enhancements

- **Per-user rate limiting**: Individual limits per API key
- **Dynamic rate limiting**: Adjust limits based on usage patterns
- **Rate limit sharing**: Distribute limits across multiple instances
- **Advanced algorithms**: Token bucket, sliding window counters