# Frontend Lint Errors Fix Guide

This document provides a step-by-step guide to fix all remaining lint errors in the frontend codebase (src/ directory).

## Error Categories

### 1. func-style (Expected a function expression)
**Rule:** Functions should be defined as `const x = () => {}` instead of `function x() {}`

**Files affected:**
- src/components/app-sidebar.tsx (line 44)
- src/components/connection-badge.tsx (line 7)
- src/components/layout/footer.tsx (line 5)
- src/components/loading/page-skeleton.tsx (lines 5, 18)
- src/components/logs/log-console.tsx (line 8)
- src/components/logs/log-dock.tsx (line 9)
- src/components/status-badge.tsx (line 7)
- src/components/ui/badge.tsx (line 29)
- src/components/ui/calendar.tsx (lines 11, 140)
- src/components/ui/carousel.tsx (line 31)
- src/components/ui/chart.tsx (lines 26, 301)
- src/components/ui/menubar.tsx (lines 8, 14, 20, 26, 32)
- src/components/ui/sidebar.tsx (line 40)
- src/components/ui/skeleton.tsx (line 5)
- src/components/ui/spinner.tsx (line 6)
- src/hooks/use-auth.ts (lines 29, 64, 116, 139)
- src/hooks/use-device-stream.ts (line 18)
- src/hooks/use-logs.ts (line 12)
- src/hooks/use-mobile.tsx (line 5)
- src/hooks/use-server-health.ts (line 5)
- src/lib/device-stream-context.tsx (lines 13, 19)
- src/lib/settings.test.ts (lines 83, 139, 143, 208, 212, 389)
- src/lib/vyzorix-api.ts (lines 64, 68, 72, 90, 102, 121, 136, 141, 152, 176)
- src/lib/vyzorix-auth.test.ts (line 60)
- src/lib/vyzorix-auth.ts (lines 65, 73, 81, 90, 99, 109, 125, 143, 162, 184, 199, 212, 236, 253, 265, 285, 299, 316, 322, 348)
- src/lib/vyzorix-config.tsx (lines 71, 98, 170, 176)
- src/routes/__root.tsx (lines 16, 38, 105, 124)
- src/routes/_app.alerts.tsx (lines 41, 96, 221)
- src/routes/_app.diagnostics.tsx (lines 35, 145, 154)
- src/routes/_app.logs.tsx (line 12)
- src/routes/_app.settings.advanced.tsx (line 17)
- src/routes/_app.settings.appearance.tsx (lines 16, 23, 67)
- src/routes/_app.settings.connection.tsx (lines 21, 31, 241)
- src/routes/_app.settings.index.tsx (lines 60, 120)
- src/routes/_app.settings.notifications.tsx (lines 16, 94)
- src/routes/_app.settings.operator.tsx (lines 25, 233)
- src/routes/_app.settings.thresholds.tsx (lines 16, 125)
- src/routes/_app.settings.tsx (line 19)
- src/routes/_app.tsx (lines 29, 39)
- src/routes/_app.updates.tsx (lines 20, 192)
- src/routes/auth.callback.tsx (line 27)
- src/routes/forgot-password.tsx (line 20)
- src/routes/login.tsx (line 24)
- src/routes/reset-password.tsx (line 20)
- src/routes/verify-email.tsx (line 20)
- src/server.ts (line 21)

**Fix process:**
1. For each function definition like `function FunctionName() { ... }`, convert to `const FunctionName = () => { ... }`
2. For exported functions, change `export function FunctionName() { ... }` to `export const FunctionName = () => { ... }`

### 2. @typescript-eslint/explicit-function-return-type
**Rule:** All functions must have an explicit return type annotation

**Fix process:**
1. Identify functions missing return types
2. Add return type annotation, e.g., `const fn = (): string => { ... }`

**Key locations:**
- src/hooks/use-server-health.ts (line 5): Change `export function useServerHealth(` to `export function useServerHealth(): UseQueryResult<{ ok: boolean }, Error> {`
- src/hooks/use-auth.ts (lines 43, 150)
- src/hooks/use-device-stream.ts (lines 38, 98)
- src/hooks/use-mobile.tsx (line 10)
- src/lib/config.server.ts (line 19)
- src/lib/vyzorix-config.tsx (lines 122, 158, 159, 160, 161)
- src/router.tsx (line 6)
- src/routes/_app.settings.connection.tsx (lines 48, 66, 87)
- src/routes/_app.updates.tsx (line 39)
- All UI components in src/components/ui/

