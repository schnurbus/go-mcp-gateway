# GitHub Actions Workflows

## Docker Image Publishing

The `docker-publish.yml` workflow automatically builds and publishes Docker images to GitHub Container Registry (GHCR) when releases are created.

### Automatic Triggers

**On Release Published:**
- Triggers automatically when a new GitHub release is published
- Extracts version from the release tag (e.g., `v1.0.0`)
- Creates multiple image tags:
  - `ghcr.io/schnurbus/go-mcp-gateway:1.0.0` (version)
  - `ghcr.io/schnurbus/go-mcp-gateway:1.0` (major.minor)
  - `ghcr.io/schnurbus/go-mcp-gateway:1` (major)
  - `ghcr.io/schnurbus/go-mcp-gateway:latest` (if on default branch)

### Manual Triggers

You can manually trigger the workflow from the Actions tab:
1. Go to Actions → "Build and Push Docker Image"
2. Click "Run workflow"
3. Enter a custom tag (e.g., `dev`, `staging`, `latest`)

### Multi-Platform Support

The workflow builds images for:
- `linux/amd64` (Intel/AMD 64-bit)
- `linux/arm64` (ARM 64-bit, including Apple Silicon)

### Caching

The workflow uses GitHub Actions cache to speed up subsequent builds by caching Docker layers.

### Security

- Uses `GITHUB_TOKEN` for authentication (no manual secrets required)
- Generates build provenance attestations for supply chain security
- Runs with minimal permissions (read contents, write packages)

### Usage

After a release is published, pull the image:

```bash
# Pull latest
docker pull ghcr.io/schnurbus/go-mcp-gateway:latest

# Pull specific version
docker pull ghcr.io/schnurbus/go-mcp-gateway:v1.0.0
```

Use in docker-compose:

```yaml
services:
  gateway:
    image: ghcr.io/schnurbus/go-mcp-gateway:latest
```

### Creating a Release

To trigger the workflow, create a new release:

1. **Via GitHub UI:**
   - Go to Releases → "Create a new release"
   - Create a new tag (e.g., `v1.0.0`)
   - Add release notes
   - Publish release

2. **Via GitHub CLI:**
   ```bash
   gh release create v1.0.0 --title "v1.0.0" --notes "Release notes"
   ```

3. **Via Git:**
   ```bash
   git tag -a v1.0.0 -m "Version 1.0.0"
   git push origin v1.0.0
   # Then create the release on GitHub
   ```

### Troubleshooting

**Image not appearing in GHCR:**
- Check that the workflow completed successfully in the Actions tab
- Verify that packages write permission is enabled
- Ensure the repository visibility settings allow package publishing

**Build failures:**
- Check the workflow logs for specific error messages
- Verify that Dockerfile builds successfully locally
- Ensure all required files are not excluded by .dockerignore

**Permission errors:**
- The workflow uses `GITHUB_TOKEN` which should have automatic permissions
- If issues persist, check repository settings → Actions → General → Workflow permissions
