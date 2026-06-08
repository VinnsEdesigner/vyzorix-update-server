# Vyzorix CI Linter Pass Strategy

> **Document Version:** 1.0  
> **Status:** Active  
> **Last Updated:** 2026-06-08  
> **Purpose:** Comprehensive guide to pass all CI linters with aggressive settings

---

## CI STATUS: ✅ ALL PASSING

| Check | Status | Details |
|-------|--------|---------|
| **Go Linting** | ✅ PASS | 57 aggressive linters enabled |
| **Go Tests** | ✅ PASS | All 11 packages pass |
| **TypeScript** | ✅ PASS | Type check clean |
| **ESLint** | ✅ PASS | 12 warnings (UI folder ignored) |

---

## Table of Contents

1. [Overview](#1-overview)
2. [Aggressive Linter Configuration](#2-aggressive-linter-configuration)
3. [Go Linter Rules (golangci-lint)](#3-go-linter-rules-golangci-lint)
4. [TypeScript/ESLint Rules](#4-typescripteslint-rules)
5. [CI Pipeline](#5-ci-pipeline)
6. [Quick Reference](#6-quick-reference)

---

## 1. Overview

### CI Jobs
| Job | Tool | Purpose |
|-----|------|---------|
| `lint` | ESLint + TypeScript | Frontend code quality |
| `test` | Vitest | Frontend unit tests |
| `go-lint` | golangci-lint | Go code quality |
| `go-test` | go test | Go unit tests |

### Current Status
- **Go Linting:** ✅ PASS (57 aggressive linters enabled)
- **TypeScript:** ✅ PASS (type check clean)
- **ESLint:** ✅ PASS (13 warnings in /ui/ folder - ignored)

---

## 2. Aggressive Linter Configuration

### Go: `.golangci.yml` Location
```
apps/api/.golangci.yml
```

### TypeScript: ESLint Config Location
```
apps/web/eslint.config.js
```

---

## 3. Go Linter Rules (golangci-lint)

### 3.1 Enabled Linters (57 total)

#### Error Handling (Critical)
| Linter | Purpose | Fix |
|--------|---------|-----|
| `errcheck` | Unchecked errors | Always check `err` return values |
| `errorlint` | Error wrapping issues | Use `errors.Is`, `errors.As` |
| `err113` | Inefficient error handling | Don't wrap with `fmt.Errorf` inside `errors.Wrap` |
| `errname` | Naming convention | Prefix errors with `Err`, suffix types with `Error` |
| `nilerr` | Return nil when checking err != nil | Return actual error |
| `nilnesserr` | Check err != nil but return nil error | Return meaningful error |
| `nilnil` | Return nil error + non-nil value | Return only nil error or only value |
| `rowserrcheck` | sql.Rows.Err unchecked | Check `rows.Err()` |
| `sqlclosecheck` | SQL resources not closed | Use `defer rows.Close()` |

#### Code Quality
| Linter | Purpose | Fix |
|--------|---------|-----|
| `unused` | Unused code | Remove or use `_` prefix |
| `staticcheck` | Static analysis issues | Follow recommendations |
| `gosimple` | Simplify code | Use idiomatic Go |
| `govet` | Vet suspicious constructs | Fix vet warnings |
| `ineffassign` | Dead assignments | Remove unused assignments |
| `copyloopvar` | Loop variable copy | Use pointer or index |
| `predeclared` | Shadow predeclared identifiers | Rename variables |

#### Style & Formatting
| Linter | Purpose | Fix |
|--------|---------|-----|
| `gofmt` | Format code | Run `gofmt -s` |
| `gofumpt` | Format code (strict) | Run `gofumpt` |
| `goimports` | Format imports | Run `goimports` |
| `godot` | Comments end with period | Add `.` to comments |
| `misspell` | Misspelled words | Fix spelling |
| `revive` | Code style | Follow revive rules |
| `stylecheck` | Code style (strict) | Follow stylecheck rules |
| `gci` | Import order | Group: std, default, local |
| `decorder` | Declaration order | Group: const, var, func |

#### Complexity
| Linter | Setting | Limit |
|--------|---------|-------|
| `gocyclo` | min-complexity | 25 |
| `funlen` | lines | 60 |
| `funlen` | statements | 40 |
| `nestif` | min-complexity | 8 |
| `maintidx` | min-complexity | 70 |

#### Security
| Linter | Purpose | Fix |
|--------|---------|-----|
| `gosec` | Security issues | Fix security vulnerabilities |
| `bodyclose` | HTTP body not closed | Use `defer resp.Body.Close()` |

#### Best Practices
| Linter | Purpose | Fix |
|--------|---------|-----|
| `exhaustive` | Switch exhaustive | Add all enum cases |
| `exhaustruct` | Struct initialization | Initialize all fields |
| `iface` | Interface pollution | Keep interfaces small |
| `importas` | Import aliases | Use consistent aliases |
| `noctx` | HTTP without context | Pass `ctx` to HTTP requests |
| `contextcheck` | Context not passed | Pass context to functions |
| `containedctx` | Struct with context | Use context as parameter |

#### Naming & Conventions
| Linter | Purpose | Fix |
|--------|---------|-----|
| `godox` | TODO/FIXME comments | Use proper format: `// TODO(author):` |
| `goprintffuncname` | Printf-like naming | Name with `f` suffix |
| `errchkjson` | JSON encoding errors | Check error returns |
| `perfsprint` | Slow Sprint usage | Use `strconv` for numbers |

#### Testing
| Linter | Setting | Fix |
|--------|---------|-------|
| `paralleltest` | Missing t.Parallel() | Add `t.Parallel()` to tests |
| `testableexamples` | Examples need output | Add expected output |
| `testpackage` | Test in same package | Use `_test` package |
| `testifylint` | Testify usage | Use proper assertions |

#### Performance
| Linter | Purpose | Fix |
|--------|---------|-----|
| `prealloc` | Preallocate slices | Use `make(len)` |
| `mirror` | Wrong mirror usage | Use `bytes.TrimSuffix` |
| `nakedret` | Naked returns | Use named returns |

#### Imports
| Linter | Purpose | Fix |
|--------|---------|-----|
| `depguard` | Blocked imports | Remove disallowed imports |
| `gomoddirectives` | Replace/retract directives | Remove `replace` directives |
| `gomodguard` | Module allow/block | Update allow list |

#### Logging
| Linter | Purpose | Fix |
|--------|---------|-----|
| `loggercheck` | Logger key-value pairs | Use consistent keys |
| `sloglint` | log/slog style | Use structured logging |

---

### 3.2 Common Fix Patterns

#### Error Handling
```go
// ❌ BAD - unchecked error
_ = os.ReadFile("test.txt")

// ✅ GOOD - check error
data, err := os.ReadFile("test.txt")
if err != nil {
    return fmt.Errorf("read file: %w", err)
}
```

#### Context Propagation
```go
// ❌ BAD - no context
resp, err := http.Get(url)

// ✅ GOOD - with context
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
```

#### SQL Resource Cleanup
```go
// ❌ BAD - rows not checked
rows, err := db.Query("SELECT * FROM users")
defer rows.Close()

// ✅ GOOD - proper cleanup
rows, err := db.Query("SELECT * FROM users")
if err != nil {
    return err
}
defer rows.Close()
for rows.Next() {
    // process row
}
if err := rows.Err(); err != nil {
    return err
}
```

#### Import Ordering
```go
// ✅ GOOD - grouped imports
import (
    "context"
    "fmt"

    "github.com/gin-gonic/gin"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
    "github.com/VinnsEdesigner/vyzorix/apps/api/pkg/config"
)
```

---

### 3.3 Running Linters Locally

```bash
# Setup
export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin

# Run all linters
cd apps/api
golangci-lint run ./...

# Run with verbose output
golangci-lint run -v ./...

# Run specific linter
golangci-lint run --enable=errcheck ./...

# Auto-fix where possible
golangci-lint run --fix ./...
```

---

## 4. TypeScript/ESLint Rules

### 4.1 Configuration Location
```
apps/web/eslint.config.js
```

### 4.2 Key Rules

#### React Hooks (Critical)
| Rule | Level | Fix |
|------|-------|-----|
| `react-hooks/exhaustive-deps` | error | Add all dependencies to useEffect/useCallback |

#### Unused Code
| Rule | Level | Fix |
|------|-------|-----|
| `@typescript-eslint/no-unused-vars` | error | Remove or prefix with `_` |
| `unused-imports/no-unused-imports` | error | Remove unused imports |

#### Best Practices
| Rule | Level | Fix |
|------|-------|-----|
| `no-console` | warn | Remove or use `console.warn/error` |
| `no-alert` | error | Remove `alert()` calls |
| `no-debugger` | error | Remove `debugger` statements |
| `no-eval` | error | Remove `eval()` calls |
| `no-implicit-coercion` | error | Use `Boolean()` not `!!` |

#### Stylistic
| Rule | Level | Fix |
|------|-------|-----|
| `max-len` | error | Max 120 chars per line |
| `comma-dangle` | error | Trailing commas in multiline |
| `semi` | error | Always end with semicolon |
| `one-var` | error | Never use `var a, b` |

#### Import Rules
| Rule | Level | Fix |
|------|-------|-----|
| `import/order` | error | Group: external, internal, relative |
| `import/no-duplicates` | error | No duplicate imports |
| `import/no-cycle` | error | No circular dependencies |

### 4.3 UI Components Folder Exception

The `/apps/web/src/components/ui/` folder has relaxed rules:
- `react-refresh/only-export-components` is set to **warn** (not error)
- This is intentional because UI libraries export utility functions alongside components

### 4.4 Running ESLint Locally

```bash
# Run all linting
pnpm lint

# Run typecheck
pnpm typecheck

# Run specific package
pnpm --filter @vyzorix/web lint
```

---

## 5. CI Pipeline

### 5.1 GitHub Actions Workflow

Located at: `.github/workflows/ci.yml`

```yaml
jobs:
  lint:        # pnpm lint
  test:        # pnpm test
  go-lint:     # golangci-lint + go test
```

### 5.2 Local CI Simulation

```bash
# Install dependencies
pnpm install --no-frozen-lockfile

# Run all checks
pnpm lint
pnpm typecheck
pnpm test

# Go checks
cd apps/api
export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin
go mod tidy
golangci-lint run ./...
go test ./...
```

---

## 6. Quick Reference

### Go: Files to Never Commit With Issues
| Pattern | Issue |
|---------|-------|
| `_, _ =` | Unchecked error |
| `http.Get(` | No context |
| `fmt.Printf(` | Use log/slog |
| `// TODO` | Use `// TODO(author):` |
| `errors.New(fmt.Sprintf(` | Use `fmt.Errorf` |

### TypeScript: Files to Never Commit With Issues
| Pattern | Issue |
|---------|-------|
| `// eslint-disable` | Avoid unless necessary |
| `any` | Use `unknown` |
| `!important` | CSS !important abuse |
| `console.log` | Use console.warn/error |

### Pre-commit Checklist
```bash
# Go
cd apps/api
golangci-lint run ./...   # Must pass
go test ./...             # Must pass

# Frontend
cd apps/web
pnpm lint                 # Must pass
pnpm typecheck            # Must pass
```

---

## Document Changelog

| Date | Change |
|------|--------|
| 2026-06-08 | Initial creation - 57 Go linters enabled, all passing |

---

**End of Document**