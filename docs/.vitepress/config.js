import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'CCProxy',
  description: 'Universal AI proxy supporting Claude Code, Groq Kimi K2, OpenAI, Gemini, Mistral, XAI Grok, and Ollama. Seamless integration with any AI provider through a unified API.',
  ignoreDeadLinks: true,
  
  head: [
    // SEO meta tags
    ['meta', { name: 'keywords', content: 'AI proxy, Claude Code, Kimi K2, Groq, OpenAI, Gemini, Mistral, XAI, Grok, Ollama, API proxy, multi-provider, artificial intelligence' }],
    ['meta', { name: 'author', content: 'CCProxy' }],
    ['meta', { property: 'og:title', content: 'CCProxy - Multi-Provider AI Proxy' }],
    ['meta', { property: 'og:description', content: 'Universal AI proxy supporting Claude Code with Kimi K2, OpenAI, Gemini, Mistral, XAI Grok, and Ollama providers.' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:image', content: '/ccproxy_icon.png' }],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:title', content: 'CCProxy - Multi-Provider AI Proxy' }],
    ['meta', { name: 'twitter:description', content: 'Universal AI proxy for Claude Code with Kimi K2, OpenAI, Gemini, and more providers.' }],
    ['meta', { name: 'twitter:image', content: '/ccproxy_icon.png' }],
    
    // Favicons
    ['link', { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }],
    ['link', { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/favicon-32x32.png' }],
    ['link', { rel: 'icon', type: 'image/png', sizes: '16x16', href: '/favicon-16x16.png' }],
    ['link', { rel: 'apple-touch-icon', sizes: '180x180', href: '/apple-icon-180x180.png' }],
    ['link', { rel: 'manifest', href: '/manifest.json' }],
    ['meta', { name: 'msapplication-config', content: '/browserconfig.xml' }]
  ],

  themeConfig: {
    logo: '/ccproxy_icon.png',
    
    search: {
      provider: 'local',
      options: {
        translations: {
          button: {
            buttonText: 'Search docs',
            buttonAriaLabel: 'Search documentation'
          },
          modal: {
            displayDetails: 'Display detailed list',
            resetButtonTitle: 'Reset search',
            backButtonTitle: 'Close search',
            noResultsText: 'No results for',
            footer: {
              selectText: 'to select',
              navigateText: 'to navigate',
              closeText: 'to close'
            }
          }
        }
      }
    },
    
    nav: [
      { text: 'Guide', link: '/guide/' },
      { text: 'Providers', link: '/providers/' },
      { text: 'Kimi K2', link: '/kimi-k2' },
      { text: 'API', link: '/api/' },
      { 
        text: 'Community', 
        items: [
          { text: 'GitHub', link: 'https://github.com/praneybehl/ccproxy' },
          { text: 'Discussions', link: 'https://github.com/praneybehl/ccproxy/discussions' },
          { text: 'Issues & Bug Reports', link: 'https://github.com/praneybehl/ccproxy/issues' },
          { text: 'Feature Requests', link: 'https://github.com/praneybehl/ccproxy/issues/new?template=feature_request.md' }
        ]
      }
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
          text: 'Featured',
          items: [
            { text: '🚀 Kimi K2 Integration', link: '/kimi-k2' }
          ]
        },
        {
          text: 'Advanced',
          items: [
            { text: 'Advanced Workflows', link: '/guide/advanced-workflows' },
            { text: 'Docker Deployment', link: '/guide/docker' },
            { text: 'Environment Variables', link: '/guide/environment' },
            { text: 'Logging', link: '/guide/logging' },
            { text: 'Health Checks', link: '/guide/health-checks' }
          ]
        }
      ],
      
      '/providers/': [
        {
          text: 'Featured Provider',
          items: [
            { text: '⚡ Kimi K2 (Groq)', link: '/kimi-k2' }
          ]
        },
        {
          text: 'All AI Providers',
          items: [
            { text: 'Overview', link: '/providers/' },
            { text: 'Groq (Kimi K2)', link: '/providers/groq' },
            { text: 'OpenRouter', link: '/providers/openrouter' },
            { text: 'OpenAI GPT', link: '/providers/openai' },
            { text: 'XAI Grok', link: '/providers/xai' },
            { text: 'Google Gemini', link: '/providers/gemini' },
            { text: 'Mistral AI', link: '/providers/mistral' },
            { text: 'Ollama Local', link: '/providers/ollama' }
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
      ],
      
      '/kimi-k2': [
        {
          text: 'Kimi K2 Guide',
          items: [
            { text: '🚀 Overview', link: '/kimi-k2' },
            { text: 'Groq Setup', link: '/providers/groq' },
            { text: 'OpenRouter Setup', link: '/providers/openrouter' },
            { text: 'Quick Start', link: '/guide/quick-start' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/praneybehl/ccproxy' }
    ],

    footer: {
      message: 'Released under the MIT License. <a href="https://github.com/praneybehl/ccproxy" target="_blank">⭐ GitHub</a> • <a href="https://github.com/praneybehl/ccproxy/discussions" target="_blank">💬 Join Discussions</a> • <a href="https://github.com/praneybehl/ccproxy/issues" target="_blank">🐛 Report Issues</a>',
      copyright: 'Copyright © 2025 Praney Behl - Universal AI Proxy for Claude Code'
    },

    editLink: {
      pattern: 'https://github.com/praneybehl/ccproxy/edit/main/docs/:path',
      text: 'Edit this page on GitHub'
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