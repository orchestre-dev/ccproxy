---
layout: home
title: CCProxy - AI Request Proxy for Claude Code | Multi-Provider LLM Gateway
description: CCProxy is the premier AI request proxy for Claude Code, enabling seamless integration with OpenAI GPT-4, Google Gemini, DeepSeek, Kimi K2, and more. Transform Claude Code into a multi-provider AI development platform.
keywords: CCProxy, AI proxy for Claude Code, Claude Code proxy server, LLM gateway, AI model router, OpenAI proxy, Anthropic proxy, Google Gemini proxy, multi-provider AI, Claude Code integration

hero:
  name: "CCProxy"
  text: "AI Request Proxy for Claude Code"
  tagline: "Enable Claude Code to work with OpenAI, Google Gemini, DeepSeek, and 7+ AI providers through intelligent routing"
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
      text: Share Feedback
      link: https://github.com/orchestre-dev/ccproxy/discussions

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

<SocialShare />

## 🌟 Featured: Kimi K2 + Claude Code

<div class="showcase-grid">
  <div class="showcase-item">
    <div class="showcase-title">⚡ Ultra-Fast Kimi K2</div>
    <div class="showcase-description">
      Experience blazing-fast inference with Moonshot AI's Kimi K2 via Groq or OpenRouter. Sub-second response times with 32B activated parameters and 1T total parameters for exceptional AI performance.
    </div>
    <a href="/kimi-k2" class="showcase-link">Learn More →</a>
  </div>
  
  <div class="showcase-item">
    <div class="showcase-title">💰 Incredible Value</div>
    <div class="showcase-description">
      Kimi K2 offers exceptional value - significantly cheaper than GPT-4 with comparable quality. Available via Groq for ultra-fast inference or OpenRouter for reliable access.
    </div>
    <a href="/kimi-k2" class="showcase-link">See Details →</a>
  </div>
  
  <div class="showcase-item">
    <div class="showcase-title">🔧 Zero Setup Required</div>
    <div class="showcase-description">
      Works instantly with Claude Code. No API changes, no code modifications - just set two environment variables and unlock the power of Kimi K2 through CCProxy.
    </div>
    <a href="/guide/quick-start" class="showcase-link">Quick Start →</a>
  </div>
</div>

## Quick Start

```bash
# Install CCProxy with one command
curl -sSL https://raw.githubusercontent.com/orchestre-dev/ccproxy/main/install.sh | bash

# Configure and start (example with Groq + Kimi K2)
export PROVIDER=groq GROQ_API_KEY=your_key
ccproxy &

# Connect Claude Code
export ANTHROPIC_BASE_URL=http://localhost:3456
claude-code "Help me with coding tasks"
```

**[Complete setup guide →](/guide/quick-start)** • **[Installation options →](/guide/installation)**

## Why Choose CCProxy?

CCProxy transforms Claude Code into a **universal AI development tool** by connecting it to the best AI providers available. Instead of being limited to Claude models, unlock the power of:

<div class="showcase-grid">
  <div class="showcase-item">
    <div class="showcase-title">🎯 Top AI Providers</div>
    <div class="showcase-description">
      <strong>Groq + Kimi K2:</strong> Ultra-fast inference with sub-second response times<br><br>
      <strong>OpenRouter + Kimi K2:</strong> Access to 100+ models including Kimi K2<br><br>
      <strong>OpenAI GPT:</strong> Industry-standard models with latest features<br><br>
      <strong>XAI Grok:</strong> Real-time data access and X integration<br><br>
      <strong>Google Gemini:</strong> Advanced multimodal capabilities<br><br>
      <strong>Mistral AI:</strong> European privacy-focused models<br><br>
      <strong>Ollama:</strong> Complete privacy with local models
    </div>
  </div>
  
  <div class="showcase-item">
    <div class="showcase-title">💡 Perfect for Developers</div>
    <div class="showcase-description">
      <strong>Code Generation:</strong> Intelligent completion and suggestions<br><br>
      <strong>Debugging:</strong> Advanced error analysis and solutions<br><br>
      <strong>Architecture:</strong> System design and optimization guidance<br><br>
      <strong>Documentation:</strong> Automated docs and comments<br><br>
      <strong>Testing:</strong> Comprehensive test generation
    </div>
  </div>
  
  <div class="showcase-item">
    <div class="showcase-title">🛡️ Enterprise Ready</div>
    <div class="showcase-description">
      <strong>Health Monitoring:</strong> Built-in status endpoints<br><br>
      <strong>Docker Support:</strong> Cross-platform deployment<br><br>
      <strong>Logging:</strong> Structured request/response logging<br><br>
      <strong>Security:</strong> API key management and validation<br><br>
      <strong>Scalability:</strong> High-performance proxy architecture
    </div>
  </div>
