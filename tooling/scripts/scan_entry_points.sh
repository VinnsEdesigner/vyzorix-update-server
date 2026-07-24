#!/bin/bash
# Scan cmd/api/main.go and remaining internal/ root files

cd "$(dirname "$0")/.."

echo ""
echo "                    SCANNING: cmd/api/main.go & Remaining Files                                      "
echo ""
echo ""

echo ""
echo "1. cmd/api/main.go (NEW Entry Point)"
echo ""
echo ""

if [ -f "cmd/api/main.go" ]; then
    echo "Lines: $(wc -l < cmd/api/main.go)"
    echo ""
    echo "Imports (NEW structure):"
    grep "^import" -A 40 cmd/api/main.go | grep -E "vyzorix.*internal" | head -20
    echo ""
    echo "Functions:"
    grep "^func " cmd/api/main.go | head -10
fi

echo ""
echo ""
echo "2. main.go (OLD Entry Point)"
echo ""
echo ""

if [ -f "main.go" ]; then
    echo "Lines: $(wc -l < main.go)"
    echo ""
    echo "Imports (OLD structure):"
    grep "^import" -A 30 main.go | grep -E "vyzorix.*pkg\|vyzorix.*internal" | head -20
fi

echo ""
echo ""
echo "3. internal/ ROOT FILES (to be migrated)"
echo ""
echo ""

echo "Files in internal/ root:"
for f in internal/*.go; do
    if [ -f "$f" ]; then
        name=$(basename "$f")
        funcs=$(grep -c "^func " "$f" 2>/dev/null || echo "0")
        types=$(grep -c "^type " "$f" 2>/dev/null || echo "0")
        
        # Check if migrated
        migrated=""
        case "$name" in
            "email.go") grep -q "type.*Service" internal/infrastructure/email/*.go 2>/dev/null && migrated=" migrated" || migrated="  pending" ;;
            "command_signer.go") grep -q "CommandSigner" internal/infrastructure/crypto/*.go 2>/dev/null && migrated=" migrated" || migrated="  pending" ;;
            *) migrated="  pending" ;;
        esac
        
        echo "  $name: $types types, $funcs funcs ($migrated)"
    fi
done

echo ""
echo ""
echo "4. ENTRY POINT COMPARISON"
echo ""
echo ""

echo ""
echo " Entry Point                                          Structure Used                              "
echo ""

if [ -f "cmd/api/main.go" ]; then
    NEW_IMPORTS=$(grep -c "internal/infrastructure\|internal/domain\|internal/application" cmd/api/main.go 2>/dev/null || echo "0")
    OLD_IMPORTS=$(grep -c "pkg/" cmd/api/main.go 2>/dev/null || echo "0")
    echo " cmd/api/main.go                                     NEW ($NEW_IMPORTS internal imports)        "
fi

if [ -f "main.go" ]; then
    OLD_IMPORTS=$(grep -c "pkg/" main.go 2>/dev/null || echo "0")
    echo " main.go                                              OLD ($OLD_IMPORTS pkg imports)          "
fi

echo ""

echo ""
echo ""
echo "5. SUMMARY"
echo ""
echo ""

echo "NEW Entry Point (cmd/api/main.go):"
if [ -f "cmd/api/main.go" ]; then
    NEW_INT=$(grep "vyzorix/apps/api/internal" cmd/api/main.go 2>/dev/null | wc -l)
    echo "   Uses NEW imports: $NEW_INT imports"
    echo "   Wire up NEW architecture"
fi

echo ""
echo "OLD Entry Point (main.go):"
if [ -f "main.go" ]; then
    OLD_PKG=$(grep "vyzorix/apps/api/pkg" main.go 2>/dev/null | wc -l)
    echo "    Uses OLD imports: $OLD_PKG imports"
    echo "    Needs update to use NEW structure"
fi

echo ""
echo "Pending migrations:"
for f in internal/email.go internal/command_signer.go; do
    if [ -f "$f" ]; then
        echo "    $f"
    fi
done

echo ""
echo ""
