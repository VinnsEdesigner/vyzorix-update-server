#!/bin/bash
echo "╔══════════════════════════════════════════════════════════════════════════════════════════╗"
echo "║                    GAP ANALYSIS: pkg/crypto vs internal/infrastructure/crypto         ║"
echo "╚══════════════════════════════════════════════════════════════════════════════════════════╝"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "OLD: pkg/crypto/hmac.go"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
grep -E "^type |^func " pkg/crypto/hmac.go | grep -v "_test"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "NEW: internal/infrastructure/crypto/"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
for f in internal/infrastructure/crypto/*.go; do
    if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
        echo "--- $(basename $f) ---"
        grep -E "^type |^func " "$f" | grep -v "_test"
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "MISSING FROM NEW:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

MISSING=""
for func in "NewNonceCache" "NonceCache" "Verifier" "ReadAndVerifyHTTP" "ReadAndVerify" "Verify"; do
    if ! grep -rq "$func" internal/infrastructure/crypto/ 2>/dev/null; then
        echo "  ❌ $func"
        MISSING="yes"
    fi
done

if [ -z "$MISSING" ]; then
    echo "  ✅ All crypto functions migrated!"
else
    echo ""
    echo "  ⚠️  These functions need migration to internal/infrastructure/crypto/"
fi