### 3. @typescript-eslint/prefer-nullish-coalescing
**Rule:** Use `??` instead of `||` for nullish coalescing

**Fix process:**
1. Find patterns like `x || y` where `x` could be `null` or `undefined`
2. Replace with `x ?? y`

**Key locations:**
- src/components/ui/chart.tsx (lines 44, 66, 81, 133, 172, 270)
- src/components/ui/progress.tsx (line 19)
- src/components/ui/toggle-group.tsx (lines 43, 44)
- src/integrations/supabase/client.ts (lines 9, 11)
- src/integrations/supabase/client.server.ts (line 39)
- src/server.ts (line 11)

### 4. no-nested-ternary
**Rule:** Do not nest ternary expressions

**Fix process:**
1. Refactor nested ternary into if/else or separate variables
2. Example: `a ? b ? c : d : e` should become:
   ```typescript
   const result = a && b ? c : a ? d : e;
   // Or better:
   let result;
   if (a && b) result = c;
   else if (a) result = d;
   else result = e;
   ```

**Key locations:**
- src/routes/_app.alerts.tsx (line 198)
- src/routes/_app.settings.connection.tsx (lines 148, 152)
- src/routes/_app.updates.tsx (line 105)

### 5. no-return-await
**Rule:** Remove redundant `await` on return values

**Location:**
- src/hooks/use-auth.ts (line 125)

**Fix:** Change `return await someAsyncFunction()` to `return someAsyncFunction()`

### 6. require-await
**Rule:** Async functions must use `await`

**Location:**
- src/lib/api/example.functions.ts (line 16)

**Fix:** Either add `await` inside the async function or remove `async` keyword if not needed

### 7. prettier/prettier (formatting issues)
**Rule:** Code must be properly formatted

**Fix process:**
1. Run `npm run lint -- --fix` to auto-fix formatting issues
2. Manual fixes needed for:
   - src/components/ui/drawer.tsx (line 9): Delete extra newline
   - src/components/ui/form.tsx (line 32): Delete extra newline
   - src/components/ui/pagination.tsx (line 52): Delete extra newline
   - src/components/ui/resizable.tsx (line 21): Delete extra newline
   - src/hooks/use-server-health.ts (line 5): Put parameters on single line
   - src/routes/_app.settings.connection.tsx (lines 147, 151): Put ternary on single line

## Quick Fix Commands

```bash
# Auto-fix all fixable errors (prettier and some others)
cd src && npm run lint -- --fix

# Check remaining errors
npm run lint
```

## File-by-File Fix Instructions

### 1. src/hooks/use-server-health.ts
Line 5: Change from multi-line to single line:
```typescript
// FROM:
export function useServerHealth(
  serverUrl: string,
): UseQueryResult<{ ok: boolean }, Error> {

// TO:
export function useServerHealth(serverUrl: string): UseQueryResult<{ ok: boolean }, Error> {
```

### 2. src/routes/_app.settings.connection.tsx
Lines 147-152: Fix nested ternary and formatting:
```typescript
// FROM:
health.data?.ok ? "default" : health.isError ? "destructive" : "secondary"

// TO:
health.data?.ok ? "default" : health.isError ? "destructive" : "secondary"
```
(Use single line, or refactor to if/else)

### 3. src/hooks/use-auth.ts
Line 125: Remove redundant await:
```typescript
// FROM:
return await someFunction();

// TO:
return someFunction();
```

### 4. src/lib/api/example.functions.ts
Line 16: Either add await or remove async based on function logic.

### 5. src/components/ui/drawer.tsx
Line 9: Delete extra blank line after imports.

### 6. src/components/ui/form.tsx
Line 32: Delete extra blank line.

### 7. src/components/ui/pagination.tsx
Line 52: Delete extra blank line.

### 8. src/components/ui/resizable.tsx
Line 21: Delete extra blank line.

## ESLint-Disable Comment Removal Process

The following process was used to remove all `eslint-disable` comments from the frontend codebase:

### Process Steps

1. **Identify all files with eslint-disable comments:**
   ```bash
   grep -r "eslint-disable" src/ --include="*.tsx" --include="*.ts" -l
   ```

2. **Remove all single-line eslint-disable comments:**
   ```bash
   find src -name "*.tsx" -o -name "*.ts" | xargs sed -i '/\/\/ eslint-disable/d'
   ```

