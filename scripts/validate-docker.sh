#!/bin/bash
# Script to validate Docker configuration files

set -e

echo "Validating Docker configuration files..."

# Check required files exist
required_files=(
    "Dockerfile"
    "Dockerfile.dev"
    "Dockerfile.multiarch"
    ".dockerignore"
    "docker-compose.yml"
    "docker-compose.dev.yml"
    "docker/entrypoint.sh"
)

for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        echo "✓ $file exists"
    else
        echo "✗ $file is missing"
        exit 1
    fi
done

# Validate Dockerfile syntax (basic check)
echo -e "\nValidating Dockerfile syntax..."
for dockerfile in Dockerfile*; do
    echo "Checking $dockerfile..."
    # Check for FROM instruction
    if ! grep -q "^FROM" "$dockerfile"; then
        echo "✗ $dockerfile missing FROM instruction"
        exit 1
    fi
    # Check for proper stage naming
    if grep -q "AS builder" "$dockerfile" && ! grep -q "FROM.*builder" "$dockerfile"; then
        echo "✗ $dockerfile has unused builder stage"
        exit 1
    fi
    echo "✓ $dockerfile basic syntax OK"
done

# Validate docker-compose syntax
echo -e "\nValidating docker-compose files..."
for compose_file in docker-compose*.yml; do
    echo "Checking $compose_file..."
    # Check for version
    if ! grep -q "^version:" "$compose_file"; then
        echo "✗ $compose_file missing version"
        exit 1
    fi
    # Check for services
    if ! grep -q "^services:" "$compose_file"; then
        echo "✗ $compose_file missing services section"
        exit 1
    fi
    echo "✓ $compose_file basic syntax OK"
done

# Check .dockerignore
echo -e "\nValidating .dockerignore..."
important_ignores=("*.test" "coverage.out" ".git")
for ignore in "${important_ignores[@]}"; do
    if grep -q "$ignore" .dockerignore; then
        echo "✓ .dockerignore includes $ignore"
    else
        echo "⚠ .dockerignore should include $ignore"
    fi
done

# Check entrypoint script
echo -e "\nValidating entrypoint script..."
if [ -x "docker/entrypoint.sh" ]; then
    echo "✓ docker/entrypoint.sh is executable"
else
    echo "✗ docker/entrypoint.sh is not executable"
    exit 1
fi

# Validate build context size
echo -e "\nEstimating build context size..."
context_size=$(du -sh . --exclude=.git --exclude=dist --exclude=tmp 2>/dev/null | cut -f1)
echo "Build context size: $context_size"

echo -e "\n✅ Docker configuration validation passed!"