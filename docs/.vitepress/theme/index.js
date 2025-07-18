// Custom theme extending the default VitePress theme
import DefaultTheme from 'vitepress/theme'
import SocialShare from './components/SocialShare.vue'
import CookieConsent from './components/CookieConsent.vue'
import Layout from './components/Layout.vue'
import Mermaid from 'vitepress-plugin-mermaid/Mermaid.vue'
import './style.css'

export default {
  ...DefaultTheme,
  enhanceApp({ app }) {
    // Register the SocialShare component globally
    app.component('SocialShare', SocialShare)
    // Register the CookieConsent component globally
    app.component('CookieConsent', CookieConsent)
    // Register the Mermaid component globally
    app.component('Mermaid', Mermaid)
  },
  Layout
}