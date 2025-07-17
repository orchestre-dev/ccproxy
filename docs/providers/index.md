# Supported Providers

CCProxy supports 7 major AI providers, each with unique strengths and characteristics. This page provides an overview of all supported providers.

## Provider Overview

| Provider | Speed | Cost | Models | Use Case |
|----------|-------|------|--------|----------|
| **[Groq](/providers/groq)** | ⚡⚡⚡ | 💰 | 15+ | Ultra-fast inference |
| **[OpenRouter](/providers/openrouter)** | ⚡⚡ | 💰💰 | 100+ | Model diversity |
| **[OpenAI](/providers/openai)** | ⚡⚡ | 💰💰💰 | 10+ | Industry standard |
| **[XAI (Grok)](/providers/xai)** | ⚡⚡ | 💰💰 | 3+ | Real-time data |
| **[Google Gemini](/providers/gemini)** | ⚡⚡ | 💰💰 | 5+ | Multimodal AI |
| **[Mistral AI](/providers/mistral)** | ⚡⚡ | 💰💰 | 8+ | European choice |
| **[Ollama](/providers/ollama)** | ⚡ | 🆓 | 50+ | Local & private |

## Quick Setup

Each provider requires different setup steps. Here's the minimal configuration for each:

::: code-group

```bash [Groq]
export PROVIDER=groq
export GROQ_API_KEY=gsk_your_key_here
```

```bash [OpenRouter]
export PROVIDER=openrouter
export OPENROUTER_API_KEY=sk-or-v1-your_key_here
```

```bash [OpenAI]
export PROVIDER=openai
export OPENAI_API_KEY=sk-your_key_here
```

```bash [XAI (Grok)]
export PROVIDER=xai
export XAI_API_KEY=xai-your_key_here
```

```bash [Google Gemini]
export PROVIDER=gemini
export GEMINI_API_KEY=your_key_here
```

```bash [Mistral AI]
export PROVIDER=mistral
export MISTRAL_API_KEY=your_key_here
```

```bash [Ollama]
export PROVIDER=ollama
export OLLAMA_MODEL=llama3.2
# Requires Ollama running locally
```

:::

## Provider Comparison

### Performance & Speed

```mermaid
graph TB
    A[Ultra Fast<br/>< 100ms] --> B[Groq]
    C[Fast<br/>< 500ms] --> D[OpenRouter]
    C --> E[OpenAI]
    C --> F[XAI]
    C --> G[Gemini]
    C --> H[Mistral]
    I[Variable<br/>depends on model] --> J[Ollama]
```

### Cost Comparison

| Provider | Free Tier | Pricing Model | Best For |
|----------|-----------|---------------|----------|
| **Groq** | ✅ Generous | Pay-per-use | Speed & development |
| **OpenRouter** | ✅ Limited | Pay-per-use | Model variety |
| **OpenAI** | ✅ Limited | Pay-per-use | Enterprise reliability |
| **XAI** | ❌ | Pay-per-use | Real-time data |
| **Gemini** | ✅ Generous | Pay-per-use | Multimodal tasks |
| **Mistral** | ❌ | Pay-per-use | European compliance |
| **Ollama** | ✅ Unlimited | Local/Free | Privacy & control |

For current pricing information, visit each provider's official pricing page.

### Model Capabilities

| Provider | Text | Code | Function Calls | Vision | Reasoning |
|----------|------|------|---------------|--------|-----------|
| **Groq** | ✅ | ✅ | ✅ | ❌ | ✅ |
| **OpenRouter** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **OpenAI** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **XAI** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Gemini** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Mistral** | ✅ | ✅ | ✅ | ❌ | ✅ |
| **Ollama** | ✅ | ✅ | ✅ | ✅* | ✅ |

*Depends on specific model

## Important: Tool Calling Requirement

**⚠️ Critical for Claude Code Users**: You must select models that support **tool calling** or **function calling** capabilities, as Claude Code requires these features to operate correctly. When choosing models from any provider, verify they support function calling.

## Provider Selection Guide

### Choose **Groq** if you want:
- ⚡ Fastest possible inference speeds
- 💰 Cost-effective pricing
- 🆓 Generous free tier
- 📊 Simple, reliable service

### Choose **OpenRouter** if you want:
- 🎯 Access to 100+ different models
- 🔄 Model routing and fallbacks
- 🧪 Experimentation with cutting-edge models
- 📈 Usage analytics

### Choose **OpenAI** if you want:
- 🏭 Industry-standard models
- 🛠️ Extensive tooling and ecosystem
- 👁️ Advanced vision capabilities
- 🎯 Proven reliability

### Choose **XAI (Grok)** if you want:
- 📰 Real-time information access
- 🐦 X/Twitter integration
- 🆕 Cutting-edge capabilities
- 🚀 Elon Musk's AI vision

### Choose **Google Gemini** if you want:
- 🎥 Advanced multimodal capabilities
- 🏗️ Google's latest technology
- 📊 Strong analytical capabilities
- 🔍 Integration with Google services

### Choose **Mistral AI** if you want:
- 🇪🇺 European AI alternative
- 🔒 Strong privacy focus
- 💼 Enterprise-grade features
- 🎯 Multilingual excellence

### Choose **Ollama** if you want:
- 🔒 Complete privacy (local processing)
- 🌐 Offline capabilities
- 💸 Zero ongoing costs
- 🎛️ Full control over your models

## Model Selection Guidelines

When selecting models from any provider:

1. **Verify Tool Calling Support**: Ensure the model supports function calling/tool use
2. **Check Current Availability**: Model availability changes frequently
3. **Review Pricing**: Visit the provider's official pricing page for current rates
4. **Test Performance**: Different models excel at different tasks

For current model lists, capabilities, and pricing, always check the provider's official documentation.

## Getting Started

1. **Pick a provider** based on your needs
2. **Get an API key** from the provider's website
3. **Configure CCProxy** with your chosen provider
4. **Start coding** with Claude Code!

Ready to dive deeper? Click on any provider above to see detailed setup instructions, model lists, and configuration options.