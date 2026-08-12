# AGENTS.md — vyzorix-update-server

## Repository layout
- `apps/api/` — Go server (go.mod). Does NOT consume the TS SDK.
- `apps/VyzoriX_web/` — Vite + TanStack Start web app. Consumes `@vyzorix/api-client` (workspace symlink) for connectivity only.
- `apps/VyzoriX_mobile/` — React Native app.
- `packages/API_Client/` — `@vyzorix/api-client` TS SDK (source-only, consumed via workspace, no published build).
- `packages/config/`, `packages/ui/` — shared workspace packages.

## API_Client (`packages/API_Client`) conventions
- **TypeScript**: `npx tsc --noEmit --ignoreDeprecations 6.0` (baseUrl is deprecated in TS 7.0; the `--ignoreDeprecations 6.0` flag is REQUIRED or tsc errors). Internal path alias `@/*` → `./src/*` (defined in the package's own tsconfig, NOT in the web app's tsconfig).
- **Entry points (web/node separation)**:
  - `@vyzorix/api-client` (root, `src/index.ts`) — universal/browser-safe surface: domain types/mappers/validators, config, REST client (axios), connectivity, batching, GraphQL. Deliberately does NOT export `crypto`/`security`/`websocket`/`device` so `node:crypto`/`Buffer` never enter a browser bundle.
  - `@vyzorix/api-client/node` (`src/node.ts`) — Node-only: re-exports root + `crypto` (node:crypto HMAC/AES-256-GCM/HTTP signing), `security` (SSL pinning via node:crypto/tls), `websocket` (WebSocket client), `device` (device client; signs requests via node:crypto).
  - `package.json` `exports` uses `browser`/`node`/`import`/`default` conditions.
- **Shared domain primitives** live in `src/domain/_shared/` (`DeviceStatus`, `Pagination`, `MetricThreshold`, `RawPagination`, `paginationFromRaw`, `validateIMEI`, `ValidationResult`, ...). Each context entity file re-exports the shared types it needs via `import type` + `export type`. Do NOT duplicate type definitions across contexts.
- **Singleton state**: mutable per-session state is consolidated into single state objects (`clientState` in `rest-client.ts`, `authContextState` in `auth-context.ts`, `graphqlState` in `graphql-client.ts`, `batcherState` in `request-batcher.ts`). `resetClientState()` / `auth.clear()` / `resetBatchers()` reset for tests. Isolated single lazy-singleton instance vars behind `getX()` factories (e.g. `defaultMonitor`, `deviceClient`, `sslPinningInstance`) are the accepted pattern.
- **Naming to avoid context collisions**: `diagnosticTelemetryFromRaw` (not `telemetryFromRaw`), `timelineEventTypeLabel` (not `getEventTypeLabel`), `RawRegisteredDevice` (not `RawDevice`) in registration context.
- **Build**: `pnpm build` runs `tsup src/index.ts src/node.ts --format cjs,esm`.
- **React peer**: `peerDependencies.react` is `>=18.2.0` (covers both the RN 18 line and the web app's React 19). Do NOT narrow it back to `^18.2.0` — that excludes the web consumer's React 19 and forces pnpm to install a duplicate react 18 into the SDK context (lockfile drift, see git history).
- **Install**: `pnpm install` at repo root (corepack `packageManager: pnpm@11.5.2`). `pnpm install --frozen-lockfile` is the CI mode and is currently green for the SDK's own peer deps; remaining `pnpm peers check` warnings (eslint/typescript/tsconfck/vite/react-native-reanimated) are pre-existing dev-tooling drift in other packages, not SDK-caused.

## Working in this repo
- pnpm workspace (`pnpm-workspace.yaml`); root `package.json` workspaces `apps/*` + `packages/*`.
- When typechecking the web app, pre-existing `@/domain/...` resolution errors from API_Client source are expected (web tsconfig `@/*` → its own `./src/*`, not API_Client's). These are NOT caused by SDK changes.
