---
layout: home

hero:
  name: "CCProxy"
  text: "Multi-Provider AI Proxy"
  tagline: "Seamlessly connect Claude Code to any AI provider with zero configuration changes"
  image:
    src: /logo.svg
    alt: CCProxy
  actions:
    - theme: brand
      text: Get Started
      link: /guide/
    - theme: alt
      text: View on GitHub
      link: https://github.com/praneybehl/ccproxy

features:
  - icon: 🚀
    title: Multiple AI Providers
    details: Support for Groq, OpenRouter, OpenAI, XAI (Grok), Google Gemini, Mistral AI, and Ollama
  - icon: 🔄
    title: Anthropic Compatible
    details: Perfect format conversion between Anthropic API and any supported provider
  - icon: ⚙️
    title: Zero Config Changes
    details: Use with Claude Code by simply setting environment variables
  - icon: 🛠️
    title: Tool Support
    details: Full support for function calling and tool use across all providers
  - icon: 📊
    title: Health Monitoring
    details: Built-in health checks and status endpoints for monitoring
  - icon: 🐳
    title: Docker Ready
    details: Cross-platform binaries and Docker images for easy deployment
---

## Quick Example

```bash
# Install CCProxy
wget https://github.com/praneybehl/ccproxy/releases/latest/download/ccproxy-linux-amd64
chmod +x ccproxy-linux-amd64

# Configure for Groq
export PROVIDER=groq
export GROQ_API_KEY=your_groq_api_key

# Start the proxy
./ccproxy-linux-amd64

# Use with Claude Code
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED
claude
```

## Why CCProxy?

CCProxy bridges the gap between Claude Code's Anthropic API format and other AI providers. Instead of being limited to Claude models, you can now use:

- **Groq**: Ultra-fast inference with Llama, Mixtral, and more
- **OpenRouter**: Access to 100+ models from various providers
- **OpenAI**: GPT-4o, GPT-4 Turbo, and other OpenAI models
- **XAI**: Grok models with real-time data access
- **Google Gemini**: Gemini 2.0 Flash and Pro models
- **Mistral AI**: Mistral Large, Medium, and specialized models
- **Ollama**: Local models for privacy and offline usage

All while maintaining perfect compatibility with Claude Code's interface, including full support for tool calling and function use.