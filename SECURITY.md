# Security Guide

This document outlines security best practices for the go-mcp-gateway project, especially important when publishing as a public repository.

## Table of Contents
- [Before Making Repository Public](#before-making-repository-public)
- [GitHub Repository Settings](#github-repository-settings)
- [Secrets Management](#secrets-management)
- [Branch Protection](#branch-protection)
- [GitHub Actions Security](#github-actions-security)
- [Dependency Security](#dependency-security)
- [Security Scanning](#security-scanning)
- [Incident Response](#incident-response)

## Before Making Repository Public

### Critical Pre-Flight Checklist

- [ ] **Verify no secrets in git history**
  ```bash
  # Check for common secret patterns
  git log -p | grep -i "client_secret\|password\|api_key\|token"

  # Use git-secrets or similar tools
  git secrets --scan-history
  ```

- [ ] **Ensure .env is not committed**
  ```bash
  git status
  # .env should NOT appear in the list
  ```

- [ ] **Review all files for sensitive data**
  ```bash
  # Search for potential secrets
  grep -r "client_secret" . --exclude-dir=.git
  grep -r "password" . --exclude-dir=.git
  ```

- [ ] **Rotate ALL credentials if any were committed**
  - Generate new Google OAuth credentials
  - Update all production deployments
  - Revoke old credentials in Google Cloud Console

### If Secrets Were Already Committed

If you've already committed secrets to git history, you MUST:

1. **Rotate all compromised credentials immediately**
   - Google OAuth: Create new credentials in Google Cloud Console
   - Delete old credentials
   - Update all deployments

2. **Clean git history** (use with caution):
   ```bash
   # Option 1: Using BFG Repo-Cleaner
   bfg --replace-text passwords.txt
   git reflog expire --expire=now --all
   git gc --prune=now --aggressive

   # Option 2: Using git-filter-repo
   git filter-repo --path .env --invert-paths
   ```

3. **Force push to remote** (if already pushed):
   ```bash
   git push origin --force --all
   git push origin --force --tags
   ```

4. **Notify collaborators** to re-clone the repository

## GitHub Repository Settings

### Required Settings

1. **Enable Vulnerability Alerts**
   - Go to: Settings → Security → Code security and analysis
   - Enable: "Dependency graph"
   - Enable: "Dependabot alerts"
   - Enable: "Dependabot security updates"

2. **Enable Secret Scanning**
   - Settings → Security → Code security and analysis
   - Enable: "Secret scanning"
   - Enable: "Push protection" (prevents pushing secrets)

3. **Configure Branch Protection**
   - Settings → Branches → Add rule for `main`
   - See [Branch Protection](#branch-protection) section

4. **Restrict GitHub Actions Permissions**
   - Settings → Actions → General → Workflow permissions
   - Select: "Read repository contents and packages permissions"
   - Check: "Allow GitHub Actions to create and approve pull requests"

5. **Enable Private Vulnerability Reporting**
   - Settings → Security → Code security and analysis
   - Enable: "Private vulnerability reporting"

## Secrets Management

### Never Commit These Files

```
.env
.env.*
*.pem
*.key
*.crt
*.p12
*.pfx
config/secrets.yaml
credentials.json
```

### Using GitHub Secrets

For CI/CD, use GitHub Secrets instead of committing credentials:

**Setting Secrets:**
1. Repository → Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add secrets:
   - `OAUTH_GOOGLE_CLIENT_ID`
   - `OAUTH_GOOGLE_CLIENT_SECRET`
   - `REDIS_PASSWORD` (if needed)

**Using in Workflows:**
```yaml
env:
  OAUTH_GOOGLE_CLIENT_ID: ${{ secrets.OAUTH_GOOGLE_CLIENT_ID }}
```

### Environment Variables Best Practices

1. **Use .env.example as template**
   ```bash
   cp .env.example .env
   # Then edit .env with actual values
   ```

2. **Document all required variables**
   - Keep .env.example up to date
   - Add comments explaining each variable
   - Never put real values in .env.example

3. **Use strong, unique credentials**
   - Different credentials for dev/staging/production
   - Rotate credentials regularly (every 90 days recommended)
   - Use strong random passwords for Redis

## Branch Protection

Configure these rules for the `main` branch:

### Settings → Branches → Branch protection rules

```
Branch name pattern: main

☑ Require a pull request before merging
  ☑ Require approvals (at least 1)
  ☑ Dismiss stale pull request approvals when new commits are pushed
  ☑ Require review from Code Owners (enforces .github/CODEOWNERS)

☑ Require status checks to pass before merging
  ☑ Require branches to be up to date before merging
  Status checks:
    - build-and-push (if you add a CI workflow)
    - tests (if you add a test workflow)

☑ Require conversation resolution before merging

☑ Require signed commits (recommended)

☑ Require linear history

☑ Do not allow bypassing the above settings
```

### For Solo Projects

If you're the only contributor, minimum protection:
```
☑ Require status checks to pass before merging
☑ Do not allow bypassing the above settings
```

## GitHub Actions Security

### Workflow Security Best Practices

1. **Pin action versions** (already done in our workflow):
   ```yaml
   # Good
   uses: actions/checkout@v4

   # Better
   uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
   ```

2. **Minimal permissions**:
   ```yaml
   permissions:
     contents: read
     packages: write
   ```

3. **Review third-party actions**
   - Only use actions from verified publishers
   - Check the action's source code before using
   - Prefer official actions from: actions/, github/, docker/

4. **Protect workflow files**
   - Don't allow external contributors to modify workflows
   - Review workflow changes carefully

5. **Use environments for deployments**:
   ```yaml
   environment:
     name: production
     url: https://mcp.example.com
   ```

### GITHUB_TOKEN Security

The workflow uses `GITHUB_TOKEN` which is automatically provided:
- Expires after the workflow completes
- Scoped to the repository
- No need to create personal access tokens

## Dependency Security

### Automated Dependency Updates

1. **Enable Dependabot** (Settings → Security):
   - Automatic pull requests for dependency updates
   - Security updates have high priority

2. **Create `.github/dependabot.yml`**:
   ```yaml
   version: 2
   updates:
     - package-ecosystem: "gomod"
       directory: "/"
       schedule:
         interval: "weekly"
       open-pull-requests-limit: 10

     - package-ecosystem: "github-actions"
       directory: "/"
       schedule:
         interval: "weekly"
       open-pull-requests-limit: 10

     - package-ecosystem: "docker"
       directory: "/"
       schedule:
         interval: "weekly"
       open-pull-requests-limit: 5
   ```

### Manual Dependency Checks

```bash
# Check for outdated dependencies
go list -u -m all

# Audit for known vulnerabilities
go list -json -m all | nancy sleuth

# Update dependencies
go get -u ./...
go mod tidy
```

## Security Scanning

### 1. Set Up CodeQL Analysis

Create `.github/workflows/codeql-analysis.yml`:

```yaml
name: "CodeQL"

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 0 * * 1'  # Weekly on Mondays

jobs:
  analyze:
    name: Analyze
    runs-on: ubuntu-latest
    permissions:
      security-events: write
      contents: read

    steps:
    - name: Checkout repository
      uses: actions/checkout@v4

    - name: Initialize CodeQL
      uses: github/codeql-action/init@v3
      with:
        languages: go

    - name: Autobuild
      uses: github/codeql-action/autobuild@v3

    - name: Perform CodeQL Analysis
      uses: github/codeql-action/analyze@v3
```

### 2. Container Security Scanning

Add to your docker-publish workflow:

```yaml
- name: Run Trivy vulnerability scanner
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest
    format: 'sarif'
    output: 'trivy-results.sarif'

- name: Upload Trivy results to GitHub Security tab
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: 'trivy-results.sarif'
```

### 3. Pre-commit Hooks

Install pre-commit hooks to catch secrets before commit:

```bash
# Install git-secrets
git clone https://github.com/awslabs/git-secrets.git
cd git-secrets
make install

# Configure
git secrets --install
git secrets --register-aws
git secrets --add 'client_secret'
git secrets --add 'password.*='
```

## Security Headers and Runtime

### Application Security

The application should implement these security measures:

1. **HTTPS in Production**
   - Use reverse proxy (nginx, Caddy) with TLS
   - Redirect HTTP to HTTPS
   - Set `BASE_URL` to https:// in production

2. **Security Headers** (add to your Fiber app):
   ```go
   app.Use(func(c *fiber.Ctx) error {
       c.Set("X-Content-Type-Options", "nosniff")
       c.Set("X-Frame-Options", "DENY")
       c.Set("X-XSS-Protection", "1; mode=block")
       c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
       c.Set("Content-Security-Policy", "default-src 'self'")
       return c.Next()
   })
   ```

3. **Rate Limiting** (add to your app):
   ```go
   import "github.com/gofiber/fiber/v2/middleware/limiter"

   app.Use(limiter.New(limiter.Config{
       Max: 100,
       Expiration: 1 * time.Minute,
   }))
   ```

### Redis Security

1. **Use password authentication**:
   ```env
   REDIS_PASSWORD=<strong-random-password>
   ```

2. **Use TLS for Redis connection** (in production):
   ```go
   redis.NewClient(&redis.Options{
       TLSConfig: &tls.Config{
           MinVersion: tls.VersionTLS12,
       },
   })
   ```

3. **Network isolation**:
   - Don't expose Redis port publicly
   - Use Docker networks or VPC
   - Bind Redis to localhost only

## Incident Response

### If a Secret is Leaked

1. **Immediate Actions** (within 1 hour):
   - Rotate the compromised credential immediately
   - Revoke old credential in the provider (Google Cloud Console)
   - Check access logs for unauthorized usage
   - Update all production deployments with new credentials

2. **Investigation** (within 24 hours):
   - Determine scope of exposure
   - Review logs for suspicious activity
   - Check if credential was used maliciously
   - Document the incident

3. **Prevention** (within 1 week):
   - Clean git history if needed
   - Enable push protection
   - Add pre-commit hooks
   - Update documentation

### If a Vulnerability is Found

1. **For Public Vulnerabilities**:
   - Create a GitHub Security Advisory
   - Fix in a private fork
   - Release patch version
   - Publish advisory with fix

2. **For Private Vulnerabilities**:
   - Enable private vulnerability reporting
   - Communicate with reporter privately
   - Develop and test fix
   - Coordinate disclosure

## Security Checklist for Going Public

- [ ] All secrets removed from code and git history
- [ ] `.gitignore` properly configured
- [ ] `.env.example` created (without real values)
- [ ] GitHub secret scanning enabled
- [ ] Push protection enabled
- [ ] Dependabot enabled
- [ ] Branch protection rules configured
- [ ] CodeQL analysis configured
- [ ] Security policy (SECURITY.md) published
- [ ] All production credentials rotated
- [ ] Team educated on security practices

## Reporting Security Issues

**DO NOT** open public issues for security vulnerabilities.

Instead:
1. Use GitHub's private vulnerability reporting (if enabled)
2. Or email: [your-security-email@example.com]
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We aim to respond within 48 hours and provide a fix within 7 days for critical issues.

## References

- [GitHub Security Best Practices](https://docs.github.com/en/code-security)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Policy](https://go.dev/security/policy)
- [OAuth 2.0 Security Best Current Practice](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics)
- [Redis Security](https://redis.io/docs/management/security/)
