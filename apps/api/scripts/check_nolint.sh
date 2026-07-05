#!/bin/bash
# Script to find all nolint and noinline comments in Go files

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Checking for nolint/noinline comments..."
echo "=========================================="
echo ""

# Find Go files and search for nolint and noinline comments
FOUND=0
COUNT=0

# Search for nolint comments
echo -e "${YELLOW}Searching for 'nolint' comments...${NC}"
echo ""

# Use grep to find all nolint occurrences with line numbers
mapfile -t NOLINT_LINES < <(grep -rn "//go:nolint\|// nolint\|//nolint" --include="*.go" . 2>/dev/null || true)

if [ ${#NOLINT_LINES[@]} -gt 0 ]; then
    FOUND=1
    COUNT=$((COUNT + ${#NOLINT_LINES[@]}))
    for line in "${NOLINT_LINES[@]}"; do
        echo -e "${RED}✗${NC} $line"
    done
    echo ""
else
    echo -e "${GREEN}✓ No nolint comments found${NC}"
    echo ""
fi

# Search for noinline comments
echo -e "${YELLOW}Searching for 'noinline' comments...${NC}"
echo ""

mapfile -t NOINLINE_LINES < <(grep -rn "//go:noinline\|// noinline\|//noinline" --include="*.go" . 2>/dev/null || true)

if [ ${#NOINLINE_LINES[@]} -gt 0 ]; then
    FOUND=1
    COUNT=$((COUNT + ${#NOINLINE_LINES[@]}))
    for line in "${NOINLINE_LINES[@]}"; do
        echo -e "${RED}✗${NC} $line"
    done
    echo ""
else
    echo -e "${GREEN}✓ No noinline comments found${NC}"
    echo ""
fi

# Search for linter skip comments (//lint comments)
echo -e "${YELLOW}Searching for linter skip patterns...${NC}"
echo ""

mapfile -t LINT_SKIP_LINES < <(grep -rn "//lint:\|//lint:" --include="*.go" . 2>/dev/null || true)

if [ ${#LINT_SKIP_LINES[@]} -gt 0 ]; then
    FOUND=1
    COUNT=$((COUNT + ${#LINT_SKIP_LINES[@]}))
    for line in "${LINT_SKIP_LINES[@]}"; do
        echo -e "${RED}✗${NC} $line"
    done
    echo ""
else
    echo -e "${GREEN}✓ No lint skip patterns found${NC}"
    echo ""
fi

# Summary
echo "=========================================="
echo "Summary:"
echo "=========================================="
if [ $FOUND -eq 1 ]; then
    echo -e "${RED}Found $COUNT linter suppression comments${NC}"
    echo ""
    echo "To remove all nolint/noinline comments, run:"
    echo "  ./scripts/remove_nolint.sh"
    exit 1
else
    echo -e "${GREEN}No linter suppression comments found!${NC}"
    exit 0
fi
