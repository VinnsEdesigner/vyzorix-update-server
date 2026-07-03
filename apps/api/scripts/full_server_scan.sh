#!/bin/bash
# Comprehensive scan: ALL OLD files vs NEW structure

cd "$(dirname "$0")/.."

echo "╔══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
echo "║                    COMPLETE SERVER SCAN: ALL FILES vs NEW STRUCTURE                                              ║"
echo "╚══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╝"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. MAIN ENTRY POINTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "main.go (OLD - uses pkg/):"
echo "  Imports:"
grep "^import" -A 30 main.go | grep -E "vyzorix.*pkg" | head -5
echo ""

echo "cmd/api/main.go (NEW - uses internal/):"
echo "  Imports:"
grep "^import" -A 40 cmd/api/main.go | grep -E "vyzorix.*internal" | head -10
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. REMAINING OLD internal/ FILES"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "internal/ directories still with OLD logic:"
find internal -maxdepth 1 -type d 2>/dev/null | while read dir; do
    if [ -d "$dir" ] && [ "$dir" != "internal" ]; then
        name=$(basename "$dir")
        if [ ! -d "internal/infrastructure/$name" ] && [ ! -d "internal/domain/$name" ] && [ ! -d "internal/application/$name" ]; then
            files=$(find "$dir" -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
            echo "  ⚠️  $name/: $files files (NO migration exists)"
        fi
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. INTERNAL/API/"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "internal/api/server.go:"
if [ -f "internal/api/server.go" ]; then
    echo "  Status: ⚠️  API Server (needs NEW version)"
    grep -E "^func |^type " internal/api/server.go | head -10
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. HANDLERS STATUS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

OLD_HANDLERS=$(find internal/api/handlers -maxdepth 1 -name "auth_*.go" -o -name "device.go" -o -name "command.go" -o -name "admin_*.go" 2>/dev/null | grep -v "_test" | wc -l)
NEW_HANDLERS=$(find internal/api/handlers -maxdepth 1 -type d 2>/dev/null | wc -l)

echo "OLD flat handlers: $OLD_HANDLERS files"
echo "NEW organized handlers: $NEW_HANDLERS directories"
echo ""

echo "Flat handlers still using pkg/models:"
find internal/api/handlers -maxdepth 1 -name "*.go" 2>/dev/null | while read f; do
    if grep -q "pkg/models\|pkg/storage" "$f" 2>/dev/null; then
        echo "  ⚠️  $(basename $f)"
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5. FILES USING OLD pkg/ IMPORTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "Files importing pkg/models:"
find . -name "*.go" -type f 2>/dev/null | xargs grep -l "vyzorix/apps/api/pkg/models" 2>/dev/null | grep -v "_test" | head -20

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "6. COMPLETE STRUCTURE COMPARISON"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "NEW STRUCTURE (✅ COMPLETE):"
echo ""
echo "  internal/infrastructure/"
ls -d internal/infrastructure/*/ 2>/dev/null | while read d; do
    name=$(basename "$d")
    files=$(find "$d" -name "*.go" | wc -l)
    echo "    ✅ $name/ ($files files)"
done

echo ""
echo "  internal/domain/"
ls -d internal/domain/*/ 2>/dev/null | while read d; do
    name=$(basename "$d")
    echo "    ✅ $name/"
done

echo ""
echo "  internal/application/"
ls -d internal/application/*/ 2>/dev/null | while read d; do
    name=$(basename "$d")
    echo "    ✅ $name/"
done

echo ""
echo "OLD STRUCTURE (⚠️ BEING PHASED):"
echo ""
echo "  pkg/"
ls -d pkg/*/ 2>/dev/null | while read d; do
    name=$(basename "$d")
    files=$(find "$d" -name "*.go" | wc -l)
    echo "    ⚠️  $name/ ($files files)"
done

echo ""
echo "  internal/"
find internal -maxdepth 1 -type d 2>/dev/null | while read d; do
    name=$(basename "$d")
    if [ "$name" != "internal" ]; then
        if [ ! -d "internal/infrastructure/$name" ] && [ ! -d "internal/domain/$name" ] && [ ! -d "internal/application/$name" ] && [ "$name" != "api" ]; then
            echo "    ⚠️  $name/"
        fi
    fi
done

echo ""
echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════"
echo "📊 SUMMARY"
echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════"
echo ""
echo "NEW structure:"
echo "  ✅ infrastructure/: 10 directories"
echo "  ✅ domain/: 10+ directories"
echo "  ✅ application/: 6 directories"
echo "  ✅ ws/: WebSocket hub"
echo "  ✅ api/responses/: API responses"
echo ""
echo "OLD structure (being phased):"
echo "  ⚠️  pkg/: config, crypto, logging, models, storage"
echo "  ⚠️  internal/: audit, email, fcm, ssr, command_signer"
echo "  ⚠️  internal/api/handlers/: flat handlers"
echo "  ⚠️  main.go: OLD entry point"
echo ""
echo "Next step: Create NEW internal/api/server.go and cmd/api/main.go using NEW structure"
echo ""
echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════"
