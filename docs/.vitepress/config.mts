import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

// https://vitepress.dev/reference/site-config
export default withMermaid(defineConfig({
  title: "CCProxy - AI Request Proxy for Claude Code",
  description: "The premier AI request proxy for Claude Code, enabling seamless integration with OpenAI, Google Gemini, Anthropic Claude, and more. High-performance multi-provider LLM gateway.",
  
  // SEO improvements
  head: [
    ['meta', { name: 'keywords', content: 'CCProxy, AI proxy for Claude Code, Claude Code proxy, LLM gateway, AI model router, OpenAI proxy, Anthropic proxy, Google Gemini proxy, multi-provider AI' }],
    ['meta', { name: 'author', content: 'CCProxy Team' }],
    ['meta', { property: 'og:title', content: 'CCProxy - AI Request Proxy for Claude Code' }],
    ['meta', { property: 'og:description', content: 'Enable Claude Code to work with multiple AI providers through intelligent routing and API translation.' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:url', content: 'https://ccproxy.orchestre.dev' }],
    ['meta', { property: 'og:image', content: '/og-image.png' }],
    ['link', { rel: 'icon', href: '/favicon.ico' }],
    ['link', { rel: 'apple-touch-icon', href: '/apple-touch-icon.png' }],
    ['link', { rel: 'manifest', href: '/site.webmanifest' }]
  ],

  // Clean URLs
  cleanUrls: true,

  // Sitemap generation for SEO
  sitemap: {
    hostname: 'https://ccproxy.orchestre.dev'
  },

  themeConfig: {
    // SEO-friendly site title
    siteTitle: 'CCProxy',
    
    logo: '/ccproxy_icon.png',
    
    // Navigation
    nav: [
      { text: 'Guide', link: '/guide/' },
      { text: 'Providers', link: '/providers/' },
      { text: 'Kimi K2', link: '/kimi-k2' },
      { text: 'Blog', link: '/blog/' },
      { text: 'API', link: '/api/' },
      {
        text: 'Community',
        items: [
          { text: 'GitHub', link: 'https://github.com/orchestre-dev/ccproxy' },
          { text: 'Discussions', link: 'https://github.com/orchestre-dev/ccproxy/discussions' },
          { text: 'Issues & Bug Reports', link: 'https://github.com/orchestre-dev/ccproxy/issues' },
          { text: 'Feature Requests', link: 'https://github.com/orchestre-dev/ccproxy/issues/new?template=feature_request.md' }
        ]
      },
      { text: 'Orchestre', link: 'https://orchestre.dev' }
    ],

    // Sidebar
    sidebar: {
      '/guide/': [
        {
          text: 'Getting Started',
          items: [
            { text: 'Quick Start', link: '/guide/quick-start' },
            { text: 'Installation', link: '/guide/installation' },
            { text: 'Configuration', link: '/guide/configuration' },
            { text: 'Environment Variables', link: '/guide/environment' }
          ]
        },
        {
          text: 'Advanced Topics',
          items: [
            { text: 'Advanced Workflows', link: '/guide/advanced-workflows' },
            { text: 'Intelligent Routing', link: '/guide/routing' },
            { text: 'Security', link: '/guide/security' },
            { text: 'Performance', link: '/guide/performance' },
            { text: 'Monitoring', link: '/guide/monitoring' },
            { text: 'Health Checks', link: '/guide/health-checks' },
            { text: 'Logging', link: '/guide/logging' }
          ]
        },
        {
          text: 'Development',
          items: [
            { text: 'Development Guide', link: '/guide/development' },
            { text: 'Testing', link: '/guide/testing' },
            { text: 'Contributing', link: '/guide/contributing' }
          ]
        }
      ],
      '/api/': [
        {
          text: 'API Reference',
          items: [
            { text: 'Overview', link: '/api/' },
            { text: 'Architecture', link: '/api/architecture' },
            { text: 'Messages API', link: '/api/messages' },
            { text: 'Claude Code Integration', link: '/api/claude-code' },
            { text: 'Health API', link: '/api/health' },
            { text: 'Status API', link: '/api/status' },
            { text: 'Error Handling', link: '/api/errors' }
          ]
        }
      ],
      '/providers/': [
        {
          text: 'AI Providers',
          items: [
            { text: 'Overview', link: '/providers/' },
            { text: 'OpenAI', link: '/providers/openai' },
            { text: 'Google Gemini', link: '/providers/gemini' },
            { text: 'Mistral AI', link: '/providers/mistral' },
            { text: 'Groq', link: '/providers/groq' },
            { text: 'OpenRouter', link: '/providers/openrouter' },
            { text: 'xAI', link: '/providers/xai' },
            { text: 'Ollama', link: '/providers/ollama' }
          ]
        }
      ],
      '/blog/': [
        {
          text: 'Blog Posts',
          items: [
            { text: 'All Posts', link: '/blog/' },
            { text: '🎉 v1+ Release Announcement', link: '/blog/ccproxy-release-multi-model-revolution' },
            { text: 'Qwen3 235B: Top Model', link: '/blog/qwen3-235b-top-model' },
            { text: 'OpenAI Integration', link: '/blog/openai-claude-code-integration' },
            { text: 'Google Gemini Guide', link: '/blog/google-gemini-claude-code-multimodal' },
            { text: 'Mistral AI & Privacy', link: '/blog/mistral-ai-claude-code-privacy-first' },
            { text: 'Groq Performance', link: '/blog/groq-claude-code-future-ai-development' },
            { text: 'Ollama Local AI', link: '/blog/ollama-claude-code-complete-privacy' },
            { text: 'xAI Grok Real-Time', link: '/blog/xai-grok-claude-code-real-time' },
            { text: 'Kimi K2 Guide', link: '/blog/kimi-k2-claude-code-ultimate-guide' }
          ]
        }
      ]
    },

    // Social links
    socialLinks: [
      { icon: 'github', link: 'https://github.com/orchestre-dev/ccproxy' }
    ],

    // Footer
    footer: {
      message: 'Released under the MIT License. ⭐ <a href="https://github.com/orchestre-dev/ccproxy">GitHub</a> • 💬 <a href="https://github.com/orchestre-dev/ccproxy/discussions">Join Discussions</a> • 🐛 <a href="https://github.com/orchestre-dev/ccproxy/issues">Report Issues</a>',
      copyright: 'Copyright © 2025, Made with ❤️ by <a href="https://orchestre.dev">Orchestre</a> for the Claude Code community'
    },

    // Search
    search: {
      provider: 'local',
      options: {
        placeholder: 'Search CCProxy docs...'
      }
    },

    // Edit link
    editLink: {
      pattern: 'https://github.com/orchestre-dev/ccproxy/edit/main/docs/:path',
      text: 'Edit this page on GitHub'
    }
  },

  // Mermaid configuration
  mermaid: {
    theme: 'default'
  },

  // Build configuration
  vite: {
    build: {
      chunkSizeWarningLimit: 1000
    }
  }
}))