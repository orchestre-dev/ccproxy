// Custom theme extending the default VitePress theme
import DefaultTheme from 'vitepress/theme'
import SocialShare from './components/SocialShare.vue'
import './style.css'

export default {
  ...DefaultTheme,
  enhanceApp({ app }) {
    // Register the SocialShare component globally
    app.component('SocialShare', SocialShare)
  }
}