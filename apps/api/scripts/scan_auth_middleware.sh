#!/bin/bash
# Scan internal/auth/ and internal/api/middleware/

cd "$(dirname "$0")/.."

echo "╔════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
echo "║                    SCANNING: internal/auth/ & internal/api/middleware/                                          ║"
echo "╚════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. internal/auth/ (JWT, OAuth, Password, TOTP, Rate Limiting, Session)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ -d "internal/auth" ]; then
    echo "Files:"
    for f in internal/auth/*.go; do
        if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
            funcs=$(grep -c "^func " "$f" 2>/dev/null || echo "0")
            types=$(grep -c "^type " "$f" 2>/dev/null || echo "0")
            echo "  $(basename $f): $types types, $funcs funcs"
        fi
    done
    
    echo ""
    echo "Subdirectories:"
    ls -la internal/auth/secretstore/ 2>/dev/null && echo "  secretstore/: Key storage" || echo "  (no secretstore)"
    
    echo ""
    echo "Key Logic:"
    grep -E "^type |^func " internal/auth/*.go 2>/dev/null | grep -v "_test" | head -40
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. internal/api/middleware/"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ -d "internal/api/middleware" ]; then
    echo "Files:"
    for f in internal/api/middleware/*.go; do
        if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
            funcs=$(grep -c "^func " "$f" 2>/dev/null || echo "0")
            types=$(grep -c "^type " "$f" 2>/dev/null || echo "0")
            echo "  $(basename $f): $types types, $funcs funcs"
        fi
    done
    
    echo ""
    echo "Key Logic:"
    grep -E "^func " internal/api/middleware/*.go 2>/dev/null | grep -v "_test"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. MIGRATION STATUS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "┌────────────────────────────────────────────────┬────────────────────────────────────────┐"
echo "│ Component                                        │ Status                                  │"
echo "├────────────────────────────────────────────────┼────────────────────────────────────────┤"

# Check if auth services exist in application layer
if [ -f "internal/application/auth/service.go" ]; then
    echo "│ JWT/Session Management                           │ ✅ In application/auth/service.go     │"
else
    echo "│ JWT/Session Management                           │ ⚠️  internal/auth/ (needs migration) │"
fi

# Check OAuth
if grep -rq "Google\|GitHub\|OAuth" internal/application/ 2>/dev/null; then
    echo "│ OAuth (Google, GitHub)                          │ ✅ In application layer               │"
else
    echo "│ OAuth (Google, GitHub)                          │ ⚠️  internal/auth/ (needs migration) │"
fi

# Check password
if grep -rq "HashPassword\|VerifyPassword" internal/infrastructure/auth/ 2>/dev/null; then
    echo "│ Password Hashing                                │ ✅ In infrastructure/auth/            │"
else
    echo "│ Password Hashing                                │ ⚠️  internal/auth/ (needs migration) │"
fi

# Check TOTP
if grep -rq "TOTP\|Totp\|mfa" internal/domain/ 2>/dev/null || grep -rq "TOTP" internal/application/ 2>/dev/null; then
    echo "│ TOTP/MFA                                       │ ✅ In domain/application layer       │"
else
    echo "│ TOTP/MFA                                       │ ⚠️  internal/auth/ (needs migration) │"
fi

# Check rate limiting
if grep -rq "RateLimit" internal/application/ 2>/dev/null || grep -rq "RateLimit" internal/infrastructure/ 2>/dev/null; then
    echo "│ Rate Limiting                                  │ ✅ In infrastructure/application     │"
else
    echo "│ Rate Limiting                                  │ ⚠️  internal/auth/ (needs migration) │"
fi

# Check middleware
if [ -d "internal/api/middleware" ]; then
    files=$(find internal/api/middleware -name "*.go" | grep -v "_test" | wc -l)
    echo "│ API Middleware                                  │ ⚠️  internal/api/middleware/ ($files files) │"
fi

echo "└────────────────────────────────────────────────┴────────────────────────────────────────┘"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. MISSING FROM NEW STRUCTURE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

MISSING=""
if [ -f "internal/auth/jwt.go" ] && ! grep -rq "func.*GenerateToken\|func.*ValidateToken" internal/application/ 2>/dev/null; then
    echo "❌ JWT token generation/validation"
    MISSING="yes"
fi

if [ -f "internal/auth/google_token.go" ] && ! grep -rq "GoogleID\|google" internal/domain/ 2>/dev/null; then
    echo "❌ Google OAuth token verification"
    MISSING="yes"
fi

if [ -f "internal/auth/github.go" ] && ! grep -rq "GitHub\|github" internal/domain/ 2>/dev/null; then
    echo "❌ GitHub OAuth"
    MISSING="yes"
fi

if [ -f "internal/auth/lockout.go" ]; then
    echo "❌ Account lockout"
    MISSING="yes"
fi

if [ -f "internal/auth/totp.go" ]; then
    echo "❌ TOTP/MFA"
    MISSING="yes"
fi

if [ -f "internal/auth/ratelimit.go" ]; then
    echo "❌ Rate limiting"
    MISSING="yes"
fi

if [ -f "internal/auth/origin.go" ]; then
    echo "❌ Origin/CORS validation"
    MISSING="yes"
fi

if [ -f "internal/auth/secretstore/" ]; then
    echo "❌ Secret store"
    MISSING="yes"
fi

if [ -z "$MISSING" ]; then
    echo "✅ All auth logic migrated!"
fi

echo ""
echo "════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
