#!/bin/bash
# Surgical analysis: Check for pkg/ imports in NEW structure
# Usage: ./scripts/check_pkg_imports.sh

set -e
cd "$(dirname "$0")/.."

echo ""
echo "           SURGICAL ANALYSIS: pkg/ imports in NEW structure              "
echo ""
echo ""

# Find all files importing from pkg/ (excluding the pkg/ directory itself and tests)
echo " Files still importing from pkg/:"
echo ""
grep -rl '"github.com/VinnsEdesigner/vyzorix/apps/api/pkg' --include="*.go" . 2>/dev/null | \
    grep -v "^./pkg/" | \
    grep -v "_test.go" | \
    sort | \
    while read file; do
        echo "  $file"
    done

echo ""
echo ""
echo " BREAKDOWN BY pkg/ PACKAGE:"
echo ""

# For each pkg subpackage, find which files import it
for pkg in config crypto email fcm models storage utils; do
    count=$(grep -rl '"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/'$pkg'' --include="*.go" . 2>/dev/null | grep -v "^./pkg/" | grep -v "_test.go" | wc -l)
    if [ "$count" -gt 0 ]; then
        echo ""
        echo " pkg/$pkg ($count files):"
        grep -rl '"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/'$pkg'' --include="*.go" . 2>/dev/null | \
            grep -v "^./pkg/" | \
            grep -v "_test.go" | \
            sort | \
            while read file; do
                echo "    → $file"
            done
    fi
done

echo ""
echo ""
echo " SUMMARY:"
echo ""

# Categorize files
old_handlers=$(grep -rl '"github.com/VinnsEdesigner/vyzorix/apps/api/pkg' --include="*.go" ./internal/api/handlers 2>/dev/null | grep -v "_test.go" | wc -l)
storage_refs=$(grep -rl '"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage' --include="*.go" . 2>/dev/null | grep -v "^./pkg/" | grep -v "_test.go" | wc -l)
model_refs=$(grep -rl '"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models' --include="*.go" . 2>/dev/null | grep -v "^./pkg/" | grep -v "_test.go" | wc -l)

echo "  • OLD handlers using pkg/...:      $old_handlers"
echo "  • pkg/storage references:          $storage_refs"  
echo "  • pkg/models references:           $model_refs"

echo ""
echo ""
echo " CRITICAL: Files importing pkg/storage OR pkg/models:"
echo ""
grep -rl '"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage\|pkg/models' --include="*.go" ./internal ./cmd 2>/dev/null | \
    grep -v "_test.go" | \
    grep -v "^./pkg/" | \
    sort | \
    while read file; do
        echo "   $file"
    done

echo ""
echo ""
echo " LEGITIMATE: Files importing pkg/config (expected):"
echo ""
grep -rl '"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config' --include="*.go" ./internal ./cmd 2>/dev/null | \
    grep -v "_test.go" | \
    grep -v "^./pkg/" | \
    sort | \
    while read file; do
        echo "   $file"
    done

echo ""
echo ""
echo " TOTAL: $(grep -rl '"github.com/VinnsEdesigner/vyzorix/apps/api/pkg' --include="*.go" . 2>/dev/null | grep -v "^./pkg/" | grep -v "_test.go" | wc -l) files need migration review"
echo ""
