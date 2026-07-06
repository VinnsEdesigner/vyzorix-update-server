#!/bin/bash
# Gap Analysis: OLD pkg/models/ vs NEW internal/domain/
# Shows migration status of all types

set -e
cd "$(dirname "$0")/.."

echo ""
echo "                    GAP ANALYSIS: OLD pkg/models/ vs NEW internal/domain/              "
echo ""
echo ""

# Count migrated types
MIGRATED=0
declare -a MIGRATED_LIST
declare -a MISSING_LIST

for f in pkg/models/*.go; do
    if ! echo "$f" | grep -q "_test" && [ -f "$f" ]; then
        while IFS= read -r line; do
            type=$(echo "$line" | sed "s/type \([^ ]*\).*/\1/" | tr -d " {")
            if [ -n "$type" ]; then
                # Search in NEW (domain + api/responses)
                result=$(grep -rl "type $type " internal/domain/ internal/api/responses/ 2>/dev/null)
                if [ -n "$result" ]; then
                    MIGRATED=$((MIGRATED + 1))
                    domain=$(dirname "$result" | sed "s|internal/||" | xargs basename)
                    MIGRATED_LIST+=(" $type → $domain")
                else
                    MISSING_LIST+=(" $type")
                fi
            fi
        done < <(grep "^type " "$f")
    fi
done

TOTAL=${#MIGRATED_LIST[@]}
MISSING=${#MISSING_LIST[@]}
let "TOTAL=TOTAL+MISSING"

echo ""
echo " FINAL SUMMARY"
echo ""
echo ""
echo "  Total types in OLD:    $TOTAL"
echo "  Migrated to NEW:       $MIGRATED"
echo "  Missing (not migrated): $MISSING"
echo ""

if [ $MISSING -eq 0 ]; then
    echo ""
    echo " ALL $TOTAL TYPES MIGRATED FROM OLD TO NEW STRUCTURE!"
    echo ""
else
    echo ""
    echo "MIGRATED TYPES:"
    echo ""
    for item in "${MIGRATED_LIST[@]}"; do echo "  $item"; done | sort
    
    echo ""
    echo ""
    echo "MISSING TYPES:"
    echo ""
    for item in "${MISSING_LIST[@]}"; do echo "  $item"; done | sort
fi
