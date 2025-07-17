---
title: OpenAI with Claude Code - Industry-Leading AI Models via CCProxy
description: Access GPT-4, GPT-4o, and o1 reasoning models with Claude Code through CCProxy. Experience enterprise-grade AI with vision capabilities, function calling, and production reliability.
keywords: OpenAI, Claude Code, CCProxy, GPT-4, GPT-4o, o1 reasoning, AI proxy, enterprise AI, vision AI, function calling, production AI
---

# OpenAI Provider

<div class="social-share">
  <button class="share-twitter" onclick="shareToTwitter()">
    🐦 Share on Twitter
  </button>
  <button class="share-linkedin" onclick="shareToLinkedIn()">
    💼 Share on LinkedIn
  </button>
  <button class="share-reddit" onclick="shareToReddit()">
    🔗 Share on Reddit
  </button>
  <button class="share-copy" onclick="copyToClipboard()">
    📋 Copy Link
  </button>
</div>

**OpenAI sets the industry standard** for AI models, offering the most mature ecosystem and reliable performance for production applications. Through **CCProxy integration with Claude Code**, you can harness the full power of GPT-4, GPT-4o, and advanced reasoning models while maintaining your familiar development workflow.

## 🏭 Why Choose OpenAI for Claude Code?

- 🥇 **Industry standard**: Most mature and reliable AI models with proven enterprise adoption
- 🛠️ **Rich ecosystem**: Extensive tooling and integrations that work seamlessly with Claude Code
- 👁️ **Advanced vision**: Best-in-class image understanding and multimodal capabilities
- 🎯 **Proven reliability**: Battle-tested in production environments worldwide
- 🧠 **Advanced reasoning**: Access to o1 models for complex problem-solving
- 🔧 **Perfect Claude Code integration**: Zero configuration changes required with CCProxy

## Setup

### 1. Get an API Key

