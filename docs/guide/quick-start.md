---
title: Quick Start - Get CCProxy Running in 2 Minutes
description: Get CCProxy running with Kimi K2 and Claude Code in under 2 minutes. Fast setup guide for immediate AI development productivity.
keywords: CCProxy quick start, Kimi K2 setup, Claude Code integration, AI proxy setup
---

# Quick Start

<SocialShare />

Get CCProxy running with Kimi K2 in under 2 minutes.

## 1. Install CCProxy

```bash
curl -sSL https://raw.githubusercontent.com/orchestre-dev/ccproxy/main/install.sh | bash
```

## 2. Start with Kimi K2

```bash
export PROVIDER=groq GROQ_API_KEY=your_groq_api_key
ccproxy &
```

## 3. Connect Claude Code

```bash
export ANTHROPIC_BASE_URL=http://localhost:7187
claude-code "Write a Python function to reverse a string"
```

## 🎉 Done!

Claude Code now uses your chosen AI provider. Try:

```bash
claude-code "Explain this code and suggest improvements" < your-file.py
claude-code "Create a REST API for user management"
claude-code "Debug this error: TypeError: 'int' object is not subscriptable"
```

## Next Steps

- **[Full Installation Guide](/guide/installation)** - Multiple installation methods
- **[Configuration Guide](/guide/configuration)** - Advanced provider setup
- **[Kimi K2 Guide](/kimi-k2)** - Optimize for ultra-fast development

## Troubleshooting

**Connection refused?** Check if CCProxy is running on port 7187.

**API key error?** Verify your provider API key is correct.

**Need help?** [💬 GitHub Discussions](https://github.com/orchestre-dev/ccproxy/discussions) • [🐛 Report Issues](https://github.com/orchestre-dev/ccproxy/issues)