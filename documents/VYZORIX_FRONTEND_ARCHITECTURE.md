# Vyzorix Frontend Architecture: Codegen & Data Layer Modernization

## Context

Vyzorix is a device management platform with a Go backend (Gin + swaggo) and a
TypeScript frontend (Vite + React + TanStack Query). The current data layer pipeline is:

```
Go handler annotations (swaggo)
  ‚Üí swag init ‚Üí swagger.json (Swagger 2.0)
  ‚Üí build_openapi3.py ‚Üí openapi3.json (OpenAPI 3.0)
  ‚Üí orval ‚Üí packages/API_Client/src/generated/ (typed axios functions)
  ‚Üí build_zod.py ‚Üí vyzorixUpdateServerAPI.zod.ts (zod schemas)
  ‚Üí domain/ modules (re-exports + hand-rolled business rules)
  ‚Üí apps/VyzoriX_web/src/hooks/ (manual TanStack Query wrappers)
```

This document outlines six architectural improvements, inspired by patterns
observed in production observability platforms, that Vyzorix can adopt ‚Äî with
naming conventions unique to Vyzorix's device-management domain.

---

## 1. Vyzorix Query SDK ‚Äî Auto-Generated Typed Hooks

### Problem

The current pipeline generates typed axios functions via orval, but every hook
in `apps/VyzoriX_web/src/hooks/` is hand-written boilerplate:

```typescript
// 40+ files like this, each ~20 lines of wrapper:
export function useDevices(params?: DeviceParams) {
  return useQuery({
    queryKey: queryKeys.devices(params),
    queryFn: () => getDevices().getDevices(),
  });
}
```

Cache invalidation is manual:
```typescript
onSuccess: () => queryClient.invalidateQueries({ queryKey: ['devices'] }),
```

If a handler adds a new field, the hook doesn't know about it. If a mutation
should refetch related queries, the developer must remember to wire it.

### Proposal

Replace orval with `@rtk-query/codegen-openapi` (or a similar OpenAPI-to-hook
codegen) that reads `openapi3.json` and generates typed query/mutation hooks
directly ‚Äî no manual hook files.

The swaggo `@Tags` annotations already partition the API into tags:
`@Tags devices`, `@Tags updates`, `@Tags auth`, etc. These map directly to
cache tag types for automatic invalidation:

```go
// CreateDevice handler
// @Tags devices
// @Success 201 {object} openapi.DeviceListItem
```

Generates:
```typescript
// packages/API_Client/src/generated/vyzorix-sdk.ts
export const vyzorixApi = createApi({
  baseQuery: vyzorixBaseQuery({ baseURL: '/api' }),
  tagTypes: ['devices', 'updates', 'auth', 'inbox', 'organizations', ...],
  endpoints: (builder) => ({
    getDevices: builder.query<DeviceListResult, void>({
      query: () => '/v1/devices',
      providesTags: ['devices'],
    }),
    deregisterDevice: builder.mutation<SuccessResult, string>({
      query: (imei) => ({ url: `/v1/devices/${imei}`, method: 'DELETE' }),
      invalidatesTags: ['devices'],
    }),
  }),
});

export const {
  useGetDevicesQuery,
  useDeregisterDeviceMutation,
} = vyzorixApi;
```

### What gets deleted (~30 pure API wrapper hooks)

Vyzorix has 58 hook files total. They fall into three categories:

**~30 pure API wrappers (deleted)** — files like `use-devices.ts`,
`use-alerts.ts`, `use-versions.ts`, `use-sessions.ts`, `use-settings.ts`,
`use-contact-points.ts`, `use-organizations.ts`. These are thin wrappers
that call a generated function inside `useQuery`/`useMutation`. No business
logic. RTK Query generates these automatically.

- Manual `queryClient.invalidateQueries` calls
- Manual `queryKey` definitions in `lib/query-keys.ts`

### What stays (~28 hooks)

**~15 orchestration hooks (kept but simplified)** — files like
`use-commands.ts` (command polling with state machine callbacks),
`use-realtime.ts` (WebSocket subscriptions), `_graphql-fallback.ts`
files (REST fails → try GraphQL), `use-login.ts` (MFA challenge flow),
`use-auth-session.ts` (auth context). These have real logic beyond
"call API X" — they stay, but get simpler because cache invalidation
is automatic and they call generated hooks directly.