1. Visit [platform.openai.com](https://platform.openai.com)
2. Sign up for an account
3. Navigate to the API Keys section
4. Generate a new API key

### 2. Configure CCProxy

Set the following environment variables:

```bash
export PROVIDER=openai
export OPENAI_API_KEY=sk-your_openai_api_key_here
```

### 3. Optional Configuration

```bash
# Custom model (default: gpt-4o)
export OPENAI_MODEL=gpt-4o-mini

# Custom max tokens (default: 16384)
export OPENAI_MAX_TOKENS=8192

# Custom base URL (default: https://api.openai.com/v1)
export OPENAI_BASE_URL=https://api.openai.com/v1

# Organization ID (optional)
export OPENAI_ORGANIZATION=org-your_org_id_here
```

## Available Models

OpenAI provides access to various model families:

- **GPT-4 Series** - Latest and most capable models with multimodal support
- **GPT-3.5 Series** - Cost-effective models for simpler tasks
- **o1 Series** - Advanced reasoning models for complex problem-solving
- **Specialized Models** - Fine-tuned models for specific use cases

**🔧 Critical for Claude Code**: You must select models that support **tool calling** or **function calling** capabilities, as Claude Code requires these features to operate correctly.

### Model Selection Guidelines

When choosing OpenAI models:

1. **Verify Tool Support**: Ensure the model supports function calling
2. **Check Current Availability**: OpenAI's model lineup evolves frequently
3. **Consider Cost vs Performance**: Balance quality needs with budget
4. **Review Context Limits**: Different models have different context windows
5. **Test Capabilities**: Some models excel at specific tasks

For current model availability, capabilities, and pricing, visit [OpenAI's official models page](https://platform.openai.com/docs/models).

## Pricing

### Free Tier
- Free credits for new users
- Limited time availability
- Perfect for testing and development

### Pay-as-you-go Pricing
- Competitive per-token pricing
- No minimum spend required
- Volume discounts available for enterprise

For current, accurate pricing information, visit [OpenAI's official pricing page](https://openai.com/pricing).

## Configuration Examples

### Basic Setup

```bash
# .env file
PROVIDER=openai
OPENAI_API_KEY=sk-your_api_key_here
```

### High-Performance Setup

```bash
# For maximum speed
PROVIDER=openai
OPENAI_API_KEY=sk-your_api_key_here
OPENAI_MODEL=gpt-4o-mini
OPENAI_MAX_TOKENS=4096
```

### Quality-Focused Setup

```bash
# For best quality
PROVIDER=openai
OPENAI_API_KEY=sk-your_api_key_here
OPENAI_MODEL=gpt-4o
OPENAI_MAX_TOKENS=16384
```

### Reasoning-Focused Setup

```bash
# For complex reasoning tasks
PROVIDER=openai
OPENAI_API_KEY=sk-your_api_key_here
OPENAI_MODEL=o1-preview
OPENAI_MAX_TOKENS=8192
```

## Usage with Claude Code

Once configured, use Claude Code normally:

```bash
# Set CCProxy as the API endpoint
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED

# Use Claude Code
claude "Write a Python script to analyze CSV data"
```

## Features

### ✅ Fully Supported
- Text generation
- Function calling
- Tool use
- Streaming responses
- Vision/image input
- JSON mode
- Custom temperature/top_p
- Structured outputs
- Reasoning models

### ⚠️ Model Dependent
- Real-time data access (none)
- File uploads (vision models only)
- Code execution (none)

## Advanced Features

### Vision Capabilities

Use GPT-4o or GPT-4-turbo with images:

```bash
# Configure for vision tasks
export OPENAI_MODEL=gpt-4o
```

### JSON Mode

Force JSON responses:

```bash
# Enable JSON mode in requests
# (handled automatically by CCProxy when tools are used)
```

### Function Calling

OpenAI has the most robust function calling:

```bash
# All models support function calling
# No additional configuration needed
```

## Performance Tips

### 1. Choose the Right Model

```bash
# For development/testing
export OPENAI_MODEL=gpt-4o-mini

# For production
export OPENAI_MODEL=gpt-4o

# For complex reasoning
export OPENAI_MODEL=o1-preview
```

### 2. Optimize Token Usage

```bash
# Set appropriate limits
export OPENAI_MAX_TOKENS=2048

# Use shorter prompts when possible
```

### 3. Handle Rate Limits

OpenAI has different rate limits per model:

| Model | Requests/min | Tokens/min |
|-------|--------------|------------|
| **gpt-4o** | 10,000 | 30,000,000 |
| **gpt-4o-mini** | 30,000 | 200,000,000 |
| **gpt-4-turbo** | 10,000 | 2,000,000 |
| **o1-preview** | 20 | 20,000 |

## Troubleshooting

### Rate Limit Errors

```json
{
  "error": {
    "message": "Rate limit exceeded",
    "type": "rate_limit_error"
  }
}
```

**Solution**: Implement exponential backoff or upgrade to higher tier.

### Authentication Errors

```json
{
  "error": {
    "message": "Incorrect API key provided",
    "type": "invalid_request_error"
  }
}
```

**Solution**: Verify your API key is correct and active.

### Model Access Errors

```json
{
  "error": {
    "message": "The model does not exist",
    "type": "invalid_request_error"
  }
}
```

**Solution**: Check if you have access to the model. Some models require specific tier access.

### Content Policy Violations

```json
{
  "error": {
    "message": "Your request was rejected as a result of our safety system",
    "type": "invalid_request_error"
  }
}
```

**Solution**: Modify your prompt to comply with OpenAI's usage policies.

### Insufficient Quota

```json
{
  "error": {
    "message": "You exceeded your current quota",
    "type": "insufficient_quota"
  }
}
```

**Solution**: Add credits to your OpenAI account or upgrade your plan.

## Usage Monitoring

### View Usage

```bash
# Check OpenAI usage
curl https://api.openai.com/v1/usage \
  -H "Authorization: Bearer $OPENAI_API_KEY"

# Monitor CCProxy logs
tail -f ccproxy.log

# Check status
curl http://localhost:7187/status
```

### Cost Management

```bash
# Set usage limits in OpenAI dashboard
# Monitor spending regularly
# Use gpt-4o-mini for development
```

## Best Practices

### 1. Model Selection

- **Development**: Use `gpt-4o-mini` for cost efficiency
- **Production**: Use `gpt-4o` for reliability
- **Complex reasoning**: Use `o1-preview` for difficult problems

### 2. Prompt Engineering

```bash
# Use clear, specific prompts
# Include examples when needed
# Structure complex requests
```

### 3. Error Handling

```bash
# Implement retry logic
# Handle rate limits gracefully
# Monitor for content policy issues
```

### 4. Security

```bash
# Never log API keys
# Use environment variables
# Rotate keys regularly
```

## Integration Examples

### Python SDK

```python
import openai

# Point to CCProxy
client = openai.OpenAI(
    api_key="NOT_NEEDED",
    base_url="http://localhost:7187"
)

response = client.chat.completions.create(
    model="claude-3-sonnet",  # Will be mapped to OpenAI model
    messages=[{"role": "user", "content": "Hello!"}],
    max_tokens=100
)
```

### Node.js SDK

```javascript
import OpenAI from 'openai';

const client = new OpenAI({
    apiKey: 'NOT_NEEDED',
    baseURL: 'http://localhost:7187'
});

const response = await client.chat.completions.create({
    model: 'claude-3-sonnet',
    messages: [{ role: 'user', content: 'Hello!' }],
    max_tokens: 100
});
```

## Monitoring and Analytics

### OpenAI Dashboard

Monitor usage at [platform.openai.com/usage](https://platform.openai.com/usage):

- Request counts
- Token usage
- Cost breakdown
- Error rates

### CCProxy Monitoring

```bash
# Real-time logs
tail -f ccproxy.log | grep openai

# Status endpoint
curl http://localhost:7187/status

# Health check
curl http://localhost:7187/health
```

## Next Steps

- Explore [function calling](/guide/function-calling) with OpenAI models
- Learn about [vision capabilities](/guide/vision) with GPT-4o
- Set up [usage monitoring](/guide/monitoring) for cost control
- Compare with [other providers](/providers/) including [Groq with Kimi K2](/providers/groq) for speed

<script>
function shareToTwitter() {
  const url = encodeURIComponent(window.location.href);
  const text = encodeURIComponent('🏭 OpenAI + Claude Code + CCProxy = Enterprise-grade AI development! Access GPT-4, GPT-4o, and o1 reasoning models seamlessly');
  window.open(`https://twitter.com/intent/tweet?url=${url}&text=${text}`, '_blank');
}

function shareToLinkedIn() {
  const url = encodeURIComponent(window.location.href);
  window.open(`https://www.linkedin.com/sharing/share-offsite/?url=${url}`, '_blank');
}

function shareToReddit() {
  const url = encodeURIComponent(window.location.href);
  const title = encodeURIComponent('OpenAI with Claude Code - Industry-Leading AI Models via CCProxy');
  window.open(`https://reddit.com/submit?url=${url}&title=${title}`, '_blank');
}

function copyToClipboard() {
  navigator.clipboard.writeText(window.location.href).then(() => {
    const button = event.target;
    const originalText = button.textContent;
    button.textContent = '✅ Copied!';
    setTimeout(() => {
      button.textContent = originalText;
    }, 2000);
  });
}
</script>