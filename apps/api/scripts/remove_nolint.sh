#!/bin/bash
# Script to remove all nolint, noinline, and lint: comments from Go files

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Removing linter suppression comments..."
echo "=========================================="
echo ""

# Patterns to remove (lines containing only these comments, or the comments within lines)
# We handle multiple patterns:
# 1. Lines that are ONLY a nolint/noinline comment
# 2. Lines that have nolint/noinline at the end after other code

# For lines that are ONLY the comment (no code before //).
echo -e "${YELLOW}Removing standalone '// nolint' and '// noinline' comments...${NC}"

# Remove lines that are exactly a nolint/noinline comment (with optional whitespace)
find . -name "*.go" -type f -exec sed -i -E 's/^[[:space:]]*\/\/[[:space:]]*(nolint|noinline)[[:space:]]*$//' {} \;.

# Remove nolint/noinline from end of lines (when there's code before it)
find . -name "*.go" -type f -exec sed -i -E 's/[[:space:]]*\/\/[[:space:]]*(nolint|noinline)[[:space:]]*$//' {} \;.

# Handle //go:nolint and //go:noinline directives.
find . -name "*.go" -type f -exec sed -i -E 's/[[:space:]]*\/\/[[:space:]]*\/\/go:nolint[[:space:]]*$//' {} \;.
find . -name "*.go" -type f -exec sed -i -E 's/[[:space:]]*\/\/[[:space:]]*go:nolint[[:space:]]*$//' {} \;.
find . -name "*.go" -type f -exec sed -i -E 's/[[:space:]]*\/\/[[:space:]]*\/\/go:noinline[[:space:]]*$//' {} \;.
find . -name "*.go" -type f -exec sed -i -E 's/[[:space:]]*\/\/[[:space:]]*go:noinline[[:space:]]*$//' {} \;.

# Handle lint: directives (e.g., //lint:ST1001 this is fine).
find . -name "*.go" -type f -exec sed -i -E 's/[[:space:]]*\/\/[[:space:]]*lint:[A-Za-z]+[[:space:]]+.*$//' {} \;.

# Remove any resulting empty comment lines (/// becomes empty, or just //).
find . -name "*.go" -type f -exec sed -i -E 's/^[[:space:]]*\/\/[[:space:]]*$//' {} \;.

# Clean up multiple empty lines that may have been created
find . -name "*.go" -type f -exec sed -i -E '/^$/N;/^\n$/D' {} \;

echo -e "${GREEN} Linter suppression comments removed${NC}"
echo ""
echo "Review the changes with:"
echo "  git diff --stat"
echo ""
echo "Then run linters to check for issues:"
echo "  cd apps/api && golangci-lint run ./..."
