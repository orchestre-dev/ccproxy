---
title: OpenRouter with Claude Code - 100+ AI Models via CCProxy
description: Access 100+ AI models including Claude, GPT-4, Llama, and Mistral with Claude Code through CCProxy and OpenRouter. Perfect for model experimentation and cost optimization with unified API access.
keywords: OpenRouter, Claude Code, CCProxy, 100+ AI models, model fallbacks, Claude 3.5, GPT-4, Llama, model comparison, unified AI API
---

# OpenRouter Provider

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

**OpenRouter unlocks the full AI ecosystem** by providing access to 100+ different AI models through a single API. When combined with **Claude Code and CCProxy**, OpenRouter becomes the ultimate platform for AI experimentation, allowing you to find the perfect model for each development task.

## 🎯 Why Choose OpenRouter for Claude Code?

- 🎯 **100+ models**: Access to all major AI models through familiar Claude Code interface
- 🔄 **Model fallbacks**: Automatic failover between models for maximum reliability
- 📊 **Usage analytics**: Detailed tracking and monitoring integrated with CCProxy
- 💰 **Cost optimization**: Compare and choose the most cost-effective models for each task
- 🧪 **Perfect for experimentation**: Test different models without changing your Claude Code workflow
- ⚡ **Unified API**: Single integration for Claude 3.5, GPT-4, Llama, Mistral, and more

## Setup

### 1. Get an API Key

