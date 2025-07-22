# Documentation Deployment

This document describes how the CCProxy documentation is built and deployed to Cloudflare Pages.

## Overview

- **Production URL**: https://ccproxy.orchestre.dev
- **Preview URLs**: https://pr-{number}.ccproxy.pages.dev
- **Cloudflare Pages Project**: `ccproxy`

## Automatic Deployment

### Production Deployment

When changes are pushed to the `main` branch that affect documentation:
1. GitHub Actions builds the VitePress documentation
2. Deploys to Cloudflare Pages production environment
3. Purges Cloudflare cache for immediate updates

### Preview Deployments

When a pull request is created or updated:
1. GitHub Actions builds the documentation with PR-specific base URL
2. Deploys to a preview environment at `https://pr-{number}.ccproxy.pages.dev`
3. Adds a comment to the PR with the preview URL

## Required Secrets

Configure these secrets in your GitHub repository settings:

- `CLOUDFLARE_API_TOKEN` - API token with Pages deployment permissions
- `CLOUDFLARE_ACCOUNT_ID` - Your Cloudflare account ID
- `CLOUDFLARE_ZONE_ID` - Zone ID for cache purging (production only)
- `GA_MEASUREMENT_ID` - Google Analytics measurement ID (optional)

## Local Development

```bash
cd docs
npm install
npm run dev
```

Visit http://localhost:5173 to preview the documentation locally.

## Manual Deployment

If needed, you can deploy manually using Wrangler:

```bash
cd docs
npm run build
npx wrangler pages deploy .vitepress/dist --project-name=ccproxy
```

## Configuration Files

- `.github/workflows/docs.yml` - GitHub Actions workflow
- `docs/wrangler.jsonc` - Wrangler configuration
- `docs/.vitepress/config.js` - VitePress configuration

## Troubleshooting

### Build Failures

1. Check Node.js version (requires v20+)
2. Verify all dependencies are installed: `npm ci`
3. Check for VitePress build errors in logs

### Deployment Failures

1. Verify Cloudflare API credentials are correct
2. Check if the Pages project exists in Cloudflare dashboard
3. Ensure wrangler version is up to date

### Preview URL Not Working

1. Wait 1-2 minutes for deployment to complete
2. Check GitHub Actions logs for deployment URL
3. Verify PR number in the URL is correct