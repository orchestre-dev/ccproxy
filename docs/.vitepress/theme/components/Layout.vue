<template>
  <DefaultTheme.Layout>
    <template #layout-bottom>
      <CookieConsent />
      <AnalyticsTracker />
    </template>
  </DefaultTheme.Layout>
</template>

<script setup>
import DefaultTheme from 'vitepress/theme'
import CookieConsent from './CookieConsent.vue'
import AnalyticsTracker from './AnalyticsTracker.vue'
import { useAnalytics } from '../composables/useAnalytics'
import { useRoute } from 'vitepress'
import { watch, onMounted, onUnmounted, nextTick } from 'vue'

const route = useRoute()
const { trackPageView, trackScrollDepth, trackTimeOnPage, trackCodeCopy, trackOutboundLink } = useAnalytics()

// Track initial page view
onMounted(() => {
  trackPageView()
  
  // Enable scroll depth tracking for blog posts
  if (route.path.startsWith('/blog/')) {
    trackScrollDepth()
    trackTimeOnPage()
  }
})

// Track page views on route change
watch(() => route.path, (newPath) => {
  nextTick(() => {
    trackPageView(newPath)
    
    // Enable scroll depth tracking for blog posts
    if (newPath.startsWith('/blog/')) {
      trackScrollDepth()
      trackTimeOnPage()
    }
  })
})

// Global click handler for tracking
const clickHandler = (e) => {
  const target = e.target.closest('a')
  if (!target) return
  
  const href = target.getAttribute('href')
  if (!href) return
  
  // Track outbound links
  if (href.startsWith('http') && !href.includes('ccproxy.orchestre.dev')) {
    trackOutboundLink(href, target.textContent || 'Unknown')
  }
  
  // Track code copy buttons
  const copyButton = e.target.closest('.copy')
  if (copyButton) {
    const codeBlock = copyButton.closest('.language-')
    const language = codeBlock ? codeBlock.className.match(/language-(\w+)/)?.[1] : 'unknown'
    trackCodeCopy(language)
  }
}

onMounted(() => {
  document.addEventListener('click', clickHandler)
})

onUnmounted(() => {
  document.removeEventListener('click', clickHandler)
})
</script>