# Intelligent Model Routing Guide

## Overview

CCProxy automatically selects the best AI model based on your request characteristics, ensuring optimal performance, cost efficiency, and response quality. The intelligent routing system analyzes multiple factors including token count, model availability, task complexity, and specific request parameters to make routing decisions.

## How Routing Works

The router evaluates requests in a priority order, selecting the first matching rule:

1. **Explicit Provider Selection**: When you specify "provider,model" format
2. **Direct Model Routes**: Configured model-specific routes
3. **Token-Based Routing**: Automatic selection for long context (>60K tokens)
4. **Background Task Routing**: Routes to fast models for background processing
5. **Thinking Parameter Routing**: Complex reasoning tasks requiring deep thinking
6. **Default Route**: Fallback to configured default model

## Routing Rules Priority

### 1. Explicit Provider Selection
When you specify a provider and model using comma notation:
```json
{
  "model": "anthropic,claude-opus-4-20250720"
}
```

### 2. Direct Model Routes
Model names that match configured routes:
```json
{
  "model": "gpt-4-turbo"  // Routes to specific provider
}
```

### 3. Automatic Token-Based Routing
Requests exceeding 60,000 tokens automatically route to long-context models:
```json
{
  "messages": [...],  // If total tokens > 60K, routes to longContext
}
```

### 4. Background Task Routing
Models ending in "-haiku" trigger fast model routing:
```json
{
  "model": "claude-3-5-haiku-20241022"
}
```

### 5. Thinking Parameter Routing
Complex reasoning tasks with thinking enabled (boolean parameter):
```json
{
  "model": "claude-3-5-sonnet-20241022",
  "thinking": true  // Boolean flag, not thinkingBudget
}
```

### 6. Default Route
When no specific rules match, uses the configured default model.

## Recommended Routes by Use Case

### General Purpose (Default)

Best for everyday tasks, balanced between performance and cost.

**Recommended Models:**
- **Anthropic**: claude-3-5-sonnet-20241022
- **OpenAI**: gpt-4-turbo-2024-04-09
- **Google**: gemini-1.5-flash
- **DeepSeek**: deepseek-chat (Note: 8192 token limit)

**Configuration Example:**
```json
{
  "providers": {
    "anthropic": {
      "models": {
        "claude-3-5-sonnet-20241022": {
          "route": "default"
        }
      }
    }
  },
  "router": {
    "routes": {
      "default": "anthropic,claude-3-5-sonnet-20241022"
    }
  }
}
```

### Long Context (>60K tokens)

For processing large documents, extensive conversations, or code analysis.

**Recommended Models:**
- **Anthropic**: claude-3-opus-20240229 (200K context)
- **OpenAI**: gpt-4-32k (32K context)
- **Google**: gemini-1.5-pro (1M context)

**Configuration Example:**
```json
{
  "router": {
    "routes": {
      "longContext": "google,gemini-1.5-pro"
    },
    "rules": {
      "tokenThreshold": 60000
    }
  }
}
```

### Fast/Background Tasks

For quick responses, simple tasks, or high-volume processing.

**Recommended Models:**
- **Anthropic**: claude-3-haiku-20240307
- **OpenAI**: gpt-3.5-turbo
- **Google**: gemini-1.5-flash
- **DeepSeek**: deepseek-chat (Note: 8192 token limit)

**Configuration Example:**
```json
{
  "router": {
    "routes": {
      "fast": "anthropic,claude-3-haiku-20240307",
      "background": "openai,gpt-3.5-turbo"
    },
    "modelPatterns": {
      "haiku": "fast",
      "turbo": "fast"
    }
  }
}
```

### Complex Reasoning (Think Mode)

For tasks requiring deep analysis, multi-step reasoning, or complex problem solving. The router checks for a boolean `thinking` field in the request.

**Recommended Models:**
- **Anthropic**: claude-3-opus-20240229
- **OpenAI**: o1-preview
- **Google**: gemini-1.5-pro
- **DeepSeek**: deepseek-coder

**Configuration Example:**
```json
{
  "router": {
    "routes": {
      "think": "anthropic,claude-3-opus-20240229"
    }
}
```

**Note**: When using Claude Code, only providers that support function calling will work properly. This includes Anthropic, OpenAI, and Google Gemini. DeepSeek and some other providers may have limited or no function calling support.

