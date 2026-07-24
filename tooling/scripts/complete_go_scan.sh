#!/bin/bash
# COMPLETE Go file scan - categorize ALL remaining files

cd "$(dirname "$0")/.."

echo ""
echo "                    COMPLETE Go FILE SCAN - ALL REMAINING FILES                                       "
echo ""
echo ""

echo ""
echo "1. ROOT Go FILES"
echo ""
echo ""

echo "Root level Go files:"
for f in *.go; do
    if [ -f "$f" ]; then
        echo "    $f"
    fi
done

echo ""
echo ""
echo "2. OLD internal/ FILES (to be phased)"
echo ""
echo ""

echo "OLD directories in internal/ (migrated to infrastructure):"
for dir in internal/audit internal/auth internal/fcm internal/metrics internal/ssr; do
    if [ -d "$dir" ]; then
        files=$(find "$dir" -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
        echo "    $dir/: $files files"
    fi
done

echo ""
echo "Root files in internal/:"
for f in internal/*.go; do
    if [ -f "$f" ]; then
        echo "    $f"
    fi
done

echo ""
echo ""
echo "3. OLD pkg/ FILES (to be phased)"
echo ""
echo ""

echo "pkg/config/:"
for f in pkg/config/*.go; do
    if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
        echo "    $(basename $f)"
    fi
done

echo ""
echo "pkg/crypto/:"
for f in pkg/crypto/*.go; do
    if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
        echo "    $(basename $f)"
    fi
done

echo ""
echo "pkg/logging/:"
for f in pkg/logging/*.go; do
    if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
        echo "    $(basename $f)"
    fi
done

echo ""
echo "pkg/models/:"
for f in pkg/models/*.go; do
    if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
        echo "    $(basename $f)"
    fi
done

echo ""
echo "pkg/storage/:"
for f in pkg/storage/*.go; do
    if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
        echo "    $(basename $f)"
    fi
done

echo ""
echo ""
echo "4. HANDLER FILES - OLD vs NEW"
echo ""
echo ""

echo "OLD flat handlers (in internal/api/handlers/*.go):"
OLD_HANDLERS=$(find internal/api/handlers -maxdepth 1 -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
echo "  Count: $OLD_HANDLERS files"
for f in internal/api/handlers/*.go; do
    if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
        echo "    $(basename $f)"
    fi
done

echo ""
echo "NEW organized handlers (in internal/api/handlers/*/):"
for dir in internal/api/handlers/*/; do
    if [ -d "$dir" ]; then
        name=$(basename "$dir")
        files=$(find "$dir" -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
        echo "   $name/: $files files"
    fi
done

echo ""
echo ""
echo "5. MIDDLEWARE FILES - OLD vs NEW"
echo ""
echo ""

echo "OLD middleware (in internal/api/middleware/):"
OLD_MW=$(find internal/api/middleware -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
echo "  Count: $OLD_MW files"

echo ""
echo "NEW middleware (in internal/infrastructure/middleware/):"
NEW_MW=$(find internal/infrastructure/middleware -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
echo "  Count: $NEW_MW files"

echo ""
echo ""
echo "6. MIGRATION STATUS SUMMARY"
echo ""
echo ""

echo ""
echo " Category                                             Status                                          "
echo ""

echo " NEW infrastructure/ (COMPLETE)                       $(find internal/infrastructure -name "*.go" -type f 2>/dev/null | wc -l) files                          "
echo " NEW domain/ (COMPLETE)                              $(find internal/domain -name "*.go" -type f 2>/dev/null | wc -l) files                             "
echo " NEW application/ (COMPLETE)                        $(find internal/application -name "*.go" -type f 2>/dev/null | wc -l) files                           "
echo " NEW organized handlers/ (IN PROGRESS)              $(find internal/api/handlers -maxdepth 1 -name "*.go" -type f 2>/dev/null | wc -l) files                         "
echo " OLD pkg/ (TO BE PHASED)                             $(find pkg -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l) files                          "
echo " OLD internal/ (TO BE PHASED)                        $(find internal/audit internal/auth internal/fcm internal/metrics internal/ssr internal/*.go -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l) files                         "
echo ""

echo ""
echo ""
echo "7. FILES STILL USING OLD pkg/ IMPORTS"
echo ""
echo ""

echo "Files importing pkg/models:"
USING_PKG_MODELS=$(find . -name "*.go" -type f 2>/dev/null | xargs grep -l "vyzorix/apps/api/pkg/models" 2>/dev/null | grep -v "_test" | wc -l)
echo "  Count: $USING_PKG_MODELS"

echo ""
echo "Files importing pkg/storage:"
USING_PKG_STORAGE=$(find . -name "*.go" -type f 2>/dev/null | xargs grep -l "vyzorix/apps/api/pkg/storage" 2>/dev/null | grep -v "_test" | wc -l)
echo "  Count: $USING_PKG_STORAGE"

echo ""
echo ""
echo "NEXT: Migrate handlers from flat to organized structure, then phase out pkg/"
echo ""
