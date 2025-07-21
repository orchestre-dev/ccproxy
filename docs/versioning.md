# Versioning and Release Management

CCProxy uses semantic versioning with conventional commits for automated release management.

## Overview

The project uses a combination of:
- **Semantic Versioning (SemVer)**: `MAJOR.MINOR.PATCH` format
- **Conventional Commits**: Standardized commit messages
- **Automated Releases**: Triggered on merge to main branch
- **Version Management Script**: Local version management tools

## Semantic Versioning

### Version Format: `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes that are not backwards compatible
- **MINOR**: New features that are backwards compatible
- **PATCH**: Bug fixes that are backwards compatible

### Examples
- `1.0.0` → `1.0.1` (patch: bug fix)
- `1.0.0` → `1.1.0` (minor: new feature)
- `1.0.0` → `2.0.0` (major: breaking change)

## Conventional Commits

### Format
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

| Type | Description | Version Bump |
|------|-------------|--------------|
| `feat` | New feature | MINOR |
| `fix` | Bug fix | PATCH |
| `docs` | Documentation only | None |
| `style` | Code style changes | None |
| `refactor` | Code refactoring | None |
| `perf` | Performance improvements | PATCH |
| `test` | Adding/updating tests | None |
| `build` | Build system changes | None |
| `ci` | CI configuration changes | None |
| `chore` | Other changes | None |
| `revert` | Reverting previous commit | PATCH |

### Breaking Changes

Add `!` after the type or include `BREAKING CHANGE:` in the footer:

```bash
feat!: change API response format
# or
feat: add new authentication method

BREAKING CHANGE: API now requires authentication header
```

### Examples

```bash
# Feature (minor bump)
feat: add user authentication
feat(api): add rate limiting

# Bug fix (patch bump)
fix: prevent memory leak in server
fix(auth): correct token validation

# Breaking change (major bump)
feat!: redesign API endpoints
feat(api)!: change response format

# Documentation (no bump)
docs: update installation guide
docs(api): add authentication examples

# Chore (no bump)
chore: update dependencies
chore(ci): improve test coverage
```

## Automated Release Process

### How It Works

1. **Commit Analysis**: When code is merged to `main`, the system analyzes commits since the last release
2. **Version Calculation**: Determines the appropriate version bump based on conventional commit types
3. **Release Creation**: Automatically creates a new release with:
   - Updated version number
   - Generated changelog
   - Cross-platform binaries
   - Docker images
   - Release notes

### Workflow Triggers

- **Automatic**: Every merge to `main` branch (if version bump needed)
- **Manual**: Workflow dispatch with version type selection

### Release Artifacts

Each release includes:
- **Binaries**: macOS (Intel/Apple Silicon), Linux (x64/ARM64), Windows (x64)
- **Docker Images**: Multi-architecture container images
- **Checksums**: SHA256 verification files
- **Release Notes**: Automatically generated from commits
- **Changelog**: Updated project changelog

## Version Management Script

Use the `scripts/version.sh` script for local version management:

### Available Commands

```bash
# Show current version
./scripts/version.sh current

# Show suggested next version
./scripts/version.sh suggest

# Show what next version would be
./scripts/version.sh next [auto|patch|minor|major]

# Bump version and create commit/tag
./scripts/version.sh bump [auto|patch|minor|major]

# Generate changelog
./scripts/version.sh changelog

# Validate conventional commits
./scripts/version.sh check

# Show help
./scripts/version.sh help
```

### Examples

```bash
# Check current version
$ ./scripts/version.sh current
1.2.3

# See suggested version bump
$ ./scripts/version.sh suggest
Current version: 1.2.3
Suggested bump: minor
Next version: 1.3.0

# Automatically bump version
$ ./scripts/version.sh bump auto
▶️  Current version: 1.2.3
ℹ️  Auto-detected bump type: minor
▶️  Bumping version: 1.2.3 → 1.3.0
ℹ️  Updated internal/version/version.go with version 1.3.0
ℹ️  Generated changelog in CHANGELOG.md
ℹ️  Created commit for version 1.3.0
ℹ️  Created tag v1.3.0
ℹ️  ✅ Version bumped to 1.3.0
```

## Makefile Integration

Version management is integrated into the Makefile:

```bash
# Version information
make version                # Show current git version
make version-current        # Show version from version.go
make version-suggest        # Suggest next version
make version-next           # Show next version (auto)

# Version bumping
make version-bump           # Auto-bump based on commits
make version-bump-patch     # Force patch bump
make version-bump-minor     # Force minor bump
make version-bump-major     # Force major bump

# Validation and changelog
make version-check          # Validate conventional commits
make changelog              # Generate changelog
```

## Git Configuration

### Commit Message Template

Set up the conventional commit template:

```bash
git config commit.template .gitmessage
```

This will show commit type hints when you run `git commit`.

### Pre-commit Hook (Optional)

Add a pre-commit hook to validate commit messages:

```bash
#!/bin/sh
# .git/hooks/commit-msg
./scripts/version.sh check
```

## CI/CD Configuration

### GitHub Actions Workflows

1. **Auto Release** (`.github/workflows/auto-release.yml`)
   - Triggers on main branch push
   - Analyzes commits for version bump
   - Creates release if needed

2. **Manual Release** (`.github/workflows/release.yml`)
   - Triggers on git tag push
   - Manual release creation

### Environment Variables

Required for CI/CD:
- `GITHUB_TOKEN`: Automatically provided by GitHub
- `HOMEBREW_TAP_TOKEN`: For Homebrew formula updates (optional)

## Best Practices

### For Developers

1. **Use Conventional Commits**: Always follow the conventional commit format
2. **Validate Before Push**: Run `make version-check` before pushing
3. **Review Changes**: Use `make version-suggest` to see what release would be created
4. **Test Locally**: Use the version script to test version bumping locally

### For Releases

1. **Let CI Handle It**: The automated system handles most releases
2. **Manual Override**: Use workflow dispatch for special cases
3. **Breaking Changes**: Always use `!` or `BREAKING CHANGE:` for breaking changes
4. **Documentation**: Update documentation with breaking changes

### Commit Message Tips

```bash
# Good
feat: add user profile management
fix: resolve memory leak in auth service
docs: update API documentation
chore: update dependencies

# Bad
Add new feature
Fixed bug
Updated docs
Various changes
```

## Troubleshooting

### Common Issues

**Version not bumping automatically**
- Check if commits follow conventional format: `make version-check`
- Verify commits contain features/fixes since last release

**Manual version bump needed**
- Use: `./scripts/version.sh bump [type]`
- Or trigger manual workflow in GitHub Actions

**Release creation failed**
- Check GitHub Actions logs
- Verify repository permissions
- Ensure no duplicate tags exist

### Recovery

If automated versioning gets out of sync:

```bash
# Reset to current state
git tag -d v1.2.3  # Delete incorrect tag
./scripts/version.sh bump patch  # Create correct version
git push origin main --tags  # Push correction
```

## Migration Guide

### From Manual Versioning

1. **Update commit format**: Start using conventional commits
2. **Set initial version**: Ensure `internal/version/version.go` has correct version
3. **Create baseline tag**: `git tag v1.0.0 && git push origin v1.0.0`
4. **Enable automation**: Merges to main will now auto-release

### Version File Location

The version is stored in `internal/version/version.go`:

```go
package version

var (
    Version   = "1.0.0"
    BuildTime = "2025-07-21_15:30:00"
    Commit    = "abc123"
)
```

This file is automatically updated by the version script and used by the build system.