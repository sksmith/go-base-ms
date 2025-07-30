#!/bin/bash

# Initialize the first release tag
# This script should be run once to set up the initial v0.1.0 tag

set -e

echo "🚀 Initializing first release..."

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Not in a git repository. Please run 'git init' first."
    exit 1
fi

# Check if there are any existing tags
if git describe --tags --abbrev=0 2>/dev/null; then
    echo "⚠️  Git tags already exist. Use 'make release-patch|minor|major' instead."
    exit 1
fi

# Check if working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    echo "❌ Git working directory is not clean. Please commit or stash changes."
    git status --short
    exit 1
fi

# Check if we have any commits
if ! git log --oneline -1 > /dev/null 2>&1; then
    echo "❌ No commits found. Please make at least one commit first."
    exit 1
fi

# Create the initial tag
INITIAL_TAG="v0.1.0"
echo "📝 Creating initial tag: $INITIAL_TAG"

git tag -a "$INITIAL_TAG" -m "Initial release $INITIAL_TAG"

echo "✅ Initial tag $INITIAL_TAG created!"
echo ""
echo "📋 Next steps:"
echo "1. Push the tag: git push origin $INITIAL_TAG"
echo "2. This will trigger the GitHub Actions release workflow"
echo "3. Use 'make release-patch|minor|major' for future releases"
echo ""
echo "🔧 Available make commands:"
echo "  make version          - Show current version"
echo "  make next-versions    - Show next available versions"
echo "  make release-dry-run  - Test release locally"
echo "  make release-patch    - Create patch release (x.x.X)"
echo "  make release-minor    - Create minor release (x.X.0)"
echo "  make release-major    - Create major release (X.0.0)"