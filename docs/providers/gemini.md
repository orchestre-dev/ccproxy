---
title: Google Gemini with Claude Code - Advanced Multimodal AI via CCProxy
description: Access Google's Gemini Pro, Flash, and Lite models with Claude Code through CCProxy. Experience multimodal AI with 2M context windows and excellent vision capabilities.
keywords: Google Gemini, Claude Code, CCProxy, Gemini 2.5 Pro, multimodal AI, long context, vision AI, Google DeepMind, AI proxy, 2M context window
---

# Google Gemini Provider

**Google's Gemini models** provide cutting-edge multimodal AI capabilities. When integrated with **Claude Code through CCProxy**, Gemini offers **long-context understanding** with up to 2M tokens and sophisticated multimodal processing for vision and text tasks.

<SocialShare />

## 🎥 Why Choose Google Gemini for Claude Code?

- 🎯 **Multimodal excellence**: Superior vision and text understanding with Claude Code integration
- 🏗️ **Google's latest tech**: Cutting-edge AI from Google DeepMind accessible via CCProxy
- 📊 **Massive context windows**: Up to 2M tokens for comprehensive code and document analysis
- 💰 **Flexible pricing tiers**: From ultra-fast lite to pro models
- 🔍 **Superior analytics**: Outstanding at data analysis and complex reasoning tasks
- ⚡ **Claude Code optimized**: Seamless integration with intelligent routing
- ⚠️ **Limited function calling**: Basic tool support - may have compatibility issues with some Claude Code features

## Setup

### 1. Get an API Key

