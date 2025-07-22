---
title: OpenRouter with Claude Code - 100+ AI Models via CCProxy
description: Access 100+ AI models including Qwen3 235B (AIME25 70.3), Kimi K2, Grok, Claude, and GPT-4 with Claude Code through CCProxy and OpenRouter. Features the top-performing Qwen3 235B model with record-breaking benchmarks.
keywords: OpenRouter, Claude Code, CCProxy, Qwen3 235B, AIME25, Kimi K2, Grok, 100+ AI models, model fallbacks, Claude 3.5, GPT-4, Llama, model comparison, unified AI API, top performance, benchmark
---

# OpenRouter Provider

**OpenRouter unlocks the full AI ecosystem** by providing access to 100+ different AI models through a single API, including the **top-performing Qwen3 235B** with its record-breaking AIME25 score of 70.3. As a **major supporter of cutting-edge models**, OpenRouter offers reliable access to Qwen3 235B, Kimi K2's ultra-fast inference, Grok's real-time data capabilities, alongside Claude, GPT-4, and many others. When combined with **Claude Code and CCProxy**, OpenRouter becomes the ultimate platform for AI experimentation, allowing you to leverage the most advanced models for each development task.

<SocialShare />

## 🎯 Why Choose OpenRouter for Claude Code?

- 🏆 **Top performance**: Access to Qwen3 235B with AIME25 score of 70.3 (vs GPT-4o's 26.7)
- ⚡ **Ultra-fast options**: Kimi K2 with 128K context for lightning-fast inference
- 🌐 **Real-time data**: Grok models with current web information access
- 🎯 **100+ models**: Access to all major AI models through familiar Claude Code interface
- 🔄 **Model fallbacks**: Automatic failover between models for maximum reliability
- 📊 **Usage analytics**: Detailed tracking and monitoring integrated with CCProxy
- 💰 **Cost optimization**: Compare and choose the most cost-effective models for each task
- 🧪 **Perfect for experimentation**: Test different models without changing your Claude Code workflow
- ⚡ **Unified API**: Single integration for Qwen3, Kimi K2, Grok, Claude 3.5, GPT-4, and more

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
export OPENROUTER_MODEL=moonshotai/kimi-k2-instruct

# Custom max tokens (default: 16384)
export OPENROUTER_MAX_TOKENS=8192

# Custom base URL (default: https://openrouter.ai/api/v1)
export OPENROUTER_BASE_URL=https://openrouter.ai/api/v1

# Site URL for tracking (optional)
export OPENROUTER_SITE_URL=https://yourapp.com

# Site name for tracking (optional)
export OPENROUTER_SITE_NAME="Your App Name"
```

## 🌟 Top Models Available

### Qwen3 235B A22B 2507 - Leading Performance

**The current top-performing model on OpenRouter**, Qwen3 235B delivers exceptional capabilities:

- **🏆 AIME25 Score: 70.3** - Crushing GPT-4o's 26.7 score
- **💡 Advanced reasoning**: State-of-the-art mathematical and logical capabilities
- **📊 Massive scale**: 235B parameters with optimized inference
- **🚀 Production ready**: Reliable performance for demanding applications

### Kimi K2 - Ultra-Fast Inference

- **⚡ Lightning fast**: Optimized for rapid responses
- **📄 128K context**: Handle large documents and codebases
- **🎯 Tool calling**: Full Claude Code compatibility
- **💰 Cost effective**: Great performance-to-price ratio

### Grok Models - Real-Time Data

- **🌐 Real-time access**: Current information and web data
- **🔄 Dynamic updates**: Always up-to-date responses
- **🛠️ Tool support**: Compatible with Claude Code workflows
- **📈 Continuous learning**: Incorporates latest information

## Available Models

OpenRouter provides access to 100+ AI models from leading providers including:

- **Qwen** - Top-performing Qwen3 235B with record-breaking benchmarks
- **Anthropic** - Claude series with advanced reasoning
- **OpenAI** - GPT-4 series and reasoning models
- **Moonshot AI** - Kimi K2 with ultra-fast inference
- **xAI** - Grok models with real-time data access
- **Meta** - Llama models for open-source applications  
- **Google** - Gemini models with multimodal capabilities
- **Mistral** - European privacy-focused models
- **Many others** - Including specialized coding and reasoning models

**🔧 Critical for Claude Code**: You must select models that support **tool calling** or **function calling** capabilities, as Claude Code requires these features to operate correctly.

### Model Selection Guidelines

When choosing models on OpenRouter:

1. **Verify Tool Support**: Ensure the model supports function calling
2. **Check Availability**: Model availability changes frequently
3. **Review Capabilities**: Different models excel at different tasks
4. **Consider Cost**: Pricing varies significantly between models

For current model availability, capabilities, and pricing, visit [OpenRouter's official models page](https://openrouter.ai/models).

## Pricing

OpenRouter offers competitive pricing with transparent costs:

### Free Tier
- Free credits for new users
- No monthly fees  
- Pay-as-you-use pricing

### Cost Structure
- Competitive per-token pricing across all models
- No minimum spend required
- Volume discounts available

For current, accurate pricing information, visit [OpenRouter's official pricing page](https://openrouter.ai/models).

## Configuration Examples

### Basic Setup

```bash
# .env file
PROVIDER=openrouter
OPENROUTER_API_KEY=sk-or-v1-your_api_key_here
```

### Top Performance Setup - Qwen3 235B

```bash
# For the highest performing model (AIME25: 70.3)
PROVIDER=openrouter
OPENROUTER_API_KEY=sk-or-v1-your_api_key_here
OPENROUTER_MODEL=qwen/qwen-3-235b-a22b-2507
OPENROUTER_MAX_TOKENS=16384
```

### Ultra-Fast Setup - Kimi K2

```bash
# For ultra-fast inference with large context
PROVIDER=openrouter
OPENROUTER_API_KEY=sk-or-v1-your_api_key_here
OPENROUTER_MODEL=moonshotai/kimi-k2-instruct
OPENROUTER_MAX_TOKENS=8192
```

### Real-Time Data Setup - Grok

```bash
# For real-time information access
PROVIDER=openrouter
OPENROUTER_API_KEY=sk-or-v1-your_api_key_here
OPENROUTER_MODEL=xai/grok-beta
OPENROUTER_MAX_TOKENS=8192
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
export ANTHROPIC_BASE_URL=http://localhost:3456
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

### Model Routing Examples

Configure intelligent model routing based on task requirements:

```bash
# For maximum performance tasks
export OPENROUTER_MODEL=qwen/qwen-3-235b-a22b-2507
export OPENROUTER_FALLBACK=anthropic/claude-3-opus

# For speed-critical applications  
export OPENROUTER_MODEL=moonshotai/kimi-k2-instruct
export OPENROUTER_FALLBACK=openai/gpt-4o-mini

# For real-time information needs
export OPENROUTER_MODEL=xai/grok-beta
export OPENROUTER_FALLBACK=xai/grok-2
```

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
# For maximum performance: Use the top-scoring model
export OPENROUTER_MODEL=qwen/qwen-3-235b-a22b-2507

# For speed: Use ultra-fast models
export OPENROUTER_MODEL=moonshotai/kimi-k2-instruct

# For quality: Use larger, more capable models
export OPENROUTER_MODEL=anthropic/claude-3-opus

# For real-time data: Use models with web access
export OPENROUTER_MODEL=xai/grok-beta

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

### For Top Performance
- `qwen/qwen-3-235b-a22b-2507` - **#1 Overall** - AIME25 score 70.3 (vs GPT-4o's 26.7)
- `anthropic/claude-3-opus` - Highest quality responses
- `openai/o1-preview` - Advanced reasoning

### For Ultra-Fast Speed
- `moonshotai/kimi-k2-instruct` - **Fastest inference** with 128K context
- `openai/gpt-4o-mini` - Fast and cost-effective
- `anthropic/claude-3-haiku` - Ultra-fast responses

### For Real-Time Data
- `xai/grok-beta` - **Real-time web access** and current information
- `xai/grok-2` - Enhanced reasoning with real-time data

### For General Use
- `anthropic/claude-3.5-sonnet` - Best all-around performance
- `openai/gpt-4o` - Strong reasoning and tool use

### For Cost
- `qwen/qwen-2.5-72b-instruct` - Great quality-to-cost ratio
- `meta-llama/llama-3.1-405b-instruct` - Open source powerhouse

### For Coding
- `deepseek/deepseek-coder` - Specialized for code
- `anthropic/claude-3.5-sonnet` - Excellent code understanding
- `qwen/qwen-3-235b-a22b-2507` - Superior problem solving

## Monitoring

Monitor your OpenRouter usage:

```bash
# Check CCProxy logs
tail -f ccproxy.log

# Check OpenRouter status
curl http://localhost:3456/status

# View usage analytics on OpenRouter dashboard
```

## Next Steps

- Explore [other providers](/providers/) for comparison and specialized use cases
- Learn about model fallbacks for production reliability
- Set up [usage monitoring](/guide/monitoring) to optimize model selection and costs
- Try [Groq with Kimi K2](/providers/groq) for ultra-fast inference alongside OpenRouter