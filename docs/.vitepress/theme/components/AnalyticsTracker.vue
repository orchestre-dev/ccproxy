<template>
  <div class="analytics-tracker" style="display: none;"></div>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vitepress'
import { useAnalytics } from '../composables/useAnalytics'

const route = useRoute()
const { trackEvent, trackCTA, trackDownload, trackConversion } = useAnalytics()

// Store references for cleanup
let observer = null
const copyButtonHandlers = new WeakMap()

// Optimized selectors
const installSelectors = ['curl -sSL', 'irm https://raw.githubusercontent.com']

onMounted(() => {
  // Track installation commands (optimized)
  const codeElements = document.querySelectorAll('pre code')
  const installTypes = { 'curl -sSL': 'unix', 'irm https://raw.githubusercontent.com': 'windows' }
  
  codeElements.forEach(el => {
    const text = el.textContent || ''
    for (const [selector, type] of Object.entries(installTypes)) {
      if (text.includes(selector)) {
        trackEvent('install_command_view', {
          install_type: type,
          page_path: route.path
        })
        break
      }
    }
  })
  
  // Enhanced click tracking for CTAs
  const clickHandler = (e) => {
    const target = e.target.closest('a')
    if (!target || !target.href) return
    
    // Track download links
    if (target.href.includes('github.com/orchestre-dev/ccproxy/releases')) {
      const fileName = target.href.split('/').pop() || 'ccproxy-release'
      trackDownload(fileName, 'binary')
      trackConversion('download_binary', 0)
    }
    
    // Track "Get Started" CTAs
    if (target.textContent?.toLowerCase().includes('get started') || 
        target.href.includes('/guide/')) {
      trackCTA('get_started', route.path, target.href)
    }
    
    // Track API key signup links
    if (target.href.includes('openai.com') || 
        target.href.includes('anthropic.com') ||
        target.href.includes('aistudio.google.com')) {
      trackCTA('api_key_signup', route.path, target.href)
    }
    
    // Track documentation navigation
    if (target.href.includes('/providers/') || 
        target.href.includes('/api/') ||
        target.href.includes('/guide/')) {
      trackEvent('docs_navigation', {
        from_page: route.path,
        to_page: target.href,
        link_text: target.textContent
      })
    }
  }
  
  document.addEventListener('click', clickHandler)
  
  // Track copy button clicks for installation commands
  const copyButtonHandler = (button) => {
    const handler = () => {
      const codeBlock = button.closest('div[class*="language-"]')
      const language = codeBlock?.className.match(/language-(\w+)/)?.[1] || 'unknown'
      const code = codeBlock?.querySelector('code')?.textContent || ''
      
      // Special tracking for installation commands
      if (code.includes('curl -sSL') && code.includes('install.sh')) {
        trackConversion('install_copy_unix', 0)
        trackEvent('install_command_copy', {
          install_type: 'unix',
          method: 'curl'
        })
      } else if (code.includes('irm') && code.includes('install.ps1')) {
        trackConversion('install_copy_windows', 0)
        trackEvent('install_command_copy', {
          install_type: 'windows',
          method: 'powershell'
        })
      }
    }
    
    button.addEventListener('click', handler)
    copyButtonHandlers.set(button, handler)
  }
  
  observer = new MutationObserver(() => {
    const copyButtons = document.querySelectorAll('.copy')
    copyButtons.forEach(button => {
      if (!button.dataset.analyticsTracked) {
        button.dataset.analyticsTracked = 'true'
        copyButtonHandler(button)
      }
    })
  })
  
  observer.observe(document.body, { childList: true, subtree: true })
  
  // Cleanup function
  onUnmounted(() => {
    document.removeEventListener('click', clickHandler)
    
    if (observer) {
      observer.disconnect()
    }
    
    // Remove all copy button handlers
    const copyButtons = document.querySelectorAll('.copy[data-analytics-tracked]')
    copyButtons.forEach(button => {
      const handler = copyButtonHandlers.get(button)
      if (handler) {
        button.removeEventListener('click', handler)
        copyButtonHandlers.delete(button)
      }
    })
  })
})
</script>