1. Visit [aistudio.google.com](https://aistudio.google.com)
2. Sign up with your Google account
3. Navigate to "Get API key"
4. Generate a new API key

### 2. Configure CCProxy

Set the following environment variables:

```bash
export PROVIDER=gemini
export GEMINI_API_KEY=your_gemini_api_key_here
```

Alternative API key environment variables:
```bash
# Also supported
export GOOGLE_API_KEY=your_gemini_api_key_here
```

### 3. Optional Configuration

```bash
# Custom model (default: gemini-2.5-flash)
export GEMINI_MODEL=gemini-2.5-pro

# Custom max tokens (default: 16384)
export GEMINI_MAX_TOKENS=8192

# Custom base URL (default: https://generativelanguage.googleapis.com)
export GEMINI_BASE_URL=https://generativelanguage.googleapis.com
```

## Available Models

### Latest Models (July 2025)

Google's newest Gemini models offer state-of-the-art performance:

- **gemini-2.5-pro** - Top-tier model
  - Excels at complex reasoning and analysis
  - Best for tasks requiring deep understanding
  
- **gemini-2.5-flash** - Balanced performance model
  - Optimal balance of speed and quality
  - Ideal default model for most tasks
  
- **gemini-2.5-flash-lite** - Preview model, lowest latency/cost
  - Maximum speed for simple tasks
  - Extremely cost-effective
  
- **gemini-2.0-flash** - Generally available with 1M token context
  - Stable, production-ready model
  - Excellent for long-context tasks

### Legacy Models

- **Gemini 1.5 Series** - Previous generation with long context windows
- **Gemini 1.0 Series** - Stable models for general use

**🔧 Critical for Claude Code**: You must select models that support **tool calling** or **function calling** capabilities, as Claude Code requires these features to operate correctly.

⚠️ **Important**: The Gemini transformer in CCProxy has limited tool support. While basic function calling works, complex tool interactions may fail. The transformer:
- Maps tool definitions to Gemini's function format
- Does not support provider-specific features like `thinkingBudget`
- May have compatibility issues with complex Claude Code operations

For full Claude Code compatibility, consider using Anthropic or OpenAI providers instead.

### Supported Parameters

CCProxy supports the following standard parameters for Gemini models:

- **temperature**: Controls randomness (0.0 to 1.0)
- **top_p**: Nucleus sampling parameter (0.0 to 1.0)
- **top_k**: Top-k sampling parameter (integer)
- **max_tokens**: Maximum response length

### Model Selection Guidelines

When choosing Gemini models:

1. **Verify Tool Support**: Ensure the model supports function calling
2. **Check Current Availability**: Google's model lineup evolves frequently
3. **Consider Context Needs**: Gemini offers very long context windows (up to 2M tokens)
4. **Review Multimodal Needs**: Some models excel at vision and document analysis
5. **Test Performance**: Different models balance speed vs quality differently

For current model availability, capabilities, and pricing, visit [Google AI Studio](https://aistudio.google.com).

### Routing Recommendations

CCProxy can automatically route to the optimal Gemini model based on your usage pattern:

- **"default" route**: `gemini-2.5-flash` - Balanced performance
- **"longContext" route**: `gemini-2.5-pro` - Best for complex analysis
- **"background" route**: `gemini-2.5-flash-lite` - Fastest, cheapest for simple tasks

Configure routing in your CCProxy config:

```json
{
  "routing": {
    "default": "gemini-2.5-flash",
    "longContext": "gemini-2.5-pro",
    "background": "gemini-2.5-flash-lite"
  }
}
```

## Pricing

### Free Tier
Google AI Studio offers a generous free tier:
- High request limits for development and testing
- Generous daily token allowances
- Perfect for getting started

### Paid Usage
- Competitive per-token pricing
- Pay-as-you-use model
- Volume discounts available

For current, accurate pricing information, visit [Google AI Studio](https://aistudio.google.com).

## Configuration Examples

### Basic Setup

```bash
# .env file
PROVIDER=gemini
GEMINI_API_KEY=your_api_key_here
```

### High-Performance Setup

```bash
# For maximum speed
PROVIDER=gemini
GEMINI_API_KEY=your_api_key_here
GEMINI_MODEL=gemini-2.5-flash-lite
GEMINI_MAX_TOKENS=4096
```

### Quality-Focused Setup

```bash
# For best quality and long context
PROVIDER=gemini
GEMINI_API_KEY=your_api_key_here
GEMINI_MODEL=gemini-2.5-pro
GEMINI_MAX_TOKENS=16384
```

## Usage with Claude Code

Once configured, use Claude Code normally:

```bash
# Set CCProxy as the API endpoint
export ANTHROPIC_BASE_URL=http://localhost:3456
# Claude Code will use CCProxy, no direct Anthropic API key needed

# Use Claude Code
claude "Analyze this image and explain what you see"
```

## Features

### ✅ Fully Supported
- Text generation
- Function calling
- Tool use
- Streaming responses
- Vision/image input
- Long context (up to 2M tokens)
- JSON mode
- Custom temperature
- Multimodal understanding

### ⚠️ Model Dependent
- Real-time data access (limited)
- Code execution capabilities
- File uploads (vision models only)

### ❌ Not Supported
- Audio processing
- Video analysis (coming soon)

## Multimodal Capabilities

### Vision Understanding

Gemini excels at vision tasks:

```bash
# Image analysis
claude "What's in this image and what's the context?"

# Document analysis
claude "Extract and summarize the key information from this document"

# Chart and graph analysis
claude "Analyze this chart and explain the trends"
```

### Long Context Processing

With up to 2M tokens of context:

```bash
# Large document analysis
claude "Summarize this entire research paper"

# Multi-document comparison
claude "Compare these three reports and highlight differences"

# Code repository analysis
claude "Analyze this entire codebase and suggest improvements"
```

## Performance Tips

### 1. Choose the Right Model

```bash
# For speed and cost efficiency
export GEMINI_MODEL=gemini-2.5-flash-lite

# For balanced performance
export GEMINI_MODEL=gemini-2.5-flash

# For complex reasoning and analysis
export GEMINI_MODEL=gemini-2.5-pro

# For stable long-context tasks
export GEMINI_MODEL=gemini-2.0-flash
```

### 2. Optimize Token Usage

```bash
# Use appropriate max tokens for your use case
export GEMINI_MAX_TOKENS=2048  # For short responses
export GEMINI_MAX_TOKENS=8192  # For detailed analysis
```

### 3. Leverage Multimodal Features

```bash
# Combine text and image analysis
# Use long context for comprehensive analysis
# Take advantage of the generous free tier
```

## Advanced Features

### Function Calling

Gemini has robust function calling capabilities:

```json
{
  "tools": [
    {
      "name": "get_weather",
      "description": "Get current weather",
      "input_schema": {
        "type": "object",
        "properties": {
          "location": {"type": "string"}
        }
      }
    }
  ]
}
```

### JSON Mode

Force structured JSON responses:

```bash
# Gemini supports JSON mode for structured outputs
# Automatically enabled when using tools
```

### Safety Settings

Gemini includes built-in safety filters:

- Harassment detection
- Hate speech filtering
- Sexually explicit content blocking
- Dangerous content prevention

## Use Cases

### 1. Document Analysis

```bash
# Legal document review
claude "Analyze this contract and highlight key terms"

# Research paper summarization
claude "Summarize the key findings from this research"
```

### 2. Data Analysis

```bash
# Chart analysis
claude "What trends do you see in this sales chart?"

# Statistical analysis
claude "Analyze this dataset and provide insights"
```

### 3. Code Understanding

```bash
# Code review
claude "Review this code and suggest improvements"

# Architecture analysis
claude "Analyze this system architecture and identify potential issues"
```

### 4. Creative Tasks

```bash
# Image-based creativity
claude "Create a story based on this image"

# Multimodal content creation
claude "Write a blog post about this infographic"
```

## Troubleshooting

### Rate Limit Errors

```json
{
  "error": {
    "message": "Quota exceeded",
    "type": "quota_exceeded"
  }
}
```

**Solution**: Wait for quota reset or upgrade to paid usage.

### API Key Errors

```json
{
  "error": {
    "message": "API key not valid",
    "type": "invalid_argument"
  }
}
```

**Solution**: Verify your API key is correct and has proper permissions.

### Model Not Found

```json
{
  "error": {
    "message": "Model not found",
    "type": "not_found"
  }
}
```

**Solution**: Check the available models in [Google AI Studio](https://aistudio.google.com).

### Content Safety Blocks

```json
{
  "error": {
    "message": "The request was blocked by safety filters",
    "type": "safety_error"
  }
}
```

**Solution**: Modify your content to comply with safety guidelines.

### Large Context Limits

```json
{
  "error": {
    "message": "Input too long",
    "type": "invalid_argument"
  }
}
```

**Solution**: Reduce input size or use a model with larger context window.

## Best Practices

### 1. Model Selection

```bash
# Use gemini-2.5-flash for most tasks (balanced)
# Use gemini-2.5-flash-lite for high-volume, simple tasks
# Use gemini-2.5-pro for complex analysis and reasoning
# Use gemini-2.0-flash for stable long-context needs
```

### 2. Context Management

```bash
# Take advantage of long context windows
# Structure large inputs clearly
# Use appropriate chunking for very large documents
```

### 3. Multimodal Usage

```bash
# Combine text and image inputs effectively
# Use vision for document analysis
# Leverage charts and graphs analysis
```

### 4. Cost Optimization

```bash
# Start with the generous free tier
# Monitor usage in Google AI Studio
# Use gemini-2.5-flash-lite for maximum cost efficiency
# Use routing to automatically select cost-effective models
```

## Integration Examples

### Python with Google SDK

```python
import google.generativeai as genai

# Configure to use CCProxy
genai.configure(
    api_key="NOT_NEEDED",
    client_options={"api_endpoint": "http://localhost:3456"}
)

model = genai.GenerativeModel('claude-3-sonnet')  # Maps to Gemini
response = model.generate_content("Explain quantum computing")
```

### Anthropic SDK via CCProxy

```python
import anthropic

client = anthropic.Anthropic(
    api_key="NOT_NEEDED",
    base_url="http://localhost:3456"
)

response = client.messages.create(
    model="claude-3-sonnet",  # Maps to Gemini
    messages=[{"role": "user", "content": "Hello!"}],
    max_tokens=100
)
```

## Monitoring

### Google AI Studio

Monitor usage at [aistudio.google.com](https://aistudio.google.com):

- Request counts and quotas
- Token usage
- Model performance
- Error rates

### CCProxy Monitoring

```bash
# Real-time logs
tail -f ccproxy.log | grep gemini

# Status endpoint
curl http://localhost:3456/status

# Health check with Gemini status
curl http://localhost:3456/health
```

## Comparison with Other Providers

### Strengths
- 🎯 Excellent multimodal capabilities
- 💰 Generous free tier
- 📏 Very long context windows (up to 2M tokens)
- 🔍 Strong at analysis tasks
- ⚡ Flexible performance options (lite to pro)
- 🛠️ Robust function calling support

### Considerations
- 🛡️ Strong safety filters (may block some content)
- 🚀 Newer ecosystem (fewer third-party tools)
- 🌐 Geographic availability varies

## Future Developments

Google is rapidly improving Gemini:

- Video understanding capabilities
- Enhanced reasoning models
- Better code generation
- Improved multimodal features

Stay updated at [ai.google.dev](https://ai.google.dev) and [aistudio.google.com](https://aistudio.google.com).

## Next Steps

- Explore long context use cases with Gemini's 2M token windows
- Learn about multimodal capabilities for vision and document analysis
- Set up [usage monitoring](/guide/monitoring) to optimize your Gemini usage
- Compare with [other providers](/providers/) including [Groq for speed](/providers/groq) and [OpenAI for reliability](/providers/openai)