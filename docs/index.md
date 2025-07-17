---
layout: home
title: CCProxy - Universal AI Proxy for Claude Code with Kimi K2 Support
description: Universal AI proxy supporting Claude Code with Groq Kimi K2, OpenAI GPT, Google Gemini, Mistral AI, XAI Grok, and Ollama. Experience blazing-fast inference with zero configuration changes.
keywords: CCProxy, Claude Code, AI proxy, Kimi K2, Groq, OpenAI, Gemini, Mistral, XAI, Grok, Ollama, multi-provider

hero:
  name: "CCProxy"
  text: "Universal AI Proxy for Claude Code"
  tagline: "Connect Claude Code to Kimi K2, GPT, Gemini, and 7+ AI providers with zero config changes"
  image:
    src: /ccproxy_icon.png
    alt: CCProxy - Universal AI Proxy
  actions:
    - theme: brand
      text: Get Started
      link: /guide/
    - theme: alt
      text: Try Kimi K2
      link: /kimi-k2
    - theme: alt
      text: View on GitHub
      link: https://github.com/praneybehl/ccproxy

features:
  - icon: ⚡
    title: Ultra-Fast Kimi K2 Support
    details: Experience blazing-fast inference with Moonshot AI's Kimi K2 through Groq's LPU infrastructure - sub-second response times!
  - icon: 🌐
    title: 7+ AI Providers
    details: Groq, OpenRouter, OpenAI, XAI (Grok), Google Gemini, Mistral AI, and Ollama - all in one proxy
  - icon: 🔄
    title: Perfect Claude Code Integration
    details: Seamless Anthropic API compatibility - just set environment variables and you're ready to go
  - icon: 🛠️
    title: Advanced Tool Support
    details: Full function calling and tool use across all providers with intelligent format conversion
  - icon: 📊
    title: Production Ready
    details: Built-in health monitoring, logging, and Docker deployment for enterprise use
  - icon: 🏆
    title: Best-in-Class Performance
    details: Optimized for speed with intelligent caching and connection pooling
---

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

## 🌟 Featured: Kimi K2 + Claude Code

<div class="showcase-grid">
  <div class="showcase-item">
    <div class="showcase-title">🚀 Lightning Speed</div>
    <div class="showcase-description">
      Experience sub-second response times with Kimi K2 via Groq's LPU infrastructure. 50x faster than traditional GPU inference.
    </div>
    <a href="/kimi-k2" class="showcase-link">Learn More →</a>
  </div>
  
  <div class="showcase-item">
    <div class="showcase-title">💰 Incredible Value</div>
    <div class="showcase-description">
      At $0.20/1M input tokens, Kimi K2 offers exceptional value - 50x cheaper than GPT-4 with comparable quality.
    </div>
    <a href="/kimi-k2#performance-comparison" class="showcase-link">See Pricing →</a>
  </div>
  
  <div class="showcase-item">
    <div class="showcase-title">🧠 Smart Code Understanding</div>
    <div class="showcase-description">
      8K context window optimized for code. Perfect for refactoring, debugging, and architectural discussions.
    </div>
    <a href="/providers/groq" class="showcase-link">Setup Guide →</a>
  </div>
  
  <div class="showcase-item">
    <div class="showcase-title">🔧 Zero Setup Required</div>
    <div class="showcase-description">
      Works instantly with Claude Code. No API changes, no code modifications - just set two environment variables.
    </div>
    <a href="/guide/quick-start" class="showcase-link">Quick Start →</a>
  </div>
</div>

## Quick Start Example

```bash
# Download CCProxy (macOS)
wget https://github.com/praneybehl/ccproxy/releases/latest/download/ccproxy-darwin-amd64
chmod +x ccproxy-darwin-amd64

# Configure for Kimi K2 via Groq (recommended)
export PROVIDER=groq
export GROQ_API_KEY=your_groq_api_key
export GROQ_MODEL=moonshotai/kimi-k2-instruct

# Start the proxy
./ccproxy-darwin-amd64

# Configure Claude Code (in new terminal)
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED

# Use Claude Code normally - now powered by Kimi K2! 🚀
claude-code "Help me optimize this database query"
```

## Why Choose CCProxy?

CCProxy transforms Claude Code into a **universal AI development tool** by connecting it to the best AI providers available. Instead of being limited to Claude models, unlock the power of:

### 🎯 **Top AI Providers**
- **Groq + Kimi K2**: Sub-second inference, $0.20/1M tokens
- **OpenRouter**: 100+ models, competitive pricing
- **OpenAI GPT**: Industry-standard models with latest features
- **XAI Grok**: Real-time data access and X integration
- **Google Gemini**: Advanced multimodal capabilities
- **Mistral AI**: European privacy-focused models
- **Ollama**: Complete privacy with local models

### 💡 **Perfect for Developers**
- **Code Generation**: Intelligent completion and suggestions
- **Debugging**: Advanced error analysis and solutions
- **Architecture**: System design and optimization guidance
- **Documentation**: Automated docs and comments
- **Testing**: Comprehensive test generation

### 🛡️ **Enterprise Ready**
- **Health Monitoring**: Built-in status endpoints
- **Docker Support**: Cross-platform deployment
- **Logging**: Structured request/response logging
- **Security**: API key management and validation
- **Scalability**: High-performance proxy architecture

---

**Ready to supercharge your Claude Code experience?** Get started in under 2 minutes and join thousands of developers using CCProxy for AI-powered development!

<script>
function shareToTwitter() {
  const url = encodeURIComponent(window.location.href);
  const text = encodeURIComponent('🚀 CCProxy: Universal AI Proxy for Claude Code with Kimi K2 support! Experience blazing-fast AI development with 7+ providers.');
  window.open(`https://twitter.com/intent/tweet?url=${url}&text=${text}`, '_blank');
}

function shareToLinkedIn() {
  const url = encodeURIComponent(window.location.href);
  window.open(`https://www.linkedin.com/sharing/share-offsite/?url=${url}`, '_blank');
}

function shareToReddit() {
  const url = encodeURIComponent(window.location.href);
  const title = encodeURIComponent('CCProxy - Universal AI Proxy for Claude Code with Kimi K2 Support');
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