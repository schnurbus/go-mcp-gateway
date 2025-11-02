#!/bin/bash
# Security verification script
# Run this before making your first commit or before going public

set -e

echo "🔒 Security Verification Script"
echo "================================"
echo ""

ERRORS=0
WARNINGS=0

# Check 1: Verify .env is ignored
echo "📋 Check 1: Verifying .env is ignored by git..."
if git check-ignore .env > /dev/null 2>&1; then
    echo "   ✓ .env is properly ignored"
else
    echo "   ✗ ERROR: .env is NOT ignored by git!"
    ERRORS=$((ERRORS + 1))
fi
echo ""

# Check 2: Verify .env is not staged
echo "📋 Check 2: Verifying .env is not staged for commit..."
if git diff --staged --name-only | grep -q "^\.env$"; then
    echo "   ✗ ERROR: .env is staged for commit!"
    echo "   Run: git reset .env"
    ERRORS=$((ERRORS + 1))
else
    echo "   ✓ .env is not staged"
fi
echo ""

# Check 3: Search for secrets in code files (excluding markdown docs)
echo "📋 Check 3: Searching for secrets in staged code files..."
FOUND_SECRETS=0

# Get staged files excluding markdown documentation and this script
STAGED_CODE_FILES=$(git diff --staged --name-only --diff-filter=ACM | grep -v "\.md$" | grep -v "SECURITY" | grep -v "\.example$" | grep -v "verify-security.sh")

if [ -n "$STAGED_CODE_FILES" ]; then
    # Check for Google OAuth client secrets in code files only
    if echo "$STAGED_CODE_FILES" | xargs git diff --staged | grep -i "GOCSPX-" > /dev/null 2>&1; then
        echo "   ✗ ERROR: Found Google OAuth client secret in staged code files!"
        FOUND_SECRETS=1
    fi

    # Check for generic secret patterns in code files
    if echo "$STAGED_CODE_FILES" | xargs git diff --staged | grep -E "(password|secret|api_key|token).*=.*['\"][a-zA-Z0-9]{30,}['\"]" > /dev/null 2>&1; then
        echo "   ⚠ WARNING: Found potential secrets in staged code files"
        echo "   Please review manually"
        WARNINGS=$((WARNINGS + 1))
    fi
fi

if [ $FOUND_SECRETS -eq 0 ]; then
    echo "   ✓ No secrets found in staged code files"
else
    ERRORS=$((ERRORS + 1))
fi
echo ""

# Check 4: Verify .env.example has no real secrets
echo "📋 Check 4: Verifying .env.example has no real secrets..."
if grep -q "GOCSPX-" .env.example 2>/dev/null; then
    echo "   ✗ ERROR: Found real secret in .env.example!"
    ERRORS=$((ERRORS + 1))
elif grep -q "641764879875" .env.example 2>/dev/null; then
    echo "   ✗ ERROR: Found real client ID in .env.example!"
    ERRORS=$((ERRORS + 1))
else
    echo "   ✓ .env.example contains no real secrets"
fi
echo ""

# Check 5: Verify required files exist
echo "📋 Check 5: Verifying security files exist..."
MISSING_FILES=0
for file in .gitignore SECURITY.md .github/dependabot.yml .github/workflows/codeql-analysis.yml; do
    if [ ! -f "$file" ]; then
        echo "   ✗ Missing: $file"
        MISSING_FILES=1
    fi
done

if [ $MISSING_FILES -eq 0 ]; then
    echo "   ✓ All security files present"
else
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# Check 6: List files to be committed
echo "📋 Check 6: Files staged for commit:"
git diff --staged --name-only | while read file; do
    echo "   - $file"
done
echo ""

# Check 7: Verify no large binary files
echo "📋 Check 7: Checking for large files..."
LARGE_FILES=$(git diff --staged --stat | awk '{if ($3 ~ /\|/ && $5 > 1000000) print $1}')
if [ -n "$LARGE_FILES" ]; then
    echo "   ⚠ WARNING: Found large files (>1MB):"
    echo "$LARGE_FILES" | while read file; do
        echo "   - $file"
    done
    WARNINGS=$((WARNINGS + 1))
else
    echo "   ✓ No large files detected"
fi
echo ""

# Summary
echo "================================"
echo "📊 Summary"
echo "================================"
echo "Errors: $ERRORS"
echo "Warnings: $WARNINGS"
echo ""

if [ $ERRORS -gt 0 ]; then
    echo "❌ SECURITY VERIFICATION FAILED"
    echo ""
    echo "Please fix the errors above before committing."
    echo "Your credentials may be exposed if you continue!"
    exit 1
elif [ $WARNINGS -gt 0 ]; then
    echo "⚠️  VERIFICATION PASSED WITH WARNINGS"
    echo ""
    echo "Please review the warnings above."
    echo "You may proceed with caution."
    exit 0
else
    echo "✅ SECURITY VERIFICATION PASSED"
    echo ""
    echo "Your repository is ready for commit!"
    echo ""
    echo "Next steps:"
    echo "1. git commit -m 'Initial commit'"
    echo "2. git push"
    echo "3. Enable GitHub security features (see .github/SECURITY_CHECKLIST.md)"
    exit 0
fi
