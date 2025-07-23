// Custom theme extending the default VitePress theme
import DefaultTheme from 'vitepress/theme'
import SocialShare from './components/SocialShare.vue'
import CookieConsent from './components/CookieConsent.vue'
import Layout from './components/Layout.vue'
import Mermaid from 'vitepress-plugin-mermaid/Mermaid.vue'
import NewsletterForm from './components/NewsletterForm.vue'
import AnalyticsTracker from './components/AnalyticsTracker.vue'
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
    // Register the NewsletterForm component globally
    app.component('NewsletterForm', NewsletterForm)
    // Register the AnalyticsTracker component globally
    app.component('AnalyticsTracker', AnalyticsTracker)
  },
  Layout
}