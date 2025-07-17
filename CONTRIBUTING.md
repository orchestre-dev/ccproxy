# Contributing to CCProxy

Thank you for your interest in contributing to CCProxy! We appreciate your help in making this project better.

## 🎯 Contribution Scope

**We welcome contributions for:**
- 🐛 **Bug fixes** - Help us squash issues and improve stability
- 📖 **Documentation improvements** - Enhance guides, fix typos, add examples
- 🧪 **Testing** - Add test cases, improve test coverage
- 🔧 **Build and tooling** - Improve CI/CD, build scripts, development tools
- 📦 **Provider integrations** - Bug fixes for existing AI provider support
- 🎨 **Website and docs** - UI/UX improvements, design enhancements

**We do NOT accept contributions for:**
- ✋ **Core architecture changes** - Fundamental proxy logic and design
- ✋ **New major features** - Core functionality additions
- ✋ **Breaking API changes** - Changes that affect backward compatibility

## 🚀 Getting Started

### Prerequisites
- Go 1.21 or later
- Git
- Basic familiarity with AI APIs and proxy concepts

### Development Setup
1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/your-username/ccproxy.git
   cd ccproxy
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run tests:
   ```bash
   go test ./...
   ```

## 🐛 Bug Fixes

### Before You Start
1. **Check existing issues** - Search [GitHub Issues](https://github.com/praneybehl/ccproxy/issues) to avoid duplicates
2. **Create an issue** - Describe the bug using our [bug report template](https://github.com/praneybehl/ccproxy/issues/new?template=bug_report.md)
3. **Get approval** - Wait for maintainer acknowledgment before starting work

### Bug Fix Process
1. **Create a branch** from `main`:
   ```bash
   git checkout -b fix/descriptive-bug-name
   ```
2. **Write tests** - Add test cases that reproduce the bug
3. **Fix the issue** - Make minimal changes to resolve the problem
4. **Test thoroughly** - Ensure all tests pass
5. **Submit a PR** - Use our pull request template

## 📖 Documentation Contributions

### Documentation Types
- **User guides** - Installation, configuration, usage examples
- **API documentation** - Endpoint descriptions, examples
- **Provider guides** - AI provider setup and configuration
- **Troubleshooting** - Common issues and solutions

### Documentation Guidelines
- **Clear and concise** - Reduce cognitive load for users
- **Tested examples** - Verify all code examples work
- **Consistent formatting** - Follow existing style patterns
- **Mobile-friendly** - Ensure responsive design

## 🧪 Testing Guidelines

### Test Requirements
- **Unit tests** for all bug fixes
- **Integration tests** for provider interactions
- **Documentation tests** for code examples
- **No test coverage reduction** - Maintain or improve coverage

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...
```

## 📝 Pull Request Process

### Before Submitting
- [ ] Tests pass locally
- [ ] Code follows project style (run `golangci-lint run`)
- [ ] Documentation updated if needed
- [ ] No breaking changes
- [ ] PR template completed

### PR Requirements
1. **Descriptive title** - Clear summary of changes
2. **Detailed description** - What, why, and how
3. **Linked issue** - Reference the related issue
4. **Test evidence** - Show tests pass
5. **Documentation** - Update relevant docs

### Review Process
1. **Automated checks** - CI/CD must pass
2. **Maintainer review** - Code quality and scope check
3. **Testing verification** - Functionality validation
4. **Approval and merge** - By project maintainers

## 🎨 Code Style

### Go Style
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` and `golangci-lint`
- Write clear, self-documenting code
- Add comments for complex logic

### Commit Messages
```
type: brief description

Longer explanation if needed

Fixes #123
```

**Types:** `fix`, `docs`, `test`, `ci`, `style`, `refactor`

## 🔍 Provider Bug Fixes

### Supported Providers
- Groq (Kimi K2)
- OpenRouter
- OpenAI
- Google Gemini
- Mistral AI
- XAI (Grok)
- Ollama

### Provider Fix Guidelines
- **Test with real APIs** - Verify fixes work with actual providers
- **Handle edge cases** - Error conditions, rate limits, timeouts
- **Update documentation** - Reflect any configuration changes
- **Backward compatibility** - Don't break existing configs

## 🆘 Getting Help

### Communication Channels
- **[GitHub Discussions](https://github.com/praneybehl/ccproxy/discussions)** - Questions and community support
- **[GitHub Issues](https://github.com/praneybehl/ccproxy/issues)** - Bug reports and feature requests
- **Pull Request comments** - Implementation-specific discussions

### Response Times
- **Issues**: Within 2-3 business days
- **Pull Requests**: Within 1 week
- **Discussions**: Community-driven, varies

## 📋 Issue Labels

- `bug` - Something isn't working
- `documentation` - Improvements or additions to docs
- `good first issue` - Good for newcomers
- `help wanted` - Extra attention is needed
- `provider-*` - Provider-specific issues
- `wontfix` - This will not be worked on

## 🙏 Recognition

Contributors are recognized in:
- **Release notes** - For significant contributions
- **Contributors list** - In repository documentation
- **Community mentions** - In discussions and social media

## 📜 License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for helping make CCProxy better!** 🚀

For questions about contributing, please [start a discussion](https://github.com/praneybehl/ccproxy/discussions) or [open an issue](https://github.com/praneybehl/ccproxy/issues).