#!/bin/bash
# =============================================
# TYPESCRIPT STRICT VALIDATOR
# Ensures strict TypeScript practices
# =============================================

set -euo pipefail

echo "[TYPESCRIPT STRICT] Running TypeScript strict validation..."

WARNINGS=0

# Check for 'any' type usage
echo "Checking for 'any' type usage..."
if grep -rE ":\s*any\b|type\s+any\s*=" apps/web/src --include="*.ts" --include="*.tsx" 2>/dev/null | grep -v "_test.ts" | grep -v "node_modules" > /dev/null; then
    ANY_COUNT=$(grep -rE ":\s*any\b|type\s+any\s*=" apps/web/src --include="*.ts" --include="*.tsx" 2>/dev/null | grep -v "_test.ts" | grep -v "node_modules" | wc -l)
    echo "   [WARNING] Found $ANY_COUNT 'any' type usages"
    echo "   Consider using 'unknown' instead for better type safety"
    WARNINGS=$((WARNINGS + ANY_COUNT))
fi

# Check for non-null assertions (!)
echo "Checking for non-null assertions..."
if grep -rE "!(\s|$|\(|;|\)|\])" apps/web/src --include="*.ts" --include="*.tsx" 2>/dev/null | grep -v "_test.ts" | grep -v "node_modules" > /dev/null; then
    NON_NULL_COUNT=$(grep -rE "!(\s|$|\(|;|\)|\])" apps/web/src --include="*.ts" --include="*.tsx" 2>/dev/null | grep -v "_test.ts" | grep -v "node_modules" | wc -l)
    echo "   [WARNING] Found $NON_NULL_COUNT non-null assertions (!)"
    echo "   Consider using proper null checks instead"
    WARNINGS=$((WARNINGS + NON_NULL_COUNT))
fi

# Check for @ts-ignore
echo "Checking for @ts-ignore comments..."
if grep -rE "@ts-ignore|@ts-expect-error" apps/web/src --include="*.ts" --include="*.tsx" 2>/dev/null | grep -v "_test.ts" > /dev/null; then
    IGNORE_COUNT=$(grep -rE "@ts-ignore|@ts-expect-error" apps/web/src --include="*.ts" --include="*.tsx" 2>/dev/null | grep -v "_test.ts" | wc -l)
    echo "   [WARNING] Found $IGNORE_COUNT TypeScript ignore comments"
    WARNINGS=$((WARNINGS + IGNORE_COUNT))
fi

# Check for TODO in TypeScript
echo "Checking for TODO comments..."
if grep -rE "TODO:|FIXME:" apps/web/src --include="*.ts" --include="*.tsx" 2>/dev/null | grep -v "_test.ts" > /dev/null; then
    TODO_COUNT=$(grep -rE "TODO:|FIXME:" apps/web/src --include="*.ts" --include="*.tsx" 2>/dev/null | grep -v "_test.ts" | wc -l)
    echo "   [INFO] Found $TODO_COUNT TODO/FIXME comments"
fi

# Check for missing prop types in React components
echo "Checking for missing PropTypes..."
if grep -rE "function\s+[A-Z][a-zA-Z]+\(" apps/web/src --include="*.tsx" 2>/dev/null | grep -v "_test.ts" > /dev/null; then
    COMPONENT_COUNT=$(grep -rE "function\s+[A-Z][a-zA-Z]+\(" apps/web/src --include="*.tsx" 2>/dev/null | grep -v "_test.ts" | wc -l)
    echo "   [INFO] Found $COMPONENT_COUNT React components (manual prop validation recommended)"
fi

# Verify tsconfig has strict mode
echo ""
echo "Checking tsconfig.json..."
if [[ -f "apps/web/tsconfig.json" ]]; then
    if grep -q '"strict": true' apps/web/tsconfig.json; then
        echo "   [OK] Strict mode is enabled"
    else
        echo "   [WARNING] Strict mode not enabled in tsconfig.json"
        echo "   Consider adding: \"strict\": true"
    fi
    
    if grep -q '"noImplicitAny": true' apps/web/tsconfig.json; then
        echo "   [OK] noImplicitAny is enabled"
    else
        echo "   [WARNING] noImplicitAny not enabled"
    fi
fi

echo ""
if [[ "$WARNINGS" -gt 0 ]]; then
    echo "[WARNING] TypeScript strict validation found $WARNINGS warnings"
    echo "   Review the warnings above for potential improvements"
else
    echo "[OK] TypeScript strict validation passed!"
fi

exit 0