1. Visit [openrouter.ai](https://openrouter.ai)
2. Sign up for an account
3. Go to the API Keys section
4. Generate a new API key

### 2. Configure CCProxy

Set the following environment variables:

```bash
export PROVIDER=openrouter
export OPENROUTER_API_KEY=sk-or-v1-your_openrouter_api_key_here
```

### 3. Optional Configuration

```bash
# Custom model (default: anthropic/claude-3.5-sonnet)
export OPENROUTER_MODEL=openai/gpt-4o

# Custom max tokens (default: 16384)
export OPENROUTER_MAX_TOKENS=8192

# Custom base URL (default: https://openrouter.ai/api/v1)
export OPENROUTER_BASE_URL=https://openrouter.ai/api/v1

# Site URL for tracking (optional)
export OPENROUTER_SITE_URL=https://yourapp.com

# Site name for tracking (optional)
export OPENROUTER_SITE_NAME="Your App Name"
```

## Available Models

OpenRouter offers 100+ models across different categories:

### Top Claude Models
| Model | Context | Speed | Cost/1M tokens |
|-------|---------|-------|----------------|
| **anthropic/claude-3.5-sonnet** | 200K | ⚡⚡ | $3.00/$15.00 |
| **anthropic/claude-3-opus** | 200K | ⚡ | $15.00/$75.00 |
| **anthropic/claude-3-haiku** | 200K | ⚡⚡⚡ | $0.25/$1.25 |

### Top OpenAI Models
| Model | Context | Speed | Cost/1M tokens |
|-------|---------|-------|----------------|
| **openai/gpt-4o** | 128K | ⚡⚡ | $2.50/$10.00 |
| **openai/gpt-4o-mini** | 128K | ⚡⚡⚡ | $0.15/$0.60 |
| **openai/o1-preview** | 32K | ⚡ | $15.00/$60.00 |

### Top Open Source Models
| Model | Context | Speed | Cost/1M tokens |
|-------|---------|-------|----------------|
| **meta-llama/llama-3.1-405b-instruct** | 32K | ⚡ | $2.70/$2.70 |
| **qwen/qwen-2.5-72b-instruct** | 32K | ⚡⚡ | $0.40/$1.20 |
| **mistralai/mistral-large** | 128K | ⚡⚡ | $2.00/$6.00 |

### Specialized Models
| Model | Specialty | Context | Cost/1M tokens |
|-------|-----------|---------|----------------|
| **deepseek/deepseek-coder** | Coding | 128K | $0.14/$0.28 |
| **perplexity/llama-3.1-sonar-large-128k-online** | Web Search | 128K | $5.00/$5.00 |
| **anthropic/claude-3-sonnet:beta** | Latest Beta | 200K | $3.00/$15.00 |

## Pricing

OpenRouter offers competitive pricing with transparent costs:

### Free Tier
- $5 free credits for new users
- No monthly fees
- Pay-as-you-use pricing

### Cost Structure
- **Input tokens**: $0.06 - $15.00 per 1M tokens
- **Output tokens**: $0.06 - $75.00 per 1M tokens
- **No minimum spend**
- **Volume discounts** available

## Configuration Examples

### Basic Setup

```bash
# .env file
PROVIDER=openrouter
OPENROUTER_API_KEY=sk-or-v1-your_api_key_here
```

### High-Performance Setup

```bash
# For speed-focused applications
PROVIDER=openrouter
OPENROUTER_API_KEY=sk-or-v1-your_api_key_here
OPENROUTER_MODEL=openai/gpt-4o-mini
OPENROUTER_MAX_TOKENS=4096
```

### Quality-Focused Setup

```bash
# For best quality responses
PROVIDER=openrouter
OPENROUTER_API_KEY=sk-or-v1-your_api_key_here
OPENROUTER_MODEL=anthropic/claude-3-opus
OPENROUTER_MAX_TOKENS=16384
```

### Cost-Optimized Setup

```bash
# For cost-effective usage
PROVIDER=openrouter
OPENROUTER_API_KEY=sk-or-v1-your_api_key_here
OPENROUTER_MODEL=qwen/qwen-2.5-72b-instruct
OPENROUTER_MAX_TOKENS=8192
```

## Usage with Claude Code

Once configured, use Claude Code normally:

```bash
# Set CCProxy as the API endpoint
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED

# Use Claude Code
claude "Compare different sorting algorithms"
```

## Features

### ✅ Supported
- Text generation
- Function calling
- Tool use
- Streaming responses
- Vision capabilities (model dependent)
- JSON mode (model dependent)
- Custom temperature/top_p
- Model fallbacks
- Usage tracking

### ⚠️ Model Dependent
- Vision/image input
- Real-time data access
- Code execution
- File uploads

## Advanced Features

### Model Fallbacks

Configure automatic failover between models:

```bash
# Set primary and fallback models
export OPENROUTER_MODEL=anthropic/claude-3.5-sonnet
export OPENROUTER_FALLBACK=openai/gpt-4o
```

### Usage Tracking

OpenRouter provides detailed analytics:

```bash
# Add tracking headers
export OPENROUTER_SITE_URL=https://yourapp.com
export OPENROUTER_SITE_NAME="Your App Name"
```

### Custom Headers

```bash
# Add custom tracking
export OPENROUTER_X_TITLE="Your Request Title"
```

## Performance Tips

### 1. Choose the Right Model

```bash
# For speed: Use smaller, faster models
export OPENROUTER_MODEL=openai/gpt-4o-mini

# For quality: Use larger, more capable models
export OPENROUTER_MODEL=anthropic/claude-3-opus

# For cost: Use open-source models
export OPENROUTER_MODEL=qwen/qwen-2.5-72b-instruct
```

### 2. Optimize Token Usage

```bash
# Reduce max tokens for faster responses
export OPENROUTER_MAX_TOKENS=1024

# Use appropriate context length
export OPENROUTER_MAX_TOKENS=4096
```

### 3. Monitor Usage

Check your usage and costs:

```bash
# View OpenRouter dashboard
curl -H "Authorization: Bearer $OPENROUTER_API_KEY" \
  https://openrouter.ai/api/v1/auth/key
```

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

**Solution**: OpenRouter has generous rate limits. Wait and retry, or check your usage.

### Authentication Errors

```json
{
  "error": {
    "message": "Invalid API key",
    "type": "authentication_error"
  }
}
```

**Solution**: Verify your API key is correct and has sufficient credits.

### Model Not Available

```json
{
  "error": {
    "message": "Model not found or not available",
    "type": "invalid_request_error"
  }
}
```

**Solution**: Check the [OpenRouter models page](https://openrouter.ai/models) for available models.

### Insufficient Credits

```json
{
  "error": {
    "message": "Insufficient credits",
    "type": "insufficient_quota"
  }
}
```

**Solution**: Add credits to your OpenRouter account.

## Model Selection Guide

### For General Use
- `anthropic/claude-3.5-sonnet` - Best all-around performance
- `openai/gpt-4o` - Strong reasoning and tool use

### For Speed
- `openai/gpt-4o-mini` - Fast and cost-effective
- `anthropic/claude-3-haiku` - Ultra-fast responses

### For Quality
- `anthropic/claude-3-opus` - Highest quality responses
- `openai/o1-preview` - Advanced reasoning

### For Cost
- `qwen/qwen-2.5-72b-instruct` - Great quality-to-cost ratio
- `meta-llama/llama-3.1-405b-instruct` - Open source powerhouse

### For Coding
- `deepseek/deepseek-coder` - Specialized for code
- `anthropic/claude-3.5-sonnet` - Excellent code understanding

## Monitoring

Monitor your OpenRouter usage:

```bash
# Check CCProxy logs
tail -f ccproxy.log

# Check OpenRouter status
curl http://localhost:7187/status

# View usage analytics on OpenRouter dashboard
```

## Next Steps

- Explore [other providers](/providers/) for comparison and specialized use cases
- Learn about [model fallbacks](/guide/fallbacks) for production reliability
- Set up [usage monitoring](/guide/monitoring) to optimize model selection and costs
- Try [Groq with Kimi K2](/providers/groq) for ultra-fast inference alongside OpenRouter

<script>
function shareToTwitter() {
  const url = encodeURIComponent(window.location.href);
  const text = encodeURIComponent('🎯 OpenRouter + Claude Code + CCProxy = Access to 100+ AI models! Claude 3.5, GPT-4, Llama, Mistral, and more through unified API');
  window.open(`https://twitter.com/intent/tweet?url=${url}&text=${text}`, '_blank');
}

function shareToLinkedIn() {
  const url = encodeURIComponent(window.location.href);
  window.open(`https://www.linkedin.com/sharing/share-offsite/?url=${url}`, '_blank');
}

function shareToReddit() {
  const url = encodeURIComponent(window.location.href);
  const title = encodeURIComponent('OpenRouter with Claude Code - 100+ AI Models via CCProxy');
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