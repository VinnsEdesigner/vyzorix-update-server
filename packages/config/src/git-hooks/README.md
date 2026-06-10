# @vyzorix/config/git-hooks - Git Hooks Configuration
# Husky + lint-staged for pre-commit validation

## Setup Instructions

1. Install dependencies:
\`\`\`bash
pnpm add -D husky lint-staged
\`\`\`

2. Initialize Husky:
\`\`\`bash
npx husky init
\`\`\`

3. Update package.json scripts:
\`\`\`json
{
  "scripts": {
    "prepare": "husky install",
    "lint-staged": "lint-staged"
  },
  "lint-staged": {
    "*.{ts,tsx,js,jsx}": [
      "eslint --fix",
      "prettier --write"
    ],
    "*.{json,md,css}": [
      "prettier --write"
    ],
    "*.go": [
      "gofmt -s -w"
    ]
  }
}
\`\`\`

4. Create the pre-commit hook:
\`\`\`bash
npx husky add .husky/pre-commit "pnpm lint-staged"
\`\`\`

## Available Hooks

### pre-commit
- Runs ESLint on staged files
- Formats with Prettier
- Type checks with TypeScript
- Runs tests on staged files

### commit-msg
- Validates commit message format
- Enforces Conventional Commits

### pre-push
- Runs full test suite
- Builds the project

## Commit Message Format

\`\`\`
<type>(<scope>): <subject>

<body>

<footer>
\`\`\`

### Types
- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Formatting, missing semi colons, etc
- **refactor**: Code refactoring
- **test**: Adding tests
- **chore**: Maintenance tasks

### Examples
\`\`\`bash
git commit -m "feat(auth): add Google OAuth support"
git commit -m "fix(dashboard): resolve device list pagination"
git commit -m "docs(api): update endpoint documentation"
\`\`\`

## Troubleshooting

### Hooks not running
\`\`\`bash
# Check if husky is installed
cat .husky/pre-commit

# Reinstall hooks
pnpm prepare
\`\`\`

### Skip hooks temporarily
\`\`\`bash
git commit --no-verify -m "WIP: temporary commit"
\`\`\`