</div>

## 🚀 Beyond Fast AI Access: Complete Development Stack

CCProxy gets you **fast, affordable AI access**. But what about **production-ready code**?

<div class="showcase-grid">
  <div class="showcase-item">
    <div class="showcase-title">🏗️ CCProxy: Infrastructure Layer</div>
    <div class="showcase-description">
      <strong>Ultra-fast AI access:</strong> Kimi K2 via Groq in sub-seconds<br><br>
      <strong>Cost-effective:</strong> Cheaper alternatives to expensive APIs<br><br>
      <strong>Multi-provider:</strong> 7+ AI providers in one proxy<br><br>
      <strong>Claude Code compatible:</strong> Zero configuration changes
    </div>
  </div>
  
  <div class="showcase-item">
    <div class="showcase-title">🧠 Orchestre: Intelligence Layer</div>
    <div class="showcase-description">
      <strong>Production-ready code:</strong> Transform AI slop into real applications<br><br>
      <strong>Context intelligence:</strong> Teaches AI your project conventions<br><br>
      <strong>Multi-AI review:</strong> Quality assurance from different perspectives<br><br>
      <strong>Ship MVPs in 3 days:</strong> From prototype to production
    </div>
    <a href="https://orchestre.dev" class="showcase-link">Learn More →</a>
  </div>
  
  <div class="showcase-item">
    <div class="showcase-title">⚡ The Complete Stack</div>
    <div class="showcase-description">
      <strong>Start with CCProxy:</strong> Blazing-fast, affordable AI access<br><br>
      <strong>Add Orchestre:</strong> Context-aware, production-ready development<br><br>
      <strong>Ship faster:</strong> MVPs 10x faster with professional architecture<br><br>
      <strong>Perfect together:</strong> Infrastructure + Intelligence = Success
    </div>
    <a href="https://orchestre.dev" class="showcase-link">Explore Stack →</a>
  </div>
</div>

### Why Developers Choose Both

> *"I use CCProxy with Groq's Kimi K2 for ultra-fast AI responses during development. Then Orchestre's context intelligence ensures my code is production-ready from day one. Shipped my SaaS MVP in 3 days."*

**The Modern AI Development Workflow:**
1. **CCProxy** handles blazing-fast, cost-effective AI infrastructure
2. **Orchestre** adds context intelligence and quality assurance  
3. **You** ship production applications in days, not months

---

**Ready to supercharge your Claude Code experience?** Get started in under 2 minutes and join thousands of developers using CCProxy for AI-powered development!

## 💬 Community & Support

- **[🗨️ GitHub Discussions](https://github.com/orchestre-dev/ccproxy/discussions)** - Ask questions, share tips, and connect with the community
- **[🐛 Report Issues](https://github.com/orchestre-dev/ccproxy/issues)** - Found a bug? Let us know so we can fix it
- **[✨ Feature Requests](https://github.com/orchestre-dev/ccproxy/issues/new?template=feature_request.md)** - Suggest new features and improvements
- **[📖 Edit Documentation](https://github.com/orchestre-dev/ccproxy/tree/main/docs)** - Help improve our docs