---
title: Contributing Guide - CCProxy Open Source Contribution
description: Learn how to contribute to CCProxy. Guidelines for submitting issues, pull requests, and participating in the community.
keywords: CCProxy contributing, open source, pull requests, code contribution, community guidelines
---

# Contributing Guide

<SocialShare />

Thank you for your interest in contributing to CCProxy! This guide will help you get started with contributing to the project.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct:

- **Be respectful**: Treat everyone with respect and consideration
- **Be collaborative**: Work together to resolve conflicts  
- **Be inclusive**: Welcome newcomers and help them get started
- **Be professional**: Focus on what is best for the community

## How to Contribute

### Reporting Issues

Found a bug or have a feature request? 

1. **Search existing issues** first to avoid duplicates
2. **Create a new issue** with a clear title and description
3. **Include details**:
   - CCProxy version
   - Go version
   - Operating system
   - Steps to reproduce
   - Expected vs actual behavior
   - Error messages/logs

Example issue:
```markdown
### Description
CCProxy crashes when processing large requests with streaming enabled.

### Environment
- CCProxy version: v1.2.3
- Go version: 1.21
- OS: Ubuntu 22.04

### Steps to Reproduce
1. Start ccproxy with default config
2. Send request with 50k tokens
3. Enable streaming in request

### Expected Behavior
Request should process successfully

### Actual Behavior
Server crashes with panic: runtime error

### Logs
```
panic: runtime error: slice bounds out of range
...
```
```

### Suggesting Features

1. **Open a discussion** first for major features
2. **Explain the use case** and benefits
3. **Consider alternatives** you've explored
4. **Be specific** about the implementation

### Submitting Pull Requests

#### 1. Fork and Clone

```bash
# Fork on GitHub, then:
git clone https://github.com/YOUR_USERNAME/ccproxy.git
cd ccproxy
git remote add upstream https://github.com/orchestre-dev/ccproxy.git
```

#### 2. Create a Branch

```bash
# Update main
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name
# Or for fixes
git checkout -b fix/issue-description
```

#### 3. Make Changes

Follow our coding standards:

```go
// Package comment is required
// Package router handles request routing logic
package router

import (
    "context"
    "fmt"
    
    // Standard library first
    // Then external packages
    // Then internal packages
    "github.com/orchestre-dev/ccproxy/internal/config"
)

// RouterService handles routing decisions
type RouterService struct {
    config *config.Config
}

// NewRouterService creates a new router service
// It accepts a configuration and returns a configured service
func NewRouterService(cfg *config.Config) (*RouterService, error) {
    if cfg == nil {
        return nil, fmt.Errorf("config is required")
    }
    
    return &RouterService{
        config: cfg,
    }, nil
}

// SelectProvider chooses the best provider for a request
func (s *RouterService) SelectProvider(ctx context.Context, req Request) (string, error) {
    // Always check context first
    if err := ctx.Err(); err != nil {
        return "", fmt.Errorf("context cancelled: %w", err)
    }
    
    // Implementation
    return s.selectProvider(req)
}
```

#### 4. Write Tests

All code must have tests:

```go
func TestNewRouterService(t *testing.T) {
    tests := []struct {
        name    string
        config  *config.Config
        wantErr bool
    }{
        {
            name:    "valid config",
            config:  &config.Config{},
            wantErr: false,
        },
        {
            name:    "nil config",
            config:  nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc, err := NewRouterService(tt.config)
            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, svc)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, svc)
            }
        })
    }
}
```

#### 5. Update Documentation

Update relevant documentation:
- API documentation for new endpoints
- Configuration examples for new options
- README updates for major features

#### 6. Commit Your Changes

Follow conventional commits:

```bash
# Format: <type>(<scope>): <subject>

# Features
git commit -m "feat(router): add provider failover support"

# Bug fixes
git commit -m "fix(auth): handle empty API keys gracefully"

# Documentation
git commit -m "docs(api): update endpoint documentation"

# Tests
git commit -m "test(router): add failover test cases"

# Refactoring
git commit -m "refactor(provider): simplify connection logic"
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting, semicolons, etc)
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

#### 7. Push and Create PR

```bash
# Push to your fork
git push origin feature/your-feature-name

# Create pull request on GitHub
```

Pull Request template:
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change)
- [ ] New feature (non-breaking change)
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No new warnings
- [ ] Tests added/updated
- [ ] All tests passing
```

## Development Guidelines

### Code Style

1. **Format**: Always run `gofmt`
2. **Lint**: Pass `golangci-lint`
3. **Comments**: Export functions need comments
4. **Errors**: Wrap errors with context
5. **Logging**: Use structured logging

### Testing Requirements

- Minimum 80% code coverage
- All new features must have tests
- Integration tests for API changes
- Benchmark tests for performance-critical code

### Documentation Standards

- Clear, concise comments
- Examples for complex functions
- Update relevant docs
- Include diagrams where helpful

## Review Process

### What to Expect

1. **Automated checks** run on all PRs
2. **Code review** from maintainers
3. **Feedback** may be provided
4. **Changes** may be requested
5. **Approval** when ready
6. **Merge** by maintainers

### Review Criteria

- **Correctness**: Does it work as intended?
- **Design**: Is it well-architected?
- **Testing**: Are tests comprehensive?
- **Performance**: Any performance impacts?
- **Security**: Any security concerns?
- **Documentation**: Is it well-documented?

### Handling Feedback

- Be open to suggestions
- Ask questions if unclear
- Make requested changes
- Update PR description if needed
- Be patient with the process

## Community

### Getting Help

- **Discussions**: Ask questions in [GitHub Discussions](https://github.com/orchestre-dev/ccproxy/discussions)
- **Issues**: Report bugs via [GitHub Issues](https://github.com/orchestre-dev/ccproxy/issues)
- **Documentation**: Read the [docs](https://ccproxy.orchestre.dev)

### Communication Channels

- GitHub Issues: Bug reports and features
- GitHub Discussions: Questions and ideas
- Pull Requests: Code contributions

### Recognition

Contributors are recognized in:
- Release notes
- Contributors file
- Project documentation

## Release Process

### Versioning

We use semantic versioning (SemVer):
- **Major** (X.0.0): Breaking changes
- **Minor** (0.X.0): New features
- **Patch** (0.0.X): Bug fixes

### Release Cycle

- Monthly minor releases
- Patch releases as needed
- Major releases with notice

## Legal

### License

CCProxy is licensed under the MIT License. By contributing, you agree that your contributions will be licensed under the same license.

### Developer Certificate of Origin

By contributing, you certify that:
1. The contribution is your original work
2. You have the right to submit it
3. You understand it will be public
4. You grant the project license rights

## Quick Start Checklist

New contributor? Follow these steps:

1. [ ] Read this contributing guide
2. [ ] Fork the repository
3. [ ] Clone your fork locally
4. [ ] Set up development environment
5. [ ] Find an issue to work on (look for "good first issue")
6. [ ] Create a feature branch
7. [ ] Make your changes
8. [ ] Write/update tests
9. [ ] Run tests locally
10. [ ] Commit with conventional commits
11. [ ] Push to your fork
12. [ ] Create pull request
13. [ ] Respond to feedback
14. [ ] Celebrate your contribution! 🎉

## Thank You!

Your contributions make CCProxy better for everyone. We appreciate your time and effort in improving the project.

## Next Steps

- [Development Setup](/guide/development) - Set up your environment
- [Testing Guide](/guide/testing) - Learn about testing
- [API Documentation](/api/) - Understand the architecture