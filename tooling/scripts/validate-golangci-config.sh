#!/bin/bash
# =============================================
# GOLANGCI-LINT CONFIG VALIDATOR
# Prevents disabling critical linters
# =============================================

set -euo pipefail

CONFIG_FILE="apps/api/.golangci.yml"

echo "🔒 Validating golangci-lint configuration..."

# Linters that MUST remain enabled (critical for code quality)
REQUIRED_LINTERS=(
    "errcheck"
    "unused"
    "staticcheck"
    "govet"
    "gosec"
    "bodyclose"
    "revive"
    "gofmt"
)

echo "Checking for required linters in $CONFIG_FILE..."

# Check if config file exists
if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "❌ ERROR: $CONFIG_FILE not found!"
    exit 1
fi

# Check that disable-all is not removing required linters
if grep -q "disable-all: true" "$CONFIG_FILE"; then
    echo "✅ Found 'disable-all: true' - checking required linters are enabled..."
    
    for linter in "${REQUIRED_LINTERS[@]}"; do
        if grep -E "^\s*-\s+$linter$" "$CONFIG_FILE" > /dev/null 2>&1; then
            echo "   ✅ $linter is enabled"
        else
            echo "   ❌ ERROR: Required linter '$linter' is not enabled!"
            echo "   This linter is critical for code quality."
            echo "   Please do NOT disable it."
            exit 1
        fi
    done
fi

# Check for suspicious patterns (disabled linters that should remain enabled)
SUSPICIOUS_PATTERNS=(
    "# - errcheck"
    "# - unused"
    "# - staticcheck"
    "# - gosec"
)

for pattern in "${SUSPICIOUS_PATTERNS[@]}"; do
    if grep -qF "$pattern" "$CONFIG_FILE"; then
        echo "❌ ERROR: Found suspicious pattern '$pattern'"
        echo "   Critical linter appears to be commented out (disabled)."
        echo "   This is not allowed for critical linters."
        exit 1
    fi
done

# Check minimum linter count (should have at least 20 enabled)
enabled_count=$(grep -cE "^\s*-\s+[a-z]" "$CONFIG_FILE" 2>/dev/null || echo "0")
echo "📊 Found $enabled_count linters enabled"

if [[ "$enabled_count" -lt 20 ]]; then
    echo "❌ WARNING: Very few linters enabled ($enabled_count). This seems suspicious."
    exit 1
fi

echo ""
echo "✅ golangci-lint configuration is VALID!"
echo "   All critical linters are enabled."
echo "   The linter config is LOCKED for quality enforcement."
echo ""

exit 0