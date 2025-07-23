# DeepSeek Provider

## Overview

DeepSeek is an advanced AI provider offering state-of-the-art language models with exceptional reasoning capabilities, competitive coding performance, and significant improvements in accuracy and reliability.

## ⚠️ Important Limitations

- **NO Function Calling Support**: The DeepSeek transformer in CCProxy does not support function calling/tools. This makes DeepSeek **incompatible with Claude Code**, which requires function calling for its operations. Only use DeepSeek through CCProxy for direct API calls that don't require tools.
- **Max Tokens Hard Limit**: DeepSeek enforces a hard limit of 8192 max_tokens. Any request exceeding this will be automatically capped at 8192.
- **Claude Code Compatibility**: Due to lack of function calling support, DeepSeek is NOT recommended for use with Claude Code. Consider using Anthropic, OpenAI, or Gemini providers instead.

## Latest Models (July 2025)

### Available Models

- **deepseek-chat** (maps to DeepSeek-V3-0324) - Latest V3 with improved post-training
- **deepseek-reasoner** (maps to DeepSeek-R1-0528) - Advanced reasoning model with built-in reasoning capabilities
- **deepseek-coder-v2** - Specialized for coding tasks

### Key Features

#### DeepSeek-R1-0528 (deepseek-reasoner)
- **87.5% on AIME 2025** (up from 70%) - Exceptional mathematical reasoning
- **45-50% less hallucination** - Significantly improved reliability
- Advanced multi-step reasoning capabilities via `reasoning_content` field
- Optimal for complex problem-solving tasks
- **Note**: While this model has excellent reasoning capabilities, its limited function calling support may impact Claude Code compatibility

#### DeepSeek-V3-0324 (deepseek-chat)
- Better reasoning and analytical capabilities
- Enhanced coding skills with improved debugging
- Basic tool-use capabilities (limited compared to other providers)
- General-purpose model with balanced performance

#### DeepSeek-Coder-V2
- Specialized for software development tasks
- Superior code generation and understanding
- Multi-language programming support
- Optimized for technical documentation

### Technical Specifications
- Supports system prompts natively
- Built-in reasoning capabilities via `reasoning_content` field (no need for `<think>` tags)
- Streaming support for real-time responses with special handling for reasoning content
- JSON mode support for structured outputs
- **Max tokens**: Hard-capped at 8192 tokens

## Configuration

### Basic Configuration

```json
{
  "providers": {
    "deepseek": {
      "apiKey": "${DEEPSEEK_API_KEY}",
      "baseURL": "https://api.deepseek.com/v1",
      "models": {
        "default": "deepseek-chat",
        "think": "deepseek-reasoner",
        "code": "deepseek-coder-v2"
      }
    }
  }
}
```

### Environment Variables

```bash
export DEEPSEEK_API_KEY="your-api-key-here"
export CCPROXY_CONFIG="/path/to/config.json"
```

## Routing Recommendations

### Default Route
Use **deepseek-chat** for general-purpose tasks:
- Conversational AI
- Content generation
- Basic reasoning tasks
- Tool use and function calling

### Think Route
Use **deepseek-reasoner** for complex reasoning:
- Mathematical problems
- Logic puzzles
- Multi-step reasoning
- Scientific analysis
- Strategic planning

### Code Route
Use **deepseek-coder-v2** for development tasks:
- Code generation
- Code review and debugging
- Technical documentation
- API design
- Algorithm implementation

## Integration Examples

### Claude Code Configuration

```json
{
  "routes": {
    "default": {
      "provider": "deepseek",
      "model": "deepseek-chat"
    },
    "think": {
      "provider": "deepseek",
      "model": "deepseek-reasoner"
    },
    "longContext": {
      "provider": "deepseek",
      "model": "deepseek-chat"
    }
  }
}
```

### API Request Example

