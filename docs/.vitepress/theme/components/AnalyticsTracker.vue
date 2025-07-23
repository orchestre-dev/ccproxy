<template>
  <div class="analytics-tracker" style="display: none;"></div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRoute } from 'vitepress'
import { useAnalytics } from '../composables/useAnalytics'

const route = useRoute()
const { trackEvent, trackCTA, trackDownload, trackConversion } = useAnalytics()

onMounted(() => {
  // Track installation commands
  const installCommands = [
    { selector: 'code:contains("curl -sSL")', event: 'install_command_view', type: 'unix' },
    { selector: 'code:contains("irm https://raw.githubusercontent.com")', event: 'install_command_view', type: 'windows' }
  ]
  
  installCommands.forEach(({ selector, event, type }) => {
    const elements = document.querySelectorAll('pre code')
    elements.forEach(el => {
      if (el.textContent?.includes(selector.match(/contains\("(.+)"\)/)?.[1] || '')) {
        trackEvent(event, {
          install_type: type,
          page_path: route.path
        })
      }
    })
  })
  
  // Enhanced click tracking for CTAs
  document.addEventListener('click', (e) => {
    const target = e.target
    
    // Track download links
    if (target.href?.includes('github.com/orchestre-dev/ccproxy/releases')) {
      const fileName = target.href.split('/').pop() || 'ccproxy-release'
      trackDownload(fileName, 'binary')
      trackConversion('download_binary', 0)
    }
    
    // Track "Get Started" CTAs
    if (target.textContent?.toLowerCase().includes('get started') || 
        target.href?.includes('/guide/')) {
      trackCTA('get_started', route.path, target.href || '/guide/')
    }
    
    // Track API key signup links
    if (target.href?.includes('openai.com') || 
        target.href?.includes('anthropic.com') ||
        target.href?.includes('aistudio.google.com')) {
      trackCTA('api_key_signup', route.path, target.href)
    }
    
    // Track documentation navigation
    if (target.href?.includes('/providers/') || 
        target.href?.includes('/api/') ||
        target.href?.includes('/guide/')) {
      trackEvent('docs_navigation', {
        from_page: route.path,
        to_page: target.href,
        link_text: target.textContent
      })
    }
  })
  
  // Track copy button clicks for installation commands
  const observer = new MutationObserver(() => {
    const copyButtons = document.querySelectorAll('.copy')
    copyButtons.forEach(button => {
      if (!button.dataset.analyticsTracked) {
        button.dataset.analyticsTracked = 'true'
        button.addEventListener('click', () => {
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
        })
      }
    })
  })
  
  observer.observe(document.body, { childList: true, subtree: true })
})
</script>