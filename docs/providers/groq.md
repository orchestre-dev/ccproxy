---
title: Groq with Claude Code - Ultra-Fast AI Inference with Kimi K2 Support
description: Experience lightning-fast AI development with Groq LPU infrastructure through CCProxy. Access Kimi K2, Llama 3.1, and Mixtral models with Claude Code for sub-second inference speeds.
keywords: Groq, Claude Code, CCProxy, Kimi K2, fast AI inference, LPU, Llama 3.1, Mixtral, ultra-fast AI, AI proxy, sub-second responses
---

# Groq Provider

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

**Groq revolutionizes AI development** with ultra-fast inference speeds through their groundbreaking LPU (Language Processing Unit) technology. When combined with **Claude Code and CCProxy**, Groq delivers an unmatched development experience with **sub-second response times** and access to cutting-edge models like **Kimi K2**.

## 🚀 Why Choose Groq for Claude Code?

- ⚡ **Ultra-fast inference**: Sub-second response times with LPU technology
- 🌟 **Kimi K2 access**: Revolutionary Moonshot AI model with 8K context
- 💰 **Cost-effective**: Competitive pricing with generous free tier
- 🎯 **Simple API**: Easy integration with Claude Code via CCProxy
- 🔄 **High throughput**: Excellent for high-volume AI development workflows
- 🚀 **Perfect Claude Code integration**: Seamless compatibility with zero configuration changes

## Setup

### 1. Get an API Key