**~13 non-API hooks (kept, unchanged)** — files like
`use-vyzor-analytics.ts`, `use-connectivity.ts`, `use-debounce.ts`,
`use-pagination.ts`, `use-time-range.ts`, `use-vyzor-language.ts`,
`use-vyzor-error-recovery.ts`. These don't call the API at all — they're
UI utilities that are unaffected by the codegen change.

### What else stays

- `domain/` layer (zod schemas + business validators)
- `restClient` transport (wrapped as `vyzorixBaseQuery`)
- `openapi/schemas.go` (Go wire types)
- `build_openapi3.py` (spec normalization)
- `build_zod.py` (zod generation)

### Naming

- Package: `@vyzorix/api-client` (existing)
- Generated file: `packages/API_Client/src/generated/vyzorix-sdk.ts`
- Hook exports: `useGetDevicesQuery`, `usePushUpdateMutation`, etc.
- Cache tags: derived from swaggo `@Tags` values (`devices`, `updates`, etc.)
- Base query: `vyzorixBaseQuery` (wraps the existing `restClient`)

---

## 2. Vyzorix Schema Definitions ‚Äî CUE-Based Config Types

### Problem

Device settings, thresholds, notification settings, and update push configs are
defined as Go structs in `openapi/schemas.go`. These must manually mirror the
`gin.H` keys that handlers emit. If a handler changes a JSON key, the schema
silently drifts ‚Äî the generated TypeScript type says one thing, the server
sends another.

Current drift detection: the pre-commit swagger drift gate checks that
annotations match `swagger.json`, but it does NOT check that `openapi.X` schemas
match what the handler actually returns in `gin.H`.

### Proposal

Define configuration schemas in CUE files (`.cue`), then generate both Go
structs AND TypeScript types from the same source. CUE is a configuration
language with type constraints, defaults, and validation built in.

```
tooling/schema/
‚îú‚îÄ‚îÄ device_settings.cue      ‚Üí device thresholds, custom name, location, metadata
‚îú‚îÄ‚îÄ notification_settings.cue ‚Üí email/push/webhook notification config
‚îú‚îÄ‚îÄ update_push.cue           ‚Üí push update request shape
‚îú‚îÄ‚îÄ inbox_request.cue         ‚Üí device registration request
‚îî‚îÄ‚îÄ gen.go                    ‚Üí reads .cue, outputs Go + TS
```

Example `.cue` file:
```cue
// device_settings.cue
package vyzorix

#DeviceThresholds: {
	riskWarn:    int | *0
	riskCrit:    int | *0
	thermalWarn: int | *0
	thermalCrit: int | *0
	bufferWarn:  int | *0
	bufferCrit:  int | *0
}
```

Generated Go:
```go
// zz_generated_device_settings.go
type DeviceThresholds struct {
	RiskWarn    int `json:"riskWarn"`
	RiskCrit    int `json:"riskCrit"`
	// ...
}
```

Generated TypeScript:
```typescript
// device_settings.gen.ts
export interface DeviceThresholds {
  riskWarn: number;
  riskCrit: number;
  // ...
}
```

### Pipeline change

```
.cue files ‚Üí [cuetsy] ‚Üí Go structs + TS types
Go structs ‚Üí [swaggo] ‚Üí swagger.json ‚Üí openapi3.json
TS types ‚Üí imported by domain/ modules
```

The handler returns the generated Go struct directly (not `gin.H`), so the
wire format is always consistent with the schema. The OpenAPI spec is generated
from the Go struct, which is generated from CUE ‚Äî one source of truth.

### What gets deleted

- Manual `openapi/schemas.go` struct definitions for config types
- The "does `gin.H` match `openapi.X`?" class of drift bugs

### Naming

- Schema dir: `tooling/schema/`
- CUE files: `device_settings.cue`, `notification_settings.cue`, etc.
- Go output: `zz_generated_*.go` in `internal/api/openapi/`
- TS output: `*.gen.ts` in `packages/API_Client/src/generated/schema/`
- Generator: `tooling/codegen/build_schema.py` (or `gen.go` using cuetsy)

---

## 3. Vyzorix Request Deduplicator ‚Äî Transport-Level Request Sharing

### Problem

The current `_batching/request-batcher.ts` collapses duplicate GETs within a
50ms window. POST/PUT/PATCH/DELETE have basic in-flight dedup (exact body
hash match). But with the multi-conversation WebSocket manager, each background
conversation fires `useUserConversation()` independently ‚Äî if two conversations
share the same backend, they can fire duplicate API calls that aren't deduped.

