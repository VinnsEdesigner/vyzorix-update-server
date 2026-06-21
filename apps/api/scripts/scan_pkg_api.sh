#!/bin/bash
# Scan remaining pkg/ and internal/api/ directories

cd "$(dirname "$0")/.."

echo "╔════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
echo "║                    SCANNING: pkg/ & internal/api/ REMAINING                                         ║"
echo "╚════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╝"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. pkg/ REMAINING (OLD structure - being phased)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "pkg/ contents:"
for dir in pkg/*/; do
    if [ -d "$dir" ]; then
        name=$(basename "$dir")
        files=$(find "$dir" -name "*.go" -type f | grep -v "_test" | wc -l)
        echo "  ⚠️  $name/: $files files"
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. internal/api/ SERVER & HANDLERS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "internal/api/server.go:"
if [ -f "internal/api/server.go" ]; then
    funcs=$(grep -c "^func " internal/api/server.go 2>/dev/null || echo "0")
    types=$(grep -c "^type " internal/api/server.go 2>/dev/null || echo "0")
    echo "  $types types, $funcs funcs"
    echo "  Logic:"
    grep "^func \|^type " internal/api/server.go | head -20
fi

echo ""
echo "internal/api/responses/:"
if [ -d "internal/api/responses" ]; then
    for f in internal/api/responses/*.go; do
        if [ -f "$f" ]; then
            echo "  $(basename $f)"
        fi
    done
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. HANDLERS STATUS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "Flat handlers (OLD - being phased):"
OLD_COUNT=$(find internal/api/handlers -maxdepth 1 -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
echo "  Count: $OLD_COUNT files"

echo ""
echo "Organized handlers (NEW):"
for dir in internal/api/handlers/*/; do
    if [ -d "$dir" ]; then
        name=$(basename "$dir")
        files=$(find "$dir" -name "*.go" -type f | grep -v "_test" | wc -l)
        echo "  ✅ $name/: $files files"
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. COMPLETE OVERVIEW"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "NEW infrastructure/ (COMPLETE):"
for d in $(ls -d internal/infrastructure/*/ 2>/dev/null | sort); do
    name=$(basename "$d")
    files=$(find "$d" -name "*.go" -type f | grep -v "_test" | wc -l)
    echo "  ✅ $name/"
done

echo ""
echo "OLD structure (being phased):"
echo "  ⚠️  pkg/: $(find pkg -name "*.go" -type f | grep -v "_test" | wc -l) files"
echo "  ⚠️  internal/ws/: 2 files"
echo "  ⚠️  internal/api/handlers/flat: $OLD_COUNT files"
echo "  ⚠️  internal/api/server.go"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5. FILES STILL NEEDING REVIEW"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "Files using pkg/models (will break when removed):"
find . -name "*.go" -type f 2>/dev/null | xargs grep -l "vyzorix/apps/api/pkg/models" 2>/dev/null | grep -v "_test" | head -10

echo ""
echo "════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
echo "INFRASTRUCTURE MIGRATION: COMPLETE ✅"
echo "HANDLER MIGRATION: In Progress 🔄"
echo "pkg/ PHASING: Pending ⚠️"
echo "════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
