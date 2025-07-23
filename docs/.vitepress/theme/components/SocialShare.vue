<template>
  <div class="social-share">
    <span class="share-label">Share this page:</span>
    <div class="share-buttons">
      <button 
        class="share-btn twitter"
        @click="shareToTwitter"
        :title="twitterTooltip"
        aria-label="Share on Twitter"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
          <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
        </svg>
      </button>
      
      <button 
        class="share-btn linkedin"
        @click="shareToLinkedIn"
        title="Share on LinkedIn"
        aria-label="Share on LinkedIn"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
          <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/>
        </svg>
      </button>
      
      <button 
        class="share-btn reddit"
        @click="shareToReddit"
        title="Share on Reddit"
        aria-label="Share on Reddit"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0zm5.01 4.744c.688 0 1.25.561 1.25 1.249a1.25 1.25 0 0 1-2.498.056l-2.597-.547-.8 3.747c1.824.07 3.48.632 4.674 1.488.308-.309.73-.491 1.207-.491.968 0 1.754.786 1.754 1.754 0 .716-.435 1.333-1.01 1.614a3.111 3.111 0 0 1 .042.52c0 2.694-3.13 4.87-7.004 4.87-3.874 0-7.004-2.176-7.004-4.87 0-.183.015-.366.043-.534A1.748 1.748 0 0 1 4.028 12c0-.968.786-1.754 1.754-1.754.463 0 .898.196 1.207.49 1.207-.883 2.878-1.43 4.744-1.487l.885-4.182a.342.342 0 0 1 .14-.197.35.35 0 0 1 .238-.042l2.906.617a1.214 1.214 0 0 1 1.108-.701zM9.25 12C8.561 12 8 12.562 8 13.25c0 .687.561 1.248 1.25 1.248.687 0 1.248-.561 1.248-1.249 0-.688-.561-1.249-1.249-1.249zm5.5 0c-.687 0-1.248.561-1.248 1.25 0 .687.561 1.248 1.249 1.248.688 0 1.249-.561 1.249-1.249 0-.687-.562-1.249-1.25-1.249zm-5.466 3.99a.327.327 0 0 0-.231.094.33.33 0 0 0 0 .463c.842.842 2.484.913 2.961.913.477 0 2.105-.056 2.961-.913a.361.361 0 0 0 .029-.463.33.33 0 0 0-.464 0c-.547.533-1.684.73-2.512.73-.828 0-1.979-.196-2.512-.73a.326.326 0 0 0-.232-.095z"/>
        </svg>
      </button>
      
      <button 
        class="share-btn copy"
        @click="copyToClipboard"
        :title="copyButtonText"
        aria-label="Copy page link"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
          <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useData } from 'vitepress'
import { useAnalytics } from '../composables/useAnalytics'

const { page } = useData()
const copyButtonText = ref('Copy link')
const { trackSocialShare } = useAnalytics()

const twitterTooltip = computed(() => {
  return 'Share on Twitter/X'
})

function shareToTwitter() {
  const url = encodeURIComponent(window.location.href)
  const text = '🚀 CCProxy + Claude Code = Universal AI development! Connect Claude Code to 100+ open source LLMs with zero config changes!'
  
  trackSocialShare('twitter', 'documentation')
  window.open(`https://twitter.com/intent/tweet?url=${url}&text=${encodeURIComponent(text)}`, '_blank')
}

function shareToLinkedIn() {
  const url = encodeURIComponent(window.location.href)
  
  trackSocialShare('linkedin', 'documentation')
  window.open(`https://www.linkedin.com/sharing/share-offsite/?url=${url}`, '_blank')
}

function shareToReddit() {
  const url = encodeURIComponent(window.location.href)
  const title = encodeURIComponent(page.value.title)
  
  trackSocialShare('reddit', 'documentation')
  window.open(`https://reddit.com/submit?url=${url}&title=${title}`, '_blank')
}

function copyToClipboard() {
  navigator.clipboard.writeText(window.location.href).then(() => {
    copyButtonText.value = '✅ Copied!'
    trackSocialShare('copy_link', 'documentation')
    setTimeout(() => {
      copyButtonText.value = 'Copy link'
    }, 2000)
  })
}
</script>

<style scoped>
.social-share {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  margin: 16px 0 32px auto;
  max-width: fit-content;
}

.share-label {
  font-size: 12px;
  font-weight: 400;
  color: var(--vp-c-text-3);
  white-space: nowrap;
  margin-right: 4px;
}

.share-buttons {
  display: flex;
  gap: 4px;
}

.share-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
  position: relative;
  opacity: 0.7;
}

.share-btn:hover {
  opacity: 1;
  transform: translateY(-1px);
}

.share-btn.twitter {
  background: #1da1f2;
  color: white;
}

.share-btn.twitter:hover {
  background: #1991db;
}

.share-btn.linkedin {
  background: #0077b5;
  color: white;
}

.share-btn.linkedin:hover {
  background: #006ba1;
}

.share-btn.reddit {
  background: #ff4500;
  color: white;
}

.share-btn.reddit:hover {
  background: #e03d00;
}

.share-btn.copy {
  background: var(--vp-c-brand-1);
  color: white;
}

.share-btn.copy:hover {
  background: var(--vp-c-brand-2);
}

/* Tooltip styles */
.share-btn {
  position: relative;
}

.share-btn::after {
  content: attr(title);
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  background: var(--vp-c-bg-elv);
  color: var(--vp-c-text-1);
  padding: 6px 8px;
  border-radius: 4px;
  font-size: 12px;
  white-space: nowrap;
  opacity: 0;
  visibility: hidden;
  transition: all 0.2s ease;
  z-index: 1000;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 4px;
}

.share-btn:hover::after {
  opacity: 1;
  visibility: visible;
}

@media (max-width: 640px) {
  .social-share {
    justify-content: center;
    margin: 12px 0 24px 0;
  }
  
  .share-label {
    font-size: 11px;
  }
  
  .share-btn {
    width: 24px;
    height: 24px;
  }
}
</style>