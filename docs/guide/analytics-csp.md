# Analytics and Content Security Policy (CSP)

This guide explains how to configure Content Security Policy headers when using Google Analytics with CCProxy documentation.

## CSP Requirements for Google Analytics

The Google Analytics implementation in CCProxy requires the following CSP directives:

### Required CSP Headers

```http
Content-Security-Policy: 
  script-src 'self' https://www.googletagmanager.com 'unsafe-inline';
  connect-src 'self' https://www.google-analytics.com https://analytics.google.com https://region1.google-analytics.com;
  img-src 'self' https://www.google-analytics.com;
```

### Explanation of Directives

1. **script-src**:
   - `'self'` - Allow scripts from the same origin
   - `https://www.googletagmanager.com` - Google Tag Manager scripts
   - `'unsafe-inline'` - Required for the GA initialization script

2. **connect-src**:
   - `'self'` - Allow XHR/fetch to same origin
   - `https://www.google-analytics.com` - GA data collection endpoint
   - `https://analytics.google.com` - GA configuration endpoint
   - `https://region1.google-analytics.com` - Regional GA endpoints

3. **img-src**:
   - `'self'` - Allow images from same origin
   - `https://www.google-analytics.com` - GA tracking pixels

## Alternative: External Script Approach

To avoid using `'unsafe-inline'`, you can move the GA initialization to an external file:

1. Create `/public/js/analytics.js`:
```javascript
window.dataLayer = window.dataLayer || [];
function gtag(){dataLayer.push(arguments);}
gtag('js', new Date());

// Default consent mode
gtag('consent', 'default', {
  'analytics_storage': 'denied',
  'ad_storage': 'denied',
  'ad_user_data': 'denied',
  'ad_personalization': 'denied',
  'wait_for_update': 500
});

// Configuration
gtag('config', 'G-R0JGBZ98R7', {
  'anonymize_ip': true,
  'cookie_flags': 'SameSite=None;Secure'
});
```

2. Update VitePress config to load the external script:
```javascript
head: [
  ['script', { async: true, src: 'https://www.googletagmanager.com/gtag/js?id=G-R0JGBZ98R7' }],
  ['script', { src: '/js/analytics.js' }]
]
```

3. Use stricter CSP without `'unsafe-inline'`:
```http
Content-Security-Policy: 
  script-src 'self' https://www.googletagmanager.com;
  connect-src 'self' https://www.google-analytics.com https://analytics.google.com;
  img-src 'self' https://www.google-analytics.com;
```

## Nginx Configuration Example

```nginx
location / {
  add_header Content-Security-Policy "script-src 'self' https://www.googletagmanager.com 'unsafe-inline'; connect-src 'self' https://www.google-analytics.com https://analytics.google.com https://region1.google-analytics.com; img-src 'self' https://www.google-analytics.com;" always;
  try_files $uri $uri/ /index.html;
}
```

## Cloudflare Pages Configuration

Add these headers in your `_headers` file:

```
/*
  Content-Security-Policy: script-src 'self' https://www.googletagmanager.com 'unsafe-inline'; connect-src 'self' https://www.google-analytics.com https://analytics.google.com https://region1.google-analytics.com; img-src 'self' https://www.google-analytics.com;
```

## Testing Your CSP

1. Use browser developer tools to check for CSP violations
2. Monitor the console for blocked resources
3. Use [CSP Evaluator](https://csp-evaluator.withgoogle.com/) to validate your policy

## Privacy Considerations

The current implementation includes:
- IP anonymization enabled
- Cookie consent required before tracking
- No advertising or personalization tracking
- Secure cookie settings with SameSite=None

These privacy features work alongside CSP to ensure both security and user privacy.