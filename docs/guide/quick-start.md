# Quick Start

Get CCProxy running in under 5 minutes!

## Prerequisites

- An API key from your chosen provider (Groq, OpenAI, etc.)
- Claude Code installed on your system

## Step 1: Download CCProxy

Choose your platform and download the latest release:

::: code-group

```bash [Linux (AMD64)]
wget https://github.com/praneybehl/ccproxy/releases/latest/download/ccproxy-linux-amd64
chmod +x ccproxy-linux-amd64
sudo mv ccproxy-linux-amd64 /usr/local/bin/ccproxy
```

```bash [Linux (ARM64)]
wget https://github.com/praneybehl/ccproxy/releases/latest/download/ccproxy-linux-arm64
chmod +x ccproxy-linux-arm64
sudo mv ccproxy-linux-arm64 /usr/local/bin/ccproxy
```

```bash [macOS (Intel)]
wget https://github.com/praneybehl/ccproxy/releases/latest/download/ccproxy-darwin-amd64
chmod +x ccproxy-darwin-amd64
sudo mv ccproxy-darwin-amd64 /usr/local/bin/ccproxy
```

```bash [macOS (Apple Silicon)]
wget https://github.com/praneybehl/ccproxy/releases/latest/download/ccproxy-darwin-arm64
chmod +x ccproxy-darwin-arm64
sudo mv ccproxy-darwin-arm64 /usr/local/bin/ccproxy
```

```powershell [Windows]
# Download from GitHub releases
# https://github.com/praneybehl/ccproxy/releases/latest/download/ccproxy-windows-amd64.exe
```

:::

## Step 2: Choose Your Provider

Pick one of the supported providers and get an API key:

### Option A: Groq (Recommended for Speed)

1. Visit [console.groq.com](https://console.groq.com)
2. Sign up for a free account
3. Generate an API key
4. Note: Free tier includes generous limits

### Option B: OpenAI

1. Visit [platform.openai.com](https://platform.openai.com)
2. Sign up and add billing information
3. Generate an API key
4. Note: Pay-per-use pricing

### Option C: Other Providers

See our [Provider Guide](/providers/) for setup instructions for Gemini, Mistral, XAI, OpenRouter, and Ollama.

## Step 3: Configure Environment

Create a `.env` file or set environment variables:

::: code-group

```bash [Groq]
export PROVIDER=groq
export GROQ_API_KEY=gsk_your_groq_api_key_here
```

```bash [OpenAI]
export PROVIDER=openai
export OPENAI_API_KEY=sk-your_openai_api_key_here
```

```bash [Gemini]
export PROVIDER=gemini
export GEMINI_API_KEY=your_gemini_api_key_here
```

```bash [Mistral]
export PROVIDER=mistral
export MISTRAL_API_KEY=your_mistral_api_key_here
```

:::

## Step 4: Start CCProxy

```bash
ccproxy
```

You should see output like:
```
{"level":"info","msg":"Successfully initialized groq provider with model moonshotai/kimi-k2-instruct","time":"2025-01-17T10:30:00.000Z"}
{"level":"info","msg":"Starting server on 0.0.0.0:7187","time":"2025-01-17T10:30:00.001Z"}
```

## Step 5: Configure Claude Code

In a new terminal, set Claude Code to use CCProxy:

```bash
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED
```

## Step 6: Test It!

Now use Claude Code normally:

```bash
claude "Hello! What model are you using?"
```

Claude Code will now use your chosen provider instead of Claude! 🎉

## Verify It's Working

Check the CCProxy logs - you should see requests being processed:

```json
{
  "action": "anthropic_request",
  "level": "info",
  "messages": 1,
  "model": "claude-3-sonnet",
  "provider": "groq",
  "time": "2025-01-17T10:30:01.000Z"
}
```

You can also check the status endpoint:

```bash
curl http://localhost:7187/status
```

## Next Steps

- [Explore different providers](/providers/) to find the best fit
- [Learn about configuration](/guide/configuration) options
- [Set up health monitoring](/guide/health-checks) for production use
- [Deploy with Docker](/guide/docker) for scalability

## Troubleshooting

### Port Already in Use
```bash
# Use a different port
export SERVER_PORT=8080
ccproxy
```

### API Key Issues
- Verify your API key is correct
- Check you have sufficient credits/quota
- Ensure the provider environment variable matches your chosen provider

### Connection Issues
- Check your internet connection
- Verify the provider's API is accessible
- Try a different provider to isolate the issue

Need more help? Check our [troubleshooting guide](/guide/troubleshooting) or open an issue on [GitHub](https://github.com/praneybehl/ccproxy/issues).