```bash
curl -X POST http://localhost:3456/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-ccproxy-key" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "messages": [{"role": "user", "content": "Explain quantum computing"}],
    "max_tokens": 1024
  }'
```

CCProxy will automatically route to DeepSeek based on your configuration.

### Streaming Example

```javascript
const response = await fetch('http://localhost:3456/v1/messages', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'x-api-key': 'your-ccproxy-key',
    'anthropic-version': '2023-06-01'
  },
  body: JSON.stringify({
    model: 'claude-3-5-sonnet-20241022',
    messages: [{ role: 'user', content: 'Write a Python function' }],
    stream: true
  })
});

const reader = response.body.getReader();
// Handle streaming response
```

## Advanced Configuration

### Model-Specific Parameters

```json
{
  "providers": {
    "deepseek": {
      "apiKey": "${DEEPSEEK_API_KEY}",
      "models": {
        "default": "deepseek-chat",
        "think": "deepseek-reasoner",
        "code": "deepseek-coder-v2"
      },
      "parameters": {
        "temperature": 0.7,
        "top_p": 0.95,
        "frequency_penalty": 0,
        "presence_penalty": 0
      }
    }
  }
}
```

### Multi-Route Configuration

```json
{
  "routes": {
    "math": {
      "provider": "deepseek",
      "model": "deepseek-reasoner",
      "parameters": {
        "temperature": 0.1
      }
    },
    "creative": {
      "provider": "deepseek", 
      "model": "deepseek-chat",
      "parameters": {
        "temperature": 0.9
      }
    },
    "debug": {
      "provider": "deepseek",
      "model": "deepseek-coder-v2",
      "parameters": {
        "temperature": 0
      }
    }
  }
}
```

## Performance Considerations

### Model Selection
- **deepseek-reasoner**: Best for accuracy-critical tasks, slightly higher latency
- **deepseek-chat**: Balanced performance for most use cases
- **deepseek-coder-v2**: Optimized for code generation speed

### Optimization Tips
1. Use appropriate models for specific tasks
2. Adjust temperature based on task requirements
3. Enable streaming for better perceived performance
4. Consider token limits for cost optimization

## Troubleshooting

### Common Issues

**API Key Errors**
```bash
# Verify API key is set
echo $DEEPSEEK_API_KEY

# Test direct API access
curl https://api.deepseek.com/v1/models \
  -H "Authorization: Bearer $DEEPSEEK_API_KEY"
```

**Model Not Found**
- Ensure you're using the correct model names: `deepseek-chat`, `deepseek-reasoner`, or `deepseek-coder-v2`
- Check CCProxy configuration for proper model mapping

**Rate Limiting**
- DeepSeek has rate limits per API key
- Implement retry logic with exponential backoff
- Consider using multiple API keys for high-volume applications

## Best Practices

1. **Model Selection**: Choose models based on task requirements
   - For Claude Code: Consider using other providers with better function calling support
   - For reasoning tasks without tool use: deepseek-reasoner is excellent
   - For general chat: deepseek-chat works well within its limitations
2. **Temperature Settings**: Lower for factual tasks, higher for creative tasks
3. **System Prompts**: Leverage native system prompt support for consistent behavior
4. **Token Limits**: Be aware of the 8192 token hard limit when designing prompts
5. **Error Handling**: Implement robust error handling for API failures and function calling limitations
6. **Monitoring**: Track usage and performance metrics, especially tool use failures

## When to Use DeepSeek

**Good Use Cases:**
- Mathematical reasoning and problem solving
- Code generation without extensive tool use
- General conversation and Q&A
- Tasks requiring deep reasoning but not function calling

**Consider Alternatives When:**
- Using Claude Code with its extensive function calling requirements
- Needing reliable tool/function calling capabilities
- Requiring more than 8192 max tokens
- Building applications heavily dependent on tool use

## See Also

- [CCProxy Configuration Guide](/guide/configuration)
- [Provider Overview](/providers/)
- [API Documentation](/api/)