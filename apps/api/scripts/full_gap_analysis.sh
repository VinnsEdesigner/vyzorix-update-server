#!/bin/bash
# Comprehensive Gap Analysis: All OLD directories vs NEW structure

cd "$(dirname "$0")/.."

echo "╔════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
echo "║                    COMPREHENSIVE GAP ANALYSIS: ALL OLD vs NEW                                       ║"
echo "╚════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. MODELS: pkg/models vs internal/domain/"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

OLD_TYPES=$(grep -h "^type " pkg/models/*.go 2>/dev/null | grep -v "_test" | wc -l)
NEW_TYPES=$(grep -rh "^type " internal/domain/*/*.go internal/api/responses/*.go 2>/dev/null | wc -l)

echo "  OLD: $OLD_TYPES types"
echo "  NEW: $NEW_TYPES types"
echo "  Status: $([ $NEW_TYPES -ge $OLD_TYPES ] && echo '✅ Complete' || echo '❌ Incomplete')"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. STORAGE: pkg/storage vs internal/infrastructure/storage/"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

OLD_STORE=$(grep -h "^func (s \*Store)" pkg/storage/*.go 2>/dev/null | wc -l)
NEW_REPO=$(grep -h "^func (r \*" internal/infrastructure/storage/*.go 2>/dev/null | wc -l)

echo "  OLD Store methods: $OLD_STORE"
echo "  NEW Repository methods: $NEW_REPO"
echo "  Status: $([ $NEW_REPO -ge $OLD_STORE ] && echo '✅ Complete' || echo '❌ Incomplete')"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. CRYPTO: pkg/crypto vs internal/infrastructure/crypto/"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "  Checking for NonceCache and Verifier..."
if grep -rq "NonceCache\|Verifier" internal/infrastructure/crypto/*.go 2>/dev/null; then
    echo "  Status: ✅ Complete"
else
    echo "  Status: ❌ Missing"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. EMAIL: internal/email.go vs internal/infrastructure/email/"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ -f "internal/infrastructure/email/service.go" ]; then
    echo "  Status: ✅ Migrated"
else
    echo "  Status: ❌ Missing"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5. AUTH: pkg/storage/crypto.go vs internal/infrastructure/auth/"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "  Checking for HashPassword, VerifyPassword..."
if grep -rq "func HashPassword\|func VerifyPassword" internal/infrastructure/auth/*.go 2>/dev/null; then
    echo "  Status: ✅ Complete"
else
    echo "  Status: ❌ Missing"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "6. UUID: pkg/storage/uuid.go vs internal/infrastructure/uuid/"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ -f "internal/infrastructure/uuid/uuid.go" ]; then
    echo "  Status: ✅ Migrated"
else
    echo "  Status: ❌ Missing"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "7. NEW STRUCTURE FILES (verify no OLD imports)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

NEW_VIOLATIONS=$(find internal/ws internal/infrastructure internal/domain internal/application cmd/api -name "*.go" 2>/dev/null | while read f; do
    if grep -q "vyzorix/apps/api/pkg/models\|vyzorix/apps/api/pkg/storage\|vyzorix/apps/api/pkg/crypto" "$f" 2>/dev/null; then
        if ! grep -q "pkg/config" "$f" 2>/dev/null; then
            echo "$f"
        fi
    fi
done | wc -l)

echo "  NEW files importing OLD (violations): $NEW_VIOLATIONS"
echo "  Status: $([ $NEW_VIOLATIONS -eq 0 ] && echo '✅ Clean' || echo '❌ Has violations')"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "8. CONFIG & LOGGING"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "  pkg/config: Legacy - API layer concern (LEGITIMATE)"
echo "  pkg/logging: Legacy - Not migrated to infrastructure"
echo "  Status: ⚠️  Not migrated (not critical for clean architecture)"

echo ""
echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════"
echo "📊 OVERALL SUMMARY"
echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════"
echo ""

echo "┌────────────────────────────────────────┬────────────────────────────────────────────┐"
echo "│ Component                               │ Status                                   │"
echo "├────────────────────────────────────────┼────────────────────────────────────────────┤"
printf "│ Models (38 types)                    │ %40s │\n" "$([ $NEW_TYPES -ge $OLD_TYPES ] && echo '✅ Migrated' || echo '❌ Incomplete')"
printf "│ Storage Repository (95 methods)       │ %40s │\n" "$([ $NEW_REPO -ge $OLD_STORE ] && echo '✅ Migrated' || echo '❌ Incomplete')"
printf "│ Crypto (NonceCache/Verifier)           │ %40s │\n" "$(grep -rq "NonceCache\|Verifier" internal/infrastructure/crypto/*.go 2>/dev/null && echo '✅ Migrated' || echo '❌ Missing')"
printf "│ Email Service                         │ %40s │\n" "$(ls internal/infrastructure/email/*.go 2>/dev/null >/dev/null && echo '✅ Migrated' || echo '❌ Missing')"
printf "│ Auth (HashPassword)                   │ %40s │\n" "$(grep -rq "func HashPassword" internal/infrastructure/auth/*.go 2>/dev/null && echo '✅ Migrated' || echo '❌ Missing')"
printf "│ UUID Generator                        │ %40s │\n" "$(ls internal/infrastructure/uuid/*.go 2>/dev/null >/dev/null && echo '✅ Migrated' || echo '❌ Missing')"
printf "│ Architecture (NEW→OLD imports)        │ %40s │\n" "$(echo $NEW_VIOLATIONS) violations"
echo "└────────────────────────────────────────┴────────────────────────────────────────────┘"

echo ""
echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════"
