<template>
  <DefaultTheme.Layout>
    <template #layout-bottom>
      <CookieConsent />
    </template>
  </DefaultTheme.Layout>
</template>

<script setup>
import DefaultTheme from 'vitepress/theme'
import CookieConsent from './CookieConsent.vue'
</script>