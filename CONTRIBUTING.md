# Contributing to CCProxy

Thank you for your interest in contributing to CCProxy! This document provides guidelines and information for contributors.

## Development Setup

1. Clone the repository:
```bash
git clone https://github.com/orchestre-dev/ccproxy.git
cd ccproxy
```

2. Install Go 1.21 or later from [golang.org](https://golang.org)

3. Install dependencies:
```bash
go mod download
```

4. Run tests:
```bash
make test
```

## Code Standards

- Follow standard Go formatting (`go fmt`)
- Write tests for new functionality
- Ensure all tests pass including race detection (`make test-race`)
- Add appropriate documentation comments

## Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages:

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `test:` Test additions or modifications
- `refactor:` Code refactoring
- `chore:` Maintenance tasks
- `ci:` CI/CD changes

Include `BREAKING CHANGE:` in the commit body or `!` after the type for breaking changes.

## Release Process

### Automatic Releases

CCProxy uses an automated release pipeline that creates releases when source code changes are merged to main.

**Releases are triggered when:**
- Source code files are modified:
  - `*.go` files (excluding `*_test.go`)
  - `go.mod` or `go.sum`
  - `Makefile`
  - `scripts/*.sh`
  - `.github/workflows/*.yml`
- AND the commits include version-bump-worthy changes (`feat:`, `fix:`, or `BREAKING CHANGE:`)

**Releases are NOT triggered for:**
- Documentation-only changes (`*.md` files, `docs/` directory)
- Example configurations (`examples/` directory)
- Blog posts or VitePress changes (`.vitepress/` directory)
- Test file changes (`*_test.go`)

This prevents unnecessary version bumps when only documentation or non-functional changes are made.

### Manual Version Bumping

You can manually check what version bump would occur:
```bash
./scripts/version.sh suggest
```

### Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes using conventional commits
4. Push to your fork (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Changelog

The changelog is automatically generated from conventional commits. When you open a PR:
- The CI will validate your commit messages
- A changelog preview will be added as a PR comment (only for source code changes)
- The CHANGELOG.md will be automatically updated when merged

## Testing

Before submitting a PR, ensure:

1. All tests pass:
```bash
make test
```

2. No race conditions:
```bash
make test-race
```

3. Code is properly formatted:
```bash
go fmt ./...
```

4. Build succeeds:
```bash
make build
```

## Documentation

- Update relevant documentation for new features
- Add code comments for complex logic
- Update the README if adding new functionality
- Consider adding examples in the `examples/` directory

## Getting Help

- Open an issue for bugs or feature requests
- Join discussions in the [GitHub Discussions](https://github.com/orchestre-dev/ccproxy/discussions)
- Check existing issues before creating new ones

## License

By contributing to CCProxy, you agree that your contributions will be licensed under the MIT License.