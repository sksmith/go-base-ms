# Release Management

This project uses [GoReleaser](https://goreleaser.com/) for automated release management with strict version control to prevent overwrites.

## 🚀 Quick Start

### First Time Setup
```bash
# Initialize the first release (creates v0.1.0)
make release-init
git push origin v0.1.0
```

### Regular Releases
```bash
# Check current version and next options
make version
make next-versions

# Create releases (automatically increments)
make release-patch    # v1.0.0 → v1.0.1 (bug fixes)
make release-minor    # v1.0.1 → v1.1.0 (new features)  
make release-major    # v1.1.0 → v2.0.0 (breaking changes)
```

## 📋 Release Process

1. **Automatic Version Management**: GoReleaser automatically increments version numbers based on the latest git tag
2. **Semantic Versioning**: Uses semver (v1.2.3) format
3. **Git Tags**: Creates and pushes git tags automatically
4. **GitHub Releases**: Creates GitHub releases with changelogs
5. **Multi-Arch Builds**: Builds for Linux, macOS, Windows (amd64/arm64)
6. **Docker Images**: Publishes to GitHub Container Registry
7. **Checksums**: Generates checksums for all artifacts

## 🛡️ Version Protection

The system prevents version overwrites by:
- **Git Tag Validation**: Checks existing tags before creating new ones
- **Automatic Increment**: Always creates the next logical version
- **Clean Repo Check**: Ensures working directory is clean before releasing
- **Push Protection**: GitHub prevents force-pushing tags

## 🔧 Available Commands

| Command | Description | Example Output |
|---------|-------------|----------------|
| `make version` | Show current version | `Current version: v1.0.0` |
| `make next-versions` | Show next available versions | `Next patch: v1.0.1`<br>`Next minor: v1.1.0`<br>`Next major: v2.0.0` |
| `make release-init` | Initialize first release | Creates `v0.1.0` |
| `make release-dry-run` | Test release locally | Builds without publishing |
| `make release-snapshot` | Create snapshot build | For testing (no tags) |
| `make release-patch` | Create patch release | `v1.0.0` → `v1.0.1` |
| `make release-minor` | Create minor release | `v1.0.1` → `v1.1.0` |
| `make release-major` | Create major release | `v1.1.0` → `v2.0.0` |
| `make release-clean` | Clean artifacts | Removes `dist/` folder |

## 📦 Release Artifacts

Each release creates:
- **Binaries**: Cross-compiled for multiple platforms
- **Docker Images**: Multi-arch containers in GHCR
- **Checksums**: SHA256 checksums for verification
- **Release Notes**: Auto-generated changelogs
- **Source Code**: Git archives

## 🐳 Docker Images

Images are published to GitHub Container Registry:
```bash
# Pull specific version
docker pull ghcr.io/USERNAME/go-base-ms:v1.0.0

# Pull latest
docker pull ghcr.io/USERNAME/go-base-ms:latest
```

## 🔍 Version Information

The application exposes version information via:
- **Logs**: Startup logs show version details
- **API Endpoint**: `GET /version` returns JSON with version info
- **Build Metadata**: Embedded in binary via ldflags

Example version response:
```json
{
  "version": "v1.0.0",
  "commit": "abc123def456",
  "date": "2024-01-01T00:00:00Z",
  "built_by": "goreleaser"
}
```

## 🔐 GitHub Actions

The release workflow is triggered automatically when you push a tag:
1. Checks out code
2. Sets up Go environment
3. Logs into GitHub Container Registry
4. Runs GoReleaser to build and publish
5. Updates GitHub release with artifacts

## 🐛 Troubleshooting

**Problem**: `make release-patch` fails with "tag already exists"
**Solution**: Use `make version` to check current version, then use appropriate increment command

**Problem**: Docker build fails during release
**Solution**: Check Dockerfile.goreleaser and ensure all required files are included

**Problem**: GitHub Actions fails to publish
**Solution**: Ensure repository has proper permissions for GITHUB_TOKEN and packages

## 📝 Best Practices

1. **Always test locally first**: Use `make release-dry-run`
2. **Use semantic versioning**: Patch for fixes, minor for features, major for breaking changes
3. **Clean working directory**: Commit all changes before releasing
4. **Review changelogs**: Check generated release notes before publishing
5. **Test Docker images**: Verify containers work after release