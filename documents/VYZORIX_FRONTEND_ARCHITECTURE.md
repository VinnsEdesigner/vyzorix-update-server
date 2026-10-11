# Vyzorix Frontend Architecture: Codegen & Data Layer Modernization

## Context

Vyzorix is a device management platform with a Go backend (Gin + swaggo) and a
TypeScript frontend (Vite + React + TanStack Query). The current data layer pipeline is:

```
Go handler annotations (swaggo)
  → swag init → swagger.json (Swagger 2.0)
  → build_openapi3.py → openapi3.json (OpenAPI 3.0)
  → orval → packages/API_Client/src/generated/ (typed axios functions)
  → build_zod.py → vyzorixUpdateServerAPI.zod.ts (zod schemas)
  → domain/ modules (re-exports + hand-rolled business rules)
  → apps/VyzoriX_web/src/hooks/ (manual TanStack Query wrappers)
```

This document outlines six architectural improvements, inspired by patterns
observed in production observability platforms, that Vyzorix can adopt — with
naming conventions unique to Vyzorix's device-management domain.

---

## 1. Vyzorix Query SDK — Auto-Generated Typed Hooks

### Problem

The current pipeline generates typed axios functions via orval, but every hook
in `apps/VyzoriX_web/src/hooks/` was hand-written boilerplate:

```typescript
// 40+ files like this, each ~20 lines of wrapper:
export function useDevices(params?: DeviceParams) {
  return useQuery({
    queryKey: queryKeys.devices(params),
    queryFn: () => getDevices().getDevices(),
  });
}
```

Cache invalidation was manual:
```typescript
onSuccess: () => queryClient.invalidateQueries({ queryKey: ['devices'] }),
```

If a handler adds a new field, the hook doesn't know about it. If a mutation
should refetch related queries, the developer must remember to wire it.

### Proposal

Generate typed query/mutation hooks directly from `openapi3.json` and drive
cache invalidation from the spec, so mutations refetch related queries with no
hand-written `invalidateQueries()` calls.

The swaggo `@Tags` annotations already partition the API into tags:
`@Tags devices`, `@Tags updates`, `@Tags auth`, etc. These map directly to
cache tag types for automatic invalidation:

```go
// CreateDevice handler
// @Tags devices
// @Success 201 {object} openapi.DeviceListItem
```

Implemented via orval's React Query client with `override.query`:

```javascript
// orval.config.js — vyzorixRQ output
query: {
  version: 5,
  useInvalidate: true,          // emit invalidateGet* helpers
  shouldExportQueryKey: true,   // emit getGet*QueryKey (shared key scheme)
  queryKey: {                   // org-scope every generated key
    path: 'apps/VyzoriX_web/src/generated-rq/query-key-mutator.ts',
    name: 'organizationScopedQueryKey',
    default: true,
  },
  mutationInvalidates: [
    { onMutations: ['postDeviceConfirm', 'deleteDevicesImei', /* … */],
      invalidates: [{ query: 'getDevices' }, { query: 'getDeviceCount' }, /* … */] },
    // … one entry per domain tag
  ],
},
```

Generates hooks whose mutations invalidate the org-scoped keys of their
related queries:

```typescript
// apps/VyzoriX_web/src/generated-rq/devices/device-management.ts
export const getGetDevicesQueryKey = (params?: GetDevicesParams) =>
  organizationScopedQueryKey({ params }, { url: `/devices` });

export const useDeleteDevicesImei = (…) => {
  const onSuccess = (…) => {
    queryClient.invalidateQueries({ queryKey: getGetDevicesQueryKey() });
    queryClient.invalidateQueries({ queryKey: getGetDeviceCountQueryKey() });
    // …
  };
};
```

The domain wrapper hooks (`src/hooks/*`) become thin re-exports — they keep
only business logic (org gating, wire→domain normalization, GraphQL fallback,
polling state machines) and no longer call `invalidateQueries` for the
mutation-driven cases. WebSocket-driven invalidation (external events that no
mutation produces) stays, but now invalidates the same generated org-scoped
keys.

### What gets deleted (~30 pure API wrapper hooks)

Vyzorix has 58 hook files total. They fall into three categories:

