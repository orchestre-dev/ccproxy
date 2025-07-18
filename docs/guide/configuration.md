---
title: Configuration Guide - CCProxy Provider Setup
description: Complete configuration guide for CCProxy. Set up Groq Kimi K2, OpenAI, Gemini, Mistral, XAI Grok, OpenRouter, and Ollama providers with Claude Code.
keywords: CCProxy configuration, Groq setup, OpenAI config, Gemini setup, Mistral config, XAI Grok, OpenRouter, Ollama, Claude Code integration
---

# Configuration Guide

<SocialShare />

Configure CCProxy to work with your preferred AI provider. CCProxy supports 7+ major providers, each with their own configuration options and capabilities.

## Quick Setup

The fastest way to get started is using our installation script:

```bash
# Install CCProxy with auto-detection
curl -sSL https://raw.githubusercontent.com/orchestre-dev/ccproxy/main/install.sh | bash

# Then configure your provider (see examples below)
export PROVIDER=groq
export GROQ_API_KEY=your_api_key
ccproxy
```

## Environment Variables

CCProxy uses environment variables for configuration. Create a `.env` file or export variables directly:

### Core Configuration

```bash
# Required: Choose your provider
PROVIDER=groq

# Optional: Server configuration
PORT=7187                    # Default: 7187
LOG_LEVEL=info              # Options: debug, info, warn, error
LOG_FORMAT=json             # Options: json, text
```

## Provider-Specific Configuration

### 🚀 Groq (Recommended for Kimi K2)

**Ultra-fast inference with Kimi K2 support**

```bash
# Required
PROVIDER=groq
GROQ_API_KEY=gsk_your_groq_api_key_here

# Optional
GROQ_MODEL=moonshotai/kimi-k2-instruct    # Default model
GROQ_MAX_TOKENS=8192                      # Default: 16384
GROQ_BASE_URL=https://api.groq.com/openai/v1  # Default URL
```

