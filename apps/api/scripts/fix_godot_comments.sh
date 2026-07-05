#!/bin/bash
# fix_godot_comments.sh - Fixes comments without trailing period
set -e
TARGET_DIR="${1:-.}"
echo "Fixing godot comments..."
find "$TARGET_DIR" -name "*.go" -type f | grep -v "cmd/verify" | while read -r file; do
    sed -i -E 's|^(\s*//[^/!][a-zA-Z0-9])$|\1.|g' "$file"
done
echo "Done!"