**Pure API wrappers (become thin re-exports)** — files like
`use-devices.ts`, `use-alerts.ts`, `use-versions.ts`, `use-sessions.ts`,
`use-settings.ts`, `use-contact-points.ts`, `use-organizations.ts`. These keep
their names and signatures (so route/page consumers don't change) but drop
their hand-written `queryKey` overrides and `invalidateQueries` calls; the
generated hooks own the keys and the invalidation.

- Manual `queryClient.invalidateQueries` calls (mutation-driven cases)
- Manual `queryKey` definitions in `lib/query-keys.ts` for generated-backed
  queries — the org-scoped `getGet*QueryKey` helpers are now canonical, and
  `lib/query-keys.ts` is deleted (no app code referenced it after the migration)

### What stays (~28 hooks)

**~15 orchestration hooks (simplified)** — files like
`use-commands.ts` (command polling with state machine callbacks),
`use-realtime.ts` (WebSocket subscriptions), `_graphql-fallback.ts`
files (REST fails → try GraphQL), `use-login.ts` (MFA challenge flow),
`use-auth-session.ts` (auth context). These have real logic beyond
"call API X" — they stay, but are simpler because cache invalidation
is automatic (or uses the generated `invalidateGet*` helpers where a GraphQL
fallback bypasses the generated mutation) and they call generated hooks
directly.

**~13 non-API hooks (kept, unchanged)** — files like
`use-vyzor-analytics.ts`, `use-connectivity.ts`, `use-debounce.ts`,
`use-pagination.ts`, `use-time-range.ts`, `use-vyzor-language.ts`,
`use-vyzor-error-recovery.ts`. These don't call the API at all — they're
UI utilities that are unaffected by the codegen change.

### What else stays

- `domain/` layer (zod schemas + business validators)
- `restClient` transport (wired into orval via the `rest-bridge` mutator)
- `openapi/schemas.go` (Go wire types)
- `build_openapi3.py` (spec normalization)
- `build_zod.py` (zod generation)

### Limitation: parameterized queries

orval's `query.mutationInvalidates` can only reference queries whose generated
`getGet*QueryKey()` takes **no runtime path params**. For a parameterized query
(e.g. `getOrganizationsIdMembers(id)`, `getConfigVersionsResource(resource)`,
`getCommandDispatchIdStatus(dispatchId)`), orval emits the key helper with zero
arguments, so the generated invalidation call fails to compile
(`TS2554: Expected N arguments, but got 0`). Those operations are therefore
intentionally left out of `mutationInvalidates`:

- **commands** (`postDeviceImeiCommand`, `deleteCommandDispatchId`,
  `postCommandDispatchIdRetry`, `postDevicesImeiCommandConfirm`) — invalidated by
  hand in `hooks/commands/use-commands.ts` using the generated
  `getGet*QueryKey` / `invalidateGet*` helpers;
- **organization members** (`deleteOrganizationsIdMembersMemberId` and the member
  `role` / `suspend` / `reinstate` / `transfer` mutations) — invalidated by the
  realtime WebSocket member-event handler;
- **config-version restore** (`postConfigVersionsResourceVersionRestore`) — no UI
  consumer yet.

When adding UI for one of these, invalidate by hand with the generated key
helper; do not add it to `mutationInvalidates`.

### Naming

- Package: `@vyzorix/api-client` (existing)
- Generated hooks: `apps/VyzoriX_web/src/generated-rq/` (tags-split)
- Hook exports: `useGetDevices`, `usePushUpdate`, etc.
- Query keys: `getGetDevicesQueryKey`, … (org-scoped via
  `organizationScopedQueryKey`)
- Invalidation helpers: `invalidateGetDevices`, …
- Domain wrappers: `apps/VyzoriX_web/src/hooks/*` (unchanged names)

---

## 2. Vyzorix Schema Definitions — CUE-Based Config Types

### Problem

Device settings, thresholds, notification settings, and update push configs
were defined as Go structs in `openapi/schemas.go`, hand-mirroring the `gin.H`
keys that handlers emit. If a handler changed a JSON key, the schema silently
drifted — the generated TypeScript type said one thing, the server sent another.
The pre-commit swagger gate checked annotations against `swagger.json`, but not
that `openapi.X` matched what handlers actually returned.

### Implemented (complete)

The CUE schemas in `tooling/schema/*.cue` (30 files, 226 types) are now the
**single source of truth**. `tooling/codegen/build_schemas.py` emits the
`internal/api/openapi` package directly, so `swag init` reads generated structs
and there is no second hand-maintained copy. Pipeline:

```
tooling/schema/*.cue
  -> build_schemas.py -> internal/api/openapi/schemas.go
  -> swag init        -> swagger.json -> openapi3.json
  -> orval            -> generated-rq hooks + generated/schemas wire types
```

What was done to get here:

- **Timestamps unified to `int64` Unix millis.** The live server DTOs
  (`internal/application/dto`, e.g. `CreatedAt = now.UnixMilli()`) were already
  int64; CUE said `string` and `openapi` said `time.Time`. CUE now declares
  `int64` (optional → `*int64`), matching the real wire format. Handlers that
  built RFC3339 strings were converted to `.UnixMilli()`.
- **Typed maps preserved.** CUE expresses the value type
  (`config: {[string]: string}`, `metrics: {[string]: float64}`,
  `requests_by_scope: {[string]: int64}`, `toggles: {[string]: bool}`), so the
  generator emits `map[string]string` / `float64` / `int64` / `bool` rather than
  collapsing to `any`.
- **One package.** The separate `internal/api/schema` package is deleted; the 22
  handlers that returned `schema.X` now return `openapi.X`.
- **Enforcement.** `tooling/hooks/pre-commit` runs a CUE drift gate: it
  regenerates the Go + TS to a temp dir and fails if either committed artifact
  is stale. The swagger drift gate then covers `openapi` → `swagger.json`.
- **Lint.** `.golangci.yml` exempts the generated `internal/api/openapi/schemas.go`
  from `govet/fieldalignment` (the generator controls field order).

Result: editing a `.cue` file is the only way to change a wire DTO, and CI fails
if the generated artifacts are not committed.

### Handler `gin.H` migration (complete)

Because `swag init` reads the generated `openapi` structs, a handler emitting a
`gin.H` key absent from the schema would be invisible to the spec. Every handler
response now returns a generated `openapi.X` struct — **no `gin.H` literals
remain in non-test handler code**.

Reconciling the last sites surfaced genuine under-declarations that were fixed in
`tooling/schema/*.cue` (then regenerated), notably:

- `DeviceEvent` gained `severity`/`source`/`timestamp`; `DeviceEventListResult`
  gained `has_more`/`count`.
- `ConnectionMetricsResult`, `DeviceTransferResult`, `TelemetryEntry`,
  `TelemetryHistoryQueryResult`, `ClientCredential`, `LoginResult` /
  `LoginWithTokensResult`, `MFAOperator`, `LockoutStatusResult`,
  `CommandConfirmResult`, `ContactPointTestResult`, `AdminAPIKeyListResult`,
  `UpdateCheckerResult` gained the fields the handlers actually emit.
- New types: `TelemetryCleanupResult`, `MFAVerifiedResult`,
  `AdminClientActionResult`, `WebhookInfoResult`, `ClientCredentialUpdateResult`,
  `ClientSecretRotateResult`.

The pre-commit CUE-drift and swagger-drift gates keep the schema and the emitted
wire format in lockstep.

### Naming

- Schema dir: `tooling/schema/`
- CUE files: one per domain (`devices.cue`, `inbox.cue`, `updates.cue`, ...)
- Go output: `apps/api/internal/api/openapi/schemas.go` (generated)
- TS output: none from CUE — TypeScript comes from the spec via orval (`generated/schemas/`)
- Generator: `tooling/codegen/build_schemas.py`
- Drift gate: `tooling/hooks/pre-commit` (CUE schema drift check)

---

## 3. Vyzorix Request Deduplicator — Transport-Level Request Sharing

### Problem

The current `_batching/request-batcher.ts` collapses duplicate GETs within a
50ms window. POST/PUT/PATCH/DELETE have basic in-flight dedup (exact body
hash match). But with the multi-conversation WebSocket manager, each background
conversation fires `useUserConversation()` independently — if two conversations
share the same backend, they can fire duplicate API calls that aren't deduped.

### Proposal

Port a transport-level request queue that deduplicates ALL HTTP methods —
not just GETs. The queue accepts requests, the worker processes them with
dedup, and the response is shared across all callers that requested the same
operation.

```
Component A → restClient.get('/v1/devices')
Component B → restClient.get('/v1/devices')  (fired 5ms later)
                    ↓
            VyzorixRequestQueue
                    ↓
            One HTTP request to server
                    ↓
            Response shared to both A and B
```

### Implementation

`packages/API_Client/src/vyzorServer/rest/_shared/request-queue.ts` defines
`VyzorixRequestQueue`, now the single dedup primitive `restClient` routes every
HTTP method through:

- hashes `method + url + params/body` (stable, key-order-independent) into a
  dedup key;
- shares the in-flight promise across all callers with the same key — so the
  duplicate-POST case (two conversations firing the same mutation) fires one
  request, not two;
- **keys the entry by the active org as well as method+url+body**: REST org
  scoping rides in the `X-Organization-ID` header, not the URL, so an org switch
  would otherwise let one org share another's response. `setOrganizationContext`
  also resets the queue.
- **is bypassed when the caller supplies an `AbortSignal`**: generated queryFns
  pass TanStack Query's cancellation signal; sharing one promise would let a
  single unmount abort the request for every other sharer.
- registers the in-flight entry synchronously before the executor runs, so a
  caller arriving on the same tick joins it. JS is single-threaded, so no lock
  is needed; the design's "single worker" is intentionally omitted because
  serialising unrelated requests would defeat HTTP/2 multiplexing;
- can serve a completed GET from a short cache (`staleMs`), **disabled by
  default**. A non-zero window is an opaque HTTP-level cache that TanStack
  Query's `invalidateQueries` cannot see, which would make "mutate then refetch"
  observe stale data — so it is opt-in, not the default. In-flight dedup is
  always on.

`restClient` still owns the transport concerns inside the executor it hands the
queue: the connectivity monitor's offline queue, the circuit breaker, token
refresh, and HMAC signing. The old `RESTRequestBatcher` and the per-method
`inFlightRequests` map are removed (the GraphQL batcher is unaffected).

### Naming

- Class: `VyzorixRequestQueue`
- File: `packages/API_Client/src/vyzorServer/rest/_shared/request-queue.ts`
- Config: `{ staleMs: 0, maxQueueSize: 100 }` (`staleMs` opt-in)
- Integration: `restClient` delegates to `getRequestQueue().execute()`

---

## 4. Vyzorix Realtime Channel — Keep `graphql-ws` + the Go Hub

### Decision

**Keep the existing `graphql-ws` client + in-process Go hub.** Do NOT adopt an
external broker (Centrifugo or otherwise) at the current scale. The earlier
"replace raw WebSocket with Centrifugo" proposal targeted code that does not
exist in this repository and solved a problem we do not yet have.

### What actually exists (as of this writing)

- **Frontend**: `apps/VyzoriX_web/src/stores/websocket-store.ts` (154 lines)
  wraps `graphql-ws`'s `createClient` — reconnect (`retryAttempts: Infinity`),
  keepalive (`keepAlive: 10s`), `connectionAckWaitTimeout`, and connection-state
  events all come from `graphql-ws`, not hand-rolled code. Subscriptions use the
  code-first generated documents from `@vyzorix/api-client/generated-graphql`.
  Transport path: `/{org}/graphql/ws` (`server_graphql.go`).
- **Backend**: an in-process hub at `apps/api/internal/ws/` (~4.6k LOC) —
  `hub.go` (register/unregister, broadcast), `subscriptions.go`,
  `telemetry_filter.go`, `message_queue.go` (replay on reconnect),
  `ws_rate_limiter.go`, `compression.go`, `channel/`. Device auth is HMAC
  request signing; org scoping rides the org path segment. This is domain logic,
  not boilerplate.

There is **no** `use-websocket.ts` or `PersistentWebSocketProvider` in this repo
(those are OpenHands files); the "~200 lines of fragile browser lifecycle" the
old §4 described are not present here.

### Why not an external broker now

An external broker (Centrifugo/NATS/Redis pub-sub) exists to solve **horizontal
fan-out**: when N API replicas each hold some WebSocket clients, a message
produced on replica A must reach subscribers on B/C/D. This deployment runs
`numInstances: 1` (`render.yaml`) with **no** Redis/NATS/broker anywhere, so that
problem does not exist yet.

At one instance an in-process hub is strictly simpler and faster (no network hop
to a broker). Adopting Centrifugo today would mean deleting a working, tested,
domain-integrated hub and re-expressing its logic (HMAC auth, telemetry
filtering, org scoping, reconnect replay, rate limiting) as JWT claims + publish
calls — a large rewrite for zero capability gain, plus a new service to deploy
and secure.

### Decision rule

- **Stay on `graphql-ws` + the in-process hub while the API runs at one
  instance.**
- **Revisit when the API scales to `numInstances > 1`** (or a second service
  needs to publish into the same realtime stream).
- **Then, prefer a broker backplane behind the existing hub** (NATS or Redis):
  keep `Hub.Broadcast…` and additionally publish to a shared subject; each
  replica relays to its local clients. This preserves all domain logic and adds
  only cross-instance fan-out — less work and less new surface than replacing
  the stack.
- **Choose Centrifugo (sidecar) only if** you also want its off-the-shelf
  platform features — presence, channel history, an admin UI, broad client SDKs
  (mobile/desktop) — and are willing to own the domain logic as JWT + publish
  calls. That is a product decision, not a technical necessity.

### If/when a broker is introduced

- Backplane: `internal/realtime/` (NATS/Redis publish-subscribe around the hub).
- Channels: `vyzorix:{domain}:{id}:{event_type}` (e.g.
  `vyzorix:device:{imei}:telemetry`).
- The frontend client (`graphql-ws`) stays; only the hub's fan-out changes.

### What is explicitly NOT done

- No Centrifugo dependency (absent from `go.mod` / `package.json`).
- No `centrifuge-js`; subscriptions stay typed GraphQL documents.

---

## 5. Vyzorix Data Pipeline — Keep Zustand + TanStack Query (No RxJS)

### Decision

**Keep the existing data stack: TanStack Query for REST, Zustand stores for
realtime, and one shared merge primitive at the seam. Do NOT introduce RxJS.**
The earlier "unify on RxJS Observables" proposal targeted code that does not
exist in this repository (a 1300-line `conversation-websocket-context.tsx`,
`OpenHandsEvent`, `isAgentServerEvent`, `useObservable`) — those are OpenHands
files — and rxjs is not a dependency.

### What actually exists (as of this writing)

The realtime/data layer is ~870 lines total:

- REST: TanStack Query via the generated react-query hooks (§1).
- Realtime: `apps/VyzoriX_web/src/stores/websocket-store.ts` (graphql-ws) +
  `hooks/realtime/*` + `stores/metrics-realtime-store.ts` feeding Zustand
  stores (`timeline-stream-store`, `log-stream-store`, `diagnostics-store`).
- The one genuine "two paradigms, one buffer" seam is the device timeline:
  REST history (`useDeviceTimeline`) and live WS events (`useDiagnosticStream`)
  both feed `timeline-stream-store`.

### The real fix (shipped)

`timeline-stream-store` had a **dedup asymmetry**: `appendBatch` (REST history)
deduped by event id, but `append` (live WS) did not. The same event arriving via
both paths — a live event, then a history fetch covering the same window, or
vice-versa — produced a duplicate row.

Both paths now go through one `mergeNewestFirst(existing, incoming)` helper that
dedupes by id, prepends newest-first, caps at the ring size, and returns the
existing array unchanged (no state churn) when nothing is new. Covered by three
regression tests including both cross-path orderings.

### Why not RxJS

- It is a dependency + a whole Observable/`useObservable` layer to unify ~870
  lines that already work and are typed. The paradigm shift is not justified by
  the complexity present.
- RxJS in React adds real footguns: subscription leaks, re-render storms, and
  SSR friction (this app uses TanStack Start / SSR).
- The actual problem — REST history + live events landing in one buffer — is
  solved by a single merge function, not by rewriting the data layer.

### Revisit trigger

Reconsider a unified observable layer only if a **second** realtime-merge site
appears (e.g. a second stream that must combine REST history + WS live data with
non-trivial operators). One seam does not justify a paradigm.

### Naming (if the trigger is ever hit)

- Shared merge: `mergeNewestFirst` (currently private to `timeline-stream-store`).
- If promoted: `packages/API_Client/src/vyzorServer/data/` with
  `VyzorixDataPipeline.combine(history, live)` — only then.

---

## 6. Vyzorix Test Selectors — Version-Aware E2E Selectors

### Problem

Tests hardcode API URLs and DOM selectors:
```typescript
await page.goto('/v1/devices/123');
await expect(page.locator('[data-testid="device-list"]')).toBeVisible();
```

When a route changes, tests break silently — no compiler, no type check.

### Proposal

Generate a `selectors` module from the route definitions and OpenAPI paths:

```typescript
// packages/API_Client/src/vyzorServer/testing/selectors.ts
export const vyzorixSelectors = {
  pages: {
    devices: {
      list: { url: '/v1/devices', testId: 'device-list' },
      detail: (imei: string) => ({ url: `/v1/devices/${imei}`, testId: 'device-detail' }),
    },
    updates: {
      versions: { url: '/v1/updates/versions', testId: 'update-versions' },
      push: { url: '/v1/updates/push', testId: 'update-push-form' },
    },
  },
  api: {
    devices: {
      list: 'GET /v1/devices',
      get: (imei: string) => `GET /v1/devices/${imei}`,
      deregister: (imei: string) => `DELETE /v1/devices/${imei}`,
    },
  },
};
```

Tests reference selectors instead of hardcoding:
```typescript
await page.goto(vyzorixSelectors.pages.devices.list.url);
await expect(page.locator(`[data-testid="${vyzorixSelectors.pages.devices.list.testId}"]`)).toBeVisible();
```

### Generation

A script reads `openapi3.json` paths + route definitions and generates the
selectors file. When an API path changes, only the selector file updates —
tests reference the selector, not the raw URL.

### Naming

- File: `packages/API_Client/src/vyzorServer/testing/vyzorix-selectors.ts`
- Generator: `tooling/codegen/build_selectors.py`
- Export: `vyzorixSelectors`
- Pattern: `vyzorixSelectors.pages.{domain}.{action}.{url|testId}`

---

## Implementation Priority

| Phase | What | Effort | Impact | Risk |
|-------|------|-------|--------|------|
| 1 | Vyzorix Request Deduplicator (#3) | Low | Prevents duplicate API calls | Low |
| 2 | Vyzorix Query SDK (#1) | Medium | Deletes 40+ manual hooks, auto cache invalidation | Medium |
| 3 | Vyzorix Schema Definitions (#2) | Medium-high | Eliminates config schema drift | Medium |
| 4 | Vyzorix Realtime Channel (#4) | — | No change: keep `graphql-ws` + Go hub until multi-instance | — |
| 5 | Vyzorix Test Selectors (#6) | Low | Future-proofs test suite | Low |
| 6 | Vyzorix Data Pipeline (#5) | — | No change: keep Zustand + TanStack Query; one merge primitive | — |

Phases 1 and 5 can be done independently and in parallel. Phase 2 depends on
the OpenAPI spec being stable (which it is after the annotation work). Phases 4
and 6 are deliberately deferred: the existing `graphql-ws` + in-process Go hub
is the right design at one instance (§4), and the RxJS data-pipeline rewrite is
not justified at the current data-layer size — the REST↔WS seam is handled by one
shared merge primitive (§5).

---

## Current State (as of this writing)

Already completed:
- ✅ swaggo annotations for all 191 operations (146 paths)
- ✅ build_openapi3.py (Swagger 2 → OpenAPI 3 normalization)
- ✅ orval generated SDK (207 schemas, 0 untyped responses)
- ✅ build_zod.py (207 zod schemas generated)
- ✅ domain layer migrated (generated types + zod + hand-rolled .refine())
- ✅ restClient mutator (circuit breaker, HMAC signing, retry, offline queue)
- ✅ openapi/schemas.go (207 typed wire types, all 24 unknowns fixed)
- ✅ CUE schema layer + build_schemas.py + CUE drift gate (§2; single source of truth)

Not yet started:
- ❌ Vyzorix Test Selectors (#6)

Decided — Vyzorix Data Pipeline (#5):
- ✅ **Keep Zustand + TanStack Query; no RxJS.** The proposal targeted OpenHands
  files that don't exist here and rxjs isn't a dependency. The real REST↔WS seam
  (device timeline) is fixed with one `mergeNewestFirst(existing, incoming)`
  helper in `timeline-stream-store` (both `append` and `appendBatch` dedupe by
  id); 3 regression tests added. Revisit only if a second realtime-merge site
  appears.

Decided — Vyzorix Realtime Channel (#4):
- ✅ **Keep `graphql-ws` + the in-process Go hub.** No external broker at the
  current scale (`numInstances: 1`, no Redis/NATS). The old Centrifugo proposal
  targeted files that don't exist here and solved a fan-out problem we don't have.
  Revisit only at multi-instance; prefer a NATS/Redis backplane behind the hub.

Completed — Vyzorix Request Deduplicator (#3):
- ✅ `VyzorixRequestQueue` (rest/_shared/request-queue.ts) dedupes ALL methods via
  in-flight promise sharing keyed on method+url+body; `restClient` routes every
  method through it. Old GET-only `RESTRequestBatcher` + `inFlightRequests` map
  removed. Optional GET stale cache (`staleMs`) is opt-in / disabled by default
  to avoid a second opaque cache TanStack Query cannot invalidate.

Completed — Vyzorix Schema Definitions (#2):
- ✅ `tooling/schema/*.cue` (30 files, 226 types) is the single source of truth;
  `tooling/codegen/build_schemas.py` emits `internal/api/openapi/schemas.go` (the
  package `swag init` reads). TypeScript is NOT generated from CUE — it comes from the spec
  via swag -> openapi3.json -> orval (`generated/schemas/`).
- ✅ Timestamps unified to int64 Unix millis; typed maps preserved; the separate
  `internal/api/schema` package deleted (22 handlers repointed to `openapi.X`).
- ✅ CUE schema drift gate in `tooling/hooks/pre-commit`.
- ⏳ Follow-up: ~35 handler files still emit `gin.H` directly, so a key absent from
  the schema is not yet a compile error (see §2 "Remaining").

In progress — Vyzorix Query SDK (#1, Phase 2):
- ✅ Data-layer hooks migrated to orval-generated react-query hooks (`src/generated-rq/`):
  commands, diagnostics, events, logs, metrics, registration, updates — including
  the GraphQL-fallback orchestration hooks, with domain types preserved via the `TData`
  generic + internal wire-type casts (committed 2026-08-31; tsc 0 errors, vitest 577 green).
- ✅ Also migrated: alerts, apikey (operator/admin), devices, notifications, organizations,
  sessions, settings, telemetry, inbox (create/register/resend/ack/dismiss), updates
  (push/sync/cancel/detail).
- ✅ **Auto cache invalidation (the §1 goal) is now complete.** `orval.config.js`
  enables `query.useInvalidate` + `shouldExportQueryKey`, org-scopes every generated
  key via `query-key-mutator.ts::organizationScopedQueryKey`, and declares
  `query.mutationInvalidates` per domain tag. The hand-written hooks no longer pass
  a `queryKey` override or call `invalidateQueries` for mutation-driven cases — the
  generated mutations invalidate the related org-scoped queries themselves. GraphQL
  fallbacks (which bypass the generated mutation) call the generated `invalidateGet*`
  helpers, and realtime WebSocket handlers invalidate the same canonical keys.
  Regression coverage: `src/test/hooks/use-devices.test.ts` asserts the device list
  refetches after a deregister mutation with no manual wiring (tsc 0 errors,
  vitest green).
- ✅ Auth orchestration hooks ported onto the generated mutation hooks:
  `auth/use-login.ts` (`usePostAuthLoginTokens` + store ingest via `onSuccess`),
  `auth/use-mfa.ts` (`usePostAuthMfaVerify` + async `/me` reconcile in `onSuccess`),
  `auth/use-logout.ts` (`usePostAuthLogout` + `onSettled` clear),
  `auth/use-password-reset.ts` / `auth/use-register.ts` (thin generated-hook
  re-exports), `auth/use-me.ts` (`useSelectOrganization` uses the generated
  mutation). The generated mutations carry the org-scoped auto-invalidation; the
  wrappers keep only the auth-store side effects. Tests updated to the generated
  `{ data }` variables + response envelope.
- ✅ Auto-invalidation gaps closed for GraphQL-fallback and realtime paths
  (documented in §1 "Limitation: parameterized queries"): commands invalidate via
  the generated key helpers, organization-member changes via the realtime
  member-event handler.
- No remaining old-style hook wrappers outside the intentional non-data utilities.

