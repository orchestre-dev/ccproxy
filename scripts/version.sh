#!/usr/bin/env bash
# CCProxy Version Management Script
# Handles semantic versioning with conventional commits support

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
VERSION_FILE="internal/version/version.go"
CHANGELOG_FILE="CHANGELOG.md"

# Print colored message
print_msg() {
    local color=$1
    shift
    echo -e "${color}$*${NC}"
}

# Print info message
info() {
    print_msg "$GREEN" "ℹ️  $*"
}

# Print error message
error() {
    print_msg "$RED" "❌ $*"
}

# Print warning message
warn() {
    print_msg "$YELLOW" "⚠️  $*"
}

# Print step message
step() {
    print_msg "$BLUE" "▶️  $*"
}

# Show help
show_help() {
    cat << EOF
CCProxy Version Management Script

Usage: $0 [COMMAND] [OPTIONS]

Commands:
  current                    Show current version
  next [TYPE]               Calculate next version based on commit history
  bump [TYPE]               Bump version and create commit/tag
  changelog                 Generate changelog from git history
  check                     Validate conventional commits since last tag
  suggest                   Suggest next version based on commits

Version Types:
  patch                     Bug fixes (default)
  minor                     New features (backwards compatible)
  major                     Breaking changes
  auto                      Automatically determine from commits

Options:
  -h, --help               Show this help message
  -v, --verbose            Verbose output
  --dry-run               Show what would be done without making changes
  --no-commit             Don't create git commit
  --no-tag                Don't create git tag

Examples:
  $0 current               # Show current version
  $0 bump auto             # Auto-bump based on commits
  $0 bump minor            # Force minor version bump
  $0 next auto             # Show what next version would be
  $0 changelog             # Generate changelog

Conventional Commits:
  feat:     New feature (minor bump)
  fix:      Bug fix (patch bump)
  docs:     Documentation only
  style:    Formatting, no code change
  refactor: Code change that neither fixes bug nor adds feature
  perf:     Performance improvement
  test:     Adding tests
  chore:    Updating build tasks, etc.

  BREAKING CHANGE: or ! after type (major bump)

EOF
}

# Get current version from version.go
get_current_version() {
    if [ ! -f "$VERSION_FILE" ]; then
        echo "0.0.0"
        return
    fi
    
    local version
    version=$(grep -E 'Version\s*=\s*"' "$VERSION_FILE" | sed -E 's/.*"([^"]+)".*/\1/' 2>/dev/null)
    if [ -n "$version" ] && [ "$version" != "" ]; then
        echo "$version"
    else
        echo "0.0.0"
    fi
}

# Get latest git tag
get_latest_tag() {
    git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"
}

# Parse semantic version
parse_version() {
    local version=$1
    version=${version#v}  # Remove 'v' prefix if present
    
    if [[ ! "$version" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)(-.*)?$ ]]; then
        error "Invalid version format: $version"
        exit 1
    fi
    
    echo "${BASH_REMATCH[1]} ${BASH_REMATCH[2]} ${BASH_REMATCH[3]}"
}

# Increment version
increment_version() {
    local current=$1
    local type=$2
    
    read -r major minor patch <<< "$(parse_version "$current")"
    
    case "$type" in
        major)
            ((major++))
            minor=0
            patch=0
            ;;
        minor)
            ((minor++))
            patch=0
            ;;
        patch)
            ((patch++))
            ;;
        *)
            error "Invalid version type: $type"
            exit 1
            ;;
    esac
    
    echo "${major}.${minor}.${patch}"
}

# Analyze commits since last tag to determine version bump
analyze_commits() {
    local last_tag
    last_tag=$(get_latest_tag)
    
    local commits
    if [ "$last_tag" = "v0.0.0" ]; then
        commits=$(git log --pretty=format:"%s" --no-merges)
    else
        commits=$(git log "${last_tag}..HEAD" --pretty=format:"%s" --no-merges)
    fi
    
    local has_breaking=false
    local has_feature=false
    local has_fix=false
    
    while IFS= read -r commit; do
        if [[ "$commit" =~ BREAKING[[:space:]]*CHANGE|!: ]] || [[ "$commit" =~ ^[^:]+!: ]]; then
            has_breaking=true
        elif [[ "$commit" =~ ^feat(\(.+\))?: ]]; then
            has_feature=true
        elif [[ "$commit" =~ ^fix(\(.+\))?: ]]; then
            has_fix=true
        fi
    done <<< "$commits"
    
    if [ "$has_breaking" = true ]; then
        echo "major"
    elif [ "$has_feature" = true ]; then
        echo "minor"
    elif [ "$has_fix" = true ]; then
        echo "patch"
    else
        echo "none"
    fi
}

