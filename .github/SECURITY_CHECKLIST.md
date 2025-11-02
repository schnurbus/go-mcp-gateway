# Security Checklist Before Going Public

## Critical - Do Before First Push

### 1. Verify No Secrets in Repository
```bash
# Check what will be committed
git status

# Verify .env is NOT in the list
# It should be ignored by .gitignore

# Double check .env is properly ignored
git check-ignore .env
# Should output: .env

# Search for any secrets in files to be committed
grep -r "client_secret" --exclude-dir=.git --exclude=".env*" --exclude="SECURITY.md" .
grep -r "GOCSPX-" --exclude-dir=.git --exclude=".env*" .
```

### 2. Verify Your Current Secrets
```bash
# Check if secrets appear anywhere except .env
grep -r "GOCSPX-" . --exclude-dir=.git --exclude=".env"
grep -r "client_secret" . --exclude-dir=.git --exclude=".env"
```

**⚠️ CRITICAL**: If any secrets are found in files OTHER than `.env` and security documentation, you MUST:
1. Remove them from those files
2. Rotate the credentials in Google Cloud Console
3. Update your `.env` with new credentials

### 3. First Commit Safety Check
```bash
# Add files to staging
git add .

# Verify .env is NOT staged
git status
# Should show: ".env" under "Untracked files" or not listed at all

# Preview what will be committed
git diff --staged --name-only

# If .env appears, STOP and run:
git reset .env
```

## GitHub Repository Setup (After First Push)

### Enable Security Features

1. **Go to Repository Settings → Code security and analysis**
   ```
   ☑ Dependency graph
   ☑ Dependabot alerts
   ☑ Dependabot security updates
   ☑ Secret scanning
   ☑ Push protection (CRITICAL!)
   ☑ Private vulnerability reporting
   ```

2. **Set up Branch Protection (Settings → Branches)**
   - Add rule for `main` branch
   - Minimum settings:
     ```
     ☑ Require a pull request before merging
       ☑ Require approvals: 1
       ☑ Require review from Code Owners
     ☑ Require status checks to pass before merging
     ```

3. **Configure Actions Permissions (Settings → Actions → General)**
   ```
   Workflow permissions:
   ○ Read repository contents and packages permissions (recommended)

   ☑ Allow GitHub Actions to create and approve pull requests
   ```

## Credential Rotation (Production)

Since your current `.env` contains real credentials that may have been viewed during development:

### Rotate Google OAuth Credentials

1. **Go to [Google Cloud Console](https://console.cloud.google.com/)**
2. Navigate to: APIs & Services → Credentials
3. Find your existing OAuth 2.0 Client ID (check your .env file for the CLIENT_ID value)
4. Click the trash icon to delete it
5. Create new OAuth 2.0 Client ID:
   - Application type: Web application
   - Name: go-mcp-gateway (production)
   - Authorized redirect URIs: Add your production callback URL
   - Click "Create"
6. Copy the new client ID and secret
7. Update your production `.env` file
8. **DO NOT commit the .env file**

### For Different Environments

Create separate OAuth credentials for each environment:

- **Development**: `go-mcp-gateway-dev`
  - Redirect URI: `http://localhost:8080/oauth/callback`

- **Staging**: `go-mcp-gateway-staging`
  - Redirect URI: `https://staging.yourdomain.com/oauth/callback`

- **Production**: `go-mcp-gateway-prod`
  - Redirect URI: `https://yourdomain.com/oauth/callback`

## Quick Security Verification

Run these commands before pushing:

```bash
# 1. Verify .env is ignored
git check-ignore .env && echo "✓ .env is properly ignored" || echo "✗ ERROR: .env is NOT ignored!"

# 2. Check for accidental secret commits
git grep -i "client_secret" $(git ls-files) && echo "✗ ERROR: Found secrets!" || echo "✓ No secrets found"

# 3. Verify only safe files are staged
git diff --staged --name-only

# 4. Check .env.example has no real values
grep "GOCSPX-" .env.example && echo "✗ ERROR: Real secret in .env.example!" || echo "✓ .env.example is safe"
```

## Post-Publication Monitoring

After making the repository public:

### Week 1: Daily Checks
- Check Security tab for any alerts
- Monitor Dependabot PRs
- Review any secret scanning alerts
- Check Actions workflow results

### Ongoing: Weekly Checks
- Review and merge Dependabot PRs
- Check Security advisories
- Monitor Actions workflow results
- Review access logs for unusual activity

## Emergency Response

### If Secret Scanning Detects a Secret

1. **Immediate (within 1 hour)**:
   ```bash
   # Remove the secret from the file
   # DO NOT just edit - the secret is in git history

   # You must clean the history:
   git filter-repo --path path/to/file --invert-paths
   git push --force
   ```

2. **Rotate the credential** in Google Cloud Console

3. **Check logs** for any unauthorized access

### If Someone Reports a Security Issue

1. Thank them
2. Assess severity
3. Create private security advisory
4. Fix in private
5. Release patch
6. Publish advisory

## Ready to Go Public Checklist

- [ ] ✓ .env is not committed (verified with `git status`)
- [ ] ✓ .env.example contains no real secrets
- [ ] ✓ .gitignore properly configured
- [ ] ✓ No secrets in any committed files
- [ ] ✓ Production credentials rotated
- [ ] ✓ Different credentials for dev/staging/prod
- [ ] ✓ Security features will be enabled after first push
- [ ] ✓ Team members understand security practices
- [ ] ✓ Monitoring plan in place

## First Commit Commands

Once all checks pass:

```bash
# Stage all files
git add .

# Final verification
git status
# Ensure .env is NOT in "Changes to be committed"

# Commit
git commit -m "Initial commit"

# Create main branch if needed
git branch -M main

# Add remote (replace with your repo URL)
git remote add origin https://github.com/schnurbus/go-mcp-gateway.git

# Push
git push -u origin main

# Immediately go to GitHub and enable security features!
```

## After First Push - Enable Security NOW

**IMMEDIATELY after pushing, go to GitHub and enable:**

1. Settings → Code security → Enable ALL security features
2. Settings → Branches → Add branch protection rules
3. Settings → Actions → Set correct permissions

This checklist ensures your repository is secure before and after going public.
