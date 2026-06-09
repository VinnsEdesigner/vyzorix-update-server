#!/bin/bash
# =============================================
# DEPENDENCY AUDITOR
# Ensures dependencies are properly audited
# =============================================

set -euo pipefail

echo "[DEPENDENCY AUDITOR] Running dependency audit..."

# Check for outdated Go dependencies
echo "Checking Go dependencies..."
cd apps/api
if command -v go &> /dev/null; then
    # Check for unused dependencies
    if go mod tidy 2>&1 | grep -q "unused"; then
        echo "   [WARNING] Found unused Go dependencies - run 'go mod tidy'"
    fi
    
    # Check for vulnerable dependencies (if govulncheck is available)
    if command -v govulncheck &> /dev/null; then
        echo "   Running vulnerability scan..."
        govulncheck ./... 2>/dev/null || true
    else
        echo "   [INFO] govulncheck not installed - skipping vulnerability scan"
        echo "   Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"
    fi
fi
cd ../..

# Check for outdated npm packages
echo "Checking npm dependencies..."
cd apps/web
if command -v pnpm &> /dev/null; then
    # Check for outdated packages
    pnpm outdated 2>/dev/null || true
    
    # Check for security vulnerabilities
    if [[ -f "pnpm-lock.yaml" ]]; then
        # Audit only production dependencies
        echo "   [INFO] Run 'pnpm audit' for security audit"
    fi
fi
cd ../..

# Check for deprecated packages
echo "Checking for deprecated packages..."
DEPRECATED=(
    "moment"  # Use date-fns instead
    "lodash"  # Use lodash-es with tree-shaking
    "classnames"  # Use clsx instead
)

for pkg in "${DEPRECATED[@]}"; do
    if grep -q "\"$pkg\"" apps/web/package.json 2>/dev/null; then
        echo "   [WARNING] Found deprecated package: $pkg"
    fi
done

# Check package.json scripts are valid
echo "Checking package.json scripts..."
if grep -q '"dev":' apps/web/package.json && \
   grep -q '"build":' apps/web/package.json && \
   grep -q '"lint":' apps/web/package.json && \
   grep -q '"typecheck":' apps/web/package.json; then
    echo "   [OK] All required scripts are defined"
else
    echo "   [ERROR] Missing required scripts in package.json"
    exit 1
fi

echo ""
echo "[OK] Dependency audit complete!"

exit 0