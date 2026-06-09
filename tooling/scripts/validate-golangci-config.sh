#!/bin/bash
# =============================================
# GOLANGCI-LINT CONFIG VALIDATOR
# Prevents disabling critical linters
# Ensures aggressive linter configuration is enforced
# =============================================

set -euo pipefail

CONFIG_FILE="apps/api/.golangci.yml"

echo "[GOLANGCI GUARDIAN] Validating golangci-lint configuration..."

# Linters that MUST remain enabled (aggressive configuration - 42 linters)
REQUIRED_LINTERS=(
    "errcheck"
    "errorlint"
    "err113"
    "errname"
    "nilerr"
    "nilnesserr"
    "nilnil"
    "rowserrcheck"
    "sqlclosecheck"
    "unused"
    "staticcheck"
    "gosimple"
    "govet"
    "ineffassign"
    "copyloopvar"
    "predeclared"
    "gofmt"
    "gofumpt"
    "goimports"
    "godot"
    "misspell"
    "revive"
    "stylecheck"
    "gci"
    "decorder"
    "gocyclo"
    "funlen"
    "nestif"
    "maintidx"
    "gosec"
    "bodyclose"
    "exhaustive"
    "exhaustruct"
    "iface"
    "noctx"
    "contextcheck"
    "containedctx"
    "asciicheck"
    "bidichk"
    "dupword"
    "godox"
    "goprintffuncname"
    "errchkjson"
    "perfsprint"
    "paralleltest"
    "testableexamples"
    "testpackage"
    "testifylint"
    "prealloc"
    "mirror"
    "nakedret"
    "gomoddirectives"
    "gomodguard"
    "loggercheck"
    "sloglint"
)

echo "Checking for required linters in $CONFIG_FILE..."

# Check if config file exists
if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "[ERROR] $CONFIG_FILE not found!"
    exit 1
fi

# Check that disable-all is not removing required linters
if grep -q "disable-all: true" "$CONFIG_FILE"; then
    echo "[OK] Found 'disable-all: true' - checking required linters are enabled..."
    
    for linter in "${REQUIRED_LINTERS[@]}"; do
        if grep -E "^\s*-\s+$linter$" "$CONFIG_FILE" > /dev/null 2>&1; then
            echo "   [OK] $linter is enabled"
        else
            echo "   [ERROR] Required linter '$linter' is not enabled!"
            echo "   This linter is critical for code quality."
            echo "   Please do NOT disable it."
            exit 1
        fi
    done
fi

# Check for suspicious patterns (disabled linters that should remain enabled)
SUSPICIOUS_PATTERNS=(
    "# - errcheck"
    "# - errorlint"
    "# - err113"
    "# - unused"
    "# - staticcheck"
    "# - gosec"
    "# - godot"
    "# - revive"
    "# - gofmt"
)

for pattern in "${SUSPICIOUS_PATTERNS[@]}"; do
    if grep -qF "$pattern" "$CONFIG_FILE"; then
        echo "[ERROR] Found suspicious pattern '$pattern'"
        echo "   Critical linter appears to be commented out (disabled)."
        echo "   This is not allowed for critical linters."
        exit 1
    fi
done

# Check minimum linter count (should have at least 40 enabled for aggressive config)
enabled_count=$(grep -cE "^\s*-\s+[a-z]" "$CONFIG_FILE" 2>/dev/null || echo "0")
echo "[INFO] Found $enabled_count linters enabled"

if [[ "$enabled_count" -lt 40 ]]; then
    echo "[ERROR] Very few linters enabled ($enabled_count). Aggressive config requires at least 40."
    exit 1
fi

# Check for aggressive linter settings
echo ""
echo "Checking for aggressive linter settings..."

if grep -q "funlen" "$CONFIG_FILE"; then
    echo "   [OK] funlen (function length) linter is enabled"
else
    echo "   [WARNING] funlen linter is not enabled"
fi

if grep -q "gocyclo" "$CONFIG_FILE"; then
    echo "   [OK] gocyclo (cyclomatic complexity) linter is enabled"
else
    echo "   [WARNING] gocyclo linter is not enabled"
fi

if grep -q "errorlint" "$CONFIG_FILE"; then
    echo "   [OK] errorlint linter is enabled"
else
    echo "   [WARNING] errorlint linter is not enabled"
fi

echo ""
echo "[OK] golangci-lint configuration is VALID!"
echo "   All critical linters are enabled."
echo "   The linter config is LOCKED for quality enforcement."
echo ""

exit 0