# Validate conventional commits
validate_commits() {
    local last_tag
    last_tag=$(get_latest_tag)
    
    local commits
    if [ "$last_tag" = "v0.0.0" ]; then
        commits=$(git log --pretty=format:"%s" --no-merges)
    else
        commits=$(git log "${last_tag}..HEAD" --pretty=format:"%s" --no-merges)
    fi
    
    local invalid_commits=()
    local valid_types="feat|fix|docs|style|refactor|perf|test|chore|build|ci|revert"
    
    while IFS= read -r commit; do
        if [ -n "$commit" ] && [[ ! "$commit" =~ ^(${valid_types})(\(.+\))?!?:\ .+ ]]; then
            invalid_commits+=("$commit")
        fi
    done <<< "$commits"
    
    if [ ${#invalid_commits[@]} -gt 0 ]; then
        warn "Found non-conventional commits:"
        for commit in "${invalid_commits[@]}"; do
            echo "  - $commit"
        done
        return 1
    fi
    
    info "All commits follow conventional commit format ✅"
    return 0
}

# Create or update version.go file
update_version_file() {
    local version=$1
    local build_time
    build_time=$(date -u '+%Y-%m-%d_%H:%M:%S')
    local commit
    commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    
    mkdir -p "$(dirname "$VERSION_FILE")"
    
    cat > "$VERSION_FILE" << EOF
package version

// Version information
var (
    // Version is the current version of CCProxy
    Version = "${version}"
    
    // BuildTime is when this binary was built
    BuildTime = "${build_time}"
    
    // Commit is the git commit hash this was built from
    Commit = "${commit}"
)
EOF
    
    info "Updated $VERSION_FILE with version $version"
}

# Generate changelog
generate_changelog() {
    local last_tag
    last_tag=$(get_latest_tag)
    local current_version
    current_version=$(get_current_version)
    
    step "Generating changelog..."
    
    # Create changelog header if file doesn't exist
    if [ ! -f "$CHANGELOG_FILE" ]; then
        cat > "$CHANGELOG_FILE" << EOF
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

EOF
    fi
    
    # Generate changelog entries
    local temp_changelog
    temp_changelog=$(mktemp)
    
    # Header
    cat "$CHANGELOG_FILE" > "$temp_changelog"
    
    # New version section
    echo "## [${current_version}] - $(date '+%Y-%m-%d')" >> "$temp_changelog"
    echo "" >> "$temp_changelog"
    
    # Get commits since last tag
    local commits
    if [ "$last_tag" = "v0.0.0" ]; then
        commits=$(git log --pretty=format:"%s" --no-merges --reverse)
    else
        commits=$(git log "${last_tag}..HEAD" --pretty=format:"%s" --no-merges --reverse)
    fi
    
    # Categorize commits
    local features=()
    local fixes=()
    local breaking=()
    local other=()
    
    while IFS= read -r commit; do
        if [[ "$commit" =~ BREAKING[[:space:]]*CHANGE ]] || [[ "$commit" =~ ^[^:]+!: ]]; then
            breaking+=("$commit")
        elif [[ "$commit" =~ ^feat(\(.+\))?: ]]; then
            features+=("${commit#feat*: }")
        elif [[ "$commit" =~ ^fix(\(.+\))?: ]]; then
            fixes+=("${commit#fix*: }")
        else
            other+=("$commit")
        fi
    done <<< "$commits"
    
    # Add sections
    if [ ${#breaking[@]} -gt 0 ]; then
        echo "### 💥 BREAKING CHANGES" >> "$temp_changelog"
        echo "" >> "$temp_changelog"
        for item in "${breaking[@]}"; do
            echo "- $item" >> "$temp_changelog"
        done
        echo "" >> "$temp_changelog"
    fi
    
    if [ ${#features[@]} -gt 0 ]; then
        echo "### ✨ Features" >> "$temp_changelog"
        echo "" >> "$temp_changelog"
        for item in "${features[@]}"; do
            echo "- $item" >> "$temp_changelog"
        done
        echo "" >> "$temp_changelog"
    fi
    
    if [ ${#fixes[@]} -gt 0 ]; then
        echo "### 🐛 Bug Fixes" >> "$temp_changelog"
        echo "" >> "$temp_changelog"
        for item in "${fixes[@]}"; do
            echo "- $item" >> "$temp_changelog"
        done
        echo "" >> "$temp_changelog"
    fi
    
    if [ ${#other[@]} -gt 0 ]; then
        echo "### 🔧 Other Changes" >> "$temp_changelog"
        echo "" >> "$temp_changelog"
        for item in "${other[@]}"; do
            echo "- $item" >> "$temp_changelog"
        done
        echo "" >> "$temp_changelog"
    fi
    
    # Replace original file
    mv "$temp_changelog" "$CHANGELOG_FILE"
    
    info "Generated changelog in $CHANGELOG_FILE"
}

# Main command handling
main() {
    local command="${1:-help}"
    local dry_run=false
    local verbose=false
    local no_commit=false
    local no_tag=false
    
    # Parse flags
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dry-run)
                dry_run=true
                shift
                ;;
            --verbose|-v)
                verbose=true
                shift
                ;;
            --no-commit)
                no_commit=true
                shift
                ;;
            --no-tag)
                no_tag=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                break
                ;;
        esac
    done
    
    case "$command" in
        current)
            current_version=$(get_current_version)
            echo "$current_version"
            ;;
            
        next)
            local bump_type="${2:-auto}"
            local current_version
            current_version=$(get_current_version)
            
            if [ "$bump_type" = "auto" ]; then
                bump_type=$(analyze_commits)
                if [ "$bump_type" = "none" ]; then
                    echo "$current_version"
                    info "No version bump needed"
                    exit 0
                fi
            fi
            
            local next_version
            next_version=$(increment_version "$current_version" "$bump_type")
            echo "$next_version"
            ;;
            
        bump)
            local bump_type="${2:-auto}"
            local current_version
            current_version=$(get_current_version)
            
            step "Current version: $current_version"
            
            if [ "$bump_type" = "auto" ]; then
                bump_type=$(analyze_commits)
                if [ "$bump_type" = "none" ]; then
                    info "No version bump needed"
                    exit 0
                fi
                info "Auto-detected bump type: $bump_type"
            fi
            
            local new_version
            new_version=$(increment_version "$current_version" "$bump_type")
            
            step "Bumping version: $current_version → $new_version"
            
            if [ "$dry_run" = true ]; then
                info "DRY RUN - Would bump version to $new_version"
                exit 0
            fi
            
            # Update version file
            update_version_file "$new_version"
            
            # Generate changelog
            generate_changelog
            
            # Git operations
            if [ "$no_commit" = false ]; then
                git add "$VERSION_FILE" "$CHANGELOG_FILE"
                git commit -m "chore(release): bump version to $new_version"
                info "Created commit for version $new_version"
            fi
            
            if [ "$no_tag" = false ]; then
                git tag -a "v$new_version" -m "Release v$new_version"
                info "Created tag v$new_version"
            fi
            
            info "✅ Version bumped to $new_version"
            ;;
            
        changelog)
            generate_changelog
            ;;
            
        check)
            step "Validating conventional commits..."
            if validate_commits; then
                info "✅ All commits are valid"
            else
                error "❌ Some commits don't follow conventional format"
                exit 1
            fi
            ;;
            
        suggest)
            local current_version
            current_version=$(get_current_version)
            local suggested_type
            suggested_type=$(analyze_commits)
            
            echo "Current version: $current_version"
            echo "Suggested bump: $suggested_type"
            
            if [ "$suggested_type" != "none" ]; then
                local next_version
                next_version=$(increment_version "$current_version" "$suggested_type")
                echo "Next version: $next_version"
            else
                echo "No version bump needed"
            fi
            ;;
            
        help|--help|-h)
            show_help
            ;;
            
        *)
            error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Check dependencies
check_deps() {
    local missing=()
    
    command -v git >/dev/null 2>&1 || missing+=("git")
    
    if [ ${#missing[@]} -ne 0 ]; then
        error "Missing required commands: ${missing[*]}"
        exit 1
    fi
}

# Run main function
check_deps
main "$@"