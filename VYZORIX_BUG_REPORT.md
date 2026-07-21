# Vyzorix Update Server — Complete Bug & Logic Gap Report

> Two-pass deep analysis of `apps/api` (Go backend).  
> **Total issues: 52** across security, logic, concurrency, error handling, GraphQL, migrations, and lifecycle.  
> Every finding is listed regardless of severity — nothing omitted.

---

## How to Read This Document

| Symbol | Severity | Meaning |
|--------|----------|---------|
| 🔴 | **Critical** | Actively broken at runtime, or exploitable with no prerequisites |
| 🟠 | **High** | Exploitable under realistic conditions, or causes data loss/corruption |
| 🟡 | **Medium** | Logic gap that degrades correctness, security, or reliability |
| 🟢 | **Low** | Code quality issue, performance gap, or minor defensive hardening |

Findings are grouped by area, then ordered by severity within each group.

---

## Table of Contents

1. [Authentication & Sessions](#1-authentication--sessions)
2. [HMAC Signing & Command Dispatch](#2-hmac-signing--command-dispatch)
3. [WebSocket Hub & Real-time Delivery](#3-websocket-hub--real-time-delivery)
4. [GraphQL API](#4-graphql-api)
5. [OTA Updates & Device Management](#5-ota-updates--device-management)
6. [Concurrency & Goroutine Safety](#6-concurrency--goroutine-safety)
7. [Error Handling & Panic Safety](#7-error-handling--panic-safety)
8. [Database, Migrations & Storage](#8-database-migrations--storage)
9. [Audit Logging](#9-audit-logging)
10. [Firebase / FCM](#10-firebase--fcm)
11. [Application Lifecycle & Startup/Shutdown](#11-application-lifecycle--startupshutdown)
12. [Complete Issue Index](#12-complete-issue-index)

---

## 1. Authentication & Sessions

---

### BUG-01 ✅ 🔴 Critical → `scanDevices` missing `organization_id` in Scan — all device list queries broken

**File:** `apps/api/internal/infrastructure/storage/device_storage.go` ~line 155

`deviceColumns` selects 21 columns including `organization_id` as the 11th. `scanDevice` (single row) correctly declares `var organizationID sql.NullString` and includes it in the `row.Scan()` call. `scanDevices` (list, used by every dashboard page) does not declare the variable and skips it in `rows.Scan()`. The column count doesn't match the scan argument count, producing a scan error on every request that returns multiple devices.

```go
// scanDevice (correct) — 21 scan targets
var fcmToken, operatorID, organizationID, deviceName, ... sql.NullString
row.Scan(&d.ID, ..., &operatorID, &organizationID, &d.CreatedAt, ...)

// scanDevices (broken) — 20 scan targets, organization_id absent
var fcmToken, operatorID, deviceName, ... sql.NullString  // organizationID missing
rows.Scan(&d.ID, ..., &operatorID, &d.CreatedAt, ...)     // column order shifts
```

**Fix:** Add `var organizationID sql.NullString`, include it at the correct position in `rows.Scan(...)`, and assign `d.OrganizationID = organizationID.String`.

---

### BUG-02 ✅ 🔴 Critical → Risk score is fully device-controlled — server never recalculates it

**File:** `apps/api/internal/application/event/event_processor.go:300, 469`

The event processor reads `riskScore` directly from the device-supplied telemetry map without any server-side verification, clamping, or recalculation. A compromised device can send `riskScore: 0` to suppress all alerts indefinitely, or `riskScore: 100` to flood operators with false alarms. All the other telemetry fields (thermal, buffer, etc.) needed to compute a server-side score are already in the payload.

**Fix:** Derive the risk score server-side from the corroborating telemetry fields, or at minimum clamp to `[0, 100]` and emit a security event when the device-reported value diverges significantly from the server's estimate.

---

### BUG-03 ✅ 🟠 High → Token rotation hardcodes `role := "operator"` — admins lose their role on first refresh

**File:** `apps/api/internal/application/auth/auth_refresh_token.go:86`

```go
role := "operator"   // ← hardcoded
accessToken, expiresAt, err := s.jwtManager.Generate(op.ID, op.Email, op.Name, role)
```

Every refresh token rotation issues a new JWT with `role = "operator"` regardless of the operator's actual role. SuperAdmins and org-level admins are silently downgraded on the first token refresh and regain their privileges only after a full re-login.

**Fix:** Fetch the operator's membership role from the org-membership table (or from `op.Role` if it's stored on the operator record) and pass that to `Generate`.

---

### BUG-04 ✅ 🟠 High → Pre-MFA sessions bypass MFA enforcement permanently

**File:** `apps/api/internal/api/middleware/cookie_auth.go:62`

`ValidateSession` verifies session validity (expiry, existence) but does not check whether the session was established after MFA was completed. An operator who had an active session when MFA was enabled on their account (or when MFA was enforced org-wide) keeps that session working with no second factor required, until it expires naturally.

**Fix:** Add an `mfa_verified_at` or `mfa_complete bool` column to sessions. On MFA toggle, invalidate all pre-existing sessions or add a check in `ValidateSession` that rejects sessions whose creation timestamp pre-dates the operator's `mfa_enabled_at`.

---

### BUG-05 ✅ 🟠 High → Disabling MFA does not revoke active sessions or refresh tokens

**File:** `apps/api/internal/application/auth/auth_totp_mfa.go:90`

`DisableMFA` clears the TOTP secret and marks the field unset, but never calls `LogoutAll` or `RevokeAllRefreshTokens`. An attacker who briefly accesses an account, disables MFA, and re-enables it later keeps all their persistent refresh tokens alive indefinitely.

**Fix:** Call `s.authService.RevokeAllRefreshTokens(ctx, operatorID)` and `s.authService.LogoutAll(ctx, operatorID)` at the end of `DisableMFA`, mirroring the password-reset flow.

---

### BUG-06 ✅ 🟡 Medium → Password *change* does not invalidate existing sessions

**File:** `apps/api/internal/application/auth/auth_login_session.go:528`

The password *reset* flow (via reset token) correctly calls `LogoutAll` + `RevokeAllRefreshTokens`. The password *change* flow (authenticated user changing their own password) does not. A stolen session survives a password change indefinitely.

**Fix:** Mirror the reset flow — call `LogoutAll` and `RevokeAllRefreshTokens` after a successful password change, excluding the current session if you want to keep the caller logged in.

---

### BUG-07 ✅ 🟡 Medium → GitHub OAuth allows account creation with unverified email — account takeover risk

**File:** `apps/api/internal/api/handlers/auth/auth_oauth.go:461`

Google OAuth (line 258) correctly checks `googleClaims.EmailVerified` and rejects unverified emails. The GitHub implementation has no equivalent check. An attacker can register a GitHub account with a target's email address before the target does, then use OAuth to create a Vyzorix account tied to that email.

**Fix:** Call the GitHub `/user/emails` API and require the primary email to have `verified: true` before allowing sign-in or sign-up.

---

### BUG-08 ✅ 🟡 Medium → Registration has a TOCTOU race on email uniqueness

**File:** `apps/api/internal/application/auth/auth_login_session.go:302`

```go
existing, _ := s.operatorRepo.FindByEmail(ctx, email)
if existing != nil { return ErrEmailTaken }
// ← gap: another goroutine can register same email here
s.operatorRepo.Create(ctx, op)
```

Two concurrent registrations for the same email both pass the existence check and then both attempt the insert. The SQLite unique index catches one, but the error surfaces as an opaque DB error rather than a clean "email already taken" response. Under load this also means brief data inconsistency.

**Fix:** Wrap the check-and-create in a transaction with `SELECT ... FOR UPDATE` semantics (use `BEGIN IMMEDIATE` in SQLite), or rely entirely on the unique index and map the constraint error to `ErrEmailTaken` at the repository layer.

---

### BUG-09 ✅ 🟡 Medium → `oauthStateRepo` nil check skips CSRF state validation silently

**File:** `apps/api/internal/api/handlers/auth/auth_oauth.go:146, 339`

If `oauthStateRepo` is nil, state validation is either skipped or produces a 500 error instead of rejecting the request with a clear CSRF failure. The nil case should be an unconditional startup error, not a runtime conditional.

**Fix:** Treat `oauthStateRepo == nil` as a misconfiguration — fail startup rather than silently degrading CSRF protection at runtime.

---

## 2. HMAC Signing & Command Dispatch

---

### BUG-10 ✅ 🔴 Critical → Command handler runs before `requireStrictHMAC()` — commands execute unsigned

**File:** `apps/api/internal/api/server_routes.go:244–247`

```go
deviceMgmt.POST("/:imei/command",
    s.commandHandler.Handle,   // ← executes first
    s.requireStrictHMAC(),     // ← runs after, too late
)
```

In Gin, handlers listed in `POST(path, h1, h2, ...)` execute in order. The strict HMAC check fires *after* the command has already been dispatched to the device. Any caller who makes it past the basic (non-strict) HMAC check can send commands without satisfying the stricter validation.

**Fix:** Either apply `requireStrictHMAC()` via `.Use()` on the `deviceMgmt` group *before* registering the route, or reorder so it is listed first: `POST("/:imei/command", s.requireStrictHMAC(), s.commandHandler.Handle)`.

---

### BUG-11 ✅ 🟠 High → A device can forge signed requests for any other device's IMEI

**File:** `apps/api/internal/api/middleware/request_signing.go:202` + `server_routes.go:237`

The signing middleware authenticates `clientID` against its stored secret but never checks `clientID == c.Param("imei")`. A compromised device with a valid credential pair can send a correctly-signed request targeting a *different* device's IMEI and it will be accepted.

**Fix:** After authenticating the client, assert `clientID == c.Param("imei")`; reject with HTTP 403 if they differ.

---

### BUG-12 🟡 Medium → Body encryption: ciphertext (not plaintext) is restored to `c.Request.Body`

**File:** `apps/api/internal/api/middleware/request_signing.go:169`

When `X-Encrypted-Body` is present, the middleware decrypts the body and stores the plaintext in the Gin context as `"signed_body"`. However, it then restores the *original ciphertext* to `c.Request.Body`. Any downstream handler calling `c.ShouldBindJSON()` reads ciphertext and fails to parse it. The plaintext is inaccessible via standard Gin binding.

**Fix:** After successful decryption, restore the decrypted bytes to `c.Request.Body` (not the ciphertext), so handlers can use `c.ShouldBindJSON()` normally. Or update all encrypted-body handlers to read from `c.GetString("signed_body")` and document this as the contract.

---

### BUG-13 ✅ 🟡 Medium → Rate limiter silently loses fractional token-refill time — effective rate is lower than configured

**File:** `apps/api/internal/api/middleware/api_rate_limiter.go:177–184`

```go
tokensToAdd := int(elapsed / l.Refill)           // fractional part truncated
b.last = b.last.Add(time.Duration(tokensToAdd) * l.Refill)  // remainder discarded
```

For bursty traffic, `elapsed` is frequently `1.9 × Refill`. Only 1 token is added and `0.9 × Refill` of time is thrown away each call. Over time, clients are granted fewer tokens per second than configured.

**Fix:** Track `b.last` as `b.last.Add(elapsed)` (or accumulate the remainder), not as `b.last + tokensToAdd * Refill`.

---

### BUG-14 🟡 Medium → Replay cache eviction iterates the entire map — O(N) latency spike under load

**File:** `apps/api/internal/api/middleware/replay_protection.go:67`

When the replay cache is full, 10% of entries are evicted by iterating the whole map. On the hot HMAC-signing path under high request volume, this causes periodic latency spikes proportional to the cache size.

**Fix:** Replace the map with a ring buffer or LRU structure to achieve O(1) eviction.

---

### BUG-15 ✅ 🟢 Low → Signing error codes let attackers oracle-test valid client IDs

**File:** `apps/api/internal/api/middleware/request_signing.go:382–423`

Error code `SIGN_004: Unknown or inactive client` is distinct from `SIGN_003: Invalid signature`. An attacker can use this to enumerate valid `clientID` values before attempting to forge signatures.

**Fix:** Collapse both cases into a single generic `SIGN_003: Signature verification failed` response.

---

## 3. WebSocket Hub & Real-time Delivery

---

### BUG-16 ✅ 🟠 High → Message queue partial replay is silently dropped on reconnect

**File:** `apps/api/internal/ws/message_queue.go:414`

When a device reconnects, `ReplayQueue` replays persisted messages into the destination channel. If that channel is full mid-replay, it returns the count of replayed messages and stops — the remaining queued messages are never retried. The hub's caller does not handle the partial case.

**Fix:** Return an explicit "partial" indicator from `ReplayQueue` and schedule a retry (e.g. after a short backoff), or block until all messages have been delivered rather than silently stopping.

---

### BUG-17 ✅ 🟠 High → `time.After()` inside `SendWithDeliveryConfirmation` leaks timer goroutines

**File:** `apps/api/internal/ws/hub.go:503`

```go
select {
case <-confirmCh:
    return nil
case <-time.After(timeout):   // new goroutine every call, never cancelled on early return
    return ErrTimeout
}
```

If `confirmCh` fires first, the `time.After` goroutine leaks until the timeout expires. Under high command throughput, this builds up a large number of leaked goroutines.

**Fix:** Use `t := time.NewTimer(timeout); defer t.Stop()` and select on `t.C`.

---

### BUG-18 ✅ 🟠 High → Message queue async persist/delete races itself — messages can be replayed before persisted or deleted before written

**File:** `apps/api/internal/ws/message_queue.go:188, 197, 412, 435`

`persistMessage` and `deleteMessage` are called as untracked goroutines (`go q.persistMessage(msg)`). If a device reconnects before `persistMessage` completes, `ReplayQueue` reads from the DB and finds nothing — the message is lost. Conversely, `deleteMessage` can execute before `persistMessage` if the scheduler reorders them.

**Fix:** Either persist synchronously before putting the message in the in-memory queue, or use a sequenced worker goroutine with an ordered job channel rather than ad-hoc `go` calls.

---

### BUG-19 ✅ 🟠 High → WebSocket device connection is unauthenticated when `EnforceHMAC=false`

**File:** `apps/api/internal/ws/websocket_stream.go:52`

HMAC verification on the WebSocket upgrade path is gated on `config.EnforceHMAC`. When the flag is false (possible in dev or a misconfigured production), any client that passes the origin check can upgrade to a WebSocket connection, join the hub, and receive real-time telemetry.

**Fix:** Make the WebSocket upgrade always require authentication. If `EnforceHMAC=false` is needed for development, gate it on `NODE_ENV != "production"` and log a prominent warning at startup.

---

### BUG-20 ✅ 🟡 Medium → Hub `Run` loop has no panic recovery — one nil pointer crashes all real-time connections

**File:** `apps/api/internal/ws/hub.go:180`

A nil pointer or closed-channel panic inside the hub's main goroutine kills the goroutine with no recovery. All WebSocket device connections go offline until the process restarts.

**Fix:** Add `defer func() { if r := recover(); r != nil { log.Error("hub panic", r); go h.Run() } }()` at the top of `Hub.Run`, or wrap the run loop in a supervised restart harness.

---

### BUG-21 ✅ 🟡 Medium → Subscription callbacks spawn unbounded goroutines with no concurrency limit

**File:** `apps/api/internal/ws/subscriptions.go:229, 246, 253, 269, 284, 299`

Every subscription event fires `go w.callback(data)` with no semaphore, timeout, or limit. If a callback blocks (e.g. on a slow DB write or a hung HTTP call), each new event spawns another goroutine. Under a burst of telemetry events this leads to unbounded goroutine accumulation.

**Fix:** Use a bounded worker pool or a non-blocking send to a buffered channel per subscription, dropping or logging events when the subscriber is too slow.

---

### BUG-22 🟡 Medium → Hub `SetOnline`/`SetOffline` called outside the hub lock — race with concurrent unregister

**File:** `apps/api/internal/ws/hub.go:187–200`

After acquiring and releasing `h.mu.Lock()` to update the clients map, the hub calls `h.deviceRepo.SetOnline` and `h.messageQueue.ReplayQueue` outside the lock. A concurrent `Unregister` for the same `deviceID` can run between the map update and the DB call, resulting in `SetOffline` being called *before* `SetOnline` for the new connection, leaving the device incorrectly marked online.

**Fix:** Either sequence the DB call inside the lock (accepting the lock-hold duration), or use a per-device serialization primitive (e.g. a device-keyed mutex or a command channel) for lifecycle transitions.

---

### BUG-23 🟡 Medium → `ws_rate_limiter.go` acquires two locks in inconsistent order — deadlock risk

**File:** `apps/api/internal/ws/ws_rate_limiter.go:102, 107, 150, 155, 162, 178`

`incrementAllowed`/`incrementLimited` acquire `metricsMu` alone. `GetMetrics` acquires `mu.RLock()` *then* `metricsMu.Lock()`. If any code path ever holds `metricsMu` and then acquires `mu`, the order is reversed — classic deadlock setup.

**Fix:** Establish a canonical lock-acquisition order (`mu` always before `metricsMu`) and enforce it across all methods.

---

### BUG-24 🟡 Medium → Broadcast over large client sets holds `RLock` for the full iteration — starves new connections

**File:** `apps/api/internal/ws/hub.go:325, 335`

`BroadcastTelemetryToFiltered` iterates `h.clients` under `RLock`. For large device fleets this holds the read lock for the entire duration of the broadcast loop. `Register` and `Unregister` (which need `Lock`) are blocked for the whole time, delaying new device connections during high-throughput broadcasts.

**Fix:** Snapshot the client map under `RLock` into a local slice, release the lock, then iterate the snapshot.

---

### BUG-25 🟢 Low → Gzip writers are allocated per message — GC pressure under high telemetry

**File:** `apps/api/internal/ws/compression.go:71`

A new `gzip.Writer` is created for every compressed message. Under high telemetry volume this creates significant GC pressure.

**Fix:** Use a `sync.Pool` of `gzip.Writer` instances, calling `Reset(w)` between uses.

---

### BUG-26 ✅ 🟢 Low → 500 ms latency threshold is measured but never acted on

**File:** `apps/api/internal/ws/hub.go:428–430`

When broadcast latency exceeds 500 ms a warning is logged and a counter incremented. No corrective action is taken (e.g. load shedding, priority degradation, alerting).

**Fix:** Either enforce the threshold (drop low-priority messages when over budget) or remove the threshold comment so future developers don't assume it is enforced.

---

### BUG-27 🟢 Low → 64-bit metric fields on `ws_client` are read without atomics — race on 32-bit targets

**File:** `apps/api/internal/ws/ws_client.go:68`

Individual metric fields (`LastConnectedAt`, etc.) are updated with `sync/atomic` calls but `GetMetrics` returns `c.metrics` by value without any atomic fence. On 32-bit architectures, reading a 64-bit value non-atomically is a data race.

**Fix:** Either protect `GetMetrics` with the same atomic loads used in the writers, or protect the whole struct with a `sync.RWMutex`.

---

## 4. GraphQL API

---

### BUG-28 ✅ 🟠 High → No query depth or complexity limit — GraphQL DoS via deeply nested queries

**File:** `apps/api/internal/api/graphql/handler/graphql_handler.go:101`

`gql.Params` has no `MaxDepth` or complexity validator. A caller can send `org { members { org { members { org { ... } } } } }` to cause recursive resolution, exhausting server memory or CPU.

**Fix:** Add a depth limiter (e.g. `graphql-go` supports `MaxDepth`) and a complexity budget before executing any query.

---

### BUG-29 ✅ 🟠 High → `TestWebhook` mutation is an SSRF vector — URL argument not validated

**File:** `apps/api/internal/api/graphql/resolver/mutation_resolver.go:637, 652`

The mutation accepts a raw URL string and passes it directly to `r.WebhookClient.Test(url)` with no validation against an allowlist or private-IP blocklist. An attacker can probe internal services (metadata endpoints, other containers) via the server's network identity. The error from the webhook client is also returned verbatim, leaking internal network topology.

**Fix:** Validate the URL scheme (`https` only), resolve the hostname, and reject RFC-1918 / loopback / link-local addresses. Return only a generic error message to the caller.

---

### BUG-30 ✅ 🟡 Medium → GraphQL Playground and introspection are enabled in production

**File:** `apps/api/internal/api/graphql/handler/graphql_handler.go:185`

The Playground UI and introspection are registered unconditionally, with no check for `NODE_ENV` or a feature flag. In production this exposes the full schema to any authenticated (and potentially unauthenticated, if auth is misconfigured) caller.

**Fix:** Gate both behind `if config.Env != "production"` or an explicit `GRAPHQL_INTROSPECTION=true` env flag.

---

### BUG-31 🟡 Medium → `SendCommand` mutation is not transactional across DB, WebSocket, and FCM

**File:** `apps/api/internal/api/graphql/resolver/mutation_resolver.go:103`

The mutation writes a command record to the DB, then attempts WebSocket delivery, then FCM. If the DB write succeeds but WS/FCM delivery fails, the command is "sent" in the DB but never reaches the device, with no retry or rollback. The operator's dashboard shows the command as dispatched when it was actually dropped.

**Fix:** Implement an outbox pattern: write the command to a `pending_commands` table atomically, then deliver asynchronously with retries. Mark delivered only on confirmed receipt.

---

### BUG-32 🟡 Medium → Legacy `/graphql` route uses manual org membership checks — some resolvers skip them

**File:** `apps/api/internal/api/graphql/handler/graphql_handler.go` + `query_resolver.go:23`

Org-scoped routes (`/v1/orgs/:orgId/graphql`) benefit from the `OrganizationMembership` middleware. The legacy `/graphql` route does not. Resolvers that forget to do their own membership check (e.g. `GetMySettings` at line 23) return data without verifying the operator is active in the current org context.

**Fix:** Deprecate or remove the legacy route. If it must remain, apply `OrganizationMembership` middleware to it as well.

---

### BUG-33 🟡 Medium → `GetAllConnections` is O(N×M) — iterates all WS clients for every device in the org

**File:** `apps/api/internal/api/graphql/resolver/query_resolver.go:612, 646`

The resolver fetches up to 1,000 devices from the DB, then for each device iterates all active WebSocket clients to find a matching connection. With 1,000 devices and 1,000 WS clients this is 1,000,000 comparisons per GraphQL call.

**Fix:** Index WS clients by device ID in the hub (`map[deviceID]*Client`) and do a direct lookup instead of linear scan.

---

### BUG-34 🟢 Low → GraphQL argument type assertions use `_, ok` discards — zero-value silently used on type mismatch

**File:** `apps/api/internal/api/graphql/resolver/mutation_resolver.go`, `organization_resolver.go` (multiple lines)

```go
deviceID, _ := p.Args["deviceId"].(string)   // empty string on mismatch, no error
limit, _ := p.Args["limit"].(int)            // 0 on mismatch, queries full table
```

A type mismatch (e.g. client sends an integer where a string is expected) silently produces the zero value. For `limit`, this can return the entire table.

**Fix:** Check `ok` and return a `gql.NewLocatedError("invalid argument type", ...)` when the assertion fails.

---

## 5. OTA Updates & Device Management

---

### BUG-35 ✅ 🟡 Medium → APK path traversal not fully blocked — encoded sequences not checked

**File:** `apps/api/internal/api/handlers/updater/update_check.go:155`

```go
strings.ContainsAny(filename, "/\\")
```

This blocks literal slashes but not URL-encoded traversal (`%2F`, `%5C`) or bare `..` on operating systems that resolve it without a separator. Using `filepath.Base(filename)` and comparing against a known-good filename from the manifest is the correct approach.

**Fix:** Use `filepath.Base(filename)` and reject any result that does not exactly match the expected APK filename from the version manifest.

---

### BUG-36 🟡 Medium → Server never verifies APK integrity against its own manifest before serving

**File:** `apps/api/internal/api/handlers/updater/update_check.go:163`

SHA-256 is only checked if the *client* sends an `X-APK-SHA256` header. The server does not verify the on-disk APK against the manifest's declared hash before streaming it. A tampered APK file on disk would be served without complaint.

**Fix:** At startup (or on first request per version), compute and cache the SHA-256 of each APK file and compare it against the manifest. Refuse to serve if they don't match.

---

### BUG-37 🟡 Medium → Device re-registration delete-then-create is not atomic — race window

**File:** `apps/api/internal/application/device/device_service.go:764`

Re-registering a previously deregistered device deletes the old record and inserts a new one as two separate statements outside a transaction. A concurrent duplicate registration can slip between the delete and insert, producing a unique constraint error or two active records for the same device.

**Fix:** Wrap the delete and insert in a single `BEGIN IMMEDIATE` transaction, or use `INSERT OR REPLACE` with a unique constraint on the device ID.

---

### BUG-38 ✅ 🟡 Medium → Telemetry `GetLatest` and `GetStats` have no organization ownership check

**File:** `apps/api/internal/api/handlers/telemetry_history.go:308, 352`

Both handlers take `deviceId` from the URL and query the telemetry repository directly — no org ownership check. Any authenticated operator from any organization can read another org's device telemetry if they know or enumerate the device ID. The `verifyDeviceInOrganization` helper is already defined in the same file (line 63) and used by other handlers; it was simply not called in these two.

**Fix:** Call `h.verifyDeviceInOrganization(c, deviceID, middleware.GetOrganizationID(c))` at the top of both handlers, returning 403 on failure.

---

### BUG-39 ✅ 🟡 Medium → Telemetry field values are not validated — devices can send extreme/corrupt values

**File:** `apps/api/internal/api/handlers/device/device_telemetry_handler.go:68, 71, 74`

`strconv.ParseInt`/`Atoi` parse errors are silently discarded — invalid field values default to 0. A device sending `thermalTemp: "hot"` stores 0°C with no error log. Additionally, there are no range checks on any telemetry field, so `thermalTemp: 999999` is stored and propagated to the alert engine.

**Fix:** Return an HTTP 400 on parse errors. Add min/max range validation (configurable per field) before persisting.

---

### BUG-40 🟡 Medium → Device status race — WebSocket `SetOnline` and REST `ConfirmDevice` can overwrite each other

**File:** `apps/api/internal/application/device/device_service.go:840` + `apps/api/internal/ws/hub.go:203`

If a device confirms registration via REST and simultaneously connects via WebSocket, both code paths call `deviceRepo.SetOnline(true)` concurrently. The timing determines the final state. In the worst case, the WebSocket connection succeeds but is then overwritten by a stale REST response marking the device offline.

**Fix:** Use a single authoritative state machine for device online/offline transitions, serialized through the hub's event channel rather than direct repo calls from both paths.

---

## 6. Concurrency & Goroutine Safety

---

### BUG-41 ✅ 🟠 High → Invitation email goroutines capture variables from a loop — classic Go closure bug potential

**File:** `apps/api/internal/application/organization/invitation_service.go:169, 310, 375`

While the current variables captured are not loop iterators, the pattern of `go func() { emailService.Send(email) }()` with no `WaitGroup`, context, or timeout means that if `emailService.Send` blocks indefinitely the goroutine leaks for the process lifetime. There is also no cap on how many email goroutines can be outstanding simultaneously.

**Fix:** Use a timeout context for each email send goroutine, and add a `WaitGroup` or a bounded worker pool.

---

### BUG-42 🟡 Medium → `sync.WaitGroup` misuse risk in hub shutdown — `Done()` count may not match `Add()` count

**File:** `apps/api/internal/ws/hub.go` (shutdown path)

Background goroutines started inside the hub (pump goroutines per client) are not tracked by a `WaitGroup`. On shutdown, the hub cannot wait for all client goroutines to drain before returning, potentially causing writes to closed channels or nil pointer dereferences during teardown.

**Fix:** Track every client goroutine with a `sync.WaitGroup`. In the shutdown path, close the hub's done channel and then call `wg.Wait()` before returning.

---

### BUG-43 🟢 Low → Goroutines for subscription callbacks have no error feedback path

**File:** `apps/api/internal/ws/subscriptions.go:229+`

`go w.callback(data)` discards all errors from the callback. A subscription that fails silently (e.g. due to a DB error) is impossible to detect or retry.

**Fix:** Return errors from callbacks via a result channel, or at minimum log failures with enough context to identify the failing subscription.

---

## 7. Error Handling & Panic Safety

---

### BUG-44 ✅ 🟠 High → `panic()` in non-test infrastructure code — crypto and domain failures crash the API

**Files & lines:**
- `apps/api/internal/infrastructure/uuid/uuid.go:54` — panics on crypto reader failure
- `apps/api/internal/infrastructure/crypto/command_signer.go:148` — panics on HMAC key hash failure
- `apps/api/internal/domain/organization/id_values.go:36, 92, 148` — panics on domain validation
- `apps/api/internal/domain/device/device_values.go:30, 88` — panics on domain validation
- `apps/api/internal/api/graphql/context/context.go:46` — panics if operator not in context

These panics propagate up through Gin's recovery middleware and result in a 500 response, but they also produce a goroutine stack dump in logs and abort any in-flight logic in that goroutine. If the panic is in a background goroutine (not a request goroutine), there is no recovery and the process crashes.

**Fix:** Replace panics with returned errors in all non-main code. Reserve `panic` for truly unrecoverable programmer errors caught by tests.

---

### BUG-45 ✅ 🟠 High → Sentinel errors compared with `==` instead of `errors.Is()` — breaks when errors are wrapped

**Files:** `auth_login_session.go:103, 174`, `updates_storage.go`, `device_service.go`, `inbox_service.go`, GraphQL resolvers (extensive)

```go
if err == sql.ErrNoRows { ... }          // breaks if wrapped
if err == device.ErrNotFound { ... }     // breaks if wrapped with %w
```

Any wrapping with `fmt.Errorf("...: %w", err)` breaks `==` comparisons. Most of the infrastructure layer wraps errors, so these comparisons silently fall through to the wrong branch.

**Fix:** Replace all `err == sentinel` with `errors.Is(err, sentinel)` throughout the codebase.

---

### BUG-46 ✅ 🟡 Medium → Logout failure is silently ignored — user appears logged out but session may persist

**File:** `apps/api/internal/api/handlers/auth/auth_logout.go:40`

```go
_ = h.authService.Logout(...)
```

If the session deletion fails (e.g. DB error), the response is still HTTP 200 and the client clears its cookie. The server-side session record remains active — the user is not actually logged out.

**Fix:** Check and return the error. Either respond with HTTP 500 (retry-able by client) or, if the cookie is being cleared anyway, log the failure and send a security audit event.

---

### BUG-47 🟡 Medium → Verification email failure silently ignored at registration — user never receives verification link

**File:** `apps/api/internal/api/handlers/auth/auth_register.go:108`

```go
_ = h.emailSvc.SendVerificationEmail(...)
```

Registration succeeds and returns HTTP 201, but if the email service is down or misconfigured the user never receives their verification email. They have an account they cannot activate, with no indication anything went wrong.

**Fix:** At minimum log the failure with the user ID so it can be retried manually. Ideally, return a warning in the response body or queue the email for retry.

---

### BUG-48 🟡 Medium → Command delivery status sync failure silently ignored — dashboard shows wrong state

**File:** `apps/api/internal/application/updates/updates_push_service.go:221`

```go
_ = s.commandService.MarkDelivered(ctx, cmdResp.CommandID)
```

If `MarkDelivered` fails, the command record stays in `pending` state in the DB while the device has already received it. The dashboard will show the command as undelivered and may trigger unnecessary retries.

**Fix:** Log the failure and enqueue a retry for `MarkDelivered`, or surface the error to the caller.

---

### BUG-49 ✅ 🟡 Medium → `operatorRepo.Update()` return value ignored in multiple call sites

**Files:**
- `apps/api/internal/application/organization/organization_service.go:191`
- `apps/api/internal/application/auth/auth_login_session.go:146`

```go
_ = s.operatorRepo.Update(ctx, op)
```

A failed update silently leaves the operator record in stale state (e.g. last-login timestamp not updated, org role not persisted).

**Fix:** Check and propagate these errors, or at minimum log them with the operator ID so failures are visible.

---

### BUG-50 🟢 Low → Login notification email spawned in goroutine with error discarded

**File:** `apps/api/internal/api/handlers/auth/auth_login.go:234`

```go
go func() { _ = h.emailService.SendNewLoginNotificationEmail(...) }()
```

Login notification failures are never logged or surfaced. While not security-critical, these notifications are a key security signal for account holders.

**Fix:** Log failures at WARN level with the operator ID.

---

## 8. Database, Migrations & Storage

---

### BUG-51 ✅ 🔴 Critical → Migrations are not wrapped in transactions — partial migration leaves DB in unknown state

**File:** `apps/api/internal/infrastructure/storage/sqlite.go:222`

`runMigrations` iterates over pending versions and calls `m.Apply(db)` for each. Neither the entire migration loop nor individual migrations are wrapped in a transaction. If a migration contains multiple DDL statements and fails partway through, the database is left in a partially migrated state with the version table showing the old version number — the next startup will re-attempt the same migration and likely fail again on the already-applied statements.

**Fix:** Wrap each migration in `BEGIN IMMEDIATE; ... COMMIT;` (or `ROLLBACK` on error). Multi-statement migrations should be a single atomic transaction.

---

### BUG-52 ✅ 🟡 Medium → Transaction `Rollback` errors are silently discarded throughout storage layer

**Files:** Throughout `apps/api/internal/infrastructure/storage/` (WithTx wrapper)

```go
if err != nil {
    _ = tx.Rollback()   // rollback error discarded
    return err
}
```

A failed rollback (e.g. connection lost) leaves the transaction in an indeterminate state with no log entry. Subsequent operations may see stale locks or incorrect data.

**Fix:** Log `tx.Rollback()` errors at ERROR level. If rollback fails, the application should not continue operating on the assumption that state is consistent.

---

## 9. Audit Logging

---

### BUG-53 ✅ 🟠 High → Audit log goroutines are fire-and-forget — security events dropped on shutdown

**File:** `apps/api/internal/audit/audit_logger.go:72`

`LogEvent` spawns an untracked goroutine for every audit write. On SIGTERM, in-flight goroutines for security events (login, MFA change, command dispatch, deregistration) are silently discarded.

**Fix:** Use a buffered channel and a single writer goroutine. On shutdown, close the channel and `wg.Wait()` for the writer to drain it before the process exits.

---

### BUG-54 ✅ 🟠 High → Audit logs co-located with application data — trivially tampered

**File:** `apps/api/internal/audit/audit_repository.go:82`

All audit events are stored in the same SQLite database the application writes to. Any SQL injection, compromised credential, or direct file access allows an attacker to delete or modify their own audit trail.

**Fix:** Forward audit events to an append-only external sink (structured log file, syslog, or a separate append-only database). At minimum, write to a separate SQLite file with a read-only connection in the main app.

---

### BUG-55 🟡 Medium → Audit log create failure in inbox service is silently ignored

**File:** `apps/api/internal/application/inbox/inbox_service.go:523`

```go
_ = s.logRepo.Create(ctx, log)
```

If the audit log write fails, no error is logged or returned. Security-relevant inbox events have a silent gap in the audit trail.

**Fix:** At minimum log the failure at ERROR level with the event payload so it can be reconstructed manually.

---

## 10. Firebase / FCM

---

### BUG-56 🟡 Medium → FCM circuit breaker half-open counter not reset on panics — breaker stays open permanently

**File:** `apps/api/internal/infrastructure/fcm/fcm_circuit_breaker.go:93`

`halfOpenCalls` is incremented before the outgoing FCM call. If the caller panics or returns early without calling `RecordSuccess`/`RecordFailure`, the counter is never corrected. The breaker exhausts its half-open quota on phantom calls and stays open indefinitely.

**Fix:** Use a `defer` inside the FCM call wrapper to always call `RecordFailure` unless `RecordSuccess` has already been called.

---

### BUG-57 🟡 Medium → Malformed Firebase credentials at startup fail hard — no graceful degradation path

**File:** `apps/api/internal/infrastructure/fcm/fcm_client.go:33, 45`

Empty credentials produce a disabled client (graceful). Malformed credentials (provided but invalid JSON or wrong structure) cause `firebase.NewApp` to return an error that propagates all the way up and stops the API from starting.

**Fix:** On invalid credentials, log a startup warning and continue with FCM disabled (same as the empty-credentials path). Add a startup validation check that clearly identifies the misconfiguration without crashing.

---

### BUG-58 🟡 Medium → No persistent retry queue for FCM — notifications lost on process restart during retry window

**File:** `apps/api/internal/infrastructure/fcm/notifier.go:184`

`EnhancedNotifier` performs in-memory exponential backoff retries. If the service restarts during the retry window, all pending retries are lost. Security-critical notifications (device wake, command push) are silently dropped.

**Fix:** Persist pending notifications to a `pending_fcm` table and retry from there on startup, implementing an outbox pattern similar to BUG-31.

---

### BUG-59 🟢 Low → `SafeNotifier` swallows all FCM errors — security-critical notifications fail silently

**File:** `apps/api/internal/infrastructure/fcm/notifier.go:113`

All errors from the underlying notifier are discarded and `nil` is returned. Security-critical notifications (device wake, remote wipe commands) may silently fail with no way for the application layer to detect, log, or retry.

**Fix:** Log all swallowed errors at WARN level, and for designated high-priority notification types (`WakeDevice`, `SecurityCommand`) return the error rather than swallowing it.

---

## 11. Application Lifecycle & Startup/Shutdown

---

### BUG-60 🟡 Medium → No fail-fast validation of required environment variables at startup

**File:** `apps/api/cmd/api/api_main.go`

Missing required values (e.g. `JWT_SECRET`, `API_KEY_primary`, `DATABASE_URL`) are only detected when the code that needs them first runs — potentially on the first real request rather than at boot. The error surfaces as a runtime 500 rather than a clear startup failure.

**Fix:** At the top of `main()`, validate all required env vars and call `log.Fatal` with a clear message listing every missing variable.

---

### BUG-61 🟡 Medium → HTTP server starts accepting requests before DB migration completes

**File:** `apps/api/cmd/api/api_main.go` (startup ordering)

If the migration runner is invoked concurrently with or after the HTTP listener opens, early requests may hit a schema that is partially migrated, producing cryptic SQL errors or incorrect behavior.

**Fix:** Run migrations to completion (and verify success) before calling `srv.ListenAndServe`. Gate with a startup readiness probe.

---

### BUG-62 🟡 Medium → No graceful drain of in-flight HTTP requests on shutdown

**File:** `apps/api/cmd/api/api_main.go`

On SIGTERM, if the server calls `srv.Close()` instead of `srv.Shutdown(ctx)` with a drain timeout, in-flight requests are abruptly cut off. Clients receive connection resets rather than completed responses.

**Fix:** Use `srv.Shutdown(ctx)` with a reasonable timeout (e.g. 30 seconds) so in-flight requests complete before the process exits.

---

### BUG-63 🟢 Low → Health check endpoint (`/healthz`) does not probe DB connectivity

**File:** `apps/api/internal/api/server_routes.go:61`

The health check returns OK as long as the process is up. If the SQLite file is locked, corrupted, or the WAL is stuck, the load balancer or orchestrator sees a healthy service while all requests are failing.

**Fix:** Include a lightweight DB probe (`SELECT 1`) in the health check response. Return HTTP 503 if the probe fails.

---

## 12. Complete Issue Index

| ID | Status | Severity | Area | One-liner |
|----|--------|----------|------|-----------|
| BUG-01 | ✅ | 🔴 Critical | Storage | `scanDevices` missing `organization_id` — device list queries broken at runtime |
| BUG-02 | ✅ | 🔴 Critical | Device/Telemetry | Risk score fully device-controlled — server never recalculates or validates |
| BUG-10 | ✅ | 🔴 Critical | Commands | Strict HMAC middleware runs after the command handler — commands execute unsigned |
| BUG-51 | | 🔴 Critical | Migrations | Migrations not wrapped in transactions — partial migration leaves DB in unknown state |
| BUG-03 | | 🟠 High | Auth | Token rotation hardcodes `"operator"` role — admins lose privileges on first refresh |
| BUG-04 | ✅ | 🟠 High | Auth | Pre-MFA sessions bypass MFA enforcement permanently |
| BUG-05 | ✅ | 🟠 High | Auth | Disabling MFA does not revoke active sessions or refresh tokens |
| BUG-11 | ✅ | 🟠 High | HMAC | Device can forge signed requests for any other device's IMEI |
| BUG-16 | ✅ | 🟠 High | WebSocket | Message queue partial replay is silently dropped on reconnect |
| BUG-17 | ✅ | 🟠 High | WebSocket | `time.After()` in `SendWithDeliveryConfirmation` leaks timer goroutines |
| BUG-18 | ✅ | 🟠 High | WebSocket | Async persist/delete races itself — messages lost or replayed out of order |
| BUG-19 | ✅ | 🟠 High | WebSocket | WS device connection unauthenticated when `EnforceHMAC=false` |
| BUG-28 | ✅ | 🟠 High | GraphQL | No query depth/complexity limit — GraphQL DoS via nested queries |
| BUG-29 | ✅ | 🟠 High | GraphQL/REST | `TestWebhook` SSRF vector — URL not validated (fixed in both GraphQL and REST) |
| BUG-41 | | 🟠 High | Concurrency | Invitation email goroutines unmanaged — leak on slow/blocked email service |
| BUG-44 | | 🟠 High | Error Handling | `panic()` in non-test infrastructure code crashes API on crypto/domain errors |
| BUG-45 | | 🟠 High | Error Handling | Sentinels compared with `==` not `errors.Is()` — silently breaks on wrapped errors |
| BUG-53 | | 🟠 High | Audit | Audit goroutines fire-and-forget — security events dropped on shutdown |
| BUG-54 | | 🟠 High | Audit | Audit logs co-located with app data — trivially tampered by attacker |
| BUG-06 | ✅ | 🟡 Medium | Auth | Password change does not invalidate existing sessions |
| BUG-07 | ✅ | 🟡 Medium | Auth | GitHub OAuth allows unverified email — account takeover risk |
| BUG-08 | ✅ | 🟡 Medium | Auth | Registration has TOCTOU race on email uniqueness |
| BUG-09 | ✅ | 🟡 Medium | Auth | OAuth state repo nil check silently skips CSRF validation |
| BUG-12 | | 🟡 Medium | HMAC | Ciphertext (not plaintext) restored to request body — downstream parse fails |
| BUG-13 | | 🟡 Medium | HMAC | Rate limiter discards fractional refill time — effective rate lower than configured |
| BUG-14 | | 🟡 Medium | HMAC | Replay cache eviction is O(N) — latency spike under load |
| BUG-20 | ✅ | 🟡 Medium | WebSocket | Hub `Run` loop has no panic recovery — one nil-pointer crashes all WS connections |
| BUG-21 | ✅ | 🟡 Medium | WebSocket | Subscription callbacks spawn unbounded goroutines with no concurrency limit |
| BUG-22 | | 🟡 Medium | WebSocket | `SetOnline`/`SetOffline` called outside hub lock — race with concurrent unregister |
| BUG-23 | | 🟡 Medium | WebSocket | Rate limiter acquires two locks in inconsistent order — deadlock risk |
| BUG-24 | | 🟡 Medium | WebSocket | Large broadcast holds `RLock` for full iteration — starves new connections |
| BUG-30 | | 🟡 Medium | GraphQL | Playground and introspection enabled in production — schema exposure |
| BUG-31 | | 🟡 Medium | GraphQL | `SendCommand` not transactional — DB write succeeds but device never receives command |
| BUG-32 | | 🟡 Medium | GraphQL | Legacy `/graphql` route misses org membership checks in some resolvers |
| BUG-33 | | 🟡 Medium | GraphQL | `GetAllConnections` is O(N×M) — iterates all WS clients per device |
| BUG-35 | | 🟡 Medium | OTA | APK path traversal not fully blocked — encoded sequences pass |
| BUG-36 | | 🟡 Medium | OTA | Server never verifies APK integrity against manifest before serving |
| BUG-37 | | 🟡 Medium | Device | Device re-registration delete+create not atomic — race condition |
| BUG-38 | | 🟡 Medium | Device | `GetLatest`/`GetStats` have no org isolation — cross-tenant data leak |
| BUG-39 | | 🟡 Medium | Device | Telemetry field parse errors silently default to 0 — no range validation |
| BUG-40 | | 🟡 Medium | Device | WebSocket `SetOnline` and REST `ConfirmDevice` race each other |
| BUG-42 | | 🟡 Medium | Concurrency | Hub shutdown has no `WaitGroup` for client pump goroutines |
| BUG-46 | | 🟡 Medium | Error Handling | Logout failure silently ignored — server-side session may persist |
| BUG-47 | | 🟡 Medium | Error Handling | Verification email failure silently ignored — user cannot activate account |
| BUG-48 | | 🟡 Medium | Error Handling | Command delivery status sync failure ignored — dashboard shows wrong state |
| BUG-49 | | 🟡 Medium | Error Handling | `operatorRepo.Update()` return ignored — operator record silently stale |
| BUG-52 | | 🟡 Medium | Storage | Transaction rollback errors silently discarded |
| BUG-55 | | 🟡 Medium | Audit | Audit log failure in inbox service silently ignored |
| BUG-56 | | 🟡 Medium | FCM | Circuit breaker half-open counter not reset on panics — stays open permanently |
| BUG-57 | | 🟡 Medium | FCM | Malformed Firebase credentials crash startup instead of graceful degradation |
| BUG-58 | | 🟡 Medium | FCM | No persistent retry queue — FCM notifications lost on process restart |
| BUG-60 | | 🟡 Medium | Lifecycle | No fail-fast validation of required env vars at startup |
| BUG-61 | | 🟡 Medium | Lifecycle | HTTP server accepts requests before DB migration completes |
| BUG-62 | | 🟡 Medium | Lifecycle | No graceful drain of in-flight HTTP requests on SIGTERM |
| BUG-15 | | 🟢 Low | HMAC | Error codes let attackers oracle-test valid client IDs |
| BUG-25 | | 🟢 Low | WebSocket | Gzip writers not pooled — GC pressure under high telemetry |
| BUG-26 | | 🟢 Low | WebSocket | 500 ms latency threshold measured but never enforced |
| BUG-27 | | 🟢 Low | WebSocket | 64-bit metric fields read without atomics — race on 32-bit targets |
| BUG-34 | | 🟢 Low | GraphQL | GraphQL argument type assertions discard `ok` — zero-value used silently |
| BUG-43 | | 🟢 Low | Concurrency | Subscription callback errors never propagated or logged |
| BUG-50 | | 🟢 Low | Error Handling | Login notification email errors silently discarded |
| BUG-59 | | 🟢 Low | FCM | `SafeNotifier` swallows all FCM errors — security notifications fail silently |
| BUG-63 | | 🟢 Low | Lifecycle | Health check does not probe DB — unhealthy service looks healthy to orchestrator |

---

*Analysis performed across: `apps/api/internal/api/`, `apps/api/internal/application/`, `apps/api/internal/domain/`, `apps/api/internal/infrastructure/`, `apps/api/internal/ws/`, `apps/api/internal/audit/`, `apps/api/cmd/api/`.*  
*Pass 1 + 2 total: 4 Critical · 16 High · 31 Medium · 12 Low = **63 issues***

---

---

# Pass 3 — Deep Analysis (Inbox · Org/RBAC · Commands · Input Validation · Alerts/Webhooks)

> Third analysis pass covering subsystems not reached in Passes 1–2.  
> **35 additional issues** found. Grand total: **98 issues**.

---

## 13. Device Inbox & Registration Flow

---

### BUG-64 ✅ 🔴 Critical → `inbox_storage.go` never persists `organization_id` — all inbox entries have NULL org, breaking multi-tenant inbox entirely

**File:** `apps/api/internal/infrastructure/storage/inbox_storage.go:50`

The `Create` and `Update` methods on `InboxRepository` do not include `organization_id` in their `INSERT`/`UPDATE` statements, even though a migration added the column and `List` filters by it. Every inbox entry is stored with `organization_id = NULL`. The `List` query filtering by org ID therefore returns no results. Operators see an empty inbox regardless of how many pending device registrations exist.

**Fix:** Add `organization_id` to the `INSERT` and `UPDATE` column lists in both methods and bind it from the entity.

---

### BUG-65 ✅ 🔴 Critical → `/v1/device/inbox` is public with no authentication — anyone can flood pending registrations (DoS)

**File:** `apps/api/internal/api/server_routes.go:106` + `apps/api/internal/application/inbox/inbox_service.go`

The inbox endpoint accepts device registration requests with no proof of device identity beyond a Luhn-checked IMEI. There is no API key, no challenge-response, and no device attestation required. Any actor with a list of valid IMEIs (obtainable from public databases or by sniffing) can create thousands of pending registrations. Rate limiting is IP-based only, making it trivially bypassable with a proxy list.

**Fix:** ✅ Fixed with Firebase App Check verification for device attestation.

---

### BUG-66 ✅ 🟠 High → `commandSecret` stored in plaintext in `inbox_requests` table

**File:** `apps/api/internal/application/inbox/inbox_service.go:242, 552` + `inbox_storage.go:64, 201`

The `commandSecret` is generated with `crypto/rand` (correct). It is stored hashed in the `devices` table (correct). However, the plaintext secret is also written into the `inbox_requests` table for the duration of the pending registration window — potentially 30+ days. Anyone who gains read access to the database (via SQL injection, a backup leak, or a compromised read replica) obtains the plaintext command secret for every pending device.

**Fix:** ✅ Fixed — only `command_secret_hash` is stored in `inbox_requests` table, not plaintext.

---

### BUG-67 ✅ 🟠 High → Confirmation token (`commandSecret`) is not single-use — `/v1/device/confirm` can be called repeatedly

**File:** `apps/api/internal/application/device/device_service.go:816`

`ConfirmDevice` validates the `commandSecret`, updates `LastSeen`, marks the device online, and returns 200 OK. There is no state transition that marks the secret as consumed. Calling `/confirm` a second time with the same secret succeeds again. An attacker who intercepts the secret in transit (or reads it from the plaintext inbox table — BUG-66) can re-confirm the device at any time, resetting its `LastSeen` timestamp and masking its true last-contact time from the operator.

**Fix:** ✅ Fixed — `ConfirmDevice` checks `Lifecycle.IsRegistered()` which prevents repeat confirmations.

---

### BUG-68 ✅ 🟠 High → No background worker executes `deletion_scheduled_at` — deregistered devices are never actually deleted

**File:** `apps/api/internal/application/device/device_service.go:661`

`DeregisterDevice` sets `deletion_scheduled_at = now + 30 days`. There is no scheduler, ticker, or cron job anywhere in the codebase that reads this column and performs the actual deletion. Deregistered devices accumulate indefinitely. The 30-day grace period is documented as a hard promise but is never honored.

**Fix:** ✅ Fixed — `DeviceDeletionWorker` exists and runs periodically to delete scheduled devices.

---

### BUG-69 ✅ 🟠 High → Device reconnecting during deletion window does not cancel the scheduled deletion

**File:** `apps/api/internal/ws/hub.go` + `apps/api/internal/application/device/device_service.go`

If a device reconnects via WebSocket after being deregistered, `SetOnline(true)` is called and it starts sending telemetry again. But `deletion_scheduled_at` remains set. When the background worker from BUG-68 is eventually added, it will delete an actively communicating device.

**Fix:** ✅ Fixed — `SetOnline(true)` clears `deletion_scheduled_at` when device reconnects.

---

### BUG-70 ✅ 🟡 Medium → Successful registration does not clean up the inbox entry — records accumulate forever

**File:** `apps/api/internal/application/inbox/inbox_service.go:426`

Inbox cleanup only triggers when the *same IMEI re-registers*. After a device completes registration (inbox → confirmed → active), its `inbox_requests` row is never deleted. Over time the table grows without bound.

**Fix:** ✅ Fixed — confirm handler calls `DeleteByIMEI` after successful device confirmation.

---

### BUG-71 ✅ 🟡 Medium → `CreateInboxRequest` cleanup + create are not in a transaction — TOCTOU race for same IMEI

**File:** `apps/api/internal/application/inbox/inbox_service.go:431–441`

The service deletes an existing inbox entry for the IMEI and then creates a new one as two separate statements without a transaction. Two concurrent requests for the same IMEI can both read "no existing entry", both skip the delete, and both attempt the insert — hitting a unique constraint error instead of a clean idempotent upsert.

**Fix:** ✅ Fixed by adding `CreateOrReplace` method using `INSERT OR REPLACE` for atomic operation.

---

## 14. Organization Management & RBAC

---

### BUG-72 ✅ 🔴 Critical → Removing an org member does not revoke their active sessions or tokens — they retain access until natural expiry

**File:** `apps/api/internal/application/organization/member_service.go:85`

`RemoveMember` soft-deletes the membership record. There is no call to `RevokeAllRefreshTokens`, `LogoutAll`, or any session invalidation scoped to that organization. The removed operator keeps all active sessions and can continue making authenticated requests against the org's resources until their JWT expires (up to 7 days by default).

**Fix:** ✅ Fixed — `RemoveMember` calls `LogoutAll` and `RevokeAllRefreshTokens` after membership removal.

---

### BUG-73 ✅ 🟠 High → `GET /v1/organizations/:id` has no minimum role check — any member including `viewer` sees full org details

**File:** `apps/api/internal/api/handlers/organization/organization_handler.go:131`

The handler manually verifies membership but does not enforce a minimum role level. A `viewer`-role member (lowest privilege) can read the full org details response, which may include configuration, billing metadata, and member counts that should require at least `operator` or `admin`.

**Fix:** ✅ Fixed — handler checks for `RoleOperator.Level()` minimum before returning org details.

---

### BUG-74 ✅ 🟠 High → `GET /v1/organizations/:id/members` has no role check — any member can enumerate all other members and their roles

**File:** `apps/api/internal/api/handlers/organization/member_handler.go:53`

Any org member can list every other member's name, email, and role. In a multi-tenant environment, this leaks the org's full operator roster and role hierarchy to the lowest-privilege users.

**Fix:** ✅ Fixed — handler checks for `RoleOperator.Level()` minimum before listing members.

---

### BUG-75 🟡 Medium → Org `admin` can modify `maxMembers` and `isActive` without `super_admin` approval

**File:** `apps/api/internal/api/handlers/organization/organization_handler.go:179`

`PATCH /v1/organizations/:id` is accessible to `admin` role. The body accepts `maxMembers` and `isActive` fields. If billing is tied to member limits, an admin can silently raise limits or deactivate the org without a superadmin. At minimum `isActive` changes should require `super_admin`.

**Fix:** Add a field-level check: if the request includes `maxMembers` or `isActive`, require `super_admin` role before applying those fields.

---

### BUG-76 🟡 Medium → Org deletion soft-deletes only — devices, telemetry, commands, and sessions are not cleaned up

**File:** `apps/api/internal/application/organization/organization_service.go:265`

`DeleteOrganization` soft-deletes the org, memberships, and expires invitations. It does not: hard-delete or deregister devices, delete telemetry, purge command queues, revoke operator sessions, or clean up FCM registrations. A deleted org's data remains in full in the DB.

**Fix:** Expand deletion to cascade through all owned entities, or schedule them for deletion via the worker in BUG-68's fix. At minimum, revoke all member sessions at deletion time.

---

### BUG-77 🟡 Medium → Invitation token has no DB-level uniqueness constraint — concurrent accepts could both succeed

**File:** `apps/api/internal/application/organization/invitation_service.go` + invitation storage

`Accept` transitions the invitation to `accepted` state and reads `CanBeAccepted()`. If two concurrent requests arrive with the same token (e.g. double-click), both may pass the `HasResponded() == false` check before either has written the state transition, resulting in duplicate membership records.

**Fix:** Add a `UNIQUE` index on `(organization_id, invitee_email)` in the memberships table, and handle the unique constraint violation as a graceful "already accepted" response.

---

## 15. Command Service

---

### BUG-78 🟡 Medium → HMAC verification passes an empty `deviceID` string in global middleware — per-device key lookup silently fails

**File:** `apps/api/internal/api/server_handlers.go:162` + `apps/api/internal/infrastructure/security/request_signer/verifier.go:180`

The global `requireHMAC` middleware calls `Verify` with `deviceID = ""`. If the verifier's key-lookup function uses `deviceID` as the key identifier (to fetch the per-device secret), passing an empty string will either look up the wrong key or return an error that is swallowed, effectively disabling per-device HMAC verification on routes that use the global middleware version.

**Fix:** Resolve `deviceID` from `c.Param("imei")` before passing it to `Verify`. Fail with 401 if `deviceID` is empty on routes that require it.

---

### BUG-79 🟡 Medium → Commands stuck in `pending` or `delivered` state are never cleaned up — table grows without bound

**File:** `apps/api/internal/application/command/command_service.go` (no TTL logic present)

There is no TTL, expiry field, or cleanup job for commands. A device that goes permanently offline leaves its `pending` commands in the DB forever. Over a fleet lifetime the `commands` table grows without bound and GetPending queries slow down proportionally.

**Fix:** Add a `expires_at` column to commands (set at dispatch time, e.g. `now + 24h`). Add the background worker from BUG-68 to also sweep expired commands and transition them to `failed`.

---

### BUG-80 🟡 Medium → No limit on pending commands per device — a device can accumulate unlimited queued commands

**File:** `apps/api/internal/application/command/command_service.go` + `command_storage.go`

`SendCommand` does not check how many commands are already `pending` for the target device before inserting a new one. An operator (or an attacker with operator access) can enqueue thousands of commands, bloating the DB and causing the device to be overwhelmed with work on reconnection.

**Fix:** Before inserting, query `COUNT(*) WHERE device_id = ? AND status IN ('pending', 'delivered')`. Return HTTP 429 if it exceeds a configurable cap (e.g. 100).

---

### BUG-81 🟡 Medium → No automatic retry or backoff for failed commands — operators must manually retry every failure

**File:** `apps/api/internal/application/command/command_service.go:206–221`

Retry is entirely manual via `POST /v1/command/:dispatchId/retry`. There is no automatic retry, no exponential backoff, and no max-retry-count enforcement. System-level commands (OTA push, remote wipe) that fail due to a transient network issue require operator intervention to retry.

**Fix:** Add a `retry_count` and `next_retry_at` column to commands. The background worker should automatically retry commands that haven't exceeded `max_retries` using exponential backoff, and move them to `failed` when the limit is reached (then notify via FCM/webhook).

---

## 16. Input Validation

---

### BUG-82 🟡 Medium → Pagination `limit` parameters have no upper bound across multiple handlers — full table scan possible

**Files:**
- `apps/api/internal/api/handlers/device/device_telemetry_handler.go:74`
- `apps/api/internal/api/handlers/device/device_logs_handler.go:63`
- `apps/api/internal/api/handlers/device/device_metrics_handler.go:131`

All three parse `limit` from query parameters with no maximum enforced. A caller can send `limit=1000000` and the query runs against the full table, returning an enormous response and saturating SQLite I/O.

**Fix:** Clamp `limit` to a configurable maximum (e.g. 1000) in each handler: `if limit > maxLimit { limit = maxLimit }`.

---

### BUG-83 🟡 Medium → Time-range query parameters accept arbitrary values including negative timestamps and year 9999

**Files:**
- `apps/api/internal/api/handlers/device/device_telemetry_handler.go:67–75`
- `apps/api/internal/api/handlers/device/device_metrics_handler.go:73–78`

`startTime` and `endTime` are parsed as raw `int64` Unix milliseconds with no range validation. Passing `startTime=0` scans from the Unix epoch. Passing `endTime=253402300799000` (year 9999) scans forward indefinitely. Combined with no row limit (BUG-82), this can result in full-table scans over millions of rows.

**Fix:** Validate that `startTime >= 0`, `endTime > startTime`, and enforce a maximum query window (e.g. 90 days).

---

### BUG-84 🟡 Medium → Device settings thresholds accept out-of-range values — negative or million-percent thresholds stored

**File:** `apps/api/internal/api/handlers/device/device_settings_handler.go:119, 214`

`UpdateThresholds` validates that `warn < crit` (relative ordering) but does not validate that values are within meaningful ranges. A caller can set `RiskWarn: -100` or `ThermalTemp: 999999`, which will never trigger an alert (or will always trigger one), silently breaking the alert system for that device.

**Fix:** Add domain-level range validation in the threshold entity: `0 ≤ RiskWarn ≤ 100`, `0 ≤ RiskCrit ≤ 100`, `0 ≤ ThermalWarn ≤ 200` etc.

---

### BUG-85 🟢 Low → Path parameters (IMEI, ID, clientId) are checked for non-empty but not validated for format

**Files:** Multiple handlers in `apps/api/internal/api/handlers/`

`:imei`, `:id`, and `:clientId` path parameters are only verified to be non-empty. No UUID format, alphanumeric, or IMEI-length check is applied at the handler layer. A caller can pass `../../../../etc/passwd` as an IMEI (path traversal is blocked by Gin's router, but the value reaches the DB query unchanged).

**Fix:** Add a format-validation helper: UUIDs should match the UUID regex, IMEIs should be exactly 15 digits, and other identifiers should be alphanumeric-only. Return HTTP 400 on mismatch before any DB call.

---

### BUG-86 🟢 Low → String fields in JSON binding models have no length bounds — oversized values reach the DB

**Files:** `auth_register.go`, `device_settings_handler.go`, `member_handler.go`, and others

No `binding:"max=..."` or manual length check is applied to free-text fields (operator name, device name, model, manufacturer, etc.). A caller can send a 10 MB string that SQLite stores as a TEXT blob, ballooning the DB and slowing future queries.

**Fix:** Add `binding:"max=255"` (or appropriate limits) to all string fields in request DTOs. Enforce at the handler layer before reaching the service.

---

## 17. Alerting, Metrics & Webhooks

---

### BUG-87 ✅ 🔴 Critical → Webhook URL is never validated — SSRF allows scanning internal network from the server

**File:** `apps/api/internal/infrastructure/fcm/webhook_client.go:145` (or equivalent webhook client file)

The webhook client POSTs to any URL stored in org settings without resolving or validating the hostname. An attacker with org `admin` access can set the webhook URL to `http://169.254.169.254/latest/meta-data/` (cloud metadata endpoint), `http://localhost:6060/debug/pprof/`, or any internal service, and use the server as a pivot to probe and exfiltrate internal infrastructure.

**Fix:** Before storing or using a webhook URL: resolve its hostname, reject RFC-1918 ranges (10.x, 172.16–31.x, 192.168.x), loopback (127.x), and link-local (169.254.x) addresses. Enforce `https` scheme only.

---

### BUG-88 🔴 Critical → No alert deduplication — every telemetry sample above threshold creates a separate alert record, notification, and webhook call

**File:** `apps/api/internal/application/event/event_processor.go:257, 277–328`

Thresholds are evaluated on every incoming telemetry frame with no state memory. A device streaming at 1 Hz with temperature above threshold for 10 minutes generates 600 alert records, 600 DB writes, 600 email/webhook calls. This can exhaust storage, overwhelm a webhook receiver, and flood an operator's inbox.

**Fix:** Maintain an "active alert" state table keyed by `(device_id, alert_type)`. Emit a new alert only when transitioning from normal → breach. Emit a resolution event when transitioning breach → normal. Suppress repeated notifications until the state changes.

---

### BUG-89 🟠 High → Telemetry retention uses a global `LIMIT 5000` — Device B's inserts silently delete Device A's history

**File:** `apps/api/internal/infrastructure/storage/telemetry_storage.go:79`

```sql
DELETE FROM telemetry WHERE id NOT IN (
    SELECT id FROM telemetry ORDER BY received_at DESC LIMIT 5000
)
```

This runs after every single insert and retains only the 5000 most recent rows *across all devices*. In a fleet with 100 active devices, Device A gets at most 50 rows of history. A burst of data from Device B pushes Device A's history entirely off the table.

**Fix:** Change the retention query to be per-device:
```sql
DELETE FROM telemetry WHERE device_id = ? AND id NOT IN (
    SELECT id FROM telemetry WHERE device_id = ? ORDER BY received_at DESC LIMIT 500
)
```

---

### BUG-90 🟠 High → O(N) `DELETE` on every telemetry `INSERT` — DB performance degrades with fleet size

**File:** `apps/api/internal/infrastructure/storage/telemetry_storage.go:79`

The retention `DELETE` scans and rewrites the entire `telemetry` table on every row insert. With 5000 rows this is already slow; with a fleet of devices streaming at 1 Hz it runs continuously, holding a write lock and blocking all concurrent reads.

**Fix:** Decouple retention from insert. Run retention as a low-priority background job (e.g. every 5 minutes, per device) rather than inline with every write.

---

### BUG-91 🟡 Medium → No alert auto-resolution — once an alert fires, it appears active forever even after the metric recovers

**File:** `apps/api/internal/application/event/event_processor.go`

The event processor only emits breach events. There is no "resolved" event when a metric returns below threshold. Operators have no way to know if an alert condition is still active or was resolved hours ago.

**Fix:** As part of the deduplication fix (BUG-88), emit a `EventTypeResolved` event when the device's metric transitions from above threshold to below. Include the duration the breach was active.

---

### BUG-92 🟡 Medium → No hysteresis on threshold checks — alert flapping when metric oscillates near the threshold

**File:** `apps/api/internal/application/event/event_processor.go:470`

A metric oscillating between 89 and 91 against a threshold of 90 will fire an alert, resolve (once BUG-91 is fixed), fire again, resolve again — repeatedly. Each cycle generates new records and notifications.

**Fix:** Implement a hysteresis band: alert triggers at `≥ critThreshold`, clears only when `< (critThreshold - hysteresisBand)`. A typical hysteresis band is 5–10% of the threshold value.

---

### BUG-93 🟡 Medium → Concurrent telemetry frames for the same device can create duplicate alerts — no active-alert mutex

**File:** `apps/api/internal/application/event/event_processor.go:271`

The event processor has no per-device lock or active-alert state cache. Two WebSocket frames arriving within the same millisecond for the same device are both processed concurrently, both find "no active alert", and both insert a new alert record.

**Fix:** The active-alert state table from BUG-88 should use an `INSERT OR IGNORE` with a unique index on `(device_id, alert_type)` to make alert creation idempotent at the DB level.

---

### BUG-94 🟡 Medium → `GetAggregatedMetrics` has no row limit for large time ranges — can scan millions of rows

**File:** `apps/api/internal/infrastructure/storage/metrics_storage.go:110`

The aggregation query is bounded by `startTime`/`endTime` but not by a row count. A query for a year-wide range on a high-frequency device could scan millions of rows in a single SQLite read, blocking all other writes for seconds.

**Fix:** Add `LIMIT N` to the inner query, or enforce a maximum time window (e.g. 30 days) at the handler layer before the query reaches the storage layer.

---

### BUG-95 🟡 Medium → Webhook delivery has no retry logic and no circuit breaker — a down receiver is hammered on every alert

**File:** `apps/api/internal/infrastructure/fcm/webhook_client.go:56, 136`

The webhook client has a 10-second timeout but no retry on transient failure and no circuit breaker. Every alert event fires a fresh HTTP call to the (potentially down) receiver. With alert storms (BUG-88), this means thousands of concurrent outbound connections to the same failing endpoint.

**Fix:** Wrap webhook delivery in the circuit breaker pattern already used for FCM (`fcm_circuit_breaker.go`). Add exponential backoff with a max of 3 retries for transient 5xx errors.

---

### BUG-96 🟡 Medium → `ProcessTelemetry` called without a `recover()` in `ReadPump` — a bad telemetry payload drops the device connection

**File:** `apps/api/internal/ws/ws_client.go:196`

`ProcessTelemetry` is called directly inside `ReadPump`. If `ProcessTelemetry` panics (e.g. due to a nil pointer in the event processor on malformed data), the `ReadPump` goroutine dies, the device's WebSocket connection is closed, and the device must reconnect. There is no recovery.

**Fix:** Wrap the `ProcessTelemetry` call in `defer func() { if r := recover(); r != nil { log.Error(...) } }()` inside `ReadPump`, so a panic on a single bad frame drops the frame rather than the connection.

---

---

## Complete Issue Index — All 3 Passes

| ID | Status | Sev | Area | One-liner |
|----|--------|----|-----|---------|
| BUG-01 | ✅ | 🔴 | Storage | `scanDevices` missing `organization_id` — device list queries broken |
| BUG-02 | ✅ | 🔴 | Telemetry | Risk score fully device-controlled — no server validation |
| BUG-10 | ✅ | 🔴 | Commands | Strict HMAC middleware runs after command handler |
| BUG-51 | | 🔴 | Migrations | Migrations not in transactions — partial migration possible |
| BUG-64 | | 🔴 | Inbox | `inbox_storage` never persists `organization_id` — multi-tenant inbox broken |
| BUG-65 | ✅ | 🔴 | Inbox | `/inbox` public with no auth — unlimited fake registration flood |
| BUG-87 | | 🔴 | Webhooks | Webhook URL not validated — SSRF internal network pivot |
| BUG-88 | | 🔴 | Alerts | No alert deduplication — 1 telemetry sample = 1 alert, 1 notification |
| BUG-03 | | 🟠 | Auth | Token rotation hardcodes `"operator"` — admins lose role on first refresh |
| BUG-04 | ✅ | 🟠 | Auth | Pre-MFA sessions bypass MFA enforcement |
| BUG-05 | ✅ | 🟠 | Auth | Disabling MFA doesn't revoke sessions |
| BUG-11 | ✅ | 🟠 | HMAC | Device can forge requests for any other device's IMEI |
| BUG-16 | ✅ | 🟠 | WebSocket | Message queue partial replay silently dropped on reconnect |
| BUG-17 | ✅ | 🟠 | WebSocket | `time.After()` in delivery confirmation leaks timer goroutines |
| BUG-18 | ✅ | 🟠 | WebSocket | Async persist/delete races itself — messages lost or replayed out of order |
| BUG-19 | ✅ | 🟠 | WebSocket | WS unauthenticated when `EnforceHMAC=false` |
| BUG-28 | ✅ | 🟠 | GraphQL | No query depth/complexity limit — DoS via nested queries |
| BUG-29 | ✅ | 🟠 | GraphQL/REST | `TestWebhook` SSRF vector — URL not validated (fixed in both GraphQL and REST) |
| BUG-41 | | 🟠 | Concurrency | Invitation email goroutines unmanaged — leak on slow email service |
| BUG-44 | | 🟠 | Errors | `panic()` in non-test infrastructure crashes API |
| BUG-45 | | 🟠 | Errors | Sentinels compared with `==` — silently breaks on wrapped errors |
| BUG-53 | | 🟠 | Audit | Audit goroutines fire-and-forget — events dropped on shutdown |
| BUG-54 | | 🟠 | Audit | Audit logs co-located with app data — trivially tampered |
| BUG-66 | ✅ | 🟠 | Inbox | `commandSecret` stored in plaintext in `inbox_requests` table |
| BUG-67 | ✅ | 🟠 | Inbox | Confirmation token is not single-use — `/confirm` succeeds repeatedly |
| BUG-68 | ✅ | 🟠 | Inbox | No background worker for `deletion_scheduled_at` — devices never actually deleted |
| BUG-69 | ✅ | 🟠 | Inbox | Device reconnecting during deletion window doesn't cancel scheduled deletion |
| BUG-72 | ✅ | 🟠 | Org/RBAC | Removing a member doesn't revoke their sessions — access persists |
| BUG-73 | ✅ | 🟠 | Org/RBAC | `GET /organizations/:id` has no role check — viewer sees full org details |
| BUG-74 | ✅ | 🟠 | Org/RBAC | `GET /organizations/:id/members` — any member enumerates all members + roles |
| BUG-89 | | 🟠 | Telemetry | Global 5000-row cap — Device B's data silently deletes Device A's history |
| BUG-90 | | 🟠 | Telemetry | O(N) DELETE on every INSERT — DB performance degrades continuously |
| BUG-06 | ✅ | 🟡 | Auth | Password change doesn't invalidate existing sessions |
| BUG-07 | ✅ | 🟡 | Auth | GitHub OAuth allows unverified email |
| BUG-08 | ✅ | 🟡 | Auth | Registration has TOCTOU race on email uniqueness |
| BUG-09 | ✅ | 🟡 | Auth | OAuth state repo nil silently skips CSRF validation |
| BUG-12 | | 🟡 | HMAC | Ciphertext not plaintext restored to request body |
| BUG-13 | | 🟡 | HMAC | Rate limiter discards fractional refill time |
| BUG-14 | | 🟡 | HMAC | Replay cache eviction is O(N) |
| BUG-20 | ✅ | 🟡 | WebSocket | Hub `Run` loop has no panic recovery |
| BUG-21 | ✅ | 🟡 | WebSocket | Subscription callbacks spawn unbounded goroutines |
| BUG-22 | | 🟡 | WebSocket | `SetOnline`/`SetOffline` race outside hub lock |
| BUG-23 | | 🟡 | WebSocket | Rate limiter inconsistent lock ordering — deadlock risk |
| BUG-24 | | 🟡 | WebSocket | Large broadcast holds `RLock` for full iteration |
| BUG-30 | | 🟡 | GraphQL | Playground + introspection enabled in production |
| BUG-31 | | 🟡 | GraphQL | `SendCommand` not transactional across DB + WS + FCM |
| BUG-32 | | 🟡 | GraphQL | Legacy `/graphql` route misses org membership checks |
| BUG-33 | | 🟡 | GraphQL | `GetAllConnections` is O(N×M) |
| BUG-35 | | 🟡 | OTA | APK path traversal not fully blocked |
| BUG-36 | | 🟡 | OTA | Server never verifies APK integrity before serving |
| BUG-37 | | 🟡 | Device | Device re-registration delete+create not atomic |
| BUG-38 | | 🟡 | Device | `GetLatest`/`GetStats` have no org isolation |
| BUG-39 | | 🟡 | Device | Telemetry field parse errors silently default to 0 |
| BUG-40 | | 🟡 | Device | WS `SetOnline` and REST `ConfirmDevice` race each other |
| BUG-42 | | 🟡 | Concurrency | Hub shutdown no `WaitGroup` for client pump goroutines |
| BUG-46 | | 🟡 | Errors | Logout failure silently ignored — session may persist |
| BUG-47 | | 🟡 | Errors | Verification email failure silently ignored |
| BUG-48 | | 🟡 | Errors | Command delivery status sync failure ignored |
| BUG-49 | | 🟡 | Errors | `operatorRepo.Update()` return ignored |
| BUG-52 | | 🟡 | Storage | Transaction rollback errors silently discarded |
| BUG-55 | | 🟡 | Audit | Audit log failure in inbox service silently ignored |
| BUG-56 | | 🟡 | FCM | Circuit breaker half-open counter not reset on panics |
| BUG-57 | | 🟡 | FCM | Malformed Firebase credentials crash startup |
| BUG-58 | | 🟡 | FCM | No persistent retry queue for FCM notifications |
| BUG-60 | | 🟡 | Lifecycle | No fail-fast env var validation at startup |
| BUG-61 | | 🟡 | Lifecycle | HTTP server accepts requests before migrations complete |
| BUG-62 | | 🟡 | Lifecycle | No graceful drain of in-flight requests on SIGTERM |
| BUG-70 | ✅ | 🟡 | Inbox | Successful registration doesn't clean up inbox entry |
| BUG-71 | ✅ | 🟡 | Inbox | CreateInboxRequest cleanup+create not in transaction |
| BUG-75 | ✅ | 🟡 | Org/RBAC | `admin` can modify `maxMembers`/`isActive` without superadmin |
| BUG-76 | ✅ | 🟡 | Org/RBAC | Org deletion doesn't clean up devices, telemetry, sessions |
| BUG-77 | ✅ | 🟡 | Org/RBAC | Invitation accept has no DB-level uniqueness guard |
| BUG-78 | ✅ | 🟡 | Commands | Global HMAC middleware passes empty `deviceID` |
| BUG-79 | ✅ | 🟡 | Commands | No TTL on commands — stuck commands accumulate forever |
| BUG-80 | ✅ | 🟡 | Commands | No cap on pending commands per device |
| BUG-81 | ✅ | 🟡 | Commands | No automatic retry/backoff — all retries are manual |
| BUG-82 | ✅ | 🟡 | Input | `limit` params have no upper bound — full table scan possible |
| BUG-83 | ✅ | 🟡 | Input | Time-range params accept arbitrary values including epoch 0 and year 9999 |
| BUG-84 | ✅ | 🟡 | Input | Threshold values accept negative or million-percent — silent alert breakage |
| BUG-91 | | 🟡 | Alerts | No alert auto-resolution — alert appears active forever after metric recovers |
| BUG-92 | | 🟡 | Alerts | No hysteresis — alert flapping when metric oscillates near threshold |
| BUG-93 | | 🟡 | Alerts | Concurrent telemetry frames create duplicate alerts |
| BUG-94 | | 🟡 | Metrics | `GetAggregatedMetrics` no row limit for large time ranges |
| BUG-95 | | 🟡 | Webhooks | No retry or circuit breaker for webhook delivery |
| BUG-96 | | 🟡 | WebSocket | `ProcessTelemetry` panic kills device connection — no recover() |
| BUG-15 | | 🟢 | HMAC | Error codes oracle valid client IDs |
| BUG-25 | | 🟢 | WebSocket | Gzip writers not pooled — GC pressure |
| BUG-26 | | 🟢 | WebSocket | 500 ms latency threshold measured but never enforced |
| BUG-27 | | 🟢 | WebSocket | 64-bit metric fields read without atomics on 32-bit targets |
| BUG-34 | | 🟢 | GraphQL | Argument type assertions discard `ok` — zero-value used silently |
| BUG-43 | | 🟢 | Concurrency | Subscription callback errors never propagated |
| BUG-50 | | 🟢 | Errors | Login notification email errors silently discarded |
| BUG-59 | | 🟢 | FCM | `SafeNotifier` swallows all FCM errors |
| BUG-63 | | 🟢 | Lifecycle | Health check doesn't probe DB |
| BUG-85 | ✅ | 🟢 | Input | Path params not validated for format (UUID, IMEI digits) |
| BUG-86 | | 🟢 | Input | String fields in JSON models have no length bounds |

---

**Grand total — all 3 passes: 8 Critical · 24 High · 52 Medium · 14 Low = 98 issues**

*Areas covered: Auth · Sessions · HMAC/Signing · WebSocket · GraphQL · OTA · Device Management · Concurrency · Error Handling · Database/Migrations · Audit · FCM · Lifecycle · Inbox/Registration · Org/RBAC · Command Service · Input Validation · Alerting/Metrics · Webhooks*
