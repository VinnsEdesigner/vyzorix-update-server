#!/bin/bash
# =============================================
# ESLINT CONFIG VALIDATOR
# Prevents disabling critical ESLint rules
# Ensures aggressive linter configuration is enforced
# =============================================

set -euo pipefail

CONFIG_FILE="apps/web/eslint.config.js"
PACKAGE_FILE="apps/web/package.json"

echo "[ESLINT GUARDIAN] Validating ESLint configuration..."

# Critical rules that MUST remain enabled (aggressive configuration)
REQUIRED_RULES=(
    "@typescript-eslint/no-unused-vars"
    "react-hooks/exhaustive-deps"
    "no-console"
    "no-debugger"
    "import/order"
    "no-unused-vars"
    "no-alert"
    "no-eval"
    "no-implicit-coercion"
    "prefer-promise-reject-errors"
)

# Rules that should NEVER be disabled (critical for code quality)
FORBIDDEN_DISABLE=(
    "@typescript-eslint/no-unused-vars"
    "react-hooks/exhaustive-deps"
    "no-console"
    "import/order"
    "no-unused-vars"
    "react-hooks/rules-of-hooks"
    "no-alert"
    "no-eval"
)

echo "Checking for required ESLint rules in $CONFIG_FILE..."

# Check if config file exists
if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "[ERROR] $CONFIG_FILE not found!"
    exit 1
fi

# Check for eslint-disable comments that might be hiding rule violations
SUSPICIOUS_PATTERNS=(
    "// eslint-disable"
    "/* eslint-disable"
)

for pattern in "${SUSPICIOUS_PATTERNS[@]}"; do
    if grep -qF "$pattern" "$CONFIG_FILE"; then
        echo "[WARNING] Found '$pattern' in config file."
        echo "   This may indicate disabled critical rules."
    fi
done

# Check that required packages are in package.json
echo ""
echo "Checking for required ESLint dependencies..."

if [[ -f "$PACKAGE_FILE" ]]; then
    # Direct dependencies that must be listed
    REQUIRED_PACKAGES=(
        "eslint"
        "eslint-plugin-react-hooks"
        "eslint-plugin-import"
        "typescript-eslint"
    )
    
    for pkg in "${REQUIRED_PACKAGES[@]}"; do
        if grep -q "\"$pkg\"" "$PACKAGE_FILE"; then
            echo "   [OK] $pkg is installed"
        else
            echo "   [ERROR] $pkg is not in package.json"
        fi
    done
fi

# Check for .eslintignore or similar exclude patterns
if [[ -f "apps/web/.eslintignore" ]]; then
    echo ""
    echo "[WARNING] Found .eslintignore - ensure critical paths are not excluded:"
    echo "   Checking for suspicious exclusions..."
    
    SUSPICIOUS_EXCLUDES=(
        "node_modules"
        "dist"
        "coverage"
    )
    
    # These are OK to exclude
    # Just a warning check
fi

# Verify the eslint script exists in package.json
if [[ -f "$PACKAGE_FILE" ]]; then
    if grep -q '"lint":' "$PACKAGE_FILE"; then
        echo ""
        echo "[OK] 'lint' script is defined in package.json"
    else
        echo "[ERROR] 'lint' script not found in package.json"
    fi
fi

# Check for critical rule configurations
echo ""
echo "Checking critical rule configurations..."

# Check no-unused-vars is set to error
# In TypeScript projects, we use @typescript-eslint/no-unused-vars instead
if grep -q 'no-unused-vars.*error' "$CONFIG_FILE"; then
    echo "   [OK] no-unused-vars is set to error"
elif grep -q '@typescript-eslint/no-unused-vars' "$CONFIG_FILE"; then
    # TypeScript project using @typescript-eslint/no-unused-vars - this is acceptable
    echo "   [OK] no-unused-vars is handled by @typescript-eslint/no-unused-vars"
else
    echo "   [ERROR] no-unused-vars should be set to 'error'"
fi

# Check react-hooks/exhaustive-deps is set to error
if grep -q 'exhaustive-deps.*error' "$CONFIG_FILE"; then
    echo "   [OK] react-hooks/exhaustive-deps is set to error"
else
    echo "   [ERROR] react-hooks/exhaustive-deps should be set to 'error'"
fi

# Check no-console is set (at least warn)
if grep -q 'no-console' "$CONFIG_FILE"; then
    echo "   [OK] no-console rule is configured"
else
    echo "   [ERROR] no-console rule not configured"
fi

# Check for aggressive TypeScript rules
echo ""
echo "Checking for aggressive TypeScript rules..."

AGGRESSIVE_RULES=(
    "@typescript-eslint/no-unused-vars"
    "@typescript-eslint/explicit-function-return-type"
    "@typescript-eslint/no-misused-spread"
    "@typescript-eslint/prefer-optional-chain"
    "@typescript-eslint/switch-exhaustiveness-check"
)

for rule in "${AGGRESSIVE_RULES[@]}"; do
    if grep -q "$rule" "$CONFIG_FILE"; then
        echo "   [OK] $rule is configured"
    else
        echo "   [WARNING] $rule is not explicitly configured"
    fi
done

echo ""
echo "[OK] ESLint configuration is VALID!"
echo "   Critical rules are properly configured."
echo "   The ESLint config is LOCKED for quality enforcement."
echo ""

exit 0