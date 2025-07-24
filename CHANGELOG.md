# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]


### ✨ New Features

- improve parameter handling and code quality
- add route-level parameter configuration

### 🐛 Bug Fixes

- resolve lint error in pipeline.go

### 📚 Documentation

- update CHANGELOG
## [1.6.0] - 2025-07-23

### ✨ Features

- Add Product Hunt badge to docs homepage

## [1.5.0] - 2025-07-23

### ✨ Features

- **Human-readable environment variables**: CCProxy now automatically detects provider-specific environment variables like `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, etc. This eliminates the need for confusing indexed variables (`CCPROXY_PROVIDERS_0_API_KEY`).
- **Simplified configuration**: API keys can now be omitted from config.json when environment variables are set.
- **Backward compatibility**: The indexed format still works for users who prefer it.
- **Multi-format support**: Supports `GEMINI_API_KEY`/`GOOGLE_API_KEY` for Gemini and `XAI_API_KEY`/`GROK_API_KEY` for XAI.
- **AWS Bedrock support**: Automatically combines `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` for Bedrock authentication.
- Add sample configurations and update docs for OpenAI models
- Add environment variable demo and simple config example

### 🐛 Bug Fixes

- Correct apikey field name in example configurations
- Resolve Go version compatibility and formatting issues

### 📚 Documentation

- Major update to all documentation to use human-readable environment variables
- Add comprehensive environment variable reference table
- Add comprehensive Ollama provider documentation
- Add comprehensive Groq provider documentation
- Update provider documentation with environment variables
- Update Docker documentation with human-readable env vars
- Update guide configuration and environment documentation
- Update README to use human-readable environment variable

### 🔧 Other Changes

- Update k8s deployment with human-readable env vars
- Add .serena/ to .gitignore

## [1.4.0] - 2025-07-23

### ✨ Features

- Add comprehensive analytics tracking system
- Integrate analytics tracking into components

### 🐛 Bug Fixes

- Address all PR review comments for analytics implementation

### 📚 Documentation

- Align documentation with actual implementation
- Refocus homepage on open source LLMs
- Simplify README to focus on essentials
- Add environment example and CSP documentation

## [1.3.0] - 2025-07-22

### ✨ Features

- Improve installation UX with automatic config setup
- Add Windows support and platform-specific clarity
- Enhance installer safety and validation

## [1.2.6] - 2025-07-22

### 🐛 Bug Fixes

- Fix installation script cleanup trap causing file not found error

## [1.2.5] - 2025-07-22

### 🐛 Bug Fixes

- Remove vitepress --version command that triggers dev server
- Clean up docs workflow and simplify build process
- Update blog post and clean up GPT-4 references

### 📚 Documentation

- Fix readme formatting

## [1.2.4] - 2025-07-22

### 🐛 Bug Fixes

- Force vitepress build mode with CI=true environment variable

## [1.2.3] - 2025-07-22

### 🐛 Bug Fixes

- Ensure docs workflow runs build instead of dev server

## [1.2.2] - 2025-07-22

### 🐛 Bug Fixes

- Add gitHubToken to Cloudflare Pages deployment

## [1.2.1] - 2025-07-22

### 🐛 Bug Fixes

- Resolve all dead links in documentation

## [1.2.0] - 2025-07-22

### ✨ Features

- Add CCProxy v1+ release announcement
- Feature Kimi K2 and update setup messaging
- Add newsletter component to all blog posts
- Move Kimi K2 guide to 2nd position in blog sidebar

### 🐛 Bug Fixes

- Update v1.0 release post with accurate claims and Kimi K2
- Update Kimi K2 guide with correct configuration
- Correct route configuration documentation and examples
- Resolve YAML syntax error in docs workflow
- Ensure docs workflow runs build instead of dev server

### 🔧 Other Changes

- Refactor blog release post for Claude Code users
- Add detailed routing reference and fix route examples

## [1.1.0] - 2025-07-22

### ✨ Features

- Update navigation and footer with community links
- Add newsletter signup form component
- Add info cards and newsletter form
- Add checksum generation script
- Add Qwen3 235B announcement post

### 🐛 Bug Fixes

- Create working install.sh and correct documentation
- Update Quick Start Configuration section with working example
- Fix install.sh security vulnerabilities
- Address critical vulnerabilities in install.sh

### 📚 Documentation

- Enhance SEO and fix VitePress deployment
- Fix misleading documentation and remove non-functional examples
- Fix VitePress configuration and add SocialShare consistently
- Update provider docs with July 2025 models
- Add comprehensive routing guide
- Clarify model selection and configuration
- Update with July 2025 models and routing info
- Update with implementation learnings

### 🔧 Other Changes

- Update og image
- Remove unsupported provider references
- Style newsletter with accent-colored borders
- Update homepage to reflect 5 providers with 100+ models via OpenRouter
- Add Qwen3 235B, Kimi K2, and Grok models to OpenRouter docs
- Convert unsupported providers to redirect pages
- Update Claude PR Assistant and Code Review workflows

## [1.0.4] - 2025-07-22

### 🐛 Bug Fixes

- Address all low severity GoSec G104 unhandled error issues
- Correct spelling of 'canceled' in comments

## [1.0.3] - 2025-07-22

### 🐛 Bug Fixes

- Address all medium and high severity GoSec issues

## [1.0.2] - 2025-07-22

### 🐛 Bug Fixes

- Fix pre-release version detection to exclude rc tags
- Format version.go with proper tab indentation
- Update version.sh to generate Go files with tab indentation

## [1.0.1] - 2025-07-22

### 🐛 Bug Fixes

- Fix auto-release version bump and remove Docker from all pipelines
- Fix Go version inconsistency and file formatting

## [1.0.0] - 2025-07-17

### ✨ Initial Release

**Core Features:**
- **Multi-provider support**: Anthropic, OpenAI, Groq, DeepSeek, Gemini, OpenRouter
- **API translation**: Converts Anthropic API format to provider-specific formats
- **Intelligent routing**: Automatic model selection based on token count, model type, and parameters
- **Streaming support**: Server-Sent Events (SSE) for real-time responses
- **Claude Code integration**: Auto-start, environment variable management, reference counting
- **Process management**: Background service with PID file locking and graceful shutdown

**Security:**
- API key validation (Bearer token and x-api-key header)
- Localhost-only enforcement when no API key configured
- Request size limits (10MB default) to prevent DoS attacks
- Provider error response sanitization

**Performance:**
- Memory: <20MB baseline usage
- Startup: <100ms cold start
- Connection pooling and reuse
- Efficient memory management

**Developer Experience:**
- Single static binary with no external dependencies
- Cross-platform: Linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64)
- Comprehensive test coverage
- Interactive setup command
- Extensive documentation with VitePress

**Build System:**
- Cross-platform build scripts
- Docker support
- Automated CI/CD with GitHub Actions

[unreleased]: https://github.com/orchestre-dev/ccproxy/compare/v1.6.0...HEAD
[1.6.0]: https://github.com/orchestre-dev/ccproxy/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/orchestre-dev/ccproxy/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/orchestre-dev/ccproxy/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/orchestre-dev/ccproxy/compare/v1.2.6...v1.3.0
[1.2.6]: https://github.com/orchestre-dev/ccproxy/compare/v1.2.5...v1.2.6
[1.2.5]: https://github.com/orchestre-dev/ccproxy/compare/v1.2.4...v1.2.5
[1.2.4]: https://github.com/orchestre-dev/ccproxy/compare/v1.2.3...v1.2.4
[1.2.3]: https://github.com/orchestre-dev/ccproxy/compare/v1.2.2...v1.2.3
[1.2.2]: https://github.com/orchestre-dev/ccproxy/compare/v1.2.1...v1.2.2
[1.2.1]: https://github.com/orchestre-dev/ccproxy/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/orchestre-dev/ccproxy/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/orchestre-dev/ccproxy/compare/v1.0.4...v1.1.0
[1.0.4]: https://github.com/orchestre-dev/ccproxy/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/orchestre-dev/ccproxy/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/orchestre-dev/ccproxy/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/orchestre-dev/ccproxy/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/orchestre-dev/ccproxy/releases/tag/v1.0.0## [1.7.0] - 2025-07-24

### ✨ Features

- add route-level parameter configuration
- improve parameter handling and code quality

### 🐛 Bug Fixes

- resolve lint error in pipeline.go

### 🔧 Other Changes

- docs: update CHANGELOG
- ci: add automatic changelog generation
- chore: update changelog for PR #29

