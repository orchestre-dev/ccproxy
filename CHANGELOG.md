# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2025-07-22

### 🐛 Bug Fixes

- Fix auto-release version bump and remove Docker from all pipelines
- Fix Go version inconsistency and file formatting

## [1.0.2] - 2025-07-22

### 🐛 Bug Fixes

- Fix pre-release version detection to exclude rc tags
- Format version.go with proper tab indentation
- Update version.sh to generate Go files with tab indentation

## [1.0.3] - 2025-07-22

### 🐛 Bug Fixes

- Address all medium and high severity GoSec issues

## [1.0.4] - 2025-07-22

### 🐛 Bug Fixes

- Address all low severity GoSec G104 unhandled error issues
- Correct spelling of 'canceled' in comments

## [1.1.0] - 2025-07-22

### ✨ Features

- update navigation and footer with community links
- add newsletter signup form component
- add info cards and newsletter form
- add checksum generation script
- add Qwen3 235B announcement post

### 🐛 Bug Fixes

- Create working install.sh and correct documentation
- Update Quick Start Configuration section with working example
-  install.sh security vulnerabilities
- address critical vulnerabilities in install.sh

### 🔧 Other Changes

- docs: Enhance SEO and fix VitePress deployment
- chore: update og image
- docs: Fix misleading documentation and remove non-functional examples
- docs: Fix VitePress configuration and add SocialShare consistently
- docs(providers): update provider docs with July 2025 models
- docs(guide): add comprehensive routing guide
- docs(guides): clarify model selection and configuration
- docs(readme): update with July 2025 models and routing info
- docs(claude): update with implementation learnings
- refactor: remove unsupported provider references
- style(newsletter): add accent-colored borders to input fields
- docs(homepage): update to reflect 5 providers with 100+ models via OpenRouter
- docs(openrouter): add Qwen3 235B, Kimi K2, and Grok models
- docs(providers): convert unsupported providers to redirect pages
- docs(kimi-k2): clarify access via OpenRouter, not direct Groq
- docs(readme): update to show 5 implemented providers
- docs(blog): update index with Qwen3 post and accurate provider count
- Update Claude PR Assistant workflow
- Update Claude Code Review workflow

