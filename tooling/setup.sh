#!/usr/bin/env bash
set -euo pipefail
REPO_ROOT=$(git rev-parse --show-toplevel)
HOOKS_DIR="$REPO_ROOT/tooling/hooks"
if [ ! -d "$HOOKS_DIR" ]; then
  echo "Error: hooks directory not found at $HOOKS_DIR"
  exit 1
fi
git config core.hooksPath "tooling/hooks"
chmod +x "$HOOKS_DIR"/*
echo "Hooks configured (core.hooksPath = tooling/hooks)"
echo "Run this once after cloning: bash tooling/setup.sh"
