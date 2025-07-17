import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'CCProxy - Multi-Provider AI Proxy',
  description: 'Universal AI proxy supporting Claude Code, Groq Kimi K2, OpenAI, Gemini, Mistral, XAI Grok, and Ollama. Seamless integration with any AI provider through a unified API.',
  base: '/ccproxy/',
  ignoreDeadLinks: true,
  
  head: [
    // SEO meta tags
    ['meta', { name: 'keywords', content: 'AI proxy, Claude Code, Kimi K2, Groq, OpenAI, Gemini, Mistral, XAI, Grok, Ollama, API proxy, multi-provider, artificial intelligence' }],
    ['meta', { name: 'author', content: 'CCProxy' }],
    ['meta', { property: 'og:title', content: 'CCProxy - Multi-Provider AI Proxy' }],
    ['meta', { property: 'og:description', content: 'Universal AI proxy supporting Claude Code with Groq Kimi K2, OpenAI, Gemini, Mistral, XAI Grok, and Ollama providers.' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:image', content: '/ccproxy/ccproxy_icon.png' }],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:title', content: 'CCProxy - Multi-Provider AI Proxy' }],
    ['meta', { name: 'twitter:description', content: 'Universal AI proxy for Claude Code with Kimi K2, OpenAI, Gemini, and more providers.' }],
    ['meta', { name: 'twitter:image', content: '/ccproxy/ccproxy_icon.png' }],
    
    // Favicons
    ['link', { rel: 'icon', type: 'image/x-icon', href: '/ccproxy/favicon.ico' }],
    ['link', { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/ccproxy/favicon-32x32.png' }],
    ['link', { rel: 'icon', type: 'image/png', sizes: '16x16', href: '/ccproxy/favicon-16x16.png' }],
    ['link', { rel: 'apple-touch-icon', sizes: '180x180', href: '/ccproxy/apple-icon-180x180.png' }],
    ['link', { rel: 'manifest', href: '/ccproxy/manifest.json' }],
    ['meta', { name: 'msapplication-config', content: '/ccproxy/browserconfig.xml' }],
    
    // Custom CSS for glow effect and social sharing
    ['style', {}, `
      .VPHero .VPImage {
        filter: drop-shadow(0 0 20px rgba(0, 255, 127, 0.3));
        transition: filter 0.3s ease;
      }
      .VPHero .VPImage:hover {
        filter: drop-shadow(0 0 30px rgba(0, 255, 127, 0.5));
      }
      .social-share {
        display: flex;
        gap: 10px;
        margin: 20px 0;
        align-items: center;
        flex-wrap: wrap;
      }
      .social-share button {
        padding: 8px 16px;
        border: none;
        border-radius: 6px;
        cursor: pointer;
        font-size: 14px;
        font-weight: 500;
        transition: all 0.2s;
        display: flex;
        align-items: center;
        gap: 6px;
      }
      .share-twitter { background: #1da1f2; color: white; }
      .share-linkedin { background: #0077b5; color: white; }
      .share-reddit { background: #ff4500; color: white; }
      .share-copy { background: #6b7280; color: white; }
      .social-share button:hover { 
        transform: translateY(-1px); 
        opacity: 0.9;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
      }
      .showcase-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
        gap: 20px;
        margin: 40px 0;
      }
      .showcase-item {
        background: var(--vp-c-bg-soft);
        border: 1px solid var(--vp-c-border);
        border-radius: 12px;
        padding: 24px;
        transition: all 0.3s;
      }
      .showcase-item:hover {
        border-color: var(--vp-c-brand-1);
        box-shadow: 0 8px 32px rgba(0, 255, 127, 0.1);
        transform: translateY(-2px);
      }
      .showcase-title {
        font-size: 18px;
        font-weight: 600;
        margin-bottom: 12px;
        color: var(--vp-c-brand-1);
      }
      .showcase-description {
        color: var(--vp-c-text-2);
        line-height: 1.6;
        margin-bottom: 16px;
      }
      .showcase-link {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        color: var(--vp-c-brand-1);
        text-decoration: none;
        font-weight: 500;
        transition: color 0.2s;
      }
      .showcase-link:hover {
        color: var(--vp-c-brand-2);
      }
    `]
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
          text: 'Featured',
          items: [
            { text: '🚀 Kimi K2 Integration', link: '/kimi-k2' }
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
      message: 'Released under the MIT License.',
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
  },
  
  // Custom CSS variables for lime green/electric blue theme
  vite: {
    css: {
      preprocessorOptions: {
        scss: {
          additionalData: `
            :root {
              --vp-c-brand-1: #00ff7f;
              --vp-c-brand-2: #00e86b;
              --vp-c-brand-3: #00d660;
              --vp-c-brand-soft: rgba(0, 255, 127, 0.14);
              --vp-c-brand-softer: rgba(0, 255, 127, 0.07);
              --vp-c-brand-softest: rgba(0, 255, 127, 0.04);
            }
            
            .dark {
              --vp-c-brand-1: #00ffff;
              --vp-c-brand-2: #00e6e6;
              --vp-c-brand-3: #00cccc;
              --vp-c-brand-soft: rgba(0, 255, 255, 0.16);
              --vp-c-brand-softer: rgba(0, 255, 255, 0.08);
              --vp-c-brand-softest: rgba(0, 255, 255, 0.04);
            }
          `
        }
      }
    }
  }
})