### Proposal

Port a transport-level request queue that deduplicates ALL HTTP methods ‚Äî
not just GETs. The queue accepts requests, the worker processes them with
dedup, and the response is shared across all callers that requested the same
operation.

```
Component A ‚Üí restClient.get('/v1/devices')
Component B ‚Üí restClient.get('/v1/devices')  (fired 5ms later)
                    ‚Üì
            VyzorixRequestQueue
                    ‚Üì
            One HTTP request to server
                    ‚Üì
            Response shared to both A and B
```

### Implementation

Replace the current `request-batcher.ts` with a `VyzorixRequestQueue` that:
- Hashes the method + URL + body to create a dedup key
- If a request with the same key is in-flight, returns the existing promise
- If a GET was fetched within the stale window (configurable, default 5s),
  returns the cached response without a new request
- Processes the queue in a single worker to avoid race conditions

### Naming

- Class: `VyzorixRequestQueue`
- File: `packages/API_Client/src/vyzorServer/rest/_shared/request-queue.ts`
- Config: `{ staleMs: 5000, maxQueueSize: 100 }`
- Integration: `restClient` delegates to `VyzorixRequestQueue.execute()`

---

## 4. Vyzorix Realtime Channel ‚Äî Centrifuge-Based WebSocket Manager

### Problem

The current WebSocket implementation (`use-websocket.ts`) is hand-rolled with:
- Manual keepalive pings (every 20s when tab is hidden)
- Manual 3-second reconnect with attempt counting
- Manual auth via first-message JSON protocol
- Manual connection state tracking (`"CONNECTING" | "OPEN" | "CLOSED"`)
- Manual `visibilitychange` handler for background tab recovery
- The `PersistentWebSocketProvider` manages multiple WS connections for
  background conversations, each with its own `useWebSocket` instance

This is ~200 lines of fragile browser-specific lifecycle code.

### Proposal

Replace raw WebSocket with Centrifugo (a pub/sub broker with a Go SDK and
JavaScript client). The Go backend runs a Centrifugo server; the frontend
uses `centrifuge-js` to connect and subscribe to channels.

```
Go backend ‚Üí Centrifugo server (embedded or sidecar)
    ‚Üì
Centrifugo handles:
  - WebSocket connection lifecycle (connect, reconnect, backoff)
  - Authentication (JWT or token-based)
  - Channel subscriptions (per-conversation, per-device, per-org)
  - Message delivery guarantees (at-least-once)
  - Connection state notifications
    ‚Üì
Frontend (centrifuge-js) ‚Üí subscribes to channels
  - vyzorix:device:{imei}:telemetry
  - vyzorix:device:{imei}:events
  - vyzorix:updates:push:{pushId}
  - vyzorix:conversation:{id}:events
```

### What gets deleted

- `use-websocket.ts` (entire file ‚Äî replaced by centrifuge-js)
- `PersistentWebSocketProvider` (centrifuge manages multiple subscriptions
  over a single connection, not multiple WS connections)
- Manual keepalive/reconnect logic
- Manual `visibilitychange` handler
- The `headless` prop on `ConversationWebSocketProvider` (Centrifugo handles
  background subscriptions natively)

### What stays

- The event processing logic in `ConversationWebSocketProvider` (what to do
  with received events ‚Äî addEvent, setExecutionStatus, etc.)
- The event store (`use-event-store`)
- The REST history preloading (`useConversationHistory`)

### Naming

- Go module: `internal/realtime/centrifuge/` (Centrifugo server config)
- JS client: `packages/API_Client/src/vyzorServer/realtime/centrifuge-client.ts`
- Channel naming: `vyzorix:{domain}:{id}:{event_type}`
  - `vyzorix:device:{imei}:telemetry`
  - `vyzorix:device:{imei}:events`
  - `vyzorix:updates:push:{pushId}`
  - `vyzorix:inbox:{imei}:status`
- Connection service: `VyzorixRealtimeService` (wraps Centrifuge, exposes
  `subscribe(channel)`, `publish(channel, data)`, `getConnectionState()`)

---

## 5. Vyzorix Data Pipeline ‚Äî RxJS Unified Data Layer

### Problem

The frontend uses two completely different paradigms for data:
- REST: TanStack Query (Promises, cache, refetch)
- WebSocket: imperative `onMessage` handler with `if (isAgentServerEvent(event))`
  chains (1300+ lines in `conversation-websocket-context.tsx`)

