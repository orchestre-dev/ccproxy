<template>
  <div class="early-access-section">
    <div class="early-access-container">
      <div class="header-content">
        <span class="icon">📬</span>
        <h3>Get Updates</h3>
        <span class="divider">•</span>
        <span class="tagline">Stay informed about new features and providers</span>
      </div>
      
      <form 
        v-if="!submitted"
        @submit.prevent="handleSubmit"
        class="early-access-form"
      >
        <input 
          type="text" 
          id="name"
          v-model="formData.name"
          placeholder="Your name" 
          class="form-input"
          required
        />
        
        <input 
          type="email" 
          id="email"
          v-model="formData.email"
          placeholder="your@email.com" 
          class="form-input"
          required
        />
        
        <input type="hidden" name="_subject" value="Newsletter Signup for CCProxy" />
        
        <button type="submit" class="submit-button" :disabled="loading">
          <span v-if="!loading">Subscribe</span>
          <span v-else>...</span>
        </button>
      </form>
      
      <p class="fine-print" v-if="!submitted">
        🤝 We promise to only send you the good stuff. No spam, just pure CCProxy goodness.
      </p>
      
      <div v-else class="success-message">
        <span>✅</span>
        <span>You're on the list! We'll be in touch soon.</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAnalytics } from '../composables/useAnalytics'

const formData = ref({
  name: '',
  email: ''
})

const loading = ref(false)
const submitted = ref(false)
const { trackFormSubmit, trackConversion } = useAnalytics()

const handleSubmit = async () => {
  loading.value = true
  
  const formDataToSend = new FormData()
  formDataToSend.append('name', formData.value.name)
  formDataToSend.append('email', formData.value.email)
  formDataToSend.append('_subject', 'New Newsletter Signup for CCProxy')
  
  try {
    const response = await fetch('https://formspree.io/f/xyzpeyed', {
      method: 'POST',
      body: formDataToSend,
      headers: {
        'Accept': 'application/json'
      }
    })
    
    if (response.ok) {
      submitted.value = true
      
      // Track form submission
      trackFormSubmit('newsletter_signup', {
        form_location: 'blog_post',
        form_type: 'newsletter'
      })
      
      // Track as conversion
      trackConversion('newsletter_signup', 0)
    } else {
      throw new Error('Form submission failed')
    }
  } catch (error) {
    alert('Sorry, there was an error. Please try again or email us directly.')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.early-access-section {
  margin: 40px auto;
  max-width: 1152px;
}

.early-access-container {
  background: var(--vp-c-bg-alt);
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  padding: 20px 24px;
  transition: all 0.25s ease;
}

.early-access-container:hover {
  border-color: var(--vp-c-brand-1);
  transform: translateY(-1px);
}

.header-content {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.icon {
  font-size: 24px;
}

.header-content h3 {
  font-size: 20px;
  font-weight: 700;
  margin: 0;
  color: var(--vp-c-text-1);
}

.divider {
  color: var(--vp-c-divider);
  font-size: 12px;
}

.tagline {
  font-size: 16px;
  color: var(--vp-c-text-2);
}

.early-access-form {
  display: flex;
  gap: 10px;
  align-items: center;
}

.form-input {
  flex: 1;
  min-width: 140px;
  padding: 10px 16px;
  font-size: 16px;
  border: 2px solid var(--vp-c-brand-1);
  border-radius: 9999px;
  background: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  transition: all 0.2s ease;
  opacity: 0.7;
}

.form-input:hover {
  opacity: 0.85;
}

.form-input:focus {
  outline: none;
  border-color: var(--vp-c-brand-1);
  background: var(--vp-c-bg-elv);
  opacity: 1;
}

.form-input::placeholder {
  color: var(--vp-c-text-3);
  font-size: 16px;
}

.submit-button {
  padding: 10px 24px;
  font-size: 16px;
  font-weight: 500;
  background-color: var(--vp-button-brand-bg);
  color: var(--vp-button-brand-text);
  border: 1px solid var(--vp-button-brand-bg);
  border-radius: 24px;
  cursor: pointer;
  transition: all 0.25s ease;
  white-space: nowrap;
}

.submit-button:hover:not(:disabled) {
  background-color: var(--vp-button-brand-hover-bg);
  border-color: var(--vp-button-brand-hover-bg);
  color: var(--vp-button-brand-text);
  transform: translateY(-1px);
}

.submit-button:active {
  transform: translateY(0);
}

.submit-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.success-message {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #2563eb;
  font-weight: 500;
  font-size: 16px;
}

.fine-print {
  margin-top: 12px;
  font-size: 14px;
  color: var(--vp-c-text-3);
  text-align: center;
  line-height: 1.5;
}

/* Mobile responsive */
@media (max-width: 768px) {
  .early-access-section {
    margin: 32px auto;
  }
  
  .early-access-container {
    padding: 16px;
  }
  
  .divider,
  .tagline {
    display: none;
  }
  
  .early-access-form {
    flex-direction: column;
    gap: 8px;
  }
  
  .form-input {
    width: 100%;
  }
  
  .submit-button {
    width: 100%;
    padding: 10px;
  }
}

/* Ultra compact on very small screens */
@media (max-width: 480px) {
  .header-content {
    margin-bottom: 10px;
  }
  
  .icon {
    font-size: 16px;
  }
  
  .header-content h3 {
    font-size: 18px;
  }
}

/* Dark mode adjustments */
.dark .early-access-container {
  background: var(--vp-c-bg-elv);
}

.dark .early-access-container:hover {
  border-color: var(--vp-c-brand-1);
}

.dark .form-input {
  background: var(--vp-c-bg);
  border-color: var(--vp-c-brand-1);
}

.dark .form-input:focus {
  background: var(--vp-c-bg-soft);
  border-color: var(--vp-c-brand-1);
}

/* Button styles now use VitePress theme variables which handle dark mode automatically */

.dark .success-message {
  color: #60a5fa;
}
</style>