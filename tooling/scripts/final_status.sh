#!/bin/bash
# FINAL COMPREHENSIVE STATUS - ALL DIRECTORIES

cd "$(dirname "$0")/.."

echo ""
echo "                    FINAL STATUS - COMPLETE DIRECTORY SCAN                                           "
echo ""
echo ""

echo ""
echo "1. NEW STRUCTURE (MIGRATION COMPLETE)"
echo ""
echo ""

echo "internal/domain/:"
for d in internal/domain/*/; do
    if [ -d "$d" ]; then
        name=$(basename "$d")
        files=$(find "$d" -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
        echo "   $name/ ($files files)"
    fi
done

echo ""
echo "internal/application/:"
for d in internal/application/*/; do
    if [ -d "$d" ]; then
        name=$(basename "$d")
        echo "   $name/"
    fi
done

echo ""
echo "internal/infrastructure/:"
for d in $(ls -d internal/infrastructure/*/ 2>/dev/null | sort); do
    name=$(basename "$d")
    files=$(find "$d" -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
    echo "   $name/ ($files files)"
done

echo ""
echo "internal/api/responses/:"
for f in internal/api/responses/*.go; do
    if [ -f "$f" ]; then
        echo "   $(basename $f)"
    fi
done

echo ""
echo ""
echo "2. OLD STRUCTURE (TO BE PHASED)"
echo ""
echo ""

echo "OLD pkg/:"
for d in pkg/*/; do
    if [ -d "$d" ]; then
        name=$(basename "$d")
        files=$(find "$d" -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
        echo "    $name/ ($files files)"
    fi
done

echo ""
echo "OLD internal/ directories:"
for dir in internal/audit internal/auth internal/fcm internal/metrics internal/ssr; do
    if [ -d "$dir" ]; then
        files=$(find "$dir" -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
        echo "    $dir/ ($files files)"
    fi
done

echo ""
echo "OLD internal/ root files:"
for f in internal/*.go; do
    if [ -f "$f" ]; then
        echo "    $(basename $f)"
    fi
done

echo ""
echo ""
echo "3. FILES STILL USING OLD IMPORTS"
echo ""
echo ""

echo "Files importing OLD packages:"
echo ""
echo "  pkg/models: $(find . -name "*.go" -type f 2>/dev/null | xargs grep -l 'vyzorix/apps/api/pkg/models' 2>/dev/null | grep -v "_test" | wc -l) files"
echo "  pkg/storage: $(find . -name "*.go" -type f 2>/dev/null | xargs grep -l 'vyzorix/apps/api/pkg/storage' 2>/dev/null | grep -v "_test" | wc -l) files"
echo "  pkg/config: $(find . -name "*.go" -type f 2>/dev/null | xargs grep -l 'vyzorix/apps/api/pkg/config' 2>/dev/null | grep -v "_test" | wc -l) files"
echo "  internal/auth: $(find . -name "*.go" -type f 2>/dev/null | xargs grep -l 'vyzorix/apps/api/internal/auth' 2>/dev/null | grep -v "_test" | wc -l) files"
echo "  internal/ws: $(find . -name "*.go" -type f 2>/dev/null | xargs grep -l 'vyzorix/apps/api/internal/ws' 2>/dev/null | grep -v "_test" | wc -l) files"

echo ""
echo ""
echo "4. SUMMARY COUNTS"
echo ""
echo ""

NEW_TOTAL=$(find internal/infrastructure internal/domain internal/application internal/api/responses -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
OLD_TOTAL=$(find pkg internal/audit internal/auth internal/fcm internal/metrics internal/ssr internal -maxdepth 1 -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)

echo ""
echo " Category                                             Count                                          "
echo ""
echo " NEW infrastructure/domain/application             $NEW_TOTAL files                          "
echo " OLD pkg/internal (to be phased)                   $OLD_TOTAL files                          "
echo ""

echo ""
echo ""
echo "INFRASTRUCTURE MIGRATION:  COMPLETE"
echo "HANDLER MIGRATION:  IN PROGRESS"
echo "pkg/ PHASING:   PENDING"
echo ""
