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

## Go API (`apps/api`) storage layer
- **Dual backend via a single `*storage.SQLite` handle.** `internal/infrastructure/storage/sqlite.go` is backend-agnostic: `Open(cfg)` picks SQLite (local file, `mattn/go-sqlite3` CGO driver) or Turso (remote libSQL over HTTP, pure-Go `tursodatabase/libsql-client-go` driver) from `cfg.Backend` (auto-detected: turso when `TURSO_DB_URL` set). Both drivers are imported side-by-side (`sqlite3` + `libsql`); both register with `database/sql`.
- **One migration registry, two backends.** `runMigrations(db)` runs the same 47-version SQLite-compatible DDL against either backend via the `schema_migrations` ledger. NEVER fork a separate Turso migration path — that was the weakness of the audio-scope-view reference project. libSQL is a superset of SQLite DDL.
- **Repository layer is unchanged across backends.** All repositories take `*sql.DB`; the libSQL driver implements `database/sql` fully, so `Querier`/`WithTx`/`TxManager` work identically. Transactions pull the `*sql.Tx` from context via `transaction.TxFromContext(ctx)` — executing through `s.DB()` inside a `WithTx` body DEADLOCKS the single-writer SQLite pool (conn held by the tx, body waits for a free conn).
- **Config**: `config.Config` carries `DatabaseBackend`, `TursoDatabaseURL`, `TursoAuthToken`, and pool tuning (`DatabaseMaxOpenConns` etc.). `ResolvedDatabaseBackend()` resolves `auto|sqlite|turso`. `validateDatabaseConfig()` enforces that turso requires URL+token and rejects `http://` in production. Wire DI builds the `storage.Config` in `apps/api/internal/api/wire/providers.go::buildStorageConfig` and `ProvideSQLite(cfg, logger)`.
- **Production hardening on the Turso path**: configurable pool (default 16/8 open/idle, 30m lifetime, 5m idle), `RequestTimeout`-bounded pings, a background health-check goroutine (stopped by `Close()`), `Info()` metadata for `/health` that redacts the auth token from the URL. Auth token is appended to the libSQL DSN as `?authToken=` and never logged.
- **Tests**: `internal/infrastructure/storage/sqlite_test.go` covers local file open+migrate, tx commit/rollback, nil-config and missing-credential guard, and a live `TestOpenTurso_Remote` (skipped unless `TURSO_DB_URL` + a token env var are present; prefers `TURSO_VYZOR_SCOPE_DB_TOKEN` then `TURSO_AUTH_TOKEN`). Run Go tests with `CGO_ENABLED=1` (sqlite3 is CGO).
- **Docker**: `apps/api/Dockerfile` builds with `CGO_ENABLED=1`. The libSQL pure-Go driver coexists fine.

## Live-testing findings & fixes (dual-storage Turso validation)
Bugs surfaced by running the full request flow against Turso cloud and fixed. All were schema/context bugs that local SQLite either masked or would also hit; Turso's stricter remote-driver behavior exposed them.

- **Migration v48 `add_devices_model_column`** (`048_device_model.go`): the domain `Device`, `deviceColumns`, and the device INSERT/UPDATE all reference a `model` column that no prior migration created. Fresh-schema SELECT failed with "no such column: model" → 500 on `GET /v1/device/status`. Idempotent ALTER via `PRAGMA table_info`.
- **Migration v49 `relax_devices_command_secret_null`** (`049_relax_command_secret_null.go`): `devices.command_secret TEXT NOT NULL` (legacy plaintext column) but `DeviceRepository.Create` never sets it (only `command_secret_hash` is written). Every device creation failed with "NOT NULL constraint failed: devices.command_secret". Table rebuild (shadow copy + copy + rename) makes the column nullable; preserves all other columns/indexes; idempotent via `PRAGMA table_info`/`foreign_key_list`.
- **Migration v50 `fix_device_settings_fk`** (`050_fix_device_settings_fk.go`): migration 042 created `device_settings` with `FOREIGN KEY (device_imei) REFERENCES devices(imei)`, but `devices` has no `imei` column (IMEEI lives in `devices.id`). The malformed FK blocked ALL device deletion ("foreign key mismatch") and CASCADE never fired. Table rebuild corrects the FK to `devices(id) ON DELETE CASCADE`. Idempotent.
- **`checkDeviceStatus` error sentinel mismatch** (`application/inbox/inbox_service.go`): compared `err != device.ErrNotFound` (domain sentinel) but the injected `DeviceLookup` is `*deviceapplication.Service`, which returns `application.ErrDeviceNotFound` (different variable, same message). `POST /v1/device/inbox` returned 500 instead of proceeding. Fixed to `errors.Is` against both sentinels; added `deviceapp` import.
- **Idempotency record silently dropped on Turso** (`api/middleware/idempotency_key.go::storeIdempotencyRecord`): the async store goroutine used `c.Request.Context()`, which is cancelled the instant the response is written. libSQL (Turso) honors context cancellation → the INSERT aborts → no idempotency record persisted → retries re-execute the handler and hit the domain 409 instead of replaying the cached 201. Fixed: detached `context.WithTimeout(context.Background(), 10s)` for the async write, with error logging. This was Turso-specific in impact (local CGO SQLite file I/O ignores context cancellation).
- **Run order**: migrations are registered in `sqlite.go` `migrations` slice (v1…v50). Each `Apply` fn + `Name` + `Version`. The `schema_migrations` ledger tracks applied versions.

## Running the server against Turso (local dev)
- `cd /tmp/vyzorix-run` (or any dir with the built binary + a `start.sh` that sets `TURSO_DB_URL`/`TURSO_VYZOR_SCOPE_DB_TOKEN` from `keys.env`). `SSR_ENABLE_SSR=false ENABLE_SSR=false ./start.sh` (SSR off avoids the missing `./web` build dir warnings).
- Server on `:3000`. `/healthz` reports `database:"ok"` when the Turso ping succeeds.
- Valid test IMEI: must be 15-digit Luhn. Helper `/tmp/imei.py` generates valid ones (TAC `490154` + serial + Luhn check digit). FCM token validation requires ≥100 printable ASCII chars.
- Probe the live Turso DB directly: small Go programs under `database/sql` + `github.com/tursodatabase/libsql-client-go/libsql`, DSN = `TURSO_DB_URL?authToken=TURSO_VYZOR_SCOPE_DB_TOKEN` (fall back to `TURSO_AUTH_TOKEN`). `PRAGMA table_info`/`foreign_key_list` work over libSQL.
- Build: `cd apps/api && CGO_ENABLED=1 go build -o /tmp/vyzorix-server ./cmd/api`. Tests: `CGO_ENABLED=1 go test ./...` (all green). Vet: `CGO_ENABLED=1 go vet ./...` clean.
