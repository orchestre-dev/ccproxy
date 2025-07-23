---
layout: home
title: CCProxy - AI Request Proxy for Claude Code | Multi-Provider LLM Gateway
description: CCProxy is the premier AI request proxy for Claude Code, enabling seamless integration with OpenAI models, Google Gemini, DeepSeek, and 100+ open source LLMs. Transform Claude Code into a multi-provider AI development platform through standard OpenAI-compatible API translation.
keywords: CCProxy, AI proxy for Claude Code, Claude Code proxy server, LLM gateway, AI model router, OpenAI proxy, Anthropic proxy, Google Gemini proxy, multi-provider AI, Claude Code integration

hero:
  name: "CCProxy"
  text: "AI Request Proxy for Claude Code"
  tagline: "Enable Claude Code to work with OpenAI, Google Gemini, DeepSeek, and 100+ models via OpenRouter through intelligent routing"
  image:
    src: /ccproxy_icon.png
    alt: CCProxy - Universal AI Proxy
  actions:
    - theme: brand
      text: Get Started
      link: /guide/
    - theme: alt
      text: Share Feedback
      link: https://github.com/orchestre-dev/ccproxy/discussions

features:
  - icon: 🌐
    title: 5 AI Providers with 100+ Models via OpenRouter
    details: Anthropic, OpenAI, Google Gemini, DeepSeek, and OpenRouter (100+ models including open source leaders)
  - icon: 🚀
    title: Leading Open Source LLMs - Qwen3, Kimi K2, and More
    details: Access top open source models like Qwen3 235B (70.3 AIME), Qwen2.5-Coder-32B (73.7% Aider), Kimi K2 (65.8% SWE-bench), DeepSeek, and dozens more via OpenRouter
  - icon: 🔄
    title: Perfect Claude Code Integration
    details: Seamless Anthropic API compatibility - just set environment variables and you're ready to go
  - icon: 🛠️
    title: Advanced Tool Support
    details: Full function calling and tool use across all providers with intelligent format conversion. Note that Claude Code requires function calling support.
  - icon: 📊
    title: Production Ready
    details: Built-in health monitoring, logging, and Docker deployment for enterprise use
  - icon: 🏆
    title: Best-in-Class Performance
    details: Optimized for speed with intelligent caching and connection pooling
---

<SocialShare />

---

<NewsletterForm />

## Quick Start

```bash
# Install CCProxy with one command
curl -sSL https://raw.githubusercontent.com/orchestre-dev/ccproxy/main/install.sh | bash

# Create a config file with your API keys
cat > ~/.ccproxy/config.json << EOF
{
  "providers": [
    {
      "name": "openai",
      "api_key": "sk-...",
      "models": ["gpt-4.1", "gpt-4.1-mini", "o3"],
      "enabled": true
    }
  ],
  "routes": {
    "default": {
      "provider": "openai",
      "model": "gpt-4.1"
    }
  },
  "note": "CCProxy transforms standard OpenAI-compatible requests to each provider's format"
}
EOF

# Start CCProxy
ccproxy start

# Connect Claude Code
export ANTHROPIC_BASE_URL=http://localhost:3456
claude "Help me with coding tasks"
```

**[Complete setup guide →](/guide/quick-start)** • **[Installation options →](/guide/installation)**

## What CCProxy Does

CCProxy is a **translation proxy** that:
- ✅ Converts OpenAI-compatible API requests to provider-specific formats
- ✅ Routes requests to different providers based on model names
- ✅ Supports standard parameters (temperature, max_tokens, etc.)
- ✅ Handles streaming responses (SSE)
- ✅ Provides function calling support for compatible clients
- ✅ Offers health monitoring and logging

