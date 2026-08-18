#!/usr/bin/env bash
#
# Installs the pre-commit hook into the local .git/hooks/ directory.
# Run this once after cloning the repo:
#
#   bash tooling/hooks/install.sh
#
set -euo pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)
HOOK_SRC="$REPO_ROOT/tooling/hooks/pre-commit"
HOOK_DST="$REPO_ROOT/.git/hooks/pre-commit"

if [ ! -f "$HOOK_SRC" ]; then
  echo "Error: hook source not found at $HOOK_SRC"
  exit 1
fi

# Create hooks dir if it doesn't exist
mkdir -p "$REPO_ROOT/.git/hooks"

# Copy the hook
cp "$HOOK_SRC" "$HOOK_DST"
chmod +x "$HOOK_DST"

echo "Installed pre-commit hook to .git/hooks/pre-commit"
echo ""
echo "The hook will run on every 'git commit' that stages Go files under apps/api/."
echo "It runs: go build, go vet, golangci-lint v2.12.2"
echo ""
echo "To skip for a single commit (emergency): git commit --no-verify"
echo "To uninstall: rm .git/hooks/pre-commit"
