#!/bin/bash
# =============================================
# GIT CONVENTION VALIDATOR
# Ensures commits follow conventional commits
# =============================================

set -euo pipefail

echo "[GIT CONVENTIONS] Validating git conventions..."

# Conventional commits pattern
COMMIT_PATTERN="^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|opt)\([a-z0-9_-]+\): .+"

# Check if we're in a git repo
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "[WARNING] Not in a git repository - skipping convention check"
    exit 0
fi

# Get the commit message (from argument or last commit)
COMMIT_MSG="${1:-$(git log -1 --format=%B 2>/dev/null | head -1)}"

if [[ -z "$COMMIT_MSG" ]]; then
    echo "[WARNING] No commit message found - skipping"
    exit 0
fi

echo "Checking commit message: $COMMIT_MSG"

# Check conventional commit format
if [[ "$COMMIT_MSG" =~ $COMMIT_PATTERN ]]; then
    echo "   [OK] Commit follows conventional commits format"
else
    echo "   [WARNING] Commit doesn't follow conventional commits format"
    echo "   Expected: type(scope): description"
    echo "   Types: feat, fix, docs, style, refactor, test, chore, perf, ci, build, opt"
    echo "   Example: feat(auth): add OAuth2 support"
    echo ""
    echo "   This is a warning, not an error. CI will still pass."
fi

# Check for merge commits
if [[ "$COMMIT_MSG" == merge* ]] || [[ "$COMMIT_MSG" == Merge* ]]; then
    echo "   [INFO] Merge commit detected - convention check skipped"
fi

# Check for revert commits
if [[ "$COMMIT_MSG" == revert* ]]; then
    echo "   [INFO] Revert commit detected - convention check skipped"
fi

# Check branch name conventions
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BRANCH_PATTERN="^(main|develop|feature/|fix/|hotfix/|release/|bugfix/)"

echo ""
echo "Checking branch name: $CURRENT_BRANCH"

if [[ "$CURRENT_BRANCH" =~ $BRANCH_PATTERN ]]; then
    echo "   [OK] Branch follows naming convention"
else
    echo "   [WARNING] Branch doesn't follow common naming convention"
    echo "   Expected patterns: main, develop, feature/*, fix/*, hotfix/*, release/*"
fi

echo ""
echo "[OK] Git convention check complete!"

exit 0