**Get your API key:** [console.groq.com](https://console.groq.com/)

**Available Models:** [View current models and pricing →](https://console.groq.com/docs/models)

### 🌐 OpenRouter

**Access to 100+ models including Kimi K2**

```bash
# Required
PROVIDER=openrouter
OPENROUTER_API_KEY=sk-or-v1-your_openrouter_key_here

# Optional
OPENROUTER_MODEL=moonshotai/kimi-k2-instruct  # Default model
OPENROUTER_MAX_TOKENS=8192                    # Default: 16384
OPENROUTER_SITE_URL=https://yoursite.com      # For tracking
OPENROUTER_SITE_NAME=YourApp                  # For tracking
```

**Get your API key:** [openrouter.ai](https://openrouter.ai/)

**Available Models:** [Browse 100+ models and pricing →](https://openrouter.ai/models)

### 🤖 OpenAI

**Industry-standard GPT models**

```bash
# Required
PROVIDER=openai
OPENAI_API_KEY=sk-your_openai_api_key_here

# Optional
OPENAI_MODEL=gpt-4.1                         # Default model (2025)
OPENAI_MAX_TOKENS=16384                      # Default: 16384
OPENAI_ORGANIZATION=org-your_org_id          # Optional
OPENAI_BASE_URL=https://api.openai.com/v1    # Default URL
```

**Get your API key:** [platform.openai.com](https://platform.openai.com/)

**Available Models:** [View current models and pricing →](https://platform.openai.com/docs/models)

### 🧠 Google Gemini

**Advanced multimodal AI**

```bash
# Required
PROVIDER=gemini
GEMINI_API_KEY=your_gemini_api_key_here

# Optional
GEMINI_MODEL=gemini-2.0-flash                # Default model (2025)
GEMINI_MAX_TOKENS=32768                      # Default: 16384
GEMINI_BASE_URL=https://generativelanguage.googleapis.com  # Default
```

**Get your API key:** [aistudio.google.com](https://aistudio.google.com/)

**Available Models:** [View current models and pricing →](https://ai.google.dev/gemini-api/docs/models)

### 🇪🇺 Mistral AI

**European privacy-focused AI**

```bash
# Required
PROVIDER=mistral
MISTRAL_API_KEY=your_mistral_api_key_here

# Optional
MISTRAL_MODEL=mistral-large-latest           # Default model
MISTRAL_MAX_TOKENS=32768                     # Default: 16384
MISTRAL_BASE_URL=https://api.mistral.ai/v1   # Default URL
```

**Get your API key:** [console.mistral.ai](https://console.mistral.ai/)

**Available Models:** [View current models and pricing →](https://docs.mistral.ai/getting-started/models/models_overview/)

### 🔥 XAI (Grok)

**Real-time data access with X integration**

```bash
# Required
PROVIDER=xai
XAI_API_KEY=xai-your_xai_api_key_here

# Optional
XAI_MODEL=grok-4                            # Default model (2025)
XAI_MAX_TOKENS=16384                        # Default: 16384
XAI_BASE_URL=https://api.x.ai/v1            # Default URL
```

**Get your API key:** [console.x.ai](https://console.x.ai/) ($25 free credits/month)

**Available Models:** [View current models and pricing →](https://docs.x.ai/docs/models)

### 🏠 Ollama

**Local models for complete privacy**

```bash
# Required
PROVIDER=ollama
OLLAMA_MODEL=llama3.3                       # Must be downloaded locally

# Optional
OLLAMA_BASE_URL=http://localhost:11434      # Default Ollama URL
OLLAMA_MAX_TOKENS=16384                     # Default: 16384
```

**Setup Requirements:**
1. Install Ollama: [ollama.ai](https://ollama.ai/)
2. Download a model: `ollama pull llama3.3`
3. Start Ollama: `ollama serve`

**Available Models:** [Browse all local models →](https://ollama.com/library)

## Claude Code Integration

Once CCProxy is configured and running, integrate with Claude Code:

```bash
# Set Claude Code to use CCProxy
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED

# Verify it's working
claude-code "Hello, can you help me with coding?"
```

## Advanced Configuration

### Custom Configuration File

Create a `config.yaml` file for advanced settings:

```yaml
# config.yaml
server:
  port: 7187
  host: "0.0.0.0"
  read_timeout: "30s"
  write_timeout: "30s"

logging:
  level: "info"
  format: "json"
  file: "ccproxy.log"

provider:
  name: "groq"
  timeout: "30s"
  max_retries: 3
  retry_delay: "1s"

# Provider-specific settings
groq:
  api_key: "${GROQ_API_KEY}"
  model: "moonshotai/kimi-k2-instruct"
  max_tokens: 8192
  base_url: "https://api.groq.com/openai/v1"
```

Run with config file:
```bash
ccproxy --config config.yaml
```

### Multiple Provider Setup

For switching between providers easily:

```bash
# Create multiple .env files
cp .env .env.groq
cp .env .env.openai
cp .env .env.gemini

# Switch providers
source .env.groq && ccproxy    # Use Groq
source .env.openai && ccproxy  # Use OpenAI
source .env.gemini && ccproxy  # Use Gemini
```

### Load Balancing (Coming Soon)

Configure multiple providers for failover:

```yaml
providers:
  - name: "groq"
    weight: 70
    config:
      api_key: "${GROQ_API_KEY}"
      model: "moonshotai/kimi-k2-instruct"
  
  - name: "openrouter"
    weight: 30
    config:
      api_key: "${OPENROUTER_API_KEY}"
      model: "moonshotai/kimi-k2-instruct"
```

## Environment Files

Create provider-specific `.env` files:

### .env.kimi-groq
```bash
# Ultra-fast Kimi K2 via Groq
PROVIDER=groq
GROQ_API_KEY=gsk_your_key_here
GROQ_MODEL=moonshotai/kimi-k2-instruct
LOG_LEVEL=info
```

### .env.kimi-openrouter
```bash
# Kimi K2 via OpenRouter with fallbacks
PROVIDER=openrouter
OPENROUTER_API_KEY=sk-or-v1-your_key_here
OPENROUTER_MODEL=moonshotai/kimi-k2-instruct
OPENROUTER_SITE_NAME=YourApp
LOG_LEVEL=info
```

### .env.local-privacy
```bash
# Complete privacy with Ollama
PROVIDER=ollama
OLLAMA_MODEL=llama3.2
OLLAMA_BASE_URL=http://localhost:11434
LOG_LEVEL=debug
```

## Validation

Test your configuration:

```bash
# Health check
curl http://localhost:7187/health

# Provider status
curl http://localhost:7187/status

# Test message
curl -X POST http://localhost:7187/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [{"role": "user", "content": "Hello!"}],
    "max_tokens": 100
  }'
```

## Troubleshooting

### Common Issues

**API Key Invalid:**
- Verify your API key is correct
- Check for extra spaces or newlines
- Ensure the key has proper permissions

**Connection Refused:**
- Check if the provider's API is accessible
- Verify your network connection
- Try a different provider

**Model Not Found:**
- Verify the model name is correct
- Check if the model is available for your API key
- Try the default model for your provider

**Rate Limiting:**
- Switch to a different provider
- Implement retry logic
- Consider upgrading your API plan

### Getting Help

- 📖 [Installation Guide](/guide/installation)
- 🚀 [Quick Start](/guide/quick-start)  
- ⭐ [Kimi K2 Setup](/kimi-k2)
- 💬 [Ask Questions](https://github.com/orchestre-dev/ccproxy/discussions) - Community support
- 🐛 [Report Issues](https://github.com/orchestre-dev/ccproxy/issues) - Bug reports and feature requests