import { onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vitepress'

// Analytics helper composable
export function useAnalytics() {
  const route = useRoute()
  
  // Check if gtag is available
  const isAnalyticsAvailable = () => {
    return typeof window !== 'undefined' && typeof window.gtag === 'function'
  }

  // Track page view
  const trackPageView = (pagePath, pageTitle) => {
    if (!isAnalyticsAvailable()) return
    
    window.gtag('event', 'page_view', {
      page_path: pagePath || route.path,
      page_title: pageTitle || document.title,
      page_location: window.location.href
    })
  }

  // Track custom events
  const trackEvent = (eventName, parameters = {}) => {
    if (!isAnalyticsAvailable()) return
    
    window.gtag('event', eventName, {
      ...parameters,
      page_path: route.path
    })
  }

  // Track click events
  const trackClick = (category, label, value = 1) => {
    trackEvent('click', {
      event_category: category,
      event_label: label,
      value: value
    })
  }

  // Track download events
  const trackDownload = (fileName, fileType) => {
    trackEvent('file_download', {
      file_name: fileName,
      file_extension: fileType,
      link_text: fileName
    })
  }

  // Track social share events
  const trackSocialShare = (method, contentType = 'article') => {
    trackEvent('share', {
      method: method,
      content_type: contentType,
      item_id: route.path
    })
  }

  // Track scroll depth
  const trackScrollDepth = () => {
    if (!isAnalyticsAvailable()) return
    
    let scrollDepths = [25, 50, 75, 90, 100]
    let scrolledDepths = new Set()
    
    const calculateScrollDepth = () => {
      const windowHeight = window.innerHeight
      const documentHeight = document.documentElement.scrollHeight
      const scrollTop = window.scrollY || document.documentElement.scrollTop
      const scrollPercentage = Math.round((scrollTop + windowHeight) / documentHeight * 100)
      
      scrollDepths.forEach(depth => {
        if (scrollPercentage >= depth && !scrolledDepths.has(depth)) {
          scrolledDepths.add(depth)
          trackEvent('scroll', {
            percent_scrolled: depth,
            page_path: route.path
          })
        }
      })
    }
    
    const throttledScroll = throttle(calculateScrollDepth, 500)
    
    onMounted(() => {
      window.addEventListener('scroll', throttledScroll)
    })
    
    onUnmounted(() => {
      window.removeEventListener('scroll', throttledScroll)
    })
  }

  // Track time on page
  const trackTimeOnPage = () => {
    if (!isAnalyticsAvailable()) return
    
    let startTime = Date.now()
    let isVisible = true
    
    const handleVisibilityChange = () => {
      if (document.hidden) {
        isVisible = false
        const timeSpent = Math.round((Date.now() - startTime) / 1000)
        trackEvent('timing_complete', {
          name: 'time_on_page',
          value: timeSpent,
          event_category: 'engagement'
        })
      } else {
        isVisible = true
        startTime = Date.now()
      }
    }
    
    onMounted(() => {
      document.addEventListener('visibilitychange', handleVisibilityChange)
    })
    
    onUnmounted(() => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
      if (isVisible) {
        const timeSpent = Math.round((Date.now() - startTime) / 1000)
        trackEvent('timing_complete', {
          name: 'time_on_page',
          value: timeSpent,
          event_category: 'engagement'
        })
      }
    })
  }

  // Track code copy events
  const trackCodeCopy = (codeLanguage = 'unknown') => {
    trackEvent('copy_code', {
      code_language: codeLanguage,
      event_category: 'engagement'
    })
  }

  // Track form submissions
  const trackFormSubmit = (formName, formData = {}) => {
    trackEvent('generate_lead', {
      form_name: formName,
      ...formData
    })
  }

  // Track search
  const trackSearch = (searchTerm) => {
    trackEvent('search', {
      search_term: searchTerm
    })
  }

  // Track video engagement
  const trackVideo = (action, videoTitle, videoProvider = 'youtube') => {
    trackEvent(`video_${action}`, {
      video_title: videoTitle,
      video_provider: videoProvider
    })
  }

  // Track outbound links
  const trackOutboundLink = (url, linkText) => {
    trackEvent('click', {
      event_category: 'outbound',
      event_label: url,
      link_text: linkText,
      link_url: url,
      outbound: true
    })
  }

  // Track CTA clicks
  const trackCTA = (ctaName, ctaLocation, ctaDestination) => {
    trackEvent('select_content', {
      content_type: 'cta',
      item_id: ctaName,
      location: ctaLocation,
      destination: ctaDestination
    })
  }

  // Track errors
  const trackError = (errorMessage, errorSource) => {
    trackEvent('exception', {
      description: errorMessage,
      fatal: false,
      error_source: errorSource
    })
  }

  // Enhanced ecommerce - track conversions
  const trackConversion = (conversionType, value = 0, currency = 'USD') => {
    trackEvent('conversion', {
      conversion_type: conversionType,
      value: value,
      currency: currency,
      transaction_id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    })
  }

  return {
    trackPageView,
    trackEvent,
    trackClick,
    trackDownload,
    trackSocialShare,
    trackScrollDepth,
    trackTimeOnPage,
    trackCodeCopy,
    trackFormSubmit,
    trackSearch,
    trackVideo,
    trackOutboundLink,
    trackCTA,
    trackError,
    trackConversion
  }
}

// Utility function to throttle scroll events
function throttle(func, delay) {
  let timeoutId
  let lastExecTime = 0
  
  return function (...args) {
    const currentTime = Date.now()
    
    if (currentTime - lastExecTime > delay) {
      func.apply(this, args)
      lastExecTime = currentTime
    } else {
      clearTimeout(timeoutId)
      timeoutId = setTimeout(() => {
        func.apply(this, args)
        lastExecTime = Date.now()
      }, delay - (currentTime - lastExecTime))
    }
  }
}