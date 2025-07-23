---
title: Ollama (Local Models) - Run AI Models Locally with Claude Code
description: Use Ollama's local AI models with Claude Code through CCProxy. Complete privacy with models running on your own hardware.
keywords: Ollama, local AI, Claude Code, CCProxy, privacy, offline AI, Llama, Mistral, CodeLlama
---

# Ollama Provider

Run AI models locally on your own hardware with complete privacy. Ollama provides an OpenAI-compatible API that works seamlessly with CCProxy, allowing you to use Claude Code without sending data to external servers.

## Why Choose Ollama?

- 🔒 **Complete Privacy**: All processing happens locally - your data never leaves your machine
- 💰 **Free to Use**: No API costs, just your electricity
- 🚀 **Fast Response**: No network latency for model calls
- 🌐 **Works Offline**: After initial model download, no internet required
- 🛠️ **Easy Setup**: Simple installation and model management

## Setup

### 1. Install Ollama

**macOS/Linux:**
```bash
curl -fsSL https://ollama.com/install.sh | sh
```

**Windows:**
Download from [ollama.com/download](https://ollama.com/download)

### 2. Download a Model

Choose a model based on your needs:

```bash
# For general use (7B parameters, ~4GB)
ollama pull llama3.1

# For coding tasks (7B parameters, ~4GB)
ollama pull codellama

# For smaller systems (3B parameters, ~2GB)
ollama pull phi3

# For fast responses (7B parameters, ~4GB)
ollama pull mistral
```

### 3. Configure CCProxy

Create or update your CCProxy configuration:

```json
{
  "providers": [
    {
      "name": "openai",
      "api_base_url": "http://localhost:11434/v1",
      "api_key": "ollama",
      "models": ["llama3.1", "codellama", "mistral"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "llama3.1"
    }
  }
}
```

**Important**: Use `"openai"` as the provider name since Ollama provides an OpenAI-compatible API.

### 4. Start CCProxy

```bash
ccproxy start
ccproxy code
```

## Available Models

### General Purpose
- **llama3.1** (8B/70B) - Latest Llama model with tool calling support
- **llama3** (8B/70B) - Previous generation, still excellent
- **mistral** (7B) - Fast and efficient

### Coding Specialized
- **codellama** (7B/13B/34B) - Optimized for code generation
- **deepseek-coder** (1.3B/6.7B) - Efficient coding model
- **qwen2.5-coder** (Various sizes) - Strong coding capabilities

### Small Models (Low Resource)
- **phi3** (3B) - Microsoft's efficient model
- **gemma** (2B/7B) - Google's lightweight models
- **tinyllama** (1.1B) - Tiny but capable

## Configuration Examples

### Basic Setup

Minimal configuration using Llama 3.1:

```json
{
  "providers": [
    {
      "name": "openai",
      "api_base_url": "http://localhost:11434/v1",
      "api_key": "ollama",
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "llama3.1"
    }
  }
}
```

### Multi-Model Setup

Different models for different tasks:

```json
{
  "providers": [
    {
      "name": "openai",
      "api_base_url": "http://localhost:11434/v1",
      "api_key": "ollama",
      "models": ["llama3.1", "codellama", "phi3"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "llama3.1"
    },
    "background": {
      "provider": "openai",
      "model": "phi3"
    },
    "think": {
      "provider": "openai",
      "model": "llama3.1:70b"
    }
  }
}
```

### Remote Ollama Server

If running Ollama on another machine:

```json
{
  "providers": [
    {
      "name": "openai",
      "api_base_url": "http://192.168.1.100:11434/v1",
      "api_key": "ollama",
      "enabled": true
    }
  ]
}
```

## Function Calling Support

Ollama supports function calling with compatible models:

**Models with Tool Support:**
- ✅ **llama3.1** - Full tool calling support
- ✅ **mistral** - Basic tool support
- ❌ **codellama** - No tool support (code generation only)
- ❌ **phi3** - No tool support

For Claude Code usage, we recommend **llama3.1** as it has the best tool calling compatibility.

## Performance Optimization

### Model Selection by Hardware

**High-End (32GB+ RAM, GPU):**
- llama3.1:70b
- codellama:34b

**Mid-Range (16GB RAM):**
- llama3.1:8b
- codellama:13b
- mistral:7b

**Low-End (8GB RAM):**
- phi3:3b
- gemma:2b
- tinyllama:1.1b

### Ollama Performance Settings

Set environment variables before starting Ollama:

```bash
# Number of parallel requests
export OLLAMA_NUM_PARALLEL=2

# GPU layers (if you have a GPU)
export OLLAMA_NUM_GPU=999

# CPU threads
export OLLAMA_NUM_THREAD=8
```

## Troubleshooting

### Ollama Not Responding

1. Check if Ollama is running:
   ```bash
   ollama list
   ```

2. Start Ollama service:
   ```bash
   ollama serve
   ```

3. Verify API endpoint:
   ```bash
   curl http://localhost:11434/v1/models
   ```

### Model Download Issues

```bash
# Check available models
ollama list

# Remove and re-download
ollama rm llama3.1
ollama pull llama3.1
```

### Memory Issues

For large models on limited RAM:
1. Use smaller models (phi3, gemma)
2. Reduce context size in Ollama
3. Close other applications

### CCProxy Connection Issues

Ensure your configuration uses:
- Provider name: `"openai"` (not "ollama")
- Correct base URL with `/v1` suffix
- API key set to `"ollama"`

## Best Practices

1. **Model Selection**: Start with smaller models and upgrade based on performance needs
2. **Privacy**: Disable Ollama telemetry for complete privacy:
   ```bash
   export OLLAMA_NO_ANALYTICS=true
   ```
3. **Updates**: Keep models updated:
   ```bash
   ollama pull llama3.1
   ```
4. **Resource Management**: Monitor system resources when using large models

## Security Considerations

- Ollama binds to `localhost` by default (secure)
- To expose Ollama to network, use:
  ```bash
  OLLAMA_HOST=0.0.0.0 ollama serve
  ```
  ⚠️ Only do this on trusted networks

## Next Steps

- Explore different models for your use case
- Fine-tune models with Ollama's Modelfile feature
- Set up GPU acceleration if available
- Consider running Ollama on a dedicated server

For more information, visit [ollama.com](https://ollama.com).