3. **Handle block-style eslint-disable comments (if any):**
   ```bash
   find src -name "*.tsx" -o -name "*.ts" | xargs sed -i '/\/\* eslint-disable \*\//d'
   ```

4. **Verify removal:**
   ```bash
   grep -r "eslint-disable" src/ --include="*.tsx" --include="*.ts"
   ```
   (Should return no results)

### Files That Had ESLint-Disable Comments Removed

The following files had eslint-disable comments removed:
- src/routes/_app.diagnostics.tsx
- src/routes/_app.logs.tsx
- src/routes/verify-email.tsx
- src/routes/_app.settings.thresholds.tsx
- src/routes/forgot-password.tsx
- src/routes/reset-password.tsx
- src/routes/_app.settings.index.tsx
- src/routes/auth.callback.tsx
- src/routes/__root.tsx
- src/routes/_app.alerts.tsx
- src/routes/_app.tsx
- src/routes/_app.device.tsx
- src/routes/login.tsx
- src/routes/_app.settings.advanced.tsx
- src/routes/_app.settings.tsx
- src/routes/_app.updates.tsx
- src/routes/_app.settings.notifications.tsx
- src/routes/_app.settings.connection.tsx
- src/routes/_app.settings.appearance.tsx
- src/routes/_app.settings.operator.tsx
- src/routes/_app.dashboard.tsx
- src/server.ts
- src/components/ui/chart.tsx
- src/components/ui/alert-dialog.tsx
- src/components/ui/breadcrumb.tsx
- src/components/ui/sheet.tsx
- src/components/ui/carousel.tsx
- src/components/ui/form.tsx
- src/components/ui/spinner.tsx
- src/components/ui/resizable.tsx
- src/components/ui/dialog.tsx
- src/components/ui/badge.tsx
- src/components/ui/skeleton.tsx
- src/components/ui/sidebar.tsx
- src/components/ui/toggle-group.tsx
- src/components/ui/menubar.tsx
- src/components/ui/command.tsx
- src/components/ui/dropdown-menu.tsx
- src/components/ui/drawer.tsx
- src/components/ui/progress.tsx
- src/components/ui/sonner.tsx
- src/components/ui/calendar.tsx
- src/components/ui/context-menu.tsx
- src/components/ui/pagination.tsx
- src/components/layout/footer.tsx
- src/components/app-sidebar.tsx
- src/components/loading/page-skeleton.tsx
- src/components/status-badge.tsx
- src/components/logs/log-dock.tsx
- src/components/logs/log-console.tsx
- src/routeTree.gen.ts
- src/lib/api/example.functions.ts
- src/lib/settings.test.ts
- src/lib/vyzorix-auth.test.ts
- src/lib/vyzorix-auth.ts
- src/lib/device-stream-context.tsx
- src/lib/vyzorix-api.ts
- src/lib/config.server.ts
- src/lib/vyzorix-config.tsx
- src/hooks/use-server-health.ts
- src/hooks/use-mobile.tsx
- src/hooks/use-logs.ts
- src/hooks/use-auth.ts
- src/hooks/use-device-stream.ts
- src/router.tsx
- src/integrations/supabase/client.server.ts
- src/integrations/supabase/client.ts

### Why Remove ESLint-Disable Comments?

1. **Code Quality**: ESLint-disable comments hide legitimate code issues
2. **Maintainability**: Over time, these comments accumulate and mask problems
3. **Technical Debt**: Each disable is a reminder that something needs proper fixing
4. **Best Practice**: Modern linting should guide code quality, not be circumvented

### Alternative Approach

Instead of blanket removal, you could:
1. Evaluate each disable individually
2. Fix the underlying issue the disable was masking
3. Then remove the disable comment once the fix is verified

## Notes

- The `func-style` errors are numerous (80+). They require converting function declarations to arrow functions.
- The `explicit-function-return-type` errors also require adding return types to all functions.
- Consider using a codemod or ESLint's `--fix` option where applicable.
- Test files (settings.test.ts, vyzorix-auth.test.ts) also have func-style errors that should be fixed.
- The route files in src/routes/ have the most func-style errors and should be addressed in groups.

## Automated Approach

To fix all `func-style` and `explicit-function-return-type` errors programmatically, consider:

1. Using `@typescript-eslint/eslint-plugin` with `func-style: ["error", "expression"]`
2. Running a codemod like `ts-codemod` to convert function declarations to arrow functions
3. Or manually fixing each file following the list above