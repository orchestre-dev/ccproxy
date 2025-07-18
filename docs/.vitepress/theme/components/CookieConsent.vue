<template>
  <Transition name="cookie-consent">
    <div v-if="showBanner" class="cookie-consent">
      <div class="cookie-consent-inner">
        <svg class="cookie-icon" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 2a10 10 0 1 0 10 10 4 4 0 0 1-5-5 4 4 0 0 1-5-5"></path>
          <path d="M8.5 8.5v.01"></path>
          <path d="M16 15.5v.01"></path>
          <path d="M12 12v.01"></path>
          <path d="M11 17v.01"></path>
          <path d="M7 14v.01"></path>
        </svg>
        <div class="cookie-consent-content">
          <p>
            We use analytics cookies to improve our documentation. 
            <a href="/privacy" rel="noopener">Learn more</a>
          </p>
        </div>
        <div class="cookie-consent-actions">
          <button @click="reject" class="cookie-consent-button secondary">
            Reject
          </button>
          <button @click="accept" class="cookie-consent-button primary">
            Accept
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const showBanner = ref(false)

const CONSENT_KEY = 'ccproxy-cookie-consent'
const CONSENT_EXPIRY_DAYS = 365

function setConsent(granted) {
  const consentData = {
    timestamp: Date.now(),
    granted: granted
  }
  localStorage.setItem(CONSENT_KEY, JSON.stringify(consentData))
  
  // Update Google Analytics consent
  if (typeof gtag !== 'undefined') {
    gtag('consent', 'update', {
      'analytics_storage': granted ? 'granted' : 'denied',
      'ad_storage': 'denied',
      'ad_user_data': 'denied',
      'ad_personalization': 'denied'
    })
  }
  
  showBanner.value = false
}

function accept() {
  setConsent(true)
}

function reject() {
  setConsent(false)
}

onMounted(() => {
  // Check if consent has been given
  const storedConsent = localStorage.getItem(CONSENT_KEY)
  
  if (storedConsent) {
    try {
      const consent = JSON.parse(storedConsent)
      const daysSinceConsent = (Date.now() - consent.timestamp) / (1000 * 60 * 60 * 24)
      
      // Re-show banner if consent is older than expiry period
      if (daysSinceConsent > CONSENT_EXPIRY_DAYS) {
        showBanner.value = true
      } else {
        // Apply stored consent
        if (typeof gtag !== 'undefined') {
          gtag('consent', 'update', {
            'analytics_storage': consent.granted ? 'granted' : 'denied',
            'ad_storage': 'denied',
            'ad_user_data': 'denied',
            'ad_personalization': 'denied'
          })
        }
      }
    } catch (e) {
      showBanner.value = true
    }
  } else {
    // First visit - show banner and set default consent mode
    showBanner.value = true
    
    if (typeof gtag !== 'undefined') {
      // Set initial consent mode to denied
      gtag('consent', 'default', {
        'analytics_storage': 'denied',
        'ad_storage': 'denied',
        'ad_user_data': 'denied',
        'ad_personalization': 'denied',
        'wait_for_update': 500
      })
    }
  }
})
</script>

<style scoped>
.cookie-consent {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 100;
  width: auto;
  min-width: 520px;
  background: var(--vp-c-bg-elv);
  backdrop-filter: blur(12px) saturate(180%);
  padding: 0.75rem 1.25rem;
  border-radius: 8px;
  border: 1px solid var(--vp-c-divider);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.1), 0 2px 6px rgba(0, 0, 0, 0.08);
}

html.dark .cookie-consent {
  background: rgba(30, 30, 32, 0.98);
  border-color: rgba(255, 255, 255, 0.1);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.4), 0 2px 6px rgba(0, 0, 0, 0.2);
}

.cookie-consent-inner {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.cookie-icon {
  color: var(--vp-c-brand);
  flex-shrink: 0;
  opacity: 0.8;
  width: 20px;
  height: 20px;
}

.cookie-consent-content {
  flex: 1;
  display: flex;
  align-items: center;
}

.cookie-consent-content p {
  margin: 0;
  color: var(--vp-c-text-1);
  font-size: 0.8125rem;
  line-height: 1.2;
}

.cookie-consent-content a {
  color: var(--vp-c-brand);
  text-decoration: none;
  font-weight: 500;
  transition: opacity 0.2s;
}

.cookie-consent-content a:hover {
  opacity: 0.8;
  text-decoration: underline;
}

.cookie-consent-actions {
  display: flex;
  gap: 0.75rem;
  margin-left: 1rem;
}

.cookie-consent-button {
  padding: 0.3rem 0.875rem;
  border: none;
  border-radius: 5px;
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
  min-width: 65px;
  text-align: center;
}

.cookie-consent-button.primary {
  background: var(--vp-c-brand);
  color: white;
}

.cookie-consent-button.primary:hover {
  background: var(--vp-c-brand-dark);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.cookie-consent-button.secondary {
  background: transparent;
  color: var(--vp-c-text-2);
  border: 1px solid var(--vp-c-divider);
}

.cookie-consent-button.secondary:hover {
  color: var(--vp-c-text-1);
  border-color: var(--vp-c-divider-dark);
  background: var(--vp-c-bg-soft);
}

/* Transition animations */
.cookie-consent-enter-active,
.cookie-consent-leave-active {
  transition: transform 0.3s ease, opacity 0.3s ease;
}

.cookie-consent-enter-from {
  transform: translateY(100%);
  opacity: 0;
}

.cookie-consent-leave-to {
  transform: translateY(100%);
  opacity: 0;
}

/* Tablet responsiveness */
@media (max-width: 768px) {
  .cookie-consent {
    min-width: auto;
    width: calc(100% - 32px);
    max-width: 480px;
  }
}

/* Mobile responsiveness */
@media (max-width: 640px) {
  .cookie-consent {
    min-width: auto;
    width: calc(100% - 24px);
    bottom: 12px;
    padding: 0.625rem 1rem;
  }
  
  .cookie-consent-inner {
    gap: 0.75rem;
  }
  
  .cookie-icon {
    width: 18px;
    height: 18px;
  }
  
  .cookie-consent-content p {
    font-size: 0.75rem;
  }
  
  .cookie-consent-content a {
    display: inline;
  }
  
  .cookie-consent-actions {
    gap: 0.5rem;
    margin-left: 0.75rem;
  }
  
  .cookie-consent-button {
    padding: 0.25rem 0.75rem;
    font-size: 0.75rem;
    min-width: 55px;
  }
}

/* Very small mobile */
@media (max-width: 380px) {
  .cookie-consent {
    padding: 0.5rem 0.75rem;
  }
  
  .cookie-consent-inner {
    flex-wrap: wrap;
    gap: 0.5rem;
  }
  
  .cookie-icon {
    display: none;
  }
  
  .cookie-consent-content {
    width: 100%;
    text-align: center;
  }
  
  .cookie-consent-actions {
    width: 100%;
    margin-left: 0;
    justify-content: center;
    gap: 0.5rem;
  }
  
  .cookie-consent-button {
    flex: 1;
    max-width: 120px;
  }
}
</style>