# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### ✨ New Features

- **Human-readable environment variables**: CCProxy now automatically detects provider-specific environment variables like `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, etc. This eliminates the need for confusing indexed variables (`CCPROXY_PROVIDERS_0_API_KEY`).
- **Simplified configuration**: API keys can now be omitted from config.json when environment variables are set.
- **Backward compatibility**: The indexed format still works for users who prefer it.
- **Multi-format support**: Supports `GEMINI_API_KEY`/`GOOGLE_API_KEY` for Gemini and `XAI_API_KEY`/`GROK_API_KEY` for XAI.
- **AWS Bedrock support**: Automatically combines `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` for Bedrock authentication.

### 📚 Documentation

- Updated all documentation to use human-readable environment variables
- Added comprehensive environment variable reference table
- Created examples demonstrating the new feature

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

## [1.2.0] - 2025-07-22

### ✨ Features

- add CCProxy v1+ release announcement
- feature Kimi K2 and update setup messaging
- add newsletter component to all blog posts
- move Kimi K2 guide to 2nd position in blog sidebar

### 🐛 Bug Fixes

- update v1.0 release post with accurate claims and Kimi K2
- update Kimi K2 guide with correct configuration
- correct route configuration documentation and examples
- resolve YAML syntax error in docs workflow
- ensure docs workflow runs build instead of dev server

### 🔧 Other Changes

- refactor(blog): optimize release post for Claude Code users
- docs: add detailed routing reference and fix route examples

## [1.2.1] - 2025-07-22

### 🐛 Bug Fixes

- resolve all dead links in documentation

## [1.2.2] - 2025-07-22

### 🐛 Bug Fixes

- add gitHubToken to Cloudflare Pages deployment

## [1.2.3] - 2025-07-22

### 🐛 Bug Fixes

- ensure docs workflow runs build instead of dev server

## [1.2.4] - 2025-07-22

### 🐛 Bug Fixes

- force vitepress build mode with CI=true environment variable

## [1.2.5] - 2025-07-22

### 🐛 Bug Fixes

- remove vitepress --version command that triggers dev server
- clean up docs workflow and simplify build process
- update blog post and clean up GPT-4 references

### 🔧 Other Changes

- docs: fix readme

## [1.2.6] - 2025-07-22

### 🐛 Bug Fixes

- installation script cleanup trap causing file not found error

## [1.3.0] - 2025-07-22

### ✨ Features

- improve installation UX with automatic config setup
- add Windows support and platform-specific clarity
- enhance installer safety and validation

## [1.4.0] - 2025-07-23

### ✨ Features

- add comprehensive analytics tracking system
- integrate analytics tracking into components

### 🐛 Bug Fixes

- address all PR review comments for analytics implementation

### 🔧 Other Changes

- docs: align documentation with actual implementation
- docs: refocus homepage on open source LLMs
- docs: simplify README to focus on essentials
- docs: add environment example and CSP documentation

