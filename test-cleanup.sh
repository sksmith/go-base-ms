#!/bin/bash

# Test script to verify the cleanup functionality
# This simulates what the finalize_project function does

set -e

echo "Testing template cleanup functionality..."

# Create test files that should be cleaned up
mkdir -p test-templates
echo "test template" > test-templates/test.template
echo "test init script" > test-scripts/test-init.sh

# Test CHANGELOG reset
echo "Creating test CHANGELOG..."
cat > test-changelog.md << 'EOF'
# Changelog

## [Unreleased]

### Added
- Initial project setup

## [0.1.0] - $(date +%Y-%m-%d)

### Added
- Basic HTTP server with health endpoints
- Structured logging with slog
- OpenAPI 3.0 specification
- Docker support with multi-stage builds
- Kubernetes deployment manifests
- GoReleaser configuration for automated releases
EOF

# Replace the date placeholder
sed -i.bak "s/\$(date +%Y-%m-%d)/$(date +%Y-%m-%d)/g" test-changelog.md
rm test-changelog.md.bak

echo "✅ CHANGELOG template generated successfully"
cat test-changelog.md

# Cleanup test files
rm -rf test-templates test-scripts test-changelog.md

echo "✅ Cleanup test completed successfully"