These can't compose. You can't merge a REST history fetch with a WebSocket
live stream into a single data pipeline. The "REST first, then WS takes over"
logic is hand-coded with timestamps and dedup.

### Proposal

Unify on RxJS Observables for both REST and realtime. The REST layer returns
`Observable<T>` instead of `Promise<T>`. The WebSocket layer emits events
as an `Observable<VyzorixEvent>`. The UI subscribes to a merged stream that
combines history + live updates.

```typescript
// History (REST) + Live (WS) merged into one observable
const deviceEvents$ = merge(
  from(fetchDeviceHistory(imei)),     // REST: past events
  realtimeService.subscribe(`vyzorix:device:${imei}:events`),  // WS: live events
).pipe(
  scan((events, event) => [...events, event], [] as VyzorixEvent[]),
  shareReplay(1),
);

// UI component
const events = useObservable(deviceEvents$, []);
```

### What gets deleted

- The `useLayoutEffect` that clears events on conversation switch (RxJS
  `switchMap` handles this naturally ‚Äî unsubscribe from old, subscribe to new)
- The timestamp-based dedup logic (RxJS `distinctUntilChanged` by event ID)
- The manual "is REST done before WS connects?" gate (RxJS `combineLatest`
  waits for both)
- Most of the 1300-line `conversation-websocket-context.tsx`

### What stays

- TanStack Query for non-realtime data (settings, org list, etc.) ‚Äî wrapped
  via `lastValueFrom()` where RxJS is needed
- The domain layer (zod schemas, validators)
- The event store (fed by the RxJS subscription)

### Naming

- Package: `packages/API_Client/src/vyzorServer/data/`
- REST wrapper: `VyzorixDataPipeline.fromRest(url)` ‚Üí `Observable<T>`
- WS wrapper: `VyzorixDataPipeline.fromRealtime(channel)` ‚Üí `Observable<T>`
- Merged: `VyzorixDataPipeline.combine(history$, live$)` ‚Üí `Observable<T[]>`
- Event type: `VyzorixEvent` (replaces the current `OpenHandsEvent`-inspired type)

### Effort

This is the highest-risk change ‚Äî it's a paradigm shift from Promises to
Observables across the data layer. Recommend doing this LAST, after the other
five improvements are in place, and only if the WS event handler complexity
justifies it.

---

## 6. Vyzorix Test Selectors ‚Äî Version-Aware E2E Selectors

### Problem

Tests hardcode API URLs and DOM selectors:
```typescript
await page.goto('/v1/devices/123');
await expect(page.locator('[data-testid="device-list"]')).toBeVisible();
```

When a route changes, tests break silently ‚Äî no compiler, no type check.

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
selectors file. When an API path changes, only the selector file updates ‚Äî
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
| 4 | Vyzorix Realtime Channel (#4) | Medium | Replaces fragile WS keepalive/reconnect | Medium |
| 5 | Vyzorix Test Selectors (#6) | Low | Future-proofs test suite | Low |
| 6 | Vyzorix Data Pipeline (#5) | High | Simplifies WS event handler | High |

Phases 1 and 5 can be done independently and in parallel. Phase 2 depends on
the OpenAPI spec being stable (which it is after the annotation work). Phase
4 depends on the Go backend accepting Centrifugo. Phase 6 should only be
attempted after all others are complete, as it's a paradigm shift.

---

## Current State (as of this writing)

Already completed:
- ‚úÖ swaggo annotations for all 191 operations (146 paths)
- ‚úÖ build_openapi3.py (Swagger 2 ‚Üí OpenAPI 3 normalization)
- ‚úÖ orval generated SDK (207 schemas, 0 untyped responses)
- ‚úÖ build_zod.py (207 zod schemas generated)
- ‚úÖ domain layer migrated (generated types + zod + hand-rolled .refine())
- ‚úÖ restClient mutator (circuit breaker, HMAC signing, retry, offline queue)
- ‚úÖ openapi/schemas.go (207 typed wire types, all 24 unknowns fixed)

Not yet started:
- ❌ Vyzorix Query SDK (RTK Query — would delete ~30 pure API wrapper hooks)
- ❌ CUE schema definitions
- ❌ Request deduplicator
- ❌ Centrifuge realtime channel
- ❌ RxJS data pipeline
- ❌ Version-aware test selectors