1. Visit [console.groq.com](https://console.groq.com)
2. Sign up for a free account
3. Navigate to the API Keys section
4. Generate a new API key

### 2. Configure CCProxy

Set the following environment variables:

```bash
export PROVIDER=groq
export GROQ_API_KEY=gsk_your_groq_api_key_here
```

### 3. Optional Configuration

```bash
# Custom model (default: moonshotai/kimi-k2-instruct)
export GROQ_MODEL=llama-3.1-70b-versatile

# Custom max tokens (default: 16384)
export GROQ_MAX_TOKENS=8192

# Custom base URL (default: https://api.groq.com/openai/v1)
export GROQ_BASE_URL=https://api.groq.com/openai/v1
```

## Available Models

| Model | Context | Speed | Best For |
|-------|---------|-------|----------|
| **moonshotai/kimi-k2-instruct** | 8K | ⚡⚡⚡ | **Claude Code optimized (default)** |
| **llama-3.1-405b-reasoning** | 128K | ⚡⚡ | Complex reasoning |
| **llama-3.1-70b-versatile** | 128K | ⚡⚡⚡ | Balanced performance |
| **llama-3.1-8b-instant** | 128K | ⚡⚡⚡ | Fast responses |
| **mixtral-8x7b-32768** | 32K | ⚡⚡⚡ | Code and math |
| **gemma2-9b-it** | 8K | ⚡⚡⚡ | Efficient inference |

## 🌟 Featured: Kimi K2 + Claude Code Integration

**Kimi K2** (`moonshotai/kimi-k2-instruct`) represents the **perfect AI model for Claude Code users**. This revolutionary model from Moonshot AI, optimized through Groq's LPU infrastructure, delivers:

- **Sub-second inference**: Experience the fastest AI responses for development
- **8K context window**: Perfect for code analysis and documentation
- **Claude Code optimized**: Seamlessly integrates with your existing workflow
- **Exceptional value**: At $0.20/1M input tokens, it's 50x cheaper than GPT-4
- **Zero configuration**: Works instantly with CCProxy - just set environment variables

### Kimi K2 Performance Advantages

```mermaid
graph TB
    A[Kimi K2 + Claude Code] --> B[Sub-second Responses]
    A --> C[8K Context for Code]
    A --> D[50x Cheaper than GPT-4]
    A --> E[Perfect Claude Integration]
    B --> F[10x Faster Development]
    C --> G[Complete File Analysis]
    D --> H[Cost-Effective Scaling]
    E --> I[Zero Setup Friction]
```

## Pricing

Groq offers competitive pricing with a generous free tier:

| Tier | Requests/Day | Requests/Minute | Tokens/Minute |
|------|--------------|-----------------|---------------|
| **Free** | 14,400 | 30 | 6,000 |
| **Pay-as-you-go** | Unlimited | 30 | 6,000 |
| **Pro** | Unlimited | 6,000 | 6,000,000 |

### Cost per 1M Tokens

| Model | Input | Output |
|-------|-------|--------|
| Llama 3.1 8B | $0.05 | $0.08 |
| Llama 3.1 70B | $0.59 | $0.79 |
| Llama 3.1 405B | $2.70 | $2.70 |
| Mixtral 8x7B | $0.24 | $0.24 |

## Configuration Examples

### Basic Setup

```bash
# .env file
PROVIDER=groq
GROQ_API_KEY=gsk_your_api_key_here
```

### High-Performance Setup

```bash
# For maximum speed
PROVIDER=groq
GROQ_API_KEY=gsk_your_api_key_here
GROQ_MODEL=llama-3.1-8b-instant
GROQ_MAX_TOKENS=4096
```

### Quality-Focused Setup

```bash
# For best quality
PROVIDER=groq
GROQ_API_KEY=gsk_your_api_key_here
GROQ_MODEL=llama-3.1-405b-reasoning
GROQ_MAX_TOKENS=16384
```

## Usage with Claude Code

Once configured, use Claude Code normally:

```bash
# Set CCProxy as the API endpoint
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED

# Use Claude Code
claude "Explain quantum computing in simple terms"
```

## Features

### ✅ Supported
- Text generation
- Function calling
- Tool use
- Streaming responses
- JSON mode
- Custom temperature/top_p

### ❌ Not Supported
- Vision/image input
- File uploads
- Real-time data access

## Performance Tips

### 1. Choose the Right Model
- Use `llama-3.1-8b-instant` for speed
- Use `llama-3.1-405b-reasoning` for quality
- Use `mixtral-8x7b-32768` for code tasks

### 2. Optimize Token Usage
```bash
# Reduce max tokens for faster responses
export GROQ_MAX_TOKENS=1024
```

### 3. Batch Requests
If making multiple requests, consider batching them to stay within rate limits.

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

**Solution**: Wait and retry, or upgrade to a paid plan for higher limits.

### Authentication Errors
```json
{
  "error": {
    "message": "Invalid API key",
    "type": "authentication_error"
  }
}
```

**Solution**: Verify your API key is correct and hasn't expired.

### Model Not Found
```json
{
  "error": {
    "message": "Model not found",
    "type": "invalid_request_error"
  }
}
```

**Solution**: Check the model name against the available models list above.

## Advanced Configuration

### Custom Headers
```bash
# Add custom headers for tracking
export GROQ_SITE_URL=https://yourapp.com
export GROQ_SITE_NAME="Your App Name"
```

### Timeout Settings
```bash
# Increase timeout for long requests
export GROQ_TIMEOUT=120s
```

## Monitoring

Check Groq usage and performance:

```bash
# View CCProxy logs
tail -f ccproxy.log

# Check status endpoint
curl http://localhost:7187/status
```

## Next Steps

- Experience [Kimi K2 with Claude Code](/kimi-k2) for cutting-edge AI development
- Explore [other providers](/providers/) for different use cases
- Learn about [configuration options](/guide/configuration)
- Set up [health monitoring](/guide/health-checks) for production

<script>
function shareToTwitter() {
  const url = encodeURIComponent(window.location.href);
  const text = encodeURIComponent('🚀 Groq + Claude Code + Kimi K2 = Ultra-fast AI development! Experience sub-second inference with CCProxy');
  window.open(`https://twitter.com/intent/tweet?url=${url}&text=${text}`, '_blank');
}

function shareToLinkedIn() {
  const url = encodeURIComponent(window.location.href);
  window.open(`https://www.linkedin.com/sharing/share-offsite/?url=${url}`, '_blank');
}

function shareToReddit() {
  const url = encodeURIComponent(window.location.href);
  const title = encodeURIComponent('Groq with Claude Code - Ultra-Fast AI Inference with Kimi K2 Support');
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