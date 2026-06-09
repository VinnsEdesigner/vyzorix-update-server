# Vyzorix CI Validation Scripts

This directory contains automated validation scripts that run as CI guards to ensure code quality and security standards are maintained.

## Overview

These scripts act as **guardians** that run BEFORE linting to validate configuration integrity. If they fail, the CI pipeline stops, preventing bad code from being merged.

## Available Validators

### 1. Go Linter Config Guardian
**File:** `validate-golangci-config.sh`

**Purpose:** Ensures critical Go linters remain enabled and the `.golangci.yml` is not weakened.

**Checks:**
-  8 critical linters are enabled (errcheck, unused, staticcheck, govet, gosec, bodyclose, revive, gofmt)
-  Critical linters are NOT commented out
- At least 20 linters are enabled

**What it prevents:**
```bash
#  THIS WILL FAIL CI:
# - errcheck  # Let me just disable this
```

### 2. ESLint Config Guardian
**File:** `validate-eslint-config.sh`

**Purpose:** Ensures critical ESLint rules remain configured and required packages are installed.

**Checks:**
- Critical rules are configured (no-unused-vars, react-hooks/exhaustive-deps, no-console)
- Required packages are installed (eslint, react-hooks, import)
- 'lint' script is defined in package.json

### 3. Security Validator
**File:** `validate-security.sh`

**Purpose:** Scans for common security vulnerabilities and hardcoded secrets.

**Checks:**
-  Hardcoded passwords, API keys, tokens in code
-  SQL injection risks (fmt.Sprintf with SQL)
-  eval() usage in frontend code
-  innerHTML assignments (XSS risk)
-  Security TODO/FIXME comments
-  console.log in production code

### 4. Dependency Auditor
**File:** `validate-dependencies.sh`

**Purpose:** Ensures dependencies are properly maintained and audited.

**Checks:**
-  Unused Go dependencies
-  Vulnerable dependencies (via govulncheck)
-  Outdated npm packages
-  Deprecated packages (moment, lodash, classnames)
- Required npm scripts are defined

### 5. Git Convention Validator
**File:** `validate-git-conventions.sh`

**Purpose:** Validates commit messages and branch names follow conventions.

**Checks:**
- Commit follows conventional commits format
- Branch follows naming conventions (main, develop, feature/*, etc.)

**Format:** `type(scope): description`
- Types: feat, fix, docs, style, refactor, test, chore, perf, ci, build, opt

### 6. TypeScript Strict Validator
**File:** `validate-typescript-strict.sh`

**Purpose:** Ensures strict TypeScript practices are followed.

**Checks:**
-  'any' type usage (should use 'unknown')
-  Non-null assertions (!) usage
-  @ts-ignore/@ts-expect-error comments
- TODO/FIXME comments count
- Strict mode in tsconfig.json

## Running Locally

All validators can be run locally before committing:

```bash
# Make all scripts executable
chmod +x tooling/scripts/*.sh

# Run all validators
./tooling/scripts/validate-golangci-config.sh
./tooling/scripts/validate-eslint-config.sh
./tooling/scripts/validate-security.sh
./tooling/scripts/validate-dependencies.sh
./tooling/scripts/validate-git-conventions.sh
./tooling/scripts/validate-typescript-strict.sh
```

## CI Integration

These validators run as separate CI jobs in `.github/workflows/ci.yml`:

```yaml
linter-config-guardian:     # Validates Go linter config
eslint-config-guardian:     # Validates ESLint config
security-validator:         # Security scan
dependency-auditor:         # Dependency check
typescript-strict-validator: # TypeScript strict check
git-convention-validator:  # Commit/branch validation
```

## Adding New Validators

1. Create script in this directory
2. Make it executable (`chmod +x`)
3. Add it to CI workflow
4. Document in this README

## Exit Codes

- `0` - Validation passed
- `1` - Validation failed (blocks CI)

## Best Practices

1. **Don't disable validators** - They exist to maintain quality
2. **Fix the root cause** - If a validator fails, fix the code, don't disable the check
3. **Run locally first** - Catch issues before pushing
4. **Keep validators updated** - Update rules as codebase evolves
