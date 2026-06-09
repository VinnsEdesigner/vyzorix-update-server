#!/bin/bash
# =============================================
# SECURITY VALIDATOR
# Checks for common security issues
# =============================================

set -euo pipefail

echo "[SECURITY VALIDATOR] Running security validation..."

ISSUES_FOUND=0

# Check for hardcoded secrets in code
echo "Checking for hardcoded secrets..."
SECRET_PATTERNS=(
    "password\s*=\s*['\"][^'\"]{1,20}['\"]"
    "api[_-]?key\s*=\s*['\"][^'\"]{10,}['\"]"
    "secret\s*=\s*['\"][^'\"]{10,}['\"]"
    "token\s*=\s*['\"][^'\"]{20,}['\"]"
)

for pattern in "${SECRET_PATTERNS[@]}"; do
    if grep -rE "$pattern" apps/api --include="*.go" 2>/dev/null | grep -v "_test.go" | grep -v "example" | grep -v "test" > /dev/null; then
        echo "   [WARNING] Found potential hardcoded secret (pattern: $pattern)"
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
    fi
done

# Check for SQL injection vulnerabilities
echo "Checking for potential SQL injection..."
if grep -rE "fmt\.Sprintf.*SELECT|fmt\.Sprintf.*INSERT|fmt\.Sprintf.*UPDATE" apps/api --include="*.go" 2>/dev/null | grep -v "_test.go" > /dev/null; then
    echo "   [WARNING] Found potential SQL injection risk (fmt.Sprintf with SQL)"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

# Check for eval usage
echo "Checking for eval() usage..."
if grep -rE "eval\s*\(" apps/web/src --include="*.ts" --include="*.tsx" 2>/dev/null > /dev/null; then
    echo "   [ERROR] eval() found in frontend code!"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

# Check for dangerous innerHTML usage
echo "Checking for dangerous DOM manipulation..."
if grep -rE "innerHTML\s*=" apps/web/src --include="*.ts" --include="*.tsx" 2>/dev/null | grep -v "dangerouslySetInnerHTML" > /dev/null; then
    echo "   [WARNING] Found innerHTML assignment (XSS risk)"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

# Check for TODO security comments
echo "Checking for security TODO comments..."
if grep -rE "TODO.*security|FIXME.*auth|XXX.*vuln" apps/api apps/web 2>/dev/null > /dev/null; then
    echo "   [WARNING] Found security-related TODO/FIXME comments"
fi

# Check .env.example exists and is complete
echo "Checking environment template..."
if [[ -f "apps/api/.env.example" ]]; then
    REQUIRED_ENV_VARS=(
        "DATABASE_URL"
        "JWT_SECRET"
        "TOKEN_SECRET"
        "FIREBASE_CREDENTIALS"
    )
    
    for var in "${REQUIRED_ENV_VARS[@]}"; do
        if ! grep -qE "^$var=" "apps/api/.env.example" 2>/dev/null; then
            echo "   [WARNING] Missing $var in .env.example"
        fi
    done
fi

# Check for console.log in production code (excluding tests)
echo "Checking for console.log in production..."
if grep -rE "console\.log\s*\(" apps/web/src --include="*.ts" --include="*.tsx" 2>/dev/null | grep -v "test" | grep -v "console\.warn\|console\.error" > /dev/null; then
    LOG_COUNT=$(grep -rE "console\.log\s*\(" apps/web/src --include="*.ts" --include="*.tsx" 2>/dev/null | grep -v "test" | wc -l)
    echo "   [WARNING] Found $LOG_COUNT console.log statements in production code"
fi

echo ""
if [[ "$ISSUES_FOUND" -gt 0 ]]; then
    echo "[WARNING] Security validation found $ISSUES_FOUND potential issues"
    echo "   Review the warnings above before proceeding"
    # Not failing - just warnings for now
else
    echo "[OK] Security validation passed!"
fi

exit 0