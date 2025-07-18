import { defineConfig } from 'vitepress'
import { loadEnv } from 'vite'
import { withMermaid } from 'vitepress-plugin-mermaid'

// Load environment variables
const env = loadEnv('', process.cwd(), '')

export default withMermaid(defineConfig({
  title: env.VITE_APP_TITLE || 'CCProxy',
  description: 'Universal AI proxy supporting Claude Code, Groq Kimi K2, OpenAI, Gemini, Mistral, XAI Grok, and Ollama. Seamless integration with any AI provider through a unified API.',
  base: process.env.BASE_URL || '/',
  ignoreDeadLinks: true,
  cleanUrls: true,
  
  // Enable sitemap generation
  sitemap: {
    hostname: 'https://ccproxy.orchestre.dev',
    transformItems: (items) => {
      // Add custom logic for sitemap items if needed
      return items.filter((item) => !item.url.includes('404'))
    }
  },
  
  head: [
    // SEO meta tags
    ['meta', { name: 'keywords', content: 'AI proxy, Claude Code, Kimi K2, Groq, OpenAI, Gemini, Mistral, XAI, Grok, Ollama, API proxy, multi-provider, artificial intelligence' }],
    ['meta', { name: 'author', content: 'CCProxy' }],
    ['meta', { property: 'og:title', content: 'CCProxy - Multi-Provider AI Proxy' }],
    ['meta', { property: 'og:description', content: 'Universal AI proxy supporting Claude Code with Kimi K2, OpenAI, Gemini, Mistral, XAI Grok, and Ollama providers.' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:url', content: 'https://ccproxy.orchestre.dev' }],
    ['meta', { property: 'og:image', content: 'https://ccproxy.orchestre.dev/og-image.png' }],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:title', content: 'CCProxy - Multi-Provider AI Proxy' }],
    ['meta', { name: 'twitter:description', content: 'Universal AI proxy for Claude Code with Kimi K2, OpenAI, Gemini, and more providers.' }],
    ['meta', { name: 'twitter:image', content: 'https://ccproxy.orchestre.dev/og-image.png' }],
    
    // Favicons
    ['link', { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }],
    ['link', { rel: 'icon', type: 'image/png', sizes: '96x96', href: '/favicon-96x96.png' }],
    ['link', { rel: 'apple-touch-icon', sizes: '180x180', href: '/apple-touch-icon.png' }],
    ['link', { rel: 'manifest', href: '/site.webmanifest' }],
    ['meta', { name: 'msapplication-config', content: '/browserconfig.xml' }],
    
    // Google Analytics with Consent Mode v2 - Configure with GA_MEASUREMENT_ID environment variable
    ...(env.GA_MEASUREMENT_ID ? [
      ['script', { async: true, src: `https://www.googletagmanager.com/gtag/js?id=${env.GA_MEASUREMENT_ID}` }],
      ['script', {}, `
        window.dataLayer = window.dataLayer || [];
        function gtag(){dataLayer.push(arguments);}
        
        // Set default consent mode before GA initialization
        gtag('consent', 'default', {
          'analytics_storage': 'denied',
          'ad_storage': 'denied',
          'ad_user_data': 'denied',
          'ad_personalization': 'denied',
          'wait_for_update': 500
        });
        
        gtag('js', new Date());
        
        // Configure GA with privacy-focused settings
        gtag('config', '${env.GA_MEASUREMENT_ID}', {
          anonymize_ip: true,
          cookie_flags: 'SameSite=None;Secure',
          allow_google_signals: false,
          allow_ad_personalization_signals: false,
          ads_data_redaction: true,
          url_passthrough: true
        });
      `]
    ] : [])
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
            { text: 'Architecture', link: '/api/architecture' },
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
      ],
      
      '/blog/': [
        {
          text: 'Latest Posts',
          items: [
            { text: 'Claude Code Reliability Challenges', link: '/blog/claude-code-reliability-challenges-solution' },
            { text: 'Kimi K2 + Claude Code Guide', link: '/blog/kimi-k2-claude-code-ultimate-guide' },
            { text: 'Groq + Claude Code Future', link: '/blog/groq-claude-code-future-ai-development' },
            { text: 'OpenAI Integration Guide', link: '/blog/openai-claude-code-integration' },
            { text: 'Gemini Multimodal AI', link: '/blog/google-gemini-claude-code-multimodal' }
          ]
        },
        {
          text: 'Provider Deep Dives',
          items: [
            { text: 'Mistral AI Privacy-First', link: '/blog/mistral-ai-claude-code-privacy-first' },
            { text: 'XAI Grok Real-Time', link: '/blog/xai-grok-claude-code-real-time' },
            { text: 'Ollama Complete Privacy', link: '/blog/ollama-claude-code-complete-privacy' }
          ]
        },
        {
          text: 'Categories',
          items: [
            { text: 'Performance & Speed', link: '/blog/#performance-speed' },
            { text: 'Privacy & Security', link: '/blog/#privacy-security' },
            { text: 'AI Integration', link: '/blog/#ai-integration' },
            { text: 'Multimodal AI', link: '/blog/#multimodal' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/orchestre-dev/ccproxy' }
    ],

    footer: {
      message: 'Released under the MIT License. <a href="https://github.com/orchestre-dev/ccproxy" target="_blank">⭐ GitHub</a> • <a href="https://github.com/orchestre-dev/ccproxy/discussions" target="_blank">💬 Join Discussions</a> • <a href="https://github.com/orchestre-dev/ccproxy/issues" target="_blank">🐛 Report Issues</a>',
      copyright: 'Copyright © 2025, Made with ❤️ by Orchestre for the Claude code community'
    },

    editLink: {
      pattern: 'https://github.com/orchestre-dev/ccproxy/edit/main/docs/:path',
      text: 'Edit this page on GitHub'
    }
  },

  markdown: {
    theme: {
      light: 'github-light',
      dark: 'github-dark'
    },
    lineNumbers: true
  },

  // Mermaid configuration
  mermaid: {
    theme: 'default',
    darkTheme: 'dark'
  }
}))