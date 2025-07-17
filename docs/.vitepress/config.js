import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'CCProxy',
  description: 'Multi-Provider AI Proxy for Claude Code',
  base: '/ccproxy/',
  
  head: [
    ['link', { rel: 'icon', href: '/ccproxy/favicon.ico' }]
  ],

  themeConfig: {
    logo: '/logo.svg',
    
    nav: [
      { text: 'Guide', link: '/guide/' },
      { text: 'Providers', link: '/providers/' },
      { text: 'API', link: '/api/' },
      { text: 'GitHub', link: 'https://github.com/praneybehl/ccproxy' }
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Getting Started',
          items: [
            { text: 'Introduction', link: '/guide/' },
            { text: 'Quick Start', link: '/guide/quick-start' },
            { text: 'Installation', link: '/guide/installation' },
            { text: 'Configuration', link: '/guide/configuration' }
          ]
        },
        {
          text: 'Advanced',
          items: [
            { text: 'Docker Deployment', link: '/guide/docker' },
            { text: 'Environment Variables', link: '/guide/environment' },
            { text: 'Logging', link: '/guide/logging' },
            { text: 'Health Checks', link: '/guide/health-checks' }
          ]
        }
      ],
      
      '/providers/': [
        {
          text: 'Supported Providers',
          items: [
            { text: 'Overview', link: '/providers/' },
            { text: 'Groq', link: '/providers/groq' },
            { text: 'OpenRouter', link: '/providers/openrouter' },
            { text: 'OpenAI', link: '/providers/openai' },
            { text: 'XAI (Grok)', link: '/providers/xai' },
            { text: 'Google Gemini', link: '/providers/gemini' },
            { text: 'Mistral AI', link: '/providers/mistral' },
            { text: 'Ollama', link: '/providers/ollama' }
          ]
        },
        {
          text: 'Provider Comparison',
          items: [
            { text: 'Feature Matrix', link: '/providers/comparison' },
            { text: 'Performance', link: '/providers/performance' },
            { text: 'Cost Analysis', link: '/providers/costs' }
          ]
        }
      ],
      
      '/api/': [
        {
          text: 'API Reference',
          items: [
            { text: 'Overview', link: '/api/' },
            { text: 'Messages Endpoint', link: '/api/messages' },
            { text: 'Health Endpoints', link: '/api/health' },
            { text: 'Status Endpoint', link: '/api/status' }
          ]
        },
        {
          text: 'Integration',
          items: [
            { text: 'Claude Code', link: '/api/claude-code' },
            { text: 'Error Handling', link: '/api/errors' },
            { text: 'Rate Limiting', link: '/api/rate-limits' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/praneybehl/ccproxy' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2025 Pranay Behl'
    },

    editLink: {
      pattern: 'https://github.com/praneybehl/ccproxy/edit/main/docs/:path',
      text: 'Edit this page on GitHub'
    },

    search: {
      provider: 'local'
    }
  },

  markdown: {
    theme: {
      light: 'github-light',
      dark: 'github-dark'
    },
    lineNumbers: true
  }
})