CCProxy does **not**:
- ❌ Add new capabilities to Claude Code beyond API translation
- ❌ Support provider-specific features without function calling
- ❌ Modify request/response content (it's a pass-through proxy)

## Why Choose CCProxy?

CCProxy transforms Claude Code into a **universal AI development tool** by connecting it to the best AI providers available. CCProxy acts as a translation layer, converting standard OpenAI-compatible requests into each provider's specific format. Instead of being limited to Claude models, unlock the power of:

<div class="showcase-grid">
  <div class="showcase-item">
    <div class="showcase-title">🎯 Top AI Providers</div>
    <div class="showcase-description">
      <strong>OpenRouter:</strong> 100+ models including Qwen3 235B (#1 reasoning), Qwen2.5-Coder, Kimi K2, XAI Grok, Mistral, and more<br><br>
      <strong>OpenAI GPT:</strong> Industry-standard models with latest features<br><br>
      <strong>Google Gemini:</strong> Advanced multimodal capabilities<br><br>
      <strong>DeepSeek:</strong> Cost-effective Chinese models with strong coding abilities<br><br>
      <strong>Anthropic:</strong> Native Claude support for seamless integration
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
      <strong>Ultra-fast AI access:</strong> Qwen3 235B and 100+ models via OpenRouter<br><br>
      <strong>Cost-effective:</strong> Cheaper alternatives to expensive APIs<br><br>
      <strong>Multi-provider:</strong> 5 AI providers with 100+ models in one proxy<br><br>
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

> *"I use CCProxy with OpenRouter's Qwen3 235B for superior reasoning during development. Then Orchestre's context intelligence ensures my code is production-ready from day one. Shipped my SaaS MVP in 3 days."*

**The Modern AI Development Workflow:**
1. **CCProxy** handles blazing-fast, cost-effective AI infrastructure
2. **Orchestre** adds context intelligence and quality assurance  
3. **You** ship production applications in days, not months

---

<div class="info-cards">
  <div class="info-card">
    <div class="card-icon">📋</div>
    <h3>Latest Models</h3>
    <p>CCProxy supports the newest models from each provider:</p>
    <ul>
      <li><strong>Open Source:</strong> Qwen3 235B, Qwen2.5-Coder-32B, Kimi K2, Mistral, and 100+ more via OpenRouter</li>
      <li><strong>Anthropic:</strong> Claude Opus 4 & Sonnet 4 (July 2025)</li>
      <li><strong>OpenAI:</strong> Latest models including o3/o4-mini (July 2025)</li>
      <li><strong>Google:</strong> Gemini 2.5 family (July 2025)</li>
      <li><strong>DeepSeek:</strong> V3-0324 & R1-0528 (2025)</li>
    </ul>
    <p class="note">Note: CCProxy supports standard OpenAI-compatible parameters. Provider-specific features require function calling support.</p>
    <a href="/guide/routing" class="card-link">See routing guide →</a>
  </div>
  
  <div class="info-card orchestre-card">
    <div class="card-icon">🚀</div>
    <h3>Stop getting AI slop</h3>
    <p>CCProxy handles the AI infrastructure. <strong>Orchestre</strong> adds the context intelligence that transforms generic AI code into production-ready software that actually works.</p>
    <p class="highlight">✨ Ship MVPs in 3 days with battle-tested recipes</p>
    <a href="https://orchestre.dev" class="cta-button">Get Started with Orchestre</a>
  </div>
</div>

<style>
.info-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
  gap: 24px;
  margin: 48px 0;
}

.info-card {
  background: var(--vp-c-bg-alt);
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  padding: 32px;
  transition: all 0.25s ease;
  position: relative;
}

.info-card:hover {
  border-color: var(--vp-c-brand-1);
  transform: translateY(-2px);
}

.dark .info-card {
  background: var(--vp-c-bg-elv);
}

.info-card .card-icon {
  font-size: 48px;
  line-height: 1;
  margin-bottom: 16px;
  display: block;
}

.info-card h3 {
  margin: 0 0 16px 0;
  font-size: 24px;
  font-weight: 600;
  letter-spacing: -0.02em;
  line-height: 32px;
}

.info-card p {
  font-size: 16px;
  line-height: 24px;
  color: var(--vp-c-text-2);
  margin: 0 0 12px 0;
}

.info-card ul {
  list-style: none;
  padding: 0;
  margin: 16px 0;
}

.info-card li {
  font-size: 14px;
  line-height: 24px;
  color: var(--vp-c-text-2);
  padding: 4px 0;
}

.info-card .note {
  font-size: 14px;
  line-height: 20px;
  color: var(--vp-c-text-3);
  margin-top: 16px;
}

.info-card .highlight {
  color: #2563eb;
  font-weight: 500;
  font-size: 16px;
  line-height: 24px;
}

.dark .info-card .highlight {
  color: #60a5fa;
}

/* Fix for Orchestre card highlight text */
.orchestre-card .highlight {
  color: #0066ff;
}

.dark .orchestre-card .highlight {
  color: #00aaff;
}

.orchestre-card .cta-button {
  color: #000000;
}

.orchestre-card .cta-button:hover {
  color: #000000;
}

.card-link {
  display: inline-flex;
  align-items: center;
  margin-top: 16px;
  color: var(--vp-c-brand-1);
  text-decoration: none !important;
  font-weight: 500;
  font-size: 16px;
  transition: color 0.25s;
}

.card-link:hover {
  color: var(--vp-c-brand-2);
}

.cta-button {
  display: inline-block;
  margin-top: 24px;
  padding: 12px 24px;
  background-color: var(--vp-button-brand-bg);
  color: var(--vp-button-brand-text);
  text-decoration: none !important;
  font-weight: 500;
  font-size: 16px;
  line-height: 24px;
  border-radius: 24px;
  transition: all 0.25s ease;
  text-align: center;
  border: 1px solid var(--vp-button-brand-bg);
}

.cta-button:hover {
  background-color: var(--vp-button-brand-hover-bg);
  border-color: var(--vp-button-brand-hover-bg);
  color: var(--vp-button-brand-text);
  text-decoration: none !important;
  transform: translateY(-1px);
}

.orchestre-card {
  background: var(--vp-c-bg-alt);
}

.dark .orchestre-card {
  background: var(--vp-c-bg-elv);
}

@media (max-width: 768px) {
  .info-cards {
    grid-template-columns: 1fr;
    gap: 16px;
  }
  
  .info-card {
    padding: 24px;
  }
}
</style>

---

**Ready to supercharge your Claude Code experience?** Get started in under 2 minutes and join thousands of developers using CCProxy for AI-powered development!

## 💬 Community & Support

- **[🗨️ GitHub Discussions](https://github.com/orchestre-dev/ccproxy/discussions)** - Ask questions, share tips, and connect with the community
- **[🐛 Report Issues](https://github.com/orchestre-dev/ccproxy/issues)** - Found a bug? Let us know so we can fix it
- **[✨ Feature Requests](https://github.com/orchestre-dev/ccproxy/issues/new?template=feature_request.md)** - Suggest new features and improvements
- **[📖 Edit Documentation](https://github.com/orchestre-dev/ccproxy/tree/main/docs)** - Help improve our docs