## Advanced Configuration Examples

### Multi-Provider Fallback
```json
{
  "router": {
    "routes": {
      "default": "anthropic,claude-3-5-sonnet-20241022"
    },
    "fallbacks": {
      "anthropic": ["openai,gpt-4-turbo", "google,gemini-1.5-flash"],
      "openai": ["anthropic,claude-3-5-sonnet-20241022"],
      "google": ["anthropic,claude-3-5-sonnet-20241022"]
    }
  }
}
```

### Cost-Optimized Routing
```json
{
  "router": {
    "routes": {
      "default": "google,gemini-1.5-flash",
      "fast": "anthropic,claude-3-haiku-20240307"
    }
  }
}
```

### Direct Model Routes
```json
{
  "router": {
    "routes": {
      "gpt-4-turbo": "openai,gpt-4-turbo-2024-04-09",
      "claude-3-5-sonnet-20241022": "anthropic,claude-3-5-sonnet-20241022",
      "gemini-1.5-pro": "google,gemini-1.5-pro"
    }
  }
}
```

### Multiple Provider Configuration
```json
{
  "router": {
    "routes": {
      "default": "anthropic,claude-3-5-sonnet-20241022",
      "longContext": "google,gemini-1.5-pro",
      "fast": "openai,gpt-3.5-turbo",
      "think": "anthropic,claude-3-opus-20240229"
    }
  }
}
```

## Best Practices

### 1. Token Count Awareness
- Monitor your typical token usage
- Set appropriate thresholds for long-context routing
- Consider chunking large requests

### 2. Cost Management
- Use fast models for simple tasks
- Reserve premium models for complex reasoning
- Implement usage quotas per model

### 3. Latency Optimization
- Route time-sensitive requests to fast models
- Use geographic routing for global deployments
- Implement timeout-based fallbacks

### 4. Error Handling
- Configure fallback routes for each provider
- Set up circuit breakers for failing models
- Monitor provider availability

### 5. Testing Routes
```bash
# Test specific routing
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-key" \
  -d '{
    "model": "anthropic,claude-3-5-sonnet-20241022",
    "messages": [{"role": "user", "content": "Test"}]
  }'

# Test automatic routing (long context)
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-key" \
  -d '{
    "messages": [{"role": "user", "content": "Very long content..."}]
  }'
```

## Monitoring and Debugging

### View Current Routes
```bash
curl http://localhost:3456/health \
  -H "x-api-key: your-key" | jq .router
```

### Route Decision Logging
Enable debug logging to see routing decisions:
```json
{
  "logging": {
    "level": "debug",
    "routingDecisions": true
  }
}
```

### Metrics Collection
Monitor routing patterns:
- Route hit rates
- Token distribution
- Provider utilization
- Fallback frequency

## Common Routing Scenarios

### Scenario 1: Development vs Production
```json
{
  "environments": {
    "development": {
      "router": {
        "routes": {
          "default": "openai,gpt-3.5-turbo"
        }
      }
    },
    "production": {
      "router": {
        "routes": {
          "default": "anthropic,claude-3-5-sonnet-20241022"
        }
      }
    }
  }
}
```

### Scenario 2: Model Pattern Routing
```json
{
  "router": {
    "routes": {
      "default": "anthropic,claude-3-5-sonnet-20241022",
      "fast": "anthropic,claude-3-haiku-20240307"
    },
    "modelPatterns": {
      "haiku": "fast"
    }
  }
}
```

## Troubleshooting

### Route Not Working
1. Check provider configuration
2. Verify API keys are set
3. Ensure model names are correct
4. Review routing priority

### Unexpected Model Selection
1. Enable debug logging
2. Check token count
3. Verify routing rules
4. Test with explicit routing

### Performance Issues
1. Monitor route latency
2. Check provider availability
3. Review fallback configuration
4. Optimize route selection

## Limitations

- **Function Calling**: When using Claude Code, only providers with function calling support will work properly (Anthropic, OpenAI, Google Gemini)
- **DeepSeek**: Limited to 8192 tokens per request
- **Routing Rules**: Currently supports token-based, model pattern, and thinking parameter routing
- **Configuration**: Routes must be explicitly configured in the config file