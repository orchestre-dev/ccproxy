# Cloudflare Pages Deployment Guide

## Setup Instructions

### 1. Cloudflare Pages Project Setup

1. **Go to Cloudflare Dashboard** → Pages → Create a project
2. **Connect to Git** → Select the `orchestre-dev/ccproxy` repository
3. **Configure build settings**:
   - **Framework preset**: VitePress
   - **Build command**: `npm run docs:build`
   - **Build output directory**: `docs/.vitepress/dist`
   - **Root directory**: `/` (leave empty for root)
   - **Node.js version**: `18.x`

### 2. Environment Variables

No environment variables are required for the documentation build.

### 3. Custom Domain Setup

1. **Go to your Cloudflare Pages project** → Custom domains
2. **Add custom domain**: `ccproxy.orchestre.dev`
3. **Verify DNS configuration** in your Cloudflare DNS settings:
   ```
   Type: CNAME
   Name: ccproxy
   Target: orchestre-dev-ccproxy.pages.dev (or your assigned pages.dev domain)
   Proxy status: Proxied (orange cloud)
   ```

### 4. Build Settings

The build will automatically:
- Install dependencies with `npm install`
- Build the VitePress site with `npm run docs:build`
- Deploy to `https://ccproxy.orchestre.dev`

### 5. Automatic Deployments

- **Production deployments**: Triggered on push to `main` branch
- **Preview deployments**: Triggered on pull requests
- **Build time**: ~2-3 minutes

### 6. SSL/TLS

Cloudflare automatically provides:
- Free SSL certificate for `ccproxy.orchestre.dev`
- HTTP to HTTPS redirects
- Modern TLS protocols

### 7. Performance Features

Cloudflare Pages includes:
- Global CDN distribution
- Automatic static asset optimization
- Brotli/Gzip compression
- Edge caching

## Testing Deployment

After setup, verify:
1. ✅ Site loads at `https://ccproxy.orchestre.dev`
2. ✅ All internal links work correctly
3. ✅ Search functionality works
4. ✅ Social sharing works
5. ✅ GitHub links point to `orchestre-dev/ccproxy`
6. ✅ SEO meta tags include correct domain

## Maintenance

- **Monitor builds**: Check Cloudflare Pages dashboard for build status
- **Branch previews**: Use for testing before merging to main
- **Analytics**: Enable in Cloudflare dashboard for traffic insights