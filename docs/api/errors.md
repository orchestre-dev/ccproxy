---
title: Error Handling - CCProxy API Reference
description: Understand CCProxy error responses, status codes, and troubleshooting. Handle API errors gracefully in your Claude Code integration.
keywords: CCProxy errors, API error handling, troubleshooting, status codes
---

# Error Handling

Understand CCProxy error responses and handle them gracefully.

<SocialShare />

## Error Response Format

All errors follow a consistent JSON format:

```json
{
  "error": {
    "type": "authentication_error",
    "message": "Invalid API key provided",
    "code": "invalid_api_key",
    "details": {
      "provider": "groq",
      "suggestion": "Check your GROQ_API_KEY environment variable"
    }
  }
}
```

## HTTP Status Codes

| Code | Type | Description |
|------|------|-------------|
| `400` | Bad Request | Invalid request format or parameters |
| `401` | Unauthorized | Invalid or missing API key |
| `403` | Forbidden | API key lacks required permissions |
| `429` | Rate Limited | Too many requests |
| `500` | Internal Error | CCProxy internal error |
| `502` | Bad Gateway | Provider API error |
| `503` | Service Unavailable | Provider temporarily unavailable |

## Common Error Types

### Authentication Errors

**Invalid API Key**
```json
{
  "error": {
    "type": "authentication_error",
    "message": "Invalid API key for provider 'groq'",
    "code": "invalid_api_key"
  }
}
```

**Missing API Key**
```json
{
  "error": {
    "type": "authentication_error", 
    "message": "API key not configured for provider 'openai'",
    "code": "missing_api_key"
  }
}
```

### Rate Limiting

```json
{
  "error": {
    "type": "rate_limit_error",
    "message": "Rate limit exceeded",
    "code": "rate_limit_exceeded",
    "details": {
      "retry_after": 60,
      "limit": 5000,
      "remaining": 0
    }
  }
}
```

### Provider Errors

**Model Not Found**
```json
{
  "error": {
    "type": "provider_error",
    "message": "Model 'invalid-model' not found",
    "code": "model_not_found",
    "details": {
      "provider": "groq",
      "available_models": ["moonshotai/kimi-k2-instruct", "llama-3.1-70b-versatile"]
    }
  }
}
```

**Provider Unavailable**
```json
{
  "error": {
    "type": "provider_error",
    "message": "Provider 'groq' is temporarily unavailable",
    "code": "provider_unavailable",
    "details": {
      "retry_after": 30
    }
  }
}
```

## Handling Errors in Code

### Python Example
```python
import requests

try:
    response = requests.post(
        "http://localhost:3456/v1/messages",
        json={"model": "claude-3-sonnet", "messages": [...]},
        headers={"Content-Type": "application/json"}
    )
    response.raise_for_status()
    return response.json()
    
except requests.exceptions.HTTPError as e:
    error = e.response.json().get("error", {})
    
    if error.get("code") == "rate_limit_exceeded":
        retry_after = error.get("details", {}).get("retry_after", 60)
        print(f"Rate limited, retry after {retry_after} seconds")
        
    elif error.get("code") == "invalid_api_key":
        print("Invalid API key, check configuration")
        
    else:
        print(f"API error: {error.get('message')}")
```

### JavaScript Example
```javascript
async function callCCProxy(messages) {
  try {
    const response = await fetch('http://localhost:3456/v1/messages', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        model: 'claude-3-sonnet',
        messages
      })
    });

    if (!response.ok) {
      const error = await response.json();
      
      switch (error.error?.code) {
        case 'rate_limit_exceeded':
          const retryAfter = error.error.details?.retry_after || 60;
          throw new Error(`Rate limited, retry after ${retryAfter}s`);
          
        case 'provider_unavailable':
          throw new Error('Provider temporarily unavailable');
          
        default:
          throw new Error(error.error?.message || 'Unknown error');
      }
    }

    return await response.json();
    
  } catch (error) {
    console.error('CCProxy error:', error.message);
    throw error;
  }
}
```

## Troubleshooting Guide

### Provider Connection Issues

1. **Check API key**: Verify your provider API key is correct
2. **Check network**: Ensure CCProxy can reach provider APIs
3. **Check rate limits**: Monitor your provider usage
4. **Check model availability**: Verify model exists for your provider

### CCProxy Issues

1. **Check logs**: Enable debug logging with `LOG_LEVEL=debug`
2. **Check health**: Use `/health` and `/status` endpoints
3. **Check configuration**: Verify environment variables
4. **Check resources**: Monitor memory and CPU usage

### Common Solutions

**Authentication Failed**
```bash
# Check environment variables
echo $GROQ_API_KEY
echo $PROVIDER

# Test provider directly
curl -H "Authorization: Bearer $GROQ_API_KEY" \
  https://api.groq.com/openai/v1/models
```

**Connection Refused**
```bash
# Check if CCProxy is running
curl http://localhost:3456/health

# Check port availability
lsof -i :3456
```

## Best Practices

1. **Implement retry logic** with exponential backoff
2. **Handle rate limits** gracefully with delays
3. **Log errors** for debugging and monitoring
4. **Validate requests** before sending to CCProxy
5. **Monitor health** endpoints regularly

## See Also

- [Status Endpoint](/api/status) - Monitor proxy health
- [Health Endpoints](/api/health) - Simple health checks